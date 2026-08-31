package auth

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// testUserID 為整合測試專用使用者 ID(避免與真實資料碰撞)。
const testUserID = 424242

// openTestValkey 嘗試連線本機 Valkey;不可用時回傳 nil。
func openTestValkey(t *testing.T) *redis.Client {
	t.Helper()
	c := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		_ = c.Close()
		t.Skipf("Valkey 不可用,跳過整合測試: %v", err)
	}
	t.Cleanup(func() {
		ctx := context.Background()
		c.Del(ctx, tokenVersionKey(testUserID), refreshUserSetKey(testUserID))
		_ = c.Close()
	})
	return c
}

// TestValkeyTokenIssueRefreshRotation 以真實 Valkey 驗證 token 發行 / refresh 旋轉 / 撤銷
// (T13 驗收;Valkey 缺省時本測試自動 skip,由 MemoryStore unit 測試覆蓋)。
func TestValkeyTokenIssueRefreshRotation(t *testing.T) {
	c := openTestValkey(t)
	ctx := context.Background()
	tm := NewTokenManager("it-secret", NewRedisStore(c))

	access, err := tm.IssueAccess(ctx, TokenSubject{UserID: testUserID, CompanyID: 1, Role: "customer"})
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	if _, err := tm.VerifyAccess(ctx, access); err != nil {
		t.Fatalf("VerifyAccess: %v", err)
	}

	r1, err := tm.IssueRefresh(ctx, testUserID, 0)
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}
	r2, err := tm.RotateRefresh(ctx, r1, testUserID, 0)
	if err != nil {
		t.Fatalf("RotateRefresh: %v", err)
	}
	if _, _, err := tm.VerifyRefresh(ctx, r1); err != ErrInvalidRefresh {
		t.Fatalf("旋轉後舊 refresh 應失效,got %v", err)
	}
	if _, _, err := tm.VerifyRefresh(ctx, r2); err != nil {
		t.Fatalf("新 refresh 應有效: %v", err)
	}

	// RevokeAll:全部 refresh 撤銷 + token_version 遞增(舊 access 亦失效)
	if err := tm.RevokeAll(ctx, testUserID); err != nil {
		t.Fatalf("RevokeAll: %v", err)
	}
	if _, _, err := tm.VerifyRefresh(ctx, r2); err != ErrInvalidRefresh {
		t.Fatalf("RevokeAll 後 refresh 應失效,got %v", err)
	}
	if _, err := tm.VerifyAccess(ctx, access); err != ErrTokenRevoked {
		t.Fatalf("RevokeAll 後 access 應失效,got %v", err)
	}
}

// TestValkeyLoginLock 以真實 Valkey 驗證登入失敗計數與鎖定(T12 驗收)。
func TestValkeyLoginLock(t *testing.T) {
	c := openTestValkey(t)
	ctx := context.Background()
	lock := NewLoginLock(NewRedisStore(c))
	account := "IT-LOCK-001"
	t.Cleanup(func() { _ = lock.Clear(ctx, account) })

	for range MaxLoginFailures {
		if _, err := lock.RecordFailure(ctx, account); err != nil {
			t.Fatalf("RecordFailure: %v", err)
		}
	}
	locked, err := lock.IsLocked(ctx, account)
	if err != nil || !locked {
		t.Fatalf("5 次失敗後應鎖定: locked=%v err=%v", locked, err)
	}
	if err := lock.Clear(ctx, account); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	locked, _ = lock.IsLocked(ctx, account)
	if locked {
		t.Fatal("Clear 後不應鎖定")
	}
}

// TestValkeyWebSessionStore 以真實 Valkey 驗證 scs session store(Web session cookie 路徑)。
func TestValkeyWebSessionStore(t *testing.T) {
	c := openTestValkey(t)
	ctx := context.Background()
	st := NewSessionStore(NewRedisStore(c))
	token := "it-session-token"

	if err := st.CommitCtx(ctx, token, []byte("user:42"), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CommitCtx: %v", err)
	}
	b, found, err := st.FindCtx(ctx, token)
	if err != nil || !found || string(b) != "user:42" {
		t.Fatalf("FindCtx: found=%v data=%q err=%v", found, b, err)
	}
	if err := st.DeleteCtx(ctx, token); err != nil {
		t.Fatalf("DeleteCtx: %v", err)
	}
	if _, found, err := st.FindCtx(ctx, token); err != nil || found {
		t.Fatalf("刪除後不應找到: found=%v err=%v", found, err)
	}
}
