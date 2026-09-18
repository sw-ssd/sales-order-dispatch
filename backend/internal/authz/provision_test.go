package authz_test

import (
	"context"
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
)

// newProvisionEnv 建立 enttest sqlite + 記憶體 OpenFGA;回傳 engine 與 DB。
func newProvisionEnv(t *testing.T) (*openfga.Engine, *ent.Client) {
	t.Helper()
	db := enttest.Open(t, "sqlite3", "file:prov_"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	fgaClient, err := ofga.NewMemory(context.Background(), "prov-store")
	if err != nil {
		t.Fatalf("NewMemory: %v", err)
	}
	t.Cleanup(fgaClient.Close)
	return openfga.New(fgaClient), db
}

func itoa(v int) string { return strconv.Itoa(v) }

// TestProvisionRoleAbilities 驗證 role_permissions(非 inverted)translate 成 role→ability
// tuples,且 inverted 規則不轉為 allow tuple。
func TestProvisionRoleAbilities(t *testing.T) {
	e, db := newProvisionEnv(t)
	ctx := context.Background()
	r := db.Role.Create().SetCode("staff").SetName("門市").SetDataScope(role.DataScopeDepartment).SetIsSystem(true).SaveX(ctx)
	// read → can_read;cancel(inverted) → 不寫;update → can_write。
	db.RolePermission.Create().SetRoleID(r.ID).SetResource("sales_order").SetAction("read").SaveX(ctx)
	db.RolePermission.Create().SetRoleID(r.ID).SetResource("sales_order").SetAction("cancel").SetInverted(true).SaveX(ctx)
	db.RolePermission.Create().SetRoleID(r.ID).SetResource("customer").SetAction("update").SaveX(ctx)

	if err := authz.Provision(ctx, e, db); err != nil {
		t.Fatalf("Provision: %v", err)
	}
	userset := "role:" + itoa(r.ID) + "#assigned"
	if ok, _ := e.Check(ctx, userset, "can_read", "ability:sales_order"); !ok {
		t.Error("ability:sales_order can_read 應存在")
	}
	if ok, _ := e.Check(ctx, userset, "can_write", "ability:customer"); !ok {
		t.Error("ability:customer can_write 應存在")
	}
	if ok, _ := e.Check(ctx, userset, "can_write", "ability:sales_order"); ok {
		t.Error("inverted cancel 不得轉為 can_write allow tuple")
	}
}

// TestProvisionRemovesStaleTuples 驗證 reconcile 會清除殘留:角色權限移除、使用者角色變更後,
// 對應舊 tuples 應被刪除(P2 修正)。
func TestProvisionRemovesStaleTuples(t *testing.T) {
	e, db := newProvisionEnv(t)
	ctx := context.Background()
	r1 := db.Role.Create().SetCode("staff").SetName("門市").SetDataScope(role.DataScopeDepartment).SetIsSystem(true).SaveX(ctx)
	r2 := db.Role.Create().SetCode("admin").SetName("管理").SetDataScope(role.DataScopeCompany).SetIsSystem(true).SaveX(ctx)
	co := db.Company.Create().SetName("C公司").SetIdentifier("C-1").SaveX(ctx)

	// 初次:staff 具 sales_order read;使用者角色為 staff。
	perm := db.RolePermission.Create().SetRoleID(r1.ID).SetResource("sales_order").SetAction("read").SaveX(ctx)
	u := db.User.Create().SetEmail("a@x.com").SetName("A").SetStatus(user.StatusActive).SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)
	if err := authz.Provision(ctx, e, db); err != nil {
		t.Fatalf("Provision(初始): %v", err)
	}
	userset1 := "role:" + itoa(r1.ID) + "#assigned"
	if ok, _ := e.Check(ctx, userset1, "can_read", "ability:sales_order"); !ok {
		t.Fatal("初始 role→ability 應存在")
	}
	if ok, _ := e.Check(ctx, "user:"+itoa(u.ID), "assigned", "role:"+itoa(r1.ID)); !ok {
		t.Fatal("初始 user→role 應存在")
	}

	// 變更:移除該權限;使用者改為 admin 角色。
	db.RolePermission.DeleteOneID(perm.ID).ExecX(ctx)
	db.User.UpdateOneID(u.ID).SetRole("admin").SaveX(ctx)
	if err := authz.Provision(ctx, e, db); err != nil {
		t.Fatalf("Provision(變更): %v", err)
	}
	// 舊 role→ability 應已刪除。
	if ok, _ := e.Check(ctx, userset1, "can_read", "ability:sales_order"); ok {
		t.Error("移除權限後舊 role→ability tuple 應被刪除")
	}
	// 舊 user→role 應已刪除、新 user→role 應存在。
	if ok, _ := e.Check(ctx, "user:"+itoa(u.ID), "assigned", "role:"+itoa(r1.ID)); ok {
		t.Error("角色變更後舊 user→role tuple 應被刪除")
	}
	if ok, _ := e.Check(ctx, "user:"+itoa(u.ID), "assigned", "role:"+itoa(r2.ID)); !ok {
		t.Error("角色變更後新 user→role tuple 應存在")
	}
}

// TestProvisionUserRolesSkipsUnknownRole 驗證未知角色與非 active 使用者不被指派。
func TestProvisionUserRolesSkipsUnknownRole(t *testing.T) {
	e, db := newProvisionEnv(t)
	ctx := context.Background()
	r := db.Role.Create().SetCode("staff").SetName("門市").SetDataScope(role.DataScopeDepartment).SetIsSystem(true).SaveX(ctx)
	co := db.Company.Create().SetName("C公司").SetIdentifier("C-1").SaveX(ctx)
	active := db.User.Create().SetEmail("active@x.com").SetName("Active").SetStatus(user.StatusActive).SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)
	db.User.Create().SetEmail("pending@x.com").SetName("Pending").SetStatus(user.StatusPending).SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)
	unknown := db.User.Create().SetEmail("unknown@x.com").SetName("Unknown").SetStatus(user.StatusActive).SetRole("no_such_role").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)

	if err := authz.Provision(ctx, e, db); err != nil {
		t.Fatalf("Provision: %v", err)
	}
	if ok, err := e.Check(ctx, "user:"+itoa(active.ID), "assigned", "role:"+itoa(r.ID)); err != nil || !ok {
		t.Fatalf("active 使用者應有 role assigned tuple: ok=%v err=%v", ok, err)
	}
	if ok, _ := e.Check(ctx, "user:"+itoa(unknown.ID), "assigned", "role:"+itoa(r.ID)); ok {
		t.Error("未知角色使用者不應被指派至 staff role")
	}
}
