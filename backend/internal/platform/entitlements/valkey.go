package entitlements

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ValkeyCache 為權益快取的 production 實作：**跨行程、跨 replica 共用同一份快取**。
//
// 為什麼一定要共用（而不是像 MemoryCache 一樣各行程一份）：訂閱狀態有三個寫入來源，而它們分屬
// 不同行程 —— API（T9 的平台 RPC）、排程（billing 的掃描式轉移）、事件 consumer（凍結／復原）。
// 只有三個行程共用同一顆 Valkey，其中一個刪掉的鍵才會對其他兩個立刻生效。行程內快取做不到這件事：
// 排程停了某個租戶，API 那台的判定還是舊的，直到 TTL 到期為止。
//
// 失效語意：**寫入路徑顯式 Delete（見 Invalidate／InvalidateAll），TTL 僅為保底** ——
// 正確性不依賴 TTL 到期，`ttl` 只是「萬一某條寫入路徑漏了失效」的收斂上限。
type ValkeyCache struct{ client *redis.Client }

// NewValkeyCache 建立 Valkey 權益快取。client 由呼叫端建立（連線設定屬組態，不屬判定層）。
//
// 回傳 Cache 介面（C-32）：呼叫端只該依賴契約，不該拿到 client 去下別的指令 ——
// key 的形態（ent:{companyID}）是判定層的內政，從外面拼 key 就是複製一份會漂移的契約。
func NewValkeyCache(client *redis.Client) Cache { return &ValkeyCache{client: client} }

// Get 取快取；不存在（redis.Nil）回 ok=false 且**不是錯誤** —— 未命中是常態，不是故障。
func (v *ValkeyCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	raw, err := v.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return raw, true, nil
}

// Set 寫入快取。**ttl <= 0 一律不入庫**（與 MemoryCache.Set 同一語意）。
//
// 為什麼不能沿用 Valkey 的原生語意（0 = 永不過期）：判定層的 `ttl <= 0` 意思是「不快取」，
// 兩端各說各話的失效模式是**最壞的一種** —— 營運把 TTL 設成 0 想關掉快取，實際卻得到
// 「永久快取這一份權益」，之後所有失效都靠寫入路徑記得刪；漏一條路徑就是永遠讀舊方案。
// 呼叫端（Service.state）本來就會先擋，這裡再擋一次是為了讓**快取介面的契約**只有一種讀法。
func (v *ValkeyCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return v.client.Set(ctx, key, val, ttl).Err()
}

// Delete 刪除單一鍵。刪不存在的鍵不是錯誤（DEL 回 0）—— 失效必須是冪等的：同一租戶連續兩次
// 狀態異動、或 API 與排程都刪了同一個鍵，都不該讓呼叫端看到錯誤。
func (v *ValkeyCache) Delete(ctx context.Context, key string) error {
	return v.client.Del(ctx, key).Err()
}

// Keys 以 SCAN 逐批收集符合 pattern 的鍵（供 InvalidateAll 全量失效）。
//
// 不用 KEYS：KEYS 在大鍵空間時會**阻塞整個 Valkey**（單執行緒掃完整個 keyspace），而權益快取
// 是所有租戶共用的 —— 一次方案改動就能讓全站判定一起卡住。SCAN 是游標式的分批掃描，
// 每次只碰 COUNT 個桶，代價可控。
//
// 刻意的取捨：SCAN 可能回重複鍵（同一批之間資料變動時），故不保證「不重複」；對 Delete 而言
// 重複刪同一個鍵無害（冪等），所以不需要去重。
func (v *ValkeyCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	var out []string
	var cursor uint64
	for {
		keys, next, err := v.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		out = append(out, keys...)
		if next == 0 {
			return out, nil
		}
		cursor = next
	}
}

// Scanner 為「支援掃描鍵」的快取。只有需要全量失效（方案／價目異動）時才要求這個能力，
// 故它不進 Cache 介面：單租戶失效（絕大多數路徑）只需要 Get／Set／Delete。
type Scanner interface {
	Keys(ctx context.Context, pattern string) ([]string, error)
}

// cachePrefix 為權益快取鍵的前綴（cacheKey 產生 "ent:{companyID}"）：全量失效靠它掃描。
const cachePrefix = "ent:"

// Invalidate 清除某租戶的權益快取 —— **所有寫入路徑的唯一失效入口**（方案、override、
// 訂閱狀態異動），讓 key 的形態（ent:{companyID}）只有一個地方知道。
//
// c 為 nil（該行程沒有接上快取）時 no-op：呼叫端（billing／consumer／T9 的服務層）的快取依賴是
// **可選**的，讓它們各自判空就是讓 N 個地方各自可能忘記判而 panic，而「沒有快取」本來就等於
// 「沒有東西要失效」。
//
// 錯誤不經 internal/errcode：這個錯誤唯一的去處是呼叫端的 log（快取失效失敗不得讓已落地的
// 帳務／設定寫入回錯誤），不跨網路就沒有穩定對外碼可言。
func Invalidate(ctx context.Context, c Cache, companyID int) error {
	if c == nil {
		return nil
	}
	if err := c.Delete(ctx, cacheKey(companyID)); err != nil {
		return fmt.Errorf("失效權益快取(company=%d): %w", companyID, err)
	}
	return nil
}

// ErrScanUnsupported 為「這個快取實作不支援全量失效」的**哨兵錯誤**（實作缺 Scanner）。
//
// 為什麼要是可辨識的哨兵而不是一句錯誤字串：呼叫端（T9 的平台寫入 RPC）對這兩種結果的處置
// 完全不同 —— 「不支援」（本機／單 replica 的 MemoryCache）是**已知的部署形態**，TTL 就是
// 收斂上界，log 一行說明即可；「掃描／刪除失敗」（Valkey 故障）是意外，要留下可追的痕跡。
// 兩者都只進 log（已落地的寫入不得因快取回錯——見 Invalidate 的說明），但字串比對會讓
// 「Valkey 回了同樣的字」被當成「不支援」而靜默。
var ErrScanUnsupported = errors.New("此快取實作不支援全量失效（缺少 Keys）")

// InvalidateAll 清除**所有**租戶的權益快取（方案的內容與價目異動時使用）。
//
// 為什麼是「全部」而不是「用到該方案的租戶」：判定快照含方案的 entitlements，改一格方案等於
// 改到每一個用它的租戶；而「誰在用哪個方案」得先查一遍訂閱表（多一次跨表查詢、還要處理查詢與
// 刪除之間的競態：查完到刪完之間新指派的租戶會漏掉），模糊地全刪反而正確且更簡單。
//
// ponytail: 逐鍵 SCAN + DEL；方案／價目的改動是人工操作、頻率極低（v1 的租戶量也小），
// 量體成長到刪除延遲可感時再改成版本號鍵（ent:v{n}:{companyID}，改方案只動版本號）。
func InvalidateAll(ctx context.Context, c Cache) error {
	if c == nil {
		return nil
	}
	sc, ok := c.(Scanner)
	if !ok {
		return ErrScanUnsupported
	}
	keys, err := sc.Keys(ctx, cachePrefix+"*")
	if err != nil {
		return fmt.Errorf("掃描權益快取鍵: %w", err)
	}
	for _, k := range keys {
		if err := c.Delete(ctx, k); err != nil {
			return fmt.Errorf("失效權益快取(%s): %w", k, err)
		}
	}
	return nil
}
