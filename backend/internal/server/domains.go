// Package server 集中 DI（D31）。InitDomains() 為組裝唯一來源：
// 逐 domain 呼叫 mountXxx()，新增 domain 只動此檔。
package server

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	authzopenfga "github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	domainauth "github.com/salesorder/sales-order-1.0/backend/internal/domain/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/handlers"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/services"
	"github.com/salesorder/sales-order-1.0/backend/third_party/cache"
	"github.com/salesorder/sales-order-1.0/backend/third_party/database"
	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
)

// InitDomains 逐 domain 組裝 repo→usecase→handler 並掛上 router。
// 新增 domain 只動此一檔（D31）：於 InitDomains() 加一行呼叫對應的 mountXxx()，
// 並在下方定義該 mountXxx() 的組裝鏈。組裝唯一來源為此檔，禁止在 main.go 或各套件自行組裝。
func (s *Server) InitDomains() {
	s.mountAuth()
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
	s.mountOpenFGA(entClient)
	valkeyClient := cache.NewClient(s.cfg.Cache.ValkeyAddr)
	if err := cache.Ping(context.Background(), valkeyClient); err != nil {
		log.Printf("auth: 略過掛載（Valkey: %v）", err)
		return
	}

	kv := auth.NewRedisStore(valkeyClient)
	tokens := auth.NewTokenManager(s.cfg.Auth.JWTSecret, kv, entClient)
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
	// AbilityService(T9/D30):CASL 規則下發給前端 @casl/ability 初始化。
	abilityPath, abilityHandler := salesorderv1connect.NewAbilityServiceHandler(domainauth.NewAbilityHandler(entClient, domainauth.Config{DeveloperAccountEnabled: s.cfg.API.DeveloperAccountEnabled}))
	apiMux.Handle(abilityPath, abilityHandler)
	services.RegisterCompanyServices(apiMux, entClient) // CompanyService/DepartmentService(T20)
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

// openEntClient 開啟 PostgreSQL ent client（委派 third_party/database，統一初始化路徑，D31）。
func (s *Server) openEntClient() (*ent.Client, error) {
	return database.OpenEnt(s.cfg.Database.DatabaseURL)
}

// mountOpenFGA 建立內嵌 OpenFGA 授權引擎(D32)並注入 Server。
// datastore 與業務共用 PostgreSQL(單一 store);dsn 沿用 Database.DatabaseURL。
// production 為 fail-closed:引擎建立失敗即終止啟動(fail-fast),避免授權閘門被靜默繞過。
// 非 production 開發降級:失敗時 log 並以 nil 繼續(此環境由各服務層既有授權承擔)。
func (s *Server) mountOpenFGA(db *ent.Client) {
	if !s.cfg.OpenFGA.Enabled {
		return
	}
	dsn := s.cfg.OpenFGA.DatabaseURL
	if dsn == "" {
		dsn = s.cfg.Database.DatabaseURL
	}
	client, err := ofga.NewPostgres(context.Background(), dsn, s.cfg.OpenFGA.StoreName)
	if err != nil {
		if s.cfg.API.Env == "production" {
			log.Fatalf("config: ENV=production 且 OPENFGA_ENABLED=true 但 OpenFGA 引擎建立失敗,拒絕啟動: %v", err)
		}
		log.Printf("openfga: 略過授權引擎掛載(engine: %v),回退既有 RLS/Casbin 授權", err)
		return
	}
	s.SetOpenFGA(authzopenfga.New(client))
	// 供給 OpenFGA 授權資料(role_permissions→role ability;users→role assigned),
	// 使 middleware Check 得以判定(修復零 tuple → 全員 deny)。production 供給失敗即終止。
	if err := authz.Provision(context.Background(), authzopenfga.New(client), db); err != nil {
		if s.cfg.API.Env == "production" {
			log.Fatalf("config: OpenFGA 授權資料供給失敗,拒絕啟動: %v", err)
		}
		log.Printf("openfga: 授權資料供給失敗(略過): %v", err)
	}
	log.Println("openfga: 內嵌授權引擎已掛載(D32)")
}
