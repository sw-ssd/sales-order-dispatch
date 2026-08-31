package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OIDC state 與 registration token 效期。
const (
	StateTTL             = 10 * time.Minute
	RegistrationTokenTTL = 30 * time.Minute
	GoogleIssuerURL      = "https://accounts.google.com"
)

// OIDCIdentity 為 ID token 驗證通過後的身分(Google Workspace 員工)。
type OIDCIdentity struct {
	Email        string
	Name         string
	HostedDomain string // hd claim;可對應 companies.identifier 解析所屬公司
}

// OIDCVerifier 抽象 ID token 驗證,供測試以 fake 替換。
type OIDCVerifier interface {
	VerifyIDToken(ctx context.Context, rawIDToken string) (*OIDCIdentity, error)
}

// OAuthExchanger 抽象授權碼交換(回傳 raw ID token),供測試以 fake 替換。
type OAuthExchanger interface {
	Exchange(ctx context.Context, code string) (rawIDToken string, err error)
}

// NewGoogleOAuthConfig 組裝 Google OAuth2 config(scope: openid email profile)。
func NewGoogleOAuthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// googleExchanger 以 oauth2.Config 交換授權碼並抽取 id_token。
type googleExchanger struct {
	cfg *oauth2.Config
}

// NewGoogleOAuthExchanger 建立正式 OAuthExchanger。
func NewGoogleOAuthExchanger(cfg *oauth2.Config) OAuthExchanger {
	return &googleExchanger{cfg: cfg}
}

// Exchange 交換授權碼並回傳 raw ID token。
func (e *googleExchanger) Exchange(ctx context.Context, code string) (string, error) {
	tok, err := e.cfg.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("auth: 交換授權碼: %w", err)
	}
	raw, _ := tok.Extra("id_token").(string)
	if raw == "" {
		return "", errors.New("auth: 授權回應缺 id_token")
	}
	return raw, nil
}

// GoogleVerifier 以 Google OIDC discovery 驗證 ID token(簽章、aud、exp)。
type GoogleVerifier struct {
	v *oidc.IDTokenVerifier
}

// NewGoogleVerifier 建立 GoogleVerifier;需網路存取 Google discovery 端點。
func NewGoogleVerifier(ctx context.Context, clientID string) (*GoogleVerifier, error) {
	provider, err := oidc.NewProvider(ctx, GoogleIssuerURL)
	if err != nil {
		return nil, fmt.Errorf("auth: 載入 Google OIDC provider: %w", err)
	}
	return &GoogleVerifier{
		v: provider.Verifier(&oidc.Config{ClientID: clientID}),
	}, nil
}

// VerifyIDToken 驗證並抽取身分。
func (g *GoogleVerifier) VerifyIDToken(ctx context.Context, rawIDToken string) (*OIDCIdentity, error) {
	idtok, err := g.v.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("auth: ID token 驗證失敗: %w", err)
	}
	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Hd    string `json:"hd"`
	}
	if err := idtok.Claims(&claims); err != nil {
		return nil, fmt.Errorf("auth: 解析 ID token claims: %w", err)
	}
	if claims.Email == "" {
		return nil, errors.New("auth: ID token 缺 email")
	}
	return &OIDCIdentity{Email: claims.Email, Name: claims.Name, HostedDomain: claims.Hd}, nil
}

// NewState 產生隨機 CSRF state(一次性,存 KVStore)。
func NewState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: 產生 state: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// NewRegistrationToken 產生一次性 registration token(首次 OIDC 登入、未建帳號時)。
func NewRegistrationToken() (string, error) {
	return NewState()
}

// OneTimeStore 以 KVStore 實作一次性 token(state / registration token)儲存。
type OneTimeStore struct {
	kv KVStore
}

// NewOneTimeStore 建立 OneTimeStore。
func NewOneTimeStore(kv KVStore) *OneTimeStore {
	return &OneTimeStore{kv: kv}
}

// Put 存入一次性 token。
func (o *OneTimeStore) Put(ctx context.Context, key, value string, ttl time.Duration) error {
	return o.kv.Set(ctx, key, value, ttl)
}

// GetAndDelete 取出並刪除一次性 token(用完即刪;不存在 ok=false)。
func (o *OneTimeStore) GetAndDelete(ctx context.Context, key string) (string, bool, error) {
	v, ok, err := o.kv.Get(ctx, key)
	if err != nil || !ok {
		return "", ok, err
	}
	if err := o.kv.Delete(ctx, key); err != nil {
		return "", false, err
	}
	return v, true, nil
}

// StateKey 回傳 OAuth state 的儲存鍵。
func StateKey(state string) string {
	return "auth:state:" + state
}

// RegistrationKey 回傳 registration token 的儲存鍵。
func RegistrationKey(token string) string {
	return "auth:reg:" + token
}
