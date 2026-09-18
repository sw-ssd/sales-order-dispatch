# Backend 03 — AuditService 稽核查詢 API 執行計畫（A4 / 細部 2.6.3）

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 落地 `AuditService.ListAuditLogs` 稽核日誌查詢 API（分頁、時間窗、行為/資源/操作者篩選、操作者名稱 join、範圍限定）。

**Architecture:** 既有 `audit_logs` 表（02 批次已建，00009/00010）。新增 `audit/v1` proto + Connect service（放 `internal/services`，reuse `parseID`/`toConnectError`/`normalizePage`）；範圍：super 全域＋可選 company、company_admin 僅自己公司、其餘角色拒絕；未帶時間窗時套用近 3 個月（D27）；時間降冪；操作者名稱以 user_id 一次 join users。

**Tech Stack:** Go / ent（`audit_logs`）/ Connect-RPC（buf）/ PostgreSQL / enttest(sqlite)。

**Spec:** `docs/superpowers/plans/backend/detail/03-metadicts-audit.md` §2.6.3；決策 D18、D27。

---

## Global Constraints

- 一律繁中註解與 commit message；測試僅 stdlib `testing` + enttest(sqlite)；提交前 `task check`。
- 範圍：super(data_scope=all) 全系統、可選 `company_id` 篩選；company_admin 強制自己公司（忽略請求帶的他公司）；其餘角色 `permission_denied`。
- 篩選皆 AND；排序 `created_at` 降冪；分頁 `meta`、`per_page ≤ 100`。
- 未帶任何時間篩選時套用預設近 3 個月時間窗（D27）,防全表掃描。
- `audit_logs` 不可由 API 寫入（只查詢）；查詢本身不寫稽核（避免自我遞迴）。

---

## Task 1: proto + codegen
- Create `backend/proto/audit/v1/audit.proto`（`AuditService.ListAuditLogs`、`AuditLog` message、`ListAuditLogsRequest/Response`、`salesorder.v1.Pagination`）
- buf generate（僅 Go, template 產至 `internal/proto`）
- `buf lint` PASS、`go build` PASS

## Task 2: AuditService 實作（TDD）
- Create `backend/internal/services/audit_service.go`
  - 範圍推導 `isSuperIdentity` / `isCompanyAdmin`；company_admin fail-closed 缺公司 → `permission_denied`
  - 時間窗：`from`/`to`（RFC3339 驗證,非法 → `invalid_argument`）；皆空套用近 3 個月
  - 篩選：`action`（驗證合法值）、`resource_type`+`resource_id`、`user_id`
  - 分頁 + `created_at` 降冪；操作者名稱一次 join `users`
- Create `backend/internal/services/audit_service_test.go`：company_admin 跨公司被強制、filter、時間窗、名稱 join、分頁、非法輸入
- Modify `backend/internal/server/domains.go`：`RegisterAuditServices(apiMux, entClient)`

## Task 3: 複審 + 文件對齊
- 複審（requesting-code-review 內聯）；README 03 → ✅ 完成、03 計畫 Task 6 → ✅
- `task check` 全綠

---

*最後更新：2026-09-18（A4 稽核查詢 API）*
