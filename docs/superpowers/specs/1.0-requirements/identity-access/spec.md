# identity-access 需求規格

> 來源：原 OpenSpec delta spec（OpenSpec 工作流已停用，2026-08-03 遷移至 docs/）。


## Requirements

### Requirement: 員工 OAuth2/OIDC 登入

系統 SHALL 以 OAuth2/OIDC 提供員工與管理員登入，1.0 僅支援 Google Workspace 作為身分提供者。Web 端 MUST 採用 PKCE 流程：使用者點選「以 Google 登入」後跳轉至 provider，回調成功後後端建立 session 並設定 httpOnly cookie，Web 端不得使用 JWT 作為登入態載體。App 端 MUST 透過系統瀏覽器完成 OAuth2，回調成功後後端核發 JWT（access token 與 refresh token）。系統 MUST NOT 為員工帳號儲存密碼（`password_hash` 僅客戶帳號使用）。員工帳號 MUST NOT 被刪除，僅可停用。

#### Scenario: Web 員工完成 PKCE 登入

- **WHEN** Web 使用者在登入頁點選「以 Google 登入」，並於 Google Workspace 完成授權後回到回調端點
- **THEN** 後端驗證授權碼（PKCE）後建立 session，並在回應中設定 httpOnly cookie
- **AND** 後續 Web 請求以該 cookie 識別身分，前端不持有任何 token

#### Scenario: App 員工完成 OAuth2 登入

- **WHEN** App 使用者以系統瀏覽器完成 Google OAuth2 授權並回調成功
- **THEN** 後端核發 access JWT 與 refresh token 給 App
- **AND** App 後續請求以 `authorization: Bearer <jwt>` 攜帶憑證

#### Scenario: 員工帳號僅可停用

- **WHEN** 管理者嘗試刪除任一員工帳號
- **THEN** 系統拒絕刪除操作，僅提供停用（`status` 設為 `inactive`）
- **AND** 停用後該帳號無法再登入，其既有 token 立即失效

#### Scenario: 所屬公司停用時員工登入被拒

- **WHEN** 員工所屬公司的 `status` 非 `active`，該員工嘗試登入
- **THEN** 系統拒絕登入，不建立 session、不核發 token
- **AND** 該公司既有已登入的 session / token 於下一次請求時失效

### Requirement: guest 註冊完成頁與審核

首次 OAuth2 登入且系統無對應帳號時，系統 SHALL 導向註冊完成頁，要求使用者選擇所屬公司並輸入姓名；未完成註冊完成頁前 MUST NOT 建立帳號，亦不得進入系統。註冊完成頁的公司清單 MUST 僅回傳公開基本資訊（`display_name`），不得暴露其他公司資料。完成註冊後系統 SHALL 建立 `guest` 帳號（`status = pending`），歸屬所選公司。`guest` 帳號在審核前 MUST NOT 操作任何業務資料。審核 SHALL 由 `super` 或該公司的 `company_admin` 執行，核准時指派部門與角色；`dept_admin` MUST NOT 審核 guest。

#### Scenario: 首次 OAuth 登入導向註冊完成頁

- **WHEN** 一個系統中不存在的 Google 帳號首次完成 OAuth2 登入
- **THEN** 系統不建立帳號，導向註冊完成頁要求選擇公司與輸入姓名
- **AND** 在完成註冊前該使用者無法存取任何系統功能

#### Scenario: 註冊完成頁公司清單僅含公開資訊

- **WHEN** 註冊完成頁載入可選公司清單
- **THEN** 回應中每家公司僅包含 `display_name` 等公開基本資訊
- **AND** 回應不包含公司內部設定、成員或業務資料

#### Scenario: 完成註冊建立 pending guest

- **WHEN** 使用者於註冊完成頁選擇公司並輸入姓名後送出
- **THEN** 系統建立 `guest` 帳號，`status` 為 `pending` 並歸屬所選公司
- **AND** 該帳號在審核通過前無法操作業務資料

#### Scenario: company_admin 核准 guest

- **WHEN** 該公司的 `company_admin`（或 `super`）核准一個 pending 的 guest，並指派部門與角色
- **THEN** 該帳號 `status` 轉為 `active`，取得所指派的部門與角色
- **AND** 該帳號即可依新角色登入使用系統

#### Scenario: dept_admin 不得審核 guest

- **WHEN** `dept_admin` 嘗試核准 guest 或對 guest 指派角色
- **THEN** 系統拒絕該操作
- **AND** guest 維持 `pending` 狀態不變

### Requirement: 客戶密碼登入與密碼政策

系統 SHALL 提供客戶以帳號名稱（`users.account_name`，客戶內唯一）+ 密碼登入；**每個客戶有 1 個主帳號 + 多個子帳號**（主帳號於建立客戶時產生、`is_primary = true`、帳號名稱預設為客戶名稱；子帳號如主廚由店家以主帳號新增），各帳號獨立密碼、可獨立停用 / 重置，任一帳號的登出、停用或密碼重置 MUST NOT 影響其他帳號（`customer_code` 為客戶身分識別，非登入識別）。**主帳號 SHALL 僅供管理登入**：登入後僅能使用帳號管理功能，所有業務 API（下單、訂單歷史、退貨、專屬商品、促銷等）對主帳號 session MUST 拒絕（403）；**子帳號 SHALL 為唯一的 App 業務登入身分**（老闆本人需下單 / 查單時，另建子帳號使用）。QR Code 登入的帳號選擇清單 SHALL 僅列子帳號。密碼 MUST 以 Argon2id 雜湊儲存。建立客戶帳號時系統 SHALL 產生臨時密碼，效期為 24 小時（`temp_password_expires_at`），客戶首次以臨時密碼登入後 MUST 強制修改密碼才能繼續使用。密碼強度 MUST 最少 8 字元。臨時密碼過期或客戶忘記密碼時，SHALL 由 `dept_admin` 以上角色重新產生（重置）臨時密碼；密碼重置 MUST 連帶解除登入鎖定，並使該帳號既有 token 失效。

#### Scenario: 臨時密碼首次登入強制修改

- **WHEN** 客戶以系統產生的臨時密碼首次登入成功
- **THEN** 系統要求立即設定新密碼
- **AND** 未完成密碼修改前不得使用其他系統功能

#### Scenario: 新密碼少於 8 字元被拒

- **WHEN** 客戶於強制修改密碼時提交少於 8 字元的新密碼
- **THEN** 系統拒絕並提示密碼強度不足
- **AND** 帳號維持需修改密碼的狀態

#### Scenario: 臨時密碼超過 24 小時失效

- **WHEN** 客戶在臨時密碼產生超過 24 小時後才嘗試以該密碼登入
- **THEN** 系統拒絕登入，提示臨時密碼已過期
- **AND** 須由 `dept_admin` 以上角色重新產生臨時密碼後方可登入

#### Scenario: dept_admin 重置密碼連帶解鎖並作廢 token

- **WHEN** `dept_admin` 以上角色為客戶重置密碼
- **THEN** 系統產生新的臨時密碼（效期 24 小時）、解除該帳號的登入鎖定，並將 `token_version` + 1
- **AND** 該帳號先前核發的 JWT 與 refresh token 立即失效

### Requirement: 客戶登入錯誤鎖定

系統 SHALL 對客戶密碼登入實施錯誤鎖定：連續 5 次密碼錯誤 MUST 鎖定帳號，鎖定期間拒絕登入，30 分鐘後自動解除。`dept_admin` 以上角色 SHALL 可手動解鎖；重置密碼時 MUST 連帶解除鎖定。登入成功時 SHALL 將連續錯誤計數（`failed_login_attempts`）歸零。

#### Scenario: 連續 5 次錯誤後鎖定

- **WHEN** 客戶連續第 5 次輸入錯誤密碼
- **THEN** 系統鎖定該帳號並記錄鎖定時間（`locked_at`）
- **AND** 鎖定期間即使輸入正確密碼也拒絕登入

#### Scenario: 鎖定 30 分鐘後自動解除

- **WHEN** 帳號被鎖定且自鎖定時間起已超過 30 分鐘
- **THEN** 系統自動解除鎖定
- **AND** 客戶可以正確密碼正常登入，錯誤計數歸零

#### Scenario: 登入成功重置錯誤計數

- **WHEN** 客戶在連續錯誤未達 5 次前以正確密碼登入成功
- **THEN** 系統將 `failed_login_attempts` 歸零
- **AND** 帳號不進入鎖定狀態

### Requirement: QR Code 登入

系統 SHALL 提供 QR Code 深層連結登入：連結內含經簽章的一次性 token，token MUST 編碼 `company_id`、`customer_code` 與過期時間；後端驗證簽章與效期後 MUST 以 token 內的 `company_id` 定位所屬公司，再帶入對應客戶身分，由使用者選擇任一**店家子帳號**完成登入（主帳號為管理用途、業務子帳號專供所屬業務使用，皆不列於 QR 清單），不得僅以 `customer_code` 跨公司識別客戶。QR Code 連結 SHALL 採 Universal Link / App Link 形式（如 `https://<domain>/customer_account_qrcode/{token}`）：已安裝 App 時直接開啟並完成登入，未安裝時導向商店下載頁。token MUST 僅可使用一次，驗證成功或過期後即失效。

#### Scenario: 有效 QR token 登入

- **WHEN** 客戶開啟含有效簽章 token 的 QR Code 連結且已安裝 App
- **THEN** 後端驗證 token 簽章與效期，依 `company_id` + `customer_code` 定位客戶身分，使用者選擇任一店家子帳號完成登入（業務子帳號不列於清單）
- **AND** 該 token 隨即作廢，不可再次使用

### Requirement: 帳號管理深層連結

系統 SHALL 提供帳號管理深層連結 `https://<domain>/customer_account_manage`（Universal Link / App Link）：點擊後已安裝 App 直接開啟並導向「帳號管理」登入流程（以主帳號帳密登入），未安裝則導向商店下載頁。連結 MUST NOT 內含任何登入憑證（登入仍須主帳號帳密）；業務於 App 新增客戶當場交付帳密時 SHALL 一併顯示該連結（可為 QR Code）。

#### Scenario: 點擊管理連結開啟帳號管理

- **WHEN** 店家點擊業務交付的管理網址鏈結且已安裝 App
- **THEN** App 開啟並導向「帳號管理」登入流程
- **AND** 店家以主帳號帳密登入後進入帳號管理畫面

#### Scenario: 未安裝 App 導向商店

- **WHEN** 店家點擊管理網址鏈結但未安裝 App
- **THEN** 系統導向 App Store / Google Play 下載頁

#### Scenario: 過期或重複使用 token 被拒

- **WHEN** 客戶使用已過期或已兌換過的 QR token
- **THEN** 系統拒絕登入
- **AND** 不核發任何 session 或 token

#### Scenario: 多公司環境以 company_id 定位客戶

- **WHEN** 系統中存在不同公司下相同 `customer_code` 的客戶帳號
- **THEN** 系統必須依 token 內的 `company_id` 定位正確公司的客戶
- **AND** 不得登入到其他公司的同名客戶帳號

### Requirement: Session 與 Token 生命週期

Web 端 SHALL 使用 httpOnly cookie + CSRF token 維護 session。App 端 SHALL 使用 JWT：access token 效期 1 小時，refresh token 效期 30 天且採旋轉制——每次 refresh MUST 核發新 refresh token 並作廢舊 token，後端 MUST 僅儲存 refresh token 的雜湊值。系統 SHALL 以 `users.token_version` 支援 token 撤銷：JWT 與 refresh token 皆帶 `tv` claim，驗證時比對資料庫目前值；停用帳號、強制登出、密碼重置、角色變更時 `token_version` MUST + 1，舊 token 立即失效。API access token SHALL 僅供 server-to-server 呼叫，以獨立 header `X-Api-Token` 驗證，MUST NOT 配置於 Web / App 客戶端。

#### Scenario: refresh token 旋轉作廢舊 token

- **WHEN** App 以有效 refresh token 請求換發新 token
- **THEN** 後端核發新的 access token 與新的 refresh token，並作廢舊的 refresh token
- **AND** 再次以舊 refresh token 請求時被拒絕

#### Scenario: 後端僅存 refresh token 雜湊

- **WHEN** 系統儲存任一 refresh token
- **THEN** 資料庫中僅保存其雜湊值
- **AND** 驗證時以雜湊比對，無法從儲存內容還原原始 token

#### Scenario: 角色變更使既有 token 即時失效

- **WHEN** 使用者角色被變更，`token_version` + 1
- **THEN** 該使用者既有 JWT（`tv` claim 為舊值）於下一次請求驗證失敗
- **AND** 既有 refresh token 亦無法再換發新 token

#### Scenario: X-Api-Token 不接受客戶端使用

- **WHEN** 請求以 `X-Api-Token` 存取僅限 server-to-server 的 API
- **THEN** 系統驗證該 token 後放行合法的伺服器端呼叫
- **AND** Web / App 客戶端發佈內容中不包含任何 `X-Api-Token` 設定

### Requirement: 強制登出

系統 SHALL 提供強制登出功能，`super`、`company_admin`、`dept_admin` 可依其管理範圍對帳號執行強制登出。執行時 MUST 刪除該帳號的 Web session，並將 `token_version` + 1 使其 JWT 與 refresh token 立即失效。強制登出 MUST 寫入稽核日誌（`action = force_logout`）。

#### Scenario: 管理者強制登出帳號

- **WHEN** `super` / `company_admin` / `dept_admin` 對其管理範圍內的帳號執行強制登出
- **THEN** 系統刪除該帳號所有 Web session，並將 `token_version` + 1
- **AND** 該帳號既有 JWT 與 refresh token 立即失效，下一次請求即被要求重新登入

#### Scenario: 強制登出寫入稽核

- **WHEN** 任一強制登出操作完成
- **THEN** `audit_logs` 新增一筆 `action = force_logout` 記錄
- **AND** 記錄包含操作人、被登出帳號、時間、IP 與 user agent

系統 SHALL 內建 `developer` 角色帳號供開發/測試使用，其登入 MUST 受環境開關 `DEVELOPER_ACCOUNT_ENABLED` 控制：開關為 `false` 時 developer 帳號無法登入。後端啟動時若 `ENV=production` 且 `DEVELOPER_ACCOUNT_ENABLED=true`，MUST 拒絕啟動（fail fast）並輸出錯誤訊息。開發環境 migration SHALL seed 一組本機開發者帳號（僅 `ENV=development` 時建立）。developer 帳號的所有操作 MUST 照常寫入稽核日誌，不因繞過權限而省略。

#### Scenario: 開關關閉時 developer 無法登入

- **WHEN** `DEVELOPER_ACCOUNT_ENABLED=false`，任何開發者帳號嘗試登入
- **THEN** 系統拒絕登入
- **AND** 不核發 session 或 token

#### Scenario: 生產環境誤開開關時拒絕啟動

- **WHEN** 後端以 `ENV=production` 且 `DEVELOPER_ACCOUNT_ENABLED=true` 啟動
- **THEN** 啟動程序 fail fast，拒絕提供服務
- **AND** 輸出明確錯誤訊息指出開發者帳號不得於生產環境啟用

#### Scenario: 開發環境 seed 本機開發者帳號

- **WHEN** 於 `ENV=development` 環境執行 migration
- **THEN** 系統建立一組本機開發者帳號
- **AND** 於 `ENV=production` 執行 migration 時不建立該帳號

#### Scenario: developer 操作照常寫稽核

- **WHEN** developer 帳號執行任何寫入或敏感操作
- **THEN** 系統照常將該操作寫入 `audit_logs`
- **AND** 稽核記錄可識別操作者為 developer 帳號

### Requirement: 店家自助管理登入帳號

店家 SHALL 以**主帳號**於 App 管理自己客戶底下的帳號：新增子帳號（填寫帳號名稱，走臨時密碼流程）、停用子帳號、重置子帳號密碼。**子帳號 SHALL 僅能登入使用，無帳號管理權限**。範圍 MUST 僅限自己客戶（data_scope self），不得觸及其他客戶或員工帳號。**建立客戶時自動附帶的業務子帳號 SHALL 專供該客戶的所屬業務（`default_sales_rep_id`）使用**，憑證交付所屬業務；於店家帳號管理清單顯示為「系統預設（業務使用）」並灰化（反白）：店家可檢視但 MUST NOT 可改名、停用或重置（店家無該帳號密碼，實質僅所屬業務可登入）。後台 SHALL 可正常管理，並於主責業務變更時將業務子帳號移交新主責業務（重置密碼轉交）。防呆：**主帳號 MUST NOT 可由店家停用或重置**；後台 `dept_admin` 以上 SHALL 仍可停用 / 重置任何店家帳號（含主帳號，重置主帳號密碼可轉交新負責人）作為逃生門。帳號管理異動 MUST 寫入稽核日誌。

#### Scenario: 店家新增子帳號

- **WHEN** 主帳號於 App「帳號管理」新增「主廚」帳號
- **THEN** 系統建立該子帳號（`account_name` = 主廚、`is_primary = false`、臨時密碼 24 小時、首登強改）並綁定同一客戶
- **AND** 寫入稽核日誌

#### Scenario: 店家停用離職子帳號

- **WHEN** 主帳號停用「主廚」子帳號
- **THEN** 該帳號立即無法登入，其他帳號不受影響

#### Scenario: 主帳號不可由店家停用或重置

- **WHEN** 店家嘗試停用或重置自己客戶的主帳號（含當前登入的帳號）
- **THEN** 系統拒絕該操作並提示主帳號須由後台管理

#### Scenario: 主帳號僅供管理、不提供業務登入

- **WHEN** 主帳號登入後嘗試呼叫下單、訂單歷史、退貨或專屬商品等業務 API
- **THEN** 系統拒絕該請求（403）
- **AND** App 僅顯示「帳號管理」畫面，不提供業務頁籤

#### Scenario: 自動附帶業務子帳號店家不可管理

- **WHEN** 店家於帳號管理頁檢視帳號清單
- **THEN** 建立客戶時自動附帶的業務子帳號顯示為「系統預設」並灰化（反白），改名 / 停用 / 重置按鈕不可用
- **AND** 店家可正常新增與管理自行建立的子帳號

#### Scenario: 後台可管理並移交業務子帳號

- **WHEN** `dept_admin` 以上於後台檢視該客戶帳號
- **THEN** 自動附帶的業務子帳號可正常改名、停用或重置（不受店家灰化限制）
- **AND** 客戶主責業務（`default_sales_rep_id`）變更時，後台可將業務子帳號移交新主責業務（重置密碼轉交）

#### Scenario: 子帳號無管理權限

- **WHEN** 子帳號嘗試新增、停用或重置其他帳號
- **THEN** 系統拒絕該操作

#### Scenario: 範圍限制僅自己客戶

- **WHEN** 店家嘗試新增、停用或重置其他客戶或員工的帳號
- **THEN** 系統拒絕（RLS self 範圍）

#### Scenario: 後台逃生門

- **WHEN** 店家帳號遭誤停用或鎖定（含主帳號），`dept_admin` 以上執行重置或重新啟用
- **THEN** 系統允許操作，店家帳號恢復可用
- **AND** 重置主帳號密碼後，可由店家轉交新的主帳號負責人
