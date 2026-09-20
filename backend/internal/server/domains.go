// Package server 集中 DI（D31）。InitDomains() 為組裝唯一來源：
// 逐 domain 呼叫 mountXxx()，新增 domain 只動此檔。
package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"connectrpc.com/connect"

	"github.com/go-chi/chi/v5"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	authzopenfga "github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	domainauth "github.com/salesorder/sales-order-1.0/backend/internal/domain/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/handlers"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	postgresstore "github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/services"
	"github.com/salesorder/sales-order-1.0/backend/third_party/cache"
	"github.com/salesorder/sales-order-1.0/backend/third_party/database"
	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
)

// entitlementCacheTTL 為權益快照的行程內快取 TTL（縮短＝更即時、更多平台庫查詢；
// Plan C 換成 Valkey 快取時一併調整）。
const entitlementCacheTTL = 60 * time.Second

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
	s.tokens = tokens // 供 authzMiddleware Bearer JWT 路徑逐請求驗證(01 1.6/A2)
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
	// 權益守衛（平台域）先建立再掛載四個業務服務：建構子要求表態，漏掛即編譯失敗。
	entSvc := s.mountEntitlements(entClient)
	apiMux := http.NewServeMux()
	authPath, authHandler := salesorderv1connect.NewAuthServiceHandler(h, connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(entClient)))
	apiMux.Handle(authPath, authHandler)
	handlers.RegisterRoleHandler(apiMux, entClient) // RoleService(T18)
	// AbilityService(T9/D30):CASL 規則下發給前端 @casl/ability 初始化。
	abilityPath, abilityHandler := salesorderv1connect.NewAbilityServiceHandler(domainauth.NewAbilityHandler(entClient, domainauth.Config{DeveloperAccountEnabled: s.cfg.API.DeveloperAccountEnabled}), connect.WithInterceptors(requestid.Interceptor(), dbtenant.Interceptor(entClient)))
	apiMux.Handle(abilityPath, abilityHandler)
	services.RegisterCompanyServices(apiMux, entClient, entSvc)                          // CompanyService/DepartmentService(T20)
	services.RegisterUserServices(apiMux, entClient, entSvc)                             // UserService(02 Task 3)
	services.RegisterMetadictServices(apiMux, entClient)                                 // MetadictService(03 Task 2)
	services.RegisterAuditServices(apiMux, entClient)                                    // AuditService(03 Task 6, A4)
	services.RegisterCustomerServices(apiMux, entClient, s.cfg.Auth.FrontendURL, entSvc) // CustomerService(04 Task 1-2 + D22 帳號交付 URL)
	// 04 Task 3.4 部門級主檔(Warehouse/Route/ProcessingSpec/ProductCategory)。
	services.RegisterWarehouseService(apiMux, entClient)
	services.RegisterRouteService(apiMux, entClient)
	services.RegisterProcessingSpecService(apiMux, entClient)
	services.RegisterProductCategoryService(apiMux, entClient)
	services.RegisterProductService(apiMux, entClient, entSvc) // 04 Task 3.3 商品主檔
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

// mountEntitlements 建立平台權益判定服務並寫入 Server（供四個業務服務的配額守衛、
// 以及 T9/T10 的平台端／租戶端投影取用）。
// 平台域走 admin 連線：platform schema 對業務角色 app_rw 零權限（S9），store 只能是 owner 池。
// 連線不可用即拒絕啟動（比照 mountOpenFGA 的立場）：靜默降級成 Unlimited() 等於關掉全部配額，
// 而業務服務的建構子已強制每個呼叫端表態，不提供「未掛守衛」的 production 退路。
func (s *Server) mountEntitlements(db *ent.Client) *entitlements.Service {
	adminDB, err := database.OpenSQL(s.cfg.Database.AdminDSN())
	if err != nil {
		log.Fatalf("platform: admin 連線不可用,拒絕以無守衛狀態啟動(守衛缺席＝配額形同虛設): %v", err)
	}
	// 快取：v1 用行程內 MemoryCache；Valkey 實作與寫入後失效屬 Plan C 的訂閱寫入路徑,屆時換此處。
	svc := entitlements.New(postgresstore.New(adminDB), services.NewEntitlementCounter(db),
		entitlements.NewMemoryCache(), entitlementCacheTTL)
	s.entitlements = svc
	log.Println("platform: 權益守衛已掛載（entitlements.Service → 四個業務服務）")
	return svc
}

// openEntClient 開啟業務 PostgreSQL ent client:連線仍委派 third_party/database(D31),
// 但業務 client **必須**經 dbtenant.NewClient 建立,RLS driver 裝飾器才會生效
// (請求交易的 SET LOCAL 就在 driver.Tx 內套用)。測試/CLI/seed 不經此路徑。
func (s *Server) openEntClient() (*ent.Client, error) {
	db, err := database.OpenSQL(s.cfg.Database.DatabaseURL)
	if err != nil {
		return nil, err
	}
	return dbtenant.NewClient(db), nil
}

// mountOpenFGA 建立內嵌 OpenFGA 授權引擎(D32)並注入 Server。
// datastore 與業務共用 PostgreSQL(單一 store);dsn 沿用 owner DSN(AdminDSN)。
// OPENFGA_ENABLED=true 即 fail-fast(F2 裁定):所有環境 bootstrap 失敗都終止啟動,
// 因為引擎缺席(或零 tuple)會讓 GetAbility 永遠回空、前端守衛 fail-closed 把全站
// 使用者擋在 /403,比啟動失敗更難診斷。OPENFGA_ENABLED=false 為文件化退路(語意不變)。
func (s *Server) mountOpenFGA(db *ent.Client) {
	if !s.cfg.OpenFGA.Enabled {
		return
	}
	dsn := s.cfg.OpenFGA.DatabaseURL
	if dsn == "" {
		dsn = s.cfg.Database.AdminDSN()
	}
	client, err := ofga.NewPostgres(context.Background(), dsn, s.cfg.OpenFGA.StoreName)
	if err != nil {
		log.Fatalf("openfga: 授權引擎建立失敗(OPENFGA_ENABLED=true 即拒絕啟動,不降級): %v;請先執行 `go run ./cmd/migrate up` 建立 OpenFGA datastore schema,或設 OPENFGA_ENABLED=false 明示停用", err)
	}
	engine := authzopenfga.New(client)
	s.SetOpenFGA(engine)
	// 供給 OpenFGA 授權資料(role_permissions→role ability;users→role assigned),
	// 使 middleware Check 得以判定(修復零 tuple → 全員 deny);供給失敗同樣終止啟動。
	if err := authz.Provision(context.Background(), engine, db); err != nil {
		log.Fatalf("openfga: 授權資料供給失敗(拒絕以零 tuple 啟動,否則全站將被擋在 /403): %v", err)
	}
	log.Println("openfga: 內嵌授權引擎已掛載(D32)")
}
