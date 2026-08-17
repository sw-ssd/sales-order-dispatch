# master-data 需求規格

> 來源：原 OpenSpec delta spec（OpenSpec 工作流已停用，2026-08-03 遷移至 docs/）。


主檔管理：客戶（含編號取號、多登入帳號與偏好送貨日）、客戶地址簿／聯絡人、商品與單位換算、部門級主檔（倉別／車次／分切規格／商品分類）、metadicts 字典，以及全系統軟刪除慣例。

## Requirements

### Requirement: 客戶建立與編號取號

系統 SHALL 於建立客戶時，以公司定義的客戶編號前綴（`companies.customer_code_prefix`）加上自增序號（補零至 6 位）產生 `customer_code`（例如 `TY000123`），透過 `customer_counters` 樂觀鎖取號，且取號與客戶建檔 MUST 於同一資料庫交易完成；`company_id + customer_code` MUST 唯一，且 `customer_code` 建立後不可修改。

#### Scenario: 正常取號建立客戶

- **WHEN** dept_admin 於公司前綴為 `TY`、計數器 `next_seq` 為 123 的狀態下建立客戶
- **THEN** 客戶建立成功且 `customer_code` 為 `TY000123`
- **AND** `customer_counters.next_seq` 遞增為 124 且與客戶建檔屬於同一交易

#### Scenario: 樂觀鎖衝突自動重試

- **WHEN** 兩個請求同時以相同的 `customer_counters.version` 取號
- **THEN** 其中一個更新因 version 衝突失敗並重試
- **AND** 最終兩位客戶取得不同且連續的 `customer_code`，無重複或跳號遺失

#### Scenario: customer_code 不可修改

- **WHEN** 任何角色對已存在的客戶送出修改 `customer_code` 的更新請求
- **THEN** 系統拒絕該欄位異動，既有 `customer_code` 維持不變

#### Scenario: 公司內唯一性

- **WHEN** 同一公司內嘗試建立與既有未刪除客戶相同 `customer_code` 的資料
- **THEN** 資料庫唯一性限制阻止寫入，操作失敗
- **AND** 不同公司之間的 `customer_code` 各自獨立計數、互不干擾

### Requirement: 客戶建立同步產生登入帳號與臨時密碼

系統 SHALL 於建立客戶時同步建立**主帳號**（`users`，`customer_id` 關聯客戶主檔、角色 `customer`、`account_name` 預設為客戶名稱、`is_primary = true`），並同時自動建立 **1 個業務子帳號**（`is_primary = false`、`account_name` 預設為「客戶名稱（業務）」），各產生隨機臨時密碼，臨時密碼效期 MUST 為 24 小時（`temp_password_expires_at`），客戶首次登入 MUST 被強制要求修改密碼後才能繼續使用；建立完成時主帳號帳密交付店家（由業務轉交）、**業務子帳號憑證交付該客戶所屬業務**（見 sales-orders / App 新增客戶流程）。**其他子帳號**（如主廚）由店家用主帳號新增（見 identity-access capability）或後台新增：各帳號 `account_name` 於客戶內唯一、密碼獨立管理，新增或停用任一帳號不影響客戶主檔與其他帳號；每客戶 MUST 恰有一個 `is_primary = true` 的主帳號；自動附帶的業務子帳號於店家管理清單灰化、店家不可管理（見 identity-access capability）。

#### Scenario: 建立客戶產生帳號與臨時密碼

- **WHEN** dept_admin 成功建立一位新客戶
- **THEN** 系統同步建立一筆 `users` 客戶帳號並關聯該客戶
- **AND** 產生臨時密碼且其效期為建立後 24 小時

#### Scenario: 首次登入強制修改密碼

- **WHEN** 客戶以 `customer_code` 與未過期臨時密碼首次登入
- **THEN** 系統要求先設定新密碼（最少 8 字元）
- **AND** 完成修改前不得進入其他功能

#### Scenario: 臨時密碼過期

- **WHEN** 客戶於臨時密碼建立超過 24 小時後才嘗試登入
- **THEN** 登入被拒絕並提示臨時密碼已失效
- **AND** 需由 dept_admin 以上重新產生臨時密碼

### Requirement: 客戶主檔維護

系統 SHALL 提供客戶主檔的新增、查詢、修改與軟刪除功能，欄位包含 `tax_id`、`payment_method_id`、`settlement_method_id`、`customer_type_id`、`invoice_type_id`、`default_sales_rep_id`（預設負責業務，供通知路由使用）、`preferred_delivery_days`（偏好送貨日：星期一到六核取）與 `promo_tag_ids`（套用促銷分類標籤，見 notifications capability），並依角色與部門資料範圍限制可見與可操作的客户。

#### Scenario: 設定預設負責業務

- **WHEN** dept_admin 編輯客戶並指定 `default_sales_rep_id` 為同部門的業務帳號
- **THEN** 客戶主檔儲存該預設負責業務，供後續下單與通知路由使用

#### Scenario: 設定偏好送貨日

- **WHEN** dept_admin 為客戶勾選星期二與星期四為送貨日
- **THEN** 客戶主檔儲存 `preferred_delivery_days`，後續下單時非勾選日自動順延至下一個勾選日（順延規則屬 sales-orders capability）

#### Scenario: 列表預設排除已刪除客戶

- **WHEN** 使用者查詢客戶列表且未開啟「顯示已刪除」
- **THEN** 結果僅包含 `deleted_at IS NULL` 的客戶

### Requirement: 客戶地址簿與聯絡人

系統 SHALL 支援每位客戶維護多筆地址（`customer_addresses`，`type` 為 `shipping` / `billing` / `other`，含 `is_default` 標記）與多筆聯絡人（`customer_contacts`，含 `name`、`title`、`email`、`phone`、`is_default`），地址與聯絡人 MUST 隸屬客戶並繼承其公司與部門歸屬。

#### Scenario: 新增多筆不同類型地址

- **WHEN** 使用者為同一客戶分別新增 `shipping` 與 `billing` 地址
- **THEN** 兩筆地址皆成功建立並各自保留類型
- **AND** 每種用途可查詢到對應地址

#### Scenario: 標記預設地址

- **WHEN** 客戶已有多筆 `shipping` 地址，使用者將其中一筆設為 `is_default`
- **THEN** 該筆地址成為預設出貨地址
- **AND** 同類型地址至多一筆為預設

#### Scenario: 地址與聯絡人軟刪除

- **WHEN** 使用者刪除客戶的某筆地址或聯絡人
- **THEN** 該筆資料寫入 `deleted_at` 且不再出現於預設查詢結果
- **AND** 既有歷史單據關聯的資料不受影響

### Requirement: 商品主檔維護

系統 SHALL 提供商品主檔的新增、查詢、修改與軟刪除，商品隸屬部門（`department_id`）；`code` 於同一部門內 MUST 唯一（搭配軟刪除部分唯一索引）；商品 SHALL 記錄 `inventory_warehouse_id`（產品分庫）與 `picking_warehouse_id`（揀貨倉別）；商品可用的分切規格 MUST 以關聯表 `product_cutting_specs`（`product_id` + `cutting_spec_id`）維護，不得使用陣列欄位。

#### Scenario: 部門內商品編號唯一

- **WHEN** 於同一部門建立與既有未刪除商品相同 `code` 的商品
- **THEN** 建立失敗並回報編號重複
- **AND** 不同部門可使用相同的商品 `code`

#### Scenario: 軟刪除後可重建同編號商品

- **WHEN** 某商品已被軟刪除，使用者於同部門以相同 `code` 建立新商品
- **THEN** 建立成功（部分唯一索引僅約束 `deleted_at IS NULL` 的資料）

#### Scenario: 關聯分切規格

- **WHEN** 使用者為商品勾選多個部門的分切規格
- **THEN** 系統於 `product_cutting_specs` 建立對應關聯列
- **AND** 商品查詢可取得其可用分切規格清單

### Requirement: 商品單位與換算率

系統 SHALL 以 `product_units` 維護每項商品的單位設定：`unit_code` 對應 metadicts 單位字典，每項商品 MUST 恰好有一個 `is_base = true` 的基本單位，其餘單位以 `conversion_rate` 記錄換算為基本單位的比率（如 1 條 = 0.6 kg 則 `conversion_rate` 為 0.6），並以 `sort_order` 排序。

#### Scenario: 設定基本單位與換算單位

- **WHEN** 使用者為商品設定基本單位 `kg`（`is_base = true`）與換算單位 `條`（`conversion_rate = 0.6`）
- **THEN** 兩筆單位皆建立成功
- **AND** 1 條換算為 0.6 kg 基本單位數量

#### Scenario: 每商品僅一個基本單位

- **WHEN** 商品已有基本單位，使用者嘗試再將另一單位設為 `is_base = true`
- **THEN** 系統阻止出現第二個基本單位（拒絕操作或要求先異動原基本單位）

#### Scenario: 換算率邊界

- **WHEN** 使用者嘗試將 `conversion_rate` 設為 0 或負數
- **THEN** 系統拒絕儲存並提示換算率必須為正數

### Requirement: 部門級主檔獨立管理

倉別（`warehouses`）、車次（`routes`）、分切規格（`cutting_specs`）與商品分類（`product_categories`）SHALL 為部門級實體，由各部門獨立新增、修改與軟刪除，跨部門互不可見、不可選用；分切規格 SHALL 記錄其加工／配送揀貨歸屬。

#### Scenario: 部門各自維護倉別

- **WHEN** 甲部門與乙部門各自建立倉別資料
- **THEN** 各部門僅能查詢與選用自己部門的倉別
- **AND** 甲部門使用者無法在商品或單據上選用乙部門的倉別

#### Scenario: 軟刪除車次後不可再選用

- **WHEN** dept_admin 軟刪除某車次
- **THEN** 該車次不再出現於新單據的車次選項
- **AND** 既有訂單關聯的該車次名稱仍可正常顯示

### Requirement: metadicts 字典兩層可見性

系統 SHALL 以單一 `metadicts` 表（`type`、`code`、`display_name`、`department_id`、`sort_order`、`is_active`）承載「系統預設 + 部門擴充」字典：`department_id IS NULL` 為系統預設、所有部門可見，`department_id = <部門 ID>` 為部門擴充、僅該部門可見；查詢 MUST 回傳系統預設與當前部門擴充的聯集並排除已軟刪除資料；`type + code + department_id` MUST 唯一（`department_id` 的 NULL 以 PG15 `NULLS NOT DISTINCT` 或部分唯一索引處理）；1.0 不支援部門覆寫或停用系統預設值。

#### Scenario: 查詢合併系統預設與部門擴充

- **WHEN** 甲部門使用者查詢「單位」字典，系統預設有 `kg`，甲部門擴充有 `條`，乙部門擴充有 `箱`
- **THEN** 結果包含 `kg` 與 `條`
- **AND** 不包含乙部門的 `箱`

#### Scenario: 系統預設與部門擴充可共用 code

- **WHEN** 系統預設已存在 `type=unit, code=pk, department_id=NULL`
- **THEN** 甲部門仍可建立 `type=unit, code=pk, department_id=甲部門`
- **AND** 同一部門內重複建立相同 `type + code` 則被拒絕

#### Scenario: 1.0 不支援覆寫系統預設

- **WHEN** dept_admin 嘗試修改或停用 `department_id IS NULL` 的系統預設字典值
- **THEN** 系統拒絕該操作

### Requirement: metadicts 維護權限

`super` SHALL 可維護系統級預設字典（`department_id IS NULL`）；`dept_admin` 與 `staff` SHALL 僅可維護自己所屬部門的擴充值，不得維護系統預設或其他部門的資料。

#### Scenario: super 維護系統級字典

- **WHEN** super 新增一筆 `department_id IS NULL` 的字典值
- **THEN** 建立成功且所有部門查詢皆可見

#### Scenario: dept_admin 越權維護被拒

- **WHEN** dept_admin 嘗試新增系統級字典或修改其他部門的擴充值
- **THEN** 系統依權限拒絕並回傳禁止存取錯誤

### Requirement: 軟刪除慣例

業務實體（含 `customers`、`customer_addresses`、`customer_contacts`、`products`、`warehouses`、`routes`、`cutting_specs`、`product_categories`、`metadicts`）SHALL 統一以 `deleted_at TIMESTAMP NULL` 作為軟刪除標記；業務唯一性限制 MUST 搭配部分唯一索引（`WHERE deleted_at IS NULL`）；列表查詢預設排除已刪除資料；復原 MUST 清空 `deleted_at` 並寫入稽核日誌；硬刪除 MUST 僅限 `super` 於特定管理介面執行並強制留存稽核紀錄；`notifications`、`audit_logs`、`sales_order_events` 不適用軟刪除、不可刪除。

#### Scenario: 軟刪除主檔的歷史關聯仍可顯示

- **WHEN** 某客戶或商品已被軟刪除，使用者檢視引用該主檔的歷史訂單或單據
- **THEN** 歷史資料仍正常顯示其名稱與編號
- **AND** 該主檔不可再被選用於新單據

#### Scenario: 復原軟刪除資料

- **WHEN** 管理員復原一筆已軟刪除的商品
- **THEN** 該商品 `deleted_at` 清空並重新出現於列表
- **AND** 稽核日誌記錄復原操作（操作人、時間、異動摘要）

#### Scenario: 硬刪除僅限 super

- **WHEN** 非 super 角色嘗試硬刪除任何主檔資料
- **THEN** 系統拒絕
- **AND** super 執行硬刪除時系統強制寫入稽核紀錄

#### Scenario: 記錄型實體不可刪除

- **WHEN** 任何角色（含 super）嘗試刪除 `notifications`、`audit_logs` 或 `sales_order_events` 的資料
- **THEN** 系統不提供刪除途徑，資料永久保留
