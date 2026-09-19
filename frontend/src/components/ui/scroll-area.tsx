import { ScrollArea as ArkScrollArea } from "@ark-ui/solid";
import { Show, splitProps, type Component, type JSX } from "solid-js";
import { cn } from "@/lib/cn";

/**
 * 捲動容器：量測、拖曳、鍵盤與滾輪行為全部交給 Ark UI 的 ScrollArea，
 * 內部結構為 `Root > Viewport > Content`，加上依 `orientation` 掛載的 `Scrollbar/Thumb` 與 `Corner`。
 *
 * 對外維持單一入口（`class` + `children`）。高度一律由呼叫端以 `max-h-*`/`h-*` 下在 Root 上：
 * Root 用 `flex flex-col` 把高度交給 `Viewport`（Ark 為它設 `overflow: auto`）。
 * Viewport 不能改用 Ark 範例的 `h-full`：`height: 100%` 在 `max-h` 父層下會量到比可見區更高的高度
 * （Edge 實測 viewport 472px / 可見 378px），尾端內容會被 Root 的 `overflow-hidden` 裁掉而捲不到。
 *
 * 樣式一律走語意 token；捲軸的顯隱交給 Ark 的 `data-hover`/`data-scrolling`（指標在 Root 內就會標記），
 * 深色模式交由 token 翻轉，不寫顏色相關的深色變體。
 */
export interface ScrollAreaProps extends JSX.HTMLAttributes<HTMLDivElement> {
  /** 要顯示哪一（幾）向的捲軸，預設 `vertical` */
  orientation?: "vertical" | "horizontal" | "both";
  class?: string;
  children?: JSX.Element;
}

export const ScrollArea: Component<ScrollAreaProps> = (props) => {
  const [local, rest] = splitProps(props, ["orientation", "class", "children"]);
  const showsVertical = () => (local.orientation ?? "vertical") !== "horizontal";
  const showsHorizontal = () => (local.orientation ?? "vertical") !== "vertical";

  const scrollbarClass =
    "pointer-events-none select-none p-0.5 opacity-0 transition-opacity duration-300 data-[hover]:pointer-events-auto data-[scrolling]:pointer-events-auto data-[hover]:opacity-100 data-[scrolling]:opacity-100";

  return (
    <ArkScrollArea.Root
      class={cn("relative flex flex-col overflow-hidden", local.class)}
      {...rest}
    >
      <ArkScrollArea.Viewport class="min-h-0 w-full flex-1 rounded-[inherit] focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
        <ArkScrollArea.Content>{local.children}</ArkScrollArea.Content>
      </ArkScrollArea.Viewport>

      <Show when={showsVertical()}>
        <ArkScrollArea.Scrollbar
          orientation="vertical"
          class={cn("hidden w-2.5 data-[overflow-y]:flex", scrollbarClass)}
        >
          <ArkScrollArea.Thumb class="w-1.5 rounded-lg bg-border hover:bg-muted-foreground/50" />
        </ArkScrollArea.Scrollbar>
      </Show>

      <Show when={showsHorizontal()}>
        <ArkScrollArea.Scrollbar
          orientation="horizontal"
          class={cn("hidden h-2.5 data-[overflow-x]:flex", scrollbarClass)}
        >
          <ArkScrollArea.Thumb class="h-1.5 rounded-lg bg-border hover:bg-muted-foreground/50" />
        </ArkScrollArea.Scrollbar>
      </Show>

      <ArkScrollArea.Corner />
    </ArkScrollArea.Root>
  );
};
