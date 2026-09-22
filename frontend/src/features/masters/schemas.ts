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

/**
 * 倉別表單的欄位規則（新增／編輯共用）。
 *
 * 只鏡射後端必填：`code` 與 `name`（三個主檔共用 `codeName`，trim 後非空）。
 * `address` 可空 —— 建立時後端只在非空才寫入，更新時 presence 整欄覆寫
 * （送空字串即清除），所以這裡不加格式規則（不製造後端沒有的限制）。
 */
export const warehouseSchema = vali.object({
  code: vali.pipe(vali.string(), vali.trim(), vali.nonEmpty("請輸入倉別代號")),
  name: vali.pipe(vali.string(), vali.trim(), vali.nonEmpty("請輸入倉別名稱")),
  address: vali.string(),
  isActive: vali.boolean(),
});

/**
 * 商品分類表單的欄位規則（新增／編輯共用）。
 *
 * 只鏡射後端必填：`code` 與 `name`（`codeName`／`trimNonEmpty`）。
 * `sortOrder` 是字串表單值（送出前 `toSortOrder` 轉 `int32`），後端不驗格式
 * —— 空或非數字一律歸 0（同 `routeSchema`：不製造後端沒有的限制）。
 */
export const productCategorySchema = vali.object({
  code: vali.pipe(vali.string(), vali.trim(), vali.nonEmpty("請輸入分類代號")),
  name: vali.pipe(vali.string(), vali.trim(), vali.nonEmpty("請輸入分類名稱")),
  sortOrder: vali.string(),
  isActive: vali.boolean(),
});

/**
 * 分切規格表單的欄位規則（新增／編輯共用）。
 *
 * 只鏡射後端必填：`code` 與 `name`（`codeName`）。`kind` 是開放集
 * （metadicts `processing_kind` 背書，後端空值歸 `other`）→ 自由輸入、不驗格式。
 * `attributes` 可空：留空＝不帶（送出時歸空 Struct `{}`），非空必須 parse 成
 * **JSON 物件** —— 後端欄位是 `google.protobuf.Struct`（TS 型別 `JsonObject`），
 * 陣列／純量序列化會失敗，所以擋在欄位層：這裡驗不過，送出前就被 `isFieldsValid`
 * 攔下，不打 API。兩個 appliesTo 旗標刻意不在此驗「至少其一」：跨欄規則放單欄
 * validator 會誤導到某一欄，後端 `validateSpecFlags` 的原文由頁面 banner 呈現。
 */
export const processingSpecSchema = vali.object({
  code: vali.pipe(vali.string(), vali.trim(), vali.nonEmpty("請輸入規格代號")),
  name: vali.pipe(vali.string(), vali.trim(), vali.nonEmpty("請輸入規格名稱")),
  kind: vali.string(),
  appliesToProcessing: vali.boolean(),
  appliesToPicking: vali.boolean(),
  sortOrder: vali.string(),
  isActive: vali.boolean(),
  attributes: vali.pipe(
    vali.string(),
    vali.check((value) => {
      const raw = value.trim();
      if (raw === "") return true; // 空白＝不帶 attributes（送出時歸空 Struct）。
      try {
        const parsed: unknown = JSON.parse(raw);
        return typeof parsed === "object" && parsed !== null && !Array.isArray(parsed);
      } catch {
        return false;
      }
    }, "屬性須為 JSON 物件，或留空")
  ),
});
