# Backend 使用者管理（UserService）+ 最小稽核地基 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 實作 `UserService`（02 Task 3 使用者 CRUD + AssignRole + Deactivate + ForceLogout），並建立最小 `audit_logs` 稽核地基，使後端使用者管理可端到端運作且關鍵操作依 D18 同行寫入稽核。

**Architecture:** 沿用既有 services 模式（`internal/services/*_service.go` + `proto/salesorder/v1/*.proto`）。使用者 CRUD 以 `*ent.Client` 直作，管理範圍控制（super / company_admin / dept_admin）在 usecase 入口統一判斷，RLS 兜底。授權閘門沿用現有 `server.protectedRPC`（OpenFGA Check），UserService 各 RPC path 加入對映。audit 以最小 `internal/audit` 包提供同事務 recorder，供 AssignRole/Deactivate/ForceLogout 寫稽核（D18）。角色指派同步 OpenFGA `assigned` tuple（對齊 role_service `syncRolePermissions` 模式）。

**Tech Stack:** Go 1.25.7、ent、connectrpc、PostgreSQL（pgx）、goose、buf（Go 型別生成）。

**Spec:**
- `docs/superpowers/plans/backend/detail/02-tenancy-users.md` Task 2.3（2.3.1/2.3.2/2.3.3）
- `docs/superpowers/plans/backend/detail/01-auth.md` 1.6.5（middleware，已完成）、1.7（ForceLogout）、1.11（developer，已部分完成）
- `docs/superpowers/plans/backend/detail/03-metadicts-audit.md` 2.6.1/2.6.2（audit_logs 表 + recorder，最小範圍）
- 決策 D5/D6/D18/D22/D28

## Global Constraints

- 繁體中文註解與 commit message（專案慣例）。
- 測試只用 stdlib `testing` + enttest（sqlite）；禁止 testify/testcontainers 新測試依賴。
- 授權：未登入 → `Unauthenticated`、越權 → `PermissionDenied`、跨租戶 → `PermissionDenied`、輸入驗證失敗 → `InvalidArgument`、不存在 → `NotFound`、狀態不允許 → `FailedPrecondition`、重複 → `AlreadyExists`。
- 稽核（D18）與業務異動同一 DB 交易，同成功同失敗；敏感欄位（密碼雜湊）不入 snapshot。
- 提交前必跑 `task check`（fmt+vet+lint+test）。
- 不在本計畫範圍：`metadicts` 字典檔（03 剩餘）、audit 查詢 API（03 2.6.3）、用戶 schema 唯一性改「公司內唯一」（01 Task 1 殘留，沿用現有 email 全局唯一）。

---

### Task 1: audit_logs 表 migration + minimal ent schema

**Files:**
- Create: `backend/ent/schema/auditlog.go`
- Create: `backend/database/migrations/00009_audit_logs.sql`
- Run: `go generate ./ent`（ent 產碼，無 gen 檔則 `go generate ./ent/generate.go`）

**Interfaces:**
- Consumes: 既有 `users`/`companies`/`departments` 表（00005）。
- Produces: `audit_logs` 表 + `ent.AuditLog` 實體；欄位 `id, company_id, department_id(可空), user_id, action, resource_type, resource_id, before_snapshot(jsonb,可空), after_snapshot(jsonb,可空), ip_address, user_agent, created_at`；索引 `(company_id, created_at)`。**不建** `deleted_at`/`updated_at`（D10 §5.4）。

- [ ] **Step 1: 寫 ent schema + migration**

`ent/schema/auditlog.go`：`action` 為 `field.String`（不對應 enum 欄位層級，action 值由業務界定；對齊 detail 2.6.1）。`before/after_snapshot` 為 `field.JSON`（可空）。`00009_audit_logs.sql`：Up 建表 + 索引，Down DROP TABLE；必含 Up/Down。

- [ ] **Step 2: 產生 ent code**

Run: `cd backend && go generate ./ent/generate.go`。Expected: `ent/auditlog/` 產出。若既有產碼流程是 `go generate ./ent`，以該入口為準。

- [ ] **Step 3: 驗證 migration 可執行**

Run: 依專案 `migrate:up`（需 DB）或人工檢核 SQL。若本機無 DB 可跳過，標註待整合驗證。

- [ ] **Step 4: Commit**

```bash
git add backend/ent/schema/auditlog.go backend/database/migrations/00009_audit_logs.sql backend/ent
git commit -m "feat(backend): audit_logs 表 + ent schema(D18 稽核地基)"
```

---

### Task 2: internal/audit recorder（同事務寫入入口）

**Files:**
- Create: `backend/internal/audit/recorder.go`
- Test: `backend/internal/audit/recorder_test.go`

**Interfaces:**
- Consumes: `ent.Tx`（交易內 client）、請求脈絡（操作者 / 公司 / 部門 / IP / User-Agent）。
- Produces: `audit.Record(ctx, tx *ent.Tx, e Entry) error`；`Entry{Action, ResourceType, ResourceID, Before, After map[string]any}`。脈絡 `WithContext(ctx, Meta)` 由 middleware 注入；`MetaFrom(ctx)` 讀取。IP/User-Agent 於此 Task 以 `Meta` 顯式傳入（middleware 注入段於 Task 5 串接）。

- [ ] **Step 1: 寫失敗測試**

`recorder_test.go`：`TestRecordWritesSnapshot`（enttest sqlite，開 tx + auditlog 建立；assert action/resource_type/after_snapshot 寫入且不洩漏敏感欄位——以測試傳入含 password_hash 的 snapshot，assert 不落盤）。

- [ ] **Step 2: 跑測試確認失敗**

Run: `go test ./internal/audit/... -v`。Expected: FAIL（audit 包不存在）。

- [ ] **Step 3: 實作 recorder**

`recorder.go`：`Record` 以 tx client 建立 `ent.AuditLog`，`Before`/`After` 以 `map[string]any` → `jsonpb`/`structpb` 或 `json.Marshal` 落 `before/after_snapshot`（依 AuditLog schema JSON 欄位型別）。敏感欄位過濾：掃 `Before`/`After` 鍵，跳過 `password_hash`/`hash`/`secret`/`token`。

- [ ] **Step 4: 跑測試確認通過**

Run: `go test ./internal/audit/... -v`。Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/audit/recorder.go backend/internal/audit/recorder_test.go
git commit -m "feat(backend): audit recorder 同事務寫入 D18"
```

---

### Task 3: user.proto 定義 + buf 生成

**Files:**
- Create: `backend/proto/salesorder/v1/user.proto`
- Run: `task proto:gen`（buf 產生 Go；Dart/TS 為附帶產出，不需逐一核對）

**Interfaces:**
- Produces: `salesorderv1.UserService`（7 RPC）+ `User` message；`User` 含 `id, name, email, phone, employee_no, status, role, department_id, company_id, is_customer, account_name`。
- RPC 集合：
  - `ListUsers(page, page_size, company_id, department_id, role, status) → {users, pagination}`
  - `GetUser(user_id) → {user}`
  - `CreateUser(name, email, company_id, department_id, role, phone, employee_no) → {user}`（員工帳號，不設密碼）
  - `UpdateUser(user_id, name?, department_id?, phone?, employee_no?, status?) → {user}`
  - `AssignRole(user_id, role, department_id?) → {user}`（含 guest 審核：status pending → active）
  - `Deactivate(user_id) → {}`
  - `ForceLogout(user_id) → {}`

> 註：proto field 用 `optional`（出現即更新）對齊 company.proto。`User` 重複欄位用 `repeated` 於 List。

- [ ] **Step 1: 寫 user.proto**

比照 `company.proto`/`role.proto` 風格（package salesorder.v1、go_package、註解）。

- [ ] **Step 2: 跑 buf generate**

Run: `cd backend && task proto:gen`。Expected: `internal/proto/salesorder/v1/user.pb.go` + `salesorderv1connect/user.connect.go` 產出且編譯過。

- [ ] **Step 3: 確認編譯**

Run: `cd backend && go build ./...`。Expected: PASS。

- [ ] **Step 4: Commit**

```bash
git add backend/proto/salesorder/v1/user.proto backend/internal/proto/salesorder/v1
git commit -m "feat(proto): UserService 定義(v1)"
```

---

### Task 4: UserService CRUD + 管理範圍控制

**Files:**
- Create: `backend/internal/services/user_service.go`
- Test: `backend/internal/services/user_service_test.go`
- Modify: `backend/internal/server/server.go`（`protectedRPC` 加入 7 條 UserService path → `{"user", "read"/"write"}`）
- Modify: `backend/internal/server/domains.go`（`mountAuth` 掛載 `RegisterUserServices`）

**Interfaces:**
- Consumes: `authz.Identity`/`authz.IdentityFrom`、`auth.RolesFor`/`auth.ScopeForRole`、`ent.Client`、`ent.Tx`。
- Produces: `NewUserService(db)` + `RegisterUserServices(mux *http.ServeMux, db *ent.Client)`；範圍函式 `scopeForTarget(ctx, target *ent.User) error`（super 不限 / company_admin 公司 / dept_admin 部門 + staff）。

- [ ] **Step 1: 寫失敗測試（範圍矩陣）**

`user_service_test.go`：授權門檻矩陣（未登入 / guest / staff / dept_admin / company_admin / super）—— ListUsers 與 UpdateUser 各測。`dept_admin` 改他部門 → `PermissionDenied`；`company_admin` 改他公司 → `PermissionDenied`；未登入 → `Unauthenticated`。測試以 `authz.WithIdentity` 注入 ctx + enttest sqlite。

- [ ] **Step 2: 跑測試確認失敗**

Run: `go test ./internal/services/... -run User -v`。Expected: FAIL（user_service 不存在）。

- [ ] **Step 3: 實作範圍控制 + CRUD**

`user_service.go`：`scopeForTarget` 依操作者 `Identity.Roles` 判斷目標使用者範圍；`ListUsers` 依 `Identity` 強制注入 company_id/department_id 過濾（dept_admin 固定自己部門 + staff）；`CreateUser`（員工帳號，`password_hash` 以既有 schema 必填——若 schema `NotEmpty()` 卡住則先補暫時值，待 01 首登 OAuth 改寫，於註解標註）；`GetUser`；`UpdateUser`（僅更新出現欄位）—— 全部走 `authz.IdentityFrom` 門檻（未登入 → `Unauthenticated`）。

- [ ] **Step 4: 跑測試確認通過**

Run: `go test ./internal/services/... -v`。Expected: PASS。

- [ ] **Step 5: 掛載 + protectedRPC 對映**

`server.go` `protectedRPC` 加：`/salesorder.v1.UserService/ListUsers`、`GetUser` → `{"user","read"}`；`Create`/`Update`/`AssignRole`/`Deactivate`/`ForceLogout` → `{"user","write"}`。`domains.go` `mountAuth` 加 `services.RegisterUserServices(apiMux, entClient)`。

- [ ] **Step 6: 全量測試 + task check**

Run: `cd backend && task check`。Expected: 全過。

- [ ] **Step 7: Commit**

```bash
git add backend/internal/services/user_service.go backend/internal/services/user_service_test.go backend/internal/server/server.go backend/internal/server/domains.go
git commit -m "feat(backend): UserService 使用者 CRUD + 管理範圍控制"
```

---

### Task 5: AssignRole（含 guest 審核）+ OpenFGA assigned tuple 同步 + audit + token_version

**Files:**
- Modify: `backend/internal/services/user_service.go`
- Test: `backend/internal/services/user_service_test.go`
- Modify: `backend/internal/server/server.go`（middleware 注入 audit Meta：IP / User-Agent）

**Interfaces:**
- Consumes: `ent.Tx`、`audit.Record`、`authz.EngineFrom`（`WriteTuple`/`DeleteTuple`）、`auth.RolesFor`。
- Produces: `AssignRole` 於交易內：角色變更時 `token_version + 1` + 寫 `audit_logs`（action `role_change`）→ **交易 commit 後**同步 OpenFGA `assigned` tuple（刪舊 role、寫新 role——對齊 role_service `syncRolePermissions` 模式，OpenFGA 與業務非同交易）。guest 審核：status pending → active 並指派 role/department。developer 帳號操作照常寫稽核。

- [ ] **Step 1: 寫失敗測試**

`user_service_test.go`：`TestAssignRoleBumpsTokenVersion`（AssignRole 後 user.token_version +1）；`TestAssignRoleWritesAuditSameTx`（注入稽核寫入失敗 → 業務異動不存在 [回滾驗證]）；`TestAssignRoleGuestApprove`（guest pending → active）；`TestAssignRoleRoleChange`（openfga 引擎注入時 assert 對應 assigned tuple 異動，缺引擎時略過）。

- [ ] **Step 2: 跑測試確認失敗**

Run: `go test ./internal/services/... -run AssignRole -v`。Expected: FAIL。

- [ ] **Step 3: 實作**

交易內更新 role（+department_id，guest 審核含 status）→ `token_version +1` → `audit.Record(tx, entry{action:"role_change"})` → `tx.Commit()` → (`EngineFrom` 非 nil 時) `DeleteTuple(old)` + `WriteTuple(new)`。audit 失敗 → 整體回滾。

- [ ] **Step 4: middleware 注入 audit Meta**

`server.go` `authzMiddleware`：解析身分後，組 `audit.Meta{UserID, CompanyID, DepartmentID, IP, UserAgent}` 並 `audit.WithContext(ctx, meta)`；`userAgent`/`ip` 自 `r.Header`/`r.RemoteAddr`。

- [ ] **Step 5: 跑測試確認通過**

Run: `go test ./internal/services/... -v`。Expected: PASS。

- [ ] **Step 6: Commit**

```bash
git add backend/internal/services/user_service.go backend/internal/server/server.go
git commit -m "feat(backend): UserService AssignRole + OpenFGA tuple 同步 + 稽核(D18)"
```

---

### Task 6: Deactivate + ForceLogout + 稽核

**Files:**
- Modify: `backend/internal/services/user_service.go`
- Test: `backend/internal/services/user_service_test.go`

**Interfaces:**
- Consumes: `ent.Tx`、`audit`、`authz.Identity`。
- Produces: `Deactivate`（同一交易：範圍判斷 → status=inactive → token_version+1 → audit action `update`）→ commit 後刪該用戶 Web session（Valkey，若 session store 可達；失敗僅記 log，tv 已 +1 由 middleware 兜底）。`ForceLogout`（同一交易：範圍判斷 → token_version+1 → audit action `force_logout`）；範圍對自己呼叫 → `InvalidArgument`；dept_admin 僅自己部門 staff、company_admin 僅自己公司。

- [ ] **Step 1: 寫失敗測試**

`TestDeactivateBumpsTokenVersionAndAudit`（deactivate 後 status inactive、tv+1、audit 存在）；`TestForceLogoutBumpsTokenVersionAndAudit`；`TestForceLogoutSelf`（對自己 → `InvalidArgument`）；`TestDeactivateScope`（dept_admin 停用他部門 → `PermissionDenied`）。

- [ ] **Step 2: 跑測試確認失敗**

Run: `go test ./internal/services/... -run "Deactivate|ForceLogout" -v`。Expected: FAIL。

- [ ] **Step 3: 實作**

`Deactivate` / `ForceLogout` 各自：`scopeForTarget` → tx 交易內更新 + tv+1 + `audit.Record` → commit →（Deactivate）Valkey session 刪除 best-effort。

- [ ] **Step 4: 跑測試確認通過**

Run: `go test ./internal/services/... -v`。Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/user_service.go backend/internal/services/user_service_test.go
git commit -m "feat(backend): UserService Deactivate + ForceLogout + 稽核(D18)"
```

---

### Task 7: 收尾驗證 + 文件更新

**Files:**
- Modify: `docs/superpowers/plans/backend/2026-08-17-backend-02-tenancy-users-plan.md`（Task 3 勾選 + 狀態）
- Modify: `docs/superpowers/plans/README.md`（02 狀態「🟡 部分」→ 註記 UserService 已落地）
- Modify: 本計畫檔「收尾註記」加實作結果

- [ ] **Step 1: 全量 `task check` + `task vuln`**

Run: `cd backend && task check && task vuln`。Expected: fmt/vet/lint/test 全過；vuln 無高風險。

- [ ] **Step 2: 更新 02 計畫 Task 3 勾選**

將 02 Task 3 各 Step 打勾（`- [x]`），Task 6/7 殘留標註。

- [ ] **Step 3: 更新計畫總索引**

`plans/README.md`：02-tenancy-users 說明補「UserService 已落地（2026-09-18）」。

- [ ] **Step 4: Commit**

```bash
git add docs/superpowers/plans/backend/2026-08-17-backend-02-tenancy-users-plan.md docs/superpowers/plans/README.md docs/superpowers/plans/backend/2026-09-18-backend-user-service-plan.md
git commit -m "docs(plans): 02 Task 3 UserService 完工狀態更新"
```

---

## 執行驗收（DoD）

- `task check` 全過；`task vuln` 無高風險弱點。
- `audit_logs` 表 + ent 實體 + migration（00009）落地；無 `deleted_at`。
- `internal/audit.Record` 同事務寫稽核，敏感欄位不外洩。
- `UserService` 7 RPC 全數實作並掛載，`protectedRPC` 對映正確。
- AssignRole 含 guest 審核 + token_version+1 + OpenFGA assigned tuple 同步 + audit。
- Deactivate / ForceLogout 含範圍控制 + tv+1 + audit。
- 02 計畫 Task 3 勾選更新、README 同步。
