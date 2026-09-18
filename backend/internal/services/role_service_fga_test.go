package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"
	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
)

// newRoleTestServerFGA 同 newRoleTestServer,但另注入 OpenFGA engine(ctx shim),
// 供 role_permissions 異動 → OpenFGA tuples 同步測試(D32/Task8)。
func newRoleTestServerFGA(t *testing.T, id authz.Identity) (salesorderv1connect.RoleServiceClient, *ent.Client, *openfga.Engine) {
	t.Helper()
	db := enttest.Open(t, "sqlite3", "file:entfga?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })

	fgaClient, err := ofga.NewMemory(context.Background(), "role-sync-store")
	if err != nil {
		t.Fatalf("NewMemory: %v", err)
	}
	t.Cleanup(fgaClient.Close)
	fga := openfga.New(fgaClient)

	mux := http.NewServeMux()
	RegisterRoleServices(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithCASLEnabled(ctx, true)
		ctx = authz.WithDB(ctx, db)
		ctx = authz.WithEngine(ctx, fga)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	client := salesorderv1connect.NewRoleServiceClient(http.DefaultClient, ts.URL)
	return client, db, fga
}

// TestUpdateRolePermissionsSyncsOpenFGA D32/Task8 驗收:UpdateRolePermissions 寫入
// role_permissions 後,將角色權限 translate 成 OpenFGA tuples(ability:<resource>#can_<action>
// @role:<roleID>#assigned)並同步;移除的權限對應 tuple 被刪除。
func TestUpdateRolePermissionsSyncsOpenFGA(t *testing.T) {
	ctx := context.Background()
	super := authz.Identity{UserID: "u0", CompanyID: "c1", Role: "super", Roles: []string{"super"}}
	client, db, fga := newRoleTestServerFGA(t, super)

	r := db.Role.Create().SetCode("regional").SetName("區域經理").SetDataScope(role.DataScopeCompany).SetIsSystem(false).SaveX(ctx)
	roleID := strconvID(r.ID)
	user := "role:" + roleID + "#assigned"

	// 寫入兩筆權限:sales_order/read(→can_read)、customer/read(→can_read)。
	_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
		RoleId: roleID,
		Permissions: []*v1.Permission{
			{Resource: "sales_order", Action: "read", SortOrder: 1},
			{Resource: "customer", Action: "read", SortOrder: 2},
		},
	}))
	if err != nil {
		t.Fatalf("UpdateRolePermissions: %v", err)
	}
	// sales_order read → can_read tuple 存在。
	ok, err := fga.Check(ctx, user, "can_read", "ability:sales_order")
	if err != nil || !ok {
		t.Fatalf("sales_order can_read tuple 應存在,ok=%v err=%v", ok, err)
	}
	// customer read → can_read tuple 存在。
	ok, err = fga.Check(ctx, user, "can_read", "ability:customer")
	if err != nil || !ok {
		t.Fatalf("customer can_read tuple 應存在,ok=%v err=%v", ok, err)
	}
	// 未授權 write → false。
	ok, err = fga.Check(ctx, user, "can_write", "ability:sales_order")
	if err != nil || ok {
		t.Fatalf("sales_order can_write 應不存在(未授予),ok=%v err=%v", ok, err)
	}

	// 全量取代:移除 customer/read,保留 sales_order/read,新增 company/read。
	_, err = client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
		RoleId: roleID,
		Permissions: []*v1.Permission{
			{Resource: "sales_order", Action: "read", SortOrder: 1},
			{Resource: "company", Action: "read", SortOrder: 2},
		},
	}))
	if err != nil {
		t.Fatalf("UpdateRolePermissions(取代): %v", err)
	}
	// 保留的 tuple 仍在。
	if ok, _ := fga.Check(ctx, user, "can_read", "ability:sales_order"); !ok {
		t.Fatal("取代後 sales_order can_read 應仍在")
	}
	// 新增的 tuple 寫入。
	if ok, _ := fga.Check(ctx, user, "can_read", "ability:company"); !ok {
		t.Fatal("取代後 company can_read 應存在")
	}
	// 移除的 tuple 已刪。
	if ok, _ := fga.Check(ctx, user, "can_read", "ability:customer"); ok {
		t.Fatal("取代後 customer can_read 應已被刪除")
	}
}
