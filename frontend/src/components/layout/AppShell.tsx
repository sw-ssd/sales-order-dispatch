import { Splitter, type SplitterPanelData, type SplitterResizeDetails } from "@ark-ui/solid";
import { useRouterState } from "@tanstack/solid-router";
import { createEffect, createSignal, Show, type ParentProps } from "solid-js";
import {
  SIDEBAR_WIDTH,
  SIDEBAR_WIDTH_ICON,
  SidebarInset,
  SidebarProvider,
  ThemeProvider,
  useSidebar,
} from "~/components/ui";
import Sidebar from "./Sidebar";
import Topbar from "./Topbar";

export type AppShellProps = ParentProps<{
  /** true = 只渲染頁面內容（登入頁、403 這類自己就是完整版面的頁面）。 */
  chromeless: boolean;
}>;

/**
 * 展開時的兩個面板。側欄標 `preserve-pixel-size`：視窗縮放時側欄維持像素寬（只被 12–30rem
 * 夾住），而不是跟著比例縮放——側欄在寬螢幕上不該變成三分之一步寬；內容面板吃掉剩下的寬度。
 */
const EXPANDED_PANELS: SplitterPanelData[] = [
  { id: "sidebar", minSize: "12rem", maxSize: "30rem", resizeBehavior: "preserve-pixel-size" },
  { id: "content", minSize: "20rem" },
];

/** 收合成 icon rail 時的面板：側欄固定 `minSize === maxSize === 3rem`，拖曳與方向鍵都動不了它。 */
const COLLAPSED_PANELS: SplitterPanelData[] = [
  {
    id: "sidebar",
    minSize: SIDEBAR_WIDTH_ICON,
    maxSize: SIDEBAR_WIDTH_ICON,
    resizeBehavior: "preserve-pixel-size",
  },
  { id: "content", minSize: "20rem" },
];

/** 側欄與內容之間的拖曳把手：1px 分隔線 + 兩側各 4px 的熱區（`role="separator"`，方向鍵可調）。 */
const RESIZE_TRIGGER =
  "relative w-2 shrink-0 cursor-col-resize bg-transparent outline-hidden after:absolute after:inset-y-0 after:left-1/2 after:w-px after:-translate-x-1/2 after:bg-border hover:after:bg-primary focus-visible:after:bg-primary data-[dragging]:after:bg-primary";

/** 內容區（頂欄 + 頁面）。shell 是唯一的內距所有者，頁面本身不再自帶內距。 */
function ShellInset(props: ParentProps<{ pathname: string }>) {
  return (
    <SidebarInset class="h-dvh overflow-y-auto">
      <Topbar pathname={props.pathname} />
      <div class="grow p-4 lg:p-6">{props.children}</div>
    </SidebarInset>
  );
}

/**
 * 側欄 + 內容的版面：桌面是 Ark `splitter` 的兩個面板（側欄與內容之間的把手可拖曳調寬、
 * 方向鍵可調），側欄 `collapsible="icon"` 收合成 3rem 圖示列——**不用 `offcanvas`**：收合時的
 * `inert` 會連側欄內的切換熱區一起打死。行動寬度讓側欄變成 Ark `drawer`（off-canvas），
 * 內容佔滿整個寬度。
 */
function ShellChrome(props: ParentProps) {
  const { isMobile, open, setOpenMobile } = useSidebar();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  /** 拖曳／方向鍵調出來的寬度（Ark 回報的是百分比；還沒調過就用預設 16rem）。 */
  const [draggedSize, setDraggedSize] = createSignal<SplitterResizeDetails["size"] | null>(null);

  // 換頁後收起行動版抽屜，避免它蓋住剛進去的頁面（沿用 v2 的行為）。
  createEffect(() => {
    pathname();
    setOpenMobile(false);
  });

  return (
    <div class="flex h-dvh w-full min-w-80 overflow-hidden bg-background text-foreground">
      <Show
        when={!isMobile()}
        fallback={
          <>
            <Sidebar pathname={pathname()} />
            <ShellInset pathname={pathname()}>{props.children}</ShellInset>
          </>
        }
      >
        <Splitter.Root
          orientation="horizontal"
          panels={open() ? EXPANDED_PANELS : COLLAPSED_PANELS}
          /**
           * `size` 是受控值：Ark 在受控模式下**不會**自己改內部尺寸，只把使用者的調整回報到
           * `onResize`——所以拖曳／方向鍵的結果必須寫回這裡才會生效（面板寬度因此就是側欄寬度：
           * 展開 16rem 或上次拖曳的寬度，收合成 icon rail 3rem）。
           */
          size={open() ? (draggedSize() ?? [SIDEBAR_WIDTH]) : [SIDEBAR_WIDTH_ICON]}
          onResize={(details) => {
            if (open()) setDraggedSize(details.size);
          }}
          class="flex h-dvh w-full"
        >
          <Splitter.Panel id="sidebar" class="transition-[flex-grow] duration-200 ease-linear">
            <Sidebar pathname={pathname()} />
          </Splitter.Panel>
          <Splitter.ResizeTrigger
            id="sidebar:content"
            aria-label="調整側邊欄寬度"
            /**
             * Ark 直接把 `orientation="horizontal"` 映射成 `aria-orientation="horizontal"`，但那指的是
             * **面板排列方向**；依 APG「Window Splitter」，畫面上這條上下向、左右拖曳的分隔線，
             * accessible name 該是 `vertical`（鍵盤也是 Left/Right）。Ark 的 ResizeTrigger 是
             * `mergeProps(getResizeTriggerProps(), restProps)`——使用者給的屬性優先，所以顯式覆寫即可，
             * 不影響 `data-orientation` 與內部的鍵盤運算。
             */
            aria-orientation="vertical"
            class={RESIZE_TRIGGER}
          />
          <Splitter.Panel id="content">
            <ShellInset pathname={pathname()}>{props.children}</ShellInset>
          </Splitter.Panel>
        </Splitter.Root>
      </Show>
    </div>
  );
}

/**
 * app shell。`chromeless` 的路徑（登入頁、403）直接渲染頁面本身——它們各自就是完整版面。
 *
 * `ThemeProvider` 不分 chromeless：深色 class 掛在 `document.documentElement`，登入頁也要跟著
 * 使用者的偏好（切換器只出現在 shell 的頂欄）。
 */
export default function AppShell(props: AppShellProps) {
  return (
    <ThemeProvider>
      <Show when={!props.chromeless} fallback={props.children}>
        <SidebarProvider>
          <ShellChrome>{props.children}</ShellChrome>
        </SidebarProvider>
      </Show>
    </ThemeProvider>
  );
}
