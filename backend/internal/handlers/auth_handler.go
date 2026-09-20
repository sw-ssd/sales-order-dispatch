// Package handlers 提供認證 HTTP 端點:AuthService Connect-RPC(Login / Refresh / Logout /
// RegisterComplete)與 Google OIDC 導向 / callback(公開 REST 端點)。
package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alexedwards/scs/v2"
	"golang.org/x/oauth2"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/services"
)

const (
	// RoleGuest 為首次 OIDC 登入的員工角色(待審核 / 待完成註冊)。
	RoleGuest = "guest"

	// RegistrationTokenCookie 為首次 OIDC 登入(尚未建帳號)時派發的一次性註冊憑證 cookie。
	RegistrationTokenCookie = "registration_token"
)

// seatGuard 為席位守衛所需的最小介面（consumer 端定義）：AuthHandler 只用 CheckLimit，
// `*entitlements.Service` 與 `entitlements.Unlimited()` 都直接滿足它。
//
// **未注入（nil）時建帳號路徑一律拒絕**（見 guardSeats）：漏注入等於無守衛＝靜默超額，
// 而這是 fail-closed 的護欄、不是可選功能；生產的組裝鏈固定注入（domains.go 的 mountAuth）。
type seatGuard interface {
	CheckLimit(ctx context.Context, companyID int, feature string, delta int) error
}

// AuthDeps 集中 AuthHandler 依賴。
type AuthDeps struct {
	Cfg       *config.Config
	DB        *ent.Client
	Tokens    *auth.TokenManager
	Lockout   *auth.LoginLock
	OneTime   *auth.OneTimeStore
	Sessions  *scs.SessionManager
	OAuth     *oauth2.Config
	Exchanger auth.OAuthExchanger
	Verifier  auth.OIDCVerifier
	// Entitlements 為席位上限守衛（見 guardSeats）。
	Entitlements seatGuard
}

// AuthHandler 實作 salesorder.v1.AuthService 與 OIDC 公開端點。
type AuthHandler struct {
	salesorderv1connect.UnimplementedAuthServiceHandler
	deps AuthDeps
}

// NewAuthHandler 建立 AuthHandler。
func NewAuthHandler(deps AuthDeps) *AuthHandler {
	return &AuthHandler{deps: deps}
}

// SetOIDC 注入 OIDC 依賴（OAuth config / exchanger / verifier），
// 由 server 組裝時在 DB 與 Valkey 就緒後呼叫（避免 Google discovery 阻斷其餘路由）。
func (h *AuthHandler) SetOIDC(cfg *oauth2.Config, exchanger auth.OAuthExchanger, verifier auth.OIDCVerifier) {
	h.deps.OAuth = cfg
	h.deps.Exchanger = exchanger
	h.deps.Verifier = verifier
}

// guardSeats 在**建立 users 列之前**檢查席位上限（席位＝該公司所有非 inactive 帳號，
// 見 internal/services/counters.go）。AuthHandler 有兩條會新增 users 列的路徑：
// OIDC 首次登入（hd 對應公司 → 直接建 guest）與 RegisterComplete 的 registration-token 分支
// （建 guest）。兩者都不需要管理權，漏了守衛就是「員工用公司網域的 Google 帳號首次登入即靜默
// 超額佔席位」，而席位的超額正是這條路徑造成的。
//
// 公司由呼叫端明示帶入：這些是**無租戶身分**的系統路徑（見 systemScope），不能像業務服務那樣
// 由身分推導公司，故走 services.GuardQuotaForCompany（與 guardQuota 同一份實作）。
//
// 未注入守衛（nil）→ **一律拒絕**：漏注入等於無守衛，而這是 fail-closed 的護欄不是可選功能；
// 生產的組裝鏈固定注入（domains.go 的 mountAuth），所以 nil 只可能來自組裝漏掉。
func (h *AuthHandler) guardSeats(ctx context.Context, companyID int) error {
	if h.deps.Entitlements == nil {
		return errcode.SysInternal.Wrap(errors.New("auth: 未注入席位守衛（AuthDeps.Entitlements）"))
	}
	return services.GuardQuotaForCompany(ctx, h.deps.Entitlements, companyID, entitlements.LimitSeats, 1)
}

// errCodeID 取 errcode 註冊碼（例 PLAT-5001）。OIDC 是 HTTP redirect，沒有 ErrorInfo 通道
// （見 redirectError），只能把碼放進 query 由前端查表顯示可行動訊息
// （frontend/src/lib/errcode.ts 的碼 → 訊息對照）；取不到碼一律當系統錯誤。
func errCodeID(err error) string {
	ce, ok := err.(*connect.Error)
	if !ok {
		return errcode.SysInternal.ID()
	}
	for _, d := range ce.Details() {
		v, derr := d.Value()
		if derr != nil {
			continue
		}
		if info, ok := v.(*v1.ErrorInfo); ok {
			return info.GetCode()
		}
	}
	return errcode.SysInternal.ID()
}

// ---------------------------------------------------------------------------
// AuthService RPC

// systemScope 在系統範圍(scope=all)交易內執行**未登入路徑**的 DB 工作(登入/註冊/OIDC/guest),
// 並把該交易注入 ctx:同一段程式碼裡的 dbtenant.Client(ctx, h.deps.DB) 因而落在這條系統交易上
// (entity 的 lazy edge 查詢也才不會落到已提交的交易上)。
//
// 為何需要:這些路徑都還沒有身分,而 users/companies 在 00028 之後受 RLS 約束 —— 走請求交易的
// (無 scope)只會回 0 列 → 登入一律「客戶編號或密碼錯誤」、註冊一律「公司不存在」。
func (h *AuthHandler) systemScope(ctx context.Context, fn func(context.Context) error) error {
	return dbtenant.SystemScopeTx(ctx, h.deps.DB, func(tx *ent.Tx) error {
		return fn(dbtenant.WithTenantTx(ctx, tx))
	})
}

// userByEmail 以系統範圍載入使用者(未登入路徑的 email 查找;回傳 ent.IsNotFound 供呼叫端判別)。
func (h *AuthHandler) userByEmail(ctx context.Context, email string) (*ent.User, error) {
	var u *ent.User
	err := h.systemScope(ctx, func(ctx context.Context) error {
		var qerr error
		u, qerr = dbtenant.Client(ctx, h.deps.DB).User.Query().Where(user.EmailEQ(email)).Only(ctx)
		return qerr
	})
	return u, err
}

// Login 客戶密碼登入(T12):以 customer_code(= users.account_name)查客戶帳號,
// bcrypt 驗證密碼;連續 5 次失敗鎖定 30 分鐘(失敗計數存 Valkey,不區分帳號是否存在)。
func (h *AuthHandler) Login(ctx context.Context, req *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error) {
	customerCode := strings.TrimSpace(req.Msg.GetCustomerCode())
	password := req.Msg.GetPassword()
	if customerCode == "" || password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("客戶編號與密碼不可為空"))
	}

	unlockAt, err := h.deps.Lockout.LockedUntil(ctx, customerCode)
	if err != nil {
		return nil, internal(err)
	}
	if !unlockAt.IsZero() {
		return nil, errcode.AuthLocked.Error(map[string]string{"until": unlockAt.Format("15:04")})
	}

	var u *ent.User
	if err := h.systemScope(ctx, func(ctx context.Context) error {
		var qerr error
		u, qerr = dbtenant.Client(ctx, h.deps.DB).User.Query().
			Where(user.AccountNameEQ(customerCode), user.IsCustomerEQ(true)).WithCompany().Only(ctx)
		return qerr
	}); err != nil {
		if !ent.IsNotFound(err) {
			return nil, internal(err)
		}
		h.recordFailure(ctx, customerCode)
		return nil, invalidCredentials()
	}
	// A2 公司停用連鎖(2.1.3):公司非 active → permission_denied(AUTH-4002),不核發憑證、不計失敗。
	// 軟刪除的公司(P2-A)一併視同停用:已刪租戶不得再換發憑證。
	if u.Edges.Company != nil && (u.Edges.Company.Status != company.StatusActive || u.Edges.Company.DeletedAt != nil) {
		return nil, errcode.AuthCompanyInactive.Error(nil)
	}
	if u.Status != user.StatusActive {
		h.recordFailure(ctx, customerCode)
		return nil, invalidCredentials()
	}
	if !auth.VerifyPassword(u.PasswordHash, password) {
		h.recordFailure(ctx, customerCode)
		return nil, invalidCredentials()
	}
	if err := h.deps.Lockout.Clear(ctx, customerCode); err != nil {
		return nil, internal(err)
	}
	return h.issueTokenPair(ctx, u)
}

// Refresh 以 refresh token 旋轉換發新 token 對(T13):驗證 token 存在、使用者 token_version
// 未變(撤銷比對)、帳號仍 active,舊 refresh 作廢並發新。
func (h *AuthHandler) Refresh(ctx context.Context, req *connect.Request[v1.RefreshRequest]) (*connect.Response[v1.RefreshResponse], error) {
	plain := strings.TrimSpace(req.Msg.GetRefreshToken())
	if plain == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("refresh token 不可為空"))
	}
	uid, pinnedTV, err := h.deps.Tokens.VerifyRefresh(ctx, plain)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("refresh token 無效"))
	}
	liveTV, err := h.deps.Tokens.CurrentTokenVersion(ctx, uid)
	if err != nil {
		return nil, internal(err)
	}
	if liveTV != pinnedTV {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("refresh token 已撤銷,請重新登入"))
	}
	var u *ent.User
	if err := h.systemScope(ctx, func(ctx context.Context) error {
		var qerr error
		u, qerr = dbtenant.Client(ctx, h.deps.DB).User.Get(ctx, uid)
		return qerr
	}); err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("帳號不存在或已停用"))
	}
	if u.Status != user.StatusActive {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("帳號不存在或已停用"))
	}

	subject, err := h.subjectFromUser(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	access, err := h.deps.Tokens.IssueAccess(ctx, subject)
	if err != nil {
		return nil, internal(err)
	}
	newRefresh, err := h.deps.Tokens.RotateRefresh(ctx, plain, uid, pinnedTV)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("refresh token 無效"))
	}
	return connect.NewResponse(&v1.RefreshResponse{
		AccessToken:  access,
		RefreshToken: newRefresh,
		ExpiresIn:    int64(auth.AccessTokenTTL / time.Second),
	}), nil
}

// Logout 撤銷 refresh token 並清除 Web session(冪等)。
func (h *AuthHandler) Logout(ctx context.Context, req *connect.Request[v1.LogoutRequest]) (*connect.Response[v1.LogoutResponse], error) {
	if plain := strings.TrimSpace(req.Msg.GetRefreshToken()); plain != "" {
		if err := h.deps.Tokens.RevokeRefresh(ctx, plain); err != nil {
			return nil, internal(err)
		}
	}
	if err := h.deps.Sessions.Destroy(ctx); err != nil {
		return nil, internal(err)
	}
	return connect.NewResponse(&v1.LogoutResponse{}), nil
}

// RegisterComplete 完成員工註冊(T17):guest 選公司、填姓名後轉 pending(待審核)。
// 身分來源依序:registration token(首次 OIDC 未建帳號 → 建立 guest)→ Web session → Bearer JWT(更新既有 guest)。
func (h *AuthHandler) RegisterComplete(ctx context.Context, req *connect.Request[v1.RegisterCompleteRequest]) (*connect.Response[v1.RegisterCompleteResponse], error) {
	name := strings.TrimSpace(req.Msg.GetName())
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("姓名不可為空"))
	}
	companyID, err := parseID(req.Msg.GetCompanyId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("公司 ID 格式錯誤"))
	}
	// 軟刪除(P2-A):已刪除的公司不得再作為註冊/登入的租戶。
	var co *ent.Company
	if err := h.systemScope(ctx, func(ctx context.Context) error {
		var qerr error
		co, qerr = dbtenant.Client(ctx, h.deps.DB).Company.Query().
			Where(company.ID(companyID), company.DeletedAtIsNil()).Only(ctx)
		return qerr
	}); err != nil {
		if ent.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("公司不存在"))
		}
		return nil, internal(err)
	}
	if co.Status != company.StatusActive {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("公司未啟用"))
	}

	// 路徑 1:registration token(首次 OIDC 登入、hd 未對應公司時由 callback 派發)
	if token := registrationTokenFromRequest(req); token != "" {
		return h.registerWithToken(ctx, token, name, companyID)
	}

	// 路徑 2/3:既有 guest(Web session 或 App Bearer JWT)更新公司與姓名、轉 pending
	uid := auth.SessionUserID(ctx, h.deps.Sessions)
	if uid == 0 {
		claims, err := h.authenticateBearer(ctx, req.Header().Get("Authorization"))
		if err != nil {
			return nil, err
		}
		uid = claims.UserID
	}
	return h.completeGuest(ctx, uid, name, companyID)
}

// QRLogin 於後續計畫實作(QR 兌換前段);本 wave 回 Unimplemented。
// (內嵌 UnimplementedAuthServiceHandler 已覆蓋。)

// ---------------------------------------------------------------------------
// OIDC 公開端點

// GoogleLogin 導向 Google 授權頁(T11):產生一次性 state 存 Valkey,回跳前端登入。
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.deps.OAuth == nil || h.deps.Exchanger == nil || h.deps.Verifier == nil {
		http.Error(w, "OIDC 未設定", http.StatusServiceUnavailable)
		return
	}
	state, err := auth.NewState()
	if err != nil {
		http.Error(w, "伺服器錯誤", http.StatusInternalServerError)
		return
	}
	if err := h.deps.OneTime.Put(r.Context(), auth.StateKey(state), "1", auth.StateTTL); err != nil {
		http.Error(w, "伺服器錯誤", http.StatusInternalServerError)
		return
	}
	opts := []oauth2.AuthCodeOption{}
	if hd := strings.TrimSpace(h.deps.Cfg.Auth.GoogleHostedDomain); hd != "" {
		opts = append(opts, oauth2.SetAuthURLParam("hd", hd))
	}
	http.Redirect(w, r, h.deps.OAuth.AuthCodeURL(state, opts...), http.StatusFound)
}

// GoogleCallback 處理 Google 回調(T11):驗 state、換授權碼、驗 ID token;
// 依 email find-or-create User(role=guest),Web 設 session cookie、App 回 JWT。
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	code, state := q.Get("code"), q.Get("state")
	client := q.Get("client") // web(預設)| app
	if code == "" || state == "" {
		h.redirectError(w, r, "invalid_state")
		return
	}
	if _, ok, err := h.deps.OneTime.GetAndDelete(r.Context(), auth.StateKey(state)); err != nil || !ok {
		h.redirectError(w, r, "invalid_state")
		return
	}
	rawIDToken, err := h.deps.Exchanger.Exchange(r.Context(), code)
	if err != nil {
		h.redirectError(w, r, "oauth_failed")
		return
	}
	id, err := h.deps.Verifier.VerifyIDToken(r.Context(), rawIDToken)
	if err != nil {
		h.redirectError(w, r, "oauth_failed")
		return
	}

	u, err := h.userByEmail(r.Context(), id.Email)
	switch {
	case err == nil:
		switch u.Status {
		case user.StatusPending:
			// 待審核:回跳等待頁,不發憑證
			http.Redirect(w, r, h.deps.Cfg.Auth.FrontendURL+"/register-complete?status=pending", http.StatusFound)
		case user.StatusInactive:
			h.redirectError(w, r, "account_inactive")
		default:
			h.completeLogin(w, r, u, client)
		}
		return
	case ent.IsNotFound(err):
		// 首次登入:hd 對應公司 → 建立 guest(role=guest,status=active);
		// 否則派發一次性 registration token,由 RegisterComplete 建帳號(選公司)。
		co, err := h.resolveCompanyByHD(r.Context(), id.HostedDomain)
		if err != nil {
			h.redirectError(w, r, "server_error")
			return
		}
		if co == nil {
			h.issueRegistration(w, r, id, client)
			return
		}
		// 席位守衛：這條路徑會新增 users 列（＝佔一個席位），且不需要管理權。
		if gerr := h.guardSeats(r.Context(), co.ID); gerr != nil {
			h.redirectError(w, r, errCodeID(gerr))
			return
		}
		name := id.Name
		if name == "" {
			name = emailLocalPart(id.Email)
		}
		var created *ent.User
		if err := h.systemScope(r.Context(), func(ctx context.Context) error {
			var cerr error
			created, cerr = dbtenant.Client(ctx, h.deps.DB).User.Create().
				SetEmail(id.Email).
				SetName(name).
				SetStatus(user.StatusActive).
				SetRole(RoleGuest).
				SetIsCustomer(false).
				SetPasswordHash(auth.OIDCPasswordSentinel).
				SetCompanyID(co.ID).
				Save(ctx)
			return cerr
		}); err != nil {
			h.redirectError(w, r, "server_error")
			return
		}
		h.completeLogin(w, r, created, client)
		return
	default:
		h.redirectError(w, r, "server_error")
	}
}

// ---------------------------------------------------------------------------
// 內部輔助

// issueTokenPair 為成功登入核發 access + refresh token 對。
func (h *AuthHandler) issueTokenPair(ctx context.Context, u *ent.User) (*connect.Response[v1.LoginResponse], error) {
	subject, err := h.subjectFromUser(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	access, err := h.deps.Tokens.IssueAccess(ctx, subject)
	if err != nil {
		return nil, internal(err)
	}
	tv, err := h.deps.Tokens.CurrentTokenVersion(ctx, u.ID)
	if err != nil {
		return nil, internal(err)
	}
	refresh, err := h.deps.Tokens.IssueRefresh(ctx, u.ID, tv)
	if err != nil {
		return nil, internal(err)
	}
	return connect.NewResponse(&v1.LoginResponse{
		AccessToken:        access,
		RefreshToken:       refresh,
		ExpiresIn:          int64(auth.AccessTokenTTL / time.Second),
		MustChangePassword: u.MustChangePassword, // A3 1.5.2:前端導向改密碼頁
	}), nil
}

// subjectFromUser 由**使用者 ID**組裝 token subject(company / department 於**自己的系統範圍交易**
// 內載入)。刻意不吃呼叫端的 entity:憑證簽發路徑(登入/OIDC/refresh)都還沒有租戶 scope,而那些
// entity 是別的(已提交的)交易載入的 —— 對它做 lazy edge 查詢會落在已結束的交易上。
func (h *AuthHandler) subjectFromUser(ctx context.Context, userID int) (auth.TokenSubject, error) {
	var s auth.TokenSubject
	err := h.systemScope(ctx, func(ctx context.Context) error {
		u, err := dbtenant.Client(ctx, h.deps.DB).User.Query().
			WithCompany().WithDepartment().Where(user.ID(userID)).Only(ctx)
		if err != nil {
			return err
		}
		s = auth.TokenSubject{UserID: u.ID, Role: u.Role, MustChangePassword: u.MustChangePassword}
		if u.Edges.Company != nil {
			s.CompanyID = u.Edges.Company.ID
		}
		if u.Edges.Department != nil {
			s.DepartmentID = u.Edges.Department.ID
		}
		return nil
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return auth.TokenSubject{}, connect.NewError(connect.CodeFailedPrecondition, errors.New("使用者未歸屬公司"))
		}
		return auth.TokenSubject{}, internal(err)
	}
	if s.CompanyID == 0 {
		return auth.TokenSubject{}, connect.NewError(connect.CodeFailedPrecondition, errors.New("使用者未歸屬公司"))
	}
	return s, nil
}

// completeLogin 於 callback 登入成功後核發憑證:web 設 session cookie,app 於回跳 URL 帶 JWT。
func (h *AuthHandler) completeLogin(w http.ResponseWriter, r *http.Request, u *ent.User, client string) {
	ctx := r.Context()
	subject, err := h.subjectFromUser(ctx, u.ID)
	if err != nil {
		h.redirectError(w, r, "server_error")
		return
	}
	access, err := h.deps.Tokens.IssueAccess(ctx, subject)
	if err != nil {
		h.redirectError(w, r, "server_error")
		return
	}
	tv, err := h.deps.Tokens.CurrentTokenVersion(ctx, u.ID)
	if err != nil {
		h.redirectError(w, r, "server_error")
		return
	}
	refresh, err := h.deps.Tokens.IssueRefresh(ctx, u.ID, tv)
	if err != nil {
		h.redirectError(w, r, "server_error")
		return
	}
	if client == "app" {
		target := fmt.Sprintf("%s/auth/callback?access_token=%s&refresh_token=%s&expires_in=%d",
			h.deps.Cfg.Auth.FrontendURL, url.QueryEscape(access), url.QueryEscape(refresh),
			int64(auth.AccessTokenTTL/time.Second))
		http.Redirect(w, r, target, http.StatusFound)
		return
	}
	auth.EstablishWebSession(ctx, h.deps.Sessions, u.ID, u.Role, tv)
	http.Redirect(w, r, h.deps.Cfg.Auth.FrontendURL+"/", http.StatusFound)
}

// issueRegistration 首次登入且 hd 無對應公司:派發一次性 registration token(cookie + app query)。
func (h *AuthHandler) issueRegistration(w http.ResponseWriter, r *http.Request, id *auth.OIDCIdentity, client string) {
	token, err := auth.NewRegistrationToken()
	if err != nil {
		h.redirectError(w, r, "server_error")
		return
	}
	if err := h.deps.OneTime.Put(r.Context(), auth.RegistrationKey(token), id.Email, auth.RegistrationTokenTTL); err != nil {
		h.redirectError(w, r, "server_error")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     RegistrationTokenCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.deps.Cfg.Auth.SessionSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(auth.RegistrationTokenTTL.Seconds()),
	})
	target := h.deps.Cfg.Auth.FrontendURL + "/register-complete"
	if client == "app" {
		target += "?registration_token=" + url.QueryEscape(token)
	}
	http.Redirect(w, r, target, http.StatusFound)
}

// registerWithToken 以 registration token 建立 guest(role=guest,status=pending,歸屬所選公司)。
func (h *AuthHandler) registerWithToken(ctx context.Context, token, name string, companyID int) (*connect.Response[v1.RegisterCompleteResponse], error) {
	email, ok, err := h.deps.OneTime.GetAndDelete(ctx, auth.RegistrationKey(token))
	if err != nil {
		return nil, internal(err)
	}
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("註冊憑證無效或已過期,請重新以 Google 登入"))
	}
	// 席位守衛：這條路徑會新增 users 列（＝佔一個席位），公司由請求帶入但已在上游驗證
	// （RegisterComplete 先確認公司存在、未軟刪除且啟用）；此處仍不採身分推導（無身分）。
	if err := h.guardSeats(ctx, companyID); err != nil {
		return nil, err
	}
	// 去重與建帳號同一條系統交易(未登入路徑;users 受 RLS 約束,見 systemScope)。
	if err := h.systemScope(ctx, func(ctx context.Context) error {
		db := dbtenant.Client(ctx, h.deps.DB)
		exists, qerr := db.User.Query().Where(user.EmailEQ(email)).Exist(ctx)
		if qerr != nil {
			return internal(qerr)
		}
		if exists {
			// 已知的識別碼（email）重複 → SYS-2001；email 放 details（訊息樣板不含參數）。
			return errcode.SysConflict.Error(map[string]string{"email": email})
		}
		if _, cerr := db.User.Create().
			SetEmail(email).
			SetName(name).
			SetStatus(user.StatusPending).
			SetRole(RoleGuest).
			SetIsCustomer(false).
			SetPasswordHash(auth.OIDCPasswordSentinel).
			SetCompanyID(companyID).
			Save(ctx); cerr != nil {
			return internal(cerr)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.RegisterCompleteResponse{}), nil
}

// completeGuest 更新既有 guest:設定公司與姓名、狀態轉 pending,並使既有 token 失效。
func (h *AuthHandler) completeGuest(ctx context.Context, uid int, name string, companyID int) (*connect.Response[v1.RegisterCompleteResponse], error) {
	var resp *connect.Response[v1.RegisterCompleteResponse]
	err := h.systemScope(ctx, func(ctx context.Context) error {
		db := dbtenant.Client(ctx, h.deps.DB)
		u, err := db.User.Get(ctx, uid)
		if err != nil {
			if ent.IsNotFound(err) {
				return connect.NewError(connect.CodeUnauthenticated, errors.New("帳號不存在"))
			}
			return internal(err)
		}
		if u.Role != RoleGuest {
			return connect.NewError(connect.CodeFailedPrecondition, errors.New("僅 guest 帳號可完成註冊"))
		}
		// 身分異動(公司歸屬、狀態)後既有 access/refresh 立即失效(D5):token_version+1 併入
		// **同一個敘述**。為何不呼叫 Tokens.BumpTokenVersion:那會另開一條交易 UPDATE 同一列,
		// 而本交易的列鎖尚未放開 → 兩條交易互等,死結(「同一請求內對同一列開第二條交易」是禁例)。
		if _, err := db.User.UpdateOneID(uid).
			SetName(name).
			SetStatus(user.StatusPending).
			SetCompanyID(companyID).
			AddTokenVersion(1).
			Save(ctx); err != nil {
			return internal(err)
		}
		resp = connect.NewResponse(&v1.RegisterCompleteResponse{})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// authenticateBearer 解析並驗證 App Bearer JWT。
func (h *AuthHandler) authenticateBearer(ctx context.Context, authorization string) (*auth.Claims, error) {
	raw := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
	if raw == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
	}
	claims, err := h.deps.Tokens.VerifyAccess(ctx, raw)
	if err != nil {
		if errors.Is(err, auth.ErrTokenRevoked) || errors.Is(err, auth.ErrInvalidToken) {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("登入已失效,請重新登入"))
		}
		return nil, internal(err)
	}
	return claims, nil
}

// resolveCompanyByHD 依 Google Workspace 網域(hd)對應 companies.identifier 解析所屬公司。
// 已軟刪除(P2-A)或非 active 的公司不可解析:被刪公司的識別碼可被新公司重用,若不過濾
// deleted_at 就會解析到舊的幽靈公司。
func (h *AuthHandler) resolveCompanyByHD(ctx context.Context, hd string) (*ent.Company, error) {
	if hd == "" {
		return nil, nil
	}
	var co *ent.Company
	err := h.systemScope(ctx, func(ctx context.Context) error {
		var qerr error
		co, qerr = dbtenant.Client(ctx, h.deps.DB).Company.Query().
			Where(company.IdentifierEQ(hd), company.StatusEQ(company.StatusActive), company.DeletedAtIsNil()).
			Only(ctx)
		return qerr
	})
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return co, nil
}

// recordFailure 記錄登入失敗(失敗計數故障不阻斷錯誤回應)。
func (h *AuthHandler) recordFailure(ctx context.Context, customerCode string) {
	_, _ = h.deps.Lockout.RecordFailure(ctx, customerCode)
}

// redirectError 回跳登入頁並帶錯誤碼。
func (h *AuthHandler) redirectError(w http.ResponseWriter, r *http.Request, code string) {
	http.Redirect(w, r, h.deps.Cfg.Auth.FrontendURL+"/login?error="+code, http.StatusFound)
}

// registrationTokenFromRequest 自 cookie 或 X-Registration-Token header 讀取 registration token。
func registrationTokenFromRequest(req *connect.Request[v1.RegisterCompleteRequest]) string {
	if c, err := (&http.Request{Header: req.Header()}).Cookie(RegistrationTokenCookie); err == nil && c.Value != "" {
		return c.Value
	}
	return strings.TrimSpace(req.Header().Get("X-Registration-Token"))
}

// parseID 將字串 ID 轉為 ent 自增 int ID。
func parseID(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("空 ID")
	}
	return strconv.Atoi(s)
}

func emailLocalPart(email string) string {
	if i := strings.IndexByte(email, '@'); i > 0 {
		return email[:i]
	}
	return email
}

// invalidCredentials 為登入／改密碼的憑證錯誤:AUTH-4003(對外仍 Unauthenticated)。
// 刻意不區分「帳號不存在」與「密碼錯誤」(防帳號列舉),故所有呼叫點共用此碼。
func invalidCredentials() error {
	return errcode.AuthBadCredentials.Error(nil)
}

func internal(err error) error {
	return connect.NewError(connect.CodeInternal, err)
}
