import { Code, ConnectError } from "@connectrpc/connect";
import { ErrorInfoSchema } from "@/lib/proto/salesorder/v1/common_pb";

/**
 * 造一個**線路形狀**的錯誤：後端 `errcode.Code.Error` 的產物就是帶 `ErrorInfo` detail 的
 * ConnectError。樣板照 `platform-console/src/test-helpers.ts`（同組 runtime 已驗證）。
 *
 * 為什麼測試不能用裸 `new ConnectError("訊息", Code.X)`：那**不是**後端送出的形狀 ——
 * 沒有 `ErrorInfo` 就沒有碼、沒有 `details`，前端只能退回 connect 碼的分類文案。
 * 用假形狀寫的斷言會鎖住某個頁面剛好寫死的那句話，卻沒驗到真實路徑
 * （2026-09-26 收斂 `errorMessage` 時，9 個測試因此變紅，全是假形狀）。
 *
 * 兩種後端形狀都要能造：
 * - **帶碼**（服務層用 `errcode.X.Error(...)`）：`{ code, message, details }`。
 * - **裸訊息**（服務層用 `connect.NewError(connect.CodeX, errors.New("中文"))`，
 *   例「部門仍有使用者,無法刪除」）：傳 `undefined` 給 details 化不了，改用
 *   `bareConnectError()`。
 */
export function apiError(
  code: string,
  message: string,
  details: Record<string, string> = {},
  connectCode: Code = Code.InvalidArgument
): ConnectError {
  return new ConnectError(message, connectCode, undefined, [
    { desc: ErrorInfoSchema, value: { code, message, details } },
  ]);
}

/**
 * 後端**沒有** ErrorInfo 的形狀（服務層直接 `connect.NewError`）。
 * 純中文訊息時前端會原樣採用（那是唯一的使用者訊息來源）；`Code.Internal` 例外，
 * 那種訊息是內部字串（「缺少租戶交易(context)」），前端一律改寫。
 */
export function bareConnectError(message: string, connectCode: Code): ConnectError {
  return new ConnectError(message, connectCode);
}
