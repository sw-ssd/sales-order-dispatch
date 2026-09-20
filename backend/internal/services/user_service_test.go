package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/department"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
)

// newUserTestServer 建立自建 db(empty) 的 UserService server 並注入身分。
func newUserTestServer(t *testing.T, id authz.Identity) (salesorderv1connect.UserServiceClient, *ent.Client) {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_fk=1"
	db := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = db.Close() })
	return newUserTestServerWithDB(t, id, db), db
}

// newUserTestServerWithDB 以已建立的 db 建立 UserService server 並注入身分。
func newUserTestServerWithDB(t *testing.T, id authz.Identity, db *ent.Client) salesorderv1connect.UserServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	RegisterUserServices(mux, db, entitlements.Unlimited())
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithDB(ctx, db)
		// 模擬 middleware 注入稽核來源資訊(I9)。
		ctx = audit.WithMeta(ctx, audit.Meta{IP: "127.0.0.1", UserAgent: "test-agent"})
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return salesorderv1connect.NewUserServiceClient(http.DefaultClient, ts.URL)
}

// seedUserCompany 建立公司與兩部門,回傳 (companyID, deptA, deptB)。
func seedUserCompany(t *testing.T, db *ent.Client) (int, int, int) {
	t.Helper()
	ctx := context.Background()
	co, err := db.Company.Create().SetName("公司A").SetIdentifier("co-a").Save(ctx)
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	da, err := db.Department.Create().SetCompanyID(co.ID).SetName("部門甲").Save(ctx)
	if err != nil {
		t.Fatalf("dept: %v", err)
	}
	dbb, err := db.Department.Create().SetCompanyID(co.ID).SetName("部門乙").Save(ctx)
	if err != nil {
		t.Fatalf("dept: %v", err)
	}
	return co.ID, da.ID, dbb.ID
}

func uItoa(i int) string { return strconv.Itoa(i) }

// --- 審查修復:授權下限與授予上限(C1/C2/I1/I2/I3/I5/I6)---

// TestListUsersNonManagerDenied:非管理角色(staff)不得列舉使用者(C1)。
func TestListUsersNonManagerDenied(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	id := authz.Identity{UserID: "9", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "staff", Roles: []string{"staff", "customer"}}
	client := newUserTestServerWithDB(t, id, db)
	_, err := client.ListUsers(ctx, connect.NewRequest(&v1.ListUsersRequest{}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("staff 列舉使用者應回 permission_denied,得到 %v", err)
	}
}

// TestListUsersSuperDepartmentFilter:super 帶 department_id 時確實套用篩選(I3)。
func TestListUsersSuperDepartmentFilter(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, deptB := seedUserCompany(t, db)
	if _, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("fa@t.com").SetName("甲").SetRole("staff").SetPasswordHash("x").Save(ctx); err != nil {
		t.Fatalf("user: %v", err)
	}
	if _, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptB).SetEmail("fb@t.com").SetName("乙").SetRole("staff").SetPasswordHash("x").Save(ctx); err != nil {
		t.Fatalf("user: %v", err)
	}
	id := authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}}
	client := newUserTestServerWithDB(t, id, db)
	resp, err := client.ListUsers(ctx, connect.NewRequest(&v1.ListUsersRequest{DepartmentId: uItoa(deptA)}))
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(resp.Msg.GetUsers()) != 1 || resp.Msg.GetUsers()[0].GetDepartmentId() != uItoa(deptA) {
		t.Fatalf("期望僅部門甲 1 人,得到 %d 人", len(resp.Msg.GetUsers()))
	}
}

// TestCreateUserCannotGrantSuperRole:company_admin 不得建立 super 帳號(C2)。
func TestCreateUserCannotGrantSuperRole(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)
	_, err := client.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
		Name: "壞人", Email: "bad@t.com", CompanyId: uItoa(coID), Role: "super",
	}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("company_admin 建立 super 應回 permission_denied,得到 %v", err)
	}
	// 確認未落庫。
	n, _ := db.User.Query().Count(ctx)
	if n != 0 {
		t.Errorf("拒絕後不應建立帳號,得到 %d 筆", n)
	}
}

// TestCreateUserSoftDeletedCompany:P2-A 後續——super 以請求指定已軟刪除的公司 → not_found、不落庫。
// CreateUser 是唯一以「請求」指定 company_id 的掛載路徑;不擋就會把活帳號掛進已刪租戶
// (該帳號之後仍能通過登入與身分解析,因為那些路徑讀的是「使用者的公司」)。
func TestCreateUserSoftDeletedCompany(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, _, _ := seedUserCompany(t, db)
	db.Company.UpdateOneID(coID).SetDeletedAt(time.Now().UTC()).SaveX(ctx)

	id := authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}}
	client := newUserTestServerWithDB(t, id, db)
	_, err := client.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
		Name: "幽靈", Email: "ghost@t.com", CompanyId: uItoa(coID), Role: "staff",
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("已刪除公司下建帳號應回 not_found,得到 %v", err)
	}
	if n, _ := db.User.Query().Count(ctx); n != 0 {
		t.Fatalf("已刪除公司下不得建帳號,得到 %d 筆", n)
	}
}

// TestAssignRoleCannotGrantSuper:company_admin 不得把任何人(含自己)升為 super(C2)。
func TestAssignRoleCannotGrantSuper(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	self, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("ca@t.com").SetName("ca").SetRole("company_admin").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	id := authz.Identity{UserID: uItoa(self.ID), CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)
	_, err = client.AssignRole(ctx, connect.NewRequest(&v1.AssignRoleRequest{UserId: uItoa(self.ID), Role: "super"}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("自我升為 super 應回 permission_denied,得到 %v", err)
	}
	fresh, _ := db.User.Get(ctx, self.ID)
	if fresh.Role != "company_admin" {
		t.Errorf("角色不應變更,得到 %s", fresh.Role)
	}
}

// TestDeptAdminCannotGrantDeptAdmin:dept_admin 僅能授予 staff(規格 2.3.2)。
func TestDeptAdminCannotGrantDeptAdmin(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	target, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("s2@t.com").SetName("s").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	id := authz.Identity{UserID: "5", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)
	_, err = client.AssignRole(ctx, connect.NewRequest(&v1.AssignRoleRequest{UserId: uItoa(target.ID), Role: "dept_admin"}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("dept_admin 授予 dept_admin 應回 permission_denied,得到 %v", err)
	}
}

// TestAssignRoleDoesNotReactivateInactive:AssignRole 不得將已停用帳號默默復活(I5)。
func TestAssignRoleDoesNotReactivateInactive(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	target, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("in@t.com").SetName("i").SetRole("staff").SetStatus("inactive").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)
	resp, err := client.AssignRole(ctx, connect.NewRequest(&v1.AssignRoleRequest{UserId: uItoa(target.ID), Role: "staff"}))
	if err != nil {
		t.Fatalf("AssignRole: %v", err)
	}
	if resp.Msg.GetUser().GetStatus() != "inactive" {
		t.Errorf("已停用帳號不應被復活,得到 %s", resp.Msg.GetUser().GetStatus())
	}
}

// TestAssignRoleAuditFailureRollsBack:稽核寫入失敗 → 業務異動回滾(D18 不變式)。
func TestAssignRoleAuditFailureRollsBack(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	target, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("rb@t.com").SetName("rb").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	// 注入失敗:操作者 UserID 無法解析 → audit.Record 因缺操作者脈絡而失敗 → 交易須回滾。
	id := authz.Identity{UserID: "not-a-number", CompanyID: uItoa(coID), Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)
	_, err = client.AssignRole(ctx, connect.NewRequest(&v1.AssignRoleRequest{UserId: uItoa(target.ID), Role: "dept_admin"}))
	if connect.CodeOf(err) != connect.CodeInternal {
		t.Fatalf("稽核失敗應回 internal,得到 %v", err)
	}
	fresh, gerr := db.User.Get(ctx, target.ID)
	if gerr != nil {
		t.Fatalf("get: %v", gerr)
	}
	if fresh.Role != "staff" || fresh.TokenVersion != 0 {
		t.Errorf("交易應回滾:role=%s tv=%d", fresh.Role, fresh.TokenVersion)
	}
	n, _ := db.AuditLog.Query().Count(ctx)
	if n != 0 {
		t.Errorf("不應有稽核殘留,得到 %d", n)
	}
}

// TestCreateUserRejectsForeignDepartment:department_id 須屬於目標公司(I6)。
func TestCreateUserRejectsForeignDepartment(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, _, _ := seedUserCompany(t, db)
	// 另一家公司與其部門。
	coB, err := db.Company.Create().SetName("公司B").SetIdentifier("co-b").Save(ctx)
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	deptB, err := db.Department.Create().SetCompanyID(coB.ID).SetName("B部門").Save(ctx)
	if err != nil {
		t.Fatalf("dept: %v", err)
	}
	id := authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}}
	client := newUserTestServerWithDB(t, id, db)
	_, err = client.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
		Name: "x", Email: "x@t.com", CompanyId: uItoa(coID), DepartmentId: uItoa(deptB.ID), Role: "staff",
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("跨公司部門應回 invalid_argument,得到 %v", err)
	}
}

// TestAssignSuperCustomRole:super 可授予「既有自訂角色」(殘留 #1)。
func TestAssignSuperCustomRole(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	// 建立自訂角色(非 is_system, active)。
	if _, err := db.Role.Create().SetCode("ops_manager").SetName("營運主管").SetDataScope("department").SetIsSystem(false).SetIsActive(true).Save(ctx); err != nil {
		t.Fatalf("role: %v", err)
	}
	target, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("cu@t.com").SetName("c").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	id := authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}}
	client := newUserTestServerWithDB(t, id, db)
	resp, err := client.AssignRole(ctx, connect.NewRequest(&v1.AssignRoleRequest{UserId: uItoa(target.ID), Role: "ops_manager"}))
	if err != nil {
		t.Fatalf("super 授予自訂角色應成功:%v", err)
	}
	if resp.Msg.GetUser().GetRole() != "ops_manager" {
		t.Errorf("期望 role=ops_manager,得到 %s", resp.Msg.GetUser().GetRole())
	}
}

// TestCompanyAdminCannotGrantCustomRole:company_admin 不得授予自訂角色(殘留 #1 之安全下限)。
func TestCompanyAdminCannotGrantCustomRole(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	if _, err := db.Role.Create().SetCode("ops_manager").SetName("營運主管").SetDataScope("department").SetIsSystem(false).SetIsActive(true).Save(ctx); err != nil {
		t.Fatalf("role: %v", err)
	}
	target, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("c2@t.com").SetName("c2").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)
	_, err = client.AssignRole(ctx, connect.NewRequest(&v1.AssignRoleRequest{UserId: uItoa(target.ID), Role: "ops_manager"}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("company_admin 授予自訂角色應回 permission_denied,得到 %v", err)
	}
}

// TestCreateCustomRoleNotEscalateFromCompanyAdmin:company_admin 不得建立自訂角色帳號(殘留 #1)。
func TestCreateCustomRoleNotEscalateFromCompanyAdmin(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, _, _ := seedUserCompany(t, db)
	if _, err := db.Role.Create().SetCode("ops_manager").SetName("營運主管").SetDataScope("department").SetIsSystem(false).SetIsActive(true).Save(ctx); err != nil {
		t.Fatalf("role: %v", err)
	}
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)
	_, err := client.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
		Name: "x", Email: "c@t.com", CompanyId: uItoa(coID), Role: "ops_manager",
	}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("company_admin 建立自訂角色帳號應回 permission_denied,得到 %v", err)
	}
}

// TestCreateAndUpdateUserWriteAudit:CreateUser / UpdateUser 亦須寫稽核(I4;2.3.1)。
func TestCreateAndUpdateUserWriteAudit(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	id := authz.Identity{UserID: "1", Role: "super", Roles: []string{"super"}}
	client := newUserTestServerWithDB(t, id, db)

	resp, err := client.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
		Name: "新人", Email: "new@t.com", CompanyId: uItoa(coID), DepartmentId: uItoa(deptA), Role: "staff",
	}))
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	newID := resp.Msg.GetUser().GetId()

	name := "改名"
	if _, err := client.UpdateUser(ctx, connect.NewRequest(&v1.UpdateUserRequest{UserId: newID, Name: &name})); err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}

	audits, err := db.AuditLog.Query().All(ctx)
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	var created, updated bool
	for _, a := range audits {
		if a.Action == "create" && a.ResourceID == newID {
			created = true
		}
		if a.Action == "update" && a.ResourceID == newID {
			updated = true
			if a.BeforeSnapshot["name"] != "新人" || a.AfterSnapshot["name"] != "改名" {
				t.Errorf("update 快照不正確: before=%v after=%v", a.BeforeSnapshot, a.AfterSnapshot)
			}
		}
	}
	if !created || !updated {
		t.Errorf("期望 create 與 update 稽核各一筆 (create=%v update=%v)", created, updated)
	}
}

// TestAuditRecordsSourceMeta:稽核寫入應帶 middleware 注入的 IP / User-Agent(I9)。
func TestAuditRecordsSourceMeta(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, _, _ := seedUserCompany(t, db)
	target, err := db.User.Create().SetCompanyID(coID).SetEmail("meta@t.com").SetName("m").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)
	if _, err := client.Deactivate(ctx, connect.NewRequest(&v1.DeactivateRequest{UserId: uItoa(target.ID)})); err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	a, err := db.AuditLog.Query().Only(ctx)
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	if a.IPAddress != "127.0.0.1" || a.UserAgent != "test-agent" {
		t.Errorf("稽核缺來源資訊: ip=%q ua=%q", a.IPAddress, a.UserAgent)
	}
	// after_snapshot 應為變更後狀態(I2)。
	if a.AfterSnapshot["status"] != "inactive" {
		t.Errorf("after_snapshot 應為 inactive,得到 %v", a.AfterSnapshot)
	}
	if a.BeforeSnapshot["status"] != "active" {
		t.Errorf("before_snapshot 應為 active,得到 %v", a.BeforeSnapshot)
	}
}

// TestScopeFailClosedOnMissingCompany:身分缺 company_id 時不得放行(I1)。
func TestScopeFailClosedOnMissingCompany(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, _, _ := seedUserCompany(t, db)
	target, err := db.User.Create().SetCompanyID(coID).SetEmail("o@t.com").SetName("o").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	// company_admin 但 CompanyID 為空(異常身分)→ 應 fail-closed。
	id := authz.Identity{UserID: "2", Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)
	_, err = client.GetUser(ctx, connect.NewRequest(&v1.GetUserRequest{UserId: uItoa(target.ID)}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("缺公司脈絡應回 permission_denied,得到 %v", err)
	}
}

func TestAssignRoleBumpsTokenVersionAndAudit(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	// 目標使用者(guest, pending)。
	target, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("g@t.com").SetName("guest").SetRole("guest").SetStatus("pending").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	// 操作者 company_admin(同公司)。
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)

	resp, err := client.AssignRole(ctx, connect.NewRequest(&v1.AssignRoleRequest{
		UserId: uItoa(target.ID), Role: "staff", DepartmentId: uItoa(deptA),
	}))
	if err != nil {
		t.Fatalf("AssignRole: %v", err)
	}
	if resp.Msg.GetUser().GetStatus() != "active" {
		t.Errorf("期望 active,得到 %s", resp.Msg.GetUser().GetStatus())
	}
	if resp.Msg.GetUser().GetRole() != "staff" {
		t.Errorf("期望 role staff,得到 %s", resp.Msg.GetUser().GetRole())
	}

	// token_version 遞增(由 0 → 1)。
	fresh, err := db.User.Get(ctx, target.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if fresh.TokenVersion != 1 {
		t.Errorf("期望 token_version=1,得到 %d", fresh.TokenVersion)
	}

	// 稽核存在(action=role_change)。
	audits, err := db.AuditLog.Query().All(ctx)
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	found := false
	for _, a := range audits {
		if a.Action == "role_change" && a.ResourceID == uItoa(target.ID) {
			found = true
		}
	}
	if !found {
		t.Errorf("期望存在 role_change 稽核紀錄,得到 %d 筆", len(audits))
	}
}

func TestAssignRoleGuestDeniedNoManager(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, _ := seedUserCompany(t, db)
	target, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("g2@t.com").SetName("g").SetRole("guest").SetStatus("pending").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	// staff 非管理者 → permission_denied。
	id := authz.Identity{UserID: "3", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "staff", Roles: []string{"staff", "customer"}}
	client := newUserTestServerWithDB(t, id, db)
	_, err = client.AssignRole(ctx, connect.NewRequest(&v1.AssignRoleRequest{UserId: uItoa(target.ID), Role: "staff"}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("期望 permission_denied,得到 %v", err)
	}
}

func TestDeactivateBumpsTokenVersionAndAudit(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, _, _ := seedUserCompany(t, db)
	target, err := db.User.Create().SetCompanyID(coID).SetEmail("d@t.com").SetName("d").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)

	if _, err := client.Deactivate(ctx, connect.NewRequest(&v1.DeactivateRequest{UserId: uItoa(target.ID)})); err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	fresh, err := db.User.Get(ctx, target.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if fresh.Status != "inactive" {
		t.Errorf("期望 inactive,得到 %s", fresh.Status)
	}
	if fresh.TokenVersion != 1 {
		t.Errorf("期望 token_version=1,得到 %d", fresh.TokenVersion)
	}
	// 稽核存在。
	n, err := db.AuditLog.Query().Count(ctx)
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	if n < 1 {
		t.Errorf("期望稽核紀錄,得到 %d", n)
	}
}

func TestForceLogoutBumpsAndBlocksSelf(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, _, _ := seedUserCompany(t, db)
	target, err := db.User.Create().SetCompanyID(coID).SetEmail("f@t.com").SetName("f").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)

	if _, err := client.ForceLogout(ctx, connect.NewRequest(&v1.ForceLogoutRequest{UserId: uItoa(target.ID)})); err != nil {
		t.Fatalf("ForceLogout: %v", err)
	}
	fresh, _ := db.User.Get(ctx, target.ID)
	if fresh.TokenVersion != 1 {
		t.Errorf("期望 token_version=1,得到 %d", fresh.TokenVersion)
	}

	// 對自己 → invalid_argument。
	// 操作者 id=2,目標 id=2(無此列但身份成立)。
	_, err = client.ForceLogout(ctx, connect.NewRequest(&v1.ForceLogoutRequest{UserId: "2"}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("對自己期望 invalid_argument,得到 %v", err)
	}
}

func TestDeactivateOutOfScopeDenied(t *testing.T) {
	ctx := context.Background()
	_, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, deptB := seedUserCompany(t, db)
	// 部門乙的 staff。
	other, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptB).SetEmail("ob@t.com").SetName("ob").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	// dept_admin(部門甲)想停用部門乙帳號。
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)
	_, err = client.Deactivate(ctx, connect.NewRequest(&v1.DeactivateRequest{UserId: uItoa(other.ID)}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("期望 permission_denied,得到 %v", err)
	}
}

func TestListUsersScope(t *testing.T) {
	ctx := context.Background()

	t.Run("未登入回 Unauthenticated", func(t *testing.T) {
		client, _ := newUserTestServer(t, authz.Identity{})
		_, err := client.ListUsers(ctx, connect.NewRequest(&v1.ListUsersRequest{}))
		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Fatalf("期望 unauthenticated,得到 %v", err)
		}
	})

	t.Run("dept_admin 僅見自己部門", func(t *testing.T) {
		_, db := newUserTestServer(t, authz.Identity{})
		coID, deptA, deptB := seedUserCompany(t, db)
		// 部門甲一個 staff、部門乙一個 staff。
		if _, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("a1@t.com").SetName("甲一").SetRole("staff").SetPasswordHash("x").Save(ctx); err != nil {
			t.Fatalf("user: %v", err)
		}
		if _, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptB).SetEmail("b1@t.com").SetName("乙一").SetRole("staff").SetPasswordHash("x").Save(ctx); err != nil {
			t.Fatalf("user: %v", err)
		}
		// 以 dept_admin 身分重建 client(注入身分)。
		id := authz.Identity{UserID: "1", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
		client := newUserTestServerWithDB(t, id, db)

		resp, err := client.ListUsers(ctx, connect.NewRequest(&v1.ListUsersRequest{}))
		if err != nil {
			t.Fatalf("ListUsers: %v", err)
		}
		if len(resp.Msg.GetUsers()) != 1 {
			t.Fatalf("期望 dept_admin 見 1 人,得到 %d", len(resp.Msg.GetUsers()))
		}
		if got := resp.Msg.GetUsers()[0].GetDepartmentId(); got != uItoa(deptA) {
			t.Fatalf("dept_admin 看到跨部門使用者 dept=%s", got)
		}
	})
}

func TestUpdateUserScope(t *testing.T) {
	ctx := context.Background()
	client, db := newUserTestServer(t, authz.Identity{})
	coID, deptA, deptB := seedUserCompany(t, db)
	staff, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).SetEmail("s@t.com").SetName("staff").SetRole("staff").SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("user: %v", err)
	}

	t.Run("dept_admin 更新自己部門 staff 成功", func(t *testing.T) {
		id := authz.Identity{UserID: "1", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
		c := newUserTestServerWithDB(t, id, db)
		name := "改名"
		if _, err := c.UpdateUser(ctx, connect.NewRequest(&v1.UpdateUserRequest{UserId: uItoa(staff.ID), Name: &name})); err != nil {
			t.Fatalf("UpdateUser: %v", err)
		}
	})

	t.Run("dept_admin 更新他部門使用者被拒", func(t *testing.T) {
		other, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptB).SetEmail("o@t.com").SetName("o").SetRole("staff").SetPasswordHash("x").Save(ctx)
		if err != nil {
			t.Fatalf("user: %v", err)
		}
		id := authz.Identity{UserID: "1", CompanyID: uItoa(coID), DepartmentID: uItoa(deptA), Role: "dept_admin", Roles: []string{"dept_admin", "staff"}}
		c := newUserTestServerWithDB(t, id, db)
		name := "x"
		_, err = c.UpdateUser(ctx, connect.NewRequest(&v1.UpdateUserRequest{UserId: uItoa(other.ID), Name: &name}))
		if connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("期望 permission_denied,得到 %v", err)
		}
	})

	t.Run("未登入更新被拒", func(t *testing.T) {
		name := "x"
		_, err := client.UpdateUser(ctx, connect.NewRequest(&v1.UpdateUserRequest{UserId: uItoa(staff.ID), Name: &name}))
		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Fatalf("期望 unauthenticated,得到 %v", err)
		}
	})
}

// TestListUsersPagination 複審 Minor 4:分頁 meta(Total/PageSize)與跨頁切分正確(走 pageList)。
func TestListUsersPagination(t *testing.T) {
	ctx := context.Background()
	db := enttest.Open(t, "sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	coID, deptA, _ := seedUserCompany(t, db)
	for i, n := range []string{"甲", "乙", "丙"} {
		if _, err := db.User.Create().SetCompanyID(coID).SetDepartmentID(deptA).
			SetEmail("u" + uItoa(i+1) + "@t.com").SetName(n).SetRole("staff").SetPasswordHash("x").Save(ctx); err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}
	id := authz.Identity{UserID: "2", CompanyID: uItoa(coID), Role: "company_admin", Roles: []string{"company_admin", "dept_admin", "staff"}}
	client := newUserTestServerWithDB(t, id, db)

	p1, err := client.ListUsers(ctx, connect.NewRequest(&v1.ListUsersRequest{Page: 1, PageSize: 2}))
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if got := p1.Msg.GetPagination(); got.GetTotal() != 3 || got.GetPageSize() != 2 || len(p1.Msg.GetUsers()) != 2 {
		t.Fatalf("第1頁分頁應 total=3 size=2 users=2,got total=%d size=%d users=%d", got.GetTotal(), got.GetPageSize(), len(p1.Msg.GetUsers()))
	}
	p2, err := client.ListUsers(ctx, connect.NewRequest(&v1.ListUsersRequest{Page: 2, PageSize: 2}))
	if err != nil {
		t.Fatalf("ListUsers p2: %v", err)
	}
	if len(p2.Msg.GetUsers()) != 1 {
		t.Fatalf("第2頁應剩 1 筆,got %d", len(p2.Msg.GetUsers()))
	}
}

// TestDepartmentRowLockDialectGuard D1:兩側的部門列鎖都只允許在 PostgreSQL 生效。
// SQLite 沒有 FOR SHARE/FOR UPDATE,ent 於該 dialect 會讓整條查詢報錯
// (`sql: SELECT .. FOR UPDATE/SHARE not supported in SQLite`,dialect/sql/builder.go 的
// Selector.For)—— sqlite 路徑(sqlite3 預設回歸套件)必須完全不寫入該子句;PG 則必須寫入,
// 否則掛載(CreateUser/UpdateUser/AssignRole)與部門刪除不再互斥(D1 的缺陷會回來)。
func TestDepartmentRowLockDialectGuard(t *testing.T) {
	locks := []struct {
		name string
		lock func(*sql.Selector)
	}{
		{"掛載端 FOR SHARE", lockDepartmentForShare},
		{"刪除端 FOR UPDATE", lockDepartmentForDelete},
	}
	for _, l := range locks {
		for _, tc := range []struct {
			dialect  string
			wantLock bool
		}{
			{dialect: dialect.SQLite, wantLock: false},
			{dialect: dialect.Postgres, wantLock: true},
		} {
			t.Run(l.name+"/"+tc.dialect, func(t *testing.T) {
				sel := sql.Dialect(tc.dialect).Select().From(sql.Table(department.Table))
				l.lock(sel)
				if err := sel.Err(); err != nil {
					t.Fatalf("selector 於 %s 不應報錯: %v", tc.dialect, err)
				}
				query, _ := sel.Query()
				if got := strings.Contains(query, "FOR "); got != tc.wantLock {
					t.Fatalf("%s 的查詢 %q 含列鎖子句 = %v,want %v", tc.dialect, query, got, tc.wantLock)
				}
			})
		}
	}
}
