/**
 * 待收款（未付期別）的呈現與匯出。
 *
 * 兩件事都寫成**純函式**：CSV 的字串輸出（跳脫照 RFC 4180）與狀態欄的文字 ——
 * 這份 CSV 是營運貼進試算表對帳用的，一列歪掉就會少一筆帳。
 *
 * 「逾期」**不是後端狀態**：後端只回 `open`（未付），逾期是同一個狀態再加上「期末已過」的事實
 * （見 `PlatformAdminService.ListReceivables` 的註解：服務層不改語意，由前端標記），
 * 所以 `now` 由呼叫端注入——同一個輸入永遠得到同一個答案，測試也不必跟時鐘賽跑。
 */
export type ReceivableRow = {
  companyName: string;
  planCode: string;
  periodNo: number;
  amount: string;
  periodEnd: string;
  status: string;
};

const HEADER = ["公司", "方案", "期別", "金額", "到期日", "狀態"] as const;

/** 狀態欄：後端只回 `open`（未付）；期末已過即逾期，其餘狀態原樣呈現（後端才是定義者）。 */
export function receivableStatus(
  row: { status: string; periodEnd: string },
  now = Date.now(),
): string {
  if (row.status !== "open") return row.status;
  const periodEnd = Date.parse(row.periodEnd);
  return Number.isFinite(periodEnd) && periodEnd < now ? "逾期" : "未付";
}

/**
 * 待收款 CSV：固定六欄（公司、方案、期別、金額、到期日、狀態）。
 *
 * 期別與金額原樣輸出（後端的期別是數字、金額已是 `money.FormatCents` 的兩位小數字串），
 * 欄位含逗號／雙引號／換行／前後空白時以雙引號包住、內部雙引號加倍。
 */
export function receivablesCsv(rows: readonly ReceivableRow[], now = Date.now()): string {
  const table = [
    HEADER as readonly string[],
    ...rows.map((row) => [
      row.companyName,
      row.planCode,
      String(row.periodNo),
      row.amount,
      row.periodEnd,
      receivableStatus(row, now),
    ]),
  ];
  return (
    table
      .map((cells) =>
        cells.map((v) => (/[",\r\n]|^\s|\s$/.test(v) ? `"${v.replaceAll('"', '""')}"` : v)).join(","),
      )
      .join("\n") + "\n"
  );
}

/** 觸發瀏覽器下載：副作用（Blob、錨點）只留在這裡，組字串的純函式才測得動。 */
export function downloadCsv(filename: string, csv: string): void {
  const url = URL.createObjectURL(new Blob([csv], { type: "text/csv;charset=utf-8" }));
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.click();
  URL.revokeObjectURL(url);
}
