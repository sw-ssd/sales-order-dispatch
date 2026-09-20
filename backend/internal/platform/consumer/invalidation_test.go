// Task 8：consumer 派送事件時的**權益快取失效**。
//
// consumer 是訂閱狀態的第三個寫入路徑（前兩個是 T9 的平台 RPC 與 billing 的掃描式轉移）：它把
// subscription.suspended／expired／reactivated 翻成公司狀態變更，而公司在凍結／復原之後，
// 該租戶的權益快取（ent:{companyID}）必須跟著失效 —— 否則 console 顯示「已凍結」而配額判定
// 還在用舊快照放行。
//
// 邊界與 billing 同一條：**只對真的改了公司狀態的事件（mapped 且成功認領）失效**。
// 未對應型別（period.opened…）只被認領、不動公司狀態；狀態變更失敗的那筆整筆回滾 ——
// 兩者都不得刪快取。
package consumer_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// recordingCache 只記錄失效（Delete）呼叫；其餘 Cache 方法為 no-op。
type recordingCache struct{ deleted []string }

func (c *recordingCache) Get(context.Context, string) ([]byte, bool, error) { return nil, false, nil }
func (c *recordingCache) Set(context.Context, string, []byte, time.Duration) error {
	return nil
}
func (c *recordingCache) Delete(_ context.Context, key string) error {
	c.deleted = append(c.deleted, key)
	return nil
}

// subscription.suspended（凍結）→ 交易提交後失效該租戶的快取。
func TestDispatchInvalidatesCacheForSuspendedEvent(t *testing.T) {
	c, _, _, _ := newConsumer(t, store.Event{ID: 1, AggregateType: "subscription", AggregateID: 5,
		EventType: "subscription.suspended", Payload: []byte(`{"company_id":42}`)})
	cache := &recordingCache{}
	c.WithCache(cache)

	if _, err := c.DispatchOnce(context.Background(), 10); err != nil {
		t.Fatalf("派送: %v", err)
	}
	assertInvalidated(t, cache, "ent:42")
}

// subscription.expired（G7：取消且期末已過 → 凍結）同樣要失效。
func TestDispatchInvalidatesCacheForExpiredEvent(t *testing.T) {
	c, _, _, _ := newConsumer(t, store.Event{ID: 1, AggregateType: "subscription", AggregateID: 5,
		EventType: "subscription.expired", Payload: []byte(`{"company_id":43}`)})
	cache := &recordingCache{}
	c.WithCache(cache)

	if _, err := c.DispatchOnce(context.Background(), 10); err != nil {
		t.Fatalf("派送: %v", err)
	}
	assertInvalidated(t, cache, "ent:43")
}

// 未對應型別（period.opened）：只認領、不動公司狀態 → 不得失效。
//
// 為什麼要擋：這類事件每期每租戶都會產生，若也失效，等於每開一期就把該租戶的快取清一次 ——
// 快取被自己的副作用打穿，而 log 看起來完全正常。
func TestDispatchDoesNotInvalidateForUnmappedEvent(t *testing.T) {
	c, _, _, _ := newConsumer(t, store.Event{ID: 1, AggregateType: "subscription", AggregateID: 5,
		EventType: "period.opened", Payload: []byte(`{"company_id":42}`)})
	cache := &recordingCache{}
	c.WithCache(cache)

	if _, err := c.DispatchOnce(context.Background(), 10); err != nil {
		t.Fatalf("派送: %v", err)
	}
	if len(cache.deleted) != 0 {
		t.Fatalf("未改公司狀態的事件不得失效快取，got %v", cache.deleted)
	}
}

// 產品域失敗（整筆回滾、事件未被認領）→ 不得失效；且要能重試。
func TestDispatchFailureDoesNotInvalidateCache(t *testing.T) {
	c, _, _, setter := newConsumer(t, store.Event{ID: 1, AggregateType: "subscription",
		AggregateID: 5, EventType: "subscription.suspended", Payload: []byte(`{"company_id":42}`)})
	setter.err = errors.New("模擬產品域失敗")
	cache := &recordingCache{}
	c.WithCache(cache)

	if _, err := c.DispatchOnce(context.Background(), 10); err == nil {
		t.Fatal("產品域失敗必須回報（errors.Join），不得靜默")
	}
	if len(cache.deleted) != 0 {
		t.Fatalf("失敗的事件不得失效快取，got %v", cache.deleted)
	}
}

// 沒接上快取（cache=nil）時照常派送：可選依賴的意義。
func TestConsumerWorksWithoutCache(t *testing.T) {
	c, _, _, setter := newConsumer(t, store.Event{ID: 1, AggregateType: "subscription",
		AggregateID: 5, EventType: "subscription.suspended", Payload: []byte(`{"company_id":42}`)})

	if n, err := c.DispatchOnce(context.Background(), 10); err != nil || n != 1 {
		t.Fatalf("未接快取時派送應照常運作: n=%d err=%v", n, err)
	}
	if len(setter.calls) != 1 {
		t.Fatalf("公司狀態仍應被變更一次，got %v", setter.calls)
	}
}

func assertInvalidated(t *testing.T, c *recordingCache, want ...string) {
	t.Helper()
	if len(c.deleted) != len(want) {
		t.Fatalf("失效的鍵不符: got %v, want %v", c.deleted, want)
	}
	for i, k := range want {
		if c.deleted[i] != k {
			t.Fatalf("失效的鍵不符: got %v, want %v", c.deleted, want)
		}
	}
}
