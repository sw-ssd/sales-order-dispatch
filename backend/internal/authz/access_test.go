package authz_test

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
)

// 註:01 計畫的 middleware.WithIdentity / testutil.CtxWithDB 尚未落地,
// 此處以 authz 自身的 WithIdentity/WithDB shim 取代(契約形狀相同,斷言不變)。

func TestAccessibleFilter(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	role := client.Role.Create().SetCode("staff").SetName("門市").SetDataScope("department").SetIsSystem(true).SaveX(ctx)
	client.RolePermission.Create().SetRoleID(role.ID).SetResource("sales_order").SetAction("read").
		SetConditions(map[string]any{"department_id": "${user.department_id}"}).SaveX(ctx)
	client.RolePermission.Create().SetRoleID(role.ID).SetResource("sales_order").SetAction("read").
		SetInverted(true).SetConditions(map[string]any{"status": "voided"}).SetSortOrder(1).SaveX(ctx)

	id := authz.Identity{UserID: "u1", CompanyID: "c1", DepartmentID: "d1", Roles: []string{"staff"}}

	t.Run("開啟時回條件片段", func(t *testing.T) {
		c := authz.WithCASLEnabled(authz.WithIdentity(authz.WithDB(ctx, client), id), true)
		clause, args, denied := authz.AccessibleFilter(c, "read", "sales_order")
		if denied {
			t.Fatal("denied = true, want false")
		}
		if clause != `("department_id" = ?) AND (NOT ("status" = ?))` {
			t.Fatalf("clause = %q", clause)
		}
		if len(args) != 2 || args[0] != "d1" || args[1] != "voided" {
			t.Fatalf("args = %#v, want [d1 voided]", args)
		}
	})
	t.Run("關閉時回空條件", func(t *testing.T) {
		c := authz.WithCASLEnabled(authz.WithIdentity(authz.WithDB(ctx, client), id), false)
		clause, args, denied := authz.AccessibleFilter(c, "read", "sales_order")
		if denied {
			t.Fatal("denied = true, want false")
		}
		if clause != "" {
			t.Fatalf("clause = %q, want empty", clause)
		}
		if args != nil {
			t.Fatalf("args = %#v, want nil", args)
		}
	})
	t.Run("無規則 resource denied", func(t *testing.T) {
		c := authz.WithCASLEnabled(authz.WithIdentity(authz.WithDB(ctx, client), id), true)
		_, _, denied := authz.AccessibleFilter(c, "read", "customer")
		if !denied {
			t.Fatal("denied = false, want true")
		}
	})
}
