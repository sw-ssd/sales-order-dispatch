<!-- 由 `go generate ./internal/errcode` 產生（來源：backend/internal/errcode/codes_*.go），請勿手改。 -->

# 錯誤碼一覽

**已遷移路徑**的對外錯誤一律帶 `ErrorInfo`（`code` 與 `trace_id`）；**尚未遷移者**列於
`backend/internal/services/errcode_baseline.txt`（基線只減不增，由守門測試強制）。
**唯一真相來源**是 `backend/internal/errcode`，本檔是它的投影——要改碼表請改 `codes_*.go` 後重跑 `go generate ./internal/errcode`。

- 碼的形態為 `域-4位數`；對外 connect 碼由區段決定（見 `sectionRules`）。
- **碼發佈後不得重用或改義**，廢止只標狀態；訊息中的 `{param}` 由 `ErrorInfo.details` 帶入（前端顯示前請自行填入）。
- 共 22 碼（AUTH 7／CUST 3／PLAT 5／SYS 7）。

| 碼 | 域 | Go 常數 | connect 碼 | 訊息 | 參數 | 狀態 |
|---|---|---|---|---|---|---|
| AUTH-3001 | AUTH | `AuthRegistrationRequired` | failed_precondition | 尚未完成註冊 | — | 使用中 |
| AUTH-3002 | AUTH | `AuthTempPasswordExpired` | failed_precondition | 臨時密碼已過期，請聯繫管理員重置 | — | 使用中 |
| AUTH-3003 | AUTH | `AuthLocked` | failed_precondition | 帳號已鎖定，請於 {until} 後再試 | `until` | 使用中 |
| AUTH-3004 | AUTH | `AuthPasswordChangeRequired` | failed_precondition | 首次登入須先修改密碼 | — | 使用中 |
| AUTH-4001 | AUTH | `AuthUnauthenticated` | unauthenticated | 未登入 | — | 使用中 |
| AUTH-4002 | AUTH | `AuthCompanyInactive` | permission_denied | 所屬公司已停用 | — | 使用中 |
| AUTH-4003 | AUTH | `AuthBadCredentials` | unauthenticated | 帳號或密碼錯誤 | — | 使用中 |
| CUST-1001 | CUST | `CustomerNameRequired` | invalid_argument | 客戶名稱不可為空 | — | 使用中 |
| CUST-2001 | CUST | `CustomerCodeExists` | already_exists | 客戶編號 {code} 已存在 | `code` | 使用中 |
| CUST-3001 | CUST | `CustomerDeleted` | failed_precondition | 客戶已刪除，無法更新 | — | 使用中 |
| PLAT-3001 | PLAT | `PlatformSubscriptionInactive` | failed_precondition | 訂閱狀態不允許此操作 | — | 使用中 |
| PLAT-3002 | PLAT | `PlatformPaymentConflict` | failed_precondition | 收款衝突：{reason} | `reason` | 使用中 |
| PLAT-3003 | PLAT | `PlatformOperatorGovernance` | failed_precondition | 此操作會讓平台失去可管理性：{reason} | `reason` | 使用中 |
| PLAT-5001 | PLAT | `PlatformLimitExceeded` | failed_precondition | 已達方案上限（{used}/{limit}），請升級方案 | `used`, `limit` | 使用中 |
| PLAT-5002 | PLAT | `PlatformFeatureNotInPlan` | failed_precondition | 目前方案未包含此功能，請升級方案 | — | 使用中 |
| SYS-1001 | SYS | `SysInvalidArgument` | invalid_argument | 參數驗證失敗 | — | 使用中 |
| SYS-2001 | SYS | `SysConflict` | already_exists | 資料衝突，請確認識別碼是否已被使用 | — | 使用中 |
| SYS-3001 | SYS | `SysScopeViolation` | failed_precondition | 資料超出目前的存取範圍,無法完成此操作 | — | 使用中 |
| SYS-3002 | SYS | `SysConstraintViolation` | failed_precondition | 資料違反資料庫約束,無法完成此操作(請確認識別碼是否已被使用、參照對象是否仍存在) | — | 使用中 |
| SYS-4001 | SYS | `SysPermissionDenied` | permission_denied | 缺少權限 | — | 使用中 |
| SYS-4002 | SYS | `SysNotFound` | not_found | 資源不存在或無權存取 | — | 使用中 |
| SYS-9000 | SYS | `SysInternal` | internal | 系統忙碌，請稍後再試 | — | 使用中 |
