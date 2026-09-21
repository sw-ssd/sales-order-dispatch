// Task 7 的單元測試:排程的**編排**語意(順序、摘要、單飛、panic 復原、可重跑)。
//
// 假 deps(C-08:介面在 cron 套件內)說明:
//   - Billing 用 spy 包住**真的** billing.Billing(掃描語意屬 Task 5,不重造;spy 只記呼叫順序);
//   - Consumer 用**真的** consumer.Consumer(事件映射／認領語意屬 Task 6),它的兩個依賴才假:
//     假的系統交易把「認領」寫回假 store(故第二趟的 undispatched 查詢自然是空的 —— 可重跑是
//     資料造成的,不是假物件硬回 0),假的 setter 記下凍結有沒有真的走到產品域入口;
//   - Store 用 store.NewFakeBilling(事件／訂閱／期別／設定的唯一來源),需要「跑到一半」的閘門時
//     再包一層 gateStore;
//   - Lock 用記憶體版單飛鎖(真鎖是 PostgreSQL advisory lock,由整合測試釘住)。
package cron_test

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/consumer"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/cron"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// at 造 UTC 時間(期別日期一律 UTC,免得時區把「日」的邊界弄模糊)。
func at(year int, month time.Month, day, hour int) time.Time {
	return time.Date(year, month, day, hour, 0, 0, 0, time.UTC)
}

// callLog 記下一趟排程呼叫外部依賴的順序。帳務與派送兩個 spy 共用同一份,順序才是整體的順序
// (各自記一份只能知道「誰被呼叫過」,看不出「派送在產生期別之後」)。
type callLog struct {
	mu    sync.Mutex
	calls []string
}

func (l *callLog) record(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.calls = append(l.calls, name)
}

func (l *callLog) order() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return slices.Clone(l.calls)
}

// spyBilling 包住真的帳務:記下呼叫,可指定某一步失敗(驗「部分完成 + 錯誤」)。
type spyBilling struct {
	inner  *billing.Billing
	log    *callLog
	failOn string
}

func (s *spyBilling) call(name string) error {
	s.log.record(name)
	if s.failOn == name {
		return errors.New("假造的失敗:" + name)
	}
	return nil
}

func (s *spyBilling) ExpireTrials(ctx context.Context, now time.Time, graceDays int) (int, error) {
	if err := s.call("ExpireTrials"); err != nil {
		return 0, err
	}
	return s.inner.ExpireTrials(ctx, now, graceDays)
}

func (s *spyBilling) MarkPastDue(ctx context.Context, now time.Time, graceDays int) (int, error) {
	if err := s.call("MarkPastDue"); err != nil {
		return 0, err
	}
	return s.inner.MarkPastDue(ctx, now, graceDays)
}

func (s *spyBilling) SuspendOverdue(ctx context.Context, now time.Time) (int, error) {
	if err := s.call("SuspendOverdue"); err != nil {
		return 0, err
	}
	return s.inner.SuspendOverdue(ctx, now)
}

func (s *spyBilling) ExpireCancelled(ctx context.Context, now time.Time) (int, error) {
	if err := s.call("ExpireCancelled"); err != nil {
		return 0, err
	}
	return s.inner.ExpireCancelled(ctx, now)
}

func (s *spyBilling) EnsureNextPeriod(ctx context.Context, companyID int, now time.Time, leadDays int) (bool, error) {
	if err := s.call("EnsureNextPeriod"); err != nil {
		return false, err
	}
	return s.inner.EnsureNextPeriod(ctx, companyID, now, leadDays)
}

// spyDispatcher 包住真的 consumer:回傳值仍是「本趟認領的事件數」,只是把呼叫記進順序。
type spyDispatcher struct {
	inner *consumer.Consumer
	log   *callLog
}

func (s *spyDispatcher) DispatchOnce(ctx context.Context, limit int) (int, error) {
	s.log.record("DispatchOnce")
	return s.inner.DispatchOnce(ctx, limit)
}

// fakeSystemTx 假的系統範圍交易:認領＝把該事件標記已派送(寫回假 store),故第二趟的
// UndispatchedEvents 自然變空。條件式認領(0 列即跳過)的語意屬 Task 6 的測試,不在這裡重測。
type fakeSystemTx struct {
	f *store.FakeBilling
}

func (t *fakeSystemTx) Run(ctx context.Context, fn func(context.Context, consumer.Tx) error) error {
	return fn(ctx, fakeTx{t.f})
}

type fakeTx struct{ f *store.FakeBilling }

// Client 在單元測試裡沒有 ent:產品域入口是假的,不會用到它。
func (fakeTx) Client() *ent.Client { return nil }

func (t fakeTx) Claim(ctx context.Context, eventID int64) (bool, error) {
	return true, t.f.MarkEventDispatchedTx(ctx, nil, eventID)
}

// recordingSetter 記下產品域收到的狀態變更(凍結有沒有真的走到唯一入口)。
type recordingSetter struct {
	mu    sync.Mutex
	calls []int
}

func (s *recordingSetter) SetStatus(_ context.Context, _ *ent.Client, companyID int,
	_ company.Status, reason string, _ authz.Identity) error {
	if reason == "" {
		return errors.New("狀態變更必須有原因")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, companyID)
	return nil
}

func (s *recordingSetter) companies() []int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.calls)
}

// fakeLocker 為記憶體版單飛鎖:「同一時間只有一個持有者」＋取鎖／解鎖次數(解鎖必須發生,
// 否則下一趟全被擋住 —— panic 路徑尤其要看這個)。
type fakeLocker struct {
	mu       sync.Mutex
	held     bool
	tries    int
	releases int
}

func (l *fakeLocker) TryLock(context.Context) (func(context.Context) error, bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tries++
	if l.held {
		return nil, false, nil
	}
	l.held = true
	return func(ctx context.Context) error {
		l.mu.Lock()
		defer l.mu.Unlock()
		// 未結項 #19：假物件不得比真實作更嚴 —— 真實作的解鎖是一次 SQL（QueryRowContext），
		// 呼叫端 RunGuarded 永遠以 context.WithoutCancel 包過才傳進來，故 ctx.Err() 在此
		// 恆為 nil；舊的 ctx.Err() 閘門測的是「呼叫端沒包 WithoutCancel」，而那條路生產
		// 根本走不到（且真實作遇到死 ctx 會直接讓 SQL 失敗、丟棄連線放鎖，不會回 ctx.Err()）。
		// 守門對象改為「解鎖被呼叫且鎖被放掉」本身。
		l.held = false
		l.releases++
		return nil
	}, true, nil
}

func (l *fakeLocker) state() (held bool, tries, releases int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.held, l.tries, l.releases
}

// gateStore 讓第一趟停在「產生期別」的來源查詢上:此時鎖在手上、前面三個掃描已完成,
// 第二趟才真的與它重疊(競態要能被斷言就必須可控)。
type gateStore struct {
	inner   cron.Store
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (g *gateStore) Setting(ctx context.Context, key string) (string, error) {
	return g.inner.Setting(ctx, key)
}

func (g *gateStore) ActiveOrTrialingSubscriptions(ctx context.Context) ([]store.Subscription, error) {
	// 只有**第一次**呼叫會被擋住:第二趟在單飛正常時根本到不了這裡,而突變驗證(把鎖拿掉)時
	// 它必須能跑完並以斷言失敗收場,不能死在閘門上(那就看不出是哪一條錯了)。
	blocked := false
	g.once.Do(func() {
		blocked = true
		close(g.entered)
	})
	if blocked {
		<-g.release
	}
	return g.inner.ActiveOrTrialingSubscriptions(ctx)
}

func (g *gateStore) OverdueReceivablePeriods(ctx context.Context, now time.Time) ([]store.Period, error) {
	return g.inner.OverdueReceivablePeriods(ctx, now)
}

func (g *gateStore) TrialingSubscriptionsWithoutTrialEnd(ctx context.Context) ([]store.Subscription, error) {
	return g.inner.TrialingSubscriptionsWithoutTrialEnd(ctx)
}

// 未結項 #14:服務中卻沒有 open 期別的租戶，排程每一趟都靜默跳過。最新一期已 paid 且下一期
// 又開不出來(價目缺失)的訂閱，會永遠停在 active —— 既不被催收(只掃 open)，也不再被開帳
// (逐租戶失敗只記進 log)。這筆狀態必須從 cron 摘要看得見，否則 operator 的帳務視圖與
// 排程行為長期不一致。
func TestRunOnceCountsServiceableButUnbilled(t *testing.T) {
	f := store.NewFakeBilling()
	// 沒有價目:EnsureNextPeriod 開不出下一期(只記錯誤、不中斷)。
	seedSub(f, 42, "active", "monthly", nil, at(2026, time.October, 3, 3))
	deps, _, _, _ := newDeps(f)
	now := at(2026, time.October, 1, 3)

	s, err := cron.RunOnce(context.Background(), deps, now, cron.Params{GraceDays: 7, LeadDays: 14, EventBatch: 200})
	if err == nil {
		t.Fatal("缺價目應讓本趟帶錯誤回來")
	}
	if s.Unbilled != 1 {
		t.Fatalf("服務中卻無 open 期別的租戶應被計入摘要，got %d", s.Unbilled)
	}
}

// seedSub 種一列訂閱與它的第 1 期(open),期末由呼叫端決定;回傳訂閱 id。
func seedSub(f *store.FakeBilling, companyID int, status, cycle string, graceUntil *time.Time, end time.Time) int64 {
	id := f.PutSubscription(store.Subscription{
		CompanyID: companyID, Status: status, PlanID: 1, SeatCount: 3,
		BillingCycle: cycle, GraceUntil: graceUntil,
	})
	f.PutPeriod(store.Period{
		SubscriptionID: id, PeriodNo: 1, Status: "open",
		PeriodStart: end.AddDate(0, -1, 0), PeriodEnd: end,
		PlanID: 1, SeatCount: 3, AmountCents: 195000, Currency: "TWD",
	})
	return id
}

// newDeps 組出真帳務 + 真 consumer 的假 deps,回傳呼叫順序與 setter(看凍結有沒有走到產品域)。
func newDeps(f *store.FakeBilling) (cron.Deps, *callLog, *spyBilling, *recordingSetter) {
	log := &callLog{}
	spy := &spyBilling{inner: billing.NewBilling(f), log: log}
	setter := &recordingSetter{}
	return cron.Deps{
		Billing:  spy,
		Consumer: &spyDispatcher{inner: consumer.New(f, &fakeSystemTx{f: f}, setter), log: log},
		Store:    f,
		Lock:     &fakeLocker{},
	}, log, spy, setter
}

// seedTrialingSub 種一筆試用中的訂閱(trial_ends_at = trialEnds)與它的第 1 期(期末 end)。
func seedTrialingSub(f *store.FakeBilling, companyID int, trialEnds, end time.Time) int64 {
	id := f.PutSubscription(store.Subscription{
		CompanyID: companyID, Status: "trialing", PlanID: 1, SeatCount: 3,
		BillingCycle: "monthly", TrialEnds: &trialEnds,
	})
	f.PutPeriod(store.Period{
		SubscriptionID: id, PeriodNo: 1, Status: "open",
		PeriodStart: end.AddDate(0, -1, 0), PeriodEnd: end,
		PlanID: 1, SeatCount: 3, AmountCents: 195000, Currency: "TWD",
	})
	return id
}

// 一趟排程的順序與可重跑:逾期末付 → past_due;寬限已過 → suspended;已取消期末已過 → expired(G7);
// 進入提前窗 → 開下一期;事件在同一趟內被認領(含凍結)。第二趟不得產生任何第二個副作用。
//
// 待收款清單是**狀態**不是動作,故第二趟不歸零 —— 這一條同時釘住「receivables 不是本趟處理數」。
// 未結項 #40:trialing 但沒有到期日的訂閱永遠停在試用（ExpireTrials 刻意不碰：
// 不讓排程猜）。這種列可用、不催收、不凍結 —— 摘要必須有一欄讓 operator 看見它，
// 否則它與「正常試用中」在可觀測性上不可區分。
func TestRunOnceCountsStuckTrialing(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	p := cron.Params{GraceDays: 7, LeadDays: 14, EventBatch: 100}
	ctx := context.Background()

	f := store.NewFakeBilling()
	f.PutSetting("system_actor_user_id", "7")
	f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})
	// 42:有到期日且已過期的試用 → 轉 past_due（StuckTrialing 不得含它）。
	trialEnds := now.Add(-time.Hour)
	seedTrialingSub(f, 42, trialEnds, now.AddDate(0, 0, 10))
	// 43:trialing 但沒有到期日 → 永遠停在試用，摘要必須計 1。
	id := f.PutSubscription(store.Subscription{CompanyID: 43, Status: "trialing",
		PlanID: 1, SeatCount: 3, BillingCycle: "monthly"})
	f.PutPeriod(store.Period{SubscriptionID: id, PeriodNo: 1, Status: "open",
		PeriodStart: now.AddDate(0, -1, 0), PeriodEnd: now.AddDate(0, 0, 10),
		PlanID: 1, SeatCount: 3, AmountCents: 195000, Currency: "TWD"})

	deps, _, _, _ := newDeps(f)
	s, err := cron.RunOnce(ctx, deps, now, p)
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if s.TrialsExpired != 1 {
		t.Fatalf("有到期日的試用應轉走，got %d", s.TrialsExpired)
	}
	if s.StuckTrialing != 1 {
		t.Fatalf("無到期日的試用應被計入摘要，got %d", s.StuckTrialing)
	}
	// 重跑：轉走的不再轉，但卡住的仍在 —— 狀態欄位不得歸零。
	second, err := cron.RunOnce(ctx, deps, now, p)
	if err != nil {
		t.Fatalf("第二趟: %v", err)
	}
	if second.StuckTrialing != 1 {
		t.Fatalf("卡住的試用仍在，重跑不得歸零，got %d", second.StuckTrialing)
	}
}

func TestRunOnceIsIdempotent(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	p := cron.Params{GraceDays: 7, LeadDays: 14, EventBatch: 100}
	ctx := context.Background()

	f := store.NewFakeBilling()
	f.PutSetting("system_actor_user_id", "7")
	f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})
	// 42:逾期未付(→past_due);43:寬限已過(→suspended);44:已取消且期末已過(→expired,G7);
	// 45:服務中且期末在提前窗內(→開下一期);46:試用已到期(→past_due,且**不得**被開新期)。
	seedSub(f, 42, "active", "monthly", nil, now.Add(-time.Hour))
	pastGrace := now.Add(-24 * time.Hour)
	seedSub(f, 43, "past_due", "monthly", &pastGrace, now.Add(-2*time.Hour))
	seedSub(f, 44, "cancelled", "monthly", nil, now.Add(-2*time.Hour))
	seedSub(f, 45, "active", "monthly", nil, now.AddDate(0, 0, 10))
	// 46 的期末刻意放在提前窗**內**（10 天 < LeadDays=14；45 正是靠這點才被開出第 2 期）:
	// 若試用到期沒有先轉走狀態,本趟就會替它開出第 2 期。
	sub46 := seedTrialingSub(f, 46, now.Add(-time.Hour), now.AddDate(0, 0, 10))

	deps, log, _, setter := newDeps(f)

	first, err := cron.RunOnce(ctx, deps, now, p)
	if err != nil {
		t.Fatalf("第一趟: %v", err)
	}
	// 四類掃描的順序(逾期 → 停用 → 取消到期 → 產生期別)＋派送放最後:本趟產生的每個事件
	// (含凍結)都在同一趟內被認領,不留「事件寫了但沒送」的窗口。
	wantOrder := []string{"ExpireTrials", "MarkPastDue", "SuspendOverdue", "ExpireCancelled", "EnsureNextPeriod", "DispatchOnce"}
	if got := log.order(); !slices.Equal(got, wantOrder) {
		t.Fatalf("掃描順序不符: got %v want %v", got, wantOrder)
	}
	if first.TrialsExpired != 1 || first.PastDue != 1 || first.Suspended != 1 ||
		first.ExpiredCancelled != 1 || first.PeriodsOpened != 1 {
		t.Fatalf("第一趟計數不符: %+v", first)
	}
	// 5 筆事件:trial_ended／past_due／suspended／expired／period.opened(最後一筆是本趟開期產生的,
	// 派送放在產生期別之後才會在同一趟被認領)。
	if first.Dispatched != 5 {
		t.Fatalf("第一趟應認領 5 筆事件(trial_ended／past_due／suspended／expired／period.opened),got %d",
			first.Dispatched)
	}
	// 試用到期只把狀態轉成 past_due(仍可用、尚未凍結)→ **不得**有產品域動作;
	// 而且它轉走之後就不在服務中 → 不會被開下一期(46 只有第 1 期)。
	if _, err := f.OpenPeriodByNoTx(ctx, nil, sub46, 2); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("試用已到期的租戶不得被開第 2 期（轉 past_due 之後就不在服務中）: %v", err)
	}
	// 凍結真的走到產品域唯一入口:43(suspended)與 44(expired→凍結)。
	if got := setter.companies(); !slices.Equal(got, []int{43, 44}) {
		t.Fatalf("凍結應走到產品域入口(43、44),got %v", got)
	}
	// 待收款:42／43／44 的第 1 期仍是 open 且期末已過(45 已開下一期,不是待收款)。
	if first.Receivables != 3 {
		t.Fatalf("待收款應為 3 筆(open 且期末已過),got %d", first.Receivables)
	}

	second, err := cron.RunOnce(ctx, deps, now, p)
	if err != nil {
		t.Fatalf("第二趟: %v", err)
	}
	if second.TrialsExpired != 0 || second.PastDue != 0 || second.Suspended != 0 ||
		second.ExpiredCancelled != 0 || second.PeriodsOpened != 0 || second.Dispatched != 0 {
		t.Fatalf("重跑不得重複轉移或重複派送: %+v", second)
	}
	if second.Receivables != first.Receivables {
		t.Fatalf("待收款是狀態不是動作,重跑不得歸零: got %d want %d", second.Receivables, first.Receivables)
	}
	if got := setter.companies(); !slices.Equal(got, []int{43, 44}) {
		t.Fatalf("重跑不得再動產品域,got %v", got)
	}
	if events := len(f.Events()); events != 5 {
		t.Fatalf("重跑不得產生第二個事件,got %d 筆", events)
	}
}

// 部分完成如實回報:某一步失敗即中止(不再產生期別、不派送),但**已完成的計數不得被吞掉** ——
// billing 的每個掃描各自是一個交易,已完成的是已落地的帳務事實,摘要若不報,帳就看不出來。
func TestRunOnceReportsPartialSummaryOnFailure(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	p := cron.Params{GraceDays: 7, LeadDays: 14, EventBatch: 100}
	f := store.NewFakeBilling()
	f.PutSetting("system_actor_user_id", "7")
	seedSub(f, 42, "active", "monthly", nil, now.Add(-time.Hour))

	deps, log, spy, _ := newDeps(f)
	spy.failOn = "SuspendOverdue"

	got, err := cron.RunOnce(context.Background(), deps, now, p)
	if err == nil {
		t.Fatal("失敗必須回報(呼叫端以非零離開碼收場)")
	}
	if !strings.Contains(err.Error(), "停用欠費") {
		t.Fatalf("錯誤應說出是哪一步失敗,got %v", err)
	}
	if got.PastDue != 1 {
		t.Fatalf("已完成的轉移必須留在摘要裡,got %+v", got)
	}
	if got.Dispatched != 0 {
		t.Fatalf("中止後不得派送,got %d", got.Dispatched)
	}
	want := []string{"ExpireTrials", "MarkPastDue", "SuspendOverdue"}
	if order := log.order(); !slices.Equal(order, want) {
		t.Fatalf("失敗即中止(不得繼續後面的步驟): got %v want %v", order, want)
	}
}

// I-1:單一租戶的「產生期別」失敗(例:缺當期生效價目 → Task 5 會大聲失敗)**不得讓派送被跳過**。
// 派送派的是前面掃描(逾期／停用／取消到期)寫下的事件;跳過它,事件會持續積壓,
// 所有租戶的 subscription.suspended／expired 永遠到不了產品域,而 console 那頭早就顯示 suspended
// —— 帳務狀態與產品域長期不一致,而且症狀只有「排程回錯誤」。
func TestRunOnceKeepsDispatchingWhenPeriodOpeningFails(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	p := cron.Params{GraceDays: 7, LeadDays: 14, EventBatch: 100}
	ctx := context.Background()

	f := store.NewFakeBilling()
	f.PutSetting("system_actor_user_id", "7")
	// 方案 1 有月繳價目(43 開得出下一期);方案 2 沒有價目 → 該租戶的產生期別大聲失敗。
	f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})
	// 42:逾期未付(→past_due,它的 subscription.past_due 就是「必須照樣派送」的那一筆)。
	seedSub(f, 42, "active", "monthly", nil, now.Add(-time.Hour))
	// 43:服務中、期末在提前窗內,但**沒有價目** → 產生期別失敗。
	broken := f.PutSubscription(store.Subscription{
		CompanyID: 43, Status: "trialing", PlanID: 2, SeatCount: 3, BillingCycle: "monthly"})
	f.PutPeriod(store.Period{SubscriptionID: broken, PeriodNo: 1, Status: "open",
		PeriodStart: now.AddDate(0, -1, 0), PeriodEnd: now.AddDate(0, 0, 10),
		PlanID: 2, SeatCount: 3, AmountCents: 195000, Currency: "TWD"})
	// 44:服務中、期末在提前窗內且有價目 → 迴圈必須繼續跑完它(不能因 43 而放棄其他租戶)。
	seedSub(f, 44, "active", "monthly", nil, now.AddDate(0, 0, 10))

	deps, log, _, _ := newDeps(f)
	got, err := cron.RunOnce(ctx, deps, now, p)
	if err == nil {
		t.Fatal("產生期別的失敗必須回報")
	}
	if !strings.Contains(err.Error(), "產生期別(company=43)") {
		t.Fatalf("錯誤要說得出是哪一個租戶,got %v", err)
	}
	// ① 派送仍被呼叫(且在最後):42 的事件真的被認領走了。
	order := log.order()
	if order[len(order)-1] != "DispatchOnce" {
		t.Fatalf("派送必須照做且放最後,got %v", order)
	}
	if got.Dispatched != 2 {
		// 42 的 subscription.past_due ＋ 44 的 period.opened(44 的期別照開)。
		t.Fatalf("積壓的事件仍必須被派送,got %d", got.Dispatched)
	}
	if left, _ := f.UndispatchedEvents(ctx, 100); len(left) != 0 {
		t.Fatalf("積壓的事件必須被派送掉,剩 %d 筆", len(left))
	}
	// ② 部分完成的計數照舊:42 逾期、44 開出下一期(43 的失敗不影響它)。
	if got.PastDue != 1 || got.PeriodsOpened != 1 {
		t.Fatalf("計數應保留部分完成的那份: %+v", got)
	}
}

// 單飛:兩個執行重疊時,第二個**什麼都不做**且不算失敗(回零值摘要、Locked=false)。
// 少了它,兩趟會同時對同一批訂閱做狀態轉移與事件派送(排程沒有列鎖,CAS 靠這一層兜住)。
func TestRunGuardedSecondOverlappingRunDoesNothing(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	p := cron.Params{GraceDays: 7, LeadDays: 14, EventBatch: 100}
	f := store.NewFakeBilling()
	f.PutSetting("system_actor_user_id", "7")
	seedSub(f, 42, "active", "monthly", nil, now.Add(-time.Hour))

	lock := &fakeLocker{}
	deps, log, _, _ := newDeps(f)
	deps.Lock = lock
	deps.Store = &gateStore{inner: f, entered: make(chan struct{}), release: make(chan struct{})}

	first := make(chan cron.Summary, 1)
	firstErr := make(chan error, 1)
	go func() {
		s, err := cron.RunGuarded(context.Background(), deps, now, p)
		first <- s
		firstErr <- err
	}()

	gate := deps.Store.(*gateStore)
	<-gate.entered // 第一趟已進到掃描之中(鎖在手上)

	// 第二趟:取不到鎖 → 直接結束(不是錯誤)。
	// 以「呼叫數不變」斷言(不是寫死的步驟數):新增一個掃描步驟不該讓這條測試紅。
	callsBeforeSecondRun := len(log.order())
	second, err := cron.RunGuarded(context.Background(), deps, now, p)
	if err != nil {
		t.Fatalf("取不到鎖不是錯誤: %v", err)
	}
	if second != (cron.Summary{}) {
		t.Fatalf("第二個執行不得處理任何事: %+v", second)
	}
	if got := log.order(); len(got) != callsBeforeSecondRun {
		t.Fatalf("第二趟不得再呼叫任何掃描,got %v", got)
	}

	close(gate.release)
	if err := <-firstErr; err != nil {
		t.Fatalf("第一趟: %v", err)
	}
	if got := <-first; !got.Locked || got.PastDue != 1 {
		t.Fatalf("第一趟應取得鎖並完成轉移,got %+v", got)
	}
	if held, tries, releases := lock.state(); held || tries != 2 || releases != 1 {
		t.Fatalf("鎖應被取兩次、解一次且已釋放: held=%v tries=%d releases=%d", held, tries, releases)
	}
}

// panic 復原(C-09):排程是無人看管的行程,panic 裸奔會讓下一次觸發也一起死,而症狀只有
// CrashLoopBackOff(看起來像部署問題,不是「這一趟沒做成」)。復原後必須回報、**保留已完成的
// 計數**(爆掉的那趟在 log 裡要看得出「做到哪裡」),且鎖要放掉。
func TestRunGuardedRecoversPanic(t *testing.T) {
	lock := &fakeLocker{}
	// 逾期 2 筆之後在停用那一步 panic:兩筆已完成的是已落地的帳務事實,摘要不得被清成零。
	deps := cron.Deps{Billing: stubBilling{pastDue: 2, panicOn: "SuspendOverdue"}, Store: store.NewFakeBilling(), Lock: lock}

	got, err := cron.RunGuarded(context.Background(), deps, at(2026, time.October, 1, 3),
		cron.Params{GraceDays: 7, LeadDays: 14, EventBatch: 100})
	if err == nil {
		t.Fatal("panic 必須被復原並回報,不得靜默")
	}
	if !strings.Contains(err.Error(), "panic") {
		t.Fatalf("錯誤應看得出是 panic,got %v", err)
	}
	if !got.Locked || got.PastDue != 2 {
		t.Fatalf("panic 的那趟仍持有鎖,且已完成的計數必須留在摘要裡: %+v", got)
	}
	if held, _, releases := lock.state(); held || releases != 1 {
		t.Fatalf("panic 後鎖必須被釋放(否則下一趟永遠被擋): held=%v releases=%d", held, releases)
	}
}

// 取鎖本身 panic 也要被收斂(typed nil 的 Locker 之類):recover 必須涵蓋 TryLock 那一段,
// 否則排程會以 panic 收場 —— 那正是 RunGuarded 存在的理由。
func TestRunGuardedRecoversPanicInLock(t *testing.T) {
	_, err := cron.RunGuarded(context.Background(), cron.Deps{Lock: panicLocker{}},
		at(2026, time.October, 1, 3), cron.Params{})
	if err == nil || !strings.Contains(err.Error(), "panic") {
		t.Fatalf("TryLock 的 panic 必須被復原並回報,got %v", err)
	}
}

// 逾時(或取消)不得讓單飛鎖留在連線上(M-4):一趟卡住會一直握著鎖,後續每一趟都只印「跳過」且以 0
// 收場 —— 對外看起來就是排程靜默停擺。故 RunGuarded 以 context.WithoutCancel 解鎖,這裡釘住它。
func TestRunGuardedReleasesLockWhenContextExpires(t *testing.T) {
	lock := &fakeLocker{}
	deps := cron.Deps{Billing: stubBilling{waitForCtx: true}, Store: store.NewFakeBilling(), Lock: lock}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := cron.RunGuarded(ctx, deps, at(2026, time.October, 1, 3), cron.Params{})
	if err == nil {
		t.Fatal("逾時必須回報錯誤(不得當成正常完成)")
	}
	if held, _, releases := lock.state(); held || releases != 1 {
		t.Fatalf("逾時後鎖必須被釋放(否則下一趟永遠被擋): held=%v releases=%d", held, releases)
	}
}

// 沒有鎖就沒有單飛:寧可這一趟不跑,也不要兩個執行一起改帳。
func TestRunGuardedRefusesWithoutLock(t *testing.T) {
	_, err := cron.RunGuarded(context.Background(), cron.Deps{}, at(2026, time.October, 1, 3),
		cron.Params{})
	if err == nil {
		t.Fatal("缺單飛鎖必須拒絕執行")
	}
}

// 營運參數一律來自 platform.settings:缺席或值不合理回錯誤,**不得用猜的預設值**
// (無聲的錯誤寬限期在帳面上看不出來)。
func TestLoadParams(t *testing.T) {
	tests := []struct {
		name     string
		settings map[string]string
		want     cron.Params
		wantErr  string
	}{
		{name: "四鍵齊備", want: cron.Params{GraceDays: 7, LeadDays: 14, EventBatch: cron.EventBatch},
			settings: map[string]string{"grace_days": "7", "lead_days": " 14 "}},
		{name: "缺 grace_days", wantErr: "grace_days",
			settings: map[string]string{"lead_days": "14"}},
		{name: "缺 lead_days", wantErr: "lead_days",
			settings: map[string]string{"grace_days": "7"}},
		{name: "不是整數", wantErr: "不是整數",
			settings: map[string]string{"grace_days": "七天", "lead_days": "14"}},
		{name: "超出範圍", wantErr: "不合理",
			settings: map[string]string{"grace_days": "-1", "lead_days": "14"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := store.NewFakeBilling()
			for k, v := range tt.settings {
				f.PutSetting(k, v)
			}
			got, err := cron.LoadParams(context.Background(), f)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("應回錯誤(含 %q),got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadParams: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %+v want %+v", got, tt.want)
			}
		})
	}
}

// stubBilling 可程式化的假帳務:指定「哪一步 panic」與「逾期幾筆」(panic 路徑的斷言需要
// 「前面已完成幾筆」)。
type stubBilling struct {
	pastDue int
	panicOn string
	// waitForCtx 讓第一步一直等到 ctx 結束才回錯誤(模擬「一趟跑太久被逾時砍掉」)。
	waitForCtx bool
}

func (b stubBilling) call(name string) {
	if b.panicOn == name {
		panic("假造的 panic:" + name)
	}
}

func (b stubBilling) ExpireTrials(context.Context, time.Time, int) (int, error) {
	b.call("ExpireTrials")
	return 0, nil
}
func (b stubBilling) MarkPastDue(ctx context.Context, _ time.Time, _ int) (int, error) {
	b.call("MarkPastDue")
	if b.waitForCtx {
		<-ctx.Done()
		return 0, ctx.Err()
	}
	return b.pastDue, nil
}
func (b stubBilling) SuspendOverdue(context.Context, time.Time) (int, error) {
	b.call("SuspendOverdue")
	return 0, nil
}
func (b stubBilling) ExpireCancelled(context.Context, time.Time) (int, error) {
	b.call("ExpireCancelled")
	return 0, nil
}
func (b stubBilling) EnsureNextPeriod(context.Context, int, time.Time, int) (bool, error) {
	b.call("EnsureNextPeriod")
	return false, nil
}

// stubBilling 的 Store 接口已新增 TrialingSubscriptionsWithoutTrialEnd：
// stub 沒有 store，回空集合（不影響 panic／順序路徑的斷言）。
func (b stubBilling) TrialingSubscriptionsWithoutTrialEnd(context.Context) ([]store.Subscription, error) {
	return nil, nil
}

// 未結項 #18:排程的待收款含 G5 平台自營公司，而 console 的 ListReceivables 已排除它。
// 平台對自己開出的期別不是應收帳款 —— 兩處的「租戶」定義必須一致，否則排程摘要與
// console 待收款頁長期對不上（operator 會以為有一筆收不到的錢）。
func TestRunOnceReceivablesExcludesPlatformCompany(t *testing.T) {
	now := at(2026, time.October, 1, 3)
	p := cron.Params{GraceDays: 7, LeadDays: 14, EventBatch: 100}
	ctx := context.Background()

	f := store.NewFakeBilling()
	f.PutSetting("system_actor_user_id", "7")
	f.PutPlanPrice(1, "monthly", store.Price{BaseCents: 150000, SeatCents: 15000, Currency: "TWD"})
	// 42:一般租戶，逾期未付 → 待收款 1 筆。
	seedSub(f, 42, "active", "monthly", nil, now.Add(-time.Hour))
	// 9001:平台自營公司（由 PutPlatformCompany 標記，真 store 在 SQL 內排除）。
	// 即使有逾期未付的期別，也不得計入待收款（console 的 ListReceivables 用同一謂詞排除它）。
	f.PutPlatformCompany(9001)
	seedSub(f, 9001, "active", "monthly", nil, now.Add(-time.Hour))

	deps, _, _, _ := newDeps(f)
	s, err := cron.RunOnce(ctx, deps, now, p)
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if s.Receivables != 1 {
		t.Fatalf("待收款應排除平台自營公司，got %d", s.Receivables)
	}
}

// panicLocker:取鎖本身就 panic(測試 recover 是否涵蓋 TryLock 那一段)。
type panicLocker struct{}

func (panicLocker) TryLock(context.Context) (func(context.Context) error, bool, error) {
	panic("假造的鎖 panic")
}
