import * as v from "valibot";

/**
 * 客戶 modal 表單的欄位規則（新增與編輯共用）。
 *
 * 只鏡射後端必填：`name` 必填（純空白視為未填）；`taxId` 選填、`deletedAt` 不進表單。
 * 統一編號格式規則只有後端知道，刻意不在前端新增。
 */
export const customerSchema = v.object({
  name: v.pipe(v.string(), v.trim(), v.nonEmpty("請輸入客戶名稱")),
  taxId: v.string(),
});
