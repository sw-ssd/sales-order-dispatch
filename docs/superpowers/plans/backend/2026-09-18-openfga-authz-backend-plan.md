# 後端 OpenFGA 授權對齊與地基補齊 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把後端授權引擎由 Casbin + CASL 遷移至內嵌 OpenFGA + RLS（D32），並補齊核心表 migration / 角色 seed / Argon2id，使後端可端到端運作且授權決策由 OpenFGA 驅動。

**Architecture:** 在後端程序內以 OpenFGA Go library 內嵌一個實例（共用 PostgreSQL datastore、單一 store + authorization model）。新 `internal/authz/openfga` 提供 `Check`/`ListObjects` 供 Connect middleware 使用；`role_permissions` 保留為角色權限定義來源，異動時 translate 成 OpenFGA tuples；RLS（data_scope）保留為資料庫兜底。移除 Casbin 與 CASL engine 與依賴。`GetAbility` 改為 OpenFGA 驅動。

**Tech Stack:** Go 1.25.7、OpenFGA（github.com/openfga/openfga + datastore）、ent、connectrpc、PostgreSQL（pgx）、envconfig、goose、golang.org/x/crypto/argon2。

**Spec:** `docs/superpowers/specs/2026-09-18-openfga-authz-design.md`

## Global Constraints

- 繁體中文註解與 commit message（專案慣例）。
- 設定密鑰 production 空值/預設值 → `Server.Init()` fail-fast；驗證端對空密鑰 fail-closed。
- 測試只用 stdlib `testing` + enttest（sqlite）；禁止 testify/testcontainers 新測試依賴。OpenFGA 相關測試用 openfga 記憶體 datastore + 真實 Postgres 整合測試（D21）。
- 任一新授權門檻：未登入 → `Unauthenticated`、越權 → `PermissionDenied`；跨租戶 → `PermissionDenied`；輸入驗證失敗 → `InvalidArgument`。
- `development` 預設 developer 逃生門；`production` 誤開 → 拒絕啟動。
- 提交前必跑 `task check`（fmt+vet+lint+test）。

---

### Task 1: 核心表 migration + 7 內建角色 seed + developer 帳號

**Files:**
- Create: `backend/database/migrations/00005_core_schema.sql`（roles / users / companies / departments；role_permissions 補 FK）
- Create: `backend/database/migrations/00006_seed_roles.sql`（7 內建角色 + 角色權限 + dev developer 帳號）
- Modify: `backend/cmd/seed/main.go`（改為冪等執行角色/權限 seed）
- Test: `backend/cmd/seed/seed_test.go`

**Interfaces:**
- Produces: DB 具備 `roles`（7 內建 + `is_system` + `data_scope`）、`role_permissions`、`companies`、`departments`、`users` 表；`developer` 帳號僅 `ENV=development`。
- Consumes: 既有 `ent/schema/{role,user,company,department,rolepermission}` 定義。

- [ ] **Step 1: 寫 migration 00005（core schema，對齊 ent schema）**

`backend/database/migrations/00005_core_schema.sql`（Up：建表，Down：DROP）。欄位對齊 `ent/schema/`（見 role.go/user.go/company.go/department.go/rolepermission.go）。role_permissions 將 `role_id bigint` 改為 `REFERENCES roles(id)`；加入 `created_at`/`updated_at`/`deleted_at`（依各 ent schema）。

- [ ] **Step 2: 寫 migration 00006（seed roles/permissions）**

`backend/database/migrations/00006_seed_roles.sql`：UPSERT 7 內建角色（super/all、company_admin/company、dept_admin/department、staff/department、customer/self、guest/self、developer/all，`is_system=true`）。依 `ent/schema/rolepermission.go` 與既有 `internal/services/role_service.go` 的預設權限建立 `role_permissions` 種子（resource×action×conditions）。開發者帳號以 `-- DO NOT seed in production` 註記並用 `INSERT ... WHERE current_setting('salesorder.env') = 'development'` 防護。

- [ ] **Step 3: 改造 seed 入口為冪等**

`backend/cmd/seed/main.go`：連 DB → 執行 00006 的 seed 邏輯（或直接呼叫 goose up；若 seed 走 migration 則 main.go 僅做連線檢查並 log「seed 已由 migration 00006 提供」）。維持「冪等」。

- [ ] **Step 4: 寫並跑 seed/migration 測試**

`backend/cmd/seed/seed_test.go`：以 enttest sqlite 驗證 7 內建角色存在、`data_scope` 正確、developer 僅 dev。Run: `go test ./cmd/seed/... -v`。Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add backend/database/migrations/00005_core_schema.sql backend/database/migrations/00006_seed_roles.sql backend/cmd/seed/main.go backend/cmd/seed/seed_test.go
git commit -m "feat(backend): core schema migration + 7 內建角色 seed + developer 帳號"
```

---

### Task 2: 密碼雜湊 bcrypt → Argon2id

**Files:**
- Modify: `backend/internal/auth/password.go`
- Test: `backend/internal/auth/password_test.go`

**Interfaces:**
- Produces: `HashPassword(plain string) (string, error)` 與 `VerifyPassword(hash, plain string) bool` 改用 Argon2id；簽名不變，呼叫端（auth_handler.go Login）不需改。
- Consumes: 既有呼叫端（不變）。

- [ ] **Step 1: 寫失敗測試**

`password_test.go`：`TestHashPasswordArgon2`、`TestVerifyPasswordArgon2`、`TestVerifyPasswordRejectsBcrypt`（舊 bcrypt hash 驗證失敗）。

- [ ] **Step 2: 跑測試確認失敗**

Run: `go test ./internal/auth/... -run Argon -v`。Expected: FAIL（目前 bcrypt）。

- [ ] **Step 3: 實作 Argon2id**

`password.go`：以 `golang.org/x/crypto/argon2` 實作（使用 argon2id，內嵌 salt + params 於 hash 字串，格式 `$argon2id$v=19$m=65536,t=1,p=4$<salt>$<key>`）。

- [ ] **Step 4: 跑測試確認通過**

Run: `go test ./internal/auth/... -v`。Expected: PASS；確認 bcrypt 舊 hash 被拒。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/auth/password.go backend/internal/auth/password_test.go
git commit -m "fix(backend): 密碼雜湊 bcrypt → Argon2id(對齊 identity-access 規格)"
```

---

### Task 3: 內嵌 OpenFGA 起實例 + config

**Files:**
- Create: `backend/config/openfga.go`、Modify: `backend/config/config.go`
- Create: `backend/third_party/openfga/openfga.go`
- Modify: `backend/.env.example`
- Test: `backend/third_party/openfga/openfga_test.go`

**Interfaces:**
- Produces: `config.OpenFGA` struct（`OPENFGA_ENABLED`、`OPENFGA_DATABASE_URL`、`OPENFGA_STORE_NAME`）；`openfga.NewClient(ctx, cfg, entDB)` → `*openfga.Client`（持 store id + server 實例）；`Client.Check(ctx, user, relation, object)`、`Client.ListObjects(...)`、`Client.WriteTuple/DeleteTuple(...)`。
- Consumes: `config.Config`；PostgreSQL 連線（與業務同庫，單一 store）。

- [ ] **Step 1: 建 config 檔並聚合**

`config/openfga.go`：`OpenFGA` struct。`config.go` 聚合欄位 + `mustProcess(&c.OpenFGA)`。

- [ ] **Step 2: 寫失敗測試（embed 起實例 + 建 store）**

`third_party/openfga/openfga_test.go`：以 openfga 記憶體 datastore 起內嵌 server，`CreateStore` → assert store ID 非空。

- [ ] **Step 3: 跑測試確認失敗**

Run: `go test ./third_party/openfga/... -v`。Expected: FAIL（openfga 未引入/未實作）。

- [ ] **Step 4: 引入依賴並實作**

Run: `cd backend && go get github.com/openfga/openfga@latest github.com/openfga/go-sdk@latest`。
`third_party/openfga/openfga.go`：以 `github.com/openfga/openfga/pkg/server` 用記憶體（測試）/ postgres（正式）datastore 起 `*server.Server`，`CreateStore`，將 `*server.Server` 包進 Client。註解說明 import/型別須對齊已解析的 openfga 版本（`server.New`、`openfgav1`、`typesystem`）。

- [ ] **Step 5: 跑測試確認通過 + 更新 .env.example**

Run: `go test ./third_party/openfga/... -v`。Expected: PASS。

- [ ] **Step 6: Commit**

```bash
git add backend/config/openfga.go backend/config/config.go backend/third_party/openfga/openfga.go backend/third_party/openfga/openfga_test.go backend/.env.example backend/go.mod backend/go.sum
git commit -m "feat(backend): 內嵌 OpenFGA 起實例 + config(D32)"
```

---

### Task 4: 授權 model DSL（.fga）

**Files:**
- Create: `backend/third_party/openfga/model.fga`
- Modify: `backend/third_party/openfga/openfga.go`（載入 model）

**Interfaces:**
- Produces: `openfga.WriteModel(ctx)` 將 `model.fga` 寫入 OpenFGA；未來領域擴充此檔。
- Consumes: Task 3 的 Client。

- [ ] **Step 1: 寫 model.fga**

定義租戶型別與 relations：`type user`、`type company`、`type department`、`type role`（`relation member: [user]` 等）、核心資源 type（`role`、`company`、`department`、`ability`）帶 `can_read`/`can_write` relations + userset rewrite 綁 `role#member`（對齊授權 model 取向 1）。

- [ ] **Step 2: 寫失敗測試**

`openfga_test.go`：`TestWriteModel` → write model 後 `ReadAuthorizationModel` 回傳 type definitions 含核心資源。

- [ ] **Step 3: 跑測試確認失敗** → Run `go test ./third_party/openfga/... -run WriteModel -v`。Expected: FAIL。

- [ ] **Step 4: 實作 model 載入 + 寫入**

`openfga.go`：以 `typesystem`/`WriteAuthorizationModel` 將 `model.fga` 轉 `TypeDefinition` 寫入。

- [ ] **Step 5: 跑測試確認通過 + Commit**

```bash
git add backend/third_party/openfga/model.fga backend/third_party/openfga/openfga.go backend/third_party/openfga/openfga_test.go
git commit -m "feat(backend): OpenFGA authorization model DSL(取向1 每資源型別+userset)"
```

---

### Task 5: internal/authz/openfga 封裝（Check / ListObjects）+ 換掉 CASL facade

**Files:**
- Create: `backend/internal/authz/openfga/openfga.go`
- Modify: `backend/internal/authz/access.go`（CASL 執行面移除，改 OpenFGA）
- Test: `backend/internal/authz/openfga/openfga_test.go`

**Interfaces:**
- Produces: `Check(ctx, subject, relation, object) (bool, error)`、`ListObjects(ctx, subject, relation, type) ([]string, error)`、`WriteTuple/DeleteTuple`；`authz.AccessibleFilter`/`Can` 語意改由 OpenFGA 提供。
- Consumes: Task 3/4 Client；`authz.Identity`。

- [ ] **Step 1: 寫失敗測試**

`openfga_test.go`：`TestCheckAllow`、`TestCheckDeny`、`TestListObjects`（寫 tuple 後 check/list 符合預期）。

- [ ] **Step 2: 跑測試確認失敗** → Run `go test ./internal/authz/openfga/... -v`。Expected: FAIL。

- [ ] **Step 3: 實作封裝**

`internal/authz/openfga/openfga.go`：包 `*openfga.Client` 的 `Check`/`ListObjects`/`WriteTuple`/`DeleteTuple`。

- [ ] **Step 4: 改 access.go 移除 CASL 執行面**

`access.go`：移除 `AccessibleFilter`/`Can`（CASL evaluator 呼叫）；介面改由 OpenFGA `Check` 提供；`Registry`/`FieldRegistry` 相關保留與否依 Task 6/8 決定（條件語意映射至 OpenFGA condition）。

- [ ] **Step 5: 跑測試確認通過** → Run `go test ./internal/authz/... -v`。Expected: PASS。

- [ ] **Step 6: Commit**

```bash
git add backend/internal/authz/openfga/openfga.go backend/internal/authz/openfga/openfga_test.go backend/internal/authz/access.go
git commit -m "feat(backend): authz/openfga Check/ListObjects 封裝;CASL 執行面改由 OpenFGA"
```

---

### Task 6: middleware 接線（identity → OpenFGA + RLS）

**Files:**
- Modify: `backend/internal/server/server.go`、`backend/internal/server/domains.go`
- Test: `backend/internal/server/server_test.go`

**Interfaces:**
- Consumes: Task 5 `authz/openfga`；`authz.Identity`。
- Produces: 受保護 RPC 路徑在進入 handler 前以 OpenFGA `Check` 判定；身分注入 ctx（OpenFGA subject + RLS scope）。

- [ ] **Step 1: 寫失敗測試**

`server_test.go`：`TestAuthzMiddlewareCheck`（有權→放行、無權→PermissionDenied、未登入→Unauthenticated）。

- [ ] **Step 2: 跑測試確認失敗** → Run `go test ./internal/server/... -v`。Expected: FAIL。

- [ ] **Step 3: 實作 middleware 接線**

`server.go` `authzMiddleware`：session/identity 就緒後，對受保護 RPC path 以 `authz/openfga.Check` 判定；developer 跳過 Check。`domains.go` 傳入 OpenFGA client。

- [ ] **Step 4: 跑測試確認通過** → Run `go test ./internal/server/... -v`。Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/server/server.go backend/internal/server/domains.go backend/internal/server/server_test.go
git commit -m "feat(backend): middleware 以 OpenFGA Check 做授權門檻(D32)"
```

---

### Task 7: RLS policy 落地核心表

**Files:**
- Modify: `backend/database/migrations/00002_rls_policies.sql`（補實際政策）或新增 `00007_rls_core_policies.sql`
- Test: `backend/internal/auth/rls_test.go`（延伸）

**Interfaces:**
- Consumes: `auth.RLSStatements`/`ApplyRLS`（已實作）。
- Produces: 核心表具 `FORCE ROW LEVEL SECURITY` 與依 `data_scope` 的政策。

- [ ] **Step 1: 寫政策 SQL（新增 00007）**

`00007_rls_core_policies.sql`：對 core 表 `ENABLE ROW LEVEL SECURITY` + `FORCE`；依 `app.current_data_scope` 寫 policy（all→USING(true)；company→company_id=app.current_company_id；department→department_id=...；self→id=app.current_user_id）。

- [ ] **Step 2: 延伸 rls_test**

`rls_test.go`：對各 data_scope 驗證 `RLSStatements` 產出含對應 `SET LOCAL app.current_*` 且順序決定性。

- [ ] **Step 3: 跑測試確認通過** → Run `go test ./internal/auth/... -v`。Expected: PASS。

- [ ] **Step 4: Commit**

```bash
git add backend/database/migrations/00007_rls_core_policies.sql backend/internal/auth/rls_test.go
git commit -m "feat(backend): RLS core 表政策落地(資料範圍 data_scope 兜底)"
```

---

### Task 8: role_permissions 異動 → OpenFGA tuples 同步

**Files:**
- Modify: `backend/internal/services/role_service.go`
- Test: `backend/internal/services/role_service_test.go`

**Interfaces:**
- Consumes: Task 5 `WriteTuple/DeleteTuple`。
- Produces: `UpdateRolePermissions` 成功後將對應角色/權限 translate 成 OpenFGA tuples 寫入（Write/Delete，對齊 D32「role_permissions 為定義來源，異動寫入 OpenFGA」）。

- [ ] **Step 1: 寫失敗測試**

`role_service_test.go`：`TestUpdateRolePermissionsSyncsOpenFGA`（更新後 assert 對應 tuple 存在）。

- [ ] **Step 2: 跑測試確認失敗** → Run `go test ./internal/services/... -run OpenFGA -v`。Expected: FAIL。

- [ ] **Step 3: 實作同步**

`role_service.go` `UpdateRolePermissions`：權限寫 `role_permissions` 後，逐條 translate（resource×action → relation + conditions）並 `WriteTuple`/`DeleteTuple`。

- [ ] **Step 4: 跑測試確認通過** → Run `go test ./internal/services/... -v`。Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/role_service.go backend/internal/services/role_service_test.go
git commit -m "feat(backend): role_permissions 異動同步 OpenFGA tuples(D32)"
```

---

### Task 9: 移除 Casbin/CASL code；GetAbility 改 OpenFGA 驅動

**Files:**
- Delete: `backend/internal/auth/casbin.go`、`backend/internal/auth/rbac.go`、`backend/internal/authz/casl/`、`backend/internal/domain/auth/ability.go`（CASL 載入）、舊測試
- Modify: `backend/internal/authz/access.go`、`backend/internal/domain/auth/ability.go`
- Modify: `backend/go.mod`（移除 casbin 依賴）
- Test: 對等 OpenFGA 測試

**Interfaces:**
- Consumes: Task 5。
- Produces: 後端不再有 Casbin/CASL engine/依賴；`GetAbility` 改由 OpenFGA `Check`/`ListObjects` 組裝該身分可用 (resource, action) 集合（前端 proxy 用）。

- [ ] **Step 1: 改 GetAbility 為 OpenFGA 驅動**

`ability.go`：`GetAbility` 以 OpenFGA `ListObjects`/`Check` 產出該身分可用權限清單（取代 CASL rules 載入）。

- [ ] **Step 2: 寫對等測試**

`ability_test.go`：`TestGetAbilityOpenFGA`（developer 全開、staff 依 tuple）。

- [ ] **Step 3: 跑測試確認通過** → Run `go test ./internal/domain/auth/... -v`。Expected: PASS。

- [ ] **Step 4: 刪除 Casbin/CASL code + 移除依賴**

刪 `casbin.go`/`rbac.go`/`casl/`；清理 `access.go` 內 casl import；`go mod tidy` 移除 casbin/casbin + govaluate；確認編譯過。

- [ ] **Step 5: 補齊受影響測試**

更新 `server_test.go`/`services_test.go`/`role_service_casl_test.go` 等移除 casl 參考，改 OpenFGA 對等。

- [ ] **Step 6: 全量跑測試 + `task check`**

Run: `task check`。Expected: fmt/vet/lint/test 全過。

- [ ] **Step 7: Commit**

```bash
git add -A backend
git commit -m "refactor(backend): 移除 Casbin/CASL;CASL 依賴退場;GetAbility 改 OpenFGA 驅動"
```

---

### Task 10: 整合測試補齊（D21 對應授權面）

**Files:**
- Create: `backend/internal/authz/openfga/integration_test.go`（依 `OPENFGA_*`/DSN skip 保護）
- Test: 真實 Postgres 整合

**Interfaces:**
- Produces: OpenFGA + RLS 的 cross-table 授權隔離整合測試（跨公司/部門/self）。

- [ ] **Step 1: 寫整合測試**（無 Postgres 自動 skip）
- [ ] **Step 2: 跑測試確認通過** → Run `go test ./internal/authz/... -tags integration -v`。Expected: PASS（無 DSN 時 SKIP）。
- [ ] **Step 3: Commit** `git add -A backend && git commit -m "test(backend): OpenFGA+RLS 授權隔離整合測試(D21)"`

---

## 執行驗收（DoD）
- `task check` 全過；`task vuln` 無高風險弱點。
- 無 Casbin/CASL 依賴與 code；`GetAbility` 由 OpenFGA 驅動。
- 核心表 migration + 7 內建角色 seed + developer（dev only）已落地。
- RLS 核心表政策生效；密碼為 Argon2id。
- OpenFGA + RLS 授權隔離整合測試（D21）通過。

*最後更新：2026-09-18*
