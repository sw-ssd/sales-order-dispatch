# 01 — 認證與授權(Task 1.1–1.8、1.11)

> - 對應原計畫:Task 1.1「users 與角色 Ent schema」、1.2「Casbin model 與 policy 儲存」、1.3「RLS policies」、1.4「OAuth2/OIDC 員工登入」、1.5「客戶密碼登入」、1.6「Session / JWT 管理」、1.7「強制登出」、1.8「ability API」、1.11「開發者帳號與繞過機制」(規格 §3、§4、決策 D1–D6、D8、D15、D18、D21)。
> - 對應規格書 v1.0.34 章節:§3 認證與授權(§3.1 員工 OAuth2、§3.2 客戶帳密、§3.3 Session/JWT、§3.4 RBAC/RLS/CASL 三層)、§4.2 客戶帳號(臨時密碼、鎖定)、§5.1 資料表(users/companies/departments)、§14 安全。
> - 相依文件:無(本文件為所有 domain 的前置);被相依:`02-tenancy-users.md`(2.9 接管 1.8)、其餘全部文件(認證 middleware、RLS 注入、稽核入口)。
> - 範圍註記:Task 1.9(前端登入與路由守衛)、1.10(App 登入)為前端/App 範圍,不在本文件;本文件僅提供其所需後端介面(1.4 導向端點、1.6 JWT/refresh、1.8 ability、3.8 QR token 見 `04-master-data.md`)。

---

## 共通規則

以下規則適用本文件所有子功能,各欄不再重複(完整定義見 `00-index.md` §3):

1. **交易與稽核**:登入成功/失敗、強制登出、密碼重置、guest 審核等關鍵操作與 audit log 寫入皆同一 DB 交易(D18);外部呼叫(如 OIDC token 交換)不納入交易。
2. **軟刪除**:`users` 採 `deleted_at` 軟刪除,查詢預設排除;唯一性以部分唯一索引處理(D10)。
3. **多租戶**:`users` 帶 `company_id` / `department_id`;Casbin 管功能(domain = company_id)、RLS 管資料範圍、CASL 管 UI(D3)。RLS 注入為最後防線。
4. **錯誤處理**:統一以 Connect code 表述;認證失敗一律 `unauthenticated` 且不透露帳號是否存在;授權失敗 `permission_denied`。
5. **密碼安全**:雜湊一律 Argon2id;密碼、refresh token 僅存雜湊,不明文落盤、不入稽核快照。

---

## 子功能 1.1.1: Company schema

- **目標**: 建立公司主檔 Ent schema,支撐多租戶頂層實體與公開發現。
- **檔案**:
  - Create `backend/ent/schema/company.go`
  - Create 對應 migration
- **介面**: 無對外 RPC;產出 `companies` 表與 `ent.Company`。
  - 欄位:`id`(UUID)、`name`、`tax_id`(統編,可空)、`identifier`(公開發現代碼,全系統唯一)、`customer_code_prefix`(客戶編碼前綴,大寫英數 1–4 碼,全系統唯一,D7)、`status`(enum:`active` / `suspended`)、`public_info`(JSONB,公開簡介/電話/地址)、`capabilities`(JSONB,功能開關)、`logo_file_id`(關聯 `file_assets`,可空;Phase 2 Task 2.4 啟用)、`created_at` / `updated_at` / `deleted_at`。
- **實作邏輯**:
  1. `identifier` 與 `customer_code_prefix` 各自建部分唯一索引(`WHERE deleted_at IS NULL`);`customer_code_prefix` 正規化為大寫後比對。
  2. `status = suspended` 時的登入阻斷在認證 middleware 處理(見 2.1.3),schema 層只存狀態。
  3. 軟刪除掛 `deleted_at` mixin。
- **錯誤處理**: schema 層無 RPC 錯誤;唯一索引衝突由 2.1.2 轉譯為 `already_exists`。
- **驗收**:
  - [ ] migration 於乾淨資料庫執行成功,部分唯一索引存在。
  - [ ] Ent code gen 後可建立 company;同 prefix 不同公司建立第二筆被拒。

## 子功能 1.1.2: Department schema

- **目標**: 建立部門主檔 Ent schema,為業務資料範圍隔離單位。
- **檔案**:
  - Create `backend/ent/schema/department.go`
  - Create 對應 migration
- **介面**: 無對外 RPC;產出 `departments` 表與 `ent.Department`。
  - 欄位:`id`、`company_id`(外鍵,必填)、`name`、`status`(enum:`active` / `suspended`)、`created_at` / `updated_at` / `deleted_at`。
- **實作邏輯**:
  1. `company_id` 外鍵關聯 `companies`;同公司內 `name` 部分唯一索引(`WHERE deleted_at IS NULL`)。
  2. 索引:`(company_id)` 供公司內部門列表查詢。
  3. 軟刪除掛 `deleted_at` mixin;部門軟刪除不連動刪除使用者(由 2.2.1 擋下仍有使用者的刪除)。
- **錯誤處理**: schema 層無 RPC 錯誤。
- **驗收**:
  - [ ] migration 成功;同公司同名部門第二筆被拒,跨公司同名允許。

## 子功能 1.1.3: User schema(員工 + 客戶帳號)

- **目標**: 建立使用者 Ent schema,同時承載員工(OAuth)與客戶帳號(帳密,一主多子)。`相依: 1.1.1、1.1.2`
- **檔案**:
  - Create `backend/ent/schema/user.go`
  - Create 對應 migration
- **介面**: 無對外 RPC;產出 `users` 表與 `ent.User`。
  - 共用欄位:`id`、`company_id`、`department_id`(客戶帳號可空)、`status`(enum:`pending` / `active` / `suspended`)、`role`(enum:`super` / `company_admin` / `dept_admin` / `staff` / `guest` / `customer` / `developer`)、`data_scope`(enum:`all` / `company` / `department` / `self`;2.9.5 注入 RLS)、`token_version`(整數,預設 0,撤銷用)、`created_at` / `updated_at` / `deleted_at`。
  - 員工欄位:`email`(公司內唯一)、`name`、`phone`、`employee_no`、`password_hash` 恆空。
  - 客戶帳號欄位:`is_customer`(bool)、`customer_id`(關聯 customers,Task 3.1 建表後加外鍵)、`account_name`(客戶內唯一)、`is_primary`(bool,每客戶恰一 true)、`password_hash`(Argon2id)、`temp_password_expires_at`(可空)、`must_change_password`(bool)、`failed_login_attempts`(整數,預設 0)、`locked_at`(可空)。
- **實作邏輯**:
  1. 部分唯一索引:員工 `(company_id, email) WHERE deleted_at IS NULL AND is_customer = false`;客戶帳號 `(customer_id, account_name) WHERE deleted_at IS NULL AND is_customer = true`。
  2. 「每客戶恰一主帳號」以部分唯一索引 `(customer_id) WHERE is_primary = true AND deleted_at IS NULL` 保證最多一個;「至少一個」由 3.1.4 建檔流程保證。
  3. `guest` 為員工審核前狀態:`status = pending` + `role = guest`,審核通過才轉正式角色(見 1.4.4)。
  4. 軟刪除掛 `deleted_at` mixin。
- **錯誤處理**: schema 層無 RPC 錯誤;唯一衝突由上層轉譯 `already_exists`。
- **驗收**:
  - [ ] migration 成功,三組部分唯一索引存在。
  - [ ] 同客戶建立第二個 `is_primary = true` 帳號被 DB 拒絕。
  - [ ] `task backend:ent:gen` 產生碼可編譯(原 Task 1.1 驗收)。

## 子功能 1.2.1: Casbin model 定義

- **目標**: 定義 RBAC with domain 授權模型,功能權限以 domain = company_id 隔離(D3)。
- **檔案**:
  - Create `backend/config/casbin_model.conf`
- **介面**: model 字串:`request = sub, dom, obj, act`;`policy = p, sub, dom, obj, act`;`role = g, _, _, _`(使用者→角色、角色繼承均在 domain 內);matcher 比對 `sub`/`dom` 相等 + `obj`/`act` keyMatch。
- **實作邏輯**:
  1. `dom` 一律帶 company_id;`super` 與 `developer` 不靠 model 特判,分別由 policy(seeder 發全域規則)與 middleware(1.11.2)處理。
  2. `obj` 命名慣例:domain 資源路徑(如 `customers`、`sales_orders`);`act` 對齊 CRUD 動詞,供 2.9.3 功能權限矩陣與 2.10 policy 管理共用詞彙。
- **錯誤處理**: model 檔語法錯誤 → enforcer 初始化失敗,啟動 fail-fast。
- **驗收**:
  - [ ] enforcer 以 model 初始化成功;`Enforce("u1","c1","customers","read")` 可依 policy 正確回應 allow/deny。

## 子功能 1.2.2: PostgreSQL adapter 初始化

- **目標**: Casbin policy 持久化於 PostgreSQL,多 replica 共享同一份規則。`相依: 1.2.1`
- **檔案**:
  - Create `backend/internal/authz/casbin.go`
  - Create 對應 migration(`casbin_rule` 表)
- **介面**: `authz.NewEnforcer(db, modelPath) (*authz.Enforcer, error)`;`Enforcer.Enforce(sub, dom, obj, act) (bool, error)`;`Enforcer.AddPolicy / RemovePolicy / ListPolicies`(供 2.10 使用)。
- **實作邏輯**:
  1. 以 pgx adapter 初始化;`casbin_rule` 表與業務表同庫,不加 RLS(內部表,僅後端存取)。
  2. `AddPolicy`/`RemovePolicy` 後呼叫 `LoadPolicy` 或開啟 auto-notify,確保多 replica 即時生效(2.10.2 完整處理;此處先保證單機正確)。
  3. enforcer 以 singleton 注入 middleware(1.6.5)與各 domain handler。
- **錯誤處理**: DB 連線失敗 → 啟動 fail-fast;`Enforce` 執行期錯誤記 log 並 deny(預設拒絕)。
- **驗收**:
  - [ ] 重啟後 policy 仍存在;新增規則後不重啟即生效。

## 子功能 1.2.3: 預設 policy seeder

- **目標**: 部署後內建五角色(super/company_admin/dept_admin/staff/customer)具備規格定義的預設功能權限。`相依: 1.2.2`
- **檔案**:
  - Create `backend/internal/authz/seed.go`
  - Create seed migration 或啟動時冪等 seeder
- **介面**: `authz.SeedDefaultPolicies(ctx, enforcer) error`(冪等)。
- **實作邏輯**:
  1. 依規格 §3.4 預設矩陣建立各角色 `p` 規則;`super` 發全域規則(dom = `*`)。
  2. seeder 冪等:以「規則不存在才新增」方式執行,重跑不產生重複列;使用者後續於 2.10 調整過的規則不被覆蓋(seeder 只補缺,不還原)。
  3. `g` 規則(使用者→角色指派)不在 seeder,由 2.3.1 使用者 CRUD 動態建立。
- **錯誤處理**: seed 失敗 → 啟動 fail-fast(無預設權限的系統不可用)。
- **驗收**:
  - [ ] 全新部署後:`staff` 無法存取他部門資源、`company_admin` 可跨部門限自己公司、`super` 可存取所有公司(原 Task 1.2 驗收)。
  - [ ] 重跑 seeder 不產生重複 policy、不覆蓋人工調整。

## 子功能 1.3.1: RLS policy migration

- **目標**: 在 PostgreSQL 層對租戶表啟用列級安全,作為授權最後防線(D3)。
- **檔案**:
  - Create `backend/database/migrations/00003_rls_policies.sql`
- **介面**: 無 RPC;對 `users` / `companies` / `departments` 等表 `ENABLE ROW LEVEL SECURITY` 並建立 policy。
- **實作邏輯**:
  1. policy 依 session variables `app.current_company_id` / `app.current_department_id` 篩選;讀寫皆套用。
  2. `app.current_data_scope = 'all'`(super/developer)或 company 級角色按 `company_id` 比對、department 級加比對 `department_id`、self(客戶)加比對 `customer_id`;data_scope 注入見 1.3.2,等級語義見 2.9.5。
  3. migration 具備 down 版本(撤 policy + DISABLE RLS),供 rollback。
- **錯誤處理**: migration 失敗即部署失敗;session variable 未注入時 policy 預設不命中任何列(fail-closed)。
- **驗收**:
  - [ ] 未注入 session variable 的連線查 `users` 回 0 列。
  - [ ] D21 RLS 整合測試:不同 company/department 連線互不可見。

## 子功能 1.3.2: session 變數注入 hook

- **目標**: 每筆請求的 DB 連線自動注入當前使用者的租戶上下文,業務程式碼不需手動設定。`相依: 1.3.1`
- **檔案**:
  - Create `backend/internal/database/rls.go`
  - Update Ent client 建立處(connection hook)
- **介面**: `rls.SetContext(ctx, identity) (context.Context, error)`;`identity` 含 `company_id` / `department_id` / `data_scope` / `customer_id`;Ent client 於取得連線、開交易時自動執行 `SET LOCAL`。
- **實作邏輯**:
  1. middleware(1.6.5)解析身分後呼叫 `rls.SetContext`,將 identity 放入 ctx。
  2. Ent driver 包一層 hook:每次從 pool 取連線或開 tx,先 `SET LOCAL app.current_company_id = ...` 等四個變數;`SET LOCAL` 僅作用於當前交易,交易外查詢以顯式交易包裝,避免連線重用殘留。
  3. 無身分的公開端點(2.4.3 公開發現、3.8 QR 兌換)注入受限的唯讀上下文或走獨立無 RLS 表,不得跳過注入沿用前一連線狀態。
- **錯誤處理**: 注入失敗(連線錯誤)→ 請求 `internal` 錯誤並記 log;不得降級為不注入繼續執行。
- **驗收**:
  - [ ] 同一 pool 連線先後服務兩家公司請求,各自僅見自己資料(無殘留洩漏)。
  - [ ] 未經 middleware 的內部呼叫若未注入,查詢 fail-closed 回 0 列。

## 子功能 1.3.3: 高權角色繞過機制

- **目標**: `super` / `developer` 可跨租戶存取,不受 RLS 阻擋。`相依: 1.3.2`
- **檔案**:
  - Update `backend/internal/database/rls.go`
  - Update `backend/database/migrations/00003_rls_policies.sql`
- **介面**: identity 增加 `bypass_rls bool`;policy 增加 `OR current_setting('app.current_data_scope') = 'all'` 類豁免條件。
- **實作邏輯**:
  1. 繞過不透過 `BYPASSRLS` 角色屬性(連線池共用同一 DB 角色無法逐請求切換),改以 policy 內判斷 `app.current_data_scope = 'all'` 放行。
  2. `super` 由 1.6.5 依 role 設定 `data_scope = 'all'`;`developer` 由 1.11.2 設定,兩者都照常寫稽核。
  3. 寫入路徑同樣套用:繞過者新增資料仍須顯式帶入目標 `company_id` / `department_id`(由 API 參數指定),不得靠豁免寫入無歸屬資料。
- **錯誤處理**: `data_scope = 'all'` 但角色非 super/developer(middleware 異常)→ 記 security log 並拒絕。
- **驗收**:
  - [ ] `super` token 可查全部公司資料;`staff` token 注入同語句僅見自己部門。

## 子功能 1.4.1: OAuth2 導向端點

- **目標**: 員工從 Web/App 發起 Google Workspace 登入,後端產生授權導向 URL。
- **檔案**:
  - Create `backend/internal/domain/auth/oauth.go`
  - Update `backend/config/config.go`(OAuth client id/secret/redirect/hd 限制)
- **介面**: `GET /api/v1/auth/oauth/:provider`(provider = `google`)→ 302 導向 Google 授權頁;query 帶 `state`(CSRF,一次性,存 Valkey,TTL 10 分鐘)與 `redirect`(登入完成回跳前端路徑,白名單校驗)。
- **實作邏輯**:
  1. 組裝 authorization URL(scope `openid email profile`,hosted domain 限定公司 Workspace 網域,多網域以設定清單允許)。
  2. `state` 與 `redirect` 綁定存 Valkey;callback 僅接受存在且未使用的 state,用完即刪。
- **錯誤處理**: 未知 provider → `not_found`;`redirect` 不在白名單 → `invalid_argument`;設定缺 client id → 啟動 fail-fast。
- **驗收**:
  - [ ] 瀏覽器訪問端點正確導向 Google;偽造 redirect 被拒。

## 子功能 1.4.2: callback 與 ID token 驗證

- **目標**: 處理 Google 回調,驗證授權碼與 ID token,識別使用者身分。`相依: 1.4.1`
- **檔案**:
  - Update `backend/internal/domain/auth/oauth.go`
  - Create `backend/internal/domain/auth/handler.go`
- **介面**: `GET /api/v1/auth/oauth/:provider/callback?code=&state=` → 依結果 302:已審核使用者 → 建立 session/JWT 回跳前端;新使用者 → 回跳註冊完成頁(帶一次性 registration token)。
- **實作邏輯**:
  1. 驗 `state`(存在、未用、未過期),以 `code` 向 Google 換 token,驗 ID token 簽章、`aud`、`exp`、hosted domain。
  2. 以 email + 所屬公司查 `users`:存在且 `status = active` → 走 1.6 發放;`pending` → 回跳「等待審核」頁;不存在 → 發一次性 registration token(Valkey,TTL 30 分鐘,綁 email),回跳註冊完成頁,**此時不建帳號**。
  3. 公司歸屬判斷:依 hd 網域對應 companies;對不到時註冊完成頁讓使用者自選公司(1.4.3)。
- **錯誤處理**: state 無效/重放 → `unauthenticated`;token 交換或驗簽失敗 → `unauthenticated`;非允許網域 → `permission_denied`。一律不透露帳號存在與否。
- **驗收**:
  - [ ] 既有 active 員工完成 callback 取得 session;重放同一 state 第二次被拒。

## 子功能 1.4.3: 首登註冊完成(建立 guest)

- **目標**: 新員工選公司 + 輸入姓名,完成後才建立 `pending` guest 帳號;未完成前系統無此帳號。`相依: 1.4.2、1.1.3`
- **檔案**:
  - Update `backend/internal/domain/auth/handler.go`
  - Create `backend/internal/domain/auth/usecase.go`
- **介面**: Connect-RPC `AuthService.CompleteRegistration(registration_token, company_id, name) → { status: "pending" }`。
- **實作邏輯**:
  1. 驗 registration token(Valkey 取出 email,用完即刪);`company_id` 須存在且 `status = active`。
  2. 同一交易:建立 `users`(email 取自 token、`name`、`role = guest`、`status = pending`、`company_id`) + audit log(註冊申請);交易外無副作用。
  3. 同 email 同公司已存在帳號(含軟刪除) → `already_exists`,引導聯繫管理員。
- **錯誤處理**: token 過期/不存在 → `unauthenticated`(重新走 OAuth);公司無效 → `invalid_argument`;重複 → `already_exists`。
- **驗收**:
  - [ ] 完成註冊前 DB 無該使用者;完成後存在 `pending` guest,登入僅見「等待審核」。

## 子功能 1.4.4: guest 審核流程

- **目標**: company_admin 審核自己公司的 pending guest,指派正式角色與部門。`相依: 1.4.3`
- **檔案**:
  - Update `backend/internal/domain/auth/usecase.go`
  - Update proto(`AuthService` 或併入 `UserService` 審核 RPC)
- **介面**: Connect-RPC `ListPendingGuests()`(僅自己公司)、`ApproveGuest(user_id, role, department_id)`、`RejectGuest(user_id, reason)`。
- **介面細節**: `ApproveGuest` 的 `role` 限 `company_admin` / `dept_admin` / `staff`;`dept_admin`/`staff` 必填 `department_id`。
- **實作邏輯**:
  1. 審核者範圍檢查:company_admin 僅見/僅審自己 `company_id` 的 pending guest(RLS + Casbin 雙層)。
  2. `ApproveGuest` 同一交易:更新 `role`/`department_id`/`status = active` + 建 Casbin `g` 規則(使用者→角色於該 company domain)+ audit log。
  3. `RejectGuest` 同一交易:直接軟刪除該 guest 帳號(`deleted_at`;不另設 rejected 狀態,審核軌跡由稽核保存)+ audit log(含原因);被拒 email 可重新走 1.4.3 註冊(部分唯一索引不擋軟刪除列)。
- **錯誤處理**: 跨公司操作 → `permission_denied`;目標非 pending → `failed_precondition`;角色不合法 → `invalid_argument`。
- **驗收**:
  - [ ] company_admin 可看到並審核自己公司的 pending guest(原 Task 1.4 驗收);A 公司 admin 不可見 B 公司 guest。
  - [ ] 核准後該員工重新登入可操作被授權功能。

## 子功能 1.5.1: 帳密驗證(一主多子)

- **目標**: 客戶以 `account_name` + 密碼登入;每客戶 1 主帳號(僅帳號管理)+ 多子帳號(業務身分),各帳號獨立停用/重置。`相依: 1.1.3`
- **檔案**:
  - Create `backend/internal/domain/auth/customer.go`
  - Update `backend/internal/domain/auth/handler.go`
- **介面**: Connect-RPC `AuthService.CustomerLogin(company_identifier, customer_code, account_name, password) → { tokens..., must_change_password, is_primary }`。
- **實作邏輯**:
  1. 依 `company_identifier` + `customer_code` 定位 customer,再以 `account_name` 找帳號;任一查無 → 統一 `unauthenticated`(不透露哪段錯)。
  2. 檢查順序:帳號 `status = active` → 公司/客戶未停用 → 鎖定狀態(1.5.3)→ 臨時密碼效期(1.5.2)→ Argon2id 比對。
  3. 主帳號(`is_primary = true`)登入成功但身分僅限帳號管理:後續呼叫業務 API 由 middleware 依 `is_primary` 回 `permission_denied`(403);子帳號為業務登入身分。
  4. 成功:清 `failed_login_attempts` / `locked_at`,走 1.6 發放 JWT/session;回傳 `must_change_password` 供前端導向改密碼頁。
  5. 成功與失敗皆寫稽核(失敗不含密碼);失敗計數更新與稽核同一交易。
- **錯誤處理**: 帳號/密碼錯誤、帳號停用、公司停用 → `unauthenticated`;鎖定中 → `failed_precondition`(附解鎖剩餘時間);臨時密碼過期 → `failed_precondition`(提示聯繫管理員)。
- **驗收**:
  - [ ] 主帳號可登入但呼叫下單 API 回 `permission_denied`;子帳號可正常下單。
  - [ ] 停用任一子帳號不影響主帳號與其他子帳號登入(原 Task 1.5 驗收)。

## 子功能 1.5.2: 臨時密碼與首登強制修改

- **目標**: 新建/重置的客戶帳號持 24 小時臨時密碼,首次登入必須改密碼才能進系統。`相依: 1.5.1`
- **檔案**:
  - Update `backend/internal/domain/auth/customer.go`
- **介面**: Connect-RPC `AuthService.ChangePassword(old_password, new_password)`(登入態,`must_change_password = true` 時唯一可用 RPC)。
- **實作邏輯**:
  1. 產生臨時密碼(3.1.4 建檔、1.5.4 重置):隨機 ≥ 12 字元,寫 `password_hash` + `temp_password_expires_at = now + 24h` + `must_change_password = true` + 清鎖定,同一交易。
  2. 登入時 `must_change_password = true` → 核發的 JWT 帶受限 claim,僅 `ChangePassword` 可通過 middleware;其餘 RPC `failed_precondition`。
  3. `ChangePassword`:驗舊密碼 → 新密碼強度 ≥ 8 字元 → 更新雜湊、清 `must_change_password` 與 `temp_password_expires_at`、`token_version + 1`(舊 session 全失效,以新密碼重登),同一交易 + 稽核。
  4. 臨時密碼過期 → 登入 `failed_precondition`,僅能由 dept_admin 以上重新產生(1.5.4)。
- **錯誤處理**: 舊密碼不符 → `unauthenticated`;強度不足 → `invalid_argument`;過期 → `failed_precondition`。
- **驗收**:
  - [ ] 未改密碼前呼叫任何業務 RPC 被拒;改完可正常操作(原 Task 1.5 驗收)。
  - [ ] 臨時密碼超過 24 小時登入被拒。

## 子功能 1.5.3: 登入鎖定(5 次 / 30 分鐘)

- **目標**: 連續 5 次密碼錯誤鎖定帳號 30 分鐘,防止離線暴力破解。`相依: 1.5.1`
- **檔案**:
  - Update `backend/internal/domain/auth/customer.go`
- **介面**: 無新增 RPC;登入回應在鎖定時附 `locked_until`。
- **實作邏輯**:
  1. 密碼錯誤:`failed_login_attempts + 1`;達 5 次設 `locked_at = now`,計數更新與稽核同一交易。
  2. 登入檢查:`locked_at` 距今 < 30 分鐘 → `failed_precondition`(附剩餘秒數);≥ 30 分鐘 → 自動解除(清 `locked_at` 與計數)後繼續驗密碼。
  3. 鎖定判定與計數更新需對同一 user 列 `SELECT ... FOR UPDATE`,避免併發登入請求同時通過第 5 次。
  4. 重置密碼(1.5.4)連帶解鎖;員工 OAuth 登入不適用本機制(Google 側防護)。
- **錯誤處理**: 鎖定中 → `failed_precondition`;其餘沿用 1.5.1。
- **驗收**:
  - [ ] 連續 5 次錯誤後正確密碼也被拒;30 分鐘後(測試以時間注入模擬)自動解除可登入(原 Task 1.5 驗收;D21 鎖定解除整合測試)。
  - [ ] 併發 10 次錯誤嘗試,最終計數與鎖定狀態一致。

## 子功能 1.5.4: 密碼重置 API

- **目標**: dept_admin 以上可為自己範圍內的客戶帳號重新產生臨時密碼。`相依: 1.5.2`
- **檔案**:
  - Update `backend/internal/domain/auth/customer.go`
  - Update proto(`AuthService.ResetCustomerPassword`)
- **介面**: Connect-RPC `AuthService.ResetCustomerPassword(user_id) → { temp_password, expires_at }`(僅此一次回傳明文,不落盤)。
- **實作邏輯**:
  1. 範圍檢查:dept_admin 限自己部門客戶、company_admin 限自己公司、super 不限;目標須 `is_customer = true`。
  2. 同一交易:產生臨時密碼並更新雜湊/效期/`must_change_password`、清鎖定、`token_version + 1`、寫稽核(操作者 + 目標帳號,不含密碼)。
  3. 回傳的明文僅存在於本次回應;交付通路(業務轉交)屬作業流程,系統不留存。
- **錯誤處理**: 範圍外目標 → `permission_denied`;目標非客戶帳號 → `invalid_argument`;目標不存在 → `not_found`。
- **驗收**:
  - [ ] 重置後舊密碼與舊 session 立即失效,新臨時密碼可登入且強制改密碼。
  - [ ] staff 呼叫被拒;稽核含操作者與目標。

## 子功能 1.6.1: scs + Valkey session store

- **目標**: Web 端以 httpOnly cookie session 為登入態,session 資料存 Valkey 供多 replica 共享。
- **檔案**:
  - Create `backend/internal/session/manager.go`
  - Update `backend/config/config.go`(Valkey 連線、session TTL)
- **介面**: `session.Manager`(scs 封裝):`Load(ctx)` / `Put(ctx, key, val)` / `Destroy(ctx)`;cookie 屬性 httpOnly、Secure、SameSite=Lax。
- **實作邏輯**:
  1. scs 以 Valkey 為 store;session 內容:user_id、role、company_id、department_id、is_customer、is_primary。
  2. session TTL 對齊 refresh token 30 天,滑動展延;登出/強制登出時 `Destroy` 刪 Valkey key。
  3. Valkey 連線失敗 → 啟動 fail-fast(session 為關鍵路徑,不降級記憶體)。
- **錯誤處理**: 執行期 Valkey 錯誤 → 請求 `unavailable` 並記 log;不得靜默視為未登入後繼續。
- **驗收**:
  - [ ] Web 登入後取得 httpOnly cookie(原 Task 1.6 驗收);重啟 backend session 仍有效(存 Valkey)。

## 子功能 1.6.2: access JWT 簽發與驗證

- **目標**: App/API 以 1 小時 access JWT 為登入態,claim 帶 `tv` 供撤銷比對。`相依: 1.1.3`
- **檔案**:
  - Update `backend/internal/session/manager.go`
  - Create `backend/internal/session/jwt.go`
- **介面**: `jwt.Issue(identity) (token string, exp time.Time)`;`jwt.Verify(token) (*Claims, error)`;Claims:`sub`(user_id)、`role`、`cid`(company_id)、`did`(department_id)、`cust`(customer_id)、`prim`(is_primary)、`tv`(token_version)、`exp`。
- **實作邏輯**:
  1. HS256 對稱簽章(金鑰自 config `JWT_SECRET`,僅後端自簽自驗;日後需跨服務驗簽再升 RS256);TTL 1 小時。
  2. 簽發時讀取當前 `users.token_version` 寫入 `tv`;驗證僅做簽章與 `exp`,撤銷比對在 1.6.4。
  3. 客戶受限態(`must_change_password`)以獨立 claim 標記,供 middleware 限制可用 RPC(1.5.2)。
- **錯誤處理**: 簽章/格式/過期 → `unauthenticated`;金鑰未設定 → 啟動 fail-fast。
- **驗收**:
  - [ ] App 請求帶 `Authorization: Bearer` 可通過驗證(原 Task 1.6 驗收);竄改 claim 被拒。

## 子功能 1.6.3: refresh token 旋轉

- **目標**: 30 天 refresh token 旋轉制:每次換發新 token 並作廢舊 token,後端僅存雜湊。`相依: 1.6.2`
- **檔案**:
  - Update `backend/internal/session/jwt.go`
  - Create 對應 migration(`refresh_tokens` 表)
- **介面**: Connect-RPC `AuthService.RefreshToken(refresh_token) → { access_token, refresh_token, expires_in }`;`refresh_tokens` 表:`id`、`user_id`、`token_hash`(唯一)、`expires_at`、`rotated_at`(可空)、`created_at`。
- **實作邏輯**:
  1. 換發流程:雜湊查表 → 驗 `expires_at`、未 rotated、user `status = active` 且 `users.token_version` 現值與發放 access JWT 之 `tv` 一致(比對現值,不在 `refresh_tokens` 表冗餘記錄,撤銷語義與 1.6.4 同源) → 同一交易標記舊列 `rotated_at`、寫新列、簽新 access JWT。
  2. 舊 token 重放(已 rotated 的 hash 再次出現)→ 視為外洩:同一交易作廢該使用者全部 refresh token + `token_version + 1` + security log。
  3. 過期列清理:換發與驗證查詢時順手 `DELETE` 該使用者已過期列(1.0 不引入背景 job 框架);表不加 RLS(內部表)。
- **錯誤處理**: token 不存在/過期/已旋轉 → `unauthenticated`;重放偵測 → `unauthenticated` 並全數撤銷。
- **驗收**:
  - [ ] access 過期後可換發,舊 refresh token 立即失效(原 Task 1.6 驗收)。
  - [ ] 重放舊 token 後,該使用者所有 refresh token 與 JWT 全失效。

## 子功能 1.6.4: token_version 撤銷比對 middleware

- **目標**: 每次請求比對 JWT `tv` 與 DB 現值,實現強制登出/改密碼後舊 token 立即失效。`相依: 1.6.2`
- **檔案**:
  - Create `backend/internal/middleware/auth.go`(比對段)
- **介面**: 併入 `middleware.Authenticate` 流程;比對失敗一律 `unauthenticated`。
- **實作邏輯**:
  1. 驗簽通過後以 `sub` 查 `users.token_version`;`tv != 現值` → 拒絕。
  2. 查詢成本控制:同請求只查一次;熱路徑可加短 TTL(≤ 60 秒)per-user 快取於 Valkey,撤銷操作(1.7.1、1.5.2、1.5.4)主動刪快取 — 快取與主動失效成對實作,不得只加快取。
  3. Web session 路徑同樣比對(session 內記發放時 tv),保證 cookie 登入態也被即時撤銷。
- **錯誤處理**: 比對不符、user 不存在/非 active → `unauthenticated`;DB 錯誤 → `unavailable`。
- **驗收**:
  - [ ] `token_version` 變更後,舊 JWT / refresh token / Web session 立即失效(原 Task 1.6 驗收)。

## 子功能 1.6.5: Authenticate middleware

- **目標**: 統一解析 Web cookie session 與 Bearer JWT,注入身分供 Casbin、RLS、handler 使用。`相依: 1.6.1、1.6.2、1.6.4`
- **檔案**:
  - Create `backend/internal/middleware/auth.go`
- **介面**: `middleware.Authenticate(sessionManager, enforcer, db)`;產出 ctx identity(user_id/role/company_id/department_id/customer_id/is_primary/data_scope),供 `rls.SetContext`(1.3.2)與各 handler。
- **實作邏輯**:
  1. 依請求類型擇一:有 cookie → session 路徑;有 `Authorization: Bearer` → JWT 路徑;兩者皆無 → `unauthenticated`。
  2. 身分確立後依序:公司 `status` 檢查(2.1.3 連鎖阻斷)→ `tv` 比對(1.6.4)→ 主帳號業務 API 限制(1.5.1:`is_primary` 且路徑屬業務 RPC → `permission_denied`)→ Casbin `Enforce`(1.2.2)→ `rls.SetContext`。
  3. 公開端點(1.4.1/1.4.2、2.4.3、3.8 兌換)以路由白名單跳過本 middleware。
  4. developer 繞過在 1.11.2 疊加於此 middleware 之後。
- **錯誤處理**: 未登入/憑證無效 → `unauthenticated`;權限不足 → `permission_denied`;公司停用 → `unauthenticated`(附公司停用碼供前端提示)。
- **驗收**:
  - [ ] 無憑證呼叫受保護 RPC 回 `unauthenticated`;合法身分可到 handler 且 RLS 已注入。

## 子功能 1.6.6: X-Api-Token(server-to-server)

- **目標**: 提供機器對機器呼叫的靜態 token 驗證,僅限 server-to-server,不開放一般客戶端。
- **檔案**:
  - Update `backend/internal/middleware/auth.go`
  - Update `backend/config/config.go`(token 清單與對應身分/範圍)
- **介面**: `middleware.ApiTokenAuthenticate()` 驗證 header `X-Api-Token`;通過後注入預先設定的機器身分(固定 role/company/data_scope)。
- **實作邏輯**:
  1. token 自 config 載入(值僅存雜湊比對);每組 token 綁定預設身分與允許的 RPC 白名單。
  2. 機器身分不走 `token_version`(無 users 列),但仍過 Casbin 與 RLS 注入;呼叫寫稽核(actor 標記 api-token 名稱)。
  3. 同時帶 cookie/JWT 與 `X-Api-Token` 時以使用者憑證優先,避免降級混淆。
- **錯誤處理**: token 無效/未設定 → `unauthenticated`;白名單外 RPC → `permission_denied`。
- **驗收**:
  - [ ] 合法 token 可呼叫白名單 RPC;同 token 呼叫其他 RPC 被拒(原 Task 1.6 驗收)。

## 子功能 1.7.1: 強制登出核心(session 刪除 + token_version+1)

- **目標**: 單一入口讓指定使用者所有登入態(Web session、JWT、refresh token)立即失效。`相依: 1.6.1、1.6.4`
- **檔案**:
  - Update `backend/internal/session/manager.go`
  - Update `backend/internal/domain/users/usecase.go`
- **介面**: usecase `ForceLogoutUser(ctx, target_user_id, actor) error`(內部函式,供 1.7.2 RPC 與日後其他觸發點重用)。
- **實作邏輯**:
  1. 同一交易:`users.token_version + 1` + 作廢全部 refresh token(標記 rotated 或刪列)+ audit log;Valkey session key 刪除在交易提交後執行(外部儲存,失敗僅記 log 重試一次 — `tv` 已 +1,session 殘留不影響撤銷語義,因 1.6.4 對 Web session 也比對)。
  2. 主動刪 1.6.4 的 per-user 快取。
  3. 客戶主帳號被強制登出不連動子帳號(各帳號獨立,D22 同構);公司停用(2.1.3)走 status 阻斷,不在此功能。
- **錯誤處理**: 目標不存在 → `not_found`;DB 錯誤 → `internal`。
- **驗收**:
  - [ ] 執行後該使用者 Web cookie 與 JWT 下次請求皆 `unauthenticated`(原 Task 1.7 驗收)。

## 子功能 1.7.2: ForceLogout API + 稽核

- **目標**: super / company_admin / dept_admin 可於管理介面觸發強制登出。`相依: 1.7.1`
- **檔案**:
  - Update proto(`UserService.ForceLogout`)
  - Update `backend/internal/domain/users/usecase.go`
- **介面**: Connect-RPC `UserService.ForceLogout(user_id) → {}`。
- **實作邏輯**:
  1. 範圍檢查:dept_admin 限自己部門 staff/客戶帳號、company_admin 限自己公司、super 不限;不得對自己呼叫(避免自鎖,改用一般登出)。
  2. 呼叫 1.7.1 核心;稽核含操作者、目標、時間(D18 同事務)。
- **錯誤處理**: 範圍外 → `permission_denied`;對自己呼叫 → `invalid_argument`;其餘沿用 1.7.1。
- **驗收**:
  - [ ] dept_admin 可強登自己部門帳號、不可強登他部門;稽查 log 可查到操作記錄。

## 子功能 1.8.1: 內建預設規則產生 CASL JSON

- **目標**: 前端取得當前使用者的 CASL ability JSON,控制選單與按鈕顯示(Phase 1 先以內建規則;Phase 2 `2.9.4` 起改由 `role_permissions` 表驅動,屆時本子功能僅保留序列化層)。
- **檔案**:
  - Create `backend/internal/domain/auth/ability.go`
  - Update proto(`AbilityService`)
- **介面**: Connect-RPC `AbilityService.GetAbility() → { rules: [{action, subject, conditions?, inverted?}] }`(CASL.js 可消費的 JSON)。
- **實作邏輯**:
  1. Phase 1:依 role(內建五角色)從硬編碼預設矩陣產生規則;conditions 帶 `company_id`/`department_id`/`customer_id` 讓前端可做資料範圍判斷。
  2. 規則詞彙(action/subject)與 Casbin `act`/`obj`(1.2.1)同源,避免前後端權限語意分歧;主帳號額外附加「僅帳號管理」反向規則。
  3. 回應可快取(短 TTL 60 秒,對齊前端 ability store);2.9.4 接管後改為讀 `role_permissions` 並於該表異動時失效。
- **錯誤處理**: 未登入 → `unauthenticated`;未知角色 → 回空規則(前端預設全隱藏,fail-closed)。
- **驗收**:
  - [ ] 前端可取得 ability 並正確控制按鈕顯示(原 Task 1.8 驗收);主帳號僅見帳號管理入口。

## 子功能 1.11.1: 開發者開關與啟動防護

- **目標**: `DEVELOPER_ACCOUNT_ENABLED` 僅開發/測試可用;production 誤開時後端拒絕啟動。
- **檔案**:
  - Update `backend/config/config.go`
  - Update `backend/cmd/server/main.go`(啟動檢查)
- **介面**: config 欄位 `DeveloperAccountEnabled bool`(development 預設 true、test 預設 true、production 預設 false)。
- **實作邏輯**:
  1. 啟動時檢查:`ENV = production` 且開關為 true → 印出明確錯誤並 fail-fast(原 Task 1.11 驗收)。
  2. 開關狀態寫入啟動 log(不含任何憑證),供部署檢查(8.7 上線清單會驗 `DEVELOPER_ACCOUNT_ENABLED=false`)。
- **錯誤處理**: 上述情境直接拒絕啟動,無降級。
- **驗收**:
  - [ ] `ENV=production` + 開關 true → 程序退出且 log 指出原因;production + false 正常啟動。

## 子功能 1.11.2: developer middleware 繞過

- **目標**: developer 角色且開關啟用時跳過 Casbin 並以 `data_scope = all` 注入 RLS,可跨公司/部門存取所有 API。`相依: 1.6.5、1.3.3、1.11.1`
- **檔案**:
  - Create `backend/internal/middleware/developer.go`
- **介面**: 疊加於 `Authenticate` 之後的 middleware:身分 `role = developer` 且開關啟用 → 設定 identity `data_scope = all`、`bypass_casbin = true`。
- **實作邏輯**:
  1. 繞過只發生在「開關啟用 + 身分為 developer」同時成立;開關關閉時 developer 帳號於登入階段即被拒(1.11.3 seed 不建/登入擋下)。
  2. 繞過 Casbin 但**不繞過稽核**:developer 操作照常寫 `audit_logs`(原 Task 1.11 要求),actor 標記 developer。
  3. 寫入操作仍須依 1.3.3 規則顯式帶入目標 `company_id`/`department_id`。
- **錯誤處理**: 開關關閉下出現 developer 身分 → `unauthenticated` 並記 security log(異常狀態)。
- **驗收**:
  - [ ] 開發環境 developer 可跨公司存取所有 API;關閉開關後 developer 無法登入(原 Task 1.11 驗收)。

## 子功能 1.11.3: developer 角色與帳號 seed

- **目標**: 種入 `developer` 系統角色與開發環境預設開發者帳號。`相依: 1.2.3、1.11.1`
- **檔案**:
  - Update `backend/internal/authz/seed.go`(角色)
  - Create seed migration 或開發環境 seeder(帳號)
- **介面**: `developer` 角色列於 `users.role` enum(1.1.3 已含);`is_system = true`、`data_scope = all`(2.9.1 的 roles 表落地時同步)。
- **實作邏輯**:
  1. 角色 seed 冪等且標記 `is_system`,2.9.2 的角色 CRUD 不得刪除/改名。
  2. 開發者帳號僅 `ENV = development` 時 seed(固定帳號名,密碼自環境變數,無預設值 — 未設定則略過並記 log);test/production 不 seed。
  3. 該帳號 `company_id` 指向 seed 用的開發用公司(若無則建立標記性開發公司),避免無歸屬資料。
- **錯誤處理**: 非 development 環境嘗試 seed 帳號 → 略過並記 log;角色 seed 失敗 → 啟動 fail-fast。
- **驗收**:
  - [ ] development 啟動後可用開發者帳號登入;test/production DB 無該帳號。

---

## 整合測試重點(D21)

1. **認證全流程**:OAuth 首登 → 註冊完成 → 審核 → 登入操作;客戶臨時密碼 → 強制改密碼 → 正常操作。
2. **RLS 隔離**:同語句在不同 company/department/self 身分的可見性;連線池重用無殘留(1.3.1/1.3.2)。
3. **鎖定解除**:5 次鎖定、30 分鐘自動解除、重置連帶解鎖、併發嘗試一致性(1.5.3)。
4. **撤銷即時性**:強制登出/改密碼後,JWT、refresh token、Web session 三路立即失效(1.6.4/1.7.1)。
5. **refresh 旋轉**:換發作廢舊 token、重放偵測全數撤銷(1.6.3)。
6. **developer 防護**:production 誤開 fail-fast;開關關閉無法登入;繞過期間稽核完整(1.11.x)。

---

*最後更新:2026-08-17*
