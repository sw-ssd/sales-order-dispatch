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

/**
 * 地址簿表單（新增與編輯共用）。
 *
 * 只鏡射後端必填：`recipientName` 與 `addressLine` 非空（`AddAddress` 對空白回
 * `InvalidArgument`）、`type` 只收 `shipping|billing|other`（後端 `validAddressType`）。
 * 其餘欄位選填；`isDefault` 不進欄位驗證器（checkbox 不走 `fieldValidators` 的字串簽章）。
 */
export const addressSchema = v.object({
  type: v.picklist(["shipping", "billing", "other"]),
  recipientName: v.pipe(v.string(), v.trim(), v.nonEmpty("請輸入收件人")),
  phone: v.string(),
  addressLine: v.pipe(v.string(), v.trim(), v.nonEmpty("請輸入地址")),
  city: v.string(),
  postalCode: v.string(),
});

/**
 * 聯絡人表單（新增與編輯共用）。
 *
 * `name` 必填；`email` 選填但填了就要像 email —— 後端 `AddContact` 也用
 * `contactEmailRe` 擋，前端先擋省一次必然失敗的來回（格式權威仍在後端）。
 */
export const contactSchema = v.object({
  name: v.pipe(v.string(), v.trim(), v.nonEmpty("請輸入姓名")),
  title: v.string(),
  email: v.pipe(
    v.string(),
    v.regex(/^$|^[^\s@]+@[^\s@]+\.[^\s@]+$/, "email 格式非法")
  ),
  phone: v.string(),
});

/**
 * 專屬清單表單（新增／編輯共用）。
 *
 * `productId` 只在新增時可選（後端 `UpdateCustomerProduct` 不可改 customer/product
 * —— 想換商品＝刪一列再加一列，保留原列的稽核軌跡）；`defaultQty` 是十進位文字、
 * 空＝不設預設數量，格式權威在後端 `ParseQty`（這裡只攔最常見的形狀，省一次來回）。
 */
export const customerProductSchema = v.object({
  productId: v.pipe(v.string(), v.nonEmpty("請選擇產品")),
  aliasName: v.string(),
  defaultQty: v.pipe(
    v.string(),
    v.regex(/^$|^\d+(\.\d+)?$/, "數量須為十進位數字（可小數）")
  ),
  cutNote: v.string(),
});
