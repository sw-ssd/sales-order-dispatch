# 前端 UI 重構：Ark UI × Tailkit 實作計畫（v2）

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `frontend/` 的 UI 層重建為「Ark UI 管行為、Tailkit 管視覺、本地複合元件當唯一組合點」，並新增 app shell 與 5 頁換皮。

**Architecture:** 視覺來自 Tailkit 的結構與尺寸（取 `tech: "html"` 後手動移植成 SolidJS），顏色一律改寫為 `src/index.css` 的語意 token（Tailkit 的色階字面值與其 `dark:` 顏色變體全部丟棄，交由 token 在深色模式自動翻轉）；互動元件的行為由 `@ark-ui/solid` primitives 承擔，取代 Tailkit HTML 版的 Alpine。

**Tech Stack:** SolidJS 1.9、`@ark-ui/solid` 5.39.2、Tailwind CSS v4.3、`@tailwindcss/forms`、`@tailwindcss/typography`、TanStack Router/Query、Vitest + jsdom + `@solidjs/testing-library`、pnpm workspace。

**Spec:** `docs/superpowers/specs/2026-09-19-frontend-tailkit-ark-design.md`

## Global Constraints

- 套件管理器 pnpm；指令一律在 `frontend/` 執行；lockfile 在 repo 根。
- **Tailkit 取碼**：用 kernel 工具 `tailkit_search(query)` 與 `tailkit_code(identifiers, tech)`（`tech` 用 `"html"`）。這兩個工具在父 session 的 kernel 內執行，OAuth token 不落地、不進對話。**Tailkit 不是 npm 依賴，不得寫進 `package.json`。**
- **允許新增的依賴只有** `@tailwindcss/forms`、`@tailwindcss/typography`（Tailkit 官方要求）；不得新增其他依賴。
- **顏色一律語意 token**（下表）。Tailkit markup 中的色階字面值（`secondary-*`、`orange-*`、`emerald-*`、`white`、`black/N`）**不得**原樣保留。
- **顏色相關的 `dark:` 變體一律刪除**（token 已在深色模式翻轉）；非顏色的 `dark:`（如 `dark:shadow-none`）保留。
- **對外 API 照 Ark 語意，且不外洩 `details` 物件**：回呼一律吃純值（`onOpenChange(open: boolean)`、`onValueChange(value: string)`、`onCheckedChange(checked: boolean)`）。
- Ark 5.39.2 **沒有** `Portal` 匯出 → 用 `solid-js/web` 的 `Portal`。
- **測試禁則**：不攔截或手動推進 rAF、不 mock Ark 內部、不斷言 Ark 私有屬性、不留「只斷言沒被呼叫」的假綠案例；非同步一律用真實等待（`waitFor` / microtask tick），不用假時鐘。
- `pnpm lint` 必須 **0 warning**。
- 既有 `src/lib/ability/**`、`src/lib/transport.ts`、`src/lib/query-client.ts`、`src/lib/proto/**`、`router/index.tsx` 的路由路徑與守衛 **不得改動**。
- 註解、文件、commit message 一律繁體中文。
- **Tailkit 授權邊界**：Tailkit 為付費授權產品。取回的 markup 只作為改寫參考；**不得把 Tailkit 原始檔案原樣入庫**，只保留改寫後的 SolidJS 元件。`tech` 一律用 `"html"`。

### Tailkit → 語意 token 映射表（T2 起每個任務都適用）

| Tailkit 類別 | 改成 |
|---|---|
| `bg-white` / `dark:bg-secondary-800` | `bg-card` |
| `bg-secondary-50` / `dark:bg-secondary-900` | `bg-background`（頁面底）或 `bg-muted`（區塊底） |
| `bg-secondary-100` / `dark:bg-secondary-800/50` | `bg-muted` |
| `border-secondary-200` / `dark:border-secondary-700` | `border-border` |
| `text-secondary-800/900/950` / `dark:text-secondary-100/50` | `text-foreground` |
| `text-secondary-500/600` / `dark:text-secondary-400` | `text-muted-foreground` |
| `placeholder-secondary-500` / `dark:placeholder-secondary-400` | `placeholder:text-muted-foreground` |
| `bg-secondary-900/75`（modal backdrop） | `bg-foreground/75` |
| `ring-secondary-200/50`、`ring-offset-secondary-900` | `ring-border`、`ring-offset-background` |
| `shadow-secondary-300/25` | `shadow-border/25`（無對應者直接刪 shadow 色） |
| `orange-*`（品牌色：`bg-orange-700`、`text-orange-500/600`、`border-orange-*`） | `bg-primary` / `text-primary` / `border-primary`；淺底用 `bg-primary/10` |
| `emerald-*`（成功狀態） | `bg-success/15`、`text-success`（token 於 T1 建立，值於 T4 校正） |
| `orange-100/900`（表格警告狀態） | `bg-warning/15`、`text-warning`（token 於 T1 建立，值於 T4 校正） |
| 資訊狀態（Tailkit 無對應色） | `bg-info/15`、`text-info`（token 於 T4 建立） |
| 第三方品牌色（`text-[#1877f2]` 等） | **保留字面值**（唯一例外） |

---

## 檔案結構

| 檔案 | 動作 | 責任 |
|---|---|---|
| `frontend/package.json` | 改 | `+ @tailwindcss/forms`、`+ @tailwindcss/typography`、`− @kobalte/core` |
| `frontend/src/index.css` | 改 | Tailkit 的 `@plugin`/`@theme`、dark variant 對齊、`--success`/`--warning` token |
| `frontend/index.html` | 改 | Inter 字型（Bunny.net）與 `lang` 保持 `zh-Hant` |
| `frontend/src/components/ui/{button,card,badge}.tsx` | 改寫 | Tailkit 結構 + 語意 token |
| `frontend/src/components/ui/{input,label,field}.tsx` | 改寫 | 同上 |
| `frontend/src/components/ui/{table,pagination,spinner}.tsx` | 改寫 | 同上 |
| `frontend/src/components/ui/{dialog,tabs,checkbox}.tsx` | 重寫 | Ark 行為 + Tailkit 結構 |
| `frontend/src/components/ui/scroll-area.tsx` | 不動 | 自刻保留（Tailkit 無對應） |
| `frontend/src/components/layout/{AppShell,Sidebar,Topbar}.tsx` | 新增 | Tailkit Light Sidebar 版型 |
| `frontend/src/App.tsx` | 改 | 掛載 AppShell（登入／403 除外） |
| `frontend/src/features/**`（5 頁） | 改寫 | 換皮，邏輯不動 |
| `frontend/nikala.config.json`、`.cursorrules`、`.cursor/rules/nikala.mdc` | 刪除 | Nikala 痕跡 |

---

### Task 1: 依賴、Tailwind 設定、token 基礎

**Files:**
- Modify: `frontend/package.json`、`frontend/src/index.css`、`frontend/index.html`

**Interfaces:**
- Produces: `@tailwindcss/forms`、`@tailwindcss/typography` 可用；`--success`、`--warning` token 可用；`dark` variant 語意對齊 Tailkit。

- [ ] **Step 1: 安裝 Ark 依賴與 Tailwind 外掛**

```bash
cd frontend && pnpm add @ark-ui/solid@^5.39.2 && pnpm add -D @tailwindcss/forms @tailwindcss/typography
```

驗證：`pnpm ls @ark-ui/solid solid-js` 顯示 5.39.2 與 1.9.x，無 peer 警告。（`@ark-ui/solid` 是 T5–T7 的前置；`@kobalte/core` 仍在，Task 10 才移除。）

- [ ] **Step 2: 改 `src/index.css`**

在 `@import "tailwindcss";` 之後加入：

```css
/* Tailkit 官方要求的 Tailwind 外掛 */
@plugin "@tailwindcss/forms";
@plugin "@tailwindcss/typography";

/* 深色模式：對齊 Tailkit（.dark 自身與其子孫都套用） */
@custom-variant dark (&:where(.dark, .dark *));
```

在 `:root` 內追加狀態色（數值自訂，需與既有 `--primary` 的 oklch 風格一致）：

```css
  --success: oklch(0.65 0.15 155);
  --success-foreground: oklch(0.985 0 0);
  --warning: oklch(0.75 0.15 95);
  --warning-foreground: oklch(0.205 0 0);
```

在 `@theme inline`（或既有對應區塊）內讓它們變成 Tailwind 顏色：

```css
  --color-success: var(--success);
  --color-success-foreground: var(--success-foreground);
  --color-warning: var(--warning);
  --color-warning-foreground: var(--warning-foreground);
```

在 `@theme` 內追加 Tailkit 的項目：`--default-font-family: "Inter", ui-sans-serif, system-ui, sans-serif;`（**必須帶 fallback**：Tailwind preflight 的 `--theme(--default-font-family, <系統堆疊>)` 只在該鍵未定義時才用系統堆疊，一旦定義就整串被取代，bunny.net 被 CSP／離線阻擋時拉丁字會退成 serif）、`--spacing-8xl: 90rem; --spacing-9xl: 105rem; --spacing-10xl: 120rem;`、`--animate-spin-slow: spin-slow 8s linear infinite;`（含對應 `@keyframes spin-slow`）。

- [ ] **Step 3: 字型載入 `index.html`**

依 Tailkit 官方 HTML 結構文件，在 `<head>` 內加入（`lang="zh-Hant"` 與 `<title>` 不動）：

```html
<link rel="preconnect" href="https://fonts.bunny.net" />
<link href="https://fonts.bunny.net/css2?family=Inter:wght@300;400;500;600;700;800;900&display=swap" rel="stylesheet" />
```

- [ ] **Step 4: 驗證**

```bash
cd frontend && pnpm typecheck && pnpm lint && pnpm test && pnpm build
```

Expected: 全綠、0 warning。`@tailwindcss/forms` 會重設表單元素預設樣式 → 若 `pnpm test` 有 UI 相關失敗，先確認是否為 forms plugin 造成（記錄在報告）。

- [ ] **Step 5: 目視核對 forms plugin 的影響**

啟動 dev server，逐頁看：登入頁（店家登入表單）、公司／部門 modal 表單、角色權限矩陣。記錄任何「輸入框高度／邊框／focus ring」的非預期位移到報告（Task 3 會一併處理）。

- [ ] **Step 6: Commit**

```bash
git add frontend/package.json frontend/src/index.css frontend/index.html pnpm-lock.yaml
git commit -m "chore(frontend): Tailwind 外掛與 token 基礎（Tailkit 前置）"
```

---

### Task 2: 靜態元件 A — button / card / badge

**Files:**
- Modify: `frontend/src/components/ui/button.tsx`、`card.tsx`、`badge.tsx`

**Interfaces:**
- Consumes: kernel 工具 `tailkit_search` / `tailkit_code`
- Produces: 三個元件的對外 props 與 variants **完全不變**（頁面不得改）

- [ ] **Step 1: 取 Tailkit 結構**

```
tailkit_code("a-c-buttons-01,a-c-buttons-03,a-c-buttons-08,a-c-cards-09,a-c-badges-01,a-c-badges-04", tech="html")
```

（先用 `tailkit_search("button variants")` 等查詢確認是否有更貼近的變體；把實際取用的 id 記在報告。）

- [ ] **Step 2: 依映射表改寫三個元件**

規則：
1. 保留既有 props 簽章與 `cva` 的 variant 名稱；只換 class 與內部結構。
2. 顏色一律照 §映射表 換成語意 token；刪除顏色相關的 `dark:` 變體。
3. 既有 `focus-visible:ring-*`、disabled 樣式不可退化。

- [ ] **Step 3: 驗證**

```bash
cd frontend && pnpm typecheck && pnpm lint && pnpm test && pnpm build
```

Expected: 全綠、0 warning。

- [ ] **Step 4: 目視**

dev server 上比對：公司／部門頁的按鈕（primary/secondary/outline/danger）、卡片、狀態 badge（含深色模式切換）。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ui/button.tsx frontend/src/components/ui/card.tsx frontend/src/components/ui/badge.tsx
git commit -m "refactor(frontend): button/card/badge 改用 Tailkit 結構與語意 token"
```

---

### Task 3: 靜態元件 B — input / label / field

**Files:**
- Modify: `frontend/src/components/ui/input.tsx`、`label.tsx`、`field.tsx`

**Interfaces:**
- Produces: 對外 props 不變；`Field`/`FieldLabel` 仍可包住 `Input`。

- [ ] **Step 1: 取 Tailkit 結構**

```
tailkit_code("a-c-form-elements-01,a-c-form-elements-16,a-c-form-input-groups-01,a-c-form-layouts-01", tech="html")
```

- [ ] **Step 2: 依映射表改寫**

要點：`label` 的尺寸/字重照 Tailkit；`input` 的 `placeholder` 用 `placeholder:text-muted-foreground`；錯誤態（若有）用 `border-destructive`；`field` 的間距照 `form-layouts`。

- [ ] **Step 3: 驗證**

```bash
cd frontend && pnpm typecheck && pnpm lint && pnpm test && pnpm build
```

- [ ] **Step 4: 目視**：登入頁兩個 tab 的表單、公司／部門 modal 表單（含 focus ring、錯誤訊息）。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ui/input.tsx frontend/src/components/ui/label.tsx frontend/src/components/ui/field.tsx
git commit -m "refactor(frontend): input/label/field 改用 Tailkit 結構與語意 token"
```

---

### Task 4: 靜態元件 C — table / pagination / spinner

**Files:**
- Modify: `frontend/src/components/ui/table.tsx`、`pagination.tsx`、`spinner.tsx`、`frontend/src/index.css`（狀態色 token）

**Interfaces:**
- Produces: 對外 props 不變；`--success`/`--warning`/`--info` 在淺色與深色模式下都能當**文字色**使用（T2 實測淺色 `bg-warning/15 + text-warning` 對比僅約 1.8:1）

- [ ] **Step 0: 修狀態色 token（`index.css`）**

T2 的回報指出：`--warning: oklch(0.75 0.15 95)` 當文字用對比不足，且 `info` 無 token（T2 暫借 primary）。作法是讓 status token 在淺色模式偏深、深色模式偏亮，並在 `.dark` 區塊覆寫：

```css
:root {
  --success: oklch(0.55 0.13 155);
  --warning: oklch(0.55 0.13 85);
  --info: oklch(0.55 0.13 250);
}
.dark {
  --success: oklch(0.80 0.13 155);
  --warning: oklch(0.82 0.13 85);
  --info: oklch(0.80 0.13 250);
}
```

`@theme inline` 需補 `--color-info: var(--info);`（`--color-success`/`--color-warning` 已存在）。驗收：`bg-warning/15 text-warning` 在淺色的對比 ≥ 4.5:1（用 chrome-headless-shell 量 computed 值或以 WCAG 公式核算，把數字寫進報告）。

- [ ] **Step 0b: 回改 `info` variant（T2 的權宜）**

T2 在 `--info` 尚不存在時借用 primary：`frontend/src/components/ui/button.tsx:34`（`border-primary/30 bg-primary/10 text-primary hover:bg-primary/15`）與 `frontend/src/components/ui/badge.tsx:26`（`bg-primary/10 text-primary`）。token 落地後改為映射表指定的 `bg-info/15`、`text-info`（hover 用 `bg-info/20`）。這 2 檔屬本 task 可改範圍。

- [ ] **Step 1: 取 Tailkit 結構**

```
tailkit_code("a-c-tables-13,a-c-tables-01,a-c-pagination-01", tech="html")
```

`spinner` 在 Tailkit 無直接對應（其最接近者是 Progress Bars）→ 保留現行結構，只把顏色換成語意 token。

- [ ] **Step 2: 依映射表改寫**

表格的狀態標籤用新 token：成功 `bg-success/15 text-success`、警告 `bg-warning/15 text-warning`、停用 `bg-destructive/15 text-destructive`。

- [ ] **Step 3: 驗證**

```bash
cd frontend && pnpm typecheck && pnpm lint && pnpm test && pnpm build
```

- [ ] **Step 4: 目視**：公司／部門／角色三頁的表格、分頁列、載入中狀態（深色模式一起看）。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ui/table.tsx frontend/src/components/ui/pagination.tsx frontend/src/components/ui/spinner.tsx
git commit -m "refactor(frontend): table/pagination/spinner 改用 Tailkit 結構與語意 token"
```

---

### Task 5: dialog（Ark 行為 + Tailkit Modals）

**Files:**
- Modify: `frontend/src/components/ui/dialog.tsx`
- Test: `frontend/src/components/ui/dialog.test.tsx`（新增）

**Interfaces:**
- Consumes: `@ark-ui/solid` 的 `Dialog`；`solid-js/web` 的 `Portal`
- Produces（對外，頁面不用改）:
  - `Dialog(open?: boolean, onOpenChange?: (open: boolean) => void, closeOnOutsideClick?: boolean, children)`
  - `DialogContent(showCloseButton?: boolean, blur?: boolean, class?, children)`
  - `DialogHeader`、`DialogFooter`、`DialogTitle`、`DialogDescription`、`DialogClose`
  - **不再匯出** `DialogOverlay`、`DialogTrigger`（全 repo 無呼叫端）

- [ ] **Step 1: 寫失敗測試**

```tsx
import { render, screen, fireEvent, waitFor } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";
import { createSignal } from "solid-js";
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogTitle } from "./dialog";

const Harness = (props: {
  onOpenChange?: (open: boolean) => void;
  closeOnOutsideClick?: boolean;
  showCloseButton?: boolean;
}) => {
  const [open, setOpen] = createSignal(true);
  return (
    <Dialog
      open={open()}
      closeOnOutsideClick={props.closeOnOutsideClick}
      onOpenChange={(value) => {
        setOpen(value);
        props.onOpenChange?.(value);
      }}
    >
      <DialogContent showCloseButton={props.showCloseButton}>
        <DialogTitle>新增公司</DialogTitle>
        <DialogDescription>建立新的公司主檔</DialogDescription>
        <DialogClose>取消</DialogClose>
      </DialogContent>
    </Dialog>
  );
};

describe("Dialog", () => {
  it("open 時顯示標題與說明", () => {
    render(() => <Harness />);
    expect(screen.getByText("新增公司")).toBeTruthy();
    expect(screen.getByText("建立新的公司主檔")).toBeTruthy();
  });

  it("點關閉鈕回報 onOpenChange(false)", () => {
    const onOpenChange = vi.fn();
    render(() => <Harness onOpenChange={onOpenChange} />);
    fireEvent.click(screen.getByText("取消"));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("showCloseButton=false 時沒有關閉鈕", () => {
    render(() => <Harness showCloseButton={false} />);
    expect(screen.queryByText("Close")).toBeNull();
  });

  it("按 Escape 回報 onOpenChange(false)", async () => {
    const onOpenChange = vi.fn();
    render(() => <Harness onOpenChange={onOpenChange} />);
    fireEvent.keyDown(document.body, { key: "Escape", code: "Escape" });
    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
  });
});
```

- [ ] **Step 2: 跑測試確認紅**

```bash
cd frontend && pnpm exec vitest run src/components/ui/dialog.test.tsx
```

Expected: FAIL（舊實作不存在 `closeOnOutsideClick` 在 `Dialog` 上、且仍依賴 Kobalte）。

- [ ] **Step 3: 實作**

```
tailkit_code("a-c-modals-01,a-c-modals-02", tech="html")
```

Ark anatomy：`Dialog.Root > Portal > Backdrop + Positioner > Content`，`Content` 內含 `Title`/`Description`/`CloseTrigger`。要點：
- `closeOnOutsideClick` 直接對應 Root 的 `closeOnInteractOutside`（**不要**再從 Content 往上推）。
- `onOpenChange={(details) => props.onOpenChange?.(details.open)}`。
- 置中交給 `Positioner`（`fixed inset-0 flex items-center justify-center p-4`），`Content` 不再自帶 `translate`。
- 顏色照映射表：backdrop `bg-foreground/75`（`blur` 為真時加 `backdrop-blur-sm`）、卡片 `bg-card border-border`、標題 `text-foreground`、說明 `text-muted-foreground`。
- 動畫類別以 Ark 的 `data-[state=open]` / `data-[state=closed]` 表達。

- [ ] **Step 4: 跑測試至綠**

```bash
cd frontend && pnpm exec vitest run src/components/ui/dialog.test.tsx
```

若 Escape 案例在 jsdom 無法驅動，**刪掉該案例**並在報告註明交由 Task 10 瀏覽器驗證；不得留假綠。

- [ ] **Step 5: 驗證**

```bash
cd frontend && pnpm typecheck && pnpm lint && pnpm test && pnpm build
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/ui/dialog.tsx frontend/src/components/ui/dialog.test.tsx
git commit -m "refactor(frontend): dialog 改用 Ark 行為 + Tailkit Modals 結構"
```

---

### Task 6: tabs（Ark 行為 + Tailkit Tabs）

**Files:**
- Modify: `frontend/src/components/ui/tabs.tsx`
- Test: `frontend/src/components/ui/tabs.test.tsx`（新增）

**Interfaces:**
- Produces（對外）: `Tabs(value?, defaultValue?, onValueChange?: (value: string) => void, orientation?, class?)`、`TabsList`、`TabsTrigger(value, disabled?)`、`TabsContent(value)`

- [ ] **Step 1: 寫失敗測試**

```tsx
import { render, screen, fireEvent } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";
import { createSignal } from "solid-js";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./tabs";

const Harness = (props: { onValueChange?: (v: string) => void; disableSecond?: boolean }) => {
  const [value, setValue] = createSignal("employee");
  return (
    <Tabs
      value={value()}
      onValueChange={(next) => {
        setValue(next);
        props.onValueChange?.(next);
      }}
    >
      <TabsList>
        <TabsTrigger value="employee">員工</TabsTrigger>
        <TabsTrigger value="store" disabled={props.disableSecond}>店家</TabsTrigger>
      </TabsList>
      <TabsContent value="employee">員工內容</TabsContent>
      <TabsContent value="store">店家內容</TabsContent>
    </Tabs>
  );
};

describe("Tabs", () => {
  it("受控 value 決定顯示的 panel", () => {
    render(() => <Harness />);
    expect(screen.getByText("員工內容")).toBeTruthy();
    expect(screen.queryByText("店家內容")).toBeNull();
  });

  it("點第二個 tab 帶回新值並切換內容", () => {
    const onValueChange = vi.fn();
    render(() => <Harness onValueChange={onValueChange} />);
    fireEvent.click(screen.getByText("店家"));
    expect(onValueChange).toHaveBeenCalledWith("store");
    expect(screen.getByText("店家內容")).toBeTruthy();
  });

  it("disabled trigger 不切換", () => {
    const onValueChange = vi.fn();
    render(() => <Harness disableSecond onValueChange={onValueChange} />);
    fireEvent.click(screen.getByText("店家"));
    expect(onValueChange).not.toHaveBeenCalled();
  });

  it("選中的 trigger 對輔助科技標示為已選", () => {
    render(() => <Harness />);
    expect(screen.getByRole("tab", { selected: true }).textContent).toBe("員工");
  });
});
```

- [ ] **Step 2: 跑測試確認紅**

```bash
cd frontend && pnpm exec vitest run src/components/ui/tabs.test.tsx
```

- [ ] **Step 3: 實作**

```
tailkit_code("a-c-tabs-11,a-c-tabs-01", tech="html")
```

Ark anatomy：`Tabs.Root/List/Trigger/Content`；`onValueChange={(details) => props.onValueChange?.(details.value)}`；`orientation` 直傳；以 `lazyMount` + `unmountOnExit` 保持「未選中 panel 不在 DOM」的既有語意。
樣式：選中態用 Ark 的 `data-selected:` 變體（Tailkit 的 `bg-white`/`dark:bg-secondary-900` → `data-selected:bg-card`）；`data-orientation` 供垂直版面用。
**不要**手寫 `role="tab"`/`aria-selected`（Ark 提供），也**不要**攔截 rAF 來測鍵盤。

- [ ] **Step 4: 跑測試至綠**

```bash
cd frontend && pnpm exec vitest run src/components/ui/tabs.test.tsx
```

- [ ] **Step 5: 更新呼叫端（否則 typecheck 會紅）**

`frontend/src/features/auth/pages/LoginPage.tsx` 目前傳 `onChange={(value) => setTab(value as LoginTab)}` → 改為 `onValueChange={...}`（**只改 prop 名，邏輯不動**）。

- [ ] **Step 6: 驗證**：`pnpm typecheck && pnpm lint && pnpm test && pnpm build`

- [ ] **Step 7: Commit**

```bash
git add frontend/src/components/ui/tabs.tsx frontend/src/components/ui/tabs.test.tsx frontend/src/features/auth/pages/LoginPage.tsx
git commit -m "refactor(frontend): tabs 改用 Ark 行為 + Tailkit Tabs 結構"
```

---

### Task 7: checkbox（Ark 行為 + Tailkit Form Elements）

**Files:**
- Modify: `frontend/src/components/ui/checkbox.tsx`
- Test: `frontend/src/components/ui/checkbox.test.tsx`（新增）

**Interfaces:**
- Produces（對外）: `Checkbox(checked?, defaultChecked?, onCheckedChange?: (checked: boolean) => void, disabled?, class?, …aria)`

- [ ] **Step 1: 寫失敗測試**

```tsx
import { render, screen, fireEvent } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";
import { Checkbox } from "./checkbox";

const isChecked = () => {
  const el = screen.getByRole("checkbox") as HTMLInputElement;
  return el.tagName === "INPUT" ? el.checked : el.getAttribute("aria-checked") === "true";
};

describe("Checkbox", () => {
  it("受控 checked=false 點擊回報 true", () => {
    const onCheckedChange = vi.fn();
    render(() => <Checkbox checked={false} onCheckedChange={onCheckedChange} aria-label="客戶 檢視" />);
    fireEvent.click(screen.getByRole("checkbox"));
    expect(onCheckedChange).toHaveBeenCalledWith(true);
  });

  it("受控 checked=true 點擊回報 false", () => {
    const onCheckedChange = vi.fn();
    render(() => <Checkbox checked onCheckedChange={onCheckedChange} aria-label="客戶 檢視" />);
    fireEvent.click(screen.getByRole("checkbox"));
    expect(onCheckedChange).toHaveBeenCalledWith(false);
  });

  it("非受控 defaultChecked 可切換", () => {
    render(() => <Checkbox defaultChecked aria-label="商品 檢視" />);
    expect(isChecked()).toBe(true);
    fireEvent.click(screen.getByRole("checkbox"));
    expect(isChecked()).toBe(false);
  });

  it("disabled 不觸發", () => {
    const onCheckedChange = vi.fn();
    render(() => <Checkbox disabled onCheckedChange={onCheckedChange} aria-label="訂單 檢視" />);
    fireEvent.click(screen.getByRole("checkbox"));
    expect(onCheckedChange).not.toHaveBeenCalled();
  });

  it("aria-label 成為可辨識名稱", () => {
    render(() => <Checkbox aria-label="客戶 檢視" />);
    expect(screen.getByRole("checkbox", { name: "客戶 檢視" })).toBeTruthy();
  });
});
```

- [ ] **Step 2: 跑測試確認紅**

```bash
cd frontend && pnpm exec vitest run src/components/ui/checkbox.test.tsx
```

- [ ] **Step 3: 實作**

```
tailkit_code("a-c-form-elements-16,a-c-form-elements-01", tech="html")
```

Ark anatomy：`Checkbox.Root`（渲染 `<label>`）`> Control + Indicator + HiddenInput`。要點：
- `onCheckedChange={(details) => props.onCheckedChange?.(details.checked === true)}`。
- 勾選記號放 `Indicator`；`Control` 承載方框樣式（Tailkit 的 checkbox 尺寸/圓角照抄，顏色用 `border-primary`/`bg-primary text-primary-foreground`）。
- `aria-label` 的落點以測試結果為準：`Root` 取不到名稱時改轉發到 `HiddenInput`（**只放一處**）。
- `CheckboxProps` 目前 extends `JSX.ButtonHTMLAttributes<HTMLButtonElement>`；若 typecheck 因 Root 是 `<label>` 而報錯，改基底為 `JSX.LabelHTMLAttributes<HTMLLabelElement>`（對外 props 不變），不要用 `as any`。

- [ ] **Step 4: 跑測試至綠**

```bash
cd frontend && pnpm exec vitest run src/components/ui/checkbox.test.tsx
```

- [ ] **Step 5: 更新呼叫端（否則 typecheck 會紅）**

`frontend/src/features/users/components/PermissionMatrix.tsx` 目前傳 `onChange={(value) => toggle(resource, action, value)}` → 改為 `onCheckedChange={...}`（只改 prop 名與參數語意，邏輯不動）。

- [ ] **Step 6: 驗證**：`pnpm typecheck && pnpm lint && pnpm test && pnpm build`

- [ ] **Step 7: 刪除死碼 hook**

T6 已把 `tabs.tsx` 的用法移除，本 task 移除 `checkbox.tsx` 之後 `hooks/create-controllable-signal.ts` 即無人引用 → 刪除該檔，並用 `grep` 工具確認 `create-controllable-signal` 於 `frontend/src` 0 命中。若仍有引用 → 停下來回報，不要保留。

- [ ] **Step 8: Commit**

```bash
git add frontend/src/components/ui/checkbox.tsx frontend/src/components/ui/checkbox.test.tsx frontend/src/features/users/components/PermissionMatrix.tsx frontend/src/hooks/create-controllable-signal.ts
git commit -m "refactor(frontend): checkbox 改用 Ark 行為 + Tailkit Form Elements 結構"
```

---

### Task 8: App shell（Tailkit Light Sidebar）

**Files:**
- Create: `frontend/src/components/layout/AppShell.tsx`、`Sidebar.tsx`、`Topbar.tsx`
- Modify: `frontend/src/App.tsx`、`frontend/src/router/index.tsx`（僅為了區分登入／403 不套 shell）

**Interfaces:**
- Consumes: `~/components/ui/*`、TanStack Router 的 `useRouterState`/`Link`
- Produces: `<AppShell>{children}</AppShell>`；`App.tsx` 依當前路徑決定是否包 shell

- [ ] **Step 1: 取 Tailkit 版型**

```
tailkit_code("a-l-light-sidebar-01,a-l-light-sidebar-03", tech="html")
```

- [ ] **Step 2: 實作 shell**

- 側邊欄項目依 Pixso 稿順序，只列出存在的路由（首頁、公司、部門、角色權限），其餘目標渲染為 `<span aria-disabled="true">` 佔位（不可點、`text-muted-foreground/60`），**不得**指向不存在的路徑。
- `AppShell` 用 `useRouterState({ select: (s) => s.location.pathname })` 判斷當前項並套用選中樣式。
- 登入頁（`/login`）與 403（`/403`）不套 shell。
- 行動寬度下側邊欄收合為頂列按鈕（依 Tailkit 版型的響應式做法，或最小可用替代並在報告說明）。

- [ ] **Step 3: 驗證**

```bash
cd frontend && pnpm typecheck && pnpm lint && pnpm test && pnpm build
```

- [ ] **Step 4: 目視**：桌面與手機寬度、深色模式、選中態、佔位項不可點。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/layout frontend/src/App.tsx frontend/src/router/index.tsx
git commit -m "feat(frontend): 新增 app shell（Tailkit Light Sidebar）"
```

---

### Task 9: 頁面換皮（5 頁）

**Files:**
- Modify: `frontend/src/features/auth/pages/LoginPage.tsx`、`ForbiddenPage.tsx`、`frontend/src/features/users/pages/{CompaniesPage,DepartmentsPage,RolesPage}.tsx`（必要時含其 components）

**Interfaces:**
- Consumes: Task 1–8 的元件與 token
- Produces: 頁面外觀照 Tailkit；**邏輯、欄位、API 呼叫、守衛一律不動**

- [ ] **Step 1: 取 Tailkit 樣板**

```
tailkit_code("a-p-sign-in-01,a-p-errors-01", tech="html")
```

- [ ] **Step 2: 換皮**

- 登入頁：Tailkit Sign In 版面（左右分欄或置中卡片，依取回樣板），兩個 tab 與 Google 登入流程照舊。
- 403：Tailkit Errors 版面，文案與導向不動。
- 三頁 users：標題區塊（Page Headings）、表格、modal、表單照 Task 2–7 的元件；頁面層只保留版面與間距。
- 顏色一律語意 token；不得為了「像 Tailkit」而保留色階字面值。

- [ ] **Step 3: 驗證**

```bash
cd frontend && pnpm typecheck && pnpm lint && pnpm test && pnpm build
```

- [ ] **Step 4: 視覺比對（允許開瀏覽器，收尾必須乾淨）**

5 頁 × {桌面, 手機} × {淺色, 深色} 對照，記錄任何非預期差異。做法：dev server 用 `hub` 起，截圖／computed-style 用 playwright MCP 或 `chrome-headless-shell`。
**收尾衛生（強制）**：任務結束前 `pkill -f chrome-headless-shell` 並確認 `pgrep -f chrome-headless-shell | wc -l` 為 0、刪掉你建的暫存 profile、不留暫存檔在 repo。若瀏覽器反覆崩潰（本機 Chrome 153 有此紀錄），降級為靜態佐證並在報告明寫未做像素級驗證，不要硬撐重試。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/features
git commit -m "refactor(frontend): 5 頁改用 Tailkit 版面與語意 token"
```

---

### Task 10: Nikala 清理與端到端驗收

**Files:**
- Delete: `frontend/nikala.config.json`、`frontend/.cursorrules`、`frontend/.cursor/rules/nikala.mdc`（並移除空的 `.cursor/rules/`）
- Modify: `frontend/AGENTS.md`

- [ ] **Step 1: 刪除 Nikala 痕跡並改 `AGENTS.md`**

標題改為 `# Frontend（SolidJS）Development Guidelines`；刪除 Nikala 引言與 ThemeScript 條款；保留 splitProps／`children()`／SSR guard 三條；新增：UI 元件庫為 **Ark UI（行為）× Tailkit（視覺）**，互動元件一律以 Ark primitives 實作於 `src/components/ui/`，Tailkit 僅作為結構與樣式來源（非 npm 依賴），顏色一律語意 token。

- [ ] **Step 2: 移除 Kobalte 依賴**

```bash
cd frontend && pnpm remove @kobalte/core
```

- [ ] **Step 3: 殘留檢查**

用 `grep` 工具（本 harness 封鎖 shell grep）：
- `kobalte`（path：`frontend/src; frontend/package.json; pnpm-lock.yaml`）→ 0 命中
- `Nikala`（path：`frontend`）→ 0 命中
- **Tailkit 色階字面值**：搜 class 形式的色階前綴（path：`frontend/src`）→ 0 命中：
  `(bg|text|border|ring|divide|placeholder|from|to|via|shadow)-(secondary|orange|emerald|zinc|slate)-`
  （**不要**用裸字 `secondary-`：`src/index.css` 的 `--color-secondary`／`--secondary-foreground` 等 token 名稱是合法的，會誤命中。）

- [ ] **Step 4: 全套驗收**

```bash
cd frontend && pnpm typecheck && pnpm lint && pnpm test && pnpm build
```

Expected: 全綠、**0 warning**。

- [ ] **Step 5: 瀏覽器實測（允許，收尾必須乾淨）**

1. `/login`：兩 tab 切換（含鍵盤 ArrowLeft/ArrowRight）、Google 按鈕、店家登入表單送出錯誤路徑
2. 公司／部門：modal 開關、Escape、點 backdrop、表單驗證、CRUD 成功後表格更新
3. 角色權限：矩陣 checkbox 勾選、disabled 格、儲存
4. shell：側邊欄選中態、佔位項不可點、手機寬度收合
5. 深色模式：以上全部再看一次（`document.documentElement.classList.add('dark')`）

**收尾衛生（強制）**：`pkill -f chrome-headless-shell` 後確認 `pgrep -f chrome-headless-shell | wc -l` 為 0、刪除暫存 profile、不留暫存檔。若瀏覽器反覆崩潰（本機 Chrome 153 有此紀錄），改做靜態佐證並明列未驗項，不要硬撐。

- [ ] **Step 6: impeccable detector**

用 `impeccable detect --json` 掃改動後的 UI 檔案；有發現則修，或在報告中說明保留原因。

- [ ] **Step 7: Commit**

```bash
git add -A frontend/nikala.config.json frontend/.cursorrules frontend/.cursor frontend/AGENTS.md frontend/package.json pnpm-lock.yaml
git commit -m "chore(frontend): 移除 Kobalte 與 Nikala 痕跡，文件改以 Ark UI × Tailkit 為準"
```

---

## 驗收對照（spec §10）

| spec 驗收 | 步驟 |
|---|---|
| 1. typecheck/lint/test/build 全綠且 lint 0 warning | Task 10 Step 4（各 task 各自也跑） |
| 2. dev server 目視 5 個互動面 | Task 10 Step 5（允許開瀏覽器；崩潰時可降級為靜態佐證並明列未驗項） |
| 3. 改動前後視覺比對 | Task 9 Step 4（同上；收尾衛生強制） |
| 4. `kobalte`/`Nikala`/色階字面值歸零 | Task 10 Step 3 |
| 5. impeccable detector | Task 10 Step 6 |

---

*建立：2026-09-19（v2）*
