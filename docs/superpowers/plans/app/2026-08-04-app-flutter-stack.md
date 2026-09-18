# App Flutter 技術棧與認證基礎 執行計畫（現況對齊版）

> **性質**：原為「App Flutter 技術棧基礎（D29）」目標型執行計畫（含內嵌目標程式碼，10 Tasks 描述 solidart/disco/auto_route/fquery/Sembast 完整落地）。經 2026-09-18 以實際程式碼盤點重建，為**反映現況的執行計畫**。
>
> **對應設計**：`docs/superpowers/specs/2026-08-04-app-flutter-stack-design.md`（D29）、決策 `D5/D29`
> **狀態基準**：2026-09-18 盤點（app 實際 code + pubspec）

---

## 執行狀態總覽

| Task | 內容 | 狀態 | 實際產物 |
|---|---|---|---|
| 1 | App 專案骨架與依賴鎖版 | ✅ 完成 | `app/pubspec.yaml`（依賴鎖版）、`main.dart` |
| 2 | core/config 與 core/errors | ⬜ 未開始 | —（無 `app/lib/core/`） |
| 3 | core/cache — QueryCache factory | ⬜ 未開始 | — |
| 4 | core/cache — Sembast 唯讀快取鏡像 | ⬜ 未開始 | — |
| 5 | core/auth — TokenStore、AuthSession、單飛 refresh | 🟡 部分 | `features/auth/{token_storage,auth_repository,auth_config}.dart` |
| 6 | core/network — AuthInterceptor 與 Transport factory | 🟡 部分 | `features/auth/auth_transport.dart`（connectrpc） |
| 7 | router 骨架（auto_route + guards） | 🟡 部分 | `router/app_router.dart`（auto_route 手寫路由表，無 guard 產碼） |
| 8 | 根佈線（CacheProvider + disco ProviderScope + app.dart） | ⬜ 未開始 | `app.dart` 僅 connectrpc DI，無 fquery/disco 佈線 |
| 9 | fquery 資料層慣例 | ⬜ 未開始 | — |
| 10 | 文件回寫（D29 / v1.0.35 / 主計畫 v2.10.0） | ⬜ 未開始 | — |

**實作範圍**：App 目前完成**骨架、認證（auth feature）、auto_route 路由、connectrpc 客戶端**；**尚未落地 solidart / disco / fquery / Sembast 快取鏡像**（pubspec 已鎖版但 code 未 import，`app.dart` 註解明示「後續掛 CacheProvider(fquery)+根 ProviderScope(disco)」）。

---

## Task 1: App 專案骨架與依賴鎖版

**實際狀態：✅ 完成**

**實際產物（已存在）：**
- `app/pubspec.yaml` — 已鎖版：`auto_route ^11.1.0`、`solidart ^2.8.6`、`disco ^2.0.0`、`fquery ^3.1.0`、`sembast ^3.8.0`、`flutter_secure_storage ^11.0.0`、`flutter_web_auth_2 ^5.0.0`、`connectrpc ^1.0.0`、`crypto`（PKCE）、`firebase_messaging`、`talker`
- `app/lib/main.dart`、`main_dev.dart`、`main_prod.dart`

**說明**：骨架與依賴版本基準已建立（含 flavors）。**注意**：實際 `flutter_web_auth_2`、`sembast ^3.8.0` 與原計畫鎖版（`flutter_web_auth`、`sembast ^3.7.0`）略有出入，以實際 pubspec 為準。

- [x] **Step 1: 建立 App 骨架** — main.dart / flavors
- [x] **Step 2: 加入執行期依賴並鎖版** — pubspec.yaml
- [x] **Step 3: Commit**

## Task 2: core/config 與 core/errors

**實際狀態：⬜ 未開始**

**說明**：`app/lib/` 下**無 `core/` 目錄**；未建立 `core/config`、`core/errors`（實際設定在 `app/lib/config.dart`，未分層 core）。

- [ ] **Step 1: 建立 core/config 與 core/errors**
- [ ] **Step 2: 測試 + Commit**

## Task 3: core/cache — QueryCache factory 與預設策略

**實際狀態：⬜ 未開始**

**說明**：無 QueryCache / fquery factory（server state 管理未落地）。

## Task 4: core/cache — Sembast 唯讀快取鏡像（CacheRepository）

**實際狀態：⬜ 未開始**

**說明**：無 Sembast 唯讀快取鏡像（write-through + 啟動 seed 未實作）。

## Task 5: core/auth — TokenStore、AuthSession、單飛 refresh

**實際狀態：🟡 部分完成**

**實際產物（已存在）：**
- `app/lib/features/auth/token_storage.dart`（`SecureTokenStorage`，flutter_secure_storage）
- `app/lib/features/auth/auth_repository.dart`（登入/refresh 控制）
- `app/lib/features/auth/auth_config.dart`
- `app/lib/features/auth/pages/login_page.dart`、`identity_select_page.dart`（身分選擇登入）

**說明**：token storage、auth repository（refresh）、登入/身分選擇頁已實作（於 `features/auth/` 而非計畫規劃的 `core/auth/`）。**未含** AuthSession 抽象與正式 403→登出/單飛 refresh 控制器（部分於 repository）。是否收斂至 `core/` 待決定。

- [x] **Step 1: token storage** — `SecureTokenStorage`
- [x] **Step 2: auth repository（refresh）** — auth_repository.dart
- [ ] **Step 3: AuthSession 抽象 + 403→登出 + 單飛 refresh 控制器** — 部分合併入 repository，獨立 session 層待

## Task 6: core/network — AuthInterceptor 與 Transport factory

**實際狀態：🟡 部分完成**

**實際產物（已存在）：**
- `app/lib/features/auth/auth_transport.dart` — connectrpc transport（`createAuthServiceClient`）

**說明**：connectrpc client/transport 已建立（於 `features/auth/`）。**未含**正式 AuthInterceptor（Bearer 附加、401→refresh 旋轉重試、403→登出）的完整實作於共用 network 層。

- [x] **Step 1: connectrpc transport** — auth_transport.dart
- [ ] **Step 2: AuthInterceptor（Bearer/refresh 重試/403 登出）** — 待

## Task 7: router 骨架（auto_route + guards）

**實際狀態：🟡 部分完成**

**實際產物（已存在）：**
- `app/lib/router/app_router.dart` — auto_route 手寫路由表（login + identity_select 頁），以建構式注入 `AuthRepository`

**說明**：auto_route 路由骨架已建立（手寫 `Route`/`PageInfo`，未用 generator）。**未含** guard、`route_access.dart`、佔位頁與路由產碼。

- [x] **Step 1: auto_route 路由表** — app_router.dart（login/identity_select）
- [ ] **Step 2: guards + route_access + 佔位頁** — 待
- [ ] **Step 3: 路由產碼（auto_route_generator）** — 目前手寫，未用 generator

## Task 8: 根佈線（CacheProvider + disco ProviderScope + app.dart/main.dart）

**實際狀態：⬜ 未開始**

**說明**：`app.dart` 目前僅 connectrpc DI + `MaterialApp.router`；註解明示「後續在此掛 CacheProvider(fquery)+根 ProviderScope(disco)」。fquery QueryCache、disco ProviderScope 皆未落地。

- [ ] **Step 1: 根佈線（CacheProvider + ProviderScope）**
- [ ] **Step 2: 測試 + Commit**

## Task 9: fquery 資料層慣例（QueryKeys + seed + 範例 repository）

**實際狀態：⬜ 未開始**

**說明**：無 fquery QueryKeys / seed / repository 慣例（無下單等資料層）。

## Task 10: 文件回寫（D29 / 規格書 v1.0.35 / 主計畫 v2.10.0）

**實際狀態：⬜ 未開始**

**說明**：D29 技術棧文件回寫（規格書升版、主計畫 v2.10.0）未執行。

---

*最後更新：2026-09-18（app-flutter-stack 現況對齊重建）*
