// Package auth 提供認證基礎元件:Google OIDC、密碼/鎖定、JWT + refresh token、Web session。
// 本 package 只依賴標準函式庫與第三方 SDK,不依賴 ent(呼叫端負責把 *ent.User 轉為 TokenSubject)。
package auth

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// KVStore 抽象後端鍵值儲存(Valkey 為正式實作;測試用 MemoryStore)。
// 涵蓋 token_version、refresh token、一次性 token(state / registration)、登入失敗計數。
type KVStore interface {
	// Get 回傳 key 的值;不存在時 ok=false(不視為錯誤)。
	Get(ctx context.Context, key string) (string, bool, error)
	// Set 寫入值並設定 TTL。
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	// Expire 為既有 key 設定 TTL。
	Expire(ctx context.Context, key string, ttl time.Duration) error
	// Delete 刪除 key(不存在時為 no-op)。
	Delete(ctx context.Context, key string) error
	// Incr 將 key 遞增 1 並回傳新值;key 不存在時視為 0 → 1。
	Incr(ctx context.Context, key string) (int64, error)
	// SetAdd 將 member 加入集合。
	SetAdd(ctx context.Context, key, member string) error
	// SetRemove 將 member 移出集合。
	SetRemove(ctx context.Context, key, member string) error
	// SetMembers 列出集合全部 member。
	SetMembers(ctx context.Context, key string) ([]string, error)
}

// RedisStore 以 go-redis(Valkey 相容)實作 KVStore。
type RedisStore struct {
	c *redis.Client
}

// NewRedisStore 建立 RedisStore。
func NewRedisStore(c *redis.Client) *RedisStore {
	return &RedisStore{c: c}
}

func (s *RedisStore) Get(ctx context.Context, key string) (string, bool, error) {
	v, err := s.c.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}

func (s *RedisStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.c.Set(ctx, key, value, ttl).Err()
}

func (s *RedisStore) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return s.c.Expire(ctx, key, ttl).Err()
}

func (s *RedisStore) Delete(ctx context.Context, key string) error {
	return s.c.Del(ctx, key).Err()
}

func (s *RedisStore) Incr(ctx context.Context, key string) (int64, error) {
	return s.c.Incr(ctx, key).Result()
}

func (s *RedisStore) SetAdd(ctx context.Context, key, member string) error {
	return s.c.SAdd(ctx, key, member).Err()
}

func (s *RedisStore) SetRemove(ctx context.Context, key, member string) error {
	return s.c.SRem(ctx, key, member).Err()
}

func (s *RedisStore) SetMembers(ctx context.Context, key string) ([]string, error) {
	return s.c.SMembers(ctx, key).Result()
}

// consumeRefreshScript 原子「讀取並作廢」refresh token:Lua 內 GET + DEL + SREM 一氣呵成,
// 避免並發旋轉時兩個請求同時讀到同一 token(重放偵測)。不存在時回 nil。
var consumeRefreshScript = redis.NewScript(`
local v = redis.call('GET', KEYS[1])
if not v then
  return nil
end
redis.call('DEL', KEYS[1])
redis.call('SREM', KEYS[2], ARGV[1])
return v
`)

// ConsumeRefresh 原子讀取並刪除 refresh token 記錄,並自使用者集合移除;
// 不存在 → ok=false。供 TokenManager.RotateRefresh 做並發安全的旋轉。
func (s *RedisStore) ConsumeRefresh(ctx context.Context, key, userSetKey, member string) (string, bool, error) {
	v, err := consumeRefreshScript.Run(ctx, s.c, []string{key, userSetKey}, member).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	val, ok := v.(string)
	if !ok {
		return "", false, errors.New("auth: consume refresh 回傳型別錯誤")
	}
	return val, true, nil
}

// memEntry 為 MemoryStore 的一筆記錄。
type memEntry struct {
	value   string
	expires time.Time // zero = 不過期
}

// MemoryStore 供測試與無 Valkey 環境使用的記憶體 KVStore。
type MemoryStore struct {
	mu   sync.Mutex
	m    map[string]memEntry
	sets map[string]map[string]struct{}
}

// NewMemoryStore 建立 MemoryStore。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{m: make(map[string]memEntry), sets: make(map[string]map[string]struct{})}
}

func (s *MemoryStore) getLocked(key string) (string, bool) {
	e, ok := s.m[key]
	if !ok {
		return "", false
	}
	if !e.expires.IsZero() && time.Now().After(e.expires) {
		delete(s.m, key)
		return "", false
	}
	return e.value, true
}

func (s *MemoryStore) Get(ctx context.Context, key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.getLocked(key)
	return v, ok, nil
}

func (s *MemoryStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := memEntry{value: value}
	if ttl > 0 {
		e.expires = time.Now().Add(ttl)
	}
	s.m[key] = e
	return nil
}

func (s *MemoryStore) Expire(ctx context.Context, key string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.m[key]
	if !ok {
		return nil
	}
	if ttl > 0 {
		e.expires = time.Now().Add(ttl)
	} else {
		e.expires = time.Time{}
	}
	s.m[key] = e
	return nil
}

func (s *MemoryStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
	return nil
}

func (s *MemoryStore) Incr(ctx context.Context, key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var n int64
	if v, ok := s.getLocked(key); ok {
		n, _ = strconv.ParseInt(v, 10, 64)
	}
	n++
	s.m[key] = memEntry{value: strconv.FormatInt(n, 10)}
	return n, nil
}

func (s *MemoryStore) SetAdd(ctx context.Context, key, member string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sets[key] == nil {
		s.sets[key] = make(map[string]struct{})
	}
	s.sets[key][member] = struct{}{}
	return nil
}

func (s *MemoryStore) SetRemove(ctx context.Context, key, member string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if set, ok := s.sets[key]; ok {
		delete(set, member)
		if len(set) == 0 {
			delete(s.sets, key)
		}
	}
	return nil
}

func (s *MemoryStore) SetMembers(ctx context.Context, key string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0, len(s.sets[key]))
	for m := range s.sets[key] {
		out = append(out, m)
	}
	return out, nil
}

// ConsumeRefresh 原子讀取並刪除 refresh token 記錄(鎖保護,與 Redis Lua 等價);
// 不存在 → ok=false。
func (s *MemoryStore) ConsumeRefresh(ctx context.Context, key, userSetKey, member string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.getLocked(key)
	if !ok {
		return "", false, nil
	}
	delete(s.m, key)
	if set, ok := s.sets[userSetKey]; ok {
		delete(set, member)
		if len(set) == 0 {
			delete(s.sets, userSetKey)
		}
	}
	return v, true, nil
}
