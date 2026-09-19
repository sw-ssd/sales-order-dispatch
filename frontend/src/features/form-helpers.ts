import * as v from "valibot";

/**
 * 所有 `createForm` 共用的表單層選項（`createForm(() => ({ ...appFormOptions, … }))`）。
 *
 * `canSubmitWhenInvalid: true`（form-core 1.33.5 `FormApi.d.ts:142` 的官方選項）讓 `canSubmit` 恆為
 * true。少了它就會踩到 F3：`_handleSubmit` 在「`canSubmit` 為 false 且這是第一次送出
 * （`submissionAttempts <= 1`）」時直接 return（`FormApi.js:509-518`），跳過 `validateAllFields`。
 * 任一欄被互動後 blur 標紅（touched + invalid）就會讓 `canSubmit` 變 false，於是那一次送出
 * 一個欄位都不驗證（modal 重開後只標紅第一欄，其餘空的必填欄要再按一次才會出現）。
 * 設了它，送出永遠走完整驗證；無效表單仍由送出前的 `isFieldsValid` 檢查擋下
 * （`FormApi.js:524-539`），不會真的打 API。
 *
 * 三個頁面的送出鈕都不是以 `canSubmit` 停用（守門用 `isSubmitting`），故無其他行為改變。
 */
export const appFormOptions = { canSubmitWhenInvalid: true } as const;

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

/**
 * 欄位錯誤訊息只取第一則；`meta.errors` 的型別是 `unknown[]`，
 * 而 validator（`fieldValidators`）回傳的必定是字串，非字串一律不顯示。
 */
export function firstMessage(errors: unknown[]): string | undefined {
  const [first] = errors;
  return typeof first === "string" ? first : undefined;
}
