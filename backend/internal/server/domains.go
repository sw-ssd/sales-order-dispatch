// Package server 集中 DI（D31）。InitDomains() 為組裝唯一來源：
// 逐 domain 呼叫 mountXxx()，新增 domain 只動此檔。
package server

import (
	"context"
	"database/sql"
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
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
	postgresstore "github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/services"
	"github.com/salesorder/sales-order-1.0/backend/third_party/cache"
	"github.com/salesorder/sales-order-1.0/backend/third_party/database"
	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
)

// entitlementCacheTTL 為權益快照的快取 TTL（縮短＝更即時、更多平台庫查詢）。
//
// 這只是**保底**：正確性不靠它 —— 方案／override／訂閱狀態異動時，寫入路徑會顯式刪除該租戶的鍵
// （見 entitlements.Invalidate／InvalidateAll，以及三個寫入來源：平台 RPC、billing 排程、consumer）。
// TTL 的用途是「萬一某條寫入路徑漏了失效」時的最長收斂時間（見 billing.invalidate 的說明）。
const entitlementCacheTTL = 60 * time.Second

// InitDomains 逐 domain 組裝 repo→usecase→handler 並掛上 router。
// 新增 domain 只動此一檔（D31）：於 InitDomains() 加一行呼叫對應的 mountXxx()，
// 並在下方定義該 mountXxx() 的組裝鏈。組裝唯一來源為此檔，禁止在 main.go 或各套件自行組裝。
func (s *Server) InitDomains() {
	s.mountAuth()
	s.mountPlatformAuth()
}

// platformAdminDBOf 回共用的平台 admin 池：InitDomains 已建即共用，否則自開。
// 未結項 #9(Plan B)：此前 mountEntitlements 與 mountPlatformAuth 各自 OpenSQL，
// 同一個行程兩個池。收斂點在「取用時」而非 InitDomains —— mountAuth 在無 Valkey／
// 無 ent 時會提前 return（開發降級），InitDomains 先開池會讓 TestVersionEndpoint
// 這類無 DB 測試 Fatal（實測）。共用由呼叫端傳池：先到的開，後到的用。
func (s *Server) platformAdminDBOf() *sql.DB {
	if s.platformAdminDB == nil {
		db, err := database.OpenSQL(s.cfg.Database.AdminDSN())
		if err != nil {
			log.Fatalf("platform: admin 連線不可用,拒絕以無守衛狀態啟動(守衛缺席＝配額形同虛設): %v", err)
		}
		s.platformAdminDB = db
	}
	return s.platformAdminDB
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

	// 權益守衛（平台域）先建立再掛載業務服務與 auth handler：建構子要求表態，漏掛即編譯失敗；
	// auth handler 的兩條建帳號路徑（OIDC 首次登入、RegisterComplete）也吃它做席位守衛，
	// 漏注入會在執行期 fail-closed 擋住註冊（見 handlers.guardSeats），故此處必須先建立。
	entSvc := s.mountEntitlements(entClient)

	h := handlers.NewAuthHandler(handlers.AuthDeps{
		Cfg:          s.cfg,
		DB:           entClient,
		Tokens:       tokens,
		Lockout:      lockout,
		OneTime:      oneTime,
		Sessions:     sessions,
		Entitlements: entSvc,
	})

	// /api/v1 底下所有 Connect-RPC 共用一個 ServeMux:connect 產生的 handler 依
	// r.URL.Path 全路徑分派,掛載時剝除 /api/v1 前綴(與 RegisterCompanyServices 慣例一致)。
	// LoadAndSave + authzMiddleware 包在最外層:session 身分 → authz.Identity/RLS ctx(T14 Step 4)。
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
	services.RegisterSalesOrderService(apiMux, entClient)      // SalesOrderService(05 Task 4)
	// T10/T10b 租戶端權益投影：租戶後台／App 的「我的方案與用量」。掛在 /api/v1 之下（租戶
	// session ＋ RLS），**不是** /platform/ —— 那裡是 operator cookie 與平台工具的路徑範圍。
	services.RegisterTenantEntitlementService(apiMux, entClient, entSvc)
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
	adminDB := s.platformAdminDBOf()
	// 快取：**跨行程共用的 Valkey**（排程與 consumer 也是寫入來源，只有共用同一顆快取，
	// 它們的失效才會對 API 生效）；Valkey 不可用時退回行程內記憶體 —— 快取是加速器，
	// 不是啟動前提，沒有它配額判定照常運作（只是各行程各記一份、跨行程失效不生效）。
	entCache := s.openEntitlementCache()
	counters := services.NewEntitlementCounter(db)
	svc := entitlements.New(postgresstore.New(adminDB), counters, entCache, entitlementCacheTTL)
	s.entitlements, s.entitlementCache, s.entitlementCounter = svc, entCache, counters
	log.Println("platform: 權益守衛已掛載（entitlements.Service → 四個業務服務）")
	return svc
}

// openEntitlementCache 選用權益快取：Valkey 可用即用，否則退回行程內記憶體。
//
// 判準沿用 repo 既有的 Valkey 設定來源（config.Cache.ValkeyAddr）＋ 一次 Ping —— 不看「位址字串
// 是否為空」：它有預設值 localhost:6379，永遠非空，而以 Ping 判定才與 mountAuth 的降級語意
// 一致（連得上才算設定）。**不 fail-fast**：Valkey 掛掉只該讓快取退化（每請求多打一次平台庫），
// 不該讓服務起不來 —— 判定層對快取故障本來就會回源（見 entitlements.Service.state）。
func (s *Server) openEntitlementCache() entitlements.Cache {
	client := cache.NewClient(s.cfg.Cache.ValkeyAddr)
	if err := cache.Ping(context.Background(), client); err != nil {
		log.Printf("platform: 略過 Valkey 權益快取（%v）→ 改用行程內記憶體"+
			"（多 replica／跨行程的失效不會生效，最長 TTL 內可能讀到舊權益）", err)
		return entitlements.NewMemoryCache()
	}
	return entitlements.NewValkeyCache(client)
}

// mountPlatformAuth 掛載平台工具的登入端點（/platform/auth/*）並把 operatorauth.Service 記在
// Server 上（T9 的 PlatformAdminService 以它的 Interceptor 擋下非 operator）。
//
// 與租戶 auth 完全分離（D38/S8）：不同 secret、不同 audience、不同 cookie（名稱與 Path=/platform）。
// 設定整組未設 = 不掛載（開發環境友善）；production 由 Init() 保證整組齊備或整組不設。
//
// 兩段式降級，兩個理由：
//   - interceptor 只需設定即生效：已簽發的 operator token 在 Google discovery 暫時不可用時仍
//     驗得了，重啟不會把平台工具整組鎖死（登入進不來，但既有工作階段不中斷）。
//   - 登入端點**一律註冊**，只是缺 OIDC 依賴時回 503：回 404 會讓人以為路由沒寫（掛載問題），
//     503 才是實情（暫時不可用）。
func (s *Server) mountPlatformAuth() {
	if !s.cfg.Platform.Configured() {
		return
	}
	// 共用 mountEntitlements 建立的池（先到先開、後到共用，未結項 #9）。
	adminDB := s.platformAdminDBOf()
	opAuth := operatorauth.New(operatorauth.Config{
		Secret:        s.cfg.Platform.OperatorJWTSecret,
		CookieDomain:  s.cfg.Platform.CookieDomain,
		ConsoleURL:    s.cfg.Platform.ConsoleURL,
		AllowedDomain: s.cfg.Platform.AllowedEmailDomain,
		// http 的開發環境不得設 Secure（瀏覽器不回送），其餘環境一律 Secure。
		CookieSecure: s.cfg.API.Env != "development",
	}, postgresstore.NewOperators(adminDB))
	s.operatorAuth = opAuth
	s.router.Get(operatorauth.LoginPath, opAuth.Login)
	s.router.Get(operatorauth.CallbackPath, opAuth.Callback)

	// 平台 RPC 掛在**字面 /platform/ 之下**（瀏覽器路徑成 /platform/platform.v1.PlatformAdminService/…）。
	//
	// 不能照 Connect 的自然路徑掛在根：operator cookie 的 Path 是 /platform，而 RFC 6265 的
	// path-match 是逐段前綴 —— /platform 對 /platform.v1.… 不成立（未涵蓋的第一個字元是 "."），
	// 於是瀏覽器**不會送出** cookie，登入看似成功但每個 RPC 都回 401。反過來把 cookie 的 Path
	// 放寬成 "/" 也不行（operator cookie 會跟著送往租戶 API）。兩者既然必須是同一個值，這裡
	// 就用 CookiePath 當前綴（不寫字面字串，免得日後改了一邊）。
	//
	// 掛在此處（登入路由之後、OIDC 依賴檢查之前）：RPC 只需要 interceptor，不需要 Google
	// discovery —— OIDC 暫時不可用時既有工作階段仍能讀取。
	// 平台寫入的帳務狀態機（T9）：與排程／consumer 用同一組寫入介面（admin 連線），並接上
	// **同一顆權益快取** —— 少了 .WithCache，收款把訂閱帶回 active 之後，判定層最長一個 TTL 內
	// 仍讀舊權益（服務層自己的失效只涵蓋它經手的寫入，billing 內部的狀態轉移要看這裡）。
	// 快取未掛載時為 nil：entitlements.Invalidate 對 nil 是 no-op（沒有快取＝沒有東西要失效）。
	platformBilling := billing.NewBilling(postgresstore.New(adminDB)).WithCache(s.entitlementCache)
	platformMux := http.NewServeMux()
	services.RegisterPlatformAdminService(platformMux, postgresstore.NewAdmin(adminDB),
		platformBilling, s.entitlementCache, s.entitlementCounter, opAuth)
	s.router.Mount(operatorauth.CookiePath, http.StripPrefix(operatorauth.CookiePath, platformMux))

	clientID := s.cfg.Auth.GoogleClientID
	if clientID == "" {
		log.Println("platform: GOOGLE_CLIENT_ID 未設定，平台登入端點回 503（interceptor 仍生效）")
		return
	}
	// 回呼網址由租戶 OIDC 的 redirect URL 推導：兩者同一個 API 來源，只差路徑，
	// 故不需要另一個環境變數（也就沒有「設了 console 卻忘記設 API 網址」的失配）。
	redirectURL, err := operatorauth.RedirectURL(s.cfg.Auth.GoogleRedirectURL)
	if err != nil {
		log.Printf("platform: 平台登入端點回 503（%v）", err)
		return
	}
	verifier, err := auth.NewGoogleVerifier(context.Background(), clientID)
	if err != nil {
		log.Printf("platform: 平台登入端點回 503（Google discovery: %v）", err)
		return
	}
	oauthCfg := auth.NewGoogleOAuthConfig(clientID, s.cfg.Auth.GoogleClientSecret, redirectURL)
	opAuth.WithOIDC(oauthCfg, auth.NewGoogleOAuthExchanger(oauthCfg), verifier)
	log.Println("platform: 平台操作者認證已掛載（OIDC ＋ operator JWT ＋ cookie）")
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
