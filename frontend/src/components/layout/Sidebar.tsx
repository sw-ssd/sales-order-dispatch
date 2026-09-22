import { Link } from "@tanstack/solid-router";
import type { LucideIcon } from "lucide-solid";
import {
  Bell,
  Building2,
  Milestone,
  ClipboardList,
  CreditCard,
  Home,
  Network,
  Package,
  Printer,
  Route,
  Scissors,
  ScrollText,
  ShieldCheck,
  Tags,
  Truck,
  Undo2,
  UserCog,
  Users,
  Warehouse,
} from "lucide-solid";
import { For, Show, type Component } from "solid-js";
import { Dynamic } from "solid-js/web";
import {
  Sidebar as SidebarRoot,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "~/components/ui";

/** 側邊欄可導向的路由：只列 router 實際註冊的路徑。 */
export type NavRoute =
  | "/"
  | "/customers"
  | "/orders"
  | "/returns"
  | "/notifications"
  | "/products"
  | "/masters/routes"
  | "/masters/warehouses"
  | "/masters/categories"
  | "/masters/processing-specs"
  | "/printing"
  | "/dispatch"
  | "/users/companies"
  | "/users/departments"
  | "/users/roles"
  | "/users/users"
  | "/audit"
  | "/account";

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
 * 其餘尚無路由的畫面是佔位，避免指向不存在的路徑。圖示取自既有依賴 `lucide-solid`。
 */
const NAV_SECTIONS: NavSection[] = [
  {
    items: [{ label: "首頁", icon: Home, to: "/" }],
  },
  {
    heading: "營運",
    items: [
      { label: "客戶總表", icon: Users, to: "/customers" },
      { label: "商品總表", icon: Package, to: "/products" },
      { label: "訂單管理", icon: ClipboardList, to: "/orders" },
      { label: "退貨管理", icon: Undo2, to: "/returns" },
      { label: "車次主檔", icon: Milestone, to: "/masters/routes" },
      { label: "倉別主檔", icon: Warehouse, to: "/masters/warehouses" },
      { label: "商品分類", icon: Tags, to: "/masters/categories" },
      { label: "分切規格", icon: Scissors, to: "/masters/processing-specs" },
      { label: "派車規劃", icon: Route, to: "/dispatch" },
      { label: "單據列印", icon: Printer, to: "/printing" },
    ],
  },
  {
    heading: "人員管理",
    items: [
      { label: "公司", icon: Building2, to: "/users/companies" },
      { label: "部門", icon: Network, to: "/users/departments" },
      { label: "角色權限", icon: ShieldCheck, to: "/users/roles" },
      { label: "使用者", icon: UserCog, to: "/users/users" },
      { label: "稽核日誌", icon: ScrollText, to: "/audit" },
    ],
  },
  {
    heading: "帳號",
    // 租戶後台的訂閱資訊收在這裡（spec §2.4 規則 2）：頁面上是唯讀的方案與用量卡片，
    // 需要提醒時由 shell 的 banner 提示。
    items: [
      { label: "通知中心", icon: Bell, to: "/notifications" },
      { label: "帳號／方案", icon: CreditCard, to: "/account" },
    ],
  },
];

/** 目前路徑對應的導覽標題；不在導覽表內（例如登入頁）回 undefined。 */
export function pageTitleFor(pathname: string): string | undefined {
  for (const section of NAV_SECTIONS) {
    const hit = section.items.find((item) => item.to === pathname);
    if (hit) return hit.label;
  }
  return undefined;
}

export interface SidebarProps {
  /** 目前路徑（由 AppShell 的 useRouterState 取得）。 */
  pathname: string;
}

/**
 * app 層的側邊欄：導覽資料（`NAV_SECTIONS`）+ 元件庫的 sidebar 部件。
 *
 * 兩個 v2 遺留的無障礙問題在這裡被結構本身修掉（元件層的解剖見 `~/components/ui/sidebar`）：
 * 品牌列放在 `SidebarHeader`（在 `<nav>` 之外），所以品牌即使是連到目前路徑的 `<Link to="/">`
 * 也不會在同一個 landmark 裡產生第二個 `aria-current="page"`；行動版抽屜由 Ark `drawer` 持有，
 * 關閉時整棵內容離開無障礙樹與 tab 順序。
 *
 * 收合成 icon rail（寬度只有 3rem）時，品牌文字要收進無障礙樹、整列改為置中；這件事用側欄根層的
 * `data-collapsible="icon"` + `group/sidebar` 表達（元件的公開樣式契約），而不是讀 context——
 * 這個元件是 `SidebarProvider`/`SidebarLayoutProvider` 的**外層**，讀不到它們的 context。
 */
const Sidebar: Component<SidebarProps> = (props) => {
  return (
    <SidebarRoot collapsible="icon" class="w-full">
      <SidebarHeader class="group-data-[collapsible=icon]/sidebar:items-center">
        <Link
          to="/"
          class="group inline-flex items-center gap-2 overflow-hidden text-lg font-bold tracking-wide text-foreground"
        >
          <Truck class="size-5 flex-none text-primary transition group-hover:scale-110" />
          <span class="truncate group-data-[collapsible=icon]/sidebar:sr-only">
            多公司訂出貨系統
          </span>
        </Link>
      </SidebarHeader>

      <SidebarContent>
        <For each={NAV_SECTIONS}>
          {(section) => (
            <SidebarGroup>
              <Show when={section.heading}>
                <SidebarGroupLabel>{section.heading}</SidebarGroupLabel>
              </Show>
              <SidebarGroupContent>
                <SidebarMenu>
                  <For each={section.items}>
                    {(item) => (
                      <SidebarMenuItem>
                        <Show
                          when={item.to}
                          fallback={
                            // 尚無路由的畫面：不可點的佔位（沿用 v2 的 `aria-disabled` 語意）。
                            <SidebarMenuButton
                              as="span"
                              aria-disabled="true"
                              tooltip={item.label}
                              class="cursor-not-allowed text-muted-foreground/70"
                            >
                              <Dynamic component={item.icon} />
                              <span>{item.label}</span>
                            </SidebarMenuButton>
                          }
                        >
                          {(to) => (
                            /**
                             * `as={Link}` 讓 `to` 退化成 `string`（拿不到 Link 的泛型），型別擋不下
                             * 不存在的路徑——路徑由 `NavRoute` 這個聯集守住（見 ui/sidebar/README.md）。
                             */
                            <SidebarMenuButton
                              as={Link}
                              to={to()}
                              isActive={to() === props.pathname}
                              tooltip={item.label}
                            >
                              <Dynamic component={item.icon} />
                              <span>{item.label}</span>
                            </SidebarMenuButton>
                          )}
                        </Show>
                      </SidebarMenuItem>
                    )}
                  </For>
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          )}
        </For>
      </SidebarContent>
    </SidebarRoot>
  );
};

export default Sidebar;
