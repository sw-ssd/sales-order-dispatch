// 權益計數器：以業務 ent client 計算「有效筆數」（平台域不認得業務 schema，故由本域提供）。
package services

import (
	"context"
	"fmt"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/ent/department"
	"github.com/salesorder/sales-order-1.0/backend/ent/product"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
)

type entitlementCounter struct{ db *ent.Client }

// NewEntitlementCounter 建立計數器（於 server.InitDomains 注入 entitlements.Service）。
func NewEntitlementCounter(db *ent.Client) entitlements.Counter {
	return &entitlementCounter{db: db}
}

// Count 依 feature 計算有效用量（spec §3.2）：
//   - 席位 = 未停用的帳號數（users 無軟刪除欄位，停用即 status='inactive' → 釋放席位）；
//   - 其餘 = 該公司未軟刪除的列數（軟刪除不佔額度）。
//
// 一律經 dbtenant.Client(ctx, …) 取當前請求的 scoped client：這四張表都是 RLS
// ENABLE+FORCE，只有請求交易（scope 已 SET LOCAL）看得到列，直接用建構子的 db 會在
// 請求路徑回 0（形同無限額）。非請求路徑（CLI／單元測試）落回 fallback client ——
// 與其他服務同一慣例；要跨租戶計數的 CLI 須自行開 SystemScopeTx。
func (c *entitlementCounter) Count(ctx context.Context, companyID int, feature string) (int, error) {
	db := dbtenant.Client(ctx, c.db)
	switch feature {
	case entitlements.LimitSeats:
		return db.User.Query().
			Where(user.HasCompanyWith(company.ID(companyID)), user.StatusNEQ(user.StatusInactive)).
			Count(ctx)
	case entitlements.LimitCustomers:
		return db.Customer.Query().
			Where(customer.CompanyIDEQ(companyID), customer.DeletedAtIsNil()).
			Count(ctx)
	case entitlements.LimitProducts:
		return db.Product.Query().
			Where(product.CompanyIDEQ(companyID), product.DeletedAtIsNil()).
			Count(ctx)
	case entitlements.LimitDepartments:
		return db.Department.Query().
			Where(department.HasCompanyWith(company.ID(companyID)), department.DeletedAtIsNil()).
			Count(ctx)
	default:
		// 不得回 0：0 等於「用量為零」，判定層會誤判為未超額而放行。
		return 0, fmt.Errorf("未定義的計數 feature: %s", feature)
	}
}
