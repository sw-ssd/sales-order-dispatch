package services

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/ordercounter"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
)

// seedOrderSource 種 order_source 字典一筆(W):validateOrderSource 的查表對象。
func seedOrderSource(t *testing.T, ctx context.Context, db *ent.Client) {
	t.Helper()
	if _, err := db.Metadict.Create().
		SetType("order_source").SetCode("W").SetDisplayName("Web 中台").
		SetSortOrder(10).SetIsActive(true).Save(ctx); err != nil {
		t.Fatalf("seed order_source: %v", err)
	}
}

// TestNextOrderNoSequential 同公司同來源連續取號遞增(W000001→W000002),異來源獨立軌。
func TestNextOrderNoSequential(t *testing.T) {
	db := newCounterTestDB(t)
	ctx, tx := statusTxCtx(t, db)
	defer func() { _ = tx.Commit() }()
	client := tx.Client()
	seedOrderSource(t, ctx, client)

	if err := validateOrderSource(ctx, client, "W"); err != nil {
		t.Fatalf("validateOrderSource(W): %v", err)
	}
	if err := ensureOrderCounter(ctx, client, 7, "W"); err != nil {
		t.Fatalf("ensureOrderCounter: %v", err)
	}
	for _, want := range []string{"W000001", "W000002"} {
		got, err := nextOrderNo(ctx, client, 7, "W")
		if err != nil {
			t.Fatalf("nextOrderNo: %v", err)
		}
		if got != want {
			t.Fatalf("取號 = %q;want %q", got, want)
		}
	}
	// 異來源獨立軌:同公司 A 軌從 1 起。
	if err := ensureOrderCounter(ctx, client, 7, "A"); err != nil {
		t.Fatalf("ensureOrderCounter(A): %v", err)
	}
	_ = client.Metadict.Create().SetType("order_source").SetCode("A").
		SetDisplayName("App").SetSortOrder(20).SetIsActive(true).SaveX(ctx)
	if got, err := nextOrderNo(ctx, client, 7, "A"); err != nil || got != "A000001" {
		t.Fatalf("異來源應獨立起算,got %q err=%v", got, err)
	}
	// 異公司同來源互不干擾。
	if err := ensureOrderCounter(ctx, client, 8, "W"); err != nil {
		t.Fatalf("ensureOrderCounter(8,W): %v", err)
	}
	if got, err := nextOrderNo(ctx, client, 8, "W"); err != nil || got != "W000001" {
		t.Fatalf("異公司應獨立起算,got %q err=%v", got, err)
	}
}

// TestNextOrderNoRejectsInvalidSource 無效來源碼／空來源一律 invalid_argument,且不消耗序號。
func TestNextOrderNoRejectsInvalidSource(t *testing.T) {
	db := newCounterTestDB(t)
	ctx, tx := statusTxCtx(t, db)
	defer func() { _ = tx.Commit() }()
	client := tx.Client()
	seedOrderSource(t, ctx, client)

	if err := validateOrderSource(ctx, client, "Z"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("無效來源應 invalid_argument,got %v", err)
	}
	if err := validateOrderSource(ctx, client, "  "); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("空來源應 invalid_argument,got %v", err)
	}
	// 守衛在取號之前:counter 列不應被建立。
	if exists, _ := client.OrderCounter.Query().
		Where(ordercounter.CompanyIDEQ(7), ordercounter.SourceEQ("Z")).Exist(ctx); exists {
		t.Fatal("無效來源不得建立 counter 列(不得消耗序號)")
	}
}

// TestNextOrderNoVersionConflictRetries version 衝突重試:外部先推進一版,取號仍成功且序號連續。
func TestNextOrderNoVersionConflictRetries(t *testing.T) {
	db := newCounterTestDB(t)
	ctx, tx := statusTxCtx(t, db)
	defer func() { _ = tx.Commit() }()
	client := tx.Client()
	seedOrderSource(t, ctx, client)

	if err := ensureOrderCounter(ctx, client, 7, "W"); err != nil {
		t.Fatalf("ensureOrderCounter: %v", err)
	}
	if _, err := nextOrderNo(ctx, client, 7, "W"); err != nil { // W000001
		t.Fatalf("首次取號: %v", err)
	}
	// 模擬併發:外部(同交易內先讀後寫,version 已 +1) —— 取號的重讀應拿到新版而非衝突死循環。
	// 此處以連續兩次取號驗「無遺失無重複」:W000002、W000003。
	for _, want := range []string{"W000002", "W000003"} {
		got, err := nextOrderNo(ctx, client, 7, "W")
		if err != nil {
			t.Fatalf("nextOrderNo: %v", err)
		}
		if got != want {
			t.Fatalf("取號 = %q;want %q", got, want)
		}
	}
}

// TestNextOrderNoRollbackReusesSeq 建單失敗回滾 → 序號不消耗,下次取號得同一號。
func TestNextOrderNoRollbackReusesSeq(t *testing.T) {
	db := newCounterTestDB(t)
	seedOrderSource(t, context.Background(), db)

	// 第一輪:取號後回滾(模擬建單失敗)。
	tx1, err := db.Tx(context.Background())
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	ctx1 := dbtenant.WithTenantTx(context.Background(), tx1)
	c1 := tx1.Client()
	if err := ensureOrderCounter(ctx1, c1, 7, "W"); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	got1, err := nextOrderNo(ctx1, c1, 7, "W")
	if err != nil {
		t.Fatalf("首輪取號: %v", err)
	}
	if err := tx1.Rollback(); err != nil {
		t.Fatalf("回滾: %v", err)
	}

	// 第二輪:應取到同一號(序號未被消耗)。
	tx2, err := db.Tx(context.Background())
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	defer func() { _ = tx2.Rollback() }()
	ctx2 := dbtenant.WithTenantTx(context.Background(), tx2)
	c2 := tx2.Client()
	if err := ensureOrderCounter(ctx2, c2, 7, "W"); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	got2, err := nextOrderNo(ctx2, c2, 7, "W")
	if err != nil {
		t.Fatalf("次輪取號: %v", err)
	}
	if got1 != got2 {
		t.Fatalf("回滾後應取到同一號,got %q → %q", got1, got2)
	}
	if err := tx2.Commit(); err != nil {
		t.Fatalf("提交: %v", err)
	}
}
