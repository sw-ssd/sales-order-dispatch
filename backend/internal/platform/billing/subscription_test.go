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
	"strconv"
	"strings"
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

// 未結項 #22（RED）：無訂閱＋未知方案時，存在檢查應優先於參數檢查 ——
// 「這家公司沒有合約」是比「方案代碼打錯」更根本的事實。舊順序先查方案，回 SYS-4002
// 的 plan_code；新順序先驗訂閱，回 PLAT-3001 的「無合約」。
func TestChangePlanPrefersMissingSubscriptionOverUnknownPlan(t *testing.T) {
	f := subscriptionFixture()
	_, err := billing.NewBilling(f).ChangePlan(context.Background(),
		billing.ChangePlanInput{CompanyID: 9999, PlanCode: "archived", ActorOperatorID: 7, Reason: "換方案"})
	if errorCodeOf(t, err) != "PLAT-3001" {
		t.Fatalf("無訂閱＋未知方案應優先回 PLAT-3001（無合約），got %v", err)
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

// ---------------------------------------------------------------------------
// 開通（CreateSubscription）：建立訂閱 ＋ 第一期 ＋ 事件 ＋ 稽核，同一個交易（B-1）。
//
// 這一組測試是 B-1 的契約面：在它之前**全 repo 沒有任何程式路徑會建立 platform.subscriptions**
// （沒有訂閱就沒有可收款期別 → RecordPayment 回 PLAT-3001、EnsureNextPeriod 無期別即 no-op）。
// ---------------------------------------------------------------------------

// createFixture 種出可開通的假 store：方案 std（月繳 1000.00 ＋ 每席 200.00、年繳 10000.00 ＋
// 每席 2000.00）與方案 pro（**刻意沒有價目**：驗「缺價目不得開出 0 元期別」）。
func createFixture() *store.FakeBilling {
	f := store.NewFakeBilling()
	f.PutPlan("std", 1)
	f.PutPlan("pro", 2)
	f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 100000, SeatCents: 20000, Currency: "TWD"})
	f.PutPlanPrice(1, "yearly", store.Price{BaseCents: 1000000, SeatCents: 200000, Currency: "TWD"})
	return f
}

// nextMonthSameDay 回「下個月的同一個日號」（該日不存在時取當月最後一日）—— addBillingPeriod 的
// 月底錨點規則。測試自己算一次才驗得出「1/31 → 2/28」這類夾擠：若實作改用 time.AddDate，
// 1/31 會被正規化成 3/3（跳過整個 2 月），這裡就會紅。
func nextMonthSameDay(from time.Time) time.Time {
	lastDay := time.Date(from.Year(), from.Month()+2, 0, 0, 0, 0, 0, from.Location()).Day()
	day := from.Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(from.Year(), from.Month()+1, day,
		from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), from.Location())
}

// TestCreateSubscriptionOpensFirstPeriodAndAudits 驗開通的落地：訂閱列（狀態／方案／週期／席位）、
// 第一期（金額＝基價＋席位數×每席價、價格快照、period_no=1、期末依週期）、subscription.created
// 事件（payload 帶得出識別欄位）、恰一筆平台稽核、提交後失效該租戶快取。
func TestCreateSubscriptionOpensFirstPeriodAndAudits(t *testing.T) {
	f := createFixture()
	cache := &recordingCache{}

	before := time.Now()
	created, err := billing.NewBilling(f).WithCache(cache).CreateSubscription(context.Background(),
		billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "std", BillingCycle: "monthly", SeatCount: 3,
			ActorOperatorID: 7, Reason: "客戶簽約開通"})
	if err != nil {
		t.Fatalf("CreateSubscription: %v", err)
	}
	if created.Status != "active" || created.TrialEnds != nil {
		t.Fatalf("沒有 trial_ends_at 應直接 active，got %q (%v)", created.Status, created.TrialEnds)
	}

	// 訂閱列
	sub := readSub(t, f, 42)
	if sub.ID != created.SubscriptionID {
		t.Fatalf("回傳的訂閱 id 必須是剛建立那筆: %d vs %d", created.SubscriptionID, sub.ID)
	}
	if sub.Status != "active" || sub.PlanID != 1 || sub.BillingCycle != "monthly" || sub.SeatCount != 3 {
		t.Fatalf("訂閱列不符: %+v", sub)
	}

	// 第一期：金額 = 1000.00 ＋ 3 × 200.00 = 1600.00（快照齊備）
	periods, err := f.PeriodsByStatus(context.Background(), "open")
	if err != nil || len(periods) != 1 {
		t.Fatalf("應恰好一期 open: %v (%d 筆)", err, len(periods))
	}
	p := periods[0]
	if p.SubscriptionID != sub.ID || p.PeriodNo != 1 {
		t.Fatalf("期別必須是該訂閱的第 1 期: %+v", p)
	}
	if p.PlanID != 1 || p.UnitPriceCents != 100000 || p.SeatPriceCents != 20000 ||
		p.SeatCount != 3 || p.AmountCents != 160000 || p.Currency != "TWD" {
		t.Fatalf("價格快照不符（基價 1000.00、每席 200.00、3 席、合計 1600.00）: %+v", p)
	}
	if p.PeriodStart.Before(before) || p.PeriodStart.After(time.Now()) {
		t.Fatalf("第一期應自現在起算，got %v", p.PeriodStart)
	}
	if want := nextMonthSameDay(p.PeriodStart); !p.PeriodEnd.Equal(want) {
		t.Fatalf("月繳的期末應為下月同日（月底夾擠）: got %v want %v", p.PeriodEnd, want)
	}
	if created.FirstPeriod == nil || created.FirstPeriod.PeriodNo != 1 || created.FirstPeriod.AmountCents != 160000 {
		t.Fatalf("回傳值必須帶第一期（金額是已落地的快照）: %+v", created.FirstPeriod)
	}

	// 事件：consumer 不得為了補欄位再查一次 DB
	var payload struct {
		CompanyID      int    `json:"company_id"`
		SubscriptionID int64  `json:"subscription_id"`
		Reason         string `json:"reason"`
	}
	eventPayload(t, f, "subscription.created", &payload)
	if payload.CompanyID != 42 || payload.SubscriptionID != sub.ID || payload.Reason != "客戶簽約開通" {
		t.Fatalf("subscription.created 的 payload 不符: %+v", payload)
	}
	if types := eventTypes(f); len(types) != 1 {
		t.Fatalf("開通只寫一個事件（第一期的 period.opened 不寫：那是產期的路徑）: %v", types)
	}

	// 稽核：恰一筆，actor 是真的 operator，after 帶得出開通後的狀態與金額
	audits := f.Audits()
	if len(audits) != 1 {
		t.Fatalf("恰寫一筆稽核，got %d", len(audits))
	}
	a := audits[0]
	if a.Action != "subscription.create" || a.TargetType != "subscription" ||
		a.TargetID != strconv.FormatInt(sub.ID, 10) {
		t.Fatalf("稽核的動作／目標錯誤: %+v", a)
	}
	if a.OperatorID != 7 || a.Reason != "客戶簽約開通" {
		t.Fatalf("稽核的 actor（operator）／原因錯誤: %+v", a)
	}
	var nilBefore map[string]any
	if err := json.Unmarshal(a.Before, &nilBefore); err != nil || len(nilBefore) != 0 {
		t.Fatalf("開通沒有「之前」，before 應為空: %s (%v)", a.Before, err)
	}
	var after map[string]any
	if err := json.Unmarshal(a.After, &after); err != nil {
		t.Fatalf("after 不是 JSON: %v (%s)", err, a.After)
	}
	if after["status"] != "active" || after["amount_cents"] != float64(160000) ||
		after["billing_cycle"] != "monthly" || after["seat_count"] != float64(3) ||
		after["plan_code"] != "std" || after["period_no"] != float64(1) {
		t.Fatalf("稽核必須記下開通的內容: %+v", after)
	}

	assertDeleted(t, cache, "ent:42")
}

// TestCreateSubscriptionWithTrialStartsTrialing 驗試用的狀態選擇與第一期：trial_ends_at 在未來
// → trialing（且列上記下到期日），**第一期照開** —— 沒有期別時 EnsureNextPeriod 是 no-op
// （沒有當前期別就不知道起訖與期別號），試用到期時排程不會替他開帳。
func TestCreateSubscriptionWithTrialStartsTrialing(t *testing.T) {
	f := createFixture()
	trialEnds := time.Now().AddDate(0, 0, 14).Truncate(time.Second)

	created, err := billing.NewBilling(f).CreateSubscription(context.Background(),
		billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "std", BillingCycle: "monthly", SeatCount: 2,
			TrialEnds: &trialEnds, ActorOperatorID: 7, Reason: "POC 試用"})
	if err != nil {
		t.Fatalf("CreateSubscription（試用）: %v", err)
	}
	if created.Status != "trialing" {
		t.Fatalf("有未來 trial_ends_at 應 trialing，got %q", created.Status)
	}
	sub := readSub(t, f, 42)
	if sub.Status != "trialing" || sub.TrialEnds == nil || !sub.TrialEnds.Equal(trialEnds) {
		t.Fatalf("試用到期日必須落在訂閱列上: %+v", sub)
	}
	periods, err := f.PeriodsByStatus(context.Background(), "open")
	if err != nil || len(periods) != 1 || periods[0].PeriodNo != 1 {
		t.Fatalf("試用也必須有第一期: %v (%d 筆)", err, len(periods))
	}
	if periods[0].AmountCents != 140000 { // 1000.00 ＋ 2 × 200.00
		t.Fatalf("試用期的金額仍是價目快照（2 席 = 1400.00），got %d", periods[0].AmountCents)
	}
}

// TestCreateSubscriptionRejectsInvalidInput 驗參數邊界：每一項都必須**什麼都不留**（沒有訂閱、
// 沒有期別、沒有事件、沒有稽核、沒有失效）—— 開通是「一次寫入 = 一個交易」，失敗不得留半成品。
func TestCreateSubscriptionRejectsInvalidInput(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	// 上限 365 天（maxTrialDays）：試用是不收錢地放行全部權益，沒有上界就等於送出無限期免費。
	tooFar := time.Now().AddDate(0, 0, 366)
	cases := []struct {
		name string
		in   billing.CreateSubscriptionInput
		code string
	}{
		{"reason 全空白", billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "std", BillingCycle: "monthly", SeatCount: 3,
			ActorOperatorID: 7, Reason: "   "}, "SYS-1001"},
		{"plan_code 空", billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: " ", BillingCycle: "monthly", SeatCount: 3,
			ActorOperatorID: 7, Reason: "開通"}, "SYS-1001"},
		{"billing_cycle 空", billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "std", SeatCount: 3, ActorOperatorID: 7, Reason: "開通"}, "SYS-1001"},
		{"billing_cycle 未支援", billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "std", BillingCycle: "weekly", SeatCount: 3,
			ActorOperatorID: 7, Reason: "開通"}, "SYS-1001"},
		{"席位 0", billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "std", BillingCycle: "monthly", SeatCount: 0,
			ActorOperatorID: 7, Reason: "開通"}, "SYS-1001"},
		{"席位負數", billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "std", BillingCycle: "monthly", SeatCount: -1,
			ActorOperatorID: 7, Reason: "開通"}, "SYS-1001"},
		{"試用期已過", billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "std", BillingCycle: "monthly", SeatCount: 3,
			TrialEnds: &past, ActorOperatorID: 7, Reason: "開通"}, "SYS-1001"},
		{"試用期超過上限", billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "std", BillingCycle: "monthly", SeatCount: 3,
			TrialEnds: &tooFar, ActorOperatorID: 7, Reason: "開通"}, "SYS-1001"},
		{"方案不存在", billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "ghost", BillingCycle: "monthly", SeatCount: 3,
			ActorOperatorID: 7, Reason: "開通"}, "SYS-4002"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := createFixture()
			cache := &recordingCache{}
			_, err := billing.NewBilling(f).WithCache(cache).CreateSubscription(context.Background(), tc.in)
			if got := errorCodeOf(t, err); got != tc.code {
				t.Fatalf("應回 %s，got %q (%v)", tc.code, got, err)
			}
			assertNothingWritten(t, f, cache)
		})
	}

	// 方案存在但該週期沒有價目：**不得**開出 0 元期別（那等於免費送方案）。
	f := createFixture()
	cache := &recordingCache{}
	_, err := billing.NewBilling(f).WithCache(cache).CreateSubscription(context.Background(),
		billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "pro", BillingCycle: "monthly", SeatCount: 3,
			ActorOperatorID: 7, Reason: "開通"})
	if got := errorCodeOf(t, err); got != "PLAT-3001" {
		t.Fatalf("缺價目應 PLAT-3001，got %q (%v)", got, err)
	}
	if reason := errorInfoOf(t, err).GetDetails()["reason"]; !strings.Contains(reason, "不得開出 0 元期別") {
		t.Fatalf("錯誤必須說出「不得開出 0 元期別」，got %q", reason)
	}
	assertNothingWritten(t, f, cache)
}

// TestCreateSubscriptionAcceptsTrialAtUpperBound 驗上限的**邊界方向**：364 天可以、366 天不行
// （off-by-one 寫反的話，這一條會紅；單看「366 天被拒」看不出界線畫在哪）。
func TestCreateSubscriptionAcceptsTrialAtUpperBound(t *testing.T) {
	within := time.Now().AddDate(0, 0, 364).Truncate(time.Second)
	f := createFixture()
	created, err := billing.NewBilling(f).CreateSubscription(context.Background(),
		billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "std", BillingCycle: "monthly", SeatCount: 1,
			TrialEnds: &within, ActorOperatorID: 7, Reason: "一年期 POC"})
	if err != nil || created.Status != "trialing" {
		t.Fatalf("364 天的試用應被接受: %v (%+v)", err, created)
	}
	if sub := readSub(t, f, 42); sub.TrialEnds == nil || !sub.TrialEnds.Equal(within) {
		t.Fatalf("試用到期日應原樣落地，got %v", sub.TrialEnds)
	}
}

// priceLookupFails 讓取價以**基礎設施錯誤**（不是「沒有價目」）失敗：M-1 的映射要用它。
type priceLookupFails struct{ *store.FakeBilling }

func (priceLookupFails) CurrentPriceTx(context.Context, *sql.Tx, int64, string) (store.Price, error) {
	return store.Price{}, errors.New("連線中斷")
}

// TestCreateSubscriptionDistinguishesPriceLookupFailure 驗 M-1：「沒有該週期的價目」與「查價失敗」
// 必須是兩種答案。前者是資料問題（operator 要改的是價目 → PLAT-3001），後者是基礎設施問題
// （連線中斷、死鎖 → SYS-9000）；把後者也講成「這個方案沒有價目」會讓 operator 去改一個沒壞的設定。
func TestCreateSubscriptionDistinguishesPriceLookupFailure(t *testing.T) {
	f := createFixture()
	_, err := billing.NewBilling(priceLookupFails{f}).CreateSubscription(context.Background(),
		billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "std", BillingCycle: "monthly", SeatCount: 3,
			ActorOperatorID: 7, Reason: "開通"})
	if connect.CodeOf(err) != connect.CodeInternal || errorCodeOf(t, err) != "SYS-9000" {
		t.Fatalf("查價失敗（非缺價目）應 SYS-9000，got %v", err)
	}
	assertNothingWritten(t, f, &recordingCache{})
}

// TestCreateSubscriptionRollsBackWhenEventWriteFails 驗**回滾涵蓋訂閱列與第一期**（不只交易前的檢查）：
// 事件寫入失敗 → 訂閱、期別、稽核三者都不存在。少了這一條，「一次寫入 = 一個交易」只在
// 前置檢查那一段被驗過 —— 而開通要嘛整份成立，要嘛整份不成立。
func TestCreateSubscriptionRollsBackWhenEventWriteFails(t *testing.T) {
	f := createFixture()
	_, err := billing.NewBilling(failEvents{f}).CreateSubscription(context.Background(),
		billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "std", BillingCycle: "monthly", SeatCount: 3,
			ActorOperatorID: 7, Reason: "開通"})
	if err == nil {
		t.Fatal("事件寫入失敗時不得回成功")
	}
	if connect.CodeOf(err) != connect.CodeInternal || errorCodeOf(t, err) != "SYS-9000" {
		t.Fatalf("基礎設施失敗應 SYS-9000，got %v", err)
	}
	if sub, err := f.OpenSubscriptionTx(context.Background(), nil, 42); err != nil || sub != nil {
		t.Fatalf("回滾後不得留下訂閱列: %+v (%v)", sub, err)
	}
	if periods, err := f.PeriodsByStatus(context.Background(), "open"); err != nil || len(periods) != 0 {
		t.Fatalf("回滾後不得留下期別: %v (%d 筆)", err, len(periods))
	}
	if audits := f.Audits(); len(audits) != 0 {
		t.Fatalf("回滾後不得留下稽核，got %+v", audits)
	}
	if events := f.Events(); len(events) != 0 {
		t.Fatalf("回滾後不得留下事件，got %v", eventTypes(f))
	}
}

// TestCreateSubscriptionRejectsSecondLiveSubscription 驗 00029 的
// subscriptions_active_company_unique：同一公司不得同時有兩份未取消的合約 —— 兩份並行的合約
// 沒有「哪一份生效」的定義。回**已註冊的** SYS-2001（不是裸 connect 錯誤、也不是 5xx）。
//
// 已取消的合約不佔這條唯一鍵：要再服務是**新合約**（與 CancelSubscription 同一立場）。
func TestCreateSubscriptionRejectsSecondLiveSubscription(t *testing.T) {
	f := createFixture()
	f.PutSubscription(store.Subscription{
		ID: 5, CompanyID: 42, PlanID: 1, SeatCount: 3, BillingCycle: "monthly", Status: "active"})
	cache := &recordingCache{}

	_, err := billing.NewBilling(f).WithCache(cache).CreateSubscription(context.Background(),
		billing.CreateSubscriptionInput{
			CompanyID: 42, PlanCode: "std", BillingCycle: "monthly", SeatCount: 9,
			ActorOperatorID: 7, Reason: "誤按第二次"})
	if connect.CodeOf(err) != connect.CodeAlreadyExists || errorCodeOf(t, err) != "SYS-2001" {
		t.Fatalf("已有未取消的訂閱應 SYS-2001，got %v", err)
	}
	if sub := readSub(t, f, 42); sub.ID != 5 || sub.SeatCount != 3 {
		t.Fatalf("失敗不得改動既有合約: %+v", sub)
	}
	assertNothingWrittenAfter(t, f, cache, 1) // 夾具那筆仍在服務中，且不得多出第二筆

	// 已取消 → 允許開新合約（新的一筆列，舊的 cancelled 不動）。
	f.PutSubscription(store.Subscription{
		ID: 6, CompanyID: 43, PlanID: 1, SeatCount: 5, BillingCycle: "monthly", Status: "cancelled"})
	created, err := billing.NewBilling(f).CreateSubscription(context.Background(),
		billing.CreateSubscriptionInput{
			CompanyID: 43, PlanCode: "std", BillingCycle: "monthly", SeatCount: 2,
			ActorOperatorID: 7, Reason: "重新簽約"})
	if err != nil {
		t.Fatalf("已取消的合約不得擋住新合約: %v", err)
	}
	if created.SubscriptionID == 6 {
		t.Fatalf("必須是**新**合約（不得復活舊的）: %d", created.SubscriptionID)
	}
	sub := readSub(t, f, 43)
	if sub.ID != created.SubscriptionID || sub.Status != "active" {
		t.Fatalf("現行訂閱應是新合約: %+v", sub)
	}
}

// assertNothingWritten 斷言這次失敗沒有留下任何痕跡（沒有訂閱、沒有期別、沒有事件、沒有稽核、
// 沒有失效）。假 store 的 WithTx 整份還原 + 前置檢查在任何寫入之前 ＝ 呼叫端看得到「什麼都沒動」。
func assertNothingWritten(t *testing.T, f *store.FakeBilling, cache *recordingCache) {
	t.Helper()
	assertNothingWrittenAfter(t, f, cache, 0)
}

// assertNothingWrittenAfter 同 assertNothingWritten，但容許已經有 wantSubs 份訂閱列（夾具種的）。
func assertNothingWrittenAfter(t *testing.T, f *store.FakeBilling, cache *recordingCache, wantSubs int) {
	t.Helper()
	subs, err := f.ActiveOrTrialingSubscriptions(context.Background())
	if err != nil {
		t.Fatalf("讀訂閱: %v", err)
	}
	if len(subs) != wantSubs {
		t.Fatalf("失敗不得建立訂閱: 期望 %d 筆在服務中，got %d", wantSubs, len(subs))
	}
	periods, err := f.PeriodsByStatus(context.Background(), "open")
	if err != nil || len(periods) != 0 {
		t.Fatalf("失敗不得開期別: %v (%d 筆)", err, len(periods))
	}
	if len(f.Audits()) != 0 || len(f.Events()) != 0 {
		t.Fatalf("失敗不得寫稽核／事件: audits=%+v events=%v", f.Audits(), eventTypes(f))
	}
	if len(cache.deleted) != 0 {
		t.Fatalf("失敗不得失效快取: %v", cache.deleted)
	}
}
