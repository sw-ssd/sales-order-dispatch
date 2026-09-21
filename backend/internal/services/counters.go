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
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
)

type entitlementCounter struct{ db *ent.Client }

// NewEntitlementCounter 建立計數器（於 server.InitDomains 注入 entitlements.Service）。
func NewEntitlementCounter(db *ent.Client) entitlements.Counter {
	return &entitlementCounter{db: db}
}

// Count 依 feature 計算「該公司」的有效用量（spec §3.2）：
//   - 席位 = 未停用的帳號數（users 無軟刪除欄位，停用即 status='inactive' → 釋放席位）；
//   - 其餘 = 該公司未軟刪除的列數（軟刪除不佔額度）。
//
// 未結項 #1(Plan B)：席位口徑含客戶帳號 —— CreateCustomer 建兩列 users（主帳＋子帳，
// customer_service.go），而席位計數含所有非 inactive 帳號（is_customer 未排除）。
// 該路徑只受 LimitCustomers 守衛 → 已達席位上限仍可藉「建客戶」超額佔席位。
// 選項：(a) 席位只算非客戶帳號（is_customer=false，需改 spec §3.2＋此處；建議）；
// (b) CreateCustomer 同時掛 LimitSeats 守衛。改之前不得動此計數（牽一髮：守衛入口三處）。
//
// 可見範圍（關鍵）：配額是**公司層**的（spec §3.2 的 WHERE company_id=?），但 RLS 會把可見範圍
// 縮到請求 scope。故依 auth.RLSFrom(ctx).DataScope 分流：
//
//   - company／all：請求交易本身即可見整個公司（all 亦靠明示的 company_id 條件），直接在請求
//     交易內計數。請求交易看得到同請求尚未提交的列（守衛在寫入前呼叫 → 同請求內連續建立時
//     前一筆也算得到）。
//   - department／self（dept_admin／staff 的請求）：**改走獨立的系統範圍交易**
//     （dbtenant.SystemScopeTx，scope=all）在公司層計數。理由：這兩種 scope 在請求交易內只看得到
//     本部門／本人的列，直接數會把配額低報成「每部門一份」，而 dept_admin／staff 可以建立
//     客戶／商品／使用者 → 掛上守衛後就是**超額放行（fail-open）**。
//   - **空 scope（＝無租戶身分）**：同樣走系統範圍交易。這是 00028 **ENABLE＋FORCE RLS** 的
//     直接後果：未設 scope 的請求交易（或完全沒有請求交易 —— OIDC 首次登入、CLI）把 users 等表
//     濾成 **0 列** → `used=0` → 任何上限都不觸發，守衛成了**裝飾品**。無身分 ≠ 沒有限制：
//     公司由呼叫端明示帶入（services.GuardQuotaForCompany），計數就必須看得見整個公司。
//     （sqlite 的 enttest 沒有 RLS，這條路徑在單元層「怎麼數都對」——真容器整合測試才量得出來。）
//
// SystemScopeTx 的代價（刻意接受）：①它是第二條連線（AGENTS §9-6），但只做唯讀 SELECT、
// 不取任何列鎖，不會與請求交易互鎖；②它看不到請求交易內未提交的列 —— 守衛一律在寫入之前呼叫，
// 該情境不存在，且「低報」屬 fail-closed 方向，遠比超額放行安全。
//
// 未特別處理 DataScopeAll：平台層身分（super／developer）的配額略過寫在守衛入口
// （services.guardQuota），判定層只認 scope 的可見性。
func (c *entitlementCounter) Count(ctx context.Context, companyID int, feature string) (int, error) {
	switch auth.RLSFrom(ctx).DataScope {
	case auth.DataScopeCompany, auth.DataScopeAll:
		return countFeature(ctx, dbtenant.Client(ctx, c.db), companyID, feature)
	default:
		var n int
		err := dbtenant.SystemScopeTx(ctx, c.db, func(tx *ent.Tx) error {
			var err error
			n, err = countFeature(ctx, tx.Client(), companyID, feature)
			return err
		})
		return n, err
	}
}

// countFeature 在指定 client 上依 feature 計數。company_id 條件一律明示（不靠 RLS 兜底），
// 這樣在 scope=all 的系統範圍交易內也只數該公司。
//
// **一律沿用呼叫端的 ctx**（不得用 context.Background()）：department／self scope 的守衛會經
// SystemScopeTx 開第二條連線，交易若綁在 background ctx 上就不隨請求取消／逾時結束 ——
// 客戶端早已放棄，慢查詢仍握著請求交易與一條額外連線（F-4）。ctx 上的值不影響本查詢：
// 可見範圍由**傳入的 client** 決定，不是 ctx。
func countFeature(ctx context.Context, db *ent.Client, companyID int, feature string) (int, error) {
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
