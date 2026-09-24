# 訂出貨系統 — AI Agent 專案導覽

> 本文件位於 `docs/`，是**多公司訂出貨系統 1.0**（monorepo，`backend`／`frontend`／`app` 三個子專案同倉）的導覽。各子專案的 `AGENTS.md` 是該子專案的權威指引，修改前請一併參閱。
>
> **規劃文件**：整體規劃整合於 `docs/PLANNING_OVERVIEW.md`，決策層為 `docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md`（D1–D32），計畫現況見 `docs/superpowers/plans/README.md`，規劃細節請先讀 `docs/PLANNING_OVERVIEW.md`。
> - `backend/AGENTS.md`：Go 後端詳細指引
> - `frontend/AGENTS.md`：SolidJS 前端詳細指引
> - `app/AGENTS.md`：Flutter App 詳細指引

---

## 1. 專案概覽

這是**多公司訂出貨系統 1.0** 的程式碼庫，主要用於業務員（salesrep）與客戶（customer）建立、查詢、管理銷售訂單（sales order）。三個子專案同倉，各自另有 `AGENTS.md`：

| 子專案 | 技術 | 用途 | 版本 |
|--------|------|------|------|
| `backend` | Go（`go 1.25.7`）；ent + Connect-RPC + OpenFGA + RLS + goose + Valkey | API 後端（Connect-RPC，前綴 `/api/v1`） | 見 `backend/go.mod` |
| `frontend` | SolidJS（`solid-js ^1.9.15`）+ TypeScript + Vite 8 + Tailwind CSS v4 | 網頁中台（SPA） | 見 `frontend/package.json` |
| `platform-console` | SolidJS + TypeScript + Vite | 平台營運後台（operator 用；掛在 `/platform/`） | 見 `platform-console/package.json` |
| `app` | Flutter + Dart | 跨平台行動 App（iOS／Android，雙 flavor：`dev`／`prod`） | 見 `app/pubspec.yaml`；Flutter 版本由 `.fvmrc`（`stable`）決定 |

- **主要語言**：文件與程式碼註解以**繁體中文**為主；識別字、檔案名稱、套件名稱維持英文原文。
- **API 前綴**：後端統一使用 `/api/v1`。
- **存放庫結構**：單一 Git 倉庫（根目錄即 monorepo 根）：根 `package.json`＋`pnpm-workspace.yaml`（工作區 `backend`／`frontend`／`platform-console`／`app`／`infra`，turbo 編排）與根 `Taskfile.yml`（`includes` 四個子專案，另提供 `infra:*`／`fga:*`）。子專案不再各自獨立成倉。

---

## 2. 專案結構與關鍵設定檔

```text
.
├── .github/workflows/ci.yml   # CI：Go 測試（含 integration）與前端四道 gate
├── app/                       # Flutter 行動 App（fvm；dev/prod 雙 flavor）
│   ├── lib/                   # main_dev/main_prod、config.dart、router/、features/、gen/（proto 生成碼）
│   ├── test/                  # flutter_test
│   ├── android/ / ios/        # 平台專案（ios/setup_flavors.rb 為一次性注入腳本）
│   ├── pubspec.yaml / pubspec.lock / analysis_options.yaml
│   └── AGENTS.md / Taskfile.yml / .fvmrc
├── backend/                   # Go 後端（ent + Connect-RPC + OpenFGA + RLS + goose + Valkey）
│   ├── cmd/                   # 入口：server / migrate / seed
│   ├── config/                # envconfig 逐檔設定（api、storage、observability…）
│   ├── database/migrations/   # goose 遷移檔（NNNNN_name.sql，必含 Up/Down）
│   ├── ent/                   # Ent schema 與產生碼（改 schema 後 go generate ./ent）
│   ├── internal/
│   │   ├── services/          # Connect-RPC handler（各 service 提供 Register*Services）
│   │   ├── domain/<name>/     # usecase + repository 介面
│   │   ├── proto/             # buf 產生碼（Go 型別與 *connect）
│   │   └── auth/ authz/ audit/ handlers/ server/ testsupport/
│   ├── proto/                 # proto 定義唯一來源（改動後 task proto:gen）
│   ├── third_party/           # 外部整合（database / openfga / cache）
│   ├── AGENTS.md / Taskfile.yml / go.mod / go.sum
│   └── buf.yaml / buf.gen.yaml / .air.toml / .env.example
├── docs/                      # 規劃、規範與報告
│   ├── AGENTS.md              # 本文件（導覽層級）
│   ├── PLANNING_OVERVIEW.md / FUNCTION_LIST.md
│   ├── superpowers/           # specs/（設計書與決策 D1–D32）、plans/（backend 01~09…）、reports/
│   ├── design/                # 版面與設計進度存檔
│   └── archive/               # 歸檔（docs.zip）
├── platform-console/          # 平台營運後台（SolidJS + Vite；獨立於租戶端，有自己的 proto 生成碼）
│   ├── src/                   # pages/、lib/api.ts、lib/proto/
│   └── package.json / Taskfile.yml
├── frontend/                  # SolidJS 租戶網頁中台（Vite + Vitest）
│   ├── src/
│   │   ├── features/<name>/   # 頁面與領域元件（users、auth…，各自 pages/ 與 components/）
│   │   ├── components/ui/     # Ark UI（行為）× 語意 token 的基礎元件（含 sidebar/、demo/）
│   │   ├── components/layout/ # AppShell / Sidebar / Topbar
│   │   ├── lib/               # transport、query-client、ability、proto 生成碼
│   │   └── router/ App.tsx main.tsx index.css test-setup.ts
│   ├── scripts/               # （casl-golden-gen.mjs 已失效：CASL 已移除，見下方註）
│   ├── index.html / vite.config.ts / vitest.config.ts / tsconfig.json / eslint.config.js
│   └── package.json / AGENTS.md / Taskfile.yml
├── .superpowers/sdd/          # SDD 工作區（各波計畫、報告與複審紀錄）
├── docker-compose.dev.yml     # 開發基礎設施（PostgreSQL / Valkey / Gotenberg）
├── Taskfile.yml               # 根 Taskfile：infra:start|stop、fga:start|stop、includes 三個子專案
├── package.json / pnpm-workspace.yaml / pnpm-lock.yaml / turbo.json
│                              # ↑ pnpm 工作區（backend/frontend/app）與 turbo 編排
├── fga-mcp.json               # OpenFGA MCP 連線設定（唯讀）
└── README.md
```

---

## 3. 常用建置與執行指令

### 3.1 backend（Go）

需安裝 [Task](https://taskfile.dev/) 與 Go 1.25；開發時建議一併安裝 `air`、Docker/Podman。

```bash
cd backend

# 啟動本地基礎設施（Postgres + Valkey + Mailpit）
task infra:start

# 熱重載開發（air，監聽 0.0.0.0:3080）
task dev

# 直接執行
task run

# 資料庫遷移（goose，migration 檔在 backend/database/migrations/）
task backend:migrate:up      # 套用全部未套用的 migration
task backend:migrate:status  # 目前版本
task backend:migrate:down    # 回退一版

# 種子資料（冪等；development 才會建示範租戶）
task backend:seed

# 程式碼產生
task backend:proto:gen     # buf：proto → Go / TS(前端+console) / Dart
go generate ./ent          # Ent 產生碼（改 ent/schema 後）

# 測試
task backend:test                 # go test -count=1 ./...
task backend:test:integration     # go test -tags integration -count=1 ./...
task backend:check                # fmt + vet + golangci-lint + test（CI 同款閘門）

# 品質檢查（check 已含前三項；單獨跑時）
go fmt ./...  /  go vet ./...  /  golangci-lint run

# 平台排程（單趟執行；非長駐服務）
task backend:platform:cron
```

**註**：本專案**沒有** `task swagger`／`task oapigen`／`task typego`／`task gcp:deploy`（舊版殘留）；對外契約一律以 protobuf 為來源，型別由 `task backend:proto:gen` 產生。

### 3.2 frontend（SolidJS）

需安裝 pnpm。

```bash
cd frontend
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

### 3.3 app（Flutter）

建議透過 FVM（Flutter Version Management）使用固定版本 `3.35.2`；VS Code 啟動設定位於 `.vscode/launch.json`。

```bash
cd app
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
task app:test              # flutter test（單元／Widget）
fvm flutter test integration_test/app_flow_test.dart            # iOS 模擬器實機整合
fvm flutter test integration_test/app_flow_test.dart -d emulator-5554 --flavor dev  # Android（必須帶 flavor）
```

> 注意：`app/Taskfile.yml` 內部分任務直接呼叫 `flutter` / `dart`，若系統 PATH 未掛載 Flutter，請改用 `fvm flutter ...` / `fvm dart ...` 手動執行對應指令。

---

## 4. 程式碼組織與主要慣例

### 4.0 程式碼索引與查詢（codebase-memory-mcp）

進行**任何**程式碼索引或程式碼查詢（找定義、找實作、找呼叫端／被呼叫端、跨專案跳轉、影響範圍分析、結構探索）時，**一律先經過 codebase-memory-mcp**，以其作為程式碼知識圖譜的權威來源，取代直覺式廣域 grep。具體做法：

- 查定義／實作／關係：先以 codebase-memory 的圖形查詢（搜尋節點、查關係、追蹤呼叫路徑）定位，不足處再以 grep 補足。
- 改動涉及既有程式碼時，先確認目標檔在 codebase-memory 索引中的覆蓋狀態（coverage）；對未完整索引（parse_partial／skipped）的檔，以 grep 與直接讀檔為準。
- 索引的建置／更新／狀態亦經 codebase-memory-mcp 執行（index_status / index_repository / …）。

> **定案：專案程式碼索引與查詢皆需先經過 codebase-memory-mcp。**

### 4.1 backend

後端以**服務層**為主（`internal/services/`，121 檔、一檔一 service 或一組相關 service），再依職責分層：

- `internal/services/`：Connect-RPC handler（各 service 提供 `Register*Service(s)`）＋服務層業務邏輯與授權（`requireScope`／`requireRole`／各 domain 的 scope 函式）。
- `internal/domain/<name>/`：**真正的 DDD 子域**（`auth`／`customers`／`fileassets`／`products`／`roles`），含 usecase 與 repository 介面。
- `internal/platform/`：SaaS 平台域（billing／entitlements／cron／operatorauth／store…），**不受 RLS**。
- `internal/auth`（JWT／session／密碼／角色繼承）、`internal/authz`（OpenFGA engine 與 provision）、`internal/dbtenant`（租戶交易與 RLS driver 裝飾）、`internal/audit`、`internal/errcode`（錯誤碼註冊表）、`internal/resterr`、`internal/obs/requestid`。
- `cmd/`：入口有 `server`／`migrate`／`seed`／`platform-cron`／`gen-errcodes`。

路由與中介層組裝在 `internal/server/`（`server.go` 的 `authzMiddleware` ＋ `protectedRPC` 授權表；`domains.go` 掛載全部服務，一律掛在 `/api/v1` 之下）。

### 4.2 frontend

- **路由**：TanStack Solid Router，**程式式定義**於 `src/router/index.tsx`（`createRoute` + `lazyRouteComponent`），**沒有**檔案式路由或 `routeTree.gen.ts`。
- **資料取得**：領域為單位放 `src/features/<name>/`（各自 `pages/`、`queries.ts`），共用查詢工具在 `src/lib/`。
- **狀態管理**：伺服器狀態用 TanStack Solid Query；認證狀態為 SolidJS store。
- **UI 元件**：`src/components/ui/` 為 Ark UI（行為）＋語意 token 的基礎元件；樣式以 Tailwind CSS v4 與 `cn()`（`clsx` + `tailwind-merge`）組合。
- **權限**：`src/lib/ability/` 的權限集合（由後端 `GetAbility` 投影載入）；`@casl/ability` 已移除。

### 4.3 app（Flutter）

- `lib/core/`：API transport（單一 `ApiClient`，Bearer ＋ 401 單飛 refresh 重試）、設定（`config.dart`）。
- `lib/features/<name>/`：依領域分（`auth`／`orders`／`returns`／`notifications`／`shell`…），各自 pages／repository／provider。
- `lib/router/`：`auto_route` 路由表。
- `lib/ui/`：**平台自適應元件**（`adaptive.dart`：iOS 走 Cupertino、Android 走 Material，整合測試以語意標籤斷言）；`ui/themes/` 為主題。
- `lib/gen/`：proto 產生碼（`task app:gen` 之 build_runner 產生部分為 Riverpod/Freezed 等）。

Flavor 分 `dev` / `prod`，入口 `lib/main_dev.dart` / `lib/main_prod.dart`（**API base URL 直接寫在入口檔**）。產生檔已入版控，改來源檔後須重跑 `task app:gen`。

---

## 5. 測試策略

| 子專案 | 測試框架 | 現況 | 執行指令 |
|--------|----------|------|----------|
| backend | 標準 `testing`；整合測試用 `testcontainers-go`（tag `integration`） | 166 個測試檔（含 75 個整合探針） | `task backend:test` / `task backend:test:integration` / `task backend:check` |
| frontend | Vitest + jsdom + `@solidjs/testing-library` | 41 個測試檔（366 tests） | `task frontend:test` |
| console | Vitest | 12 個測試檔 | `task console:test` |
| app | `flutter_test` ＋ `integration_test`（真機／模擬器） | 單元 + 整合各一組 | `task app:test`（單元）／`fvm flutter test integration_test/...`（實機） |

- 後端整合測試以 `testcontainers-go` 起一次性 PostgreSQL，套真 migration 後建 schema／session／OpenFGA engine（**非** `ory/dockertest`）。
- 整合測試必須帶 `-count=1`（Taskfile 與 CI 皆已內建），否則快取會回報假綠。
- 前端 `vitest` 不快取通過結果，故不需 `-count=1`。
- App 目前**沒有** Maestro flow（舊版殘留）；實機驗證走 `integration_test/app_flow_test.dart`（Android 需 `--flavor dev`）。

---

## 6. 部署流程

本倉**目前沒有部署管線**。`.github/workflows/ci.yml` 只有五個驗證 job，**沒有任何 deploy job**：

| job | 內容 |
|-----|------|
| `go` | lint、`go test -count=1 ./...`、錯誤碼 baseline 守門、產生檔冪等（proto 三端＋console 的 errcode.ts）、build、govulncheck |
| `go-integration` | `go test -tags integration -count=1 ./...`（testcontainers 拋棄式容器） |
| `frontend` | 租戶中台四道 gate（typecheck／lint／test／build） |
| `platform-console` | typecheck／lint／test／build |
| `flutter` | `flutter pub get` + `flutter analyze`（**不建置 App**） |

repo 內也無 `firebase.json`／Dockerfile／fastlane 設定。部署方式待補（規劃見 `docs/PLANNING_OVERVIEW.md`）。知名取捨：`flutter` job 只做 `pub get` + `analyze`，**不建置 App**，故 Android／iOS 的原生建置問題（如 `app/android/gradle/wrapper` 變更）在 CI 不會被驗到 —— 需在本機以 `--flavor dev` 實機測試把關。

---

## 7. 安全與機密注意事項

### 環境變數與憑證

- **backend**：
  - 唯一的 env 範本是 `backend/.env.example`（**不含真實憑證**）；`.env*` 已被 `.gitignore` 排除（保留 `.env.example`）。真實密鑰只放部署環境或本地未追蹤檔案。
  - 生產環境請設定 `SESSION_SECURE=true`、`SESSION_SAME_SITE=lax`（或更嚴格）；`JWT_SECRET` 留空或預設值會被 `Init()` fail-fast 拒絕。
  - `API_TOKENS`（若啟用）只放 token 的 **SHA-256**，不得放原文。
  - `cmd/token` 產生的 JWT 無 `exp`，請評估是否符合安全需求。

- **frontend**：
  - `.env` / `.env.production` 可能包含機敏資訊，已阻擋直接讀取，請勿提交真實機密。
- **app**：無 env 檔機制 —— API base URL 直接寫在 `lib/main_dev.dart`（`http://localhost:3080/api/v1`）與 `lib/main_prod.dart`（`https://api.example.com/api/v1`，佔位）。**prod 的 base URL 目前是範例值，上架前必須改為真實網域**。無 fastlane／截圖 flow／CI 發布設定。

### 認證與授權

- 後端支援三種認證（皆由 `server.authzMiddleware` 逐請求驗證，**順序即優先序**）：
  1. **Session Cookie**：Web 中台，`scs` session（store 為 Valkey）。
  2. **API JWT Token**：`Authorization: Bearer` header，用於 Mobile App 與 API 客戶端（access 1h／refresh 30d，旋轉制，帶 `tv` claim）。
  3. **`X-Api-Token`**：靜態、**僅 server-to-server**（見上方「認證用憑證」）。
- 後端授權為 **OpenFGA + RLS**（D32）：受保護 RPC 由 middleware 做 OpenFGA `Check`（developer/super 逃生門、無引擎 fail-closed），`role_permissions` 為權限定義來源並同步 tuples；服務層 `requireScope`/`requireRole`（純 Go ACL）為 fallback；資料範圍由 RLS 兜底（38 張業務表 ENABLE+FORCE，`dbtenant`）。
- 前端權限為權限集合查詢（`Can`／guards；@casl/ability 已於 2026-09-19 移除，集合由後端 `GetAbility` 投影載入）。

### 資料安全

- 後端密碼使用 `argon2id.CreateHash` 雜湊。
- App 的 session info、cookies、HTTP 快取以 Sembast 存於應用程式快取目錄（`app_storage.db`），未額外加密。
- App 的 `AuthSessionManager.clearSession()` 會清除 session、cookies、HTTP 快取、圖片快取；`clearAuthCookies()` 僅清 cookies 且不發通知，用於登入流程失敗時避免狀態不一致。

### Deep Link

App 與前端**目前都未實作** deep link 路由（`customer_account_qrcode`／`customer_account_manage` 皆無對應處理）。規格要求的兩條深層連結（QR 登入、帳號管理）見 `docs/superpowers/specs/1.0-requirements/identity-access/spec.md`；實作時需同時處理 iOS Universal Link 與 Android App Link，且連結**不得內含登入憑證**。後端已提供對應 API（QR 產生見 `CustomerService.GetCustomerQRCode`；帳號管理見 `CustomerAccountService`）。

---

## 8. 跨專案協作須知

### 開發時的啟動順序

1. 先啟動後端與資料庫：
   ```bash
   cd backend
   task infra:start
   task dev
   ```
2. 再啟動前端或 App：
   ```bash
   cd frontend && pnpm run dev
   # 或
   cd app && fvm flutter run --flavor dev --target lib/main_dev.dart
   ```

### 認證用憑證（三種，優先序固定）

`server.authzMiddleware` 每請求依序嘗試，**使用者憑證優先**（同時帶多種時不得降級成機器身分）：

1. **Web session cookie**（`scs`，store 為 Valkey）—— 網頁中台。
2. **`Authorization: Bearer <JWT>`** —— App 與 API 客戶端（access 1h／refresh 30d 旋轉制，帶 `tv` claim）。
3. **`X-Api-Token`（靜態）** —— **僅供 server-to-server／M2M**（設定見 `backend/config/auth.go` 的 `Auth.APITokens`）。

> `X-Api-Token` **MUST NOT** 配置於 Web 或 App 客戶端（規格 §4.3、細部 1.6.6）：token 是**共用長效**機密，放進前端 bundle／App 封包等於公開。設定只存原文的 SHA-256、綁定真實使用者、每組 token 另有 RPC 前綴白名單。**目前尚無任何呼叫方**（`cmd/platform-cron` 直連資料庫不經 HTTP），屬規格要求的 M2M 預備能力。
> 前端與 App **都沒有**任何 API token 環境變數；不要把憑證注入客戶端產物。

### 型別同步

protobuf 為唯一型別來源：改 `backend/proto/**` 後執行 `task backend:proto:gen`（buf），一次產生 Go（`backend/internal/proto`）、前端與 console 的 TS（各自 `src/lib/proto`）以及 App 的 Dart（`app/lib/gen`）。前端／App 一律匯入生成碼，沒有獨立的 DTO 產生步驟。

---

## 9. 給 AI Agent 的快速檢查清單

開始修改前，建議確認：

1. **是否在正確的子專案工作？** 子專案位於 `backend`、`frontend`、`app`；根目錄是 monorepo 根（pnpm 工作區與根 Taskfile），不是任一子專案的目錄。
2. **後端是否已啟動？** 前端開發與 App 實機整合測試都需要後端在線（`task backend:run`）。
3. **修改後端 Ent schema 後**，是否已執行 `go generate ./ent` 並新增對應 migration？（改 proto 則 `task backend:proto:gen`，產生檔一併提交——CI 有冪等檢查。）
4. **修改 App 帶 `@riverpod`／`@freezed`／`@JsonSerializable` 等標註的檔案後**，是否已執行 `task app:gen`（build_runner）並提交產生檔？
5. **新增後端環境變數後**，是否已更新 `backend/.env.example`？（前端目前無 env 檔機制；App 的設定值寫在 `lib/main_dev.dart`／`lib/main_prod.dart`。）
6. **新增使用者可見文字時**，請維持繁體中文。
7. **若更動了本文件提及的架構、指令或流程，請同步更新本文件與對應子專案的 `AGENTS.md`。**
8. **進行程式碼索引／查詢前**，是否已先經過 codebase-memory-mcp（見 §4.0）？
9. **改動受保護 RPC 時**，是否已在 `server.protectedRPC` 補上（resource, action）？未列入的路徑不受 OpenFGA 閘門保護。

---

*最後更新：2026-09-25（修正舊版殘留：目錄名、任務名、認證機制、測試框架與部署段落；以實際 repo 狀態逐項查證）*
