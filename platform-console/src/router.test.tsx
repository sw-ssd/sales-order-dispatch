import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { createMemoryHistory, RouterProvider } from "@tanstack/solid-router";
import { render, screen, waitFor } from "@solidjs/testing-library";
import { Code, ConnectError } from "@connectrpc/connect";
import { beforeEach, describe, expect, it, vi } from "vitest";

const listTenants = vi.fn();
const getOperatorSelf = vi.fn();
vi.mock("./lib/api", () => ({
  platform: {
    listTenants: (...args: unknown[]) => listTenants(...args),
    getOperatorSelf: (...args: unknown[]) => getOperatorSelf(...args),
  },
  loginUrl: "/platform/auth/google",
}));

import { resetSession } from "./lib/session";
import { createAppRouter } from "./router";

function renderAt(path: string) {
  const router = createAppRouter(createMemoryHistory({ initialEntries: [path] }));
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(() => (
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  ));
  return router;
}

/**
 * 路由守衛的行為契約（C-34：這些斷言必須真的會紅）。
 * 把 src/lib/guard.ts 的 `if (!(await ensureSession())) throw redirect(...)` 拿掉，
 * 這兩個測試就會失敗（見 task-10-report.md 的 RED 紀錄）。
 */
describe("路由守衛", () => {
  beforeEach(() => {
    listTenants.mockReset();
    getOperatorSelf.mockReset();
    getOperatorSelf.mockResolvedValue({ operatorId: "42", email: "ops@example.com", role: "admin" });
    resetSession();
  });

  it("未登入：/tenants 被導向 /login，且不渲染任何租戶資料", async () => {
    getOperatorSelf.mockRejectedValue(new ConnectError("未登入", Code.Unauthenticated));
    const router = renderAt("/tenants");

    await waitFor(() => expect(screen.getByRole("link", { name: "以 Google 登入" })).toBeTruthy());
    expect(router.state.location.pathname).toBe("/login");
    expect(screen.queryByRole("table")).toBeNull();
  });

  it("已登入：停在 /tenants 並渲染租戶清單", async () => {
    listTenants.mockResolvedValue({
      tenants: [
        {
          companyId: "7",
          companyName: "甲公司",
          planCode: "std",
          planName: "標準",
          subscriptionStatus: "active",
          seatCount: 8,
          overdue: false,
        },
      ],
      pagination: { page: 1, pageSize: 50, total: 1 },
    });
    const router = renderAt("/tenants");

    await waitFor(() => expect(screen.getByText("甲公司")).toBeTruthy());
    expect(router.state.location.pathname).toBe("/tenants");
    expect(screen.getByRole("table", { name: "租戶清單" })).toBeTruthy();
  });

  it("登入頁不需要 operator：/login 直接渲染登入引導", async () => {
    const router = renderAt("/login");

    await waitFor(() => expect(screen.getByRole("link", { name: "以 Google 登入" })).toBeTruthy());
    expect(router.state.location.pathname).toBe("/login");
    // 登入頁不套 console 外框（否則會出現按了又被守衛彈回的「登出」鈕與整排導覽，
    // 且 <main> 會嵌套在 <main> 內）。
    expect(screen.queryByRole("navigation", { name: "主導覽" })).toBeNull();
    expect(screen.queryByRole("button", { name: "登出" })).toBeNull();
    expect(listTenants).not.toHaveBeenCalled();
    expect(getOperatorSelf).not.toHaveBeenCalled();
  });

  // 未結項 #25 後半：「/login」精確字串對未來的「/login/verify」子路由會誤套外框。
  // RED：目前以 !== 比對，子路由會被包進 console 外框（含需登入的主導覽）。
  it("未來的 /login 子路由同樣不套外框（前綴比對）", async () => {
    // 子路由尚未註冊、無法走整棵 router 證明 —— 直接驗外框判定函式（RED：舊寫法對子路由回 false）。
    const { isLoginTree } = await import("./router");
    expect(isLoginTree("/login")).toBe(true);
    expect(isLoginTree("/login/verify")).toBe(true);
    expect(isLoginTree("/loginox")).toBe(false);
    expect(isLoginTree("/tenants")).toBe(false);
  });
});
