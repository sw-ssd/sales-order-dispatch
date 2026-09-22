package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
)

// meBody 為 /me 成功回應的形狀(測試只取用到的欄位)。
type meBody struct {
	UserID  string `json:"user_id"`
	Role    string `json:"role"`
	Company *struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		LogoURL string `json:"logo_url"`
	} `json:"company"`
}

// errBody 為 resterr 錯誤協定的頂層欄位(details 內含 base64 的 ErrorInfo)。
type errBody struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Details []map[string]string `json:"details"`
}

// meCall 以（可為零值的）身分呼叫 GET /me;id 為零值時不注入身分＝未登入路徑。
func meCall(t *testing.T, h *AuthHandler, id authz.Identity) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if id.UserID != "" {
			ctx = authz.WithIdentity(ctx, id)
		}
		h.Me(w, r.WithContext(ctx))
	})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/me", nil))
	return rec
}

// decodeMe 解譯 200 回應。
func decodeMe(t *testing.T, rec *httptest.ResponseRecorder) meBody {
	t.Helper()
	var b meBody
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("解譯 /me 回應: %v,body=%s", err, rec.Body.String())
	}
	return b
}

// seedCompany 建一間公司（identifier 逐測試唯一;logo 可指定）。
func seedCompany(t *testing.T, db *ent.Client, ident, logoURL string, deletedAt *time.Time) int {
	t.Helper()
	b := db.Company.Create().SetName("公司-" + ident).SetIdentifier(ident)
	if logoURL != "" {
		b = b.SetLogoURL(logoURL)
	}
	if deletedAt != nil {
		b = b.SetDeletedAt(*deletedAt)
	}
	c := b.SaveX(t.Context())
	return c.ID
}

// TestMeUnauthenticated 未注入身分 → 401,且錯誤協定可追（ErrorInfo 帶 AUTH-4001 與 trace_id）。
func TestMeUnauthenticated(t *testing.T) {
	db := enttest.Open(t, "sqlite3", "file:me-unauth?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	h := NewAuthHandler(AuthDeps{DB: db})

	rec := meCall(t, h, authz.Identity{})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("未登入應 401,got %d body=%s", rec.Code, rec.Body.String())
	}
	var b errBody
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("解譯錯誤回應: %v", err)
	}
	if b.Code != "unauthenticated" || b.Message == "" {
		t.Fatalf("錯誤協定頂層應為 unauthenticated + 非空訊息,got code=%q message=%q", b.Code, b.Message)
	}
	if len(b.Details) == 0 {
		t.Fatal("錯誤應帶 details(內含 ErrorInfo)")
	}
	raw, err := base64.RawStdEncoding.DecodeString(b.Details[0]["value"])
	if err != nil {
		t.Fatalf("解碼 detail: %v", err)
	}
	var info commonv1.ErrorInfo
	if err := proto.Unmarshal(raw, &info); err != nil {
		t.Fatalf("還原 ErrorInfo: %v", err)
	}
	if info.GetCode() != "AUTH-4001" {
		t.Fatalf("ErrorInfo.code 應 AUTH-4001,got %q", info.GetCode())
	}
	if info.GetTraceId() == "" {
		t.Fatal("ErrorInfo.trace_id 應非空(REST 寫出邊界 Stamp)")
	}
}

// TestMeReturnsOwnCompany 身分帶 company_id → 200 回身分與該公司品牌（公司由身分定錨）。
func TestMeReturnsOwnCompany(t *testing.T) {
	db := enttest.Open(t, "sqlite3", "file:me-co?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	coID := seedCompany(t, db, "ME-CO", "/api/v1/files/logo-x/download", nil)
	h := NewAuthHandler(AuthDeps{DB: db})

	rec := meCall(t, h, authz.Identity{
		UserID: "9", CompanyID: strconv.Itoa(coID), Role: "super", Roles: []string{"super"},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("應 200,got %d body=%s", rec.Code, rec.Body.String())
	}
	b := decodeMe(t, rec)
	if b.UserID != "9" || b.Role != "super" {
		t.Fatalf("身分欄位不符,got user_id=%q role=%q", b.UserID, b.Role)
	}
	if b.Company == nil {
		t.Fatal("所屬公司應非 null")
	}
	if b.Company.ID != strconv.Itoa(coID) || b.Company.Name != "公司-ME-CO" ||
		b.Company.LogoURL != "/api/v1/files/logo-x/download" {
		t.Fatalf("公司欄位不符,got %+v", b.Company)
	}
}

// TestMeCompanyNullShapes 無公司身分、公司不存在(company_id 損壞)與已軟刪公司:一律 company=null
// （查無與已軟刪不可區分——前端一律降級為預設圖示,不洩漏公司存在與否）。
func TestMeCompanyNullShapes(t *testing.T) {
	db := enttest.Open(t, "sqlite3", "file:me-null?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = db.Close() })
	deleted := time.Now()
	delID := seedCompany(t, db, "ME-DEL", "", &deleted)
	h := NewAuthHandler(AuthDeps{DB: db})

	cases := []struct {
		name, companyID string
	}{
		{"無公司(未註冊/員工未歸屬)", ""},
		{"company_id 損壞", "not-a-number"},
		{"已軟刪公司", strconv.Itoa(delID)},
		{"不存在的公司", "999999"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := meCall(t, h, authz.Identity{
				UserID: "3", CompanyID: tc.companyID, Role: "staff", Roles: []string{"staff"},
			})
			if rec.Code != http.StatusOK {
				t.Fatalf("身分存在即應 200,got %d body=%s", rec.Code, rec.Body.String())
			}
			b := decodeMe(t, rec)
			if b.UserID != "3" {
				t.Fatalf("身分仍應回傳,got %q", b.UserID)
			}
			if b.Company != nil {
				t.Fatalf("company 應 null,got %+v", b.Company)
			}
		})
	}
}
