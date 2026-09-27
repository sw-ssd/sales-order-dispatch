---
name: 多公司訂出貨系統 — Web 中台
description: 出貨作業的控制室：平整、克制、以邊框定義形狀的高密度桌機工具
colors:
  background: "oklch(1 0 0)"
  foreground: "oklch(0.1450 0 0)"
  card: "oklch(1 0 0)"
  card-foreground: "oklch(0.1450 0 0)"
  popover: "oklch(1 0 0)"
  primary: "oklch(0.53 0.18 70)"
  primary-foreground: "oklch(0.9850 0 0)"
  secondary: "oklch(0.9700 0 0)"
  muted: "oklch(0.9700 0 0)"
  muted-foreground: "oklch(0.53 0 0)"
  accent: "oklch(0.9700 0 0)"
  destructive: "oklch(0.505 0.213 27.5)"
  destructive-foreground: "oklch(1 0 0)"
  success: "oklch(0.48 0.13 155)"
  success-foreground: "oklch(0.985 0 0)"
  warning: "oklch(0.50 0.13 95)"
  warning-foreground: "oklch(0.985 0 0)"
  info: "oklch(0.50 0.13 250)"
  border: "oklch(0.9220 0 0)"
  grid-line: "oklch(0.9220 0 0 / 0.75)"
  input: "oklch(0.9220 0 0)"
  ring: "oklch(0.58 0 0)"
  sidebar: "oklch(0.9850 0 0)"
  sidebar-foreground: "oklch(0.1450 0 0)"
  sidebar-accent: "oklch(0.9700 0 0)"
  sidebar-border: "oklch(0.9220 0 0)"
typography:
  display:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 700
  title:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "1.125rem"
    fontWeight: 600
    lineHeight: 1
  body:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
  label:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 500
  caption:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 600
    letterSpacing: "0.05em"
rounded:
  sm: "0.375rem"
  md: "0.5rem"
  lg: "0.625rem"
  xl: "0.875rem"
spacing:
  xs: "0.25rem"
  sm: "0.5rem"
  md: "0.75rem"
  lg: "1rem"
  xl: "1.25rem"
  2xl: "1.5rem"
components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-foreground}"
    rounded: "{rounded.lg}"
    padding: "0.5rem 1rem"
    height: "2.25rem"
    typography: "body"
  button-primary-hover:
    backgroundColor: "oklch(0.53 0.18 70 / 0.95)"
    textColor: "{colors.primary-foreground}"
    rounded: "{rounded.lg}"
  button-destructive:
    backgroundColor: "{colors.destructive}"
    textColor: "{colors.destructive-foreground}"
    rounded: "{rounded.lg}"
    height: "2.25rem"
  button-outline:
    backgroundColor: "{colors.card}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.lg}"
    height: "2.25rem"
  button-ghost:
    backgroundColor: "transparent"
    textColor: "{colors.foreground}"
    rounded: "{rounded.lg}"
    height: "2.25rem"
  input:
    backgroundColor: "{colors.card}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.lg}"
    padding: "0.5rem 0.75rem"
    height: "2.5rem"
  badge:
    backgroundColor: "{colors.muted}"
    textColor: "{colors.muted-foreground}"
    rounded: "{rounded.sm}"
    padding: "0.25rem 0.5rem"
  card:
    backgroundColor: "{colors.card}"
    textColor: "{colors.card-foreground}"
    rounded: "{rounded.lg}"
    padding: "1.25rem"
  table-head:
    backgroundColor: "{colors.muted}"
    textColor: "{colors.foreground}"
    padding: "1rem 0.75rem"
  table-cell:
    textColor: "{colors.foreground}"
    padding: "0.75rem"
---

# Design System: 多公司訂出貨系統 — Web 中台

## Overview

**Creative North Star: "The Control Room"（控制室）**

中台是一間控制室：多路訊號（訂單、派車、庫別、退貨）同時進場，使用者的工作是**判斷**，不是瀏覽。因此畫面的第一責任是讓狀態可掃視、異常可定位，其次才是任何形式的愉悅。裝飾性的漸層、插圖、動效在這裡都是雜訊；節奏來自表格的行高、標題的層級與留白的一致性。（唯一的例外是 `body` 的圖紙格線——它是底材，不是內容：不搶對比、不進資料區，見 Layout 的底層與 The Paper-Under-Glass Rule。）

系統自帶一套完整語意 token（oklch，淺／深兩套），元件只消費 token，不持有顏色。這規則讓深色模式是同一套值的翻轉，而不是第二套設計——也讓平台營運主控台能直接以路徑別名共用同一份樣式而不分裂。

形狀語言是**平整且克制的**：1px 邊框定義形狀，圓角 10px（0.625rem）統一，靜止狀態只用最低一階的陰影（`--shadow-xs`，`0 1px 3px hsl(0 0% 0% / 0.05)`）或完全不用。深度靠層級與底色差（`bg-muted` 帶狀區）表達，不靠投影堆疊。

**Key Characteristics:**
- 高資訊密度：表格列高約 45px、表頭 52px，內容區無 `max-width`，寬度全給資料
- 單一強調色：托盤陶土（Pallet Clay）只出現在主要動作、選中狀態與焦點環
- 邊框優先：卡片、輸入框、表格、側欄都以 1px `--border` 分隔
- 唯一的裝飾是 body 的 24px 圖紙格線（`--grid-line`）：它是「紙」，不是內容——所有承載內容的表面都不透明地壓在它上面
- 深色是同一組 token 的覆寫（`.dark`），元件內禁止顏色的 `dark:` 變體
- 狀態一律有文字，顏色只是加速判讀

## Colors

調色盤是**單一強調色的中性系統**：整個介面由近黑／近白的無彩灰階構成，唯一的彩度來自暖橙主色與四個語意色（成功／警告／資訊／危險）。

### Primary
- **托盤陶土（Pallet Clay）** (`oklch(0.53 0.18 70)`): 唯一的主要動作色。用於 `Button` default variant 的底、`Badge` default、`Input` 焦點邊框、行內動作連結（`text-primary hover:underline`）、`ScrollArea` 焦點環、側欄焦點環、圖表第一色。亮度被刻意壓到 L 0.53：白字疊上去才有 4.5:1。

### Semantic（語意色，非裝飾）
- **出貨綠（success）** (`oklch(0.48 0.13 155)`): 已完成、已核准、成功提示帶（`bg-success/15`）。
- **備料黃（warning）** (`oklch(0.50 0.13 95)`): 處理中／待審核／帳號待啟用、接近額度上限。
- **調度藍（info）** (`oklch(0.50 0.13 250)`): 待處理（訂單專用）。**注意**：全站唯一把 `pending` 配藍的是訂單狀態（`OrdersPage.tsx` 的 `STATUS_VARIANTS`）；退貨與使用者頁的 `pending` 配黃。兩種都反映現行實作，改動任一端前先確認另一端。
- **退貨紅（destructive）** (`oklch(0.505 0.213 27.5)`): 已作廢、刪除、錯誤橫幅（`bg-destructive/15`）。
- **中性灰（secondary／muted）**: **終態與靜止態**一律中性——`cancelled`（訂單）、`inactive`（使用者）、`read`（通知）。綠代表「這件事有個好結果」，灰代表「這件事結束了、不需要再看」；把後者塗綠會讓列表整片發綠，稀釋真正該注意的成功訊號。

### Neutral
- **淨白（background / card / popover）** (`oklch(1 0 0)`): 頁面底、卡片底、浮層底。三者同值，靠邊框與陰影分層；差別在**頁面底疊了格線**，卡片與浮層維持不透明的白，格線只從間距裡透出來。
- **圖紙格線（grid-line）** (`oklch(0.9220 0 0 / 0.75)`): 頁面底紋專用，值刻意取 `--border` 疊 75% 透明度——畫面上的**實線**（表格列、卡片外框、側欄分隔）必須永遠比底紋強，否則資料格線與裝飾格線會打架。只鋪在 `body`，不得挪去畫分隔線。深色沿用同一條 75% 推導（`--border` 在 `.dark` 較亮），所以深色下會比淺色顯眼一些——這是同一條規則的結果，不是兩套值，不要為了「看起來一樣」去改其中一邊。
- **墨黑（foreground / card-foreground）** (`oklch(0.1450 0 0)`): 主要文字，全站唯一的主文色。
- **灰面板（muted / secondary / accent）** (`oklch(0.9700 0 0)`): 表頭色帶、卡片標題帶、篩選列、hover 底、`Badge` secondary。
- **說明灰（muted-foreground）** (`oklch(0.53 0 0)`): 說明文字、次要標籤、表格空狀態。與 `muted` 底的對比依 4.5:1 定案。
- **格線灰（border / input）** (`oklch(0.9220 0 0)`): 所有 1px 分隔線、輸入框邊框、卡片外框。
- **側欄面板（sidebar / sidebar-accent / sidebar-border）** (`oklch(0.9850 0 0)`): 導覽專用的一組，值略高於 `background`，讓側欄與內容區在無陰影下仍可分辨。

深色不另立值：同一組 token 在 `.dark` 覆寫（例如 primary → `oklch(0.75 0.16 70)`），實際值見 `.impeccable/design.json` 的 `colorMeta.*.darkValue`。

### Named Rules
**The One Accent Rule.** 托盤陶土在任一畫面的覆蓋面積不超過一成。它的稀有性就是它的作用——主色一多，主要動作就不再可辨。

**The Token-Only Rule.** 元件內只准寫語意 token（`bg-card`、`text-muted-foreground`、`border-border`）。禁寫色階字面值（`zinc-*`、`orange-*`、`emerald-*`），也禁寫顏色相關的 `dark:` 變體——深色由 `.dark` 覆寫同一組 token 達成。

**The Contrast-Carrying Rule.** 任何改動 token 亮度的人都必須重新量測：白字疊主色的 hover 用 `/95` 而非 `/90`，因為 90% 只有 4.36:1（不合格）、95% 是 4.75:1。焦點環 `--ring` 也因同樣理由定在不透明 L 0.58（對白底 4.29:1）。

**The Business-Timezone Rule.** 任何**瞬間**（`created_at`、`printed_at`、`trial_ends_at`…）都必須經 `lib/datetime.ts` 的 `formatDateTime`／`formatDate` 轉成營業時區（UTC+8）再顯示，**禁止** `slice(0, 19)` 這類字串裁切。理由是可量測的：後端 5 個服務輸出 `.UTC()`、15 個輸出伺服器本地偏移，而字串裁切會把偏移整個丟掉——同一個入庫瞬間在訂單頁與客戶頁因此相差八小時，且該差距取決於部署機器的 `TZ`（換機器就變）。純日曆日欄位（`expected_delivery_date`、`target_date`）是日期不是瞬間，原樣顯示不轉換。

## Typography

**Display Font:** Inter（經 fonts.bunny.net 載入，權重 300–900）
**Body Font:** 同一套 Inter + `ui-sans-serif, system-ui` 系統堆疊；中文由系統 CJK 字型承接（`Noto Sans` 在堆疊內）
**Label/Mono Font:** `--font-mono`（`ui-monospace, SFMono-Regular, Menlo`）——只用於分頁摘要與 demo 頁的 token 名稱，不進正式版面

**Character:** 單一字族、單一語氣：這是工具，不是刊物。層級完全靠字級與字重（1.5rem/700 → 1.125rem/600 → 0.875rem/400）建立，不使用襯線、不用斜體、不用字距裝飾（唯一例外是側欄分組標籤的 `tracking-wider` 大寫）。

### Hierarchy
- **Display**（700, 1.5rem/2rem）: 頁面 `h1`。全頁唯一，出現在頁首左側。
- **Title**（600, 1.125rem, `leading-none`）: 卡片標題（`CardTitle`，實作為 `<h3>`）。
- **Body**（400, 0.875rem）: 表格內容、篩選標籤、說明文。表格允許 `whitespace-nowrap`，靠橫向捲動而非換行保密度。
- **Label**（500, 0.875rem）: 表單標籤（`FieldLabel`）、按鈕文字（`font-semibold` 0.875rem）。
- **Caption**（600, 0.75rem, `tracking-wider`, 大寫）: 側欄分組標籤、`Badge` 文字、表格內次要資訊。

### Named Rules
**The Two-Weights Rule.** 每一段文字只用 400 與 600 兩級（標題 700 是唯一例外）。連續三種字重會讓密集表格看起來像在喊叫。

**The Monospace-Is-Evidence Rule.** 等寬字只用於需要逐位對齊的數字（分頁摘要、編號），不用來營造技術感。

## Layout

**空間模型是「側欄 + 單一內容柱」，內容柱沒有 `max-width`。**

- 外框：`flex h-dvh w-full min-w-80 overflow-hidden`（**不設** `bg-background`，見下）；桌機是 Ark Splitter 兩面板（側欄 12–30rem 可拖、內容最小 20rem）。
- 底層：格線鋪在 `body` 上（`background-size: 24px 24px`，兩條 1px `linear-gradient`），隨 viewport canvas 固定、不跟著內容捲動——內容是從紙上滑過去的。因此**殼層、`SidebarInset` 與頁面 `main` 都不得自帶不透明底色**：加了就把底紋蓋掉，整片變成死白。側欄（`bg-sidebar`）、頂列（`bg-card`）、卡片與表格仍是不透明表面，格線只出現在它們之間的間距。
- 側欄寬：展開 **16rem**、收斂成 icon rail **3rem**、行動抽屜 **18rem**；寬度以 CSS 變數 `--sidebar-width*` 注入，過場 `200ms ease-linear`。
- 頂列：`sticky top-0 z-30 h-16`，`bg-card` + 下緣 1px 邊框；標題是 `<p>`（頁面自己持有唯一 `h1`）。
- 內容柱：`h-dvh overflow-y-auto`，**外框是唯一的 padding 擁有者**（`p-4 lg:p-6`），頁面本身不加上下留白。
- 分頁面版型順序（固定）：頁首 → 錯誤橫幅（可選）→ 一張 `Card`（內含篩選列 → `Table` → 分頁列）→ 對話框（在卡片外）。
- 篩選列：`flex flex-wrap items-end gap-3 border-b border-border bg-muted px-3 py-3`；條件多時改用 `grid gap-3 sm:grid-cols-2 lg:grid-cols-4`。
- 分頁列：`flex flex-wrap items-center justify-between gap-3 border-t border-border px-3 py-3`。
- 斷點是 **JS**（`matchMedia("(max-width: 767px)")`），不是 class：窄於 768px 時 Splitter 整個不掛載，側欄改為 Ark drawer（backdrop `bg-foreground/75`、`z-50`），導覽切換後自動關閉。

密度基準（實測自現行 class）：

| 位置 | 值 |
|---|---|
| 表格列高 | ≈45px（`p-3` + 20px 行高 + 1px 框） |
| 表頭高 | ≈52px（`py-4`） |
| 篩選控件高 | 輸入框 40px、按鈕 36px、checkbox 列 36px |
| 卡片內距 | 標題帶／頁腳 `px-5 py-4`，內容 `p-5` |
| 側欄項目 | 列高 ≈32px（`py-1.5`），圖示 16px |

### Named Rules
**The Full-Bleed Content Rule.** 內容柱不設 `max-width`。資料是主體，把它收成 1200px 的置中欄只是為了美觀而犧牲掃視寬度。

**The Shell-Owns-Padding Rule.** 頁面不得自帶外距；間距只由外框的 `p-4 lg:p-6` 與元件自身的內距產生。

**The Paper-Under-Glass Rule.** 底紋只在間距裡出現。任何承載內容的表面——卡片、表格、篩選列、頂列、側欄、輸入框、對話框——都必須是不透明底色；格線若穿過某個內容區塊，那個區塊就是漏了一層底。反過來說，殼層與頁面不得再補 `bg-background`：紙已經在 body 上了。

## Elevation & Depth

**預設是平的。** 靜止狀態只有兩階：完全不投影，或最低一階 `--shadow-xs`（`0 1px 3px hsl(0 0% 0% / 0.05)`，用在卡片與 hover 的 outline／secondary 按鈕）。層級主要靠**底色差**（`bg-muted` 色帶：卡片標題、表頭、篩選列）與 **z-index + 邊框**（頂列 `z-30`、對話框 `z-50`）表達，不靠投影堆疊。

對話框與抽屜用背景遮罩建立深度：`fixed inset-0 z-50 bg-foreground/75`（＋預設 `backdrop-blur-sm`）。這是全系統唯一允許的強制變暗手法。

### Named Rules
**The Flat-By-Default Rule.** 表面靜止時不投影。陰影只作為狀態反應（hover 的 `shadow-xs`）或最低限度的卡片分界。

**The No-Shadow-Stacking Rule.** 同一畫面不得同時出現兩階以上的陰影。要再分層就用底色或邊框。

## Shapes

圓角統一由 `--radius: 0.625rem` 推導，**只使用四階**：`sm` 6px（`Badge`、檔案選擇鈕）、`md` 8px（側欄項目、分頁籤）、`lg` 10px（按鈕、輸入框、卡片、對話框、頁籤容器）、`xl` 14px（保留，目前未使用）。邊框一律 1px `--border`（hover 或 focus 時可轉為主色）。圓形只出現在頭像與開關類元件，不用於容器。

整體輪廓是**矩形為主、小圓角收邊**：沒有膠囊形大按鈕、沒有超橢圓、沒有裝飾性斜角。清單與表格是全寬貼齊的矩形，讓欄位邊緣能對齊成一條線。

### Named Rules
**The 10px Rule.** 容器與控件都用 `rounded-lg`（10px）；小元件降一階到 6–8px。沒有第五種圓角。

## Components

元件分三層：Ark UI 提供行為（狀態機、焦點、ARIA），元件本身定義語意 token 的組裝，頁面只負責資料。**元件對外只暴露 `data-state`／`data-*` 屬性，呼叫端不得自己刻 ARIA。**

### Buttons
出貨作業的動詞：動作明確、尺寸一致、次要動作用邊框或純文字。
- **Shape:** `rounded-lg`（10px）+ 1px 邊框（`ghost`／`link` 為透明框），字重 600、0.875rem，圖示 16px。
- **Sizes:** `xs` 24px（表格內密集動作）、`sm` 32px、`default` 36px、`lg` 40px、`icon` 36×36。
- **Primary（default）:** 托盤陶土底 + 近白字，hover 為同色 95%（不是 90%，見對比規則）。
- **Destructive:** 退貨紅底 + 白字。
- **Outline / Secondary:** 白底 + 格線框，hover 換 `bg-muted` 並加 `shadow-xs`；兩者的差別只在底色（`card` vs `secondary`），選用時依「這是不是動作」而非「哪個好看」。
- **Ghost / Link:** 無框；`link` 只用於極次要的導覽（如「查看全部」）。
- **狀態:** focus-visible 一律 3px `ring-ring`（只有 Checkbox 是 1px、Badge 的 2px 環因根節點不可聚焦而實際不顯示）；disabled 為 `opacity-50` + `pointer-events-none`；loading 會同時 `disabled` 並掛 `aria-busy="true"` 與內嵌 Spinner。

### Cards / Containers
資訊的容器，也是唯一的分層工具。
- **Corner Style:** `rounded-lg`（10px），`overflow-hidden`。
- **Background:** `bg-card`；標題帶與頁腳用 `bg-muted` 形成頂／底色帶。
- **Border:** 1px `--border`。
- **Shadow Strategy:** `shadow-xs`，見 Elevation。
- **Internal Padding:** 標題帶／頁腳 `px-5 py-4`，內容 `p-5`。
- 標題硬綁 `<h3>`：外層若已是 `h1`，中間必須補一個 `h2`（或改用 section 標題），否則跳級。

### Inputs / Fields
- **Style:** 白底（`bg-card`，不是透明）+ 1px 格線框 + `rounded-lg`，高 40px，內距 `px-3 py-2`；桌機字級降到 0.875rem（`md:text-sm`），行動維持 1rem 以避免 iOS 自動縮放。
- **Focus:** 邊框轉主色 + 3px `ring-primary/50`；錯誤態改為 `border-destructive` + `ring-destructive/50`。
- **Labels:** 標籤／說明／錯誤一律走 `Field`（`grid gap-1.5`），標籤 0.875rem/500，說明與錯誤 0.875rem、錯誤用 `text-destructive` 且 `role="alert"`。
- **深色陷阱:** `@tailwindcss/forms` 對原生控件硬編白底且無深色變體，`select`／`textarea`／`input[type=date|...]` 與 checkbox/radio 由 `index.css` 的 `@layer components` 統一回填語意色；新控件若繞過元件層，要確認仍落在這條覆蓋內。
- **必須顯式傳 `invalid`：** `Field` 不會從子元件推斷錯誤（Solid 無子節點偵測）；不傳就沒有 `aria-invalid` 也沒有錯誤訊息。

### Tables
中台最核心的元件，密度與對齊即是它的設計。
- **Style:** 外層 `overflow-x-auto`，`<table class="min-w-full align-middle text-sm whitespace-nowrap">`——**不換行**，長內容靠橫向捲動。
- **Head:** `bg-muted px-3 py-4 text-left font-semibold`，下緣 1px 框；色帶是唯一的表頭裝飾。
- **Row:** `border-b border-border`，hover 為 `bg-muted/50`（半透明，不破壞選中態），選中態 `bg-muted`。
- **Cell:** `p-3 align-middle`。
- **行內動作:** 純文字連結（`font-medium text-primary hover:underline`；危險動作用 `text-destructive`），容器 `flex justify-end gap-3`——不放大按鈕，避免撐高列。
- **命中區:** 行內文字連結的實際命中框只有 28×20，低於 WCAG 2.5.8（AA）的 24×24。以 `::after` 覆蓋層**只擴垂直**（水平方向相鄰連結相距 12px，橫擴會互相重疊）：滑鼠環境上下各 2px（→24px），`@media (pointer: coarse)` 擴到 44px。列高與欄寬不變。**不分指標類型**——游標精度受限與螢幕放大同樣受影響，不是觸控專屬問題。

### Badges
靜態狀態標籤，唯讀、無互動。
- **Style:** `rounded-sm`（6px）、`px-2 py-1`、0.75rem/600；語意色一律「淡底 + 同色深字」（`bg-success/15 text-success`、`bg-destructive/15 text-destructive`），只有 default 用實底主色。
- **State:** 無 hover、無 disabled。因為不可聚焦，它身上的 `ring-2` 焦點樣式實際不會出現——不要把 Badge 當按鈕用。

### Navigation（Sidebar）
導覽是控制室的儀表板列：固定、可收斂、位置永恆。
- **Root:** `bg-sidebar text-sidebar-foreground` + 右緣 1px 框，寬度由 CSS 變數驅動。
- **Structure:** 4 個分組 / 19 個項目。分組標籤 `text-xs font-semibold tracking-wider uppercase text-sidebar-foreground/70`；`SidebarContent` 是全站**唯一**的 `<nav>` landmark（品牌列刻意在它之外），因此 `aria-current="page"` 不會重複。
- **Item:** `rounded-md px-2 py-1.5`（≈32px 列高）、圖示 `size-4`。未選中為 `text-sidebar-foreground`，hover 轉 `bg-sidebar-accent`，**選中由 `data-active="true"` 驅動**（`bg-sidebar-accent` + `font-semibold` + `text-sidebar-accent-foreground`）並同步 `aria-current="page"`；比對是精確相等，不做前綴匹配。
- **Collapsed rail:** 項目置中、標籤 `sr-only`，hover 以右側 Tooltip 補回名稱。
- **Placeholder:** 有圖示與標籤但尚未有路由的項目渲染為 `aria-disabled="true"` 的 span（`text-muted-foreground/70`、`cursor-not-allowed`），不以連結偽裝。

### Dialog
- **Style:** 遮罩 `bg-foreground/75` + `backdrop-blur-sm`；內容白底、`rounded-lg`、`max-w-lg`（32rem，寬表單另給 `max-w-3xl`），內文包在 `ScrollArea max-h-[80vh]` + `space-y-4 p-6`。
- **Behavior:** 關閉鈕固定右上（36×36、3px 焦點環）；`DialogContent` 關閉時整棵子樹卸載——**表單欄位狀態必須放在對話框內容的子元件裡**，否則下次開啟會殘留上一筆的值。

### Signature: PlanBanner
內容柱頂列與頁面之間的一條權益／訂閱狀態橫幅。它不是通知、也不是錯誤：把「這家公司目前的方案與到期狀態」放在每個畫面的固定位置，讓權益問題在使用者撞到 PLAT-5001 之前就被看見。

## Do's and Don'ts

### Do:
- **Do** 只寫語意 token（`bg-card`、`text-muted-foreground`、`border-border`），並在深色模式驗證一次。
- **Do** 把顏色改動當成對比度改動：任何 `--primary`／`--ring`／`--muted-foreground` 的亮度調整都要實測 4.5:1（文字）與 3:1（非文字）。
- **Do** 讓狀態有文字。顏色只是加速判讀，`Badge` 的語意色永遠搭配同義的繁中字。
- **Do** 用 `data-state`／`data-active` 屬性表達狀態，讓元件與頁面各司其職。
- **Do** 在表格內把列動作寫成文字連結（`text-primary hover:underline`），維持 45px 列高的密度。
- **Do** 讓頁面沿用固定順序：頁首 → 錯誤帶 → 一張卡片（篩選／表格／分頁）→ 對話框。
- **Do** 把表單放進對話框時，同時傳 `invalid`，否則錯誤訊息與 `aria-invalid` 都不會出現。

### Don't:
- **Don't** 寫色階字面值（`zinc-*`、`orange-*`、`emerald-*`）或顏色相關的 `dark:` 變體。
- **Don't** 在內容柱上追加 `max-width`／置中容器——資料寬度就是版型寬度。
- **Don't** 在殼層（`AppShell`／`SidebarInset`）或頁面 `main` 補 `bg-background`——那會蓋掉 `body` 的圖紙格線，讓底層變成死白；承載內容的表面才需要不透明底色。
- **Don't** 疊加陰影；也不要為了「有層次」給靜止的卡片加 `shadow-md` 以上。
- **Don't** 引入第五種圓角或膠囊形大按鈕。
- **Don't** 自己刻 ARIA 或 `role`（Ark 已提供），也不要在元件外層改寫它的 `data-*` 契約。
- **Don't** 把 `Badge` 當可點元件（它不可聚焦），也別把 `Card` 當可點區塊（要就放真正的 `<a>`／`<button>`）。
- **Don't** 在元件內硬編訊息字串；錯誤一律走 `ErrorInfo` 的碼與已渲染訊息。
- **Don't** 用動畫補足層級——過場只用在狀態變化（`transition-colors`）與必要的位移（側欄 200ms），進入動畫一律不做。
