// Package server 集中 DI：Server struct 持有共享依賴，
// Init() 收斂 fail-fast 啟動檢查，InitDomains() 逐 domain 組裝（D31）。
package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/handlers"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/third_party/cache"
)

// Version 目前版本；正式版號由建置 ldflags 注入。
var Version = "0.1.0-dev"

// Server 持有全部共享依賴與路由。
type Server struct {
	cfg    *config.Config
	router *chi.Mux
}

// New 建立 Server 並掛上全域 middleware 與基礎路由。
func New(cfg *config.Config) *Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	s := &Server{cfg: cfg, router: r}
	r.Get("/api/v1/version", s.handleVersion)
	return s
}

// Init 執行 fail-fast 啟動檢查（DB/Valkey 連線、必要 secret 等，於各 domain 計畫補上）。
func (s *Server) Init() error {
	// 設計書 §4.4 啟動防護：production 誤開 developer 繞過授權 → 拒絕啟動。
	if s.cfg.API.Env == "production" && s.cfg.API.DeveloperAccountEnabled {
		return fmt.Errorf("config: ENV=production 且 DEVELOPER_ACCOUNT_ENABLED=true 會繞過 Casbin/RLS,拒絕啟動")
	}
	return nil
}

// Run 啟動 HTTP 服務。
func (s *Server) Run() error {
	return http.ListenAndServe(s.cfg.API.Addr, s.router)
}

// Handler 暴露路由供測試與掛載。
func (s *Server) Handler() http.Handler {
	return s.router
}

// authzMiddleware 將 scs session 身分轉換為 authz.Identity + RLS scope 注入 ctx（T14 Step 4）。
// 必須位於 sessions.LoadAndSave 之後（ctx 才帶 session 資料）。
// 未登入 / 查無使用者 / developer 關閉時以零值身分通過（fail-closed：CASL 規則載入全略過 → denied，
// Casbin 由各服務層以 EnforceAny 判斷）。
func (s *Server) authzMiddleware(entClient *ent.Client, sessions *scs.SessionManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = authz.WithCASLEnabled(ctx, s.cfg.API.CASLEnforcementEnabled)
		ctx = authz.WithDB(ctx, entClient)

		userID := auth.SessionUserID(ctx, sessions)
		if userID > 0 {
			if id, scope, ok := s.identityFor(ctx, entClient, userID); ok {
				ctx = authz.WithIdentity(ctx, id)
				ctx = auth.WithRLS(ctx, scope)
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// identityFor 由使用者載入身分與 RLS scope（company/department eager-load）。
// 帳號不存在 / 非 active / developer 關閉 → ok=false（零值身分，fail-closed）。
func (s *Server) identityFor(ctx context.Context, entClient *ent.Client, userID int) (authz.Identity, auth.RLSScope, bool) {
	u, err := entClient.User.Query().WithCompany().WithDepartment().Where(user.ID(userID)).Only(ctx)
	if err != nil || u.Status != user.StatusActive {
		return authz.Identity{}, auth.RLSScope{}, false
	}
	// developer 帳號僅在開關啟用時繞過 Casbin/RLS（設計書 §4.4）。
	if u.Role == "developer" && !s.cfg.API.DeveloperAccountEnabled {
		return authz.Identity{}, auth.RLSScope{}, false
	}

	var companyID, deptID string
	if u.Edges.Company != nil {
		companyID = strconv.FormatInt(int64(u.Edges.Company.ID), 10)
	}
	if u.Edges.Department != nil {
		deptID = strconv.FormatInt(int64(u.Edges.Department.ID), 10)
	}
	id := authz.Identity{
		UserID:       strconv.FormatInt(int64(u.ID), 10),
		CompanyID:    companyID,
		DepartmentID: deptID,
		Role:         u.Role,
		Roles:        auth.RolesFor(u.Role), // 依 Casbin g 展開(含自身)
	}
	scope := auth.RLSScope{
		UserID:       id.UserID,
		CompanyID:    companyID,
		DepartmentID: deptID,
		DataScope:    auth.ScopeForRole(u.Role),
	}
	return id, scope, true
}

// mountAuth 組裝 auth domain：ent client + Valkey client → token/鎖定/一次性 store →
// scs session → AuthService Connect handler 與 OIDC 公開端點。
// 開發降級：DB / Valkey / Google discovery 不可用時略過掛載並 log（正式環境由 Init() fail-fast 保證）。
func (s *Server) mountAuth() {
	entClient, err := s.openEntClient()
	if err != nil {
		log.Printf("auth: 略過掛載（ent client: %v）", err)
		return
	}
	valkeyClient := cache.NewClient(s.cfg.Cache.ValkeyAddr)
	if err := cache.Ping(context.Background(), valkeyClient); err != nil {
		log.Printf("auth: 略過掛載（Valkey: %v）", err)
		return
	}

	kv := auth.NewRedisStore(valkeyClient)
	tokens := auth.NewTokenManager(s.cfg.Auth.JWTSecret, kv)
	lockout := auth.NewLoginLock(kv)
	oneTime := auth.NewOneTimeStore(kv)
	sessions := auth.WebSessionManager(auth.NewSessionStore(kv),
		s.cfg.Auth.SessionLifetime, s.cfg.Auth.SessionSecure, s.cfg.Auth.SessionSameSite)

	h := handlers.NewAuthHandler(handlers.AuthDeps{
		Cfg:      s.cfg,
		DB:       entClient,
		Tokens:   tokens,
		Lockout:  lockout,
		OneTime:  oneTime,
		Sessions: sessions,
	})

	// /api/v1 底下所有 Connect-RPC 共用一個 ServeMux:connect 產生的 handler 依
	// r.URL.Path 全路徑分派,掛載時剝除 /api/v1 前綴(與 RegisterCompanyServices 慣例一致)。
	// LoadAndSave + authzMiddleware 包在最外層:session 身分 → authz.Identity/RLS ctx(T14 Step 4)。
	apiMux := http.NewServeMux()
	authPath, authHandler := salesorderv1connect.NewAuthServiceHandler(h)
	apiMux.Handle(authPath, authHandler)
	handlers.RegisterRoleHandler(apiMux, entClient) // RoleService(T18)
	s.router.Mount("/api/v1", http.StripPrefix("/api/v1", sessions.LoadAndSave(s.authzMiddleware(entClient, sessions, apiMux))))

	// OIDC 公開端點：需 Google client id 與 discovery 可用
	clientID := s.cfg.Auth.GoogleClientID
	if clientID == "" {
		log.Println("auth: GOOGLE_CLIENT_ID 未設定，略過 OIDC 路由")
		return
	}
	verifier, err := auth.NewGoogleVerifier(context.Background(), clientID)
	if err != nil {
		log.Printf("auth: 略過 OIDC 路由（Google discovery: %v）", err)
		return
	}
	oauthCfg := auth.NewGoogleOAuthConfig(clientID, s.cfg.Auth.GoogleClientSecret, s.cfg.Auth.GoogleRedirectURL)
	h.SetOIDC(oauthCfg, auth.NewGoogleOAuthExchanger(oauthCfg), verifier)

	s.router.Group(func(r chi.Router) {
		r.Use(sessions.LoadAndSave)
		r.Get("/api/v1/auth/google", h.GoogleLogin)
		r.Get("/api/v1/auth/google/callback", h.GoogleCallback)
	})
}

// openEntClient 開啟 PostgreSQL ent client（pgx driver）。
func (s *Server) openEntClient() (*ent.Client, error) {
	db, err := sql.Open("pgx", s.cfg.Database.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	drv := entsql.OpenDB(dialect.Postgres, db)
	return ent.NewClient(ent.Driver(drv)), nil
}

func (s *Server) handleVersion(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"version": Version})
}
