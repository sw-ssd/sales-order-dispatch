import { splitProps, type Component, type JSX, type ParentComponent, Show } from "solid-js";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/cn";
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, MoreHorizontal } from "lucide-solid";

/* --- 1. Button Variants for Pagination Links ---
 *
 * 結構取自 Tailkit（a-c-pagination-01 Simple）：外框式按鈕、`px-4`／`font-semibold`，選中頁以 `bg-muted` 標示。
 * 圓角刻意不放在 base（Tailkit 的 `rounded-lg` 只加在群組頭尾）：由 PaginationContent 的
 * `[&>li:first-child>*]`／`[&>li:last-child>*]` 提供，相鄰按鈕共用邊框則由 `-space-x-px` 併攏。
 * 顏色一律語意 token，Tailkit 的色階字面值與其 `dark:` 顏色變體全部丟棄。
 */
export const paginationButtonVariants = cva(
  "inline-flex cursor-pointer select-none items-center justify-center gap-2 whitespace-nowrap font-semibold transition-colors focus-visible:outline-hidden focus-visible:ring-3 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default:
          "border border-border bg-card text-foreground hover:bg-muted data-[active=true]:bg-muted data-[active=true]:font-bold",
        outline:
          "border border-border bg-card text-foreground hover:bg-muted data-[active=true]:border-primary data-[active=true]:bg-primary/10 data-[active=true]:text-primary data-[active=true]:font-bold",
        ghost:
          "border border-transparent text-foreground hover:bg-accent hover:text-accent-foreground data-[active=true]:bg-accent data-[active=true]:text-foreground data-[active=true]:font-bold",
        flat: "border border-transparent text-foreground hover:bg-muted data-[active=true]:bg-muted data-[active=true]:text-foreground data-[active=true]:font-bold",
      },
      size: {
        default: "h-9 min-w-9 px-4 text-sm",
        sm: "h-8 min-w-8 px-3 text-xs",
        lg: "h-10 min-w-10 px-4 text-base",
        icon: "h-9 w-9 p-0 text-sm",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

/* --- 2. Pagination Root --- */
export interface PaginationProps extends JSX.HTMLAttributes<HTMLElement> {
  class?: string;
}

export const Pagination: ParentComponent<PaginationProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <nav
      role="navigation"
      aria-label="pagination"
      class={cn("mx-auto flex w-full justify-center", local.class)}
      {...rest}
    >
      {local.children}
    </nav>
  );
};

/* --- 3. PaginationContent --- */
export interface PaginationContentProps extends JSX.HTMLAttributes<HTMLUListElement> {
  class?: string;
}

/**
 * 分頁項目的群組容器：相鄰項目共用邊框（`-space-x-px`），只有群組的頭尾帶圓角，
 * 對應 Tailkit a-c-pagination-01 桌機版的 `-mr-px` + `rounded-l-lg`/`rounded-r-lg`。
 */
export const PaginationContent: ParentComponent<PaginationContentProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <ul
      class={cn(
        "flex flex-row flex-wrap items-center -space-x-px [&>li:first-child>*]:rounded-l-lg [&>li:last-child>*]:rounded-r-lg",
        local.class
      )}
      {...rest}
    >
      {local.children}
    </ul>
  );
};

/* --- 4. PaginationItem --- */
export interface PaginationItemProps extends JSX.HTMLAttributes<HTMLLIElement> {
  class?: string;
}

export const PaginationItem: ParentComponent<PaginationItemProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <li class={cn("", local.class)} {...rest}>
      {local.children}
    </li>
  );
};

/* --- 5. PaginationLink --- */
export interface PaginationLinkProps
  extends JSX.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof paginationButtonVariants> {
  isActive?: boolean;
  class?: string;
}

export const PaginationLink: ParentComponent<PaginationLinkProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "variant", "size", "isActive", "children"]);

  return (
    <button
      type="button"
      aria-current={local.isActive ? "page" : undefined}
      data-active={local.isActive ? "true" : "false"}
      class={cn(
        paginationButtonVariants({
          variant: local.variant,
          size: local.size,
        }),
        local.class
      )}
      {...rest}
    >
      {local.children}
    </button>
  );
};

/* --- 6. PaginationPrevious --- */
export interface PaginationPreviousProps
  extends JSX.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof paginationButtonVariants> {
  class?: string;
  hideText?: boolean;
}

export const PaginationPrevious: Component<PaginationPreviousProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "variant", "size", "hideText"]);

  return (
    <PaginationLink
      aria-label="Go to previous page"
      variant={local.variant}
      size={local.size}
      class={cn("gap-1 pl-3", local.hideText && "p-0", local.class)}
      {...rest}
    >
      <ChevronLeft class="size-4" />
      <Show when={!local.hideText}>
        <span>Previous</span>
      </Show>
    </PaginationLink>
  );
};

/* --- 7. PaginationNext --- */
export interface PaginationNextProps
  extends JSX.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof paginationButtonVariants> {
  class?: string;
  hideText?: boolean;
}

export const PaginationNext: Component<PaginationNextProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "variant", "size", "hideText"]);

  return (
    <PaginationLink
      aria-label="Go to next page"
      variant={local.variant}
      size={local.size}
      class={cn("gap-1 pr-3", local.hideText && "p-0", local.class)}
      {...rest}
    >
      <Show when={!local.hideText}>
        <span>Next</span>
      </Show>
      <ChevronRight class="size-4" />
    </PaginationLink>
  );
};

/* --- 8. PaginationFirst --- */
export interface PaginationFirstProps
  extends JSX.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof paginationButtonVariants> {
  class?: string;
  hideText?: boolean;
}

export const PaginationFirst: Component<PaginationFirstProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "variant", "size", "hideText"]);

  return (
    <PaginationLink
      aria-label="Go to first page"
      variant={local.variant}
      size={local.size}
      class={cn("p-0", local.class)}
      {...rest}
    >
      <ChevronsLeft class="size-4" />
      <Show when={!local.hideText}>
        <span class="sr-only">First page</span>
      </Show>
    </PaginationLink>
  );
};

/* --- 9. PaginationLast --- */
export interface PaginationLastProps
  extends JSX.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof paginationButtonVariants> {
  class?: string;
  hideText?: boolean;
}

export const PaginationLast: Component<PaginationLastProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "variant", "size", "hideText"]);

  return (
    <PaginationLink
      aria-label="Go to last page"
      variant={local.variant}
      size={local.size}
      class={cn("p-0", local.class)}
      {...rest}
    >
      <ChevronsRight class="size-4" />
      <Show when={!local.hideText}>
        <span class="sr-only">Last page</span>
      </Show>
    </PaginationLink>
  );
};

/* --- 10. PaginationEllipsis --- */
export interface PaginationEllipsisProps extends JSX.HTMLAttributes<HTMLSpanElement> {
  class?: string;
}

/**
 * 省略號：與相鄰按鈕同框（Tailkit a-c-pagination-01 用帶外框的 `div` 表示）。
 */
export const PaginationEllipsis: Component<PaginationEllipsisProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <span
      aria-hidden="true"
      class={cn(
        "flex h-9 min-w-9 items-center justify-center border border-border bg-card px-4 text-muted-foreground",
        local.class
      )}
      {...rest}
    >
      <MoreHorizontal class="size-4" />
      <span class="sr-only">More pages</span>
    </span>
  );
};

/* --- 11. PaginationSummary --- */
export interface PaginationSummaryProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

export const PaginationSummary: ParentComponent<PaginationSummaryProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <div class={cn("text-sm text-muted-foreground select-none", local.class)} {...rest}>
      {local.children}
    </div>
  );
};
