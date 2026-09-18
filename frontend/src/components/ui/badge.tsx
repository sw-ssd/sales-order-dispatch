import { splitProps, type Component, type JSX } from "solid-js";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/cn";

/**
 * Class variance authority configuration for badge styling variants.
 *
 * 視覺結構取自 Tailkit（a-c-badges-01/03：rounded-sm、px-2、py-1、text-xs、leading-4、
 * 無邊框軟底），顏色一律改寫為 index.css 的語意 token；色階字面值與深色模式顏色變體一併刪除。
 */
export const badgeVariants = cva(
  "inline-flex items-center rounded-sm px-2 py-1 text-xs leading-4 font-semibold focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground",
        secondary: "bg-muted text-muted-foreground",
        destructive: "bg-destructive text-destructive-foreground",
        outline: "border border-border text-foreground",
        success: "bg-success/15 text-success",
        warning: "bg-warning/15 text-warning",
        info: "bg-primary/10 text-primary",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

/**
 * Props interface for the Badge component extending standard HTML div attributes.
 */
export interface BadgeProps
  extends JSX.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {
  class?: string;
}

/**
 * 徽章元件（SolidJS）：靜態標籤，視覺照 Tailkit + 語意 token。
 */
export const Badge: Component<BadgeProps> = (props) => {
  // Use splitProps to preserve SolidJS reactivity
  const [local, rest] = splitProps(props, ["variant", "class", "children"]);

  return (
    <div class={cn(badgeVariants({ variant: local.variant }), local.class)} {...rest}>
      {local.children}
    </div>
  );
};
