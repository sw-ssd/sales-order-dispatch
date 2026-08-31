// Package server 集中 DI：Server struct 持有共享依賴，
// Init() 收斂 fail-fast 啟動檢查，InitDomains() 逐 domain 組裝（D31）。
package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
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

	// AuthService Connect-RPC（Web session cookie 亦經 scs 中介層讀寫）。
	// connect 產生的 handler 依 r.URL.Path 全路徑分派,故以 http.StripPrefix 剝除
	// /api/v1 前綴後掛載(與 services.RegisterCompanyServices 的掛載慣例一致)。
	_, authHandler := salesorderv1connect.NewAuthServiceHandler(h)
	s.router.Mount("/api/v1", http.StripPrefix("/api/v1", sessions.LoadAndSave(authHandler)))

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
