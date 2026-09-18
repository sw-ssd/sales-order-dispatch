import { splitProps, type Component, type JSX } from "solid-js";
import { cn } from "@/lib/cn";

/**
 * Props interface for the Input component extending standard HTML input attributes.
 */
export interface InputProps extends JSX.InputHTMLAttributes<HTMLInputElement> {
  class?: string;
}

/**
 * Input 元件（SolidJS）：視覺結構取自 Tailkit（a-c-form-elements-01/16、a-c-form-layouts-02），
 * 顏色一律改寫為 index.css 的語意 token；Tailkit 的色階字面值與深色模式專用的顏色變體全部丟棄，
 * 交由 token 自動翻轉。
 *
 * 同時收斂 `@tailwindcss/forms` 對原生 input 的三項副作用（該外掛寫在 base layer，
 * 本元件的 utilities 一律覆蓋）：
 * 1. 外掛硬編碼 `background-color: #fff` → 以 `bg-card` 覆蓋，深色模式不再白底。
 * 2. 外掛在 `:focus` 加上 1px 藍色 ring 與同色邊框 → 以 `focus-visible:` 的
 *    `border-primary` + `ring-3 ring-primary/50` 覆蓋。
 * 3. 外掛把 placeholder 固定成外掛自帶的灰階 → 以 `placeholder:text-muted-foreground` 覆蓋。
 */
export const Input: Component<InputProps> = (props) => {
  // Use splitProps to preserve SolidJS reactivity for destructured props
  const [local, rest] = splitProps(props, ["class", "type"]);

  return (
    <input
      type={local.type || "text"}
      class={cn(
        "block w-full rounded-lg border border-border bg-card px-3 py-2 text-base text-foreground transition-colors placeholder:text-muted-foreground file:mr-4 file:rounded-sm file:border-0 file:bg-primary/10 file:px-4 file:py-2 file:text-sm file:font-semibold file:text-primary focus-visible:border-primary focus-visible:ring-3 focus-visible:ring-primary/50 focus-visible:outline-none aria-invalid:border-destructive aria-invalid:focus-visible:ring-destructive/50 disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
        local.class
      )}
      {...rest}
    />
  );
};
