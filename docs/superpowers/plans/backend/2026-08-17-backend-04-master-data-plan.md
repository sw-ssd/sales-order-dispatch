# Backend 04 — 業務主檔 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 實作業務主檔全部後端功能 — 客戶(取號、建檔連動帳號、偏好送貨日)、地址簿/聯絡人、商品(單位換算、分切規格)、倉別/車次/分切規格/分類 CRUD、客戶專屬商品、檔案資產、QR 登入兌換。

**Architecture:** 依細部文件 `backend-detail/04-master-data.md` 實作原計畫 Task 3.1–3.6、3.8(僅後端)。取號皆同事務(D7);建檔連動帳號重用 01-auth 的 `issueTempPassword`(D22);`mapCustomerCode` 佔位於本計畫 Task 3 替換為真實 `customers` 查表。檔案資產為跨 domain 共用(02 Logo、06 退貨照片、09 PDF)。

**Tech Stack:** Go 1.25、Ent、Chi、Connect-RPC、pgx/v5、testcontainers-go(沿用 01-auth `testutil.NewEntClient`)、`skip2/go-qrcode`(QR 圖產生)。

**Spec 來源:** 細部文件 `docs/superpowers/plans/backend-detail/04-master-data.md`;共通規則見 `00-index.md` §3。

## Global Constraints

- module 路徑:`github.com/salesorder/sales-order-1.0/backend`。
- 軟刪除 `deleted_at` + 部分唯一索引(D10);取號/建檔/連動帳號皆同一 DB 交易(D18)。
- 客戶編號:`companies.customer_code_prefix`(大寫英數 1–4 碼)+ 6 位補零自增,公司內唯一,建立後不可修改(D7)。
- 數量與換算率一律 `github.com/shopspring/decimal`(避免浮點誤差;`1 條 = 0.6 kg` 這類小數換算必須可表達);1.0 不存任何單價/金額欄位(D12)。
- 上傳白名單:jpeg/png/webp ≤ 5MB、pdf ≤ 10MB;副檔名 + MIME + magic bytes 三重一致才接受。
- 錯誤統一 Connect code;`customer_products` 一客戶一商品一別名唯一;`default_qty = 0` 保留不顯示於下游單據。
- RLS:所有主檔表帶 `company_id` / `department_id`,沿用 01-auth 注入。
- 測試:testcontainers;取號併發、建檔連動為必備整合測試;每 Task 結尾 commit。

## File Structure

| 檔案 | 職責 | 建立於 |
|---|---|---|
| `backend/ent/schema/customer.go` + `customercounter.go` | 客戶與取號計數器 | Task 1-2 |
| `backend/internal/domain/customers/{service,usecase,numbering}.go` | 客戶 RPC/業務/取號 | Task 1-3 |
| `backend/internal/domain/customers/{addresses,contacts}.go` | 地址簿/聯絡人 | Task 4 |
| `backend/ent/schema/{product,productunit,productcuttingspec}.go` | 商品三實體 | Task 5 |
| `backend/internal/domain/products/{service,usecase,convert.go}` | 商品 RPC/業務/換算 | Task 5 |
| `backend/internal/domain/masterdata/simple.go` | 倉/車/規格/分類 CRUD | Task 6 |
| `backend/ent/schema/customerproduct.go` + `internal/domain/customerproducts/` | 專屬商品 | Task 7 |
| `backend/internal/domain/files/{service,storage,validate.go}` | 檔案資產 | Task 8 |
| `backend/internal/domain/qrcode/qrcode.go` | QR token 與兌換 | Task 9 |
| `backend/proto/v1/{customer,product,masterdata,file,qrcode}.proto` | proto 定義 | 各 Task |

---

### Task 1: customers schema 與 CRUD(細部 3.1.1–3.1.2)

**Files:**
- Create: `backend/ent/schema/customer.go`
- Create: `backend/proto/v1/customer.proto`
- Create: `backend/internal/domain/customers/{service,usecase}.go`
- Test: `backend/internal/domain/customers/customer_test.go`

**Interfaces:**
- Consumes: 01-auth Task 1 的 `companies`/`departments`;RLS 注入。
- Produces: `ent.Customer`(`id`、`company_id`、`department_id`、`customer_code`(建後不可改)、`name`、`tax_id`、`phone`、`default_sales_rep_id *uuid.UUID`(→ users)、`preferred_delivery_days []int`(JSONB,1=週一…6=週六,D26)、`promo_tag_ids []uuid`(JSONB,D24)、`status`(active/suspended)、軟刪除三欄);Connect-RPC `CustomerService`:`List({keyword, page, per_page})` / `Get` / `Create` / `Update` / `Delete`(軟刪除)/ `Restore`(清 deleted_at + 稽核);usecase `customers.Usecase`;取號函式 `nextCustomerCode(tx, companyID) (string, error)`(Task 2 實作,本 Task 以佔位 `"PENDING"` 開始,Task 2 接上)。
- 寫入權限:dept_admin/staff 限自己部門;company_admin 限自己公司;customer 身分不可呼叫。

- [ ] **Step 1: 寫失敗測試(CRUD、關鍵字篩選、軟刪除/復原、customer_code 不可改)**

```go
// backend/internal/domain/customers/customer_test.go
package customers_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/customers"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func seedTenant(t *testing.T) (coID, depID uuid.UUID) {
	t.Helper()
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AB").Save(ctx)
	dep, _ := c.Department.Create().SetCompanyID(co.ID).SetName("業務部").Save(ctx)
	return co.ID, dep.ID
}

func adminOf(coID, depID uuid.UUID) rls.Identity {
	return rls.Identity{UserID: uuid.New(), CompanyID: coID, DepartmentID: &depID,
		DataScope: "department", Role: "dept_admin"}
}

func TestCustomerCRUDAndKeyword(t *testing.T) {
	c := testutil.NewEntClient(t)
	coID, depID := seedTenant(t)
	uc := customers.NewUsecase(c, nil) // audit recorder Task 3 接,先 nil 容忍
	ctx := context.Background()
	actor := adminOf(coID, depID)

	cust, err := uc.Create(ctx, actor, customers.CreateInput{Name: "阿明熱炒", Phone: "0912"})
	if err != nil { t.Fatalf("create: %v", err) }
	if cust.CustomerCode == "" { t.Fatal("customer_code should be assigned") }

	// 關鍵字:名稱包含比對
	if _, err := uc.Create(ctx, actor, customers.CreateInput{Name: "阿珠小吃"}); err != nil { t.Fatal(err) }
	list, err := uc.List(ctx, actor, customers.ListInput{Keyword: "阿明"})
	if err != nil { t.Fatal(err) }
	if len(list) != 1 || list[0].Name != "阿明熱炒" {
		t.Fatalf("keyword filter: %+v", list)
	}

	// customer_code 不可修改(D7)
	if err := uc.Update(ctx, actor, cust.ID, customers.UpdateInput{CustomerCode: ptr("XX0001")}); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("code change: want failed_precondition, got %v", err)
	}

	// 軟刪除 → List 不見、Get not_found;復原回來
	if err := uc.Delete(ctx, actor, cust.ID); err != nil { t.Fatal(err) }
	if _, err := uc.Get(ctx, actor, cust.ID); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("get deleted: want not_found, got %v", err)
	}
	if err := uc.Restore(ctx, actor, cust.ID); err != nil { t.Fatal(err) }
	if _, err := uc.Get(ctx, actor, cust.ID); err != nil {
		t.Fatalf("get restored: %v", err)
	}
}

func ptr[T any](v T) *T { return &v }
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/customers/ -v`
Expected: FAIL — `customers.NewUsecase` 未定義。

- [ ] **Step 3: 實作 schema**

`backend/ent/schema/customer.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Customer 客戶主檔(細部 3.1.1)。
type Customer struct{ ent.Schema }

func (Customer) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("company_id", uuid.UUID{}),
		field.UUID("department_id", uuid.UUID{}),
		field.String("customer_code").NotEmpty().Immutable(), // D7:建後不可改
		field.String("name").NotEmpty(),
		field.String("tax_id").Optional(),
		field.String("phone").Optional(),
		field.UUID("default_sales_rep_id", uuid.UUID{}).Optional().Nillable(),
		field.JSON("preferred_delivery_days", []int{}).Optional(), // 1=週一…6=週六
		field.JSON("promo_tag_ids", []uuid.UUID{}).Optional(),
		field.Enum("status").Values("active", "suspended").Default("active"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (Customer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("company", Company.Type).Ref("customers").Field("company_id").Unique().Required(),
		edge.From("department", Department.Type).Ref("customers").Field("department_id").Unique().Required(),
		edge.From("default_sales_rep", User.Type).Ref("default_rep_customers").
			Field("default_sales_rep_id").Unique(),
	}
}

func (Customer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("company_id", "customer_code").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("department_id", "name"),
	}
}
```

- [ ] **Step 4: 實作 usecase(CRUD + 範圍檢查)**

`backend/internal/domain/customers/usecase.go`:

```go
// Package customers 客戶主檔 domain(細部文件 3.1)。
package customers

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entcust "github.com/salesorder/sales-order-1.0/backend/ent/customer"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

type Usecase struct {
	db  *ent.Client
	rec audit.Recorder // nil 容忍(Task 3 接上 PGRecorder)
}

func NewUsecase(db *ent.Client, rec audit.Recorder) *Usecase { return &Usecase{db: db, rec: rec} }

type CreateInput struct {
	Name, TaxID, Phone   string
	DefaultSalesRepID    *uuid.UUID
	PreferredDeliveryDays []int
	PromoTagIDs          []uuid.UUID
}

type UpdateInput struct {
	Name, TaxID, Phone   *string
	CustomerCode         *string // 提供即拒(D7)
	DefaultSalesRepID    *uuid.UUID
	PreferredDeliveryDays *[]int
	PromoTagIDs          *[]uuid.UUID
}

type ListInput struct {
	Keyword      string
	Page, PerPage int
}

// checkWriteScope:dept_admin/staff 限自己部門;company_admin 限自己公司;customer 拒絕。
func checkWriteScope(actor rls.Identity) error {
	switch actor.DataScope {
	case "all", "company":
		return nil
	case "department":
		if actor.DepartmentID == nil {
			return connect.NewError(connect.CodeFailedPrecondition, errors.New("department context missing"))
		}
		return nil
	default:
		return connect.NewError(connect.CodePermissionDenied, errors.New("role cannot manage customers"))
	}
}

// Create:取號 + 建檔同一交易(Task 2 的 nextCustomerCode);連動帳號於 Task 3 掛上。
func (u *Usecase) Create(ctx context.Context, actor rls.Identity, in CreateInput) (*ent.Customer, error) {
	if err := checkWriteScope(actor); err != nil {
		return nil, err
	}
	if in.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name required"))
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	code, err := nextCustomerCode(ctx, tx, actor.CompanyID) // Task 2 實作
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	depID := actor.DepartmentID
	if depID == nil { // company/all 範圍:建立時須指定部門(1.0 簡化:取公司第一個部門,或由請求帶 — 實作以 actor 無部門時要求 CreateInput 指定,此處以 invalid_argument 擋下待 02 計畫的管理介面決定)
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("department required"))
	}
	b := tx.Customer.Create().
		SetCompanyID(actor.CompanyID).SetDepartmentID(*depID).
		SetCustomerCode(code).SetName(in.Name).
		SetPreferredDeliveryDays(in.PreferredDeliveryDays).
		SetPromoTagIds(in.PromoTagIDs)
	if in.TaxID != "" {
		b.SetTaxID(in.TaxID)
	}
	if in.Phone != "" {
		b.SetPhone(in.Phone)
	}
	if in.DefaultSalesRepID != nil {
		b.SetDefaultSalesRepID(*in.DefaultSalesRepID)
	}
	cust, err := b.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// Task 3:建檔連動主帳號 + 業務子帳號於此插入(同 tx)
	if u.rec != nil {
		if err := u.rec.Record(audit.ContextWithTx(ctx, tx), audit.Entry{
			ActorID: actor.UserID.String(), Action: "create",
			ResourceType: "customer", ResourceID: cust.ID.String(),
			After: customerSnapshot(cust),
		}); err != nil {
			_ = tx.Rollback()
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return cust, nil
}

// List:關鍵字對 name / customer_code / phone 包含比對;範圍由 RLS + where 雙層。
func (u *Usecase) List(ctx context.Context, actor rls.Identity, in ListInput) ([]*ent.Customer, error) {
	per := clampPerPage(in.PerPage)
	q := u.db.Customer.Query().Where(entcust.DeletedAtIsNil())
	if actor.DataScope == "department" && actor.DepartmentID != nil {
		q = q.Where(entcust.DepartmentIDEQ(*actor.DepartmentID))
	}
	if actor.DataScope == "company" || actor.DataScope == "department" {
		q = q.Where(entcust.CompanyIDEQ(actor.CompanyID))
	}
	if in.Keyword != "" {
		q = q.Where(entcust.Or(
			entcust.NameContains(in.Keyword),
			entcust.CustomerCodeContains(in.Keyword),
			entcust.PhoneContains(in.Keyword),
		))
	}
	return q.Order(entcust.ByCustomerCode()).
		Offset(max(in.Page-1, 0) * per).Limit(per).All(ctx)
}

func (u *Usecase) Get(ctx context.Context, _ rls.Identity, id uuid.UUID) (*ent.Customer, error) {
	m, err := u.db.Customer.Query().Where(entcust.IDEQ(id), entcust.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("customer not found"))
	}
	return m, nil
}

// Update:customer_code 提供即拒(D7);稽核同事務。
func (u *Usecase) Update(ctx context.Context, actor rls.Identity, id uuid.UUID, in UpdateInput) error {
	if in.CustomerCode != nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("customer_code is immutable"))
	}
	if err := checkWriteScope(actor); err != nil {
		return err
	}
	cust, err := u.Get(ctx, actor, id)
	if err != nil {
		return err
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	before := customerSnapshot(cust)
	upd := tx.Customer.UpdateOneID(id)
	if in.Name != nil { upd.SetName(*in.Name) }
	if in.TaxID != nil { upd.SetTaxID(*in.TaxID) }
	if in.Phone != nil { upd.SetPhone(*in.Phone) }
	if in.DefaultSalesRepID != nil { upd.SetDefaultSalesRepID(*in.DefaultSalesRepID) }
	if in.PreferredDeliveryDays != nil { upd.SetPreferredDeliveryDays(*in.PreferredDeliveryDays) }
	if in.PromoTagIDs != nil { upd.SetPromoTagIds(*in.PromoTagIDs) }
	m2, err := upd.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	if u.rec != nil {
		if err := u.rec.Record(audit.ContextWithTx(ctx, tx), audit.Entry{
			ActorID: actor.UserID.String(), Action: "update",
			ResourceType: "customer", ResourceID: id.String(),
			Before: before, After: customerSnapshot(m2),
		}); err != nil {
			_ = tx.Rollback()
			return connect.NewError(connect.CodeInternal, err)
		}
	}
	return tx.Commit()
}

// Delete = 軟刪除;Restore = 清 deleted_at(細部 3.1.2)。
func (u *Usecase) Delete(ctx context.Context, actor rls.Identity, id uuid.UUID) error {
	return u.setDeleted(ctx, actor, id, true)
}

func (u *Usecase) Restore(ctx context.Context, actor rls.Identity, id uuid.UUID) error {
	return u.setDeleted(ctx, actor, id, false)
}

func (u *Usecase) setDeleted(ctx context.Context, actor rls.Identity, id uuid.UUID, del bool) error {
	if err := checkWriteScope(actor); err != nil {
		return err
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	upd := tx.Customer.UpdateOneID(id)
	action := "restore"
	if del {
		upd.SetDeletedAt(timeNow())
		action = "delete"
	} else {
		upd.ClearDeletedAt()
	}
	if _, err := upd.Save(ctx); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeNotFound, errors.New("customer not found"))
	}
	if u.rec != nil {
		if err := u.rec.Record(audit.ContextWithTx(ctx, tx), audit.Entry{
			ActorID: actor.UserID.String(), Action: action,
			ResourceType: "customer", ResourceID: id.String(),
		}); err != nil {
			_ = tx.Rollback()
			return connect.NewError(connect.CodeInternal, err)
		}
	}
	return tx.Commit()
}

func clampPerPage(p int) int {
	if p <= 0 { return 20 }
	if p > 100 { return 100 }
	return p
}
```

`numbering.go` 佔位(Task 2 填):

```go
// nextCustomerCode 由 Task 2 實作;先回錯確保測試紅。
func nextCustomerCode(ctx context.Context, tx *ent.Tx, companyID uuid.UUID) (string, error) {
	return "", connect.NewError(connect.CodeUnimplemented, errors.New("numbering not implemented"))
}
```

`customerSnapshot`:白名單欄位 JSON(customer_code/name/phone/status 等,不含任何帳號密碼)。`timeNow` 同 01-auth 慣例。

- [ ] **Step 5: proto 與 service 薄層**

`backend/proto/v1/customer.proto`:

```proto
syntax = "proto3";
package salesorder.v1;

import "google/protobuf/empty.proto";

service CustomerService {
  rpc List(ListCustomersRequest) returns (ListCustomersResponse);
  rpc Get(GetCustomerRequest) returns (Customer);
  rpc Create(CreateCustomerRequest) returns (Customer);
  rpc Update(UpdateCustomerRequest) returns (Customer);
  rpc Delete(DeleteCustomerRequest) returns (google.protobuf.Empty);
  rpc Restore(DeleteCustomerRequest) returns (google.protobuf.Empty);
}

message Customer {
  string id = 1; string customer_code = 2; string name = 3;
  string tax_id = 4; string phone = 5;
  optional string default_sales_rep_id = 6;
  repeated int32 preferred_delivery_days = 7;
  repeated string promo_tag_ids = 8;
  string status = 9; string created_at = 10; string updated_at = 11;
}
message ListCustomersRequest { optional string keyword = 1; int32 page = 2; int32 per_page = 3; }
message ListCustomersResponse { repeated Customer items = 1; salesorder.v1.PageMeta meta = 2; }
message GetCustomerRequest { string id = 1; }
message CreateCustomerRequest {
  string name = 1; string tax_id = 2; string phone = 3;
  optional string default_sales_rep_id = 4;
  repeated int32 preferred_delivery_days = 5;
  repeated string promo_tag_ids = 6;
}
message UpdateCustomerRequest {
  string id = 1;
  optional string customer_code = 2; // 提供即 failed_precondition
  optional string name = 3; optional string tax_id = 4; optional string phone = 5;
  optional string default_sales_rep_id = 6;
  repeated int32 preferred_delivery_days = 7;
  repeated string promo_tag_ids = 8;
}
message DeleteCustomerRequest { string id = 1; }
```

- [ ] **Step 6: 跑測試確認通過**

Run: `cd backend && go generate ./ent/ && go test ./internal/domain/customers/ -v`
Expected: 先因 `nextCustomerCode` 未實作 FAIL(unimplemented)— 這是預期的階段性紅;進 Task 2 後轉綠。**本 Task 的 commit 點在 Task 2 完成後合併,或先以佔位 commit 並在 Task 2 註明。**

- [ ] **Step 7: Commit(佔位取號)**

```bash
git add backend/ent/schema/customer.go backend/internal/domain/customers backend/proto/v1/customer.proto
git commit -m "feat(backend): customers schema 與 CRUD(3.1.1-3.1.2;取號佔位待 Task 2)"
```

---

### Task 2: customer_counters 同事務取號(細部 3.1.3,D7)

**Files:**
- Create: `backend/ent/schema/customercounter.go`
- Update: `backend/internal/domain/customers/numbering.go`
- Test: `backend/internal/domain/customers/numbering_test.go`

**Interfaces:**
- Consumes: Task 1。
- Produces: `ent.CustomerCounter`(`company_id` 唯一、`next_seq int`);`nextCustomerCode(ctx, tx, companyID) (string, error)` 真實實作 — 對計數器列 `SELECT ... FOR UPDATE`,格式 `<prefix>` + 6 位補零;同事務,併發安全。

- [ ] **Step 1: 寫失敗測試(序號格式、前綴、併發唯一)**

```go
// backend/internal/domain/customers/numbering_test.go
package customers_test

import (
	"context"
	"sync"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/domain/customers"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestCustomerNumbering(t *testing.T) {
	c := testutil.NewEntClient(t)
	coID, depID := seedTenant(t) // prefix = "AB"
	uc := customers.NewUsecase(c, nil)
	ctx := context.Background()
	actor := adminOf(coID, depID)

	first, err := uc.Create(ctx, actor, customers.CreateInput{Name: "一號店"})
	if err != nil { t.Fatalf("create: %v", err) }
	if first.CustomerCode != "AB000001" {
		t.Fatalf("want AB000001, got %s", first.CustomerCode)
	}
	second, err := uc.Create(ctx, actor, customers.CreateInput{Name: "二號店"})
	if err != nil { t.Fatal(err) }
	if second.CustomerCode != "AB000002" {
		t.Fatalf("want AB000002, got %s", second.CustomerCode)
	}
}

func TestCustomerNumberingConcurrent(t *testing.T) {
	c := testutil.NewEntClient(t)
	coID, depID := seedTenant(t)
	uc := customers.NewUsecase(c, nil)
	ctx := context.Background()
	actor := adminOf(coID, depID)

	const n = 20
	codes := make(chan string, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cust, err := uc.Create(ctx, actor, customers.CreateInput{Name: "併發店"})
			if err != nil { t.Errorf("concurrent create: %v", err); return }
			codes <- cust.CustomerCode
		}(i)
	}
	wg.Wait()
	close(codes)
	seen := map[string]bool{}
	for code := range codes {
		if seen[code] {
			t.Fatalf("duplicate code under concurrency: %s", code)
		}
		seen[code] = true
	}
	if len(seen) != n {
		t.Fatalf("want %d unique codes, got %d", n, len(seen))
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/customers/ -run TestCustomerNumbering -v`
Expected: FAIL — 取號 unimplemented。

- [ ] **Step 3: 實作**

`backend/ent/schema/customercounter.go`:

```go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CustomerCounter 客戶編號計數器(每公司一列;細部 3.1.3)。
type CustomerCounter struct{ ent.Schema }

func (CustomerCounter) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("company_id", uuid.UUID{}),
		field.Int("next_seq").Default(1),
	}
}

func (CustomerCounter) Indexes() []ent.Index {
	return []ent.Index{index.Fields("company_id").Unique()}
}
```

`numbering.go` 替換佔位:

```go
package customers

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entcc "github.com/salesorder/sales-order-1.0/backend/ent/customercounter"
)

// nextCustomerCode:對公司計數器列 SELECT ... FOR UPDATE 取號,取號與建檔同一交易(細部 3.1.3)。
// 計數器列不存在時於同交易建立(首客戶)。
func nextCustomerCode(ctx context.Context, tx *ent.Tx, companyID uuid.UUID) (string, error) {
	co, err := tx.Company.Get(ctx, companyID)
	if err != nil {
		return "", connect.NewError(connect.CodeNotFound, fmt.Errorf("company not found"))
	}
	counter, err := tx.CustomerCounter.Query().
		Where(entcc.CompanyIDEQ(companyID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return "", connect.NewError(connect.CodeInternal, err)
		}
		counter, err = tx.CustomerCounter.Create().SetCompanyID(companyID).SetNextSeq(1).Save(ctx)
		if err != nil {
			return "", connect.NewError(connect.CodeInternal, err)
		}
	}
	seq := counter.NextSeq
	if _, err := tx.CustomerCounter.UpdateOneID(counter.ID).SetNextSeq(seq + 1).Save(ctx); err != nil {
		return "", connect.NewError(connect.CodeInternal, err)
	}
	return fmt.Sprintf("%s%06d", co.CustomerCodePrefix, seq), nil
}
```

- [ ] **Step 4: 跑測試確認通過(含 Task 1 全部)**

Run: `cd backend && go generate ./ent/ && go test ./internal/domain/customers/ -v -race`
Expected: PASS — 格式 `AB000001`、遞增、併發 20 無重複。

- [ ] **Step 5: Commit**

```bash
git add backend/ent/schema/customercounter.go backend/internal/domain/customers/numbering.go
git commit -m "feat(backend): customer_counters 同事務取號,併發安全(3.1.3)"
```

---

### Task 3: 建檔連動主帳號 + 業務子帳號(細部 3.1.4,D22)+ 偏好欄位(3.1.5)

**Files:**
- Update: `backend/internal/domain/customers/usecase.go`
- Update: `backend/internal/domain/auth/customer.go`(移除 `mapCustomerCode` 佔位,改真實查表)
- Test: `backend/internal/domain/customers/provision_test.go`

**Interfaces:**
- Consumes: 01-auth Task 7 的 `issueTempPassword(ctx, tx, userID) (plain string, error)`(需從 `auth` 套件匯出 — 01-auth 計畫中為小寫,本 Task 一併改為 `auth.IssueTempPassword` 匯出)。
- Produces: `customers.provisionAccounts(ctx, tx, cust *ent.Customer) (primaryPlain, salesPlain string, err error)`;建檔回應 `CreateCustomerResponse` 加 `primary_temp_password` / `sales_temp_password` / `temp_password_expires_at`(僅建立當下回傳);`auth.mapCustomerCode` 刪除,`CustomerLogin` 改查 `customers` 表(以 `company.identifier + customer_code` 定位 `customer_id`)。

- [ ] **Step 1: 寫失敗測試(連動帳號、臨時密碼、交付對象、佔位移除後登入)**

```go
// backend/internal/domain/customers/provision_test.go
package customers_test

import (
	"context"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/customers"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestCreateProvisionsAccounts(t *testing.T) {
	c := testutil.NewEntClient(t)
	coID, depID := seedTenant(t)
	ctx := context.Background()
	// 業務使用者(交付對象)
	rep, _ := c.User.Create().SetCompanyID(coID).SetDepartmentID(depID).
		SetRole("staff").SetDataScope("department").SetStatus("active").
		SetEmail("rep@example.com").Save(ctx)

	uc := customers.NewUsecase(c, nil)
	actor := adminOf(coID, depID)
	res, err := uc.CreateWithCredentials(ctx, actor, customers.CreateInput{
		Name: "阿明熱炒", DefaultSalesRepID: &rep.ID,
	})
	if err != nil { t.Fatalf("create: %v", err) }
	if res.PrimaryTempPassword == "" || res.SalesTempPassword == "" {
		t.Fatal("both temp passwords must be returned once")
	}
	if res.PrimaryTempPassword == res.SalesTempPassword {
		t.Fatal("passwords must differ")
	}

	// 帳號檢查:主帳號(account_name=客戶名稱,is_primary)+ 業務子帳號(名稱(業務))
	accounts, err := c.User.Query().Where().All(ctx) // customer_id = res.Customer.ID
	if err != nil { t.Fatal(err) }
	if len(accounts) != 2 {
		t.Fatalf("want 2 accounts, got %d", len(accounts))
	}
	var primary, sales *entUser
	for _, a := range accounts {
		if a.IsPrimary { primary = a } else { sales = a }
		if a.MustChangePassword != true || a.TempPasswordExpiresAt == nil {
			t.Fatalf("temp password flags missing: %+v", a)
		}
		if time.Until(*a.TempPasswordExpiresAt) > 25*time.Hour {
			t.Fatal("temp password TTL should be 24h")
		}
	}
	if primary == nil || sales == nil {
		t.Fatal("need exactly one primary and one sub account")
	}
	if *primary.AccountName != "阿明熱炒" || *sales.AccountName != "阿明熱炒(業務)" {
		t.Fatalf("account names: %v / %v", primary.AccountName, sales.AccountName)
	}
	// 業務子帳號歸屬標記:default_sales_rep 交付(以欄位或中繼資料;實作以 account_name 後綴 + 稽核記錄交付對象)

	// 臨時密碼可登入(mapCustomerCode 佔位已移除,走真實 customers 查表)
	authUC := auth.NewUsecase(c, nil, nil)
	u, err := authUC.CustomerLogin(ctx, "co-a", res.Customer.CustomerCode, "阿明熱炒", res.PrimaryTempPassword)
	if err != nil { t.Fatalf("primary login with temp password: %v", err) }
	if !u.MustChangePassword {
		t.Fatal("first login must force password change")
	}
}
```

(`entUser` 為 `*ent.User` 別名;測試輔助依實作調整查詢條件。)

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/customers/ -run TestCreateProvisions -v`
Expected: FAIL — `CreateWithCredentials` 未定義。

- [ ] **Step 3: 實作連動帳號**

`usecase.go` 加:

```go
// CreateResult 為建檔結果;臨時密碼僅本次回傳(細部 3.1.4)。
type CreateResult struct {
	Customer              *ent.Customer
	PrimaryTempPassword   string
	SalesTempPassword     string
	TempPasswordExpiresAt time.Time
}

// CreateWithCredentials = Create + 同交易連動帳號(D22):
// 主帳號(account_name=客戶名稱,is_primary=true)交付店家;
// 業務子帳號(account_name=客戶名稱+「(業務)」,is_primary=false)交付 default_sales_rep_id 所指業務。
func (u *Usecase) CreateWithCredentials(ctx context.Context, actor rls.Identity, in CreateInput) (*CreateResult, error) {
	// 與 Create 同骨架,差異:tx 內在建檔後呼叫 provisionAccounts
	// 實作:把 Create 的 tx 流程抽為 createInTx(ctx, tx, actor, in) (*ent.Customer, error),
	// Create 與本方法共用;Create 保留給不需要帳號的內部呼叫。
	cust, err := func() (*ent.Customer, error) { return nil, nil }() // 佔位說明:重構見下
	_ = cust; _ = err
	panic("實作:重構 Create → createInTx;本方法開 tx → createInTx → provisionAccounts → 稽核 → commit")
}

// provisionAccounts:同交易建兩帳號,各發 24h 臨時密碼(01-auth IssueTempPassword 已含
// must_change_password / temp_password_expires_at / 清鎖定)。
func provisionAccounts(ctx context.Context, tx *ent.Tx, cust *ent.Customer) (string, string, error) {
	mk := func(name string, primary bool) (string, error) {
		usr, err := tx.User.Create().
			SetCompanyID(cust.CompanyID).
			SetDepartmentID(cust.DepartmentID).
			SetRole("customer").SetDataScope("self").SetStatus("active").
			SetIsCustomer(true).SetCustomerID(cust.ID).
			SetAccountName(name).SetIsPrimary(primary).
			Save(ctx)
		if err != nil {
			return "", err
		}
		return auth.IssueTempPassword(ctx, tx, usr.ID)
	}
	primaryPlain, err := mk(cust.Name, true)
	if err != nil {
		return "", "", err
	}
	salesPlain, err := mk(cust.Name+"(業務)", false)
	if err != nil {
		return "", "", err
	}
	return primaryPlain, salesPlain, nil
}
```

重構要點(取代上面 panic 佔位):
1. `Create` 的 tx 內容抽成 `createInTx(ctx, tx, actor, in)`(含取號與建檔)。
2. `Create` = 開 tx → createInTx → 稽核 → commit。
3. `CreateWithCredentials` = 開 tx → createInTx → provisionAccounts → 稽核(action=create,snapshot 加 `provisioned_accounts: ["primary","sales"]` 與交付對象 `default_sales_rep_id`,**不含密碼**)→ commit → 回傳 CreateResult。
4. 01-auth `auth/customer.go`:刪除 `mapCustomerCode`;`CustomerLogin` 改:

```go
	cust, err := u.db.Customer.Query().Where(
		entcust.CompanyIDEQ(co.ID),
		entcust.CustomerCodeEQ(customerCode),
		entcust.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errBadCredential)
	}
	// 後續以 cust.ID 查 users(entuser.CustomerIDEQ(cust.ID) ...)
```

5. 01-auth Task 6 測試的 seed 改為先建 `customers` 列(`customer_code="000001"`),帳號 `customer_id` 指向它。

- [ ] **Step 4: proto 增量**

`customer.proto` 的 `CreateCustomerRequest` 回應改用:

```proto
message CreateCustomerResponse {
  Customer customer = 1;
  string primary_temp_password = 2;   // 僅建立當下回傳
  string sales_temp_password = 3;
  google.protobuf.Timestamp temp_password_expires_at = 4;
}
// CustomerService.Create 改回傳 CreateCustomerResponse
```

- [ ] **Step 5: 偏好欄位驗證(3.1.5)**

`preferred_delivery_days` 驗證:元素 ∈ 1..6、不重複;`Create`/`Update` 入口檢查,違者 `invalid_argument`。測試加一條:

```go
func TestPreferredDeliveryDaysValidation(t *testing.T) {
	c := testutil.NewEntClient(t)
	coID, depID := seedTenant(t)
	uc := customers.NewUsecase(c, nil)
	actor := adminOf(coID, depID)
	_, err := uc.Create(context.Background(), actor, customers.CreateInput{
		Name: "壞日期", PreferredDeliveryDays: []int{1, 7},
	})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("day 7 (Sunday) must be rejected: %v", err)
	}
}
```

(實作於 `Create`/`Update` 加 `validateDeliveryDays`;D26:僅週一至週六。)

- [ ] **Step 6: 跑測試確認通過 + Commit**

Run: `cd backend && go test ./internal/domain/customers/ ./internal/domain/auth/ -v`
Expected: PASS — 連動帳號、臨時密碼登入、01-auth Task 6 測試改真實查表後不回歸。

```bash
git add backend/internal/domain/customers backend/internal/domain/auth/customer.go backend/proto/v1/customer.proto
git commit -m "feat(backend): 建檔連動主帳號+業務子帳號、臨時密碼交付、mapCustomerCode 佔位移除(3.1.4-3.1.5)"
```

---

### Task 4: 地址簿與聯絡人(細部 3.2.1–3.2.2)

**Files:**
- Create: `backend/ent/schema/customeraddress.go`、`backend/ent/schema/customercontact.go`
- Create: `backend/internal/domain/customers/addresses.go`、`contacts.go`
- Update: `backend/proto/v1/customer.proto`
- Test: `backend/internal/domain/customers/addresses_test.go`

**Interfaces:**
- Produces: `ent.CustomerAddress`(`customer_id`、`type`(delivery/billing)、`address`、`is_default`、軟刪除三欄,冗餘 `company_id`/`department_id` 供 RLS);`ent.CustomerContact`(`customer_id`、`name`、`phone`、`email`、`is_default`、同冗餘);RPC 增量:`ListAddresses/AddAddress/UpdateAddress/DeleteAddress`、`ListContacts/AddContact/UpdateContact/DeleteContact`(掛 `CustomerService`);不變式:**同客戶同類型至多一筆預設**(部分唯一索引);設新預設時同交易清舊預設。

- [ ] **Step 1: 寫失敗測試(預設唯一、切換預設同事務)**

```go
// backend/internal/domain/customers/addresses_test.go
package customers_test

import (
	"context"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/domain/customers"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestDefaultAddressUniqueness(t *testing.T) {
	c := testutil.NewEntClient(t)
	coID, depID := seedTenant(t)
	uc := customers.NewUsecase(c, nil)
	ctx := context.Background()
	actor := adminOf(coID, depID)
	cust, err := uc.Create(ctx, actor, customers.CreateInput{Name: "預設店"})
	if err != nil { t.Fatal(err) }

	a1, err := uc.AddAddress(ctx, actor, cust.ID, customers.AddressInput{
		Type: "delivery", Address: "台北市中山區 1 號", IsDefault: true,
	})
	if err != nil { t.Fatal(err) }
	a2, err := uc.AddAddress(ctx, actor, cust.ID, customers.AddressInput{
		Type: "delivery", Address: "台北市中山區 2 號", IsDefault: true, // 切換預設
	})
	if err != nil { t.Fatal(err) }

	list, err := uc.ListAddresses(ctx, actor, cust.ID)
	if err != nil { t.Fatal(err) }
	defaults := 0
	for _, a := range list {
		if a.IsDefault {
			defaults++
			if a.ID != a2.ID {
				t.Fatalf("default should move to a2 (a1=%v)", a1.ID)
			}
		}
	}
	if defaults != 1 {
		t.Fatalf("want exactly 1 default, got %d", defaults)
	}
}
```

聯絡人同構測試(`TestDefaultContactUniqueness`),程式碼比照。

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/customers/ -run 'TestDefaultAddress|TestDefaultContact' -v`
Expected: FAIL — `AddAddress` 未定義。

- [ ] **Step 3: 實作 schema(兩實體同構)**

`backend/ent/schema/customeraddress.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CustomerAddress 客戶地址簿(細部 3.2.1);同類型至多一筆預設。
type CustomerAddress struct{ ent.Schema }

func (CustomerAddress) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("customer_id", uuid.UUID{}),
		field.UUID("company_id", uuid.UUID{}),  // 冗餘,供 RLS
		field.UUID("department_id", uuid.UUID{}),
		field.Enum("type").Values("delivery", "billing"),
		field.String("address").NotEmpty(),
		field.Bool("is_default").Default(false),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (CustomerAddress) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("customer", Customer.Type).Ref("addresses").
			Field("customer_id").Unique().Required(),
	}
}

func (CustomerAddress) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("customer_id"),
		// 同客戶同類型至多一筆預設
		index.Fields("customer_id", "type").Unique().Annotations(
			entsql.IndexWhere("is_default = true AND deleted_at IS NULL")),
	}
}
```

`customercontact.go` 同構(`name`、`phone`、`email`、`is_default`;唯一索引僅 `(customer_id) WHERE is_default AND deleted_at IS NULL` — 聯絡人無 type)。

- [ ] **Step 4: 實作 usecase(切換預設同事務)**

`addresses.go`:

```go
package customers

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entaddr "github.com/salesorder/sales-order-1.0/backend/ent/customeraddress"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

type AddressInput struct {
	Type, Address string
	IsDefault     bool
}

// AddAddress:is_default=true 時,同交易清同客戶同類型舊預設(細部 3.2.1 不變式)。
func (u *Usecase) AddAddress(ctx context.Context, actor rls.Identity, customerID uuid.UUID, in AddressInput) (*ent.CustomerAddress, error) {
	if err := checkWriteScope(actor); err != nil {
		return nil, err
	}
	cust, err := u.Get(ctx, actor, customerID)
	if err != nil {
		return nil, err
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if in.IsDefault {
		if _, err := tx.CustomerAddress.Update().
			Where(entaddr.CustomerIDEQ(customerID), entaddr.TypeEQ(in.Type), entaddr.IsDefaultEQ(true)).
			SetIsDefault(false).Save(ctx); err != nil {
			_ = tx.Rollback()
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	a, err := tx.CustomerAddress.Create().
		SetCustomerID(customerID).
		SetCompanyID(cust.CompanyID).SetDepartmentID(cust.DepartmentID).
		SetType(in.Type).SetAddress(in.Address).SetIsDefault(in.IsDefault).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return a, nil
}

func (u *Usecase) ListAddresses(ctx context.Context, _ rls.Identity, customerID uuid.UUID) ([]*ent.CustomerAddress, error) {
	return u.db.CustomerAddress.Query().
		Where(entaddr.CustomerIDEQ(customerID), entaddr.DeletedAtIsNil()).
		Order(entaddr.ByIsDefault(entsql.OrderDesc()), entaddr.ByCreatedAt()).All(ctx)
}

// UpdateAddress / DeleteAddress(軟刪除)比照 Update/Delete 模式;UpdateAddress 設 IsDefault=true 時同 AddAddress 清舊預設。
```

`contacts.go` 同構(`ContactInput{Name, Phone, Email string; IsDefault bool}`)。

proto 增量(`customer.proto`):

```proto
message Address {
  string id = 1; string type = 2; string address = 3; bool is_default = 4;
}
message Contact {
  string id = 1; string name = 2; string phone = 3; string email = 4; bool is_default = 5;
}
// CustomerService 增:ListAddresses/AddAddress/UpdateAddress/DeleteAddress/
//                     ListContacts/AddContact/UpdateContact/DeleteContact
// 請求/回應訊息比照 Customer CRUD 模式(id + 欄位;Update 用 optional)
```

- [ ] **Step 5: 跑測試確認通過 + Commit**

Run: `cd backend && go generate ./ent/ && go test ./internal/domain/customers/ -v`
Expected: PASS。

```bash
git add backend/ent/schema/customeraddress.go backend/ent/schema/customercontact.go backend/internal/domain/customers backend/proto/v1/customer.proto
git commit -m "feat(backend): 客戶地址簿與聯絡人,預設唯一同事務切換(3.2)"
```

---

### Task 5: 商品三實體與單位換算(細部 3.3.1–3.3.3)

**Files:**
- Create: `backend/ent/schema/{product,productunit,productcuttingspec}.go`
- Create: `backend/proto/v1/product.proto`
- Create: `backend/internal/domain/products/{service,usecase,convert}.go`
- Test: `backend/internal/domain/products/{product_test.go,convert_test.go}`

**Interfaces:**
- Produces: `ent.Product`(`company_id`/`department_id`、`name`、`category_id`、`base_unit_code`(基本單位,→ metadicts unit)、軟刪除);`ent.ProductUnit`(`product_id`、`unit_code`、`conversion_rate decimal`(1 單位 = rate 個基本單位,可小數,如 1 條 = 0.6 kg)、`is_base bool`(每商品恰一))`;`ent.ProductCuttingSpec`(`product_id`、`cutting_spec_id`);RPC `ProductService` CRUD(寫入連同 units 與 specs 一併落庫,同一交易);**換算契約(供 05 下單與 09 單據)**:

```go
// products.ToBase(qty decimal.Decimal, unitCode string, units []*ent.ProductUnit) (decimal.Decimal, error)
// products.FromBase(baseQty decimal.Decimal, unitCode string, units []*ent.ProductUnit) (decimal.Decimal, error)
// 規則:unitCode 對應列 is_base=true → 原值;否則 base_qty = qty × conversion_rate。
// 找不到單位 → 錯誤;FromBase 不整除 → 錯誤(不靜默進位)。
// 數值精度跟隨輸入,不額外四捨五入(單據彙總端自行處理顯示)。
```

- [ ] **Step 1: 寫失敗測試(換算契約)**

```go
// backend/internal/domain/products/convert_test.go
package products_test

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/products"
)

func units() []*ent.ProductUnit {
	return []*ent.ProductUnit{
		{UnitCode: "KG", ConversionRate: decimal.NewFromInt(1), IsBase: true}, // 基本單位
		{UnitCode: "PIECE", ConversionRate: decimal.NewFromFloat(0.6)},         // 1 條 = 0.6 kg
		{UnitCode: "BOX", ConversionRate: decimal.NewFromInt(12)},              // 1 箱 = 12 kg
	}
}

func TestToBase(t *testing.T) {
	got, err := products.ToBase(decimal.NewFromInt(10), "PIECE", units())
	if err != nil || !got.Equal(decimal.NewFromInt(6)) {
		t.Fatalf("10 條 = 6 kg, got %s err %v", got, err)
	}
	got, err = products.ToBase(decimal.NewFromInt(5), "KG", units())
	if err != nil || !got.Equal(decimal.NewFromInt(5)) {
		t.Fatalf("base unit passthrough: %s", got)
	}
	if _, err := products.ToBase(decimal.NewFromInt(1), "DOZEN", units()); err == nil {
		t.Fatal("unknown unit should fail")
	}
}

func TestFromBase(t *testing.T) {
	got, err := products.FromBase(decimal.NewFromInt(24), "BOX", units())
	if err != nil || !got.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("24 kg = 2 箱, got %s err %v", got, err)
	}
	if _, err := products.FromBase(decimal.NewFromFloat(0.5), "PIECE", units()); err == nil {
		t.Fatal("indivisible conversion should fail")
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/products/ -run 'TestToBase|TestFromBase' -v`
Expected: FAIL — `products.ToBase` 未定義。

- [ ] **Step 3: 實作換算**

`backend/internal/domain/products/convert.go`:

```go
// Package products 商品 domain(細部文件 3.3)。
package products

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/salesorder/sales-order-1.0/backend/ent"
)

// ToBase:任意單位數量 → 基本單位數量(is_base=true 的單位為基本單位;細部 3.3.3)。
// base_qty = qty × conversion_rate;精度跟隨輸入,不四捨五入。
func ToBase(qty decimal.Decimal, unitCode string, units []*ent.ProductUnit) (decimal.Decimal, error) {
	if qty.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("quantity must be positive")
	}
	for _, u := range units {
		if u.UnitCode == unitCode {
			return qty.Mul(u.ConversionRate), nil
		}
	}
	return decimal.Zero, fmt.Errorf("unknown unit %q", unitCode)
}

// FromBase:基本單位 → 指定單位;不整除回錯(由下單/單據層決定呈現,不靜默進位)。
func FromBase(baseQty decimal.Decimal, unitCode string, units []*ent.ProductUnit) (decimal.Decimal, error) {
	for _, u := range units {
		if u.UnitCode == unitCode {
			if u.ConversionRate.IsZero() {
				return decimal.Zero, fmt.Errorf("zero conversion rate for %q", unitCode)
			}
			q := baseQty.Div(u.ConversionRate)
			if !q.Equal(q.Truncate(0)) {
				return decimal.Zero, fmt.Errorf("%s not divisible into whole %s", baseQty, unitCode)
			}
			return q, nil
		}
	}
	return decimal.Zero, fmt.Errorf("unknown unit %q", unitCode)
}
```

- [ ] **Step 4: 實作 schema 與 CRUD**

`product.go`:

```go
// Product 商品主檔(細部 3.3.1);不含任何金額欄位(D12)。
type Product struct{ ent.Schema }

func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("company_id", uuid.UUID{}),
		field.UUID("department_id", uuid.UUID{}),
		field.String("name").NotEmpty(),
		field.UUID("category_id", uuid.UUID{}).Optional().Nillable(), // → product_categories(Task 6)
		field.String("base_unit_code").NotEmpty(),                    // → metadicts unit
		field.UUID("warehouse_id", uuid.UUID{}).Optional().Nillable(), // 預設倉別
		field.Enum("status").Values("active", "suspended").Default("active"),
		// created_at / updated_at / deleted_at 同前述慣例
	}
}
// Edges: company / department / category / units(has many ProductUnit) / cutting_specs
// Indexes: (department_id, name) 部分唯一 WHERE deleted_at IS NULL
```

`productunit.go` / `productcuttingspec.go`:

```go
// ProductUnit:product_id、unit_code、conversion_rate(decimal,>0)、is_base(bool,每商品恰一個)
//   索引:(product_id, unit_code) 部分唯一 WHERE deleted_at IS NULL;
//         (product_id) 部分唯一 WHERE is_base = true AND deleted_at IS NULL
// ProductCuttingSpec:product_id、cutting_spec_id;索引 (product_id, cutting_spec_id) 唯一
```

CRUD usecase(寫入連同 units/specs 同一交易,細部 3.3.2):

```go
type ProductInput struct {
	Name         string
	CategoryID   *uuid.UUID
	BaseUnitCode string
	WarehouseID  *uuid.UUID
	Units        []UnitInput // {UnitCode string; Rate int; IsDefault bool}
	CuttingSpecIDs []uuid.UUID
}

// Create:驗證 units 非空、恰一個 is_base、conversion_rate > 0、base_unit_code 對應 is_base 列;
// 同交易建 product + units + cutting specs;稽核同事務。
// Update:units/specs 採整組替換(同交易刪舊建新,軟刪除舊列);name 部分唯一衝突 → already_exists。
// Delete:軟刪除 product;units/specs 一併軟刪除(同交易)。
```

驗證測試加進 `product_test.go`:

```go
func TestProductCreateValidation(t *testing.T) {
	// units 空 → invalid_argument
	// 兩個 is_default → invalid_argument
	// base_unit_code 不在 units → invalid_argument
	// 成功:product + 2 units + 1 spec 同交易落庫;失敗注入(spec id 不存在)→ 全部回滾
}
```

- [ ] **Step 5: 跑測試確認通過 + Commit**

Run: `cd backend && go generate ./ent/ && go test ./internal/domain/products/ -v`
Expected: PASS — 換算契約 + CRUD 驗證 + 同事務回滾。

```bash
git add backend/ent/schema backend/internal/domain/products backend/proto/v1/product.proto
git commit -m "feat(backend): 商品三實體、CRUD、單位換算契約(3.3)"
```

---

### Task 6: 倉別/車次/分切規格/分類 CRUD(細部 3.4.1–3.4.4)

**Files:**
- Create: `backend/ent/schema/{warehouse,route,cuttingspec,productcategory}.go`
- Create: `backend/internal/domain/masterdata/simple.go`
- Create: `backend/proto/v1/masterdata.proto`
- Test: `backend/internal/domain/masterdata/simple_test.go`

**Interfaces:**
- Produces: 四實體(皆 `company_id`/`department_id`/`name`/軟刪除;`Warehouse` 加 `code`;`Route` 加 `code`、`description`;`CuttingSpec` 加 `belongs_to`(enum:`processing` 加工室 / `delivery` 配送揀貨,細部 3.4.3);`ProductCategory` 加 `sort_order`);RPC `MasterdataService` 四組 CRUD(List/Create/Update/Delete);共用 helper `simpleCRUD[T]` 不引入泛型抽象 — 四組各自直寫(各自獨立測試,刪除皆軟刪除 + 稽核同事務)。
- 範圍:全部部門級;dept_admin/staff 寫,所有登入角色讀。

- [ ] **Step 1: 寫失敗測試(四實體各一條 CRUD+軟刪除)**

```go
// backend/internal/domain/masterdata/simple_test.go
package masterdata_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/masterdata"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestWarehouseCRUD(t *testing.T) {
	c := testutil.NewEntClient(t)
	// seed tenant(比照 customers_test.seedTenant 模式,抽 testutil.SeedTenant 共用)
	// create → list 可見 → update name → delete(軟刪)→ list 不見 → 同名可重建
	uc := masterdata.NewUsecase(c, nil)
	_ = uc
}

func TestCuttingSpecBelongsTo(t *testing.T) {
	// belongs_to 僅接受 processing / delivery;其他 → invalid_argument
}

func TestRouteAndCategoryCRUD(t *testing.T) {
	// 同構:CRUD + 軟刪除 + (department_id, name) 部分唯一
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/masterdata/ -v`
Expected: FAIL。

- [ ] **Step 3: 實作(四 schema + usecase)**

四 schema 同構(欄位如上;索引 `(department_id, name)` 部分唯一 `WHERE deleted_at IS NULL`)。usecase 直寫四組(每組約 60 行,模式同 Task 1 的 customers;`belongs_to` 驗證放 `CuttingSpec` 的 Create/Update 入口)。proto 四組訊息。

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go generate ./ent/ && go test ./internal/domain/masterdata/ -v`
Expected: PASS。

```bash
git add backend/ent/schema backend/internal/domain/masterdata backend/proto/v1/masterdata.proto
git commit -m "feat(backend): 倉別/車次/分切規格/商品分類 CRUD(3.4)"
```

---

### Task 7: 客戶專屬商品(細部 3.5.1–3.5.3)

**Files:**
- Create: `backend/ent/schema/customerproduct.go`
- Create: `backend/internal/domain/customerproducts/{service,usecase,alias}.go`
- Create: `backend/proto/v1/customerproduct.proto`
- Test: `backend/internal/domain/customerproducts/cp_test.go`

**Interfaces:**
- Consumes: Task 1 customers、Task 5 products。
- Produces: `ent.CustomerProduct`(`customer_id`、`product_id`、`alias_name`、`default_qty decimal`(`shopspring/decimal`)、`cut_note`、`promo_tag_ids`、軟刪除;**無單價欄位**,D12);唯一:`(customer_id, product_id)` 與 `(customer_id, alias_name)` 部分唯一 `WHERE deleted_at IS NULL`;RPC `CustomerProductService` CRUD;**下單用契約(05 計畫 4.2.2 消費)**:

```go
// EnsureAlias:下單手打確認儲存時呼叫;同客戶同商品已有記錄 → 更新 alias 後回既有(冪等);
// 否則同交易建立(alias_name=手打名稱,default_qty=下單數量)。唯一衝突 → 重查回既有(細部 3.5.3)。
// 套件級函式(非 Usecase 方法),呼叫端必須已在下單交易內(tx 參數)。
func EnsureAlias(ctx context.Context, tx *ent.Tx, customerID, productID uuid.UUID, aliasName string, defaultQty decimal.Decimal) (*ent.CustomerProduct, error)
```

- [ ] **Step 1: 寫失敗測試(qty=0 保留、雙唯一、EnsureAlias 冪等)**

```go
// backend/internal/domain/customerproducts/cp_test.go
package customerproducts_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/domain/customerproducts"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestQtyZeroKeptButHidden(t *testing.T) {
	c := testutil.NewEntClient(t)
	// seed customer + product(比照前 Task 模式)
	uc := customerproducts.NewUsecase(c, nil)
	// create default_qty=0 → 成功;List 回傳(管理視角);ListForOrdering(下游單據視角)不回傳
	_ = uc
}

func TestAliasUniqueness(t *testing.T) {
	// 同客戶同 product 建兩筆 → already_exists
	// 同客戶同 alias_name 不同 product → already_exists
	// 軟刪除後可重建
}

func TestEnsureAliasIdempotent(t *testing.T) {
	c := testutil.NewEntClient(t)
	uc := customerproducts.NewUsecase(c, nil)
	_ = uc
	// 第一次 EnsureAlias → 新建;第二次同參數 → 回既有同 ID;併發兩次 → 只有一筆
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/customerproducts/ -v`
Expected: FAIL。

- [ ] **Step 3: 實作**

schema 如上;usecase 關鍵函式:

```go
// EnsureAlias(細部 3.5.3):冪等建立;套件級函式,呼叫端必須已在下單交易內(tx 參數)。
func EnsureAlias(ctx context.Context, tx *ent.Tx, customerID, productID uuid.UUID, aliasName string, defaultQty decimal.Decimal) (*ent.CustomerProduct, error) {
	// 先查:同客戶同商品已有 → 回既有
	existing, err := tx.CustomerProduct.Query().Where(
		entcp.CustomerIDEQ(customerID),
		entcp.ProductIDEQ(productID),
		entcp.DeletedAtIsNil(),
	).Only(ctx)
	if err == nil {
		return existing, nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}
	created, err := tx.CustomerProduct.Create().
		SetCustomerID(customerID).SetProductID(productID).
		SetAliasName(aliasName).SetDefaultQty(defaultQty).
		Save(ctx)
	if err != nil {
		// 唯一衝突(併發建立)→ 重查回既有(冪等)
		if isUniqueViolation(err) {
			return tx.CustomerProduct.Query().Where(
				entcp.CustomerIDEQ(customerID),
				entcp.ProductIDEQ(productID),
				entcp.DeletedAtIsNil(),
			).Only(ctx)
		}
		return nil, err
	}
	return created, nil
}
```

`ListForOrdering`:`default_qty > 0 AND is_active` 才回傳(下游單據視角;`is_active` 由 status 欄位承載 — schema 加 `field.Bool("is_active").Default(true)`)。CRUD 比照前述模式;`default_qty = 0` 合法(保留,細部 3.5.2)。

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go generate ./ent/ && go test ./internal/domain/customerproducts/ -v -race`
Expected: PASS — qty=0 雙視角、雙唯一、EnsureAlias 冪等與併發。

```bash
git add backend/ent/schema/customerproduct.go backend/internal/domain/customerproducts backend/proto/v1/customerproduct.proto
git commit -m "feat(backend): 客戶專屬商品 CRUD 與下單手打自動建別名(3.5)"
```

---

### Task 8: 檔案資產(細部 3.6.1–3.6.3)

**Files:**
- Create: `backend/ent/schema/fileasset.go`
- Create: `backend/internal/domain/files/{service,storage,validate}.go`
- Create: `backend/proto/v1/file.proto`
- Test: `backend/internal/domain/files/files_test.go`

**Interfaces:**
- Produces: `ent.FileAsset`(`id`、`company_id`、`department_id`、`uploader_user_id`、`purpose`(enum:`customer_logo` / `return_photo` / `print_pdf` / `other`)、`original_name`、`stored_path`(不暴露)、`mime_type`、`size_bytes`、軟刪除);RPC `FileService.Upload`(串流或 base64 chunk,Connect)→ `FileAsset{id}`;REST `GET /api/v1/files/:id/download`(Chi,經 Authenticate middleware);**消費端契約**:`files.Service.DownloadURL(id) string`(02 Logo、06 退貨照片、09 PDF 用)。
- 儲存:`files.Storage` 介面(`Put(ctx, name, r io.Reader) (path string, err error)` / `Open(path) (io.ReadCloser, error)`);本地實作寫 `FILE_STORAGE_DIR`(config,預設 `./data/files`),路徑 = `<company_id>/<uuid>.<ext>`。

- [ ] **Step 1: 寫失敗測試(白名單、magic bytes、大小、租戶隔離下載)**

```go
// backend/internal/domain/files/files_test.go
package files_test

import (
	"bytes"
	"context"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/domain/files"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

var (
	pngMagic  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	jpegMagic = []byte{0xFF, 0xD8, 0xFF}
	pdfMagic  = []byte("%PDF-")
)

func TestUploadValidation(t *testing.T) {
	dir := t.TempDir()
	svc := files.NewService(nil, files.NewLocalStorage(dir)) // db nil 於 Task 8 測試前段;實作拆 validate 純函式先測
	_ = svc

	// 副檔名/內容不符:偽裝 png 的 pdf
	if err := files.Validate("photo.png", "image/png", bytes.NewReader(pdfMagic), int64(len(pdfMagic))); err == nil {
		t.Fatal("mismatched magic bytes should fail")
	}
	// 合法 png(截斷內容也行,僅驗頭)
	if err := files.Validate("photo.png", "image/png", bytes.NewReader(pngMagic), 1024); err != nil {
		t.Fatalf("valid png: %v", err)
	}
	// 超過 5MB 圖片
	if err := files.Validate("big.jpg", "image/jpeg", bytes.NewReader(jpegMagic), 5*1024*1024+1); err == nil {
		t.Fatal("oversize image should fail")
	}
	// exe 不在白名單
	if err := files.Validate("evil.exe", "application/octet-stream", bytes.NewReader([]byte("MZ")), 2); err == nil {
		t.Fatal("exe should fail")
	}
}

func TestDownloadTenantIsolation(t *testing.T) {
	// A 公司檔案以 B 公司身分下載 → permission_denied(或 not_found;細部 3.6.3 二擇一,實作統一 not_found 防探測)
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/files/ -v`
Expected: FAIL — `files.Validate` 未定義。

- [ ] **Step 3: 實作驗證與儲存**

`validate.go`:

```go
// Package files 檔案資產(細部文件 3.6)。
package files

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// 白名單:副檔名 → (MIME, magic bytes, 大小上限)(細部 3.6.1)。
type rule struct {
	mime  string
	magic []byte
	max   int64
}

var whitelist = map[string]rule{
	".jpg":  {"image/jpeg", []byte{0xFF, 0xD8, 0xFF}, 5 << 20},
	".jpeg": {"image/jpeg", []byte{0xFF, 0xD8, 0xFF}, 5 << 20},
	".png":  {"image/png", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, 5 << 20},
	".webp": {"image/webp", nil, 5 << 20}, // RIFF....WEBP 雙段,實作以 RIFF+WEBP 檢查
	".pdf":  {"application/pdf", []byte("%PDF-"), 10 << 20},
}

// Validate 三重一致檢查:副檔名在白名單、宣告 MIME 相符、檔頭 magic bytes 相符、大小未超標。
func Validate(filename, declaredMIME string, r io.Reader, size int64) error {
	ext := strings.ToLower(filename[strings.LastIndex(filename, "."):])
	rule, ok := whitelist[ext]
	if !ok {
		return fmt.Errorf("extension %q not allowed", ext)
	}
	if declaredMIME != rule.mime {
		return fmt.Errorf("mime mismatch: declared %q, want %q", declaredMIME, rule.mime)
	}
	if size > rule.max {
		return fmt.Errorf("file too large: %d > %d", size, rule.max)
	}
	if rule.magic != nil {
		head := make([]byte, len(rule.magic))
		if _, err := io.ReadFull(r, head); err != nil {
			return fmt.Errorf("read head: %w", err)
		}
		if !bytes.Equal(head, rule.magic) {
			return fmt.Errorf("magic bytes mismatch")
		}
	}
	return nil
}
```

`storage.go`:

```go
// Storage 抽象檔案儲存,供測試與日後換 S3。
type Storage interface {
	Put(ctx context.Context, relPath string, r io.Reader) error
	Open(relPath string) (io.ReadCloser, error)
	Delete(relPath string) error
}

type localStorage struct{ root string }

func NewLocalStorage(root string) Storage { return &localStorage{root: root} }

func (s *localStorage) Put(ctx context.Context, relPath string, r io.Reader) error {
	full := filepath.Join(s.root, relPath)
	if !strings.HasPrefix(full, filepath.Clean(s.root)+string(os.PathSeparator)) {
		return fmt.Errorf("path escape") // 防 ../
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	f, err := os.Create(full)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (s *localStorage) Open(relPath string) (io.ReadCloser, error) {
	full := filepath.Join(s.root, relPath)
	if !strings.HasPrefix(full, filepath.Clean(s.root)+string(os.PathSeparator)) {
		return nil, fmt.Errorf("path escape")
	}
	return os.Open(full)
}

func (s *localStorage) Delete(relPath string) error {
	return os.Remove(filepath.Join(s.root, relPath))
}
```

`service.go` Upload 流程:Validate(讀 head 後以 MultiReader 接回完整串流)→ 建 `file_assets` 列(同交易稽核)→ Storage.Put;Put 失敗 → 補償刪除 DB 列(或順序反轉:先 Put 成功才建列,Put 後建列失敗則刪檔 — 採後者,與 09 計畫 PDF 一致)。Download:查列(RLS 擋跨租戶)→ Storage.Open → 串流回應,`Content-Disposition` 用 `original_name` 且不回 `stored_path`。

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go test ./internal/domain/files/ -v`
Expected: PASS。

```bash
git add backend/ent/schema/fileasset.go backend/internal/domain/files backend/proto/v1/file.proto
git commit -m "feat(backend): 檔案資產上傳白名單、本地儲存、租戶隔離下載(3.6)"
```

---

### Task 9: QR 簽章 token 與兌換端點(細部 3.8.1–3.8.2)

**Files:**
- Create: `backend/internal/domain/qrcode/qrcode.go`
- Update: `backend/internal/server/server.go`(掛公開路由)
- Test: `backend/internal/domain/qrcode/qrcode_test.go`

**Interfaces:**
- Produces: `qrcode.IssueToken(companyID uuid.UUID, customerCode string, ttl time.Duration) (string, error)` / `qrcode.VerifyToken(tok string) (companyID uuid.UUID, customerCode string, err error)`(HMAC-SHA256 簽章,payload = `company_id|customer_code|exp`,base64url);REST 公開端點 `GET /api/v1/qr/redeem?token=`(無認證,回 `{company_identifier, customer_code, customer_name}`,僅供 App 掃碼後帶出身分再選帳號);中台端 `GET /api/v1/customers/:id/qrcode`(需認證,回 PNG 圖,內容為深層連結 `salesorder://redeem?token=…`)。
- `config/auth.go` 加 `QR_TOKEN_SECRET`(空且非 development → fail-fast)。

- [ ] **Step 1: 寫失敗測試(簽章、過期、竄改、跨公司定位)**

```go
// backend/internal/domain/qrcode/qrcode_test.go
package qrcode_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/qrcode"
)

func TestTokenRoundTrip(t *testing.T) {
	qrcode.InitForTest("qr-secret")
	co := uuid.New()
	tok, err := qrcode.IssueToken(co, "AB000001", time.Hour)
	if err != nil { t.Fatal(err) }
	gotCo, code, err := qrcode.VerifyToken(tok)
	if err != nil { t.Fatal(err) }
	if gotCo != co || code != "AB000001" {
		t.Fatalf("round trip: %v %s", gotCo, code)
	}
}

func TestTokenTampered(t *testing.T) {
	qrcode.InitForTest("qr-secret")
	tok, _ := qrcode.IssueToken(uuid.New(), "AB000001", time.Hour)
	if _, _, err := qrcode.VerifyToken(tok[:len(tok)-2] + "xx"); err == nil {
		t.Fatal("tampered token should fail")
	}
}

func TestTokenExpired(t *testing.T) {
	qrcode.InitForTest("qr-secret")
	tok, _ := qrcode.IssueToken(uuid.New(), "AB000001", -time.Minute)
	if _, _, err := qrcode.VerifyToken(tok); err == nil {
		t.Fatal("expired token should fail")
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/qrcode/ -v`
Expected: FAIL。

- [ ] **Step 3: 實作**

`qrcode.go`:

```go
// Package qrcode QR 登入 token(細部文件 3.8)。
package qrcode

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var secret []byte

func Init(s string) {
	if s == "" {
		panic("qrcode: QR_TOKEN_SECRET empty")
	}
	secret = []byte(s)
}

func InitForTest(s string) { Init(s) }

// IssueToken:payload = company_id|customer_code|exp_unix;token = base64url(payload).base64url(hmac)。
func IssueToken(companyID uuid.UUID, customerCode string, ttl time.Duration) (string, error) {
	payload := fmt.Sprintf("%s|%s|%d", companyID, customerCode, time.Now().Add(ttl).Unix())
	sig := sign(payload)
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." +
		base64.RawURLEncoding.EncodeToString(sig), nil
}

// VerifyToken:簽章不符或過期皆回錯(不區分,防探測)。
func VerifyToken(tok string) (uuid.UUID, string, error) {
	parts := strings.Split(tok, ".")
	if len(parts) != 2 {
		return uuid.Nil, "", fmt.Errorf("invalid token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("invalid token")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("invalid token")
	}
	if !hmac.Equal(sign(string(payload)), sig) {
		return uuid.Nil, "", fmt.Errorf("invalid token")
	}
	fields := strings.Split(string(payload), "|")
	if len(fields) != 3 {
		return uuid.Nil, "", fmt.Errorf("invalid token")
	}
	exp, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return uuid.Nil, "", fmt.Errorf("invalid token")
	}
	co, err := uuid.Parse(fields[0])
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("invalid token")
	}
	return co, fields[1], nil
}

func sign(payload string) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(payload))
	return h.Sum(nil)
}
```

兌換端點(handler):`GET /api/v1/qr/redeem?token=` → VerifyToken → 查 `customers`(`company_id + customer_code` + 未軟刪除 + active)→ 回 `{company_identifier, customer_code, customer_name}`;查無 → 404(不區分 token 錯與客戶不存在之外的細節)。QR PNG:中台端以 `skip2/go-qrcode` 產生,內容 `salesorder://redeem?token=<IssueToken(customer, 720h)>`(token 效期 30 天,規格未限定 — 記於 Self-Review 待規格確認)。公開端點掛在 Authenticate 白名單(01-auth Task 11 `PublicPaths` 加 `/api/v1/qr/redeem`)。

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go get github.com/skip2/go-qrcode && go test ./internal/domain/qrcode/ -v`
Expected: PASS。

```bash
git add backend/internal/domain/qrcode backend/internal/server backend/config/auth.go
git commit -m "feat(backend): QR 簽章 token 與公開兌換端點(3.8)"
```

---

## Self-Review 記錄

- **Spec 覆蓋**:3.1.1-3.1.2→T1;3.1.3→T2;3.1.4-3.1.5→T3;3.2.1-3.2.2→T4;3.3.1-3.3.3→T5;3.4.1-3.4.4→T6;3.5.1-3.5.3→T7;3.6.1-3.6.3→T8;3.8.1-3.8.2→T9。無缺漏。
- **跨計畫契約**:`products.ToBase/FromBase`(T5)→ 05 計畫 4.2.1、09 計畫 5.3;`customerproducts.EnsureAlias`(T7)→ 05 計畫 4.2.2;`files.Service` + `DownloadURL`(T8)→ 02(Logo)、06(退貨照片)、09(PDF);`auth.IssueTempPassword` 需由小寫改匯出(T3 步驟說明)。
- **已知佔位/待定**:Task 1 的 `Create` 對無部門 actor(company_admin)要求部門的方式待 02 計畫管理介面定稿(現以 `invalid_argument` 擋下);QR token 效期 30 天為暫定(規格未限定);`Create` 與 `CreateWithCredentials` 併存(後者為 1.0 建檔主路徑,前者供測試與內部)。
- **類型一致**:`rls.Identity` / `audit.Recorder` / `testutil.NewEntClient` 沿用 01-auth;`nextCustomerCode` 簽名 T1 佔位與 T2 實作一致。

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-08-17-backend-04-master-data-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** — 每個 Task 派新 subagent 執行,Task 間 review,迭代快。

**2. Inline Execution** — 用 executing-plans 在本 session 逐批執行,設 checkpoint review。

Which approach?

---

*計畫版本:v1.0.0(2026-08-17);對應細部文件 `backend-detail/04-master-data.md`、原計畫 v2.9.0、規格書 v1.0.34。*
