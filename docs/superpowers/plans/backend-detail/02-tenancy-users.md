---
title: "Backend 細部計畫 02 — 多租戶主檔:公司/部門/使用者、角色權限、Casbin Policy 管理"
tasks: [2.1, 2.2, 2.3, 2.4, 2.9, 2.10]
spec_sections: ["3.1 租戶層級", "3.1.1 公司識別與公開資訊", "3.2 角色定義", "3.3 後端授權(Casbin)", "3.4 前端權限(CASL)", "4.1 員工/管理員", "4.3 Session/Token 儲存", "5.1 核心實體", "5.2 關鍵欄位", "5.4 軟刪除慣例", "17.3 預留設計"]
depends_on: ["01-auth.md(Task 1.1 users schema、Task 1.2 Casbin、Task 1.3 RLS hook、Task 1.6 session/JWT、Task 1.7 強制登出、Task 1.8 GetAbility、Task 1.11 developer 角色)"]
decisions: [D3, D4, D5, D6, D7, D8, D9, D10, D17, D18, D20, D21]
---

# 02 — 多租戶主檔:公司 / 部門 / 使用者 / 角色權限 / Policy 管理

> 對應原計畫 Phase 2 的 Task 2.1、2.2、2.3、2.4、2.9、2.10(Task 2.5、2.6 見 `03-metadicts-audit.md`;Task 2.7、2.8、2.11 為前端,不在本文件)。
> 子功能編號可雙向追溯原計畫 Task;每個子功能六欄(目標/檔案/介面/實作邏輯/錯誤處理/驗收)即為該單元的 done 定義。

## 共通規則(各子功能引用、不重複)

1. **軟刪除(D10)**:`companies` / `departments` / `users`(僅客戶帳號)/ `roles`(僅自訂角色)統一 `deleted_at` + 部分唯一索引(`WHERE deleted_at IS NULL`);列表預設排除已刪除;復原 = 清 `deleted_at` + 寫稽核;員工帳號不刪除、僅停用(規格 4.1);內建角色 `is_system = true` 不可刪除(規格 5.4)。
2. **多租戶(D3)**:業務資料帶 `company_id` / `department_id`;Casbin 管功能(domain = `company_id`)、RLS 管資料範圍(`data_scope` 等級,不依角色名稱)、CASL 管 UI;RLS session 變數注入為最後防線,usecase 層仍須自行加查詢條件。
3. **交易與稽核(D18)**:主檔異動 + `audit_logs`、角色變更 + `token_version + 1`、policy 異動 + enforcer 更新,皆同一 DB 交易,同成功同失敗;稽核記錄 `before_snapshot` / `after_snapshot`。
4. **錯誤約定**:統一 Connect code — `unauthenticated`(登入失效、公司停用後的已登入請求)、`permission_denied`(角色/範圍不符)、`not_found`(不存在或已軟刪除)、`failed_precondition`(狀態不允許、防鎖死阻擋、樂觀鎖衝突)、`invalid_argument`(格式驗證失敗)、`already_exists`(唯一約束衝突)。
5. **company_admin 範圍**:凡「管理所屬公司」的操作,usecase 入口先比對操作者 `company_id` 與目標資源 `company_id`,不符即 `permission_denied`,RLS 兜底。

---

## Task 2.1 公司管理 API

### 子功能 2.1.1: 公司 CRUD API

- **目標**: 提供公司主檔的列表、查詢、建立、更新、軟刪除;`super` 可建立/停用/刪除,`company_admin` 僅能查看與編輯自己公司。
- **檔案**:
  - Create: `backend/internal/domain/companies/handler.go`(Connect handler)
  - Create: `backend/internal/domain/companies/usecase.go`
  - Create: `backend/internal/domain/companies/repository.go`(Ent)
  - Update: `backend/proto`(CompanyService proto 定義,依 Phase 0 proto 目錄慣例)
- **介面**: Connect-RPC `CompanyService`
  - `ListCompanies`(分頁參數、status 篩選、關鍵字)→ 公司陣列 + 總數
  - `GetCompany`(company_id)→ 公司完整欄位
  - `CreateCompany`(name、tax_id、display_name、identifier、customer_code_prefix 等)→ 新公司
  - `UpdateCompany`(company_id + 可變欄位集合,含 status)→ 更新後公司
  - `DeleteCompany`(company_id)→ 空回應(軟刪除)
- **實作邏輯**:
  1. handler 解析請求後進 usecase;usecase 入口先做角色判斷:Create/Delete 與 status 變更僅 `super`;`company_admin` 僅能 Update 自己公司且不允許改 status;其餘角色一律拒絕。
  2. List:`super` 見全部;`company_admin` 強制過濾為自己公司(忽略其篩選參數中的公司條件);其他角色不開放此 RPC。
  3. Create:必填 `name`、`identifier`;`identifier` 與 `customer_code_prefix` 的唯一性規則見 2.1.2;初始 status 為 active。
  4. Update:僅更新請求中出現的欄位;`identifier` 建立後不可修改(見 2.1.2);status 變更為非 active 時觸發 2.1.3 的連鎖行為。
  5. Delete:軟刪除(設 `deleted_at`),不阻擋該公司既有歷史資料顯示;軟刪除後同名公司可重建(部分唯一索引保證)。
  6. 交易邊界:Create / Update / Delete 的資料變更與 `audit_logs`(action = create / update / delete、resource_type = company)同一交易。
  7. 所有讀取預設排除 `deleted_at IS NOT NULL`。
- **錯誤處理**:
  - 非 `super` 呼叫 Create/Delete → `permission_denied`
  - `company_admin` 嘗試讀寫其他公司或變更 status → `permission_denied`
  - company_id 不存在或已軟刪除 → `not_found`
  - Update 嘗試修改 `identifier` → `invalid_argument`
  - 唯一性衝突(identifier / customer_code_prefix)→ `already_exists`(細節見 2.1.2)
- **驗收**:
  - [ ] `super` 可建立、更新、停用、軟刪除公司,列表與單筆查詢正確。
  - [ ] `company_admin` 僅能查看/編輯自己公司,對其他公司操作回 `permission_denied`。
  - [ ] 軟刪除後列表不再出現,但同名/同 identifier 公司可重建。
  - [ ] 每筆異動同事務寫入 `audit_logs`,含前後快照。

### 子功能 2.1.2: identifier 與 customer_code_prefix 唯一性

- **目標**: 保證 `identifier`(穩定公司代號,D20)與 `customer_code_prefix`(客戶編號前綴,D7)的格式與全系統唯一性。相依: 2.1.1
- **檔案**:
  - Update: `backend/internal/domain/companies/usecase.go`(校驗邏輯)
  - Update: `backend/database/migrations/000XX_companies.sql`(唯一索引)
- **介面**: 無新增 RPC;為 2.1.1 Create/Update 的欄位驗證規則。
  - `companies.identifier`:全系統唯一,部分唯一索引(`WHERE deleted_at IS NULL`)。
  - `companies.customer_code_prefix`:大寫英數 1–4 碼,全系統唯一,部分唯一索引,允許 NULL(未設定前不可建立客戶編號,Phase 3 取號時檢查)。
- **實作邏輯**:
  1. `identifier` 驗證:非空、URL-safe(小寫英數與連字號),因作為公開端點路徑與跨協定穩定代號(規格 17.3),建立後不可修改。
  2. `customer_code_prefix` 驗證:符合大寫英數 1–4 碼格式,不符合回 `invalid_argument`;允許由 NULL 設定為有效值,也允許修改。
  3. 唯一性檢查分兩層:usecase 先 SELECT 檢查(給出明確錯誤訊息),DB 部分唯一索引為最終保證;索引違反由 repository 轉譯為 `already_exists`。
  4. 修改 `customer_code_prefix` 僅影響後續新編碼:不回頭改既有 `customers.customer_code`、不重置 `customer_counters`(D7);usecase 不得出現任何對既有客戶編號或計數器的更新。
  5. 軟刪除公司後其 identifier / prefix 釋出,可供新公司使用(部分唯一索引語意)。
  6. 交易邊界:唯一性檢查與公司寫入同一交易,避免併發建立同值;DB 唯一索引捕捉競態漏網者。
- **錯誤處理**:
  - prefix 格式不符(非大寫英數、長度超過 4 或為 0)→ `invalid_argument`
  - identifier 格式不符 → `invalid_argument`
  - identifier 已被其他未刪除公司使用 → `already_exists`
  - prefix 已被其他未刪除公司使用 → `already_exists`
- **驗收**:
  - [ ] 建立重複 identifier 或 prefix 回 `already_exists`,錯誤訊息指明欄位。
  - [ ] prefix 大小寫/長度不符回 `invalid_argument`。
  - [ ] 修改 prefix 後既有客戶編號不變、`customer_counters` 不重置。
  - [ ] 軟刪除公司後其 identifier / prefix 可被新公司重用。

### 子功能 2.1.3: 公司停用連鎖登入阻斷

- **目標**: 公司 status 非 active 時,該公司所有帳號(員工 + 客戶)無法登入,已登入 session / token 於下次請求失效(規格 3.1)。相依: 2.1.1;機制掛接 `01-auth.md` Task 1.6 的認證 middleware
- **檔案**:
  - Update: `backend/internal/auth`(認證 middleware,加入 company status 檢查)
  - Update: `backend/internal/domain/companies/usecase.go`(status 變更流程)
- **介面**: 無新增 RPC;影響所有需認證端點與登入端點的行為。
- **實作邏輯**:
  1. 認證 middleware 在驗證 session(Web)或 JWT(App)通過後,取出使用者 `company_id`,查詢該公司 status;查詢與既有 `token_version` 比對同一資料庫往返內完成,不額外快取,保證停用即時生效。
  2. 公司 status 非 active:已登入請求一律回 `unauthenticated`(Web 端同時清除 session cookie),讓前端導回登入頁;App 端 JWT 失效,需重新登入。
  3. 登入端點(OAuth2 callback 建 session、客戶帳密登入、QR token 兌換)在建 session/發 token 前檢查公司 status,非 active 回 `permission_denied` 並註明公司已停用,不建立任何憑證。
  4. `developer` 帳號不受公司停用阻斷(其繞過語意見 D8,且開發者帳號歸屬公司停用不應鎖住開發工作);其餘角色無例外。
  5. 公司停用本身 = 2.1.1 的 Update status;status 變更 + `audit_logs` 同一交易(D18)。
  6. 公司恢復 active 後,既有帳號可正常登入;停用期間不主動刪除 session 資料,由 middleware 逐請求阻擋。
- **錯誤處理**:
  - 已登入請求所屬公司已停用 → `unauthenticated`
  - 停用公司帳號嘗試登入 → `permission_denied`(公司已停用)
  - 公司 status 查詢期間公司已被軟刪除 → 視同停用,回 `unauthenticated`
- **驗收**:
  - [ ] 公司停用後,該公司員工 Web session 下一次請求即失效並導回登入頁。
  - [ ] 公司停用後,該公司客戶 App JWT 請求回 `unauthenticated`。
  - [ ] 停用公司任何帳號登入回 `permission_denied`,不核發憑證。
  - [ ] 恢復 active 後帳號可正常登入;`developer` 帳號不受影響。

---

## Task 2.2 部門管理 API

### 子功能 2.2.1: 部門 CRUD API

- **目標**: 提供部門主檔的列表、查詢、建立、更新、軟刪除;`super` 全域管理,`company_admin` 管理所屬公司部門。
- **檔案**:
  - Create: `backend/internal/domain/departments/handler.go`
  - Create: `backend/internal/domain/departments/usecase.go`
  - Create: `backend/internal/domain/departments/repository.go`(Ent)
  - Update: `backend/proto`(DepartmentService proto 定義)
- **介面**: Connect-RPC `DepartmentService`
  - `ListDepartments`(分頁、company_id 篩選)→ 部門陣列 + 總數
  - `GetDepartment`(department_id)→ 部門完整欄位
  - `CreateDepartment`(company_id、name)→ 新部門
  - `UpdateDepartment`(department_id + 可變欄位)→ 更新後部門
  - `DeleteDepartment`(department_id)→ 空回應(軟刪除)
- **實作邏輯**:
  1. 範圍判斷:`super` 可對任意公司建立/管理;`company_admin` 所有操作限定自己 `company_id` — Create 時請求 `company_id` 必須等於操作者公司,List 強制過濾自己公司,Get/Update/Delete 先查目標部門 `company_id` 再比對。
  2. 部門為租戶層級最小管理單位,業務實體皆掛 `department_id`;部門本身僅 `company_id` 一層租戶屬性。
  3. Delete:軟刪除;不阻擋歷史訂單/單據顯示該部門名稱(規格 5.4);軟刪除部門不可再被選用於新建資料,其下使用者後續操作由 RLS `department_id` 比對自然失效,不額外連鎖停用帳號。
  4. 交易邊界:Create / Update / Delete 與 `audit_logs`(resource_type = department)同一交易。
  5. 列表預設排除 `deleted_at IS NOT NULL`;管理介面「顯示已刪除」由查詢參數控制。
- **錯誤處理**:
  - `company_admin` 對其他公司部門操作 → `permission_denied`
  - 無管理角色(如 `staff`)呼叫寫入 RPC → `permission_denied`
  - department_id 不存在或已軟刪除 → `not_found`
  - Create 缺 `company_id` / `name` → `invalid_argument`
- **驗收**:
  - [ ] `super` 可對任意公司建立/更新/軟刪除部門。
  - [ ] `company_admin` 可管理所屬公司部門,對他公司部門回 `permission_denied`。
  - [ ] 軟刪除部門不再出現於列表與選單,歷史資料仍可顯示其名稱。
  - [ ] 異動皆同事務寫入 `audit_logs`。

---

## Task 2.3 使用者管理 API

### 子功能 2.3.1: 使用者 CRUD + 角色指派

- **目標**: 提供使用者帳號的列表、查詢、建立(審核)、更新與角色指派;員工帳號來源為 OAuth 註冊完成後的 `guest` 審核(D6),客戶帳號由 Phase 3 客戶主檔流程建立,本 RPC 管理兩者。相依: `01-auth.md` Task 1.1(users schema)
- **檔案**:
  - Create: `backend/internal/domain/users/handler.go`
  - Create: `backend/internal/domain/users/usecase.go`
  - Create: `backend/internal/domain/users/repository.go`(Ent)
  - Update: `backend/proto`(UserService proto 定義)
- **介面**: Connect-RPC `UserService`
  - `ListUsers`(分頁、company_id / department_id / role / status 篩選)→ 使用者陣列 + 總數
  - `GetUser`(user_id)→ 使用者完整欄位(不含 `password_hash`)
  - `CreateUser`(姓名、company_id、department_id、角色)→ 新使用者(員工帳號,供 super 直接建立情境)
  - `UpdateUser`(user_id + 姓名 / department_id / phone / employee_no 等)→ 更新後使用者
  - `AssignRole`(user_id、role_code)→ 更新後使用者(角色指派,含 guest 審核:status pending → active 並指派部門與角色)
  - `Deactivate`(user_id)、`ForceLogout`(user_id)見 2.3.2 / 2.3.3
- **實作邏輯**:
  1. 範圍判斷統一在 usecase 入口:`super` 全域;`company_admin` 限自己公司;`dept_admin` 限自己部門且目標帳號角色僅能為 `staff`(見 2.3.2)。
  2. List:依操作者範圍強制注入 `company_id` / `department_id` 過濾條件;RLS 兜底。
  3. Create(員工帳號):僅建立帳號主檔與指派,不產生密碼(員工走 OAuth2,`password_hash` 為 NULL);客戶帳號不由本 RPC 建立(Phase 3 客戶建立流程自動產生主帳號 + 業務子帳號,D22/D28)。
  4. AssignRole(角色指派/審核):
     a. 查目標使用者,確認在操作者管理範圍內。
     b. `guest` 審核:僅 `super` 或該公司 `company_admin` 可執行(`dept_admin` 不審核 guest,規格 4.1);指派部門與角色後 status 由 pending 轉 active。
     c. 角色異動於同一交易內完成三件事:更新使用者角色關聯、更新 Casbin g 規則(移除舊 `g, user, 舊role, domain`、新增 `g, user, 新role, domain`)、`users.token_version + 1`(舊 token 立即失效,D5)。
     d. 角色變更寫 `audit_logs`(action = role_change),同事務。
  5. 交易邊界:所有寫入 RPC 的資料變更 + g 規則異動 + token_version + audit 同一 DB 交易;Casbin enforcer 記憶體態於交易提交後更新,提交失敗則不更新。
  6. 員工帳號不可刪除(無 Delete RPC);客戶帳號軟刪除由 Phase 3 客戶帳號管理提供,不在本 RPC。
- **錯誤處理**:
  - 操作目標超出管理範圍(跨公司、跨部門、`dept_admin` 操作非 staff 帳號)→ `permission_denied`
  - `dept_admin` 嘗試審核 guest → `permission_denied`
  - user_id 不存在 → `not_found`
  - 指派的 role_code 不存在或已停用 → `invalid_argument`
  - AssignRole 對 pending 以外狀態執行審核流程 → `failed_precondition`
- **驗收**:
  - [ ] `super` 可管理所有使用者;`company_admin` 限自己公司;`dept_admin` 限自己部門 staff。
  - [ ] guest 審核後帳號轉 active 並取得指派部門與角色,可用新角色登入。
  - [ ] 角色變更後舊 JWT/session 立即失效(token_version + 1),Casbin g 規則同步更新。
  - [ ] 角色變更寫入 `audit_logs`(role_change)且與異動同事務。

### 子功能 2.3.2: 停用與管理範圍控制

- **目標**: 實作 `Deactivate` 停用帳號,並以一致規則落實三級管理範圍(super 全域 / company_admin 公司 / dept_admin 部門 staff)。相依: 2.3.1
- **檔案**:
  - Update: `backend/internal/domain/users/usecase.go`(範圍判斷共用函式)
  - Update: `backend/internal/domain/users/handler.go`
- **介面**: Connect-RPC `UserService.Deactivate`(user_id)→ 空回應;範圍判斷規則適用 UserService 全部寫入 RPC。
- **實作邏輯**:
  1. 管理範圍共用判斷(供 2.3.1 / 2.3.2 / 2.3.3 複用):
     a. `super`:不設限。
     b. `company_admin`:目標使用者 `company_id` 必須等於操作者公司。
     c. `dept_admin`:目標使用者 `company_id` 與 `department_id` 皆須等於操作者,且目標角色必須為 `staff`(規格 3.2:dept_admin 僅管理該部門內 staff 帳號)。
     d. 任何不符 → `permission_denied`;RLS(`app.current_company_id` / `app.current_department_id` / `app.current_data_scope`)為最後防線,即使 usecase 漏判也不越界。
  2. Deactivate 流程(同一交易):
     a. 範圍判斷通過後,status 設為 inactive。
     b. `users.token_version + 1`,舊 JWT / refresh token 立即失效(D5)。
     c. 刪除該使用者的 Web session(Valkey)。
     d. 寫 `audit_logs`(action = update,resource_type = user,快照含 status 變更)。
  3. 邊界規則:已停用帳號重複停用 → `failed_precondition`;停用不影響同客戶其他帳號(一主多子各自獨立,D22)。
  4. 重新啟用走 2.3.1 `UpdateUser` 將 status 改回 active(同範圍控制)。
- **錯誤處理**:
  - 目標超出管理範圍 → `permission_denied`
  - user_id 不存在 → `not_found`
  - 帳號已為 inactive → `failed_precondition`
  - 未認證呼叫 → `unauthenticated`
- **驗收**:
  - [ ] 停用後該帳號所有 token 與 session 立即失效,無法再呼叫任何 API。
  - [ ] `company_admin` 停用他公司帳號、`dept_admin` 停用他部門或非 staff 帳號,皆回 `permission_denied`。
  - [ ] 停用同客戶其中一個帳號,不影響其他帳號登入。
  - [ ] 停用異動與稽核同事務。

### 子功能 2.3.3: 強制登出(ForceLogout)範圍銜接

- **目標**: `UserService.ForceLogout` 的機制、交易邊界、錯誤處理與驗收**統一由 `01-auth.md` 1.7.1 / 1.7.2 定義**,本文件不重複規格;此處僅補充 Task 2.3 範圍內的銜接點。相依: 2.3.2;`01-auth.md` 1.7
- **檔案**: 同 1.7.2(`backend/internal/domain/users/usecase.go`、proto `UserService`)。
- **介面**: 同 1.7.2,Connect-RPC `UserService.ForceLogout(user_id)`。
- **實作邏輯**:
  1. 範圍判斷使用 2.3.2 的管理範圍函式(dept_admin 限自己部門、company_admin 限自己公司、super 不限),與 1.7.2 的範圍檢查為同一實作。
  2. 被強制登出者帳號 `status` 不變、可立即重新登入;如需禁止登入,使用 2.3.2 的停用(Deactivate),兩者語義不同。
- **錯誤處理**: 同 1.7.2;範圍外目標 → `permission_denied`。
- **驗收**:
  - [ ] dept_admin 可強登自己部門帳號、不可強登他部門(與 1.7.2 驗收一致,整合測試只跑一次,兩文件交叉引用)。

---

## Task 2.4 公司識別與公開資訊 API

### 子功能 2.4.1: Logo 上傳關聯 file_assets

- **目標**: 提供公司 Logo 上傳,落本地儲存並建立 `file_assets` 記錄,更新 `companies.logo_url`。相依: 2.1.1;儲存與白名單規則依 D17
- **檔案**:
  - Create: `backend/internal/domain/companies/branding.go`
  - Update: `backend/internal/domain/files`(file_assets repository,若 Phase 3 檔案 domain 尚未建立,先於此建立最小寫入能力)
- **介面**: REST `POST /api/v1/companies/{company_id}/logo`(multipart 上傳;REST 保留用於檔案上傳,D4)→ 回傳 file_asset_id 與 url。
- **實作邏輯**:
  1. 權限:僅 `super` 可上傳/更換 Logo(規格 3.1.1);`company_admin` 可編輯其他識別欄位但 Logo 上傳回 `permission_denied`。
  2. 白名單雙重檢查(D17):僅接受 jpeg / png / webp 且 ≤ 5 MB;先驗副檔名,再讀檔頭 magic bytes 驗證實際格式,兩者皆通過才落碟;任一不符回 `invalid_argument`。
  3. 檔案寫入本地儲存路徑後,開 DB 交易:建立 `file_assets` 記錄(`owner_type` = company、`owner_id` = company_id、`storage_path`、`url`、`mime_type`、`size_bytes`、`created_by`)+ 更新 `companies.logo_url` + 寫 `audit_logs`;交易失敗時清除已落碟檔案,避免孤兒檔。
  4. 舊 Logo 的 `file_assets` 記錄保留(歷史可查),`logo_url` 指向新檔;1.0 不主動清除舊檔。
  5. `company_id` 必須存在且未軟刪除,否則 `not_found`。
- **錯誤處理**:
  - 非 `super` 上傳 → `permission_denied`
  - 未認證 → `unauthenticated`
  - 格式不符、magic bytes 不符、超過 5 MB → `invalid_argument`
  - company_id 不存在 → `not_found`
- **驗收**:
  - [ ] `super` 上傳合法圖檔後 `companies.logo_url` 更新,`file_assets` 有對應記錄且可經 url 下載。
  - [ ] 偽造副檔名(如 exe 改名 png)被 magic bytes 檢查擋下。
  - [ ] `company_admin` 上傳 Logo 回 `permission_denied`。
  - [ ] DB 交易失敗時不殘留孤兒檔案。

### 子功能 2.4.2: UpdateBranding / UpdatePublicInfo

- **目標**: 提供公司識別(簡稱、主色)與公開資訊(聯絡方式、條款連結、capabilities)的更新 RPC。相依: 2.1.1
- **檔案**:
  - Update: `backend/internal/domain/companies/branding.go`
  - Update: `backend/internal/domain/companies/usecase.go`
- **介面**: Connect-RPC `CompanyService`
  - `UpdateBranding`(company_id、display_name、primary_color;logo 由 2.4.1 上傳後此處僅接受既有 file_asset 的 url)→ 更新後公司
  - `UpdatePublicInfo`(company_id、public_email、public_phone、public_address、terms_url、privacy_url、`public_info` JSON、`capabilities` JSON)→ 更新後公司
- **實作邏輯**:
  1. 權限:`super` 全域;`company_admin` 僅限自己公司(比對 company_id);`capabilities` 為 AI 協定預留欄位(規格 17.3),1.0 僅 `super` 可維護,`company_admin` 提交 capabilities 欄位回 `permission_denied`。
  2. 欄位驗證:`primary_color` 須為 hex 色碼格式;`terms_url` / `privacy_url` / `public_email` 做基本格式驗證;`public_info` / `capabilities` 須為合法 JSON 物件;不符回 `invalid_argument`。
  3. 僅更新請求中出現的欄位,未出現欄位保持原值(JSON 欄位整體覆寫,不做深層 merge)。
  4. 交易邊界:更新 + `audit_logs`(resource_type = company)同一交易;顯示場景(登入頁、側邊欄、App、單據表頭)由讀取端即時反映,不需額外失效機制。
- **錯誤處理**:
  - `company_admin` 修改其他公司 → `permission_denied`
  - `company_admin` 提交 capabilities → `permission_denied`
  - 無管理角色呼叫 → `permission_denied`
  - 欄位格式不符 → `invalid_argument`
  - company_id 不存在 → `not_found`
- **驗收**:
  - [ ] `super` / `company_admin` 可更新自己範圍公司的識別與公開資訊。
  - [ ] capabilities 僅 `super` 可改。
  - [ ] 非法色碼、非法 URL、非法 JSON 回 `invalid_argument`。
  - [ ] 異動寫入 `audit_logs` 且同事務。

### 子功能 2.4.3: 公開發現端點

- **目標**: 提供無認證的公開端點,依 `identifier` 回傳公司公開資訊與 capabilities,供登入頁識別與未來 AI Agent 發現(D20、規格 17.3)。相依: 2.1.2
- **檔案**:
  - Create: `backend/internal/domain/companies/public.go`
- **介面**: REST `GET /api/v1/companies/public/{identifier}`(無需認證,D4)→ 回傳 display_name、logo_url、primary_color、public_email、public_phone、public_address、terms_url、privacy_url、`public_info`、`capabilities`。
- **實作邏輯**:
  1. 以 `identifier` 查詢未軟刪除且 status 為 active 的公司;查無資料回 `not_found`(不區分「不存在」與「已停用」,避免列舉探測)。
  2. 僅回傳白名單欄位(上列公開欄位);內部欄位(`tax_id` 之外的內部管理欄位、`customer_code_prefix`、status 等)一律不輸出。
  3. 不掛認證 middleware,但保留全域 middleware(Logger、Recover、CORS);加回應快取標頭供短時間快取,降低公開端點被查詢放大。
  4. 本端點為未來 `.well-known/agent-manifest` 等協定端點的資料來源,回傳結構保持穩定,欄位擴充採新增不破壞。
- **錯誤處理**:
  - identifier 不存在、已軟刪除或公司已停用 → `not_found`
  - identifier 為空或格式非法 → `invalid_argument`
- **驗收**:
  - [ ] 無認證可取得公司公開資訊,回傳不含任何內部欄位。
  - [ ] 停用或不存在公司的 identifier 回 `not_found`。
  - [ ] 回傳包含 capabilities 與 public_info,可供未來協定適配層取用。

---

## Task 2.9 角色與功能權限 API

### 子功能 2.9.1: roles + role_permissions schema 與 migration seed

- **目標**: 建立 `roles` 與 `role_permissions` 表,並以 migration seed 七個內建角色與其預設功能權限(D9)。相依: `01-auth.md` Task 1.11(developer 角色已 seed,本 migration 需冪等)
- **檔案**:
  - Create: `backend/internal/domain/roles/schema.go`(Ent schema)
  - Create: `backend/database/migrations/000XX_roles_seed.sql`
- **介面**: Ent 實體
  - `roles`:id、`code`、`name`、`data_scope`(all / company / department / self)、`is_system`、`is_active`、`deleted_at`;`code` 部分唯一索引(`WHERE deleted_at IS NULL`)。
  - `role_permissions`:id、`role_id`、`resource`(order / customer / product / dispatch / print / user / company 等)、`action`(create / read / update / delete / manage / print / dispatch 等);`role_id + resource + action` 唯一。
- **實作邏輯**:
  1. Ent schema 對齊規格 5.2 欄位定義;`roles` 適用軟刪除但僅自訂角色可用(規格 5.4)。
  2. migration seed 七內建角色(規格 3.2),`is_system = true`、`is_active = true`;`data_scope` 對應:super = all、company_admin = company、dept_admin = department、staff = department、customer = self、guest = self(待審核,僅本人帳號資料)、developer = all(D8)。
  3. seed 各內建角色預設 `role_permissions`,依規格 3.2 各角色功能範圍展開為 resource × action 組合;此為 CASL ability 的預設來源。
  4. developer 角色已由 Task 1.11 seed:本 migration 一律使用「不存在才插入」的冪等寫法,重複執行不產生重複列、不覆寫已被 Web 調整過的權限(判斷依據為 `code` 是否已存在;已存在則略過該角色整組 seed)。
  5. seed 與建表同一 migration,確保新環境開箱即用。
- **錯誤處理**: migration 層無 Connect code;重複執行必須無副作用(冪等),語法或約束錯誤使 goose 中止並保留版本紀錄。
- **驗收**:
  - [ ] migration 後七個內建角色存在,`is_system = true`,data_scope 與上表一致。
  - [ ] 每個內建角色有對應預設 `role_permissions`。
  - [ ] migration 重複執行(或 developer 已存在)不報錯、不複寫既有資料。

### 子功能 2.9.2: 角色 CRUD(is_system 保護、自訂角色必填 data_scope)

- **目標**: 提供角色列表、查詢、建立、更新、軟刪除;內建角色受保護,自訂角色必須指定資料範圍等級。相依: 2.9.1
- **檔案**:
  - Create: `backend/internal/domain/roles/handler.go`
  - Create: `backend/internal/domain/roles/usecase.go`
  - Create: `backend/internal/domain/roles/repository.go`
  - Update: `backend/proto`(RoleService proto 定義)
- **介面**: Connect-RPC `RoleService`
  - `ListRoles`(分頁、含 is_system / is_active 篩選)→ 角色陣列
  - `GetRole`(role_id)→ 角色完整欄位
  - `CreateRole`(code、name、data_scope)→ 新自訂角色
  - `UpdateRole`(role_id + name / is_active;內建角色不可改 code / data_scope)→ 更新後角色
  - `DeleteRole`(role_id)→ 空回應(軟刪除,僅自訂角色)
- **實作邏輯**:
  1. 寫入權限:建立、更新、刪除角色僅 `super`(D9:super 可新增自訂角色;roles 為系統級資料);List / Get 開放 `super` 與 `company_admin`(後者供自己 domain 指派與 policy 設定時選用)。
  2. Create 自訂角色:必填 `code`、`name`、`data_scope`,且 `data_scope` 僅允許 `company` 或 `department`(規格 3.2);`is_system` 恆為 false,API 層拒絕任何嘗試建立 is_system 的請求。
  3. Update:目標 `is_system = true` 時,拒絕修改 `code` 與 `data_scope`(D9);`name`、`is_active` 可調整。
  4. Delete:僅自訂角色可軟刪除;刪除前檢查無使用者仍指派該角色(Casbin g 規則查詢),有使用者綁定回 `failed_precondition` 並提示先改派;通過後同一交易軟刪除角色、清理其 g 規則與 `role_permissions`、寫稽核。
  5. 交易邊界:角色寫入 + 關聯清理 + `audit_logs`(resource_type = role)同一交易。
- **錯誤處理**:
  - 非 `super` 呼叫寫入 RPC → `permission_denied`
  - 建立時缺 data_scope 或值非 company / department → `invalid_argument`
  - code 重複(未軟刪除角色)→ `already_exists`
  - 修改內建角色 code / data_scope、刪除內建角色 → `failed_precondition`
  - 刪除仍有使用者綁定的角色 → `failed_precondition`
  - role_id 不存在 → `not_found`
- **驗收**:
  - [ ] 七個內建角色不可刪除、不可改 code / data_scope。
  - [ ] 自訂角色可建立(必填 data_scope)、停用、軟刪除;軟刪除後同 code 可重建。
  - [ ] 有使用者綁定的角色不可刪除。
  - [ ] 異動皆同事務寫入 `audit_logs`。

### 子功能 2.9.3: 功能權限矩陣編輯 API

- **目標**: 提供角色功能權限(resource × action)的查詢與全量更新,供 Web 權限矩陣編輯。相依: 2.9.1
- **檔案**:
  - Update: `backend/internal/domain/roles/handler.go`
  - Update: `backend/internal/domain/roles/usecase.go`
- **介面**: Connect-RPC `RoleService`
  - `GetPermissions`(role_id)→ 該角色全部 resource × action 組合
  - `UpdatePermissions`(role_id、permissions 陣列:resource + action)→ 更新後權限清單
- **實作邏輯**:
  1. 寫入權限僅 `super`;`company_admin` 可 GetPermissions 查閱(供其了解自己 domain 可用角色能力),不可 Update。
  2. UpdatePermissions 採全量覆寫語意:同一交易內刪除該角色既有 `role_permissions` 並插入新組合;resource / action 須為系統已定義的列舉值,未知值回 `invalid_argument`。
  3. 內建角色的功能權允許調整(D9:預設值可於 Web 調);但本表僅驅動前端 CASL UI 顯示,後端授權由 Casbin policy 決定 — 防鎖死保證在 2.10.3 處理,此處不額外限制清空權限,惟 response 與稽核須如实反映。
  4. 同一交易寫 `audit_logs`(resource_type = role,快照含前後權限組合)。
  5. 生效路徑:前端 ability 快取 60 秒 TTL 到期或權限異動後主動重載(規格 3.4),GetAbility 下次呼叫即反映新權限(見 2.9.4),無需重啟。
- **錯誤處理**:
  - 非 `super` 呼叫 UpdatePermissions → `permission_denied`
  - role_id 不存在 → `not_found`
  - permissions 含未知 resource / action → `invalid_argument`
- **驗收**:
  - [ ] 修改角色權限後,GetAbility 回傳的 ability JSON 即時反映(60 秒 TTL 內或主動重載後)。
  - [ ] 更新為全量覆寫:未提交的既有權限被移除。
  - [ ] 未知 resource / action 回 `invalid_argument`。
  - [ ] 權限異動寫入稽核且同事務。

### 子功能 2.9.4: GetAbility 改由 role_permissions 驅動

- **目標**: `AbilityService.GetAbility` 由 Phase 1 內建預設規則改為依 `role_permissions` 表動態產生 CASL JSON(接管 `01-auth.md` Task 1.8 的實作)。相依: 2.9.1;被 2.9.3 引用
- **檔案**:
  - Update: `backend/internal/domain/ability`(Task 1.8 既有實作,替換規則來源)
  - Update: `backend/internal/domain/roles/repository.go`(提供依角色查權限)
- **介面**: Connect-RPC `AbilityService.GetAbility`(無參數,依當前認證身分)→ CASL 規則陣列(action + subject)。
- **實作邏輯**:
  1. 由認證 context 取當前使用者的全部角色(g 規則),查 `role_permissions` 取所有 resource × action 組合;多角色取聯集,對應為 CASL 的 subject(resource)× action 規則。
  2. 角色已停用(`is_active = false`)或其 `data_scope` 等級不影響本 RPC — 本 RPC 只反映功能權限矩陣;資料範圍由 RLS 處理(2.9.5)。
  3. `developer` 角色且開關啟用時,直接回傳 manage all 規則,不查表(對齊繞過語意,D8)。
  4. `guest` 無任何業務權限,回傳空規則或僅帳號相關最小規則。
  5. 回傳格式維持 Phase 1 契約不變,前端無需修改;前端快取策略(60 秒 TTL 或異動後主動重載)不變(規格 3.4)。
  6. 本 RPC 為唯讀,不寫稽核。
- **錯誤處理**:
  - 未認證 → `unauthenticated`
  - 使用者無任何角色(異常狀態)→ 回傳空規則陣列,不視為錯誤
- **驗收**:
  - [ ] 各內建角色取得之 ability 與 2.9.1 seed 的 `role_permissions` 一致。
  - [ ] 經 2.9.3 修改權限後,下一次 GetAbility 即反映新規則。
  - [ ] 多角色使用者取得聯集規則;developer 取得 manage all;guest 取得空規則。
  - [ ] 回傳格式與 Phase 1 契約相容,前端不需改動。

### 子功能 2.9.5: RLS data_scope 注入(app.current_data_scope)

- **目標**: RLS session 注入改由角色 `data_scope` 等級驅動(不依角色名稱),讓自訂角色不改程式即獲得正確資料範圍(D3)。相依: 2.9.1;掛接 `01-auth.md` Task 1.3 的 connection hook
- **檔案**:
  - Update: `backend/internal/auth`(middleware 解析 data_scope)
  - Update: `backend/internal/infra`(Ent connection hook,注入 `app.current_data_scope`)
  - Update: `backend/database/migrations/000XX_rls_data_scope.sql`(RLS policy 納入 data_scope 判斷)
- **介面**: 無新增 RPC;影響所有受 RLS 保護表的查詢。session 變數:`app.current_company_id`、`app.current_department_id`、`app.current_data_scope`、`app.current_customer_id`。
- **實作邏輯**:
  1. middleware 於認證通過後,依當前使用者角色查 `roles.data_scope`;多角色取最寬等級(all > company > department > self),注入 request context。
  2. Ent connection hook 於每個交易開啟時以 `SET LOCAL` 注入四個 session 變數;`SET LOCAL` 僅存續於交易內,交易結束自動還原,杜絕跨請求污染。
  3. RLS policy 依 `app.current_data_scope` 分支:
     a. `all`:繞過資料範圍限制(super、developer)。
     b. `company`:僅 `company_id = app.current_company_id` 的資料可見可寫。
     c. `department`:再加 `department_id = app.current_department_id`。
     d. `self`:客戶帳號以 `app.current_customer_id` 比對,僅見自己的訂單與主檔。
  4. 無角色或查無 `data_scope` 的異常身分:注入最嚴格等級(self),寧可過濾過嚴不可洩漏。
  5. `developer` 且開關啟用:注入 `data_scope = all`(D8)。
  6. 業務表 RLS policy 統一依本規則改寫;repository 層仍加顯式查詢條件,RLS 為最後防線(D3)。
- **錯誤處理**: 注入失敗(session 變數設定錯誤)使交易失敗並回內部錯誤,不得降級為無 RLS 查詢;未認證請求不進入本流程(`unauthenticated` 由認證層處理)。
- **驗收**:
  - [ ] 四種 data_scope 等級各有對應查詢結果:all 全見、company 限同公司、department 限同部門、self 僅本人。
  - [ ] 自訂角色指定 data_scope 後,不需改程式即獲得對應資料範圍。
  - [ ] 多角色使用者取最寬等級;無角色身分僅見自身資料。
  - [ ] 跨請求無 session 變數殘留(連續不同身分請求互不污染)。

---

## Task 2.10 API 權限管理 API(Casbin policy)

### 子功能 2.10.1: 預設 p 規則 seed

- **目標**: 以 migration seed 各角色預設 Casbin p 規則,取代/承接 Task 1.2 程式內 seeder,使預設 API 權限開箱即用且可由 Web 調整(D9)。相依: `01-auth.md` Task 1.2(enforcer 與 adapter)
- **檔案**:
  - Create: `backend/database/migrations/000XX_casbin_policies_seed.sql`
- **介面**: Casbin `casbin_rules` 表(ptype = p;v0 = role_code、v1 = domain、v2 = 資源路徑、v3 = 動作)。
  - Connect-RPC 資源路徑 = RPC method path(如 `/v1.SalesOrderService.ListSalesOrders`),動作恆為 POST;REST 公開端點 = URL path + HTTP method(規格 3.3)。
- **實作邏輯**:
  1. 依規格 3.2 角色定義與 3.3 授權模型,列出各角色對各 Service method 的預設允許規則;domain 一律為公司層級佔位(各公司 domain 的實際綁定由 g 規則承載)。
  2. seed 採冪等寫法(不存在才插入),重複執行無副作用;已被 Web 調整過的環境不因重跑 migration 被覆寫。
  3. `developer` 角色不 seed 任何 p 規則(其繞過由 middleware 實現,D8);`guest` 僅 seed 完成註冊所需的最小端點。
  4. 程式內 seeder(Task 1.2 Step 3)改為僅在 migration 資料缺失時報錯提示,不再負責寫入,避免雙來源。
  5. enforcer 啟動時自 PostgreSQL adapter 載入全部規則;seed 完成後首次啟動即生效。
- **錯誤處理**: migration 層無 Connect code;語法/約束錯誤由 goose 中止並保留版本紀錄;冪等保證重複執行不報錯。
- **驗收**:
  - [ ] migration 後 `casbin_rules` 含與規格 3.2 一致的各角色預設 p 規則。
  - [ ] enforcer 載入後,各角色對應 API 依預設規則放行/拒絕。
  - [ ] 重複執行 migration 不產生重複規則、不覆寫 Web 調整結果。

### 子功能 2.10.2: Policy CRUD 即時生效

- **目標**: 提供 policy 的列表、新增、刪除,異動經 enforcer 記憶體更新即時生效,不需重啟(D9)。相依: 2.10.1
- **檔案**:
  - Create: `backend/internal/domain/policies/handler.go`
  - Create: `backend/internal/domain/policies/usecase.go`
  - Create: `backend/internal/domain/policies/repository.go`(casbin_rules 直操作 + enforcer 協調)
- **介面**: Connect-RPC `PolicyService`
  - `ListPolicies`(分頁、role / domain 篩選)→ p 規則陣列(role、domain、資源路徑、動作)
  - `AddPolicy`(role、domain、資源路徑、動作)→ 新規則
  - `DeletePolicy`(rule_id 或完整四元組)→ 空回應
- **實作邏輯**:
  1. 寫入前驗證:role 必須存在(查 `roles`);資源路徑須為合法 RPC method path 或 REST path 格式;動作 Connect 端點僅允許 POST、REST 端點允許對應 HTTP method;任一不符回 `invalid_argument`。
  2. Add:先查重(同四元組已存在回 `already_exists`);同一交易寫入 `casbin_rules` + `audit_logs`;交易提交成功後呼叫 enforcer 新增規則使記憶體態同步,提交失敗則不動記憶體,保證 DB 與記憶體一致。
  3. Delete:同樣先 DB 交易(刪除 + 稽核)再更新記憶體;防鎖死檢查見 2.10.3。
  4. 多 replica 一致性:異動提交後經 Valkey pub/sub 廣播 policy 變更事件(複用 D14 的跨 replica 廣播基礎設施),其他 replica 收到後自 adapter 重新載入規則;單體部署時此步驟無害。
  5. List 依操作者範圍過濾(見 2.10.3);讀取直查 DB 而非記憶體,保證所見即持久態。
- **錯誤處理**:
  - 未認證 → `unauthenticated`
  - 無權限(非 super / company_admin,或跨 domain)→ `permission_denied`
  - role 不存在、路徑/動作格式非法 → `invalid_argument`
  - 重複新增同四元組 → `already_exists`
  - 刪除不存在的規則 → `not_found`
  - 觸發防鎖死 → `failed_precondition`(見 2.10.3)
- **驗收**:
  - [ ] 新增 policy 後,對應角色下一次 API 呼叫即被放行,無需重啟。
  - [ ] 刪除 policy 後,對應呼叫即被拒絕。
  - [ ] 重啟後異動結果仍存(DB 為準,記憶體重新載入一致)。
  - [ ] 每筆異動寫入 `audit_logs` 且與規則異動同事務。

### 子功能 2.10.3: domain 範圍控制 + 防鎖死

- **目標**: `super` 管理全域 policy,`company_admin` 僅限自己公司 domain;防止移除操作者自身角色的最後一個管理權限(D9)。相依: 2.10.2
- **檔案**:
  - Update: `backend/internal/domain/policies/usecase.go`
- **介面**: 無新增 RPC;為 2.10.2 各 RPC 與 2.10.4 的前置判斷規則。
- **實作邏輯**:
  1. domain 範圍判斷:
     a. `super`:可操作任意 domain 的 policy。
     b. `company_admin`:List 強制過濾自己公司 domain;Add 的 domain 參數必須等於自己公司;Delete 先查目標規則 domain 再比對;不符一律 `permission_denied`(規格 3.3)。
     c. 其餘角色不開放 PolicyService 寫入。
  2. 防鎖死檢查(Delete 與 2.9.2 角色異動連動):
     a. 刪除 policy 前,若該規則屬於「操作者自身角色於該 domain 的 policy 管理能力」(即 PolicyService 相關 method 的允許規則),檢查刪除後該角色在該 domain 是否仍有至少一條 policy 管理規則。
     b. 若為最後一條,拒絕刪除並回 `failed_precondition`,訊息指明會造成管理權限鎖死。
     c. 同一交易內完成檢查與刪除,避免併發下兩個請求各自通過檢查(檢查查詢與刪除同交易,由交易隔離保證)。
  3. `super` 角色的全域管理規則同受防鎖死保護:不可刪到 super 失去 PolicyService 管理能力。
  4. 所有異動寫 `audit_logs`(resource_type = casbin_policy,快照含規則四元組),同事務。
- **錯誤處理**:
  - `company_admin` 查/增/刪其他公司 domain 的 policy → `permission_denied`
  - 非管理角色呼叫 → `permission_denied`
  - 刪除將移除自身最後管理權限 → `failed_precondition`
- **驗收**:
  - [ ] `company_admin` 僅見、僅能改自己公司 domain 的 policy;對他公司 domain 操作回 `permission_denied`。
  - [ ] 嘗試刪除操作者自身角色最後一條 policy 管理規則時被拒,系統不鎖死。
  - [ ] 併發刪除情境下防鎖死仍成立(同事務檢查)。
  - [ ] super 全域管理規則同受保護。

### 子功能 2.10.4: ListGrouping(g 規則檢視)

- **目標**: 提供 Casbin g 規則(使用者 → 角色 @ domain)的唯讀檢視,供 Web 權限管理頁核對指派現況。相依: 2.10.2
- **檔案**:
  - Update: `backend/internal/domain/policies/handler.go`
  - Update: `backend/internal/domain/policies/usecase.go`
- **介面**: Connect-RPC `PolicyService.ListGrouping`(分頁、role / domain / user 篩選)→ g 規則陣列(user、role、domain)。
- **實作邏輯**:
  1. 唯讀查詢,直查 `casbin_rules`(ptype = g),不讀 enforcer 記憶體。
  2. 範圍控制:`super` 見全部;`company_admin` 強制過濾自己公司 domain;其餘角色不開放。
  3. 支援依 user / role / domain 組合篩選與分頁;結果附對應使用者顯示名稱與角色名稱(join `users` / `roles`,僅顯示用途)。
  4. g 規則的新增/移除不由本 RPC 提供 — 使用者角色指派統一走 2.3.1 `AssignRole`(單一寫入路徑,避免雙來源);本 RPC 僅檢視。
- **錯誤處理**:
  - 未認證 → `unauthenticated`
  - 非 super / company_admin 呼叫 → `permission_denied`
- **驗收**:
  - [ ] `super` 可檢視全部 g 規則;`company_admin` 僅見自己公司 domain。
  - [ ] 2.3.1 角色指派異動後,ListGrouping 結果同步反映。
  - [ ] 篩選與分頁正確;回傳含使用者與角色顯示名稱。

---

## 整合測試重點

依 D21(授權/RLS 需整合測試,dockertest 起真實 PostgreSQL)與本文件驗收欄,Phase 2 至少覆蓋以下路徑:

1. **公司停用連鎖阻斷**:建立公司 + 員工 + 客戶帳號並登入 → 停用公司 → 驗證 Web session 與 App JWT 下一次請求皆 `unauthenticated`、重新登入 `permission_denied` → 恢復 active 後可登入;developer 帳號不受影響(2.1.3)。
2. **唯一性衝突**:併發建立同 identifier / 同 prefix 公司,一者成功一者 `already_exists`;軟刪除後可重用;修改 prefix 後既有客戶編號與計數器不變(2.1.2,與 Phase 3 取號聯動驗證)。
3. **三級管理範圍**:`company_admin` 對他公司使用者/部門/policy 操作、`dept_admin` 對他部門或非 staff 帳號操作,皆 `permission_denied`,且 RLS 層亦查不到越界資料(2.2.1、2.3.2、2.10.3)。
4. **角色變更撤銷**:指派新角色後舊 token 立即失效、g 規則更新、GetAbility 反映新權限;guest 審核全流程(2.3.1、2.9.4)。
5. **data_scope 四等級 RLS 隔離**:同表分別以 all / company / department / self 身分查詢,結果集各異;自訂角色指定 data_scope 不改程式即生效;連續不同身分請求無 session 變數污染(2.9.5)。
6. **policy 即時生效與防鎖死**:新增 policy 後對應呼叫即放行、刪除即拒絕;刪除自身最後管理規則被 `failed_precondition` 擋下;重啟後狀態一致(2.10.2、2.10.3)。
7. **公開端點**:無認證取得公開資訊且不含內部欄位;停用公司回 `not_found`(2.4.3)。
8. **稽核同事務**:上述每條路徑驗證 `audit_logs` 與業務異動同成功同失敗,快照完整(D18)。
