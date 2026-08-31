# 05. 銷售訂單(Backend 細部實作計畫)

---
**文件標題**:銷售訂單 — 訂單 API 與下單流程
**對應原計畫 Task**:4.1(訂單 API)、4.2(下單流程與單位換算)— `docs/superpowers/plans/2026-07-17-sales-order-1-0-tasks.md`
**對應規格書 v1.0.34 章節**:`docs/superpowers/specs/1.0-requirements/sales-orders/spec.md`(訂單狀態機、取消派車、訂單作廢與更正、訂單異動事件與稽核記錄、訂單編號與併發取號、不儲存任何金額欄位、下單流程、訂單明細與單位換算、客戶專屬商品清單);另參照 `master-data`(customers.preferred_delivery_days、customer_products)與 `notifications`(訂單建立通知)。
**相關決策**:D3(租戶三層)、D7(訂單編號取號)、D10(軟刪除)、D11(訂單來源字典)、D12(不存金額)、D13(狀態機與事件)、D14(派車欄位與 version)、D18(稽核同事務)、D21(整合測試)、D26(偏好送貨日順延)。
**相依文件**:`00-index.md`(共通規則)、`01-auth.md`(認證與 data_scope)、`03-metadicts-audit.md`(訂單來源字典、audit_logs)、`04-master-data.md`(3.1 客戶、3.3 商品與單位換算、3.5 客戶專屬商品)。**下游**:`06-returns.md`(退貨不更動訂單狀態)、`08-dispatch.md`(派車確認 → processing、取消派車回退、看板 version)、`09-printing.md`(processing 方可正式列印)。

---

## 共通規則(本文件適用,細節見 `00-index.md` §3)

- **交易邊界**:取號 + 建單、狀態異動 + `sales_order_events`、作廢/取消派車 + `audit_logs`,皆同一 DB 交易,同成功同失敗(D7、D13、D18)。各子功能「實作邏輯」欄標明交易邊界。
- **軟刪除**:`sales_orders` / `sales_order_items` 採 `deleted_at` + 部分唯一索引(D10);`sales_order_events` 為僅追加記錄,**不適用**軟刪除,不得修改或刪除(D10、規格書「事件記錄不可竄改」)。
- **多租戶**:訂單與明細皆帶 `company_id` / `department_id`;Casbin 管功能權限、RLS 注入(`app.current_company_id` / `app.current_department_id` / `app.current_data_scope`)為最後防線(D3)。客戶帳號 data_scope 為 self,僅見自己客戶的訂單。
- **金額禁令(D12)**:`sales_orders` / `sales_order_items` **不得**出現任何金額欄位(單價、小計、稅額、折扣、總額、幣別);schema review 與 migration lint 應阻擋此類欄位。時間欄位一律 `timestamptz`,顯示 UTC+8。
- **樂觀鎖衝突**一律回 `failed_precondition`(取號衝突由後端自動重試,不外拋;拖放/編輯 version 衝突外拋供前端重查)。
- **錯誤碼約定**:`unauthenticated` / `permission_denied` / `not_found` / `failed_precondition` / `invalid_argument` / `already_exists`,語意見 `00-index.md` §3.4。
- **角色用語**:「dept_admin 以上」指 `dept_admin`、`company_admin`、`super`(及開發環境 `developer` 逃生門);實際判定由 Casbin policy 執行,非硬編碼角色字串。

---

## Task 4.1:訂單 API

### 子功能 4.1.1: 訂單 schema(sales_orders / sales_order_items / sales_order_events)

- **目標**: 建立訂單三表 Ent schema 與 migration,承載狀態機、來源、派車欄位與事件軌跡;不含任何金額欄位(D12)。
- **檔案**:
  - Create `backend/internal/domain/salesorders/schema/sales_order.go`(Ent schema)
  - Create `backend/internal/domain/salesorders/schema/sales_order_item.go`
  - Create `backend/internal/domain/salesorders/schema/sales_order_event.go`
  - Create `backend/migrations/`(對應 migration,依 Phase 0 migrate 流程)
- **介面**(Ent 實體):
  - `sales_orders`:`id`(UUID)、`company_id`、`department_id`、`order_no`(公司內唯一)、`customer_id`、`source`(訂單來源碼,對應 metadicts 系統級「訂單來源」,如 `W` / `A`)、`status`(enum:`pending` / `processing` / `completed` / `cancelled` / `voided`,預設 `pending`)、`expected_delivery_date`(date)、`sales_rep_id`(負責業務)、`note`、`dispatched_at` / `dispatched_by`(nullable)、`route_id` / `delivery_sequence`(nullable,看板位置)、`version`(整數樂觀鎖,D14 看板拖放用)、`created_by`、`created_at` / `updated_at` / `deleted_at`。
  - `sales_order_items`:`id`、`sales_order_id`(FK)、`company_id` / `department_id`(冗餘以利 RLS)、`product_id`(FK,nullable — 見 4.2.2 手打品名)、`display_name`(名稱快照:別名或手打名稱;軟刪除商品後歷史訂單仍可顯示,D10)、`qty`(numeric)、`unit`(文字,下單時選用單位)、`base_qty`(numeric,換算基本單位後數量,4.2.1)、`cutting_spec_id`(nullable)、`special_cut_note`、`warehouse_id`(nullable)、`sort_order`、`created_at` / `updated_at` / `deleted_at`。
  - `sales_order_events`:`id`、`sales_order_id`(FK)、`company_id`、`event_type`(enum:`create` / `edit` / `dispatch` / `dispatch_cancel` / `cancel` / `complete` / `void`)、`actor_id`、`reason`(nullable,僅 dispatch_cancel / void 必填語意)、`payload`(JSON,異動摘要,如狀態前後值)、`created_at`。**無** `updated_at` / `deleted_at`(僅追加)。
  - 索引:`sales_orders` 部分唯一 `(company_id, order_no) WHERE deleted_at IS NULL`;`(company_id, department_id, status, expected_delivery_date)` 複合索引供列表與看板;`sales_order_events` 索引 `(sales_order_id, created_at)`。
- **實作邏輯**:
  1. 依上列欄位定義三張 Ent schema;`status` / `event_type` 以 enum 型別約束,不允許自由文字。
  2. 確認欄位清單不含任何金額語意欄位(price / amount / subtotal / tax / discount / total / currency),D12 為硬性邊界。
  3. `order_no` 唯一性以部分唯一索引實作(軟刪除後同號不可重用亦不可撞號 — 序號由計數器保證不回退,見 4.1.2)。
  4. `dispatched_at` / `dispatched_by` / `route_id` / `delivery_sequence` 於派車域(08-dispatch)寫入,本 schema 僅定義欄位與預設 NULL。
  5. `version` 預設 1,任何看板位置或狀態異動由 update 時遞增(拖放比對見 08-dispatch)。
  6. Migration 一併建立 `sales_orders` / `sales_order_items` / `sales_order_events` 的 RLS policy(company/department/data_scope 三級,同 02-tenancy 模式);events 表僅開 SELECT/INSERT,REVOKE UPDATE/DELETE。
- **錯誤處理**: migration 失敗 → 部署失敗即停(fail-fast);RLS policy 缺失 → 視為部署錯誤,不得進入運行。
- **驗收**:
  - [ ] 三表建立成功,欄位與索引如上;`sales_order_events` 無更新/刪除權限。
  - [ ] 主表與明細不含任何金額欄位(可於 migration lint 或 schema 測試斷言)。
  - [ ] `(company_id, order_no)` 部分唯一索引生效:同公司同號拒絕,異公司同號可共存。
  - [ ] RLS 開啟後,跨公司/跨部門連線查不到他方訂單。

### 子功能 4.1.2: order_counters 樂觀鎖取號

- **目標**: 實作「來源碼 + 6 位補零序號」訂單編號,`order_counters` 依 `company_id + source` 一列,樂觀鎖更新,version 衝突自動重試;取號與建單同一交易(D7)。相依: 4.1.1。
- **檔案**:
  - Create `backend/internal/domain/salesorders/schema/order_counter.go`
  - Create `backend/internal/domain/salesorders/numbering.go`(取號邏輯)
  - Create `backend/migrations/`(order_counters 表)
- **介面**:
  - Ent 實體 `order_counters`:`id`、`company_id`、`source`、`next_seq`(下一可用序號)、`version`(樂觀鎖)、`created_at` / `updated_at`;唯一鍵 `(company_id, source)`。
  - 內部函式(非 RPC):`NextOrderNo(ctx, tx, companyID, source) (orderNo string, err error)` — 僅接受已開啟的交易物件,不自行開交易。
- **實作邏輯**(全部在呼叫方傳入的交易內執行):
  1. 查詢 `(company_id, source)` 列;不存在則插入 `next_seq = 1, version = 1` 並直接取用序號 1(插入撞唯一鍵時轉為重新查詢)。
  2. 以「`version` 相符才更新」的條件更新:`next_seq + 1`、`version + 1`;影響列數為 0 代表併發衝突。
  3. 衝突時重讀該列並重試,重試上限(建議 5 次,指數退避可省略 — 量級低);超過上限回內部錯誤(實務上不應發生)。
  4. 取號成功 → 組裝編號:來源碼(字典 `code`,如 `W`)+ 序號格式化為 6 位補零(如 `W000123`);序號超過 999999 時自然增長位數(不截斷)。
  5. 交易由呼叫方(4.1.4 Create)統一提交或回滾:建單失敗 → 整體回滾,序號不消耗,下次取號仍取得同一序號。
  6. 不同來源各自獨立遞增;不同公司互不干擾。
- **錯誤處理**: `source` 非系統字典有效碼 → `invalid_argument`;重試上限耗盡 → `internal`(並記錄告警日誌);交易回滾不對外暴露錯誤碼,由上層 Create 統一回應。
- **驗收**:
  - [ ] 同公司同來源連續取號:序號連續遞增且無重複(如 `W000123`、`W000124`)。
  - [ ] 同公司不同來源各自獨立序號軌(如 `A000001`)。
  - [ ] 併發兩請求同來源取號:一方 version 衝突重試,兩者取得不同且連續序號,無遺失(整合測試,見文末)。
  - [ ] 取號後建單失敗回滾:下次取號取得同一序號(序號不被消耗)。

### 子功能 4.1.3: 訂單狀態機

- **目標**: 以單一狀態機模組守衛所有狀態流轉:`pending ⇄ processing → completed`、`pending → cancelled`、`completed → voided`;所有異動寫 `sales_order_events`;作廢需 dept_admin 以上 + 原因(D13)。相依: 4.1.1。
- **檔案**:
  - Create `backend/internal/domain/salesorders/statemachine.go`(轉移表與守衛)
- **介面**: 內部函式(非 RPC):`CanTransition(from, to status) bool`、`Transition(ctx, tx, order, to, actor, reason) error` — 由 usecase 層(Cancel / Complete / Void 與 08-dispatch 的派車確認 / 取消派車)統一呼叫,禁止各 RPC 自行改 `status` 欄位。
- **實作邏輯**(每個轉移皆在單一交易內完成:狀態更新 + 事件 + 必要時稽核):
  1. 定義合法轉移表:`pending → processing`(派車確認,由 08-dispatch 觸發,同時寫 `dispatched_at` / `dispatched_by`);`processing → pending`(取消派車,dept_admin 以上 + 原因,清 `dispatched_at` / `dispatched_by`,**保留** `route_id` / `delivery_sequence`);`processing → completed`(出貨完成);`pending → cancelled`(取消);`completed → voided`(作廢,dept_admin 以上 + 原因)。其餘組合一律非法。
  2. 終態規則:`cancelled` 與 `voided` 不接受任何轉移;`completed` 僅可轉 `voided`,不提供直接編輯(更正 = 作廢 + 重建,新單備註記載原單號)。
  3. 每次轉移寫一筆 `sales_order_events`:`create` / `edit` / `dispatch` / `dispatch_cancel` / `cancel` / `complete` / `void`,含 `actor_id`、`created_at`、`reason`(dispatch_cancel 與 void 必填)。
  4. `dispatch_cancel` 與 `void` 除事件外,**同交易**寫入 `audit_logs`(操作人、時間、原因、異動前後狀態,D13 / D18)。
  5. 角色檢查:`dispatch_cancel`、`void` 需 dept_admin 以上(由 Casbin 判定);`cancel` 依下單來源角色規則(見 4.1.4)。
  6. 狀態更新採條件更新(WHERE status = 預期前值):影響列數 0 → 狀態已被併發改變,回 `failed_precondition`。
  7. 任何轉移遞增 `sales_orders.version`(供看板樂觀鎖,D14)。
- **錯誤處理**: 非法轉移(含終態異動、completed 直接編輯嘗試)→ `failed_precondition`,訊息指明目前狀態與允許動作;角色不足 → `permission_denied`;應填原因未填 → `invalid_argument`;併發狀態改變 → `failed_precondition`(前端重查)。
- **驗收**:
  - [ ] 合法路徑全通過:pending → processing → completed;pending → cancelled;processing → pending(回退);completed → voided。
  - [ ] 非法路徑全拒絕:編輯 processing/completed、直接取消 processing(提示須先取消派車)、對 cancelled/voided 任何異動。
  - [ ] 每次轉移各產生一筆對應 `event_type` 事件,含操作人/時間/原因。
  - [ ] `dispatch_cancel` 與 `void` 同時存在事件與 `audit_logs` 記錄,且同一交易(模擬稽核寫入失敗 → 狀態不變)。
  - [ ] 取消派車後 `dispatched_at` / `dispatched_by` 清空、`route_id` / `delivery_sequence` 原值保留。
  - [ ] 作廢由 staff 執行或未填原因 → 拒絕,狀態維持。

### 子功能 4.1.4: 訂單 CRUD + 軟刪除與取消

- **目標**: 提供 `SalesOrderService` 的 List / Get / Create / Update / Cancel / Complete / Void RPC;待處理可編輯可取消、處理中起不可編輯;軟刪除僅限 pending / cancelled。相依: 4.1.1、4.1.2、4.1.3。
- **檔案**:
  - Create `backend/proto/v1/salesorder.proto`
  - Create `backend/internal/domain/salesorders/handler.go`(Connect handler)
  - Create `backend/internal/domain/salesorders/usecase.go`
  - Create `backend/internal/domain/salesorders/repo.go`
- **介面**(Connect-RPC,package `v1`,Service `salesorder.v1.SalesOrderService`):
  - `ListOrders(ListOrdersRequest{page, page_size, status?, customer_id?, expected_delivery_date_range?, source?, keyword?}) → ListOrdersResponse{orders[], total}` — 預設排除 `deleted_at`;客戶帳號 data_scope self 自動限縮為自己。
  - `GetOrder(GetOrderRequest{id}) → GetOrderResponse{order, items[]}`。
  - `CreateOrder(CreateOrderRequest{customer_id, source, expected_delivery_date, note?, sales_rep_id?, items[]{product_id?, manual_name?, display_name, qty, unit, cutting_spec_id?, special_cut_note?, warehouse_id?}, save_alias[]?}) → CreateOrderResponse{order}` — 下單組裝邏輯見 4.2.1–4.2.5。
  - `UpdateOrder(UpdateOrderRequest{id, version, expected_delivery_date?, note?, items[]?}) → UpdateOrderResponse{order}` — 僅 pending;攜帶 `version` 樂觀鎖。
  - `CancelOrder(CancelOrderRequest{id, reason?}) → CancelOrderResponse{order}`。
  - `CompleteOrder(CompleteOrderRequest{id}) → CompleteOrderResponse{order}`。
  - `VoidOrder(VoidOrderRequest{id, reason}) → VoidOrderResponse{order}`。
- **實作邏輯**:
  1. **Create**(交易邊界 = 整個步驟):驗證客戶存在且未軟刪除、呼叫者可見(Casbin + RLS)→ 依 4.2.3 檢查明細品項合法性 → 依 4.2.1 逐項計算 `base_qty` → 依 4.2.2 處理手打與別名 → 依 4.2.5 順延 `expected_delivery_date` → 呼叫 4.1.2 取號(同事務)→ 插入訂單(`status = pending`、`version = 1`)與明細 → 寫 `create` 事件 → 提交。提交後(交易外)觸發訂單建立通知,交 07-notifications(業務下單推播該客戶子帳號;客戶自行下單不另行通知)。
  2. **Update**:載入訂單 → 狀態必須 `pending`,否則 `failed_precondition`(processing 提示須先取消派車)→ 比對請求 `version` 與現值,不符回 `failed_precondition` → 整單替換明細(軟刪除舊明細 + 插入新明細,重跑 4.2.1 換算與 4.2.5 順延)→ `version + 1` → 寫 `edit` 事件(含異動摘要 payload)→ 同交易提交。
  3. **Cancel**:僅 `pending` → 走 4.1.3 轉移 `cancelled`,寫 `cancel` 事件。客戶帳號僅能取消自己客戶的訂單(RLS self);員工依 Casbin 權限。
  4. **Complete**:僅 `processing` → 轉移 `completed`,寫 `complete` 事件。
  5. **Void**:僅 `completed`,dept_admin 以上,`reason` 必填 → 轉移 `voided`,事件 + 稽核同交易(4.1.3)。
  6. **軟刪除**(Delete / 或 Update 旗標,依 proto 定一種):僅 `pending` 或 `cancelled` 訂單可軟刪除 → 設 `deleted_at` + 連同明細軟刪除 + 寫稽核(D10/D18),同交易。`processing` / `completed` / `voided` 拒絕(歷史與單據需保留)。
  7. List / Get:預設排除軟刪除;查無或已軟刪除 → `not_found`;RLS 為最後防線,Casbin 先擋功能權限。
- **錯誤處理**: 未登入/token 失效 → `unauthenticated`;無訂單功能權限、客戶越權操作他人訂單 → `permission_denied`;訂單不存在或已軟刪除 → `not_found`;狀態不允許(編輯 processing/completed、取消非 pending、作廢非 completed、軟刪除 processing/completed/voided)、version 衝突 → `failed_precondition`;輸入驗證失敗(空明細、qty ≤ 0、缺客戶、作廢未填原因)→ `invalid_argument`;唯一衝突(理論上不應發生,取號保證)→ `already_exists`。
- **驗收**:
  - [ ] 待處理訂單可編輯明細/出貨日/備註,可取消;取消後拒絕任何異動與編輯。
  - [ ] 處理中訂單 Update 被拒並提示須先取消派車;completed 訂單不可編輯。
  - [ ] Create 全程單一交易:明細任一項驗證失敗 → 訂單、序號、別名皆不殘留。
  - [ ] 軟刪除僅 pending/cancelled 成功;軟刪除後 List 預設不可見、Get 回 not_found、稽核有記錄。
  - [ ] version 不符的 Update 被拒,重查後以新 version 可成功。

### 子功能 4.1.5: ListEvents 異動軌跡查詢

- **目標**: 提供訂單異動軌跡查詢,供訂單詳情頁與 App 顯示完整操作歷程。相依: 4.1.1、4.1.3。
- **檔案**:
  - Update `backend/proto/v1/salesorder.proto`(新增 RPC 與訊息)
  - Update `backend/internal/domain/salesorders/handler.go`、`usecase.go`、`repo.go`
- **介面**(Connect-RPC):`ListOrderEvents(ListOrderEventsRequest{sales_order_id, page?, page_size?}) → ListOrderEventsResponse{events[]{event_type, actor_id, actor_name, reason?, payload?, created_at}, total}` — 依 `created_at` 升序回傳。
- **實作邏輯**:
  1. 驗證呼叫者對該訂單的可見性:先走 GetOrder 同套權限(Casbin 功能權限 + RLS 資料範圍;客戶帳號 self 僅自己訂單),訂單不可見 → 同 `not_found`(不洩漏存在性)。
  2. 查詢 `sales_order_events`,關聯操作者姓名快照(join users 取當前姓名即可,1.0 不做姓名快照)。
  3. 分頁預設足夠大(事件量小),保持升序;不回傳他公司/他部門事件(RLS 自動限縮)。
  4. 不提供任何修改/刪除事件的 API;DB 層已 REVOKE(4.1.1)。
- **錯誤處理**: 訂單不存在/不可見 → `not_found`;未登入 → `unauthenticated`;無訂單讀取權限 → `permission_denied`。
- **驗收**:
  - [ ] 建立 → 編輯 → 派車 → 取消派車 → 完成 → 作廢的訂單,依序回傳六類事件,各含操作人、時間、原因(適用時)。
  - [ ] 客戶帳號僅能查自己訂單的事件;查他人訂單回 `not_found`。
  - [ ] 系統無任何更新/刪除事件的入口(API 與 DB 權限雙重)。

---

## Task 4.2:下單流程與單位換算

> 本子 Task 全部落於 `backend/internal/domain/salesorders/usecase.go` 的 Create / Update 組裝流程,與 4.1.4 的交易邊界一致(整個 Create 一個交易)。

### 子功能 4.2.1: 單位換算套用

- **目標**: 下單/編輯時依選用單位自動換算,於明細 `base_qty` 保存基本單位數量,供揀貨單與加工單彙總。相依: 3.3.3(商品單位換算,`04-master-data.md`)。
- **檔案**:
  - Update `backend/internal/domain/salesorders/usecase.go`
- **介面**: 內部函式(非 RPC):`ResolveBaseQty(ctx, tx, productID, unit, qty) (baseQty, err error)` — 於 Create / Update 逐明細呼叫;輸入為 CreateOrderRequest 的 `qty` / `unit`。
- **實作邏輯**:
  1. 讀取商品的 `product_units` 清單(3.3.3),找出 `unit` 對應列;找不到 → `invalid_argument`(單位不屬於該商品)。
  2. 若該列 `is_base = true`:`base_qty = qty`,不換算。
  3. 否則 `base_qty = qty × conversion_rate`(換算為 `is_base = true` 單位的數量,如 1 條 = 0.6 kg,10 條 → 6 kg)。
  4. 數值以高精度 decimal 計算,避免浮點誤差;結果精度跟隨商品單位定義,不額外四捨五入(單據彙總端自行處理顯示)。
  5. 手打品名(無 `product_id`,見 4.2.2)無換算來源:`base_qty = qty` 直接照填,不阻擋下單。
  6. 於 Create / Update 交易中逐項寫入明細 `qty` / `unit` / `base_qty`;換算失敗 → 整單回滾。
- **錯誤處理**: 單位不存在於該商品 → `invalid_argument`;`qty ≤ 0` → `invalid_argument`;商品不存在或已軟刪除 → `not_found`(以驗證失敗表述於 `invalid_argument` 細節亦可,統一採 `invalid_argument` 附欄位錯誤)。
- **驗收**:
  - [ ] 非基本單位:`qty=10`、`unit=條`、換算率 0.6 → `base_qty=6`。
  - [ ] 基本單位:`base_qty = qty`。
  - [ ] 無效單位或非正數量 → `invalid_argument`,整單不建立。
  - [ ] Update 改單位或數量後,`base_qty` 重新計算且正確。

### 子功能 4.2.2: 手打商品別名建立

- **目標**: 業務下單時可手打商品名稱;選擇儲存時自動建立/更新該客戶的 `customer_products` 別名;從商品總表選用而未建檔的商品亦自動加入清單。相依: 3.5.3(下單自動建立別名,`04-master-data.md`)、4.2.1。
- **檔案**:
  - Update `backend/internal/domain/salesorders/usecase.go`
- **介面**: CreateOrderRequest 明細層支援兩種品項形狀:`{product_id, display_name, save_alias?}`(選品改名/手打綁定)或 `{manual_name}`(純手打,無商品關聯);`save_alias = true` 時於同交易寫入 `customer_products`。
- **實作邏輯**:
  1. 手打品名(業務限定;客戶帳號不可手打,見 4.2.3):明細以 `manual_name` 建立,`product_id` 為 NULL,`display_name = manual_name`,`base_qty = qty`(4.2.1 第 5 點)。
  2. 選品 + 改名(「A 商品在不同客戶叫 BCDEF」需求):`product_id` 關聯總表商品,`display_name` 存業務輸入名稱快照。
  3. `save_alias = true` 且帶 `product_id` → 同交易 upsert `customer_products`(customer_id + product_id):存在(含軟刪除復原語意依 3.5.3)則更新 `alias_name`;不存在則建立,`alias_name = display_name`、其餘預設值(default_qty 可帶本次 qty)。唯一鍵衝突(同客戶同商品已有別名)→ 更新既有別名,**不**回 `already_exists`(下單不應被別名擋下);若業務明確要保留舊別名,前端不送 `save_alias`。
  4. 選用總表商品而該客戶 `customer_products` 尚無此商品 → 同交易自動建立清單記錄(預設值帶入,依 3.5.3),`display_name` 預設為商品名稱。
  5. 「詢問是否儲存」為前端互動;後端僅依 `save_alias` 旗標執行,不產生中間態。
  6. 全部寫入與訂單 Create / Update 同一交易;`customer_products` 寫入失敗 → 整單回滾。
- **錯誤處理**: 客戶帳號送出手打品項 → `permission_denied`;`manual_name` 空白 → `invalid_argument`;客戶不存在/已軟刪除 → `not_found`;別名 upsert 撞唯一索引(併發下同客戶同商品)→ 重讀後改為更新,仍失敗則整單回滾回 `internal`。
- **驗收**:
  - [ ] 手打品項下單成功,明細 `product_id` 為 NULL、`display_name` 為手打名稱。
  - [ ] `save_alias = true` 下單後,`customer_products` 出現/更新該客戶對該商品的別名,下次帶出清單即見新別名。
  - [ ] 同客戶同商品不重複建第二筆清單記錄(一客戶一商品一筆)。
  - [ ] 選用總表商品自動加入該客戶清單;清單建立失敗時整單不存在(交易回滾)。

### 子功能 4.2.3: 客戶限定專屬清單下單

- **目標**: 客戶帳號下單時,客戶固定為自己,且品項必須來自自己的 `customer_products`;業務帳號不受此限(可總表選品或手打)。相依: 3.5(客戶專屬商品清單)、4.1.4。
- **檔案**:
  - Update `backend/internal/domain/salesorders/usecase.go`
- **介面**: 無新增 RPC;為 `CreateOrder` / `UpdateOrder` 內的品項來源守衛。判定依據 = 呼叫者身分(客戶帳號 vs 員工)與其 data_scope(self)。
- **實作邏輯**:
  1. 識別呼叫者類型:客戶帳號(含主帳號 — 主帳號依 D28 業務 API 一律 `permission_denied`,不進入本流程)與員工帳號。
  2. 客戶子帳號下單:忽略請求中的 `customer_id`,強制為自己客戶;逐項檢查 `product_id` 必須存在於該客戶未軟刪除的 `customer_products`;任一品項不在清單 → 拒絕整單。
  3. 客戶品項不允許 `manual_name`(手打為業務功能);送出即拒絕。
  4. 員工下單:不限品項來源,但 `customer_id` 必須在其 data_scope(部門/公司)內且可見;`sales_rep_id` 預設為操作者本人,可指定同部門業務。
  5. 守衛在取號之前執行:品項不合法不消耗序號。
  6. 雙重防線:usecase 守衛為主,RLS(self 範圍)為最後防線 — 即使守衛遺漏,客戶連線亦讀寫不到他人資料。
- **錯誤處理**: 主帳號呼叫 → `permission_denied`;品項不在專屬清單、客戶送出手打品項 → `invalid_argument`(訊息指明品項與原因);`customer_id` 越權(員工跨範圍)→ `permission_denied`;客戶/商品已軟刪除 → `not_found`。
- **驗收**:
  - [ ] 客戶子帳號以清單內商品下單成功;夾帶清單外商品 → 整單拒絕、無序號消耗。
  - [ ] 客戶請求帶他人 `customer_id` → 被忽略並落單於自己(或拒絕,實作擇一並於 API 文件註明;建議拒絕以避免誤單)。
  - [ ] 主帳號下單 → `permission_denied`。
  - [ ] 業務對自己部門客戶下單成功;跨部門客戶(data_scope department)→ `permission_denied`。

### 子功能 4.2.4: source 記錄(Web / App)

- **目標**: 每筆訂單記錄下單來源,來源碼決定訂單編號前綴;來源值受系統字典約束。相依: 4.1.2、03-metadicts(訂單來源字典,D11 系統級固定)。
- **檔案**:
  - Update `backend/internal/domain/salesorders/usecase.go`
  - Update `backend/proto/v1/salesorder.proto`(CreateOrderRequest.source 欄位說明)
- **介面**: `CreateOrderRequest.source`(字串,字典碼,如 `W` = Web 中台、`A` = App);`SalesOrder.source` 於 List / Get 回傳。
- **實作邏輯**:
  1. 驗證 `source` 為 metadicts「訂單來源」系統級字典的有效 `code` 且 `is_active`;無效 → `invalid_argument`。
  2. 同一 `source` 值同時用於:訂單 `source` 欄位、4.1.2 取號軌道(`order_counters.source`)、編號前綴(來源碼本身)。
  3. 來源由呼叫端如實申報(Web 中台送 `W`、App 送 `A`);1.0 不做通道指紋校驗,但 employee/customer 身分與 source 組合不設限(業務可從 Web 或 App 下單)。
  4. `source` 建立後不可修改(Update 不接受此欄位)。
  5. List 支援以 `source` 篩選(4.1.4 介面已含)。
- **錯誤處理**: 缺 `source` 或無效碼 → `invalid_argument`;字典查詢失敗(部署缺 seed)→ `internal` 並告警。
- **驗收**:
  - [ ] Web 下單編號形如 `W000xxx`,App 下單形如 `A000xxx`,兩軌序號各自遞增。
  - [ ] 無效來源碼被拒;訂單建立後 `source` 不可變更。
  - [ ] List 以 `source` 篩選結果正確。

### 子功能 4.2.5: 偏好送貨日順延(D26)

- **目標**: 下單/編輯選擇預計出貨日時,若客戶設有偏好送貨日且所選日非勾選日,自動順延至下一個勾選日;未勾選任何日則維持原選擇。相依: 3.1.5(customers.preferred_delivery_days,`04-master-data.md`)。
- **檔案**:
  - Update `backend/internal/domain/salesorders/usecase.go`
- **介面**: 內部函式(非 RPC):`AdjustDeliveryDate(preferredDays []bool(週一至週六), date) adjustedDate`;Create / Update 回應中的 `expected_delivery_date` 為**順延後**結果(前端據以於確認畫面顯示)。
- **實作邏輯**:
  1. 讀取 `customers.preferred_delivery_days`(JSON 布林陣列,索引對應星期一至六)。
  2. 客戶未勾選任何日(全 false 或空)→ 維持原選擇,不順延。
  3. 所選日期的星期落於勾選日(星期日視同非勾選 — 勾選範圍僅週一至週六)→ 維持原選擇。
  4. 否則逐日往後找下一個勾選日(跨週持續搜尋,至多 7 天內必命中,因至少勾選一日)→ 以該日為 `expected_delivery_date`。
  5. 順延於 Create 與 Update(修改出貨日時)皆執行;順延結果直接落庫,不回傳「原選擇 + 建議」雙值(單一事實來源)。
  6. 日期一律以 UTC+8 營業日曆判斷星期,避免伺服器時區造成偏移。
- **錯誤處理**: `expected_delivery_date` 缺漏或格式錯誤 → `invalid_argument`;客戶不存在 → `not_found`;偏好陣列長度異常(非 6)→ 視同未勾選,維持原選擇並記錄警告日誌(不擋單)。
- **驗收**:
  - [ ] 偏好為週二、週四,選週三 → 自動順延為週四。
  - [ ] 選日落於勾選日 → 維持原日。
  - [ ] 客戶未勾選任何日 → 維持原選擇。
  - [ ] 跨週情境(偏好僅週一,選週五)→ 順延至下週一。
  - [ ] Update 修改出貨日時同樣套用順延。
  - [ ] 整合測試覆蓋本規則(D21 關鍵路徑,見文末)。

---

## 整合測試重點

依 D21,下列三條為本 domain 的關鍵路徑整合測試(真實 PostgreSQL、RLS 啟用、Connect handler 端到端):

1. **訂單狀態機**(對應 4.1.3 / 4.1.4):
   - 全合法路徑走查:建立(pending)→ 派車確認(processing,由 dispatch 域觸發)→ 完成(completed);pending → cancelled;processing →(取消派車,dept_admin + 原因)→ pending;completed →(作廢,dept_admin + 原因)→ voided。
   - 非法路徑:編輯 processing / completed、直接取消 processing(訊息須提示先取消派車)、對 cancelled / voided 任何異動、staff 作廢、未填原因作廢 — 全部拒絕且狀態不變。
   - 每條路徑斷言 `sales_order_events` 事件型別與筆數;取消派車與作廢另斷言 `audit_logs` 存在且同交易(可注入稽核寫入失敗驗證狀態回滾)。
2. **取號併發**(對應 4.1.2):
   - 同公司同來源 N 條併發 Create:全部成功、編號不重複、序號連續無遺失。
   - 同公司異來源並行:兩軌序號各自連續。
   - 取號後強制建單失敗(如明細驗證失敗):交易回滾、序號未消耗,下次取號取得同一序號。
3. **偏好送貨日順延**(對應 4.2.5):
   - 偏好週二/週四選週三 → 落庫為週四;選週四 → 維持;未勾選 → 維持;僅勾週一選週五 → 順延至下週一(跨週)。
   - Update 修改出貨日同樣順延;回應值即順延後日期。

另建議併入同批測試(非 D21 硬性列舉但為本文件驗收所依):客戶專屬清單守衛(4.2.3)跨客戶越權、軟刪除後 List/Get 不可見(4.1.4)、`sales_order_events` 無更新刪除入口(4.1.5)、明細不含金額欄位的 schema 斷言(4.1.1)。

---

*最後更新:2026-08-17*
