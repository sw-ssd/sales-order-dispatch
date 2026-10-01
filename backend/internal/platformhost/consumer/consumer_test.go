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
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
	"github.com/salesorder/sales-order-1.0/backend/internal/platformhost/consumer"
)

// fakeEvents 假的事件來源:只回還沒被認領的事件(模擬 SQL 的 dispatched_at IS NULL 查詢)。
// emitted 記錄 consumer/worker 在交易內寫出的 outbox 事件(模擬 platform.events 的 INSERT),
// 測試據此斷言「狀態變更請求」確實被發出(而非直接改 companies)。
type fakeEvents struct {
	events  []store.Event // 全部事件(含 consumer 主迴圈發出的 company.status_changed outbox)。
	claimed map[int64]bool
	// emitted 為 consumer 主迴圈發出的 company.status_changed outbox(測試斷言用;與 events 同步)。
	emitted []store.Event
	stale   bool

	actorID    int64
	actorErr   error
	actorCalls int
}

func (f *fakeEvents) UndispatchedEvents(context.Context, int) ([]store.Event, error) {
	var out []store.Event
	for _, e := range f.events {
		if e.EventType == "company.status_changed" {
			continue // outbox 由 worker 獨佔,主迴圈不得回頭撿到(與 postgres 一致)。
		}
		if f.stale || !f.claimed[e.ID] {
			out = append(out, e)
		}
	}
	return out, nil
}

// UndispatchedCompanyEvents 回尚未派送的 company.status_changed outbox(worker 用)。
func (f *fakeEvents) UndispatchedCompanyEvents(context.Context, int) ([]store.Event, error) {
	var out []store.Event
	for _, e := range f.events {
		if e.EventType != "company.status_changed" {
			continue
		}
		if !f.claimed[e.ID] {
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
	events  *fakeEvents
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

// Emit 記錄 consumer 發出的 company.status_changed outbox(模擬 platform.events 的 INSERT),
// 與認領同一個交易(由 fakeSystemTx.Run 的回滾語意保證)。
func (t fakeTx) Emit(_ context.Context, aggregateID uuid.UUID, _ string, payload []byte) error {
	ev := store.Event{
		ID:            int64(len(t.f.events.events) + 1),
		AggregateType: "company",
		AggregateID:   aggregateID,
		EventType:     "company.status_changed",
		Payload:       slices.Clone(payload),
	}
	t.f.events.events = append(t.f.events.events, ev)
	t.f.events.emitted = append(t.f.events.emitted, ev)
	return nil
}

// Claim 條件式認領:同一筆只會成功一次(UPDATE ... WHERE dispatched_at IS NULL 的 0 列等價物)。
func (t fakeTx) Claim(_ context.Context, eventID int64) (bool, error) {
	if t.f.claimed[eventID] {
		return false, nil
	}
	t.f.claimed[eventID] = true
	t.f.claims++
	return true, nil
}

// newConsumer 組出受測的 consumer:假的來源(actorID=7)＋假的系統範圍交易。
// p3 outbox 反轉後 consumer 不再直接呼叫產品域入口,狀態變更改由 worker 消費發出的
// company.status_changed outbox 執行;故本測試只驗證 outbox 是否正確發出(見 fakeEvents.emitted)。
func newConsumer(t *testing.T, events ...store.Event) (*consumer.Consumer, *fakeEvents, *fakeSystemTx) {
	t.Helper()
	src := &fakeEvents{events: events, actorID: 7}
	tx := &fakeSystemTx{events: src, claimed: map[int64]bool{}}
	// consumer.New 第三參數(CompanyStatusSetter)在 outbox 反轉後未被 dispatch 使用,傳 ProductDomain{} 佔位。
	return consumer.New(src, tx, consumer.ProductDomain{}), src, tx
}

// emittedStatusChanged 取 fakeEvents 發出的第 idx 筆 company.status_changed outbox 的 payload 欄位。
func emittedStatusChanged(t *testing.T, src *fakeEvents, idx int) (companyID int, status, reason string) {
	t.Helper()
	if idx >= len(src.emitted) {
		t.Fatalf("應有第 %d 筆 company.status_changed outbox,實際 %d 筆: %+v", idx, len(src.emitted), src.emitted)
	}
	var p struct {
		CompanyID int    `json:"company_id"`
		Status    string `json:"status"`
		Reason    string `json:"reason"`
	}
	if err := json.Unmarshal(src.emitted[idx].Payload, &p); err != nil {
		t.Fatalf("outbox payload 不是合法 JSON: %v", err)
	}
	return p.CompanyID, p.Status, p.Reason
}

// ① subscription.suspended → 發 company.status_changed outbox(凍結),且認領事件。
func TestDispatchOnceSuspendsCompany(t *testing.T) {
	c, src, tx := newConsumer(t, store.Event{ID: 1, EventType: "subscription.suspended",
		Payload: []byte(`{"company_id":42,"reason":"overdue"}`)})

	n, err := c.DispatchOnce(context.Background(), 100)
	if err != nil || n != 1 {
		t.Fatalf("應派送 1 筆: n=%d err=%v", n, err)
	}
	if len(src.emitted) != 1 {
		t.Fatalf("應發出 1 筆 company.status_changed outbox,got %d", len(src.emitted))
	}
	companyID, status, reason := emittedStatusChanged(t, src, 0)
	if companyID != 42 || status != string(company.StatusSuspended) {
		t.Fatalf("應凍結公司 42,got companyID=%d status=%q", companyID, status)
	}
	if !strings.Contains(reason, "逾期未付") {
		t.Fatalf("凍結原因應說明逾期未付(稽核要說得出為什麼),got %q", reason)
	}
	if !tx.claimed[1] {
		t.Fatal("事件 1 應被認領(dispatched_at)")
	}
}

// ② subscription.expired(G7)→ 發凍結 outbox。漏了它,consumer 會把它當未知型別只認領 ——
// G7 靜默失效,而帳與事件都看起來正常。
func TestDispatchOnceSuspendsExpiredCancelledSubscription(t *testing.T) {
	c, src, tx := newConsumer(t, store.Event{ID: 9, EventType: "subscription.expired",
		Payload: []byte(`{"company_id":42,"subscription_id":3,"reason":"cancelled_at_period_end"}`)})

	n, err := c.DispatchOnce(context.Background(), 100)
	if err != nil || n != 1 {
		t.Fatalf("expired 應派送 1 筆: n=%d err=%v", n, err)
	}
	if len(src.emitted) != 1 {
		t.Fatalf("expired 必須發凍結 outbox(G7),got %d 筆", len(src.emitted))
	}
	companyID, status, reason := emittedStatusChanged(t, src, 0)
	if companyID != 42 || status != string(company.StatusSuspended) {
		t.Fatalf("expired 應凍結公司 42,got companyID=%d status=%q", companyID, status)
	}
	if !strings.Contains(reason, "期末") {
		t.Fatalf("expired 應說明期末已過,got %q", reason)
	}
	if !tx.claimed[9] {
		t.Fatal("事件 9 應被認領")
	}
}

// ③ subscription.reactivated → 發復原 active outbox。
func TestDispatchOnceReactivatesCompany(t *testing.T) {
	c, src, tx := newConsumer(t, store.Event{ID: 2, EventType: "subscription.reactivated",
		Payload: []byte(`{"company_id":42,"from":"suspended"}`)})

	if _, err := c.DispatchOnce(context.Background(), 100); err != nil {
		t.Fatalf("派送: %v", err)
	}
	if len(src.emitted) != 1 {
		t.Fatalf("應發出 1 筆 outbox,got %d", len(src.emitted))
	}
	companyID, status, reason := emittedStatusChanged(t, src, 0)
	if companyID != 42 || status != string(company.StatusActive) {
		t.Fatalf("應以 active 復原,got companyID=%d status=%q", companyID, status)
	}
	if !strings.Contains(reason, "補款復原") {
		t.Fatalf("復原原因應說明補款,got %q", reason)
	}
	if !tx.claimed[2] {
		t.Fatal("事件 2 應被認領")
	}
}

// ④ 同一批事件跑第二趟:沒有可派送的事件 → 不再認領、不再發 outbox(排程每日重跑的安全網)。
func TestDispatchOnceIsIdempotentAcrossRuns(t *testing.T) {
	c, src, tx := newConsumer(t, store.Event{ID: 1, EventType: "subscription.suspended",
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
	if len(src.emitted) != 1 {
		t.Fatalf("第二趟不得再發 outbox,got %d 筆", len(src.emitted))
	}
}

// ⑤ 事件已被別的執行搶先認領(條件式 UPDATE 0 列)→ 跳過:不報錯、不算派送、不重複副作用。
func TestDispatchOnceSkipsEventClaimedByAnotherRun(t *testing.T) {
	c, src, tx := newConsumer(t, store.Event{ID: 1, EventType: "subscription.suspended",
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
	if len(src.emitted) != 0 {
		t.Fatalf("不得發出第二個 outbox,got %+v", src.emitted)
	}
}

// ⑥ 沒有產品域動作的型別(past_due 在寬限期內仍提供服務、period.* 目前只供通知)→
// 只認領、不發 outbox:不認領的話排程每趟都會重掃同一筆。
func TestDispatchOnceClaimsUnmappedEventTypesWithoutSideEffects(t *testing.T) {
	c, src, tx := newConsumer(t,
		store.Event{ID: 1, EventType: "subscription.past_due", Payload: []byte(`{"company_id":42}`)},
		store.Event{ID: 2, EventType: "period.opened", Payload: []byte(`{"company_id":42}`)})

	n, err := c.DispatchOnce(context.Background(), 100)
	if err != nil {
		t.Fatalf("未對應型別不得讓一趟失敗: %v", err)
	}
	if n != 2 {
		t.Fatalf("兩筆都應被認領(否則每趟重掃),got %d", n)
	}
	if len(src.emitted) != 0 {
		t.Fatalf("未對應型別不得發 outbox,got %+v", src.emitted)
	}
	if !tx.claimed[1] || !tx.claimed[2] {
		t.Fatal("兩筆都應標記為已派送")
	}
}

// ⑧ payload 壞掉(缺 company_id／不是物件)→ 錯誤往外、不認領、不發 outbox:留在待派送清單供人看,
// 不得靜默跳過(那筆事件的副作用就此消失)。
func TestDispatchOnceRejectsEventWithoutCompanyID(t *testing.T) {
	for name, payload := range map[string]string{
		"缺 company_id": `{"reason":"overdue"}`,
		"不是物件":         `[]`,
	} {
		t.Run(name, func(t *testing.T) {
			c, src, tx := newConsumer(t, store.Event{ID: 1,
				EventType: "subscription.suspended", Payload: []byte(payload)})

			if _, err := c.DispatchOnce(context.Background(), 100); err == nil {
				t.Fatal("壞 payload 必須往外報錯")
			}
			if tx.claimed[1] {
				t.Fatal("壞 payload 不得被認領")
			}
			if len(src.emitted) != 0 {
				t.Fatalf("壞 payload 不得發 outbox,got %+v", src.emitted)
			}
		})
	}
}

// ⑪ payload 壞掉的那筆同樣不得堵住後面(它自己留在待派送清單,後面照常派送)。
func TestDispatchOnceContinuesPastBrokenPayload(t *testing.T) {
	c, src, tx := newConsumer(t,
		store.Event{ID: 1, EventType: "subscription.suspended", Payload: []byte(`{}`)},
		store.Event{ID: 2, EventType: "subscription.suspended", Payload: []byte(`{"company_id":43}`)})

	n, err := c.DispatchOnce(context.Background(), 100)
	if err == nil {
		t.Fatal("壞 payload 必須回報")
	}
	if n != 1 || len(src.emitted) != 1 {
		t.Fatalf("壞 payload 只該擋住自己: n=%d emitted=%+v", n, src.emitted)
	}
	if companyID, _, _ := emittedStatusChanged(t, src, 0); companyID != 43 {
		t.Fatalf("正確的事件應發 outbox 給公司 43,got companyID=%d", companyID)
	}
	if tx.claimed[1] {
		t.Fatal("壞 payload 不得被認領")
	}
	if !tx.claimed[2] {
		t.Fatal("公司 43 的事件應被認領")
	}
}

// ⑨ 系統 actor 只在真的要發 outbox 時才解析:未設定的 platform.settings 不該讓「整批都未對應型別」
// 的一趟失敗(那會讓那些事件每趟被重掃);反之要發 outbox 時沒有 actor 就必須失敗(fail-closed:
// 稽核沒有主體,寧可不動)。
func TestDispatchOnceResolvesSystemActorOnlyForMappedEvents(t *testing.T) {
	t.Run("只有未對應型別 → 不必解析 actor", func(t *testing.T) {
		c, src, _ := newConsumer(t, store.Event{ID: 1,
			EventType: "period.opened", Payload: []byte(`{}`)})
		src.actorErr = errors.New("platform.settings 沒有 system_actor_user_id")

		n, err := c.DispatchOnce(context.Background(), 100)
		if err != nil || n != 1 {
			t.Fatalf("未對應型別應只認領: n=%d err=%v", n, err)
		}
		if src.actorCalls != 0 {
			t.Fatalf("不需要 actor 就不該查設定,got %d 次", src.actorCalls)
		}
		if len(src.emitted) != 0 {
			t.Fatalf("不得發 outbox,got %+v", src.emitted)
		}
	})

	t.Run("要發 outbox 但沒有 actor → 失敗且不認領", func(t *testing.T) {
		c, src, tx := newConsumer(t, store.Event{ID: 1,
			EventType: "subscription.suspended", Payload: []byte(`{"company_id":42}`)})
		src.actorErr = errors.New("platform.settings 沒有 system_actor_user_id")

		if _, err := c.DispatchOnce(context.Background(), 100); err == nil {
			t.Fatal("沒有稽核主體不得發 outbox 變更公司狀態")
		}
		if tx.claimed[1] {
			t.Fatal("不得認領(留待 actor 修好後重試)")
		}
		if len(src.emitted) != 0 {
			t.Fatalf("不得發 outbox,got %+v", src.emitted)
		}
	})
}

// 未結項 #16:actor 解析失敗時，每個 mapped 事件各 append 一條同根因錯誤 → errors.Join 訊息
// O(N) 膨脹（僅 log 噪音）。同一趟內平台設定不會自己變好，actor 失敗只應記一次。
func TestDispatchOnceDedupesActorFailure(t *testing.T) {
	events := []store.Event{
		{ID: 1, EventType: "subscription.suspended", Payload: []byte(`{"company_id":42,"reason":"overdue"}`)},
		{ID: 2, EventType: "subscription.suspended", Payload: []byte(`{"company_id":43,"reason":"overdue"}`)},
		{ID: 3, EventType: "subscription.suspended", Payload: []byte(`{"company_id":44,"reason":"overdue"}`)},
	}
	c, src, _ := newConsumer(t, events...)
	src.actorErr = errors.New("模擬平台設定缺失")

	_, err := c.DispatchOnce(context.Background(), 10)
	if err == nil {
		t.Fatal("actor 失敗應回錯誤")
	}
	// 同根因只記一次：外層只有一條「事件 1 …無法派送」（定位用），不因三筆事件變三條。
	joined := err.Error()
	if n := strings.Count(joined, "無法派送"); n != 1 {
		t.Fatalf("同根因錯誤應只記一次，got %d 次: %v", n, err)
	}
	if !strings.Contains(joined, "事件 1(subscription.suspended)") {
		t.Fatalf("應記下第一個事件的 id 與型別（定位用）: %v", err)
	}
	if got := src.actorCalls; got != 1 {
		t.Fatalf("同一趟 actor 只應解析一次，got %d 次", got)
	}
	if len(src.emitted) != 0 {
		t.Fatalf("actor 缺失時不得發任何 outbox,got %+v", src.emitted)
	}
}
