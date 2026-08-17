# multi-tenancy 需求規格

> 來源：原 OpenSpec delta spec（OpenSpec 工作流已停用，2026-08-03 遷移至 docs/）。


## Requirements

### Requirement: 兩層租戶架構與資料歸屬

系統 SHALL 採用「公司（Company）→ 部門（Department）」兩層租戶架構，Department MUST 隸屬於唯一一個 Company。所有業務資料（客戶、商品、訂單、派車、單據等）MUST 帶有 `company_id` 與 `department_id` 以標示所屬租戶；資料可見範圍 SHALL 依使用者角色的資料範圍等級決定：`super` 可跨公司存取全部資料，`company_admin` 可跨部門存取但僅限自己所屬公司，其他角色僅限自己所屬部門。（RLS 與 Casbin 的機制細節屬 authorization capability，不在此規範。）

#### Scenario: 依角色決定資料可見範圍

- **GIVEN** 系統中存在 A 公司甲部門、A 公司乙部門、B 公司甲部門的訂單
- **WHEN** `super` 查詢訂單列表
- **THEN** 系統回傳三個租戶的訂單

#### Scenario: company_admin 限自己公司跨部門

- **GIVEN** 隸屬 A 公司的 `company_admin` 已登入
- **WHEN** 查詢訂單列表
- **THEN** 系統回傳 A 公司甲、乙兩部門的訂單
- **AND** 不回傳 B 公司的任何訂單

#### Scenario: 一般角色限自己部門

- **GIVEN** 隸屬 A 公司甲部門的 `dept_admin` 或 `staff` 已登入
- **WHEN** 查詢該部門的業務資料
- **THEN** 系統僅回傳 A 公司甲部門的資料
- **AND** 對 A 公司乙部門或 B 公司資料的直接存取請求被拒絕

### Requirement: 公司停用連鎖

當公司 `status` 非 `active` 時，系統 SHALL 阻擋該公司所有帳號（含員工與客戶）登入，且該公司已登入使用者的 session / token MUST 於下一次請求時失效。公司恢復為 `active` 後，帳號 SHALL 可重新登入。

#### Scenario: 停用後禁止登入

- **GIVEN** A 公司 `status` 已變更為非 `active`
- **WHEN** A 公司任一部門的員工或客戶帳號嘗試登入
- **THEN** 系統拒絕登入並提示公司已被停用
- **AND** 不簽發新的 session 或 token

#### Scenario: 已登入 session 於下次請求失效

- **GIVEN** A 公司某客戶帳號持有有效 session 且正在操作
- **WHEN** 公司被停用後該帳號發出下一個 API 請求
- **THEN** 系統判定其 session / token 已失效並拒絕該請求

#### Scenario: 重新啟用後恢復登入

- **GIVEN** A 公司曾被停用，現已恢復為 `active`
- **WHEN** A 公司帳號以正確憑證登入
- **THEN** 系統允許登入並正常簽發 session / token

### Requirement: 公司識別與公開資訊管理

系統 SHALL 提供公司識別欄位（`logo_url`、`primary_color`、`display_name`）與公開資訊欄位（`public_email`、`public_phone`、`public_address`、`terms_url`、`privacy_url` 等），供登入頁、Web 側邊欄頂部、App 首頁與關於我們頁面、單據表頭（PDF）依登入使用者所屬公司動態呈現。僅 `super` 與 `company_admin` SHALL 可編輯所屬公司的識別與公開資訊；僅 `super` SHALL 可上傳或更換 Logo 圖檔。

#### Scenario: 依公司動態呈現識別資訊

- **GIVEN** A 公司已設定 `logo_url`、`primary_color` 與 `display_name`
- **WHEN** A 公司使用者開啟登入頁、Web 側邊欄、App 首頁或列印單據
- **THEN** 各場景呈現 A 公司的 Logo、主色與簡稱，而非其他公司的識別資訊

#### Scenario: 權限允許編輯

- **WHEN** `super` 或 A 公司的 `company_admin` 更新 A 公司的公開資訊欄位
- **THEN** 系統儲存更新，後續顯示場景採用新值

#### Scenario: 權限拒絕編輯與 Logo 上傳

- **WHEN** `dept_admin` / `staff` 嘗試編輯公司識別或公開資訊
- **THEN** 系統拒絕該操作
- **AND** `company_admin` 嘗試上傳或更換 Logo 圖檔時，系統同樣拒絕（僅 `super` 可上傳 Logo）

### Requirement: 客戶編號前綴維護

每間公司 MUST 設定 `customer_code_prefix` 作為客戶編號前綴，格式 SHALL 為大寫英數 1–4 碼，且全系統唯一（使 `customer_code` 可全域定位客戶帳號）。前綴僅 `super` / `company_admin` SHALL 可於公司設定維護。修改前綴 SHALL 僅影響後續新產生的客戶編號，不影響既有客戶編號，且取號計數器不重置。（編號產生與取號行為屬 master-data capability，不在此規範。）

#### Scenario: 合法前綴設定成功

- **WHEN** `super` 或 `company_admin` 將公司前綴設為 `TY` 且全系統無其他公司使用
- **THEN** 系統儲存該前綴，後續新客戶編號以 `TY` 開頭產生

#### Scenario: 格式不符拒絕

- **WHEN** 設定前綴為小寫字母、含符號、空值或超過 4 碼
- **THEN** 系統拒絕儲存並提示格式錯誤

#### Scenario: 重複前綴拒絕且修改不追溯

- **GIVEN** A 公司前綴為 `TY` 且已存在客戶編號 `TY000123`
- **WHEN** B 公司嘗試設定前綴 `TY`
- **THEN** 系統以唯一性衝突拒絕
- **AND** 當 A 公司前綴改為 `TYX` 後，既有客戶 `TY000123` 維持不變，後續新編號改用 `TYX`，計數器續接不歸零

### Requirement: AI 整合預留與公開發現端點

每間公司 MUST 具有 `identifier`（全系統唯一的穩定公司代號，建立後不可變更）以及 `capabilities`、`public_info` 兩個 JSON 欄位，作為未來 MCP-SERVER / ACP / A2A 整合的資料來源。系統 SHALL 提供公開發現端點 `GET /api/v1/companies/public/{identifier}`，無需認證，僅回傳該公司的公開資訊與 `capabilities`；該端點 MUST NOT 暴露任何需認證的業務資源或內部欄位（如 `smtp_config`、停用公司之詳細狀態）。

#### Scenario: 免認證取得公開資訊

- **GIVEN** A 公司 `identifier` 為 `acme` 且已設定 `capabilities` 與 `public_info`
- **WHEN** 未認證的外部呼叫端請求 `GET /api/v1/companies/public/acme`
- **THEN** 系統回傳 A 公司的公開資訊與 `capabilities`
- **AND** 回應不含任何需認證的業務資料或內部敏感欄位

#### Scenario: 不存在的 identifier

- **WHEN** 請求 `GET /api/v1/companies/public/{identifier}` 且該 identifier 不存在（或公司已軟刪除）
- **THEN** 系統回傳找不到資源的錯誤，不洩漏公司是否存在以外的資訊

#### Scenario: identifier 穩定不可變更

- **WHEN** 對已建立的公司嘗試修改 `identifier`
- **THEN** 系統拒絕變更，`identifier` 維持原值

### Requirement: 公司與部門 CRUD、軟刪除與唯一性

系統 SHALL 提供公司與部門的建立、查詢、更新、刪除管理功能，僅 `super` 可建立或停用公司，`super` / `company_admin` 可管理公司內部門。公司與部門 MUST 採軟刪除（`deleted_at`），列表查詢預設排除已刪除資料。公司的 `identifier`、`tax_id`、`customer_code_prefix` SHALL 全系統唯一，且唯一性限制 SHALL 搭配軟刪除部分唯一索引（僅對未刪除列生效），使軟刪除後可重建同名資料。

#### Scenario: 建立公司與部門

- **WHEN** `super` 建立公司並於其下建立部門
- **THEN** 系統建立公司與部門主檔，部門隸屬該公司
- **AND** 公司 `status` 預設為 `active`

#### Scenario: 唯一性衝突

- **GIVEN** A 公司已使用 `tax_id` 與 `identifier`
- **WHEN** 建立或更新另一間公司使用相同的 `tax_id` 或 `identifier`
- **THEN** 系統以唯一性衝突拒絕

#### Scenario: 軟刪除與重建

- **GIVEN** A 公司已軟刪除
- **WHEN** 查詢公司列表
- **THEN** 預設結果不含 A 公司
- **AND** 使用 A 公司原 `identifier` / `tax_id` / `customer_code_prefix` 建立新公司可成功
