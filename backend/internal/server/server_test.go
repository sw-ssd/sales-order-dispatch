package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
)

func TestVersionEndpoint(t *testing.T) {
	s := New(config.New())
	s.InitDomains()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("回應非 JSON: %v", err)
	}
	if body["version"] == "" {
		t.Error("回應應含 version 欄位")
	}
}

// TestInitRejectsInsecureJWTSecret P1-3 驗收:production 環境使用空或預設 JWT 密鑰
// 時拒絕啟動(fail-fast);自訂密鑰(且未誤開 developer)可啟動。
func TestInitRejectsInsecureJWTSecret(t *testing.T) {
	cases := []struct {
		name   string
		secret string
	}{
		{"JWT_SECRET 為空", ""},
		{"JWT_SECRET 為預設常數", config.DefaultJWTSecret},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := New(&config.Config{API: config.API{Env: "production"}, Auth: config.Auth{JWTSecret: tc.secret}})
			if err := s.Init(); err == nil {
				t.Fatal("production 使用不安全 JWT 密鑰應拒絕啟動")
			}
		})
	}

	t.Run("production + developer 仍拒絕(既有防護)", func(t *testing.T) {
		s := New(&config.Config{API: config.API{Env: "production", DeveloperAccountEnabled: true}, Auth: config.Auth{JWTSecret: "custom-secret"}})
		if err := s.Init(); err == nil {
			t.Fatal("production 誤開 developer 應拒絕啟動")
		}
	})

	t.Run("production + 自訂密鑰可啟動", func(t *testing.T) {
		s := New(&config.Config{API: config.API{Env: "production"}, Auth: config.Auth{JWTSecret: "custom-secret"}})
		if err := s.Init(); err != nil {
			t.Fatalf("production + 自訂 JWT 密鑰應可啟動,got %v", err)
		}
	})

	t.Run("development 使用預設密鑰不阻擋", func(t *testing.T) {
		s := New(&config.Config{API: config.API{Env: "development"}, Auth: config.Auth{JWTSecret: config.DefaultJWTSecret}})
		if err := s.Init(); err != nil {
			t.Fatalf("development 使用預設密鑰不應阻擋,got %v", err)
		}
	})
}

// newIdentityTestEnv 建立測試用 Server(DeveloperAccountEnabled=true 避免 developer 防護
// 干擾)與 scs session manager。
func newIdentityTestEnv() (*Server, *scs.SessionManager) {
	s := &Server{cfg: &config.Config{API: config.API{DeveloperAccountEnabled: true}}}
	sessions := auth.WebSessionManager(memstore.New(), 30*24*time.Hour, false, "lax")
	return s, sessions
}

// openIdentityDB 開啟測試用 enttest sqlite client(唯一 DSN,避免跨測試共享)。
func openIdentityDB(t *testing.T, dsn string) *ent.Client {
	t.Helper()
	db := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestIdentityForTokenVersionMismatch P2-5 驗收:identityFor 比對 session 簽發時的
// token_version 與 DB 現值——不一致(改密碼/停用/強制登出已 bump)→ ok=false。
func TestIdentityForTokenVersionMismatch(t *testing.T) {
	s, _ := newIdentityTestEnv()
	ctx := context.Background()
	db := openIdentityDB(t, "file:identity-tv?mode=memory&cache=shared&_fk=1")
	co := db.Company.Create().SetName("測試公司").SetIdentifier("T-1").SaveX(ctx)
	u := db.User.Create().
		SetEmail("tv@example.com").SetName("測試").SetStatus(user.StatusActive).
		SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)

	// session 簽發時 tv=0,DB 亦為 0 → 一致 → 身分成立。
	id, scope, ok := s.identityFor(ctx, db, u.ID, 0)
	if !ok || id.UserID == "" || scope.CompanyID != "1" {
		t.Fatalf("tv 一致應建立身分,got ok=%v id=%+v", ok, id)
	}
	// 舊版 session(未記錄 tv,-1)→ 不比對 → 仍成立(部署相容)。
	if _, _, ok := s.identityFor(ctx, db, u.ID, -1); !ok {
		t.Fatal("未記錄 tv 的舊版 session 應仍成立(不比對)")
	}
	// bump 至 1:session 仍為 0 → 不一致 → 身分失效(401 落點)。
	db.User.UpdateOneID(u.ID).AddTokenVersion(1).SaveX(ctx)
	if _, _, ok := s.identityFor(ctx, db, u.ID, 0); ok {
		t.Fatal("tv 不一致應身分失效(ok=false)")
	}
	// 新 session 帶 tv=1 → 一致 → 成立。
	if _, _, ok := s.identityFor(ctx, db, u.ID, 1); !ok {
		t.Fatal("tv=1 一致應建立身分")
	}
}

// TestAuthzMiddlewareDestroysStaleSession P2-5 驗收:authzMiddleware 偵測 session
// token_version 與 DB 不符時銷毀 session(強制登出),後續請求不再視為已登入。
func TestAuthzMiddlewareDestroysStaleSession(t *testing.T) {
	s, sessions := newIdentityTestEnv()
	ctx := context.Background()
	db := openIdentityDB(t, "file:identity-mw?mode=memory&cache=shared&_fk=1")
	co := db.Company.Create().SetName("測試公司").SetIdentifier("T-2").SaveX(ctx)
	u := db.User.Create().
		SetEmail("mw@example.com").SetName("測試").SetStatus(user.StatusActive).
		SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)

	// 以 seed handler 寫入帶 tv=0 的 session,取得 cookie。
	seed := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth.EstablishWebSession(r.Context(), sessions, u.ID, u.Role, 0)
		w.WriteHeader(http.StatusNoContent)
	})
	rec := httptest.NewRecorder()
	sessions.LoadAndSave(seed).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/seed", nil))
	var cookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == auth.SessionCookieName {
			cookie = c
			break
		}
	}
	if cookie == nil {
		t.Fatal("seed: 未取得 session cookie")
	}

	// 受 authzMiddleware 保護的 handler:回報 ctx 身分。
	var gotID authz.Identity
	probe := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = authz.IdentityFrom(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})
	mw := sessions.LoadAndSave(s.authzMiddleware(db, sessions, probe))

	req := func() *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/probe", nil)
		r.AddCookie(cookie)
		return r
	}

	// tv 一致(0=0)→ 身分注入。
	rec = httptest.NewRecorder()
	mw.ServeHTTP(rec, req())
	if gotID.UserID == "" {
		t.Fatal("tv 一致時應注入身分")
	}

	// bump 至 1:session tv=0 不符 → 身分失效且 session 被銷毀。
	db.User.UpdateOneID(u.ID).AddTokenVersion(1).SaveX(ctx)
	rec = httptest.NewRecorder()
	mw.ServeHTTP(rec, req())
	if gotID.UserID != "" {
		t.Fatal("tv 不一致時不應注入身分")
	}
	// 同一 cookie 再送一次:session 已銷毀 → 視為未登入。
	rec = httptest.NewRecorder()
	mw.ServeHTTP(rec, req())
	if gotID.UserID != "" {
		t.Fatal("session 銷毀後應視為未登入(不注入身分)")
	}
	if strings.TrimSpace(rec.Header().Get("Set-Cookie")) == "" {
		t.Log("注意:session 銷毀後應清除 cookie(scs 依實作決定是否重送)")
	}
}
