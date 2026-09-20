package auth

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"
)

import (
	"golang.org/x/crypto/bcrypt"
)

func TestHashVerifyPassword(t *testing.T) {
	hash, err := HashPassword("secret-123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !VerifyPassword(hash, "secret-123") {
		t.Fatal("正確密碼應通過")
	}
	if VerifyPassword(hash, "wrong") {
		t.Fatal("錯誤密碼不應通過")
	}
	// OIDC 帳號的 sentinel 非合法 bcrypt,密碼登入必失敗
	if VerifyPassword(OIDCPasswordSentinel, "anything") {
		t.Fatal("OIDC sentinel 不得通過密碼驗證")
	}
	if VerifyPassword("", "anything") {
		t.Fatal("空雜湊不得通過")
	}
}

func TestLoginLock(t *testing.T) {
	ctx := context.Background()
	lock := NewLoginLock(NewMemoryStore())

	locked, err := lock.IsLocked(ctx, "C001")
	if err != nil || locked {
		t.Fatalf("初始不應鎖定: locked=%v err=%v", locked, err)
	}
	for i := 1; i <= MaxLoginFailures; i++ {
		n, err := lock.RecordFailure(ctx, "C001")
		if err != nil {
			t.Fatalf("RecordFailure: %v", err)
		}
		if n != i {
			t.Fatalf("失敗次數 = %d, want %d", n, i)
		}
	}
	locked, err = lock.IsLocked(ctx, "C001")
	if err != nil || !locked {
		t.Fatalf("5 次失敗後應鎖定: locked=%v err=%v", locked, err)
	}

	// 鎖定期間再失敗,次數維持不超過
	if _, err := lock.RecordFailure(ctx, "C001"); err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}
	locked, _ = lock.IsLocked(ctx, "C001")
	if !locked {
		t.Fatal("鎖定期間應維持鎖定")
	}

	// 登入成功 → 清除
	if err := lock.Clear(ctx, "C001"); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	locked, _ = lock.IsLocked(ctx, "C001")
	if locked {
		t.Fatal("Clear 後不應鎖定")
	}
}

func TestLoginLockTTLExpiry(t *testing.T) {
	ctx := context.Background()
	kv := NewMemoryStore()
	lock := NewLoginLock(kv)

	// 直接對儲存設定極短 TTL,驗證鎖定窗口(TTL)到期自動解除
	for i := 0; i < MaxLoginFailures; i++ {
		if _, err := lock.RecordFailure(ctx, "C002"); err != nil {
			t.Fatalf("RecordFailure: %v", err)
		}
	}
	if err := kv.Expire(ctx, loginFailKey("C002"), 30*time.Millisecond); err != nil {
		t.Fatalf("Expire: %v", err)
	}
	locked, _ := lock.IsLocked(ctx, "C002")
	if !locked {
		t.Fatal("TTL 到期前應鎖定")
	}
	time.Sleep(60 * time.Millisecond)
	locked, _ = lock.IsLocked(ctx, "C002")
	if locked {
		t.Fatal("TTL 到期後應自動解除鎖定")
	}
}

// TestLoginLockLockedUntil 釘住「解鎖時間」契約，且走**真實流程**（RecordFailure，不靠人工 Expire）：
// ①鎖定期間計數鍵的窗口仍在（第一次失敗起算 LockDuration）——這同時釘住 `MemoryStore.Incr`
// 不得清掉既有 TTL（否則計數鍵永不過期、鎖定永不解鎖，與 Redis 的 INCR 語意分歧）；
// ②解鎖時間落在 (now, now+LockDuration]；③解鎖時間讀的是**KV 的真實剩餘 TTL**，
// 不是 fallback 的「now + LockDuration」假值。
func TestLoginLockLockedUntil(t *testing.T) {
	ctx := context.Background()
	kv := NewMemoryStore()
	lock := NewLoginLock(kv)

	if until, err := lock.LockedUntil(ctx, "C003"); err != nil || !until.IsZero() {
		t.Fatalf("未鎖定時應回零值時間: until=%v err=%v", until, err)
	}
	for range MaxLoginFailures {
		if _, err := lock.RecordFailure(ctx, "C003"); err != nil {
			t.Fatalf("RecordFailure: %v", err)
		}
	}
	// ① 真實流程後窗口仍在（首次失敗時設定；後續 Incr 不得清掉）。
	ttl, err := kv.TTL(ctx, loginFailKey("C003"))
	if err != nil {
		t.Fatalf("TTL: %v", err)
	}
	if ttl <= 0 {
		t.Fatal("多次失敗後計數鍵不得失去 TTL（否則鎖定永不解鎖；Redis 的 INCR 會保留 TTL）")
	}
	if ttl > LockDuration {
		t.Fatalf("剩餘窗口 = %v，不得超過 LockDuration = %v", ttl, LockDuration)
	}
	// ② 解鎖時間落在窗口內。
	until, err := lock.LockedUntil(ctx, "C003")
	if err != nil {
		t.Fatalf("LockedUntil: %v", err)
	}
	if remaining := time.Until(until); remaining <= 0 || remaining > LockDuration {
		t.Fatalf("解鎖時間應落在 (now, now+%v] 內,得到 remaining=%v", LockDuration, remaining)
	}
	// ③ 換一個「剩餘 20 分鐘」的計數鍵（鎖定狀態的真實樣貌）→ 解鎖時間必須跟著縮短；
	// 若實作回 fallback（now + LockDuration = 30 分）就抓得到。
	if err := kv.Set(ctx, loginFailKey("C004"), strconv.Itoa(MaxLoginFailures), 20*time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	until4, err := lock.LockedUntil(ctx, "C004")
	if err != nil {
		t.Fatalf("LockedUntil(20 分窗口): %v", err)
	}
	if remaining := time.Until(until4); remaining <= 0 || remaining > 21*time.Minute {
		t.Fatalf("解鎖時間應反映 KV 的真實剩餘 TTL（約 20 分）,得到 remaining=%v（fallback 假值＝30 分）", remaining)
	}
}

func TestHashPasswordArgon2id(t *testing.T) {
	hash, err := HashPassword("secret-123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	// Argon2id 編碼格式: $argon2id$v=19$m=...,t=...,p=...$<salt>$<key>
	if len(hash) < 30 || !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Fatalf("非預期雜湊格式: %q", hash)
	}
	if !VerifyPassword(hash, "secret-123") {
		t.Fatal("Argon2id 正確密碼應通過")
	}
	if VerifyPassword(hash, "wrong") {
		t.Fatal("錯誤密碼不應通過")
	}
	// 每次 salt 隨機 → 相同明文產生不同 hash
	h2, err := HashPassword("secret-123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == h2 {
		t.Fatal("同一明文兩次雜湊應不同(salt 隨機)")
	}
	// 舊 bcrypt 雜湊必須被拒(Argon2id 取代 bcrypt,規格)
	legacyBcrypt, err := bcrypt.GenerateFromPassword([]byte("secret-123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	if VerifyPassword(string(legacyBcrypt), "secret-123") {
		t.Fatal("舊 bcrypt 雜湊不得通過 Argon2id 驗證")
	}
}
