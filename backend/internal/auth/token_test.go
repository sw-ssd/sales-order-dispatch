package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // sqlite in-memory 測試驅動

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
)

// tokenTestSeq 為 newTokenTestEnv 產生唯一 DSN/identifier 的序號。
var tokenTestSeq atomic.Int64

// newTokenTestEnv 建立 enttest sqlite 記憶體 DB(company + user)與 TokenManager。
// token_version 由 users.token_version 欄位驅動;refresh token 存於 MemoryStore。
// DSN 以測試名+序號區隔,避免同 package 多測試(或同測試多次呼叫)共享同一 in-memory db。
func newTokenTestEnv(t *testing.T) (*TokenManager, *ent.User, context.Context) {
	t.Helper()
	name := fmt.Sprintf("%s-%d", strings.ReplaceAll(t.Name(), "/", "_"), tokenTestSeq.Add(1))
	dsn := "file:" + name + "?mode=memory&cache=shared&_fk=1"
	db := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()

	co := db.Company.Create().
		SetName("測試公司-" + name).
		SetIdentifier("T-" + name).
		SaveX(ctx)
	u := db.User.Create().
		SetEmail("tok-" + name + "@example.com").
		SetName("測試使用者").
		SetStatus(user.StatusActive).
		SetRole("staff").
		SetPasswordHash("x").
		SetCompanyID(co.ID).
		SaveX(ctx)
	tm := NewTokenManager("test-secret", NewMemoryStore(), db)
	return tm, u, ctx
}

func TestAccessTokenIssueVerify(t *testing.T) {
	tm, u, ctx := newTokenTestEnv(t)

	access, err := tm.IssueAccess(ctx, TokenSubject{UserID: u.ID, CompanyID: 3, DepartmentID: 5, Role: "customer"})
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	claims, err := tm.VerifyAccess(ctx, access)
	if err != nil {
		t.Fatalf("VerifyAccess: %v", err)
	}
	if claims.UserID != u.ID || claims.CompanyID != 3 || claims.DepartmentID != 5 || claims.Role != "customer" {
		t.Fatalf("claims 不符: %+v", claims)
	}
	if claims.TokenVersion != 0 {
		t.Fatalf("初始 token_version 應為 0(users.token_version 預設),got %d", claims.TokenVersion)
	}
	// token_version 讀自 DB 欄位
	if tv, err := tm.CurrentTokenVersion(ctx, u.ID); err != nil || tv != 0 {
		t.Fatalf("DB token_version 應為 0,got %d err=%v", tv, err)
	}
	exp := claims.ExpiresAt.Time
	if d := time.Until(exp); d > AccessTokenTTL || d < AccessTokenTTL-time.Minute {
		t.Fatalf("exp 應約 1h,got %v", d)
	}
}

func TestAccessTokenRevokedByBump(t *testing.T) {
	tm, u, ctx := newTokenTestEnv(t)

	access, err := tm.IssueAccess(ctx, TokenSubject{UserID: u.ID, Role: "staff"})
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	if err := tm.BumpTokenVersion(ctx, u.ID); err != nil {
		t.Fatalf("BumpTokenVersion: %v", err)
	}
	if _, err := tm.VerifyAccess(ctx, access); err != ErrTokenRevoked {
		t.Fatalf("bump 後舊 access 應回 ErrTokenRevoked,got %v", err)
	}
	// 新簽發的 access 帶新 tv,可通過
	access2, err := tm.IssueAccess(ctx, TokenSubject{UserID: u.ID, Role: "staff"})
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	if _, err := tm.VerifyAccess(ctx, access2); err != nil {
		t.Fatalf("新 access 應可驗證: %v", err)
	}
}

func TestRefreshIssueRotateRevoke(t *testing.T) {
	tm, u, ctx := newTokenTestEnv(t)

	r1, err := tm.IssueRefresh(ctx, u.ID, 0)
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}
	uid, tv, err := tm.VerifyRefresh(ctx, r1)
	if err != nil || uid != u.ID || tv != 0 {
		t.Fatalf("VerifyRefresh: uid=%d tv=%d err=%v", uid, tv, err)
	}

	// 旋轉:舊 token 立即失效(重放偵測)
	r2, err := tm.RotateRefresh(ctx, r1, u.ID, 0)
	if err != nil {
		t.Fatalf("RotateRefresh: %v", err)
	}
	if r2 == r1 {
		t.Fatal("旋轉後新 token 不得與舊相同")
	}
	if _, _, err := tm.VerifyRefresh(ctx, r1); err != ErrInvalidRefresh {
		t.Fatalf("旋轉後舊 refresh 重放應回 ErrInvalidRefresh,got %v", err)
	}

	// 撤銷單一 token(冪等)
	if err := tm.RevokeRefresh(ctx, r2); err != nil {
		t.Fatalf("RevokeRefresh: %v", err)
	}
	if err := tm.RevokeRefresh(ctx, r2); err != nil {
		t.Fatalf("RevokeRefresh 冪等: %v", err)
	}
	if _, _, err := tm.VerifyRefresh(ctx, r2); err != ErrInvalidRefresh {
		t.Fatalf("撤銷後 refresh 應失效,got %v", err)
	}
}

func TestRefreshRevokedByBumpAndRevokeAll(t *testing.T) {
	tm, u, ctx := newTokenTestEnv(t)

	r1, err := tm.IssueRefresh(ctx, u.ID, 0)
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}
	if _, err := tm.IssueRefresh(ctx, u.ID, 0); err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}
	// token_version bump → 全部 refresh 失效(簽發時 tv 與 DB 現值不符)
	if err := tm.BumpTokenVersion(ctx, u.ID); err != nil {
		t.Fatalf("Bump: %v", err)
	}
	if _, _, err := tm.VerifyRefresh(ctx, r1); err != ErrRefreshRevoked {
		t.Fatalf("bump 後 refresh 應回 ErrRefreshRevoked,got %v", err)
	}

	// RevokeAll:撤銷全部 refresh + tv+1
	tm2, u2, ctx2 := newTokenTestEnv(t)
	a, err := tm2.IssueRefresh(ctx2, u2.ID, 0)
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}
	b, err := tm2.IssueRefresh(ctx2, u2.ID, 0)
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}
	if err := tm2.RevokeAll(ctx2, u2.ID); err != nil {
		t.Fatalf("RevokeAll: %v", err)
	}
	for _, r := range []string{a, b} {
		if _, _, err := tm2.VerifyRefresh(ctx2, r); err != ErrInvalidRefresh {
			t.Fatalf("RevokeAll 後 refresh 應失效,got %v", err)
		}
	}
	// RevokeAll 使 DB 欄位遞增為 1
	tv, err := tm2.CurrentTokenVersion(ctx2, u2.ID)
	if err != nil || tv != 1 {
		t.Fatalf("RevokeAll 後 DB tv 應為 1,got %d err=%v", tv, err)
	}
}

func TestRefreshInvalidToken(t *testing.T) {
	tm, _, ctx := newTokenTestEnv(t)
	if _, _, err := tm.VerifyRefresh(ctx, "garbage"); err != ErrInvalidRefresh {
		t.Fatalf("亂碼 refresh 應 ErrInvalidRefresh,got %v", err)
	}
	if _, _, err := tm.VerifyRefresh(ctx, ""); err != ErrInvalidRefresh {
		t.Fatalf("空 refresh 應 ErrInvalidRefresh,got %v", err)
	}
}
// TestRotateRefreshConcurrentReplay P1-4 驗收:並發重放同一 refresh token 時,讀-刪-寫
// 以原子原語完成,恰一個請求消耗成功,其餘必回 ErrRefreshRevoked(第二次重放必拒)。
func TestRotateRefreshConcurrentReplay(t *testing.T) {
	tm, u, ctx := newTokenTestEnv(t)

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

// TestVerifyAccessEmptySecret P1-3 驗收:空 JWT 密鑰時 VerifyAccess 一律拒絕
// (fail-closed,防空鑰簽章繞過;正式環境由 Server.Init 擋下)。
func TestVerifyAccessEmptySecret(t *testing.T) {
	tm, u, ctx := newTokenTestEnv(t)
	empty := NewTokenManager("", NewMemoryStore(), tm.db)
	access, err := tm.IssueAccess(ctx, TokenSubject{UserID: u.ID, Role: "staff"})
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	if _, err := empty.VerifyAccess(ctx, access); err != ErrInvalidToken {
		t.Fatalf("空密鑰 VerifyAccess 應 ErrInvalidToken,got %v", err)
	}
	if _, err := empty.VerifyAccess(ctx, "bogus"); err != ErrInvalidToken {
		t.Fatalf("空密鑰 + 亂碼 token 應 ErrInvalidToken,got %v", err)
	}
}

// TestCurrentTokenVersionMissingUser 驗證 fail-closed:使用者不存在(已刪除)時
// token_version 讀取回 ErrInvalidToken,既有 token 無法通過。
func TestCurrentTokenVersionMissingUser(t *testing.T) {
	tm, _, ctx := newTokenTestEnv(t)
	if _, err := tm.CurrentTokenVersion(ctx, 999999); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("不存在使用者應回 ErrInvalidToken,got %v", err)
	}
}
