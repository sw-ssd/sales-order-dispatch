// Task 8：帳務寫入路徑的**權益快取失效**（不是 TTL 收斂，是提交後主動 Delete）。
//
// 為什麼每條路徑都要失效：訂閱狀態是權益判定的輸入之一（entitlements.usable），而訂閱狀態有
// **三個寫入來源、分屬不同行程** —— 平台 RPC（T9）、排程（本套件的掃描式轉移）、事件 consumer。
// 快取在 Valkey 上是三個行程共用的，所以任何一個行程刪了鍵，其他兩個看到的就已經是新的。
//
// 這幾條測的兩件事，缺一都不算接上：
//
//	① 狀態**真的**改變（掃描有轉移、收款有回 active）→ 提交後對該租戶 Delete；
//	② 寫入失敗或根本沒轉移 → **不得**刪（失敗卻刪是白刪：資料沒變、下一次判定又要重建快取）。
//
// 「提交後」這半由 billing_integration_test／entitlements 的失效整合測試以真 PostgreSQL 把關
// （假 store 的 WithTx 是整份快照還原，看不到 commit 邊界）。
package billing_test

import (
	"context"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/billing"
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

// ① 停用掃描：只有**真的**被轉成 suspended 的租戶要失效；寬限未過者不得被刪。
//
// 為什麼寬限未過的也要檢查：Delete 是多餘的往來，但真正的風險是「沒轉移卻刪」掩蓋了掃描條件
// 壞掉（例如查詢改成無條件回全表）—— 那種缺陷在日誌上只會看到一堆失效，看起來一切正常。
func TestSuspendOverdueInvalidatesOnlyChangedTenants(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	expiredGrace := now.Add(-time.Hour)
	futureGrace := now.Add(time.Hour)
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "past_due",
		PlanID: 1, BillingCycle: "monthly", GraceUntil: &expiredGrace})
	f.PutSubscription(store.Subscription{ID: 6, CompanyID: 43, Status: "past_due",
		PlanID: 1, BillingCycle: "monthly", GraceUntil: &futureGrace})

	cache := &recordingCache{}
	n, err := billing.NewBilling(f).WithCache(cache).SuspendOverdue(context.Background(), now)
	if err != nil || n != 1 {
		t.Fatalf("應只停用寬限已過的那一家: n=%d err=%v", n, err)
	}
	assertDeleted(t, cache, "ent:42")
}

// ② 逾期掃描：被轉成 past_due 的租戶要失效（cached 快照裡的 status 已與 DB 不一致）。
func TestMarkPastDueInvalidatesChangedTenants(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "active",
		PlanID: 1, BillingCycle: "monthly"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open",
		PeriodStart: now.AddDate(0, -1, 0), PeriodEnd: now.Add(-time.Hour)})

	cache := &recordingCache{}
	if n, err := billing.NewBilling(f).WithCache(cache).MarkPastDue(
		context.Background(), now, 7); err != nil || n != 1 {
		t.Fatalf("應轉移一家為 past_due: n=%d err=%v", n, err)
	}
	assertDeleted(t, cache, "ent:42")
}

// ③ 收款：訂閱可能由 trialing／past_due／suspended 回到 active（權益因此改變）→ 失效。
func TestRecordPaymentInvalidatesTenant(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "suspended",
		PlanID: 1, BillingCycle: "monthly"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open",
		AmountCents: 195000, PeriodStart: time.Now()})

	cache := &recordingCache{}
	if _, err := billing.NewBilling(f).WithCache(cache).RecordPayment(context.Background(),
		billing.RecordPaymentInput{CompanyID: 42, Provider: "manual", ActorOperatorID: 7,
			Reason: "匯款入帳"}); err != nil {
		t.Fatalf("收款: %v", err)
	}
	assertDeleted(t, cache, "ent:42")
}

// ④ **寫入失敗不得失效**：整份交易回滾了，快取卻被刪掉的話，下一次判定會重建出同樣的舊資料
// （白做一次平台庫查詢），而且掩蓋了「這次寫入其實失敗了」的事實。
func TestRecordPaymentFailureDoesNotInvalidate(t *testing.T) {
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "suspended",
		PlanID: 1, BillingCycle: "monthly"})
	f.PutPeriod(store.Period{ID: 9, SubscriptionID: 5, PeriodNo: 1, Status: "open",
		AmountCents: 195000, PeriodStart: time.Now()})

	cache := &recordingCache{}
	_, err := billing.NewBilling(failEvents{f}).WithCache(cache).RecordPayment(context.Background(),
		billing.RecordPaymentInput{CompanyID: 42, Provider: "manual", ActorOperatorID: 7,
			Reason: "匯款入帳"})
	if err == nil {
		t.Fatal("事件寫入失敗必須讓整筆收款失敗")
	}
	if len(cache.deleted) != 0 {
		t.Fatalf("寫入失敗不得失效快取，got %v", cache.deleted)
	}
}

// ⑤ 掃描沒有任何轉移 → 不得失效（否則每趟排程都會把全租戶的快取清掉，讓快取形同虛設）。
func TestNoTransitionDoesNotInvalidate(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "active",
		PlanID: 1, BillingCycle: "monthly"})

	cache := &recordingCache{}
	b := billing.NewBilling(f).WithCache(cache)
	if n, err := b.MarkPastDue(context.Background(), now, 7); err != nil || n != 0 {
		t.Fatalf("沒有逾期者: n=%d err=%v", n, err)
	}
	if n, err := b.SuspendOverdue(context.Background(), now); err != nil || n != 0 {
		t.Fatalf("沒有可停用者: n=%d err=%v", n, err)
	}
	if len(cache.deleted) != 0 {
		t.Fatalf("沒有轉移不得失效快取，got %v", cache.deleted)
	}
}

// ⑥ 沒接上快取（cache=nil）時，同一條路徑必須照常運作（可選依賴的意義就在這裡）。
func TestBillingWorksWithoutCache(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	expiredGrace := now.Add(-time.Hour)
	f := store.NewFakeBilling()
	f.PutSubscription(store.Subscription{ID: 5, CompanyID: 42, Status: "past_due",
		PlanID: 1, BillingCycle: "monthly", GraceUntil: &expiredGrace})

	if n, err := billing.NewBilling(f).SuspendOverdue(context.Background(), now); err != nil || n != 1 {
		t.Fatalf("未接快取時停用掃描應照常運作: n=%d err=%v", n, err)
	}
}

// assertDeleted 斷言「恰好失效了這些鍵」：缺一（漏失效）與多一（誤失效）都要看得見。
func assertDeleted(t *testing.T, c *recordingCache, want ...string) {
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
