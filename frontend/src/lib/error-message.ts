import { Code, ConnectError } from "@connectrpc/connect";
import { CODE_MESSAGES } from "./errcode";
import { ErrorInfoSchema } from "./proto/salesorder/v1/common_pb";

/**
 * 把 RPC 錯誤轉成**給作業人員看**的繁中訊息（全站唯一來源）。
 *
 * 原本 20 個頁面各自複製一份 `errorMessage()`，同一個錯誤碼在不同頁面顯示不同句子
 * （`Code.FailedPrecondition` 有四種寫法、`Code.Unavailable` 有「連線至／連線到」兩種），
 * 使用者于是無法從訊息本身判斷「這是不是同一件事」。這裡收斂成一份。
 *
 * ## 取值順序（每一條都對應一個實際觀察到的缺陷）
 *
 * 1. **後端 `ErrorInfo.details.reason`（中文者）** —— 後端把可行動的原因放在這裡
 *    （「看板需選擇部門」「該車次當日無待派訂單」「資料已變更，請重新載入」…），
 *    但它**不會**被寫進 message：`errcode.Code.Render` 只替換訊息樣板裡的 `{佔位符}`，
 *    裸的 `reason` 參數留在 details。於是前端若只讀 message，使用者只看到樣板原文
 *    （SYS-1001「參數驗證失敗」），真正的線索被丟掉。
 *    只接受含中文的值：`reason` 也承載機器可讀的 token（`already_inactive`、
 *    `primary_account`…），那些不該直接端到使用者面前。
 * 2. **後端 message（與碼表樣板不同者）** —— 已經渲染過、帶實際數值
 *    （例「已達方案上限（9/10），請升級方案」）。與樣板逐字相同時代表沒有補充資訊，跳過。
 * 3. **碼表文案（`CODE_MESSAGES`，代入 `details`）** —— 這是預設值。碼表的字串比
 *    connect 碼的分類文案精確得多：`SYS-3001`（超出存取範圍）與 `SYS-3002`（違反約束）
 *    同樣是 `FailedPrecondition`，用分類文案會說成「目前狀態不允許此操作」，語意走鐘。
 *    少數對使用者不具行動指引的碼列在 `OWN_COPY_CODES`，改用第 4 條。
 *    代入後仍含 `{` 的（後端沒給該參數）不可用，會把 `{used}` 原樣印給使用者。
 * 4. **本檔的 connect 碼分類文案（`BY_CODE`）** —— 通用退路。
 * 5. 退回 connect 的 `rawMessage`，最後才是「操作失敗，請稍後再試」。
 *
 * **不顯示錯誤碼**：租戶端使用者是店家與業務，`SYS-1001` 對他們沒有意義
 * （平台營運主控台才顯示碼，operator 要用它開單）。追查用的 trace id 在 ErrorInfo 裡，
 * 由客服依時間反查。
 *
 * ## 標點
 * 一律全形（`，`、`（`）。zh-Hant 的句子用半形逗號是排版缺陷，也讓 200% 縮放下的
 * 字距忽寬忽窄。
 */

/**
 * 選項：僅在需要保留**領域特定**說明時使用。
 *
 * `permissionHint` 只覆寫 `PermissionDenied` 的文案。只有公告頁用它：那裡的拒絕原因不是
 * 「角色不足」而是「這則公告的範圍不在你的管理範圍內」，通用文案會把可修復的線索弄丟。
 * 其餘情況不要傳 —— 每個頁面各寫一份說明就是這支函式被抽出來的原因。
 */
export interface ErrorMessageOptions {
  permissionHint?: string;
}

/** 後端以裸 Connect 錯誤（沒有 ErrorInfo）回覆、但訊息本身是寫給使用者看的中文。 */
const CJK = /[\u4e00-\u9fff]/;

/** 這些碼代表伺服器端故障，其訊息是給工程師的（「缺少租戶交易(context)」），不可外洩。 */
const ENGINEERING_CODES: readonly Code[] = [Code.Internal, Code.Unknown, Code.DataLoss];

export function errorMessage(err: unknown, options?: ErrorMessageOptions): string {
  if (!(err instanceof ConnectError)) {
    return NETWORK_FALLBACK;
  }
  if (err.code === Code.PermissionDenied && options?.permissionHint) {
    return fullWidthPunctuation(options.permissionHint);
  }
  const info = err.findDetails(ErrorInfoSchema)[0];
  if (info) {
    const reason = info.details["reason"] ?? "";
    if (CJK.test(reason)) return fullWidthPunctuation(reason);
    // 後端渲染過的訊息，且與碼表樣板不同 → 它帶了實際數值（例「已達方案上限（9/10）」）。
    if (info.message && info.message !== CODE_MESSAGES[info.code]) {
      return fullWidthPunctuation(info.message);
    }
    // 碼表文案：絕大多數比 connect 碼的分類文案精確 —— SYS-3001（超出存取範圍）與
    // SYS-3002（違反約束）同為 FailedPrecondition，用分類文案會說成「狀態不允許」，
    // 語意完全走鐘。故預設採用碼表，只有少數幾個碼例外（見 OWN_COPY_CODES）。
    const template = CODE_MESSAGES[info.code];
    if (template && !OWN_COPY_CODES.has(info.code)) {
      const filled = fillTemplate(template, info.details);
      // 佔位符沒填滿時（`{used}` 會原樣印給使用者）寧可退回通用文案。
      if (!filled.includes("{")) return fullWidthPunctuation(filled);
    }
  } else if (!ENGINEERING_CODES.includes(err.code) && CJK.test(err.rawMessage)) {
    // 沒有 ErrorInfo，但訊息本身是中文：後端服務層手寫的使用者訊息
    // （「部門仍有使用者,無法刪除」「公司仍設定客戶編號前綴」「客戶編號取號衝突,請稍後重試」…）。
    // 這些是後端唯一的使用者訊息來源，比 connect 碼的分類文案具體。
    //
    // 限定「非工程碼 + 含中文」兩個條件：`Internal` 的裸訊息是「缺少租戶交易(context)」
    // 「開啟租戶交易失敗」這類內部字串（見 backend 各 service），絕不可外洩；
    // 純 ASCII 的（"address id 格式錯誤" 的英文部分、SQL 片段）也不採用。
    //
    // 只在此分支採用（有 ErrorInfo 時不採用）：ErrorInfo 是後端的正式契約，
    // 兩者都在時以結構化的那份為準，否則 SYS-1001 又會退回「參數驗證失敗」。
    return fullWidthPunctuation(err.rawMessage);
  }
  return BY_CODE[err.code] ?? UNKNOWN_FALLBACK;
}

/**
 * 只把「前後都是中文」的 ASCII 逗號換成全形。
 *
 * 後端仍有以 `connect.NewError` 直接寫訊息的呼叫點（如「部門仍有使用者,無法刪除」）——
 * 那些字串不經 registry，逗號是半形。與其逐一改動後端（那是另一件事），在出口統一。
 * 條件限定「兩側都是中文」才安全：`1,000`、`(identifier)` 這類不會被動到。
 */
function fullWidthPunctuation(message: string): string {
  return message.replace(/(?<=[\u4e00-\u9fff]),(?=[\u4e00-\u9fff])/g, "，");
}

/** 碼表樣板的參數代入（`{param}` ← `details`）；與平台主控台的 `fillTemplate` 同法。 */
function fillTemplate(template: string, details: { [key: string]: string }): string {
  return template.replace(/\{(\w+)\}/g, (whole, key: string) => details[key] ?? whole);
}

/**
 * 這幾個碼改用本檔文案，因為碼表那句對**租戶端使用者**不具行動指引。
 *
 * 這是逐碼判斷，不是可推導的規則 —— 碼表的字串同時餵給 server log、平台營運主控台與
 * App，寫法是為那些場合取捨的（SYS-1001 的「參數驗證失敗」對寫 log 的人剛好，對店家
 * 等於沒說）。改動前請先確認碼表那句真的不適合使用者，不要為了統一而統一。
 */
const OWN_COPY_CODES = new Set([
  // 「參數驗證失敗」→ 說出該做什麼。
  "SYS-1001",
  // 「缺少權限」→ 加上可以找誰／做什麼。
  "SYS-4001",
  // 「資源不存在或無權存取」→ 列表頁的使用者看到的是「資料被刪了」還是「打錯 id」。
  "SYS-4002",
]);

/** 連線層失敗：`Unavailable` 之外的網路錯誤（fetch 失敗、CORS、後端未啟動）也走這句。 */
const NETWORK_FALLBACK = "無法連線至伺服器，請確認後端服務已啟動";

/** 伺服器端故障（5xx）：使用者無能為力，只需知道不是自己操作錯、且可以再試。 */
const SERVER_FALLBACK = "伺服器暫時無法使用，請稍後再試";

/** 其餘未知失敗：可行動的下一步只有「再試一次」。 */
const UNKNOWN_FALLBACK = "操作失敗，請稍後再試";

/**
 * connect 碼 → 使用者文案。
 *
 * `Unavailable` 用「無法連線至伺服器」而非「無法連線到伺服器」：後者是登入頁的舊寫法，
 * 同一個意思不該有兩種說法。
 */
const BY_CODE: Partial<Record<Code, string>> = {
  [Code.InvalidArgument]: "輸入資料有誤，請檢查後再試",
  [Code.NotFound]: "資料不存在或已被刪除",
  [Code.AlreadyExists]: "資料已存在",
  [Code.FailedPrecondition]: "目前狀態不允許此操作",
  [Code.PermissionDenied]: "沒有權限執行此操作",
  [Code.Unauthenticated]: "請先登入",
  [Code.Unavailable]: NETWORK_FALLBACK,
  // 5xx 是伺服器端的問題，不是使用者的操作錯 —— 說成「操作失敗」會讓人重試同一個動作
  // 卻不知道該等。內部字串（「缺少租戶交易(context)」）一律不外洩，只說明現況與下一步。
  [Code.Internal]: SERVER_FALLBACK,
  [Code.Unknown]: SERVER_FALLBACK,
  [Code.DataLoss]: SERVER_FALLBACK,
};
