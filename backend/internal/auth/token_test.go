package auth

import (
	"context"
	"testing"
	"time"
)

func TestAccessTokenIssueVerify(t *testing.T) {
	ctx := context.Background()
	tm := NewTokenManager("test-secret", NewMemoryStore())

	access, err := tm.IssueAccess(ctx, TokenSubject{UserID: 7, CompanyID: 3, DepartmentID: 5, Role: "customer"})
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	claims, err := tm.VerifyAccess(ctx, access)
	if err != nil {
		t.Fatalf("VerifyAccess: %v", err)
	}
	if claims.UserID != 7 || claims.CompanyID != 3 || claims.DepartmentID != 5 || claims.Role != "customer" {
		t.Fatalf("claims 不符: %+v", claims)
	}
	if claims.TokenVersion != 0 {
		t.Fatalf("初始 token_version 應為 0,got %d", claims.TokenVersion)
	}
	exp := claims.ExpiresAt.Time
	if d := time.Until(exp); d > AccessTokenTTL || d < AccessTokenTTL-time.Minute {
		t.Fatalf("exp 應約 1h,got %v", d)
	}
}

func TestAccessTokenRevokedByBump(t *testing.T) {
	ctx := context.Background()
	tm := NewTokenManager("test-secret", NewMemoryStore())

	access, err := tm.IssueAccess(ctx, TokenSubject{UserID: 1, Role: "staff"})
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	if err := tm.BumpTokenVersion(ctx, 1); err != nil {
		t.Fatalf("BumpTokenVersion: %v", err)
	}
	if _, err := tm.VerifyAccess(ctx, access); err != ErrTokenRevoked {
		t.Fatalf("bump 後舊 access 應回 ErrTokenRevoked,got %v", err)
	}
	// 新簽發的 access 帶新 tv,可通過
	access2, err := tm.IssueAccess(ctx, TokenSubject{UserID: 1, Role: "staff"})
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}
	if _, err := tm.VerifyAccess(ctx, access2); err != nil {
		t.Fatalf("新 access 應可驗證: %v", err)
	}
}

func TestRefreshIssueRotateRevoke(t *testing.T) {
	ctx := context.Background()
	tm := NewTokenManager("test-secret", NewMemoryStore())

	r1, err := tm.IssueRefresh(ctx, 42, 0)
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}
	uid, tv, err := tm.VerifyRefresh(ctx, r1)
	if err != nil || uid != 42 || tv != 0 {
		t.Fatalf("VerifyRefresh: uid=%d tv=%d err=%v", uid, tv, err)
	}

	// 旋轉:舊 token 立即失效(重放偵測)
	r2, err := tm.RotateRefresh(ctx, r1, 42, 0)
	if err != nil {
		t.Fatalf("RotateRefresh: %v", err)
	}
	if r2 == r1 {
		t.Fatal("旋轉後新 token 不得與舊相同")
	}
	if _, _, err := tm.VerifyRefresh(ctx, r1); err != ErrInvalidRefresh {
		t.Fatalf("旋轉後舊 refresh 應失效,got %v", err)
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
	ctx := context.Background()
	tm := NewTokenManager("test-secret", NewMemoryStore())

	r1, err := tm.IssueRefresh(ctx, 9, 0)
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}
	if _, err := tm.IssueRefresh(ctx, 9, 0); err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}
	// token_version bump → 全部 refresh 失效(簽發時 tv 與現值不符)
	if err := tm.BumpTokenVersion(ctx, 9); err != nil {
		t.Fatalf("Bump: %v", err)
	}
	if _, _, err := tm.VerifyRefresh(ctx, r1); err != ErrRefreshRevoked {
		t.Fatalf("bump 後 refresh 應回 ErrRefreshRevoked,got %v", err)
	}

	// RevokeAll:撤銷全部 refresh + tv+1
	tm2 := NewTokenManager("test-secret", NewMemoryStore())
	a, err := tm2.IssueRefresh(ctx, 9, 0)
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}
	b, err := tm2.IssueRefresh(ctx, 9, 0)
	if err != nil {
		t.Fatalf("IssueRefresh: %v", err)
	}
	if err := tm2.RevokeAll(ctx, 9); err != nil {
		t.Fatalf("RevokeAll: %v", err)
	}
	for _, r := range []string{a, b} {
		if _, _, err := tm2.VerifyRefresh(ctx, r); err != ErrInvalidRefresh {
			t.Fatalf("RevokeAll 後 refresh 應失效,got %v", err)
		}
	}
	tv, err := tm2.CurrentTokenVersion(ctx, 9)
	if err != nil || tv != 1 {
		t.Fatalf("RevokeAll 後 tv 應為 1,got %d err=%v", tv, err)
	}
}

func TestRefreshInvalidToken(t *testing.T) {
	ctx := context.Background()
	tm := NewTokenManager("test-secret", NewMemoryStore())
	if _, _, err := tm.VerifyRefresh(ctx, "garbage"); err != ErrInvalidRefresh {
		t.Fatalf("亂碼 refresh 應 ErrInvalidRefresh,got %v", err)
	}
	if _, _, err := tm.VerifyRefresh(ctx, ""); err != ErrInvalidRefresh {
		t.Fatalf("空 refresh 應 ErrInvalidRefresh,got %v", err)
	}
}
