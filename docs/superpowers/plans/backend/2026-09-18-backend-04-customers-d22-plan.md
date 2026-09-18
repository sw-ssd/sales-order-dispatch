# Backend 04 — customers D22 建檔連動帳號 執行計畫（04 Task 3，子功能 3.1.4）

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立客戶時於同一交易自動建立 1 主帳號（角色 `customer`、`is_primary=true`）＋1 業務子帳號（`is_primary=false`、`system_generated=true`），各發一組 24h 效期隨機臨時密碼、`must_change_password=true`；`CreateCustomerResponse` 擴充回傳兩帳號名稱＋臨時密碼＋`account_manage_url`（規格 §9.4 憑證交付）。

**Architecture:** `users` 增 `customer_id`（可空 FK→customers）、`is_primary`、`system_generated`（migration 00014）；`CreateCustomer` 複合交易內取號→建客戶→建主帳號→建業務子帳號→寫稽核（客戶＋兩帳號）同成功同失敗（D18）。臨時密碼重用 `auth.GenerateTempPassword` / `auth.HashPassword`；`account_manage_url` 以 `config.Auth.FrontendURL`＋`/customer_account_manage` 組出（經 domains.go 注入）。

**Tech Stack:** Go / ent / Connect-RPC（buf）/ PostgreSQL（goose）/ enttest(sqlite)。

**Spec:** `docs/superpowers/plans/backend/detail/04-master-data.md` §3.1.4（D22）；決策 D7/D10/D18/D22；密碼機制造 `01-auth.md` §1.5.2/1.5.4。

**狀態基準：** 2026-09-18。customers 核心批（Task 1–2）已 commit；`user` schema 缺 `customer_id/is_primary/system_generated`；`CreateCustomerResponse` 未含帳號交付欄位；`CustomerService` 未注入 FrontendURL。

---

## Global Constraints

- 繁中註解與 commit message；測試僅 stdlib `testing` + enttest(sqlite)；提交前 `task check`。
- 建檔連動帳號與取號/建客戶/稽核同一 DB 交易，任一失敗整交易回滾（D18），不產生孤兒客戶/帳號。
- 每客戶恆恰一 `is_primary=true` 帳號；資料庫以部分唯一索引 `(customer_id) WHERE is_primary=true AND customer_id IS NOT NULL` 兜底。
- 主帳號 `system_generated=false`；業務子帳號 `system_generated=true`、`is_primary=false`（灰化判斷：店家對系統自動帳號改名/停用/重置一律拒絕，由 01/02 清單與拒絕邏輯後續實作，本批只寫入可識別標記）。
- 兩帳號皆角色 `customer`、`is_customer=true`、`must_change_password=true`、`temp_password_expires_at=now+24h`；兩組臨時密碼相異。
- `user.email` 全域唯一 → 以「全域唯一之 customer_code」生成佔位 email（主 `customer.{code}@system.local`、子 `salesrep.{code}@system.local`），保證不撞既有帳號。
- `default_sales_rep_id` 於 `CreateCustomer` 改為**必填**（D22 交付流程必要；未提供/非法 → `invalid_argument`）。
- 錯誤：重複帳號名（防禦）`already_exists`；`default_sales_rep_id` 缺失/非法 `invalid_argument`；users 層約束違反映射 `already_exists/invalid_argument`。

---

## File Structure

- `backend/ent/schema/user.go` — 增 `customer_id`(Int, Optional/Nillable)、`is_primary`(Bool, Default false)、`system_generated`(Bool, Default false)
- `backend/ent/` — codegen（`go generate ./ent`）
- `backend/database/migrations/00014_user_customer_account.sql` — `ALTER TABLE users ADD COLUMN IF NOT EXISTS customer_id bigint / is_primary boolean NOT NULL DEFAULT false / system_generated boolean NOT NULL DEFAULT false` + FK + 查詢索引 + 部分唯一索引（每客戶一主帳號）
- `backend/proto/customers/v1/customer.proto` — `CreateCustomerResponse` 增 `primary_account_name` / `primary_temp_password` / `sales_rep_account_name` / `sales_rep_temp_password` / `account_manage_url`
- `backend/internal/proto/customers/v1/` — buf 產碼（Go；Dart 已知 gap 不產）
- `backend/internal/services/customer_service.go` — Create 必填 rep；tx 內建兩帳號＋寫稽核；回傳交付欄位
- `backend/internal/server/domains.go` — `RegisterCustomerServices` 注入 `s.cfg.Auth.FrontendURL`
- 測試：`backend/internal/services/customer_service_test.go`（擴充 D22 案例＋既有案例補 rep）

---

## Task 1: user schema + codegen + migration 00014（TDD 編譯層）

- [ ] `user.go` 增三欄；`go generate ./ent`；migration `00014_user_customer_account.sql`
- [ ] `task check` 全綠 → commit

## Task 2: proto + codegen
- [ ] `customer.proto` `CreateCustomerResponse` 增五欄；buf generate（Go）→ lint/build PASS → commit

## Task 3: CreateCustomer 連動建帳號（TDD）
- [ ] 測試先行：D22（回應含兩帳號名/臨時密碼/url；DB 兩帳號欄位正確、24h 效期、must_change 皆 true）；`default_sales_rep_id` 缺→`invalid_argument`；既有案例補 rep
- [ ] 實作 `customer_service.go`（必填 rep、建兩帳號＋稽核、回傳交付欄位）→ `task check` → commit

## Task 4: 掛載 + 複審 + 文件對齊
- [ ] `domains.go` 注入 FrontendURL；複審（requesting-code-review 內聯）；04 計畫 3.1.4 勾✅、README 04 反映、批次計畫登錄
- [ ] `task check` 全綠 → commit

---

## 複審登錄（requesting-code-review, 2026-09-18）

- ✅ 內聯唯讀複審完成（本 harness 無通用子代理派送，採內聯）：`default_sales_rep_id` 於 Create 改為必填；兩帳號與客戶/取號/稽核同一交易（D18），任一 `buildCustomerAccount` 失敗於 commit 前回傳並回滾，不產生孤兒；`user.email` 以全域唯一之 `customer_code` 生成佔位不撞既有帳號；`account_manage_url` 由 `config.Auth.FrontendURL`＋`/customer_account_manage` 經 domains.go 注入。
- **註記 #1（防禦性、可接受）**：規格 3.1.4 步驟 4「帳號名稱於客戶內唯一」未以 DB 唯一索引強制——新建客戶恰一次建立兩帳號且名稱由客戶名稱推導（理論不可能重複），`account_name` 亦非唯一欄位；採服務層保證，不另建索引。
- **註記 #2**：「每客戶恰一 `is_primary=true`」由 migration `00014` 部分唯一索引 `(customer_id) WHERE is_primary=true AND customer_id IS NOT NULL` 兜底（Postgres）；ent/sqlite 測試不建該索引，由服務層正確寫入（測試內驗證）。
- 已知 gap：App Dart `lib/gen` 未同步（本機無 `protoc-gen-dart`）；buf 重產對不相關前端 proto 的版本註解 churn 已還原，僅保留本任務 `customer_pb.*`；A2「App JWT 路徑停用阻斷」併入 01 Task 11 未作。

### 複審 recheck（requesting-code-review, 2026-09-18 第二輪）

- ✅ **已修 A（Important）**：D22 核心契約「交付的臨時密碼須可登入」原測試僅驗非空/相異，未驗明文能對上儲存雜湊。`assertD22Accounts` 加 `auth.VerifyPassword(hash, 回傳明文)` 驗主/子兩帳號，並加「兩帳號密碼不可互相登入」負向斷言。
- ✅ **已修 B（Important）**：補 `TestCreateCustomerAccountFailureRollsBack`——預占主帳號 email 使建帳碰撞失敗，驗客戶列=0、計數器回滾為 1、無孤兒帳號（D18 同交易回滾，規格 3.1.4 驗收項）。
- **Minor 待觀察**：`customerTempPasswordTTL`（services）與 `tempPasswordTTL`（handlers）同值 24h 之常數分置兩套件（可接受，均為規格 1.5.2）。
- **Minor 待辦（依賴後續）**：`is_primary/system_generated` 尚未經 UserService proto 下發店面清單灰化（01/02 帳號管理範圍）。
