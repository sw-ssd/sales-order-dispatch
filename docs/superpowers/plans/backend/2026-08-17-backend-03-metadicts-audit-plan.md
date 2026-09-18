# Backend 03 — 字典檔與稽核日誌 執行計畫（現況對齊版）

> **性質**：原為目標型執行計畫（含內嵌目標程式碼）。經 2026-09-18 盤點（codebase-memory 知識圖譜 + git）重建，為**反映現況的執行計畫**。
>
> **狀態基準**：2026-09-18 盤點。**本計畫對應領域（metadicts / audit）實際尚未實作任何 code**——`ent/schema` 僅 company/department/role/rolepermission/user 五檔、migrations 僅 4 個、`internal/` 無 metadicts/audit 目錄、無 audit.Recorder DB 實作。以下為保留的目標計畫架構，全數 ⬜ 未開始。
>
> **對應設計**：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）、決策 `D11/D18/D27`
> **細部文件**：`docs/superpowers/plans/backend/detail/03-metadicts-audit.md`、共通規則 `detail/00-index.md` §3
> **前置依賴**：01-auth（`audit.Recorder` 介面 / Noop、testutil、RLS 注入）、02（軟刪除/範圍慣例）

---

## 執行狀態總覽

| Task | 內容 | 狀態 |
|---|---|---|
| 1 | metadicts schema、migration、系統預設 seed（細部 2.5.1） | ✅ 完成（2026-09-18）|
| 2 | MetadictService CRUD 與軟刪除（細部 2.5.2） | ✅ 完成（2026-09-18）|
| 3 | 合併查詢與 ListOptions（細部 2.5.3–2.5.4） | ✅ 完成（2026-09-18，隨 Task 2 一併落地）|
| 4 | audit_logs schema 與 migration（細部 2.6.1） | ✅ 完成（2026-09-18，02-UserService 批次提前落地:00009/00010）|
| 5 | audit.Recorder 的 DB 實作（細部 2.6.2，接管 01 NoopRecorder） | ✅ 完成（2026-09-18，02-UserService 批次提前落地:internal/audit）|
| 6 | 稽核查詢 API（細部 2.6.3） | ⬜ 未開始（= A4,待後續批次）|

**實作範圍**：metadicts（Task 1–3）＋ audit 地基（Task 4–5）皆已落地；**剩 Task 6 稽核查詢 API**。執行細節見 `2026-09-18-backend-metadicts-execution-plan.md`。

---

## Global Constraints（目標，精煉自原檔）

- `metadicts` 單表兩層（`department_id IS NULL` = 系統預設，D11）；軟刪除 + 部分唯一索引 `(type, code, department_id)` 兩道（D10）。
- `audit_logs` 不可變（無 `deleted_at` / `updated_at`）；稽核寫入失敗 = 交易回滾，不得略過（D18）；snapshot 永不收錄敏感欄位。
- 稽核查詢預設時間窗近 3 個月（D27）；字典 `type` 合法值 unit/payment_method/settlement_method/customer_type/invoice_type/order_source（order_source 固定不可異動）。
- 錯誤統一 Connect code；列表分頁 `per_page ≤ 100`。

## 待辦 Task 詳情

### Task 1: metadicts schema、migration、系統預設 seed（細部 2.5.1）
建立 `metadicts` 表（公司/部門層級可見性）＋ migration ＋ 系統預設字典 seed（冪等）。

### Task 2: MetadictService CRUD 與軟刪除（細部 2.5.2）
MetadictService CRUD、軟刪除/復原、部門擴充與系統預設合併查詢。

### Task 3: 合併查詢與 ListOptions（細部 2.5.3–2.5.4）
系統預設 + 部門擴充合併查詢邏輯、ListOptions 篩選/分頁。

### Task 4: audit_logs schema 與 migration（細部 2.6.1）
`audit_logs` schema（actor/resource/before/after/action）＋ migration（immutable）。

### Task 5: audit.Recorder 的 DB 實作（細部 2.6.2）
接管 01-auth 的 `audit.NoopRecorder` 為 DB 實作；recorder 僅接受交易內 Ent client（杜絕脫鉤寫入，D18）。

### Task 6: 稽核查詢 API（細部 2.6.3）
稽核日誌查詢 API（分頁、時間窗、篩選；預設近 3 個月，D27）。

---

*最後更新：2026-09-18（03-metadicts-audit 現況對齊重建；領域未實作）*
