---
title: 後端細部計畫 09 — 單據列印(四種模板 / Gotenberg PDF / 列印與預覽 API)
source_tasks: Task 5.3 四種單據模板、Task 5.4 Gotenberg PDF 產生服務、Task 5.5 列印與預覽 API(docs/superpowers/plans/2026-07-17-sales-order-1-0-tasks.md v2.9.0)
spec_sections: 規格書 v1.0.34 第 7 章「派車與單據列印」(§7.2 單據列印與列印流程);§5.1 / §5.2(print_logs、print_previews、file_assets 欄位);§6.4(base_qty 基本單位彙總);§14.1(print_logs 保留 2 年、print_previews 保留 90 天)
depends_on:
  - 04-master-data.md(Task 3.6 檔案資產 API:file_assets 落檔與下載 URL)
  - 08-dispatch.md(Task 5.1 / 5.2:派車確認後訂單 status = processing、route_id / delivery_sequence 資料來源)
decisions: D3(RLS 最後防線)、D4(Connect-RPC)、D12(全系統無金額)、D15(Gotenberg、四單據內容限制、重印需原因、PDF 全留存)、D17(本地儲存 + 白名單)、D18(稽核同事務)
---

# 09 — 單據列印細部計畫

## 共通規則(各子功能不重複贅述)

- **無金額**:全系統不儲存、不顯示任何金額(D12);四種單據的 view model 從源頭就不含金額欄位,模板層無從渲染價格。
- **多租戶**:print_logs / print_previews / file_assets 皆帶 `company_id` / `department_id`;查詢由 repository 注入部門條件,PostgreSQL RLS(`app.current_company_id` / `app.current_department_id` / `app.current_data_scope`)為最後防線(D3)。
- **軟刪除**:file_assets 適用 `deleted_at` 軟刪除(D10);print_logs / print_previews 為記錄類資料,不軟刪除,改由保留排程管理(print_logs 2 年、print_previews 90 天,§14.1)。
- **同事務**:print_logs / print_previews 與其 file_assets 元資料記錄、以及正式列印的 audit_logs(action = print)同一資料庫交易寫入,同成功同失敗(D18);PDF 實體檔案因涉及外部 Gotenberg 呼叫,先產生落檔再進交易,交易失敗時補償刪除檔案。
- **樂觀鎖衝突**:凡比對 version 失敗一律回 Connect `failed_precondition`,由前端重新整理。
- **錯誤碼約定**:未登入 `unauthenticated`、權限不足 `permission_denied`、資源不存在 `not_found`、狀態/前置條件不符 `failed_precondition`、參數缺失或格式錯誤 `invalid_argument`、唯一性衝突 `already_exists`;Gotenberg 等上游基礎設施故障回 `internal` 並記錄 trace。
- **紙張與字型**:四種單據皆 A4;字型統一 Noto Sans CJK TC(D15)。

---

## Task 5.3 — 四種單據模板

### 子功能 5.3.1: 單車總表模板(dispatch_summary)

- **目標**: 建立 A4 HTML/CSS 模板,依車次分組列出該車當日所有店家與明細摘要,不含任何金額欄位(D12 / D15)。
- **檔案**: Create `backend/internal/print/templates/dispatch_summary.html`
- **介面**: Go html/template 模板;輸入 view model 含:車次(代碼、名稱)、出貨日期、列印時間、公司名稱,以及店家列表——每店含 `customer_code`、店名、配送地址、配送順序(`delivery_sequence`)、品項摘要(品名、數量、單位、特殊分切備註)。view model 不含金額欄位。
- **實作邏輯**:
  1. 一份文件對應單一車次 + 單一出貨日期;頁首固定呈現公司名稱、單據名稱「單車總表」、車次、出貨日期、列印時間。
  2. 店家區塊依 `delivery_sequence` 升冪排列,與派車看板順序一致,司機按單配送。
  3. 每店區塊內列出品項摘要:品名、數量、單位;有 `special_cut_note` 的品項附註提醒,供司機核對加工品。
  4. 店家數過多跨頁時,表頭於每頁重複(page-break 與 thead 重複規則以 CSS 控制)。
  5. 模板不提供任何金額/單價/小計欄位;view model 建構處(5.4.2)即不帶金額,雙重防呆。
- **錯誤處理**: 模板渲染層不直接回 Connect code;車次不存在等資料問題由 5.4.2 組合層回 `not_found`;模板缺欄位屬開發期錯誤,以模板渲染測試涵蓋,不上線後才暴露。
- **驗收**:
  - [ ] 文件依車次分組,一車一份。
  - [ ] 店家排序與 `delivery_sequence` 一致。
  - [ ] 全文件無任何金額、單價、小計字樣或欄位。
  - [ ] 特殊分切備註正確顯示於對應品項。

### 子功能 5.3.2: 對點單模板(delivery_note)

- **目標**: 建立每個店家一張 A4 的對點單模板,依車次 → 店家分組,不顯示價格(D15)。`相依: 5.3.1`(共用模板目錄與 view model 慣例)
- **檔案**: Create `backend/internal/print/templates/delivery_note.html`
- **介面**: Go html/template 模板;輸入 view model 含:車次、出貨日期、店家資訊(`customer_code`、店名、地址、聯絡人、電話)、訂單明細列表(品名、數量、單位、分切規格、品項備註),以及頁尾簽收欄(客戶簽名、收貨日期留白)。view model 不含價格欄位。
- **實作邏輯**:
  1. 一次列印可涵蓋同車次多店家,每店強制換頁(page-break-after),確保每店獨立一張 A4 可撕交客戶核對。
  2. 店家順序依 `delivery_sequence` 升冪;頁首呈現車次、出貨日期、店名與 `customer_code`。
  3. 明細僅列品名、數量、單位;有分切規格者併列規格名稱,方便現場對點加工品。
  4. 頁尾保留簽收區:客戶簽名欄與收貨日期欄固定為空白,供現場手寫。
  5. 模板無價格/金額欄位;view model 建構處不帶價格。
- **錯誤處理**: 同 5.3.1,渲染層不回 Connect code;店家不存在由組合層回 `not_found`。
- **驗收**:
  - [ ] 每個店家恰一張 A4,跨店強制換頁。
  - [ ] 單張僅含該店家品項,不混入其他店。
  - [ ] 全文件無價格/金額欄位。
  - [ ] 簽收欄列印為空白可手寫。

### 子功能 5.3.3: 揀貨單模板(picking_list)

- **目標**: 建立依「車次 → 倉別 → 商品分類 → 品名」排序的揀貨單模板,跨店家彙總數量供倉庫揀貨。`相依: 5.3.1`
- **檔案**: Create `backend/internal/print/templates/picking_list.html`
- **介面**: Go html/template 模板;輸入 view model 含:車次、出貨日期、倉別名稱,以及彙總品項列表——每列含商品分類、品名、彙總數量(基本單位,來自 `sales_order_items.base_qty`)、需分切提示、揀貨核對欄(空白)。
- **實作邏輯**:
  1. 一份文件對應單一車次 + 單一倉別 + 單一出貨日期;倉別由 `sales_order_items.warehouse_id` 決定。
  2. 品項排序:先商品分類(依分類 `sort_order` 或名稱),再品名筆畫/字序;同商品跨店家合併為一列。
  3. 彙總數量以基本單位呈現,加總各明細 `base_qty`(§6.4),不混用下單單位,避免倉庫誤讀。
  4. 有分切規格或 `special_cut_note` 的品項加註「需加工」提示,引導揀貨人員轉交加工室(與 5.3.4 銜接)。
  5. 每列保留空白揀貨核對欄,供實揀打勾或手寫實揀量。
- **錯誤處理**: 倉別不存在由組合層回 `not_found`;渲染層不回 Connect code。
- **驗收**:
  - [ ] 排序符合車次 → 倉別 → 商品分類 → 品名。
  - [ ] 同商品跨店家彙總為一列,數量以基本單位呈現。
  - [ ] 需加工品項有明確提示。
  - [ ] 揀貨核對欄為空白可手寫。

### 子功能 5.3.4: 加工單模板(processing_list)

- **目標**: 建立分「加工室揀 / 配送揀」兩區塊的加工單模板,顯示原始數量,「加工後數量」欄列印空白供手寫回填,1.0 不回寫系統(D15)。`相依: 5.3.1`
- **檔案**: Create `backend/internal/print/templates/processing_list.html`
- **介面**: Go html/template 模板;輸入 view model 含:車次、出貨日期,以及兩個區塊列表——「加工室揀」區塊(倉別、品名、原始數量、基本單位、分切規格)、「配送揀」區塊(店家、品名、原始數量、單位、特殊分切備註);兩區塊每列皆含「加工後數量」欄,模板固定渲染為空白儲存格。
- **實作邏輯**:
  1. 資料範圍:僅納入有 `cutting_spec_id` 或 `special_cut_note` 的訂單明細;無加工需求的明細不出現。
  2. 區塊順序固定:先「加工室揀」後「配送揀」(§7.2),跨頁時區塊標題重複。
  3. 「加工室揀」區塊:依倉別 → 品名排序,顯示需從倉庫領至加工室的原料原始數量(基本單位)與分切規格。
  4. 「配送揀」區塊:依店家 `delivery_sequence` → 品名排序,顯示加工完成後隨車配送至各店的原始數量與下單單位。
  5. 「加工後數量」欄位兩區塊皆固定空白——模板層不輸出任何數值,由加工人員現場手寫回填;系統不提供回填 API(1.0 不回寫)。
- **錯誤處理**: 渲染層不回 Connect code;無加工品項時由組合層標記空表(5.4.3 不產生 PDF)。
- **驗收**:
  - [ ] 僅含需加工的品項。
  - [ ] 「加工室揀」區塊在前、「配送揀」區塊在後。
  - [ ] 原始數量與分切規格正確顯示。
  - [ ] 「加工後數量」欄於所有列皆列印為空白。

---

## Task 5.4 — Gotenberg PDF 產生服務

### 子功能 5.4.1: Gotenberg client(HTML → PDF)

- **目標**: 實作後端對 Gotenberg 的 HTTP client,將 HTML/CSS 轉為 A4 PDF。
- **檔案**: Create `backend/internal/print/client.go`
- **介面**: 內部 Go client,方法接受 HTML 內容(byte 切片)與轉換選項(紙張 A4、邊界、等待頁面載入完成),回傳 PDF byte 切片;對外不暴露為 RPC。
- **實作邏輯**:
  1. 以 multipart/form-data 組請求:主文件 `index.html` 加上模板引用的資源(字型宣告、CSS),送 Gotenberg 的 Chromium HTML 轉換路由。
  2. 轉換參數固定:紙張 A4、列印背景圖形開啟、合理頁邊距、等待網路靜止再輸出,確保字型與樣式載入完成。
  3. 呼叫設逾時(整體 deadline),避免 Gotenberg 卡死拖住 RPC。
  4. 回應非成功時讀取 Gotenberg 錯誤內容,轉為內部錯誤並帶 trace 資訊。
  5. 重試規則:連線錯誤與上游 5xx 以有限次數(至多 2 次)指數退避重試;4xx(如 HTML 無法解析)屬請求錯誤,不重試直接失敗。
  6. PDF 全程在記憶體傳遞,不落暫存檔;交由 5.4.3 寫入正式儲存。
- **錯誤處理**: Gotenberg 不可達或轉換失敗(重試用盡)→ `internal`;HTML 內容被上游拒絕(4xx)→ `invalid_argument`(屬模板/資料問題,開發期應被測試攔下);逾時 → `internal` 並記錄。
- **驗收**:
  - [ ] 合法 HTML 可轉出 A4 PDF。
  - [ ] 上游 5xx 有有限重試,4xx 不重試。
  - [ ] 呼叫有逾時上限,不無限等待。

### 子功能 5.4.2: 資料組合(車次 / 日期 / 單據類型)

- **目標**: 依單據類型從訂單資料組出各模板的 view model,落實 5.3 的分組、排序與彙總規則。`相依: 5.3.1–5.3.4`
- **檔案**: Create `backend/internal/print/service.go`
- **介面**: 內部 service 方法,輸入:單據類型(`dispatch_summary` / `delivery_note` / `picking_list` / `processing_list`)、`route_id`、`target_date`,以及依類型而定的選擇器(對點單可指定 `customer_id` 單印一店;揀貨單可指定 `warehouse_id` 單印一倉);輸出:對應 view model 與「是否空表」標記。供 5.5 的 Preview / Print RPC 呼叫,非獨立端點。
- **實作邏輯**:
  1. 依 `department_id`(RLS 注入)+ `route_id` + `target_date`(對應 `sales_orders.expected_delivery_date`)查詢訂單與明細,聯結 customers、routes、warehouses、product_categories、cutting_specs 補齊顯示名稱。
  2. 單車總表:依 `delivery_sequence` 排店家,品項以原始下單數量與單位列出。
  3. 對點單:依車次 → 店家分組;指定 `customer_id` 時僅取該店;明細含分切規格。
  4. 揀貨單:依車次 → 倉別 → 商品分類 → 品名排序;同商品跨店合併,數量加總 `base_qty` 以基本單位呈現。
  5. 加工單:僅取有 `cutting_spec_id` 或 `special_cut_note` 的明細,分「加工室揀」(倉別 → 品名)與「配送揀」(配送順序 → 品名)兩集合。
  6. 邊界規則:查無任何符合明細時回傳空集合標記,由 5.4.3 決定不產生 PDF;軟刪除的主檔(如已刪商品)仍顯示歷史名稱編號(D10)。
  7. 狀態過濾不在此層:預覽不限狀態、正式列印限 `processing` 的判斷由 5.5.3 執行,本層只負責資料成形。
- **錯誤處理**: `route_id` / `customer_id` / `warehouse_id` 不存在或不屬當前部門 → `not_found`(RLS 下查不到視同不存在);缺必要參數或類型與選擇器組合不合法 → `invalid_argument`。
- **驗收**:
  - [ ] 四種類型各產出符合 5.3 分組/排序/彙總規則的 view model。
  - [ ] 查無資料時回傳空表標記而非錯誤。
  - [ ] 跨部門資料在 RLS 下不可見。

### 子功能 5.4.3: PDF 產生並關聯 file_assets(空表不產生、Noto Sans CJK TC)

- **目標**: 串接模板渲染 → Gotenberg 轉換 → file_assets 落檔的完整產線;空表不產生 PDF;字型使用 Noto Sans CJK TC(D15)。`相依: 3.6(Task 3.6 檔案資產 API)、5.4.1、5.4.2`
- **檔案**: Update `backend/internal/print/service.go`
- **介面**: 內部 service 方法,輸入 view model 與單據類型,輸出 `file_asset_id` 與下載 URL;file_assets 記錄的 `owner_type` 對應列印用途(print_log / print_preview)、`owner_id` 於 5.5 寫入記錄後回填或同事務建立、`mime_type` 為 application/pdf。
- **實作邏輯**:
  1. 空表守門:view model 標記為空表時直接回 `failed_precondition`(無可列印資料),不呼叫 Gotenberg、不寫任何記錄——空表不印(D15)。
  2. 以對應模板渲染 HTML;字型於模板 CSS 以 `Noto Sans CJK TC` 字族宣告,Gotenberg 環境需具備該字型,確保中文正確嵌入。
  3. 呼叫 5.4.1 client 取得 PDF;PDF 大小須在 file_assets 白名單上限(pdf ≤ 10 MB,D17)內,超出屬異常回 `internal` 並告警。
  4. 經 Task 3.6 的檔案資產服務寫入本地 volume 並建立 file_assets 元資料(`company_id` / `department_id` 帶當前租戶,`created_by` 帶操作人)。
  5. 交易邊界:Gotenberg 呼叫與檔案落檔在 DB 交易外完成;file_assets 元資料與 print_logs / print_previews 記錄、audit_logs(正式列印)於同一交易寫入(D18);交易失敗時補償刪除已落檔的 PDF,避免孤兒檔。
  6. 每次產生皆為新 PDF、新 file_asset,即使內容相同也不覆用——PDF 全留存,保留「當初印了什麼」(D15)。
- **錯誤處理**: 空表 → `failed_precondition`;Gotenberg 故障 → `internal`(依 5.4.1 規則);儲存失敗 → `internal` 並補償清理;租戶上下文缺失 → `unauthenticated`。
- **驗收**:
  - [ ] 空表不產生 PDF,且不留下任何 file_assets / 列印記錄。
  - [ ] PDF 中文以 Noto Sans CJK TC 正確顯示,無缺字方塊。
  - [ ] 每次產生皆關聯新的 file_asset,可經下載 URL 取回。
  - [ ] 交易失敗時不殘留孤兒 PDF 檔。

---

## Task 5.5 — 列印與預覽 API

### 子功能 5.5.1: print_logs / print_previews schema

- **目標**: 建立正式列印與預覽兩張記錄表,欄位對齊規格 §5.2,供 5.5.2–5.5.4 使用。
- **檔案**: Create `backend/ent/schema/printlog.go`、`backend/ent/schema/printpreview.go`,以及對應 DB migration(RLS policy 與索引隨 migration 建立)
- **介面**: Ent 實體:
  - `PrintLog`:`company_id`、`department_id`、`document_type`(`dispatch_summary` / `delivery_note` / `picking_list` / `processing_list`)、`route_id`、`customer_id`(可空,對點單單店時使用)、`warehouse_id`(可空,揀貨單單倉時使用)、`target_date`、`is_reprint`、`reprint_reason`(可空,重印時必填)、`printed_by`、`printed_at`、`file_asset_id`。
  - `PrintPreview`:`company_id`、`department_id`、`document_type`、`route_id`、`customer_id`(可空)、`warehouse_id`(可空)、`target_date`、`previewed_by`、`previewed_at`、`file_asset_id`。
- **實作邏輯**:
  1. 兩表為記錄類資料,不適用軟刪除(同 audit_logs / sales_order_events 的處理,D10)。
  2. RLS policy:以 `app.current_company_id` / `app.current_department_id` / `app.current_data_scope` 限制可見範圍,developer 角色繞過(D3 / D8)。
  3. 索引:以(`department_id`、`target_date`、`document_type`)支援列印記錄查詢;`file_asset_id` 建外鍵參照 file_assets。
  4. 判斷「是否已有正式列印記錄」的查詢以(部門、`document_type`、`route_id`、`target_date`、選用的 `customer_id` / `warehouse_id`)比對,供 5.5.3 / 5.5.4 推導 `is_reprint`。
  5. 保留排程另由維運工作處理(print_logs 2 年、print_previews 90 天,§14.1),schema 層僅備註,不在本子功能實作排程。
- **錯誤處理**: schema 層無 Connect code;migration 失敗屬部署期問題,須可重複升降。
- **驗收**:
  - [ ] migration 可升可降,兩表欄位與 §5.2 一致。
  - [ ] RLS 生效:跨部門帳號查不到他部門列印記錄。
  - [ ] 重印判斷查詢有對應索引,不全表掃描。

### 子功能 5.5.2: Preview API(不影響正式記錄)

- **目標**: 提供任何訂單狀態皆可使用的預覽,產生 PDF 寫入 print_previews,完全不觸碰 print_logs(D15)。`相依: 5.4.3、5.5.1`
- **檔案**: Create `backend/internal/domain/prints/preview.go`(service handler 與 repository)
- **介面**: Connect-RPC `PrintService.Preview`;請求訊息含 `document_type`、`route_id`、`target_date`、選用 `customer_id` / `warehouse_id`;回應訊息含 `preview_id`、`file_asset_id`、`download_url`。
- **實作邏輯**:
  1. 認證與授權:走既有 auth interceptor;需具備 resource `print` 的 read/print 權限(dept_admin / staff 於自己部門,RLS 限制範圍)。
  2. 參數驗證:單據類型合法、選擇器與類型組合合法(如揀貨單才接受 `warehouse_id`)。
  3. **不檢查訂單狀態**——pending、processing、completed 皆可預覽(§7.2 列印流程第 1 步)。
  4. 呼叫 5.4.2 組資料、5.4.3 產 PDF;空表同樣不產生,回 `failed_precondition`。
  5. 同事務寫入 print_previews 與 file_assets 元資料;預覽不寫 audit_logs(print_previews 本身即為記錄,D18 稽核保留給正式列印)。
  6. 回傳下載 URL(經 Task 3.6 檔案下載端點),前端直接開啟或下載。
- **錯誤處理**: 未登入 → `unauthenticated`;無 print 權限或跨部門 → `permission_denied`(RLS 查不到資源時以 `not_found` 呈現);車次/店家/倉別不存在 → `not_found`;參數組合不合法 → `invalid_argument`;空表 → `failed_precondition`。
- **驗收**:
  - [ ] pending 狀態訂單可預覽並取得 PDF 下載 URL。
  - [ ] 預覽後 print_logs 無任何新增,print_previews 新增一筆且關聯 PDF。
  - [ ] 空表預覽回 `failed_precondition` 且不寫記錄。

### 子功能 5.5.3: Print API(檢查 status = processing,PDF 經 file_assets 下載 URL)

- **目標**: 提供正式列印:範圍內訂單須全為 `processing`,產生 PDF 寫入 print_logs(`is_reprint = false`),並提供列印記錄查詢(D15)。`相依: 5.4.3、5.5.1、Task 5.1(派車確認)`
- **檔案**: Create `backend/internal/domain/prints/print.go`、`backend/internal/domain/prints/list_logs.go`(service handler 與 repository)
- **介面**:
  - Connect-RPC `PrintService.Print`;請求訊息含 `document_type`、`route_id`、`target_date`、選用 `customer_id` / `warehouse_id`、選用 `reprint_reason`(重印時必填,規則見 5.5.4);回應訊息含 `print_log_id`、`file_asset_id`、`download_url`、`is_reprint`。
  - Connect-RPC `PrintService.ListLogs`;請求含日期範圍、單據類型、車次篩選與分頁;回應為 print_logs 列表,每筆含操作人、時間、`is_reprint`、`reprint_reason` 與 `download_url`。
- **實作邏輯**:
  1. 認證與授權:resource `print` × action `print`;dept_admin / staff 於自己部門。
  2. 狀態檢查:鎖定範圍(部門 + 車次 + 出貨日期 + 選用選擇器)內的目標訂單,逐筆確認 `status = processing`;任一筆不符(尚未派車確認、已取消派車退回 pending、已作廢)→ 整批拒絕回 `failed_precondition`,訊息指出不符合的訂單單號。
  3. 首印/重印推導:依 5.5.1 的比對鍵查 print_logs;無既有記錄 → 首印,`is_reprint = false`,`reprint_reason` 必須為空;有既有記錄 → 轉重印規則(5.5.4)。
  4. 呼叫 5.4.2 / 5.4.3 產 PDF(空表守門照常生效)。
  5. 同事務寫入 print_logs、file_assets 元資料與 audit_logs(action = print,含單據類型、車次、日期、是否重印與原因)(D18);狀態檢查與寫入在同一交易完成,避免檢查後狀態被並發改變的窗口——交易內以當前讀取再次確認狀態,不一致則整體回滾。
  6. PDF 一律以 file_assets 下載 URL 交付,不在 RPC 回應夾帶二進位。
  7. ListLogs 僅讀,依 RLS 範圍過濾,分頁上限沿用全系統 per_page ≤ 100 慣例。
- **錯誤處理**: 未登入 → `unauthenticated`;無權限 → `permission_denied`;車次/店家/倉別不存在 → `not_found`;訂單非 processing → `failed_precondition`;首印帶入 `reprint_reason` 或參數組合不合法 → `invalid_argument`;空表 → `failed_precondition`;交易內狀態再確認失敗 → `failed_precondition` 並回滾。
- **驗收**:
  - [ ] 含 pending 訂單的車次正式列印被拒,回 `failed_precondition`。
  - [ ] 首印成功:print_logs 一筆 `is_reprint = false`、audit_logs 一筆、PDF 可經下載 URL 取得。
  - [ ] ListLogs 依部門範圍回傳,含重印標記與原因。
  - [ ] 狀態檢查與記錄寫入同成功同失敗,無半套記錄。

### 子功能 5.5.4: 重印(必填原因、is_reprint = true、開放 dept_admin/staff)

- **目標**: 已有正式列印記錄後的重印:`dept_admin` / `staff` 皆可執行,必填 `reprint_reason`,新記錄標 `is_reprint = true`,並產生新 PDF 留存(D15)。`相依: 5.5.3`
- **檔案**: Update `backend/internal/domain/prints/print.go`(重用 `PrintService.Print` 端點的重印分支)
- **介面**: 沿用 `PrintService.Print`;當比對鍵已存在 print_logs 時進入重印分支,請求的 `reprint_reason` 轉為必填;回應 `is_reprint = true`。
- **實作邏輯**:
  1. 重印判定:以(部門、`document_type`、`route_id`、`target_date`、選用 `customer_id` / `warehouse_id`)查到任一既有 print_log 即屬重印。
  2. 原因守門:`reprint_reason` 空白(空字串或純空白)→ `invalid_argument`;原因原文留存於新記錄。
  3. 權限:開放 `dept_admin` 與 `staff`(D15),更高角色(super / company_admin 於範圍內)當然可行;customer / guest 等外部身分一律 `permission_denied`。
  4. 狀態仍須全部 `processing`:若範圍內訂單已被取消派車退回 pending(§7.1 情境:提示需重新列印),重印同樣拒絕,待重新派車確認後再印——避免印出與看板不一致的單據。
  5. 每次重印產生**新 PDF、新 file_asset、新 print_log**(`is_reprint = true`、`printed_by` / `printed_at` 為本次操作),舊記錄與舊 PDF 保留不動,完整留存列印軌跡。
  6. 同事務寫入 print_logs + file_assets 元資料 + audit_logs(action = print,after_snapshot 含 `is_reprint` 與原因)(D18)。
- **錯誤處理**: 未填原因 → `invalid_argument`;外部身分 → `permission_denied`;訂單非 processing → `failed_precondition`;比對鍵資源不存在 → `not_found`;未登入 → `unauthenticated`。
- **驗收**:
  - [ ] 未填原因的重印回 `invalid_argument`,不產生任何記錄。
  - [ ] `staff` 身分可完成重印;customer / guest 被拒。
  - [ ] 重印後 print_logs 新增 `is_reprint = true` 且含原因,舊記錄與舊 PDF 仍在。
  - [ ] 訂單已退回 pending 時重印回 `failed_precondition`。

---

## 整合測試重點

- **端到端列印流程**:派車確認(Task 5.1,訂單轉 processing)→ 四種單據各執行正式列印 → print_logs / file_assets / audit_logs 三處記錄同時存在,PDF 可經下載 URL 取回且中文(Noto Sans CJK TC)正確顯示。
- **狀態守門**:車次內含 pending 訂單時正式列印整批拒絕(`failed_precondition`)且無半套記錄;取消派車後重印同樣被拒。
- **預覽隔離**:任何狀態預覽成功,print_logs 計數不變;print_previews 與其 PDF 獨立留存。
- **重印軌跡**:首印 → 重印(填原因)→ 再重印,三筆 print_logs 依序存在,僅首筆 `is_reprint = false`,三份 PDF 皆可取回;缺原因的重印被拒且不留記錄。
- **空表不印**:無明細的車次/日期/類型組合,預覽與正式列印皆回 `failed_precondition`,file_assets 無新增、本地無孤兒檔。
- **排序與分組**:揀貨單 PDF 內容符合車次 → 倉別 → 商品分類 → 品名且同商品以 `base_qty` 彙總;加工單兩區塊順序正確且「加工後數量」全空白;對點單每店一頁;單車總表店家順序與 `delivery_sequence` 一致;四單據全文無金額字樣。
- **多租戶**:A 部門 staff 對 B 部門車次預覽/列印,呈現 `not_found` 或 `permission_denied`;ListLogs 跨部門不可見(RLS 生效)。
- **Gotenberg 故障**:模擬上游 5xx 觀察有限重試與 `internal` 回應;交易失敗路徑確認 PDF 補償刪除、不留孤兒檔。
