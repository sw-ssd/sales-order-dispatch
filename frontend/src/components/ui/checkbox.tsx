import { splitProps, type Component, type JSX } from "solid-js";
import { Checkbox as ArkCheckbox } from "@ark-ui/solid";
import { cn } from "@/lib/cn";

/**
 * 勾選框：勾選狀態機、鍵盤操作與 ARIA 一律交給 Ark UI
 * （`Root` 渲染 `<label>`、`HiddenInput` 才是真正的 `<input type="checkbox">`），
 * 只把回呼收斂為純值，避免 Ark 的 `details` 物件外洩到呼叫端。
 *
 * 結構取自 Tailkit（a-c-form-elements-16 Form Elements: Disabled，Checkboxes Inline）：
 * 方框 `size-4 rounded-sm`、勾選態為 `bg-primary` + `text-primary-foreground`；
 * Tailkit 的色階字面值與深色模式專用的顏色變體一律刪除，交由 index.css 的語意 token 翻轉。
 *
 * 視覺焦點在鍵盤操作時落在被隱藏的 `HiddenInput` 上，所以焦點環畫在 `Control`，
 * 用 `peer` 指向同層、置於 `Control` 之前的 `HiddenInput`。
 * `HiddenInput` 另加 `sr-only`：Ark 只給它 1px 的行內樣式，但 `@tailwindcss/forms`
 * 會為原生 `input[type=checkbox]` 上底色（`#fff`／`checked` 的 checkmark）與 focus ring
 * （2px 藍框），`sr-only` 的裁切讓這些繪製完全不落地，只留下 Ark 的隱形輸入。
 */
export interface CheckboxProps
  extends Omit<
    JSX.LabelHTMLAttributes<HTMLLabelElement>,
    "onChange" | "class" | "children"
  > {
  /** 受控勾選狀態 */
  checked?: boolean;
  /** 非受控的初始勾選狀態 */
  defaultChecked?: boolean;
  /** 勾選變更回呼（純值） */
  onCheckedChange?: (checked: boolean) => void;
  disabled?: boolean;
  class?: string;
}

export const Checkbox: Component<CheckboxProps> = (props) => {
  const [local, rest] = splitProps(props, [
    "checked",
    "defaultChecked",
    "onCheckedChange",
    "disabled",
    "class",
  ]);

  return (
    <ArkCheckbox.Root
      checked={local.checked}
      defaultChecked={local.defaultChecked}
      disabled={local.disabled}
      onCheckedChange={(details) => local.onCheckedChange?.(details.checked === true)}
      class={cn(
        "inline-flex cursor-pointer items-center gap-2 data-[disabled]:cursor-not-allowed",
        local.class
      )}
      {...rest}
    >
      <ArkCheckbox.HiddenInput class="peer sr-only" />
      <ArkCheckbox.Control
        class={cn(
          "flex size-4 shrink-0 items-center justify-center rounded-sm border border-primary text-primary-foreground transition-colors",
          "peer-focus-visible:ring-1 peer-focus-visible:ring-ring peer-focus-visible:outline-none",
          "data-[state=checked]:bg-primary",
          "data-[disabled]:opacity-50"
        )}
      >
        <ArkCheckbox.Indicator>
          <svg
            class="size-3.5 fill-none stroke-current stroke-[3]"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M5 13l4 4L19 7"
            />
          </svg>
        </ArkCheckbox.Indicator>
      </ArkCheckbox.Control>
    </ArkCheckbox.Root>
  );
};
