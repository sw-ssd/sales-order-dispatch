// Task 5：訂閱生命週期的**掃描式**轉移（產生下一期／逾期／寬限後停用／取消期末）單元契約。
//
// 用 store.NewFakeBilling（Task 3 的記憶體 BillingStore：語意與真 store 對齊，含 WithTx 的整份
// 快照還原）驗四件事：
//
//	① 狀態機逐格：active 且期末已過 → past_due（設寬限期）；past_due 且寬限已過 → suspended；
//	   不在表上的列（已取消／已停用）不得被排程動手；
//	② 產生下一期：期別長度依 billing_cycle（G1：年繳 +1 年）、金額與價格欄位取「當期生效價」
//	   快照、期別號遞增、起日接續前期期末；
//	③ G7：cancelled 且期末已過 → subscription.expired（consumer 據此凍結公司），
//	   **不得**改變訂閱狀態（cancelled 是終態）；
//	④ 可重跑：四個掃描各跑第二趟 —— 不得有第二個期別、第二個事件、第二次狀態轉移，
//	   也不得補寫稽核（判斷只依「當前狀態 ＋ 時間」，不依賴呼叫次數）。
//
// **排程不寫平台稽核**：platform.audit_logs.operator_id 是 NOT NULL 且 FK 到 platform.operators，
// 而排程沒有 operator 主體（store.SystemActor 給的是租戶 users.id，不能當 operator_id，
// 硬寫會被 FK 23503 擋下）。凍結／復原的稽核由 consumer 經 SetCompanyStatus 落租戶稽核。
//
// 掃描集合的邊界（期末 < now、grace_until IS NOT NULL、已發過 subscription.expired 者不再選中）
// 由 store 的 SQL 與假實作各自把關（fake_billing_test.go 有契約測試），本檔驗的是呼叫端的行為。
package billing_test

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// seedSub 種一列訂閱（ID 固定 5，斷言好讀）與它的第 1 期（open）；期末由 periodEnd 決定。
func seedSub(f *store.FakeBilling, sub store.Subscription, periodEnd time.Time) int64 {
	sub.ID = 5
	id := f.PutSubscription(sub)
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: id, PeriodNo: 1, Status: "open",
		PeriodEnd: periodEnd, PlanID: sub.PlanID, SeatCount: sub.SeatCount,
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

// 逾期轉移：active 且期末已過 → past_due ＋ 寬限期，並發事件（payload 帶得出公司與寬限期）；
// 重跑（狀態已非 active，掃描不含此列）不得再轉移、不得重複發事件。
func TestMarkPastDueSetsGraceAndEmitsOnce(t *testing.T) {
	now := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, BillingCycle: "monthly"},
		now.Add(-time.Hour))
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
	}
	eventPayload(t, f, "subscription.past_due", &p)
	if p.CompanyID != 42 || p.GraceUntil != wantGrace.Format(time.RFC3339) {
		t.Fatalf("事件需帶得出公司與寬限期: %+v", p)
	}

	// 重跑：掃描條件（active）已排除它 → 不轉移、不發第二個事件、不補寫稽核。
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

// 寬限已過 → suspended 並發事件（該事件驅動產品域凍結）；不得殘留寬限期。
func TestSuspendOverdueEmitsEventAndClearsGrace(t *testing.T) {
	now := time.Date(2026, 10, 10, 3, 0, 0, 0, time.UTC)
	grace := now.Add(-time.Hour)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "past_due", PlanID: 1,
		BillingCycle: "monthly", GraceUntil: &grace}, now.Add(-72*time.Hour))
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
		CompanyID int `json:"company_id"`
	}
	eventPayload(t, f, "subscription.suspended", &p)
	if p.CompanyID != 42 {
		t.Fatalf("事件需帶得出公司（consumer 據此凍結）: %+v", p)
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
	now := time.Date(2026, 10, 10, 3, 0, 0, 0, time.UTC)
	grace := now.Add(time.Hour)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "past_due", PlanID: 1,
		BillingCycle: "monthly", GraceUntil: &grace}, now.Add(-72*time.Hour))

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
	now := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)
	for _, status := range []string{"cancelled", "suspended"} {
		t.Run(status, func(t *testing.T) {
			f := store.NewFakeBilling()
			seedSub(f, store.Subscription{CompanyID: 42, Status: status, PlanID: 1,
				BillingCycle: "monthly"}, now.Add(-time.Hour))
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

// 產生下一期：期別號遞增、起日接續前期期末、價格與金額為當期生效價快照；重跑不得多開一期
// 或重複發 period.opened。
func TestEnsureNextPeriodOpensOnceWithPriceSnapshot(t *testing.T) {
	now := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 10, 3, 3, 0, 0, 0, time.UTC) // 2 天後 → 在 leadDays=14 的窗內
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
		BillingCycle: "monthly"}, periodEnd)
	f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})
	b := billing.NewBilling(f)

	created, err := b.EnsureNextPeriod(context.Background(), 42, now, 14)
	if err != nil || !created {
		t.Fatalf("應建立下一期: created=%v err=%v", created, err)
	}
	next := readPeriod(t, f, 5, 2)
	if next.PeriodNo != 2 || !next.PeriodStart.Equal(periodEnd) ||
		!next.PeriodEnd.Equal(time.Date(2026, 11, 3, 3, 0, 0, 0, time.UTC)) {
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
	}
	eventPayload(t, f, "period.opened", &p)
	if p.CompanyID != 42 || p.PeriodNo != 2 || p.AmountCents != 195000 || p.BillingCycle != "monthly" {
		t.Fatalf("事件需帶得出公司、期別號、金額與週期: %+v", p)
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
	now := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
		BillingCycle: "monthly"}, now.Add(48*time.Hour))
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
	end := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 10, 25, 3, 0, 0, 0, time.UTC) // 在 leadDays=14 的提前窗內
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
		BillingCycle: "yearly"}, end)
	f.PutPlanPrice(1, "yearly", store.Price{BaseCents: 1620000, SeatCents: 162000, Currency: "TWD"})

	created, err := billing.NewBilling(f).EnsureNextPeriod(context.Background(), 42, now, 14)
	if err != nil || !created {
		t.Fatalf("應建立下一期: created=%v err=%v", created, err)
	}
	next := readPeriod(t, f, 5, 2)
	if want := time.Date(2027, 10, 31, 0, 0, 0, 0, time.UTC); !next.PeriodEnd.Equal(want) {
		t.Fatalf("年繳期末應為 %s（+1 年），got %s", want, next.PeriodEnd)
	}
	if !next.PeriodStart.Equal(end) {
		t.Fatalf("新期別起日應接續前期期末 %s，got %s", end, next.PeriodStart)
	}
	if next.AmountCents != 2106000 { // 16200 + 1620 × 3
		t.Fatalf("年繳金額應為 2106000，got %d", next.AmountCents)
	}
}

// billing_cycle 為空字串或未知值時必須**大聲失敗**（G1）：默默當成月繳會讓年繳只收 1 個月。
// 價目刻意也存在於該週期底下 —— 這樣只有「週期無效」能讓它失敗，否則測不到真正的原因。
func TestEnsureNextPeriodRejectsInvalidBillingCycle(t *testing.T) {
	now := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)
	for _, cycle := range []string{"", "weekly"} {
		t.Run("cycle="+cycle, func(t *testing.T) {
			f := store.NewFakeBilling()
			seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
				BillingCycle: cycle}, now.Add(48*time.Hour))
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
	now := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)
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
				BillingCycle: "monthly"}, tc.periodEnd)
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
	now := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)
	for _, status := range []string{"past_due", "suspended", "cancelled"} {
		t.Run(status, func(t *testing.T) {
			f := store.NewFakeBilling()
			seedSub(f, store.Subscription{CompanyID: 42, Status: status, PlanID: 1, SeatCount: 3,
				BillingCycle: "monthly"}, now.Add(48*time.Hour))
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

// brokenLookup 讓「查下一期是否已存在」以非 sql.ErrNoRows 的錯誤失敗（連線中斷／權限）。
type brokenLookup struct{ *store.FakeBilling }

func (brokenLookup) OpenPeriodByNoTx(context.Context, *sql.Tx, int64, int) (*store.Period, error) {
	return nil, errors.New("模擬查詢失敗")
}

// 「期別不存在」只能由 sql.ErrNoRows 判定：其他錯誤必須中止，不得當成「不存在」而續開新期
// —— 否則一次連線抖動就會開出重複期別（帳會多收一期）。
func TestEnsureNextPeriodAbortsOnLookupFailure(t *testing.T) {
	now := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "active", PlanID: 1, SeatCount: 3,
		BillingCycle: "monthly"}, now.Add(48*time.Hour))
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
	now := time.Date(2026, 11, 1, 3, 0, 0, 0, time.UTC)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "cancelled", PlanID: 1,
		BillingCycle: "monthly"}, now.Add(-time.Hour))
	b := billing.NewBilling(f)

	n, err := b.ExpireCancelled(context.Background(), now)
	if err != nil || n != 1 {
		t.Fatalf("應處理 1 筆: n=%d err=%v", n, err)
	}
	// payload 契約：consumer 不得為了補欄位再查一次 DB，故兩個識別碼都要帶。
	var p struct {
		CompanyID      int   `json:"company_id"`
		SubscriptionID int64 `json:"subscription_id"`
	}
	eventPayload(t, f, "subscription.expired", &p)
	if p.CompanyID != 42 || p.SubscriptionID != 5 {
		t.Fatalf("payload 需帶得出公司與訂閱: %+v", p)
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
	now := time.Date(2026, 11, 1, 3, 0, 0, 0, time.UTC)
	f := store.NewFakeBilling()
	seedSub(f, store.Subscription{CompanyID: 42, Status: "cancelled", PlanID: 1,
		BillingCycle: "monthly"}, now.Add(time.Hour))

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
	now := time.Date(2026, 11, 1, 3, 0, 0, 0, time.UTC)
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
