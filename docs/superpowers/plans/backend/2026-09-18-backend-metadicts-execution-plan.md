# Backend 03 — Metadicts 字典檔 執行計畫

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 落地 `metadicts` 字典檔領域（系統級預設 + 部門擴充），含 Ent schema、migration＋seed、Connect-RPC CRUD＋軟刪除＋稽核、合併查詢與 ListOptions 表單選項。

**Architecture:** 單一 `metadicts` 表以 `department_id IS NULL`＝系統預設、非 NULL＝部門擴充（D11）；CRUD 走既有 `services`／`domain` 層模式（比照 `user_service.go`），寫入與 `audit.Record` 同一 DB 交易（D18）；可見性以「系統預設＋當前部門擴充」合併查詢，RLS policy 僅「定義」不 ENABLE（比照 00010，D3 待接線後生效）。

**Tech Stack:** Go / ent（`go generate ./ent`）/ Connect-RPC（buf codegen）/ PostgreSQL / goose migration / enttest(sqlite) 單元測試。

**Spec:** `docs/superpowers/plans/backend/detail/03-metadicts-audit.md`（Task 2.5·子功能 2.5.1–2.5.4）、`docs/superpowers/plans/backend/2026-08-17-backend-03-metadicts-audit-plan.md`（Task 1–3）

**狀態基準：** 2026-09-18 盤點。audit 部分（`audit.Recorder` DB + `audit_logs` schema/migration 00009/00010）已由「02-UserService」批次提前落地；本計畫**僅處理 metadicts 字典檔**（Task 1–3）。

> **✅ 執行結果（2026-09-18）**：Task 1–6 全數完成並 commit。Task 5（合併查詢 + ListOptions）隨 Task 4 的 `metadictScope` 一併落地（非獨立 commit）。RLS D3 接線仍待（`00011` policy 僅定義不 ENABLE）。唯一約束（部分唯一索引）之 DB 層驗證於 sqlite 單元測試不套用，待 Postgres CI 驗證。

---

## Global Constraints

- 一律繁體中文程式註解與 commit message（AGENTS.md）。
- **測試只用 stdlib `testing` + enttest(sqlite)**；禁止 testify / testcontainers 新依賴（backend/AGENTS.md §4）。
- 提交前必跑 `task check`（fmt+vet+lint+test）；`task vuln` 因工具鏈錯配不跑（環境問題）。
- `metadicts` 單表兩層（`department_id IS NULL`＝系統預設，D11）；軟刪除 + 部分唯一索引 `(type, code, department_id)` 兩道（D10）。
- 錯誤統一以 Connect code 表述；列表分頁沿用 `meta` 格式、`per_page ≤ 100`。
- `order_source` 為系統級固定、API 不可異動。
- 寫入範圍：super(`data_scope=all`)→`department_id NULL`；dept_admin/staff→自動帶當前部門，**不接受請求自帶 `department_id`**。
- `type`/`code` 建立後不可修改；Update 僅允許 `display_name`/`sort_order`/`is_active`。
- RLS policy 僅定義不 ENABLE（比照 00010），不觸動 OpenFGA 共用表。
- codegen 命令：`go generate ./ent`（schema）；proto：`buf generate`（依 repo 既有 T16 慣例）。

---

## File Structure

- `backend/ent/schema/metadict.go` — Ent schema（欄位 + 部分唯一索引 + RLS 註解）
- `backend/ent/` — codegen 產碼（`go generate ./ent` 產生 `metadict.go` 查詢器等）
- `backend/database/migrations/00011_metadicts.sql` — 建表 + 索引 + RLS policy + 冪等 seed
- `backend/proto/metadict/v1/metadict.proto` — MetadictService RPC 定義
- `backend/internal/proto/metadict/v1/metadictv1connect/` + `metadict.pb.go` — buf 產碼
- `backend/internal/domain/metadicts/service.go` — Connect handler（List/Get/Create/Update/Delete/ListOptions）
- `backend/internal/domain/metadicts/usecase.go` — 交易 + 範圍推導 + audit.Record
- `backend/internal/domain/metadicts/repository.go` — 合併查詢（系統預設 + 部門擴充）、ListOptions
- `backend/internal/domain/metadicts/service_test.go` — 單元測試（enttest sqlite）
- `backend/internal/server/domains.go` — `mountMetadict()` 掛載（比照 `mountAuth` 中 `RegisterUserServices`）

> 註：既有 UserService 直接放 `backend/internal/services/`。metadicts 依細部文件建議放 `internal/domain/metadicts/`，比照 `domainauth.NewAbilityHandler` 之 `domain/` 慣例。於 Task 2 掛載時決定用哪一層並維持一致。

---

## Task 1: metadicts schema + codegen

**Files:**
- Create: `backend/ent/schema/metadict.go`
- Modify: `backend/ent/`（codegen 產碼）
- Test: 以 `task check` 驗證編譯與既有測試不回歸

**Interfaces:**
- Consumes: 既有 `ent` 套件、`role.go`/`auditlog.go` schema 慣例
- Produces: `ent.Metadict` 實體與 `ent.MetadictQuery` 等查詢 API；`metadicts` 資料表欄位

- [ ] **Step 1: 建立 `metadict.go` schema**
  欄位：`type`(String, NotEmpty)、`code`(String, NotEmpty)、`display_name`(String, NotEmpty)、`department_id`(Int, Optional/Nillable)、`sort_order`(Int, Default 0)、`is_active`(Bool, Default true)、`deleted_at`(Time, Optional/Nillable)、`created_at`/`updated_at`(Time, Default mutableNow)。部分唯一索引兩道（D10/D11）：
  - `metadicts_sys_unique` → `(type, code)` `WHERE department_id IS NULL AND deleted_at IS NULL`
  - `metadicts_dept_unique` → `(type, code, department_id)` `WHERE department_id IS NOT NULL AND deleted_at IS NULL`
  查詢索引：`(type, department_id, is_active)`。

- [ ] **Step 2: 產生 code**
  Run: `go generate ./ent`
  Expected: 產生 `backend/ent/metadict.go`、`ent/metadict/metadict.go`、mutation 等。

- [ ] **Step 3: 驗證編譯**
  Run: `task check`
  Expected: PASS，schema 註冊無誤、既有測試不回歸。

- [ ] **Step 4: Commit**
  `git add -A && git commit -m "feat(backend): metadicts Ent schema(系統預設+部門擴充、部分唯一索引 D10/D11)"`

---

## Task 2: migration 00011 + RLS policy + 冪等 seed

**Files:**
- Create: `backend/database/migrations/00011_metadicts.sql`
- Test: 以乾淨+重複執行驗證冪等（若有本地/CI 資料庫），或依既有的 migration 測試慣例

**Interfaces:**
- Consumes: `00010` RLS policy 模式（`app.current_data_scope`/`app.current_company_id`/`app.current_department_id`，僅定義不 ENABLE）
- Produces: `metadicts` 表 + 索引 + RLS policy + 系統預設 seed（`unit`/`payment_method`/`settlement_method`/`customer_type`/`invoice_type`，`department_id NULL`；`order_source` 固定 `W`/`A`）

- [ ] **Step 1: 撰寫 migration**
  建表 `metadicts`（欄位對齊 Task 1 schema）；三組索引；RLS policy `core_metadicts_scope`：
  - USING：`data_scope='all'` 或 `department_id IS NULL OR department_id = current_department_id`
  - WITH CHECK：`data_scope='all'` 或 `department_id = current_department_id`
  冪等 seed：`INSERT ... SELECT ... WHERE NOT EXISTS`（或 `ON CONFLICT DO NOTHING` 搭配索引），系統預設 `unit`/`payment_method`/`settlement_method`/`customer_type`/`invoice_type` + `order_source`(`W`/`A`)。
  Down：`DROP TABLE metadicts`。

- [ ] **Step 2: RLS 註記**
  檔頭註解：僅定義不 ENABLE（比照 00010），待 D3 每請求交易層落定後再啟用。

- [ ] **Step 3: 驗證**
  Run: 既有 migration 執行路徑（若 `make db-up` / `task` 有對應 target）或 migration 測試；確認重複執行第二次不報錯、不重複 seed。
  Expected: PASS（若環境無 DB，則記錄待 CI 驗證，不 blocking 本任務提交編譯層）。

- [ ] **Step 4: Commit**
  `git add -A && git commit -m "feat(backend): migration 00011 metadicts 建表+RLS policy+系統預設 seed(冪等)"`

---

## Task 3: proto + codegen MetadictService

**Files:**
- Create: `backend/proto/metadict/v1/metadict.proto`
- Modify: `backend/internal/proto/metadict/v1/...`（buf 產碼）
- Test: `task check` 編譯

**Interfaces:**
- Consumes: 既有 buf/proto 慣例（比照 `proxy/salesorder/v1/user.proto`）
- Produces: `metadictv1connect.MetadictServiceHandler`／`Client`，`Metadict`/`ListMetadictsRequest`/`...` message

- [ ] **Step 1: 撰寫 proto**
  `MetadictService` RPC：
  - `List(ListMetadictsRequest{type, include_inactive, include_deleted, page, per_page}) returns (ListMetadictsResponse{items, meta})`
  - `Get(GetMetadictRequest{id}) returns (Metadict)`
  - `Create(CreateMetadictRequest{type, code, display_name, sort_order, is_active}) returns (Metadict)`
  - `Update(UpdateMetadictRequest{id, display_name, sort_order, is_active}) returns (Metadict)`
  - `Delete(DeleteMetadictRequest{id}) returns (Empty)`
  - `ListOptions(ListOptionsRequest{type, keyword}) returns (ListOptionsResponse{options})`
  `Metadict` message：`id, type, code, display_name, department_id(optional), sort_order, is_active, created_at, updated_at`。
  `Option` message：`code, display_name`。

- [ ] **Step 2: buf 產生**
  Run: repo 既有 proto 產生命令（如 `buf generate` / 對應 `task` target）
  Expected: 產生 `metadict.pb.go` + `metadictv1connect/`。

- [ ] **Step 3: 驗證編譯**
  Run: `task check`
  Expected: PASS。

- [ ] **Step 4: Commit**
  `git add -A && git commit -m "feat(backend): proto/metadict/v1 MetadictService RPC(CRUD+ListOptions) 與 buf 產碼"`

---

## Task 4: MetadictService CRUD + 軟刪除 + 稽核（2.5.2）

**Files:**
- Create: `backend/internal/domain/metadicts/service.go`、`usecase.go`、`repository.go`
- Test: `backend/internal/domain/metadicts/service_test.go`
- Modify: `backend/internal/server/domains.go`（`mountMetadict`）

**Interfaces:**
- Consumes: Task 1/3 產物；`authz.IdentityFrom(ctx)`、`audit.MetaFrom(ctx)`、`audit.Record(ctx, tx, audit.Entry{...})`
- Produces: `MetadictService` Connect handler 掛載於 `/api/v1`；Create/Update/Delete 寫入 audit（action `create`/`update`/`delete`, resource_type `metadict`）
- 範圍推導：`IdentityFrom` → super 建 `department_id NULL`；dept_admin/staff 強制 `current department`（`IdentityFrom` 之部門脈絡），不接受請求自帶 `department_id`。
- Update 僅 `display_name`/`sort_order`/`is_active`；改 `type`/`code` → `failed_precondition`；`order_source`（系統固定）Update/Delete → `failed_precondition`。
- Delete = 軟刪除（`deleted_at`），並寫稽核。

- [ ] **Step 1: 先寫失敗測試（TDD）**
  用 enttest(sqlite) 覆蓋：
  - super 建立系統級列（`department_id NULL`）；dept_admin 建立列自動帶自己部門。
  - dept_admin 對系統級/他部門列 Update/Delete → `permission_denied`（應用層）。
  - Update 帶 `code` 變更 / 對 `order_source` 異動 → `failed_precondition`。
  - 軟刪除後 List 預設查不到、Get → `not_found`；同 `code` 可立即重建（`already_exists` 不再觸發）。
  - Create/Update/Delete 成功後 audit_logs 有對應紀錄（同一 tx）；注入稽核寫入失敗 → 業務異動回滾（D18）。
  Run: 確認測試 FAIL（服務未定義）。

- [ ] **Step 2: 實作**
  `service.go`（Connect handler，轉發 usecase）、`usecase.go`（交易 + 範圍推導 + `audit.Record`）、`repository.go`（`WithTx` 與範圍 where）。

- [ ] **Step 3: 掛載**
  `domains.go`：`InitDomains()` 加 `s.mountMetadict()`；`mountMetadict` 比照 `mountAuth`，開 entClient → `metadicts.RegisterMetadictServices(apiMux, entClient)`（或 `NewMetadictServiceHandler`）。

- [ ] **Step 4: 跑測試**
  Run: `task check`
  Expected: PASS，全部新測試綠。

- [ ] **Step 5: Commit**
  `git add -A && git commit -m "feat(backend): MetadictService CRUD+軟刪除+audit(D10/D18)、範圍推導、mountMetadict 掛載"`

---

## Task 5: 合併查詢 + ListOptions（2.5.3–2.5.4）

**Files:**
- Modify: `backend/internal/domain/metadicts/repository.go`、`usecase.go`、`service.go`
- Test: `backend/internal/domain/metadicts/service_test.go`（擴充）

**Interfaces:**
- Consumes: Task 4 產物
- Produces: `List`/`Get` 合併「系統預設 + 當前部門擴充」（`department_id IS NULL OR = 當前部門`），部門間隔離；`ListOptions` 回可選用選項（`is_active=true` + 軟刪除排除 + keyword 過濾）
- 無部門角色（company_admin）僅見系統預設；客戶端身分 ListOptions 僅回系統預設（非錯誤）。

- [ ] **Step 1: 先寫失敗測試**
  覆蓋：
  - 部門 A 使用者 List 見系統預設 + 部門 A 擴充，不見部門 B 擴充。
  - 無部門角色僅見系統預設；不回 `invalid_argument`、不漏系統值。
  - 停用（`is_active=false`）後 List 仍回（管理視角）、ListOptions 立即排除。
  - ListOptions：含啟用值、排除軟刪除/停用、keyword 過濾、排序 `sort_order`→`code`；客戶端僅系統預設。
  Run: 確認 FAIL。

- [ ] **Step 2: 實作**
  `repository.go` 合併查詢 + 排序；`usecase.go` ListOptions 條件 + 可選短 TTL 快取（快取失效於 Create/Update/Delete）。

- [ ] **Step 3: 跑測試**
  Run: `task check`
  Expected: PASS。

- [ ] **Step 4: Commit**
  `git add -A && git commit -m "feat(backend): MetadictService 合併查詢(系統+部門)與 ListOptions(2.5.3-2.5.4)"`

---

## Task 6: 範圍複審 + README 對齊 + 收尾

**Files:**
- Modify: `docs/superpowers/plans/README.md`（03 行 → metadicts Task 1-3 已落地，audit 已提前落地；狀態 🟡 部分）
- Modify: `docs/superpowers/plans/backend/2026-08-17-backend-03-metadicts-audit-plan.md`（Task 1-5 狀態更新）
- Test: `task check` 全綠

- [ ] **Step 1: 複審**
  對本批次以 requesting-code-review skill 流程（內聯唯讀）複審，處理發現項。

- [ ] **Step 2: 更新索引**
  README 與 03 計畫「反映現況」對齊。

- [ ] **Step 3: 最終驗證**
  Run: `task check`
  Expected: PASS。

- [ ] **Step 4: Commit**
  `git add -A && git commit -m "docs(plans): 03-metadicts 反映現況(metadicts 落地)並對齊 README"`

---

## 複審登錄（requesting-code-review, 2026-09-18）

- ✅ **已修 #1**：`ListOptions` 原用 `metadictScope(super)`（identity → 洩漏各部門私有選項），已改為 super **僅系統預設**（與 `ListMetadicts` 預設一致）；補測試 `TestListOptionsSuperSystemOnly`。
- ✅ **已修 #3**：`CreateMetadict` 加 `code` 長度上限（`maxMetadictCodeLen = 64` → `invalid_argument`）。
- ⬜ **#2 待辦**：D18「稽核寫入失敗 → 業務回滾」尚未有測試（sqlite 強制稽核失敗注入較難）；建議以 ent hook 或後續整合測試（Postgres）補。
- ✅ **已修 #5**：`GetMetadict` 合併為單次查詢（scope 條件併入首查，範圍外一律 `not_found`）。
- **#6 註記（非缺陷）**：系統級字典被軟刪後**同一支 migration 再執行**會 re-seed（goose 版本化正常不會重跑;down/up 或全新 DB 才觸發;部分唯一索引允許已刪+現行共存）。無需動作。
- ⬜ **#4 待決**：`include_inactive` 為死欄位;根因是「欄位註解 vs §2.5.3 驗收」矛盾。待決：移除欄位 / 改為具語意 / 保留並標註。

---

*最後更新：2026-09-18（metadicts 執行計畫；含複審修正）*
