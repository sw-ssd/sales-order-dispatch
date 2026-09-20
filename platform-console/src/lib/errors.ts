import { ConnectError } from "@connectrpc/connect";
import { CODE_MESSAGES } from "./errcode";
import { ErrorInfoSchema } from "./proto/salesorder/v1/common_pb";

/**
 * 把平台 API 的錯誤轉成可顯示的繁中字串。
 *
 * **不硬編訊息**：後端 `errcode.Error` 已把繁中訊息放在 ErrorInfo.message（碼表來源是
 * `backend/internal/errcode`）；`errcode.ts` 是同一份碼表的 TS 投影，用於訊息缺席時補位。
 * 兩者都沒有（連線失敗等）才退回 connect 的訊息。
 */
export function describeError(err: unknown): string {
  if (err instanceof ConnectError) {
    const info = err.findDetails(ErrorInfoSchema)[0];
    if (info) {
      const known = CODE_MESSAGES[info.code];
      const message = info.message || known;
      return message ? `${info.code}：${message}` : info.code;
    }
    return err.rawMessage || err.message;
  }
  return err instanceof Error ? err.message : String(err);
}
