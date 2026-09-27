/**
 * 時間顯示的單一來源。
 *
 * ## 為什麼需要這支函式（兩個真實缺陷）
 *
 * 1. **`slice(0, 19)` 完全忽略時區偏移**，只印字串裡那個牆上時間。後端各服務對同一個
 *    `timestamptz` 的序列化並不統一：5 個服務明確 `.UTC()`（輸出 `...Z`），其餘 15 個
 *    直接 `Format`（pgx 交回的是**伺服器本地**時區，輸出 `+08:00`）。同一個入庫瞬間，
 *    訂單頁顯示 `17:43:09`、客戶頁顯示 `01:43:09` —— 相差八小時，且兩頁相鄰。
 * 2. **這個差異取決於伺服器的 `TZ` 環境變數**（部署檔沒有釘任何 TZ）。同一份程式在
 *    `TZ=UTC` 的機器上兩者會一致、在 `TZ=Asia/Taipei` 上差八小時 —— 也就是說這是
 *    「換一台機器就變」的潛伏缺陷，不是單純的排版不一致。
 *
 * 因此：**一律解析成瞬間，再以營業時區輸出**。顯示不該取決於後端剛好序列化成哪個偏移。
 *
 * ## 為什麼是固定 +8 而不是 `Intl` 的 `Asia/Taipei`
 *
 * 後端算營業日曆用的是 `time.FixedZone("UTC+8", 8*3600)`
 * （`internal/services/sales_order_assembly.go`，出貨日調整依 UTC+8 曆）。前端若改用
 * 具名時區，等於用另一套規則描述同一件事；台北自 1979 年起無日光節約，兩者恆等。
 * 固定偏移同時不需要執行環境帶完整 ICU/tzdata，在測試與瀏覽器上結果逐位元相同。
 *
 * ## 純日期欄位不要經過這裡
 *
 * `YYYY-MM-DD` 的欄位（`expectedDeliveryDate`、`targetDate`）是**日曆日**，不是瞬間；
 * 轉時區會把它們平移一天。那些欄位原樣顯示。
 */

/** 營業時區偏移：UTC+8。與後端 `time.FixedZone("UTC+8", 8*3600)` 同一組規則。 */
const BUSINESS_OFFSET_MS = 8 * 60 * 60 * 1000;

/** 空值與無法解析時顯示的佔位字元（與各頁原本的 `"—"` 一致）。 */
export const EMPTY_DATETIME = "—";

/**
 * 把後端傳來的 RFC3339 瞬間轉成 `YYYY-MM-DD HH:mm:ss`（營業時區 UTC+8）。
 *
 * 不吃偏移的寫法（`slice(0, 19)`）會讓同一個瞬間在 `.UTC()` 與本地時區的服務之間
 * 相差八小時；這裡先用 `Date` 解析掉偏移，再加固定偏移讀 UTC 欄位。
 */
export function formatDateTime(value: string | undefined | null): string {
  if (!value) return EMPTY_DATETIME;
  const t = Date.parse(value);
  if (Number.isNaN(t)) return EMPTY_DATETIME;
  const d = new Date(t + BUSINESS_OFFSET_MS);
  const p = (n: number) => String(n).padStart(2, "0");
  return (
    `${d.getUTCFullYear()}-${p(d.getUTCMonth() + 1)}-${p(d.getUTCDate())}` +
    ` ${p(d.getUTCHours())}:${p(d.getUTCMinutes())}:${p(d.getUTCSeconds())}`
  );
}

/**
 * 只取日期（營業時區 UTC+8）的 `YYYY-MM-DD`。
 *
 * 給「後端送的是瞬間、但畫面只需要日期」的少數情況用；純日期欄位請原樣顯示。
 */
export function formatDate(value: string | undefined | null): string {
  const full = formatDateTime(value);
  return full === EMPTY_DATETIME ? full : full.slice(0, 10);
}
