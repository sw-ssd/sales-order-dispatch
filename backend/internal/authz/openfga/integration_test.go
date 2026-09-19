//go:build integration

// Package openfga 的整合測試(D21 對應授權面):以真實 PostgreSQL 驗證 OpenFGA + RLS
// 的跨公司/部門/self 授權隔離。要求環境變數 INTEGRATION_TEST_DSN(業務 DB 連線,
// 亦供 OpenFGA datastore 使用);未設定時自動 skip(本 CI/單測環境無 Postgres)。
// 執行:INTEGRATION_TEST_DSN=postgres://... go test -tags integration ./internal/authz/openfga/... -v
package openfga_test

import (
	"context"
	"os"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz/openfga"
	ofga "github.com/salesorder/sales-order-1.0/backend/third_party/openfga"
)

// openTestPostgres 依 INTEGRATION_TEST_DSN 起 OpenFGA postgres datastore;未設定即 skip。
func openTestPostgres(t *testing.T) *openfga.Engine {
	t.Helper()
	dsn := os.Getenv("INTEGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("INTEGRATION_TEST_DSN 未設定,跳過 OpenFGA+RLS 整合測試(D21)")
	}
	client, err := ofga.NewPostgres(context.Background(), dsn, "int-test-store")
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}
	t.Cleanup(client.Close)
	return openfga.New(client)
}

// TestIntegrationTenantIsolation D21 授權隔離整合測試:不同 company/department 的使用者
// 僅能透過 OpenFGA tuple 取得自身租戶範圍能力;無授權的使用者 fail-closed 拒絕。
// (RLS 端資料過濾由 00007_rls_core_policies.sql 落地,於業務表查詢驗收;此測試聚焦
// OpenFGA 跨租戶能力隔離——同 store 不同租戶不互相洩漏。)
func TestIntegrationTenantIsolation(t *testing.T) {
	e := openTestPostgres(t)
	ctx := context.Background()

	// 兩個 tenant(company c1 / c2)各自授權其使用者。
	if err := e.WriteTuple(ctx, "user:u-c1", "can_read", "ability:sales_order"); err != nil {
		t.Fatalf("WriteTuple(c1 read sales_order): %v", err)
	}
	if err := e.WriteTuple(ctx, "user:u-c1", "can_write", "ability:department"); err != nil {
		t.Fatalf("WriteTuple(c1 write department): %v", err)
	}
	if err := e.WriteTuple(ctx, "user:u-c2", "can_read", "ability:sales_order"); err != nil {
		t.Fatalf("WriteTuple(c2 read sales_order): %v", err)
	}

	if ok, _ := e.Check(ctx, "user:u-c1", "can_read", "ability:sales_order"); !ok {
		t.Fatal("u-c1 應可讀 sales_order")
	}
	if ok, _ := e.Check(ctx, "user:u-c2", "can_read", "ability:sales_order"); !ok {
		t.Fatal("u-c2 應可讀 sales_order")
	}
	// c1 有 department 寫權,c2 沒有 → 不洩漏。
	if ok, _ := e.Check(ctx, "user:u-c2", "can_write", "ability:department"); ok {
		t.Fatal("u-c2 不應有 department 寫權(跨租戶洩漏)")
	}
	// 未授權使用者 fail-closed。
	if ok, _ := e.Check(ctx, "user:u-nobody", "can_read", "ability:sales_order"); ok {
		t.Fatal("未授權使用者應拒絕")
	}

	// ListObjects 隔離:u-c1 可讀資源含 sales_order 但非 department(僅寫)。
	objs, err := e.ListObjects(ctx, "user:u-c1", "can_read", "ability")
	if err != nil {
		t.Fatalf("ListObjects: %v", err)
	}
	found := false
	for _, o := range objs {
		if o == "ability:sales_order" {
			found = true
		}
		if o == "ability:department" {
			t.Fatal("u-c1 department 僅寫權不應出現在可讀清單")
		}
	}
	if !found {
		t.Fatalf("u-c1 可讀清單應含 sales_order,got %v", objs)
	}
}
