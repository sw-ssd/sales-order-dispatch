import { Menu } from "lucide-solid";
import { Button } from "~/components/ui/button";
import { pageTitleFor } from "./Sidebar";

export interface TopbarProps {
  /** 目前路徑（由 AppShell 的 useRouterState 取得）。 */
  pathname: string;
  onMenuClick: () => void;
}

/**
 * 頂欄（Tailkit Light Sidebar 版型的 page header）：行動寬度放收合鈕，右側顯示目前頁面標題。
 * 標題用 `p` 而非 `h1`——各頁自己帶著 Tailkit Page Headings 的 `h1`，避免同頁兩個 h1。
 */
export default function Topbar(props: TopbarProps) {
  return (
    <header class="sticky top-0 z-30 flex h-16 flex-none items-center gap-3 border-b border-border bg-card px-4 lg:px-6">
      <Button
        variant="outline"
        size="icon"
        class="lg:hidden"
        aria-label="開啟側邊欄"
        onClick={() => props.onMenuClick()}
      >
        <Menu />
      </Button>
      <p class="text-base font-semibold text-foreground">
        {pageTitleFor(props.pathname) ?? "多公司訂出貨系統"}
      </p>
    </header>
  );
}
