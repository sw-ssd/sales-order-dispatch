# audit-compliance 需求規格

> 來源：原 OpenSpec delta spec（OpenSpec 工作流已停用，2026-08-03 遷移至 docs/）。


## Requirements

### Requirement: 稽核日誌記錄關鍵操作

系統 SHALL 將關鍵操作寫入獨立的 `audit_logs` 表，記錄範圍至少涵蓋：登入（`login`）、登出（`logout`）、下單與訂單異動（`create` / `update`）、主檔異動（`create` / `update` / `delete`）、列印（`print`）、強制登出（`force_logout`）、角色變更（`role_change`）、取消派車（`dispatch_cancel`）、作廢（`void`）。每筆稽核記錄 MUST 包含：`company_id`、`department_id`、`user_id`（操作人）、`action`（操作類型）、`resource_type`、`resource_id`、`before_snapshot`、`after_snapshot`（異動前後摘要）、`ip_address`、`user_agent` 及操作時間。稽核日誌 MUST 與業務資料實體分開儲存，以利日後依時間分區或封存。

#### Scenario: 下單操作寫入稽核日誌

- **WHEN** 使用者成功建立一筆銷售訂單
- **THEN** `audit_logs` 新增一筆 `action` 為 `create`、`resource_type` 為訂單、`resource_id` 為該訂單 ID 的記錄
- **AND** 記錄含操作人 `user_id`、所屬 `company_id` / `department_id`、`after_snapshot`、`ip_address`、`user_agent` 與操作時間

#### Scenario: 主檔異動記錄異動前後摘要

- **WHEN** 使用者修改客戶主檔的電話欄位
- **THEN** `audit_logs` 新增一筆 `action` 為 `update` 的記錄
- **AND** `before_snapshot` 與 `after_snapshot` 分別記載修改前與修改後的內容摘要

#### Scenario: 特權操作寫入專屬 action

- **WHEN** 管理員對使用者執行強制登出、變更角色、取消派車或作廢訂單
- **THEN** `audit_logs` 分別寫入 `action` 為 `force_logout`、`role_change`、`dispatch_cancel`、`void` 的記錄
- **AND** 每筆記錄皆含操作人、目標 `resource_type` / `resource_id` 與操作時間

### Requirement: 稽核日誌與業務操作同一資料庫交易

稽核日誌 MUST 與觸發它的業務操作在同一資料庫交易中**同步寫入**，同成功或同失敗；系統 SHALL NOT 採用非同步佇列寫入稽核日誌。

#### Scenario: 業務操作成功時稽核記錄一併提交

- **WHEN** 一筆受稽核的業務操作（如下單）交易提交成功
- **THEN** 對應的稽核記錄在同一交易中可被查詢到，不存在業務成功但稽核缺失的情況

#### Scenario: 業務操作失敗時不殘留稽核記錄

- **WHEN** 受稽核的業務操作因驗證失敗或錯誤而回滾
- **THEN** 該操作對應的稽核記錄不會出現在 `audit_logs` 中

#### Scenario: 稽核寫入失敗時業務操作一併回滾

- **WHEN** 稽核記錄寫入失敗（如違反約束）
- **THEN** 觸發該稽核的業務操作整體回滾，不會留下業務已變更但無稽核軌跡的狀態

### Requirement: 稽核日誌不可刪除與查詢範圍

`audit_logs` 不適用軟刪除，系統 SHALL NOT 提供任何刪除或修改稽核記錄的操作介面；線上資料到期僅得封存（見資料保留期限）。稽核日誌查詢 MUST 依角色限制範圍：`super` 可查詢全部公司，`company_admin` 僅可查詢所屬公司；`super` 的跨公司查詢行為本身 MUST 記錄於 `audit_logs`。

#### Scenario: 任何角色皆無法刪除稽核記錄

- **WHEN** 任何角色（含 `super`）嘗試刪除或修改 `audit_logs` 中的記錄
- **THEN** 系統拒絕該操作，既有稽核記錄保持不變

#### Scenario: company_admin 僅查詢所屬公司稽核日誌

- **WHEN** `company_admin` 查詢稽核日誌
- **THEN** 結果僅包含其所屬 `company_id` 的記錄，不會回傳其他公司的稽核資料

#### Scenario: super 跨公司查詢留下稽核軌跡

- **WHEN** `super` 跨公司查詢資料
- **THEN** 該查詢行為被寫入 `audit_logs`，記錄操作人、被查詢的範圍與時間

### Requirement: 記錄類資料保留期限與到期處理

系統 SHALL 依下列期限對記錄類資料執行到期處理：`notifications` 保留 180 天後排程刪除；`audit_logs` 預設保留 **3 個月**，由 `super` / `company_admin` 於管理頁面設定（1 / 3 / 6 / 12 個月或永久），到期由排程刪除；`print_logs`（含關聯 PDF）保留 2 年後排程刪除；`print_previews`（含關聯 PDF）保留 90 天後排程刪除。排程刪除 MUST 一併清除關聯的 PDF 檔案，不殘留孤立檔案。

#### Scenario: notifications 到期排程刪除

- **WHEN** 排程作業執行且存在建立時間超過 180 天的 `notifications`
- **THEN** 這些記錄被刪除，180 天內的記錄不受影響

#### Scenario: audit_logs 到期依設定刪除

- **WHEN** `audit_logs` 記錄線上留存超過管理頁設定的保留期限（預設 3 個月）
- **THEN** 保留排程將該批記錄刪除
- **AND** 管理者變更保留期限後，到期處理依新設定執行

#### Scenario: print_logs 到期連同關聯 PDF 刪除

- **WHEN** 排程作業執行且存在建立時間超過 2 年的 `print_logs`
- **THEN** 這些記錄及其 `file_asset_id` 關聯的 PDF 檔案一併刪除

#### Scenario: print_previews 到期連同關聯 PDF 刪除

- **WHEN** 排程作業執行且存在建立時間超過 90 天的 `print_previews`
- **THEN** 這些記錄及其關聯的 PDF 檔案一併刪除

### Requirement: 軟刪除資料保留與硬刪除

採軟刪除的業務資料在刪除後 SHALL 保留 30 天；屆滿 30 天後由 `super` 執行硬刪除或封存。保留期間內資料 MUST 維持可還原狀態，且不在一般查詢結果中出現。

#### Scenario: 軟刪除 30 天內資料保留可還原

- **WHEN** 一筆客戶主檔被軟刪除未滿 30 天
- **THEN** 該記錄仍保留於資料庫、不出現在一般查詢結果，且可被還原

#### Scenario: 軟刪除屆滿 30 天由 super 處置

- **WHEN** 軟刪除記錄屆滿 30 天
- **THEN** 僅 `super` 可對其執行硬刪除或封存，硬刪除後資料自資料庫移除

### Requirement: 個人資料保護與當事人權利

系統 MUST 將客戶名稱、電話、地址、統一編號、聯絡人資料、員工帳號與 OAuth2 識別資訊視為個人資料。資料存取 SHALL 依角色與部門隔離實施最小權限。當事人權利（客戶資料查詢、更正、刪除）SHALL 由 `super` / `company_admin` 於中台操作，且相關異動 MUST 寫入稽核日誌。

#### Scenario: 最小權限隔離個資存取

- **WHEN** `staff` 嘗試查詢非其資料範圍（`data_scope`）內的客戶個資
- **THEN** 系統拒絕或過濾該查詢，不回傳範圍外的個人資料

#### Scenario: 當事人更正請求由管理角色執行

- **WHEN** 客戶要求更正其名稱或聯絡人資料
- **THEN** 由 `super` 或 `company_admin` 於中台執行更正
- **AND** 更正異動以 `update` 寫入 `audit_logs`（含 `before_snapshot` / `after_snapshot`）

#### Scenario: 一般員工無法執行當事人刪除請求

- **WHEN** `staff` 或非管理角色嘗試執行客戶資料刪除
- **THEN** 系統拒絕該操作

### Requirement: 傳輸與儲存安全

所有 API、Connect-RPC（含 server streaming）連線 MUST 強制 TLS 1.2 以上；生產環境 SHALL NOT 使用 `InsecureSkipVerify` 跳過憑證驗證。PostgreSQL 與檔案儲存 MUST 使用磁區加密。客戶密碼 MUST 以 Argon2id 雜湊儲存；員工採 OAuth2，後端 SHALL NOT 儲存員工密碼。Web session MUST 採 httpOnly cookie 搭配 CSRF token，session store 啟用 TTL；App 的 JWT MUST 儲存於 `flutter_secure_storage`。

#### Scenario: 明文 HTTP 連線被拒絕

- **WHEN** 用戶端以未加密連線存取 API 或 server streaming
- **THEN** 系統拒絕或導向 TLS 1.2 以上加密連線

#### Scenario: 客戶密碼以 Argon2id 儲存

- **WHEN** 客戶設定或變更密碼
- **THEN** 資料庫僅儲存 Argon2id 雜湊值，任何查詢皆無法取得明文密碼

#### Scenario: 員工帳號不儲存密碼

- **WHEN** 員工經 OAuth2 登入
- **THEN** 後端僅保存 `oauth_provider` / `oauth_subject` 識別資訊，不儲存密碼雜湊

#### Scenario: Web session 採 httpOnly cookie 且有 TTL

- **WHEN** 使用者於 Web 中台登入取得 session
- **THEN** session 以 httpOnly cookie 發放、非讀取操作需驗證 CSRF token
- **AND** session 於 TTL 屆期後自動失效，需重新登入

### Requirement: 應用程式安全防護

系統 MUST 使用 Ent ORM 與參數化查詢防止 SQL 注入；前端 SHALL 對輸出跳脫，Rich Text 內容 MUST 經過消毒（sanitize）後方可渲染；Web 中台所有非讀取操作 MUST 攜帶有效的 CSRF token。

#### Scenario: 惡意輸入不造成 SQL 注入

- **WHEN** 使用者在查詢或表單欄位輸入 SQL 語法片段
- **THEN** 輸入被視為純資料處理，不改變查詢語意，亦不揭露資料庫錯誤細節

#### Scenario: Rich Text 惡意腳本被消毒

- **WHEN** 使用者提交含 `<script>` 或事件屬性的 Rich Text 內容並於前端渲染
- **THEN** 危險標籤與屬性被移除或跳脫，腳本不會執行

#### Scenario: 缺少 CSRF token 的寫入請求被拒絕

- **WHEN** Web 中台收到未攜帶有效 CSRF token 的非讀取請求
- **THEN** 系統拒絕該請求，業務資料不被變更

### Requirement: 關鍵端點速率限制

系統 SHALL 對登入、QR Code 兌換、密碼重置端點實作 rate limit；超過限制次數的請求 MUST 被拒絕並回傳明確的錯誤回應。

#### Scenario: 登入端點超過嘗試上限被拒絕

- **WHEN** 同一來源在短時間內對登入端點發送超過限制次數的請求
- **THEN** 後續請求被 rate limit 拒絕，直至限制窗口重置

#### Scenario: QR Code 兌換端點超過嘗試上限被拒絕

- **WHEN** 同一來源在短時間內對 QR Code 兌換端點發送超過限制次數的請求
- **THEN** 後續兌換請求被拒絕，防止暴力列舉兌換碼

#### Scenario: 密碼重置端點超過嘗試上限被拒絕

- **WHEN** 同一來源在短時間內對密碼重置端點發送超過限制次數的請求
- **THEN** 後續重置請求被拒絕，防止重置信濫發

### Requirement: 安全標頭與依賴漏洞掃描

生產環境的 HTTP 回應 MUST 啟用 HSTS、CSP、X-Frame-Options 等安全標頭。CI 流程 SHALL 執行 `govulncheck`、`npm audit` 及 Flutter 依賴掃描（或 Dependabot），發現漏洞時 MUST 使檢查失敗以警示處理。

#### Scenario: 生產回應攜帶安全標頭

- **WHEN** 用戶端向生產環境發出請求
- **THEN** 回應標頭包含 HSTS、CSP 與 X-Frame-Options

#### Scenario: 依賴掃描發現已知漏洞時 CI 失敗

- **WHEN** CI 中的 `govulncheck` 或 `npm audit` 發現已知漏洞
- **THEN** 該次 CI 檢查標記為失敗，漏洞未被處理前不得視為通過

### Requirement: 稽核異常監控告警

系統 SHALL 對異常存取模式產生告警，至少涵蓋：大量登入失敗、非預期跨部門查詢。告警 MUST 可供維運人員及時察覺並追溯對應的稽核記錄。

#### Scenario: 大量登入失敗觸發告警

- **WHEN** 短時間內登入失敗次數超過告警門檻
- **THEN** 系統產生異常告警，且相關失敗嘗試可追溯來源

#### Scenario: 非預期跨部門查詢觸發告警

- **WHEN** 偵測到不符合角色資料範圍慣例的跨部門查詢行為
- **THEN** 系統產生異常告警，並可對照 `audit_logs` 追查操作人與查詢範圍
