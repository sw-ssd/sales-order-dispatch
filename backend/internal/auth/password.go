package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

// 密碼政策:連續 5 次錯誤鎖定 30 分鐘(T12;密碼以 Argon2id 雜湊儲存,對齊 identity-access 規格)。
const (
	MaxLoginFailures = 5
	LockDuration     = 30 * time.Minute

	// OIDCPasswordSentinel 為僅 OIDC 登入(員工)帳號的 password_hash 佔位值。
	// 員工走 Google 登入不存密碼(規格 4.1);sentinel 非合法 Argon2id,密碼登入必失敗。
	OIDCPasswordSentinel = "!"
)

// Argon2id 參數(OWASP 建議等級)。
const (
	argon2Time    uint32 = 1
	argon2Memory  uint32 = 64 * 1024 // 64 MiB
	argon2Threads uint8  = 4
	argon2KeyLen  uint32 = 32
	argon2SaltLen        = 16
)

// HashPassword 以 Argon2id 雜湊明文密碼,編碼格式(對齊 PHP password_hash):
// $argon2id$v=19$m=<memory>,t=<time>,p=<threads>$<salt_b64>$<key_b64>
func HashPassword(plain string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("auth: 產生 salt 失敗: %w", err)
	}
	key := argon2.IDKey([]byte(plain), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Time, argon2Threads, b64Salt, b64Key), nil
}

// VerifyPassword 驗證密碼是否與 Argon2id 雜湊相符;非合法 Argon2id 格式(如 OIDC sentinel、
// 舊 bcrypt、空值)一律視為不符(fail-closed)。
func VerifyPassword(hash, plain string) bool {
	parts := strings.Split(hash, "$")
	// ["", "argon2id", "v=19", "m=..,t=..,p=..", salt, key] -> len 6
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	m, t, p, err := parseArgon2Params(parts[3])
	if err != nil {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}
	expected := argon2.IDKey([]byte(plain), salt, t, m, p, uint32(len(key)))
	return subtle.ConstantTimeCompare(expected, key) == 1
}

// parseArgon2Params 解析 "m=...,t=...,p=..." 參數段。
func parseArgon2Params(s string) (memory, time uint32, threads uint8, err error) {
	for _, kv := range strings.Split(s, ",") {
		kv = strings.TrimSpace(kv)
		key, val, _ := strings.Cut(kv, "=")
		switch key {
		case "m":
			memory, err = parseUint32(val)
		case "t":
			time, err = parseUint32(val)
		case "p":
			var n uint64
			n, err = strconv.ParseUint(val, 10, 8)
			threads = uint8(n)
		}
		if err != nil {
			return 0, 0, 0, err
		}
	}
	if memory == 0 || time == 0 || threads == 0 {
		return 0, 0, 0, errors.New("auth: 非法 argon2 參數")
	}
	return memory, time, threads, nil
}

func parseUint32(s string) (uint32, error) {
	n, err := strconv.ParseUint(s, 10, 32)
	return uint32(n), err
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

// tempPasswordChars 為臨時密碼字元集(去除易混淆的 0/O/1/l/I, A3 1.5.2/1.5.4)。
const tempPasswordChars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789"

// TempPasswordLength 為臨時密碼長度(≥ 12,規格 1.5.2)。
const TempPasswordLength = 12

// GenerateTempPassword 產生隨機臨時密碼(≥ 12 字元,字母+數字)。
func GenerateTempPassword() (string, error) {
	b := make([]byte, TempPasswordLength)
	max := big.NewInt(int64(len(tempPasswordChars)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("auth: 產生臨時密碼失敗: %w", err)
		}
		b[i] = tempPasswordChars[n.Int64()]
	}
	return string(b), nil
}
