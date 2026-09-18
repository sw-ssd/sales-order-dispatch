package main

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
)

// newEntTestClient 以 sqlite 建立獨立 ent client(enttest 慣例,避免共享 cache)。
func newEntTestClient(t *testing.T) *ent.Client {
	t.Helper()
	return enttest.Open(t, "sqlite3", "file:seed_"+t.Name()+"?mode=memory&cache=shared&_fk=1")
}

func TestSeedBuiltinRoles(t *testing.T) {
	ctx := context.Background()
	client := newEntTestClient(t)
	if err := SeedBuiltinRoles(ctx, client); err != nil {
		t.Fatalf("SeedBuiltinRoles: %v", err)
	}
	want := map[string]role.DataScope{
		"super":         role.DataScopeAll,
		"company_admin": role.DataScopeCompany,
		"dept_admin":    role.DataScopeDepartment,
		"staff":         role.DataScopeDepartment,
		"customer":      role.DataScopeSelf,
		"guest":         role.DataScopeSelf,
		"developer":     role.DataScopeAll,
	}
	got, err := client.Role.Query().All(ctx)
	if err != nil {
		t.Fatalf("query roles: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("roles count = %d, want %d", len(got), len(want))
	}
	for _, r := range got {
		ds, ok := want[r.Code]
		if !ok {
			t.Errorf("unexpected role code %q", r.Code)
			continue
		}
		if r.DataScope != ds {
			t.Errorf("role %s data_scope = %s, want %s", r.Code, r.DataScope, ds)
		}
		if !r.IsSystem {
			t.Errorf("role %s is_system = false, want true", r.Code)
		}
	}
	// 冪等：再次執行不重複建立。
	if err := SeedBuiltinRoles(ctx, client); err != nil {
		t.Fatalf("SeedBuiltinRoles twice: %v", err)
	}
	n, err := client.Role.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != len(want) {
		t.Errorf("roles count after re-seed = %d, want %d", n, len(want))
	}
}

// TestSeedBuiltinRolePermissions 驗證權限 seed:種子全數落列、冪等不重複、
// conditions IS NULL(種子無條件規則)。
func TestSeedBuiltinRolePermissions(t *testing.T) {
	ctx := context.Background()
	client := newEntTestClient(t)
	if err := SeedBuiltinRoles(ctx, client); err != nil {
		t.Fatalf("SeedBuiltinRoles: %v", err)
	}
	want := auth.BuiltinRolePermissions()
	if err := SeedBuiltinRolePermissions(ctx, client); err != nil {
		t.Fatalf("SeedBuiltinRolePermissions: %v", err)
	}

	roleCode := map[int]string{}
	roles, err := client.Role.Query().All(ctx)
	if err != nil {
		t.Fatalf("query roles: %v", err)
	}
	for _, r := range roles {
		roleCode[r.ID] = r.Code
	}
	perms, err := client.RolePermission.Query().All(ctx)
	if err != nil {
		t.Fatalf("query perms: %v", err)
	}
	got := map[string]bool{}
	for _, p := range perms {
		got[roleCode[p.RoleID]+"\x00"+p.Resource+"\x00"+p.Action] = true
		if p.Conditions != nil {
			t.Errorf("種子權限 %s/%s conditions 應為 NULL", p.Resource, p.Action)
		}
	}
	// 每個種子均落列。
	for _, s := range want {
		if !got[s.Role+"\x00"+s.Resource+"\x00"+s.Action] {
			t.Errorf("缺少種子權限 %s/%s/%s", s.Role, s.Resource, s.Action)
		}
	}
	// 種子集合 = 落列集合(無多餘列)。
	if len(perms) != len(want) {
		t.Errorf("role_permissions 列數 = %d, want %d", len(perms), len(want))
	}
	// 冪等:再次執行不增加列數。
	if err := SeedBuiltinRolePermissions(ctx, client); err != nil {
		t.Fatalf("SeedBuiltinRolePermissions twice: %v", err)
	}
	n, err := client.RolePermission.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != len(want) {
		t.Errorf("re-seed 後列數 = %d, want %d", n, len(want))
	}
}
