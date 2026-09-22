# Backend 02 — 多租戶與使用者 執行計畫（現況對齊版）

> **性質**：原為目標型執行計畫（含內嵌目標程式碼）。經 2026-09-18 以實際程式碼盤點（codebase-memory 知識圖譜 + git）重建，改為**反映現況的執行計畫**。
>
> **對應設計**：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）、決策 `D3/D6/D9/D22/D28`
> **細部文件**：`docs/superpowers/plans/backend/detail/02-tenancy-users.md`
> **狀態基準**：2026-09-18 盤點（知識圖譜 + git + backend code）

---

## 執行狀態總覽

| Task | 內容 | 狀態 | 實際產物 |
|---|---|---|---|
| 1 | Company schema 擴充 + CompanyService CRUD + 唯一性 + 停用連鎖 | ✅ 完成（2026-09-18, A2）| `internal/services/company_service.go`（CRUD + status 變更稽核 D18）＋ `server.go` identityFor/middleware 阻斷（2.1.3）|
| 2 | 部門管理 API | ✅ 完成 | `company_service.go`（Department CRUD）、`ent/schema/department.go` |
| 3 | 使用者 CRUD + 角色指派 + 停用 + ForceLogout | 🟡 部分 | `user_service.go`、`user.proto`；AssignRole/Deactivate/ForceLogout 含 D18 稽核 + tv+1；主帳號連鎖(D22)待 Phase 3 |
| 4 | Logo/Branding/PublicInfo/公開發現端點 | 🟡 部分 | PublicInfo 欄位序列化已做；Logo 上傳已落地（2026-09-22：`fileassets.logo`＋`GET /me`＋Web 上傳/側邊欄顯示；權限為 **company_admin 限所屬公司**，spec 3.1.1 已同步修訂）；殘：UpdateBranding／公開發現端點 |
| 5 | roles + role_permissions schema + RoleService CRUD | ✅ 完成 | `role_service.go`、migration `00003`、`role.proto` |
| 6 | 功能權限矩陣 + GetAbility 表驅動 + RLS data_scope 注入 | 🟡 部分 | ability 表驅動已做；data_scope 注入待 |
| 7 | Casbin policy 管理 API + 防鎖死 + ListGrouping | 🟡 部分 | 防鎖死/條件驗證已做；Casbin policy 管理 API 待 |

---

## Task 1: Company schema 擴充 + CompanyService CRUD + 唯一性 + 停用連鎖（細部 2.1.1–2.1.3）

**實際狀態：🟡 部分完成**

**實際產物（已存在）：**
- `backend/internal/services/company_service.go` — `CompanyService`：`ListCompanies` / `GetCompany` / `CreateCompany` / `UpdateCompany` / `DeleteCompany`
- `backend/proto/salesorder/v1/company.proto` — `CompanyService`（5 RPC）
- `backend/ent/schema/company.go`

**說明**：Company CRUD 已實作並掛載 Connect（`requireScope` 做資源/動作授權）。**待補**：停用連鎖（停用公司 → 連鎖部門/使用者）、`rls.Identity` 擴充、登入端點檢查被停用公司（2.1.3）。

- [x] **Step 1: Company CRUD** — 5 個 RPC 全數實作
- [x] **Step 2: proto CompanyService** — company.proto / company.pb.go
- [x] **Step 3: 停用連鎖（2.1.3）** — 停用公司 → middleware 逐請求阻擋（unauthenticated, 保留 session）+ 登入端點 `permission_denied` + developer 豁免（2026-09-18, A2 完成）

**待辦摘要**：停用連鎖與登入端點檢查已完成（A2）。

---

## Task 2: 部門管理 API（細部 2.2.1）

**實際狀態：✅ 完成**

**實際產物（已存在）：**
- `company_service.go` — `DepartmentService`：`ListDepartments` / `GetDepartment` / `CreateDepartment` / `UpdateDepartment` / `DeleteDepartment`
- `ent/schema/department.go`、`company.proto`（DepartmentService）

**說明**：部門 CRUD 已實作並掛載。

- [x] **Step 1: Department CRUD** — 5 個 RPC 全數實作
- [x] **Step 2: proto DepartmentService** — company.proto / department 產碼

---

## Task 3: 使用者 CRUD + 角色指派 + 停用 + ForceLogout 範圍銜接（細部 2.3.1–2.3.3）

**實際狀態：🟡 部分完成（2026-09-18 UserService 已落地）**

**實際產物（已存在）：**
- `backend/proto/salesorder/v1/user.proto` + `internal/proto/salesorder/v1/user.{pb.go,connect.go}`（UserService 7 RPC）
- `backend/internal/services/user_service.go` — ListUsers/GetUser/CreateUser/UpdateUser/AssignRole/Deactivate/ForceLogout
- `backend/internal/server/server.go`（protectedRPC 加 7 條 user path）、`backend/internal/server/domains.go`（註冊掛載）
- `backend/internal/audit/recorder.go` + `ent/schema/auditlog.go` + migration `00009`（稽核地基,D18）

**說明**：UserService CRUD + 範圍控制（super 全域/company_admin 公司/dept_admin 部門 staff）+ AssignRole（含 guest 審核）+ Deactivate + ForceLogout 已實作；角色指派/停用/強登皆**同一交易寫稽核（D18）**並 `token_version+1`（D5），AssignRole 另同步 OpenFGA assigned tuple。

- [x] **Step 1: 使用者 CRUD** — `UserService`（List/Get/Create/Update、範圍）＋ user.proto 已實作
- [x] **Step 2: 角色指派 + 停用** — AssignRole（含 guest 審核 / D18 稽核 / tv+1）、Deactivate（D18 稽核 / tv+1）已實作
- [x] **Step 3: ForceLogout 範圍銜接** — ForceLogout（含不能對自己、D18 稽核 / tv+1）已實作

**待辦摘要**：**主帳號連鎖子帳號（D22）未實作**——屬客戶帳號管理（一主多子）範疇，為 Phase 3/04 客戶主檔流程；本 Task 3 聚焦員工帳號管理。01 計畫 Task 7（首登強改密碼）與 01 Task 11（X-Api-Token）亦尚未落地（另列於 01 計畫）。

---

## Task 4: Logo 上傳 / Branding / PublicInfo / 公開發現端點（細部 2.4.1–2.4.3）

**實際狀態：🟡 部分完成**

**實際產物（已存在）：**
- `company_service.go` — `PublicInfo` 欄位於 `companyToProto` 序列化（`structpb.NewStruct`）
- `ent/schema/company.go` — `public_info`、`capabilities` 欄位

**說明**：`public_info` / `capabilities` 欄位與序列化已就緒；**Logo 檔案上傳已落地（2026-09-22）**——`domain/fileassets/logo.go` REST `POST /api/v1/companies/{company_id}/logo`（**company_admin 限所屬公司**、白名單副檔名＋magic bytes 三重驗證、檔記錄＋`companies.logo_url`＋稽核同交易、舊檔保留），`GET /api/v1/me` 回身分與公司品牌，Web 端 CompaniesPage 上傳對話框（company_admin 於本公司列顯示）＋側邊欄 `BrandLogo` 顯示。**仍未含**：Branding（`UpdateBranding`）、公開發現端點（`/api/v1/companies/public/{identifier}`）。

- [x] **Step 1: PublicInfo/capabilities 欄位** — company schema + 序列化
- [x] **Step 2: Logo 上傳** — FileStore（04 計畫）＋ `companies.logo_url` 更新（`fileassets.logo`；權限為 company_admin 限所屬公司，spec 3.1.1 已同步修訂；驗收見 04 計畫 Task 3.6 與 `logo_integration_test`：401／super 403／dept_admin 403／400／他公司與不存在皆 404／成功＋url 可下載／軟刪 404 八段）
- [ ] **Step 3: 公開發現端點** — `/api/v1/companies/public/{identifier}`

---

## Task 5: roles + role_permissions schema 與 seed + RoleService CRUD（細部 2.9.1–2.9.2）

**實際狀態：✅ 完成**

**實際產物（已存在）：**
- `backend/internal/services/role_service.go` — `ListRoles` / `GetRolePermissions` / `UpdateRolePermissions` / `ListConditionFields`
- `backend/ent/schema/{role.go,rolepermission.go}`
- migration `backend/database/migrations/00003_role_permissions_casl.sql`
- `backend/proto/salesorder/v1/role.proto`
- git commit `a60b0d4`（T18）、`c204b26`、`ef1bf04`

**說明**：7 內建角色 + role_permissions（CASL 三欄 + 條件唯一鍵）schema/migration、RoleService CRUD、防鎖死與條件驗證、company_admin 限自己公司皆已實作。

- [x] **Step 1: role/rolepermission schema** — ent schema + 產生碼
- [x] **Step 2: migration seed** — 00003（冪等）
- [x] **Step 3: RoleService CRUD** — ListRoles/GetRolePermissions/UpdateRolePermissions/ListConditionFields
- [x] **Step 4: 防鎖死 + 條件驗證 + 範圍** — validateNoLockout / validateConditions / validateOwnCompany

---

## Task 6: 功能權限矩陣 + GetAbility 表驅動 + RLS data_scope 注入（細部 2.9.3–2.9.5）

**實際狀態：🟡 部分完成**

**實際產物（已存在）：**
- `backend/internal/domain/auth/ability.go` — `BuildAbility` 已**表驅動**（`Where(rolepermission.HasRoleWith(role.CodeIn(id.Roles...)))` 依 sort_order）
- `role_service.go` — `ListConditionFields`
- frontend RolesPage / PermissionMatrix（git `0c51977`，T19）

**說明**：GetAbility 已由 `role_permissions` 表驅動（完成 2.9.4）；功能權限矩陣前端已做。**未含** RLS `data_scope` 注入（user 無 `data_scope` 欄位，相依 01 Task 1 Step 3 補齊後才能注入）。

- [x] **Step 1: GetAbility 表驅動** — `BuildAbility` 使用 rolepermission
- [x] **Step 2: 功能權限矩陣** — RolesPage / PermissionMatrix 前端
- [ ] **Step 3: RLS data_scope 注入** — 依 user.data_scope 於 middleware/Identity 展開（相依 01 Task 1/11）

**待辦摘要**：data_scope 注入（相依 auth 地基補齊）。

---

## Task 7: Casbin policy 管理 API + 預設 p 規則 seed + 防鎖死 + ListGrouping（細部 2.10.1–2.10.4）

**實際狀態：🟡 部分完成**

**實際產物（已存在）：**
- `role_service.go` — 防鎖死（`validateNoLockout`）、`ListConditionFields`
- Casbin 執行層（01 Task 2）

**說明**：防鎖死與條件驗證邏輯已做。**未含**完整 Casbin policy 管理 API（`policies` domain、p 規則 CRUD 即時生效）、預設 p 規則 seed 的 DB 化（01 Task 2 目前用 string-adapter）、`ListGrouping`。

- [x] **Step 1: 防鎖死 + 條件驗證** — validateNoLockout / validateConditions
- [ ] **Step 2: Casbin policy 管理 API** — policies CRUD 即時生效（相依 01 Task 2 遷 DB adapter）
- [ ] **Step 3: 預設 p 規則 seed（冪等）**
- [ ] **Step 4: ListGrouping**

**待辦摘要**：Casbin policy 管理 API（相依 01 Task 2 遷 DB adapter）。

---

*最後更新：2026-09-18（02-tenancy-users 現況對齊重建，以 codebase-memory 驗證）*
