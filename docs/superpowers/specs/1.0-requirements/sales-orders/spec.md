# sales-orders 需求規格

> 來源：原 OpenSpec delta spec（OpenSpec 工作流已停用，2026-08-03 遷移至 docs/）。


## Purpose

定義多公司訂出貨系統的訂單核心：訂單狀態機（`pending` / `processing` / `completed` / `cancelled` / `voided`）、訂單編號取號、下單流程、單位換算、退貨申請，以及客戶專屬商品清單（`customer_products`）。派車看板操作屬 `dispatch`、單據列印屬 `printing`、通知通道與路由屬 `notifications`，本 capability 僅定義訂單本身的狀態與資料行為。

## Requirements

### Requirement: 訂單狀態機

訂單 SHALL 具有狀態 `pending`（待處理）、`processing`（處理中）、`completed`（已完成）、`cancelled`（已取消）、`voided`（已作廢），且 MUST 僅依下列規則流轉：`pending` ⇄ `processing` → `completed`；`pending` → `cancelled`；`completed` → `voided`。`pending` 訂單可編輯、可取消；`processing` 訂單（已派車確認，以 `dispatched_at` 追蹤）不可編輯、不可直接取消；`completed` 訂單不可編輯；`cancelled` 與 `voided` 為終態。派車確認時系統 MUST 將訂單狀態設為 `processing` 並記錄 `dispatched_at` 與 `dispatched_by`。

#### Scenario: 正常生命週期

- **WHEN** 一筆 `pending` 訂單被派車確認，隨後由管理人員標記出貨完成
- **THEN** 訂單狀態依序由 `pending` 轉為 `processing` 再轉為 `completed`
- **AND** 轉為 `processing` 時寫入 `dispatched_at` 與 `dispatched_by`

#### Scenario: 待處理訂單可編輯可取消

- **WHEN** 使用者編輯一筆 `pending` 訂單的明細、預計出貨日或備註，或取消其訂單
- **THEN** 編輯成功；取消後狀態轉為 `cancelled`
- **AND** `cancelled` 訂單不再接受任何狀態異動或編輯

#### Scenario: 非法狀態流轉被拒絕

- **WHEN** 使用者嘗試編輯 `processing` 或 `completed` 訂單、直接取消 `processing` 訂單、或對 `cancelled` / `voided` 訂單執行任何狀態異動
- **THEN** 系統 MUST 拒絕操作並維持原狀態
- **AND** 直接取消 `processing` 訂單時提示須先取消派車退回 `pending`

### Requirement: 取消派車（處理中退回待處理）

派錯車時，`dept_admin` 以上角色 SHALL 可將 `processing` 訂單退回 `pending`，且 MUST 填寫原因。退回時系統 MUST 清除 `dispatched_at` 與 `dispatched_by`，並保留 `route_id` 與 `delivery_sequence`，使訂單停留在看板原位待重新派車。若該車次已有正式列印記錄，操作時 MUST 提示需重新列印。

#### Scenario: 授權角色取消派車成功

- **WHEN** `dept_admin` 以上角色對一筆 `processing` 訂單執行取消派車並填寫原因
- **THEN** 訂單狀態退回 `pending`，`dispatched_at` 與 `dispatched_by` 被清除
- **AND** `route_id` 與 `delivery_sequence` 保持原值不變

#### Scenario: 權限不足或未填原因被拒絕

- **WHEN** `staff` 角色嘗試取消派車，或授權角色未填寫原因即送出
- **THEN** 系統 MUST 拒絕操作，訂單維持 `processing` 且 `dispatched_at` / `dispatched_by` 不變

#### Scenario: 車次已有正式列印記錄時提示重印

- **WHEN** 取消派車的訂單所屬車次已存在正式列印記錄
- **THEN** 系統完成退回 `pending`，並提示操作者該車次單據需重新列印

### Requirement: 訂單作廢與更正

內容錯誤的 `completed` 訂單，`dept_admin` 以上角色 SHALL 可作廢，且 MUST 填寫原因。作廢後狀態為 `voided`，屬終態不可恢復。系統 MUST NOT 提供 `completed` 訂單的直接編輯；更正方式為作廢後重新建立訂單，新單備註 SHALL 記載原單號。

#### Scenario: 授權角色作廢已完成訂單

- **WHEN** `dept_admin` 以上角色對一筆 `completed` 訂單執行作廢並填寫原因
- **THEN** 訂單狀態轉為 `voided`

#### Scenario: 作廢為終態且更正採重建

- **WHEN** 使用者嘗試對 `voided` 訂單執行任何狀態異動或編輯
- **THEN** 系統 MUST 拒絕
- **AND** 更正該訂單時須建立新訂單，新單備註記載原單號

#### Scenario: 非已完成訂單或權限不足不可作廢

- **WHEN** 使用者對 `pending` / `processing` / `cancelled` 訂單執行作廢、由 `staff` 角色執行作廢、或未填寫原因
- **THEN** 系統 MUST 拒絕操作並維持原狀態

### Requirement: 訂單異動事件與稽核記錄

訂單的所有異動 SHALL 寫入 `sales_order_events`，包含 `event_type`（`create` / `edit` / `dispatch` / `dispatch_cancel` / `cancel` / `complete` / `void`）、操作人（`actor_id`）、時間（`created_at`）與原因（`reason`，適用時）。取消派車（`dispatch_cancel`）與作廢（`void`）MUST 另寫入 `audit_logs`。`sales_order_events` 為僅追加記錄，MUST NOT 被修改或刪除。

#### Scenario: 各類異動皆產生事件

- **WHEN** 訂單經歷建立、編輯、派車、取消派車、取消、完成、作廢
- **THEN** 每個動作各寫入一筆對應 `event_type` 的 `sales_order_events`，含操作人、時間與原因

#### Scenario: 取消派車與作廢另寫稽核日誌

- **WHEN** 訂單被取消派車或作廢
- **THEN** 除 `sales_order_events` 外，`audit_logs` 亦新增一筆記錄，含操作人、時間、原因及異動前後狀態

#### Scenario: 事件記錄不可竄改

- **WHEN** 任何角色嘗試修改或刪除既有的 `sales_order_events` 記錄
- **THEN** 系統 MUST 拒絕

### Requirement: 訂單編號與併發取號

訂單編號 SHALL 由「下單來源碼 + 6 位補零自增序號」組成（如 `W000123`），來源碼對應系統級字典「訂單來源」（如 `W` = Web 中台、`A` = App），且在公司內唯一。系統 MUST 以 `order_counters`（依 `company_id` + `source` 一列）樂觀鎖取號：更新時比對 `version`，衝突時重試；取號與訂單建立 MUST 在同一資料庫交易內完成。

#### Scenario: 依來源各自遞增且公司內唯一

- **WHEN** 同一公司先後以來源 `W` 建立兩筆訂單，另以來源 `A` 建立一筆訂單
- **THEN** 訂單編號分別為 `W` 序號連續遞增（如 `W000123`、`W000124`）與 `A` 自有序號（如 `A000001`）
- **AND** 同一公司內不存在重複訂單編號

#### Scenario: 併發取號不重複

- **WHEN** 同一公司同一來源同時有兩個下單請求併發取號
- **THEN** 樂觀鎖 `version` 衝突的一方重試，兩筆訂單取得不同且連續的序號
- **AND** 無重複訂單編號、無序號遺失

#### Scenario: 取號與訂單建立同一交易

- **WHEN** 取號成功後訂單建立失敗導致交易回滾
- **THEN** `order_counters` 的序號一併回滾不消耗，下一次取號仍取得該序號

### Requirement: 不儲存任何金額欄位

訂單、訂單明細、商品與客戶專屬商品 SHALL NOT 包含任何金額欄位（單價、小計、稅額、折扣、總額）。`sales_orders` 僅記錄數量、單位、規格與出貨資訊，不記錄 `subtotal` / `tax_rate` / `tax_inclusive` / `tax_amount` / `discount_amount` / `total_amount` / `currency`；`sales_order_items` 不記錄 `price` / `amount`。系統 SHALL NOT 於任何單據、報表或通知中輸出金額。時間欄位一律 `timestamptz`，顯示時區 UTC+8。

#### Scenario: 訂單不包含金額欄位

- **WHEN** 檢視訂單主檔與明細的資料結構
- **THEN** 不存在任何單價、小計、稅額、折扣或總額欄位

#### Scenario: 單據與報表不含金額

- **WHEN** 產生單車總表、對點單、揀貨單或加工單
- **THEN** 單據內容不含任何金額資訊

### Requirement: 下單流程

業務下單 SHALL 先選擇客戶，再帶出該客戶的 `customer_products`；業務可從商品總表選擇或手打商品名稱，手打時系統 SHALL 詢問是否儲存為該客戶別名。客戶自行下單時客戶固定為自己，且 MUST 只能從自己的 `customer_products` 選擇商品。訂單 MUST 記錄 `source`（訂單來源）與 `expected_delivery_date`（預計出貨日），建立後狀態為 `pending`。選擇預計出貨日時，若客戶設有偏好送貨日（`customers.preferred_delivery_days`，星期一到六）且所選日期非勾選日，系統 SHALL 自動順延至下一個有勾選的日期；客戶未勾選任何日期時維持原選擇。訂單提交後 SHALL 觸發訂單建立通知；通知通道與路由規則屬 `notifications` capability。

#### Scenario: 業務下單並手打商品

- **WHEN** 業務選擇客戶後，系統帶出其 `customer_products`，業務改以手打輸入商品名稱並提交
- **THEN** 系統詢問是否將手打名稱儲存為該客戶的商品別名
- **AND** 訂單記錄 `source`、`expected_delivery_date` 與負責業務 `sales_rep_id`，狀態為 `pending`

#### Scenario: 客戶只能從專屬清單下單

- **WHEN** 客戶帳號下單時嘗試選擇不在自己 `customer_products` 中的商品
- **THEN** 系統 MUST 拒絕該商品加入訂單
- **AND** 客戶可從自己的 `customer_products` 正常選品下單

#### Scenario: 提交後觸發通知

- **WHEN** 訂單提交建立成功
- **THEN** 系統觸發訂單建立通知（實際通道與收件路由由 `notifications` 決定）

#### Scenario: 非偏好送貨日自動順延

- **WHEN** 客戶偏好送貨日為星期二與星期四，業務選擇星期三為預計出貨日下單
- **THEN** `expected_delivery_date` 自動順延為星期四（下一個勾選日）
- **AND** 業務於確認畫面看到順延後的日期

#### Scenario: 偏好送貨日已勾選時維持原日

- **WHEN** 客戶偏好送貨日含所選的預計出貨日
- **THEN** `expected_delivery_date` 維持原選擇，不順延

#### Scenario: 客戶未設定偏好送貨日

- **WHEN** 客戶 `preferred_delivery_days` 為空（未勾選任何日期）
- **THEN** 系統不自動順延，維持業務所選日期

### Requirement: 訂單明細與單位換算

每筆訂單明細 SHALL 包含商品、數量（`qty`）、單位（`unit`）、分切規格（`cutting_spec_id`）、特殊切法備註（`special_cut_note`）與倉別（`warehouse_id`），MUST NOT 包含單價或金額欄位。下單選擇單位後系統 SHALL 依 `product_units` 的 `conversion_rate` 自動換算數量，並於 `base_qty` 保存換算為基本單位（`is_base = true`）的數量，供揀貨單與加工單彙總。

#### Scenario: 選擇非基本單位自動換算

- **WHEN** 商品基本單位為 kg，換算率 1 條 = 0.6 kg，下單選擇單位「條」、數量 10
- **THEN** 明細 `qty` 為 10、`unit` 為條，`base_qty` 自動換算為 6（kg）

#### Scenario: 選擇基本單位時 base_qty 等於數量

- **WHEN** 下單直接選擇商品的基本單位（`is_base = true`）
- **THEN** `base_qty` 等於 `qty`

#### Scenario: 明細攜帶分切與倉別資訊

- **WHEN** 下單時為明細指定分切規格、特殊切法備註與倉別
- **THEN** 明細保存 `cutting_spec_id`、`special_cut_note` 與 `warehouse_id`，供後續單據使用

### Requirement: 退貨申請（客戶發起）

客戶 SHALL 可從 App 發起退貨申請，品項來源為「歷史訂單勾選」或「客戶專屬商品清單選擇」二擇一（1.0 兩者並存，UX 依試用回饋收斂）。申請 MUST 包含數量與原因，SHALL 可附照片（多張，經 `file_assets` 留存）與備註。送出後狀態為 `pending`，由主責業務（`customers.default_sales_rep_id`，未設定則部門 `dept_admin`）審核為 `approved` 或 `rejected`（拒絕 MUST 填原因）。通過後客戶端 SHALL 顯示退貨證明畫面，可出示給配送司機。退貨 MUST NOT 修改原訂單狀態；審核結果推播發起客戶帳號（屬 notifications capability）。公司端不建置配送專屬頁面，取貨由業務自行通知配送。

#### Scenario: 客戶從歷史訂單發起退貨

- **WHEN** 客戶在訂單歷史中選擇訂單並勾選品項、填寫數量與原因、上傳照片後送出
- **THEN** 系統建立 `pending` 退貨申請並關聯原訂單與品項

#### Scenario: 客戶從專屬商品清單發起退貨

- **WHEN** 客戶從自己的 `customer_products` 選擇品項發起退貨
- **THEN** 系統建立 `pending` 退貨申請（不關聯訂單）

#### Scenario: 業務審核通過顯示退貨證明

- **WHEN** 主責業務將退貨申請審核為 `approved`
- **THEN** 客戶端顯示退貨證明畫面（含申請編號與通過狀態），並推播通知發起帳號

#### Scenario: 業務拒絕需填原因

- **WHEN** 主責業務將退貨申請審核為 `rejected`
- **THEN** 系統要求填寫拒絕原因，客戶端顯示拒絕狀態與原因

#### Scenario: 退貨不改變原訂單狀態

- **WHEN** 退貨申請建立或審核完成
- **THEN** 原訂單 `status` 維持不變，退貨異動僅寫入退貨申請與稽核日誌

### Requirement: 客戶專屬商品清單（customer_products）

系統 SHALL 維護 `customer_products`，記錄 `customer_id`、`product_id`、`alias_name`、`default_qty`、`custom_price` 與 `default_cut_note`。同一客戶對同一商品 MUST 只有一筆（`customer_id + product_id` 唯一，以部分唯一索引 `WHERE deleted_at IS NULL` 實作）。清單首次由業務在客戶檔案頁手動建立，或下單時自動加入；刪除採軟刪除（`deleted_at`）。`default_qty` 設為 0 時保留於清單，但後續單據不顯示。

#### Scenario: 建立與預設值帶入

- **WHEN** 業務在客戶檔案頁為客戶新增專屬商品，或下單選用商品總表商品時自動加入清單
- **THEN** 系統建立 `customer_products` 記錄
- **AND** 後續下單選用該商品時帶入 `alias_name`、`default_qty`、`custom_price` 與 `default_cut_note` 作為預設值

#### Scenario: 同客戶同商品不可重複

- **WHEN** 為同一客戶對同一商品建立第二筆 `customer_products`
- **THEN** 系統 MUST 拒絕，維持一客戶一商品一筆（一個別名）

#### Scenario: 數量為 0 保留但單據不顯示

- **WHEN** 清單項目的 `default_qty` 被設為 0
- **THEN** 該項目仍保留於客戶專屬清單
- **AND** 後續產生的單據不顯示該項目

#### Scenario: 軟刪除後可重建

- **WHEN** 使用者刪除某筆 `customer_products` 後，再為同客戶同商品重新建立
- **THEN** 原記錄標記 `deleted_at` 並自清單隱藏，新記錄可成功建立不觸發唯一衝突
- **AND** 歷史訂單明細不受影響，仍正常顯示
