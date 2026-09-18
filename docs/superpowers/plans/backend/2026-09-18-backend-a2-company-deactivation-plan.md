# Backend 02 — A2 公司停用連鎖登入阻斷 執行計畫（細部 2.1.3）

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 公司 `status` 非 `active` 時，該公司所有帳號（員工＋客戶）無法登入；已登入 session/token 於下一次請求失效（`unauthenticated`）；`developer` 豁免。

**Architecture:** 兩個阻斷點 + 一項稽核：
1. `identityFor`（authzMiddleware 逐請求）：載入使用者時帶公司 edge，若公司非 active 且非 developer → `ok=false`（零值身分 → `authorizeRPC` 回 `unauthenticated`；Web session 由 middleware 銷毀）。
2. `Login`（客戶帳密）：載入使用者帶公司 edge，公司非 active → `permission_denied`（不核發憑證）。
3. `CompanyService.Update`：status 變更（停用/啟用）與 `audit.Record` 同一交易（D18）。

RLS 接線（D3 每請求交易的 `ApplyRLS` 啟用）為獨立設計任務，不在本波（`00007/00010/00011` policy 維持僅定義）。

**Tech Stack:** Go / ent（`WithCompany` edge）/ Connect-RPC / PostgreSQL / enttest(sqlite)。

**Spec:** `docs/superpowers/plans/backend/detail/02-tenancy-users.md` §2.1.3；`docs/superpowers/specs/1.0-requirements/multi-tenancy/spec.md`「公司停用連鎖」；02 計畫 Task 1 Step 3。

**狀態基準：** 2026-09-18。`identityFor` 已查 `u.Status==active` 但**未查公司 status**；Login 未載入公司；`CompanyService.Update` 未寫稽核。

> **✅ 執行結果（2026-09-18）**：Task 1–3 全數完成並 commit（`development`）。交付：`RLSScope.CompanyActive` + middleware 逐請求阻擋（`unauthenticated`,不解銷 session,開發者豁免,恢復 active 可續用）；Login 停用公司 `permission_denied`；`CompanyService.Update` status 變更寫稽核（D18）。RLS D3 接線（`ApplyRLS` 啟用）為獨立設計任務,未在此波。

---

## Global Constraints

- 繁中註解與 commit message；測試僅 stdlib `testing` + enttest(sqlite)；提交前 `task check`。
- 已登入請求所屬公司非 active → `unauthenticated`（Web 清除 session cookie）。
- 停用公司帳號登入 → `permission_denied`（公司已停用），不核發憑證。
- 公司於查詢期間被軟刪除 → 視同停用 → `unauthenticated`。
- `developer` 帳號不受公司停用阻斷；其餘角色無例外。
- 公司 status 變更與稽核同一交易（D18）。

---

## File Structure

- `backend/internal/server/server.go` — `identityFor` 加公司 status 檢查（`u.Role != "developer" && company 非 active → ok=false`）
- `backend/internal/handlers/auth_handler.go` — `Login` 載入公司 edge + 非 active → `permission_denied`
- `backend/internal/services/company_service.go` — `Update` status 變更時寫稽核（D18）
- 測試：`backend/internal/server/server_test.go`（middleware 阻斷）、`backend/internal/handlers/auth_handler_test.go`（登入阻斷）、`backend/internal/services/company_service_test.go`（稽核）

---

## Task 1: middleware 逐請求阻斷（TDD）
- [ ] `identityFor`：載入 user 已帶 `WithCompany`；加條件：`u.Role != "developer" && u.Edges.Company != nil && u.Edges.Company.Status != company.StatusActive` → `ok=false`
- [ ] 測試（server_test.go）：停用公司 + active 使用者 → identityFor ok=false；`authzMiddleware` 該請求不注入身分且 Web session 被銷毀；developer 不受影響
- [ ] `task check` → commit

## Task 2: Login 阻斷（TDD）
- [ ] `Login`：查詢改 `.WithCompany()`；加條件：公司非 active → `permission_denied`（公司已停用），不核發憑證
- [ ] 測試（auth_handler_test.go）：停用公司客戶登入 → `permission_denied`；恢復 active → 可登入
- [ ] `task check` → commit

## Task 3: CompanyService.Update 稽核（D18）+ 文件對齊 + 複審
- [ ] `Update`：status 變更（含停用/啟用）與 `audit.Record`（action `update`, resource_type `company`）同一交易
- [ ] 複審（requesting-code-review 內聯）；02 計畫 Task 1 停用連鎖 Step 3 ✅、README 02 對齊
- [ ] `task check` 全綠 → commit

---

*最後更新：2026-09-18（A2 公司停用連鎖）*
