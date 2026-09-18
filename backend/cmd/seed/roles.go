package main

import (
	"context"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
)

// builtinRole 為 7 個內建角色(code/name/data_scope/is_system=true)(D9)。
// code 與 data_scope 不可修改、內建角色不可刪除(於 RoleService 防護)。
type builtinRole struct {
	Code      string
	Name      string
	DataScope role.DataScope
}

var builtinRoles = []builtinRole{
	{Code: "super", Name: "超級管理員", DataScope: role.DataScopeAll},
	{Code: "company_admin", Name: "公司管理員", DataScope: role.DataScopeCompany},
	{Code: "dept_admin", Name: "部門管理員", DataScope: role.DataScopeDepartment},
	{Code: "staff", Name: "一般員工", DataScope: role.DataScopeDepartment},
	{Code: "customer", Name: "客戶", DataScope: role.DataScopeSelf},
	{Code: "guest", Name: "待審核訪客", DataScope: role.DataScopeSelf},
	{Code: "developer", Name: "開發人員（逃生門）", DataScope: role.DataScopeAll},
}

// SeedBuiltinRoles 冪等建立 7 內建角色(依 code 查詢,已存在即略過)。
func SeedBuiltinRoles(ctx context.Context, client *ent.Client) error {
	for _, b := range builtinRoles {
		exists, err := client.Role.Query().Where(role.CodeEQ(b.Code)).Exist(ctx)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if _, err := client.Role.Create().
			SetCode(b.Code).
			SetName(b.Name).
			SetDataScope(b.DataScope).
			SetIsSystem(true).
			SetIsActive(true).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

// SeedDeveloper 於 development 環境建立本機開發者帳號(design §4.4);production 不建立。
// 需既有 company 錨定(companyID);companyID<=0 時略過並回傳 true(表示略過)。
func SeedDeveloper(ctx context.Context, client *ent.Client, env string, companyID int) error {
	if env == "production" || companyID <= 0 {
		return nil
	}
	exists, err := client.User.Query().Where(user.EmailEQ("developer@local.dev")).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = client.User.Create().
		SetEmail("developer@local.dev").
		SetName("Local Developer").
		SetRole("developer").
		SetStatus(user.StatusActive).
		SetPasswordHash("!").
		SetCompanyID(companyID).
		Save(ctx)
	return err
}

// firstCompanyID 回傳任一家公司 ID;無公司時回傳 0(developer 錨點缺失略過)。
func firstCompanyID(ctx context.Context, client *ent.Client) int {
	c, err := client.Company.Query().First(ctx)
	if err != nil {
		return 0
	}
	return c.ID
}
