# Sidebar（多部件）

側邊欄。解剖取自 [solid-ui](https://github.com/stefan-karger/solid-ui) 的 sidebar，**行為層換成 Ark UI**：
行動版抽屜用 `drawer`、子選單用 `collapsible`、收合時的標籤用 `tooltip`、快捷鍵用 `hotkeys`。

本專案第一個「資料夾型」元件：`context.tsx` 放狀態與持久化，`parts.tsx` 放所有部件，本檔以外的
`sidebar.test.tsx` 守住三件事（收合切換＋記憶、行動版關閉時不可 Tab、`aria-current` 唯一）。

## 用法

```tsx
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarGroupContent,
  SidebarInset,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarProvider,
  SidebarTrigger,
} from "~/components/ui";
import { Link } from "@tanstack/solid-router";

<SidebarProvider>
  <Sidebar>
    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupLabel>營運</SidebarGroupLabel>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton as={Link} to="/users/companies" tooltip="客戶總表">
                <Users />
                <span>客戶總表</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>
  </Sidebar>
  <SidebarInset>{/* 頁面內容 */}</SidebarInset>
</SidebarProvider>;
```

`SidebarTrigger`（收合鈕）自己不放進側邊欄，由呼叫端擺在 Topbar 之類的位置。

## 部件 → Ark 對應

| 部件 | 底層 | 說明 |
| --- | --- | --- |
| `SidebarProvider` | `useHotkeys`（`hotkeys`） | 狀態來源與 `mod+B`；沒有 DOM |
| `Sidebar` | `drawer`（行動版分支） | `collapsible="none"`／行動寬度／桌面三分支 |
| `SidebarContent` | `<nav>` | 唯一的導覽 landmark |
| `SidebarHeader` / `SidebarFooter` | — | 品牌列／置底，**在 `<nav>` 之外** |
| `SidebarGroup` / `GroupLabel` / `GroupContent` | — | 分組 |
| `SidebarMenu` / `MenuItem` | `<ul>` / `<li>` | 清單語意 |
| `SidebarMenuButton` | `tooltip`（收合時） | polymorphic `as`，可接 TanStack `Link` |
| `SidebarMenuSub` / `MenuSubItem` | `collapsible` | 可收合子選單 |
| `SidebarTrigger` | — | 收合鈕（`Button` variant `ghost`），桌面切換／行動版開抽屜 |
| `SidebarRail` | — | 側欄內緣熱區，純 toggle、不進 tab 順序；**行動版不渲染** |
| `SidebarInset` | — | 主要內容區 |

行動版抽屜另外內建 `Drawer.CloseTrigger`（具名「關閉導覽選單」）：純觸控裝置才有關得掉的按鈕，
不必只靠點 scrim 或 Esc。

Ark `splitter` **不由本元件持有**：`Sidebar` 的 Panel 是「側欄／內容」兩個面板，屬於 AppShell（Task 6）
的組裝。`Splitter.ResizeTrigger` 不能脫離 `Splitter.Root` 使用，所以 `SidebarRail` 不做拖曳把手，
只提供同樣位置的滑鼠切換熱區。

## `data-*` 與 CSS 變數

桌面根層（`Sidebar` 的桌面分支）：`data-state="expanded|collapsed"`、`data-collapsible="icon"`（收合成
icon rail 時；展開或 `offcanvas` 時為空字串）、`data-variant="sidebar|floating|inset"`、
`data-side="left|right"`；行動版抽屜的 `Drawer.Content` 帶 `data-side`。

寬度一律由 CSS 變數決定，掛在桌面根層與行動版抽屜的 `Content` 上：

| 變數 | 值 | 用途 |
| --- | --- | --- |
| `--sidebar-width` | `16rem` | 桌面展開 |
| `--sidebar-width-icon` | `3rem` | 桌面收合成 icon rail |
| `--sidebar-width-mobile` | `18rem` | 行動版抽屜 |

導覽項目的選中狀態用 `data-active="true"`（`aria-current="page"` 同時由 `isActive` 產生）。
顏色一律 `--sidebar-*` 語意 token，沒有任何 `dark:` 變體。

## 狀態與持久化

- context 形狀：`{state, open, setOpen, isMobile, openMobile, setOpenMobile, toggleSidebar}`。
- `isMobile` 由 `matchMedia("(max-width: 767px)")` 驅動（`SIDEBAR_MOBILE_BREAKPOINT = 768`）。
- 桌面狀態持久化在 localStorage **`ui:sidebar`**（`"true"`／`"false"`）：
  - **讀**：`SidebarProvider` 初始化時讀一次（`createSignal(readStoredOpen())`）。
  - **寫**：`createEffect` 監看 `open()`，且**只在值真的變過之後**才寫——掛載時不覆寫既有值。
  - 兩者都包 `try/catch`：SSR（沒有 `window`）與 storage 被封鎖／配額用盡時，只失去記憶，
    不影響當次操作（回退「展開」）。
- `mod+B`（macOS ⌘B／其他平台 Ctrl+B）由 Ark `useHotkeys` 註冊，不手寫 keydown listener。

## 收合時的行為

三種 `collapsible`：

| 值 | 收合後 | 內容能否 Tab |
| --- | --- | --- |
| `icon`（預設） | 寬度 `3rem`，只剩圖示 | 可以（圖示仍是連結），標籤轉 tooltip |
| `offcanvas` | 寬度 `0px`，整欄收起 | **不可以**：根層加 `inert` |
| `none` | 不隨 `open` 改變寬度 | 可以 |

`offcanvas` 用 `inert` 而非 `hidden`：`hidden`（`display:none`）會讓 `transition-[width]` 的收合動畫
整段消失，`inert` 只切斷互動與無障礙樹。`inert` 以 `attr:` 設成**屬性**（不是 property），SSR 與
jsdom 才看得到。

`collapsible="icon"` 且桌面收合時，`SidebarMenuButton` 的直接子節點 `<span>` 會變 `sr-only`
（無障礙名稱保留），可見標籤改由 Ark `tooltip` 提供，位置在右側——因此**標籤文字請包在 `<span>` 裡**。
`SidebarMenuSub` 的觸發列同理。

## 兩個 v2 遺留的無障礙問題

1. **側欄關閉時仍可 Tab**：行動版交給 Ark `drawer`（關閉時 Presence 把 `Positioner`／`Content` 標成
   `hidden`，內容離開無障礙樹與 tab 順序），桌面 `collapsible="offcanvas"` 收合時加 `inert`；
   兩者都以「關閉後 `queryByRole("link")` 取不到／不在可 Tab 集合」守住，Esc 與抽屜內的關閉鈕都能關。
2. **同一個 `<nav>` 內兩個 `aria-current="page"`**：`<nav>` 只包在 `SidebarContent` 上，品牌列所在的
   `SidebarHeader` 在它之外；`aria-current="page"` 只由 `SidebarMenuButton` 的 `isActive` 產生。
   品牌列請不要放進 `SidebarContent`。

## 已知限制

- **`as={Link}` 會讓 TanStack typed routes 退化**：`SidebarMenuButton` 的 props 用
  `ComponentProps<As>` 取 `as` 指定元件的屬性，拿不到 `Link` 的泛型，因此 `to` 只檢查到 `string`
  ——`<SidebarMenuButton as={Link} to="/不存在的路由">` **不會**被 typecheck 擋下（一般
  `<Link to="…">` 會）。Task 6 接路由時請自行確認路徑存在。
- `variant="floating"` / `"inset"` 只有邊距、圓角與陰影差異，**尚無實戰頁面**驗證。
- `collapsible="none"` 不隨 `open` 改變寬度（寬度固定 `--sidebar-width`），也不回應 `SidebarTrigger`。
- `tooltip` 的定位固定 `placement="right"`：`side="right"` 的側欄應該顯示在左側，目前沒有把 `side`
  傳進 `SidebarMenuButton`。
- `SidebarRail` 只能滑鼠點（`tabindex="-1"`）；鍵盤請用 `SidebarTrigger` 或 `mod+B`。
- 子選單收合成 icon rail 時目前只把標籤收進 `sr-only`，沒有做浮出式次級選單。
