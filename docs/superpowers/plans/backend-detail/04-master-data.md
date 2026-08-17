# 04 — 業務主檔(Master Data)後端細部實作計畫

> **文件標題**:業務主檔後端細部實作計畫(客戶 / 地址聯絡人 / 商品 / 部門級主檔 / 客戶專屬商品 / 檔案資產 / QR Code)
> **對應原計畫 Task 編號**:Task 3.1–3.6、Task 3.8(Task 3.8 原含前端元件 `frontend/src/components/customer-qrcode.tsx`,本文件僅涵蓋後端部分,前端元件屬前端計畫範圍,於 3.8 節註明;Task 3.7 全為前端,不屬本文件;Task 3.9 為 Phase 驗收,依 `00-index.md` §5 不拆)
> **對應規格書 v1.0.34 章節**:§4.2(客戶帳號一主多子、臨時密碼、QR 登入)、§5.4(統一軟刪除慣例)、§9.4(業務流程:App 新增客戶與憑證交付)、§14(安全:檔案上傳白名單、Rate Limiting);並對應 capability 規格 `1.0-requirements/master-data`、`1.0-requirements/file-assets`、`1.0-requirements/identity-access`(QR Code 登入)、`1.0-requirements/sales-orders`(customer_products)
> **相依文件**:`01-auth.md`(臨時密碼機制、JWT/Session、主帳號業務 API 403)、`02-tenancy-users.md`(companies.customer_code_prefix、users 實體、角色權限)、`03-metadicts-audit.md`(metadicts 字典、稽核日誌寫入)、`05-sales-orders.md`(下單手打別名、base_qty 換算、偏好送貨日順延的消費端)、`07-notifications.md`(promo_tags 推播分群的消費端)
> **相關決策**:D7(客戶編號取號)、D10(軟刪除)、D18(稽核同事務)、D22(客戶帳號一主多子)、D24(促銷分類標籤)、D26(偏好送貨日)

---

## 0. 共通規則(本文件適用,各子功能不重複贅述)

1. **軟刪除**:本文件所有業務實體統一 `deleted_at TIMESTAMP NULL` + 部分唯一索引(`WHERE deleted_at IS NULL`);列表查詢預設排除已刪除;復原 = 清空 `deleted_at` + 同事務寫稽核(D10)。
2. **多租戶**:所有實體帶 `company_id` / `department_id`;功能面由 Casbin 把關(domain = company_id),資料面由 RLS 注入(`app.current_company_id` 等)作最後防線;dept_admin / staff 僅能操作所屬部門資料。
3. **交易與稽核**:取號 + 建檔、主檔異動 + audit log,皆同一 DB 交易,同成功同失敗(D18);各子功能「實作邏輯」欄以「交易邊界」標明範圍。
4. **錯誤碼約定**:統一以 Connect code 表述:`unauthenticated` / `permission_denied` / `not_found` / `failed_precondition`(含樂觀鎖衝突)/ `invalid_argument` / `already_exists`。
5. **權限基準**:主檔管理開放 `dept_admin` / `staff`(限所屬部門);`customer` 主帳號對業務 API 一律 `permission_denied`(見 `01-auth.md`)。
6. 完整共通規則見 `00-index.md` §3,本節僅摘錄本文件直接引用者。

---

## 1. Task 3.1:客戶主檔 API

### 子功能 3.1.1: customers Ent schema(含 default_sales_rep_id)

- **目標**: 定義 `customers` 實體結構、索引與軟刪除慣例,作為所有客戶功能的資料基礎。
- **檔案**: Create `backend/internal/domain/customers/schema.go`(Ent schema);Create `backend/internal/domain/customers/customer_counter.go`(見 3.1.3)
- **介面**: Ent 實體 `Customer`,欄位:`id`(uuid)、`company_id`、`department_id`、`customer_code`(文字,見 3.1.3)、`name`、`tax_id`、`payment_method_id` / `settlement_method_id` / `customer_type_id` / `invoice_type_id`(皆指向 metadicts 字典值,見 `03-metadicts-audit.md`)、`default_sales_rep_id`(指向 users,預設負責業務,供通知路由使用)、`preferred_delivery_days`(JSON 布林陣列,長度 6,對應星期一至星期六,見 3.1.5)、`promo_tag_ids`(JSON 陣列,見 3.1.5)、`created_by` / `updated_by`、`created_at` / `updated_at`、`deleted_at`。
- **實作邏輯**:
  1. 定義上述欄位與型別;`preferred_delivery_days` 與 `promo_tag_ids` 採 JSON 欄位,預設值分別為「全 false」與空陣列。
  2. 建立部分唯一索引:`(company_id, customer_code) WHERE deleted_at IS NULL`(公司內唯一,軟刪除後可重用編號)。
  3. 建立一般索引:`(department_id) WHERE deleted_at IS NULL`(部門列表查詢用)、`(default_sales_rep_id)`(通知路由反查用)。
  4. `default_sales_rep_id` 設外鍵至 users,允許 NULL(建檔時尚未指派業務的情境),但 3.1.4 的業務子帳號交付流程要求建檔時提供。
  5. 字典類欄位(payment / settlement / customer_type / invoice_type)不設硬外鍵,於 3.1.2 的寫入路徑驗證其存在於 metadicts 且符合 type。
- **錯誤處理**: 本體為 schema 定義,無執行期錯誤;migration 失敗屬部署議題,不在此列。
- **驗收**:
  - [ ] migration 後 `customers` 表存在,欄位與索引如上。
  - [ ] 同公司同 `customer_code` 兩筆未刪除資料無法並存;一筆軟刪除後同碼可再建。
  - [ ] `default_sales_rep_id` 可為 NULL,亦可指向同公司 users。

### 子功能 3.1.2: 客戶 CRUD + 軟刪除 + 關鍵字篩選

- **目標**: 提供客戶主檔的新增、查詢、修改、軟刪除與復原,並支援關鍵字篩選。`相依: 3.1.1`
- **檔案**: Create `backend/internal/domain/customers/service.go`、`backend/internal/domain/customers/repo.go`
- **介面**: Connect-RPC `CustomerService`:
  - `ListCustomers(ListCustomersRequest) returns (ListCustomersResponse)` — Request 含 `keyword`、`page` / `page_size`、`include_deleted`、`sort`;Response 含 `customers[]` 與分頁 `meta`(依規格 API 分頁標準)。
  - `GetCustomer(GetCustomerRequest) returns (GetCustomerResponse)` — 以 `id` 取單筆。
  - `CreateCustomer(CreateCustomerRequest) returns (CreateCustomerResponse)` — 內部串接 3.1.3 取號與 3.1.4 建帳號;Response 除客戶資料外,一次性回傳主帳號與業務子帳號的帳號名稱與臨時密碼,以及帳號管理深層連結(規格 §9.4)。
  - `UpdateCustomer(UpdateCustomerRequest) returns (UpdateCustomerResponse)` — 欄位式更新;`customer_code` 不在可更新欄位集合內。
  - `DeleteCustomer(DeleteCustomerRequest) returns (DeleteCustomerResponse)` — 軟刪除。
  - `RestoreCustomer(RestoreCustomerRequest) returns (RestoreCustomerResponse)` — 復原。
- **實作邏輯**:
  1. 所有入口先驗身分與角色:未登入拒絕;`dept_admin` / `staff` 以外角色(含 customer 主帳號)拒絕;RLS 注入當前 company / department / data_scope。
  2. List:預設僅回 `deleted_at IS NULL`;`include_deleted = true` 時加註刪除標記一併回傳(供管理介面「顯示已刪除」開關)。`keyword` 非空時對 `name`、`customer_code`、`tax_id` 做不區分大小寫的模糊比對(三者 OR)。分頁與排序依 Request 參數,排序欄位白名單化(僅允許 name / customer_code / created_at)。
  3. Get:以 id + 租戶範圍查詢;不存在或已軟刪除(且未要求含刪除)視同找不到。
  4. Create:驗證必填(name)與字典欄位合法性(每個 metadicts 引用須存在、`is_active`、type 相符);`default_sales_rep_id` 若提供,須為同公司同部門、具業務身分且未停用的使用者;進入 3.1.3 / 3.1.4 的複合交易。`customer_code` 一律由系統產生,Request 攜帶該欄位視為無效輸入。
  5. Update:僅允許白名單欄位;Request 嘗試異動 `customer_code` 一律拒絕(D7:建立後不可修改),不回傳成功而忽略。`default_sales_rep_id` 異動須重新驗證目標業務合法性;主責業務變更時,業務子帳號憑證移交屬後台操作(見 3.1.4 步驟 6),不在本 RPC 自動處理。
  6. Delete:交易邊界 = 更新 `deleted_at` + 寫稽核日誌(操作人、時間、異動摘要),同成功同失敗(D18)。軟刪除不連動刪除地址、聯絡人、專屬商品與帳號;歷史單據關聯不受影響。
  7. Restore:清空 `deleted_at` + 同事務寫稽核;復原前重新檢查 `(company_id, customer_code)` 部分唯一索引不衝突(正常情況不會衝突,因編號在建檔時已唯一且不可改)。
- **錯誤處理**:
  - 未登入或 session 失效:`unauthenticated`。
  - 角色不符、跨部門操作(由 Casbin / RLS 擋下):`permission_denied`。
  - id 不存在或已軟刪除:`not_found`。
  - 必填缺失、字典引用非法、`default_sales_rep_id` 指向非法業務、嘗試修改 `customer_code`、排序欄位不在白名單:`invalid_argument`。
  - `(company_id, customer_code)` 唯一衝突(正常流程不應發生,屬防禦):`already_exists`。
- **驗收**:
  - [ ] dept_admin / staff 可管理所屬部門客戶;跨部門 id 操作被拒。
  - [ ] 關鍵字可比對名稱、客戶編號、統編;列表預設排除已刪除,開啟 include_deleted 可見。
  - [ ] 修改 `customer_code` 的請求被拒且原值不變。
  - [ ] 軟刪除與復原皆留稽核紀錄;歷史單據引用軟刪除客戶仍可顯示名稱與編號。

### 子功能 3.1.3: customer_counters 同事務取號(D7)

- **目標**: 建立客戶時以「公司前綴 + 6 位補零自增序號」產生公司內唯一、建立後不可修改的 `customer_code`,取號與建檔同一資料庫交易。`相依: 3.1.1`;前置依賴 `02-tenancy-users.md` 的 `companies.customer_code_prefix`(大寫英數 1–4 碼、全系統唯一)。
- **檔案**: Create `backend/internal/domain/customers/customer_counter.go`(Ent schema);Update `backend/internal/domain/customers/service.go`(CreateCustomer 串接)
- **介面**: Ent 實體 `CustomerCounter`,欄位:`company_id`(主鍵)、`next_seq`(整數,下一個可用序號)、`version`(整數,樂觀鎖版本)、`updated_at`。對外不暴露 RPC,僅供 `CustomerService.CreateCustomer` 內部使用。
- **實作邏輯**:
  1. 交易邊界 = 「取號 + 客戶建檔 + 建帳號(3.1.4)+ 稽核」整體一個資料庫交易,同成功同失敗(D18)。
  2. 交易內讀取 `customer_counters` 中該 `company_id` 的列;不存在則插入初始列(`next_seq = 1`、`version = 0`)後重新讀取。
  3. 取當前 `next_seq` 為本次序號,以 `version` 為條件執行樂觀更新(`next_seq + 1`、`version + 1`,更新條件帶舊 `version`)。
  4. 若更新影響列數為 0(version 已被併發交易推進),整個交易回滾,從步驟 2 重新開始;重試設上限(例如 5 次),超限回錯。
  5. 以公司 `customer_code_prefix` + 序號補零至 6 位組成 `customer_code`(如 `TY000123`);序號超過 6 位時自然延伸(不截斷、不回繞)。
  6. 以該 `customer_code` 建立客戶列;唯一索引 `(company_id, customer_code)` 為最終防線,衝突即回滾重試(理論上不應發生,作為防禦)。
  7. 公司前綴日後修改不影響既有客戶編號,計數器不重置(D7);前綴不存在或公司被停用時拒絕建檔。
- **錯誤處理**:
  - 樂觀鎖重試超過上限:`failed_precondition`(提示稍後重試,不回傳半成品)。
  - 公司不存在 / 無前綴 / 已停用:`failed_precondition`。
  - 唯一索引防禦性衝突且重試耗盡:`already_exists`。
- **驗收**:
  - [ ] 前綴 `TY`、計數器 `next_seq = 123` 時建立客戶,得到 `TY000123`,且 `next_seq` 同交易遞增為 124。
  - [ ] 併發兩請求同 version 取號:一方 version 衝突重試,最終兩客戶取得不同且連續編號,無重複、無跳號遺失(見「整合測試重點」)。
  - [ ] 取號成功但後續建帳號失敗時,整交易回滾,計數器不遞增、客戶不存在。

### 子功能 3.1.4: 建檔連動主帳號 + 業務子帳號(D22)

- **目標**: 建立客戶時同一交易自動建立 1 個主帳號 + 1 個業務子帳號,各發一組 24 小時效期臨時密碼,業務子帳號憑證交付 `default_sales_rep_id` 所指業務。`相依: 3.1.3`;密碼雜湊、臨時密碼效期與首登強制修改機制見 `01-auth.md`。
- **檔案**: Update `backend/internal/domain/customers/service.go`(CreateCustomer 複合流程)
- **介面**: 不新增 RPC;`CreateCustomerResponse` 訊息形狀擴充為:`customer`(客戶主檔)、`primary_account`(`account_name` + `temp_password`,交付店家、由業務轉交)、`sales_rep_account`(`account_name` + `temp_password`,交付所屬業務)、`account_manage_url`(帳號管理深層連結 `https://<domain>/customer_account_manage`,規格 v1.0.34 §9.4)。臨時密碼僅於本次回應出現,系統不留明文。
- **實作邏輯**:
  1. 在 3.1.3 的同一資料庫交易內,於 `users` 建立主帳號:關聯 `customer_id`、角色 `customer`、`is_primary = true`、`account_name` 預設為客戶名稱。
  2. 同交易建立業務子帳號:`is_primary = false`、`account_name` 預設為「客戶名稱(業務)」、標記為系統自動附帶(`system_generated = true` 或等效欄位,供灰化判斷)。
  3. 兩帳號各自產生密碼學安全的隨機臨時密碼,雜湊後儲存,並設 `temp_password_expires_at = 建立時間 + 24 小時`、`must_change_password = true`(規格 §4.2);首次登入強制改密、過期拒絕登入的行為由 `01-auth.md` 的登入流程落實。
  4. 帳號名稱於客戶內唯一;若客戶名稱或「客戶名稱(業務)」與既有帳號衝突(理論上新建客戶不會,防禦處理),整交易回滾。
  5. 任一步驟失敗(取號、建客戶、建任一帳號、寫稽核)整交易回滾,不產生孤兒客戶或孤兒帳號。
  6. 交付規則:主帳號帳密供業務轉交店家;業務子帳號帳密供 `default_sales_rep_id` 所指業務留存使用(業務以客戶身分代客操作)。主責業務日後變更時,由後台重置業務子帳號密碼並轉交新業務(移交操作屬 `02-tenancy-users.md` 的使用者管理範圍,不在本 RPC)。
  7. 灰化規則(資料面契約):店家以主帳號呼叫帳號管理清單時,系統自動附帶的業務子帳號以「系統預設(業務使用)」標記回傳且店家對其改名 / 停用 / 重置一律拒絕(後端強制,非僅 UI 灰化);後台(dept_admin 以上)可正常管理。清單 API 與拒絕邏輯實作於 `01-auth.md` / `02-tenancy-users.md`,本子功能負責在建檔時寫入可識別的標記。
  8. 店家日後可自行以主帳號新增其他子帳號(如主廚),不影響本流程建立的兩個帳號;每客戶恆恰有一個 `is_primary = true` 帳號,此約束由 users 層保證(見 `02-tenancy-users.md`)。
- **錯誤處理**:
  - 帳號名稱客戶內重複(防禦):`already_exists`,整交易回滾。
  - `default_sales_rep_id` 未提供或非法(建檔交付流程必要):`invalid_argument`。
  - users 層任一約束違反:整交易回滾,錯誤碼隨底層原因映射(`already_exists` / `invalid_argument`)。
- **驗收**:
  - [ ] 建立客戶後,同交易存在 1 主帳號(`is_primary = true`、名稱 = 客戶名稱)與 1 業務子帳號(名稱 = 客戶名稱(業務)、系統標記)。
  - [ ] 回應一次性回傳兩組帳號名稱 + 臨時密碼與帳號管理深層連結;兩組臨時密碼不同,效期均為 24 小時,首登強制修改。
  - [ ] 建帳號任一環節失敗時客戶主檔與計數器皆回滾。
  - [ ] 店家帳號管理清單中業務子帳號帶「系統預設」標記,店家對其管理操作被後端拒絕。

### 子功能 3.1.5: preferred_delivery_days 與 promo_tag_ids 欄位維護(D26 / D24)

- **目標**: 維護客戶的偏好送貨日(星期一至六核取)與促銷分類標籤訂閱,供下單順延與促銷推播分群使用。`相依: 3.1.2`
- **檔案**: Update `backend/internal/domain/customers/service.go`(Create / Update 驗證與寫入)
- **介面**: 不新增 RPC;`CreateCustomerRequest` / `UpdateCustomerRequest` 納入 `preferred_delivery_days`(布林陣列,長度 6,索引 0 = 星期一)與 `promo_tag_ids`(重複字串 / ID 陣列);`Customer` 訊息回傳同兩欄位。
- **實作邏輯**:
  1. `preferred_delivery_days` 驗證:必須為長度恰好 6 的布林陣列(星期一至六;星期日不送貨,不在核取範圍,規格 D26);缺漏或長度不符拒絕。全 false(未勾選任何日)為合法值,語意為「不啟用順延」。
  2. 順延判斷(非勾選日自動順延至下一勾選日)屬下單流程,實作於 `05-sales-orders.md`,本欄位僅負責儲存與驗證;本文件承諾的契約:讀取端可取得原樣陣列,順延端自行計算。
  3. `promo_tag_ids` 驗證:每個 id 須存在於同部門的 `promo_tags` 且未軟刪除(promo_tags 主檔屬 `07-notifications.md` 的促銷分群範圍);陣列去重後儲存;空陣列合法(不訂閱任何促銷分類)。
  4. 兩欄位異動皆走 3.1.2 的 Update 路徑:白名單欄位、同事務寫稽核(D18)。
  5. 促銷標籤被軟刪除後,客戶已訂閱的殘留 id 不自動清除;推播分群查詢時以「標籤未刪除」為準過濾(消費端責任,見 `07-notifications.md`)。
- **錯誤處理**:
  - 陣列長度不為 6、元素型別錯誤:`invalid_argument`。
  - `promo_tag_ids` 含不存在、跨部門或已刪除標籤:`invalid_argument`(訊息指出非法 id)。
- **驗收**:
  - [ ] 勾選星期二與星期四可正確儲存與讀回(陣列索引 1、3 為 true)。
  - [ ] 長度非 6 的陣列被拒;全 false 可儲存。
  - [ ] 引用跨部門或已刪除促銷標籤被拒;去重後儲存。

---

## 2. Task 3.2:客戶地址簿與聯絡人 API

### 子功能 3.2.1: 地址簿 CRUD + 預設地址

- **目標**: 每位客戶可維護多筆地址,同類型至多一筆預設地址;地址隸屬客戶並繼承其公司與部門歸屬。`相依: 3.1.2`
- **檔案**: Create `backend/internal/domain/customers/addresses.go`(含 Ent schema `CustomerAddress` 與 service 方法)
- **介面**: Connect-RPC `CustomerService` 擴充:`ListAddresses` / `AddAddress` / `UpdateAddress` / `DeleteAddress`(Request 皆帶 `customer_id`;地址訊息含 `id`、`type`(列舉:`shipping` / `billing` / `other`)、收件地址欄位群、`is_default`)。Ent 實體 `CustomerAddress`:`id`、`customer_id`、`company_id`、`department_id`(自建檔客戶複寫,供 RLS)、`type`、`is_default`、`deleted_at` 與稽核欄位。
- **實作邏輯**:
  1. 寫入前先以 `customer_id` 查客戶,確認存在、未軟刪除且在當前租戶範圍;地址列的 `company_id` / `department_id` 由客戶複寫,不信任 Request 攜帶值。
  2. `type` 僅接受列舉三值。
  3. 預設地址唯一性:資料庫建部分唯一索引 `(customer_id, type) WHERE is_default = true AND deleted_at IS NULL`;設定 `is_default = true` 時,於同一交易內先將該客戶同類型其他地址的 `is_default` 清為 false,再設目標列;索引為最終防線,衝突即回滾。
  4. 新增地址時若為該客戶該類型第一筆,自動設 `is_default = true`。
  5. Delete 為軟刪除 + 同事務稽核;被刪除的是預設地址時,不自動遞補預設(下次下單由業務另選),僅於回應/清單呈現無預設狀態。
  6. List 預設排除已刪除;歷史單據快照欄位不受刪除影響。
- **錯誤處理**:
  - 客戶不存在 / 已刪除:`not_found`;跨部門客戶:`permission_denied`。
  - 地址 id 不屬於指定客戶或已刪除:`not_found`。
  - `type` 非法、必填地址欄位缺漏:`invalid_argument`。
  - 預設唯一索引防禦性衝突:`already_exists`(並重試語義交給交易內步驟 3,正常不應透出)。
- **驗收**:
  - [ ] 同一客戶可新增 `shipping` 與 `billing` 地址並各自查得。
  - [ ] 多筆 `shipping` 中設定某筆為預設後,同類型其餘地址自動失去預設;同類型至多一筆預設。
  - [ ] 刪除地址後不再出現於預設查詢,歷史單據顯示不受影響。

### 子功能 3.2.2: 聯絡人 CRUD + 預設聯絡人

- **目標**: 每位客戶可維護多筆聯絡人,至多一筆預設聯絡人;繼承客戶租戶歸屬。`相依: 3.1.2`
- **檔案**: Create `backend/internal/domain/customers/contacts.go`(含 Ent schema `CustomerContact` 與 service 方法)
- **介面**: Connect-RPC `CustomerService` 擴充:`ListContacts` / `AddContact` / `UpdateContact` / `DeleteContact`;聯絡人訊息含 `id`、`name`、`title`、`email`、`phone`、`is_default`。Ent 實體 `CustomerContact` 欄位同訊息,另加 `customer_id`、`company_id`、`department_id`、`deleted_at` 與稽核欄位。
- **實作邏輯**:
  1. 客戶存在性、租戶複寫、軟刪除與稽核規則同 3.2.1 步驟 1、5、6。
  2. 預設聯絡人不分類型:部分唯一索引 `(customer_id) WHERE is_default = true AND deleted_at IS NULL`;設預設採 3.2.1 步驟 3 的同交易「先清後設」。
  3. 首筆聯絡人自動設為預設(同 3.2.1 步驟 4)。
  4. `email` 格式驗證(若提供);`name` 必填;`phone` 不設格式硬限制(保留國際碼彈性),僅限長度上限。
- **錯誤處理**: 同 3.2.1 對應項;email 格式非法:`invalid_argument`。
- **驗收**:
  - [ ] 客戶可有多筆聯絡人;設定預設後至多一筆 `is_default = true`。
  - [ ] 首筆新增自動成為預設;刪除預設聯絡人後清單呈現無預設。
  - [ ] 軟刪除聯絡人不影響歷史單據顯示。

---

## 3. Task 3.3:商品主檔 API

### 子功能 3.3.1: products schema(單位換算、倉別、分切規格關聯)

- **目標**: 定義 `products`、`product_units`、`product_cutting_specs` 三實體,承載商品主檔、多組單位換算率與分切規格關聯。
- **檔案**: Create `backend/internal/domain/products/schema.go`(三個 Ent schema)
- **介面**: Ent 實體:
  - `Product`:`id`、`company_id`、`department_id`、`code`(部門內唯一)、`name`、`category_id`(指向 product_categories,可 NULL)、`inventory_warehouse_id`(產品分庫)、`picking_warehouse_id`(揀貨倉別)、`description`、`is_active`、`deleted_at` 與稽核欄位。
  - `ProductUnit`:`id`、`product_id`、`unit_code`(對應 metadicts `type = unit` 的字典值)、`conversion_rate`(十進位,換算為基本單位的比率)、`is_base`(布林)、`sort_order`。
  - `ProductCuttingSpec`:`id`、`product_id`、`cutting_spec_id`(關聯表,不得以陣列欄位儲存,規格 master-data)。
- **實作邏輯**:
  1. `products.code` 建部分唯一索引 `(department_id, code) WHERE deleted_at IS NULL`(部門內唯一,軟刪除可重建)。
  2. `product_units` 建部分唯一索引 `(product_id) WHERE is_base = true`(每商品至多一個基本單位;服務層保證「恰好一個」,見 3.3.2/3.3.3),另建 `(product_id, unit_code)` 唯一索引(同商品同單位僅一筆)。
  3. `conversion_rate` 採十進位型別(不用浮點),精度足以表達如 0.6 的換算率。
  4. `product_cutting_specs` 建 `(product_id, cutting_spec_id)` 唯一索引。
  5. 倉別與分類引用不設跨表硬外鍵以外的額外約束,同部門合法性於 3.3.2 寫入路徑驗證。
- **錯誤處理**: schema 定義,無執行期錯誤。
- **驗收**:
  - [ ] 三表建立,索引如上;同部門同 `code` 未刪除商品不可並存,軟刪除後可重建同碼。
  - [ ] 一商品無法寫入第二個 `is_base = true` 的單位(索引擋下)。

### 子功能 3.3.2: 商品 CRUD + 軟刪除 + 分類關聯

- **目標**: 提供商品新增、查詢、修改、軟刪除與復原,寫入時連同單位設定與分切規格關聯一併落庫。`相依: 3.3.1`
- **檔案**: Create `backend/internal/domain/products/service.go`、`backend/internal/domain/products/repo.go`
- **介面**: Connect-RPC `ProductService`:`ListProducts`(keyword、category_id 篩選、分頁 meta、include_deleted)/ `GetProduct`(含單位清單與分切規格清單)/ `CreateProduct` / `UpdateProduct` / `DeleteProduct` / `RestoreProduct`。Create / Update 的 Request 訊息形狀含 `units[]`(`unit_code`、`conversion_rate`、`is_base`、`sort_order`)與 `cutting_spec_ids[]`,採整組替換語意。
- **實作邏輯**:
  1. 身分、角色(dept_admin / staff 限本部門)與 RLS 檢查同 3.1.2 步驟 1。
  2. Create / Update 的交易邊界 = 商品主列 + `product_units` 整組替換 + `product_cutting_specs` 整組替換 + 稽核,同一交易(D18)。
  3. 寫入前驗證:`code` 必填且符合部門命名習慣(長度上限);`category_id`、`inventory_warehouse_id`、`picking_warehouse_id` 若提供,須為同部門、未軟刪除的主檔(分類見 3.4.4,倉別見 3.4.1);`cutting_spec_ids` 每個須為同部門未刪除分切規格(見 3.4.3)。
  4. 單位組驗證(細節見 3.3.3):恰一個 `is_base = true`;基本單位 `conversion_rate` 恆為 1;其餘 `conversion_rate > 0`;`unit_code` 皆存在於 metadicts `unit` 字典(系統預設 + 本部門擴充聯集)。
  5. List:預設排除軟刪除;`keyword` 對 `code`、`name` 模糊比對;`category_id` 精確篩選。
  6. Delete:軟刪除 + 稽核;不連動刪除單位與關聯列(保留供歷史查詢),但新單據選品僅列未刪除商品。
  7. Restore:清 `deleted_at` + 稽核;復原前檢查 `(department_id, code)` 部分唯一索引(同期間可能已有同碼新商品,衝突則拒絕復原)。
- **錯誤處理**:
  - 同 3.1.2 的身分 / 範圍 / 不存在對應碼。
  - 同部門同 `code` 未刪除商品已存在:`already_exists`;復原時撞碼同。
  - 倉別 / 分類 / 分切規格跨部門或已刪除:`invalid_argument`。
  - 單位組非法(無基本單位、多個基本單位、換算率非正數、單位字典不存在):`invalid_argument`。
- **驗收**:
  - [ ] 商品可設定多組單位換算率並隨 GetProduct 讀回。
  - [ ] 商品可指定產品分庫與揀貨倉別,且只能選本部門倉別。
  - [ ] 軟刪除後同部門可以同 `code` 重建;軟刪除商品不出現於新單據選品,歷史單據仍顯示。

### 子功能 3.3.3: 單位換算計算邏輯

- **目標**: 提供「任意單位數量 ↔ 基本單位數量」的換算契約,供下單(訂單明細 `base_qty`)、揀貨單與加工單彙總使用。`相依: 3.3.2`;消費端見 `05-sales-orders.md`。
- **檔案**: Create `backend/internal/domain/products/conversion.go`
- **介面**: 內部套件函式(不暴露 RPC):輸入「商品 id + 單位 `unit_code` + 數量」,輸出「基本單位數量(`base_qty`)+ 基本單位 `unit_code`」;另提供反向換算(基本單位 → 指定單位)供單據顯示。
- **實作邏輯**:
  1. 換算規則:基本單位數量 = 輸入數量 × 該單位 `conversion_rate`(如 1 條 × 0.6 = 0.6 kg);反向 = 基本單位數量 ÷ `conversion_rate`。
  2. 計算採十進位算術,不用二進位浮點;精度規則:中間計算保留足夠位數,輸出依商品單位慣例四捨五入至小數 3 位(供彙總一致性)。
  3. 輸入單位不存在於該商品的 `product_units` 時拒絕,不做字典層級的猜測換算。
  4. 輸入數量為 0 合法(換算結果 0);負數拒絕。
  5. 「每商品恰好一個基本單位」由 3.3.2 寫入驗證 + 3.3.1 索引雙重保證;換算入口若發現資料異常(零個或多個基本單位,理論不可能),視為前置條件失敗並記錄,不自行修復。
  6. 下單端契約:訂單明細保存原始 `qty` + `unit`,並以本邏輯算出 `base_qty` 一併保存;事後商品換算率修改不回溯歷史明細(快照語意,消費端責任,見 `05-sales-orders.md`)。
- **錯誤處理**(映射至呼叫端 Connect code):
  - 商品不存在 / 已刪除:`not_found`。
  - 單位未設定於該商品、數量為負:`invalid_argument`。
  - 商品單位組異常(基本單位個數不為 1):`failed_precondition`。
- **驗收**:
  - [ ] 基本單位 `kg` + 換算單位「條」(0.6)的商品,輸入 2 條得 1.2 kg;反向 1.2 kg 得 2 條。
  - [ ] 換算率設為 0 或負數於 3.3.2 寫入時即被拒。
  - [ ] 十進位精度:0.1 kg 級連續換算不產生浮點誤差。

---

## 4. Task 3.4:倉庫 / 車次 / 分切規格 / 商品分類 API(皆部門級)

> 四者皆為部門級實體:帶 `company_id` / `department_id`、軟刪除、跨部門互不可見不可選用;CRUD 形狀一致(List / Create / Update / Delete / Restore + 分頁 meta),權限開放 dept_admin / staff 限本部門。各子功能僅述差異。

### 子功能 3.4.1: warehouses CRUD

- **目標**: 各部門獨立維護倉別主檔,供商品分庫 / 揀貨倉與訂單明細倉別選用。
- **檔案**: Create `backend/internal/domain/warehouses/`(schema、service、repo)
- **介面**: Connect-RPC `WarehouseService`:`ListWarehouses` / `CreateWarehouse` / `UpdateWarehouse` / `DeleteWarehouse` / `RestoreWarehouse`。Ent 實體 `Warehouse`:`id`、`company_id`、`department_id`、`code`、`name`、`address`、`is_active`、`deleted_at` 與稽核欄位。
- **實作邏輯**:
  1. 寫入時 `company_id` / `department_id` 由當前 session 租戶注入,不接受 Request 指定跨部門值。
  2. `code` 部門內部分唯一 `(department_id, code) WHERE deleted_at IS NULL`。
  3. 軟刪除 + 稽核同交易;軟刪除後商品 / 單據不可再選用,既有引用(商品分庫、訂單明細)仍顯示名稱。
  4. List 僅回本部門資料,排除已刪除(可開 include_deleted)。
- **錯誤處理**: 跨部門存取 `permission_denied`;不存在 `not_found`;同碼 `already_exists`;必填缺漏 `invalid_argument`。
- **驗收**:
  - [ ] 甲、乙部門各自建倉,互不可見、商品與單據不可跨部門選用。
  - [ ] 軟刪除倉別後新單據選項消失,既有引用正常顯示。

### 子功能 3.4.2: routes CRUD

- **目標**: 各部門獨立維護車次主檔,供訂單 / 派車選用。
- **檔案**: Create `backend/internal/domain/routes/`
- **介面**: Connect-RPC `RouteService`(五法同 3.4.1 形狀)。Ent 實體 `Route`:`id`、`company_id`、`department_id`、`code`、`name`、`description`、`sort_order`、`is_active`、`deleted_at` 與稽核欄位。
- **實作邏輯**:
  1. 同 3.4.1 步驟 1–4;`sort_order` 供派車看板與下拉排序。
  2. 軟刪除車次後,派車(見 `08-dispatch.md`)與新訂單不可再選用;既有訂單關聯車次名稱正常顯示。
- **錯誤處理**: 同 3.4.1。
- **驗收**:
  - [ ] 部門各自維護車次,跨部門不可見不可選。
  - [ ] 軟刪除車次不再出現於新單據選項,歷史訂單車次名稱仍顯示。

### 子功能 3.4.3: cutting_specs CRUD

- **目標**: 各部門獨立維護分切規格,記錄其加工 / 配送揀貨歸屬,供商品關聯(3.3.1)與訂單明細選用。
- **檔案**: Create `backend/internal/domain/cuttingspecs/`
- **介面**: Connect-RPC `CuttingSpecService`(五法同上)。Ent 實體 `CuttingSpec`:`id`、`company_id`、`department_id`、`code`、`name`、`applies_to`(列舉:加工單歸屬 / 配送揀貨歸屬,決定該規格出現在加工單或揀貨單)、`sort_order`、`is_active`、`deleted_at` 與稽核欄位。
- **實作邏輯**:
  1. 同 3.4.1 步驟 1–4;`code` 部門內部分唯一。
  2. `applies_to` 必填且僅接受列舉值;此歸屬供列印單據(見 `09-printing.md`)決定規格顯示於加工單或揀貨單。
  3. 被商品關聯(`product_cutting_specs`)中的分切規格軟刪除時不強制清除關聯列;商品編輯時關聯清單過濾已刪除規格,新關聯不得引用已刪除規格。
- **錯誤處理**: 同 3.4.1;`applies_to` 非法值 `invalid_argument`。
- **驗收**:
  - [ ] 分切規格帶加工 / 揀貨歸屬並可供商品多選關聯。
  - [ ] 軟刪除規格不可再被新關聯引用,既有商品查詢之規格清單自動排除。

### 子功能 3.4.4: product_categories CRUD

- **目標**: 各部門獨立維護商品分類,供商品主檔歸類與列表篩選。
- **檔案**: Create `backend/internal/domain/productcategories/`
- **介面**: Connect-RPC `ProductCategoryService`(五法同上)。Ent 實體 `ProductCategory`:`id`、`company_id`、`department_id`、`code`、`name`、`sort_order`、`is_active`、`deleted_at` 與稽核欄位。
- **實作邏輯**:
  1. 同 3.4.1 步驟 1–4;`code` 部門內部分唯一。
  2. 軟刪除分類不連動清空既有商品的 `category_id`(保留歷史顯示);商品寫入路徑(3.3.2)拒絕引用已刪除分類。
  3. 1.0 分類為單層,不實作樹狀父子結構。
- **錯誤處理**: 同 3.4.1。
- **驗收**:
  - [ ] 商品可指定本部門分類並以分類篩選列表。
  - [ ] 軟刪除分類後商品不可新設該分類,既有商品分類名稱仍顯示。

---

## 5. Task 3.5:客戶專屬商品清單 API

### 子功能 3.5.1: customer_products schema(alias_name / default_qty / cut_note / promo_tag_ids,不存單價)

- **目標**: 定義客戶專屬商品清單實體,一客戶一商品一筆,1.0 不儲存單價欄位。
- **檔案**: Create `backend/internal/domain/customerproducts/schema.go`
- **介面**: Ent 實體 `CustomerProduct`:`id`、`customer_id`、`product_id`、`company_id`、`department_id`(自客戶複寫,供 RLS)、`alias_name`(客戶慣用名稱)、`default_qty`(預設數量,十進位)、`cut_note`(預設特殊切法備註)、`promo_tag_ids`(JSON 陣列,促銷分類標記,D24)、`deleted_at` 與稽核欄位。**明確不含** `custom_price` 或任何單價 / 金額欄位(1.0 單據不涉價格)。
- **實作邏輯**:
  1. 部分唯一索引 `(customer_id, product_id) WHERE deleted_at IS NULL`:同一客戶對同一商品僅一筆(一個別名);軟刪除後可重建不衝突。
  2. `default_qty` 預設 0;0 為合法值,語意「保留於清單但後續單據不顯示」(見 3.5.2)。
  3. `alias_name` 預設為商品名稱,可由業務改為客戶慣用語;長度設上限。
  4. `promo_tag_ids` 驗證規則同 3.1.5 步驟 3(同部門、未刪除、去重)。
- **錯誤處理**: schema 定義,無執行期錯誤。
- **驗收**:
  - [ ] 同客戶同商品寫入第二筆未刪除記錄被索引擋下;軟刪除後可重建。
  - [ ] 實體不包含任何單價欄位。

### 子功能 3.5.2: customer_products CRUD + 軟刪除 + qty=0 保留不顯示

- **目標**: 提供客戶專屬清單的新增、查詢、修改、軟刪除;`default_qty = 0` 的項目保留於清單但下游單據不顯示。`相依: 3.5.1`
- **檔案**: Create `backend/internal/domain/customerproducts/service.go`、`backend/internal/domain/customerproducts/repo.go`
- **介面**: Connect-RPC `CustomerProductService`:`ListCustomerProducts`(Request 帶 `customer_id`;Response 為該客戶清單,含商品基本資料聯播)/ `AddCustomerProduct` / `UpdateCustomerProduct` / `DeleteCustomerProduct`。List 提供 `for_order` 旗標:`true` 時(下單 / 單據用途)排除 `default_qty = 0` 與已軟刪除項目;`false` 時(客戶檔案管理用途)僅排除已軟刪除,保留 qty=0 項目。
- **實作邏輯**:
  1. 身分與範圍:dept_admin / staff 限本部門;`customer` 身分(App 客戶端)可查自己清單但僅 `for_order` 語意(data_scope self,主帳號依規格 §4.2 對業務 API 一律 403,僅子帳號可用)。
  2. Add:驗證客戶與商品皆存在、未刪除、同租戶;`customer_id + product_id` 已存在未刪除記錄時拒絕(一客戶一商品一筆);`alias_name` 未提供時帶入商品名稱。
  3. Update:可改 `alias_name`、`default_qty`(含改為 0)、`cut_note`、`promo_tag_ids`;`customer_id` / `product_id` 不可改(改商品 = 刪舊建新)。
  4. Delete:軟刪除 + 同事務稽核;歷史訂單明細快照不受影響。
  5. List 排序:`sort_order` 無設計,以更新時間或別名排序,供 App 快速下單穩定呈現。
- **錯誤處理**:
  - 客戶 / 商品不存在或跨租戶:`not_found` / `permission_denied`。
  - 同客戶同商品重複新增:`already_exists`。
  - `default_qty` 為負、別名超長、促銷標籤非法:`invalid_argument`。
  - 主帳號呼叫:`permission_denied`。
- **驗收**:
  - [ ] 一客戶一商品僅一筆;數量設 0 後管理視圖仍可見、下單帶出(for_order)不顯示。
  - [ ] 軟刪除後重建同商品不觸發唯一衝突,歷史訂單明細顯示不受影響。

### 子功能 3.5.3: 下單手打自動建別名

- **目標**: 業務下單手打商品名稱並確認儲存時,自動為該客戶建立專屬清單記錄;重複情境冪等,不產生重複別名。`相依: 3.5.2`;呼叫端為下單流程(見 `05-sales-orders.md`)。
- **檔案**: Update `backend/internal/domain/customerproducts/service.go`
- **介面**: Connect-RPC `CustomerProductService` 增加 `EnsureCustomerProduct(EnsureCustomerProductRequest) returns (EnsureCustomerProductResponse)`;Request 含 `customer_id`、`product_id`(手打名稱先於商品總表模糊比對,命中則帶 product_id;未命中的純手打品項不建清單,僅存於訂單明細文字)、`alias_name`(手打名稱);Response 回傳既有或新建的記錄及 `created` 旗標。
- **實作邏輯**:
  1. 僅在下單流程中、使用者對「是否儲存為該客戶別名」確認同意後呼叫;未確認則不呼叫、不清單化。
  2. 查 `(customer_id, product_id)` 未刪除記錄:存在則不更動任何欄位,直接回傳既有記錄(`created = false`)——別名以既有為準,手打文字僅保留於當次訂單明細。
  3. 不存在則建立:`alias_name` = 手打名稱、`default_qty` = 0、`cut_note` 空;與訂單建立**不**同交易(清單建立失敗不應回滾訂單),但以冪等語意重試。
  4. 併發或重複呼叫導致唯一索引衝突時:捕捉 `already_exists`,改為重新讀取既有記錄回傳(冪等收斂),不向呼叫端報錯。
  5. 存在已軟刪除的同客戶同商品記錄時:新建一筆(不復活舊記錄,避免舊別名 / 舊預設值殘留影響),部分唯一索引允許。
  6. 商品已軟刪除時拒絕建立(下單端本就不應選到已刪除商品,此為防禦)。
- **錯誤處理**:
  - 客戶 / 商品不存在、跨租戶、商品已刪除:`not_found` / `permission_denied` / `failed_precondition`(商品已刪除)。
  - `product_id` 缺失(純手打未命中商品):不視為錯誤,呼叫端應不呼叫本 RPC;若仍呼叫以 `invalid_argument` 拒絕。
  - 唯一衝突:內部吸收,回傳既有記錄(見步驟 4),不透出錯誤。
- **驗收**:
  - [ ] 下單手打並確認儲存後,該客戶專屬清單出現以手打名稱為別名的新記錄。
  - [ ] 同客戶同商品重複觸發(含併發)僅有一筆記錄,第二次呼叫回傳既有記錄且 `created = false`。
  - [ ] 未命中商品總表的純手打品項不產生清單記錄。

---

## 6. Task 3.6:檔案資產 API

### 子功能 3.6.1: 上傳驗證(MIME 白名單、副檔名 + magic bytes 雙重檢查)

- **目標**: 所有上傳先經白名單與一致性驗證:圖片 jpeg / png / webp ≤ 5 MB,PDF ≤ 10 MB;宣告 MIME、副檔名、magic bytes 三者一致才接受。
- **檔案**: Create `backend/internal/domain/fileassets/validate.go`
- **介面**: REST `POST /api/v1/files/upload`(multipart/form-data;欄位:`file`、`owner_type`、`owner_id`);驗證邏輯為內部函式,輸入檔案串流與宣告 MIME,輸出正規化後的 MIME 與大小,或拒絕原因。
- **實作邏輯**:
  1. 身分:需已登入(員工或客戶身分皆可,依用途端點規範);未認證拒絕。
  2. 白名單對照表(宣告 MIME ↔ 允許副檔名 ↔ magic bytes 簽名 ↔ 大小上限):
     - `image/jpeg`:副檔名 `.jpg` / `.jpeg`;magic bytes `FF D8 FF`;上限 5 MB。
     - `image/png`:副檔名 `.png`;magic bytes `89 50 4E 47 0D 0A 1A 0A`;上限 5 MB。
     - `image/webp`:副檔名 `.webp`;RIFF 標頭 `52 49 46 46` + 第 8–11 位元組 `57 45 42 50`(`WEBP`);上限 5 MB。
     - `application/pdf`:副檔名 `.pdf`;magic bytes `25 50 44 46`(`%PDF`);上限 10 MB。
  3. 檢查順序:先比大小上限(讀取時以限制讀取器截斷,超限即拒,不完整落盤)→ 宣告 MIME 是否在白名單 → 副檔名是否與宣告 MIME 匹配 → 讀取檔頭比對 magic bytes;任一不符即拒絕,不寫入儲存、不建記錄。
  4. 副檔名比對不區分大小寫;無副檔名拒絕。
  5. 1.0 不實作病毒掃描(規格明定);白名單以外類型(如 gif、svg、zip)一律拒絕。
- **錯誤處理**:
  - 未認證:`unauthenticated`。
  - MIME 非白名單、副檔名不匹配、magic bytes 不符、缺副檔名:`invalid_argument`(訊息指出失敗環節,不回傳檔案內容相關細節)。
  - 超過大小上限:`invalid_argument`(訊息含上限值)。
- **驗收**:
  - [ ] 白名單內 jpeg / png / webp ≤ 5 MB、pdf ≤ 10 MB 可通過驗證。
  - [ ] 改副檔名的偽裝檔(如 exe 改名 .jpg)被 magic bytes 檢查拒絕;副檔名與宣告 MIME 不符被拒;超限被拒。

### 子功能 3.6.2: 本地儲存 + file_assets 元資料

- **目標**: 驗證通過的檔案寫入本地掛載(volume / NFS),並建立完整 `file_assets` 元資料記錄。`相依: 3.6.1`
- **檔案**: Create `backend/internal/domain/fileassets/storage.go`、`backend/internal/domain/fileassets/schema.go`、`backend/internal/domain/fileassets/service.go`
- **介面**: 同上 REST 端點;成功回應 JSON:`id`、`url`(下載相對路徑,見 3.6.3)、`filename`、`original_filename`、`mime_type`、`size_bytes`。Ent 實體 `FileAsset`:`id`、`company_id`、`department_id`、`owner_type`、`owner_id`、`filename`(儲存檔名)、`original_filename`、`mime_type`、`size_bytes`、`storage_path`、`url`、`created_by`、`created_at`、`deleted_at`。
- **實作邏輯**:
  1. 儲存檔名由系統產生(隨機 uuid + 正規化副檔名),不採用使用者原始檔名,避免路徑穿越與撞名;`original_filename` 原樣保留(含中文)。
  2. `storage_path` 依 `<儲存根目錄>/<company_id>/<yyyy>/<mm>/<filename>` 組織,便於備份與排查;寫入前確認目錄存在(不存在則建立)。
  3. `owner_type` / `owner_id` 關聯驗證:對應資料須存在且屬上傳者租戶範圍(如公司 Logo、公告圖片);一筆 `file_assets` 至多對應一個 owner;`company_id` / `department_id` 自 session 或 owner 資料推得,不信任 Request。
  4. 寫入順序:先將檔案完整寫入本地儲存並 fsync 確認 → 成功後於資料庫交易建立 `file_assets` 記錄(+ 稽核);檔案寫入失敗則不建記錄(不留孤兒記錄);資料庫寫入失敗則刪除已落盤檔案(不留孤兒檔案)。
  5. 備份至 GCS 屬 ops 排程(規格 file-assets),不在本流程。
- **錯誤處理**:
  - owner 不存在或跨租戶:`invalid_argument`(關聯非法)/ `permission_denied`。
  - 儲存寫入失敗(掛載不可用、磁碟錯誤):`internal` 類錯誤以 `failed_precondition` 之外的伺服器錯誤回傳(REST 對應 500),不建記錄。
  - 未認證:`unauthenticated`。
- **驗收**:
  - [ ] 上傳成功後檔案落於本地掛載,`file_assets` 記錄含全部必備欄位,`original_filename` 保留中文原檔名、`filename` 為系統產生。
  - [ ] 儲存失敗不留 `file_assets` 記錄;資料庫失敗不留孤兒檔案。

### 子功能 3.6.3: 下載端點

- **目標**: 以 `GET /api/v1/files/:id/download` 提供檔案下載,不暴露本地路徑,並強制租戶隔離。`相依: 3.6.2`
- **檔案**: Create `backend/internal/domain/fileassets/download.go`
- **介面**: REST `GET /api/v1/files/:id/download`;成功回應:檔案內容串流,標頭帶正確 `Content-Type`(依 `mime_type`)與 `Content-Disposition`(`filename*` 編碼 `original_filename`,支援中文檔名)。
- **實作邏輯**:
  1. 需已認證;以 `:id` 查 `file_assets`,不存在或已軟刪除一律回找不到(不區分原因,避免存在性探測)。
  2. 租戶判定:記錄的 `company_id` 須與當前使用者相符,且依其 `data_scope` 判定部門範圍;跨公司一律拒絕;RLS 注入為最後防線。
  3. 自 `storage_path` 開檔串流回傳;檔案遺失(記錄在、檔案不在)記錄告警並回找不到。
  4. 回應不含 `storage_path`;不提供目錄瀏覽;`id` 為 uuid 不具可列舉性。
  5. 軟刪除(`DELETE /api/v1/files/:id`,管理用途)僅標記 `deleted_at`,實體檔案保留供稽核;軟刪除後下載一律拒絕。
- **錯誤處理**:
  - 未認證:`unauthenticated`(401)。
  - 跨公司或超資料範圍:`permission_denied`(403;或統一以 404 避免洩漏存在性,實作選 404,與步驟 1 一致)。
  - 不存在 / 已刪除 / 檔案遺失:`not_found`(404)。
- **驗收**:
  - [ ] 同公司使用者可下載並取得正確 MIME 與原始檔名。
  - [ ] 跨公司以 id 請求被拒;未認證被拒;已軟刪除 id 下載回找不到。

---

## 7. Task 3.8:QR Code 產生與兌換(僅後端)

> **範圍註明**:原 Task 3.8 含前端元件 `frontend/src/components/customer-qrcode.tsx`(QR 圖呈現與下載分享按鈕),該元件屬前端計畫範圍,本文件不涵蓋;本節僅拆後端:簽章 token 機制(3.8.1)與產生 / 兌換端點(3.8.2)。App 掃碼後的登入完成流程(session 核發)屬 `01-auth.md`,此處僅到「回傳可選子帳號清單」為止。

### 子功能 3.8.1: QR 簽章 token 產生與驗證(company_id + customer_code + exp)

- **目標**: 產生與驗證經簽章的一次性 QR token,payload 編碼 `company_id` + `customer_code` + 過期時間,杜絕僞造與跨公司誤定位。`相依: 3.1.2`
- **檔案**: Create `backend/internal/domain/customers/qrcode.go`
- **介面**: 內部函式:`GenerateQRToken(company_id, customer_code) → token 字串`;`VerifyQRToken(token) → (company_id, customer_code)`,失敗回具體原因(格式非法 / 簽章不符 / 已過期 / 已使用)。
- **實作邏輯**:
  1. payload 結構:`company_id`、`customer_code`、`exp`(過期時間)、`jti`(唯一識別,一次性控制用)、`purpose`(固定值,防與其他簽章 token 混用)。
  2. 簽章採伺服器端對稱密鑰(HMAC-SHA256)或等效機制;密鑰由環境設定注入,不進版本庫;payload 未經加密,不得放入任何憑證資訊(僅定位用)。
  3. 效期:預設 30 天(供印製 / 轉發場景),由設定檔調整;`exp` 過期即拒。
  4. 一次性:兌換成功後將 `jti` 記入 Valkey(TTL ≥ token 剩餘效期),已存在的 `jti` 拒絕;Valkey 不可用時降級為僅驗簽章與效期並記錄告警(安全與可用性取捨,見稽核規格 rate limit 章節精神)。
  5. 驗證順序:格式解析 → 簽章比對 → `purpose` 檢查 → `exp` → `jti` 一次性 → 以 `company_id` + `customer_code` 定位客戶(不可僅以 `customer_code` 跨公司識別,規格 identity-access);客戶不存在或已軟刪除拒絕。
  6. Rate limit:兌換端點套用來源限流(見 3.8.2 步驟 4 與稽核規格),防暴力列舉。
- **錯誤處理**(供 3.8.2 映射):
  - 格式非法 / 簽章不符 / purpose 不符:對外統一 `unauthenticated`(不區分,防探測)。
  - 已過期 / 已使用 / 客戶不存在或已刪除:`failed_precondition`(對外訊息不區分細節)。
- **驗收**:
  - [ ] token 含 `company_id` + `customer_code` + `exp`,竄改任一欄位即簽章驗證失敗。
  - [ ] 過期 token、已兌換 token 皆被拒;不同公司同 `customer_code` 依 `company_id` 正確定位。

### 子功能 3.8.2: QR Code 產生與兌換端點(REST 公開端點)

- **目標**: 提供中台產生客戶登入 QR Code(含深層連結)與 App 掃碼後的公開兌換端點。`相依: 3.8.1`
- **檔案**: Create `backend/internal/domain/customers/qrcode.go`(同上檔,產生端);Create `backend/internal/httpapi/qr.go`(REST 兌換端點,路徑依公開 REST 路由慣例)
- **介面**:
  - 產生(需認證):Connect-RPC `CustomerService.GetCustomerQRCode(GetCustomerQRCodeRequest) returns (GetCustomerQRCodeResponse)`;Request 帶 `customer_id`;Response 含 `qr_url`(深層連結 `https://<domain>/customer_account_qrcode/{token}`)與 `qr_image`(PNG 位元組,供前端下載 / 分享;前端元件不屬本文件)。
  - 兌換(公開):REST `POST /api/v1/auth/qr/redeem`,body 含 `token`;成功回應 JSON:`company`(公司識別資訊)、`customer_code`、客戶顯示名稱、`accounts[]`(可選的**店家子帳號**清單:僅 `account_name` 與 id,不含主帳號、不含業務子帳號,規格 §4.2)。後續以所選子帳號 + 密碼完成登入,屬 `01-auth.md`。
- **實作邏輯**:
  1. 產生端:dept_admin / staff 限本部門;驗證客戶存在且未刪除 → 以 3.8.1 產生 token → 組深層連結(Universal Link / App Link 形式,已安裝 App 直接開啟,未安裝導向商店)→ 產生 QR PNG → 寫稽核(產生者、客戶、時間,D18 同事務範圍為本次操作自身)。
  2. 每次呼叫產生新 token;舊 token 依其自身 `exp` 與一次性狀態獨立有效或失效,不互相作廢。
  3. 兌換端:**公開端點、不需登入**;解析並驗證 token(3.8.1 步驟 5 全鏈);通過後立即將 `jti` 標記已使用(先標記再回應,避免同 token 併發雙兌換)。
  4. 兌換端套用 rate limit(同來源短時間超次拒絕,規格 audit-compliance);失敗與成功的回應皆不洩漏 token 各欄位細節。
  5. 帳號清單過濾:僅回傳該客戶 `is_primary = false` 且非系統業務子帳號(3.1.4 步驟 7 標記)的啟用中帳號;主帳號(管理用途)與業務子帳號(專供所屬業務)皆排除。
  6. 兌換成功寫稽核(兌換時間、客戶、來源);兌換本身**不**核發 session,登入於子帳號選定後另行完成。
- **錯誤處理**:
  - 產生端:未認證 `unauthenticated`;跨部門 `permission_denied`;客戶不存在 `not_found`。
  - 兌換端:token 無效 `unauthenticated`;過期 / 已使用 / 客戶已刪除 `failed_precondition`;超 rate limit 以拒絕回應(HTTP 429 對應,REST 端點不依 Connect code)。
  - 兌換後無任何可選子帳號(店家尚未自建):`failed_precondition`(提示以主帳號登入建立子帳號)。
- **驗收**:
  - [ ] 中台可為本部門客戶產生 QR,回傳深層連結與 PNG;連結內 token 可被兌換端驗證。
  - [ ] 掃碼兌換回傳公司、客戶與店家子帳號清單,清單不含主帳號與業務子帳號。
  - [ ] 同一 token 第二次兌換被拒;App 可依 QR token 完成掃描登入的前段(選帳號)流程。

---

## 8. 整合測試重點

依 `00-index.md` §3.5,以下行為必須以整合測試(真實 PostgreSQL / Valkey)覆蓋,不得以單元測試替代:

1. **取號併發(3.1.3,必要)**:對同一公司併發 N 個 CreateCustomer,驗證 (a) 全部成功或明確失敗、(b) 成功的 `customer_code` 互不重複且序號連續無跳號遺失、(c) `customer_counters.next_seq` 最終值 = 起始值 + 成功筆數、(d) version 衝突確實發生並被重試吸收。另測「建帳號失敗 → 計數器回滾不遞增」。
2. **一主多子建檔原子性(3.1.4)**:注入帳號建立失敗,驗證客戶主檔、計數器、兩帳號、稽核全部回滾;成功路徑驗證恰一個 `is_primary`、兩組臨時密碼不同且 24h 效期。
3. **軟刪除部分唯一索引**:客戶 `customer_code`、商品 `code`、customer_products `(customer_id, product_id)` 三組,各自驗證「未刪除重複被拒 → 軟刪除 → 同鍵重建成功」。
4. **預設唯一性(3.2.1 / 3.2.2)**:併發設定同客戶同類型兩筆地址為預設,最終恰一筆 `is_default = true`。
5. **單位換算(3.3.3)**:多組單位換算與 `base_qty` 快照語意;換算率修改後歷史明細不變。
6. **RLS / 跨部門隔離(3.1–3.5)**:以乙部門身分對甲部門資源逐一嘗試讀寫(客戶、商品、倉別、車次、分切規格、分類、專屬清單),全部 `permission_denied` / `not_found`。
7. **檔案上傳三重檢查與隔離(3.6)**:偽裝檔、超限檔、跨公司下載、軟刪除後下載各案例;儲存失敗 / DB 失敗的孤兒清理。
8. **QR token(3.8)**:簽章竄改、過期、一次性(含併發雙兌換)、跨公司同 `customer_code` 定位、rate limit 觸發、清單排除主帳號與業務子帳號。
9. **qty=0 過濾(3.5.2)**:管理視圖與 for_order 視圖的差異斷言。

---

*最後更新:2026-08-17*
