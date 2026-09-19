import { For, Show, splitProps, type Component, type JSX, type ParentComponent } from "solid-js";
import { Pagination as ArkPagination } from "@ark-ui/solid";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/cn";
import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  MoreHorizontal,
} from "lucide-solid";

/**
 * 分頁：頁碼狀態（`count`／`page`／`pageSize`）與 `api().pages` 產生的頁碼範圍一律交給 Ark UI，
 * 以下樣式部件只決定外觀（Tailkit a-c-pagination-01 的結構 + 語意 token）。
 *
 * 對外 API：`Pagination` 直接收 `count`／`page`／`pageSize`／`onPageChange`，不傳子節點時渲染
 * 「上一頁／頁碼範圍／下一頁」的預設版型；傳子節點則整組取代（部件仍可在 Ark 的狀態上自行組裝）。
 *
 * 檔案順序：先是樣式部件與預設版型，Root 放最後（它依賴前者）。
 */

/* --- 1. Button Variants for Pagination Parts ---
 *
 * 結構取自 Tailkit（a-c-pagination-01 Simple）：外框式按鈕、`px-4`／`font-semibold`，選中頁以 `bg-muted` 標示。
 * 圓角刻意不放在 base（Tailkit 的 `rounded-lg` 只加在群組頭尾）：由 PaginationContent 的
 * `[&>li:first-child>*]`／`[&>li:last-child>*]` 提供，相鄰按鈕共用邊框則由 `-space-x-px` 併攏。
 * 選中狀態讀 Ark 的 `data-selected`（只在選中時標記），顏色一律語意 token，
 * Tailkit 的色階字面值與其 `dark:` 顏色變體全部丟棄。
 */
export const paginationButtonVariants = cva(
  "inline-flex cursor-pointer select-none items-center justify-center gap-2 whitespace-nowrap font-semibold transition-colors focus-visible:outline-hidden focus-visible:ring-3 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default:
          "border border-border bg-card text-foreground hover:bg-muted data-selected:bg-muted data-selected:font-bold",
        outline:
          "border border-border bg-card text-foreground hover:bg-muted data-selected:border-primary data-selected:bg-primary/10 data-selected:text-primary data-selected:font-bold",
        ghost:
          "border border-transparent text-foreground hover:bg-accent hover:text-accent-foreground data-selected:bg-accent data-selected:text-foreground data-selected:font-bold",
        flat: "border border-transparent text-foreground hover:bg-muted data-selected:bg-muted data-selected:text-foreground data-selected:font-bold",
      },
      size: {
        default: "h-9 min-w-9 px-4 text-sm",
        sm: "h-8 min-w-8 px-3 text-xs",
        lg: "h-10 min-w-10 px-4 text-base",
        icon: "h-9 w-9 px-0 text-sm",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

/* --- 2. Pagination Content --- */
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

/* --- 3. Pagination Item（頁碼鈕） --- */
export type PaginationItemProps = Omit<ArkPagination.ItemProps, "type"> &
  VariantProps<typeof paginationButtonVariants>;

/**
 * 頁碼鈕：`type="page"` 由包裝層固定，選中狀態（`data-selected`、`aria-current="page"`）與點擊事件
 * 由 Ark 依目前頁碼決定（因此不再需要自帶 `isActive`）。
 */
export const PaginationItem: Component<PaginationItemProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "variant", "size"]);

  return (
    <ArkPagination.Item
      type="page"
      class={cn(
        paginationButtonVariants({ variant: local.variant, size: local.size }),
        local.class
      )}
      {...rest}
    />
  );
};

/* --- 4. Pagination Ellipsis --- */
export type PaginationEllipsisProps = ArkPagination.EllipsisProps;

/**
 * 省略號：與相鄰按鈕同框（Tailkit a-c-pagination-01 用帶外框的 `div` 表示）。
 * `index` 由 Ark 用於產生元素 id，頁碼範圍中每個省略號都要有不重複的值。
 */
export const PaginationEllipsis: Component<PaginationEllipsisProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <ArkPagination.Ellipsis
      aria-hidden="true"
      class={cn(
        "flex h-9 min-w-9 items-center justify-center border border-border bg-card px-4 text-muted-foreground",
        local.class
      )}
      {...rest}
    >
      <MoreHorizontal class="size-4" />
    </ArkPagination.Ellipsis>
  );
};

/* --- 5. Pagination Previous --- */
export interface PaginationPreviousProps
  extends ArkPagination.PrevTriggerProps,
    VariantProps<typeof paginationButtonVariants> {
  /** 只留圖示，文字不渲染（`px-0` 蓋掉 size variant 的左右內距，讓按鈕維持方形） */
  hideText?: boolean;
}

/**
 * 上一頁：無障礙名稱（`aria-label`）由 Ark 依 `translations` 給，首頁時 Ark 直接標 `disabled`。
 * 文字預設「上一頁」，可用 children 覆寫。
 */
export const PaginationPrevious: Component<PaginationPreviousProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "variant", "size", "hideText", "children"]);

  return (
    <ArkPagination.PrevTrigger
      class={cn(
        paginationButtonVariants({ variant: local.variant, size: local.size }),
        local.hideText ? "px-0" : "gap-1 pl-3",
        local.class
      )}
      {...rest}
    >
      <ChevronLeft class="size-4" />
      <Show when={!local.hideText}>{local.children ?? "上一頁"}</Show>
    </ArkPagination.PrevTrigger>
  );
};

/* --- 6. Pagination Next --- */
export interface PaginationNextProps
  extends ArkPagination.NextTriggerProps,
    VariantProps<typeof paginationButtonVariants> {
  /** 只留圖示，文字不渲染（`px-0` 蓋掉 size variant 的左右內距，讓按鈕維持方形） */
  hideText?: boolean;
}

/** 下一頁：與 `PaginationPrevious` 對稱，末頁時由 Ark 標 `disabled`。 */
export const PaginationNext: Component<PaginationNextProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "variant", "size", "hideText", "children"]);

  return (
    <ArkPagination.NextTrigger
      class={cn(
        paginationButtonVariants({ variant: local.variant, size: local.size }),
        local.hideText ? "px-0" : "gap-1 pr-3",
        local.class
      )}
      {...rest}
    >
      <Show when={!local.hideText}>{local.children ?? "下一頁"}</Show>
      <ChevronRight class="size-4" />
    </ArkPagination.NextTrigger>
  );
};

/* --- 7. Pagination First --- */
export type PaginationFirstProps = ArkPagination.FirstTriggerProps &
  VariantProps<typeof paginationButtonVariants>;

/** 第一頁：純圖示（無障礙名稱由 Ark 的 `firstTriggerLabel` 提供），首頁時標 `disabled`。 */
export const PaginationFirst: Component<PaginationFirstProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "variant", "size"]);

  return (
    <ArkPagination.FirstTrigger
      class={cn(
        paginationButtonVariants({ variant: local.variant, size: local.size }),
        "px-0",
        local.class
      )}
      {...rest}
    >
      <ChevronsLeft class="size-4" />
    </ArkPagination.FirstTrigger>
  );
};

/* --- 8. Pagination Last --- */
export type PaginationLastProps = ArkPagination.LastTriggerProps &
  VariantProps<typeof paginationButtonVariants>;

/** 最後一頁：與 `PaginationFirst` 對稱。 */
export const PaginationLast: Component<PaginationLastProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "variant", "size"]);

  return (
    <ArkPagination.LastTrigger
      class={cn(
        paginationButtonVariants({ variant: local.variant, size: local.size }),
        "px-0",
        local.class
      )}
      {...rest}
    >
      <ChevronsRight class="size-4" />
    </ArkPagination.LastTrigger>
  );
};

/* --- 9. Pagination Summary --- */
export interface PaginationSummaryProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

/** 範圍摘要（第 X–Y 筆,共 N 筆）：數字用等寬字，字級沿用清單頁內文基準 `text-sm`。 */
export const PaginationSummary: ParentComponent<PaginationSummaryProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <div
      class={cn("text-sm font-mono text-muted-foreground select-none", local.class)}
      {...rest}
    >
      {local.children}
    </div>
  );
};

/* --- 10. 預設版型 --- */
/**
 * 傳給 `Pagination` 的子節點會整組取代這裡。
 */
const PaginationDefaultControls: Component = () => (
  <PaginationContent>
    <li>
      <PaginationPrevious />
    </li>
    <ArkPagination.Context>
      {(api) => (
        <For each={api().pages}>
          {(page, index) => (
            <li>
              {page.type === "page" ? (
                <PaginationItem value={page.value}>{page.value}</PaginationItem>
              ) : (
                <PaginationEllipsis index={index()} />
              )}
            </li>
          )}
        </For>
      )}
    </ArkPagination.Context>
    <li>
      <PaginationNext />
    </li>
  </PaginationContent>
);

/* --- 11. Pagination Root --- */
export interface PaginationProps
  extends Omit<
    ArkPagination.RootProps,
    "count" | "page" | "pageSize" | "onPageChange" | "translations" | "class"
  > {
  /** 總筆數（後端 pagination.total） */
  count: number;
  /** 目前頁碼（1-based，受控） */
  page: number;
  /** 每頁筆數 */
  pageSize: number;
  /** 頁碼變更（回報新頁碼） */
  onPageChange: (page: number) => void;
  /** 逐項覆寫 Ark 的繁中預設字串（未給的項目沿用繁中預設） */
  translations?: NonNullable<ArkPagination.RootProps["translations"]>;
  class?: string;
}

/** Ark 的預設字串是英文（`next page` 等），本專案 UI 為繁中，故以繁中覆寫。 */
const paginationTranslations = {
  rootLabel: "分頁",
  firstTriggerLabel: "第一頁",
  prevTriggerLabel: "上一頁",
  nextTriggerLabel: "下一頁",
  lastTriggerLabel: "最後一頁",
  itemLabel: ({ page }: { page: number }) => `第 ${page} 頁`,
} satisfies NonNullable<ArkPagination.RootProps["translations"]>;

export const Pagination: ParentComponent<PaginationProps> = (props) => {
  const [local, rest] = splitProps(props, [
    "count",
    "page",
    "pageSize",
    "onPageChange",
    "translations",
    "class",
    "children",
  ]);

  return (
    <ArkPagination.Root
      count={local.count}
      page={local.page}
      pageSize={local.pageSize}
      onPageChange={(details) => local.onPageChange(details.page)}
      translations={{ ...paginationTranslations, ...local.translations }}
      class={cn("mx-auto flex w-full justify-center", local.class)}
      {...rest}
    >
      <Show when={local.children} fallback={<PaginationDefaultControls />}>
        {local.children}
      </Show>
    </ArkPagination.Root>
  );
};
