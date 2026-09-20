//go:build integration

// Task 8 的 Valkey 快取整合驗收（真 Valkey 容器，樣板：internal/testsupport）。
//
// 單元測試用假快取驗契約（誰刪了哪個鍵），這裡驗的是**真的 Valkey 上**的行為：
// 未命中不是錯誤（redis.Nil 的處理）、TTL 真的被設（不是永不過期）、
// InvalidateAll 的 SCAN MATCH 只掃 ent:* 而不碰同一顆 Valkey 上的其他鍵（session、鎖）。
//
// 執行：task test:integration -- -count=1 -run TestIntegrationValkeyCache -v
package entitlements_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
	"github.com/salesorder/sales-order-1.0/backend/third_party/cache"
)

func TestIntegrationValkeyCacheRoundTrip(t *testing.T) {
	client := cache.NewClient(testsupport.Valkey(t))
	defer func() { _ = client.Close() }()
	c := entitlements.NewValkeyCache(client)
	ctx := t.Context()

	// ① Set 後 Get 命中，且 TTL 真的寫進 Valkey（ttl 是「保底」的那個上限，不是裝飾品）。
	if err := c.Set(ctx, "ent:42", []byte(`{"status":"active"}`), time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	raw, ok, err := c.Get(ctx, "ent:42")
	if err != nil || !ok {
		t.Fatalf("Get 應命中: ok=%v err=%v", ok, err)
	}
	if string(raw) != `{"status":"active"}` {
		t.Fatalf("值不符: %q", raw)
	}
	if ttl, err := client.TTL(ctx, "ent:42").Result(); err != nil || ttl <= 0 || ttl > time.Minute {
		t.Fatalf("TTL 應為 (0, 1m]，got %v err=%v", ttl, err)
	}

	// ② 未命中不是錯誤（redis.Nil 必須收斂成 ok=false）—— 判定層靠這個語意決定要不要回源。
	if raw, ok, err := c.Get(ctx, "ent:404"); err != nil || ok || raw != nil {
		t.Fatalf("未命中應為 (nil,false,nil)，got raw=%q ok=%v err=%v", raw, ok, err)
	}

	// ③ 單租戶失效（寫入路徑的入口）：鍵消失，重複刪不報錯（失效必須冪等）。
	if err := entitlements.Invalidate(ctx, c, 42); err != nil {
		t.Fatalf("Invalidate: %v", err)
	}
	if _, ok, err := c.Get(ctx, "ent:42"); err != nil || ok {
		t.Fatalf("Invalidate 後不得命中: ok=%v err=%v", ok, err)
	}
	if err := entitlements.Invalidate(ctx, c, 42); err != nil {
		t.Fatalf("對已消失的鍵失效不得報錯: %v", err)
	}

	// ④ 全量失效（方案／價目異動）：清掉所有 ent:*，**不得**波及同一顆 Valkey 上的其他鍵。
	if _, err := client.Set(ctx, "sess:keep", "1", time.Minute).Result(); err != nil {
		t.Fatalf("seed sess:keep: %v", err)
	}
	for i := 1; i <= 3; i++ {
		if err := c.Set(ctx, key(i), []byte(`{"status":"active"}`), time.Minute); err != nil {
			t.Fatalf("Set ent:%d: %v", i, err)
		}
	}
	if err := entitlements.InvalidateAll(ctx, c); err != nil {
		t.Fatalf("InvalidateAll: %v", err)
	}
	for i := 1; i <= 3; i++ {
		if _, ok, err := c.Get(ctx, key(i)); err != nil || ok {
			t.Fatalf("全量失效後 ent:%d 不得命中: ok=%v err=%v", i, ok, err)
		}
	}
	if v, err := client.Get(ctx, "sess:keep").Result(); err != nil || v != "1" {
		t.Fatalf("全量失效不得動到非權益鍵（session 等），got %q err=%v", v, err)
	}
}

func key(companyID int) string {
	// 顯式拼出契約形態（ent:{companyID}）：判定層的 cacheKey 是內政，但這個形態是跨行程的契約
	// ——排程與 consumer 刪的就是它。
	return "ent:" + strconv.Itoa(companyID)
}
