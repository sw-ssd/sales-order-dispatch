package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	platformv1 "github.com/salesorder/sales-order-1.0/backend/contracts/proto/platform/v1"
	platformv1connect "github.com/salesorder/sales-order-1.0/backend/contracts/proto/platform/v1/platformv1connect"
	"github.com/salesorder/platform/entitlements"
	"github.com/salesorder/platform/server"
)

// TestPlatformServerSmoke 驗收 phase-2 獨立服務:handler 穿透到 entitlements.Service。
// Unlimited 模式不查 store,CheckLimit 直接通過;重點是 server.Register 掛載與 Connect 路徑通。
func TestPlatformServerSmoke(t *testing.T) {
	mux := http.NewServeMux()
	server.Register(mux, entitlements.Unlimited())

	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := platformv1connect.NewTenantEntitlementServiceClient(http.DefaultClient, srv.URL)
	req := connect.NewRequest(&platformv1.CheckLimitRequest{
		ProductId:         "sales-order",
		CompanyInternalId: uuid.New().String(),
		Feature:           entitlements.LimitSeats,
		Delta:             1,
	})
	if _, err := client.CheckLimit(context.Background(), req); err != nil {
		t.Fatalf("Unlimited 模式下 CheckLimit 應通過,卻: %v", err)
	}
}
