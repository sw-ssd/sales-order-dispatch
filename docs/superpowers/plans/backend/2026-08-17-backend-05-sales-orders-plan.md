# Backend 05 — 銷售訂單 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 實作銷售訂單後端 — 訂單三表 schema、樂觀鎖取號、狀態機、CRUD 與異動軌跡、下單組裝邏輯(單位換算、手打別名、客戶專屬清單守衛、來源記錄、偏好送貨日順延)。

**Architecture:** 依細部文件 `detail/05-sales-orders.md` 實作原計畫 Task 4.1 / 4.2。狀態機集中於單一模組,所有轉移走 `Transition()`(條件更新 + 事件 + 必要時稽核,同一交易);取號與建單同交易(D7);不存任何金額欄位(D12)。

**Tech Stack:** Go 1.25、Ent、Connect-RPC、`shopspring/decimal`(數量/換算,對齊 04 計畫契約)、testcontainers-go。

**Spec 來源:** 細部文件 `docs/superpowers/plans/backend/detail/05-sales-orders.md`;共通規則見 `00-index.md` §3。

## Global Constraints

- module 路徑:`github.com/salesorder/sales-order-1.0/backend`。
- **金額禁令(D12)**:`sales_orders` / `sales_order_items` 不得出現 price/amount/subtotal/tax/discount/total/currency 欄位;schema 測試斷言。
- `sales_order_events` 僅追加:無 `updated_at`/`deleted_at`,migration REVOKE UPDATE/DELETE。
- 數量(`qty`/`base_qty`)用 `decimal.Decimal`(04 計畫契約);手打品項 `base_qty = qty`。
- 訂單編號:來源碼(metadicts `order_source`,如 `W`/`A`)+ 6 位補零;`order_counters` 依 `(company_id, source)` 一列,樂觀鎖 + 自動重試上限 5 次;取號僅在呼叫方交易內執行。
- 狀態機:合法轉移僅 `pending→processing`、`processing→pending`(取消派車)、`processing→completed`、`pending→cancelled`、`completed→voided`;終態 `cancelled`/`voided`;條件更新(WHERE status=前值)影響 0 列 → `failed_precondition`;任何轉移遞增 `version`(D14)。
- 客戶帳號下單:品項必須來自自己 `customer_products`;忽略請求 `customer_id` 越權值並**拒絕**(細部 4.2.3 驗收採「拒絕以避免誤單」);手打品項為業務限定。
- 偏好送貨日順延(D26):`preferred_delivery_days` 為長度 6 的 bool 陣列(週一至週六);未勾選 → 維持;非勾選日 → 順延至下一勾選日(至多 7 天命中);時區固定 UTC+8。
- 稽核:`dispatch_cancel` / `void` / 軟刪除同交易寫 `audit_logs`(03 計畫 `audit.NewPGRecorder`);寫入失敗整體回滾(D18)。
- 通知(D18,對齊 07 計畫契約):通知記錄(fcm + in_app 各一筆 `pending`)於下單**同一交易**內經 `triggers.OnOrderCreated(ctx, tx, order, actor)`(07 計畫 T5)建立並收集通知 ID;**提交後**呼叫 `notification.Dispatch(ctx, db, sender, rec, ids)`(07 計畫 T4)發送;發送失敗僅標 `failed` 不重試、不回滾業務(D16)。業務下單推客戶子帳號群(排除主帳號);客戶自行下單不通知(07 觸發函式內建路由,本計畫不重複)。
- 錯誤統一 Connect code;樂觀鎖/狀態衝突一律 `failed_precondition`。

## File Structure

| 檔案 | 職責 | 建立於 |
|---|---|---|
| `backend/ent/schema/{salesorder,salesorderitem,salesorderevent,ordercounter}.go` | 訂單四實體 | Task 1-2 |
| `backend/internal/domain/salesorders/numbering.go` | 取號 | Task 2 |
| `backend/internal/domain/salesorders/statemachine.go` | 狀態機 | Task 3 |
| `backend/internal/domain/salesorders/{handler,usecase,repo}.go` | RPC 與業務 | Task 4-5 |
| `backend/internal/domain/salesorders/ordering.go` | 下單組裝(換算/別名/守衛/順延)+ 通知觸發掛接 | Task 5 |
| `backend/proto/v1/salesorder.proto` | SalesOrderService | Task 4 |

---

### Task 1: 訂單三表 schema(細部 4.1.1)

**Files:**
- Create: `backend/ent/schema/salesorder.go`、`salesorderitem.go`、`salesorderevent.go`
- Create: `backend/database/migrations/00010_sales_orders.sql`(部分唯一索引、RLS、events REVOKE)
- Test: `backend/internal/domain/salesorders/schema_test.go`

**Interfaces:**
- Consumes: 04 計畫 `customers`/`products`;01-auth RLS 注入。
- Produces: `ent.SalesOrder` / `ent.SalesOrderItem` / `ent.SalesOrderEvent`(欄位完全依細部 4.1.1 介面欄:`order_no`、`source`、`status` enum、`expected_delivery_date`(date)、`sales_rep_id`、`dispatched_at`/`dispatched_by`/`route_id`/`delivery_sequence` nullable、`version` 預設 1;items 含 `product_id` nullable、`display_name`、`qty`/`base_qty` decimal、`unit`、`cutting_spec_id`、`special_cut_note`、`warehouse_id`、`sort_order`;events 含 `event_type` enum(create/edit/dispatch/dispatch_cancel/route_assign/cancel/complete/void,共八值 — `route_assign` 由 08-dispatch 計畫寫入,細部 5.1.1)、`actor_id`、`reason`、`payload` JSONB,無 updated/deleted)。

- [ ] **Step 1: 寫失敗測試(金額禁令斷言、部分唯一、events 不可變)**

```go
// backend/internal/domain/salesorders/schema_test.go
package salesorders_test

import (
	"context"
	"strings"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

// bannedMoneyFields D12:主表與明細不得含金額語意欄位。
var bannedMoneyFields = []string{"price", "amount", "subtotal", "tax", "discount", "total", "currency"}

func TestNoMoneyFields(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	for _, table := range []string{"sales_orders", "sales_order_items"} {
		rows, err := c.DB().QueryContext(ctx,
			`SELECT column_name FROM information_schema.columns WHERE table_name = $1`, table)
		if err != nil { t.Fatal(err) }
		defer rows.Close()
		for rows.Next() {
			var col string
			if err := rows.Scan(&col); err != nil { t.Fatal(err) }
			for _, banned := range bannedMoneyFields {
				if strings.Contains(col, banned) {
					t.Fatalf("D12 violated: %s.%s", table, col)
				}
			}
		}
	}
}

func TestOrderNoPartialUnique(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	// seed 兩公司(模式同 03/04 計畫 seedTenant);同公司同 order_no 拒絕、異公司同號可存
}

func TestEventsImmutable(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	_ = ctx
	// 建一筆 event 後嘗試 UPDATE/DELETE → DB 權限拒絕(migration REVOKE);
	// 以 SQL 直連驗證:UPDATE sales_order_events SET ... → error
	_ = c
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/salesorders/ -v`
Expected: FAIL — `ent.SalesOrder` 未定義。

- [ ] **Step 3: 實作 schema(三實體,欄位依細部 4.1.1 逐字)**

```go
// backend/ent/schema/salesorder.go(重點欄位;慣例欄位同前)
type SalesOrder struct{ ent.Schema }

func (SalesOrder) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("company_id", uuid.UUID{}),
		field.UUID("department_id", uuid.UUID{}),
		field.String("order_no").NotEmpty(),
		field.UUID("customer_id", uuid.UUID{}),
		field.String("source").NotEmpty().MaxLen(8), // metadicts order_source 碼
		field.Enum("status").Values("pending", "processing", "completed", "cancelled", "voided").Default("pending"),
		field.Time("expected_delivery_date"),
		field.UUID("sales_rep_id", uuid.UUID{}).Optional().Nillable(),
		field.Text("note").Optional(),
		field.Time("dispatched_at").Optional().Nillable(),
		field.UUID("dispatched_by", uuid.UUID{}).Optional().Nillable(),
		field.UUID("route_id", uuid.UUID{}).Optional().Nillable(),
		field.Int("delivery_sequence").Optional().Nillable(),
		field.Int("version").Default(1),
		field.UUID("created_by", uuid.UUID{}),
		// created_at / updated_at / deleted_at 同慣例
	}
}
// Indexes: (company_id, order_no) 部分唯一 WHERE deleted_at IS NULL;
//          (company_id, department_id, status, expected_delivery_date) 複合
```

`SalesOrderItem`:`qty` / `base_qty` 用 `field.Other(qty, decimal.Decimal{})` 或 `field.Float` — 採 `field.Other` + `SchemaType` decimal(對齊 04 計畫 `shopspring/decimal` 契約);其餘欄位依細部 4.1.1。
`SalesOrderEvent`:`event_type` enum 八值(七值狀態機事件 + `route_assign`,08-dispatch 計畫寫入)、`reason` 可空、`payload` JSONB;**無** updated_at/deleted_at。

`00010_sales_orders.sql`:

```sql
-- +goose Up
ALTER TABLE sales_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales_order_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales_order_events ENABLE ROW LEVEL SECURITY;
-- 三表同一組 policy:tenant_isolation(company)+ data_scope(department/self 比對 customer_id)
-- (policy 本體比照 01-auth Task 3 的 users 模式,self 範圍比對 app.current_customer_id)
REVOKE UPDATE, DELETE ON sales_order_events FROM PUBLIC; -- 僅追加(細部 4.1.1 步驟 6)
-- +goose Down
GRANT UPDATE, DELETE ON sales_order_events TO PUBLIC;
ALTER TABLE sales_order_events DISABLE ROW LEVEL SECURITY;
ALTER TABLE sales_order_items DISABLE ROW LEVEL SECURITY;
ALTER TABLE sales_orders DISABLE ROW LEVEL SECURITY;
```

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go generate ./ent/ && go test ./internal/domain/salesorders/ -v`
Expected: PASS。

```bash
git add backend/ent/schema backend/database/migrations/00010_sales_orders.sql backend/internal/domain/salesorders
git commit -m "feat(backend): 訂單三表 schema,金額禁令與 events 僅追加(4.1.1)"
```

---

### Task 2: order_counters 樂觀鎖取號(細部 4.1.2,D7)

**Files:**
- Create: `backend/ent/schema/ordercounter.go`
- Create: `backend/internal/domain/salesorders/numbering.go`
- Test: `backend/internal/domain/salesorders/numbering_test.go`

**Interfaces:**
- Produces: `ent.OrderCounter`(`company_id`、`source`、`next_seq`、`version`;唯一 `(company_id, source)`);`salesorders.NextOrderNo(ctx context.Context, tx *ent.Tx, companyID uuid.UUID, source string) (string, error)` — **僅接受已開啟交易**,不自行開交易;重試上限 5。

- [ ] **Step 1: 寫失敗測試(連續、雙軌獨立、併發、回滾不消耗)**

```go
// backend/internal/domain/salesorders/numbering_test.go
package salesorders_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/domain/salesorders"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestNextOrderNoSequence(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID := seedCompany(t, c) // 輔助:建公司回 ID(比照 03/04 模式)

	tx, _ := c.Tx(ctx)
	n1, err := salesorders.NextOrderNo(ctx, tx, coID, "W")
	if err != nil { t.Fatal(err) }
	n2, err := salesorders.NextOrderNo(ctx, tx, coID, "W")
	if err != nil { t.Fatal(err) }
	m1, err := salesorders.NextOrderNo(ctx, tx, coID, "A")
	if err != nil { t.Fatal(err) }
	_ = tx.Commit()
	if n1 != "W000001" || n2 != "W000002" || m1 != "A000001" {
		t.Fatalf("numbering: %s %s %s", n1, n2, m1)
	}
	// 無效來源碼
	tx, _ = c.Tx(ctx)
	if _, err := salesorders.NextOrderNo(ctx, tx, coID, "X"); err == nil {
		t.Fatal("unknown source should be invalid_argument")
	}
	_ = tx.Rollback()
}

func TestNextOrderNoConcurrent(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID := seedCompany(t, c)

	const n = 20
	results := make(chan string, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tx, err := c.Tx(ctx)
			if err != nil { t.Errorf("tx: %v", err); return }
			no, err := salesorders.NextOrderNo(ctx, tx, coID, "W")
			if err != nil { t.Errorf("take: %v", err); _ = tx.Rollback(); return }
			if err := tx.Commit(); err != nil { t.Errorf("commit: %v", err); return }
			results <- no
		}()
	}
	wg.Wait()
	close(results)
	seen := map[string]bool{}
	for no := range results {
		if seen[no] {
			t.Fatalf("duplicate order_no: %s", no)
		}
		seen[no] = true
	}
	if len(seen) != n {
		t.Fatalf("want %d unique, got %d", n, len(seen))
	}
}

func TestRollbackDoesNotConsume(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID := seedCompany(t, c)

	tx, _ := c.Tx(ctx)
	first, err := salesorders.NextOrderNo(ctx, tx, coID, "W")
	if err != nil { t.Fatal(err) }
	_ = tx.Rollback() // 建單失敗情境

	tx2, _ := c.Tx(ctx)
	second, err := salesorders.NextOrderNo(ctx, tx2, coID, "W")
	if err != nil { t.Fatal(err) }
	_ = tx2.Commit()
	if first != second {
		t.Fatalf("rollback consumed seq: %s then %s", first, second)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/salesorders/ -run TestNextOrderNo -v`
Expected: FAIL。

- [ ] **Step 3: 實作**

`backend/ent/schema/ordercounter.go`:

```go
// OrderCounter 訂單取號計數器(每公司每來源一列;細部 4.1.2)。
type OrderCounter struct{ ent.Schema }

func (OrderCounter) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("company_id", uuid.UUID{}),
		field.String("source").NotEmpty().MaxLen(8),
		field.Int("next_seq").Default(1),
		field.Int("version").Default(1),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (OrderCounter) Indexes() []ent.Index {
	return []ent.Index{index.Fields("company_id", "source").Unique()}
}
```

`numbering.go`:

```go
package salesorders

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entoc "github.com/salesorder/sales-order-1.0/backend/ent/ordercounter"
	entmd "github.com/salesorder/sales-order-1.0/backend/ent/metadict"
)

const maxTakeRetries = 5

// NextOrderNo 於呼叫方交易內取號(細部 4.1.2):樂觀鎖更新,衝突自動重試,上限 5 次。
// 回滾由呼叫方負責 — 序號不消耗(計數器遞增與建單同生滅)。
func NextOrderNo(ctx context.Context, tx *ent.Tx, companyID uuid.UUID, source string) (string, error) {
	// 來源碼驗證:系統級 order_source 字典有效且啟用(細部 4.1.2 錯誤處理)
	valid, err := tx.Metadict.Query().Where(
		entmd.TypeEQ("order_source"), entmd.CodeEQ(source),
		entmd.DepartmentIDIsNil(), entmd.IsActiveEQ(true), entmd.DeletedAtIsNil(),
	).Exist(ctx)
	if err != nil {
		return "", connect.NewError(connect.CodeInternal, err)
	}
	if !valid {
		return "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown order source %q", source))
	}

	var seq int
	for attempt := 0; attempt < maxTakeRetries; attempt++ {
		row, err := tx.OrderCounter.Query().Where(
			entoc.CompanyIDEQ(companyID), entoc.SourceEQ(source),
		).Only(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return "", connect.NewError(connect.CodeInternal, err)
			}
			// 首列:插入撞唯一鍵 → 轉重新查詢(細部 4.1.2 步驟 1)
			row, err = tx.OrderCounter.Create().
				SetCompanyID(companyID).SetSource(source).SetNextSeq(1).SetVersion(1).Save(ctx)
			if err != nil {
				continue
			}
		}
		seq = row.NextSeq
		// 條件更新:version 相符才更新(細部 4.1.2 步驟 2)
		affected, err := tx.OrderCounter.Update().
			Where(entoc.IDEQ(row.ID), entoc.VersionEQ(row.Version)).
			SetNextSeq(row.NextSeq + 1).SetVersion(row.Version + 1).
			Save(ctx)
		if err == nil && affected == 1 {
			return fmt.Sprintf("%s%06d", source, seq), nil
		}
		time.Sleep(time.Duration(attempt+1) * 5 * time.Millisecond) // 量級低,線性退避即可
	}
	return "", connect.NewError(connect.CodeInternal, errors.New("order number contention exceeded retries"))
}
```

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go generate ./ent/ && go test ./internal/domain/salesorders/ -v -race`
Expected: PASS — 連續、雙軌、併發 20 無重複、回滾不消耗。

```bash
git add backend/ent/schema/ordercounter.go backend/internal/domain/salesorders/numbering.go
git commit -m "feat(backend): order_counters 樂觀鎖取號,衝突自動重試(4.1.2)"
```

---

### Task 3: 訂單狀態機(細部 4.1.3,D13)

**Files:**
- Create: `backend/internal/domain/salesorders/statemachine.go`
- Test: `backend/internal/domain/salesorders/statemachine_test.go`

**Interfaces:**
- Consumes: Task 1 schema;03 計畫 `audit.Recorder`。
- Produces(供 Task 4 與 08-dispatch 計畫消費):

```go
type Status = string // "pending" | "processing" | "completed" | "cancelled" | "voided"

// CanTransition 查合法轉移表。
func CanTransition(from, to Status) bool

// Transition 唯一狀態入口;交易內:條件更新(status=前值)+ 事件 + version+1;
// dispatch 寫 dispatched_at/by;dispatch_cancel 清 dispatched_* 保留 route/seq + 稽核;
// void 需 reason + 稽核;角色檢查由呼叫方 Casbin 完成,本函式僅驗狀態與原因。
func Transition(ctx context.Context, tx *ent.Tx, rec audit.Recorder,
	orderID uuid.UUID, from, to Status, actorID uuid.UUID, reason string) error
```

- [ ] **Step 1: 寫失敗測試(合法全通、非法全拒、事件與稽核、回退保留看板位)**

```go
// backend/internal/domain/salesorders/statemachine_test.go
package salesorders_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/domain/salesorders"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestTransitionTable(t *testing.T) {
	legal := [][2]string{
		{"pending", "processing"}, {"processing", "pending"},
		{"processing", "completed"}, {"pending", "cancelled"}, {"completed", "voided"},
	}
	for _, tr := range legal {
		if !salesorders.CanTransition(tr[0], tr[1]) {
			t.Errorf("should be legal: %s -> %s", tr[0], tr[1])
		}
	}
	illegal := [][2]string{
		{"pending", "completed"}, {"pending", "voided"}, {"processing", "cancelled"},
		{"cancelled", "pending"}, {"voided", "pending"}, {"completed", "pending"},
		{"completed", "processing"}, {"cancelled", "cancelled"},
	}
	for _, tr := range illegal {
		if salesorders.CanTransition(tr[0], tr[1]) {
			t.Errorf("should be illegal: %s -> %s", tr[0], tr[1])
		}
	}
}

func TestTransitionWritesEventAndBumpsVersion(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	order := seedOrder(t, c, "pending") // 輔助:建 pending 訂單回 *ent.SalesOrder
	actor := uuid.New()

	tx, _ := c.Tx(ctx)
	err := salesorders.Transition(ctx, tx, nil, order.ID, "pending", "cancelled", actor, "")
	if err != nil { t.Fatalf("cancel: %v", err) }
	if err := tx.Commit(); err != nil { t.Fatal(err) }

	got := c.SalesOrder.GetX(ctx, order.ID)
	if got.Status != "cancelled" || got.Version != order.Version+1 {
		t.Fatalf("after cancel: status=%s version=%d", got.Status, got.Version)
	}
	ev, err := c.SalesOrderEvent.Query().Where().All(ctx) // sales_order_id = order.ID
	if err != nil || len(ev) != 1 || ev[0].EventType != "cancel" || ev[0].ActorID != actor {
		t.Fatalf("event: %v %+v", err, ev)
	}
	// 終態再轉移 → failed_precondition
	tx, _ = c.Tx(ctx)
	err = salesorders.Transition(ctx, tx, nil, order.ID, "cancelled", "pending", actor, "")
	_ = tx.Rollback()
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("terminal transition: want failed_precondition, got %v", err)
	}
}

func TestDispatchCancelClearsDispatchKeepsBoardPosition(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	order := seedOrder(t, c, "processing")
	// seed 時已設 dispatched_at/by 與 route_id/delivery_sequence
	tx, _ := c.Tx(ctx)
	err := salesorders.Transition(ctx, tx, nil, order.ID, "processing", "pending", uuid.New(), "車輛調度")
	if err != nil { t.Fatal(err) }
	if err := tx.Commit(); err != nil { t.Fatal(err) }

	got := c.SalesOrder.GetX(ctx, order.ID)
	if got.DispatchedAt != nil || got.DispatchedBy != nil {
		t.Fatal("dispatch fields should be cleared")
	}
	if got.RouteID == nil || got.DeliverySequence == nil {
		t.Fatal("board position must be preserved (細部 4.1.3)")
	}
	ev := c.SalesOrderEvent.Query().Where().OnlyX(ctx)
	if ev.EventType != "dispatch_cancel" || ev.Reason == nil || *ev.Reason != "車輛調度" {
		t.Fatalf("dispatch_cancel event: %+v", ev)
	}
}

func TestVoidRequiresReason(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	order := seedOrder(t, c, "completed")
	tx, _ := c.Tx(ctx)
	err := salesorders.Transition(ctx, tx, nil, order.ID, "completed", "voided", uuid.New(), "")
	_ = tx.Rollback()
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("void without reason: want invalid_argument, got %v", err)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/salesorders/ -run 'TestTransition|TestDispatchCancel|TestVoid' -v`
Expected: FAIL — `CanTransition` / `Transition` 未定義。

- [ ] **Step 3: 實作**

`statemachine.go`:

```go
package salesorders

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entso "github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
)

type Status = string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
	StatusVoided     Status = "voided"
)

// transitions 合法轉移表(細部 4.1.3 步驟 1);鍵外組合一律非法。
var transitions = map[Status][]Status{
	StatusPending:    {StatusProcessing, StatusCancelled},
	StatusProcessing: {StatusPending, StatusCompleted},
	StatusCompleted:  {StatusVoided},
	StatusCancelled:  {},
	StatusVoided:     {},
}

// eventTypeOf 轉移 → 事件型別。
func eventTypeOf(from, to Status) string {
	switch {
	case from == StatusPending && to == StatusProcessing:
		return "dispatch"
	case from == StatusProcessing && to == StatusPending:
		return "dispatch_cancel"
	case from == StatusProcessing && to == StatusCompleted:
		return "complete"
	case from == StatusPending && to == StatusCancelled:
		return "cancel"
	case from == StatusCompleted && to == StatusVoided:
		return "void"
	}
	return "edit"
}

func CanTransition(from, to Status) bool {
	for _, t := range transitions[from] {
		if t == to {
			return true
		}
	}
	return false
}

// Transition:狀態機唯一入口(細部 4.1.3)。條件更新確保併發安全;
// dispatch_cancel / void 必填原因且同交易寫稽核(D13/D18);rec 可為 nil(單元測試)。
func Transition(ctx context.Context, tx *ent.Tx, rec audit.Recorder,
	orderID uuid.UUID, from, to Status, actorID uuid.UUID, reason string) error {

	if !CanTransition(from, to) {
		return connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("cannot transition %s -> %s", from, to))
	}
	et := eventTypeOf(from, to)
	if (et == "dispatch_cancel" || et == "void") && reason == "" {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("reason required"))
	}

	upd := tx.SalesOrder.Update().
		Where(entso.IDEQ(orderID), entso.StatusEQ(from)). // 條件更新(步驟 6)
		SetStatus(to)
	switch et {
	case "dispatch":
		upd.SetDispatchedAt(time.Now()).SetDispatchedBy(actorID)
	case "dispatch_cancel":
		upd.ClearDispatchedAt().ClearDispatchedBy() // 保留 route_id / delivery_sequence
	}
	affected, err := upd.Save(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if affected != 1 {
		return connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("order not in %s (concurrent change)", from))
	}
	if _, err := tx.SalesOrder.Update().Where(entso.IDEQ(orderID)).
		SetVersion(intVersionPlusOne(ctx, tx, orderID)).Save(ctx); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}

	ev := tx.SalesOrderEvent.Create().
		SetSalesOrderID(orderID).
		SetEventType(et).SetActorID(actorID).
		SetPayload(map[string]any{"from": from, "to": to})
	if reason != "" {
		ev.SetReason(reason)
	}
	// company_id 冗餘自訂單
	order, err := tx.SalesOrder.Get(ctx, orderID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	ev.SetCompanyID(order.CompanyID)
	if _, err := ev.Save(ctx); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}

	if (et == "dispatch_cancel" || et == "void") && rec != nil {
		if err := rec.Record(audit.ContextWithTx(ctx, tx), audit.Entry{
			ActorID: actorID.String(), Action: et,
			ResourceType: "sales_order", ResourceID: order.OrderNo,
			Before: fmt.Sprintf(`{"status":%q}`, from),
			After:  fmt.Sprintf(`{"status":%q,"reason":%q}`, to, reason),
		}); err != nil {
			return connect.NewError(connect.CodeInternal, err) // 呼叫方回滾
		}
	}
	return nil
}
```

(`intVersionPlusOne`:讀現值 +1;或改用 `AddVersion(1)` Ent 增量語法 — 實作採 `Add(1)` 原子遞增。)

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go test ./internal/domain/salesorders/ -v`
Expected: PASS。

```bash
git add backend/internal/domain/salesorders/statemachine.go
git commit -m "feat(backend): 訂單狀態機,事件+稽核同事務(4.1.3)"
```

---

### Task 4: SalesOrderService CRUD 與 ListEvents(細部 4.1.4–4.1.5)

**Files:**
- Create: `backend/proto/v1/salesorder.proto`
- Create: `backend/internal/domain/salesorders/{handler,usecase,repo}.go`
- Test: `backend/internal/domain/salesorders/usecase_test.go`

**Interfaces:**
- Consumes: Task 2 取號、Task 3 狀態機、03 計畫稽核、04 計畫 customers/products。
- Produces: Connect-RPC `SalesOrderService`:`ListOrders` / `GetOrder` / `CreateOrder` / `UpdateOrder` / `CancelOrder` / `CompleteOrder` / `VoidOrder` / `DeleteOrder`(軟刪除,僅 pending/cancelled)/ `ListOrderEvents`;usecase `salesorders.Usecase`;Create 的明細組裝委派 Task 5 的 `ordering.go`(本 Task 以最小品項驗證先行:非空、qty>0;換算/別名/順延於 Task 5 掛入)。

- [ ] **Step 1: 寫失敗測試(Create 同事務、Update version 鎖、軟刪除限制、事件軌跡)**

```go
// backend/internal/domain/salesorders/usecase_test.go
package salesorders_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/salesorders"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestCreateOrderHappyPath(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, depID, custID, productID := seedOrderFixtures(t, c) // 輔助:公司+部門+客戶+商品(含 units)+ order_source seed
	staff := rls.Identity{UserID: uuid.New(), CompanyID: coID, DepartmentID: &depID, DataScope: "department", Role: "staff"}

	uc := salesorders.NewUsecase(c, nil, nil) // rec, notifier Task 5 接
	order, err := uc.CreateOrder(ctx, staff, salesorders.CreateOrderInput{
		CustomerID: custID,
		Source:     "W",
		ExpectedDeliveryDate: "2026-08-20",
		Items: []salesorders.OrderItemInput{{
			ProductID: &productID, DisplayName: "去骨雞腿",
			Qty: decimal.NewFromInt(5), Unit: "KG",
		}},
	})
	if err != nil { t.Fatalf("create: %v", err) }
	if order.OrderNo != "W000001" || order.Status != "pending" || order.Version != 1 {
		t.Fatalf("order: %+v", order)
	}
	// 明細與 create 事件同交易落庫
	items := c.SalesOrderItem.Query().Where().AllX(ctx)
	if len(items) != 1 || !items[0].BaseQty.Equal(decimal.NewFromInt(5)) {
		t.Fatalf("items: %+v", items)
	}
	ev := c.SalesOrderEvent.Query().Where().OnlyX(ctx)
	if ev.EventType != "create" {
		t.Fatalf("event: %+v", ev)
	}
}

func TestCreateOrderRollbackOnInvalidItem(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, depID, custID, productID := seedOrderFixtures(t, c)
	staff := rls.Identity{UserID: uuid.New(), CompanyID: coID, DepartmentID: &depID, DataScope: "department", Role: "staff"}
	uc := salesorders.NewUsecase(c, nil, nil)

	_, err := uc.CreateOrder(ctx, staff, salesorders.CreateOrderInput{
		CustomerID: custID, Source: "W", ExpectedDeliveryDate: "2026-08-20",
		Items: []salesorders.OrderItemInput{
			{ProductID: &productID, DisplayName: "好", Qty: decimal.NewFromInt(1), Unit: "KG"},
			{ProductID: &productID, DisplayName: "壞", Qty: decimal.NewFromInt(1), Unit: "NOPE"}, // 無效單位
		},
	})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("want invalid_argument, got %v", err)
	}
	// 整單回滾:無訂單、無序號消耗
	if n := c.SalesOrder.Query().CountX(ctx); n != 0 {
		t.Fatalf("order should not exist, got %d", n)
	}
	tx, _ := c.Tx(ctx)
	no, _ := salesorders.NextOrderNo(ctx, tx, coID, "W")
	_ = tx.Commit()
	if no != "W000001" {
		t.Fatalf("seq should not be consumed, got %s", no)
	}
}

func TestUpdateOrderVersionConflict(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, depID, custID, productID := seedOrderFixtures(t, c)
	staff := rls.Identity{UserID: uuid.New(), CompanyID: coID, DepartmentID: &depID, DataScope: "department", Role: "staff"}
	uc := salesorders.NewUsecase(c, nil, nil)
	order, _ := uc.CreateOrder(ctx, staff, salesorders.CreateOrderInput{
		CustomerID: custID, Source: "W", ExpectedDeliveryDate: "2026-08-20",
		Items: []salesorders.OrderItemInput{{ProductID: &productID, DisplayName: "雞", Qty: decimal.NewFromInt(1), Unit: "KG"}},
	})

	// version 不符 → failed_precondition
	err := uc.UpdateOrder(ctx, staff, order.ID, salesorders.UpdateOrderInput{
		Version: order.Version + 99, Note: ptr("改"),
	})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("stale version: want failed_precondition, got %v", err)
	}
	// 正確 version → 成功,version+1,edit 事件
	err = uc.UpdateOrder(ctx, staff, order.ID, salesorders.UpdateOrderInput{
		Version: order.Version, Note: ptr("改"),
	})
	if err != nil { t.Fatalf("update: %v", err) }
	if got := c.SalesOrder.GetX(ctx, order.ID); got.Version != order.Version+1 || got.Note != "改" {
		t.Fatalf("after update: %+v", got)
	}
}

func TestDeleteRestrictions(t *testing.T) {
	// pending → 軟刪除可;processing → failed_precondition;軟刪除後 List 不見、Get not_found
}

func TestListEventsOrdered(t *testing.T) {
	// create → edit → cancel 三事件依 created_at 升序;客戶帳號查他人訂單 → not_found
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/salesorders/ -run 'TestCreateOrder|TestUpdateOrder|TestDelete|TestListEvents' -v`
Expected: FAIL — `Usecase` 未定義。

- [ ] **Step 3: 實作 proto**

`backend/proto/v1/salesorder.proto`(訊息形狀依細部 4.1.4 介面逐字;重點):

```proto
syntax = "proto3";
package salesorder.v1;

service SalesOrderService {
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
  rpc UpdateOrder(UpdateOrderRequest) returns (UpdateOrderResponse);
  rpc CancelOrder(CancelOrderRequest) returns (CancelOrderResponse);
  rpc CompleteOrder(CompleteOrderRequest) returns (CompleteOrderResponse);
  rpc VoidOrder(VoidOrderRequest) returns (VoidOrderResponse);
  rpc DeleteOrder(DeleteOrderRequest) returns (DeleteOrderResponse);
  rpc ListOrderEvents(ListOrderEventsRequest) returns (ListOrderEventsResponse);
}

message OrderItemInput {
  optional string product_id = 1;   // 與 manual_name 擇一
  optional string manual_name = 2;  // 純手打(業務限定)
  string display_name = 3;
  string qty = 4;                   // decimal 字串(proto3 無 decimal)
  string unit = 5;
  optional string cutting_spec_id = 6;
  optional string special_cut_note = 7;
  optional string warehouse_id = 8;
  bool save_alias = 9;              // Task 5
}
message SalesOrder {
  string id = 1; string order_no = 2; string customer_id = 3; string source = 4;
  string status = 5; string expected_delivery_date = 6; // date
  optional string sales_rep_id = 7; string note = 8;
  optional string route_id = 9; optional int32 delivery_sequence = 10;
  int32 version = 11; string created_at = 12;
}
// CreateOrderRequest{customer_id, source, expected_delivery_date, note?, sales_rep_id?, items[]}
// UpdateOrderRequest{id, version, expected_delivery_date?, note?, items[]?}
// VoidOrderRequest{id, reason(必填)};CancelOrderRequest{id, reason?}
// ListOrderEventsResponse{events[]{event_type, actor_id, actor_name, reason?, payload?, created_at}}
```

- [ ] **Step 4: 實作 usecase**

`usecase.go`(骨架,明細組裝呼叫 Task 5 的 `assembleItems`;本 Task 先以基本驗證版):

```go
package salesorders

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entso "github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/notification"
	"github.com/salesorder/sales-order-1.0/backend/internal/notification/triggers"
)

type Usecase struct {
	db       *ent.Client
	rec      audit.Recorder
	sender notification.Sender // 07 計畫 T4 `Sender` 介面;可 nil(通知關閉,測試注入 FakeSender)
}

func NewUsecase(db *ent.Client, rec audit.Recorder, sender notification.Sender) *Usecase {
	return &Usecase{db: db, rec: rec, sender: sender}
}

type OrderItemInput struct {
	ProductID      *uuid.UUID
	ManualName     *string
	DisplayName    string
	Qty            decimal.Decimal
	Unit           string
	CuttingSpecID  *uuid.UUID
	SpecialCutNote *string
	WarehouseID    *uuid.UUID
	SaveAlias      bool
}

type CreateOrderInput struct {
	CustomerID           uuid.UUID
	Source               string
	ExpectedDeliveryDate string // YYYY-MM-DD
	Note                 string
	SalesRepID           *uuid.UUID
	Items                []OrderItemInput
}

type UpdateOrderInput struct {
	Version              int
	ExpectedDeliveryDate *string
	Note                 *string
	Items                []OrderItemInput // 非空時整單替換
}

// CreateOrder 全程單一交易(細部 4.1.4 步驟 1):
// 守衛(4.2.3)→ 明細組裝(4.2.1/4.2.2/4.2.5)→ 取號(4.1.2)→ 建單+明細 → create 事件 → commit。
// 提交後(交易外)通知(4.4.3 路由,07 計畫)。
func (u *Usecase) CreateOrder(ctx context.Context, actor rls.Identity, in CreateOrderInput) (*ent.SalesOrder, error) {
	if err := u.guardOrderingScope(ctx, actor, &in); err != nil { // Task 5;本 Task 先驗 customer 存在可見
		return nil, err
	}
	if len(in.Items) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("items required"))
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	items, err := assembleItems(ctx, tx, actor, in.CustomerID, in.Items) // Task 5
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	deliveryDate, err := adjustDeliveryDate(ctx, tx, in.CustomerID, in.ExpectedDeliveryDate) // Task 5
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	orderNo, err := NextOrderNo(ctx, tx, actor.CompanyID, in.Source)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	ob := tx.SalesOrder.Create().
		SetCompanyID(actor.CompanyID).
		SetCustomerID(in.CustomerID).
		SetSource(in.Source).
		SetOrderNo(orderNo).
		SetExpectedDeliveryDate(deliveryDate).
		SetCreatedBy(actor.UserID).
		SetNote(in.Note)
	if actor.DepartmentID != nil {
		ob.SetDepartmentID(*actor.DepartmentID)
	} else {
		// company/all 範圍下單:部門取自客戶歸屬
		ob.SetDepartmentID(lookupCustomerDepartment(ctx, tx, in.CustomerID))
	}
	if in.SalesRepID != nil {
		ob.SetSalesRepID(*in.SalesRepID)
	} else {
		ob.SetSalesRepID(actor.UserID) // 預設操作者本人(細部 4.2.3 步驟 4)
	}
	order, err := ob.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	for i, it := range items {
		if _, err := tx.SalesOrderItem.Create().
			SetSalesOrderID(order.ID).
			SetCompanyID(order.CompanyID).SetDepartmentID(order.DepartmentID).
			SetDisplayName(it.DisplayName).
			SetQty(it.Qty).SetUnit(it.Unit).SetBaseQty(it.BaseQty).
			SetSortOrder(i).
			SetNillableProductID(it.ProductID).
			SetNillableCuttingSpecID(it.CuttingSpecID).
			SetNillableSpecialCutNote(it.SpecialCutNote).
			SetNillableWarehouseID(it.WarehouseID).
			Save(ctx); err != nil {
			_ = tx.Rollback()
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	if _, err := tx.SalesOrderEvent.Create().
		SetSalesOrderID(order.ID).SetCompanyID(order.CompanyID).
		SetEventType("create").SetActorID(actor.UserID).
		SetPayload(map[string]any{"order_no": orderNo, "source": in.Source}).
		Save(ctx); err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 通知掛點:Task 5 於本函式內插入 — 交易內 triggers.OnOrderCreated 建檔(在 Commit 前),
	// 提交後 notification.Dispatch 發送(07 計畫契約);Task 4 先留 nil sender 不發通知。
	_ = u.sender
	return order, nil
}

// UpdateOrder:僅 pending;version 比對;整單替換明細(軟刪舊+插新,重跑組裝與順延)。
func (u *Usecase) UpdateOrder(ctx context.Context, actor rls.Identity, id uuid.UUID, in UpdateOrderInput) error {
	order, err := u.getVisible(ctx, actor, id)
	if err != nil {
		return err
	}
	if order.Status != StatusPending {
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("only pending orders can be edited; cancel dispatch first if processing"))
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	// 樂觀鎖:version 相符才更新
	affected, err := tx.SalesOrder.Update().
		Where(entso.IDEQ(id), entso.VersionEQ(in.Version)).
		SetVersion(in.Version + 1).
		SetNillableNote(in.Note).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	if affected != 1 {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("version conflict; refetch"))
	}
	if in.ExpectedDeliveryDate != nil {
		d, err := adjustDeliveryDate(ctx, tx, order.CustomerID, *in.ExpectedDeliveryDate)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if _, err := tx.SalesOrder.UpdateOneID(id).SetExpectedDeliveryDate(d).Save(ctx); err != nil {
			_ = tx.Rollback()
			return connect.NewError(connect.CodeInternal, err)
		}
	}
	if in.Items != nil {
		items, err := assembleItems(ctx, tx, actor, order.CustomerID, in.Items)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		// 軟刪舊明細 + 插新(同交易)
		if _, err := tx.SalesOrderItem.Update().
			Where(entsoitems.SalesOrderIDEQ(id), entsoitems.DeletedAtIsNil()).
			SetDeletedAt(time.Now()).Save(ctx); err != nil {
			_ = tx.Rollback()
			return connect.NewError(connect.CodeInternal, err)
		}
		for i, it := range items {
			// 插入同 CreateOrder 迴圈(抽共用 insertItems(ctx, tx, order, items))
			_ = i; _ = it
		}
	}
	if _, err := tx.SalesOrderEvent.Create().
		SetSalesOrderID(id).SetCompanyID(order.CompanyID).
		SetEventType("edit").SetActorID(actor.UserID).
		SetPayload(map[string]any{"summary": "order edited"}).
		Save(ctx); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	return tx.Commit()
}

// Cancel / Complete / Void:委派狀態機;Void 角色檢查(dept_admin 以上)由 Casbin 於 handler 層完成。
func (u *Usecase) CancelOrder(ctx context.Context, actor rls.Identity, id uuid.UUID, reason string) error {
	return u.transition(ctx, actor, id, StatusPending, StatusCancelled, reason)
}
func (u *Usecase) CompleteOrder(ctx context.Context, actor rls.Identity, id uuid.UUID) error {
	return u.transition(ctx, actor, id, StatusProcessing, StatusCompleted, "")
}
func (u *Usecase) VoidOrder(ctx context.Context, actor rls.Identity, id uuid.UUID, reason string) error {
	return u.transition(ctx, actor, id, StatusCompleted, StatusVoided, reason)
}

func (u *Usecase) transition(ctx context.Context, actor rls.Identity, id uuid.UUID, from, to Status, reason string) error {
	if _, err := u.getVisible(ctx, actor, id); err != nil {
		return err
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if err := Transition(ctx, tx, u.rec, id, from, to, actor.UserID, reason); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// DeleteOrder:軟刪除,僅 pending/cancelled;明細連帶;稽核同事務(細部 4.1.4 步驟 6)。
func (u *Usecase) DeleteOrder(ctx context.Context, actor rls.Identity, id uuid.UUID) error {
	order, err := u.getVisible(ctx, actor, id)
	if err != nil {
		return err
	}
	if order.Status != StatusPending && order.Status != StatusCancelled {
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("only pending/cancelled orders can be deleted"))
	}
	// tx:軟刪訂單+明細 + 稽核(action=delete, resource=sales_order)
	// (實作比照前述模式)
	return nil
}
```

`ListOrderEvents`:先 `getVisible`(不可見 → `not_found` 不洩漏存在性)→ 查 events 升序 + join users 取操作者姓名。`getVisible`:查單 + 未軟刪除 + actor 範圍(self 僅 `customer_id = actor.CustomerID`)。

- [ ] **Step 5: 跑測試確認通過**

Run: `cd backend && go generate ./ent/ && go test ./internal/domain/salesorders/ -v`
Expected: PASS(`assembleItems` / `adjustDeliveryDate` 本 Task 先以最小版落地:基本驗證 + 原樣日期;Task 5 補完整邏輯與測試)。

- [ ] **Step 6: Commit**

```bash
git add backend/proto/v1/salesorder.proto backend/internal/domain/salesorders
git commit -m "feat(backend): SalesOrderService CRUD、version 鎖、事件軌跡(4.1.4-4.1.5)"
```

---

### Task 5: 下單組裝邏輯(細部 4.2.1–4.2.5)

**Files:**
- Create: `backend/internal/domain/salesorders/ordering.go`
- Update: `backend/internal/domain/salesorders/usecase.go`(通知掛接段)
- Test: `backend/internal/domain/salesorders/ordering_test.go`

**Interfaces:**
- Consumes: 04 計畫 `products.ToBase`(decimal 契約)、`customerproducts.EnsureAlias(ctx, tx, customerID, productID, aliasName, defaultQty)`(套件級函式);03 計畫 metadicts order_source;07 計畫 `triggers.OnOrderCreated` / `notification.Dispatch` / `notification.Sender` / `notification.FakeSender`。
- Produces:
  - `assembleItems(ctx, tx, actor, customerID, items) ([]assembledItem, error)`(換算 + 手打/別名 + 清單守衛)
  - `adjustDeliveryDate(ctx, tx, customerID, date string) (time.Time, error)`(D26 順延)
  - `guardOrderingScope(ctx, actor, in *CreateOrderInput) error`(客戶/業務守衛)
  - CreateOrder 通知掛接(交易內建檔 + 提交後 Dispatch;07 計畫函式直呼,不自定介面)

- [ ] **Step 1: 寫失敗測試(換算、清單守衛、手打、順延四情境)**

```go
// backend/internal/domain/salesorders/ordering_test.go
package salesorders_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/salesorders"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestUnitConversionApplied(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, depID, custID, productID := seedOrderFixtures(t, c) // 商品含 PIECE rate 0.6 KG
	staff := rls.Identity{UserID: uuid.New(), CompanyID: coID, DepartmentID: &depID, DataScope: "department", Role: "staff"}
	uc := salesorders.NewUsecase(c, nil, nil)

	order, err := uc.CreateOrder(ctx, staff, salesorders.CreateOrderInput{
		CustomerID: custID, Source: "W", ExpectedDeliveryDate: "2026-08-19", // 週三(客戶無偏好 → 維持)
		Items: []salesorders.OrderItemInput{{
			ProductID: &productID, DisplayName: "雞腿",
			Qty: decimal.NewFromInt(10), Unit: "PIECE",
		}},
	})
	if err != nil { t.Fatal(err) }
	item := c.SalesOrderItem.Query().Where().OnlyX(ctx)
	if !item.BaseQty.Equal(decimal.NewFromInt(6)) { // 10 條 × 0.6 = 6 kg
		t.Fatalf("base_qty: want 6, got %s", item.BaseQty)
	}
	_ = order
}

func TestCustomerRestrictedToOwnProducts(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, depID, custID, productID := seedOrderFixtures(t, c)
	cust := rls.Identity{UserID: uuid.New(), CompanyID: coID, DepartmentID: &depID,
		DataScope: "self", Role: "customer", CustomerID: &custID}
	uc := salesorders.NewUsecase(c, nil, nil)

	// 商品不在客戶專屬清單 → invalid_argument 整單拒絕
	_, err := uc.CreateOrder(ctx, cust, salesorders.CreateOrderInput{
		CustomerID: custID, Source: "A", ExpectedDeliveryDate: "2026-08-19",
		Items: []salesorders.OrderItemInput{{
			ProductID: &productID, DisplayName: "雞腿", Qty: decimal.NewFromInt(1), Unit: "KG",
		}},
	})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("off-list item: want invalid_argument, got %v", err)
	}
	// 客戶手打 → permission_denied
	_, err = uc.CreateOrder(ctx, cust, salesorders.CreateOrderInput{
		CustomerID: custID, Source: "A", ExpectedDeliveryDate: "2026-08-19",
		Items: []salesorders.OrderItemInput{{
			ManualName: ptr("手打品"), DisplayName: "手打品", Qty: decimal.NewFromInt(1), Unit: "KG",
		}},
	})
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("customer manual item: want permission_denied, got %v", err)
	}
	// 客戶帶他人 customer_id → 拒絕(細部 4.2.3 驗收:拒絕以避免誤單)
	otherCust := uuid.New()
	_, err = uc.CreateOrder(ctx, cust, salesorders.CreateOrderInput{
		CustomerID: otherCust, Source: "A", ExpectedDeliveryDate: "2026-08-19",
		Items: []salesorders.OrderItemInput{{
			ProductID: &productID, DisplayName: "雞腿", Qty: decimal.NewFromInt(1), Unit: "KG",
		}},
	})
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("foreign customer_id: want permission_denied, got %v", err)
	}
}

func TestSaveAliasCreatesCustomerProduct(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, depID, custID, productID := seedOrderFixtures(t, c)
	staff := rls.Identity{UserID: uuid.New(), CompanyID: coID, DepartmentID: &depID, DataScope: "department", Role: "staff"}
	uc := salesorders.NewUsecase(c, nil, nil)

	_, err := uc.CreateOrder(ctx, staff, salesorders.CreateOrderInput{
		CustomerID: custID, Source: "W", ExpectedDeliveryDate: "2026-08-19",
		Items: []salesorders.OrderItemInput{{
			ProductID: &productID, DisplayName: "去骨雞腿(客稱)",
			Qty: decimal.NewFromInt(2), Unit: "KG", SaveAlias: true,
		}},
	})
	if err != nil { t.Fatal(err) }
	// customer_products 出現別名;同商品再次下單 save_alias → 更新既有,不重複建列
	cp := c.CustomerProduct.Query().Where().OnlyX(ctx)
	if cp.AliasName != "去骨雞腿(客稱)" {
		t.Fatalf("alias: %+v", cp)
	}
	_, err = uc.CreateOrder(ctx, staff, salesorders.CreateOrderInput{
		CustomerID: custID, Source: "W", ExpectedDeliveryDate: "2026-08-20",
		Items: []salesorders.OrderItemInput{{
			ProductID: &productID, DisplayName: "去骨雞腿(新稱)",
			Qty: decimal.NewFromInt(3), Unit: "KG", SaveAlias: true,
		}},
	})
	if err != nil { t.Fatal(err) }
	if n := c.CustomerProduct.Query().Where().CountX(ctx); n != 1 {
		t.Fatalf("one customer-product row expected, got %d", n)
	}
	if got := c.CustomerProduct.Query().Where().OnlyX(ctx); got.AliasName != "去骨雞腿(新稱)" {
		t.Fatalf("alias should be updated: %s", got.AliasName)
	}
}

func TestPreferredDeliveryDayRollover(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, depID, custID, productID := seedOrderFixtures(t, c)
	// 客戶偏好:週二、週四(preferred_delivery_days = [false,true,false,true,false,false])
	c.Customer.UpdateOneID(custID).
		SetPreferredDeliveryDays([]int{2, 4}).ExecX(ctx)
	staff := rls.Identity{UserID: uuid.New(), CompanyID: coID, DepartmentID: &depID, DataScope: "department", Role: "staff"}
	uc := salesorders.NewUsecase(c, nil, nil)

	// 2026-08-19 是週三 → 順延至週四 2026-08-20
	order, err := uc.CreateOrder(ctx, staff, salesorders.CreateOrderInput{
		CustomerID: custID, Source: "W", ExpectedDeliveryDate: "2026-08-19",
		Items: []salesorders.OrderItemInput{{
			ProductID: &productID, DisplayName: "雞腿", Qty: decimal.NewFromInt(1), Unit: "KG",
		}},
	})
	if err != nil { t.Fatal(err) }
	if got := order.ExpectedDeliveryDate.Format("2006-01-02"); got != "2026-08-20" {
		t.Fatalf("rollover: want 2026-08-20 (Thu), got %s", got)
	}

	// 跨週:偏好僅週一,選週五 2026-08-21 → 順延至 2026-08-24(下週一)
	c.Customer.UpdateOneID(custID).SetPreferredDeliveryDays([]int{1}).ExecX(ctx)
	order2, err := uc.CreateOrder(ctx, staff, salesorders.CreateOrderInput{
		CustomerID: custID, Source: "W", ExpectedDeliveryDate: "2026-08-21",
		Items: []salesorders.OrderItemInput{{
			ProductID: &productID, DisplayName: "雞腿", Qty: decimal.NewFromInt(1), Unit: "KG",
		}},
	})
	if err != nil { t.Fatal(err) }
	if got := order2.ExpectedDeliveryDate.Format("2006-01-02"); got != "2026-08-24" {
		t.Fatalf("cross-week rollover: want 2026-08-24 (Mon), got %s", got)
	}
	_ = time.Now
}
```

注意:`seedOrderFixtures` 的客戶 `preferred_delivery_days` 欄位 — 04 計畫 schema 為 `[]int`(星期編號 1-6),細部文件 4.2.5 介面描述為 bool 陣列;**統一採 `[]int`(勾選的星期編號)**,04 計畫已落地此型別,`adjustDeliveryDate` 依 `[]int` 實作(Self-Review 記錄此決策)。

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/salesorders/ -run 'TestUnitConversion|TestCustomerRestricted|TestSaveAlias|TestPreferredDelivery' -v`
Expected: FAIL — 最小版 `assembleItems` 不做換算/守衛/別名,`adjustDeliveryDate` 未順延。

- [ ] **Step 3: CreateOrder 通知掛接(07 計畫契約)**

不新增介面/套件;直接消費 07 計畫產物。修改 Task 4 `CreateOrder` 的通知掛點段為:

```go
	// --- 通知(D18):同一交易內建 pending 記錄(07 計畫 T5),提交後 Dispatch(07 計畫 T4)---
	var notifyIDs []uuid.UUID
	if u.sender != nil {
		ids, err := triggers.OnOrderCreated(ctx, tx, triggers.OrderInfo{
			ID: order.ID, OrderNo: order.OrderNo,
			CompanyID: order.CompanyID, DepartmentID: &order.DepartmentID,
			CustomerID: in.CustomerID, CustomerName: customerName(ctx, tx, in.CustomerID),
			ItemCount: len(items),
		}, triggers.ActorInfo{
			UserID: actor.UserID, Role: actor.Role, IsCustomer: actor.Role == "customer",
		})
		if err != nil {
			_ = tx.Rollback() // 建檔失敗同交易回滾(D18)
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		notifyIDs = ids
	}
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 提交後發送;失敗僅標 failed,不回滾(D16)
	if u.sender != nil && len(notifyIDs) > 0 {
		if err := notification.Dispatch(ctx, u.db, u.sender, u.rec, notifyIDs); err != nil {
			log.Printf("salesorders: dispatch notifications: %v", err)
		}
	}
	return order, nil
}
```

測試注入 `notification.FakeSender`(07 計畫 T4 提供);`TestCreateOrderHappyPath` 加斷言:業務下單後 `notifications` 表該客戶每子帳號 fcm/in_app 各一筆;客戶自行下單 → 0 筆。`customerName` 為簡單查詢輔助函式。

- [ ] **Step 4: 實作 ordering.go**

```go
package salesorders

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entcp "github.com/salesorder/sales-order-1.0/backend/ent/customerproduct"
	entcust "github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/customerproducts"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/products"
)

// bizLocation 營業時區(D26 步驟 6:固定 UTC+8)。
var bizLocation = time.FixedZone("UTC+8", 8*3600)

// guardOrderingScope(細部 4.2.3):客戶守衛 + 員工範圍。
func (u *Usecase) guardOrderingScope(ctx context.Context, actor rls.Identity, in *CreateOrderInput) error {
	if actor.Role == "customer" {
		if actor.IsPrimary {
			return connect.NewError(connect.CodePermissionDenied,
				errors.New("primary account is management-only"))
		}
		if actor.CustomerID == nil {
			return connect.NewError(connect.CodeFailedPrecondition, errors.New("customer context missing"))
		}
		if in.CustomerID != *actor.CustomerID {
			return connect.NewError(connect.CodePermissionDenied,
				errors.New("cannot order for another customer"))
		}
		for _, it := range in.Items {
			if it.ManualName != nil {
				return connect.NewError(connect.CodePermissionDenied,
					errors.New("manual items are staff-only"))
			}
		}
		return nil
	}
	// 員工:customer 必須存在、未軟刪除,且在其 data_scope(department 由 RLS/where 雙層)
	cust, err := u.db.Customer.Query().Where(
		entcust.IDEQ(in.CustomerID), entcust.DeletedAtIsNil(), entcust.StatusEQ("active"),
	).Only(ctx)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, errors.New("customer not found"))
	}
	if actor.DataScope == "department" && actor.DepartmentID != nil && cust.DepartmentID != *actor.DepartmentID {
		return connect.NewError(connect.CodePermissionDenied, errors.New("customer out of department scope"))
	}
	return nil
}

type assembledItem struct {
	ProductID      *uuid.UUID
	DisplayName    string
	Qty            decimal.Decimal
	BaseQty        decimal.Decimal
	Unit           string
	CuttingSpecID  *uuid.UUID
	SpecialCutNote *string
	WarehouseID    *uuid.UUID
}

// assembleItems(細部 4.2.1/4.2.2/4.2.3 品項層):
// 客戶身分 → product_id 必須在自己 customer_products;手打已於 guard 擋。
// 員工 → 選品換算 ToBase;手打 product_id=NULL, base_qty=qty;save_alias → EnsureAlias(同 tx)。
func assembleItems(ctx context.Context, tx *ent.Tx, actor rls.Identity, customerID uuid.UUID, items []OrderItemInput) ([]assembledItem, error) {
	out := make([]assembledItem, 0, len(items))
	for _, it := range items {
		if it.Qty.LessThanOrEqual(decimal.Zero) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("qty must be positive"))
		}
		ai := assembledItem{
			ProductID: it.ProductID, DisplayName: it.DisplayName, Qty: it.Qty, Unit: it.Unit,
			CuttingSpecID: it.CuttingSpecID, SpecialCutNote: it.SpecialCutNote, WarehouseID: it.WarehouseID,
		}
		if it.ProductID == nil {
			// 純手打(僅員工能走到這):base_qty = qty(細部 4.2.1 步驟 5)
			if it.ManualName == nil || *it.ManualName == "" {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("manual_name or product_id required"))
			}
			ai.DisplayName = *it.ManualName
			ai.BaseQty = it.Qty
			out = append(out, ai)
			continue
		}
		// 客戶守衛:品項須在自己專屬清單(細部 4.2.3 步驟 2)
		if actor.Role == "customer" {
			ok, err := tx.CustomerProduct.Query().Where(
				entcp.CustomerIDEQ(customerID),
				entcp.ProductIDEQ(*it.ProductID),
				entcp.DeletedAtIsNil(),
			).Exist(ctx)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, err)
			}
			if !ok {
				return nil, connect.NewError(connect.CodeInvalidArgument,
					errors.New("item not in customer product list"))
			}
		}
		// 換算(4.2.1)
		units, err := tx.ProductUnit.Query().Where(
			entpu.ProductIDEQ(*it.ProductID), entpu.DeletedAtIsNil(),
		).All(ctx)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		base, err := products.ToBase(it.Qty, it.Unit, units)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		ai.BaseQty = base
		// 別名(4.2.2 步驟 3):save_alias → upsert;衝突改更新,不擋單
		if it.SaveAlias {
			// 套件級函式直呼(04 計畫 T7 契約)
			if _, err := customerproducts.EnsureAlias(ctx, tx, customerID, *it.ProductID, it.DisplayName, it.Qty); err != nil {
				return nil, connect.NewError(connect.CodeInternal, err)
			}
		} else if actor.Role != "customer" {
			// 選用總表商品而清單尚無 → 自動建立(4.2.2 步驟 4,display_name 預設商品名)
			exists, err := tx.CustomerProduct.Query().Where(
				entcp.CustomerIDEQ(customerID), entcp.ProductIDEQ(*it.ProductID), entcp.DeletedAtIsNil(),
			).Exist(ctx)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, err)
			}
			if !exists {
				prod, err := tx.Product.Get(ctx, *it.ProductID)
				if err != nil {
					return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("product not found"))
				}
				if _, err := customerproducts.EnsureAlias(ctx, tx, customerID, *it.ProductID, prod.Name, it.Qty); err != nil {
					return nil, connect.NewError(connect.CodeInternal, err)
				}
			}
		}
		out = append(out, ai)
	}
	return out, nil
}

// adjustDeliveryDate(細部 4.2.5,D26):非勾選日順延至下一勾選日;未勾選維持。
func adjustDeliveryDate(ctx context.Context, tx *ent.Tx, customerID uuid.UUID, date string) (time.Time, error) {
	d, err := time.ParseInLocation("2006-01-02", date, bizLocation)
	if err != nil {
		return time.Time{}, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid expected_delivery_date"))
	}
	cust, err := tx.Customer.Get(ctx, customerID)
	if err != nil {
		return time.Time{}, connect.NewError(connect.CodeNotFound, errors.New("customer not found"))
	}
	days := cust.PreferredDeliveryDays // []int,勾選的星期編號 1=週一…6=週六
	if len(days) == 0 {
		return d, nil // 未勾選 → 維持
	}
	preferred := map[int]bool{}
	for _, day := range days {
		if day >= 1 && day <= 6 {
			preferred[day] = true
		}
	}
	if len(preferred) == 0 {
		return d, nil
	}
	// 逐日往後找(至多 7 天必命中;星期日 weekday=0 視同非勾選)
	for i := 0; i < 7; i++ {
		cand := d.AddDate(0, 0, i)
		wd := int(cand.Weekday()) // Sunday=0
		if preferred[wd] {
			return cand, nil
		}
	}
	return d, nil // 理論不可達
}
```

契約同步說明:`customerproducts.EnsureAlias` 需從 04 計畫 Task 7 的方法改為套件級函式 `customerproducts.EnsureAlias(ctx, tx, customerID, productID, aliasName string, defaultQty decimal.Decimal)`(04 計畫中的方法版僅為包裝;實作時以套件級函式為準,04 計畫 Self-Review 已記)。`defaultQty` 型別隨 decimal 契約改 `decimal.Decimal`。

- [ ] **Step 5: 跑測試確認通過 + Commit**

Run: `cd backend && go test ./internal/domain/salesorders/ -v -race`
Expected: PASS — 換算 10 條=6kg、清單守衛三路徑、別名建立與更新、順延兩情境。

```bash
git add backend/internal/domain/salesorders/ordering.go backend/internal/domain/salesorders/usecase.go backend/internal/domain/customerproducts
git commit -m "feat(backend): 下單組裝(換算/別名/清單守衛/送貨日順延)(4.2)"
```

---

## Self-Review 記錄

- **Spec 覆蓋**:4.1.1→T1;4.1.2→T2;4.1.3→T3;4.1.4-4.1.5→T4;4.2.1-4.2.5→T5。無缺漏。
- **整合測試重點對應**(細部文件文末三條):狀態機走查→T3 測試 + T4 `transition`;取號併發/雙軌/回滾→T2 三測試;順延四情境→T5 `TestPreferredDeliveryDayRollover`。
- **跨計畫契約**:`products.ToBase`(04 T5,decimal)→ T5;`customerproducts.EnsureAlias` 為套件級函式且 `defaultQty` 為 `decimal.Decimal`(04 T7 已對齊);通知採 07 計畫契約(`triggers.OnOrderCreated` 交易內建檔 + `notification.Dispatch` 提交後發送,`notification.Sender`/`FakeSender` 注入)——本計畫不自定 Notifier 介面;`Transition()`(T3)→ 08 計畫派車確認/取消呼叫;`preferred_delivery_days` 型別統一 `[]int`(04 T1 schema 已落地;細部文件 4.2.5 的 bool 陣列描述以 `[]int` 為準)。
- **已知佔位**:01-auth Task 6 `mapCustomerCode` 已於 04 T3 移除;Casbin 角色檢查(void/dispatch_cancel 需 dept_admin 以上)於 handler 層接 enforcer(02 計畫 T7 完成後接線);`lookupCustomerDepartment` 為簡單查詢函式。
- **類型一致**:`Status` 常數、`Transition` 簽名、`OrderItemInput`/`assembledItem` 欄位在 T3/T4/T5 一致。

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/backend/2026-08-17-backend-05-sales-orders-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** — 每個 Task 派新 subagent 執行,Task 間 review,迭代快。

**2. Inline Execution** — 用 executing-plans 在本 session 逐批執行,設 checkpoint review。

Which approach?

---

*計畫版本:v1.0.0(2026-08-17);對應細部文件 `detail/05-sales-orders.md`、原計畫 v2.9.0、規格書 v1.0.34。*
