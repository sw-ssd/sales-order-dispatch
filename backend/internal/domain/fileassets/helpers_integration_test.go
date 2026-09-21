//go:build integration

package fileassets_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/ent"
)

// itoaFile 整數字串(測試輔助)。
func itoaFile(n int) string { return strconv.Itoa(n) }

// migrateBusinessUpFile 以 goose 全量 up(與 services 同路徑同版本表)。
func migrateBusinessUpFile(t *testing.T, dsn string) {
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
	if err := goose.RunContext(t.Context(), "up", db, "../../../database/migrations"); err != nil {
		t.Fatalf("goose up: %v", err)
	}
}

// openPGEntClientFile 以 goose 建出的庫開 ent client(不重建 schema)。
func openPGEntClientFile(t *testing.T, dsn string) (*sql.DB, *ent.Client) {
	t.Helper()
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, sqlDB)))
	t.Cleanup(func() { _ = client.Close() })
	return sqlDB, client
}

// seedSecondCompany 建第二家公司(跨公司下載用)。
func seedSecondCompany(t *testing.T, ctx context.Context, db *ent.Client) int {
	t.Helper()
	var coID int
	seedTxFile(t, db, func(tx *ent.Tx) error {
		co, err := tx.Company.Create().SetName("檔案公司B").SetIdentifier("F2-" + t.Name()).
			SetStatus("active").Save(ctx)
		if err != nil {
			return err
		}
		coID = co.ID
		return nil
	})
	return coID
}

func jsonDecode(resp *http.Response, v any) error {
	return json.NewDecoder(resp.Body).Decode(v)
}
