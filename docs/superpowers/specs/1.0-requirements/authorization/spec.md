# authorization 需求規格

> 來源：原 OpenSpec delta spec（OpenSpec 工作流已停用，2026-08-03 遷移至 docs/）。


Casbin RBAC with domain（後端授權）+ PostgreSQL RLS（資料範圍）+ CASL.js（前端能力）的三層權限機制，涵蓋角色定義、policy 管理、權限設置頁面與開發者繞過。

## Requirements

### Requirement: 內建角色定義

系統 MUST 提供 7 個內建角色：`super`、`company_admin`、`dept_admin`、`staff`、`customer`、`guest`、`developer`，由 migration seed 建立，且 `is_system = true`。內建角色 SHALL 不可刪除，且其 `code` 與 `data_scope` 不可修改。角色定義 MUST 包含 `data_scope` 屬性，決定 RLS 資料範圍等級，而非依角色名稱判斷。`staff` 角色 SHALL 依所屬部門資料範圍兼任會計檢視與對帳相關操作，不另設獨立會計角色。

#### Scenario: 內建角色由 migration seed 建立

- **WHEN** 系統於全新資料庫執行 migration
- **THEN** `roles` 表存在 `super`、`company_admin`、`dept_admin`、`staff`、`customer`、`guest`、`developer` 共 7 筆角色
- **AND** 每筆內建角色 `is_system = true`，且具備對應的 `data_scope`（`developer` 為 `all`）

#### Scenario: 內建角色不可刪除

- **WHEN** 任何使用者嘗試刪除 `is_system = true` 的角色
- **THEN** 系統拒絕該操作並回傳錯誤，角色保持存在

#### Scenario: 內建角色 code 與 data_scope 不可修改

- **WHEN** `super` 嘗試修改內建角色的 `code` 或 `data_scope`
- **THEN** 系統拒絕該修改並回傳錯誤，原有 `code` 與 `data_scope` 維持不變

#### Scenario: staff 兼任會計檢視

- **WHEN** 具 `staff` 角色的使用者存取會計檢視或對帳功能
- **THEN** 系統允許其依所屬部門資料範圍執行檢視與對帳操作

### Requirement: 自訂角色管理

`super` SHALL 可新增自訂角色；自訂角色 MUST 指定 `data_scope`，且僅允許 `company` 或 `department` 兩種等級。`company_admin` SHALL 可管理自己公司 domain 內的自訂角色。自訂角色套用軟刪除（`deleted_at`），其 `data_scope` 建立後決定 RLS 資料範圍行為。

#### Scenario: super 建立自訂角色

- **WHEN** `super` 建立自訂角色並指定 `data_scope = department`
- **THEN** 系統建立該角色且 `is_system = false`
- **AND** 被指派該角色的使用者，其 RLS 資料範圍依 `department` 等級生效

#### Scenario: 自訂角色缺少或不合法的 data_scope

- **WHEN** 建立自訂角色時未指定 `data_scope`，或指定 `all` / `self` 以外的公司/部門等級以外之值
- **THEN** 系統拒絕建立並回傳驗證錯誤

#### Scenario: company_admin 跨公司管理自訂角色被拒絕

- **WHEN** `company_admin` 嘗試修改或刪除其他公司 domain 的自訂角色
- **THEN** 系統拒絕該操作並回傳權限不足錯誤

### Requirement: Casbin 後端授權

系統 MUST 使用 Casbin RBAC with domain 進行後端授權，domain 為 `company_id`，policy 儲存於 PostgreSQL。Connect-RPC 請求的 obj MUST 為 RPC method path（如 `/v1.SalesOrderService.ListSalesOrders`），act 恆為 `POST`；REST 公開端點的 obj MUST 為 URL path、act 為對應 HTTP method。Middleware MUST 從已驗證的身分解析 `company_id`、`department_id`、`roles` 並注入 request context 供授權判斷使用。Casbin SHALL 僅負責「角色 × 公司（domain）× 資源 × 動作」的授權判斷；部門層級資料範圍 MUST NOT 由 Casbin domain 承載，統一交由 RLS policy 與 repository 查詢條件實現。各角色的預設 Casbin policy（p 規則）MUST 由 migration seed 依內建角色定義建立。

#### Scenario: 具備 policy 的角色通過授權

- **WHEN** 角色 `staff` 在 `company_01` domain 存在 `p, staff, company_01, /v1.SalesOrderService.ListSalesOrders, POST` 規則，且該公司 `staff` 使用者呼叫該 RPC
- **THEN** middleware 授權通過，請求進入業務處理

#### Scenario: 無 policy 的請求被拒絕

- **WHEN** 角色 `staff` 在 `company_01` 無對應 p 規則，卻呼叫 `/v1.CompanyService.UpdateCompany`
- **THEN** middleware 拒絕請求並回傳權限不足錯誤，不執行業務邏輯

#### Scenario: 跨公司 domain 不套用他公司 policy

- **WHEN** 隸屬 `company_01` 的使用者身分被用於 `company_02` domain 的授權判斷
- **THEN** `company_01` 的 p / g 規則不生效，請求被拒絕

#### Scenario: REST 公開端點以 URL path 與 method 判斷

- **WHEN** 受保護的 REST 端點收到請求
- **THEN** Casbin 以 URL path 為 obj、HTTP method 為 act 進行判斷，而非 RPC method path

#### Scenario: 預設 policy 由 migration seed 建立

- **WHEN** 系統於全新資料庫執行 migration
- **THEN** Casbin p 規則依內建角色定義預先建立，各角色對應的 API 權限即為預設值

### Requirement: RLS 資料範圍

系統 MUST 以 PostgreSQL RLS 實現資料範圍隔離。每個請求 MUST 注入 session variables：`app.current_company_id`、`app.current_department_id`、`app.current_data_scope`、`app.current_customer_id`。RLS policy MUST 依當前角色的 `data_scope` 等級判斷：`all` 繞過 RLS；`company` 僅限同 `company_id` 資料；`department` 僅限同 `department_id` 資料；`self` 僅限本人資料，客戶帳號以 `app.current_customer_id` 比對資料的 `customer_id`。

#### Scenario: data_scope 為 all 繞過 RLS

- **WHEN** 當前角色 `data_scope = all` 的使用者查詢業務資料
- **THEN** RLS 不限制查詢結果，可跨公司讀取全部資料

#### Scenario: data_scope 為 company 限同公司

- **WHEN** 當前角色 `data_scope = company` 的使用者查詢資料
- **THEN** 查詢結果僅包含 `company_id` 等於 `app.current_company_id` 的資料
- **AND** 其他公司的資料不出現在結果中

#### Scenario: data_scope 為 department 限同部門

- **WHEN** 當前角色 `data_scope = department` 的使用者查詢資料
- **THEN** 查詢結果僅包含 `department_id` 等於 `app.current_department_id` 的資料
- **AND** 同公司其他部門的資料不出現在結果中

#### Scenario: data_scope 為 self 的客戶僅見本人資料

- **WHEN** `customer` 角色（`data_scope = self`）的客戶帳號查詢訂單
- **THEN** 查詢結果僅包含 `customer_id` 等於 `app.current_customer_id` 的訂單與相關主檔
- **AND** 其他客戶的資料不出現在結果中

### Requirement: CASL 前端能力

Web MUST 於登入後向後端 `AbilityService.GetAbility` 載入當前使用者的 ability JSON 並快取，快取 TTL 為 60 秒，或於權限異動後主動重新載入；不得在每次路由切換時重新請求。Ability MUST 由後端依 `role_permissions` 表（resource × action）動態產生，預設值依內建角色定義由 migration seed 建立，非前端靜態規則。Web SHALL 使用 ability 進行路由守衛與按鈕/選單顯示控制。App 因功能較少，SHALL 使用簡單 role 判斷，不載入 CASL ability。

後端 MUST 以同一份 `role_permissions` 規則（含 `conditions`、`inverted`）在 `CASL_ENFORCEMENT_ENABLED = true` 時執行：list 查詢依規則產生 SQL 過濾條件（無允許規則時回空集）；寫入/更新/取消等操作對實體做 `can(action, instance)` 條件檢查。開關關閉時後端跳過 CASL 執行層，行為等同 Casbin + RLS（RLS 仍為租戶隔離最後防線）。

#### Scenario: 登入後載入並快取 ability

- **WHEN** 使用者登入 Web 中台
- **THEN** 前端向 `AbilityService.GetAbility` 取得 ability JSON 一次並快取
- **AND** 60 秒 TTL 內的路由切換不再請求 ability

#### Scenario: 權限異動後重新載入

- **WHEN** 使用者權限於後端異動，且前端快取到期或收到異動通知
- **THEN** 前端重新向 `AbilityService.GetAbility` 載入最新 ability，顯示控制隨之更新

#### Scenario: ability 由 role_permissions 動態產生

- **WHEN** 後端修改某角色的 `role_permissions`（resource × action）
- **THEN** 該角色使用者下次取得的 ability JSON 反映最新權限，不需修改前端程式碼

#### Scenario: ability 控制路由與按鈕

- **WHEN** 使用者 ability 不含某 resource 的 `read` action
- **THEN** 前端路由守衛阻止進入對應頁面，相關選單與按鈕不顯示

#### Scenario: list 查詢套用規則條件

- **WHEN** `staff` 角色的 `role_permissions` 含 `read sales_order conditions={"department_id": "${user.department_id}"}`，且開關啟用
- **THEN** 該使用者 list 訂單僅回符合條件的資料，且結果仍受 RLS 限制（交集語意）

#### Scenario: 實例檢查攔截狀態越權

- **WHEN** 規則為 `cancel sales_order conditions={"status": "pending"}`，staff 嘗試取消 `processing` 訂單，且開關啟用
- **THEN** 後端回 `permission_denied`，不執行取消

#### Scenario: 開關關閉時降級

- **WHEN** `CASL_ENFORCEMENT_ENABLED = false`
- **THEN** list 查詢不附加規則條件、實例檢查放行，Casbin 與 RLS 行為不變；`GetAbility` 回傳不受影響

#### Scenario: 規則異動前後端同源生效

- **WHEN** `super` 於權限設置頁修改某角色規則的 conditions
- **THEN** 下一次 `GetAbility`（前端）與下一次 list/實例檢查（後端）皆反映新條件，不需重啟

### Requirement: 權限管理範圍與防鎖死

`super` SHALL 可管理全域角色與所有 Casbin policy；`company_admin` SHALL 僅可管理自己公司 domain 的 policy 與自訂角色。系統 MUST 防止移除操作者自身角色的最後一個管理權限，避免權限管理鎖死。所有權限異動 MUST 寫入 `audit_logs`。Policy 異動 MUST 即時生效，不需重啟服務。

#### Scenario: company_admin 僅管理自己公司 domain

- **WHEN** `company_admin` 嘗試新增、修改或刪除其他公司 domain 的 Casbin policy
- **THEN** 系統拒絕該操作並回傳權限不足錯誤
- **AND** 自己公司 domain 的 policy 管理操作正常完成

#### Scenario: 防鎖死保護

- **WHEN** 操作者嘗試移除自身角色所持有的最後一個權限管理權限
- **THEN** 系統拒絕該異動並回傳錯誤，該管理權限保留

#### Scenario: 權限異動寫入稽核日誌

- **WHEN** 任何角色、`role_permissions` 或 Casbin policy 被新增、修改或刪除
- **THEN** 系統寫入一筆 `audit_logs`，記錄操作者、異動內容與前後狀態

#### Scenario: policy 異動即時生效

- **WHEN** `super` 刪除某角色對某 RPC method path 的 p 規則
- **THEN** 該角色使用者下一次請求該 RPC 即被拒絕，不需重啟後端服務

### Requirement: 開發者繞過

當使用者角色為 `developer` 且環境開關 `DEVELOPER_ACCOUNT_ENABLED = true` 時，middleware MUST 跳過 Casbin 授權檢查，且 RLS session MUST 注入 `data_scope = all`，使其可跨公司/部門存取所有資料與功能。當開關關閉時 MUST NOT 執行繞過。

#### Scenario: 開關啟用時繞過授權與資料範圍

- **WHEN** 角色為 `developer` 且 `DEVELOPER_ACCOUNT_ENABLED = true` 的使用者呼叫任何業務 RPC
- **THEN** middleware 跳過 Casbin 檢查，請求直接通過授權
- **AND** RLS session 注入 `data_scope = all`，可查詢跨公司/部門資料

#### Scenario: 開關關閉時不繞過

- **WHEN** `DEVELOPER_ACCOUNT_ENABLED = false`
- **THEN** middleware 不執行 developer 繞過邏輯
- **AND** developer 帳號請求依一般授權規則處理，不享有跨公司資料存取

### Requirement: Web 權限設置頁面

Web 中台 MUST 提供兩個權限設置頁面：「角色權限設置」與「API 權限設置」，僅 `super` 與 `company_admin` 可存取。角色權限設置頁 SHALL 提供角色 CRUD（自訂角色）與功能權限矩陣（resource × action）編輯；API 權限設置頁 SHALL 提供 Casbin p 規則（角色 × domain × path × method）CRUD，並支援依角色與 domain 篩選。`company_admin` 使用兩頁面時 MUST 僅見自己公司 domain 的資料。

#### Scenario: super 使用角色權限設置頁

- **WHEN** `super` 開啟角色權限設置頁
- **THEN** 可檢視所有角色、建立/編輯自訂角色，並於功能權限矩陣編輯各角色的 resource × action 權限
- **AND** 內建角色顯示為預設值且不可刪除

#### Scenario: super 使用 API 權限設置頁

- **WHEN** `super` 開啟 API 權限設置頁並依角色與 domain 篩選
- **THEN** 顯示符合篩選條件的 Casbin p 規則（角色 × domain × path × method）
- **AND** 可對 p 規則執行新增、修改、刪除

#### Scenario: company_admin 僅見自己公司 domain

- **WHEN** `company_admin` 開啟角色權限設置頁或 API 權限設置頁
- **THEN** 僅顯示自己公司 domain 的自訂角色與 policy
- **AND** 其他公司 domain 的資料不可見、不可操作

#### Scenario: 無權角色不可存取權限設置頁

- **WHEN** `dept_admin`、`staff` 或其他非管理角色嘗試存取權限設置頁
- **THEN** 系統拒絕存取，頁面不顯示
