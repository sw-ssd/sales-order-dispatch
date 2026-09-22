import * as vali from "valibot";

/**
 * 車次表單的欄位規則（新增／編輯共用）。
 *
 * 只鏡射後端必填：`code` 與 `name`（`codeName` 要求兩者 trim 後非空）。
 * `description`／`sortOrder`／`isActive` 皆可空可省 —— 後端 `CreateRoute` 只驗 code/name，
 * `sort_order` 是 int32 但**不驗格式**（`route_service.go` 無驗證路徑），
 * 所以這裡也不加格式規則：空或非數字一律歸 0（不製造後端沒有的限制）。
 */
export const routeSchema = vali.object({
  code: vali.pipe(vali.string(), vali.trim(), vali.nonEmpty("請輸入車次代號")),
  name: vali.pipe(vali.string(), vali.trim(), vali.nonEmpty("請輸入車次名稱")),
  description: vali.string(),
  sortOrder: vali.string(),
  isActive: vali.boolean(),
});

/**
 * 表單字串型排序 → 後端 `int32 sort_order`。
 *
 * 非數字與空字串都歸 0（0 是合法順位，且後端不驗格式 —— 見 `routeSchema` 註解）。
 */
export function toSortOrder(raw: string): number {
  const n = Number(raw.trim());
  return Number.isFinite(n) ? Math.trunc(n) : 0;
}
