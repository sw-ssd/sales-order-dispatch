package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
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

func TestCallbackCreatesGuestUser(t *testing.T) {
	// T11 驗收:mock Google token → 使用者建立(role=guest),Web 設 session 並回跳前端。
	e := newTestEnv(t)
	coID := mustCreateCompany(t, e, "example.com") // identifier = hd 網域

	state := "state-1"
	if err := e.handler.deps.OneTime.Put(e.ctx, auth.StateKey(state), "1", time.Minute); err != nil {
		t.Fatalf("寫入 state: %v", err)
	}
	rec := e.callback(t, state, "")

	if rec.Code != http.StatusFound {
		t.Fatalf("callback 應 302,got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "http://localhost:3000/" {
		t.Fatalf("回跳應至前端根路徑,got %q", loc)
	}
	// Web session cookie 已設定
	foundCookie := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == auth.SessionCookieName && c.Value != "" {
			foundCookie = true
		}
	}
	if !foundCookie {
		t.Fatal("callback 應設定 session cookie")
	}

	// 使用者已建立:role=guest,status=active,歸屬 hd 對應公司
	u, err := e.db.User.Query().Where(user.EmailEQ("emp@example.com")).Only(e.ctx)
	if err != nil {
		t.Fatalf("使用者應已建立: %v", err)
	}
	if u.Role != RoleGuest {
		t.Fatalf("role 應為 guest,got %q", u.Role)
	}
	if u.Status != user.StatusActive {
		t.Fatalf("首次登入 guest 應為 active(待完成註冊),got %q", u.Status)
	}
	co, err := u.QueryCompany().Only(e.ctx)
	if err != nil || co.ID != coID {
		t.Fatalf("guest 應歸屬 hd 對應公司,got company=%v err=%v", co, err)
	}

	// state 一次性:重放被拒
	rec2 := e.callback(t, state, "")
	if rec2.Code == http.StatusFound && rec2.Header().Get("Location") == "http://localhost:3000/" {
		t.Fatal("state 重放不應成功")
	}
}
