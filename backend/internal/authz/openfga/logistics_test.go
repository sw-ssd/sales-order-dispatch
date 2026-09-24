package openfga_test

import (
	"context"
	"testing"
)

// TestLogisticsDeliveryTenantAndAssigneeEdges 是 logistics model(D32/10.8)的契約測試:
//
//   - 租戶 parent 邊:delivery 掛 company/department tuple(物件建立時經 AfterCommit 寫入);
//   - 司機指派:delivery#driver 寫 userset 主體「driver:<id>#assignee」→ 司機本人
//     沿 driver#assignee 到達 can_read/can_write(10.8:被指派司機可操作其 delivery);
//   - assigned_by(指派人)對同一 delivery 亦可讀寫;
//   - 未指派且無部門資格者一律 deny(他人 403);
//   - 重指派:以最新 tuple 為準 —— 判決完全由 tuples 承載,刪舊寫新即完成移交。
//
// 註:department#member/#admin 在生產由 role_permissions/users 派生的成員 provision 寫入
// (現行 provision 只產 role→ability 與 user→role);此測試自行組出成員邊,驗證 model
// 改寫式的正確性,不依賴 provision 擴充(D32 本批不含成員 provision)。
func TestLogisticsDeliveryTenantAndAssigneeEdges(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	e := newEngine(t)

	write := func(user, relation, object string) {
		t.Helper()
		if err := e.WriteTuple(ctx, user, relation, object); err != nil {
			t.Fatalf("WriteTuple(%s,%s,%s): %v", user, relation, object, err)
		}
	}
	check := func(user, relation, object string) bool {
		t.Helper()
		ok, err := e.Check(ctx, user, relation, object)
		if err != nil {
			t.Fatalf("Check(%s,%s,%s): %v", user, relation, object, err)
		}
		return ok
	}

	// 部門成員/管理邊(生產由成員 provision 寫;此處手工組出)。
	write("role:staff#assigned", "member", "department:100")
	write("role:dept_admin#assigned", "admin", "department:100")
	write("user:manager1", "assigned", "role:dept_admin")
	// 司機物件:d1/d2 的被指派人身分(assignee: [user] → subject=user,object=driver)。
	write("user:driver1", "assignee", "driver:d1")
	write("user:driver2", "assignee", "driver:d2")
	// delivery:f1 的租戶 parent 邊 + 指派(10.4 AssignmentService 的 tuple 合約;
	// subject 一律 userset 形式 —— company:X#member / department:X#member / #admin,
	// 物件建立時經 AfterCommit 寫入)。
	write("company:1#member", "company", "logistics_delivery:f1")
	write("department:100#member", "department", "logistics_delivery:f1")
	write("department:100#admin", "manager", "logistics_delivery:f1")
	write("driver:d1#assignee", "driver", "logistics_delivery:f1")
	write("user:manager1", "assigned_by", "logistics_delivery:f1")
	// manager2 同為 dept_admin 但**不是** assigned_by:寫權限應走 department#admin → manager 邊。
	write("user:manager2", "assigned", "role:dept_admin")

	t.Run("被指派司機可讀寫自己的 delivery", func(t *testing.T) {
		if !check("user:driver1", "can_read", "logistics_delivery:f1") {
			t.Fatal("被指派司機應可讀")
		}
		if !check("user:driver1", "can_write", "logistics_delivery:f1") {
			t.Fatal("被指派司機應可寫(Start/Complete/Cancel 路徑)")
		}
	})
	t.Run("指派人可讀寫", func(t *testing.T) {
		if !check("user:manager1", "can_read", "logistics_delivery:f1") {
			t.Fatal("assigned_by 應可讀")
		}
		if !check("user:manager1", "can_write", "logistics_delivery:f1") {
			t.Fatal("assigned_by 應可寫")
		}
	})
	t.Run("部門管理者的寫權限走 department#admin → manager 邊", func(t *testing.T) {
		if !check("user:manager2", "can_write", "logistics_delivery:f1") {
			t.Fatal("dept_admin(非指派人)應可寫(manager 邊)")
		}
		if !check("user:manager2", "can_read", "logistics_delivery:f1") {
			t.Fatal("dept_admin 應可讀")
		}
	})
	t.Run("部門成員可讀、未指派者 403", func(t *testing.T) {
		// role:staff#assigned 是 department:100 的 member → 該角色的使用者可讀。
		if err := e.WriteTuple(ctx, "user:staff1", "assigned", "role:staff"); err != nil {
			t.Fatalf("WriteTuple(role assign): %v", err)
		}
		if !check("user:staff1", "can_read", "logistics_delivery:f1") {
			t.Fatal("部門成員應可讀")
		}
		if check("user:staff1", "can_write", "logistics_delivery:f1") {
			t.Fatal("未被指派的成員不得寫(寫限 dept_admin/被指派者/指派人)")
		}
		if check("user:stranger", "can_read", "logistics_delivery:f1") {
			t.Fatal("無任何邊的使用者不得讀(他人 403)")
		}
	})
	t.Run("重指派:判決跟著 tuple 走", func(t *testing.T) {
		if err := e.DeleteTuple(ctx, "driver:d1#assignee", "driver", "logistics_delivery:f1"); err != nil {
			t.Fatalf("DeleteTuple(舊指派): %v", err)
		}
		write("driver:d2#assignee", "driver", "logistics_delivery:f1")
		if check("user:driver1", "can_read", "logistics_delivery:f1") {
			t.Fatal("被替換的司機不得再讀")
		}
		if !check("user:driver2", "can_read", "logistics_delivery:f1") {
			t.Fatal("新司機應可讀")
		}
		if !check("user:driver2", "can_write", "logistics_delivery:f1") {
			t.Fatal("新司機應可寫")
		}
	})
}
