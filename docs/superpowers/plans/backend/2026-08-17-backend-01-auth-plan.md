# Backend 01 — 認證與授權地基 執行計畫（現況對齊版）

> **性質**：本文件原為「目標型」執行計畫（v 初版，含內嵌目標程式碼）。經 2026-09-18 以實際程式碼盤點重建，改為**反映現況的執行計畫**：每個 Task 標示實際實作狀態、對應真實產物檔案路徑，未完成部分保留為待辦。
>
> **對應設計**：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）、決策 `D5/D6/D7/D8/D9`
> **細部文件**：`docs/superpowers/plans/backend/detail/01-auth.md`
> **狀態基準**：2026-09-18 盤點（git + backend code）
> **覆寫備註**：D32 於主計畫標頭指示「Task 14 改 OpenFGA + RLS」，但**實際實作仍為 Casbin**（`internal/auth/casbin.go` 註解亦採 Casbin）。本文件以實際 code 為準，OpenFGA 遷移列為待辦（見 Task 14）。

---

## 執行狀態總覽

| Task | 內容 | 狀態 | 實際產物 |
|---|---|---|---|
| 1 | Company/Department/User Ent schema | 🟡 部分 | `ent/schema/{company,department,user}.go`（user 欄位精簡） |
| 2 | Casbin model + enforcer | ✅ 完成 | `internal/auth/casbin.go`、`internal/auth/{rbac_model.conf,rbac_policy.csv}` |
| 3 | RLS policies + 注入 | 🟡 部分 | `internal/auth/rls.go`、migration `00002_rls_policies.sql`（無 WrapDriver 注入） |
| 4 | OAuth2 導向與 callback | ✅ 完成 | `internal/auth/oidc.go`、`handlers/auth_handler.go` |
| 5 | 註冊完成與 guest 審核 | 🟡 部分 | `RegisterComplete` handler |
| 6 | 客戶帳密登入 | ✅ 完成 | `internal/auth/password.go`、`handlers/auth_handler.go` |
| 7 | 臨時密碼與首登強制修改 | ✅ 完成（2026-09-18, A3）| `internal/handlers/auth_password.go`、migration `00012`（must_change/temp 效期）、受限 claim + middleware 攔截 |
| 8 | 登入鎖定與密碼重置 | ✅ 完成（2026-09-18, A3）| `internal/auth/password.go`（鎖定）+ `ResetCustomerPassword`（auth_password.go）|
| 9 | scs+Valkey session 與 access JWT | ✅ 完成 | `internal/auth/{session,token,stores}.go` |
| 10 | refresh 旋轉與 token_version | ✅ 完成 | `internal/auth/token.go`、migration `00004` |
| 11 | Authenticate 與 X-Api-Token middleware | ⬜ 未開始 | —（無 middleware 目錄） |
| 12 | 強制登出 | 🟡 部分 | `token.go`（BumpTokenVersion）、`Logout` handler |
| 13 | ability API | ✅ 完成 | `internal/domain/auth/ability.go`、`proto/ability.proto` |
| 14 | developer 逃生門與 audit 介面 | ⬜ 未開始 | —（無 audit.Recorder、無 developer bypass） |

**實作範圍說明**：本計畫涵蓋的認證/授權地基已實作 auth 主路徑（OIDC 登入、客戶帳密登入、JWT/refresh、session、Casbin RBAC、RLS 語句、ability API、role 權限 CRUD）。未含 middleware、developer/audit、首登強改密碼等後段項目。

---

## Task 1: Company / Department / User Ent schema（細部 1.1.1–1.1.3）

**實際狀態：🟡 部分完成**

**實際產物（已存在）：**
- `backend/ent/schema/company.go`
- `backend/ent/schema/department.go`
- `backend/ent/schema/user.go`

**與計畫差異**：實際 `user.go` 欄位為 `email / name / status / role / phone / employee_no / is_customer / account_name / token_version / password_hash`。**缺少**計畫要求的 `data_scope`、`customer_id`、`is_primary`、`temp_password_expires_at`、`must_change_password`、`failed_login_attempts`、`locked_at`、`locale`、`company_id`、`department_id`（role 欄位存在但無 enum 完整約束、無 data_scope）。

- [x] **Step 1: 建立 Company / Department schema** — `company.go` / `department.go` 已存在
- [x] **Step 2: 建立 User schema 基本欄位** — `user.go` 已存在
- [ ] **Step 3: 補齊 User 欄位** — `data_scope`（enum all/company/department/self）、`customer_id`、`is_primary`、`temp_password_expires_at`、`must_change_password`、`failed_login_attempts`、`locked_at`、`locale`、`company_id`、`department_id`；`role` 改為 enum 約束。需同步 migration（目前 00001_init_schema 以現有欄位為準）
- [ ] **Step 4: 建立 testutil** — `backend/internal/testutil/db.go` 不存在；DB 相依測試目前散見各 package，未集中
- [ ] **Step 5: 產生 Ent code（`task ent:gen`）與測試**

**待辦摘要**：補齊 user 欄位 ＋ 建 testutil。

---

## Task 2: Casbin model、enforcer、預設 seeder（細部 1.2.1–1.2.3）

**實際狀態：✅ 完成（執行層）**

**實際產物（已存在）：**
- `backend/internal/auth/casbin.go` — `BuiltinRoles`（7 角色）、`Enforce` / `EnforceAny`
- `backend/internal/auth/rbac_model.conf`、`backend/internal/auth/rbac_policy.csv`（go:embed 內嵌）
- `backend/internal/auth/rbac.go`

**說明**：已實作進程級 enforcer（model + policy 來自 internal/auth go:embed 內嵌）。**與計畫差異**：未用 PG DB adapter 與 `SeedDefaultPolicies`（計畫要求），改以 string-adapter（註解明示「production 遷移至 casbin_rules 表後改為 DB adapter」）。

- [x] **Step 1: 實作 Casbin model + 7 內建角色 policy** — `casbin.go`、`rbac_model.conf`、`rbac_policy.csv`
- [x] **Step 2: 實作 enforcer（進程級、內嵌 policy）** — `casbinEnforcer()` / `Enforce` / `EnforceAny`
- [ ] **Step 3: 遷移 DB adapter** — 以 `casbin_rules` 表取代 string-adapter 並提供 `SeedDefaultPolicies` 冪等 seeder（目前 seed 於 `cmd/seed` 未含此項）

**待辦摘要**：DB adapter + 冪等 seeder（非阻塞，可後置）。

---

## Task 3: RLS policies、注入 hook、高權繞過（細部 1.3.1–1.3.3）

**實際狀態：🟡 部分完成**

**實際產物（已存在）：**
- `backend/internal/auth/rls.go` — `RLSScope`、`WithRLS`/`RLSFrom`、`RLSStatements`、`ApplyRLS`、`ScopeForRole`
- migration `backend/database/migrations/00002_rls_policies.sql`

**說明**：已實作 RLS 語句產生與 `SET LOCAL` 套用，及 migration。**與計畫差異**：無 `authz.Identity`/`NewContext`/`WrapDriver`（Ent client 每交易自動 SET LOCAL 的注入層未建；`third_party/database/ent.go` 註解明示該 hook 待加）。

- [x] **Step 1: 實作 RLS 語句層** — `rls.go`（RLSStatements / ApplyRLS / ScopeForRole）
- [x] **Step 2: RLS migration** — `00002_rls_policies.sql`
- [ ] **Step 3: 實作 Identity + WrapDriver 注入** — 建立 `authz.Identity`、`rls.NewContext`/`FromContext`、`rls.WrapDriver(dialect.Driver)`，於 Ent client 建立處（`third_party/database/ent.go`）掛接，使每筆交易自動 `SET LOCAL`
- [ ] **Step 4: 高權繞過（developer）** — 見 Task 14

**待辦摘要**：Identity/Context/WrapDriver 注入層。

---

## Task 4: OAuth2 導向與 callback（細部 1.4.1–1.4.2）

**實際狀態：✅ 完成**

**實際產物（已存在）：**
- `backend/internal/auth/oidc.go` — `NewGoogleOAuthConfig`、`googleExchanger`、`NewGoogleVerifier`、`NewState`、`NewRegistrationToken`、`NewOneTimeStore`
- `backend/internal/handlers/auth_handler.go` — `GoogleLogin`、`GoogleCallback`
- `backend/internal/server/server.go` — OAuth 路由掛載
- `backend/config/auth.go` — OIDC 設定

**說明**：Google Workspace OIDC 導向 / callback、state 生命週期、id_token 驗證、一次性 token store 皆已實作。

- [x] **Step 1: config 擴充** — `config/auth.go`
- [x] **Step 2: 實作 OAuth 導向** — `GoogleLogin` + `NewState` + redirect
- [x] **Step 3: 實作 callback / 驗證** — `GoogleCallback` + `NewGoogleVerifier` + `VerifyIDToken`
- [x] **Step 4: 註冊/登入流程銜接** — `completeLogin` / `issueTokenPair`

**狀態**：無待辦。

---

## Task 5: 註冊完成與 guest 審核（細部 1.4.3–1.4.4）

**實際狀態：🟡 部分完成**

**實際產物（已存在）：**
- `handlers/auth_handler.go` — `RegisterComplete`
- git commit `9780324`（員工註冊完成 guest→pending）

**說明**：員工 OIDC 首登的 `RegisterComplete`（guest→pending 過渡）已實作。**未含** guest 審核管理介面（`super`/`company_admin` 審核）與審核流程的完整閉環。

- [x] **Step 1: RegisterComplete handler** — guest→pending 過渡
- [ ] **Step 2: guest 審核管理** — 審核清單/通過/拒絕 API 與介面（尚未實作）

---

## Task 6: 客戶帳密登入（細部 1.5.1）

**實際狀態：✅ 完成**

**實際產物（已存在）：**
- `backend/internal/auth/password.go` — `HashPassword`/`VerifyPassword`、`LoginLock`（RecordFailure/IsLocked/Clear）
- `handlers/auth_handler.go` — `Login`
- git commit `c396a47`（客戶密碼登入與 5 次失敗鎖定）

**說明**：客戶帳密登入與失敗鎖定已實作。

- [x] **Step 1: 密碼雜湊** — `HashPassword`/`VerifyPassword`
- [x] **Step 2: CustomerLogin** — `Login` handler + 主帳號業務拒絕
- [x] **Step 3: 失敗鎖定** — `LoginLock`（Valkey 計次）

---

## Task 7: 臨時密碼與首登強制修改（細部 1.5.2）

**實際狀態：✅ 完成（2026-09-25 查證）**

**實際產物：**
- 欄位：`ent/schema/user.go` 的 `must_change_password`、`temp_password_expires_at`（migration 已含）。
- 臨時密碼發放：`auth.GenerateTempPassword`（≥12 字元）＋`auth.HashPassword`；效期常數 `handlers.tempPasswordTTL`（now+24h）。建立客戶（`customer_service.buildCustomerAccount`）與新增子帳號（`customer_account_service`）皆產生臨時密碼並回傳**僅此一次**。
- 首登強改：middleware 檢查 `Identity.MustChangePassword` —— 除了 `ChangePassword` 之外**所有** RPC 回 `AUTH-3004`（`errcode.AuthPasswordChangeRequired`）；`AuthService.ChangePassword` 為登入態唯一可用者，改完清除 `must_change_password` 並 `tv+1`。
- proto：`ChangePassword`、`ResetCustomerPassword`（`auth.proto`）；`LoginResponse.must_change_password` 供客戶端導向。

- [x] **Step 1: 補 user 欄位**
- [x] **Step 2: 臨時密碼發放**（24h 效期）
- [x] **Step 3: 首登強制修改**（middleware 閘門 AUTH-3004）
- [x] **Step 4: proto + 測試**（`auth_password.go` 相關測試）

**待辦摘要**：無。**注意**：App 端**未**實作改密碼頁 —— 臨時密碼登入後會收到 `AUTH-3004`，目前 App 把它當一般錯誤顯示。這是既有缺口（非本次變更），已記於 `04-master-data` 計畫的 Task 6.7 殘項。

---

## Task 8: 登入鎖定與密碼重置（細部 1.5.3–1.5.4）

**實際狀態：🟡 部分完成**

**實際產物（已存在）：**
- `internal/auth/password.go` — `LoginLock`（鎖定完成）

**說明**：登入 5 次失敗鎖定 30 分鐘已實作（`LoginLock`）。**未含**管理員密碼重置流程（清除/重設客戶或業務密碼）。

- [x] **Step 1: 登入鎖定** — `LoginLock`（RecordFailure/IsLocked）
- [ ] **Step 2: 密碼重置** — 管理員清除/重置密碼 API（`UserService`）與流程

---

## Task 9: scs + Valkey session 與 access JWT（細部 1.6.1–1.6.2）

**實際狀態：✅ 完成**

**實際產物（已存在）：**
- `backend/internal/auth/session.go` — `NewSessionStore`、`WebSessionManager`、`EstablishWebSession`
- `backend/internal/auth/token.go` — `NewTokenManager`、`IssueAccess`/`VerifyAccess`
- `backend/internal/auth/stores.go` — `NewRedisStore`（KVStore 實作）
- `backend/third_party/cache/valkey.go`

**說明**：access JWT（1h）＋ scs session manager（Web）＋ Valkey KV 皆已實作。

- [x] **Step 1: JWT access token** — `IssueAccess`/`VerifyAccess`
- [x] **Step 2: scs session manager** — `WebSessionManager` / `EstablishWebSession`
- [x] **Step 3: Valkey KVStore** — `newStore`/`RedisStore`

---

## Task 10: refresh token 旋轉與 token_version 比對（細部 1.6.3–1.6.4）

**實際狀態：✅ 完成**

**實際產物（已存在）：**
- `internal/auth/token.go` — `IssueRefresh`/`VerifyRefresh`/`ConsumeRefresh`、`IssueAccess`；`CurrentTokenVersion`/`BumpTokenVersion`
- migration `00004_users_token_version.sql`
- git commit `fc17e63`、`f1bd1a9`（refresh 原子旋轉、session tv 比對）、`37a84c4`

**說明**：refresh token 旋轉、token_version 撤銷（改密碼/停用/角色變更/強制登出 → tv+1 全數失效）、session tv 比對已實作。

- [x] **Step 1: refresh 發放** — `IssueRefresh`
- [x] **Step 2: refresh 旋轉** — `ConsumeRefresh`（原子）
- [x] **Step 3: token_version 比對** — `CurrentTokenVersion`/`BumpTokenVersion`；migration 00004

---

## Task 11: Authenticate 與 X-Api-Token middleware（細部 1.6.5–1.6.6）

**實際狀態：✅ 完成（2026-09-25）**

**實際產物（已存在）：**
- `Server.authzMiddleware`（`internal/server/server.go`）＝規格所稱的 `Authenticate`：解析 scs session／`Authorization: Bearer` JWT（`BearerToken`＋`TokenManager.VerifyAccess`）→ `identityFor` 載入使用者與公司狀態 → 注入 `authz.Identity` 與 `auth.RLSScope`。**未另立 middleware 目錄**：本專案所有業務路由共用 `authzMiddleware` 這一層（`domains.go` 以 `sessions.LoadAndSave(s.authzMiddleware(...))` 掛在 `/api/v1`），再以 `protectedRPC` 表逐 path 做 OpenFGA Check（D32 取代 Casbin）。
- **X-Api-Token（1.6.6）已完成（2026-09-25）**：`internal/auth/apitoken.go`（`ParseAPITokens`／`MatchAPIToken`／`AllowsRPC`）＋ `authzMiddleware` 的第三條憑證路徑。設計取捨見檔頭：只存 SHA-256（原文不落設定檔）、綁**真實使用者**（稽核 `user_id` 有 FK 到 `users`，且權限判定自動沿用該人角色）、每 token 的 `rpc_prefixes` 白名單（空 = 不允許任何 RPC，fail-closed）。使用者憑證優先（session → JWT → API token，避免降級混淆）。設定 `API_TOKENS`（JSON，`config/auth.go`＋`.env.example` 已載明），格式錯誤即啟動 Fatal。
- **PublicPaths 白名單**：未以「路徑清單跳過 middleware」實作，而是**白名單只對 `protectedRPC` 內的路徑生效** —— 公開端點（QR 兌換／登入／`/me`）本就不在該表，自然不被授權與 token 白名單限制（語意等價且少一份要同步的清單）。
- 驗收已由測試覆蓋：`TestAuthzMiddlewareBearerJWT`（session/JWT）、`TestAuthzMiddlewareAPIToken`（白名單內外、無效 token、綁定已停用、公開端點不受限、使用者憑證優先）、`internal/auth/apitoken_test.go`（解析嚴格性、雜湊比對大小寫不拘、空名單 fail-closed）。

- [x] **Step 1: config 擴充** — `Auth.APITokens`（JSON：name／sha256／user_id／rpc_prefixes）
- [x] **Step 2: 實作 `Authenticate`** — session 與 Bearer JWT 兩路徑（`authzMiddleware`）
- [x] **Step 3: 實作 `ApiTokenAuthenticate`** — `X-Api-Token` 雜湊比對＋RPC 前綴白名單
- [x] **Step 4: 公開路徑** — 以「白名單只管 `protectedRPC` 內路徑」達成（無需獨立清單）
- [x] **Step 5: 全域掛載** — `/api/v1` 統一掛載

**待辦摘要**：無。**注意**：`X-Api-Token` 目前**尚無消費者**（`cmd/platform-cron` 直連資料庫、不經 HTTP），此路徑為規格要求的 M2M 能力預備；啟用時只需設 `API_TOKENS`。

**待辦摘要**：完整未開始，為後續 domain（02~09）共用地基。

---

## Task 12: 強制登出（細部 1.7.1–1.7.2）

**實際狀態：🟡 部分完成**

**實際產物（已存在）：**
- `internal/auth/token.go` — `BumpTokenVersion`（撤銷）
- `handlers/auth_handler.go` — `Logout`

**說明**：`Logout`（自願登出，撤銷 token）與 `BumpTokenVersion` 已實作。**未含**管理員**強制**登出指定使用者的完整 `ForceLogout` usecase/API（`UserService.ForceLogout(user_id)`）。

- [x] **Step 1: 自願登出** — `Logout` handler + token 撤銷
- [ ] **Step 2: 管理員強制登出** — `UserService.ForceLogout(user_id)` + `forceLogoutUser`（交易內 tv+1、Valkey session 刪除為提交後動作）

---

## Task 13: ability API（細部 1.8.1）

**實際狀態：✅ 完成**

**實際產物（已存在）：**
- `backend/proto/salesorder/v1/ability.proto` — `AbilityService.GetAbility`
- `backend/internal/proto/salesorder/v1/ability.pb.go`
- `backend/internal/domain/auth/ability.go` — `BuildAbility`
- git commit `1a6c122`、`ef1bf04`

**說明**：`GetAbility` 產出 CASL JSON 規則（Action/Subject/Conditions/Inverted）已實作；front 端 `@casl/ability` 消費。

- [x] **Step 1: proto AbilityService** — `GetAbility`
- [x] **Step 2: BuildAbility 規則表** — domain 內建預設規則
- [ ] **Step 3: 規則來源改由 `role_permissions` 驅動** — 計畫要求 Phase 2 改用表驅動；目前 `role_service.go` 已提供 RoleService.UpdateRolePermissions 寫入 `role_permissions`，但 ability 規則是否已由表驅動需確認（見 02 計畫）

---

## Task 14: developer 逃生門與 audit 介面（細部 1.11.1–1.11.3）

**實際狀態：✅ 完成（2026-09-25 查證）**

**實際產物：**
- config：`config.API` 的 `Env`、`DeveloperAccountEnabled`（`DEVELOPER_ACCOUNT_ENABLED`）。
- fail-fast：`server.Init` 於 `ENV=production && DeveloperAccountEnabled` 時拒絕啟動。
- developer 繞過：`authzMiddleware` 的 `authorizeRPC` 對 `developer`/`super` 放行（`identityFor` 亦僅在開關開啟時接受 developer 帳號），且**稽核照寫**（不因繞權而省略）。
- audit：`internal/audit` 套件（`Recorder` 介面語意由 `audit.Record` 承載，含 snapshot 淨化、trace 標記、`_actor_kind`）；`NoopRecorder` 未另立 —— 需要時以不注入 meta 表達（本專案無「無稽核」情境）。

> D32 衝突已解：授權為 **OpenFGA + RLS**（Casbin 已移除），見 `2026-09-18-openfga-authz-backend-plan.md`。

- [x] **Step 1: config** + fail-fast 防護
- [x] **Step 2: developer bypass**（OpenFGA Check 與 RLS 皆放行）
- [x] **Step 3: audit 介面**（`internal/audit` 已存在並廣泛使用）
- [x] **Step 4: 啟動防護**（`Init()` 實作）

---

*最後更新：2026-09-25（Task 7 臨時密碼/首登強改、Task 14 developer 逃生門與 audit 對齊實際實作；此前 2026-09-24 已對齊 Task 11 X-Api-Token）*
