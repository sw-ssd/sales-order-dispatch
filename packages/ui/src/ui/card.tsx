import { Match, Switch, splitProps, type Component, type JSX } from "solid-js";
import { cn } from "./lib/cn";

export interface CardProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

/**
 * Root Card container component.
 *
 * 視覺結構取自 Tailkit（a-c-cards-01/03/09：flex-col、overflow-hidden、rounded-lg、
 * shadow-xs、標題／頁尾為 bg-muted 色帶），顏色改用語意 token。
 */
export const Card: Component<CardProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <div
      class={cn(
        "flex flex-col overflow-hidden rounded-lg border border-border bg-card text-card-foreground shadow-xs",
        local.class
      )}
      {...rest}
    />
  );
};

export interface CardHeaderProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

/**
 * Header wrapper for the Card component.
 */
export const CardHeader: Component<CardHeaderProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <div
      class={cn("flex flex-col gap-1.5 bg-muted px-5 py-4", local.class)}
      {...rest}
    />
  );
};

export interface CardTitleProps extends JSX.HTMLAttributes<HTMLHeadingElement> {
  class?: string;
  /**
   * 標題層級，預設 3。
   *
   * 卡片標題固定用 `h3` 會讓「頁面 `h1` → 卡片 `h3`」跳過 `h2`，螢幕閱讀器的文件大綱
   * 因此失去一層（首頁就是這個情況）。`Card` 不該假設自己嵌在什麼層級，故由呼叫端指定。
   */
  as?: "h2" | "h3" | "h4";
}

/**
 * Main title heading for the Card header.
 */
export const CardTitle: Component<CardTitleProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "as"]);

  const classes = () => cn("font-semibold leading-none tracking-tight text-lg", local.class);
  // `Switch`／`Match` 而非三元鏈：`local.as` 是反應式讀取，寫在元件的頂層 return 運算式裡
  // 只會被求值一次（`solid/components-return-once`、`solid/reactivity`），層級改變不會反映。
  // `Match` 的條件本身是追蹤範圍，改用這裡每個分支都是完整的 JSX 字面值 ——
  // `<Dynamic>` 會讓 `JSX.HTMLAttributes` 收斂成 `IntrinsicAttributes`（`class` 消失），故不用。
  return (
    <Switch fallback={<h3 class={classes()} {...rest} />}>
      <Match when={local.as === "h2"}>
        <h2 class={classes()} {...rest} />
      </Match>
      <Match when={local.as === "h4"}>
        <h4 class={classes()} {...rest} />
      </Match>
    </Switch>
  );
};

export interface CardDescriptionProps extends JSX.HTMLAttributes<HTMLParagraphElement> {
  class?: string;
}

/**
 * Subtitle description text for the Card header.
 */
export const CardDescription: Component<CardDescriptionProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <p
      class={cn("text-sm text-muted-foreground", local.class)}
      {...rest}
    />
  );
};

export interface CardContentProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

/**
 * Main body content area for the Card.
 */
export const CardContent: Component<CardContentProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <div
      class={cn("grow p-5", local.class)}
      {...rest}
    />
  );
};

export interface CardFooterProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

/**
 * Footer action area for the Card.
 */
export const CardFooter: Component<CardFooterProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <div
      class={cn(
        "flex items-center gap-2 bg-muted px-5 py-4 text-sm text-muted-foreground",
        local.class
      )}
      {...rest}
    />
  );
};
