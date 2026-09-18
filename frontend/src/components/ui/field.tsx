import { splitProps, type Component, type JSX } from "solid-js";
import { cn } from "@/lib/cn";
import { Label, type LabelProps } from "./label";

export interface FieldProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

/**
 * 單一欄位的包裝層（label／控件／說明／錯誤），對應 Tailkit 表單裡的
 * `div.space-y-1`（a-c-form-elements-01、a-c-form-layouts-02）。
 * Tailkit 欄位與欄位之間的 `space-y-6` 屬頁面層版面，交由頁面自行處理。
 */
export const Field: Component<FieldProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <div class={cn("grid w-full gap-1.5", local.class)} {...rest}>
      {local.children}
    </div>
  );
};

export interface FieldLabelProps extends LabelProps {
  class?: string;
  children?: JSX.Element;
}

/** 欄位標籤：尺寸與字重由 `Label`（Tailkit `inline-block font-medium`）提供。 */
export const FieldLabel: Component<FieldLabelProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <Label class={local.class} {...rest}>
      {local.children}
    </Label>
  );
};

export interface FieldDescriptionProps
  extends JSX.HTMLAttributes<HTMLParagraphElement> {
  class?: string;
}

/** 欄位說明文字：Tailkit 的次要灰階 → `text-muted-foreground`。 */
export const FieldDescription: Component<FieldDescriptionProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <p class={cn("text-sm text-muted-foreground", local.class)} {...rest}>
      {local.children}
    </p>
  );
};

export interface FieldErrorProps
  extends JSX.HTMLAttributes<HTMLParagraphElement> {
  class?: string;
}

/** 欄位錯誤訊息：搭配 `Input` 的 `aria-invalid:border-destructive` 使用。 */
export const FieldError: Component<FieldErrorProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <p
      role="alert"
      class={cn("text-sm font-medium text-destructive", local.class)}
      {...rest}
    >
      {local.children}
    </p>
  );
};
