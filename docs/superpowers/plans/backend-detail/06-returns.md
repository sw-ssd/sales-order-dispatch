# 06 — 退貨申請與審核(Task 4.7,僅後端)

> - 對應原計畫:Task 4.7「退貨申請與審核」(規格 §6.6、決策 D25 / D23 / D18 / D21)。
> - 對應規格書 v1.0.34 章節:§6.6 退貨流程(客戶發起)、§10.1 通知路由、§5.2 資料表(`return_requests` / `return_request_items`)。
> - 相依文件:`04-master-data.md`(Task 3.6 file_assets 上傳、Task 3.5 客戶專屬商品清單)、`05-sales-orders.md`(Task 4.1/4.2 歷史訂單)、`03-metadicts-audit.md`(Task 2.6 稽核日誌)、`07-notifications.md`(Task 4.3/4.4 通知與 FCM)。
> - 範圍註記:原 Task 4.7 同時涵蓋 App 端入口(`sales-order-app` 訂單歷史 / 專屬商品頁的退貨入口、退貨證明畫面),屬 Phase 6 App 範圍,本文件**僅拆後端部分**;App 入口與證明畫面的 UI 實作不在此文件。
> - 公司端**不建置配送專屬頁面**(D25):系統僅記錄申請與審核狀態,取貨由業務自行通知配送,後端無配送取貨 API。

---

## 共通規則

以下規則適用本文件所有子功能,各欄不再重複(完整定義見 `00-index.md` §3):

1. **交易與稽核**:退貨申請建檔、審核狀態異動與對應 audit log 寫入皆為同一 DB 交易,同成功同失敗(D18);各子功能「實作邏輯」欄標出交易邊界。
2. **軟刪除**:`return_requests` / `return_request_items` 採 `deleted_at` 軟刪除,查詢預設排除;唯一性需求以部分唯一索引處理(D10)。
3. **多租戶**:兩表皆帶 `company_id` / `department_id`;Casbin 管功能權限,RLS(`app.current_company_id` / `app.current_department_id` / `app.current_data_scope`)管資料範圍,為最後防線。客戶帳號 data_scope 為 self,僅見自己客戶的退貨申請;主帳號呼叫本業務 API 一律 `permission_denied`(主帳號僅供帳號管理)。
4. **錯誤處理**:統一以 Connect code 表述;樂觀鎖 / 狀態衝突一律 `failed_precondition`。
5. **不修改原訂單狀態**:退貨流程全程**不得**異動 `sales_orders` 任何欄位與狀態,僅在品項上保留對來源訂單的參照(D25、規格 §6.6)。

---

## 子功能 4.7.1: return_requests / return_request_items schema

- **目標**: 建立退貨申請主檔與品項檔的 Ent schema 與 migration,支撐 pending → approved / rejected 的審核流程。
- **檔案**:
  - Create `backend/internal/domain/returns/schema/return_request.go`(Ent schema)
  - Create `backend/internal/domain/returns/schema/return_request_item.go`(Ent schema)
  - Create 對應 migration(依專案 migration 目錄慣例)
- **介面**: 無對外 RPC;產出兩張表與 Ent 實體供 4.7.2–4.7.5 使用。
  - `return_requests` 欄位(公司/部門級,規格 §5.2):
    - `id`(UUID 主鍵)、`company_id`、`department_id`、`customer_id`(必填,退貨所屬客戶)
    - `created_by_user_id`(必填,發起帳號;即推播目標帳號,D23)
    - `status`(enum:`pending` / `approved` / `rejected`,預設 `pending`)
    - `remark`(整張申請的備註,可空)
    - `reviewed_by_user_id`(可空,審核業務)、`reviewed_at`(可空)、`reject_reason`(`rejected` 時必填)
    - `version`(整數,樂觀鎖,預設 0)
    - `created_at` / `updated_at` / `deleted_at`
  - `return_request_items` 欄位(關聯 `return_requests`,規格 §5.2):
    - `id`、`return_request_id`(外鍵)、`company_id`、`department_id`(冗餘帶入,供 RLS)
    - `source_type`(enum:`order_item` / `customer_product`;標示品項來源,1.0 兩者並存,D25)
    - `sales_order_id`、`sales_order_item_id`(`source_type = order_item` 時必填,否則為空;僅參照用途)
    - `customer_product_id`、`product_id`(`source_type = customer_product` 時 `customer_product_id` 必填;品項一律快照 `product_id` 與商品名稱 / 規格 / 單位,避免來源日後異動影響證明內容)
    - `quantity`(正數)、`unit`、`reason`(品項退貨原因,必填)
    - `photo_file_ids`(JSONB 陣列,元素為 `file_assets.id`,可多張、可空)
    - `created_at` / `updated_at` / `deleted_at`
- **實作邏輯**:
  1. 依上列欄位定義 Ent schema,`status` / `source_type` 以 enum 型別限定合法值。
  2. `return_request_items.return_request_id` 設外鍵;主檔軟刪除時品項一併軟刪除(於 usecase 層同事務處理,DB 層不做 cascade delete)。
  3. 索引:`return_requests` 建 `(company_id, customer_id, status)`、`(company_id, department_id, status)` 複合索引(客戶自查、業務待審清單);品項表建 `return_request_id` 索引。
  4. 軟刪除:兩表掛 `deleted_at` mixin;`return_requests` 無業務唯一約束需求(同一客戶可有多張 pending,1.0 不限制併發申請),故不需部分唯一索引。
  5. RLS policy:兩表啟用 RLS,以 `company_id` / `department_id` 比對 session 注入值;客戶 self 範圍再比對 `customer_id`。
- **錯誤處理**: schema 層無 RPC 錯誤;migration 失敗即部署失敗,不進入執行期。
- **驗收**:
  - [ ] migration 可於乾淨資料庫從頭執行成功,兩表、enum、索引、RLS policy 齊全。
  - [ ] Ent code gen 後兩實體可供 CRUD 使用,`status` 僅接受三個合法值。
  - [ ] 以不同 `company_id` session 查詢,RLS 正確隔離(D21 RLS 整合測試範圍)。

## 子功能 4.7.2: 發起退貨(客戶端 Create / List)

- **目標**: 客戶子帳號從「歷史訂單勾選」或「專屬商品清單選擇」兩種來源擇品發起退貨申請,附數量 / 原因 / 備註 / 照片;並提供發起帳號查詢自己客戶的申請列表與明細。`相依: 4.7.1`;照片依賴 Task 3.6 上傳 API;來源資料依賴 Task 4.1/4.2(歷史訂單)與 Task 3.5(專屬商品)。
- **檔案**:
  - Create `backend/internal/domain/returns/create.go`(usecase)
  - Create `backend/internal/domain/returns/list.go`(usecase)
  - Create `backend/internal/domain/returns/handler.go`(Connect handler)
  - Create `backend/internal/domain/returns/repo.go`(Ent 查詢)
  - Update proto 定義檔(依專案 proto 目錄慣例)新增 `ReturnService`
- **介面**: Connect-RPC `ReturnService`
  - `Create(CreateReturnRequestRequest) → CreateReturnRequestResponse`:請求含 `items[]`(每項含 `source_type`、`sales_order_item_id` 或 `customer_product_id`、`quantity`、`reason`、`photo_file_ids[]`)、`remark`;回應含 `id`、`status`(必為 `pending`)。
  - `List(ListReturnRequestsRequest) → ListReturnRequestsResponse`:分頁 + `status` 篩選;客戶帳號固定只回自己 `customer_id` 的申請。
  - `Get(GetReturnRequestRequest) → ReturnRequest`:單筆含品項明細與照片下載 URL(經 Task 3.6 下載端點)。
  - 照片取得流程:客戶端先以 Task 3.6 `POST /api/v1/files/upload`(jpeg / png / webp ≤ 5 MB)逐張上傳取得 `file_assets.id`,再於 Create 時帶入 `photo_file_ids`;退貨 API 本身不處理 binary 上傳。
- **實作邏輯**:
  1. 身分判斷:session 必須為客戶**子帳號**(`customer_id` 非空、`is_primary = false`);主帳號一律拒絕。`customer_id` 一律取自 session,不信任請求傳入值。
  2. 驗證品項:`items` 至少 1 筆;每筆 `quantity` 為正數、`reason` 非空;`source_type` 與參照欄位配對正確(`order_item` 必帶 `sales_order_item_id`,`customer_product` 必帶 `customer_product_id`)。
  3. 來源核實(兩來源並存,D25):
     - `order_item`:該 `sales_order_item` 所屬訂單必須屬於本客戶且未被軟刪除;1.0 **不檢核**訂單狀態與可退數量上限(無庫存概念,是否超退由業務審核把關)。
     - `customer_product`:該專屬商品必須屬於本客戶且未被軟刪除。
  4. 照片核實:每個 `photo_file_ids` 元素必須存在於 `file_assets`、屬於同公司且未被軟刪除;張數上限依專案常數(建議 ≤ 5 張,寫入規格待確認項)。
  5. **交易邊界(單一交易)**:寫入 `return_requests`(status = `pending`、快照 `company_id` / `department_id` / `customer_id` / `created_by_user_id`)+ 全部 `return_request_items`(快照商品名稱 / 規格 / 單位)+ audit log(動作 `return_request.create`,見 4.7.5);任一步失敗整筆 rollback(D18)。
  6. List / Get:客戶 self 範圍強制加 `customer_id = session.customer_id`;回應中照片的 `file_assets.id` 轉為下載 URL。
- **錯誤處理**:
  - `unauthenticated`:未登入或 session 失效。
  - `permission_denied`:主帳號呼叫;staff / 其他公司帳號呼叫 Create;Get 他人客戶的申請(RLS 擋下者視同 `not_found`)。
  - `invalid_argument`:`items` 為空、數量非正數、原因空白、來源欄位配對錯誤、照片張數超上限。
  - `not_found`:參照的訂單品項 / 專屬商品 / 照片檔案不存在、已軟刪除或不屬於本客戶。
- **驗收**:
  - [ ] 子帳號可分別以歷史訂單品項與專屬商品兩種來源建立申請,狀態為 `pending`。
  - [ ] 申請可附多張照片,Get 回應含可下載 URL。
  - [ ] 主帳號呼叫 Create 回 `permission_denied`;帶入他人客戶的來源品項回 `not_found`。
  - [ ] List 僅回自己客戶的申請;跨公司 session 經 RLS 查不到資料。
  - [ ] 建立成功時 `audit_logs` 同交易存在一筆 `return_request.create` 記錄。

## 子功能 4.7.3: 審核 API(approved / rejected)

- **目標**: 主責業務(或部門 dept_admin)於 Web / App 審核退貨申請:通過(`approved`)或拒絕(`rejected`,必填原因);審核**不修改原訂單狀態**。`相依: 4.7.2`
- **檔案**:
  - Create `backend/internal/domain/returns/review.go`(usecase)
  - Update `backend/internal/domain/returns/handler.go`(掛載 Review)
- **介面**: Connect-RPC `ReturnService`
  - `Review(ReviewReturnRequestRequest) → ReturnRequest`:請求含 `id`、`decision`(`approved` / `rejected`)、`reject_reason`(`decision = rejected` 時必填)、`expected_version`(樂觀鎖);回應為更新後的申請。
  - `ListPending`(或 `List` 帶 `status = pending` 篩選):staff 視角,依 data_scope 回 department / company 範圍的待審清單。
- **實作邏輯**:
  1. 權限判斷:呼叫者為 staff 身分,且為下列之一——該申請客戶的 `default_sales_rep_id` 主責業務,或該部門 `dept_admin` 以上(規格 §6.6);其餘 staff 拒絕。
  2. 讀取申請(帶 `FOR UPDATE` 或以 `version` 條件更新):不存在或已軟刪除回 `not_found`。
  3. 狀態判斷:僅 `status = pending` 可審核;已 `approved` / `rejected` 者回 `failed_precondition`(不允許二次審核 / 撤銷,1.0 無此流程)。
  4. 樂觀鎖:以 `id + version = expected_version` 為更新條件,影響列數為 0 且該筆仍存在 → `failed_precondition`(前端重新載入後再審)。
  5. 輸入判斷:`decision = rejected` 時 `reject_reason` 空白 → `invalid_argument`;`approved` 時忽略 `reject_reason`。
  6. **交易邊界(單一交易)**:更新 `status` / `reviewed_by_user_id` / `reviewed_at` / `reject_reason` / `version + 1` + 寫 audit log(動作 `return_request.review`,記錄 decision 與原因,見 4.7.5)。**本交易不得觸碰 `sales_orders` / `sales_order_items` / `sales_order_events`**(D25)。
  7. 審核通過後的推播於交易提交後觸發(見 4.7.5)。
- **錯誤處理**:
  - `unauthenticated`:未登入。
  - `permission_denied`:非主責業務且非 dept_admin 以上;客戶帳號(含子帳號)呼叫 Review。
  - `not_found`:申請不存在或已軟刪除(含跨範圍被 RLS 擋下)。
  - `failed_precondition`:非 `pending` 狀態;樂觀鎖衝突。
  - `invalid_argument`:`decision` 非法值;拒絕未填原因。
- **驗收**:
  - [ ] 主責業務可將 pending 申請審為 approved;dept_admin 亦可審核。
  - [ ] 拒絕時未填原因回 `invalid_argument`;填原因後審為 rejected 且 `reject_reason` 留存。
  - [ ] 已審核申請再次 Review 回 `failed_precondition`;併發兩人同審僅一人成功(另一人樂觀鎖衝突)。
  - [ ] 非主責業務的一般 staff 審核回 `permission_denied`。
  - [ ] 審核前後原訂單 `status` 與所有欄位不變;`audit_logs` 同交易存在一筆 `return_request.review`。

## 子功能 4.7.4: 退貨證明資料輸出

- **目標**: 審核通過後,客戶端可取得退貨證明資料(App 顯示為可出示配送司機的畫面;後端僅供資料,不產 PDF)。`相依: 4.7.3`
- **檔案**:
  - Create `backend/internal/domain/returns/certificate.go`(usecase)
  - Update `backend/internal/domain/returns/handler.go`
- **介面**: Connect-RPC `ReturnService`
  - `GetCertificate(GetReturnCertificateRequest) → ReturnCertificate`:請求含申請 `id`;回應含——申請編號(`id`)、客戶名稱與 `customer_code`、申請時間、品項清單(快照之商品名稱 / 規格 / 單位 / 數量 / 原因 / 照片下載 URL)、審核結果、審核業務姓名、審核時間、申請狀態。
  - 證明僅對 `approved` 申請輸出完整內容;`pending` / `rejected` 不提供證明(客戶端顯示審核結果走 4.7.2 的 Get)。
- **實作邏輯**:
  1. 權限判斷:呼叫者為該申請所屬客戶的子帳號(self 範圍),或具審核權限的 staff(主責業務 / dept_admin 以上,供業務側查證)。
  2. 讀取申請與品項;`status != approved` 回 `failed_precondition`(尚無證明可出示)。
  3. 品項內容一律取自申請時快照欄位(商品名稱 / 規格 / 單位),不即時 join 商品主檔,確保證明內容與審核當下一致。
  4. 照片的 `photo_file_ids` 轉為 Task 3.6 下載 URL 一併回傳,供司機現場核對。
  5. 證明查詢為唯讀,不寫稽核(查詢非 D18 關鍵操作);不產生 PDF、不經 Gotenberg(Phase 5 列印範圍不含退貨證明)。
- **錯誤處理**:
  - `unauthenticated`:未登入。
  - `permission_denied`:主帳號呼叫;他人客戶的子帳號;無審核權限的 staff。
  - `not_found`:申請不存在或已軟刪除。
  - `failed_precondition`:申請非 `approved` 狀態。
- **驗收**:
  - [ ] approved 申請可取回證明,含申請編號、客戶、品項快照、審核業務與時間、照片 URL。
  - [ ] 來源商品日後改名 / 改規格,證明內容仍顯示申請時快照。
  - [ ] pending / rejected 申請取證明回 `failed_precondition`。
  - [ ] 其他客戶的子帳號取證明回 `not_found`(RLS 隔離)。

## 子功能 4.7.5: 審核結果推播 + 申請與審核稽核

- **目標**: 審核完成後推播審核結果給發起帳號;申請建立與審核異動皆寫入稽核日誌。`相依: 4.7.3`;通知基礎相依 Task 4.3(範本 / 記錄)與 4.4(FCM / 站內發送),稽核相依 Task 2.6。
- **檔案**:
  - Create `backend/internal/domain/returns/notify.go`(審核結果通知組裝)
  - Update `backend/internal/domain/returns/review.go`(提交後觸發通知)
  - Update `backend/internal/domain/returns/create.go`(寫稽核,於 4.7.2 交易內)
  - Update `backend/internal/domain/notifications/`(新增退貨審核結果範本,依 07-notifications.md 既有範本機制)
- **介面**: 無新增對外 RPC;使用既有內部介面——
  - Task 2.6 `AuditService.Record`(內部,同事務寫入)。
  - Task 4.3 / 4.4 通知範本渲染與發送(`fcm` + `in_app` 兩通道;1.0 無 Email)。
  - 通知範本變數:申請編號、審核結果(通過 / 拒絕)、拒絕原因(拒絕時)、客戶名稱。
- **實作邏輯**:
  1. 稽核(同事務,D18):Create 交易內寫 `return_request.create`(含申請 id、品項數、customer_id、操作者 user_id);Review 交易內寫 `return_request.review`(含申請 id、decision、reject_reason、審核者 user_id)。任一步失敗整筆業務操作 rollback。
  2. 推播路由(D23):審核完成後,通知對象為**發起帳號**(`return_requests.created_by_user_id` 單一帳號,非該客戶全部帳號);主帳號不接收業務通知,但發起者必為子帳號(主帳號無法 Create),故此路由天然只落於子帳號。
  3. 觸發時機:Review **交易提交後**才發送(避免 rollback 後發出假通知);通知 `notifications` 記錄建立與 FCM / 站內發送依 Task 4.4 既有流程。
  4. 失敗處理(D16 修訂):發送失敗**不重試**,`notifications.status` 標 `failed` 並記錄原因;FCM 回報 token 失效時刪除對應 `user_devices`;通知失敗**不回滾**已完成的審核。
  5. 站內通知同時入通知中心,客戶子帳號於 App 可查審核結果;推播文案不含金額(全系統無金額欄位)。
- **錯誤處理**:
  - 通知範本缺失或渲染失敗:記錄 `failed` 通知與後端 log,審核結果不回滾(稽核已同事務保證,通知為事後副作用)。
  - 發起帳號已停用或無有效裝置 token:僅站內通知可達,FCM 側記錄 `failed`;不視為錯誤回覆給審核者。
  - 本子功能不新增對外錯誤碼;Review 本身錯誤碼見 4.7.3。
- **驗收**:
  - [ ] 審核通過 / 拒絕後,發起帳號收到 FCM 推播與站內通知;拒絕時文案含拒絕原因。
  - [ ] 同一客戶的其他子帳號與主帳號不收到此審核通知(僅發起帳號)。
  - [ ] 發送失敗時 `notifications.status = failed` 且原因可查,審核狀態不受影響。
  - [ ] `audit_logs` 可依資源篩選出 `return_request.create` 與 `return_request.review`,且與業務操作同事務(模擬寫稽核失敗時業務操作一併 rollback)。

---

## 整合測試重點(D21:退貨審核列為關鍵路徑,需整合測試)

後端以 testify + dockertest 對真實 PostgreSQL(含 RLS)執行,至少覆蓋:

1. **審核狀態機**:pending → approved / rejected 成立;重複審核、`approved` 再改 `rejected` 皆回 `failed_precondition`。
2. **樂觀鎖併發**:兩個 staff 同時 Review 同一申請,僅一筆成功,另一筆 `failed_precondition`。
3. **審核權限**:主責業務與 dept_admin 可審;非主責 staff、客戶帳號、主帳號呼叫 Review / Create 分別回 `permission_denied`。
4. **RLS 隔離**:客戶 A 的子帳號 List / Get / GetCertificate 查不到客戶 B 的申請;跨公司 staff 不可見。
5. **不修改原訂單**:審核前後比對 `sales_orders` 與 `sales_order_events`,確認零異動(D25)。
6. **稽核同事務**:Create / Review 成功必有對應 audit log;注入稽核寫入失敗,業務操作整筆 rollback(D18)。
7. **推播路由**:審核後僅發起帳號有通知記錄(FCM + 站內),同客戶其他帳號無;FCM 失敗標 `failed` 不重試、不影響審核結果。
8. **雙來源建立**:歷史訂單品項與專屬商品兩種 `source_type` 皆可建單;帶入他人客戶的來源品項回 `not_found`。
9. **退貨證明快照**:審核通過後修改來源商品名稱,GetCertificate 仍回申請時快照內容。

---

*最後更新:2026-08-17*
