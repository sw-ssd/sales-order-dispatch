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

/**
 * 欄位級 validator 工廠：`onBlur` 與 `onSubmit` 共用同一條 valibot 規則
 * （不掛 `onChange`，否則每次按鍵就標紅）。通過回 `undefined`，失敗回該欄第一則訊息。
 *
 * 以 `v.safeParse` 接入，不依賴 Standard Schema 的版本特性。
 */
export function fieldValidators(schema: v.GenericSchema<string, string>) {
  const validate = ({ value }: { value: string }): string | undefined => {
    const result = v.safeParse(schema, value);
    return result.success ? undefined : result.issues[0].message;
  };
  return { onBlur: validate, onSubmit: validate };
}
