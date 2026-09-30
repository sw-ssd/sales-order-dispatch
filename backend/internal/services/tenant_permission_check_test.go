package services

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	v1 "github.com/salesorder/sales-order-1.0/backend/contracts/proto/salesorder/v1"
)

// TestServiceLayerPermissionOnDBRolePermissions 驗證服務層的逐動作授權以 DB
// role_permissions 為準(D33),且不會因為二元閘門而放行過寬。
//
// 這個測試釘住的是 OpenFGA 閘門 **無法**表達的區分:middleware 只判 can_read/can_write,
// 而 company_admin 有 company/update 卻沒有 company/create|delete —— 若服務層不比對
// 逐動作權限,「update」的二元授權就會放行「create」與「delete」。
//
// 另外驗證自訂角色:自訂角色只存在於 role_permissions 表,內建 ACL(rolePolicy)看不懂,
// 故必須由 DB 查詢提供答案(D33 之前自訂角色在服務層一律被拒)。
func TestServiceLayerPermissionOnDBRolePermissions(t *testing.T) {
	ctx := context.Background()

	t.Run("company_admin:DB 有 update 無 create → create 拒、update 准", func(t *testing.T) {
		cc, _, db := newTestServerWithIdentity(t, authz.Identity{
			UserID: "2", CompanyID: "c1", Role: "company_admin", Roles: auth.RolesFor("company_admin"),
		})
		r := db.Role.Create().SetCode("company_admin").SetName("公司管理員").
			SetDataScope("company").SetIsSystem(true).SaveX(ctx)
		for _, act := range []string{"read", "update"} {
			db.RolePermission.Create().SetRoleID(r.ID).SetResource("company").SetAction(act).SetInverted(false).SaveX(ctx)
		}
		co := db.Company.Create().SetName("公司A").SetIdentifier("A-1").SaveX(ctx)

		if _, err := cc.UpdateCompany(ctx, connect.NewRequest(&v1.UpdateCompanyRequest{
			CompanyId: strconvID(co.ID), Name: protoStr("公司A改"),
		})); err != nil {
			t.Fatalf("DB 有 company/update,UpdateCompany 應通過,got %v", err)
		}
		if _, err := cc.CreateCompany(ctx, connect.NewRequest(&v1.CreateCompanyRequest{
			Name: "公司B", Identifier: "B-1",
		})); connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("DB 無 company/create,CreateCompany 應 PermissionDenied,got %v", err)
		}
		if _, err := cc.DeleteCompany(ctx, connect.NewRequest(&v1.DeleteCompanyRequest{
			CompanyId: strconvID(co.ID),
		})); connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("DB 無 company/delete,DeleteCompany 應 PermissionDenied,got %v", err)
		}
	})

	t.Run("自訂角色:DB 授予即可通過(內建 ACL 看不懂自訂角色)", func(t *testing.T) {
		cc, _, db := newTestServerWithIdentity(t, authz.Identity{
			UserID: "9", CompanyID: "c1", Role: "auditor", Roles: auth.RolesFor("auditor"),
		})
		r := db.Role.Create().SetCode("auditor").SetName("稽核員").
			SetDataScope("company").SetIsSystem(false).SaveX(ctx)
		db.RolePermission.Create().SetRoleID(r.ID).SetResource("company").SetAction("read").SetInverted(false).SaveX(ctx)
		db.Company.Create().SetName("公司A").SetIdentifier("A-1").SaveX(ctx)

		if _, err := cc.ListCompanies(ctx, connect.NewRequest(&v1.ListCompaniesRequest{})); err != nil {
			t.Fatalf("DB 授予 company/read,自訂角色 ListCompanies 應通過,got %v", err)
		}
	})

	t.Run("inverted 規則不構成授權", func(t *testing.T) {
		cc, _, db := newTestServerWithIdentity(t, authz.Identity{
			UserID: "8", CompanyID: "c1", Role: "probe", Roles: auth.RolesFor("probe"),
		})
		r := db.Role.Create().SetCode("probe").SetName("探針").SetDataScope("company").SetIsSystem(false).SaveX(ctx)
		// 只有 inverted(拒絕)規則,不得轉為允許。
		db.RolePermission.Create().SetRoleID(r.ID).SetResource("company").SetAction("read").SetInverted(true).SaveX(ctx)

		if _, err := cc.ListCompanies(ctx, connect.NewRequest(&v1.ListCompaniesRequest{})); connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Fatalf("只有 inverted 規則時應 PermissionDenied,got %v", err)
		}
	})
}
