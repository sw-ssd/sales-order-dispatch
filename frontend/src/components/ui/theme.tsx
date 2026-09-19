import { ToggleGroup } from "@ark-ui/solid";
import { Monitor, Moon, Sun, type LucideIcon } from "lucide-solid";
import {
  createContext,
  createEffect,
  createSignal,
  For,
  onCleanup,
  useContext,
  type Accessor,
  type Component,
  type ParentComponent,
} from "solid-js";
import { Dynamic } from "solid-js/web";
import { cn } from "@/lib/cn";

/**
 * 深色模式（Tailkit 的 class-based dark mode），**全 app 唯一的機制**：偏好三態 `light`／`dark`／
 * `system`，持久化在 localStorage `ui:theme`，套用結果是 `document.documentElement` 上的 `.dark`
 * （`index.css` 的 `.dark` token 與 `@custom-variant dark`）。
 *
 * 兩段接力，缺一不可：
 * 1. `index.html` 的 anti-FOUC 內聯腳本：在任何 CSS 與 bundle 之前同步套用同一個規則（不引用外部檔，
 *    因此那段邏輯與這裡是刻意重複的兩份；改規則時兩邊都要改）。
 * 2. 本檔：接手之後的切換（含 `system` 時跟著 `prefers-color-scheme` 即時變）。
 */

export type Theme = "light" | "dark" | "system";

/** 主題偏好的持久化 key（值為 `"light"`／`"dark"`／`"system"`）。 */
export const THEME_STORAGE_KEY = "ui:theme";

/** 系統深色偏好的媒體查詢。 */
export const THEME_DARK_QUERY = "(prefers-color-scheme: dark)";

/** 深色 class：掛在 `document.documentElement`（Tailkit 的 `.dark` 與其子孫都套用）。 */
export const THEME_DARK_CLASS = "dark";

export function isTheme(value: unknown): value is Theme {
  return value === "light" || value === "dark" || value === "system";
}

/** 讀取持久化的偏好；沒存過、值不合法、`window` 不存在（SSR）或 storage 被封鎖時回退 `system`。 */
function readStoredTheme(): Theme {
  try {
    const stored = window.localStorage.getItem(THEME_STORAGE_KEY);
    return isTheme(stored) ? stored : "system";
  } catch {
    return "system";
  }
}

/** 寫入偏好；無痕模式／配額用盡時只失去記憶，不影響本次切換。 */
function writeStoredTheme(theme: Theme): void {
  try {
    window.localStorage.setItem(THEME_STORAGE_KEY, theme);
  } catch {
    // 忽略：無法寫入時仍然照常切換，只是下次進來不會記得。
  }
}

/** 系統現在是否偏好深色；無 `window`／`matchMedia`（SSR、測試環境）時一律 false。 */
function readSystemDark(): boolean {
  if (typeof window === "undefined" || typeof window.matchMedia !== "function") {
    return false;
  }
  return window.matchMedia(THEME_DARK_QUERY).matches;
}

export interface ThemeContextValue {
  /** 使用者的偏好（`system` 代表跟隨系統，不是「現在的顏色」）。 */
  theme: Accessor<Theme>;
  setTheme: (theme: Theme) => void;
}

const ThemeContext = createContext<ThemeContextValue>();

/**
 * 深色模式容器。掛載時只「讀」localStorage，只有在偏好真的被切換過之後才寫回——使用者的既有選擇
 * 不會被預設值覆寫（與 `SidebarProvider` 同一套慣例）。
 */
export const ThemeProvider: ParentComponent = (props) => {
  const [theme, setThemeSignal] = createSignal(readStoredTheme());
  const [systemDark, setSystemDark] = createSignal(readSystemDark());

  // `system` 的變化只能靠媒體查詢事件得知（`matches` 不是 signal）。監聽整個 app 生命週期：
  // 成本是一個 listener，換來「使用者在系統設定裡切深色，這裡立刻跟上」。
  if (typeof window !== "undefined" && typeof window.matchMedia === "function") {
    const query = window.matchMedia(THEME_DARK_QUERY);
    const onChange = (event: MediaQueryListEvent) => setSystemDark(event.matches);
    query.addEventListener("change", onChange);
    onCleanup(() => query.removeEventListener("change", onChange));
  }

  const setTheme = (next: Theme) => {
    setThemeSignal(next);
    writeStoredTheme(next);
  };

  createEffect(() => {
    const dark = theme() === "dark" || (theme() === "system" && systemDark());
    document.documentElement.classList.toggle(THEME_DARK_CLASS, dark);
  });

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>{props.children}</ThemeContext.Provider>
  );
};

/** 取得主題狀態；不在 `ThemeProvider` 內時直接拋錯，避免靜默失效。 */
export function useTheme(): ThemeContextValue {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error("useTheme 必須在 <ThemeProvider> 內使用");
  }
  return context;
}

interface ThemeOption {
  value: Theme;
  /** 無障礙名稱（按鈕只有圖示，名稱由 `aria-label` 提供）。 */
  label: string;
  icon: LucideIcon;
}

/** 三態的顯示順序：淺色 → 跟隨系統 → 深色。 */
const THEME_OPTIONS: ThemeOption[] = [
  { value: "light", label: "淺色", icon: Sun },
  { value: "system", label: "跟隨系統", icon: Monitor },
  { value: "dark", label: "深色", icon: Moon },
];

/** 分段控制外觀（顏色一律語意 token，沒有任何顏色相關的 `dark:` 變體）。 */
const SWITCHER =
  "inline-flex items-center gap-0.5 rounded-lg border border-border bg-muted p-0.5";
const SWITCHER_ITEM =
  "inline-flex size-8 cursor-pointer items-center justify-center rounded-md text-muted-foreground outline-hidden transition-colors hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring data-[state=on]:bg-card data-[state=on]:text-foreground data-[state=on]:shadow-xs [&_svg]:size-4";

export interface ThemeSwitcherProps {
  class?: string;
}

/**
 * 主題切換器：Ark `toggle-group` 的單選分段控制（`role="radiogroup"` + 每項 `role="radio"`，
 * 方向鍵可切、`deselectable={false}` 保證永遠有一項選中）。
 */
export const ThemeSwitcher: Component<ThemeSwitcherProps> = (props) => {
  const { theme, setTheme } = useTheme();

  return (
    <ToggleGroup.Root
      aria-label="主題"
      value={[theme()]}
      deselectable={false}
      onValueChange={(details) => {
        const next = details.value[0];
        if (isTheme(next)) setTheme(next);
      }}
      class={cn(SWITCHER, props.class)}
    >
      <For each={THEME_OPTIONS}>
        {(option) => (
          <ToggleGroup.Item value={option.value} aria-label={option.label} class={SWITCHER_ITEM}>
            <Dynamic component={option.icon} />
          </ToggleGroup.Item>
        )}
      </For>
    </ToggleGroup.Root>
  );
};
