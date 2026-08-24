# Backend 07 — 通知系統 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 實作 backend Phase 4 通知系統全部機制 — `notification_templates` / `notifications` / `user_devices` schema、範本渲染與語系退回、通知中心查詢與已讀 API、裝置註冊/註銷與失效 token 清理、promo_tags 資料層、FCM / 站內雙通道發送、下單與專屬商品觸發路由、失敗標記不重試。

**Architecture:** 依 `docs/superpowers/plans/backend-detail/07-notifications.md`(下稱「細部文件」,子功能編號 4.3.x / 4.4.x)實作。Chi router + Connect-RPC;通知建檔(`pending`)與觸發方業務**同一 DB 交易**(D18),FCM 外部呼叫於交易提交後執行,失敗僅標 `failed`、**不重試**(D16);通道僅 `fcm` / `in_app`;FCM 以 `notification.Sender` 介面抽象,測試注入 `FakeSender`。

**Tech Stack:** Go 1.25、Ent(entgo.io)、Chi v5、Connect-RPC、pgx/v5、firebase.google.com/go/v4(FCM Admin SDK)、prometheus/client_golang(失敗指標)、testcontainers-go(整合測試)。

**Spec 來源:** 細部文件 `docs/superpowers/plans/backend-detail/07-notifications.md`;共通規則見 `docs/superpowers/plans/backend-detail/00-index.md` §3。

## Global Constraints

- module 路徑:`github.com/salesorder/sales-order-1.0/backend`;所有路徑相對 repo root。
- 軟刪除(D10):`notification_templates` / `user_devices` / `promo_tags` 適用軟刪除 + 部分唯一索引 `WHERE deleted_at IS NULL`;**`notifications` 無 `deleted_at` 欄位,通知記錄不可刪除**(規格 §5.4),不提供 Delete RPC。
- 通道限制(D16):`channel` 僅允許 `fcm` / `in_app`,無 Email;出現其他值一律 `invalid_argument`。
- 交易邊界(D18):通知建檔(status=`pending`)與觸發方業務(下單、專屬商品建立、裝置註冊)**同一 DB 交易**;FCM 呼叫屬外部 I/O,**不得包在 DB 交易內**,於提交後執行;提交後發送失敗**不回滾業務**,只標 `failed`。
- 不重試(D16):`failed` 為終態,1.0 無重試佇列、無定時補發;任何人不得在本 Phase 偷加 retry。
- 狀態機:`pending` → `sent` → `read`;`pending` → `failed`(終態);`pending` → `read`(站內未推播即讀);`read` / `failed` 皆終態不可再轉。
- 稽核:裝置註冊、失效 token 清除寫稽核,與資料異動同一 DB 交易(D18);`audit.Recorder` 介面與 no-op 由 **01-auth-plan Task 14 提供**,DB 實作由 **03-metadicts-audit-plan(Task 2.6)** 接管。敏感欄位(fcm_token 全文)永不入稽核,僅記 token 前 12 碼摘要。
- 錯誤:RPC 層統一 Connect code — `unauthenticated` / `permission_denied` / `not_found` / `failed_precondition` / `invalid_argument` / `already_exists`;內部套件以 sentinel error 表達,RPC 層轉碼。
- RLS:`notification_templates` / `notifications` 以 `company_id` + `department_id` 隔離;`user_devices` 以 `company_id` 隔離;`promo_tags` 以 `company_id` + `department_id` 隔離。RLS 注入(`rls.NewContext` / `rls.WrapDriver` / `testutil.NewEntClientWithRLS`)由 **01-auth-plan Task 3 提供**。
- proto 檔統一置 `backend/proto/v1/<domain>.proto`(照 01-auth 計畫慣例);細部文件所列 `proto/notification/v1/*.proto` 對應至此。
- Ent schema 統一置 `backend/ent/schema/`(Ent codegen 限制,照 01-auth 計畫 Task 1 慣例);細部文件所列 `internal/domain/notifications/schema/` 對應至此。
- migration 採 goose 格式(`-- +goose Up/Down`);檔名序號若與他 Phase 計畫衝突,執行時順延,內容不變。
- FCM 抽象:所有 FCM 呼叫經 `notification.Sender` 介面;測試一律注入 `notification.FakeSender`,不觸網;`FCM_DISABLED=true` 時降級為僅記日誌(僅供開發)。
- 測試:DB 相依測試走 `testutil.NewEntClient`(**01-auth-plan Task 1 提供**);`go test ./...` 必須全綠。
- 每個 Task 結尾 commit;commit message 格式 `feat(backend): …` / `test(backend): …`。

## File Structure

| 檔案 | 職責 | 建立於 |
|---|---|---|
| `backend/ent/schema/notificationtemplate.go` | 範本表 schema(4.3.1) | Task 1 |
| `backend/ent/schema/notification.go` | 通知記錄表 schema(無 `deleted_at`) | Task 1 |
| `backend/ent/schema/userdevice.go` | 裝置表 schema | Task 1 |
| `backend/database/migrations/00013_notifications_rls.sql` | 三表 RLS policy | Task 1 |
| `backend/internal/domain/notifications/render.go` | 範本選取、語系退回、`{{變數}}` 渲染(4.3.2) | Task 2 |
| `backend/internal/domain/notifications/repo.go` | 通知記錄資料存取(4.3.3) | Task 2 |
| `backend/internal/domain/notifications/service.go` | NotificationService / NotificationTemplateService handler | Task 2 |
| `backend/proto/v1/notification.proto` | 通知中心 + Preview proto | Task 2 |
| `backend/internal/domain/notifications/device_service.go` | DeviceService + PurgeInvalidTokens(4.3.4) | Task 3 |
| `backend/proto/v1/device.proto` | DeviceService proto | Task 3 |
| `backend/ent/schema/promotag.go` | promo_tags schema(4.3.5) | Task 3 |
| `backend/database/migrations/00014_promo_tags_rls.sql` | promo_tags RLS policy | Task 3 |
| `backend/internal/notification/fcm.go` | `Sender` 介面、FCM 實作、FakeSender、`Dispatch`(4.4.1) | Task 4 |
| `backend/internal/notification/inapp.go` | 站內通道發送(4.4.2) | Task 4 |
| `backend/internal/notification/failmark.go` | `MarkFailed` 共用失敗標記 + 失敗指標(4.4.5) | Task 4 |
| `backend/internal/notification/triggers/order_created.go` | 下單觸發路由(4.4.3) | Task 5 |
| `backend/internal/notification/triggers/customer_product_created.go` | 專屬商品觸發路由(4.4.4) | Task 5 |

---

### Task 1: notification_templates / notifications / user_devices schema + RLS(細部 4.3.1)

**Files:**
- Create: `backend/ent/schema/notificationtemplate.go`
- Create: `backend/ent/schema/notification.go`
- Create: `backend/ent/schema/userdevice.go`
- Create: `backend/database/migrations/00013_notifications_rls.sql`
- Test: `backend/internal/domain/notifications/schema_test.go`

**Interfaces:**
- Consumes: `testutil.NewEntClient` / `testutil.NewEntClientWithRLS` / `rls.NewContext` / `rls.Identity`(皆由 01-auth-plan Task 1、Task 3 提供)。
- Produces: Ent 實體 `ent.NotificationTemplate` / `ent.Notification` / `ent.UserDevice` 產生碼(謂詞套件 `ent/notificationtemplate`、`ent/notification`、`ent/userdevice`)。欄位:
  - `NotificationTemplate`:`company_id`、`department_id`(nillable,NULL = 公司層範本)、`code`、`name`、`channel`(enum: fcm/in_app)、`subject`、`body`、`locale`、`is_active`(default true)、`created_at`/`updated_at`/`deleted_at`。
  - `Notification`:`company_id`、`department_id`(nillable)、`user_id`、`template_id`(nillable)、`channel`(enum: fcm/in_app)、`title`、`content`、`payload`(JSON `map[string]any`)、`status`(enum: pending/sent/failed/read,default pending)、`failure_reason`(nillable)、`sent_at`/`read_at`(nillable)、`created_at`/`updated_at`;**無 `deleted_at`**。
  - `UserDevice`:`user_id`、`company_id`、`platform`(enum: ios/android/web)、`fcm_token`、`device_name`、`last_seen_at`、`created_at`/`updated_at`/`deleted_at`。

- [ ] **Step 1: 寫失敗測試(唯一索引、無軟刪除、RLS 隔離)**

`backend/internal/domain/notifications/schema_test.go`:

```go
package notifications_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func seedCompanyDept(t *testing.T, c interface {
	Save(context.Context) (interface{}, error)
}) {
}

// setupCo 建一公司一部門,回傳 (companyID, deptID)。直接用 ent client:
func setupCo(t *testing.T, ctx context.Context, c *ent.Client) (uuid.UUID, uuid.UUID) {
	t.Helper()
	co, err := c.Company.Create().SetName("甲").SetIdentifier("co-a").
		SetCustomerCodePrefix("AA").Save(ctx)
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	dept, err := c.Department.Create().SetCompanyID(co.ID).SetName("業務部").Save(ctx)
	if err != nil {
		t.Fatalf("department: %v", err)
	}
	return co.ID, dept.ID
}

func TestTemplatePartialUniqueIndex(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, deptID := setupCo(t, ctx, c)

	mk := func() error {
		_, err := c.NotificationTemplate.Create().
			SetCompanyID(coID).SetDepartmentID(deptID).
			SetCode("order_created").SetName("訂單成立").
			SetChannel("fcm").SetSubject("新訂單").SetBody("{{order_no}}").
			SetLocale("zh-TW").Save(ctx)
		return err
	}
	if err := mk(); err != nil {
		t.Fatalf("first template: %v", err)
	}
	if err := mk(); err == nil {
		t.Fatal("duplicate (company,dept,code,channel,locale) should fail")
	}
	// 軟刪除後可重建
	tmpl, err := c.NotificationTemplate.Query().Only(ctx)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if _, err := c.NotificationTemplate.UpdateOneID(tmpl.ID).
		SetDeletedAt(time.Now()).Save(ctx); err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	if err := mk(); err != nil {
		t.Fatalf("recreate after soft delete: %v", err)
	}
}

func TestTemplateCompanyLevelUniqueIndex(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, _ := setupCo(t, ctx, c)

	mk := func() error {
		_, err := c.NotificationTemplate.Create().
			SetCompanyID(coID). // department_id NULL = 公司層
			SetCode("order_created").SetName("訂單成立").
			SetChannel("in_app").SetSubject("新訂單").SetBody("b").
			SetLocale("zh-TW").Save(ctx)
		return err
	}
	if err := mk(); err != nil {
		t.Fatalf("first company-level template: %v", err)
	}
	if err := mk(); err == nil {
		t.Fatal("duplicate company-level template should fail")
	}
}

func TestUserDeviceTokenUnique(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, _ := setupCo(t, ctx, c)

	mk := func(userID uuid.UUID) error {
		_, err := c.UserDevice.Create().
			SetUserID(userID).SetCompanyID(coID).
			SetPlatform("ios").SetFcmToken("tok-1").
			SetDeviceName("iPhone").SetLastSeenAt(time.Now()).Save(ctx)
		return err
	}
	if err := mk(uuid.New()); err != nil {
		t.Fatalf("first device: %v", err)
	}
	if err := mk(uuid.New()); err == nil {
		t.Fatal("same fcm_token for another user should fail while first is active")
	}
}

func TestNotificationNoSoftDeleteField(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, deptID := setupCo(t, ctx, c)

	n, err := c.Notification.Create().
		SetCompanyID(coID).SetDepartmentID(deptID).SetUserID(uuid.New()).
		SetChannel("in_app").SetTitle("t").SetContent("c").
		SetPayload(map[string]any{"order_id": uuid.New().String()}).Save(ctx)
	if err != nil {
		t.Fatalf("create notification: %v", err)
	}
	// schema 層無 deleted_at:產生碼不含 SetDeletedAt;以反射確認欄位不存在
	if _, ok := reflect.TypeOf(ent.NotificationUpdateOne{}).MethodByName("SetDeletedAt"); ok {
		t.Fatal("notifications must not have deleted_at (規格 §5.4)")
	}
	if n.Status != "pending" {
		t.Fatalf("default status = %q, want pending", n.Status)
	}
}

func TestNotificationRLSIsolation(t *testing.T) {
	c := testutil.NewEntClientWithRLS(t)
	ctx := context.Background()
	superCtx := rls.NewContext(ctx, rls.Identity{UserID: uuid.New(), DataScope: "all"})
	// 建兩家公司(01-auth Task 1 schema)
	coA, err := c.Company.Create().SetName("甲").SetIdentifier("co-a").
		SetCustomerCodePrefix("AA").Save(superCtx)
	if err != nil {
		t.Fatalf("company A: %v", err)
	}
	coB, err := c.Company.Create().SetName("乙").SetIdentifier("co-b").
		SetCustomerCodePrefix("BB").Save(superCtx)
	if err != nil {
		t.Fatalf("company B: %v", err)
	}
	ctxA := rls.NewContext(ctx, rls.Identity{UserID: uuid.New(), CompanyID: coA.ID, DataScope: "company"})
	ctxB := rls.NewContext(ctx, rls.Identity{UserID: uuid.New(), CompanyID: coB.ID, DataScope: "company"})
	if _, err := c.Notification.Create().
		SetCompanyID(coA.ID).SetUserID(uuid.New()).
		SetChannel("in_app").SetTitle("t").SetContent("c").
		SetPayload(map[string]any{}).Save(ctxA); err != nil {
		t.Fatalf("create in A: %v", err)
	}
	nB, err := c.Notification.Query().Count(ctxB)
	if err != nil {
		t.Fatalf("query B: %v", err)
	}
	if nB != 0 {
		t.Fatalf("RLS isolation violated: B sees %d notifications of A", nB)
	}
}
```

測試檔頭部需補的 import(`reflect` 與 `ent`):

```go
import (
	"reflect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
)
```

(移除佔位的 `seedCompanyDept` 空函式 — 僅為讓本 Step 可獨立閱讀;`setupCo` 為實際輔助。)

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/notifications/ -v`
Expected: FAIL — `ent.NotificationTemplate` / `ent.Notification` / `ent.UserDevice` 未定義(編譯失敗)。

- [ ] **Step 3: 實作三個 schema**

`backend/ent/schema/notificationtemplate.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// NotificationTemplate 為通知範本(細部文件 4.3.1)。
// department_id 為 NULL 時代表公司層範本(語系退回終點)。
type NotificationTemplate struct{ ent.Schema }

func (NotificationTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("company_id", uuid.UUID{}),
		field.UUID("department_id", uuid.UUID{}).Optional().Nillable(),
		field.String("code").NotEmpty(), // 範本代號,如 order_created
		field.String("name").NotEmpty(),
		field.Enum("channel").Values("fcm", "in_app"),
		field.String("subject").NotEmpty(),
		field.Text("body").NotEmpty(), // 含 {{變數}} 佔位
		field.String("locale").NotEmpty(),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (NotificationTemplate) Indexes() []ent.Index {
	return []ent.Index{
		// 部門層:同範圍同通道同語系只有一個啟用中範本代號
		index.Fields("company_id", "department_id", "code", "channel", "locale").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL AND department_id IS NOT NULL")),
		// 公司層:department_id 為 NULL 的範本另行唯一(Postgres NULL 不參與相等比較)
		index.Fields("company_id", "code", "channel", "locale").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL AND department_id IS NULL")),
		index.Fields("company_id", "department_id"),
	}
}
```

`backend/ent/schema/notification.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Notification 為通知記錄(細部文件 4.3.1)。
// 無 deleted_at:通知記錄不可刪除(規格 §5.4);status 狀態機見細部文件 4.3.3。
type Notification struct{ ent.Schema }

func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("company_id", uuid.UUID{}),
		field.UUID("department_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("user_id", uuid.UUID{}), // 接收者
		field.UUID("template_id", uuid.UUID{}).Optional().Nillable(), // 手動/降級通知無範本
		field.Enum("channel").Values("fcm", "in_app"),
		field.String("title").NotEmpty(),
		field.Text("content").NotEmpty(),
		field.JSON("payload", map[string]any{}), // 導頁資訊,如 order_id;通知系統不解析
		field.Enum("status").Values("pending", "sent", "failed", "read").Default("pending"),
		field.String("failure_reason").Optional().Nillable(),
		field.Time("sent_at").Optional().Nillable(),
		field.Time("read_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		// 通知中心:本人未讀過濾 + created_at 倒序分頁
		index.Fields("user_id", "status", "created_at"),
		index.Fields("company_id", "department_id"),
	}
}
```

`backend/ent/schema/userdevice.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserDevice 為 App 裝置 FCM token 歸屬(細部文件 4.3.1)。
// 一個 token 同時只屬一個使用者;換帳登入由 4.3.4 重新歸屬。
type UserDevice struct{ ent.Schema }

func (UserDevice) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("company_id", uuid.UUID{}),
		field.Enum("platform").Values("ios", "android", "web"),
		field.String("fcm_token").NotEmpty(),
		field.String("device_name").NotEmpty(),
		field.Time("last_seen_at"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (UserDevice) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("fcm_token").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("user_id"),
		index.Fields("company_id"),
	}
}
```

- [ ] **Step 4: 實作 RLS migration**

`backend/database/migrations/00013_notifications_rls.sql`:

```sql
-- +goose Up
-- 細部文件 4.3.1 步驟 5:通知三表 RLS。
-- 未注入 session variables 時 current_setting(..., true) 回 NULL → fail-closed(01-auth Task 3 慣例)。

ALTER TABLE notification_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_devices ENABLE ROW LEVEL SECURITY;

-- notification_templates:公司隔離;department 級僅見自己部門與公司層(department_id IS NULL)範本
CREATE POLICY tenant_isolation ON notification_templates
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true));

CREATE POLICY department_scope ON notification_templates
  USING (current_setting('app.current_data_scope', true) IN ('all', 'company')
         OR department_id IS NULL
         OR department_id::text = current_setting('app.current_department_id', true));

-- notifications:公司 + 部門隔離(接收端再於應用層限定本人 user_id)
CREATE POLICY tenant_isolation ON notifications
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true));

CREATE POLICY department_scope ON notifications
  USING (current_setting('app.current_data_scope', true) IN ('all', 'company')
         OR department_id IS NULL
         OR department_id::text = current_setting('app.current_department_id', true));

-- user_devices:公司隔離(細部文件 4.3.1 步驟 5)
CREATE POLICY tenant_isolation ON user_devices
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true));

-- +goose Down
DROP POLICY IF EXISTS tenant_isolation ON user_devices;
DROP POLICY IF EXISTS department_scope ON notifications;
DROP POLICY IF EXISTS tenant_isolation ON notifications;
DROP POLICY IF EXISTS department_scope ON notification_templates;
DROP POLICY IF EXISTS tenant_isolation ON notification_templates;
ALTER TABLE user_devices DISABLE ROW LEVEL SECURITY;
ALTER TABLE notifications DISABLE ROW LEVEL SECURITY;
ALTER TABLE notification_templates DISABLE ROW LEVEL SECURITY;
```

- [ ] **Step 5: 產生碼 + 跑測試確認通過**

Run: `cd backend && go generate ./ent/ && goose -dir database/migrations postgres "$(go run ./cmd/devdb -dsn)" up` (或以測試庫手動套用 migration 後) `go test ./internal/domain/notifications/ -v`

RLS 測試(`TestNotificationRLSIsolation`)執行前,須先對測試庫套用 `00013_notifications_rls.sql`;於 `testutil.NewEntClientWithRLS` 中已含 migration 套用邏輯(01-auth-plan Task 3 提供)時直接通過,否則在該 helper 補套用本檔 — 以 01-auth 計畫既有機制為準,不另建第二套。

Expected: PASS — 5 個測試全綠。

- [ ] **Step 6: Commit**

```bash
git add backend/ent/schema/notificationtemplate.go backend/ent/schema/notification.go backend/ent/schema/userdevice.go backend/ent backend/database/migrations/00013_notifications_rls.sql backend/internal/domain/notifications
git commit -m "feat(backend): 通知三表 schema 與 RLS policy(4.3.1)"
```

---

### Task 2: 範本渲染 + 通知記錄/已讀 API(細部 4.3.2–4.3.3)

**Files:**
- Create: `backend/internal/domain/notifications/render.go`
- Create: `backend/internal/domain/notifications/repo.go`
- Create: `backend/internal/domain/notifications/service.go`
- Create: `backend/proto/v1/notification.proto`
- Test: `backend/internal/domain/notifications/render_test.go`
- Test: `backend/internal/domain/notifications/service_test.go`

**Interfaces:**
- Consumes: Task 1 三表;`rls.Identity` / `rls.FromContext`(01-auth-plan Task 3、Task 11;`IsPrimary` 欄位由 01-auth-plan Task 13 擴充)。
- Produces:
  - `notifications.TemplateQuery{CompanyID uuid.UUID; DepartmentID *uuid.UUID; Code, Channel, Locale string}`
  - `notifications.Rendered{TemplateID uuid.UUID; Title, Content string; Missing []string}`
  - `notifications.Render(ctx, db *ent.Client, q TemplateQuery, vars map[string]string) (Rendered, error)` — sentinel error:`ErrTemplateNotFound`、`ErrInvalidArgument`(RPC 層轉 `not_found` / `invalid_argument`);`notifications.DefaultLocale = "zh-TW"`。
  - `notifications.Repo`:`CreateOne(ctx, tx *ent.Tx, p CreateParams) (uuid.UUID, error)`;`ListByUser(ctx, db, userID, page, pageSize, unreadOnly) ([]*ent.Notification, total int64, unread int64, err error)`;`MarkReadBatch(ctx, db, userID, ids []uuid.UUID) (int, error)`;`UnreadCount(ctx, db, userID) (int64, error)`。
  - Connect-RPC `NotificationService`:`List` / `MarkRead` / `UnreadCount`;`NotificationTemplateService`:`Preview`。
  - 後續 Task 4/5 以 `Render` + `Repo.CreateOne` 建通知;`MarkReadBatch` 狀態機為全域唯一已讀入口。

- [ ] **Step 1: 寫失敗測試(渲染與語系退回)**

`backend/internal/domain/notifications/render_test.go`:

```go
package notifications_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/domain/notifications"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func seedTemplate(t *testing.T, ctx context.Context, c *ent.Client,
	coID uuid.UUID, deptID *uuid.UUID, code, channel, locale, subject, body string) {
	t.Helper()
	b := c.NotificationTemplate.Create().
		SetCompanyID(coID).SetCode(code).SetName(code).
		SetChannel(channel).SetSubject(subject).SetBody(body).SetLocale(locale)
	if deptID != nil {
		b.SetDepartmentID(*deptID)
	}
	if _, err := b.Save(ctx); err != nil {
		t.Fatalf("seed template: %v", err)
	}
}

func TestRenderFullSubstitution(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, deptID := setupCo(t, ctx, c)
	seedTemplate(t, ctx, c, coID, &deptID, "order_created", "in_app", "zh-TW",
		"{{customer_name}} 新訂單", "訂單 {{order_no}} 共 {{item_count}} 項")

	r, err := notifications.Render(ctx, c, notifications.TemplateQuery{
		CompanyID: coID, DepartmentID: &deptID,
		Code: "order_created", Channel: "in_app", Locale: "zh-TW",
	}, map[string]string{"customer_name": "好市多", "order_no": "SO-001", "item_count": "3"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if r.Title != "好市多 新訂單" || r.Content != "訂單 SO-001 共 3 項" {
		t.Fatalf("got %q / %q", r.Title, r.Content)
	}
	if len(r.Missing) != 0 {
		t.Fatalf("unexpected missing: %v", r.Missing)
	}
}

func TestRenderMissingVarPreserved(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, deptID := setupCo(t, ctx, c)
	seedTemplate(t, ctx, c, coID, &deptID, "order_created", "in_app", "zh-TW",
		"新訂單", "訂單 {{order_no}} 金額 {{amount}}")

	r, err := notifications.Render(ctx, c, notifications.TemplateQuery{
		CompanyID: coID, DepartmentID: &deptID,
		Code: "order_created", Channel: "in_app", Locale: "zh-TW",
	}, map[string]string{"order_no": "SO-002", "unused": "x"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if r.Content != "訂單 SO-002 金額 {{amount}}" {
		t.Fatalf("missing var should be preserved, got %q", r.Content)
	}
	if len(r.Missing) != 1 || r.Missing[0] != "amount" {
		t.Fatalf("missing = %v, want [amount]", r.Missing)
	}
}

func TestRenderLocaleFallbackChain(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, deptID := setupCo(t, ctx, c)
	// 只有預設語系部門範本 + 公司層 en 範本;請求 en-US 部門範本
	seedTemplate(t, ctx, c, coID, &deptID, "order_created", "fcm", "zh-TW", "部門預設", "b1")
	seedTemplate(t, ctx, c, coID, nil, "order_created", "fcm", "en", "company en", "b2")

	// 1. 部門無 en-US → 退回部門預設語系 zh-TW
	r, err := notifications.Render(ctx, c, notifications.TemplateQuery{
		CompanyID: coID, DepartmentID: &deptID,
		Code: "order_created", Channel: "fcm", Locale: "en-US",
	}, nil)
	if err != nil || r.Title != "部門預設" {
		t.Fatalf("fallback to dept default locale failed: %q err=%v", r.Title, err)
	}

	// 2. 另一 code 只有公司層範本 → 退回公司層
	seedTemplate(t, ctx, c, coID, nil, "customer_product_created", "fcm", "zh-TW", "公司層", "b3")
	r, err = notifications.Render(ctx, c, notifications.TemplateQuery{
		CompanyID: coID, DepartmentID: &deptID,
		Code: "customer_product_created", Channel: "fcm", Locale: "en-US",
	}, nil)
	if err != nil || r.Title != "公司層" {
		t.Fatalf("fallback to company level failed: %q err=%v", r.Title, err)
	}

	// 3. 全無範本 → ErrTemplateNotFound
	_, err = notifications.Render(ctx, c, notifications.TemplateQuery{
		CompanyID: coID, DepartmentID: &deptID,
		Code: "nonexistent", Channel: "fcm", Locale: "zh-TW",
	}, nil)
	if !errors.Is(err, notifications.ErrTemplateNotFound) {
		t.Fatalf("want ErrTemplateNotFound, got %v", err)
	}
}

func TestRenderFCMTruncation(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, deptID := setupCo(t, ctx, c)
	long := strings.Repeat("長", 600)
	seedTemplate(t, ctx, c, coID, &deptID, "order_created", "fcm", "zh-TW", long, long)

	r, err := notifications.Render(ctx, c, notifications.TemplateQuery{
		CompanyID: coID, DepartmentID: &deptID,
		Code: "order_created", Channel: "fcm", Locale: "zh-TW",
	}, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if utf8.RuneCountInString(r.Title) > 100 || utf8.RuneCountInString(r.Content) > 500 {
		t.Fatalf("fcm channel should truncate: title=%d body=%d runes",
			utf8.RuneCountInString(r.Title), utf8.RuneCountInString(r.Content))
	}
	// in_app 不截斷
	seedTemplate(t, ctx, c, coID, &deptID, "order_created", "in_app", "zh-TW", "s", long)
	r, err = notifications.Render(ctx, c, notifications.TemplateQuery{
		CompanyID: coID, DepartmentID: &deptID,
		Code: "order_created", Channel: "in_app", Locale: "zh-TW",
	}, nil)
	if err != nil || utf8.RuneCountInString(r.Content) != 600 {
		t.Fatalf("in_app should not truncate: %d runes err=%v", utf8.RuneCountInString(r.Content), err)
	}
}

func TestRenderInvalidChannel(t *testing.T) {
	c := testutil.NewEntClient(t)
	_, err := notifications.Render(context.Background(), c, notifications.TemplateQuery{
		CompanyID: uuid.New(), Code: "x", Channel: "email", Locale: "zh-TW",
	}, nil)
	if !errors.Is(err, notifications.ErrInvalidArgument) {
		t.Fatalf("email channel must be rejected (D16), got %v", err)
	}
}
```

測試需補 import:`errors`、`strings`、`unicode/utf8`、`ent`。

- [ ] **Step 2: 寫失敗測試(通知記錄 List / MarkRead / UnreadCount 狀態機)**

`backend/internal/domain/notifications/service_test.go`:

```go
package notifications_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	v1 "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/notifications"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func identityCtx(ctx context.Context, userID, coID uuid.UUID, deptID *uuid.UUID, isPrimary bool) context.Context {
	return rls.NewContext(ctx, rls.Identity{
		UserID: userID, CompanyID: coID, DepartmentID: deptID,
		DataScope: "company", IsPrimary: isPrimary,
	})
}

func seedNotification(t *testing.T, ctx context.Context, c *ent.Client,
	coID uuid.UUID, deptID *uuid.UUID, userID uuid.UUID, status string) *ent.Notification {
	t.Helper()
	b := c.Notification.Create().
		SetCompanyID(coID).SetUserID(userID).
		SetChannel("in_app").SetTitle("t").SetContent("c").
		SetPayload(map[string]any{}).SetStatus(status)
	if deptID != nil {
		b.SetDepartmentID(*deptID)
	}
	n, err := b.Save(ctx)
	if err != nil {
		t.Fatalf("seed notification: %v", err)
	}
	return n
}

func newService(t *testing.T) (*notifications.NotificationServiceHandler, *ent.Client) {
	t.Helper()
	c := testutil.NewEntClient(t)
	return notifications.NewNotificationServiceHandler(c), c
}

func TestListOnlyOwnNotifications(t *testing.T) {
	svc, c := newService(t)
	ctx := context.Background()
	coID, deptID := setupCo(t, ctx, c)
	me, other := uuid.New(), uuid.New()
	seedNotification(t, ctx, c, coID, &deptID, me, "sent")
	seedNotification(t, ctx, c, coID, &deptID, me, "read")
	seedNotification(t, ctx, c, coID, &deptID, other, "sent")

	resp, err := svc.List(identityCtx(ctx, me, coID, &deptID, false),
		connect.NewRequest(&v1.ListNotificationsRequest{Page: 1, PageSize: 10}))
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.Msg.Total != 2 || len(resp.Msg.Notifications) != 2 {
		t.Fatalf("total = %d, want 2 (own only)", resp.Msg.Total)
	}
	if resp.Msg.UnreadCount != 1 {
		t.Fatalf("unread_count = %d, want 1", resp.Msg.UnreadCount)
	}

	// unread_only
	resp, err = svc.List(identityCtx(ctx, me, coID, &deptID, false),
		connect.NewRequest(&v1.ListNotificationsRequest{Page: 1, PageSize: 10, UnreadOnly: true}))
	if err != nil {
		t.Fatalf("list unread: %v", err)
	}
	if resp.Msg.Total != 1 || resp.Msg.Notifications[0].Status != "sent" {
		t.Fatalf("unread_only failed: %+v", resp.Msg)
	}
}

func TestMarkReadStateMachine(t *testing.T) {
	svc, c := newService(t)
	ctx := context.Background()
	coID, deptID := setupCo(t, ctx, c)
	me := uuid.New()
	sent := seedNotification(t, ctx, c, coID, &deptID, me, "sent")
	pending := seedNotification(t, ctx, c, coID, &deptID, me, "pending")
	failed := seedNotification(t, ctx, c, coID, &deptID, me, "failed")

	// sent + pending → read
	resp, err := svc.MarkRead(identityCtx(ctx, me, coID, &deptID, false),
		connect.NewRequest(&v1.MarkReadRequest{
			NotificationIds: []string{sent.ID.String(), pending.ID.String()},
		}))
	if err != nil {
		t.Fatalf("mark read: %v", err)
	}
	if resp.Msg.MarkedCount != 2 {
		t.Fatalf("marked = %d, want 2", resp.Msg.MarkedCount)
	}
	got, _ := c.Notification.Get(ctx, sent.ID)
	if got.Status != "read" || got.ReadAt == nil {
		t.Fatalf("sent -> read failed: status=%s read_at=%v", got.Status, got.ReadAt)
	}

	// 冪等:重複呼叫不報錯、marked=0
	resp, err = svc.MarkRead(identityCtx(ctx, me, coID, &deptID, false),
		connect.NewRequest(&v1.MarkReadRequest{NotificationIds: []string{sent.ID.String()}}))
	if err != nil || resp.Msg.MarkedCount != 0 {
		t.Fatalf("idempotent mark read failed: %v marked=%d", err, resp.Msg.MarkedCount)
	}

	// failed 不可轉 read → failed_precondition
	_, err = svc.MarkRead(identityCtx(ctx, me, coID, &deptID, false),
		connect.NewRequest(&v1.MarkReadRequest{NotificationIds: []string{failed.ID.String()}}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("failed -> read must be failed_precondition, got %v", err)
	}

	// 他人通知 ID → not_found(不洩漏存在性)
	other := seedNotification(t, ctx, c, coID, &deptID, uuid.New(), "sent")
	_, err = svc.MarkRead(identityCtx(ctx, me, coID, &deptID, false),
		connect.NewRequest(&v1.MarkReadRequest{NotificationIds: []string{other.ID.String()}}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("others' notification must be not_found, got %v", err)
	}

	// 主帳號 → permission_denied(D22)
	_, err = svc.MarkRead(identityCtx(ctx, me, coID, &deptID, true),
		connect.NewRequest(&v1.MarkReadRequest{NotificationIds: []string{sent.ID.String()}}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("primary account must be permission_denied, got %v", err)
	}

	// 超過單批上限 → invalid_argument
	ids := make([]string, 101)
	for i := range ids {
		ids[i] = uuid.NewString()
	}
	_, err = svc.MarkRead(identityCtx(ctx, me, coID, &deptID, false),
		connect.NewRequest(&v1.MarkReadRequest{NotificationIds: ids}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("over-limit batch must be invalid_argument, got %v", err)
	}
}

func TestUnreadCount(t *testing.T) {
	svc, c := newService(t)
	ctx := context.Background()
	coID, deptID := setupCo(t, ctx, c)
	me := uuid.New()
	seedNotification(t, ctx, c, coID, &deptID, me, "pending")
	seedNotification(t, ctx, c, coID, &deptID, me, "sent")
	seedNotification(t, ctx, c, coID, &deptID, me, "read")
	seedNotification(t, ctx, c, coID, &deptID, me, "failed")

	resp, err := svc.UnreadCount(identityCtx(ctx, me, coID, &deptID, false),
		connect.NewRequest(&v1.UnreadCountRequest{}))
	if err != nil {
		t.Fatalf("unread count: %v", err)
	}
	if resp.Msg.Count != 2 {
		t.Fatalf("count = %d, want 2 (pending + sent)", resp.Msg.Count)
	}
}

func TestPreviewTemplate(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, deptID := setupCo(t, ctx, c)
	tmpl, err := c.NotificationTemplate.Create().
		SetCompanyID(coID).SetDepartmentID(deptID).
		SetCode("order_created").SetName("訂單成立").SetChannel("in_app").
		SetSubject("{{customer_name}} 新訂單").SetBody("訂單 {{order_no}}").
		SetLocale("zh-TW").Save(ctx)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	svc := notifications.NewNotificationTemplateServiceHandler(c)
	resp, err := svc.Preview(identityCtx(ctx, uuid.New(), coID, &deptID, false),
		connect.NewRequest(&v1.PreviewTemplateRequest{
			TemplateId: tmpl.ID.String(),
			Variables:  map[string]string{"customer_name": "好市多"},
		}))
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if resp.Msg.Title != "好市多 新訂單" || resp.Msg.Content != "訂單 {{order_no}}" {
		t.Fatalf("preview got %q / %q", resp.Msg.Title, resp.Msg.Content)
	}
	if len(resp.Msg.MissingVariables) != 1 || resp.Msg.MissingVariables[0] != "order_no" {
		t.Fatalf("missing = %v", resp.Msg.MissingVariables)
	}
}
```

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/notifications/ -run 'TestRender|TestList|TestMarkRead|TestUnreadCount|TestPreview' -v`
Expected: FAIL — `notifications.Render` / `NewNotificationServiceHandler` 未定義、`gen/proto/v1` 無通知訊息(編譯失敗)。

- [ ] **Step 4: 實作 render.go**

`backend/internal/domain/notifications/render.go`:

```go
// Package notifications 提供通知範本渲染與通知中心 API(細部文件 4.3.2、4.3.3)。
package notifications

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	enttemplate "github.com/salesorder/sales-order-1.0/backend/ent/notificationtemplate"
)

// DefaultLocale 為範本語系退回鏈的預設語系。
const DefaultLocale = "zh-TW"

// FCM 訊息長度上限(rune);in_app 不截斷,由前端截顯。
const (
	FCMTitleMaxRunes = 100
	FCMBodyMaxRunes  = 500
)

var (
	// ErrTemplateNotFound:範本不存在或已停用;RPC 層轉 not_found,
	// 觸發方降級規則見細部文件 4.3.2 錯誤處理。
	ErrTemplateNotFound = errors.New("notifications: template not found")
	// ErrInvalidArgument:code/channel 不合法;RPC 層轉 invalid_argument。
	ErrInvalidArgument = errors.New("notifications: invalid argument")
)

// TemplateQuery 為範本選取條件。DepartmentID 為 nil 表示直接查公司層範本。
type TemplateQuery struct {
	CompanyID    uuid.UUID
	DepartmentID *uuid.UUID
	Code         string
	Channel      string // fcm | in_app
	Locale       string // 空字串視為 DefaultLocale
}

// Rendered 為渲染結果。Missing 為範本中出現但 variables 未提供的佔位符清單。
type Rendered struct {
	TemplateID uuid.UUID
	Title      string
	Content    string
	Missing    []string
}

var placeholderRe = regexp.MustCompile(`\{\{([A-Za-z0-9_]+)\}\}`)

// Render 選取範本並替換 {{變數}};純計算 + 唯讀查詢,不寫庫(細部文件 4.3.2 步驟 5)。
// 語系退回鏈:(部門,指定語系) → (部門,預設語系) → (公司層,指定語系) → (公司層,預設語系) → ErrTemplateNotFound。
func Render(ctx context.Context, db *ent.Client, q TemplateQuery, vars map[string]string) (Rendered, error) {
	if q.Code == "" || (q.Channel != "fcm" && q.Channel != "in_app") {
		return Rendered{}, fmt.Errorf("%w: code/channel", ErrInvalidArgument)
	}
	if q.Locale == "" {
		q.Locale = DefaultLocale
	}
	tmpl, err := selectTemplate(ctx, db, q)
	if err != nil {
		return Rendered{}, err
	}
	title, missT := substitute(tmpl.Subject, vars)
	body, missB := substitute(tmpl.Body, vars)
	missing := unionSorted(missT, missB)
	if len(missing) > 0 {
		// 缺漏變數保留原文、記警告日誌,不阻斷發送(細部文件 4.3.2 步驟 3)
		log.Printf("notifications: template %s(%s/%s) missing variables %v",
			q.Code, q.Channel, q.Locale, missing)
	}
	if q.Channel == "fcm" {
		title = truncateRunes(title, FCMTitleMaxRunes)
		body = truncateRunes(body, FCMBodyMaxRunes)
	}
	return Rendered{TemplateID: tmpl.ID, Title: title, Content: body, Missing: missing}, nil
}

// selectTemplate 依退回鏈查詢啟用中且未軟刪除的範本。
func selectTemplate(ctx context.Context, db *ent.Client, q TemplateQuery) (*ent.NotificationTemplate, error) {
	type key struct {
		dept   *uuid.UUID
		locale string
	}
	chain := []key{}
	if q.DepartmentID != nil {
		chain = append(chain, key{q.DepartmentID, q.Locale})
		if q.Locale != DefaultLocale {
			chain = append(chain, key{q.DepartmentID, DefaultLocale})
		}
	}
	chain = append(chain, key{nil, q.Locale})
	if q.Locale != DefaultLocale {
		chain = append(chain, key{nil, DefaultLocale})
	}
	for _, k := range chain {
		preds := []predicate.NotificationTemplate{
			enttemplate.CompanyIDEQ(q.CompanyID),
			enttemplate.CodeEQ(q.Code),
			enttemplate.ChannelEQ(q.Channel),
			enttemplate.LocaleEQ(k.locale),
			enttemplate.IsActiveEQ(true),
			enttemplate.DeletedAtIsNil(),
		}
		if k.dept != nil {
			preds = append(preds, enttemplate.DepartmentIDEQ(*k.dept))
		} else {
			preds = append(preds, enttemplate.DepartmentIDIsNil())
		}
		tmpl, err := db.NotificationTemplate.Query().Where(preds...).First(ctx)
		if err == nil {
			return tmpl, nil
		}
		if !ent.IsNotFound(err) {
			return nil, fmt.Errorf("notifications: select template: %w", err)
		}
	}
	return nil, ErrTemplateNotFound
}

// substitute 全字吻合替換 {{變數名}}(變數名僅英數與底線,由 placeholderRe 保證);
// variables 未提供者保留原文,回傳缺漏清單。輸出即純文字,不再二次解析(注入防護)。
func substitute(tmpl string, vars map[string]string) (string, []string) {
	missing := map[string]bool{}
	out := placeholderRe.ReplaceAllStringFunc(tmpl, func(m string) string {
		name := placeholderRe.FindStringSubmatch(m)[1]
		if v, ok := vars[name]; ok {
			return v
		}
		missing[name] = true
		return m
	})
	names := make([]string, 0, len(missing))
	for n := range missing {
		names = append(names, n)
	}
	return out, names
}

func unionSorted(a, b []string) []string {
	set := map[string]bool{}
	for _, s := range append(a, b...) {
		set[s] = true
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max])
}
```

(`selectTemplate` 需補 import:`"entgo.io/ent/dialect/predicate"`;`enttemplate.ChannelEQ` 參數型別為產生碼 enum `notificationtemplate.Channel`,字串字面量以 `enttemplate.Channel(q.Channel)` 轉型 — 實作時照產生碼簽名調整。)

- [ ] **Step 5: 實作 repo.go 與 proto**

`backend/internal/domain/notifications/repo.go`:

```go
package notifications

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entnotification "github.com/salesorder/sales-order-1.0/backend/ent/notification"
)

// MaxMarkReadBatch 為 MarkRead 單批 ID 上限。
const MaxMarkReadBatch = 100

var (
	// ErrNotFound:通知不存在或不屬當前使用者(不洩漏存在性);RPC 層轉 not_found。
	ErrNotFound = errors.New("notifications: not found")
	// ErrFailedPrecondition:狀態機不允許的轉移;RPC 層轉 failed_precondition。
	ErrFailedPrecondition = errors.New("notifications: failed precondition")
)

// CreateParams 為建檔參數;TemplateID 可為 nil(手動/降級通知,細部文件 4.3.1 步驟 4)。
type CreateParams struct {
	CompanyID    uuid.UUID
	DepartmentID *uuid.UUID
	UserID       uuid.UUID
	TemplateID   *uuid.UUID
	Channel      string
	Title        string
	Content      string
	Payload      map[string]any
}

// Repo 為通知記錄資料存取;無 Delete 方法(通知記錄不可刪除,規格 §5.4)。
type Repo struct{}

// CreateOne 於觸發方交易內建一筆 pending 通知(D18 同交易)。
func (Repo) CreateOne(ctx context.Context, tx *ent.Tx, p CreateParams) (uuid.UUID, error) {
	if p.Channel != "fcm" && p.Channel != "in_app" {
		return uuid.Nil, fmt.Errorf("%w: channel", ErrInvalidArgument)
	}
	b := tx.Notification.Create().
		SetCompanyID(p.CompanyID).SetUserID(p.UserID).
		SetChannel(entnotification.Channel(p.Channel)).
		SetTitle(p.Title).SetContent(p.Content).SetPayload(p.Payload)
	if p.DepartmentID != nil {
		b.SetDepartmentID(*p.DepartmentID)
	}
	if p.TemplateID != nil {
		b.SetTemplateID(*p.TemplateID)
	}
	n, err := b.Save(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("notifications: create: %w", err)
	}
	return n.ID, nil
}

// ListByUser 僅回傳本人通知,created_at 倒序分頁;unreadOnly 過濾 status IN (pending, sent)。
// 同函式回傳 unread_count(細部文件 4.3.3 步驟 1)。
func (Repo) ListByUser(ctx context.Context, db *ent.Client, userID uuid.UUID,
	page, pageSize int, unreadOnly bool) ([]*ent.Notification, int64, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	base := db.Notification.Query().Where(entnotification.UserIDEQ(userID))
	q := base.Clone()
	if unreadOnly {
		q = q.Where(entnotification.StatusIn(
			entnotification.StatusPending, entnotification.StatusSent))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("notifications: count: %w", err)
	}
	items, err := q.
		Order(ent.Desc(entnotification.FieldCreatedAt)).
		Limit(pageSize).Offset((page - 1) * pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("notifications: list: %w", err)
	}
	unread, err := base.Clone().Where(entnotification.StatusIn(
		entnotification.StatusPending, entnotification.StatusSent)).Count(ctx)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("notifications: unread count: %w", err)
	}
	return items, int64(total), int64(unread), nil
}

// UnreadCount 供通知鈴角標輪詢。
func (Repo) UnreadCount(ctx context.Context, db *ent.Client, userID uuid.UUID) (int64, error) {
	n, err := db.Notification.Query().Where(
		entnotification.UserIDEQ(userID),
		entnotification.StatusIn(entnotification.StatusPending, entnotification.StatusSent),
	).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("notifications: unread count: %w", err)
	}
	return int64(n), nil
}

// MarkReadBatch 於單一交易將本人 pending/sent 通知轉 read 並寫 read_at(細部文件 4.3.3 步驟 2-3)。
// 狀態機:failed 不允許轉 read;read 冪等略過;他人 ID 視同不存在(ErrNotFound,不洩漏存在性);
// read/failed 皆終態,無後續轉移路徑。回傳實際異動筆數。
func (Repo) MarkReadBatch(ctx context.Context, db *ent.Client, userID uuid.UUID, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 || len(ids) > MaxMarkReadBatch {
		return 0, fmt.Errorf("%w: ids length %d", ErrInvalidArgument, len(ids))
	}
	tx, err := db.Tx(ctx)
	if err != nil {
		return 0, fmt.Errorf("notifications: tx: %w", err)
	}
	mine, err := tx.Notification.Query().Where(
		entnotification.IDEQ... , // 見下行實際寫法
	).All(ctx)
	_ = mine
	// 實際查詢:本人 + ID 集合
	mine, err = tx.Notification.Query().Where(
		entnotification.UserIDEQ(userID),
		entnotification.IDIn(ids...),
	).All(ctx)
	if err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("notifications: query: %w", err)
	}
	if len(mine) != len(ids) {
		_ = tx.Rollback()
		return 0, ErrNotFound
	}
	for _, n := range mine {
		if n.Status == entnotification.StatusFailed {
			_ = tx.Rollback()
			return 0, fmt.Errorf("%w: notification %s is failed", ErrFailedPrecondition, n.ID)
		}
	}
	marked, err := tx.Notification.Update().Where(
		entnotification.UserIDEQ(userID),
		entnotification.IDIn(ids...),
		entnotification.StatusIn(entnotification.StatusPending, entnotification.StatusSent),
	).
		SetStatus(entnotification.StatusRead).
		SetReadAt(time.Now()).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("notifications: mark read: %w", err)
	}
	return marked, tx.Commit()
}
```

(移除上方 `entnotification.IDEQ...` 佔位查詢段 — 僅為說明查詢形狀;以第二段實際查詢為準。)

`backend/proto/v1/notification.proto`:

```proto
syntax = "proto3";
package salesorder.v1;

import "google/protobuf/timestamp.proto";

service NotificationService {
  rpc List(ListNotificationsRequest) returns (ListNotificationsResponse);
  rpc MarkRead(MarkReadRequest) returns (MarkReadResponse);
  rpc UnreadCount(UnreadCountRequest) returns (UnreadCountResponse);
}

service NotificationTemplateService {
  rpc Preview(PreviewTemplateRequest) returns (PreviewTemplateResponse);
}

message Notification {
  string id = 1;
  string channel = 2;
  string title = 3;
  string content = 4;
  string payload_json = 5;
  string status = 6;
  google.protobuf.Timestamp sent_at = 7;
  google.protobuf.Timestamp read_at = 8;
  google.protobuf.Timestamp created_at = 9;
}

message ListNotificationsRequest {
  int32 page = 1;
  int32 page_size = 2;
  bool unread_only = 3;
}
message ListNotificationsResponse {
  repeated Notification notifications = 1;
  int64 total = 2;
  int64 unread_count = 3;
}

message MarkReadRequest { repeated string notification_ids = 1; }
message MarkReadResponse { int32 marked_count = 1; }

message UnreadCountRequest {}
message UnreadCountResponse { int64 count = 1; }

message PreviewTemplateRequest {
  string template_id = 1;
  map<string, string> variables = 2;
}
message PreviewTemplateResponse {
  string title = 1;
  string content = 2;
  repeated string missing_variables = 3;
}
```

- [ ] **Step 6: 實作 service.go(Connect handler)**

`backend/internal/domain/notifications/service.go`:

```go
package notifications

import (
	"context"
	"encoding/json"
	"errors"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	v1 "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// NotificationServiceHandler 實作通知中心 API(細部文件 4.3.3)。
type NotificationServiceHandler struct {
	db   *ent.Client
	repo Repo
}

func NewNotificationServiceHandler(db *ent.Client) *NotificationServiceHandler {
	return &NotificationServiceHandler{db: db}
}

// actor 取當前身分;未登入 → unauthenticated;主帳號無業務身分(D22)→ permission_denied。
func actor(ctx context.Context) (rls.Identity, error) {
	id, ok := rls.FromContext(ctx)
	if !ok {
		return rls.Identity{}, connect.NewError(connect.CodeUnauthenticated, errors.New("not authenticated"))
	}
	if id.IsPrimary {
		return rls.Identity{}, connect.NewError(connect.CodePermissionDenied, errors.New("primary account has no business identity"))
	}
	return id, nil
}

func toProto(n *ent.Notification) (*v1.Notification, error) {
	payload, err := json.Marshal(n.Payload)
	if err != nil {
		return nil, err
	}
	msg := &v1.Notification{
		Id:          n.ID.String(),
		Channel:     string(n.Channel),
		Title:       n.Title,
		Content:     n.Content,
		PayloadJson: string(payload),
		Status:      string(n.Status),
		CreatedAt:   timestamppb.New(n.CreatedAt),
	}
	if n.SentAt != nil {
		msg.SentAt = timestamppb.New(*n.SentAt)
	}
	if n.ReadAt != nil {
		msg.ReadAt = timestamppb.New(*n.ReadAt)
	}
	return msg, nil
}

func (h *NotificationServiceHandler) List(ctx context.Context,
	req *connect.Request[v1.ListNotificationsRequest]) (*connect.Response[v1.ListNotificationsResponse], error) {
	id, err := actor(ctx)
	if err != nil {
		return nil, err
	}
	items, total, unread, err := h.repo.ListByUser(ctx, h.db, id.UserID,
		int(req.Msg.Page), int(req.Msg.PageSize), req.Msg.UnreadOnly)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	msgs := make([]*v1.Notification, 0, len(items))
	for _, n := range items {
		m, err := toProto(n)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		msgs = append(msgs, m)
	}
	return connect.NewResponse(&v1.ListNotificationsResponse{
		Notifications: msgs, Total: total, UnreadCount: unread,
	}), nil
}

func (h *NotificationServiceHandler) MarkRead(ctx context.Context,
	req *connect.Request[v1.MarkReadRequest]) (*connect.Response[v1.MarkReadResponse], error) {
	id, err := actor(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(req.Msg.NotificationIds))
	for _, s := range req.Msg.NotificationIds {
		u, err := uuid.Parse(s)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid notification id"))
		}
		ids = append(ids, u)
	}
	marked, err := h.repo.MarkReadBatch(ctx, h.db, id.UserID, ids)
	switch {
	case errors.Is(err, ErrInvalidArgument):
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, ErrNotFound):
		return nil, connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, ErrFailedPrecondition):
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	case err != nil:
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.MarkReadResponse{MarkedCount: int32(marked)}), nil
}

func (h *NotificationServiceHandler) UnreadCount(ctx context.Context,
	_ *connect.Request[v1.UnreadCountRequest]) (*connect.Response[v1.UnreadCountResponse], error) {
	id, err := actor(ctx)
	if err != nil {
		return nil, err
	}
	n, err := h.repo.UnreadCount(ctx, h.db, id.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.UnreadCountResponse{Count: n}), nil
}

// NotificationTemplateServiceHandler 實作範本 Preview(細部文件 4.3.2 介面)。
type NotificationTemplateServiceHandler struct {
	db *ent.Client
}

func NewNotificationTemplateServiceHandler(db *ent.Client) *NotificationTemplateServiceHandler {
	return &NotificationTemplateServiceHandler{db: db}
}

func (h *NotificationTemplateServiceHandler) Preview(ctx context.Context,
	req *connect.Request[v1.PreviewTemplateRequest]) (*connect.Response[v1.PreviewTemplateResponse], error) {
	if _, err := actor(ctx); err != nil {
		return nil, err
	}
	tmplID, err := uuid.Parse(req.Msg.TemplateId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid template id"))
	}
	tmpl, err := h.db.NotificationTemplate.Get(ctx, tmplID)
	if ent.IsNotFound(err) {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("template not found"))
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// Preview 直接對指定範本渲染(不走退回鏈);缺漏變數保留原文並回報清單
	title, missT := substitute(tmpl.Subject, req.Msg.Variables)
	body, missB := substitute(tmpl.Body, req.Msg.Variables)
	if tmpl.Channel == "fcm" {
		title = truncateRunes(title, FCMTitleMaxRunes)
		body = truncateRunes(body, FCMBodyMaxRunes)
	}
	return connect.NewResponse(&v1.PreviewTemplateResponse{
		Title: title, Content: body, MissingVariables: unionSorted(missT, missB),
	}), nil
}
```

- [ ] **Step 7: 產生 proto + 跑測試確認通過**

Run: `cd backend && buf generate && go generate ./ent/ && go get connectrpc.com/connect && go test ./internal/domain/notifications/ -v`
Expected: PASS — render 5 測試 + service 5 測試全綠。

- [ ] **Step 8: Commit**

```bash
git add backend/internal/domain/notifications backend/proto/v1/notification.proto backend/gen
git commit -m "feat(backend): 範本渲染語系退回鏈與通知中心 List/MarkRead/UnreadCount API(4.3.2-4.3.3)"
```

---

### Task 3: DeviceService 註冊/註銷 + 失效 token 清理 + promo_tags 資料層(細部 4.3.4–4.3.5)

**Files:**
- Create: `backend/internal/domain/notifications/device_service.go`
- Create: `backend/proto/v1/device.proto`
- Create: `backend/ent/schema/promotag.go`
- Create: `backend/database/migrations/00014_promo_tags_rls.sql`
- Test: `backend/internal/domain/notifications/device_service_test.go`
- Test: `backend/internal/domain/notifications/promotag_test.go`

**Interfaces:**
- Consumes: Task 1 `user_devices`;`audit.Recorder` 介面(**由 01-auth-plan Task 14 提供**;DB 實作由 03-metadicts-audit-plan Task 2.6 接管,本 Task 以 `audit.NoopRecorder` 與測試 fake 注入);`rls.Identity`。
- Produces:
  - Connect-RPC `DeviceService`:`Register(RegisterDeviceRequest{platform, fcm_token, device_name}) returns (RegisterDeviceResponse{device_id})`;`Unregister(UnregisterDeviceRequest{fcm_token}) returns (UnregisterDeviceResponse{})`。
  - `notifications.PurgeInvalidTokens(ctx context.Context, db *ent.Client, tokens []string, rec audit.Recorder) error` — Task 4 FCM 發送迴路呼叫。
  - Ent 實體 `ent.PromoTag`(謂詞套件 `ent/promotag`):`company_id`、`department_id`、`code`、`name`、`is_active`(default true)、`created_at`/`updated_at`/`deleted_at`;部分唯一索引 `(company_id, department_id, code) WHERE deleted_at IS NULL`。
  - 欄位規格(4.3.5,**實際落地於 04-master-data-plan 的 schema 檔**):`products.promo_tag_ids` / `customer_products.promo_tag_ids` / `customers.promo_tag_ids` 皆為 `field.JSON("promo_tag_ids", []uuid.UUID{}).Optional()`(預設空陣列、無外鍵;有效性檢核責任在寫入方 Phase 7 Task 7.4;`promo_tags` 軟刪除時**不反向清理**宿主欄位)。本計畫僅交付 `promo_tags` 表本身;三處宿主欄位由 04-master-data-plan 對應 Task 加入(相依: `backend-detail/04-master-data.md` 商品/專屬商品/客戶 schema 子功能),Phase 7 Task 7.4 直接複用 `backend/internal/domain/promotions/` 目錄續建 CRUD 與選群推播(**跨界:CRUD RPC、客戶訂閱 API、依分類選群推播皆屬 Phase 7 Task 7.4,本計畫不出現對應 RPC**)。

- [ ] **Step 1: 寫失敗測試(裝置生命週期與 purge)**

`backend/internal/domain/notifications/device_service_test.go`:

```go
package notifications_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	v1 "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	entuserdevice "github.com/salesorder/sales-order-1.0/backend/ent/userdevice"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/notifications"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

// fakeRecorder 記錄稽核呼叫(斷言 purge 寫稽核用)。
type fakeRecorder struct{ entries []audit.Entry }

func (f *fakeRecorder) Record(_ context.Context, e audit.Entry) error {
	f.entries = append(f.entries, e)
	return nil
}

func newDeviceService(t *testing.T, rec audit.Recorder) (*notifications.DeviceServiceHandler, *ent.Client) {
	t.Helper()
	c := testutil.NewEntClient(t)
	return notifications.NewDeviceServiceHandler(c, rec), c
}

func TestRegisterIdempotentAndTransfer(t *testing.T) {
	rec := &fakeRecorder{}
	svc, c := newDeviceService(t, rec)
	ctx := context.Background()
	coID, _ := setupCo(t, ctx, c)
	userA, userB := uuid.New(), uuid.New()

	// A 註冊
	resp, err := svc.Register(identityCtx(ctx, userA, coID, nil, false),
		connect.NewRequest(&v1.RegisterDeviceRequest{
			Platform: "ios", FcmToken: "tok-1", DeviceName: "iPhone A",
		}))
	if err != nil {
		t.Fatalf("register A: %v", err)
	}
	// A 重複註冊:冪等,同 device_id
	resp2, err := svc.Register(identityCtx(ctx, userA, coID, nil, false),
		connect.NewRequest(&v1.RegisterDeviceRequest{
			Platform: "ios", FcmToken: "tok-1", DeviceName: "iPhone A",
		}))
	if err != nil || resp2.Msg.DeviceId != resp.Msg.DeviceId {
		t.Fatalf("idempotent register failed: %v id=%q want %q", err, resp2.Msg.DeviceId, resp.Msg.DeviceId)
	}
	// B 用同一 token 註冊(換帳登入):歸屬轉移,新 device_id
	resp3, err := svc.Register(identityCtx(ctx, userB, coID, nil, false),
		connect.NewRequest(&v1.RegisterDeviceRequest{
			Platform: "android", FcmToken: "tok-1", DeviceName: "Pixel B",
		}))
	if err != nil {
		t.Fatalf("register B: %v", err)
	}
	if resp3.Msg.DeviceId == resp.Msg.DeviceId {
		t.Fatal("transfer should create a new device row")
	}
	// 舊記錄已軟刪;A 名下無有效裝置;B 名下有一筆
	oldID := uuid.MustParse(resp.Msg.DeviceId)
	old, err := c.UserDevice.Get(ctx, oldID)
	if err != nil || old.DeletedAt == nil {
		t.Fatalf("old device should be soft-deleted: %v %+v", err, old)
	}
	n, err := c.UserDevice.Query().Where(
		entuserdevice.UserIDEQ(userA), entuserdevice.DeletedAtIsNil()).Count(ctx)
	if err != nil || n != 0 {
		t.Fatalf("A should have no active device, got %d err=%v", n, err)
	}
	n, err = c.UserDevice.Query().Where(
		entuserdevice.UserIDEQ(userB), entuserdevice.DeletedAtIsNil()).Count(ctx)
	if err != nil || n != 1 {
		t.Fatalf("B should have one active device, got %d err=%v", n, err)
	}
	// 註冊有稽核
	if len(rec.entries) == 0 {
		t.Fatal("register should record audit entries")
	}
}

func TestUnregisterIdempotent(t *testing.T) {
	svc, c := newDeviceService(t, &fakeRecorder{})
	ctx := context.Background()
	coID, _ := setupCo(t, ctx, c)
	me := uuid.New()

	resp, err := svc.Register(identityCtx(ctx, me, coID, nil, false),
		connect.NewRequest(&v1.RegisterDeviceRequest{
			Platform: "ios", FcmToken: "tok-9", DeviceName: "iPhone",
		}))
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	// 註銷 → 軟刪除
	if _, err := svc.Unregister(identityCtx(ctx, me, coID, nil, false),
		connect.NewRequest(&v1.UnregisterDeviceRequest{FcmToken: "tok-9"})); err != nil {
		t.Fatalf("unregister: %v", err)
	}
	d, _ := c.UserDevice.Get(ctx, uuid.MustParse(resp.Msg.DeviceId))
	if d.DeletedAt == nil {
		t.Fatal("device should be soft-deleted after unregister")
	}
	// 重複註銷冪等成功
	if _, err := svc.Unregister(identityCtx(ctx, me, coID, nil, false),
		connect.NewRequest(&v1.UnregisterDeviceRequest{FcmToken: "tok-9"})); err != nil {
		t.Fatalf("repeated unregister should succeed, got %v", err)
	}
	// 他人 token → not_found
	if _, err := svc.Unregister(identityCtx(ctx, me, coID, nil, false),
		connect.NewRequest(&v1.UnregisterDeviceRequest{FcmToken: "not-mine"})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("others' token should be not_found, got %v", err)
	}
	// 主帳號註冊 → permission_denied(D22)
	if _, err := svc.Register(identityCtx(ctx, me, coID, nil, true),
		connect.NewRequest(&v1.RegisterDeviceRequest{
			Platform: "ios", FcmToken: "tok-p", DeviceName: "x",
		})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("primary account register should be permission_denied, got %v", err)
	}
	// 空 token → invalid_argument
	if _, err := svc.Register(identityCtx(ctx, me, coID, nil, false),
		connect.NewRequest(&v1.RegisterDeviceRequest{
			Platform: "ios", FcmToken: "", DeviceName: "x",
		})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("empty token should be invalid_argument, got %v", err)
	}
}

func TestPurgeInvalidTokens(t *testing.T) {
	rec := &fakeRecorder{}
	svc, c := newDeviceService(t, rec)
	ctx := context.Background()
	coID, _ := setupCo(t, ctx, c)
	me := uuid.New()
	for _, tok := range []string{"dead-1", "dead-2", "alive-1"} {
		if _, err := svc.Register(identityCtx(ctx, me, coID, nil, false),
			connect.NewRequest(&v1.RegisterDeviceRequest{
				Platform: "ios", FcmToken: tok, DeviceName: "dev " + tok,
			})); err != nil {
			t.Fatalf("register %s: %v", tok, err)
		}
	}
	if err := notifications.PurgeInvalidTokens(ctx, c, []string{"dead-1", "dead-2"}, rec); err != nil {
		t.Fatalf("purge: %v", err)
	}
	n, err := c.UserDevice.Query().Where(entuserdevice.DeletedAtIsNil()).Count(ctx)
	if err != nil || n != 1 {
		t.Fatalf("only alive-1 should remain active, got %d err=%v", n, err)
	}
	// 每筆清除一則稽核(原因 = FCM 回報失效)
	if len(rec.entries) != 2 {
		t.Fatalf("purge should record 2 audit entries, got %d", len(rec.entries))
	}
	// 重複 purge 冪等(已軟刪不再處理、不再記稽核)
	if err := notifications.PurgeInvalidTokens(ctx, c, []string{"dead-1"}, rec); err != nil {
		t.Fatalf("re-purge: %v", err)
	}
	if len(rec.entries) != 2 {
		t.Fatalf("re-purge should not add audit entries, got %d", len(rec.entries))
	}
}
```

- [ ] **Step 2: 寫失敗測試(promo_tags schema 與 RLS)**

`backend/internal/domain/notifications/promotag_test.go`:

```go
package notifications_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestPromoTagPartialUnique(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, deptID := setupCo(t, ctx, c)

	mk := func() error {
		_, err := c.PromoTag.Create().
			SetCompanyID(coID).SetDepartmentID(deptID).
			SetCode("summer").SetName("夏季促銷").Save(ctx)
		return err
	}
	if err := mk(); err != nil {
		t.Fatalf("first tag: %v", err)
	}
	if err := mk(); err == nil {
		t.Fatal("duplicate (company, dept, code) should fail")
	}
	// 同 code 不同部門可共存
	dept2, err := c.Department.Create().SetCompanyID(coID).SetName("業務二部").Save(ctx)
	if err != nil {
		t.Fatalf("dept2: %v", err)
	}
	if _, err := c.PromoTag.Create().
		SetCompanyID(coID).SetDepartmentID(dept2.ID).
		SetCode("summer").SetName("夏季促銷(二部)").Save(ctx); err != nil {
		t.Fatalf("same code in another department should be allowed: %v", err)
	}
}

func TestPromoTagRLSDepartmentIsolation(t *testing.T) {
	c := testutil.NewEntClientWithRLS(t)
	ctx := context.Background()
	superCtx := rls.NewContext(ctx, rls.Identity{UserID: uuid.New(), DataScope: "all"})
	co, err := c.Company.Create().SetName("甲").SetIdentifier("co-a").
		SetCustomerCodePrefix("AA").Save(superCtx)
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	dept1, _ := c.Department.Create().SetCompanyID(co.ID).SetName("一部").Save(superCtx)
	dept2, _ := c.Department.Create().SetCompanyID(co.ID).SetName("二部").Save(superCtx)
	if _, err := c.PromoTag.Create().SetCompanyID(co.ID).SetDepartmentID(dept1.ID).
		SetCode("t1").SetName("一部的標籤").Save(superCtx); err != nil {
		t.Fatalf("seed tag: %v", err)
	}
	// 二部視角(department 級)看不到一部的標籤
	ctxD2 := rls.NewContext(ctx, rls.Identity{
		UserID: uuid.New(), CompanyID: co.ID, DepartmentID: &dept2.ID, DataScope: "department",
	})
	n, err := c.PromoTag.Query().Count(ctxD2)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 0 {
		t.Fatalf("department isolation violated: dept2 sees %d tags of dept1", n)
	}
}
```

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/notifications/ -run 'TestRegister|TestUnregister|TestPurge|TestPromoTag' -v`
Expected: FAIL — `NewDeviceServiceHandler` / `PurgeInvalidTokens` / `ent.PromoTag` 未定義(編譯失敗)。

- [ ] **Step 4: 實作 device_service.go 與 proto**

`backend/internal/domain/notifications/device_service.go`:

```go
package notifications

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entuserdevice "github.com/salesorder/sales-order-1.0/backend/ent/userdevice"
	v1 "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// MaxFCMTokenLen 為 fcm_token 長度上限。
const MaxFCMTokenLen = 512

// DeviceServiceHandler 實作裝置註冊/註銷(細部文件 4.3.4)。
type DeviceServiceHandler struct {
	db  *ent.Client
	rec audit.Recorder // 01-auth-plan Task 14 介面;DB 實作由 03-metadicts-audit-plan 接管
}

func NewDeviceServiceHandler(db *ent.Client, rec audit.Recorder) *DeviceServiceHandler {
	return &DeviceServiceHandler{db: db, rec: rec}
}

// tokenDigest 為稽核用的 token 摘要(敏感欄位全文永不入稽核)。
func tokenDigest(token string) string {
	if len(token) > 12 {
		return token[:12] + "…"
	}
	return token
}

func (h *DeviceServiceHandler) Register(ctx context.Context,
	req *connect.Request[v1.RegisterDeviceRequest]) (*connect.Response[v1.RegisterDeviceResponse], error) {
	id, err := actor(ctx)
	if err != nil {
		return nil, err
	}
	platform := req.Msg.Platform
	token := req.Msg.FcmToken
	name := req.Msg.DeviceName
	if token == "" || len(token) > MaxFCMTokenLen || name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid fcm_token/device_name"))
	}
	if platform != "ios" && platform != "android" && platform != "web" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid platform"))
	}

	tx, err := h.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	now := time.Now()
	// 含軟刪除查詢同一 token(細部文件 4.3.4 步驟 1)
	existing, err := tx.UserDevice.Query().
		Where(entuserdevice.FcmTokenEQ(token)).
		Order(ent.Desc(entuserdevice.FieldCreatedAt)).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var deviceID uuid.UUID
	switch {
	case existing != nil && existing.UserID == id.UserID:
		// 本人已持有:更新平台/名稱/last_seen_at 並復原(清 deleted_at),冪等回傳原 ID
		if _, err := tx.UserDevice.UpdateOneID(existing.ID).
			SetPlatform(entuserdevice.Platform(platform)).
			SetDeviceName(name).
			SetLastSeenAt(now).
			ClearDeletedAt().
			Save(ctx); err != nil {
			_ = tx.Rollback()
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		deviceID = existing.ID
	case existing != nil:
		// 他人持有(換帳登入同一裝置):原記錄軟刪除,為當前使用者新建
		if _, err := tx.UserDevice.UpdateOneID(existing.ID).
			SetDeletedAt(now).Save(ctx); err != nil {
			_ = tx.Rollback()
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		fallthrough
	default:
		d, err := tx.UserDevice.Create().
			SetUserID(id.UserID).SetCompanyID(id.CompanyID).
			SetPlatform(entuserdevice.Platform(platform)).
			SetFcmToken(token).SetDeviceName(name).
			SetLastSeenAt(now).
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("device register conflict: %v", err))
		}
		deviceID = d.ID
	}
	// 裝置註冊稽核與寫入同一 DB 交易(D18);token 僅記摘要
	if err := h.rec.Record(ctx, audit.Entry{
		ActorID: id.UserID.String(), ActorName: id.UserID.String(),
		Action: "device.register", ResourceType: "user_devices",
		ResourceID: deviceID.String(),
		After:      fmt.Sprintf("platform=%s token=%s", platform, tokenDigest(token)),
	}); err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.RegisterDeviceResponse{DeviceId: deviceID.String()}), nil
}

func (h *DeviceServiceHandler) Unregister(ctx context.Context,
	req *connect.Request[v1.UnregisterDeviceRequest]) (*connect.Response[v1.UnregisterDeviceResponse], error) {
	id, err := actor(ctx)
	if err != nil {
		return nil, err
	}
	token := req.Msg.FcmToken
	if token == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("fcm_token required"))
	}
	n, err := h.db.UserDevice.Update().Where(
		entuserdevice.FcmTokenEQ(token),
		entuserdevice.UserIDEQ(id.UserID),
		entuserdevice.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if n == 0 {
		// 查無記錄(從未註冊)→ 冪等成功;但 token 屬他人 → not_found(不允許代註銷)
		owned, err := h.db.UserDevice.Query().Where(
			entuserdevice.FcmTokenEQ(token), entuserdevice.DeletedAtIsNil()).Exist(ctx)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		if owned {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("device not found"))
		}
	}
	return connect.NewResponse(&v1.UnregisterDeviceResponse{}), nil
}

// PurgeInvalidTokens 將 FCM 回報失效的 token 對應 user_devices 全部軟刪除
// (系統行為,不限當前使用者;RLS 以對應公司範圍由呼叫方注入,細部文件 4.3.4 步驟 4)。
// 每筆清除寫稽核(原因 = FCM 回報失效)。冪等:已軟刪的 token 略過。
func PurgeInvalidTokens(ctx context.Context, db *ent.Client, tokens []string, rec audit.Recorder) error {
	if len(tokens) == 0 {
		return nil
	}
	tx, err := db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("notifications: purge tx: %w", err)
	}
	devices, err := tx.UserDevice.Query().Where(
		entuserdevice.FcmTokenIn(tokens...),
		entuserdevice.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("notifications: purge query: %w", err)
	}
	now := time.Now()
	for _, d := range devices {
		if _, err := tx.UserDevice.UpdateOneID(d.ID).SetDeletedAt(now).Save(ctx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("notifications: purge device %s: %w", d.ID, err)
		}
		if err := rec.Record(ctx, audit.Entry{
			ActorID: "system", ActorName: "fcm-purge",
			Action: "device.purge_invalid", ResourceType: "user_devices",
			ResourceID: d.ID.String(),
			Before:     fmt.Sprintf("token=%s reason=fcm-unregistered", tokenDigest(d.FcmToken)),
		}); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("notifications: purge audit: %w", err)
		}
		log.Printf("notifications: purged invalid device %s (token %s)", d.ID, tokenDigest(d.FcmToken))
	}
	return tx.Commit()
}
```

`backend/proto/v1/device.proto`:

```proto
syntax = "proto3";
package salesorder.v1;

service DeviceService {
  rpc Register(RegisterDeviceRequest) returns (RegisterDeviceResponse);
  rpc Unregister(UnregisterDeviceRequest) returns (UnregisterDeviceResponse);
}

message RegisterDeviceRequest {
  string platform = 1;    // ios | android | web
  string fcm_token = 2;
  string device_name = 3;
}
message RegisterDeviceResponse { string device_id = 1; }

message UnregisterDeviceRequest { string fcm_token = 1; }
message UnregisterDeviceResponse {}
```

- [ ] **Step 5: 實作 promotag.go schema 與 RLS migration**

`backend/ent/schema/promotag.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PromoTag 為促銷分類標籤,部門級(細部文件 4.3.5,D24)。
// 僅資料層:CRUD RPC、客戶訂閱 API、依分類選群推播屬 Phase 7 Task 7.4,本 Phase 不建。
// 軟刪除時不反向清理宿主(products / customer_products / customers)的 promo_tag_ids;
// 殘留 ID 在推播選群時自然失效(細部文件 4.3.5 步驟 3)。
type PromoTag struct{ ent.Schema }

func (PromoTag) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("company_id", uuid.UUID{}),
		field.UUID("department_id", uuid.UUID{}),
		field.String("code").NotEmpty(),
		field.String("name").NotEmpty(),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (PromoTag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("company_id", "department_id", "code").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}
```

宿主欄位規格(落地於 04-master-data-plan 的 schema 檔,本計畫不修改那些檔案;執行時若 04 尚未落地,於其 schema 建立後補入,三處形狀一致):

```go
// 加至 backend/ent/schema/product.go、customerproduct.go、customer.go(04-master-data-plan 檔案)的 Fields():
field.JSON("promo_tag_ids", []uuid.UUID{}).Optional(),
// TODO(相依: backend-detail/04-master-data.md 商品/專屬商品/客戶 schema 子功能;
//  有效性檢核在寫入方 Phase 7 Task 7.4,JSON 陣列無法外鍵)
```

`backend/database/migrations/00014_promo_tags_rls.sql`:

```sql
-- +goose Up
-- 細部文件 4.3.5 步驟 1:promo_tags 部門級 RLS。

ALTER TABLE promo_tags ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON promo_tags
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true));

CREATE POLICY department_scope ON promo_tags
  USING (current_setting('app.current_data_scope', true) IN ('all', 'company')
         OR department_id::text = current_setting('app.current_department_id', true));

-- +goose Down
DROP POLICY IF EXISTS department_scope ON promo_tags;
DROP POLICY IF EXISTS tenant_isolation ON promo_tags;
ALTER TABLE promo_tags DISABLE ROW LEVEL SECURITY;
```

- [ ] **Step 6: 產生碼 + 跑測試確認通過**

Run: `cd backend && buf generate && go generate ./ent/ && go test ./internal/domain/notifications/ -v`
Expected: PASS — 裝置 3 測試 + promo_tags 2 測試全綠(Task 1/2 既有測試不回歸)。

- [ ] **Step 7: Commit**

```bash
git add backend/internal/domain/notifications/device_service.go backend/proto/v1/device.proto backend/ent/schema/promotag.go backend/ent backend/database/migrations/00014_promo_tags_rls.sql backend/gen
git commit -m "feat(backend): DeviceService 註冊註銷、失效 token purge、promo_tags 資料層(4.3.4-4.3.5)"
```

---

### Task 4: FCM client + 站內發送 + 失敗標記不重試(細部 4.4.1、4.4.2、4.4.5)

**Files:**
- Create: `backend/internal/notification/fcm.go`
- Create: `backend/internal/notification/inapp.go`
- Create: `backend/internal/notification/failmark.go`
- Test: `backend/internal/notification/dispatch_test.go`
- Test: `backend/internal/notification/failmark_test.go`

**Interfaces:**
- Consumes: Task 1 `notifications` / `user_devices`;Task 2 `notifications.Render` / `notifications.Repo`;Task 3 `notifications.PurgeInvalidTokens`;`audit.Recorder`(01-auth-plan Task 14)。
- Produces(套件 `notification`,singular;供 Task 5 與未來觸發方使用):
  - `notification.Message{Title, Body string; Data map[string]string}`
  - `notification.SendResult{Token string; OK bool; Invalid bool; ErrCategory string; ErrDetail string}` — `Invalid` = FCM 回報 unregistered / invalid-registration-token。
  - `notification.Sender` 介面:`Send(ctx context.Context, tokens []string, msg Message) ([]SendResult, error)`。
  - `notification.NewFCMSender(credentialsPath string, disabled bool, timeout time.Duration) (Sender, error)` — 缺憑證且非 disabled → 啟動期 fail-fast(error);disabled → 回傳僅記日誌的降級 sender(僅供開發,文件註明)。
  - `notification.FakeSender{Results []SendResult; Err error}`(測試用,實作 `Sender`)。
  - `notification.Dispatch(ctx context.Context, db *ent.Client, sender Sender, rec audit.Recorder, ids []uuid.UUID) error` — 交易提交後呼叫,逐筆處理 pending 通知:in_app → 標 sent;fcm → 查裝置發送,任一裝置成功即 sent、失效 token purge、全部失敗 → `MarkFailed`。
  - `notification.SendInApp(ctx context.Context, tx *ent.Tx, q notifications.TemplateQuery, userIDs []uuid.UUID, vars map[string]string, payload map[string]any) ([]uuid.UUID, error)` — 站內通道建檔(交易內,pending;送達由 `Dispatch` 標記)。
  - `notification.MarkFailed(ctx context.Context, db *ent.Client, id uuid.UUID, reason FailReason) error` — 僅在 `status = pending` 時生效(條件更新),寫警告日誌 + prometheus 計數。
  - `notification.FailReason` 型別(string 列舉):`FailNoDevice = "no_device"`、`FailTimeout = "fcm_timeout"`、`FailQuota = "fcm_quota"`、`FailUnregistered = "fcm_unregistered"`、`FailTemplateMissing = "template_missing"`、`FailOther = "fcm_other"`;`failure_reason` 欄位寫入 `"<category>: <detail 摘要(截 200 rune)>"`。

- [ ] **Step 1: 寫失敗測試(Dispatch 雙通道、失效 token、多裝置)**

`backend/internal/notification/dispatch_test.go`:

```go
package notification_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entnotification "github.com/salesorder/sales-order-1.0/backend/ent/notification"
	entuserdevice "github.com/salesorder/sales-order-1.0/backend/ent/userdevice"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/notifications"
	"github.com/salesorder/sales-order-1.0/backend/internal/notification"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func seedUser(t *testing.T, ctx context.Context, c *ent.Client, coID uuid.UUID) uuid.UUID {
	t.Helper()
	u, err := c.User.Create().SetCompanyID(coID).SetRole("customer").
		SetDataScope("self").SetStatus("active").SetIsCustomer(true).
		SetCustomerID(uuid.New()).SetAccountName("子帳號").SetIsPrimary(false).
		SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u.ID
}

func seedDevice(t *testing.T, ctx context.Context, c *ent.Client, userID, coID uuid.UUID, token string) {
	t.Helper()
	if _, err := c.UserDevice.Create().SetUserID(userID).SetCompanyID(coID).
		SetPlatform("ios").SetFcmToken(token).SetDeviceName("dev").
		SetLastSeenAt(time.Now()).Save(ctx); err != nil {
		t.Fatalf("seed device: %v", err)
	}
}

func createPending(t *testing.T, ctx context.Context, c *ent.Client,
	userID, coID uuid.UUID, channel string) uuid.UUID {
	t.Helper()
	tx, err := c.Tx(ctx)
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
	id, err := (notifications.Repo{}).CreateOne(ctx, tx, notifications.CreateParams{
		CompanyID: coID, UserID: userID, Channel: channel,
		Title: "t", Content: "c", Payload: map[string]any{},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	return id
}

func TestDispatchInAppMarksSent(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, _ := setupCo(t, ctx, c)
	uid := seedUser(t, ctx, c, coID)
	id := createPending(t, ctx, c, uid, coID, "in_app")

	if err := notification.Dispatch(ctx, c, &notification.FakeSender{}, audit.NoopRecorder{}, []uuid.UUID{id}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	n, _ := c.Notification.Get(ctx, id)
	if n.Status != "sent" || n.SentAt == nil {
		t.Fatalf("in_app should be sent immediately: status=%s", n.Status)
	}
}

func TestDispatchFCMResultClassification(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, _ := setupCo(t, ctx, c)
	uid := seedUser(t, ctx, c, coID)
	seedDevice(t, ctx, c, uid, coID, "good-tok")
	seedDevice(t, ctx, c, uid, coID, "dead-tok")
	id := createPending(t, ctx, c, uid, coID, "fcm")

	sender := &notification.FakeSender{Results: []notification.SendResult{
		{Token: "good-tok", OK: true},
		{Token: "dead-tok", OK: false, Invalid: true},
	}}
	if err := notification.Dispatch(ctx, c, sender, audit.NoopRecorder{}, []uuid.UUID{id}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	// 任一裝置成功即 sent(細部文件 4.4.1 步驟 3)
	n, _ := c.Notification.Get(ctx, id)
	if n.Status != "sent" {
		t.Fatalf("one device success should mark sent, got %s", n.Status)
	}
	// 失效 token 已被 purge 軟刪除
	dead, err := c.UserDevice.Query().Where(
		entuserdevice.FcmTokenEQ("dead-tok")).Only(ctx)
	if err != nil || dead.DeletedAt == nil {
		t.Fatalf("dead-tok should be soft-deleted: %v %+v", err, dead)
	}
	// 成功裝置 last_seen_at 更新
	good, _ := c.UserDevice.Query().Where(entuserdevice.FcmTokenEQ("good-tok")).Only(ctx)
	if time.Since(good.LastSeenAt) > time.Minute {
		t.Fatal("successful device last_seen_at should be refreshed")
	}
}

func TestDispatchFCMAllDevicesFail(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, _ := setupCo(t, ctx, c)
	uid := seedUser(t, ctx, c, coID)
	seedDevice(t, ctx, c, uid, coID, "tok-a")
	id := createPending(t, ctx, c, uid, coID, "fcm")

	sender := &notification.FakeSender{Results: []notification.SendResult{
		{Token: "tok-a", OK: false, ErrCategory: string(notification.FailTimeout), ErrDetail: "deadline exceeded"},
	}}
	if err := notification.Dispatch(ctx, c, sender, audit.NoopRecorder{}, []uuid.UUID{id}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	n, _ := c.Notification.Get(ctx, id)
	if n.Status != "failed" || n.FailureReason == nil {
		t.Fatalf("all devices failed should mark failed, got %s reason=%v", n.Status, n.FailureReason)
	}
	if got := *n.FailureReason; got == "" || got[:11] != "fcm_timeout" {
		t.Fatalf("failure_reason should start with category, got %q", got)
	}
	if n.SentAt != nil {
		t.Fatal("failed notification must keep sent_at NULL")
	}
}

func TestDispatchFCMNoDevice(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, _ := setupCo(t, ctx, c)
	uid := seedUser(t, ctx, c, coID) // 無裝置
	id := createPending(t, ctx, c, uid, coID, "fcm")

	sender := &notification.FakeSender{}
	if err := notification.Dispatch(ctx, c, sender, audit.NoopRecorder{}, []uuid.UUID{id}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	n, _ := c.Notification.Get(ctx, id)
	if n.Status != "failed" || n.FailureReason == nil || (*n.FailureReason)[:9] != "no_device" {
		t.Fatalf("no device should mark failed(no_device), got %s %v", n.Status, n.FailureReason)
	}
	if sender.Calls != 0 {
		t.Fatal("sender must not be called when user has no device")
	}
}

func TestDispatchSenderErrorAndPartialFailure(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, _ := setupCo(t, ctx, c)
	u1 := seedUser(t, ctx, c, coID)
	u2 := seedUser(t, ctx, c, coID)
	seedDevice(t, ctx, c, u1, coID, "tok-1")
	seedDevice(t, ctx, c, u2, coID, "tok-2")
	id1 := createPending(t, ctx, c, u1, coID, "fcm")
	id2 := createPending(t, ctx, c, u2, coID, "fcm")

	// 第一人 sender 回傳整批錯誤;第二人成功 → 部分失敗互不影響(細部文件 4.4.5 步驟 3)
	sender := &notification.FakeSender{
		ErrByToken: map[string]error{"tok-1": errors.New("connection reset")},
		Results:    []notification.SendResult{{Token: "tok-2", OK: true}},
	}
	if err := notification.Dispatch(ctx, c, sender, audit.NoopRecorder{}, []uuid.UUID{id1, id2}); err != nil {
		t.Fatalf("dispatch should not return error for partial failure, got %v", err)
	}
	n1, _ := c.Notification.Get(ctx, id1)
	n2, _ := c.Notification.Get(ctx, id2)
	if n1.Status != "failed" {
		t.Fatalf("u1 should be failed, got %s", n1.Status)
	}
	if n2.Status != "sent" {
		t.Fatalf("u2 should be sent, got %s", n2.Status)
	}
}
```

(`FakeSender.ErrByToken` 為 fake 的測試控制欄位,見 Step 3 實作;`sender.Calls` 記錄呼叫次數。)

- [ ] **Step 2: 寫失敗測試(MarkFailed 條件更新、不重試、指標)**

`backend/internal/notification/failmark_test.go`:

```go
package notification_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	promtestutil "github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/salesorder/sales-order-1.0/backend/internal/notification"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestMarkFailedOnlyFromPending(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, _ := setupCo(t, ctx, c)
	uid := seedUser(t, ctx, c, coID)
	id := createPending(t, ctx, c, uid, coID, "fcm")

	before := promtestutil.ToFloat64(notification.FailuresTotal.WithLabelValues("fcm", string(notification.FailQuota)))
	if err := notification.MarkFailed(ctx, c, id, notification.FailQuota, "quota exceeded"); err != nil {
		t.Fatalf("mark failed: %v", err)
	}
	n, _ := c.Notification.Get(ctx, id)
	if n.Status != "failed" || n.FailureReason == nil || !strings.HasPrefix(*n.FailureReason, "fcm_quota") {
		t.Fatalf("got status=%s reason=%v", n.Status, n.FailureReason)
	}
	after := promtestutil.ToFloat64(notification.FailuresTotal.WithLabelValues("fcm", string(notification.FailQuota)))
	if after != before+1 {
		t.Fatalf("failures_total should increment: %v -> %v", before, after)
	}

	// 終態不可再轉:對 failed 再 MarkFailed 為 no-op,不覆蓋 reason
	if err := notification.MarkFailed(ctx, c, id, notification.FailOther, "x"); err != nil {
		t.Fatalf("re-mark should be no-op success, got %v", err)
	}
	n, _ = c.Notification.Get(ctx, id)
	if !strings.HasPrefix(*n.FailureReason, "fcm_quota") {
		t.Fatalf("terminal state must not be overwritten, got %q", *n.FailureReason)
	}
	// 1.0 無重試佇列:failed 後狀態停留,無任何自動重發(此行為由無 retry 程式路徑保證,
	// 任何新增 retry 的 PR 應被 review 拒絕 — D16)
}

func TestMarkFailedDoesNotOverwriteSent(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coID, _ := setupCo(t, ctx, c)
	uid := seedUser(t, ctx, c, coID)
	id := createPending(t, ctx, c, uid, coID, "in_app")

	if err := notification.Dispatch(ctx, c, &notification.FakeSender{}, auditNoop(), []uuid.UUID{id}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	// 雙通道競態:已 sent 的記錄再被 MarkFailed → 略過(細部文件 4.4.5 步驟 5)
	if err := notification.MarkFailed(ctx, c, id, notification.FailOther, "late failure"); err != nil {
		t.Fatalf("mark failed on sent: %v", err)
	}
	n, _ := c.Notification.Get(ctx, id)
	if n.Status != "sent" {
		t.Fatalf("sent must not be overwritten by failed, got %s", n.Status)
	}
}
```

(`auditNoop()` 為測試輔助,回傳 `audit.NoopRecorder{}`。)

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/notification/ -v`
Expected: FAIL — `notification.Dispatch` / `MarkFailed` / `FakeSender` 未定義(編譯失敗)。

- [ ] **Step 4: 實作 fcm.go(Sender 介面、FCM 實作、FakeSender、Dispatch)**

`backend/internal/notification/fcm.go`:

```go
// Package notification 為通知發送引擎:FCM / 站內通道、失敗標記(細部文件 4.4)。
package notification

import (
	"context"
	"fmt"
	"log"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"google.golang.org/api/option"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entnotification "github.com/salesorder/sales-order-1.0/backend/ent/notification"
	entuserdevice "github.com/salesorder/sales-order-1.0/backend/ent/userdevice"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/notifications"
)

// Message 為跨通道訊息;Data 攜帶導頁資訊(如 order_id)。
type Message struct {
	Title string
	Body  string
	Data  map[string]string
}

// SendResult 為單一 token 的發送結果。Invalid = FCM 回報 unregistered /
// invalid-registration-token(對應 user_devices 應清除,細部文件 4.4.1 步驟 3)。
type SendResult struct {
	Token       string
	OK          bool
	Invalid     bool
	ErrCategory string // FailReason 列舉值;OK/Invalid 時為空
	ErrDetail   string
}

// Sender 抽象 FCM 發送;測試注入 FakeSender,不觸網。
type Sender interface {
	Send(ctx context.Context, tokens []string, msg Message) ([]SendResult, error)
}

// fcmSender 為 FCM Admin SDK 實作;multicast 一次最多 500 token。
type fcmSender struct {
	client  *messaging.Client
	timeout time.Duration
}

// NewFCMSender 以 service account 憑證建立 Sender。
// 缺憑證且 disabled=false → 回傳 error(啟動期 fail-fast,細部文件 4.4.1 介面);
// disabled=true(FCM_DISABLED) → 回傳僅記日誌的降級 sender,僅供開發環境使用。
func NewFCMSender(credentialsPath string, disabled bool, timeout time.Duration) (Sender, error) {
	if disabled {
		log.Print("notification: FCM_DISABLED=true, sender degraded to log-only (development only)")
		return logSender{}, nil
	}
	if credentialsPath == "" {
		return nil, fmt.Errorf("notification: FCM credentials path required (or set FCM_DISABLED=true in development)")
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, fmt.Errorf("notification: firebase app: %w", err)
	}
	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("notification: firebase messaging: %w", err)
	}
	return &fcmSender{client: client, timeout: timeout}, nil
}

const fcmMulticastLimit = 500

func (s *fcmSender) Send(ctx context.Context, tokens []string, msg Message) ([]SendResult, error) {
	results := make([]SendResult, 0, len(tokens))
	for start := 0; start < len(tokens); start += fcmMulticastLimit {
		end := start + fcmMulticastLimit
		if end > len(tokens) {
			end = len(tokens)
		}
		batch := tokens[start:end]
		callCtx, cancel := context.WithTimeout(ctx, s.timeout)
		resp, err := s.client.SendEachForMulticast(callCtx, &messaging.MulticastMessage{
			Tokens:       batch,
			Notification: &messaging.Notification{Title: msg.Title, Body: msg.Body},
			Data:         msg.Data,
		})
		cancel()
		if err != nil {
			// 整批呼叫失敗(連線/逾時):逐 token 標同類錯誤
			cat := classifyFCMError(err)
			for _, tok := range batch {
				results = append(results, SendResult{Token: tok, ErrCategory: cat, ErrDetail: err.Error()})
			}
			continue
		}
		for i, r := range resp.Responses {
			if r.Success {
				results = append(results, SendResult{Token: batch[i], OK: true})
				continue
			}
			switch {
			case messaging.IsUnregistered(r.Error) || messaging.IsInvalidArgument(r.Error):
				results = append(results, SendResult{Token: batch[i], Invalid: true})
			default:
				results = append(results, SendResult{
					Token: batch[i], ErrCategory: classifyFCMError(r.Error), ErrDetail: r.Error.Error(),
				})
			}
		}
	}
	return results, nil
}

// classifyFCMError 將 FCM 錯誤分類為 FailReason 列舉(逾時 / 配額 / 其他)。
func classifyFCMError(err error) string {
	switch {
	case messaging.IsQuotaExceeded(err):
		return string(FailQuota)
	case errors.Is(err, context.DeadlineExceeded):
		return string(FailTimeout)
	default:
		return string(FailOther)
	}
}

// logSender 為 FCM_DISABLED 開發降級:僅記日誌,一律回報成功。
type logSender struct{}

func (logSender) Send(_ context.Context, tokens []string, msg Message) ([]SendResult, error) {
	log.Printf("notification: [FCM_DISABLED] would send to %d tokens: %q", len(tokens), msg.Title)
	out := make([]SendResult, len(tokens))
	for i, t := range tokens {
		out[i] = SendResult{Token: t, OK: true}
	}
	return out, nil
}

// FakeSender 為測試用 Sender:Results 為依序回傳結果;Err 為整批錯誤;
// ErrByToken 可依 token 指定整批錯誤;Calls 記錄呼叫次數。
type FakeSender struct {
	Results    []SendResult
	Err        error
	ErrByToken map[string]error
	Calls      int
	LastTokens []string
}

func (f *FakeSender) Send(_ context.Context, tokens []string, _ Message) ([]SendResult, error) {
	f.Calls++
	f.LastTokens = tokens
	if f.Err != nil {
		return nil, f.Err
	}
	byToken := map[string]SendResult{}
	for _, r := range f.Results {
		byToken[r.Token] = r
	}
	out := make([]SendResult, 0, len(tokens))
	for _, t := range tokens {
		if err, ok := f.ErrByToken[t]; ok {
			return nil, err
		}
		if r, ok := byToken[t]; ok {
			out = append(out, r)
		} else {
			out = append(out, SendResult{Token: t, OK: true})
		}
	}
	return out, nil
}

// Dispatch 於觸發方交易提交後呼叫(時序邊界:FCM 外部 I/O 不得在 DB 交易內,
// 細部文件 4.4.1 步驟 2)。逐筆處理 pending 通知,部分失敗不影響他人;
// 發送失敗只標 failed,不回滾業務、不重試(D16)。
func Dispatch(ctx context.Context, db *ent.Client, sender Sender, rec audit.Recorder, ids []uuid.UUID) error {
	for _, id := range ids {
		if err := dispatchOne(ctx, db, sender, rec, id); err != nil {
			// 單筆內部錯誤僅記日誌,不阻斷其他通知(狀態停留 pending 由人工依日誌追查)
			log.Printf("notification: dispatch %s: %v", id, err)
		}
	}
	return nil
}

func dispatchOne(ctx context.Context, db *ent.Client, sender Sender, rec audit.Recorder, id uuid.UUID) error {
	n, err := db.Notification.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}
	if n.Status != entnotification.StatusPending {
		return nil // 已被其他路徑處理(read/failed 終態)
	}
	switch n.Channel {
	case entnotification.ChannelInApp:
		// 站內通道無外部投遞環節:建檔成功即視為送達(細部文件 4.4.2 步驟 2,採
		// 「提交後立即標 sent」策略,與 FCM 通道保持 pending → sent 狀態機一致)
		return markSent(ctx, db, id)
	case entnotification.ChannelFcm:
		return dispatchFCM(ctx, db, sender, rec, n)
	default:
		return fmt.Errorf("unknown channel %q", n.Channel)
	}
}

func dispatchFCM(ctx context.Context, db *ent.Client, sender Sender, rec audit.Recorder, n *ent.Notification) error {
	devices, err := db.UserDevice.Query().Where(
		entuserdevice.UserIDEQ(n.UserID),
		entuserdevice.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return fmt.Errorf("query devices: %w", err)
	}
	if len(devices) == 0 {
		// 接收者無有效裝置:直接 failed(no_device),站內通道不受影響(細部文件 4.4.2 步驟 4)
		return MarkFailed(ctx, db, n.ID, FailNoDevice, "user has no active device")
	}
	tokens := make([]string, len(devices))
	data := map[string]string{}
	for k, v := range n.Payload {
		if s, ok := v.(string); ok {
			data[k] = s
		}
	}
	for i, d := range devices {
		tokens[i] = d.FcmToken
	}
	results, err := sender.Send(ctx, tokens, Message{Title: n.Title, Body: n.Content, Data: data})
	if err != nil {
		// 整批呼叫失敗:視同全部失敗,分類由錯誤決定
		return MarkFailed(ctx, db, n.ID, FailReason(classifyFCMError(err)), err.Error())
	}
	var invalid []string
	anyOK := false
	var lastCat, lastDetail string
	now := time.Now()
	for i, r := range results {
		switch {
		case r.OK:
			anyOK = true
			// last_seen_at 僅在 Register 與成功發送時更新(細部文件 4.3.4 步驟 5)
			if _, err := db.UserDevice.UpdateOneID(devices[i].ID).SetLastSeenAt(now).Save(ctx); err != nil {
				log.Printf("notification: update last_seen_at device %s: %v", devices[i].ID, err)
			}
		case r.Invalid:
			invalid = append(invalid, r.Token)
		default:
			lastCat, lastDetail = r.ErrCategory, r.ErrDetail
		}
	}
	if len(invalid) > 0 {
		if err := notifications.PurgeInvalidTokens(ctx, db, invalid, rec); err != nil {
			log.Printf("notification: purge invalid tokens: %v", err)
		}
	}
	if anyOK {
		// 任一裝置成功即 sent(細部文件 4.4.1 步驟 3-4)
		return markSent(ctx, db, n.ID)
	}
	if lastCat == "" {
		// 全部 token 皆失效被清除
		return MarkFailed(ctx, db, n.ID, FailUnregistered, "all device tokens unregistered")
	}
	return MarkFailed(ctx, db, n.ID, FailReason(lastCat), lastDetail)
}

// markSent 條件更新 pending → sent(競態下不覆蓋 read/failed 終態)。
func markSent(ctx context.Context, db *ent.Client, id uuid.UUID) error {
	_, err := db.Notification.Update().Where(
		entnotification.IDEQ(id),
		entnotification.StatusEQ(entnotification.StatusPending),
	).SetStatus(entnotification.StatusSent).SetSentAt(time.Now()).Save(ctx)
	if err != nil {
		return fmt.Errorf("mark sent: %w", err)
	}
	return nil
}
```

(`fcm.go` 需補 import:`"errors"`。)

- [ ] **Step 5: 實作 inapp.go 與 failmark.go**

`backend/internal/notification/inapp.go`:

```go
package notification

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/notifications"
)

// SendInApp 於觸發方交易內為每位接收者渲染 in_app 範本並建一筆 pending 通知
// (細部文件 4.4.2 步驟 1);送達標記由提交後的 Dispatch 完成。
// 範本缺失 → 降級建 template_id 為空的通知(細部文件 4.3.2 錯誤處理);
// 接收者已停用/不存在 → 該筆略過並記日誌,不影響其他接收者。
func SendInApp(ctx context.Context, tx *ent.Tx, q notifications.TemplateQuery,
	userIDs []uuid.UUID, vars map[string]string, payload map[string]any) ([]uuid.UUID, error) {
	q.Channel = "in_app"
	ids := make([]uuid.UUID, 0, len(userIDs))
	repo := notifications.Repo{}
	for _, uid := range userIDs {
		ok, err := activeUser(ctx, tx, uid)
		if err != nil {
			return nil, fmt.Errorf("notification: check user %s: %w", uid, err)
		}
		if !ok {
			log.Printf("notification: skip disabled/unknown recipient %s", uid)
			continue
		}
		r, err := notifications.Render(ctx, tx.Client(), q, vars)
		var tmplID *uuid.UUID
		var title, content string
		switch {
		case err == nil:
			tmplID, title, content = &r.TemplateID, r.Title, r.Content
		case errors.Is(err, notifications.ErrTemplateNotFound):
			// 降級:無範本仍以變數摘要建檔,通知不中斷(細部文件 4.3.2 錯誤處理)
			log.Printf("notification: template %s missing, degrade to template-less record", q.Code)
			title, content = q.Code, summarizeVars(vars)
		default:
			return nil, fmt.Errorf("notification: render %s: %w", q.Code, err)
		}
		id, err := repo.CreateOne(ctx, tx, notifications.CreateParams{
			CompanyID: q.CompanyID, DepartmentID: q.DepartmentID, UserID: uid,
			TemplateID: tmplID, Channel: "in_app",
			Title: title, Content: content, Payload: payload,
		})
		if err != nil {
			return nil, err // 建檔失敗 → 隨觸發方交易回滾(上層感知 internal,D18)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// activeUser 確認接收者存在且未停用、未軟刪除。
func activeUser(ctx context.Context, tx *ent.Tx, uid uuid.UUID) (bool, error) {
	return tx.User.Query().Where(
		entuser.IDEQ(uid),
		entuser.StatusEQ("active"),
		entuser.DeletedAtIsNil(),
	).Exist(ctx)
}

// summarizeVars 為降級通知的內容摘要(鍵值平鋪,穩定排序)。
func summarizeVars(vars map[string]string) string {
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+vars[k])
	}
	return strings.Join(parts, " ")
}
```

(`inapp.go` 需補 import:`"sort"`、`"strings"`、`entuser "github.com/salesorder/sales-order-1.0/backend/ent/user"`;`entuser.StatusEQ` 參數照產生碼 enum 轉型。)

`backend/internal/notification/failmark.go`:

```go
package notification

import (
	"context"
	"fmt"
	"log"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entnotification "github.com/salesorder/sales-order-1.0/backend/ent/notification"
)

// FailReason 為失敗原因列舉(細部文件 4.4.5 介面)。
type FailReason string

const (
	FailNoDevice        FailReason = "no_device"
	FailTimeout         FailReason = "fcm_timeout"
	FailQuota           FailReason = "fcm_quota"
	FailUnregistered    FailReason = "fcm_unregistered"
	FailTemplateMissing FailReason = "template_missing"
	FailOther           FailReason = "fcm_other"
)

// FailuresTotal 為通知失敗計數,依 channel / reason 分維度(D19,/metrics 暴露)。
var FailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "salesorder",
	Name:      "notification_failures_total",
	Help:      "通知發送失敗計數(不重試,failed 為終態)",
}, []string{"channel", "reason"})

const failureReasonMaxRunes = 200

// MarkFailed 統一失敗收斂:不重試、status=failed、記錄原因(細部文件 4.4.5)。
// 僅在 status=pending 時生效(條件更新):已被標 sent(雙通道競態)或 failed/read
// (終態)的記錄略過,不覆蓋。自身 SQL 失敗只記錯誤日誌,不再升級。
func MarkFailed(ctx context.Context, db *ent.Client, id uuid.UUID, reason FailReason, detail string) error {
	if utf8.RuneCountInString(detail) > failureReasonMaxRunes {
		detail = string([]rune(detail)[:failureReasonMaxRunes])
	}
	full := string(reason)
	if detail != "" {
		full = fmt.Sprintf("%s: %s", reason, detail)
	}
	updated, err := db.Notification.Update().Where(
		entnotification.IDEQ(id),
		entnotification.StatusEQ(entnotification.StatusPending),
	).
		SetStatus(entnotification.StatusFailed).
		SetFailureReason(full).
		Save(ctx)
	if err != nil {
		// 通知狀態停留 pending,由人工依日誌追查(細部文件 4.4.5 錯誤處理)
		log.Printf("notification: MarkFailed %s sql error: %v", id, err)
		return fmt.Errorf("notification: mark failed: %w", err)
	}
	if updated == 0 {
		return nil // 非 pending(已 sent/failed/read):略過
	}
	log.Printf("notification: %s marked failed (%s)", id, full)
	n, err := db.Notification.Get(ctx, id)
	if err == nil {
		FailuresTotal.WithLabelValues(string(n.Channel), string(reason)).Inc()
	}
	return nil
}
```

- [ ] **Step 6: 跑測試確認通過**

Run: `cd backend && go get firebase.google.com/go/v4 google.golang.org/api github.com/prometheus/client_golang && go test ./internal/notification/ -v`
Expected: PASS — dispatch 5 測試 + failmark 2 測試全綠。

- [ ] **Step 7: Commit**

```bash
git add backend/internal/notification
git commit -m "feat(backend): FCM Sender 抽象、Dispatch 雙通道發送、站內建檔、MarkFailed 不重試(4.4.1/4.4.2/4.4.5)"
```

---

### Task 5: 通知路由 — 下單與專屬商品觸發(細部 4.4.3–4.4.4)

**Files:**
- Create: `backend/internal/notification/triggers/order_created.go`
- Create: `backend/internal/notification/triggers/customer_product_created.go`
- Test: `backend/internal/notification/triggers/triggers_test.go`

**Interfaces:**
- Consumes: Task 2 `notifications.Render` / `notifications.Repo`;Task 4 `notification.Dispatch` / `notification.SendInApp`;`users` 表 `is_customer` / `is_primary` / `status` / `customer_id`(01-auth-plan Task 1);`customers.default_sales_rep_id`(**由 04-master-data-plan 提供**,本 Task 以參數傳入,不直接依賴該表產生碼)。
- Produces:
  - `triggers.OrderInfo{ID uuid.UUID; OrderNo string; CompanyID uuid.UUID; DepartmentID *uuid.UUID; CustomerID uuid.UUID; CustomerName string; ItemCount int}`
  - `triggers.ActorInfo{UserID uuid.UUID; Role string; IsCustomer bool}`
  - `triggers.OnOrderCreated(ctx context.Context, tx *ent.Tx, order OrderInfo, actor ActorInfo) ([]uuid.UUID, error)` — 業務下單為該客戶全部子帳號(排除主帳號)建 fcm/in_app 各一筆 pending;客戶自行下單回傳 `nil, nil`。
  - `triggers.CustomerProductInfo{ID uuid.UUID; CompanyID uuid.UUID; DepartmentID uuid.UUID; CustomerID uuid.UUID; CustomerName string; ProductName string; DefaultSalesRepID *uuid.UUID}`
  - `triggers.OnCustomerProductCreated(ctx context.Context, tx *ent.Tx, cp CustomerProductInfo, reviewEnabled bool) ([]uuid.UUID, error)` — 推主責業務,NULL/停用退回部門 dept_admin;`reviewEnabled` 對應設定旗標 `NOTIFY_CUSTOMER_PRODUCT_REVIEW`(預設關閉;開啟改用範本 `customer_product_review`,待定決議後僅翻旗標與調範本,不改程式結構)。
  - 呼叫點(相依: `backend-detail/05-sales-orders.md` 下單流程、`backend-detail/04-master-data.md` 專屬商品建立流程):觸發方於**業務交易內**呼叫 `OnXxx` 取得通知 ID,**提交後**呼叫 `notification.Dispatch(ctx, db, sender, rec, ids)`;注入程式碼由 05/04 計畫執行時落地(本計畫 Task 5 提供函式與下方整合範例,標 TODO)。

- [ ] **Step 1: 寫失敗測試(下單路由)**

`backend/internal/notification/triggers/triggers_test.go`:

```go
package triggers_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entnotification "github.com/salesorder/sales-order-1.0/backend/ent/notification"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/notification"
	"github.com/salesorder/sales-order-1.0/backend/internal/notification/triggers"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func setup(t *testing.T) (context.Context, *ent.Client, uuid.UUID, uuid.UUID) {
	t.Helper()
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, err := c.Company.Create().SetName("甲").SetIdentifier("co-a").
		SetCustomerCodePrefix("AA").Save(ctx)
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	dept, err := c.Department.Create().SetCompanyID(co.ID).SetName("業務部").Save(ctx)
	if err != nil {
		t.Fatalf("dept: %v", err)
	}
	return ctx, c, co.ID, dept.ID
}

func seedAccount(t *testing.T, ctx context.Context, c *ent.Client,
	coID uuid.UUID, customerID uuid.UUID, name string, primary bool, status string) uuid.UUID {
	t.Helper()
	u, err := c.User.Create().SetCompanyID(coID).SetRole("customer").
		SetDataScope("self").SetStatus(status).SetIsCustomer(true).
		SetCustomerID(customerID).SetAccountName(name).SetIsPrimary(primary).
		SetPasswordHash("x").Save(ctx)
	if err != nil {
		t.Fatalf("seed account %s: %v", name, err)
	}
	return u.ID
}

func seedTemplate(t *testing.T, ctx context.Context, c *ent.Client,
	coID uuid.UUID, deptID *uuid.UUID, code, channel string) {
	t.Helper()
	b := c.NotificationTemplate.Create().
		SetCompanyID(coID).SetCode(code).SetName(code).
		SetChannel(channel).SetSubject("{{customer_name}} 通知").
		SetBody("{{customer_name}} {{order_no}} {{item_count}} {{product_name}}").
		SetLocale("zh-TW")
	if deptID != nil {
		b.SetDepartmentID(*deptID)
	}
	if _, err := b.Save(ctx); err != nil {
		t.Fatalf("seed template: %v", err)
	}
}

func orderInfo(coID, deptID, customerID uuid.UUID) triggers.OrderInfo {
	return triggers.OrderInfo{
		ID: uuid.New(), OrderNo: "SO-001", CompanyID: coID, DepartmentID: &deptID,
		CustomerID: customerID, CustomerName: "好市多", ItemCount: 3,
	}
}

func TestOrderCreatedByStaffNotifiesSubAccounts(t *testing.T) {
	ctx, c, coID, deptID := setup(t)
	seedTemplate(t, ctx, c, coID, &deptID, "order_created", "fcm")
	seedTemplate(t, ctx, c, coID, &deptID, "order_created", "in_app")
	customerID := uuid.New()
	sub1 := seedAccount(t, ctx, c, coID, customerID, "子一", false, "active")
	sub2 := seedAccount(t, ctx, c, coID, customerID, "子二", false, "active")
	primary := seedAccount(t, ctx, c, coID, customerID, "老闆", true, "active")
	suspended := seedAccount(t, ctx, c, coID, customerID, "停用子", false, "suspended")

	tx, _ := c.Tx(ctx)
	ids, err := triggers.OnOrderCreated(ctx, tx, orderInfo(coID, deptID, customerID),
		triggers.ActorInfo{UserID: uuid.New(), Role: "staff", IsCustomer: false})
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	// 每位子帳號 fcm + in_app 各一筆;主帳號與停用帳號無
	if len(ids) != 4 {
		t.Fatalf("want 4 notifications (2 subs x 2 channels), got %d", len(ids))
	}
	for _, uid := range []uuid.UUID{sub1, sub2} {
		n, err := c.Notification.Query().Where(
			entnotification.UserIDEQ(uid)).Count(ctx)
		if err != nil || n != 2 {
			t.Fatalf("sub %s should have 2 notifications, got %d err=%v", uid, n, err)
		}
	}
	for _, uid := range []uuid.UUID{primary, suspended} {
		n, _ := c.Notification.Query().Where(entnotification.UserIDEQ(uid)).Count(ctx)
		if n != 0 {
			t.Fatalf("primary/suspended account %s should have no notification, got %d", uid, n)
		}
	}
	// payload 帶訂單 ID 供 App 導頁
	one, _ := c.Notification.Get(ctx, ids[0])
	if one.Payload["order_id"] == "" {
		t.Fatalf("payload should carry order_id, got %v", one.Payload)
	}
	// 提交後發送:in_app 標 sent;fcm 無裝置標 failed(no_device)
	if err := notification.Dispatch(ctx, c, &notification.FakeSender{}, audit.NoopRecorder{}, ids); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	inApp, _ := c.Notification.Query().Where(
		entnotification.UserIDEQ(sub1),
		entnotification.ChannelEQ(entnotification.ChannelInApp),
	).Only(ctx)
	if inApp.Status != "sent" {
		t.Fatalf("in_app should be sent after dispatch, got %s", inApp.Status)
	}
}

func TestOrderCreatedByCustomerNotifiesNobody(t *testing.T) {
	ctx, c, coID, deptID := setup(t)
	seedTemplate(t, ctx, c, coID, &deptID, "order_created", "in_app")
	customerID := uuid.New()
	sub1 := seedAccount(t, ctx, c, coID, customerID, "子一", false, "active")
	seedAccount(t, ctx, c, coID, customerID, "子二", false, "active")

	tx, _ := c.Tx(ctx)
	ids, err := triggers.OnOrderCreated(ctx, tx, orderInfo(coID, deptID, customerID),
		triggers.ActorInfo{UserID: sub1, Role: "customer", IsCustomer: true})
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if ids != nil {
		t.Fatalf("customer self-order should notify nobody, got %d ids", len(ids))
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	n, _ := c.Notification.Query().Count(ctx)
	if n != 0 {
		t.Fatalf("no notification should exist, got %d", n)
	}
}

func TestOrderCreatedNoSubAccountsIsNotError(t *testing.T) {
	ctx, c, coID, deptID := setup(t)
	customerID := uuid.New() // 無任何帳號
	tx, _ := c.Tx(ctx)
	ids, err := triggers.OnOrderCreated(ctx, tx, orderInfo(coID, deptID, customerID),
		triggers.ActorInfo{UserID: uuid.New(), Role: "staff", IsCustomer: false})
	if err != nil || ids != nil {
		t.Fatalf("no sub-accounts should log-and-skip, got ids=%v err=%v", ids, err)
	}
	_ = tx.Commit()
}
```

- [ ] **Step 2: 寫失敗測試(專屬商品路由與檢核旗標)**

同檔 `triggers_test.go` 續加:

```go
func seedStaff(t *testing.T, ctx context.Context, c *ent.Client,
	coID, deptID uuid.UUID, role, status string) uuid.UUID {
	t.Helper()
	u, err := c.User.Create().SetCompanyID(coID).SetDepartmentID(deptID).
		SetRole(role).SetDataScope("department").SetStatus(status).
		SetName(role).Save(ctx)
	if err != nil {
		t.Fatalf("seed staff: %v", err)
	}
	return u.ID
}

func cpInfo(coID, deptID uuid.UUID, repID *uuid.UUID) triggers.CustomerProductInfo {
	return triggers.CustomerProductInfo{
		ID: uuid.New(), CompanyID: coID, DepartmentID: deptID,
		CustomerID: uuid.New(), CustomerName: "好市多", ProductName: "特規螺絲",
		DefaultSalesRepID: repID,
	}
}

func TestCustomerProductNotifiesSalesRep(t *testing.T) {
	ctx, c, coID, deptID := setup(t)
	seedTemplate(t, ctx, c, coID, &deptID, "customer_product_created", "in_app")
	seedTemplate(t, ctx, c, coID, &deptID, "customer_product_created", "fcm")
	rep := seedStaff(t, ctx, c, coID, deptID, "staff", "active")
	admin := seedStaff(t, ctx, c, coID, deptID, "dept_admin", "active")

	tx, _ := c.Tx(ctx)
	ids, err := triggers.OnCustomerProductCreated(ctx, tx, cpInfo(coID, deptID, &rep), false)
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	_ = tx.Commit()
	if len(ids) != 2 { // 主責業務 fcm + in_app
		t.Fatalf("want 2 notifications for sales rep, got %d", len(ids))
	}
	n, _ := c.Notification.Query().Where(entnotification.UserIDEQ(rep)).Count(ctx)
	if n != 2 {
		t.Fatalf("rep should have 2 notifications, got %d", n)
	}
	n, _ = c.Notification.Query().Where(entnotification.UserIDEQ(admin)).Count(ctx)
	if n != 0 {
		t.Fatalf("dept_admin should not be notified when rep is set, got %d", n)
	}
}

func TestCustomerProductFallbackToDeptAdmin(t *testing.T) {
	ctx, c, coID, deptID := setup(t)
	seedTemplate(t, ctx, c, coID, &deptID, "customer_product_created", "in_app")
	admin1 := seedStaff(t, ctx, c, coID, deptID, "dept_admin", "active")
	admin2 := seedStaff(t, ctx, c, coID, deptID, "dept_admin", "active")
	disabledRep := seedStaff(t, ctx, c, coID, deptID, "staff", "suspended")

	// 情況一:未設主責業務 → 全部啟用中 dept_admin
	tx, _ := c.Tx(ctx)
	ids, err := triggers.OnCustomerProductCreated(ctx, tx, cpInfo(coID, deptID, nil), false)
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	_ = tx.Commit()
	if len(ids) != 2 { // 2 admins x in_app(僅 in_app 有範本)
		t.Fatalf("want 2 (admins x in_app), got %d", len(ids))
	}
	for _, uid := range []uuid.UUID{admin1, admin2} {
		n, _ := c.Notification.Query().Where(entnotification.UserIDEQ(uid)).Count(ctx)
		if n != 1 {
			t.Fatalf("admin %s should have 1 notification, got %d", uid, n)
		}
	}

	// 情況二:主責業務已停用 → 退回 dept_admin
	tx, _ = c.Tx(ctx)
	ids, err = triggers.OnCustomerProductCreated(ctx, tx, cpInfo(coID, deptID, &disabledRep), false)
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	_ = tx.Commit()
	if len(ids) != 2 {
		t.Fatalf("disabled rep should fall back to dept_admins, got %d ids", len(ids))
	}
	n, _ := c.Notification.Query().Where(entnotification.UserIDEQ(disabledRep)).Count(ctx)
	if n != 0 {
		t.Fatalf("disabled rep should receive nothing, got %d", n)
	}
}

func TestCustomerProductNoRecipientLogsOnly(t *testing.T) {
	ctx, c, coID, deptID := setup(t)
	// 無主責業務且部門無 dept_admin → 記警告日誌,業務照常成功(不發送)
	tx, _ := c.Tx(ctx)
	ids, err := triggers.OnCustomerProductCreated(ctx, tx, cpInfo(coID, deptID, nil), false)
	if err != nil || ids != nil {
		t.Fatalf("no recipient should succeed without sending, got ids=%v err=%v", ids, err)
	}
	_ = tx.Commit()
}

func TestCustomerProductReviewFlagSwitchesTemplate(t *testing.T) {
	ctx, c, coID, deptID := setup(t)
	seedTemplate(t, ctx, c, coID, &deptID, "customer_product_created", "in_app")
	seedTemplate(t, ctx, c, coID, &deptID, "customer_product_review", "in_app")
	rep := seedStaff(t, ctx, c, coID, deptID, "staff", "active")

	// 旗標開啟 → 使用「待檢核」語意範本 customer_product_review
	tx, _ := c.Tx(ctx)
	ids, err := triggers.OnCustomerProductCreated(ctx, tx, cpInfo(coID, deptID, &rep), true)
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	_ = tx.Commit()
	if len(ids) != 1 {
		t.Fatalf("want 1 notification, got %d", len(ids))
	}
	n, _ := c.Notification.Get(ctx, ids[0])
	tmpl, _ := c.NotificationTemplate.Get(ctx, *n.TemplateID)
	if tmpl.Code != "customer_product_review" {
		t.Fatalf("review flag on should use review template, got %s", tmpl.Code)
	}

	// 旗標關閉 → 使用「已新增」語意範本 customer_product_created
	tx, _ = c.Tx(ctx)
	ids, err = triggers.OnCustomerProductCreated(ctx, tx, cpInfo(coID, deptID, &rep), false)
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	_ = tx.Commit()
	n, _ = c.Notification.Get(ctx, ids[0])
	tmpl, _ = c.NotificationTemplate.Get(ctx, *n.TemplateID)
	if tmpl.Code != "customer_product_created" {
		t.Fatalf("review flag off should use created template, got %s", tmpl.Code)
	}
}
```

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/notification/triggers/ -v`
Expected: FAIL — `triggers.OnOrderCreated` / `OnCustomerProductCreated` 未定義(編譯失敗)。

- [ ] **Step 4: 實作 order_created.go**

`backend/internal/notification/triggers/order_created.go`:

```go
// Package triggers 為業務事件的通知路由(細部文件 4.4.3、4.4.4;路由總表見細部文件 §0)。
package triggers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entuser "github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/notifications"
)

// OrderInfo 為觸發所需最小訂單資訊(由 05-sales-orders-plan 下單流程填入)。
type OrderInfo struct {
	ID           uuid.UUID
	OrderNo      string
	CompanyID    uuid.UUID
	DepartmentID *uuid.UUID
	CustomerID   uuid.UUID
	CustomerName string
	ItemCount    int
}

// ActorInfo 為操作者身分;IsCustomer 由 01-auth 身分模型判別(D22)。
type ActorInfo struct {
	UserID     uuid.UUID
	Role       string
	IsCustomer bool
}

// OnOrderCreated 於訂單建立的同一 DB 交易內呼叫(D18):
// 業務/內部員工下單 → 為該客戶全部「啟用中子帳號」建 fcm/in_app 各一筆 pending;
// 主帳號排除(管理用途,D22);客戶自行下單 → 不發任何通知。
// 回傳建立的通知 ID;呼叫方於交易提交後交給 notification.Dispatch。
// 冪等:通知建檔與訂單同交易,訂單建立失敗(如取號衝突耗盡重試)一併回滾,無殘留 pending。
func OnOrderCreated(ctx context.Context, tx *ent.Tx, order OrderInfo, actor ActorInfo) ([]uuid.UUID, error) {
	// 客戶子帳號自行下單:直接結束,避免客戶收到自己操作的通知(細部文件 4.4.3 步驟 1)
	if actor.IsCustomer {
		return nil, nil
	}
	// 接收者:該客戶全部帳號中 is_primary=false 且未停用、未軟刪除(步驟 2)
	subs, err := tx.User.Query().Where(
		entuser.CustomerIDEQ(order.CustomerID),
		entuser.IsCustomerEQ(true),
		entuser.IsPrimaryEQ(false),
		entuser.StatusEQ("active"),
		entuser.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("triggers: query sub-accounts: %w", err)
	}
	if len(subs) == 0 {
		// 客戶無子帳號:記日誌正常結束,不視為錯誤、不阻斷下單
		log.Printf("triggers: order_created customer %s has no active sub-accounts; skip", order.CustomerID)
		return nil, nil
	}
	vars := map[string]string{
		"customer_name": order.CustomerName,
		"order_no":      order.OrderNo,
		"item_count":    strconv.Itoa(order.ItemCount),
	}
	payload := map[string]any{"type": "order_created", "order_id": order.ID.String()}
	userIDs := make([]uuid.UUID, len(subs))
	for i, s := range subs {
		userIDs[i] = s.ID
	}
	q := notifications.TemplateQuery{
		CompanyID: order.CompanyID, DepartmentID: order.DepartmentID,
		Code: "order_created", Locale: "",
	}
	ids := make([]uuid.UUID, 0, len(userIDs)*2)
	// 雙通道:同一接收者 fcm / in_app 各一筆,已讀互不影響(細部文件 4.4.2 步驟 3)
	for _, channel := range []string{"fcm", "in_app"} {
		q.Channel = channel
		created, err := createForUsers(ctx, tx, q, userIDs, vars, payload)
		if err != nil {
			return nil, err
		}
		ids = append(ids, created...)
	}
	return ids, nil
}

// createForUsers 為每位接收者渲染範本並建一筆 pending 通知(雙通道共用);
// 範本缺失 → 降級建 template_id 為空的通知(細部文件 4.3.2 錯誤處理);
// 建檔失敗 → 回傳 error 使觸發方交易回滾(上層感知 internal)。
// 通知觸發本身不另寫稽核(下單稽核已涵蓋,細部文件 4.4.3 步驟 4)。
func createForUsers(ctx context.Context, tx *ent.Tx, q notifications.TemplateQuery,
	userIDs []uuid.UUID, vars map[string]string, payload map[string]any) ([]uuid.UUID, error) {
	repo := notifications.Repo{}
	ids := make([]uuid.UUID, 0, len(userIDs))
	for _, uid := range userIDs {
		q.Locale = accountLocale(uid)
		r, err := notifications.Render(ctx, tx.Client(), q, vars)
		var tmplID *uuid.UUID
		var title, content string
		switch {
		case err == nil:
			tmplID, title, content = &r.TemplateID, r.Title, r.Content
		case errors.Is(err, notifications.ErrTemplateNotFound):
			log.Printf("triggers: template %s/%s missing, degrade to template-less record", q.Code, q.Channel)
			title, content = q.Code, fmt.Sprintf("%s %s", vars["customer_name"], vars["order_no"])
		default:
			return nil, fmt.Errorf("triggers: render %s/%s: %w", q.Code, q.Channel, err)
		}
		id, err := repo.CreateOne(ctx, tx, notifications.CreateParams{
			CompanyID: q.CompanyID, DepartmentID: q.DepartmentID, UserID: uid,
			TemplateID: tmplID, Channel: q.Channel,
			Title: title, Content: content, Payload: payload,
		})
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// accountLocale 取接收者帳號語系。
// TODO(相依: 01-auth-plan Task 1 已落地 users.locale):執行時
// 改為查詢該欄位;現階段一律回空字串,由 notifications.Render 退回預設語系(DefaultLocale)。
func accountLocale(_ uuid.UUID) string { return "" }
```

- [ ] **Step 5: 實作 customer_product_created.go**

`backend/internal/notification/triggers/customer_product_created.go`:

```go
package triggers

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entuser "github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/notifications"
)

// CustomerProductInfo 為觸發所需最小專屬商品資訊(由 04-master-data-plan 建立流程填入;
// DefaultSalesRepID 對應 customers.default_sales_rep_id)。
type CustomerProductInfo struct {
	ID                uuid.UUID
	CompanyID         uuid.UUID
	DepartmentID      uuid.UUID
	CustomerID        uuid.UUID
	CustomerName      string
	ProductName       string
	DefaultSalesRepID *uuid.UUID
}

// OnCustomerProductCreated 於專屬商品建立的同一 DB 交易內呼叫(僅後台/Web 觸發;
// 批次匯入逐筆觸發,細部文件 4.4.4 步驟 1)。
// 接收者:主責業務(default_sales_rep_id)啟用中 → 主責業務;
// 為 NULL 或已停用/離職 → 同部門全部啟用中 dept_admin;兩者皆無 → 記警告日誌不發送。
// reviewEnabled 對應設定旗標 NOTIFY_CUSTOMER_PRODUCT_REVIEW(檢核待定項,D16/D23):
// true → 範本 customer_product_review(「待檢核」語意);false(預設)→ customer_product_created
// (「已新增」語意);待定決議後僅翻旗標與調範本,不改程式結構。
func OnCustomerProductCreated(ctx context.Context, tx *ent.Tx, cp CustomerProductInfo, reviewEnabled bool) ([]uuid.UUID, error) {
	recipients, err := resolveProductRecipients(ctx, tx, cp)
	if err != nil {
		return nil, err
	}
	if len(recipients) == 0 {
		log.Printf("triggers: customer_product_created for customer %s: no sales rep and no dept_admin in dept %s; skip",
			cp.CustomerID, cp.DepartmentID)
		return nil, nil
	}
	code := "customer_product_created"
	if reviewEnabled {
		code = "customer_product_review"
	}
	vars := map[string]string{
		"customer_name": cp.CustomerName,
		"product_name":  cp.ProductName,
	}
	payload := map[string]any{
		"type": "customer_product_created", "customer_product_id": cp.ID.String(),
	}
	q := notifications.TemplateQuery{
		CompanyID: cp.CompanyID, DepartmentID: &cp.DepartmentID, Code: code,
	}
	ids := make([]uuid.UUID, 0, len(recipients)*2)
	for _, channel := range []string{"fcm", "in_app"} {
		q.Channel = channel
		created, err := createForUsers(ctx, tx, q, recipients, vars, payload)
		if err != nil {
			return nil, err // 建檔失敗 → 專屬商品建立交易回滾,回 internal
		}
		ids = append(ids, created...)
	}
	return ids, nil
}

// resolveProductRecipients 解析接收者(細部文件 4.4.4 步驟 2)。
func resolveProductRecipients(ctx context.Context, tx *ent.Tx, cp CustomerProductInfo) ([]uuid.UUID, error) {
	if cp.DefaultSalesRepID != nil {
		ok, err := tx.User.Query().Where(
			entuser.IDEQ(*cp.DefaultSalesRepID),
			entuser.StatusEQ("active"),
			entuser.DeletedAtIsNil(),
		).Exist(ctx)
		if err != nil {
			return nil, fmt.Errorf("triggers: check sales rep: %w", err)
		}
		if ok {
			return []uuid.UUID{*cp.DefaultSalesRepID}, nil
		}
		log.Printf("triggers: sales rep %s inactive/missing, fall back to dept_admin", *cp.DefaultSalesRepID)
	}
	admins, err := tx.User.Query().Where(
		entuser.CompanyIDEQ(cp.CompanyID),
		entuser.DepartmentIDEQ(cp.DepartmentID),
		entuser.RoleEQ("dept_admin"),
		entuser.StatusEQ("active"),
		entuser.DeletedAtIsNil(),
	).IDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("triggers: query dept_admin: %w", err)
	}
	return admins, nil
}
```

- [ ] **Step 6: 整合範例(觸發方呼叫點,由 05/04 計畫落地)**

於 `backend/internal/domain/orders/`(05-sales-orders-plan 檔案)訂單建立流程插入:

```go
	// TODO(相依: 07-notifications-plan Task 5):下單通知路由(細部文件 4.4.3)
	notifyIDs, err := triggers.OnOrderCreated(ctx, tx, triggers.OrderInfo{
		ID: order.ID, OrderNo: order.OrderNo, CompanyID: order.CompanyID,
		DepartmentID: order.DepartmentID, CustomerID: order.CustomerID,
		CustomerName: customerName, ItemCount: len(items),
	}, triggers.ActorInfo{UserID: actor.UserID, Role: actor.Role, IsCustomer: actor.IsCustomer})
	if err != nil {
		_ = tx.Rollback() // 通知建檔失敗 → 下單整體回滾,呼叫端收 internal(D18)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// FCM 外部呼叫在提交後;失敗只標 failed,不影響訂單(D16/D18)
	if len(notifyIDs) > 0 {
		go func() {
			if err := notification.Dispatch(context.Background(), db, fcmSender, auditRec, notifyIDs); err != nil {
				log.Printf("notification dispatch: %v", err)
			}
		}()
	}
```

於 `backend/internal/domain/products/`(04-master-data-plan 檔案)專屬商品建立流程插入同形片段,呼叫 `triggers.OnCustomerProductCreated(ctx, tx, info, cfg.NotifyCustomerProductReview)`;config 加:

```go
// NotifyCustomerProductReview 對應環境變數 NOTIFY_CUSTOMER_PRODUCT_REVIEW(預設 false;
// 檢核待定項,決議後僅翻旗標與調範本,細部文件 4.4.4 步驟 3)。
NotifyCustomerProductReview bool
```

- [ ] **Step 7: 跑測試確認通過**

Run: `cd backend && go test ./internal/notification/... -v`
Expected: PASS — triggers 7 測試 + Task 4 既有 7 測試全綠。

- [ ] **Step 8: Commit**

```bash
git add backend/internal/notification/triggers
git commit -m "feat(backend): 下單推子帳號、專屬商品推主責業務通知路由(4.4.3-4.4.4)"
```

---


### Task 6: 派車通知 adapter(對接 08-dispatch 計畫 `DispatchNotifier`)

**Files:**
- Create: `backend/internal/notification/triggers/dispatch_notifier.go`
- Test: `backend/internal/notification/triggers/dispatch_notifier_test.go`

**Interfaces:**
- Consumes: Task 2 `notifications.Render` / `notifications.Repo.CreateOne`;Task 4 `notification.Dispatch` / `notification.Sender`;08-dispatch 計畫 T1 定義的 `salesorders.DispatchNotifier` / `salesorders.DispatchNotification`(本 Task 實作該介面;依賴方向:triggers → salesorders 型別,無循環)。
- Produces: `triggers.NewDispatchNotifier(db *ent.Client, sender Sender, rec audit.Recorder) salesorders.DispatchNotifier` — `InitDomains()` 組裝點以此替換 08 計畫的 `noopDispatchNotifier`(08 計畫 T1 TODO 解除)。
- 範本:`dispatch`(fcm / in_app 兩通道;變數 `order_no` / `route_name` / `expected_delivery_date`);範本由管理者經範本 CRUD 建立,缺失降級為無範本記錄(同 T5 行為)。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/notification/triggers/dispatch_notifier_test.go`:

```go
package triggers_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/notification"
	"github.com/salesorder/sales-order-1.0/backend/internal/notification/triggers"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestDispatchNotifierCreatesAndDispatches(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx, coID, deptID := setup(t)
	custID := uuid.New()
	sub := seedAccount(t, ctx, c, coID, custID, "子帳號", false, "active")
	_ = seedAccount(t, ctx, c, coID, custID, "主帳號", true, "active")

	sender := &notification.FakeSender{}
	n := triggers.NewDispatchNotifier(c, sender, audit.NoopRecorder{})
	err := n.NotifyDispatched(ctx, salesorders.DispatchNotification{
		CompanyID: coID, DepartmentID: deptID,
		SalesOrderID: uuid.New(), CustomerID: custID,
		OrderNo: "W000001", RouteName: "北一車", ExpectedDeliveryDate: "2026-08-22",
	})
	if err != nil { t.Fatalf("notify: %v", err) }
	// 僅子帳號(排除主帳號)fcm + in_app 各一筆,範本缺失時降級無範本記錄
	var notifs []struct{ UserID uuid.UUID; Channel string }
	rows := c.Notification.Query().Where().AllX(ctx)
	if len(rows) != 2 {
		t.Fatalf("want 2 notifications, got %d", len(rows))
	}
	for _, r := range rows {
		if r.UserID != sub {
			t.Fatalf("wrong recipient: %+v", r)
		}
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/notification/triggers/ -run TestDispatchNotifier -v`
Expected: FAIL — `NewDispatchNotifier` 未定義。

- [ ] **Step 3: 實作 dispatch_notifier.go**

```go
package triggers

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/notifications"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/salesorders"
	"github.com/salesorder/sales-order-1.0/backend/internal/notification"
)

// dispatchNotifier 實作 salesorders.DispatchNotifier(08-dispatch 計畫 T1 介面)。
// 接收者路由同 OnOrderCreated:該客戶全部啟用中子帳號(排除主帳號)。
// fire-and-record:任何失敗僅記 log,不回錯給派車流程(D16)。
type dispatchNotifier struct {
	db     *ent.Client
	sender notification.Sender
	rec    audit.Recorder
}

func NewDispatchNotifier(db *ent.Client, sender notification.Sender, rec audit.Recorder) salesorders.DispatchNotifier {
	return &dispatchNotifier{db: db, sender: sender, rec: rec}
}

func (d *dispatchNotifier) NotifyDispatched(ctx context.Context, n salesorders.DispatchNotification) error {
	tx, err := d.db.Tx(ctx)
	if err != nil {
		log.Printf("triggers: dispatch notify tx: %v", err)
		return nil // fire-and-record
	}
	recipients, err := activeSubAccounts(ctx, tx, n.CustomerID) // 同 T5 的接收者查詢
	if err != nil {
		_ = tx.Rollback()
		log.Printf("triggers: dispatch notify recipients: %v", err)
		return nil
	}
	vars := map[string]string{
		"order_no": n.OrderNo, "route_name": n.RouteName,
		"expected_delivery_date": n.ExpectedDeliveryDate,
	}
	var ids []uuid.UUID
	for _, uid := range recipients {
		for _, channel := range []string{"fcm", "in_app"} {
			q := notifications.TemplateQuery{
				CompanyID: n.CompanyID, DepartmentID: &n.DepartmentID,
				Code: "dispatch", Channel: channel, Locale: notifications.DefaultLocale,
			}
			var tmplID *uuid.UUID
			title, content := q.Code, fmt.Sprintf("%s %s", n.OrderNo, n.RouteName)
			r, err := notifications.Render(ctx, d.db, q, vars)
			switch {
			case err == nil:
				tmplID, title, content = &r.TemplateID, r.Title, r.Content
			case errors.Is(err, notifications.ErrTemplateNotFound):
				// 降級無範本記錄
			default:
				_ = tx.Rollback()
				log.Printf("triggers: dispatch notify render: %v", err)
				return nil
			}
			id, err := notifications.NewRepo().CreateOne(ctx, tx, notifications.CreateParams{
				CompanyID: n.CompanyID, DepartmentID: &n.DepartmentID, UserID: uid,
				TemplateID: tmplID, Channel: channel, Title: title, Content: content,
				Payload: map[string]any{"sales_order_id": n.SalesOrderID.String()},
			})
			if err != nil {
				_ = tx.Rollback()
				log.Printf("triggers: dispatch notify create: %v", err)
				return nil
			}
			ids = append(ids, id)
		}
	}
	if err := tx.Commit(); err != nil {
		log.Printf("triggers: dispatch notify commit: %v", err)
		return nil
	}
	if len(ids) > 0 {
		if err := notification.Dispatch(ctx, d.db, d.sender, d.rec, ids); err != nil {
			log.Printf("triggers: dispatch notify dispatch: %v", err)
		}
	}
	return nil
}
```

(`activeSubAccounts` 於 T5 已有接收者查詢邏輯,抽為共用函式。)

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go test ./internal/notification/... -v`
Expected: PASS — 接收者僅子帳號、雙通道各一筆、範本缺失降級。

```bash
git add backend/internal/notification/triggers/dispatch_notifier.go
git commit -m "feat(backend): 派車通知 adapter,實作 08 計畫 DispatchNotifier(4.4.3 派車情境)"
```

---

## Self-Review 記錄

- **Spec 覆蓋**:細部文件 10 子功能 → Task 對應:4.3.1→T1;4.3.2→T2(render.go + Preview);4.3.3→T2(repo.go + service.go);4.3.4→T3(device_service.go + PurgeInvalidTokens);4.3.5→T3(promotag.go + RLS + 宿主欄位規格);4.4.1→T4(fcm.go Sender/NewFCMSender/Dispatch);4.4.2→T4(inapp.go + Dispatch in_app 分支);4.4.3→T5(order_created.go);4.4.4→T5(customer_product_created.go + 旗標);4.4.5→T4(failmark.go MarkFailed + FailuresTotal);派車通知(08 計畫 5.1.4 觸發的 `DispatchNotifier` 實作)→T6(dispatch_notifier.go)。無缺漏。
- **細部文件「整合測試重點」覆蓋**:①端到端路由→T5 Step 1(`TestOrderCreatedByStaffNotifiesSubAccounts` 含提交後 Dispatch、主帳號排除、`TestOrderCreatedByCustomerNotifiesNobody`);②專屬商品三組合→T5 Step 2;③裝置生命週期→T3 Step 1(註冊/歸屬轉移/註銷)+ T4 Step 1(`dead-tok` purge 後軟刪除);④狀態機→T2 Step 2(MarkRead 狀態機)+ T4 Step 2(終態不可再轉、sent 不被 failed 覆蓋);⑤交易一致性→T4 Dispatch 失敗僅標 failed + T5 建檔失敗回滾路徑(程式碼路徑;整合環境強迫失敗屬執行期驗收,照細部文件由 Phase 4 Task 4.8 執行);⑥RLS 隔離→T1 `TestNotificationRLSIsolation`、T3 `TestPromoTagRLSDepartmentIsolation`;⑦範本退回→T2 `TestRenderLocaleFallbackChain`。
- **已知佔位(皆標 TODO + 接手方,屬跨 domain/跨 Phase 依賴,非本計畫範圍)**:
  - `products` / `customer_products` / `customers` 三表 `promo_tag_ids` 欄位(→ 04-master-data-plan 對應 schema Task;本計畫 T3 提供精確欄位碼與行為註記);其 CRUD、客戶訂閱、選群推播 → Phase 7 Task 7.4。
  - 觸發點注入:訂單建立流程(05-sales-orders-plan Task 5 已直接消費 `triggers.OnOrderCreated` + `notification.Dispatch`);退貨審核結果(06-returns-plan Task 3 直接消費 `Render`/`CreateOne`/`Dispatch`,範本 `return_reviewed`);派車通知(本計畫 T6 `NewDispatchNotifier` 實作 08 計畫 `DispatchNotifier` 介面);專屬商品建立流程(→ 04-master-data-plan 執行時掛接 `OnCustomerProductCreated`)。
  - `triggers.accountLocale`(→ `users.locale` 已於 01-auth-plan Task 1 schema 落地;執行時改為查詢該欄位,空值退回預設語系)。
  - `audit.Recorder` DB 實作(→ 03-metadicts-audit-plan Task 2.6;介面與 no-op 由 01-auth-plan Task 14 提供)。
  - `NOTIFY_CUSTOMER_PRODUCT_REVIEW` 檢核待定項:旗標預設關閉,決議後翻旗標 + 調範本,不改程式結構(細部文件 4.4.4 步驟 3)。
  - migration 檔名已全域定序(00013 / 00014;見 00 計畫群跨檔編號約定)。
- **類型一致**:`notifications.TemplateQuery` / `Rendered` / `Repo.CreateOne(CreateParams)`(T2 定義,T4 `SendInApp`、T5 `createForUsers` 重用簽名一致);`notification.Sender` / `SendResult` / `Dispatch` / `MarkFailed(FailReason)`(T4 定義,T5 呼叫一致);`notifications.PurgeInvalidTokens`(T3 定義,T4 `dispatchFCM` 呼叫一致);`rls.Identity{UserID, CompanyID, DepartmentID, DataScope, IsPrimary}`(01-auth-plan Task 3/13 定義,T2/T3 `actor()` 使用一致);`audit.Recorder.Record(ctx, Entry)`(01-auth-plan Task 14 定義,T3/T4 注入一致)。
- **狀態機一致**:`pending → sent → read`、`pending → failed`(終態)、`pending → read` 僅三條合法路徑;`MarkReadBatch`(T2)拒絕 failed、`MarkFailed`(T4)僅作用於 pending、`markSent`(T4)僅作用於 pending,三路互不重疊。

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-08-17-backend-07-notifications-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** — 每個 Task 派新 subagent 執行,Task 間 review,迭代快。

**2. Inline Execution** — 用 executing-plans 在本 session 逐批執行,設 checkpoint review。

Which approach?

---

*計畫版本:v1.0.0(2026-08-17);對應細部文件 backend-detail/07-notifications.md v1.0.0、原計畫 v2.9.0、規格書 v1.0.34。*
