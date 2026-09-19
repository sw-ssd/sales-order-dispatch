# 前端表格 Phase 3 實作計畫（三波）

> **For agentic workers:** REQUIRED SUB-SKILL: use superpowers:subagent-driven-development. Steps use `- [ ]`.
> **Goal:** 三張清單改用 solid-query 取資料、TanStack Table（manual）渲染、並支援伺服器端排序（跨 proto/後端/前端）。
> **Spec:** `docs/superpowers/specs/2026-09-19-frontend-tables-phase3-design.md`（D1–D11 為裁定，違反即須回報）
> **前置已定案**：Phase 1（元件庫）、Phase 2（表單）完成；後端 F1/F2 修復且後端可正常起（可登入、OpenFGA 有真權限）

## Global Constraints（每個 task 都適用）

- **新依賴只允許** `@tanstack/solid-table`（`@tanstack/solid-query@^5.102.8` 已在）。不加 devtools。
- **查詢慣例沿用 `lib/ability/service.ts` 既有樣板**（`queryOptions` + Connect client + `queryClient.invalidateQueries`）。
  - `queryKey` 含全部參數；`placeholderData: (prev) => prev`；**`retry` 謂詞只重試 `Unavailable`/`Unknown`/`DeadlineExceeded`**（不得重試確定性錯誤）。
  - **篩選條件是頁面 signal（草稿），只在 submit 時進 query key**；不得每按鍵就查詢。
  - 空/錯誤/載入狀態由 query 狀態驅動，**移除**各自的 `loading`/`error` signal。
- **Table 一律 manual 模式**（`manualPagination` + `rowCount` + `getCoreRowModel`）；**Ark `ListPagination` 保留為唯一分頁 UI**，接 table 實例（D2/D3）。
- **`ui/table.tsx`、`ui/pagination.tsx` 對外 API 不得改動**；第一版不做列選取。
- **排序契約**：`sort`（白名單字串）+ `desc`（bool）；`sort` 空 → 用服務預設排序並**忽略 `desc`**；非法值 → `InvalidArgument` 列出白名單；**任何排序變更都要回第 1 頁**（D1）。
- **不得弱化既有斷言**（Phase 1/2 的測試只能新增；若認為某斷言必須改 → 先回報）。**Phase 2 的 modal 與登入測試必須持續全綠**（含 `lib/transport.ts` 統一與部門公司下拉改寫之後）。
- 前端四道 gate：`cd frontend && pnpm typecheck && pnpm lint && pnpm test && pnpm build`（lint 0 warning）。
- 後端：`cd backend && go test ./...`；動到 SQL/PG 專屬行為時用 `task test:integration`（`//go:build integration` + testcontainers；預設路徑不得起容器）。
- proto 變更後必跑 `task proto:gen`；**Dart 生成需 fvm pin 的 SDK**（`AGENTS.md` §2 的 PATH 指示）。缺 SDK 時 Go/TS 仍須完成，Dart 未生成要**明確回報**。
- 瀏覽器一律 **Microsoft Edge CDP**（`--headless=new` + 真滑鼠事件）；**禁用** Chrome for Testing／`chrome-headless-shell`／Playwright MCP。vite 綁 `::1` → **用 `localhost`**。收尾 `pkill -f "Microsoft Edge.*--headless"` 且 0 殘留。
- 後端/容器收尾：`podman ps -a` 0 殘留（`podman container prune` 清 Created）。
- Harness：shell `grep`/`rg`/`ls`/`find` 被封鎖（用內建工具）；不要 `git add -A`；暫存檔放 `/tmp` 或刪除；註解/commit 一律繁體中文。

---

# 波 3A：資料層（solid-query）

## Task 1: `lib/transport.ts` 統一 ＋ 共用查詢工具

**Files:** `frontend/src/lib/transport.ts`（既有）、`lib/query-client.ts`（既有，可能加共用 `retry` 謂詞）、5 個 inline client 的頁面（`CompaniesPage.tsx:47-49`、`DepartmentsPage.tsx:50-56`、`RolesPage.tsx:15-18`、`LoginPage.tsx:23-25`）、`lib/ability/service.ts`

- [ ] **Step 1** 先讀 `lib/transport.ts`、`lib/ability/service.ts`、`lib/query-client.ts`，確認既有慣例（不要發明第二套）。
- [ ] **Step 2** 把 5 處 inline `createClient(Service, createConnectTransport({baseUrl:"/api/v1"}))` 換成 `createClient(Service, transport)`（**不改任何呼叫端的 payload/行為**）。
- [ ] **Step 3** 在 `lib/query-client.ts`（或新 `lib/query-retry.ts`）加**共用 retry 謂詞**：只重試 `Unavailable`/`Unknown`/`DeadlineExceeded`；附單元測試（餵 `ConnectError` 各 code，斷言 true/false）。
- [ ] **Step 4** 四道 gate ＋ 既有測試全綠（**特別確認 Phase 2 的登入與 modal 測試未被動**：`git diff` 不得出現那些測試檔的刪除行）。
- [ ] **Step 5: commit** `refactor(frontend): Connect client 統一走 lib/transport 並加共用 retry 謂詞`

**Acceptance:** 5 處 inline client 消失；`transport` 是唯一來源；retry 謂詞有單元測試；Phase 1/2 測試全綠。

## Task 2: 公司清單 query 化（**樣板任務**，先做）

**Files:** `frontend/src/features/users/pages/CompaniesPage.tsx`、新增 `features/users/queries.ts`（或 `companies.queries.ts`）、`CompaniesPage.test.tsx`

- [ ] **Step 1（RED）** 先寫 payload 契約測試：①首屏 → `listCompanies({page:1,pageSize:20,status:undefined,keyword:undefined})` ②切到第 2 頁 → `page:2` ③篩選 submit（keyword/status）→ 帶參數**且 `page:1`** ④建立/編輯/刪除成功後 → 清單被重新取得（spy 呼叫數 +1）。
- [ ] **Step 2** 實作 `queryOptions`：`queryKey: ["companies", {page,pageSize,status,keyword}]`、`placeholderData`、共用 retry 謂詞；頁面以 `createQuery` 取值，**移除 `companies`/`total`/`loading`/`error` signals**（保留 `page`/`keyword`/`statusFilter` 草稿）——`total` 改由 query 結果推導（`data.pagination.total`），列資料同理；**不要**另留一份 signal 快取（避免兩份真相）。
- [ ] **Step 3** `maxPage` 超頁退回：**夾到合法頁碼並只重取一次**（不得遞迴無限重取）＋測試（回傳 total 使當前頁超出 → 斷言最終請求的 page 合法且請求次數有界）。
- [ ] **Step 4** mutation（modal 的建立/編輯/刪除）成功後 `invalidateQueries({ queryKey: ["companies"] })`；Phase 2 的 modal 測試保持綠。
- [ ] **Step 5** 四道 gate；**commit** `refactor(frontend): 公司清單改用 solid-query（樣板）`

**Acceptance:** ①②③④ + 超頁退回有測試且請求次數有界；`loading`/`error` signal 已移除；Phase 2 測試未弱化。

## Task 3: 部門清單 query 化 ＋ 公司下拉 `createInfiniteQuery`

**Files:** `features/users/pages/DepartmentsPage.tsx`、`features/users/queries.ts`、`DepartmentsPage.test.tsx`

- [ ] **Step 1（RED）** 測試：部門清單 payload（`page`/`pageSize`/`companyId`）＋超頁退回；公司下拉：首屏載入 50 筆、「載入更多」後**累積**且不重複（`companyPage` 遞增、選項數單調增加）。
- [ ] **Step 2** 依 Task 2 的樣板實作部門清單 query；公司下拉以 `createInfiniteQuery`（若該版本無此 API → 單一 query + 累積 `pageSize`，報告說明）。
- [ ] **Step 3** 公司下拉**同時是 Phase 2 部門 modal 的 `<select>` 選項來源** → modal 測試必須綠；「載入更多」UX 保留。
- [ ] **Step 4** 四道 gate；**commit** `refactor(frontend): 部門清單與公司下拉改用 solid-query`

## Task 4: 角色清單 query 化

**Files:** `features/users/pages/RolesPage.tsx`、`features/users/queries.ts`、新增 `RolesPage.test.tsx`

- [ ] **Step 1（RED）** 測試：`listRoles({page,pageSize})` payload、換頁、超頁退回、**換頁清空已選角色**（現行 `RolesPage:74-78` 的行為）。
- [ ] **Step 2** 依樣板實作 query；`getRolePermissions` 維持現行呼叫方式（可一併 query 化，但**不得改變**矩陣載入時機與 `invalidateQueries(["ability"])` 的既有行為）。
- [ ] **Step 3** 四道 gate；**commit** `refactor(frontend): 角色清單改用 solid-query`

## Task 5: 3A 波驗收

- [ ] **Step 1** 四道 gate（貼輸出）＋ `git diff` 確認 Phase 1/2 測試檔無刪除行。
- [ ] **Step 2** **真後端 Edge 實測**（後端可正常起：`task infra:start` → `migrate:up` → `seed` → `run`；前端 `pnpm dev`，用 `localhost`）：三頁清單載入、換頁、篩選、建立/編輯/刪除後清單即時更新、權限矩陣仍正常。
- [ ] **Step 3** 收尾（Edge/容器 0 殘留、repo 無暫存檔）；**Step 4 回報**（含未驗項）。

---

# 波 3B：TanStack Table（manual）

## Task 6: 公司表格改 TanStack Table（**樣板任務**，先做）

**Files:** `features/users/pages/CompaniesPage.tsx`、`CompaniesPage.test.tsx`

- [ ] **Step 1（RED）** 測試：①表頭欄位數與 `columns` 一致 ②列數等於資料數 ③空狀態/載入列的 `colspan === columns.length` ④分頁 UI 顯示的頁碼/總數與查詢結果一致 ⑤點下一頁 → `page:2`（沿用 Task 2 的斷言）。
- [ ] **Step 2** `createSolidTable({ data, columns, state:{pagination}, manualPagination:true, rowCount: total(), getCoreRowModel: getCoreRowModel(), onPaginationChange })`；`flexRender` 餵 `TableHead`/`TableCell`；`columnHelper` 定義 6 欄（含操作欄 JSX）。
- [ ] **Step 3** Ark `ListPagination` 接 table 實例（`page={pageIndex+1}`、`count={rowCount}`、`onPageChange` → `setPageIndex`）——**不重寫分頁元件**；確認既有 `pagination.test.tsx` 未被動。**`page` signal 由 table 的 pagination state 取代**（query key 的 `page` 一律取自 table 狀態，不得同時保留兩份頁碼來源）。
- [ ] **Step 4** `colspan` 常數由 `columns.length` 推導；確認 `ui/table.tsx` 零修改（`git diff` 應為空）。
- [ ] **Step 5** 四道 gate ＋ Edge 實測（載入/換頁/篩選/空狀態）；**commit** `refactor(frontend): 公司表格改 TanStack Table（manual；樣板）`

## Task 7: 部門表格改 TanStack Table

- [ ] **Step 1（RED）** 同 Task 6 的 ①–⑤（部門 4 欄）。
- [ ] **Step 2** 依樣板實作；`ui/table.tsx` 與分頁元件不得動。
- [ ] **Step 3** 四道 gate ＋ Edge 實測；**commit** `refactor(frontend): 部門表格改 TanStack Table（manual）`

## Task 8: 角色清單改為表格（保留選取互動）

- [ ] **Step 1（RED）** 測試：①列渲染（代碼/名稱/系統/狀態/ID）②**點「選取」按鈕 → 載入該角色權限矩陣**（現行行為）③**換頁行為＝清空舊選取 → 自動選取新頁第一筆並載入其矩陣**（T4 已以測試釘住此現行行為，**不是**單純清空；不得在此 task 改變它）④鍵盤可達（按鈕為可聚焦元素、有名稱）。
- [ ] **Step 2** `<ul>` → `ui/Table` + TanStack Table（manual）；以**列內按鈕**承載選取語意（不是整列可點）。
- [ ] **Step 3** 四道 gate ＋ Edge 實測（含用鍵盤操作）；**commit** `refactor(frontend): 角色清單改為表格並保留選取互動`

## Task 9: 3B 波驗收

- [ ] **Step 1** 四道 gate（貼輸出）；`ui/**` 與 `pagination.test.tsx` 零變更的 `git diff` 證明。
- [ ] **Step 2** 真後端 Edge 實測：三頁表格、換頁、篩選、角色選取→矩陣、建立/編輯/刪除後列更新。
- [ ] **Step 3** 收尾；**Step 4 回報**。

---

# 波 3C：伺服器端排序（跨 proto/後端/前端）

## Task 10: proto 新增 `sort` / `desc` ＋ 生成

**Files:** `backend/proto/salesorder/v1/company.proto`（`ListCompaniesRequest`、`ListDepartmentsRequest`）、`role.proto`（`ListRolesRequest`）

- [ ] **Step 1** 依既有 `customer.proto:63-64` 的風格加註解與欄位：`string sort = N; // 白名單:<列出>（空=預設排序）`、`bool desc = N+1; // 是否降冪（sort 空時忽略）`。
- [ ] **Step 2** `cd backend && task proto:gen`（Go/TS/Dart；Dart 需 fvm pin 的 SDK，缺則明確回報未生成）。
- [ ] **Step 3** 確認生成的 Go/TS 檔含新欄位（貼證據）；四道 frontend gate ＋ 後端 `go test ./...`（此時未改服務，行為不變）。
- [ ] **Step 4: commit** `feat(proto): ListCompanies/ListDepartments/ListRoles 新增 sort 與 desc`

## Task 11: 公司服務排序白名單（**樣板任務**，先做）

**Files:** `backend/internal/services/company_service.go`（＋既有測試檔新增案例）

- [ ] **Step 1（RED）** 測試：`sort=""` → `id DESC`（現行行為）；`sort="name"` → 依 name 升冪；`sort="name", desc=true` → 降冪；`sort="status"` 可用；`sort="bogus"` → `InvalidArgument` 且訊息列出白名單；**`sort=""`, `desc=true` → 仍為預設排序**。
- [ ] **Step 2** 照 `customerSortField`（`customer_service.go:243-247`）的樣板實作映射與 `ent.Asc/Desc`；**不得改動**未帶 `sort` 時的既有排序與分頁語意。
- [ ] **Step 3** 後端 `go test ./...`；**commit** `feat(backend): 公司清單支援 sort/desc 白名單（樣板）`

## Task 12: 部門服務排序

- [ ] **Step 1（RED）** 同 Task 11 的案例（白名單 `name`/`id`；預設 `id DESC`）。
- [ ] **Step 2** 依樣板實作；**Step 3** `go test ./...`；**commit** `feat(backend): 部門清單支援 sort/desc 白名單`

## Task 13: 角色服務排序

- [ ] **Step 1（RED）** 白名單 `code`/`name`/`id`；**預設 `id ASC`（現行行為，與上兩者不同！）**。
- [ ] **Step 2** 依樣板實作；**Step 3** `go test ./...`；**commit** `feat(backend): 角色清單支援 sort/desc 白名單`

## Task 14: 前端點表頭排序（三頁）

**Files:** 三個頁面的 `columns`（`enableSorting` + header 按鈕）、query key 加上 `sort`/`desc`、`CompaniesPage.test.tsx`/`DepartmentsPage.test.tsx`/`RolesPage.test.tsx`

- [ ] **Step 1（RED）** 測試：①點 `name` 表頭 → 請求帶 `sort:"name"`、`desc:false` 且 **`page:1`** ②再點同欄 → `desc:true` ③點另一欄 → 新欄 `desc:false` ④`aria-sort` 為 `ascending`/`descending`/`none` 正確 ⑤可排欄位**僅限**白名單（公司 `name`/`identifier`/`tax_id`/`id`；部門 `name`/`id`；角色 `code`/`name`/`id`）。
- [ ] **Step 2** TanStack 的 `sorting` state ↔ query key 映射（`sorting[0]?.id ?? ""`、`sorting[0]?.desc ?? false`）；`manualSorting: true`；表頭以按鈕承載（可聚焦、有名稱、`aria-sort` 下在 `<th>`）。
- [ ] **Step 3** 四道 gate；**commit** `feat(frontend): 三張表支援點表頭排序（伺服器端）`

## Task 15: 3C 波驗收（含真後端跨頁排序）

- [ ] **Step 1** 四道 gate ＋ 後端 `go test ./...`（貼輸出）。
- [ ] **Step 2** 真後端 Edge 實測（**關鍵：跨頁正確性**）：以足夠資料量（seed 或先建立 >20 筆）驗：升冪第 1 頁 vs 第 2 頁的接續性（不得出現只排當前頁的假象）、降冪、換頁後排序保持、改排序回第 1 頁、非法 `sort` 不會從 UI 產生（且以 `curl`/probe 驗後端回 `InvalidArgument`）。
- [ ] **Step 3** 收尾（Edge/容器/進程 0 殘留）；**Step 4 回報**（含未驗項）。
- [ ] **Step 5** 最終全分支複審（另行指派）涵蓋 `1980b86..HEAD` 的整個 Phase 3。

---

## 驗收對照（spec §5）

| spec 驗收 | 步驟 |
|---|---|
| 1. 四道 gate ＋ 後端測試 | 每波驗收（T5/T9/T15 Step 1） |
| 2. 真後端端到端（清單/換頁/篩選/排序/CRUD/矩陣） | T5/T9/T15 Step 2 |
| 3. 表格測試 ①–⑦ ＋ 後端白名單測試 | T2–T4、T6–T8、T11–T14 |
| 4. 無新依賴（solid-table 除外）、`ui/**` API 未變 | T6 Step 4、T9 Step 1 |
| 5. `impeccable detect` | T15 Step 2 附帶 |

*建立：2026-09-19*
