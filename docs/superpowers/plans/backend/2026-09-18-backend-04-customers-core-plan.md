# Backend 04 — customers 主檔核心 執行計畫（Task 1 + Task 2 批）

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 落地 `customers` 主檔核心：schema、`customer_code` 取號（D7，公司前綴＋6 位自增、樂觀鎖 counter）、CRUD＋軟刪除＋復原＋關鍵字篩選＋字典/業務 reference 驗證。

**Architecture:** 三張表（`company.customer_code_prefix` 前置欄位、`customers`、`customer_counters`）+ Connect-RPC `CustomerService`（List/Get/Create/Update/Delete/Restore）。取號與建檔同一交易（D18）；`customer_code` 建立後不可改（D7）；字典欄位（payment/settlement/customer_type/invoice_type）驗證指向 `metadicts`、`default_sales_rep_id` 驗證指向同公司同部門有效業務；軟刪除＋部分唯一索引（D10）。

**Tech Stack:** Go / ent / Connect-RPC（buf）/ PostgreSQL（goose）/ `shopspring/decimal`（後續換算才用，本批未用）/ enttest(sqlite)。

**Spec:** `docs/superpowers/plans/backend/detail/04-master-data.md` §3.1.1–3.1.3；決策 D7/D10/D18。

**狀態基準：** 2026-09-18。`customers`/`customer_counters` 無 code；`company` 缺 `customer_code_prefix`。

> **✅ 執行結果（2026-09-18）**：Task 1–5 全數完成並 commit（`development`）。交付：`company.customer_code_prefix` + `customers`/`customer_counters` schema + migration `00013`；`CustomerService` 6 RPC（List/Get/Create/Update/Delete/Restore）+ 取號 counter（D7 樂觀鎖、取號+建檔+稽核同交易）+ 字典/業務驗證 + `include_deleted`/關鍵字/排序白名單。Task 3（D22 帳號連動）留後續批（需 user 增欄位）。

---

## Global Constraints

- 繁中註解與 commit message；測試僅 stdlib `testing` + enttest(sqlite)；提交前 `task check`。
- `customer_code` = `companies.customer_code_prefix`（大寫英數 1–4）＋ 6 位補零自增；公司內唯一、建立後不可改（D7）。
- 取號（樂觀鎖 counter）+ 建客戶 + 稽核＝同一 DB 交易（D18）；重試上限 5。
- 軟刪除 `deleted_at` + 部分唯一索引 `(company_id, customer_code) WHERE deleted_at IS NULL`（D10）；列表預設排除已刪除、`include_deleted` 可開。
- 字典欄位驗證：指向 `metadicts` 且 `is_active`、type 相符、未刪除；`default_sales_rep_id` 為同公司同部門、業務身分、active 未刪的 user。
- 錯誤：未登入/角色不符/跨部門 → `unauthenticated/permission_denied`；不存在/已刪 → `not_found`；必填缺失/字典非法/試改 code/排序白名單外 → `invalid_argument`；counter 衝突防禦 → `already_exists`；樂觀鎖重試超限 → `failed_precondition`。
- 排序白名單：`name` / `customer_code` / `created_at`。
- 主檔管理開放 `dept_admin`/`staff`（限所屬部門）；`customer` 主帳號對業務 API 一律 `permission_denied`。

---

## File Structure

- `backend/ent/schema/company.go` — 增 `customer_code_prefix`(String, Optional, 大寫英數 1–4)
- `backend/ent/schema/customer.go` — `customers` 實體
- `backend/ent/schema/customercounter.go` — `customer_counters` 實體（company_id 主鍵 / next_seq / version 樂觀鎖）
- `backend/ent/` — codegen
- `backend/database/migrations/00013_customers.sql` — company 前置欄位 + `customers` + `customer_counters` + 索引 + RLS policy（僅定義不 ENABLE）
- `backend/proto/customers/v1/customer.proto` — `CustomerService`（List/Get/Create/Update/Delete/Restore）+ message
- `backend/internal/proto/customers/v1/` — buf 產碼
- `backend/internal/services/customer_service.go` — CRUD + counter 取號 + 驗證 + 稽核
- `backend/internal/services/customer_service_test.go` — 單元測試
- `backend/internal/server/domains.go` — `RegisterCustomerServices`

---

## Task 1: company.customer_code_prefix + customers/customer_counters schema + migration 00013（TDD 編譯層）

- [ ] `company.go` 增 `customer_code_prefix`；`customer.go`（company_id、department_id 可空、customer_code、name、tax_id 可空、四字典 id 可空、default_sales_rep_id 可空、preferred_delivery_days JSON、promo_tag_ids JSON、created_by/updated_by 可空、time 欄位、軟刪除）；`customercounter.go`（company_id 主鍵、next_seq、version）
- [ ] `go generate ./ent`；migration `00013_customers.sql`（含部分唯一索引、查詢索引、RLS policy 僅定義）
- [ ] `task check` 全綠 → commit

## Task 2: proto + codegen
- [ ] `customers/v1/customer.proto`：`CustomerService`（ListCustomers/GetCustomer/CreateCustomer/UpdateCustomer/DeleteCustomer/RestoreCustomer）+ `Customer` message（含可空欄位用 wrapper）
- [ ] buf generate（Go）；`buf lint`/`go build` PASS → commit

## Task 3: CustomerService CRUD（TDD）
- [ ] 測試先行：List 關鍵字/分頁/排序白名單/跨部門拒絕；Get 不存在→not_found；Create 字典/業務驗證＋code 不可自帶；Update 試改 code 拒絕；Delete/Restore 軟刪除＋稽核
- [ ] 實作 `customer_service.go`（範圍推導、驗證、交易＋稽核）→ `task check` → commit

## Task 4: customer_code 取號（D7，TDD）
- [ ] 測試先行：`customer_counters` 樂觀鎖取號（前綴＋補零、同交易遞增、併發 version 衝突重試 5 次、取號成功但建檔失敗回滾不遞增）
- [ ] 實作 `CreateCustomer` 接入 counter（取號＋建客戶＋稽核同一交易；公司無前綴/停用→`failed_precondition`）→ `task check` → commit

## Task 5: 掛載 + 複審 + 文件對齊
- [ ] `domains.go` `RegisterCustomerServices`；複審（requesting-code-review 內聯）；04 計畫 Task 1/2 勾✅、README 04 反映
- [ ] `task check` 全綠 → commit

---

*最後更新：2026-09-18（04 customers 核心批）*
