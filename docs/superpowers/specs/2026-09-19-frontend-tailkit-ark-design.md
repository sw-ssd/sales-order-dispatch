# 前端 UI 重構：Ark UI（行為）× Tailkit（視覺）— 設計（spec v2）

> **狀態**：設計已定案（2026-09-19），待實作計畫
> **範圍**：`frontend/`（SolidJS Web 中台）
> **前一版**：`attempt-1`（Ark 單獨遷移 + 保留本地包裝）已作廢並重置，原因見 §0

---

## 0. 為什麼重來（attempt-1 的根因）

attempt-1 的四個 commit 已由 `git reset --keep 7d1ac4b` 移除（其 spec/plan/ledger 原存於 `.superpowers/sdd/…ark-ui-migration-plan/attempt-1/`；該工作區已於專案收尾時刪除，故此處僅存歷史敘述——**下方表格即為其結論的完整留存**）。作廢的原因不是程式碼品質（dialog 改寫通過複審），而是**設計層的兩個決策逼出補丁**：

| 根因 | 症狀 |
|---|---|
| 沿用 attempt-1 自訂的對外 API（`closeOnOutsideClick` 放在 `DialogContent`），而 Ark 只接受在 `Dialog.Root` | 需要 `Dialog` → `DialogContent` 的 context 上呈；且該 prop 實際上沒有任何呼叫端 |
| 測試策略未先定（禁止 mock／禁止斷言內部屬性），而 Ark 的 `send` 走 `queueMicrotask`、鍵盤選取走兩段 rAF | Tabs 測試攔截 rAF 逐 frame 推進以「還原瀏覽器語意」，屬模擬而非真實行為 |
| 只換元件庫，未定視覺權威 | 元件是 Kobalte/Nikala 血統、Pixso 稿是 blue/slate、`nikala.config.json` 是 zinc/amber、`index.css` 是 shadcn 風格 amber → 四套視覺敘述並存 |

v2 對應：對外 API 照 Ark 語意重新設計、測試策略先定且不模擬內部機制、視覺權威明確指定為 Tailkit。

## 1. 兩個權威，各管一件事

| 權威 | 來源 | 負責 |
|---|---|---|
| **Ark UI** (`@ark-ui/solid` 5.39.2) | 計畫 `2026-08-05-...-implementation-plan.md:222`（2026-08-31 決策） | **行為**：無障礙語意、狀態機、鍵盤操作、焦點管理 |
| **Tailkit 5.1** | 使用者指定（`tailkit.com`，已購買授權，MCP 已授權） | **視覺**：結構、尺寸、間距、字型、元件外觀、版型、頁面 |
| 本地複合元件 `src/components/ui/*` | 本設計 | 兩者的**唯一組合點**：頁面只認本地元件 |

Tailkit 官方僅提供 HTML / React / Vue / Alpine（**無 Solid**），且 HTML 版的互動靠 Alpine → 取回的 markup 一律手動移植成 SolidJS，互動層改由 Ark 承擔。Tailkit **不是** npm 依賴，不會進 bundle。

## 2. 決策摘要（訪談定案，2026-09-19）

1. **保留本地複合元件包裝層**，但全新重寫（不沿用 attempt-1 的實作）。
2. 對外 API **照 Ark 語意**重新設計，但不外洩 Ark 的 `details` 物件（`onValueChange(string)` 這類純值回呼）。
3. **顏色走語意 token**：Tailkit 的結構與尺寸照抄，顏色一律改寫為專案語意 token（`bg-card`、`text-muted-foreground`、`border-border`…），不用 Tailkit 的色票字面值。
4. **範圍一次到位**：theme/token + 13 個本地元件 + app shell + 現有 5 頁換皮。
5. 測試策略先定（§10）：只斷言使用者／輔助科技可觀察的行為，**不攔截 rAF、不 mock Ark 內部機制**；等待一律用真實的非同步等待。

## 3. 範圍

**In**
- `src/index.css`：保留語意 token 層，追加 Tailkit 官方 `@theme` 項目；深色模式 variant 對齊 Tailkit
- 新依賴：`@tailwindcss/forms`、`@tailwindcss/typography`（Tailkit 官方要求）
- 13 個本地元件重寫（互動 3 個 = Ark 行為 + Tailkit 結構；靜態 10 個 = Tailkit 結構）
- 新增 app shell（側邊欄 + 頂欄）
- 現有 5 頁換皮：登入、403、公司、部門、角色權限
- 刪除 Nikala 痕跡（`nikala.config.json`、`.cursorrules`、`.cursor/rules/nikala.mdc`、`AGENTS.md` 相關段落、元件註解字串）

**Out（本次不做）**
- Pixso 稿的 9 個未建 Web 畫面（shell 會為其保留位置，見 §7）
- 後端、proto、Flutter App（`app/`）任何改動
- CASL → OpenFGA 的 UI 權限來源遷移
- `scroll-area` 維持自刻（Tailkit 無對應元件；Ark 雖有 ScrollArea primitive，但它是「行為」層的替換，與本次視覺範圍無關）

## 4. 分層架構

```
頁面（features/*/pages/*.tsx）
   │  只認本地元件與語意 token
   ▼
src/components/ui/*            ← 組合點：Tailkit 結構 + 語意 token
   ├── 互動元件（dialog / tabs / checkbox）─ Ark primitives 提供行為
   └── 靜態元件（button / card / …）──────── 純 Tailwind markup
   ▼
@ark-ui/solid（行為）           Tailwind v4 + index.css token（視覺）
```

## 5. 依賴與設定變更

| 項目 | 變更 | 依據 |
|---|---|---|
| `package.json` | `+ @tailwindcss/forms`、`+ @tailwindcss/typography` | Tailkit 安裝文件：「某些 UI 元件需要這兩個外掛才能正確渲染」 |
| `src/index.css` | `@plugin "@tailwindcss/forms"`、`@plugin "@tailwindcss/typography"`；`@theme` 追加 `--default-font-family: "Inter"`、`--spacing-8xl/9xl/10xl`、`--animate-spin-slow` | Tailkit v4 設定文件 |
| 深色模式 | `@custom-variant dark (&:is(.dark *))` → `(&:where(.dark, .dark *))` | 對齊 Tailkit；差別是 `.dark` 自身也套用深色樣式 |
| 字型 | Inter 為預設；**採用 Tailkit 官方文件的 Bunny.net 載入方式**（`fonts.bunny.net`，GDPR 友善、無需自管資產）；Caveat 只在行銷版位需要時才載 | 現況：`index.html` 無任何字型載入、`--font-sans` 為系統字串 |
| 語意 token | **保留** `--background/--foreground/--primary/…`，不引入 Tailkit 色階 | §2 決策 3 |
| 狀態色 | 若 Tailkit 元件需要 success / warning（`--destructive` 只涵蓋錯誤），**擴充語意 token**（`--success` / `--warning` 與其 foreground），不引入 Tailkit 色階字面值 | 現行 token 集無成功／警告色，見 §6 狀態標籤元件 |

## 6. 元件對應表（13 個本地元件）

Tailkit identifier 格式：`a-*` = Application UI（`a-c-*` 元件、`a-l-*` 版型、`a-p-*` 頁面）。實際 id 於實作時以 MCP 檢索（範例：`a-c-modals-01`）。

| 本地元件 | Tailkit 來源類別 | Ark | 對外 API（v2） |
|---|---|---|---|
| `dialog.tsx` | Components → **Modals** | `Dialog` (`Root/Trigger/Backdrop/Positioner/Content/Title/Description/CloseTrigger`) | `Dialog(open, onOpenChange(open: boolean), closeOnOutsideClick?)`、`DialogContent(showCloseButton, blur, class)` |
| `tabs.tsx` | Components → **Tabs** | `Tabs` (`Root/List/Trigger/Content`) | `Tabs(value, defaultValue, onValueChange(value: string), orientation)` |
| `checkbox.tsx` | Components → **Form Elements** | `Checkbox` (`Root/Control/Indicator/HiddenInput`) | `Checkbox(checked, defaultChecked, onCheckedChange(checked: boolean), disabled)` |
| `button.tsx` | Components → **Buttons** | — | `Button(variant, size, loading, …)` 不變 |
| `card.tsx` | Components → **Cards** | — | `Card(…)` 不變 |
| `badge.tsx` | Components → **Badges** | — | `Badge(variant, …)` 不變 |
| `input.tsx` | Components → **Form Elements** | — | `Input(…)` 不變 |
| `label.tsx` | Components → **Form Elements** | — | `Label(…)` 不變 |
| `field.tsx` | Components → **Form Layouts** | — | `Field/FieldLabel(…)` 不變 |
| `table.tsx` | Components → **Tables** | — | `Table(…)` 不變 |
| `pagination.tsx` | Components → **Pagination** | — | `Pagination(…)` 不變 |
| `spinner.tsx` | 無直接對應（Tailkit 有 Progress Bars） | — | 不變；若 Tailkit 無合適樣式則保留現行視覺並改為語意 token |
| `scroll-area.tsx` | 無對應 | — | 不變（自刻保留） |

「不變」= 對外 props 與行為不變，僅內部 markup/樣式改為 Tailkit 結構 + 語意 token。

## 7. App shell

- 新增 `src/components/layout/`（側邊欄 + 頂欄），來源為 Tailkit **Layouts → Light Sidebar / Stacked**（依 Pixso 稿為淺色商務風）。
- 側邊欄項目依 Pixso 稿的 Web 導覽順序；**未實作的目標只渲染為 disabled 佔位**（不做假連結、不做假頁面），頁面落地時逐項啟用。
- shell 掛在既有 root route（`router/index.tsx` 的 `rootRoute` 目前直接渲染 `App`，而 `App.tsx` 只是 pass-through）。
- 登入頁與 403 頁**不套 shell**（Tailkit Pages → Sign In / Errors 為獨立版面）。

## 8. 頁面換皮（5 頁）

| 頁面 | Tailkit 來源 | 不動的部分 |
|---|---|---|
| `features/auth/pages/LoginPage.tsx` | Pages → **Sign In** | 員工／店家兩個 tab 的邏輯、Google OIDC、店家登入表單流程 |
| `features/auth/pages/ForbiddenPage.tsx` | Pages → **Errors** | 文案與導向 |
| `features/users/pages/CompaniesPage.tsx` | Layouts + Components（Tables/Modals/Buttons） | 資料載入、CRUD、`requireAbility` 守衛 |
| `features/users/pages/DepartmentsPage.tsx` | 同上 | 同上 |
| `features/users/pages/RolesPage.tsx` | Tables + **Form Switches/Checkboxes**（權限矩陣） | `PermissionMatrix` 的資料結構與 ability 邏輯 |

## 9. Tailkit 取碼與授權處理

- 取碼管道：Tailkit MCP（`https://tailkit.com/mcp`，OAuth 已完成，憑證存於 `~/.omp/agent/agent.db` 的 `mcp_oauth:profile:default:https://tailkit.com/mcp`）。以 `tech: "html"` 取結構與類別，互動語意一律改寫為 Ark。
- **調色盤缺口**：Tailkit markup 使用 `secondary-*` 等自訂色階，而公開文件與 MCP resource 都未定義該調色盤（推測在授權下載包內）。因此**不依賴**其色值：顏色一律映射為語意 token（§2 決策 3）；映射規則在實作時由實際類別歸納，並以視覺比對驗證。
- **授權邊界**：Tailkit 為付費授權產品。取回的 markup 僅作為改寫參考、不得整包轉散布；本專案只保留改寫後的 SolidJS 元件，不把 Tailkit 的原始檔案原樣入庫。
- 取碼工具在 kernel 內以代理工具實作（token 不落地、不進對話），subagent 以工具名取用。

## 10. 測試與驗收

**互動元件（dialog / tabs / checkbox）**：Vitest + jsdom + `@solidjs/testing-library`，只斷言可觀察行為：
- dialog：`open` 時標題可見；關閉鈕 → `onOpenChange(false)`；`showCloseButton={false}` 無關閉鈕
- tabs：受控 `value` 決定可見 panel；點擊 → `onValueChange` 收到新值且 panel 切換；`disabled` trigger 不切換；選中狀態以 `getByRole("tab", { selected: true })` 查詢
- checkbox：`checked` + `onCheckedChange(true/false)`；`defaultChecked` 非受控；`disabled` 不觸發
- **禁止**：攔截／手動推進 rAF、mock Ark 內部、斷言 Ark 私有屬性、以「不觸發」為唯一訊號的假綠案例
- 非同步等待一律用真實等待（`await waitFor(...)`／microtask tick），不使用假時鐘

**靜態元件**：不硬寫 class 斷言（plumbing）；以 §「驗收」的視覺比對與 lint 為準。

**驗收**（全部可執行）
1. `pnpm typecheck && pnpm lint && pnpm test && pnpm build` 全綠，且 **lint 0 warning**（attempt-1 留下過 `solid/reactivity` warning）
2. `pnpm dev` 目視：登入（兩個 tab、鍵盤操作）、403、公司／部門（modal 開關、Escape、外點、捲動）、角色權限（checkbox 勾選與 disabled）、shell（側邊欄未實作項為 disabled）
3. 改動前後對照：每個換皮頁面在桌面與手機寬度各截一次，確認無非預期視覺漂移
4. 殘留歸零：`kobalte`、`Nikala` 於 `frontend/` 與 lockfile 皆 0 命中；`nikala.config.json`、`.cursorrules`、`.cursor/rules/nikala.mdc` 已刪
5. impeccable detector 掃過改動後的 UI（有發現即修，或明確說明保留原因）

## 11. 風險

- **R1（最大）**：Tailkit 無 Solid 版，13 個元件 + shell + 5 頁全部手動移植；Alpine 互動全部要換成 Ark，工作量集中在這裡。
- **R2**：Tailkit 更新與我們的客製會漂移 → 以語意 token 吸收顏色變動；結構變動需人工 diff。
- **R3**：新增 `@tailwindcss/forms`（會重設表單元素預設樣式）與 `typography` → 既有表單與富文本樣式需逐一核對，可能出現非預期位移。
- **R4**：深色模式 variant 由 `:is(.dark *)` 改為 `:where(.dark, .dark *)`，影響面是全站；需在兩種模式下都比對。
- **R5**：字型（Inter）體積與載入來源（Bunny CDN vs 自架）影響首屏；自架需處理 CJK 子集化以外的拉丁字型授權（Inter 為 OFL，可自架）。
- **R6**：`spinner` 在 Tailkit 無直接對應 → 可能保留現行視覺，屬「未被 Tailkit 覆蓋」的清單，需在驗收時明列。

## 12. 後續（非本次）

Pixso 稿 9 個畫面的落地（每個畫面屆時以 `$impeccable shape` 走一次）、`scroll-area` 是否換 Ark ScrollArea、`spinner`/`empty state` 等 Tailkit 對應補齊、CASL → OpenFGA。

---

*建立：2026-09-19（v2；v1 見 attempt-1 存檔）*
