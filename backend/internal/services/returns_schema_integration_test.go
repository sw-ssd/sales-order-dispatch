//go:build integration

package services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationReturnRequestsSchemaUpDown 驗證 00039:兩表存在 + 索引 + app_rw CRUD 全授 +
// Down 乾淨 + 重上冪等(比照 print_logs 測試;RLS ENABLE 留待服務收斂批)。
func TestIntegrationReturnRequestsSchemaUpDown(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateReturnsUp(t, dsn)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	defer func() { _ = db.Close() }()

	for _, tbl := range []string{"return_requests", "return_request_items"} {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM ` + tbl).Scan(&n); err != nil {
			t.Fatalf("%s 不存在: %v", tbl, err)
		}
	}
	// 索引存在(客戶自查 + 部門待審 + 明細外鍵)。
	for _, idx := range []string{"return_requests_customer_status_idx",
		"return_requests_dept_status_idx", "return_request_items_request_idx"} {
		var exists bool
		if err := db.QueryRow(`SELECT to_regclass($1) IS NOT NULL`, "public."+idx).Scan(&exists); err != nil {
			t.Fatalf("查索引 %s: %v", idx, err)
		}
		if !exists {
			t.Fatalf("索引 %s 應存在", idx)
		}
	}
	// 業務實體:app_rw CRUD 全授。
	for _, tbl := range []string{"return_requests", "return_request_items"} {
		var sel, ins, upd, del bool
		q := `SELECT has_table_privilege('app_rw', $1, 'SELECT'),
			has_table_privilege('app_rw', $1, 'INSERT'),
			has_table_privilege('app_rw', $1, 'UPDATE'),
			has_table_privilege('app_rw', $1, 'DELETE')`
		if err := db.QueryRow(q, tbl).Scan(&sel, &ins, &upd, &del); err != nil {
			t.Fatalf("查授權 %s: %v", tbl, err)
		}
		if !sel || !ins || !upd || !del {
			t.Fatalf("%s 應 CRUD 全授,got %v %v %v %v", tbl, sel, ins, upd, del)
		}
	}

	// down-to 38 只回滾 00039:兩表消失;重上冪等。
	ctx := context.Background()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("dialect: %v", err)
	}
	goose.SetTableName("goose_db_version")
	if err := goose.RunContext(ctx, "down-to", db, "../../database/migrations", "38"); err != nil {
		t.Fatalf("down-to 38: %v", err)
	}
	for _, tbl := range []string{"return_requests", "return_request_items"} {
		var exists bool
		if err := db.QueryRow(`SELECT to_regclass($1) IS NOT NULL`, "public."+tbl).Scan(&exists); err != nil {
			t.Fatalf("查表存在: %v", err)
		}
		if exists {
			t.Fatalf("Down 後 %s 應消失", tbl)
		}
	}
	if err := goose.RunContext(ctx, "up", db, "../../database/migrations"); err != nil {
		t.Fatalf("重上: %v", err)
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM return_requests`).Scan(&n); err != nil {
		t.Fatalf("重上後 return_requests 應存在: %v", err)
	}
}

// migrateReturnsUp 以 goose 全量 up。
func migrateReturnsUp(t *testing.T, dsn string) {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("dialect: %v", err)
	}
	goose.SetTableName("goose_db_version")
	goose.SetBaseFS(nil)
	if err := goose.RunContext(t.Context(), "up", db, "../../database/migrations"); err != nil {
		t.Fatalf("goose up: %v", err)
	}
}
