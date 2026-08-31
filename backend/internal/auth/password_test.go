package auth

import (
	"context"
	"testing"
	"time"
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
