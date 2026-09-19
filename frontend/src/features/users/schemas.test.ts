import { describe, expect, it } from "vitest";
import { fieldValidators } from "../form-helpers";
import { companySchema } from "./schemas";

const name = fieldValidators(companySchema.entries.name);
const identifier = fieldValidators(companySchema.entries.identifier);

describe("公司 modal 欄位規則", () => {
  it("公司名稱：未填有繁中訊息、填了無訊息（onBlur 與 onSubmit 一致）", () => {
    expect(name.onBlur({ value: "" })).toBe("請輸入公司名稱");
    expect(name.onSubmit({ value: "" })).toBe("請輸入公司名稱");
    expect(name.onBlur({ value: "測試公司" })).toBeUndefined();
    expect(name.onSubmit({ value: "測試公司" })).toBeUndefined();
  });

  it("識別碼：未填有繁中訊息、填了無訊息（onBlur 與 onSubmit 一致）", () => {
    expect(identifier.onBlur({ value: "" })).toBe("請輸入識別碼(identifier)");
    expect(identifier.onSubmit({ value: "" })).toBe("請輸入識別碼(identifier)");
    expect(identifier.onBlur({ value: "C-001" })).toBeUndefined();
    expect(identifier.onSubmit({ value: "C-001" })).toBeUndefined();
  });

  it("純空白視為未填（改寫前是 trim 後再檢查）", () => {
    expect(name.onSubmit({ value: "   " })).toBe("請輸入公司名稱");
    expect(identifier.onSubmit({ value: " " })).toBe("請輸入識別碼(identifier)");
  });

  it("選填欄位（統一編號、狀態）不會產生錯誤", () => {
    const taxId = fieldValidators(companySchema.entries.taxId);
    const status = fieldValidators(companySchema.entries.status);
    expect(taxId.onSubmit({ value: "" })).toBeUndefined();
    expect(taxId.onSubmit({ value: "12345678" })).toBeUndefined();
    expect(status.onSubmit({ value: "" })).toBeUndefined();
    expect(status.onSubmit({ value: "active" })).toBeUndefined();
  });
});
