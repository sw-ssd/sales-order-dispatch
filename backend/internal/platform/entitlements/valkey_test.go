// Task 8：權益快取的失效入口（Invalidate／InvalidateAll）與「快取故障不得拒絕服務」的單元契約。
//
// 這裡用假快取（記錄 Delete、可假裝故障），因為要驗的是**判定層與寫入層之間的契約**：
// 誰刪了哪個鍵、nil 快取怎麼辦、快取壞掉時判定還在不在。真 Valkey 的鍵值語意（含 SCAN）由
// valkey_integration_test.go 以真容器把關。
package entitlements_test

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

// fakeCache 記錄失效呼叫；scanKeys 模擬「掃得到哪些鍵」；failKeys 模擬特定鍵刪除失敗。
type fakeCache struct {
	deleted  []string
	scanKeys []string
	scanErr  error
	failKeys map[string]error
	// getErr／setErr 模擬 Valkey 故障（連不上、權限、逾時）。
	getErr error
	setErr error
}

func (c *fakeCache) Get(context.Context, string) ([]byte, bool, error) {
	if c.getErr != nil {
		return nil, false, c.getErr
	}
	return nil, false, nil
}

func (c *fakeCache) Set(context.Context, string, []byte, time.Duration) error { return c.setErr }
func (c *fakeCache) Delete(_ context.Context, key string) error {
	if err, ok := c.failKeys[key]; ok {
		return err
	}
	c.deleted = append(c.deleted, key)
	return nil
}

// Keys 只實作 ent:* 的前綴過濾：真 Valkey 版走 SCAN MATCH（見 valkey_integration_test.go）。
func (c *fakeCache) Keys(_ context.Context, pattern string) ([]string, error) {
	if c.scanErr != nil {
		return nil, c.scanErr
	}
	prefix := strings.TrimSuffix(pattern, "*")
	var out []string
	for _, k := range c.scanKeys {
		if strings.HasPrefix(k, prefix) {
			out = append(out, k)
		}
	}
	return out, nil
}

// ① 單租戶失效：鍵的形態是 ent:{companyID}（跨行程共用的契約，寫錯就是刪了也沒生效）。
func TestInvalidateDeletesTenantKey(t *testing.T) {
	c := &fakeCache{}
	if err := entitlements.Invalidate(context.Background(), c, 42); err != nil {
		t.Fatalf("Invalidate: %v", err)
	}
	if len(c.deleted) != 1 || c.deleted[0] != "ent:42" {
		t.Fatalf("應刪除 ent:42，got %v", c.deleted)
	}
}

// ② 沒接上快取（nil）＝沒有東西要失效：可選依賴的呼叫端不該到處判空。
func TestInvalidateWithNilCacheIsNoOp(t *testing.T) {
	if err := entitlements.Invalidate(context.Background(), nil, 42); err != nil {
		t.Fatalf("nil 快取應為 no-op，got %v", err)
	}
	if err := entitlements.InvalidateAll(context.Background(), nil); err != nil {
		t.Fatalf("nil 快取的全量失效應為 no-op，got %v", err)
	}
}

// ③ 全量失效（方案內容／價目異動）：只刪 ent:*，不得波及別的用途的鍵（同一顆 Valkey 也放 session）。
func TestInvalidateAllRemovesEveryTenantKey(t *testing.T) {
	c := &fakeCache{scanKeys: []string{"ent:1", "ent:2", "sess:abc", "lock:1"}}
	if err := entitlements.InvalidateAll(context.Background(), c); err != nil {
		t.Fatalf("InvalidateAll: %v", err)
	}
	if len(c.deleted) != 2 {
		t.Fatalf("應刪除 2 個租戶鍵，got %v", c.deleted)
	}
	for _, k := range c.deleted {
		if k != "ent:1" && k != "ent:2" {
			t.Fatalf("不得刪除不相關的鍵：%v", c.deleted)
		}
	}
}

// ④ 不支援掃描的快取（行程內 MemoryCache）必須**大聲**回錯誤，不得靜默假裝成功 ——
// 靜默的話，「改了方案卻沒人失效」會完全沒有痕跡。
func TestInvalidateAllUnsupportedCacheFailsLoudly(t *testing.T) {
	err := entitlements.InvalidateAll(context.Background(), entitlements.NewMemoryCache())
	if err == nil {
		t.Fatal("沒有 Keys 的快取做全量失效時必須回錯誤")
	}
}

// ④-b 不支援全量失效必須是**可辨識的哨兵**(T9)：呼叫端對它的處置與「掃描失敗」完全不同
// （前者是已知的降級部署，TTL 就是收斂上界；後者是 Valkey 故障，要留下可追的痕跡）。
// 以字串比對的話，「Valkey 回了同一句話」會被當成不支援而靜默。
func TestInvalidateAllUnsupportedCacheReturnsSentinel(t *testing.T) {
	err := entitlements.InvalidateAll(context.Background(), entitlements.NewMemoryCache())
	if !errors.Is(err, entitlements.ErrScanUnsupported) {
		t.Fatalf("必須回哨兵 ErrScanUnsupported，got %v", err)
	}
	// 掃描失敗**不得**被誤判成「不支援」（否則故障會被當成部署形態而靜默）。
	c := &fakeCache{scanErr: errors.New("模擬 Valkey 故障")}
	if err := entitlements.InvalidateAll(context.Background(), c); errors.Is(err, entitlements.ErrScanUnsupported) {
		t.Fatalf("掃描失敗不得回 ErrScanUnsupported，got %v", err)
	}
}

// ⑤ 掃描失敗不得當成「掃到 0 個鍵」：那是把故障偽裝成「沒有東西要清」。
func TestInvalidateAllScanFailureIsReported(t *testing.T) {
	c := &fakeCache{scanErr: errors.New("模擬 Valkey 故障")}
	if err := entitlements.InvalidateAll(context.Background(), c); err == nil {
		t.Fatal("掃描失敗必須回錯誤")
	}
}

// ⑤b 未結項 #20：部分鍵刪除失敗不得 early-return —— 否則第一個壞鍵後面的租戶永遠不清掉，
// 且呼叫端在 log 裡只看到第一個鍵、誤以為只壞了一個。
// RED:目前第一個 DEL 失敗即 return，第二鍵 ent:2 未被嘗試。
func TestInvalidateAllContinuesPastDeleteFailure(t *testing.T) {
	c := &fakeCache{
		scanKeys: []string{"ent:1", "ent:2"},
		failKeys: map[string]error{"ent:1": errors.New("模擬單鍵刪除失敗")},
	}
	err := entitlements.InvalidateAll(context.Background(), c)
	if err == nil {
		t.Fatal("部分刪除失敗必須回錯誤")
	}
	// 第二鍵必須仍被嘗試（fakeCache 只記錄成功的刪除，故 ent:2 必須出現在 deleted）。
	found := false
	for _, d := range c.deleted {
		if d == "ent:2" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("第一鍵失敗後仍須嘗試後續鍵，deleted=%v", c.deleted)
	}
	if !strings.Contains(err.Error(), "ent:1") {
		t.Fatalf("錯誤必須指出是哪個鍵失敗，got %v", err)
	}
}

// ⑥ **快取故障不得拒絕服務**：Valkey 掛掉時判定回源（配額守衛還在，只是慢），
// 而不是把全站業務寫入一起打掉。
func TestJudgementSurvivesCacheFailure(t *testing.T) {
	f := store.NewFake()
	f.PutFeature(seatsDef)
	f.PutPlan("std", stdPlan)
	f.PutSubscription(*subWithStatus("active"))

	var logs bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(prev)

	broken := &fakeCache{getErr: errors.New("模擬 Valkey 不可用"), setErr: errors.New("模擬 Valkey 不可用")}
	svc := entitlements.New(f, counting{seats: 1}, broken, time.Minute)

	if err := svc.CheckLimit(context.Background(), 1, seats, 1); err != nil {
		t.Fatalf("快取故障時判定仍須放行（回源），got %v", err)
	}
	// log 要吵：效能退化的原因必須看得見（否則就只是「系統偶爾很慢」）。
	if got := logs.String(); !strings.Contains(got, "權益快取") {
		t.Fatalf("快取故障必須留下 log，got %q", got)
	}
}
