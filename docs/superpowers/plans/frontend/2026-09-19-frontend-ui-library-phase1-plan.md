# 前端 UI 元件庫 Phase 1 實作計畫

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `frontend/src/components/ui/` 變成有清單、有文件、有 demo 頁的自家元件庫；四個靜態元件改用 Ark primitives（Ark 在內、對外 API 盡量維持）；sidebar 依 solid-ui 解剖用 Ark 完整重做。

**Architecture:** Ark 提供行為與 a11y 關聯，Tailkit 結構＋語意 token 提供視覺，本地元件是唯一入口。Sidebar 以 CSS 變數（`--sidebar-width*`）與 `data-state/collapsible/variant/side` 驅動版面，行為交給 Ark（`collapsible`/`drawer`/`splitter`/`tooltip`/`hotkeys`）。

**Tech Stack:** SolidJS 1.9、`@ark-ui/solid` 5.39.2、Tailwind v4.3、TanStack Router/Query、Vitest + jsdom + testing-library、pnpm workspace。

**Spec:** `docs/superpowers/specs/2026-09-19-frontend-ui-library-phase1-design.md`

## Global Constraints

- 套件管理器 pnpm；指令在 `frontend/` 執行；lockfile 在 repo 根。
- **不新增依賴**：本階段只用已安裝的 `@ark-ui/solid`、`lucide-solid`、`class-variance-authority`、Tailwind v4。Phase 2–4 的 TanStack Form/Table、Pragmatic D&D **不在本階段安裝**。
- **`ui/**` 依賴規則（本階段新規則）**：不得 import `features/**`、`router/**`、`lib/proto/**`；只允許 `@ark-ui/solid`、`lucide-solid`、`~/lib/cn`、`class-variance-authority`、solid-js。
- 顏色一律語意 token（含 `--sidebar-*`）；無顏色相關 `dark:` 變體；無 Tailkit 色階字面值。
- **對外 API**：除 `Pagination`（本階段唯一變更）外，其餘元件與頁面 API 不變。
- **測試禁則**（沿用 v2）：不攔截 rAF、不 mock Ark 內部、不斷言 Ark 私有屬性、不留只斷言「沒被呼叫」的假綠；非同步用真實等待。
- `pnpm lint` 必須 **0 warning**；四道 gate（typecheck/lint/test/build）全綠。
- **瀏覽器**：視覺驗證用 Microsoft Edge（`--headless=new` + CDP）；**禁用** Chrome for Testing／`chrome-headless-shell`。收尾強制：`pkill -f "Microsoft Edge.*--headless"`、確認 0 殘留、`rm -rf /tmp/msedge-*`、不留暫存檔在 repo。
- **任何新增、刪除、改名元件的 task，必須同步更新 `ui/index.ts`（barrel）與 `ui/registry.json`**；Task 7 做最終對齊檢查。（已知會觸發的：Task 2 刪 `label.tsx`、Task 4 改 `pagination` 型別、Task 5 新增 `sidebar`。）
- 註解、文件、commit message 一律繁體中文。
- Harness：shell `grep`/`rg`/`ls`/`find` 被封鎖（用內建工具）；不要 `git add -A`。

### 已查證的 Ark 5.39.2 事實（不要憑記憶）

| primitive | parts / props 重點 |
|---|---|
| `field` | `Root(invalid/required/disabled/readOnly/target)`、`Label`、`Input`、`Textarea`、`Select`、`ErrorText`、`HelperText`、`RequiredIndicator`、`Item` |
| `scroll-area` | `Root`、`Viewport`、`Content`、`Scrollbar(orientation)`、`Thumb`、`Corner` |
| `pagination` | `Root(count/page/defaultPage/pageSize/type/getPageUrl/onPageChange/onPageSizeChange/siblingCount/boundaryCount/translations)`、`Item(type="page", value)`、`Ellipsis(index)`、`Prev/Next/First/LastTrigger` |
| `splitter` | `Root(panels/size/defaultSize/orientation/onResize/keyboardResizeBy)`、`Panel(id)`、`ResizeTrigger(id/disabled)`、`ResizeTriggerIndicator`、`Registry` |
| `collapsible` | `Root(open/defaultOpen/onOpenChange/disabled/lazyMount/unmountOnExit)`、`Trigger`、`Content`、`Indicator` |
| `drawer` | `Root(open/defaultOpen/onOpenChange/closeOnEscape/closeOnInteractOutside/modal/trapFocus/restoreFocus/preventScroll/snapPoints/swipeDirection)`、`Trigger`、`Backdrop`、`Positioner`、`Content`、`Title`、`Description`、`CloseTrigger`、`Grabber`、`SwipeArea` |
| `hotkeys` | hook 形式：`useHotkeys({commands})`；command = `{hotkey, action, enabled?, options:{preventDefault?,enableOnFormTags?}}` |
| `tooltip` | `Root`/`Trigger`/`Positioner`/`Content`/`Arrow` |

**取碼工具**（kernel 代理，dispatch 時以 `tools:` 掛給 agent）：`ark_props(component)`、`ark_examples(component)`、`ark_example(component, example_id)`、`ark_styling(component)`。

### solid-ui sidebar 的既有事實（我們照它的解剖，但換成 Ark）

`MOBILE_BREAKPOINT=768`、`SIDEBAR_WIDTH=16rem`、`SIDEBAR_WIDTH_MOBILE=18rem`、`SIDEBAR_WIDTH_ICON=3rem`、快捷鍵 `b` + meta/ctrl、狀態持久化（它用 cookie `sidebar:state`；我們用 localStorage `ui:sidebar`）。context 形狀：`{state, open, setOpen, isMobile, openMobile, setOpenMobile, toggleSidebar}`。版面靠 `data-state|data-collapsible|data-variant|data-side` 與 `--sidebar-width*` CSS 變數。

---

## 檔案結構

| 檔案 | 動作 | 責任 |
|---|---|---|
| `frontend/src/components/ui/index.ts` | 新增 | barrel：所有公開元件與型別 |
| `frontend/src/components/ui/registry.json` | 新增 | 機器可讀元件清單（見 spec §2.3 schema） |
| `frontend/src/components/ui/*.md`、`ui/sidebar/README.md` | 新增 | 每元件文件 |
| `frontend/src/components/ui/demo/UiDemoPage.tsx`（或 `features/` 外） | 新增 | dev-only `/ui` 展示頁 |
| `frontend/src/router/index.tsx` | 改 | 僅新增 dev-only 路由（動態 import） |
| `frontend/src/components/ui/field.tsx`、`label.tsx` | 改 / 刪 | Field → Ark；`label.tsx` 刪除 |
| `frontend/src/components/ui/scroll-area.tsx` | 改 | Ark ScrollArea |
| `frontend/src/components/ui/pagination.tsx`、`hooks/create-pagination.ts`、`features/users/components/ListPagination.tsx` | 改 / 刪 | Ark Pagination + 呼叫端 |
| `frontend/src/components/ui/sidebar/{index,context,parts}.tsx`、`sidebar.test.tsx` | 新增 | sidebar 多部件 |
| `frontend/src/components/layout/AppShell.tsx`（及 Sidebar/Topbar） | 改寫 | 換用新 sidebar 與 splitter |

---

### Task 1: 元件庫庫化骨架

**Files:** 新增 `ui/index.ts`、`ui/registry.json`、`ui/demo/UiDemoPage.tsx`；改 `router/index.tsx`

- [ ] **Step 1: barrel**

建立 `ui/index.ts`，把現有 12 個元件的公開匯出集中（`export * from "./button"` …）。**不要** export `ui/demo/**`。

- [ ] **Step 2: registry.json**

照 spec §2.3 的 schema 為**當時存在的每一個元件**寫一筆（`name/path/kind/ark/deps/tokens/variants/state/docs`）；後續 task 新增或改動的元件由 Task 7 對齊。`ark` 欄對靜態元件填本階段導入的 primitive；`state` 一律 `stable`。

- [ ] **Step 3: dev-only demo 路由**

在 `router/index.tsx` 加入（**只加路由，不動既有路由與守衛**）：

```tsx
const uiDemoRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/ui",
  component: lazy(() => import("~/components/ui/demo/UiDemoPage")),
});
```

`UiDemoPage` 需以 `import.meta.env.DEV` 判斷：非 DEV 時渲染「not available」空頁（避免正式站可用）。

- [ ] **Step 4: 驗證 tree-shake**

```bash
cd frontend && pnpm build && node -e "const fs=require('fs');const d='dist/assets';const f=fs.readdirSync(d).filter(x=>x.endsWith('.js'));const hits=f.filter(x=>fs.readFileSync(d+'/'+x,'utf8').includes('UiDemoPage')||fs.readFileSync(d+'/'+x,'utf8').includes('元件庫展示'));console.log('demo chunk leaked:', hits.length? hits : 'none')"
```

Expected: `none`（正式 bundle 不含 demo 樹）。

- [ ] **Step 5: 驗證**：`pnpm typecheck && pnpm lint && pnpm test && pnpm build` 全綠 0 warning

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/ui/index.ts frontend/src/components/ui/registry.json frontend/src/components/ui/demo frontend/src/router/index.tsx
git commit -m "feat(frontend): 元件庫庫化（barrel、registry、dev-only demo 路由）"
```

---

### Task 2: Field → Ark（並刪除死碼 label.tsx）

**Files:** 改 `ui/field.tsx`；刪 `ui/label.tsx`；新增 `ui/field.test.tsx`

- [ ] **Step 1: 取 Ark 用法**

用 `ark_example("field", <example_id>)`（先 `ark_examples("field")`）與 `ark_props("field")` 確認 anatomy；以 Tailkit 的 form-layouts 結構為樣式基礎（沿用現有 class 與語意 token）。

- [ ] **Step 2: 寫測試（先紅）**

```tsx
import { render, screen } from "@solidjs/testing-library";
import { describe, expect, it } from "vitest";
import { Field, FieldError, FieldLabel } from "./field";
import { Input } from "./input";

describe("Field", () => {
  it("label 與 input 以 for/id 關聯", () => {
    render(() => (
      <Field>
        <FieldLabel for="company-name">公司名稱</FieldLabel>
        <Input id="company-name" />
      </Field>
    ));
    expect(screen.getByLabelText("公司名稱")).toBeTruthy();
  });

  it("錯誤訊息存在時 input 帶 aria-invalid 且被 aria-describedby 指到", () => {
    render(() => (
      <Field invalid>
        <FieldLabel for="company-name">公司名稱</FieldLabel>
        <Input id="company-name" />
        <FieldError>名稱必填</FieldError>
      </Field>
    ));
    const input = screen.getByLabelText("公司名稱");
    expect(input.getAttribute("aria-invalid")).toBe("true");
    const describedBy = input.getAttribute("aria-describedby");
    expect(describedBy).toBeTruthy();
    expect(document.getElementById(describedBy!)?.textContent).toContain("名稱必填");
  });
});
```

- [ ] **Step 3: 跑測試確認紅**：`pnpm exec vitest run src/components/ui/field.test.tsx`

- [ ] **Step 4: 實作**

`Field` → `Field.Root`（`invalid` 由 `props.invalid` 或「有 `FieldError` 子節點」推導）；`FieldLabel` → `Field.Label`（`for` 直傳）；`FieldError` → `Field.ErrorText`；`FieldDescription` → `Field.HelperText`。對外 props 不變（`class`/`children`/`for`/`invalid` 之外不加新必填項）。

- [ ] **Step 5: 刪除 `label.tsx`**

T3 步驟：`label.tsx` 唯一消費者是 `field.tsx`，改完後刪檔；`FieldLabelProps` 改自行定義。用 `grep` 工具確認 `ui/label` 於 `frontend/src` 0 命中（頁面用的是 `FieldLabel`，不受影響）。

- [ ] **Step 6: 測試綠 + 四道 gate**：`pnpm typecheck && pnpm lint && pnpm test && pnpm build`

- [ ] **Step 7: Commit**

```bash
git add frontend/src/components/ui/field.tsx frontend/src/components/ui/field.test.tsx frontend/src/components/ui/label.tsx
git commit -m "refactor(frontend): Field 改用 Ark Field（對外 API 不變，刪除死碼 label.tsx）"
```

---

### Task 3: ScrollArea → Ark

**Files:** 改 `ui/scroll-area.tsx`；新增 `ui/scroll-area.test.tsx`

- [ ] **Step 1: 取 Ark 用法**：`ark_examples("scroll-area")` → `ark_example(...)`；`ark_styling("scroll-area")` 取得各 part 的 data 屬性。

- [ ] **Step 2: 寫測試（先紅）**

```tsx
it("內容超出高度時可捲動", () => {
  render(() => (
    <ScrollArea class="h-20">
      <div style={{ height: "400px" }}>內容</div>
    </ScrollArea>
  ));
  const viewport = screen.getByText("內容").parentElement!;
  expect(viewport.scrollHeight).toBeGreaterThan(viewport.clientHeight);
});
```

（jsdom 的 layout 為零，`scrollHeight/clientHeight` 需以 `Object.defineProperty` 假造；若不可行，改成斷言 Ark 的 `Viewport` 與 `Scrollbar` 結構存在且 `Content` 在 `Viewport` 內——**不要**留無意義的斷言。）

- [ ] **Step 3: 實作**

`ScrollArea`（單一入口，維持 `class`/`children`）內部：`ScrollArea.Root > Viewport > Content` + 兩向 `Scrollbar/Thumb` + `Corner`；樣式用語意 token（`border-border`、`bg-border`）。

- [ ] **Step 4: 呼叫端零改動驗證**：`dialog.tsx` 仍以 `<ScrollArea class="max-h-[80vh] w-full">` 使用；公司/部門 modal 在 Edge 實測可捲動（截圖或 computed style）。

- [ ] **Step 5: 四道 gate + Commit**

```bash
git add frontend/src/components/ui/scroll-area.tsx frontend/src/components/ui/scroll-area.test.tsx
git commit -m "refactor(frontend): ScrollArea 改用 Ark ScrollArea（對外 API 不變）"
```

---

### Task 4: Pagination → Ark（本階段唯一 API 變更）

**Files:** 改 `ui/pagination.tsx`、`features/users/components/ListPagination.tsx`；刪 `hooks/create-pagination.ts`；新增 `ui/pagination.test.tsx`

- [ ] **Step 1: 寫測試（先紅）**

```tsx
it("點下一頁回報新頁碼", () => {
  const onPageChange = vi.fn();
  render(() => <Pagination count={50} page={1} pageSize={10} onPageChange={onPageChange} />);
  fireEvent.click(screen.getByRole("button", { name: /next/i }));
  expect(onPageChange).toHaveBeenCalledWith(2);
});

it("首頁時上一頁不可用", () => {
  render(() => <Pagination count={50} page={1} pageSize={10} />);
  expect(screen.getByRole("button", { name: /previous|上一頁/i })).toBeDisabled();
});
```

（實際 accessible name 依 Ark 的 `translations` 決定；測試用 `getByRole` 抓到的名稱寫死前先實測。）

- [ ] **Step 2: 實作**

包裝層收 `count`/`page`/`pageSize`/`onPageChange` → `Pagination.Root`；頁碼以 `api().pages` 渲染 `Item`；`Prev/Next/First/LastTrigger`、`Ellipsis` 保留為樣式部件。對外仍匯出同名部件。

- [ ] **Step 3: 改 `ListPagination`**：改傳 `count`/`page`/`pageSize`/`onPageChange`；分頁切片改由頁碼驅動。

- [ ] **Step 4: 刪 `create-pagination.ts`**：確認 `grep` 工具在 `frontend/src` 0 命中後刪除。

- [ ] **Step 5: 四道 gate + Edge 實測**（公司/部門頁換頁、首末頁 disabled、ellipsis 位置）

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/ui/pagination.tsx frontend/src/components/ui/pagination.test.tsx frontend/src/features/users/components/ListPagination.tsx frontend/src/hooks/create-pagination.ts
git commit -m "refactor(frontend): Pagination 改用 Ark（頁碼狀態交 Ark，ListPagination 跟隨）"
```

---

### Task 5: Sidebar 多部件（Ark 拼裝）

**Files:** 新增 `ui/sidebar/{index.tsx,context.tsx,parts.tsx,README.md}`、`ui/sidebar/sidebar.test.tsx`

- [ ] **Step 1: 取 Ark 用法**：`ark_examples` + `ark_example` 取 `drawer`、`collapsible`、`splitter`、`tooltip` 的 `basic/controlled` 範例；`ark_styling` 取各 part 的 data 屬性。

- [ ] **Step 2: context（`context.tsx`）**

照 solid-ui 的形狀：`{state, open, setOpen, isMobile, openMobile, setOpenMobile, toggleSidebar}`；常數 `MOBILE_BREAKPOINT=768`、`--sidebar-width: 16rem`／`--sidebar-width-icon: 3rem`／行動版 `18rem`；持久化用 localStorage key `ui:sidebar`（讀取時 try/catch，SSR/無 window 時回退預設）。

- [ ] **Step 3: 寫測試（先紅）**

```tsx
it("切換收合後狀態被記住", async () => {
  localStorage.clear();
  render(() => <SidebarProvider><Sidebar><SidebarContent>內容</SidebarContent></Sidebar></SidebarProvider>);
  fireEvent.click(screen.getByRole("button", { name: /toggle sidebar/i }));
  await waitFor(() => expect(localStorage.getItem("ui:sidebar")).toBe("false"));
});

it("行動版關閉時導覽連結不在 tab 順序", () => {
  // 以 matchMedia stub 觸發 isMobile
  render(() => <SidebarProvider><Sidebar><SidebarContent><a href="/x">連結</a></SidebarContent></Sidebar></SidebarProvider>);
  expect(screen.queryByRole("link", { name: "連結" })).toBeNull();
});
```

- [ ] **Step 4: 實作部件**

`parts.tsx`：`Sidebar`（`Switch`：`collapsible==="none"` / `isMobile` 用 Ark `drawer` / 桌面 div，帶 `data-state|data-collapsible|data-variant|data-side`）、`SidebarHeader/Footer/Content/Group/GroupLabel/GroupContent/Menu/MenuItem/MenuButton`（polymorphic `as`，接 TanStack `Link`）、`SidebarMenuSub`（Ark `collapsible`）、`Trigger`、`Rail`、`Inset`。
**Ark `splitter` 的歸屬**：由 **Task 6 的 AppShell** 持有（`splitter` 的 Panel 是「側欄 / 內容」兩個面板）；本 task 的 `Rail` 只負責渲染 `splitter` 的 `ResizeTrigger` 外觀與提供 toggle（若 Task 6 判斷無法共用，於報告說明並改為純 toggle 按鈕）。
收合時標籤用 Ark `tooltip`；`Cmd/Ctrl+B` 用 `useHotkeys`（**取代** solid-ui 的手寫 keydown listener）。

- [ ] **Step 5: 四道 gate + Edge 實測**（桌面收合/展開/記憶、行動 drawer 的 Esc 與 focus、關閉時不可 Tab、`aria-current` 唯一）

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/ui/sidebar
git commit -m "feat(frontend): sidebar 多部件（solid-ui 解剖 + Ark 行為）"
```

---

### Task 6: AppShell 換用新 sidebar 與 splitter

**Files:** 改寫 `components/layout/AppShell.tsx`（必要時 `Sidebar.tsx`/`Topbar.tsx`）；改 `router/index.tsx`（僅 HomePage 內距）；新增 `src/components/ui/theme.tsx` + `theme.test.tsx`；改 `index.html`（anti-FOUC 腳本）；同步 `ui/index.ts` 與 `ui/registry.json`

- [ ] **Step 0: 深色模式機制（Tailkit class-based dark mode；2026-09-19 使用者指示）**

現況：`index.css` 已有 `.dark` token 與 `@custom-variant dark (&:where(.dark, .dark *))`，但**沒有任何地方套用 `.dark`** → 深色模式在執行期不可達（過去都以注入 class 驗證）。本步驟讓它真的可達：

1. 新增 `src/components/ui/theme.tsx`（或 `lib/theme.ts`，落在 `ui/**` 則遵守依賴規則）：狀態 `"light" | "dark" | "system"`、以 localStorage key `ui:theme` 持久化、跟隨 `prefers-color-scheme`、切換時對 `document.documentElement` 加/移除 `.dark`；讀取與寫入都要 try/catch（無 window／隱私模式）。
2. `index.html`：在 `<head>` 的最前面加**anti-FOUC 內聯腳本**（同步讀 localStorage／系統偏好並設 `.dark`，早於任何 CSS 與 bundle；不得引用外部檔案）。
3. shell（本 task 的 `AppShell`/`Topbar` 或 `SidebarFooter`）：放一個切換器（`light | dark | system` 三態；用 Ark `menu` 或 `toggle-group` 實作，圖示用既有 `lucide-solid`）。
4. 測試（jsdom）：切到 dark → `document.documentElement` 有 `.dark` 且 localStorage 寫入；切到 system → 依 `matchMedia` stub 的結果決定；預設（無儲存值）為 `system`。
5. **驗證方式改變**：本 task 之後的視覺驗證（含 Task 8）改用**真實切換器**，不再注入 class。

- [ ] **Step 1: 以 `SidebarProvider` + `Sidebar` + `SidebarInset` 組裝**；側欄與內容之間用 Ark `splitter`（`Panel` + `ResizeTrigger`），保留 v2 的內距歸屬決定（shell 是唯一內距所有者）。

- [ ] **Step 2: 修 v2 遺留**：`router/index.tsx` 的 `HomePage` 自帶 `p-8` 與 shell 內距疊加 → 移除頁面層內距（只改 class，不動路由/守衛）。

- [ ] **Step 3: 修 v2 的兩項 a11y**：`Sidebar` 品牌列不再是 `Link to="/"`（或移出 `<nav>`）→ 同一 nav 只有一個 `aria-current="page"`；行動版關閉時 `invisible`（若 Step 5 的 Ark drawer 已自然滿足，則以測試證明並在報告說明）。

- [ ] **Step 4: 四道 gate + Edge 實測**（桌面/手機 × **以真實切換器切淺/深色**、拖曳調寬、`Cmd/Ctrl+B`、`/login` 與 `/403` 不套 shell；另確認重新整理後不閃色（anti-FOUC）與偏好記憶）

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/layout frontend/src/router/index.tsx
git commit -m "refactor(frontend): AppShell 換用新 sidebar 與 splitter"
```

---

### Task 7: demo 頁內容、registry/README 補齊、依賴規則

**Files:** 改 `ui/demo/UiDemoPage.tsx`；新增各元件 `.md` 與 `ui/sidebar/README.md`；改 `ui/registry.json`

- [ ] **Step 1: demo 頁**：逐元件展示所有變體與狀態（button 各 variant/size、badge、card、input/label/field（含錯誤態）、table、pagination、spinner、dialog、tabs、checkbox、scroll-area、sidebar 三種 variant × 兩種 side × 三種 collapsible）；含深色預覽切換按鈕（切換 `document.documentElement.classList` 的 `dark`）。

- [ ] **Step 2: README**：每個元件一份，含用途、API 表、Ark 對應、a11y 要點、範例（**不重複程式碼的長篇說明**）。

- [ ] **Step 3: registry 補齊**：與實際檔案/primitive/variants 對齊。

- [ ] **Step 4: 依賴規則驗證**：用 `grep` 工具搜 `ui/**` 內是否有 `features/`、`router/`、`lib/proto` 匯入 → 0 命中。

- [ ] **Step 4b: import 改走 barrel**

把 `features/**`、`components/layout/**`、`App.tsx` 的元件匯入由逐檔深引（例如 `~/components/ui/dialog`）改為 `~/components/ui`（TypeScript 會指出所有遺漏）。**只改 import 路徑，不動任何邏輯**。改完以 `grep` 工具確認 `from "~/components/ui/` 於 `frontend/src/features`、`frontend/src/components/layout` 0 命中（`ui/**` 內部互相引用不受此限）。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ui
git commit -m "docs(frontend): 元件庫 demo 頁、README 與 registry 補齊"
```

---

### Task 8: 端到端驗收

- [ ] **Step 1: 四道 gate**：`pnpm typecheck && pnpm lint && pnpm test && pnpm build`（0 warning）

- [ ] **Step 2: Edge 實測清單**（逐項記錄結果與截圖路徑）
1. `/ui`：每元件所有變體渲染、**以真實切換器切換淺/深色**（Task 6 的主題切換器）
2. shell：桌面收合/展開 + 重整後記憶、拖曳調寬、`Cmd/Ctrl+B`、行動 drawer（Esc、focus 回到觸發鈕、關閉後 Tab 不進導覽）、**主題切換器的三態（light/dark/system）與重整後不閃色**
3. 既有 5 頁：登入（兩 tab、表單錯誤關聯）、403、公司/部門（分頁換頁、modal 捲動、CRUD）、角色權限（checkbox、儲存）
4. `aria-current` 唯一性（同一 nav）

- [ ] **Step 3: `impeccable detect`** 掃 `src/components/**` 與改動頁面 → 有發現即修或說明保留理由。

- [ ] **Step 4: 收尾衛生**：`pkill -f "Microsoft Edge.*--headless"`、`pgrep -f "Microsoft Edge.*--headless" | wc -l` 為 0、`rm -rf /tmp/msedge-*`、repo 無暫存檔。

- [ ] **Step 5: 回報**：把 Step 1–4 的實際輸出貼進報告（含未驗項）。

---

## 驗收對照（spec §5）

| spec 驗收 | 步驟 |
|---|---|
| 1. 四道 gate 全綠、lint 0 warning | Task 8 Step 1（各 task 各自也跑） |
| 2. `/ui` demo 以 Edge 實測 | Task 7 Step 1 + Task 8 Step 2.1 |
| 3. shell 實測（收合/記憶/拖寬/快捷鍵/drawer） | Task 6 Step 4 + Task 8 Step 2.2 |
| 4. 既有 5 頁在新 API 下正常 | Task 8 Step 2.3 |
| 5. registry 與 README 齊備 + 依賴規則 | Task 7 Step 3–4 |
| 6. impeccable detect | Task 8 Step 3 |

---

*建立：2026-09-19*
