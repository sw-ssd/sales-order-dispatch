import { splitProps, type Component } from "solid-js";
import { Tabs as ArkTabs } from "@ark-ui/solid";
import { cn } from "@/lib/cn";

/**
 * 分頁根層：選取狀態、鍵盤操作、ARIA 關聯一律交給 Ark UI，
 * 只把回呼收斂為純值，避免 Ark 的 `details` 物件外洩到呼叫端。
 *
 * 結構取自 Tailkit（a-c-tabs-11 In Card Alternate）：膠囊在淺底容器內，
 * 選中膠囊改寫為 `bg-card shadow-xs`、未選中改寫為 `text-muted-foreground`；
 * Tailkit 的色階字面值與深色模式專用的顏色變體一律刪除，交由 index.css 的語意 token 翻轉。
 *
 * 一律 `lazyMount` + `unmountOnExit`：維持「未選中的 panel 不在 DOM」的既有語意，
 * 因此這兩個 prop 不開放呼叫端覆寫。
 */
export interface TabsProps
  extends Omit<ArkTabs.RootProps, "onValueChange" | "lazyMount" | "unmountOnExit"> {
  /** 目前選中的分頁值（受控） */
  value?: string;
  /** 非受控的初始分頁值 */
  defaultValue?: string;
  /** 選取變更回呼 */
  onValueChange?: (value: string) => void;
  /** 版面方向（預設 horizontal） */
  orientation?: "horizontal" | "vertical";
  class?: string;
}

export const Tabs: Component<TabsProps> = (props) => {
  const [local, rest] = splitProps(props, [
    "value",
    "defaultValue",
    "onValueChange",
    "orientation",
    "class",
    "children",
  ]);

  return (
    <ArkTabs.Root
      value={local.value}
      defaultValue={local.defaultValue}
      orientation={local.orientation}
      onValueChange={(details) => local.onValueChange?.(details.value)}
      lazyMount
      unmountOnExit
      class={cn(
        "flex w-full flex-col gap-2 data-[orientation=vertical]:flex-row data-[orientation=vertical]:gap-4",
        local.class
      )}
      {...rest}
    >
      {local.children}
    </ArkTabs.Root>
  );
};

export interface TabsListProps extends ArkTabs.ListProps {
  class?: string;
}

/** 分頁列：Tailkit 的膠囊容器改寫為 `bg-muted`。 */
export const TabsList: Component<TabsListProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <ArkTabs.List
      class={cn(
        "inline-flex h-9 items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground",
        "data-[orientation=vertical]:h-auto data-[orientation=vertical]:w-auto data-[orientation=vertical]:flex-col data-[orientation=vertical]:items-stretch",
        local.class
      )}
      {...rest}
    >
      {local.children}
    </ArkTabs.List>
  );
};

export interface TabsTriggerProps extends ArkTabs.TriggerProps {
  /** 對應的 TabsContent 值 */
  value: string;
  class?: string;
}

/** 分頁按鈕：ARIA 角色與選取狀態、鍵盤操作皆由 Ark 提供，樣式靠 `data-selected`、`data-orientation` 變體。 */
export const TabsTrigger: Component<TabsTriggerProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <ArkTabs.Trigger
      class={cn(
        "inline-flex cursor-pointer items-center justify-center gap-2 rounded-md px-3 py-1 text-sm font-medium whitespace-nowrap transition-colors",
        "ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
        "disabled:pointer-events-none disabled:opacity-50",
        "hover:bg-card/50 hover:text-foreground",
        "data-selected:bg-card data-selected:text-foreground data-selected:shadow-xs",
        "data-[orientation=vertical]:justify-start",
        local.class
      )}
      {...rest}
    >
      {local.children}
    </ArkTabs.Trigger>
  );
};

export interface TabsContentProps extends ArkTabs.ContentProps {
  /** 對應的 TabsTrigger 值 */
  value: string;
  class?: string;
}

/** 分頁內容：未選中時由 Ark 直接卸載（Root 的 lazyMount + unmountOnExit）。 */
export const TabsContent: Component<TabsContentProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <ArkTabs.Content
      class={cn(
        "mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
        "data-[orientation=vertical]:mt-0 data-[orientation=vertical]:flex-1",
        local.class
      )}
      {...rest}
    >
      {local.children}
    </ArkTabs.Content>
  );
};
