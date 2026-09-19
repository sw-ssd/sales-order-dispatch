import { render, screen } from "@solidjs/testing-library";
import { describe, expect, it } from "vitest";
import { Field, FieldError, FieldLabel } from "./field";
import { Input } from "./input";

describe("Field", () => {
  it("label 與 input 以 for/id 關聯", () => {
    render(() => (
      <Field>
        <FieldLabel for="company-name">公司名稱</FieldLabel>
        <Input id="company-name" />
      </Field>
    ));
    expect(screen.getByLabelText("公司名稱")).toBeTruthy();
  });

  it("錯誤訊息存在時 input 帶 aria-invalid 且被 aria-describedby 指到", () => {
    render(() => (
      <Field invalid>
        <FieldLabel for="company-name">公司名稱</FieldLabel>
        <Input id="company-name" />
        <FieldError>名稱必填</FieldError>
      </Field>
    ));
    const input = screen.getByLabelText("公司名稱");
    expect(input.getAttribute("aria-invalid")).toBe("true");
    const describedBy = input.getAttribute("aria-describedby");
    expect(describedBy).toBeTruthy();
    expect(document.getElementById(describedBy!)?.textContent).toContain(
      "名稱必填"
    );
  });

  it("Field 外的 FieldError（登入頁用法）仍可顯示錯誤文字", () => {
    render(() => <FieldError>帳號或密碼錯誤</FieldError>);
    expect(screen.getByText("帳號或密碼錯誤")).toBeTruthy();
  });
});
