import { createSignal } from "solid-js";
import { platform } from "./api";

/** unknown=尚未探針（畫面先當未登入處理）；前端不儲存任何 token。 */
export type SessionStatus = "unknown" | "authenticated" | "anonymous";

const [status, setStatus] = createSignal<SessionStatus>("unknown");

/** 登入狀態訊號（唯讀）：只有探針與登出會改它。 */
export const sessionStatus = status;

/**
 * 探針結果的 TTL：HttpOnly cookie 的效期只有後端知道，前端只能「過一陣子回頭問一次」。
 *
 * 永久快取會讓 session 失效後（12h 到期、operator 被停用）導航仍穿過守衛，只看到頁面上的
 * 401 文案卻不回首頁——不是 fail-open（後端仍擋），但 UX 誤導。TTL 把這個窗口壓在 30 秒內，
 * 也讓一般導航不必每次都多打一個 RPC。
 */
export const PROBE_TTL_MS = 30_000;

let probe: Promise<boolean> | undefined;
let probeAt = 0;
let loggedOut = false;
// 未結項 #25 中段：探針世代計數 —— TTL 一過就同步替換 probe，舊探針晚回會覆寫新探針已寫的狀態。
// 回呼只在「自己仍是最新一代」時寫狀態、回真值；過期世代的結果直接丟棄（呼叫端拿到 false
// ＝「這個結果已不可信」，fail-closed 方向）。
let probeGen = 0;

/**
 * 單一探針：以最小的平台查詢（1 筆租戶）問後端「這個 cookie 還算數嗎」。
 *
 * 不新增認證端點、不看 cookie 內容（HttpOnly）：**唯一**可信的來源就是後端本身。
 * 任何失敗（Unauthenticated／連線失敗／非 operator 白名單）一律視為未登入
 * —— fail-closed，寧可把使用者擋在登入頁，也不要讓後續畫面以為有 session。
 * 結果快取 PROBE_TTL_MS；登出後（loggedOut）不再探針。
 */
export async function ensureSession(): Promise<boolean> {
  if (loggedOut) return false;
  if (!probe || Date.now() - probeAt >= PROBE_TTL_MS) {
    probeAt = Date.now();
    const gen = ++probeGen;
    probe = platform.listTenants({ page: 1, pageSize: 1 }).then(
      () => {
        // 未結項 #25 前半：logout 不取消已在飛行的探針 —— 其 .then 回來必須檢查 loggedOut，
        // 否則「登出 → 飛行探針成功」會把狀態翻回 authenticated（守衛仍 fail-closed、
        // 後端仍擋，但使用者會被閃回主控台再看到 401 文案）。
        if (loggedOut || gen !== probeGen) return false;
        setStatus("authenticated");
        return true;
      },
      () => {
        if (loggedOut || gen !== probeGen) return false;
        setStatus("anonymous");
        return false;
      },
    );
  }
  return probe;
}

/**
 * 登出（純前端，v1 無登出端點）。
 *
 * `platform_session` 是 HttpOnly，前端**無法**刪除它，所以這裡只把本地狀態歸零
 * （導航立刻回到登入頁），cookie 仍有效至效期結束；要真正即時撤銷，只有後端做得到：
 * 停用 `platform.operators` 該列（`operatorauth.verify` 每次請求都回查白名單）。
 * 因此登出後**不再探針**——否則 TTL 一過，後端會照 cookie 判定「仍登入」，
 * 把剛按了登出的人又放回主控台。換帳號＝關掉分頁重開（或由後端撤銷）。
 */
export function logout() {
  loggedOut = true;
  setStatus("anonymous");
  probe = Promise.resolve(false);
}

/** 測試用：清掉快取的探針結果（正式流程不需要）。世代計數同步推進 —— 讓 reset 前
 *  仍在飛行的探針變成過期世代，其回呼不再寫狀態（測試之間不互相污染）。 */
export function resetSession() {
  loggedOut = false;
  probeAt = 0;
  probeGen++;
  setStatus("unknown");
  probe = undefined;
}
