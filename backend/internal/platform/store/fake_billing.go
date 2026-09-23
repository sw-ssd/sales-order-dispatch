package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FakeBilling 為 BillingStore 的記憶體實作:供帳務(收款)、排程與 outbox consumer 的**單元測試**
// 使用(不進 production 路徑),讓那些測試不必起容器。Fake(唯讀判定層)與它分開:判定層讀權益,
// 帳務讀寫訂閱與期別,兩者的資料來源不同,混成一個物件只會讓種子資料有兩份。
//
// 與 SQL 實作**語意一致**的地方(不一致會讓單元測試失真):
//   - 期別不存在 → sql.ErrNoRows(呼叫端只認 sql.ErrNoRows／ent.NotFoundError 代表「不存在」,
//     其他錯誤必須中止);
//   - OpenSubscriptionTx **不預先濾掉 cancelled**(F-8/C-02):優先未取消,只有全是 cancelled
//     時才取它;完全沒有列回 (nil, nil);
//   - CreateSubscriptionTx:同一公司已有未取消的訂閱 → ErrConflict(部分唯一索引
//     subscriptions_active_company_unique;已取消者不佔這條鍵);
//   - OpenPeriodTx 以 (subscription_id, period_no) 為冪等鍵:已存在即回既有期別,不新增不覆寫;
//   - MarkPeriodPaidTx:同交易號重送是完全 no-op(既有付款憑據一個都不動)、交易號不同即拒絕、
//     **同一 provider ＋ 交易號不得入帳兩期**(00029 的 periods_provider_ref_unique)、
//     note 為空字串時保留原值、狀態為 void 等一律拒絕並說出狀態;
//   - SetSeatCountTx／SetSubscriptionPlanTx:訂閱不存在 → sql.ErrNoRows(不得靜默成功);
//     PlanIDByCodeTx 對「已歸檔」視同不存在(假實作只認 PutPlan 註冊過的方案);
//   - EmitEventTx:空 payload 存成 '{}'(SQL 的 jsonb 欄位不接受空字串);
//   - RecordAuditTx 的 reason 必填(空字串即拒絕);
//   - 排程三個查詢的集合邊界(最新一期**仍是 open** 且期末 < now、grace_until IS NOT NULL、
//     已發過 subscription.expired 者不再選中)逐一照 SQL 的謂詞寫。
//
// 交易語意:記憶體裡沒有真的交易,**tx 參數一律忽略**;唯一表達「同一交易」的地方是 WithTx ——
// 它進入時整份快照、fn 回錯誤時整份還原,讓「失敗不留半成品」的斷言在單元測試裡也成立。因此
// 假實作**不會**擋下「沒開交易就直接寫」的誤用,那條界線由整合測試(真 store ＋ 真 PostgreSQL)
// 把關;列鎖(FOR UPDATE)同理,記憶體裡不存在。
type FakeBilling struct {
	mu sync.Mutex

	nextSubID    int64
	nextPeriodID int64
	nextEventID  int64

	subs    []Subscription
	periods []Period
	events  []Event
	// dispatched 為已派送事件的 id 集合(platform.events.dispatched_at IS NOT NULL)。
	dispatched map[int64]bool
	prices     map[priceKey]Price
	audits     []AuditRecord
	settings   map[string]string
	// plans 為 code → 方案 id(**只註冊 active 的方案**):PlanIDByCodeTx 對已歸檔者視同不存在,
	// 而「歸檔」是 SQL 的 WHERE 條件,假實作不需要另一份狀態 —— 不註冊就是不存在。
	plans map[string]int64
	// platformCompanies 為平台自營公司的公司 id 集合（G5）：OverdueReceivablePeriods
	// 排除它們（真 store 在 SQL 內 JOIN companies 以 identifier='platform' 排除；
	// 假實作沒有 companies 表，故由 PutPlatformCompany 標記）。
	platformCompanies map[int]bool
}

// priceKey 為價目的鍵(方案 × 計費週期);SQL 端另有 effective_from 的生效順序,假實作只保留
// 「當期生效價」一筆 —— 單元測試要驗的是「取到哪一筆價」,價目史屬遷移與真 store 的範疇。
type priceKey struct {
	planID int64
	cycle  string
}

// AuditRecord 為 FakeBilling 記下的平台稽核。介面沒有讀稽核的方法(正式路徑由 console 查
// platform.audit_logs),單元測試要斷言「動了錢有沒有留痕」得看得到它。
type AuditRecord struct {
	OperatorID                           int64
	Action, TargetType, TargetID, Reason string
	Before, After                        []byte
}

var _ BillingStore = (*FakeBilling)(nil)

// NewFakeBilling 建立空的記憶體帳務 store(供單元測試與 CLI)。
func NewFakeBilling() *FakeBilling {
	return &FakeBilling{
		dispatched:        map[int64]bool{},
		prices:            map[priceKey]Price{},
		settings:          map[string]string{},
		plans:             map[string]int64{},
		platformCompanies: map[int]bool{},
	}
}

// PutSubscription 種一列訂閱(ID 為 0 時自動配),回傳它的 ID。
func (f *FakeBilling) PutSubscription(s Subscription) int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s.ID == 0 {
		f.nextSubID++
		s.ID = f.nextSubID
	} else if s.ID > f.nextSubID {
		f.nextSubID = s.ID
	}
	f.subs = append(f.subs, cloneSubscription(s))
	return s.ID
}

// PutPeriod 種一期(ID 為 0 時自動配),回傳它的 ID —— 排程與 consumer 的測試需要「已經有一期」
// 的起點(上一期未付、期末已過…)。
func (f *FakeBilling) PutPeriod(p Period) int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	if p.ID == 0 {
		f.nextPeriodID++
		p.ID = f.nextPeriodID
	} else if p.ID > f.nextPeriodID {
		f.nextPeriodID = p.ID
	}
	if p.Status == "" {
		p.Status = "open"
	}
	f.periods = append(f.periods, clonePeriod(p))
	return p.ID
}

// PutPlanPrice 設定某方案某計費週期的當期生效價(CurrentPriceTx 的來源)。
func (f *FakeBilling) PutPlanPrice(planID int64, cycle string, p Price) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.prices[priceKey{planID: planID, cycle: cycle}] = p
}

// PutPlan 註冊一個可被指派的方案(code → id)。未註冊者(含已歸檔的方案)在 PlanIDByCodeTx
// 一律視同不存在 —— 與 SQL 的 `status = 'active'` 條件同語意。
func (f *FakeBilling) PutPlan(code string, id int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.plans[code] = id
}

// PutSetting 種一筆營運參數(含 system_actor_user_id)。
func (f *FakeBilling) PutSetting(key, value string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.settings[key] = value
}

// Audits 回傳已寫入的稽核(副本)。
func (f *FakeBilling) Audits() []AuditRecord {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]AuditRecord, len(f.audits))
	for i, a := range f.audits {
		a.Before = slices.Clone(a.Before)
		a.After = slices.Clone(a.After)
		out[i] = a
	}
	return out
}

// Events 回傳**全部**已寫入的事件,含已派送者(副本);未派送者請用 UndispatchedEvents。
func (f *FakeBilling) Events() []Event {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]Event, len(f.events))
	for i, e := range f.events {
		e.Payload = slices.Clone(e.Payload)
		out[i] = e
	}
	return out
}

// CreateSubscriptionTx 建立一筆訂閱。同一公司已有未取消的訂閱 → ErrConflict
// (00029 的 subscriptions_active_company_unique:已取消的訂閱不佔這條鍵)。
func (f *FakeBilling) CreateSubscriptionTx(_ context.Context, _ *sql.Tx,
	in CreateSubscriptionInput) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range f.subs {
		if f.subs[i].CompanyID == in.CompanyID && f.subs[i].Status != "cancelled" {
			return 0, ErrConflict
		}
	}
	f.nextSubID++
	f.subs = append(f.subs, Subscription{
		ID:           f.nextSubID,
		CompanyID:    in.CompanyID,
		PlanID:       in.PlanID,
		SeatCount:    in.SeatCount,
		BillingCycle: in.BillingCycle,
		Status:       in.Status,
		TrialEnds:    clonePtr(in.TrialEnds),
	})
	return f.nextSubID, nil
}

// OpenSubscriptionTx 取該公司的現行訂閱(見型別說明:不預先濾掉 cancelled)。
func (f *FakeBilling) OpenSubscriptionTx(_ context.Context, _ *sql.Tx, companyID int) (*Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var best *Subscription
	for i := range f.subs {
		s := &f.subs[i]
		if s.CompanyID != companyID {
			continue
		}
		// 與 SQL 的 `ORDER BY (status='cancelled'), id DESC LIMIT 1` 同序:未取消優先,
		// 都取消時取 id 大者(最新一筆)。
		if best == nil || betterSubscription(*s, *best) {
			best = s
		}
	}
	if best == nil {
		return nil, nil
	}
	out := cloneSubscription(*best)
	return &out, nil
}

// betterSubscription 回報 a 是否比 b 更該被選為「現行訂閱」。
func betterSubscription(a, b Subscription) bool {
	ac, bc := a.Status == "cancelled", b.Status == "cancelled"
	if ac != bc {
		return !ac
	}
	return a.ID > b.ID
}

// SetSubscriptionStatusTx 更新訂閱狀態與寬限期(未取消時寬限期為 nil，代表清空)。
// CAS 語意與真 store 同：expectedStatus 非空且現狀不符時回 ErrStatusChanged（不改狀態）。
func (f *FakeBilling) SetSubscriptionStatusTx(_ context.Context, _ *sql.Tx, subID int64, status string, graceUntil *time.Time, expectedStatus ...string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	i := f.subIndex(subID)
	if i < 0 {
		return fmt.Errorf("訂閱 %d 不存在: %w", subID, sql.ErrNoRows)
	}
	if len(expectedStatus) > 0 && expectedStatus[0] != "" && f.subs[i].Status != expectedStatus[0] {
		return fmt.Errorf("訂閱 %d 狀態已變更(預期 %q): %w", subID, expectedStatus[0], ErrStatusChanged)
	}
	f.subs[i].Status = status
	f.subs[i].GraceUntil = clonePtr(graceUntil)
	// 與真 store 的 cancelled_at CASE 同語意：首次取消記時間，重複取消不推進。
	if status == "cancelled" && f.subs[i].CancelledAt == nil {
		now := time.Now()
		f.subs[i].CancelledAt = &now
	}
	return nil
}

// OpenPeriodTx 建立一期;已存在同 (subscription_id, period_no) 即回既有期別(不新增不覆寫)。
func (f *FakeBilling) OpenPeriodTx(_ context.Context, _ *sql.Tx, in OpenPeriodInput) (*Period, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if existing, ok := f.periodByNo(in.SubscriptionID, in.PeriodNo); ok {
		out := clonePeriod(existing)
		return &out, nil
	}
	f.nextPeriodID++
	p := clonePeriod(Period{
		ID:             f.nextPeriodID,
		SubscriptionID: in.SubscriptionID,
		PeriodNo:       in.PeriodNo,
		PeriodStart:    in.PeriodStart,
		PeriodEnd:      in.PeriodEnd,
		PlanID:         in.PlanID,
		UnitPriceCents: in.UnitPriceCents,
		SeatPriceCents: in.SeatPriceCents,
		SeatCount:      in.SeatCount,
		AmountCents:    in.AmountCents,
		Currency:       in.Currency,
		Status:         "open",
	})
	f.periods = append(f.periods, p)
	return &p, nil
}

// OpenPeriodByNoTx 以 (subscription_id, period_no) 取期別;不存在時回 sql.ErrNoRows。
func (f *FakeBilling) OpenPeriodByNoTx(_ context.Context, _ *sql.Tx, subID int64, periodNo int) (*Period, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.periodByNo(subID, periodNo)
	if !ok {
		return nil, sql.ErrNoRows
	}
	out := clonePeriod(p)
	return &out, nil
}

// CurrentPeriodTx 取訂閱最新一期(period_no 最大者);無期別回 (nil, nil)。
func (f *FakeBilling) CurrentPeriodTx(_ context.Context, _ *sql.Tx, subID int64) (*Period, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var best *Period
	for i := range f.periods {
		p := &f.periods[i]
		if p.SubscriptionID != subID {
			continue
		}
		if best == nil || p.PeriodNo > best.PeriodNo {
			best = p
		}
	}
	if best == nil {
		return nil, nil
	}
	out := clonePeriod(*best)
	return &out, nil
}

// CurrentPriceTx 取該方案在指定計費週期的當期生效價;沒有價目即錯誤(不得靜默回 0 元)。
func (f *FakeBilling) CurrentPriceTx(_ context.Context, _ *sql.Tx, planID int64, cycle string) (Price, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.prices[priceKey{planID: planID, cycle: cycle}]
	if !ok {
		// 與 SQL 同語意:沒有價目回 store.ErrNotFound(「查價失敗」才回其他錯誤)。
		return Price{}, fmt.Errorf("方案 %d 沒有 %s 週期的價目: %w", planID, cycle, ErrNotFound)
	}
	return p, nil
}

// MarkPeriodPaidTx 標記期別已付款(見型別說明:同交易號重送 no-op 且不動既有憑據、不同交易號
// 拒絕、同一交易號不得入帳兩期、note 空字串保留原值、其他狀態一律拒絕)。
func (f *FakeBilling) MarkPeriodPaidTx(_ context.Context, _ *sql.Tx, id int64, paidAt time.Time,
	invoiceNo, invoiceStatus, buyerTaxID, carrier, provider, externalRef, note string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	i := f.periodIndex(id)
	if i < 0 {
		return fmt.Errorf("期別 %d 不存在: %w", id, sql.ErrNoRows)
	}
	p := &f.periods[i]
	// 00029 的 periods_provider_ref_unique(部分唯一索引):同一 provider ＋ 交易號只能入帳一期。
	// 少了這一條,T4 的「同一筆交易號不得重複入帳」會在假實作上假綠。
	// 未結項 #8（已關一半）：跨期衝突的「呼叫端可觀察行為」已由三層測試釘住 ——
	// fake 層（fake_billing_test：跨期拒絕＋期別不動）、真 SQL 層（billing_integration_test：
	// 唯一索引拒絕＋兩期皆不動）、billing 層（billing_test：PLAT-3002＋事件／稽核不寫）。
	// 真 store 靠 UPDATE 的 0 列分流（狀態／交易號不合即 0 列，再查狀態決定三種錯誤），
	// 假實作用預檢＋格式化字串 —— 形狀不同但呼叫端只認 errors.Is(err, sql.ErrNoRows)＝
	// 「不存在」，其餘一律收斂，故不影響行為。殘：`cancelled_at` 無法由 fake 表達
	// （fake 的 Subscription 無該欄；取消時間語意只由整合測試覆蓋）。
	if externalRef != "" && p.ExternalRef != externalRef {
		for j := range f.periods {
			if j != i && f.periods[j].ExternalRef == externalRef && f.periods[j].PaymentProvider == provider {
				return fmt.Errorf("交易號 %q（provider %q）已入帳於期別 %d，不得重複入帳",
					externalRef, provider, f.periods[j].ID)
			}
		}
	}
	switch {
	case p.Status == "open":
		p.Status = "paid"
		p.PaidAt = &paidAt
		p.InvoiceNo = invoiceNo
		p.PaymentProvider = provider
		p.ExternalRef = externalRef
	case p.Status == "paid" && p.ExternalRef == externalRef:
		// 重複入帳(webhook 重播／呼叫端重試):完全 no-op,已入帳的憑據一個都不動
		// (SQL 端同樣只在 status='open' 時寫那四個欄位)。
	case p.Status == "paid":
		return fmt.Errorf("期別 %d 已付款（交易號 %q）且交易號不同（%q），拒絕覆蓋",
			id, p.ExternalRef, externalRef)
	default:
		// void 等狀態:說出真正的狀態,不得講成「已付款」。
		return fmt.Errorf("期別 %d 狀態為 %q，不得入帳", id, p.Status)
	}
	if note != "" {
		p.Note = note
	}
	return nil
}

// SetSeatCountTx 更新席位數;訂閱不存在回 sql.ErrNoRows(與 SQL 同語意)。
func (f *FakeBilling) SetSeatCountTx(_ context.Context, _ *sql.Tx, subID int64, seats int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	i := f.subIndex(subID)
	if i < 0 {
		return fmt.Errorf("訂閱 %d 不存在: %w", subID, sql.ErrNoRows)
	}
	f.subs[i].SeatCount = seats
	return nil
}

// SetSubscriptionPlanTx 改訂閱的方案;訂閱不存在回 sql.ErrNoRows(與 SQL 同語意)。
func (f *FakeBilling) SetSubscriptionPlanTx(_ context.Context, _ *sql.Tx, subID, planID int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	i := f.subIndex(subID)
	if i < 0 {
		return fmt.Errorf("訂閱 %d 不存在: %w", subID, sql.ErrNoRows)
	}
	f.subs[i].PlanID = planID
	return nil
}

// PlanIDByCodeTx 以 code 取方案 id;未註冊(不存在／已歸檔)回 sql.ErrNoRows。
func (f *FakeBilling) PlanIDByCodeTx(_ context.Context, _ *sql.Tx, code string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id, ok := f.plans[code]
	if !ok {
		return 0, fmt.Errorf("方案 %q 不存在或已歸檔: %w", code, sql.ErrNoRows)
	}
	return id, nil
}

// PeriodsByStatus 取指定狀態的期別,依 (subscription_id, period_no) 排序(與 SQL 同序)。
func (f *FakeBilling) PeriodsByStatus(_ context.Context, status string) ([]Period, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []Period
	for i := range f.periods {
		if f.periods[i].Status == status {
			out = append(out, clonePeriod(f.periods[i]))
		}
	}
	slices.SortFunc(out, func(a, b Period) int {
		if a.SubscriptionID != b.SubscriptionID {
			return cmpInt64(a.SubscriptionID, b.SubscriptionID)
		}
		return a.PeriodNo - b.PeriodNo
	})
	return out, nil
}

// OverdueReceivablePeriods 回待收款期別:open 且期末已過，且排除平台自營公司。
// 假實作沒有 companies 表，故平台自營公司由 PutPlatformCompany 標記
// （真 store 在 SQL 內 JOIN companies 以 identifier='platform' 排除）。
func (f *FakeBilling) OverdueReceivablePeriods(_ context.Context, now time.Time) ([]Period, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	subCompany := map[int64]int{}
	for i := range f.subs {
		subCompany[f.subs[i].ID] = f.subs[i].CompanyID
	}
	var out []Period
	for i := range f.periods {
		p := f.periods[i]
		if p.Status != "open" || !p.PeriodEnd.Before(now) {
			continue
		}
		if f.platformCompanies[subCompany[p.SubscriptionID]] {
			continue
		}
		out = append(out, clonePeriod(p))
	}
	return out, nil
}

// PutPlatformCompany 標記某公司為平台自營公司（G5）：OverdueReceivablePeriods 排除它。
func (f *FakeBilling) PutPlatformCompany(companyID int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.platformCompanies == nil {
		f.platformCompanies = map[int]bool{}
	}
	f.platformCompanies[companyID] = true
}

// ActiveSubscriptionsWithDueOpenPeriod:active 且最新一期**仍是 open** 且已過期末
// (SQL 的 `cur.status = 'open' AND cur.period_end < now`);已付款／已作廢的期別不算逾期(C-1)。
func (f *FakeBilling) ActiveSubscriptionsWithDueOpenPeriod(_ context.Context, _ *sql.Tx, now time.Time) ([]Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.filterSubs(func(s Subscription) bool {
		if s.Status != "active" {
			return false
		}
		cur, ok := f.latestPeriod(s.ID)
		return ok && cur.Status == "open" && cur.PeriodEnd.Before(now)
	}), nil
}

// PastDueSubscriptionsExpiredGrace:past_due 且寬限期已過;grace_until 為 nil **不算**已過
// (與 SQL 的 `grace_until IS NOT NULL` 一致)。
func (f *FakeBilling) PastDueSubscriptionsExpiredGrace(_ context.Context, _ *sql.Tx, now time.Time) ([]Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.filterSubs(func(s Subscription) bool {
		return s.Status == "past_due" && s.GraceUntil != nil && s.GraceUntil.Before(now)
	}), nil
}

// TrialingSubscriptionsExpiredTrial:trialing 且 trial_ends_at 已過(SQL 的
// `trial_ends_at IS NOT NULL AND trial_ends_at < now`;無到期日不算到期)。
func (f *FakeBilling) TrialingSubscriptionsExpiredTrial(_ context.Context, _ *sql.Tx, now time.Time) ([]Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.filterSubs(func(s Subscription) bool {
		return s.Status == "trialing" && s.TrialEnds != nil && s.TrialEnds.Before(now)
	}), nil
}

// TrialingSubscriptionsWithoutTrialEnd:trialing 但沒有到期日（未結項 #40）。
// ExpireTrials 刻意不碰它們，故它們永遠停在 trialing —— 這裡只列出來，不轉移。
func (f *FakeBilling) TrialingSubscriptionsWithoutTrialEnd(_ context.Context) ([]Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.filterSubs(func(s Subscription) bool {
		return s.Status == "trialing" && s.TrialEnds == nil
	}), nil
}

// CancelledSubscriptionsPastPeriodEnd:cancelled 且最新一期已過期末,**且**尚未發過
// subscription.expired(排程可重跑而不重複發事件)。
func (f *FakeBilling) CancelledSubscriptionsPastPeriodEnd(_ context.Context, _ *sql.Tx, now time.Time) ([]Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.filterSubs(func(s Subscription) bool {
		if s.Status != "cancelled" || f.expiredEmitted(s.ID) {
			return false
		}
		cur, ok := f.latestPeriod(s.ID)
		return ok && cur.PeriodEnd.Before(now)
	}), nil
}

// ActiveOrTrialingSubscriptions:仍在服務中的訂閱(依 id 排序,與 SQL 同序)。
func (f *FakeBilling) ActiveOrTrialingSubscriptions(context.Context) ([]Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := f.filterSubs(func(s Subscription) bool {
		return s.Status == "active" || s.Status == "trialing"
	})
	slices.SortFunc(out, func(a, b Subscription) int { return cmpInt64(a.ID, b.ID) })
	return out, nil
}

// EmitEventTx 寫入 outbox 事件(未派送)。空 payload(nil／空切片)存成 '{}':真 store 的
// jsonb 欄位不接受空字串,不這樣對齊的話 consumer 的單元測試拿到的 payload 會與真環境不同。
func (f *FakeBilling) EmitEventTx(_ context.Context, _ *sql.Tx, aggregateType string,
	aggregateID int64, eventType string, payload []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	f.nextEventID++
	f.events = append(f.events, Event{
		ID:            f.nextEventID,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Payload:       slices.Clone(payload),
	})
	return nil
}

// UndispatchedEvents 取未派送事件(依 id 排序,同 SQL)。
func (f *FakeBilling) UndispatchedEvents(_ context.Context, limit int) ([]Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []Event
	for i := range f.events {
		e := f.events[i]
		if f.dispatched[e.ID] {
			continue
		}
		e.Payload = slices.Clone(e.Payload)
		out = append(out, e)
		if limit > 0 && len(out) == limit {
			break
		}
	}
	return out, nil
}

// MarkEventDispatchedTx 標記事件已派送(找不到該 id 不報錯,與 UPDATE 0 列同語意)。
func (f *FakeBilling) MarkEventDispatchedTx(_ context.Context, _ *sql.Tx, id int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dispatched[id] = true
	return nil
}

// RecordAuditTx 寫入平台稽核;reason 必填(空字串與全空白即拒絕,不寫半筆)。
func (f *FakeBilling) RecordAuditTx(_ context.Context, _ *sql.Tx, operatorID int64,
	action, targetType, targetID, reason string, before, after []byte) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("平台稽核必須提供原因")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.audits = append(f.audits, AuditRecord{
		OperatorID: operatorID, Action: action, TargetType: targetType, TargetID: targetID,
		Reason: reason, Before: slices.Clone(before), After: slices.Clone(after),
	})
	return nil
}

// WithTx 執行 fn 並在它回錯誤時**整份還原**(見型別說明:記憶體沒有真的交易,這是唯一表達
// 「失敗不留半成品」的地方)。
func (f *FakeBilling) WithTx(_ context.Context, fn func(*sql.Tx) error) error {
	f.mu.Lock()
	saved := f.snapshot()
	f.mu.Unlock()

	if err := fn(nil); err != nil {
		f.mu.Lock()
		f.restore(saved)
		f.mu.Unlock()
		return err
	}
	return nil
}

// SystemActor 讀 system_actor_user_id(與 SQL 同語意:鍵不存在回 sql.ErrNoRows、非整數即錯誤)。
func (f *FakeBilling) SystemActor(ctx context.Context) (int64, error) {
	v, err := f.Setting(ctx, "system_actor_user_id")
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("system_actor_user_id 格式錯誤: %q", v)
	}
	return id, nil
}

// Setting 讀營運參數;鍵不存在回 sql.ErrNoRows。
func (f *FakeBilling) Setting(_ context.Context, key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.settings[key]
	if !ok {
		return "", sql.ErrNoRows
	}
	return v, nil
}

// UpsertSettingTx 寫入營運參數。
func (f *FakeBilling) UpsertSettingTx(_ context.Context, _ *sql.Tx, key, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.settings[key] = value
	return nil
}

// --- 內部:查表與快照(呼叫端已持有鎖) ---

func (f *FakeBilling) subIndex(id int64) int {
	return slices.IndexFunc(f.subs, func(s Subscription) bool { return s.ID == id })
}

func (f *FakeBilling) periodIndex(id int64) int {
	return slices.IndexFunc(f.periods, func(p Period) bool { return p.ID == id })
}

func (f *FakeBilling) periodByNo(subID int64, periodNo int) (Period, bool) {
	for _, p := range f.periods {
		if p.SubscriptionID == subID && p.PeriodNo == periodNo {
			return p, true
		}
	}
	return Period{}, false
}

// latestPeriod 為該訂閱 period_no 最大的一期(對應 SQL 的 LATERAL 子查詢)。
func (f *FakeBilling) latestPeriod(subID int64) (Period, bool) {
	var best Period
	var found bool
	for _, p := range f.periods {
		if p.SubscriptionID != subID {
			continue
		}
		if !found || p.PeriodNo > best.PeriodNo {
			best, found = p, true
		}
	}
	return best, found
}

// expiredEmitted 回報該訂閱是否已發過 subscription.expired(對應 SQL 的 NOT EXISTS)。
func (f *FakeBilling) expiredEmitted(subID int64) bool {
	return slices.ContainsFunc(f.events, func(e Event) bool {
		return e.AggregateType == "subscription" && e.AggregateID == subID &&
			e.EventType == "subscription.expired"
	})
}

// filterSubs 依序回傳符合條件的訂閱(副本)。
func (f *FakeBilling) filterSubs(keep func(Subscription) bool) []Subscription {
	var out []Subscription
	for i := range f.subs {
		if keep(f.subs[i]) {
			out = append(out, cloneSubscription(f.subs[i]))
		}
	}
	return out
}

// fakeBillingState 為 WithTx 的快照(整份狀態的深拷貝)。
type fakeBillingState struct {
	nextSubID, nextPeriodID, nextEventID int64
	subs                                 []Subscription
	periods                              []Period
	events                               []Event
	dispatched                           map[int64]bool
	prices                               map[priceKey]Price
	audits                               []AuditRecord
	settings                             map[string]string
	platformCompanies                    map[int]bool
}

// snapshot 呼叫端必須已持有 f.mu（WithTx 在 Lock 期間呼叫）：snapshot 內部不再加鎖，
// 否則外層 Lock＋內層 Lock 在同一 goroutine 自死鎖（sync.Mutex 不可重入）。
func (f *FakeBilling) snapshot() fakeBillingState {
	s := fakeBillingState{
		nextSubID: f.nextSubID, nextPeriodID: f.nextPeriodID, nextEventID: f.nextEventID,
		subs:              make([]Subscription, len(f.subs)),
		periods:           make([]Period, len(f.periods)),
		events:            make([]Event, len(f.events)),
		dispatched:        make(map[int64]bool, len(f.dispatched)),
		prices:            make(map[priceKey]Price, len(f.prices)),
		audits:            make([]AuditRecord, len(f.audits)),
		settings:          make(map[string]string, len(f.settings)),
		platformCompanies: make(map[int]bool, len(f.platformCompanies)),
	}
	for i, x := range f.subs {
		s.subs[i] = cloneSubscription(x)
	}
	for i, x := range f.periods {
		s.periods[i] = clonePeriod(x)
	}
	for i, x := range f.events {
		x.Payload = slices.Clone(x.Payload)
		s.events[i] = x
	}
	for k, v := range f.dispatched {
		s.dispatched[k] = v
	}
	for k, v := range f.prices {
		s.prices[k] = v
	}
	for i, x := range f.audits {
		x.Before = slices.Clone(x.Before)
		x.After = slices.Clone(x.After)
		s.audits[i] = x
	}
	for k, v := range f.settings {
		s.settings[k] = v
	}
	for k, v := range f.platformCompanies {
		s.platformCompanies[k] = v
	}
	return s
}

func (f *FakeBilling) restore(s fakeBillingState) {
	f.nextSubID, f.nextPeriodID, f.nextEventID = s.nextSubID, s.nextPeriodID, s.nextEventID
	f.subs, f.periods, f.events = s.subs, s.periods, s.events
	f.dispatched, f.prices, f.audits, f.settings = s.dispatched, s.prices, s.audits, s.settings
	f.platformCompanies = s.platformCompanies
}

// clonePeriod 深拷貝一期(PaidAt 是指標:PG 每次掃描都是新配置,假實作也不得共用)。
func clonePeriod(p Period) Period {
	p.PaidAt = clonePtr(p.PaidAt)
	return p
}

// cmpInt64 為 slices.SortFunc 的三向比較。
func cmpInt64(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
