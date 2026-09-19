import { fireEvent, render, screen, waitFor } from "@solidjs/testing-library";
import { Users } from "lucide-solid";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { SIDEBAR_STORAGE_KEY, SidebarProvider } from "./context";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubItem,
  SidebarRail,
  SidebarTrigger,
} from "./parts";

/**
 * jsdom 沒有 `matchMedia`，而側邊欄的 `isMobile` 由 `matchMedia("(max-width: 767px)")` 驅動；
 * 這裡補上可控 stub（含 `change` 事件），讓測試能決定自己跑在行動還是桌面寬度。
 */
function stubMatchMedia(initialMatches: boolean) {
  const listeners = new Set<(event: MediaQueryListEvent) => void>();
  const query = {
    matches: initialMatches,
    media: "(max-width: 767px)",
    addEventListener: (_type: "change", listener: (event: MediaQueryListEvent) => void) => {
      listeners.add(listener);
    },
    removeEventListener: (_type: "change", listener: (event: MediaQueryListEvent) => void) => {
      listeners.delete(listener);
    },
  };
  vi.stubGlobal("matchMedia", vi.fn(() => query));
  return {
    setMatches(matches: boolean) {
      query.matches = matches;
      listeners.forEach((listener) => listener({ matches } as MediaQueryListEvent));
    },
  };
}

/** 側邊欄的桌面根層：`data-state|data-collapsible|data-variant|data-side` 的公開樣式契約。 */
const sidebarRoot = () => document.querySelector<HTMLElement>("[data-variant]")!;

const toggleButton = () => screen.getByRole("button", { name: "切換側邊欄" });

/** 一個最小但完整的側邊欄：收合鈕 + 側欄 + 一個導覽連結。 */
const SidebarApp = () => (
  <SidebarProvider>
    <SidebarTrigger />
    <Sidebar>
      <SidebarContent>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton as="a" href="/users/companies">
              <span>客戶總表</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarContent>
    </Sidebar>
  </SidebarProvider>
);

describe("Sidebar", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("收合鈕切換桌面狀態並寫入 localStorage，再點一次可展開", async () => {
    render(SidebarApp);

    expect(sidebarRoot().getAttribute("data-state")).toBe("expanded");
    expect(sidebarRoot().getAttribute("data-collapsible")).toBe("");
    // 掛載時只讀取、不覆寫：使用者沒動過就沒有存檔。
    expect(localStorage.getItem(SIDEBAR_STORAGE_KEY)).toBeNull();

    fireEvent.click(toggleButton());

    await waitFor(() => expect(sidebarRoot().getAttribute("data-state")).toBe("collapsed"));
    expect(sidebarRoot().getAttribute("data-collapsible")).toBe("icon");
    expect(localStorage.getItem(SIDEBAR_STORAGE_KEY)).toBe("false");

    fireEvent.click(toggleButton());

    await waitFor(() => expect(localStorage.getItem(SIDEBAR_STORAGE_KEY)).toBe("true"));
    expect(sidebarRoot().getAttribute("data-state")).toBe("expanded");
  });

  it("重新掛載（重整）後維持上次的收合狀態", async () => {
    const first = render(SidebarApp);
    fireEvent.click(toggleButton());
    await waitFor(() => expect(localStorage.getItem(SIDEBAR_STORAGE_KEY)).toBe("false"));
    first.unmount();

    render(SidebarApp);

    expect(sidebarRoot().getAttribute("data-state")).toBe("collapsed");
    expect(sidebarRoot().getAttribute("data-collapsible")).toBe("icon");
  });

  it("行動版抽屜關閉時導覽連結不在無障礙樹／tab 順序，Esc 可關閉", async () => {
    stubMatchMedia(true);
    render(SidebarApp);

    // 關閉：Ark 的 Presence 把 Positioner/Content 標成 hidden，連結因此不可 Tab。
    expect(screen.queryByRole("link", { name: "客戶總表" })).toBeNull();
    expect(screen.queryByRole("dialog")).toBeNull();

    fireEvent.click(toggleButton());

    const link = await screen.findByRole("link", { name: "客戶總表" });
    expect(link).toBeTruthy();
    // dialog 的無障礙名稱來自 Drawer.Title（沒有 Title 就會是未命名的 dialog）。
    expect(screen.getByRole("dialog", { name: "導覽選單" })).toBeTruthy();

    fireEvent.keyDown(document.body, { key: "Escape" });

    await waitFor(() => expect(screen.queryByRole("link", { name: "客戶總表" })).toBeNull());
  });

  it("收合成 icon rail 時標籤改用 Ark tooltip，展開時不掛 tooltip", async () => {
    localStorage.setItem(SIDEBAR_STORAGE_KEY, "false");
    const app = () => (
      <SidebarProvider>
        <Sidebar>
          <SidebarContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton as="a" href="/users/companies" tooltip="客戶總表">
                  <Users />
                  <span>客戶總表</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarContent>
        </Sidebar>
      </SidebarProvider>
    );

    const collapsed = render(app);

    // 標籤仍在無障礙名稱裡（sr-only），可見標籤改由 Ark tooltip 提供。
    expect(screen.getByRole("link", { name: "客戶總表" })).toBeTruthy();
    expect(screen.getByRole("tooltip", { hidden: true }).textContent).toBe("客戶總表");
    collapsed.unmount();

    localStorage.setItem(SIDEBAR_STORAGE_KEY, "true");
    render(app);

    expect(screen.queryByRole("tooltip", { hidden: true })).toBeNull();
  });

  it("Ctrl+B（Ark useHotkeys 的 mod+B）切換側邊欄", async () => {
    render(SidebarApp);

    // jsdom 的 platform 不是 macOS，Ark 因此把 `mod` 解析成 Ctrl。
    fireEvent.keyDown(document.body, { key: "b", code: "KeyB", ctrlKey: true });

    await waitFor(() => expect(localStorage.getItem(SIDEBAR_STORAGE_KEY)).toBe("false"));
    expect(sidebarRoot().getAttribute("data-state")).toBe("collapsed");
  });

  it("同一個 nav 內只有一個 aria-current=page，品牌列在 nav 之外", () => {
    render(() => (
      <SidebarProvider>
        <Sidebar>
          <SidebarHeader>
            {/* 品牌列：v2 是 `<Link to="/">`，連到目前路徑時會自己帶上 aria-current="page"。 */}
            <a href="/" aria-current="page">
              品牌
            </a>
          </SidebarHeader>
          <SidebarContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton as="a" href="/" isActive>
                  <span>首頁</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton as="a" href="/users/roles">
                  <span>角色權限</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarContent>
        </Sidebar>
      </SidebarProvider>
    ));

    const nav = screen.getByRole("navigation", { name: "主要導覽" });
    expect(nav.querySelectorAll('[aria-current="page"]')).toHaveLength(1);
    expect(nav.contains(screen.getByRole("link", { name: "品牌" }))).toBe(false);
    expect(nav.contains(screen.getByRole("link", { name: "首頁" }))).toBe(true);
    expect(screen.getByRole("link", { name: "角色權限" }).hasAttribute("aria-current")).toBe(false);
  });

  it("子選單由 Ark collapsible 收合，關閉時子連結不在 tab 順序", async () => {
    render(() => (
      <SidebarProvider>
        <Sidebar>
          <SidebarContent>
            <SidebarMenuSub label="人員管理" icon={<Users />}>
              <SidebarMenuSubItem>
                <SidebarMenuButton as="a" href="/users/roles">
                  <span>角色權限</span>
                </SidebarMenuButton>
              </SidebarMenuSubItem>
            </SidebarMenuSub>
          </SidebarContent>
        </Sidebar>
      </SidebarProvider>
    ));

    const trigger = screen.getByRole("button", { name: "人員管理" });
    expect(trigger.getAttribute("aria-expanded")).toBe("false");
    expect(screen.queryByRole("link", { name: "角色權限" })).toBeNull();

    fireEvent.click(trigger);

    expect(await screen.findByRole("link", { name: "角色權限" })).toBeTruthy();
  });

  it("所有部件可組裝，Rail 也切換側邊欄（但不進 tab 順序）", async () => {
    render(() => (
      <SidebarProvider>
        <div class="flex">
          <Sidebar>
            <SidebarHeader>品牌</SidebarHeader>
            <SidebarContent>
              <SidebarGroup>
                <SidebarGroupLabel>營運</SidebarGroupLabel>
                <SidebarGroupContent>
                  <SidebarMenu>
                    <SidebarMenuItem>
                      <SidebarMenuButton tooltip="客戶總表">
                        <Users />
                        <span>客戶總表</span>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
            </SidebarContent>
            <SidebarFooter>v1.0.0</SidebarFooter>
            <SidebarRail />
          </Sidebar>
          <SidebarInset>頁面內容</SidebarInset>
        </div>
      </SidebarProvider>
    ));

    expect(screen.getByText("頁面內容")).toBeTruthy();
    expect(screen.getByRole("navigation", { name: "主要導覽" })).not.toBeNull();
    expect(screen.getByRole("button", { name: "客戶總表" })).toBeTruthy();

    const rail = screen
      .getAllByRole("button", { name: "切換側邊欄" })
      .find((button) => button.getAttribute("tabindex") === "-1");
    expect(rail).toBeTruthy();

    fireEvent.click(rail!);

    await waitFor(() => expect(localStorage.getItem(SIDEBAR_STORAGE_KEY)).toBe("false"));
  });
});
