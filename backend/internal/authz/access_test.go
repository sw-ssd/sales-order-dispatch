package authz_test

import (
	"context"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
)

func TestIdentityContext(t *testing.T) {
	ctx := context.Background()
	// 未注入 → 零值。
	if id := authz.IdentityFrom(ctx); id.UserID != "" {
		t.Fatalf("未注入身分應為零值,got %#v", id)
	}
	id := authz.Identity{UserID: "u1", CompanyID: "c1", Roles: []string{"staff"}}
	got := authz.IdentityFrom(authz.WithIdentity(ctx, id))
	if got.UserID != "u1" || got.CompanyID != "c1" {
		t.Fatalf("身分注入後讀回不符,got %#v", got)
	}
}

func TestRegistryConditionFields(t *testing.T) {
	ctx := context.Background()
	reg := authz.Registry(ctx)
	fields := reg.ConditionFields("sales_order")
	if len(fields) == 0 {
		t.Fatal("sales_order 應註冊條件欄位白名單(未註冊→空)")
	}
	// 未註冊資源 → 空(不 panic)。
	if got := reg.ConditionFields("no_such_resource"); len(got) != 0 {
		t.Fatalf("未註冊資源條件欄位應為空,got %#v", got)
	}
}
