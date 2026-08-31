# 08 — 派車與串流看板(Backend 細部實作計畫)

---
對應原計畫 Task:**5.1 派車 API**、**5.2 Connect 串流即時派車看板(僅後端部分:proto / 串流 handler / pub-sub / heartbeat;前端重連與降級輪詢僅註記,實作屬原計畫 Task 5.6)**
對應規格書 v1.0.34 章節:**7.1 派車看板**、**6.1 訂單狀態機(取消派車原因與回退)**;需求層見 `docs/superpowers/specs/1.0-requirements/dispatch/spec.md`
相關決策:D3(租戶三層分工)、D5(認證雙軌)、D10(軟刪除)、D13(狀態機與回退)、D14(派車批次確認 + 看板樂觀鎖 + Connect 串流)、D16(通知通道)、D18(稽核同事務)、D21(整合測試)
相依文件:`01-auth.md`(auth interceptor / data_scope 注入)、`04-master-data.md`(車次 `routes` 主檔)、`05-sales-orders.md`(4.1.3 訂單狀態機、`sales_order_events`)、`07-notifications.md`(4.3 範本、4.4 FCM/站內發送)、`09-printing.md`(`print_logs` 正式列印記錄)
---

## 共通規則(本文件適用,各子功能不重複)

1. **交易與稽核(D18)**:狀態異動 + `sales_order_events`、取消派車 + `audit_logs`,皆同一 DB 交易,同成功同失敗;各子功能「實作邏輯」欄標出交易邊界。
2. **樂觀鎖**:所有看板拖放/指派異動比對 `sales_orders.version`,衝突一律回 `failed_precondition`,由前端重新整理看板,後端不自動重試業務語意衝突(D14)。
3. **多租戶(D3)**:派車操作僅能作用於操作者 `data_scope` 範圍內的部門訂單;repository 查詢條件 + RLS(`app.current_company_id` / `app.current_department_id` / `app.current_data_scope`)為最後防線。跨部門看板事件不得外洩。
4. **錯誤處理約定**:統一以 Connect code 表述(unauthenticated / permission_denied / not_found / failed_precondition / invalid_argument / already_exists),見 `00-index.md` §3.4。
5. **事件推播語意(D14)**:`BoardEvent` 僅作「看板查詢失效」提示,前端收到後全量重查,後端不保證事件順序、不補發斷線期間事件;狀態正確性以重查結果為準。

---

## Task 5.1:派車 API

對應原計畫 `### Task 5.1`。核心檔案 `backend/internal/domain/salesorders/dispatch.go`;RPC 定義於 `backend/proto/v1/dispatch.proto` 的 `DispatchService`。

### 子功能 5.1.1:AssignRoute(指派車次與配送順序)

- **目標**:拖放提交時更新訂單的 `route_id` / `delivery_sequence`,以樂觀鎖防止併發覆蓋。
- **檔案**:
  - Create/Update: `backend/proto/v1/dispatch.proto`(AssignRoute RPC 與訊息)
  - Create/Update: `backend/internal/domain/salesorders/dispatch.go`(usecase + repository)
- **介面**:Connect-RPC `DispatchService.AssignRoute`
  - Request:`sales_order_id`、`route_id`(可為空值語意,代表拖回「未指派」)、`delivery_sequence`、`version`(客戶端讀取時的 `sales_orders.version`)、`expected_delivery_date`(看板當前日期,用於同車次順位重排範圍)。
  - Response:更新後的 `sales_order_id`、`route_id`、`delivery_sequence`、新 `version`。
- **實作邏輯**:
  1. 驗證輸入:`sales_order_id` 必填;`route_id` 有值時 `delivery_sequence` 必須 ≥ 1;`version` 必填(缺漏視同無法比對)。
  2. 讀取訂單(帶 RLS 注入的查詢):不存在或已軟刪除 → `not_found`;狀態非 `pending` → `failed_precondition`(已派車/完成/取消/作廢的訂單不可再拖放)。
  3. `route_id` 有值時,驗證該車次存在、未軟刪除、且與訂單同 `department_id`;跨部門車次 → `permission_denied`(不暴露存在性細節時以 `not_found` 處理,二擇一後全文一致,本文件採 `permission_denied` 僅限功能權限,資料不存在一律 `not_found`)。
  4. 開啟交易(交易邊界開始):
     - 以「`sales_order_id` + `version` 相符」為條件更新 `route_id` / `delivery_sequence` 並 `version + 1`;影響列數為 0 → 樂觀鎖衝突,整個交易 rollback,回 `failed_precondition`。
     - 同車次順位重排:目標車次(同 `expected_delivery_date`、同部門、未軟刪除、狀態 `pending`)中 `delivery_sequence` ≥ 新順位的其他訂單依序後移;若訂單自原車次移出,原車次空出的順位不回填(順位允許空洞,由看板排序顯示吸收,避免一次拖放連鎖更新大量列)。
     - 寫入一筆 `sales_order_events`(event_type = route_assign,含新舊 `route_id` / `delivery_sequence`)。
  5. 提交交易;提交成功後發佈 `BoardEvent` 至 Valkey pub/sub(見 5.2.2),並於 Response 帶回新 `version`。
- **錯誤處理**:
  - `unauthenticated`:未登入或 session/token 失效。
  - `permission_denied`:Casbin 無 dispatch 功能權限,或 `data_scope` 不含該訂單部門。
  - `not_found`:訂單或車次不存在、已軟刪除。
  - `failed_precondition`:狀態非 `pending`;`version` 樂觀鎖衝突(前端提示「資料已被他人變更」並重整理看板)。
  - `invalid_argument`:參數缺漏或順位 < 1。
- **驗收**:
  - [ ] 拖放至新車次/新順位成功,`version` 遞增,目標車次順位 ≥ 新順位者依序後移。
  - [ ] 以過期 `version` 提交回 `failed_precondition`,資料不變。
  - [ ] 拖回「未指派」(`route_id` 清空)可行,原車次順位空洞不影響顯示排序。
  - [ ] 非 `pending` 訂單指派被拒;跨部門車次不可指派。
  - [ ] 提交成功後同部門看板連線收到 `BoardEvent`(配合 5.2.2 驗證)。

### 子功能 5.1.2:車次批次 Confirm(派車確認)

- **目標**:以車次為批次單位,一次確認該車次當日全部待派訂單,逐筆轉 `processing`。`相依: 4.1.3`(訂單狀態機)
- **檔案**:
  - Create/Update: `backend/proto/v1/dispatch.proto`(Confirm RPC 與訊息)
  - Create/Update: `backend/internal/domain/salesorders/dispatch.go`
- **介面**:Connect-RPC `DispatchService.Confirm`
  - Request:`route_id`、`expected_delivery_date`。
  - Response:逐筆結果清單(每筆含 `sales_order_id`、成功/失敗、失敗原因),以及成功筆數彙總;已成功與失敗訂單在回應中明確區分。
- **實作邏輯**:
  1. 驗證輸入:`route_id`、`expected_delivery_date` 必填;車次存在、未軟刪除、屬操作者資料範圍內部門,否則 `not_found`。
  2. 撈出批次候選:同車次、同 `expected_delivery_date`、同部門、未軟刪除、狀態 = `pending` 且 `route_id` 不為空的訂單;**未指派車次的訂單不納入任何批次**。
  3. 逐筆處理,**每筆訂單各自一個交易**(交易邊界 = 單筆訂單;批次不採全有全無,單筆失敗不影響他筆,對應規格「已成功與失敗明確區分」):
     - 以「當前狀態仍為 `pending`」為條件更新:狀態 → `processing`、寫入 `dispatched_at`(批次內各筆共用同一時間值)、`dispatched_by` = 操作者、`version + 1`。
     - 條件不符(已被取消/已被他人派車)→ 該筆記為失敗並帶原因,不更新任何欄位、不重複寫入 `dispatched_at`,繼續下一筆。
     - 同交易寫入一筆 `sales_order_events`(event_type = dispatch)。
  4. 全部處理完畢後,對每筆成功訂單觸發派車通知(見 5.1.4),並對每筆成功訂單發佈 `BoardEvent`(見 5.2.2)。
  5. 回傳逐筆結果;批次內 0 筆候選訂單 → `failed_precondition`(該車次當日無待派訂單)。
- **錯誤處理**:
  - `unauthenticated` / `permission_denied`:同 5.1.1;`staff` 以上可執行確認(看板頁適用角色 dept_admin / staff)。
  - `not_found`:車次不存在或不在資料範圍。
  - `failed_precondition`:當日該車次無 `pending` 待派訂單;單筆狀態不符記於該筆結果而非整批失敗。
  - `invalid_argument`:缺 `route_id` / 日期。
- **驗收**:
  - [ ] 車次當日 5 筆 `pending` 訂單一次確認後全部轉 `processing`,各筆 `dispatched_at` 相同、`dispatched_by` 為操作者。
  - [ ] 每筆成功訂單各寫一筆 `sales_order_events`(event_type = dispatch)。
  - [ ] 批次中已被取消的訂單回報失敗原因、欄位不變,其餘訂單照常成功。
  - [ ] 未指派車次的 `pending` 訂單不被任何批次納入。
  - [ ] 每筆成功訂單觸發派車通知並發佈看板事件。

### 子功能 5.1.3:CancelDispatch(取消派車)

- **目標**:dept_admin 以上將 `processing` 訂單退回 `pending`,清除派車資訊但保留看板位置,並對已列印車次回傳重印警告。
- **檔案**:
  - Create/Update: `backend/proto/v1/dispatch.proto`(CancelDispatch RPC 與訊息)
  - Create/Update: `backend/internal/domain/salesorders/dispatch.go`
- **介面**:Connect-RPC `DispatchService.CancelDispatch`
  - Request:`sales_order_id`、`reason`(必填)、`acknowledge_reprint`(布林;該車次已正式列印時,前端要求使用者確認後重送並帶 true)。
  - Response:更新後訂單狀態、`route_id` / `delivery_sequence`(保留值)、`reprint_warning`(布林,該車次當日存在正式列印記錄)。
- **實作邏輯**:
  1. 權限:角色須 dept_admin 以上(Casbin resource = dispatch、action = cancel);`staff` 或更低 → `permission_denied`。
  2. 驗證輸入:`reason` 空白 → `invalid_argument`(原因必填,訂單維持不變)。
  3. 讀取訂單:不存在 → `not_found`;狀態非 `processing` → `failed_precondition`(僅處理中可取消派車)。
  4. 檢查該車次當日是否存在正式列印記錄(`print_logs`,排除預覽 `print_previews`):
     - 存在且 `acknowledge_reprint` 不為 true → 不執行取消,Response 帶 `reprint_warning = true`,由前端提示「該車次已列印,取消後需重新列印相關單據」並要求確認。
     - 存在且已確認,或不存在 → 繼續。
  5. 開啟交易(交易邊界開始,D18):
     - 清除 `dispatched_at` / `dispatched_by`,狀態退回 `pending`,`version + 1`;**保留** `route_id` / `delivery_sequence`(訂單停留看板原車次原順位)。
     - 寫入一筆 `sales_order_events`(event_type = dispatch_cancel,含原因)。
     - 寫入一筆 `audit_logs`(action = dispatch_cancel,含操作者、原因、訂單快照)。
  6. 提交交易;提交成功後發佈 `BoardEvent`(見 5.2.2),Response 帶回保留的看板位置與 `reprint_warning` 最終狀態。
- **錯誤處理**:
  - `unauthenticated`:未登入。
  - `permission_denied`:角色低於 dept_admin,或訂單不在資料範圍。
  - `not_found`:訂單不存在或已軟刪除。
  - `failed_precondition`:狀態非 `processing`。
  - `invalid_argument`:`reason` 為空。
- **驗收**:
  - [ ] 取消後狀態退回 `pending`,`dispatched_at` / `dispatched_by` 清空,`route_id` / `delivery_sequence` 保留原值。
  - [ ] 同事務寫入 `sales_order_events`(dispatch_cancel,含原因)與 `audit_logs`(dispatch_cancel)。
  - [ ] 未填原因 → `invalid_argument`,訂單完全不變。
  - [ ] 車次已有正式列印記錄且未確認 → 回傳重印警告且不執行;確認後取消照常執行。
  - [ ] `staff` 嘗試取消 → `permission_denied`,訂單維持 `processing`。
  - [ ] 提交成功後同部門看板連線收到 `BoardEvent`。

### 子功能 5.1.4:派車通知觸發

- **目標**:批次確認成功後,對每筆已派訂單發送派車通知(FCM + 站內)。`相依: 4.4`(FCM / 站內通知發送)、`4.3`(通知範本)
- **檔案**:
  - Create/Update: `backend/internal/domain/salesorders/dispatch.go`(確認流程中的通知觸發點)
  - Update: `backend/internal/notification/`(若派車事件型別尚未涵蓋,補對應範本 type;通道與發送機制本身屬 07-notifications)
- **介面**:內部呼叫通知模組發送介面(非 RPC);通知範本於 `notification_templates` 以公司/部門級維護,事件型別 = dispatch,變數含訂單編號、車次名稱、預計出貨日。
- **實作邏輯**:
  1. 於 5.1.2 批次交易中**僅收集**成功派車的訂單清單,不在 DB 交易內發送;通知於交易提交後逐筆觸發(發送失敗不回滾派車,依 D16 失敗僅標記)。
  2. 每筆訂單:查詢該訂單部門適用的 dispatch 範本(`fcm` / `in_app` 兩通道);無範本 → 略過該通道。
  3. 套用範本變數(訂單編號、車次、日期),建立 `notifications` 記錄並呼叫 4.4 的發送流程。
  4. 發送失敗不重試:`notifications.status` 標記 `failed` 並記錄原因;FCM 回報失效 token 由 4.4 既有機制刪除,不在此重複實作。
- **錯誤處理**:
  - 通知為 fire-and-record:發送失敗僅落 `notifications.status = failed`,不向 Confirm 呼叫者回錯、不影響批次結果。
  - 範本不存在或範本變數缺漏 → 該通道記 failed 並帶原因。
- **驗收**:
  - [ ] 批次確認成功後,每筆派車訂單各產生 `notifications` 記錄(fcm / in_app 依範本)。
  - [ ] 通知內容含訂單編號、車次名稱、預計出貨日。
  - [ ] 發送失敗時訂單仍為 `processing`,`notifications` 可查失敗原因。
  - [ ] 無對應範本時不發送、不報錯。

---

## Task 5.2:Connect 串流即時派車看板(僅後端)

對應原計畫 `### Task 5.2` 的後端部分(Step 1、Step 2、heartbeat)。前端 connect-web 訂閱、invalidate 重查、exponential backoff 重連與 30 秒輪詢降級屬原計畫 Task 5.6,本文件僅於 5.2.3 註記後端配合點。核心檔案 `backend/internal/domain/salesorders/watch_board.go`。

### 子功能 5.2.1:WatchBoard proto 與串流 handler

- **目標**:提供部門看板訂閱的 server streaming RPC,連線走既有認證,註冊後接收本部門看板事件。
- **檔案**:
  - Create/Update: `backend/proto/v1/dispatch.proto`(WatchBoard streaming RPC 與 `BoardEvent`)
  - Create: `backend/internal/domain/salesorders/watch_board.go`(handler + 訂閱註冊表)
- **介面**:Connect-RPC `DispatchService.WatchBoard`(server streaming)
  - Request:`expected_delivery_date`(看板日期;事件依部門推送,日期僅供前端語意,後端可依訂閱註冊)。
  - Stream Response `BoardEvent`:`sales_order_id`、`route_id`、`delivery_sequence`、`version`、事件種類(route_assign / dispatch / dispatch_cancel)、`department_id`。
  - 認證:與其他 RPC 相同,Web 走 httpOnly cookie、App 走 token header,由既有 auth interceptor 處理;**不實作一次性 ticket、不接受 URL 參數憑證**(D14)。
- **實作邏輯**:
  1. auth interceptor 先行:無有效 session/JWT → `unauthenticated`,不建立任何訂閱。
  2. 由注入的 `data_scope` 解析訂閱部門:看板為部門級,訂閱 key = `department_id`;無部門上下文(未選部門)或 `data_scope` 無法定位部門 → `failed_precondition`。
  3. 於程序內訂閱註冊表(department_id → 連線集合)登記本連線,並向 5.2.2 的 Valkey 訂閱轉發層註冊該部門 channel。
  4. 迴圈等待:收到該部門事件 → 寫入串流;客戶端斷線(context 取消)→ 從註冊表移除、若該部門於本 replica 已無連線則退訂 Valkey channel,釋放資源後結束 handler。
  5. 事件僅轉發不過濾日期:後端按部門廣播,前端自行判斷是否影響當前日期看板並 invalidate(事件僅作失效提示)。
- **錯誤處理**:
  - `unauthenticated`:未攜帶有效 cookie/token 建立串流。
  - `permission_denied`:無看板功能權限。
  - `failed_precondition`:無法解析部門上下文。
  - 串流寫入失敗(客戶端已離線)→ 清理註冊並靜默結束,不回錯給已斷線端。
- **驗收**:
  - [ ] 有效 cookie/token 可建立串流;未認證連線被拒且不建立訂閱。
  - [ ] 部門 A 連線只收到部門 A 事件,部門 B 連線不收部門 A 事件。
  - [ ] 客戶端斷線後註冊表無殘留,該部門無連線時本 replica 退訂對應 Valkey channel。
  - [ ] `BoardEvent` 欄位齊全(sales_order_id / route_id / delivery_sequence / version / 事件種類)。

### 子功能 5.2.2:Valkey pub/sub 跨 replica 事件轉發

- **目標**:派車 mutation 提交後以部門分 channel 發佈事件,任意 replica 上的看板連線皆能收到(D14)。
- **檔案**:
  - Create/Update: `backend/internal/domain/salesorders/watch_board.go`(訂閱轉發層)
  - Create/Update: `backend/internal/domain/salesorders/dispatch.go`(三個 mutation 的發佈點)
- **介面**:內部機制,無 RPC;channel 命名以 `department_id` 分段(一部門一 channel),訊息載體 = 序列化後的 `BoardEvent`。
- **實作邏輯**:
  1. 發佈點統一規則:AssignRoute / Confirm / CancelDispatch 的 DB 交易**提交成功後**才發佈(交易內不發佈,避免事件先於可見資料送達;rollback 不產生任何事件)。
  2. 發佈內容:該筆訂單最新 `route_id` / `delivery_sequence` / `version` 與事件種類;Confirm 批次對每筆成功訂單各發一則。
  3. 訂閱轉發層:每 replica 對「本機有連線的部門」訂閱對應 Valkey channel;收到訊息 → 反序列化 → 依註冊表轉發該部門所有本機連線。
  4. 容錯語意:pub/sub 為 at-most-once,斷線期間事件遺失屬預期,由前端重連後全量重查補齊(D14);發佈失敗(Valkey 異常)僅記錄錯誤日誌與指標,不影響 mutation 結果、不重試。
  5. 指標:發佈/轉發/連線數納入 `/metrics`(配合 D19 監控)。
- **錯誤處理**:
  - Valkey 連線中斷 → 訂閱層自動重連並重新訂閱本機仍有連線的部門 channel;重連期間事件遺失由前端重查語意吸收,後端不補發。
  - 訊息反序列化失敗 → 丟棄該則並記錄,不影響其他連線。
- **驗收**:
  - [ ] replica 1 提交的派車變更,replica 2 上的同部門串流連線即時收到事件。
  - [ ] 三種 mutation(AssignRoute / Confirm / CancelDispatch)各自觸發對應事件種類。
  - [ ] DB 交易 rollback 時無任何事件發佈。
  - [ ] 事件依部門隔離,跨部門連線收不到。
  - [ ] Valkey 異常時 mutation 仍成功,僅推播降級(記錄可查)。

### 子功能 5.2.3:串流 heartbeat 與前端降級註記

- **目標**:server 定期送出 heartbeat,防止 ingress / proxy idle timeout 切斷長串流;前端重連與輪詢降級為前端職責,此處僅註記後端配合點。
- **檔案**:
  - Create/Update: `backend/proto/v1/dispatch.proto`(`BoardEvent` 增加 heartbeat 種類,或採 oneof 事件載體)
  - Create/Update: `backend/internal/domain/salesorders/watch_board.go`
- **介面**:`BoardEvent` 事件種類含 `heartbeat`(無業務欄位);發送間隔以常數設定,須小於 ingress idle timeout(預設 25 秒,低於常見 30 秒閾值,可組態調整)。
- **實作邏輯**:
  1. 每條 WatchBoard 串流建立獨立 ticker,間隔到達且期間無業務事件送出時,送出 heartbeat 事件。
  2. heartbeat 寫入失敗視同客戶端斷線,走 5.2.1 的清理流程。
  3. heartbeat 不寫 `sales_order_events`、不發佈至 Valkey,純連線維持用途,不進入跨 replica 轉發。
  4. **前端職責註記(後端不實作)**:前端收到任何事件(含 heartbeat)僅作看板查詢 invalidate 全量重查,不做 cache patch;斷線以 exponential backoff 重連並於重連後全量重查;連續建立失敗達閾值降級為 30 秒輪詢(視窗隱藏暫停、聚焦立即重查),串流可建立時恢復推播——以上屬原計畫 Task 5.6 前端實作,後端僅需保證:輪詢降級使用的看板查詢 RPC 與串流並存可用、heartbeat 間隔可由部署組態調整以配合不同 ingress。
- **錯誤處理**:
  - heartbeat 僅盡力發送;連續寫入失敗即清理連線,不向上層回錯。
  - ticker 間隔組態 ≥ ingress timeout 時啟動日誌警告(防誤配)。
- **驗收**:
  - [ ] 無業務事件時,串流連線於間隔內持續收到 heartbeat,長時間閒置不被 ingress 切斷。
  - [ ] heartbeat 不產生任何業務事件記錄、不進入 Valkey channel。
  - [ ] 看板查詢 RPC 在串流不可用時仍可正常回應(供降級輪詢使用)。

---

## 整合測試重點

1. **樂觀鎖衝突(5.1.1)**:兩併發 AssignRoute 攜帶相同 `version` 提交,僅一筆成功,另一筆回 `failed_precondition` 且資料不變;成功方 `version` 遞增。
2. **跨 replica 事件送達(5.2.2)**:起兩個後端程序共享同一 Valkey;於程序 A 執行 AssignRoute / Confirm / CancelDispatch,程序 B 上的 WatchBoard 連線收到對應 `BoardEvent`;部門 B 連線確認收不到部門 A 事件。
3. **批次確認部分失敗(5.1.2)**:批次中混入已取消訂單,回應明確區分成功/失敗,失敗筆無 `dispatched_at`、成功筆各有 dispatch 事件與通知記錄。
4. **取消派車事務(5.1.3)**:取消成功後 `sales_order_events` 與 `audit_logs` 同時存在(D18 同事務);已列印車次未確認時不執行且回警告;`staff` 呼叫回 `permission_denied`。
5. **串流生命週期(5.2.1 / 5.2.3)**:未認證連線被拒;閒置串流持續收到 heartbeat 不被切斷;客戶端斷線後註冊表清空、本 replica 對無連線部門退訂 channel。
6. **交易 rollback 不發事件**:強制 mutation 交易失敗,驗證 Valkey 無任何事件發佈、看板連線無感知。
