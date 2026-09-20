// T9：訂閱生命週期的**營運寫入**（席位／改方案／取消）的單元契約。
//
// 用 store.NewFakeBilling（語意與真 store 對齊，含 WithTx 的整份快照還原）驗四件事：
//
//	① 狀態機邊界：cancelled 是終態（不得改席位／改方案）、已取消者重複取消為 no-op；
//	② v1 不做按日比例計費：席位與方案**下一期生效**，當期期別的計畫與金額快照一個字都不動；
//	③ 事件與稽核同一個交易：取消發 subscription.cancelled（payload 帶 reason），三條路徑各
//	   恰寫一筆平台稽核且 actor 是 operator；
//	④ 失效只有該租戶：狀態／方案／席位改了，判定快照就過期（提交後刪 ent:{companyID}）。
//
// 假實作沒有真的交易與列鎖，那條界線由 platform_admin_write_integration_test.go（真 PostgreSQL
// ＋ 真 Valkey）把關。
package billing_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	connect "connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// periodEnd 為夾具的當期期末（＝下一期的起日，也是「下一期生效」的生效點）。
var periodEnd = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

// subscriptionFixture 種一家 active 的訂閱（方案 1，5 席）與它的一期（open，期末 periodEnd）。
func subscriptionFixture() *store.FakeBilling {
	f := store.NewFakeBilling()
	f.PutPlan("std", 1)
	f.PutPlan("pro", 2)
	f.PutSubscription(store.Subscription{
		ID: 5, CompanyID: 42, PlanID: 1, SeatCount: 5, BillingCycle: "monthly", Status: "active"})
	f.PutPeriod(store.Period{
		ID: 9, SubscriptionID: 5, PeriodNo: 1, PlanID: 1, SeatCount: 5, AmountCents: 150000,
		Status: "open", PeriodStart: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), PeriodEnd: periodEnd})
	return f
}

// onlySubscription 取該公司唯一的訂閱（夾具只有一家）。
func onlySubscription(t *testing.T, f *store.FakeBilling, status string) store.Subscription {
	t.Helper()
	subs, err := f.ActiveOrTrialingSubscriptions(context.Background())
	if err != nil {
		t.Fatalf("讀訂閱: %v", err)
	}
	if status == "active" {
		if len(subs) != 1 {
			t.Fatalf("應有一筆仍在服務中的訂閱，got %d", len(subs))
		}
		return subs[0]
	}
	if len(subs) != 0 {
		t.Fatalf("狀態 %s 的訂閱不得被視為仍在服務中: %+v", status, subs)
	}
	return store.Subscription{}
}

// TestSetSeatCountUpdatesSubscriptionAndAudits 驗席位寫入：只改 subscriptions.seat_count、
// 恰寫一筆稽核（before／after 帶得出新舊席位）、提交後失效該租戶快取。
//
// 當期期別的 seat_count 是**開帳當下的快照**，不得跟著改（否則調席位等於回溯改帳）。
func TestSetSeatCountUpdatesSubscriptionAndAudits(t *testing.T) {
	f := subscriptionFixture()
	cache := &recordingCache{}

	seats, err := billing.NewBilling(f).WithCache(cache).SetSeatCount(context.Background(),
		billing.SetSeatCountInput{CompanyID: 42, SeatCount: 10, ActorOperatorID: 7, Reason: "擴編"})
	if err != nil || seats != 10 {
		t.Fatalf("SetSeatCount: seats=%d err=%v", seats, err)
	}
	if got := onlySubscription(t, f, "active").SeatCount; got != 10 {
		t.Fatalf("訂閱席位應更新為 10，got %d", got)
	}
	periods, err := f.PeriodsByStatus(context.Background(), "open")
	if err != nil || len(periods) != 1 {
		t.Fatalf("讀期別: %v (%d 筆)", err, len(periods))
	}
	if periods[0].SeatCount != 5 {
		t.Fatalf("當期期別的席位快照不得被改動，got %d", periods[0].SeatCount)
	}

	audits := f.Audits()
	if len(audits) != 1 {
		t.Fatalf("恰寫一筆稽核，got %d", len(audits))
	}
	a := audits[0]
	if a.Action != "subscription.set_seats" || a.TargetType != "subscription" || a.TargetID != "5" {
		t.Fatalf("稽核的動作／目標錯誤: %+v", a)
	}
	if a.OperatorID != 7 || a.Reason != "擴編" {
		t.Fatalf("稽核的 actor（operator）／原因錯誤: %+v", a)
	}
	var before, after struct {
		SeatCount int `json:"seat_count"`
	}
	if err := json.Unmarshal(a.Before, &before); err != nil {
		t.Fatalf("before 不是 JSON: %v (%s)", err, a.Before)
	}
	if err := json.Unmarshal(a.After, &after); err != nil {
		t.Fatalf("after 不是 JSON: %v (%s)", err, a.After)
	}
	if before.SeatCount != 5 || after.SeatCount != 10 {
		t.Fatalf("稽核必須記下新舊值（%d → %d）", before.SeatCount, after.SeatCount)
	}
	assertDeleted(t, cache, "ent:42")
}

// TestSetSeatCountRejectsCancelledSubscription 驗 cancelled 是終態：能改席位的合約才是活的，
// 對已取消的合約調席位只會讓 console 顯示一個「怎麼調都不會有下一期」的合約。
func TestSetSeatCountRejectsCancelledSubscription(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutPlan("std", 1)
	f.PutSubscription(store.Subscription{
		ID: 5, CompanyID: 42, PlanID: 1, SeatCount: 5, BillingCycle: "monthly", Status: "cancelled"})
	cache := &recordingCache{}

	_, err := billing.NewBilling(f).WithCache(cache).SetSeatCount(context.Background(),
		billing.SetSeatCountInput{CompanyID: 42, SeatCount: 10, ActorOperatorID: 7, Reason: "擴編"})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition || errorCodeOf(t, err) != "PLAT-3001" {
		t.Fatalf("已取消的訂閱應 PLAT-3001，got %v", err)
	}
	if len(f.Audits()) != 0 {
		t.Fatalf("失敗不得寫稽核，got %+v", f.Audits())
	}
	if len(cache.deleted) != 0 {
		t.Fatalf("失敗不得失效快取，got %v", cache.deleted)
	}
}

// TestSubscriptionOpsRequireValidInput 驗三條路徑的共同前置：reason 必填（平台稽核的 reason
// 是 NOT NULL，空字串等於一筆沒有原因的權限變更）、席位必須為正、方案代碼必填。
func TestSubscriptionOpsRequireValidInput(t *testing.T) {
	book := subscriptionFixture()
	b := billing.NewBilling(book)
	ctx := context.Background()

	if _, err := b.SetSeatCount(ctx, billing.SetSeatCountInput{
		CompanyID: 42, SeatCount: 5, ActorOperatorID: 7, Reason: "  "}); errorCodeOf(t, err) != "SYS-1001" {
		t.Fatalf("空 reason 應 SYS-1001，got %v", err)
	}
	if _, err := b.SetSeatCount(ctx, billing.SetSeatCountInput{
		CompanyID: 42, SeatCount: 0, ActorOperatorID: 7, Reason: "降席位"}); errorCodeOf(t, err) != "SYS-1001" {
		t.Fatalf("0 席應 SYS-1001，got %v", err)
	}
	if _, err := b.ChangePlan(ctx, billing.ChangePlanInput{
		CompanyID: 42, PlanCode: "pro", ActorOperatorID: 7, Reason: ""}); errorCodeOf(t, err) != "SYS-1001" {
		t.Fatalf("空 reason 應 SYS-1001，got %v", err)
	}
	if _, err := b.CancelSubscription(ctx, billing.CancelSubscriptionInput{
		CompanyID: 42, AtPeriodEnd: true, ActorOperatorID: 7, Reason: "\t"}); errorCodeOf(t, err) != "SYS-1001" {
		t.Fatalf("空 reason 應 SYS-1001，got %v", err)
	}
	// 前置驗證必須在動任何資料之前：沒有稽核、沒有事件，訂閱也還是原樣。
	if len(book.Audits()) != 0 || len(book.Events()) != 0 {
		t.Fatalf("前置驗證失敗時不得寫稽核或事件: %+v / %v", book.Audits(), eventTypes(book))
	}
	if got := onlySubscription(t, book, "active").SeatCount; got != 5 {
		t.Fatalf("前置驗證失敗時不得動到訂閱，got seat=%d", got)
	}
}

// TestChangePlanTakesEffectNextPeriod 驗改方案是「下一期生效」：subscriptions.plan_id 改了，
// **當期期別的 plan_id 與金額不動**（價格快照是已開帳的事實），回應的生效日為當期期末。
func TestChangePlanTakesEffectNextPeriod(t *testing.T) {
	f := subscriptionFixture()
	cache := &recordingCache{}

	effectiveFrom, err := billing.NewBilling(f).WithCache(cache).ChangePlan(context.Background(),
		billing.ChangePlanInput{CompanyID: 42, PlanCode: "pro", ActorOperatorID: 7, Reason: "升級方案"})
	if err != nil {
		t.Fatalf("ChangePlan: %v", err)
	}
	if !effectiveFrom.Equal(periodEnd) {
		t.Fatalf("生效日應為當期期末（%s），got %s", periodEnd, effectiveFrom)
	}
	if got := onlySubscription(t, f, "active").PlanID; got != 2 {
		t.Fatalf("訂閱的方案應改為 2，got %d", got)
	}
	periods, err := f.PeriodsByStatus(context.Background(), "open")
	if err != nil || len(periods) != 1 || periods[0].PlanID != 1 || periods[0].AmountCents != 150000 {
		t.Fatalf("當期期別的方案與金額快照不得被改動: %v (%+v)", err, periods)
	}
	audits := f.Audits()
	if len(audits) != 1 || audits[0].Action != "subscription.change_plan" || audits[0].OperatorID != 7 {
		t.Fatalf("應恰寫一筆改方案的稽核: %+v", audits)
	}
	assertDeleted(t, cache, "ent:42")
}

// TestChangePlanRejectsUnknownOrArchivedPlan 驗不存在／已歸檔的方案一律 SYS-4002，且不留痕跡。
// 假實作只註冊 active 的方案 —— 未註冊即「不存在或已歸檔」（與 SQL 的 status='active' 同語意）。
func TestChangePlanRejectsUnknownOrArchivedPlan(t *testing.T) {
	f := subscriptionFixture()
	cache := &recordingCache{}

	_, err := billing.NewBilling(f).WithCache(cache).ChangePlan(context.Background(),
		billing.ChangePlanInput{CompanyID: 42, PlanCode: "archived", ActorOperatorID: 7, Reason: "換方案"})
	if connect.CodeOf(err) != connect.CodeNotFound || errorCodeOf(t, err) != "SYS-4002" {
		t.Fatalf("不存在／已歸檔的方案應 SYS-4002，got %v", err)
	}
	if len(f.Audits()) != 0 || len(cache.deleted) != 0 {
		t.Fatalf("失敗不得寫稽核或失效快取: %+v / %v", f.Audits(), cache.deleted)
	}
	if got := onlySubscription(t, f, "active").PlanID; got != 1 {
		t.Fatalf("失敗不得改動訂閱，got plan=%d", got)
	}
}

// TestChangePlanSamePlanIsNoOp 驗「與現行方案相同」不寫稽核：一筆 before 與 after 一模一樣的
// 稽核會讓「誰真的改了方案」需要逐筆比對才看得出來。
func TestChangePlanSamePlanIsNoOp(t *testing.T) {
	f := subscriptionFixture()

	if _, err := billing.NewBilling(f).ChangePlan(context.Background(),
		billing.ChangePlanInput{CompanyID: 42, PlanCode: "std", ActorOperatorID: 7, Reason: "確認"}); err != nil {
		t.Fatalf("同方案應為 no-op 而非錯誤: %v", err)
	}
	if len(f.Audits()) != 0 {
		t.Fatalf("no-op 不得寫稽核，got %+v", f.Audits())
	}
}

// TestCancelSubscriptionAtPeriodEnd 驗期末終止：狀態轉 cancelled、發 subscription.cancelled
// （payload 帶 company_id／service_until／reason）、恰一筆稽核、失效該租戶快取。
//
// **不發 subscription.suspended**：取消是期末終止，期末前仍提供服務；凍結由排程的
// ExpireCancelled 發 subscription.expired、consumer 接手（spec §5.6）。
func TestCancelSubscriptionAtPeriodEnd(t *testing.T) {
	f := subscriptionFixture()
	cache := &recordingCache{}

	cancelled, err := billing.NewBilling(f).WithCache(cache).CancelSubscription(context.Background(),
		billing.CancelSubscriptionInput{CompanyID: 42, AtPeriodEnd: true, ActorOperatorID: 7, Reason: "客戶不續約"})
	if err != nil {
		t.Fatalf("CancelSubscription: %v", err)
	}
	if !cancelled.ServiceUntil.Equal(periodEnd) {
		t.Fatalf("服務應提供到期末（%s），got %s", periodEnd, cancelled.ServiceUntil)
	}
	if cancelled.CancelledAt.IsZero() {
		t.Fatal("應記下取消時間")
	}
	onlySubscription(t, f, "cancelled") // 不再是「仍在服務中」的訂閱

	if got := eventTypes(f); len(got) != 1 || got[0] != "subscription.cancelled" {
		t.Fatalf("應只發 subscription.cancelled，got %v", got)
	}
	var payload struct {
		CompanyID    int    `json:"company_id"`
		ServiceUntil string `json:"service_until"`
		Reason       string `json:"reason"`
	}
	eventPayload(t, f, "subscription.cancelled", &payload)
	if payload.CompanyID != 42 || payload.Reason != "客戶不續約" {
		t.Fatalf("事件 payload 必須自帶動手所需的欄位與原因: %+v", payload)
	}
	if payload.ServiceUntil != periodEnd.UTC().Format(time.RFC3339) {
		t.Fatalf("事件應帶服務到期時間，got %q", payload.ServiceUntil)
	}
	audits := f.Audits()
	if len(audits) != 1 || audits[0].Action != "subscription.cancel" || audits[0].OperatorID != 7 {
		t.Fatalf("應恰寫一筆取消稽核: %+v", audits)
	}
	assertDeleted(t, cache, "ent:42")
}

// TestCancelSubscriptionTwiceIsNoOp 驗重複取消：不重發事件、不重寫稽核、不推進 cancelled_at
// （那記的是取消發生的時間點，被重跑推進去就不再是事實）。
func TestCancelSubscriptionTwiceIsNoOp(t *testing.T) {
	f := subscriptionFixture()
	b := billing.NewBilling(f)
	in := billing.CancelSubscriptionInput{
		CompanyID: 42, AtPeriodEnd: true, ActorOperatorID: 7, Reason: "客戶不續約"}

	first, err := b.CancelSubscription(context.Background(), in)
	if err != nil {
		t.Fatalf("第一次取消: %v", err)
	}
	second, err := b.CancelSubscription(context.Background(), in)
	if err != nil {
		t.Fatalf("第二次取消應為 no-op 而非錯誤: %v", err)
	}
	if !second.CancelledAt.IsZero() {
		t.Fatalf("no-op 不得回報新的取消時間，got %s", second.CancelledAt)
	}
	if !second.ServiceUntil.Equal(first.ServiceUntil) {
		t.Fatalf("no-op 的服務到期時間不得改變: %s → %s", first.ServiceUntil, second.ServiceUntil)
	}
	if got := eventTypes(f); len(got) != 1 {
		t.Fatalf("重複取消不得重發事件，got %v", got)
	}
	if len(f.Audits()) != 1 {
		t.Fatalf("重複取消不得重寫稽核，got %+v", f.Audits())
	}
}

// TestCancelSubscriptionRejectsImmediateTermination 驗 at_period_end=false 一律拒絕（PLAT-3001）：
// v1 沒有按日比例計費，立即終止會產生一筆「已收但不再服務」的期別 —— 那是退款流程。
func TestCancelSubscriptionRejectsImmediateTermination(t *testing.T) {
	f := subscriptionFixture()
	cache := &recordingCache{}

	_, err := billing.NewBilling(f).WithCache(cache).CancelSubscription(context.Background(),
		billing.CancelSubscriptionInput{CompanyID: 42, AtPeriodEnd: false, ActorOperatorID: 7, Reason: "立即停"})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition || errorCodeOf(t, err) != "PLAT-3001" {
		t.Fatalf("立即終止應 PLAT-3001，got %v", err)
	}
	if len(f.Audits()) != 0 || len(f.Events()) != 0 || len(cache.deleted) != 0 {
		t.Fatalf("拒絕時不得留下任何痕跡: audits=%+v events=%v deleted=%v",
			f.Audits(), eventTypes(f), cache.deleted)
	}
	onlySubscription(t, f, "active")
}

// TestSubscriptionOpsWithoutCache 驗沒接快取時三條路徑照常運作：快取是加速器，不是前提
// （排程／CLI／單元測試不該因為沒有 Valkey 就不能改訂閱）。
func TestSubscriptionOpsWithoutCache(t *testing.T) {
	f := subscriptionFixture()
	b := billing.NewBilling(f)
	ctx := context.Background()

	if _, err := b.SetSeatCount(ctx, billing.SetSeatCountInput{
		CompanyID: 42, SeatCount: 6, ActorOperatorID: 7, Reason: "擴編"}); err != nil {
		t.Fatalf("未接快取時改席位應照常運作: %v", err)
	}
	if _, err := b.ChangePlan(ctx, billing.ChangePlanInput{
		CompanyID: 42, PlanCode: "pro", ActorOperatorID: 7, Reason: "升級"}); err != nil {
		t.Fatalf("未接快取時改方案應照常運作: %v", err)
	}
	if _, err := b.CancelSubscription(ctx, billing.CancelSubscriptionInput{
		CompanyID: 42, AtPeriodEnd: true, ActorOperatorID: 7, Reason: "不續約"}); err != nil {
		t.Fatalf("未接快取時取消應照常運作: %v", err)
	}
}

// TestSubscriptionOpsRejectMissingSubscription 驗沒有合約的公司一律 PLAT-3001（而不是 5xx）：
// 「這家公司沒有可調整的合約」是營運要看得懂的結果。
func TestSubscriptionOpsRejectMissingSubscription(t *testing.T) {
	// 方案要存在，否則 ChangePlan 會先停在「方案不存在」（兩個都是 4xx，但這條測的是「沒有合約」）。
	book := store.NewFakeBilling()
	book.PutPlan("std", 1)
	b := billing.NewBilling(book)
	ctx := context.Background()

	calls := map[string]func() error{
		"SetSeatCount": func() error {
			_, err := b.SetSeatCount(ctx, billing.SetSeatCountInput{
				CompanyID: 99, SeatCount: 5, ActorOperatorID: 7, Reason: "擴編"})
			return err
		},
		"ChangePlan": func() error {
			_, err := b.ChangePlan(ctx, billing.ChangePlanInput{
				CompanyID: 99, PlanCode: "std", ActorOperatorID: 7, Reason: "換方案"})
			return err
		},
		"CancelSubscription": func() error {
			_, err := b.CancelSubscription(ctx, billing.CancelSubscriptionInput{
				CompanyID: 99, AtPeriodEnd: true, ActorOperatorID: 7, Reason: "不續約"})
			return err
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			err := call()
			if errors.Is(err, sql.ErrNoRows) {
				t.Fatalf("沒有訂閱不是「列不存在」的驅動層錯誤: %v", err)
			}
			if connect.CodeOf(err) != connect.CodeFailedPrecondition || errorCodeOf(t, err) != "PLAT-3001" {
				t.Fatalf("沒有合約應 PLAT-3001，got %v", err)
			}
		})
	}
}
