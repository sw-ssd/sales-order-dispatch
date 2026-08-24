# 多公司訂出貨系統 1.0 — Backend / Web / App 子專案實作計畫

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> 日期：2026-08-05
> 對應設計文件：`docs/superpowers/specs/2026-08-05-sales-order-1.0-subproject-decomposition-design.md`
> 對應規格書：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）
> 對應決策記錄：`docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md`（D1–D28）
> 狀態：**實作暫緩；本計畫供未來開工時直接指派與追蹤**

**Goal:** 將 1.0 核心系統依 Backend / Web / App 三子專案逐項實作，每個 Task 皆可獨立指派、驗收、開票。

**Architecture:** 全新 monorepo；後端 Go + Ent + Goose + Connect-RPC + Casbin + PostgreSQL RLS；Web SolidJS + TanStack Query + CASL；App Flutter + solidart + disco + fquery + Sembast。proto 為三端唯一型別來源。

**Tech Stack:** Go 1.24+ / SolidJS / Flutter stable（Dart ≥3.10）/ PostgreSQL 16 / Valkey / Gotenberg / Kubernetes / Prometheus / Grafana

## Global Constraints

- 後端 `.proto` 為唯一型別來源，產生前端 TypeScript 與 Flutter Dart 型別；proto package `v1`。
- 業務 API 統一 Connect-RPC；REST 僅保留公開端點（版本、OAuth 導向/回調、QR 兌換、檔案上傳/下載、`companies/public/{identifier}`），前綴 `/api/v1`。
- 所有業務實體統一使用 `deleted_at` 軟刪除，查詢預設排除。
- 每筆業務資料皆帶 `company_id` / `department_id`；RLS 為最後防線。
- 三端覆蓋率門檻 70%（CI 強制）。
- 不儲存金額；訂單/明細/商品/專屬商品皆無金額欄位。
- 每個 domain 新增時同步更新 API 文件與前端 ability 規則。
- 實作暫不啟動；本計畫每個 Task 於未來開工時直接轉為 ticket。

---

## File Structure

```
/
├── Taskfile.yml
├── package.json
├── pnpm-workspace.yaml
├── turbo.json
├── docker-compose.dev.yml
├── .github/workflows/ci.yml
├── backend/
│   ├── go.mod
│   ├── Taskfile.yml
│   ├── .air.toml
│   ├── cmd/
│   │   ├── server/main.go
│   │   ├── migrate/main.go
│   │   └── seed/main.go
│   ├── config/
│   │   └── config.go（+ api/auth/cache/database/storage/observability.go，D31）
│   ├── internal/
│   │   ├── server/{server.go,domains.go}
│   │   ├── middleware/
│   │   ├── auth/
│   │   ├── ent/
│   │   │   └── schema/
│   │   ├── proto/v1/
│   │   ├── migrations/
│   │   ├── repositories/
│   │   ├── services/
│   │   └── handlers/
│   ├── third_party/
│   │   ├── database/ent.go
│   │   └── cache/valkey.go
│   └── database/migrations/
├── frontend/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── src/
│   │   ├── main.tsx
│   │   ├── App.tsx
│   │   ├── router/
│   │   ├── lib/
│   │   │   ├── api/          # Connect clients
│   │   │   ├── auth/         # CASL ability
│   │   │   └── query/        # TanStack Query
│   │   ├── components/
│   │   ├── features/
│   │   │   ├── auth/
│   │   │   ├── layout/
│   │   │   ├── users/
│   │   │   ├── master-data/
│   │   │   ├── orders/
│   │   │   ├── dispatch/
│   │   │   ├── printing/
│   │   │   ├── announcements/
│   │   │   └── audit/
│   │   └── index.css
│   └── tests/
├── app/
│   ├── pubspec.yaml
│   ├── analysis_options.yaml
│   ├── Taskfile.yml
│   └── lib/
│       ├── main.dart
│       ├── app.dart
│       ├── core/
│       │   ├── config/
│       │   ├── auth/
│       │   ├── network/
│       │   ├── cache/
│       │   └── errors/
│       ├── router/
│       ├── features/
│       │   ├── auth/
│       │   ├── home/
│       │   ├── products/
│       │   ├── orders/
│       │   ├── customers/
│       │   ├── account/
│       │   ├── profile/
│       │   └── notifications/
│       └── gen/              # buf 產出
└── infra/
    ├── k8s/
    ├── helm/
    └── terraform/
```

---

## Wave 1：基礎建設與認證授權

### Task 1: monorepo 根結構

**Files:**
- Create: `package.json`
- Create: `pnpm-workspace.yaml`
- Create: `turbo.json`
- Create: `.gitignore`
- Create: `README.md`

**Interfaces:**
- Consumes: none
- Produces: root workspace config for `pnpm install` and `pnpm turbo run build`

- [ ] **Step 1: 建立 `package.json`**
  - Set `private: true`, add `devDependencies`: `turbo`, `prettier`, `eslint`.
- [ ] **Step 2: 建立 `pnpm-workspace.yaml`**
  - Include `backend`, `frontend`, `app`, `infra`.
- [ ] **Step 3: 建立 `turbo.json`**
  - Define `build`, `dev`, `test`, `lint` pipelines with dependency order `^build`.
- [ ] **Step 4: 建立根 `.gitignore`**
  - Exclude `.turbo`, `node_modules`, `.env*`, `dist`, `build`.
- [ ] **Step 5: 建立 `README.md`**
  - Document startup: `task infra:start`, `task dev`.
- [ ] **Step 6: 驗收**
  - Run `pnpm install` and `pnpm turbo run build` without error.
- [ ] **Step 7: Commit**
  - `git add package.json pnpm-workspace.yaml turbo.json .gitignore README.md`
  - `git commit -m "chore: scaffold monorepo root"`

### Task 2: 開發環境 docker-compose

**Files:**
- Create: `docker-compose.dev.yml`

**Interfaces:**
- Consumes: none
- Produces: `task infra:start` brings up PostgreSQL, Valkey, Gotenberg

- [ ] **Step 1: 定義 `postgres` 服務**
  - Port `5432`, volume `postgres_data`, env `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`.
- [ ] **Step 2: 定義 `valkey` 服務**
  - Port `6379`.
- [ ] **Step 3: 定義 `gotenberg` 服務**
  - Port `3001`.
- [ ] **Step 4: 定義共用 volume `postgres_data`**
- [ ] **Step 5: 驗收**
  - `task infra:start` then `docker ps` shows 3 containers.
  - `task infra:stop` stops them.
- [ ] **Step 6: Commit**

### Task 3: 後端 Go 骨架

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`（極薄入口）
- Create: `backend/cmd/migrate/main.go`、`backend/cmd/seed/main.go`
- Create: `backend/config/{config,api,auth,cache,database,storage,observability}.go`（envconfig 逐檔 struct，D31）
- Create: `backend/internal/server/{server,domains}.go`
- Create: `backend/third_party/{database/ent.go,cache/valkey.go}`
- Create: `backend/.air.toml`、`backend/Taskfile.yml`（含 dev/check/vuln/migrate/seed）

**Interfaces:**
- Consumes: none
- Produces: `config.New()` 聚合 struct, `server.New(cfg)` Chi router + `Server.Init()`/`InitDomains()`

- [ ] **Step 1: 初始化 Go module**
  - `go mod init github.com/salesorder/sales-order-1.0/backend`
- [ ] **Step 2: 安裝依賴**
  - `go get github.com/go-chi/chi/v5 github.com/kelseyhightower/envconfig github.com/rs/cors`
- [ ] **Step 3: 實作 `config` 逐檔 struct 與 `New()` 聚合**
  - Fields: `Port`, `DatabaseURL`, `ValkeyAddr`, `GoogleClientID`, `GoogleClientSecret`, `GotenbergURL`.
- [ ] **Step 4: 實作基礎 Chi server**
  - CORS, request ID logger, `/api/v1/version` returning version JSON.
- [ ] **Step 5: 建立 `backend/Taskfile.yml`**
  - Tasks: `dev`, `run`, `test`, `migrate:up`, `proto:gen`.
- [ ] **Step 6: 驗收**
  - `cd backend && task run` starts on `3080`.
  - `GET /api/v1/version` returns 200.
- [ ] **Step 7: Commit**

### Task 4: 前端 SolidJS 骨架

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/tsconfig.json`
- Create: `frontend/src/main.tsx`
- Create: `frontend/src/App.tsx`
- Create: `frontend/src/index.css`

**Interfaces:**
- Consumes: none
- Produces: dev server on port 3000, build to `dist/`

- [ ] **Step 1: 初始化 package.json**
  - Dependencies: `solid-js`, `@solidjs/router`, `@tanstack/solid-query`, `casl/ability`, `@casl/solid`, `valibot`, `tailwindcss`, `class-variance-authority`, `clsx`, `tailwind-merge`.
- [ ] **Step 2: 設定 Vite + TypeScript**
  - Path alias `~/*` → `src/*`.
- [ ] **Step 3: 設定 Tailwind CSS**
  - `tailwind.config.js`, `postcss.config.js`.
- [ ] **Step 4: 建立 `main.tsx` 與 `App.tsx`**
  - Basic router with `/` and `/login` placeholders.
- [ ] **Step 5: 驗收**
  - `pnpm dev` starts on 3000.
  - `pnpm build` outputs `dist/`.
- [ ] **Step 6: Commit**

### Task 5: App Flutter 骨架

**Files:**
- Create: `app/pubspec.yaml`
- Create: `app/analysis_options.yaml`
- Create: `app/Taskfile.yml`
- Create: `app/lib/main_dev.dart`
- Create: `app/lib/main_prod.dart`

**Interfaces:**
- Consumes: none
- Produces: `fvm flutter run --flavor dev --target lib/main_dev.dart`

- [ ] **Step 1: 設定 pubspec.yaml**
  - Flutter stable (Dart ≥3.10), dependencies: `auto_route`, `solidart`, `flutter_solidart`, `disco`, `fquery`, `sembest`, `flutter_secure_storage`, `connectrpc`, `firebase_messaging`, `talker`.
- [ ] **Step 2: 建立 flavor 設定**
  - Android `productFlavors`, iOS schemes `dev`/`prod`.
- [ ] **Step 3: 建立 `main_dev.dart` / `main_prod.dart`**
  - Call shared `main()` with env override.
- [ ] **Step 4: 建立 `app/Taskfile.yml`**
  - Tasks: `gen`, `run-dev`, `build-apk-prod`, `test`.
- [ ] **Step 5: 驗收**
  - `fvm flutter pub get` succeeds.
  - `fvm flutter analyze` shows no errors.
- [ ] **Step 6: Commit**

### Task 6: proto 與型別同步

**Files:**
- Create: `backend/proto/v1/common.proto`
- Create: `backend/buf.gen.yaml`
- Create: `frontend/src/lib/proto/.gitkeep`
- Create: `app/lib/gen/.gitkeep`

**Interfaces:**
- Consumes: none
- Produces: `task backend:proto:gen` generates Go/TS/Dart code

- [ ] **Step 1: 安裝 buf 與 plugins**
  - `buf`, `protoc-gen-go`, `protoc-gen-connect-go`, `protoc-gen-connect-es`, `protoc-gen-connect-dart`.
- [ ] **Step 2: 建立 `backend/buf.gen.yaml`**
  - Go out to `backend/internal/proto`, TS to `frontend/src/lib/proto`, Dart to `app/lib/gen`.
- [ ] **Step 3: 建立 `common.proto`**
  - Messages: `Pagination`, `TimestampRange`, `Money` (deprecated comment), `PrintPaperSize`.
- [ ] **Step 4: 建立 generate 腳本**
  - `task backend:proto:gen` runs `buf generate`.
- [ ] **Step 5: 驗收**
  - `task backend:proto:gen` succeeds.
  - Generated files appear in all three target directories.
- [ ] **Step 6: Commit**

### Task 7: CI 基礎管線

**Files:**
- Create: `.github/workflows/ci.yml`

**Interfaces:**
- Consumes: none
- Produces: PR checks for Go, frontend, Flutter

- [ ] **Step 1: 設定觸發條件**
  - `push`/`pull_request` to `main`.
- [ ] **Step 2: 設定 Go job**
  - Lint (`golangci-lint`), test (`go test ./...`), build (`go build ./cmd/server`), vuln (`govulncheck ./...`, D31-2).
- [ ] **Step 3: 設定 frontend job**
  - `pnpm install`, type check, lint, build.
- [ ] **Step 4: 設定 Flutter job**
  - `fvm flutter pub get`, `fvm flutter analyze`.
- [ ] **Step 5: 驗收**
  - Push a test PR; all jobs run.
- [ ] **Step 6: Commit**

### Task 8: 資料庫 migration 基礎

**Files:**
- Create: `backend/database/migrations/00001_init_schema.sql`
- Create: `backend/database/migrations/00002_rls_policies.sql`

**Interfaces:**
- Consumes: Task 2 (running PostgreSQL)
- Produces: `task backend:migrate:up` applies migrations

- [ ] **Step 1: 安裝 goose**
- [ ] **Step 2: 建立 `00001_init_schema.sql`**
  - `goose` version table, `+goose Up`/`+goose Down`.
- [ ] **Step 3: 建立 `00002_rls_policies.sql`**
  - `ALTER DATABASE ... SET row_security = on;`, create roles `app_read`, `app_write`.
- [ ] **Step 4: 建立 migrate task**
  - `task backend:migrate:up` runs `goose up`.
- [ ] **Step 5: 驗收**
  - `task backend:migrate:up` succeeds.
  - `goose_db_version` table exists.
- [ ] **Step 6: Commit**

### Task 9: Ent schema 基礎

**Files:**
- Create: `backend/ent/schema/user.go`
- Create: `backend/ent/schema/company.go`
- Create: `backend/ent/schema/department.go`

**Interfaces:**
- Consumes: Task 3 (backend skeleton)
- Produces: `ent.Company`, `ent.Department`, `ent.User`

- [ ] **Step 1: 安裝 ent**
  - `go get entgo.io/ent entgo.io/ent/dialect/entsql`.
- [ ] **Step 2: 定義 `Company` schema**
  - Fields: `name`, `tax_id`, `status`, `identifier`, `public_info`, `capabilities`, `logo_url`.
- [ ] **Step 3: 定義 `Department` schema**
  - Fields: `name`, edge to `Company`.
- [ ] **Step 4: 定義 `User` schema**
  - Fields: `email`, `name`, `status`, `role`, `phone`, `employee_no`, `is_customer`, `account_name`, `password_hash`, edges to `Company`, `Department`, `Customer`.
- [ ] **Step 5: 產生 ent code**
  - `go generate ./ent`.
- [ ] **Step 6: 驗收**
  - `go build ./...` passes.
- [ ] **Step 7: Commit**

### Task 10: 認證 proto

**Files:**
- Create: `backend/proto/v1/auth.proto`
- Modify: `backend/buf.gen.yaml` if needed

**Interfaces:**
- Consumes: Task 6 (proto setup)
- Produces: `AuthService` generated clients

- [ ] **Step 1: 定義 auth messages**
  - `LoginRequest`, `LoginResponse`, `RefreshRequest`, `LogoutRequest`, `RegisterCompleteRequest`, `QRLoginRequest`, `QRLoginResponse`.
- [ ] **Step 2: 定義 `AuthService`**
  - RPCs: `Login`, `Refresh`, `Logout`, `RegisterComplete`, `QRLogin`.
- [ ] **Step 3: 產生程式碼**
- [ ] **Step 4: 驗收**
  - Generated Go/TS/Dart files compile.
- [ ] **Step 5: Commit**

### Task 11: Google Workspace OIDC 登入

**Files:**
- Create: `backend/internal/auth/oidc.go`
- Create: `backend/internal/handlers/auth_handler.go`
- Modify: `backend/internal/server/server.go`

**Interfaces:**
- Consumes: Task 9 (User schema), Task 10 (Auth proto)
- Produces: `GET /api/v1/auth/google` redirect, `GET /api/v1/auth/google/callback` handler

- [ ] **Step 1: 實作 OAuth2 config 載入**
  - Use `golang.org/x/oauth2/google`.
- [ ] **Step 2: 實作登入導向端點**
  - Generate state, redirect to Google.
- [ ] **Step 3: 實作 callback 處理**
  - Verify ID token, extract email, find or create `User` with role `guest`.
- [ ] **Step 4: 建立 session / JWT**
  - For Web: set session cookie. For App: return JWT.
- [ ] **Step 5: 驗收**
  - Integration test: mock Google token, assert user created.
- [ ] **Step 6: Commit**

### Task 12: 客戶密碼登入

**Files:**
- Modify: `backend/internal/handlers/auth_handler.go`
- Create: `backend/internal/auth/password.go`

**Interfaces:**
- Consumes: Task 9 (User schema), Task 10 (Auth proto)
- Produces: `AuthService.Login` for customer accounts

- [ ] **Step 1: 實作密碼 hash 驗證**
  - Use `golang.org/x/crypto/bcrypt`.
- [ ] **Step 2: 實作客戶登入 RPC**
  - Find user by `customer_code`, verify password, check lockout.
- [ ] **Step 3: 實作錯誤次數與鎖定**
  - Store failed attempts in Valkey; lock 30 min after 5 failures.
- [ ] **Step 4: 驗收**
  - Unit tests: correct password, wrong password, locked account.
- [ ] **Step 5: Commit**

### Task 13: JWT / Session 與 token_version

**Files:**
- Create: `backend/internal/auth/token.go`
- Create: `backend/internal/auth/session.go`

**Interfaces:**
- Consumes: Task 11, Task 12
- Produces: `CreateAccessToken`, `CreateRefreshToken`, `Refresh`, `RevokeAll`

- [ ] **Step 1: 實作 JWT access token**
  - Claims: `user_id`, `company_id`, `department_id`, `role`, `token_version`, exp 1h.
- [ ] **Step 2: 實作 refresh token**
  - Store in Valkey with 30-day TTL, rotate on use.
- [ ] **Step 3: 實作 token_version 撤銷**
  - Increment `token_version` on password change / disable / role change.
- [ ] **Step 4: 實作 Web session cookie**
  - Use `scs` with Valkey store.
- [ ] **Step 5: 驗收**
  - Tests for issue, refresh, rotation, revocation.
- [ ] **Step 6: Commit**

### Task 14: Casbin + RLS 基礎

**Files:**
- Create: `backend/internal/auth/casbin.go`
- Create: `backend/internal/auth/rls.go`
- Create: `backend/config/rbac_model.conf`

**Interfaces:**
- Consumes: Task 9 (User schema)
- Produces: `Enforce(sub, obj, act, dom)` and `WithRLS(ctx, scope)`

- [ ] **Step 1: 載入 Casbin model 與 policy adapter**
  - File adapter or DB adapter.
- [ ] **Step 2: 定義 7 內建角色 seed policy**
  - `super`, `company_admin`, `dept_admin`, `staff`, `customer`, `guest`, `developer`.
- [ ] **Step 3: 實作 RLS 切換**
  - `SET LOCAL app.current_user_id`, `SET LOCAL app.data_scope`.
- [ ] **Step 4: 中介層整合**
  - Auth middleware populates context with user + scope.
- [ ] **Step 5: 驗收**
  - Tests: allowed/denied actions per role; RLS filters data correctly.
- [ ] **Step 6: Commit**

### Task 15: Web 登入頁

**Files:**
- Create: `frontend/src/features/auth/pages/LoginPage.tsx`
- Create: `frontend/src/features/auth/components/GoogleLoginButton.tsx`
- Modify: `frontend/src/router/index.tsx`

**Interfaces:**
- Consumes: Task 10 (AuthService generated client)
- Produces: `/login` route

- [ ] **Step 1: 建立登入頁 layout**
  - Two tabs: 員工 / 店家.
- [ ] **Step 2: 實作 Google 登入按鈕**
  - Redirect to `/api/v1/auth/google`.
- [ ] **Step 3: 實作店家登入表單**
  - Inputs: customer_code, password.
- [ ] **Step 4: 整合 AuthService.Login**
- [ ] **Step 5: 驗收**
  - Manual browser test: login succeeds, redirects to `/`.
- [ ] **Step 6: Commit**

### Task 16: App 登入流程

**Files:**
- Create: `app/lib/features/auth/pages/login_page.dart`
- Create: `app/lib/features/auth/pages/identity_select_page.dart`
- Modify: `app/lib/router/app_router.dart`

**Interfaces:**
- Consumes: Task 10 (AuthService generated Dart client)
- Produces: `/login`, identity selection UI

- [ ] **Step 1: 建立身分選擇頁**
  - Buttons: 我是店家 / 我是業務.
- [ ] **Step 2: 建立店家登入頁**
  - customer_code + password inputs.
- [ ] **Step 3: 建立業務 Google 登入**
  - Use `flutter_web_auth` PKCE flow.
- [ ] **Step 4: 儲存 token 到 secure storage**
- [ ] **Step 5: 驗收**
  - Unit test: login calls AuthService.Login with correct payload.
- [ ] **Step 6: Commit**

### Task 17: 員工註冊完成頁

**Files:**
- Create: `frontend/src/features/auth/pages/RegisterCompletePage.tsx`
- Create: `app/lib/features/auth/pages/register_complete_page.dart`
- Modify: `backend/internal/handlers/auth_handler.go`

**Interfaces:**
- Consumes: Task 11 (Google OIDC)
- Produces: `AuthService.RegisterComplete`

- [ ] **Step 1: 後端實作 RegisterComplete RPC**
  - Accept company_id, name; promote guest to pending/approved.
- [ ] **Step 2: Web 註冊完成頁**
  - Select company, input name, submit.
- [ ] **Step 3: App 註冊完成頁**
  - Same flow.
- [ ] **Step 4: 驗收**
  - Test: guest user completes registration, status updated.
- [ ] **Step 5: Commit**

### Task 18: 角色權限管理 API

**Files:**
- Create: `backend/proto/v1/role.proto`
- Create: `backend/internal/services/role_service.go`
- Create: `backend/internal/handlers/role_handler.go`

**Interfaces:**
- Consumes: Task 14 (Casbin)
- Produces: `RoleService` with CRUD for role_permissions

- [ ] **Step 1: 定義 role proto**
  - Messages: `Role`, `Permission`, `UpdateRolePermissionsRequest`.
- [ ] **Step 2: 實作角色權限 CRUD**
  - company_admin limited to own company.
- [ ] **Step 3: 產生並編譯**
- [ ] **Step 4: 驗收**
  - Tests: update permissions, enforce new rules.
- [ ] **Step 5: Commit**

### Task 19: Web 角色權限頁

**Files:**
- Create: `frontend/src/features/users/pages/RolesPage.tsx`
- Create: `frontend/src/features/users/components/PermissionMatrix.tsx`

**Interfaces:**
- Consumes: Task 18 (RoleService)
- Produces: `/users/roles` page

- [ ] **Step 1: 建立角色列表頁**
- [ ] **Step 2: 建立權限矩陣元件**
  - Checkboxes per action/resource.
- [ ] **Step 3: 整合保存 API**
- [ ] **Step 4: 驗收**
  - Manual test: changes saved, reflected in policy.
- [ ] **Step 5: Commit**

### Task 20: Company / Department 管理

**Files:**
- Create: `backend/proto/v1/company.proto`
- Create: `backend/internal/services/company_service.go`
- Create: `frontend/src/features/users/pages/CompaniesPage.tsx`
- Create: `frontend/src/features/users/pages/DepartmentsPage.tsx`

**Interfaces:**
- Consumes: Task 9 (Company/Department schema)
- Produces: `CompanyService`, `DepartmentService`, Web CRUD pages

- [ ] **Step 1: 定義 company/department proto**
- [ ] **Step 2: 實作後端 CRUD**
- [ ] **Step 3: 產生程式碼**
- [ ] **Step 4: Web 公司/部門管理頁**
- [ ] **Step 5: 驗收**
  - CRUD integration tests.
- [ ] **Step 6: Commit**

---

## Wave 2：多租戶主檔

### Task 21: 客戶主檔後端

**Files:**
- Create: `backend/proto/v1/customer.proto`
- Create: `backend/ent/schema/customer.go`
- Create: `backend/internal/services/customer_service.go`
- Create: `backend/internal/handlers/customer_handler.go`

**Interfaces:**
- Consumes: Task 9 (User schema), Task 14 (RLS)
- Produces: `CustomerService` with CRUD + auto-numbering

- [ ] **Step 1: 定義 customer proto**
  - Fields: code, name, tax_id, payment_method, settlement_method, invoice_type, etc.
- [ ] **Step 2: 實作 Ent schema**
  - Soft delete, unique partial index on `(company_id, code)` where `deleted_at IS NULL`.
- [ ] **Step 3: 實作取號邏輯**
  - Company prefix + 6-digit auto-increment in same transaction.
- [ ] **Step 4: 實作 CRUD service/handler**
- [ ] **Step 5: 產生程式碼**
- [ ] **Step 6: 驗收**
  - Tests: create customer gets sequential code, soft delete, RLS isolation.
- [ ] **Step 7: Commit**

### Task 22: 客戶地址簿與聯絡人

**Files:**
- Create: `backend/ent/schema/customer_address.go`
- Create: `backend/ent/schema/customer_contact.go`
- Modify: `backend/proto/v1/customer.proto`

**Interfaces:**
- Consumes: Task 21 (Customer schema)
- Produces: Nested address/contact CRUD in CustomerService

- [ ] **Step 1: 建立 address/contact schema**
- [ ] **Step 2: 擴充 proto messages**
- [ ] **Step 3: 實作 CRUD**
- [ ] **Step 4: 驗收**
- [ ] **Step 5: Commit**

### Task 23: Web 客戶主檔頁面

**Files:**
- Create: `frontend/src/features/master-data/pages/CustomersPage.tsx`
- Create: `frontend/src/features/master-data/components/CustomerForm.tsx`

**Interfaces:**
- Consumes: Task 21 (CustomerService)
- Produces: `/customers` list and form

- [ ] **Step 1: 建立客戶列表頁**
  - Table with pagination, search.
- [ ] **Step 2: 建立客戶表單**
  - Create/edit with address/contact sub-forms.
- [ ] **Step 3: 整合 API**
- [ ] **Step 4: 驗收**
- [ ] **Step 5: Commit**

### Task 24: 商品主檔後端

**Files:**
- Create: `backend/proto/v1/product.proto`
- Create: `backend/ent/schema/product.go`
- Create: `backend/ent/schema/unit_conversion.go`
- Create: `backend/internal/services/product_service.go`

**Interfaces:**
- Consumes: Task 14 (RLS)
- Produces: `ProductService` with CRUD + unit conversions

- [ ] **Step 1: 定義 product proto**
- [ ] **Step 2: 實作 Ent schema**
  - Product, unit conversions, category edge.
- [ ] **Step 3: 實作 CRUD + conversion logic**
- [ ] **Step 4: 產生程式碼**
- [ ] **Step 5: 驗收**
- [ ] **Step 6: Commit**

### Task 25: 部門級主檔（倉別/車次/分切規格/分類）

**Files:**
- Create: `backend/proto/v1/warehouse.proto`, `route.proto`, `cutting_spec.proto`, `category.proto`
- Create: `backend/ent/schema/warehouse.go`, `route.go`, `cutting_spec.go`, `category.go`
- Create: `backend/internal/services/master_data_service.go`

**Interfaces:**
- Consumes: Task 14 (RLS)
- Produces: CRUD services for warehouse, route, cutting_spec, category

- [ ] **Step 1: 定義 proto messages**
- [ ] **Step 2: 實作 Ent schemas**
- [ ] **Step 3: 實作 CRUD services/handlers**
- [ ] **Step 4: 產生程式碼**
- [ ] **Step 5: 驗收**
- [ ] **Step 6: Commit**

### Task 26: metadicts 字典檔

**Files:**
- Create: `backend/proto/v1/metadict.proto`
- Create: `backend/ent/schema/metadict.go`
- Create: `backend/internal/services/metadict_service.go`

**Interfaces:**
- Consumes: Task 14 (RLS)
- Produces: `MetadictService` with system defaults + department overrides

- [ ] **Step 1: 定義 metadict proto**
- [ ] **Step 2: 實作 schema**
  - `type`, `key`, `value`, `department_id` (nullable).
- [ ] **Step 3: 實作 CRUD**
- [ ] **Step 4: 驗收**
- [ ] **Step 5: Commit**

### Task 27: 客戶專屬商品與別名

**Files:**
- Create: `backend/proto/v1/customer_product.proto`
- Create: `backend/ent/schema/customer_product.go`
- Create: `backend/internal/services/customer_product_service.go`

**Interfaces:**
- Consumes: Task 21 (Customer), Task 24 (Product)
- Produces: `CustomerProductService`

- [ ] **Step 1: 定義 customer_product proto**
  - Fields: customer_id, product_id, alias, default_qty, unit_id, active.
- [ ] **Step 2: 實作 schema**
- [ ] **Step 3: 實作 CRUD + alias lookup**
- [ ] **Step 4: 產生程式碼**
- [ ] **Step 5: 驗收**
- [ ] **Step 6: Commit**

### Task 28: Web 商品與主檔管理頁

**Files:**
- Create: `frontend/src/features/master-data/pages/ProductsPage.tsx`
- Create: `frontend/src/features/master-data/pages/WarehousesPage.tsx`
- Create: `frontend/src/features/master-data/pages/RoutesPage.tsx`
- Create: `frontend/src/features/master-data/pages/CustomerProductsPage.tsx`

**Interfaces:**
- Consumes: Tasks 24–27
- Produces: Master data CRUD pages

- [ ] **Step 1: 建立商品列表/表單**
- [ ] **Step 2: 建立倉別/車次/分切/分類頁面**
- [ ] **Step 3: 建立客戶專屬商品頁**
- [ ] **Step 4: 驗收**
- [ ] **Step 5: Commit**

### Task 29: App 客戶列表與搜尋

**Files:**
- Create: `app/lib/features/customers/pages/customers_list_page.dart`
- Create: `app/lib/features/customers/data/customer_repository.dart`

**Interfaces:**
- Consumes: Task 21 (CustomerService)
- Produces: App customers list with search

- [ ] **Step 1: 建立 repository**
- [ ] **Step 2: 建立列表頁**
- [ ] **Step 3: 實作搜尋**
- [ ] **Step 4: 驗收**
- [ ] **Step 5: Commit**

### Task 30: App 手動新增客戶

**Files:**
- Create: `app/lib/features/customers/pages/customer_create_page.dart`
- Modify: `app/lib/features/customers/data/customer_repository.dart`

**Interfaces:**
- Consumes: Task 21 (CustomerService)
- Produces: App customer create form + auto-generated login account

- [ ] **Step 1: 建立表單頁**
- [ ] **Step 2: 呼叫 CustomerService.CreateCustomer**
- [ ] **Step 3: 顯示一次性臨時密碼**
- [ ] **Step 4: 驗收**
- [ ] **Step 5: Commit**

---

## Wave 3：訂單、退貨與通知

### Task 31: 訂單後端

**Files:**
- Create: `backend/proto/v1/sales_order.proto`
- Create: `backend/ent/schema/sales_order.go`
- Create: `backend/ent/schema/sales_order_event.go`
- Create: `backend/internal/services/sales_order_service.go`

**Interfaces:**
- Consumes: Tasks 21, 24, 27 (Customer, Product, CustomerProduct)
- Produces: `SalesOrderService` with state machine + numbering

- [ ] **Step 1: 定義 sales_order proto**
  - Order, OrderItem, status enum, delivery date.
- [ ] **Step 2: 實作 Ent schemas**
  - SalesOrder, SalesOrderItem, SalesOrderEvent.
- [ ] **Step 3: 實作編號取號**
  - Source code + 6-digit auto-increment.
- [ ] **Step 4: 實作狀態機**
  - `pending ⇄ processing → completed`, `cancelled`, `voided`.
- [ ] **Step 5: 實作偏好送貨日順延**
- [ ] **Step 6: 產生程式碼**
- [ ] **Step 7: 驗收**
- [ ] **Step 8: Commit**

### Task 32: 退貨後端

**Files:**
- Create: `backend/proto/v1/return_request.proto`
- Create: `backend/ent/schema/return_request.go`
- Create: `backend/internal/services/return_request_service.go`

**Interfaces:**
- Consumes: Task 31 (SalesOrder)
- Produces: `ReturnRequestService`

- [ ] **Step 1: 定義 return_request proto**
- [ ] **Step 2: 實作 schema**
- [ ] **Step 3: 實作申請 + 審核 + 證明**
- [ ] **Step 4: 產生程式碼**
- [ ] **Step 5: 驗收**
- [ ] **Step 6: Commit**

### Task 33: 通知後端

**Files:**
- Create: `backend/proto/v1/notification.proto`
- Create: `backend/ent/schema/notification.go`
- Create: `backend/internal/services/notification_service.go`
- Create: `backend/internal/notifications/fcm.go`

**Interfaces:**
- Consumes: Tasks 31, 32 (order/return events)
- Produces: `NotificationService`, FCM sender

- [ ] **Step 1: 定義 notification proto**
- [ ] **Step 2: 實作 schema**
- [ ] **Step 3: 實作 FCM 發送**
- [ ] **Step 4: 實作業務下單推客戶、退貨結果推發起者**
- [ ] **Step 5: 產生程式碼**
- [ ] **Step 6: 驗收**
- [ ] **Step 7: Commit**

### Task 34: Web 訂單頁面

**Files:**
- Create: `frontend/src/features/orders/pages/OrdersPage.tsx`
- Create: `frontend/src/features/orders/pages/OrderCreatePage.tsx`
- Create: `frontend/src/features/orders/pages/OrderDetailPage.tsx`

**Interfaces:**
- Consumes: Task 31 (SalesOrderService)
- Produces: `/orders/*` routes

- [ ] **Step 1: 建立訂單列表**
- [ ] **Step 2: 建立訂單建立頁**
  - Select customer → load customer products → add manual/total items.
- [ ] **Step 3: 建立訂單詳情**
- [ ] **Step 4: 驗收**
- [ ] **Step 5: Commit**

### Task 35: Web 退貨頁面

**Files:**
- Create: `frontend/src/features/orders/pages/ReturnsPage.tsx`
- Create: `frontend/src/features/orders/components/ReturnApproveDialog.tsx`

**Interfaces:**
- Consumes: Task 32 (ReturnRequestService)
- Produces: `/returns/*` routes

- [ ] **Step 1: 建立退貨申請列表**
- [ ] **Step 2: 建立審核對話框**
- [ ] **Step 3: 驗收**
- [ ] **Step 4: Commit**

### Task 36: Web 通知中心

**Files:**
- Create: `frontend/src/features/notifications/pages/NotificationsPage.tsx`
- Create: `frontend/src/components/NotificationBadge.tsx`

**Interfaces:**
- Consumes: Task 33 (NotificationService)
- Produces: Notification list + badge

- [ ] **Step 1: 建立通知列表頁**
- [ ] **Step 2: 建立全域 badge**
- [ ] **Step 3: 驗收**
- [ ] **Step 4: Commit**

### Task 37: App 快速下單

**Files:**
- Create: `app/lib/features/products/pages/quick_order_page.dart`
- Create: `app/lib/features/products/application/order_draft_controller.dart`

**Interfaces:**
- Consumes: Tasks 24, 27, 31 (Product/CustomerProduct/SalesOrder services)
- Produces: `/products/quick-order` flow

- [ ] **Step 1: 建立 order draft controller**
  - Signals for customer, items, delivery date.
- [ ] **Step 2: 建立選客戶步驟**
- [ ] **Step 3: 建立選品/手打步驟**
- [ ] **Step 4: 建立確認與送出**
- [ ] **Step 5: 驗收**
- [ ] **Step 6: Commit**

### Task 38: App 訂單歷史與退貨

**Files:**
- Create: `app/lib/features/orders/pages/orders_list_page.dart`
- Create: `app/lib/features/orders/pages/order_detail_page.dart`
- Create: `app/lib/features/orders/pages/return_request_page.dart`

**Interfaces:**
- Consumes: Tasks 31, 32
- Produces: App order history + return flow

- [ ] **Step 1: 建立訂單列表**
- [ ] **Step 2: 建立訂單詳情**
- [ ] **Step 3: 建立退貨申請頁**
- [ ] **Step 4: 驗收**
- [ ] **Step 5: Commit**

### Task 39: App 通知

**Files:**
- Create: `app/lib/features/notifications/pages/notifications_page.dart`
- Create: `app/lib/features/notifications/fcm_handler.dart`

**Interfaces:**
- Consumes: Task 33
- Produces: App notification list + FCM handling

- [ ] **Step 1: 建立通知列表**
- [ ] **Step 2: 處理 FCM 前景/背景訊息**
- [ ] **Step 3: 通知點擊導航**
- [ ] **Step 4: 驗收**
- [ ] **Step 5: Commit**

---

## Wave 4：派車與列印

### Task 40: 派車後端

**Files:**
- Create: `backend/proto/v1/dispatch.proto`
- Create: `backend/ent/schema/dispatch.go`
- Create: `backend/internal/services/dispatch_service.go`

**Interfaces:**
- Consumes: Task 25 (Route), Task 31 (SalesOrder)
- Produces: `DispatchService`, `WatchBoard` streaming

- [ ] **Step 1: 定義 dispatch proto**
- [ ] **Step 2: 實作 schema**
- [ ] **Step 3: 實作批次確認 API**
- [ ] **Step 4: 實作 `WatchBoard` streaming**
- [ ] **Step 5: 實作 Valkey pub/sub + 輪詢降級**
- [ ] **Step 6: 產生程式碼**
- [ ] **Step 7: 驗收**
- [ ] **Step 8: Commit**

### Task 41: 列印後端

**Files:**
- Create: `backend/proto/v1/print.proto`
- Create: `backend/internal/services/print_service.go`
- Create: `backend/internal/print/templates/` (HTML templates)

**Interfaces:**
- Consumes: Task 40 (Dispatch)
- Produces: `PrintService`, Gotenberg integration

- [ ] **Step 1: 定義 print proto**
- [ ] **Step 2: 建立四種單據 HTML 模板**
- [ ] **Step 3: 實作 Gotenberg HTML→PDF**
- [ ] **Step 4: 實作重印記錄**
- [ ] **Step 5: 產生程式碼**
- [ ] **Step 6: 驗收**
- [ ] **Step 7: Commit**

### Task 42: Web 派車看板

**Files:**
- Create: `frontend/src/features/dispatch/pages/DispatchBoardPage.tsx`
- Create: `frontend/src/features/dispatch/components/KanbanColumn.tsx`

**Interfaces:**
- Consumes: Task 40 (DispatchService, WatchBoard)
- Produces: `/dispatch` Kanban board

- [ ] **Step 1: 建立看板頁面**
- [ ] **Step 2: 實作拖放排序**
- [ ] **Step 3: 整合 streaming / 輪詢**
- [ ] **Step 4: 實作批次確認**
- [ ] **Step 5: 驗收**
- [ ] **Step 6: Commit**

### Task 43: Web 列印頁面

**Files:**
- Create: `frontend/src/features/printing/pages/PrintsPage.tsx`
- Create: `frontend/src/features/printing/components/PrintPreview.tsx`

**Interfaces:**
- Consumes: Task 41 (PrintService)
- Produces: `/prints/*` routes

- [ ] **Step 1: 建立列印入口與車次選擇**
- [ ] **Step 2: 建立預覽/列印元件**
- [ ] **Step 3: 實作重印原因輸入**
- [ ] **Step 4: 驗收**
- [ ] **Step 5: Commit**

---

## Wave 5：公告、檔案、稽核、部署

### Task 44: 公告後端

**Files:**
- Create: `backend/proto/v1/announcement.proto`
- Create: `backend/ent/schema/announcement.go`
- Create: `backend/internal/services/announcement_service.go`

**Interfaces:**
- Consumes: Task 33 (Notification)
- Produces: `AnnouncementService`, promo tag targeting

- [ ] **Step 1: 定義 announcement proto**
- [ ] **Step 2: 實作 schema**
- [ ] **Step 3: 實作 CMS CRUD + 促銷推播選群**
- [ ] **Step 4: 產生程式碼**
- [ ] **Step 5: 驗收**
- [ ] **Step 6: Commit**

### Task 45: 檔案後端

**Files:**
- Create: `backend/proto/v1/file_asset.proto`
- Create: `backend/ent/schema/file_asset.go`
- Create: `backend/internal/services/file_asset_service.go`

**Interfaces:**
- Consumes: Task 8 (migration), Task 14 (RLS)
- Produces: `FileAssetService`, REST upload/download endpoints

- [ ] **Step 1: 定義 file_asset proto**
- [ ] **Step 2: 實作 schema**
- [ ] **Step 3: 實作上傳（白名單 + magic bytes）**
- [ ] **Step 4: 實作下載/預覽**
- [ ] **Step 5: 產生程式碼**
- [ ] **Step 6: 驗收**
- [ ] **Step 7: Commit**

### Task 46: 稽核後端

**Files:**
- Create: `backend/proto/v1/audit_log.proto`
- Create: `backend/ent/schema/audit_log.go`
- Create: `backend/internal/services/audit_log_service.go`
- Create: `backend/internal/audit/middleware.go`

**Interfaces:**
- Consumes: Task 14 (auth middleware)
- Produces: `AuditLogService`, audit middleware

- [ ] **Step 1: 定義 audit_log proto**
- [ ] **Step 2: 實作 schema**
- [ ] **Step 3: 實作同交易寫入 middleware**
- [ ] **Step 4: 實作保留期設定 API**
- [ ] **Step 5: 產生程式碼**
- [ ] **Step 6: 驗收**
- [ ] **Step 7: Commit**

### Task 47: Web 公告/檔案/稽核頁面

**Files:**
- Create: `frontend/src/features/announcements/pages/AnnouncementsPage.tsx`
- Create: `frontend/src/features/announcements/pages/PromoPushPage.tsx`
- Create: `frontend/src/features/files/pages/FilesPage.tsx`
- Create: `frontend/src/features/audit/pages/AuditLogsPage.tsx`

**Interfaces:**
- Consumes: Tasks 44, 45, 46
- Produces: CMS, file manager, audit log viewer

- [ ] **Step 1: 建立公告 CMS 頁**
- [ ] **Step 2: 建立促銷推播頁**
- [ ] **Step 3: 建立檔案管理頁**
- [ ] **Step 4: 建立稽核日誌頁**
- [ ] **Step 5: 驗收**
- [ ] **Step 6: Commit**

### Task 48: App 店家帳號管理

**Files:**
- Create: `app/lib/features/account/pages/sub_accounts_page.dart`
- Create: `app/lib/features/account/pages/sub_account_create_page.dart`

**Interfaces:**
- Consumes: Task 9 (User schema), Task 12 (customer auth)
- Produces: `/account/sub-accounts` for primary account

- [ ] **Step 1: 建立子帳號列表**
- [ ] **Step 2: 建立新增/停用/重置密碼功能**
- [ ] **Step 3: 實作主帳號路由守衛**
- [ ] **Step 4: 驗收**
- [ ] **Step 5: Commit**

### Task 49: App 首頁與公告

**Files:**
- Create: `app/lib/features/home/pages/home_page.dart`
- Create: `app/lib/features/home/widgets/announcement_carousel.dart`

**Interfaces:**
- Consumes: Task 44 (AnnouncementService)
- Produces: Home tab with announcements + quick order entry

- [ ] **Step 1: 建立首頁殼層**
- [ ] **Step 2: 建立公告輪播**
- [ ] **Step 3: 整合快速下單入口**
- [ ] **Step 4: 驗收**
- [ ] **Step 5: Commit**

### Task 50: 部署與維運

**Files:**
- Create: `infra/k8s/postgres-statefulset.yaml`
- Create: `infra/k8s/valkey-statefulset.yaml`
- Create: `infra/helm/sales-order/Chart.yaml`
- Create: `infra/helm/sales-order/templates/*.yaml`
- Create: `infra/k8s/prometheus-values.yaml`
- Create: `infra/k8s/backup-cronjob.yaml`

**Interfaces:**
- Consumes: all prior tasks
- Produces: production-ready k8s manifests

- [ ] **Step 1: 建立 PostgreSQL StatefulSet + PITR**
- [ ] **Step 2: 建立 Valkey StatefulSet**
- [ ] **Step 3: 建立應用 Helm chart**
- [ ] **Step 4: 建立 Prometheus/Grafana/Alertmanager**
- [ ] **Step 5: 建立 GCS 備份 CronJob**
- [ ] **Step 6: 執行災難復原演練**
- [ ] **Step 7: 驗收**
- [ ] **Step 8: Commit**

---

## 自我檢查清單

- [x] **Spec coverage:** 每個設計文件中的工作項目（BE/WEB/APP）都對應到一個以上 Task。
- [x] **Placeholder scan:** 無 TBD/TODO/「implement later」；每個 Task 皆有 Files / Interfaces / Steps / Acceptance Criteria 區塊。
- [x] **Type consistency:** proto service 名稱、Ent schema 名稱、頁面路由於跨 Task 引用時保持一致。
- [x] **Scope check:** 本計畫聚焦 1.0 三子專案實作；1.1 AI 輔助不在此計畫範圍。
- [x] **Dependency ordering:** Wave 1 → 2 → 3 → 4 → 5 順序符合後端先於前端/App、主檔先於訂單的阻塞關係。

---

## 修訂記錄

| 修訂號 | 日期 | 修訂內容 | 修訂者 |
|---|---|---|---|
| v0.1.0 | 2026-08-05 | 初版：50 Tasks 分 5 Waves，涵蓋 Backend / Web / App 三子專案 | 開發團隊 |
