# 多公司訂出貨系統 1.0 — 實作計畫

> **版本：v2.9.0**（2026-08-03：對應規格書 v1.0.34；新增退貨 / 店家帳號管理（一主多子）/ 促銷推播任務；不存金額、無 Email、偏好送貨日等需求更新）
> 依據規格書 `docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）。
> 本文件為**唯一實作計畫**，整合原里程碑總覽與詳細任務清單；舊版計畫（`2026-07-16-sales-order-2.0-phase1.md`、`2026-07-17-multi-company-sales-order-1-0.md`、`2026-07-17-multi-company-sales-order-1-0-companies.md`、`2026-07-17-sales-order-1-0-milestones.md`）已於 2026-07-17 收斂刪除。
> 每個 Task 包含 Goal、Files、Interfaces、Steps、Acceptance Criteria，使用 checkbox 追蹤。

---

## 里程碑總覽

| Phase | 名稱 | 目標 | 主要交付物 | 相依於 |
|:---:|---|---|---|---|
| 0 | 基礎建設 | 建立 monorepo、開發環境、專案骨架 | Taskfile、docker-compose、Go/SolidJS/Flutter 骨架、CI/CD | — |
| 1 | 認證與授權 | 實作多租戶認證與 RBAC | OAuth2（Google Workspace）、JWT/Session、Casbin、RLS、CASL.js、開發者帳號 | Phase 0 |
| 2 | 多租戶主檔 | 公司、部門、使用者、角色權限、字典、稽核 | companies/departments/users、roles/role_permissions、metadicts、audit_logs | Phase 1 |
| 3 | 業務主檔 | 客戶、商品、倉庫、車次、分切規格 | customers、customer_counters、products、warehouses、routes、cutting_specs、customer_products | Phase 2 |
| 4 | 訂單與通知 | 下單流程、狀態機、退貨、通知系統 | sales_orders、return_requests、notifications（FCM/站內） | Phase 3 |
| 5 | 派車與列印 | 派車看板、四種單據、列印記錄 | Kanban、Connect 串流、print_logs、Gotenberg | Phase 4 |
| 6 | App 功能 | 業務/客戶 App 下單、新增客戶與歷史查詢 | Flutter 登入、底部導覽、快速下單、新增客戶、推播 | Phase 4 |
| 7 | 公告與 UI | 公告管理、中台介面、公司識別呈現 | CMS、側邊欄、主題 | Phase 2 |
| 8 | 部署與維運 | k8s 部署、備份監控、安全合規、上線 | Helm/k8s manifests、GCS 備份、災難復原、Big Bang | Phase 5–7 |

---

## 跨 Phase 注意事項

1. **型別同步**：後端 `.proto` 為唯一來源，產生前端 TypeScript 與 Flutter Dart 型別；proto package `v1`，每 domain 一個 Service（如 `company.v1.CompanyService`）。
2. **API 協定**：業務 API 統一 Connect-RPC；REST 僅保留公開端點（版本查詢、OAuth 導向/回調、QR 兌換、檔案上傳/下載、`companies/public/{identifier}`），前綴 `/api/v1`。
3. **軟刪除**：所有業務實體統一使用 `deleted_at`，查詢預設排除。
4. **多租戶**：每筆業務資料皆帶 `company_id` / `department_id`，RLS 為最後防線。
5. **測試**：每個 Phase 結束前需有對應單元測試與至少一條整合測試路徑；三端覆蓋率門檻 70%（CI 強制）。
6. **文件**：每個 domain 新增時同步更新 API 文件與前端 ability 規則。

---

## Phase 0: 基礎建設與專案初始化

### Task 0.1: 建立 monorepo 根目錄結構

**Goal:** 建立 pnpm workspace + Turborepo 的 monorepo 骨架。

**Files:**
- Create: `package.json`
- Create: `pnpm-workspace.yaml`
- Create: `turbo.json`
- Create: `.gitignore`
- Create: `README.md`

**Interfaces:**
- `pnpm install` 可於根目錄安裝所有 workspace 依賴。
- `pnpm turbo run build` 可依序建置 backend/frontend/app。

**Steps:**
- [ ] Step 1: 建立 `package.json`，設定 `private: true` 與 `devDependencies`（turbo、prettier、eslint）。
- [ ] Step 2: 建立 `pnpm-workspace.yaml`，納入 `backend`、`frontend`、`app`、`infra`。
- [ ] Step 3: 建立 `turbo.json`，定義 `build`、`dev`、`test`、`lint` pipeline。
- [ ] Step 4: 建立根目錄 `.gitignore`，排除 `.turbo`、`node_modules`、環境檔案。
- [ ] Step 5: 建立 `README.md`，說明啟動方式與目錄結構。

**Acceptance Criteria:**
- [ ] `pnpm install` 成功。
- [ ] `pnpm turbo run build` 可執行（即使 backend/frontend/app 尚無內容，至少不報錯）。

---

### Task 0.2: 建立開發環境 docker-compose

**Goal:** 提供一致的本地開發基礎設施。

**Files:**
- Create: `docker-compose.dev.yml`

**Interfaces:**
- `task infra:start` 啟動 PostgreSQL、Valkey、Gotenberg（1.0 無 Email，不含 Mailpit）。
- `task infra:stop` 停止所有容器。

**Steps:**
- [ ] Step 1: 定義 `postgres` 服務（port 5432、volume、env）。
- [ ] Step 2: 定義 `valkey` 服務（port 6379）。
- [ ] Step 3: 定義 `gotenberg` 服務（port 3001）。
- [ ] Step 4: 定義共用 volume。

**Acceptance Criteria:**
- [ ] `task infra:start` 後 `docker ps` 看到 3 個容器運行（PostgreSQL、Valkey、Gotenberg，無 Email 不含 Mailpit）。
- [ ] `task infra:stop` 後容器停止。

---

### Task 0.3: 建立根目錄 Taskfile

**Goal:** 提供統一的開發指令入口。

**Files:**
- Create: `Taskfile.yml`

**Interfaces:**
- `task dev` 啟動後端與前端開發伺服器。
- `task infra:start/stop` 控制基礎設施。

**Steps:**
- [ ] Step 1: 定義 `dev`、`infra:start`、`infra:stop`、`test`、`lint` 任務。
- [ ] Step 2: 使用 `task -p` 平行啟動 backend dev 與 frontend dev。

**Acceptance Criteria:**
- [ ] `task --list` 顯示所有任務。
- [ ] `task infra:start` 成功。

---

### Task 0.4: 建立後端專案骨架

**Goal:** 建立 Go 後端基本結構。

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/config/config.go`
- Create: `backend/internal/server/server.go`
- Create: `backend/Taskfile.yml`

**Interfaces:**
- `config.Load()` 載入環境變數。
- `server.New(cfg)` 回傳 Chi router。

**Steps:**
- [ ] Step 1: `go mod init github.com/salesorder/sales-order-1.0/backend`。
- [ ] Step 2: 安裝 chi、cors、viper、pgx、ent、casbin、jwt、connect-go、oauth2/oidc 等依賴。
- [ ] Step 3: 實作 `config.Config` 與 `Load()`。
- [ ] Step 4: 實作基礎 Chi server（middleware、CORS、健康檢查路由）。
- [ ] Step 5: 建立 `backend/Taskfile.yml`（dev、run、test、migrate、proto:gen）。

**Acceptance Criteria:**
- [ ] `cd backend && go run ./cmd/server` 可啟動並監聽 3080。
- [ ] `GET /api/v1/version` 回傳版本資訊。

---

### Task 0.5: 建立前端專案骨架

**Goal:** 建立 SolidJS + TypeScript + Vite 前端結構。

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/tsconfig.json`
- Create: `frontend/src/main.tsx`
- Create: `frontend/src/App.tsx`
- Create: `frontend/src/index.css`

**Interfaces:**
- `pnpm dev` 啟動開發伺服器於 port 3000。
- `pnpm build` 輸出至 `dist/`。

**Steps:**
- [ ] Step 1: 初始化 package.json，安裝 solid-js、@solidjs/router、@tanstack/solid-query、@tanstack/solid-table、casl/ability、valibot、tailwindcss、shadcn-solid 相關依賴。
- [ ] Step 2: 設定 Vite 6 + TypeScript 5.9。
- [ ] Step 3: 設定 Tailwind CSS 與 `~/*` path alias。
- [ ] Step 4: 建立 `main.tsx`、`App.tsx`、基礎路由。

**Acceptance Criteria:**
- [ ] `pnpm dev` 成功啟動。
- [ ] `pnpm build` 成功輸出 dist。

---

### Task 0.6: 建立 App 專案骨架

**Goal:** 建立 Flutter 專案結構與 flavor 設定。

**Files:**
- Create: `app/pubspec.yaml`
- Create: `app/lib/main_dev.dart`
- Create: `app/lib/main_prod.dart`
- Create: `app/analysis_options.yaml`
- Create: `app/Taskfile.yml`

**Interfaces:**
- `fvm flutter run --flavor dev --target lib/main_dev.dart` 可啟動 App。
- `fvm flutter build apk --flavor prod --target lib/main_prod.dart` 可建置。

**Steps:**
- [ ] Step 1: 設定 `pubspec.yaml`（Flutter 3.35.2、Dart ≥3.9.0）。
- [ ] Step 2: 安裝依賴：connectrpc、auto_route、get_it、dio、sembast、flutter_secure_storage、firebase_messaging、talker。
- [ ] Step 3: 建立 `main_dev.dart` / `main_prod.dart`。
- [ ] Step 4: 設定 Android/iOS flavor。
- [ ] Step 5: 建立 `app/Taskfile.yml`。

**Acceptance Criteria:**
- [ ] `fvm flutter pub get` 成功。
- [ ] `fvm flutter analyze` 無嚴重錯誤。

---

### Task 0.7: 設定 proto 與型別同步流程

**Goal:** 建立後端為唯一來源的 proto 定義與產生流程。

**Files:**
- Create: `backend/proto/v1/common.proto`
- Create: `backend/proto/v1/auth.proto`
- Create: `backend/buf.gen.yaml`
- Create: `frontend/src/lib/proto/.gitkeep`
- Create: `app/lib/proto/.gitkeep`

**Interfaces:**
- `task backend:proto:gen` 產生 Go / TypeScript / Dart 程式碼。

**Steps:**
- [ ] Step 1: 安裝 buf、protoc-gen-go、protoc-gen-connect-go、protoc-gen-connect-es、protoc-gen-connect-dart。
- [ ] Step 2: 建立 `buf.gen.yaml` 與基礎 proto 檔案。
- [ ] Step 3: 設定產生腳本，將 TypeScript/Dart 輸出至 frontend/app。

**Acceptance Criteria:**
- [ ] `task backend:proto:gen` 成功產生三端程式碼。

---

### Task 0.8: 建立 CI/CD 基礎管線

**Goal:** 建立 GitHub Actions 基礎 CI。

**Files:**
- Create: `.github/workflows/ci.yml`

**Interfaces:**
- PR 觸發 lint、test、build。

**Steps:**
- [ ] Step 1: 設定觸發條件（push/PR 至 main）。
- [ ] Step 2: 設定 Go lint、test、build。
- [ ] Step 3: 設定 frontend type check、lint、build。
- [ ] Step 4: 設定 flutter analyze。

**Acceptance Criteria:**
- [ ] CI pipeline 在 PR 時成功執行。

---

### Task 0.9: 建立資料庫 migration 基礎

**Goal:** 設定 Goose migration 與開發資料庫初始化。

**Files:**
- Create: `backend/database/migrations/00001_init_schema.sql`
- Create: `backend/database/migrations/00002_rls_policies.sql`

**Interfaces:**
- `task backend:migrate:up` 套用 migration。

**Steps:**
- [ ] Step 1: 安裝 goose。
- [ ] Step 2: 建立 `00001_init_schema.sql`，包含 `goose` 版本控制表。
- [ ] Step 3: 建立 `00002_rls_policies.sql`，預備 RLS 啟用。

**Acceptance Criteria:**
- [ ] `task backend:migrate:up` 成功。
- [ ] PostgreSQL 中出現 goose 版本控制表。

---

### Task 0.10: Phase 0 驗收

**Goal:** 確認基礎建設可運作。

**Acceptance Criteria:**
- [ ] `task infra:start` 成功。
- [ ] `cd backend && task dev` 可啟動後端。
- [ ] `cd frontend && pnpm dev` 可啟動前端。
- [ ] `cd app && fvm flutter run --flavor dev` 可啟動 App。
- [ ] CI pipeline 通過。

---

## Phase 1: 認證與授權

### Task 1.1: 設計 users 與角色 Ent schema

**Goal:** 建立使用者資料模型，支援員工與客戶帳號。

**Files:**
- Create: `backend/ent/schema/user.go`
- Create: `backend/ent/schema/company.go`
- Create: `backend/ent/schema/department.go`

**Interfaces:**
- `ent.User`、`ent.Company`、`ent.Department` 產生後可用。

**Steps:**
- [ ] Step 1: 定義 `Company` schema（name、tax_id、status、identifier、public_info、capabilities、logo_url 等）。
- [ ] Step 2: 定義 `Department` schema（name、company_id）。
- [ ] Step 3: 定義 `User` schema（email、name、status、role、company_id、department_id、phone、employee_no、is_customer、customer_id、account_name、password_hash）。
- [ ] Step 4: 執行 `task backend:ent:gen`。

**Acceptance Criteria:**
- [ ] `ent` 產生碼可編譯。
- [ ] 可透過 Ent 建立 company/department/user。

---

### Task 1.2: 實作 Casbin model 與 policy 儲存

**Goal:** 建立 RBAC with domain 授權模型。

**Files:**
- Create: `backend/config/casbin_model.conf`
- Create: `backend/internal/authz/casbin.go`

**Interfaces:**
- `authz.Enforcer` 提供 `Enforce(sub, dom, obj, act)`。

**Steps:**
- [ ] Step 1: 定義 Casbin model（`p, sub, dom, obj, act` / `g, _, _, _`）。
- [ ] Step 2: 使用 PostgreSQL adapter 初始化 enforcer。
- [ ] Step 3: 建立預設 policy seeder（super、company_admin、dept_admin、staff、customer 權限）。
- [ ] Step 4: 寫單元測試驗證權限判斷。

**Acceptance Criteria:**
- [ ] `staff` 無法存取其他部門資源。
- [ ] `super` 可存取所有公司資源。
- [ ] `company_admin` 可跨部門但僅限自己公司。

---

### Task 1.3: 實作 RLS policies

**Goal:** 在 PostgreSQL 層實現租戶隔離。

**Files:**
- Create: `backend/database/migrations/00003_rls_policies.sql`
- Create: `backend/internal/database/rls.go`

**Interfaces:**
- `rls.SetContext(ctx, companyID, departmentID)` 注入 session variables。

**Steps:**
- [ ] Step 1: 為 user/company/department 等表啟用 RLS。
- [ ] Step 2: 建立 policy：依 `app.current_company_id` / `app.current_department_id` 篩選。
- [ ] Step 3: 建立 super / company_admin 繞過 RLS 的機制。
- [ ] Step 4: 在 Ent client 建立 connection hook 自動注入。

**Acceptance Criteria:**
- [ ] 整合測試驗證不同 role 只能看到自己權限範圍內的資料。

---

### Task 1.4: 實作 OAuth2/OIDC 員工登入

**Goal:** 支援 Google Workspace 登入。

**Files:**
- Create: `backend/internal/domain/auth/handler.go`
- Create: `backend/internal/domain/auth/oauth.go`
- Create: `backend/internal/domain/auth/usecase.go`

**Interfaces:**
- `GET /api/v1/auth/oauth/:provider` 導向 provider。
- `GET /api/v1/auth/oauth/:provider/callback` 處理回調。

**Steps:**
- [ ] Step 1: 設定 OAuth2 config（Google Workspace）。
- [ ] Step 2: 實作 OIDC ID token 驗證。
- [ ] Step 3: 首次登入跳轉註冊完成頁（選擇公司 + 輸入姓名），完成後建立 `guest` user（pending，歸屬所選公司）；未完成前不建立帳號。
- [ ] Step 4: 已審核使用者發放 JWT / session。

**Acceptance Criteria:**
- [ ] 可用 Google 帳號登入並取得 session。
- [ ] 首次登入完成註冊頁後建立 `guest`（歸屬所選公司），無法操作資料。
- [ ] company_admin 可看到並審核自己公司的 pending guest。

---

### Task 1.5: 實作客戶密碼登入

**Goal:** 支援客戶一主多子帳號（`account_name` + 密碼；主帳號僅管理、子帳號業務登入）。

**Files:**
- Update: `backend/internal/domain/auth/handler.go`
- Create: `backend/internal/domain/auth/customer.go`

**Interfaces:**
- Connect-RPC：`AuthService.CustomerLogin`、`AuthService.ResetCustomerPassword`

**Steps:**
- [ ] Step 1: 驗證 `account_name`（客戶內唯一）+ password（Argon2id）；每個客戶 1 個主帳號（`is_primary`）+ 多個子帳號，各帳號獨立停用/重置；主帳號僅供管理登入（業務 API 403）、子帳號為業務登入身分。
- [ ] Step 2: 首次登入強制修改密碼；密碼強度最少 8 字元。
- [ ] Step 3: 臨時密碼效期 1 天（`temp_password_expires_at`），過期需 dept_admin 以上重新產生。
- [ ] Step 4: 連續 5 次密碼錯誤鎖定（`failed_login_attempts` / `locked_at`），30 分鐘自動解除；重置密碼連帶解鎖。
- [ ] Step 5: 產生 customer JWT / session。
- [ ] Step 6: dept_admin 以上可重置客戶密碼。

**Acceptance Criteria:**
- [ ] 客戶可用帳號名稱 + 密碼登入；每個客戶 1 主帳號 + 多子帳號（老闆/主廚），任一帳號異動不影響其他帳號；主帳號登入僅能帳號管理，業務 API 被拒。
- [ ] 首次登入未改密碼前無法進入系統。
- [ ] 臨時密碼過期無法登入；連續 5 次錯誤鎖定、30 分鐘後解除。

---

### Task 1.6: 實作 Session / JWT 管理

**Goal:** 統一管理 Web cookie 與 App/API token。

**Files:**
- Create: `backend/internal/session/manager.go`
- Create: `backend/internal/middleware/auth.go`

**Interfaces:**
- `middleware.Authenticate(session)` 解析並注入當前使用者。
- `middleware.ApiTokenAuthenticate()` 驗證 `X-Api-Token`（僅 server-to-server）。

**Steps:**
- [ ] Step 1: 使用 scs + Valkey 實作 session store。
- [ ] Step 2: 實作 JWT 產生與驗證（access token 1 小時，claim 含 `tv` = `users.token_version`）。
- [ ] Step 3: 實作 refresh token（30 天、旋轉制：每次 refresh 核發新 token 並作廢舊 token；後端僅存雜湊）。
- [ ] Step 4: 實作 `token_version` 驗證 middleware：每次請求比對 JWT `tv` 與資料庫目前值，不符即拒絕。
- [ ] Step 5: 實作 Authenticate middleware。
- [ ] Step 6: 實作 ApiToken middleware（`X-Api-Token`，僅 server-to-server）。

**Acceptance Criteria:**
- [ ] Web 登入後可取得 httpOnly cookie。
- [ ] App 請求帶 `authorization` JWT 可通過驗證；`X-Api-Token` 僅限 server-to-server。
- [ ] access token 過期後可以 refresh token 換發，舊 refresh token 旋轉作廢。
- [ ] `token_version` 變更後，舊 JWT / refresh token 立即失效。

---

### Task 1.7: 實作強制登出

**Goal:** 管理員可強制指定使用者登出。

**Files:**
- Update: `backend/internal/session/manager.go`
- Update: `backend/internal/domain/users/usecase.go`

**Interfaces:**
- Connect-RPC：`UserService.ForceLogout`。

**Steps:**
- [ ] Step 1: 記錄使用者 session blacklist 或刪除 Valkey session。
- [ ] Step 2: `users.token_version + 1`，使該使用者所有 JWT / refresh token 立即失效。
- [ ] Step 3: 提供 API 給 super / company_admin / dept_admin 使用。
- [ ] Step 4: 寫入 audit log。

**Acceptance Criteria:**
- [ ] 強制登出後該使用者下次請求被拒。

---

### Task 1.8: 實作 ability API

**Goal:** 前端可取得當前使用者的 ability JSON。

**Files:**
- Create: `backend/internal/domain/auth/ability.go`

**Interfaces:**
- Connect-RPC：`AbilityService.GetAbility`

**Steps:**
- [ ] Step 1: 依 role + company + department 產生 ability 規則（Phase 1 先以內建預設規則實作；Phase 2 Task 2.9 起改由 `role_permissions` 表驅動）。
- [ ] Step 2: 回傳 CASL.js 可消费的 JSON。

**Acceptance Criteria:**
- [ ] 前端可取得 ability 並正確控制按鈕顯示。

---

### Task 1.9: 前端登入與路由守衛

**Goal:** Web 中台登入流程與權限控制。

**Files:**
- Create: `frontend/src/routes/signin.tsx`
- Create: `frontend/src/routes/onboarding.tsx`（OAuth 首登註冊完成頁：選公司 + 姓名）
- Create: `frontend/src/lib/auth/index.ts`
- Create: `frontend/src/constant/ability.ts`

**Interfaces:**
- 未登入導向 `/signin`。
- 依 ability 控制路由與元件顯示。

**Steps:**
- [ ] Step 1: 實作登入頁面與 OAuth2 跳轉。
- [ ] Step 2: 實作註冊完成頁（新帳號選擇公司 + 輸入姓名，完成後建立 pending guest）。
- [ ] Step 3: 實作 ability store（登入後載入一次，短 TTL 60 秒快取）。
- [ ] Step 4: 設定路由守衛。

**Acceptance Criteria:**
- [ ] 未登入訪問 `/admin` 自動導向 `/signin`。
- [ ] 依角色顯示/隱藏對應選單。

---

### Task 1.10: App OAuth2 / QR Code 登入

**Goal:** App 身分選擇與登入。

**Files:**
- Create: `app/lib/layer_presentation/stories/login/role_select_screen.dart`
- Create: `app/lib/layer_business/services/auth_service.dart`
- Create: `app/lib/layer_business/services/deep_link_service.dart`

**Interfaces:**
- 選擇「我是店家」或「我是業務」。
- 業務使用 OAuth2；客戶使用帳號 + 密碼或 QR Code（QR 帶入客戶身分後選擇已綁定帳號）。

**Steps:**
- [ ] Step 1: 實作身分選擇頁。
- [ ] Step 2: 使用 flutter_web_auth / app_auth 實作 OAuth2。
- [ ] Step 3: 實作 QR Code 深層連結解析（驗證含 company_id 的 token）。
- [ ] Step 4: 儲存 JWT 至 flutter_secure_storage。

**Acceptance Criteria:**
- [ ] 業務可用 OAuth2 登入。
- [ ] 客戶可用 QR Code 直接登入。

---

### Task 1.11: 開發者帳號與繞過機制

**Goal:** 提供不受限制的開發者帳號，僅開發/測試環境可用，上架時可關閉。

**Files:**
- Create: `backend/internal/middleware/developer.go`
- Modify: `backend/config/config.go`

**Steps:**
- [ ] Step 1: config 新增 `DEVELOPER_ACCOUNT_ENABLED`（development 預設 true、production 預設 false）。
- [ ] Step 2: 啟動防護：`ENV=production` 且開關為 true 時拒絕啟動（fail fast）。
- [ ] Step 3: seed `developer` 角色（`is_system=true`、`data_scope=all`）；開發環境 seed 預設開發者帳號（僅 `ENV=development`）。
- [ ] Step 4: middleware：developer 角色且開關啟用時，跳過 Casbin 檢查，RLS 注入 `data_scope=all`；開關關閉時 developer 帳號無法登入。
- [ ] Step 5: developer 操作照常寫入 `audit_logs`。

**Acceptance Criteria:**
- [ ] 開發環境 developer 帳號可跨公司/部門存取所有 API。
- [ ] 開關關閉後 developer 帳號無法登入。
- [ ] `ENV=production` + 開關 true 時後端拒絕啟動。

---

### Task 1.12: Phase 1 驗收

**Goal:** 確認認證與授權機制完整可用。

**Acceptance Criteria:**
- [ ] 員工 OAuth2 登入流程完整。
- [ ] 客戶密碼與 QR Code 登入完整。
- [ ] Casbin + RLS 權限隔離正確。
- [ ] 開發者帳號於開發環境可不受限制存取；開關關閉後無法登入。
- [ ] 強制登出有效。
- [ ] Web/App 皆可取得 ability。

---

## Phase 2: 多租戶主檔

### Task 2.1: 公司管理 API

**Goal:** CRUD 公司主檔。

**Files:**
- Create: `backend/internal/domain/companies/*`

**Interfaces:**
- Connect-RPC：`CompanyService`（List / Get / Create / Update / Delete）

**Steps:**
- [ ] Step 1: 設計 request/response model。
- [ ] Step 2: 實作 handler、usecase、repository。
- [ ] Step 3: 實作 identifier 唯一性檢查。
- [ ] Step 4: 公司停用時，該公司所有帳號無法登入（認證 middleware 檢查 company status）。
- [ ] Step 5: 實作 `customer_code_prefix` 欄位維護（大寫英數 1–4 碼、全系統唯一校驗；修改僅影響後續新編碼，計數器不重置）。

**Acceptance Criteria:**
- [ ] super 可建立/停用公司。
- [ ] company_admin 僅能查看/編輯自己公司。
- [ ] 公司可設定客戶編號前綴，且全系統唯一。

---

### Task 2.2: 部門管理 API

**Goal:** CRUD 部門主檔。

**Files:**
- Create: `backend/internal/domain/departments/*`

**Interfaces:**
- Connect-RPC：`DepartmentService`（List / Get / Create / Update / Delete）

**Acceptance Criteria:**
- [ ] super 可建立部門。
- [ ] company_admin 可管理所屬公司部門。

---

### Task 2.3: 使用者管理 API

**Goal:** 使用者 CRUD、角色指派、停用、強制登出。

**Files:**
- Create: `backend/internal/domain/users/*`

**Interfaces:**
- Connect-RPC：`UserService`（List / Get / Create / Update / Deactivate / ForceLogout）

**Acceptance Criteria:**
- [ ] super 可管理所有使用者。
- [ ] company_admin 可管理所屬公司使用者。
- [ ] dept_admin 可管理所屬部門 staff。

---

### Task 2.4: 公司識別與公開資訊 API

**Goal:** 管理公司 Logo、主色、公開資訊。

**Files:**
- Create: `backend/internal/domain/companies/branding.go`
- Create: `backend/internal/domain/companies/public.go`

**Interfaces:**
- `GET /api/v1/companies/public/:identifier`（無需認證）
- Connect-RPC：`CompanyService.UpdateBranding`、`CompanyService.UpdatePublicInfo`

**Steps:**
- [ ] Step 1: 實作 Logo 上傳 API，關聯 `file_assets`。
- [ ] Step 2: 實作公開資訊與 capabilities 更新。
- [ ] Step 3: 實作公開發現端點。

**Acceptance Criteria:**
- [ ] 公開端點無需認證即可取得公司公開資訊。
- [ ] super / company_admin 可編輯公司識別。

---

### Task 2.5: metadicts 字典檔 API

**Goal:** 系統級預設 + 部門擴充的字典管理。

**Files:**
- Create: `backend/internal/domain/metadicts/*`

**Interfaces:**
- Connect-RPC：`MetadictService`（List / Get / Create / Update / Delete / ListOptions）

**Steps:**
- [ ] Step 1: 設計 `metadicts` 資料表（type、code、display_name、department_id）。
- [ ] Step 2: 實作查詢邏輯：系統預設 + 當前部門擴充。
- [ ] Step 3: 實作選項服務供表單下拉使用。

**Acceptance Criteria:**
- [ ] 各部門可看到系統預設與自己的擴充值。
- [ ] 部門間擴充值互相隔離。

---

### Task 2.6: audit_logs 稽核日誌 API

**Goal:** 記錄關鍵操作並提供查詢。

**Files:**
- Create: `backend/internal/domain/auditlogs/*`

**Interfaces:**
- Connect-RPC：`AuditService.Record`（內部使用）、`AuditService.List`

**Steps:**
- [ ] Step 1: 設計 `audit_logs` 表與 Ent schema。
- [ ] Step 2: 在 usecase 層統一寫入 audit log。
- [ ] Step 3: 實作查詢 API（依時間、操作類型、資源篩選）。

**Acceptance Criteria:**
- [ ] 登入、下單、主檔異動皆記錄 audit log。
- [ ] super / company_admin 可查詢所屬範圍 audit log。

---

### Task 2.7: 公司/部門/使用者 前端頁面

**Goal:** Web 中台管理介面。

**Files:**
- Create: `frontend/src/routes/admin/companies.tsx`
- Create: `frontend/src/routes/admin/departments.tsx`
- Create: `frontend/src/routes/admin/users.tsx`

**Acceptance Criteria:**
- [ ] super 可管理公司與部門。
- [ ] 管理人員名單支援角色指派與停用。
- [ ] 公司表單可維護客戶編號前綴（`customer_code_prefix`）。

---

### Task 2.8: 公司識別設定前端頁面

**Goal:** Web 中台設定公司 Logo、主色、公開資訊。

**Files:**
- Create: `frontend/src/routes/admin/company-branding.tsx`

**Acceptance Criteria:**
- [ ] 可上傳 Logo 並預覽。
- [ ] 可設定主色並即時套用。

---

### Task 2.9: 角色與功能權限 API

**Goal:** 建立 `roles` / `role_permissions` 表與管理 API，內建角色與權限為預設值。

**Files:**
- Create: `backend/internal/domain/roles/*`
- Create: `backend/database/migrations/000XX_roles_seed.sql`

**Interfaces:**
- Connect-RPC：`RoleService`（List / Get / Create / Update / Delete / GetPermissions / UpdatePermissions）
- `AbilityService.GetAbility` 改由 `role_permissions` 驅動

**Steps:**
- [ ] Step 1: 建立 `roles` schema（`code`、`name`、`data_scope`、`is_system`、`is_active`）與 `role_permissions` schema（`role_id`、`resource`、`action`）。
- [ ] Step 2: migration seed 七個內建角色（super/company_admin/dept_admin/staff/customer/guest/developer）與其功能權限（依規格書 3.2 定義；developer 角色已由 Phase 1 Task 1.11 建立）。
- [ ] Step 3: 角色 CRUD API；內建角色（`is_system = true`）不可刪除、不可修改 `code` 與 `data_scope`；自訂角色必須指定 `data_scope`（`company` 或 `department`）。
- [ ] Step 4: 角色功能權限編輯 API（resource × action 矩陣）。
- [ ] Step 5: `AbilityService.GetAbility` 改為依 `role_permissions` 動態產生 CASL JSON。
- [ ] Step 6: RLS 注入 `app.current_data_scope`，依角色 `data_scope` 決定資料範圍。

**Acceptance Criteria:**
- [ ] migration 後七個內建角色與預設權限存在。
- [ ] 內建角色不可刪除；自訂角色可建立、停用、軟刪除。
- [ ] 修改角色權限後，ability JSON 即時反映。

---

### Task 2.10: API 權限管理 API（Casbin policy）

**Goal:** 以 Web 管理 Casbin policy，預設值依角色定義 seed。

**Files:**
- Create: `backend/internal/domain/policies/*`
- Create: `backend/database/migrations/000XX_casbin_policies_seed.sql`

**Interfaces:**
- Connect-RPC：`PolicyService`（List / Add / Delete；p 規則：role × domain × 資源路徑 × 動作，資源路徑為 REST path 或 Connect RPC method path）
- `PolicyService.ListGrouping`（g 規則檢視）

**Steps:**
- [ ] Step 1: migration seed 預設 p 規則（依規格書 3.2 各角色 API 權限）。
- [ ] Step 2: policy CRUD API，異動後即時生效（`e.AddPolicy` / `e.RemovePolicy`，不需重啟）。
- [ ] Step 3: 權限範圍控制：`super` 全域；`company_admin` 僅限自己公司 domain 的 policy。
- [ ] Step 4: 防鎖死檢查：不可移除操作者自身角色的最後一個管理權限。
- [ ] Step 5: policy 異動寫入 `audit_logs`。

**Acceptance Criteria:**
- [ ] 預設 policy 存在且與角色定義一致。
- [ ] 新增/刪除 policy 後即時生效。
- [ ] `company_admin` 無法修改其他公司 domain 的 policy。

---

### Task 2.11: 權限設置前端頁面

**Goal:** Web 中台角色權限與 API 權限管理介面。

**Files:**
- Create: `frontend/src/routes/admin/roles.tsx`
- Create: `frontend/src/routes/admin/api-policies.tsx`

**Acceptance Criteria:**
- [ ] 角色權限設置頁：角色列表、功能權限矩陣編輯、新增自訂角色。
- [ ] API 權限設置頁：policy 列表、依角色/domain 篩選、新增/刪除規則。
- [ ] `company_admin` 僅見自己公司 domain 的資料。

---

### Task 2.12: Phase 2 驗收

**Goal:** 確認多租戶主檔完整。

**Acceptance Criteria:**
- [ ] 公司/部門/使用者 CRUD 完整。
- [ ] 角色權限與 API 權限可於 Web 調整預設值並即時生效。
- [ ] metadicts 查詢與擴充正確。
- [ ] audit log 記錄完整。
- [ ] 公開發現端點可用。

---

## Phase 3: 業務主檔

### Task 3.1: 客戶主檔 API

**Goal:** 部門客戶 CRUD。

**Files:**
- Create: `backend/internal/domain/customers/*`

**Interfaces:**
- Connect-RPC：`CustomerService`（List / Get / Create / Update / Delete）

**Steps:**
- [ ] Step 1: 設計 `customers` Ent schema（含 `default_sales_rep_id` 預設負責業務）。
- [ ] Step 2: 實作 CRUD 與軟刪除。
- [ ] Step 3: 實作依關鍵字篩選。
- [ ] Step 4: 實作 `customer_counters` 樂觀鎖取號：建立客戶時自動產生 `customer_code`（公司定義前綴 + 自增 ID，序號補零至 6 位，公司內唯一，建立後不可修改），取號與建檔同一資料庫交易（同 6.5 訂單編號機制）。
- [ ] Step 5: 建立客戶時同步建立**主帳號 + 1 個業務子帳號**（主帳號 `account_name` = 客戶名稱、`is_primary = true`；業務子帳號 `is_primary = false`、`account_name` 預設「客戶名稱（業務）」），各產生臨時密碼（24 小時效期、首登強制修改，見規格 §4.2）；主帳號帳密交付店家（業務轉交）、**業務子帳號憑證交付所屬業務**（`default_sales_rep_id`）；店家另可自行新增子帳號（如主廚）；業務子帳號於店家管理清單顯示「系統預設（業務使用）」灰化（不可管理），主責業務變更時後台移交。
- [ ] Step 6: `customers` 補 `preferred_delivery_days`（星期一到六核取）與 `promo_tag_ids`（套用促銷分類）欄位維護。

**Acceptance Criteria:**
- [ ] dept_admin / staff 可管理所屬部門客戶。
- [ ] 建立客戶時自動產生客戶編號、主帳號與業務子帳號（兩組臨時密碼）；店家可另新增子帳號。

---

### Task 3.2: 客戶地址簿與聯絡人 API

**Goal:** 管理客戶多筆地址與聯絡人。

**Files:**
- Create: `backend/internal/domain/customers/addresses.go`
- Create: `backend/internal/domain/customers/contacts.go`

**Interfaces:**
- Connect-RPC：`CustomerService`（ListAddresses / AddAddress / UpdateAddress / DeleteAddress）
- Connect-RPC：`CustomerService`（ListContacts / AddContact / UpdateContact / DeleteContact）

**Acceptance Criteria:**
- [ ] 客戶可有多筆地址與聯絡人。
- [ ] 可設定預設地址/聯絡人。

---

### Task 3.3: 商品主檔 API

**Goal:** 部門商品 CRUD 與單位換算。

**Files:**
- Create: `backend/internal/domain/products/*`

**Interfaces:**
- Connect-RPC：`ProductService`（List / Get / Create / Update / Delete）

**Steps:**
- [ ] Step 1: 設計 `products` schema（含 code、unit conversion、warehouses、cutting specs）。
- [ ] Step 2: 實作單位換算儲存與計算。
- [ ] Step 3: 實作軟刪除與分類關聯。

**Acceptance Criteria:**
- [ ] 商品可設定多組單位換算率。
- [ ] 商品可選擇產品分庫與揀貨倉別。

---

### Task 3.4: 倉庫 / 車次 / 分切規格 / 商品分類 API

**Goal:** 部門級字典與主檔管理。

**Files:**
- Create: `backend/internal/domain/warehouses/*`
- Create: `backend/internal/domain/routes/*`
- Create: `backend/internal/domain/cuttingspecs/*`
- Create: `backend/internal/domain/productcategories/*`

**Acceptance Criteria:**
- [ ] 各部門可獨立管理自己的倉庫、車次、分切規格、商品分類。

---

### Task 3.5: 客戶專屬商品清單 API

**Goal:** 管理 `customer_products`。

**Files:**
- Create: `backend/internal/domain/customerproducts/*`

**Interfaces:**
- Connect-RPC：`CustomerProductService`（List / Add / Update / Delete）

**Steps:**
- [ ] Step 1: 設計 `customer_products` schema（alias_name、default_qty、cut_note、promo_tag_ids；不存單價）。
- [ ] Step 2: 實作軟刪除（`deleted_at`）。
- [ ] Step 3: 實作下單時自動建立別名。

**Acceptance Criteria:**
- [ ] 一個客戶對一個商品只能有一個別名。
- [ ] 數量設為 0 時保留但後續單據不顯示。

---

### Task 3.6: 檔案資產 API

**Goal:** 上傳與管理 Logo、公告圖片、列印 PDF。

**Files:**
- Create: `backend/internal/domain/fileassets/*`

**Interfaces:**
- `POST /api/v1/files/upload`
- `GET /api/v1/files/:id/download`

**Steps:**
- [ ] Step 1: 實作本地 volume 儲存。
- [ ] Step 2: 上傳限制：MIME 白名單（jpeg / png / webp ≤ 5 MB、pdf ≤ 10 MB），副檔名與 magic bytes 雙重檢查。
- [ ] Step 3: 記錄 `file_assets` 元資料。
- [ ] Step 4: 提供下載 URL。

**Acceptance Criteria:**
- [ ] 可上傳圖片並取得 URL。
- [ ] 可下載已上傳檔案。

---

### Task 3.7: 業務主檔前端頁面

**Goal:** Web 中台客戶與商品管理介面。

**Files:**
- Create: `frontend/src/routes/admin/customers.tsx`
- Create: `frontend/src/routes/admin/products.tsx`
- Create: `frontend/src/routes/admin/warehouses.tsx`
- Create: `frontend/src/routes/admin/routes.tsx`
- Create: `frontend/src/routes/admin/cutting-specs.tsx`

**Acceptance Criteria:**
- [ ] DataTable 支援分頁、排序、篩選。
- [ ] Sheet 表單可新增/編輯。

---

### Task 3.8: QR Code 產生與下載

**Goal:** Web 中台產生客戶登入 QR Code。

**Files:**
- Create: `backend/internal/domain/customers/qrcode.go`
- Create: `frontend/src/components/customer-qrcode.tsx`

**Acceptance Criteria:**
- [ ] QR Code 內含簽章 token（company_id + customer_code + exp）。
- [ ] 可下載/分享 QR Code。

---

### Task 3.9: Phase 3 驗收

**Goal:** 確認業務主檔完整。

**Acceptance Criteria:**
- [ ] 客戶、商品、倉庫、車次、分切規格 CRUD 完整。
- [ ] 客戶專屬商品清單運作正常。
- [ ] QR Code 可被 App 正確掃描登入。

---

## Phase 4: 訂單與通知

### Task 4.1: 訂單 API

**Goal:** 銷售訂單 CRUD 與狀態管理。

**Files:**
- Create: `backend/internal/domain/salesorders/*`

**Interfaces:**
- Connect-RPC：`SalesOrderService`（List / Get / Create / Update / Cancel / Complete / Void / ListEvents）

**Steps:**
- [ ] Step 1: 設計 `sales_orders` / `sales_order_items` / `sales_order_events` schema（含 `source`、`voided` 狀態；不存任何金額欄位）。
- [ ] Step 2: 設計 `order_counters` schema（`company_id` + `source` + `next_seq` + `version`）。
- [ ] Step 3: 實作訂單編號取號：樂觀鎖更新 `order_counters`（version 衝突重試），取號與訂單建立同一交易；編號 = 來源碼 + 6 位補零序號（如 `W000123`）。
- [ ] Step 4: 實作狀態機（含 `voided` 終態；1.0 不儲存任何金額欄位，見規格書 6.5）；所有狀態異動寫入 `sales_order_events`。
- [ ] Step 5: 實作軟刪除與取消邏輯。

**Acceptance Criteria:**
- [ ] 待處理訂單可編輯、可取消。
- [ ] 處理中訂單不可編輯。
- [ ] 併發建立訂單時編號不重複（樂觀鎖重試生效）。
- [ ] 訂單與明細不含任何金額欄位；狀態流轉正確。
- [ ] 作廢需 dept_admin 以上並填寫原因，作廢後為終態；異動記錄可查。

---

### Task 4.2: 下單流程與單位換算

**Goal:** 支援業務/客戶下單、手打商品、別名建立。

**Files:**
- Update: `backend/internal/domain/salesorders/usecase.go`

**Steps:**
- [ ] Step 1: 下單時依選擇單位自動換算數量。
- [ ] Step 2: 業務手打商品名稱時，詢問是否儲存為客戶別名。
- [ ] Step 3: 客戶只能從自己的 `customer_products` 下單。
- [ ] Step 4: 下單時記錄 `source`（Web / App）；1.0 不儲存金額欄位。
- [ ] Step 5: 選擇預計出貨日時依 `customers.preferred_delivery_days` 自動順延至下一個勾選日（未勾選則維持原選擇）。

**Acceptance Criteria:**
- [ ] 單位換算計算正確。
- [ ] 非偏好送貨日下單自動順延至下一個勾選日。
- [ ] 手打商品可自動建立 `customer_products`。

---

### Task 4.3: 通知系統 API

**Goal:** 通知範本與通知記錄。

**Files:**
- Create: `backend/internal/domain/notifications/*`

**Interfaces:**
- Connect-RPC：`NotificationTemplateService`（List / Get / Create / Update / Delete）
- Connect-RPC：`NotificationService`（List / MarkRead）

**Steps:**
- [ ] Step 1: 設計 `notification_templates` / `notifications` / `user_devices` schema。
- [ ] Step 2: 實作範本渲染（變數替換）。
- [ ] Step 3: 實作通知記錄與已讀。
- [ ] Step 4: 實作 `DeviceService`（Register / Unregister）；FCM 回報 token 失效時刪除對應 `user_devices`。
- [ ] Step 5: 通知範本與記錄限 `fcm` / `in_app` 兩通道（無 Email）；路由：業務下單推客戶全部子帳號（主帳號不接收業務通知）、後台新增專屬商品推主責業務（檢核待定）、退貨審核結果推發起帳號。
- [ ] Step 6: 促銷推播資料：`promo_tags` 部門級標籤、商品 / 專屬商品 `promo_tag_ids`、客戶 `promo_tag_ids` 套用。

**Acceptance Criteria:**
- [ ] 可依範本產生通知。
- [ ] 通知狀態正確流轉（pending → sent → read）。
- [ ] App 可註冊 / 註銷裝置 token；失效 token 自動清除。

---

### Task 4.4: FCM / 站內通知發送

**Goal:** 整合 FCM 推播與站內通知（1.0 無 Email / SMTP）。

**Files:**
- Create: `backend/internal/notification/fcm.go`
- Create: `backend/internal/notification/inapp.go`

**Steps:**
- [ ] Step 1: 整合 FCM SDK。
- [ ] Step 2: 實作站內通知（`notifications` 記錄 + 通知中心 API）。
- [ ] Step 3: 訂單建立觸發 FCM + 站內：業務下單時推播該客戶全部**子帳號**（主帳號為管理用途、不接收業務通知），客戶自行下單不另行通知；後台新增客戶專屬商品推播主責業務（`default_sales_rep_id` → dept_admin，檢核待定）；退貨審核結果推播發起帳號。
- [ ] Step 4: 發送失敗不重試，`notifications.status` 標記 `failed` 並記錄原因。

**Acceptance Criteria:**
- [ ] 業務下單後客戶子帳號收到 FCM 推播與站內通知（主帳號不收）。
- [ ] 後台新增專屬商品後主責業務收到推播。
- [ ] 失敗記錄可查；FCM 失效 token 自動刪除。

---

### Task 4.5: Web 下單頁面

**Goal:** Web 中台訂單管理與下單。

**Files:**
- Create: `frontend/src/routes/admin/sales-orders.tsx`
- Create: `frontend/src/routes/admin/sales-orders/new.tsx`

**Acceptance Criteria:**
- [ ] 可選擇客戶並帶出 `customer_products`。
- [ ] 可新增/刪除訂單明細。
- [ ] 提交後發送通知。

---

### Task 4.6: App 快速下單

**Goal:** App「商品」頁快速下單流程。

**Files:**
- Create: `app/lib/layer_presentation/stories/admin/tabs/salesorder/salesorder_screen.dart`

**Steps:**
- [ ] Step 1: 業務選擇客戶後帶出 `customer_products`。
- [ ] Step 2: 客戶直接從自己的清單選擇商品。
- [ ] Step 3: 輸入數量/單位/規格/備註，提交訂單。

**Acceptance Criteria:**
- [ ] 業務/客戶皆可從「商品」頁下單。
- [ ] 下單後可在「訂單歷史」查看。

---

### Task 4.7: 退貨申請與審核

**Goal:** 客戶發起退貨、業務審核、退貨證明（規格 §6.6）。

**Files:**
- Create: `backend/internal/domain/returns/*`
- Update: `sales-order-app` 訂單歷史 / 專屬商品頁（退貨入口）

**Interfaces:**
- Connect-RPC：`ReturnService`（Create / List / Review）

**Steps:**
- [ ] Step 1: 設計 `return_requests` / `return_request_items` schema（status：`pending` / `approved` / `rejected`）。
- [ ] Step 2: 客戶端發起退貨：從歷史訂單勾選品項 或 從專屬商品清單選擇（兩者並存），填數量/原因/備註、上傳照片（file_assets）。
- [ ] Step 3: 業務審核（approved / rejected，拒絕需原因）；通過後客戶端顯示退貨證明。
- [ ] Step 4: 審核結果推播發起帳號；申請與審核寫入稽核日誌；不修改原訂單狀態。
- [ ] Step 5: 公司端不建置配送頁面（取貨由業務自行通知配送）。

**Acceptance Criteria:**
- [ ] 客戶可從兩種來源發起退貨並附照片。
- [ ] 業務審核後客戶端顯示退貨證明。
- [ ] 原訂單狀態不變；稽核與推播正確。

---

### Task 4.8: Phase 4 驗收

**Goal:** 確認訂單與通知完整。

**Acceptance Criteria:**
- [ ] Web / App 皆可建立訂單。
- [ ] 訂單狀態正確（不含金額欄位）。
- [ ] 訂單建立後客戶收到 FCM / 站內通知。

---

## Phase 5: 派車與單據列印

### Task 5.1: 派車 API

**Goal:** 指派訂單車次與配送順序，確認派車。

**Files:**
- Create/Update: `backend/internal/domain/salesorders/dispatch.go`

**Interfaces:**
- Connect-RPC：`DispatchService.AssignRoute`（指派車次與配送順序）
- Connect-RPC：`DispatchService.Confirm`（批次派車確認）
- Connect-RPC：`DispatchService.CancelDispatch`（取消派車，需原因）

**Steps:**
- [ ] Step 1: 更新 `route_id` / `delivery_sequence`，樂觀鎖比對 `sales_orders.version`（衝突回傳錯誤，前端重新整理看板）。
- [ ] Step 2: 派車確認以車次批次執行：一次確認該車次當日所有待派訂單，逐筆更新 `dispatched_at` / `dispatched_by` / `status = processing`。
- [ ] Step 3: 發送派車通知。
- [ ] Step 4: 取消派車：dept_admin 以上，需填寫原因；清除 `dispatched_at` / `dispatched_by`、狀態退回 `pending`，保留 `route_id` / `delivery_sequence`；該車次已正式列印時回傳警告提示需重新列印。

**Acceptance Criteria:**
- [ ] 訂單可指派車次與順序。
- [ ] 派車確認後狀態變為處理中。
- [ ] 取消派車後訂單退回待處理，原因寫入 `sales_order_events` 與 `audit_logs`。

---

### Task 5.2: Connect 串流即時派車看板

**Goal:** Web 派車看板即時更新（Connect server streaming）。

**Files:**
- Create: `backend/internal/domain/salesorders/watch_board.go`
- Create: `frontend/src/routes/admin/dispatch.tsx`

**Steps:**
- [ ] Step 1: proto 定義 `DispatchService.WatchBoard` streaming RPC 與 `BoardEvent`（含 `sales_order_id`、`route_id`、`delivery_sequence`、`version`），實作串流 handler 與部門訂閱註冊，cookie／JWT 走既有 auth interceptor（不實作一次性 ticket）。
- [ ] Step 2: 派車 mutation（AssignRoute／Confirm／CancelDispatch）提交後發佈事件至 Valkey pub/sub（以 `department_id` 分 channel），各 replica 串流 handler 訂閱轉發本地連線；server 定期 heartbeat 防 ingress idle timeout。
- [ ] Step 3: 前端看板以 connect-web 訂閱，事件觸發看板查詢 invalidate 全量重查（不做 cache patch）；自身 mutation 成功後同樣 invalidate。
- [ ] Step 4: 斷線以 exponential backoff 重連，重連後全量重查；連續失敗達閾值降級 30 秒輪詢（視窗隱藏暫停、聚焦重查）；樂觀鎖衝突提示並重新載入看板資料。

**Acceptance Criteria:**
- [ ] 同部門其他連線即時收到看板事件（含跨 replica）。
- [ ] 斷線重連後看板為最新狀態、串流不可用時降級輪詢可用。
- [ ] 樂觀鎖衝突提示並重載。

---

### Task 5.3: 四種單據模板

**Goal:** 使用 HTML/CSS 建立四種單據模板。

**Files:**
- Create: `backend/internal/print/templates/dispatch_summary.html`
- Create: `backend/internal/print/templates/delivery_note.html`
- Create: `backend/internal/print/templates/picking_list.html`
- Create: `backend/internal/print/templates/processing_list.html`

**Acceptance Criteria:**
- [ ] 單車總表依車次分組，不顯示金額。
- [ ] 對點單依車次 → 店家分組，不顯示價格。
- [ ] 揀貨單依車次 → 倉別 → 商品分類 → 商品名稱排序。
- [ ] 加工單分為加工室揀 / 配送揀兩區塊；「加工後數量」欄位列印為空白供手寫回填。

---

### Task 5.4: Gotenberg PDF 產生服務

**Goal:** 整合 Gotenberg 將 HTML 轉為 PDF。

**Files:**
- Create: `backend/internal/print/service.go`
- Create: `backend/internal/print/client.go`

**Steps:**
- [ ] Step 1: 實作 Gotenberg client。
- [ ] Step 2: 依車次/日期/單據類型組合資料。
- [ ] Step 3: 產生 PDF 並關聯 `file_assets`。

**Acceptance Criteria:**
- [ ] 空表不產生 PDF。
- [ ] PDF 使用 Noto Sans CJK TC 字型。

---

### Task 5.5: 列印與預覽 API

**Goal:** 正式列印與預覽記錄。

**Files:**
- Create: `backend/internal/domain/prints/*`

**Interfaces:**
- Connect-RPC：`PrintService.Preview`、`PrintService.Print`、`PrintService.ListLogs`（PDF 經 `file_assets` 以下載 URL 取得）

**Steps:**
- [ ] Step 1: 實作 `print_logs` / `print_previews` 記錄。
- [ ] Step 2: 正式列印檢查訂單狀態為 `processing`。
- [ ] Step 3: 重印需填寫原因。

**Acceptance Criteria:**
- [ ] 預覽不影響正式列印記錄。
- [ ] 重印記錄 `is_reprint = true` 與 `reprint_reason`。

---

### Task 5.6: Web 派車與列印頁面

**Goal:** Web 中台派車規劃與單據列印介面。

**Files:**
- Create: `frontend/src/routes/admin/dispatch-planning.tsx`

**Acceptance Criteria:**
- [ ] Kanban 可拖放調整車次與順序。
- [ ] 可依預計出貨日篩選。
- [ ] 可預覽與列印四種單據。

---

### Task 5.7: Phase 5 驗收

**Goal:** 確認派車與列印完整。

**Acceptance Criteria:**
- [ ] 派車看板即時同步。
- [ ] 四種單據可正確產生 PDF。
- [ ] 列印記錄完整。

---

## Phase 6: App 功能

### Task 6.1: App 首頁

**Goal:** 公告輪播、最新消息、快速下單入口。

**Files:**
- Create/Update: `app/lib/layer_presentation/stories/admin/tabs/home/home_screen.dart`

**Acceptance Criteria:**
- [ ] 顯示公告 Banner 輪播。
- [ ] 顯示最新消息列表。
- [ ] 依所屬公司顯示 Logo 與主色。

---

### Task 6.2: App 訂單歷史

**Goal:** 查看自己的訂單歷史，支援下拉重新整理。

**Files:**
- Create/Update: `app/lib/layer_presentation/stories/admin/tabs/history/history_screen.dart`

**Acceptance Criteria:**
- [ ] 業務看到自己部門的訂單。
- [ ] 客戶只看到自己的訂單。
- [ ] 下拉重新整理運作正常。

---

### Task 6.3: App 功能頁

**Goal:** 客戶快速查詢、QR Code、關於我們、設定、登出。

**Files:**
- Create/Update: `app/lib/layer_presentation/stories/admin/tabs/profile/profile_screen.dart`

**Acceptance Criteria:**
- [ ] 可查看公司公開資訊。
- [ ] 業務可展示客戶 QR Code。
- [ ] 可登出並清除 secure storage。

---

### Task 6.4: App 離線快取

**Goal:** 快取登入資訊、客戶列表、商品列表、訂單歷史。

**Files:**
- Create/Update: `app/lib/layer_data/local/*`

**Acceptance Criteria:**
- [ ] 離線時可查看已快取資料。
- [ ] 下單時若無網路顯示錯誤，不允許提交。

---

### Task 6.5: App FCM 推播

**Goal:** 接收訂單狀態更新與派車通知。

**Files:**
- Create: `app/lib/layer_business/services/fcm_service.dart`

**Steps:**
- [ ] Step 1: 整合 firebase_messaging。
- [ ] Step 2: 註冊 FCM token 至後端。
- [ ] Step 3: 處理前景/背景/終止狀態推播。

**Acceptance Criteria:**
- [ ] App 可收到並顯示 FCM 推播。

---

### Task 6.6: App 新增客戶（手動表單）

**Goal:** 業務於 App 直接建立客戶主檔與登入帳號（規格 §9.4，v1.0.26 納入 1.0）。

**Files:**
- Create/Update: `app/lib/layer_presentation/stories/admin/customers/*`

**Interfaces:**
- Connect-RPC：`CustomerService.Create`（回傳 `customer_code`、主帳號與業務子帳號的名稱與臨時密碼）

**Steps:**
- [ ] Step 1: 建立新增客戶表單（主檔必填欄位、地址、聯絡人）。
- [ ] Step 2: 呼叫 `CustomerService.Create`，後端自動產生 `customer_code`（公司前綴 + 自增 ID）與臨時密碼。
- [ ] Step 3: 成功頁顯示主帳號帳密（供業務轉交店家）、業務子帳號帳密（供所屬業務留存使用）與管理網址鏈結（`customer_account_manage` 深層連結，可顯示為 QR Code）。

**Acceptance Criteria:**
- [ ] staff / dept_admin 可於 App 新增客戶。
- [ ] 系統自動產生前綴 + 自增 ID 的登入帳號與 24 小時效期臨時密碼。
- [ ] 臨時密碼僅於建立成功時顯示一次。

---

### Task 6.7: 店家帳號管理（主帳號自助）

**Goal:** 店家以主帳號於 App 管理子帳號（規格 §4.2 / §9.3）。

**Files:**
- Create: `backend/internal/domain/customeraccounts/*`（self 範圍 API）
- Update: `sales-order-app` 功能頁（帳號管理入口）

**Interfaces:**
- Connect-RPC：`CustomerAccountService`（List / Add / Deactivate / ResetPassword），僅主帳號可呼叫、範圍僅限自己客戶

**Steps:**
- [ ] Step 1: 店家自助 API（限主帳號、data_scope self：僅能操作自己客戶底下的帳號；子帳號無管理權限）；主帳號 session 僅授權帳號管理，業務 API 一律 403。
- [ ] Step 2: 新增子帳號（填寫帳號名稱、`is_primary = false`、臨時密碼 24h、首登強改）與重置子帳號密碼。
- [ ] Step 3: 停用子帳號；防呆：主帳號不可由店家停用 / 重置（含當前登入帳號）。
- [ ] Step 4: App 帳號管理頁（列表 / 新增 / 停用 / 重置）；主帳號登入僅顯示帳號管理畫面；自動附帶的業務子帳號顯示為「系統預設（業務使用）」並灰化（反白），店家不可改名 / 停用 / 重置；QR 登入清單排除主帳號與業務子帳號。
- [ ] Step 5: 處理帳號管理深層連結（`customer_account_manage`）路由 → 導向帳號管理登入；未安裝 App 導向商店。
- [ ] Step 6: 帳號管理異動寫稽核；後台 dept_admin 逃生門（可停用 / 重置任何店家帳號，含重置主帳號密碼轉交）；後台停用主帳號連鎖停用子帳號。

**Acceptance Criteria:**
- [ ] 主帳號登入僅能帳號管理，業務功能不可用；店家以主帳號可新增 / 停用 / 重置自建子帳號，子帳號無法管理帳號；自動附帶的業務子帳號灰化不可管理且店家無其密碼；後台可移交業務子帳號。
- [ ] 主帳號不可由店家停用 / 重置；後台逃生門可用；停用主帳號連鎖子帳號。
- [ ] 帳號管理異動皆有稽核記錄。

---

### Task 6.8: Phase 6 驗收

**Goal:** 確認 App 功能完整。

**Acceptance Criteria:**
- [ ] App 四個底部頁籤運作正常。
- [ ] 業務/客戶皆可下單。
- [ ] 業務可於 App 新增客戶並交付登入帳號與臨時密碼。
- [ ] 推播可正常接收。
- [ ] 店家可於 App 自行管理名下帳號。

---

## Phase 7: 公告與 UI 強化

### Task 7.1: 公告內容管理 API

**Goal:** Banner / 最新消息 / 圖文文章 CRUD。

**Files:**
- Create: `backend/internal/domain/articles/*`

**Interfaces:**
- Connect-RPC：`AnnouncementService`（List / Get / Create / Update / Delete）

**Steps:**
- [ ] Step 1: 設計 articles schema（title、content、slug、image、platform、category）。
- [ ] Step 2: 實作圖片上傳與 WebP 處理。
- [ ] Step 3: 實作依平台（Web/App）篩選。

**Acceptance Criteria:**
- [ ] super / company_admin / dept_admin 可依範圍管理公告。
- [ ] App 只收到標記為 App 的公告。

---

### Task 7.2: Web 中台導覽與主題

**Goal:** 側邊欄動態群組、主題切換、通知鈴。

**Files:**
- Create: `frontend/src/components/layout/sidebar.tsx`
- Create: `frontend/src/components/layout/theme-switcher.tsx`
- Create: `frontend/src/components/layout/notification-bell.tsx`

**Acceptance Criteria:**
- [ ] 側邊欄依角色動態顯示選單群組。
- [ ] 可切換 Light / Dark / System 主題。
- [ ] 通知鈴顯示未讀通知數量。

---

### Task 7.3: Dashboard 頁面

**Goal:** 今日待出貨、待處理訂單數量、快速連結。

**Files:**
- Create: `frontend/src/routes/admin/dashboard.tsx`

**Acceptance Criteria:**
- [ ] 顯示今日待出貨訂單數。
- [ ] 顯示待處理訂單數。
- [ ] 提供快速連結至常用功能。

---

### Task 7.4: 促銷推播與公告分離

**Goal:** 促銷分類標籤、客戶套用、依分類選群推播（規格 §10.1）。

**Files:**
- Create: `backend/internal/domain/promotions/*`
- Update: `sales-order-frontend` 客戶主檔與商品頁（標籤）、促銷推播頁

**Interfaces:**
- Connect-RPC：`PromoService`（Tag CRUD / Assign / Broadcast）

**Steps:**
- [ ] Step 1: `promo_tags` 部門級標籤 CRUD。
- [ ] Step 2: 商品 / 客戶專屬商品標記 `promo_tag_ids`；客戶主檔套用 `promo_tag_ids`。
- [ ] Step 3: 後台促銷推播頁：依分類選客戶群，經 FCM + 站內發送。
- [ ] Step 4: App 首頁消息區顯示公告（所有客戶可見，與推播分離）。

**Acceptance Criteria:**
- [ ] 可依分類標籤選群推播，未套用標籤的客戶不收到。
- [ ] 公告於首頁消息顯示，與推播互不取代。

---

### Task 7.5: Phase 7 驗收

**Goal:** 確認公告與 UI 完整。

**Acceptance Criteria:**
- [ ] 公告管理與顯示正常。
- [ ] 側邊欄依公司/角色動態呈現。
- [ ] 主題切換生效。
- [ ] 促銷推播依分類選群正確送達。

---

## Phase 8: 部署、安全與維運

### Task 8.1: Kubernetes manifests

**Goal:** 建立 k8s 部署設定。

**Files:**
- Create: `infra/k8s/backend-deployment.yaml`
- Create: `infra/k8s/frontend-deployment.yaml`
- Create: `infra/k8s/postgres-statefulset.yaml`
- Create: `infra/k8s/valkey-statefulset.yaml`
- Create: `infra/k8s/traefik-ingress.yaml`

**Acceptance Criteria:**
- [ ] `kubectl apply -f infra/k8s` 可部署完整系統。
- [ ] Traefik 正確轉發 REST 與 Connect-RPC（含 server streaming）。

---

### Task 8.2: GitHub Actions CD 管線

**Goal:** 自動化建置與部署。

**Files:**
- Create: `.github/workflows/cd.yml`

**Steps:**
- [ ] Step 1: 建置 backend / frontend Docker image。
- [ ] Step 2: 推送至 Artifact Registry。
- [ ] Step 3: 部署至 k8s。

**Acceptance Criteria:**
- [ ] 合併至 main 後自動部署至 staging。
- [ ] 手動觸發可部署至 production。

---

### Task 8.3: 備份策略實作

**Goal:** PostgreSQL / Valkey / 檔案備份至 GCS。

**Files:**
- Create: `infra/k8s/backup-cronjob.yaml`
- Create: `backend/scripts/backup-postgres.sh`
- Create: `backend/scripts/backup-valkey.sh`

**Acceptance Criteria:**
- [ ] 每日備份成功上傳 GCS。
- [ ] 可從 GCS 還原至測試環境。

---

### Task 8.4: 監控與告警

**Goal:** 基礎設施與應用監控。

**Files:**
- Create: `infra/k8s/monitoring/prometheus.yaml`
- Create: `infra/k8s/monitoring/grafana.yaml`
- Create: `infra/k8s/monitoring/alertmanager.yaml`
- Create: `infra/k8s/monitoring/alerts.yaml`（告警規則）

**Steps:**
- [ ] Step 1: 部署 Prometheus + Grafana + Alertmanager 至同叢集（規格書 12.3.2 定案）。
- [ ] Step 2: 後端暴露 `/metrics`（API 延遲、錯誤率、請求量、Connect-RPC 狀態碼）。
- [ ] Step 3: 建立 Grafana 儀表板：基礎設施（CPU / 記憶體 / 磁碟 / Pod 重啟）與業務指標（登入失敗、訂單建立量、列印次數）。
- [ ] Step 4: 建立 Alertmanager 告警規則：服務不可用、錯誤率驟升、備份失敗、磁碟空間不足；通道 Email / Slack / Webhook。

**Acceptance Criteria:**
- [ ] 可查看 CPU / 記憶體 / 請求延遲儀表板。
- [ ] 備份失敗或錯誤率驟升時發送告警。

---

### Task 8.5: 安全強化

**Goal:** 實作規格書第 14 章安全與合規。

**Files:**
- Update: `backend/internal/server/server.go`（安全標頭）
- Update: `backend/internal/middleware/ratelimit.go`
- Update: `frontend/src/components/rich-text-editor.tsx`（XSS sanitize）

**Steps:**
- [ ] Step 1: 啟用 HSTS、CSP、X-Frame-Options。
- [ ] Step 2: 實作 rate limit middleware。
- [ ] Step 3: 確保所有通訊走 TLS 1.2+。
- [ ] Step 4: 實作 Rich Text sanitize。
- [ ] Step 5: 實作資料保留排程：`notifications` 180 天、`print_previews` 90 天、`print_logs` 2 年排程刪除；`audit_logs` 預設保留 3 個月、管理頁可設定 1 / 3 / 6 / 12 個月或永久，到期排程刪除（D27）。

**Acceptance Criteria:**
- [ ] 生產環境通過基本安全檢查清單。

---

### Task 8.6: 災難復原演練

**Goal:** 驗證 RTO/RPO 目標。

**Files:**
- Create: `docs/runbooks/disaster-recovery.md`
- Create: `docs/runbooks/pitr-restore.md`（PostgreSQL WAL 歸檔 PITR 還原手冊）

**Steps:**
- [ ] Step 1: 撰寫 PITR runbook：從 GCS 備份與 WAL 歸檔還原 PostgreSQL 至指定時間點。
- [ ] Step 2: 上線前完成首次 PITR 還原演練（自建 StatefulSet 必要驗證，見規格書 12.3.3）。
- [ ] Step 3: 演練 Valkey 從 RDB/AOF 備份還原。
- [ ] Step 4: 演練後檢核 RTO 4 小時 / RPO 1 小時是否達成，更新復原手冊。

**Acceptance Criteria:**
- [ ] 演練可在 4 小時內還原核心服務。
- [ ] 資料遺失不超過 1 小時。
- [ ] 上線前首次 PITR 還原演練完成並留存紀錄。

---

### Task 8.7: Big Bang 上線準備

**Goal:** 完成上線前檢查清單。

**Files:**
- Create: `docs/runbooks/production-launch-checklist.md`

**Steps:**
- [ ] Step 1: 準備生產環境變數與憑證。
- [ ] Step 2: 執行最終壓力測試。
- [ ] Step 3: 確認不從舊系統匯入資料。
- [ ] Step 4: 確認生產環境 `DEVELOPER_ACCOUNT_ENABLED=false`（開發者帳號關閉）。
- [ ] Step 5: 安排上線時間與回滾計畫。

**Acceptance Criteria:**
- [ ] 上線檢查清單所有項目通過。

---

### Task 8.8: Phase 8 驗收

**Goal:** 確認系統可上線。

**Acceptance Criteria:**
- [ ] 生產環境部署成功。
- [ ] 備份與監控運作正常。
- [ ] 災難復原演練通過。
- [ ] Big Bang 上線完成。

---

## 附錄：跨 Phase 共用元件

| 元件 | 負責 Phase | 說明 |
|---|---|---|
| DataTable | 2–7 | 分頁、排序、篩選、欄位隱藏 |
| Sheet 表單 | 2–7 | 側邊滑出新增/編輯 |
| File Uploader | 2–7 | 圖片/PDF 上傳 |
| Notification Service | 4 | FCM / 站內通知（無 Email） |
| Audit Logger | 1–8 | 關鍵操作記錄 |
| Ability Provider | 1 | 前後端權限同步 |

---

*計畫版本：v2.7.1*  
*日期：2026-07-17*  
*對應規格書：v1.0.34*
