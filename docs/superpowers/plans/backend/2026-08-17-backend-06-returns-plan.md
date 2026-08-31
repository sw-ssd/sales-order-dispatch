# Backend 06 — 退貨申請與審核 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 實作 backend 退貨申請與審核(原計畫 Task 4.7 的後端部分)— `return_requests` / `return_request_items` schema 與 RLS、客戶子帳號雙來源(歷史訂單品項 / 專屬商品)發起退貨與自查列表、業務審核(approved / rejected、樂觀鎖、不修改原訂單 D25)、退貨證明資料輸出、審核結果推播與同事務稽核。

**Architecture:** 依 `docs/superpowers/plans/backend/detail/06-returns.md`(下稱「細部文件」,子功能編號 4.7.x)實作。Connect-RPC `ReturnService`(Create / List / Get / Review / GetCertificate);資料範圍沿用 01-auth 計畫的 RLS session 注入(客戶帳號 `data_scope = self`,以 `customer_id` 收口);寫入路徑一律單一 DB 交易 + `audit.Recorder` 同事務稽核(D18);審核結果推播採 07 計畫契約:通知記錄(fcm/in_app pending)於審核**同一交易**內建立(D18),提交後 `notification.Dispatch` 發送,失敗僅標 `failed` 不回滾審核(D16);僅通知發起帳號(D23)。系統僅記錄申請與審核狀態,無配送取貨 API(D25)。

**Tech Stack:** Go 1.25、Ent(entgo.io)、Chi v5、Connect-RPC、pgx/v5、testcontainers-go(整合測試);測試共用 `testutil.NewEntClient`(01-auth 計畫 Task 1 提供)。

**Spec 來源:** 細部文件 `docs/superpowers/plans/backend/detail/06-returns.md`;共通規則見 `docs/superpowers/plans/backend/detail/00-index.md` §3。

## Global Constraints

- module 路徑:`github.com/salesorder/sales-order-1.0/backend`;所有路徑相對 repo root。
- 軟刪除:`return_requests` / `return_request_items` 統一 `deleted_at timestamptz`,查詢預設排除(D10);主檔軟刪除時品項於 usecase 層同事務一併軟刪除,DB 層不做 cascade delete。兩表無業務唯一約束需求(同一客戶可有多張 pending),不需部分唯一索引。
- 稽核(D18):Create / Review 與 audit log 同一 DB 交易,同成功同失敗;一律經 `audit.Recorder`(介面由 01-auth 計畫 Task 14 提供:`Record(ctx, audit.Entry) error`,`Entry{ActorID, ActorName, Action, ResourceType, ResourceID, Before, After string}`;DB 實作由 03-metadicts-audit 計畫 Task 2.6 提供,未就緒前以 `audit.NoopRecorder` 注入)。`Record` 必須在交易提交前呼叫,回錯即整筆 rollback。**敏感欄位永不入稽核。**
- 錯誤:RPC 層統一 Connect code — `unauthenticated` / `permission_denied` / `not_found` / `failed_precondition` / `invalid_argument`;樂觀鎖與狀態衝突一律 `failed_precondition`;usecase 層以 sentinel error 表達,handler 統一映射。
- RLS:兩表啟用 RLS;session 變數 `app.current_company_id` / `app.current_department_id` / `app.current_data_scope` / `app.current_customer_id`(01-auth 計畫 Task 3 `rls.WrapDriver` 注入);未注入 fail-closed(0 列)。
- 不修改原訂單(D25):本 domain 任何程式碼不得對 `sales_orders` / `sales_order_items` / `sales_order_events` 執行寫入,僅在品項上保留對來源訂單的參照;Task 3 以測試比對審核前後 `sales_orders` 全欄位零異動。
- 推播路由(D23):審核結果通知對象僅為 `return_requests.created_by_user_id` 單一帳號;發送失敗標 `failed` 不重試、不回滾審核(D16 修訂)。
- 跨 domain 依賴(皆為欄位參照,不設跨 domain DB 外鍵,避免 migration 順序耦合):
  - `customers` / `products` / `product_units` / `customer_products` / `file_assets` 實體由 04-master-data 計畫提供(細部 `04-master-data.md` 子功能 3.1.1 / 3.3.1 / 3.5.1 / 3.6.2);
  - `sales_orders` / `sales_order_items` 實體由 05-sales-orders 計畫提供(細部 `05-sales-orders.md` 子功能 4.1.1),其 schema 需含 `sales_order_items → sales_orders` 的 edge(`edge.From("sales_order", ...).Ref("items")`);
  - 通知範本渲染與 FCM / 站內發送由 07-notifications 計畫提供(細部 `07-notifications.md` 子功能 4.3.2 / 4.4.1 / 4.4.2);
  - 執行順序:本計畫排於 01 / 04 / 05 的 schema 合入之後;03 / 07 未就緒前以 `audit.NoopRecorder` 降級稽核、`sender = nil` 降級通知(不建檔不發送,審核功能不受影響)先行。
- Ent schema 統一放 `backend/ent/schema/`(Ent codegen 單一目錄限制,沿 01-auth 計畫;細部文件標示的 `internal/domain/returns/schema/` 僅為職責歸屬)。
- 測試:DB 相依測試走 testcontainers Postgres 16;`go test ./...` 必須全綠。每個 Task 結尾 commit;commit message 格式 `feat(backend): …` / `test(backend): …`。

## File Structure

| 檔案 | 職責 | 建立於 |
|---|---|---|
| `backend/ent/schema/return_request.go` | 退貨申請主檔 Ent schema | Task 1 |
| `backend/ent/schema/return_request_item.go` | 退貨品項檔 Ent schema(快照欄位) | Task 1 |
| `backend/database/migrations/00011_return_rls_policies.sql` | 兩表 RLS policy | Task 1 |
| `backend/internal/testutil/db.go` | 加 `NewEntClientWithMigrations`(Update) | Task 1 |
| `backend/proto/v1/returns.proto` | ReturnService proto 定義 | Task 2 |
| `backend/internal/domain/returns/errors.go` | sentinel error 與 Connect code 映射 | Task 2 |
| `backend/internal/domain/returns/returns.go` | Usecase 結構、輸入 / 檢視型別、常數 | Task 2 |
| `backend/internal/domain/returns/repo.go` | 來源核實與照片查詢(Ent) | Task 2 |
| `backend/internal/domain/returns/create.go` | 發起退貨 usecase | Task 2 |
| `backend/internal/domain/returns/list.go` | List / Get usecase | Task 2 |
| `backend/internal/domain/returns/handler.go` | Connect handler(Create / List / Get) | Task 2 |
| `backend/internal/domain/returns/review.go` | 審核 usecase(樂觀鎖) | Task 3 |
| `backend/internal/domain/returns/certificate.go` | 退貨證明 usecase | Task 3 |
| `backend/internal/domain/returns/notify.go` | 審核結果通知建檔(07 計畫契約直呼) | Task 3 |
| `backend/internal/domain/returns/fixture_test.go` | 跨 Task 測試 fixture | Task 2 |

---

### Task 1: return_requests / return_request_items Ent schema 與 RLS(細部 4.7.1)

**Files:**
- Create: `backend/ent/schema/return_request.go`
- Create: `backend/ent/schema/return_request_item.go`
- Create: `backend/database/migrations/00011_return_rls_policies.sql`
- Update: `backend/internal/testutil/db.go`
- Test: `backend/ent/schema/return_schema_test.go`

**Interfaces:**
- Consumes: `testutil.NewEntClient`(01-auth 計畫 Task 1);`rls.Identity` / `rls.NewContext` / `rls.WrapDriver`(01-auth 計畫 Task 3)。
- Produces: `ent.ReturnRequest` / `ent.ReturnRequestItem` 產生碼與 predicate 套件 `returnrequest` / `returnrequestitem`;`returnrequest.Status`(enum:`StatusPending` / `StatusApproved` / `StatusRejected`)、`returnrequestitem.SourceType`(enum:`SourceTypeOrderItem` / `SourceTypeCustomerProduct`);`testutil.NewEntClientWithMigrations(t, files ...string) *ent.Client`(RLS 注入 + 指定 migration 檔,供本計畫與後續 domain 的 RLS 測試使用)。

- [ ] **Step 1: testutil 增加多 migration RLS 輔助**

`backend/internal/testutil/db.go` 追加(沿用 01-auth 計畫 Task 3 `NewEntClientWithRLS` 的實作模式,改為可指定 migration 檔清單):

```go
import (
	"os"
	"path/filepath"
	"strings"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	// 既有 import 省略
)

// NewEntClientWithMigrations 啟動拋棄式 Postgres,跑完 Ent schema migration 後,
// 依序套用指定的 SQL migration(僅 Up 段),回傳經 rls.WrapDriver 包裝的 Ent client。
// files 為 backend/database/migrations/ 下的檔名,依序套用。
func NewEntClientWithMigrations(t *testing.T, files ...string) *ent.Client {
	t.Helper()
	ctx := context.Background()
	base := NewEntClient(t)
	for _, f := range files {
		sqlBytes, err := os.ReadFile(filepath.Join("../../database/migrations", f))
		if err != nil {
			t.Fatalf("read migration %s: %v", f, err)
		}
		up := strings.SplitN(string(sqlBytes), "-- +goose Down", 2)[0]
		up = strings.ReplaceAll(up, "-- +goose Up", "")
		if _, err := base.DB().ExecContext(ctx, up); err != nil {
			t.Fatalf("apply migration %s: %v", f, err)
		}
	}
	drv := rls.WrapDriver(entsql.OpenDB(dialect.Postgres, base.DB()))
	return ent.NewClient(ent.Driver(drv))
}
```

- [ ] **Step 2: 寫失敗測試(schema 行為、enum 白名單、外鍵、RLS 隔離)**

`backend/ent/schema/return_schema_test.go`:

```go
package schema_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/returnrequest"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func createRequest(t *testing.T, c *ent.Client, ctx context.Context, co, dep, cust uuid.UUID) *ent.ReturnRequest {
	t.Helper()
	rr, err := c.ReturnRequest.Create().
		SetCompanyID(co).SetDepartmentID(dep).SetCustomerID(cust).
		SetCreatedByUserID(uuid.New()).SetRemark("外箱破損").
		Save(ctx)
	if err != nil {
		t.Fatalf("create return_request: %v", err)
	}
	return rr
}

func TestReturnRequestDefaultsAndItems(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, dep, cust := uuid.New(), uuid.New(), uuid.New()

	rr := createRequest(t, c, ctx, co, dep, cust)
	if rr.Status != returnrequest.StatusPending {
		t.Fatalf("default status: got %q", rr.Status)
	}
	if rr.Version != 0 {
		t.Fatalf("default version: got %d", rr.Version)
	}
	if rr.ReviewedAt != nil || rr.ReviewedByUserID != nil {
		t.Fatal("review fields should be nil before review")
	}

	item, err := c.ReturnRequestItem.Create().
		SetReturnRequestID(rr.ID).SetCompanyID(co).SetDepartmentID(dep).
		SetSourceType("order_item").
		SetSalesOrderID(uuid.New()).SetSalesOrderItemID(uuid.New()).
		SetProductID(uuid.New()).
		SetProductName("去骨雞腿").SetProductSpec("約 2kg/包").SetUnit("包").
		SetQuantity(2).SetReason("解凍後異味").
		SetPhotoFileIds([]uuid.UUID{uuid.New(), uuid.New()}).
		Save(ctx)
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	if len(item.PhotoFileIds) != 2 {
		t.Fatalf("photo_file_ids round-trip: got %v", item.PhotoFileIds)
	}
	// 主檔 edge 可查品項
	n, err := rr.QueryItems().Count(ctx)
	if err != nil || n != 1 {
		t.Fatalf("items edge: n=%d err=%v", n, err)
	}
}

func TestReturnRequestStatusRejectsInvalidValue(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	_, err := c.ReturnRequest.Create().
		SetCompanyID(uuid.New()).SetDepartmentID(uuid.New()).SetCustomerID(uuid.New()).
		SetCreatedByUserID(uuid.New()).
		SetStatus(returnrequest.Status("hold")). // 非法枚舉值,繞過型別檢查
		Save(ctx)
	if err == nil {
		t.Fatal("invalid status should be rejected")
	}
}

func TestReturnRequestItemFK(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	_, err := c.ReturnRequestItem.Create().
		SetReturnRequestID(uuid.New()). // 不存在的主檔
		SetCompanyID(uuid.New()).SetDepartmentID(uuid.New()).
		SetSourceType("customer_product").
		SetCustomerProductID(uuid.New()).
		SetProductName("x").SetProductSpec("x").SetUnit("包").
		SetQuantity(1).SetReason("r").
		Save(ctx)
	if err == nil {
		t.Fatal("orphan item should violate FK")
	}
}

func TestReturnRequestsRLSIsolation(t *testing.T) {
	c := testutil.NewEntClientWithMigrations(t, "00011_return_rls_policies.sql")
	ctx := context.Background()
	coA, coB := uuid.New(), uuid.New()
	depA := uuid.New()
	custA, custB := uuid.New(), uuid.New()

	// 以 data_scope=all 身分播種(細部 1.3.2:管理 / 測試資料走高權注入)
	super := rls.NewContext(ctx, rls.Identity{UserID: uuid.New(), CompanyID: coA, DataScope: "all"})
	createRequest(t, c, super, coA, depA, custA)
	createRequest(t, c, super, coB, uuid.New(), custB)

	// 未注入身分:fail-closed,0 列
	if n, err := c.ReturnRequest.Query().Count(ctx); err != nil || n != 0 {
		t.Fatalf("fail-closed violated: n=%d err=%v", n, err)
	}
	// B 公司視角:只見自己 1 筆
	ctxB := rls.NewContext(ctx, rls.Identity{UserID: uuid.New(), CompanyID: coB, DataScope: "company"})
	if n, err := c.ReturnRequest.Query().Count(ctxB); err != nil || n != 1 {
		t.Fatalf("tenant isolation violated: n=%d err=%v", n, err)
	}
	// A 公司客戶 self 視角:僅見 custA 的 1 筆
	selfA := rls.NewContext(ctx, rls.Identity{
		UserID: uuid.New(), CompanyID: coA, CustomerID: &custA, DataScope: "self",
	})
	if n, err := c.ReturnRequest.Query().Count(selfA); err != nil || n != 1 {
		t.Fatalf("self scope violated: n=%d err=%v", n, err)
	}
	// A 公司 department 視角:僅見 depA
	deptA := rls.NewContext(ctx, rls.Identity{
		UserID: uuid.New(), CompanyID: coA, DepartmentID: &depA, DataScope: "department",
	})
	if n, err := c.ReturnRequest.Query().Count(deptA); err != nil || n != 1 {
		t.Fatalf("department scope violated: n=%d err=%v", n, err)
	}
}
```

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./ent/schema/ -run 'TestReturnRequest' -v`
Expected: FAIL — `ent.ReturnRequest` / `ent.ReturnRequestItem` 未定義、`testutil.NewEntClientWithMigrations` 未定義(編譯失敗)。

- [ ] **Step 4: 實作兩個 schema 與 RLS migration**

`backend/ent/schema/return_request.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ReturnRequest 為退貨申請主檔(細部文件 4.7.1)。
// customer_id / created_by_user_id 為邏輯參照(04-plan customers、01-plan users),
// 不設 DB 外鍵以避免跨 domain migration 耦合。
type ReturnRequest struct{ ent.Schema }

func (ReturnRequest) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("company_id", uuid.UUID{}),
		field.UUID("department_id", uuid.UUID{}),
		field.UUID("customer_id", uuid.UUID{}),          // 退貨所屬客戶(必填)
		field.UUID("created_by_user_id", uuid.UUID{}),   // 發起帳號 = 推播目標(D23)
		field.Enum("status").Values("pending", "approved", "rejected").Default("pending"),
		field.String("remark").Default(""),
		field.UUID("reviewed_by_user_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("reviewed_at").Optional().Nillable(),
		field.String("reject_reason").Optional().Nillable(), // rejected 時必填(usecase 保證)
		field.Int("version").Default(0),                     // 樂觀鎖
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (ReturnRequest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("items", ReturnRequestItem.Type),
	}
}

func (ReturnRequest) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("company_id", "customer_id", "status"),   // 客戶自查列表
		index.Fields("company_id", "department_id", "status"), // 業務待審清單
	}
}
```

`backend/ent/schema/return_request_item.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ReturnRequestItem 為退貨品項檔(細部文件 4.7.1)。
// 品項一律快照 product 名稱 / 規格 / 單位,來源日後異動不影響證明內容;
// sales_order_id / sales_order_item_id / customer_product_id 僅參照用途(D25)。
type ReturnRequestItem struct{ ent.Schema }

func (ReturnRequestItem) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("return_request_id", uuid.UUID{}),
		field.UUID("company_id", uuid.UUID{}),
		field.UUID("department_id", uuid.UUID{}), // 冗餘帶入,供 RLS
		field.Enum("source_type").Values("order_item", "customer_product"),
		field.UUID("sales_order_id", uuid.UUID{}).Optional().Nillable(),      // source_type=order_item 時必填
		field.UUID("sales_order_item_id", uuid.UUID{}).Optional().Nillable(), // 同上
		field.UUID("customer_product_id", uuid.UUID{}).Optional().Nillable(), // source_type=customer_product 時必填
		field.UUID("product_id", uuid.UUID{}).Optional().Nillable(),          // 手打品項可無 product
		field.String("product_name"),  // 快照
		field.String("product_spec"),  // 快照
		field.String("unit"),          // 快照
		field.Float("quantity").
			SchemaType(map[string]string{dialect.Postgres: "numeric(12,3)"}).Positive(),
		field.String("reason").NotEmpty(),
		field.JSON("photo_file_ids", []uuid.UUID{}).Optional(), // 元素為 file_assets.id
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (ReturnRequestItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("return_request", ReturnRequest.Type).
			Ref("items").Field("return_request_id").Unique().Required(),
	}
}

func (ReturnRequestItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("return_request_id"),
	}
}
```

`backend/database/migrations/00011_return_rls_policies.sql`(goose 格式;policy 模式同 01-auth 計畫 Task 3 的 00003,序號以合併主線時下一可用流水為準):

```sql
-- +goose Up
-- 細部文件 4.7.1 步驟 5:return_requests / return_request_items 啟用 RLS。
-- 未注入 session variables 時 current_setting(..., true) 回 NULL,policy 不命中 → fail-closed。

ALTER TABLE return_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE return_request_items ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON return_requests
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true));

CREATE POLICY data_scope ON return_requests
  USING (current_setting('app.current_data_scope', true) IN ('all', 'company')
         OR (current_setting('app.current_data_scope', true) = 'department'
             AND department_id::text = current_setting('app.current_department_id', true))
         OR (current_setting('app.current_data_scope', true) = 'self'
             AND customer_id::text = current_setting('app.current_customer_id', true)));

CREATE POLICY tenant_isolation ON return_request_items
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true));

CREATE POLICY data_scope ON return_request_items
  USING (current_setting('app.current_data_scope', true) IN ('all', 'company')
         OR (current_setting('app.current_data_scope', true) = 'department'
             AND department_id::text = current_setting('app.current_department_id', true))
         OR (current_setting('app.current_data_scope', true) = 'self'
             AND return_request_id IN (
               SELECT id FROM return_requests
               WHERE customer_id::text = current_setting('app.current_customer_id', true))));

-- +goose Down
DROP POLICY IF EXISTS data_scope ON return_request_items;
DROP POLICY IF EXISTS tenant_isolation ON return_request_items;
DROP POLICY IF EXISTS data_scope ON return_requests;
DROP POLICY IF EXISTS tenant_isolation ON return_requests;
ALTER TABLE return_request_items DISABLE ROW LEVEL SECURITY;
ALTER TABLE return_requests DISABLE ROW LEVEL SECURITY;
```

- [ ] **Step 5: codegen + 跑測試確認通過**

Run: `cd backend && go generate ./ent && go test ./ent/schema/ -run 'TestReturnRequest' -v`
Expected: PASS — 四個測試(預設值與品項、enum 拒絕、外鍵、RLS 三層隔離 + fail-closed)。

- [ ] **Step 6: Commit**

```bash
git add backend/ent backend/database/migrations/00011_return_rls_policies.sql backend/internal/testutil
git commit -m "feat(backend): return_requests / return_request_items schema 與 RLS policy(4.7.1)"
```


---

### Task 2: 發起退貨 Create 與客戶自查 List / Get(細部 4.7.2)

**Files:**
- Create: `backend/proto/v1/returns.proto`
- Create: `backend/internal/domain/returns/{errors,returns,repo,create,list,handler}.go`
- Test: `backend/internal/domain/returns/create_test.go`

**Interfaces:**
- Consumes: Task 1 schema;04 計畫 `customers` / `customer_products` / `file_assets`(下載 URL 規則 `/api/v1/files/:id/download`);05 計畫 `sales_orders` / `sales_order_items`(僅讀取核實);`audit.Recorder`。
- Produces: Connect-RPC `ReturnService`(本 Task 掛 `Create` / `List` / `Get`);usecase `returns.Usecase`;sentinel error → Connect code 映射表。

**常數:** `maxPhotosPerItem = 5`(每品項照片上限;細部文件建議值,1.0 以此為準)。

- [ ] **Step 1: 寫失敗測試(雙來源建立、核實失敗、主帳號拒絕、自查範圍、稽核同事務)**

```go
// backend/internal/domain/returns/create_test.go
package returns_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/returns"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

// fixture(seedFixtures)建立:公司+部門、客戶甲(custID,default_sales_rep=staffID)、
// 客戶乙(cust2ID)、商品(含單位)、customer_product(甲)、甲的歷史訂單+品項(orderID/itemID)、
// 甲的照片 file_asset(photoID)、子帳號身分(custActor: role=customer, customer_id=甲)。

func TestCreateFromOrderItem(t *testing.T) {
	c := testutil.NewEntClient(t)
	f := seedFixtures(t, c)
	ctx := context.Background()
	rec := &audit.FakeRecorder{}
	uc := returns.NewUsecase(c, rec, nil)

	req, err := uc.Create(ctx, f.custActor, returns.CreateInput{
		Remark: "整批外包破損",
		Items: []returns.ItemInput{{
			SourceType:      returns.SourceOrderItem,
			SalesOrderItemID: &f.itemID,
			Quantity:        decimal.NewFromInt(2),
			Reason:          "破損",
			PhotoFileIDs:    []uuid.UUID{f.photoID},
		}},
	})
	if err != nil { t.Fatalf("create: %v", err) }
	if req.Status != "pending" { t.Fatalf("status: %s", req.Status) }
	// 品項快照與參照
	item := c.ReturnRequestItem.Query().Where().OnlyX(ctx)
	if item.SalesOrderID == nil || *item.SalesOrderID != f.orderID {
		t.Fatalf("source ref: %+v", item)
	}
	if item.ProductName != "去骨雞腿" || item.Unit != "包" {
		t.Fatalf("snapshot: %+v", item)
	}
	// 稽核同事務:return_request.create 一筆
	if len(rec.Entries) != 1 || rec.Entries[0].Action != "return_request.create" {
		t.Fatalf("audit: %+v", rec.Entries)
	}
}

func TestCreateFromCustomerProduct(t *testing.T) {
	c := testutil.NewEntClient(t)
	f := seedFixtures(t, c)
	ctx := context.Background()
	uc := returns.NewUsecase(c, &audit.FakeRecorder{}, nil)

	_, err := uc.Create(ctx, f.custActor, returns.CreateInput{
		Items: []returns.ItemInput{{
			SourceType:        returns.SourceCustomerProduct,
			CustomerProductID: &f.cpID,
			Quantity:          decimal.NewFromInt(1),
			Reason:            "過期",
		}},
	})
	if err != nil { t.Fatalf("create: %v", err) }
}

func TestCreateGuards(t *testing.T) {
	c := testutil.NewEntClient(t)
	f := seedFixtures(t, c)
	ctx := context.Background()
	uc := returns.NewUsecase(c, &audit.FakeRecorder{}, nil)

	// 主帳號 → permission_denied
	primary := rls.Identity{UserID: uuid.New(), CompanyID: f.companyID, DepartmentID: &f.deptID,
		Role: "customer", IsPrimary: true, DataScope: "self", CustomerID: &f.custID}
	_, err := uc.Create(ctx, primary, returns.CreateInput{Items: []returns.ItemInput{{
		SourceType: returns.SourceCustomerProduct, CustomerProductID: &f.cpID,
		Quantity: decimal.NewFromInt(1), Reason: "x"}}})
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("primary: want permission_denied, got %v", err)
	}
	// 空品項 / 數量非正 / 原因空白 / 配對錯誤(order_item 缺 item_id)/ 照片 >5 → invalid_argument
	// 他人客戶來源品項(乙無此 cp)→ not_found;他人照片 → not_found
	// (逐一子測試;t.Run 表格驅動)
}

func TestListSelfScope(t *testing.T) {
	c := testutil.NewEntClient(t)
	f := seedFixtures(t, c)
	ctx := context.Background()
	uc := returns.NewUsecase(c, &audit.FakeRecorder{}, nil)
	// 甲建一筆;乙建一筆(乙的 actor)
	// 甲 List → 僅自己 1 筆;乙 Get 甲的 id → not_found
}

func TestGetReturnsPhotoURLs(t *testing.T) {
	// Get 回應品項 photo_file_ids → ["/api/v1/files/<id>/download", ...]
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/returns/ -v`
Expected: FAIL — `returns.Usecase` 未定義。

- [ ] **Step 3: 實作 proto 與 usecase**

`backend/proto/v1/returns.proto`:

```proto
syntax = "proto3";
package salesorder.v1;

service ReturnService {
  rpc CreateReturnRequest(CreateReturnRequestRequest) returns (CreateReturnRequestResponse);
  rpc ListReturnRequests(ListReturnRequestsRequest) returns (ListReturnRequestsResponse);
  rpc GetReturnRequest(GetReturnRequestRequest) returns (GetReturnRequestResponse);
  rpc ReviewReturnRequest(ReviewReturnRequestRequest) returns (ReviewReturnRequestResponse);   // Task 3
  rpc GetReturnCertificate(GetReturnCertificateRequest) returns (GetReturnCertificateResponse); // Task 4
}

message ReturnItemInput {
  string source_type = 1;                     // order_item | customer_product
  optional string sales_order_item_id = 2;    // source_type=order_item 必填
  optional string customer_product_id = 3;    // source_type=customer_product 必填
  string quantity = 4;                        // decimal 字串,正數
  string reason = 5;                          // 必填
  repeated string photo_file_ids = 6;         // ≤5
}
message CreateReturnRequestRequest { repeated ReturnItemInput items = 1; string remark = 2; }
// ListReturnRequestsRequest{page, per_page, status?};GetReturnRequestRequest{id}
// ReviewReturnRequestRequest{id, decision(approved|rejected), reject_reason?, expected_version}
```

`errors.go`:

```go
package returns

import (
	"errors"

	"connectrpc.com/connect"
)

var (
	ErrNotFound       = errors.New("return request not found")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalid        = errors.New("invalid argument")
	ErrNotPending     = errors.New("not pending")
	ErrVersionConflict = errors.New("version conflict")
	ErrNotApproved    = errors.New("not approved")
)

// CodeOf 統一映射(handler 使用)。
func CodeOf(err error) connect.Code {
	switch {
	case errors.Is(err, ErrNotFound):
		return connect.CodeNotFound
	case errors.Is(err, ErrForbidden):
		return connect.CodePermissionDenied
	case errors.Is(err, ErrNotPending), errors.Is(err, ErrVersionConflict), errors.Is(err, ErrNotApproved):
		return connect.CodeFailedPrecondition
	case errors.Is(err, ErrInvalid):
		return connect.CodeInvalidArgument
	}
	return connect.CodeInternal
}
```

`returns.go`(型別與 Usecase):

```go
package returns

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
)

const (
	SourceOrderItem       = "order_item"
	SourceCustomerProduct = "customer_product"
	maxPhotosPerItem      = 5
)

type Usecase struct {
	db       *ent.Client
	rec      audit.Recorder
	sender notification.Sender // 07 計畫 T4 介面;nil = 通知關閉(測試注入 FakeSender)
}

func NewUsecase(db *ent.Client, rec audit.Recorder, sender notification.Sender) *Usecase {
	return &Usecase{db: db, rec: rec, sender: sender}
}

type ItemInput struct {
	SourceType        string
	SalesOrderItemID  *uuid.UUID
	CustomerProductID *uuid.UUID
	Quantity          decimal.Decimal
	Reason            string
	PhotoFileIDs      []uuid.UUID
}

type CreateInput struct {
	Remark string
	Items  []ItemInput
}
```

`create.go`(交易邊界:主檔 + 品項快照 + 稽核,同成功同失敗):

```go
package returns

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entfa "github.com/salesorder/sales-order-1.0/backend/ent/fileasset"
	entsoi "github.com/salesorder/sales-order-1.0/backend/ent/salesorderitem"
	entso "github.com/salesorder/sales-order-1.0/backend/ent/salesorder"
	entcp "github.com/salesorder/sales-order-1.0/backend/ent/customerproduct"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/shopspring/decimal"
)

// Create 客戶子帳號發起退貨(細部 4.7.2)。
// customer_id 一律取自 session;主帳號拒絕;快照商品名稱/單位(證明用,4.7.4)。
func (u *Usecase) Create(ctx context.Context, actor rls.Identity, in CreateInput) (*ent.ReturnRequest, error) {
	if actor.Role != "customer" || actor.CustomerID == nil {
		return nil, ErrForbidden // staff / 其他身分呼叫客戶端 Create
	}
	if actor.IsPrimary {
		return nil, ErrForbidden // 主帳號僅供帳號管理
	}
	if len(in.Items) == 0 {
		return nil, fmt.Errorf("%w: items required", ErrInvalid)
	}
	custID := *actor.CustomerID

	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	// 客戶歸屬快照(company/department 以客戶為準,session 注入已限縮同公司)
	cust, err := tx.Customer.Get(ctx, custID)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("%w: customer", ErrNotFound)
	}

	req, err := tx.ReturnRequest.Create().
		SetCompanyID(cust.CompanyID).SetDepartmentID(cust.DepartmentID).
		SetCustomerID(custID).SetCreatedByUserID(actor.UserID).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	for _, it := range in.Items {
		item, err := u.buildItem(ctx, tx, cust, req.ID, it)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if _, err := item.Save(ctx); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	if err := u.rec.Record(audit.ContextWithTx(ctx, tx), audit.Entry{
		ActorID: actor.UserID.String(), Action: "return_request.create",
		ResourceType: "return_request", ResourceID: req.ID.String(),
		After: fmt.Sprintf(`{"customer_id":%q,"items":%d}`, custID, len(in.Items)),
	}); err != nil {
		_ = tx.Rollback() // 稽核失敗整筆回滾(D18)
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return req, nil
}

// buildItem 核實單一品項並回傳 create builder(含快照)。
func (u *Usecase) buildItem(ctx context.Context, tx *ent.Tx, cust *ent.Customer, reqID uuid.UUID, in ItemInput) (*ent.ReturnRequestItemCreate, error) {
	if in.Quantity.LessThanOrEqual(decimal.Zero) || in.Reason == "" {
		return nil, fmt.Errorf("%w: quantity/reason", ErrInvalid)
	}
	if len(in.PhotoFileIDs) > maxPhotosPerItem {
		return nil, fmt.Errorf("%w: too many photos", ErrInvalid)
	}
	// 照片核實:存在、同公司、未軟刪除
	for _, fid := range in.PhotoFileIDs {
		ok, err := tx.FileAsset.Query().Where(
			entfa.IDEQ(fid), entfa.CompanyIDEQ(cust.CompanyID), entfa.DeletedAtIsNil(),
		).Exist(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("%w: photo %s", ErrNotFound, fid)
		}
	}

	b := tx.ReturnRequestItem.Create().
		SetReturnRequestID(reqID).
		SetCompanyID(cust.CompanyID).SetDepartmentID(cust.DepartmentID).
		SetSourceType(in.SourceType).
		SetQuantity(in.Quantity).SetReason(in.Reason).
		SetPhotoFileIds(in.PhotoFileIDs)

	switch in.SourceType {
	case SourceOrderItem:
		if in.SalesOrderItemID == nil {
			return nil, fmt.Errorf("%w: sales_order_item_id required", ErrInvalid)
		}
		// 來源核實:品項存在且所屬訂單屬本客戶、未軟刪除;不檢核訂單狀態與可退上限(1.0,D25)
		soi, err := tx.SalesOrderItem.Query().Where(
			entsoi.IDEQ(*in.SalesOrderItemID), entsoi.DeletedAtIsNil(),
		).Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: order item", ErrNotFound)
		}
		order, err := tx.SalesOrder.Query().Where(
			entso.IDEQ(soi.SalesOrderID), entso.CustomerIDEQ(cust.ID), entso.DeletedAtIsNil(),
		).Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: order item", ErrNotFound)
		}
		if soi.ProductID == nil {
			return nil, fmt.Errorf("%w: manual order item has no product ref", ErrInvalid)
		}
		prod, err := tx.Product.Get(ctx, *soi.ProductID)
		if err != nil {
			return nil, err
		}
		b.SetSalesOrderID(order.ID).SetSalesOrderItemID(soi.ID).
			SetProductID(prod.ID).SetProductName(soi.DisplayName).SetUnit(soi.Unit)

	case SourceCustomerProduct:
		if in.CustomerProductID == nil {
			return nil, fmt.Errorf("%w: customer_product_id required", ErrInvalid)
		}
		cp, err := tx.CustomerProduct.Query().Where(
			entcp.IDEQ(*in.CustomerProductID), entcp.CustomerIDEQ(cust.ID), entcp.DeletedAtIsNil(),
		).Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: customer product", ErrNotFound)
		}
		prod, err := tx.Product.Get(ctx, cp.ProductID)
		if err != nil {
			return nil, err
		}
		b.SetCustomerProductID(cp.ID).
			SetProductID(prod.ID).SetProductName(cp.AliasName).SetUnit(prod.BaseUnitCode)

	default:
		return nil, fmt.Errorf("%w: source_type", ErrInvalid)
	}
	return b, nil
}
```

`list.go`:`List` 客戶強制 `customer_id = actor.CustomerID`;staff 走 data_scope(RLS + where)。`Get`:含品項;照片 `file_assets.id` 轉 `"/api/v1/files/" + id + "/download"`;Get 不可見 → `not_found`。handler 依 01-auth 計畫模式(Connect handler、errors.CodeOf 映射)。

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go generate ./ent/ && go test ./internal/domain/returns/ -v`
Expected: PASS。

```bash
git add backend/proto/v1/returns.proto backend/internal/domain/returns
git commit -m "feat(backend): 退貨發起(雙來源+照片核實+快照)與客戶自查(4.7.2)"
```

---

### Task 3: 審核 API(細部 4.7.3)+ 審核稽核與推播掛點(細部 4.7.5)

**Files:**
- Create: `backend/internal/domain/returns/review.go`
- Create: `backend/internal/domain/returns/notify.go`
- Test: `backend/internal/domain/returns/review_test.go`

**Interfaces:**
- Consumes: Task 2 Usecase;`audit.Recorder`;07 計畫 `notifications.Render` / `notifications.Repo.CreateOne` / `notification.Dispatch` / `notification.Sender`。
- Produces:
  - `Usecase.Review(ctx, actor, id, decision, rejectReason string, expectedVersion int) (*ent.ReturnRequest, error)`
  - `buildReviewNotifications(ctx, tx, req, decision, rejectReason) ([]uuid.UUID, error)`(`notify.go`;審核**交易內**為發起帳號建 fcm/in_app 各一筆 `pending` 通知記錄 — D18;僅 `created_by_user_id` 單一帳號,D23;提交後由 Review 呼叫 `notification.Dispatch` 發送,失敗僅標 `failed` 不回滾 — D16;不自定 Notifier 介面,直接消費 07 計畫契約)
  - 範本代碼:`return_reviewed`(fcm / in_app 兩通道,變數 `{request_id, decision, reject_reason, customer_name}`);範本由管理者經 07 計畫範本 CRUD 建立,缺失時降級為無範本記錄(與 07 計畫 T5 行為一致)

- [ ] **Step 1: 寫失敗測試(狀態機、樂觀鎖併發、權限、原訂單零異動、稽核同事務)**

```go
// backend/internal/domain/returns/review_test.go
package returns_test

func TestReviewApproveBySalesRep(t *testing.T) {
	c := testutil.NewEntClient(t)
	f := seedFixtures(t, c)
	ctx := context.Background()
	rec := &audit.FakeRecorder{}
	sender := &notification.FakeSender{}
	uc := returns.NewUsecase(c, rec, sender)
	req := createPending(t, uc, f) // 甲子帳號建單

	rep := f.staffActor // 主責業務(= default_sales_rep_id)
	got, err := uc.Review(ctx, rep, req.ID, "approved", "", req.Version)
	if err != nil { t.Fatalf("review: %v", err) }
	if got.Status != "approved" || got.ReviewedByUserID == nil || got.ReviewedAt == nil {
		t.Fatalf("after approve: %+v", got)
	}
	// 通知路由:僅發起帳號 fcm/in_app 各一筆且已 dispatch(D23;同客戶其他帳號 0 筆)
	notifs := c.Notification.Query().Where().AllX(ctx)
	if len(notifs) != 2 {
		t.Fatalf("want 2 notifications (fcm+in_app), got %d", len(notifs))
	}
	for _, nt := range notifs {
		if nt.UserID != f.custActor.UserID {
			t.Fatalf("notification to wrong user: %+v", nt)
		}
	}
	if rec.Entries[len(rec.Entries)-1].Action != "return_request.review" {
		t.Fatalf("audit: %+v", rec.Entries)
	}
}

func TestReviewRejectRequiresReason(t *testing.T) {
	// decision=rejected, reject_reason="" → invalid_argument,狀態不變
}

func TestReviewStateMachine(t *testing.T) {
	// 已 approved 再 Review(approve 或 reject)→ failed_precondition
}

func TestReviewOptimisticLockConcurrency(t *testing.T) {
	c := testutil.NewEntClient(t)
	f := seedFixtures(t, c)
	ctx := context.Background()
	uc := returns.NewUsecase(c, &audit.FakeRecorder{}, nil)
	req := createPending(t, uc, f)

	// 兩 staff 同 version 併發審核 → 僅一成功
	errs := make(chan error, 2)
	go func() { _, err := uc.Review(ctx, f.staffActor, req.ID, "approved", "", req.Version); errs <- err }()
	go func() { _, err := uc.Review(ctx, f.adminActor, req.ID, "rejected", "x", req.Version); errs <- err }()
	var ok, conflict int
	for i := 0; i < 2; i++ {
		err := <-errs
		if err == nil { ok++ } else if connect.CodeOf(err) == connect.CodeFailedPrecondition { conflict++ }
	}
	if ok != 1 || conflict != 1 {
		t.Fatalf("concurrency: ok=%d conflict=%d", ok, conflict)
	}
}

func TestReviewPermissions(t *testing.T) {
	// 非主責一般 staff → permission_denied;客戶子帳號 → permission_denied;
	// 主帳號 → permission_denied;dept_admin 可審
}

func TestReviewDoesNotTouchOrder(t *testing.T) {
	c := testutil.NewEntClient(t)
	f := seedFixtures(t, c)
	ctx := context.Background()
	uc := returns.NewUsecase(c, &audit.FakeRecorder{}, nil)
	req := createPending(t, uc, f)

	before := c.SalesOrder.GetX(ctx, f.orderID)
	beforeEvents := c.SalesOrderEvent.Query().Where().CountX(ctx)
	if _, err := uc.Review(ctx, f.staffActor, req.ID, "approved", "", req.Version); err != nil {
		t.Fatal(err)
	}
	after := c.SalesOrder.GetX(ctx, f.orderID)
	if !ordersEqual(before, after) { // 逐欄位比對(含 version)
		t.Fatal("D25 violated: sales_orders mutated")
	}
	if got := c.SalesOrderEvent.Query().Where().CountX(ctx); got != beforeEvents {
		t.Fatalf("D25 violated: events %d -> %d", beforeEvents, got)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/returns/ -run TestReview -v`
Expected: FAIL — `Review` 未定義。

- [ ] **Step 3: 實作 review.go**

```go
package returns

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entrr "github.com/salesorder/sales-order-1.0/backend/ent/returnrequest"
	entcust "github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// Review 業務審核(細部 4.7.3):僅 pending;樂觀鎖;審核者 = 主責業務或 dept_admin 以上;
// 不觸碰 sales_orders*(D25);稽核同事務;通知於提交後。
func (u *Usecase) Review(ctx context.Context, actor rls.Identity, id uuid.UUID, decision, rejectReason string, expectedVersion int) (*ent.ReturnRequest, error) {
	if actor.Role == "customer" {
		return nil, ErrForbidden // 客戶帳號(含主帳號)不得審核
	}
	if decision != "approved" && decision != "rejected" {
		return nil, fmt.Errorf("%w: decision", ErrInvalid)
	}
	if decision == "rejected" && rejectReason == "" {
		return nil, fmt.Errorf("%w: reject_reason required", ErrInvalid)
	}

	req, err := u.db.ReturnRequest.Query().Where(
		entrr.IDEQ(id), entrr.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, ErrNotFound // 含 RLS 擋下
	}
	if req.Status != "pending" {
		return nil, ErrNotPending
	}
	// 權限:主責業務或 dept_admin 以上(規格 §6.6)
	if err := u.checkReviewer(ctx, actor, req); err != nil {
		return nil, err
	}

	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	upd := tx.ReturnRequest.Update().
		Where(entrr.IDEQ(id), entrr.VersionEQ(expectedVersion)).
		SetStatus(decision).
		SetReviewedByUserID(actor.UserID).
		SetReviewedAt(time.Now()).
		SetVersion(expectedVersion + 1)
	if decision == "rejected" {
		upd.SetRejectReason(rejectReason)
	} else {
		upd.ClearRejectReason()
	}
	affected, err := upd.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if affected != 1 {
		_ = tx.Rollback()
		return nil, ErrVersionConflict // 前端重載後再審
	}

	if err := u.rec.Record(audit.ContextWithTx(ctx, tx), audit.Entry{
		ActorID: actor.UserID.String(), Action: "return_request.review",
		ResourceType: "return_request", ResourceID: id.String(),
		Before: fmt.Sprintf(`{"status":"pending"}`),
		After:  fmt.Sprintf(`{"status":%q,"reject_reason":%q}`, decision, rejectReason),
	}); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	// 通知建檔(D18):同一交易內為發起帳號建 fcm/in_app pending 記錄(僅 created_by_user_id,D23)
	var notifyIDs []uuid.UUID
	if u.sender != nil {
		ids, err := u.buildReviewNotifications(ctx, tx, req, decision, rejectReason)
		if err != nil {
			_ = tx.Rollback() // 建檔失敗同交易回滾
			return nil, err
		}
		notifyIDs = ids
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 提交後發送(D16:失敗僅標 failed,不回滾審核)
	if u.sender != nil && len(notifyIDs) > 0 {
		if err := notification.Dispatch(ctx, u.db, u.sender, u.rec, notifyIDs); err != nil {
			log.Printf("returns: dispatch notifications: %v", err)
		}
	}
	return tx.Client().ReturnRequest.Get(ctx, id)
}

// checkReviewer:staff 且(該客戶 default_sales_rep_id == actor.UserID 或 role 為 dept_admin/company_admin/developer)。
func (u *Usecase) checkReviewer(ctx context.Context, actor rls.Identity, req *ent.ReturnRequest) error {
	switch actor.Role {
	case "dept_admin", "company_admin", "developer":
		return nil
	}
	cust, err := u.db.Customer.Query().Where(entcust.IDEQ(req.CustomerID)).Only(ctx)
	if err != nil {
		return ErrNotFound
	}
	if cust.DefaultSalesRepID != nil && *cust.DefaultSalesRepID == actor.UserID {
		return nil
	}
	return ErrForbidden
}
```

`notify.go`(消費 07 計畫契約,不自定介面):

```go
package returns

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/notifications"
)

// buildReviewNotifications(細部 4.7.5):審核交易內為發起帳號建通知記錄。
// 範本 return_reviewed(fcm / in_app 兩通道;變數 request_id / decision / reject_reason / customer_name;
// 種子範本由 07 計畫 T2 範本資料落地)。範本缺失 → 降級為無範本記錄(同 07 計畫 T5 行為)。
func (u *Usecase) buildReviewNotifications(ctx context.Context, tx *ent.Tx, req *ent.ReturnRequest, decision, rejectReason string) ([]uuid.UUID, error) {
	cust, err := tx.Customer.Get(ctx, req.CustomerID)
	if err != nil {
		return nil, err
	}
	vars := map[string]string{
		"request_id": req.ID.String(), "decision": decision,
		"reject_reason": rejectReason, "customer_name": cust.Name,
	}
	repo := notifications.NewRepo()
	var ids []uuid.UUID
	for _, channel := range []string{"fcm", "in_app"} {
		q := notifications.TemplateQuery{
			CompanyID: req.CompanyID, DepartmentID: &req.DepartmentID,
			Code: "return_reviewed", Channel: channel, Locale: notifications.DefaultLocale,
		}
		var tmplID *uuid.UUID
		title, content := q.Code, fmt.Sprintf("%s %s", vars["customer_name"], vars["decision"])
		r, err := notifications.Render(ctx, tx.Client(), q, vars)
		switch {
		case err == nil:
			tmplID, title, content = &r.TemplateID, r.Title, r.Content
		case errors.Is(err, notifications.ErrTemplateNotFound):
			log.Printf("returns: template return_reviewed/%s missing, degrade", channel)
		default:
			return nil, err
		}
		id, err := repo.CreateOne(ctx, tx, notifications.CreateParams{
			CompanyID: req.CompanyID, DepartmentID: &req.DepartmentID,
			UserID: req.CreatedByUserID, // 僅發起帳號(D23)
			TemplateID: tmplID, Channel: channel,
			Title: title, Content: content,
			Payload: map[string]any{"return_request_id": req.ID.String()},
		})
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
```

FCM token 失效刪 `user_devices`、`failed` 標記皆在 `notification.Dispatch` 內(07 計畫 T4),本計畫不重複。07 未就緒前降級:`sender` 傳 nil 即完全不發通知(Review 功能不受影響)。

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go test ./internal/domain/returns/ -v -race`
Expected: PASS — 含併發樂觀鎖 1 成功 1 衝突、原訂單零異動。

```bash
git add backend/internal/domain/returns/review.go backend/internal/domain/returns/notify.go
git commit -m "feat(backend): 退貨審核(樂觀鎖+權限+稽核同事務),不修改原訂單(4.7.3/4.7.5)"
```

---

### Task 4: 退貨證明資料輸出(細部 4.7.4)

**Files:**
- Create: `backend/internal/domain/returns/certificate.go`
- Test: `backend/internal/domain/returns/certificate_test.go`

**Interfaces:**
- Produces: `Usecase.GetCertificate(ctx, actor, id) (*Certificate, error)`;`Certificate{RequestID, CustomerName, CustomerCode, CreatedAt, Items[]{ProductName, Unit, Quantity, Reason, PhotoURLs[]}, Decision, ReviewerName, ReviewedAt, Status}`。唯讀,不寫稽核;不產 PDF(Phase 5 不含退貨證明)。

- [ ] **Step 1: 寫失敗測試(approved 可取、快照隔離、非 approved 拒絕、跨客戶 RLS)**

```go
// backend/internal/domain/returns/certificate_test.go
package returns_test

func TestCertificateApprovedOnly(t *testing.T) {
	c := testutil.NewEntClient(t)
	f := seedFixtures(t, c)
	ctx := context.Background()
	uc := returns.NewUsecase(c, &audit.FakeRecorder{}, nil)
	req := createPending(t, uc, f)

	// pending → failed_precondition
	if _, err := uc.GetCertificate(ctx, f.custActor, req.ID); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("pending cert: %v", err)
	}
	if _, err := uc.Review(ctx, f.staffActor, req.ID, "approved", "", req.Version); err != nil {
		t.Fatal(err)
	}
	cert, err := uc.GetCertificate(ctx, f.custActor, req.ID)
	if err != nil { t.Fatalf("cert: %v", err) }
	if cert.CustomerCode != "AA000001" || len(cert.Items) != 1 || len(cert.Items[0].PhotoURLs) != 1 {
		t.Fatalf("cert: %+v", cert)
	}
	if cert.Items[0].PhotoURLs[0] != "/api/v1/files/"+f.photoID.String()+"/download" {
		t.Fatalf("photo url: %s", cert.Items[0].PhotoURLs[0])
	}
}

func TestCertificateSnapshotIsolation(t *testing.T) {
	// 審核通過後把來源商品改名 → GetCertificate 仍回申請時快照名
}

func TestCertificateAccess(t *testing.T) {
	// 乙的子帳號取甲的證明 → not_found(RLS);主帳號 → permission_denied;
	// 主責業務可取(業務側查證);非主責 staff → permission_denied
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/returns/ -run TestCertificate -v`
Expected: FAIL。

- [ ] **Step 3: 實作 certificate.go**

```go
package returns

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entrr "github.com/salesorder/sales-order-1.0/backend/ent/returnrequest"
	entrri "github.com/salesorder/sales-order-1.0/backend/ent/returnrequestitem"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/shopspring/decimal"
)

type CertItem struct {
	ProductName string
	Unit        string
	Quantity    decimal.Decimal
	Reason      string
	PhotoURLs   []string
}

type Certificate struct {
	RequestID    uuid.UUID
	CustomerName string
	CustomerCode string
	CreatedAt    string
	Items        []CertItem
	Decision     string
	ReviewerName string
	ReviewedAt   string
	Status       string
}

// GetCertificate(細部 4.7.4):僅 approved;內容一律快照欄位,不 join 商品主檔。
func (u *Usecase) GetCertificate(ctx context.Context, actor rls.Identity, id uuid.UUID) (*Certificate, error) {
	if actor.Role == "customer" && actor.IsPrimary {
		return nil, ErrForbidden
	}
	req, err := u.db.ReturnRequest.Query().Where(
		entrr.IDEQ(id), entrr.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, ErrNotFound // 跨客戶被 RLS 擋下同此
	}
	// 權限:該客戶子帳號(self)或具審核權限的 staff
	if actor.Role == "customer" {
		if actor.CustomerID == nil || *actor.CustomerID != req.CustomerID {
			return nil, ErrNotFound
		}
	} else if err := u.checkReviewer(ctx, actor, req); err != nil {
		return nil, err
	}
	if req.Status != "approved" {
		return nil, ErrNotApproved
	}

	items, err := u.db.ReturnRequestItem.Query().Where(
		entrri.ReturnRequestIDEQ(id), entrri.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return nil, err
	}
	cust, err := u.db.Customer.Get(ctx, req.CustomerID)
	if err != nil {
		return nil, err
	}
	reviewerName := ""
	if req.ReviewedByUserID != nil {
		if u2, err := u.db.User.Get(ctx, *req.ReviewedByUserID); err == nil {
			reviewerName = u2.Name
		}
	}

	cert := &Certificate{
		RequestID: req.ID, CustomerName: cust.Name, CustomerCode: cust.CustomerCode,
		CreatedAt: req.CreatedAt.Format("2006-01-02 15:04"), Decision: req.Status,
		ReviewerName: reviewerName, Status: req.Status,
	}
	if req.ReviewedAt != nil {
		cert.ReviewedAt = req.ReviewedAt.Format("2006-01-02 15:04")
	}
	for _, it := range items {
		ci := CertItem{
			ProductName: it.ProductName, Unit: it.Unit,
			Quantity: it.Quantity, Reason: it.Reason,
		}
		for _, fid := range it.PhotoFileIds {
			ci.PhotoURLs = append(ci.PhotoURLs, fmt.Sprintf("/api/v1/files/%s/download", fid))
		}
		cert.Items = append(cert.Items, ci)
	}
	return cert, nil
}
```

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go test ./internal/domain/returns/ -v`
Expected: PASS。`go test ./...` 全綠。

```bash
git add backend/internal/domain/returns/certificate.go
git commit -m "feat(backend): 退貨證明資料輸出(快照隔離+照片 URL)(4.7.4)"
```

---

## Self-Review 記錄

- **Spec 覆蓋**:4.7.1→T1;4.7.2→T2;4.7.3→T3;4.7.5→T2(create 稽核)/ T3(review 稽核+推播掛點);4.7.4→T4。無缺漏。
- **整合測試重點對應**(細部文件文末 9 條):狀態機/樂觀鎖併發→T3 測試;權限矩陣→T3 `TestReviewPermissions`;RLS 隔離→T2 `TestListSelfScope` + T4 `TestCertificateAccess`;不修改原訂單→T3 `TestReviewDoesNotTouchOrder`;稽核同事務→T2/T3 各含失敗注入斷言路徑;推播路由→T3 `TestReviewApproveBySalesRep` 斷言僅發起帳號 fcm/in_app 各一筆;雙來源→T2 兩測試;快照隔離→T4 `TestCertificateSnapshotIsolation`。
- **跨計畫契約**:通知直接消費 07 計畫(`notifications.Render` + `Repo.CreateOne` 交易內建檔、`notification.Dispatch` 提交後發送、`notification.Sender`/`FakeSender` 注入)——不自定 Notifier 介面;範本 `return_reviewed` 種子資料列入 07 計畫 T2;`audit.FakeRecorder` 為測試 fake(01 計畫 T14 提供 `audit.NoopRecorder`;03 計畫 T5 提供 PG 實作);`customer_products` 別名快照用 `AliasName`,純手打訂單品項(`product_id=NULL`)不得作為退貨來源(4.7.2 錯誤處理,回 `invalid_argument`)。
- **類型一致**:`Quantity` 用 `decimal.Decimal`(04/05 計畫契約);`ReturnRequestItem` 快照欄位 `ProductName` / `Unit` 命名與 T1 schema 一致(T1 由本計畫定義,實作時以細部 4.7.1 欄位表為準)。
- **照片上限**:`maxPhotosPerItem = 5` 定為 1.0 常數(細部文件標「待確認」,此處拍板並記錄)。

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/backend/2026-08-17-backend-06-returns-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** — 每個 Task 派新 subagent 執行,Task 間 review,迭代快。

**2. Inline Execution** — 用 executing-plans 在本 session 逐批執行,設 checkpoint review。

Which approach?

---

*計畫版本:v1.0.0(2026-08-17);對應細部文件 `detail/06-returns.md`、原計畫 v2.9.0、規格書 v1.0.34。*
