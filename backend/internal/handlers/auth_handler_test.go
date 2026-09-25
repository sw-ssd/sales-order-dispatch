package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// fakeVerifier 測試用 OIDCVerifier。
type fakeVerifier struct {
	id  *auth.OIDCIdentity
	err error
}

func (f *fakeVerifier) VerifyIDToken(_ context.Context, _ string) (*auth.OIDCIdentity, error) {
	return f.id, f.err
}

// fakeExchanger 測試用 OAuthExchanger。
type fakeExchanger struct {
	raw string
	err error
}

func (f *fakeExchanger) Exchange(_ context.Context, _ string) (string, error) {
	return f.raw, f.err
}

// fakeSeatGuard 為席位守衛的測試替身：記錄以哪個公司／feature 查過，並依設定回錯
// （nil＝未達上限）。用替身而非真品，是因為本套件要驗的是「**有沒有在建立 users 列之前查**」
// 與「被擋時回哪個碼」，計數正確性屬 internal/services 的測試範圍。
type fakeSeatGuard struct {
	calls []string // "companyID feature delta"
	err   error
}

func (f *fakeSeatGuard) CheckLimit(_ context.Context, companyID int, feature string, delta int) error {
	f.calls = append(f.calls, fmt.Sprintf("%d %s %d", companyID, feature, delta))
	return f.err
}

// seatLimitExceeded 為「已達席位上限」的守衛回應（PLAT-5001，帶 details；與 CheckLimit 同碼）。
func seatLimitExceeded() error {
	return errcode.PlatformLimitExceeded.Error(map[string]string{
		"feature": entitlements.LimitSeats, "used": "10", "limit": "10",
	})
}

// testEnv 組裝 handlers 測試環境(enttest sqlite + MemoryStore + fake OIDC)。
type testEnv struct {
	ctx       context.Context
	db        *ent.Client
	kv        *auth.MemoryStore
	tokens    *auth.TokenManager
	handler   *AuthHandler
	sessions  *scs.SessionManager
	verifier  *fakeVerifier
	exch      *fakeExchanger
	rpc       salesorderv1connect.AuthServiceClient
	seatGuard *fakeSeatGuard
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	db := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })

	kv := auth.NewMemoryStore()
	tokens := auth.NewTokenManager("test-secret", kv, db)
	lockout := auth.NewLoginLock(kv)
	oneTime := auth.NewOneTimeStore(kv)
	sessions := auth.WebSessionManager(memstore.New(), 30*24*time.Hour, false, "lax")

	cfg := &config.Config{}
	cfg.Auth.FrontendURL = "http://localhost:3000"

	verifier := &fakeVerifier{id: &auth.OIDCIdentity{Email: "emp@example.com", Name: "張三", HostedDomain: "example.com"}}
	exch := &fakeExchanger{raw: "fake-id-token"}
	seatGuard := &fakeSeatGuard{}
	h := NewAuthHandler(AuthDeps{
		Cfg: cfg, DB: db, Tokens: tokens, Lockout: lockout, OneTime: oneTime, Sessions: sessions,
		Entitlements: seatGuard,
	})
	h.SetOIDC(auth.NewGoogleOAuthConfig("cid", "csec", "http://localhost:3080/api/v1/auth/google/callback"), exch, verifier)

	// Connect RPC client(掛載 handler + scs 中介層,支援 Web session cookie)
	path, connectHandler := salesorderv1connect.NewAuthServiceHandler(h)
	mux := http.NewServeMux()
	mux.Handle(path, sessions.LoadAndSave(connectHandler))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	rpc := salesorderv1connect.NewAuthServiceClient(http.DefaultClient, ts.URL)

	return &testEnv{
		ctx: context.Background(), db: db, kv: kv, tokens: tokens,
		handler: h, sessions: sessions, verifier: verifier, exch: exch, rpc: rpc,
		seatGuard: seatGuard,
	}
}

// seedWebSession 建立一個已登入的 Web session cookie(經 LoadAndSave 中介層寫入)。
// token_version 取 DB 現值,與完整登入流程一致(供 authzMiddleware 驗證比對)。
func (e *testEnv) seedWebSession(t *testing.T, userID int, role string) string {
	t.Helper()
	tv, err := e.tokens.CurrentTokenVersion(e.ctx, userID)
	if err != nil {
		t.Fatalf("seedWebSession: CurrentTokenVersion: %v", err)
	}
	seed := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth.EstablishWebSession(r.Context(), e.sessions, userID, role, tv)
		w.WriteHeader(http.StatusNoContent)
	})
	rec := httptest.NewRecorder()
	e.sessions.LoadAndSave(seed).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/seed", nil))
	for _, c := range rec.Result().Cookies() {
		if c.Name == auth.SessionCookieName {
			return c.Value
		}
	}
	t.Fatal("seedWebSession: 未取得 session cookie")
	return ""
}

// callback 以測試替身走完整 OIDC callback(前置 state 已寫入)。
func (e *testEnv) callback(t *testing.T, state string, client string) *httptest.ResponseRecorder {
	t.Helper()
	url := "/api/v1/auth/google/callback?code=testcode&state=" + state
	if client != "" {
		url += "&client=" + client
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	e.sessions.LoadAndSave(http.HandlerFunc(e.handler.GoogleCallback)).ServeHTTP(rec, req)
	return rec
}

func mustCreateCompany(t *testing.T, e *testEnv, identifier string) int {
	t.Helper()
	co, err := e.db.Company.Create().SetName("公司-" + identifier).SetIdentifier(identifier).Save(e.ctx)
	if err != nil {
		t.Fatalf("建公司: %v", err)
	}
	return co.ID
}
func TestLoginSuccessWrongPasswordLockout(t *testing.T) {
	// T12 驗收:正確密碼、錯誤密碼、鎖定。
	e := newTestEnv(t)
	coID := mustCreateCompany(t, e, "co-a")
	hash, err := auth.HashPassword("secret-123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	e.db.User.Create().
		SetEmail("cust@example.com").SetName("店家甲").SetStatus(user.StatusActive).
		SetRole("customer").SetIsCustomer(true).SetAccountName("C001").SetPasswordHash(hash).
		SetCompanyID(coID).SaveX(e.ctx)

	// 正確密碼
	resp, err := e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "C001", Password: "secret-123"}))
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if resp.Msg.GetAccessToken() == "" || resp.Msg.GetRefreshToken() == "" {
		t.Fatal("應核發 access + refresh token")
	}
	if resp.Msg.GetExpiresIn() != int64(auth.AccessTokenTTL/time.Second) {
		t.Fatalf("expires_in 應為 3600,got %d", resp.Msg.GetExpiresIn())
	}
	// JWT 可驗證且 claim 正確
	claims, err := e.tokens.VerifyAccess(e.ctx, resp.Msg.GetAccessToken())
	if err != nil {
		t.Fatalf("access token 驗證: %v", err)
	}
	if claims.Role != "customer" || claims.CompanyID != coID {
		t.Fatalf("claims 不符: %+v", claims)
	}

	// 錯誤密碼 → Unauthenticated(不透露)
	if _, err := e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "C001", Password: "wrong"})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("錯誤密碼應 Unauthenticated,got %v", err)
	}
	// 不存在的帳號 → Unauthenticated(不透露帳號存在與否)
	if _, err := e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "NO-SUCH", Password: "x"})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("不存在帳號應 Unauthenticated,got %v", err)
	}

	// 再失敗 3 次(累計 4 次)
	for range 3 {
		_, _ = e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "C001", Password: "wrong"}))
	}
	// 第 5 次失敗(此後 count=5)
	if _, err := e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "C001", Password: "wrong"})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("第 5 次失敗應 Unauthenticated,got %v", err)
	}
	// 鎖定:即使密碼正確也拒絕
	if _, err := e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "C001", Password: "secret-123"})); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("鎖定期間應 FailedPrecondition,got %v", err)
	}
}

// TestLoginExpiredTempPasswordRejected A3(1.5.2)spec「臨時密碼超過 24 小時失效」:
// 密碼本身正確(must_change=true 且已過 temp_password_expires_at)→ **登入階段**即拒
// (AUTH-3002),不是登入後才在改密碼頁被擋。須由 dept_admin 以上重置後才能再登入。
func TestLoginExpiredTempPasswordRejected(t *testing.T) {
	e := newTestEnv(t)
	coID := mustCreateCompany(t, e, "co-exp")
	hash, err := auth.HashPassword("temp-123456")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	u := e.db.User.Create().
		SetEmail("exp@example.com").SetName("過期店家").SetStatus(user.StatusActive).
		SetRole("customer").SetIsCustomer(true).SetAccountName("EXP-1").SetPasswordHash(hash).
		SetCompanyID(coID).SetMustChangePassword(true).
		SetTempPasswordExpiresAt(time.Now().Add(-time.Hour)).SaveX(e.ctx)

	_, err = e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "EXP-1", Password: "temp-123456"}))
	if err == nil {
		t.Fatal("臨時密碼已過期仍登入成功 —— 應回 AUTH-3002")
	}
	// 碼而非僅 connect 碼:前端據 ErrorInfo.code 決定提示(connect 碼同為 failed_precondition,
	// 只斷言 connect 碼的話,AUTH-3002/AUTH-3004/AUTH-3003 三者互換也測不出來)。
	if code := authErrorInfo(t, err).GetCode(); code != errcode.AuthTempPasswordExpired.ID() {
		t.Fatalf("應回 %s,got %q(connect=%v)", errcode.AuthTempPasswordExpired.ID(), code, connect.CodeOf(err))
	}
	// 憑證正確不算失敗:不該讓一次正常登入把帳號推向鎖定。
	unlockAt, lerr := auth.NewLoginLock(e.kv).LockedUntil(e.ctx, "EXP-1")
	if lerr != nil {
		t.Fatalf("LockedUntil: %v", lerr)
	}
	if !unlockAt.IsZero() {
		t.Fatalf("憑證正確的過期登入不應計入鎖定,got %v", unlockAt)
	}

	// 重置後(新臨時密碼、效期重新起算)即可登入,回到受限態而非被鎖死。
	fresh, err := auth.GenerateTempPassword()
	if err != nil {
		t.Fatalf("GenerateTempPassword: %v", err)
	}
	freshHash, err := auth.HashPassword(fresh)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	e.db.User.UpdateOneID(u.ID).SetPasswordHash(freshHash).
		SetTempPasswordExpiresAt(time.Now().Add(24 * time.Hour)).SaveX(e.ctx)
	resp, err := e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "EXP-1", Password: fresh}))
	if err != nil {
		t.Fatalf("重置後應可登入: %v", err)
	}
	if !resp.Msg.GetMustChangePassword() {
		t.Fatal("重置後登入應仍為受限態(must_change_password=true)")
	}
}

func TestCallbackUnknownHDIssuesRegistrationToken(t *testing.T) {
	// hd 無對應公司:不建帳號,派發 registration token 並回跳註冊完成頁(1.4.3)。
	e := newTestEnv(t)
	e.verifier.id = &auth.OIDCIdentity{Email: "new@other.example", Name: "李四", HostedDomain: "other.example"}

	state := "state-2"
	if err := e.handler.deps.OneTime.Put(e.ctx, auth.StateKey(state), "1", time.Minute); err != nil {
		t.Fatalf("寫入 state: %v", err)
	}
	rec := e.callback(t, state, "")
	if rec.Code != http.StatusFound {
		t.Fatalf("callback 應 302,got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); !strings.HasSuffix(loc, "/register-complete") {
		t.Fatalf("應回跳註冊完成頁,got %q", loc)
	}
	// 未建帳號
	if n, err := e.db.User.Query().Where(user.EmailEQ("new@other.example")).Count(e.ctx); err != nil || n != 0 {
		t.Fatalf("hd 未對應公司時不得建帳號,count=%d err=%v", n, err)
	}
	// registration token cookie 已設定
	var regToken string
	for _, c := range rec.Result().Cookies() {
		if c.Name == RegistrationTokenCookie {
			regToken = c.Value
		}
	}
	if regToken == "" {
		t.Fatal("應設定 registration_token cookie")
	}

	// 完成註冊:選公司 + 姓名 → 建立 guest(status=pending)
	coID := mustCreateCompany(t, e, "co-b")
	req := connect.NewRequest(&v1.RegisterCompleteRequest{CompanyId: itoa(coID), Name: "李四"})
	req.Header().Set("Cookie", RegistrationTokenCookie+"="+regToken)
	if _, err := e.rpc.RegisterComplete(e.ctx, req); err != nil {
		t.Fatalf("RegisterComplete: %v", err)
	}
	u, err := e.db.User.Query().Where(user.EmailEQ("new@other.example")).Only(e.ctx)
	if err != nil {
		t.Fatalf("註冊完成後應有帳號: %v", err)
	}
	if u.Role != RoleGuest || u.Status != user.StatusPending || u.Name != "李四" {
		t.Fatalf("guest 應為 pending: role=%q status=%q name=%q", u.Role, u.Status, u.Name)
	}
	if co, err := u.QueryCompany().Only(e.ctx); err != nil || co.ID != coID {
		t.Fatalf("帳號應歸屬所選公司,got %v err=%v", co, err)
	}
}

func TestRegisterCompleteGuestToPending(t *testing.T) {
	// T17 驗收:guest 完成註冊 → 狀態更新為 pending。
	e := newTestEnv(t)
	coA := mustCreateCompany(t, e, "co-a")
	coB := mustCreateCompany(t, e, "co-b")

	guest := e.db.User.Create().
		SetEmail("guest@example.com").SetName("舊名").SetStatus(user.StatusActive).
		SetRole(RoleGuest).SetIsCustomer(false).SetPasswordHash(auth.OIDCPasswordSentinel).
		SetCompanyID(coA).SaveX(e.ctx)

	cookie := e.seedWebSession(t, guest.ID, RoleGuest)
	req := connect.NewRequest(&v1.RegisterCompleteRequest{CompanyId: itoa(coB), Name: "新名"})
	req.Header().Set("Cookie", auth.SessionCookieName+"="+cookie)

	if _, err := e.rpc.RegisterComplete(e.ctx, req); err != nil {
		t.Fatalf("RegisterComplete: %v", err)
	}
	after := e.db.User.GetX(e.ctx, guest.ID)
	if after.Status != user.StatusPending {
		t.Fatalf("guest 完成註冊後狀態應為 pending,got %q", after.Status)
	}
	if after.Name != "新名" {
		t.Fatalf("姓名應更新,got %q", after.Name)
	}
	if co, err := after.QueryCompany().Only(e.ctx); err != nil || co.ID != coB {
		t.Fatalf("公司應更新為所選公司,got %v err=%v", co, err)
	}
	if after.Role != RoleGuest {
		t.Fatalf("角色維持 guest,got %q", after.Role)
	}
	// 身分異動後 token_version 遞增(既有 token 失效)
	if tv, err := e.tokens.CurrentTokenVersion(e.ctx, guest.ID); err != nil || tv != 1 {
		t.Fatalf("token_version 應遞增為 1,got %d err=%v", tv, err)
	}

	// 非 guest 呼叫 → failed_precondition
	staff := e.db.User.Create().
		SetEmail("staff@example.com").SetName("王五").SetStatus(user.StatusActive).
		SetRole("staff").SetIsCustomer(false).SetPasswordHash(auth.OIDCPasswordSentinel).
		SetCompanyID(coA).SaveX(e.ctx)
	cookie2 := e.seedWebSession(t, staff.ID, "staff")
	req2 := connect.NewRequest(&v1.RegisterCompleteRequest{CompanyId: itoa(coA), Name: "王五"})
	req2.Header().Set("Cookie", auth.SessionCookieName+"="+cookie2)
	if _, err := e.rpc.RegisterComplete(e.ctx, req2); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("非 guest 應 failed_precondition,got %v", err)
	}
}

func TestRegisterCompleteRequiresAuth(t *testing.T) {
	// 無 session / 無 token / 無 registration token → Unauthenticated
	e := newTestEnv(t)
	coID := mustCreateCompany(t, e, "co-a")
	if _, err := e.rpc.RegisterComplete(e.ctx, connect.NewRequest(&v1.RegisterCompleteRequest{CompanyId: itoa(coID), Name: "某人"})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("未登入應 Unauthenticated,got %v", err)
	}
	// 無效公司 → InvalidArgument
	guest := e.db.User.Create().
		SetEmail("g@example.com").SetName("g").SetStatus(user.StatusActive).
		SetRole(RoleGuest).SetIsCustomer(false).SetPasswordHash(auth.OIDCPasswordSentinel).
		SetCompanyID(coID).SaveX(e.ctx)
	cookie := e.seedWebSession(t, guest.ID, RoleGuest)
	req := connect.NewRequest(&v1.RegisterCompleteRequest{CompanyId: "999999", Name: "某人"})
	req.Header().Set("Cookie", auth.SessionCookieName+"="+cookie)
	if _, err := e.rpc.RegisterComplete(e.ctx, req); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("公司不存在應 InvalidArgument,got %v", err)
	}
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
func TestRefreshRotation(t *testing.T) {
	// T13:refresh 旋轉換發;舊 token 重放被拒。
	e := newTestEnv(t)
	coID := mustCreateCompany(t, e, "co-a")
	hash, _ := auth.HashPassword("pw-123456")
	e.db.User.Create().
		SetEmail("cust2@example.com").SetName("店家乙").SetStatus(user.StatusActive).
		SetRole("customer").SetIsCustomer(true).SetAccountName("C002").SetPasswordHash(hash).
		SetCompanyID(coID).SaveX(e.ctx)

	login, err := e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "C002", Password: "pw-123456"}))
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	r1 := login.Msg.GetRefreshToken()

	ref, err := e.rpc.Refresh(e.ctx, connect.NewRequest(&v1.RefreshRequest{RefreshToken: r1}))
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if ref.Msg.GetAccessToken() == "" || ref.Msg.GetRefreshToken() == "" || ref.Msg.GetRefreshToken() == r1 {
		t.Fatal("Refresh 應旋轉發新 token 對")
	}
	r2 := ref.Msg.GetRefreshToken()

	// 舊 token 重放 → Unauthenticated
	if _, err := e.rpc.Refresh(e.ctx, connect.NewRequest(&v1.RefreshRequest{RefreshToken: r1})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("舊 refresh 重放應 Unauthenticated,got %v", err)
	}
	// 新 token 可繼續旋轉
	if _, err := e.rpc.Refresh(e.ctx, connect.NewRequest(&v1.RefreshRequest{RefreshToken: r2})); err != nil {
		t.Fatalf("新 refresh 應可旋轉: %v", err)
	}

	// Logout 撤銷後 → Unauthenticated
	login2, err := e.rpc.Login(e.ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "C002", Password: "pw-123456"}))
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if _, err := e.rpc.Logout(e.ctx, connect.NewRequest(&v1.LogoutRequest{RefreshToken: login2.Msg.GetRefreshToken()})); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if _, err := e.rpc.Refresh(e.ctx, connect.NewRequest(&v1.RefreshRequest{RefreshToken: login2.Msg.GetRefreshToken()})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("Logout 後 refresh 應失效,got %v", err)
	}
	// Logout 冪等
	if _, err := e.rpc.Logout(e.ctx, connect.NewRequest(&v1.LogoutRequest{RefreshToken: login2.Msg.GetRefreshToken()})); err != nil {
		t.Fatalf("Logout 應冪等: %v", err)
	}
}

// openAuthDB 開啟 enttest sqlite db(供 auth 測試)。
func openAuthDB(t *testing.T) *ent.Client {
	t.Helper()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// newIdentifiedAuthClientWithDB 以指定 db 建立注入身分 + 稽核來源的 AuthService client。
func newIdentifiedAuthClientWithDB(t *testing.T, db *ent.Client, id authz.Identity) salesorderv1connect.AuthServiceClient {
	t.Helper()
	kv := auth.NewMemoryStore()
	sessions := auth.WebSessionManager(memstore.New(), 30*24*time.Hour, false, "lax")
	h := NewAuthHandler(AuthDeps{
		DB: db, Tokens: auth.NewTokenManager("test-secret", kv, db),
		Lockout: auth.NewLoginLock(kv), OneTime: auth.NewOneTimeStore(kv), Sessions: sessions,
		// 本環境只驗密碼路徑（不建帳號）；席位守衛注入 Unlimited 等同「不受配額限制」，
		// 語意與注入前相同（AuthDeps.Entitlements 為 nil 時是 fail-closed：不注入會擋住建帳號）。
		Entitlements: entitlements.Unlimited(),
	})
	// 與生產一致(internal/server/domains.go:70)掛上 dbtenant.HandlerOption:請求層租戶交易是
	// A3 密碼路徑的必要條件(auth_password.go 以 dbtenant.TxFrom 取請求交易;缺它 → internal 錯誤)。
	path, handler := salesorderv1connect.NewAuthServiceHandler(h, dbtenant.HandlerOption(db))
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = audit.WithMeta(ctx, audit.Meta{IP: "1.2.3.4", UserAgent: "t"})
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return salesorderv1connect.NewAuthServiceClient(http.DefaultClient, ts.URL)
}

// seedAuthCompany 建立 company 回傳其真實 id(供 auth 測試 FK)。
func seedAuthCompany(t *testing.T, db *ent.Client) int {
	t.Helper()
	co, err := db.Company.Create().SetName("公司").SetIdentifier("T-" + t.Name()).Save(context.Background())
	if err != nil {
		t.Fatalf("seed company: %v", err)
	}
	return co.ID
}

// TestChangePasswordSuccess A3 1.5.2:改密碼成功 → hash 更新、must_change 清空、temp 效期清空、token_version+1、稽核存在。
func TestChangePasswordSuccess(t *testing.T) {
	ctx := context.Background()
	db := openAuthDB(t)
	cid := seedAuthCompany(t, db)
	old, _ := auth.HashPassword("OldPass123")
	u := db.User.Create().SetCompanyID(cid).SetEmail("cp@t.com").SetName("改密").SetRole("customer").SetIsCustomer(true).
		SetPasswordHash(old).SetMustChangePassword(true).SetTempPasswordExpiresAt(time.Now().Add(24 * time.Hour)).SaveX(ctx)
	client := newIdentifiedAuthClientWithDB(t, db, authz.Identity{UserID: strconv.Itoa(u.ID), CompanyID: strconv.Itoa(cid), Role: "customer", Roles: []string{"customer"}})
	if _, err := client.ChangePassword(ctx, connect.NewRequest(&v1.ChangePasswordRequest{OldPassword: "OldPass123", NewPassword: "NewPass123456"})); err != nil {
		t.Fatalf("ChangePassword 應成功: %v", err)
	}
	got := db.User.GetX(ctx, u.ID)
	if !auth.VerifyPassword(got.PasswordHash, "NewPass123456") {
		t.Fatal("新密碼應可驗證")
	}
	if got.MustChangePassword {
		t.Fatal("must_change_password 應清為 false")
	}
	if got.TempPasswordExpiresAt != nil {
		t.Fatal("temp_password_expires_at 應清空")
	}
	if got.TokenVersion != 1 {
		t.Fatalf("token_version 應 +1 為 1,得到 %d", got.TokenVersion)
	}
	if n, _ := db.AuditLog.Query().Count(ctx); n != 1 {
		t.Fatalf("改密碼應寫 1 筆稽核,得到 %d", n)
	}
}

// TestChangePasswordWrongOld:舊密碼錯誤 → unauthenticated,不動資料。
func TestChangePasswordWrongOld(t *testing.T) {
	ctx := context.Background()
	db := openAuthDB(t)
	cid := seedAuthCompany(t, db)
	old, _ := auth.HashPassword("OldPass123")
	u := db.User.Create().SetCompanyID(cid).SetEmail("cp2@t.com").SetName("甲").SetRole("customer").SetPasswordHash(old).SaveX(ctx)
	client := newIdentifiedAuthClientWithDB(t, db, authz.Identity{UserID: strconv.Itoa(u.ID), CompanyID: strconv.Itoa(cid), Role: "customer"})
	if _, err := client.ChangePassword(ctx, connect.NewRequest(&v1.ChangePasswordRequest{OldPassword: "WrongOld1", NewPassword: "NewPass123456"})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("舊密碼錯應 unauthenticated,got %v", err)
	}
	if got := db.User.GetX(ctx, u.ID); got.TokenVersion != 0 {
		t.Fatalf("失敗不應 bump token_version,得到 %d", got.TokenVersion)
	}
}

// TestChangePasswordTooShort:新密碼 < 8 → invalid_argument。
func TestChangePasswordTooShort(t *testing.T) {
	ctx := context.Background()
	db := openAuthDB(t)
	cid := seedAuthCompany(t, db)
	old, _ := auth.HashPassword("OldPass123")
	u := db.User.Create().SetCompanyID(cid).SetEmail("cp3@t.com").SetName("乙").SetRole("customer").SetPasswordHash(old).SaveX(ctx)
	client := newIdentifiedAuthClientWithDB(t, db, authz.Identity{UserID: strconv.Itoa(u.ID), CompanyID: strconv.Itoa(cid), Role: "customer"})
	if _, err := client.ChangePassword(ctx, connect.NewRequest(&v1.ChangePasswordRequest{OldPassword: "OldPass123", NewPassword: "short"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("新密碼過短應 invalid_argument,got %v", err)
	}
}

// seedCustomerUser 建立客戶帳號,回傳其 id / company id。
func seedCustomerUser(t *testing.T, db *ent.Client, cid int, email, name string, customer bool) int {
	t.Helper()
	u, err := db.User.Create().SetCompanyID(cid).SetEmail(email).SetName(name).SetRole("customer").SetIsCustomer(customer).
		SetPasswordHash("x").Save(context.Background())
	if err != nil {
		t.Fatalf("seed customer: %v", err)
	}
	return u.ID
}

// seedCompanyNamed 建立指定 identifier 的公司,回傳其 id。
func seedCompanyNamed(t *testing.T, db *ent.Client, ident string) int {
	t.Helper()
	co, err := db.Company.Create().SetName("公司-" + ident).SetIdentifier(ident).Save(context.Background())
	if err != nil {
		t.Fatalf("seed company named: %v", err)
	}
	return co.ID
}

// TestResetCustomerPasswordSuccess A3 1.5.4:super 重置客戶 → 回傳 ≥12 臨時密碼、舊 hash 替換、
// must_change=true、效期 24h、token_version+1、稽核存在。
func TestResetCustomerPasswordSuccess(t *testing.T) {
	ctx := context.Background()
	db := openAuthDB(t)
	cid := seedAuthCompany(t, db)
	target := seedCustomerUser(t, db, cid, "cust@t.com", "客戶", true)
	client := newIdentifiedAuthClientWithDB(t, db, authz.Identity{UserID: "1", CompanyID: strconv.Itoa(cid), Role: "super", Roles: []string{"super"}})
	resp, err := client.ResetCustomerPassword(ctx, connect.NewRequest(&v1.ResetCustomerPasswordRequest{UserId: strconv.Itoa(target)}))
	if err != nil {
		t.Fatalf("ResetCustomerPassword: %v", err)
	}
	if len(resp.Msg.GetTempPassword()) < 12 {
		t.Fatalf("臨時密碼應 ≥ 12 字元,得到 %d", len(resp.Msg.GetTempPassword()))
	}
	got := db.User.GetX(ctx, target)
	if !auth.VerifyPassword(got.PasswordHash, resp.Msg.GetTempPassword()) {
		t.Fatal("回傳的臨時密碼應可驗證")
	}
	if !got.MustChangePassword {
		t.Fatal("must_change_password 應為 true")
	}
	if got.TempPasswordExpiresAt == nil {
		t.Fatal("temp_password_expires_at 應已設定")
	}
	if got.TokenVersion != 1 {
		t.Fatalf("token_version 應 +1 為 1,得到 %d", got.TokenVersion)
	}
	if n, _ := db.AuditLog.Query().Count(ctx); n != 1 {
		t.Fatalf("重置應寫 1 筆稽核,得到 %d", n)
	}
}

// TestResetCustomerPasswordScopeDenied:staff 與跨公司 company_admin 皆被拒。
func TestResetCustomerPasswordScopeDenied(t *testing.T) {
	ctx := context.Background()
	db := openAuthDB(t)
	cidA := seedCompanyNamed(t, db, "co-a")
	cidB := seedCompanyNamed(t, db, "co-b")
	targetB := seedCustomerUser(t, db, cidB, "custb@t.com", "客戶B", true)

	// staff → permission_denied
	staffClient := newIdentifiedAuthClientWithDB(t, db, authz.Identity{UserID: "9", CompanyID: strconv.Itoa(cidA), Role: "staff", Roles: []string{"staff"}})
	if _, err := staffClient.ResetCustomerPassword(ctx, connect.NewRequest(&v1.ResetCustomerPasswordRequest{UserId: strconv.Itoa(targetB)})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("staff 重置應 permission_denied,got %v", err)
	}
	// company_admin(cidA) 重置 cidB 客戶 → permission_denied
	adminClient := newIdentifiedAuthClientWithDB(t, db, authz.Identity{UserID: "2", CompanyID: strconv.Itoa(cidA), Role: "company_admin", Roles: []string{"company_admin"}})
	if _, err := adminClient.ResetCustomerPassword(ctx, connect.NewRequest(&v1.ResetCustomerPasswordRequest{UserId: strconv.Itoa(targetB)})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("跨公司 company_admin 重置應 permission_denied,got %v", err)
	}
}

// TestCompanyDeactivationBlocksLogin A2(2.1.3):停用公司客戶登入 → permission_denied,不核發憑證;
// 恢復 active 後可正常登入。
func TestCompanyDeactivationBlocksLogin(t *testing.T) {
	e := newTestEnv(t)
	ctx := context.Background()
	co, err := e.db.Company.Create().SetName("停用測試公司").SetIdentifier("T-co-login").Save(ctx)
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	hash, _ := auth.HashPassword("pw-123456")
	if _, err := e.db.User.Create().SetCompanyID(co.ID).SetEmail("custacct@t.com").SetName("店家").SetRole("customer").SetIsCustomer(true).
		SetAccountName("ACC01").SetPasswordHash(hash).Save(ctx); err != nil {
		t.Fatalf("cust user: %v", err)
	}

	login := func() *connect.Response[v1.LoginResponse] {
		r, err := e.rpc.Login(ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "ACC01", Password: "pw-123456"}))
		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		return r
	}

	// 公司 active:可登入。
	if r := login(); r.Msg.GetAccessToken() == "" {
		t.Fatal("公司 active 時應核發 access token")
	}

	// 停用公司 → permission_denied。
	e.db.Company.UpdateOneID(co.ID).SetStatus(company.StatusInactive).SaveX(ctx)
	if _, err := e.rpc.Login(ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "ACC01", Password: "pw-123456"})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("停用公司登入應 permission_denied,got %v", err)
	}

	// 恢復 active → 可登入。
	e.db.Company.UpdateOneID(co.ID).SetStatus(company.StatusActive).SaveX(ctx)
	if r := login(); r.Msg.GetAccessToken() == "" {
		t.Fatal("恢復 active 後應可登入並核發 token")
	}
}

// TestSoftDeletedCompanyBlocksLogin P2-A 後續:公司被軟刪除(狀態仍 active)→ 客戶登入 permission_denied,
// 不核發憑證(復原才可再登入)。修復前登入只看 status,已刪租戶可繼續換發 token。
func TestSoftDeletedCompanyBlocksLogin(t *testing.T) {
	e := newTestEnv(t)
	ctx := context.Background()
	co, err := e.db.Company.Create().SetName("已刪測試公司").SetIdentifier("T-co-softdel").Save(ctx)
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	hash, _ := auth.HashPassword("pw-123456")
	if _, err := e.db.User.Create().SetCompanyID(co.ID).SetEmail("softdel@t.com").SetName("店家").SetRole("customer").SetIsCustomer(true).
		SetAccountName("ACC02").SetPasswordHash(hash).Save(ctx); err != nil {
		t.Fatalf("cust user: %v", err)
	}
	login := func() error {
		_, err := e.rpc.Login(ctx, connect.NewRequest(&v1.LoginRequest{CustomerCode: "ACC02", Password: "pw-123456"}))
		return err
	}
	// 前置:未刪除前可登入(證明後面的拒絕來自 deleted_at,不是別的失效原因)。
	if err := login(); err != nil {
		t.Fatalf("軟刪除前應可登入,got %v", err)
	}
	e.db.Company.UpdateOneID(co.ID).SetDeletedAt(time.Now().UTC()).SaveX(ctx)
	if got := connect.CodeOf(login()); got != connect.CodePermissionDenied {
		t.Fatalf("已刪除公司登入應 permission_denied,got %v", got)
	}
	// 復原 → 可再登入。
	e.db.Company.UpdateOneID(co.ID).ClearDeletedAt().SaveX(ctx)
	if err := login(); err != nil {
		t.Fatalf("復原後應可登入,got %v", err)
	}
}

// TestResetCustomerPasswordNonCustomer:目標非客戶帳號 → invalid_argument。
func TestResetCustomerPasswordNonCustomer(t *testing.T) {
	ctx := context.Background()
	db := openAuthDB(t)
	cid := seedAuthCompany(t, db)
	emp := seedCustomerUser(t, db, cid, "emp@t.com", "員工", false)
	client := newIdentifiedAuthClientWithDB(t, db, authz.Identity{UserID: "1", CompanyID: strconv.Itoa(cid), Role: "super", Roles: []string{"super"}})
	if _, err := client.ResetCustomerPassword(ctx, connect.NewRequest(&v1.ResetCustomerPasswordRequest{UserId: strconv.Itoa(emp)})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("目標非客戶應 invalid_argument,got %v", err)
	}
}

// ---------------------------------------------------------------------------
// F-2（最終全分支審查）：席位上線的**未守衛建帳號路徑**

// 席位＝該公司所有非 inactive 的 users 列（internal/services/counters.go），但只有 6 個寫入
// RPC 與 UpdateUser 復原掛了守衛。以下兩條會新增 users 列的路徑**原本不受任何檢查**：
//   - OIDC 首次登入（hd 對應公司 → 直接建 guest，status=active）；
//   - RegisterComplete 的 registration-token 分支（建 guest，status=pending）。
// 兩者都不需要管理權，員工用公司網域的 Google 帳號首次登入就能靜默超額佔席位。
//
// 驗收：達上限 → 被擋（PLAT-5001 對外碼）且**不落 users 列**；未達上限 → 照常成功。

// TestCallbackGuestCreationChecksSeatLimit：OIDC 首次登入建 guest 前必須先檢查席位。
func TestCallbackGuestCreationChecksSeatLimit(t *testing.T) {
	t.Run("未達上限：照常建 guest 並登入", func(t *testing.T) {
		e := newTestEnv(t)
		coID := mustCreateCompany(t, e, "example.com") // hd 對應公司 → 直接建 guest

		state := "state-seat-ok"
		if err := e.handler.deps.OneTime.Put(e.ctx, auth.StateKey(state), "1", time.Minute); err != nil {
			t.Fatalf("寫入 state: %v", err)
		}
		rec := e.callback(t, state, "")
		if rec.Code != http.StatusFound {
			t.Fatalf("callback 應 302,got %d", rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "http://localhost:3000/" {
			t.Fatalf("未達上限時應完成登入回跳首頁,got %q", loc)
		}
		u, err := e.db.User.Query().Where(user.EmailEQ("emp@example.com")).Only(e.ctx)
		if err != nil {
			t.Fatalf("未達上限時應建立 guest: %v", err)
		}
		if u.Role != RoleGuest || u.Status != user.StatusActive {
			t.Fatalf("guest 應為 active: role=%q status=%q", u.Role, u.Status)
		}
		want := strconv.Itoa(coID) + " " + entitlements.LimitSeats + " 1"
		if got := e.seatGuard.calls; len(got) != 1 || got[0] != want {
			t.Fatalf("應以 hd 對應的公司查一次席位（%q）,got %v", want, got)
		}
	})

	t.Run("達上限：不得建 users 列，回跳帶 PLAT-5001", func(t *testing.T) {
		e := newTestEnv(t)
		coID := mustCreateCompany(t, e, "example.com")
		e.seatGuard.err = seatLimitExceeded()

		state := "state-seat-blocked"
		if err := e.handler.deps.OneTime.Put(e.ctx, auth.StateKey(state), "1", time.Minute); err != nil {
			t.Fatalf("寫入 state: %v", err)
		}
		rec := e.callback(t, state, "")
		if rec.Code != http.StatusFound {
			t.Fatalf("callback 應 302（回跳登入頁）,got %d", rec.Code)
		}
		// OIDC 是 HTTP redirect，沒有 ErrorInfo 通道 → 碼放進 query 供前端查表顯示訊息
		// （frontend/src/lib/errcode.ts 的 PLAT-5001＝「已達方案上限（…），請升級方案」）。
		if loc := rec.Header().Get("Location"); !strings.Contains(loc, "error=PLAT-5001") {
			t.Fatalf("被擋時應回跳 login?error=PLAT-5001,got %q", loc)
		}
		if n := e.db.User.Query().Where(user.EmailEQ("emp@example.com")).CountX(e.ctx); n != 0 {
			t.Fatalf("達上限時不得建立 users 列（席位就是這個路徑超額的）,count=%d", n)
		}
		want := strconv.Itoa(coID) + " " + entitlements.LimitSeats + " 1"
		if got := e.seatGuard.calls; len(got) != 1 || got[0] != want {
			t.Fatalf("應以 hd 對應的公司查一次席位（%q）,got %v", want, got)
		}
	})

	t.Run("未注入守衛：fail-closed（漏注入＝無守衛）", func(t *testing.T) {
		e := newTestEnv(t)
		mustCreateCompany(t, e, "example.com")
		e.handler.deps.Entitlements = nil

		state := "state-no-guard"
		if err := e.handler.deps.OneTime.Put(e.ctx, auth.StateKey(state), "1", time.Minute); err != nil {
			t.Fatalf("寫入 state: %v", err)
		}
		rec := e.callback(t, state, "")
		if loc := rec.Header().Get("Location"); strings.Contains(loc, "http://localhost:3000/") && !strings.Contains(loc, "error=") {
			t.Fatalf("未注入守衛時不得放行建帳號,got %q", loc)
		}
		if n := e.db.User.Query().Where(user.EmailEQ("emp@example.com")).CountX(e.ctx); n != 0 {
			t.Fatalf("未注入守衛時不得建立 users 列,count=%d", n)
		}
	})
}

// TestRegisterCompleteChecksSeatLimit：registration-token 分支建 guest 前必須先檢查席位。
func TestRegisterCompleteChecksSeatLimit(t *testing.T) {
	newRegisterToken := func(t *testing.T, e *testEnv, email string) string {
		t.Helper()
		token, err := auth.NewRegistrationToken()
		if err != nil {
			t.Fatalf("NewRegistrationToken: %v", err)
		}
		if err := e.handler.deps.OneTime.Put(e.ctx, auth.RegistrationKey(token), email, time.Minute); err != nil {
			t.Fatalf("寫入 registration token: %v", err)
		}
		return token
	}

	t.Run("未達上限：照常建 guest（pending）", func(t *testing.T) {
		e := newTestEnv(t)
		coID := mustCreateCompany(t, e, "co-seat")
		token := newRegisterToken(t, e, "new@other.example")

		req := connect.NewRequest(&v1.RegisterCompleteRequest{CompanyId: itoa(coID), Name: "李四"})
		req.Header().Set("Cookie", RegistrationTokenCookie+"="+token)
		if _, err := e.rpc.RegisterComplete(e.ctx, req); err != nil {
			t.Fatalf("RegisterComplete: %v", err)
		}
		u, err := e.db.User.Query().Where(user.EmailEQ("new@other.example")).Only(e.ctx)
		if err != nil {
			t.Fatalf("未達上限時應建立 guest: %v", err)
		}
		if u.Role != RoleGuest || u.Status != user.StatusPending {
			t.Fatalf("guest 應為 pending: role=%q status=%q", u.Role, u.Status)
		}
		want := strconv.Itoa(coID) + " " + entitlements.LimitSeats + " 1"
		if got := e.seatGuard.calls; len(got) != 1 || got[0] != want {
			t.Fatalf("應以所選公司查一次席位（%q）,got %v", want, got)
		}
	})

	t.Run("達上限：不得建 users 列，回 PLAT-5001", func(t *testing.T) {
		e := newTestEnv(t)
		coID := mustCreateCompany(t, e, "co-seat")
		e.seatGuard.err = seatLimitExceeded()
		token := newRegisterToken(t, e, "new@other.example")

		req := connect.NewRequest(&v1.RegisterCompleteRequest{CompanyId: itoa(coID), Name: "李四"})
		req.Header().Set("Cookie", RegistrationTokenCookie+"="+token)
		_, err := e.rpc.RegisterComplete(e.ctx, req)
		if connect.CodeOf(err) != connect.CodeFailedPrecondition {
			t.Fatalf("達席位上限應 failed_precondition,got %v", err)
		}
		if got := authErrorInfo(t, err).GetCode(); got != "PLAT-5001" {
			t.Fatalf("ErrorInfo.code = %q；want PLAT-5001（前端據以導向升級方案）", got)
		}
		if n := e.db.User.Query().Where(user.EmailEQ("new@other.example")).CountX(e.ctx); n != 0 {
			t.Fatalf("達上限時不得建立 users 列,count=%d", n)
		}

		// F-5：被席位守衛擋下的那一次**不得燒掉一次性 registration token** —— 席位守衛（以及任何
		// 會拒絕的前置檢查）必須在消費憑證之前。這裡換一個未滿席的公司、帶著**同一個憑證**重試：
		// 憑證若已被消費，第二次會是 unauthenticated 而不是成功。
		e.seatGuard.err = nil
		other := mustCreateCompany(t, e, "co-room")
		req2 := connect.NewRequest(&v1.RegisterCompleteRequest{CompanyId: itoa(other), Name: "李四"})
		req2.Header().Set("Cookie", RegistrationTokenCookie+"="+token)
		if _, err := e.rpc.RegisterComplete(e.ctx, req2); err != nil {
			t.Fatalf("同一個 registration token 在未滿席的公司應仍有效（被擋那一次不得消費憑證）: %v", err)
		}
		u := e.db.User.Query().Where(user.EmailEQ("new@other.example")).OnlyX(e.ctx)
		if co, err := u.QueryCompany().Only(e.ctx); err != nil || co.ID != other {
			t.Fatalf("重試應把 guest 建在未滿席的公司 %d,got %v err=%v", other, co, err)
		}
	})

	t.Run("email 已註冊（SYS-2001）：不得燒掉憑證", func(t *testing.T) {
		e := newTestEnv(t)
		coID := mustCreateCompany(t, e, "co-dup")
		e.db.User.Create().
			SetEmail("dup@other.example").SetName("既有").SetStatus(user.StatusActive).
			SetRole("staff").SetIsCustomer(false).SetPasswordHash(auth.OIDCPasswordSentinel).
			SetCompanyID(coID).SaveX(e.ctx)
		token := newRegisterToken(t, e, "dup@other.example")

		attempt := func() error {
			req := connect.NewRequest(&v1.RegisterCompleteRequest{CompanyId: itoa(coID), Name: "李四"})
			req.Header().Set("Cookie", RegistrationTokenCookie+"="+token)
			_, err := e.rpc.RegisterComplete(e.ctx, req)
			return err
		}
		err := attempt()
		if got := authErrorInfo(t, err).GetCode(); got != "SYS-2001" {
			t.Fatalf("email 重複應回 SYS-2001,got %v", err)
		}
		// 第二次仍必須走到同一個「前置檢查失敗」（而不是 unauthenticated）→ 憑證沒被消費。
		err = attempt()
		if connect.CodeOf(err) == connect.CodeUnauthenticated {
			t.Fatalf("被前置檢查擋下不得消費 registration token（第二次應仍是 SYS-2001）,got %v", err)
		}
		if got := authErrorInfo(t, err).GetCode(); got != "SYS-2001" {
			t.Fatalf("第二次應仍是 SYS-2001,got %q err=%v", got, err)
		}
	})
}
