import { Link } from "@tanstack/solid-router";
import type { LucideIcon } from "lucide-solid";
import {
  Building2,
  ClipboardList,
  Home,
  Network,
  Package,
  Printer,
  Route,
  ShieldCheck,
  Truck,
  Users,
  X,
} from "lucide-solid";
import { For, Show } from "solid-js";
import { Dynamic } from "solid-js/web";
import { Button } from "~/components/ui/button";
import { cn } from "~/lib/cn";

/** 側邊欄可導向的路由：只列 router 實際註冊的路徑。 */
export type NavRoute =
  | "/"
  | "/users/companies"
  | "/users/departments"
  | "/users/roles";

interface NavItem {
  label: string;
  icon: LucideIcon;
  /** 未實作的畫面沒有 `to`，渲染為不可點的佔位。 */
  to?: NavRoute;
}

interface NavSection {
  /** 分組標題（Tailkit 版型的 "Projects"/"Account" 位置）。 */
  heading?: string;
  items: NavItem[];
}

/**
 * 導覽資料來源：依 Pixso 稿的 Web 畫面順序
 * （docs/design/2026-09-18-pixso-版面美化-進度存檔.md §2.1：Dashboard → 客戶總表 → 商品總表 →
 * 訂單管理 → 派車規劃 → 單據列印 → 管理人員名單 → 角色權限）。只有已註冊的路由帶 `to`，
 * 其餘尚無路由的畫面是佔位，避免指向不存在的路徑。圖示取自既有依賴 `lucide-solid`（repo 內
 * `pagination.tsx` 已在用），不照搬 Tailkit 版型的 heroicons／icon font。
 */
const NAV_SECTIONS: NavSection[] = [
  {
    items: [{ label: "首頁", icon: Home, to: "/" }],
  },
  {
    heading: "營運",
    items: [
      { label: "客戶總表", icon: Users },
      { label: "商品總表", icon: Package },
      { label: "訂單管理", icon: ClipboardList },
      { label: "派車規劃", icon: Route },
      { label: "單據列印", icon: Printer },
    ],
  },
  {
    heading: "人員管理",
    items: [
      { label: "公司", icon: Building2, to: "/users/companies" },
      { label: "部門", icon: Network, to: "/users/departments" },
      { label: "角色權限", icon: ShieldCheck, to: "/users/roles" },
    ],
  },
];

/** Tailkit 導覽項目的共用結構（尺寸/間距照抄，顏色改語意 token）；`group` 供圖示的 `group-hover:text-primary` 使用。 */
const NAV_ITEM =
  "group flex items-center gap-2 rounded-lg border px-2.5 py-2 text-sm font-medium transition-colors";

/** 目前路徑對應的導覽標題；不在導覽表內（例如登入頁）回 undefined。 */
export function pageTitleFor(pathname: string): string | undefined {
  for (const section of NAV_SECTIONS) {
    const hit = section.items.find((item) => item.to === pathname);
    if (hit) return hit.label;
  }
  return undefined;
}

function NavPlaceholder(props: { item: NavItem }) {
  return (
    <span
      aria-disabled="true"
      class={cn(NAV_ITEM, "cursor-not-allowed border-transparent text-muted-foreground/60")}
    >
      <Dynamic component={props.item.icon} class="size-5 flex-none" />
      <span class="grow">{props.item.label}</span>
    </span>
  );
}

export interface SidebarProps {
  /** 目前路徑（由 AppShell 的 useRouterState 取得）。 */
  pathname: string;
  /** 行動寬度的收合狀態。 */
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export default function Sidebar(props: SidebarProps) {
  return (
    <nav
      id="page-sidebar"
      aria-label="主要導覽"
      class={cn(
        "fixed top-0 bottom-0 left-0 z-50 flex h-full w-full -translate-x-full flex-col border-r border-border bg-card transition-transform duration-500 ease-out lg:w-64 lg:translate-x-0",
        props.open && "translate-x-0",
      )}
    >
      <div class="flex h-16 w-full flex-none items-center justify-between px-4 lg:justify-center">
        <Link
          to="/"
          class="group inline-flex items-center gap-2 text-lg font-bold tracking-wide text-foreground"
        >
          <Truck class="size-5 flex-none text-primary transition group-hover:scale-110" />
          <span>多公司訂出貨系統</span>
        </Link>
        <div class="lg:hidden">
          <Button
            variant="outline"
            size="icon"
            aria-label="關閉側邊欄"
            onClick={() => props.onOpenChange(false)}
          >
            <X />
          </Button>
        </div>
      </div>

      <div class="overflow-y-auto">
        <div class="w-full p-4">
          <div class="space-y-1">
            <For each={NAV_SECTIONS}>
              {(section) => (
                <>
                  <Show when={section.heading}>
                    <div class="px-3 pt-5 pb-2 text-xs font-semibold tracking-wider text-muted-foreground uppercase">
                      {section.heading}
                    </div>
                  </Show>
                  <For each={section.items}>
                    {(item) => (
                      <Show when={item.to} fallback={<NavPlaceholder item={item} />}>
                        {(to) => (
                          <Link
                            to={to()}
                            class={cn(
                              NAV_ITEM,
                              item.to === props.pathname
                                ? "border-primary/20 bg-primary/10 text-foreground"
                                : "border-transparent text-foreground hover:bg-primary/10 active:border-primary/20",
                            )}
                          >
                            <Dynamic
                              component={item.icon}
                              class={cn(
                                "size-5 flex-none",
                                item.to === props.pathname
                                  ? "text-primary"
                                  : "text-muted-foreground group-hover:text-primary",
                              )}
                            />
                            <span class="grow">{item.label}</span>
                          </Link>
                        )}
                      </Show>
                    )}
                  </For>
                </>
              )}
            </For>
          </div>
        </div>
      </div>
    </nav>
  );
}
