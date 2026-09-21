import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { createMemoryHistory, RouterProvider } from "@tanstack/solid-router";
import { fireEvent, render, screen, waitFor } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { connectErrorWithInfo } from "../test-helpers";

const listTenants = vi.fn();
const getOperatorSelf = vi.fn();
vi.mock("../lib/api", () => ({
  loginUrl: "/platform/auth/google",
  platform: {
    listTenants: (...args: unknown[]) => listTenants(...args),
    getOperatorSelf: (...args: unknown[]) => getOperatorSelf(...args),
  },
}));

import { resetSession } from "../lib/session";
import { createAppRouter } from "../router";

/**
 * 契約：伺服器端分頁與篩選（不是把一堆租戶抓到前端再過濾）、逾期標記只出現在逾期者、
 * 三態（載入／空／錯誤）都有可讀狀態。斷言看的是**送出的請求參數**與 DOM 狀態，
 * 不是「元件有 render」。
 *
 * 頁面走真實路由渲染（`useParams`／`Link` 需要 router context）；路由守衛會先打一次探針
 * （`pageSize: 1`），故以參數區分「探針」與「頁面查詢」。
 */
const probe = { tenants: [], pagination: { page: 1, pageSize: 1, total: 0 } };

const rows = [
  {
    companyId: "1",
    companyName: "甲公司",
    planCode: "std",
    planName: "標準",
    subscriptionStatus: "active",
    seatCount: 8,
    currentPeriodEnd: "2026-10-31T00:00:00Z",
    overdue: false,
  },
  {
    companyId: "2",
    companyName: "乙公司",
    planCode: "pro",
    planName: "專業",
    subscriptionStatus: "past_due",
    seatCount: 12,
    currentPeriodEnd: "2026-09-15T00:00:00Z",
    overdue: true,
  },
  {
    companyId: "3",
    companyName: "丙公司",
    planCode: "free",
    planName: "免費",
    subscriptionStatus: "none",
    seatCount: 0,
    currentPeriodEnd: "",
    overdue: false,
  },
];

/** 頁面查詢（非探針）的回應。 */
function respond(tenants: typeof rows, total = tenants.length, page = 1) {
  listTenants.mockImplementation((req: { pageSize?: number }) =>
    req.pageSize === 1
      ? Promise.resolve(probe)
      : Promise.resolve({ tenants, pagination: { page, pageSize: 20, total } }),
  );
}

function renderPage() {
  const router = createAppRouter(createMemoryHistory({ initialEntries: ["/tenants"] }));
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(() => (
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  ));
  return router;
}

describe("TenantsPage", () => {
  beforeEach(() => {
    resetSession();
    listTenants.mockReset();
    getOperatorSelf.mockReset();
    getOperatorSelf.mockResolvedValue({ operatorId: "42", email: "ops@example.com", role: "admin" });
    respond(rows);
  });

  it("列出三家租戶，且僅逾期者顯示標記", async () => {
    renderPage();

    // 結構斷言：表頭 + 3 列。
    await waitFor(() => expect(screen.getAllByRole("row")).toHaveLength(4));
    expect(screen.getByText("乙公司")).toBeTruthy();
    expect(screen.getAllByText("逾期")).toHaveLength(1);
    expect(screen.getByText("乙公司").closest("tr")?.textContent).toContain("逾期");
  });

  it("載入中顯示載入狀態（不是白屏）", async () => {
    const { promise } = Promise.withResolvers<never>(); // 永不解析＝停在載入中
    listTenants.mockImplementation((req: { pageSize?: number }) =>
      req.pageSize === 1 ? Promise.resolve(probe) : promise,
    );
    renderPage();

    await screen.findByRole("button", { name: "查詢" });
    expect(screen.getByText("載入中…")).toBeTruthy();
    expect(screen.queryByRole("table")).toBeNull();
  });

  it("空清單顯示空狀態而非空白表格", async () => {
    respond([], 0);
    renderPage();

    await waitFor(() => expect(screen.getByText(/目前沒有租戶/)).toBeTruthy());
    expect(screen.queryAllByRole("row")).toHaveLength(0);
  });

  it("SYS-4001 顯示可行動訊息（含缺哪個角色與下一步）", async () => {
    listTenants.mockImplementation((req: { pageSize?: number }) =>
      req.pageSize === 1
        ? Promise.resolve(probe)
        : Promise.reject(
            connectErrorWithInfo("SYS-4001", {
              message: "缺少權限",
              details: { required_role: "admin" },
            }),
          ),
    );
    renderPage();

    const alert = await screen.findByRole("alert");
    expect(alert.textContent).toContain("SYS-4001");
    expect(alert.textContent).toContain("admin");
    expect(alert.textContent).toMatch(/operator 帳號/);
  });

  it("關鍵字與狀態篩選走伺服器端，且回到第 1 頁", async () => {
    respond(rows, 45, 1);
    renderPage();
    await screen.findByText("乙公司");

    fireEvent.input(screen.getByLabelText("公司關鍵字"), { target: { value: "甲" } });
    fireEvent.change(screen.getByLabelText("訂閱狀態"), { target: { value: "past_due" } });
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));

    await waitFor(() =>
      expect(listTenants).toHaveBeenLastCalledWith({
        page: 1,
        pageSize: 20,
        keyword: "甲",
        status: "past_due",
      }),
    );
  });

  it("翻頁把頁碼帶進請求（分頁由後端決定）", async () => {
    respond(rows, 45, 1);
    renderPage();
    await screen.findByText("乙公司");
    expect(screen.getByText(/第 1–20 筆,共 45 筆/)).toBeTruthy();

    respond(rows, 45, 2);
    fireEvent.click(screen.getByRole("button", { name: "下一頁" }));

    await waitFor(() =>
      expect(listTenants).toHaveBeenLastCalledWith({
        page: 2,
        pageSize: 20,
        keyword: "",
        status: "",
      }),
    );
  });

  it("總筆數未超過一頁時不顯示分頁控制", async () => {
    renderPage();
    await screen.findByText("乙公司");
    expect(screen.queryByText(/共 3 筆/)).toBeNull();
  });
});
