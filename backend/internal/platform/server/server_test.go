package server_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"errors"
	"github.com/google/uuid"


	"github.com/salesorder/sales-order-1.0/backend/contracts/errcode"
	platformv1 "github.com/salesorder/sales-order-1.0/backend/contracts/proto/platform/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/entitlements"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/server"
	"github.com/salesorder/sales-order-1.0/backend/internal/platform/store"
)

func ptr[T any](v T) *T { return &v }

func TestCheckLimitHandler(t *testing.T) {
	f := store.NewFake()
	internalID := uuid.MustParse("00000000-0000-0000-0000-000000000042")
	f.PutSubscription(store.Subscription{
		CompanyID: 42, InternalID: internalID, PlanCode: "std", Status: "active",
	})
	f.PutFeature(store.Feature{Code: entitlements.LimitSeats, Type: "integer", Unit: "席"})
	f.PutPlan("std", []store.Entitlement{{FeatureCode: entitlements.LimitSeats, Enabled: true, Limit: ptr(int64(2))}})
	svc := entitlements.New(f, nil, nil, 0)

	h := server.NewTenantEntitlementService(svc)

	// 超額:cur=1 + delta=2 > limit=2 → PLAT-5001。
	req := connect.NewRequest(&platformv1.CheckLimitRequest{
		ProductId:         "sales-order",
		CompanyInternalId: internalID.String(),
		Feature:           entitlements.LimitSeats,
		Delta:             2,
		CurrentUsed:       1, // 超額:1 + 2 > limit 2 → PLAT-5001
	})
	if _, err := h.CheckLimit(context.Background(), req); err == nil {
		t.Fatal("超額應回 PLAT-5001")
	} else {
		var ce *connect.Error
		if !errors.As(err, &ce) || ce.Code() != errcode.PlatformLimitExceeded.ConnectCode() {
			t.Fatalf("應為 PLAT-5001,卻: %v", err)
		}
	}

	// 未超額:1 + 1 = 2 不超。
	req.Msg.Delta = 1
	if _, err := h.CheckLimit(context.Background(), req); err != nil {
		t.Fatalf("未超額應通過,卻: %v", err)
	}
}

