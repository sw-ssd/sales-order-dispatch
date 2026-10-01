import {
  createContext,
  createEffect,
  createSignal,
  onCleanup,
  useContext,
  type Accessor,
  type ParentComponent,
} from "solid-js";
import { useHotkeys } from "@ark-ui/solid/hotkeys";

/**
 * 側邊欄狀態來源（solid-ui 的 context 形狀：`state`／`open`／`isMobile`／`openMobile`／
 * `toggleSidebar`），行為層改用 Ark UI：`mod+B` 交給 `useHotkeys`，行動版抽屜交給 `drawer`。
 *
 * 與 solid-ui 的差異（本專案的既有事實）：
 * - 持久化用 localStorage `ui:sidebar`（solid-ui 用 cookie `sidebar:state`）。
 * - 讀寫一律包 try/catch：SSR（無 `window`）與無痕／配額用盡（storage 拋錯）都只是失去記憶，
 *   不影響當次操作。
 */

/** 行動版斷點（px）：視窗寬度小於此值時，側邊欄改用 off-canvas 抽屜。 */
export const SIDEBAR_MOBILE_BREAKPOINT = 768;

/** 桌面展開寬度。 */
export const SIDEBAR_WIDTH = "16rem";
/** 桌面收合成 icon rail 的寬度。 */
export const SIDEBAR_WIDTH_ICON = "3rem";
/** 行動版抽屜寬度。 */
export const SIDEBAR_WIDTH_MOBILE = "18rem";

/** 桌面收合狀態的持久化 key（值為 `"true"`／`"false"`）。 */
export const SIDEBAR_STORAGE_KEY = "ui:sidebar";

/**
 * 掛在側邊欄根層的 CSS 變數（`--sidebar-width*`）：寬度只由這組變數決定，
 * 收合／展開切換的是變數的消費者而非散落各處的固定寬度。
 */
export const SIDEBAR_WIDTH_VARS = {
  "--sidebar-width": SIDEBAR_WIDTH,
  "--sidebar-width-icon": SIDEBAR_WIDTH_ICON,
  "--sidebar-width-mobile": SIDEBAR_WIDTH_MOBILE,
} as const;

export interface SidebarContextValue {
  /** 桌面狀態；行動版一律回報 `expanded`（行動版看 `openMobile`）。 */
  state: Accessor<"expanded" | "collapsed">;
  /** 桌面是否展開。 */
  open: Accessor<boolean>;
  setOpen: (open: boolean) => void;
  /** 是否為行動寬度（`matchMedia("(max-width: 767px)")`）。 */
  isMobile: Accessor<boolean>;
  /** 行動版抽屜是否開啟。 */
  openMobile: Accessor<boolean>;
  setOpenMobile: (open: boolean) => void;
  /** 切換桌面展開／收合（`SidebarTrigger`、`SidebarRail` 與 `mod+B` 共用）。 */
  toggleSidebar: () => void;
}

const SidebarContext = createContext<SidebarContextValue>();

/**
 * 內部：由 `Sidebar` 依 `collapsible`／`isMobile` 決定「這棵子樹現在是不是收合的 icon rail」，
 * 供 `SidebarMenuButton` 之類的部件決定標籤要不要 `sr-only`、要不要掛 Ark tooltip。
 * 由 `Sidebar` 統一提供（而不是各部件自己從 `state` 推導），`collapsible="offcanvas"`／
 * `"none"` 與行動版抽屜才不會被誤判成收合。
 */
const SidebarLayoutContext = createContext<Accessor<boolean>>();

export const SidebarLayoutProvider = SidebarLayoutContext.Provider;

/** 目前的 icon rail 收合狀態；不在 `Sidebar` 內時一律 false。 */
export function useSidebarCollapsed(): Accessor<boolean> {
  return useContext(SidebarLayoutContext) ?? (() => false);
}

/** 讀取持久化的桌面狀態；沒存過、`window` 不存在（SSR）或 storage 被封鎖時回退「展開」。 */
function readStoredOpen(): boolean {
  try {
    return window.localStorage.getItem(SIDEBAR_STORAGE_KEY) !== "false";
  } catch {
    return true;
  }
}

/** 寫入桌面狀態；無痕模式／配額用盡時只失去記憶，不影響本次操作。 */
function writeStoredOpen(open: boolean): void {
  try {
    window.localStorage.setItem(SIDEBAR_STORAGE_KEY, String(open));
  } catch {
    // 忽略：無法寫入時仍然照常切換，只是下次進來不會記得。
  }
}

/** `matchMedia` 驅動的 `isMobile`；無 `window`（SSR）或無 `matchMedia`（jsdom）時回退 false。 */
function createIsMobile(): Accessor<boolean> {
  if (typeof window === "undefined" || typeof window.matchMedia !== "function") {
    return () => false;
  }
  const query = window.matchMedia(`(max-width: ${SIDEBAR_MOBILE_BREAKPOINT - 1}px)`);
  const [isMobile, setIsMobile] = createSignal(query.matches);
  const onChange = (event: MediaQueryListEvent) => setIsMobile(event.matches);
  query.addEventListener("change", onChange);
  onCleanup(() => query.removeEventListener("change", onChange));
  return isMobile;
}

/**
 * 側邊欄狀態容器。掛載時只「讀」localStorage；只有在狀態真的被改過之後才寫回，
 * 因此使用者的既有選擇不會被預設值覆寫。
 */
export const SidebarProvider: ParentComponent = (props) => {
  const [open, setOpen] = createSignal(readStoredOpen());
  const [openMobile, setOpenMobile] = createSignal(false);
  const isMobile = createIsMobile();
  const toggleSidebar = () => setOpen((current) => !current);

  let previous: boolean | undefined;
  createEffect(() => {
    const next = open();
    if (previous !== undefined && previous !== next) writeStoredOpen(next);
    previous = next;
  });

  // `mod+B`：macOS 是 ⌘B、其他平台是 Ctrl+B。一律交給 Ark 的 hotkeys（不再手寫 keydown
  // listener）；`preventDefault` 擋掉瀏覽器／可編輯區對這組鍵的預設綁定。
  useHotkeys({
    commands: [
      {
        hotkey: "mod+b",
        action: () => toggleSidebar(),
        label: "切換側邊欄",
        options: { preventDefault: true },
      },
    ],
  });

  return (
    <SidebarContext.Provider
      value={{
        state: () => (open() ? "expanded" : "collapsed"),
        open,
        setOpen,
        isMobile,
        openMobile,
        setOpenMobile,
        toggleSidebar,
      }}
    >
      {props.children}
    </SidebarContext.Provider>
  );
};

/** 取得側邊欄狀態；不在 `SidebarProvider` 內時直接拋錯，避免靜默失效。 */
export function useSidebar(): SidebarContextValue {
  const context = useContext(SidebarContext);
  if (!context) {
    throw new Error("useSidebar 必須在 <SidebarProvider> 內使用");
  }
  return context;
}
