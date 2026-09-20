//go:build integration

// 配額守衛的真 PostgreSQL 端到端測試（T6）：真 PG ＋ app_rw（00022 的 NOBYPASSRLS 業務角色）
// ＋ 真 RLS ＋ 真 entitlements.Service（store.Fake ＋ NewEntitlementCounter）＋ 真 service handler。
//
// 為何 sqlite 不夠（entitlement_guard_test.go 只能驗語意）：sqlite（enttest）沒有 RLS，
// 計數器看得到全部列，**量不出**「請求 scope 把可見範圍縮小」造成的低報。而配額是公司層的
// （spec §3.2 的 WHERE company_id=?），dept_admin／staff 的請求 scope 是 department，卻可以
// 建立客戶 → 就著請求交易數會把 10 位看成 6 位（本部門）→ 超額放行（fail-open）。
// 這正是 T5 修掉的那個洞，本檔的 department scope 案例就是它的端到端證明。
//
// 為什麼一定要 app_rw：容器／測試的 admin 是 superuser，PG 的 superuser 永遠繞過 RLS
// （FORCE 亦然）→ 以 admin 連線「怎麼數都對」，測不出低報。
package services

import (
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1/customersv1connect"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationSeatGuard 席位上限在真 PG ＋ RLS 下端到端生效：公司 10 席（上限 10）→
// CreateUser 回 failed_precondition ＋ PLAT-5001 且不落庫；停用一席釋放配額後可再建。
func TestIntegrationSeatGuard(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)
	admin := openRawDB(t, adminDSN)

	co := insertRLSCompany(t, admin, "席次公司", "GUARD-SEAT")
	// 稽核列 users 有 FK（00010）→ 操作者必須是真實使用者；他也佔一席。
	actor := insertGuardUser(t, admin, co, 0, "guard-seat-actor@example.com", "company_admin")
	for i := range 9 {
		insertCounterUser(t, admin, co, fmt.Sprintf("guard-seat%d@example.com", i), "active", 0)
	}
	if n := countRows(t, admin, `SELECT count(*) FROM users WHERE company_users = $1 AND status <> 'inactive'`, co); n != 10 {
		t.Fatalf("前置：公司應有 10 席，got %d", n)
	}

	// app_rw ＋ dbtenant.NewClient：連線池 >1 —— 守衛在非 company scope 會另開一條系統範圍
	// 交易（AGENTS §9-6；池設 1 會與請求交易互鎖）。
	client := openAppRoleEntClient(t, adminDSN)
	entSvc := guardEntitlements(t, client, co, entitlements.LimitSeats, 10)
	id := authz.Identity{
		UserID: itoa(actor), CompanyID: itoa(co), Role: "company_admin", Roles: []string{"company_admin"},
	}
	uc := salesorderv1connect.NewUserServiceClient(http.DefaultClient,
		guardURL(t, client, id, guardCompanyScope(co), entSvc, func(m *http.ServeMux, e entitlementChecker) {
			RegisterUserServices(m, client, e)
		}))

	_, err := uc.CreateUser(t.Context(), connect.NewRequest(&v1.CreateUserRequest{
		Name: "第 11 人", Email: "guard-over@example.com", CompanyId: itoa(co), Role: "staff",
	}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("已達席位上限應回 failed_precondition（引導升級方案），got %v", err)
	}
	if got := errorInfoOf(t, err).GetCode(); got != "PLAT-5001" {
		t.Fatalf("席位超限必須帶 ErrorInfo.code=PLAT-5001，got %q", got)
	}
	if n := countRows(t, admin, `SELECT count(*) FROM users WHERE email = 'guard-over@example.com'`); n != 0 {
		t.Fatalf("被擋後不得落庫，got %d 筆", n)
	}

	// 停用一席即釋放配額（spec §3.2 規則 1）→ 可再建（此時真的走完整寫入路徑）。
	if _, err := admin.Exec(`UPDATE users SET status = 'inactive' WHERE company_users = $1 AND email = 'guard-seat0@example.com'`, co); err != nil {
		t.Fatalf("停用一席: %v", err)
	}
	if _, err := uc.CreateUser(t.Context(), connect.NewRequest(&v1.CreateUserRequest{
		Name: "遞補", Email: "guard-refill@example.com", CompanyId: itoa(co), Role: "staff",
	})); err != nil {
		t.Fatalf("停用釋放席位後應可建立: %v", err)
	}
	if n := countRows(t, admin, `SELECT count(*) FROM users WHERE email = 'guard-refill@example.com'`); n != 1 {
		t.Fatalf("遞補帳號應已落庫，got %d 筆", n)
	}
}

// TestIntegrationDepartmentScopeGuardBlocksOverLimit dept_admin（請求 scope = department）在
// **公司層**超額時必須被擋：公司共 10 位客戶（本部門 6 位、另一部門 4 位），上限 10 → 第 11 位
// （本部門的第 7 位）必須回 failed_precondition ＋ PLAT-5001。
//
// 這是低報 bug 的端到端證明：修好前，計數就著請求交易只會看到本部門 6 位（另 4 位被 RLS
// 擋在部門外）→ 6+1 ≤ 10 → 放行（fail-open，本測試 RED）；修好後走系統範圍交易取公司總數 10
// → 擋下。
func TestIntegrationDepartmentScopeGuardBlocksOverLimit(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)
	admin := openRawDB(t, adminDSN)

	co := insertRLSCompany(t, admin, "客戶公司", "GUARD-CUST")
	setCompanyCodePrefix(t, admin, co, "GD")
	deptA := insertCounterDepartment(t, admin, co, "甲部門", false)
	deptB := insertCounterDepartment(t, admin, co, "乙部門", false)
	actor := insertGuardUser(t, admin, co, deptA, "guard-dept-actor@example.com", "dept_admin")
	rep := insertGuardUser(t, admin, co, deptA, "guard-dept-rep@example.com", "staff")

	// 公司 10 位有效客戶：本部門 6 位（另 1 位軟刪除不佔額度）、另一部門 4 位。
	for i := range 6 {
		insertGuardCustomer(t, admin, co, deptA, fmt.Sprintf("GD10%04d", i), false)
	}
	insertGuardCustomer(t, admin, co, deptA, "GD109999", true)
	for i := range 4 {
		insertGuardCustomer(t, admin, co, deptB, fmt.Sprintf("GD20%04d", i), false)
	}
	if n := countRows(t, admin, `SELECT count(*) FROM customers WHERE company_id = $1 AND deleted_at IS NULL`, co); n != 10 {
		t.Fatalf("前置：公司應有 10 位有效客戶，got %d", n)
	}

	client := openAppRoleEntClient(t, adminDSN)
	entSvc := guardEntitlements(t, client, co, entitlements.LimitCustomers, 10)
	id := authz.Identity{
		UserID: itoa(actor), CompanyID: itoa(co), DepartmentID: itoa(deptA),
		Role: "dept_admin", Roles: []string{"dept_admin", "staff"},
	}
	cc := customersv1connect.NewCustomerServiceClient(http.DefaultClient,
		guardURL(t, client, id, guardDepartmentScope(co, deptA), entSvc, func(m *http.ServeMux, e entitlementChecker) {
			RegisterCustomerServices(m, client, "http://localhost:3000", e)
		}))

	_, err := cc.CreateCustomer(t.Context(), connect.NewRequest(&customersv1.CreateCustomerRequest{
		Name: "超額客戶", DefaultSalesRepId: itoa(rep),
	}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("公司層客戶數已達上限時 dept_admin 不得放行（低報＝超額放行的回歸），got %v", err)
	}
	if got := errorInfoOf(t, err).GetCode(); got != "PLAT-5001" {
		t.Fatalf("客戶超限必須帶 ErrorInfo.code=PLAT-5001，got %q", got)
	}
	if n := countRows(t, admin, `SELECT count(*) FROM customers WHERE company_id = $1 AND deleted_at IS NULL`, co); n != 10 {
		t.Fatalf("被擋後不得新增客戶，got %d", n)
	}
}

// insertGuardUser 以 admin 連線建一位使用者並回傳 id（admin 為 superuser，不受 RLS 約束，
// 故 fixture 不必設 scope）。departmentID=0 表示不屬任何部門。
func insertGuardUser(t *testing.T, db *sql.DB, companyID, departmentID int, email, role string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO users (email, name, role, status, password_hash, company_users, department_users)
		 VALUES ($1, '守衛使用者', $2, 'active', 'x', $3, NULLIF($4, 0)) RETURNING id`,
		email, role, companyID, departmentID).Scan(&id); err != nil {
		t.Fatalf("建使用者 %s: %v", email, err)
	}
	return id
}

// insertGuardCustomer 以 admin 連線建一筆客戶（軟刪除者直接標記 deleted_at）。
func insertGuardCustomer(t *testing.T, db *sql.DB, companyID, departmentID int, code string, softDeleted bool) {
	t.Helper()
	deletedAt := "NULL"
	if softDeleted {
		deletedAt = "now()"
	}
	if _, err := db.Exec(
		`INSERT INTO customers (company_id, department_id, customer_code, name, deleted_at)
		 VALUES ($1, $2, $3, '客戶', `+deletedAt+`)`, companyID, departmentID, code); err != nil {
		t.Fatalf("建客戶 %s（公司 %d）: %v", code, companyID, err)
	}
}
