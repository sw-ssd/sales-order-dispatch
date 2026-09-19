import { SidebarTrigger, ThemeSwitcher } from "~/components/ui";
import { pageTitleFor } from "./Sidebar";

export interface TopbarProps {
  /** 目前路徑（由 AppShell 的 useRouterState 取得）。 */
  pathname: string;
}

/**
 * 頂欄（Tailkit Light Sidebar 版型的 page header）：左側是側邊欄切換鈕（桌面切換收合、行動寬度
 * 開抽屜），中間是目前頁面標題，右側是主題切換器。
 *
 * 標題用 `p` 而非 `h1`——各頁自己帶著 `h1`，避免同頁兩個 h1。
 */
export default function Topbar(props: TopbarProps) {
  return (
    <header class="sticky top-0 z-30 flex h-16 flex-none items-center gap-3 border-b border-border bg-card px-4 lg:px-6">
      <SidebarTrigger />
      <p class="text-base font-semibold text-foreground">
        {pageTitleFor(props.pathname) ?? "多公司訂出貨系統"}
      </p>
      <ThemeSwitcher class="ms-auto" />
    </header>
  );
}
