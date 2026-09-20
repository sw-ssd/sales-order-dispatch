// Task 6 的單元測試:consumer 把平台事件翻成產品域的公司狀態變更。
//
// 為什麼要在**介面層**做假實作(而不是像 billing 一樣假整個 store):consumer 的兩個依賴
// (讀事件／系統範圍交易)都已經窄到只剩介面,假實作因此能把「同一交易」與「條件式認領」這兩個
// 真正會出錯的語意**建模**出來,並讓測試看得見 consumer 的決策:
//
//   - fakeSystemTx.claimed 就是 platform.events.dispatched_at:認領是條件式的(同 id 只成功一次),
//     fn 失敗時整筆回滾(還原認領)—— 與 store.FakeBilling.WithTx 的 snapshot/restore 同慣例;
//   - fakeEvents.UndispatchedEvents 只回「還沒被認領」的事件,與 SQL 的 dispatched_at IS NULL 同義,
//     故同一批跑第二趟自然是空的(可重跑的前提);
//   - fakeSetter 記下每次呼叫與「呼叫當下是否在交易內」—— 狀態變更跑到交易外就是孤立狀態,
//     那是這條路徑最貴的失效(公司被凍結但事件沒派送、或反之)。
//
// 真 PG 才看得到的部分(認領的 RLS、commit/rollback、稽核欄位落地)在 consumer_integration_test.go。
package consumer_test

import (
	"context"
	"errors"
	"maps"
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/consumer"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// fakeEvents 假的事件來源:只回還沒被認領的事件(模擬 SQL 的 dispatched_at IS NULL 查詢)。
type fakeEvents struct {
	events []store.Event
	// claimed 與 fakeSystemTx 共用同一張 map:認領是**資料庫**的事,不是 consumer 的狀態。
	claimed map[int64]bool
	// stale 模擬「讀取之後、認領之前被別的執行搶先」的競態:這趟仍看到同一筆事件。
	stale bool

	actorID    int64
	actorErr   error
	actorCalls int
}

func (f *fakeEvents) UndispatchedEvents(context.Context, int) ([]store.Event, error) {
	var out []store.Event
	for _, e := range f.events {
		if f.stale || !f.claimed[e.ID] {
			out = append(out, e)
		}
	}
	return out, nil
}

func (f *fakeEvents) SystemActor(context.Context) (int64, error) {
	f.actorCalls++
	if f.actorErr != nil {
		return 0, f.actorErr
	}
	return f.actorID, nil
}

// fakeSystemTx 假的系統範圍交易:claimed 是那筆事件的 dispatched_at,fn 回錯誤即回滾認領。
type fakeSystemTx struct {
	claimed map[int64]bool
	claims  int
	openErr error
	// active 為 fn 執行中(供 setter 斷言狀態變更落在交易內)。
	active bool
}

func (f *fakeSystemTx) Run(ctx context.Context, fn func(context.Context, consumer.Tx) error) error {
	if f.openErr != nil {
		return f.openErr
	}
	saved := maps.Clone(f.claimed)
	f.active = true
	err := fn(ctx, fakeTx{f})
	f.active = false
	if err != nil {
		// 回滾:原地還原被認領的列(不換 map 物件 —— fakeEvents 共用同一張)。
		for id := range f.claimed {
			if !saved[id] {
				delete(f.claimed, id)
			}
		}
		return err
	}
	return nil
}

type fakeTx struct{ f *fakeSystemTx }

// Client 在單元測試裡沒有 ent:產品域入口是假的,不會用到它。
func (t fakeTx) Client() *ent.Client { return nil }

// Claim 條件式認領:同一筆只會成功一次(UPDATE ... WHERE dispatched_at IS NULL 的 0 列等價物)。
func (t fakeTx) Claim(_ context.Context, eventID int64) (bool, error) {
	if t.f.claimed[eventID] {
		return false, nil
	}
	t.f.claimed[eventID] = true
	t.f.claims++
	return true, nil
}

// statusCall 記錄一次產品域狀態變更呼叫;inTx 為呼叫當下是否在交易內。
type statusCall struct {
	companyID int
	status    company.Status
	reason    string
	actor     authz.Identity
	inTx      bool
}

type fakeSetter struct {
	tx    *fakeSystemTx
	calls []statusCall
	err   error
	// failFor > 0 時只讓該公司的變更失敗(模擬「這一家永遠不會成功」:例公司已軟刪除)。
	failFor int
}

func (f *fakeSetter) SetStatus(_ context.Context, _ *ent.Client, companyID int,
	status company.Status, reason string, actor authz.Identity) error {
	if f.err != nil || companyID == f.failFor {
		return errors.New("模擬產品域失敗")
	}
	f.calls = append(f.calls, statusCall{
		companyID: companyID, status: status, reason: reason, actor: actor, inTx: f.tx.active,
	})
	return nil
}

// newConsumer 組出受測的 consumer:假的來源(actorID=7)＋假的系統範圍交易＋假的產品域入口。
func newConsumer(t *testing.T, events ...store.Event) (*consumer.Consumer, *fakeEvents, *fakeSystemTx, *fakeSetter) {
	t.Helper()
	tx := &fakeSystemTx{claimed: map[int64]bool{}}
	src := &fakeEvents{events: events, claimed: tx.claimed, actorID: 7}
	setter := &fakeSetter{tx: tx}
	return consumer.New(src, tx, setter), src, tx, setter
}

// ① subscription.suspended → 凍結公司,且在**交易內**呼叫產品域的唯一入口並認領事件。
func TestDispatchOnceSuspendsCompany(t *testing.T) {
	c, _, tx, setter := newConsumer(t, store.Event{ID: 1, EventType: "subscription.suspended",
		Payload: []byte(`{"company_id":42,"reason":"overdue"}`)})

	n, err := c.DispatchOnce(context.Background(), 100)
	if err != nil || n != 1 {
		t.Fatalf("應派送 1 筆: n=%d err=%v", n, err)
	}
	if len(setter.calls) != 1 {
		t.Fatalf("應呼叫一次狀態入口,got %d", len(setter.calls))
	}
	got := setter.calls[0]
	if got.companyID != 42 || got.status != company.StatusSuspended {
		t.Fatalf("應凍結公司 42,got %+v", got)
	}
	if !strings.Contains(got.reason, "逾期未付") {
		t.Fatalf("凍結原因應說明逾期未付(稽核要說得出為什麼),got %q", got.reason)
	}
	if !got.inTx {
		t.Fatal("狀態變更必須在認領的同一交易內(交易外呼叫會留下孤立狀態)")
	}
	if got.actor.UserID != "7" {
		t.Fatalf("稽核主體應為 platform.settings 的系統 actor(users.id=7),got %q", got.actor.UserID)
	}
	if !tx.claimed[1] {
		t.Fatal("事件 1 應被認領(dispatched_at)")
	}
}

// ② subscription.expired(G7)→ 凍結。suspended 也涵蓋「已取消且期末已過」:漏了這一列,
// 排程發得出事件、consumer 卻把它當未知型別只認領 —— G7 靜默失效,而帳與事件都看起來正常。
func TestDispatchOnceSuspendsExpiredCancelledSubscription(t *testing.T) {
	c, _, tx, setter := newConsumer(t, store.Event{ID: 9, EventType: "subscription.expired",
		Payload: []byte(`{"company_id":42,"subscription_id":3,"reason":"cancelled_at_period_end"}`)})

	n, err := c.DispatchOnce(context.Background(), 100)
	if err != nil || n != 1 {
		t.Fatalf("expired 應派送 1 筆: n=%d err=%v", n, err)
	}
	if len(setter.calls) != 1 || setter.calls[0].status != company.StatusSuspended {
		t.Fatalf("expired 必須凍結公司(G7),got %+v", setter.calls)
	}
	if got := setter.calls[0]; got.companyID != 42 || !strings.Contains(got.reason, "期末") {
		t.Fatalf("expired 應凍結公司 42 並說明期末已過,got %+v", got)
	}
	if !tx.claimed[9] {
		t.Fatal("事件 9 應被認領")
	}
}

// ③ subscription.reactivated → 復原為 active。
func TestDispatchOnceReactivatesCompany(t *testing.T) {
	c, _, tx, setter := newConsumer(t, store.Event{ID: 2, EventType: "subscription.reactivated",
		Payload: []byte(`{"company_id":42,"from":"suspended"}`)})

	if _, err := c.DispatchOnce(context.Background(), 100); err != nil {
		t.Fatalf("派送: %v", err)
	}
	if len(setter.calls) != 1 || setter.calls[0].status != company.StatusActive {
		t.Fatalf("應以 active 復原,got %+v", setter.calls)
	}
	if !strings.Contains(setter.calls[0].reason, "補款復原") {
		t.Fatalf("復原原因應說明補款,got %q", setter.calls[0].reason)
	}
	if !tx.claimed[2] {
		t.Fatal("事件 2 應被認領")
	}
}

// ④ 同一批事件跑第二趟:沒有可派送的事件 → 不再認領、不再寫狀態(排程每日重跑的安全網)。
func TestDispatchOnceIsIdempotentAcrossRuns(t *testing.T) {
	c, _, tx, setter := newConsumer(t, store.Event{ID: 1, EventType: "subscription.suspended",
		Payload: []byte(`{"company_id":42}`)})

	first, err := c.DispatchOnce(context.Background(), 100)
	if err != nil || first != 1 {
		t.Fatalf("第一趟應派送 1 筆: n=%d err=%v", first, err)
	}
	second, err := c.DispatchOnce(context.Background(), 100)
	if err != nil || second != 0 {
		t.Fatalf("第二趟應無事可做: n=%d err=%v", second, err)
	}
	if tx.claims != 1 {
		t.Fatalf("同一筆事件只應被認領一次,got %d", tx.claims)
	}
	if len(setter.calls) != 1 {
		t.Fatalf("第二趟不得再變更狀態(也不得留第二筆稽核),got %d 次呼叫", len(setter.calls))
	}
}

// ⑤ 事件已被別的執行搶先認領(條件式 UPDATE 0 列)→ 跳過:不報錯、不算派送、不重複副作用。
// 這是「重跑／補跑／單飛鎖失效」下唯一真正擋得住重複副作用的閘門(讀取與認領之間一定有縫)。
func TestDispatchOnceSkipsEventClaimedByAnotherRun(t *testing.T) {
	c, src, tx, setter := newConsumer(t, store.Event{ID: 1, EventType: "subscription.suspended",
		Payload: []byte(`{"company_id":42}`)})
	src.stale = true // 這趟仍看得到它(別的執行在讀取之後才認領)
	tx.claimed[1] = true

	n, err := c.DispatchOnce(context.Background(), 100)
	if err != nil {
		t.Fatalf("被搶先認領不是錯誤(排程不該因此紅): %v", err)
	}
	if n != 0 {
		t.Fatalf("被搶先認領不得算成本趟派送,got %d", n)
	}
	if len(setter.calls) != 0 {
		t.Fatalf("不得產生第二個副作用,got %+v", setter.calls)
	}
}

// ⑥ 沒有產品域動作的型別(past_due 在寬限期內仍提供服務、period.* 目前只供通知)→
// 只認領、不失敗:不認領的話排程每趟都會重掃同一筆。
func TestDispatchOnceClaimsUnmappedEventTypesWithoutSideEffects(t *testing.T) {
	c, _, tx, setter := newConsumer(t,
		store.Event{ID: 1, EventType: "subscription.past_due", Payload: []byte(`{"company_id":42}`)},
		store.Event{ID: 2, EventType: "period.opened", Payload: []byte(`{"company_id":42}`)})

	n, err := c.DispatchOnce(context.Background(), 100)
	if err != nil {
		t.Fatalf("未對應型別不得讓一趟失敗: %v", err)
	}
	if n != 2 {
		t.Fatalf("兩筆都應被認領(否則每趟重掃),got %d", n)
	}
	if len(setter.calls) != 0 {
		t.Fatalf("未對應型別不得有副作用,got %+v", setter.calls)
	}
	if !tx.claimed[1] || !tx.claimed[2] {
		t.Fatal("兩筆都應標記為已派送")
	}
}

// ⑦ 產品域寫入失敗 → 整筆回滾(認領一起還原):不得留下「狀態沒改、事件卻已派送」的孤兒,
// 且修好之後下一趟必須重試成功。
func TestDispatchOnceRollsBackClaimWhenProductDomainFails(t *testing.T) {
	c, _, tx, setter := newConsumer(t, store.Event{ID: 1, EventType: "subscription.suspended",
		Payload: []byte(`{"company_id":42}`)})
	setter.err = errors.New("模擬產品域失敗")

	if _, err := c.DispatchOnce(context.Background(), 100); err == nil {
		t.Fatal("產品域失敗必須往外傳(排程要看得見)")
	}
	if tx.claimed[1] {
		t.Fatal("失敗的事件不得留成已派送(否則公司永遠不會被凍結)")
	}

	setter.err = nil
	n, err := c.DispatchOnce(context.Background(), 100)
	if err != nil || n != 1 {
		t.Fatalf("修好後下一趟應重試成功: n=%d err=%v", n, err)
	}
	if len(setter.calls) != 1 || setter.calls[0].companyID != 42 {
		t.Fatalf("重試應真的呼叫狀態入口,got %+v", setter.calls)
	}
}

// ⑧ payload 壞掉(缺 company_id／不是物件)→ 錯誤往外、不認領:留在待派送清單供人看,
// 不得靜默跳過(那筆事件的副作用就此消失)。
func TestDispatchOnceRejectsEventWithoutCompanyID(t *testing.T) {
	for name, payload := range map[string]string{
		"缺 company_id": `{"reason":"overdue"}`,
		"不是物件":         `[]`,
	} {
		t.Run(name, func(t *testing.T) {
			c, _, tx, setter := newConsumer(t, store.Event{ID: 1,
				EventType: "subscription.suspended", Payload: []byte(payload)})

			if _, err := c.DispatchOnce(context.Background(), 100); err == nil {
				t.Fatal("壞 payload 必須往外報錯")
			}
			if tx.claimed[1] {
				t.Fatal("壞 payload 不得被認領")
			}
			if len(setter.calls) != 0 {
				t.Fatalf("壞 payload 不得有副作用,got %+v", setter.calls)
			}
		})
	}
}

// ⑩ 佇列頭部有一筆**永遠不會成功**的事件時,不得堵住後面的事件:單筆失敗只記下錯誤,
// 迴圈跑完再一次往外傳(errors.Join)。事件依 id 排序、每趟從第一筆未派送者開始,若一失敗就 return,
// 後面所有租戶的凍結／復原(含 G7 的期末停用)永遠不會被處理,而症狀只有「排程回錯誤」。
func TestDispatchOnceContinuesPastPermanentlyFailingEvent(t *testing.T) {
	c, _, tx, setter := newConsumer(t,
		store.Event{ID: 1, EventType: "subscription.suspended", Payload: []byte(`{"company_id":42}`)},
		store.Event{ID: 2, EventType: "subscription.suspended", Payload: []byte(`{"company_id":43}`)})
	setter.failFor = 42 // 公司 42 永遠失敗(例:已軟刪除 → SetCompanyStatus 一律 NotFound)

	n, err := c.DispatchOnce(context.Background(), 100)
	if err == nil {
		t.Fatal("沒做成的事件必須回報(排程要看得見),不得靜默")
	}
	if n != 1 {
		t.Fatalf("後面的公司 43 仍必須被派送,got %d", n)
	}
	if len(setter.calls) != 1 || setter.calls[0].companyID != 43 {
		t.Fatalf("只有公司 43 該被變更,got %+v", setter.calls)
	}
	if tx.claimed[1] {
		t.Fatal("失敗的事件不得被認領(不得吞掉凍結)")
	}
	if !tx.claimed[2] {
		t.Fatal("公司 43 的事件應被認領")
	}

	// 下趟重試頭部那筆:仍然失敗、仍然不阻塞、也不重複副作用(43 已派送故不再出現)。
	n, err = c.DispatchOnce(context.Background(), 100)
	if err == nil {
		t.Fatal("永久失敗的事件下趟仍應回報錯誤")
	}
	if n != 0 || len(setter.calls) != 1 {
		t.Fatalf("重試不得重複副作用: n=%d calls=%d", n, len(setter.calls))
	}
}

// ⑪ payload 壞掉的那筆同樣不得堵住後面(它自己留在待派送清單,後面照常派送)。
func TestDispatchOnceContinuesPastBrokenPayload(t *testing.T) {
	c, _, tx, setter := newConsumer(t,
		store.Event{ID: 1, EventType: "subscription.suspended", Payload: []byte(`{}`)},
		store.Event{ID: 2, EventType: "subscription.suspended", Payload: []byte(`{"company_id":43}`)})

	n, err := c.DispatchOnce(context.Background(), 100)
	if err == nil {
		t.Fatal("壞 payload 必須回報")
	}
	if n != 1 || len(setter.calls) != 1 || setter.calls[0].companyID != 43 {
		t.Fatalf("壞 payload 只該擋住自己: n=%d calls=%+v", n, setter.calls)
	}
	if tx.claimed[1] {
		t.Fatal("壞 payload 不得被認領")
	}
}

// ⑨ 系統 actor 只在真的要寫產品域時才解析:未設定的 platform.settings 不該讓「整批都未對應型別」
// 的一趟失敗(那會讓那些事件每趟被重掃);反之要寫狀態時沒有 actor 就必須失敗(fail-closed:
// 稽核沒有主體,寧可不動)。
func TestDispatchOnceResolvesSystemActorOnlyForMappedEvents(t *testing.T) {
	t.Run("只有未對應型別 → 不必解析 actor", func(t *testing.T) {
		c, src, _, setter := newConsumer(t, store.Event{ID: 1,
			EventType: "period.opened", Payload: []byte(`{}`)})
		src.actorErr = errors.New("platform.settings 沒有 system_actor_user_id")

		n, err := c.DispatchOnce(context.Background(), 100)
		if err != nil || n != 1 {
			t.Fatalf("未對應型別應只認領: n=%d err=%v", n, err)
		}
		if src.actorCalls != 0 {
			t.Fatalf("不需要 actor 就不該查設定,got %d 次", src.actorCalls)
		}
		if len(setter.calls) != 0 {
			t.Fatalf("不得有副作用,got %+v", setter.calls)
		}
	})

	t.Run("要寫產品域但沒有 actor → 失敗且不認領", func(t *testing.T) {
		c, src, tx, setter := newConsumer(t, store.Event{ID: 1,
			EventType: "subscription.suspended", Payload: []byte(`{"company_id":42}`)})
		src.actorErr = errors.New("platform.settings 沒有 system_actor_user_id")

		if _, err := c.DispatchOnce(context.Background(), 100); err == nil {
			t.Fatal("沒有稽核主體不得變更公司狀態")
		}
		if tx.claimed[1] {
			t.Fatal("不得認領(留待 actor 修好後重試)")
		}
		if len(setter.calls) != 0 {
			t.Fatalf("不得有副作用,got %+v", setter.calls)
		}
	})
}
