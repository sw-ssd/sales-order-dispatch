import { splitProps, type Component, type JSX } from "solid-js";
import { cn } from "@/lib/cn";

/**
 * 表格元件（SolidJS）。
 *
 * 結構取自 Tailkit（a-c-tables-01 Bordered、a-c-tables-13 In Card Alternate）：
 * 表頭底色帶、較緊湊的 cell 內距、外框與圓角留給外層容器（Card）決定（tables-13 的
 * 卡內表格本身無外框），顏色一律改寫為 index.css 的語意 token。
 */

export interface TableProps extends JSX.HTMLAttributes<HTMLTableElement> {
  class?: string;
}

/**
 * 表格外框：負責水平溢出的捲動容器。
 */
export const Table: Component<TableProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <div class="relative w-full overflow-x-auto">
      <table
        class={cn("min-w-full align-middle text-sm whitespace-nowrap", local.class)}
        {...rest}
      />
    </div>
  );
};

export interface TableHeaderProps extends JSX.HTMLAttributes<HTMLTableSectionElement> {
  class?: string;
}

/**
 * 表頭區塊。
 */
export const TableHeader: Component<TableHeaderProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return <thead class={cn("[&_tr]:border-b [&_tr]:border-border", local.class)} {...rest} />;
};

export interface TableBodyProps extends JSX.HTMLAttributes<HTMLTableSectionElement> {
  class?: string;
}

/**
 * 表格主體區塊。
 */
export const TableBody: Component<TableBodyProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return <tbody class={cn("[&_tr:last-child]:border-0", local.class)} {...rest} />;
};

export interface TableFooterProps extends JSX.HTMLAttributes<HTMLTableSectionElement> {
  class?: string;
}

/**
 * 表尾區塊（色帶底）。
 */
export const TableFooter: Component<TableFooterProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <tfoot
      class={cn(
        "border-t border-border bg-muted font-medium [&>tr]:last:border-b-0",
        local.class
      )}
      {...rest}
    />
  );
};

export interface TableRowProps extends JSX.HTMLAttributes<HTMLTableRowElement> {
  class?: string;
}

/**
 * 資料列：底線分隔 + hover 高亮（沿用改版前的互動提示，Tailkit 樣板未提供 hover 態）。
 */
export const TableRow: Component<TableRowProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <tr
      class={cn(
        "border-b border-border transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted",
        local.class
      )}
      {...rest}
    />
  );
};

export interface TableHeadProps extends JSX.ThHTMLAttributes<HTMLTableCellElement> {
  class?: string;
}

/**
 * 表頭儲存格：色帶底 + 語意前景色（Tailkit 表頭的半透明色帶映射為 `bg-muted`）。
 */
export const TableHead: Component<TableHeadProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <th
      class={cn(
        "bg-muted px-3 py-4 text-left font-semibold text-foreground [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px]",
        local.class
      )}
      {...rest}
    />
  );
};

export interface TableCellProps extends JSX.TdHTMLAttributes<HTMLTableCellElement> {
  class?: string;
}

/**
 * 資料儲存格。
 */
export const TableCell: Component<TableCellProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <td
      class={cn(
        "p-3 align-middle [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px]",
        local.class
      )}
      {...rest}
    />
  );
};

export interface TableCaptionProps extends JSX.HTMLAttributes<HTMLTableCaptionElement> {
  class?: string;
}

/**
 * 表格說明（無障礙用）。
 */
export const TableCaption: Component<TableCaptionProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <caption class={cn("mt-4 text-xs text-muted-foreground pb-2", local.class)} {...rest} />
  );
};
