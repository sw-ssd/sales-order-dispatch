import { Field as ArkField, useFieldContext } from "@ark-ui/solid/field";
import { Show, splitProps, type Component, type JSX } from "solid-js";
import { cn } from "@/lib/cn";

export interface FieldProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
  /**
   * 是否為錯誤狀態。Ark `Field.Root` 由此推導 `aria-invalid`、錯誤訊息 id 與 `data-invalid`；
   * 沒有子節點偵測（Solid 的 JSX 在編譯期就會把子元件展開成 DOM，拿不到子元件型別），
   * 因此「欄位內有 `FieldError`」時必須由呼叫端明確傳 `invalid`。
   */
  invalid?: boolean;
}

/**
 * 單一欄位的包裝層（label／控件／說明／錯誤），a11y 關聯交由 Ark `Field`，
 * 版面沿用 Tailkit 表單的 `div.space-y-1`（a-c-form-elements-01、a-c-form-layouts-02）。
 * Tailkit 欄位與欄位之間的 `space-y-6` 屬頁面層版面，交由頁面自行處理。
 */
export const Field: Component<FieldProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children", "invalid"]);

  return (
    <ArkField.Root
      invalid={local.invalid ?? false}
      class={cn("grid w-full gap-1.5", local.class)}
      {...rest}
    >
      {local.children}
    </ArkField.Root>
  );
};

export interface FieldLabelProps
  extends JSX.LabelHTMLAttributes<HTMLLabelElement> {
  class?: string;
  children?: JSX.Element;
}

/** 尺寸與字重照 Tailkit（a-c-form-elements-01/16、a-c-form-layouts-02/03），顏色用語意 token。 */
const labelClass =
  "inline-block text-sm font-medium text-foreground select-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70";

/** 欄位標籤：轉發 Ark `Field.Label`，`for` 直通，維持與頁面控件的 id 關聯。 */
export const FieldLabel: Component<FieldLabelProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <ArkField.Label class={cn(labelClass, local.class)} {...rest}>
      {local.children}
    </ArkField.Label>
  );
};

export interface FieldDescriptionProps
  extends JSX.HTMLAttributes<HTMLSpanElement> {
  class?: string;
}

const descriptionClass = "block text-sm text-muted-foreground";

/**
 * 欄位說明文字：Ark `Field.HelperText`（`block` 維持原本 `<p>` 的區塊版面）。
 * 沒有 Ark context（放在 `Field` 外）時退回獨立 `<span>`，與 `FieldError` 的 fallback 對稱。
 */
export const FieldDescription: Component<FieldDescriptionProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);
  const field = useFieldContext();

  return (
    <Show
      when={field}
      fallback={
        <span class={cn(descriptionClass, local.class)} {...rest}>
          {local.children}
        </span>
      }
    >
      <ArkField.HelperText class={cn(descriptionClass, local.class)} {...rest}>
        {local.children}
      </ArkField.HelperText>
    </Show>
  );
};

export interface FieldErrorProps
  extends JSX.HTMLAttributes<HTMLSpanElement> {
  class?: string;
}

const errorClass = "block text-sm font-medium text-destructive";

/**
 * 欄位錯誤訊息：Ark `Field.ErrorText`，只在 `Field` 為 invalid 時渲染。
 * 登入頁把 `FieldError` 直接放在表單層（不在 `Field` 內），該用法沒有 Ark context，
 * 此時退回獨立的 `role="alert"`（`block` 維持原本 `<p>` 的區塊版面）。
 */
export const FieldError: Component<FieldErrorProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);
  const field = useFieldContext();

  return (
    <Show
      when={field}
      fallback={
        <span role="alert" class={cn(errorClass, local.class)} {...rest}>
          {local.children}
        </span>
      }
    >
      <ArkField.ErrorText class={cn(errorClass, local.class)} {...rest}>
        {local.children}
      </ArkField.ErrorText>
    </Show>
  );
};
