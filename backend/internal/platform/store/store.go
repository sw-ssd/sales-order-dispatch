// Package store 定義平台域的資料模型與兩個介面:唯讀的 Store(方案／權益／訂閱／例外)與
// 交易版的 BillingStore(期別／訂閱狀態／事件／稽核／settings 的寫入)。
//
// 平台域與業務域不 JOIN、不共用任何一張表,只以 company_id 對照租戶;存取一律走 admin
// 連線(DATABASE_ADMIN_URL):platform schema 對業務角色 app_rw 零權限(00029／S9)。
// Store **只讀**,供判定層與平台 CLI 使用;BillingStore 的每個寫入方法都接受 *sql.Tx ——
// 平台的一次寫入(期別／訂閱狀態／事件／稽核)必須是同一個 commit,交易由呼叫端擁有。
package store

import (
	"context"
	"database/sql"
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
	// ID 為訂閱列主鍵:期別(Period.SubscriptionID)與狀態變更(SetSubscriptionStatusTx)都以它
	// 為鍵,故 BillingStore 的取列與唯讀的 Store.Subscription 都會帶出它。
	ID        int64
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

// CreateSubscriptionInput 為建立訂閱(開通)所需的欄位。第一期不在此 —— 期別是呼叫端在
// 同一個交易內接著開的(它需要訂閱 id 與當期生效價)。
//
// Status 由呼叫端決定(trialing／active),store 不重複狀態機:哪些狀態可以用、怎麼轉移
// 是業務判定,寫在 SQL 裡就會有第二份答案。
type CreateSubscriptionInput struct {
	CompanyID    int
	PlanID       int64
	SeatCount    int
	BillingCycle string // monthly | yearly
	Status       string // trialing | active
	TrialEnds    *time.Time
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

// Period 為一期帳(platform.subscription_periods)。金額欄位一律是**分**(int64):DB 存
// numeric(12,2),兩端各自換算,中間不經 float64(帳務不接受 0.01 的誤差)。
//
// 價格欄位是簽約當下的**快照**(PlanID／UnitPriceCents／SeatPriceCents／SeatCount):方案日後
// 調價不得改動既有期別,否則調價等於回溯改帳。
type Period struct {
	ID              int64
	SubscriptionID  int64
	PeriodNo        int
	PeriodStart     time.Time
	PeriodEnd       time.Time
	PlanID          int64
	UnitPriceCents  int64
	SeatPriceCents  int64
	SeatCount       int
	AmountCents     int64
	Currency        string
	Status          string // open | paid | void
	PaidAt          *time.Time
	InvoiceNo       string
	PaymentProvider string
	ExternalRef     string
	// Note 為短收／溢收等人工註記(G8):付款時寫入,不改變期別金額;空字串代表未提供。
	Note string
}

// OpenPeriodInput 為建立一期所需的快照欄位。
// 不含 Note 與付款欄位:註記與付款憑據屬**付款事件**,只能經 MarkPeriodPaidTx 寫入。
type OpenPeriodInput struct {
	SubscriptionID int64
	PeriodNo       int
	PeriodStart    time.Time
	PeriodEnd      time.Time
	PlanID         int64
	UnitPriceCents int64
	SeatPriceCents int64
	SeatCount      int
	AmountCents    int64
	Currency       string
}

// Price 為「當期生效價」(GB1):同一方案同一計費週期可能多次調價,取 effective_from 最新者;
// 期別金額由它快照而來,方案調價後新期別用新價、舊期別不變。
type Price struct {
	BaseCents int64
	SeatCents int64
	Currency  string
}

// Event 為 outbox 事件(platform.events):跨域副作用(凍結公司、通知)由此驅動,consumer 以
// dispatched_at IS NULL 取件、處理成功後標記。
type Event struct {
	ID            int64
	AggregateType string
	AggregateID   int64
	EventType     string
	Payload       []byte
}

// BillingStore 為平台帳務的寫入介面(實作:postgres 的 admin 連線;FakeBilling 供單元測試)。
//
// **每個寫入方法都接受 *sql.Tx**:平台的一次寫入(期別／訂閱狀態／事件／稽核)必須落在同一個
// commit —— 收到款卻沒寫事件、或寫了事件卻沒改期別,都是無法回補的帳務事實。
//
// 交易由**呼叫端**擁有:store 不開交易也不 commit(唯一的例外是 WithTx,它是呼叫端宣告「這一段
// 要一個交易」的地方)。這些交易是**平台寫入的交易**,與租戶請求的 interceptor 交易無關:業務表
// 的交易帶 RLS 的 data_scope,平台表不套 RLS 也不經租戶連線(spec §3.3),兩者不可混用。
type BillingStore interface {
	// CreateSubscriptionTx 建立一筆訂閱(開通的第一步),回傳新列的 id。
	//
	// **同一公司已有未取消的訂閱時回 store.ErrConflict**(00029 的
	// subscriptions_active_company_unique)—— 不得讓 23505 冒上去變成 SYS-9000:那是
	// operator 的輸入情境(這家公司已經有合約),console 要顯示得出來、也才擋得住雙擊。
	// 已取消的訂閱不佔這條唯一鍵:要再服務是**新合約**(見 CancelSubscription)。
	CreateSubscriptionTx(ctx context.Context, tx *sql.Tx, in CreateSubscriptionInput) (int64, error)
	// OpenSubscriptionTx 取該租戶的現行訂閱並以 FOR UPDATE 鎖住該列(併發的收款／逾期轉移必須互斥)。
	// 無訂閱回 (nil, nil)。**不得**預先濾掉 cancelled:只有一筆已取消合約的租戶要拿得到它,
	// 判定層才分得出「已取消」與「從未訂閱」(F-8;取法與平台投影同源)。
	OpenSubscriptionTx(ctx context.Context, tx *sql.Tx, companyID int) (*Subscription, error)
	// SetSubscriptionStatusTx 更新訂閱狀態與寬限期(cancelled 時一併記下 cancelled_at)。
	SetSubscriptionStatusTx(ctx context.Context, tx *sql.Tx, subID int64, status string, graceUntil *time.Time) error
	// OpenPeriodTx 建立一期(status=open);unique(subscription_id, period_no) 使其可重跑:
	// 已存在即回既有期別,不新增也不覆寫。
	OpenPeriodTx(ctx context.Context, tx *sql.Tx, in OpenPeriodInput) (*Period, error)
	// OpenPeriodByNoTx 以 (subscription_id, period_no) 取期別;不存在時回 sql.ErrNoRows
	// (呼叫端只認 sql.ErrNoRows／ent.NotFoundError 代表「不存在」,其他錯誤必須中止)。
	OpenPeriodByNoTx(ctx context.Context, tx *sql.Tx, subID int64, periodNo int) (*Period, error)
	// CurrentPeriodTx 取訂閱最新一期(期別產生與逾期判定用);無期別回 (nil, nil)。
	CurrentPeriodTx(ctx context.Context, tx *sql.Tx, subID int64) (*Period, error)
	// CurrentPriceTx 取方案在指定計費週期的當期生效價(cycle: monthly | yearly,G1);
	// 沒有該週期的價目即錯誤(不得靜默用 0,那等於免費送方案)。
	CurrentPriceTx(ctx context.Context, tx *sql.Tx, planID int64, cycle string) (Price, error)
	// MarkPeriodPaidTx 標記期別已付款。note(G8)為短收／溢收的人工註記:空字串保留原值,
	// 不得用空字串清掉既有註記。同交易號重送視為重複入帳(no-op),交易號不同則拒絕覆蓋。
	//
	// invoiceNo／invoiceStatus／buyerTaxID／carrier 為**開票資訊**(spec §3 的
	// subscription_periods 欄位):期別是開票的對象,故與付款憑據同一組參數一起寫 —— 開票是
	// 付款事件的一部分,分開一支方法等於允許「收了錢但發票欄位沒落地」(T4 的缺口)。
	MarkPeriodPaidTx(ctx context.Context, tx *sql.Tx, id int64, paidAt time.Time,
		invoiceNo, invoiceStatus, buyerTaxID, carrier, provider, externalRef, note string) error
	// SetSeatCountTx 更新席位數(下一次產期即用新席位數計價;當期期別的金額快照不動)。
	SetSeatCountTx(ctx context.Context, tx *sql.Tx, subID int64, seats int) error
	// SetSubscriptionPlanTx 改訂閱的方案(下一期生效;當期期別的金額快照不動)。
	// 訂閱不存在時回 sql.ErrNoRows(不得靜默成功)。
	SetSubscriptionPlanTx(ctx context.Context, tx *sql.Tx, subID, planID int64) error
	// PlanIDByCodeTx 以 code 取方案 id;不存在**或已歸檔**回 sql.ErrNoRows ——
	// 已歸檔的方案不得再被指派(新合約只能掛在賣得動的方案上)。
	PlanIDByCodeTx(ctx context.Context, tx *sql.Tx, code string) (int64, error)
	// PeriodsByStatus 取指定狀態的期別(收款清單與對帳用),依 (subscription_id, period_no) 排序。
	PeriodsByStatus(ctx context.Context, status string) ([]Period, error)
	// ActiveSubscriptionsWithDueOpenPeriod 回 active、最新一期**仍是 open** 且已過期末者(轉 past_due)。
	// 最新一期已付款(is paid)或作廢(void)者不算逾期:逾期後才繳清的客戶不得被再次催收(C-1)。
	ActiveSubscriptionsWithDueOpenPeriod(ctx context.Context, tx *sql.Tx, now time.Time) ([]Subscription, error)
	// PastDueSubscriptionsExpiredGrace 回 past_due 且寬限期已過者(轉 suspended)。
	PastDueSubscriptionsExpiredGrace(ctx context.Context, tx *sql.Tx, now time.Time) ([]Subscription, error)
	// CancelledSubscriptionsPastPeriodEnd 回 cancelled 且最新期別已過期末、且尚未發過
	// subscription.expired 者(G7):排程可重跑而不重複發事件。
	CancelledSubscriptionsPastPeriodEnd(ctx context.Context, tx *sql.Tx, now time.Time) ([]Subscription, error)
	// ActiveOrTrialingSubscriptions 回仍在服務中的訂閱(排程逐租戶產生下一期用)。
	ActiveOrTrialingSubscriptions(ctx context.Context) ([]Subscription, error)
	// EmitEventTx 寫入 outbox 事件(與期別／狀態同一個交易)。
	EmitEventTx(ctx context.Context, tx *sql.Tx, aggregateType string, aggregateID int64, eventType string, payload []byte) error
	// UndispatchedEvents 取未派送事件(依 id 排序,consumer 用)。
	UndispatchedEvents(ctx context.Context, limit int) ([]Event, error)
	// MarkEventDispatchedTx 標記事件已派送(重試次數記在 attempts)。
	MarkEventDispatchedTx(ctx context.Context, tx *sql.Tx, id int64) error
	// RecordAuditTx 寫入平台稽核(S9:actor 為 operator_id,不 FK 租戶 users)。
	// reason 必填:空字串即拒絕 —— 動到錢與權限的操作必須留下「為什麼」。
	RecordAuditTx(ctx context.Context, tx *sql.Tx, operatorID int64, action, targetType, targetID, reason string, before, after []byte) error
	// WithTx 開一個交易並把 *sql.Tx 交給 fn;fn 回錯誤即回滾,否則提交。
	WithTx(ctx context.Context, fn func(*sql.Tx) error) error
	// SystemActor 讀 platform.settings.system_actor_user_id:排程／consumer 的稽核主體(G5)。
	SystemActor(ctx context.Context) (int64, error)
	// Setting 讀 platform.settings 的營運參數(試用／寬限／提前天數)。
	Setting(ctx context.Context, key string) (string, error)
	// UpsertSettingTx 寫入營運參數(供 console 維護;seed 亦寫同一張表)。
	UpsertSettingTx(ctx context.Context, tx *sql.Tx, key, value string) error
}
