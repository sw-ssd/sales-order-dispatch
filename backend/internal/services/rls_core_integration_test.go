//go:build integration

package services

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"connectrpc.com/connect"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	authzopenfga "github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
)

// 本檔為核心域(companies / departments / users / roles / role_permissions)＋ 依賴它們最重的
// 三支 service(company / user / role)在 00028 ENABLE + FORCE 之後的探針。
//
// 為何核心域要單獨一組探針:
//   - 這五張表被幾乎所有服務讀取(每個 RPC 的身分解析、稽核 FK、角色查詢都會摸到),任何一條
//     未收斂的存取在 RLS 生效後就是全站黑屏;而 sqlite 單元測試(無 RLS 語意)與 superuser 連線
//     (PG 的 superuser 永遠繞過 RLS,FORCE 亦然)對它完全沒有鑑別力。
//   - 四支測試分別釘住四個不同面:
//     ① TestIntegrationRLSCoreIsolation:app_rw 直連的 scope 矩陣(all/company/department/self
//     × 讀/寫)＋ 未設 scope 的 fail-closed;
//     ② TestIntegrationCoreServicesUnderAppRole:真 handler × 請求層租戶交易,逐條走過三支
//     service 的 List/Get/Create/Update/Delete,每條都斷言回來的列屬身分所屬公司/部門;
//     ③ TestIntegrationCoreWritesStayInRequestTx:**強制 handler 之後失敗**再斷言「業務列與稽核列
//     都沒有落地」—— 這是「9 處自開交易已收斂至請求交易」的唯一直接證據(RLS 可見性對自開交易
//     完全無感:裝飾器對任何 client.Tx(ctx) 都會套 ctx 的 SET LOCAL,功能上與請求交易等價,
//     差別只在 D18 原子性與鎖);
//     ④ TestIntegrationRLSCoreEnableMigrationDown:00028 的 Up/Down 對稱且不得動 policy。

// TestIntegrationRLSCoreIsolation 以 app_rw 直連驗證核心五表的隔離(00028 ENABLE + FORCE):
// 未設 scope → 五表全數 0 列(fail-closed);四種 scope 等級各自看到該看到的列;跨租戶寫入被
// WITH CHECK 擋下、跨租戶 UPDATE 影響 0 列。
//
// 五張表的 policy 語意不同(00023/00025),故斷言逐表寫明而非抽象化:
//   - companies:id = 當前公司(all,或任何帶了 company_id 的 scope 都看得到自己那家)。
//   - departments:all / company+company_id / department+department_id —— **self 無分支**(0 列)。
//   - users:all / company+company_id / department+department_id / self+user_id。
//   - roles / role_permissions:scope 非空即可讀寫(共享目錄,寫入權威在服務層 ACL)。
func TestIntegrationRLSCoreIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()

	coA := insertRLSCompany(t, admin, "A", "CORE-A")
	coB := insertRLSCompany(t, admin, "B", "CORE-B")
	deptA := insertRLSDepartment(t, admin, coA, "A 部門")
	deptB := insertRLSDepartment(t, admin, coB, "B 部門")
	userA := insertRLSUser(t, admin, coA, "core-a@example.com", "staff")
	setRLSUserDepartment(t, admin, userA, deptA)
	userB := insertRLSUser(t, admin, coB, "core-b@example.com", "staff")
	setRLSUserDepartment(t, admin, userB, deptB)
	roleID := insertRLSCoreRole(t, admin, "core_rls_role")
	insertRLSCoreRolePermission(t, admin, roleID, "customer", "read")

	app := openAppRoleDB(t, adminDSN)

	t.Run("未設 scope → 核心五表全數 fail-closed", func(t *testing.T) {
		for _, table := range coreTables {
			if n := countRows(t, app, `SELECT count(*) FROM `+table); n != 0 {
				t.Fatalf("未設 scope 時 %s 必須 0 列(fail-closed),got %d", table, n)
			}
		}
		// 未設 scope 的寫入也必須被擋:不是「查不到」而是根本寫不進去。
		if err := appExecScoped(t, app, nil,
			`INSERT INTO companies (name, identifier, status) VALUES ('X', 'CORE-X', 'active')`); !isRLSPolicyViolation(err) {
			t.Fatalf("未設 scope 的公司寫入必須被 WITH CHECK 擋下(42501),got %v", err)
		}
	})

	t.Run("scope=all → 系統範圍可見可寫全部公司", func(t *testing.T) {
		tx := appTx(t, app, []string{`SET LOCAL app.current_data_scope = 'all'`})
		defer func() { _ = tx.Rollback() }()
		assertCoreCount(t, tx, `SELECT count(*) FROM companies WHERE identifier IN ('CORE-A','CORE-B')`, 2, nil, "companies(all)")
		assertCoreCount(t, tx, `SELECT count(*) FROM departments WHERE name IN ('A 部門','B 部門')`, 2, nil, "departments(all)")
		assertCoreCount(t, tx, `SELECT count(*) FROM users WHERE email IN ('core-a@example.com','core-b@example.com')`, 2, nil, "users(all)")
		assertCoreCount(t, tx, `SELECT count(*) FROM roles WHERE id = $1`, 1, roleID, "roles(all)")
		assertCoreCount(t, tx, `SELECT count(*) FROM role_permissions WHERE role_id = $1`, 1, roleID, "role_permissions(all)")
		// 系統範圍寫得進任何公司(SystemScopeTx 的等價性:未登入路徑靠它)。
		if _, err := tx.Exec(
			`INSERT INTO users (email, name, role, password_hash, company_users)
			 VALUES ('core-all@example.com', '系統範圍', 'staff', 'x', $1)`, coB); err != nil {
			t.Fatalf("scope=all 應可寫入任一公司的使用者: %v", err)
		}
	})

	t.Run("scope=company A → 只見 A 的列", func(t *testing.T) {
		tx := appTx(t, app, companyScope(coA))
		defer func() { _ = tx.Rollback() }()
		assertCoreCount(t, tx, `SELECT count(*) FROM companies WHERE id = $1`, 1, coA, "companies(自己)")
		assertCoreCount(t, tx, `SELECT count(*) FROM companies WHERE id = $1`, 0, coB, "companies(他家)")
		assertCoreCount(t, tx, `SELECT count(*) FROM departments WHERE id = $1`, 1, deptA, "departments(自己)")
		assertCoreCount(t, tx, `SELECT count(*) FROM departments WHERE id = $1`, 0, deptB, "departments(他家)")
		assertCoreCount(t, tx, `SELECT count(*) FROM users WHERE id = $1`, 1, userA, "users(自己)")
		assertCoreCount(t, tx, `SELECT count(*) FROM users WHERE id = $1`, 0, userB, "users(他家)")
		// 共享目錄:任何非空 scope 都可讀。
		assertCoreCount(t, tx, `SELECT count(*) FROM roles WHERE id = $1`, 1, roleID, "roles")
		assertCoreCount(t, tx, `SELECT count(*) FROM role_permissions WHERE role_id = $1`, 1, roleID, "role_permissions")
	})

	t.Run("scope=department A → 只見自己部門的列", func(t *testing.T) {
		tx := appTx(t, app, departmentScope(coA, deptA))
		defer func() { _ = tx.Rollback() }()
		assertCoreCount(t, tx, `SELECT count(*) FROM users WHERE id = $1`, 1, userA, "users(自己部門)")
		assertCoreCount(t, tx, `SELECT count(*) FROM users WHERE id = $1`, 0, userB, "users(他部門)")
		assertCoreCount(t, tx, `SELECT count(*) FROM departments WHERE id = $1`, 1, deptA, "departments(自己部門)")
		assertCoreCount(t, tx, `SELECT count(*) FROM departments WHERE id = $1`, 0, deptB, "departments(他部門)")
		assertCoreCount(t, tx, `SELECT count(*) FROM companies WHERE id = $1`, 1, coA, "companies(自己公司)")
		assertCoreCount(t, tx, `SELECT count(*) FROM role_permissions WHERE role_id = $1`, 1, roleID, "role_permissions")
	})

	t.Run("scope=self → 只見自己那一列,departments 無分支", func(t *testing.T) {
		tx := appTx(t, app, selfScope(coA, userA))
		defer func() { _ = tx.Rollback() }()
		assertCoreCount(t, tx, `SELECT count(*) FROM users WHERE id = $1`, 1, userA, "users(自己)")
		assertCoreCount(t, tx, `SELECT count(*) FROM users WHERE id = $1`, 0, userB, "users(他人)")
		// self 不是「自己部門」:departments 的 policy 沒有 self 分支 → 0 列(fail-closed)。
		assertCoreCount(t, tx, `SELECT count(*) FROM departments WHERE id = $1`, 0, deptA, "departments(self 無分支)")
		assertCoreCount(t, tx, `SELECT count(*) FROM companies WHERE id = $1`, 1, coA, "companies(自己公司)")
		assertCoreCount(t, tx, `SELECT count(*) FROM roles WHERE id = $1`, 1, roleID, "roles(共享目錄)")
	})

	t.Run("跨租戶寫入 → 被 WITH CHECK 擋下", func(t *testing.T) {
		if err := appExecScoped(t, app, companyScope(coA),
			`INSERT INTO users (email, name, role, password_hash, company_users)
			 VALUES ('core-cross@example.com', '跨租戶', 'staff', 'x', $1)`, coB); !isRLSPolicyViolation(err) {
			t.Fatalf("以 A 的身分新增 B 公司的使用者必須被擋(42501),got %v", err)
		}
		if err := appExecScoped(t, app, companyScope(coA),
			`INSERT INTO companies (name, identifier, status) VALUES ('新公司', 'CORE-NEW', 'active')`); !isRLSPolicyViolation(err) {
			t.Fatalf("以 A 的身分新增別家公司必須被擋(42501),got %v", err)
		}
		// 自己公司的寫入必須成功 —— 否則上面兩條只是「全部都寫不進去」的假證據。
		if err := appExecScoped(t, app, companyScope(coA),
			`INSERT INTO users (email, name, role, password_hash, company_users)
			 VALUES ('core-own@example.com', '自己公司', 'staff', 'x', $1)`, coA); err != nil {
			t.Fatalf("以 A 的身分新增 A 公司的使用者應成功,got %v", err)
		}
		// department:同部門可寫、他部門被擋。
		if err := appExecScoped(t, app, departmentScope(coA, deptA),
			`INSERT INTO users (email, name, role, password_hash, company_users, department_users)
			 VALUES ('core-dept@example.com', '自己部門', 'staff', 'x', $1, $2)`, coA, deptA); err != nil {
			t.Fatalf("以 A 部門身分新增同部門使用者應成功,got %v", err)
		}
		if err := appExecScoped(t, app, departmentScope(coA, deptA),
			`INSERT INTO users (email, name, role, password_hash, company_users, department_users)
			 VALUES ('core-dept-x@example.com', '他部門', 'staff', 'x', $1, $2)`, coA, deptB); !isRLSPolicyViolation(err) {
			t.Fatalf("以 A 部門身分新增他部門使用者必須被擋(42501),got %v", err)
		}
		// self:UPDATE 自己那列可過;INSERT(新列 id ≠ user_id)必被擋。
		if n, err := coreExecScoped(t, app, selfScope(coA, userA),
			`UPDATE users SET phone = '0900000000' WHERE id = $1`, userA); err != nil || n != 1 {
			t.Fatalf("self 範圍更新自己那一列應成功且影響 1 列,got n=%d err=%v", n, err)
		}
		if err := appExecScoped(t, app, selfScope(coA, userA),
			`INSERT INTO users (email, name, role, password_hash, company_users)
			 VALUES ('core-self@example.com', 'self 新增', 'staff', 'x', $1)`, coA); !isRLSPolicyViolation(err) {
			t.Fatalf("self 範圍不得新增使用者(新列 id ≠ user_id,42501),got %v", err)
		}
	})

	t.Run("跨租戶 UPDATE → 影響 0 列且真值不變", func(t *testing.T) {
		n, err := coreExecScoped(t, app, companyScope(coA),
			`UPDATE users SET name = '被跨租戶改壞' WHERE id = $1`, userB)
		if err != nil {
			t.Fatalf("跨租戶 UPDATE 在 USING 過濾下應是 0 列(非錯誤),got %v", err)
		}
		if n != 0 {
			t.Fatalf("以 A 的身分更新 B 公司的使用者必須影響 0 列,got %d", n)
		}
		if got := countRows(t, admin, `SELECT count(*) FROM users WHERE id = $1 AND name = '公司管理員'`, userB); got != 1 {
			t.Fatal("B 公司的使用者被跨租戶改動了(admin 真值檢查失敗)")
		}
	})
}

// TestIntegrationCoreServicesUnderAppRole 以 app_rw + 請求層租戶交易跑**真 handler**,逐條走過
// company / department / user / role 四組 RPC 的 List/Get/Create/Update/Delete,每條都斷言
// 「回來的列屬身分所屬公司(或該 scope 該見的集合)」,並斷言寫入路徑的稽核列真的落地(D18)。
//
// 三種身分各代表一種 scope 等級(super=all、company_admin=company、dept_admin=department):
// 「service 收了 dbtenant.Client 但 scope 等級不對」在單一 scope 的探針下看不出來(T8 的教訓)。
func TestIntegrationCoreServicesUnderAppRole(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()

	coA := insertRLSCompany(t, admin, "A", "CORE-SVC-A")
	coB := insertRLSCompany(t, admin, "B", "CORE-SVC-B")
	deptA := insertRLSDepartment(t, admin, coA, "A 部門")
	deptOther := insertRLSDepartment(t, admin, coA, "A 他部門")
	insertRLSDepartment(t, admin, coB, "B 部門")
	superID := insertRLSUser(t, admin, coA, "core-super@example.com", "super")
	adminA := insertRLSUser(t, admin, coA, "core-admin-a@example.com", "company_admin")
	adminB := insertRLSUser(t, admin, coB, "core-admin-b@example.com", "company_admin")
	deptAdminA := insertRLSUser(t, admin, coA, "core-deptadmin-a@example.com", "dept_admin")
	setRLSUserDepartment(t, admin, deptAdminA, deptA)
	memberOther := insertRLSUser(t, admin, coA, "core-other-dept@example.com", "staff")
	setRLSUserDepartment(t, admin, memberOther, deptOther)
	roleID := insertRLSCoreRole(t, admin, "core_svc_role")
	insertRLSCoreRolePermission(t, admin, roleID, "customer", "read")

	client := openAppRoleEntClient(t, adminDSN)
	super := authz.Identity{UserID: itoa(superID), Role: "super", Roles: []string{"super"}}
	superScope := auth.RLSScope{UserID: itoa(superID), DataScope: auth.DataScopeAll, CompanyActive: true}

	// ① super(scope=all):完整生命週期(create/update/delete)+ 稽核落地。
	t.Run("super(all):公司/部門/使用者的完整生命週期與稽核落地", func(t *testing.T) {
		c := newCoreAppRoleServer(t, client, super, superScope, true)

		company := callCore(t, c.companies.CreateCompany, &v1.CreateCompanyRequest{
			Name: "新公司", Identifier: "CORE-SVC-NEW",
		}).Company
		newCo := mustItoa(t, company.GetId())
		updated := callCore(t, c.companies.UpdateCompany, &v1.UpdateCompanyRequest{
			CompanyId: company.GetId(), Name: protoStr("新公司改名"),
		}).Company
		if updated.GetName() != "新公司改名" {
			t.Fatalf("UpdateCompany 未生效:%+v", updated)
		}
		// 稽核的 company_id 取「目標公司」(super 無所屬公司)。
		assertAuditRow(t, admin, "company", "update", company.GetId(), newCo)
		if got := callCore(t, c.companies.GetCompany, &v1.GetCompanyRequest{CompanyId: company.GetId()}).Company; got.GetName() != "新公司改名" {
			t.Fatalf("GetCompany 未回更新後的值:%+v", got)
		}
		if err := callCoreErr(t, c.companies.DeleteCompany, &v1.DeleteCompanyRequest{CompanyId: company.GetId()}); err != nil {
			t.Fatalf("無部門/成員的公司應可軟刪除:%v", err)
		}
		assertCoreSoftDeleted(t, admin, "companies", company.GetId(), true)
		assertAuditRow(t, admin, "company", "delete", company.GetId(), newCo)

		department := callCore(t, c.departments.CreateDepartment, &v1.CreateDepartmentRequest{
			CompanyId: itoa(coA), Name: "SVC 部門",
		}).Department
		if department.GetCompanyId() != itoa(coA) {
			t.Fatalf("CreateDepartment 回傳的公司不符:%+v", department)
		}
		updatedDept := callCore(t, c.departments.UpdateDepartment, &v1.UpdateDepartmentRequest{
			DepartmentId: department.GetId(), Name: protoStr("SVC 部門改名"),
		}).Department
		if updatedDept.GetName() != "SVC 部門改名" {
			t.Fatalf("UpdateDepartment 未生效:%+v", updatedDept)
		}
		// 註:部門的 create/update 沒有稽核(僅 delete 有),故不在此斷言稽核列。
		if got := callCore(t, c.departments.GetDepartment, &v1.GetDepartmentRequest{DepartmentId: department.GetId()}).Department; got.GetName() != "SVC 部門改名" {
			t.Fatalf("GetDepartment 未回更新後的值:%+v", got)
		}

		user := callCore(t, c.users.CreateUser, &v1.CreateUserRequest{
			CompanyId: itoa(coA), DepartmentId: department.GetId(),
			Email: "core-svc-new@example.com", Name: "新員工", Role: "staff",
		}).User
		if user.GetCompanyId() != itoa(coA) || user.GetDepartmentId() != department.GetId() {
			t.Fatalf("CreateUser 回傳的公司/部門不符:%+v", user)
		}
		assertAuditRow(t, admin, "user", "create", user.GetId(), coA)
		if got := callCore(t, c.users.GetUser, &v1.GetUserRequest{UserId: user.GetId()}).User; got.GetEmail() != "core-svc-new@example.com" {
			t.Fatalf("GetUser 未回剛建立的帳號:%+v", got)
		}

		if err := callCoreErr(t, c.users.UpdateUser, &v1.UpdateUserRequest{
			UserId: user.GetId(), Name: protoStr("新員工改名"),
		}); err != nil {
			t.Fatalf("UpdateUser: %v", err)
		}
		assertAuditRow(t, admin, "user", "update", user.GetId(), coA)
		if err := callCoreErr(t, c.users.AssignRole, &v1.AssignRoleRequest{
			UserId: user.GetId(), Role: "dept_admin",
		}); err != nil {
			t.Fatalf("AssignRole: %v", err)
		}
		assertAuditRow(t, admin, "user", "role_change", user.GetId(), coA)
		if err := callCoreErr(t, c.users.ForceLogout, &v1.ForceLogoutRequest{UserId: user.GetId()}); err != nil {
			t.Fatalf("ForceLogout: %v", err)
		}
		assertAuditRow(t, admin, "user", "force_logout", user.GetId(), coA)
		if err := callCoreErr(t, c.users.Deactivate, &v1.DeactivateRequest{UserId: user.GetId()}); err != nil {
			t.Fatalf("Deactivate: %v", err)
		}
		if n := countRows(t, admin, `SELECT count(*) FROM users WHERE id = $1 AND status = 'inactive'`, mustItoa(t, user.GetId())); n != 1 {
			t.Fatal("Deactivate 未把帳號設為 inactive")
		}

		// 部門仍有成員 → 不可刪(軟刪除不變式);成員移出後才可刪。
		err := callCoreErr(t, c.departments.DeleteDepartment, &v1.DeleteDepartmentRequest{DepartmentId: department.GetId()})
		if connect.CodeOf(err) != connect.CodeFailedPrecondition {
			t.Fatalf("部門仍有成員時刪除應 failed_precondition,got %v", err)
		}
		if err := callCoreErr(t, c.users.UpdateUser, &v1.UpdateUserRequest{
			UserId: user.GetId(), DepartmentId: protoStr(itoa(deptA)),
		}); err != nil {
			t.Fatalf("把成員移出部門: %v", err)
		}
		if err := callCoreErr(t, c.departments.DeleteDepartment, &v1.DeleteDepartmentRequest{DepartmentId: department.GetId()}); err != nil {
			t.Fatalf("成員移出後應可刪部門: %v", err)
		}
		assertCoreSoftDeleted(t, admin, "departments", department.GetId(), true)
		assertAuditRow(t, admin, "department", "delete", department.GetId(), coA)

		// RoleService:共享目錄(roles/role_permissions)在任何非空 scope 下可讀寫。
		assertRoleIDPresent(t, callCore(t, c.roles.ListRoles, &v1.ListRolesRequest{}), strconv.Itoa(roleID))
		perms := callCore(t, c.roles.GetRolePermissions, &v1.GetRolePermissionsRequest{RoleId: strconv.Itoa(roleID)}).Permissions
		if len(perms) != 1 {
			t.Fatalf("GetRolePermissions 應回 1 筆,got %d", len(perms))
		}
		if err := callCoreErr(t, c.roles.UpdateRolePermissions, &v1.UpdateRolePermissionsRequest{
			RoleId: strconv.Itoa(roleID),
			Permissions: []*v1.Permission{
				{Resource: "customer", Action: "read", SortOrder: 0},
				{Resource: "customer", Action: "update", SortOrder: 1},
			},
		}); err != nil {
			t.Fatalf("UpdateRolePermissions: %v", err)
		}
		if n := countRows(t, admin, `SELECT count(*) FROM role_permissions WHERE role_id = $1`, roleID); n != 2 {
			t.Fatalf("UpdateRolePermissions 應留下 2 筆權限,got %d", n)
		}
	})

	// ② company_admin(scope=company):List/Get 只能看到自己公司。
	t.Run("company_admin(company):List/Get 僅自己公司", func(t *testing.T) {
		c := newCoreAppRoleServer(t, client, authz.Identity{
			UserID: itoa(adminA), CompanyID: itoa(coA), Role: "company_admin", Roles: []string{"company_admin"},
		}, auth.RLSScope{
			UserID: itoa(adminA), CompanyID: itoa(coA), DataScope: auth.DataScopeCompany, CompanyActive: true,
		}, true)

		companies := callCore(t, c.companies.ListCompanies, &v1.ListCompaniesRequest{}).Companies
		if len(companies) != 1 || companies[0].GetIdentifier() != "CORE-SVC-A" {
			t.Fatalf("company_admin 的 ListCompanies 應僅 [CORE-SVC-A],got %+v", companies)
		}
		err := callCoreErr(t, c.companies.GetCompany, &v1.GetCompanyRequest{CompanyId: itoa(coB)})
		if connect.CodeOf(err) != connect.CodeNotFound {
			t.Fatalf("他公司的 GetCompany 應 NotFound(RLS 過濾),got %v", err)
		}

		departments := callCore(t, c.departments.ListDepartments, &v1.ListDepartmentsRequest{}).Departments
		for _, d := range departments {
			if d.GetCompanyId() != itoa(coA) {
				t.Fatalf("ListDepartments 混入他公司的部門:%+v", d)
			}
		}
		// 明確以他公司 id 篩選也必須回空(RLS 是最後一道閘門)。
		if l := callCore(t, c.departments.ListDepartments, &v1.ListDepartmentsRequest{CompanyId: itoa(coB)}).Departments; len(l) != 0 {
			t.Fatalf("以他公司 id 篩選應回 0 筆,got %d", len(l))
		}

		users := callCore(t, c.users.ListUsers, &v1.ListUsersRequest{}).Users
		if len(users) == 0 {
			t.Fatal("company_admin 應看得到自己公司的成員(非空)")
		}
		for _, u := range users {
			if u.GetCompanyId() != itoa(coA) {
				t.Fatalf("ListUsers 混入他公司的成員:%+v", u)
			}
		}

		// 跨公司的 CreateUser 不得成功(RLS 的 WITH CHECK 是最後一道閘門)。
		err = callCoreErr(t, c.users.CreateUser, &v1.CreateUserRequest{
			CompanyId: itoa(coB), Email: "core-cross-create@example.com", Name: "跨租戶", Role: "staff",
		})
		if err == nil {
			t.Fatal("company_admin 不得在他公司建帳號")
		}
		if n := countRows(t, admin, `SELECT count(*) FROM users WHERE email = 'core-cross-create@example.com'`); n != 0 {
			t.Fatalf("被擋下的建帳號不得落地,got %d 列", n)
		}
	})

	// ③ dept_admin(scope=department):只看得到自己部門的成員。
	t.Run("dept_admin(department):ListUsers 僅自己部門", func(t *testing.T) {
		c := newCoreAppRoleServer(t, client, authz.Identity{
			UserID: itoa(deptAdminA), CompanyID: itoa(coA), DepartmentID: itoa(deptA),
			Role: "dept_admin", Roles: []string{"dept_admin"},
		}, auth.RLSScope{
			UserID: itoa(deptAdminA), CompanyID: itoa(coA), DepartmentID: itoa(deptA),
			DataScope: auth.DataScopeDepartment, CompanyActive: true,
		}, true)

		users := callCore(t, c.users.ListUsers, &v1.ListUsersRequest{}).Users
		if len(users) == 0 {
			t.Fatal("dept_admin 應看得到自己部門的成員(非空)")
		}
		for _, u := range users {
			if u.GetDepartmentId() != itoa(deptA) {
				t.Fatalf("dept_admin 看到不在自己部門的成員:%+v", u)
			}
			if u.GetId() == itoa(memberOther) {
				t.Fatal("dept_admin 看到同公司他部門的成員(RLS 與 ACL 都應排除)")
			}
		}
		// 同公司他部門/他公司的成員:RLS 與 ACL 都不得放行。
		for _, target := range []int{memberOther, adminA, adminB} {
			if err := callCoreErr(t, c.users.GetUser, &v1.GetUserRequest{UserId: itoa(target)}); err == nil {
				t.Fatalf("dept_admin 不得讀取部門外成員 %d", target)
			}
		}
	})

	// ④ 負向對照:同一連線池、不注入 scope → 核心表讀不到任何列、寫入失敗。
	t.Run("未注入 scope → 讀 0 列且寫入失敗(fail-closed)", func(t *testing.T) {
		c := newCoreAppRoleServer(t, client, super, auth.RLSScope{}, false)

		if l := callCore(t, c.companies.ListCompanies, &v1.ListCompaniesRequest{}).Companies; len(l) != 0 {
			t.Fatalf("未注入 scope 時 ListCompanies 必須 0 筆,got %d", len(l))
		}
		if l := callCore(t, c.users.ListUsers, &v1.ListUsersRequest{}).Users; len(l) != 0 {
			t.Fatalf("未注入 scope 時 ListUsers 必須 0 筆,got %d", len(l))
		}
		err := callCoreErr(t, c.companies.CreateCompany, &v1.CreateCompanyRequest{
			Name: "無 scope 公司", Identifier: "CORE-NOSCOPE",
		})
		if err == nil {
			t.Fatal("未注入 scope 時 CreateCompany 必須失敗(WITH CHECK)")
		}
		if n := countRows(t, admin, `SELECT count(*) FROM companies WHERE identifier = 'CORE-NOSCOPE'`); n != 0 {
			t.Fatalf("被擋下的建公司不得落地,got %d 列", n)
		}
	})
}

// coreFailMarker 是 failAfterHandler 回傳的固定訊息:測試以它確認「失敗確實發生在 handler 之後」,
// 而不是服務層自己早一步失敗(那會讓「沒有落地」變成假證據 —— 請求根本沒走到寫入)。
const coreFailMarker = "core-probe: 強制於 handler 之後失敗"

// failAfterHandler 為掛在 dbtenant.Interceptor **內層**的 interceptor:先讓 handler 完成(所有
// 業務寫入與稽核都已於請求交易內執行),再回傳錯誤 → 外層的 dbtenant.Interceptor 看到 err 即
// rollback 整個請求交易。
type failAfterHandler struct{}

func (failAfterHandler) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if _, err := next(ctx, req); err != nil {
			return nil, err
		}
		return nil, connect.NewError(connect.CodeInternal, errors.New(coreFailMarker))
	}
}

func (failAfterHandler) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (failAfterHandler) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

// assertForcedFailure 斷言錯誤確實來自 failAfterHandler:若服務層早一步失敗(例如 RLS 擋下查詢),
// 請求根本沒走到寫入,「沒有落地」就不是收斂的證據。
func assertForcedFailure(t *testing.T, err error) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), coreFailMarker) {
		t.Fatalf("請求必須在 handler 之後被強制失敗(否則「沒有落地」是假證據),got %v", err)
	}
}

// TestIntegrationCoreWritesStayInRequestTx 證明核心域**所有寫入路徑**都在**同一個請求交易**內:
// 以 failAfterHandler 讓每個 RPC 在 handler 成功之後失敗,外層請求交易因此 rollback —— 此時
// 業務列與稽核列必須**全部不落地**。
//
// 為何需要這一條(而不是只斷言 RLS 可見性):裝飾器對任何 client.Tx(ctx) 都會套 ctx 的 SET LOCAL,
// 故「服務自己開的交易」在 RLS 下**功能等價**(一樣看得到、寫得進),只有原子性與鎖不同
// (T8 §6.5 已記錄此盲點)。收斂前這 9 處(s.db.Tx)各自 commit → rollback 對它們無效 → 本測試紅。
func TestIntegrationCoreWritesStayInRequestTx(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()

	coA := insertRLSCompany(t, admin, "A", "CORE-TX-A")
	deptA := insertRLSDepartment(t, admin, coA, "A 部門")
	coEmpty := insertRLSCompany(t, admin, "空", "CORE-TX-EMPTY")
	deptEmpty := insertRLSDepartment(t, admin, coEmpty, "空部門")
	coBare := insertRLSCompany(t, admin, "無部門", "CORE-TX-BARE")
	superID := insertRLSUser(t, admin, coA, "core-tx-super@example.com", "super")
	target := insertRLSUser(t, admin, coA, "core-tx-target@example.com", "staff")
	setRLSUserDepartment(t, admin, target, deptA)
	roleID := insertRLSCoreRole(t, admin, "core_tx_role")
	insertRLSCoreRolePermission(t, admin, roleID, "customer", "read")

	client := openAppRoleEntClient(t, adminDSN)
	super := authz.Identity{UserID: itoa(superID), CompanyID: itoa(coA), Role: "super", Roles: []string{"super"}}
	scope := auth.RLSScope{
		UserID: itoa(superID), CompanyID: itoa(coA), DataScope: auth.DataScopeAll, CompanyActive: true,
	}
	// 每個 RPC 各自一個 server:失敗注入是「該請求」的性質,共用會讓後續請求全數失敗。
	newFailing := func(t *testing.T) coreClients { return newCoreFailingServer(t, client, super, scope) }

	t.Run("CreateCompany 不落地", func(t *testing.T) {
		assertForcedFailure(t, callCoreErr(t, newFailing(t).companies.CreateCompany, &v1.CreateCompanyRequest{
			Name: "回滾公司", Identifier: "CORE-TX-ROLLBACK",
		}))
		if n := countRows(t, admin, `SELECT count(*) FROM companies WHERE identifier = 'CORE-TX-ROLLBACK'`); n != 0 {
			t.Fatalf("請求交易回滾後不得留下公司列,got %d", n)
		}
	})

	t.Run("UpdateCompany(自開交易)不落地且不留稽核", func(t *testing.T) {
		assertForcedFailure(t, callCoreErr(t, newFailing(t).companies.UpdateCompany, &v1.UpdateCompanyRequest{
			CompanyId: itoa(coA), Name: protoStr("回滾改名"),
		}))
		if n := countRows(t, admin, `SELECT count(*) FROM companies WHERE id = $1 AND name = 'A'`, coA); n != 1 {
			t.Fatal("回滾後公司名稱必須維持原值")
		}
		assertNoAuditRow(t, admin, "company", "update", itoa(coA))
	})

	t.Run("DeleteCompany(自開交易)不落地且不留稽核", func(t *testing.T) {
		assertForcedFailure(t, callCoreErr(t, newFailing(t).companies.DeleteCompany, &v1.DeleteCompanyRequest{
			CompanyId: itoa(coBare),
		}))
		assertCoreSoftDeleted(t, admin, "companies", itoa(coBare), false)
		assertNoAuditRow(t, admin, "company", "delete", itoa(coBare))
	})

	t.Run("CreateDepartment 不落地", func(t *testing.T) {
		assertForcedFailure(t, callCoreErr(t, newFailing(t).departments.CreateDepartment, &v1.CreateDepartmentRequest{
			CompanyId: itoa(coA), Name: "回滾部門",
		}))
		if n := countRows(t, admin, `SELECT count(*) FROM departments WHERE name = '回滾部門'`); n != 0 {
			t.Fatalf("請求交易回滾後不得留下部門列,got %d", n)
		}
	})

	t.Run("UpdateDepartment 不落地", func(t *testing.T) {
		assertForcedFailure(t, callCoreErr(t, newFailing(t).departments.UpdateDepartment, &v1.UpdateDepartmentRequest{
			DepartmentId: itoa(deptA), Name: protoStr("回滾部門改名"),
		}))
		if n := countRows(t, admin, `SELECT count(*) FROM departments WHERE id = $1 AND name = 'A 部門'`, deptA); n != 1 {
			t.Fatal("回滾後部門名稱必須維持原值")
		}
	})

	t.Run("DeleteDepartment(自開交易)不落地且不留稽核", func(t *testing.T) {
		assertForcedFailure(t, callCoreErr(t, newFailing(t).departments.DeleteDepartment, &v1.DeleteDepartmentRequest{
			DepartmentId: itoa(deptEmpty),
		}))
		assertCoreSoftDeleted(t, admin, "departments", itoa(deptEmpty), false)
		assertNoAuditRow(t, admin, "department", "delete", itoa(deptEmpty))
	})

	t.Run("CreateUser(自開交易)不落地且不留稽核", func(t *testing.T) {
		assertForcedFailure(t, callCoreErr(t, newFailing(t).users.CreateUser, &v1.CreateUserRequest{
			CompanyId: itoa(coA), Email: "core-tx-create@example.com", Name: "回滾員工", Role: "staff",
		}))
		if n := countRows(t, admin, `SELECT count(*) FROM users WHERE email = 'core-tx-create@example.com'`); n != 0 {
			t.Fatalf("請求交易回滾後不得留下使用者列,got %d", n)
		}
		if n := countRows(t, admin, `SELECT count(*) FROM audit_logs WHERE resource_type = 'user' AND action = 'create'`); n != 0 {
			t.Fatalf("請求交易回滾後不得留下稽核列,got %d", n)
		}
	})

	t.Run("UpdateUser(自開交易)不落地且不留稽核", func(t *testing.T) {
		assertForcedFailure(t, callCoreErr(t, newFailing(t).users.UpdateUser, &v1.UpdateUserRequest{
			UserId: itoa(target), Name: protoStr("回滾改名"),
		}))
		if n := countRows(t, admin, `SELECT count(*) FROM users WHERE id = $1 AND name = '公司管理員'`, target); n != 1 {
			t.Fatal("回滾後使用者名稱必須維持原值")
		}
		assertNoAuditRow(t, admin, "user", "update", itoa(target))
	})

	t.Run("AssignRole(自開交易)不落地且不留稽核", func(t *testing.T) {
		assertForcedFailure(t, callCoreErr(t, newFailing(t).users.AssignRole, &v1.AssignRoleRequest{
			UserId: itoa(target), Role: "dept_admin",
		}))
		if n := countRows(t, admin,
			`SELECT count(*) FROM users WHERE id = $1 AND role = 'staff' AND token_version = 0`, target); n != 1 {
			t.Fatal("回滾後角色與 token_version 必須維持原值")
		}
		assertNoAuditRow(t, admin, "user", "role_change", itoa(target))
	})

	t.Run("Deactivate(自開交易)不落地且不留稽核", func(t *testing.T) {
		assertForcedFailure(t, callCoreErr(t, newFailing(t).users.Deactivate, &v1.DeactivateRequest{
			UserId: itoa(target),
		}))
		if n := countRows(t, admin,
			`SELECT count(*) FROM users WHERE id = $1 AND status = 'active' AND token_version = 0`, target); n != 1 {
			t.Fatal("回滾後 status 與 token_version 必須維持原值")
		}
		assertNoAuditRow(t, admin, "user", "update", itoa(target))
	})

	t.Run("ForceLogout(自開交易)不落地且不留稽核", func(t *testing.T) {
		assertForcedFailure(t, callCoreErr(t, newFailing(t).users.ForceLogout, &v1.ForceLogoutRequest{
			UserId: itoa(target),
		}))
		if n := countRows(t, admin, `SELECT count(*) FROM users WHERE id = $1 AND token_version = 0`, target); n != 1 {
			t.Fatal("回滾後 token_version 必須維持原值")
		}
		assertNoAuditRow(t, admin, "user", "force_logout", itoa(target))
	})

	t.Run("UpdateRolePermissions(自開交易)不落地", func(t *testing.T) {
		assertForcedFailure(t, callCoreErr(t, newFailing(t).roles.UpdateRolePermissions, &v1.UpdateRolePermissionsRequest{
			RoleId: strconv.Itoa(roleID),
			Permissions: []*v1.Permission{
				{Resource: "customer", Action: "read", SortOrder: 0},
				{Resource: "customer", Action: "update", SortOrder: 1},
			},
		}))
		if n := countRows(t, admin, `SELECT count(*) FROM role_permissions WHERE role_id = $1`, roleID); n != 1 {
			t.Fatalf("回滾後 role_permissions 必須維持原樣(1 筆),got %d", n)
		}
	})
}

// TestIntegrationOpenFGAHookAfterCommit OpenFGA tuple 同步必須排在請求交易 **commit 之後**
// (dbtenant.AfterCommit)。兩條斷言各釘一半:
//
//	① 回滾路徑(handler 之後強制失敗)→ 掛鉤被丟棄:tuple 不得寫入,業務列亦未落地;
//	② 成功路徑 → 掛鉤跑了:新權限的 can_write tuple 存在。
//
// 把同步移回交易內(commit 前)會讓 ① 紅 —— OpenFGA 停在被回滾的 DB 狀態(新增方向 = 多授權,
// fail-open,且要等下次 authz.Provision reconcile 才修正)。
// 兩個子測試各用**自己的 in-memory 引擎**,互不污染;順序刻意「先回滾、後成功」,讓兩者都在
// 「DB 初始為 1 筆權限、請求欲增為 2 筆」的同一 diff 上比較。
func TestIntegrationOpenFGAHookAfterCommit(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin := openRawDB(t, adminDSN)
	defer func() { _ = admin.Close() }()

	coA := insertRLSCompany(t, admin, "A", "CORE-HOOK-A")
	superID := insertRLSUser(t, admin, coA, "core-hook-super@example.com", "super")
	roleID := insertRLSCoreRole(t, admin, "core_hook_role")
	insertRLSCoreRolePermission(t, admin, roleID, "customer", "read")

	client := openAppRoleEntClient(t, adminDSN)
	super := authz.Identity{UserID: itoa(superID), CompanyID: itoa(coA), Role: "super", Roles: []string{"super"}}
	scope := auth.RLSScope{UserID: itoa(superID), CompanyID: itoa(coA), DataScope: auth.DataScopeAll, CompanyActive: true}
	userset := "role:" + strconv.Itoa(roleID) + "#assigned"

	// addUpdate 把角色權限由 1 筆(read)擴為 2 筆(read + update)→ can_write 需要新 tuple。
	addUpdate := func() *v1.UpdateRolePermissionsRequest {
		return &v1.UpdateRolePermissionsRequest{
			RoleId: strconv.Itoa(roleID),
			Permissions: []*v1.Permission{
				{Resource: "customer", Action: "read", SortOrder: 0},
				{Resource: "customer", Action: "update", SortOrder: 1},
			},
		}
	}
	// newEngine 建立一個乾淨的 in-memory 引擎(store 名須符合 OpenFGA 的命名 regex,故不用 t.Name())。
	newEngine := func(t *testing.T, name string) (*authzopenfga.Engine, func(string) bool) {
		t.Helper()
		fgaClient, err := ofga.NewMemory(t.Context(), "t9-hook-"+name)
		if err != nil {
			t.Fatalf("NewMemory: %v", err)
		}
		t.Cleanup(fgaClient.Close)
		e := authzopenfga.New(fgaClient)
		return e, func(relation string) bool {
			ok, cerr := e.Check(t.Context(), userset, relation, "ability:customer")
			if cerr != nil {
				t.Fatalf("engine.Check(%s): %v", relation, cerr)
			}
			return ok
		}
	}

	t.Run("回滾路徑:掛鉤被丟棄 → tuple 未寫入且業務列未落地", func(t *testing.T) {
		engine, has := newEngine(t, "rollback")
		c := newCoreEngineServer(t, client, super, scope, engine, true)
		assertForcedFailure(t, callCoreErr(t, c.roles.UpdateRolePermissions, addUpdate()))
		if n := countRows(t, admin, `SELECT count(*) FROM role_permissions WHERE role_id = $1`, roleID); n != 1 {
			t.Fatalf("請求交易回滾後 role_permissions 必須維持原樣(1 筆),got %d", n)
		}
		if has("can_write") {
			t.Fatal("交易回滾時 post-commit 掛鉤不得執行:OpenFGA 會停在被回滾的 DB 狀態(多授權/fail-open)")
		}
	})

	t.Run("成功路徑:commit 後掛鉤執行 → tuple 已寫入", func(t *testing.T) {
		engine, has := newEngine(t, "success")
		c := newCoreEngineServer(t, client, super, scope, engine, false)
		callCore(t, c.roles.UpdateRolePermissions, addUpdate())
		// syncRolePermissions 只同步「差異」(全量對齊是開機的 authz.Provision 負責),故此處只斷言
		// 新增的那個 tuple 已被寫入。
		if !has("can_write") {
			t.Fatal("commit 成功後 post-commit 掛鉤應寫入 can_write tuple")
		}
		if n := countRows(t, admin, `SELECT count(*) FROM role_permissions WHERE role_id = $1`, roleID); n != 2 {
			t.Fatalf("成功路徑應留下 2 筆權限,got %d", n)
		}
	})
}

// TestIntegrationRLSCoreEnableMigrationDown 00028 的 Up/Down 必須對稱(ENABLE + FORCE ↔
// NO FORCE + DISABLE),且 Down 不得動到任何 policy(五張表的 policy 由 00023/00025 定義:少了
// 它們,回退後的環境與 00027 的狀態不一致,而旗標上看不出來),並可重複套用。
func TestIntegrationRLSCoreEnableMigrationDown(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)

	admin := openRawDB(t, dsn)
	defer func() { _ = admin.Close() }()

	assertCoreRLSFlags(t, admin, true, true)
	assertCorePoliciesPresent(t, admin)

	migrateBusinessDownTo(t, dsn, "27")
	assertCoreRLSFlags(t, admin, false, false)
	assertCorePoliciesPresent(t, admin)

	migrateBusinessUp(t, dsn)
	assertCoreRLSFlags(t, admin, true, true)
	assertCorePoliciesPresent(t, admin)
}

// coreTables 為核心五表(00028 的 ENABLE + FORCE 範圍)。
var coreTables = []string{"companies", "departments", "users", "roles", "role_permissions"}

// corePolicies 為核心五表的 policy(00023 建立、00025 以 NULLIF 重建);00028 只准動旗標。
var corePolicies = []struct{ table, policy string }{
	{"companies", "core_companies_scope"},
	{"departments", "core_departments_scope"},
	{"users", "core_users_scope"},
	{"roles", "core_roles_read"},
	{"role_permissions", "core_role_permissions_read"},
}

// assertCoreRLSFlags 逐表斷言 ENABLE / FORCE 旗標。
func assertCoreRLSFlags(t *testing.T, db *sql.DB, enabled, forced bool) {
	t.Helper()
	for _, table := range coreTables {
		var gotEnabled, gotForced bool
		if err := db.QueryRow(
			`SELECT relrowsecurity, relforcerowsecurity FROM pg_class WHERE relname = $1`, table).
			Scan(&gotEnabled, &gotForced); err != nil {
			t.Fatalf("查 %s 的 RLS 旗標: %v", table, err)
		}
		if gotEnabled != enabled || gotForced != forced {
			t.Fatalf("%s 的旗標應為 ENABLE=%v FORCE=%v,got ENABLE=%v FORCE=%v",
				table, enabled, forced, gotEnabled, gotForced)
		}
	}
}

// assertCorePoliciesPresent 斷言核心五表的 policy 仍存在(00028 的 Up/Down 皆不得 DROP)。
func assertCorePoliciesPresent(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, p := range corePolicies {
		var n int
		if err := db.QueryRow(
			`SELECT count(*) FROM pg_policies WHERE tablename = $1 AND policyname = $2`, p.table, p.policy).
			Scan(&n); err != nil {
			t.Fatalf("查 policy %s/%s: %v", p.table, p.policy, err)
		}
		if n != 1 {
			t.Fatalf("policy %s 應存在於 %s(00028 不得動 policy),got %d", p.policy, p.table, n)
		}
	}
}

// assertCoreSoftDeleted 斷言軟刪除旗標(admin 真值查詢,不受 RLS 影響)。
func assertCoreSoftDeleted(t *testing.T, db *sql.DB, table, id string, wantDeleted bool) {
	t.Helper()
	expr := `SELECT count(*) FROM ` + table + ` WHERE id = $1 AND deleted_at IS NULL`
	if wantDeleted {
		expr = `SELECT count(*) FROM ` + table + ` WHERE id = $1 AND deleted_at IS NOT NULL`
	}
	if n := countRows(t, db, expr, id); n != 1 {
		t.Fatalf("%s(id=%s) 的 deleted_at 狀態不符(期望已刪除=%v)", table, id, wantDeleted)
	}
}

// assertNoAuditRow 斷言某筆業務寫入沒有留下稽核列(回滾的證據)。
func assertNoAuditRow(t *testing.T, db *sql.DB, resourceType, action, resourceID string) {
	t.Helper()
	if n := countRows(t, db,
		`SELECT count(*) FROM audit_logs WHERE resource_type = $1 AND action = $2 AND resource_id = $3`,
		resourceType, action, resourceID); n != 0 {
		t.Fatalf("%s/%s(resource_id=%s) 不得留下稽核列,got %d", resourceType, action, resourceID, n)
	}
}

// assertRoleIDPresent 斷言角色清單含指定 id(共享目錄在任何非空 scope 下都讀得到)。
func assertRoleIDPresent(t *testing.T, resp *v1.ListRolesResponse, wantID string) {
	t.Helper()
	for _, r := range resp.GetRoles() {
		if r.GetId() == wantID {
			return
		}
	}
	t.Fatalf("ListRoles 缺少 id=%s", wantID)
}

// assertCoreCount 以指定交易斷言單一純量(RLS 可見性,與 admin 的真值查詢分開)。
func assertCoreCount(t *testing.T, tx *sql.Tx, query string, want int, arg any, what string) {
	t.Helper()
	var got int
	var err error
	if arg == nil {
		err = tx.QueryRow(query).Scan(&got)
	} else {
		err = tx.QueryRow(query, arg).Scan(&got)
	}
	if err != nil {
		t.Fatalf("%s:%v", what, err)
	}
	if got != want {
		t.Fatalf("%s:期望 %d,got %d", what, want, got)
	}
}

// coreExecScoped 以 app_rw 開**獨立**交易、套用 scope 後執行單一語句,回傳影響列數。
// 每個受測語句各自一個交易:一句失敗會 abort 整個交易,共用交易會讓後續語句只剩 25P02。
func coreExecScoped(t *testing.T, app *sql.DB, stmts []string, query string, args ...any) (int64, error) {
	t.Helper()
	tx, err := app.Begin()
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	execScope(t, tx, stmts)
	res, err := tx.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		t.Fatalf("RowsAffected: %v", err)
	}
	return n, nil
}

// insertRLSCoreRole 以 admin 連線建一列自訂角色(roles 無租戶欄;policy 只要求 scope 非空)。
func insertRLSCoreRole(t *testing.T, db *sql.DB, code string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO roles (code, name, data_scope, is_system, is_active)
		 VALUES ($1, '核心域探針角色', 'company', false, true) RETURNING id`, code).Scan(&id); err != nil {
		t.Fatalf("建角色 %s: %v", code, err)
	}
	return id
}

// insertRLSCoreRolePermission 以 admin 連線建一筆角色權限。
func insertRLSCoreRolePermission(t *testing.T, db *sql.DB, roleID int, resource, action string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO role_permissions (role_id, resource, action, sort_order)
		 VALUES ($1, $2, $3, 0) RETURNING id`, roleID, resource, action).Scan(&id); err != nil {
		t.Fatalf("建角色權限 %s/%s: %v", resource, action, err)
	}
	return id
}

// coreClients 為核心域四支 service 的 RPC client(同一 mux、同一身分與 scope)。
type coreClients struct {
	companies   salesorderv1connect.CompanyServiceClient
	departments salesorderv1connect.DepartmentServiceClient
	users       salesorderv1connect.UserServiceClient
	roles       salesorderv1connect.RoleServiceClient
}

// openAppRoleEntClient 建立 app_rw 的業務 client:必須經 dbtenant.NewClient(RLS 裝飾器才生效)。
//
// **連線池不設上限 1**:未登入／系統範圍路徑的 SystemScopeTx 會在請求交易之外再開一條交易
// (第二條連線);池若只有 1 條,它會與請求交易互鎖成死結。上限 4 是明確的「≥2」,同時讓
// 「漏收斂的查詢」不至於無聲佔滿池。
func openAppRoleEntClient(t *testing.T, adminDSN string) *ent.Client {
	t.Helper()
	pool := openAppRoleDB(t, adminDSN)
	pool.SetMaxOpenConns(4)
	client := dbtenant.NewClient(pool)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// newCoreAppRoleServer 以業務 client(app_rw + dbtenant.NewClient)掛載三支核心 domain 的 handler,
// 掛載方式與生產一致(dbtenant.HandlerOption → 每個 RPC 都在請求交易內執行)。
// withScope=false 模擬「漏掛 RLS scope」(負向對照);生產的 scope 由 authzMiddleware 依身分導出。
func newCoreAppRoleServer(t *testing.T, client *ent.Client, id authz.Identity, scope auth.RLSScope, withScope bool) coreClients {
	t.Helper()
	mux := http.NewServeMux()
	mountCoreHandlers(mux, client, dbtenant.HandlerOption(client))
	return serveCoreClients(t, mux, id, scope, withScope, nil)
}

// mountCoreHandlers 以 opts 掛上核心域四支 handler(與 RegisterCompanyServices／RegisterUserServices／
// RegisterRoleServices 相同的組裝,差別是能把額外 interceptor 疊在 dbtenant.Interceptor 之後)。
func mountCoreHandlers(mux *http.ServeMux, client *ent.Client, opts ...connect.HandlerOption) {
	companyPath, companyHandler := salesorderv1connect.NewCompanyServiceHandler(NewCompanyService(client), opts...)
	mux.Handle(companyPath, companyHandler)
	departmentPath, departmentHandler := salesorderv1connect.NewDepartmentServiceHandler(NewDepartmentService(client), opts...)
	mux.Handle(departmentPath, departmentHandler)
	userPath, userHandler := salesorderv1connect.NewUserServiceHandler(NewUserService(client), opts...)
	mux.Handle(userPath, userHandler)
	rolePath, roleHandler := salesorderv1connect.NewRoleServiceHandler(NewRoleService(client), opts...)
	mux.Handle(rolePath, roleHandler)
}

// newCoreFailingServer 與 newCoreAppRoleServer 相同的組裝,差別在 interceptor 疊法:
// dbtenant.Interceptor 在外、failAfterHandler 在內(同一組 WithInterceptors 的參數順序即外→內),
// 故 handler 的寫入先進入請求交易、再由外層 rollback。
func newCoreFailingServer(t *testing.T, client *ent.Client, id authz.Identity, scope auth.RLSScope) coreClients {
	t.Helper()
	return newCoreEngineServer(t, client, id, scope, nil, true)
}

// newCoreEngineServer 與 newCoreFailingServer 相同的組裝,差別是能把 OpenFGA 引擎放進 ctx
// (供 tuple 同步的 post-commit 掛鉤斷言);failing=true 時在 handler 之後強制失敗(驗證回滾丟棄掛鉤)。
func newCoreEngineServer(t *testing.T, client *ent.Client, id authz.Identity, scope auth.RLSScope, engine *authzopenfga.Engine, failing bool) coreClients {
	t.Helper()
	mux := http.NewServeMux()
	interceptors := []connect.Interceptor{dbtenant.Interceptor(client)}
	if failing {
		interceptors = append(interceptors, failAfterHandler{})
	}
	mountCoreHandlers(mux, client, connect.WithInterceptors(interceptors...))
	return serveCoreClients(t, mux, id, scope, true, engine)
}

// serveCoreClients 以指定身分與 scope 包住 mux 並起 httptest server。
func serveCoreClients(t *testing.T, mux *http.ServeMux, id authz.Identity, scope auth.RLSScope, withScope bool, engine *authzopenfga.Engine) coreClients {
	t.Helper()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = audit.WithMeta(ctx, audit.Meta{IP: "10.0.0.9", UserAgent: "t9-core-probe"})
		if engine != nil {
			ctx = authz.WithEngine(ctx, engine)
		}
		if withScope {
			ctx = auth.WithRLS(ctx, scope)
		}
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return coreClients{
		companies:   salesorderv1connect.NewCompanyServiceClient(http.DefaultClient, ts.URL),
		departments: salesorderv1connect.NewDepartmentServiceClient(http.DefaultClient, ts.URL),
		users:       salesorderv1connect.NewUserServiceClient(http.DefaultClient, ts.URL),
		roles:       salesorderv1connect.NewRoleServiceClient(http.DefaultClient, ts.URL),
	}
}

// callCore 以單一請求呼叫 RPC 並回傳回應訊息(錯誤即測試失敗)。
// 直接吃生成 client 的方法值,讓呼叫點與 RPC 簽章之間不隔一層閉包。
func callCore[M, T any](t *testing.T, call func(context.Context, *connect.Request[M]) (*connect.Response[T], error), msg *M) *T {
	t.Helper()
	resp, err := call(t.Context(), connect.NewRequest(msg))
	if err != nil {
		t.Fatalf("RPC 失敗: %v", err)
	}
	return resp.Msg
}

// callCoreErr 以單一請求呼叫 RPC,錯誤交由呼叫端斷言。
func callCoreErr[M, T any](t *testing.T, call func(context.Context, *connect.Request[M]) (*connect.Response[T], error), msg *M) error {
	t.Helper()
	_, err := call(t.Context(), connect.NewRequest(msg))
	return err
}
