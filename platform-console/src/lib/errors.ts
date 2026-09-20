import { ConnectError } from "@connectrpc/connect";
import { CODE_MESSAGES, ERR_SYS_PERMISSION_DENIED } from "./errcode";
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
    if (!info) return err.rawMessage || err.message;

    const message = info.message || fillTemplate(CODE_MESSAGES[info.code], info.details);
    const head = message ? `${info.code}：${message}` : info.code;
    const hint = actionHint(info);
    return hint ? `${head}（${hint}）` : head;
  }
  return err instanceof Error ? err.message : String(err);
}

/** 碼表投影的樣板（`{param}` 由 `ErrorInfo.details` 帶入）：只在後端的 message 缺席時才會用到。 */
function fillTemplate(template: string | undefined, details: { [key: string]: string }): string {
  if (!template) return "";
  return template.replace(/\{(\w+)\}/g, (whole, key: string) => details[key] ?? whole);
}

/**
 * 逐碼補「現在該做什麼」。後端的訊息說的是**發生什麼事**，可行動的那一步只有前端補得上。
 *
 * SYS-4001（缺少權限）：`requireAdmin` 只用在操作者管理（建立／停用 operator），
 * 但它同時是「角色不足」的通用碼，`details.required_role` 指出缺哪個角色。
 *
 * **console 不依角色隱藏按鈕**：前端拿不到 operator 的角色（v1 沒有「查自己」的 RPC），
 * 猜出來的隱藏會變成假的安全感——權限一律由後端判定，前端負責把拒絕說清楚。
 */
function actionHint(info: { code: string; details: { [key: string]: string } }): string {
  if (info.code !== ERR_SYS_PERMISSION_DENIED) return "";
  const role = info.details["required_role"];
  return role
    ? `此操作需要 ${role} 角色，請改用該角色的 operator 帳號`
    : "請改用具備該權限的 operator 帳號";
}
