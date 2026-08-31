package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/ent/rolepermission"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz/casl"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1/salesorderv1connect"

	"google.golang.org/protobuf/types/known/structpb"
)

// newRoleTestServer 建立 enttest sqlite client + 注入身分的 RoleService HTTP server。
// 以測試中介層模擬 server authzMiddleware(身分 + CASL 開關 + DB 注入 ctx)。
func newRoleTestServer(t *testing.T, id authz.Identity) (salesorderv1connect.RoleServiceClient, *ent.Client) {
	t.Helper()
	db := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })

	mux := http.NewServeMux()
	RegisterRoleServices(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), id)
		ctx = authz.WithCASLEnabled(ctx, true)
		ctx = authz.WithDB(ctx, db)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	client := salesorderv1connect.NewRoleServiceClient(http.DefaultClient, ts.URL)
	return client, db
}

// mustStruct 建立 structpb.Struct(map 必須 JSON 相容)。
func mustStruct(t *testing.T, m map[string]any) *structpb.Struct {
	t.Helper()
	s, err := structpb.NewStruct(m)
	if err != nil {
		t.Fatalf("structpb.NewStruct: %v", err)
	}
	return s
}

// staffEvaluator 依 DB 內 staff 角色規則建立 CASL evaluator(模擬 authz 規則載入)。
func staffEvaluator(t *testing.T, db *ent.Client, ctx context.Context) *casl.Evaluator {
	t.Helper()
	rows, err := db.RolePermission.Query().
		Where(rolepermission.HasRoleWith(role.CodeEQ("staff"))).
		Order(rolepermission.BySortOrder()).
		All(ctx)
	if err != nil {
		t.Fatalf("載入 staff 規則: %v", err)
	}
	rules := make([]casl.Rule, 0, len(rows))
	for _, rp := range rows {
		conds, err := casl.ParseConditions(rp.Conditions)
		if err != nil {
			t.Fatalf("ParseConditions(%#v): %v", rp.Conditions, err)
		}
		rules = append(rules, casl.Rule{Action: rp.Action, Subject: rp.Resource, Conditions: conds, Inverted: rp.Inverted})
	}
	return casl.NewEvaluator(rules, casl.Identity{UserID: "u1", CompanyID: "c1", DepartmentID: "d1"})
}

// TestRolePermissionCRUD 驗收 brief Step 4:更新權限後新規則生效(經 CASL evaluator 強制)。
func TestRolePermissionCRUD(t *testing.T) {
	ctx := context.Background()
	super := authz.Identity{UserID: "u0", CompanyID: "c1", Role: "super", Roles: []string{"super"}}
	client, db := newRoleTestServer(t, super)

	db.Role.Create().SetCode("staff").SetName("門市人員").SetDataScope(role.DataScopeDepartment).SetIsSystem(true).SaveX(ctx)

	// ListRoles:可列出角色。
	list, err := client.ListRoles(ctx, connect.NewRequest(&v1.ListRolesRequest{}))
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if len(list.Msg.GetRoles()) != 1 || list.Msg.GetRoles()[0].GetCode() != "staff" {
		t.Fatalf("ListRoles 結果不符: %#v", list.Msg.GetRoles())
	}

	// GetRolePermissions:初始為空。
	got, err := client.GetRolePermissions(ctx, connect.NewRequest(&v1.GetRolePermissionsRequest{RoleId: list.Msg.GetRoles()[0].GetId()}))
	if err != nil {
		t.Fatalf("GetRolePermissions: %v", err)
	}
	if len(got.Msg.GetPermissions()) != 0 {
		t.Fatalf("初始權限應為空,got %#v", got.Msg.GetPermissions())
	}

	// UpdateRolePermissions:寫入帶條件與無條件規則。
	upd, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
		RoleId: list.Msg.GetRoles()[0].GetId(),
		Permissions: []*v1.Permission{
			{Resource: "sales_order", Action: "read", SortOrder: 1},
			{Resource: "sales_order", Action: "cancel", Conditions: mustStruct(t, map[string]any{"status": "pending"}), SortOrder: 2},
		},
	}))
	if err != nil {
		t.Fatalf("UpdateRolePermissions: %v", err)
	}
	if len(upd.Msg.GetPermissions()) != 2 {
		t.Fatalf("更新後應有 2 筆權限,got %d", len(upd.Msg.GetPermissions()))
	}

	// 回讀:依 sort_order 升冪。
	got, err = client.GetRolePermissions(ctx, connect.NewRequest(&v1.GetRolePermissionsRequest{RoleId: list.Msg.GetRoles()[0].GetId()}))
	if err != nil {
		t.Fatalf("GetRolePermissions: %v", err)
	}
	perms := got.Msg.GetPermissions()
	if len(perms) != 2 || perms[0].GetAction() != "read" || perms[1].GetAction() != "cancel" {
		t.Fatalf("權限順序/內容不符: %#v", perms)
	}
	if perms[1].GetConditions().GetFields()["status"].GetStringValue() != "pending" {
		t.Fatalf("conditions 未回讀: %#v", perms[1].GetConditions())
	}

	// 新規則強制生效(CASL evaluator 由 DB 規則載入,與 authz.loadRules 同構)。
	e := staffEvaluator(t, db, ctx)
	if !e.Can("cancel", "sales_order", map[string]any{"status": "pending"}) {
		t.Error("cancel(pending) 應允許(新規則生效)")
	}
	if e.Can("cancel", "sales_order", map[string]any{"status": "processing"}) {
		t.Error("cancel(processing) 應拒絕(條件不命中)")
	}
	if !e.Can("read", "sales_order", nil) {
		t.Error("read 無條件規則應命中")
	}
	if e.Can("delete", "sales_order", nil) {
		t.Error("未授權動作應拒絕")
	}
}

// TestRolePermissionCompanyScope 驗收「company_admin 限自己公司」:
// 不可改內建角色;conditions.company_id 僅可為自身公司。
func TestRolePermissionCompanyScope(t *testing.T) {
	ctx := context.Background()
	admin := authz.Identity{UserID: "u2", CompanyID: "c1", Role: "company_admin", Roles: []string{"company_admin"}}

	t.Run("company_admin 不可修改內建角色", func(t *testing.T) {
		client, db := newRoleTestServer(t, admin)
		sys := db.Role.Create().SetCode("staff").SetName("門市人員").SetDataScope(role.DataScopeDepartment).SetIsSystem(true).SaveX(ctx)
		_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
			RoleId:      strconvID(sys.ID),
			Permissions: []*v1.Permission{{Resource: "sales_order", Action: "read"}},
		}))
		if connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("company_admin 改內建角色應 PermissionDenied,got %v", err)
		}
	})

	t.Run("company_admin 條件不得引用其他公司", func(t *testing.T) {
		client, db := newRoleTestServer(t, admin)
		custom := db.Role.Create().SetCode("regional").SetName("區域經理").SetDataScope(role.DataScopeCompany).SetIsSystem(false).SaveX(ctx)
		_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
			RoleId: strconvID(custom.ID),
			Permissions: []*v1.Permission{
				{Resource: "sales_order", Action: "read", Conditions: mustStruct(t, map[string]any{"company_id": "c2"})},
			},
		}))
		if connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("引用其他公司應 PermissionDenied,got %v", err)
		}
	})

	t.Run("company_admin 可管理自訂角色且限自己公司", func(t *testing.T) {
		client, db := newRoleTestServer(t, admin)
		custom := db.Role.Create().SetCode("regional").SetName("區域經理").SetDataScope(role.DataScopeCompany).SetIsSystem(false).SaveX(ctx)
		_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
			RoleId: strconvID(custom.ID),
			Permissions: []*v1.Permission{
				{Resource: "sales_order", Action: "read", Conditions: mustStruct(t, map[string]any{"company_id": "c1"}), SortOrder: 1},
				{Resource: "customer", Action: "read", SortOrder: 2},
			},
		}))
		if err != nil {
			t.Fatalf("company_admin 改自訂角色應允許,got %v", err)
		}
	})
}

// TestRolePermissionDenied 驗收權限門檻:無身分 / 無管理權角色一律拒絕。
func TestRolePermissionDenied(t *testing.T) {
	ctx := context.Background()

	t.Run("未登入 Unauthenticated", func(t *testing.T) {
		client, _ := newRoleTestServer(t, authz.Identity{})
		_, err := client.ListRoles(ctx, connect.NewRequest(&v1.ListRolesRequest{}))
		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Fatalf("未登入應 Unauthenticated,got %v", err)
		}
	})

	for _, id := range []authz.Identity{
		{UserID: "u3", CompanyID: "c1", Role: "staff", Roles: []string{"staff"}},
		{UserID: "u4", CompanyID: "c1", Role: "guest", Roles: []string{"guest"}},
		{UserID: "u5", CompanyID: "c1", Role: "dept_admin", Roles: []string{"dept_admin"}},
	} {
		client, _ := newRoleTestServer(t, id)
		_, err := client.ListRoles(ctx, connect.NewRequest(&v1.ListRolesRequest{}))
		if connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("%s ListRoles 應 PermissionDenied,got %v", id.Role, err)
		}
	}
}

// TestUpdateRolePermissionsValidation 輸入驗證:空欄位 / 重複規則 / 不存在角色。
func TestUpdateRolePermissionsValidation(t *testing.T) {
	ctx := context.Background()
	super := authz.Identity{UserID: "u0", CompanyID: "c1", Role: "super", Roles: []string{"super"}}
	client, db := newRoleTestServer(t, super)
	custom := db.Role.Create().SetCode("regional").SetName("區域經理").SetDataScope(role.DataScopeCompany).SaveX(ctx)
	roleID := strconvID(custom.ID)

	_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
		RoleId:      roleID,
		Permissions: []*v1.Permission{{Resource: " ", Action: "read"}},
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("空 resource 應 InvalidArgument,got %v", err)
	}

	_, err = client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
		RoleId: roleID,
		Permissions: []*v1.Permission{
			{Resource: "sales_order", Action: "read"},
			{Resource: "sales_order", Action: "read"},
		},
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("重複規則應 InvalidArgument,got %v", err)
	}

	_, err = client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
		RoleId:      "99999",
		Permissions: []*v1.Permission{{Resource: "sales_order", Action: "read"}},
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("不存在角色應 NotFound,got %v", err)
	}
}

func strconvID(id int) string {
	return fmt.Sprintf("%d", id)
}
