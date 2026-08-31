package auth

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/redis/go-redis/v9"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
)

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
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// TestValkeyRotateRefreshConcurrentReplay P1-4 驗收:以真實 Valkey + Lua 原子消耗驗證
// 並發重放同一 refresh token 時恰一個成功、其餘拒絕(Valkey 缺省時自動 skip)。
func TestValkeyRotateRefreshConcurrentReplay(t *testing.T) {
	c := openTestValkey(t)
	ctx := context.Background()

	name := strings.ReplaceAll(t.Name(), "/", "_")
	db := enttest.Open(t, "sqlite3", "file:"+name+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	co := db.Company.Create().SetName("整合公司").SetIdentifier("IT-" + name).SaveX(ctx)
	u := db.User.Create().
		SetEmail("it-rot-" + name + "@example.com").SetName("整合").SetStatus(user.StatusActive).
		SetRole("customer").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)
	t.Cleanup(func() { _ = c.Del(context.Background(), refreshUserSetKey(u.ID)) })

	tm := NewTokenManager("it-secret", NewRedisStore(c), db)

	r1, err := tm.IssueRefresh(ctx, u.ID, 0)
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}
	const workers = 8
	results := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := tm.RotateRefresh(ctx, r1, u.ID, 0)
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	success, revoked := 0, 0
	for err := range results {
		switch {
		case err == nil:
			success++
		case errors.Is(err, ErrRefreshRevoked):
			revoked++
		default:
			t.Fatalf("RotateRefresh 意外錯誤: %v", err)
		}
	}
	if success != 1 {
		t.Fatalf("並發旋轉應恰 1 個成功,got %d", success)
	}
	if revoked != workers-1 {
		t.Fatalf("並發重放應 %d 個被拒,got %d", workers-1, revoked)
	}
	if _, _, err := tm.VerifyRefresh(ctx, r1); err != ErrInvalidRefresh {
		t.Fatalf("並發旋轉後舊 token 應失效,got %v", err)
	}
}

// TestValkeyTokenIssueRefreshRotation 以真實 Valkey 驗證 token 發行 / refresh 旋轉 / 撤銷
// (T13 驗收;Valkey 缺省時本測試自動 skip,由 MemoryStore unit 測試覆蓋)。
func TestValkeyTokenIssueRefreshRotation(t *testing.T) {
	c := openTestValkey(t)
	ctx := context.Background()

	// token_version 讀寫自 DB(users.token_version);refresh token 走真實 Valkey。
	name := strings.ReplaceAll(t.Name(), "/", "_")
	db := enttest.Open(t, "sqlite3", "file:"+name+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	co := db.Company.Create().SetName("整合公司").SetIdentifier("IT-" + name).SaveX(ctx)
	u := db.User.Create().
		SetEmail("it-token-" + name + "@example.com").SetName("整合").SetStatus(user.StatusActive).
		SetRole("customer").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)
	t.Cleanup(func() {
		ctx := context.Background()
		c.Del(ctx, refreshUserSetKey(u.ID))
	})

	tm := NewTokenManager("it-secret", NewRedisStore(c), db)

	access, err := tm.IssueAccess(ctx, TokenSubject{UserID: u.ID, CompanyID: 1, Role: "customer"})
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	if _, err := tm.VerifyAccess(ctx, access); err != nil {
		t.Fatalf("VerifyAccess: %v", err)
	}

	r1, err := tm.IssueRefresh(ctx, u.ID, 0)
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}
	r2, err := tm.RotateRefresh(ctx, r1, u.ID, 0)
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
	if err := tm.RevokeAll(ctx, u.ID); err != nil {
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
