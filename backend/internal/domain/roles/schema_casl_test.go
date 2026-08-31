package roles_test

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/rolepermission"
)

func TestRolePermissionCASLColumns(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	role := client.Role.Create().SetCode("staff").SetName("門市人員").
		SetDataScope("department").SetIsSystem(true).SaveX(ctx)

	t.Run("conditions/inverted/sort_order 寫入讀回", func(t *testing.T) {
		rp := client.RolePermission.Create().
			SetRoleID(role.ID).SetResource("sales_order").SetAction("cancel").
			SetConditions(map[string]any{"status": "pending"}).
			SetInverted(false).SetSortOrder(10).
			SaveX(ctx)
		got := client.RolePermission.GetX(ctx, rp.ID)
		if got.Conditions["status"] != "pending" {
			t.Fatalf("conditions = %#v, want status=pending", got.Conditions)
		}
		if got.Inverted {
			t.Fatal("inverted = true, want false")
		}
		if got.SortOrder != 10 {
			t.Fatalf("sort_order = %d, want 10", got.SortOrder)
		}
	})

	t.Run("同 resource×action 不同 conditions 可並存", func(t *testing.T) {
		client.RolePermission.Create().
			SetRoleID(role.ID).SetResource("sales_order").SetAction("read").
			SetConditions(map[string]any{"department_id": "${user.department_id}"}).SetSortOrder(1).
			SaveX(ctx)
		client.RolePermission.Create().
			SetRoleID(role.ID).SetResource("sales_order").SetAction("read").
			SetConditions(map[string]any{"created_by": "${user.id}"}).SetSortOrder(2).
			SaveX(ctx)
		n := client.RolePermission.Query().
			Where(rolepermission.RoleID(role.ID), rolepermission.Resource("sales_order"), rolepermission.Action("read")).
			CountX(ctx)
		if n != 2 {
			t.Fatalf("count = %d, want 2", n)
		}
	})

	t.Run("同 resource×action 相同 conditions 被唯一索引拒絕", func(t *testing.T) {
		client.RolePermission.Create().
			SetRoleID(role.ID).SetResource("customer").SetAction("read").
			SetConditions(map[string]any{"company_id": "c1"}).
			SaveX(ctx)
		_, err := client.RolePermission.Create().
			SetRoleID(role.ID).SetResource("customer").SetAction("read").
			SetConditions(map[string]any{"company_id": "c1"}).
			Save(ctx)
		if err == nil {
			t.Fatal("want unique index error, got nil")
		}
	})
}
