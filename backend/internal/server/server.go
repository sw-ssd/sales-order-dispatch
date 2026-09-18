// Package server 集中 DI：Server struct 持有共享依賴，
// Init() 收斂 fail-fast 啟動檢查，InitDomains() 逐 domain 組裝（D31）。
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"connectrpc.com/connect"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	authzopenfga "github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
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
		// 設計 §3(D31):production fail-fast 需涵蓋 DB/Valkey 連線,
		// 避免 auth/service 因 infra 不可用而靜默不掛載(見 domains.go mountAuth 開發降級)。
		if err := database.Probe(context.Background(), s.cfg.Database.DatabaseURL); err != nil {
			return fmt.Errorf("config: ENV=production 無法連線資料庫: %w", err)
		}
		vc := cache.NewClient(s.cfg.Cache.ValkeyAddr)
		if err := cache.Ping(context.Background(), vc); err != nil {
			return fmt.Errorf("config: ENV=production 無法連線 Valkey: %w", err)
		}
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
// 未登入 / 查無使用者 / developer 關閉 / session token_version 與 DB 不符時以零值身分通過,
// 受保護 RPC 再由 authorizeRPC 以 OpenFGA Check 判定（未登入→Unauthenticated）。
// token_version 不符代表改密碼 / 停用 / 強制登出已 bump——一併銷毀 session 使該請求即時登出。
func (s *Server) authzMiddleware(entClient *ent.Client, sessions *scs.SessionManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = authz.WithCASLEnabled(ctx, s.cfg.API.CASLEnforcementEnabled)
		ctx = authz.WithDB(ctx, entClient)
		if s.fga != nil {
			ctx = authz.WithEngine(ctx, s.fga)
		}

		userID := auth.SessionUserID(ctx, sessions)
		if userID > 0 {
			sessionTV := auth.SessionTokenVersion(ctx, sessions)
			if id, scope, ok := s.identityFor(ctx, entClient, userID, sessionTV); ok {
				ctx = authz.WithIdentity(ctx, id)
				ctx = auth.WithRLS(ctx, scope)
			} else if sessionTV >= 0 {
				// session 已記錄 tv 卻身分失效（token_version 變更 / 帳號停用等）→ 銷毀 session 強制登出。
				_ = sessions.Destroy(ctx)
			}
		}
		// OpenFGA 授權閘門(D32):受保護 RPC path 以 OpenFGA Check 判定;developer 跳過。
		if rpc, ok := protectedRPC[r.URL.Path]; ok {
			if err := s.authorizeRPC(ctx, rpc); err != nil {
				writeConnectError(w, err)
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
		return connect.NewError(connect.CodeUnauthenticated, errors.New("未登入"))
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
		// production 由 mountOpenFGA fail-fast 避免此態;此處為防線(即使單元測試亦不誤放行)。
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
		return connect.NewError(connect.CodePermissionDenied, errors.New("無權限執行此操作"))
	}
	return nil
}

// writeConnectError 以 Connect 錯誤協定寫出錯誤回應(供 middleware)。
// 以 JSON 物件(單次 Marshal)輸出合法 body,並依 connect code 對映正確 HTTP 狀態:
// unauthenticated→401、permission_denied→403、invalid_argument→400,其餘→500。
func writeConnectError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusForCode(connect.CodeOf(err)))
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    connect.CodeOf(err).String(),
		"message": err.Error(),
	})
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

// identityFor 由使用者載入身分與 RLS scope（company/department eager-load）。
// sessionTokenVersion 為 session 簽發時記錄的 token_version；與 DB 目前值不符
// （改密碼 / 停用 / 角色變更 / 強制登出已 bump）→ ok=false（fail-closed）。
// 帳號不存在 / 非 active / developer 關閉 → ok=false（零值身分，fail-closed）。
func (s *Server) identityFor(ctx context.Context, entClient *ent.Client, userID int, sessionTokenVersion int) (authz.Identity, auth.RLSScope, bool) {
	u, err := entClient.User.Query().WithCompany().WithDepartment().Where(user.ID(userID)).Only(ctx)
	if err != nil || u.Status != user.StatusActive {
		return authz.Identity{}, auth.RLSScope{}, false
	}
	// token_version 比對：session 簽發時的 tv(-1 = 未記錄的舊版 session,不比對)與 DB 現值
	// 不一致 → 身分失效(401 落點)。
	if sessionTokenVersion >= 0 && sessionTokenVersion != u.TokenVersion {
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

func (s *Server) handleVersion(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"version": Version})
}
