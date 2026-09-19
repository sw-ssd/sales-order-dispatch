import { describe, expect, it } from "vitest";
import { fieldValidators } from "../form-helpers";
import { loginSchema } from "./schemas";

const customerCode = fieldValidators(loginSchema.entries.customerCode);
const password = fieldValidators(loginSchema.entries.password);

describe("登入表單欄位規則", () => {
  it("客戶編號：未填有繁中訊息、填了無訊息（onBlur 與 onSubmit 一致）", () => {
    expect(customerCode.onBlur({ value: "" })).toBe("請輸入客戶編號");
    expect(customerCode.onSubmit({ value: "" })).toBe("請輸入客戶編號");
    expect(customerCode.onBlur({ value: "S-001" })).toBeUndefined();
    expect(customerCode.onSubmit({ value: "S-001" })).toBeUndefined();
  });

  it("密碼：未填有繁中訊息、填了無訊息（onBlur 與 onSubmit 一致）", () => {
    expect(password.onBlur({ value: "" })).toBe("請輸入密碼");
    expect(password.onSubmit({ value: "" })).toBe("請輸入密碼");
    expect(password.onBlur({ value: "pw-12345" })).toBeUndefined();
    expect(password.onSubmit({ value: "pw-12345" })).toBeUndefined();
  });

  it("純空白視為已填（與 HTML required 的判定一致）", () => {
    expect(customerCode.onBlur({ value: " " })).toBeUndefined();
    expect(password.onSubmit({ value: " " })).toBeUndefined();
  });
});
