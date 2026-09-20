// 由 `go generate ./internal/errcode` 產生（來源：backend/internal/errcode/codes_*.go），請勿手改。
// 碼的唯一真相來源是後端 registry；本檔只是投影，CI 會驗證同步（.github/workflows/ci.yml）。

/** 對外錯誤碼（`ErrorInfo.code`；形如 "域-4位數"）。 */
export const ERR_AUTH_REGISTRATION_REQUIRED = "AUTH-3001";
export const ERR_AUTH_TEMP_PASSWORD_EXPIRED = "AUTH-3002";
export const ERR_AUTH_LOCKED = "AUTH-3003";
export const ERR_AUTH_PASSWORD_CHANGE_REQUIRED = "AUTH-3004";
export const ERR_AUTH_UNAUTHENTICATED = "AUTH-4001";
export const ERR_AUTH_COMPANY_INACTIVE = "AUTH-4002";
export const ERR_AUTH_BAD_CREDENTIALS = "AUTH-4003";
export const ERR_CUSTOMER_NAME_REQUIRED = "CUST-1001";
export const ERR_CUSTOMER_CODE_EXISTS = "CUST-2001";
export const ERR_CUSTOMER_DELETED = "CUST-3001";
export const ERR_PLATFORM_SUBSCRIPTION_INACTIVE = "PLAT-3001";
export const ERR_PLATFORM_PAYMENT_CONFLICT = "PLAT-3002";
export const ERR_PLATFORM_LIMIT_EXCEEDED = "PLAT-5001";
export const ERR_PLATFORM_FEATURE_NOT_IN_PLAN = "PLAT-5002";
export const ERR_SYS_INVALID_ARGUMENT = "SYS-1001";
export const ERR_SYS_CONFLICT = "SYS-2001";
export const ERR_SYS_SCOPE_VIOLATION = "SYS-3001";
export const ERR_SYS_CONSTRAINT_VIOLATION = "SYS-3002";
export const ERR_SYS_PERMISSION_DENIED = "SYS-4001";
export const ERR_SYS_NOT_FOUND = "SYS-4002";
export const ERR_SYS_INTERNAL = "SYS-9000";

/** 碼 → 繁中訊息樣板（`{param}` 由 `ErrorInfo.details` 帶入）。 */
export const CODE_MESSAGES: Record<string, string> = {
  "AUTH-3001": "尚未完成註冊",
  "AUTH-3002": "臨時密碼已過期，請聯繫管理員重置",
  "AUTH-3003": "帳號已鎖定，請於 {until} 後再試",
  "AUTH-3004": "首次登入須先修改密碼",
  "AUTH-4001": "未登入",
  "AUTH-4002": "所屬公司已停用",
  "AUTH-4003": "帳號或密碼錯誤",
  "CUST-1001": "客戶名稱不可為空",
  "CUST-2001": "客戶編號 {code} 已存在",
  "CUST-3001": "客戶已刪除，無法更新",
  "PLAT-3001": "訂閱狀態不允許此操作",
  "PLAT-3002": "收款衝突：{reason}",
  "PLAT-5001": "已達方案上限（{used}/{limit}），請升級方案",
  "PLAT-5002": "目前方案未包含此功能，請升級方案",
  "SYS-1001": "參數驗證失敗",
  "SYS-2001": "資料衝突，請確認識別碼是否已被使用",
  "SYS-3001": "資料超出目前的存取範圍,無法完成此操作",
  "SYS-3002": "資料違反資料庫約束,無法完成此操作(請確認識別碼是否已被使用、參照對象是否仍存在)",
  "SYS-4001": "缺少權限",
  "SYS-4002": "資源不存在或無權存取",
  "SYS-9000": "系統忙碌，請稍後再試",
};
