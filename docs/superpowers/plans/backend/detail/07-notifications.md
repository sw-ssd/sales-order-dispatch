# 07 — 通知系統(Task 4.3、4.4)細部實作計畫

> **文件標題**:通知範本、通知記錄、裝置管理、FCM / 站內發送
> **對應原計畫 Task**:4.3(通知系統 API)、4.4(FCM / 站內通知發送)
> **對應規格書 v1.0.34 章節**:§5.1 核心實體(`notification_templates` / `notifications` / `user_devices`)、§5.2 關鍵欄位、§5.4 軟刪除慣例、§6.2 下單流程(通知觸發)、§6.3 客戶專屬商品清單、§10.1 通知
> **對應決策記錄**:D3(租戶隔離)、D10(軟刪除)、D16(通知兩通道、失敗不重試、失效 token 即刪)、D18(稽核同事務)、D21(測試策略)、D22(一主多子)、D23(App 推播路由)、D24(促銷推播與公告分離)
> **相依文件**:`01-auth.md`(身分、主帳號/子帳號判別)、`02-tenancy-users.md`(使用者/角色、dept_admin)、`03-metadicts-audit.md`(稽核日誌寫入)、`04-master-data.md`(customers.default_sales_rep_id、customer_products、promo 欄位宿主)、`05-sales-orders.md`(訂單建立觸發點)、`06-returns.md`(退貨審核觸發點)

---

## 0. 共通規則(本文件適用,引用 `00-index.md` §3,不重複全文)

- **交易與稽核(D18)**:通知觸發方的業務交易(下單、審核、新增專屬商品)與稽核日誌同一 DB 交易;通知記錄(`notifications`)的建立採同事務寫入 `pending`,實際對 FCM 的外部呼叫在交易提交後執行(見 4.4.1),提交後失敗不回滾業務,只標記 `failed`。
- **軟刪除(D10)**:`notification_templates`、`user_devices` 適用軟刪除 + 部分唯一索引;`notifications` **不適用軟刪除**(通知記錄不可刪除,規格 §5.4)。
- **多租戶(D3)**:`notification_templates` / `notifications` 帶 `company_id` + `department_id`;`user_devices` 帶 `user_id` + `company_id`;RLS 注入為最後防線。
- **通道限制(D16)**:1.0 僅 `fcm` / `in_app` 兩通道,**無 Email**;任何通道欄位出現其他值一律拒絕。
- **錯誤處理約定**:統一以 Connect code 表述(unauthenticated / permission_denied / not_found / failed_precondition / invalid_argument / already_exists);樂觀鎖衝突一律 `failed_precondition`。
- **路由總表(D16 / D23,規格 §10.1)**:

| 觸發事件 | 接收者 | 通道 |
|---|---|---|
| 業務下單(訂單建立) | 該客戶全部**子帳號**(主帳號排除) | FCM + 站內 |
| 客戶自行下單 | 不另行通知 | — |
| 後台新增客戶專屬商品 | 主責業務(`customers.default_sales_rep_id` → 部門 `dept_admin`;業務檢核待定) | FCM + 站內 |
| 退貨審核結果 | 發起退貨的客戶帳號 | FCM + 站內 |
| 促銷推播(Phase 7 Task 7.4) | 依 `promo_tags` 分類訂閱之客戶群**子帳號** | FCM + 站內 |

- **語言**:接收者之通知以其帳號 `locale`(無則部門/公司預設)選取範本;範本缺對應語系時退回同 `code` 之預設語系範本。

---

## Task 4.3:通知系統 API

### 子功能 4.3.1: notification_templates / notifications / user_devices schema

- **目標**: 建立通知系統三張表的 Ent schema 與 migration,作為後續所有通知子功能的資料基礎。
- **檔案**:
  - Create `backend/internal/domain/notifications/schema/notification_template.go`
  - Create `backend/internal/domain/notifications/schema/notification.go`
  - Create `backend/internal/domain/notifications/schema/user_device.go`
  - Create `backend/migrations/`(對應 migration 檔)
- **介面**: Ent 實體(無對外 RPC):
  - `NotificationTemplate`:`company_id`、`department_id`、`code`(範本代號,如 `order_created`)、`name`、`channel`(`fcm` / `in_app`)、`subject`、`body`(含 `{{變數}}` 佔位)、`locale`、`is_active`、`deleted_at`
  - `Notification`:`company_id`、`department_id`、`user_id`(接收者)、`template_id`(可空,手動通知無範本)、`channel`、`title`、`content`、`payload`(JSON,攜帶導頁資訊如訂單 ID)、`status`(`pending` / `sent` / `failed` / `read`)、`failure_reason`、`sent_at`、`read_at`(無 `deleted_at`,不可刪除)
  - `UserDevice`:`user_id`、`company_id`、`platform`(`ios` / `android` / `web`)、`fcm_token`、`device_name`、`last_seen_at`、`deleted_at`
- **實作邏輯**:
  1. `notification_templates` 以部分唯一索引 `(company_id, department_id, code, channel, locale) WHERE deleted_at IS NULL` 保證同範圍同通道同語系只有一個啟用中範本代號。
  2. `user_devices` 以部分唯一索引 `(fcm_token) WHERE deleted_at IS NULL` 保證同一 token 全表唯一(一個裝置 token 同時只屬一個使用者;換帳登入由 4.3.4 重新歸屬)。
  3. `notifications` 建立 `(user_id, status, created_at)` 索引供通知中心查詢未讀與分頁;不建 `deleted_at`,RLS 原則照常注入 `company_id` / `department_id`。
  4. `notifications.template_id` 外鍵允許 NULL(人工/系統直發通知不經範本);`payload` 為 JSON 物件,結構由觸發方定義,通知系統不解析。
  5. 三表皆納入 RLS policy:`user_devices` 以 `company_id` 隔離;`notifications` 以 `company_id` + `department_id`;`notification_templates` 以 `company_id` + `department_id`。
- **錯誤處理**: migration 層失敗為啟動期 fail-fast(不對外暴露 Connect code);唯一索引衝突於上層轉 `already_exists`。
- **驗收**:
  - [ ] migration 可於空庫與既有庫重複套用無誤。
  - [ ] 三表欄位、索引、RLS policy 與上述定義一致。
  - [ ] `notifications` 無 `deleted_at` 欄位,刪除語句在 schema 層無對應欄位可寫。

### 子功能 4.3.2: 範本渲染(變數替換)

- **目標**: 提供範本選取與 `{{變數}}` 渲染的內部函式,供所有觸發方與 `NotificationTemplateService` 預覽使用。相依: 4.3.1
- **檔案**:
  - Create `backend/internal/domain/notifications/render.go`
- **介面**: 內部函式(非 RPC):輸入 `(company_id, department_id, code, channel, locale, variables map)`;輸出 `(title, content)` 字串;`NotificationTemplateService` 另提供 `Preview` RPC(輸入範本 ID + 變數 map,回傳渲染結果)。
- **實作邏輯**:
  1. 選取範本:依 `(company_id, department_id, code, channel, locale)` 且 `is_active = true`、未軟刪除查詢;查無時依序退回:同 code 同 channel 之部門預設語系 → 同 code 同 channel 之公司層範本(`department_id` 為公司預設) → 視為失敗。
  2. 渲染:對 `subject` 與 `body` 逐一替換 `{{變數名}}`;變數名僅允許英數與底線,比對採全字吻合。
  3. 未定義變數:出現在範本中但 `variables` 未提供的佔位符,保留原文不替換,並將缺漏變數名清單寫入警告日誌(不阻斷發送,避免範本小改即癱瘓通知)。
  4. 注入防護:渲染結果不再二次解析(輸出即純文字);FCM 通道 title/body 依 FCM 訊息長度上限截斷,站內通道不截斷但由前端截顯。
  5. 本函式純計算、不寫庫;寫庫由 4.4.2 / 4.4.3 負責。
- **錯誤處理**: 範本不存在或已停用 → `not_found`(觸發方捕獲後降級為僅建 `template_id` 為空的通知或記 `failed`,見 4.4.5);`code` / `channel` 為空或不合法 → `invalid_argument`;Preview 呼叫者無範本讀取權限 → `permission_denied`。
- **驗收**:
  - [ ] 提供完整變數時,`subject` / `body` 全部佔位符被正確替換。
  - [ ] 缺漏變數保留原文且有警告日誌;多餘變數被忽略。
  - [ ] 語系退回順序(指定語系 → 預設語系 → 公司層)行為正確;全無範本時回 `not_found`。

### 子功能 4.3.3: 通知記錄 + 已讀 API(pending → sent → read)

- **目標**: 提供通知中心的查詢與已讀 RPC,讓 Web / App 使用者取得自己的通知並標記已讀。相依: 4.3.1
- **檔案**:
  - Create `backend/internal/domain/notifications/service.go`
  - Create `backend/internal/domain/notifications/repo.go`
  - Create `proto/notification/v1/notification.proto`
- **介面**: Connect-RPC `NotificationService`:
  - `List(ListNotificationsRequest{ page, page_size, unread_only }) returns (ListNotificationsResponse{ notifications[], total, unread_count })` — 僅回傳當前使用者自己的通知。
  - `MarkRead(MarkReadRequest{ notification_ids[] }) returns (MarkReadResponse{ marked_count })`
  - `UnreadCount(UnreadCountRequest{}) returns (UnreadCountResponse{ count })` — 供通知鈴角標輪詢。
  - proto 訊息含 `id / channel / title / content / payload / status / sent_at / read_at`。
- **實作邏輯**:
  1. List:以當前使用者 `user_id` 過濾(RLS 已限縮公司/部門,此處再限定本人),依 `created_at` 倒序分頁;`unread_only = true` 時加 `status IN ('pending','sent')` 過濾;同查詢或平行查詢回傳 `unread_count`。
  2. MarkRead:僅接受當前使用者本人的通知 ID;狀態機僅允許 `sent` → `read`(以及 `pending` → `read`,站內通道尚未推播即被讀的情況);已 `read` 者冪等略過;`failed` 不允許轉 `read`(發送失敗的通知不該被當成已讀訊息)。
  3. MarkRead 批次於單一交易更新 `status = 'read'`、`read_at = now()`;回傳實際異動筆數。
  4. 狀態流轉總表:`pending`(建檔)→ `sent`(通道發送成功,4.4.1/4.4.2)→ `read`(使用者已讀);失敗分支 `pending` → `failed`(4.4.5,終態);`read` / `failed` 皆終態,不可再轉。
  5. 通知記錄不可刪除(規格 §5.4):不提供 Delete RPC,亦不軟刪除。
- **錯誤處理**: 未登入 → `unauthenticated`;對他人通知 ID 呼叫 MarkRead → 視同不存在,`not_found`(不洩漏存在性);主帳號呼叫(主帳號無業務身分,D22)→ `permission_denied`;空 ID 陣列或超過單批上限 → `invalid_argument`。
- **驗收**:
  - [ ] List 只回傳本人通知,跨使用者資料不出現;分頁與 `unread_count` 正確。
  - [ ] MarkRead 對 `sent` 通知轉 `read` 並寫 `read_at`;重複呼叫冪等。
  - [ ] `failed` 通知 MarkRead 回 `failed_precondition` 或略過並於回應反映未異動。
  - [ ] 狀態機無 `read`/`failed` 之後的非法轉移路徑。

### 子功能 4.3.4: DeviceService 註冊 / 註銷 + FCM 失效 token 自動刪除

- **目標**: App 登入後註冊裝置 FCM token、登出/換帳時註銷;FCM 回報 token 失效時自動清除,避免無效推播(D16)。相依: 4.3.1
- **檔案**:
  - Create `backend/internal/domain/notifications/device_service.go`
  - Create `proto/notification/v1/device.proto`
- **介面**: Connect-RPC `DeviceService`:
  - `Register(RegisterDeviceRequest{ platform, fcm_token, device_name }) returns (RegisterDeviceResponse{ device_id })`
  - `Unregister(UnregisterDeviceRequest{ fcm_token }) returns (UnregisterDeviceResponse{})`
  - 內部介面 `PurgeInvalidTokens(tokens[] string)`:供 4.4.1 FCM 發送迴路呼叫。
- **實作邏輯**:
  1. Register:以 `fcm_token` 查 `user_devices`(含軟刪除);已存在且屬當前使用者 → 更新 `platform` / `device_name` / `last_seen_at` 並復原(清 `deleted_at`),冪等回傳原 ID;已存在但屬**其他使用者**(換帳登入同一裝置)→ 原記錄軟刪除,為當前使用者新建一筆,保證 token 同一時刻只屬一人。
  2. Register 寫入與稽核日誌(裝置註冊)同一 DB 交易(D18)。
  3. Unregister:依 `fcm_token` + 當前使用者軟刪除對應記錄;查無記錄時冪等成功(不報錯),避免登出流程被裝置狀態卡住。
  4. PurgeInvalidTokens:FCM 回應錯誤碼為 `unregistered` / `invalid-registration-token` 時,將該批 token 對應的 `user_devices` 全部軟刪除(不限當前使用者,屬系統行為,以服務帳號身分執行、RLS 以對應公司範圍注入);每筆清除寫稽核(原因 = FCM 回報失效)。
  5. `last_seen_at` 僅在 Register 與成功發送(4.4.1)時更新,供後續清理久未使用裝置的排程參考(1.0 不建清理排程)。
- **錯誤處理**: 未登入 → `unauthenticated`;`fcm_token` 為空或超長 → `invalid_argument`;主帳號註冊裝置(主帳號不登入 App 業務,D22)→ `permission_denied`;Unregister 指定 token 不屬本人 → `not_found`。
- **驗收**:
  - [ ] App 可註冊 token 並重複註冊冪等;換帳登入後 token 歸屬轉移且舊使用者查不到。
  - [ ] 登出註銷後該 token 不再收到推播;重複註銷不報錯。
  - [ ] 模擬 FCM 回報 unregistered 後,對應 `user_devices` 被軟刪除且有稽核記錄。

### 子功能 4.3.5: promo_tags 與套用欄位資料層(商品 / 專屬商品 / 客戶 promo_tag_ids)

- **目標**: 建立促銷分類標籤的資料層(schema 與欄位),供 Phase 7 Task 7.4 的 CRUD 與分群推播使用;**本子功能僅建資料層,不建 CRUD RPC 與推播 UI**。相依: 4.3.1
- **檔案**:
  - Create `backend/internal/domain/promotions/schema/promo_tag.go`
  - Update `backend/internal/domain/products/schema/`(商品、專屬商品實體加 `promo_tag_ids` 欄位)
  - Update `backend/internal/domain/customers/schema/`(客戶實體加 `promo_tag_ids` 欄位)
  - Create `backend/migrations/`(對應 migration 檔)
- **介面**: Ent 實體(無對外 RPC):
  - `PromoTag`:`company_id`、`department_id`(部門級標籤,D24)、`code`、`name`、`is_active`、`deleted_at`;部分唯一索引 `(company_id, department_id, code) WHERE deleted_at IS NULL`。
  - 欄位層:`products.promo_tag_ids`、`customer_products.promo_tag_ids`、`customers.promo_tag_ids` 皆為 JSON 陣列(存 `promo_tags.id`),預設空陣列。
- **實作邏輯**:
  1. `promo_tags` 為部門級:RLS 注入 `company_id` + `department_id`,跨部門不可見。
  2. 三處 `promo_tag_ids` 僅存 ID 陣列,不建外鍵(JSON 陣列無法外鍵);**有效性檢核責任在寫入方**(Phase 7 Task 7.4 的 CRUD 與客戶套用 API):寫入時需驗證陣列中每個 ID 存在於同公司同部門且未軟刪除的 `promo_tags`。
  3. `promo_tags` 軟刪除時**不反向清理**各宿主的 `promo_tag_ids`(殘留 ID 在推播選群時自然失效);於文件註明此行為,避免 Phase 7 實作者誤以為有級聯。
  4. **跨界註明**:本子功能歸屬原計畫 Task 4.3 Step 6,但 `promo_tags` 的 CRUD RPC、客戶套用(訂閱)API、依分類選群推播皆屬 **Phase 7 Task 7.4**(`backend/internal/domain/promotions/*` 其餘部分與前端頁面);此處僅交付 schema 與欄位,Phase 7 直接複用 `backend/internal/domain/promotions/` 目錄續建。
- **錯誤處理**: 資料層無 RPC;後續寫入方驗證失敗時 → `invalid_argument`(引用不存在的標籤 ID)、`permission_denied`(跨部門標籤)。
- **驗收**:
  - [ ] `promo_tags` schema、部分唯一索引、RLS policy 就緒。
  - [ ] 商品 / 專屬商品 / 客戶三表新增 `promo_tag_ids` JSON 欄位,既有資料遷移後預設空陣列。
  - [ ] 文件明確標示 CRUD 與選群推播屬 Task 7.4,本 Phase 不出現對應 RPC。

---

## Task 4.4:FCM / 站內通知發送

### 子功能 4.4.1: FCM client 整合

- **目標**: 建立對 Firebase Cloud Messaging 的發送通道封裝,供所有觸發方統一呼叫。相依: 4.3.1、4.3.4
- **檔案**:
  - Create `backend/internal/notification/fcm.go`
- **介面**: 內部套件函式(非 RPC):`Send(user_ids[] , title, content, payload) → per-user 結果(成功 / failed + 原因 / 失效 token 清單)`;初始化時由設定檔注入 service account 憑證路徑,缺憑證於啟動期 fail-fast(開發環境可設 `FCM_DISABLED=true` 降級為僅記日誌,文件註明此開關僅供開發)。
- **實作邏輯**:
  1. 以 FCM Admin SDK 初始化;批次發送採 multicast(一次最多 500 token),內部依 `user_ids` 查出所有未軟刪除 `user_devices` 的 token。
  2. **時序邊界**:FCM 呼叫屬外部 I/O,**不得包在業務 DB 交易內**;觸發方先在交易內建 `notifications`(status = `pending`)並提交,提交成功後才呼叫本函式,再依結果更新狀態(同事務規則僅涵蓋建檔,見 §0 共通規則)。
  3. 結果分類:成功 → 對應通知標 `sent` + `sent_at`;錯誤碼 `unregistered` / `invalid-registration-token` → 呼叫 4.3.4 `PurgeInvalidTokens` 清除,且該使用者若仍有其他有效裝置則不影響其 `sent` 判定(任一裝置成功即 `sent`);其他錯誤(配額、內部錯誤)→ 交 4.4.5 標 `failed`。
  4. 單一使用者多裝置:對其全部 token 發送;全部失效才視為發送失敗。
  5. 逾時與連線:設定單次呼叫逾時(秒級),逾時歸類為失敗,**不重試**(D16)。
- **錯誤處理**: 內部套件不直接回 Connect code;上層觸發方於「通知建檔失敗」時回 `internal` 並使業務交易回滾(D18,業務與通知建檔同交易);FCM 呼叫本身失敗不回滾業務,只反映於通知狀態。
- **驗收**:
  - [ ] 以 FCM 模擬/測試專案驗證 multicast 發送與結果分類(成功 / 失效 token / 其他失敗)。
  - [ ] 失效 token 觸發 `user_devices` 軟刪除;多裝置使用者任一成功即 `sent`。
  - [ ] FCM 逾時或 5xx 時業務交易不回滾,通知標 `failed`。

### 子功能 4.4.2: 站內通知發送(通知中心 API 資料流)

- **目標**: 建立站內通道的發送路徑——寫入 `notifications` 記錄供通知中心(4.3.3)讀取。相依: 4.3.1、4.3.2
- **檔案**:
  - Create `backend/internal/notification/inapp.go`
- **介面**: 內部函式 `SendInApp(user_ids[], template_code, variables, payload) → 每使用者一筆 notifications`;對外讀取面即 4.3.3 的 `NotificationService`(通知中心 API),本通道不另建寫入 RPC。
- **實作邏輯**:
  1. 站內發送 = 對每位接收者以 4.3.2 渲染 `in_app` 通道範本,建一筆 `notifications`(channel = `in_app`)。
  2. 站內通道無外部投遞環節:建檔成功即視為送達,交易提交後立即批次更新 `status = 'sent'`、`sent_at`(若採「建檔即 sent」策略,則建檔時直接寫 `sent`,但狀態機文件仍以 pending → sent 描述 FCM 雙通道一致性;實作擇一並於程式註解標明)。
  3. 與 FCM 雙通道觸發時:同一位接收者產生**兩筆**通知記錄(channel 各一),已讀互不影響;通知中心預設只顯示 `in_app` 通道,FCM 紀錄供追蹤,前端查詢帶 channel 過濾。
  4. 接收者無有效裝置時,FCM 通道記錄直接 `failed`(原因 = 無裝置),站內通道仍正常送達。
- **錯誤處理**: 範本缺失 → 依 4.3.2 降級規則;接收者 ID 無效或已停用 → 該筆略過並記日誌,不影響其他接收者;交易內建檔失敗 → 隨業務交易回滾(上層感知為 `internal`)。
- **驗收**:
  - [ ] 觸發後接收者於通知中心(List)看到站內通知,`unread_count` 增加。
  - [ ] 雙通道觸發產生 fcm / in_app 各一筆,通知中心僅顯示 in_app。
  - [ ] 停用帳號不產生通知記錄,其餘接收者不受影響。

### 子功能 4.4.3: 下單觸發路由(業務下單推子帳號;客戶自行下單不通知)

- **目標**: 在訂單建立流程掛上通知觸發,落實 D23 路由:業務下單推該客戶全部子帳號、主帳號排除;客戶自行下單不通知。相依: 4.4.1、4.4.2;觸發點相依 `05-sales-orders.md`(Task 4.2 下單流程)
- **檔案**:
  - Create `backend/internal/notification/triggers/order_created.go`
  - Update `backend/internal/domain/orders/`(訂單建立流程注入通知觸發)
- **介面**: 內部觸發函式 `OnOrderCreated(order, actor)`;使用範本 `code = order_created`,變數至少含 `{{customer_name}}`、`{{order_no}}`、`{{item_count}}`;payload 帶訂單 ID 供 App 導頁。
- **實作邏輯**:
  1. 身分判別:檢查下單操作者角色——**業務/內部員工**下單才觸發;**客戶子帳號自行下單**直接結束,不發任何通知(避免客戶收到自己操作的通知)。
  2. 接收者解析:取訂單客戶的全部帳號,過濾 `is_primary = false`(主帳號為管理用途,D22)且帳號未停用;結果為空(客戶無子帳號)時記日誌並正常結束,不視為錯誤。
  3. 於訂單建立的**同一 DB 交易**內為每位子帳號建 `notifications`(fcm / in_app 各一筆,status = `pending`);交易提交後依 4.4.1 / 4.4.2 完成發送與狀態更新。
  4. 稽核:通知觸發本身不另寫稽核(下單稽核已涵蓋,D18);僅裝置清除(4.3.4)與失敗標記(4.4.5)留軌跡。
  5. 冪等:通知觸發以訂單 ID 為鍵,訂單建立交易重試(樂觀鎖取號衝突重試,見 05 文件)時不會產生重複通知——通知建檔與訂單同交易,訂單失敗即一併回滾。
- **錯誤處理**: 客戶資料異常(無帳號)→ 記日誌不阻斷下單;通知建檔 SQL 失敗 → 整個下單交易回滾,呼叫端收 `internal`;發送階段失敗 → 僅標 `failed`(4.4.5),不影響訂單。
- **驗收**:
  - [ ] 業務代客下單後,該客戶所有子帳號收到 FCM 推播與站內通知,主帳號無任何通知。
  - [ ] 客戶子帳號自行下單後,該客戶任何帳號(含其他子帳號)皆無通知。
  - [ ] 訂單建立失敗(如取號衝突耗盡重試)時無殘留 `pending` 通知。

### 子功能 4.4.4: 後台新增專屬商品推主責業務(default_sales_rep_id → dept_admin,檢核待定)

- **目標**: 後台/Web 為客戶新增專屬商品(`customer_products`)時,推播主責業務;主責業務未設定時退回部門 `dept_admin`;是否需要業務檢核為待定項,實作保留開關。相依: 4.4.1、4.4.2;觸發點相依 `04-master-data.md`(Task 3.x 專屬商品建立)
- **檔案**:
  - Create `backend/internal/notification/triggers/customer_product_created.go`
  - Update `backend/internal/domain/products/`(專屬商品建立流程注入通知觸發)
- **介面**: 內部觸發函式 `OnCustomerProductCreated(customer_product, actor)`;範本 `code = customer_product_created`,變數含 `{{customer_name}}`、`{{product_name}}`。
- **實作邏輯**:
  1. 觸發條件:僅限後台 / Web 建立專屬商品(App 端無此操作);批次匯入建立時逐筆觸發或彙整一則,實作採**逐筆觸發**(1.0 量級可接受,避免彙整視窗複雜度)。
  2. 接收者解析:取該客戶 `default_sales_rep_id`;為 NULL 或該業務已停用/離職 → 退回同部門具 `dept_admin` 角色的所有啟用中使用者;兩者皆無 → 記警告日誌,不發送。
  3. **檢核待定項**(D16/D23「是否需要業務檢核,待定」):實作以設定旗標 `NOTIFY_CUSTOMER_PRODUCT_REVIEW` 預留——開啟時通知文案改為「待檢核」語意、關閉時為「已新增」語意;旗標預設關閉(即不檢核,新增即通知),待定決議後僅需翻旗標與調範本,不改程式結構。
  4. 交易邊界與冪等同 4.4.3:通知建檔與專屬商品建立同交易,發送在提交後。
- **錯誤處理**: 無接收者(未設主責業務且部門無 dept_admin)→ 記警告日誌,業務操作照常成功;範本缺失 → 4.3.2 降級規則;通知建檔失敗 → 專屬商品建立交易回滾,回 `internal`。
- **驗收**:
  - [ ] 後台新增專屬商品後,主責業務收到 FCM + 站內通知。
  - [ ] 客戶未設 `default_sales_rep_id` 時,dept_admin 收到通知;主責業務已停用時亦退回 dept_admin。
  - [ ] 檢核旗標切換時通知文案語意隨之改變,無需改程式。

### 子功能 4.4.5: 發送失敗標記 failed 不重試(記錄原因)

- **目標**: 統一所有通道的失敗收斂:不重試、`notifications.status = failed`、記錄原因,供人工追蹤(D16)。相依: 4.4.1、4.4.2
- **檔案**:
  - Update `backend/internal/notification/fcm.go`(失敗分類接入)
  - Create `backend/internal/notification/failmark.go`(共用失敗標記函式)
- **介面**: 內部函式 `MarkFailed(notification_ids[], reason)`;`reason` 為列舉字串(fcm 錯誤碼分類:逾時 / 配額 / 無裝置 / 範本缺失 / 其他 + 原始訊息摘要),寫入 `notifications.failure_reason`。
- **實作邏輯**:
  1. 任何通道發送失敗(4.4.1 分類為「其他錯誤」、無有效裝置、範本缺失降級失敗)皆呼叫 `MarkFailed`:單筆交易更新 `status = 'failed'` + `failure_reason`,`sent_at` 保持 NULL。
  2. **不重試**:1.0 無重試佇列、無定時補發;`failed` 為終態(見 4.3.3 狀態機)。文件與程式註解明示重試佇列列後續版本評估,不得在本 Phase 偷加 retry。
  3. 部分失敗:同一觸發多接收者時逐接收者獨立標記,一人失敗不影響他人 `sent`。
  4. 可觀測性:每次 `MarkFailed` 寫警告級日誌(含通知 ID、接收者、reason),並計入 `/metrics` 業務指標(通知失敗計數,依 channel / reason 分維度,D19);管理端追查以日誌 + DB 查詢為主,1.0 不建失敗通知管理頁。
  5. 併發防護:`MarkFailed` 僅在 `status = 'pending'` 時生效(條件更新),若記錄已被標 `sent`(雙通道競態)則略過,避免覆蓋成功狀態。
- **錯誤處理**: `MarkFailed` 自身 SQL 失敗 → 記錯誤日誌(通知狀態停留 `pending`,由人工依日誌追查,不再升級);不對外暴露 Connect code。
- **驗收**:
  - [ ] 模擬 FCM 5xx / 逾時後,對應通知為 `failed` 且 `failure_reason` 可讀。
  - [ ] 失敗後無任何自動重發行為;`failed` 狀態不可再被 MarkRead 或重標。
  - [ ] 多接收者情境下部分失敗僅影響失敗者;失敗計數出現在 `/metrics`。

---

## 整合測試重點

依 D21(關鍵路徑整合測試)與本 Task 驗收基準,需以 dockertest + testcontainers(或 FCM 模擬層)覆蓋:

1. **端到端路由(4.4.3)**:業務代客下單 → 客戶全部子帳號收到 in_app 通知(通知中心可查到)且主帳號無;客戶子帳號自行下單 → 全客戶帳號皆無通知。此為 Phase 4 驗收(Task 4.8)直接依據。
2. **專屬商品路由(4.4.4)**:有/無 `default_sales_rep_id`、業務停用三種組合下接收者正確(主責業務 / dept_admin / 僅日誌)。
3. **裝置生命週期(4.3.4 + 4.4.1)**:註冊 → 換帳歸屬轉移 → 註銷;FCM 回報 unregistered → `user_devices` 軟刪除 + 稽核,且再次下單不再對該 token 發送。
4. **狀態機(4.3.3 + 4.4.5)**:pending → sent → read 正向流轉;pending → failed 終態不可再轉;failed 不可 MarkRead;MarkRead 冪等;雙通道競態下 `sent` 不被 `failed` 覆蓋。
5. **交易一致性(D18)**:強迫通知建檔失敗 → 下單/新增專屬商品整體回滾、無殘留通知;強迫 FCM 失敗 → 業務仍成功、通知為 `failed`。
6. **RLS 隔離(D3)**:A 公司使用者 List/MarkRead 永遠碰不到 B 公司通知;跨公司 token 清除僅影響對應公司範圍。
7. **範本退回(4.3.2)**:語系退回鏈與缺漏變數保留原文的行為在整合環境可重現。
