# Backend 03 — 字典檔與稽核日誌 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 實作 metadicts 字典檔(系統預設 + 部門擴充兩層)與 audit_logs 稽核日誌(schema、usecase 層統一同事務寫入機制、查詢 API)。

**Architecture:** 依細部文件 `detail/03-metadicts-audit.md` 實作原計畫 Task 2.5 / 2.6。字典檔單表 `metadicts` 以 `department_id IS NULL` 表達系統預設(D11);稽核以 `audit.Recorder` 介面(01-auth Task 14 已定義)的 DB 實作落地,recorder 僅接受交易內 Ent client,從簽名杜絕脫鉤寫入(D18)。

**Tech Stack:** Go 1.25、Ent、Chi、Connect-RPC、pgx/v5、testcontainers-go(沿用 01-auth 計畫 Task 1 的 `testutil.NewEntClient`)。

**Spec 來源:** 細部文件 `docs/superpowers/plans/backend/detail/03-metadicts-audit.md`;共通規則見 `00-index.md` §3。

## Global Constraints

- module 路徑:`github.com/salesorder/sales-order-1.0/backend`;所有路徑相對 repo root。
- `metadicts` 軟刪除 `deleted_at` + 部分唯一索引(D10);`audit_logs` **不建** `deleted_at` / `updated_at`(不可變,D27 保留期限由 Phase 8 維運處理)。
- 唯一性:`(type, code, department_id)` 以兩道部分唯一索引實作(系統預設 `WHERE department_id IS NULL AND deleted_at IS NULL`;部門擴充 `WHERE department_id IS NOT NULL AND deleted_at IS NULL`)。
- 稽核寫入失敗 = 整體交易回滾;不得降級略過(D18);snapshot 永不收錄敏感欄位(密碼雜湊、token)。
- 錯誤統一 Connect code;列表分頁 `meta` 格式,`per_page` ≤ 100;稽核查詢預設時間窗 = 近 3 個月(D27)。
- 字典類型 `type` 合法值:`unit` / `payment_method` / `settlement_method` / `customer_type` / `invoice_type` / `order_source`;`order_source` 為系統固定,API 不可異動。
- RLS 注入值沿用 01-auth:`app.current_company_id` / `app.current_department_id` / `app.current_data_scope`。
- 測試:DB 相依走 testcontainers;每 Task 結尾 commit(`feat(backend): …`)。

## File Structure

| 檔案 | 職責 | 建立於 |
|---|---|---|
| `backend/ent/schema/metadict.go` | 字典檔 schema | Task 1 |
| `backend/database/migrations/00008_metadicts.sql` | 部分唯一索引、RLS、系統預設 seed | Task 1 |
| `backend/internal/domain/metadicts/{service,usecase,repository}.go` | 字典 RPC / 業務 / 查詢 | Task 2-3 |
| `backend/proto/v1/metadict.proto` | MetadictService | Task 2 |
| `backend/ent/schema/auditlog.go` | 稽核表 schema | Task 4 |
| `backend/database/migrations/00009_audit_logs.sql` | 索引、RLS、寫入權限收緊 | Task 4 |
| `backend/internal/audit/pg_recorder.go` | `audit.Recorder` 的 DB 實作(接管 01-auth NoopRecorder) | Task 5 |
| `backend/internal/domain/auditlogs/query.go` | 稽核查詢 API | Task 6 |
| `backend/proto/v1/audit.proto` | AuditService | Task 6 |

---

### Task 1: metadicts schema、migration、系統預設 seed(細部 2.5.1)

**Files:**
- Create: `backend/ent/schema/metadict.go`
- Create: `backend/database/migrations/00008_metadicts.sql`
- Test: `backend/internal/domain/metadicts/metadict_schema_test.go`

**Interfaces:**
- Consumes: 01-auth Task 1 的 `companies` / `departments`;`testutil.NewEntClientWithRLS`(RLS 包裝)。
- Produces: `ent.Metadict`(欄位 `id`、`type`、`code`、`display_name`、`department_id *uuid.UUID`、`sort_order int`、`is_active bool`、`created_at`、`updated_at`、`deleted_at *time.Time`);合法 type 常數 `metadicts.TypeUnit` 等(見 Task 2);seed 後 `order_source` 固定值 `W`(Web)/`A`(App)可查,供 05-sales-orders 取號使用。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/domain/metadicts/metadict_schema_test.go`:

```go
package metadicts_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestMetadictPartialUniqueIndex(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	dep := newDept(t, c, "業務部") // 見 Step 3 輔助

	mk := func(depID *uuid.UUID, code string) error {
		b := c.Metadict.Create().SetType("unit").SetCode(code).SetDisplayName(code)
		if depID != nil {
			b.SetDepartmentID(*depID)
		}
		_, err := b.Save(ctx)
		return err
	}

	// 系統預設與部門擴充同 code 可並存
	if err := mk(nil, "KG"); err != nil {
		t.Fatalf("system row: %v", err)
	}
	if err := mk(&dep, "KG"); err != nil {
		t.Fatalf("dept row same code should coexist: %v", err)
	}
	// 同範圍重複被拒
	if err := mk(nil, "KG"); err == nil {
		t.Fatal("duplicate system code should fail")
	}
	if err := mk(&dep, "KG"); err == nil {
		t.Fatal("duplicate dept code should fail")
	}
	// 軟刪除後同範圍可立即重建(部分唯一索引生效)
	row, err := c.Metadict.Query().Where().First(ctx) // type=unit code=KG department NULL
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Metadict.UpdateOneID(row.ID).SetDeletedAt(testutil.Now()).Save(ctx); err != nil {
		t.Fatal(err)
	}
	if err := mk(nil, "KG"); err != nil {
		t.Fatalf("recreate after soft delete: %v", err)
	}
}

func TestMetadictSeedIdempotent(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	if err := applyMetadictMigration(t, c); err != nil { // 見 Step 3:跑 00004 SQL 的 seed 段
		t.Fatal(err)
	}
	n1, err := c.Metadict.Query().Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n1 == 0 {
		t.Fatal("seed should insert system dictionaries")
	}
	if err := applyMetadictMigration(t, c); err != nil { // 重跑冪等
		t.Fatalf("re-seed: %v", err)
	}
	if n2, _ := c.Metadict.Query().Count(ctx); n2 != n1 {
		t.Fatalf("seed not idempotent: %d -> %d", n1, n2)
	}
	// order_source 固定值存在
	exists, err := c.Metadict.Query().Where().Exist(ctx) // type=order_source code=W
	if err != nil || !exists {
		t.Fatalf("order_source W missing: %v", err)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/metadicts/ -v`
Expected: FAIL — `ent.Metadict`、`newDept`、`testutil.Now`、`applyMetadictMigration` 未定義(編譯失敗)。

- [ ] **Step 3: 實作 schema**

`backend/ent/schema/metadict.go`:

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

// Metadict 字典檔:department_id IS NULL = 系統預設,非 NULL = 部門擴充(D11,細部 2.5.1)。
type Metadict struct{ ent.Schema }

func (Metadict) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("type").NotEmpty(),
		field.String("code").NotEmpty().MaxLen(32),
		field.String("display_name").NotEmpty(),
		field.UUID("department_id", uuid.UUID{}).Optional().Nillable(),
		field.Int("sort_order").Default(0),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (Metadict) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("department", Department.Type).Ref("metadicts").
			Field("department_id").Unique(),
	}
}

func (Metadict) Indexes() []ent.Index {
	// 部分唯一索引在 migration SQL 建立(Ent annotations 對雙軌 NULL 語意支援有限);
	// 這裡只建查詢索引。
	return []ent.Index{
		index.Fields("type", "department_id", "is_active"),
	}
}
```

`backend/database/migrations/00008_metadicts.sql`:

```sql
-- +goose Up
-- 細部 2.5.1:雙軌部分唯一索引 + RLS + 系統預設 seed(冪等)。

CREATE UNIQUE INDEX metadicts_system_type_code_unique
  ON metadicts (type, code)
  WHERE department_id IS NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX metadicts_dept_type_code_unique
  ON metadicts (type, code, department_id)
  WHERE department_id IS NOT NULL AND deleted_at IS NULL;

ALTER TABLE metadicts ENABLE ROW LEVEL SECURITY;

-- 讀取:系統預設或當前部門;all 全見
CREATE POLICY metadicts_read ON metadicts FOR SELECT
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR department_id IS NULL
         OR department_id::text = current_setting('app.current_department_id', true));

-- 寫入:all(系統級)或自己部門(擴充)
CREATE POLICY metadicts_write ON metadicts FOR ALL
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR department_id::text = current_setting('app.current_department_id', true));

-- 系統預設 seed(冪等:ON CONFLICT DO NOTHING 依賴上方部分唯一索引)
-- 單位
INSERT INTO metadicts (id, type, code, display_name, sort_order, is_active, created_at, updated_at)
SELECT gen_random_uuid(), 'unit', v.code, v.name, v.ord, true, now(), now()
FROM (VALUES ('KG','公斤',1),('G','公克',2),('PACK','包',3),('BOX','箱',4),('PIECE','件',5)) AS v(code,name,ord)
ON CONFLICT DO NOTHING;
-- 付款方式 / 結帳方式 / 客戶類型 / 發票類型(各 seed 基礎值,細節依規格 §5.3)
INSERT INTO metadicts (id, type, code, display_name, sort_order, is_active, created_at, updated_at)
SELECT gen_random_uuid(), 'payment_method', v.code, v.name, v.ord, true, now(), now()
FROM (VALUES ('CASH','現金',1),('TRANSFER','轉帳',2),('CREDIT','月結',3)) AS v(code,name,ord)
ON CONFLICT DO NOTHING;
-- 訂單來源(系統固定,API 不可異動)
INSERT INTO metadicts (id, type, code, display_name, sort_order, is_active, created_at, updated_at)
SELECT gen_random_uuid(), 'order_source', v.code, v.name, v.ord, true, now(), now()
FROM (VALUES ('W','Web 中台',1),('A','App',2)) AS v(code,name,ord)
ON CONFLICT DO NOTHING;

-- +goose Down
DROP POLICY IF EXISTS metadicts_write ON metadicts;
DROP POLICY IF EXISTS metadicts_read ON metadicts;
ALTER TABLE metadicts DISABLE ROW LEVEL SECURITY;
DROP INDEX IF EXISTS metadicts_dept_type_code_unique;
DROP INDEX IF EXISTS metadicts_system_type_code_unique;
DELETE FROM metadicts WHERE department_id IS NULL;
```

測試輔助(放入 `backend/internal/domain/metadicts/helpers_test.go`):

```go
package metadicts_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func newDept(t *testing.T, c *ent.Client, name string) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	co, err := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	if err != nil { t.Fatal(err) }
	d, err := c.Department.Create().SetCompanyID(co.ID).SetName(name).Save(ctx)
	if err != nil { t.Fatal(err) }
	return d.ID
}

// applyMetadictMigration 讀 00004 SQL 的 seed 段(INSERT ... ON CONFLICT DO NOTHING)執行。
func applyMetadictMigration(t *testing.T, c *ent.Client) error {
	t.Helper()
	raw, err := os.ReadFile("../../../database/migrations/00008_metadicts.sql")
	if err != nil { return err }
	up := strings.SplitN(string(raw), "-- +goose Down", 2)[0]
	_, err = c.DB().ExecContext(context.Background(), up)
	return err
}
```

`testutil` 補 `func Now() time.Time { return time.Now() }`。

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && go generate ./ent/ && go test ./internal/domain/metadicts/ -v`
Expected: PASS — 雙軌唯一、軟刪後重建、seed 冪等、order_source 存在。

- [ ] **Step 5: Commit**

```bash
git add backend/ent/schema/metadict.go backend/database/migrations/00008_metadicts.sql backend/internal/domain/metadicts backend/internal/testutil
git commit -m "feat(backend): metadicts schema、雙軌部分唯一索引、系統預設 seed(2.5.1)"
```

---

### Task 2: MetadictService CRUD 與軟刪除(細部 2.5.2)

**Files:**
- Create: `backend/proto/v1/metadict.proto`
- Create: `backend/internal/domain/metadicts/service.go`
- Create: `backend/internal/domain/metadicts/usecase.go`
- Create: `backend/internal/domain/metadicts/repository.go`
- Test: `backend/internal/domain/metadicts/service_test.go`

**Interfaces:**
- Consumes: Task 1 schema;`rls.Identity`(01-auth);`audit.Recorder`(01-auth Task 14;Task 5 換 DB 實作,本 Task 以 fake recorder 測試同事務語義)。
- Produces: Connect-RPC `MetadictService`:`List({type, include_inactive, include_deleted, page, per_page})` / `Get({id})` / `Create({type, code, display_name, sort_order, is_active})` / `Update({id, display_name, sort_order, is_active})` / `Delete({id})`;usecase `metadicts.Usecase{Create/Update/Delete/...}`;合法 type 常數:

```go
const (
	TypeUnit              = "unit"
	TypePaymentMethod     = "payment_method"
	TypeSettlementMethod  = "settlement_method"
	TypeCustomerType      = "customer_type"
	TypeInvoiceType       = "invoice_type"
	TypeOrderSource       = "order_source" // 系統固定,API 不可異動
)
```

- [ ] **Step 1: 寫失敗測試(範圍推導、不可變欄位、同事務稽核)**

`backend/internal/domain/metadicts/service_test.go`:

```go
package metadicts_test

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/metadicts"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

// fakeRecorder 記錄呼叫;fail=true 時模擬稽核寫入失敗(驗證回滾)。
type fakeRecorder struct {
	entries []audit.Entry
	fail    bool
}

func (f *fakeRecorder) Record(_ context.Context, e audit.Entry) error {
	if f.fail {
		return errors.New("injected audit failure")
	}
	f.entries = append(f.entries, e)
	return nil
}

func superIdentity(coID uuid.UUID) rls.Identity {
	return rls.Identity{UserID: uuid.New(), CompanyID: coID, DataScope: "all", Role: "super"}
}

func deptAdminIdentity(coID, depID uuid.UUID) rls.Identity {
	return rls.Identity{UserID: uuid.New(), CompanyID: coID, DepartmentID: &depID, DataScope: "department", Role: "dept_admin"}
}

func TestCreateScopeDerivation(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	dep, _ := c.Department.Create().SetCompanyID(co.ID).SetName("業務部").Save(ctx)
	rec := &fakeRecorder{}
	uc := metadicts.NewUsecase(c, rec)

	// super → 系統級(department_id NULL)
	m, err := uc.Create(ctx, superIdentity(co.ID), metadicts.CreateInput{
		Type: metadicts.TypeUnit, Code: "BAG", DisplayName: "袋",
	})
	if err != nil { t.Fatalf("super create: %v", err) }
	if m.DepartmentID != nil {
		t.Fatal("super create must be system-level (NULL department)")
	}

	// dept_admin → 強制帶入自己部門
	m2, err := uc.Create(ctx, deptAdminIdentity(co.ID, dep.ID), metadicts.CreateInput{
		Type: metadicts.TypeUnit, Code: "BAG", DisplayName: "袋(業務部)",
	})
	if err != nil { t.Fatalf("dept create: %v", err) }
	if m2.DepartmentID == nil || *m2.DepartmentID != dep.ID {
		t.Fatal("dept_admin create must carry own department_id")
	}

	// 稽核:兩次 create 各一筆 action=create
	if len(rec.entries) != 2 || rec.entries[0].Action != "create" {
		t.Fatalf("audit missing: %+v", rec.entries)
	}
}

func TestUpdateRejectsImmutableFields(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	rec := &fakeRecorder{}
	uc := metadicts.NewUsecase(c, rec)
	m, _ := uc.Create(ctx, superIdentity(co.ID), metadicts.CreateInput{
		Type: metadicts.TypeUnit, Code: "BAG", DisplayName: "袋",
	})

	// code 不可改
	if err := uc.Update(ctx, superIdentity(co.ID), m.ID, metadicts.UpdateInput{
		Code: ptr("BOX"),
	}); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("code change: want failed_precondition, got %v", err)
	}
	// 系統固定字典不可異動
	src, _ := uc.Create(ctx, superIdentity(co.ID), metadicts.CreateInput{
		Type: metadicts.TypeOrderSource, Code: "POS", DisplayName: "POS",
	})
	if err := uc.Update(ctx, superIdentity(co.ID), src.ID, metadicts.UpdateInput{
		DisplayName: ptr("改"),
	}); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("order_source update: want failed_precondition, got %v", err)
	}
	if err := uc.Delete(ctx, superIdentity(co.ID), src.ID); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("order_source delete: want failed_precondition, got %v", err)
	}
}

func TestAuditFailureRollsBack(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	rec := &fakeRecorder{fail: true}
	uc := metadicts.NewUsecase(c, rec)

	_, err := uc.Create(ctx, superIdentity(co.ID), metadicts.CreateInput{
		Type: metadicts.TypeUnit, Code: "RB", DisplayName: "回滾驗證",
	})
	if connect.CodeOf(err) != connect.CodeInternal {
		t.Fatalf("audit failure: want internal, got %v", err)
	}
	// 業務異動不存在(回滾)
	exists, _ := c.Metadict.Query().Where().Exist(ctx) // code=RB
	if exists {
		t.Fatal("business row must not exist when audit write fails (D18)")
	}
}

func TestSoftDeleteAndRecreate(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	rec := &fakeRecorder{}
	uc := metadicts.NewUsecase(c, rec)
	id := superIdentity(co.ID)

	m, _ := uc.Create(ctx, id, metadicts.CreateInput{Type: metadicts.TypeUnit, Code: "DEL", DisplayName: "刪"})
	if err := uc.Delete(ctx, id, m.ID); err != nil { t.Fatalf("delete: %v", err) }
	// Get 回 not_found
	if _, err := uc.Get(ctx, id, m.ID); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("get deleted: want not_found, got %v", err)
	}
	// 同 code 可重建
	if _, err := uc.Create(ctx, id, metadicts.CreateInput{Type: metadicts.TypeUnit, Code: "DEL", DisplayName: "刪2"}); err != nil {
		t.Fatalf("recreate: %v", err)
	}
	// 稽核有 delete
	var actions []string
	for _, e := range rec.entries { actions = append(actions, e.Action) }
	if !contains(actions, "create") || !contains(actions, "delete") {
		t.Fatalf("audit actions: %v", actions)
	}
}

func ptr[T any](v T) *T { return &v }
func contains(s []string, v string) bool {
	for _, x := range s { if x == v { return true } }
	return false
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/metadicts/ -run 'TestCreate|TestUpdate|TestAudit|TestSoftDelete' -v`
Expected: FAIL — `metadicts.NewUsecase` / `CreateInput` 未定義。

- [ ] **Step 3: 實作 proto**

`backend/proto/v1/metadict.proto`:

```proto
syntax = "proto3";
package salesorder.v1;

import "google/protobuf/empty.proto";

service MetadictService {
  rpc List(ListMetadictsRequest) returns (ListMetadictsResponse);
  rpc Get(GetMetadictRequest) returns (Metadict);
  rpc Create(CreateMetadictRequest) returns (Metadict);
  rpc Update(UpdateMetadictRequest) returns (Metadict);
  rpc Delete(DeleteMetadictRequest) returns (google.protobuf.Empty);
  rpc ListOptions(ListOptionsRequest) returns (ListOptionsResponse); // Task 3
}

message Metadict {
  string id = 1; string type = 2; string code = 3; string display_name = 4;
  optional string department_id = 5; int32 sort_order = 6; bool is_active = 7;
  string created_at = 8; string updated_at = 9; // RFC3339
}
message PageMeta { int32 page = 1; int32 per_page = 2; int32 total = 3; }

message ListMetadictsRequest {
  optional string type = 1; bool include_inactive = 2; bool include_deleted = 3;
  int32 page = 4; int32 per_page = 5;
}
message ListMetadictsResponse { repeated Metadict items = 1; PageMeta meta = 2; }
message GetMetadictRequest { string id = 1; }
message CreateMetadictRequest {
  string type = 1; string code = 2; string display_name = 3;
  int32 sort_order = 4; bool is_active = 5;
}
message UpdateMetadictRequest {
  string id = 1;
  optional string code = 2;         // 提供即拒絕(failed_precondition)
  optional string display_name = 3;
  optional int32 sort_order = 4;
  optional bool is_active = 5;
}
message DeleteMetadictRequest { string id = 1; }
message ListOptionsRequest { string type = 1; optional string keyword = 2; }
message ListOptionsResponse {
  repeated Option options = 1;
  message Option { string code = 1; string display_name = 2; }
}
```

- [ ] **Step 4: 實作 usecase**

`backend/internal/domain/metadicts/usecase.go`:

```go
// Package metadicts 字典檔 domain(細部文件 2.5)。
package metadicts

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entmd "github.com/salesorder/sales-order-1.0/backend/ent/metadict"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// 合法字典類型(Global Constraints)。
const (
	TypeUnit             = "unit"
	TypePaymentMethod    = "payment_method"
	TypeSettlementMethod = "settlement_method"
	TypeCustomerType     = "customer_type"
	TypeInvoiceType      = "invoice_type"
	TypeOrderSource      = "order_source" // 系統固定,API 不可異動
)

var validTypes = map[string]bool{
	TypeUnit: true, TypePaymentMethod: true, TypeSettlementMethod: true,
	TypeCustomerType: true, TypeInvoiceType: true, TypeOrderSource: true,
}

// CreateInput / UpdateInput 為 RPC 層轉來的輸入。
type CreateInput struct {
	Type, Code, DisplayName string
	SortOrder               int
	IsActive                bool
}

type UpdateInput struct {
	Code        *string // 提供即拒絕
	DisplayName *string
	SortOrder   *int
	IsActive    *bool
}

type Usecase struct {
	db  *ent.Client
	rec audit.Recorder
}

func NewUsecase(db *ent.Client, rec audit.Recorder) *Usecase {
	return &Usecase{db: db, rec: rec}
}

// Create:super → 系統級(NULL department);dept_admin/staff → 強制帶入自己部門(細部 2.5.2 步驟 1)。
func (u *Usecase) Create(ctx context.Context, actor rls.Identity, in CreateInput) (*ent.Metadict, error) {
	if !validTypes[in.Type] || in.Code == "" || in.DisplayName == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid type/code/display_name"))
	}
	var deptID *uuid.UUID
	switch actor.DataScope {
	case "all":
		// 系統級
	case "department":
		if actor.DepartmentID == nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("department context missing"))
		}
		deptID = actor.DepartmentID
	default: // company / self(客戶)不可寫字典
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("cannot create metadict"))
	}

	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	b := tx.Metadict.Create().
		SetType(in.Type).SetCode(in.Code).SetDisplayName(in.DisplayName).
		SetSortOrder(in.SortOrder).SetIsActive(in.IsActive)
	if deptID != nil {
		b.SetDepartmentID(*deptID)
	}
	m, err := b.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		if isUniqueViolation(err) {
			return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("code exists in scope"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 稽核同事務(D18);失敗 → 回滾
	if err := u.rec.Record(withTx(ctx, tx), audit.Entry{
		ActorID: actor.UserID.String(), Action: "create",
		ResourceType: "metadict", ResourceID: m.ID.String(),
		After: snapshotOf(m),
	}); err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	invalidateOptionsCache(in.Type) // Task 3 快取失效
	return m, nil
}

// Update:僅 display_name / sort_order / is_active;code 提供即拒;order_source 拒絕(細部 2.5.2 步驟 2-3)。
func (u *Usecase) Update(ctx context.Context, actor rls.Identity, id uuid.UUID, in UpdateInput) error {
	if in.Code != nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("code is immutable"))
	}
	m, err := u.writable(ctx, actor, id)
	if err != nil {
		return err
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	before := snapshotOf(m)
	upd := tx.Metadict.UpdateOneID(id)
	if in.DisplayName != nil {
		upd.SetDisplayName(*in.DisplayName)
	}
	if in.SortOrder != nil {
		upd.SetSortOrder(*in.SortOrder)
	}
	if in.IsActive != nil {
		upd.SetIsActive(*in.IsActive)
	}
	m2, err := upd.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	if err := u.rec.Record(withTx(ctx, tx), audit.Entry{
		ActorID: actor.UserID.String(), Action: "update",
		ResourceType: "metadict", ResourceID: id.String(),
		Before: before, After: snapshotOf(m2),
	}); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	if err := tx.Commit(); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	invalidateOptionsCache(m.Type)
	return nil
}

// Delete = 軟刪除;復原不在 1.0 API(細部 2.5.2 步驟 5)。
func (u *Usecase) Delete(ctx context.Context, actor rls.Identity, id uuid.UUID) error {
	m, err := u.writable(ctx, actor, id)
	if err != nil {
		return err
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	before := snapshotOf(m)
	if _, err := tx.Metadict.UpdateOneID(id).SetDeletedAt(timeNow()).Save(ctx); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	if err := u.rec.Record(withTx(ctx, tx), audit.Entry{
		ActorID: actor.UserID.String(), Action: "delete",
		ResourceType: "metadict", ResourceID: id.String(), Before: before,
	}); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	if err := tx.Commit(); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	invalidateOptionsCache(m.Type)
	return nil
}

// writable:取列 + 寫入權限檢查(系統級僅 all;部門列僅同部門 department;order_source 拒絕異動)。
func (u *Usecase) writable(ctx context.Context, actor rls.Identity, id uuid.UUID) (*ent.Metadict, error) {
	m, err := u.db.Metadict.Query().Where(entmd.IDEQ(id), entmd.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("metadict not found"))
	}
	if m.Type == TypeOrderSource {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("order_source is system-fixed"))
	}
	if m.DepartmentID == nil {
		if actor.DataScope != "all" {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("system-level row"))
		}
		return m, nil
	}
	if actor.DataScope != "department" || actor.DepartmentID == nil || *actor.DepartmentID != *m.DepartmentID {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("out of department scope"))
	}
	return m, nil
}
```

`repository.go` / helpers(`isUniqueViolation`、`snapshotOf`、`withTx`、`timeNow`、`invalidateOptionsCache` 佔位)於同檔或 `repository.go` 落地;`withTx(ctx, tx)` 把 `*ent.Tx` 放入 ctx 供 recorder 使用(Task 5 的 DB recorder 從 ctx 取 tx,確保同一交易);`invalidateOptionsCache` 本 Task 先為 no-op,Task 3 接真實快取。`snapshotOf` 只收錄非敏感顯示欄位:

```go
func snapshotOf(m *ent.Metadict) string {
	b, _ := json.Marshal(map[string]any{
		"type": m.Type, "code": m.Code, "display_name": m.DisplayName,
		"department_id": m.DepartmentID, "sort_order": m.SortOrder, "is_active": m.IsActive,
	})
	return string(b)
}
```

`service.go`(Connect handler)為薄層:解析 proto → 呼叫 usecase → 映射 Connect code(usecase 已回 Connect error,直接回傳);`Get` 實作 `writable` 的讀版(讀權限由 Task 3 合併查詢邏輯處理)。

- [ ] **Step 5: 跑測試確認通過**

Run: `cd backend && go test ./internal/domain/metadicts/ -v`
Expected: PASS — 範圍推導、不可變欄位、稽核回滾、軟刪重建。

- [ ] **Step 6: Commit**

```bash
git add backend/proto/v1/metadict.proto backend/internal/domain/metadicts
git commit -m "feat(backend): MetadictService CRUD、範圍推導、同事務稽核(2.5.2)"
```

---

### Task 3: 合併查詢與 ListOptions(細部 2.5.3–2.5.4)

**Files:**
- Update: `backend/internal/domain/metadicts/repository.go`
- Update: `backend/internal/domain/metadicts/service.go`
- Test: `backend/internal/domain/metadicts/query_test.go`

**Interfaces:**
- Consumes: Task 2 usecase。
- Produces: `metadicts.VisibleScope(actor rls.Identity) predicate`(固定可見範圍條件,供 List/Get/ListOptions 與其他 domain 複用);`Usecase.ListOptions(ctx, actor, type, keyword) ([]Option, error)`(`Option{Code, DisplayName string}`);`invalidateOptionsCache(type string)` 真實實作(記憶體快取,type 為 key,TTL 60 秒)。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/domain/metadicts/query_test.go`:

```go
package metadicts_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/metadicts"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestMergedVisibility(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	depA, _ := c.Department.Create().SetCompanyID(co.ID).SetName("A部").Save(ctx)
	depB, _ := c.Department.Create().SetCompanyID(co.ID).SetName("B部").Save(ctx)
	rec := &fakeRecorder{}
	uc := metadicts.NewUsecase(c, rec)
	super := superIdentity(co.ID)

	// 系統預設 + A 擴充 + B 擴充
	uc.Create(ctx, super, metadicts.CreateInput{Type: "unit", Code: "SYS", DisplayName: "系統值"})
	uc.Create(ctx, deptAdminIdentity(co.ID, depA.ID), metadicts.CreateInput{Type: "unit", Code: "DA", DisplayName: "A擴充"})
	uc.Create(ctx, deptAdminIdentity(co.ID, depB.ID), metadicts.CreateInput{Type: "unit", Code: "DB", DisplayName: "B擴充"})

	// A 部門視角:系統 + A,看不見 B
	items, err := uc.List(ctx, deptAdminIdentity(co.ID, depA.ID), metadicts.ListInput{Type: ptr("unit")})
	if err != nil { t.Fatal(err) }
	codes := map[string]bool{}
	for _, m := range items { codes[m.Code] = true }
	if !codes["SYS"] || !codes["DA"] || codes["DB"] {
		t.Fatalf("merged visibility wrong: %v", codes)
	}

	// 無部門角色(company_admin):僅系統預設,不報錯
	ca := rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DataScope: "company", Role: "company_admin"}
	items, err = uc.List(ctx, ca, metadicts.ListInput{Type: ptr("unit")})
	if err != nil { t.Fatal(err) }
	for _, m := range items {
		if m.DepartmentID != nil {
			t.Fatalf("company scope should only see system rows, got %+v", m)
		}
	}
}

func TestListOptionsFiltersAndCacheInvalidation(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	rec := &fakeRecorder{}
	uc := metadicts.NewUsecase(c, rec)
	super := superIdentity(co.ID)

	m, _ := uc.Create(ctx, super, metadicts.CreateInput{Type: "unit", Code: "OPT", DisplayName: "選項", IsActive: true})

	// 客戶身分(self)僅見系統預設,不報錯(細部 2.5.4 步驟 5)
	cust := rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DataScope: "self", Role: "customer"}
	opts, err := uc.ListOptions(ctx, cust, "unit", "")
	if err != nil { t.Fatal(err) }
	found := false
	for _, o := range opts { if o.Code == "OPT" { found = true } }
	if !found { t.Fatalf("customer should see system option: %v", opts) }

	// 停用後立即消失(快取失效生效)
	if err := uc.Update(ctx, super, m.ID, metadicts.UpdateInput{IsActive: ptr(false)}); err != nil {
		t.Fatal(err)
	}
	opts, err = uc.ListOptions(ctx, cust, "unit", "")
	if err != nil { t.Fatal(err) }
	for _, o := range opts {
		if o.Code == "OPT" {
			t.Fatal("inactive option should disappear immediately (cache invalidated)")
		}
	}

	// 非法 type
	if _, err := uc.ListOptions(ctx, cust, "no-such-type", ""); err == nil {
		t.Fatal("invalid type should be invalid_argument")
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/metadicts/ -run 'TestMerged|TestListOptions' -v`
Expected: FAIL — `uc.List` / `ListInput` / `ListOptions` 未定義。

- [ ] **Step 3: 實作合併查詢與選項服務**

`repository.go` 加:

```go
// ListInput 為 List 的篩選。
type ListInput struct {
	Type           *string
	IncludeInactive bool
	IncludeDeleted  bool // 僅 super(細部 2.5.3 步驟 2)
	Page, PerPage   int
}

// VisibleScope:系統預設 OR 當前部門;company/all 無部門時僅系統預設(細部 2.5.3 步驟 1、3、6)。
func visibleScope(actor rls.Identity) []predicate.Metadict {
	switch {
	case actor.DataScope == "all":
		return nil // 管理視角全見(含各部門擴充;細部 2.5.3 步驟 3 的指定部門檢視由 ListInput 另行支援)
	case actor.DataScope == "department" && actor.DepartmentID != nil:
		return []predicate.Metadict{entmd.Or(
			entmd.DepartmentIDIsNil(),
			entmd.DepartmentIDEQ(*actor.DepartmentID),
		)}
	default: // company / self:僅系統預設
		return []predicate.Metadict{entmd.DepartmentIDIsNil()}
	}
}

// List 合併查詢;排序 sort_order、code 升冪(細部 2.5.3 步驟 4)。
func (u *Usecase) List(ctx context.Context, actor rls.Identity, in ListInput) ([]*ent.Metadict, error) {
	per := in.PerPage
	if per <= 0 {
		per = 20
	}
	if per > 100 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("per_page exceeds 100"))
	}
	q := u.db.Metadict.Query().Where(visibleScope(actor)...)
	if !in.IncludeDeleted {
		q = q.Where(entmd.DeletedAtIsNil())
	} else if actor.DataScope != "all" {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("include_deleted is super-only"))
	}
	if in.Type != nil {
		q = q.Where(entmd.TypeEQ(*in.Type))
	}
	if !in.IncludeInactive {
		q = q.Where(entmd.IsActiveEQ(true))
	}
	return q.Order(entmd.BySortOrder(), entmd.ByCode()).
		Offset(max(in.Page-1, 0) * per).Limit(per).All(ctx)
}

// ListOptions:可選用選項 = 合併可見範圍 + is_active + 未刪除 + type(細部 2.5.4 步驟 1)。
func (u *Usecase) ListOptions(ctx context.Context, actor rls.Identity, typ, keyword string) ([]Option, error) {
	if !validTypes[typ] {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("unknown type"))
	}
	if v, ok := optionsCache.get(typ, actor.DepartmentID); ok {
		return v, nil
	}
	q := u.db.Metadict.Query().Where(visibleScope(actor)...).
		Where(entmd.TypeEQ(typ), entmd.IsActiveEQ(true), entmd.DeletedAtIsNil())
	if keyword != "" {
		q = q.Where(entmd.Or(
			entmd.DisplayNameContains(keyword),
			entmd.CodeContainsFold(keyword),
		))
	}
	rows, err := q.Order(entmd.BySortOrder(), entmd.ByCode()).All(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	opts := make([]Option, 0, len(rows))
	for _, r := range rows {
		opts = append(opts, Option{Code: r.Code, DisplayName: r.DisplayName})
	}
	optionsCache.set(typ, actor.DepartmentID, opts)
	return opts, nil
}

// Option 為下拉選項(精簡,細部 2.5.4 介面)。
type Option struct {
	Code        string
	DisplayName string
}
```

記憶體快取(`repository.go` 同檔):

```go
// optionsCache:type+department → 選項,TTL 60 秒;寫入(Create/Update/Delete)主動失效(細部 2.5.4 步驟 4)。
var optionsCache = newTTLCache(60 * time.Second)

type ttlCache struct {
	mu sync.Mutex
	m  map[string]cacheEntry
	ttl time.Duration
}
type cacheEntry struct {
	opts      []Option
	expiresAt time.Time
}

func newTTLCache(ttl time.Duration) *ttlCache { return &ttlCache{m: map[string]cacheEntry{}, ttl: ttl} }

func cacheKey(typ string, dep *uuid.UUID) string {
	if dep == nil {
		return typ + "|sys"
	}
	return typ + "|" + dep.String()
}

func (c *ttlCache) get(typ string, dep *uuid.UUID) ([]Option, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[cacheKey(typ, dep)]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.opts, true
}

func (c *ttlCache) set(typ string, dep *uuid.UUID, opts []Option) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[cacheKey(typ, dep)] = cacheEntry{opts: opts, expiresAt: time.Now().Add(c.ttl)}
}

// invalidateOptionsCache:該 type 所有範圍的快取全失效(簡單正確;量級小不需精細)。
func invalidateOptionsCache(typ string) {
	optionsCache.mu.Lock()
	defer optionsCache.mu.Unlock()
	for k := range optionsCache.m {
		if strings.HasPrefix(k, typ+"|") {
			delete(optionsCache.m, k)
		}
	}
}
```

RLS 驗證測試(整合測試重點 1)補一條:以 `testutil.NewEntClientWithRLS` + `rls.NewContext` 切換 A/B 部門 session,DB 層互不可見 — 測試骨架同 01-auth Task 3 的 `TestRLSTenantIsolation`,查 `Metadict` 即可,程式碼照該模式改寫。

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && go test ./internal/domain/metadicts/ -v`
Expected: PASS — 合併可見性、company 僅系統預設、停用即時消失、非法 type。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/metadicts
git commit -m "feat(backend): metadicts 合併查詢與 ListOptions 選項服務含快取失效(2.5.3-2.5.4)"
```

---

### Task 4: audit_logs schema 與 migration(細部 2.6.1)

**Files:**
- Create: `backend/ent/schema/auditlog.go`
- Create: `backend/database/migrations/00009_audit_logs.sql`
- Test: `backend/internal/domain/auditlogs/schema_test.go`

**Interfaces:**
- Produces: `ent.AuditLog`(欄位 `id`、`company_id`、`department_id *uuid.UUID`、`user_id *uuid.UUID`、`action`、`resource_type`、`resource_id string`、`before_snapshot` / `after_snapshot` JSONB 可空、`ip_address`、`user_agent`、`created_at`;**無** `updated_at` / `deleted_at`);action 常數集合(見下);RLS:讀限 `all` 或同公司;寫僅應用連線。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/domain/auditlogs/schema_test.go`:

```go
package auditlogs_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestAuditLogShape(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)

	l, err := c.AuditLog.Create().
		SetCompanyID(co.ID).
		SetAction("login").
		SetResourceType("user").
		SetResourceID(uuid.New().String()).
		SetIPAddress("203.0.113.1").
		SetUserAgent("test-agent").
		Save(ctx)
	if err != nil {
		t.Fatalf("create audit log: %v", err)
	}
	// JSONB 原樣讀回
	snap := map[string]any{"name": "改前", "n": float64(1)}
	l2, err := c.AuditLog.UpdateOneID(l.ID).SetAfterSnapshot(snap).Save(ctx)
	if err != nil {
		t.Fatalf("snapshot write: %v", err)
	}
	if l2.AfterSnapshot["name"] != "改前" {
		t.Fatalf("snapshot round-trip: %v", l2.AfterSnapshot)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/auditlogs/ -v`
Expected: FAIL — `ent.AuditLog` 未定義。

- [ ] **Step 3: 實作**

`backend/ent/schema/auditlog.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AuditLog 稽核日誌:不可變 — 無 updated_at / deleted_at,不提供 Update/Delete 路徑(細部 2.6.1)。
type AuditLog struct{ ent.Schema }

func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("company_id", uuid.UUID{}),
		field.UUID("department_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("user_id", uuid.UUID{}).Optional().Nillable(), // 系統/排程操作可空
		field.String("action").NotEmpty(),
		field.String("resource_type").NotEmpty(),
		field.String("resource_id").NotEmpty(), // 容納 UUID 與單號
		field.JSON("before_snapshot", map[string]any{}).Optional(),
		field.JSON("after_snapshot", map[string]any{}).Optional(),
		field.String("ip_address").Optional(),
		field.String("user_agent").Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		// 注意:無 updated_at / deleted_at
	}
}

func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("company_id", "created_at"),
		index.Fields("resource_type", "resource_id"),
		index.Fields("user_id", "created_at"),
	}
}
```

`backend/database/migrations/00009_audit_logs.sql`:

```sql
-- +goose Up
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

-- 讀:super(all)全見;company_admin 限自己公司(細部 2.6.3)
CREATE POLICY audit_logs_read ON audit_logs FOR SELECT
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true));

-- 寫:僅後端應用連線(以 current_data_scope 存在與否區隔;RLS 為最後防線,
-- 真正的寫入收斂在 audit.Recorder 的 tx-only 簽名,細部 2.6.2)
CREATE POLICY audit_logs_write ON audit_logs FOR INSERT
  WITH CHECK (current_setting('app.current_company_id', true) IS NOT NULL);

-- 不提供 UPDATE/DELETE policy → 預設拒絕(不可變)
-- +goose Down
DROP POLICY IF EXISTS audit_logs_write ON audit_logs;
DROP POLICY IF EXISTS audit_logs_read ON audit_logs;
ALTER TABLE audit_logs DISABLE ROW LEVEL SECURITY;
```

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go generate ./ent/ && go test ./internal/domain/auditlogs/ -v`
Expected: PASS。

```bash
git add backend/ent/schema/auditlog.go backend/database/migrations/00009_audit_logs.sql backend/internal/domain/auditlogs
git commit -m "feat(backend): audit_logs schema 與 RLS(2.6.1)"
```

---

### Task 5: audit.Recorder 的 DB 實作(細部 2.6.2,接管 01-auth NoopRecorder)

**Files:**
- Create: `backend/internal/audit/pg_recorder.go`
- Update: `backend/internal/server/domains.go`(組裝處以 DB recorder 取代 NoopRecorder)
- Test: `backend/internal/audit/pg_recorder_test.go`

**Interfaces:**
- Consumes: Task 4 `ent.AuditLog`;01-auth Task 14 `audit.Recorder` / `audit.Entry`;Task 2 的 `withTx(ctx, tx)` 慣例。
- Produces: `audit.NewPGRecorder() audit.Recorder`;`audit.TxFromContext(ctx) (*ent.Tx, bool)`;`audit.ContextWithTx(ctx, tx) context.Context`;脈絡鍵 `audit.ContextWithRequestMeta(ctx, ip, userAgent)`。**契約:`Record` 在非交易 ctx 呼叫 → 回錯誤(程式錯誤,細部 2.6.2 錯誤處理)。**

- [ ] **Step 1: 寫失敗測試**

`backend/internal/audit/pg_recorder_test.go`:

```go
package audit_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entmd "github.com/salesorder/sales-order-1.0/backend/ent/metadict"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestRecordRequiresTx(t *testing.T) {
	rec := audit.NewPGRecorder()
	err := rec.Record(context.Background(), audit.Entry{Action: "create"})
	if err == nil {
		t.Fatal("Record without tx must fail (程式錯誤防呆)")
	}
}

func TestRecordWritesInsideTx(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	rec := audit.NewPGRecorder()

	tx, _ := c.Tx(ctx)
	actor := rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DataScope: "all"}
	txCtx := audit.ContextWithRequestMeta(audit.ContextWithTx(ctx, tx), "198.51.100.7", "go-test")
	if err := rec.Record(withIdentity(txCtx, actor), audit.Entry{
		ActorID: actor.UserID.String(), Action: "update",
		ResourceType: "metadict", ResourceID: uuid.New().String(),
		Before: `{"display_name":"舊"}`, After: `{"display_name":"新"}`,
	}); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	l, err := c.AuditLog.Query().Where().Only(ctx)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if l.Action != "update" || l.IPAddress != "198.51.100.7" || *l.UserID != actor.UserID {
		t.Fatalf("audit row wrong: %+v", l)
	}
	if l.AfterSnapshot["display_name"] != "新" {
		t.Fatalf("snapshot: %v", l.AfterSnapshot)
	}
}

func TestRollbackLeavesNoAudit(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	rec := audit.NewPGRecorder()

	tx, _ := c.Tx(ctx)
	txCtx := audit.ContextWithTx(ctx, tx)
	// 業務異動 + 稽核後回滾
	if _, err := tx.Metadict.Create().SetType("unit").SetCode("X").SetDisplayName("X").Save(ctx); err != nil {
		t.Fatal(err)
	}
	_ = rec.Record(withIdentity(txCtx, rls.Identity{UserID: uuid.New(), CompanyID: co.ID}), audit.Entry{
		Action: "create", ResourceType: "metadict", ResourceID: "x",
	})
	_ = tx.Rollback()
	n, _ := c.AuditLog.Query().Count(ctx)
	if n != 0 {
		t.Fatalf("rollback must leave no audit row, got %d", n)
	}
	if n, _ := c.Metadict.Query().Where(entmd.CodeEQ("X")).Count(ctx); n != 0 {
		t.Fatal("rollback must leave no business row")
	}
}

// withIdentity 測試輔助:把 rls.Identity 放入 ctx(生產由 middleware 注入)。
func withIdentity(ctx context.Context, id rls.Identity) context.Context {
	return rls.NewContext(ctx, id)
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/audit/ -v`
Expected: FAIL — `audit.NewPGRecorder` / `ContextWithTx` 未定義。

- [ ] **Step 3: 實作**

`backend/internal/audit/pg_recorder.go`:

```go
package audit

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

type txKey struct{}
type metaKey struct{}

// ContextWithTx 把業務交易放入 ctx;recorder 據此保證同事務寫入(D18)。
func ContextWithTx(ctx context.Context, tx *ent.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFromContext 取交易;未在交易內回 ok=false。
func TxFromContext(ctx context.Context) (*ent.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*ent.Tx)
	return tx, ok
}

// ContextWithRequestMeta 由 middleware 放入 IP / User-Agent(細部 2.6.2 步驟 2)。
func ContextWithRequestMeta(ctx context.Context, ip, userAgent string) context.Context {
	return context.WithValue(ctx, metaKey{}, [2]string{ip, userAgent})
}

// PGRecorder 為 Recorder 的 PostgreSQL 實作;僅接受交易 ctx,否則視為程式錯誤。
type PGRecorder struct{}

func NewPGRecorder() *PGRecorder { return &PGRecorder{} }

func (PGRecorder) Record(ctx context.Context, e Entry) error {
	tx, ok := TxFromContext(ctx)
	if !ok {
		return errors.New("audit: Record called outside transaction (programming error)")
	}
	b := tx.AuditLog.Create().
		SetAction(e.Action).
		SetResourceType(e.ResourceType).
		SetResourceID(e.ResourceID)
	// 操作者與租戶自 rls.Identity;Entry.ActorID 可覆寫(api-token:<名稱> / developer)
	if id, ok := rls.FromContext(ctx); ok {
		b.SetCompanyID(id.CompanyID)
		if id.DepartmentID != nil {
			b.SetDepartmentID(*id.DepartmentID)
		}
		if uid, err := uuid.Parse(e.ActorID); err == nil {
			b.SetUserID(uid)
		}
	}
	if ipUA, ok := ctx.Value(metaKey{}).([2]string); ok {
		b.SetIPAddress(ipUA[0]).SetUserAgent(ipUA[1])
	}
	if e.Before != "" {
		var m map[string]any
		if err := json.Unmarshal([]byte(e.Before), &m); err == nil {
			b.SetBeforeSnapshot(m)
		}
	}
	if e.After != "" {
		var m map[string]any
		if err := json.Unmarshal([]byte(e.After), &m); err == nil {
			b.SetAfterSnapshot(m)
		}
	}
	_, err := b.Save(ctx)
	return err // 失敗由呼叫方回滾整體交易(細部 2.6.2 步驟 4)
}
```

`InitDomains()` 組裝處:`auditRecorder := audit.NewPGRecorder()` 取代 `audit.NoopRecorder{}`,注入各 domain usecase;並加 request-meta middleware:`r.Use(func(next http.Handler) ... ContextWithRequestMeta(r.Context(), r.RemoteAddr, r.UserAgent()))`。

注意:Task 2 測試的 `fakeRecorder` 走 `withTx(ctx, tx)`;統一更名為 `audit.ContextWithTx` — 實作時把 metadicts usecase 的 `withTx` 呼叫改為 `audit.ContextWithTx`(`metadicts/usecase.go` 三處),`snapshotOf` 回傳的 JSON 字串與 `Entry.After` 型別已相容。

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go test ./internal/audit/ ./internal/domain/metadicts/ -v`
Expected: PASS — 無 tx 拒絕、tx 內寫入成功、回滾無孤兒(雙向)。

```bash
git add backend/internal/audit backend/internal/server backend/internal/domain/metadicts/usecase.go
git commit -m "feat(backend): audit.Recorder DB 實作,tx-only 簽名落地 D18(2.6.2)"
```

---

### Task 6: 稽核查詢 API(細部 2.6.3)

**Files:**
- Create: `backend/proto/v1/audit.proto`
- Create: `backend/internal/domain/auditlogs/query.go`
- Test: `backend/internal/domain/auditlogs/query_test.go`

**Interfaces:**
- Consumes: Task 4/5。
- Produces: Connect-RPC `AuditService.List(ListAuditLogsRequest{time_from, time_to, action, resource_type, resource_id, user_id, company_id(super 篩選用), page, per_page}) returns (ListAuditLogsResponse{items, meta})`;`items` 元素含操作者顯示名稱(join users);`Record` RPC 不在 proto(內部以 Go 介面呼叫,細部 2.6.2 介面註:gateway 不暴露)。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/domain/auditlogs/query_test.go`:

```go
package auditlogs_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/auditlogs"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func seedLogs(t *testing.T, coA, coB uuid.UUID) {
	t.Helper()
	c := testutil.NewEntClient(t) // 與測試主體同 container(testutil 單例);實作改為傳入 client
	_ = c
	// 實作:直接用傳入的 client 建兩公司各一筆
}

func TestListScopeAndFilters(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coA, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	coB, _ := c.Company.Create().SetName("乙").SetIdentifier("co-b").SetCustomerCodePrefix("BB").Save(ctx)
	now := time.Now()
	for _, tc := range []struct {
		co   uuid.UUID
		act  string
		ago  time.Duration
	}{{coA.ID, "login", time.Hour}, {coA.ID, "update", 2 * time.Hour}, {coB.ID, "login", time.Hour}} {
		if _, err := c.AuditLog.Create().SetCompanyID(tc.co).SetAction(tc.act).
			SetResourceType("user").SetResourceID("r1").SetCreatedAt(now.Add(-tc.ago)).Save(ctx); err != nil {
			t.Fatal(err)
		}
	}
	q := auditlogs.NewQuery(c)

	// company_admin:僅自己公司
	ca := rls.Identity{UserID: uuid.New(), CompanyID: coA.ID, DataScope: "company", Role: "company_admin"}
	res, err := q.List(ctx, ca, auditlogs.ListInput{})
	if err != nil { t.Fatal(err) }
	if len(res.Items) != 2 {
		t.Fatalf("company scope: want 2, got %d", len(res.Items))
	}
	// action 篩選
	res, err = q.List(ctx, ca, auditlogs.ListInput{Action: ptrStr("login")})
	if err != nil { t.Fatal(err) }
	if len(res.Items) != 1 || res.Items[0].Action != "login" {
		t.Fatalf("action filter: %+v", res.Items)
	}
	// super 跨公司 + company_id 篩選
	super := rls.Identity{UserID: uuid.New(), CompanyID: coA.ID, DataScope: "all", Role: "super"}
	res, err = q.List(ctx, super, auditlogs.ListInput{CompanyID: &coB.ID})
	if err != nil { t.Fatal(err) }
	if len(res.Items) != 1 {
		t.Fatalf("super company filter: want 1, got %d", len(res.Items))
	}
	// staff 被拒
	staff := rls.Identity{UserID: uuid.New(), CompanyID: coA.ID, DataScope: "department", Role: "staff"}
	if _, err := q.List(ctx, staff, auditlogs.ListInput{}); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("staff: want permission_denied, got %v", err)
	}
	// per_page 超上限
	if _, err := q.List(ctx, super, auditlogs.ListInput{PerPage: 101}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("per_page: want invalid_argument, got %v", err)
	}
	// 預設時間窗:4 個月前的紀錄不在無篩選結果(D27 預設近 3 個月)
	if _, err := c.AuditLog.Create().SetCompanyID(coA.ID).SetAction("login").
		SetResourceType("user").SetResourceID("old").
		SetCreatedAt(now.Add(-4 * 30 * 24 * time.Hour)).Save(ctx); err != nil {
		t.Fatal(err)
	}
	res, _ = q.List(ctx, ca, auditlogs.ListInput{})
	for _, it := range res.Items {
		if it.ResourceID == "old" {
			t.Fatal("default window should exclude records older than 3 months")
		}
	}
}

func ptrStr(s string) *string { return &s }
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/auditlogs/ -run TestListScope -v`
Expected: FAIL — `auditlogs.NewQuery` / `ListInput` 未定義。(`seedLogs` 佔位函式刪除,測試已內聯 seed。)

- [ ] **Step 3: 實作 proto 與查詢**

`backend/proto/v1/audit.proto`:

```proto
syntax = "proto3";
package salesorder.v1;

service AuditService {
  rpc List(ListAuditLogsRequest) returns (ListAuditLogsResponse);
}

message AuditLogItem {
  string id = 1; string action = 2; string resource_type = 3; string resource_id = 4;
  optional string user_id = 5; string user_display_name = 6;
  string company_id = 7; optional string department_id = 8;
  string before_snapshot = 9; string after_snapshot = 10; // JSON 字串,可空字串表無
  string ip_address = 11; string user_agent = 12; string created_at = 13;
}
message ListAuditLogsRequest {
  optional string time_from = 1; // RFC3339
  optional string time_to = 2;
  optional string action = 3;
  optional string resource_type = 4;
  optional string resource_id = 5;
  optional string user_id = 6;
  optional string company_id = 7; // 僅 super 有效
  int32 page = 8; int32 per_page = 9;
}
message ListAuditLogsResponse {
  repeated AuditLogItem items = 1;
  salesorder.v1.PageMeta meta = 2; // 若 PageMeta 已定義於 metadict.proto,抽到 common.proto 共用
}
```

`backend/internal/domain/auditlogs/query.go`:

```go
// Package auditlogs 稽核查詢(細部 2.6.3)。
package auditlogs

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	ental "github.com/salesorder/sales-order-1.0/backend/ent/auditlog"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// defaultWindow 對齊 D27:無篩選時預設近 3 個月。
const defaultWindow = 90 * 24 * time.Hour

type ListInput struct {
	TimeFrom, TimeTo               *time.Time
	Action, ResourceType, ResourceID *string
	UserID, CompanyID              *uuid.UUID
	Page, PerPage                  int
}

type Item struct {
	ID               uuid.UUID
	Action           string
	ResourceType     string
	ResourceID       string
	UserID           *uuid.UUID
	UserDisplayName  string
	CompanyID        uuid.UUID
	DepartmentID     *uuid.UUID
	Before, After    string
	IP, UserAgent    string
	CreatedAt        time.Time
}

type ListResult struct {
	Items []Item
	Total int
}

type Query struct{ db *ent.Client }

func NewQuery(db *ent.Client) *Query { return &Query{db: db} }

// List:super 全域可篩公司;company_admin 強制自己公司;其餘角色拒絕(細部 2.6.3 步驟 1)。
func (q *Query) List(ctx context.Context, actor rls.Identity, in ListInput) (*ListResult, error) {
	switch actor.Role {
	case "super":
	case "company_admin":
		in.CompanyID = &actor.CompanyID // 強制限定,忽略請求自帶他公司
	default:
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("audit query not allowed"))
	}
	if in.PerPage > 100 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("per_page exceeds 100"))
	}
	if in.PerPage <= 0 {
		in.PerPage = 20
	}
	if in.TimeFrom != nil && in.TimeTo != nil && in.TimeFrom.After(*in.TimeTo) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("time range inverted"))
	}

	qy := q.db.AuditLog.Query()
	// 預設時間窗防呆(細部 2.6.3 步驟 4)
	from := time.Now().Add(-defaultWindow)
	if in.TimeFrom != nil {
		from = *in.TimeFrom
	}
	qy = qy.Where(ental.CreatedAtGTE(from))
	if in.TimeTo != nil {
		qy = qy.Where(ental.CreatedAtLTE(*in.TimeTo))
	}
	if in.CompanyID != nil {
		qy = qy.Where(ental.CompanyIDEQ(*in.CompanyID))
	}
	if in.Action != nil {
		qy = qy.Where(ental.ActionEQ(*in.Action))
	}
	if in.ResourceType != nil {
		qy = qy.Where(ental.ResourceTypeEQ(*in.ResourceType))
	}
	if in.ResourceID != nil {
		qy = qy.Where(ental.ResourceIDEQ(*in.ResourceID))
	}
	if in.UserID != nil {
		qy = qy.Where(ental.UserIDEQ(*in.UserID))
	}

	total, err := qy.Clone().Count(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	rows, err := qy.Order(ental.ByCreatedAt(entsql.OrderDesc())).
		Offset(max(in.Page-1, 0) * in.PerPage).Limit(in.PerPage).All(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res := &ListResult{Total: total}
	for _, r := range rows {
		res.Items = append(res.Items, toItem(r))
	}
	return res, nil
}
```

(`toItem` 序列化 snapshot map → JSON 字串;操作者顯示名稱以 `user_id` join `users.name`,批次查一次避免 N+1;import 補 `entsql "entgo.io/ent/dialect/sql"`;`max` 用內建(Go 1.21+)。查詢行為本身不寫稽核 — 細部 2.6.3 步驟 6。)

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go test ./internal/domain/auditlogs/ -v`
Expected: PASS — 範圍、篩選、權限、上限、預設時間窗。

```bash
git add backend/proto/v1/audit.proto backend/internal/domain/auditlogs
git commit -m "feat(backend): 稽核查詢 API(範圍推導、預設 3 個月窗)(2.6.3)"
```

---

## Self-Review 記錄

- **Spec 覆蓋**:2.5.1→T1;2.5.2→T2;2.5.3+2.5.4→T3;2.6.1→T4;2.6.2→T5;2.6.3→T6。無缺漏。
- **整合測試重點對應**:RLS 隔離(T3 尾段測試)、同事務一致性(T2 `TestAuditFailureRollsBack` + T5 `TestRollbackLeavesNoAudit`)、軟刪重建(T1)、合併查詢邊界(T3)、選項即時性(T3 快取失效)、稽核完整性(T5;敏感欄位排除由 `snapshotOf` 白名單欄位保證)、查詢防呆(T6)。
- **已知佔位**:01-auth 各 `TODO(Phase 2 Task 2.6)` 稽核寫入點在本計畫 T5 後可逐一接上(以 `audit.ContextWithTx` + `PGRecorder`);`common.proto` 的 `PageMeta` 抽取於首個需要共用它的 Task 順手完成。
- **類型一致**:`audit.Entry{ActorID, Action, ResourceType, ResourceID, Before, After string}`(01-auth T14)= 本計畫 T2/T5 使用一致;`rls.Identity` 含 `Role` / `IsPrimary`(01-auth T13 擴充)。

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/backend/2026-08-17-backend-03-metadicts-audit-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** — 每個 Task 派新 subagent 執行,Task 間 review,迭代快。

**2. Inline Execution** — 用 executing-plans 在本 session 逐批執行,設 checkpoint review。

Which approach?

---

*計畫版本:v1.0.0(2026-08-17);對應細部文件 `detail/03-metadicts-audit.md`、原計畫 v2.9.0、規格書 v1.0.34。*
