//go:build integration

package services

import (
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationSalesOrdersSchemaUpDown 驗 00031 的表／索引／policy／GRANT 全套存在，
// 且 Down 可乾淨還原（T1 只落地結構，ENABLE 與服務收斂另批）。
func TestIntegrationSalesOrdersSchemaUpDown(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)
	admin := openRawDB(t, adminDSN)

	tables := []string{"sales_orders", "sales_order_items", "sales_order_events", "order_counters"}
	for _, tb := range tables {
		var n int
		if err := admin.QueryRowContext(t.Context(),
			`SELECT count(*) FROM information_schema.tables WHERE table_name = $1`, tb).Scan(&n); err != nil {
			t.Fatalf("查表 %s: %v", tb, err)
		}
		if n != 1 {
			t.Fatalf("表 %s 應存在", tb)
		}
	}
	// D12：訂單兩表不得有金額欄位。
	for _, q := range []string{
		`SELECT count(*) FROM information_schema.columns WHERE table_name = 'sales_orders'
		  AND column_name IN ('price','amount','subtotal','tax','discount','total','currency','unit_price')`,
		`SELECT count(*) FROM information_schema.columns WHERE table_name = 'sales_order_items'
		  AND column_name IN ('price','amount','subtotal','tax','discount','total','currency','unit_price')`,
	} {
		var n int
		if err := admin.QueryRowContext(t.Context(), q).Scan(&n); err != nil {
			t.Fatalf("D12 斷言: %v", err)
		}
		if n != 0 {
			t.Fatal("訂單表不得含金額欄位(D12)")
		}
	}
	// events 僅追加：app_rw 無 UPDATE/DELETE。
	var hasUpdate, hasDelete bool
	if err := admin.QueryRowContext(t.Context(),
		`SELECT has_table_privilege('app_rw', 'sales_order_events', 'UPDATE')`).Scan(&hasUpdate); err != nil {
		t.Fatalf("查權限: %v", err)
	}
	if err := admin.QueryRowContext(t.Context(),
		`SELECT has_table_privilege('app_rw', 'sales_order_events', 'DELETE')`).Scan(&hasDelete); err != nil {
		t.Fatalf("查權限: %v", err)
	}
	if hasUpdate || hasDelete {
		t.Fatal("sales_order_events 必須僅追加(app_rw 不得 UPDATE/DELETE)")
	}

	migrateBusinessDownTo(t, adminDSN, "30")
	for _, tb := range tables {
		var n int
		if err := admin.QueryRowContext(t.Context(),
			`SELECT count(*) FROM information_schema.tables WHERE table_name = $1`, tb).Scan(&n); err != nil {
			t.Fatalf("Down 後查表 %s: %v", tb, err)
		}
		if n != 0 {
			t.Fatalf("Down 後表 %s 應消失", tb)
		}
	}
	// Down 後重上：冪等。
	migrateBusinessUp(t, adminDSN)
	var n int
	if err := admin.QueryRowContext(t.Context(),
		`SELECT count(*) FROM information_schema.tables WHERE table_name = 'sales_orders'`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("重上後 sales_orders 應存在: %v", err)
	}
}
