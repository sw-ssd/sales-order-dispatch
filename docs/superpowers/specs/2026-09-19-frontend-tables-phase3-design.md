# 前端表格 Phase 3 — TanStack Table ＋ 資料層（solid-query）＋ 伺服器端排序（設計）

> **狀態**：設計已定案（2026-09-19），待實作計畫
> **範圍**：`frontend/` 三張清單 ＋ `backend/` 的排序契約（跨前後端）
> **前置**：Phase 1（元件庫）、Phase 2（表單 + 欄位驗證）已完成並通過複審；後端 F1/F2（migration 與 OpenFGA）亦已修復，**後端現在可正常起、可登入、可拿到真權限**

---

## 0. 背景與偵查事實（皆有 檔案:行 證據）

| 事實 | 證據 |
|---|---|
| 真正用 `ui/Table` 渲染資料的只有 **2 張**：公司（6 欄）、部門（4 欄） | `CompaniesPage.tsx:288-353`、`DepartmentsPage.tsx:374-431` |
| 角色清單目前是 `<ul>`（不是表） | `RolesPage.tsx:175-204` |
| 角色權限矩陣是 80 格受控 checkbox（非排序/分頁表） | `PermissionMatrix.tsx:101-143` |
| 三張清單全部是**伺服器端分頁＋伺服器端篩選**（每次 RPC，無本地切片） | `CompaniesPage.tsx:166-171`、`DepartmentsPage.tsx:178-182`、`RolesPage.tsx:76` |
| **proto 對這三張表完全沒有排序欄位** | `company.proto:44-49`、`company.proto:88-92`、`role.proto:33-37` |
| 排序的既有先例：`string sort` 白名單、**只升冪**、預設欄位固定、非法值 `InvalidArgument` | `proto/customers/v1/customer.proto:63-64`、`customer_service.go:243-247`、`customer_service.go:241`（`ent.Asc`） |
| 後端目前的固定排序：公司/部門 `id DESC`、角色 `id ASC` | `company_service.go:119-121`、`:319-321`、`role_service.go:104-106` |
| `ui/table.tsx` 是純呈現（無排序/選取邏輯），子元件除了 `class` 全量 `{...rest}` 直通 | `ui/table.tsx` 全檔；`TableRow`/`TableHead` 已為選取態預留樣式 |
| `ListPagination`（Ark）已共用且有 7 條測試 | `ListPagination.tsx`（3 處使用）、`pagination.test.tsx` |
| **`@tanstack/solid-query@^5.102.8` 已在依賴，且已有使用先例** | `package.json`；`lib/ability/service.ts:14-22`（`queryOptions` + Connect client）、`lib/query-client.ts`、`lib/ability/guards.ts` |
| 共用 `transport` 已存在，但 5 個頁面仍各自 inline `createClient` | `lib/transport.ts`（只有 ability service 用）；`CompaniesPage:47-49`、`DepartmentsPage:50-56`(×2)、`RolesPage:15-18`、`LoginPage:23-25` |
| 表格頁**目前 0 條測試** | Phase 1/2 只建了元件與 modal 測試 |
| 重複量（**非** Table 的守備範圍）：`errorMessage` ×4（64 行）、清單 signal ×3（49 行）、`maxPage` 退回 ×3（15 行，逐字相同）、載入/空列（36 行）、篩選列（95 行）≈ 300 行/頁量級 | `CompaniesPage:73-90/102-115/175-180/300-313/254-286`、`DepartmentsPage:97-114/121-136/186-191/384-397/311-372`、`RolesPage:29-43/52-70/81-86/169-174` |

## 1. 目標與範圍

**In（三波）**
- **3A 資料層**：三張清單（公司／部門／角色）改用 `@tanstack/solid-query` 的 `queryOptions` + `createQuery`；`lib/transport.ts` 成為唯一 Connect 來源（移除 5 處 inline client）；mutation 後以 `invalidateQueries` 前綴失效。
- **3B 表格**：公司／部門改用 `@tanstack/solid-table`（manual 模式）；角色清單由 `<ul>` 改為表格（沿用 `ui/Table`），**保留「點選角色→載入權限矩陣」的互動與可及性**。
- **3C 排序（跨後端）**：proto 新增 `sort`（字串白名單）＋ `desc`（布林）；後端各服務加白名單映射（照 `customerSortField` 樣板）；前端點表頭切換排序。

**Out**
- `PermissionMatrix`（80 格受控勾選矩陣；導入 TanStack 只會多一層抽象）
- Phase 4（Pragmatic D&D）、`customers` 頁（前端尚無頁面；其 `sort` 是否補 `desc` 另議）
- 新增 devtools、任何新依賴（`solid-table` 除外）

## 2. 已定案的介面與裁定

**D1 排序契約（使用者裁定）**：`string sort` ＋ `bool desc`。
- `sort` 為**白名單欄位名**；**空字串 → 使用該服務的預設排序，且忽略 `desc`**（保證「前端不帶參數」＝現在的行為）。
- 非法值 → `InvalidArgument`，訊息列出白名單（照既有 `customerSortField` 的寫法）。
- 白名單與**預設值**（保留現行為）：
  | 表 | 白名單 | 空 `sort` 的預設 |
  |---|---|---|
  | 公司 | `name`、`identifier`、`tax_id`、`status`、`id` | `id DESC`（不變） |
  | 部門 | `name`、`id` | `id DESC`（不變） |
  | 角色 | `code`、`name`、`id` | `id ASC`（不變） |
- 前端**只對這些欄位**開可點表頭：公司 `name`/`identifier`/`tax_id`/`id`；部門 `name`/`id`；角色 `code`/`name`/`id`（`status` 可由後端排、但前端不開 UI：三值列舉的字典序排序價值低）。
- **改排序必須回到第 1 頁**（否則跨頁語意錯誤）。

**D2 TanStack Table 用 manual 模式**：`manualPagination: true` ＋ `rowCount: total()`，`getCoreRowModel()`（不引入 client 端 row model：資料永遠只有當前頁）。理由：三張清單都是伺服器端分頁，改成客戶端等於架空 proto 的 `page`/`page_size`/`Pagination.total`。

**D3 分頁 UI 不重寫**：Ark `ListPagination` 保留為唯一分頁 UI，接 table 實例（`page = pageIndex + 1`、`count = rowCount`、`onPageChange → table.setPageIndex`）。理由：TanStack 與 Ark 是兩套狀態機，讓 table 當唯一真相、Ark 只渲染，可省下重寫 UI ＋ 7 條既有測試。**反向接線（分頁仍走現有 signal）不採**：那 TanStack 只變成 `flexRender` 容器，收益不足。

**D4 `ui/table.tsx` 零修改**：`flexRender(header.column.columnDef.header, header.getContext())` 的輸出直接餵 `TableHead`/`TableCell`。第一版**不做列選取**（`Table` 的 `{...rest}` 落在內層 `<table>`、外層 div 收不到 props 是已知限制）。

**D5 Query 慣例**（沿用 `ability/service.ts` 既有樣板）：
- `queryKey`：`["companies", { page, pageSize, sort, desc, keyword, status }]`（部門同理；角色 `["roles", { page, pageSize, sort, desc }]`）。
- **`retry` 謂詞**：只重試 `Unavailable`/`Unknown`/`DeadlineExceeded`；**不重試**確定性錯誤（`NotFound`/`InvalidArgument`/`PermissionDenied`/`Unauthenticated`/`AlreadyExists`/`FailedPrecondition`）——否則 403 會被退避重打 4 次、錯誤訊息延遲。
- `placeholderData: (prev) => prev`（v5 的 keepPreviousData 等價物）→ 換頁不閃空。
- mutation 成功後 `invalidateQueries({ queryKey: ["companies"] })`（前綴失效；`["ability"]` 的既有慣例不變）。
- 空/錯誤/載入狀態由 query 狀態驅動，**不再各自維護 `loading`/`error` signal**。
- **篩選條件仍是頁面 signal**（使用者輸入的草稿），**只在 submit 時**進入 query key（`setPage(1)` ＋ 更新 key）→ **不得**變成「每按一鍵就查詢」。同理，`page`/`pageSize` 仍由頁面持有（TanStack Table 的 pagination state 與其同步）。
- **`lib/transport.ts` 成為唯一 Connect 來源**：`LoginPage`（Phase 2 的檔案）也在其中 —— 換掉它的 inline client 後，**Phase 2 的登入測試與 modal 測試必須全數保持綠**（不得為此改弱任何斷言）。

**D6 部門頁的「公司下拉選單」**：現況是伺服器端分頁 + 前端累積（`DepartmentsPage:155-166`、`334-342` 的「載入更多」）。改以 `createInfiniteQuery`（若該版本 solid-query 無此 API，退化為單一 query + 累積 `pageSize`），**保留「載入更多」UX**。**該清單同時是 Phase 2 部門 modal 的 `<select>` 選項來源** → 改寫後 Phase 2 的 modal 測試必須保持綠。

**D7 角色清單改表格**：`<ul>` → `ui/Table`（欄位：角色代碼／名稱／系統角色／狀態／ID／操作〔選取〕）。**點選角色的行為與可及性必須逐一對齊現況**（含換頁清空選取 `RolesPage:74-78`）：以列內按鈕承載選取語意（而非整列可點），確保鍵盤可達。

**D8 欄位定義就近放置**：每個 feature 一份 `columns` 陣列（`columnHelper`），與該頁的 action 欄 JSX 同檔；不抽跨頁共用（三張表欄位差異大，強抽會變成設定物件地獄）。

**D9 空列/載入列的 `colspan` 由 `columns.length` 推導**（消滅散落三處的硬寫常數）。

**D10 測試策略**
- 前端（Vitest + jsdom）：①分頁切換的 RPC payload（第 N 頁 → `page: N`、`pageSize: 20`）②篩選 submit 的 payload 且**回到第 1 頁** ③點表頭 → 請求帶 `sort`/`desc` 且回到第 1 頁 ④再次點同欄 → `desc` 反轉 ⑤超頁退回（`maxPage` 情境；**不得無限迴圈**）⑥空狀態/載入列的 `colspan` 隨欄數 ⑦角色清單：選取行為與換頁清空選取。
- 後端（Go）：每個服務的 `sort`/`desc` 白名單與預設排序（enttest sqlite 即可，屬純查詢建構）＋非法值 → `InvalidArgument`；若有 PG 專屬行為才用 testcontainers（`//go:build integration`）。
- **不**以測試釘住 Ark/TanStack 的內部屬性。

**D11 可及性**：可排序表頭需有 `aria-sort`（`ascending`/`descending`/`none`）與可聚焦的按鈕語意；排序變更後焦點不亂跳。

## 3. 三波與檢查點

| 波 | 內容 | 完成檢查點 |
|---|---|---|
| **3A** | solid-query 接手三張清單的資料取得＋`lib/transport.ts` 統一 | 四道 frontend gate 綠；三頁行為不變（payload 逐欄相同）；後端真串接實測（登入→三頁清單載入/換頁/篩選） |
| **3B** | TanStack Table（manual）改寫公司/部門表格；角色清單改表格 | 同上 ＋ 表格頁測試（D10 ①–⑦）綠 |
| **3C** | proto `sort`/`desc` ＋ 後端白名單 ＋ 前端點表頭排序 | 同上 ＋ 後端排序測試 ＋ `task proto:gen`（Go/TS/Dart）＋ 真後端排序實測（含跨頁正確性） |

**每波結束**：四道 gate（frontend：`typecheck && lint && test && build`）＋（3C 另加後端 `go test ./...` 與需要的整合測試）＋ Edge 真後端實測 ＋ scoped 複審；最後一次全分支複審。

## 4. 風險

- **R1（最可能出錯）**：TanStack Table 與 Ark 分頁兩套狀態機的接線 → 只允許 table 為單一真相，Ark 純渲染；換頁/排序後以測試斷言 RPC payload（不是斷言內部狀態）。
- **R2**：`maxPage` 超頁退回在 query 模式下可能無限重試 → 需明確「夾到合法頁碼並只重取一次」，並有測試。
- **R3**：mutation 後清單 stale（建立/編輯/刪除後看到的還是舊資料）→ 以真後端實測（Phase 2 的 modal 流程可直接複用）。
- **R4**：角色清單 `<ul>`→表格會動到既有互動（選取角色→權限矩陣）與可及性 → 逐項對齊並以測試守住。
- **R5**：proto 變更需 `task proto:gen`，其中 Dart 生成**必須**用 fvm pin 的 Dart SDK（`AGENTS.md` §2 有 PATH 指示）；若環境缺該 SDK，**Go/TS 生成仍要完成**，Dart 部分明確回報未生成（不得靜默略過）。
- **R6**：`status` 之類列舉欄位以字典序排序（後端白名單允許、前端不開 UI）——語意已文件化。
- **R7**：資料層改寫是本次最大面積（≈300 行/頁的重複被收斂），但也是行為最容易被改壞之處 → 三頁的 payload 契約測試必須在 3A 先寫（RED → GREEN）。

## 5. 驗收（Phase 3 整體）

1. 四道 frontend gate 綠且 lint 0 warning；後端 `go test ./...`（＋需要的 `task test:integration`）綠。
2. 真後端端到端（Edge、真實滑鼠事件）：三張清單載入、換頁、篩選、排序（升/降、跨頁正確）、建立/編輯/刪除後清單更新、權限矩陣仍正常。
3. 表格頁測試覆蓋 D10 的 ①–⑦；後端排序白名單與非法值測試。
4. 無新依賴（`@tanstack/solid-table` 除外）；`ui/table.tsx`、`ui/pagination.tsx` 對外 API 未變。
5. `impeccable detect` 掃改動面。

## 6. 後續

- Phase 4：Pragmatic D&D（等派車畫面）。
- 另議：`customers` 的 `sort` 是否補 `desc`（前端尚無頁面）；`PermissionMatrix` 是否改成更好操作的 UI；`errorMessage` 跨頁合併（本階段只動與清單相關者）。

---

*建立：2026-09-19*
