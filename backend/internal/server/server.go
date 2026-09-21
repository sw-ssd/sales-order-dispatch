// Package server 集中 DI：Server struct 持有共享依賴，
// Init() 收斂 fail-fast 啟動檢查，InitDomains() 逐 domain 組裝（D31）。
package server

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	authzopenfga "github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
	"github.com/salesorder/sales-order-1.0/backend/internal/obs/requestid"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
	salesorderv1connect "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/third_party/cache"
	"github.com/salesorder/sales-order-1.0/backend/third_party/database"
)

// Version 目前版本；正式版號由建置 ldflags 注入。
var Version = "0.1.0-dev"

// Server 持有全部共享依賴與路由。
type Server struct {
	cfg    *config.Config
	router *chi.Mux
	fga    *authzopenfga.Engine // 可選:OpenFGA 授權引擎(設入後 middleware 對受保護 RPC 做 Check,D32)
	tokens *auth.TokenManager   // JWT access/refresh 管理(01 1.6;App Bearer 路徑逐請求驗證)
	// entitlements 為平台權益判定（配額守衛的來源，由 mountAuth 建立）；T9/T10 的平台端與
	// 租戶端投影亦由此取用，故留在 Server 上而非只傳進四個業務服務。
	entitlements *entitlements.Service
	// entitlementCache 為判定用的快取（Valkey 實作或缺 Valkey 時的行程內版），於 mountEntitlements
	// 一併建立。**平台寫入 RPC（T9）的注入點**：改方案／override／訂閱狀態後必須失效同一顆快取，
	// 而快取的建立（連線設定）屬組裝，T9 只該拿到介面。nil = 尚未掛載。
	entitlementCache entitlements.Cache
	// entitlementCounter 為業務域的用量計數器（席位＝未停用帳號數），於 mountEntitlements 一併
	// 建立。**平台寫入（T9）的注入點**：降席位必須先知道目前用了幾席，而平台域不認得業務 schema
	// （計數只能由業務域提供）。nil = 尚未掛載，該守衛會拒絕而不是放行。
	entitlementCounter entitlements.Counter
	// operatorAuth 為平台工具認證(mountPlatformAuth 建立)。T9 的 PlatformAdminService 以
	// operatorAuth.Interceptor() 擋下非 operator;nil = 平台工具未設定,該 RPC 不掛載。
	operatorAuth *operatorauth.Service
	// platformAdminDB 為平台 admin 連線池（未結項 #9(Plan B)：此前 mountEntitlements 與
	// mountPlatformAuth 各自 OpenSQL，同一個行程兩個池）。由 InitDomains 一次建立、
	// 兩處共用；nil = 尚未建立（測試直呼 mountXxx 時各開各的，與舊行為一致）。
	platformAdminDB *sql.DB
}

// rpcAuth 為受保護 RPC path 的 OpenFGA 對映(resource, action)。
// action 只取 read/write 兩類,對應授權 model 的 can_read/can_write。
// 新增領域時,於此表補上相應的受保護 RPC path(資料驅動能力由 role_permissions→tuples 承載)。
type rpcAuth struct {
	resource string
	action   string
}

// protectedRPC 對映受保護 RPC path → OpenFGA (resource, action)。
// path 為 Connect-RPC 全路徑(剝除 /api/v1 前綴後)。未登入/無權 → Unauthenticated/PermissionDenied。
var protectedRPC = map[string]rpcAuth{
	"/salesorder.v1.RoleService/ListRoles":              {"role", "read"},
	"/salesorder.v1.RoleService/GetRolePermissions":     {"role", "read"},
	"/salesorder.v1.RoleService/UpdateRolePermissions":  {"role", "write"},
	"/salesorder.v1.RoleService/ListConditionFields":    {"role", "read"},
	"/salesorder.v1.CompanyService/ListCompanies":       {"company", "read"},
	"/salesorder.v1.CompanyService/GetCompany":          {"company", "read"},
	"/salesorder.v1.CompanyService/CreateCompany":       {"company", "write"},
	"/salesorder.v1.CompanyService/UpdateCompany":       {"company", "write"},
	"/salesorder.v1.CompanyService/DeleteCompany":       {"company", "write"},
	"/salesorder.v1.DepartmentService/ListDepartments":  {"department", "read"},
	"/salesorder.v1.DepartmentService/GetDepartment":    {"department", "read"},
	"/salesorder.v1.DepartmentService/CreateDepartment": {"department", "write"},
	"/salesorder.v1.DepartmentService/UpdateDepartment": {"department", "write"},
	"/salesorder.v1.DepartmentService/DeleteDepartment": {"department", "write"},
	"/salesorder.v1.UserService/ListUsers":              {"user", "read"},
	"/salesorder.v1.UserService/GetUser":                {"user", "read"},
	"/salesorder.v1.UserService/CreateUser":             {"user", "write"},
	"/salesorder.v1.UserService/UpdateUser":             {"user", "write"},
	"/salesorder.v1.UserService/AssignRole":             {"user", "write"},
	"/salesorder.v1.UserService/Deactivate":             {"user", "write"},
	"/salesorder.v1.UserService/ForceLogout":            {"user", "write"},
}

// SetOpenFGA 注入 OpenFGA 授權引擎(啟動組裝時;nil 則跳過 middleware 檢查)。
func (s *Server) SetOpenFGA(e *authzopenfga.Engine) {
	s.fga = e
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
	// 啟動防護：production 不得使用空或預設 JWT 密鑰（JWT_SECRET 未設定時 envconfig 落回
	// config/auth.go 的 dev 預設常數,等同未設定 → 拒絕啟動）。
	if s.cfg.API.Env == "production" {
		if s.cfg.Auth.JWTSecret == "" {
			return fmt.Errorf("config: ENV=production 且 JWT_SECRET 為空,拒絕啟動")
		}
		if s.cfg.Auth.JWTSecret == config.DefaultJWTSecret {
			return fmt.Errorf("config: ENV=production 且 JWT_SECRET 仍為預設值(dev-only),拒絕啟動")
		}
		// 平台工具設定(D38)守護。平台工具採**顯式 opt-in**:整組不設 = 不啟用(與組裝處
		// `Platform.Configured()` 的掛載判斷一致,開發環境亦可不設);一旦設了其一(secret 或 console URL)
		// 就必須整組齊備 —— 只設一半會讓 operator 登入看似啟用卻走不通,而錯誤直到登入才爆。
		// 注意:不可用「Platform 非零值」當條件。envconfig 會替 AllowedEmailDomain 等欄位填預設值,
		// 故 config.New() 出來的 Platform 恆非零 → 那樣的條件連「整組不設」都擋,且訊息承諾的
		// 「整組留空以停用」永遠無法生效(operator 清空後重啟仍被同一條守護拒絕)。
		if (s.cfg.Platform.OperatorJWTSecret != "" || s.cfg.Platform.ConsoleURL != "") && !s.cfg.Platform.Configured() {
			return fmt.Errorf("config: ENV=production 且平台工具只設了一部分,拒絕啟動（請補齊 PLATFORM_JWT_SECRET 與 PLATFORM_CONSOLE_URL,或整組不設以停用平台工具）")
		}
		// 沿用隨 repo 公開的 dev 預設值 = 任何人都能偽造 operator token,且啟動毫無警示 → 拒絕。
		if s.cfg.Platform.Configured() && s.cfg.Platform.OperatorJWTSecret == config.DefaultJWTSecret {
			return fmt.Errorf("config: ENV=production 且 PLATFORM_JWT_SECRET 仍為預設值(dev-only),拒絕啟動")
		}
		// 兩個密鑰共用等於租戶 token 可冒充平台操作者(反之亦然)→ 拒絕啟動。訊息不帶密鑰值。
		if s.cfg.Platform.Configured() && s.cfg.Platform.OperatorJWTSecret == s.cfg.Auth.JWTSecret {
			return fmt.Errorf("config: PLATFORM_JWT_SECRET 不得與 JWT_SECRET 相同(跨用將使租戶 token 可冒充平台操作者)")
		}
		// 設計 §3(D31):production fail-fast 需涵蓋 DB/Valkey 連線,
		// 避免 auth/service 因 infra 不可用而靜默不掛載(見 domains.go mountAuth 開發降級)。
		if err := database.Probe(context.Background(), s.cfg.Database.DatabaseURL); err != nil {
			return fmt.Errorf("config: ENV=production 無法連線資料庫: %w", err)
		}
		// 業務連線必須是非 superuser(且不帶 BYPASSRLS):PG 的 superuser 恆繞過 RLS(FORCE 亦然),
		// 業務 DSN 誤指 owner 會讓 00024–00028 的租戶邊界**靜默消失**(所有端點與測試照常綠)。
		if err := assertBusinessRoleNotSuperuser(context.Background(), s.cfg.Database.DatabaseURL); err != nil {
			return fmt.Errorf("config: ENV=production 業務連線角色不合法: %w", err)
		}
		vc := cache.NewClient(s.cfg.Cache.ValkeyAddr)
		if err := cache.Ping(context.Background(), vc); err != nil {
			return fmt.Errorf("config: ENV=production 無法連線 Valkey: %w", err)
		}
	}
	return nil
}

// assertBusinessRoleNotSuperuser 以**業務連線**確認當前角色不是 superuser／不帶 BYPASSRLS。
// PostgreSQL 的 superuser(與帶 BYPASSRLS 的角色)恆繞過 RLS,`FORCE ROW LEVEL SECURITY` 亦然 ——
// 業務 DSN 誤指 owner 時租戶邊界會**靜默消失**(服務照常回應、測試照常綠),故 production 拒絕啟動。
func assertBusinessRoleNotSuperuser(ctx context.Context, dsn string) error {
	db, err := database.OpenSQL(dsn)
	if err != nil {
		return fmt.Errorf("開啟業務連線: %w", err)
	}
	defer func() { _ = db.Close() }()
	var bypassesRLS bool
	if err := db.QueryRowContext(ctx,
		`SELECT rolsuper OR rolbypassrls FROM pg_roles WHERE rolname = current_user`).Scan(&bypassesRLS); err != nil {
		return fmt.Errorf("查詢業務連線的角色屬性(pg_roles): %w", err)
	}
	if bypassesRLS {
		return errors.New("業務連線角色為 superuser 或帶 BYPASSRLS,會繞過 RLS 使租戶隔離失效;" +
			"請把 DATABASE_URL 指向非 owner 的業務角色(如 app_rw),其密碼可用 `task backend:db:app-password` 設定")
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

// bearerToken 由 Authorization header 取出 Bearer token(大小寫不敏感的前綴比對)。
// 無 Bearer 前綴或空白 → 空字串(表示無 JWT 憑證)。
func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(h) > len(prefix) && strings.EqualFold(h[:len(prefix)], prefix) {
		return strings.TrimSpace(h[len(prefix):])
	}
	return ""
}

// clientIP 由 RemoteAddr 取下 IP(去除 port;供稽核來源資訊,I9)。
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// authzMiddleware 將 scs session 身分轉換為 authz.Identity + RLS scope 注入 ctx（T14 Step 4）。
// 必須位於 sessions.LoadAndSave 之後（ctx 才帶 session 資料）。
// 未登入 / 查無使用者 / developer 關閉 / session token_version 與 DB 不符時以零值身分通過,
// 受保護 RPC 再由 authorizeRPC 以 OpenFGA Check 判定（未登入→Unauthenticated）。
// token_version 不符代表改密碼 / 停用 / 強制登出已 bump——一併銷毀 session 使該請求即時登出。
func (s *Server) authzMiddleware(entClient *ent.Client, sessions *scs.SessionManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = authz.WithCASLEnabled(ctx, s.cfg.API.CASLEnforcementEnabled)
		ctx = authz.WithDB(ctx, entClient)
		// 稽核來源資訊(IP / User-Agent)每請求注入一次,供 service 層寫稽核(I9;03 2.6.2)。
		ctx = audit.WithMeta(ctx, audit.Meta{IP: clientIP(r), UserAgent: r.UserAgent()})
		if s.fga != nil {
			ctx = authz.WithEngine(ctx, s.fga)
		}

		userID := auth.SessionUserID(ctx, sessions)
		if userID > 0 {
			// Web session 路徑：身分由 scs session 提供。
			sessionTV := auth.SessionTokenVersion(ctx, sessions)
			if id, scope, ok := s.identityFor(ctx, entClient, userID, sessionTV); ok {
				ctx = authz.WithIdentity(ctx, id)
				ctx = auth.WithRLS(ctx, scope)
			} else if sessionTV >= 0 {
				// session 已記錄 tv 卻身分失效（token_version 變更 / 帳號停用等）→ 銷毀 session 強制登出。
				_ = sessions.Destroy(ctx)
			}
		} else if tok := bearerToken(r); tok != "" && s.tokens != nil {
			// App/API Bearer JWT 路徑（01 1.6/A2 缺口）：無 cookie session 時改驗 access JWT。
			// VerifyAccess 做簽章/exp/tv 比對；identityFor 再載入使用者並判定帳號與公司 active。
			// 失敗時不注入身分(零值) → authorizeRPC 對受保護 RPC 回 unauthenticated。
			claims, err := s.tokens.VerifyAccess(ctx, tok)
			if err == nil {
				if id, scope, ok := s.identityFor(ctx, entClient, claims.UserID, claims.TokenVersion); ok {
					ctx = authz.WithIdentity(ctx, id)
					ctx = auth.WithRLS(ctx, scope)
				}
			}
		}
		// A2 公司停用連鎖(2.1.3):所屬公司非 active(非 developer)→ unauthenticated。
		// 不解銷 session(scope.CompanyActive=false 時 identity 仍注入),使公司恢復 active 後
		// 既有 session 可續用,不需重新登入。
		if id := authz.IdentityFrom(ctx); len(id.Roles) > 0 && id.Role != "developer" {
			if scope := auth.RLSFrom(ctx); !scope.CompanyActive {
				writeConnectError(w, r, connect.NewError(connect.CodeUnauthenticated, errors.New("所屬公司已停用,無法繼續操作")))
				return
			}
		}
		// A3 受限態(1.5.2):must_change_password=true 時僅放行 ChangePassword,其餘回 failed_precondition
		// (AUTH-3004),強制首登改密碼後才能使用業務 RPC。
		if id := authz.IdentityFrom(ctx); id.MustChangePassword && r.URL.Path != salesorderv1connect.AuthServiceChangePasswordProcedure {
			writeConnectError(w, r, errcode.AuthPasswordChangeRequired.Error(nil))
			return
		}
		// OpenFGA 授權閘門(D32):受保護 RPC path 以 OpenFGA Check 判定;developer 跳過。
		if rpc, ok := protectedRPC[r.URL.Path]; ok {
			if err := s.authorizeRPC(ctx, rpc); err != nil {
				writeConnectError(w, r, err)
				return
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// authorizeRPC 對受保護 RPC path 執行 OpenFGA Check 授權閘門:
// 未登入 → Unauthenticated;無權 → PermissionDenied;developer/super 逃生門(開關啟用)跳過。
// OpenFGA 停用(OPENFGA_ENABLED=false) → 回退(放行,授權由各服務層檢查承擔);
// OpenFGA 啟用卻無引擎 → 視為接線失敗,fail-closed(拒絕),避免授權被靜默繞過。
func (s *Server) authorizeRPC(ctx context.Context, rpc rpcAuth) error {
	id := authz.IdentityFrom(ctx)
	if len(id.Roles) == 0 {
		return errcode.AuthUnauthenticated.Error(nil)
	}
	// developer/super 逃生門:僅在開關啟用時(身分成立)跳過 OpenFGA 檢查。
	// super 為「全資源」管理角色(rolePolicy "*":{"*"}),其 role_permissions 亦有具體能力 tuples;
	// 逃生門提供對「未列入 adminResources 之新資源」的防禦(避免新增資源即鎖死 super)。
	if (id.Role == "developer" || id.Role == "super") && s.cfg.API.DeveloperAccountEnabled {
		return nil
	}
	if !s.cfg.OpenFGA.Enabled {
		// 刻意停用 OpenFGA → 回退語意(config.OpenFGA.Enabled 註解):放行,由服務層授權承擔。
		return nil
	}
	e := authz.EngineFrom(ctx)
	if e == nil {
		// OpenFGA 已啟用卻無引擎 → 視為建置/接線失敗,fail-closed:拒絕(避免授權被靜默繞過)。
		// 正常啟動下 OPENFGA_ENABLED=true 已由 mountOpenFGA fail-fast 擋掉此態;此處為防線
		// (即使單元測試亦不誤放行)。
		return connect.NewError(connect.CodeInternal, errors.New("OpenFGA 授權引擎未就緒"))
	}
	relation := "can_read"
	if rpc.action == "write" {
		relation = "can_write"
	}
	allowed, err := e.Check(ctx, "user:"+id.UserID, relation, "ability:"+rpc.resource)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return errcode.SysPermissionDenied.Error(map[string]string{"resource": rpc.resource, "action": rpc.action})
	}
	return nil
}

// connectErrBody 為 Connect 錯誤協定的 JSON 形狀（與 connect-go 的 wire 格式一致）：
// {"code":"permission_denied","message":"…","details":[{"type":"…","value":"<base64>"}]}。
// 既有欄位（code／message）刻意保留：只讀這兩欄的既有前端不受影響。
type connectErrBody struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Details []connectErrDetail `json:"details,omitempty"`
}

// connectErrDetail 對映 connect 的錯誤 detail（type 為去前綴的完整型別名，value 為 proto 值的
// base64；與 connect-go 內部 wire 格式相同，只是該型別未匯出，故在此重建同樣的形狀）。
type connectErrDetail struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// writeConnectError 以 Connect 錯誤協定寫出錯誤回應(供 middleware 的閘門)。
//
// 為什麼要自己寫:這些閘門在 connect handler **之前**就拒絕請求,因此 requestid.Interceptor
// 的「回應邊界補 trace_id」看不到它們 —— 客戶端只會拿到一個沒有碼、沒有 trace_id 的錯誤,
// 客服無從追查。此處以 requestid.Ensure/Stamp 補上 trace_id(與 RPC 路徑同一份實作),
// 並輸出與 connect 一致的 JSON 形狀(含 details),使客戶端能用同一套解析讀到錯誤碼。
//
// HTTP 狀態碼仍由 httpStatusForCode 依 connect code 對映(unauthenticated→401、
// permission_denied→403、invalid_argument→400、其餘→500)。
func writeConnectError(w http.ResponseWriter, r *http.Request, err error) {
	ctx, _ := requestid.Ensure(r.Context(), r.URL.Path)
	err = requestid.Stamp(ctx, err)

	body := connectErrBody{Code: connect.CodeOf(err).String(), Message: err.Error()}
	if ce, ok := err.(*connect.Error); ok {
		// 與 connect 一致:message 只用 Message()(不含 "code: " 前綴),details 逐一帶出。
		body.Message = ce.Message()
		for _, d := range ce.Details() {
			body.Details = append(body.Details, connectErrDetail{
				Type:  d.Type(),
				Value: base64.RawStdEncoding.EncodeToString(d.Bytes()),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusForCode(connect.CodeOf(err)))
	_ = json.NewEncoder(w).Encode(body)
}

// httpStatusForCode 對映 Connect code → HTTP 狀態碼(供 middleware 錯誤回應)。
func httpStatusForCode(c connect.Code) int {
	switch c {
	case connect.CodeUnauthenticated:
		return http.StatusUnauthorized // 401
	case connect.CodePermissionDenied:
		return http.StatusForbidden // 403
	case connect.CodeInvalidArgument:
		return http.StatusBadRequest // 400
	case connect.CodeInternal:
		return http.StatusInternalServerError // 500
	default:
		return http.StatusInternalServerError
	}
}

// errIdentityRejected 為「身分不成立」的內部哨兵(帳號不存在/非 active/tv 不符/developer 關閉):
// 與 DB 錯誤一樣落到 ok=false,但不需為每一種情形各留一個回傳值。
var errIdentityRejected = errors.New("server: 身分不成立(fail-closed)")

// identityFor 由使用者載入身分與 RLS scope（company/department eager-load）。
// sessionTokenVersion 為 session 簽發時記錄的 token_version；與 DB 目前值不符
// （改密碼 / 停用 / 角色變更 / 強制登出已 bump）→ ok=false（fail-closed）。
// 帳號不存在 / 非 active / developer 關閉 → ok=false（零值身分，fail-closed）。
//
// 查詢一律在**系統範圍**交易內執行:本函式是身分解析本身(middleware 階段,ctx 尚未有租戶 scope),
// 而 users/roles 在 00028 之後受 RLS 約束 —— 未包系統範圍會回 0 列 → 所有已登入請求變成 401
// 並被 middleware 順手銷毀 session。
func (s *Server) identityFor(ctx context.Context, entClient *ent.Client, userID int, sessionTokenVersion int) (authz.Identity, auth.RLSScope, bool) {
	var id authz.Identity
	var scope auth.RLSScope
	err := dbtenant.SystemScopeTx(ctx, entClient, func(tx *ent.Tx) error {
		u, err := tx.Client().User.Query().WithCompany().WithDepartment().Where(user.ID(userID)).Only(ctx)
		if err != nil || u.Status != user.StatusActive {
			return errIdentityRejected
		}
		// token_version 比對：session 簽發時的 tv(-1 = 未記錄的舊版 session,不比對)與 DB 現值
		// 不一致 → 身分失效(401 落點)。
		if sessionTokenVersion >= 0 && sessionTokenVersion != u.TokenVersion {
			return errIdentityRejected
		}
		// developer 帳號僅在開關啟用時繞過 Casbin/RLS（設計書 §4.4）。
		if u.Role == "developer" && !s.cfg.API.DeveloperAccountEnabled {
			return errIdentityRejected
		}

		var companyID, deptID string
		if u.Edges.Company != nil {
			companyID = strconv.FormatInt(int64(u.Edges.Company.ID), 10)
		}
		if u.Edges.Department != nil {
			deptID = strconv.FormatInt(int64(u.Edges.Department.ID), 10)
		}
		id = authz.Identity{
			UserID:             strconv.FormatInt(int64(u.ID), 10),
			CompanyID:          companyID,
			DepartmentID:       deptID,
			Role:               u.Role,
			Roles:              auth.RolesFor(u.Role), // 依 Casbin g 展開(含自身)
			MustChangePassword: u.MustChangePassword,  // A3 首登/臨時密碼態
		}
		// A2 公司停用連鎖(2.1.3):companyActive=false 表示公司非 active(company 為 nil 視同停用),
		// 由 middleware 阻擋該請求(unauthenticated)但不刪 session,恢復 active 後可續用。
		companyActive := u.Edges.Company != nil && u.Edges.Company.Status == company.StatusActive
		scope = auth.RLSScope{
			UserID:        id.UserID,
			CompanyID:     companyID,
			DepartmentID:  deptID,
			DataScope:     dataScopeForUser(ctx, tx.Client(), u.Role),
			CompanyActive: companyActive,
		}
		return nil
	})
	if err != nil {
		return authz.Identity{}, auth.RLSScope{}, false
	}
	return id, scope, true
}

// dataScopeForUser 取得使用者的 RLS 資料範圍(審查複審):
// 優先讀 roles.data_scope(表為權威來源,支援自訂角色);查無角色列時回退內建對映
// (auth.ScopeForRole)。自訂角色若走硬編碼對映會回空字串→RLSStatements 不注入 data_scope,
// 導致 RLS 啟用後範圍錯置(data_scope=department 的自訂角色反被視為全公司)。
// 皆無法取得 → 空字串(不注入,fail-closed 由 RLS policy 承擔)。
//
// db **必須是已套用系統範圍的 client**(唯一呼叫端 identityFor 傳入 tx.Client()):roles 在 00028
// 之後受 RLS 約束,傳入未帶 scope 的 client 會讀到 0 列 → 靜默退回內建對映 → 自訂角色範圍錯置。
func dataScopeForUser(ctx context.Context, db *ent.Client, roleCode string) auth.DataScope {
	if r, err := db.Role.Query().Where(role.CodeEQ(roleCode)).Only(ctx); err == nil && r.DataScope != "" {
		return auth.DataScope(r.DataScope)
	}
	return auth.ScopeForRole(roleCode)
}

func (s *Server) handleVersion(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"version": Version})
}
