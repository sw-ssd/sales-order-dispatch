# 03 — 字典檔(metadicts)與稽核日誌(audit_logs)細部實作計畫

> - **文件標題**:Backend 細部分解 03 — 字典檔與稽核日誌
> - **對應原計畫 Task**:`docs/superpowers/plans/reference/2026-07-17-sales-order-1-0-tasks.md` 的 **Task 2.5**(metadicts 字典檔 API)、**Task 2.6**(audit_logs 稽核日誌 API)
> - **對應規格書 v1.0.34 章節**:§5.1 核心實體清單、§5.2 實體欄位(`metadicts` / `audit_logs`)、§5.3 字典檔策略、§5.4 軟刪除慣例、§10.2 稽核日誌、§14.1 資料保留、§14.4 稽核與監控
> - **對應決策記錄**:D10(軟刪除 + 部分唯一索引)、D11(字典檔單表 + `department_id IS NULL` 語意)、D18(稽核同事務同步寫入)、D21(整合測試類別)、D27(稽核保留期限)
> - **相依文件**:`01-auth.md`(登入稽核來源、Session/JWT 身分注入、RLS middleware)、`02-tenancy-users.md`(公司/部門實體、角色與 data_scope;`2.9` 角色異動為稽核來源之一)
> - **被相依文件**:`04-master-data.md`(主檔異動稽核、單位字典選用)、`05-sales-orders.md`(下單稽核、訂單來源字典)、`06-returns.md`、`07-notifications.md`、`08-dispatch.md`(取消派車/作廢稽核)、`09-printing.md`(列印稽核)

---

## 0. 適用共通規則(引用 `00-index.md` §3,不重複定義)

- 軟刪除 `deleted_at` + 部分唯一索引(`metadicts` 適用;`audit_logs` **不適用**軟刪除,見 2.6.1)。
- 業務資料帶 `company_id` / `department_id`,Casbin 管功能、RLS 注入(`app.current_company_id` / `app.current_department_id` / `app.current_data_scope`)為資料範圍最後防線。
- 稽核與業務操作同一 DB 交易,同成功同失敗(D18);樂觀鎖衝突一律 `failed_precondition`。
- 錯誤統一以 Connect code 表述;列表分頁沿用 `meta` 格式、`per_page` 上限 100。

---

## 1. Task 2.5:metadicts 字典檔 API

> 原計畫 Goal:系統級預設 + 部門擴充的字典管理。規格依據 §5.3 與 D11。
> 字典分層備忘:倉別/車次/分切規格/商品分類為部門級**獨立實體表**(屬 `04-master-data.md`,不在此);訂單來源為系統級固定(seed,不開放異動);本 Task 處理「系統級預設 + 部門可擴充」的單位、付款方式、結帳方式、客戶類型、發票類型、規格、代切實重肉片等。

### 子功能 2.5.1: metadicts Ent schema 與 migration

- **目標**: 建立單一 `metadicts` 表,以 `department_id IS NULL` 表達系統預設、非 NULL 表達部門擴充,並 seed 系統級字典預設值。
- **檔案**:
  - Create `backend/ent/schema/metadict.go`
  - Create `backend/database/migrations/000XX_metadicts.sql`(含索引、RLS policy、seed)
  - Create `backend/internal/domain/metadicts/`(套件目錄,後續子功能填入)
- **介面**: Ent 實體 `Metadict`,欄位:`id`(UUID)、`type`(字典類型,字串,如 `unit` / `payment_method` / `settlement_method` / `customer_type` / `invoice_type` / `order_source`)、`code`(程式識別,同 type + 部門範圍內唯一)、`display_name`(顯示名稱)、`department_id`(nullable,FK `departments.id`;NULL = 系統預設)、`sort_order`(整數排序)、`is_active`(布林)、`deleted_at`(nullable timestamp)、`created_at` / `updated_at`。索引:查詢索引 `(type, department_id, is_active)`;唯一約束 `(type, code, department_id)` 搭配軟刪除。
- **實作邏輯**:
  1. 依上列欄位定義 Ent schema;`department_id` 設為 optional + nillable,對應資料庫 NULL 語意;`deleted_at` 依共通軟刪除慣例(D10)。
  2. 唯一約束以兩道部分唯一索引實作(避開 NULL 不參與唯一比較的陷阱):一道涵蓋系統預設(`WHERE department_id IS NULL AND deleted_at IS NULL`),一道涵蓋部門擴充(`WHERE department_id IS NOT NULL AND deleted_at IS NULL`);若 PostgreSQL 版本 ≥15,可改用 `NULLS NOT DISTINCT` 單一唯一索引搭配軟刪除條件,擇一並於 migration 註記(D11 允許兩種作法)。
  3. 建立 RLS policy:讀取允許 `department_id IS NULL` 或等於 session 注入的 `app.current_department_id`;寫入限定 `data_scope = all`(super,系統級)或 `department_id = app.current_department_id`(部門級擴充)。
  4. migration seed 系統級預設字典:`type` 含單位、付款方式、結帳方式、客戶類型、發票類型等,`department_id` 一律 NULL;訂單來源(`order_source`,如 `W` = Web 中台、`A` = App)seed 為系統級固定值,供 `05-sales-orders.md` 取號使用。
  5. seed 必須冪等(重複執行不產生重複列、不報錯),以唯一約束衝突時略過的方式處理。
  6. 交易邊界:本子功能僅 schema 與 migration,無運行時交易。
- **錯誤處理**: migration 失敗(語法、FK 不存在)即中止部署,不得部分套用;seed 與既有列衝突時略過不報錯(冪等);`departments` 表不存在時依賴錯誤於部署階段暴露(本檔相依 `02-tenancy-users.md` 的部門 schema 先建立)。運行時錯誤碼不適用。
- **驗收**:
  - [ ] migration 於乾淨資料庫執行成功,且重複執行第二次不報錯、不產生重複 seed 列。
  - [ ] 同一 `type + code` 在系統預設(department_id NULL)與某部門擴充可並存;同範圍內重複 `code` 被資料庫拒絕。
  - [ ] 軟刪除一列後,同範圍可立即重建同 `code` 的新列(部分唯一索引生效)。
  - [ ] seed 後可查得訂單來源 `W` / `A` 等系統級固定字典。

### 子功能 2.5.2: MetadictService CRUD 與軟刪除

- **目標**: 提供字典維護 RPC,系統級僅 super 可寫、部門擴充由所屬部門管理角色維護;刪除採軟刪除並同事務寫稽核。相依: 2.5.1、2.6.2(稽核寫入入口)。
- **檔案**:
  - Create `proto/metadict/v1/metadict.proto`
  - Create `backend/internal/domain/metadicts/service.go`(Connect handler)
  - Create `backend/internal/domain/metadicts/usecase.go`
  - Create `backend/internal/domain/metadicts/repository.go`
- **介面**: Connect-RPC `MetadictService` —
  - `List(ListMetadictsRequest{type 篩選、include_inactive、include_deleted}) returns (ListMetadictsResponse{items, meta 分頁})`
  - `Get(GetMetadictRequest{id}) returns (Metadict)`
  - `Create(CreateMetadictRequest{type, code, display_name, sort_order, is_active;scope 由身分推導}) returns (Metadict)`
  - `Update(UpdateMetadictRequest{id, display_name, sort_order, is_active}) returns (Metadict)`
  - `Delete(DeleteMetadictRequest{id}) returns (空)`
  - (`ListOptions` 另列於 2.5.4)
  - `Metadict` 訊息形狀:`id, type, code, display_name, department_id(可空), sort_order, is_active, created_at, updated_at`。
- **實作邏輯**:
  1. 寫入範圍推導:操作者為 super(`data_scope = all`)時,建立的列 `department_id` 設 NULL(系統級);dept_admin / staff 時,強制帶入 session 的當前部門 ID,**不接受請求自帶 `department_id`**(防越權建立他部門資料)。
  2. 修改限制:`type` 與 `code` 建立後不可修改(程式識別穩定性,`code` 會被 `product_units.unit_code` 等引用);Update 僅允許 `display_name` / `sort_order` / `is_active`。
  3. 系統級固定字典(如 `order_source`)不可由 API 異動:Update / Delete 目標屬此類型時拒絕,提示為系統固定值。
  4. 1.0 不支援部門覆寫系統預設(D11):部門擴充列與系統預設列各自獨立,不提供「覆寫同 code 系統值」的語意;同 `type + code` 在部門範圍內仍受唯一約束。
  5. Delete = 軟刪除:設定 `deleted_at`,不實體移除;復原不在 1.0 API 範圍(如需,由 super 管理介面清 `deleted_at` 並寫稽核,依 §5.4)。
  6. 交易邊界:Create / Update / Delete 皆為「業務異動 + `audit_logs` 寫入」同一 DB 交易(D18),action 分別為 `create` / `update` / `delete`,`resource_type = metadict`;稽核寫入失敗則整體回滾。
  7. 軟刪除主檔引用規則:已軟刪除的字典值,歷史單據仍可顯示名稱,但 ListOptions(2.5.4)不再回傳(不可再選用,依 §5.4)。
- **錯誤處理**: 未登入或 session 失效 → `unauthenticated`;非 super 嘗試建立/修改系統級列、跨部門異動、或客戶身分呼叫 → `permission_denied`;目標不存在或已軟刪除 → `not_found`;`type`/`code` 嘗試修改、異動系統固定字典 → `failed_precondition`;必填缺漏、`code` 格式不符(空白、過長)→ `invalid_argument`;同範圍 `type + code` 重複 → `already_exists`。
- **驗收**:
  - [ ] super 可建立/修改/軟刪除系統級字典;dept_admin 僅能操作自己部門的擴充列,且建立的列自動帶入自己部門 ID。
  - [ ] dept_admin 對系統級列或他部門列執行 Update / Delete 回 `permission_denied`(RLS 與應用層雙重攔截)。
  - [ ] Update 帶 `code` 變更或對 `order_source` 異動,回 `failed_precondition`。
  - [ ] 軟刪除後 List 預設查不到、Get 回 `not_found`;同 `code` 可立即重建(`already_exists` 不再觸發)。
  - [ ] Create / Update / Delete 成功後,`audit_logs` 存在對應紀錄且與業務異動同生滅(強制失敗注入稽核寫入時,業務異動一併回滾)。

### 子功能 2.5.3: 系統預設 + 部門擴充合併查詢

- **目標**: List / Get 的可見性統一為「系統預設 + 當前部門擴充」,部門間擴充值完全隔離。相依: 2.5.1。
- **檔案**:
  - Update `backend/internal/domain/metadicts/repository.go`
- **介面**: 不新增 RPC;作為 `MetadictService.List` / `Get` 的查詢實作,供 2.5.4 ListOptions 與其他 domain(如主檔表單)內部複用。
- **實作邏輯**:
  1. 可見範圍條件固定為:`department_id IS NULL` **或** `department_id = 當前部門 ID`(自 session 注入值取得,非請求參數)。
  2. 一律附加 `deleted_at IS NULL`;`include_deleted` 僅 super 管理介面可開啟(依 §5.4「顯示已刪除」開關)。
  3. super(`data_scope = all`)查詢時不綁單一部門:回傳系統預設全部 + 可依請求參數指定部門檢視其擴充(管理用途);無指定時僅回系統預設。
  4. 排序:先 `sort_order` 升冪,同值再按 `code` 升冪;系統預設與部門擴充混排,不回傳額外層級標記以外的區分(列上自有 `department_id` 可辨識來源)。
  5. RLS 為最後防線:應用層 where 條件之外,RLS policy 再保證跨部門列不可見;兩層條件一致,避免「應用層漏寫、RLS 擋下」造成的靜默少資料情形以測試覆蓋。
  6. 邊界:操作者無部門(如 company_admin 未隸屬部門)時,僅見系統預設;不得因 NULL 比較錯誤而看到全部部門擴充。
- **錯誤處理**: session 缺少部門脈絡卻以部門角色呼叫 → `failed_precondition`(帳號資料異常,需重新登入或聯繫管理員);其餘沿用 List 的一般錯誤(未認證 → `unauthenticated`)。
- **驗收**:
  - [ ] 部門 A 使用者 List 同時見到系統預設與部門 A 擴充,且完全看不到部門 B 擴充。
  - [ ] 無部門角色(company_admin)僅見系統預設,不回 `invalid_argument` 也不漏系統值。
  - [ ] 關閉某擴充值(`is_active = false`)後,List 預設仍回傳(管理視角),ListOptions 不回傳(見 2.5.4)。
  - [ ] 以 SQL 層直接驗證 RLS:換部門 session 變數後,他部門擴充列在資料庫層即不可見。

### 子功能 2.5.4: ListOptions 表單下拉選項服務

- **目標**: 提供輕量選項查詢供 Web 中台與 App 表單下拉使用,僅回「可選用」的選項。相依: 2.5.3。
- **檔案**:
  - Update `proto/metadict/v1/metadict.proto`
  - Update `backend/internal/domain/metadicts/service.go`、`usecase.go`、`repository.go`
- **介面**: Connect-RPC `MetadictService.ListOptions(ListOptionsRequest{type, 可選 keyword}) returns (ListOptionsResponse{options})`;`options` 元素形狀:`code, display_name`(精簡,不含 id、排序、狀態欄位)。回應不分頁(字典選項量級小,單次全回)。
- **實作邏輯**:
  1. 篩選條件 = 2.5.3 合併可見範圍 + `is_active = true` + `deleted_at IS NULL` + 指定 `type`;僅回可選用值,軟刪除與停用值一律排除(§5.4:不可再選用於新單據)。
  2. `keyword` 為可選:對 `display_name` 與 `code` 做前綴/包含比對,供選項多時過濾;未帶則全回。
  3. 排序同 2.5.3(`sort_order`、`code`)。
  4. 快取:字典異動頻率低,允許短 TTL 記憶體快取;Create / Update / Delete(2.5.2)成功時主動失效對應 `type` 快取,避免表單拿到舊選項。
  5. 呼叫身分:所有登入角色皆可呼叫(表單通用);客戶端身分僅回系統預設(客戶無部門脈絡),不視為錯誤。
  6. 交易邊界:純讀取,無交易與稽核(下拉查詢非 D18 關鍵操作)。
- **錯誤處理**: 未帶 `type` 或 `type` 為系統未定義值 → `invalid_argument`;未認證 → `unauthenticated`;查無選項回空陣列(非錯誤)。
- **驗收**:
  - [ ] 表單下拉取得 `unit`、`payment_method` 等 type 的選項,內容為系統預設 + 自己部門擴充的啟用值。
  - [ ] 停用或軟刪除某選項後,ListOptions 立即不再回傳(快取失效生效),但歷史單據顯示不受影響。
  - [ ] 客戶端(App 客戶身分)呼叫僅得系統預設選項,不揭露任何部門擴充。

---

## 2. Task 2.6:audit_logs 稽核日誌 API

> 原計畫 Goal:記錄關鍵操作並提供查詢。規格依據 §10.2、§14.4 與 D18、D27。

### 子功能 2.6.1: audit_logs Ent schema 與 migration

- **目標**: 建立獨立 `audit_logs` 表,記錄操作者、動作、資源、變更前後與來源網路資訊;不可軟刪除、不可由 API 修改。相依: 02-tenancy-users.md(companies / departments / users schema)。
- **檔案**:
  - Create `backend/ent/schema/auditlog.go`
  - Create `backend/database/migrations/000XX_audit_logs.sql`(含索引、RLS policy;**不建** deleted_at)
- **介面**: Ent 實體 `AuditLog`,欄位:`id`(UUID)、`company_id`(FK)、`department_id`(nullable FK;跨部門或公司級操作可空)、`user_id`(操作者 FK;系統排程操作可空並以 action 語意區分)、`action`(列舉字串:`create` / `update` / `delete` / `login` / `logout` / `print` / `force_logout` / `role_change` / `dispatch_cancel` / `void`,依 §5.2)、`resource_type`(字串,如 `metadict` / `customer` / `sales_order`)、`resource_id`(字串,容納 UUID 與單號)、`before_snapshot`(JSONB,可空)、`after_snapshot`(JSONB,可空)、`ip_address`、`user_agent`、`created_at`。索引:`(company_id, created_at)`、`(resource_type, resource_id)`、`(user_id, created_at)`;查詢篩選主力為時間與資源,索引依此配置。
- **實作邏輯**:
  1. 欄位對齊 §5.2;`created_at` 為 `timestamptz`,本表無 `updated_at`(紀錄不可變)。
  2. 明確**不建** `deleted_at`(D10 / §5.4:稽核日誌不適用軟刪除);不提供任何 Update / Delete 路徑,僅保留期限排程(D27,Phase 8 維運範圍)可直接 SQL 到期刪除,schema 不得依賴軟刪除標記。
  3. `before_snapshot` / `after_snapshot` 存異動前後摘要 JSON;login / logout 等無資源異動者兩者皆可空,以 action 語意為準。
  4. RLS policy:讀取限定 `data_scope = all`(super)或 `company_id = app.current_company_id`(company_admin);寫入僅限後端 service 角色(應用連線),不由一般查詢角色寫入。
  5. 交易邊界:本表寫入一律發生於業務交易內(2.6.2),schema 層不需額外機制。
- **錯誤處理**: migration 失敗即中止部署;FK 目標(companies / departments / users)不存在時於部署階段暴露依賴錯誤。運行時錯誤碼不適用(本表不對外開寫)。
- **驗收**:
  - [ ] migration 成功建立表與三組索引;表結構不含 `deleted_at` / `updated_at`。
  - [ ] 嘗試以應用連線對 `audit_logs` 做 Update / Delete 被資料庫角色權限拒絕。
  - [ ] `before_snapshot` / `after_snapshot` 可寫入任意合法 JSON 並原樣讀回。

### 子功能 2.6.2: usecase 層統一同事務寫入機制

- **目標**: 所有關鍵操作的稽核寫入收斂為 usecase 層單一入口,與業務異動同一 DB 交易,同成功同失敗(D18)。相依: 2.6.1;供 01-auth(登入/登出/強制登出)、02-tenancy-users(角色變更)、04-master-data(主檔異動/刪除)、05-sales-orders(下單)、08-dispatch(取消派車/作廢)、09-printing(列印)等呼叫。
- **檔案**:
  - Create `backend/internal/domain/auditlogs/recorder.go`(統一入口與 helper)
  - Create `backend/internal/domain/auditlogs/service.go`(Connect handler,`Record` 僅內部掛載、不對外公開路由)
- **介面**: Connect-RPC `AuditService.Record(RecordAuditRequest{action, resource_type, resource_id, before_snapshot, after_snapshot}) returns (空)` 標記為**內部使用**(gateway 不暴露,僅供同進程或其他內部服務呼叫);Go 層主要介面為 recorder helper:接受「交易內 Ent client + 請求脈絡(操作者、公司、部門、IP、User-Agent)+ 稽核內容」,於當前交易寫入一筆 `audit_logs`。
- **實作邏輯**:
  1. 統一入口原則:各 domain usecase 不得自行組裝 `audit_logs` 寫入,一律經 recorder;recorder 只接受交易內 client(tx client),從簽名上杜絕「業務已 commit、稽核裸寫」的脫鉤寫法。
  2. 脈絡取值:`user_id` / `company_id` / `department_id` 自 session 注入的請求脈絡取得;`ip_address` / `user_agent` 自 Connect 請求標頭取得,由 middleware 放入脈絡;缺 IP 時存空不視為錯誤。
  3. snapshot 組裝規則:Create 只填 `after_snapshot`;Update 填前後兩者(僅收錄變更欄位摘要,不整列複製,避免敏感欄位如密碼雜湊入稽核——敏感欄位一律排除);Delete 只填 `before_snapshot`。
  4. 失敗策略:稽核寫入失敗 = 整體交易回滾,業務操作對呼叫方回錯;**不得**降級為「略過稽核繼續」(D18:稽核缺漏比稽核擋下操作更糟)。
  5. 關鍵操作清單(D18 / §10.2)與 action 對應:登入/登出 → `login` / `logout`;下單、主檔新增 → `create`;主檔修改 → `update`;刪除(軟刪除) → `delete`;列印 → `print`;強制登出 → `force_logout`;角色變更 → `role_change`;取消派車 → `dispatch_cancel`;作廢 → `void`。新 domain 接入時於該 domain 文件標註對應 action。
  6. developer 帳號操作照常經 recorder 寫入,不因繞過 Casbin/RLS 而省略(§4.4)。
  7. 交易邊界:recorder 本身不開交易;它假設呼叫方已在業務交易中,寫入與業務異動共用同一 commit / rollback。
- **錯誤處理**: 脈絡缺操作者身分(應為已認證請求) → `unauthenticated`;recorder 被以非交易 client 呼叫 → 視為程式錯誤,直接失敗並回滾(`internal`,需測試捕捉);稽核寫入 DB 錯誤 → 業務 RPC 回 `internal` 且業務異動已回滾(不可部分成功)。
- **驗收**:
  - [ ] 任一關鍵操作(以登入與主檔修改為代表)成功後,`audit_logs` 各存在 action 正確、操作者與 IP 齊全的一筆。
  - [ ] 注入稽核寫入失敗時,對應業務異動在資料庫中不存在(回滾驗證)。
  - [ ] developer 帳號操作產生稽核紀錄;敏感欄位(密碼雜湊等)不出現於任何 snapshot。
  - [ ] 取消派車、作廢除 `sales_order_events` 外,另存在 `dispatch_cancel` / `void` 稽核紀錄(與 08-dispatch 聯合驗收)。

### 子功能 2.6.3: 稽核查詢 API

- **目標**: super / company_admin 依時間、操作類型、資源篩選查詢所屬範圍的稽核紀錄。相依: 2.6.1。
- **檔案**:
  - Update `proto/audit/v1/audit.proto`(或 Create)
  - Create `backend/internal/domain/auditlogs/query.go`(List handler 與 repository)
- **介面**: Connect-RPC `AuditService.List(ListAuditLogsRequest{時間起訖、action、resource_type、resource_id、user_id 篩選;page, per_page}) returns (ListAuditLogsResponse{items, meta 分頁})`;`items` 元素形狀:`id, action, resource_type, resource_id, user_id(含操作者顯示名稱), company_id, department_id, before_snapshot, after_snapshot, ip_address, user_agent, created_at`。
- **實作邏輯**:
  1. 範圍推導:super(`data_scope = all`)可查全系統,可選帶 `company_id` 篩選;company_admin 強制限定自己公司(忽略請求自帶的他公司篩選);其餘角色不可呼叫。
  2. 篩選組合:時間起訖(預設近 3 個月,對齊預設保留期限 D27)、`action` 單選、`resource_type` + `resource_id` 定位單一資源軌跡、`user_id` 查特定操作者;條件皆為 AND。
  3. 排序固定 `created_at` 降冪(最新在前);分頁沿用 `meta` 格式,`per_page` 上限 100(§ 共通規則)。
  4. 效能:時間範圍為必填防呆之一——未帶任何篩選時套用預設時間窗,避免全表掃描;查詢一律走 2.6.1 索引欄位。
  5. RLS 為最後防線:company_admin 範圍由 RLS 保證,應用層 where 為第一層。
  6. 交易邊界:純讀取,無交易;查詢行為本身不寫稽核(避免自我遞迴)。
- **錯誤處理**: 未認證 → `unauthenticated`;dept_admin / staff / customer 呼叫 → `permission_denied`;時間起訖顛倒、`per_page` 超上限 → `invalid_argument`;查無資料回空陣列(非錯誤)。
- **驗收**:
  - [ ] super 可跨公司查詢並以 `company_id` 篩選;company_admin 查詢結果僅含自己公司紀錄(含 RLS 層驗證)。
  - [ ] 以時間 + action + resource 組合篩選可正確收斂結果;`resource_type + resource_id` 可拉出單一資源完整軌跡。
  - [ ] staff 呼叫回 `permission_denied`;未帶篩選時套用預設 3 個月時間窗且分頁 `meta` 正確。

---

## 3. 整合測試重點

依 D21 六類整合測試中與本文件相關者,外加本 domain 特有恆等式:

1. **RLS 隔離**(2.5.3 / 2.6.3):部門 A / B 擴充字典互不可見;company_admin 稽核查詢不越公司;以資料庫 session 變數直接驗證,不只走應用層。
2. **同事務一致性**(D18,2.5.2 / 2.6.2):注入稽核寫入失敗 → 業務異動不存在;反之業務異動失敗 → 無孤兒稽核列。對 Create / Update / Delete 各驗一條路徑。
3. **軟刪除 + 部分唯一索引**(D10,2.5.1 / 2.5.2):軟刪除字典值後同 `type + code` 可立即重建;未刪除時重建被 `already_exists` 拒絕。
4. **合併查詢邊界**(2.5.3):無部門角色僅見系統預設;`department_id IS NULL` 比較不得誤放全部部門擴充。
5. **選項服務即時性**(2.5.4):停用/刪除後 ListOptions 立即排除(含快取失效),歷史單據顯示不受影響。
6. **稽核完整性**(2.6.2):關鍵操作清單逐項抽查「有操作必有稽核」,含 developer 帳號;snapshot 不含敏感欄位。
7. **稽核查詢效能防呆**(2.6.3):無篩選查詢套用預設時間窗,`per_page` > 100 被拒;大量紀錄下時間 + 公司篩選走索引(以 EXPLAIN 佐證)。

---

*最後更新:2026-08-17*
