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
	"testing"
	"time"

	"github.com/google/uuid"
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
	c, src, _ := newConsumer(t, store.Event{ID: 1, AggregateType: "subscription", AggregateID: uuid.MustParse("00000000-0000-0000-0000-000000000005"),
		EventType: "subscription.suspended", Payload: []byte(`{"company_id":42}`)})
	cache := &recordingCache{}
	c.WithCache(cache)

	if _, err := c.DispatchOnce(context.Background(), 10); err != nil {
		t.Fatalf("派送: %v", err)
	}
	if len(src.emitted) != 1 {
		t.Fatalf("應發出 1 筆 company.status_changed outbox,got %d", len(src.emitted))
	}
	assertInvalidated(t, cache, "ent:42")
}

// subscription.expired（G7：取消且期末已過 → 凍結）同樣要失效。
func TestDispatchInvalidatesCacheForExpiredEvent(t *testing.T) {
	c, src, _ := newConsumer(t, store.Event{ID: 1, AggregateType: "subscription", AggregateID: uuid.MustParse("00000000-0000-0000-0000-000000000005"),
		EventType: "subscription.expired", Payload: []byte(`{"company_id":43}`)})
	cache := &recordingCache{}
	c.WithCache(cache)

	if _, err := c.DispatchOnce(context.Background(), 10); err != nil {
		t.Fatalf("派送: %v", err)
	}
	if len(src.emitted) != 1 {
		t.Fatalf("應發出 1 筆 outbox,got %d", len(src.emitted))
	}
	assertInvalidated(t, cache, "ent:43")
}

// 未對應型別（period.opened）：只認領、不動公司狀態 → 不得失效。
func TestDispatchDoesNotInvalidateForUnmappedEvent(t *testing.T) {
	c, src, _ := newConsumer(t, store.Event{ID: 1, AggregateType: "subscription", AggregateID: uuid.MustParse("00000000-0000-0000-0000-000000000005"),
		EventType: "period.opened", Payload: []byte(`{"company_id":42}`)})
	cache := &recordingCache{}
	c.WithCache(cache)

	if _, err := c.DispatchOnce(context.Background(), 10); err != nil {
		t.Fatalf("派送: %v", err)
	}
	if len(src.emitted) != 0 {
		t.Fatalf("未改公司狀態的事件不得發 outbox,got %+v", src.emitted)
	}
	if len(cache.deleted) != 0 {
		t.Fatalf("未改公司狀態的事件不得失效快取，got %v", cache.deleted)
	}
}

// payload 壞掉（缺 company_id）→ 報錯、不發 outbox、不失效；且要能重試。
func TestDispatchFailureDoesNotInvalidateCache(t *testing.T) {
	c, src, _ := newConsumer(t, store.Event{ID: 1, AggregateType: "subscription",
		AggregateID: uuid.MustParse("00000000-0000-0000-0000-000000000005"), EventType: "subscription.suspended", Payload: []byte(`{}`)})
	cache := &recordingCache{}
	c.WithCache(cache)

	if _, err := c.DispatchOnce(context.Background(), 10); err == nil {
		t.Fatal("壞 payload 必須回報（errors.Join），不得靜默")
	}
	if len(src.emitted) != 0 {
		t.Fatalf("失敗的事件不得發 outbox,got %+v", src.emitted)
	}
	if len(cache.deleted) != 0 {
		t.Fatalf("失敗的事件不得失效快取，got %v", cache.deleted)
	}
}

// 沒接上快取（cache=nil）時照常派送：可選依賴的意義。
func TestConsumerWorksWithoutCache(t *testing.T) {
	c, src, _ := newConsumer(t, store.Event{ID: 1, AggregateType: "subscription",
		AggregateID: uuid.MustParse("00000000-0000-0000-0000-000000000005"), EventType: "subscription.suspended", Payload: []byte(`{"company_id":42}`)})

	if n, err := c.DispatchOnce(context.Background(), 10); err != nil || n != 1 {
		t.Fatalf("未接快取時派送應照常運作: n=%d err=%v", n, err)
	}
	if len(src.emitted) != 1 {
		t.Fatalf("公司狀態仍應被發出 outbox 一次，got %d", len(src.emitted))
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
