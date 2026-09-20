package services

// 部門軟刪除(00020)回歸:硬刪除被 audit_logs.department_id 的 FK 永久卡住。
//
// 缺陷現場(複審以真 PG 實測確認):部門 D 的成員 U 被做過任一次有稽核的操作
// (recordUserAudit 以「目標使用者的部門」寫 department_id)→ 之後把 U 調離 D →
// DeleteDepartment(D) 的前置檢查(部門仍有使用者)通過 → 硬刪除被稽核列的 FK 擋下,
// 而 toConnectError 把原因講成與稽核無關的通用約束訊息(「識別碼已被使用」)。
//
// 本檔守住:軟刪除列保留(所有 FK 有主可依)、刪除寫 action=delete 稽核(department_id 指向
// 被刪部門)、所有部門查詢排除已刪除列、前置檢查(仍有成員)不變、已軟刪除的部門不得再被
// 掛載(CreateUser/UpdateUser/AssignRole),且軟刪除的部門不得再阻擋公司刪除
// (軟刪除拿走了硬刪除的隱性保護)。真 PG 的 FK 行為由 department_soft_delete_integration_test.go 守住。
//
// M1(複審):刪除的前置條件與寫入已收斂成**單一敘述式條件更新**(見 company_service.DeleteDepartment),
// 本檔以「仍有成員 → failed_precondition / 已刪除或不存在 → not_found」的兩條分支守住該收斂的語意
// (`TestDeleteDepartmentFailureBranches`)。

import (
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/auditlog"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// TestDeleteDepartmentSoftDelete 00020:刪除部門為軟刪除 —— 列保留(稽核 FK 永遠有主可依、
// 舊行為在有稽核列時必違反 FK)、同一交易寫 action=delete 稽核、查詢不再可見、前置檢查不變。
func TestDeleteDepartmentSoftDelete(t *testing.T) {
	ctx := t.Context()
	super := authz.Identity{UserID: "1", CompanyID: "1", Role: "super", Roles: []string{"super"}}
	_, dc, db := newTestServerWithIdentity(t, super)
	uc := newUserTestServerWithDB(t, super, db)

	co := db.Company.Create().SetName("公司A").SetIdentifier("DEPT-SOFT").SaveX(ctx)
	created, err := dc.CreateDepartment(ctx, connect.NewRequest(&v1.CreateDepartmentRequest{
		CompanyId: uItoa(co.ID), Name: "門市一",
	}))
	if err != nil {
		t.Fatalf("前置 CreateDepartment: %v", err)
	}
	deptID := created.Msg.GetDepartment().GetId()
	did, err := parseID(deptID)
	if err != nil {
		t.Fatalf("解析部門 id %q: %v", deptID, err)
	}

	// 先製造缺陷現場:部門成員被稽核(寫下 department_id = 該部門的稽核列)後被調離部門。
	member, err := uc.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
		Name: "成員", Email: "dept-member@example.com", CompanyId: uItoa(co.ID), Role: "staff", DepartmentId: deptID,
	}))
	if err != nil {
		t.Fatalf("前置 CreateUser: %v", err)
	}
	if n := db.AuditLog.Query().Where(auditlog.DepartmentIDEQ(did)).CountX(ctx); n == 0 {
		t.Fatal("前置條件失敗:成員的稽核列應以該部門為 department_id(硬刪除的 FK 陷阱來源)")
	}
	if _, err := uc.UpdateUser(ctx, connect.NewRequest(&v1.UpdateUserRequest{
		UserId: member.Msg.GetUser().GetId(), DepartmentId: strPtr(""),
	})); err != nil {
		t.Fatalf("前置 UpdateUser(把成員調離部門,使刪除前置檢查通過): %v", err)
	}

	// ① 刪除必須成功(軟刪除):成員已調離、但稽核列仍指向該部門 —— 舊行為在此必被 FK 擋下。
	if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{DepartmentId: deptID})); err != nil {
		t.Fatalf("刪除有稽核列的部門必須成功(舊行為會被 audit_logs.department_id 的 FK 擋下): %v", err)
	}
	row, err := db.Department.Get(ctx, did)
	if err != nil {
		t.Fatalf("部門刪除必須只標記 deleted_at(軟刪除),列不得消失: %v", err)
	}
	if row.DeletedAt == nil {
		t.Fatal("部門刪除必須標記 deleted_at")
	}

	// ② 刪除寫稽核:action=delete、resource_type=department、department_id/resource_id 指向被刪部門。
	deletes := db.AuditLog.Query().
		Where(auditlog.ActionEQ(auditlog.ActionDelete), auditlog.ResourceTypeEQ("department")).
		AllX(ctx)
	if len(deletes) != 1 {
		t.Fatalf("刪除部門應寫 1 筆 action=delete/resource_type=department 稽核,got %d", len(deletes))
	}
	if deletes[0].DepartmentID != did {
		t.Fatalf("稽核列 department_id 應為被刪部門 %d,got %d", did, deletes[0].DepartmentID)
	}
	if deletes[0].CompanyID != co.ID {
		t.Fatalf("稽核列 company_id 應為 %d,got %d", co.ID, deletes[0].CompanyID)
	}
	if deletes[0].ResourceID != deptID {
		t.Fatalf("稽核列 resource_id 應為 %q,got %q", deptID, deletes[0].ResourceID)
	}
	if name, _ := deletes[0].BeforeSnapshot["name"].(string); name != "門市一" {
		t.Fatalf("刪除稽核應留 before 快照(name=門市一),got %v", deletes[0].BeforeSnapshot)
	}

	// ③ 查詢不可見:Get、List(全域與帶 company_id 篩選)、Update 一律排除已刪除列。
	if _, err := dc.GetDepartment(ctx, connect.NewRequest(&v1.GetDepartmentRequest{DepartmentId: deptID})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("已刪除部門 GetDepartment 應 NotFound,got %v", err)
	}
	for _, tc := range []struct {
		name string
		req  *v1.ListDepartmentsRequest
	}{
		{"全域", &v1.ListDepartmentsRequest{}},
		{"帶 company_id 篩選", &v1.ListDepartmentsRequest{CompanyId: uItoa(co.ID)}},
	} {
		list, err := dc.ListDepartments(ctx, connect.NewRequest(tc.req))
		if err != nil {
			t.Fatalf("ListDepartments(%s): %v", tc.name, err)
		}
		if list.Msg.GetPagination().GetTotal() != 0 || len(list.Msg.GetDepartments()) != 0 {
			t.Fatalf("ListDepartments(%s)不得出現已刪除部門,got total=%d items=%d",
				tc.name, list.Msg.GetPagination().GetTotal(), len(list.Msg.GetDepartments()))
		}
	}
	if _, err := dc.UpdateDepartment(ctx, connect.NewRequest(&v1.UpdateDepartmentRequest{
		DepartmentId: deptID, Name: protoStr("門市一改"),
	})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("更新已刪除部門應 NotFound,got %v", err)
	}
	// 冪等語意:已刪除的部門再刪一次 → NotFound。
	if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{DepartmentId: deptID})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("重複刪除應 NotFound,got %v", err)
	}

	// ④ 前置檢查不變:仍有成員的部門不得被刪,也不得被標記 deleted_at。
	keep := db.Department.Create().SetName("門市二").SetCompanyID(co.ID).SaveX(ctx)
	db.User.Create().SetEmail("keep@example.com").SetName("留任").SetStatus("active").
		SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SetDepartmentID(keep.ID).SaveX(ctx)
	if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{DepartmentId: uItoa(keep.ID)})); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("仍有使用者的部門應 FailedPrecondition,got %v", err)
	}
	if got := db.Department.GetX(ctx, keep.ID); got.DeletedAt != nil {
		t.Fatal("前置檢查失敗時不得標記 deleted_at")
	}
}

// TestDeleteDepartmentFailureBranches 00020 + M1:DeleteDepartment 的「該列未被更新」只由**單一**
// 條件式更新判定(條件 = 存在 且未軟刪除 且無使用者),失敗的兩種原因必須可區分 —— 否則重複刪除會
// 回「部門仍有使用者」(真正原因是不存在),而真有成員時會回 404(真正原因是現況不允許):
//
//	仍有成員    → failed_precondition(部門還在,現況不允許刪)
//	已刪除/不存在 → not_found(重複刪除的正確答案)
func TestDeleteDepartmentFailureBranches(t *testing.T) {
	ctx := t.Context()
	super := authz.Identity{UserID: "1", CompanyID: "1", Role: "super", Roles: []string{"super"}}
	_, dc, db := newTestServerWithIdentity(t, super)

	co := db.Company.Create().SetName("公司A").SetIdentifier("DEPT-BRANCH").SaveX(ctx)

	// ① 仍有成員 → failed_precondition,且不得標記 deleted_at。
	occupied := db.Department.Create().SetName("有成員").SetCompanyID(co.ID).SaveX(ctx)
	db.User.Create().SetEmail("occupied@example.com").SetName("佔用").SetStatus("active").
		SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SetDepartmentID(occupied.ID).SaveX(ctx)
	if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{
		DepartmentId: uItoa(occupied.ID),
	})); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("仍有成員的部門應 failed_precondition(不是 not_found),got %v", err)
	}
	if got := db.Department.GetX(ctx, occupied.ID); got.DeletedAt != nil {
		t.Fatal("仍有成員時不得標記 deleted_at")
	}

	// ② 已軟刪除(歷史列)與 ③ 從未存在 → not_found,不得被講成「仍有使用者」。
	gone := db.Department.Create().SetName("已刪").SetCompanyID(co.ID).SaveX(ctx)
	if _, err := db.Department.UpdateOneID(gone.ID).SetDeletedAt(time.Now().UTC()).Save(ctx); err != nil {
		t.Fatalf("標記軟刪除: %v", err)
	}
	for _, tc := range []struct {
		name string
		id   int
	}{
		{"已軟刪除", gone.ID},
		{"不存在", gone.ID + 1000},
	} {
		if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{
			DepartmentId: uItoa(tc.id),
		})); connect.CodeOf(err) != connect.CodeNotFound {
			t.Fatalf("%s 的部門應 not_found(不是 failed_precondition),got %v", tc.name, err)
		}
	}

	// 對照組:存續且無成員者仍可刪 —— 證明上面兩個失敗不是「條件永遠不成立」。
	free := db.Department.Create().SetName("可刪").SetCompanyID(co.ID).SaveX(ctx)
	if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{
		DepartmentId: uItoa(free.ID),
	})); err != nil {
		t.Fatalf("無成員的存續部門必須可刪除: %v", err)
	}
	if got := db.Department.GetX(ctx, free.ID); got.DeletedAt == nil {
		t.Fatal("刪除必須標記 deleted_at")
	}
}

// TestDepartmentQueriesExcludeSoftDeleted 00020:以 DB 直寫 deleted_at 模擬歷史遺留的軟刪除列 ——
// ListDepartments(含 total 與 company_id 篩選)與 GetDepartment 一律看不到,否則已刪除部門
// 會回到清單與詳情頁(硬刪除時代不會發生,是軟刪除引入的破口)。
func TestDepartmentQueriesExcludeSoftDeleted(t *testing.T) {
	ctx := t.Context()
	super := authz.Identity{UserID: "1", CompanyID: "1", Role: "super", Roles: []string{"super"}}
	_, dc, db := newTestServerWithIdentity(t, super)
	co := db.Company.Create().SetName("公司A").SetIdentifier("DEPT-KEEP").SaveX(ctx)
	alive := db.Department.Create().SetName("存續").SetCompanyID(co.ID).SaveX(ctx)
	gone := db.Department.Create().SetName("已刪").SetCompanyID(co.ID).SaveX(ctx)
	if _, err := db.Department.UpdateOneID(gone.ID).SetDeletedAt(time.Now().UTC()).Save(ctx); err != nil {
		t.Fatalf("標記軟刪除: %v", err)
	}

	if _, err := dc.GetDepartment(ctx, connect.NewRequest(&v1.GetDepartmentRequest{DepartmentId: uItoa(gone.ID)})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("已刪除部門在 GetDepartment 應 NotFound,got %v", err)
	}
	for _, tc := range []struct {
		name string
		req  *v1.ListDepartmentsRequest
	}{
		{"全域", &v1.ListDepartmentsRequest{}},
		{"帶 company_id 篩選", &v1.ListDepartmentsRequest{CompanyId: uItoa(co.ID)}},
	} {
		list, err := dc.ListDepartments(ctx, connect.NewRequest(tc.req))
		if err != nil {
			t.Fatalf("ListDepartments(%s): %v", tc.name, err)
		}
		if list.Msg.GetPagination().GetTotal() != 1 || len(list.Msg.GetDepartments()) != 1 {
			t.Fatalf("ListDepartments(%s)應只剩 1 筆存續部門,got total=%d items=%d",
				tc.name, list.Msg.GetPagination().GetTotal(), len(list.Msg.GetDepartments()))
		}
		if got := list.Msg.GetDepartments()[0].GetId(); got != uItoa(alive.ID) {
			t.Fatalf("ListDepartments(%s)應為存續部門 %d,got %s", tc.name, alive.ID, got)
		}
	}
}

// TestDepartmentGuardsAgainstSoftDeletedAttachment 00020:回應 P2-A 的教訓 —— 軟刪除拿走了硬刪除
// (FK)的隱性保護,故「把資料掛到已刪除部門」的每一條可達路徑都必須被服務層擋下。
// user_service 是唯一以請求指定 department_id 的掛載路徑(CreateUser / UpdateUser / AssignRole),
// 不擋就會把活帳號掛進已刪除的部門(該帳號之後仍能通過登入與身分解析)。
func TestDepartmentGuardsAgainstSoftDeletedAttachment(t *testing.T) {
	ctx := t.Context()
	super := authz.Identity{UserID: "1", CompanyID: "1", Role: "super", Roles: []string{"super"}}
	_, _, db := newTestServerWithIdentity(t, super)
	uc := newUserTestServerWithDB(t, super, db)

	co := db.Company.Create().SetName("公司A").SetIdentifier("DEPT-GUARD").SaveX(ctx)
	deptA := db.Department.Create().SetName("部門甲").SetCompanyID(co.ID).SaveX(ctx)
	deptB := db.Department.Create().SetName("部門乙").SetCompanyID(co.ID).SaveX(ctx)
	if _, err := db.Department.UpdateOneID(deptB.ID).SetDeletedAt(time.Now().UTC()).Save(ctx); err != nil {
		t.Fatalf("標記軟刪除: %v", err)
	}

	// ① CreateUser 不得把新帳號掛進已刪除部門。
	_, err := uc.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
		Name: "新人", Email: "new@example.com", CompanyId: uItoa(co.ID), Role: "staff", DepartmentId: uItoa(deptB.ID),
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("掛到已刪除部門應 invalid_argument,got %v", err)
	}
	if n := db.User.Query().CountX(ctx); n != 0 {
		t.Fatalf("掛載被擋後不得留下帳號,got %d 筆", n)
	}

	// 對照組:存續部門可正常掛載(證明拒絕來自 deleted_at,而非其他失效原因)。
	created, err := uc.CreateUser(ctx, connect.NewRequest(&v1.CreateUserRequest{
		Name: "新人", Email: "new@example.com", CompanyId: uItoa(co.ID), Role: "staff", DepartmentId: uItoa(deptA.ID),
	}))
	if err != nil {
		t.Fatalf("存續部門應可正常建帳號: %v", err)
	}
	uid := created.Msg.GetUser().GetId()

	// ② UpdateUser 不得把既有帳號移進已刪除部門,且原部門不得被改動。
	_, err = uc.UpdateUser(ctx, connect.NewRequest(&v1.UpdateUserRequest{UserId: uid, DepartmentId: strPtr(uItoa(deptB.ID))}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("改掛到已刪除部門應 invalid_argument,got %v", err)
	}
	if got := userDeptID(t, db, uid); got != deptA.ID {
		t.Fatalf("被擋後使用者的部門應維持 %d,got %d", deptA.ID, got)
	}

	// ③ AssignRole 帶 department_id 亦不得掛進已刪除部門。
	_, err = uc.AssignRole(ctx, connect.NewRequest(&v1.AssignRoleRequest{UserId: uid, Role: "staff", DepartmentId: uItoa(deptB.ID)}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("指派角色並改掛到已刪除部門應 invalid_argument,got %v", err)
	}
	if got := userDeptID(t, db, uid); got != deptA.ID {
		t.Fatalf("被擋後使用者的部門應維持 %d,got %d", deptA.ID, got)
	}
}

// TestDeleteCompanyIgnoresSoftDeletedDepartments 00020:軟刪除的部門列會保留,故 DeleteCompany 的
// 「公司仍有部門」前置檢查必須排除它們 —— 否則一個部門全數軟刪除的公司將永遠刪不掉
// (軟刪除拿走硬刪除的隱性保護後的新破口)。
func TestDeleteCompanyIgnoresSoftDeletedDepartments(t *testing.T) {
	ctx := t.Context()
	super := authz.Identity{UserID: "1", CompanyID: "1", Role: "super", Roles: []string{"super"}}
	cc, dc, db := newTestServerWithIdentity(t, super)

	co := db.Company.Create().SetName("公司A").SetIdentifier("DEPT-CO-DEL").SaveX(ctx)
	alive := db.Department.Create().SetName("存續").SetCompanyID(co.ID).SaveX(ctx)
	gone := db.Department.Create().SetName("已刪").SetCompanyID(co.ID).SaveX(ctx)

	// 未刪除的部門仍阻擋公司刪除(既有行為)。
	if _, err := cc.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: uItoa(co.ID)})); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("仍有未刪除部門應 FailedPrecondition,got %v", err)
	}
	// 兩部門都軟刪除後即可刪除公司。
	if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{DepartmentId: uItoa(alive.ID)})); err != nil {
		t.Fatalf("前置刪除部門(存續): %v", err)
	}
	if _, err := dc.DeleteDepartment(ctx, connect.NewRequest(&v1.DeleteDepartmentRequest{DepartmentId: uItoa(gone.ID)})); err != nil {
		t.Fatalf("前置刪除部門(已刪): %v", err)
	}
	if _, err := cc.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{CompanyId: uItoa(co.ID)})); err != nil {
		t.Fatalf("部門全數軟刪除後應可刪除公司: %v", err)
	}
	if row := db.Company.GetX(ctx, co.ID); row.DeletedAt == nil {
		t.Fatal("公司刪除必須標記 deleted_at")
	}
}

// userDeptID 取使用者的部門 id(0 = 無部門);ent 的 FK 欄不公開,須 eager-load department edge。
func userDeptID(t *testing.T, db *ent.Client, uid string) int {
	t.Helper()
	id, err := parseID(uid)
	if err != nil {
		t.Fatalf("解析使用者 id %q: %v", uid, err)
	}
	u, err := db.User.Query().WithDepartment().Where(user.ID(id)).Only(t.Context())
	if err != nil {
		t.Fatalf("載入使用者 %s: %v", uid, err)
	}
	if u.Edges.Department == nil {
		return 0
	}
	return u.Edges.Department.ID
}
