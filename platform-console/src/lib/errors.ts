import { ConnectError } from "@connectrpc/connect";
import { CODE_MESSAGES, ERR_PLATFORM_PAYMENT_CONFLICT, ERR_SYS_PERMISSION_DENIED } from "./errcode";
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
 *
 * PLAT-3002（收款衝突）見 `paymentConflictHint`。
 */
function actionHint(info: { code: string; details: { [key: string]: string } }): string {
  if (info.code === ERR_SYS_PERMISSION_DENIED) {
    const role = info.details["required_role"];
    return role
      ? `此操作需要 ${role} 角色，請改用該角色的 operator 帳號`
      : "請改用具備該權限的 operator 帳號";
  }
  if (info.code === ERR_PLATFORM_PAYMENT_CONFLICT) {
    return paymentConflictHint(info.details);
  }
  return "";
}

/**
 * PLAT-3002 由 `billing.RecordPayment` 產生，**同一個碼承載三種語意**：
 * `amount_mismatch`（金額不符）、`ref_mismatch`（已付款但交易號不同）、
 * `cross_period`（同一交易號已入帳另一期）。
 * 未結項 #30：優先走結構化 `details.kind`；無 kind 時（舊版後端／唯一鍵衝突無 reason）
 * 回退中文關鍵詞分流（「已付款」／「不符」皆出自 billing 的訊息本體；都不符時通用說法）。
 *
 * 三者**都不是「不可重試」**：後端刻意不把這個碼當成不可重試的依據（見平台服務層
 * `platformWriteError` 的註解——它涵蓋的失敗比「重複收款」更廣），所以指引一律給
 * 「先確認什麼、再怎麼送」，而不是叫 operator 放棄。
 */
function paymentConflictHint(details: { [key: string]: string }): string {
  const kind = details["kind"] ?? "";
  if (kind === "ref_mismatch" || (kind === "" && (details["reason"] ?? "").includes("已付款"))) {
    return "期別已付款但交易號不同：請先確認是否重複收款；若是同一筆，沿用原交易號重送（相同交易號視為重送，不會重複入帳），溢收則改用人工對帳";
  }
  if (
    kind === "amount_mismatch" ||
    (kind === "" && (details["reason"] ?? "").includes("不符"))
  ) {
    return "金額與期別金額不符：不支援部分付款，請留空金額以採用期別快照，或依期別金額輸入（差額請記於備註）";
  }
  return "可能是同一交易號已入帳另一期，或期別狀態已變更：請重新整理確認期別狀態後再送出";
}
