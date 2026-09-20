import { createSignal } from "solid-js";
import { platform } from "./api";

/** unknown=尚未探針（畫面先當未登入處理）；前端不儲存任何 token。 */
export type SessionStatus = "unknown" | "authenticated" | "anonymous";

const [status, setStatus] = createSignal<SessionStatus>("unknown");

/** 登入狀態訊號（唯讀）：只有探針與登出會改它。 */
export const sessionStatus = status;

let probe: Promise<boolean> | undefined;

/**
 * 單一探針：以最小的平台查詢（1 筆租戶）問後端「這個 cookie 還算數嗎」。
 *
 * 不新增認證端點、不看 cookie 內容（HttpOnly）：**唯一**可信的來源就是後端本身。
 * 任何失敗（Unauthenticated／連線失敗／非 operator 白名單）一律視為未登入
 * —— fail-closed，寧可把使用者擋在登入頁，也不要讓後續畫面以為有 session。
 * 結果在本次頁面生命週期內快取：導航不重打後端。
 */
export async function ensureSession(): Promise<boolean> {
  probe ??= platform.listTenants({ page: 1, pageSize: 1 }).then(
    () => {
      setStatus("authenticated");
      return true;
    },
    () => {
      setStatus("anonymous");
      return false;
    },
  );
  return probe;
}

/**
 * 登出（純前端，v1 無登出端點）。
 *
 * `platform_session` 是 HttpOnly，前端**無法**刪除它，所以這裡只把本地狀態歸零
 * （導航立刻回到登入頁）。cookie 仍有效至效期結束；要真正即時撤銷，只有後端做得到：
 * 停用 `platform.operators` 該列（`operatorauth.verify` 每次請求都回查白名單）。
 * 換帳號則需等 cookie 過期或由後端撤銷。
 */
export function logout() {
  setStatus("anonymous");
  probe = Promise.resolve(false);
}

/** 測試用：清掉快取的探針結果（正式流程不需要）。 */
export function resetSession() {
  setStatus("unknown");
  probe = undefined;
}
