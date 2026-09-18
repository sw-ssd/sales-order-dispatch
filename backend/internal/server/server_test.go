package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"connectrpc.com/connect"
	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	authzopenfga "github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
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

	t.Run("production + 自訂密鑰但 infra 不可達被 DB/Valkey fail-fast 擋下", func(t *testing.T) {
		// 單測環境無 PostgreSQL/Valkey;production 完整可啟動路徑由整合/驗收測試涵蓋。
		// 此處驗證 secret/developer 檢查通過後,infra 不可達仍會拒絕啟動(設計 §3 D31)。
		s := New(&config.Config{API: config.API{Env: "production"}, Auth: config.Auth{JWTSecret: "custom-secret"}})
		if err := s.Init(); err == nil {
			t.Fatal("production 於無 infra 環境應被 DB/Valkey fail-fast 拒絕")
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
// 干擾;OpenFGA.Enabled=true 讓 authorizeRPC 走 OpenFGA 檢查而非停用回退)
// 與 scs session manager。
func newIdentityTestEnv() (*Server, *scs.SessionManager) {
	s := &Server{cfg: &config.Config{
		API:     config.API{DeveloperAccountEnabled: true},
		OpenFGA: config.OpenFGA{Enabled: true},
	}}
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

// testSessionCookie 建立指定使用者身分的 session 並回傳 cookie(供 middleware probe)。
func testSessionCookie(t *testing.T, sessions *scs.SessionManager, userID int, role string) *http.Cookie {
	t.Helper()
	seed := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth.EstablishWebSession(r.Context(), sessions, userID, role, 0)
		w.WriteHeader(http.StatusNoContent)
	})
	rec := httptest.NewRecorder()
	sessions.LoadAndSave(seed).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/seed", nil))
	for _, c := range rec.Result().Cookies() {
		if c.Name == auth.SessionCookieName {
			return c
		}
	}
	t.Fatal("seed: 未取得 session cookie")
	return nil
}

// TestAuthzMiddlewareOpenFGA D32 驗收:受保護 RPC path 由 OpenFGA Check 判定。
// 有權 → 放行(probe 執行);未登入 → Unauthenticated;developer 跳過。write 拒絕對應之
// TestAuthorizeRPCWriteDenied 單測覆蓋(不受 HTTP 層 session 影響)。
func TestAuthzMiddlewareOpenFGA(t *testing.T) {
	s, sessions := newIdentityTestEnv()
	ctx := context.Background()
	db := openIdentityDB(t, "file:ofga-mw?mode=memory&cache=shared&_fk=1")
	co := db.Company.Create().SetName("測試公司").SetIdentifier("T-3").SaveX(ctx)
	staff := db.User.Create().SetEmail("staff@example.com").SetName("staff").SetStatus(user.StatusActive).
		SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)
	dev := db.User.Create().SetEmail("dev@example.com").SetName("dev").SetStatus(user.StatusActive).
		SetRole("developer").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)

	// 起 OpenFGA 記憶體 engine 並注入 Server。
	fgaClient, err := ofga.NewMemory(ctx, "server-test-store")
	if err != nil {
		t.Fatalf("NewMemory: %v", err)
	}
	t.Cleanup(fgaClient.Close)
	fga := authzopenfga.New(fgaClient)
	s.SetOpenFGA(fga)
	// 授予 staff 對 role 資源的 can_read(資料驅動 tuple)。
	if err := fga.WriteTuple(ctx, "user:"+strconv.FormatInt(int64(staff.ID), 10), "can_read", "ability:role"); err != nil {
		t.Fatalf("WriteTuple: %v", err)
	}
	// 受保護 path:RoleService/ListRoles → role/read。
	const protectedPath = "/salesorder.v1.RoleService/ListRoles"

	var called bool
	probe := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})
	mw := sessions.LoadAndSave(s.authzMiddleware(db, sessions, probe))

	req := func(user int, role string) *http.Request {
		r := httptest.NewRequest(http.MethodGet, protectedPath, nil)
		r.AddCookie(testSessionCookie(t, sessions, user, role))
		return r
	}

	t.Run("staff 有權 → 放行", func(t *testing.T) {
		called = false
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req(int(staff.ID), "staff"))
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", rec.Code)
		}
		if !called {
			t.Fatal("有權時 probe 應被呼叫")
		}
	})
	t.Run("未登入 → 401 Unauthenticated + 合法 JSON body", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, protectedPath, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
		var body map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("錯誤 body 應為合法 JSON: %v (body=%q)", err, rec.Body.Bytes())
		}
		if body["code"] != "unauthenticated" {
			t.Fatalf("body.code = %q, want unauthenticated", body["code"])
		}
	})
	t.Run("developer → 跳過放行", func(t *testing.T) {
		called = false
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req(int(dev.ID), "developer"))
		if rec.Code != http.StatusNoContent {
			t.Fatalf("developer status = %d, want 204", rec.Code)
		}
		if !called {
			t.Fatal("developer 應跳過 Check 並放行")
		}
	})
}

// TestAuthorizeRPCWriteDenied 驗證 authorizeRPC 對未授予 write 動作回 PermissionDenied。
func TestAuthorizeRPCWriteDenied(t *testing.T) {
	s, _ := newIdentityTestEnv()
	ctx := context.Background()
	fgaClient, err := ofga.NewMemory(ctx, "rpc-denied-store")
	if err != nil {
		t.Fatalf("NewMemory: %v", err)
	}
	t.Cleanup(fgaClient.Close)
	fga := authzopenfga.New(fgaClient)
	s.SetOpenFGA(fga)
	// staff 僅授予 can_read role;未授予 can_write。
	if err := fga.WriteTuple(ctx, "user:7", "can_read", "ability:role"); err != nil {
		t.Fatalf("WriteTuple: %v", err)
	}
	c := authz.WithEngine(authz.WithIdentity(ctx, authz.Identity{UserID: "7", Roles: []string{"staff"}}), fga)
	// role/write → can_write 未授予 → PermissionDenied。
	err = s.authorizeRPC(c, rpcAuth{resource: "role", action: "write"})
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("write 未授予應 permission_denied,got %v", err)
	}
	// role/read → can_read 已授予 → 放行。
	if err := s.authorizeRPC(c, rpcAuth{resource: "role", action: "read"}); err != nil {
		t.Fatalf("read 已授予應放行,got %v", err)
	}
}

// TestAuthorizeRPCFallbackSemantics 驗證 OPENFGA_ENABLED 開關的兩種結局:
// 停用 → 回退放行(由服務層授權承擔);啟用卻無引擎 → fail-closed(Internal)。
func TestAuthorizeRPCFallbackSemantics(t *testing.T) {
	ctx := context.Background()
	c := authz.WithIdentity(ctx, authz.Identity{UserID: "7", Roles: []string{"staff"}})

	t.Run("OpenFGA 停用 → 回退放行", func(t *testing.T) {
		s := &Server{cfg: &config.Config{API: config.API{DeveloperAccountEnabled: true}, OpenFGA: config.OpenFGA{Enabled: false}}}
		if err := s.authorizeRPC(c, rpcAuth{resource: "role", action: "read"}); err != nil {
			t.Fatalf("停用 OpenFGA 應回退放行,got %v", err)
		}
	})
	t.Run("OpenFGA 啟用但無引擎 → fail-closed", func(t *testing.T) {
		s := &Server{cfg: &config.Config{API: config.API{DeveloperAccountEnabled: true}, OpenFGA: config.OpenFGA{Enabled: true}}}
		err := s.authorizeRPC(c, rpcAuth{resource: "role", action: "read"})
		if connect.CodeOf(err) != connect.CodeInternal {
			t.Fatalf("啟用但無引擎應 fail-closed(internal),got %v", err)
		}
	})
}
