import * as v from "valibot";

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
