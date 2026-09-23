# Frontend（SolidJS）Development Guidelines

## 1. Strict SolidJS Reactivity Rules

1. **NEVER Destructure Props Directly**:
   - `const { variant, class: className } = props;` -> ❌ **FORBIDDEN** (breaks SolidJS fine-grained signal tracking).
   - `const [local, others] = splitProps(props, ["variant", "class"]);` -> ✅ **REQUIRED**.

2. **Children Inspection & Tab Hydration**:
   - ALWAYS wrap `props.children` with SolidJS's native `children(() => props.children)` memoization helper when inspecting, iterating, or rendering dynamic JSX child nodes.

3. **SSR Safety Guards**:
   - Always include `typeof window !== "undefined"` and `typeof document !== "undefined"` guards inside browser event listeners or DOM access logic.

## 2. UI 元件庫：Ark UI（行為）× Tailkit（視覺）

- 互動元件一律以 **Ark UI primitives**（`@ark-ui/solid`）實作行為（狀態機、焦點管理、鍵盤互動、ARIA），放在 `src/components/ui/`。
- **Tailkit 只作為結構與樣式的參考來源，不是 npm 依賴**：取其版面／間距／層級思路，改寫成 SolidJS 元件，不得把 Tailkit 產物原樣入庫、不得寫進 `package.json`。
- 狀態一律以 Ark 的 `data-state` / `data-selected` / `data-orientation` 表達，不要自寫 `aria-*` 或 `role`。
- 元件對外 API 照 Ark 語意，且不外洩 `details` 物件：回呼吃純值（`onOpenChange(open: boolean)`、`onValueChange(value: string)`、`onCheckedChange(checked: boolean)`）。

## 3. Tailwind CSS v4 Native Design Tokens

- All components must use semantic design tokens (`bg-background`, `text-foreground`, `bg-card`, `border-border`, `bg-primary`, etc.).
- 不得使用色階字面值（`secondary-*`、`orange-*`、`emerald-*`、`zinc-*` 等）；token 已在 `src/index.css` 定義並於深色模式翻轉，因此也不再寫顏色相關的 `dark:` 變體。

## 4. Code Search / Index via codebase-memory-mcp

All code indexing and querying in this project **must go through codebase-memory-mcp first** (see root `docs/AGENTS.md` §4.0): locate definitions/implementations/callers, trace call paths, analyze blast radius, and traverse across projects via the codebase-memory knowledge graph; fall back to grep/direct file reads only when insufficient. Confirm index coverage of a target file before modifying existing code.

## 5. 測試：斷言必須「能失敗」

本專案前端測試的用途是**契約守門**,不是覆蓋率裝飾。寫測試時逐條自問「**什麼改動會讓它紅?**」——答不出來就是假綠。已在本專案抓到的假綠形狀(勿再犯):

1. **以「某文案不存在」代表狀態** → 文案一改就失去鑑別力。改為**結構判定**(`tbody` 列數、`queryAllByRole("row")`)或斷言**應該出現**的內容。
2. **期望值由被測畫面自己產生**(`expect(x).toHaveLength(headers.length)`、`colspan == headers.length`) → 兩邊同源,整欄被刪也全綠。改為**字面值**(`toHaveLength(6)`、`colspan === "6"`)。
3. **原生元素的預設值當契約**(`button.tabIndex === 0`) → jsdom 預設即 0。刪掉,或改斷言真正的行為(`tagName`、`document.activeElement`、點擊後焦點仍在)。
4. **斷言「不存在的屬性」**(`tr.getAttribute("role") === null`) → 任何非互動元素都成立。改為斷言緊接的**行為**(點列不會選取、不會發請求)。
5. **函式形式的選項只看單點**(`retry(0)=true`、`retry(3)=false` 卻號稱「封頂 3 次」) → 上限改成 1/2 也會過。補齊中間值。
6. **突變驗證**:新增或強化的斷言,交付前至少手動把對應實作改壞一次確認它會紅(並在報告寫下該實驗);**不能只靠「測試全綠」當證據**。

## 6. SolidJS 風格指南（solidjs-patterns）

權威來源是 `.omp/skills/solidjs-patterns/`：`SKILL.md` 是 Quick Reference（依 Impact 排序的 9 類目錄），`AGENTS.md` 是完整編譯版，`rules/*.md` 單條規則一份一檔（檔頭有 `impact` 與 `tags`）。**照 SKILL.md 的 `Reference Loading` 走：只載入本次碰到的主題，不要整份讀**（整份 50+ 條會淹掉真正的重點）。

本節同 §1 起的五節：只記**覆寫**、**已落地的修正**、**未收斂**，不重述指南。

### 6.1 本專案覆寫指南（刻意偏離，不要「修復」）

1. **表單陣列不用 `createStore`**（`state-form-store`／`state-store`）：`OrdersPage.tsx:325`、`ProductsPage.tsx:270` 的草稿列用 `createSignal`。
   為什麼：這些是「送出前整批草稿」，語意就是**整批取代**（`form.reset` 與 `setItems([emptyItem()])` 同一個 `batch`）；逐欄 store 沒有收益，卻會動到 22 個既有測試的斷言路徑。
2. **signal 內放 `ReadonlySet` 是刻意的**（`reactivity-no-set-map-signal`，`NotificationsPage.tsx:131`）：該集合每次整批換（`markRead` 一次 RPC 帶整批 id），只被 `.has(id)` 讀一個 bool，沒有 per-item 訂閱需求。
3. **`createEffect` 裡讀 `.data` 是安全的、不必改成 derive**（`reactivity-derive-not-effect`）：`CustomersPage.tsx:267`、`ProductsPage.tsx:263`、`PrintPage.tsx:217`、`AuditPage.tsx:240`、`UsersPage.tsx:220`、`DepartmentsPage.tsx:323` 的 `if (query.isPlaceholderData || !query.data) return;` 都在 effect 內。
   為什麼安全：Solid resource 的 `read()` 只在 `Listener && !Listener.user` 時註冊 suspense（`solid.js` 的 `createResource`），`createEffect` 是 user-level computation（`user: true`）→ 既不拋 promise 也不 `c.increment()`，不會把路由那層 Suspense 的 fallback 頓出來。**會出事的是 JSX 讀**（render-phase，`user: false`），見 6.2-1。
   為什麼是 effect 不是 derive：它跨了 external boundary（query 狀態 → 本地 UI 狀態「把超界頁碼夾回來」），本身就是邊界同步。
4. **沒註冊資源的 effect 不用補 `onCleanup`**（`reactivity-cleanup-effects`）：`AppShell.tsx:70` 只呼叫 `setOpenMobile`，沒有 listener／timer／subscription。四個真正持有資源的 effect（`theme.tsx:92`、`sidebar/context.tsx:104`、`DispatchPage.tsx:270`…）都已有 `onCleanup`。
5. **非 SolidStart、非 RxJS**：`start-createasync-not-resource`／`start-route-preloading`／`start-use-server-validation`／`interop-from-browser-apis`／`interop-observable-export`／`data-resource-latest` 一律不適用 —— 本專案是 TanStack Router 程式化路由樹 + `@tanstack/solid-query`，沒有 `createResource`／`use server`／Observable。
6. **資料列表繼續用 `<For>`**（`rendering-use-for-not-map` 已符合，JSX 內 0 處 `.map()`）；只有 6.2-2 那兩處草稿列改 `<Index>`。

### 6.2 指南在此落地的硬規則（2026-09-23 全面查核，已修正）

1. **`data-guard-suspense` 是本專案踩到的最嚴重一條，全站 97 處改完。**
   機制：`@tanstack/solid-router` 在 client 把整條路由包在一個 `Solid.Suspense`（`Matches.js:15`），fallback 取 `router.options.defaultPendingComponent` —— 本專案原本**沒設**，所以 fallback 是 `null`。而 `createQuery().data` 走的是 `useBaseQuery` 的 Proxy：該 key **從未有過值**時落到 `resource()`，在 render-phase 讀它就 `c.increment()`，整個路由子樹被替換成 `null`。
   真實重現（真瀏覽器 + 真 Postgres，`CustomerService/ListCustomers` 延遲 4 秒、client-side 導覽到 `/customers`）：`document.body.innerText.length` **265 → 109**、持續整個 4 秒；側邊欄還在（它在 `<Outlet/>` 之外），內容區是一面空牆。
   兩層修法：
   - **讀點**：一律走 `src/lib/query-data.ts` 的 `queryData(query, (d) => …)` —— pending（含未啟用）時把 `undefined` 交給取值函式，不碰 `.data`。全站 97 處（`grep -ro "queryData("`），**已無 JSX 直接讀 `.data`**。
   - **兜底**：`router/index.tsx` 設 `defaultPendingComponent: RouteSkeleton`（`src/components/layout/RouteSkeleton.tsx`，形狀照「標題列＋篩選卡＋表格卡」，避免版面跳動）。有了它，即使有人漏用 helper 也是骨架而不是空白。
   為什麼兩層都要：helper 是正解但靠人守；骨架是機械兜底，漏一次不至於白屏。
2. **草稿列用 `<Index>` 不用 `<For>`**（`rendering-index-vs-for`）：`OrdersPage.tsx:786`、`ProductsPage.tsx:701`。
   為什麼：`patchRow` 每次以**新物件**取代該列 → `mapArray` 視為新參照而整列 remount，輸入框節點被丟掉、焦點掉回 `document.body`（agent 的 vitest 探針實測：`<For>` 時 `document.contains()` 變 false、`activeElement` 為 BODY；改 `<Index>` 後 `same-node=true`、焦點仍在該輸入框）。`<Index>` 的項目是 accessor，節點按 index 對應，就不會重建。
   為什麼只改這兩處：其餘 `<For>` 都是**資料列**（後端已排序、物件有身分），remount 正是預期行為。
3. **props 不得解構**（`reactivity-no-destructure-props`／`reactivity-no-rest-spread`）：全站 **0 處違規**，一律 `splitProps`。這一條既有習慣本來就對，是 §1-1 的功勞。

### 6.3 已知未收斂

1. **20 條路由全是靜態 import，未做 code-splitting**（`perf-lazy-load-heavy-components`／`perf-route-code-splitting`，`router/index.tsx:11-34`；只有 `/ui` demo 用 `lazyRouteComponent`）。
   為什麼沒做：改 `lazy` 會變更 build 切塊、動到 `declare module "@tanstack/solid-router"` 的 `Register` 型別註冊，以及所有直接 import 頁面的測試 —— 屬 router 契約決策，不是順手能改的。`vite build` 已在警告 chunk > 500 kB。
2. **`src/lib/ability/guards.test.ts:5-19` 的 `mockGetAbility` 沒用 `vi.hoisted`**（`testing-mock-hoisted`）。目前不會炸：該 mock 在 `vi.mock` factory 內只被包進 thunk、延後到呼叫才取值，避開了 TDZ。但這是脆弱寫法，應改 `vi.hoisted`。

