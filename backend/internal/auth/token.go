package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/salesorder/sales-order-1.0/backend/ent"
)

// Token 效期(D5):access 1 小時;refresh 30 天、使用時旋轉。
const (
	AccessTokenTTL  = time.Hour
	RefreshTokenTTL = 30 * 24 * time.Hour
)

// 錯誤 sentinel:handler 依此映射 Connect code。
var (
	ErrInvalidToken       = errors.New("auth: token 無效")
	ErrTokenRevoked       = errors.New("auth: token 已撤銷")
	ErrInvalidRefresh     = errors.New("auth: refresh token 無效")
	ErrRefreshRevoked     = errors.New("auth: refresh token 已撤銷")
	ErrAccountLocked      = errors.New("auth: 帳號已鎖定")
	ErrAccountInactive    = errors.New("auth: 帳號未啟用")
	ErrInvalidCredentials = errors.New("auth: 帳號或密碼錯誤")
)

// Claims 為 access token 內容;tv 對應 token_version(撤銷比對用,存放於 DB 的
// users.token_version 欄位)。
type Claims struct {
	UserID       int    `json:"sub"`
	Role         string `json:"role"`
	CompanyID    int    `json:"cid,omitempty"`
	DepartmentID int    `json:"did,omitempty"`
	TokenVersion int    `json:"tv"`
	jwt.RegisteredClaims
}

// TokenSubject 為簽發 access token 所需的使用者身分快照。
type TokenSubject struct {
	UserID       int
	CompanyID    int
	DepartmentID int // 0 = 無部門
	Role         string
}

// TokenManager 負責 access token 簽發/驗證、refresh token 發行/旋轉/撤銷、token_version 管理。
// token_version 存於 DB(users.token_version);refresh token 存於 KVStore(Valkey 正式 / MemoryStore 測試)。
type TokenManager struct {
	secret []byte
	kv     KVStore
	db     *ent.Client
}

// NewTokenManager 建立 TokenManager。
func NewTokenManager(jwtSecret string, kv KVStore, db *ent.Client) *TokenManager {
	return &TokenManager{secret: []byte(jwtSecret), kv: kv, db: db}
}

// refreshKey 回傳 refresh token 雜湊的鍵。
func refreshKey(hash string) string {
	return "auth:refresh:" + hash
}

// refreshUserSetKey 回傳使用者名下 refresh token 雜湊集合鍵(RevokeAll 用)。
func refreshUserSetKey(userID int) string {
	return fmt.Sprintf("auth:refresh:user:%d", userID)
}

// CurrentTokenVersion 回傳使用者目前 token_version(users.token_version)。
// 使用者不存在時回傳 ErrInvalidToken(fail-closed:已刪除使用者的 token 一律失效)。
func (m *TokenManager) CurrentTokenVersion(ctx context.Context, userID int) (int, error) {
	u, err := m.db.User.Get(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, fmt.Errorf("%w: 使用者 %d 不存在", ErrInvalidToken, userID)
		}
		return 0, err
	}
	return u.TokenVersion, nil
}

// BumpTokenVersion 遞增 users.token_version,使該使用者既有 access/refresh token 全數失效(D5)。
// 改密碼 / 停用 / 角色變更 / 強制登出時呼叫。
func (m *TokenManager) BumpTokenVersion(ctx context.Context, userID int) error {
	_, err := m.db.User.UpdateOneID(userID).
		AddTokenVersion(1).
		Save(ctx)
	return err
}

// IssueAccess 簽發 HS256 access token(exp 1h,claim 含 tv)。
func (m *TokenManager) IssueAccess(ctx context.Context, s TokenSubject) (string, error) {
	tv, err := m.CurrentTokenVersion(ctx, s.UserID)
	if err != nil {
		return "", err
	}
	now := time.Now()
	claims := Claims{
		UserID:       s.UserID,
		Role:         s.Role,
		CompanyID:    s.CompanyID,
		DepartmentID: s.DepartmentID,
		TokenVersion: tv,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(s.UserID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(m.secret)
}

// VerifyAccess 驗證 access token(簽章、exp、tv 與目前 token_version 比對)。
func (m *TokenManager) VerifyAccess(ctx context.Context, token string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(_ *jwt.Token) (any, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !parsed.Valid {
		return nil, ErrInvalidToken
	}
	live, err := m.CurrentTokenVersion(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if claims.TokenVersion != live {
		return nil, ErrTokenRevoked
	}
	return claims, nil
}

// newRefreshToken 產生 32-byte 隨機 refresh token(plaintext)與其 SHA-256 雜湊。
func newRefreshToken() (plain, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	plain = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(plain))
	return plain, hex.EncodeToString(sum[:]), nil
}

// IssueRefresh 發行 refresh token:後端僅存 SHA-256 雜湊,綁定簽發時 token_version。
func (m *TokenManager) IssueRefresh(ctx context.Context, userID, tokenVersion int) (string, error) {
	plain, hash, err := newRefreshToken()
	if err != nil {
		return "", err
	}
	value := fmt.Sprintf("%d:%d", userID, tokenVersion)
	if err := m.kv.Set(ctx, refreshKey(hash), value, RefreshTokenTTL); err != nil {
		return "", err
	}
	if err := m.kv.SetAdd(ctx, refreshUserSetKey(userID), hash); err != nil {
		return "", err
	}
	return plain, nil
}

// VerifyRefresh 查驗 refresh token 存在、未被撤銷,且使用者 token_version 未變(撤銷比對);
// 回傳綁定的 userID 與 token_version。
func (m *TokenManager) VerifyRefresh(ctx context.Context, plain string) (int, int, error) {
	hash, err := hashRefreshToken(plain)
	if err != nil {
		return 0, 0, ErrInvalidRefresh
	}
	v, ok, err := m.kv.Get(ctx, refreshKey(hash))
	if err != nil {
		return 0, 0, err
	}
	if !ok {
		return 0, 0, ErrInvalidRefresh
	}
	userID, tv, err := parseRefreshValue(v)
	if err != nil {
		return 0, 0, ErrInvalidRefresh
	}
	live, err := m.CurrentTokenVersion(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	if live != tv {
		return 0, 0, ErrRefreshRevoked
	}
	return userID, tv, nil
}

// RotateRefresh 作廢舊 refresh token 並發行新的(旋轉制;舊 token 重放會被拒)。
func (m *TokenManager) RotateRefresh(ctx context.Context, plain string, userID, tokenVersion int) (string, error) {
	hash, err := hashRefreshToken(plain)
	if err != nil {
		return "", ErrInvalidRefresh
	}
	if err := m.kv.Delete(ctx, refreshKey(hash)); err != nil {
		return "", err
	}
	if err := m.kv.SetRemove(ctx, refreshUserSetKey(userID), hash); err != nil {
		return "", err
	}
	return m.IssueRefresh(ctx, userID, tokenVersion)
}

// RevokeRefresh 撤銷單一 refresh token(冪等;不存在亦成功)。
func (m *TokenManager) RevokeRefresh(ctx context.Context, plain string) error {
	hash, err := hashRefreshToken(plain)
	if err != nil {
		return nil // 格式不符視同已撤銷,Logout 保持冪等
	}
	v, ok, err := m.kv.Get(ctx, refreshKey(hash))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	userID, _, err := parseRefreshValue(v)
	if err == nil {
		_ = m.kv.SetRemove(ctx, refreshUserSetKey(userID), hash)
	}
	return m.kv.Delete(ctx, refreshKey(hash))
}

// RevokeAll 撤銷使用者全部 refresh token 並遞增 token_version(強制登出)。
func (m *TokenManager) RevokeAll(ctx context.Context, userID int) error {
	members, err := m.kv.SetMembers(ctx, refreshUserSetKey(userID))
	if err != nil {
		return err
	}
	for _, h := range members {
		if err := m.kv.Delete(ctx, refreshKey(h)); err != nil {
			return err
		}
	}
	if err := m.kv.Delete(ctx, refreshUserSetKey(userID)); err != nil {
		return err
	}
	return m.BumpTokenVersion(ctx, userID)
}

func hashRefreshToken(plain string) (string, error) {
	if plain == "" {
		return "", ErrInvalidRefresh
	}
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:]), nil
}

func parseRefreshValue(v string) (userID, tokenVersion int, err error) {
	parts := strings.SplitN(v, ":", 2)
	if len(parts) != 2 {
		return 0, 0, errors.New("auth: refresh 值格式錯誤")
	}
	userID, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	tokenVersion, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return userID, tokenVersion, nil
}
