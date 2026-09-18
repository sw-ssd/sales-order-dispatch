# Backend 01 — A3 臨時密碼與首登強制修改 執行計畫（1.5.2 + 1.5.4）

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 落地客戶帳號「臨時密碼 + 首登強制改密碼」與密碼重置（dept_admin+ 重新發臨時密碼），含受限登入態（`must_change_password` 時僅 `ChangePassword` 可用）。

**Architecture:** 在 user schema 增 `must_change_password`＋`temp_password_expires_at`（migration 00012）。新增 AuthService RPC `ChangePassword`／`ResetCustomerPassword`；`LoginResponse` 增 `must_change_password`；access JWT `Claims` 增 `MustChangePassword` 旗標，由 `authzMiddleware` 於 `must_change_password=true` 時僅放行 `ChangePassword` path、其餘回 `failed_precondition`。改密碼/重置同一交易寫稽核（D18）＋ `token_version+1` 撤銷舊 session/憑證。

**Tech Stack:** Go / ent / Connect-RPC（buf）/ PostgreSQL / Argon2id（既有 `auth/password.go`）/ enttest(sqlite)。

**Spec:** `docs/superpowers/plans/backend/detail/01-auth.md` §1.5.2、§1.5.4；01-auth 計畫 Task 7/8。

**狀態基準：** 2026-09-18 盤點。user schema 缺 `must_change_password`/`temp_password_expires_at`；auth.proto 無 ChangePassword/Reset；LoginResponse 無 must_change_password；JWT Claims 無受限旗標。

> **✅ 執行結果（2026-09-18）**：Task 1–6 全數完成並 commit（`development`）。交付：migration 00012、ChangePassword/ResetCustomerPassword、`GenerateTempPassword`、受限 claim（Claims.MustChangePassword）＋ authzMiddleware 攔截（僅放行 ChangePassword）、LoginResponse.must_change_password。臨時密碼效期 24h；改/重置皆 `token_version+1` 並稽核（D18）。

---

## Global Constraints

- 繁中註解與 commit message；測試僅 stdlib `testing` + enttest(sqlite)；提交前 `task check`。
- 臨時密碼：隨機 ≥ 12 字元；`temp_password_expires_at = now + 24h`；`must_change_password = true`；同時清鎖定（Valkey lockout）；同一交易。
- `must_change_password=true` 時登入所發 JWT 帶受限 claim，middleware 僅放行 `ChangePassword`，其餘 `failed_precondition`。
- `ChangePassword`：驗舊密碼 → 新密碼強度 ≥ 8 字元 → 更新雜湊、清 `must_change_password`/`temp_password_expires_at`、`token_version+1`、稽核（action `update`，不含密碼）→ 交易。
- `ResetCustomerPassword`：dept_admin 限自己部門、company_admin 限公司、super 不限；目標 `is_customer=true`；回傳明文僅本次回應不落盤；稽核含操作者與目標。
- 錯誤：舊密碼錯/帳密錯/停用 → `unauthenticated`；過期 → `failed_precondition`；強度不足 → `invalid_argument`；範圍外 → `permission_denied`；目標非客戶 → `invalid_argument`。
- RLS 沿用既有定義（不新增 policy）；不觸動 OpenFGA。

---

## File Structure

- `backend/ent/schema/user.go` — 增 `must_change_password`(Bool, Default false)、`temp_password_expires_at`(Time, Optional/Nillable)
- `backend/ent/` — codegen（`go generate ./ent`）
- `backend/database/migrations/00012_user_temp_password.sql` — `ALTER TABLE users ADD ...` + 註解（冪等：`ADD COLUMN IF NOT EXISTS`）
- `backend/proto/salesorder/v1/auth.proto` — `LoginResponse.must_change_password`、`ChangePassword`/`ResetCustomerPassword` RPC + request/response
- `backend/internal/proto/salesorder/v1/` — buf 產碼
- `backend/internal/auth/password.go` — `GenerateTempPassword()`（≥12 隨機）
- `backend/internal/auth/token.go` — `Claims.MustChangePassword` 旗標 + `TokenSubject` 欄位
- `backend/internal/handlers/auth_handler.go` — Login 回 `must_change_password`；`ChangePassword`/`ResetCustomerPassword` handler（交易＋稽核）
- `backend/internal/server/server.go` — authzMiddleware 受限 claim 接線（僅放行 `ChangePassword` path）
- 測試：`backend/internal/handlers/auth_handler_test.go`（擴充）、`backend/internal/auth/token_test.go`、`backend/internal/services/user_service_test.go`（若 issueTempPassword 由 UserService 建檔用）

---

## Task 1: user schema + codegen + migration 00012

- [ ] schema 增兩欄 → `go generate ./ent` → migration `00012_user_temp_password.sql`（`ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT false; ADD COLUMN IF NOT EXISTS temp_password_expires_at timestamptz`)
- [ ] `task check` 全綠；commit

## Task 2: proto + codegen

- [ ] `LoginResponse` 增 `bool must_change_password = 4`；`AuthService` 增：
  - `ChangePassword(ChangePasswordRequest{old_password, new_password}) returns (ChangePasswordResponse)`
  - `ResetCustomerPassword(ResetCustomerPasswordRequest{user_id}) returns (ResetCustomerPasswordResponse{temp_password, expires_at})`
- [ ] buf generate（`internal/proto` Go）→ lint/build PASS → commit

## Task 3: 登入態受限 claim + middleware 接線（TDD）

- [ ] `Claims`/`TokenSubject`/`IssueAccess` 增 `MustChangePassword`；Login 核發時帶入
- [ ] `authzMiddleware`：身分 `MustChangePassword=true` 時，非 `ChangePassword` 受保護 path → `failed_precondition`（`ChangePassword` 路徑放行）
- [ ] token_test 覆蓋 claim 旗標；server 測試覆蓋受限接線；`task check` → commit

## Task 4: ChangePassword + 首登強改（TDD）

- [ ] `GenerateTempPassword()`（≥12 隨機字元）
- [ ] `ChangePassword` handler：驗舊→新密碼 ≥8→更新 hash/清旗標與效期/token_version+1→稽核(action `update`, resource_type `user`)→同交易
- [ ] 測試：改前其它 RPC 被拒、改後可正常、舊密碼錯 `unauthenticated`、強度不足 `invalid_argument`、稽核與 token_version 正確
- [ ] `task check` → commit

## Task 5: ResetCustomerPassword（TDD）

- [ ] `ResetCustomerPassword` handler：範圍推導（dept/company/super）、目標 `is_customer`、`GenerateTempPassword` + 更新 + token_version+1 + 稽核（操作者+目標）同交易
- [ ] 測試：範圍外 `permission_denied`、非客戶 `invalid_argument`、重置後舊 session 失效 + 新臨時密碼可登入且強制改
- [ ] `task check` → commit

## Task 6: 複審 + 文件對齊 + 收尾

- [ ] 複審（requesting-code-review 內聯）；README 01/02 反映 Task 7/8 落地；01-auth 計畫 Task 7→✅、Task 8 重置→✅
- [ ] `task check` 全綠 → commit

---

*最後更新：2026-09-18（A3 臨時密碼/首登強改）*
