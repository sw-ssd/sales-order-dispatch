package dbtenant

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"log"
	"slices"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/mattn/go-sqlite3" // sqlite 記憶體測試驅動(與專案既有單元測試同款)

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
)

// sqlOpen 開一個記憶體 sqlite 並建最小表:本檔只需「能開交易與查詢」(RLS 語句為空,
// 故不會在 sqlite 上執行 SET LOCAL;scope=all 的 SystemScopeTx 一律不在此測)。
func sqlOpen(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", "file:dbtenant?mode=memory&cache=shared&_fk=1")
	if err != nil {
		t.Fatalf("開 sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS ping (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("建表: %v", err)
	}
	return db
}

// Client 在沒有請求交易時必須退回傳入的 client(CLI／seed／既有 sqlite 測試都靠這條)。
func TestClientFallsBackWithoutRequestTx(t *testing.T) {
	db := sqlOpen(t)
	client := ent.NewClient(ent.Driver(Wrap(entsql.OpenDB(dialect.SQLite, db))))
	t.Cleanup(func() { _ = client.Close() })

	if got := Client(context.Background(), client); got != client {
		t.Fatal("無請求交易時應回傳 fallback client")
	}
}

func TestClientUsesRequestTx(t *testing.T) {
	db := sqlOpen(t)
	client := ent.NewClient(ent.Driver(Wrap(entsql.OpenDB(dialect.SQLite, db))))
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	tx, err := client.Tx(ctx)
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	got := Client(WithTenantTx(ctx, tx), client)
	if got == client {
		t.Fatal("有請求交易時不得回傳 fallback client")
	}
	if got != tx.Client() {
		t.Fatal("有請求交易時應回傳請求交易的 client")
	}
	if _, ok := TxFrom(WithTenantTx(ctx, tx)); !ok {
		t.Fatal("TxFrom 應取得請求交易")
	}
	if _, ok := TxFrom(ctx); ok {
		t.Fatal("未注入時 TxFrom 應回 false")
	}
}

// spyTx／spyDriver 記錄交易內執行的語句:SET LOCAL 的「有無與順序」是 RLS 能否生效的
// 全部關鍵,但 sqlite 不支援 SET LOCAL,故以假 driver 直接觀測裝飾器行為。
type spyTx struct{ execs []string }

func (t *spyTx) Exec(_ context.Context, query string, _, _ any) error {
	t.execs = append(t.execs, query)
	return nil
}
func (t *spyTx) Query(context.Context, string, any, any) error { return nil }
func (t *spyTx) Commit() error                                 { return nil }
func (t *spyTx) Rollback() error                               { return nil }

type spyDriver struct{ tx spyTx }

func (d *spyDriver) Exec(context.Context, string, any, any) error  { return nil }
func (d *spyDriver) Query(context.Context, string, any, any) error { return nil }
func (d *spyDriver) Tx(context.Context) (dialect.Tx, error)        { return &d.tx, nil }
func (d *spyDriver) Close() error                                  { return nil }
func (d *spyDriver) Dialect() string                               { return dialect.SQLite }

// Wrap 必須在交易開好、任何業務查詢之前把 ctx 的 RLS scope 套進交易(順序即 RLSStatements 的順序)。
func TestWrapAppliesScopeStatementsInTx(t *testing.T) {
	scope := auth.RLSScope{
		UserID: "u1", CompanyID: "c1", DepartmentID: "d1",
		DataScope: auth.DataScopeDepartment, CompanyActive: true,
	}
	inner := &spyDriver{}
	if _, err := Wrap(inner).Tx(auth.WithRLS(context.Background(), scope)); err != nil {
		t.Fatalf("開交易: %v", err)
	}
	if want := auth.RLSStatements(scope); !slices.Equal(inner.tx.execs, want) {
		t.Fatalf("交易內應套用 scope 語句\nwant %q\ngot  %q", want, inner.tx.execs)
	}
}

// 空 scope(未登入／CLI／seed)不得產生任何語句:否則交易一開就爆(sqlite 更直接紅)。
func TestWrapAppliesNothingWithoutScope(t *testing.T) {
	inner := &spyDriver{}
	if _, err := Wrap(inner).Tx(context.Background()); err != nil {
		t.Fatalf("開交易: %v", err)
	}
	if len(inner.tx.execs) != 0 {
		t.Fatalf("空 scope 不應執行任何語句,got %q", inner.tx.execs)
	}
}

// failDriver 讓 client.Tx 或 Commit 以含哨兵字串的錯誤失敗,用來驗證錯誤處理路徑。
type failDriver struct {
	txErr     error // 非 nil 表示開交易即失敗
	commitErr error // 非 nil 表示開得起來但提交失敗
	tx        failTx
}

type failTx struct{ commitErr error }

func (t *failTx) Exec(context.Context, string, any, any) error  { return nil }
func (t *failTx) Query(context.Context, string, any, any) error { return nil }
func (t *failTx) Commit() error                                 { return t.commitErr }
func (t *failTx) Rollback() error                               { return nil }

func (d *failDriver) Exec(context.Context, string, any, any) error  { return nil }
func (d *failDriver) Query(context.Context, string, any, any) error { return nil }
func (d *failDriver) Close() error                                  { return nil }
func (d *failDriver) Dialect() string                               { return dialect.SQLite }
func (d *failDriver) Tx(context.Context) (dialect.Tx, error) {
	if d.txErr != nil {
		return nil, d.txErr
	}
	d.tx.commitErr = d.commitErr
	return &d.tx, nil
}

// 交易開不起來／提交失敗時:對客戶端只准看到固定訊息(SQLSTATE、policy 名與 SET LOCAL
// 語句文字不得外洩 —— connect-go 逐字轉送 Message()),而根因必須落 server log
// (chi Logger 不記錄 handler error,線上追查只剩 log 一途)。
func TestInterceptorHidesRootCauseButLogsIt(t *testing.T) {
	const sentinel = "SENTINEL-DB-ERROR"
	cases := map[string]struct {
		driver *failDriver
		want   string
	}{
		"開啟交易失敗": {&failDriver{txErr: errors.New(sentinel)}, "開啟租戶交易失敗"},
		"提交失敗":   {&failDriver{commitErr: errors.New(sentinel)}, "提交交易失敗"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var logged bytes.Buffer
			prev := log.Writer()
			log.SetOutput(&logged)
			t.Cleanup(func() { log.SetOutput(prev) })

			client := ent.NewClient(ent.Driver(tc.driver))
			_, err := Interceptor(client).WrapUnary(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
				return nil, nil
			})(context.Background(), nil)
			if err == nil {
				t.Fatal("應回傳錯誤")
			}
			var connectErr *connect.Error
			if !errors.As(err, &connectErr) {
				t.Fatalf("應為 connect 錯誤,got %T: %v", err, err)
			}
			if got := connectErr.Message(); got != tc.want {
				t.Fatalf("對客戶端的訊息必須恰為 %q(不得含根因／SQLSTATE),got %q", tc.want, got)
			}
			if !strings.Contains(logged.String(), sentinel) {
				t.Fatalf("根因必須落 server log,got %q", logged.String())
			}
		})
	}
}
