# RLS 租戶隔離啟用 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 讓 PostgreSQL RLS 從「只定義 policy」變成真正生效：非 owner 應用角色、14 個既有 policy 補 `WITH CHECK`、三張漏網表補 policy、業務表 `ENABLE` + `FORCE`，並把服務層全部 DB 存取收斂到「請求層租戶交易」，使跨租戶資料在 DB 層不可見也不可寫。

**Architecture:** connect `Interceptor` 在每個 unary RPC 開一個交易並以 `auth.ApplyRLS` 套用 ctx 內的身分/範圍（`SET LOCAL app.*`），handler 回傳錯誤即 rollback、成功即 commit；服務層以 `dbtenant.Client(ctx, s.db)` 取得被 scope 約束的 client，交易邊界由請求持有。業務連線用非 owner 角色 `app_rw`（RLS 對其生效）；goose/seed/OpenFGA 用 owner 的 `DATABASE_ADMIN_URL`。

**Tech Stack:** Go 1.25、ent、connect-go、goose、PostgreSQL 16、testcontainers-go（`internal/testsupport`）

**Spec:** `docs/superpowers/specs/2026-09-20-saas-billing-entitlements-design.md`（§2.3、§6 為權威；本計畫為其第一步的實作）

## Global Constraints

- 業務連線身分為 `app_rw`：`LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOBYPASSRLS` 且**非 table owner**；migration、seed、OpenFGA、平台域一律走 `DATABASE_ADMIN_URL`（owner）
- policy 一律 `FOR ALL … USING (…) WITH CHECK (…)`；`WITH CHECK` 必須**等於或嚴於**原 `USING`，`core_metadicts_scope` 沿用原樣不得放寬
- 服務層 DB 存取一律經 `dbtenant.Client(ctx, s.db)`；`db.Tx(ctx)` 不得再自行開交易（42 處全數改為使用請求交易）
- RLS 相關測試一律 `//go:build integration` + `internal/testsupport`；**不得以 sqlite（enttest）測 RLS**（enttest 不支援 `SET`／`FORCE`）
- **RLS 的驗證必須以非 superuser 連線**（T5 實測更正）：測試容器的 `postgres` 是 superuser，而 PostgreSQL superuser **恆繞過 RLS（`FORCE` 亦然）** → 以它連線的測試全綠**不能**當作「RLS 生效」的證據（只能當 regression gate）。凡宣稱驗 RLS 的測試，必須以 `app_rw`（或專用非 superuser 角色）建立連線／client。
- 建 `companies` fixture 時**不得**寫 `created_at`／`updated_at`（該表無此欄位，見 `00005`；T5 實測踩過）。
- 每個 migration 必含 `Up`/`Down`；ENABLE 的 `Down` 必含 `NO FORCE` + `DISABLE`
- **對已 ENABLE（且 FORCE）的表做資料回填**的 migration 與 `cmd/seed`：交易內先執行 `SET LOCAL app.current_data_scope = 'all'`（FORCE 也會擋 owner）
- 錯誤一律經既有 `toConnectError` 映射，不得回傳 SQLSTATE 或 constraint 名
- 註解、commit message 一律繁體中文（repo 慣例）
- 每個任務結束前跑 `task check`；含整合測試的任務另跑 `task test:integration` 的對應 `-run`

### 交易改寫樣式（每個 domain 遷移任務都適用）

服務層既有寫法：

```go
tx, err := s.db.Tx(ctx)
if err != nil {
	return nil, toConnectError(err)
}
// … 使用 tx.…
if err := tx.Commit(); err != nil {
	return nil, toConnectError(err)
}
```

改為（交易由請求層 interceptor 擁有，服務內不再開交易、不再 commit/rollback）：

```go
// 由 ctx 取請求交易：稽核寫入需要 *ent.Tx（audit.Record 的簽章），
// 且 D18「業務寫入與稽核同一交易」正是靠它維持。
tx, ok := dbtenant.TxFrom(ctx)
if !ok {
	// 無請求交易：CLI／seed／未掛載的路徑。回明確錯誤，不要默默退回 fallback client 寫入。
	return nil, connect.NewError(connect.CodeInternal, errors.New("缺少租戶交易(context)"))
}
db := tx.Client() // 查詢與寫入都用它；同一個交易
// … 原本用 tx.X() 的地方全部改用 db.…；recordAuditBA(ctx, tx, …) 維持傳 tx
```

**兩個必須避免的錯誤**（T4 review 交接）：

1. **不要**把 `db.Tx(ctx)` 機械式改成 `dbtenant.Client(ctx, s.db).Tx(ctx)` —— `Client()` 回傳的是「綁在請求交易上的 client」，再對它開交易會得到 ent 的 `ErrTxStarted`，所有寫入路徑會壞。
2. **不要**讓稽核寫入落到另一個交易：`recordAuditBA(ctx, tx, …)` 必須續用請求交易的 `tx`，否則 D18 的「同交易」會被靜默拆開。

同時把同檔所有 `s.db.` 改為 `dbtenant.Client(ctx, s.db).`：

```bash
sd 's\.db\.' 'dbtenant.Client(ctx, s.db).' internal/services/<file>.go
grep -n 's\.db\.\|db.Tx(ctx)' internal/services/<file>.go   # 期望無輸出
```

---

## File Structure

| 路徑 | 職責 |
|---|---|
| `config/database.go`（改） | 新增 `AdminURL` 與 `AdminDSN()`（空則沿用 `DatabaseURL`，dev/CI 零設定） |
| `config/database_test.go`（新） | DSN 分流單元測試 |
| `database/migrations/00022_app_role.sql`（新） | 建立 `app_rw` 角色與授權（業務 schema 全表 DML，platform schema 一格不給） |
| `database/migrations/00023_rls_policy_with_check.sql`（新） | 14 個 policy 補 `WITH CHECK`、三張漏網表新增 policy |
| `database/migrations/00024..00027_rls_enable_*.sql`（新） | 依 domain 分批 `ENABLE` + `FORCE`（客戶／主檔／商品／字典稽核／核心） |
| `internal/dbtenant/dbtenant.go`（新） | 請求層租戶交易：`Interceptor`、`Client`、`WithTenantTx`、`SystemScopeTx` |
| `internal/dbtenant/dbtenant_test.go`（新） | 純函式／sqlite 可測的部分 |
| `internal/dbtenant/tenant_integration_test.go`（新） | 真 PG：GUC 貫穿、commit/rollback 語意 |
| `internal/testsupport/approle.go`（新） | 由容器 admin DSN 導出 `app_rw` DSN（設密碼），供整合測試 |
| `internal/services/*.go`（改） | `s.db.` → `dbtenant.Client(ctx, s.db).`；`db.Tx(ctx)` 改為使用請求交易 |
| `internal/services/company_service.go`（改） | `SetStatus` usecase 抽出（供平台域 consumer 使用；本次先提供並測試） |
| `internal/server/domains.go`（改） | 每個 `NewXServiceHandler` 掛上租戶交易 interceptor |
| `internal/server/server.go`（改） | authzMiddleware 的身分查詢改走 `SystemScopeTx` |
| `internal/handlers/auth_handler.go`（改） | 登入憑證查詢改走 `SystemScopeTx` |
| `cmd/seed/main.go`（改） | 改 admin DSN，並在交易內 `SET LOCAL … 'all'` |
| `internal/services/rls_*_integration_test.go`（新） | 各 domain 跨租戶探針與 policy 結構斷言 |
| `backend/AGENTS.md`（改） | RLS 慣例（請求層交易、回填需系統 scope、RLS 不可用 sqlite 測） |

---

### Task 1: 兩把 DSN（業務與 owner 分流）

**Files:**
- Modify: `config/database.go`
- Create: `config/database_test.go`
- Modify: `cmd/migrate/main.go:23`、`cmd/migrate/openfga.go:30-32`、`cmd/seed/main.go:14,21`、`internal/server/domains.go:111,123-125`
- Modify: `.env.example`（`Database` 段）

**Interfaces:**
- Produces: `config.Database.AdminURL`（env `DATABASE_ADMIN_URL`）、`config.Database.AdminDSN() string`

- [ ] **Step 1: 寫失敗測試（`config/database_test.go`）**

```go
package config

import "testing"

func TestAdminDSNFallsBackToDatabaseURL(t *testing.T) {
	d := Database{DatabaseURL: "postgres://app@localhost:5432/salesorder"}
	if got := d.AdminDSN(); got != "postgres://app@localhost:5432/salesorder" {
		t.Fatalf("未設 DATABASE_ADMIN_URL 時應沿用 DATABASE_URL,got %q", got)
	}
	d.AdminURL = "postgres://owner@localhost:5432/salesorder"
	if got := d.AdminDSN(); got != "postgres://owner@localhost:5432/salesorder" {
		t.Fatalf("設了 DATABASE_ADMIN_URL 時應採用它,got %q", got)
	}
}

func TestAdminURLFromEnv(t *testing.T) {
	t.Setenv("DATABASE_ADMIN_URL", "postgres://owner@db:5432/salesorder")
	var d Database
	mustProcess(&d)
	if d.AdminURL != "postgres://owner@db:5432/salesorder" {
		t.Fatalf("DATABASE_ADMIN_URL 未綁定,got %q", d.AdminURL)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./config/ -run 'TestAdmin' -v`
Expected: FAIL —`d.AdminDSN undefined`、`d.AdminURL undefined`

- [ ] **Step 3: 實作（`config/database.go`）**

```go
package config

// Database PostgreSQL 連線設定。
type Database struct {
	DatabaseURL string `envconfig:"DATABASE_URL" default:"postgres://postgres:postgres@localhost:5432/salesorder?sslmode=disable"`
	// AdminURL 為 owner 連線(goose 遷移、seed、OpenFGA、平台域)；業務連線改走非 owner 的 app_rw，
	// 使 RLS 對業務生效。空則沿用 DatabaseURL(單一角色的開發/測試環境)。
	AdminURL string `envconfig:"DATABASE_ADMIN_URL"`
}

// AdminDSN 回傳 owner 連線字串；未設定 DATABASE_ADMIN_URL 時沿用業務 DSN。
func (d Database) AdminDSN() string {
	if d.AdminURL != "" {
		return d.AdminURL
	}
	return d.DatabaseURL
}
```

- [ ] **Step 4: 把 owner 用途的 5 處改成 `cfg.Database.AdminDSN()`**
  - `cmd/migrate/main.go:23`：`sql.Open("pgx", cfg.Database.AdminDSN())`
  - `cmd/migrate/openfga.go:30-32`：`dsn := cfg.OpenFGA.DatabaseURL; if dsn == "" { dsn = cfg.Database.AdminDSN() }`
  - `cmd/seed/main.go:14,21`：兩處改用 `cfg.Database.AdminDSN()`
  - `internal/server/domains.go:123-125`（`mountOpenFGA`）：`dsn = s.cfg.Database.AdminDSN()`
  - **不動**：`domains.go:111`（`openEntClient` 業務 client）與 `server.go:121`（startup probe 要驗的是業務連線）

- [ ] **Step 5: `.env.example` 補說明**

```dotenv
# --- config/database.go (Database) ---
# 業務連線：非 owner 的 app_rw（RLS 對其生效）
DATABASE_URL=postgres://app_rw:app_rw@localhost:5432/salesorder?sslmode=disable
# owner 連線：goose 遷移/seed/OpenFGA/平台域；未設時沿用 DATABASE_URL
DATABASE_ADMIN_URL=postgres://postgres:postgres@localhost:5432/salesorder?sslmode=disable
```

- [ ] **Step 6: 跑測試確認通過並收尾**

Run: `cd backend && go test ./config/ -run 'TestAdmin' -v && go build ./...`
Expected: PASS，build 無錯

- [ ] **Step 7: Commit**

```bash
git add backend/config/database.go backend/config/database_test.go backend/cmd backend/internal/server/domains.go backend/.env.example
git commit -m "feat(backend): 業務與 owner DSN 分流（DATABASE_ADMIN_URL），為 RLS 啟用鋪路"
```

---

### Task 2: `app_rw` 角色與授權（migration 00022）

**Files:**
- Create: `database/migrations/00022_app_role.sql`
- Create: `internal/testsupport/approle.go`
- Create: `internal/services/rls_app_role_integration_test.go`
- Modify: `Taskfile.yml`（新增 `db:app-password`，本機開發用）

**Interfaces:**
- Consumes: Task 1 的 `AdminDSN()`
- Produces: PG 角色 `app_rw`；`testsupport.AppRoleDSN(t *testing.T, adminDSN string) string`

- [ ] **Step 1: 寫失敗測試（`internal/services/rls_app_role_integration_test.go`）**

```go
//go:build integration

package services

import (
	"database/sql"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationAppRolePrivileges 驗證業務角色是非 owner、且對業務表有 DML 權限：
// RLS 只在「非 owner」連線上才擋得住（owner 需 FORCE，而 FORCE 連帶要求回填走系統 scope）。
func TestIntegrationAppRolePrivileges(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)
	appDSN := testsupport.AppRoleDSN(t, adminDSN)

	db, err := sql.Open("pgx", appDSN)
	if err != nil {
		t.Fatalf("app_rw 連線: %v", err)
	}
	defer db.Close()

	var user string
	if err := db.QueryRow(`SELECT current_user`).Scan(&user); err != nil {
		t.Fatalf("查 current_user: %v", err)
	}
	if user != "app_rw" {
		t.Fatalf("業務連線應為 app_rw,got %q", user)
	}

	var isSuper, bypass bool
	if err := db.QueryRow(
		`SELECT rolsuper, rolbypassrls FROM pg_roles WHERE rolname = current_user`,
	).Scan(&isSuper, &bypass); err != nil {
		t.Fatalf("查角色屬性: %v", err)
	}
	if isSuper || bypass {
		t.Fatalf("app_rw 不得為 superuser 或 BYPASSRLS(rolsuper=%v rolbypassrls=%v)", isSuper, bypass)
	}

	var owners int
	if err := db.QueryRow(
		`SELECT count(*) FROM pg_class c JOIN pg_roles r ON r.oid = c.relowner
		  WHERE r.rolname = current_user AND c.relname IN ('customers','products','users')`,
	).Scan(&owners); err != nil {
		t.Fatalf("查 table owner: %v", err)
	}
	if owners != 0 {
		t.Fatalf("app_rw 不得是業務表 owner（owner 會繞過 RLS），got %d 張", owners)
	}

	var granted int
	if err := db.QueryRow(
		`SELECT count(*) FROM information_schema.role_table_grants
		  WHERE grantee = 'app_rw' AND table_name = 'customers'
		    AND privilege_type IN ('SELECT','INSERT','UPDATE','DELETE')`,
	).Scan(&granted); err != nil {
		t.Fatalf("查授權: %v", err)
	}
	if granted != 4 {
		t.Fatalf("app_rw 對 customers 應有 SELECT/INSERT/UPDATE/DELETE 四項授權,got %d", granted)
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && task test:integration -- -run TestIntegrationAppRolePrivileges -v`
Expected: FAIL —`role "app_rw" does not exist`

- [ ] **Step 3: 寫 migration（`database/migrations/00022_app_role.sql`）**

```sql
-- 業務連線角色(D36):非 owner、NOBYPASSRLS，使 RLS 真正生效。
-- 密碼不由 migration 設定(不落版控):本機開發用 `task db:app-password`，
-- 其他環境由部署流程以 ALTER ROLE app_rw PASSWORD ... 提供。
-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_rw') THEN
        CREATE ROLE app_rw LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOBYPASSRLS;
    END IF;
END
$$;
-- +goose StatementEnd

-- 授權：**明確列舉業務表**，不用 `ALL TABLES`／`ALTER DEFAULT PRIVILEGES`（T2 審查裁決，2026-09-20）。
-- 理由：內嵌 OpenFGA 與業務**共用同一個 database 的 public schema**，其 datastore 表
-- （tuple／authorization_model／store／assertion／changelog）與 goose 版本表都會被
-- `ALL TABLES` 與 default privileges 一併授權給業務角色——等於讓業務連線可改寫授權資料。
-- 且 `ALTER DEFAULT PRIVILEGES` 無法排除「未來由 owner 建立的非業務表」（OpenFGA 的
-- migration 會在 00022 之後繼續建表），屬不可局部修補的設計，故一併移除。
-- 代價（刻意的摩擦）：**新增業務表時，必須在其 migration 內顯式 GRANT 這四項 DML 與 sequence**。
GRANT USAGE ON SCHEMA public TO app_rw;
GRANT SELECT, INSERT, UPDATE, DELETE ON
    companies, departments, users, roles, role_permissions,
    audit_logs, metadicts,
    customers, customer_counters, customer_addresses, customer_contacts,
    warehouses, routes, processing_specs, product_categories,
    products, product_units, product_processing_specs
TO app_rw;
-- sequence 不承載租戶資料，且漏授權會讓 INSERT 失敗（與表的風險不對稱），故保留 ALL SEQUENCES。
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_rw;

-- +goose Down
-- +goose StatementBegin
-- 先撤銷同一份白名單再 DROP ROLE —— 直接 DROP ROLE 會被 ACL 依賴擋下（SQLSTATE 2BP01），
-- 使 Down 不可回滾（原計畫版本即犯此錯，由實作與審查共同發現）。
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public FROM app_rw;
REVOKE SELECT, INSERT, UPDATE, DELETE ON
    companies, departments, users, roles, role_permissions,
    audit_logs, metadicts,
    customers, customer_counters, customer_addresses, customer_contacts,
    warehouses, routes, processing_specs, product_categories,
    products, product_units, product_processing_specs
FROM app_rw;
REVOKE USAGE ON SCHEMA public FROM app_rw;
DROP ROLE IF EXISTS app_rw;
-- +goose StatementEnd
```

- [ ] **Step 4: 寫測試輔助（`internal/testsupport/approle.go`）**

```go
//go:build integration

package testsupport

import (
	"database/sql"
	"net/url"
	"testing"
)

// AppRoleDSN 由 owner DSN 導出 app_rw 連線字串：先把密碼設成固定值（容器為拋棄式，
// 且本檔是唯一設定點），再替換 userinfo。非容器模式（INTEGRATION_TEST_DSN）同樣可行，
// 但會改動該 DSN 指向的庫之角色密碼 → 需專用拋棄式資料庫。
func AppRoleDSN(t *testing.T, adminDSN string) string {
	t.Helper()
	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		t.Fatalf("開 admin 連線: %v", err)
	}
	defer admin.Close()
	if _, err := admin.Exec(`ALTER ROLE app_rw WITH PASSWORD 'app_rw'`); err != nil {
		t.Fatalf("設定 app_rw 密碼: %v（migration 00022 是否已套用？）", err)
	}

	u, err := url.Parse(adminDSN)
	if err != nil {
		t.Fatalf("解析 DSN: %v", err)
	}
	u.User = url.UserPassword("app_rw", "app_rw")
	return u.String()
}
```

- [ ] **Step 5: Taskfile 補本機開發步驟（`task db:app-password`）**

```yaml
  db:app-password:
    desc: 為本機開發資料庫設定 app_rw 密碼（migration 00022 只建角色不建密碼）
    cmds:
      - psql "$DATABASE_ADMIN_URL" -c "ALTER ROLE app_rw WITH PASSWORD 'app_rw'"
```

- [ ] **Step 6: 跑測試確認通過**

Run: `cd backend && task test:integration -- -run TestIntegrationAppRolePrivileges -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add backend/database/migrations/00022_app_role.sql backend/internal/testsupport/approle.go backend/internal/services/rls_app_role_integration_test.go backend/Taskfile.yml
git commit -m "feat(backend): 新增非 owner 業務角色 app_rw 與授權（00022）"
```

---

### Task 3: 14 個 policy 補 `WITH CHECK` ＋ 三張漏網表補 policy（migration 00023）

**Files:**
- Create: `database/migrations/00023_rls_policy_with_check.sql`
- Create: `internal/services/rls_policy_integration_test.go`

**Interfaces:**
- Produces: 18 張租戶表各有 1 個 `FOR ALL … USING … WITH CHECK …` policy

**背景（實測，勿重新推論）：** 15 個既有 policy 中 14 個缺 `WITH CHECK`（`USING` 同時被當成 INSERT/UPDATE 的檢查，但 `USING` 表達的是「可見列」，與「可寫入的新列」不是同一件事——本任務把後者寫成明示條款）；唯一已有 `WITH CHECK` 的是 `core_metadicts_scope`，且它**刻意不含 `department_id IS NULL`**（系統預設字典只有 `scope=all` 能寫），**沿用原樣**。

- [ ] **Step 1: 寫失敗測試（`internal/services/rls_policy_integration_test.go`）**

```go
//go:build integration

package services

import (
	"database/sql"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// rlsPolicyTables 為本波要求「有 FOR ALL policy 且 WITH CHECK 非空」的租戶表。
var rlsPolicyTables = []string{
	"companies", "departments", "users", "roles", "role_permissions",
	"audit_logs", "metadicts", "customers", "customer_addresses", "customer_contacts",
	"customer_counters", "warehouses", "routes", "processing_specs",
	"product_categories", "products", "product_units", "product_processing_specs",
}

// TestIntegrationRLSPolicyCoverage 檢查 policy 覆蓋率與 WITH CHECK：
// 少了 WITH CHECK，RLS 只能擋讀、擋不住「把列寫成別的 company_id」。
func TestIntegrationRLSPolicyCoverage(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateBusinessUp(t, dsn)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("連線: %v", err)
	}
	defer db.Close()

	for _, table := range rlsPolicyTables {
		t.Run(table, func(t *testing.T) {
			var total, withCheck int
			if err := db.QueryRow(
				`SELECT count(*), count(with_check) FROM pg_policies
				  WHERE schemaname = 'public' AND tablename = $1 AND cmd = 'ALL'`,
				table,
			).Scan(&total, &withCheck); err != nil {
				t.Fatalf("查 pg_policies(%s): %v", table, err)
			}
			if total == 0 {
				t.Fatalf("%s 沒有 FOR ALL policy", table)
			}
			if withCheck != total {
				t.Fatalf("%s 有 %d/%d 個 policy 缺 WITH CHECK（可跨租戶寫入）", table, total-withCheck, total)
			}
		})
	}

	// metadicts 的 WITH CHECK 必須維持原樣（不含 department_id IS NULL）：放寬會讓
	// 一般租戶寫入系統預設字典。
	var metaCheck string
	if err := db.QueryRow(
		`SELECT with_check FROM pg_policies
		  WHERE tablename = 'metadicts' AND policyname = 'core_metadicts_scope'`,
	).Scan(&metaCheck); err != nil {
		t.Fatalf("查 metadicts policy: %v", err)
	}
	if containsAll(metaCheck, "department_id IS NULL") {
		t.Fatalf("core_metadicts_scope 的 WITH CHECK 被放寬（含 department_id IS NULL）：%s", metaCheck)
	}
}

func containsAll(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) &&
		(haystack == needle || indexOf(haystack, needle) >= 0)
}

func indexOf(h, n string) int {
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return i
		}
	}
	return -1
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && task test:integration -- -run TestIntegrationRLSPolicyCoverage -v`
Expected: FAIL —14 張表報「缺 WITH CHECK」、`customer_counters`／`product_units`／`product_processing_specs` 報「沒有 FOR ALL policy」

- [ ] **Step 3: 寫 migration（`database/migrations/00023_rls_policy_with_check.sql`）**

```sql
-- 補齊 RLS 的寫入面：14 個既有 policy 原本只有 USING(D35/D36)。
-- USING 管「可見列」，WITH CHECK 管「可寫入的新列」；缺 WITH CHECK 時 PostgreSQL
-- 以 USING 代替檢查，語意上仍是「只檢查新列」，但表達不完整且審計時看不出意圖。
-- 本波一律明示 WITH CHECK，且**等於原 USING**（不放寬任何既有語意）
-- 例外：core_metadicts_scope 已有 WITH CHECK 且刻意較嚴，不動。
-- +goose Up
-- +goose StatementBegin

-- 核心三表（原 00007）
DROP POLICY IF EXISTS core_companies_scope ON companies;
CREATE POLICY core_companies_scope ON companies FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (current_setting('app.current_company_id', true))::bigint = id
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (current_setting('app.current_company_id', true))::bigint = id
    );

DROP POLICY IF EXISTS core_departments_scope ON departments;
CREATE POLICY core_departments_scope ON departments FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_departments
        )
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'department'
            AND (current_setting('app.current_department_id', true))::bigint = id
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_departments
        )
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'department'
            AND (current_setting('app.current_department_id', true))::bigint = id
        )
    );

DROP POLICY IF EXISTS core_users_scope ON users;
CREATE POLICY core_users_scope ON users FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_users
        )
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'department'
            AND (current_setting('app.current_department_id', true))::bigint = department_users
        )
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'self'
            AND (current_setting('app.current_user_id', true))::bigint = id
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_users
        )
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'department'
            AND (current_setting('app.current_department_id', true))::bigint = department_users
        )
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'self'
            AND (current_setting('app.current_user_id', true))::bigint = id
        )
    );

-- 共享目錄：語意維持「已設定身分即可讀寫」不變（角色/權限的寫入權威仍在服務層 ACL，
-- 本波不在此處收緊，避免與 role_service 的授權矩陣不一致）。
DROP POLICY IF EXISTS core_roles_read ON roles;
CREATE POLICY core_roles_read ON roles FOR ALL
    USING (COALESCE(current_setting('app.current_data_scope', true), '') <> '')
    WITH CHECK (COALESCE(current_setting('app.current_data_scope', true), '') <> '');

DROP POLICY IF EXISTS core_role_permissions_read ON role_permissions;
CREATE POLICY core_role_permissions_read ON role_permissions FOR ALL
    USING (COALESCE(current_setting('app.current_data_scope', true), '') <> '')
    WITH CHECK (COALESCE(current_setting('app.current_data_scope', true), '') <> '');

-- 稽核（原 00010）
DROP POLICY IF EXISTS core_audit_logs_scope ON audit_logs;
CREATE POLICY core_audit_logs_scope ON audit_logs FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_id
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_id
        )
    );

-- 客戶域（原 00013／00015）：公司 + 部門（department_id IS NULL 為公司層共用列）
DROP POLICY IF EXISTS core_customers_scope ON customers;
CREATE POLICY core_customers_scope ON customers FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_customer_addresses_scope ON customer_addresses;
CREATE POLICY core_customer_addresses_scope ON customer_addresses FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_customer_contacts_scope ON customer_contacts;
CREATE POLICY core_customer_contacts_scope ON customer_contacts FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

-- 部門級主檔（原 00016）：四張同型
DROP POLICY IF EXISTS core_warehouses_scope ON warehouses;
CREATE POLICY core_warehouses_scope ON warehouses FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_routes_scope ON routes;
CREATE POLICY core_routes_scope ON routes FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_processing_specs_scope ON processing_specs;
CREATE POLICY core_processing_specs_scope ON processing_specs FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

DROP POLICY IF EXISTS core_product_categories_scope ON product_categories;
CREATE POLICY core_product_categories_scope ON product_categories FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

-- 商品（原 00017）
DROP POLICY IF EXISTS core_products_scope ON products;
CREATE POLICY core_products_scope ON products FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );

-- 漏網三表：原先完全沒有 policy
-- customer_counters 的 company_id 即主鍵
DROP POLICY IF EXISTS core_customer_counters_scope ON customer_counters;
CREATE POLICY core_customer_counters_scope ON customer_counters FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR company_id = (current_setting('app.current_company_id', true))::bigint
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR company_id = (current_setting('app.current_company_id', true))::bigint
    );

-- product_units / product_processing_specs 無 company_id：以父表 products 的存在性表達。
-- 子查詢本身也受 products 的 policy 約束（同一角色），故隔離具傳遞性。
DROP POLICY IF EXISTS core_product_units_scope ON product_units;
CREATE POLICY core_product_units_scope ON product_units FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR EXISTS (SELECT 1 FROM products p WHERE p.id = product_units.product_id)
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR EXISTS (SELECT 1 FROM products p WHERE p.id = product_units.product_id)
    );

DROP POLICY IF EXISTS core_product_processing_specs_scope ON product_processing_specs;
CREATE POLICY core_product_processing_specs_scope ON product_processing_specs FOR ALL
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR EXISTS (SELECT 1 FROM products p WHERE p.id = product_processing_specs.product_id)
    )
    WITH CHECK (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR EXISTS (SELECT 1 FROM products p WHERE p.id = product_processing_specs.product_id)
    );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- 還原為「只有 USING」的原狀，並移除本波新增的三個 policy
DROP POLICY IF EXISTS core_customer_counters_scope ON customer_counters;
DROP POLICY IF EXISTS core_product_units_scope ON product_units;
DROP POLICY IF EXISTS core_product_processing_specs_scope ON product_processing_specs;
DROP POLICY IF EXISTS core_companies_scope ON companies;
CREATE POLICY core_companies_scope ON companies
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (current_setting('app.current_company_id', true))::bigint = id
    );
DROP POLICY IF EXISTS core_departments_scope ON departments;
CREATE POLICY core_departments_scope ON departments
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_departments
        )
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'department'
            AND (current_setting('app.current_department_id', true))::bigint = id
        )
    );
DROP POLICY IF EXISTS core_users_scope ON users;
CREATE POLICY core_users_scope ON users
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_users
        )
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'department'
            AND (current_setting('app.current_department_id', true))::bigint = department_users
        )
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'self'
            AND (current_setting('app.current_user_id', true))::bigint = id
        )
    );
DROP POLICY IF EXISTS core_roles_read ON roles;
CREATE POLICY core_roles_read ON roles
    USING (COALESCE(current_setting('app.current_data_scope', true), '') <> '');
DROP POLICY IF EXISTS core_role_permissions_read ON role_permissions;
CREATE POLICY core_role_permissions_read ON role_permissions
    USING (COALESCE(current_setting('app.current_data_scope', true), '') <> '');
DROP POLICY IF EXISTS core_audit_logs_scope ON audit_logs;
CREATE POLICY core_audit_logs_scope ON audit_logs
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
            AND (current_setting('app.current_company_id', true))::bigint = company_id
        )
    );
DROP POLICY IF EXISTS core_customers_scope ON customers;
CREATE POLICY core_customers_scope ON customers
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
DROP POLICY IF EXISTS core_customer_addresses_scope ON customer_addresses;
CREATE POLICY core_customer_addresses_scope ON customer_addresses
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
DROP POLICY IF EXISTS core_customer_contacts_scope ON customer_contacts;
CREATE POLICY core_customer_contacts_scope ON customer_contacts
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
DROP POLICY IF EXISTS core_warehouses_scope ON warehouses;
CREATE POLICY core_warehouses_scope ON warehouses
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
DROP POLICY IF EXISTS core_routes_scope ON routes;
CREATE POLICY core_routes_scope ON routes
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
DROP POLICY IF EXISTS core_processing_specs_scope ON processing_specs;
CREATE POLICY core_processing_specs_scope ON processing_specs
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
DROP POLICY IF EXISTS core_product_categories_scope ON product_categories;
CREATE POLICY core_product_categories_scope ON product_categories
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
DROP POLICY IF EXISTS core_products_scope ON products;
CREATE POLICY core_products_scope ON products
    USING (
        COALESCE(current_setting('app.current_data_scope', true), '') = 'all'
        OR (
            company_id = (current_setting('app.current_company_id', true))::bigint
            AND (
                COALESCE(current_setting('app.current_data_scope', true), '') = 'company'
                OR department_id IS NULL
                OR department_id = (current_setting('app.current_department_id', true))::bigint
            )
        )
    );
-- +goose StatementEnd
```

- [ ] **Step 4: 跑測試確認通過**

Run: `cd backend && task test:integration -- -run TestIntegrationRLSPolicyCoverage -v`
Expected: PASS（18 張表全數有 WITH CHECK；metadicts 未被放寬）

- [ ] **Step 5: 驗證遷移可完整回滾**

Run: `cd backend && task test:integration -- -run TestIntegrationMigrateDown -v`（既有回滾測試）
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/database/migrations/00023_rls_policy_with_check.sql backend/internal/services/rls_policy_integration_test.go
git commit -m "feat(backend): RLS policy 補 WITH CHECK 並補三張漏網表（00023）"
```

---

### Task 4: 請求層租戶交易（`internal/dbtenant`）

**Files:**
- Create: `internal/dbtenant/dbtenant.go`
- Create: `internal/dbtenant/dbtenant_test.go`
- Create: `internal/dbtenant/tenant_integration_test.go`
- Modify: `internal/server/domains.go`（每個 `NewXServiceHandler` 加 handler option）

> **⚠️ 設計更正（2026-09-20，開工後實測）**：原版計畫要 `auth.ApplyRLS(ctx, tx, scope)` 直接套在 `*ent.Tx` 上。實測本專案產生的 ent 型別（ent v0.14.6）**沒有** `ExecContext`（`ent/tx.go` 只有 `Commit`／`Rollback`／`Client`／`OnCommit`／`OnRollback`），且模組內**沒有** `entsql.OpenTx`（只有 `Open`／`OpenDB`）——所以原版程式碼**編譯不過**。改採 **`dialect.Driver` 裝飾器**：ent 的 `client.Tx(ctx)` 會呼叫 `driver.Tx(ctx)`，我們在那裡把 ctx 的 RLS scope 用 `dialect.Tx.Exec` 套進剛開好的交易。好處：**服務層既有的 42 處 `client.Tx(ctx)` 自動變成 RLS 安全，不必逐一改**；代價：仍須改讀取路徑（未包交易的查詢會 fail-closed 黑屏）。**對外 API 不變**（interceptor 仍持有 `*ent.Tx` 並負責 commit／rollback）。

**Interfaces:**
- Consumes: `auth.RLSFrom(ctx)`、`auth.RLSStatements(scope)`、`dialect.Driver`／`dialect.Tx`／`entsql.OpenDB`
- Produces:
  - `dbtenant.NewClient(db *sql.DB) *ent.Client`（＝`ent.NewClient(ent.Driver(Wrap(entsql.OpenDB(dialect.Postgres, db))))`；**業務 client 必須由它建立**，裝飾器才會生效）
  - `dbtenant.Wrap(inner dialect.Driver) dialect.Driver`
  - `dbtenant.Client(ctx context.Context, fallback *ent.Client) *ent.Client`
  - `dbtenant.WithTenantTx(ctx context.Context, tx *ent.Tx) context.Context`
  - `dbtenant.TxFrom(ctx context.Context) (*ent.Tx, bool)`
  - `dbtenant.Interceptor(client *ent.Client) connect.Interceptor`
  - `dbtenant.HandlerOption(client *ent.Client) connect.HandlerOption`
  - `dbtenant.SystemScopeTx(ctx context.Context, client *ent.Client, fn func(*ent.Tx) error) error`

- [ ] **Step 1: 寫失敗測試（`internal/dbtenant/dbtenant_test.go`）**

```go
package dbtenant

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/salesorder/sales-order-1.0/backend/ent"
	_ "modernc.org/sqlite"
)

// Client 在沒有請求交易時必須退回傳入的 client（CLI／seed／既有 sqlite 測試都靠這條）。
func TestClientFallsBackWithoutRequestTx(t *testing.T) {
	db, err := sqlOpen(t)
	if err != nil {
		t.Fatalf("開 sqlite: %v", err)
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	t.Cleanup(func() { _ = client.Close() })

	if got := Client(context.Background(), client); got != client {
		t.Fatal("無請求交易時應回傳 fallback client")
	}
}

func TestClientUsesRequestTx(t *testing.T) {
	db, err := sqlOpen(t)
	if err != nil {
		t.Fatalf("開 sqlite: %v", err)
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db)))
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
	if _, ok := TxFrom(WithTenantTx(ctx, tx)); !ok {
		t.Fatal("TxFrom 應取得請求交易")
	}
	if _, ok := TxFrom(ctx); ok {
		t.Fatal("未注入時 TxFrom 應回 false")
	}
}
```

`schema/init.sql`：測試檔內以 `CREATE TABLE` 建最小表（本次僅需能開交易與查詢，欄位可為空表）：

```go
func sqlOpen(t *testing.T) (*sql.DB, error) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:dbtenant?mode=memory&cache=shared&_fk=1")
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS ping (id INTEGER PRIMARY KEY)`); err != nil {
		return nil, err
	}
	return db, nil
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && go test ./internal/dbtenant/ -v`
Expected: FAIL —`undefined: Client`／`WithTenantTx`／`TxFrom`

- [ ] **Step 3: 實作（`internal/dbtenant/dbtenant.go`）**

```go
// Package dbtenant 提供「請求層租戶交易」：每個 unary RPC 開一個交易，交易內以
// SET LOCAL app.* 套用 RLS scope，服務層統一由 Client(ctx, s.db) 取得被約束的 client。
//
// 為何不是逐呼叫點包交易：服務層有 124 處直呼查詢與 42 處自開交易，逐點包會漏；
// 交易邊界改由請求擁有（spec §6.3）。串流 RPC 與長時工作（未來 WatchBoard／PDF）
// 不得沿用此法，屆時另立短交易邊界。
//
// 為何用 driver 裝飾器而不是 auth.ApplyRLS(ctx, tx, scope)：本專案產生的 ent 型別
// 沒有 ExecContext，ent 也沒有 OpenTx 可把 *sql.Tx 綁成 client。ent 的 client.Tx(ctx)
// 會呼叫 driver.Tx(ctx)，我們就在那裡把 ctx 的 RLS scope 套進剛開好的交易 ——
// 於是**服務層任何 client.Tx(ctx) 都自動 RLS 安全**（42 處不必逐一改）。
package dbtenant

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
)

type txCtxKey struct{}

// NewClient 建立業務 ent client：**必須**經此建立，RLS 裝飾器才會生效。
// 其他 ent client（CLI／seed／測試 fixture）走原本的 entsql.OpenDB，不受影響。
func NewClient(db *sql.DB) *ent.Client {
	return ent.NewClient(ent.Driver(Wrap(entsql.OpenDB(dialect.Postgres, db))))
}

// Wrap 以 RLS 裝飾器包住 dialect driver。
func Wrap(inner dialect.Driver) dialect.Driver { return &rlsDriver{inner: inner} }

type rlsDriver struct{ inner dialect.Driver }

func (d *rlsDriver) Exec(ctx context.Context, query string, args, v any) error {
	return d.inner.Exec(ctx, query, args, v)
}

func (d *rlsDriver) Query(ctx context.Context, query string, args, v any) error {
	return d.inner.Query(ctx, query, args, v)
}

func (d *rlsDriver) Close() error     { return d.inner.Close() }
func (d *rlsDriver) Dialect() string  { return d.inner.Dialect() }

// Tx 開交易並**立刻**套用 ctx 的 RLS scope：SET LOCAL 只在當前交易有效，
// 而這裡正是交易剛開好、任何業務查詢之前 —— 錯過這個點就再也補不上。
func (d *rlsDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	tx, err := d.inner.Tx(ctx)
	if err != nil {
		return nil, err
	}
	for _, stmt := range auth.RLSStatements(auth.RLSFrom(ctx)) {
		// dialect.Tx.Exec 的 v 參數對 SQL driver 而言是 *sql.Result（entsql.Result 是別名介面，
		// 不能寫 `&entsql.Result{}` —— 要用具體變數取址）。
		var res sql.Result
		if err := tx.Exec(ctx, stmt, []any{}, &res); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("dbtenant: 套用 RLS 失敗(%s): %w", stmt, err)
		}
	}
	return tx, nil
}

// WithTenantTx 把請求交易放進 ctx。
func WithTenantTx(ctx context.Context, tx *ent.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

// TxFrom 取出請求交易；未注入時回 false（CLI／seed／單元測試）。
func TxFrom(ctx context.Context) (*ent.Tx, bool) {
	tx, ok := ctx.Value(txCtxKey{}).(*ent.Tx)
	return tx, ok
}

// Client 回傳當前請求的 scoped client；沒有請求交易時退回 fallback。
func Client(ctx context.Context, fallback *ent.Client) *ent.Client {
	if tx, ok := TxFrom(ctx); ok {
		return tx.Client()
	}
	return fallback
}

// Interceptor 為 unary RPC 開租戶交易：開交易（RLS 由 Wrap 的 driver 裝飾器在
// Tx(ctx) 內套用，此處只負責交易邊界）→ 呼叫 handler → err == nil 則 commit、否則 rollback。
// 以 interceptor 而非 HTTP middleware 的理由：interceptor 看得到 domain error
// （HTTP 狀態碼在 connect 下與錯誤碼的對應是間接的）。
func Interceptor(client *ent.Client) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			tx, err := client.Tx(ctx)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, errors.New("開啟租戶交易失敗"))
			}
			resp, err := next(WithTenantTx(ctx, tx), req)
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
			if err := tx.Commit(); err != nil {
				return nil, connect.NewError(connect.CodeInternal, errors.New("提交交易失敗"))
			}
			return resp, nil
		}
	})
}

// HandlerOption 供 NewXServiceHandler 掛載租戶交易。
func HandlerOption(client *ent.Client) connect.HandlerOption {
	return connect.WithInterceptors(Interceptor(client))
}

// SystemScopeTx 在明確的系統範圍（scope=all）內執行 fn：供「尚無身分」的路徑使用
// （登入憑證查詢、authzMiddleware 的身分解析、seed）。刻意獨立成一個入口，
// 讓「系統範圍」在呼叫點顯眼可審計，而不是散落的 SET LOCAL。
func SystemScopeTx(ctx context.Context, client *ent.Client, fn func(*ent.Tx) error) error {
	// scope=all 必須在**開交易之前**注入 ctx：driver 裝飾器在 Tx(ctx) 內讀它。
	ctx = auth.WithRLS(ctx, auth.RLSScope{DataScope: auth.DataScopeAll})
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
```

- [ ] **Step 4: 掛載 interceptor（`internal/services/*.go` 的 `Register*Services` ＋ `internal/server/domains.go`）**

**掛在 register helper 內，而非呼叫端**：這樣 production 與測試（兩者都經 `Register*Services`）自動一致，不必改 11 個函式簽名。每個 helper 內把

```go
salesorderv1connect.NewCompanyServiceHandler(svc)
```

改成

```go
salesorderv1connect.NewCompanyServiceHandler(svc, dbtenant.HandlerOption(db))
```

逐一處理 `RegisterCompanyServices`、`RegisterUserServices`、`RegisterMetadictServices`、`RegisterAuditServices`、`RegisterCustomerServices`、`RegisterWarehouseService`、`RegisterRouteService`、`RegisterProcessingSpecService`、`RegisterProductCategoryService`、`RegisterProductService`、`RegisterRoleServices`，以及 `internal/server/domains.go` 內直接建構的 `NewAuthServiceHandler`、`NewAbilityServiceHandler`（`handlers.RegisterRoleHandler` 內若有比照辦理）。

REST 端點（OIDC 回調、QR）不掛 interceptor —— 它們屬未登入路徑，改以 Task 9 的 `SystemScopeTx` 包住其 DB 查詢。

- [ ] **Step 5: 整合測試（`internal/dbtenant/tenant_integration_test.go`）**

契約是「成功提交、失敗回滾、handler 內拿得到請求交易」。GUC 是否真的套上，由 Task 10 的跨租戶端到端探針證明（RLS 啟用後若未 `SET LOCAL`，資料會全數不可見 → 探針必紅），故此處不重複以 `current_setting` 觀測（ent 未產生可讀 GUC 的 raw query 介面）。

```go
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
	"github.com/pressly/goose/v3"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationTenantTxCommitAndRollback 驗證 interceptor 的交易語意：
// 成功 → 交易內寫入落地；handler 回傳錯誤 → 整筆回滾；handler 內取得請求交易。
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
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, sqlDB)))
	t.Cleanup(func() { _ = client.Close() })

	ctx := auth.WithRLS(context.Background(), auth.RLSScope{
		UserID: "1", CompanyID: "1", DataScope: auth.DataScopeAll, CompanyActive: true,
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
```

- [ ] **Step 6: 跑測試確認通過**

Run: `cd backend && go test ./internal/dbtenant/ -v && task test:integration -- -run TestIntegrationTenantTxCommitAndRollback -v`
Expected: 單元 PASS；整合 PASS（提交 1 列、回滾 0 列）

- [ ] **Step 7: Commit**

```bash
git add backend/internal/dbtenant backend/internal/server/domains.go
git commit -m "feat(backend): 請求層租戶交易 interceptor 與 dbtenant.Client（RLS scope 貫穿）"
```

---

### Task 5: 客戶域遷移 + ENABLE（migration 00024）

**Files:**
- Create: `database/migrations/00024_rls_enable_customers.sql`
- Create: `internal/services/rls_customers_integration_test.go`
- Modify: `internal/services/customer_service.go`（14 處 `s.db.`、4 處 `db.Tx(ctx)`）
- Modify: `internal/services/customer_address_service.go`（7 處、3 處）
- Modify: `internal/services/customer_contact_service.go`（6 處、3 處）

**Interfaces:**
- Consumes: `dbtenant.Client(ctx, s.db)`
- Produces: `customers`／`customer_counters`／`customer_addresses`／`customer_contacts` 四表 ENABLE + FORCE

- [ ] **Step 1: 寫失敗測試（`internal/services/rls_customers_integration_test.go`）**

```go
//go:build integration

package services

import (
	"database/sql"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRLSCustomersIsolation 以 app_rw 直連驗證客戶域隔離：
// 未設 scope → 看不到任何列（fail-closed）；設 A 公司 → 只看得到 A；
// 寫入 B 公司的列 → 被 WITH CHECK 擋下。
func TestIntegrationRLSCustomersIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)

	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		t.Fatalf("admin 連線: %v", err)
	}
	defer admin.Close()
	var coA, coB int
	if err := admin.QueryRow(
		`INSERT INTO companies (name, identifier, status, created_at, updated_at)
		 VALUES ('A', 'RLS-A', 'active', now(), now()) RETURNING id`).Scan(&coA); err != nil {
		t.Fatalf("建 A 公司: %v", err)
	}
	if err := admin.QueryRow(
		`INSERT INTO companies (name, identifier, status, created_at, updated_at)
		 VALUES ('B', 'RLS-B', 'active', now(), now()) RETURNING id`).Scan(&coB); err != nil {
		t.Fatalf("建 B 公司: %v", err)
	}
	for _, co := range []int{coA, coB} {
		if _, err := admin.Exec(
			`INSERT INTO customers (company_id, customer_code, name, created_at, updated_at)
			 VALUES ($1, $2, '客戶', now(), now())`, co, "C-"+string(rune('A'+co))); err != nil {
			t.Fatalf("建客戶(公司 %d): %v", co, err)
		}
	}

	appDSN := testsupport.AppRoleDSN(t, adminDSN)
	app, err := sql.Open("pgx", appDSN)
	if err != nil {
		t.Fatalf("app 連線: %v", err)
	}
	defer app.Close()

	t.Run("未設 scope → fail-closed", func(t *testing.T) {
		var n int
		if err := app.QueryRow(`SELECT count(*) FROM customers`).Scan(&n); err != nil {
			t.Fatalf("查詢: %v", err)
		}
		if n != 0 {
			t.Fatalf("未設 scope 時不得看到任何客戶,got %d", n)
		}
	})

	t.Run("scope=company A → 只見 A", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		for _, stmt := range []string{
			`SET LOCAL app.current_data_scope = 'company'`,
			`SET LOCAL app.current_company_id = '` + itoa(coA) + `'`,
		} {
			if _, err := tx.Exec(stmt); err != nil {
				t.Fatalf("%s: %v", stmt, err)
			}
		}
		var got int
		if err := tx.QueryRow(`SELECT DISTINCT company_id FROM customers`).Scan(&got); err != nil {
			t.Fatalf("查詢: %v", err)
		}
		if got != coA {
			t.Fatalf("應只看到公司 %d,got %d", coA, got)
		}
	})

	t.Run("跨租戶 INSERT → 被 WITH CHECK 擋", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		if _, err := tx.Exec(
			`SET LOCAL app.current_data_scope = 'company'`); err != nil {
			t.Fatalf("SET scope: %v", err)
		}
		if _, err := tx.Exec(
			`SET LOCAL app.current_company_id = '` + itoa(coA) + `'`); err != nil {
			t.Fatalf("SET company: %v", err)
		}
		_, err = tx.Exec(
			`INSERT INTO customers (company_id, customer_code, name, created_at, updated_at)
			 VALUES ($1, 'X-1', '別家客戶', now(), now())`, coB)
		if err == nil {
			t.Fatal("以 A 的身分寫入 B 公司的客戶必須被擋（WITH CHECK）")
		}
	})
}

// itoa 為本檔（同套件其他 RLS 整合測試共用）的整數字串轉換輔助：
// SET LOCAL 的參數無法以 $1 綁定，故以字串拼接（值來自測試 fixture 的整數，非使用者輸入）。
func itoa(i int) string { return strconv.Itoa(i) }
```

（`strconv` 需加入該檔 import。）

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && task test:integration -- -run TestIntegrationRLSCustomersIsolation -v`
Expected: FAIL —失敗點是「未設 scope → fail-closed」（此刻該表還沒 ENABLE，仍看得到列）

- [ ] **Step 3: 遷移服務路徑（三檔）**

以 `customer_service.go` 為例（其餘兩檔同法），先把 `s.db.` 一律改為 `dbtenant.Client(ctx, s.db).`：

```bash
cd backend
sd 's\.db\.' 'dbtenant.Client(ctx, s.db).' internal/services/customer_service.go
sd 's\.db\.' 'dbtenant.Client(ctx, s.db).' internal/services/customer_address_service.go
sd 's\.db\.' 'dbtenant.Client(ctx, s.db).' internal/services/customer_contact_service.go
grep -c 's\.db\.' internal/services/customer_service.go   # 期望 0（僅剩取 client 的呼叫）
```

再把該檔 4 處自開交易改為使用請求交易。原形（`DeleteCustomer` 等）：

```go
tx, err := s.db.Tx(ctx)
if err != nil {
	return nil, toConnectError(err)
}
// … 使用 tx.…
if err := tx.Commit(); err != nil {
	return nil, toConnectError(err)
}
```

改成：

```go
db := dbtenant.Client(ctx, s.db) // 請求交易（無請求時為原 client，供 CLI／測試）
// … 全部改用 db.…（不再 commit/rollback：交易由 interceptor 擁有）
```

逐檔完成後確認已無自開交易：

```bash
grep -n 'db.Tx(ctx)' internal/services/customer_service.go internal/services/customer_address_service.go internal/services/customer_contact_service.go
```
Expected: 無輸出

- [ ] **Step 4: 寫 ENABLE migration（`database/migrations/00024_rls_enable_customers.sql`）**

```sql
-- 客戶域啟用 RLS（D36）。順序:policy(00023) → ENABLE → FORCE。
-- FORCE 也約束 owner，故日後對本域做資料回填的 migration 必須先 SET LOCAL app.current_data_scope='all'。
-- +goose Up
ALTER TABLE customers            ENABLE ROW LEVEL SECURITY;
ALTER TABLE customers            FORCE  ROW LEVEL SECURITY;
ALTER TABLE customer_counters    ENABLE ROW LEVEL SECURITY;
ALTER TABLE customer_counters    FORCE  ROW LEVEL SECURITY;
ALTER TABLE customer_addresses   ENABLE ROW LEVEL SECURITY;
ALTER TABLE customer_addresses   FORCE  ROW LEVEL SECURITY;
ALTER TABLE customer_contacts    ENABLE ROW LEVEL SECURITY;
ALTER TABLE customer_contacts    FORCE  ROW LEVEL SECURITY;

-- +goose Down
ALTER TABLE customers            NO FORCE ROW LEVEL SECURITY;
ALTER TABLE customers            DISABLE ROW LEVEL SECURITY;
ALTER TABLE customer_counters    NO FORCE ROW LEVEL SECURITY;
ALTER TABLE customer_counters    DISABLE ROW LEVEL SECURITY;
ALTER TABLE customer_addresses   NO FORCE ROW LEVEL SECURITY;
ALTER TABLE customer_addresses   DISABLE ROW LEVEL SECURITY;
ALTER TABLE customer_contacts    NO FORCE ROW LEVEL SECURITY;
ALTER TABLE customer_contacts    DISABLE ROW LEVEL SECURITY;
```

- [ ] **Step 5: 跑整合測試（新測試 + 客戶域既有測試）**

Run: `cd backend && task test:integration -- -run 'TestIntegrationRLS|TestIntegrationCustomer' -v`
Expected: 全 PASS。**若客戶域既有整合測試失敗且錯誤為「查不到資料」→ 代表仍有路徑未走 `dbtenant.Client`**，回頭補（Step 3 的 grep 只涵蓋 `s.db.`，其他別名（如 `s.sqlDB`、自有 helper）也要查）。

- [ ] **Step 6: 全套守門**

Run: `cd backend && task check && task test:integration`
Expected: 全綠

- [ ] **Step 7: Commit**

```bash
git add backend/database/migrations/00024_rls_enable_customers.sql backend/internal/services/customer_service.go backend/internal/services/customer_address_service.go backend/internal/services/customer_contact_service.go backend/internal/services/rls_customers_integration_test.go
git commit -m "feat(backend): 客戶域啟用 RLS 並將路徑收斂至請求層租戶交易（00024）"
```

---

### Task 6: 主檔域（倉別／車次／加工規格／商品分類）

**Files:**
- Create: `database/migrations/00025_rls_enable_masters.sql`
- Create: `internal/services/rls_masters_integration_test.go`
- Modify: `internal/services/warehouse_service.go`（8 處、4 處）
- Modify: `internal/services/route_service.go`（8 處、4 處）
- Modify: `internal/services/processingspec_service.go`（8 處、4 處）
- Modify: `internal/services/productcategory_service.go`（8 處、4 處）
- Modify: `internal/services/department_master.go`（共用 helper，若有 DirectDB 存取）

**Interfaces:**
- Consumes: `dbtenant.Client(ctx, s.db)`
- Produces: `warehouses`／`routes`／`processing_specs`／`product_categories` ENABLE + FORCE

- [ ] **Step 1: 寫失敗測試（`internal/services/rls_masters_integration_test.go`）**

以 Task 5 的探針為樣板，物件名與表名換成四張主檔表；斷言三條：未設 scope → 0 列、scope=company A → 只見 A（各表以 `company_id` 判別）、以 A 身分寫入 B 公司的列 → 被擋。

```go
//go:build integration

package services

import (
	"database/sql"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRLSMastersIsolation 涵蓋 warehouses/routes/processing_specs/product_categories 四表。
func TestIntegrationRLSMastersIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)
	appDSN := testsupport.AppRoleDSN(t, adminDSN)

	tables := []struct {
		name   string
		insert string // 以 $1=company_id 插入一列的 SQL（其餘欄位取預設/常數）
	}{
		{"warehouses", `INSERT INTO warehouses (company_id, code, name, created_at, updated_at) VALUES ($1,'W1','倉',now(),now())`},
		{"routes", `INSERT INTO routes (company_id, code, name, created_at, updated_at) VALUES ($1,'R1','車次',now(),now())`},
		{"processing_specs", `INSERT INTO processing_specs (company_id, code, name, created_at, updated_at) VALUES ($1,'P1','規格',now(),now())`},
		{"product_categories", `INSERT INTO product_categories (company_id, code, name, created_at, updated_at) VALUES ($1,'C1','分類',now(),now())`},
	}
	for _, tbl := range tables {
		t.Run(tbl.name, func(t *testing.T) {
			admin, err := sql.Open("pgx", adminDSN)
			if err != nil {
				t.Fatalf("admin 連線: %v", err)
			}
			defer admin.Close()
			var coA, coB int
			if err := admin.QueryRow(
				`INSERT INTO companies (name, identifier, status, created_at, updated_at)
				 VALUES ('A','A-'||$1,'active',now(),now()) RETURNING id`, tbl.name).Scan(&coA); err != nil {
				t.Fatalf("建 A 公司: %v", err)
			}
			if err := admin.QueryRow(
				`INSERT INTO companies (name, identifier, status, created_at, updated_at)
				 VALUES ('B','B-'||$1,'active',now(),now()) RETURNING id`, tbl.name).Scan(&coB); err != nil {
				t.Fatalf("建 B 公司: %v", err)
			}
			if _, err := admin.Exec(tbl.insert, coA); err != nil {
				t.Fatalf("A 的資料: %v", err)
			}
			if _, err := admin.Exec(tbl.insert, coB); err != nil {
				t.Fatalf("B 的資料: %v", err)
			}

			app, err := sql.Open("pgx", appDSN)
			if err != nil {
				t.Fatalf("app 連線: %v", err)
			}
			defer app.Close()
			var noScope int
			if err := app.QueryRow(`SELECT count(*) FROM ` + tbl.name).Scan(&noScope); err != nil {
				t.Fatalf("未設 scope 查詢: %v", err)
			}
			if noScope != 0 {
				t.Fatalf("未設 scope 時 %s 必須 0 列,got %d", tbl.name, noScope)
			}

			tx, err := app.Begin()
			if err != nil {
				t.Fatalf("開交易: %v", err)
			}
			defer func() { _ = tx.Rollback() }()
			if _, err := tx.Exec(`SET LOCAL app.current_data_scope = 'company'`); err != nil {
				t.Fatalf("SET scope: %v", err)
			}
			if _, err := tx.Exec(`SET LOCAL app.current_company_id = '` + itoa(coA) + `'`); err != nil {
				t.Fatalf("SET company: %v", err)
			}
			var got int
			if err := tx.QueryRow(`SELECT DISTINCT company_id FROM ` + tbl.name).Scan(&got); err != nil {
				t.Fatalf("查詢: %v", err)
			}
			if got != coA {
				t.Fatalf("%s 應只看到公司 %d,got %d", tbl.name, coA, got)
			}
			if _, err := tx.Exec(tbl.insert, coB); err == nil {
				t.Fatalf("以 A 的身分寫入 B 公司的 %s 必須被擋", tbl.name)
			}
		})
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && task test:integration -- -run TestIntegrationRLSMastersIsolation -v`
Expected: FAIL —未設 scope 時仍看到列（表未 ENABLE）

- [ ] **Step 3: 遷移四檔路徑**

```bash
cd backend
for f in warehouse_service route_service processingspec_service productcategory_service; do
  sd 's\.db\.' 'dbtenant.Client(ctx, s.db).' internal/services/$f.go
done
grep -n 's\.db\.\|db.Tx(ctx)' internal/services/warehouse_service.go internal/services/route_service.go internal/services/processingspec_service.go internal/services/productcategory_service.go
```
Expected: 無輸出。四檔各自的自開交易（各 4 處）依 Global Constraints「交易改寫樣式」改寫為使用請求交易；若 `department_master.go` 內有共用存取，一併處理。

- [ ] **Step 4: ENABLE migration（`database/migrations/00025_rls_enable_masters.sql`）**

```sql
-- 部門級主檔啟用 RLS（D36）。
-- +goose Up
ALTER TABLE warehouses          ENABLE ROW LEVEL SECURITY;
ALTER TABLE warehouses          FORCE  ROW LEVEL SECURITY;
ALTER TABLE routes              ENABLE ROW LEVEL SECURITY;
ALTER TABLE routes              FORCE  ROW LEVEL SECURITY;
ALTER TABLE processing_specs    ENABLE ROW LEVEL SECURITY;
ALTER TABLE processing_specs    FORCE  ROW LEVEL SECURITY;
ALTER TABLE product_categories  ENABLE ROW LEVEL SECURITY;
ALTER TABLE product_categories  FORCE  ROW LEVEL SECURITY;

-- +goose Down
ALTER TABLE warehouses          NO FORCE ROW LEVEL SECURITY;
ALTER TABLE warehouses          DISABLE ROW LEVEL SECURITY;
ALTER TABLE routes              NO FORCE ROW LEVEL SECURITY;
ALTER TABLE routes              DISABLE ROW LEVEL SECURITY;
ALTER TABLE processing_specs    NO FORCE ROW LEVEL SECURITY;
ALTER TABLE processing_specs    DISABLE ROW LEVEL SECURITY;
ALTER TABLE product_categories  NO FORCE ROW LEVEL SECURITY;
ALTER TABLE product_categories  DISABLE ROW LEVEL SECURITY;
```

- [ ] **Step 5: 跑測試**

Run: `cd backend && task test:integration -- -run 'TestIntegrationRLSMasters|TestIntegrationListPageScan' -v`
Expected: 全 PASS

- [ ] **Step 6: Commit**

```bash
git add backend/database/migrations/00025_rls_enable_masters.sql backend/internal/services/warehouse_service.go backend/internal/services/route_service.go backend/internal/services/processingspec_service.go backend/internal/services/productcategory_service.go backend/internal/services/department_master.go backend/internal/services/rls_masters_integration_test.go
git commit -m "feat(backend): 部門級主檔啟用 RLS 並收斂路徑（00025）"
```

---

### Task 7: 商品域（products／product_units／product_processing_specs）

**Files:**
- Create: `database/migrations/00026_rls_enable_products.sql`
- Create: `internal/services/rls_products_integration_test.go`
- Modify: `internal/services/product_service.go`（15 處、4 處）

**Interfaces:**
- Consumes: `dbtenant.Client(ctx, s.db)`
- Produces: 三表 ENABLE + FORCE

- [ ] **Step 1: 寫失敗測試（`internal/services/rls_products_integration_test.go`）**

以 Task 5 探針為樣板，額外涵蓋**子表傳遞性**（`product_units`／`product_processing_specs` 靠父表 `products` 的 policy）：

```go
//go:build integration

package services

import (
	"database/sql"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRLSProductsIsolation 驗證商品三表隔離，含子表以父表存在性表達的傳遞性：
// product_units 沒有 company_id，只有所屬商品可見時才可見。
func TestIntegrationRLSProductsIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)
	appDSN := testsupport.AppRoleDSN(t, adminDSN)

	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		t.Fatalf("admin 連線: %v", err)
	}
	defer admin.Close()

	var coA, coB, prodA, prodB int
	if err := admin.QueryRow(`INSERT INTO companies (name, identifier, status, created_at, updated_at)
		VALUES ('A','PRD-A','active',now(),now()) RETURNING id`).Scan(&coA); err != nil {
		t.Fatalf("建 A: %v", err)
	}
	if err := admin.QueryRow(`INSERT INTO companies (name, identifier, status, created_at, updated_at)
		VALUES ('B','PRD-B','active',now(),now()) RETURNING id`).Scan(&coB); err != nil {
		t.Fatalf("建 B: %v", err)
	}
	if err := admin.QueryRow(`INSERT INTO products (company_id, code, name, created_at, updated_at)
		VALUES ($1,'PA','商品A',now(),now()) RETURNING id`, coA).Scan(&prodA); err != nil {
		t.Fatalf("建 A 商品: %v", err)
	}
	if err := admin.QueryRow(`INSERT INTO products (company_id, code, name, created_at, updated_at)
		VALUES ($1,'PB','商品B',now(),now()) RETURNING id`, coB).Scan(&prodB); err != nil {
		t.Fatalf("建 B 商品: %v", err)
	}
	if _, err := admin.Exec(`INSERT INTO product_units (product_id, unit_code, conversion_rate, is_base)
		VALUES ($1,'箱',1,true), ($2,'箱',1,true)`, prodA, prodB); err != nil {
		t.Fatalf("建單位: %v", err)
	}

	app, err := sql.Open("pgx", appDSN)
	if err != nil {
		t.Fatalf("app 連線: %v", err)
	}
	defer app.Close()

	tx, err := app.Begin()
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`SET LOCAL app.current_data_scope = 'company'`); err != nil {
		t.Fatalf("SET scope: %v", err)
	}
	if _, err := tx.Exec(`SET LOCAL app.current_company_id = '` + itoa(coA) + `'`); err != nil {
		t.Fatalf("SET company: %v", err)
	}

	var units int
	if err := tx.QueryRow(`SELECT count(*) FROM product_units`).Scan(&units); err != nil {
		t.Fatalf("查子表: %v", err)
	}
	if units != 1 {
		t.Fatalf("子表應只看到 A 的一筆單位（父表存在性傳遞）,got %d", units)
	}
	var prodIDs int
	if err := tx.QueryRow(`SELECT count(DISTINCT product_id) FROM product_units`).Scan(&prodIDs); err != nil {
		t.Fatalf("查子表商品: %v", err)
	}
	if prodIDs != 1 {
		t.Fatalf("子表不得看到 B 的商品單位,got %d 個 product_id", prodIDs)
	}
	if _, err := tx.Exec(`INSERT INTO product_units (product_id, unit_code, conversion_rate, is_base)
		VALUES ($1,'包',2,false)`, prodB); err == nil {
		t.Fatal("以 A 的身分對 B 的商品新增單位必須被擋（父表不可見）")
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && task test:integration -- -run TestIntegrationRLSProductsIsolation -v`
Expected: FAIL —子表看到 2 筆（表未 ENABLE）

- [ ] **Step 3: 遷移 `product_service.go`**

```bash
cd backend
sd 's\.db\.' 'dbtenant.Client(ctx, s.db).' internal/services/product_service.go
grep -n 's\.db\.\|db.Tx(ctx)' internal/services/product_service.go   # 期望無輸出
```
四處自開交易依 Global Constraints「交易改寫樣式」改寫為使用請求交易。

- [ ] **Step 4: ENABLE migration（`database/migrations/00026_rls_enable_products.sql`）**

```sql
-- 商品域啟用 RLS（含子表：product_units／product_processing_specs 以父表 products 的存在性表達）。
-- +goose Up
ALTER TABLE products                  ENABLE ROW LEVEL SECURITY;
ALTER TABLE products                  FORCE  ROW LEVEL SECURITY;
ALTER TABLE product_units             ENABLE ROW LEVEL SECURITY;
ALTER TABLE product_units             FORCE  ROW LEVEL SECURITY;
ALTER TABLE product_processing_specs  ENABLE ROW LEVEL SECURITY;
ALTER TABLE product_processing_specs  FORCE  ROW LEVEL SECURITY;

-- +goose Down
ALTER TABLE products                  NO FORCE ROW LEVEL SECURITY;
ALTER TABLE products                  DISABLE ROW LEVEL SECURITY;
ALTER TABLE product_units             NO FORCE ROW LEVEL SECURITY;
ALTER TABLE product_units             DISABLE ROW LEVEL SECURITY;
ALTER TABLE product_processing_specs  NO FORCE ROW LEVEL SECURITY;
ALTER TABLE product_processing_specs  DISABLE ROW LEVEL SECURITY;
```

- [ ] **Step 5: 跑測試**

Run: `cd backend && task test:integration -- -run 'TestIntegrationRLSProducts|TestIntegrationListPageScan' -v`
Expected: 全 PASS

- [ ] **Step 6: Commit**

```bash
git add backend/database/migrations/00026_rls_enable_products.sql backend/internal/services/product_service.go backend/internal/services/rls_products_integration_test.go
git commit -m "feat(backend): 商品域（含子表）啟用 RLS 並收斂路徑（00026）"
```

---

### Task 8: 字典與稽核（metadicts／audit_logs）

**Files:**
- Create: `database/migrations/00027_rls_enable_metadicts_audit.sql`
- Create: `internal/services/rls_metadict_audit_integration_test.go`
- Modify: `internal/services/metadict_service.go`（8 處、3 處）
- Modify: `internal/services/audit_service.go`（2 處）
- Modify: `internal/audit/recorder.go`（稽核寫入需在同一交易；確認其以傳入 tx 執行）

**Interfaces:**
- Consumes: `dbtenant.Client(ctx, s.db)`
- Produces: `metadicts`／`audit_logs` ENABLE + FORCE

- [ ] **Step 1: 寫失敗測試（`internal/services/rls_metadict_audit_integration_test.go`）**

三條斷言：
1. `audit_logs`：未設 scope → 0 列；scope=company A → 只見 A。
2. `metadicts` 系統預設列（`department_id IS NULL`）在 `scope=company` 下**可讀**（`USING` 允許），但**不可寫**（`WITH CHECK` 不含 `department_id IS NULL` → 寫入被擋）。
3. `metadicts` 部門列在 `scope=department` 且部門相符時可寫。

```go
//go:build integration

package services

import (
	"database/sql"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

func TestIntegrationRLSMetadictAuditIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)
	appDSN := testsupport.AppRoleDSN(t, adminDSN)

	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		t.Fatalf("admin 連線: %v", err)
	}
	defer admin.Close()

	var coA, coB int
	if err := admin.QueryRow(`INSERT INTO companies (name, identifier, status, created_at, updated_at)
		VALUES ('A','MD-A','active',now(),now()) RETURNING id`).Scan(&coA); err != nil {
		t.Fatalf("建 A: %v", err)
	}
	if err := admin.QueryRow(`INSERT INTO companies (name, identifier, status, created_at, updated_at)
		VALUES ('B','MD-B','active',now(),now()) RETURNING id`).Scan(&coB); err != nil {
		t.Fatalf("建 B: %v", err)
	}
	for _, co := range []int{coA, coB} {
		if _, err := admin.Exec(`INSERT INTO audit_logs (company_id, action, resource, created_at)
			VALUES ($1,'create','customer',now())`, co); err != nil {
			t.Fatalf("建稽核(公司 %d): %v", co, err)
		}
	}

	app, err := sql.Open("pgx", appDSN)
	if err != nil {
		t.Fatalf("app 連線: %v", err)
	}
	defer app.Close()

	var none int
	if err := app.QueryRow(`SELECT count(*) FROM audit_logs`).Scan(&none); err != nil {
		t.Fatalf("未設 scope 查稽核: %v", err)
	}
	if none != 0 {
		t.Fatalf("未設 scope 時稽核必須 0 列,got %d", none)
	}

	tx, err := app.Begin()
	if err != nil {
		t.Fatalf("開交易: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`SET LOCAL app.current_data_scope = 'company'`); err != nil {
		t.Fatalf("SET scope: %v", err)
	}
	if _, err := tx.Exec(`SET LOCAL app.current_company_id = '` + itoa(coA) + `'`); err != nil {
		t.Fatalf("SET company: %v", err)
	}
	var distinct int
	if err := tx.QueryRow(`SELECT count(DISTINCT company_id) FROM audit_logs`).Scan(&distinct); err != nil {
		t.Fatalf("查稽核: %v", err)
	}
	if distinct != 1 {
		t.Fatalf("稽核應只看到 1 家公司,got %d", distinct)
	}
	// 系統預設字典（department_id IS NULL）：可讀、不可寫。
	var readable int
	if err := tx.QueryRow(`SELECT count(*) FROM metadicts WHERE department_id IS NULL`).Scan(&readable); err != nil {
		t.Fatalf("讀系統預設字典: %v", err)
	}
	if readable == 0 {
		t.Fatal("系統預設字典在 company scope 下應可讀")
	}
	if _, err := tx.Exec(`INSERT INTO metadicts (type, code, display_name, department_id, sort_order, is_active)
		VALUES ('unit','X1','別名',NULL,0,true)`); err == nil {
		t.Fatal("以公司身分寫入系統預設字典必須被擋（WITH CHECK 不含 department_id IS NULL）")
	}
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && task test:integration -- -run TestIntegrationRLSMetadictAuditIsolation -v`
Expected: FAIL —未設 scope 時稽核仍可見

- [ ] **Step 3: 遷移兩檔路徑**

```bash
cd backend
sd 's\.db\.' 'dbtenant.Client(ctx, s.db).' internal/services/metadict_service.go
sd 's\.db\.' 'dbtenant.Client(ctx, s.db).' internal/services/audit_service.go
grep -n 's\.db\.\|db.Tx(ctx)' internal/services/metadict_service.go internal/services/audit_service.go   # 期望無輸出
```
三處自開交易依 Global Constraints「交易改寫樣式」改寫為使用請求交易；`internal/audit/recorder.go` 確認以呼叫端傳入的 `*ent.Tx` 寫入（同事務稽核 D18），非自開交易。

- [ ] **Step 4: ENABLE migration（`database/migrations/00027_rls_enable_metadicts_audit.sql`）**

```sql
-- 字典與稽核啟用 RLS（D36）。
-- +goose Up
ALTER TABLE metadicts   ENABLE ROW LEVEL SECURITY;
ALTER TABLE metadicts   FORCE  ROW LEVEL SECURITY;
ALTER TABLE audit_logs  ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs  FORCE  ROW LEVEL SECURITY;

-- +goose Down
ALTER TABLE metadicts   NO FORCE ROW LEVEL SECURITY;
ALTER TABLE metadicts   DISABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs  NO FORCE ROW LEVEL SECURITY;
ALTER TABLE audit_logs  DISABLE ROW LEVEL SECURITY;
```

- [ ] **Step 5: 跑測試**

Run: `cd backend && task test:integration -- -run 'TestIntegrationRLSMetadictAudit|TestIntegrationListPageScan' -v`
Expected: 全 PASS

- [ ] **Step 6: Commit**

```bash
git add backend/database/migrations/00027_rls_enable_metadicts_audit.sql backend/internal/services/metadict_service.go backend/internal/services/audit_service.go backend/internal/services/rls_metadict_audit_integration_test.go
git commit -m "feat(backend): 字典與稽核啟用 RLS 並收斂路徑（00027）"
```

---

### Task 9: 核心域（companies／departments／users／roles／role_permissions）＋ 認證系統範圍

**Files:**
- Create: `database/migrations/00028_rls_enable_core.sql`
- Create: `internal/services/rls_core_integration_test.go`
- Create: `internal/server/rls_core_integration_test.go`
- Modify: `internal/services/company_service.go`（19 處、3 處）
- Modify: `internal/services/user_service.go`（15 處、5 處）
- Modify: `internal/services/role_service.go`（6 處、1 處）
- Modify: `internal/server/server.go:167-210`（authzMiddleware 的身分查詢改 `SystemScopeTx`）
- Modify: `internal/handlers/auth_handler.go`（登入憑證查詢改 `SystemScopeTx`）

**Interfaces:**
- Consumes: `dbtenant.SystemScopeTx`、`dbtenant.Client`
- Produces: 核心五表 ENABLE + FORCE；未登入路徑以系統範圍查詢

**為何核心域最後：** 幾乎所有服務都會讀 `users`/`companies`/`departments`，先啟用會讓尚未遷移的路徑全數黑屏；此任務完成後全站才進入「RLS 全網生效」狀態。

- [ ] **Step 1: 寫失敗測試（`internal/services/rls_core_integration_test.go`）**

```go
//go:build integration

package services

import (
	"database/sql"
	"testing"

	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRLSCoreIsolation 驗證核心五表的隔離：
// users 以 company_users 表達所屬公司、departments 以 company_departments 表達、
// self 範圍只能看自己那一列。
func TestIntegrationRLSCoreIsolation(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)
	appDSN := testsupport.AppRoleDSN(t, adminDSN)

	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		t.Fatalf("admin 連線: %v", err)
	}
	defer admin.Close()

	var coA, coB, userA int
	if err := admin.QueryRow(`INSERT INTO companies (name, identifier, status, created_at, updated_at)
		VALUES ('A','CORE-A','active',now(),now()) RETURNING id`).Scan(&coA); err != nil {
		t.Fatalf("建 A: %v", err)
	}
	if err := admin.QueryRow(`INSERT INTO companies (name, identifier, status, created_at, updated_at)
		VALUES ('B','CORE-B','active',now(),now()) RETURNING id`).Scan(&coB); err != nil {
		t.Fatalf("建 B: %v", err)
	}
	if err := admin.QueryRow(`INSERT INTO users (email, name, status, role, password_hash, company_users, created_at, updated_at)
		VALUES ('a@example.com','A','active','staff','x',$1,now(),now()) RETURNING id`, coA).Scan(&userA); err != nil {
		t.Fatalf("建 A 使用者: %v", err)
	}
	if _, err := admin.Exec(`INSERT INTO users (email, name, status, role, password_hash, company_users, created_at, updated_at)
		VALUES ('b@example.com','B','active','staff','x',$1,now(),now())`, coB); err != nil {
		t.Fatalf("建 B 使用者: %v", err)
	}

	app, err := sql.Open("pgx", appDSN)
	if err != nil {
		t.Fatalf("app 連線: %v", err)
	}
	defer app.Close()

	t.Run("未設 scope → fail-closed", func(t *testing.T) {
		var n int
		if err := app.QueryRow(`SELECT count(*) FROM users`).Scan(&n); err != nil {
			t.Fatalf("查詢: %v", err)
		}
		if n != 0 {
			t.Fatalf("未設 scope 時 users 必須 0 列,got %d", n)
		}
	})

	t.Run("scope=company A → 只見同公司", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		if _, err := tx.Exec(`SET LOCAL app.current_data_scope = 'company'`); err != nil {
			t.Fatalf("SET scope: %v", err)
		}
		if _, err := tx.Exec(`SET LOCAL app.current_company_id = '` + itoa(coA) + `'`); err != nil {
			t.Fatalf("SET company: %v", err)
		}
		var emails int
		if err := tx.QueryRow(`SELECT count(*) FROM users WHERE email = 'b@example.com'`).Scan(&emails); err != nil {
			t.Fatalf("查詢: %v", err)
		}
		if emails != 0 {
			t.Fatal("company scope 下不得看到別家公司的使用者")
		}
		var mine int
		if err := tx.QueryRow(`SELECT count(*) FROM users WHERE id = $1`, userA).Scan(&mine); err != nil {
			t.Fatalf("查詢: %v", err)
		}
		if mine != 1 {
			t.Fatalf("應看到自己公司的使用者,got %d", mine)
		}
	})

	t.Run("跨公司寫入 → 被 WITH CHECK 擋", func(t *testing.T) {
		tx, err := app.Begin()
		if err != nil {
			t.Fatalf("開交易: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		if _, err := tx.Exec(`SET LOCAL app.current_data_scope = 'company'`); err != nil {
			t.Fatalf("SET scope: %v", err)
		}
		if _, err := tx.Exec(`SET LOCAL app.current_company_id = '` + itoa(coA) + `'`); err != nil {
			t.Fatalf("SET company: %v", err)
		}
		if _, err := tx.Exec(`INSERT INTO users (email, name, status, role, password_hash, company_users, created_at, updated_at)
			VALUES ('x@example.com','X','active','staff','x',$1,now(),now())`, coB); err == nil {
			t.Fatal("以 A 的身分新增 B 公司的使用者必須被擋")
		}
	})
}
```

- [ ] **Step 2: 跑測試確認失敗**

Run: `cd backend && task test:integration -- -run TestIntegrationRLSCoreIsolation -v`
Expected: FAIL —未設 scope 時 users 仍可見

- [ ] **Step 3: 寫「未登入查詢需系統範圍」的測試（`internal/server/rls_core_integration_test.go`）**

契約：核心表 ENABLE 後，**無 scope 的查詢必須看不到任何列**（fail-closed）——所以登入與 `authzMiddleware` 的身分解析若沒包系統範圍，全站將無法登入。同時 `dbtenant.SystemScopeTx` 必須看得到（owner 也被 `FORCE` 擋，故 seed／維運同樣得走它）。

```go
//go:build integration

package server

import (
	"database/sql"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/pressly/goose/v3"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/ent/user"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// rlsCoreMigrationsDir 與 cmd/migrate 同路徑（go test 以套件目錄為 cwd）。
const rlsCoreMigrationsDir = "../../database/migrations"

// migrateForRLSCoreTest 以 cmd/migrate 相同路徑套用全部遷移（含 00024~00028 的 ENABLE）。
func migrateForRLSCoreTest(t *testing.T, dsn string) {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("設定 dialect: %v", err)
	}
	if err := goose.RunContext(t.Context(), "up", db, rlsCoreMigrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}
}

// openEntForRLSCoreTest 開 ent client（遷移已由 goose 套用，不跑 ent auto-migrate）。
func openEntForRLSCoreTest(t *testing.T, dsn string) *ent.Client {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// TestIntegrationAuthLookupNeedsSystemScope 驗證：
// ① 無 scope → users 不可見（fail-closed）；
// ② SystemScopeTx → 可見（未登入路徑的唯一合法入口）。
func TestIntegrationAuthLookupNeedsSystemScope(t *testing.T) {
	testsupport.RequiresContainer(t)
	dsn := testsupport.Postgres(t)
	migrateForRLSCoreTest(t, dsn)
	client := openEntForRLSCoreTest(t, dsn)
	ctx := t.Context()

	if err := dbtenant.SystemScopeTx(ctx, client, func(tx *ent.Tx) error {
		co, err := tx.Client().Company.Create().
			SetName("登入測試公司").SetIdentifier("RLS-LOGIN").Save(ctx)
		if err != nil {
			return err
		}
		_, err = tx.Client().User.Create().
			SetEmail("login@example.com").SetName("登入者").SetStatus("active").
			SetRole("staff").SetPasswordHash("x").SetCompanyID(co.ID).Save(ctx)
		return err
	}); err != nil {
		t.Fatalf("系統範圍建立資料: %v", err)
	}

	visible, err := dbtenant.Client(ctx, client).User.Query().
		Where(user.EmailEQ("login@example.com")).Exist(ctx)
	if err != nil {
		t.Fatalf("無 scope 查詢: %v", err)
	}
	if visible {
		t.Fatal("無 scope 時 users 必須不可見（fail-closed）")
	}

	var found bool
	if err := dbtenant.SystemScopeTx(ctx, client, func(tx *ent.Tx) error {
		var err error
		found, err = tx.Client().User.Query().Where(user.EmailEQ("login@example.com")).Exist(ctx)
		return err
	}); err != nil {
		t.Fatalf("系統範圍查詢: %v", err)
	}
	if !found {
		t.Fatal("系統範圍應看得到剛建立的使用者")
	}
}
```

- [ ] **Step 4: 跑測試確認失敗**

Run: `cd backend && task test:integration -- -run TestIntegrationLoginLookupUnderRLS -v`
Expected: FAIL —系統範圍交易尚未實作於呼叫路徑（`SystemScopeTx` 於 Task 4 已存在，此處失敗點應為 `migrateForServerTest` 尚未實作 → 先補輔助函式再跑）

- [ ] **Step 5: 遷移三檔路徑與未登入查詢**

```bash
cd backend
for f in company_service user_service role_service; do
  sd 's\.db\.' 'dbtenant.Client(ctx, s.db).' internal/services/$f.go
done
grep -n 's\.db\.\|db.Tx(ctx)' internal/services/company_service.go internal/services/user_service.go internal/services/role_service.go   # 期望無輸出
```
- 三檔的自開交易（3／5／1 處）依 Global Constraints「交易改寫樣式」改寫為使用請求交易。
- `internal/server/server.go:167-210` 的 `authzMiddleware`：把查 `users`／`companies`／`departments`／`roles` 的段落以 `dbtenant.SystemScopeTx` 包起來（該階段尚無身分，scope 由資料推導）。**只包身分解析所需查詢**，不得把整個請求處理搬進系統範圍。
- `internal/handlers/auth_handler.go`：登入（email 查找 + 密碼驗證）、QR 兌換前的身分查詢比照辦理。

- [ ] **Step 6: ENABLE migration（`database/migrations/00028_rls_enable_core.sql`）**

```sql
-- 核心域啟用 RLS（D36）。這是最後一批：核心表被幾乎所有服務讀取，需待各域路徑
-- 都已走請求層租戶交易後才啟用。
-- +goose Up
ALTER TABLE companies         ENABLE ROW LEVEL SECURITY;
ALTER TABLE companies         FORCE  ROW LEVEL SECURITY;
ALTER TABLE departments       ENABLE ROW LEVEL SECURITY;
ALTER TABLE departments       FORCE  ROW LEVEL SECURITY;
ALTER TABLE users             ENABLE ROW LEVEL SECURITY;
ALTER TABLE users             FORCE  ROW LEVEL SECURITY;
ALTER TABLE roles             ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles             FORCE  ROW LEVEL SECURITY;
ALTER TABLE role_permissions  ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_permissions  FORCE  ROW LEVEL SECURITY;

-- +goose Down
ALTER TABLE companies         NO FORCE ROW LEVEL SECURITY;
ALTER TABLE companies         DISABLE ROW LEVEL SECURITY;
ALTER TABLE departments       NO FORCE ROW LEVEL SECURITY;
ALTER TABLE departments       DISABLE ROW LEVEL SECURITY;
ALTER TABLE users             NO FORCE ROW LEVEL SECURITY;
ALTER TABLE users             DISABLE ROW LEVEL SECURITY;
ALTER TABLE roles             NO FORCE ROW LEVEL SECURITY;
ALTER TABLE roles             DISABLE ROW LEVEL SECURITY;
ALTER TABLE role_permissions  NO FORCE ROW LEVEL SECURITY;
ALTER TABLE role_permissions  DISABLE ROW LEVEL SECURITY;
```

- [ ] **Step 7: 全套守門（此刻全網生效）**

Run: `cd backend && task check && task test:integration`
Expected: 全綠。任何「查不到資料」的失敗都代表該路徑未走 `dbtenant.Client`／未以系統範圍包裝 → 修該路徑，不得放寬為 `DISABLE RLS`

- [ ] **Step 8: Commit**

```bash
git add backend/database/migrations/00028_rls_enable_core.sql backend/internal/services/company_service.go backend/internal/services/user_service.go backend/internal/services/role_service.go backend/internal/server backend/internal/handlers backend/internal/services/rls_core_integration_test.go
git commit -m "feat(backend): 核心域啟用 RLS，未登入路徑改系統範圍交易（00028）"
```

---

### Task 10: 系統範圍收斂、seed、端到端跨租戶探針與慣例文件

**Files:**
- Modify: `cmd/seed/main.go`（交易內 `SET LOCAL … 'all'`）
- Modify: `README.md`／`backend/Taskfile.yml`（本機啟動步驟加 `task db:app-password`）
- Modify: `backend/AGENTS.md`（新增 RLS 慣例小節）
- Create: `internal/services/rls_cross_tenant_integration_test.go`

**Interfaces:**
- Consumes: 前九個任務的全部產物
- Produces: 端到端探針（每個租戶端點都驗一次跨租戶不可見）、RLS 慣例成文

- [ ] **Step 1: 寫端到端探針（`internal/services/rls_cross_tenant_integration_test.go`）**

探針的目的：**「程式漏掛 `dbtenant.Client` 或 interceptor」在單元測試看不出來，只有這條會紅**（RLS 已 ENABLE，未套 scope 的查詢會 fail-closed）。

先改既有測試 helper（`internal/services/list_pagination_integration_test.go`）——把 `newListScanServer` 改成委派版，並抽出可指定 RLS scope 的版本（既有呼叫端不受影響）：

```go
func newListScanServer(t *testing.T, db *ent.Client, companyID int) listScanClients {
	t.Helper()
	return newListScanServerWithScope(t, db, auth.RLSScope{
		UserID: "1", CompanyID: strconv.Itoa(companyID),
		DataScope: auth.DataScopeAll, CompanyActive: true,
	})
}

// newListScanServerWithScope 與原 newListScanServer 同構，只多注入 RLS scope；
// 請求層租戶交易由 Register*Services 內的 interceptor 依 ctx 的 scope 套用。
func newListScanServerWithScope(t *testing.T, db *ent.Client, scope auth.RLSScope) listScanClients {
	t.Helper()
	super := authz.Identity{
		UserID: scope.UserID, CompanyID: scope.CompanyID, Role: "super", Roles: []string{"super"},
	}
	mux := http.NewServeMux()
	RegisterCompanyServices(mux, db)
	RegisterAuditServices(mux, db)
	RegisterRoleServices(mux, db)
	RegisterCustomerServices(mux, db, "http://localhost:3000")
	RegisterProcessingSpecService(mux, db)
	RegisterProductCategoryService(mux, db)
	RegisterRouteService(mux, db)
	RegisterMetadictServices(mux, db)
	RegisterProductService(mux, db)
	RegisterWarehouseService(mux, db)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := authz.WithIdentity(r.Context(), super)
		ctx = auth.WithRLS(ctx, scope)
		ctx = authz.WithCASLEnabled(ctx, true)
		ctx = authz.WithDB(ctx, db)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return listScanClients{
		companies:   salesorderv1connect.NewCompanyServiceClient(http.DefaultClient, ts.URL),
		departments: salesorderv1connect.NewDepartmentServiceClient(http.DefaultClient, ts.URL),
		roles:       salesorderv1connect.NewRoleServiceClient(http.DefaultClient, ts.URL),
		customers:   customersv1connect.NewCustomerServiceClient(http.DefaultClient, ts.URL),
		specs:       mastersv1connect.NewProcessingSpecServiceClient(http.DefaultClient, ts.URL),
		cats:        mastersv1connect.NewProductCategoryServiceClient(http.DefaultClient, ts.URL),
		routes:      mastersv1connect.NewRouteServiceClient(http.DefaultClient, ts.URL),
		metadicts:   metadictv1connect.NewMetadictServiceClient(http.DefaultClient, ts.URL),
		products:    productsv1connect.NewProductServiceClient(http.DefaultClient, ts.URL),
		warehouses:  mastersv1connect.NewWarehouseServiceClient(http.DefaultClient, ts.URL),
		audits:      auditv1connect.NewAuditServiceClient(http.DefaultClient, ts.URL),
	}
}
```

新增探針檔：

```go
//go:build integration

package services

import (
	"context"
	"strconv"
	"testing"

	"connectrpc.com/connect"

	"github.com/salesorder/sales-order-1.0/backend/ent"
	"github.com/salesorder/sales-order-1.0/backend/internal/auth"
	"github.com/salesorder/sales-order-1.0/backend/internal/dbtenant"
	customersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/customers/v1"
	mastersv1 "github.com/salesorder/sales-order-1.0/backend/internal/proto/masters/v1"
	"github.com/salesorder/sales-order-1.0/backend/internal/testsupport"
)

// TestIntegrationRLSCrossTenantEndpoints 以 A 公司身分呼叫各租戶端點，
// 回應中不得出現 B 公司的任何列（RLS 為最後一道防線，服務層過濾失效時仍須如此）。
func TestIntegrationRLSCrossTenantEndpoints(t *testing.T) {
	testsupport.RequiresContainer(t)
	adminDSN := testsupport.Postgres(t)
	migrateBusinessUp(t, adminDSN)
	_, adminClient := openPGEntClientFromGoose(t, adminDSN)
	ctx := t.Context()

	// fixture 用 admin（superuser）建立即可：superuser 恆繞過 RLS，不受 WITH CHECK 影響。
	coA, coB := seedRLSCrossTenant(t, ctx, adminClient)
	if coA == coB {
		t.Fatal("兩家公司的 id 不得相同")
	}

	// ⚠️ 受測的 server **必須**用 app_rw 連線建立 —— 這是本探針唯一有意義的設定：
	// 測試容器的 postgres 是 superuser，而 superuser 恆繞過 RLS（FORCE 亦然），
	// 用它建的 server 不管路徑有沒有漏掛都會全綠（T5 實測）。
	// 且必須經 dbtenant.NewClient：SET LOCAL 是 driver 裝飾器在 Tx(ctx) 內套的。
	appSQL, err := sql.Open("pgx", testsupport.AppRoleDSN(t, adminDSN))
	if err != nil {
		t.Fatalf("app_rw 連線: %v", err)
	}
	t.Cleanup(func() { _ = appSQL.Close() })
	appClient := dbtenant.NewClient(appSQL)
	t.Cleanup(func() { _ = appClient.Close() })

	clientsA := newListScanServerWithScope(t, appClient, auth.RLSScope{
		UserID: "1", CompanyID: strconv.Itoa(coA),
		DataScope: auth.DataScopeCompany, CompanyActive: true,
	})

	cust, err := clientsA.customers.ListCustomers(ctx, connect.NewRequest(
		&customersv1.ListCustomersRequest{PageSize: 50},
	))
	if err != nil {
		t.Fatalf("ListCustomers: %v", err)
	}
	if len(cust.Msg.Customers) != 1 {
		t.Fatalf("A 公司應只看到 1 筆客戶（自家）,got %d", len(cust.Msg.Customers))
	}
	if got := cust.Msg.Customers[0].CompanyId; got != int64(coA) {
		t.Fatalf("回應含非本公司資料:company_id=%d（應為 %d）", got, coA)
	}

	wh, err := clientsA.warehouses.ListWarehouses(ctx, connect.NewRequest(
		&mastersv1.ListWarehousesRequest{PageSize: 50},
	))
	if err != nil {
		t.Fatalf("ListWarehouses: %v", err)
	}
	if len(wh.Msg.Warehouses) != 1 {
		t.Fatalf("A 公司應只看到 1 筆倉別（自家）,got %d", len(wh.Msg.Warehouses))
	}
	if got := wh.Msg.Warehouses[0].CompanyId; got != int64(coA) {
		t.Fatalf("回應含非本公司資料:company_id=%d（應為 %d）", got, coA)
	}
}

// seedRLSCrossTenant 以系統範圍（scope=all）建立兩家公司的客戶與倉別各一筆：
// ENABLE + FORCE 後連 owner 連線都受 RLS 約束，維運路徑必須走 SystemScopeTx。
func seedRLSCrossTenant(t *testing.T, ctx context.Context, db *ent.Client) (int, int) {
	t.Helper()
	var coA, coB int
	if err := dbtenant.SystemScopeTx(ctx, db, func(tx *ent.Tx) error {
		a, err := tx.Client().Company.Create().SetName("跨租戶 A").SetIdentifier("X-A").Save(ctx)
		if err != nil {
			return err
		}
		b, err := tx.Client().Company.Create().SetName("跨租戶 B").SetIdentifier("X-B").Save(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Client().Customer.Create().
			SetCompanyID(a.ID).SetCustomerCode("C-A").SetName("A 客戶").Save(ctx); err != nil {
			return err
		}
		if _, err := tx.Client().Customer.Create().
			SetCompanyID(b.ID).SetCustomerCode("C-B").SetName("B 客戶").Save(ctx); err != nil {
			return err
		}
		if _, err := tx.Client().Warehouse.Create().
			SetCompanyID(a.ID).SetCode("W-A").SetName("A 倉").Save(ctx); err != nil {
			return err
		}
		if _, err := tx.Client().Warehouse.Create().
			SetCompanyID(b.ID).SetCode("W-B").SetName("B 倉").Save(ctx); err != nil {
			return err
		}
		coA, coB = a.ID, b.ID
		return nil
	}); err != nil {
		t.Fatalf("系統範圍造資料: %v", err)
	}
	return coA, coB
}
```

- [ ] **Step 2: 跑探針，確認「現況全綠」**

Run: `cd backend && task test:integration -- -run TestIntegrationRLSCrossTenantEndpoints -v`
Expected: PASS。若 FAIL 且看到別家公司的列 → 該端點路徑未走 `dbtenant.Client`，修該處後重跑

- [ ] **Step 3: `cmd/seed` 改為系統範圍交易**

```go
cfg := config.New()
client, err := database.OpenEnt(cfg.Database.AdminDSN())
if err != nil {
	log.Fatalf("開啟 ent client: %v", err)
}
ctx := context.Background()
// FORCE RLS 也會約束 owner：seed 屬系統級維運，於明確的系統範圍內執行（scope=all）。
if err := dbtenant.SystemScopeTx(ctx, client, func(tx *ent.Tx) error {
	if err := SeedBuiltinRoles(ctx, tx.Client()); err != nil {
		return err
	}
	if err := SeedBuiltinRolePermissions(ctx, tx.Client()); err != nil {
		return err
	}
	return SeedDeveloper(ctx, tx.Client(), cfg.API.Env, firstCompanyID(ctx, tx.Client()))
}); err != nil {
	log.Fatalf("seed: %v", err)
}
log.Println("seed: 完成")
```

- [ ] **Step 4: 驗證 seed 在 RLS 生效後仍可跑**

Run: `cd backend && task test:integration -- -run TestIntegrationSeed -v`；若無此測試則以手動路徑驗證：
`task infra:start && task migrate:up && task seed`
Expected: `seed: 完成`（無 RLS 攔阻錯誤）

- [ ] **Step 5: `backend/AGENTS.md` 新增慣例小節**

```markdown
## 9. RLS 與租戶交易（D36，2026-09-20 起）

1. **業務連線是非 owner 的 `app_rw`**；goose／seed／OpenFGA／平台域用 `DATABASE_ADMIN_URL`。
2. **所有服務層 DB 存取走 `dbtenant.Client(ctx, s.db)`**；交易邊界由請求層的 interceptor 持有，服務內不得再 `db.Tx(ctx)`。
3. **串流 RPC／PDF 產出不得沿用請求層交易**（長交易壓連線池），需另立短交易邊界。
4. **對已 ENABLE（FORCE）的表做資料回填**的 migration 與 seed，必須在交易內先 `SET LOCAL app.current_data_scope = 'all'`（FORCE 也擋 owner）。
5. **RLS 不得用 sqlite（enttest）測**：`SET`／`FORCE` 在 sqlite 無效，必為 `//go:build integration` + `internal/testsupport`。
6. **未登入的查詢**（登入憑證、`authzMiddleware` 身分解析）走 `dbtenant.SystemScopeTx`，且僅限該查詢，不得把整個請求搬進系統範圍。
7. policy 一律 `FOR ALL … USING … WITH CHECK …`；新增租戶表時**必須同時**新增 policy 並在啟用 migration 中 `ENABLE` + `FORCE`。
```

- [ ] **Step 6: 全套守門（最終驗收）**

Run: `cd backend && task check && task test:integration && go build ./cmd/...`
Expected: 全綠

- [ ] **Step 7: Commit**

```bash
git add backend/cmd/seed backend/internal/services/rls_cross_tenant_integration_test.go backend/AGENTS.md README.md backend/Taskfile.yml
git commit -m "feat(backend): 系統範圍收斂、seed 走 admin+系統 scope、跨租戶探針與 RLS 慣例"
```

---

## 驗收對照（spec §6／§7）

| spec 要求 | 對應任務 |
|---|---|
| 非 owner 應用角色、admin DSN 分流（§6.1） | Task 1、2 |
| 14 個 policy 補 `WITH CHECK`、3 張漏網表補 policy（§6.2） | Task 3 |
| 業務表 `ENABLE` + `FORCE`（§6.2） | Task 5–9（依域分批） |
| 請求層租戶交易、含唯讀（§6.3） | Task 4（機制）＋ Task 5–9（各域遷移） |
| 未登入查詢走系統範圍（§6.3-4、§6.4） | Task 9 |
| 回填需系統 scope（§6.2-3） | Task 10 |
| 未設 GUC → 0 列、跨租戶寫入被擋（§7.2-1、-2） | Task 3（結構）＋ Task 5–9（行為） |
| `migrate up/down` 完整回滾（§6.5） | Task 3 Step 5、各 ENABLE migration 的 `Down` |
| 慣例成文（§6.3） | Task 10 Step 5 |

## 風險與對策（本計畫新增者）

| # | 風險 | 對策 |
|---|---|---|
| P1 | 遷移期間漏掉某路徑 → 該端點黑屏（功能故障） | 每域以既有整合測試＋跨租戶探針守門；核心域最後啟用，把黑屏風險集中在最後一波並由全套測試暴露 |
| P2 | `sd` 批次替換誤傷（例如區域變數也叫 `s.db`） | 每檔替換後以 `grep -n 's\.db\.' <file>` 必須為 0，且 `go build ./...` 通過 |
| P3 | 請求層長交易壓連線池 | 列入 `backend/AGENTS.md` 慣例；串流 AI／PDF 另立短交易邊界（未來任務） |
| P4 | `db.Tx(ctx)` 移除後，服務層原本的「部分失敗回滾」語意變成「整請求回滾」 | 這是刻意的語意收斂（D18 同事務稽核亦要求）；以既有整合測試確認無回歸 |

---

*建立：2026-09-20（SaaS 化 spec 的第一份實作計畫：RLS 租戶隔離啟用）*
