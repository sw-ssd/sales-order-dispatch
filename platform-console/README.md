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
2. **知道 API 的 base URL 與掛載前綴**：`VITE_API_BASE_URL`（預設空＝同 origin）＋ `PLATFORM_PATH=/platform`。
   RPC 的實際路徑是 `/platform/platform.v1.PlatformAdminService/…`——後端把平台 RPC 掛在字面
   `/platform/` 之下（`backend/internal/server/domains.go`），那段前綴同時是 cookie 的 `Path`，
   **少一段就是 404 且 cookie 不會被送出**（RFC 6265 逐段前綴比對：`/platform` 對 `/platform.v1.…`
   不成立，未涵蓋的第一個字元是 `.`）。兩端各釘一次：後端
   `internal/server/platform_admin_mount_integration_test.go` 的 `TestIntegrationPlatformAdminMount`，
   前端 `src/lib/api.test.ts`。

### 跨來源部署：**目前不支援**（console 必須與 API 同 origin）

`VITE_API_BASE_URL` 是為了「同一個反代後面、但 console 由不同路徑／服務提供」這類彈性，
**不是**拿來跨來源直連 API 的：

- 後端 CORS 允許清單目前只有 `http://localhost:3000`（`backend/internal/server/server.go`），
  console 的 dev port 是 5173 → 跨來源請求會被瀏覽器擋下（連回應都拿不到）。
- cookie 是 `SameSite=Lax`（`operatorauth/service.go`）。同 site（同 registrable domain，例如
  `console.example.com` 與 `api.example.com`）＋ 設 `CookieDomain=example.com` 可行；真正**跨 site**
  需要 `SameSite=None; Secure`，那是後端變更，不在本任務範圍。

**所以現階段一律同 origin**：dev 用 Vite proxy（下節），正式由反代把 `/platform` 指到 API。
跨來源直連需要的後端變更（CORS 允許清單 ＋ 視情況 `SameSite`）列為後續待辦。

> 失敗樣態：若路徑／來源不對，症狀是「登入完又被導回 `/login`」——探針 fail-closed（後端 404/401），
> 不是登入壞了。用 `curl` 直接看 `/platform/platform.v1.PlatformAdminService/ListTenants` 的回應碼最快。

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
`/platform/auth/google` 與 `/platform/platform.v1.…`，Vite 轉給 API，Set-Cookie 由 console 的 origin
收下（cookie 不綁 port）。開發環境的 `CookieDomain` 留空即 host-only（設 `Domain=localhost` 無效）。

> **驗證的坑**：Vite proxy 的前綴 `/platform` 對 `/platform.v1.PlatformAdminService/…`（少了掛載前綴的
> 錯誤路徑）**也成立**——請求會被轉發，只是後端回 404。所以「有被代理」不等於「路徑對」：
> 手動驗證要看**回應碼**，例如
>
> ```bash
> curl -s -o /dev/null -w "%{http_code}\n" -X POST -H 'content-type: application/json' -d '{}' \
>   http://localhost:5173/platform/platform.v1.PlatformAdminService/ListTenants   # 未帶 cookie 應為 401，不是 404
> ```

## 登入與登出

- **登入**：未登入者導向 `/login`，按「以 Google 登入」→ 後端 OIDC → 設 cookie → 導回 `ConsoleURL`。
- **登入狀態**：以單一探針（`listTenants` 1 筆）判斷，結果快取 **30 秒**（`PROBE_TTL_MS`）。cookie 的效期
  只有後端知道，永久快取會讓 session 失效後導航仍穿過守衛、只看到頁面上的 401 文案卻不回首頁。
- **登出（v1 的已知缺口）**：v1 **沒有登出端點**，且 `platform_session` 是 HttpOnly，前端**無法**刪除它。
  因此 `src/lib/session.ts` 的 `logout()` 只清本地狀態（畫面回到登入頁）並**停止探針**（否則 TTL 一過，
  後端會照仍在的 cookie 把人放回主控台）；cookie 仍有效至效期結束（12h）。要真正即時撤銷 operator 存取，
  只有後端做得到：停用 `platform.operators` 該列（`operatorauth.verify` **每次請求**都回查白名單，
  所以停用立即生效）。想補真正的登出＝後端加 `Logout` RPC ＋ 可撤銷清單，不在本任務範圍。

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
| `/login` | 登入引導（唯一不受守衛保護、也不套 console 外框的頁面） |
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
