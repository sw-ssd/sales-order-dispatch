import { Show, splitProps, type Component, type JSX } from "solid-js";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/cn";
import { Spinner } from "./spinner";

/**
 * Class variance authority configuration for button styling variants and sizes.
 *
 * 視覺結構取自 Tailkit（a-c-buttons-01/02/03/06/08），顏色一律改寫為 index.css 的
 * 語意 token；Tailkit 的色階字面值與深色模式專用的顏色變體全部丟棄，交由 token 自動翻轉。
 */
export const buttonVariants = cva(
  "inline-flex cursor-pointer items-center justify-center gap-2 whitespace-nowrap rounded-lg border text-sm font-semibold transition-colors focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
  {
    variants: {
      variant: {
        default:
          // hover 用 /95 而非 /90：白字疊在 90% 主色上只有 4.36:1，未達 AA（95% 為 4.75:1）。
          "border-primary bg-primary text-primary-foreground hover:bg-primary/95 active:bg-primary",
        destructive:
          "border-destructive bg-destructive text-destructive-foreground hover:bg-destructive/90 active:bg-destructive",
        outline:
          "border-border bg-card text-foreground hover:bg-muted hover:shadow-xs",
        secondary:
          "border-border bg-secondary text-foreground hover:bg-secondary/80 hover:shadow-xs",
        ghost: "border-transparent hover:bg-accent hover:text-accent-foreground",
        link: "border-transparent text-primary underline-offset-4 hover:underline",
        success:
          "border-success bg-success text-success-foreground hover:bg-success/90 active:bg-success",
        warning:
          "border-warning bg-warning text-warning-foreground hover:bg-warning/90 active:bg-warning",
        info: "border-info/30 bg-info/15 text-info hover:bg-info/20 active:bg-info/15",
      },
      size: {
        xs: "h-6 px-2 text-xs",
        default: "h-9 px-4 py-2",
        sm: "h-8 px-3 text-xs",
        lg: "h-10 px-6",
        icon: "h-9 w-9",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

/**
 * Props interface for the Button component extending standard HTML button attributes.
 */
export interface ButtonProps
  extends JSX.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  /** Shows a loading spinner and prevents repeated clicks while active. */
  loading?: boolean;
  class?: string;
}

/**
 * 按鈕元件（SolidJS）：行為沿用原生 button，視覺照 Tailkit + 語意 token。
 */
export const Button: Component<ButtonProps> = (props) => {
  // Use splitProps to preserve SolidJS reactivity for destructured props
  const [local, rest] = splitProps(props, [
    "variant",
    "size",
    "class",
    "children",
    "loading",
    "disabled",
  ]);

  return (
    <button
      class={cn(buttonVariants({ variant: local.variant, size: local.size }), local.class)}
      disabled={local.disabled || local.loading}
      aria-busy={local.loading ? "true" : undefined}
      {...rest}
    >
      <Show when={local.loading}>
        <Spinner size="sm" class="text-current" />
      </Show>
      {local.children}
    </button>
  );
};
