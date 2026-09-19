import { splitProps, type Component, type JSX } from "solid-js";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/cn";

/**
 * Label 樣式：`inline-block font-medium` 的結構與字重照 Tailkit
 * （a-c-form-elements-01/16、a-c-form-layouts-02/03），顏色改寫為語意 token。
 */
export const labelVariants = cva(
  "inline-block text-sm font-medium text-foreground select-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
);

export interface LabelProps
  extends JSX.LabelHTMLAttributes<HTMLLabelElement>,
  VariantProps<typeof labelVariants> {
  class?: string;
}

/**
 * 表單欄位標籤：靜態 `<label>`，用語意 token，供 `Field` 與 Ark `Checkbox` 搭配使用。
 */
export const Label: Component<LabelProps> = (props) => {
  /* Use splitProps to preserve SolidJS signal reactivity */
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <label
      class={cn(labelVariants(), local.class)}
      {...rest}
    >
      {local.children}
    </label>
  );
};
