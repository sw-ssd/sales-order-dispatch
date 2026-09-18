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

// newProvisionEnv 建立 enttest sqlite + 記憶體 OpenFGA;回傳 engine。
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

// TestProvisionRoleAbilities 驗證 role_permissions(非 inverted)translate 成
// role→ability tuples,且 inverted 規則不轉為 allow tuple(D32/Task8+P1)。
func TestProvisionRoleAbilities(t *testing.T) {
	e, db := newProvisionEnv(t)
	ctx := context.Background()
	r := db.Role.Create().SetCode("staff").SetName("門市").SetDataScope(role.DataScopeDepartment).SetIsSystem(true).SaveX(ctx)
	// read → can_read;cancel → can_write;inverted cancel → 不寫。
	db.RolePermission.Create().SetRoleID(r.ID).SetResource("sales_order").SetAction("read").SaveX(ctx)
	db.RolePermission.Create().SetRoleID(r.ID).SetResource("sales_order").SetAction("cancel").SetInverted(true).SaveX(ctx)
	db.RolePermission.Create().SetRoleID(r.ID).SetResource("customer").SetAction("update").SaveX(ctx)

	if err := authz.ProvisionRoleAbilities(ctx, e, db); err != nil {
		t.Fatalf("ProvisionRoleAbilities: %v", err)
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

// TestProvisionUserRoles 驗證 active 使用者 → role assigned tuple;未知角色略過。
func TestProvisionUserRoles(t *testing.T) {
	e, db := newProvisionEnv(t)
	ctx := context.Background()
	r := db.Role.Create().SetCode("staff").SetName("門市").SetDataScope(role.DataScopeDepartment).SetIsSystem(true).SaveX(ctx)
	co := db.Company.Create().SetName("C公司").SetIdentifier("C-1").SaveX(ctx)
	db.User.Create().SetEmail("active@x.com").SetName("Active").SetStatus(user.StatusActive).SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)
	db.User.Create().SetEmail("pending@x.com").SetName("Pending").SetStatus(user.StatusPending).SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)
	db.User.Create().SetEmail("unknown@x.com").SetName("Unknown").SetStatus(user.StatusActive).SetRole("no_such_role").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)

	if err := authz.ProvisionUserRoles(ctx, e, db); err != nil {
		t.Fatalf("ProvisionUserRoles: %v", err)
	}
	// active 使用者應 assigned 至 role:<id>。
	active := db.User.Query().Where(user.EmailEQ("active@x.com")).OnlyX(ctx)
	if ok, err := e.Check(ctx, "user:"+itoa(active.ID), "assigned", "role:"+itoa(r.ID)); err != nil || !ok {
		t.Fatalf("active 使用者應有 role assigned tuple: ok=%v err=%v", ok, err)
	}
	// unknown 角色使用者不得被指派(無對應 role → 略過)。
	unknown := db.User.Query().Where(user.EmailEQ("unknown@x.com")).OnlyX(ctx)
	if ok, _ := e.Check(ctx, "user:"+itoa(unknown.ID), "assigned", "role:"+itoa(r.ID)); ok {
		t.Error("未知角色使用者不應被指派至 staff role")
	}
}
