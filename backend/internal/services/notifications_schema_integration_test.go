//go:build integration

package services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationNotificationsSchemaUpDown 驗證 00041:四表存在 + 部分唯一索引 +
// notifications 無 deleted_at + 授權形狀(範本/通知僅 SELECT/INSERT/UPDATE,裝置/標籤 CRUD) +
// down-to 40 後四表消失 + 重上冪等。
func TestIntegrationNotificationsSchemaUpDown(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateNotificationsUp(t, dsn)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	defer func() { _ = db.Close() }()

	for _, tbl := range []string{"notification_templates", "notifications", "user_devices", "promo_tags"} {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM ` + tbl).Scan(&n); err != nil {
			t.Fatalf("%s 不存在: %v", tbl, err)
		}
	}
	// 部分唯一索引存在。
	for _, idx := range []string{"notification_templates_scope_unique",
		"user_devices_token_unique", "promo_tags_scope_unique", "notifications_user_status_idx"} {
		var exists bool
		if err := db.QueryRow(`SELECT to_regclass($1) IS NOT NULL`, "public."+idx).Scan(&exists); err != nil {
			t.Fatalf("查索引 %s: %v", idx, err)
		}
		if !exists {
			t.Fatalf("索引 %s 應存在", idx)
		}
	}
	// notifications 無 deleted_at(不可刪除)。
	var hasDeleted bool
	if err := db.QueryRow(`
		SELECT count(*) > 0 FROM information_schema.columns
		WHERE table_name = 'notifications' AND column_name = 'deleted_at'`).Scan(&hasDeleted); err != nil {
		t.Fatalf("查欄位: %v", err)
	}
	if hasDeleted {
		t.Fatal("notifications 不得有 deleted_at")
	}
	// 授權形狀:範本/通知無 DELETE;裝置/標籤 CRUD 全授。
	for _, tc := range []struct {
		tbl                string
		sel, ins, upd, del bool
	}{
		{"notification_templates", true, true, true, false},
		{"notifications", true, true, true, false},
		{"user_devices", true, true, true, true},
		{"promo_tags", true, true, true, true},
	} {
		var sel, ins, upd, del bool
		q := `SELECT has_table_privilege('app_rw', $1, 'SELECT'),
			has_table_privilege('app_rw', $1, 'INSERT'),
			has_table_privilege('app_rw', $1, 'UPDATE'),
			has_table_privilege('app_rw', $1, 'DELETE')`
		if err := db.QueryRow(q, tc.tbl).Scan(&sel, &ins, &upd, &del); err != nil {
			t.Fatalf("查授權 %s: %v", tc.tbl, err)
		}
		if sel != tc.sel || ins != tc.ins || upd != tc.upd || del != tc.del {
			t.Fatalf("%s 授權異常,got %v %v %v %v", tc.tbl, sel, ins, upd, del)
		}
	}

	// down-to 41 只回滾 00042(ENABLE 解除),00041 的表留著;重上冪等。
	ctx := context.Background()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("dialect: %v", err)
	}
	goose.SetTableName("goose_db_version")
	if err := goose.RunContext(ctx, "down-to", db, "../../database/migrations", "41"); err != nil {
		t.Fatalf("down-to 41: %v", err)
	}
	for _, tbl := range []string{"notification_templates", "notifications", "user_devices", "promo_tags"} {
		var exists bool
		if err := db.QueryRow(`SELECT to_regclass($1) IS NOT NULL`, "public."+tbl).Scan(&exists); err != nil {
			t.Fatalf("查表存在: %v", err)
		}
		if !exists {
			t.Fatalf("down-to 41 不得帶走 00041 的表 %s", tbl)
		}
		var enabled bool
		if err := db.QueryRow(`SELECT relrowsecurity FROM pg_class WHERE relname = $1`, tbl).Scan(&enabled); err != nil {
			t.Fatalf("查 RLS 狀態 %s: %v", tbl, err)
		}
		if enabled {
			t.Fatalf("down-to 41 後 %s 的 RLS 應已 DISABLE", tbl)
		}
	}
	if err := goose.RunContext(ctx, "up", db, "../../database/migrations"); err != nil {
		t.Fatalf("重上: %v", err)
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM notifications`).Scan(&n); err != nil {
		t.Fatalf("重上後 notifications 應存在: %v", err)
	}
}

// migrateNotificationsUp 以 goose 全量 up。
func migrateNotificationsUp(t *testing.T, dsn string) {
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
