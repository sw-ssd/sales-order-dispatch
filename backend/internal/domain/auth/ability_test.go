package auth_test

import (
	"context"
	"strings"
	"testing"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	auth "github.com/salesorder/sales-order-1.0/backend/internal/domain/auth"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// newTestHandler 建立 enttest sqlite 記憶體 client 與 GetAbility handler。
// DSN 以測試名區隔,避免同 package 多測試共享同一 in-memory db。
func newTestHandler(t *testing.T, developerEnabled bool) (*auth.AbilityHandler, *ent.Client, context.Context) {
	t.Helper()
	dsn := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return auth.NewAbilityHandler(client, auth.Config{DeveloperAccountEnabled: developerEnabled}), client, context.Background()
}

func TestGetAbilityWithConditions(t *testing.T) {
	h, client, ctx := newTestHandler(t, true)
	role := client.Role.Create().SetCode("staff").SetName("門市").SetDataScope("department").SetIsSystem(true).SaveX(ctx)
	client.RolePermission.Create().SetRoleID(role.ID).SetResource("sales_order").SetAction("cancel").
		SetConditions(map[string]any{"status": "pending"}).SaveX(ctx)
	client.RolePermission.Create().SetRoleID(role.ID).SetResource("sales_order").SetAction("read").
		SetConditions(map[string]any{"department_id": "${user.department_id}"}).SetSortOrder(1).SaveX(ctx)

	resp, err := h.GetAbility(
		authz.WithIdentity(ctx, authz.Identity{UserID: "u1", CompanyID: "c1", DepartmentID: "d1", Roles: []string{"staff"}}),
		connect.NewRequest(&v1.GetAbilityRequest{}),
	)
	if err != nil {
		t.Fatalf("GetAbility: %v", err)
	}
	rules := resp.Msg.GetRules()
	if len(rules) != 2 {
		t.Fatalf("rules = %d, want 2", len(rules))
	}

	if rules[0].GetAction() != "cancel" || rules[0].GetSubject() != "sales_order" {
		t.Fatalf("rules[0] = %s:%s, want cancel:sales_order", rules[0].GetAction(), rules[0].GetSubject())
	}
	if rules[0].GetInverted() {
		t.Fatal("rules[0].inverted = true, want false")
	}
	if got := rules[0].GetConditions().GetFields()["status"].GetStringValue(); got != "pending" {
		t.Fatalf("rules[0].conditions.status = %q, want pending", got)
	}

	if rules[1].GetAction() != "read" {
		t.Fatalf("rules[1].action = %q, want read", rules[1].GetAction())
	}
	// 佔位符已以身分展開為具體值。
	if got := rules[1].GetConditions().GetFields()["department_id"].GetStringValue(); got != "d1" {
		t.Fatalf("rules[1].conditions.department_id = %q, want d1(佔位符展開)", got)
	}
}

func TestGetAbilityDeveloper(t *testing.T) {
	t.Run("開關啟用回 manage-all", func(t *testing.T) {
		h, _, ctx := newTestHandler(t, true)
		resp, err := h.GetAbility(
			authz.WithIdentity(ctx, authz.Identity{Role: "developer", Roles: []string{"developer"}}),
			connect.NewRequest(&v1.GetAbilityRequest{}),
		)
		if err != nil {
			t.Fatalf("GetAbility: %v", err)
		}
		rules := resp.Msg.GetRules()
		if len(rules) != 1 {
			t.Fatalf("rules = %d, want 1", len(rules))
		}
		if rules[0].GetAction() != "manage" || rules[0].GetSubject() != "all" {
			t.Fatalf("rules[0] = %s:%s, want manage:all", rules[0].GetAction(), rules[0].GetSubject())
		}
		if rules[0].GetConditions() != nil || rules[0].GetInverted() {
			t.Fatalf("developer 規則應無條件且非 inverted: %+v", rules[0])
		}
	})
	t.Run("開關關閉視同一般身分", func(t *testing.T) {
		h, _, ctx := newTestHandler(t, false)
		resp, err := h.GetAbility(
			authz.WithIdentity(ctx, authz.Identity{Role: "developer", Roles: []string{"developer"}}),
			connect.NewRequest(&v1.GetAbilityRequest{}),
		)
		if err != nil {
			t.Fatalf("GetAbility: %v", err)
		}
		if rules := resp.Msg.GetRules(); len(rules) != 0 {
			t.Fatalf("rules = %d, want 0(開關關閉不特權)", len(rules))
		}
	})
}

func TestGetAbilityGuest(t *testing.T) {
	h, client, ctx := newTestHandler(t, true)
	// 即使系統內有 staff 角色規則,guest 身分不應取得任何規則。
	role := client.Role.Create().SetCode("staff").SetName("門市").SetDataScope("department").SetIsSystem(true).SaveX(ctx)
	client.RolePermission.Create().SetRoleID(role.ID).SetResource("sales_order").SetAction("read").SaveX(ctx)

	resp, err := h.GetAbility(
		authz.WithIdentity(ctx, authz.Identity{Roles: []string{"guest"}}),
		connect.NewRequest(&v1.GetAbilityRequest{}),
	)
	if err != nil {
		t.Fatalf("GetAbility: %v", err)
	}
	if rules := resp.Msg.GetRules(); len(rules) != 0 {
		t.Fatalf("guest rules = %d, want 0", len(rules))
	}
}

func TestGetAbilityPlaceholderFailClosedAndInverted(t *testing.T) {
	h, client, ctx := newTestHandler(t, true)
	role := client.Role.Create().SetCode("staff").SetName("門市").SetDataScope("department").SetIsSystem(true).SaveX(ctx)
	// 佔位符對應身分值為空 → 規則停用,不下發。
	client.RolePermission.Create().SetRoleID(role.ID).SetResource("sales_order").SetAction("read").
		SetConditions(map[string]any{"department_id": "${user.department_id}"}).SetSortOrder(1).SaveX(ctx)
	// inverted 規則應原樣下發;非 $eq 運算子輸出 {op: value} 形。
	client.RolePermission.Create().SetRoleID(role.ID).SetResource("sales_order").SetAction("cancel").
		SetConditions(map[string]any{"status": map[string]any{"$ne": "closed"}}).SetInverted(true).SetSortOrder(2).SaveX(ctx)

	resp, err := h.GetAbility(
		authz.WithIdentity(ctx, authz.Identity{UserID: "u1", CompanyID: "c1", Roles: []string{"staff"}}), // DepartmentID 空
		connect.NewRequest(&v1.GetAbilityRequest{}),
	)
	if err != nil {
		t.Fatalf("GetAbility: %v", err)
	}
	rules := resp.Msg.GetRules()
	if len(rules) != 1 {
		t.Fatalf("rules = %d, want 1(展開失敗規則不下發)", len(rules))
	}
	if rules[0].GetAction() != "cancel" || !rules[0].GetInverted() {
		t.Fatalf("rules[0] = %+v, want cancel + inverted", rules[0])
	}
	status := rules[0].GetConditions().GetFields()["status"].GetStructValue().GetFields()
	if got := status["$ne"].GetStringValue(); got != "closed" {
		t.Fatalf("conditions.status.$ne = %q, want closed", got)
	}
}

func TestGetAbilityNoIdentity(t *testing.T) {
	h, client, ctx := newTestHandler(t, true)
	role := client.Role.Create().SetCode("staff").SetName("門市").SetDataScope("department").SetIsSystem(true).SaveX(ctx)
	client.RolePermission.Create().SetRoleID(role.ID).SetResource("sales_order").SetAction("read").SaveX(ctx)

	resp, err := h.GetAbility(ctx, connect.NewRequest(&v1.GetAbilityRequest{}))
	if err != nil {
		t.Fatalf("GetAbility: %v", err)
	}
	// 未注入身分 → 零值 Roles → fail-closed 空規則。
	if rules := resp.Msg.GetRules(); len(rules) != 0 {
		t.Fatalf("rules = %d, want 0(未注入身分)", len(rules))
	}
}
