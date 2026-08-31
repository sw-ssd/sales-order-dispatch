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
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

const (
	// RoleGuest 為首次 OIDC 登入的員工角色(待審核 / 待完成註冊)。
	RoleGuest = "guest"

	// RegistrationTokenCookie 為首次 OIDC 登入(尚未建帳號)時派發的一次性註冊憑證 cookie。
	RegistrationTokenCookie = "registration_token"
)

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

// ---------------------------------------------------------------------------
// AuthService RPC:Refresh / Logout(T13)、RegisterComplete(T17) 於後續任務補上;
// 現階段由內嵌 UnimplementedAuthServiceHandler 回 CodeUnimplemented。

// Login 客戶密碼登入(T12):以 customer_code(= users.account_name)查客戶帳號,
// bcrypt 驗證密碼;連續 5 次失敗鎖定 30 分鐘(失敗計數存 Valkey,不區分帳號是否存在)。
func (h *AuthHandler) Login(ctx context.Context, req *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error) {
	customerCode := strings.TrimSpace(req.Msg.GetCustomerCode())
	password := req.Msg.GetPassword()
	if customerCode == "" || password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("客戶編號與密碼不可為空"))
	}

	locked, err := h.deps.Lockout.IsLocked(ctx, customerCode)
	if err != nil {
		return nil, internal(err)
	}
	if locked {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("帳號已鎖定,請 30 分鐘後再試"))
	}

	u, err := h.deps.DB.User.Query().Where(user.AccountNameEQ(customerCode), user.IsCustomerEQ(true)).Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return nil, internal(err)
		}
		h.recordFailure(ctx, customerCode)
		return nil, invalidCredentials()
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

	u, err := h.deps.DB.User.Query().Where(user.EmailEQ(id.Email)).Only(r.Context())
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
		name := id.Name
		if name == "" {
			name = emailLocalPart(id.Email)
		}
		created, err := h.deps.DB.User.Create().
			SetEmail(id.Email).
			SetName(name).
			SetStatus(user.StatusActive).
			SetRole(RoleGuest).
			SetIsCustomer(false).
			SetPasswordHash(auth.OIDCPasswordSentinel).
			SetCompanyID(co.ID).
			Save(r.Context())
		if err != nil {
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
	subject, err := h.subjectFromUser(ctx, u)
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
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(auth.AccessTokenTTL / time.Second),
	}), nil
}

// subjectFromUser 由使用者組裝 token subject(company / department 自 edges 讀取)。
func (h *AuthHandler) subjectFromUser(ctx context.Context, u *ent.User) (auth.TokenSubject, error) {
	co, err := u.QueryCompany().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return auth.TokenSubject{}, connect.NewError(connect.CodeFailedPrecondition, errors.New("使用者未歸屬公司"))
		}
		return auth.TokenSubject{}, internal(err)
	}
	s := auth.TokenSubject{UserID: u.ID, CompanyID: co.ID, Role: u.Role}
	dep, err := u.QueryDepartment().Only(ctx)
	if err == nil {
		s.DepartmentID = dep.ID
	} else if !ent.IsNotFound(err) {
		return auth.TokenSubject{}, internal(err)
	}
	return s, nil
}

// completeLogin 於 callback 登入成功後核發憑證:web 設 session cookie,app 於回跳 URL 帶 JWT。
func (h *AuthHandler) completeLogin(w http.ResponseWriter, r *http.Request, u *ent.User, client string) {
	ctx := r.Context()
	subject, err := h.subjectFromUser(ctx, u)
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
	auth.EstablishWebSession(ctx, h.deps.Sessions, u.ID, u.Role)
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

// resolveCompanyByHD 依 Google Workspace 網域(hd)對應 companies.identifier 解析所屬公司。
func (h *AuthHandler) resolveCompanyByHD(ctx context.Context, hd string) (*ent.Company, error) {
	if hd == "" {
		return nil, nil
	}
	co, err := h.deps.DB.Company.Query().Where(company.IdentifierEQ(hd), company.StatusEQ(company.StatusActive)).Only(ctx)
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

func invalidCredentials() error {
	return connect.NewError(connect.CodeUnauthenticated, errors.New("客戶編號或密碼錯誤"))
}

func internal(err error) error {
	return connect.NewError(connect.CodeInternal, err)
}
