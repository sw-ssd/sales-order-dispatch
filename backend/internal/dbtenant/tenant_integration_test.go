//go:build integration

package dbtenant_test

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"testing"

	"connectrpc.com/connect"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationTenantTxCommitAndRollback 驗證 interceptor 的交易語意：
// 成功 → 交易內寫入落地；handler 回傳錯誤 → 整筆回滾；handler 內取得請求交易；
// 且 scope 的 SET LOCAL 真的落在該交易的 session 裡（RLS 啟用後缺席即全表不可見）。
func TestIntegrationTenantTxCommitAndRollback(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	defer sqlDB.Close()
	if err := goose.RunContext(t.Context(), "up", sqlDB, "../../database/migrations"); err != nil {
		t.Fatalf("套用遷移: %v", err)
	}

	// 業務 client 必須經 dbtenant.NewClient 建立(RLS 裝飾器才生效):本測試同時驗裝飾器本身。
	client := dbtenant.NewClient(sqlDB)
	t.Cleanup(func() { _ = client.Close() })

	scope := auth.RLSScope{
		UserID: "1", CompanyID: "1", DataScope: auth.DataScopeAll, CompanyActive: true,
	}
	ctx := auth.WithRLS(context.Background(), scope)

	t.Run("scope 的 SET LOCAL 落在請求交易內", func(t *testing.T) {
		tx, err := dbtenant.Wrap(entsql.OpenDB(dialect.Postgres, sqlDB)).Tx(auth.WithRLS(t.Context(), scope))
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()

		var rows entsql.Rows
		if err := tx.Query(t.Context(), "SELECT current_setting('app.current_user_id', true)", []any{}, &rows); err != nil {
			t.Fatalf("查 GUC: %v", err)
		}
		defer rows.Close()
		if !rows.Next() {
			t.Fatal("current_setting 應回一列")
		}
		var got string
		if err := rows.Scan(&got); err != nil {
			t.Fatalf("掃描: %v", err)
		}
		if got != scope.UserID {
			t.Fatalf("交易內 app.current_user_id 應為 %q,got %q", scope.UserID, got)
		}
	})

	// handler 以 dbtenant.Client 取 client：成功路徑寫入一列公司，失敗路徑寫入後回傳錯誤。
	call := func(fail bool) (connect.AnyResponse, error) {
		return dbtenant.Interceptor(client).WrapUnary(func(
			handlerCtx context.Context, _ connect.AnyRequest,
		) (connect.AnyResponse, error) {
			if _, ok := dbtenant.TxFrom(handlerCtx); !ok {
				t.Error("handler 內必須取得請求交易")
			}
			if _, err := dbtenant.Client(handlerCtx, client).Company.Create().
				SetName("交易測試").
				SetIdentifier("RLS-TX-" + strconv.FormatBool(fail)).
				Save(handlerCtx); err != nil {
				return nil, err
			}
			if fail {
				return nil, connect.NewError(connect.CodeInternal, errors.New("刻意失敗"))
			}
			return nil, nil
		})(ctx, nil)
	}

	if _, err := call(false); err != nil {
		t.Fatalf("成功路徑: %v", err)
	}
	var committed int
	if err := sqlDB.QueryRow(
		`SELECT count(*) FROM companies WHERE identifier = 'RLS-TX-false'`,
	).Scan(&committed); err != nil {
		t.Fatalf("查已提交列: %v", err)
	}
	if committed != 1 {
		t.Fatalf("成功路徑的交易應已提交,got %d 列", committed)
	}

	if _, err := call(true); err == nil {
		t.Fatal("失敗路徑應回傳錯誤")
	}
	// 「查不到列」單獨看不足以證明回滾:未結束的交易同樣查不到未提交的列。
	// 連線已歸還 + 查不到列 ⟹ 交易真的以回滾收場(若提交,下一句會查到 1 列)。
	if inUse := sqlDB.Stats().InUse; inUse != 0 {
		t.Fatalf("失敗路徑結束後請求交易必須已結束(連線歸還),got InUse=%d", inUse)
	}
	var rolled int
	if err := sqlDB.QueryRow(
		`SELECT count(*) FROM companies WHERE identifier = 'RLS-TX-true'`,
	).Scan(&rolled); err != nil {
		t.Fatalf("查回滾列: %v", err)
	}
	if rolled != 0 {
		t.Fatalf("失敗路徑的交易必須整筆回滾,got %d 列", rolled)
	}
}
