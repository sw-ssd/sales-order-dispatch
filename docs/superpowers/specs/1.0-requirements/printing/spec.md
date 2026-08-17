# printing 需求規格

> 來源：原 OpenSpec delta spec（OpenSpec 工作流已停用，2026-08-03 遷移至 docs/）。


## Purpose

定義四種出貨單據（單車總表、對點單、揀貨單、加工單）的內容、排序規則，以及預覽、正式列印、重印的狀態門檻、記錄（`print_logs` / `print_previews`）與 PDF 產生留存規則。

## Requirements

### Requirement: 單車總表內容與排序

系統 SHALL 提供單車總表（`document_type = dispatch_summary`），內容為指定車次當日的所有店家及各店家訂單明細摘要，依車次分組呈現，且 MUST NOT 顯示任何金額資訊。

#### Scenario: 依車次分組列出店家明細摘要

- **WHEN** 使用者產生某車次某日期的單車總表
- **THEN** 單據列出該車次所有店家，各店家附訂單明細摘要
- **AND** 各店家的排列反映同車次內的配送順序

#### Scenario: 不顯示金額

- **WHEN** 使用者檢視單車總表
- **THEN** 單據中不含單價、小計、稅額或總金額等任何金額欄位

### Requirement: 對點單內容與排序

系統 SHALL 提供對點單（`document_type = delivery_note`），每個店家一張 A4，列出該店家訂單明細的品名、數量與單位，MUST NOT 顯示價格，整批單據依車次 → 店家排序。

#### Scenario: 每店家一張 A4

- **WHEN** 使用者產生某車次某日期的對點單
- **THEN** 該車次每個有訂單的店家各產生一頁 A4，列出該店家訂單明細之品名、數量、單位

#### Scenario: 不顯示價格且依車次→店家排序

- **WHEN** 使用者檢視對點單
- **THEN** 單據中不含單價、小計或總金額等任何價格欄位
- **AND** 多張對點單的整體順序依車次，同車次內依店家排列

### Requirement: 揀貨單內容與排序

系統 SHALL 提供揀貨單（`document_type = picking_list`），依車次與倉別印製，彙總同車次同倉別的待揀商品數量（以基本單位計），排序層級 MUST 為車次 → 倉別 → 商品分類 → 商品名稱。

#### Scenario: 四層排序

- **WHEN** 使用者產生某車次的揀貨單
- **THEN** 內容先按車次分組，車次內按倉別分組，倉別內按商品分類排序，同分類內按商品名稱排序

#### Scenario: 彙總同車次同倉別數量

- **WHEN** 同一車次中多筆訂單含相同商品且揀貨倉別相同
- **THEN** 揀貨單將其數量彙總為一列，顯示換算後的基本單位總量

### Requirement: 加工單內容與排序

系統 SHALL 提供加工單（`document_type = processing_list`），列出需分切加工的品項，分為「加工室揀」與「配送揀」兩區塊，「加工室揀」 MUST 排於「配送揀」之前；每列顯示原始數量，「加工後數量」欄位 MUST 列印為空白供加工人員手寫回填，1.0 MUST NOT 將回填值寫回系統。

#### Scenario: 兩區塊順序與原始數量

- **WHEN** 使用者產生某車次某日期的加工單
- **THEN** 單據先呈現「加工室揀」區塊，再呈現「配送揀」區塊
- **AND** 各品項列顯示訂單原始數量與分切規格備註

#### Scenario: 加工後數量欄位列印空白

- **WHEN** 使用者檢視加工單
- **THEN** 每列的「加工後數量」欄位為空白，供加工人員加工時手寫回填

#### Scenario: 手寫回填不回寫系統

- **WHEN** 加工人員於紙本加工單手寫回填加工後數量
- **THEN** 系統不提供將該回填值寫回訂單或任何資料表的功能，訂單數量維持不變

### Requirement: 預覽不影響正式列印記錄

系統 SHALL 允許在任何訂單狀態下預覽任一單據；每次預覽 MUST 產生 PDF 並寫入一筆 `print_previews` 記錄（含 `document_type`、`route_id`、`customer_id`、`warehouse_id`、`target_date`、`previewed_by`、`previewed_at`、`file_asset_id`），且 MUST NOT 寫入或修改 `print_logs`。

#### Scenario: 待處理狀態可預覽

- **WHEN** 訂單仍為 `pending`（尚未派車確認），使用者預覽揀貨單
- **THEN** 系統產生預覽 PDF 並寫入一筆 `print_previews` 記錄
- **AND** `print_logs` 無任何新增或變更

#### Scenario: 預覽記錄內容完整

- **WHEN** 使用者完成一次預覽
- **THEN** 產生的 `print_previews` 記錄包含預覽人、預覽時間、單據類型、目標日期與對應 PDF 的 `file_asset_id`

### Requirement: 正式列印狀態門檻

系統 SHALL 僅允許對狀態為 `processing`（已派車確認）的訂單執行正式列印；每次正式列印 MUST 產生 PDF 並寫入一筆 `print_logs` 記錄且 `is_reprint = false`。對非 `processing` 狀態的訂單執行正式列印 MUST 被拒絕。

#### Scenario: 已派車確認可正式列印

- **WHEN** 目標車次的訂單狀態為 `processing`，使用者執行正式列印
- **THEN** 系統產生 PDF 並寫入一筆 `print_logs` 記錄，`is_reprint = false`
- **AND** 記錄包含 `printed_by` 與 `printed_at`

#### Scenario: 未派車確認拒絕正式列印

- **WHEN** 訂單狀態為 `pending`，使用者嘗試正式列印
- **THEN** 系統拒絕該操作並提示需先完成派車確認
- **AND** 不產生 PDF 且不寫入 `print_logs`

### Requirement: 重印需填寫原因

在已有正式列印記錄後，`dept_admin` 與 `staff` 角色 SHALL 皆可執行重印；重印 MUST 要求填寫 `reprint_reason`，未填寫 MUST 被拒絕；每次重印 MUST 產生新 PDF 並寫入一筆 `print_logs` 記錄且 `is_reprint = true`。

#### Scenario: 填寫原因後重印成功

- **WHEN** 該單據已有正式列印記錄，`staff` 使用者填寫重印原因後執行重印
- **THEN** 系統產生新 PDF 並寫入一筆 `print_logs` 記錄，`is_reprint = true` 且 `reprint_reason` 為所填內容

#### Scenario: 未填寫原因拒絕重印

- **WHEN** 使用者執行重印但未填寫 `reprint_reason`
- **THEN** 系統拒絕該操作
- **AND** 不產生 PDF 且不寫入 `print_logs`

### Requirement: PDF 產生格式與空表規則

系統 SHALL 使用 Gotenberg 將 HTML/CSS 模板轉為 PDF，紙張 MUST 為 A4，字型 MUST 使用 Noto Sans CJK TC；當查詢結果無任何單據內容（空表）時，系統 MUST NOT 產生 PDF，亦 MUST NOT 寫入 `print_logs` 或 `print_previews`。

#### Scenario: 產生符合格式的 PDF

- **WHEN** 使用者執行任一預覽或列印且查詢結果非空
- **THEN** 系統產生 A4 尺寸、以 Noto Sans CJK TC 字型呈現中文內容的 PDF

#### Scenario: 空表不產生 PDF

- **WHEN** 指定條件下無任何符合的訂單或品項，使用者嘗試預覽或列印
- **THEN** 系統提示無內容可列印，不產生 PDF
- **AND** 不寫入任何 `print_logs` 或 `print_previews` 記錄

### Requirement: PDF 留存供查核

每次預覽、正式列印與重印產生的 PDF SHALL 皆關聯一筆 `file_assets` 記錄留存，使日後可查核當時實際列印的內容；`print_logs` 與 `print_previews` MUST 透過 `file_asset_id` 指向對應的 `file_assets`。

#### Scenario: 列印 PDF 關聯留存

- **WHEN** 一次正式列印或預覽完成
- **THEN** 產生的 PDF 建立對應 `file_assets` 記錄，且 `print_logs` / `print_previews` 的 `file_asset_id` 指向該記錄

#### Scenario: 歷史列印內容可查核

- **WHEN** 使用者調閱一筆歷史 `print_logs` 記錄
- **THEN** 可經由其 `file_asset_id` 取得當次列印的原始 PDF，內容與列印當下一致

### Requirement: 列印與預覽記錄查詢

系統 SHALL 提供 `print_logs` 與 `print_previews` 的查詢功能，支援依單據類型（`document_type`）、車次（`route_id`）與日期（`target_date`）篩選；查詢結果中的 PDF SHALL 經由 `file_assets` 的下載 URL 取得。

#### Scenario: 依條件篩選列印記錄

- **WHEN** 使用者以單據類型 `picking_list`、指定車次與日期查詢 `print_logs`
- **THEN** 系統僅回傳符合全部篩選條件的列印記錄，每筆含操作人、時間、`is_reprint` 與 `reprint_reason`

#### Scenario: 經下載 URL 取得 PDF

- **WHEN** 使用者自查詢結果開啟某筆記錄的 PDF
- **THEN** 系統提供對應 `file_assets` 的下載 URL，可下載當次產生的 PDF
