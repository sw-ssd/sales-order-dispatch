package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
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
	tokens := auth.NewTokenManager("test-secret", kv)
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
func (e *testEnv) seedWebSession(t *testing.T, userID int, role string) string {
	t.Helper()
	seed := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth.EstablishWebSession(r.Context(), e.sessions, userID, role)
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

func itoa(n int) string {
	return strconv.Itoa(n)
}
