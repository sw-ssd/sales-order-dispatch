//go:build integration

package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/salesorder/sales-order-1.0/backend/config"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/handlers"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	postgresstore "github.com/salesorder/sales-order-1.0/backend/internal/platform/store/postgres"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/services"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// 本檔驗 auth handler 的兩條**無身分建帳號路徑**（OIDC 首次登入、RegisterComplete 的
// registration-token 分支）在**真 PG ＋ app_rw ＋ 真 RLS ＋ 真 entitlements.Service** 下確實會
// 被席位上限擋下（F-6）。
//
// 為何非真容器不可（本檔存在的唯一理由）：席位計數原本在**沒有 RLS scope 的 ctx** 上執行，而
// 00028 對 users 是 ENABLE + FORCE RLS —— 未設 scope 時 policy 把 users 濾成 0 列 → `used=0`
// → 任何上限都不觸發。也就是守衛在生產是**裝飾品**：sqlite（enttest）沒有 RLS 語意，單元層
// 怎麼測都綠（那正是「測試全綠但漏洞還在」）。真容器才能量到 10 席。
//
// 突變驗證：把 counters.go 的「無 scope 走 SystemScopeTx」改回「直接在 fallback client 上數」
// → 本檔「滿席 → PLAT-5001」的斷言必須紅（實測見 finalfix-report.md）。
func TestIntegrationAuthSeatGuardUnderAppRole(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateAuthBusinessUp(t, dsn)
	admin := authOpenDB(t, dsn)

	seedSeatPlans(t, admin)

	// 滿席：10 席（上限 10）。identifier 同時是 hd → OIDC 首次登入會解析到這家公司。
	full := authInsertCompany(t, admin, "滿席公司", "full.example.com")
	for i := range 10 {
		authInsertStaffAtCompany(t, admin, full, fmt.Sprintf("full%d@example.com", i))
	}
	subscribeSeatPlan(t, admin, full, "std", "active")

	// 已訂閱但**方案不含席位** → 走 PLAT-5002（用途不變）。
	noSeatFeature := authInsertCompany(t, admin, "方案不含席位", "nofeature.example.com")
	subscribeSeatPlan(t, admin, noSeatFeature, "nostd", "active")

	// 訂閱不可用 → PLAT-3001。
	suspended := authInsertCompany(t, admin, "停用公司", "suspended.example.com")
	subscribeSeatPlan(t, admin, suspended, "std", "suspended")

	// **沒有訂閱列**（尚未開通計費）→ 不施加限制（F-7）：Plan C 的訂閱指派落地前，每個真實公司
	// 都是這個狀態，擋住它就等於「員工自助註冊與首次 OIDC 登入全被硬擋，只有管理員進得去補訂閱」。
	// 已有 10 席（遠超任何方案上限）就是要證明「不施加限制」不是靠「用量很低」僥倖通過。
	noSub := authInsertCompany(t, admin, "未開通計費", "nosub.example.com")
	for i := range 10 {
		authInsertStaffAtCompany(t, admin, noSub, fmt.Sprintf("nosub%d@example.com", i))
	}

	// **唯一的訂閱列是 cancelled**（spec §5.6 的取消是「期末終止、資料不刪除」）→ 合約不可用
	// （PLAT-3001），**不得**被當成「從未訂閱」而落進 none 的不施加限制（F-8）。同樣擺 10 席，
	// 讓「不施加限制」沒有任何僥倖空間。
	cancelled := authInsertCompany(t, admin, "已取消", "cancelled.example.com")
	for i := range 10 {
		authInsertStaffAtCompany(t, admin, cancelled, fmt.Sprintf("cancelled%d@example.com", i))
	}
	subscribeSeatPlan(t, admin, cancelled, "std", "cancelled")

	// 對照組：已訂閱（active）且**未滿席** → 必須照常成功（否則上面的「被擋」可能只是恆擋）。
	roomy := authInsertCompany(t, admin, "未滿席", "roomy.example.com")
	authInsertStaffAtCompany(t, admin, roomy, "roomy0@example.com")
	subscribeSeatPlan(t, admin, roomy, "std", "active")

	// 池 >1：守衛（無 scope）會另開一條系統範圍交易（AGENTS §9-6；池設 1 會與請求交易互鎖）。
	client := authAppRoleClientN(t, dsn, 4)
	env := newSeatGuardEnv(t, admin, client)

	t.Run("已訂閱且滿席：RegisterComplete 回 PLAT-5001 且不落 users 列", func(t *testing.T) {
		token := env.newRegistrationToken(t, "new@full.example.com")
		_, err := env.registerComplete(t, token, full)
		if connect.CodeOf(err) != connect.CodeFailedPrecondition {
			t.Fatalf("滿席時 RegisterComplete 必須被擋（failed_precondition），got %v", err)
		}
		if got := seatErrorCode(t, err); got != "PLAT-5001" {
			t.Fatalf("ErrorInfo.code = %q；want PLAT-5001（滿席 10/10）", got)
		}
		if n := authCount(t, admin, `SELECT count(*) FROM users WHERE company_users = $1`, full); n != 10 {
			t.Fatalf("被擋後 users 列數必須不變（10），got %d", n)
		}
		if n := authCount(t, admin, `SELECT count(*) FROM users WHERE email = $1`, "new@full.example.com"); n != 0 {
			t.Fatalf("被擋後不得建立帳號，got %d 列", n)
		}
	})

	t.Run("已訂閱且滿席：OIDC 首次登入回 PLAT-5001 且不落 users 列", func(t *testing.T) {
		env.verifier.id = &auth.OIDCIdentity{
			Email: "newbie@full.example.com", Name: "新人", HostedDomain: "full.example.com",
		}
		rec := env.callback(t, "seat-state-full")
		if rec.Code != http.StatusFound {
			t.Fatalf("callback 應 302（回跳登入頁），got %d", rec.Code)
		}
		if loc := rec.Header().Get("Location"); !strings.Contains(loc, "error=PLAT-5001") {
			t.Fatalf("滿席時應回跳 login?error=PLAT-5001，got %q", loc)
		}
		if n := authCount(t, admin, `SELECT count(*) FROM users WHERE company_users = $1`, full); n != 10 {
			t.Fatalf("被擋後 users 列數必須不變（10），got %d", n)
		}
	})

	t.Run("已訂閱但方案不含席位 → PLAT-5002（不得因方案未含而放行）", func(t *testing.T) {
		token := env.newRegistrationToken(t, "new@nofeature.example.com")
		_, err := env.registerComplete(t, token, noSeatFeature)
		if got := seatErrorCode(t, err); got != "PLAT-5002" {
			t.Fatalf("ErrorInfo.code = %q；want PLAT-5002（方案未含此功能）", got)
		}
		if n := authCount(t, admin, `SELECT count(*) FROM users WHERE email = $1`, "new@nofeature.example.com"); n != 0 {
			t.Fatalf("被擋後不得建立帳號，got %d 列", n)
		}
	})

	t.Run("訂閱 status=suspended → PLAT-3001", func(t *testing.T) {
		token := env.newRegistrationToken(t, "new@suspended.example.com")
		_, err := env.registerComplete(t, token, suspended)
		if got := seatErrorCode(t, err); got != "PLAT-3001" {
			t.Fatalf("ErrorInfo.code = %q；want PLAT-3001（合約不可用）", got)
		}
		if n := authCount(t, admin, `SELECT count(*) FROM users WHERE email = $1`, "new@suspended.example.com"); n != 0 {
			t.Fatalf("被擋後不得建立帳號，got %d 列", n)
		}
	})

	// F-8：**唯一一列是 cancelled** 的訂閱 → 兩條建帳號路徑都必須被擋在 PLAT-3001，且不得落列。
	// 突變：把 store 的訂閱讀取還原成 `status <> 'cancelled'`（＝預先濾掉）→ 本子測試必須紅
	// （屆時它會被當成「從未訂閱」→ 不施加限制 → 註冊成功）。
	t.Run("唯一訂閱列 status=cancelled → 兩條路徑都 PLAT-3001 且不落 users 列", func(t *testing.T) {
		before := authCount(t, admin, `SELECT count(*) FROM users WHERE company_users = $1`, cancelled)
		if before != 10 {
			t.Fatalf("前置：已取消公司應有 10 席，got %d", before)
		}
		token := env.newRegistrationToken(t, "new@cancelled.example.com")
		_, err := env.registerComplete(t, token, cancelled)
		if got := seatErrorCode(t, err); got != "PLAT-3001" {
			t.Fatalf("RegisterComplete 應回 PLAT-3001（合約不可用），got %q（err=%v）", got, err)
		}

		env.verifier.id = &auth.OIDCIdentity{
			Email: "newbie@cancelled.example.com", Name: "新人", HostedDomain: "cancelled.example.com",
		}
		rec := env.callback(t, "seat-state-cancelled")
		if loc := rec.Header().Get("Location"); !strings.Contains(loc, "error=PLAT-3001") {
			t.Fatalf("OIDC 首次登入應回跳 login?error=PLAT-3001，got %q", loc)
		}
		if n := authCount(t, admin, `SELECT count(*) FROM users WHERE company_users = $1`, cancelled); n != before {
			t.Fatalf("被擋後 users 列數必須不變（%d），got %d", before, n)
		}

		// 投影與守衛**對 cancelled 必須一致**：平台端投影看得到這份已取消的合約（status=cancelled），
		// 守衛也把同一家公司擋在 PLAT-3001 —— 兩者讀的是同一列（取法逐字相同：優先未取消、
		// 只有全是 cancelled 時才取 cancelled），不會出現「後台顯示已取消、守衛卻不限額」。
		row, _, err := postgresstore.NewAdmin(admin).GetTenant(t.Context(), itoa(cancelled))
		if err != nil {
			t.Fatalf("平台端投影查詢失敗: %v", err)
		}
		if row.Status != "cancelled" || row.PlanCode != "std" {
			t.Fatalf("投影應顯示 status=cancelled／方案 std（與守衛一致），got %+v", *row)
		}
	})

	t.Run("對照組：active 且未滿席 → 照常成功", func(t *testing.T) {
		before := authCount(t, admin, `SELECT count(*) FROM users WHERE company_users = $1`, roomy)
		token := env.newRegistrationToken(t, "new@roomy.example.com")
		if _, err := env.registerComplete(t, token, roomy); err != nil {
			t.Fatalf("未滿席的 active 訂閱必須能完成註冊（否則上面的「被擋」可能只是恆擋），got %v", err)
		}
		if n := authCount(t, admin, `SELECT count(*) FROM users WHERE company_users = $1`, roomy); n != before+1 {
			t.Fatalf("應建立帳號（%d → %d），got %d", before, before+1, n)
		}
	})

	// F-7：**沒有訂閱列＝尚未開通計費** → 不施加限制（建帳號成功），但必須留一行明確 log。
	// 突變：把 none 改回 PLAT-5002 → 本子測試必須紅（實測見 finalfix-report.md）。
	t.Run("無訂閱列（尚未開通計費）→ 不施加限制：建帳號成功且留一行 log", func(t *testing.T) {
		before := authCount(t, admin, `SELECT count(*) FROM users WHERE company_users = $1`, noSub)
		if before != 10 {
			t.Fatalf("前置：無訂閱公司應有 10 席，got %d", before)
		}
		// registration token 也要在同一個 log 捕捉窗內產生，才不會漏掉判定時的 log。
		var logs string
		var err error
		logs = captureLog(t, func() {
			var token string
			token = env.newRegistrationToken(t, "new@nosub.example.com")
			_, err = env.registerComplete(t, token, noSub)
		})
		if err != nil {
			t.Fatalf("無訂閱列（尚未開通計費）不得施加限制，RegisterComplete 應成功，got %v", err)
		}
		if n := authCount(t, admin, `SELECT count(*) FROM users WHERE company_users = $1`, noSub); n != 11 {
			t.Fatalf("應建立帳號（10 → 11），got %d", n)
		}
		if !strings.Contains(logs, "無訂閱列") {
			t.Fatalf("必須留一行明確 log（「無聲地不限制」沒人看得見），got %q", logs)
		}
	})
}

// authInsertStaffAtCompany 以 admin 連線在公司下建一個 staff 帳號（無部門：部門 FK 可空）。
func authInsertStaffAtCompany(t *testing.T, db *sql.DB, companyID int, email string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO users (email, name, role, password_hash, company_users)
		 VALUES ($1, '員工', 'staff', 'x', $2) RETURNING id`, email, companyID).Scan(&id); err != nil {
		t.Fatalf("建使用者 %s: %v", email, err)
	}
	return id
}

// seatErrorCode 由 connect error 取 ErrorInfo 的註冊碼（本檔一律以碼斷言：那才是前端據以導向
// 升級方案／收款的依據）。
func seatErrorCode(t *testing.T, err error) string {
	t.Helper()
	if err == nil {
		t.Fatal("應回錯誤，got nil")
	}
	ce, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("應為 *connect.Error，got %T（%v）", err, err)
	}
	for _, d := range ce.Details() {
		v, derr := d.Value()
		if derr != nil {
			t.Fatalf("detail 取值失敗: %v", derr)
		}
		if info, ok := v.(*v1.ErrorInfo); ok {
			return info.GetCode()
		}
	}
	t.Fatalf("錯誤未帶 ErrorInfo（%v）", err)
	return ""
}

// seatGuardEnv 為本檔的測試環境：真 AuthHandler（app_rw ＋ 請求交易）＋ 真 entitlements.Service
// （postgres store ＋ 真計數器）。**不注入 Unlimited／fake guard**：那正是本檔要排除的東西。
type seatGuardEnv struct {
	handler  *handlers.AuthHandler
	rpc      salesorderv1connect.AuthServiceClient
	oneTime  *auth.OneTimeStore
	verifier *seatFakeVerifier
	sessions *scs.SessionManager
	srv      *httptest.Server
}

// seatFakeVerifier／seatFakeExchanger 只負責把 OIDC 身分餵進 callback（Google 不在測試範圍）。
type seatFakeVerifier struct{ id *auth.OIDCIdentity }

func (f *seatFakeVerifier) VerifyIDToken(context.Context, string) (*auth.OIDCIdentity, error) {
	return f.id, nil
}

type seatFakeExchanger struct{}

func (seatFakeExchanger) Exchange(context.Context, string) (string, error) {
	return "fake-id-token", nil
}

func newSeatGuardEnv(t *testing.T, admin *sql.DB, client *ent.Client) *seatGuardEnv {
	t.Helper()
	kv := auth.NewMemoryStore()
	oneTime := auth.NewOneTimeStore(kv)
	cfg := &config.Config{}
	cfg.Auth.FrontendURL = "http://localhost:3000"
	entSvc := entitlements.New(postgresstore.New(admin), services.NewEntitlementCounter(client),
		entitlements.NewMemoryCache(), 0) // ttl=0：不快取，每次判定都回源（也讓 F-7 的 log 每次都出）
	verifier := &seatFakeVerifier{}
	sessions := auth.WebSessionManager(memstore.New(), 24*time.Hour, false, "lax")

	h := handlers.NewAuthHandler(handlers.AuthDeps{
		Cfg:      cfg,
		DB:       client,
		Tokens:   auth.NewTokenManager("seat-guard-secret", kv, client),
		Lockout:  auth.NewLoginLock(kv),
		OneTime:  oneTime,
		Sessions: sessions,
		// 與生產同構（domains.go 的 mountAuth）：真判定服務注入 auth handler。
		Entitlements: entSvc,
	})
	h.SetOIDC(auth.NewGoogleOAuthConfig("cid", "csecret", "http://localhost:3080/cb"), seatFakeExchanger{}, verifier)

	// 與生產一致掛 dbtenant.HandlerOption（請求交易）；**不注入任何身分／RLS scope** —— 這兩條
	// 路徑在生產就是無身分：請求交易裡沒有 scope，正是 F-6 的觸發條件。
	path, connectHandler := salesorderv1connect.NewAuthServiceHandler(h, dbtenant.HandlerOption(client))
	mux := http.NewServeMux()
	mux.Handle(path, connectHandler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return &seatGuardEnv{
		handler: h, rpc: salesorderv1connect.NewAuthServiceClient(http.DefaultClient, srv.URL),
		oneTime: oneTime, verifier: verifier, sessions: sessions, srv: srv,
	}
}

// newRegistrationToken 在一次性 store 放一個 registration token（比照 issueRegistration）。
func (e *seatGuardEnv) newRegistrationToken(t *testing.T, email string) string {
	t.Helper()
	token, err := auth.NewRegistrationToken()
	if err != nil {
		t.Fatalf("NewRegistrationToken: %v", err)
	}
	if err := e.oneTime.Put(t.Context(), auth.RegistrationKey(token), email, 5*time.Minute); err != nil {
		t.Fatalf("寫入 registration token: %v", err)
	}
	return token
}

// registerComplete 以 registration token 完成註冊（header 帶憑證，等同 App 的流程）。
func (e *seatGuardEnv) registerComplete(t *testing.T, token string, companyID int) (*connect.Response[v1.RegisterCompleteResponse], error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), authTimeout)
	defer cancel()
	req := connect.NewRequest(&v1.RegisterCompleteRequest{CompanyId: itoa(companyID), Name: "新人"})
	req.Header().Set("X-Registration-Token", token)
	return e.rpc.RegisterComplete(ctx, req)
}

// callback 走完整的 OIDC callback（state 已寫入一次性 store）。
func (e *seatGuardEnv) callback(t *testing.T, state string) *httptest.ResponseRecorder {
	t.Helper()
	if err := e.oneTime.Put(t.Context(), auth.StateKey(state), "1", 5*time.Minute); err != nil {
		t.Fatalf("寫入 state: %v", err)
	}
	rec := httptest.NewRecorder()
	// 生產的 OIDC 路由也掛 sessions.LoadAndSave（成功登入時要寫 web session）。
	e.sessions.LoadAndSave(http.HandlerFunc(e.handler.GoogleCallback)).ServeHTTP(rec,
		httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/callback?code=testcode&state="+state, nil))
	return rec
}

// authAppRoleClientN 以 app_rw（NOBYPASSRLS）建立業務 client；maxConns > 1 是必要的：
// 無 scope 的守衛會另開一條系統範圍交易（見 counters.go），池設 1 會與請求交易互鎖。
func authAppRoleClientN(t *testing.T, adminDSN string, maxConns int) *ent.Client {
	t.Helper()
	pool := authOpenDB(t, testsupport.AppRoleDSN(t, adminDSN))
	pool.SetMaxOpenConns(maxConns)
	client := dbtenant.NewClient(pool)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// seedSeatPlans 種下平台域夾具（admin 連線）：一個 integer feature（席位）與兩個方案 ——
// std 含席位上限 10、nostd **不含**席位（用來釘 PLAT-5002）。
func seedSeatPlans(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec(`INSERT INTO platform.features (code, type, unit) VALUES ($1,'integer','席')`,
		entitlements.LimitSeats); err != nil {
		t.Fatalf("seed feature: %v", err)
	}
	planID := func(code, name string, sortOrder int) int64 {
		var id int64
		if err := db.QueryRow(`INSERT INTO platform.plans (code, name, sort_order)
			VALUES ($1,$2,$3) RETURNING id`, code, name, sortOrder).Scan(&id); err != nil {
			t.Fatalf("seed 方案 %s: %v", code, err)
		}
		return id
	}
	stdID := planID("std", "標準", 1)
	planID("nostd", "無席位方案", 2)
	if _, err := db.Exec(`INSERT INTO platform.plan_entitlements (plan_id, feature_code, enabled, limit_value)
		VALUES ($1,$2,true,10)`, stdID, entitlements.LimitSeats); err != nil {
		t.Fatalf("seed 權益: %v", err)
	}
}

// subscribeSeatPlan 為公司建立一筆訂閱（admin 連線；platform schema 對 app_rw 零權限）。
func subscribeSeatPlan(t *testing.T, db *sql.DB, companyID int, planCode, status string) {
	t.Helper()
	var planID int64
	if err := db.QueryRow(`SELECT id FROM platform.plans WHERE code = $1`, planCode).Scan(&planID); err != nil {
		t.Fatalf("取方案 %s: %v", planCode, err)
	}
	if _, err := db.Exec(`INSERT INTO platform.subscriptions (company_id, plan_id, status)
		VALUES ($1,$2,$3)`, companyID, planID, status); err != nil {
		t.Fatalf("seed 訂閱（公司 %d）: %v", companyID, err)
	}
}

// captureLog 在 fn 執行期間收集標準 log 輸出（供斷言「無訂閱列 → 不施加限制」的那一行）。
func captureLog(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)
	fn()
	return buf.String()
}
