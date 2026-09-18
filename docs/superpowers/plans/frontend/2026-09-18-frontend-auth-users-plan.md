# Frontend（SolidJS Web 中台）執行計畫（現況對齊版）

> **性質**：新增計畫（原無 frontend 專屬計畫檔；Web 端工作散見於主計畫）。經 2026-09-18 以實際程式碼盤點建立，為**反映現況的執行計畫**。
>
> **對應設計**：`docs/superpowers/specs/2026-07-16-sales-order-1.0-design.md`（v1.0.34）、決策 `D4/D5/D9`
> **狀態基準**：2026-09-18 盤點（frontend 實際 code + package.json）
> **前置依賴**：backend proto（`*_pb` GenService）、backend 01/02（auth / role / ability API）

---

## 實際技術棧（package.json）

- SolidJS 1.9 + Vite 8 + TypeScript（native TS7 / TS6）
- `@connectrpc/connect-web`（Connect 客戶端）、`@bufbuild/protobuf`
- `@casl/ability`（UI 權限，D9/D30）
- `@tanstack/solid-query`（server state）、`@tanstack/solid-router`
- UI：`@kobalte/core`、`class-variance-authority`、`tailwind-merge`、`clsx`、`lucide-solid`
- 驗證：`valibot`；測試：Vitest + jsdom

---

## 執行狀態總覽

| 領域 | 內容 | 狀態 | 實際產物 |
|---|---|---|---|
| UI 基礎 | components/ui + hooks | 🟡 部分 | `components/ui/*`（14 元件）、`hooks/*`（4 hooks） |
| auth | 登入 / 403 / Google 登入 | ✅ 完成 | `features/auth/{LoginPage,ForbiddenPage,GoogleLoginButton}.tsx` |
| users | 公司/部門/角色權限 | 🟡 部分 | `features/users/pages/{CompaniesPage,DepartmentsPage,RolesPage}`、`PermissionMatrix`、`ListPagination` |
| ability | CASL ability 整合 | ✅ 完成 | `lib/ability/{service,context,guards,Can}.tsx(+test)` |
| network | Connect transport + query client | ✅ 完成 | `lib/transport.ts`、`lib/query-client.ts`、`lib/proto/*` |
| router | 檔案式路由 + 守衛 | ✅ 完成 | `src/router/index.tsx`、`App.tsx` |

---

## 分域詳情

### 1. auth（登入 / 403 / Google OIDC 串接）

**實際狀態：✅ 完成**

**實際產物（已存在）：**
- `frontend/src/features/auth/pages/LoginPage.tsx`
- `frontend/src/features/auth/pages/ForbiddenPage.tsx`（403 無權限頁，git `a356888`）
- `frontend/src/features/auth/components/GoogleLoginButton.tsx`（OIDC PKCE 登入，git `bcf7b15`）
- `frontend/src/lib/transport.ts`（`createConnectTransport({ baseUrl: "/api/v1" })`）

- [x] **LoginPage** — 登入頁
- [x] **ForbiddenPage** — 403 頁（deny 落點）
- [x] **GoogleLoginButton** — Google OIDC（PKCE）

### 2. ability（CASL UI 權限，D9/D30）

**實際狀態：✅ 完成**

**實際產物（已存在）：**
- `frontend/src/lib/ability/service.ts(+test)` — ability service + context
- `frontend/src/lib/ability/context.tsx` — React context
- `frontend/src/lib/ability/guards.ts(+test)` — `requireAbility` 路由守衛
- `frontend/src/lib/ability/Can.tsx(+test)` — `<Can>` 顯示控制
- git commit（ability service、Can、requireAbility、快取失效 `ef1bf04`）

- [x] **ability service / context**
- [x] **`<Can>` 顯示控制**
- [x] **`requireAbility` 路由守衛**
- [x] **ability 快取失效**

**註**：依 D32 決策，CASL 日後將改由 OpenFGA `Check`/`list-objects` 驅動；目前實作仍為 CASL（以實際 code 為準），OpenFGA 遷移列為待辦。

### 3. users（公司 / 部門 / 角色權限管理）

**實際狀態：🟡 部分完成**

**實際產物（已存在）：**
- `frontend/src/features/users/pages/CompaniesPage.tsx`
- `frontend/src/features/users/pages/DepartmentsPage.tsx`
- `frontend/src/features/users/pages/RolesPage.tsx`（角色權限設置頁 + PermissionMatrix，git `0c51977`）
- `frontend/src/features/users/components/PermissionMatrix.tsx`、`ListPagination.tsx`

**說明**：公司、部門、角色權限設置頁已實作並接上後端 Company/Department/RoleService。**未含**完整使用者管理頁（UserService 後端未實作，見 backend 02 Task 3）。

- [x] **CompaniesPage** — 公司 CRUD 頁
- [x] **DepartmentsPage** — 部門 CRUD 頁
- [x] **RolesPage + PermissionMatrix** — 角色權限設置（T19）
- [x] **ListPagination** — 分頁元件
- [ ] **使用者管理頁** — 依後端 UserService（backend 02 Task 3）落地後補

### 4. UI 基礎 + hooks + network（地基）

**實際狀態：🟡 部分完成**

**實際產物（已存在）：**
- `components/ui/*`：badge/button/card/checkbox/dialog/field/input/label/pagination/scroll-area/spinner/table/tabs
- `hooks/*`：create-controllable-signal、create-pagination、create-resize-observer、create-scroll-position
- `lib/query-client.ts`（TanStack Query）、`lib/proto/*`（Connect generated）

**說明**：UI 元件庫、hooks、Connect/proto 客戶端已建立；**未含**後續業務頁（master-data / orders / dispatch / printing 等，待各 backend domain 落地）。

- [x] **components/ui + hooks**
- [x] **Connect transport + query client + proto**
- [ ] **master-data / orders / dispatch / printing 前端頁面** — 待各 backend domain（04/05/08/09）落地

### 5. router + 守衛

**實際狀態：✅ 完成**

**實際產物（已存在）：**
- `src/router/index.tsx`（檔案式路由，TanStack Router）
- `src/App.tsx`、`src/main.tsx`

- [x] **檔案式路由** — router/index.tsx
- [x] **受保護路由接上 requireAbility** — git `77b00b4`、`2b89eb9`

---

## 待辦摘要（相依 backend）

| 待辦 | 相依 |
|---|---|
| 使用者管理頁（UserService） | backend 02 Task 3 |
| master-data 前端（客戶/商品/倉別…） | backend 04 |
| orders 前端 | backend 05 |
| dispatch 看板前端 | backend 08 |
| printing 前端 | backend 09 |
| CASL → OpenFGA 遷移（UI 權限來源） | D32 定奪 |

---

*最後更新：2026-09-18（frontend 現況對齊計畫，新增）*
