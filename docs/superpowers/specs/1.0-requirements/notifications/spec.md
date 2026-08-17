# notifications 需求規格

> 來源：原 OpenSpec delta spec（OpenSpec 工作流已停用，2026-08-03 遷移至 docs/）。


## Requirements

### Requirement: 通知範本管理

系統 SHALL 提供 `notification_templates` 實體，供管理者依公司（`company_id`）與部門（`department_id`）維護通知範本。範本 MUST 包含 `code`、`name`、`channel`、`subject`、`body`、`locale`、`is_active` 欄位；`channel` 限 `fcm`、`in_app` 兩種之一（1.0 不使用 Email）。`body` MUST 支援 `{{變數}}` 模板語法，發送時以實際資料代入。`locale` 欄位為多語預留，1.0 僅使用繁體中文。範本適用軟刪除；停用的範本 MUST NOT 被用於發送。

#### Scenario: 建立公司層級 FCM 範本

- **WHEN** 管理者為某公司建立 `channel = fcm` 的通知範本，`body` 含 `{{order_no}}` 變數
- **THEN** 系統儲存範本並於發送時將 `{{order_no}}` 代換為實際訂單編號

#### Scenario: 停用範本不再用於發送

- **WHEN** 管理者將範本的 `is_active` 設為停用
- **THEN** 後續通知發送 MUST NOT 選用該範本

#### Scenario: 不支援的通道被拒絕

- **WHEN** 建立或更新範本時 `channel` 非 `fcm`、`in_app` 之一（如 `email`）
- **THEN** 系統拒絕該操作並回傳驗證錯誤

### Requirement: 通知記錄

系統 SHALL 以 `notifications` 實體保存每則實際發送記錄，欄位 MUST 包含 `company_id`、`department_id`、`user_id`、`template_id`、`channel`、`title`、`content`、`payload`、`status`、`sent_at`、`read_at`。`status` 限 `pending`、`sent`、`failed`、`read` 四種。通知記錄不適用軟刪除，任何使用者 MUST NOT 刪除通知記錄。

#### Scenario: 發送成功記錄狀態與時間

- **WHEN** 一則通知成功送出
- **THEN** 其 `status` 為 `sent` 且 `sent_at` 記錄發送時間

#### Scenario: 通知記錄不可刪除

- **WHEN** 任何角色嘗試刪除一筆 `notifications` 記錄
- **THEN** 系統拒絕該操作，記錄仍完整保留

### Requirement: 自動通知路由

系統 SHALL 依下列規則決定自動通知路由：**訂單建立**——業務下單時推播給該客戶的全部**子帳號**（`users.customer_id` = 訂單客戶且 `is_primary = false`；主帳號為管理用途、不接收業務通知）；客戶自行下單不另行通知。**客戶專屬商品新增**（後台 / Web 建立時）——推播給該客戶 `default_sales_rep_id` 主責業務（未設定則部門 `dept_admin`；是否需業務檢核，待定）。**退貨審核結果**——推播給發起退貨的客戶帳號（必然為子帳號）。

#### Scenario: 業務下單推播客戶子帳號

- **WHEN** 業務人員為某客戶建立訂單
- **THEN** 訂單建立通知發送給該客戶的全部子帳號

#### Scenario: 客戶自行下單不另行通知

- **WHEN** 客戶自行下單成功
- **THEN** 系統不因訂單建立額外發送通知給業務

#### Scenario: 後台新增客戶專屬商品推播主責業務

- **WHEN** 後台 / Web 為客戶建立新的 `customer_products`
- **THEN** 系統推播通知給該客戶 `default_sales_rep_id`（未設定則部門 `dept_admin`）

#### Scenario: 退貨審核結果推播客戶

- **WHEN** 業務完成退貨申請審核
- **THEN** 系統將審核結果推播給發起退貨的客戶帳號

### Requirement: 促銷推播（與公告分離）

系統 SHALL 提供促銷推播：`promo_tags` 為部門級促銷分類標籤（`code`、`name`、`is_active`）；商品與客戶專屬商品 SHALL 可標記多個標籤（`promo_tag_ids`）；客戶主檔 SHALL 可套用（訂閱）分類標籤（`customers.promo_tag_ids`）。後台依分類選擇要推播的客戶群，經 FCM + 站內發送。公告（`announcements`）顯示於 App 首頁消息區、所有客戶可見，與促銷推播互不取代（推播與公告是兩件事）。

#### Scenario: 依分類標籤選擇客戶群推播

- **WHEN** 後台選取分類標籤 T 並發送促銷推播
- **THEN** 所有 `customers.promo_tag_ids` 包含 T 的客戶的子帳號收到推播
- **AND** 未套用標籤 T 的客戶不收到該推播，但仍可在首頁消息看到公告

#### Scenario: 商品標記與客戶套用標籤

- **WHEN** 後台為商品 / 客戶專屬商品標記促銷標籤，並於客戶主檔套用分類
- **THEN** 系統儲存 `promo_tag_ids`，供後續選群推播使用

### Requirement: FCM 推播與裝置管理

系統 SHALL 透過 FCM 發送 App 推播，用於訂單狀態更新與派車通知。系統 MUST 提供 `user_devices` 實體（`user_id`、`platform`、`fcm_token`、`device_name`、`last_seen_at`），供 App 註冊與註銷裝置 token；`platform` 限 `android`、`ios`。發送推播時 MUST 以接收者所有已註冊裝置的 `fcm_token` 為目標。

#### Scenario: 註冊裝置 token

- **WHEN** App 使用者登入後提交裝置的 `fcm_token`、`platform` 與 `device_name`
- **THEN** 系統建立 `user_devices` 記錄，後續推播可送達該裝置

#### Scenario: 註銷裝置 token

- **WHEN** App 使用者登出或移除裝置時請求註銷 `fcm_token`
- **THEN** 系統刪除對應 `user_devices` 記錄，後續不再對該裝置發送推播

### Requirement: FCM 失效 token 清除

系統 SHALL 在 FCM 回報裝置 token 失效（`unregistered` 或 `invalid`）時，自動刪除對應的 `user_devices` 記錄，避免持續對無效裝置發送推播。

#### Scenario: unregistered token 自動刪除

- **WHEN** 發送推播時 FCM 回報某 `fcm_token` 為 `unregistered`
- **THEN** 系統刪除該 token 對應的 `user_devices` 記錄

#### Scenario: invalid token 自動刪除

- **WHEN** 發送推播時 FCM 回報某 `fcm_token` 為 `invalid`
- **THEN** 系統刪除該 token 對應的 `user_devices` 記錄，且同一使用者的其他有效裝置不受影響

### Requirement: 發送失敗處理

通知發送失敗時，系統 SHALL 將該筆 `notifications.status` 標記為 `failed` 並記錄失敗原因。1.0 MUST NOT 實作任何自動重試機制或重試佇列。

#### Scenario: 失敗標記與原因記錄

- **WHEN** 通知發送因任何通道錯誤而失敗
- **THEN** 該筆通知 `status` 為 `failed` 且記錄失敗原因

#### Scenario: 失敗後不自動重試

- **WHEN** 一筆通知已被標記為 `failed`
- **THEN** 系統不會在後續任何時間自動重新發送該通知

### Requirement: Web 通知中心

系統 SHALL 提供 Web 通知中心（站內通知），顯示目前使用者的系統通知列表（`channel = in_app`）與未讀數量，並支援將通知標記為已讀。標記已讀時 MUST 將 `status` 更新為 `read` 並記錄 `read_at` 時間。使用者僅能檢視與操作自己的通知。

#### Scenario: 檢視通知列表與未讀數

- **WHEN** 使用者開啟通知中心
- **THEN** 系統顯示該使用者的通知列表及未讀通知數量，不包含其他使用者的通知

#### Scenario: 標記已讀流轉

- **WHEN** 使用者將一則未讀通知標記為已讀
- **THEN** 該通知 `status` 由 `sent` 變更為 `read`，`read_at` 記錄已讀時間，且未讀數量隨之減少
