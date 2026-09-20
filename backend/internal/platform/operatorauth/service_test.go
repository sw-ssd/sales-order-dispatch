package operatorauth_test

// 平台操作者認證的四條安全契約（RED 先行）：
// ① 租戶 secret 簽出的 token 不得通過平台 interceptor（跨用＝可冒充平台操作者）
// ② audience 不符（租戶 token 常無 aud=platform）拒絕
// ③ operators.status=disabled 立即失效（無需黑名單）
// ④ 正向對照：有效 token 可通過且 handler 內取得身分
//
// 另補三類邊界：sub 與白名單 id 不一致、operator token 走 Bearer（非 cookie）不得成立、
// cookie 旗標（HttpOnly／Secure／SameSite／Path／Domain）與 Callback 的逐步拒絕（失敗不得簽發 cookie）。

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/operatorauth"
)

const testSecret = "platform-secret"

// fakeStore 以記憶體提供白名單查詢（三個安全契約都不需要真 DB）。
type fakeStore struct {
	op *operatorauth.Operator
	// 登入副作用的觀測點：Callback 成功必須留下登入時間與稽核。
	touched    int64
	audited    string
	auditIP    string
	auditUA    string
	failAudit  error
	lookupFail error
}

func (f *fakeStore) OperatorByEmail(context.Context, string) (*operatorauth.Operator, error) {
	if f.lookupFail != nil {
		return nil, f.lookupFail
	}
	return f.op, nil
}

func (f *fakeStore) TouchOperatorLogin(_ context.Context, id int64, _ time.Time) error {
	f.touched = id
	return nil
}

func (f *fakeStore) AuditOperatorLogin(_ context.Context, id int64, email, ip, ua string) error {
	if f.failAudit != nil {
		return f.failAudit
	}
	f.audited = email
	f.auditIP = ip
	f.auditUA = ua
	return nil
}

func activeStore() *fakeStore {
	return &fakeStore{op: &operatorauth.Operator{
		ID: 1, Email: "ops@example.com", Name: "Ops", Role: "admin", Status: "active"}}
}

func newTestService(t *testing.T, st *fakeStore) *operatorauth.Service {
	t.Helper()
	return operatorauth.New(operatorauth.Config{Secret: testSecret}, st)
}

// callWithCookie 以指定 cookie 值走 interceptor；handler 被呼叫即視為失敗。
func callWithCookie(t *testing.T, svc *operatorauth.Service, value string) error {
	t.Helper()
	req := connect.NewRequest(&emptypb.Empty{})
	req.Header().Set("Cookie", operatorauth.CookieName+"="+value)
	_, err := svc.Interceptor().WrapUnary(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
		t.Fatal("未通過驗證的請求不得進入 handler")
		return nil, nil
	})(context.Background(), req)
	return err
}

// ① 租戶 secret 簽出的 token 不得通過平台 interceptor（跨用等於可冒充平台操作者）。
func TestTenantTokenRejected(t *testing.T) {
	svc := newTestService(t, activeStore())
	token := signHS256(t, "tenant-secret", jwt.MapClaims{"sub": 1, "aud": "platform"})
	if err := callWithCookie(t, svc, token); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("租戶 secret 簽出的 token 必須 Unauthenticated，got %v", err)
	}
}

// ② 即使 secret 相同，audience 不符（租戶 token 常無 aud=platform）也拒絕。
func TestWrongAudienceRejected(t *testing.T) {
	svc := newTestService(t, activeStore())
	token := signHS256(t, testSecret, jwt.MapClaims{"sub": 1, "aud": "tenant"})
	if err := callWithCookie(t, svc, token); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("audience 不符必須 Unauthenticated，got %v", err)
	}
}

// ③ secret 與 audience 都對，但 operators.status = disabled → 立即失效（不需黑名單）。
func TestDisabledOperatorRejected(t *testing.T) {
	st := activeStore()
	st.op.Status = "disabled"
	svc := newTestService(t, st)
	token, err := svc.IssueToken(operatorauth.Identity{OperatorID: 1, Email: "ops@example.com", Role: "admin"})
	if err != nil {
		t.Fatalf("簽 token: %v", err)
	}
	if err := callWithCookie(t, svc, token); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("已停用的 operator 必須 Unauthenticated，got %v", err)
	}
}

// ④ 正向：有效的 operator token 可以通過（否則前三條可能因為「什麼都拒絕」而假綠）。
func TestActiveOperatorAccepted(t *testing.T) {
	svc := newTestService(t, activeStore())
	token, err := svc.IssueToken(operatorauth.Identity{OperatorID: 1, Email: "ops@example.com", Role: "admin"})
	if err != nil {
		t.Fatalf("簽 token: %v", err)
	}
	req := connect.NewRequest(&emptypb.Empty{})
	req.Header().Set("Cookie", operatorauth.CookieName+"="+token)
	reached := false
	if _, err := svc.Interceptor().WrapUnary(func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		reached = true
		id, ok := operatorauth.IdentityFrom(ctx)
		if !ok {
			t.Error("handler 內必須取得 operator 身分")
		}
		if id.OperatorID != 1 || id.Email != "ops@example.com" || id.Role != "admin" {
			t.Errorf("身分不得由 token 偽造,got %+v", id)
		}
		return nil, nil
	})(context.Background(), req); err != nil {
		t.Fatalf("有效 token 應通過: %v", err)
	}
	if !reached {
		t.Fatal("handler 未被呼叫")
	}
}

// ⑤ sub 完整性：token 的 sub 必須等於白名單中該 email 的 operators.id。
// 「A 的 token ＋ B 的 email」若只認 email，身分就會由攜帶者自行拼裝。
func TestSubjectMismatchRejected(t *testing.T) {
	svc := newTestService(t, activeStore())
	token := signHS256(t, testSecret, jwt.MapClaims{
		"sub": 999, "email": "ops@example.com", "role": "admin", "aud": operatorauth.Audience})
	if err := callWithCookie(t, svc, token); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("sub 與白名單 id 不一致必須 Unauthenticated，got %v", err)
	}
}

// ⑥ operator token 只認 platform_session cookie：走 Authorization Bearer（租戶路徑的憑證位置）不得成立。
func TestOperatorTokenInBearerHeaderRejected(t *testing.T) {
	svc := newTestService(t, activeStore())
	token, err := svc.IssueToken(operatorauth.Identity{OperatorID: 1, Email: "ops@example.com", Role: "admin"})
	if err != nil {
		t.Fatalf("簽 token: %v", err)
	}
	req := connect.NewRequest(&emptypb.Empty{})
	req.Header().Set("Authorization", "Bearer "+token)
	_, err = svc.Interceptor().WrapUnary(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
		t.Fatal("Bearer 不得取得平台身分")
		return nil, nil
	})(context.Background(), req)
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("Bearer 路徑必須 Unauthenticated，got %v", err)
	}
}

// ⑦ 租戶 session cookie（名稱 session）不得通過平台 interceptor。
func TestTenantSessionCookieRejected(t *testing.T) {
	svc := newTestService(t, activeStore())
	token, err := svc.IssueToken(operatorauth.Identity{OperatorID: 1, Email: "ops@example.com", Role: "admin"})
	if err != nil {
		t.Fatalf("簽 token: %v", err)
	}
	req := connect.NewRequest(&emptypb.Empty{})
	req.Header().Set("Cookie", "session="+token)
	_, err = svc.Interceptor().WrapUnary(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
		t.Fatal("租戶 cookie 名稱不得取得平台身分")
		return nil, nil
	})(context.Background(), req)
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("租戶 cookie 必須 Unauthenticated，got %v", err)
	}
}

func signHS256(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	claims["exp"] = time.Now().Add(time.Hour).Unix()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("簽 token: %v", err)
	}
	return s
}

// --- OIDC 登入面（fake exchanger／verifier，無需網路） ---

type fakeExchanger struct {
	raw string
	err error
}

func (f fakeExchanger) Exchange(context.Context, string) (string, error) { return f.raw, f.err }

type fakeVerifier struct {
	ident *auth.OIDCIdentity
	err   error
}

func (f fakeVerifier) VerifyIDToken(context.Context, string) (*auth.OIDCIdentity, error) {
	return f.ident, f.err
}

func testOAuthConfig() *oauth2.Config {
	return auth.NewGoogleOAuthConfig("client-id", "client-secret", "https://api.example.com/platform/auth/google/callback")
}

// newOIDCService 建立含 OIDC 依賴的 Service；cfg 由呼叫端補（測試 cookie 旗標需要變化）。
func newOIDCService(t *testing.T, st *fakeStore, cfg operatorauth.Config) *operatorauth.Service {
	t.Helper()
	cfg.Secret = testSecret
	cfg.AllowedDomain = "example.com"
	cfg.ConsoleURL = "https://console.example.com"
	return operatorauth.New(cfg, st).WithOIDC(testOAuthConfig(), fakeExchanger{raw: "raw-id-token"},
		fakeVerifier{ident: &auth.OIDCIdentity{Email: "ops@example.com", Name: "Ops"}})
}

// findCookie 由回應取出指定名稱的 cookie。
func findCookie(t *testing.T, resp *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, c := range resp.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("回應缺少 cookie %q: %v", name, resp.Result().Header.Values("Set-Cookie"))
	return nil
}

// ⑧ Login 的 state cookie 必須 HttpOnly／Secure／SameSite=Lax／Path=/platform。
func TestLoginSetsStateCookieFlags(t *testing.T) {
	for _, secure := range []bool{true, false} {
		st := activeStore()
		svc := newOIDCService(t, st, operatorauth.Config{CookieSecure: secure, CookieDomain: ".example.com"})
		resp := httptest.NewRecorder()
		svc.Login(resp, httptest.NewRequest(http.MethodGet, "/platform/auth/google", nil))

		if resp.Code != http.StatusFound {
			t.Fatalf("Login 應導向 Google,got %d", resp.Code)
		}
		if loc := resp.Header().Get("Location"); !strings.Contains(loc, "accounts.google.com") ||
			!strings.Contains(loc, "state=") {
			t.Fatalf("Login 導向網址不含 state: %q", loc)
		}
		c := findCookie(t, resp, operatorauth.StateCookieName)
		if !c.HttpOnly || c.Secure != secure || c.SameSite != http.SameSiteLaxMode ||
			c.Path != operatorauth.CookiePath || c.Domain != ".example.com" || c.Value == "" {
			t.Fatalf("state cookie 旗標不符(HttpOnly/Secure=%v/SameSite=Lax/Path/Domain): %+v", secure, c)
		}
	}
}

// ⑨ Callback 成功：設 platform_session（HttpOnly／Secure／SameSite=Lax／Path=/platform）、
// 清掉 state cookie、留下登入時間與稽核、導回 ConsoleURL。
func TestCallbackSuccess(t *testing.T) {
	st := activeStore()
	svc := newOIDCService(t, st, operatorauth.Config{CookieSecure: true, CookieDomain: ".example.com"})

	req := httptest.NewRequest(http.MethodGet, "/platform/auth/google/callback?state=st-1&code=code-1", nil)
	req.AddCookie(&http.Cookie{Name: operatorauth.StateCookieName, Value: "st-1"})
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 10.0.0.1")
	req.Header.Set("User-Agent", "console-test")
	resp := httptest.NewRecorder()
	svc.Callback(resp, req)

	if resp.Code != http.StatusFound || resp.Header().Get("Location") != "https://console.example.com" {
		t.Fatalf("Callback 應導回 console,got %d %q", resp.Code, resp.Header().Get("Location"))
	}
	c := findCookie(t, resp, operatorauth.CookieName)
	if !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteLaxMode ||
		c.Path != operatorauth.CookiePath || c.Domain != ".example.com" || c.Value == "" {
		t.Fatalf("session cookie 旗標不符: %+v", c)
	}
	// 簽出的 token 必須能通過 interceptor（cookie 值與驗證路徑同一份契約）。
	if err := callWithCookie(t, svc, c.Value); err != nil {
		t.Fatalf("Callback 簽出的 token 應可通過 interceptor: %v", err)
	}
	// 一次性 state：驗完即清。
	if cleared := findCookie(t, resp, operatorauth.StateCookieName); cleared.Value != "" || cleared.MaxAge >= 0 {
		t.Fatalf("state cookie 必須驗完即清,got %+v", cleared)
	}
	if st.touched != 1 || st.audited != "ops@example.com" || st.auditIP != "203.0.113.7" || st.auditUA != "console-test" {
		t.Fatalf("登入副作用不完整: touched=%d email=%q ip=%q ua=%q",
			st.touched, st.audited, st.auditIP, st.auditUA)
	}
}

// ⑩ Callback 每一步失敗都不得簽發 session cookie，且 state 不符必須擋在最前面。
func TestCallbackRejections(t *testing.T) {
	cases := []struct {
		name     string
		op       *operatorauth.Operator
		ident    *auth.OIDCIdentity
		state    string // 請求帶的 state；"" = 不帶 state cookie
		query    string
		want     int
		auditErr error
	}{
		{name: "缺 state cookie", op: activeStore().op, ident: &auth.OIDCIdentity{Email: "ops@example.com"},
			query: "?state=st-1&code=c", want: http.StatusBadRequest},
		{name: "state 不符", op: activeStore().op, ident: &auth.OIDCIdentity{Email: "ops@example.com"},
			state: "other", query: "?state=st-1&code=c", want: http.StatusBadRequest},
		{name: "網域不符", op: activeStore().op, ident: &auth.OIDCIdentity{Email: "ops@evil.com"},
			state: "st-1", query: "?state=st-1&code=c", want: http.StatusForbidden},
		{name: "非白名單", op: nil, ident: &auth.OIDCIdentity{Email: "someone@example.com"},
			state: "st-1", query: "?state=st-1&code=c", want: http.StatusForbidden},
		{name: "已停用", op: &operatorauth.Operator{ID: 1, Email: "ops@example.com", Role: "admin", Status: "disabled"},
			ident: &auth.OIDCIdentity{Email: "ops@example.com"}, state: "st-1", query: "?state=st-1&code=c",
			want: http.StatusForbidden},
		{name: "稽核寫入失敗則不發 cookie", op: activeStore().op, ident: &auth.OIDCIdentity{Email: "ops@example.com"},
			state: "st-1", query: "?state=st-1&code=c", want: http.StatusInternalServerError,
			auditErr: errors.New("db down")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{op: tc.op, failAudit: tc.auditErr}
			svc := newOIDCService(t, st, operatorauth.Config{})
			req := httptest.NewRequest(http.MethodGet, "/platform/auth/google/callback"+tc.query, nil)
			if tc.state != "" {
				req.AddCookie(&http.Cookie{Name: operatorauth.StateCookieName, Value: tc.state})
			}
			resp := httptest.NewRecorder()
			svc.Callback(resp, req)

			if resp.Code != tc.want {
				t.Fatalf("狀態碼 want %d got %d", tc.want, resp.Code)
			}
			for _, c := range resp.Result().Cookies() {
				if c.Name == operatorauth.CookieName && c.Value != "" {
					t.Fatalf("失敗的登入不得簽發 session cookie: %+v", c)
				}
			}
		})
	}
}
