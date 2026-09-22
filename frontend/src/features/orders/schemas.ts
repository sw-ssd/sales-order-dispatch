import * as vali from "valibot";

/** 單一明細草稿（建單／編輯共用）。 */
export interface OrderItemDraft {
  productId: string;
  manualName: string;
  qty: string;
  unit: string;
  processingSpecId: string;
  specialCutNote: string;
  saveAlias: boolean;
}

/** 一列空白明細。 */
export function emptyItem(): OrderItemDraft {
  return {
    productId: "",
    manualName: "",
    qty: "",
    unit: "",
    processingSpecId: "",
    specialCutNote: "",
    saveAlias: false,
  };
}

/**
 * 訂單表單（客戶／來源／出貨日／備註）的欄位規則。
 *
 * 只鏡射後端必填：`customerId`、`source` 必填（`expectedDeliveryDate` 為 `YYYY-MM-DD`
 * 且可空；格式由 `<input type="date">` 保證，不另加規則）。明細另由 `validateOrderItems`
 * 驗證 —— 明細是動態陣列，放在 form 之外管理。
 */
export const orderSchema = vali.object({
  customerId: vali.pipe(vali.string(), vali.nonEmpty("請選擇客戶")),
  source: vali.pipe(vali.string(), vali.nonEmpty("請選擇訂單來源")),
  expectedDeliveryDate: vali.string(),
  note: vali.string(),
});

/**
 * 明細陣列驗證（對齊後端 `validateOrderItem`／`resolveBaseQty` 的下限），回第一個錯誤訊息
 * 或 `undefined`（通過）：
 * - 至少一列；數量必填且為正數（後端 `ParseQty` 成功且 `Sign() > 0`）；
 * - 單位必填（後端 `unit == ""` 即 `items` 錯）；
 * - 選了商品以商品為顯示名；沒選商品即手打列，此時品名必填。
 *
 * 單位是否為該商品的合法 `unit_code`、商品是否存在**不在此驗證**（需查後端），
 * 由 `CreateOrder` 回的 `items.unit`／`items.display_name` 錯誤接手。
 */
export function validateOrderItems(items: OrderItemDraft[]): string | undefined {
  if (items.length === 0) return "至少需一項明細";
  for (const it of items) {
    const qty = it.qty.trim();
    if (qty === "") return "數量不可為空";
    const n = Number(qty);
    if (!Number.isFinite(n) || n <= 0) return "數量須為正數";
    if (it.unit.trim() === "") return "單位不可為空";
    if (it.productId === "" && it.manualName.trim() === "")
      return "未選商品時需填品名";
  }
  return undefined;
}

/**
 * 明細草稿 → 後端 `OrderItemInput` 欄位。
 *
 * 顯示名由呼叫端給（商品列＝商品名、手打列＝品名），這裡只裁空白並在手打列帶
 * `manualName`（後端視「帶 manualName」為手打，`product_id` 忽略）。
 */
export function toOrderItemInput(it: OrderItemDraft, displayName: string) {
  return {
    productId: it.productId,
    manualName: it.productId === "" ? it.manualName.trim() : "",
    displayName: displayName.trim(),
    qty: it.qty.trim(),
    unit: it.unit.trim(),
    processingSpecId: it.processingSpecId,
    specialCutNote: it.specialCutNote.trim(),
    warehouseId: "",
    saveAlias: it.saveAlias,
  };
}
