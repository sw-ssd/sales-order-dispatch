import { splitProps, type Component, type JSX } from "solid-js";
import { cn } from "@/lib/cn";

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
}

/**
 * Main title heading for the Card header.
 */
export const CardTitle: Component<CardTitleProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <h3
      class={cn("font-semibold leading-none tracking-tight text-lg", local.class)}
      {...rest}
    />
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
