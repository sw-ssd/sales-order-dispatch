import * as v from "valibot";

/**
 * 登入表單（店家分頁）的欄位規則。
 *
 * 只鏡射改寫前的 HTML `required`：兩個欄位都必填、非空字串（純空白與瀏覽器
 * 對 `required` 的判定一致，視為已填）；刻意不新增客戶編號／密碼的格式驗證，
 * 格式規則只有後端知道。
 */
export const loginSchema = v.object({
  customerCode: v.pipe(v.string(), v.nonEmpty("請輸入客戶編號")),
  password: v.pipe(v.string(), v.nonEmpty("請輸入密碼")),
});
