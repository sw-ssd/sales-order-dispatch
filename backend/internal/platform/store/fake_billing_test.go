package store_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// FakeBilling 是 T4(收款)、T5(排程)、T6(outbox consumer)的單元測試地基:它的語意一旦與 SQL
// 實作不同,那些測試就會失真(而且失真方向是「假綠」)。以下逐一釘住它與 SQL 逐字對齊的契約:
// 期別不存在回 sql.ErrNoRows、不預先濾掉 cancelled、期別冪等鍵、付款的 three-way 語意、
// 稽核原因必填、WithTx 失敗整份還原。

// 現行訂閱的取法必須與 SQL 同序:優先未取消,只有全是 cancelled 時才取它;完全沒有列回
// (nil, nil) —— 少了這一條,T4/T5 的「已取消」案例會在單元測試裡假綠。
func TestFakeBillingOpenSubscriptionPrefersUncancelled(t *testing.T) {
	ctx := context.Background()
	f := store.NewFakeBilling()

	if sub, err := f.OpenSubscriptionTx(ctx, nil, 7); err != nil || sub != nil {
		t.Fatalf("無訂閱列應回 (nil, nil),got %+v err=%v", sub, err)
	}

	old := f.PutSubscription(store.Subscription{CompanyID: 7, PlanCode: "std", Status: "cancelled", SeatCount: 99})
	live := f.PutSubscription(store.Subscription{CompanyID: 7, PlanCode: "pro", Status: "active", SeatCount: 3})
	sub, err := f.OpenSubscriptionTx(ctx, nil, 7)
	if err != nil || sub == nil {
		t.Fatalf("OpenSubscriptionTx: %+v err=%v", sub, err)
	}
	if sub.ID != live || sub.Status != "active" || sub.SeatCount != 3 {
		t.Fatalf("同一公司有 active 與 cancelled 時應取 active,got %+v(取消的 id=%d)", *sub, old)
	}

	// 只有一筆 cancelled 的公司必須拿到它(F-8):判定層才分得出「已取消」與「從未訂閱」。
	only := store.NewFakeBilling()
	only.PutSubscription(store.Subscription{CompanyID: 8, Status: "cancelled", SeatCount: 5})
	sub, err = only.OpenSubscriptionTx(ctx, nil, 8)
	if err != nil || sub == nil {
		t.Fatalf("只有已取消訂閱的租戶必須回傳該列,got %+v err=%v", sub, err)
	}
	if sub.Status != "cancelled" || sub.SeatCount != 5 {
		t.Fatalf("回傳的必須是那一列已取消的訂閱,got %+v", *sub)
	}
}

// 期別:冪等鍵、不存在回 sql.ErrNoRows、CurrentPeriodTx 無期別回 (nil, nil)。
func TestFakeBillingPeriodsIdempotentAndNotFound(t *testing.T) {
	ctx := context.Background()
	f := store.NewFakeBilling()
	subID := f.PutSubscription(store.Subscription{CompanyID: 7, Status: "active"})
	in := store.OpenPeriodInput{
		SubscriptionID: subID, PeriodNo: 1,
		PeriodStart: time.Now(), PeriodEnd: time.Now().Add(24 * time.Hour),
		PlanID: 1, AmountCents: 195000, SeatCount: 3, Currency: "TWD",
	}
	first, err := f.OpenPeriodTx(ctx, nil, in)
	if err != nil || first == nil {
		t.Fatalf("OpenPeriodTx: %+v err=%v", first, err)
	}
	if first.Status != "open" || first.AmountCents != 195000 {
		t.Fatalf("新開的期別應為 open 且帶入快照金額,got %+v", *first)
	}
	again, err := f.OpenPeriodTx(ctx, nil, in)
	if err != nil || again.ID != first.ID {
		t.Fatalf("同期別重複開啟必須回同一列(排程可重跑),got %+v err=%v", again, err)
	}
	if _, err := f.OpenPeriodByNoTx(ctx, nil, subID, 2); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("不存在的期別應回 sql.ErrNoRows(呼叫端只認這個),got %v", err)
	}
	if got, err := f.OpenPeriodByNoTx(ctx, nil, subID, 1); err != nil || got.ID != first.ID {
		t.Fatalf("OpenPeriodByNoTx 應回同一期,got %+v err=%v", got, err)
	}
	if cur, err := f.CurrentPeriodTx(ctx, nil, subID); err != nil || cur.ID != first.ID {
		t.Fatalf("CurrentPeriodTx 應回最新一期,got %+v err=%v", cur, err)
	}
	empty := f.PutSubscription(store.Subscription{CompanyID: 9, Status: "trialing"})
	if cur, err := f.CurrentPeriodTx(ctx, nil, empty); err != nil || cur != nil {
		t.Fatalf("無期別應回 (nil, nil)(不是錯誤),got %+v err=%v", cur, err)
	}
}

// 付款的三種情境:入帳、同交易號重送 no-op(且 note 不得被空字串清掉)、交易號不同即拒絕。
func TestFakeBillingMarkPeriodPaidSemantics(t *testing.T) {
	ctx := context.Background()
	f := store.NewFakeBilling()
	subID := f.PutSubscription(store.Subscription{CompanyID: 7, Status: "active"})
	id := f.PutPeriod(store.Period{SubscriptionID: subID, PeriodNo: 1, Status: "open", AmountCents: 195000})
	paidAt := time.Now()

	if err := f.MarkPeriodPaidTx(ctx, nil, id, paidAt, "INV-1", "", "", "", "manual", "REF-1", "短收 100 元"); err != nil {
		t.Fatalf("MarkPeriodPaidTx: %v", err)
	}
	p, err := f.OpenPeriodByNoTx(ctx, nil, subID, 1)
	if err != nil {
		t.Fatalf("OpenPeriodByNoTx: %v", err)
	}
	if p.Status != "paid" || p.InvoiceNo != "INV-1" || p.ExternalRef != "REF-1" ||
		p.Note != "短收 100 元" || p.PaidAt == nil || !p.PaidAt.Equal(paidAt) {
		t.Fatalf("付款欄位未寫入,got %+v", *p)
	}
	// webhook 重播:同交易號不報錯,且**憑據不得被改動**(發票號／provider／paid_at 都要留著),
	// 空字串的 note 也不得清掉既有註記(G8)。
	if err := f.MarkPeriodPaidTx(ctx, nil, id, paidAt.Add(time.Hour), "", "", "", "", "", "REF-1", ""); err != nil {
		t.Fatalf("同交易號重送應為 no-op,got %v", err)
	}
	if p, _ := f.OpenPeriodByNoTx(ctx, nil, subID, 1); p.Note != "短收 100 元" ||
		p.InvoiceNo != "INV-1" || p.PaymentProvider != "manual" ||
		p.PaidAt == nil || !p.PaidAt.Equal(paidAt) {
		t.Fatalf("重播不得改動已入帳的憑據,got %+v", *p)
	}
	if err := f.MarkPeriodPaidTx(ctx, nil, id, paidAt, "INV-2", "", "", "", "manual", "REF-2", ""); err == nil {
		t.Fatal("已付款且交易號不同時必須拒絕覆蓋")
	}
	if p, _ := f.OpenPeriodByNoTx(ctx, nil, subID, 1); p.ExternalRef != "REF-1" {
		t.Fatalf("被拒絕的收款不得改動原交易號,got %q", p.ExternalRef)
	}
	// 同一 provider＋交易號不得入帳**兩期**(00029 的 periods_provider_ref_unique,部分唯一索引):
	// 少了這一條,T4 的「同一筆交易號不得重複入帳」在假實作上會假綠。
	other := f.PutPeriod(store.Period{SubscriptionID: subID, PeriodNo: 2, Status: "open", AmountCents: 195000})
	if err := f.MarkPeriodPaidTx(ctx, nil, other, paidAt, "INV-3", "", "", "", "manual", "REF-1", ""); err == nil {
		t.Fatal("同一交易號已入帳於另一期時必須拒絕(唯一索引)")
	}
	if p, _ := f.OpenPeriodByNoTx(ctx, nil, subID, 2); p.Status != "open" {
		t.Fatalf("被拒絕的跨期入帳不得改動期別,got %+v", *p)
	}
	// 作廢(void)的期別不得入帳,且錯誤要說出狀態(不得講成「已付款」)。
	void := f.PutPeriod(store.Period{SubscriptionID: subID, PeriodNo: 3, Status: "void"})
	err = f.MarkPeriodPaidTx(ctx, nil, void, paidAt, "INV-4", "", "", "", "manual", "", "")
	if err == nil {
		t.Fatal("作廢期別不得入帳")
	}
	if !strings.Contains(err.Error(), "void") {
		t.Fatalf("作廢期別的錯誤訊息必須說出狀態(不得講成「已付款」),got %v", err)
	}
	if p, _ := f.OpenPeriodByNoTx(ctx, nil, subID, 3); p.Status != "void" {
		t.Fatalf("被拒絕的入帳不得改動期別狀態,got %+v", *p)
	}
	if err := f.MarkPeriodPaidTx(ctx, nil, 404, paidAt, "", "", "", "", "", "", ""); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("期別不存在應回 sql.ErrNoRows,got %v", err)
	}
}

// 稽核的原因必填、價目缺週期即錯、排程三查詢的集合邊界(含「已發過 expired 不再選中」)。
func TestFakeBillingAuditPriceAndDueQueries(t *testing.T) {
	ctx := context.Background()
	f := store.NewFakeBilling()
	now := time.Now()

	if err := f.RecordAuditTx(ctx, nil, 1, "record_payment", "subscription", "7", "", nil, nil); err == nil {
		t.Fatal("平台稽核的原因為空字串時必須拒絕")
	}
	if err := f.RecordAuditTx(ctx, nil, 1, "record_payment", "subscription", "7", "匯款入帳", nil, nil); err != nil {
		t.Fatalf("RecordAuditTx: %v", err)
	}
	if audits := f.Audits(); len(audits) != 1 || audits[0].Action != "record_payment" || audits[0].OperatorID != 1 {
		t.Fatalf("稽核應留下一筆(含 operator 與 action),got %+v", audits)
	}

	f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})
	if p, err := f.CurrentPriceTx(ctx, nil, 1, "monthly"); err != nil || p.BaseCents != 150000 {
		t.Fatalf("CurrentPriceTx(monthly) = %+v err=%v", p, err)
	}
	// 不得靜默回 0 元:沒有價目等於免費送方案。
	if _, err := f.CurrentPriceTx(ctx, nil, 1, "yearly"); err == nil {
		t.Fatal("方案沒有該週期價目時必須回錯誤")
	}

	due := f.PutSubscription(store.Subscription{CompanyID: 11, PlanID: 1, Status: "active"})
	f.PutPeriod(store.Period{SubscriptionID: due, PeriodNo: 1, PeriodEnd: now.Add(-time.Hour)})
	future := f.PutSubscription(store.Subscription{CompanyID: 12, PlanID: 1, Status: "active"})
	f.PutPeriod(store.Period{SubscriptionID: future, PeriodNo: 1, PeriodEnd: now.Add(time.Hour)})
	// 16：active 但還沒有任何期別（不該被選中）；17／18：期末已過但當期已付款／已作廢
	// —— 逾期後才補繳的客戶不得被再次催收（C-1）。
	f.PutSubscription(store.Subscription{CompanyID: 16, PlanID: 1, Status: "active"})
	paidPeriod := f.PutSubscription(store.Subscription{CompanyID: 17, PlanID: 1, Status: "active"})
	f.PutPeriod(store.Period{SubscriptionID: paidPeriod, PeriodNo: 1, Status: "paid",
		PeriodEnd: now.Add(-time.Hour)})
	voidPeriod := f.PutSubscription(store.Subscription{CompanyID: 18, PlanID: 1, Status: "active"})
	f.PutPeriod(store.Period{SubscriptionID: voidPeriod, PeriodNo: 1, Status: "void",
		PeriodEnd: now.Add(-time.Hour)})
	grace := now.Add(-time.Hour)
	pastDue := f.PutSubscription(store.Subscription{CompanyID: 13, PlanID: 1, Status: "past_due", GraceUntil: &grace})
	// 14 是只設狀態、沒有寬限期的 past_due:它的作用是證明「無寬限期不算已到期」。
	f.PutSubscription(store.Subscription{CompanyID: 14, PlanID: 1, Status: "past_due"})
	gone := f.PutSubscription(store.Subscription{CompanyID: 15, PlanID: 1, Status: "cancelled"})
	f.PutPeriod(store.Period{SubscriptionID: gone, PeriodNo: 1, PeriodEnd: now.Add(-time.Hour)})

	subs, err := f.ActiveSubscriptionsWithDueOpenPeriod(ctx, nil, now)
	if err != nil || len(subs) != 1 || subs[0].ID != due {
		t.Fatalf("逾期未付應只含期末已過的 active 且當期仍 open(11；17 已付款、18 已作廢不算),got %+v err=%v", subs, err)
	}
	subs, err = f.PastDueSubscriptionsExpiredGrace(ctx, nil, now)
	// 寬限期為 nil 不算「已過」(與 SQL 的 grace_until IS NOT NULL 一致)。
	if err != nil || len(subs) != 1 || subs[0].ID != pastDue {
		t.Fatalf("寬限已過應只含 13(14 無寬限期),got %+v err=%v", subs, err)
	}
	subs, err = f.CancelledSubscriptionsPastPeriodEnd(ctx, nil, now)
	if err != nil || len(subs) != 1 || subs[0].ID != gone {
		t.Fatalf("已取消且期末已過應只含 15,got %+v err=%v", subs, err)
	}
	// 排程可重跑:發過 subscription.expired 之後不得再被選中。
	if err := f.EmitEventTx(ctx, nil, "subscription", gone, "subscription.expired", []byte(`{}`)); err != nil {
		t.Fatalf("EmitEventTx: %v", err)
	}
	if subs, _ := f.CancelledSubscriptionsPastPeriodEnd(ctx, nil, now); len(subs) != 0 {
		t.Fatalf("已發過 expired 者不得再被選中,got %+v", subs)
	}
	serving, err := f.ActiveOrTrialingSubscriptions(ctx)
	if err != nil {
		t.Fatalf("ActiveOrTrialingSubscriptions: %v", err)
	}
	if len(serving) != 5 { // 11／12／16／17／18 active;13 是 past_due、15 是 cancelled
		t.Fatalf("仍在服務中的應只有 active／trialing,got %+v", serving)
	}

	// outbox:未派送者取出後標記即不再出現(先把上一段的 expired 派送掉,只驗這一段自己寫的)。
	if pending, err := f.UndispatchedEvents(ctx, 10); err != nil || len(pending) != 1 {
		t.Fatalf("前面應只留下一筆未派送的 expired,got %+v err=%v", pending, err)
	} else if err := f.MarkEventDispatchedTx(ctx, nil, pending[0].ID); err != nil {
		t.Fatalf("MarkEventDispatchedTx: %v", err)
	}
	if err := f.EmitEventTx(ctx, nil, "subscription", due, "subscription.past_due", []byte(`{}`)); err != nil {
		t.Fatalf("EmitEventTx: %v", err)
	}
	// 無 payload（nil）的事件在真 store 會被存成 '{}'（''::jsonb 會 22P02），假實作必須一致,
	// 否則 consumer 的單元測試拿到的 payload 與真環境不同。
	if err := f.EmitEventTx(ctx, nil, "subscription", due, "subscription.suspended", nil); err != nil {
		t.Fatalf("EmitEventTx(nil payload): %v", err)
	}
	events, err := f.UndispatchedEvents(ctx, 1)
	if err != nil || len(events) != 1 || events[0].AggregateID != due ||
		events[0].EventType != "subscription.past_due" || events[0].ID != 2 {
		t.Fatalf("UndispatchedEvents 應照 id 序回一筆(且 limit 生效),got %+v err=%v", events, err)
	}
	if pending, _ := f.UndispatchedEvents(ctx, 10); len(pending) != 2 || string(pending[1].Payload) != "{}" {
		t.Fatalf("無 payload 的事件應以 '{}' 儲存(與 SQL 的 DEFAULT 一致),got %+v", pending)
	}
	if err := f.MarkEventDispatchedTx(ctx, nil, events[0].ID); err != nil {
		t.Fatalf("MarkEventDispatchedTx: %v", err)
	}
	if left, _ := f.UndispatchedEvents(ctx, 10); len(left) != 1 || left[0].EventType != "subscription.suspended" {
		t.Fatalf("已標記者不得再出現,got %+v", left)
	}
}

// WithTx 的錯誤路徑必須整份還原:排程與收款一旦中途失敗,不得留下半筆期別／事件／設定
// (記憶體沒有真的交易,這是假實作唯一能表達「同一交易」的地方)。
func TestFakeBillingWithTxRollsBackOnError(t *testing.T) {
	ctx := context.Background()
	f := store.NewFakeBilling()
	f.PutSetting("lead_days", "14")
	subID := f.PutSubscription(store.Subscription{CompanyID: 7, Status: "active"})
	before := f.PutPeriod(store.Period{SubscriptionID: subID, PeriodNo: 1, Status: "open"})
	paidAt := time.Now()

	sentinel := errors.New("付款規則拒絕")
	err := f.WithTx(ctx, func(*sql.Tx) error {
		if _, err := f.OpenPeriodTx(ctx, nil, store.OpenPeriodInput{
			SubscriptionID: subID, PeriodNo: 2, PlanID: 1, AmountCents: 195000, Currency: "TWD",
		}); err != nil {
			return err
		}
		if err := f.EmitEventTx(ctx, nil, "subscription", subID, "period.opened", []byte(`{}`)); err != nil {
			return err
		}
		if err := f.MarkPeriodPaidTx(ctx, nil, before, paidAt, "INV", "", "", "", "manual", "REF", "溢收"); err != nil {
			return err
		}
		if err := f.UpsertSettingTx(ctx, nil, "lead_days", "20"); err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("WithTx 應回傳 fn 的錯誤,got %v", err)
	}
	if events := f.Events(); len(events) != 0 {
		t.Fatalf("回滾後不得留下事件,got %+v", events)
	}
	if p, err := f.OpenPeriodByNoTx(ctx, nil, subID, 2); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("回滾後不得留下新開的期別,got %+v err=%v", p, err)
	}
	if p, _ := f.OpenPeriodByNoTx(ctx, nil, subID, 1); p.Status != "open" || p.Note != "" {
		t.Fatalf("回滾後既有期別不得被改動(付款還原),got %+v", *p)
	}
	if v, _ := f.Setting(ctx, "lead_days"); v != "14" {
		t.Fatalf("回滾後設定應還原,got %q", v)
	}
	// 成功路徑要真的留下來(否則「回滾」只是永遠回滾)。
	if err := f.WithTx(ctx, func(*sql.Tx) error {
		return f.UpsertSettingTx(ctx, nil, "lead_days", "20")
	}); err != nil {
		t.Fatalf("WithTx(成功): %v", err)
	}
	if v, _ := f.Setting(ctx, "lead_days"); v != "20" {
		t.Fatalf("成功的 WithTx 必須保留變更,got %q", v)
	}
	if actor, err := f.SystemActor(ctx); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("未設定 system_actor_user_id 應回 sql.ErrNoRows(不得靜默當 0),got %d err=%v", actor, err)
	}
	f.PutSetting("system_actor_user_id", " 7 ")
	if actor, err := f.SystemActor(ctx); err != nil || actor != 7 {
		t.Fatalf("SystemActor 應解析出 7(允許前後空白),got %d err=%v", actor, err)
	}
}
