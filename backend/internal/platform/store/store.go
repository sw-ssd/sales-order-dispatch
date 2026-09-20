// Package store 定義平台域(方案／權益／訂閱／例外)的**唯讀**存取介面與資料模型。
//
// 平台域與業務域不 JOIN、不共用任何一張表,只以 company_id 對照租戶;存取一律走 admin
// 連線(DATABASE_ADMIN_URL):platform schema 對業務角色 app_rw 零權限(00029／S9)。
// 本介面**只讀**(供判定層與平台 CLI 使用);寫入屬帳務 store,不在這裡。
package store

import (
	"context"
	"time"
)

// Feature 為可賣的功能／限額定義(platform.features);Type 決定 enabled 還是 limit 有意義。
type Feature struct {
	Code        string
	Type        string // boolean | integer
	Unit        string // 席 / 客戶 / 商品 / 部門 / GB（僅整數型有意義）
	Description string
}

// Entitlement 為方案預設權益(platform.plan_entitlements);Limit 為 nil 代表不限,
// 不是 0。
type Entitlement struct {
	FeatureCode string
	Enabled     bool
	Limit       *int64
}

// Override 為租戶例外(platform.tenant_overrides,即簽約承諾)。Enabled 與 Limit 可為
// nil:null 代表「這個維度不覆寫」,不是 false／0 —— 只看限額的例外不得讓功能被關掉。
//
// 「已撤銷」(revoked_at IS NOT NULL)的例外**不會**從 Store 出來;「已到期」的例外則照
// 樣出來,由判定層自行判斷(見 Store.Overrides)。
type Override struct {
	CompanyID   int
	FeatureCode string
	Enabled     *bool
	Limit       *int64
	ExpiresAt   *time.Time
}

// Subscription 為租戶當前的訂閱(platform.subscriptions)。取列方式與平台端投影
// (store/postgres/admin.go 的 GetTenant)**逐字相同**：優先未取消的，只有全部都是
// cancelled 時才取 cancelled —— 「已取消」與「從未訂閱」在判定層必須可區分：
// 前者 fail-closed 回 PLAT-3001，後者視為尚未開通計費、不施加限制。
// 公司完全沒有訂閱列時，Store.Subscription 回 (nil, nil) —— 「沒訂閱」不是錯誤。
type Subscription struct {
	CompanyID int
	PlanCode  string
	// PlanName 為方案名(JOIN platform.plans.name):租戶端投影要顯示它,不帶出來前端只能顯示 code。
	PlanName  string
	Status    string // trialing | active | past_due | suspended | cancelled
	PlanID    int64
	SeatCount int
	// BillingCycle 決定期別產生時「加一個月」或「加一年」(G1):年繳方案若漏帶此欄,
	// 每次只會產生一個月期別,少收 11 個月。
	BillingCycle string // monthly | yearly
	TrialEnds    *time.Time
	GraceUntil   *time.Time
}

// Store 為平台域的唯讀讀取介面。實作:postgres(admin 連線)、Fake(記憶體,供單元測試)。
type Store interface {
	// Features 回傳全部功能定義(一次載入,供型別判定與 UI 顯示)。
	Features(ctx context.Context) (map[string]Feature, error)
	// PlanEntitlements 回傳某方案的權益(方案以 code 指定)。
	PlanEntitlements(ctx context.Context, planCode string) ([]Entitlement, error)
	// Overrides 回傳某租戶「未撤銷」的例外(**不過濾到期**;到期與否由判定層負責)。
	Overrides(ctx context.Context, companyID int) ([]Override, error)
	// Subscription 回傳某租戶未取消的訂閱;無則回 (nil, nil)。
	Subscription(ctx context.Context, companyID int) (*Subscription, error)
}
