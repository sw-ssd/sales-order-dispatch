// Package qrcode 客戶 QR 簽章 token(04 計畫 Task 3.8.1):產生與驗證。
// payload 為 company_id + customer_code + exp + jti + purpose(定位用,非憑證);
// 簽章 HMAC-SHA256,密鑰由環境注入(JWT_SECRET 複用,不進版控)。
// 一次性由 Valkey jti 鍵承載;Valkey 不可用降級為僅驗簽章與效期並記告警。
package qrcode

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// Purpose 為 token 用途固定值(防與其他簽章 token 混用)。
const Purpose = "customer-qr-login"

// DefaultTTL 為 token 效期預設 30 天(供印製/轉發場景),由呼叫端覆寫。
const DefaultTTL = 30 * 24 * time.Hour

// ErrTokenUsed 表示該 jti 已被兌換過(一次性保證)。
// 呼叫端一律以 errors.Is 判別,不得比對訊息文字——訊息可能被改寫,而「已使用」
// 被誤判成 KV 錯誤會把本該拒絕的兌換放行成降級路徑。
var ErrTokenUsed = errors.New("qr: token 已使用")

// Claims 為 QR token payload。
type Claims struct {
	CompanyID    int    `json:"company_id"`
	CustomerCode string `json:"customer_code"`
	Exp          int64  `json:"exp"`
	JTI          string `json:"jti"`
	Purpose      string `json:"purpose"`
}

// Generate 產生簽章 token:payload JSON → base64url.payload + "." + hex(HMAC)。
// secret 不得為空;ttl <= 0 即 DefaultTTL。
func Generate(secret string, companyID int, customerCode string, ttl time.Duration) (token, jti string, err error) {
	if strings.TrimSpace(secret) == "" {
		return "", "", errcode.SysInternal.Error(nil)
	}
	if companyID <= 0 || strings.TrimSpace(customerCode) == "" {
		return "", "", errcode.SysInvalidArgument.Error(map[string]string{"field": "customer"})
	}
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", "", errcode.SysInternal.Error(nil)
	}
	c := Claims{
		CompanyID: companyID, CustomerCode: strings.TrimSpace(customerCode),
		Exp: time.Now().Add(ttl).Unix(), JTI: hex.EncodeToString(b[:]), Purpose: Purpose,
	}
	raw, err := json.Marshal(c)
	if err != nil {
		return "", "", errcode.SysInternal.Error(nil)
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return payload + "." + sig, c.JTI, nil
}

// Verified 為驗證通過的 claims。
type Verified = Claims

// Verify 驗證 token:格式 → 簽章(恆時比對) → purpose → exp → claims。
// 簽章/purpose 不符一律 unauthenticated(不區分,防探測);過期為 failed_precondition。
func Verify(secret, token string) (*Verified, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errcode.SysInternal.Error(nil)
	}
	payload, sig, ok := strings.Cut(strings.TrimSpace(token), ".")
	if !ok || payload == "" || sig == "" {
		return nil, errcode.AuthUnauthenticated.Error(nil)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	want := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(strings.ToLower(sig)), []byte(want)) {
		return nil, errcode.AuthUnauthenticated.Error(nil)
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, errcode.AuthUnauthenticated.Error(nil)
	}
	var c Claims
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, errcode.AuthUnauthenticated.Error(nil)
	}
	if c.Purpose != Purpose {
		return nil, errcode.AuthUnauthenticated.Error(nil)
	}
	if c.CompanyID <= 0 || c.CustomerCode == "" {
		return nil, errcode.AuthUnauthenticated.Error(nil)
	}
	if time.Now().Unix() > c.Exp {
		return nil, errcode.SysInvalidArgument.Error(map[string]string{"field": "token"})
	}
	return &c, nil
}

// JTIStore 為一次性標記承載(auth.KVStore:以 key 不存在判未用,存在判已用)。
// Valkey 不可用時呼叫端降級(僅驗簽章與效期並記告警),此處不吞錯誤。
type JTIStore interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}

// JTIKey 拼 jti 儲存鍵(命名空間防與其他一次性 token 混用)。
func JTIKey(jti string) string { return "qr-jti:" + jti }

// ClaimOnce 先佔位 jti(防併發雙兌換):有值即已使用;無值即寫入佔位。
// 佔位與讀取非原子(Valkey 無 PutNX 介面):併發雙兌換的殘餘競態由「先標記再回應」
// 縮到最小,真撞上時兩邊都回成功但 jti 同值,稽核可追(見 handler 註解)。
func ClaimOnce(ctx context.Context, kv JTIStore, jti string, ttl time.Duration) error {
	_, ok, err := kv.Get(ctx, JTIKey(jti))
	if err != nil {
		return err
	}
	if ok {
		return ErrTokenUsed
	}
	return kv.Set(ctx, JTIKey(jti), "1", ttl)
}

// token 過期剩餘時間(供 jti TTL ≥ 剩餘效期)。
func Remaining(c *Claims) time.Duration {
	d := time.Until(time.Unix(c.Exp, 0))
	if d < 0 {
		return 0
	}
	return d
}
