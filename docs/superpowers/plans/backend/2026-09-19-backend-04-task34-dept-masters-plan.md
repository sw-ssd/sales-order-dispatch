# Backend 04 — Task 3.4 部門級主檔(倉庫/車次/處理規格/商品分類)執行計畫

> **For agentic workers:** 以 superpowers:subagent-driven-development 或 executing-plans 依任務逐步實作。步驟用 `- [ ]` 追蹤。

**Goal:** 建立四個部門級主檔 CRUD：`warehouses`(3.4.1)、`routes`(3.4.2)、`processing_specs`(3.4.3，泛化自 cutting_specs)、`product_categories`(3.4.4)；皆 `company_id`/`department_id`、軟刪除、跨部門不可見；形狀一致的 `List/Create/Update/Delete/Restore` + 分頁 meta。

**Architecture:** 每個主檔一個 Ent schema + 對應 migration(00016 全四表) + Connect-RPC service。權限 dept_admin/staff 限本部門，super/company_admin 公司層。RLS 依既有「僅定義不 ENABLE」(D3 未接線)。每異動寫稽核(D18)。

**Tech Stack:** Go / ent / Connect-RPC / PostgreSQL / enttest(sqlite)。測試僅 stdlib `testing` + enttest(backend/AGENTS.md §4)。

**Spec:** `docs/superpowers/plans/backend/detail/04-master-data.md` §3.4.1–3.4.4。

**狀態基準:** 2026-09-19。04 Task 4(地址/聯絡人)於 2026-09-18 完成；本波為 Task 3.4。

## 設計重點(2026-09-19 定稿)

- **processing_specs(3.4.3)**:`code/name/kind(開放集, metadicts processing_kind 背書)/applies_to_processing/applies_to_picking(多值旗標,至少其一 true)/attributes(JSONB,後端不解析)/sort_order/is_active`。
- **combos(案例 2)記為 Phase-2**,本波不建(僅文件註記)。
- `unit` 換算屬 3.3.3,與本波無關;`product_processing_specs` 多對多在 3.3 建,本波只建 `processing_specs` 主檔。

## Global Constraints

- 繁中註解與 commit message;測試僅 stdlib + enttest(sqlite);提交前 `task check`。
- 所有主檔:寫入 `company_id`/`department_id` 由 session 租戶注入,不接受 Request 指定跨部門值。
- `code` 部門內部分唯一 `(department_id, code) WHERE deleted_at IS NULL`。
- 軟刪除 + 稽核同交易(D18);List 預設排除已刪除(可 include_deleted)。
- 跨部門存取 `not_found`(沿用 codebase 慣例);同碼 `already_exists`;必填缺漏/非法值 `invalid_argument`。

## File Structure(Services 統一放 `internal/services/`,與既有 CustomerService 同層)

- Ent schema:`backend/ent/schema/warehouse.go`、`route.go`、`processingspec.go`、`productcategory.go`(皆公司/部門欄位 + 軟刪除 + 稽核 + 索引)
- Migration:`backend/database/migrations/00016_department_masters.sql`(四表 + 部分唯一索引 + RLS 定義)
- Proto:`backend/proto/masters/v1/*.proto`(四 service,各 5 RPC = 20 RPC)→ buf 重產三端
- Services:`backend/internal/services/warehouse_service.go`、`route_service.go`、`processingspec_service.go`、`productcategory_service.go`(+ 各 test)
- 掛載:`backend/internal/server/domains.go`(RegisterXxxServices)

## Task 1: schema + migration + ent gen(TDD)
- [x] 四 schema + `go generate ./ent` 編譯;migration 00016 四表、部分唯一索引、RLS、稽核欄位
- [x] `task check` → commit

## Task 2: proto + buf 三端重產(TDD)
- [x] masters.proto 四 service(各 `ListXXX/CreateXXX/UpdateXXX/DeleteXXX/RestoreXXX`;細節計畫 3.4 無 Get RPC)
- [x] buf generate(Go/TS/Dart);後端 build 通過
- [x] `task check` → commit

## Task 3: 服務實作 + 掛載(TDD)
- [x] 共用部門級 scope helper + 各 service 五法+restore;稽核同事務;`domains.go` 掛載
- [x] `task check` → commit

## Task 4: 測試 + 文件對齊(TDD)
- [x] 測試:CRUD、跨部門 not_found、軟刪除/復原、include_deleted、processing_spec 旗標驗證+attributes
- [x] 04 計畫 §3.4.1–3.4.4 驗收打勾
- [x] `task check` 全綠 → commit

> 註:`(department_id, code)` 同碼 already_exists 與軟刪除後 code 重用為 Postgres 部分唯一索引(migration)層保證,sqlite enttest 無此索引,故未以單元測試覆蓋(與 customers 同慣例)。

## 已知缺口 / 後續
- **OpenFGA 閘門**:本波四主檔暫不納入 `protectedRPC`(**依決策 D33**);納入要件與觸發時機見 D33。
- **combos(組合包)**:Phase-2(預先定義主檔 + combo_items BOM 炸開換算 + 雙軌包裝規格)。
- **RLS 接線(ApplyRLS)**:D3 仍為獨立設計任務,本波維持「僅定義不 ENABLE」。
- **3.3 商品 / 3.5 客戶專屬**:下一波(依賴 3.4 之 warehouse/category/processing_spec 驗證)。

---

*最後更新 2026-09-19*
