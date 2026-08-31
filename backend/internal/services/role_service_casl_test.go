package services

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// TestUpdateRolePermissionsCaslValidation T10 驗收:UpdatePermissions 寫入前 CASL 條件驗證
// (未知欄位 / 非法 enum 值 → invalid_argument)與防鎖死(排除操作者自身的權限管理能力 →
// failed_precondition)。使用 newRoleTestServer(role_service_test.go)的 enttest sqlite 測試
// 中介層,身分注入與 server authzMiddleware 一致。
func TestUpdateRolePermissionsCaslValidation(t *testing.T) {
	super := authz.Identity{UserID: "u0", CompanyID: "c1", Role: "super", Roles: []string{"super"}}

	t.Run("未知欄位 invalid_argument", func(t *testing.T) {
		ctx := context.Background()
		client, db := newRoleTestServer(t, super)
		r := db.Role.Create().SetCode("staff").SetName("門市人員").SetDataScope(role.DataScopeDepartment).SetIsSystem(true).SaveX(ctx)
		_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
			RoleId: strconvID(r.ID),
			Permissions: []*v1.Permission{{Resource: "sales_order", Action: "read",
				Conditions: mustStruct(t, map[string]any{"nope": 1})}},
		}))
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Fatalf("未知欄位應 InvalidArgument,got %v", err)
		}
	})

	t.Run("非法 enum 值 invalid_argument", func(t *testing.T) {
		ctx := context.Background()
		client, db := newRoleTestServer(t, super)
		r := db.Role.Create().SetCode("staff").SetName("門市人員").SetDataScope(role.DataScopeDepartment).SetIsSystem(true).SaveX(ctx)
		_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
			RoleId: strconvID(r.ID),
			Permissions: []*v1.Permission{{Resource: "sales_order", Action: "read",
				Conditions: mustStruct(t, map[string]any{"status": "bogus"})}},
		}))
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Fatalf("非法 enum 值應 InvalidArgument,got %v", err)
		}
	})

	t.Run("未註冊資源的條件規則 invalid_argument", func(t *testing.T) {
		// 條件白名單(FieldRegistry)僅註冊 sales_order;role/policy 尚無欄位定義,
		// 帶條件的權限管理規則於驗證階段即被拒(fail-closed)。
		ctx := context.Background()
		client, db := newRoleTestServer(t, super)
		r := db.Role.Create().SetCode("staff").SetName("門市人員").SetDataScope(role.DataScopeDepartment).SetIsSystem(true).SaveX(ctx)
		_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
			RoleId: strconvID(r.ID),
			Permissions: []*v1.Permission{{Resource: "role", Action: "manage",
				Conditions: mustStruct(t, map[string]any{"company_id": "c1"})}},
		}))
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Fatalf("未註冊資源條件規則應 InvalidArgument,got %v", err)
		}
	})

	t.Run("防鎖死:排除自身的權限管理能力 failed_precondition", func(t *testing.T) {
		// super 更新自身 super 角色:cannot manage role(無條件 inverted 規則)→ 操作者
		// 喪失 role 資源管理能力 → failed_precondition。
		ctx := context.Background()
		client, db := newRoleTestServer(t, super)
		r := db.Role.Create().SetCode("super").SetName("超級管理員").SetDataScope(role.DataScopeAll).SetIsSystem(true).SaveX(ctx)
		_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
			RoleId: strconvID(r.ID),
			Permissions: []*v1.Permission{{Resource: "role", Action: "manage", Inverted: true}},
		}))
		if connect.CodeOf(err) != connect.CodeFailedPrecondition {
			t.Fatalf("防鎖死應 FailedPrecondition,got %v", err)
		}
	})

	t.Run("非自身角色不觸發防鎖死", func(t *testing.T) {
		// super 更新 staff 角色(非自身角色)的 cannot manage role:不觸發防鎖死。
		ctx := context.Background()
		client, db := newRoleTestServer(t, super)
		r := db.Role.Create().SetCode("staff").SetName("門市人員").SetDataScope(role.DataScopeDepartment).SetIsSystem(true).SaveX(ctx)
		_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
			RoleId: strconvID(r.ID),
			Permissions: []*v1.Permission{{Resource: "role", Action: "manage", Inverted: true}},
		}))
		if err != nil {
			t.Fatalf("非自身角色應不觸發防鎖死,got %v", err)
		}
	})

	t.Run("合法 conditions 寫入並可讀回", func(t *testing.T) {
		ctx := context.Background()
		client, db := newRoleTestServer(t, super)
		r := db.Role.Create().SetCode("staff").SetName("門市人員").SetDataScope(role.DataScopeDepartment).SetIsSystem(true).SaveX(ctx)
		upd, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
			RoleId: strconvID(r.ID),
			Permissions: []*v1.Permission{{Resource: "sales_order", Action: "cancel",
				Conditions: mustStruct(t, map[string]any{"status": "pending"})}},
		}))
		if err != nil {
			t.Fatalf("合法 conditions 應寫入成功,got %v", err)
		}
		if len(upd.Msg.GetPermissions()) != 1 {
			t.Fatalf("權限數應為 1,got %d", len(upd.Msg.GetPermissions()))
		}
		if got := upd.Msg.GetPermissions()[0].GetConditions().GetFields()["status"].GetStringValue(); got != "pending" {
			t.Fatalf("conditions 未回讀: got %q, want pending", got)
		}
	})
}

// TestUpdateRolePermissionsCompanyAdminPlaceholder T10 驗收:company_admin 的 id 類欄位值
// 須為 ${user.company_id} 佔位符或屬自己公司(佔位符以身分展開,允許寫入)。
func TestUpdateRolePermissionsCompanyAdminPlaceholder(t *testing.T) {
	ctx := context.Background()
	admin := authz.Identity{UserID: "u2", CompanyID: "c1", Role: "company_admin", Roles: []string{"company_admin"}}

	t.Run("company_id 佔位符允許", func(t *testing.T) {
		client, db := newRoleTestServer(t, admin)
		custom := db.Role.Create().SetCode("regional").SetName("區域經理").SetDataScope(role.DataScopeCompany).SetIsSystem(false).SaveX(ctx)
		_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
			RoleId: strconvID(custom.ID),
			Permissions: []*v1.Permission{{Resource: "sales_order", Action: "read",
				Conditions: mustStruct(t, map[string]any{"company_id": "${user.company_id}"})}},
		}))
		if err != nil {
			t.Fatalf("company_id 佔位符應允許,got %v", err)
		}
	})

	t.Run("company_id 其他公司拒絕", func(t *testing.T) {
		client, db := newRoleTestServer(t, admin)
		custom := db.Role.Create().SetCode("regional").SetName("區域經理").SetDataScope(role.DataScopeCompany).SetIsSystem(false).SaveX(ctx)
		_, err := client.UpdateRolePermissions(ctx, connect.NewRequest(&v1.UpdateRolePermissionsRequest{
			RoleId: strconvID(custom.ID),
			Permissions: []*v1.Permission{{Resource: "sales_order", Action: "read",
				Conditions: mustStruct(t, map[string]any{"company_id": "c2"})}},
		}))
		if connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("其他公司應 PermissionDenied,got %v", err)
		}
	})
}

// TestListConditionFields T10 驗收:回傳資源的條件欄位白名單(欄位、型別、運算子、enum 值);
// 未知資源回空陣列。
func TestListConditionFields(t *testing.T) {
	ctx := context.Background()
	super := authz.Identity{UserID: "u0", CompanyID: "c1", Role: "super", Roles: []string{"super"}}
	client, _ := newRoleTestServer(t, super)

	resp, err := client.ListConditionFields(ctx, connect.NewRequest(&v1.ListConditionFieldsRequest{Resource: "sales_order"}))
	if err != nil {
		t.Fatalf("ListConditionFields: %v", err)
	}
	fields := resp.Msg.GetFields()
	byName := make(map[string]*v1.ConditionField, len(fields))
	for _, f := range fields {
		byName[f.GetField()] = f
	}

	status, ok := byName["status"]
	if !ok {
		t.Fatalf("fields 應含 status,got %#v", fields)
	}
	if status.GetType() != "enum" {
		t.Fatalf("status.type 應為 enum,got %q", status.GetType())
	}
	if !containsString(status.GetOps(), "$eq") {
		t.Fatalf("status.ops 應含 $eq,got %v", status.GetOps())
	}
	if !containsString(status.GetEnum(), "pending") {
		t.Fatalf("status.enum 應含 pending,got %v", status.GetEnum())
	}

	dept, ok := byName["department_id"]
	if !ok {
		t.Fatalf("fields 應含 department_id,got %#v", fields)
	}
	if dept.GetType() != "id" {
		t.Fatalf("department_id.type 應為 id,got %q", dept.GetType())
	}

	// 未知資源回空陣列。
	empty, err := client.ListConditionFields(ctx, connect.NewRequest(&v1.ListConditionFieldsRequest{Resource: "nope"}))
	if err != nil {
		t.Fatalf("ListConditionFields(unknown): %v", err)
	}
	if len(empty.Msg.GetFields()) != 0 {
		t.Fatalf("未知資源應回空,got %#v", empty.Msg.GetFields())
	}
}

func containsString(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
