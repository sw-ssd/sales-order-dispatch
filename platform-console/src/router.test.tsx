import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { createMemoryHistory, RouterProvider } from "@tanstack/solid-router";
import { render, screen, waitFor } from "@solidjs/testing-library";
import { Code, ConnectError } from "@connectrpc/connect";
import { beforeEach, describe, expect, it, vi } from "vitest";

const listTenants = vi.fn();
vi.mock("./lib/api", () => ({
  platform: { listTenants: (...args: unknown[]) => listTenants(...args) },
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
    resetSession();
  });

  it("未登入：/tenants 被導向 /login，且不渲染任何租戶資料", async () => {
    listTenants.mockRejectedValue(new ConnectError("未登入", Code.Unauthenticated));
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
    expect(listTenants).not.toHaveBeenCalled();
  });
});
