import * as vali from "valibot";

/** 單位列草稿（建商品／編輯共用）。 */
export interface ProductUnitDraft {
  unitCode: string;
  conversionRate: string;
  isBase: boolean;
  sortOrder: number;
  sizeDesc: string;
}

/** 一列預設單位：基本單位、換算率恆為 1（後端 `基本單位之換算率恆為 1`）。 */
export function baseUnitDraft(): ProductUnitDraft {
  return { unitCode: "", conversionRate: "1", isBase: true, sortOrder: 0, sizeDesc: "" };
}

/** 一列換算單位：非基本單位，率待填。 */
export function derivedUnitDraft(sortOrder: number): ProductUnitDraft {
  return { unitCode: "", conversionRate: "", isBase: false, sortOrder, sizeDesc: "" };
}

/**
 * 商品表單的欄位規則。
 *
 * 只鏡射後端必填：`code` 與 `name`（`codeName` 要求兩者 trim 後非空）。
 * 分類／進貨倉／揀貨倉／說明皆可空（`validateCategoryRef`／`validateWarehouseRef`
 * 對空字串直接回 nil），啟用狀態是布林。單位另由 `validateProductUnits` 驗證
 * —— 單位是動態陣列，放在 form 之外管理（樣板 = 訂單頁的 `validateOrderItems`）。
 */
export const productSchema = vali.object({
  code: vali.pipe(vali.string(), vali.trim(), vali.nonEmpty("請輸入商品代號")),
  name: vali.pipe(vali.string(), vali.trim(), vali.nonEmpty("請輸入商品名稱")),
  description: vali.string(),
  isActive: vali.boolean(),
  categoryId: vali.string(),
  inventoryWarehouseId: vali.string(),
  pickingWarehouseId: vali.string(),
});

/**
 * 單位組驗證（對齊後端 `validateUnits`），回第一個錯誤訊息或 `undefined`（通過）：
 * - 至少一個單位；單位代碼不可為空且不可重複；
 * - **恰一個基本單位**（`isBase`）；其換算率恆為 1；
 * - 換算率須為正數（後端 `ParseRate` 拒 `Sign() <= 0`）。
 *
 * 單位代碼是否真在 metadicts `type=unit` 字典內**不在此驗證**（需查後端），
 * 由 `CreateProduct` 回的「單位字典無此值」接手。UI 只讓使用者從字典下拉選，
 * 所以這條路徑在正常操作下不會走到。
 */
export function validateProductUnits(units: ProductUnitDraft[]): string | undefined {
  if (units.length === 0) return "至少一個單位";
  const seen = new Set<string>();
  let baseCount = 0;
  for (const u of units) {
    const code = u.unitCode.trim();
    if (code === "") return "單位代碼不可為空";
    if (seen.has(code)) return `單位重複：${code}`;
    seen.add(code);
    const rate = Number(u.conversionRate.trim());
    if (!Number.isFinite(rate) || rate <= 0) return `單位 ${code} 的換算率須為正數`;
    if (u.isBase) {
      baseCount += 1;
      if (rate !== 1) return `基本單位 ${code} 的換算率恆為 1`;
    }
  }
  if (baseCount !== 1) return "商品須恰一個基本單位";
  return undefined;
}

/**
 * 單位草稿 → 後端 `ProductUnit` 欄位。
 *
 * `sizeDesc` 空字串由後端自己處理（`TrimSpace` 後為空即不寫）；`sortOrder` 直通。
 */
export function toProductUnitInput(u: ProductUnitDraft) {
  return {
    unitCode: u.unitCode.trim(),
    conversionRate: u.conversionRate.trim(),
    isBase: u.isBase,
    sortOrder: u.sortOrder,
    sizeDesc: u.sizeDesc.trim(),
  };
}
