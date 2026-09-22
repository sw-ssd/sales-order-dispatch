import * as vali from "valibot";

/** 四種單據（`internal/print/print.go` 的 `DocumentType` 常數，proto 為字串而非 enum）。 */
export const DOC_TYPES = [
  { code: "dispatch_summary", label: "單車總表" },
  { code: "delivery_note", label: "對點單" },
  { code: "picking_list", label: "揀貨單" },
  { code: "processing_list", label: "加工單" },
] as const;

export type DocType = (typeof DOC_TYPES)[number]["code"];

/** 單據選項值 → 中文標籤（表格與下拉共用）。 */
export const DOC_TYPE_LABELS: Record<string, string> = Object.fromEntries(
  DOC_TYPES.map((d) => [d.code, d.label])
);

/**
 * 列印表單的欄位規則。
 *
 * 只鏡射後端必填：`documentType`、`routeId`、`targetDate`（`parsePrintInput` 對
 * 空車次／非 YYYY-MM-DD 日期一律回 `InvalidArgument`）。`customerId`（對點單單印一店）
 * 與 `warehouseId`（揀貨單單印一倉）**僅兩種單據需填**，故不在此層強制 —— 由 `validateScope`
 * 依單據類型判斷，避免把規則寫成兩處（後端 `parsePrintInput` 才是權威）。
 */
export const printSchema = vali.object({
  documentType: vali.pipe(vali.string(), vali.nonEmpty("請選擇單據類型")),
  routeId: vali.pipe(vali.string(), vali.nonEmpty("請選擇車次")),
  targetDate: vali.pipe(vali.string(), vali.nonEmpty("請選擇出貨日期")),
  customerId: vali.string(),
  warehouseId: vali.string(),
});

/**
 * 依單據類型補齊選用選擇器，回錯誤訊息或 `undefined`（通過）。
 *
 * 對點單（`delivery_note`）需 `customer_id`、揀貨單（`picking_list`）需 `warehouse_id`
 * —— 後端 `parsePrintInput` 對這兩種單據的空選擇器會回 `InvalidArgument`
 * （「單印一店／一倉」），故前端先擋以省一次往返。
 */
export function validateScope(
  docType: string,
  customerId: string,
  warehouseId: string
): string | undefined {
  if (docType === "delivery_note" && customerId.trim() === "") {
    return "對點單需選擇店家";
  }
  if (docType === "picking_list" && warehouseId.trim() === "") {
    return "揀貨單需選擇倉別";
  }
  return undefined;
}
