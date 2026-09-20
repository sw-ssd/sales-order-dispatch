//go:build integration

package services

import (
	"errors"
	"testing"

	"connectrpc.com/connect"

	v1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/salesorder/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationErrorInfoReachesClient：以 connect client 發出必然失敗的請求，斷言
// ClientError 內含 ErrorInfo{code, message, trace_id 非空} —— 這條驗的是「碼跨網路到得了
// 客戶端」（P0-5 的驗收），而不只是「碼在 server 端組出來了」；trace_id 由 requestid
// interceptor 在回應邊界補（T3b），這裡一併釘住它真的出現在客戶端收到的那一份 detail 上。
func TestIntegrationErrorInfoReachesClient(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)
	_, db := openPGEntClientFromGoose(t, adminDSN)
	ctx := t.Context()

	clients := newListScanServer(t, db, 1) // 既有樣板：super 身分、公司範圍 1
	_, err := clients.companies.GetCompany(ctx, connect.NewRequest(
		&v1.GetCompanyRequest{CompanyId: "999999"}))
	if err == nil {
		t.Fatal("查不存在的公司應失敗")
	}
	var ce *connect.Error
	if !errors.As(err, &ce) {
		t.Fatalf("應為 *connect.Error,got %T", err)
	}
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("connect 碼應為 not_found,got %v", connect.CodeOf(err))
	}

	info := errorInfoOf(t, err)
	if info.GetCode() != "SYS-4002" {
		t.Fatalf("錯誤碼應為 SYS-4002,got %q", info.GetCode())
	}
	if info.GetMessage() == "" {
		t.Fatal("ErrorInfo 應帶 message（前端直接顯示，不自行組字串）")
	}
	if info.GetTraceId() == "" {
		t.Fatal("ErrorInfo 應帶 trace_id（客服回報時對 server log 用）")
	}
}
