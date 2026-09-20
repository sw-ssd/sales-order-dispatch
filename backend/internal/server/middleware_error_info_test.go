package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexedwards/scs/v2"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	authzopenfga "github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	commonv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
	"google.golang.org/protobuf/proto"
)

// mwErrBody 為 middleware 錯誤回應的 wire 形狀（與 connect 錯誤協定一致）。
type mwErrBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details []struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"details"`
}

// errorInfoOf 由 wire body 取 ErrorInfo（客戶端視角：base64 → proto）。
func (b mwErrBody) errorInfoOf(t *testing.T) *commonv1.ErrorInfo {
	t.Helper()
	for _, d := range b.Details {
		raw, err := base64.RawStdEncoding.DecodeString(d.Value)
		if err != nil {
			t.Fatalf("detail value 應為 base64: %v", err)
		}
		msg := &commonv1.ErrorInfo{}
		if err := proto.Unmarshal(raw, msg); err != nil {
			t.Fatalf("detail value 應為 ErrorInfo: %v", err)
		}
		return msg
	}
	t.Fatal("middleware 錯誤回應未帶 details(ErrorInfo)")
	return nil
}

// TestMiddlewareErrorBodyCarriesErrorInfo：middleware 閘門（不經 connect handler）的錯誤
// 也必須讓**客戶端**讀到註冊碼與 trace_id —— 這些請求不會經過 requestid.Interceptor，
// 故 trace_id 由 HTTP 邊界補（requestid.Ensure/Stamp）。以真 httptest server + 真 HTTP
// client 驗證（跨網路），而不是只看 server 端組出來的東西。
func TestMiddlewareErrorBodyCarriesErrorInfo(t *testing.T) {
	const protectedPath = "/salesorder.v1.RoleService/ListRoles" // role/read，見 protectedRPC

	mwServer := func(t *testing.T, s *Server, db *ent.Client, sessions *scs.SessionManager) *httptest.Server {
		t.Helper()
		probe := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
		ts := httptest.NewServer(sessions.LoadAndSave(s.authzMiddleware(db, sessions, probe)))
		t.Cleanup(ts.Close)
		return ts
	}

	t.Run("未登入 → AUTH-4001 + trace_id", func(t *testing.T) {
		s, sessions := newIdentityTestEnv()
		db := openIdentityDB(t, "file:mw-errcode-a?mode=memory&cache=shared&_fk=1")
		ts := mwServer(t, s, db, sessions)

		resp, err := http.Get(ts.URL + protectedPath)
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		body := decodeMWErrBody(t, resp)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("HTTP 狀態 = %d, want 401", resp.StatusCode)
		}
		if body.Code != "unauthenticated" || body.Message == "" {
			t.Fatalf("既有欄位不變:code=%q message=%q", body.Code, body.Message)
		}
		info := body.errorInfoOf(t)
		if info.GetCode() != "AUTH-4001" {
			t.Fatalf("ErrorInfo.code = %q, want AUTH-4001", info.GetCode())
		}
		if info.GetTraceId() == "" {
			t.Fatal("ErrorInfo.trace_id 不得為空（客戶端要能對上 server log）")
		}
	})

	t.Run("首登受限 → AUTH-3004 + trace_id", func(t *testing.T) {
		s, sessions := newIdentityTestEnv()
		ctx := context.Background()
		db := openIdentityDB(t, "file:mw-errcode-c?mode=memory&cache=shared&_fk=1")
		co := db.Company.Create().SetName("測試公司").SetIdentifier("T-mw-mcp").SaveX(ctx)
		u := db.User.Create().SetEmail("mcp@example.com").SetName("首登").SetStatus(user.StatusActive).
			SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SetMustChangePassword(true).SaveX(ctx)
		ts := mwServer(t, s, db, sessions)

		req, err := http.NewRequest(http.MethodGet, ts.URL+protectedPath, nil)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		req.AddCookie(testSessionCookie(t, sessions, int(u.ID), "staff"))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		body := decodeMWErrBody(t, resp)
		// 既有欄位不變：connect 碼仍為 failed_precondition（HTTP 5xx 是既有對映，
		// failed_precondition 不在 httpStatusForCode 的表內；修正屬另一件事，見報告 deferred）。
		if body.Code != "failed_precondition" {
			t.Fatalf("connect 碼 = %q, want failed_precondition", body.Code)
		}
		if body.Message != "首次登入須先修改密碼" {
			t.Fatalf("message = %q，對外文字不得改變", body.Message)
		}
		info := body.errorInfoOf(t)
		if info.GetCode() != "AUTH-3004" {
			t.Fatalf("ErrorInfo.code = %q, want AUTH-3004", info.GetCode())
		}
		if info.GetTraceId() == "" {
			t.Fatal("ErrorInfo.trace_id 不得為空")
		}
	})

	t.Run("OpenFGA 無權 → SYS-4001 + details 帶 resource/action", func(t *testing.T) {
		s, sessions := newIdentityTestEnv()
		ctx := context.Background()
		db := openIdentityDB(t, "file:mw-errcode-b?mode=memory&cache=shared&_fk=1")
		co := db.Company.Create().SetName("測試公司").SetIdentifier("T-mw-err").SaveX(ctx)
		// 刻意不寫任何 tuple：staff 對 ability:role 無 can_read。
		staff := db.User.Create().SetEmail("noperm@example.com").SetName("無權者").SetStatus(user.StatusActive).
			SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).SaveX(ctx)
		fgaClient, err := openfga.NewMemory(ctx, "mw-errcode-store")
		if err != nil {
			t.Fatalf("NewMemory: %v", err)
		}
		t.Cleanup(fgaClient.Close)
		s.SetOpenFGA(authzopenfga.New(fgaClient))
		ts := mwServer(t, s, db, sessions)

		req, err := http.NewRequest(http.MethodGet, ts.URL+protectedPath, nil)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		req.AddCookie(testSessionCookie(t, sessions, int(staff.ID), "staff"))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		body := decodeMWErrBody(t, resp)
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("HTTP 狀態 = %d, want 403", resp.StatusCode)
		}
		if body.Code != "permission_denied" {
			t.Fatalf("connect 碼 = %q, want permission_denied", body.Code)
		}
		info := body.errorInfoOf(t)
		if info.GetCode() != "SYS-4001" {
			t.Fatalf("ErrorInfo.code = %q, want SYS-4001", info.GetCode())
		}
		if info.GetTraceId() == "" {
			t.Fatal("ErrorInfo.trace_id 不得為空")
		}
		if info.GetDetails()["resource"] != "role" || info.GetDetails()["action"] != "read" {
			t.Fatalf("details 應帶 resource=role／action=read,got %v", info.GetDetails())
		}
	})
}

// decodeMWErrBody 讀出 middleware 錯誤回應的 JSON body。
func decodeMWErrBody(t *testing.T, resp *http.Response) mwErrBody {
	t.Helper()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("讀 body: %v", err)
	}
	var body mwErrBody
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("body 應為合法 JSON: %v (body=%q)", err, raw)
	}
	return body
}
