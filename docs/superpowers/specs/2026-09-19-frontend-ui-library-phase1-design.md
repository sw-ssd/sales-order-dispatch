# 前端 UI 元件庫 Phase 1 — 設計（spec v3）

> **狀態**：設計已定案（2026-09-19），待實作計畫
> **範圍**：`frontend/`（SolidJS Web 中台）
> **前置**：v2（Ark × Tailkit）已完成並通過全分支複審（`7d1ac4b..3167127`）
> **本 spec 只涵蓋 Phase 1**；Phase 2–4 各自再走一次相同的流程

---

## 0. 背景：四階段拆分（2026-09-19 使用者 directive）

使用者新增方向：form 用 TanStack Form、table 用 TanStack Table、drag and drop 用 Pragmatic D&D、sidebar 參考 solid-ui、靜態元件 Ark 有就用 Ark、最終抽出自有元件庫。經訪談拆分為四階段，**有消費者才做**：

| 階段 | 內容 | 消費者現況 |
|---|---|---|
| **Phase 1（本 spec）** | 元件庫庫化 + Ark 靜態元件（field/label/pagination/scroll-area）+ sidebar 完整重做 | 有（現有 5 頁與 shell） |
| Phase 2 | TanStack Form 重寫表單 | 有（登入、公司、部門） |
| Phase 3 | TanStack Table 重寫表格 | 有（公司、部門、角色） |
| Phase 4 | Pragmatic drag and drop | **無**（派車看板尚未建）→ 等畫面落地 |

已查證的套件座標（Phase 2/3/4 用，本階段不安裝）：`@tanstack/solid-form@1.33.5`、`@tanstack/solid-table@9.2.4`、`@atlaskit/pragmatic-drag-and-drop@3.1.0`（+ `-hitbox@2.2.2`，社群有 SolidJS 移植可參考）。

## 1. 目標與範圍

**In**
- 元件庫庫化：barrel export、`registry.json`、每元件文件、dev-only demo 路由
- 四個靜態元件導入 Ark primitives（Ark 在內、對外 API 盡量維持）
- sidebar 依 solid-ui 的 parts 解剖完整重做（Ark 拼裝），並修掉 v2 遺留的兩項 a11y 問題
- 既有 UI 元件與頁面在新 API 下維持全綠
- 刪除因本階段改動而成為死碼的檔案：`ui/label.tsx`（唯一消費者 `field.tsx` 改用 Ark `Field.Label`）與 `hooks/create-pagination.ts`（頁碼改由 Ark 產生後僅剩切片職責，若確認無其他消費者即刪）

**Out**
- TanStack Form / Table（Phase 2/3）
- Pragmatic drag and drop（Phase 4）
- 後端、proto、Flutter `app/`
- 主題切換器（`.dark` 在執行期仍不可達，屬另一條待辦）
- 已登記的 0 消費者回歸（`table.tsx` 的 `caption-bottom`、`pagination` 的 `hideText`/`PaginationSummary` 字型）→ 若與本階段同檔可順手修，否則維持 deferred

## 2. 元件庫形狀（原位置庫化）

住在 `frontend/src/components/ui/`，不抽 workspace package（目前只有一個消費者；抽包要動 pnpm workspace、tsconfig paths、Tailwind content 與建置流程，成本大於收益）。

### 2.1 檔案規則

| 情況 | 規則 |
|---|---|
| 單部件元件（button/card/badge/input/label/field/table/pagination/spinner/dialog/tabs/checkbox/scroll-area） | 維持 `ui/<name>.tsx` + `ui/<name>.test.tsx` 單檔（不為了 churn 而搬家） |
| 多部件元件（**sidebar**） | 開資料夾 `ui/sidebar/`：`index.tsx`（公開 API）、`context.tsx`（Provider/state）、`parts.tsx`（解剖部件）、`sidebar.test.tsx` |

### 2.2 Barrel

新增 `ui/index.ts` 匯出所有公開元件與型別。頁面與 layout 改從 `~/components/ui` 匯入；**不再逐檔深引**（`~/components/ui/dialog` 這種）。

### 2.3 `registry.json`（機器可讀清單）

每個元件一筆，供人與 agent 查詢（未來要抽包或做 CLI 時直接可用）：

```json
{
  "name": "sidebar",
  "path": "src/components/ui/sidebar",
  "kind": "composite",
  "ark": ["collapsible", "drawer", "splitter", "tooltip", "hotkeys"],
  "deps": ["lucide-solid", "~/lib/cn"],
  "tokens": ["--sidebar-background", "--sidebar-foreground", "--sidebar-primary", "--sidebar-accent", "--sidebar-border", "--sidebar-ring"],
  "variants": ["side:left|right", "variant:sidebar|floating|inset", "collapsible:offcanvas|icon|none"],
  "state": "stable",
  "docs": "src/components/ui/sidebar/README.md"
}
```

### 2.4 每元件文件

`ui/<name>/README.md`（多部件者）或 `ui/<name>.md`（單檔者）：用途、API 表、Ark 對應、可及性要點、範例。**不寫**與程式重複的長篇說明。

### 2.5 dev-only demo 路由 `/ui`

- 以 `import.meta.env.DEV` 閘門註冊，且該路由的元件模組**以動態 `import()` 載入**，確保正式 build 時整棵 demo 樹被 tree-shake 掉。
- 展示每個元件的所有變體與狀態（含深色預覽切換）。
- **不**出現在 sidebar 導覽；**不**進 `registry.json` 的公開清單。

### 2.6 依賴規則（庫的邊界）

`ui/**` 不得 import `features/**`、`router/**`、`lib/proto/**`。只允許依賴：`@ark-ui/solid`、`lucide-solid`、`~/lib/cn`、`class-variance-authority`、SolidJS 本身。這條規則讓「未來抽成 package」是純機械工作。

## 3. Ark 靜態元件（4 個）

原則：**Ark 在內、對外 API 盡量維持**（與 v2 對互動元件的做法一致）。

### 3.1 `field.tsx`（+ `label.tsx`）

Ark 事實（已查證）：`Field.Root(invalid/required/disabled/readOnly/target)`、`Field.Label`、`Field.Input`、`Field.ErrorText`、`Field.HelperText`、`Field.RequiredIndicator`。

- `Field` 內含 `Field.Root`；`invalid` 由包裝層推導（呼叫端傳 `invalid` 或偵測有 error 子節點）。
- `FieldLabel` → `Field.Label`；`FieldError` → `Field.ErrorText`；`FieldDescription` → `Field.HelperText`。
- 對外 props（`class`/`children`/`for`）不變；`aria-describedby`/`aria-invalid` 的正確性由 Ark 提供（目前是手寫）。

### 3.2 `scroll-area.tsx`

Ark 事實：`Root > Viewport > Content` + `Scrollbar(orientation)` + `Thumb` + `Corner`。

- 包裝層維持單一入口 `<ScrollArea class>{children}</ScrollArea>`，內部展開 Ark 解剖並附兩向 scrollbar。
- 呼叫端（`dialog.tsx` 的 `max-h-[80vh]`）**不需修改**。

### 3.3 `pagination.tsx`（**唯一對外 API 變更**）

Ark 事實：`Pagination.Root(count/page/defaultPage/pageSize/type/getPageUrl/onPageChange/siblingCount/boundaryCount/translations)` 自己持有頁碼狀態並以 `api().pages` 產生頁碼；部件為 `Item(type="page", value)`、`Ellipsis(index)`、`Prev/Next/First/LastTrigger`。

- 包裝層改收 **`count` / `page` / `pageSize` / `onPageChange`**，頁碼由 Ark 產生；`Item/Ellipsis/Trigger` 保留為樣式部件。
- `features/users/components/ListPagination.tsx` 跟著改（傳 `count`/`page`/`onPageChange`）。
- `hooks/create-pagination.ts` 只保留「資料切片」職責，不再算頁碼清單；若切片改由 Ark 的 `page`/`pageSize` 驅動後該 hook 成為死碼 → **刪除**並在報告說明。
- 其餘頁面不受影響（只有 `ListPagination` 在用）。

### 3.4 `label.tsx` → **刪除**

已查證：`label.tsx` 的唯一消費者是 `field.tsx`（`FieldLabel` 以 `LabelProps` 為基底）。`field.tsx` 改內含 Ark `Field.Label` 之後，`label.tsx` 即無消費者 → **刪除該檔**（乾淨切換，不留死元件），`FieldLabelProps` 改為自行定義（`class`/`children`/`for`）。`FieldLabel` 對外用法不變（頁面傳 `for="…"` 與 `<Input id="…">` 的關聯由測試守住）。

## 4. Sidebar 完整重做（solid-ui 解剖 + Ark）

### 4.1 部件（公開 API）

`SidebarProvider`(狀態 + 記憶) / `Sidebar` / `SidebarHeader` / `SidebarFooter` / `SidebarContent` / `SidebarGroup` / `SidebarGroupLabel` / `SidebarGroupContent` / `SidebarMenu` / `SidebarMenuItem` / `SidebarMenuButton`（polymorphic `as`，接 TanStack `Link`）/ `SidebarMenuSub` / `SidebarTrigger` / `SidebarRail`（拖曳調寬）/ `SidebarInset`。

### 4.2 Ark 對應（**不自行發明狀態機**）

| 行為 | Ark primitive |
|---|---|
| 收合 / 展開（桌面 icon-rail、分組展開） | `collapsible` |
| 手機 off-canvas | `drawer`（overlay、Esc 關、focus 管理由 Ark 提供） |
| 側欄拖曳調寬 + 鍵盤調寬 | `splitter`（`Panel` + `ResizeTrigger` + `keyboardResizeBy`） |
| 收合時顯示標籤 | `tooltip` |
| `Cmd/Ctrl+B` 切換 | `hotkeys` |

### 4.3 行為清單

- 桌面：`collapsible="icon"` 收合成 icon rail；狀態存 **localStorage**（key `ui:sidebar`；solid-ui 用 cookie 是為了 SSR，我們是 SPA）。
- 行動（<1024px）：off-canvas drawer；**關閉時必須 `visibility: hidden`（`invisible`）移出 tab 順序**（修 v2 遺留問題 M-4）。
- 導覽選中態：接 router 當前路徑；**同一個 `<nav>` 內只允許一個 `aria-current="page"`**（修 v2 遺留問題 M-2；品牌列不再是 `Link to="/"`，或移出 `<nav>`）。
- 變體：`variant="sidebar" | "floating" | "inset"`、`side="left" | "right"`、`collapsible="offcanvas" | "icon" | "none"`（完整照 solid-ui）。
- sub-menu：`SidebarMenuSub` 以 `collapsible` 驅動；目前無實際子項，於 demo 頁展示。

### 4.4 與 shell 的關係

`components/layout/AppShell.tsx` 改以 `SidebarProvider` + `Sidebar` + `SidebarInset` 組裝；`splitter` 用於側欄與內容之間（取代固定 `lg:pl-64`）。內距歸屬維持 v2 的結論（**shell 是唯一內距所有者**），並一併處理 v2 遺留的 `HomePage` 自帶 `p-8` 疊加問題。

## 5. 測試與驗收

**測試（jsdom + testing-library）**
- sidebar：收合切換與 localStorage 記憶、drawer 開/關（Esc）、關閉時導覽連結不在 tab 順序（`invisible` → `queryByRole` 取不到）、`aria-current` 唯一
- pagination：`onPageChange` 回報、`count` 邊界（`boundaryCount`/`siblingCount`）、disabled 首/末頁
- field：錯誤時 `aria-invalid`/`aria-describedby` 關聯到 `FieldError`
- scroll-area：內容可捲（scrollHeight/clientHeight）
- 禁則沿用 v2：不攔截 rAF、不 mock Ark 內部、不斷言 Ark 私有屬性

**驗收（可執行）**
1. `pnpm typecheck && pnpm lint && pnpm test && pnpm build` 全綠，**lint 0 warning**
2. `/ui` demo 頁以 Microsoft Edge 實測：每個元件所有變體渲染、深色切換（`document.documentElement.classList.add('dark')`）
3. shell 實測：桌面收合/展開 + 記憶（重整後保持）、拖曳調寬、`Cmd/Ctrl+B`、行動 drawer（Esc、focus、關閉後不可 Tab）、選中態
4. 既有 5 頁在新 API 下正常（公司/部門分頁、表單錯誤、dialog 捲動）
5. `registry.json` 與每個元件的 README 齊備；`ui/**` 依賴規則以 `grep` 驗證（無 `features/`、`router/`、`lib/proto` 匯入）
6. `impeccable detect` 掃改動面，有發現即修或說明

## 6. 風險

- **R1**：Pagination 換 Ark 是本階段唯一呼叫端變更（`ListPagination`、可能刪 `create-pagination`）；Ark 的頁碼產生邏輯與我們現行的 `siblingCount`/`boundaryCount` 行為需逐一比對（首/末頁、ellipsis 位置）。
- **R2**：`splitter` 會換掉 shell 的 DOM 結構與固定左內距 → `AppShell` 等於重寫，行動版與桌面版都要重驗。
- **R3**：`field` 的 Ark 錯誤模型（`Field.ErrorText` 需在 `Root` 內且由 `invalid` 驅動）與我們 valibot 手動錯誤的接線若不對，`aria-describedby` 會靜默斷掉 → 以測試守住。
- **R4**：sub-menu / floating / inset / right-side 變體目前**沒有消費者**（你已選完整版）→ 會實作並在 demo 頁展示，但沒有實戰驗證；未來接上時需複驗。
- **R5**：`registry.json` 與 README 會隨程式漂移 → 由 `impeccable detect`／複審檢查，不引入自動產生器（YAGNI）。

## 7. 後續階段（各自再走一次流程）

- **Phase 2**：TanStack Form（`@tanstack/solid-form`）重寫登入、公司、部門表單；與 Ark `Field` 的分工（TanStack 管狀態/驗證、Ark 管 a11y 關聯）需先定清楚。
- **Phase 3**：TanStack Table（`@tanstack/solid-table` v9）重寫三張表；`Table` 元件的 markup 保留、排序/篩選/分頁狀態交給 TanStack。
- **Phase 4**：Pragmatic drag and drop 接入派車看板（待該畫面落地）。

---

*建立：2026-09-19*
