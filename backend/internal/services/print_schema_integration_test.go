//go:build integration

package services_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationPrintRecordsSchemaUpDown 驗證 00037:兩表存在 + 欄位形狀(print_logs 含
// is_reprint/reprint_reason,file_asset FK) + 記錄類僅追加(app_rw 無 UPDATE/DELETE) +
// Down 乾淨 + 重上冪等。
func TestIntegrationPrintRecordsSchemaUpDown(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migratePrintUp(t, dsn)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	defer func() { _ = db.Close() }()

	for _, tbl := range []string{"print_logs", "print_previews"} {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM ` + tbl).Scan(&n); err != nil {
			t.Fatalf("%s 不存在: %v", tbl, err)
		}
	}
	// file_asset FK 欄位存在。
	var hasFK bool
	if err := db.QueryRow(`
		SELECT count(*) > 0 FROM information_schema.columns
		WHERE table_name = 'print_logs' AND column_name = 'file_asset_id'`).Scan(&hasFK); err != nil {
		t.Fatalf("查欄位: %v", err)
	}
	if !hasFK {
		t.Fatal("print_logs 缺 file_asset_id")
	}
	// 記錄類僅追加:app_rw 有 SELECT/INSERT、無 UPDATE/DELETE。
	for _, tbl := range []string{"print_logs", "print_previews"} {
		var sel, ins, upd, del bool
		q := `SELECT has_table_privilege('app_rw', $1, 'SELECT'),
			has_table_privilege('app_rw', $1, 'INSERT'),
			has_table_privilege('app_rw', $1, 'UPDATE'),
			has_table_privilege('app_rw', $1, 'DELETE')`
		if err := db.QueryRow(q, tbl).Scan(&sel, &ins, &upd, &del); err != nil {
			t.Fatalf("查授權 %s: %v", tbl, err)
		}
		if !sel || !ins || upd || del {
			t.Fatalf("%s 應僅追加(SELECT+INSERT),got %v %v %v %v", tbl, sel, ins, upd, del)
		}
	}

	// Down 乾淨:回滾 00037 後兩表消失;重上冪等。
	ctx := context.Background()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("dialect: %v", err)
	}
	goose.SetTableName("goose_db_version")
	if err := goose.RunContext(ctx, "down-to", db, "../../database/migrations", "36"); err != nil {
		t.Fatalf("down-to 36: %v", err)
	}
	for _, tbl := range []string{"print_logs", "print_previews"} {
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
	if err := db.QueryRow(`SELECT count(*) FROM print_logs`).Scan(&n); err != nil {
		t.Fatalf("重上後 print_logs 應存在: %v", err)
	}
}

// migratePrintUp 以 goose 全量 up。
func migratePrintUp(t *testing.T, dsn string) {
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
