// Task 5：訂閱生命週期的**掃描式**轉移（產生下一期／逾期／寬限後停用／取消期末）單元契約。
//
// 用 store.NewFakeBilling（Task 3 的記憶體 BillingStore：語意與真 store 對齊，含 WithTx 的整份
// 快照還原）驗五件事：
//
//	① 狀態機逐格：active 且期末已過 → past_due（設寬限期）；past_due 且寬限已過 → suspended；
//	   不在表上的列（已取消／已停用）不得被排程動手；
//	② **已付款／已作廢的當期不算逾期**（C-1）：逾期後才補繳的客戶不得被重新催收、不得被停用，
//	   而它的下一期仍要照常開（否則催收主線的客戶會從此停止被開帳）；
//	③ 產生下一期：期別長度依 billing_cycle（G1：年繳 +1 年）與**第一期起日的日號**為錨
//	   （I-1：月底起租者不得因 2 月夾擠而漂移）、金額與價格欄位取「當期生效價」快照；
//	④ G7：cancelled 且期末已過 → subscription.expired（consumer 據此凍結公司），
//	   **不得**改變訂閱狀態（cancelled 是終態）；
//	⑤ 可重跑：四個掃描各跑第二趟 —— 不得有第二個期別、第二個事件、第二次狀態轉移，
//	   也不得補寫稽核（判斷只依「當前狀態 ＋ 時間」，不依賴呼叫次數）。
//
// **排程不寫平台稽核**：platform.audit_logs.operator_id 是 NOT NULL 且 FK 到 platform.operators，
// 而排程沒有 operator 主體（store.SystemActor 給的是租戶 users.id，不能當 operator_id，
// 硬寫會被 FK 23503 擋下）。凍結／復原的稽核由 consumer 經 SetCompanyStatus 落租戶稽核；
// 因此**每個由排程發出的 event payload 都自帶 reason 與 company_id**（見 I-2 的契約測試）。
//
// 掃描集合的邊界（最新一期仍 open 且期末 < now、grace_until IS NOT NULL、已發過
// subscription.expired 者不再選中）由 store 的 SQL 與假實作各自把關
// （fake_billing_test.go 有契約測試），本檔驗的是呼叫端的行為。
package billing_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"slices"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// at 造 UTC 時間（期別日期一律 UTC，免得時區把「日」的邊界弄模糊）。
func at(year int, month time.Month, day, hour int) time.Time {
	return time.Date(year, month, day, hour, 0, 0, 0, time.UTC)
}

// seedSub 種一列訂閱（ID 固定 5，斷言好讀）與它的第 1 期（open）。
// start 是**帳單日的唯一來源**：期別長度以它的日號為錨（1/31 起租的客戶永遠在月底結帳，I-1）。
func seedSub(f *store.FakeBilling, sub store.Subscription, start, end time.Time) int64 {
	sub.ID = 5
	id := f.PutSubscription(sub)
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: id, PeriodNo: 1, Status: "open",
		PeriodStart: start, PeriodEnd: end, PlanID: sub.PlanID, SeatCount: sub.SeatCount,
		AmountCents: amountCents, Currency: "TWD"})
	return id
}

// eventCount 數指定型別的事件筆數（重跑不得讓它變多）。
func eventCount(f *store.FakeBilling, eventType string) int {
	n := 0
	for _, e := range f.Events() {
		if e.EventType == eventType {
			n++
		}
	}
	return n
}

// openPeriods 回目前 open 的期別（重跑不得多開一筆）。
func openPeriods(t *testing.T, f *store.FakeBilling) []store.Period {
	t.Helper()
	ps, err := f.PeriodsByStatus(context.Background(), "open")
	if err != nil {
		t.Fatalf("讀 open 期別: %v", err)
	}
	return ps
}

// latestPeriod 讀訂閱的最新一期。
func latestPeriod(t *testing.T, f *store.FakeBilling, subID int64) store.Period {
	t.Helper()
	p, err := f.CurrentPeriodTx(context.Background(), nil, subID)
	if err != nil || p == nil {
		t.Fatalf("讀最新期別: %v（%+v）", err, p)
	}
	return *p
}

// openSequentially **連續**續開 n 期：每次都以「剛開好那一期的期末」前一小時當下一次的 now
// （所以每一趟都落在自己的提前窗內），回傳新開期別的期末（依序）。
// 連續性是 I-1 的關鍵 —— 每次都給一個新日期正好繞過錨點漂移，驗不出問題。
func openSequentially(t *testing.T, f *store.FakeBilling, subID int64, companyID int,
	now time.Time, leadDays, n int) []time.Time {
	t.Helper()
	b := billing.NewBilling(f)
	var ends []time.Time
	for i := 0; i < n; i++ {
		created, err := b.EnsureNextPeriod(context.Background(), companyID, now, leadDays)
		if err != nil {
			t.Fatalf("第 %d 次續開: %v", i+1, err)
		}
		if !created {
			t.Fatalf("第 %d 次續開應成立（now=%s）", i+1, now.Format(time.RFC3339))
		}
		end := latestPeriod(t, f, subID).PeriodEnd
		ends = append(ends, end)
		now = end.Add(-time.Hour)
	}
	return ends
}

// sameInstants 比對兩個時間序列（診斷訊息要看得懂日期，不比 reflect 的輸出）。
func sameInstants(got, want []time.Time) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if !got[i].Equal(want[i]) {
			return false
		}
	}
	return true
}

// formatInstants 把時間序列印成 RFC3339（失敗訊息用）。
func formatInstants(ts []time.Time) []string {
	out := make([]string, len(ts))
	for i, t := range ts {
		out[i] = t.Format(time.RFC3339)
	}
	return out
}

// 逾期轉移：active 且期末已過 → past_due ＋ 寬限期，並發事件（payload 帶得出公司與寬限期）；
// 重跑（狀態已非 active，掃描不含此列）不得再轉移、不得重複發事件。
func TestMarkPastDueSetsGraceAndEmitsOnce(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, BillingCycle: "monthly"},
		now.AddDate(0, -1, 0), now.Add(-time.Hour))
	b := billing.NewBilling(f)

	n, err := b.MarkPastDue(context.Background(), now, 7)
	if err != nil || n != 1 {
		t.Fatalf("應轉移 1 筆: n=%d err=%v", n, err)
	}
	wantGrace := now.AddDate(0, 0, 7)
	sub := readSub(t, f, 42)
	if sub.Status != "past_due" {
		t.Fatalf("狀態應為 past_due，got %q", sub.Status)
	}
	if sub.GraceUntil == nil || !sub.GraceUntil.Equal(wantGrace) {
		t.Fatalf("寬限期應為 %s，got %v", wantGrace, sub.GraceUntil)
	}
	var p struct {
		CompanyID  int    `json:"company_id"`
		GraceUntil string `json:"grace_until"`
		Reason     string `json:"reason"`
	}
	eventPayload(t, f, "subscription.past_due", &p)
	if p.CompanyID != 42 || p.GraceUntil != wantGrace.Format(time.RFC3339) || p.Reason == "" {
		t.Fatalf("事件需帶得出公司、寬限期與原因: %+v", p)
	}

	// 重跑：掃描條件（active ＋ 當期 open）已排除它 → 不轉移、不發第二個事件、不補寫稽核。
	auditsBefore := len(f.Audits())
	n2, err := b.MarkPastDue(context.Background(), now, 7)
	if err != nil || n2 != 0 {
		t.Fatalf("重跑不應轉移: n=%d err=%v", n2, err)
	}
	if got := eventCount(f, "subscription.past_due"); got != 1 {
		t.Fatalf("重跑不得重複發事件（%d 筆）: %v", got, eventTypes(f))
	}
	if got := len(f.Audits()); got != auditsBefore {
		t.Fatalf("重跑不得補寫稽核（%d → %d）", auditsBefore, got)
	}
	if sub := readSub(t, f, 42); !sub.GraceUntil.Equal(wantGrace) {
		t.Fatalf("重跑不得改動已設定的寬限期，got %v", sub.GraceUntil)
	}
}

// C-1：**已付款／已作廢的當期不算逾期**。少了這一條，逾期後才補繳的客戶（催收主線）會被重新
// 催收、寬限期重置，最後被停用凍結；而 EnsureNextPeriod 只認服務中的訂閱、RecordPayment 又不開期
// → 客戶從此停止被開帳。
func TestMarkPastDueSkipsPaidAndVoidPeriods(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	f := store.NewFakeBilling()
	// 42：逾期後已補繳（期別 paid）；44：期別作廢（void）；43：同期但未付款（open）。
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "active",
		PlanID: 1, SeatCount: 3, BillingCycle: "monthly"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "paid",
		PeriodStart: now.AddDate(0, -1, 0), PeriodEnd: now.Add(-time.Hour)})
	f.PutSubscription(store.Subscription{ID: 6, CompanyID: 43, Status: "active",
		PlanID: 1, SeatCount: 3, BillingCycle: "monthly"})
	f.PutPeriod(store.Period{ID: 10, SubscriptionID: 6, PeriodNo: 1, Status: "open",
		PeriodStart: now.AddDate(0, -1, 0), PeriodEnd: now.Add(-time.Hour)})
	f.PutSubscription(store.Subscription{ID: 7, CompanyID: 44, Status: "active",
		PlanID: 1, SeatCount: 3, BillingCycle: "monthly"})
	f.PutPeriod(store.Period{ID: 11, SubscriptionID: 7, PeriodNo: 1, Status: "void",
		PeriodStart: now.AddDate(0, -1, 0), PeriodEnd: now.Add(-time.Hour)})
	f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})
	b := billing.NewBilling(f)

	n, err := b.MarkPastDue(context.Background(), now, 7)
	if err != nil || n != 1 {
		t.Fatalf("只有未付款的 43 該轉 past_due: n=%d err=%v", n, err)
	}
	for _, company := range []int{42, 44} {
		if sub := readSub(t, f, company); sub.Status != "active" {
			t.Fatalf("公司 %d 的當期已付款／作廢，不得轉 past_due，got %q", company, sub.Status)
		}
		if sub := readSub(t, f, company); sub.GraceUntil != nil {
			t.Fatalf("公司 %d 不得被設定寬限期，got %v", company, sub.GraceUntil)
		}
	}
	if sub := readSub(t, f, 43); sub.Status != "past_due" {
		t.Fatalf("未付款的 43 仍應轉 past_due，got %q", sub.Status)
	}
	// 事件只有 43 那一筆（重跑時掃描已排除它，這裡一併驗「不得替已繳清者發事件」）。
	var companies []int
	for _, e := range f.Events() {
		if e.EventType != "subscription.past_due" {
			t.Fatalf("不應有其他事件，got %s", e.EventType)
		}
		var p struct {
			CompanyID int `json:"company_id"`
		}
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			t.Fatalf("payload 無法解析: %v", err)
		}
		companies = append(companies, p.CompanyID)
	}
	if !slices.Equal(companies, []int{43}) {
		t.Fatalf("只該對 43 發 past_due 事件，got %v", companies)
	}

	// 同趟的下一期：42 仍服務中 → 照常開下一期（C-1 的「客戶停止被開帳」不得發生）。
	created, err := b.EnsureNextPeriod(context.Background(), 42, now, 14)
	if err != nil || !created {
		t.Fatalf("已繳清且仍服務中的租戶必須照常開下一期: created=%v err=%v", created, err)
	}
	if p := readPeriod(t, f, 5, 2); !p.PeriodEnd.Equal(at(2026, time.November, 1, 2)) {
		t.Fatalf("下一期應接續 10/01 02:00 的期末（帳單日 1 日），got %s", p.PeriodEnd)
	}
	// 43 已逾期 → 不開期（避免對不服務的租戶繼續計費）。
	if created, err := b.EnsureNextPeriod(context.Background(), 43, now, 14); err != nil || created {
		t.Fatalf("已逾期者不得開新期: created=%v err=%v", created, err)
	}
}

// 寬限已過 → suspended 並發事件（該事件驅動產品域凍結）；不得殘留寬限期。
func TestSuspendOverdueEmitsEventAndClearsGrace(t *testing.T) {
	now := at(2026, time.October, 10, 3)
	grace := now.Add(-time.Hour)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "past_due", PlanID: 1,
		BillingCycle: "monthly", GraceUntil: &grace}, now.Add(-31*24*time.Hour), now.Add(-72*time.Hour))
	b := billing.NewBilling(f)

	n, err := b.SuspendOverdue(context.Background(), now)
	if err != nil || n != 1 {
		t.Fatalf("應停用 1 筆: n=%d err=%v", n, err)
	}
	sub := readSub(t, f, 42)
	if sub.Status != "suspended" || sub.GraceUntil != nil {
		t.Fatalf("應為 suspended 且清空寬限期，got status=%q grace=%v", sub.Status, sub.GraceUntil)
	}
	var p struct {
		CompanyID int    `json:"company_id"`
		Reason    string `json:"reason"`
	}
	eventPayload(t, f, "subscription.suspended", &p)
	if p.CompanyID != 42 || p.Reason == "" {
		t.Fatalf("事件需帶得出公司與原因（consumer 據此凍結）: %+v", p)
	}

	// 重跑：狀態已非 past_due → 不再轉移、不再發事件。
	n2, err := b.SuspendOverdue(context.Background(), now)
	if err != nil || n2 != 0 {
		t.Fatalf("重跑不應轉移: n=%d err=%v", n2, err)
	}
	if got := eventCount(f, "subscription.suspended"); got != 1 {
		t.Fatalf("重跑不得重複發事件（%d 筆）: %v", got, eventTypes(f))
	}
	if sub := readSub(t, f, 42); sub.Status != "suspended" {
		t.Fatalf("重跑不得改回其他狀態，got %q", sub.Status)
	}
}

// 寬限期內不得停用（寬限是給客戶補繳的時間窗，掃描條件為 grace_until < now）。
func TestSuspendOverdueLeavesSubscriptionInsideGrace(t *testing.T) {
	now := at(2026, time.October, 10, 3)
	grace := now.Add(time.Hour)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "past_due", PlanID: 1,
		BillingCycle: "monthly", GraceUntil: &grace}, now.Add(-31*24*time.Hour), now.Add(-72*time.Hour))

	n, err := billing.NewBilling(f).SuspendOverdue(context.Background(), now)
	if err != nil || n != 0 {
		t.Fatalf("寬限未過不得停用: n=%d err=%v", n, err)
	}
	if sub := readSub(t, f, 42); sub.Status != "past_due" {
		t.Fatalf("狀態應維持 past_due，got %q", sub.Status)
	}
	if evs := eventTypes(f); len(evs) != 0 {
		t.Fatalf("不得寫事件，got %v", evs)
	}
}

// cannedScan 覆寫兩支掃描查詢的結果：真 SQL 的 WHERE 不會帶出狀態不符的列，但狀態機的閘
// 不得把查詢當成唯一防線（查詢條件改壞時，這個閘是最後一道）。
type cannedScan struct {
	*store.FakeBilling
	due     []store.Subscription // ActiveSubscriptionsWithDueOpenPeriod
	expired []store.Subscription // PastDueSubscriptionsExpiredGrace
}

func (c cannedScan) ActiveSubscriptionsWithDueOpenPeriod(context.Context, *sql.Tx, time.Time) ([]store.Subscription, error) {
	return c.due, nil
}

func (c cannedScan) PastDueSubscriptionsExpiredGrace(context.Context, *sql.Tx, time.Time) ([]store.Subscription, error) {
	return c.expired, nil
}

// 狀態機逐格（排程的列）：只有表上允許的轉移能發生 —— 已取消是終態、已停用不得變逾期，
// 且被擋下時不得留下任何寫入。
func TestScheduleSkipsIllegalTransitions(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	for _, status := range []string{"cancelled", "suspended"} {
		t.Run(status, func(t *testing.T) {
			f := store.NewFakeBilling()
			seedSub(f, store.Subscription{CompanyID: 42, Status: status, PlanID: 1,
				BillingCycle: "monthly"}, now.AddDate(0, -1, 0), now.Add(-time.Hour))
			st := cannedScan{FakeBilling: f,
				due:     []store.Subscription{{ID: 5, CompanyID: 42, Status: status}},
				expired: []store.Subscription{{ID: 5, CompanyID: 42, Status: status}},
			}
			b := billing.NewBilling(st)

			if n, err := b.MarkPastDue(context.Background(), now, 7); err != nil || n != 0 {
				t.Fatalf("不得把 %q 轉為 past_due: n=%d err=%v", status, n, err)
			}
			if n, err := b.SuspendOverdue(context.Background(), now); err != nil || n != 0 {
				t.Fatalf("不得把 %q 轉為 suspended: n=%d err=%v", status, n, err)
			}
			if sub := readSub(t, f, 42); sub.Status != status {
				t.Fatalf("狀態不得改變，got %q", sub.Status)
			}
			if evs := eventTypes(f); len(evs) != 0 {
				t.Fatalf("不得寫事件，got %v", evs)
			}
			if audits := f.Audits(); len(audits) != 0 {
				t.Fatalf("不得寫稽核，got %v", audits)
			}
		})
	}
}

// C-1 的第二層防線：即使掃描多帶了一列「當期已付款」的訂閱（查詢條件改壞時），
// MarkPastDue 也必須自己擋住 —— 催收主線的客戶不得被重新催收。
func TestMarkPastDueSkipsPaidPeriodEvenIfScanned(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1,
		BillingCycle: "monthly"}, now.AddDate(0, -1, 0), now.Add(-time.Hour))
	// 把第 1 期標成已付款（收款是付款的唯一寫入路徑，這裡直接用 store 的付款方法）。
	if err := f.MarkPeriodPaidTx(context.Background(), nil, 9, now.Add(-time.Hour), "AB12345678", "", "", "", "manual", "BANK-1", ""); err != nil {
		t.Fatalf("MarkPeriodPaidTx: %v", err)
	}
	st := cannedScan{FakeBilling: f,
		due: []store.Subscription{{ID: 5, CompanyID: 42, Status: "active"}}}

	n, err := billing.NewBilling(st).MarkPastDue(context.Background(), now, 7)
	if err != nil || n != 0 {
		t.Fatalf("當期已付款不得標逾期: n=%d err=%v", n, err)
	}
	if sub := readSub(t, f, 42); sub.Status != "active" || sub.GraceUntil != nil {
		t.Fatalf("狀態不得改變且不得設寬限期，got %+v", sub)
	}
	if evs := eventTypes(f); len(evs) != 0 {
		t.Fatalf("不得寫事件，got %v", evs)
	}
}

// 產生下一期：期別號遞增、起日接續前期期末、價格與金額為當期生效價快照；重跑不得多開一期
// 或重複發 period.opened。
func TestEnsureNextPeriodOpensOnceWithPriceSnapshot(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	periodEnd := at(2026, time.October, 3, 3) // 2 天後 → 在 leadDays=14 的窗內
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
		BillingCycle: "monthly"}, at(2026, time.September, 3, 3), periodEnd)
	f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})
	b := billing.NewBilling(f)

	created, err := b.EnsureNextPeriod(context.Background(), 42, now, 14)
	if err != nil || !created {
		t.Fatalf("應建立下一期: created=%v err=%v", created, err)
	}
	next := readPeriod(t, f, 5, 2)
	if next.PeriodNo != 2 || !next.PeriodStart.Equal(periodEnd) ||
		!next.PeriodEnd.Equal(at(2026, time.November, 3, 3)) {
		t.Fatalf("期別號應遞增且接續前期期末: %+v", next)
	}
	if next.UnitPriceCents != 150000 || next.SeatPriceCents != 15000 || next.SeatCount != 3 ||
		next.AmountCents != 195000 || next.Currency != "TWD" || next.Status != "open" {
		t.Fatalf("期別快照不符（1500 + 150 × 3 = 1950 元）: %+v", next)
	}
	var p struct {
		CompanyID    int    `json:"company_id"`
		PeriodNo     int    `json:"period_no"`
		AmountCents  int64  `json:"amount_cents"`
		BillingCycle string `json:"billing_cycle"`
		Reason       string `json:"reason"`
	}
	eventPayload(t, f, "period.opened", &p)
	if p.CompanyID != 42 || p.PeriodNo != 2 || p.AmountCents != 195000 ||
		p.BillingCycle != "monthly" || p.Reason == "" {
		t.Fatalf("事件需帶得出公司、期別號、金額、週期與原因: %+v", p)
	}

	// 重跑：已有第 2 期 → 不新增、不再發事件。
	again, err := b.EnsureNextPeriod(context.Background(), 42, now, 14)
	if err != nil || again {
		t.Fatalf("重跑應回 false: created=%v err=%v", again, err)
	}
	if ps := openPeriods(t, f); len(ps) != 2 {
		t.Fatalf("重跑不得多開期別，got %d 筆: %+v", len(ps), ps)
	}
	if _, err := f.OpenPeriodByNoTx(context.Background(), nil, 5, 3); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("不得存在第 3 期: %v", err)
	}
	if got := eventCount(f, "period.opened"); got != 1 {
		t.Fatalf("重跑不得重複發 period.opened（%d 筆）: %v", got, eventTypes(f))
	}
}

// 價格快照：新期別取「當期生效價」，不得沿用舊期別的金額與價格欄位（調價不得回溯改帳，
// 新期別也不得用舊價）。
func TestEnsureNextPeriodSnapshotUsesCurrentPrice(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
		BillingCycle: "monthly"}, at(2026, time.September, 3, 3), at(2026, time.October, 3, 3))
	// 第 1 期快照是舊價（amountCents = 195000）；調價後（1800 + 200 × 3）新期別必須用新價。
	f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 180000, SeatCents: 20000, Currency: "TWD"})

	created, err := billing.NewBilling(f).EnsureNextPeriod(context.Background(), 42, now, 14)
	if err != nil || !created {
		t.Fatalf("應建立下一期: created=%v err=%v", created, err)
	}
	next := readPeriod(t, f, 5, 2)
	if next.AmountCents != 240000 {
		t.Fatalf("新期別金額應為 240000（當期生效價快照），got %d", next.AmountCents)
	}
	if next.UnitPriceCents != 180000 || next.SeatPriceCents != 20000 {
		t.Fatalf("價格快照未更新: %+v", next)
	}
	if p := readPeriod(t, f, 5, 1); p.AmountCents != amountCents {
		t.Fatalf("既有期別不得被調價改動，got %d", p.AmountCents)
	}
}

// G1：年繳訂閱的下一期期末必須 +1 年（月繳 +1 月）—— billing_cycle 漏帶等於少收 11 個月。
func TestEnsureNextPeriodUsesBillingCycle(t *testing.T) {
	end := at(2026, time.October, 31, 0)
	now := at(2026, time.October, 25, 3) // 在 leadDays=14 的提前窗內
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
		BillingCycle: "yearly"}, at(2025, time.October, 31, 0), end)
	f.PutPlanPrice(1, "yearly", store.Price{BaseCents: 1620000, SeatCents: 162000, Currency: "TWD"})

	created, err := billing.NewBilling(f).EnsureNextPeriod(context.Background(), 42, now, 14)
	if err != nil || !created {
		t.Fatalf("應建立下一期: created=%v err=%v", created, err)
	}
	next := readPeriod(t, f, 5, 2)
	if want := at(2027, time.October, 31, 0); !next.PeriodEnd.Equal(want) {
		t.Fatalf("年繳期末應為 %s（+1 年），got %s", want, next.PeriodEnd)
	}
	if !next.PeriodStart.Equal(end) {
		t.Fatalf("新期別起日應接續前期期末 %s，got %s", end, next.PeriodStart)
	}
	if next.AmountCents != 2106000 { // 16200 + 1620 × 3
		t.Fatalf("年繳金額應為 2106000，got %d", next.AmountCents)
	}
}

// I-1：月底起租者的帳單日不得漂移 —— **連續**續開時，2 月被夾成 28 之後必須回到 31。
// 漂移的後果：每期縮成 28 天、一年開出 13 期（月底客戶每年多收一期）。
func TestEnsureNextPeriodKeepsMonthEndAnchor(t *testing.T) {
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
		BillingCycle: "monthly"}, at(2026, time.January, 31, 0), at(2026, time.February, 28, 0))
	f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})

	ends := openSequentially(t, f, 5, 42, at(2026, time.February, 20, 0), 14, 3)
	want := []time.Time{
		at(2026, time.March, 31, 0), at(2026, time.April, 30, 0), at(2026, time.May, 31, 0),
	}
	if !sameInstants(ends, want) {
		t.Fatalf("連續續開的期末應為 %v（錨定 31 日），got %v", formatInstants(want), formatInstants(ends))
	}
}

// I-1：12 期必須恰好涵蓋一年。漂移（每期 28 天）時 364 天內就會擠進第 13 期 ——
// 月底起租的客戶每年被多收一期。
func TestEnsureNextPeriodOpensTwelvePeriodsPerYear(t *testing.T) {
	start := at(2026, time.January, 31, 0)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
		BillingCycle: "monthly"}, start, at(2026, time.February, 28, 0))
	f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})

	ends := openSequentially(t, f, 5, 42, at(2026, time.February, 20, 0), 14, 12)
	want := []time.Time{
		at(2026, time.March, 31, 0), at(2026, time.April, 30, 0), at(2026, time.May, 31, 0),
		at(2026, time.June, 30, 0), at(2026, time.July, 31, 0), at(2026, time.August, 31, 0),
		at(2026, time.September, 30, 0), at(2026, time.October, 31, 0), at(2026, time.November, 30, 0),
		at(2026, time.December, 31, 0), at(2027, time.January, 31, 0), at(2027, time.February, 28, 0),
	}
	if !sameInstants(ends, want) {
		t.Fatalf("12 期的期末應為 %v，got %v", formatInstants(want), formatInstants(ends))
	}
	// 正好 12 期涵蓋一年：第 12 期（period_no = 12）期末 = 起租日 + 1 年。
	// 漂移（每期 28 天）時它會落在 2026-12-28 → 一年內擠進 13 期。
	if got := ends[10]; !got.Equal(start.AddDate(1, 0, 0)) {
		t.Fatalf("第 12 期期末應為 %s（12 期恰好一年），got %s —— 期中被縮短會讓一年開出 13 期",
			start.AddDate(1, 0, 0), got)
	}
	if _, err := f.OpenPeriodByNoTx(context.Background(), nil, 5, 14); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("不得存在第 14 期: %v", err)
	}
	if ps := openPeriods(t, f); len(ps) != 13 {
		t.Fatalf("應恰好 13 筆 open（第 1 期 ＋ 12 期續開），got %d", len(ps))
	}
}

// I-1：年繳的閏日錨點（2/29 起租）只漂一次 —— 平年取 2/28，閏年回到 2/29。
func TestEnsureNextPeriodKeepsLeapDayAnchor(t *testing.T) {
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
		BillingCycle: "yearly"}, at(2028, time.February, 29, 0), at(2029, time.February, 28, 0))
	f.PutPlanPrice(1, "yearly", store.Price{BaseCents: 1620000, SeatCents: 162000, Currency: "TWD"})

	ends := openSequentially(t, f, 5, 42, at(2029, time.February, 20, 0), 14, 3)
	want := []time.Time{
		at(2030, time.February, 28, 0), at(2031, time.February, 28, 0), at(2032, time.February, 29, 0),
	}
	if !sameInstants(ends, want) {
		t.Fatalf("閏日錨點應為 %v（2032 回到 2/29），got %v", formatInstants(want), formatInstants(ends))
	}
}

// billing_cycle 為空字串或未知值時必須**大聲失敗**（G1）：默默當成月繳會讓年繳只收 1 個月。
// 價目刻意也存在於該週期底下 —— 這樣只有「週期無效」能讓它失敗，否則測不到真正的原因。
func TestEnsureNextPeriodRejectsInvalidBillingCycle(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	for _, cycle := range []string{"", "weekly"} {
		t.Run("cycle="+cycle, func(t *testing.T) {
			f := store.NewFakeBilling()
			seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
				BillingCycle: cycle}, at(2026, time.September, 3, 3), at(2026, time.October, 3, 3))
			f.PutPlanPrice(1, cycle, store.Price{BaseCents: 1620000, SeatCents: 162000, Currency: "TWD"})

			created, err := billing.NewBilling(f).EnsureNextPeriod(context.Background(), 42, now, 14)
			if err == nil || created {
				t.Fatalf("無效週期 %q 必須失敗: created=%v err=%v", cycle, created, err)
			}
			if code := errorCodeOf(t, err); code != "PLAT-3001" {
				t.Fatalf("ErrorInfo.code = %q；want PLAT-3001", code)
			}
			if connect.CodeOf(err) != connect.CodeFailedPrecondition {
				t.Fatalf("connect 碼應為 failed_precondition，got %v", connect.CodeOf(err))
			}
			if ps := openPeriods(t, f); len(ps) != 1 {
				t.Fatalf("失敗不得留下新期別: %+v", ps)
			}
			if evs := eventTypes(f); len(evs) != 0 {
				t.Fatalf("失敗不得寫事件，got %v", evs)
			}
		})
	}
}

// 提前窗的邊界：期末前 leadDays 天「當天」才進窗（不得提早開期，也不得晚一天）。
func TestEnsureNextPeriodWaitsForLeadWindow(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	cases := []struct {
		name      string
		periodEnd time.Time
		want      bool
	}{
		{"期末前 15 天（窗外）", now.AddDate(0, 0, 15), false},
		{"期末前 14 天（窗的邊界）", now.AddDate(0, 0, 14), true},
		{"期末前 2 天", now.AddDate(0, 0, 2), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := store.NewFakeBilling()
			seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
				BillingCycle: "monthly"}, tc.periodEnd.AddDate(0, -1, 0), tc.periodEnd)
			f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})

			created, err := billing.NewBilling(f).EnsureNextPeriod(context.Background(), 42, now, 14)
			if err != nil || created != tc.want {
				t.Fatalf("created=%v；want %v（err=%v）", created, tc.want, err)
			}
			wantPeriods := 1
			if tc.want {
				wantPeriods = 2
			}
			if ps := openPeriods(t, f); len(ps) != wantPeriods {
				t.Fatalf("open 期別應為 %d 筆，got %d: %+v", wantPeriods, len(ps), ps)
			}
		})
	}
}

// 非服務中的訂閱不得開新期（停用中還開期等於對不服務的租戶繼續計費）。
func TestEnsureNextPeriodSkipsNonBillableSubscription(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	for _, status := range []string{"past_due", "suspended", "cancelled"} {
		t.Run(status, func(t *testing.T) {
			f := store.NewFakeBilling()
			seedSub(f, store.Subscription{CompanyID: 42, Status: status, PlanID: 1, SeatCount: 3,
				BillingCycle: "monthly"}, now.AddDate(0, -1, 0), now.Add(48*time.Hour))
			f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})

			created, err := billing.NewBilling(f).EnsureNextPeriod(context.Background(), 42, now, 14)
			if err != nil || created {
				t.Fatalf("狀態 %q 不得開新期: created=%v err=%v", status, created, err)
			}
			if ps := openPeriods(t, f); len(ps) != 1 {
				t.Fatalf("不得多開期別: %+v", ps)
			}
			if evs := eventTypes(f); len(evs) != 0 {
				t.Fatalf("不得寫事件，got %v", evs)
			}
		})
	}

	// 沒有訂閱（無合約）與沒有期別（開通流程還沒產生第一期）都只是「這個租戶沒事可做」，
	// 不得回錯誤：一個資料不全的租戶不該讓整趟排程中斷。
	t.Run("沒有期別", func(t *testing.T) {
		f := store.NewFakeBilling()
		f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "active",
			PlanID: 1, SeatCount: 3, BillingCycle: "monthly"})
		f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})

		created, err := billing.NewBilling(f).EnsureNextPeriod(context.Background(), 42, now, 14)
		if err != nil || created {
			t.Fatalf("沒有期別時應為 no-op: created=%v err=%v", created, err)
		}
		if ps := openPeriods(t, f); len(ps) != 0 {
			t.Fatalf("不得憑空開期: %+v", ps)
		}
	})
	t.Run("沒有訂閱", func(t *testing.T) {
		f := store.NewFakeBilling()
		created, err := billing.NewBilling(f).EnsureNextPeriod(context.Background(), 42, now, 14)
		if err != nil || created {
			t.Fatalf("沒有訂閱時應為 no-op: created=%v err=%v", created, err)
		}
	})
}

// brokenLookup 讓期別查詢以非 sql.ErrNoRows 的錯誤失敗（連線中斷／權限）。
type brokenLookup struct{ *store.FakeBilling }

func (brokenLookup) OpenPeriodByNoTx(context.Context, *sql.Tx, int64, int) (*store.Period, error) {
	return nil, errors.New("模擬查詢失敗")
}

// 「期別不存在」只能由 sql.ErrNoRows 判定：其他錯誤必須中止，不得當成「不存在」而續開新期
// —— 否則一次連線抖動就會開出重複期別（帳會多收一期）。
func TestEnsureNextPeriodAbortsOnLookupFailure(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
		BillingCycle: "monthly"}, at(2026, time.September, 3, 3), at(2026, time.October, 3, 3))
	f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})

	created, err := billing.NewBilling(brokenLookup{f}).EnsureNextPeriod(context.Background(), 42, now, 14)
	if err == nil || created {
		t.Fatalf("查詢失敗必須中止: created=%v err=%v", created, err)
	}
	if code := errorCodeOf(t, err); code != "SYS-9000" {
		t.Fatalf("ErrorInfo.code = %q；want SYS-9000", code)
	}
	if ps := openPeriods(t, f); len(ps) != 1 {
		t.Fatalf("中止時不得開新期: %+v", ps)
	}
	if evs := eventTypes(f); len(evs) != 0 {
		t.Fatalf("中止時不得寫事件，got %v", evs)
	}
}

// G7：cancelled 且期末已過 → 發 subscription.expired（由 consumer 把公司轉 suspended）；
// **不得改變訂閱狀態**（cancelled 是終態，改狀態會破壞帳與稽核的可重現性）。
func TestExpireCancelledEmitsEventWithoutChangingStatus(t *testing.T) {
	now := at(2026, time.November, 1, 3)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "cancelled", PlanID: 1,
		BillingCycle: "monthly"}, now.AddDate(0, -1, 0), now.Add(-time.Hour))
	b := billing.NewBilling(f)

	n, err := b.ExpireCancelled(context.Background(), now)
	if err != nil || n != 1 {
		t.Fatalf("應處理 1 筆: n=%d err=%v", n, err)
	}
	// payload 契約：consumer 不得為了補欄位再查一次 DB，故兩個識別碼都要帶。
	var p struct {
		CompanyID      int    `json:"company_id"`
		SubscriptionID int64  `json:"subscription_id"`
		Reason         string `json:"reason"`
	}
	eventPayload(t, f, "subscription.expired", &p)
	if p.CompanyID != 42 || p.SubscriptionID != 5 || p.Reason == "" {
		t.Fatalf("payload 需帶得出公司、訂閱與原因: %+v", p)
	}
	if sub := readSub(t, f, 42); sub.Status != "cancelled" {
		t.Fatalf("排程不得改變訂閱狀態，got %q", sub.Status)
	}

	// 可重跑：掃描的 NOT EXISTS 已排除「發過 expired」者 → 0 筆、只有一個事件、不寫稽核。
	n2, err := b.ExpireCancelled(context.Background(), now)
	if err != nil || n2 != 0 {
		t.Fatalf("重跑不應再處理: n=%d err=%v", n2, err)
	}
	if got := eventCount(f, "subscription.expired"); got != 1 {
		t.Fatalf("重跑不得重複發事件（%d 筆）: %v", got, eventTypes(f))
	}
	if audits := f.Audits(); len(audits) != 0 {
		t.Fatalf("排程不寫平台稽核（actor 必須是 platform.operators），got %v", audits)
	}
	if sub := readSub(t, f, 42); sub.Status != "cancelled" {
		t.Fatalf("重跑後狀態仍應為 cancelled，got %q", sub.Status)
	}
}

// 期末未到不得發 expired：取消是「期末終止」，期末前仍提供服務（G7）。
func TestExpireCancelledWaitsForPeriodEnd(t *testing.T) {
	now := at(2026, time.November, 1, 3)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "cancelled", PlanID: 1,
		BillingCycle: "monthly"}, now.AddDate(0, -1, 0), now.Add(time.Hour))

	n, err := billing.NewBilling(f).ExpireCancelled(context.Background(), now)
	if err != nil || n != 0 {
		t.Fatalf("期末未到不得發 expired: n=%d err=%v", n, err)
	}
	if evs := eventTypes(f); len(evs) != 0 {
		t.Fatalf("不得寫事件，got %v", evs)
	}
}

// 已取消但期末已過者才發 expired：同一趟掃描下，只有期滿的那一家被處理。
func TestExpireCancelledOnlyTouchesPastPeriodEnd(t *testing.T) {
	now := at(2026, time.November, 1, 3)
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "cancelled",
		PlanID: 1, BillingCycle: "monthly"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open",
		PeriodEnd: now.Add(-time.Hour)})
	f.PutSubscription(store.Subscription{ID: 6, CompanyID: 43, Status: "cancelled",
		PlanID: 1, BillingCycle: "monthly"})
	f.PutPeriod(store.Period{ID: 10, SubscriptionID: 6, PeriodNo: 1, Status: "open",
		PeriodEnd: now.Add(24 * time.Hour)})

	n, err := billing.NewBilling(f).ExpireCancelled(context.Background(), now)
	if err != nil || n != 1 {
		t.Fatalf("應只處理期末已過的那一家: n=%d err=%v", n, err)
	}
	var expired []int64
	for _, e := range f.Events() {
		if e.EventType == "subscription.expired" {
			expired = append(expired, e.AggregateID)
		}
	}
	if !slices.Equal(expired, []int64{5}) {
		t.Fatalf("應只對訂閱 5 發 expired，got %v", expired)
	}
}

// I-2：排程**不寫平台稽核**（operator_id 是 NOT NULL 且 FK 到 platform.operators），
// 因此每個由排程發出的事件都必須自帶 company_id 與 reason —— 事件是唯一能回答
// 「為什麼」的地方（補繳、催收、客服追查都靠它）。
func TestScheduleEventPayloadCarriesCompanyAndReason(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T) *store.FakeBilling
		want []string
	}{
		{"產生期別", func(t *testing.T) *store.FakeBilling {
			now := at(2026, time.October, 1, 3)
			f := store.NewFakeBilling()
			seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
				BillingCycle: "monthly"}, at(2026, time.September, 3, 3), at(2026, time.October, 3, 3))
			f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})
			if created, err := billing.NewBilling(f).EnsureNextPeriod(
				context.Background(), 42, now, 14); err != nil || !created {
				t.Fatalf("EnsureNextPeriod: created=%v err=%v", created, err)
			}
			return f
		}, []string{"period.opened"}},
		{"逾期", func(t *testing.T) *store.FakeBilling {
			now := at(2026, time.October, 1, 3)
			f := store.NewFakeBilling()
			seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1,
				BillingCycle: "monthly"}, now.AddDate(0, -1, 0), now.Add(-time.Hour))
			if n, err := billing.NewBilling(f).MarkPastDue(
				context.Background(), now, 7); err != nil || n != 1 {
				t.Fatalf("MarkPastDue: n=%d err=%v", n, err)
			}
			return f
		}, []string{"subscription.past_due"}},
		{"停用", func(t *testing.T) *store.FakeBilling {
			now := at(2026, time.October, 10, 3)
			grace := now.Add(-time.Hour)
			f := store.NewFakeBilling()
			seedSub(f, store.Subscription{CompanyID: 42, Status: "past_due", PlanID: 1,
				BillingCycle: "monthly", GraceUntil: &grace},
				now.AddDate(0, -1, 0), now.Add(-72*time.Hour))
			if n, err := billing.NewBilling(f).SuspendOverdue(
				context.Background(), now); err != nil || n != 1 {
				t.Fatalf("SuspendOverdue: n=%d err=%v", n, err)
			}
			return f
		}, []string{"subscription.suspended"}},
		{"取消期末", func(t *testing.T) *store.FakeBilling {
			now := at(2026, time.November, 1, 3)
			f := store.NewFakeBilling()
			seedSub(f, store.Subscription{CompanyID: 42, Status: "cancelled", PlanID: 1,
				BillingCycle: "monthly"}, now.AddDate(0, -1, 0), now.Add(-time.Hour))
			if n, err := billing.NewBilling(f).ExpireCancelled(
				context.Background(), now); err != nil || n != 1 {
				t.Fatalf("ExpireCancelled: n=%d err=%v", n, err)
			}
			return f
		}, []string{"subscription.expired"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := tc.run(t)
			var got []string
			for _, e := range f.Events() {
				var p struct {
					CompanyID int    `json:"company_id"`
					Reason    string `json:"reason"`
				}
				if err := json.Unmarshal(e.Payload, &p); err != nil {
					t.Fatalf("事件 %s 的 payload 無法解析: %v（%s）", e.EventType, err, e.Payload)
				}
				if p.CompanyID != 42 || p.Reason == "" {
					t.Fatalf("事件 %s 必須帶得出公司與原因，got %s", e.EventType, e.Payload)
				}
				got = append(got, e.EventType)
			}
			if !slices.Equal(got, tc.want) {
				t.Fatalf("事件型別應為 %v，got %v", tc.want, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// I-1（Task 15 修正輪）：**試用到期**。少了這一步，試用就是「無上界的免費放行」：
// 判定層把 trialing 當可用、EnsureNextPeriod 把 trialing 當服務中（每期照開未付期別）、
// 而 MarkPastDue 只掃 active → 試用到期後既不催收也不凍結。
// ---------------------------------------------------------------------------

// TestExpireTrialsTransitionsToPastDueOnce 驗：trialing 且試用已過 → past_due ＋ 寬限期 ＋
// 一個 subscription.trial_ended（payload 帶得出公司／寬限／原因）；重跑不再轉也不再發；
// 試用**未到**的訂閱不動。
//
// 排程不寫平台稽核（platform.audit_logs.operator_id 是 NOT NULL 且 FK 到 platform.operators，而排程
// 沒有 operator 主體；見 lifecycle.go 檔頭），故這裡反向斷言**零筆稽核** —— 那條界線若被打破，
// 生產會在 FK 上炸，而單元測試的假 store 不會。
func TestExpireTrialsTransitionsToPastDueOnce(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	f := store.NewFakeBilling()
	trialEnd := now.Add(-time.Hour)
	seedSub(f, store.Subscription{CompanyID: 42, Status: "trialing", PlanID: 1, BillingCycle: "monthly",
		TrialEnds: &trialEnd}, now.AddDate(0, -1, 0), now.AddDate(0, 0, 20))
	b := billing.NewBilling(f)

	n, err := b.ExpireTrials(context.Background(), now, 7)
	if err != nil || n != 1 {
		t.Fatalf("應轉移 1 筆: n=%d err=%v", n, err)
	}
	wantGrace := now.AddDate(0, 0, 7)
	sub := readSub(t, f, 42)
	if sub.Status != "past_due" {
		t.Fatalf("試用到期應轉 past_due，got %q", sub.Status)
	}
	if sub.GraceUntil == nil || !sub.GraceUntil.Equal(wantGrace) {
		t.Fatalf("寬限期應為 %s，got %v", wantGrace, sub.GraceUntil)
	}
	// 試用到期日不得被清掉：它是「這個租戶試用到哪天」的唯一事實（判定層與 console 都讀它）。
	if sub.TrialEnds == nil || !sub.TrialEnds.Equal(trialEnd) {
		t.Fatalf("試用到期日必須保留，got %v", sub.TrialEnds)
	}
	var p struct {
		CompanyID  int    `json:"company_id"`
		GraceUntil string `json:"grace_until"`
		Reason     string `json:"reason"`
	}
	eventPayload(t, f, "subscription.trial_ended", &p)
	if p.CompanyID != 42 || p.GraceUntil != wantGrace.Format(time.RFC3339) || p.Reason == "" {
		t.Fatalf("事件需帶得出公司、寬限期與原因: %+v", p)
	}
	if got := eventCount(f, "subscription.trial_ended"); got != 1 {
		t.Fatalf("應恰發一個事件，got %d（%v）", got, eventTypes(f))
	}
	if audits := f.Audits(); len(audits) != 0 {
		t.Fatalf("排程不寫平台稽核（沒有 operator 主體），got %+v", audits)
	}

	// 重跑：狀態已不是 trialing → 不轉、不重發、不推進寬限期。
	n2, err := b.ExpireTrials(context.Background(), now, 7)
	if err != nil || n2 != 0 {
		t.Fatalf("重跑不應轉移: n=%d err=%v", n2, err)
	}
	if got := eventCount(f, "subscription.trial_ended"); got != 1 {
		t.Fatalf("重跑不得重複發事件（%d 筆）", got)
	}
	if sub := readSub(t, f, 42); !sub.GraceUntil.Equal(wantGrace) {
		t.Fatalf("重跑不得改動已設定的寬限期，got %v", sub.GraceUntil)
	}

	// 試用**未到** → 一個字都不動（差一刻也不算到期）。
	f2 := store.NewFakeBilling()
	future := now.Add(time.Second)
	seedSub(f2, store.Subscription{CompanyID: 43, Status: "trialing", PlanID: 1, BillingCycle: "monthly",
		TrialEnds: &future}, now.AddDate(0, -1, 0), now.AddDate(0, 0, 20))
	if n, err := billing.NewBilling(f2).ExpireTrials(context.Background(), now, 7); err != nil || n != 0 {
		t.Fatalf("試用未到不得轉移: n=%d err=%v", n, err)
	}
	if sub := readSub(t, f2, 43); sub.Status != "trialing" || sub.GraceUntil != nil {
		t.Fatalf("試用未到的訂閱不得被動到: %+v", sub)
	}
	if len(f2.Events()) != 0 {
		t.Fatalf("試用未到不得發事件，got %v", eventTypes(f2))
	}
}

// TestExpireTrialsOnlyTouchesTrialingWithTrialEnd 驗掃描的集合邊界：active／past_due／cancelled
// 不管 trial_ends_at 寫了什麼都不算「試用到期」；**trialing 但沒有到期日**也不算（那是資料異常，
// 該由開通端的上限擋住，而不是被排程猜成到期）。
func TestExpireTrialsOnlyTouchesTrialingWithTrialEnd(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	f := store.NewFakeBilling()
	oldTrial := now.Add(-time.Hour)
	for i, s := range []store.Subscription{
		{CompanyID: 42, Status: "active", TrialEnds: &oldTrial},
		{CompanyID: 43, Status: "past_due", TrialEnds: &oldTrial},
		{CompanyID: 44, Status: "cancelled", TrialEnds: &oldTrial},
		{CompanyID: 45, Status: "trialing"}, // 沒有到期日
	} {
		seedSub(f, store.Subscription{PlanID: 1, BillingCycle: "monthly", Status: s.Status,
			CompanyID: s.CompanyID, TrialEnds: s.TrialEnds, ID: int64(5 + i)},
			now.AddDate(0, -1, 0), now.AddDate(0, 0, 20))
	}

	n, err := billing.NewBilling(f).ExpireTrials(context.Background(), now, 7)
	if err != nil || n != 0 {
		t.Fatalf("只有 trialing 且有到期日者才算試用到期: n=%d err=%v", n, err)
	}
	for _, companyID := range []int{42, 43, 44, 45} {
		if sub := readSub(t, f, companyID); sub.Status != map[int]string{42: "active", 43: "past_due",
			44: "cancelled", 45: "trialing"}[companyID] {
			t.Fatalf("公司 %d 的狀態不得被動到，got %q", companyID, sub.Status)
		}
	}
	if len(f.Events()) != 0 {
		t.Fatalf("不得發任何事件，got %v", eventTypes(f))
	}
}
