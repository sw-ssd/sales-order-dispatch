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

// MemoryCache 為程序內快取（單元測試與單 replica 開發用）。不實作 TTL：
// 判定路徑只依賴「同一個 ttl 內查得到」，過期由 production 的 Valkey 實作負責。
type MemoryCache struct {
	mu   sync.Mutex
	data map[string][]byte
}

func NewMemoryCache() *MemoryCache { return &MemoryCache{data: map[string][]byte{}} }

func (m *MemoryCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[key]
	return v, ok, nil
}

func (m *MemoryCache) Set(_ context.Context, key string, val []byte, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = val
	return nil
}

func (m *MemoryCache) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}
