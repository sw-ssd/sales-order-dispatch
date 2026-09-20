package entitlements

import (
	"context"
	"sync"
	"time"
)

// Cache 抽象權益快取：單元測試用記憶體、production 注入 Valkey 實作。
// 失效語意：方案／override／訂閱異動時由寫入方 Delete（不靠 TTL 正確性），TTL 僅保底。
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// MemoryCache 為程序內快取（單元測試與單 replica 開發用）。每筆自帶到期時間，Get 時判定；
// 過期即視同未命中並清掉（不讓死鍵累積），因此 New(..., ttl>0) 的 ttl 是真的會生效的。
// ttl <= 0 的 Set 不入庫 —— 與 Service「ttl <= 0 表示不快取」同一語意，兩端不得各說各話。
type MemoryCache struct {
	mu   sync.Mutex
	data map[string]memEntry
}

type memEntry struct {
	val     []byte
	expires time.Time
}

func NewMemoryCache() *MemoryCache { return &MemoryCache{data: map[string]memEntry{}} }

func (m *MemoryCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.data[key]
	if !ok {
		return nil, false, nil
	}
	if !time.Now().Before(e.expires) { // 到期（含等於）即未命中
		delete(m.data, key)
		return nil, false, nil
	}
	return e.val, true, nil
}

func (m *MemoryCache) Set(_ context.Context, key string, val []byte, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = memEntry{val: val, expires: time.Now().Add(ttl)}
	return nil
}

func (m *MemoryCache) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}
