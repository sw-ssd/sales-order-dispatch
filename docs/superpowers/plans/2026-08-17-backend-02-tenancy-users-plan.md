# Backend 02 — 多租戶主檔:公司 / 部門 / 使用者、角色權限、Casbin Policy 管理 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 實作 backend Phase 2 的多租戶主檔與授權管理 — CompanyService / DepartmentService / UserService CRUD 與三級管理範圍、公司停用連鎖阻斷、Logo 上傳與公開發現端點、roles + role_permissions 表與內建角色 seed、GetAbility 改表驅動、RLS data_scope 改由角色等級注入、Casbin policy 管理 API(即時生效、domain 範圍、防鎖死、g 規則檢視)。

**Architecture:** 依 `docs/superpowers/plans/backend-detail/02-tenancy-users.md`(下稱「細部文件」,子功能編號 2.x.y)實作。沿用 01-auth 計畫(下稱「01 計畫」)建立的地基:Ent schema、`rls.Identity` / `rls.WrapDriver`、`middleware.Authenticate`(已含公司 status 檢查掛點)、`session.Manager`、`authz.NewEnforcer`、`audit.Recorder`(Noop 佔位)。授權三層分工不變:Casbin 管功能(domain = company_id 字串)、RLS 管資料範圍(data_scope 等級,不依角色名稱)、CASL 管 UI。本計畫把「功能權限來源」從程式內建表搬到 `role_permissions` / `casbin_rule` 兩張表,讓 Web 調整即時生效。

**Tech Stack:** Go 1.25、Ent(entgo.io)、Chi v5、Connect-RPC、casbin/v2 + gorm-adapter(postgres)、pgx/v5、go-redis(Valkey,pub/sub 廣播)、testcontainers-go(整合測試);測試共用 `testutil.NewEntClient` / `testutil.NewEntClientWithRLS` / `testutil.DSN`(01 計畫 Task 1–3 提供)。

**Spec 來源:** 細部文件 `docs/superpowers/plans/backend-detail/02-tenancy-users.md`;共通規則見 `docs/superpowers/plans/backend-detail/00-index.md` §3。

## Global Constraints

- module 路徑:`github.com/salesorder/sales-order-1.0/backend`;所有路徑相對 repo root。
- 軟刪除:`companies` / `departments` / `users`(僅客戶帳號)/ `roles`(僅自訂角色)統一 `deleted_at` + 部分唯一索引 `WHERE deleted_at IS NULL`(D10);列表預設排除已刪除;員工帳號不刪除、僅停用;內建角色 `is_system = true` 不可刪除。
- 稽核:所有寫入 RPC 的業務異動與稽核同一 DB 交易(D18),快照含 `before` / `after`,**敏感欄位(密碼、token)永不入稽核**。本計畫統一呼叫 01 計畫 Task 14 的 `audit.Recorder` 介面(現為 `NoopRecorder`);DB 實作由 03-metadicts-audit 計畫 Task 2.6 接手,接手時不得改動本計畫的呼叫點簽名。
- 錯誤:RPC 層統一 Connect code — `unauthenticated` / `permission_denied` / `not_found` / `failed_precondition` / `invalid_argument` / `already_exists`;對應語意依細部文件共通規則 4。
- 範圍:凡「管理所屬公司」的操作,usecase 入口先比對操作者 `rls.Identity.CompanyID` 與目標資源 `company_id`,不符即 `permission_denied`;RLS 為最後防線,usecase 層仍自行加查詢條件(D3)。
- 用語對應:細部文件 2.3.2 的帳號狀態「inactive」對應 `users.status` 列舉值 `suspended`(01 計畫 Task 1 schema 只有 pending / active / suspended);「重新啟用」= UpdateUser 將 status 改回 `active`。
- 角色指派雙層:`users.role`(列舉)為主角色快照;實際授權以 Casbin g 規則(`g, <user_id>, <role_code>, <company_id>`)為唯一來源,自訂角色僅存在於 g 規則。身分解析(middleware)一律展開 g 規則。
- 分頁:列表 RPC 統一 `PageRequest{page, per_page}` / `PageMeta{total, page, per_page}`(本計畫 Task 1 建 `common.proto`);`per_page` 上限 100,超過回 `invalid_argument`。
- RLS session 變數:`app.current_company_id` / `app.current_department_id` / `app.current_data_scope` / `app.current_customer_id`;未注入 fail-closed(01 計畫 Task 3)。
- proto 慣例:檔案置 `backend/proto/v1/`,package `salesorder.v1`;proto 變更後跑 `cd backend && buf generate`(Phase 0 管線已配置)再跑測試。`user.proto` 已由 01 計畫 Task 12 建立(含 `UserService.ForceLogout`),本計畫於同一 service 增量新增 RPC,不重複宣告。
- 測試:DB 相依測試走 testcontainers Postgres 16;`go test ./...` 必須全綠;覆蓋率目標 70%(CI 於 Phase 0 管線強制)。
- 每個 Task 結尾 commit;commit message 格式 `feat(backend): …` / `test(backend): …`。

## File Structure

| 檔案 | 職責 | 建立於 |
|---|---|---|
| `backend/ent/schema/company.go` | 擴充 branding / 公開資訊欄位、prefix 改可空 | Task 1(更新) |
| `backend/database/migrations/00004_companies_branding.sql` | 既有庫欄位與唯一索引調整 | Task 1 |
| `backend/proto/v1/common.proto` | PageRequest / PageMeta | Task 1 |
| `backend/proto/v1/company.proto` | CompanyService(含 Task 4 增量) | Task 1、4 |
| `backend/internal/domain/companies/{usecase,repository,handler}.go` | 公司 CRUD、唯一性、停用連鎖 | Task 1 |
| `backend/internal/middleware/auth.go` | developer 豁免公司停用檢查 | Task 1(更新) |
| `backend/internal/domain/auth/customer.go` / `oauth.go` | 登入端點公司 status 檢查(2.1.3) | Task 1(更新) |
| `backend/internal/database/rls.go` | `Identity.Roles` 欄位 + `HasRole` | Task 1(更新) |
| `backend/proto/v1/department.proto`、`backend/internal/domain/departments/{usecase,repository,handler}.go` | 部門 CRUD | Task 2 |
| `backend/internal/database/txsql.go` | ent.Tx 內 raw SQL 輔助(casbin_rule 用) | Task 3 |
| `backend/proto/v1/user.proto`、`backend/internal/domain/users/{usecase,repository,handler}.go` | 使用者 CRUD、角色指派、停用、ForceLogout 範圍銜接 | Task 3 |
| `backend/internal/domain/companies/branding.go` / `public.go` | Logo 上傳、Branding/PublicInfo、公開發現端點 | Task 4 |
| `backend/config/storage.go` | `StorageRoot` | Task 4(更新) |
| `backend/ent/schema/{role,rolepermission}.go` | roles / role_permissions schema | Task 5 |
| `backend/database/migrations/00005_roles_seed.sql` | 七內建角色 + 預設權限冪等 seed | Task 5 |
| `backend/proto/v1/admin.proto` | RoleService(Task 5)、PolicyService(Task 7) | Task 5、7 |
| `backend/internal/domain/roles/{usecase,repository,handler}.go` | 角色 CRUD、權限矩陣 | Task 5、6 |
| `backend/internal/domain/auth/ability.go` | GetAbility 改 role_permissions 驅動 | Task 6(更新) |
| `backend/database/migrations/00006_rls_data_scope.sql` | RLS policy 統一四級 data_scope 改寫 | Task 6 |
| `backend/database/migrations/00007_casbin_policies_seed.sql` | 預設 p 規則(method path 詞彙)冪等 seed | Task 7 |
| `backend/internal/middleware/authz_interceptor.go` | Connect 授權攔截器(接手 01 計畫佔位) | Task 7 |
| `backend/internal/domain/policies/{usecase,repository,handler,reload.go}` | Policy CRUD、防鎖死、ListGrouping、跨 replica 廣播 | Task 7 |
| `backend/internal/authz/seed.go` | 程式內 seeder 改驗證-only | Task 7(更新) |

---

### Task 1: 公司 schema 擴充 + CompanyService CRUD + 唯一性 + 停用連鎖(細部 2.1.1–2.1.3)

**Files:**
- Update: `backend/ent/schema/company.go`
- Create: `backend/database/migrations/00004_companies_branding.sql`
- Create: `backend/proto/v1/common.proto`
- Create: `backend/proto/v1/company.proto`
- Create: `backend/internal/domain/companies/usecase.go`
- Create: `backend/internal/domain/companies/repository.go`
- Create: `backend/internal/domain/companies/handler.go`
- Update: `backend/internal/database/rls.go`(`Identity.Roles` + `HasRole`)
- Update: `backend/internal/middleware/auth.go`(developer 豁免)
- Update: `backend/internal/middleware/auth_middleware_test.go`
- Update: `backend/internal/domain/auth/customer.go`、`backend/internal/domain/auth/oauth.go`(登入前公司 status 檢查)
- Test: `backend/internal/domain/companies/usecase_test.go`

**Interfaces:**
- Consumes: 01 計畫 Task 1 `ent.Company`;Task 3 `rls.Identity`;Task 11 `middleware.Authenticate`(公司 status 檢查掛點已存在,已登入請求阻斷不在此重實作);Task 14 `audit.Recorder`;Task 6 `CustomerLogin` 與 Task 4 OAuth callback 的核發憑證路徑。
- Produces: proto `CompanyService` — `ListCompanies(PageRequest、status、keyword) → ListCompaniesResponse{companies, meta}`;`GetCompany(company_id) → Company`;`CreateCompany(name、tax_id、display_name、identifier、customer_code_prefix) → Company`;`UpdateCompany(company_id + 可變欄位) → Company`;`DeleteCompany(company_id) → Empty`。usecase `companies.Usecase{List/Get/Create/Update/Delete}`、`companies.NewUsecase(db *ent.Client, rec audit.Recorder) *Usecase`、`companies.EnsureCompanyActive(ctx, db, companyID) error`(登入端點共用,2.1.3 步驟 3)。`rls.Identity` 新增 `Roles []string`(Task 6 填入,本 Task 恆空)+ 方法 `HasRole(code string) bool`。
- 公司新欄位:`display_name`、`primary_color`、`public_email`、`public_phone`、`public_address`、`terms_url`、`privacy_url`、`logo_url`;`customer_code_prefix` 改可空(2.1.2:允許 NULL,未設定前 Phase 3 不可取號)。

- [ ] **Step 1: 寫失敗測試(CRUD、唯一性、範圍、軟刪除重建)**

`backend/internal/domain/companies/usecase_test.go`:

```go
package companies_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/companies"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func superActor() rls.Identity {
	return rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(), DataScope: "all", Role: "super"}
}

func TestCompanyCreateAndUniqueness(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	uc := companies.NewUsecase(c, audit.NoopRecorder{})

	prefix := "AB"
	co, err := uc.Create(ctx, superActor(), companies.CreateInput{
		Name: "甲公司", Identifier: "co-a", CustomerCodePrefix: &prefix,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if co.Status != "active" {
		t.Fatalf("initial status should be active, got %q", co.Status)
	}

	// identifier 重複 → already_exists(錯誤訊息指明欄位)
	_, err = uc.Create(ctx, superActor(), companies.CreateInput{Name: "乙", Identifier: "co-a"})
	if connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("dup identifier: want already_exists, got %v", err)
	}
	// prefix 重複 → already_exists
	_, err = uc.Create(ctx, superActor(), companies.CreateInput{
		Name: "丙", Identifier: "co-c", CustomerCodePrefix: &prefix,
	})
	if connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("dup prefix: want already_exists, got %v", err)
	}
	// prefix 格式不符(小寫 / 超長)→ invalid_argument
	bad := "ab"
	if _, err = uc.Create(ctx, superActor(), companies.CreateInput{
		Name: "丁", Identifier: "co-d", CustomerCodePrefix: &bad,
	}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("bad prefix: want invalid_argument, got %v", err)
	}
	// identifier 格式不符 → invalid_argument
	if _, err = uc.Create(ctx, superActor(), companies.CreateInput{
		Name: "戊", Identifier: "Co_Bad",
	}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("bad identifier: want invalid_argument, got %v", err)
	}
	// prefix 可留 NULL(2.1.2)
	if _, err = uc.Create(ctx, superActor(), companies.CreateInput{
		Name: "己", Identifier: "co-f",
	}); err != nil {
		t.Fatalf("null prefix should be allowed: %v", err)
	}
}

func TestCompanyUpdateRules(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	uc := companies.NewUsecase(c, audit.NoopRecorder{})
	co, _ := uc.Create(ctx, superActor(), companies.CreateInput{Name: "甲", Identifier: "co-a"})

	// identifier 不可修改 → invalid_argument
	ident := "co-a2"
	if _, err := uc.Update(ctx, superActor(), co.ID, companies.UpdateInput{Identifier: &ident}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("identifier update: want invalid_argument, got %v", err)
	}
	// company_admin 改自己公司非 status 欄位 OK;改 status → permission_denied
	admin := rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DataScope: "company", Role: "company_admin"}
	newName := "甲改名"
	if _, err := uc.Update(ctx, admin, co.ID, companies.UpdateInput{Name: &newName}); err != nil {
		t.Fatalf("company_admin update own: %v", err)
	}
	susp := "suspended"
	if _, err := uc.Update(ctx, admin, co.ID, companies.UpdateInput{Status: &susp}); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("company_admin status change: want permission_denied, got %v", err)
	}
	// company_admin 碰其他公司 → permission_denied
	other := rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(), DataScope: "company", Role: "company_admin"}
	if _, err := uc.Update(ctx, other, co.ID, companies.UpdateInput{Name: &newName}); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("cross-company update: want permission_denied, got %v", err)
	}
	if _, err := uc.Get(ctx, other, co.ID); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("cross-company get: want permission_denied, got %v", err)
	}
	// 不存在 → not_found
	if _, err := uc.Get(ctx, superActor(), uuid.New()); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("missing: want not_found, got %v", err)
	}
}

func TestCompanySoftDeleteAndRecreate(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	uc := companies.NewUsecase(c, audit.NoopRecorder{})
	prefix := "ZZ"
	co, _ := uc.Create(ctx, superActor(), companies.CreateInput{
		Name: "甲", Identifier: "co-a", CustomerCodePrefix: &prefix,
	})

	// 非 super 刪除 → permission_denied
	staff := rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DataScope: "department", Role: "staff"}
	if err := uc.Delete(ctx, staff, co.ID); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("staff delete: want permission_denied, got %v", err)
	}
	if err := uc.Delete(ctx, superActor(), co.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	// 軟刪除後列表與 Get 不可見
	list, total, err := uc.List(ctx, superActor(), companies.ListQuery{Page: 1, PerPage: 20})
	if err != nil || total != 0 || len(list) != 0 {
		t.Fatalf("deleted company should be hidden: total=%d err=%v", total, err)
	}
	if _, err := uc.Get(ctx, superActor(), co.ID); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("get deleted: want not_found, got %v", err)
	}
	// 同 identifier / prefix 可重建(部分唯一索引語意,2.1.2 步驟 5)
	if _, err := uc.Create(ctx, superActor(), companies.CreateInput{
		Name: "甲新", Identifier: "co-a", CustomerCodePrefix: &prefix,
	}); err != nil {
		t.Fatalf("recreate after soft delete: %v", err)
	}
}

func TestCompanyListScopeAndPaging(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	uc := companies.NewUsecase(c, audit.NoopRecorder{})
	coA, _ := uc.Create(ctx, superActor(), companies.CreateInput{Name: "甲", Identifier: "co-a"})
	_, _ = uc.Create(ctx, superActor(), companies.CreateInput{Name: "乙", Identifier: "co-b"})

	// super 見全部
	_, total, _ := uc.List(ctx, superActor(), companies.ListQuery{Page: 1, PerPage: 20})
	if total != 2 {
		t.Fatalf("super list: want 2, got %d", total)
	}
	// company_admin 強制過濾自己公司(忽略請求中的公司條件)
	admin := rls.Identity{UserID: uuid.New(), CompanyID: coA.ID, DataScope: "company", Role: "company_admin"}
	list, total, _ := uc.List(ctx, admin, companies.ListQuery{Page: 1, PerPage: 20, Keyword: "乙"})
	if total != 1 || list[0].ID != coA.ID {
		t.Fatalf("company_admin list should be forced to own company: %v", list)
	}
	// 其他角色不開放
	staff := rls.Identity{UserID: uuid.New(), CompanyID: coA.ID, DataScope: "department", Role: "staff"}
	if _, _, err := uc.List(ctx, staff, companies.ListQuery{Page: 1, PerPage: 20}); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("staff list: want permission_denied, got %v", err)
	}
	// per_page 超上限 → invalid_argument
	if _, _, err := uc.List(ctx, superActor(), companies.ListQuery{Page: 1, PerPage: 101}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("per_page>100: want invalid_argument, got %v", err)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/companies/ -v`
Expected: FAIL — `companies.NewUsecase` 未定義(編譯失敗)。

- [ ] **Step 3: 實作 schema 擴充、rls.Identity 擴充、proto、repository、usecase、handler**

`backend/ent/schema/company.go` — `customer_code_prefix` 改可空、新增 branding / 公開欄位、唯一索引排除 NULL(2.1.2):

```go
func (Company) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").NotEmpty(),
		field.String("tax_id").Optional().Nillable(),
		field.String("display_name").Optional(),
		field.String("identifier").NotEmpty(),
		// 2.1.2:允許 NULL(未設定前不可建立客戶編號);大寫英數 1–4 碼由 usecase 驗證
		field.String("customer_code_prefix").Optional().Nillable().MaxLen(4),
		field.Enum("status").Values("active", "suspended").Default("active"),
		field.String("primary_color").Optional().Nillable(),
		field.String("logo_url").Optional().Nillable(),
		field.String("public_email").Optional().Nillable(),
		field.String("public_phone").Optional().Nillable(),
		field.String("public_address").Optional().Nillable(),
		field.String("terms_url").Optional().Nillable(),
		field.String("privacy_url").Optional().Nillable(),
		field.JSON("public_info", map[string]any{}).Optional(),
		field.JSON("capabilities", map[string]any{}).Optional(),
		field.UUID("logo_file_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (Company) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("identifier").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("customer_code_prefix").Unique().Annotations(
			entsql.IndexWhere("deleted_at IS NULL AND customer_code_prefix IS NOT NULL")),
	}
}
```

`backend/database/migrations/00004_companies_branding.sql`(既有開發庫用;新環境由 Ent `Schema.Create` 建全量):

```sql
-- +goose Up
-- 細部文件 2.1.2 / 2.4.2:prefix 改可空、branding 與公開資訊欄位。
-- 索引名以 pg_indexes 實際為準(ent 產生預設為 company_<field>)。
ALTER TABLE companies ALTER COLUMN customer_code_prefix DROP NOT NULL;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS display_name text;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS primary_color text;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS logo_url text;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS public_email text;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS public_phone text;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS public_address text;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS terms_url text;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS privacy_url text;

DROP INDEX IF EXISTS company_customer_code_prefix;
CREATE UNIQUE INDEX company_customer_code_prefix ON companies (customer_code_prefix)
  WHERE deleted_at IS NULL AND customer_code_prefix IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS company_customer_code_prefix;
CREATE UNIQUE INDEX company_customer_code_prefix ON companies (customer_code_prefix)
  WHERE deleted_at IS NULL;
ALTER TABLE companies ALTER COLUMN customer_code_prefix SET NOT NULL;
ALTER TABLE companies DROP COLUMN IF EXISTS display_name;
ALTER TABLE companies DROP COLUMN IF EXISTS primary_color;
ALTER TABLE companies DROP COLUMN IF EXISTS logo_url;
ALTER TABLE companies DROP COLUMN IF EXISTS public_email;
ALTER TABLE companies DROP COLUMN IF EXISTS public_phone;
ALTER TABLE companies DROP COLUMN IF EXISTS public_address;
ALTER TABLE companies DROP COLUMN IF EXISTS terms_url;
ALTER TABLE companies DROP COLUMN IF EXISTS privacy_url;
```

`backend/internal/database/rls.go` 增量(Identity 加 `Roles`,供 Task 6 多角色展開;`HasRole` 供全部 domain 範圍判斷):

```go
// Identity 新增欄位:
//   Roles []string — 當前使用者於本 domain 的全部角色代碼(g 規則展開,Task 6 填入)
func (id Identity) HasRole(code string) bool {
	if id.Role == code {
		return true
	}
	for _, r := range id.Roles {
		if r == code {
			return true
		}
	}
	return false
}
```

`backend/proto/v1/common.proto`:

```proto
syntax = "proto3";
package salesorder.v1;

message PageRequest {
  int32 page = 1;      // 1-based
  int32 per_page = 2;  // 上限 100
}

message PageMeta {
  int32 total = 1;
  int32 page = 2;
  int32 per_page = 3;
}
```

`backend/proto/v1/company.proto`(Task 4 會增量加 UpdateBranding / UpdatePublicInfo):

```proto
syntax = "proto3";
package salesorder.v1;

import "google/protobuf/empty.proto";
import "google/protobuf/struct.proto";
import "google/protobuf/timestamp.proto";
import "v1/common.proto";

service CompanyService {
  rpc ListCompanies(ListCompaniesRequest) returns (ListCompaniesResponse);
  rpc GetCompany(GetCompanyRequest) returns (Company);
  rpc CreateCompany(CreateCompanyRequest) returns (Company);
  rpc UpdateCompany(UpdateCompanyRequest) returns (Company);
  rpc DeleteCompany(DeleteCompanyRequest) returns (google.protobuf.Empty);
}

message Company {
  string id = 1;
  string name = 2;
  optional string tax_id = 3;
  string display_name = 4;
  string identifier = 5;
  optional string customer_code_prefix = 6;
  string status = 7;
  optional string primary_color = 8;
  optional string logo_url = 9;
  optional string public_email = 10;
  optional string public_phone = 11;
  optional string public_address = 12;
  optional string terms_url = 13;
  optional string privacy_url = 14;
  google.protobuf.Struct public_info = 15;
  google.protobuf.Struct capabilities = 16;
  google.protobuf.Timestamp created_at = 17;
  google.protobuf.Timestamp updated_at = 18;
}

message ListCompaniesRequest {
  PageRequest page = 1;
  optional string status = 2;   // active | suspended
  optional string keyword = 3;  // name / display_name / identifier 模糊
}

message ListCompaniesResponse {
  repeated Company companies = 1;
  PageMeta meta = 2;
}

message GetCompanyRequest { string company_id = 1; }

message CreateCompanyRequest {
  string name = 1;
  optional string tax_id = 2;
  string display_name = 3;
  string identifier = 4;
  optional string customer_code_prefix = 5;
}

message UpdateCompanyRequest {
  string company_id = 1;
  optional string name = 2;
  optional string tax_id = 3;
  optional string display_name = 4;
  optional string identifier = 5; // 出現即 invalid_argument(建立後不可修改)
  optional string customer_code_prefix = 6;
  optional string status = 7; // 僅 super
}

message DeleteCompanyRequest { string company_id = 1; }
```

`backend/internal/domain/companies/usecase.go`:

```go
// Package companies 為公司主檔 domain(細部文件 2.1、2.4)。
package companies

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entcompany "github.com/salesorder/sales-order-1.0/backend/ent/company"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

var (
	// identifier:URL-safe 小寫英數與連字號,跨協定穩定代號(2.1.2 步驟 1、D20)
	identifierRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	// customer_code_prefix:大寫英數 1–4 碼(D7)
	prefixRe = regexp.MustCompile(`^[A-Z0-9]{1,4}$`)
)

type Usecase struct {
	db    *ent.Client
	audit audit.Recorder
}

func NewUsecase(db *ent.Client, rec audit.Recorder) *Usecase {
	return &Usecase{db: db, audit: rec}
}

// ListQuery 為列表查詢參數;CompanyID 僅 super 可用(company_admin 一律強制自己公司)。
type ListQuery struct {
	Page, PerPage int
	Status, Keyword, CompanyID string
}

type CreateInput struct {
	Name, TaxID, DisplayName, Identifier string
	CustomerCodePrefix                   *string
}

type UpdateInput struct {
	Name, TaxID, DisplayName *string
	Identifier             *string // 出現即 invalid_argument(2.1.1 錯誤表)
	CustomerCodePrefix     *string
	Status                 *string // 僅 super
}

func validateIdentifier(s string) error {
	if !identifierRe.MatchString(s) {
		return connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("identifier %q: must be lowercase alnum with hyphens", s))
	}
	return nil
}

func validatePrefix(p *string) error {
	if p == nil {
		return nil // 允許 NULL(2.1.2)
	}
	if !prefixRe.MatchString(*p) {
		return connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("customer_code_prefix %q: must be 1-4 uppercase alnum", *p))
	}
	return nil
}

// normalizePage 套用分頁預設與上限(per_page ≤ 100)。
func normalizePage(page, perPage int) (int, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage == 0 {
		perPage = 20
	}
	if perPage < 0 || perPage > 100 {
		return 0, 0, connect.NewError(connect.CodeInvalidArgument, errors.New("per_page must be 1-100"))
	}
	return page, perPage, nil
}

func (u *Usecase) List(ctx context.Context, actor rls.Identity, q ListQuery) ([]*ent.Company, int, error) {
	if !actor.HasRole("super") && !actor.HasRole("company_admin") {
		return nil, 0, connect.NewError(connect.CodePermissionDenied, errors.New("list not allowed"))
	}
	page, perPage, err := normalizePage(q.Page, q.PerPage)
	if err != nil {
		return nil, 0, err
	}
	query := u.db.Company.Query().Where(entcompany.DeletedAtIsNil())
	if actor.HasRole("company_admin") && !actor.HasRole("super") {
		// 2.1.1 步驟 2:強制過濾自己公司,忽略請求中的公司條件
		query = query.Where(entcompany.IDEQ(actor.CompanyID))
	} else if q.CompanyID != "" {
		cid, err := uuid.Parse(q.CompanyID)
		if err != nil {
			return nil, 0, connect.NewError(connect.CodeInvalidArgument, err)
		}
		query = query.Where(entcompany.IDEQ(cid))
	}
	if q.Status != "" {
		query = query.Where(entcompany.StatusEQ(entcompany.Status(q.Status)))
	}
	if q.Keyword != "" {
		kw := "%" + q.Keyword + "%"
		_ = kw // repository 層以 Where(Or(NameContains, DisplayNameContains, IdentifierContains)) 實作
		query = query.Where(entcompany.Or(
			entcompany.NameContains(q.Keyword),
			entcompany.DisplayNameContains(q.Keyword),
			entcompany.IdentifierContains(q.Keyword),
		))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, connect.NewError(connect.CodeInternal, err)
	}
	items, err := query.Order(ent.Asc(entcompany.FieldCreatedAt)).
		Offset((page - 1) * perPage).Limit(perPage).All(ctx)
	if err != nil {
		return nil, 0, connect.NewError(connect.CodeInternal, err)
	}
	return items, total, nil
}

func (u *Usecase) Get(ctx context.Context, actor rls.Identity, id uuid.UUID) (*ent.Company, error) {
	co, err := u.db.Company.Query().
		Where(entcompany.IDEQ(id), entcompany.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("company not found"))
	}
	if !actor.HasRole("super") {
		if !actor.HasRole("company_admin") || co.ID != actor.CompanyID {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("out of scope"))
		}
	}
	return co, nil
}

func (u *Usecase) Create(ctx context.Context, actor rls.Identity, in CreateInput) (*ent.Company, error) {
	if !actor.HasRole("super") {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("only super can create companies"))
	}
	if in.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name is required"))
	}
	if err := validateIdentifier(in.Identifier); err != nil {
		return nil, err
	}
	if err := validatePrefix(in.CustomerCodePrefix); err != nil {
		return nil, err
	}

	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 2.1.2 步驟 3:先 SELECT 給明確訊息;DB 部分唯一索引為最終保證(競態由 repository 轉譯)
	if err := ensureUnique(ctx, tx, in.Identifier, in.CustomerCodePrefix, nil); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	builder := tx.Company.Create().
		SetName(in.Name).SetDisplayName(in.DisplayName).
		SetIdentifier(in.Identifier).SetStatus("active")
	if in.TaxID != "" {
		builder.SetTaxID(in.TaxID)
	}
	if in.CustomerCodePrefix != nil {
		builder.SetCustomerCodePrefix(*in.CustomerCodePrefix)
	}
	co, err := builder.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, translateConstraint(err)
	}
	// D18:稽核同事務;DB 版 Recorder 由 03-metadicts-audit 計畫 Task 2.6 接手,呼叫點不變。
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "create",
		ResourceType: "company", ResourceID: co.ID.String(),
		After: snapshotCompany(co),
	})
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return co, nil
}

func (u *Usecase) Update(ctx context.Context, actor rls.Identity, id uuid.UUID, in UpdateInput) (*ent.Company, error) {
	co, err := u.db.Company.Query().
		Where(entcompany.IDEQ(id), entcompany.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("company not found"))
	}
	isSuper := actor.HasRole("super")
	if !isSuper {
		if !actor.HasRole("company_admin") || co.ID != actor.CompanyID {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("out of scope"))
		}
		if in.Status != nil {
			return nil, connect.NewError(connect.CodePermissionDenied,
				errors.New("company_admin cannot change status"))
		}
	}
	if in.Identifier != nil && *in.Identifier != co.Identifier {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("identifier is immutable"))
	}
	if err := validatePrefix(in.CustomerCodePrefix); err != nil {
		return nil, err
	}
	if in.Status != nil && *in.Status != "active" && *in.Status != "suspended" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("bad status"))
	}

	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if in.CustomerCodePrefix != nil {
		if err := ensureUnique(ctx, tx, co.Identifier, in.CustomerCodePrefix, &co.ID); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}
	up := tx.Company.UpdateOneID(co.ID)
	if in.Name != nil {
		up.SetName(*in.Name)
	}
	if in.TaxID != nil {
		up.SetTaxID(*in.TaxID)
	}
	if in.DisplayName != nil {
		up.SetDisplayName(*in.DisplayName)
	}
	if in.CustomerCodePrefix != nil {
		if *in.CustomerCodePrefix == "" {
			up.ClearCustomerCodePrefix()
		} else {
			up.SetCustomerCodePrefix(*in.CustomerCodePrefix)
		}
	}
	if in.Status != nil {
		// 2.1.3:status 變更本身即連鎖開關;已登入請求由 01 計畫 Task 11 middleware 逐請求阻斷,
		// 登入端點由 EnsureCompanyActive 擋(本 Task Step 4),此處不需主動刪 session。
		up.SetStatus(*in.Status)
	}
	updated, err := up.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, translateConstraint(err)
	}
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "update",
		ResourceType: "company", ResourceID: co.ID.String(),
		Before: snapshotCompany(co), After: snapshotCompany(updated),
	})
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return updated, nil
}

func (u *Usecase) Delete(ctx context.Context, actor rls.Identity, id uuid.UUID) error {
	if !actor.HasRole("super") {
		return connect.NewError(connect.CodePermissionDenied, errors.New("only super can delete companies"))
	}
	co, err := u.db.Company.Query().
		Where(entcompany.IDEQ(id), entcompany.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, errors.New("company not found"))
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if _, err := tx.Company.UpdateOneID(id).
		SetDeletedAt(timeNow()).Save(ctx); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "delete",
		ResourceType: "company", ResourceID: co.ID.String(),
		Before: snapshotCompany(co),
	})
	return commitOrErr(tx)
}

// EnsureCompanyActive 於核發任何憑證前檢查公司狀態(2.1.3 步驟 3)。
// 公司不存在 / 已軟刪除 / 非 active 一律拒絕;developer 豁免由呼叫方判斷(步驟 4)。
func EnsureCompanyActive(ctx context.Context, db *ent.Client, companyID uuid.UUID) error {
	co, err := db.Company.Query().
		Where(entcompany.IDEQ(companyID), entcompany.DeletedAtIsNil()).Only(ctx)
	if err != nil || co.Status != "active" {
		return connect.NewError(connect.CodePermissionDenied,
			errors.New("company is suspended"))
	}
	return nil
}
```

`backend/internal/domain/companies/repository.go`:

```go
package companies

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entcompany "github.com/salesorder/sales-order-1.0/backend/ent/company"
)

// timeNow 可注入(測試固定時間);預設 time.Now。
var timeNow = time.Now

// ensureUnique 唯一性預檢(2.1.2 步驟 3);excludeID 用於 Update 排除自己。
func ensureUnique(ctx context.Context, tx *ent.Tx, identifier string, prefix *string, excludeID *uuid.UUID) error {
	iq := tx.Company.Query().Where(
		entcompany.IdentifierEQ(identifier), entcompany.DeletedAtIsNil())
	if excludeID != nil {
		iq = iq.Where(entcompany.IDNEQ(*excludeID))
	}
	if exists, err := iq.Exist(ctx); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	} else if exists {
		return connect.NewError(connect.CodeAlreadyExists,
			errors.New("identifier already in use"))
	}
	if prefix != nil && *prefix != "" {
		pq := tx.Company.Query().Where(
			entcompany.CustomerCodePrefixEQ(*prefix), entcompany.DeletedAtIsNil())
		if excludeID != nil {
			pq = pq.Where(entcompany.IDNEQ(*excludeID))
		}
		if exists, err := pq.Exist(ctx); err != nil {
			return connect.NewError(connect.CodeInternal, err)
		} else if exists {
			return connect.NewError(connect.CodeAlreadyExists,
				errors.New("customer_code_prefix already in use"))
		}
	}
	return nil
}

// translateConstraint 將 DB 唯一索引違反(併發競態漏網)轉譯為 already_exists(2.1.2 步驟 3)。
func translateConstraint(err error) error {
	if ent.IsConstraintError(err) {
		return connect.NewError(connect.CodeAlreadyExists, err)
	}
	return connect.NewError(connect.CodeInternal, err)
}

// snapshotCompany 產生稽核快照(不含任何敏感欄位 — companies 無敏感欄位,全列即可)。
func snapshotCompany(co *ent.Company) string {
	b, _ := json.Marshal(co)
	return string(b)
}

func commitOrErr(tx *ent.Tx) error {
	if err := tx.Commit(); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}
```

`backend/internal/domain/companies/handler.go`(Connect handler;錯誤已由 usecase 轉為 connect.Error,直接透傳):

```go
package companies

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	v1 "github.com/salesorder/sales-order-1.0/backend/gen/proto/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

type Handler struct {
	v1.UnimplementedCompanyServiceHandler
	uc *Usecase
}

func NewHandler(uc *Usecase) *Handler { return &Handler{uc: uc} }

func identity(ctx context.Context) (rls.Identity, error) {
	id, ok := rls.FromContext(ctx)
	if !ok {
		return rls.Identity{}, connect.NewError(connect.CodeUnauthenticated, errNoIdentity)
	}
	return id, nil
}

func (h *Handler) CreateCompany(ctx context.Context, req *connect.Request[v1.CreateCompanyRequest]) (*connect.Response[v1.Company], error) {
	actor, err := identity(ctx)
	if err != nil {
		return nil, err
	}
	in := CreateInput{
		Name: req.Msg.Name, DisplayName: req.Msg.DisplayName,
		Identifier: req.Msg.Identifier,
	}
	if req.Msg.TaxId != nil {
		in.TaxID = *req.Msg.TaxId
	}
	if req.Msg.CustomerCodePrefix != nil {
		in.CustomerCodePrefix = req.Msg.CustomerCodePrefix
	}
	co, err := h.uc.Create(ctx, actor, in)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(toProto(co)), nil
}

// GetCompany / ListCompanies / UpdateCompany / DeleteCompany 同構:
// 解析 actor 與參數 → 呼叫 usecase → toProto;UpdateCompany 將 optional 欄位
// 逐一映射進 UpdateInput 指標欄位(identifier 出現即由 usecase 回 invalid_argument)。
```

(`toProto(*ent.Company) *v1.Company` 為欄位直轉,`google.protobuf.Struct` 由 `structpb.NewStruct(map)` 轉換;`UnimplementedCompanyServiceHandler` 的實際產生名稱依 buf 產碼套件為準。)

- [ ] **Step 4: 停用連鎖 — middleware developer 豁免 + 登入端點檢查(2.1.3)**

`backend/internal/middleware/auth.go` — 公司 status 檢查加 developer 豁免(2.1.3 步驟 4;既有測試 `TestAuthenticateCompanySuspended` 不變):

```go
			// 公司停用連鎖阻斷(2.1.3);developer 帳號不受阻斷(D8)
			if u.Role != "developer" {
				co, err := db.Company.Get(ctx, u.CompanyID)
				if err != nil || co.Status != "active" || co.DeletedAt != nil {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
			}
```

`backend/internal/middleware/auth_middleware_test.go` 追加:

```go
func TestAuthenticateDeveloperExemptFromCompanySuspension(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").Save(ctx)
	u, _ := c.User.Create().SetCompanyID(co.ID).SetRole("developer").SetDataScope("all").
		SetStatus("active").SetEmail("dev@example.com").Save(ctx)
	c.Company.UpdateOneID(co.ID).SetStatus("suspended").ExecX(ctx)

	h := middleware.Authenticate(c, nil, nil)(http.HandlerFunc(probe))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, bearerReq(t, u))
	if rec.Code != http.StatusOK {
		t.Fatalf("developer on suspended company: want 200, got %d", rec.Code)
	}
}
```

`backend/internal/domain/auth/customer.go` — `CustomerLogin` 於公司查得後、密碼比對前插入(2.1.3 步驟 3;公司不存在維持原「帳號或密碼錯誤」不洩漏語義,僅「公司存在但停用」回 permission_denied):

```go
	// 2.1.3:停用公司任何帳號登入 → permission_denied,不核發憑證。
	// 客戶帳號無 developer,不需豁免。
	if err := companies.EnsureCompanyActive(ctx, a.db, co.ID); err != nil {
		return nil, err
	}
```

`backend/internal/domain/auth/oauth.go` — callback 建 session 前(使用者查得後):

```go
	// 2.1.3:登入前檢查公司 status;developer 帳號豁免(步驟 4)。
	if u.Role != "developer" {
		if err := companies.EnsureCompanyActive(ctx, h.db, u.CompanyID); err != nil {
			// 導回登入頁並註明公司已停用,不建 session
			http.Redirect(w, r, "/login?error=company_suspended", http.StatusFound)
			return
		}
	}
```

測試追加(`backend/internal/domain/auth/customer_test.go`):

```go
func TestCustomerLoginSuspendedCompany(t *testing.T) {
	// 建立公司 + 客戶主帳號(流程同既有 TestCustomerLogin 成功案例),
	// 將公司 status 改為 suspended 後登入:
	// want connect.CodePermissionDenied;且不核發任何 token(handler 層不呼叫 TokenIssuer)。
}
```

- [ ] **Step 5: 產生碼並跑測試確認通過**

Run: `cd backend && go generate ./ent/ && buf generate && go mod tidy && go test ./internal/domain/companies/ ./internal/middleware/ ./internal/domain/auth/ -v`
Expected: PASS — 公司 CRUD/唯一性/範圍/軟刪除重建、developer 豁免、停用公司登入阻斷;01 計畫既有測試全綠(注意:01 測試建公司均帶 prefix,schema 改可空後行為相容)。

- [ ] **Step 6: Commit**

```bash
git add backend/ent/schema/company.go backend/database/migrations/00004_companies_branding.sql backend/proto/v1/common.proto backend/proto/v1/company.proto backend/internal/domain/companies backend/internal/database/rls.go backend/internal/middleware backend/internal/domain/auth
git commit -m "feat(backend): CompanyService CRUD、identifier/prefix 唯一性、公司停用連鎖阻斷(2.1)"
```

---

### Task 2: 部門管理 API(細部 2.2.1)

**Files:**
- Create: `backend/proto/v1/department.proto`
- Create: `backend/internal/domain/departments/usecase.go`
- Create: `backend/internal/domain/departments/repository.go`
- Create: `backend/internal/domain/departments/handler.go`
- Test: `backend/internal/domain/departments/usecase_test.go`

**Interfaces:**
- Consumes: 01 計畫 Task 1 `ent.Department`;Task 1 `rls.Identity.HasRole`、`audit.Recorder`、`normalizePage` 模式。
- Produces: proto `DepartmentService` — `ListDepartments(page、company_id、include_deleted)→ ListDepartmentsResponse{departments, meta}`;`GetDepartment(department_id) → Department`;`CreateDepartment(company_id、name)→ Department`;`UpdateDepartment(department_id、name、status)→ Department`;`DeleteDepartment(department_id)→ Empty`。usecase `departments.Usecase` 同名方法;`departments.NewUsecase(db *ent.Client, rec audit.Recorder)`。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/domain/departments/usecase_test.go`:

```go
package departments_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/departments"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func mkCompany(t *testing.T, c *ent.Client, ident string) *ent.Company {
	t.Helper()
	co, err := c.Company.Create().SetName("公司" + ident).SetIdentifier(ident).Save(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return co
}

func superActor() rls.Identity {
	return rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(), DataScope: "all", Role: "super"}
}

func TestDepartmentCRUDAndScope(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coA, coB := mkCompany(t, c, "co-a"), mkCompany(t, c, "co-b")
	uc := departments.NewUsecase(c, audit.NoopRecorder{})

	// super 對任意公司建立
	dep, err := uc.Create(ctx, superActor(), departments.CreateInput{CompanyID: coA.ID, Name: "業務部"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// company_admin 對自己公司建立 OK;對他公司 → permission_denied
	adminA := rls.Identity{UserID: uuid.New(), CompanyID: coA.ID, DataScope: "company", Role: "company_admin"}
	if _, err := uc.Create(ctx, adminA, departments.CreateInput{CompanyID: coA.ID, Name: "倉儲部"}); err != nil {
		t.Fatalf("admin create own: %v", err)
	}
	if _, err := uc.Create(ctx, adminA, departments.CreateInput{CompanyID: coB.ID, Name: "越權"}); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("admin create other company: want permission_denied, got %v", err)
	}
	// Get/Update/Delete 他公司部門 → permission_denied
	if _, err := uc.Get(ctx, adminA, dep.ID); err != nil { // dep 屬 coA,同公司可讀
		t.Fatalf("admin get own dept: %v", err)
	}
	depB, _ := uc.Create(ctx, superActor(), departments.CreateInput{CompanyID: coB.ID, Name: "乙公司部門"})
	if _, err := uc.Get(ctx, adminA, depB.ID); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("admin get other dept: want permission_denied, got %v", err)
	}
	newName := "改名"
	if _, err := uc.Update(ctx, adminA, depB.ID, departments.UpdateInput{Name: &newName}); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("admin update other dept: want permission_denied, got %v", err)
	}
	if err := uc.Delete(ctx, adminA, depB.ID); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("admin delete other dept: want permission_denied, got %v", err)
	}
	// staff 呼叫寫入 RPC → permission_denied
	staff := rls.Identity{UserID: uuid.New(), CompanyID: coA.ID, DataScope: "department", Role: "staff"}
	if _, err := uc.Create(ctx, staff, departments.CreateInput{CompanyID: coA.ID, Name: "x"}); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("staff create: want permission_denied, got %v", err)
	}
	// 缺 name → invalid_argument;不存在 → not_found
	if _, err := uc.Create(ctx, superActor(), departments.CreateInput{CompanyID: coA.ID}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("empty name: want invalid_argument, got %v", err)
	}
	if _, err := uc.Get(ctx, superActor(), uuid.New()); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("missing: want not_found, got %v", err)
	}
}

func TestDepartmentDeleteBlockedByUsersAndSoftDelete(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	coA := mkCompany(t, c, "co-a")
	uc := departments.NewUsecase(c, audit.NoopRecorder{})
	dep, _ := uc.Create(ctx, superActor(), departments.CreateInput{CompanyID: coA.ID, Name: "業務部"})

	// 部門內有未刪除使用者 → 禁止刪除(failed_precondition)
	if _, err := c.User.Create().SetCompanyID(coA.ID).SetDepartmentID(dep.ID).
		SetRole("staff").SetDataScope("department").SetStatus("active").Save(ctx); err != nil {
		t.Fatal(err)
	}
	if err := uc.Delete(ctx, superActor(), dep.ID); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("delete with users: want failed_precondition, got %v", err)
	}
	// 移除使用者後可刪;刪後列表不可見、同 (company,name) 可重建
	c.User.Update().Where(/* 該使用者 */).SetDeletedAt(time.Now()).ExecX(ctx)
	if err := uc.Delete(ctx, superActor(), dep.ID); err != nil {
		t.Fatalf("delete after users removed: %v", err)
	}
	list, total, _ := uc.List(ctx, superActor(), departments.ListQuery{CompanyID: coA.ID.String(), Page: 1, PerPage: 20})
	if total != 0 || len(list) != 0 {
		t.Fatalf("deleted dept should be hidden: %d", total)
	}
	if _, err := uc.Create(ctx, superActor(), departments.CreateInput{CompanyID: coA.ID, Name: "業務部"}); err != nil {
		t.Fatalf("recreate same name after soft delete: %v", err)
	}
}
```

(測試中 `c.User.Update().Where(...)` 處以實際撈出的使用者 ID 補齊 predicate;`time` import 依此補上。)

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/departments/ -v`
Expected: FAIL — `departments.NewUsecase` 未定義。

- [ ] **Step 3: 實作**

`backend/proto/v1/department.proto`:

```proto
syntax = "proto3";
package salesorder.v1;

import "google/protobuf/empty.proto";
import "v1/common.proto";

service DepartmentService {
  rpc ListDepartments(ListDepartmentsRequest) returns (ListDepartmentsResponse);
  rpc GetDepartment(GetDepartmentRequest) returns (Department);
  rpc CreateDepartment(CreateDepartmentRequest) returns (Department);
  rpc UpdateDepartment(UpdateDepartmentRequest) returns (Department);
  rpc DeleteDepartment(DeleteDepartmentRequest) returns (google.protobuf.Empty);
}

message Department {
  string id = 1;
  string company_id = 2;
  string name = 3;
  string status = 4;
}

message ListDepartmentsRequest {
  PageRequest page = 1;
  optional string company_id = 2;   // 僅 super 可用;company_admin 一律強制自己公司
  bool include_deleted = 3;          // 管理介面「顯示已刪除」(2.2.1 步驟 5)
}

message ListDepartmentsResponse {
  repeated Department departments = 1;
  PageMeta meta = 2;
}

message GetDepartmentRequest { string department_id = 1; }
message CreateDepartmentRequest { string company_id = 1; string name = 2; }

message UpdateDepartmentRequest {
  string department_id = 1;
  optional string name = 2;
  optional string status = 3; // active | suspended
}

message DeleteDepartmentRequest { string department_id = 1; }
```

`backend/internal/domain/departments/usecase.go`:

```go
// Package departments 為部門主檔 domain(細部文件 2.2.1)。
package departments

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entdept "github.com/salesorder/sales-order-1.0/backend/ent/department"
	entuser "github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

type Usecase struct {
	db    *ent.Client
	audit audit.Recorder
}

func NewUsecase(db *ent.Client, rec audit.Recorder) *Usecase {
	return &Usecase{db: db, audit: rec}
}

type ListQuery struct {
	Page, PerPage  int
	CompanyID      string
	IncludeDeleted bool
}

type CreateInput struct {
	CompanyID uuid.UUID
	Name      string
}

type UpdateInput struct {
	Name, Status *string
}

// checkCompanyScope:super 全域;company_admin 限自己公司;其餘角色拒絕(2.2.1 步驟 1)。
func checkCompanyScope(actor rls.Identity, companyID uuid.UUID) error {
	if actor.HasRole("super") {
		return nil
	}
	if actor.HasRole("company_admin") && actor.CompanyID == companyID {
		return nil
	}
	return connect.NewError(connect.CodePermissionDenied, errors.New("out of scope"))
}

func (u *Usecase) List(ctx context.Context, actor rls.Identity, q ListQuery) ([]*ent.Department, int, error) {
	if !actor.HasRole("super") && !actor.HasRole("company_admin") {
		return nil, 0, connect.NewError(connect.CodePermissionDenied, errors.New("list not allowed"))
	}
	page, perPage, err := normalizePage(q.Page, q.PerPage) // 複用 Task 1 companies 的同款實作,抽至 internal/database/paging.go
	if err != nil {
		return nil, 0, err
	}
	query := u.db.Department.Query()
	if !q.IncludeDeleted {
		query = query.Where(entdept.DeletedAtIsNil())
	}
	if actor.HasRole("company_admin") && !actor.HasRole("super") {
		query = query.Where(entdept.CompanyIDEQ(actor.CompanyID)) // 強制過濾
	} else if q.CompanyID != "" {
		cid, err := uuid.Parse(q.CompanyID)
		if err != nil {
			return nil, 0, connect.NewError(connect.CodeInvalidArgument, err)
		}
		query = query.Where(entdept.CompanyIDEQ(cid))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, connect.NewError(connect.CodeInternal, err)
	}
	items, err := query.Order(ent.Asc(entdept.FieldCreatedAt)).
		Offset((page - 1) * perPage).Limit(perPage).All(ctx)
	if err != nil {
		return nil, 0, connect.NewError(connect.CodeInternal, err)
	}
	return items, total, nil
}

func (u *Usecase) Get(ctx context.Context, actor rls.Identity, id uuid.UUID) (*ent.Department, error) {
	dep, err := u.db.Department.Query().
		Where(entdept.IDEQ(id), entdept.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("department not found"))
	}
	if err := checkCompanyScope(actor, dep.CompanyID); err != nil {
		return nil, err
	}
	return dep, nil
}

func (u *Usecase) Create(ctx context.Context, actor rls.Identity, in CreateInput) (*ent.Department, error) {
	if in.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name is required"))
	}
	if err := checkCompanyScope(actor, in.CompanyID); err != nil {
		return nil, err
	}
	// 目標公司須存在且未刪除
	if exists, err := u.db.Company.Query().Where(
		entcompany.IDEQ(in.CompanyID), entcompany.DeletedAtIsNil()).Exist(ctx); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	} else if !exists {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("company not found"))
	}

	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	dep, err := tx.Department.Create().SetCompanyID(in.CompanyID).SetName(in.Name).Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		if ent.IsConstraintError(err) { // 同公司同名(部分唯一索引)
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "create",
		ResourceType: "department", ResourceID: dep.ID.String(),
		After: snapshot(dep),
	})
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return dep, nil
}

func (u *Usecase) Update(ctx context.Context, actor rls.Identity, id uuid.UUID, in UpdateInput) (*ent.Department, error) {
	dep, err := u.Get(ctx, actor, id) // 內含 not_found + 範圍判斷
	if err != nil {
		return nil, err
	}
	if in.Status != nil && *in.Status != "active" && *in.Status != "suspended" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("bad status"))
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	up := tx.Department.UpdateOneID(id)
	if in.Name != nil {
		up.SetName(*in.Name)
	}
	if in.Status != nil {
		up.SetStatus(*in.Status)
	}
	updated, err := up.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		if ent.IsConstraintError(err) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "update",
		ResourceType: "department", ResourceID: dep.ID.String(),
		Before: snapshot(dep), After: snapshot(updated),
	})
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return updated, nil
}

func (u *Usecase) Delete(ctx context.Context, actor rls.Identity, id uuid.UUID) error {
	dep, err := u.Get(ctx, actor, id) // 內含 not_found + 範圍判斷
	if err != nil {
		return err
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	// 仍有使用者的部門禁止刪除(未刪除使用者即阻擋,含 suspended,保守處理)
	inUse, err := tx.User.Query().Where(
		entuser.DepartmentIDEQ(dep.ID), entuser.DeletedAtIsNil()).Exist(ctx)
	if err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	if inUse {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("department still has users; reassign them first"))
	}
	if _, err := tx.Department.UpdateOneID(dep.ID).
		SetDeletedAt(time.Now()).Save(ctx); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "delete",
		ResourceType: "department", ResourceID: dep.ID.String(),
		Before: snapshot(dep),
	})
	return tx.Commit()
}
```

(`snapshot` 與 Task 1 同款 JSON 序列化;`normalizePage` 於本 Task 自 `companies` 抽出至 `backend/internal/database/paging.go` 並回改 Task 1 呼叫點 — 同一 PR 內重構,避免重複實作;`entcompany` import 補上。)

handler 同 Task 1 模式:解析 actor → 映射 proto optional 欄位 → 呼叫 usecase。

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && buf generate && go test ./internal/domain/departments/ ./internal/domain/companies/ -v`
Expected: PASS — 範圍、軟刪除、有使用者阻擋刪除、同名重建;companies 測試不因 normalizePage 重構回歸。

- [ ] **Step 5: Commit**

```bash
git add backend/proto/v1/department.proto backend/internal/domain/departments backend/internal/database/paging.go backend/internal/domain/companies
git commit -m "feat(backend): DepartmentService CRUD 與公司範圍控制(2.2)"
```

---

### Task 3: 使用者 CRUD + 角色指派 + 停用 + ForceLogout 範圍銜接(細部 2.3.1–2.3.3)

**Files:**
- Create: `backend/internal/database/txsql.go`
- Update: `backend/internal/domain/users/usecase.go`(01 計畫 Task 12 檔案,擴充)
- Create: `backend/internal/domain/users/repository.go`
- Create: `backend/internal/domain/users/handler.go`
- Update: `backend/proto/v1/user.proto`(增量;`ForceLogout` 已存在,不重複宣告)
- Test: `backend/internal/domain/users/usecase_test.go`(更新)+ `backend/internal/domain/users/assignrole_test.go`(新建)

**Interfaces:**
- Consumes: 01 計畫 Task 12 `users.Usecase.ForceLogout` / `forceLogoutUser`(`NewUsecase(db, sm)` 簽名本 Task 擴充,舊測試同步更新 — clean cutover);Task 9/12 `session.Manager.DeleteSessionByUserID`;Task 2 `authz.NewEnforcer`;Task 5(本計畫)`roles` 表 — 但 AssignRole 驗證僅查 `roles` 表,Task 5 schema 先行或本 Task 以 raw SQL 查 `roles` 表皆可;本計畫 Task 順序上 roles 表於 Task 5 建,**因此本 Task 的 role_code 驗證先以內建角色列舉 + `roles` 表存在時查表雙軌:實作為「查 `roles` 表,查無表(尚未 migrate)時回落內建列舉」— 測試以 Task 5 migration 套用後的庫為準,執行順序上本 Task 測試需 roles 表,故實際執行時 Task 5 的 schema/migration 先落地或本 Task 測試暫以內建列舉路徑驗證並標 TODO(Task 5 完成後移除回落)**。
- Produces: proto `UserService` 增量 — `ListUsers(page、company_id、department_id、role、status)→ ListUsersResponse{users, meta}`;`GetUser(user_id)→ User`;`CreateUser(name、company_id、department_id、role)→ User`;`UpdateUser(user_id、name、department_id、phone、employee_no、status)→ User`;`AssignRole(user_id、role_code、department_id?)→ User`;`Deactivate(user_id)→ Empty`。usecase `Usecase.{ListUsers,GetUser,CreateUser,UpdateUser,AssignRole,Deactivate}`;`NewUsecase(db *ent.Client, sm *session.Manager, enf *casbin.Enforcer, rec audit.Recorder)`;範圍函式 `checkManageScope(actor, target) error`(2.3.2 步驟 1,供 AssignRole / Deactivate / ForceLogout / UpdateUser 共用);`txsql.ExecTx / QueryTx`(ent.Tx 內 raw SQL,操作 `casbin_rule`)。

- [ ] **Step 1: txsql 輔助 + 寫失敗測試(範圍、審核、角色異動同事務)**

`backend/internal/database/txsql.go`:

```go
// Package database — txsql 提供 ent.Tx 內 raw SQL 能力,用於 casbin_rule 等非 Ent 管理表
// (gorm-adapter 不支援外部交易,D18 要求 g 規則異動與業務異動同 DB 交易)。
package database

import (
	"context"
	"database/sql"
	"fmt"

	entsql "entgo.io/ent/dialect/sql"

	"github.com/salesorder/sales-order-1.0/backend/ent"
)

// ExecTx 在交易內執行寫入 SQL。
func ExecTx(ctx context.Context, tx *ent.Tx, query string, args ...any) error {
	stx, ok := tx.(*entsql.Tx)
	if !ok {
		return fmt.Errorf("txsql: unexpected tx type %T", tx)
	}
	_, err := stx.ExecContext(ctx, query, args...)
	return err
}

// QueryTx 在交易內執行查詢 SQL;呼叫方負責 Close。
func QueryTx(ctx context.Context, tx *ent.Tx, query string, args ...any) (*sql.Rows, error) {
	stx, ok := tx.(*entsql.Tx)
	if !ok {
		return nil, fmt.Errorf("txsql: unexpected tx type %T", tx)
	}
	return stx.QueryContext(ctx, query, args...)
}
```

`backend/internal/domain/users/assignrole_test.go`:

```go
package users_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/users"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

// captureRecorder 記錄稽核呼叫(Noop 之外的行為驗證用)。
type captureRecorder struct{ entries []audit.Entry }

func (r *captureRecorder) Record(_ context.Context, e audit.Entry) error {
	r.entries = append(r.entries, e)
	return nil
}

func seedUser(t *testing.T, c *ent.Client, coID uuid.UUID, deptID *uuid.UUID, role, status string) *ent.User {
	t.Helper()
	b := c.User.Create().SetCompanyID(coID).SetRole(role).SetStatus(status).
		SetDataScope("department").SetName("u-" + role)
	if deptID != nil {
		b.SetDepartmentID(*deptID)
	}
	u, err := b.Save(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func TestAssignRoleGuestApproval(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").Save(ctx)
	dep, _ := c.Department.Create().SetCompanyID(co.ID).SetName("業務部").Save(ctx)
	guest := seedUser(t, c, co.ID, nil, "guest", "pending")

	enf, err := authz.NewEnforcer(testutil.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	rec := &captureRecorder{}
	uc := users.NewUsecase(c, nil, enf, rec)

	// dept_admin 審核 guest → permission_denied(2.3.1 步驟 4b)
	deptAdmin := rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DepartmentID: &dep.ID,
		DataScope: "department", Role: "dept_admin"}
	if _, err := uc.AssignRole(ctx, deptAdmin, guest.ID, "staff", &dep.ID); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("dept_admin approve: want permission_denied, got %v", err)
	}
	// company_admin 審核:pending → active,指派部門與角色
	admin := rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DataScope: "company", Role: "company_admin"}
	updated, err := uc.AssignRole(ctx, admin, guest.ID, "staff", &dep.ID)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if updated.Status != "active" || updated.DepartmentID == nil || *updated.DepartmentID != dep.ID {
		t.Fatalf("approval should activate + assign dept: %+v", updated)
	}
	if updated.TokenVersion != guest.TokenVersion+1 {
		t.Fatal("token_version should bump on role change (D5)")
	}
	// g 規則更新:重載後可見 staff 角色於公司 domain
	if err := enf.LoadPolicy(); err != nil {
		t.Fatal(err)
	}
	roles := enf.GetRolesForUser(guest.ID.String(), co.ID.String())
	if len(roles) != 1 || roles[0] != "staff" {
		t.Fatalf("g rules: want [staff], got %v", roles)
	}
	// 稽核 role_change 已記錄
	found := false
	for _, e := range rec.entries {
		if e.Action == "role_change" && e.ResourceID == guest.ID.String() {
			found = true
		}
	}
	if !found {
		t.Fatal("role_change audit entry missing")
	}
	// 對非 pending 帳號走審核路徑(帶 department_id 且目標 active)→ failed_precondition
	if _, err := uc.AssignRole(ctx, admin, guest.ID, "staff", &dep.ID); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("approve non-pending: want failed_precondition, got %v", err)
	}
}

func TestAssignRoleScopeAndValidation(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").Save(ctx)
	dep, _ := c.Department.Create().SetCompanyID(co.ID).SetName("業務部").Save(ctx)
	staff := seedUser(t, c, co.ID, &dep.ID, "staff", "active")
	otherAdmin := seedUser(t, c, co.ID, nil, "company_admin", "active")

	enf, _ := authz.NewEnforcer(testutil.DSN(t))
	uc := users.NewUsecase(c, nil, enf, audit.NoopRecorder{})
	deptAdmin := rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DepartmentID: &dep.ID,
		DataScope: "department", Role: "dept_admin"}

	// dept_admin 可改派自己部門 staff
	if _, err := uc.AssignRole(ctx, deptAdmin, staff.ID, "staff", nil); err != nil {
		t.Fatalf("dept_admin assign staff: %v", err)
	}
	// dept_admin 對非 staff 帳號 → permission_denied(2.3.2 步驟 1c)
	if _, err := uc.AssignRole(ctx, deptAdmin, otherAdmin.ID, "staff", nil); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("dept_admin on admin: want permission_denied, got %v", err)
	}
	// 不存在角色 → invalid_argument;不存在使用者 → not_found
	admin := rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DataScope: "all", Role: "super"}
	if _, err := uc.AssignRole(ctx, admin, staff.ID, "no-such-role", nil); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("bad role: want invalid_argument, got %v", err)
	}
	if _, err := uc.AssignRole(ctx, admin, uuid.New(), "staff", nil); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("missing user: want not_found, got %v", err)
	}
}

func TestDeactivate(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").Save(ctx)
	target := seedUser(t, c, co.ID, nil, "staff", "active")

	uc := users.NewUsecase(c, nil, nil, audit.NoopRecorder{})
	admin := rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DataScope: "company", Role: "company_admin"}
	if err := uc.Deactivate(ctx, admin, target.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	after := c.User.GetX(ctx, target.ID)
	if after.Status != "suspended" {
		t.Fatalf("status: want suspended, got %q", after.Status)
	}
	if after.TokenVersion != target.TokenVersion+1 {
		t.Fatal("token_version should bump (舊 token 立即失效,D5)")
	}
	// 重複停用 → failed_precondition
	if err := uc.Deactivate(ctx, admin, target.ID); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("double deactivate: want failed_precondition, got %v", err)
	}
	// 重新啟用走 UpdateUser status=active(2.3.2 步驟 4)
	active := "active"
	back, err := uc.UpdateUser(ctx, admin, target.ID, users.UpdateInput{Status: &active})
	if err != nil || back.Status != "active" {
		t.Fatalf("reactivate: %v", err)
	}
}

// TestForceLogoutScopeShared 驗證 2.3.3:ForceLogout 改用與其他 RPC 相同的 checkManageScope。
func TestForceLogoutScopeShared(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").Save(ctx)
	depA, _ := c.Department.Create().SetCompanyID(co.ID).SetName("A部").Save(ctx)
	depB, _ := c.Department.Create().SetCompanyID(co.ID).SetName("B部").Save(ctx)
	targetB := seedUser(t, c, co.ID, &depB.ID, "staff", "active")

	uc := users.NewUsecase(c, nil, nil, audit.NoopRecorder{})
	deptAdminA := rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DepartmentID: &depA.ID,
		DataScope: "department", Role: "dept_admin"}
	// dept_admin 強登他部門 → permission_denied(2.3.3 與 1.7.2 同一實作)
	if err := uc.ForceLogout(ctx, deptAdminA, targetB.ID); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("cross-dept force logout: want permission_denied, got %v", err)
	}
	// 同部門 staff → 成功
	targetA := seedUser(t, c, co.ID, &depA.ID, "staff", "active")
	if err := uc.ForceLogout(ctx, deptAdminA, targetA.ID); err != nil {
		t.Fatalf("same-dept force logout: %v", err)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/users/ -v`
Expected: FAIL — `users.NewUsecase` 簽名不符 / `AssignRole` / `Deactivate` 未定義(編譯失敗)。

- [ ] **Step 3: 實作 usecase 擴充與 handler**

`backend/internal/domain/users/usecase.go`(在 01 計畫 Task 12 檔案上擴充;`ForceLogout` 的內聯範圍判斷改為 `checkManageScope`,2.3.3 銜接點):

```go
package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entuser "github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/session"
)

type Usecase struct {
	db    *ent.Client
	sm    *session.Manager // 可為 nil;非 nil 時提交後刪 Valkey session
	enf   *casbin.Enforcer // 可為 nil(測試不需要 g 規則驗證時);AssignRole 必填
	audit audit.Recorder
}

// NewUsecase 簽名自 01 計畫 Task 12 擴充(原 NewUsecase(db, sm));Task 12 測試呼叫點同步更新。
func NewUsecase(db *ent.Client, sm *session.Manager, enf *casbin.Enforcer, rec audit.Recorder) *Usecase {
	return &Usecase{db: db, sm: sm, enf: enf, audit: rec}
}

// builtinRoles 為 users.role 列舉的內建角色(01 計畫 Task 1 schema)。
var builtinRoles = map[string]bool{
	"super": true, "company_admin": true, "dept_admin": true,
	"staff": true, "guest": true, "customer": true, "developer": true,
}

// checkManageScope 為 UserService 全部寫入 RPC 的統一範圍判斷(2.3.2 步驟 1;2.3.3 同一實作)。
func checkManageScope(actor rls.Identity, target *ent.User) error {
	if actor.HasRole("super") || actor.HasRole("developer") {
		return nil
	}
	if actor.HasRole("company_admin") {
		if target.CompanyID == actor.CompanyID {
			return nil
		}
		return connect.NewError(connect.CodePermissionDenied, errors.New("out of company scope"))
	}
	if actor.HasRole("dept_admin") {
		if target.CompanyID == actor.CompanyID &&
			actor.DepartmentID != nil && target.DepartmentID != nil &&
			*target.DepartmentID == *actor.DepartmentID &&
			target.Role == "staff" { // 規格 3.2:dept_admin 僅管理部門內 staff
			return nil
		}
		return connect.NewError(connect.CodePermissionDenied, errors.New("out of department scope"))
	}
	return connect.NewError(connect.CodePermissionDenied, errors.New("role cannot manage users"))
}

func (u *Usecase) getTarget(ctx context.Context, id uuid.UUID) (*ent.User, error) {
	target, err := u.db.User.Get(ctx, id)
	if err != nil || target.DeletedAt != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	}
	return target, nil
}

// roleExists 驗證 role_code:優先查 roles 表(Task 5);表不存在(尚未 migrate)回落內建列舉。
// TODO(Task 5 完成後):移除回落路徑,一律查表(含 is_active = true 條件)。
func (u *Usecase) roleExists(ctx context.Context, code string) bool {
	// roles 表查詢以 raw SQL 實作,避免本 Task 依賴 Task 5 的 Ent 產碼;
	// 表不存在(42P01)時回落內建列舉。
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return builtinRoles[code]
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := database.QueryTx(ctx, tx,
		`SELECT 1 FROM roles WHERE code = $1 AND is_active = true AND deleted_at IS NULL LIMIT 1`, code)
	if err != nil {
		return builtinRoles[code] // undefined_table 等 → 回落
	}
	defer rows.Close()
	if rows.Next() {
		return true
	}
	return false
}

// AssignRole 角色指派與 guest 審核(2.3.1 步驟 4)。
// 交易內三件事:更新使用者、替換 g 規則、token_version + 1;enforcer 記憶體於提交後重載。
func (u *Usecase) AssignRole(ctx context.Context, actor rls.Identity, targetID uuid.UUID, roleCode string, departmentID *uuid.UUID) (*ent.User, error) {
	target, err := u.getTarget(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if !u.roleExists(ctx, roleCode) {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("role %q not found or inactive", roleCode))
	}
	if target.Status == "pending" {
		// guest 審核路徑:僅 super / company_admin(2.3.1 步驟 4b),須指派部門
		if !actor.HasRole("super") && !actor.HasRole("company_admin") {
			return nil, connect.NewError(connect.CodePermissionDenied,
				errors.New("guest approval requires super or company_admin"))
		}
		if err := checkManageScope(actor, target); err != nil {
			return nil, err
		}
		if departmentID == nil {
			return nil, connect.NewError(connect.CodeInvalidArgument,
				errors.New("department_id required for guest approval"))
		}
	} else {
		if departmentID != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition,
				errors.New("approval flow only valid for pending users"))
		}
		if err := checkManageScope(actor, target); err != nil {
			return nil, err
		}
	}

	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	before := snapshotUser(target)
	up := tx.User.UpdateOneID(targetID).
		SetTokenVersion(target.TokenVersion + 1) // D5:角色異動舊 token 立即失效
	if builtinRoles[roleCode] {
		up.SetRole(roleCode) // 內建角色同步主角色快照;自訂角色僅存 g 規則
	}
	if target.Status == "pending" {
		up.SetStatus("active").SetDepartmentID(*departmentID)
	}
	updated, err := up.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// g 規則替換(同 DB 交易,D18):移除該使用者於公司 domain 的全部舊角色,加入新角色。
	domain := target.CompanyID.String()
	if err := database.ExecTx(ctx, tx,
		`DELETE FROM casbin_rule WHERE ptype = 'g' AND v0 = $1 AND v2 = $2`,
		targetID.String(), domain); err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := database.ExecTx(ctx, tx,
		`INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('g', $1, $2, $3)`,
		targetID.String(), roleCode, domain); err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "role_change",
		ResourceType: "user", ResourceID: targetID.String(),
		Before: before, After: snapshotUser(updated),
	})
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if u.enf != nil {
		_ = u.enf.LoadPolicy() // 提交失敗不更新記憶體(2.3.1 步驟 5)
	}
	return updated, nil
}

// Deactivate 停用帳號(2.3.2):status=suspended + tv+1 + 刪 Web session + 稽核,同一交易。
func (u *Usecase) Deactivate(ctx context.Context, actor rls.Identity, targetID uuid.UUID) error {
	target, err := u.getTarget(ctx, targetID)
	if err != nil {
		return err
	}
	if err := checkManageScope(actor, target); err != nil {
		return err
	}
	if target.Status == "suspended" {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("already suspended"))
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if _, err := tx.User.UpdateOneID(targetID).
		SetStatus("suspended").SetTokenVersion(target.TokenVersion + 1).Save(ctx); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "update",
		ResourceType: "user", ResourceID: targetID.String(),
		Before: snapshotUser(target),
		After:  fmt.Sprintf(`{"status":"suspended","token_version":%d}`, target.TokenVersion+1),
	})
	if err := tx.Commit(); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if u.sm != nil {
		_ = u.sm.DeleteSessionByUserID(ctx, targetID) // 提交後動作;tv 已 +1,失敗不影響撤銷語義
	}
	return nil
}

// ForceLogout(01 計畫 Task 12 既有)— 僅替換範圍判斷(2.3.3):
// 原內聯 `actor.DataScope != "all" && target.CompanyID != actor.CompanyID` 改為:
//
//	if err := checkManageScope(actor, target); err != nil {
//		return err
//	}
//
// 其餘邏輯(self 檢查、forceLogoutUser 交易核心、提交後刪 session)不變。
```

`ListUsers` / `GetUser` / `CreateUser` / `UpdateUser`(同檔):

```go
type ListQuery struct {
	Page, PerPage                    int
	CompanyID, DepartmentID          string
	Role, Status                     string
}

type CreateInput struct {
	Name         string
	CompanyID    uuid.UUID
	DepartmentID *uuid.UUID
	Role         string // 內建列舉;客戶帳號不由本 RPC 建立(Phase 3,D22/D28)
}

type UpdateInput struct {
	Name, Phone, EmployeeNo *string
	DepartmentID            *uuid.UUID
	Status                  *string // active | suspended(重新啟用路徑,2.3.2 步驟 4)
}

func (u *Usecase) ListUsers(ctx context.Context, actor rls.Identity, q ListQuery) ([]*ent.User, int, error) {
	if !actor.HasRole("super") && !actor.HasRole("company_admin") && !actor.HasRole("dept_admin") {
		return nil, 0, connect.NewError(connect.CodePermissionDenied, errors.New("list not allowed"))
	}
	page, perPage, err := normalizePage(q.Page, q.PerPage)
	if err != nil {
		return nil, 0, err
	}
	query := u.db.User.Query().Where(entuser.DeletedAtIsNil())
	// 2.3.1 步驟 2:依操作者範圍強制注入過濾;RLS 兜底
	switch {
	case actor.HasRole("super"):
		if q.CompanyID != "" {
			cid, err := uuid.Parse(q.CompanyID)
			if err != nil {
				return nil, 0, connect.NewError(connect.CodeInvalidArgument, err)
			}
			query = query.Where(entuser.CompanyIDEQ(cid))
		}
	case actor.HasRole("company_admin"):
		query = query.Where(entuser.CompanyIDEQ(actor.CompanyID))
	default: // dept_admin:限自己部門 staff
		query = query.Where(
			entuser.CompanyIDEQ(actor.CompanyID),
			entuser.DepartmentIDEQ(*actor.DepartmentID),
			entuser.RoleEQ("staff"),
		)
	}
	if q.DepartmentID != "" {
		did, err := uuid.Parse(q.DepartmentID)
		if err != nil {
			return nil, 0, connect.NewError(connect.CodeInvalidArgument, err)
		}
		query = query.Where(entuser.DepartmentIDEQ(did))
	}
	if q.Role != "" {
		query = query.Where(entuser.RoleEQ(q.Role))
	}
	if q.Status != "" {
		query = query.Where(entuser.StatusEQ(entuser.Status(q.Status)))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, connect.NewError(connect.CodeInternal, err)
	}
	items, err := query.Order(ent.Asc(entuser.FieldCreatedAt)).
		Offset((page - 1) * perPage).Limit(perPage).All(ctx)
	if err != nil {
		return nil, 0, connect.NewError(connect.CodeInternal, err)
	}
	return items, total, nil
}

func (u *Usecase) GetUser(ctx context.Context, actor rls.Identity, id uuid.UUID) (*ent.User, error) {
	target, err := u.getTarget(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := checkManageScope(actor, target); err != nil {
		return nil, err
	}
	return target, nil // 回傳不含 password_hash(handler toProto 不映射,欄位 Sensitive)
}

func (u *Usecase) CreateUser(ctx context.Context, actor rls.Identity, in CreateInput) (*ent.User, error) {
	if in.Name == "" || !builtinRoles[in.Role] || in.Role == "customer" || in.Role == "guest" {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("name/role invalid; customer accounts are created by Phase 3 flow"))
	}
	// 範圍:super 全域;company_admin 限自己公司;dept_admin 不可建立
	if !actor.HasRole("super") {
		if !actor.HasRole("company_admin") || actor.CompanyID != in.CompanyID {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("out of scope"))
		}
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 員工帳號走 OAuth2,不產生密碼(password_hash 為 NULL,2.3.1 步驟 3)
	builder := tx.User.Create().SetCompanyID(in.CompanyID).SetName(in.Name).
		SetRole(in.Role).SetStatus("active").SetDataScope("department")
	if in.DepartmentID != nil {
		builder.SetDepartmentID(*in.DepartmentID)
	}
	created, err := builder.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := database.ExecTx(ctx, tx,
		`INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('g', $1, $2, $3)`,
		created.ID.String(), in.Role, in.CompanyID.String()); err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "create",
		ResourceType: "user", ResourceID: created.ID.String(),
		After: snapshotUser(created),
	})
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if u.enf != nil {
		_ = u.enf.LoadPolicy()
	}
	return created, nil
}

func (u *Usecase) UpdateUser(ctx context.Context, actor rls.Identity, id uuid.UUID, in UpdateInput) (*ent.User, error) {
	target, err := u.getTarget(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := checkManageScope(actor, target); err != nil {
		return nil, err
	}
	if in.Status != nil && *in.Status != "active" && *in.Status != "suspended" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("bad status"))
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	up := tx.User.UpdateOneID(id)
	if in.Name != nil {
		up.SetName(*in.Name)
	}
	if in.Phone != nil {
		up.SetPhone(*in.Phone)
	}
	if in.EmployeeNo != nil {
		up.SetEmployeeNo(*in.EmployeeNo)
	}
	if in.DepartmentID != nil {
		up.SetDepartmentID(*in.DepartmentID)
	}
	if in.Status != nil {
		up.SetStatus(*in.Status)
		if *in.Status == "suspended" {
			up.SetTokenVersion(target.TokenVersion + 1) // 停用即撤銷(D5)
		}
	}
	updated, err := up.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "update",
		ResourceType: "user", ResourceID: id.String(),
		Before: snapshotUser(target), After: snapshotUser(updated),
	})
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return updated, nil
}
```

`snapshotUser`(`repository.go`):JSON 序列化前**剔除** `password_hash` 等敏感欄位(Global Constraints):

```go
func snapshotUser(u *ent.User) string {
	m := map[string]any{
		"id": u.ID, "company_id": u.CompanyID, "department_id": u.DepartmentID,
		"email": u.Email, "name": u.Name, "role": u.Role, "status": u.Status,
		"data_scope": u.DataScope, "is_customer": u.IsCustomer,
		"token_version": u.TokenVersion,
	}
	b, _ := json.Marshal(m)
	return string(b)
}
```

`backend/proto/v1/user.proto` 增量(`UserService` 與 `ForceLogout` 已由 01 計畫 Task 12 宣告,以下為**加入既有 service 區塊**的新內容):

```proto
// 既有:service UserService { rpc ForceLogout(...) ... } — 以下 rpc 加入同一區塊:
//   rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
//   rpc GetUser(GetUserRequest) returns (User);
//   rpc CreateUser(CreateUserRequest) returns (User);
//   rpc UpdateUser(UpdateUserRequest) returns (User);
//   rpc AssignRole(AssignRoleRequest) returns (User);
//   rpc Deactivate(DeactivateRequest) returns (google.protobuf.Empty);

import "google/protobuf/timestamp.proto";
import "v1/common.proto";

message User {
  string id = 1;
  string company_id = 2;
  optional string department_id = 3;
  optional string email = 4;
  string name = 5;
  string phone = 6;
  optional string employee_no = 7;
  string role = 8;
  string data_scope = 9;
  string status = 10;
  bool is_customer = 11;
  optional string customer_id = 12;
  optional string account_name = 13;
  bool is_primary = 14;
  google.protobuf.Timestamp created_at = 15;
  google.protobuf.Timestamp updated_at = 16;
}

message ListUsersRequest {
  PageRequest page = 1;
  optional string company_id = 2;    // 僅 super 有效
  optional string department_id = 3;
  optional string role = 4;
  optional string status = 5;
}

message ListUsersResponse {
  repeated User users = 1;
  PageMeta meta = 2;
}

message GetUserRequest { string user_id = 1; }

message CreateUserRequest {
  string name = 1;
  string company_id = 2;
  optional string department_id = 3;
  string role = 4; // 內建角色;customer / guest 拒絕
}

message UpdateUserRequest {
  string user_id = 1;
  optional string name = 2;
  optional string department_id = 3;
  optional string phone = 4;
  optional string employee_no = 5;
  optional string status = 6; // active | suspended
}

message AssignRoleRequest {
  string user_id = 1;
  string role_code = 2;
  optional string department_id = 3; // guest 審核必填
}

message DeactivateRequest { string user_id = 1; }
```

`backend/internal/domain/users/handler.go`:同 Task 1 模式 — 每個 RPC 解析 actor(`rls.FromContext`,無則 `unauthenticated`)→ 映射 optional 欄位 → 呼叫 usecase → `toProtoUser`(不映射 `password_hash`)。

- [ ] **Step 4: 更新 01 計畫 Task 12 舊測試呼叫點並跑測試**

`backend/internal/domain/users/usecase_test.go`(Task 12 既有)中 `users.NewUsecase(c, nil)` 全部改為 `users.NewUsecase(c, nil, nil, audit.NoopRecorder{})`;01 計畫 Task 12 的範圍測試預期維持通過(`checkManageScope` 對 `DataScope:"company"` 他公司 actor 無角色 → permission_denied,語義相容;`DataScope:"all"` 但 Role 為空的舊測試 actor 需補 `Role: "super"` — 逐一核對修正)。

Run: `cd backend && buf generate && go test ./internal/domain/users/ -v`
Expected: PASS — guest 審核全流程、範圍矩陣、停用/重啟、ForceLogout 共用範圍函式;Task 12 既有測試全綠。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/database/txsql.go backend/internal/domain/users backend/proto/v1/user.proto
git commit -m "feat(backend): UserService CRUD、角色指派(g 規則同事務)、停用與 ForceLogout 範圍銜接(2.3)"
```

---

### Task 4: Logo 上傳 / Branding / PublicInfo / 公開發現端點(細部 2.4.1–2.4.3)

**Files:**
- Create: `backend/internal/domain/companies/branding.go`
- Create: `backend/internal/domain/companies/public.go`
- Update: `backend/proto/v1/company.proto`(增量 UpdateBranding / UpdatePublicInfo)
- Update: `backend/config/storage.go`(`StorageRoot`)
- Update: `backend/internal/server/domains.go`(`InitDomains()` 路由註冊)
- Test: `backend/internal/domain/companies/branding_test.go`、`backend/internal/domain/companies/public_test.go`

**Interfaces:**
- Consumes: Task 1 `companies.Usecase` / `identifierRe`;01 計畫 Task 11 `middleware.PublicPaths`(已含 `/api/v1/companies/public` 前綴,公開端點不需改白名單);`rls.Identity`。
- Produces: REST `POST /api/v1/companies/{companyID}/logo`(multipart,super 專用,D4)→ JSON `{file_asset_id, url}`;REST `GET /api/v1/companies/public/{identifier}`(無認證)→ 白名單公開欄位 JSON;proto `CompanyService.UpdateBranding(company_id、display_name、primary_color、logo_url)→ Company`、`UpdatePublicInfo(company_id、public_email、public_phone、public_address、terms_url、privacy_url、public_info、capabilities)→ Company`;`companies.LogoAssetWriter` 介面(佔位,見下)。
- **file_assets 佔位(TODO,04-master-data 計畫 子功能 3.6.2 接手)**:`file_assets` 表與 `fileassets.Service` 由 04 計畫提供。本 Task 定義最小介面與 stub;04 計畫落地後改注入真實實作並移除 stub,本 Task 呼叫點不變:

```go
// LogoAssetWriter 為 file_assets 寫入能力的最小介面(2.4.1 步驟 3)。
// TODO(04-master-data 計畫 子功能 3.6.2 接手):fileassets.Service 落地後改為直接注入,
// 本介面與 stubLogoAssetWriter 移除;屆時 file_assets 記錄落庫的整合驗收於 04 計畫補齊。
type LogoAssetWriter interface {
	// CreateLogoAsset 在交易內建立 file_assets 記錄,回傳 id 與下載 url。
	CreateLogoAsset(ctx context.Context, tx *ent.Tx, in LogoAsset) (id uuid.UUID, url string, err error)
	// LogoURLExists 驗證 url 為該公司既有 file_asset(UpdateBranding 用,2.4.2 步驟 1 介面說明)。
	LogoURLExists(ctx context.Context, companyID uuid.UUID, url string) (bool, error)
}

type LogoAsset struct {
	CompanyID              uuid.UUID
	StoragePath, Filename  string
	MimeType               string
	SizeBytes              int64
	CreatedBy              uuid.UUID
}
```

- [ ] **Step 1: 寫失敗測試**

`backend/internal/domain/companies/branding_test.go`:

```go
package companies_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/companies"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

// fakeLogoWriter 捕捉 CreateLogoAsset 呼叫;可注入錯誤模擬 DB 失敗。
type fakeLogoWriter struct {
	calls   []companies.LogoAsset
	failErr error
}

func (f *fakeLogoWriter) CreateLogoAsset(_ context.Context, _ *ent.Tx, in companies.LogoAsset) (uuid.UUID, string, error) {
	if f.failErr != nil {
		return uuid.Nil, "", f.failErr
	}
	f.calls = append(f.calls, in)
	id := uuid.New()
	return id, "/api/v1/files/" + id.String() + "/download", nil
}

func (f *fakeLogoWriter) LogoURLExists(_ context.Context, _ uuid.UUID, url string) (bool, error) {
	return len(url) > 10, nil
}

var pngHeader = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // + 填充至 >512B

func logoRequest(t *testing.T, companyID string, filename string, body []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, _ := w.CreateFormFile("file", filename)
	_, _ = part.Write(body)
	_ = w.Close()
	req := httptest.NewRequest("POST", "/api/v1/companies/"+companyID+"/logo", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func uploadRouter(uc *companies.Usecase, w companies.LogoAssetWriter, root string, actor rls.Identity) http.Handler {
	h := companies.NewBrandingHandler(uc, w, root)
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { // 模擬 Authenticate 注入身分
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(rw, req.WithContext(rls.NewContext(req.Context(), actor)))
		})
	})
	r.Post("/api/v1/companies/{companyID}/logo", h.UploadLogo)
	return r
}

func TestLogoUpload(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	uc := companies.NewUsecase(c, audit.NoopRecorder{})
	co, _ := uc.Create(ctx, superActor(), companies.CreateInput{Name: "甲", Identifier: "co-a"})
	root := t.TempDir()
	fw := &fakeLogoWriter{}
	srv := uploadRouter(uc, fw, root, superActor())

	body := append(pngHeader, bytes.Repeat([]byte{0}, 600)...)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, logoRequest(t, co.ID.String(), "logo.png", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("upload: want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if len(fw.calls) != 1 || fw.calls[0].MimeType != "image/png" {
		t.Fatalf("asset writer not called correctly: %+v", fw.calls)
	}
	after, _ := uc.Get(ctx, superActor(), co.ID)
	if after.LogoURL == nil || *after.LogoURL == "" {
		t.Fatal("companies.logo_url not updated")
	}

	// 偽造副檔名(exe 改名 png,magic bytes 不符)→ 400
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, logoRequest(t, co.ID.String(), "evil.png", []byte("MZ....fake-exe-bytes")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("fake png: want 400, got %d", rec.Code)
	}
	// 非 super → 403
	adminSrv := uploadRouter(uc, fw, root, rls.Identity{UserID: uuid.New(), CompanyID: co.ID, Role: "company_admin", DataScope: "company"})
	rec = httptest.NewRecorder()
	adminSrv.ServeHTTP(rec, logoRequest(t, co.ID.String(), "logo.png", body))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("company_admin upload: want 403, got %d", rec.Code)
	}
	// 不存在的公司 → 404
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, logoRequest(t, uuid.New().String(), "logo.png", body))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing company: want 404, got %d", rec.Code)
	}
}

func TestLogoUploadDBFailureCleansFile(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	uc := companies.NewUsecase(c, audit.NoopRecorder{})
	co, _ := uc.Create(ctx, superActor(), companies.CreateInput{Name: "甲", Identifier: "co-a"})
	root := t.TempDir()
	fw := &fakeLogoWriter{failErr: connect.NewError(connect.CodeInternal, errBoom)}
	srv := uploadRouter(uc, fw, root, superActor())

	body := append(pngHeader, bytes.Repeat([]byte{0}, 600)...)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, logoRequest(t, co.ID.String(), "logo.png", body))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", rec.Code)
	}
	// 孤兒檔清理:<root>/logos/<company>/ 下不得殘留任何檔案
	entries, _ := os.ReadDir(filepath.Join(root, "logos", co.ID.String()))
	if len(entries) != 0 {
		t.Fatalf("orphan file left behind: %v", entries)
	}
}
```

`backend/internal/domain/companies/public_test.go`:

```go
package companies_test

func TestPublicDiscovery(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	uc := companies.NewUsecase(c, audit.NoopRecorder{})
	co, _ := uc.Create(ctx, superActor(), companies.CreateInput{
		Name: "甲公司", Identifier: "co-a", DisplayName: "甲",
	})
	h := companies.NewPublicHandler(uc)
	r := chi.NewRouter()
	r.Get("/api/v1/companies/public/{identifier}", h.GetPublicInfo)

	// 無認證可取得;回傳不含內部欄位
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/companies/public/co-a", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var got map[string]any
	json.NewDecoder(rec.Body).Decode(&got)
	if got["display_name"] != "甲" {
		t.Fatalf("display_name missing: %v", got)
	}
	for _, internal := range []string{"id", "tax_id", "customer_code_prefix", "status", "created_at", "updated_at"} {
		if _, leaked := got[internal]; leaked {
			t.Fatalf("internal field %q leaked in public response", internal)
		}
	}
	if rec.Header().Get("Cache-Control") == "" {
		t.Fatal("cache header missing")
	}

	// 停用公司 → 404(不區分不存在與停用,2.4.3 步驟 1)
	c.Company.UpdateOneID(co.ID).SetStatus("suspended").ExecX(ctx)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/companies/public/co-a", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("suspended: want 404, got %d", rec.Code)
	}
	// 不存在 → 404;非法格式 → 400
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/companies/public/no-such", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing: want 404, got %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/companies/public/BAD_ID", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad identifier: want 400, got %d", rec.Code)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/companies/ -run 'TestLogo|TestPublic' -v`
Expected: FAIL — `NewBrandingHandler` / `NewPublicHandler` 未定義。

- [ ] **Step 3: 實作 branding.go / public.go / config / proto 增量**

`backend/config/storage.go` 加:

```go
StorageRoot string `envconfig:"STORAGE_ROOT"` // 本地檔案儲存根目錄(D17);預設 "./data/files"
```

`backend/internal/domain/companies/branding.go`:

```go
package companies

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"connectrpc.com/connect"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

const maxLogoBytes = 5 << 20 // 5 MB(D17)

// allowedLogoExt 白名單:副檔名 → 期望 MIME。
var allowedLogoExt = map[string]string{
	".jpg": "image/jpeg", ".jpeg": "image/jpeg",
	".png": "image/png", ".webp": "image/webp",
}

// sniffImageMIME 以 magic bytes 驗證實際格式(2.4.1 步驟 2 雙重檢查第二層)。
func sniffImageMIME(head []byte) string {
	switch {
	case len(head) >= 3 && head[0] == 0xFF && head[1] == 0xD8 && head[2] == 0xFF:
		return "image/jpeg"
	case len(head) >= 8 && hex.EncodeToString(head[:8]) == "89504e470d0a1a0a":
		return "image/png"
	case len(head) >= 12 && string(head[:4]) == "RIFF" && string(head[8:12]) == "WEBP":
		return "image/webp"
	}
	return ""
}

type BrandingHandler struct {
	uc     *Usecase
	writer LogoAssetWriter
	root   string
	audit  audit.Recorder
}

func NewBrandingHandler(uc *Usecase, w LogoAssetWriter, storageRoot string) *BrandingHandler {
	return &BrandingHandler{uc: uc, writer: w, root: storageRoot, audit: audit.NoopRecorder{}}
}

// UploadLogo:REST multipart(檔案上傳保留 REST,D4);僅 super(規格 3.1.1)。
func (h *BrandingHandler) UploadLogo(w http.ResponseWriter, r *http.Request) {
	actor, ok := rls.FromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !actor.HasRole("super") {
		http.Error(w, "only super can upload logo", http.StatusForbidden)
		return
	}
	companyID, err := uuid.Parse(chi.URLParam(r, "companyID"))
	if err != nil {
		http.Error(w, "bad company id", http.StatusBadRequest)
		return
	}
	if _, err := h.uc.Get(r.Context(), actor, companyID); err != nil { // 不存在/已刪除 → not_found
		http.Error(w, "company not found", http.StatusNotFound)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxLogoBytes)
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required or too large", http.StatusBadRequest)
		return
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	wantMIME, ok := allowedLogoExt[ext] // 第一層:副檔名白名單
	if !ok {
		http.Error(w, "only jpeg/png/webp allowed", http.StatusBadRequest)
		return
	}
	head := make([]byte, 512)
	n, _ := file.Read(head)
	if got := sniffImageMIME(head[:n]); got != wantMIME { // 第二層:magic bytes
		http.Error(w, "file content does not match extension", http.StatusBadRequest)
		return
	}

	// 落本地儲存:<root>/logos/<company_id>/<隨機檔名>
	dir := filepath.Join(h.root, "logos", companyID.String())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	var randB [8]byte
	_, _ = rand.Read(randB[:])
	filename := hex.EncodeToString(randB[:]) + ext
	storagePath := filepath.Join(dir, filename)
	if _, err := file.Seek(0, 0); err == nil { // multipart sectionRead 支援 Seek
		// 從頭寫入完整檔案
	}
	if err := writeFile(storagePath, file); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	size := header.Size

	// DB 交易:file_assets 記錄 + companies.logo_url + 稽核;失敗清除已落碟檔案(2.4.1 步驟 3)
	tx, err := h.uc.db.Tx(r.Context())
	if err != nil {
		_ = os.Remove(storagePath)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	assetID, url, err := h.writer.CreateLogoAsset(r.Context(), tx, LogoAsset{
		CompanyID: companyID, StoragePath: storagePath, Filename: filename,
		MimeType: wantMIME, SizeBytes: size, CreatedBy: actor.UserID,
	})
	if err != nil {
		_ = tx.Rollback()
		_ = os.Remove(storagePath)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	if _, err := tx.Company.UpdateOneID(companyID).
		SetLogoURL(url).SetLogoFileID(assetID).Save(r.Context()); err != nil {
		_ = tx.Rollback()
		_ = os.Remove(storagePath)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	_ = h.audit.Record(r.Context(), audit.Entry{
		ActorID: actor.UserID.String(), Action: "update",
		ResourceType: "company", ResourceID: companyID.String(),
		After: `{"logo_url":"` + url + `"}`,
	})
	if err := tx.Commit(); err != nil {
		_ = os.Remove(storagePath)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	// 舊 Logo 的 file_assets 記錄保留(歷史可查,2.4.1 步驟 4);1.0 不清舊檔。
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"file_asset_id": assetID.String(), "url": url,
	})
}

// writeFile 將上傳串流完整寫入 path(實作:os.Create + io.Copy + Sync)。
```

UpdateBranding / UpdatePublicInfo usecase(同檔或 usecase.go 追加):

```go
var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type BrandingInput struct {
	DisplayName, PrimaryColor, LogoURL *string
}

func (u *Usecase) UpdateBranding(ctx context.Context, actor rls.Identity, id uuid.UUID, in BrandingInput, writer LogoAssetWriter) (*ent.Company, error) {
	co, err := u.Get(ctx, actor, id) // not_found + 範圍(super 全域 / company_admin 自己公司)
	if err != nil {
		return nil, err
	}
	if in.PrimaryColor != nil && *in.PrimaryColor != "" && !hexColorRe.MatchString(*in.PrimaryColor) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("primary_color must be #RRGGBB"))
	}
	if in.LogoURL != nil && *in.LogoURL != "" {
		ok, err := writer.LogoURLExists(ctx, co.ID, *in.LogoURL) // 僅接受既有 file_asset 的 url
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		if !ok {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("logo_url is not an existing asset of this company"))
		}
	}
	// tx:僅更新出現的欄位 + 稽核(action=update, resource_type=company),同 Task 1 Update 模式。
	...
}

type PublicInfoInput struct {
	PublicEmail, PublicPhone, PublicAddress *string
	TermsURL, PrivacyURL                    *string
	PublicInfo, Capabilities                map[string]any // 整體覆寫,不做深層 merge(2.4.2 步驟 3)
	HasCapabilities                         bool           // proto presence:capabilities 是否出現於請求
}

func (u *Usecase) UpdatePublicInfo(ctx context.Context, actor rls.Identity, id uuid.UUID, in PublicInfoInput) (*ent.Company, error) {
	co, err := u.Get(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	// capabilities 為 AI 協定預留欄位(規格 17.3),1.0 僅 super 可維護(2.4.2 步驟 1)
	if in.HasCapabilities && !actor.HasRole("super") {
		return nil, connect.NewError(connect.CodePermissionDenied,
			errors.New("capabilities is super-only"))
	}
	if in.PublicEmail != nil && *in.PublicEmail != "" {
		if _, err := mail.ParseAddress(*in.PublicEmail); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("bad public_email"))
		}
	}
	for _, u2 := range []*string{in.TermsURL, in.PrivacyURL} {
		if u2 != nil && *u2 != "" {
			parsed, err := url.Parse(*u2)
			if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("url must be http(s)"))
			}
		}
	}
	// tx:更新出現的欄位(public_info / capabilities 整體覆寫)+ 稽核,同 Task 1 模式。
	_ = co
	...
}
```

`backend/internal/domain/companies/public.go`:

```go
package companies

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	entcompany "github.com/salesorder/sales-order-1.0/backend/ent/company"
)

type PublicHandler struct{ uc *Usecase }

func NewPublicHandler(uc *Usecase) *PublicHandler { return &PublicHandler{uc: uc} }

// GetPublicInfo:無認證公開端點(2.4.3);路由於 main.go 掛在 Authenticate 之外
// (01 計畫 Task 11 PublicPaths 已含 /api/v1/companies/public 前綴)。
func (h *PublicHandler) GetPublicInfo(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "identifier")
	if !identifierRe.MatchString(identifier) {
		http.Error(w, "bad identifier", http.StatusBadRequest)
		return
	}
	co, err := h.uc.db.Company.Query().Where(
		entcompany.IdentifierEQ(identifier),
		entcompany.StatusEQ("active"),
		entcompany.DeletedAtIsNil(),
	).Only(r.Context())
	if err != nil {
		// 不區分「不存在」與「已停用」,避免列舉探測(2.4.3 步驟 1)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60") // 短時間快取,降低查詢放大
	w.Header().Set("Content-Type", "application/json")
	// 白名單欄位輸出(步驟 2);結構保持穩定,擴充採新增不破壞(步驟 4)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"display_name":   co.DisplayName,
		"logo_url":       co.LogoURL,
		"primary_color":  co.PrimaryColor,
		"public_email":   co.PublicEmail,
		"public_phone":   co.PublicPhone,
		"public_address": co.PublicAddress,
		"terms_url":      co.TermsURL,
		"privacy_url":    co.PrivacyURL,
		"public_info":    co.PublicInfo,
		"capabilities":   co.Capabilities,
	})
}
```

`backend/proto/v1/company.proto` 增量(加入 `service CompanyService` 區塊):

```proto
// 加入 service 區塊:
//   rpc UpdateBranding(UpdateBrandingRequest) returns (Company);
//   rpc UpdatePublicInfo(UpdatePublicInfoRequest) returns (Company);

message UpdateBrandingRequest {
  string company_id = 1;
  optional string display_name = 2;
  optional string primary_color = 3; // #RRGGBB
  optional string logo_url = 4;      // 僅接受 2.4.1 上傳產生的既有 url
}

message UpdatePublicInfoRequest {
  string company_id = 1;
  optional string public_email = 2;
  optional string public_phone = 3;
  optional string public_address = 4;
  optional string terms_url = 5;
  optional string privacy_url = 6;
  google.protobuf.Struct public_info = 7;   // 出現即整體覆寫
  google.protobuf.Struct capabilities = 8;  // 出現且非 super → permission_denied
}
```

`main.go` 路由註冊(authenticated mux 之外):

```go
	r.Route("/api/v1/companies", func(r chi.Router) {
		r.Post("/{companyID}/logo", brandingH.UploadLogo) // 掛於 Authenticate 保護群組內
	})
	r.Get("/api/v1/companies/public/{identifier}", publicH.GetPublicInfo) // 保護群組外
```

stub(`branding.go` 同檔,04 接手前讓 main.go 可組裝):

```go
// stubLogoAssetWriter:TODO(04-master-data 計畫 子功能 3.6.2 接手)移除。
type stubLogoAssetWriter struct{}

func (stubLogoAssetWriter) CreateLogoAsset(_ context.Context, _ *ent.Tx, in LogoAsset) (uuid.UUID, string, error) {
	id := uuid.New()
	return id, "/api/v1/files/" + id.String() + "/download", nil // file_assets 記錄暫不落庫
}

func (stubLogoAssetWriter) LogoURLExists(_ context.Context, _ uuid.UUID, u string) (bool, error) {
	return strings.HasPrefix(u, "/api/v1/files/"), nil
}
```

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && buf generate && go test ./internal/domain/companies/ -v`
Expected: PASS — 上傳成敗、magic bytes 阻擋、孤兒檔清理、權限、公開端點白名單與 404/400;Task 1 測試全綠。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/companies backend/proto/v1/company.proto backend/config/storage.go backend/internal/server
git commit -m "feat(backend): Logo 上傳、Branding/PublicInfo、公開發現端點(2.4)"
```

---

### Task 5: roles + role_permissions schema 與 seed + RoleService CRUD(細部 2.9.1–2.9.2)

**Files:**
- Create: `backend/ent/schema/role.go`
- Create: `backend/ent/schema/rolepermission.go`
- Create: `backend/database/migrations/00005_roles_seed.sql`
- Create: `backend/proto/v1/admin.proto`
- Create: `backend/internal/domain/roles/usecase.go`
- Create: `backend/internal/domain/roles/repository.go`
- Create: `backend/internal/domain/roles/handler.go`
- Test: `backend/internal/domain/roles/usecase_test.go`、`backend/internal/domain/roles/seed_test.go`

**Interfaces:**
- Consumes: Task 3 `database.ExecTx / QueryTx`;`testutil.NewEntClient`;01 計畫 Task 14 developer 帳號 seed(01 未建 roles 表,本 migration 為 roles 表首次 seed,冪等保證 developer 已存在時不重複)。
- Produces: `ent.Role`(`id`、`code`、`name`、`data_scope` enum all/company/department/self、`is_system`、`is_active`、`deleted_at`;`code` 部分唯一索引);`ent.RolePermission`(`id`、`role_id`、`resource`、`action`、`created_at`;`role_id+resource+action` 唯一);`roles.ValidResources` / `roles.ValidActions` 詞彙表;proto `RoleService` — `ListRoles(page、is_system、is_active)→ ListRolesResponse{roles, meta}`;`GetRole(role_id)→ Role`;`CreateRole(code、name、data_scope)→ Role`;`UpdateRole(role_id、code?、name?、data_scope?、is_active?)→ Role`;`DeleteRole(role_id)→ Empty`;usecase `roles.Usecase` 同名方法 + `roles.NewUsecase(db, enf, rec)`;`testutil.ApplyMigrationFile(t, client, name)`(自 00003 邏輯抽出,供本 Task 與後續 migration 測試)。
- roles / role_permissions **不加 RLS**:系統級參照資料,無 `company_id` 欄位;存取由 usecase 範圍判斷 + Casbin(Task 7)控制。

- [ ] **Step 1: testutil.ApplyMigrationFile 抽取 + 寫失敗測試**

`testutil/db.go` 追加(自 `NewEntClientWithRLS` 內的 00003 讀檔邏輯抽出,原函式改呼叫之):

```go
// ApplyMigrationFile 對測試庫執行 goose migration 的 Up 段(去掉 goose 標記)。
func ApplyMigrationFile(t *testing.T, c *ent.Client, name string) {
	t.Helper()
	sqlBytes, err := os.ReadFile(filepath.Join("../../database/migrations", name))
	if err != nil {
		t.Fatalf("read migration %s: %v", name, err)
	}
	up := strings.SplitN(string(sqlBytes), "-- +goose Down", 2)[0]
	up = strings.ReplaceAll(up, "-- +goose Up", "")
	if _, err := c.DB().ExecContext(context.Background(), up); err != nil {
		t.Fatalf("apply migration %s: %v", name, err)
	}
}
```

`backend/internal/domain/roles/seed_test.go`:

```go
package roles_test

import (
	"context"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestRolesSeedIdempotent(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()

	testutil.ApplyMigrationFile(t, c, "00005_roles_seed.sql")
	// 七個內建角色,is_system = true,data_scope 對應正確(2.9.1 步驟 2)
	want := map[string]string{
		"super": "all", "company_admin": "company", "dept_admin": "department",
		"staff": "department", "customer": "self", "guest": "self", "developer": "all",
	}
	roles, err := c.Role.Query().All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) != 7 {
		t.Fatalf("want 7 builtin roles, got %d", len(roles))
	}
	for _, r := range roles {
		if !r.IsSystem || !r.IsActive {
			t.Fatalf("role %s should be system+active", r.Code)
		}
		if want[r.Code] != r.DataScope {
			t.Fatalf("role %s data_scope: want %s, got %s", r.Code, want[r.Code], r.DataScope)
		}
	}
	// 預設權限:staff 有 sales_orders/manage;guest 與 developer 無權限列
	perms, _ := c.RolePermission.Query().All(ctx)
	if len(perms) == 0 {
		t.Fatal("default role_permissions missing")
	}
	// 冪等:重複執行不報錯、列數不變(2.9.1 步驟 4)
	nRoles, nPerms := len(roles), len(perms)
	testutil.ApplyMigrationFile(t, c, "00005_roles_seed.sql")
	if got, _ := c.Role.Query().Count(ctx); got != nRoles {
		t.Fatalf("re-seed changed roles: %d -> %d", nRoles, got)
	}
	if got, _ := c.RolePermission.Query().Count(ctx); got != nPerms {
		t.Fatalf("re-seed changed permissions: %d -> %d", nPerms, got)
	}
}
```

`backend/internal/domain/roles/usecase_test.go`:

```go
package roles_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/roles"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func superActor() rls.Identity {
	return rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(), DataScope: "all", Role: "super"}
}

func TestBuiltinRoleProtection(t *testing.T) {
	c := testutil.NewEntClient(t)
	testutil.ApplyMigrationFile(t, c, "00005_roles_seed.sql")
	ctx := context.Background()
	uc := roles.NewUsecase(c, nil, audit.NoopRecorder{})

	staff, _ := c.Role.Query().Where(/* code = staff */).Only(ctx)
	// 內建角色不可改 code / data_scope → failed_precondition(D9)
	newCode := "staff2"
	if _, err := uc.Update(ctx, superActor(), staff.ID, roles.UpdateInput{Code: &newCode}); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("builtin code change: want failed_precondition, got %v", err)
	}
	scope := "self"
	if _, err := uc.Update(ctx, superActor(), staff.ID, roles.UpdateInput{DataScope: &scope}); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("builtin data_scope change: want failed_precondition, got %v", err)
	}
	// name / is_active 可調
	newName := "職員"
	if _, err := uc.Update(ctx, superActor(), staff.ID, roles.UpdateInput{Name: &newName}); err != nil {
		t.Fatalf("builtin name change should be allowed: %v", err)
	}
	// 內建角色不可刪除 → failed_precondition
	if err := uc.Delete(ctx, superActor(), staff.ID); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("builtin delete: want failed_precondition, got %v", err)
	}
}

func TestCustomRoleLifecycle(t *testing.T) {
	c := testutil.NewEntClient(t)
	testutil.ApplyMigrationFile(t, c, "00005_roles_seed.sql")
	ctx := context.Background()
	uc := roles.NewUsecase(c, nil, audit.NoopRecorder{})

	// data_scope 必填且僅 company / department(2.9.2 步驟 2)
	if _, err := uc.Create(ctx, superActor(), roles.CreateInput{Code: "sales_rep", Name: "業務代表"}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("missing data_scope: want invalid_argument, got %v", err)
	}
	if _, err := uc.Create(ctx, superActor(), roles.CreateInput{Code: "sales_rep", Name: "業務代表", DataScope: "all"}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("data_scope=all: want invalid_argument, got %v", err)
	}
	r, err := uc.Create(ctx, superActor(), roles.CreateInput{Code: "sales_rep", Name: "業務代表", DataScope: "company"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if r.IsSystem {
		t.Fatal("custom role must not be system")
	}
	// code 重複 → already_exists
	if _, err := uc.Create(ctx, superActor(), roles.CreateInput{Code: "sales_rep", Name: "x", DataScope: "company"}); connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("dup code: want already_exists, got %v", err)
	}
	// 非 super 寫入 → permission_denied
	admin := rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(), DataScope: "company", Role: "company_admin"}
	if _, err := uc.Create(ctx, admin, roles.CreateInput{Code: "x", Name: "x", DataScope: "company"}); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("company_admin create: want permission_denied, got %v", err)
	}
	// company_admin 可 List / Get(供指派與 policy 設定選用,2.9.2 步驟 1)
	if _, _, err := uc.List(ctx, admin, roles.ListQuery{Page: 1, PerPage: 20}); err != nil {
		t.Fatalf("company_admin list: %v", err)
	}

	// 綁定使用者後不可刪除(2.9.2 步驟 4):直接插一筆 g 規則
	tx, _ := c.Tx(ctx)
	if err := database.ExecTx(ctx, tx,
		`INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('g', $1, 'sales_rep', $2)`,
		uuid.New().String(), uuid.New().String()); err != nil {
		t.Fatal(err)
	}
	_ = tx.Commit()
	if err := uc.Delete(ctx, superActor(), r.ID); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("delete with bound users: want failed_precondition, got %v", err)
	}
	// 清掉綁定後可軟刪除;同 code 可重建
	tx2, _ := c.Tx(ctx)
	_ = database.ExecTx(ctx, tx2, `DELETE FROM casbin_rule WHERE ptype = 'g' AND v1 = 'sales_rep'`)
	_ = tx2.Commit()
	if err := uc.Delete(ctx, superActor(), r.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := uc.Create(ctx, superActor(), roles.CreateInput{Code: "sales_rep", Name: "重建", DataScope: "department"}); err != nil {
		t.Fatalf("recreate same code after soft delete: %v", err)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/roles/ -v`
Expected: FAIL — `ent.Role` / `roles.NewUsecase` 未定義(編譯失敗)。

- [ ] **Step 3: 實作 schema、migration seed、usecase、proto**

`backend/ent/schema/role.go`:

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

// Role 為角色定義(細部文件 2.9.1);內建 is_system=true,自訂僅軟刪除可用(規格 5.4)。
type Role struct{ ent.Schema }

func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("code").NotEmpty(),
		field.String("name").NotEmpty(),
		field.Enum("data_scope").Values("all", "company", "department", "self"),
		field.Bool("is_system").Default(false),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (Role) Edges() []ent.Edge {
	return []ent.Edge{edge.To("permissions", RolePermission.Type)}
}

func (Role) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}
```

`backend/ent/schema/rolepermission.go`:

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

// RolePermission 為角色功能權限(resource × action),CASL ability 的來源(2.9.1)。
// 角色刪除時同交易清空(2.9.2 步驟 4),本表不需 deleted_at。
type RolePermission struct{ ent.Schema }

func (RolePermission) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("role_id", uuid.UUID{}),
		field.String("resource").NotEmpty(),
		field.String("action").NotEmpty(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (RolePermission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("role", Role.Type).Ref("permissions").
			Field("role_id").Unique().Required(),
	}
}

func (RolePermission) Indexes() []ent.Index {
	return []ent.Index{index.Fields("role_id", "resource", "action").Unique()}
}
```

`backend/internal/domain/roles/repository.go` 詞彙表(2.9.3 驗證共用;subject 詞彙與 01 計畫 Task 13 CASL 契約同源):

```go
package roles

// ValidResources 為系統已定義 resource 列舉(CASL subject 詞彙)。
var ValidResources = map[string]bool{
	"all": true, "users": true, "companies": true, "departments": true,
	"roles": true, "policies": true, "customers": true, "products": true,
	"sales_orders": true, "return_requests": true, "dispatch": true,
	"printing": true, "metadicts": true, "audit_logs": true,
	"customer_products": true, "notifications": true, "account": true,
}

// ValidActions 為系統已定義 action 列舉。
var ValidActions = map[string]bool{
	"create": true, "read": true, "update": true, "delete": true,
	"manage": true, "print": true, "dispatch": true,
}
```

`backend/database/migrations/00005_roles_seed.sql`(冪等;`roles` 不加 RLS — 系統級參照表,存取由 usecase + Casbin 控制):

```sql
-- +goose Up
-- 細部文件 2.9.1:七內建角色 + 預設功能權限。
-- 冪等:code 已存在(未刪除)則略過該角色整組 seed(步驟 4),不覆寫 Web 已調整的權限。
-- 註:01 計畫未建 roles 表(developer 角色語義原由 users.role 承載),本 migration 為首次 seed。

INSERT INTO roles (id, code, name, data_scope, is_system, is_active, created_at, updated_at)
SELECT '00000000-0000-0000-0000-000000000001', 'super', '系統管理員', 'all', true, true, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'super' AND deleted_at IS NULL);

INSERT INTO roles (id, code, name, data_scope, is_system, is_active, created_at, updated_at)
SELECT '00000000-0000-0000-0000-000000000002', 'company_admin', '公司管理員', 'company', true, true, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'company_admin' AND deleted_at IS NULL);

INSERT INTO roles (id, code, name, data_scope, is_system, is_active, created_at, updated_at)
SELECT '00000000-0000-0000-0000-000000000003', 'dept_admin', '部門管理員', 'department', true, true, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'dept_admin' AND deleted_at IS NULL);

INSERT INTO roles (id, code, name, data_scope, is_system, is_active, created_at, updated_at)
SELECT '00000000-0000-0000-0000-000000000004', 'staff', '職員', 'department', true, true, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'staff' AND deleted_at IS NULL);

INSERT INTO roles (id, code, name, data_scope, is_system, is_active, created_at, updated_at)
SELECT '00000000-0000-0000-0000-000000000005', 'customer', '客戶', 'self', true, true, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'customer' AND deleted_at IS NULL);

INSERT INTO roles (id, code, name, data_scope, is_system, is_active, created_at, updated_at)
SELECT '00000000-0000-0000-0000-000000000006', 'guest', '待審核', 'self', true, true, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'guest' AND deleted_at IS NULL);

INSERT INTO roles (id, code, name, data_scope, is_system, is_active, created_at, updated_at)
SELECT '00000000-0000-0000-0000-000000000007', 'developer', '開發者', 'all', true, true, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'developer' AND deleted_at IS NULL);

-- 預設功能權限(2.9.1 步驟 3,與 01 計畫 Task 13 內建 ability 表同源);
-- 每角色僅在「尚無任何權限列」時 seed,不覆寫 Web 調整結果。
INSERT INTO role_permissions (id, role_id, resource, action, created_at)
SELECT gen_random_uuid(), '00000000-0000-0000-0000-000000000001', v.resource, v.action, now()
FROM (VALUES ('all', 'manage')) AS v(resource, action)
WHERE NOT EXISTS (SELECT 1 FROM role_permissions rp WHERE rp.role_id = '00000000-0000-0000-0000-000000000001');

INSERT INTO role_permissions (id, role_id, resource, action, created_at)
SELECT gen_random_uuid(), '00000000-0000-0000-0000-000000000002', v.resource, v.action, now()
FROM (VALUES
  ('users', 'manage'), ('customers', 'manage'), ('sales_orders', 'manage'),
  ('return_requests', 'manage'), ('dispatch', 'manage'), ('printing', 'manage'),
  ('audit_logs', 'read')) AS v(resource, action)
WHERE NOT EXISTS (SELECT 1 FROM role_permissions rp WHERE rp.role_id = '00000000-0000-0000-0000-000000000002');

INSERT INTO role_permissions (id, role_id, resource, action, created_at)
SELECT gen_random_uuid(), '00000000-0000-0000-0000-000000000003', v.resource, v.action, now()
FROM (VALUES
  ('customers', 'manage'), ('sales_orders', 'manage'), ('return_requests', 'manage'),
  ('dispatch', 'manage'), ('printing', 'manage'),
  ('users', 'read'), ('users', 'update')) AS v(resource, action)
WHERE NOT EXISTS (SELECT 1 FROM role_permissions rp WHERE rp.role_id = '00000000-0000-0000-0000-000000000003');

INSERT INTO role_permissions (id, role_id, resource, action, created_at)
SELECT gen_random_uuid(), '00000000-0000-0000-0000-000000000004', v.resource, v.action, now()
FROM (VALUES
  ('customers', 'read'), ('customers', 'create'), ('customers', 'update'),
  ('sales_orders', 'manage'),
  ('return_requests', 'read'), ('return_requests', 'update'),
  ('dispatch', 'manage'), ('printing', 'manage')) AS v(resource, action)
WHERE NOT EXISTS (SELECT 1 FROM role_permissions rp WHERE rp.role_id = '00000000-0000-0000-0000-000000000004');

INSERT INTO role_permissions (id, role_id, resource, action, created_at)
SELECT gen_random_uuid(), '00000000-0000-0000-0000-000000000005', v.resource, v.action, now()
FROM (VALUES
  ('sales_orders', 'read'), ('sales_orders', 'create'),
  ('return_requests', 'read'), ('return_requests', 'create'),
  ('customer_products', 'read'), ('notifications', 'read')) AS v(resource, action)
WHERE NOT EXISTS (SELECT 1 FROM role_permissions rp WHERE rp.role_id = '00000000-0000-0000-0000-000000000005');

-- guest(0006)與 developer(0007)無預設權限列:guest 待審核無業務權限;
-- developer 繞過由 middleware 實現(D8),不查表。

-- +goose Down
DELETE FROM role_permissions WHERE role_id IN (
  '00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000002',
  '00000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000004',
  '00000000-0000-0000-0000-000000000005','00000000-0000-0000-0000-000000000006',
  '00000000-0000-0000-0000-000000000007');
DELETE FROM roles WHERE id IN (
  '00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000002',
  '00000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000004',
  '00000000-0000-0000-0000-000000000005','00000000-0000-0000-0000-000000000006',
  '00000000-0000-0000-0000-000000000007');
```

`backend/internal/domain/roles/usecase.go`:

```go
// Package roles 為角色與功能權限 domain(細部文件 2.9)。
package roles

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entrole "github.com/salesorder/sales-order-1.0/backend/ent/role"
	entrp "github.com/salesorder/sales-order-1.0/backend/ent/rolepermission"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

type Usecase struct {
	db    *ent.Client
	enf   *casbin.Enforcer // 可為 nil;非 nil 時刪除角色後 LoadPolicy
	audit audit.Recorder
}

func NewUsecase(db *ent.Client, enf *casbin.Enforcer, rec audit.Recorder) *Usecase {
	return &Usecase{db: db, enf: enf, audit: rec}
}

type ListQuery struct {
	Page, PerPage int
	IsSystem      *bool
	IsActive      *bool
}

type CreateInput struct{ Code, Name, DataScope string }

type UpdateInput struct {
	Code, Name, DataScope *string
	IsActive              *bool
}

func (u *Usecase) List(ctx context.Context, actor rls.Identity, q ListQuery) ([]*ent.Role, int, error) {
	// List / Get 開放 super 與 company_admin(2.9.2 步驟 1)
	if !actor.HasRole("super") && !actor.HasRole("company_admin") {
		return nil, 0, connect.NewError(connect.CodePermissionDenied, errors.New("list not allowed"))
	}
	page, perPage, err := normalizePage(q.Page, q.PerPage)
	if err != nil {
		return nil, 0, err
	}
	query := u.db.Role.Query().Where(entrole.DeletedAtIsNil())
	if q.IsSystem != nil {
		query = query.Where(entrole.IsSystemEQ(*q.IsSystem))
	}
	if q.IsActive != nil {
		query = query.Where(entrole.IsActiveEQ(*q.IsActive))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, connect.NewError(connect.CodeInternal, err)
	}
	items, err := query.Order(ent.Asc(entrole.FieldCreatedAt)).
		Offset((page - 1) * perPage).Limit(perPage).All(ctx)
	if err != nil {
		return nil, 0, connect.NewError(connect.CodeInternal, err)
	}
	return items, total, nil
}

func (u *Usecase) Get(ctx context.Context, actor rls.Identity, id uuid.UUID) (*ent.Role, error) {
	if !actor.HasRole("super") && !actor.HasRole("company_admin") {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("get not allowed"))
	}
	r, err := u.db.Role.Query().Where(entrole.IDEQ(id), entrole.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("role not found"))
	}
	return r, nil
}

func (u *Usecase) Create(ctx context.Context, actor rls.Identity, in CreateInput) (*ent.Role, error) {
	if !actor.HasRole("super") { // 寫入僅 super(D9)
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("only super can manage roles"))
	}
	if in.Code == "" || in.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("code and name required"))
	}
	// 自訂角色 data_scope 必填且僅 company / department(規格 3.2);is_system 恆 false,
	// proto 不提供 is_system 欄位,API 層無從嘗試建立內建角色(2.9.2 步驟 2)。
	if in.DataScope != "company" && in.DataScope != "department" {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("custom role data_scope must be company or department"))
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if exists, err := tx.Role.Query().Where(
		entrole.CodeEQ(in.Code), entrole.DeletedAtIsNil()).Exist(ctx); err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	} else if exists {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("code already in use"))
	}
	r, err := tx.Role.Create().SetCode(in.Code).SetName(in.Name).
		SetDataScope(in.DataScope).SetIsSystem(false).SetIsActive(true).Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "create",
		ResourceType: "role", ResourceID: r.ID.String(), After: snapshotRole(r),
	})
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return r, nil
}

func (u *Usecase) Update(ctx context.Context, actor rls.Identity, id uuid.UUID, in UpdateInput) (*ent.Role, error) {
	if !actor.HasRole("super") {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("only super can manage roles"))
	}
	r, err := u.db.Role.Query().Where(entrole.IDEQ(id), entrole.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("role not found"))
	}
	// 內建角色不可改 code / data_scope(D9);name / is_active 可調
	if r.IsSystem && (in.Code != nil || in.DataScope != nil) {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("builtin role code/data_scope are immutable"))
	}
	if in.DataScope != nil && *in.DataScope != "company" && *in.DataScope != "department" && !r.IsSystem {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("custom role data_scope must be company or department"))
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	up := tx.Role.UpdateOneID(id)
	if in.Code != nil {
		if exists, err := tx.Role.Query().Where(
			entrole.CodeEQ(*in.Code), entrole.IDNEQ(id), entrole.DeletedAtIsNil()).Exist(ctx); err != nil {
			_ = tx.Rollback()
			return nil, connect.NewError(connect.CodeInternal, err)
		} else if exists {
			_ = tx.Rollback()
			return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("code already in use"))
		}
		up.SetCode(*in.Code)
	}
	if in.Name != nil {
		up.SetName(*in.Name)
	}
	if in.DataScope != nil {
		up.SetDataScope(*in.DataScope)
	}
	if in.IsActive != nil {
		up.SetIsActive(*in.IsActive)
	}
	updated, err := up.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "update",
		ResourceType: "role", ResourceID: id.String(),
		Before: snapshotRole(r), After: snapshotRole(updated),
	})
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return updated, nil
}

func (u *Usecase) Delete(ctx context.Context, actor rls.Identity, id uuid.UUID) error {
	if !actor.HasRole("super") {
		return connect.NewError(connect.CodePermissionDenied, errors.New("only super can manage roles"))
	}
	r, err := u.db.Role.Query().Where(entrole.IDEQ(id), entrole.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, errors.New("role not found"))
	}
	if r.IsSystem {
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("builtin role cannot be deleted"))
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	// 有使用者綁定(Casbin g 規則)拒絕刪除,提示先改派(2.9.2 步驟 4)
	rows, err := database.QueryTx(ctx, tx,
		`SELECT 1 FROM casbin_rule WHERE ptype = 'g' AND v1 = $1 LIMIT 1`, r.Code)
	if err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	bound := rows.Next()
	rows.Close()
	if bound {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("role still assigned to users; reassign them first"))
	}
	// 同交易:軟刪除角色、清 g 規則與 role_permissions、寫稽核
	if _, err := tx.Role.UpdateOneID(id).SetDeletedAt(time.Now()).Save(ctx); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	if err := database.ExecTx(ctx, tx,
		`DELETE FROM casbin_rule WHERE ptype = 'g' AND v1 = $1`, r.Code); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	if _, err := tx.RolePermission.Delete().
		Where(entrp.RoleIDEQ(id)).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "delete",
		ResourceType: "role", ResourceID: id.String(), Before: snapshotRole(r),
	})
	if err := tx.Commit(); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if u.enf != nil {
		_ = u.enf.LoadPolicy()
	}
	return nil
}
```

`backend/proto/v1/admin.proto`(RoleService;PolicyService 於 Task 7 增量):

```proto
syntax = "proto3";
package salesorder.v1;

import "google/protobuf/empty.proto";
import "v1/common.proto";

service RoleService {
  rpc ListRoles(ListRolesRequest) returns (ListRolesResponse);
  rpc GetRole(GetRoleRequest) returns (Role);
  rpc CreateRole(CreateRoleRequest) returns (Role);
  rpc UpdateRole(UpdateRoleRequest) returns (Role);
  rpc DeleteRole(DeleteRoleRequest) returns (google.protobuf.Empty);
}

message Role {
  string id = 1;
  string code = 2;
  string name = 3;
  string data_scope = 4;
  bool is_system = 5;
  bool is_active = 6;
}

message ListRolesRequest {
  PageRequest page = 1;
  optional bool is_system = 2;
  optional bool is_active = 3;
}

message ListRolesResponse {
  repeated Role roles = 1;
  PageMeta meta = 2;
}

message GetRoleRequest { string role_id = 1; }

message CreateRoleRequest {
  string code = 1;
  string name = 2;
  string data_scope = 3; // 自訂角色僅 company | department
}

message UpdateRoleRequest {
  string role_id = 1;
  optional string code = 2;       // 內建角色拒絕
  optional string name = 3;
  optional string data_scope = 4; // 內建角色拒絕
  optional bool is_active = 5;
}

message DeleteRoleRequest { string role_id = 1; }
```

- [ ] **Step 4: 移除 Task 3 的 roleExists 回落路徑並跑測試**

roles 表已落地,依 Task 3 的 TODO 將 `users.Usecase.roleExists` 改為一律查表(`roles` 表 + `is_active = true` + 未刪除),移除內建列舉回落;Task 3 測試加 `testutil.ApplyMigrationFile(t, c, "00005_roles_seed.sql")` 使角色可查。

Run: `cd backend && go generate ./ent/ && buf generate && go test ./internal/domain/roles/ ./internal/domain/users/ -v`
Expected: PASS — seed 冪等、內建保護、自訂生命週期、Task 3 測試不回歸。

- [ ] **Step 5: Commit**

```bash
git add backend/ent/schema/role.go backend/ent/schema/rolepermission.go backend/database/migrations/00005_roles_seed.sql backend/proto/v1/admin.proto backend/internal/domain/roles backend/internal/testutil backend/internal/domain/users
git commit -m "feat(backend): roles/role_permissions schema、內建角色 seed、RoleService CRUD(2.9.1-2.9.2)"
```

---

### Task 6: 功能權限矩陣 + GetAbility 表驅動 + RLS data_scope 注入(細部 2.9.3–2.9.5)

**Files:**
- Update: `backend/internal/domain/roles/usecase.go`、`repository.go`、`handler.go`
- Update: `backend/proto/v1/admin.proto`(RoleService 增量)
- Update: `backend/internal/domain/auth/ability.go`(接管 01 計畫 Task 13 的 builtinAbility)
- Update: `backend/internal/middleware/auth.go`(data_scope 改角色表驅動)
- Create: `backend/database/migrations/00006_rls_data_scope.sql`
- Test: `backend/internal/domain/roles/permissions_test.go`、`backend/internal/domain/auth/ability_test.go`(更新)、`backend/internal/middleware/datascope_test.go`

**Interfaces:**
- Consumes: Task 5 `ent.Role` / `ent.RolePermission` / 詞彙表;01 計畫 Task 13 `auth.AbilityRule` / `auth.Can` / AbilityService handler;Task 1 `rls.Identity.Roles` / `HasRole`。
- Produces: proto `RoleService.GetPermissions(role_id)→ RolePermissions{permissions}`、`UpdatePermissions(role_id、permissions[])→ RolePermissions`;`roles.Usecase.{GetPermissions, UpdatePermissions}`;`roles.Repository.PermissionsForRoles(ctx, codes []string) ([]Permission, error)`;`auth.Permission{Resource, Action string}`、`auth.BuildAbilityFromPermissions(id rls.Identity, perms []Permission) []AbilityRule`(取代 `BuildAbility` + `builtinAbility`,序列化格式不變);middleware `resolveIdentity`(g 規則展開 → roles 表 → 最寬 data_scope)。

- [ ] **Step 1: 寫失敗測試(權限矩陣)**

`backend/internal/domain/roles/permissions_test.go`:

```go
package roles_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/roles"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestUpdatePermissionsFullOverwrite(t *testing.T) {
	c := testutil.NewEntClient(t)
	testutil.ApplyMigrationFile(t, c, "00005_roles_seed.sql")
	ctx := context.Background()
	uc := roles.NewUsecase(c, nil, audit.NoopRecorder{})

	custom, _ := uc.Create(ctx, superActor(), roles.CreateInput{Code: "sales_rep", Name: "業務代表", DataScope: "company"})
	// 全量覆寫(2.9.3 步驟 2)
	_, err := uc.UpdatePermissions(ctx, superActor(), custom.ID, []roles.PermissionInput{
		{Resource: "customers", Action: "read"},
		{Resource: "sales_orders", Action: "manage"},
	})
	if err != nil {
		t.Fatalf("update permissions: %v", err)
	}
	perms, err := uc.GetPermissions(ctx, superActor(), custom.ID)
	if err != nil || len(perms) != 2 {
		t.Fatalf("get: %v n=%d", err, len(perms))
	}
	// 再次覆寫:未提交的既有權限被移除
	_, _ = uc.UpdatePermissions(ctx, superActor(), custom.ID, []roles.PermissionInput{
		{Resource: "customers", Action: "read"},
	})
	perms, _ = uc.GetPermissions(ctx, superActor(), custom.ID)
	if len(perms) != 1 {
		t.Fatalf("full overwrite should drop unsubmitted entries: %d", len(perms))
	}
	// 未知 resource / action → invalid_argument
	if _, err := uc.UpdatePermissions(ctx, superActor(), custom.ID, []roles.PermissionInput{
		{Resource: "nope", Action: "read"},
	}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("unknown resource: want invalid_argument, got %v", err)
	}
	// company_admin 可 Get 不可 Update(2.9.3 步驟 1)
	admin := rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(), DataScope: "company", Role: "company_admin"}
	if _, err := uc.GetPermissions(ctx, admin, custom.ID); err != nil {
		t.Fatalf("company_admin get: %v", err)
	}
	if _, err := uc.UpdatePermissions(ctx, admin, custom.ID, nil); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("company_admin update: want permission_denied, got %v", err)
	}
	// role 不存在 → not_found
	if _, err := uc.GetPermissions(ctx, superActor(), uuid.New()); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("missing role: want not_found, got %v", err)
	}
}
```

- [ ] **Step 2: 寫失敗測試(GetAbility 表驅動 + data_scope 解析)**

`backend/internal/domain/auth/ability_test.go` 更新(取代 01 計畫 Task 13 的 `TestBuildAbility*`;`Can` 保留):

```go
package auth_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/auth"
)

func TestBuildAbilityFromPermissions(t *testing.T) {
	coID := uuid.New()
	depID := uuid.New()
	rules := auth.BuildAbilityFromPermissions(rls.Identity{
		UserID: uuid.New(), CompanyID: coID, DepartmentID: &depID,
		DataScope: "department", Role: "staff", Roles: []string{"staff"},
	}, []auth.Permission{
		{Resource: "customers", Action: "read"},
		{Resource: "sales_orders", Action: "manage"},
	})
	if !auth.Can(rules, "read", "customers") || !auth.Can(rules, "create", "sales_orders") {
		t.Fatal("table-driven rules not reflected")
	}
	if auth.Can(rules, "delete", "users") {
		t.Fatal("unspecified permission must not be granted")
	}
	// 多角色聯集由呼叫方(Repository)合併,此處驗證 conditions 與主帳號反向禁止保留
	primary := auth.BuildAbilityFromPermissions(rls.Identity{
		UserID: uuid.New(), CompanyID: coID, DataScope: "self", Role: "customer", IsPrimary: true,
	}, []auth.Permission{{Resource: "sales_orders", Action: "read"}})
	if auth.Can(primary, "create", "sales_orders") {
		t.Fatal("primary account must not create orders")
	}
	if !auth.Can(primary, "manage", "account") {
		t.Fatal("primary account should manage sub-accounts")
	}
	// developer:不查表直接 manage all(D8,2.9.4 步驟 3)
	dev := auth.BuildAbilityFromPermissions(rls.Identity{Role: "developer", Roles: []string{"developer"}}, nil)
	if !auth.Can(dev, "delete", "users") {
		t.Fatal("developer should manage all")
	}
	// 無權限(guest)→ 空規則
	if got := auth.BuildAbilityFromPermissions(rls.Identity{Role: "guest"}, nil); len(got) != 0 {
		t.Fatalf("guest should be empty, got %d", len(got))
	}
}
```

`backend/internal/middleware/datascope_test.go`:

```go
package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/middleware"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

// TestDataScopeFromRolesTable:自訂角色指定 data_scope 後,不改程式即獲得對應範圍(2.9.5 驗收 2)。
func TestDataScopeFromRolesTable(t *testing.T) {
	c := testutil.NewEntClient(t)
	testutil.ApplyMigrationFile(t, c, "00005_roles_seed.sql")
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").Save(ctx)
	custom, _ := c.Role.Create().SetCode("sales_rep").SetName("業務代表").
		SetDataScope("company").SetIsSystem(false).Save(ctx)
	u, _ := c.User.Create().SetCompanyID(co.ID).SetRole("staff").SetDataScope("department").
		SetStatus("active").SetEmail("r@example.com").Save(ctx)

	enf, err := authz.NewEnforcer(testutil.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	// g 規則:使用者 → 自訂角色 @ 公司 domain
	if _, err := enf.AddGroupingPolicy(u.ID.String(), custom.Code, co.ID.String()); err != nil {
		t.Fatal(err)
	}
	_ = custom

	var gotScope string
	probe := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := rlsFromCtx(r)
		gotScope = id.DataScope
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.Authenticate(c, nil, enf)(probe)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, bearerReq(t, u))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if gotScope != "company" { // staff(department)+ sales_rep(company)→ 取最寬 = company
		t.Fatalf("want widest scope company, got %q", gotScope)
	}
}

// TestNoRoleFailsStrict:無任何 g 規則且 roles 表查無 → data_scope=self(2.9.5 步驟 4)。
func TestNoRoleFailsStrict(t *testing.T) {
	c := testutil.NewEntClient(t)
	testutil.ApplyMigrationFile(t, c, "00005_roles_seed.sql")
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").Save(ctx)
	u, _ := c.User.Create().SetCompanyID(co.ID).SetRole("staff").SetDataScope("company").
		SetStatus("active").SetEmail("n@example.com").Save(ctx)
	// roles 表刪除 staff 列,模擬「查無 data_scope 的異常身分」
	c.Role.Delete().Where(/* code = staff */).ExecX(ctx)

	enf, _ := authz.NewEnforcer(testutil.DSN(t)) // 無 g 規則
	var gotScope string
	probe := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := rlsFromCtx(r)
		gotScope = id.DataScope
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.Authenticate(c, nil, enf)(probe)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, bearerReq(t, u))
	if gotScope != "self" {
		t.Fatalf("unknown identity should fall back to self, got %q", gotScope)
	}
}
```

(`rlsFromCtx` 為既有測試輔助或就地 `rls.FromContext`;舊測試檔的 probe 可複用。)

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/roles/ ./internal/domain/auth/ ./internal/middleware/ -v`
Expected: FAIL — `UpdatePermissions` / `BuildAbilityFromPermissions` 未定義、middleware 尚未從 roles 表解析。

- [ ] **Step 4: 實作**

`roles/usecase.go` 追加(2.9.3):

```go
type PermissionInput struct{ Resource, Action string }

// GetPermissions 回傳角色全部 resource × action;super / company_admin 可讀(2.9.3 步驟 1)。
func (u *Usecase) GetPermissions(ctx context.Context, actor rls.Identity, roleID uuid.UUID) ([]*ent.RolePermission, error) {
	if _, err := u.Get(ctx, actor, roleID); err != nil { // 內含權限 + not_found
		return nil, err
	}
	perms, err := u.db.RolePermission.Query().
		Where(entrp.RoleIDEQ(roleID)).Order(ent.Asc(entrp.FieldResource, entrp.FieldAction)).All(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return perms, nil
}

// UpdatePermissions 全量覆寫(2.9.3 步驟 2):同交易刪舊插新 + 稽核(前後快照)。
// 內建角色可調(D9);防鎖死屬 Casbin policy 層(2.10.3),本表僅驅動 CASL UI,不額外限制清空。
func (u *Usecase) UpdatePermissions(ctx context.Context, actor rls.Identity, roleID uuid.UUID, perms []PermissionInput) ([]*ent.RolePermission, error) {
	if !actor.HasRole("super") {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("only super can edit permissions"))
	}
	r, err := u.db.Role.Query().Where(entrole.IDEQ(roleID), entrole.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("role not found"))
	}
	for _, p := range perms {
		if !ValidResources[p.Resource] || !ValidActions[p.Action] {
			return nil, connect.NewError(connect.CodeInvalidArgument,
				fmt.Errorf("unknown permission %s/%s", p.Resource, p.Action))
		}
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	before, _ := tx.RolePermission.Query().Where(entrp.RoleIDEQ(roleID)).All(ctx)
	if _, err := tx.RolePermission.Delete().Where(entrp.RoleIDEQ(roleID)).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	bulk := make([]*ent.RolePermissionCreate, 0, len(perms))
	for _, p := range perms {
		bulk = append(bulk, tx.RolePermission.Create().
			SetRoleID(roleID).SetResource(p.Resource).SetAction(p.Action))
	}
	if len(bulk) > 0 {
		if err := tx.RolePermission.CreateBulk(bulk...).Exec(ctx); err != nil {
			_ = tx.Rollback()
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	after, _ := tx.RolePermission.Query().Where(entrp.RoleIDEQ(roleID)).All(ctx)
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "update",
		ResourceType: "role", ResourceID: roleID.String(),
		Before: snapshotPerms(before), After: snapshotPerms(after),
	})
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 生效路徑:GetAbility 每次呼叫即查表(2.9.4),前端 60s TTL 或主動重載(規格 3.4),無需重啟。
	_ = r
	return after, nil
}
```

`roles/repository.go` 追加(GetAbility 用,2.9.4 步驟 1–2;不過濾 `is_active` — 本 RPC 只反映功能權限矩陣):

```go
// PermissionsForRoles 取多角色權限聯集(供 GetAbility);角色已停用不影響(2.9.4 步驟 2)。
func (u *Usecase) PermissionsForRoles(ctx context.Context, codes []string) ([]*ent.RolePermission, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	return u.db.RolePermission.Query().Where(
		entrp.HasRoleWith(entrole.CodeIn(codes...), entrole.DeletedAtIsNil()),
	).All(ctx)
}
```

`auth/ability.go` 改寫(刪除 `builtinAbility` 與 `BuildAbility`;`AbilityRule` / `Can` 保留,契約不變):

```go
package auth

import (
	"slices"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// Permission 為 role_permissions 的一列(CASL subject × action)。
type Permission struct{ Resource, Action string }

// BuildAbilityFromPermissions 依 role_permissions 產生 CASL 規則(細部 2.9.4,
// 接管 01 計畫 Task 13 的 builtinAbility;序列化格式與 Phase 1 契約完全一致)。
func BuildAbilityFromPermissions(id rls.Identity, perms []Permission) []AbilityRule {
	// developer 且開關啟用(開關關閉時 middleware 已拒絕,不會到此):manage all,不查表(D8)
	if id.Role == "developer" || slices.Contains(id.Roles, "developer") {
		return []AbilityRule{{Action: "manage", Subject: "all"}}
	}
	out := make([]AbilityRule, 0, len(perms)+4)
	conds := map[string]string{"company_id": id.CompanyID.String()}
	if id.DepartmentID != nil {
		conds["department_id"] = id.DepartmentID.String()
	}
	if id.CustomerID != nil {
		conds["customer_id"] = id.CustomerID.String()
	}
	for _, p := range perms {
		r := AbilityRule{Action: p.Action, Subject: p.Resource}
		if r.Subject != "all" {
			r.Conditions = conds
		}
		out = append(out, r)
	}
	if id.IsPrimary {
		// 主帳號:業務主題全部 cannot,僅開帳號管理(語義保留自 Task 13)
		for _, s := range []string{"sales_orders", "return_requests", "customer_products"} {
			out = append(out, AbilityRule{Action: "manage", Subject: s, Inverted: true})
		}
		out = append(out, AbilityRule{Action: "manage", Subject: "account"})
	}
	return out
}
```

AbilityService handler(`ability.go` 同檔或既有 handler 處)改:

```go
// GetAbility:角色自 ctx identity(Roles 由 middleware 自 g 規則展開,Task 6);
// 無任何角色(異常狀態)→ 空規則,不視為錯誤(2.9.4 錯誤表)。唯讀,不寫稽核。
func (h *AbilityHandler) GetAbility(ctx context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[v1.GetAbilityResponse], error) {
	id, ok := rls.FromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("no identity"))
	}
	codes := id.Roles
	if len(codes) == 0 && id.Role != "" {
		codes = []string{id.Role}
	}
	rows, err := h.roles.PermissionsForRoles(ctx, codes)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	perms := make([]auth.Permission, 0, len(rows))
	for _, r := range rows {
		perms = append(perms, auth.Permission{Resource: r.Resource, Action: r.Action})
	}
	rules := auth.BuildAbilityFromPermissions(id, perms)
	// ... 映射 v1.AbilityRule;回應加 Cache-Control: private, max-age=60(不變)
}
```

`middleware/auth.go` — `identityFromUser` 改為 `resolveIdentity`(2.9.5 步驟 1、4、5):

```go
// scopeRank:all > company > department > self;多角色取最寬(2.9.5 步驟 1)。
func widestScope(scopes []string) string {
	rank := map[string]int{"self": 0, "department": 1, "company": 2, "all": 3}
	best := "self" // 無角色或查無 → 最嚴,寧可過濾過嚴不可洩漏(步驟 4)
	for _, s := range scopes {
		if rank[s] > rank[best] {
			best = s
		}
	}
	return best
}

// resolveIdentity:使用者 → g 規則展開角色 → roles 表取 data_scope 最寬等級。
// users.data_scope 欄位自此降為建立時初始值,身分解析以角色表為準(D3:不依角色名稱)。
func resolveIdentity(ctx context.Context, db *ent.Client, enf *casbin.Enforcer, u *ent.User) rls.Identity {
	codes := []string{}
	if enf != nil {
		codes, _ = enf.GetRolesForUser(u.ID.String(), u.CompanyID.String())
	}
	if len(codes) == 0 && u.Role != "" {
		codes = []string{u.Role} // 無 g 規則(舊帳號)回落主角色
	}
	scopes := []string{}
	if rs, err := db.Role.Query().Where(
		entrole.CodeIn(codes...), entrole.DeletedAtIsNil()).All(ctx); err == nil {
		for _, r := range rs {
			scopes = append(scopes, r.DataScope)
		}
	}
	scope := widestScope(scopes)
	if u.Role == "developer" {
		scope = "all" // D8;DeveloperBypass 另作開關把關
	}
	return rls.Identity{
		UserID: u.ID, CompanyID: u.CompanyID, DepartmentID: u.DepartmentID,
		CustomerID: u.CustomerID, DataScope: scope,
		Role: u.Role, IsPrimary: u.IsPrimary, Roles: codes,
	}
}
```

(`Authenticate` 內 `rls.NewContext(ctx, identityFromUser(u))` 改呼叫 `resolveIdentity(ctx, db, enf, u)`;import 補 `entrole`。)

`backend/database/migrations/00006_rls_data_scope.sql`(統一四級分支,補 `WITH CHECK` 保證寫入同受約束;行為與 00003 相容,語意明確化):

```sql
-- +goose Up
-- 細部文件 2.9.5:RLS policy 統一依 app.current_data_scope 四級分支改寫。
-- data_scope 值由 middleware 自角色表注入(不依角色名稱,D3);
-- 未注入時 current_setting(..., true) 回 NULL → fail-closed(0 列)。

DROP POLICY IF EXISTS tenant_isolation ON companies;
DROP POLICY IF EXISTS tenant_isolation ON departments;
DROP POLICY IF EXISTS department_scope ON departments;
DROP POLICY IF EXISTS tenant_isolation ON users;
DROP POLICY IF EXISTS data_scope ON users;

-- companies:all 全見;其餘等級僅自己公司(客戶登入頁需讀自己公司顯示資料)
CREATE POLICY tenant_isolation ON companies
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR id::text = current_setting('app.current_company_id', true))
  WITH CHECK (current_setting('app.current_data_scope', true) = 'all'
         OR id::text = current_setting('app.current_company_id', true));

-- departments:all 全見;company 見全公司;department 僅自己部門;self 不可見
CREATE POLICY tenant_isolation ON departments
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true))
  WITH CHECK (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true));

CREATE POLICY department_scope ON departments
  USING (current_setting('app.current_data_scope', true) IN ('all', 'company')
         OR (current_setting('app.current_data_scope', true) = 'department'
             AND id::text = current_setting('app.current_department_id', true)))
  WITH CHECK (current_setting('app.current_data_scope', true) IN ('all', 'company')
         OR (current_setting('app.current_data_scope', true) = 'department'
             AND id::text = current_setting('app.current_department_id', true)));

-- users:先比對公司,再按 data_scope 收窄;self(客戶帳號)僅見同客戶帳號
CREATE POLICY tenant_isolation ON users
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true))
  WITH CHECK (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true));

CREATE POLICY data_scope ON users
  USING (current_setting('app.current_data_scope', true) IN ('all', 'company')
         OR (current_setting('app.current_data_scope', true) = 'department'
             AND department_id::text = current_setting('app.current_department_id', true))
         OR (current_setting('app.current_data_scope', true) = 'self'
             AND customer_id::text = current_setting('app.current_customer_id', true)))
  WITH CHECK (current_setting('app.current_data_scope', true) IN ('all', 'company')
         OR (current_setting('app.current_data_scope', true) = 'department'
             AND department_id::text = current_setting('app.current_department_id', true))
         OR (current_setting('app.current_data_scope', true) = 'self'
             AND customer_id::text = current_setting('app.current_customer_id', true)));

-- +goose Down
-- 還原為 00003 版本(同語句,無 WITH CHECK);略 — 執行 00003 的 Up 內容重建即可。
DROP POLICY IF EXISTS tenant_isolation ON companies;
DROP POLICY IF EXISTS tenant_isolation ON departments;
DROP POLICY IF EXISTS department_scope ON departments;
DROP POLICY IF EXISTS tenant_isolation ON users;
DROP POLICY IF EXISTS data_scope ON users;
```

RLS 四等級整合測試(2.9.5 驗收 1、4;放 `backend/internal/database/rls_datascope_test.go`,沿用 `testutil.NewEntClientWithRLS` + 套用 00006):all 全見 / company 限同公司 / department 限同部門 / self 僅本人;連續不同身分請求無 session 變數殘留(01 計畫 `TestRLSConnectionNoLeak` 已覆蓋機制,本測試補四等級結果集)。

- [ ] **Step 5: 跑測試確認通過**

Run: `cd backend && buf generate && go test ./internal/domain/roles/ ./internal/domain/auth/ ./internal/middleware/ ./internal/database/ -v`
Expected: PASS — 權限矩陣全量覆寫與詞彙驗證、表驅動 ability(含 developer/guest/主帳號)、data_scope 角色表解析與 fail-strict、RLS 四等級;01 計畫既有測試全綠(`TestBuildAbility*` 已由新測試取代,舊 `BuildAbility` 呼叫點全數清除)。

- [ ] **Step 6: Commit**

```bash
git add backend/internal/domain/roles backend/proto/v1/admin.proto backend/internal/domain/auth/ability.go backend/internal/middleware backend/database/migrations/00006_rls_data_scope.sql backend/internal/database
git commit -m "feat(backend): 功能權限矩陣、GetAbility 表驅動、RLS data_scope 角色等級注入(2.9.3-2.9.5)"
```

---

### Task 7: Casbin policy 管理 API + 預設 p 規則 seed + 防鎖死 + ListGrouping(細部 2.10.1–2.10.4)

**Files:**
- Create: `backend/database/migrations/00007_casbin_policies_seed.sql`
- Update: `backend/internal/authz/seed.go`(程式內 seeder 改驗證-only)
- Update: `backend/cmd/seed/main.go`(seeder 呼叫點)、`backend/internal/server/domains.go`(interceptor 註冊)
- Create: `backend/internal/middleware/authz_interceptor.go`
- Create: `backend/internal/domain/policies/usecase.go`
- Create: `backend/internal/domain/policies/repository.go`
- Create: `backend/internal/domain/policies/handler.go`
- Create: `backend/internal/domain/policies/reload.go`
- Update: `backend/proto/v1/admin.proto`(PolicyService 增量)
- Test: `backend/internal/domain/policies/usecase_test.go`、`backend/internal/middleware/authz_interceptor_test.go`

**Interfaces:**
- Consumes: 01 計畫 Task 2 `authz.NewEnforcer` / casbin model(matcher 的 `keyMatch2(r.obj, p.obj)` 支援 path 詞彙);Task 3 `database.ExecTx / QueryTx`;Task 5 roles 表;Task 6 `rls.Identity.Roles`;config `ValkeyAddr`(01 計畫 Task 9)。
- Produces: proto `PolicyService` — `ListPolicies(page、role、domain)→ ListPoliciesResponse{policies, meta}`;`AddPolicy(role、domain、path、action)→ PolicyRule`;`DeletePolicy(role、domain、path、action)→ Empty`(四元組定位,`casbin_rule` 無 id 欄);`ListGrouping(page、user、role、domain)→ ListGroupingResponse{groupings, meta}`。`middleware.AuthzInterceptor(enf) connect.Interceptor`(接手 01 計畫「Casbin 於 handler 層接上 → Phase 2 Task 2.10」佔位);`policies.Reloader`(Valkey pub/sub 跨 replica 重載);`authz.VerifyDefaultPolicies(ctx, e) error`(取代 `SeedDefaultPolicies`,見 Step 1)。
- p 規則詞彙(2.10.1 介面):`ptype=p`、`v0`=role_code、`v1`=domain(預設 `*`;公司特定規則為 company_id 字串)、`v2`=資源路徑 — Connect 端點用 `/{package}.{Service}/{Method}`(如 `/salesorder.v1.UserService/ListUsers`),REST 端點用 URL path;`v3`=動作 — Connect 恆 `POST`,REST 用 HTTP method。

- [ ] **Step 1: 寫失敗測試(seed 內容與冪等)**

`backend/internal/domain/policies/seed_test.go`:

```go
package policies_test

import (
	"context"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestPolicySeed(t *testing.T) {
	c := testutil.NewEntClient(t)
	_ = c
	dsn := testutil.DSN(t)
	testutil.ApplyMigrationFile(t, c, "00007_casbin_policies_seed.sql")

	e, err := authz.NewEnforcer(dsn)
	if err != nil {
		t.Fatal(err)
	}
	// 舊詞彙(resource 名稱)p 規則已清除(2.10.1 步驟:取代程式內 seeder)
	for _, p := range e.GetPolicy() {
		if len(p) > 2 && p[2] != "*" && p[2] != "" && p[2][0] != '/' {
			t.Fatalf("legacy resource-vocab rule survived: %v", p)
		}
	}
	cases := []struct {
		name               string
		sub, dom, obj, act string
		want               bool
	}{
		{"company_admin 管自己公司使用者", "company_admin", "co-1", "/salesorder.v1.UserService/ListUsers", "POST", true},
		{"company_admin 不可建公司", "company_admin", "co-1", "/salesorder.v1.CompanyService/CreateCompany", "POST", false},
		{"company_admin 管自己 domain policy", "company_admin", "co-1", "/salesorder.v1.PolicyService/AddPolicy", "POST", true},
		{"dept_admin 可看使用者", "dept_admin", "co-1", "/salesorder.v1.UserService/GetUser", "POST", true},
		{"dept_admin 不可碰 policy", "dept_admin", "co-1", "/salesorder.v1.PolicyService/AddPolicy", "POST", false},
		{"staff 僅 ability", "staff", "co-1", "/salesorder.v1.AbilityService/GetAbility", "POST", true},
		{"staff 不可管使用者", "staff", "co-1", "/salesorder.v1.UserService/Deactivate", "POST", false},
		{"super 全域", "super", "co-9", "/salesorder.v1.PolicyService/DeletePolicy", "POST", true},
		{"guest 僅 ability", "guest", "co-1", "/salesorder.v1.AbilityService/GetAbility", "POST", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := e.Enforce(tc.sub, tc.dom, tc.obj, tc.act)
			if err != nil || ok != tc.want {
				t.Fatalf("Enforce(%s) = %v, %v; want %v", tc.obj, ok, err, tc.want)
			}
		})
	}
	// 冪等:重複執行不產生重複規則(2.10.1 驗收 3)
	n := len(e.GetPolicy())
	testutil.ApplyMigrationFile(t, c, "00007_casbin_policies_seed.sql")
	e2, _ := authz.NewEnforcer(dsn)
	if got := len(e2.GetPolicy()); got != n {
		t.Fatalf("re-seed not idempotent: %d -> %d", n, got)
	}
}
```

- [ ] **Step 2: 寫失敗測試(CRUD 即時生效、範圍、防鎖死、ListGrouping)**

`backend/internal/domain/policies/usecase_test.go`:

```go
package policies_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/policies"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func setup(t *testing.T) (*policies.Usecase, *casbinEnforcer, uuid.UUID) {
	t.Helper()
	c := testutil.NewEntClient(t)
	testutil.ApplyMigrationFile(t, c, "00005_roles_seed.sql")
	testutil.ApplyMigrationFile(t, c, "00007_casbin_policies_seed.sql")
	enf, err := authz.NewEnforcer(testutil.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	return policies.NewUsecase(c, enf, nil, audit.NoopRecorder{}), enf, uuid.New()
}

func TestAddDeletePolicyImmediateEffect(t *testing.T) {
	uc, enf, coID := setup(t)
	ctx := context.Background()
	admin := rls.Identity{UserID: uuid.New(), CompanyID: coID, DataScope: "company", Role: "company_admin"}

	obj := "/salesorder.v1.CompanyService/UpdateBranding"
	// 新增前:staff 被拒
	if ok, _ := enf.Enforce("staff", coID.String(), obj, "POST"); ok {
		t.Fatal("precondition: staff should be denied")
	}
	// 新增 → 下一次 Enforce 即放行(2.10.2 驗收 1,不需重啟)
	if _, err := uc.AddPolicy(ctx, admin, policies.AddInput{
		Role: "staff", Domain: coID.String(), Path: obj, Action: "POST",
	}); err != nil {
		t.Fatalf("add: %v", err)
	}
	if ok, _ := enf.Enforce("staff", coID.String(), obj, "POST"); !ok {
		t.Fatal("policy should take effect immediately")
	}
	// 重複新增 → already_exists
	if _, err := uc.AddPolicy(ctx, admin, policies.AddInput{
		Role: "staff", Domain: coID.String(), Path: obj, Action: "POST",
	}); connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("dup: want already_exists, got %v", err)
	}
	// 刪除 → 即被拒
	if err := uc.DeletePolicy(ctx, admin, policies.DeleteInput{
		Role: "staff", Domain: coID.String(), Path: obj, Action: "POST",
	}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if ok, _ := enf.Enforce("staff", coID.String(), obj, "POST"); ok {
		t.Fatal("deleted policy should deny immediately")
	}
	// 刪除不存在 → not_found
	if err := uc.DeletePolicy(ctx, admin, policies.DeleteInput{
		Role: "staff", Domain: coID.String(), Path: obj, Action: "POST",
	}); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("delete missing: want not_found, got %v", err)
	}
}

func TestPolicyValidationAndScope(t *testing.T) {
	uc, _, coID := setup(t)
	ctx := context.Background()
	admin := rls.Identity{UserID: uuid.New(), CompanyID: coID, DataScope: "company", Role: "company_admin"}

	// role 不存在 → invalid_argument
	if _, err := uc.AddPolicy(ctx, admin, policies.AddInput{
		Role: "nope", Domain: coID.String(), Path: "/salesorder.v1.UserService/GetUser", Action: "POST",
	}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("bad role: want invalid_argument, got %v", err)
	}
	// 路徑格式非法 → invalid_argument
	if _, err := uc.AddPolicy(ctx, admin, policies.AddInput{
		Role: "staff", Domain: coID.String(), Path: "users", Action: "POST",
	}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("bad path: want invalid_argument, got %v", err)
	}
	// Connect 端點動作非 POST → invalid_argument
	if _, err := uc.AddPolicy(ctx, admin, policies.AddInput{
		Role: "staff", Domain: coID.String(), Path: "/salesorder.v1.UserService/GetUser", Action: "GET",
	}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("bad action: want invalid_argument, got %v", err)
	}
	// company_admin 對他公司 domain → permission_denied(2.10.3 步驟 1b)
	if _, err := uc.AddPolicy(ctx, admin, policies.AddInput{
		Role: "staff", Domain: uuid.New().String(), Path: "/salesorder.v1.UserService/GetUser", Action: "POST",
	}); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("cross-domain add: want permission_denied, got %v", err)
	}
	// List 強制過濾自己 domain
	list, _, err := uc.ListPolicies(ctx, admin, policies.ListQuery{Page: 1, PerPage: 100})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range list {
		if p.Domain != coID.String() {
			t.Fatalf("company_admin saw foreign domain rule: %+v", p)
		}
	}
	// staff 不開放
	staff := rls.Identity{UserID: uuid.New(), CompanyID: coID, DataScope: "department", Role: "staff"}
	if _, _, err := uc.ListPolicies(ctx, staff, policies.ListQuery{Page: 1, PerPage: 20}); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("staff list: want permission_denied, got %v", err)
	}
}

func TestAntiLockout(t *testing.T) {
	uc, _, coID := setup(t)
	ctx := context.Background()
	admin := rls.Identity{UserID: uuid.New(), CompanyID: coID, DataScope: "company", Role: "company_admin"}

	// company_admin 於自己 domain 新增唯一一條 PolicyService 管理規則後,刪除它 → failed_precondition(2.10.3)
	mgmt := policies.AddInput{
		Role: "company_admin", Domain: coID.String(),
		Path: "/salesorder.v1.PolicyService/AddPolicy", Action: "POST",
	}
	// seed 已有 company_admin dom=* 的 PolicyService/* 全域規則,先模擬「該 domain 唯一一條」:
	// 直接以 super 角色驗證全域保護:刪除 super 的 /* 規則 → failed_precondition(2.10.3 步驟 3)
	superActor := rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(), DataScope: "all", Role: "super"}
	if err := uc.DeletePolicy(ctx, superActor, policies.DeleteInput{
		Role: "super", Domain: "*", Path: "/*", Action: ".*",
	}); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("super lockout: want failed_precondition, got %v", err)
	}
	_ = mgmt
}

func TestListGrouping(t *testing.T) {
	uc, enf, coID := setup(t)
	ctx := context.Background()
	// 建 g 規則:使用者 → staff @ coID
	userID := uuid.New()
	if _, err := enf.AddGroupingPolicy(userID.String(), "staff", coID.String()); err != nil {
		t.Fatal(err)
	}
	superActor := rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(), DataScope: "all", Role: "super"}
	rows, total, err := uc.ListGrouping(ctx, superActor, policies.GroupingQuery{Page: 1, PerPage: 20, Role: "staff"})
	if err != nil || total == 0 {
		t.Fatalf("list grouping: %v total=%d", err, total)
	}
	found := false
	for _, g := range rows {
		if g.UserID == userID.String() && g.Role == "staff" && g.Domain == coID.String() {
			found = true
		}
	}
	if !found {
		t.Fatal("grouping rule not visible")
	}
	// company_admin 僅見自己 domain
	admin := rls.Identity{UserID: uuid.New(), CompanyID: coID, DataScope: "company", Role: "company_admin"}
	if _, _, err := uc.ListGrouping(ctx, admin, policies.GroupingQuery{Page: 1, PerPage: 20, Domain: uuid.New().String()}); err != nil {
		// domain 參數被忽略(強制自己公司),不應見到他公司規則;此處僅驗證不報錯
		t.Fatal(err)
	}
}
```

`backend/internal/middleware/authz_interceptor_test.go`:

```go
package middleware_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/middleware"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestAuthzInterceptor(t *testing.T) {
	c := testutil.NewEntClient(t)
	testutil.ApplyMigrationFile(t, c, "00007_casbin_policies_seed.sql")
	enf, _ := authz.NewEnforcer(testutil.DSN(t))
	coID := uuid.New()

	call := func(id rls.Identity, procedure string) error {
		interceptor := middleware.AuthzInterceptor(enf)
		unary := interceptor.WrapUnary(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			return nil, nil // 放行即成功
		})
		ctx := rls.NewContext(context.Background(), id)
		req := &fakeRequest{procedure: procedure}
		_, err := unary(ctx, req)
		return err
	}

	staff := rls.Identity{UserID: uuid.New(), CompanyID: coID, DataScope: "department", Role: "staff", Roles: []string{"staff"}}
	if err := call(staff, "/salesorder.v1.AbilityService/GetAbility"); err != nil {
		t.Fatalf("staff GetAbility should pass: %v", err)
	}
	if err := call(staff, "/salesorder.v1.UserService/Deactivate"); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("staff Deactivate: want permission_denied, got %v", err)
	}
	// developer 繞過(DataScope 已經 DeveloperBypass 升 all)
	dev := rls.Identity{UserID: uuid.New(), CompanyID: coID, DataScope: "all", Role: "developer"}
	if err := call(dev, "/salesorder.v1.PolicyService/DeletePolicy"); err != nil {
		t.Fatalf("developer bypass: %v", err)
	}
	// 無身分 → unauthenticated
	interceptor := middleware.AuthzInterceptor(enf)
	unary := interceptor.WrapUnary(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) { return nil, nil })
	if _, err := unary(context.Background(), &fakeRequest{procedure: "/salesorder.v1.UserService/GetUser"}); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("no identity: want unauthenticated, got %v", err)
	}
}
```

(`fakeRequest` 實作 `connect.AnyRequest`:`Spec()` 回 `connect.Spec{Procedure: ...}`,其餘方法回零值。)

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/policies/ ./internal/middleware/ -v`
Expected: FAIL — `policies.NewUsecase` / `middleware.AuthzInterceptor` 未定義;migration 00007 不存在。

- [ ] **Step 4: 實作 migration、驗證-only seeder、interceptor、policies domain**

`backend/database/migrations/00007_casbin_policies_seed.sql`:

```sql
-- +goose Up
-- 細部文件 2.10.1:預設 p 規則改為 method path 詞彙,取代 01 計畫 Task 2 程式內 seeder。
-- v0=role_code,v1=domain(預設 '*',公司特定規則由 Web 以 company_id 新增),
-- v2=Connect method path 或 REST path,v3=動作(Connect 恆 POST)。
-- 冪等:同四元組已存在則略過,不覆寫 Web 調整結果。

-- 清除 01 計畫程式內 seeder 的舊 resource 詞彙規則(v2 非 path 者)
DELETE FROM casbin_rule WHERE ptype = 'p' AND v2 NOT LIKE '/%';

INSERT INTO casbin_rule (ptype, v0, v1, v2, v3)
SELECT 'p', 'super', '*', '/*', '.*'
WHERE NOT EXISTS (SELECT 1 FROM casbin_rule WHERE ptype='p' AND v0='super' AND v1='*' AND v2='/*' AND v3='.*');

-- company_admin:使用者/部門全管、公司讀與品牌公開資訊更新、角色與權限查閱、自己 domain 的 policy 管理
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3)
SELECT 'p', 'company_admin', '*', v.path, 'POST'
FROM (VALUES
  ('/salesorder.v1.UserService/*'),
  ('/salesorder.v1.DepartmentService/*'),
  ('/salesorder.v1.CompanyService/ListCompanies'),
  ('/salesorder.v1.CompanyService/GetCompany'),
  ('/salesorder.v1.CompanyService/UpdateCompany'),
  ('/salesorder.v1.CompanyService/UpdateBranding'),
  ('/salesorder.v1.CompanyService/UpdatePublicInfo'),
  ('/salesorder.v1.RoleService/ListRoles'),
  ('/salesorder.v1.RoleService/GetRole'),
  ('/salesorder.v1.RoleService/GetPermissions'),
  ('/salesorder.v1.PolicyService/*'),
  ('/salesorder.v1.AbilityService/GetAbility')) AS v(path)
WHERE NOT EXISTS (SELECT 1 FROM casbin_rule WHERE ptype='p' AND v0='company_admin' AND v1='*' AND v2=v.path AND v3='POST');

-- dept_admin:部門 staff 管理(List/Get/Update/AssignRole/Deactivate/ForceLogout)
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3)
SELECT 'p', 'dept_admin', '*', v.path, 'POST'
FROM (VALUES
  ('/salesorder.v1.UserService/ListUsers'),
  ('/salesorder.v1.UserService/GetUser'),
  ('/salesorder.v1.UserService/UpdateUser'),
  ('/salesorder.v1.UserService/AssignRole'),
  ('/salesorder.v1.UserService/Deactivate'),
  ('/salesorder.v1.UserService/ForceLogout'),
  ('/salesorder.v1.AbilityService/GetAbility')) AS v(path)
WHERE NOT EXISTS (SELECT 1 FROM casbin_rule WHERE ptype='p' AND v0='dept_admin' AND v1='*' AND v2=v.path AND v3='POST');

-- staff / customer / guest:僅 ability(業務 service 規則由各 Phase 計畫隨 service 落地補 seed)
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3)
SELECT 'p', v.role, '*', '/salesorder.v1.AbilityService/GetAbility', 'POST'
FROM (VALUES ('staff'), ('customer'), ('guest')) AS v(role)
WHERE NOT EXISTS (SELECT 1 FROM casbin_rule WHERE ptype='p' AND v0=v.role AND v1='*'
  AND v2='/salesorder.v1.AbilityService/GetAbility' AND v3='POST');

-- developer 不 seed 任何 p 規則:繞過由 middleware 實現(D8,2.10.1 步驟 3)。

-- +goose Down
DELETE FROM casbin_rule WHERE ptype = 'p' AND v1 = '*';
```

`backend/internal/authz/seed.go` 改寫(2.10.1 步驟 4:單一來源,程式內 seeder 不再寫入;**重新命名** `SeedDefaultPolicies` → `VerifyDefaultPolicies`,`main.go` 呼叫點同步更新 — clean cutover,無舊名殘留):

```go
// VerifyDefaultPolicies 啟動時檢查 migration 00007 的預設規則存在;缺失即報錯提示跑 migration。
// 不再負責寫入(2.10.1 步驟 4,避免 DB migration 與程式雙來源)。
func VerifyDefaultPolicies(_ context.Context, e *casbin.Enforcer) error {
	required := [][]string{
		{"super", "*", "/*", ".*"},
		{"company_admin", "*", "/salesorder.v1.PolicyService/*", "POST"},
		{"staff", "*", "/salesorder.v1.AbilityService/GetAbility", "POST"},
	}
	for _, p := range required {
		has, err := e.HasPolicy(p)
		if err != nil {
			return fmt.Errorf("authz: check policy %v: %w", p, err)
		}
		if !has {
			return fmt.Errorf("authz: default policy %v missing — run migration 00007_casbin_policies_seed.sql", p)
		}
	}
	return nil
}
```

`backend/internal/middleware/authz_interceptor.go`:

```go
package middleware

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/casbin/casbin/v2"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// AuthzInterceptor 於 Connect handler 層執行 Casbin 功能授權(接手 01 計畫 Task 11 佔位,
// 細部 2.10.x 的「即時生效」由此放大鏡生效:enforcer 記憶體更新後下一次呼叫即用新規則)。
// 掛載:僅需認證的 service 群組(connect.WithInterceptors);公開端點(AuthService 登入路徑、
// /api/v1/companies/public)不掛。
func AuthzInterceptor(enf *casbin.Enforcer) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			id, ok := rls.FromContext(ctx)
			if !ok {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("no identity"))
			}
			// developer 繞過(D8;開關把關在 DeveloperBypass,關閉時 developer 根本到不了這裡)
			if id.Role == "developer" && id.DataScope == "all" {
				return next(ctx, req)
			}
			obj := req.Spec().Procedure // 形如 /salesorder.v1.UserService/ListUsers
			dom := id.CompanyID.String()
			roles := append([]string{id.Role}, id.Roles...)
			seen := map[string]bool{}
			for _, role := range roles {
				if role == "" || seen[role] {
					continue
				}
				seen[role] = true
				allowed, err := enf.Enforce(role, dom, obj, "POST")
				if err != nil {
					return nil, connect.NewError(connect.CodeInternal, err)
				}
				if allowed {
					return next(ctx, req)
				}
			}
			return nil, connect.NewError(connect.CodePermissionDenied,
				fmt.Errorf("%s denied for roles %v", obj, roles))
		}
	})
}
```

`backend/internal/domain/policies/usecase.go`:

```go
// Package policies 為 Casbin policy 管理 domain(細部文件 2.10)。
package policies

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"connectrpc.com/connect"
	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entrole "github.com/salesorder/sales-order-1.0/backend/ent/role"
	"github.com/salesorder/sales-order-1.0/backend/internal/audit"
	"github.com/salesorder/sales-order-1.0/backend/internal/database"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// connectPathRe:/{package}.{Service}/{Method};REST path 以 /api/ 前綴另行允許。
var (
	connectPathRe = regexp.MustCompile(`^/[a-z0-9.]+\.[A-Z][A-Za-z0-9]+/[A-Z][A-Za-z0-9]+$|^/[a-z0-9.]+\.[A-Z][A-Za-z0-9]+/\*$`)
	restPathRe    = regexp.MustCompile(`^/api/[a-z0-9/_{}\-]*$|^/\*$`)
)

type Usecase struct {
	db       *ent.Client
	enf      *casbin.Enforcer
	reloader *Reloader // 可為 nil(單體部署)
	audit    audit.Recorder
}

func NewUsecase(db *ent.Client, enf *casbin.Enforcer, rel *Reloader, rec audit.Recorder) *Usecase {
	return &Usecase{db: db, enf: enf, reloader: rel, audit: rec}
}

type PolicyRule struct{ Role, Domain, Path, Action string }
type GroupingRule struct{ UserID, UserName, Role, RoleName, Domain string }

type ListQuery struct {
	Page, PerPage int
	Role, Domain  string
}

type AddInput struct{ Role, Domain, Path, Action string }
type DeleteInput = AddInput

type GroupingQuery struct {
	Page, PerPage          int
	User, Role, Domain     string
}

// checkDomainScope(2.10.3 步驟 1):super 任意 domain;company_admin 僅自己公司;其餘拒絕。
// 回傳強制過濾用的 domain(super 回空字串 = 不過濾)。
func checkDomainScope(actor rls.Identity) (string, error) {
	if actor.HasRole("super") {
		return "", nil
	}
	if actor.HasRole("company_admin") {
		return actor.CompanyID.String(), nil
	}
	return "", connect.NewError(connect.CodePermissionDenied, errors.New("policy management not allowed"))
}

func validateRuleInput(in AddInput) error {
	isConnect := connectPathRe.MatchString(in.Path)
	if !isConnect && !restPathRe.MatchString(in.Path) {
		return connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("bad resource path %q", in.Path))
	}
	if isConnect && in.Action != "POST" {
		return connect.NewError(connect.CodeInvalidArgument,
			errors.New("connect endpoints only allow POST"))
	}
	if !isConnect {
		switch in.Action {
		case "GET", "POST", "PUT", "PATCH", "DELETE", ".*":
		default:
			return connect.NewError(connect.CodeInvalidArgument, errors.New("bad http method"))
		}
	}
	return nil
}
```

`ListPolicies` / `AddPolicy` / `DeletePolicy` / 防鎖死(同檔):

```go
func (u *Usecase) ListPolicies(ctx context.Context, actor rls.Identity, q ListQuery) ([]PolicyRule, int, error) {
	forcedDomain, err := checkDomainScope(actor)
	if err != nil {
		return nil, 0, err
	}
	page, perPage, err := normalizePage(q.Page, q.PerPage)
	if err != nil {
		return nil, 0, err
	}
	// 讀取直查 DB 而非記憶體,保證所見即持久態(2.10.2 步驟 5)
	where := `ptype = 'p'`
	args := []any{}
	domain := q.Domain
	if forcedDomain != "" {
		domain = forcedDomain // company_admin 強制過濾自己公司 domain(2.10.3)
	}
	if domain != "" {
		args = append(args, domain)
		where += fmt.Sprintf(` AND v1 = $%d`, len(args))
	}
	if q.Role != "" {
		args = append(args, q.Role)
		where += fmt.Sprintf(` AND v0 = $%d`, len(args))
	}
	var total int
	if err := u.db.QueryContext(ctx, // repository.go 提供 *sql.DB 直查包裝
		`SELECT count(*) FROM casbin_rule WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, connect.NewError(connect.CodeInternal, err)
	}
	args = append(args, perPage, (page-1)*perPage)
	rows, err := u.db.QueryContext(ctx,
		`SELECT v0, v1, v2, v3 FROM casbin_rule WHERE `+where+
			fmt.Sprintf(` ORDER BY v0, v2 LIMIT $%d OFFSET $%d`, len(args)-1, len(args)), args...)
	if err != nil {
		return nil, 0, connect.NewError(connect.CodeInternal, err)
	}
	defer rows.Close()
	var out []PolicyRule
	for rows.Next() {
		var p PolicyRule
		if err := rows.Scan(&p.Role, &p.Domain, &p.Path, &p.Action); err != nil {
			return nil, 0, connect.NewError(connect.CodeInternal, err)
		}
		out = append(out, p)
	}
	return out, total, nil
}

func (u *Usecase) AddPolicy(ctx context.Context, actor rls.Identity, in AddInput) (*PolicyRule, error) {
	forcedDomain, err := checkDomainScope(actor)
	if err != nil {
		return nil, err
	}
	if forcedDomain != "" && in.Domain != forcedDomain {
		return nil, connect.NewError(connect.CodePermissionDenied,
			errors.New("domain must be your own company"))
	}
	if err := validateRuleInput(in); err != nil {
		return nil, err
	}
	// role 必須存在(2.10.2 步驟 1)
	if exists, err := u.db.Role.Query().Where(
		entrole.CodeEQ(in.Role), entrole.DeletedAtIsNil()).Exist(ctx); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	} else if !exists {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("role %q not found", in.Role))
	}

	tx, err := u.db.Tx(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 查重(同四元組 → already_exists)
	exists, err := existsRule(ctx, tx, in.Role, in.Domain, in.Path, in.Action)
	if err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if exists {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("policy already exists"))
	}
	if err := database.ExecTx(ctx, tx,
		`INSERT INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES ('p', $1, $2, $3, $4)`,
		in.Role, in.Domain, in.Path, in.Action); err != nil {
		_ = tx.Rollback()
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "create",
		ResourceType: "casbin_policy",
		ResourceID:   fmt.Sprintf("%s|%s|%s|%s", in.Role, in.Domain, in.Path, in.Action),
		After:        fmt.Sprintf(`{"role":%q,"domain":%q,"path":%q,"action":%q}`, in.Role, in.Domain, in.Path, in.Action),
	})
	if err := tx.Commit(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 提交成功後更新記憶體態;提交失敗不動記憶體(2.10.2 步驟 2)
	if _, err := u.enf.AddPolicy(in.Role, in.Domain, in.Path, in.Action); err != nil {
		_ = u.enf.LoadPolicy() // adapter 已同事務寫入,AddPolicy 失敗僅為記憶體重複 → 重載對齊
	}
	u.broadcast(ctx)
	rule := PolicyRule(in)
	return &rule, nil
}

// isPolicyManagementRule:規則是否覆蓋 PolicyService 管理能力(防鎖死判斷標的)。
func isPolicyManagementRule(path string) bool {
	return path == "/*" || path == "/salesorder.v1.PolicyService/*" ||
		strings.HasPrefix(path, "/salesorder.v1.PolicyService/")
}

func (u *Usecase) DeletePolicy(ctx context.Context, actor rls.Identity, in DeleteInput) error {
	forcedDomain, err := checkDomainScope(actor)
	if err != nil {
		return err
	}
	if forcedDomain != "" && in.Domain != forcedDomain {
		return connect.NewError(connect.CodePermissionDenied,
			errors.New("domain must be your own company"))
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	exists, err := existsRule(ctx, tx, in.Role, in.Domain, in.Path, in.Action)
	if err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	if !exists {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeNotFound, errors.New("policy not found"))
	}
	// 防鎖死(2.10.3 步驟 2):刪除的規則屬於「操作者自身角色的 policy 管理能力」時,
	// 同事務確認刪後該角色在涵蓋 domain 內仍有至少一條管理規則。
	actorRoles := append([]string{actor.Role}, actor.Roles...)
	isOwnRole := false
	for _, r := range actorRoles {
		if r == in.Role {
			isOwnRole = true
		}
	}
	if isOwnRole && isPolicyManagementRule(in.Path) {
		remaining, err := countManagementRules(ctx, tx, in.Role)
		if err != nil {
			_ = tx.Rollback()
			return connect.NewError(connect.CodeInternal, err)
		}
		if remaining <= 1 {
			_ = tx.Rollback()
			return connect.NewError(connect.CodeFailedPrecondition,
				errors.New("deleting this rule would lock out policy management for your role"))
		}
	}
	if err := database.ExecTx(ctx, tx,
		`DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = $1 AND v1 = $2 AND v2 = $3 AND v3 = $4`,
		in.Role, in.Domain, in.Path, in.Action); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	_ = u.audit.Record(ctx, audit.Entry{
		ActorID: actor.UserID.String(), Action: "delete",
		ResourceType: "casbin_policy",
		ResourceID:   fmt.Sprintf("%s|%s|%s|%s", in.Role, in.Domain, in.Path, in.Action),
		Before:       fmt.Sprintf(`{"role":%q,"domain":%q,"path":%q,"action":%q}`, in.Role, in.Domain, in.Path, in.Action),
	})
	if err := tx.Commit(); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if _, err := u.enf.RemovePolicy(in.Role, in.Domain, in.Path, in.Action); err != nil {
		_ = u.enf.LoadPolicy()
	}
	u.broadcast(ctx)
	return nil
}

// countManagementRules:該角色仍持有的 policy 管理規則數(含 dom='*' 全域規則,因其涵蓋所有 domain)。
func countManagementRules(ctx context.Context, tx *ent.Tx, role string) (int, error) {
	rows, err := database.QueryTx(ctx, tx,
		`SELECT count(*) FROM casbin_rule WHERE ptype = 'p' AND v0 = $1
		   AND (v2 = '/*' OR v2 = '/salesorder.v1.PolicyService/*' OR v2 LIKE '/salesorder.v1.PolicyService/%')`,
		role)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var n int
	if rows.Next() {
		if err := rows.Scan(&n); err != nil {
			return 0, err
		}
	}
	return n, nil
}

func (u *Usecase) broadcast(ctx context.Context) {
	if u.reloader != nil {
		_ = u.reloader.Broadcast(ctx) // 多 replica(2.10.2 步驟 4);單體部署時無害
	}
}
```

`ListGrouping`(2.10.4;唯讀、直查 DB、join 顯示名稱):

```go
func (u *Usecase) ListGrouping(ctx context.Context, actor rls.Identity, q GroupingQuery) ([]GroupingRule, int, error) {
	forcedDomain, err := checkDomainScope(actor)
	if err != nil {
		return nil, 0, err
	}
	page, perPage, err := normalizePage(q.Page, q.PerPage)
	if err != nil {
		return nil, 0, err
	}
	where := `cr.ptype = 'g'`
	args := []any{}
	domain := q.Domain
	if forcedDomain != "" {
		domain = forcedDomain
	}
	if domain != "" {
		args = append(args, domain)
		where += fmt.Sprintf(` AND cr.v2 = $%d`, len(args))
	}
	if q.Role != "" {
		args = append(args, q.Role)
		where += fmt.Sprintf(` AND cr.v1 = $%d`, len(args))
	}
	if q.User != "" {
		args = append(args, q.User)
		where += fmt.Sprintf(` AND cr.v0 = $%d`, len(args))
	}
	var total int
	if err := u.db.QueryContext(ctx,
		`SELECT count(*) FROM casbin_rule cr WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, connect.NewError(connect.CodeInternal, err)
	}
	// join users / roles 取顯示名稱(僅顯示用途,2.10.4 步驟 3)
	args = append(args, perPage, (page-1)*perPage)
	rows, err := u.db.QueryContext(ctx,
		`SELECT cr.v0, coalesce(u2.name, ''), cr.v1, coalesce(r2.name, ''), cr.v2
		   FROM casbin_rule cr
		   LEFT JOIN users u2 ON u2.id::text = cr.v0
		   LEFT JOIN roles r2 ON r2.code = cr.v1 AND r2.deleted_at IS NULL
		  WHERE `+where+fmt.Sprintf(` ORDER BY cr.v1, cr.v0 LIMIT $%d OFFSET $%d`, len(args)-1, len(args)),
		args...)
	if err != nil {
		return nil, 0, connect.NewError(connect.CodeInternal, err)
	}
	defer rows.Close()
	var out []GroupingRule
	for rows.Next() {
		var g GroupingRule
		if err := rows.Scan(&g.UserID, &g.UserName, &g.Role, &g.RoleName, &g.Domain); err != nil {
			return nil, 0, connect.NewError(connect.CodeInternal, err)
		}
		out = append(out, g)
	}
	// g 規則新增/移除統一走 UserService.AssignRole(2.3.1,單一寫入路徑);本 RPC 僅檢視。
	return out, total, nil
}
```

(`existsRule` 與 `u.db.QueryContext` 包裝放 `repository.go`:*ent.Client 無 QueryContext — repository 以 `entsql.OpenDB` 前的 `*sql.DB` 或 `client.DB()` 暴露的底層連線執行;實作時於 `policies.NewUsecase` 額外注入 `*sql.DB`(main.go 組裝時與 Ent 共用同一 pool),usecase struct 加 `sqlDB *sql.DB` 欄位,`QueryContext` 呼叫改 `u.sqlDB.QueryContext`。以此為準。)

`backend/internal/domain/policies/reload.go`:

```go
package policies

import (
	"context"
	"log"

	"github.com/casbin/casbin/v2"
	"github.com/redis/go-redis/v9"
)

const policyChangedChannel = "casbin_policy_changed"

// Reloader 跨 replica 一致性(2.10.2 步驟 4):異動提交後經 Valkey pub/sub 廣播,
// 其他 replica 收到後自 adapter 重新載入;rdb 為 nil 時(單體)全部 no-op。
type Reloader struct {
	enf *casbin.Enforcer
	rdb *redis.Client
}

func NewReloader(enf *casbin.Enforcer, rdb *redis.Client) *Reloader {
	return &Reloader{enf: enf, rdb: rdb}
}

func (r *Reloader) Broadcast(ctx context.Context) error {
	if r.rdb == nil {
		return nil
	}
	return r.rdb.Publish(ctx, policyChangedChannel, "1").Err()
}

// Subscribe 於 main.go 以 goroutine 啟動;ctx 結束即退出。
func (r *Reloader) Subscribe(ctx context.Context) {
	if r.rdb == nil {
		return
	}
	sub := r.rdb.Subscribe(ctx, policyChangedChannel)
	defer sub.Close()
	for msg := range sub.Channel() {
		_ = msg
		if err := r.enf.LoadPolicy(); err != nil {
			log.Printf("policies: reload after broadcast: %v", err)
		}
	}
}
```

`backend/proto/v1/admin.proto` 增量(加入檔案):

```proto
service PolicyService {
  rpc ListPolicies(ListPoliciesRequest) returns (ListPoliciesResponse);
  rpc AddPolicy(AddPolicyRequest) returns (PolicyRule);
  rpc DeletePolicy(DeletePolicyRequest) returns (google.protobuf.Empty);
  rpc ListGrouping(ListGroupingRequest) returns (ListGroupingResponse);
}

message PolicyRule {
  string role = 1;
  string domain = 2;
  string path = 3;
  string action = 4;
}

message ListPoliciesRequest {
  PageRequest page = 1;
  optional string role = 2;
  optional string domain = 3; // company_admin 一律被覆寫為自己公司
}

message ListPoliciesResponse {
  repeated PolicyRule policies = 1;
  PageMeta meta = 2;
}

message AddPolicyRequest {
  string role = 1;
  string domain = 2;
  string path = 3;
  string action = 4;
}

// 四元組定位(casbin_rule 無 id 欄)
message DeletePolicyRequest {
  string role = 1;
  string domain = 2;
  string path = 3;
  string action = 4;
}

message GroupingRule {
  string user_id = 1;
  string user_name = 2;
  string role = 3;
  string role_name = 4;
  string domain = 5;
}

message ListGroupingRequest {
  PageRequest page = 1;
  optional string user = 2;
  optional string role = 3;
  optional string domain = 4;
}

message ListGroupingResponse {
  repeated GroupingRule groupings = 1;
  PageMeta meta = 2;
}
```

`InitDomains()` 組裝增量:

```go
	// seeder 改名(clean cutover):原 authz.SeedDefaultPolicies(ctx, enf) 改為
	if err := authz.VerifyDefaultPolicies(ctx, enf); err != nil {
		log.Fatal(err) // 缺預設規則 = migration 未跑,fail-fast
	}
	// 需認證的 Connect service 群組掛授權攔截器(公開路徑不掛):
	// connect.WithInterceptors(middleware.AuthzInterceptor(enf))
	// Reloader:go policies.NewReloader(enf, rdb).Subscribe(ctx)
```

- [ ] **Step 5: 跑測試確認通過**

Run: `cd backend && buf generate && go test ./internal/domain/policies/ ./internal/middleware/ ./internal/authz/ -v`
Expected: PASS — seed 內容/冪等、即時生效、範圍與驗證、防鎖死、ListGrouping、interceptor 放行/拒絕/developer 繞過;01 計畫 authz 測試中 `SeedDefaultPolicies` 呼叫點已改 `VerifyDefaultPolicies`(其寫入斷言改由 migration seed 測試承接)。

- [ ] **Step 6: 全測試回歸 + Commit**

Run: `cd backend && go test ./...`
Expected: PASS — 全倉含 01 計畫既有測試。

```bash
git add backend/database/migrations/00007_casbin_policies_seed.sql backend/internal/authz backend/internal/middleware/authz_interceptor.go backend/internal/domain/policies backend/proto/v1/admin.proto backend/internal/server
git commit -m "feat(backend): PolicyService CRUD 即時生效、domain 範圍、防鎖死、ListGrouping、授權攔截器(2.10)"
```

---

## Self-Review 記錄

- **Spec 覆蓋**:細部文件 19 子功能 → Task 對應:2.1.1–2.1.3→T1;2.2.1→T2;2.3.1–2.3.3→T3;2.4.1–2.4.3→T4;2.9.1–2.9.2→T5;2.9.3–2.9.5→T6;2.10.1–2.10.4→T7。無缺漏。
- **已知佔位(皆標 TODO + 接手 Task)**:
  - `companies.LogoAssetWriter` / `stubLogoAssetWriter`(2.4.1 的 file_assets 寫入)→ 04-master-data 計畫 子功能 3.6.2 接手,屆時改注入 `fileassets.Service` 並移除 stub,呼叫點不變。
  - `audit.Recorder` 的 DB 實作 → 03-metadicts-audit 計畫 Task 2.6 接手(現為 01 計畫 Task 14 的 `NoopRecorder`,本計畫所有稽核呼叫點已就位)。
  - Task 3 `roleExists` 的內建列舉回落路徑 → 本計畫 Task 5 Step 4 移除(同計畫內消化,不外溢)。
  - 業務 service(SalesOrder 等)的 p 規則 → 各 Phase 計畫隨 service 落地補 seed(本計畫 00007 僅 seed Phase 1–2 已存在 service)。
- **類型一致**:`rls.Identity`(T1 加 `Roles` 欄位與 `HasRole`,T6 middleware 填入)、`users.NewUsecase(db, sm, enf, rec)`(T3 擴充,01 計畫 Task 12 呼叫點同步更新)、`authz.VerifyDefaultPolicies`(T7 重新命名,main.go 同步)、`auth.BuildAbilityFromPermissions`(T6 取代 `BuildAbility`,舊測試同步改寫)、`database.ExecTx/QueryTx`(T3 建立,T5/T7 複用)、`testutil.ApplyMigrationFile`(T5 建立,T6/T7 測試複用)於各 Task 簽名一致。
- **跨檔依賴標註**:公司 status 檢查掛點依 01 計畫 Task 11 middleware(T1 僅加 developer 豁免與登入端點檢查);ForceLogout 機制依 01 計畫 Task 12(T3 僅換範圍函式);GetAbility 序列化契約依 01 計畫 Task 13(T6 僅換規則來源);casbin model 詞彙依 01 計畫 Task 2(`keyMatch2` 支援 path 詞彙,T7 不改 model)。

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-08-17-backend-02-tenancy-users-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** — 每個 Task 派新 subagent 執行,Task 間 review,迭代快。

**2. Inline Execution** — 用 executing-plans 在本 session 逐批執行,設 checkpoint review。

Which approach?

---

*計畫版本:v1.0.0(2026-08-17);對應細部文件 `backend-detail/02-tenancy-users.md` v1.0.0、原計畫 v2.9.0、規格書 v1.0.34。*
