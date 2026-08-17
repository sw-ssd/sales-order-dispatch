# 多公司訂出貨系統 1.0 — App Flutter 技術棧定案（設計文件）

> 日期：2026-08-04
> 狀態：**草案，待使用者審閱**（審閱通過後：新增決策 D29 + 規格書升版 v1.0.35 + 執行計畫升版）
> 對應：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）§2.1 / §9.6、`docs/superpowers/specs/2026-07-19-sales-order-1.0-decisions.md`（D1–D28）、`docs/superpowers/plans/2026-07-17-sales-order-1-0-tasks.md`（v2.9.0）
> 本文件為 brainstorming 流程產出；實作前須經使用者審閱本文件並依 writing-plans 產生執行計畫。

---

## 1. Context

1.0 規格 v1.0.34 已凍結為實作基準，App 技術棧原定為「Flutter 3.35.2 + connectrpc + Sembast + flutter_secure_storage」（§2.1）。本文件依 2026-08-04 討論，將 App 技術棧完整定案：

- 狀態管理：**solidart**（+ flutter_solidart）
- 依賴注入：**disco**（scoped providers）
- 路由：**auto_route**
- 資料層：**fquery**（含 Sembast 唯讀快取鏡像）——經與 dio_cache_interceptor 比較後定案（見 §3）

既有定案不變：Connect-RPC 為業務 API 唯一來源（D4）、App 認證 JWT + refresh 旋轉 + `token_version` 撤銷（D5）、下單需連網（§9.6）、三端覆蓋率 ≥70%（D21）。本文件未變更任何後端 / Web（SolidJS）決策。

## 2. 決策摘要（將寫入決策記錄為 D29）

| # | 決策 | 一句話 |
|---|---|---|
| D29 | App 技術棧定案 | solidart（狀態）+ disco（scoped DI）+ auto_route（路由）+ fquery（server state）+ Sembast（唯讀快取鏡像）+ connectrpc（既有）+ flutter_secure_storage（既有）；**dio_cache_interceptor 不採用**（§3 比較） |
| D29-2 | 持久化唯讀快取 | 「快取登入資訊、客戶列表、商品列表、訂單歷史；下單仍需連網」（§9.6）落實為：fquery 記憶體快取 + Sembast write-through 鏡像，App 重啟 / 暫時離線仍可讀取上次清單；**不實作離線寫入佇列** |
| D29-3 | Flutter 版本升版 | disco 2.0.0 / solidart 2.8.6 要求 Dart SDK ^3.10.0 → Flutter 升至 stable（Dart ≥3.10），取代凍結規格之 3.35.2（Dart 3.9） |

## 3. dio_cache_interceptor vs fquery 比較（結論：fquery）

比較基準：本 App 全部業務 API 為 Connect-RPC unary（POST 至常數 URL，如 `/v1/OrdersService/ListOrders`）；WatchBoard server streaming 為 Web 專屬（§7.1），App 無串流需求。持久化唯讀快取為需求（D29-2）。

| 面向 | dio_cache_interceptor 4.0.7 | fquery 3.1.0 |
|---|---|---|
| 協議適配 | **硬傷**。connect-dart 原生吃 `http.Client`（`protocol.Transport(baseUrl:, codec:, httpClient:)` 可注入），不吃 Dio → 需自寫 `http.BaseClient → Dio` 橋接層 + 自訂 Transport | 無。fetcher 直接呼叫 connect-dart generated client，零客製層 |
| 快取 key 語意 | **硬傷**。預設僅快取 GET；Connect 全 POST → 須設 `methods: [Post]` + `onlyIdempotent` + 自訂 `keyBuilder`（body 折進 key），否則同 URL 不同篩選互撞（cache poisoning）；每個可快取端點逐一設定 | QueryKey 由應用層定義（`['customers']`、`['orders', {status}]`），天然精確 |
| 狀態語意 | 透明 HTTP 快取：無 loading/error/data、無 retry/dedup、無 mutation state、無 invalidate API（僅 per-URL delete）→ 全部需在 solidart 手刻 | 內建 QueryBuilder/MutationBuilder 狀態、retry、dedup、`staleDuration`、`refetchInterval` 輪詢、`refetchOnMount`、dependent/parallel/infinite queries |
| 失效機制 | 手動 per-URL delete | `invalidateQueries(['orders'])`（前綴匹配 / `exact: true`）、`setQueryData` 手動更新、`removeQueries` |
| 持久化 | 磁碟 store 開箱即用（durable offline 免費） | 記憶體 only；以 Sembast write-through 鏡像補足（薄層，見 §8.4） |
| 依賴 | dio ^5.2.0+1、http_cache_core | flutter_hooks ^0.21.2、freezed ^3.0.5（widget API 下不須自行寫 hook/codegen） |
| 生態風險 | 成熟（llfbandit，長期維護） | 較新、單一維護者（163★）；與 solidart/disco 同屬小生態系風險，已由使用者接受 |

**結論**：採用 **fquery**。決定性因子為協議適配成本——dio_cache_interceptor 的「Dio 橋接層 + POST key 客製 + 手刻 server state」總成本遠高於 fquery 的「薄 Sembast 持久層」，且 fquery 的 mutation/invalidation 語意與本 App 業務流程（下單、退貨、帳號管理、專屬商品改名）一一對應。檔案上傳/下載、QR 兌換、公開資訊等 REST 端點與本決策無關，不進任何快取，以 http 或 dio 實作皆可（不引快取攔截器）。

## 4. 技術棧與版本約束

| 層 | 套件 | 版本 | 角色 |
|---|---|---|---|
| 狀態 | solidart + flutter_solidart | 2.8.6 | signals / computed / Resource；session 與畫面局部狀態 |
| DI | disco | 2.0.0 | scoped providers（`ProviderScope` 分層 + `ProviderPortal`）；多 providers 同型別、隨 widget tree 生命週期 dispose |
| 路由 | auto_route | 11.1.0 | 宣告式路由 + codegen + guards |
| 資料層 | fquery | 3.1.0 | server state（query / mutation / invalidate / retry / 輪詢）；**用 widget API（QueryBuilder/MutationBuilder），不引入 flutter_hooks** |
| 持久化 | sembast | 沿用凍結棧 | 唯讀快取鏡像（write-through，見 §8.4） |
| API | connectrpc（connect-dart） | 最新穩定 | 業務 API 唯一來源（D4）；`protocol.Transport` + interceptor；proto 由 `backend/proto` 以 buf 產生至 `app/lib/gen` |
| 認證 | flutter_secure_storage + flutter_web_auth | 沿用凍結棧 | JWT 存取（D5）；員工 Google OAuth2 PKCE（§4.1） |
| REST（檔案/QR/公開資訊） | http 或 dio（不引快取） | — | 與資料層決策無關 |

**版本約束**：disco 2.0.0 / solidart 2.8.6 要求 Dart SDK ^3.10.0 → **Flutter 升至 stable（Dart ≥3.10）**；凍結規格 §2.1「Flutter 3.35.2（Dart 3.9）」升版。鎖版於 pubspec，CI 依賴掃描沿用（§14.3）。

## 5. 分層架構（feature-first）

```
app/lib/
├── main.dart / app.dart          # CacheProvider（fquery）+ 根 ProviderScope（disco）+ MaterialApp.router
├── core/
│   ├── config/                   # env、API base URL
│   ├── auth/                     # TokenStore（secure storage）、SessionController、auth interceptor
│   ├── network/                  # Transport factory、REST http client、ConnectException → AppError
│   ├── cache/                    # QueryCache factory、Sembast store、CacheRepository
│   └── errors/                   # AppError 五類：網路 / 逾時 / 認證 / 權限 / 業務
├── features/
│   ├── auth/                     # 登入（身分選擇）、員工 OAuth 註冊完成頁、QR 登入
│   ├── home/                     # 首頁：公告消息、快速下單入口（底部）
│   ├── products/                 # 商品、客戶專屬商品、快速下單流程
│   ├── orders/                   # 訂單歷史、訂單詳情、退貨申請
│   ├── account/                  # 店家帳號管理（主帳號，D22/D28）
│   ├── profile/                  # 功能頁：客戶查詢、QR、設定、登出、關於
│   └── notifications/            # 站內通知列表、FCM handler
├── router/                       # app_router.dart + AuthGuard / RoleGuard
└── gen/                          # buf 產出（.pb.dart + connect client）
```

每個 feature 內分 `presentation/`（pages、widgets）、`application/`（controllers）、`data/`（repositories）。原則：repositories 包 generated client；controllers 以 fquery 組查詢/突變；pages 只消費 QueryBuilder + SignalBuilder；單一職責、介面清楚、可獨立測試。

## 6. 狀態與 DI（disco + solidart 分工）

- **根 scope**（MaterialApp 上）：AppConfig、TokenStore、ConnectClient、QueryCache（fquery `CacheProvider`）、SembastCacheRepository、各 feature repositories。
- **auth scope**（登入後建立）：SessionController——solidart signals 存 current user、role、company、department、`is_primary`（主帳號標記）。
- **路由級 scope**：依 disco 官方 auto_route 範例模式，快速下單等有狀態流程以 `ProviderScope` 包 route 樹，注入 `OrderDraftController`（品項 / 數量 / 單位換算 / 備註 / 客戶 signals），流程結束即 dispose，不污染全域。
- **Dialog / bottom sheet**：以 `ProviderPortal` 取得 providers。
- **分工原則**：server state 全在 fquery；solidart 只管 session 與畫面局部狀態，兩者不重疊、不互相取代。

## 7. 路由（auto_route 11.x）

- 未登入 → `/login`（身分選擇：我是店家 / 我是業務）；員工 OAuth 首登（guest）→ `/register-complete`（選公司 + 姓名，D6）。
- 登入後 → `/shell`：底部導覽 4 tab（首頁 / 商品 / 訂單歷史 / 功能），`AutoTabsScaffold`。
- 子流程：`/products/quick-order`（選客戶 → 帶出客戶專屬商品 → 總表 / 手打 → 提交）；`/orders/:id`；`/orders/:id/return`（退貨：歷史訂單勾選 + 專屬商品清單並存，D25）；`/account/sub-accounts`（主帳號自助管理，D28）；`/qr`（店家子帳號 QR 登入兌換）。
- **Guards**：
  - `AuthGuard`：無有效 session 一律導向 `/login`。
  - `RoleGuard`：**主帳號（`is_primary`）僅能進入帳號管理與功能頁，業務流程路由一律擋下**——對應 D22/D28「主帳號業務 API 403」；UI 層防呆 + 後端授權雙重防護（D3 精神）。

## 8. 資料層（fquery + Sembast 持久層）

### 8.1 QueryKey 規範

| QueryKey | 資料 | 可離線鏡像 |
|---|---|---|
| `['customers']` | 客戶列表（業務） | ✅ |
| `['customer-products', customerId]` | 客戶專屬商品清單 | ✅ |
| `['products', {filter}]` | 商品總表 | ✅ |
| `['orders', {status, page}]` | 訂單歷史 | ✅ |
| `['announcements']` | 首頁公告 | ✅ |
| `['notifications']` | 站內通知 | ❌（FCM 導向） |
| `['metadicts', type]` | 字典檔（單位、來源…） | ❌（量小，重抓即可） |
| `['routes']` / `['warehouses']` | 車次 / 倉別 | ❌ |
| `['returns']` / `['returns', id]` | 退貨申請 | ❌ |

### 8.2 讀取與輪詢策略

- 讀：`QueryBuilder` / `QueriesBuilder`（平行）；寫：`MutationBuilder`。
- 預設 `QueryCache` 組態（`DefaultQueryOptions`）：`staleDuration` 2 分鐘、`refetchOnMount: stale`、`retryCount` 2（`retryDelay` 固定延遲）、`cacheDuration` 20 分鐘 GC。
- 公告 / 通知：`refetchInterval` 5 分鐘輪詢；其餘查詢不輪詢（依 FCM 推播 + 手動刷新）。

### 8.3 Mutation 與失效對照表

| 業務動作 | Mutation 後 invalidate |
|---|---|
| 建立訂單（業務 / 客戶） | `['orders']`、`['notifications']`（D23 推播路由，通知徽章） |
| 退貨申請送出 | `['orders', id]`、`['returns']` |
| 客戶專屬商品改名 / 刪除 / qty=0 | `['customer-products', id]` |
| 新增客戶（App 手動表單） | `['customers']` |
| 帳號管理（新增 / 停用 / 重置子帳號） | `['customers']`（子帳號清單） |
| 登出 | `removeQueries` 全清（含鏡像） |

### 8.4 持久化唯讀快取（Sembast write-through）

- `CacheRepository`（core/cache）以 Sembast 為鏡像，介面：`read(key)` / `write(key, json)` / `clear()`。
- **write-through**：可鏡像清單（§8.1 標 ✅）於 fetcher 成功後寫入鏡像。
- **seed**：App 啟動時若有鏡像資料即以 `setQueryData` 回填（UI 立即顯示並標 stale）；fetcher 失敗時保留已顯示的 seed 資料（標離線 / 可下拉重試），僅在無鏡像資料時才顯示錯誤態。
- **登出清空**：`removeQueries` + 鏡像 `clear()`（防跨帳號資料殘留）。
- **下單維持需連網**（§9.6 不變）：離線時下單按鈕禁用。

## 9. 認證整合（D5 落地）

- 客戶：`customer_code` + 密碼 → `AuthService` → 後端核發 access JWT（1h）+ refresh（30d 旋轉）→ `flutter_secure_storage`。
- 員工：`flutter_web_auth` PKCE → 回調 → 後端換 JWT；guest 首登導 `/register-complete`（D6）。
- connect interceptor：每請求加 `authorization: Bearer <jwt>`；**401 → refresh 旋轉（單飛鎖，並發請求共享一次 refresh）→ 重試一次**；**403（`tv` 不符，D5）→ 清 token + `removeQueries` + 鏡像 `clear()` → 導回 `/login`**。
- QR 登入（店家子帳號）：REST 兌換 token → 以子帳號完成登入（D22/D28）。

## 10. 錯誤處理與離線

- `ConnectException` → `AppError` 五類：網路（SocketException / 連線逾時 / DeadlineExceeded）、認證（401 / refresh 失敗）、權限（403 / 404）、業務（Connect code + message → 使用者可讀文案）、未預期（其餘例外）。
- UI：統一 snackbar / 離線橫幅；離線且有鏡像 seed 時顯示「快取資料（離線）」指示。
- 鏡像 seed 失敗或無鏡像資料：顯示空態 + 重試按鈕，不假裝成功。

## 11. 測試（對齊 D21，覆蓋率 ≥70%）

- repository 單元測試：以 mock `Transport`（implement connect `Transport` interface，回傳固定回應）測 fetcher / 序列化。
- `CacheRepository` 測試：Sembast 記憶體實例，驗證 write-through / seed / clear。
- auth interceptor 測試：401 → refresh → retry 序列、403 → 登出、refresh 單飛鎖。
- QueryBuilder / MutationBuilder widget 測試：loading / error / data 三態與 invalidate 觸發。
- Maestro 整合測試：登入（店家 + 業務）→ 快速下單 → 訂單歷史 → 退貨申請 → 帳號管理（主帳號）。

## 12. 文件回寫（審閱通過後執行）

1. **決策記錄**：新增 **D29**（含 D29-2 / D29-3，本文件 §2 摘要），比照 D14 模式加註修訂來源。
2. **規格書升版 v1.0.35**：§2.1 App 技術列更新；§9.6 離線快取補實作輪廓（fquery + Sembast 鏡像）；§16 後續步驟。
3. **執行計畫升版**：Phase 0/1 App 初始化 Task 更新（scaffold、disco DI 分層、auto_route、fquery 資料層、auth wiring、Sembast 鏡像）；Task 編號維持連續。
4. 本設計文件為來源文件，保留於 `docs/superpowers/specs/`。
5. monorepo 建立（D1）後一併 commit 本文件。

## 13. Open Questions

| 問題 | 影響 | 狀態 |
|---|---|---|
| Flutter stable 確切版本（Dart ≥3.10）與 pubspec 鎖版 | 建置 | 於執行計畫 Phase 0 確認 |
| fquery 3.1.0 與 flutter_hooks / freezed 相依的鎖版與 CI 掃描 | 依賴安全 | 於執行計畫 Phase 0 處理 |
| 公告 / 通知輪詢間隔（草案 5 分鐘）是否需依 FCM 到達率調整 | UX | 試用回饋後收斂 |

## 14. 修訂記錄

| 修訂號 | 日期 | 修訂內容 | 修訂者 |
|---|---|---|---|
| v0.1.0 | 2026-08-04 | 初版草案（brainstorming 產出，待審閱） | 開發團隊 |
