import { Show, splitProps, type Component, type JSX } from "solid-js";
import { Portal } from "solid-js/web";
import { Dialog as ArkDialog } from "@ark-ui/solid";
import { ScrollArea } from "./scroll-area";
import { cn } from "@/lib/cn";

/**
 * 對話框根層：行為（開關、Escape、焦點鎖定、外部點擊）一律交給 Ark UI，
 * 只把回呼收斂為純值，避免 Ark 的 `details` 物件外洩到呼叫端。
 *
 * 卡片結構取自 Tailkit（a-c-modals-01/02）：遮罩改寫為 `bg-foreground/75`、
 * 卡片改寫為 `rounded-lg border border-border bg-card shadow-xs`，
 * Tailkit 的色階字面值與深色模式專用的顏色變體一律刪除，交由 index.css 的語意 token 翻轉。
 */
export interface DialogProps
  extends Omit<ArkDialog.RootProps, "onOpenChange" | "closeOnInteractOutside"> {
  /** 是否開啟（受控） */
  open?: boolean;
  /** 開關狀態變更回呼 */
  onOpenChange?: (open: boolean) => void;
  /** 點擊對話框外部是否關閉（預設 true） */
  closeOnOutsideClick?: boolean;
}

export const Dialog: Component<DialogProps> = (props) => {
  const [local, rest] = splitProps(props, [
    "open",
    "onOpenChange",
    "closeOnOutsideClick",
    "children",
  ]);

  return (
    <ArkDialog.Root
      open={local.open}
      onOpenChange={(details) => local.onOpenChange?.(details.open)}
      closeOnInteractOutside={local.closeOnOutsideClick}
      {...rest}
    >
      {local.children}
    </ArkDialog.Root>
  );
};

/** 關閉觸發器（Tailkit modal 右上角 X 與頁尾「取消」共用）。 */
export const DialogClose = ArkDialog.CloseTrigger;

export interface DialogContentProps extends JSX.HTMLAttributes<HTMLDivElement> {
  /** 是否顯示右上角關閉鈕（預設 true） */
  showCloseButton?: boolean;
  /** 背景遮罩是否套用模糊（預設 true） */
  blur?: boolean;
  class?: string;
}

/**
 * 卡片本體。置中交給 Ark 的 `Positioner`（不再用 `translate` 手動偏移），
 * 因此行動裝置上卡片不會被 `max-h` 裁切。
 */
export const DialogContent: Component<DialogContentProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children", "showCloseButton", "blur"]);

  return (
    <Portal>
      <ArkDialog.Backdrop
        class={cn(
          "fixed inset-0 z-50 bg-foreground/75",
          local.blur !== false && "backdrop-blur-sm"
        )}
      />
      <ArkDialog.Positioner class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <ArkDialog.Content
          class={cn(
            "relative flex w-full max-w-lg flex-col overflow-hidden rounded-lg border border-border bg-card text-card-foreground shadow-xs",
            local.class
          )}
          {...rest}
        >
          <ScrollArea class="max-h-[80vh] w-full">
            <div class="space-y-4 p-6">{local.children}</div>
          </ScrollArea>
          <Show when={local.showCloseButton !== false}>
            <ArkDialog.CloseTrigger class="absolute top-3 right-3 inline-flex cursor-pointer items-center justify-center rounded-lg border border-transparent p-2 text-foreground opacity-70 transition-colors hover:border-border hover:opacity-100 focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring">
              <svg
                class="size-4"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
              <span class="sr-only">Close</span>
            </ArkDialog.CloseTrigger>
          </Show>
        </ArkDialog.Content>
      </ArkDialog.Positioner>
    </Portal>
  );
};

export interface DialogHeaderProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

export const DialogHeader: Component<DialogHeaderProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <div class={cn("flex flex-col space-y-1.5 text-center sm:text-left", local.class)} {...rest} />
  );
};

export interface DialogFooterProps extends JSX.HTMLAttributes<HTMLDivElement> {
  class?: string;
}

export const DialogFooter: Component<DialogFooterProps> = (props) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <div
      class={cn("flex flex-col-reverse pt-4 sm:flex-row sm:justify-end sm:space-x-2", local.class)}
      {...rest}
    />
  );
};

export interface DialogTitleProps {
  class?: string;
  children?: JSX.Element;
}

export const DialogTitle: Component<DialogTitleProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <ArkDialog.Title
      class={cn("text-lg leading-none font-semibold tracking-tight text-foreground", local.class)}
      {...rest}
    >
      {local.children}
    </ArkDialog.Title>
  );
};

export interface DialogDescriptionProps {
  class?: string;
  children?: JSX.Element;
}

export const DialogDescription: Component<DialogDescriptionProps> = (props) => {
  const [local, rest] = splitProps(props, ["class", "children"]);

  return (
    <ArkDialog.Description class={cn("text-sm text-muted-foreground", local.class)} {...rest}>
      {local.children}
    </ArkDialog.Description>
  );
};
