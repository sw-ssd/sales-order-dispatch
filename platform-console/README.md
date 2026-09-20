# platform-console（平台營運主控台）

operator 用的**獨立 SPA**：租戶列表／租戶詳情與 override／方案與價目／方案權益／待收款／平台稽核。

## 邊界（硬規則）

- **只走 `platform/v1`**（`PlatformAdminService`）。不碰租戶 API（`/api/v1`），dev proxy 也不代理它。
- **不共用租戶 SPA 的路由與守衛**：路由樹在自己這裡（`src/router.tsx`），守衛是自己寫的
  `src/lib/guard.ts`。租戶 SPA 也不得引入任何 `platform.*` 能力或平台路由（S11）。
- 共用的只有 **UI 元件庫**（見下）。

## 指令

```bash
pnpm -C platform-console dev        # 開發伺服器 http://localhost:5173
pnpm -C platform-console typecheck  # tsc --noEmit
pnpm -C platform-console lint       # eslint .
pnpm -C platform-console test       # vitest run
pnpm -C platform-console build      # tsc --noEmit && vite build
```

`task console:dev|build|typecheck|lint|test` 是同一組指令（root Taskfile 的 include）。

## 認證（operator session，沿用後端既有機制，不另立一套）

後端實作見 `backend/internal/platform/operatorauth`：

| 項目 | 值 |
| --- | --- |
| 登入端點 | `GET /platform/auth/google`（OIDC；回呼 `/platform/auth/google/callback`） |
| session cookie | `platform_session`（**HttpOnly**、`Path=/platform`、`Domain` 由 `config.Platform.CookieDomain` 決定） |
| RPC 前綴 | `/platform/platform.v1.PlatformAdminService/…`（cookie 的 Path 必須涵蓋 RPC 路徑，見 operatorauth 的註解） |
| 登入後導向 | `config.Platform.ConsoleURL` |

前端因此只有兩個必要動作：

1. **每個 RPC 帶 cookie**：`src/lib/api.ts` 的 transport 以 `credentials: "include"` 包住 fetch。前端看不到
   HttpOnly cookie 的內容，所以「有沒有登入」只能問後端。
2. **知道 API 的 base URL**：`VITE_API_BASE_URL`（預設空＝同 origin）。同 origin 部署（反代把
   `/platform` 指到 API）時留空即可；console 與 API 不同 origin 時設成 API 來源。

```bash
# 跨 origin 部署時的 .env（platform-console/.env）
VITE_API_BASE_URL=https://api.example.com
```

## 開發環境（dev proxy）

console 與 API 不同 port，**登入流程必須經由 proxy**，否則 cookie 不會落在 console 的 origin：

```ts
// vite.config.ts（節錄）
server: {
  port: 5173,
  proxy: {
    // 只代理平台路徑（登入端點 ＋ RPC）。租戶 API /api/v1 不代理 —— console 只走 platform/v1。
    "/platform": { target: "http://localhost:3080", changeOrigin: true },
  },
},
```

也就是：後端起在 `:3080`（`task backend:dev`），console 起在 `:5173`；瀏覽器對 console 說
`/platform/auth/google`，Vite 轉給 API，Set-Cookie 由 console 的 origin 收下（cookie 不綁 port）。
開發環境的 `CookieDomain` 留空即 host-only（設 `Domain=localhost` 無效）。

## 登入與登出

- **登入**：未登入者導向 `/login`，按「以 Google 登入」→ 後端 OIDC → 設 cookie → 導回 `ConsoleURL`。
- **登出（v1 的已知缺口）**：v1 **沒有登出端點**，且 `platform_session` 是 HttpOnly，前端**無法**刪除它。
  因此 `src/lib/session.ts` 的 `logout()` 只清本地狀態（畫面回到登入頁），cookie 仍有效至效期結束
  （12h）。要真正即時撤銷 operator 存取，只有後端做得到：停用 `platform.operators` 該列
  （`operatorauth.verify` **每次請求**都回查白名單，所以停用立即生效）。想補真正的登出＝後端加
  `Logout` RPC ＋ 可撤銷清單，不在本任務範圍。

## 路由守衛只是 UX（後端是唯一決策者）

`src/lib/guard.ts` 的 `requireOperator()` 掛在六個頁面路由的 `beforeLoad`：未登入 → 導向 `/login`。
**它不是授權**——繞過前端（改前端碼、直接打 API）拿不到任何跨租戶資料，因為
`PlatformAdminService` 由 `operatorauth.Interceptor` 擋下，每個請求都驗簽章、audience 並回查
operator 白名單。前端隱藏／導向只是少一次「進得去但每個 RPC 都 401」的畫面。

「有沒有登入」用**單一探針**判斷：`listTenants({page:1,pageSize:1})`——不新增認證端點，
任何失敗（401／連線失敗／非白名單）一律視為未登入（fail-closed），結果在頁面生命週期內快取。

| 路徑 | 頁面 |
| --- | --- |
| `/` | 導向 `/tenants` |
| `/login` | 登入引導（唯一不受守衛保護的頁面） |
| `/tenants`、`/tenants/$tenantId` | 租戶列表／詳情＋override |
| `/plans`、`/entitlements` | 方案與價目／方案權益矩陣 |
| `/receivables`、`/audit` | 待收款／平台稽核 |

## UI 元件庫：以 alias 共用租戶 SPA 的實作

`frontend/src/components/ui`（Ark UI × Tailkit）就是唯一實作，console 不複製：

```ts
// vite.config.ts（節錄）＋ tsconfig.json 的 paths
alias: {
  "@ui/": "<repo>/frontend/src/components/ui/",
  "@ui": "<repo>/frontend/src/components/ui/index.ts",
  "@/": "<repo>/frontend/src/",
  "~/": "<repo>/frontend/src/",
},
```

- `@ui/*` 是頁面用的入口；`@/`、`~/` 是**元件庫自己的內部 import**（`@/lib/cn`、`~/components/ui/...`）
  原樣解析所需——沒有這兩個，alias 進去的元件會在 import 階段就失敗。
- console 自家程式碼一律相對路徑（`./lib/api`），不靠 `~`（那是租戶 SPA 的根）。
- 樣式同源：`src/index.css` 直接 `@import "../../frontend/src/index.css"`（設計 tokens 不複製），
  並以 `@source "../../frontend/src/components/ui"` 讓 Tailwind 掃到共用元件的 class
  （Vite root 在 platform-console，自動掃描不會涵蓋 frontend）。
- 因此 console 必須自備元件庫的執行期相依：`@ark-ui/solid`、`lucide-solid`、`class-variance-authority`、
  `clsx`、`tailwind-merge`、`@tanstack/solid-router`（sidebar 用到）、`tailwindcss` 及其 Vite 插件。
- 三個 app 之後若出現第三個消費者再抽共用 package（目前兩個，alias 足夠）。

## 產生檔（不可手改）

| 檔案 | 來源 |
| --- | --- |
| `src/lib/proto/**` | `task proto:gen`（`backend/buf.gen.yaml` 的 `bufbuild/es` → 本目錄），CI 有冪等閘門 |
| `src/lib/errcode.ts` | `cd backend && go generate ./internal/errcode`，CI 有冪等閘門 |

錯誤顯示一律走 `src/lib/errors.ts`：取 `ConnectError` 的 `ErrorInfo`（碼 ＋ 已渲染訊息），
訊息缺席時才用 `errcode.ts` 的碼表投影補位——**不硬編訊息字串**。
