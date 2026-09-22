package resterr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/errcode"
)

// TestStatusMapping 釘住「全站唯一一份」的 connect code → HTTP 狀態表。
// 這條分岐曾實際出錯:server 自己的舊表漏了 NotFound,同一個 NotFound 走 REST 是 404、
// 走 middleware 是 500。改動 status 時這個測試必須跟著被檢視,而不是被順手改綠。
func TestStatusMapping(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{errcode.SysNotFound.Error(nil), http.StatusNotFound},
		{errcode.SysInvalidArgument.Error(nil), http.StatusBadRequest},
		{errcode.SysPermissionDenied.Error(nil), http.StatusForbidden},
		{errcode.AuthUnauthenticated.Error(nil), http.StatusUnauthorized},
		{errcode.SysInternal.Error(nil), http.StatusInternalServerError},
		{connect.NewError(connect.CodeFailedPrecondition, nil), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		if got := status(tc.err); got != tc.want {
			t.Errorf("status(%v) = %d, want %d", tc.err, got, tc.want)
		}
	}
}

// TestWriteShape 驗證對外 JSON 形狀與 trace_id 的落點:middleware 閘門與 REST 端點共用此函式,
// 形狀一改就是兩邊一起改(客戶端只讀 code/message,details 帶 ErrorInfo 含 trace_id)。
func TestWriteShape(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	Write(rec, req, errcode.SysPermissionDenied.Error(map[string]string{"field": "logo"}))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("HTTP 狀態 = %d, want 403", rec.Code)
	}
	var body struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details []struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"details"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("解譯回應: %v,body=%s", err, rec.Body.String())
	}
	if body.Code != connect.CodePermissionDenied.String() {
		t.Fatalf("code = %q, want permission_denied", body.Code)
	}
	if body.Message == "" {
		t.Fatal("message 不得為空(前端直接顯示後端渲染的繁中訊息)")
	}
	if len(body.Details) == 0 {
		t.Fatal("details 應帶 ErrorInfo(含 trace_id)")
	}
}
