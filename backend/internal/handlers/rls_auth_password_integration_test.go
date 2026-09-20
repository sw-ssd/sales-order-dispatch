//go:build integration

package handlers_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alexedwards/scs/v2/memstore"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx 的 database/sql driver
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/handlers"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// 本檔為 A3(改密碼／臨時密碼)在 **RLS 真正生效** 下的探針:
//
//  1. `audit_logs` 由 00027 ENABLE + FORCE 之後,未帶 scope 的稽核寫入會被 WITH CHECK 擋下。
//     改密碼／重發臨時密碼是**唯一**在 handler 內自開交易寫稽核的生產路徑(auth_password.go),
//     收斂前它以 `h.deps.DB.Tx(ctx)` 另開交易;收斂後一律走請求交易(dbtenant.TxFrom)。
//  2. 客戶(customer)的 data_scope=self、部門管理員(dept_admin)的 data_scope=department ——
//     這兩種 scope 的稽核寫入在 00027 的政策修正前會被擋,整個 RPC 連同業務寫入一起失敗。
//     sqlite 的 handler 測試(enttest)沒有 RLS 語意,對此毫無鑑別力;必須以 app_rw(非 superuser,
//     PG 的 superuser 恆繞過 RLS)＋真 handler 走才有意義。
//
// 連線池上限 1(見 newAuthAppRoleServer)是第二個斷言:請求交易佔住唯一連線,任何自開交易或
// 繞過請求交易的存取都取不到連線 → 該 RPC 逾時(見 callAuth)。這是「全部 DB 存取都在請求交易內」
// 的直接證據。
func TestIntegrationChangePasswordAuditUnderAppRole(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateAuthBusinessUp(t, adminDSN)

	admin := authOpenDB(t, adminDSN)
	defer func() { _ = admin.Close() }()

	coA := authInsertCompany(t, admin, "A", "A3-A")
	deptA := authInsertDepartment(t, admin, coA, "部門A")
	uid := authInsertCustomer(t, admin, coA, deptA, "cust-a3@example.com", "OldPass123")
	// 負向對照用獨立帳號:兩個子測試互不依賴(RED 時不會一個失敗拖垮另一個的假設)。
	uidNoScope := authInsertCustomer(t, admin, coA, deptA, "cust-a3-noscope@example.com", "OldPass456")

	client := authAppRoleClient(t, adminDSN)
	id := authz.Identity{UserID: itoa(uid), CompanyID: itoa(coA), DepartmentID: itoa(deptA), Role: "customer", Roles: []string{"customer"}}
	scope := auth.RLSScope{UserID: itoa(uid), CompanyID: itoa(coA), DepartmentID: itoa(deptA), DataScope: auth.DataScopeSelf, CompanyActive: true}
	noScopeID := authz.Identity{UserID: itoa(uidNoScope), CompanyID: itoa(coA), DepartmentID: itoa(deptA), Role: "customer", Roles: []string{"customer"}}

	t.Run("改密碼成功且稽核列落地(scope=self)", func(t *testing.T) {
		svc := newAuthAppRoleServer(t, client, id, scope, true)
		if err := callAuthErr(t, func(ctx context.Context) error {
			_, err := svc.ChangePassword(ctx, connect.NewRequest(&v1.ChangePasswordRequest{
				OldPassword: "OldPass123", NewPassword: "NewPass123456",
			}))
			return err
		}); err != nil {
			t.Fatalf("ChangePassword(app_rw + self 範圍): %v", err)
		}
		// 業務寫入落地:hash 換成新密碼、must_change 清空、token_version +1。
		var hash string
		var mustChange bool
		var tv int
		if err := admin.QueryRow(
			`SELECT password_hash, must_change_password, token_version FROM users WHERE id = $1`, uid).
			Scan(&hash, &mustChange, &tv); err != nil {
			t.Fatalf("讀使用者: %v", err)
		}
		if !auth.VerifyPassword(hash, "NewPass123456") || mustChange || tv != 1 {
			t.Fatalf("改密碼後應為新 hash / must_change=false / tv=1,got verify=%v must_change=%v tv=%d",
				auth.VerifyPassword(hash, "NewPass123456"), mustChange, tv)
		}
		// 稽核列(D18 同交易)確實落地,且不含任何密碼材料。
		if n := authCount(t, admin,
			`SELECT count(*) FROM audit_logs
			  WHERE action = 'update' AND resource_type = 'user' AND resource_id = $1 AND company_id = $2 AND user_id = $3`,
			itoa(uid), coA, uid); n != 1 {
			t.Fatalf("改密碼應留下恰 1 列稽核(resource_id=%d, 公司 %d),got %d", uid, coA, n)
		}
		var afterSnap string
		if err := admin.QueryRow(
			`SELECT after_snapshot::text FROM audit_logs WHERE resource_type = 'user' AND resource_id = $1`, itoa(uid)).
			Scan(&afterSnap); err != nil {
			t.Fatalf("讀稽核快照: %v", err)
		}
		if !contains(afterSnap, `"must_change_password": false`) {
			t.Fatalf("稽核快照應記錄 must_change_password=false,got %s", afterSnap)
		}
		if n := authCount(t, admin,
			`SELECT count(*) FROM audit_logs
			  WHERE COALESCE(before_snapshot::text,'') LIKE '%'||$1||'%'
			     OR COALESCE(after_snapshot::text,'') LIKE '%'||$1||'%'`, "NewPass123456"); n != 0 {
			t.Fatalf("稽核不得落盤任何密碼材料,got %d 列含新密碼", n)
		}
	})

	t.Run("未注入 scope → 整筆失敗且不落地(fail-closed)", func(t *testing.T) {
		// 業務寫入與稽核同一交易:稽核被 WITH CHECK 擋 → 連業務寫入一起回滾(不會「改成功但無稽核」)。
		var beforeHash string
		var beforeTV int
		if err := admin.QueryRow(`SELECT password_hash, token_version FROM users WHERE id = $1`, uidNoScope).
			Scan(&beforeHash, &beforeTV); err != nil {
			t.Fatalf("讀使用者: %v", err)
		}
		noScope := newAuthAppRoleServer(t, client, noScopeID, auth.RLSScope{}, false)
		if err := callAuthErr(t, func(ctx context.Context) error {
			_, err := noScope.ChangePassword(ctx, connect.NewRequest(&v1.ChangePasswordRequest{
				OldPassword: "OldPass456", NewPassword: "Another123456",
			}))
			return err
		}); err == nil {
			t.Fatal("未注入 scope 時改密碼必須失敗(稽核寫入被 WITH CHECK 擋)")
		}
		var hash string
		var tv int
		if err := admin.QueryRow(`SELECT password_hash, token_version FROM users WHERE id = $1`, uidNoScope).
			Scan(&hash, &tv); err != nil {
			t.Fatalf("讀使用者: %v", err)
		}
		if hash != beforeHash || tv != beforeTV {
			t.Fatalf("被擋下的請求不得改動資料:hash 變了=%v token_version %d→%d", hash != beforeHash, beforeTV, tv)
		}
		if n := authCount(t, admin,
			`SELECT count(*) FROM audit_logs WHERE resource_type = 'user' AND resource_id = $1`, itoa(uidNoScope)); n != 0 {
			t.Fatalf("被擋下的請求不得新增稽核列,got %d", n)
		}
	})
}

// TestIntegrationResetCustomerPasswordAuditUnderAppRole A3 1.5.4 的臨時密碼簽發在 RLS 生效下:
// dept_admin(scope=department)為同部門客戶重發臨時密碼 → 必須成功且稽核列落地(company_id 取
// **目標**所屬公司,符稽核追溯語意);跨部門/跨公司目標一律 permission_denied 且不落任何稽核。
func TestIntegrationResetCustomerPasswordAuditUnderAppRole(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateAuthBusinessUp(t, adminDSN)

	admin := authOpenDB(t, adminDSN)
	defer func() { _ = admin.Close() }()

	coA := authInsertCompany(t, admin, "A", "A4-A")
	coB := authInsertCompany(t, admin, "B", "A4-B")
	deptA := authInsertDepartment(t, admin, coA, "部門A")
	deptB := authInsertDepartment(t, admin, coB, "部門B")
	actor := authInsertUser(t, admin, coA, deptA, "da-a4@example.com", "dept_admin")
	target := authInsertCustomer(t, admin, coA, deptA, "cust-a4@example.com", "OldPass123")
	otherCo := authInsertCustomer(t, admin, coB, deptB, "cust-a4-b@example.com", "OldPass123")

	client := authAppRoleClient(t, adminDSN)
	id := authz.Identity{UserID: itoa(actor), CompanyID: itoa(coA), DepartmentID: itoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin"}}
	scope := auth.RLSScope{UserID: itoa(actor), CompanyID: itoa(coA), DepartmentID: itoa(deptA), DataScope: auth.DataScopeDepartment, CompanyActive: true}
	svc := newAuthAppRoleServer(t, client, id, scope, true)

	// 同部門(同一身分所在的部門)→ 成功。
	temp := callAuth(t, func(ctx context.Context) (string, error) {
		res, err := svc.ResetCustomerPassword(ctx, connect.NewRequest(&v1.ResetCustomerPasswordRequest{UserId: itoa(target)}))
		if err != nil {
			return "", err
		}
		return res.Msg.GetTempPassword(), nil
	})
	if len(temp) < 12 {
		t.Fatalf("臨時密碼長度應 ≥ 12,got %q", temp)
	}
	var hash string
	var mustChange bool
	if err := admin.QueryRow(
		`SELECT password_hash, must_change_password FROM users WHERE id = $1`, target).Scan(&hash, &mustChange); err != nil {
		t.Fatalf("讀目標使用者: %v", err)
	}
	if !mustChange || !auth.VerifyPassword(hash, temp) {
		t.Fatalf("重發後應 must_change=true 且 hash 對應臨時密碼,got must_change=%v verify=%v", mustChange, auth.VerifyPassword(hash, temp))
	}
	// 稽核列:company_id 取目標公司、resource_id 為目標、操作者為 actor。
	if n := authCount(t, admin,
		`SELECT count(*) FROM audit_logs
		  WHERE action = 'update' AND resource_type = 'user' AND resource_id = $1 AND company_id = $2 AND user_id = $3`,
		itoa(target), coA, itoa(actor)); n != 1 {
		t.Fatalf("重發臨時密碼應留下恰 1 列稽核(resource_id=%d, 公司 %d, actor=%d),got %d", target, coA, actor, n)
	}
	var afterSnap string
	if err := admin.QueryRow(
		`SELECT after_snapshot::text FROM audit_logs WHERE resource_type = 'user' AND resource_id = $1`, itoa(target)).
		Scan(&afterSnap); err != nil {
		t.Fatalf("讀稽核快照: %v", err)
	}
	if !contains(afterSnap, `"force_reset": true`) {
		t.Fatalf("稽核快照應記錄 force_reset=true,got %s", afterSnap)
	}
	if n := authCount(t, admin,
		`SELECT count(*) FROM audit_logs
		  WHERE COALESCE(before_snapshot::text,'') LIKE '%'||$1||'%'
		     OR COALESCE(after_snapshot::text,'') LIKE '%'||$1||'%'`, temp); n != 0 {
		t.Fatalf("稽核不得落盤臨時密碼明文,got %d 列", n)
	}

	// 跨公司(同為 deptB 的客戶)→ 必拒且不得落地任何稽核或改動。
	// 錯誤碼為 not_found(單一碼):00028(核心表 ENABLE+FORCE)之後目標讀取先被 RLS 過濾,服務層的
	// resetScopeOK 根本沒機會回報「權限不足」——這是與 T5–T7 一致的既有慣例(跨租戶讀取 → NotFound),
	// 也比 permission_denied 少洩漏「該帳號存在」。
	// resetScopeOK 的 permission_denied 分支由 sqlite 單元測試守著(無 RLS 語意,ACL 是唯一閘門):
	// TestResetCustomerPasswordScopeDenied 的 staff 與「跨公司 company_admin」兩個案例。
	err := callAuthErr(t, func(ctx context.Context) error {
		_, err := svc.ResetCustomerPassword(ctx, connect.NewRequest(&v1.ResetCustomerPasswordRequest{UserId: itoa(otherCo)}))
		return err
	})
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("跨公司重發應被 RLS 擋成 not_found,got %v", err)
	}
	if n := authCount(t, admin, `SELECT count(*) FROM audit_logs WHERE resource_id = $1`, itoa(otherCo)); n != 0 {
		t.Fatalf("被擋下的重發不得留下稽核列,got %d", n)
	}
}

// newAuthAppRoleServer 以業務 client(app_rw + dbtenant.NewClient)掛載真 AuthService handler,
// 與生產一致地掛上 dbtenant.HandlerOption(請求層租戶交易;見 internal/server/domains.go:70)。
// withScope=false 模擬「漏掛 RLS scope」(負向對照);生產的 scope 由 authzMiddleware 依身分導出。
func newAuthAppRoleServer(t *testing.T, client *ent.Client, id authz.Identity, scope auth.RLSScope, withScope bool) salesorderv1connect.AuthServiceClient {
	t.Helper()
	kv := auth.NewMemoryStore()
	h := handlers.NewAuthHandler(handlers.AuthDeps{
		DB:       client,
		Tokens:   auth.NewTokenManager("test-secret", kv, client),
		Lockout:  auth.NewLoginLock(kv),
		OneTime:  auth.NewOneTimeStore(kv),
		Sessions: auth.WebSessionManager(memstore.New(), 24*time.Hour, false, "lax"),
		// 本檔驗的是密碼重設路徑（不建帳號）：席位守衛注入 Unlimited 等同不受配額限制，
		// 語意與注入前相同（nil 是 fail-closed，會擋住建帳號）。
		Entitlements: entitlements.Unlimited(),
	})
	path, handler := salesorderv1connect.NewAuthServiceHandler(h, dbtenant.HandlerOption(client))
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = audit.WithMeta(ctx, audit.Meta{IP: "10.0.0.1", UserAgent: "t8-probe"})
		if withScope {
			ctx = auth.WithRLS(ctx, scope)
		}
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return salesorderv1connect.NewAuthServiceClient(http.DefaultClient, ts.URL)
}

// authAppRoleClient 建立 app_rw 的業務 client:必須經 dbtenant.NewClient(RLS 裝飾器才生效),
// 且連線池上限 1 —— 請求交易佔住唯一連線,任何自開交易/未收斂存取都會卡住並在上限時間內逾時。
func authAppRoleClient(t *testing.T, adminDSN string) *ent.Client {
	t.Helper()
	pool := authOpenDB(t, testsupport.AppRoleDSN(t, adminDSN))
	pool.SetMaxOpenConns(1)
	client := dbtenant.NewClient(pool)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// authTimeout 為單一 RPC 的上限:逾時即代表有存取卡在請求交易外(見檔頭說明)。
const authTimeout = 15 * time.Second

// callAuth 以單一 RPC 的 deadline 執行 fn;錯誤即測試失敗。
func callAuth[T any](t *testing.T, fn func(ctx context.Context) (T, error)) T {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), authTimeout)
	defer cancel()
	v, err := fn(ctx)
	if err != nil {
		t.Fatalf("RPC 失敗(若為 deadline 逾期,代表有存取沒走請求交易): %v", err)
	}
	return v
}

// callAuthErr 以單一 RPC 的 deadline 執行 fn,錯誤交由呼叫端斷言。
func callAuthErr(t *testing.T, fn func(ctx context.Context) error) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), authTimeout)
	defer cancel()
	return fn(ctx)
}

// migrateAuthBusinessUp 以與 cmd/migrate 相同路徑套用業務遷移(同 dialect、同目錄、同版本表)。
func migrateAuthBusinessUp(t *testing.T, dsn string) {
	t.Helper()
	db := authOpenDB(t, dsn)
	defer func() { _ = db.Close() }()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	if err := goose.RunContext(t.Context(), "up", db, "../../database/migrations"); err != nil {
		t.Fatalf("套用遷移: %v", err)
	}
}

// authOpenDB 開一條 database/sql 連線(admin/owner;superuser 不受 RLS 約束,供夾具與真值查詢)。
func authOpenDB(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// authInsertCompany 以 admin 連線建一間公司並回傳 id(companies 無 created_at/updated_at)。
func authInsertCompany(t *testing.T, db *sql.DB, name, identifier string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO companies (name, identifier, status) VALUES ($1, $2, 'active') RETURNING id`,
		name, identifier).Scan(&id); err != nil {
		t.Fatalf("建公司 %s: %v", name, err)
	}
	return id
}

// authInsertDepartment 以 admin 連線建一間部門並回傳 id。
func authInsertDepartment(t *testing.T, db *sql.DB, companyID int, name string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO departments (name, company_departments) VALUES ($1, $2) RETURNING id`, name, companyID).Scan(&id); err != nil {
		t.Fatalf("建部門 %s: %v", name, err)
	}
	return id
}

// authInsertUser 以 admin 連線建一般(非客戶)帳號並回傳 id。
func authInsertUser(t *testing.T, db *sql.DB, companyID, departmentID int, email, role string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO users (email, name, role, password_hash, company_users, department_users)
		 VALUES ($1, '部門管理員', $2, 'x', $3, $4) RETURNING id`, email, role, companyID, departmentID).Scan(&id); err != nil {
		t.Fatalf("建使用者 %s: %v", email, err)
	}
	return id
}

// authInsertCustomer 以 admin 連線建一個「臨時密碼待改」的客戶帳號(hash 為 password 的 bcrypt)。
func authInsertCustomer(t *testing.T, db *sql.DB, companyID, departmentID int, email, password string) int {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("雜湊密碼: %v", err)
	}
	var id int
	if err := db.QueryRow(
		`INSERT INTO users (email, name, role, is_customer, password_hash, company_users, department_users,
		                    must_change_password, temp_password_expires_at)
		 VALUES ($1, '客戶', 'customer', true, $2, $3, $4, true, now() + interval '1 hour') RETURNING id`,
		email, hash, companyID, departmentID).Scan(&id); err != nil {
		t.Fatalf("建客戶帳號 %s: %v", email, err)
	}
	return id
}

// authCount 以 admin 連線取單一純量(真值,不受 RLS 影響)。
func authCount(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("查詢 %q: %v", query, err)
	}
	return n
}

// contains 為字串包含(僅供本檔的稽核快照斷言使用)。
func contains(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) && indexOfStr(haystack, needle) >= 0
}

// indexOfStr 回傳 needle 在 haystack 的索引;不存在回 -1。
func indexOfStr(h, n string) int {
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return i
		}
	}
	return -1
}

// itoa 為整數字串轉換(身分欄位為字串)。
func itoa(i int) string { return strconv.Itoa(i) }
