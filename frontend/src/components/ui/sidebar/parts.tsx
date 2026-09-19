import { Collapsible, Drawer, Tooltip } from "@ark-ui/solid";
import { ChevronRight, PanelLeft, X } from "lucide-solid";
import {
  Show,
  splitProps,
  type Component,
  type ComponentProps,
  type JSX,
  type ParentComponent,
  type ValidComponent,
} from "solid-js";
import { Dynamic, Portal } from "solid-js/web";
import { Button, type ButtonProps } from "../button";
import {
  SIDEBAR_WIDTH_VARS,
  SidebarLayoutProvider,
  useSidebar,
  useSidebarCollapsed,
} from "./context";
import { cn } from "~/lib/cn";

/**
 * 側邊欄各部件的解剖取自 solid-ui 的 sidebar，行為層換成 Ark UI：
 * 行動版抽屜是 `drawer`、子選單是 `collapsible`、收合時的標籤是 `tooltip`。
 *
 * 樣式驅動一律走 `data-state|data-collapsible|data-variant|data-side` 與 `--sidebar-width*`，
 * 顏色一律用 index.css 的 `--sidebar-*` 語意 token（沒有任何顏色相關的 `dark:` 變體）。
 *
 * 兩個 v2 遺留的無障礙問題在這裡修掉：
 * 1. 側欄在行動版關閉時仍可 Tab：抽屜交給 Ark 的 `drawer`，關閉時 Presence 會把 `Positioner`
 *    與 `Content` 標成 `hidden`，內容因此離開無障礙樹與 tab 順序。
 * 2. 同一個 `<nav>` 內有兩個 `aria-current="page"`：`<nav>` 只包在 `SidebarContent` 上，
 *    品牌列所在的 `SidebarHeader` 在它之外；`aria-current="page"` 只由 `SidebarMenuButton`
 *    的 `isActive` 產生。
 */

/* --- 1. 共用外觀 --- */

/** 導覽項目的共用結構：`SidebarMenuButton` 與 `SidebarMenuSub` 的觸發列共用。 */
const MENU_BUTTON =
  "group/menu-button flex w-full cursor-pointer items-center gap-2 overflow-hidden rounded-md px-2 py-1.5 text-start text-sm font-medium text-sidebar-foreground outline-hidden transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground focus-visible:ring-3 focus-visible:ring-sidebar-ring disabled:pointer-events-none disabled:opacity-50 data-[active=true]:bg-sidebar-accent data-[active=true]:font-semibold data-[active=true]:text-sidebar-accent-foreground [&_svg]:size-4 [&_svg]:shrink-0";

/** 收合成 icon rail 時只剩圖示並置中；標籤（直接子節點的 `<span>`）收進無障礙樹。 */
const MENU_BUTTON_COLLAPSED = "justify-center px-0 [&>span]:sr-only";

/** Ark tooltip 的外觀（沿用 dialog 的浮層語彙）。 */
const TOOLTIP_CONTENT =
  "z-50 rounded-md border border-border bg-foreground px-2 py-1 text-xs font-medium text-background shadow-xs";

/* --- 2. Header / Footer / Content --- */

export interface SidebarHeaderProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

/**
 * 品牌列等置頂區塊。**刻意放在 `<nav>` 之外**：品牌若也是連到目前路徑的連結，
 * 才不會在同一個 landmark 裡產生第二個 `aria-current="page"`。
 */
export const SidebarHeader: ParentComponent<SidebarHeaderProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);
  return <div class={cn("flex shrink-0 flex-col gap-2 p-2", local.class)} {...rest}>{props.children}</div>;
};

export interface SidebarFooterProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

/** 置底區塊（帳號、版本資訊等）。 */
export const SidebarFooter: ParentComponent<SidebarFooterProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);
  return <div class={cn("mt-auto flex shrink-0 flex-col gap-2 p-2", local.class)} {...rest}>{props.children}</div>;
};

export interface SidebarContentProps extends JSX.HTMLAttributes<HTMLElement> {
  /** landmark 的無障礙名稱（預設「主要導覽」）。 */
  label?: string;
  class?: string;
}

/**
 * 導覽內容區，也是整組側邊欄唯一的 `<nav>` landmark：導覽清單放進來，品牌列與頁尾留在
 * 外面，`aria-current="page"` 因此不會在同一個 nav 裡出現兩次。
 */
export const SidebarContent: ParentComponent<SidebarContentProps> = (props) => {
  const [local, rest] = splitProps(props, ["label", "class"]);
  return (
    <nav
      aria-label={local.label ?? "主要導覽"}
      class={cn("flex min-h-0 grow flex-col gap-2 overflow-y-auto p-2", local.class)}
      {...rest}
    >
      {props.children}
    </nav>
  );
};

/* --- 3. Group --- */

export interface SidebarGroupProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

/** 一組導覽項目（可搭配標題）。 */
export const SidebarGroup: ParentComponent<SidebarGroupProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);
  return <div class={cn("flex w-full min-w-0 flex-col gap-1 py-1", local.class)} {...rest}>{props.children}</div>;
};

export interface SidebarGroupLabelProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

/** 分組標題；收合成 icon rail 時整列收進無障礙樹。 */
export const SidebarGroupLabel: ParentComponent<SidebarGroupLabelProps> = (props) => {
  const collapsed = useSidebarCollapsed();
  const [local, rest] = splitProps(props, ["class"]);
  return (
    <div
      class={cn(
        "px-2 pt-2 pb-1 text-xs font-semibold tracking-wider text-sidebar-foreground/70 uppercase",
        collapsed() && "sr-only",
        local.class,
      )}
      {...rest}
    >
      {props.children}
    </div>
  );
};

export interface SidebarGroupContentProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

/** 分組內容容器。 */
export const SidebarGroupContent: ParentComponent<SidebarGroupContentProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);
  return <div class={cn("w-full min-w-0 text-sm", local.class)} {...rest}>{props.children}</div>;
};

/* --- 4. Menu --- */

export interface SidebarMenuProps extends JSX.HTMLAttributes<HTMLUListElement> {
  class?: string;
}

/** 導覽清單。 */
export const SidebarMenu: ParentComponent<SidebarMenuProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);
  return <ul class={cn("flex w-full min-w-0 flex-col gap-1", local.class)} {...rest}>{props.children}</ul>;
};

export interface SidebarMenuItemProps extends JSX.HTMLAttributes<HTMLLIElement> {
  class?: string;
}

/** 導覽清單項目。 */
export const SidebarMenuItem: ParentComponent<SidebarMenuItemProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);
  return <li class={cn("group/menu-item relative min-w-0", local.class)} {...rest}>{props.children}</li>;
};

/** `as` 指定的元素／元件專屬屬性（`href`、TanStack `to` 等）由此直接轉發。 */
export type SidebarMenuButtonProps<As extends ValidComponent = "button"> = {
  /** 渲染的元素或元件；可傳 TanStack 的 `Link`。預設 `<button type="button">`。 */
  as?: As;
  /**
   * 是否為目前頁面：true 時標 `aria-current="page"`。同一個 `<nav>` 內只應有一個為 true
   * （品牌列請放在 `SidebarContent` 之外，見 `SidebarHeader`）。
   */
  isActive?: boolean;
  /** 收合成 icon rail 時，hover／focus 以 Ark tooltip 顯示的標籤文字。 */
  tooltip?: string;
  class?: string;
  children?: JSX.Element;
} & Omit<ComponentProps<As>, "class" | "children">;

/**
 * 導覽項目：預設渲染 `<button>`，`as` 可換成 `"a"` 或 TanStack 的 `Link` 等。
 *
 * 收合成 icon rail 時，直接子節點裡的 `<span>` 標籤會變成 `sr-only`（無障礙名稱保留），
 * 可見的標籤改由 Ark `tooltip` 提供——因此**標籤文字請包在 `<span>` 裡**。
 */
export function SidebarMenuButton<As extends ValidComponent = "button">(
  props: SidebarMenuButtonProps<As>,
): JSX.Element {
  const collapsed = useSidebarCollapsed();
  const [local, rest] = splitProps(props as SidebarMenuButtonProps, [
    "as",
    "isActive",
    "tooltip",
    "class",
    "children",
  ]);

  const element = (triggerProps?: object) => (
    <Dynamic
      component={local.as ?? "button"}
      type={local.as ? undefined : "button"}
      data-active={local.isActive ? "true" : undefined}
      aria-current={local.isActive ? "page" : undefined}
      class={cn(MENU_BUTTON, collapsed() && MENU_BUTTON_COLLAPSED, local.class)}
      {...rest}
      {...triggerProps}
    >
      {local.children}
    </Dynamic>
  );

  return (
    <Show when={collapsed() && local.tooltip} fallback={element()}>
      <Tooltip.Root positioning={{ placement: "right" }}>
        <Tooltip.Trigger
          asChild={(triggerProps) => element(triggerProps as unknown as object)}
        />
        <Portal>
          <Tooltip.Positioner>
            <Tooltip.Content class={TOOLTIP_CONTENT}>{local.tooltip}</Tooltip.Content>
          </Tooltip.Positioner>
        </Portal>
      </Tooltip.Root>
    </Show>
  );
}

/* --- 5. Sub menu（Ark collapsible） --- */

export interface SidebarMenuSubProps {
  /** 觸發列文字；收合成 icon rail 時同時作為標籤。 */
  label: string;
  /** 觸發列圖示。 */
  icon?: JSX.Element;
  /** 非受控的初始展開狀態（預設收合）。 */
  defaultOpen?: boolean;
  /** 受控展開狀態。 */
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  class?: string;
  children?: JSX.Element;
}

/**
 * 可收合的子選單：展開狀態、`aria-expanded`／`aria-controls` 與高度動畫的 CSS 變數都由 Ark
 * `collapsible` 提供。關閉時 Ark 把內容標成 `hidden`，子項目因此不在 tab 順序裡。
 */
export const SidebarMenuSub: ParentComponent<SidebarMenuSubProps> = (props) => {
  const collapsed = useSidebarCollapsed();

  return (
    <Collapsible.Root
      class={cn("flex w-full min-w-0 flex-col gap-1", props.class)}
      open={props.open}
      defaultOpen={props.defaultOpen}
      onOpenChange={(details) => props.onOpenChange?.(details.open)}
    >
      <Collapsible.Trigger class={cn(MENU_BUTTON, collapsed() && MENU_BUTTON_COLLAPSED)}>
        {props.icon}
        <span class={cn("grow truncate", collapsed() && "sr-only")}>{props.label}</span>
        <Collapsible.Indicator class="ms-auto shrink-0 transition-transform data-[state=open]:rotate-90 [&_svg]:size-4">
          <ChevronRight />
        </Collapsible.Indicator>
      </Collapsible.Trigger>
      <Collapsible.Content>
        <ul class="ms-3 flex min-w-0 flex-col gap-1 border-s border-sidebar-border ps-2">
          {props.children}
        </ul>
      </Collapsible.Content>
    </Collapsible.Root>
  );
};

export interface SidebarMenuSubItemProps extends JSX.HTMLAttributes<HTMLLIElement> {
  class?: string;
}

/** 子選單的清單項目。 */
export const SidebarMenuSubItem: ParentComponent<SidebarMenuSubItemProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);
  return <li class={cn("min-w-0", local.class)} {...rest}>{props.children}</li>;
};

/* --- 6. Sidebar 根層 --- */

export type SidebarCollapsible = "offcanvas" | "icon" | "none";
export type SidebarSide = "left" | "right";
export type SidebarVariant = "sidebar" | "floating" | "inset";

export interface SidebarProps extends Omit<JSX.HTMLAttributes<HTMLDivElement>, "class"> {
  /** 側邊欄在版面上的位置（預設 left）。 */
  side?: SidebarSide;
  /** 桌面外觀變體（預設 sidebar）。 */
  variant?: SidebarVariant;
  /**
   * 收合方式（預設 icon）：icon = 收成圖示列（標籤轉 tooltip）、offcanvas = 寬度歸零整個收起
   * （同時 `inert`，內容離開無障礙樹與 tab 順序）、none = 不可收合。
   */
  collapsible?: SidebarCollapsible;
  /** 行動版抽屜的標題（同時是 `role="dialog"` 的無障礙名稱，預設「導覽選單」）。 */
  mobileTitle?: string;
  class?: string;
}

/**
 * 側邊欄根層，三種渲染分支：
 * - `collapsible="none"`：單純一欄，帶 `--sidebar-width*`，不隨狀態改變寬度。
 * - 行動寬度：Ark `drawer`（off-canvas、Esc 關閉、焦點鎖定、關閉時內容 `hidden`；內含
 *   `CloseTrigger` 關閉鈕）。
 * - 桌面：帶 `data-state|data-collapsible|data-variant|data-side` 與 `--sidebar-width*` 的
 *   `div`，寬度隨 `state` 在 `--sidebar-width` 與 `--sidebar-width-icon` 之間切換；
 *   `collapsible="offcanvas"` 收合時寬度歸零並 `inert`。
 */
export const Sidebar: ParentComponent<SidebarProps> = (props) => {
  const context = useSidebar();
  const [local, rest] = splitProps(props, [
    "side",
    "variant",
    "collapsible",
    "mobileTitle",
    "class",
    "children",
  ]);

  const side = () => local.side ?? "left";
  const variant = () => local.variant ?? "sidebar";
  const collapsible = () => local.collapsible ?? "icon";
  /** 只有「桌面 + icon + state=collapsed」才算 icon rail；其他分支一律不隱藏標籤。 */
  const collapsed = () =>
    collapsible() === "icon" && !context.isMobile() && context.state() === "collapsed";
  const offcanvasCollapsed = () =>
    collapsible() === "offcanvas" && context.state() === "collapsed";

  const inner = () => (
    <SidebarLayoutProvider value={collapsed}>{local.children}</SidebarLayoutProvider>
  );

  const inline = () => (
    <div
      data-state="expanded"
      data-collapsible=""
      data-variant={variant()}
      data-side={side()}
      /* 與桌面分支同一組變數：`w-[var(--sidebar-width)]` 要有變數才算數，缺了會退化成內容寬。 */
      style={SIDEBAR_WIDTH_VARS}
      class={cn(
        "flex h-dvh w-[var(--sidebar-width)] shrink-0 flex-col border-sidebar-border bg-sidebar text-sidebar-foreground",
        side() === "right" ? "border-l" : "border-r",
        local.class,
      )}
      {...rest}
    >
      {inner()}
    </div>
  );

  const mobile = () => (
    <Drawer.Root
      open={context.openMobile()}
      onOpenChange={(details) => context.setOpenMobile(details.open)}
      swipeDirection={side() === "right" ? "end" : "start"}
    >
      <Drawer.Backdrop class="fixed inset-0 z-50 bg-foreground/75" />
      <Drawer.Positioner
        class={cn(
          "fixed inset-0 z-50 flex items-stretch",
          side() === "right" ? "justify-end" : "justify-start",
        )}
      >
        <Drawer.Content
          data-side={side()}
          style={SIDEBAR_WIDTH_VARS}
          class="relative flex h-full w-[var(--sidebar-width-mobile)] flex-col overflow-hidden bg-sidebar text-sidebar-foreground outline-hidden"
        >
          {/* Ark 把 `Content` 標成 `role="dialog"` 並指向 Title／Description；
              兩者不渲染就會留下指不到元素的 aria-labelledby／aria-describedby。 */}
          <Drawer.Title class="sr-only">{local.mobileTitle ?? "導覽選單"}</Drawer.Title>
          <Drawer.Description class="sr-only">行動寬度下的導覽抽屜</Drawer.Description>
          {/* Ark 的 CloseTrigger 不帶無障礙名稱，名稱由這裡給。 */}
          <Drawer.CloseTrigger
            aria-label="關閉導覽選單"
            class="absolute end-1 top-1 inline-flex h-9 w-9 cursor-pointer items-center justify-center rounded-lg text-sidebar-foreground outline-hidden transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground focus-visible:ring-3 focus-visible:ring-sidebar-ring [&_svg]:size-4"
          >
            <X />
          </Drawer.CloseTrigger>
          {inner()}
        </Drawer.Content>
      </Drawer.Positioner>
    </Drawer.Root>
  );

  const desktop = () => (
    <div
      data-state={context.state()}
      data-collapsible={collapsed() ? collapsible() : ""}
      data-variant={variant()}
      data-side={side()}
      style={SIDEBAR_WIDTH_VARS}
      /* offcanvas 收合只把寬度歸零，內容仍在 DOM；`inert` 讓它同時離開無障礙樹與 tab 順序，
         而 `hidden` 會連過場動畫一起砍掉。以 `attr:` 設「屬性」而非 property：SSR 產出的是屬性，
         jsdom（測試）也只認屬性；Solid 的 JSX 型別沒宣告 `attr:`，故在此轉型。 */
      {...({ "attr:inert": offcanvasCollapsed() || undefined } as JSX.HTMLAttributes<HTMLDivElement>)}
      class={cn(
        "group/sidebar sticky top-0 flex h-dvh w-[var(--sidebar-width)] shrink-0 flex-col border-sidebar-border bg-sidebar text-sidebar-foreground transition-[width] duration-200 ease-linear",
        side() === "right" ? "border-l" : "border-r",
        collapsed() && "w-[var(--sidebar-width-icon)]",
        offcanvasCollapsed() && "w-0 border-x-0",
        variant() === "floating" && "m-2 h-[calc(100dvh-1rem)] rounded-lg border shadow-xs",
        variant() === "inset" && "m-2 h-[calc(100dvh-1rem)] rounded-lg shadow-xs",
        local.class,
      )}
      {...rest}
    >
      {inner()}
    </div>
  );

  return (
    <Show
      when={collapsible() === "none"}
      fallback={
        <Show when={context.isMobile()} fallback={desktop()}>
          {mobile()}
        </Show>
      }
    >
      {inline()}
    </Show>
  );
};

/* --- 7. Trigger / Rail / Inset --- */

export interface SidebarTriggerProps
  extends Omit<ButtonProps, "onClick" | "variant" | "size" | "aria-label"> {
  class?: string;
}

/**
 * 收合鈕：桌面切換展開／收合，行動寬度開啟抽屜（關閉交給 Esc、遮罩或抽屜內的關閉鈕）。
 * 放在側邊欄之外（例如 Topbar），底色與尺寸由呼叫端決定。
 */
export const SidebarTrigger: Component<SidebarTriggerProps> = (props) => {
  const context = useSidebar();
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <Button
      variant="ghost"
      size="icon"
      aria-label="切換側邊欄"
      aria-expanded={context.isMobile() ? context.openMobile() : context.open()}
      onClick={() => (context.isMobile() ? context.setOpenMobile(true) : context.toggleSidebar())}
      class={local.class}
      {...rest}
    >
      {local.children ?? <PanelLeft />}
    </Button>
  );
};

export interface SidebarRailProps
  extends Omit<JSX.ButtonHTMLAttributes<HTMLButtonElement>, "class" | "onClick" | "aria-label"> {
  class?: string;
}

/**
 * 側邊欄內緣的細長熱區（solid-ui 的 `Rail`）。行為是**純 toggle 按鈕**：Task 6 的 AppShell
 * 才持有 Ark `splitter`，而 `Splitter.ResizeTrigger` 不能脫離 `Splitter.Root` 使用，因此這裡
 * 不冒充拖曳把手。滑鼠可點，但不進 tab 順序（鍵盤請用 `SidebarTrigger` 或 `mod+B`）。
 *
 * 行動版（抽屜）不渲染：那裡沒有可拖的內緣，而 toggle 只會改到桌面偏好（抽屜照樣開著）。
 */
export const SidebarRail: Component<SidebarRailProps> = (props) => {
  const context = useSidebar();
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <Show when={!context.isMobile()}>
      <button
        type="button"
        aria-label="切換側邊欄"
        tabindex="-1"
        onClick={() => context.toggleSidebar()}
        class={cn(
          "absolute inset-y-0 z-20 w-4 cursor-pointer bg-transparent transition-colors hover:bg-sidebar-accent focus-visible:ring-3 focus-visible:ring-sidebar-ring group-data-[side=left]/sidebar:end-0 group-data-[side=right]/sidebar:start-0",
          local.class,
        )}
        {...rest}
      />
    </Show>
  );
};

export interface SidebarInsetProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

/** 主要內容區：接在側邊欄旁邊，吃掉剩餘寬度。 */
export const SidebarInset: ParentComponent<SidebarInsetProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);
  return (
    <div
      class={cn("flex min-w-0 grow flex-col bg-background text-foreground", local.class)}
      {...rest}
    >
      {props.children}
    </div>
  );
};
