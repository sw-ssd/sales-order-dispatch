import { splitProps, type Component, type JSX } from "solid-js";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/cn";

export const spinnerVariants = cva(
  "inline-block animate-spin rounded-lg border-2 border-current border-r-transparent text-primary",
  {
    variants: {
      size: {
        sm: "size-3",
        default: "size-4",
        lg: "size-6",
      },
    },
    defaultVariants: {
      size: "default",
    },
  }
);

export interface SpinnerProps
  extends Omit<JSX.HTMLAttributes<HTMLSpanElement>, "role">,
    VariantProps<typeof spinnerVariants> {
  /** 讀屏念出的文字（`aria-label`），預設「載入中」。 */
  label?: string;
  class?: string;
}

/** 非同步載入狀態的簡潔指示器。 */
export const Spinner: Component<SpinnerProps> = (props) => {
  const [local, rest] = splitProps(props, ["size", "label", "class"]);

  return (
    <span
      role="status"
      aria-label={local.label || "載入中"}
      class={cn(spinnerVariants({ size: local.size }), local.class)}
      {...rest}
    />
  );
};
