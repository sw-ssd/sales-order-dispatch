# 訂出貨系統 — AI Agent 專案導覽

> 本文件位於專案根目錄，統整 `sales-order-backend`、`sales-order-frontend`、`sales-order-app` 三個子專案。各子專案另有更詳細的 `AGENTS.md`，修改前請一併參閱。
>
> **1.0 重建計畫**：本文件描述現行三倉系統；「多公司訂出貨系統 1.0」（全新 monorepo 重建、規劃定稿、尚未實作）的規劃文件整合於 `docs/PLANNING_OVERVIEW.md`，決策層為 `docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md`，規劃細節請先讀 `docs/PLANNING_OVERVIEW.md`。
> - `sales-order-app/AGENTS.md`：Flutter App 詳細指引
> - `sales-order-backend/AGENTS.md`：Go 後台詳細指引
> - `sales-order-frontend/AGENTS.md`：SolidJS 前端詳細指引

---

## 1. 專案概覽

這是**訂出貨系統**（Sales Order System）的程式碼庫，主要用於業務員（salesrep）與客戶（customer）建立、查詢、管理銷售訂單（sales order），並與 NetSuite 進行資料同步。系統由三部分組成：

| 子專案 | 技術 | 用途 | 版本 |
|--------|------|------|------|
| `sales-order-backend` | Go 1.25 | RESTful API 後台服務 | 見 `go.mod` |
| `sales-order-frontend` | SolidJS 1.9 + TypeScript 5.9 + Vite 6 | 網頁中台（SPA） | `package.json` `1.0.29` |
| `sales-order-app` | Flutter 3.35.2 + Dart ≥3.9.0 | 跨平台行動 App | `pubspec.yaml` `1.2.6+25` |
| `appimg` | — | 商店截圖、App 宣傳圖、ButterKit 專案 | — |

- **主要語言**：文件與程式碼註解以**繁體中文**為主；識別字、檔案名稱、套件名稱維持英文原文。
- **API 前綴**：後端統一使用 `/api/v1`。
- **存放庫結構**：根目錄本身沒有套件設定檔（無 root `package.json`、`go.mod`、`pubspec.yaml`），三個子專案各自獨立為一個 Git 倉庫（各有 `.git`）。

---

## 2. 專案結構與關鍵設定檔

```text
.
├── appimg/                    # 共用宣傳與商店截圖素材
│   ├── android/               # Android 商店圖
│   ├── ios/                   # iOS 商店圖
│   ├── screenshots.butterkit/ # ButterKit 後製專案
│   ├── export_app_store/      # iOS 截圖成品
│   ├── export_google_play/    # Android 截圖成品
│   └── 廣宣頁面.png
├── sales-order-backend/       # Go 後台 API
│   ├── cmd/                   # 可執行入口（sw8、migrate、seed、route、token、auto_increment）
│   ├── config/                # 環境變數設定
│   ├── database/              # Goose 遷移檔、seeder、Atlas diff 工具
│   ├── ent/                   # Ent schema 與產生碼
│   ├── internal/              # 應用核心（domain、middleware、server、utility）
│   ├── third_party/           # 外部整合（NetSuite、Casbin、Postgres session store、Redis）
│   ├── testcontainers/        # dockertest 容器測試輔助
│   ├── taskfiles/             # Task 子任務拆分
│   ├── go.mod / go.sum
│   ├── Taskfile.yml
│   ├── Dockerfile
│   ├── docker-compose-*.yml
│   ├── .air.toml              # air 熱重載設定
│   ├── tygo.yaml              # 產生前端 TypeScript 型別
│   └── hexagon.env            # 範例/開發環境變數檔（含敏感資訊，已在版控中）
├── sales-order-frontend/      # SolidJS 網頁前端
│   ├── src/
│   │   ├── components/        # UI 元件（ui/form/datatable/sidebar…）
│   │   ├── pages/             # 頁面層級元件
│   │   ├── routes/            # TanStack 檔案式路由
│   │   ├── lib/               # 領域 API + TanStack Query
│   │   ├── models/            # TypeScript 型別
│   │   ├── constant/          # 常數、CASL 規則、選單
│   │   ├── env/               # 環境變數綱目（Valibot）
│   │   ├── main.tsx
│   │   └── globals.css / index.css
│   ├── public/
│   ├── test_sqls/             # 臨時 SQL 草稿（不屬建置）
│   ├── package.json
│   ├── pnpm-lock.yaml
│   ├── pnpm-workspace.yaml
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   ├── firebase.json
│   └── Taskfile.yml
└── sales-order-app/           # Flutter 行動 App
    ├── lib/
    │   ├── layer_business/    # 業務邏輯、網路、路由、服務、DI
    │   ├── layer_data/        # 模型、本地儲存、常數
    │   └── layer_presentation/# 畫面、widget、主題
    ├── android/ / ios/         # 平台專案
    ├── integration_test/       # Maestro 整合測試 flow
    ├── assets/                 # 圖片、字型、顏色 XML
    ├── fastlane_bin/ / bin/    # bundler binstub
    ├── pubspec.yaml
    ├── pubspec.lock
    ├── analysis_options.yaml
    ├── Taskfile.yml
    ├── .fvmrc                  # Flutter 版本 3.35.2
    ├── firebase.json
    └── firebase_flavor.sh
```

---

## 3. 常用建置與執行指令

### 3.1 sales-order-backend（Go）

需安裝 [Task](https://taskfile.dev/) 與 Go 1.25；開發時建議一併安裝 `air`、Docker/Podman。

```bash
cd sales-order-backend

# 啟動本地基礎設施（Postgres + Valkey + Mailpit）
task infra:start

# 熱重載開發（air，監聽 0.0.0.0:3080）
task dev

# 直接執行
task run

# 資料庫遷移
task goose:up              # 套用 migration
task db:migrate -- <name>  # 產生新的 Goose 遷移檔

# 程式碼產生
task ent:gen               # Ent
task swagger               # Swagger 文件
task oapigen               # NetSuite OpenAPI client
task typego                # 前端 TypeScript 型別

# 測試
task test:unit             # go test -short ./...
task test:integration      # go test -run Integration ./...
task test                  # 全部測試

# 品質檢查
task fmt                   # go fmt ./...
task vet                   # go vet ./...
task lint                  # golangci-lint run
task vuln                  # govulncheck ./...

# 部署
task gcp:deploy            # Cloud Run 一鍵部署
```

### 3.2 sales-order-frontend（SolidJS）

需安裝 pnpm。

```bash
cd sales-order-frontend
pnpm install

# 開發伺服器（預設 http://localhost:3000）
pnpm run dev
# 或
task dev

# 正式建置（輸出至 dist/）
pnpm run build
# 或
task build

# 預覽
pnpm run serve

# 測試
pnpm run test              # vitest run

# 部署至 Firebase Hosting
task deploy                # patch 版本、建置、登入、部署 hosting

# 新增 solid-ui 元件
task ui:add -- <component-name>
```

### 3.3 sales-order-app（Flutter）

建議透過 FVM（Flutter Version Management）使用固定版本 `3.35.2`；VS Code 啟動設定位於 `.vscode/launch.json`。

```bash
cd sales-order-app
fvm flutter pub get

# 開發執行
fvm flutter run --flavor dev --target lib/main_dev.dart
fvm flutter run --flavor prod --target lib/main_prod.dart

# 產生程式碼（Freezed、json_serializable、reactive_forms、auto_route、envied、flutter_gen、dart_mappable）
fvm dart run build_runner build --delete-conflicting-outputs

# 建置
fvm flutter build apk --flavor prod --target lib/main_prod.dart
fvm flutter build appbundle --flavor prod --target lib/main_prod.dart
fvm flutter build ios --flavor prod --target lib/main_prod.dart   # 需 macOS + Xcode

# 測試
task test:maestro          # Maestro 整合測試
fvm flutter test           # 若新增單元/Widget 測試

# 截圖（需後端在線）
task screenshots:android
task screenshots:ios
```

> 注意：`sales-order-app/Taskfile.yml` 內部分任務直接呼叫 `flutter` / `dart`，若系統 PATH 未掛載 Flutter，請改用 `fvm flutter ...` / `fvm dart ...` 手動執行對應指令。

---

## 4. 程式碼組織與主要慣例

### 4.1 sales-order-backend

採用網域分層（DDD 風格），每個 `internal/domain/<domain>` 通常包含：

- `model.go`：request/response/filter DTO
- `handler.go`：HTTP handler
- `repository.go`：Ent 資料存取
- `usecase.go`：業務邏輯
- `register.go`：路由註冊
- `transformation.go`：DTO 與 ent 模型轉換
- `*_test.go` / `*_mock.go`：測試與 mock

網域名稱多為複數英文，例如 `sales_orders`、`customers`、`departments`。`cmd/sw8/main.go` 是主入口，`internal/server/initDomains.go` 註冊所有 domain 路由。

### 4.2 sales-order-frontend

- **路由**：TanStack Solid Router 檔案式路由，定義於 `src/routes/`，自動編譯為 `src/routeTree.gen.ts`。
- **資料取得**：每個領域在 `src/lib/<domain>/` 建立 API 請求與 `queryOptions`，於 `src/lib/<domain>/index.ts` 匯出。
- **狀態管理**：伺服器狀態用 TanStack Solid Query；認證狀態用 SolidJS store 並持久化至 `localStorage`。
- **UI 元件**：`src/components/ui/` 為 shadcn/ui 風格基礎元件，使用 `class-variance-authority` + `tailwind-merge` + `clsx`（`cn()` 輔助函式）。
- **路徑別名**：`~/*` 對應 `./src/*`。

### 4.3 sales-order-app

採三層式架構：

- `lib/layer_business/`：業務邏輯、網路（Dio + interceptor）、路由（auto_route）、服務（Signal/Provider）、相依注入（GetIt + disco）
- `lib/layer_data/`：Freezed/json_serializable 模型、Sembast 本地儲存、常數、JsonConverter
- `lib/layer_presentation/`：畫面（stories）、共用 widget、主題

Flavor 分為 `dev` / `prod`，進入點分別為 `lib/main_dev.dart` / `lib/main_prod.dart`。產生檔（`*.g.dart`、`*.freezed.dart`、`*.gform.dart`、`routes.gr.dart`、`lib/gen/*` 等）皆已入版控，修改來源檔後務必重新執行 `build_runner`。

---

## 5. 測試策略

| 子專案 | 測試框架 | 現況 | 執行指令 |
|--------|----------|------|----------|
| sales-order-backend | `testify` + `suite`；整合測試用 `ory/dockertest` | 有單元與整合測試 | `task test:unit` / `task test:integration` / `task test` |
| sales-order-frontend | Vitest 3.2 + jsdom + `@solidjs/testing-library` | 目前無測試檔 | `pnpm run test` |
| sales-order-app | `flutter_test`；整合測試用 Maestro | 目前無 `test/` 目錄；有 Maestro flow | `fvm flutter test` / `task test:maestro` |

- 後端整合測試會啟動 PostgreSQL container，再建立 schema、session、Casbin enforcer。
- App 的 Maestro flow 位於 `integration_test/.maestro/`，截圖 flow 需後端在線（預設檢查 `http://localhost:3080/`）。

---

## 6. 部署流程

### sales-order-backend

部署至 GCP Cloud Run + Cloud SQL（`hexagon-salesorder-platform` 專案）：

```bash
cd sales-order-backend
task gcp:deploy
```

此流程會自動遞增 `API_VERSION` patch、建置 Docker image、推送至 Artifact Registry、部署到 Cloud Run。

### sales-order-frontend

部署至 Firebase Hosting：

```bash
cd sales-order-frontend
task deploy
```

`firebase.json` 將 `/api/**` 與 `/customer_account_qrcode/**` 重寫至 Cloud Run 後端 `hexagon-backend`，其餘路徑回傳 `index.html`（SPA fallback）。

### sales-order-app

無 CI/CD 服務，發布以本機 Fastlane + Taskfile 為主：

```bash
cd sales-order-app

# 一鍵 release（clean → 產生碼 → build 號 +1 → appbundle + ipa）
task build

# Android Play Store Beta
cd android && bundle exec fastlane android beta

# iOS TestFlight
cd ios && bundle exec fastlane ios beta
```

商店截圖後製使用 [ButterKit](https://butterkit.app)（僅支援 App Store），原始截圖由 Maestro 產生，詳見 `sales-order-app/AGENTS.md` 第 7 節。

---

## 7. 安全與機密注意事項

### 環境變數與憑證

- **sales-order-backend**：
  - `hexagon.env` 目前被追蹤在 repo 中且包含範例/真實憑證，請避免將生產環境祕鑰寫入。
  - `.env` 與 `token.export` 已被 `.gitignore` 排除。
  - 生產環境請設定 `SESSION_SECURE=true`、`SESSION_HTTP_ONLY=true`、`SESSION_SAME_SITE=lax`（或更嚴格）。
  - NetSuite client 目前設定 `InsecureSkipVerify: true`，生產環境應移除或改為正確 TLS 設定。
  - `cmd/token` 產生的 JWT 無 `exp`，請評估是否符合安全需求。

- **sales-order-frontend**：
  - `.env` / `.env.production` 可能包含機敏資訊，已阻擋直接讀取，請勿提交真實機密。
  - `VITE_API_ACCESS_TOKEN` 與 `VITE_AUTH_SECRET` 是客戶端機密，避免在 log 或 UI 中暴露。

- **sales-order-app**：
  - `.dev.env` / `.prod.env` 已被 `.gitignore` 排除；但 `dev.env.hexagon`（含實際值）與 `lib/env/*.g.dart`（混淆後的值）在版控內，請勿外洩。
  - `Taskfile.yml` 的 `upload:ios` 任務內嵌 App Store Connect API key ID 與 issuer；私鑰（`.p8`）放在 `.private_keys/`（已 gitignore）。
  - Maestro 測試 flow（`integration_test/.maestro/salesrep_login/Flow.yaml`、`screenshots/Flow.yaml`、`screenshots_store_ios/Flow.yaml`）含明碼測試帳密，請勿對外洩漏。

### 認證與授權

- 後端支援兩種認證：
  1. **Session Cookie**：透過 `scs` 管理，store 為自訂 Postgres session store；受保護路由使用 `middleware.Authenticate(session)`。
  2. **API JWT Token**：`X-Sowinsoft-Token` header，用於中台與 Mobile App；SSE/WS 可透過 query `access_token` 傳入。
- 後端授權使用 **Casbin** RBAC with domain（tenant）；例外路徑包括 `/version`、`/swagger`、`/api/v1/netsuite/*`、`/api/v1/casl_resources`、`/api/v1/permissions`。
- 前端 `/admin` 路由 loader 會檢查認證，未通過則導向 `/signin`。
- 前端 CASL ability 規則定義於 `src/constant/casl.ts`，角色包括 `super`、`admin`、`acct`、`sales`。

### 資料安全

- 後端密碼使用 `argon2id.CreateHash` 雜湊。
- App 的 session info、cookies、HTTP 快取以 Sembast 存於應用程式快取目錄（`app_storage.db`），未額外加密。
- App 的 `AuthSessionManager.clearSession()` 會清除 session、cookies、HTTP 快取、圖片快取；`clearAuthCookies()` 僅清 cookies 且不發通知，用於登入流程失敗時避免狀態不一致。

### Deep Link

- App 處理 `/customer_account_qrcode/:customerAccount` 深度連結，基礎 URL 為 `https://frontend.hexagonty.com/customer_account_qrcode`。
- 前端 `firebase.json` 也將 `/customer_account_qrcode/**` 重寫至後端，由後端提供對應內容或導引。

---

## 8. 跨專案協作須知

### 開發時的啟動順序

1. 先啟動後端與資料庫：
   ```bash
   cd sales-order-backend
   task infra:start
   task dev
   ```
2. 再啟動前端或 App：
   ```bash
   cd sales-order-frontend && pnpm run dev
   # 或
   cd sales-order-app && fvm flutter run --flavor dev --target lib/main_dev.dart
   ```

### API Access Token

後端 `cmd/token` 產生的 JWT 必須設定到前端與 App 的環境變數中：

- 前端：`VITE_API_ACCESS_TOKEN`
- App：`.dev.env` / `.prod.env` 中的 `API_ACCESS_TOKEN`

後端 `AuthInterceptor` 與 `ApiTokenMiddleware` 皆以 `X-Sowinsoft-Token` header 驗證此 token。

### 型別同步

後端修改 DTO 後，可執行 `task typego` 產生前端 TypeScript 型別至 `frontend_types/`（前端再視需要匯入）。

### 截圖與上架素材

- Maestro 截圖 flow 位於 `sales-order-app/integration_test/.maestro/`。
- 後製專案位於 `appimg/screenshots.butterkit/`。
- iOS 成品輸出至 `appimg/export_app_store/`，Android 成品輸出至 `appimg/export_google_play/`。

---

## 9. 給 AI Agent 的快速檢查清單

開始修改前，建議確認：

1. **是否在正確的子專案工作？** 根目錄無套件設定，請直接進入 `sales-order-backend`、`sales-order-frontend` 或 `sales-order-app`。
2. **後端是否已啟動？** 前端與 App 開發、截圖測試都需要後端在線。
3. **修改後端 Ent schema 後**，是否已執行 `task ent:gen` 並確認遷移檔？
4. **修改 App 的 Freezed / json_serializable / reactive_forms / auto_route / envied / dart_mappable 來源檔後**，是否已重新執行 `build_runner` 並提交產生檔？
5. **修改前端路由後**，是否已讓 Vite 重新編譯 `routeTree.gen.ts`？
6. **新增環境變數後**，是否已更新對應的 `.env.example`、Valibot 綱目或 envied 類別？
7. **新增使用者可見文字時**，請維持繁體中文（App UI 多為硬編碼）。
8. **若更動了本文件提及的架構、指令或流程，請同步更新本文件與對應子專案的 `AGENTS.md`。**

---

最後更新：根據本專案當前工作目錄內容查證整理。
