import { Code, ConnectError } from "@connectrpc/connect";
import { ErrorInfoSchema } from "./lib/proto/salesorder/v1/common_pb";

/**
 * 造一個**線路形狀**的錯誤：後端 `errcode.Code.Error` 的產物就是帶 ErrorInfo detail 的
 * ConnectError。測試要驗的是「console 顯示的是 ErrorInfo 的碼／訊息／details」，
 * 所以不能用裸 `new Error()` 代替——那會走不到正確的分支，測出來的綠是假的。
 */
export function connectErrorWithInfo(
  code: string,
  options: { message?: string; details?: Record<string, string>; connectCode?: Code } = {},
): ConnectError {
  return new ConnectError(
    options.message ?? code,
    options.connectCode ?? Code.PermissionDenied,
    undefined,
    [
      {
        desc: ErrorInfoSchema,
        value: { code, message: options.message ?? "", details: options.details ?? {} },
      },
    ],
  );
}
