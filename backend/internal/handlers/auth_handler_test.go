package handlers

import (
	"context"
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
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
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

// testEnv 組裝 handlers 測試環境(enttest sqlite + MemoryStore + fake OIDC)。
type testEnv struct {
	ctx      context.Context
	db       *ent.Client
	kv       *auth.MemoryStore
	tokens   *auth.TokenManager
	handler  *AuthHandler
	sessions *scs.SessionManager
	verifier *fakeVerifier
	exch     *fakeExchanger
	rpc      salesorderv1connect.AuthServiceClient
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
	h := NewAuthHandler(AuthDeps{
		Cfg: cfg, DB: db, Tokens: tokens, Lockout: lockout, OneTime: oneTime, Sessions: sessions,
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
