// Package operatorauth 為平台工具的認證(D38/S8):OIDC 限公司網域 + platform.operators 白名單,
// 簽發獨立 secret 的 operator JWT 置於 HttpOnly cookie。租戶 session/JWT 一律不適用。
//
// 兩層邊界是硬規則(設計 §2.2/§2.3):
//   - 不同 secret:租戶 token 不得通過平台 interceptor(反之亦然)。跨用等於平台工具門戶洞開,
//     production 由 Server.Init() 拒絕啟動來保證兩個 secret 不同。
//   - 不同 audience:`aud=platform` 使「同一 secret 誤用」也擋得住,而且租戶 app 的 Bearer
//     路徑本就不讀 platform_session cookie(名稱與 Path=/platform 都不重疊)。
package operatorauth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"

	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
)

// CookieName 為 operator session cookie;Path 限定 /platform,與租戶 session 不重疊。
const CookieName = "platform_session"

// StateCookieName 為 OIDC CSRF state 的短期 cookie(一次性,驗完即清)。
const StateCookieName = "platform_oidc_state"

// CookiePath 為 operator 兩個 cookie 的 Path:限定平台 API 與登入端點。
const CookiePath = "/platform"

// Audience 為 operator token 的固定 audience;租戶 token 不得帶此值,反之亦然。
const Audience = "platform"

// LoginPath／CallbackPath 為 OIDC 端點(掛載處需與 config 的 redirect URL 一致)。
const (
	LoginPath    = "/platform/auth/google"
	CallbackPath = "/platform/auth/google/callback"
)

// Identity 為通過驗證的平台操作者身分。
type Identity struct {
	OperatorID int64
	Email      string
	Role       string // operator | admin
}

type ctxKey struct{}

// WithIdentity 將 operator 身分注入 ctx(僅由 Interceptor 於驗證成功後呼叫)。
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// IdentityFrom 取出 operator 身分;平台 RPC 服務層以此做第二層檢查(不依賴 interceptor)。
func IdentityFrom(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(Identity)
	return id, ok
}

// Operator 為 platform.operators 的一列(白名單)。
type Operator struct {
	ID     int64
	Email  string
	Name   string
	Role   string
	Status string // active | disabled
}

// Store 為 operatorauth 需要的資料存取(postgres 實作走 admin 連線)。
type Store interface {
	// OperatorByEmail 查白名單;查無此 email 回 (nil, nil)——“不在名單”不是錯誤。
	OperatorByEmail(ctx context.Context, email string) (*Operator, error)
	TouchOperatorLogin(ctx context.Context, id int64, at time.Time) error
	AuditOperatorLogin(ctx context.Context, operatorID int64, email, ip, ua string) error
}

// Config 為 operatorauth 的設定(由 config.Platform 對應;OIDC 客戶端由 domains.go 組裝)。
type Config struct {
	Secret        string
	CookieDomain  string // 空 = host-only(開發環境;設 Domain=localhost 無效)
	ConsoleURL    string
	AllowedDomain string
	TokenLifetime time.Duration // 預設 12h
	// CookieSecure 由呼叫端依環境決定(production true)。http 的開發環境若設 Secure,
	// 瀏覽器在非 localhost 的 http 來源不會回送 cookie → 登入看似成功卻查不到身分。
	CookieSecure bool
}

// Service 提供 OIDC 登入端點與 RPC interceptor。
type Service struct {
	cfg   Config
	store Store
	oauth *oauth2.Config
	// exchanger／verifier 為既有 OIDC 封裝(internal/auth),可注入 fake。
	exchanger auth.OAuthExchanger
	verifier  auth.OIDCVerifier
}

// New 建立 Service(interceptor 與 IssueToken 只需 cfg 與 store)。
func New(cfg Config, st Store) *Service {
	if cfg.TokenLifetime == 0 {
		cfg.TokenLifetime = 12 * time.Hour
	}
	return &Service{cfg: cfg, store: st}
}

// WithOIDC 補上登入端點所需依賴;未設定時 Login／Callback 回 503(不掛載即可,不 panic)。
func (s *Service) WithOIDC(cfg *oauth2.Config, ex auth.OAuthExchanger, v auth.OIDCVerifier) *Service {
	s.oauth, s.exchanger, s.verifier = cfg, ex, v
	return s
}

// IssueToken 簽發 operator JWT(獨立 secret ＋ 固定 audience)。
func (s *Service) IssueToken(id Identity) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   id.OperatorID,
		"email": id.Email,
		"role":  id.Role,
		"aud":   Audience,
		"iat":   now.Unix(),
		"exp":   now.Add(s.cfg.TokenLifetime).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.Secret))
}

// verify 驗 token:簽章(僅 HS256)／audience／exp,再以 DB 白名單核對身分。
// **每一次請求都回查白名單**是刻意的:停用 operator 立即失效,不需要黑名單或短 TTL。
func (s *Service) verify(ctx context.Context, raw string) (Identity, error) {
	parsed, err := jwt.Parse(raw, func(*jwt.Token) (any, error) {
		return []byte(s.cfg.Secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithAudience(Audience), jwt.WithExpirationRequired())
	if err != nil || !parsed.Valid {
		return Identity{}, fmt.Errorf("operator token 無效: %w", err)
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return Identity{}, errors.New("operator token 內容無效")
	}
	sub, err := subjectID(claims["sub"])
	if err != nil {
		return Identity{}, err
	}
	email, _ := claims["email"].(string)
	if email == "" {
		return Identity{}, errors.New("operator token 缺 email")
	}
	op, err := s.store.OperatorByEmail(ctx, email)
	if err != nil {
		return Identity{}, fmt.Errorf("查詢 operator 白名單: %w", err)
	}
	if op == nil || op.Status != "active" || !strings.EqualFold(op.Email, email) {
		return Identity{}, errors.New("operator 不存在或已停用")
	}
	// 完整性:身分取自 white list 的 id,且必須與 token 自稱的 sub 一致——
	// 否則「A 的 token ＋ B 的 email」會被查回 B 的身分,等於可用被停用者的 token 冒充他人。
	if op.ID != sub {
		return Identity{}, errors.New("operator token 的 sub 與白名單不符")
	}
	return Identity{OperatorID: op.ID, Email: op.Email, Role: op.Role}, nil
}

// subjectID 解析 sub 為 int64。本套件簽出的是數字;JSON 解碼後為 float64,故兩種都接受,
// 不接受字串(避免 "1"／"01" 這類等價不同形的繞道)。
func subjectID(v any) (int64, error) {
	switch n := v.(type) {
	case int64:
		return n, nil
	case float64:
		if n != float64(int64(n)) {
			return 0, errors.New("operator token 的 sub 非整數")
		}
		return int64(n), nil
	default:
		return 0, errors.New("operator token 缺 sub 或型別不符")
	}
}

// Login 導向 Google OIDC;state 存短期 cookie(比對在 Callback,防 CSRF)。
func (s *Service) Login(w http.ResponseWriter, r *http.Request) {
	if s.oauth == nil {
		http.Error(w, "平台登入未設定", http.StatusServiceUnavailable)
		return
	}
	state, err := auth.NewState()
	if err != nil {
		http.Error(w, "無法產生 OIDC state", http.StatusInternalServerError)
		return
	}
	s.setStateCookie(w, state, int(auth.StateTTL.Seconds()))
	http.Redirect(w, r, s.oauth.AuthCodeURL(state), http.StatusFound)
}

// Callback 完成 OIDC:比對 state → 換 id_token → 驗網域 → 查白名單 → 簽 token → 設 cookie → 導回 console。
// 每一步失敗都不得簽發 token;對外錯誤訊息不含 token 內容。
func (s *Service) Callback(w http.ResponseWriter, r *http.Request) {
	if s.oauth == nil || s.exchanger == nil || s.verifier == nil {
		http.Error(w, "平台登入未設定", http.StatusServiceUnavailable)
		return
	}
	q := r.URL.Query()
	stateCookie, err := r.Cookie(StateCookieName)
	// 一次性:不論後續成敗,驗過就清(缺 cookie 時沒有可清的也無妨)。
	s.setStateCookie(w, "", -1)
	if err != nil || stateCookie.Value == "" ||
		subtle.ConstantTimeCompare([]byte(q.Get("state")), []byte(stateCookie.Value)) != 1 {
		http.Error(w, "OIDC state 不符", http.StatusBadRequest)
		return
	}

	rawIDToken, err := s.exchanger.Exchange(r.Context(), q.Get("code"))
	if err != nil {
		http.Error(w, "授權碼交換失敗", http.StatusUnauthorized)
		return
	}
	ident, err := s.verifier.VerifyIDToken(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, "ID token 驗證失敗", http.StatusUnauthorized)
		return
	}

	// 網域限制:只接受公司 Workspace 網域(不分大小寫;取 @ 之後的部分)。
	at := strings.LastIndex(ident.Email, "@")
	if at < 0 || !strings.EqualFold(ident.Email[at+1:], s.cfg.AllowedDomain) {
		http.Error(w, "此帳號不屬於允許的網域", http.StatusForbidden)
		return
	}

	op, err := s.store.OperatorByEmail(r.Context(), ident.Email)
	if err != nil {
		http.Error(w, "查詢操作者失敗", http.StatusInternalServerError)
		return
	}
	if op == nil || op.Status != "active" {
		http.Error(w, "此帳號不在平台操作者名單中", http.StatusForbidden)
		return
	}

	// 先簽、先留痕,最後才發 cookie:任何一步失敗都不會留下「已登入卻沒稽核」的狀態。
	token, err := s.IssueToken(Identity{OperatorID: op.ID, Email: op.Email, Role: op.Role})
	if err != nil {
		http.Error(w, "簽發 token 失敗", http.StatusInternalServerError)
		return
	}
	if err := s.store.TouchOperatorLogin(r.Context(), op.ID, time.Now()); err != nil {
		http.Error(w, "更新登入時間失敗", http.StatusInternalServerError)
		return
	}
	// 登入本身也是稽核對象:誰、何時、來自哪裡(來源取 X-Forwarded-For 首項)。
	if err := s.store.AuditOperatorLogin(r.Context(), op.ID, op.Email, clientIP(r), r.UserAgent()); err != nil {
		http.Error(w, "寫入稽核失敗", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: CookieName, Value: token, Path: CookiePath, Domain: s.cfg.CookieDomain,
		HttpOnly: true, Secure: s.cfg.CookieSecure, SameSite: http.SameSiteLaxMode,
		MaxAge: int(s.cfg.TokenLifetime.Seconds()),
	})
	http.Redirect(w, r, s.cfg.ConsoleURL, http.StatusFound)
}

// setStateCookie 設定／清除 OIDC state cookie(maxAge<0 = 清除)。
func (s *Service) setStateCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name: StateCookieName, Value: value, Path: CookiePath, Domain: s.cfg.CookieDomain,
		HttpOnly: true, Secure: s.cfg.CookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: maxAge,
	})
}

// clientIP 取真實來源 IP(X-Forwarded-For 首項,否則 RemoteAddr)。
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RedirectURL 由租戶 OIDC 的 redirect URL 推導平台回呼網址:兩者同一個 API 來源,
// 只差路徑(scheme/host 直接沿用,故不需要另一個環境變數)。
func RedirectURL(tenantRedirectURL string) (string, error) {
	u, err := url.Parse(tenantRedirectURL)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("OIDC redirect URL 缺少 scheme/host: %q", tenantRedirectURL)
	}
	u.Path, u.RawQuery, u.Fragment = CallbackPath, "", ""
	return u.String(), nil
}
