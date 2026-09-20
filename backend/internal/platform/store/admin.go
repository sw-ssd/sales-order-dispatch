// 平台營運工具(operator console)的投影 DTO:跨租戶視圖與稽核軌跡的查詢結果。
//
// 與 store.go 的權益判定 DTO 分開的理由:那一組服務「判定」(方案／權益／訂閱／例外),
// 這一組服務「顯示與稽核」(租戶列表、方案價目、稽核軌跡)。兩者都只走 admin 連線
// (platform schema 對業務角色 app_rw 零權限,00029/S9)、都不經 ent、都不套 RLS。
package store

import (
	"errors"
	"time"
)

// ErrNotFound 表示查詢對象不存在(查無此租戶、查無此方案)。
//
// 為什麼不和租戶端一樣合併「不存在」與「無權存取」:那裡合併是為了不讓租戶以錯誤碼探測
// 他租戶資源是否存在(oracle);平台工具只有 operator 進得來(T8 的 operatorauth),沒有
// 可探測的第三方,合併只會讓 console 顯示「系統忙碌」而不是「查無此租戶」。服務層一律
// 映射為 SYS-4002。
var ErrNotFound = errors.New("平台資料不存在")

// TenantRow 為租戶列表／詳情的一列(companies × platform.subscriptions × plans 的投影)。
//
// 欄位橫跨業務表與平台表:spec §6.4 明訂平台方的跨租戶視圖走「admin 連線 ＋ 投影查詢」,
// 不開 super 的 data_scope=all 那條路。這是**唯讀投影**,不建立跨域 FK 或依賴。
type TenantRow struct {
	CompanyID        string
	CompanyName      string
	PlanCode         string
	PlanName         string
	Status           string // 訂閱狀態:trialing | active | past_due | suspended | cancelled | none
	SeatCount        int32
	CurrentPeriodEnd *time.Time // nil = 尚無期別(proto 的空字串)
	Overdue          bool       // 有已過期未付的期別(待收款清單的定義)
}

// TenantOverrideRow 為 platform.tenant_overrides 的一列。已撤銷(revoked_at)者不回傳;
// 已到期的**照樣**回傳,由呼叫端判斷 —— 與 Store.Overrides 同語意(到期是判定層的事,
// store 不替它決定)。
//
// Enabled／Limit 用指標:NULL 代表「這個維度不覆寫」,不是 false／0 —— 只看限額的例外
// 不得讓功能被關掉。
type TenantOverrideRow struct {
	ID          string
	FeatureCode string
	Enabled     *bool
	Limit       *int64
	Reason      string
	Owner       string // 承諾者(平台側人員)
	ExpiresAt   *time.Time
}

// PlanRow 為 platform.plans 的一列,含各計費週期的**現行**價目。
type PlanRow struct {
	ID        string
	Code      string
	Name      string
	Status    string // active | archived
	SortOrder int32
	Prices    []PlanPriceRow
}

// PlanPriceRow 為某方案某計費週期的現行價目(同一週期多次調價 → 取 effective_from 最新者,
// 調價史只在資料庫,調價前的舊約期別另有快照欄位)。
//
// 金額是**字串**:proto 的欄位也是字串,而 numeric(12,2) 一旦經過 float64 就可能在最後一位
// 失真——帳務不接受「誤差 0.01」。SQL 端以 ::text 帶出,不經 Go 的數值轉換。
type PlanPriceRow struct {
	BillingCycle  string
	BasePrice     string
	SeatPrice     string
	Currency      string
	EffectiveFrom time.Time
}

// PlatformAuditRow 為 platform.audit_logs 的一列(含操作者 email)。
//
// 稽核要能回答「誰做的、對誰、為什麼」,故帶 email:只回 operator_id 的話,停用該 operator
// 之後(白名單列仍在,但沒有 UI 查得到名字)這筆稽核就讀不出人了。
type PlatformAuditRow struct {
	ID            string
	OperatorEmail string
	Action        string
	TargetType    string
	TargetID      string
	Reason        string
	CreatedAt     time.Time
}
