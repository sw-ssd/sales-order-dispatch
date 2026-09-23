import { render, screen, waitFor } from "@solidjs/testing-library";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from "@tanstack/solid-router";
import type * as ConnectRpc from "@connectrpc/connect";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ListOrdersResponse } from "~/lib/proto/salesorder/v1/salesorder_pb";

/**
 * 契約（spec §8.2：今日待出貨、待處理訂單數量、快速連結）：
 * - 三個數字各自來自 `ListOrders.total`（`pageSize: 1`，不撈清單自己數）。
 * - **查詢失敗顯示 `—` 而非 `0`** —— 兩者在營運上是相反結論（0＝真的沒有；`—`＝不知道）。
 * - 沒有 `sales_order:read` 時**不打這三支查詢**（打了必定 403，整頁只剩錯誤）。
 * - 快速連結依權限出現，不給通往 403 的入口。
 */

const { listMock } = vi.hoisted(() => ({ listMock: vi.fn() }));

vi.mock("~/lib/transport", () => ({ transport: {} }));
vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({ listOrders: listMock }),
}));

import DashboardPage from "./DashboardPage";
import { AbilityProvider } from "~/lib/ability/context";
import { createPermissions, type Permission } from "~/lib/ability/permissions";
import { createSignal } from "solid-js";

/** 一次回應：只需要 `total`（列不取）。 */
function page(total: number): ListOrdersResponse {
  return { orders: [], total } as unknown as ListOrdersResponse;
}

/**
 * 依 `status` 分派回應：首頁四支查詢都是同一支 `ListOrders`，用狀態 + 日期區分
 * （今日那兩支帶今天日期，不限日期那兩支日期為空）。
 */
function respond(byStatus: Record<string, number>, todayByStatus?: Record<string, number>) {
  listMock.mockImplementation((req: { status: string; expectedDeliveryDate: string }) => {
    if (req.expectedDeliveryDate !== "") {
      const total = todayByStatus?.[req.status];
      if (total === undefined) return Promise.resolve(page(0));
      return Promise.resolve(page(total));
    }
    const total = byStatus[req.status];
    if (total === undefined) return Promise.reject(new Error("unexpected status " + req.status));
    return Promise.resolve(page(total));
  });
}

const ALL_PERMS: Permission[] = [
  { resource: "sales_order", action: "read" },
  { resource: "dispatch", action: "read" },
  { resource: "print", action: "read" },
  { resource: "customer", action: "read" },
  { resource: "product", action: "read" },
  { resource: "notification", action: "read" },
];

function renderDashboard(perms: Permission[] = ALL_PERMS) {
  const [ability] = createSignal(createPermissions(perms));
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });

  // 快速連結用 `<Link>`，需要真的 router 才有 context。
  const rootRoute = createRootRoute({});
  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/",
    component: () => <DashboardPage />,
  });
  const router = createRouter({
    routeTree: rootRoute.addChildren([indexRoute]),
    history: createMemoryHistory({ initialEntries: ["/"] }),
  });

  render(() => (
    <QueryClientProvider client={client}>
      <AbilityProvider ability={ability}>
        <RouterProvider router={router} />
      </AbilityProvider>
    </QueryClientProvider>
  ));

  return { client, router };
}

describe("DashboardPage", () => {
  beforeEach(() => {
    listMock.mockReset();
  });

  it("顯示今日待出貨（pending＋processing 相加）、待處理、處理中三個數字", async () => {
    respond({ pending: 7, processing: 4 }, { pending: 2, processing: 1 });

    renderDashboard();

    // 今日待出貨 = 今日 pending 2 + 今日 processing 1 = 3；三個數字刻意互不相同，
    // 否則斷言「找得到 3」會分不出是哪張卡（今日合計與 processing 各自都可能剛好是 3）。
    await waitFor(() => expect(screen.getByText("3")).toBeTruthy());
    const values = Array.from(document.querySelectorAll("p.tabular-nums")).map(
      (el) => el.textContent
    );
    expect(values).toEqual(["3", "7", "4"]);
    expect(screen.getByText("今日待出貨")).toBeTruthy();
    expect(screen.getByText("待處理訂單")).toBeTruthy();
    expect(screen.getByText("處理中訂單")).toBeTruthy();
  });

  it("今日待出貨的兩支查詢只算今天的單（不含其他日期的 pending）", async () => {
    // 不限日期 pending = 7，但今天只有 1 → 卡片必須顯示 1，不是 7 也不是 8。
    respond({ pending: 7, processing: 3 }, { pending: 1, processing: 0 });

    renderDashboard();

    await waitFor(() => expect(screen.getByText("1")).toBeTruthy());
    const values = Array.from(document.querySelectorAll("p.tabular-nums")).map(
      (el) => el.textContent
    );
    expect(values).toEqual(["1", "7", "3"]);
  });

  it("查詢失敗時顯示 — 而不是 0（0 與「查不到」是相反結論）", async () => {
    listMock.mockRejectedValue(new Error("unavailable"));

    renderDashboard();

    await waitFor(() => expect(screen.getByText(/部分概況暫時無法取得/)).toBeTruthy());
    // 只斷言卡片上的數字（提示訊息本身也含一個「—」，故不能數整體出現次數）。
    const values = Array.from(document.querySelectorAll("p.tabular-nums")).map(
      (el) => el.textContent
    );
    expect(values).toEqual(["—", "—", "—"]);
    expect(screen.queryByText("0")).toBeNull();
  });

  it("沒有 sales_order:read 時不打查詢，也不顯示數字卡", async () => {
    respond({ pending: 7, processing: 3 }, { pending: 2, processing: 1 });

    renderDashboard([{ resource: "notification", action: "read" }]);

    // 給查詢一點時間；若有送出就會被 mock 記錄。
    await new Promise((r) => setTimeout(r, 50));
    expect(listMock).not.toHaveBeenCalled();
    expect(screen.queryByText("今日待出貨")).toBeNull();
  });

  it("快速連結只顯示有權限的入口", async () => {
    respond({ pending: 0, processing: 0 }, { pending: 0, processing: 0 });

    renderDashboard([
      { resource: "sales_order", action: "read" },
      { resource: "print", action: "read" },
    ]);

    await waitFor(() => expect(screen.getByText("快速連結")).toBeTruthy());
    expect(screen.getByText("訂單管理")).toBeTruthy();
    expect(screen.getByText("單據列印")).toBeTruthy();
    // 沒有 dispatch / customer / product / notification 權限
    expect(screen.queryByText("派車規劃")).toBeNull();
    expect(screen.queryByText("客戶總表")).toBeNull();
    expect(screen.queryByText("商品總表")).toBeNull();
    expect(screen.queryByText("通知中心")).toBeNull();
  });
});
