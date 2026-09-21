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

// ErrConflict 表示**識別碼已存在**(目前只有 platform.operators.email 的唯一鍵)。
//
// 為什麼不讓唯一鍵的 23505 直接冒上來:它會被 toConnectError 收斂成 SYS-9000(5xx),而
// 「這個 email 已經在名單裡」是使用者輸入問題,console 必須顯示得出來。用 `ON CONFLICT DO
// NOTHING` + 0 列判定,比在錯誤字串裡撈 constraint 名可靠(且沒有競態)。
var ErrConflict = errors.New("平台資料已存在")

// ErrLastAdmin 表示「停用這位 operator 會讓平台**沒有任何 admin**」——治理上的不變式,不是
// 資料不存在。呼叫端(服務層)據此回 PLAT-3003 並附上可行動的原因,而不是含糊的「查無此人」。
var ErrLastAdmin = errors.New("不得停用最後一位 admin")

// ErrStatusChanged 表示 CAS 更新的 0 列是「列在、但狀態已被別的寫入者改走」
// （未結項 #12）—— 不是「列不在」（那是 sql.ErrNoRows）。呼叫端對前者跳過
// （不發事件、不計數），對後者報錯。
var ErrStatusChanged = errors.New("訂閱狀態已被其他寫入者變更")

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
	// 未結項 #27：試用到期／寬限期（nil = 無；trialing 才有前者、past_due 才有後者）。
	TrialEndsAt *time.Time
	GraceUntil  *time.Time
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
	// After 為稽核的 after 映像（未結項 #4：同一請求的多列共用 `_trace_id` 鍵，
	// console 據此合併顯示。空映像即無快照，不代表無 trace）。
	After []byte
}

// TenantOverrideInput 為新增一筆租戶例外(簽約承諾)的輸入。Enabled／Limit 為指標:
// nil 代表「這個維度不覆寫」—— 只看限額的例外不得讓功能被關掉。
//
// CreatedBy 是 platform.operators.id(表上的 NOT NULL 欄):例外是**有人承諾的**,沒有承諾者
// 就不該存在。Reason 由服務層擋空字串(表上 NOT NULL,但空字串是合法的 NOT NULL)。
type TenantOverrideInput struct {
	CompanyID   int64
	FeatureCode string
	Enabled     *bool
	Limit       *int64
	Reason      string
	Owner       string
	ExpiresAt   *time.Time
	CreatedBy   int64
}

// OverrideRef 為撤銷例外後回帶的識別:權益快取的失效需要 company_id,稽核需要 feature_code。
type OverrideRef struct {
	CompanyID   int64
	FeatureCode string
}

// ReceivableRow 為待收款清單的一列(公司 × 期別)。
//
// 金額是**字串**:期別金額在 DB 是 numeric(12,2),經過 float64 就可能在最後一位失真 ——
// 帳務不接受 0.01 的誤差(與 PlanPriceRow 同一個理由)。SQL 端以 (amount*100)::bigint 帶出,
// Go 端以 money.FormatCents 轉字串。
type ReceivableRow struct {
	CompanyID   string
	CompanyName string
	PlanCode    string
	PeriodNo    int32
	Amount      string
	PeriodEnd   time.Time
	Status      string
}

// PlanPriceInput 為寫入一筆方案價目的輸入。金額以**分**(int64)進出(不經 float64);
// 寫入時由 SQL 端經 money.FormatCents 的字串轉 numeric(12,2)。
//
// 調價是**新增一列**(plan_prices 是價格史,現行價取 effective_from 最新者):改寫既有列等於
// 回溯改帳 —— 舊期別的金額雖有快照,但「當時的價目」會消失,對帳時再也查不出來。
type PlanPriceInput struct {
	PlanCode     string
	BillingCycle string
	BaseCents    int64
	SeatCents    int64
	Currency     string
}
