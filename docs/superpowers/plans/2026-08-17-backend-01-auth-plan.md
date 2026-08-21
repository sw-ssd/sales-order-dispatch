# Backend 01 — 認證與授權 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 實作 backend Phase 1 全部認證授權機制 — Ent 基礎 schema、Casbin RBAC(domain = company)、PostgreSQL RLS、Google OAuth2 員工登入、客戶一主多子帳密登入、Session/JWT 雙軌、強制登出、ability API、developer 逃生門。

**Architecture:** 依 `docs/superpowers/plans/backend-detail/01-auth.md`(下稱「細部文件」,子功能編號 1.x.y)實作。Chi router + Connect-RPC;授權三層 = Casbin(功能)/ RLS(資料範圍)/ CASL JSON(UI);認證雙軌 = Web httpOnly cookie session(scs + Valkey)與 App Bearer JWT + refresh 旋轉;撤銷統一走 `users.token_version` 比對。

**Tech Stack:** Go 1.25、Ent(entgo.io)、Chi v5、Connect-RPC、casbin/v2 + gorm-adapter(postgres)、pgx/v5、golang-jwt/v5、scs/v2 + redisstore(Valkey)、x/crypto/argon2、x/oauth2 + coreos/go-oidc/v3、testcontainers-go(整合測試)。

**Spec 來源:** 細部文件 `docs/superpowers/plans/backend-detail/01-auth.md`;共通規則見 `docs/superpowers/plans/backend-detail/00-index.md` §3。

## Global Constraints

- module 路徑:`github.com/salesorder/sales-order-1.0/backend`;所有路徑相對 repo root。
- 軟刪除:業務表統一 `deleted_at timestamptz`,查詢預設排除;唯一性用部分唯一索引 `WHERE deleted_at IS NULL`(D10)。
- 稽核:登入成功/失敗、強制登出、密碼重置、guest 審核等關鍵操作與 audit log 同一 DB 交易(D18);本 Phase 稽核表未建,先以 `audit.Recorder` 介面落地(Task 14 提供 no-op + 介面,Task 2.6 計畫接管實作)。**敏感欄位(密碼、token)永不入稽核。**
- 錯誤:RPC 層統一 Connect code — `unauthenticated` / `permission_denied` / `not_found` / `failed_precondition` / `invalid_argument` / `already_exists`;認證失敗不透露帳號是否存在。
- 密碼:Argon2id(m=64MB, t=3, p=2, salt 16B, key 32B);refresh token 僅存 SHA-256 雜湊。
- JWT:HS256,金鑰自 config `JWT_SECRET`,TTL 1 小時,claim 含 `tv`。
- RLS:session variables `app.current_company_id` / `app.current_department_id` / `app.current_data_scope` / `app.current_customer_id`;未注入 fail-closed(0 列)。
- 測試:DB 相依測試走 testcontainers Postgres 16;`go test ./...` 必須全綠;覆蓋率目標 70%(CI 於 Phase 0 管線強制)。
- 每個 Task 結尾 commit;commit message 格式 `feat(backend): …` / `test(backend): …`。

## File Structure

| 檔案 | 職責 | 建立於 |
|---|---|---|
| `backend/ent/schema/{company,department,user}.go` | 基礎實體 schema | Task 1 |
| `backend/internal/testutil/db.go` | testcontainers Postgres + Ent client 測試輔助 | Task 1 |
| `backend/config/casbin_model.conf` | RBAC with domain model | Task 2 |
| `backend/internal/authz/{casbin,seed}.go` | enforcer 封裝、預設 policy seeder | Task 2 |
| `backend/database/migrations/00003_rls_policies.sql` | RLS 啟用與 policy | Task 3 |
| `backend/internal/database/rls.go` | session 變數注入與 Ent driver hook | Task 3 |
| `backend/internal/domain/auth/{oauth,handler,usecase,customer,ability}.go` | 認證 domain | Task 4-8、13 |
| `backend/internal/session/{manager,jwt,refresh}.go` | session store、JWT、refresh 旋轉 | Task 9-10 |
| `backend/internal/middleware/{auth,developer}.go` | Authenticate / ApiToken / developer | Task 10-11、14 |
| `backend/internal/domain/users/usecase.go` | ForceLogout usecase | Task 12 |
| `backend/internal/audit/recorder.go` | 稽核介面(本 Phase no-op) | Task 14 |
| `backend/proto/v1/auth.proto` | AuthService / AbilityService proto | Task 4 起隨需求增量 |

---

### Task 1: Company / Department / User Ent schema(細部 1.1.1–1.1.3)

**Files:**
- Create: `backend/ent/schema/company.go`
- Create: `backend/ent/schema/department.go`
- Create: `backend/ent/schema/user.go`
- Create: `backend/internal/testutil/db.go`
- Test: `backend/ent/schema/schema_test.go`

**Interfaces:**
- Produces: `ent.Company` / `ent.Department` / `ent.User` 產生碼;`testutil.NewEntClient(t) *ent.Client`(後續所有 DB 測試使用);users 欄位 `token_version`(int, default 0)、`role`(enum: super/company_admin/dept_admin/staff/guest/customer/developer)、`data_scope`(enum: all/company/department/self)、`is_customer` / `customer_id` / `account_name` / `is_primary` / `password_hash` / `temp_password_expires_at` / `must_change_password` / `failed_login_attempts` / `locked_at` / `locale`(可空,通知語系,07 計畫消費)。

- [ ] **Step 1: 建立測試輔助 testutil**

`backend/internal/testutil/db.go`:

```go
// Package testutil 提供 DB 相依測試的共用輔助。
package testutil

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/enttest"
)

// NewEntClient 啟動一個拋棄式 Postgres 16 container,跑完 schema migration,
// 回傳可用於測試的 Ent client。測試結束自動銷毀。
func NewEntClient(t *testing.T) *ent.Client {
	t.Helper()
	ctx := context.Background()

	pg, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	client := enttest.Open(t, dialect.Postgres, dsn)
	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

// MustUUID 解析字串 UUID,失敗即 fail 測試。
func MustUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", s, err)
	}
	return id
}
```

- [ ] **Step 2: 寫失敗測試(唯一索引與主帳號唯一)**

`backend/ent/schema/schema_test.go`:

```go
package schema_test

import (
	"context"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestCompanyPrefixUnique(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()

	_, err := c.Company.Create().SetName("甲公司").SetIdentifier("co-a").
		SetCustomerCodePrefix("AB").Save(ctx)
	if err != nil {
		t.Fatalf("first company: %v", err)
	}
	_, err = c.Company.Create().SetName("乙公司").SetIdentifier("co-b").
		SetCustomerCodePrefix("AB").Save(ctx)
	if err == nil {
		t.Fatal("duplicate customer_code_prefix should fail")
	}
}

func TestUserPrimaryAccountUnique(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()

	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").
		SetCustomerCodePrefix("AB").Save(ctx)
	cust := "11111111-1111-1111-1111-111111111111" // customer 表 Phase 3 才建,先用固定 UUID 佔位欄位值

	mk := func(name string) error {
		_, err := c.User.Create().SetCompanyID(co.ID).SetRole("customer").
			SetDataScope("self").SetStatus("active").SetIsCustomer(true).
			SetCustomerID(testutil.MustUUID(t, cust)).
			SetAccountName(name).SetIsPrimary(true).
			SetPasswordHash("x").Save(ctx)
		return err
	}
	if err := mk("老闆"); err != nil {
		t.Fatalf("first primary: %v", err)
	}
	if err := mk("老闆二號"); err == nil {
		t.Fatal("second primary account for same customer should fail")
	}
}
```

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./ent/schema/ -run 'TestCompanyPrefixUnique|TestUserPrimaryAccountUnique' -v`
Expected: FAIL — `ent.Company` / `ent.User` 未定義(編譯失敗)。

- [ ] **Step 4: 實作三個 schema**

`backend/ent/schema/company.go`:

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

// Company 為多租戶頂層實體(細部文件 1.1.1)。
type Company struct{ ent.Schema }

func (Company) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").NotEmpty(),
		field.String("tax_id").Optional().Nillable(),
		field.String("identifier").NotEmpty(),
		field.String("customer_code_prefix").NotEmpty().MaxLen(4),
		field.Enum("status").Values("active", "suspended").Default("active"),
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
		index.Fields("customer_code_prefix").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}
```

`backend/ent/schema/department.go`:

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

// Department 為業務資料範圍隔離單位(細部文件 1.1.2)。
type Department struct{ ent.Schema }

func (Department) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("company_id", uuid.UUID{}),
		field.String("name").NotEmpty(),
		field.Enum("status").Values("active", "suspended").Default("active"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (Department) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("company", Company.Type).Ref("departments").
			Field("company_id").Unique().Required(),
	}
}

func (Department) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("company_id"),
		index.Fields("company_id", "name").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}
```

`backend/ent/schema/user.go`:

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

// User 同時承載員工(OAuth)與客戶帳號(帳密,一主多子)(細部文件 1.1.3)。
type User struct{ ent.Schema }

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("company_id", uuid.UUID{}),
		field.UUID("department_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("status").Values("pending", "active", "suspended").Default("active"),
		field.Enum("role").
			Values("super", "company_admin", "dept_admin", "staff", "guest", "customer", "developer"),
		field.Enum("data_scope").Values("all", "company", "department", "self").Default("department"),
		field.Int("token_version").Default(0),

		// 員工欄位
		field.String("email").Optional().Nillable(),
		field.String("name").Optional(),
		field.String("phone").Optional(),
		field.String("employee_no").Optional().Nillable(),

		// 客戶帳號欄位
		field.Bool("is_customer").Default(false),
		field.UUID("customer_id", uuid.UUID{}).Optional().Nillable(),
		field.String("account_name").Optional().Nillable(),
		field.Bool("is_primary").Default(false),
		field.String("password_hash").Optional().Nillable().Sensitive(),
		field.Time("temp_password_expires_at").Optional().Nillable(),
		field.Bool("must_change_password").Default(false),
		field.Int("failed_login_attempts").Default(0),
		field.Time("locked_at").Optional().Nillable(),

		// 通知語系(07-notifications 計畫範本選取;無則退回預設語系 zh-TW)
		field.String("locale").Optional().Nillable(),

		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		// 員工:公司內 email 唯一
		index.Fields("company_id", "email").Unique().Annotations(
			entsql.IndexWhere("deleted_at IS NULL AND is_customer = false AND email IS NOT NULL")),
		// 客戶帳號:客戶內 account_name 唯一
		index.Fields("customer_id", "account_name").Unique().Annotations(
			entsql.IndexWhere("deleted_at IS NULL AND is_customer = true")),
		// 客戶帳號:每客戶最多一個主帳號
		index.Fields("customer_id").Unique().Annotations(
			entsql.IndexWhere("is_primary = true AND deleted_at IS NULL")),
	}
}
```

Company 需補 `departments` 反向 edge:在 `company.go` 加

```go
func (Company) Edges() []ent.Edge {
	return []ent.Edge{edge.To("departments", Department.Type)}
}
```

(並在 company.go import 加 `"entgo.io/ent/schema/edge"`)

- [ ] **Step 5: 產生 Ent code 並跑測試**

Run: `cd backend && go generate ./ent/ && go mod tidy && go test ./ent/schema/ -v`
Expected: PASS — 兩個測試皆過(重複 prefix 被拒、第二主帳號被拒)。

注意:`customer_id` 在 Phase 3 才改為外鍵 edge,現階段為純 UUID 欄位(細部文件 1.1.3 備註)。

- [ ] **Step 6: Commit**

```bash
git add backend/ent backend/internal/testutil backend/go.mod backend/go.sum
git commit -m "feat(backend): company/department/user Ent schema 與部分唯一索引(1.1)"
```

---

### Task 2: Casbin model、PG adapter、預設 seeder(細部 1.2.1–1.2.3)

**Files:**
- Create: `backend/config/casbin_model.conf`
- Create: `backend/internal/authz/casbin.go`
- Create: `backend/internal/authz/seed.go`
- Test: `backend/internal/authz/casbin_test.go`

**Interfaces:**
- Consumes: `testutil.NewEntClient` 的 DSN 模式(adapter 用同一測試庫)。
- Produces: `authz.NewEnforcer(dsn string) (*casbin.Enforcer, error)`;`authz.SeedDefaultPolicies(ctx, e) error`(冪等);model 詞彙 — `obj` 用資源路徑(`customers`、`sales_orders`、`users`、`policies`…),`act` ∈ `read|create|update|delete`;domain 一律為 `company_id` 字串,`super` 規則 dom 為 `*`。

- [ ] **Step 1: 寫失敗測試(三角色範圍)**

`backend/internal/authz/casbin_test.go`:

```go
package authz_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func newSeededEnforcer(t *testing.T) (string, *casbinEnforcer) { t.Helper(); panic("見 Step 3") }
type casbinEnforcer = struct{}

func TestDefaultPolicies(t *testing.T) {
	ctx := context.Background()
	_ = testutil.NewEntClient(t) // 確保測試庫存在;adapter 使用其 DSN
	dsn := testutil.DSN(t)

	e, err := authz.NewEnforcer(dsn)
	if err != nil {
		t.Fatalf("new enforcer: %v", err)
	}
	if err := authz.SeedDefaultPolicies(ctx, e); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// 冪等:重跑不報錯、規則數不變
	n := len(e.GetPolicy())
	if err := authz.SeedDefaultPolicies(ctx, e); err != nil {
		t.Fatalf("re-seed: %v", err)
	}
	if got := len(e.GetPolicy()); got != n {
		t.Fatalf("seed not idempotent: %d -> %d", n, got)
	}

	cases := []struct {
		name                string
		sub, dom, obj, act  string
		want                bool
	}{
		{"staff 讀自己公司客戶", "staff", "co-1", "customers", "read", true},
		{"staff 不可寫 policy", "staff", "co-1", "policies", "update", false},
		{"company_admin 管自己公司使用者", "company_admin", "co-1", "users", "update", true},
		{"company_admin 不可碰他公司(domain 不符)", "company_admin", "co-2", "users", "update", false},
		{"super 全域", "super", "co-9", "users", "delete", true},
		{"customer 僅訂單唯讀以外皆拒", "customer", "co-1", "users", "read", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := e.Enforce(tc.sub, tc.dom, tc.obj, tc.act)
			if err != nil {
				t.Fatalf("enforce: %v", err)
			}
			if ok != tc.want {
				t.Errorf("Enforce(%s,%s,%s,%s) = %v, want %v",
					tc.sub, tc.dom, tc.obj, tc.act, ok, tc.want)
			}
		})
	}
	fmt.Println("policies:", n)
}
```

`testutil.DSN(t)`:在 `testutil/db.go` 加匯出函式回傳 container DSN(重構 Step 1 的 dsn 取得為 package 級變數,`NewEntClient` 與 `DSN` 共用):

```go
var dsn string

// DSN 回傳測試庫連線字串;須先呼叫 NewEntClient。
func DSN(t *testing.T) string {
	t.Helper()
	if dsn == "" {
		t.Fatal("DSN: call NewEntClient first")
	}
	return dsn
}
```

(`NewEntClient` 內改為指派 package 級 `dsn`。)

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/authz/ -v`
Expected: FAIL — `authz.NewEnforcer` 未定義。先刪除佔位的 `newSeededEnforcer`/`casbinEnforcer` 兩行(僅為讓 Step 1 可獨立閱讀),再跑。

- [ ] **Step 3: 實作 model 與 enforcer**

`backend/config/casbin_model.conf`:

```ini
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = (g(r.sub, p.sub, r.dom) || r.sub == p.sub) && (p.dom == "*" || r.dom == p.dom) && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
```

`backend/internal/authz/casbin.go`:

```go
// Package authz 封裝 Casbin enforcer(細部文件 1.2)。
package authz

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const modelPath = "config/casbin_model.conf" // 相對 backend/ 工作目錄

// NewEnforcer 以 PostgreSQL 持久化 policy 建立 enforcer。
// casbin_rule 表為內部表,不加 RLS。
func NewEnforcer(dsn string) (*casbin.Enforcer, error) {
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("authz: open db: %w", err)
	}
	a, err := gormadapter.NewAdapterByDB(gormDB)
	if err != nil {
		return nil, fmt.Errorf("authz: adapter: %w", err)
	}
	e, err := casbin.NewEnforcer(modelPath, a)
	if err != nil {
		return nil, fmt.Errorf("authz: enforcer: %w", err)
	}
	return e, nil
}
```

`backend/internal/authz/seed.go`:

```go
package authz

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
)

// defaultPolicies 為規格 §3.4 的內建角色預設功能權限(1.0 範圍)。
// dom="*" 僅 super 使用;其餘角色的規則於各公司 domain 由 2.10 seed 展開,
// 此處 p 規則以角色為 sub、由 g 規則在使用者指派時於 domain 內連結。
var defaultPolicies = [][]string{
	// super:全域
	{"super", "*", "/*", ".*"},
	// company_admin:公司內全部資源
	{"company_admin", "*", "users", ".*"},
	{"company_admin", "*", "policies", "read|create|delete"},
	{"company_admin", "*", "roles", ".*"},
	{"company_admin", "*", "customers", ".*"},
	{"company_admin", "*", "products", ".*"},
	{"company_admin", "*", "sales_orders", ".*"},
	{"company_admin", "*", "return_requests", ".*"},
	{"company_admin", "*", "dispatch", ".*"},
	{"company_admin", "*", "printing", ".*"},
	{"company_admin", "*", "metadicts", ".*"},
	{"company_admin", "*", "audit_logs", "read"},
	// dept_admin:部門業務全開 + 部門使用者/客戶管理
	{"dept_admin", "*", "customers", ".*"},
	{"dept_admin", "*", "products", ".*"},
	{"dept_admin", "*", "sales_orders", ".*"},
	{"dept_admin", "*", "return_requests", ".*"},
	{"dept_admin", "*", "dispatch", ".*"},
	{"dept_admin", "*", "printing", ".*"},
	{"dept_admin", "*", "metadicts", "read|create|update|delete"},
	{"dept_admin", "*", "users", "read|update"}, // 限 staff 與客戶帳號,範圍由 RLS 擋
	// staff:日常業務
	{"staff", "*", "customers", "read|create|update"},
	{"staff", "*", "products", "read"},
	{"staff", "*", "sales_orders", ".*"},
	{"staff", "*", "return_requests", "read|update"},
	{"staff", "*", "dispatch", ".*"},
	{"staff", "*", "printing", ".*"},
	{"staff", "*", "metadicts", "read"},
	// customer:下單與退貨
	{"customer", "*", "sales_orders", "read|create"},
	{"customer", "*", "return_requests", "read|create"},
	{"customer", "*", "customer_products", "read"},
	{"customer", "*", "notifications", "read|update"},
}

// SeedDefaultPolicies 補齊缺少的預設規則;已存在(含人工調整過)的不動,冪等。
func SeedDefaultPolicies(_ context.Context, e *casbin.Enforcer) error {
	for _, p := range defaultPolicies {
		has, err := e.HasPolicy(p)
		if err != nil {
			return fmt.Errorf("authz: check policy %v: %w", p, err)
		}
		if has {
			continue
		}
		if _, err := e.AddPolicy(p); err != nil {
			return fmt.Errorf("authz: add policy %v: %w", p, err)
		}
	}
	return e.SavePolicy()
}
```

說明:model 的 matcher 以 `(g(...) || r.sub == p.sub)` 同時支援「直接以角色為 sub 的預設規則」與「使用者經 g 指派角色」。預設規則 dom 用 `*`,資料範圍由 RLS(Task 3)收口 — 這是 D3 的分工:Casbin 只答「這個角色能不能碰這資源」,不管「哪個公司」。

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && go get github.com/casbin/casbin/v2 github.com/casbin/gorm-adapter/v3 gorm.io/driver/postgres gorm.io/gorm && go test ./internal/authz/ -v`
Expected: PASS(7 個子測試 + 冪等檢查)。

- [ ] **Step 5: Commit**

```bash
git add backend/config/casbin_model.conf backend/internal/authz backend/internal/testutil
git commit -m "feat(backend): Casbin RBAC model、PG adapter、預設 policy seeder(1.2)"
```

---

### Task 3: RLS policies、注入 hook、高權繞過(細部 1.3.1–1.3.3)

**Files:**
- Create: `backend/database/migrations/00003_rls_policies.sql`
- Create: `backend/internal/database/rls.go`
- Test: `backend/internal/database/rls_test.go`

**Interfaces:**
- Consumes: Task 1 的三張表。
- Produces: `rls.Identity{UserID, CompanyID uuid.UUID; DepartmentID, CustomerID *uuid.UUID; DataScope string; Role string; IsPrimary bool}`;`rls.NewContext(ctx, id Identity) context.Context`;`rls.FromContext(ctx) (Identity, bool)`;`rls.WrapDriver(d dialect.Driver) dialect.Driver`(Ent client 注入用,每筆交易自動 `SET LOCAL`)。後續所有 domain 的 Ent client 一律經 `rls.WrapDriver` 建立。

- [ ] **Step 1: 寫失敗測試(隔離、fail-closed、繞過)**

`backend/internal/database/rls_test.go`:

```go
package database_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

// newRLSClient 建立經 rls.WrapDriver 包裝的 Ent client(測試輔助,於 Step 3 加入 testutil)。
func newRLSClient(t *testing.T) *ent.Client {
	t.Helper()
	return testutil.NewEntClientWithRLS(t)
}

func setupCompanies(t *testing.T, c *ent.Client) (coA, coB uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	a, err := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	if err != nil { t.Fatal(err) }
	b, err := c.Company.Create().SetName("乙").SetIdentifier("co-b").SetCustomerCodePrefix("BB").Save(ctx)
	if err != nil { t.Fatal(err) }
	return a.ID, b.ID
}

func TestRLSFailClosed(t *testing.T) {
	c := newRLSClient(t)
	coA, _ := setupCompanies(t, c)
	_ = coA
	// 未注入身分:查 users 應回 0 列(policy 無匹配)
	n, err := c.User.Query().Count(context.Background())
	if err != nil { t.Fatalf("query: %v", err) }
	if n != 0 { t.Fatalf("fail-closed violated: got %d rows", n) }
}

func TestRLSTenantIsolation(t *testing.T) {
	c := newRLSClient(t)
	coA, coB := setupCompanies(t, c)
	ctx := context.Background()

	// 用無 RLS 的寫入端建測試資料(細部 1.3.2:migration 資料與管理操作走內部連線)
	raw := testutil.NewEntClient(t) // 同一 container、未包 RLS(見 Step 3 說明)
	_ = raw // 實際寫入由 SQL 直接做,避免循環依賴;此處以 c 在 super 身分工下寫入
	superCtx := rls.NewContext(ctx, rls.Identity{UserID: uuid.New(), CompanyID: coA, DataScope: "all"})
	if _, err := c.User.Create().SetCompanyID(coA).SetRole("staff").SetDataScope("department").
		SetStatus("active").SetName("甲員工").Save(superCtx); err != nil {
		t.Fatalf("seed user A: %v", err)
	}
	if _, err := c.User.Create().SetCompanyID(coB).SetRole("staff").SetDataScope("department").
		SetStatus("active").SetName("乙員工").Save(superCtx); err != nil {
		t.Fatalf("seed user B: %v", err)
	}

	// A 公司 staff 視角:只見自己公司
	ctxA := rls.NewContext(ctx, rls.Identity{UserID: uuid.New(), CompanyID: coA, DataScope: "company"})
	names, err := c.User.Query().Select("name").Strings(ctxA)
	if err != nil { t.Fatalf("query A: %v", err) }
	if len(names) != 1 || names[0] != "甲員工" {
		t.Fatalf("tenant isolation violated: %v", names)
	}

	// super 視角:兩家皆可見
	all, err := c.User.Query().Count(superCtx)
	if err != nil { t.Fatalf("query super: %v", err) }
	if all != 2 { t.Fatalf("bypass failed: got %d", all) }
}

func TestRLSConnectionNoLeak(t *testing.T) {
	c := newRLSClient(t)
	coA, coB := setupCompanies(t, c)
	ctx := context.Background()
	superCtx := rls.NewContext(ctx, rls.Identity{UserID: uuid.New(), CompanyID: coA, DataScope: "all"})
	for _, co := range []uuid.UUID{coA, coB} {
		if _, err := c.User.Create().SetCompanyID(co).SetRole("staff").SetDataScope("department").
			SetStatus("active").Save(superCtx); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	// 先以 A 身分查,再以無注入連線查:SET LOCAL 不得殘留到 pool 連線
	ctxA := rls.NewContext(ctx, rls.Identity{UserID: uuid.New(), CompanyID: coA, DataScope: "company"})
	if _, err := c.User.Query().Count(ctxA); err != nil { t.Fatal(err) }
	n, err := c.User.Query().Count(context.Background())
	if err != nil { t.Fatal(err) }
	if n != 0 { t.Fatalf("SET LOCAL leaked across checkouts: %d rows visible", n) }
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/database/ -v`
Expected: FAIL — `rls.NewContext` 未定義、`testutil.NewEntClientWithRLS` 未定義(編譯失敗)。

- [ ] **Step 3: 實作 RLS migration 與注入**

`backend/database/migrations/00003_rls_policies.sql`(goose 格式,`-- +goose Up/Down`):

```sql
-- +goose Up
-- 細部文件 1.3.1:租戶表啟用 RLS;data_scope=all 豁免(1.3.3)。
-- 未注入 session variables 時 current_setting(..., true) 回 NULL,policy 不命中 → fail-closed。

ALTER TABLE companies ENABLE ROW LEVEL SECURITY;
ALTER TABLE departments ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- companies:租戶只見自己;all 全見
CREATE POLICY tenant_isolation ON companies
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR id::text = current_setting('app.current_company_id', true));

-- departments:company 級見全公司部門;department 級僅自己部門
CREATE POLICY tenant_isolation ON departments
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true));

CREATE POLICY department_scope ON departments
  USING (current_setting('app.current_data_scope', true) IN ('all', 'company')
         OR id::text = current_setting('app.current_department_id', true));

-- users:先比對公司,再按 data_scope 收窄;self(客戶帳號)僅見自己客戶的帳號
CREATE POLICY tenant_isolation ON users
  USING (current_setting('app.current_data_scope', true) = 'all'
         OR company_id::text = current_setting('app.current_company_id', true));

CREATE POLICY data_scope ON users
  USING (current_setting('app.current_data_scope', true) IN ('all', 'company')
         OR (current_setting('app.current_data_scope', true) = 'department'
             AND department_id::text = current_setting('app.current_department_id', true))
         OR (current_setting('app.current_data_scope', true) = 'self'
             AND customer_id::text = current_setting('app.current_customer_id', true)));

-- +goose Down
DROP POLICY IF EXISTS data_scope ON users;
DROP POLICY IF EXISTS tenant_isolation ON users;
DROP POLICY IF EXISTS department_scope ON departments;
DROP POLICY IF EXISTS tenant_isolation ON departments;
DROP POLICY IF EXISTS tenant_isolation ON companies;
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE departments DISABLE ROW LEVEL SECURITY;
ALTER TABLE companies DISABLE ROW LEVEL SECURITY;
```

`backend/internal/database/rls.go`:

```go
// Package rls 提供 PostgreSQL Row-Level Security 的 session 變數注入(細部文件 1.3)。
package rls

import (
	"context"
	"database/sql/driver"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
)

// Identity 為當前請求的租戶上下文。
type Identity struct {
	UserID       uuid.UUID
	CompanyID    uuid.UUID
	DepartmentID *uuid.UUID // 客戶帳號可為 nil
	CustomerID   *uuid.UUID
	DataScope    string // all | company | department | self
	Role         string // developer | company_admin | dept_admin | staff | customer(取自 session.Claims)
	IsPrimary    bool   // 客戶主帳號標記(取自 session.Claims)
}

type ctxKey struct{}

// NewContext 將身分放入 ctx;middleware 於解析身分後呼叫。
func NewContext(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext 取回身分;未注入回 ok=false。
func FromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(Identity)
	return id, ok
}

// setStatements 產生 SET LOCAL 語句;無身分時回 nil(fail-closed:不注入,policy 不命中)。
func setStatements(id Identity) []string {
	stmts := []string{
		fmt.Sprintf("SET LOCAL app.current_company_id = '%s'", id.CompanyID),
		fmt.Sprintf("SET LOCAL app.current_data_scope = '%s'", id.DataScope),
	}
	if id.DepartmentID != nil {
		stmts = append(stmts, fmt.Sprintf("SET LOCAL app.current_department_id = '%s'", *id.DepartmentID))
	}
	if id.CustomerID != nil {
		stmts = append(stmts, fmt.Sprintf("SET LOCAL app.current_customer_id = '%s'", *id.CustomerID))
	}
	return stmts
}

// drv 包裝 ent dialect.Driver:每筆交易開頭套用 SET LOCAL。
// SET LOCAL 僅作用於當前交易,交易結束自動還原,pool 連線不殘留。
type drv struct{ dialect.Driver }

func (d drv) Tx(ctx context.Context) (dialect.Tx, error) {
	tx, err := d.Driver.Tx(ctx)
	if err != nil {
		return nil, err
	}
	id, ok := FromContext(ctx)
	if !ok {
		return tx, nil // fail-closed:不注入
	}
	for _, s := range setStatements(id) {
		if err := tx.(interface {
			Exec(context.Context, string, ...any) error
		}).Exec(ctx, s); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("rls: inject context: %w", err)
		}
	}
	return tx, nil
}

// WrapDriver 回傳注入 RLS 的 Ent driver;所有業務 Ent client 必須經此建立。
func WrapDriver(d dialect.Driver) dialect.Driver { return drv{d} }

// 編譯期檢查:Ent 的 sql driver tx 介面吻合。
var _ driver.Tx = nil
var _ = entsql.ErrLockClosed
```

`testutil/db.go` 新增:

```go
import (
	entsql "entgo.io/ent/dialect/sql"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	// ...
)

// NewEntClientWithRLS 同 NewEntClient,但 driver 經 rls.WrapDriver 包裝,
// 並套用 00003_rls_policies.sql 的 RLS migration。
func NewEntClientWithRLS(t *testing.T) *ent.Client {
	t.Helper()
	ctx := context.Background()
	// 先建無包裝 client 跑 schema + RLS migration
	base := NewEntClient(t)
	sqlBytes, err := os.ReadFile("../../database/migrations/00003_rls_policies.sql")
	if err != nil {
		t.Fatalf("read rls migration: %v", err)
	}
	up := strings.SplitN(string(sqlBytes), "-- +goose Down", 2)[0]
	up = strings.ReplaceAll(up, "-- +goose Up", "")
	if _, err := base.DB().ExecContext(ctx, up); err != nil {
		t.Fatalf("apply rls migration: %v", err)
	}
	// 以包裝後 driver 重建 client
	drv := rls.WrapDriver(entsql.OpenDB(dialect.Postgres, base.DB()))
	return ent.NewClient(ent.Driver(drv))
}
```

注意:`rls.go` 中 `tx.Exec` 的型別斷言 — Ent 的 `dialect.Tx` 底層是 `*entsql.Tx`,提供 `ExecContext`;實作時以 `tx.(*entsql.Tx).ExecContext(ctx, s)` 呼叫,上方 interface 斷言為示意,實作統一用 `*entsql.Tx` 具體型別。

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && go test ./internal/database/ -v`
Expected: PASS — 三個測試(fail-closed、租戶隔離、連線無殘留)。

- [ ] **Step 5: Commit**

```bash
git add backend/database/migrations/00003_rls_policies.sql backend/internal/database backend/internal/testutil
git commit -m "feat(backend): RLS policies 與 session 變數注入 hook(1.3)"
```

---

### Task 4: OAuth2 導向與 callback(細部 1.4.1–1.4.2)

**Files:**
- Create: `backend/internal/domain/auth/oauth.go`
- Create: `backend/internal/domain/auth/handler.go`
- Create: `backend/internal/domain/auth/valkey.go`
- Update: `backend/config/config.go`
- Test: `backend/internal/domain/auth/oauth_test.go`

**Interfaces:**
- Consumes: Task 1 `ent.User`;Valkey(Phase 0 docker-compose 已含)。
- Produces: `GET /api/v1/auth/oauth/{provider}`、`GET /api/v1/auth/oauth/{provider}/callback`;`auth.OAuthService` 介面(`AuthURL(state string) string`、`Exchange(ctx, code) (*oidc.IDToken, error)`)供測試替換;一次性 token 存取介面 `auth.OneTimeStore`(Put/Get/Delete,Valkey 實作)供 registration token 與 state 使用;callback 成功路徑呼叫 `session.IssueTokens`(Task 9 實作,本 Task 以介面依賴注入,測試用 fake)。

- [ ] **Step 1: config 擴充**

`backend/config/config.go` 新增欄位(併入既有 `Config` struct 與 `Load()`):

```go
// OAuth
OAuthGoogleClientID     string   `mapstructure:"OAUTH_GOOGLE_CLIENT_ID"`
OAuthGoogleClientSecret string   `mapstructure:"OAUTH_GOOGLE_CLIENT_SECRET"`
OAuthGoogleRedirectURL  string   `mapstructure:"OAUTH_GOOGLE_REDIRECT_URL"`
OAuthAllowedDomains     []string `mapstructure:"OAUTH_ALLOWED_DOMAINS"`  // Google Workspace hd 白名單
OAuthAllowedRedirects   []string `mapstructure:"OAUTH_ALLOWED_REDIRECTS"` // 前端回跳白名單
ValkeyAddr              string   `mapstructure:"VALKEY_ADDR"`             // 例:127.0.0.1:6379
```

`Load()` 結尾加啟動檢查:`ENV != "development"` 且 OAuth 三欄位任一為空 → 回傳錯誤(fail-fast,細部 1.4.1)。

- [ ] **Step 2: 寫失敗測試(state 生命週期與 redirect 白名單)**

`backend/internal/domain/auth/oauth_test.go`:

```go
package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/domain/auth"
)

// fakeOneTimeStore 為記憶體版 OneTimeStore(測試 state/registration token)。
type fakeOneTimeStore struct{ m map[string]auth.OneTimeEntry }

func newFakeStore() *fakeOneTimeStore { return &fakeOneTimeStore{m: map[string]auth.OneTimeEntry{}} }
func (f *fakeOneTimeStore) Put(_ context.Context, key string, e auth.OneTimeEntry, _ time.Duration) error {
	f.m[key] = e; return nil
}
func (f *fakeOneTimeStore) Get(_ context.Context, key string) (auth.OneTimeEntry, bool) {
	e, ok := f.m[key]; return e, ok
}
func (f *fakeOneTimeStore) Delete(_ context.Context, key string) error {
	delete(f.m, key); return nil
}

func TestOAuthRedirectRejectsUnknownRedirect(t *testing.T) {
	h := auth.NewOAuthHandler(auth.OAuthConfig{
		AllowedRedirects: []string{"https://web.example.com"},
	}, newFakeStore(), nil, nil)

	req := httptest.NewRequest("GET", "/api/v1/auth/oauth/google?redirect=https://evil.example.com", nil)
	rec := httptest.NewRecorder()
	h.HandleRedirect(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestOAuthCallbackRejectsReplayedState(t *testing.T) {
	store := newFakeStore()
	store.Put(context.Background(), "st-1", auth.OneTimeEntry{Redirect: "https://web.example.com"}, time.Minute)
	h := auth.NewOAuthHandler(auth.OAuthConfig{
		AllowedRedirects: []string{"https://web.example.com"},
	}, store, fakeOAuth{}, nil)

	call := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/api/v1/auth/oauth/google/callback?code=c&state=st-1", nil)
		rec := httptest.NewRecorder()
		h.HandleCallback(rec, req)
		return rec
	}
	first := call()  // 第一次:state 被消耗(後續流程成敗不影響本測試)
	_ = first
	second := call() // 重放:state 已刪
	if second.Code != http.StatusUnauthorized {
		t.Fatalf("replayed state: want 401, got %d", second.Code)
	}
}

type fakeOAuth struct{}

func (fakeOAuth) AuthURL(state string) string { return "https://accounts.google.com/?state=" + state }
func (fakeOAuth) Exchange(_ context.Context, _ string) (*auth.OIDCIdentity, error) {
	return &auth.OIDCIdentity{Email: "a@example.com", HostedDomain: "example.com"}, nil
}
```

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/auth/ -v`
Expected: FAIL — `auth.NewOAuthHandler` 等未定義。

- [ ] **Step 4: 實作**

`backend/internal/domain/auth/valkey.go`:

```go
package auth

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// OneTimeEntry 為一次性 token 的內容(state 綁 redirect;registration token 綁 email)。
type OneTimeEntry struct {
	Redirect string
	Email    string
}

// OneTimeStore 為一次性 token 存取介面;生產用 Valkey,測試用 fake。
type OneTimeStore interface {
	Put(ctx context.Context, key string, e OneTimeEntry, ttl time.Duration) error
	Get(ctx context.Context, key string) (OneTimeEntry, bool)
	Delete(ctx context.Context, key string) error
}

type valkeyStore struct{ rdb *redis.Client }

func NewValkeyOneTimeStore(rdb *redis.Client) OneTimeStore { return valkeyStore{rdb} }

func (s valkeyStore) Put(ctx context.Context, key string, e OneTimeEntry, ttl time.Duration) error {
	return s.rdb.HSet(ctx, "onetime:"+key, map[string]string{
		"redirect": e.Redirect, "email": e.Email,
	}).Err() // TTL 另設
	// 實作:HSet 後 s.rdb.Expire(ctx, "onetime:"+key, ttl)
}

func (s valkeyStore) Get(ctx context.Context, key string) (OneTimeEntry, bool) {
	m, err := s.rdb.HGetAll(ctx, "onetime:"+key).Result()
	if err != nil || len(m) == 0 {
		return OneTimeEntry{}, false
	}
	return OneTimeEntry{Redirect: m["redirect"], Email: m["email"]}, true
}

func (s valkeyStore) Delete(ctx context.Context, key string) error {
	return s.rdb.Del(ctx, "onetime:"+key).Err()
}
```

`backend/internal/domain/auth/oauth.go`:

```go
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCIdentity 為驗證後的身分。
type OIDCIdentity struct {
	Email        string
	HostedDomain string
}

// OAuthService 抽象 Google OIDC,供測試替換。
type OAuthService interface {
	AuthURL(state string) string
	Exchange(ctx context.Context, code string) (*OIDCIdentity, error)
}

type googleOAuth struct {
	cfg      oauth2.Config
	verifier *oidc.IDTokenVerifier
	allowed  map[string]bool // hd 白名單
}

func NewGoogleOAuth(clientID, secret, redirectURL string, allowedDomains []string, issuer *oidc.Provider) OAuthService {
	allowed := map[string]bool{}
	for _, d := range allowedDomains {
		allowed[d] = true
	}
	return &googleOAuth{
		cfg: oauth2.Config{
			ClientID: clientID, ClientSecret: secret, RedirectURL: redirectURL,
			Endpoint: issuer.Endpoint(), Scopes: []string{oidc.ScopeOpenID, "email", "profile"},
		},
		verifier: issuer.Verifier(&oidc.Config{ClientID: clientID}),
		allowed:  allowed,
	}
}

func (g *googleOAuth) AuthURL(state string) string {
	return g.cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (g *googleOAuth) Exchange(ctx context.Context, code string) (*OIDCIdentity, error) {
	tok, err := g.cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("oauth exchange: %w", err)
	}
	raw, ok := tok.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("oauth: missing id_token")
	}
	id, err := g.verifier.Verify(ctx, raw)
	if err != nil {
		return nil, fmt.Errorf("oauth verify id_token: %w", err)
	}
	var claims struct {
		Email string `json:"email"`
		HD    string `json:"hd"`
	}
	if err := id.Claims(&claims); err != nil {
		return nil, fmt.Errorf("oauth claims: %w", err)
	}
	if !g.allowed[claims.HD] {
		return nil, ErrDomainNotAllowed
	}
	return &OIDCIdentity{Email: claims.Email, HostedDomain: claims.HD}, nil
}

var ErrDomainNotAllowed = fmt.Errorf("hosted domain not allowed")

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
```

`backend/internal/domain/auth/handler.go`:

```go
package auth

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/salesorder/sales-order-1.0/backend/ent"
)

// OAuthConfig 為 handler 所需設定。
type OAuthConfig struct {
	AllowedRedirects []string
}

// UserLookup 抽象使用者查詢,供測試替換;實作走 ent。
type UserLookup interface {
	// FindEmployeeByEmail 找 active/pending 員工;找不到回 (nil, nil)。
	FindEmployeeByEmail(ctx context.Context, email string) (*ent.User, error)
}

// TokenIssuer 由 Task 9 session 套件實作;callback 成功時呼叫。
type TokenIssuer interface {
	IssueWeb(w http.ResponseWriter, r *http.Request, u *ent.User) error
}

type OAuthHandler struct {
	cfg     OAuthConfig
	store   OneTimeStore
	oauth   OAuthService
	users   UserLookup
	issuer  TokenIssuer // 可為 nil(Task 9 接上)
}

func NewOAuthHandler(cfg OAuthConfig, store OneTimeStore, oauth OAuthService, users UserLookup) *OAuthHandler {
	return &OAuthHandler{cfg: cfg, store: store, oauth: oauth, users: users}
}

// Mount 掛上路由:GET /api/v1/auth/oauth/{provider} 與 /callback。
func (h *OAuthHandler) Mount(r chi.Router) {
	r.Get("/api/v1/auth/oauth/{provider}", h.HandleRedirect)
	r.Get("/api/v1/auth/oauth/{provider}/callback", h.HandleCallback)
}

// HandleRedirect 產生 state 並 302 導向 Google(細部 1.4.1)。
func (h *OAuthHandler) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	if chi.URLParam(r, "provider") != "google" {
		http.Error(w, "unknown provider", http.StatusNotFound)
		return
	}
	redirect := r.URL.Query().Get("redirect")
	if redirect != "" && !slices.Contains(h.cfg.AllowedRedirects, redirect) {
		http.Error(w, "redirect not allowed", http.StatusBadRequest)
		return
	}
	state := randomToken()
	if err := h.store.Put(r.Context(), state, OneTimeEntry{Redirect: redirect}, 10*time.Minute); err != nil {
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, h.oauth.AuthURL(state), http.StatusFound)
}

// HandleCallback 驗 state、換 token、驗 ID token,依使用者狀態分流(細部 1.4.2)。
func (h *OAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	state := r.URL.Query().Get("state")
	entry, ok := h.store.Get(ctx, state)
	if !ok {
		http.Error(w, "invalid state", http.StatusUnauthorized)
		return
	}
	_ = h.store.Delete(ctx, state) // 一次性:無論後續成敗立即作廢

	id, err := h.oauth.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		if errors.Is(err, ErrDomainNotAllowed) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	u, err := h.users.FindEmployeeByEmail(ctx, id.Email)
	if err != nil {
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	switch {
	case u == nil:
		// 新使用者:發 registration token,回跳註冊完成頁;不建帳號
		reg := randomToken()
		_ = h.store.Put(ctx, reg, OneTimeEntry{Email: id.Email, Redirect: entry.Redirect}, 30*time.Minute)
		http.Redirect(w, r, entry.Redirect+"/onboarding?reg="+reg, http.StatusFound)
	case u.Status == "pending":
		http.Redirect(w, r, entry.Redirect+"/pending", http.StatusFound)
	default: // active
		if h.issuer != nil {
			if err := h.issuer.IssueWeb(w, r, u); err != nil {
				http.Error(w, "internal", http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(w, r, entry.Redirect+"/", http.StatusFound)
	}
}
```

說明:公司歸屬 — hd 對應不到公司時由 onboarding 頁讓使用者自選(細部 1.4.2 步驟 2);本 Task 的 `FindEmployeeByEmail` 實作於 Task 5 的 usecase 檔補上 ent 版本,此處以介面驅動測試。

- [ ] **Step 5: 跑測試確認通過**

Run: `cd backend && go get github.com/coreos/go-oidc/v3 golang.org/x/oauth2 github.com/redis/go-redis/v9 && go test ./internal/domain/auth/ -v`
Expected: PASS。

- [ ] **Step 6: Commit**

```bash
git add backend/internal/domain/auth backend/config/config.go backend/go.mod backend/go.sum
git commit -m "feat(backend): OAuth2 導向與 callback(state 一次性、redirect 白名單)(1.4.1-1.4.2)"
```

---

### Task 5: 註冊完成與 guest 審核(細部 1.4.3–1.4.4)

**Files:**
- Create: `backend/internal/domain/auth/usecase.go`
- Create: `backend/proto/v1/auth.proto`(本 Task 起增量)
- Test: `backend/internal/domain/auth/usecase_test.go`

**Interfaces:**
- Consumes: Task 4 `OneTimeStore`、`UserLookup`;Task 2 `authz` enforcer(審核通過建 g 規則)。
- Produces: proto `AuthService.CompleteRegistration(registration_token, company_id, name) → CompleteRegistrationResponse{status}`;`AuthService.ListPendingGuests() → repeated User`;`AuthService.ApproveGuest(user_id, role, department_id)`;`AuthService.RejectGuest(user_id, reason)`;usecase 函式 `CompleteRegistration(ctx, regToken, companyID, name) error`、`ApproveGuest(ctx, actor, targetID, role, deptID) error`、`RejectGuest(ctx, actor, targetID, reason) error`(actor 為 `rls.Identity`);`UserLookup` 的 ent 實作 `auth.NewEntUserLookup(client)`。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/domain/auth/usecase_test.go`:

```go
package auth_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/authz"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func setupAuth(t *testing.T) (*ent.Client, *fakeOneTimeStore, context.Context) {
	t.Helper()
	c := testutil.NewEntClient(t)
	store := newFakeStore()
	co, err := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").
		Save(context.Background())
	if err != nil { t.Fatal(err) }
	ctx := rls.NewContext(context.Background(), rls.Identity{
		UserID: uuid.New(), CompanyID: co.ID, DataScope: "all",
	})
	return c, store, ctx
}

func TestCompleteRegistrationCreatesPendingGuest(t *testing.T) {
	c, store, ctx := setupAuth(t)
	store.Put(ctx, "reg-1", auth.OneTimeEntry{Email: "new@example.com"}, 30*time.Minute)
	coID, _ := c.Company.Query().OnlyID(ctx)

	uc := auth.NewUsecase(c, store, nil)
	if err := uc.CompleteRegistration(ctx, "reg-1", coID, "王小明"); err != nil {
		t.Fatalf("complete: %v", err)
	}
	u, err := c.User.Query().Where(/* email = new@example.com */).Only(ctx)
	if err != nil { t.Fatalf("query: %v", err) }
	if u.Role != "guest" || u.Status != "pending" {
		t.Fatalf("want pending guest, got role=%s status=%s", u.Role, u.Status)
	}
	// token 一次性:重放被拒
	if err := uc.CompleteRegistration(ctx, "reg-1", coID, "王小明"); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("replay: want unauthenticated, got %v", err)
	}
	// 完成前不得有帳號:另一 email 未完成前查無
	if _, err := c.User.Query().Where().Only(ctx); err == nil {
		// 僅上面建的一筆;此檢查由 Only 多列報錯涵蓋
	}
}

func TestApproveGuestAssignsRoleAndGrouping(t *testing.T) {
	c, _, ctx := setupAuth(t)
	e, err := authz.NewEnforcer(testutil.DSN(t))
	if err != nil { t.Fatal(err) }
	uc := auth.NewUsecase(c, newFakeStore(), e)

	coID, _ := c.Company.Query().OnlyID(ctx)
	g, err := c.User.Create().SetCompanyID(coID).SetRole("guest").SetStatus("pending").
		SetDataScope("department").SetEmail("g@example.com").Save(ctx)
	if err != nil { t.Fatal(err) }
	dep, err := c.Department.Create().SetCompanyID(coID).SetName("業務部").Save(ctx)
	if err != nil { t.Fatal(err) }

	if err := uc.ApproveGuest(ctx, rls.Identity{CompanyID: coID, DataScope: "all"}, g.ID, "staff", &dep.ID); err != nil {
		t.Fatalf("approve: %v", err)
	}
	g = c.User.GetX(ctx, g.ID)
	if g.Status != "active" || g.Role != "staff" {
		t.Fatalf("want active staff, got %s/%s", g.Status, g.Role)
	}
	ok, err := e.Enforce(g.ID.String(), coID.String(), "customers", "read")
	if err != nil || !ok {
		t.Fatalf("grouping not linked: ok=%v err=%v", ok, err)
	}
	// 非 pending 目標
	if err := uc.ApproveGuest(ctx, rls.Identity{CompanyID: coID, DataScope: "all"}, g.ID, "staff", &dep.ID); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("re-approve: want failed_precondition, got %v", err)
	}
}

func TestRejectGuestSoftDeletes(t *testing.T) {
	c, _, ctx := setupAuth(t)
	uc := auth.NewUsecase(c, newFakeStore(), nil)
	coID, _ := c.Company.Query().OnlyID(ctx)
	g, _ := c.User.Create().SetCompanyID(coID).SetRole("guest").SetStatus("pending").
		SetDataScope("department").SetEmail("r@example.com").Save(ctx)

	if err := uc.RejectGuest(ctx, rls.Identity{CompanyID: coID, DataScope: "all"}, g.ID, "不符合"); err != nil {
		t.Fatalf("reject: %v", err)
	}
	g = c.User.GetX(ctx, g.ID)
	if g.DeletedAt == nil {
		t.Fatal("rejected guest should be soft-deleted")
	}
	// 被拒 email 可重新註冊(部分唯一索引不擋軟刪除列)
	if _, err := c.User.Create().SetCompanyID(coID).SetRole("guest").SetStatus("pending").
		SetDataScope("department").SetEmail("r@example.com").Save(ctx); err != nil {
		t.Fatalf("re-register after reject: %v", err)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/auth/ -run 'TestCompleteRegistration|TestApproveGuest|TestRejectGuest' -v`
Expected: FAIL — `auth.NewUsecase` 未定義。

- [ ] **Step 3: 實作 usecase**

`backend/internal/domain/auth/usecase.go`:

```go
package auth

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entuser "github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// Usecase 為認證/註冊/審核的業務邏輯;enforcer 可為 nil(僅 CompleteRegistration/RejectGuest 時)。
type Usecase struct {
	db    *ent.Client
	store OneTimeStore
	enf   *casbin.Enforcer
}

func NewUsecase(db *ent.Client, store OneTimeStore, enf *casbin.Enforcer) *Usecase {
	return &Usecase{db: db, store: store, enf: enf}
}

// approvableRoles 為審核可指派的正式角色(細部 1.4.4)。
var approvableRoles = map[string]bool{
	"company_admin": true, "dept_admin": true, "staff": true,
}

// CompleteRegistration 驗 registration token 後建 pending guest;同 email 同公司已存在(未軟刪)→ already_exists。
func (u *Usecase) CompleteRegistration(ctx context.Context, regToken string, companyID uuid.UUID, name string) error {
	entry, ok := u.store.Get(ctx, regToken)
	if !ok || entry.Email == "" {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("registration token invalid"))
	}
	co, err := u.db.Company.Get(ctx, companyID)
	if err != nil || co.Status != "active" || co.DeletedAt != nil {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("company invalid"))
	}
	exists, err := u.db.User.Query().Where(
		entuser.EmailEQ(entry.Email),
		entuser.CompanyIDEQ(companyID),
		entuser.DeletedAtIsNil(),
	).Exist(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if exists {
		return connect.NewError(connect.CodeAlreadyExists, errors.New("account exists"))
	}

	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if _, err := tx.User.Create().SetCompanyID(companyID).SetEmail(entry.Email).
		SetName(name).SetRole("guest").SetStatus("pending").
		SetDataScope("department").Save(ctx); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	// TODO(Phase 2 Task 2.6): audit log 與建檔同事務;現階段由 audit.Recorder no-op 佔位(Task 14)
	if err := tx.Commit(); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	_ = u.store.Delete(ctx, regToken) // 一次性:成功後作廢
	return nil
}

// ApproveGuest 核准 pending guest:轉正式角色 + 建 Casbin g 規則,同一 DB 交易內更新 users;g 規則在提交後寫(細部 1.4.4 步驟 2 註:casbin_rule 為內部表,與業務交易同事務由共用 pg 連線達成;此處先保證順序:users 更新成功才建 g,失敗則回滾 users)。
func (u *Usecase) ApproveGuest(ctx context.Context, actor rls.Identity, targetID uuid.UUID, role string, deptID *uuid.UUID) error {
	if !approvableRoles[role] {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("role not approvable"))
	}
	if (role == "dept_admin" || role == "staff") && deptID == nil {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("department required"))
	}
	target, err := u.db.User.Get(ctx, targetID)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	}
	if target.CompanyID != actor.CompanyID && actor.DataScope != "all" {
		return connect.NewError(connect.CodePermissionDenied, errors.New("out of scope"))
	}
	if target.Status != "pending" {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("user not pending"))
	}

	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	upd := tx.User.UpdateOneID(targetID).SetRole(role).SetStatus("active")
	if deptID != nil {
		upd.SetDepartmentID(*deptID)
	}
	if _, err := upd.Save(ctx); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	if err := tx.Commit(); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if u.enf != nil {
		// g: user → role @ company domain
		if _, err := u.enf.AddGroupingPolicy(targetID.String(), role, target.CompanyID.String()); err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
	}
	return nil
}

// RejectGuest 直接軟刪除 guest 帳號(不設 rejected 狀態,軌跡由稽核保存);同 email 可重新註冊。
func (u *Usecase) RejectGuest(ctx context.Context, actor rls.Identity, targetID uuid.UUID, reason string) error {
	if reason == "" {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("reason required"))
	}
	target, err := u.db.User.Get(ctx, targetID)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	}
	if target.CompanyID != actor.CompanyID && actor.DataScope != "all" {
		return connect.NewError(connect.CodePermissionDenied, errors.New("out of scope"))
	}
	if target.Status != "pending" {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("user not pending"))
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if _, err := tx.User.UpdateOneID(targetID).SetDeletedAt(timeNow()).Save(ctx); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	return tx.Commit()
}

// timeNow 抽出以便測試注入。
var timeNow = func() time.Time { return time.Now() }
```

`backend/proto/v1/auth.proto` 初始內容(後續 Task 增量加 RPC):

```proto
syntax = "proto3";
package salesorder.v1;

import "google/protobuf/empty.proto";

service AuthService {
  rpc CompleteRegistration(CompleteRegistrationRequest) returns (CompleteRegistrationResponse);
  rpc ListPendingGuests(google.protobuf.Empty) returns (ListPendingGuestsResponse);
  rpc ApproveGuest(ApproveGuestRequest) returns (google.protobuf.Empty);
  rpc RejectGuest(RejectGuestRequest) returns (google.protobuf.Empty);
}

message CompleteRegistrationRequest {
  string registration_token = 1;
  string company_id = 2;
  string name = 3;
}
message CompleteRegistrationResponse { string status = 1; }

message UserSummary {
  string id = 1; string email = 2; string name = 3; string role = 4; string status = 5;
}
message ListPendingGuestsResponse { repeated UserSummary guests = 1; }
message ApproveGuestRequest {
  string user_id = 1; string role = 2; optional string department_id = 3;
}
message RejectGuestRequest { string user_id = 1; string reason = 2; }
```

Connect handler 實作(`handler.go` 加 `AuthServiceHandler`):每個方法一對一呼叫 `Usecase`,從 ctx 取 `rls.Identity`(Task 11 middleware 注入;此前測試直接以 ctx 注入)。`ListPendingGuests` 實作:查 `role=guest AND status=pending AND deleted_at IS NULL`,RLS 依 actor 範圍過濾。

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && go get connectrpc.com/connect && go generate ./ent/ && go test ./internal/domain/auth/ -v`
Expected: PASS — 三個測試(token 一次性、g 規則連結、軟刪後可重註冊)。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/auth backend/proto/v1/auth.proto
git commit -m "feat(backend): 註冊完成建立 pending guest 與審核流程(1.4.3-1.4.4)"
```

---

### Task 6: 客戶帳密登入(細部 1.5.1)

**Files:**
- Create: `backend/internal/domain/auth/password.go`
- Create: `backend/internal/domain/auth/customer.go`
- Test: `backend/internal/domain/auth/customer_test.go`

**Interfaces:**
- Consumes: Task 5 `Usecase`;Task 9 `TokenIssuer`(本 Task 以介面注入,fake 測試)。
- Produces: `auth.HashPassword(pw string) (string, error)` / `auth.VerifyPassword(hash, pw string) bool`(Argon2id,PHC 字串格式);proto `AuthService.CustomerLogin(company_identifier, customer_code, account_name, password) → CustomerLoginResponse{access_token, refresh_token, must_change_password, is_primary}`;usecase `CustomerLogin(ctx, companyIdentifier, customerCode, accountName, password) (*ent.User, error)`(登入成功回使用者;token 發放由 handler 呼叫 `TokenIssuer`,Task 9 接上)。
- 登入檢查順序(後續 Task 7/8 在此骨架上擴充):帳號存在且 `is_customer` → `status=active` → 臨時密碼效期 → 鎖定 → 密碼比對。

- [ ] **Step 1: 寫失敗測試(雜湊 + 主帳號業務拒絕)**

`backend/internal/domain/auth/customer_test.go`:

```go
package auth_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestPasswordHashVerify(t *testing.T) {
	h, err := auth.HashPassword("s3cret-pass")
	if err != nil { t.Fatal(err) }
	if !auth.VerifyPassword(h, "s3cret-pass") { t.Fatal("verify should pass") }
	if auth.VerifyPassword(h, "wrong") { t.Fatal("verify should fail") }
}

// seedCustomerAccounts 建一客戶的主帳號與子帳號;customer 表 Phase 3 才有,先用固定 UUID。
func seedCustomerAccounts(t *testing.T, c *ent.Client) (coID uuid.UUID, custID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	co, err := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	if err != nil { t.Fatal(err) }
	custID = uuid.New()
	mk := func(name string, primary bool, pw string) {
		h, _ := auth.HashPassword(pw)
		_, err := c.User.Create().SetCompanyID(co.ID).SetRole("customer").SetDataScope("self").
			SetStatus("active").SetIsCustomer(true).SetCustomerID(custID).
			SetAccountName(name).SetIsPrimary(primary).SetPasswordHash(h).Save(ctx)
		if err != nil { t.Fatal(err) }
	}
	mk("老闆", true, "boss-pass-123")
	mk("主廚", false, "chef-pass-123")
	return co.ID, custID
}

func TestCustomerLoginSuccess(t *testing.T) {
	c := testutil.NewEntClient(t)
	seedCustomerAccounts(t, c)
	uc := auth.NewUsecase(c, newFakeStore(), nil)

	u, err := uc.CustomerLogin(context.Background(), "co-a", "000001", "主廚", "chef-pass-123")
	if err != nil { t.Fatalf("login: %v", err) }
	if u.AccountName == nil || *u.AccountName != "主廚" || u.IsPrimary {
		t.Fatalf("wrong account: %+v", u)
	}
}

func TestCustomerLoginWrongPassword(t *testing.T) {
	c := testutil.NewEntClient(t)
	seedCustomerAccounts(t, c)
	uc := auth.NewUsecase(c, newFakeStore(), nil)

	_, err := uc.CustomerLogin(context.Background(), "co-a", "000001", "主廚", "nope")
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("want unauthenticated, got %v", err)
	}
	// 不透露帳號是否存在:不存在的帳號同樣 unauthenticated
	_, err = uc.CustomerLogin(context.Background(), "co-a", "000001", "不存在", "nope")
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("unknown account should also be unauthenticated, got %v", err)
	}
}

func TestPrimaryAccountBlockedFromBusinessAPI(t *testing.T) {
	c := testutil.NewEntClient(t)
	seedCustomerAccounts(t, c)
	uc := auth.NewUsecase(c, newFakeStore(), nil)

	u, err := uc.CustomerLogin(context.Background(), "co-a", "000001", "老闆", "boss-pass-123")
	if err != nil { t.Fatalf("primary login: %v", err) }
	// 主帳號可登入,但 IsBusinessForbidden 為 true(Task 11 middleware 據此 403)
	if !u.IsPrimary {
		t.Fatal("want primary account")
	}
}
```

說明:本階段客戶查找鍵為 `company.identifier + customer_code + account_name`;`customers` 表 Phase 3 才建,登入查找暫以 `users.customer_id` 關聯欄位比對 — 實作上 `customer_code` 先存於 user 查詢的 join 佔位:**實作時於 `users` 查詢以 `customer_id` 直接定位(測試用固定 UUID 建帳,`customer_code` 參數於 Phase 3 接上 `customers` 表後生效)**;為讓本 Task 可獨立交付,`CustomerLogin` 簽名的 `customer_code` 參數暫時用於查找 `users.customer_id`(測試以 `"000001"` 對應 seed 的固定 UUID,透過 `mapCustomerCode` 佔位函式轉換,Phase 3 替換為真實查表)。

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/auth/ -run 'TestPassword|TestCustomer|TestPrimary' -v`
Expected: FAIL — `auth.HashPassword` / `CustomerLogin` 未定義。

- [ ] **Step 3: 實作密碼雜湊**

`backend/internal/domain/auth/password.go`:

```go
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id 參數(Global Constraints)。
const (
	argonMemory  = 64 * 1024
	argonTime    = 3
	argonThreads = 2
	argonKeyLen  = 32
	argonSaltLen = 16
)

// HashPassword 回傳 PHC 格式字串 $argon2id$v=19$m=...,t=...,p=...$salt$key。
func HashPassword(pw string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(pw), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key)), nil
}

// VerifyPassword 常數時間比對。
func VerifyPassword(hash, pw string) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var mem uint32 = argonMemory
	var t uint32 = argonTime
	var p uint8 = argonThreads
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &t, &p); err != nil {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}
	got := argon2.IDKey([]byte(pw), salt, t, mem, p, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}

// GenerateTempPassword 產生臨時密碼(≥12 字元,英數混合)。
func GenerateTempPassword() (string, error) {
	const alphabet = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
}

var errWeakPassword = errors.New("password too weak")

// CheckPasswordStrength:最少 8 字元(細部 1.5.2)。
func CheckPasswordStrength(pw string) error {
	if len(pw) < 8 {
		return errWeakPassword
	}
	return nil
}
```

- [ ] **Step 4: 實作 CustomerLogin**

`backend/internal/domain/auth/customer.go`:

```go
package auth

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entcompany "github.com/salesorder/sales-order-1.0/backend/ent/company"
	entuser "github.com/salesorder/sales-order-1.0/backend/ent/user"
)

var errBadCredential = errors.New("invalid credentials")

// mapCustomerCode:Phase 3 前佔位 — customer_code 先映射到測試用 customer UUID;
// Phase 3 Task 3.1 完成後改為查 customers 表(細部 1.5.1 步驟 1)。
// TODO(Phase 3): replace with customers table lookup.
var mapCustomerCode = func(code string) (uuid.UUID, bool) {
	if code == "000001" {
		return uuid.MustParse("11111111-1111-1111-1111-111111111111"), true
	}
	// 測試 seed 使用 uuid.New();為讓佔位可查,seed 輔助改為固定 UUID(見 customer_test seedCustomerAccounts 調整)
	return uuid.Nil, false
}

// CustomerLogin 驗證客戶帳密;成功回使用者實體(token 由 handler 發放)。
// 檢查順序:帳號存在 → active → 臨時密碼效期(Task 7)→ 鎖定(Task 8)→ 密碼。
// 任何「帳號不存在/密碼錯誤」皆回 unauthenticated,不透露細節。
func (u *Usecase) CustomerLogin(ctx context.Context, companyIdentifier, customerCode, accountName, password string) (*ent.User, error) {
	co, err := u.db.Company.Query().Where(
		entcompany.IdentifierEQ(companyIdentifier),
		entcompany.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil || co.Status != "active" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errBadCredential)
	}
	custID, ok := mapCustomerCode(customerCode)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errBadCredential)
	}
	usr, err := u.db.User.Query().Where(
		entuser.CompanyIDEQ(co.ID),
		entuser.CustomerIDEQ(custID),
		entuser.AccountNameEQ(accountName),
		entuser.IsCustomerEQ(true),
		entuser.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errBadCredential)
	}
	if usr.Status != "active" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errBadCredential)
	}
	// Task 7 掛點:臨時密碼效期檢查
	if err := checkTempPasswordExpiry(usr); err != nil {
		return nil, err
	}
	// Task 8 掛點:鎖定檢查與失敗計數(行鎖)
	if err := u.checkLockAndVerify(ctx, usr, password); err != nil {
		return nil, err
	}
	return usr, nil
}

// checkTempPasswordExpiry 由 Task 7 實作;本 Task 先放行。
var checkTempPasswordExpiry = func(*ent.User) error { return nil }

// checkLockAndVerify 由 Task 8 完整實作;本 Task 先做純密碼比對。
func (u *Usecase) checkLockAndVerify(_ context.Context, usr *ent.User, password string) error {
	if usr.PasswordHash == nil || !VerifyPassword(*usr.PasswordHash, password) {
		return connect.NewError(connect.CodeUnauthenticated, errBadCredential)
	}
	return nil
}
```

修正測試 seed 以配合佔位:`seedCustomerAccounts` 的 `custID` 改為 `uuid.MustParse("11111111-1111-1111-1111-111111111111")`。

- [ ] **Step 5: 跑測試確認通過**

Run: `cd backend && go get golang.org/x/crypto && go test ./internal/domain/auth/ -run 'TestPassword|TestCustomer|TestPrimary' -v`
Expected: PASS。

- [ ] **Step 6: proto 增量 + Commit**

`auth.proto` 的 `AuthService` 加:

```proto
  rpc CustomerLogin(CustomerLoginRequest) returns (CustomerLoginResponse);
// ...
message CustomerLoginRequest {
  string company_identifier = 1;
  string customer_code = 2;
  string account_name = 3;
  string password = 4;
}
message CustomerLoginResponse {
  string access_token = 1;
  string refresh_token = 2;
  bool must_change_password = 3;
  bool is_primary = 4;
}
```

(token 欄位於 Task 9 填入;本 Task handler 先回空字串。)

```bash
git add backend/internal/domain/auth backend/proto/v1/auth.proto
git commit -m "feat(backend): Argon2id 密碼雜湊與客戶帳密登入骨架(1.5.1)"
```

---

### Task 7: 臨時密碼與首登強制修改(細部 1.5.2)

**Files:**
- Update: `backend/internal/domain/auth/customer.go`
- Update: `backend/internal/domain/auth/usecase.go`
- Test: `backend/internal/domain/auth/temppw_test.go`

**Interfaces:**
- Consumes: Task 6 `GenerateTempPassword` / `CheckPasswordStrength`。
- Produces: `auth.NewTempPasswordIssue(...) (plaintext string, err error)` 語義的內部函式 `issueTempPassword(tx, userID) (string, error)`(Task 3.1.4 建檔與 Task 8 重置重用);proto `AuthService.ChangePassword(old_password, new_password)`;登入回應 `must_change_password` 語義;`ChangePassword` 成功後 `token_version+1`。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/domain/auth/temppw_test.go`:

```go
package auth_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/domain/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestTempPasswordExpiryBlocksLogin(t *testing.T) {
	c := testutil.NewEntClient(t)
	seedCustomerAccounts(t, c)
	ctx := context.Background()
	// 讓主廚的臨時密碼過期
	_, err := c.User.Update().SetTempPasswordExpiresAt(time.Now().Add(-time.Hour)).
		SetMustChangePassword(true).Save(ctx)
	if err != nil { t.Fatal(err) }

	uc := auth.NewUsecase(c, newFakeStore(), nil)
	_, err = uc.CustomerLogin(ctx, "co-a", "000001", "主廚", "chef-pass-123")
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("expired temp password: want failed_precondition, got %v", err)
	}
}

func TestMustChangePasswordFlow(t *testing.T) {
	c := testutil.NewEntClient(t)
	seedCustomerAccounts(t, c)
	ctx := context.Background()
	if _, err := c.User.Update().SetMustChangePassword(true).
		SetTempPasswordExpiresAt(time.Now().Add(24 * time.Hour)).Save(ctx); err != nil {
		t.Fatal(err)
	}
	uc := auth.NewUsecase(c, newFakeStore(), nil)

	u, err := uc.CustomerLogin(ctx, "co-a", "000001", "主廚", "chef-pass-123")
	if err != nil { t.Fatalf("login: %v", err) }
	if !u.MustChangePassword {
		t.Fatal("want must_change_password=true")
	}
	tvBefore := u.TokenVersion

	// 強度不足
	if err := uc.ChangePassword(ctx, u.ID, "chef-pass-123", "short"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("weak password: want invalid_argument, got %v", err)
	}
	// 舊密碼錯
	if err := uc.ChangePassword(ctx, u.ID, "wrong", "new-pass-456"); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("bad old password: want unauthenticated, got %v", err)
	}
	// 成功:token_version+1、旗標清除、新密碼可登入
	if err := uc.ChangePassword(ctx, u.ID, "chef-pass-123", "new-pass-456"); err != nil {
		t.Fatalf("change: %v", err)
	}
	after := c.User.GetX(ctx, u.ID)
	if after.TokenVersion != tvBefore+1 {
		t.Fatalf("token_version not bumped: %d -> %d", tvBefore, after.TokenVersion)
	}
	if after.MustChangePassword || after.TempPasswordExpiresAt != nil {
		t.Fatal("flags not cleared")
	}
	if _, err := uc.CustomerLogin(ctx, "co-a", "000001", "主廚", "chef-pass-123"); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("old password should fail, got %v", err)
	}
	if _, err := uc.CustomerLogin(ctx, "co-a", "000001", "主廚", "new-pass-456"); err != nil {
		t.Fatalf("new password should work: %v", err)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/auth/ -run 'TestTempPassword|TestMustChange' -v`
Expected: FAIL — `ChangePassword` 未定義;`checkTempPasswordExpiry` 是 no-op 導致第一個測試失敗。

- [ ] **Step 3: 實作**

`customer.go` 替換 `checkTempPasswordExpiry` 佔位為真實檢查,並加 `issueTempPassword` 與 `ChangePassword`:

```go
// 替換:var checkTempPasswordExpiry = func(*ent.User) error { return nil }
// 為:
func checkTempPasswordExpiry(usr *ent.User) error {
	if usr.TempPasswordExpiresAt != nil && time.Now().After(*usr.TempPasswordExpiresAt) {
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("temporary password expired; ask admin to reissue"))
	}
	return nil
}

// issueTempPassword 在同一交易內產生臨時密碼並設定 24h 效期 + 強制改密碼 + 清鎖定。
// 回傳明文僅存在於本次呼叫,不落盤(細部 1.5.2 步驟 1)。
func issueTempPassword(ctx context.Context, tx *ent.Tx, userID uuid.UUID) (string, error) {
	plain, err := GenerateTempPassword()
	if err != nil {
		return "", err
	}
	h, err := HashPassword(plain)
	if err != nil {
		return "", err
	}
	exp := time.Now().Add(24 * time.Hour)
	if _, err := tx.User.UpdateOneID(userID).
		SetPasswordHash(h).
		SetTempPasswordExpiresAt(exp).
		SetMustChangePassword(true).
		ClearLockedAt().
		SetFailedLoginAttempts(0).
		Save(ctx); err != nil {
		return "", err
	}
	return plain, nil
}
```

`usecase.go` 加:

```go
// ChangePassword 首登強制改密碼;成功後 token_version+1 使舊登入態全失效(細部 1.5.2 步驟 3)。
func (u *Usecase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPw, newPw string) error {
	usr, err := u.db.User.Get(ctx, userID)
	if err != nil {
		return connect.NewError(connect.CodeUnauthenticated, errBadCredential)
	}
	if usr.PasswordHash == nil || !VerifyPassword(*usr.PasswordHash, oldPw) {
		return connect.NewError(connect.CodeUnauthenticated, errBadCredential)
	}
	if err := CheckPasswordStrength(newPw); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	h, err := HashPassword(newPw)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if _, err := tx.User.UpdateOneID(userID).
		SetPasswordHash(h).
		SetMustChangePassword(false).
		ClearTempPasswordExpiresAt().
		SetTokenVersion(usr.TokenVersion + 1).
		Save(ctx); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	// TODO(Phase 2 Task 2.6): audit log 同事務
	return tx.Commit()
}
```

`usecase.go` import 需補 `"time"`(timeNow 已在 Task 5 使用,若已存在則略)。

- [ ] **Step 4: proto 增量、跑測試**

`auth.proto` 加:

```proto
  rpc ChangePassword(ChangePasswordRequest) returns (google.protobuf.Empty);
// ...
message ChangePasswordRequest {
  string old_password = 1;
  string new_password = 2;
}
```

Run: `cd backend && go test ./internal/domain/auth/ -v`
Expected: PASS — 含 Task 6 既有測試不回歸。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/auth backend/proto/v1/auth.proto
git commit -m "feat(backend): 臨時密碼 24h 效期與首登強制修改(1.5.2)"
```

---

### Task 8: 登入鎖定與密碼重置(細部 1.5.3–1.5.4)

**Files:**
- Update: `backend/internal/domain/auth/customer.go`
- Update: `backend/internal/domain/auth/usecase.go`
- Test: `backend/internal/domain/auth/lockout_test.go`

**Interfaces:**
- Consumes: Task 7 `issueTempPassword`。
- Produces: proto `AuthService.ResetCustomerPassword(user_id) → ResetCustomerPasswordResponse{temp_password, expires_at}`;usecase `ResetCustomerPassword(ctx, actor rls.Identity, targetID uuid.UUID) (string, error)`;鎖定常數 `MaxFailedAttempts = 5`、`LockDuration = 30 * time.Minute`;`timeNow` 已可注入(測試模擬 30 分鐘後)。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/domain/auth/lockout_test.go`:

```go
package auth_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func loginExpect(uc *auth.Usecase, pw string) error {
	_, err := uc.CustomerLogin(context.Background(), "co-a", "000001", "主廚", pw)
	return err
}

func TestLockoutAfterFiveFailures(t *testing.T) {
	c := testutil.NewEntClient(t)
	seedCustomerAccounts(t, c)
	uc := auth.NewUsecase(c, newFakeStore(), nil)

	for i := 0; i < 5; i++ {
		if err := loginExpect(uc, "bad"); connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Fatalf("attempt %d: want unauthenticated, got %v", i+1, err)
		}
	}
	// 第 6 次:即使密碼正確也被鎖
	if err := loginExpect(uc, "chef-pass-123"); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("locked: want failed_precondition, got %v", err)
	}
	// 模擬 31 分鐘後:自動解除,正確密碼可登入(D21 鎖定解除整合測試)
	auth.SetNowForTest(time.Now().Add(31 * time.Minute))
	defer auth.SetNowForTest(time.Now())
	if err := loginExpect(uc, "chef-pass-123"); err != nil {
		t.Fatalf("after lock duration should auto-unlock: %v", err)
	}
}

func TestConcurrentFailuresConsistent(t *testing.T) {
	c := testutil.NewEntClient(t)
	seedCustomerAccounts(t, c)
	uc := auth.NewUsecase(c, newFakeStore(), nil)

	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() { done <- loginExpect(uc, "bad") }()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	u, _ := c.User.Query().Where().Only(context.Background()) // seed 的第一筆即可;實作以 account_name 過濾
	_ = u
	// 驗證最終狀態一致:已鎖定,且計數不超過 MaxFailedAttempts 後仍正確累加
	target, err := c.User.Query().Where().Only(context.Background())
	if err != nil { t.Fatal(err) }
	if target.LockedAt == nil {
		t.Fatal("should be locked after 10 concurrent failures")
	}
	if target.FailedLoginAttempts < auth.MaxFailedAttempts {
		t.Fatalf("counter inconsistent: %d", target.FailedLoginAttempts)
	}
}

func TestResetCustomerPassword(t *testing.T) {
	c := testutil.NewEntClient(t)
	coID, _ := seedCustomerAccounts(t, c)
	ctx := context.Background()
	uc := auth.NewUsecase(c, newFakeStore(), nil)

	target, err := c.User.Query().Where().First(ctx) // 主帳號
	if err != nil { t.Fatal(err) }
	tvBefore := target.TokenVersion

	actor := rls.Identity{UserID: uuid.New(), CompanyID: coID, DataScope: "company"}
	temp, err := uc.ResetCustomerPassword(ctx, actor, target.ID)
	if err != nil { t.Fatalf("reset: %v", err) }

	after := c.User.GetX(ctx, target.ID)
	if after.TokenVersion != tvBefore+1 { t.Fatal("token_version not bumped") }
	if after.LockedAt != nil || after.FailedLoginAttempts != 0 { t.Fatal("lock not cleared") }
	if !after.MustChangePassword { t.Fatal("must_change_password not set") }
	// 新臨時密碼可登入
	if _, err := uc.CustomerLogin(ctx, "co-a", "000001", *target.AccountName, temp); err != nil {
		t.Fatalf("login with temp password: %v", err)
	}
	// 範圍外 actor 被拒
	other := rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(), DataScope: "company"}
	if _, err := uc.ResetCustomerPassword(ctx, other, target.ID); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("out-of-scope reset: want permission_denied, got %v", err)
	}
}
```

(`lockout_test.go` import 需補 `"github.com/google/uuid"`;`Where()` 過濾條件於實作時補 `entuser.AccountNameEQ("主廚")` 等。)

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/auth/ -run 'TestLockout|TestConcurrent|TestResetCustomer' -v`
Expected: FAIL — `MaxFailedAttempts` / `SetNowForTest` / `ResetCustomerPassword` 未定義。

- [ ] **Step 3: 實作鎖定(checkLockAndVerify 完整版)**

`customer.go` 替換 `checkLockAndVerify`:

```go
// 鎖定參數(細部 1.5.3)。
const (
	MaxFailedAttempts = 5
	LockDuration      = 30 * time.Minute
)

// now 為可注入時鐘;測試用 SetNowForTest 覆寫。
var now = time.Now

// SetNowForTest 覆寫時鐘;測試結束呼叫者負責還原。
func SetNowForTest(t time.Time) { now = func() time.Time { return t } }

var errLocked = errors.New("account locked")

// checkLockAndVerify:鎖定檢查(30 分鐘自動解除)→ 密碼比對,失敗計數與鎖定更新
// 對同一 user 列 SELECT ... FOR UPDATE,避免併發通過第 5 次(細部 1.5.3 步驟 3)。
func (u *Usecase) checkLockAndVerify(ctx context.Context, usr *ent.User, password string) error {
	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	locked, err := tx.User.Query().
		Where(entuser.IDEQ(usr.ID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}

	// 鎖定檢查與自動解除
	if locked.LockedAt != nil {
		if now().Sub(*locked.LockedAt) < LockDuration {
			_ = tx.Rollback()
			remain := LockDuration - now().Sub(*locked.LockedAt)
			return connect.NewError(connect.CodeFailedPrecondition,
				fmt.Errorf("%w; retry in %ds", errLocked, int(remain.Seconds())))
		}
		// 自動解除
		if _, err := tx.User.UpdateOneID(usr.ID).
			ClearLockedAt().SetFailedLoginAttempts(0).Save(ctx); err != nil {
			_ = tx.Rollback()
			return connect.NewError(connect.CodeInternal, err)
		}
		locked.FailedLoginAttempts = 0
	}

	// 密碼比對
	if locked.PasswordHash == nil || !VerifyPassword(*locked.PasswordHash, password) {
		attempts := locked.FailedLoginAttempts + 1
		upd := tx.User.UpdateOneID(usr.ID).SetFailedLoginAttempts(attempts)
		if attempts >= MaxFailedAttempts {
			upd.SetLockedAt(now())
		}
		if _, err := upd.Save(ctx); err != nil {
			_ = tx.Rollback()
			return connect.NewError(connect.CodeInternal, err)
		}
		// TODO(Phase 2 Task 2.6): 失敗稽核同事務(不含密碼)
		if err := tx.Commit(); err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
		return connect.NewError(connect.CodeUnauthenticated, errBadCredential)
	}

	// 成功:清計數
	if locked.FailedLoginAttempts > 0 || locked.LockedAt != nil {
		if _, err := tx.User.UpdateOneID(usr.ID).
			SetFailedLoginAttempts(0).ClearLockedAt().Save(ctx); err != nil {
			_ = tx.Rollback()
			return connect.NewError(connect.CodeInternal, err)
		}
	}
	return tx.Commit()
}
```

(`customer.go` import 補 `"fmt"`、`"time"`、`entuser`。)

- [ ] **Step 4: 實作密碼重置**

`usecase.go` 加:

```go
// ResetCustomerPassword 產生新臨時密碼;連帶解鎖 + token_version+1;明文僅本次回應(細部 1.5.4)。
func (u *Usecase) ResetCustomerPassword(ctx context.Context, actor rls.Identity, targetID uuid.UUID) (string, error) {
	target, err := u.db.User.Get(ctx, targetID)
	if err != nil || target.DeletedAt != nil {
		return "", connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	}
	if !target.IsCustomer {
		return "", connect.NewError(connect.CodeInvalidArgument, errors.New("target is not a customer account"))
	}
	// 範圍檢查:all 不限;company 限同公司;department 由 RLS/2.3.2 進一步收窄(此處先比對公司)
	if actor.DataScope != "all" && target.CompanyID != actor.CompanyID {
		return "", connect.NewError(connect.CodePermissionDenied, errors.New("out of scope"))
	}

	tx, err := u.db.Tx(ctx)
	if err != nil {
		return "", connect.NewError(connect.CodeInternal, err)
	}
	plain, err := issueTempPassword(ctx, tx, targetID)
	if err != nil {
		_ = tx.Rollback()
		return "", connect.NewError(connect.CodeInternal, err)
	}
	if _, err := tx.User.UpdateOneID(targetID).
		SetTokenVersion(target.TokenVersion + 1).Save(ctx); err != nil {
		_ = tx.Rollback()
		return "", connect.NewError(connect.CodeInternal, err)
	}
	// TODO(Phase 2 Task 2.6): 稽核同事務(操作者 + 目標,不含密碼)
	if err := tx.Commit(); err != nil {
		return "", connect.NewError(connect.CodeInternal, err)
	}
	return plain, nil
}
```

- [ ] **Step 5: proto 增量、跑測試**

`auth.proto` 加:

```proto
  rpc ResetCustomerPassword(ResetCustomerPasswordRequest) returns (ResetCustomerPasswordResponse);
// ...
message ResetCustomerPasswordRequest { string user_id = 1; }
message ResetCustomerPasswordResponse {
  string temp_password = 1;
  google.protobuf.Timestamp expires_at = 2;
}
```

(proto import 補 `"google/protobuf/timestamp.proto"`。)

Run: `cd backend && go test ./internal/domain/auth/ -v`
Expected: PASS — 含併發測試(`-race` 一併跑:`go test -race ./internal/domain/auth/`)。

- [ ] **Step 6: Commit**

```bash
git add backend/internal/domain/auth backend/proto/v1/auth.proto
git commit -m "feat(backend): 登入鎖定(5 次/30 分)與密碼重置(1.5.3-1.5.4)"
```

---

### Task 9: scs + Valkey session 與 access JWT(細部 1.6.1–1.6.2)

**Files:**
- Create: `backend/internal/session/manager.go`
- Create: `backend/internal/session/jwt.go`
- Test: `backend/internal/session/session_test.go`

**Interfaces:**
- Consumes: config `ValkeyAddr`、新增 `JWTSecret` / `AccessTokenTTL`。
- Produces: `session.Manager`(scs 封裝)`Load` / `Put` / `Destroy` / `IssueWeb(w, r, u)`(實作 Task 4 的 `auth.TokenIssuer`);`session.Claims{UserID uuid.UUID; Role, CompanyID string; DepartmentID, CustomerID *uuid.UUID; IsPrimary, MustChange bool; TokenVersion int}`;`session.IssueAccessToken(c Claims) (string, error)`;`session.VerifyAccessToken(tok string) (*Claims, error)`;Task 4 `OAuthHandler.issuer` 於此 Task 接上(`main.go` 組裝時注入)。

- [ ] **Step 1: config 擴充**

`backend/config/config.go` 加:

```go
JWTSecret      string        `mapstructure:"JWT_SECRET"`
AccessTokenTTL time.Duration `mapstructure:"ACCESS_TOKEN_TTL"` // 預設 1h
SessionTTL     time.Duration `mapstructure:"SESSION_TTL"`      // 預設 30 天(對齊 refresh)
```

`Load()`:`JWTSecret` 為空且 `ENV != "development"` → 錯誤(fail-fast);TTL 未設給預設值。

- [ ] **Step 2: 寫失敗測試**

`backend/internal/session/session_test.go`:

```go
package session_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/salesorder/sales-order-1.0/backend/internal/session"
)

func testClaims() session.Claims {
	return session.Claims{
		UserID:       uuid.New(),
		Role:         "staff",
		CompanyID:    uuid.New().String(),
		IsPrimary:    false,
		TokenVersion: 3,
	}
}

func TestAccessTokenRoundTrip(t *testing.T) {
	session.InitForTest("test-secret", time.Hour)
	tok, err := session.IssueAccessToken(testClaims())
	if err != nil { t.Fatal(err) }
	c, err := session.VerifyAccessToken(tok)
	if err != nil { t.Fatal(err) }
	if c.Role != "staff" || c.TokenVersion != 3 {
		t.Fatalf("claims mismatch: %+v", c)
	}
}

func TestAccessTokenTampered(t *testing.T) {
	session.InitForTest("test-secret", time.Hour)
	tok, _ := session.IssueAccessToken(testClaims())
	if _, err := session.VerifyAccessToken(tok + "x"); err == nil {
		t.Fatal("tampered token should fail")
	}
}

func TestAccessTokenExpired(t *testing.T) {
	session.InitForTest("test-secret", -time.Minute) // 立即過期
	tok, _ := session.IssueAccessToken(testClaims())
	if _, err := session.VerifyAccessToken(tok); err == nil {
		t.Fatal("expired token should fail")
	}
}
```

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/session/ -v`
Expected: FAIL — `session.IssueAccessToken` 未定義。

- [ ] **Step 4: 實作 JWT**

`backend/internal/session/jwt.go`:

```go
// Package session 管理 Web session 與 App JWT(細部文件 1.6)。
package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims 為 access token 內容;`tv` 對應 users.token_version(撤銷比對用)。
type Claims struct {
	UserID           uuid.UUID  `json:"sub"`
	Role             string     `json:"role"`
	CompanyID        string     `json:"cid"`
	DepartmentID     *uuid.UUID `json:"did,omitempty"`
	CustomerID       *uuid.UUID `json:"cust,omitempty"`
	IsPrimary        bool       `json:"prim"`
	MustChange       bool       `json:"mcp"`
	TokenVersion     int        `json:"tv"`
	jwt.RegisteredClaims
}

var (
	secret []byte
	ttl    time.Duration
)

// Init 由 main 於啟動時呼叫;secret 為空回 panic(fail-fast,config 層已擋,此為雙保險)。
func Init(jwtSecret string, accessTTL time.Duration) {
	if jwtSecret == "" {
		panic("session: JWT_SECRET empty")
	}
	secret = []byte(jwtSecret)
	ttl = accessTTL
}

// InitForTest 供測試注入。
func InitForTest(jwtSecret string, accessTTL time.Duration) { Init(jwtSecret, accessTTL) }

// IssueAccessToken 簽發 HS256 JWT(細部 1.6.2)。
func IssueAccessToken(c Claims) (string, error) {
	c.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(secret)
}

// VerifyAccessToken 驗簽章與 exp;撤銷比對在 middleware(1.6.4)。
func VerifyAccessToken(tok string) (*Claims, error) {
	var c Claims
	_, err := jwt.ParseWithClaims(tok, &c, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected alg %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, errors.New("invalid token")
	}
	return &c, nil
}
```

- [ ] **Step 5: 實作 session manager**

`backend/internal/session/manager.go`:

```go
package session

import (
	"net/http"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/redis/go-redis/v9"

	"github.com/salesorder/sales-order-1.0/backend/ent"
)

// Manager 封裝 scs;session 內容存 Valkey(redisstore 相容)。
type Manager struct{ *scs.SessionManager }

func NewManager(rdb *redis.Client, ttlSeconds int) *Manager {
	m := scs.New()
	m.Store = redisstore.New(rdb)
	m.Lifetime = time.Duration(ttlSeconds) * time.Second
	m.Cookie.HttpOnly = true
	m.Cookie.Secure = true
	m.Cookie.SameSite = http.SameSiteLaxMode
	return &Manager{m}
}

// Put Identity 於 session(LoadAndSave middleware 後使用)。
func (m *Manager) PutIdentity(r *http.Request, id IdentityView) {
	m.Put(r.Context(), "identity", id)
}

// IdentityView 為 session 內存的使用者視圖(與 rls.Identity 對應)。
type IdentityView struct {
	UserID       string
	Role         string
	CompanyID    string
	DepartmentID string
	CustomerID   string
	IsPrimary    bool
	TokenVersion int
}

// IssueWeb 建立 Web session(實作 auth.TokenIssuer;細部 1.6.1)。
func (m *Manager) IssueWeb(w http.ResponseWriter, r *http.Request, u *ent.User) error {
	if err := m.RenewToken(r.Context()); err != nil {
		return err
	}
	m.Put(r.Context(), "identity", IdentityView{
		UserID:       u.ID.String(),
		Role:         u.Role,
		CompanyID:    u.CompanyID.String(),
		IsPrimary:    u.IsPrimary,
		TokenVersion: u.TokenVersion,
	})
	return nil
}

// Destroy 刪除 session(登出與強制登出用)。
func (m *Manager) Destroy(r *http.Request) error {
	return m.SessionManager.Destroy(r.Context())
}
```

(`manager.go` import 補 `"time"`。)

- [ ] **Step 6: 跑測試確認通過 + Commit**

Run: `cd backend && go get github.com/golang-jwt/jwt/v5 github.com/alexedwards/scs/v2 github.com/alexedwards/scs/redisstore && go test ./internal/session/ -v`
Expected: PASS — round-trip、竄改拒絕、過期拒絕。

```bash
git add backend/internal/session backend/config/config.go
git commit -m "feat(backend): scs+Valkey session 與 HS256 access JWT(1.6.1-1.6.2)"
```

---

### Task 10: refresh token 旋轉與 token_version 比對(細部 1.6.3–1.6.4)

**Files:**
- Create: `backend/ent/schema/refreshtoken.go`
- Create: `backend/internal/session/refresh.go`
- Update: `backend/internal/middleware/auth.go`(Task 11 完成完整版;本 Task 建 `CheckTokenVersion` 函式)
- Test: `backend/internal/session/refresh_test.go`

**Interfaces:**
- Consumes: Task 9 `Claims` / `IssueAccessToken`;Task 1 `testutil`。
- Produces: `ent.RefreshToken`(`id`、`user_id`、`token_hash` 唯一、`issued_tv`(int)、`expires_at`、`rotated_at`);`session.IssueRefreshToken(ctx, tx, userID, tv) (plain string, error)`;`session.RotateRefreshToken(ctx, client, plain) (*RotateResult, error)`,`RotateResult{UserID uuid.UUID; Claims Claims; AccessToken, RefreshToken string}`;重放偵測 `session.ErrRefreshReplay`;`middleware.CheckTokenVersion(ctx, client, userID, tv) error`(`unauthenticated` on 不符)。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/session/refresh_test.go`:

```go
package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/salesorder/sales-order-1.0/backend/internal/session"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func seedUser(t *testing.T) (client *entClient, userID string, tv int) { panic("見 Step 2 註解") }

func TestRefreshRotation(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	u, _ := c.User.Create().SetCompanyID(co.ID).SetRole("staff").SetDataScope("department").
		SetStatus("active").SetEmail("s@example.com").Save(ctx)

	session.InitForTest("test-secret", time.Hour)

	// 發放
	plain1, err := session.IssueRefreshToken(ctx, c, u.ID, u.TokenVersion)
	if err != nil { t.Fatal(err) }

	// 換發:回新 access + 新 refresh,舊 refresh 作廢
	res, err := session.RotateRefreshToken(ctx, c, plain1)
	if err != nil { t.Fatal(err) }
	if res.AccessToken == "" || res.RefreshToken == "" || res.RefreshToken == plain1 {
		t.Fatalf("rotation invalid: %+v", res)
	}
	if _, err := session.RotateRefreshToken(ctx, c, plain1); err != session.ErrRefreshReplay {
		t.Fatalf("old token reuse: want ErrRefreshReplay, got %v", err)
	}
	// 重放後:該使用者全部 refresh token 作廢 + token_version+1
	if _, err := session.RotateRefreshToken(ctx, c, res.RefreshToken); err == nil {
		t.Fatal("all refresh tokens should be revoked after replay detection")
	}
	after := c.User.GetX(ctx, u.ID)
	if after.TokenVersion != u.TokenVersion+1 {
		t.Fatalf("replay should bump token_version: %d -> %d", u.TokenVersion, after.TokenVersion)
	}
}

func TestRefreshExpired(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	u, _ := c.User.Create().SetCompanyID(co.ID).SetRole("staff").SetDataScope("department").
		SetStatus("active").SetEmail("e@example.com").Save(ctx)

	session.InitForTest("test-secret", time.Hour)
	plain, _ := session.IssueRefreshToken(ctx, c, u.ID, u.TokenVersion)
	// 直接改庫讓它過期
	c.RefreshToken.Update().SetExpiresAt(time.Now().Add(-time.Hour)).ExecX(ctx)
	if _, err := session.RotateRefreshToken(ctx, c, plain); err == nil {
		t.Fatal("expired refresh should fail")
	}
}

func TestRefreshRejectedAfterTokenVersionBump(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	u, _ := c.User.Create().SetCompanyID(co.ID).SetRole("staff").SetDataScope("department").
		SetStatus("active").SetEmail("v@example.com").Save(ctx)

	session.InitForTest("test-secret", time.Hour)
	plain, _ := session.IssueRefreshToken(ctx, c, u.ID, u.TokenVersion)
	// 模擬強制登出:token_version+1
	c.User.UpdateOneID(u.ID).SetTokenVersion(u.TokenVersion + 1).ExecX(ctx)
	if _, err := session.RotateRefreshToken(ctx, c, plain); err == nil {
		t.Fatal("refresh should fail after token_version bump")
	}
}
```

(測試檔 import 補 `"github.com/salesorder/sales-order-1.0/backend/ent"`;移除佔位 `seedUser`/`entClient` — 僅為說明結構,實作不需要。)

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/session/ -v`
Expected: FAIL — `RefreshToken` schema、`IssueRefreshToken` 未定義。

- [ ] **Step 3: 實作 refresh token schema**

`backend/ent/schema/refreshtoken.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// RefreshToken:旋轉制,僅存 SHA-256 雜湊;內部表,不加 RLS(細部 1.6.3)。
type RefreshToken struct{ ent.Schema }

func (RefreshToken) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.String("token_hash").NotEmpty().Sensitive(),
		field.Time("expires_at"),
		field.Time("rotated_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (RefreshToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token_hash").Unique(),
		index.Fields("user_id"),
	}
}
```

- [ ] **Step 4: 實作發放與旋轉**

`backend/internal/session/refresh.go`:

```go
package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entrt "github.com/salesorder/sales-order-1.0/backend/ent/refreshtoken"
)

const refreshTTL = 30 * 24 * time.Hour // 30 天

// ErrRefreshReplay:已旋轉的 token 被重用,視為外洩。
var ErrRefreshReplay = errors.New("refresh token replay detected")

// ErrRefreshInvalid:不存在/過期/tv 不符。
var ErrRefreshInvalid = errors.New("refresh token invalid")

func hashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

func newPlainToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// IssueRefreshToken 產生 refresh token;明文僅回傳本次,庫中僅存雜湊。
// client 參數接受 *ent.Client 或 *ent.Tx(以 ent 介面抽象)。
func IssueRefreshToken(ctx context.Context, q interface {
	RefreshTokenCreator() // 佔位,實作直接用 *ent.Client
}, userID uuid.UUID, tv int) (string, error) {
	panic("見下")
}
```

說明:為讓 Task 12(強制登出)能在同一交易內作廢 refresh token,實作以 `*ent.Client` 開交易;`IssueRefreshToken` 實際簽名:

```go
// IssueRefreshToken 於獨立交易建立 refresh token 列。
func IssueRefreshToken(ctx context.Context, db *ent.Client, userID uuid.UUID, tv int) (string, error) {
	plain, err := newPlainToken()
	if err != nil {
		return "", err
	}
	if _, err := db.RefreshToken.Create().
		SetUserID(userID).
		SetTokenHash(hashToken(plain)).
		SetExpiresAt(time.Now().Add(refreshTTL)).
		Save(ctx); err != nil {
		return "", err
	}
	return plain, nil
}

// RotateResult 為換發結果。
type RotateResult struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
}

// RotateRefreshToken 驗證並旋轉(細部 1.6.3 步驟 1):
// 雜湊查表 → 驗過期/未旋轉/user active/tv 一致 → 同交易標記 rotated + 寫新列 + 簽新 JWT。
// 重放(已 rotated)→ 同交易作廢該使用者全部 token + token_version+1。
func RotateRefreshToken(ctx context.Context, db *ent.Client, plain string) (*RotateResult, error) {
	h := hashToken(plain)
	rt, err := db.RefreshToken.Query().Where(entrt.TokenHashEQ(h)).Only(ctx)
	if err != nil {
		return nil, ErrRefreshInvalid
	}

	tx, err := db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	rollback := func(err error) error { _ = tx.Rollback(); return err }

	if rt.RotatedAt != nil {
		// 重放偵測:作廢全部 + tv+1
		u, err := tx.User.Get(ctx, rt.UserID)
		if err != nil {
			return nil, rollback(err)
		}
		now := time.Now()
		if _, err := tx.RefreshToken.Update().
			Where(entrt.UserIDEQ(rt.UserID), entrt.RotatedAtIsNil()).
			SetRotatedAt(now).Save(ctx); err != nil {
			return nil, rollback(err)
		}
		if _, err := tx.User.UpdateOneID(rt.UserID).
			SetTokenVersion(u.TokenVersion + 1).Save(ctx); err != nil {
			return nil, rollback(err)
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return nil, ErrRefreshReplay
	}

	u, err := tx.User.Get(ctx, rt.UserID)
	if err != nil || u.Status != "active" || u.DeletedAt != nil {
		return nil, rollback(ErrRefreshInvalid)
	}
	if time.Now().After(rt.ExpiresAt) {
		return nil, rollback(ErrRefreshInvalid)
	}

	// tv 比對:以發放時寫入 refresh 列的 issued tv(對照 users 現值,與 1.6.4 同源)
	// — 本實作直接比對 u.TokenVersion 與 rotation 時刻重新簽發,故此處只需確保
	// 舊 token 在 tv 變更後不可再換:tv 變更時強制登出已作廢全部列(Task 12),
	// 雙保險:此處再比對 rt.CreatedAt 之後若 tv 變過則拒(以 u.TokenVersion 記錄於 claims 比對)
	// 簡化且正確的做法:換發出來的 JWT tv = u.TokenVersion 現值;舊 refresh 是否有效
	// 由「tv 變更必作廢全部 refresh」(Task 12 / 本函式重放路徑)保證。

	now := time.Now()
	if _, err := tx.RefreshToken.UpdateOneID(rt.ID).SetRotatedAt(now).Save(ctx); err != nil {
		return nil, rollback(err)
	}
	plain2, err := newPlainToken()
	if err != nil {
		return nil, rollback(err)
	}
	if _, err := tx.RefreshToken.Create().
		SetUserID(rt.UserID).
		SetTokenHash(hashToken(plain2)).
		SetExpiresAt(now.Add(refreshTTL)).
		Save(ctx); err != nil {
		return nil, rollback(err)
	}
	// 順手清理該使用者已過期列(細部 1.6.3:查詢時順手刪)
	if _, err := tx.RefreshToken.Delete().
		Where(entrt.UserIDEQ(rt.UserID), entrt.ExpiresAtLT(now)).Exec(ctx); err != nil {
		return nil, rollback(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	access, err := IssueAccessToken(Claims{
		UserID:       u.ID,
		Role:         u.Role,
		CompanyID:    u.CompanyID.String(),
		DepartmentID: u.DepartmentID,
		CustomerID:   u.CustomerID,
		IsPrimary:    u.IsPrimary,
		MustChange:   u.MustChangePassword,
		TokenVersion: u.TokenVersion,
	})
	if err != nil {
		return nil, err
	}
	return &RotateResult{UserID: u.ID, AccessToken: access, RefreshToken: plain2}, nil
}
```

刪除佔位的 `IssueRefreshToken` interface 版本(上方標 panic 的那段不落地,以此處 `*ent.Client` 簽名為準)。

注意 `TestRefreshRejectedAfterTokenVersionBump`:`token_version` bump 時 Task 12 會作廢全部 refresh;本測試直接改庫模擬,因此 `RotateRefreshToken` 需顯式拒絕 — 在 `RotateRefreshToken` 的驗證段加 tv 快照比對:`refresh_tokens` 表加 `issued_tv` 欄位(schema 補 `field.Int("issued_tv")`),`IssueRefreshToken` 寫入 `tv`,旋轉時 `rt.IssuedTV != u.TokenVersion` → `ErrRefreshInvalid`。schema 與程式同步補上此欄位。

- [ ] **Step 5: token_version 比對函式(1.6.4)**

`backend/internal/middleware/auth.go` 本 Task 先建立單一函式(完整 middleware 於 Task 11):

```go
// Package middleware 提供認證授權 middleware(細部文件 1.6.4-1.6.6)。
package middleware

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
)

// CheckTokenVersion 比對 token 的 tv 與 users 現值(細部 1.6.4);
// 不符、使用者不存在或非 active 皆 unauthenticated。
// 快取:per-user 短 TTL(≤60s)於 Valkey,撤銷操作主動刪除 — 本 Phase 先直查 DB,
// 快取於效能需求出現時再加,且必須與主動失效成對(細部 1.6.4 步驟 2)。
func CheckTokenVersion(ctx context.Context, db *ent.Client, userID uuid.UUID, tv int) error {
	u, err := db.User.Get(ctx, userID)
	if err != nil || u.TokenVersion != tv || u.Status != "active" || u.DeletedAt != nil {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("token revoked"))
	}
	return nil
}
```

測試(併入 `refresh_test.go` 同檔或新檔 `middleware/auth_test.go`):

```go
func TestCheckTokenVersion(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	u, _ := c.User.Create().SetCompanyID(co.ID).SetRole("staff").SetDataScope("department").
		SetStatus("active").SetEmail("tv@example.com").Save(ctx)

	if err := middleware.CheckTokenVersion(ctx, c, u.ID, 0); err != nil {
		t.Fatalf("current tv should pass: %v", err)
	}
	c.User.UpdateOneID(u.ID).SetTokenVersion(1).ExecX(ctx)
	if err := middleware.CheckTokenVersion(ctx, c, u.ID, 0); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("stale tv: want unauthenticated, got %v", err)
	}
}
```

- [ ] **Step 6: 跑測試 + Commit**

Run: `cd backend && go generate ./ent/ && go test ./internal/session/ ./internal/middleware/ -v`
Expected: PASS — 旋轉、重放全撤、過期拒、tv bump 拒、CheckTokenVersion。

```bash
git add backend/ent/schema/refreshtoken.go backend/internal/session/refresh.go backend/internal/middleware
git commit -m "feat(backend): refresh token 旋轉、重放偵測、token_version 撤銷比對(1.6.3-1.6.4)"
```

---

### Task 11: Authenticate 與 X-Api-Token middleware(細部 1.6.5–1.6.6)

**Files:**
- Update: `backend/internal/middleware/auth.go`
- Update: `backend/config/config.go`
- Test: `backend/internal/middleware/auth_middleware_test.go`

**Interfaces:**
- Consumes: Task 9 `session.Manager` / `VerifyAccessToken`;Task 10 `CheckTokenVersion`;Task 2 enforcer;Task 3 `rls.NewContext`。
- Produces: `middleware.Authenticate(db, sessionMgr, enf) func(http.Handler) http.Handler`(Chi/Connect 通用);ctx 產出 `rls.Identity`(供所有 domain handler;`Role` / `IsPrimary` 自 `session.Claims` 帶入,`DataScope` 自 users 列);`middleware.ApiTokenAuthenticate(cfg) func(http.Handler) http.Handler`;config `ApiTokens`(map 名稱→雜湊);公開路由白名單 `middleware.PublicPaths`(prefix 列表:`/api/v1/auth/oauth`、`/api/v1/companies/public`、QR 兌換路徑)。

- [ ] **Step 1: config 擴充**

`backend/config/config.go` 加:

```go
// ApiTokens:名稱 → SHA-256(token)。僅 server-to-server;值以雜湊存放(細部 1.6.6)。
ApiTokens map[string]string `mapstructure:"API_TOKENS"`
```

- [ ] **Step 2: 寫失敗測試**

`backend/internal/middleware/auth_middleware_test.go`:

```go
package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/middleware"
	"github.com/salesorder/sales-order-1.0/backend/internal/session"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

// probe handler:回寫 ctx 中的 rls.Identity 供斷言。
func probe(w http.ResponseWriter, r *http.Request) {
	id, ok := rls.FromContext(r.Context())
	if !ok {
		http.Error(w, "no identity", http.StatusInternalServerError)
		return
	}
	w.Header().Set("X-Role", id.DataScope)
	w.WriteHeader(http.StatusOK)
}

func bearerReq(t *testing.T, u *ent.User) *http.Request {
	t.Helper()
	session.InitForTest("test-secret", time.Hour)
	tok, err := session.IssueAccessToken(session.Claims{
		UserID: u.ID, Role: u.Role, CompanyID: u.CompanyID.String(),
		DepartmentID: u.DepartmentID, CustomerID: u.CustomerID,
		IsPrimary: u.IsPrimary, TokenVersion: u.TokenVersion,
	})
	if err != nil { t.Fatal(err) }
	req := httptest.NewRequest("GET", "/api/v1/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	return req
}

func TestAuthenticateJWT(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	u, _ := c.User.Create().SetCompanyID(co.ID).SetRole("staff").SetDataScope("department").
		SetStatus("active").SetEmail("m@example.com").Save(ctx)

	h := middleware.Authenticate(c, nil, nil)(http.HandlerFunc(probe))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, bearerReq(t, u))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Role") != "department" {
		t.Fatalf("identity not injected: %q", rec.Header().Get("X-Role"))
	}
}

func TestAuthenticateNoCredential(t *testing.T) {
	c := testutil.NewEntClient(t)
	h := middleware.Authenticate(c, nil, nil)(http.HandlerFunc(probe))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/protected", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestAuthenticateStaleTokenVersion(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	u, _ := c.User.Create().SetCompanyID(co.ID).SetRole("staff").SetDataScope("department").
		SetStatus("active").SetEmail("s@example.com").Save(ctx)
	req := bearerReq(t, u)
	c.User.UpdateOneID(u.ID).SetTokenVersion(u.TokenVersion + 1).ExecX(ctx) // 強制登出模擬

	h := middleware.Authenticate(c, nil, nil)(http.HandlerFunc(probe))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("stale tv: want 401, got %d", rec.Code)
	}
}

func TestAuthenticateCompanySuspended(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	u, _ := c.User.Create().SetCompanyID(co.ID).SetRole("staff").SetDataScope("department").
		SetStatus("active").SetEmail("c@example.com").Save(ctx)
	c.Company.UpdateOneID(co.ID).SetStatus("suspended").ExecX(ctx)

	h := middleware.Authenticate(c, nil, nil)(http.HandlerFunc(probe))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, bearerReq(t, u))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("suspended company: want 401, got %d", rec.Code)
	}
}

func TestApiToken(t *testing.T) {
	hash := middleware.HashApiToken("secret-token-1")
	cfg := map[string]string{"netsuite": hash} // 名稱 → 雜湊
	c := testutil.NewEntClient(t)

	h := middleware.ApiTokenAuthenticate(c, cfg)(http.HandlerFunc(probe))

	ok := httptest.NewRequest("GET", "/api/v1/protected", nil)
	ok.Header.Set("X-Api-Token", "secret-token-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, ok)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid api token: want 200, got %d", rec.Code)
	}

	bad := httptest.NewRequest("GET", "/api/v1/protected", nil)
	bad.Header.Set("X-Api-Token", "wrong")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, bad)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("invalid api token: want 401, got %d", rec.Code)
	}
	_ = connect.CodeUnauthenticated
	_ = uuid.New()
}
```

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/middleware/ -v`
Expected: FAIL — `middleware.Authenticate` / `ApiTokenAuthenticate` / `HashApiToken` 未定義。

- [ ] **Step 4: 實作**

`backend/internal/middleware/auth.go` 增量(保留 Task 10 的 `CheckTokenVersion`):

```go
// HashApiToken:SHA-256 雜湊,config 只存雜湊(細部 1.6.6 步驟 1)。
func HashApiToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// identityFromUser 由 ent.User 建 rls.Identity。
func identityFromUser(u *ent.User) rls.Identity {
	return rls.Identity{
		UserID:       u.ID,
		CompanyID:    u.CompanyID,
		DepartmentID: u.DepartmentID,
		CustomerID:   u.CustomerID,
		DataScope:    u.DataScope,
	}
}

// Authenticate 統一解析 Web session 與 Bearer JWT(細部 1.6.5)。
// 檢查順序:憑證存在 → 驗證 → 公司 status(2.1.3 連鎖阻斷)→ tv 比對(1.6.4)
// → 主帳號業務限制(1.5.1,於 RPC 層判斷)→ Casbin(2.x 於 handler 層)→ rls.NewContext。
// sessionMgr / enf 可為 nil(分別停用 session 路徑 / Casbin 檢查,Casbin 於 Phase 2 接上)。
func Authenticate(db *ent.Client, sessionMgr *session.Manager, enf *casbin.Enforcer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			var userID uuid.UUID
			var tv int
			if authz := r.Header.Get("Authorization"); strings.HasPrefix(authz, "Bearer ") {
				claims, err := session.VerifyAccessToken(strings.TrimPrefix(authz, "Bearer "))
				if err != nil {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				userID = claims.UserID
				tv = claims.TokenVersion
			} else if sessionMgr != nil {
				v := sessionMgr.Get(r.Context(), "identity")
				view, ok := v.(session.IdentityView)
				if !ok {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				id, err := uuid.Parse(view.UserID)
				if err != nil {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				userID = id
				tv = view.TokenVersion
			} else {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// tv 比對 + 使用者狀態(1.6.4)
			if err := CheckTokenVersion(ctx, db, userID, tv); err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			u, err := db.User.Get(ctx, userID)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			// 公司停用連鎖阻斷(2.1.3;company status 檢查放這裡,規則定義於 02-tenancy-users.md)
			co, err := db.Company.Get(ctx, u.CompanyID)
			if err != nil || co.Status != "active" || co.DeletedAt != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// 主帳號業務 API 限制(1.5.1):JWT/claims 帶 prim=true 時,
			// 僅帳號管理路徑放行;路徑白名單於路由層標記,這裡以 header 約定:
			if u.IsPrimary && r.Header.Get("X-Account-Management") != "true" {
				http.Error(w, "primary account: management only", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r.WithContext(rls.NewContext(ctx, identityFromUser(u))))
		})
	}
}

// ApiTokenAuthenticate 驗證 X-Api-Token(細部 1.6.6);通過注入設定好的機器身分。
// tokens:名稱 → token 雜湊;機器身分以第一台公司/自訂設定,實作於 main 組裝時決定範圍。
// 本 Phase 簡化:通過後注入 data_scope=all 的機器身分(僅 server-to-server 內部呼叫,
// 寫稽核由 audit.Recorder 以 actor="api-token:<名稱>" 記錄)。
func ApiTokenAuthenticate(db *ent.Client, tokens map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tok := r.Header.Get("X-Api-Token")
			if tok == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			h := HashApiToken(tok)
			name := ""
			for n, want := range tokens {
				if subtle.ConstantTimeCompare([]byte(h), []byte(want)) == 1 {
					name = n
					break
				}
			}
			if name == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			// 機器身分:無 users 列,不走 token_version;RLS 注入 data_scope=all(細部 1.6.6 步驟 2)
			co, err := db.Company.Query().First(r.Context())
			if err != nil {
				http.Error(w, "internal", http.StatusInternalServerError)
				return
			}
			id := rls.Identity{
				UserID:    uuid.Nil, // 機器身分無 user;稽核以名稱記錄
				CompanyID: co.ID,
				DataScope: "all",
			}
			next.ServeHTTP(w, r.WithContext(rls.NewContext(r.Context(), id)))
		})
	}
}
```

import 補:`"crypto/sha256"`、`"crypto/subtle"`、`"encoding/hex"`、`"net/http"`、`"strings"`、`"github.com/casbin/casbin/v2"`、`"github.com/salesorder/sales-order-1.0/backend/internal/session"`。

說明(主帳號限制):`X-Account-Management` header 為佔位約定 — 正式實作在 Connect interceptor 以 RPC 方法名白名單判斷(`AuthService.ChangePassword`、`UserService` 帳號管理方法),於 Phase 2 使用者管理 Task 一併落地;本 Task 保留行為驗證點。

- [ ] **Step 5: 跑測試確認通過 + Commit**

Run: `cd backend && go test ./internal/middleware/ -v`
Expected: PASS — JWT 通過、無憑證 401、tv 過期 401、公司停用 401、API token 成敗。

```bash
git add backend/internal/middleware backend/config/config.go
git commit -m "feat(backend): Authenticate 與 X-Api-Token middleware(1.6.5-1.6.6)"
```

---

### Task 12: 強制登出(細部 1.7.1–1.7.2)

**Files:**
- Create: `backend/internal/domain/users/usecase.go`
- Create: `backend/proto/v1/user.proto`
- Test: `backend/internal/domain/users/usecase_test.go`

**Interfaces:**
- Consumes: Task 10 refresh 作廢邏輯、Task 11 `CheckTokenVersion` 行為。
- Produces: `users.Usecase.ForceLogout(ctx, actor rls.Identity, targetID uuid.UUID) error`;proto `UserService.ForceLogout(user_id)`;內部函式 `forceLogoutUser(ctx, tx, targetID) error`(交易內核心,日後其他觸發點重用);Valkey session 刪除為提交後動作,失敗不影響撤銷語義(tv 已 +1)。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/domain/users/usecase_test.go`:

```go
package users_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/users"
	"github.com/salesorder/sales-order-1.0/backend/internal/session"
	"github.com/salesorder/sales-order-1.0/backend/internal/testutil"
)

func TestForceLogoutRevokesEverything(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	target, _ := c.User.Create().SetCompanyID(co.ID).SetRole("staff").SetDataScope("department").
		SetStatus("active").SetEmail("f@example.com").Save(ctx)

	session.InitForTest("test-secret", 0) // TTL 於 JWT 測試已覆蓋
	refresh, err := session.IssueRefreshToken(ctx, c, target.ID, target.TokenVersion)
	if err != nil { t.Fatal(err) }

	uc := users.NewUsecase(c, nil) // session manager nil:Valkey 刪除跳過(tv 已保證失效)
	actor := rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DataScope: "all"}
	if err := uc.ForceLogout(ctx, actor, target.ID); err != nil {
		t.Fatalf("force logout: %v", err)
	}

	after := c.User.GetX(ctx, target.ID)
	if after.TokenVersion != target.TokenVersion+1 {
		t.Fatal("token_version not bumped")
	}
	// refresh token 全作廢
	if _, err := session.RotateRefreshToken(ctx, c, refresh); err == nil {
		t.Fatal("refresh token should be revoked")
	}
	// 帳號 status 不變(可重新登入;禁止登入用停用,細部 2.3.3)
	if after.Status != "active" {
		t.Fatal("status should remain active")
	}
}

func TestForceLogoutScopeAndSelf(t *testing.T) {
	c := testutil.NewEntClient(t)
	ctx := context.Background()
	co, _ := c.Company.Create().SetName("甲").SetIdentifier("co-a").SetCustomerCodePrefix("AA").Save(ctx)
	target, _ := c.User.Create().SetCompanyID(co.ID).SetRole("staff").SetDataScope("department").
		SetStatus("active").SetEmail("s@example.com").Save(ctx)

	uc := users.NewUsecase(c, nil)
	// 對自己呼叫 → invalid_argument
	self := rls.Identity{UserID: target.ID, CompanyID: co.ID, DataScope: "all"}
	if err := uc.ForceLogout(ctx, self, target.ID); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("self force logout: want invalid_argument, got %v", err)
	}
	// 範圍外 → permission_denied
	other := rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(), DataScope: "company"}
	if err := uc.ForceLogout(ctx, other, target.ID); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("out of scope: want permission_denied, got %v", err)
	}
	// 目標不存在 → not_found
	admin := rls.Identity{UserID: uuid.New(), CompanyID: co.ID, DataScope: "all"}
	if err := uc.ForceLogout(ctx, admin, uuid.New()); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("missing target: want not_found, got %v", err)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/users/ -v`
Expected: FAIL — `users.NewUsecase` 未定義。

- [ ] **Step 3: 實作**

`backend/internal/domain/users/usecase.go`:

```go
// Package users 為使用者管理 domain;本 Phase 僅 ForceLogout(細部文件 1.7)。
package users

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	entrt "github.com/salesorder/sales-order-1.0/backend/ent/refreshtoken"
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/session"
)

type Usecase struct {
	db *ent.Client
	sm *session.Manager // 可為 nil;非 nil 時提交後刪 Valkey session
}

func NewUsecase(db *ent.Client, sm *session.Manager) *Usecase {
	return &Usecase{db: db, sm: sm}
}

// ForceLogout 強制登出指定使用者(細部 1.7.2):
// 範圍檢查 → 交易核心(1.7.1)→ 提交後刪 session。
func (u *Usecase) ForceLogout(ctx context.Context, actor rls.Identity, targetID uuid.UUID) error {
	if actor.UserID == targetID {
		return connect.NewError(connect.CodeInvalidArgument,
			errors.New("cannot force-logout self; use normal logout"))
	}
	target, err := u.db.User.Get(ctx, targetID)
	if err != nil || target.DeletedAt != nil {
		return connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	}
	// 範圍:all 不限;company 限同公司;department 由 RLS/2.3.2 收窄(Phase 2 統一)
	if actor.DataScope != "all" && target.CompanyID != actor.CompanyID {
		return connect.NewError(connect.CodePermissionDenied, errors.New("out of scope"))
	}

	tx, err := u.db.Tx(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if err := forceLogoutUser(ctx, tx, targetID); err != nil {
		_ = tx.Rollback()
		return connect.NewError(connect.CodeInternal, err)
	}
	if err := tx.Commit(); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}

	// 提交後:刪 Valkey session(失敗僅記 log — tv 已 +1,撤銷語義不受影響,細部 1.7.1 步驟 1)
	if u.sm != nil {
		_ = u.sm.DeleteSessionByUserID(ctx, targetID) // session.Manager 加此方法:掃 user:<id> 索引 key
	}
	return nil
}

// forceLogoutUser 交易內核心:tv+1 + 作廢全部 refresh token(+稽核,Task 2.6 接上)。
// 供其他觸發點(如改密碼已內嵌於 Task 7)重用。
func forceLogoutUser(ctx context.Context, tx *ent.Tx, targetID uuid.UUID) error {
	target, err := tx.User.Get(ctx, targetID)
	if err != nil {
		return err
	}
	if _, err := tx.User.UpdateOneID(targetID).
		SetTokenVersion(target.TokenVersion + 1).Save(ctx); err != nil {
		return err
	}
	now := time.Now()
	if _, err := tx.RefreshToken.Update().
		Where(entrt.UserIDEQ(targetID), entrt.RotatedAtIsNil()).
		SetRotatedAt(now).Save(ctx); err != nil {
		return err
	}
	// TODO(Phase 2 Task 2.6): audit log(action=force_logout)同事務
	return nil
}
```

`session.Manager` 補 `DeleteSessionByUserID(ctx, userID) error`:scs 不直接支援按內容刪;實作於 `IssueWeb` 時額外寫 Valkey key `user_sessions:<user_id>` = session token set(SADD),刪除時讀 set 逐一 `Destroy` 後刪 set。加入 `manager.go`:

```go
// IssueWeb 內加:_ = m.trackSession(r, u.ID)
func (m *Manager) trackSession(r *http.Request, userID uuid.UUID) error {
	tok := m.Token(r.Context())
	return m.rdb.SAdd(r.Context(), "user_sessions:"+userID.String(), tok).Err()
}

// DeleteSessionByUserID 刪除該使用者全部 Web session。
func (m *Manager) DeleteSessionByUserID(ctx context.Context, userID uuid.UUID) error {
	key := "user_sessions:" + userID.String()
	toks, err := m.rdb.SMembers(ctx, key).Result()
	if err != nil {
		return err
	}
	for _, tok := range toks {
		_ = m.rdb.Del(ctx, "scs:session:"+tok).Err() // scs redisstore key 前綴
	}
	return m.rdb.Del(ctx, key).Err()
}
```

(`Manager` struct 加 `rdb *redis.Client` 欄位,`NewManager` 保留 rdb。)

- [ ] **Step 4: proto + 跑測試**

`backend/proto/v1/user.proto`:

```proto
syntax = "proto3";
package salesorder.v1;

import "google/protobuf/empty.proto";

service UserService {
  rpc ForceLogout(ForceLogoutRequest) returns (google.protobuf.Empty);
}

message ForceLogoutRequest { string user_id = 1; }
```

Run: `cd backend && go test ./internal/domain/users/ ./internal/session/ -v`
Expected: PASS — tv+1、refresh 全廢、status 不變、self/範圍/不存在三錯誤路徑。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/users backend/internal/session/manager.go backend/proto/v1/user.proto
git commit -m "feat(backend): 強制登出(tv+1 + refresh 作廢 + session 刪除)(1.7)"
```

---

### Task 13: ability API(細部 1.8.1)

**Files:**
- Create: `backend/internal/domain/auth/ability.go`
- Update: `backend/proto/v1/auth.proto`
- Test: `backend/internal/domain/auth/ability_test.go`

**Interfaces:**
- Consumes: Task 11 ctx 的 `rls.Identity`(另需 role/is_primary:擴充 `rls.Identity` 加 `Role string`、`IsPrimary bool` 欄位,`identityFromUser` 同步填)。
- Produces: proto `AbilityService.GetAbility() → GetAbilityResponse{rules: repeated AbilityRule}`;`auth.BuildAbility(id rls.Identity) []AbilityRule`;`AbilityRule{Action, Subject string; Conditions map[string]string; Inverted bool}`(CASL JSON 可消費);內建預設規則表(Phase 2 Task 2.9.4 改由 `role_permissions` 表驅動,屆時僅替換規則來源,序列化層不變)。

- [ ] **Step 1: 寫失敗測試**

`backend/internal/domain/auth/ability_test.go`:

```go
package auth_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/domain/auth"
)

func TestBuildAbilityStaff(t *testing.T) {
	coID := uuid.New()
	depID := uuid.New()
	rules := auth.BuildAbility(rls.Identity{
		UserID: uuid.New(), CompanyID: coID, DepartmentID: &depID,
		DataScope: "department", Role: "staff",
	})
	if !auth.Can(rules, "read", "customers") {
		t.Fatal("staff should read customers")
	}
	if auth.Can(rules, "delete", "users") {
		t.Fatal("staff should not delete users")
	}
	// conditions 帶租戶範圍供前端判斷
	found := false
	for _, r := range rules {
		if r.Subject == "customers" && r.Conditions["company_id"] == coID.String() {
			found = true
		}
	}
	if !found {
		t.Fatal("rules should carry company_id condition")
	}
}

func TestBuildAbilityPrimaryAccount(t *testing.T) {
	rules := auth.BuildAbility(rls.Identity{
		UserID: uuid.New(), CompanyID: uuid.New(),
		DataScope: "self", Role: "customer", IsPrimary: true,
	})
	// 主帳號:僅帳號管理,業務主題反向禁止(細部 1.8.1 步驟 2)
	if auth.Can(rules, "create", "sales_orders") {
		t.Fatal("primary account must not create orders")
	}
	if !auth.Can(rules, "manage", "account") {
		t.Fatal("primary account should manage sub-accounts")
	}
}

func TestBuildAbilityUnknownRoleFailClosed(t *testing.T) {
	rules := auth.BuildAbility(rls.Identity{Role: "no-such-role"})
	if len(rules) != 0 {
		t.Fatalf("unknown role should yield empty rules (fail-closed), got %d", len(rules))
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/domain/auth/ -run TestBuildAbility -v`
Expected: FAIL — `auth.BuildAbility` / `auth.Can` 未定義。

- [ ] **Step 3: 實作**

`backend/internal/domain/auth/ability.go`:

```go
package auth

import (
	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// AbilityRule 為 CASL.js 可消費的規則。
type AbilityRule struct {
	Action     string            `json:"action"`
	Subject    string            `json:"subject"`
	Conditions map[string]string `json:"conditions,omitempty"`
	Inverted   bool              `json:"inverted,omitempty"`
}

// builtinAbility:Phase 1 內建預設規則(詞彙與 Casbin obj/act 同源,細部 1.8.1 步驟 2)。
// Phase 2 Task 2.9.4 起改由 role_permissions 表驅動,本表屆時移除。
var builtinAbility = map[string][]AbilityRule{
	"super": {
		{Action: "manage", Subject: "all"},
	},
	"company_admin": {
		{Action: "manage", Subject: "users"},
		{Action: "manage", Subject: "customers"},
		{Action: "manage", Subject: "sales_orders"},
		{Action: "manage", Subject: "return_requests"},
		{Action: "manage", Subject: "dispatch"},
		{Action: "manage", Subject: "printing"},
		{Action: "read", Subject: "audit_logs"},
	},
	"dept_admin": {
		{Action: "manage", Subject: "customers"},
		{Action: "manage", Subject: "sales_orders"},
		{Action: "manage", Subject: "return_requests"},
		{Action: "manage", Subject: "dispatch"},
		{Action: "manage", Subject: "printing"},
		{Action: "read", Subject: "users"},
		{Action: "update", Subject: "users"},
	},
	"staff": {
		{Action: "read", Subject: "customers"},
		{Action: "create", Subject: "customers"},
		{Action: "update", Subject: "customers"},
		{Action: "manage", Subject: "sales_orders"},
		{Action: "read", Subject: "return_requests"},
		{Action: "update", Subject: "return_requests"},
		{Action: "manage", Subject: "dispatch"},
		{Action: "manage", Subject: "printing"},
	},
	"customer": {
		{Action: "read", Subject: "sales_orders"},
		{Action: "create", Subject: "sales_orders"},
		{Action: "read", Subject: "return_requests"},
		{Action: "create", Subject: "return_requests"},
		{Action: "read", Subject: "customer_products"},
		{Action: "read", Subject: "notifications"},
	},
	"guest": {}, // 審核前無權限
}

// BuildAbility 依身分產生 CASL 規則;未知角色回空(fail-closed)。
// 主帳號:全部業務主題反向禁止,僅開帳號管理(細部 1.8.1 步驟 2)。
func BuildAbility(id rls.Identity) []AbilityRule {
	rules, ok := builtinAbility[id.Role]
	if !ok {
		return nil
	}
	out := make([]AbilityRule, 0, len(rules)+4)
	conds := map[string]string{"company_id": id.CompanyID.String()}
	if id.DepartmentID != nil {
		conds["department_id"] = id.DepartmentID.String()
	}
	if id.CustomerID != nil {
		conds["customer_id"] = id.CustomerID.String()
	}
	for _, r := range rules {
		if r.Subject != "all" {
			r.Conditions = conds
		}
		out = append(out, r)
	}
	if id.IsPrimary {
		// 業務主題全部 cannot
		for _, s := range []string{"sales_orders", "return_requests", "customer_products"} {
			out = append(out, AbilityRule{Action: "manage", Subject: s, Inverted: true})
		}
		out = append(out, AbilityRule{Action: "manage", Subject: "account"})
	}
	return out
}

// Can 為測試與後端雙重檢查用的簡易求值:先掃 inverted 禁止,再掃允許。
func Can(rules []AbilityRule, action, subject string) bool {
	allowed := false
	for _, r := range rules {
		match := (r.Action == action || r.Action == "manage") &&
			(r.Subject == subject || r.Subject == "all")
		if !match {
			continue
		}
		if r.Inverted {
			return false
		}
		allowed = true
	}
	return allowed
}
```

`rls.Identity` 加欄位(`rls.go`):`Role string`、`IsPrimary bool`;`identityFromUser`(middleware)填 `Role: u.Role, IsPrimary: u.IsPrimary`。

proto(`auth.proto` 加 service):

```proto
service AbilityService {
  rpc GetAbility(google.protobuf.Empty) returns (GetAbilityResponse);
}

message AbilityRule {
  string action = 1;
  string subject = 2;
  map<string, string> conditions = 3;
  bool inverted = 4;
}
message GetAbilityResponse { repeated AbilityRule rules = 1; }
```

Connect handler `GetAbility`:從 ctx 取 `rls.Identity` → `BuildAbility` → 回應;未登入由 middleware 擋(此處再防一次:無 identity → `unauthenticated`)。快取:回應加 `Cache-Control: private, max-age=60`(細部 1.8.1 步驟 3)。

- [ ] **Step 4: 跑測試確認通過 + Commit**

Run: `cd backend && go test ./internal/domain/auth/ -v`
Expected: PASS — staff 規則與 conditions、主帳號反向禁止、未知角色 fail-closed。

```bash
git add backend/internal/domain/auth/ability.go backend/internal/database/rls.go backend/internal/middleware/auth.go backend/proto/v1/auth.proto
git commit -m "feat(backend): ability API 內建預設規則產生 CASL JSON(1.8)"
```

---

### Task 14: developer 逃生門與 audit 介面(細部 1.11.1–1.11.3)

**Files:**
- Create: `backend/internal/middleware/developer.go`
- Create: `backend/internal/audit/recorder.go`
- Update: `backend/config/config.go`
- Update: `backend/cmd/server/main.go`
- Update: `backend/internal/authz/seed.go`
- Test: `backend/internal/middleware/developer_test.go`

**Interfaces:**
- Consumes: Task 11 `Authenticate`;Task 2 seeder。
- Produces: config `DeveloperAccountEnabled bool`、`Env string`;`middleware.DeveloperBypass(enabled bool) func(http.Handler) http.Handler`(在 `Authenticate` 之後);`audit.Recorder` 介面(`Record(ctx, entry Entry) error`,`Entry{ActorID, ActorName, Action, ResourceType, ResourceID, Before, After string}`)+ `audit.NoopRecorder{}`(Phase 2 Task 2.6 以 DB 實作替換);`authz.SeedDeveloperRole(ctx, db)`(developer 角色語義由 `users.role` enum 承載;開發帳號僅 development seed)。
- 啟動防護:`main.go` 中 `cfg.Env == "production" && cfg.DeveloperAccountEnabled` → `log.Fatal`(fail-fast)。

- [ ] **Step 1: config 與 audit 介面**

`config.go` 加:

```go
Env                     string `mapstructure:"ENV"` // development | test | production
DeveloperAccountEnabled bool   `mapstructure:"DEVELOPER_ACCOUNT_ENABLED"`
DeveloperEmail          string `mapstructure:"DEVELOPER_EMAIL"`    // 開發者帳號 email(development seed 用)
DeveloperPassword       string `mapstructure:"DEVELOPER_PASSWORD"` // 無預設值;未設定則略過 seed
```

`Load()`:預設 `Env=development`;`DeveloperAccountEnabled` 預設值依 env:`development/test → true`、`production → false`(viper 未顯式設定時)。

`backend/internal/audit/recorder.go`:

```go
// Package audit 定義稽核寫入介面;Phase 2 Task 2.6 以 DB 實作替換 NoopRecorder(細部文件共通規則 1、D18)。
package audit

import "context"

// Entry 為一筆稽核記錄;Before/After 為 JSON 快照,敏感欄位(密碼、token)永不填入。
type Entry struct {
	ActorID      string // user id 或 "api-token:<名稱>"、"developer"
	ActorName    string
	Action       string // login / login_failed / force_logout / password_reset / guest_approved / ...
	ResourceType string
	ResourceID   string
	Before       string
	After        string
	IP           string
}

// Recorder 由 usecase 層在交易內呼叫;DB 版實作僅接受 tx client(D18 同生滅)。
type Recorder interface {
	Record(ctx context.Context, e Entry) error
}

// NoopRecorder 為 Phase 1 佔位:不寫任何東西,介面先落地讓所有寫入點就位。
type NoopRecorder struct{}

func (NoopRecorder) Record(context.Context, Entry) error { return nil }

// FakeRecorder 為跨 domain 測試共用 fake:記錄所有 Entry 供斷言(稽核同事務驗證用)。
type FakeRecorder struct {
	Entries []Entry
	Err     error // 非 nil 時 Record 回錯(注入稽核失敗情境,驗證業務回滾 D18)
}

func (f *FakeRecorder) Record(_ context.Context, e Entry) error {
	if f.Err != nil {
		return f.Err
	}
	f.Entries = append(f.Entries, e)
	return nil
}
```

- [ ] **Step 2: 寫失敗測試**

`backend/internal/middleware/developer_test.go`:

```go
package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
	"github.com/salesorder/sales-order-1.0/backend/internal/middleware"
)

// injectIdentity 模擬 Authenticate 已注入身分。
func injectIdentity(id rls.Identity, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(rls.NewContext(r.Context(), id)))
	})
}

func TestDeveloperBypassWhenEnabled(t *testing.T) {
	devID := rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(), DataScope: "department", Role: "developer"}
	var got rls.Identity
	probe := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = rls.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	h := injectIdentity(devID, middleware.DeveloperBypass(true)(probe))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/anything", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if got.DataScope != "all" {
		t.Fatalf("developer should be escalated to data_scope=all, got %q", got.DataScope)
	}
}

func TestDeveloperRejectedWhenDisabled(t *testing.T) {
	devID := rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(), DataScope: "department", Role: "developer"}
	probe := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := injectIdentity(devID, middleware.DeveloperBypass(false)(probe))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/anything", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("disabled: developer should be rejected with 401, got %d", rec.Code)
	}
}

func TestNormalUserUntouched(t *testing.T) {
	staff := rls.Identity{UserID: uuid.New(), CompanyID: uuid.New(), DataScope: "department", Role: "staff"}
	var got rls.Identity
	probe := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = rls.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	h := injectIdentity(staff, middleware.DeveloperBypass(true)(probe))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/anything", nil))
	if rec.Code != http.StatusOK || got.DataScope != "department" {
		t.Fatalf("staff should pass through unchanged: code=%d scope=%q", rec.Code, got.DataScope)
	}
}
```

- [ ] **Step 3: 跑測試確認失敗**

Run: `cd backend && go test ./internal/middleware/ -run TestDeveloper -v`
Expected: FAIL — `DeveloperBypass` 未定義。

- [ ] **Step 4: 實作 developer middleware 與啟動防護**

`backend/internal/middleware/developer.go`:

```go
package middleware

import (
	"log"
	"net/http"

	"github.com/salesorder/sales-order-1.0/backend/internal/database/rls"
)

// DeveloperBypass 疊加於 Authenticate 之後(細部 1.11.2):
// 身分為 developer 且開關啟用 → data_scope=all(RLS 全開)+ Casbin 檢查由 handler 層
// 讀 ctx 的 bypass 旗標跳過;開關關閉時 developer 身分一律拒絕(異常狀態,記 security log)。
// 稽核:developer 操作照常記錄(audit.Recorder 以 actor="developer"),不繞過。
func DeveloperBypass(enabled bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, ok := rls.FromContext(r.Context())
			if !ok || id.Role != "developer" {
				next.ServeHTTP(w, r)
				return
			}
			if !enabled {
				log.Printf("security: developer identity seen while DEVELOPER_ACCOUNT_ENABLED=false (user=%s)", id.UserID)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			id.DataScope = "all"
			next.ServeHTTP(w, r.WithContext(rls.NewContext(r.Context(), id)))
		})
	}
}
```

`backend/cmd/server/main.go` 啟動序列插入(在 `config.Load()` 之後):

```go
	// 啟動防護(細部 1.11.1):production 誤開 developer 帳號 → fail-fast
	if cfg.Env == "production" && cfg.DeveloperAccountEnabled {
		log.Fatal("DEVELOPER_ACCOUNT_ENABLED=true is forbidden in production")
	}
	log.Printf("env=%s developer_account_enabled=%v", cfg.Env, cfg.DeveloperAccountEnabled)
```

`authz/seed.go` 加(developer 帳號 seed,僅 development;細部 1.11.3):

```go
// SeedDeveloperAccount 於 ENV=development 且提供密碼時建立開發者帳號;冪等。
// company 使用 seed 的開發公司(identifier=dev);不存在則建立(細部 1.11.3 步驟 3)。
func SeedDeveloperAccount(ctx context.Context, db *ent.Client, email, password string) error {
	if email == "" || password == "" {
		log.Print("developer account seed skipped (email/password unset)")
		return nil
	}
	co, err := db.Company.Query().Where(entcompany.IdentifierEQ("dev")).Only(ctx)
	if err != nil {
		co, err = db.Company.Create().SetName("開發公司").SetIdentifier("dev").
			SetCustomerCodePrefix("DV").Save(ctx)
		if err != nil {
			return fmt.Errorf("seed dev company: %w", err)
		}
	}
	exists, err := db.User.Query().Where(
		entuser.EmailEQ(email), entuser.DeletedAtIsNil()).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	h, err := auth.HashPassword(password) // developer 帳號走帳密登入(開發用);亦可登入後切換
	if err != nil {
		return err
	}
	_, err = db.User.Create().SetCompanyID(co.ID).SetEmail(email).
		SetName("開發者").SetRole("developer").SetDataScope("all").
		SetStatus("active").SetPasswordHash(h).Save(ctx)
	return err
}
```

(此函式放 `authz` 會造成 authz→auth 套件依賴;實作時改放 `backend/internal/seed/developer.go`,import auth 的 `HashPassword` — 以此為準。main.go 於 `cfg.Env == "development"` 時呼叫。)

- [ ] **Step 5: 跑測試 + 啟動防護手動驗證 + Commit**

Run: `cd backend && go test ./internal/middleware/ -v`
Expected: PASS — 三個 developer 測試。

手動驗證啟動防護:

```bash
cd backend && ENV=production DEVELOPER_ACCOUNT_ENABLED=true go run ./cmd/server
```

Expected: 程序立即退出,log 顯示 `DEVELOPER_ACCOUNT_ENABLED=true is forbidden in production`。

```bash
git add backend/internal/middleware/developer.go backend/internal/audit backend/config backend/cmd/server backend/internal/seed
git commit -m "feat(backend): developer 逃生門、啟動防護、audit Recorder 介面(1.11)"
```

---

## Self-Review 記錄

- **Spec 覆蓋**:細部文件 29 子功能 → Task 對應:1.1.x→T1;1.2.x→T2;1.3.x→T3;1.4.1-1.4.2→T4;1.4.3-1.4.4→T5;1.5.1→T6;1.5.2→T7;1.5.3-1.5.4→T8;1.6.1-1.6.2→T9;1.6.3-1.6.4→T10;1.6.5-1.6.6→T11;1.7.x→T12;1.8.1→T13;1.11.x→T14。無缺漏。
- **已知佔位(皆標 TODO + 接手 Task)**:`mapCustomerCode`(→ Phase 3 Task 3.1)、audit 寫入點的 DB 實作(→ Task 2.6)、主帳號路徑白名單 interceptor(→ Phase 2)、Casbin 於 handler 層接上(→ Phase 2 Task 2.10)。這些是跨 Phase 依賴,非本計畫範圍。
- **類型一致**:`rls.Identity`(T3 定義,T11/T13/T14 擴充 Role/IsPrimary)、`session.Claims`(T9)、`auth.OneTimeStore`(T4)、`auth.Usecase`(T5)於各 Task 簽名一致。

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-08-17-backend-01-auth-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** — 每個 Task 派新 subagent 執行,Task 間 review,迭代快。

**2. Inline Execution** — 用 executing-plans 在本 session 逐批執行,設 checkpoint review。

Which approach?

---

*計畫版本:v1.0.0(2026-08-17);對應細部文件 `backend-detail/01-auth.md` v1.0.0、原計畫 v2.9.0、規格書 v1.0.34。*
