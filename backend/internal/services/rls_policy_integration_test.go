//go:build integration

package services

import (
	"database/sql"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// rlsPolicyTables 為本波要求「有 FOR ALL policy 且 WITH CHECK 非空」的租戶表。
var rlsPolicyTables = []string{
	"companies", "departments", "users", "roles", "role_permissions",
	"audit_logs", "metadicts", "customers", "customer_addresses", "customer_contacts",
	"customer_counters", "warehouses", "routes", "processing_specs",
	"product_categories", "products", "product_units", "product_processing_specs",
}

// TestIntegrationRLSPolicyCoverage 檢查 policy 覆蓋率與 WITH CHECK：
// 少了 WITH CHECK，RLS 只能擋讀、擋不住「把列寫成別的 company_id」。
func TestIntegrationRLSPolicyCoverage(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	defer db.Close()

	for _, table := range rlsPolicyTables {
		t.Run(table, func(t *testing.T) {
			var total, withCheck int
			if err := db.QueryRow(
				`SELECT count(*), count(with_check) FROM pg_policies
				  WHERE schemaname = 'public' AND tablename = $1 AND cmd = 'ALL'`,
				table,
			).Scan(&total, &withCheck); err != nil {
				t.Fatalf("查 pg_policies(%s): %v", table, err)
			}
			if total == 0 {
				t.Fatalf("%s 沒有 FOR ALL policy", table)
			}
			if withCheck != total {
				t.Fatalf("%s 有 %d/%d 個 policy 缺 WITH CHECK（可跨租戶寫入）", table, total-withCheck, total)
			}
		})
	}

	// metadicts 的 WITH CHECK 必須維持原樣（不含 department_id IS NULL）：放寬會讓
	// 一般租戶寫入系統預設字典。
	var metaCheck string
	if err := db.QueryRow(
		`SELECT with_check FROM pg_policies
		  WHERE tablename = 'metadicts' AND policyname = 'core_metadicts_scope'`,
	).Scan(&metaCheck); err != nil {
		t.Fatalf("查 metadicts policy: %v", err)
	}
	if containsAll(metaCheck, "department_id IS NULL") {
		t.Fatalf("core_metadicts_scope 的 WITH CHECK 被放寬（含 department_id IS NULL）：%s", metaCheck)
	}
}

func containsAll(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) &&
		(haystack == needle || indexOf(haystack, needle) >= 0)
}

func indexOf(h, n string) int {
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return i
		}
	}
	return -1
}
