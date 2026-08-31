package auth

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// 密碼政策:連續 5 次錯誤鎖定 30 分鐘(T12;密碼以 bcrypt 雜湊儲存)。
const (
	MaxLoginFailures = 5
	LockDuration     = 30 * time.Minute

	// OIDCPasswordSentinel 為僅 OIDC 登入(員工)帳號的 password_hash 佔位值。
	// 員工走 Google 登入不存密碼(規格 4.1);sentinel 非合法 bcrypt,密碼登入必失敗。
	OIDCPasswordSentinel = "!"
)

// HashPassword 以 bcrypt 雜湊明文密碼。
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("auth: 密碼雜湊失敗: %w", err)
	}
	return string(b), nil
}

// VerifyPassword 驗證密碼是否與雜湊相符;儲存值非合法 bcrypt(如 OIDC 帳號)一律視為不符。
func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// loginFailKey 回傳登入失敗計數鍵(以 customer_code 為鍵,不區分帳號是否存在,避免列舉)。
func loginFailKey(account string) string {
	return "auth:login:fail:" + account
}

// LoginLock 以 KVStore 實作失敗計數與鎖定:5 次錯誤鎖 30 分鐘。
type LoginLock struct {
	kv KVStore
}

// NewLoginLock 建立 LoginLock。
func NewLoginLock(kv KVStore) *LoginLock {
	return &LoginLock{kv: kv}
}

// RecordFailure 記錄一次失敗並回傳累計失敗次數;首次失敗起算 30 分鐘窗口。
func (l *LoginLock) RecordFailure(ctx context.Context, account string) (int, error) {
	key := loginFailKey(account)
	n, err := l.kv.Incr(ctx, key)
	if err != nil {
		return 0, err
	}
	if n == 1 {
		if err := l.kv.Expire(ctx, key, LockDuration); err != nil {
			return 0, err
		}
	}
	return int(n), nil
}

// IsLocked 檢查帳號是否已達鎖定門檻。
func (l *LoginLock) IsLocked(ctx context.Context, account string) (bool, error) {
	v, ok, err := l.kv.Get(ctx, loginFailKey(account))
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return false, fmt.Errorf("auth: 失敗計數格式錯誤: %w", err)
	}
	return n >= MaxLoginFailures, nil
}

// Clear 清除失敗記錄(登入成功時)。
func (l *LoginLock) Clear(ctx context.Context, account string) error {
	return l.kv.Delete(ctx, loginFailKey(account))
}
