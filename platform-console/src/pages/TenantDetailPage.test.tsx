import { Code } from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { createMemoryHistory, RouterProvider } from "@tanstack/solid-router";
import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { connectErrorWithInfo } from "../test-helpers";

const listTenants = vi.fn(); // 路由守衛的探針
const getTenant = vi.fn();
const getPlanEntitlements = vi.fn();
const setTenantOverride = vi.fn();
const revokeTenantOverride = vi.fn();
vi.mock("../lib/api", () => ({
  loginUrl: "/platform/auth/google",
  platform: {
    listTenants: (...args: unknown[]) => listTenants(...args),
    getTenant: (...args: unknown[]) => getTenant(...args),
    getPlanEntitlements: (...args: unknown[]) => getPlanEntitlements(...args),
    setTenantOverride: (...args: unknown[]) => setTenantOverride(...args),
    revokeTenantOverride: (...args: unknown[]) => revokeTenantOverride(...args),
  },
}));

import { resetSession } from "../lib/session";
import { createAppRouter } from "../router";

/**
 * 契約：例外清單與投影的可讀狀態、`reason` 空字串在送出前就被擋（不白跑一趟）、
 * 負的上限先擋（後端守衛會回 SYS-1001）、寫入成功後相關查詢自動失效（不靠手動重整）、
 * 失敗時顯示 ErrorInfo 的訊息（SYS-4001 要可行動）。
 */
const tenant = {
  companyId: "1",
  companyName: "甲公司",
  planCode: "std",
  planName: "標準",
  subscriptionStatus: "active",
  seatCount: 8,
  currentPeriodEnd: "2026-10-31T00:00:00Z",
  overdue: false,
};

const activeOverride = {
  id: "11",
  featureCode: "limit.seats",
  enabledSet: false,
  enabled: false,
  limitSet: true,
  limitValue: 30n,
  reason: "簽約承諾",
  owner: "ops@example.com",
  expiresAt: "2027-01-01T00:00:00Z",
};

const expiredOverride = {
  id: "12",
  featureCode: "feature.printing",
  enabledSet: true,
  enabled: true,
  limitSet: false,
  limitValue: 0n,
  reason: "去年的 POC",
  owner: "ops@example.com",
  expiresAt: "2025-01-01T00:00:00Z",
};

const planEntitlements = {
  entitlements: [
    { featureCode: "limit.seats", enabled: true, limitSet: true, limitValue: 10n },
    { featureCode: "feature.printing", enabled: false, limitSet: false, limitValue: 0n },
  ],
  features: [
    { code: "limit.seats", type: "integer", unit: "席", description: "席位" },
    { code: "feature.printing", type: "boolean", unit: "", description: "列印" },
  ],
};

function renderAt() {
  const router = createAppRouter(createMemoryHistory({ initialEntries: ["/tenants/1"] }));
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  render(() => (
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  ));
  return router;
}

/** 等頁面真的出現（守衛探針 → 租戶卡片 → 投影表）。 */
async function ready() {
  await screen.findByText("甲公司");
  return screen.findByRole("table", { name: "權益現況（投影）" });
}

function rowOf(table: HTMLElement, feature: string): HTMLElement {
  const row = within(table).getByText(feature).closest("tr");
  if (!row) throw new Error(`找不到 ${feature} 的列`);
  return row;
}

describe("TenantDetailPage", () => {
  beforeEach(() => {
    resetSession();
    for (const fn of [
      listTenants,
      getTenant,
      getPlanEntitlements,
      setTenantOverride,
      revokeTenantOverride,
    ]) {
      fn.mockReset();
    }
    listTenants.mockResolvedValue({ tenants: [], pagination: { page: 1, pageSize: 1, total: 0 } });
    getTenant.mockResolvedValue({ tenant, overrides: [activeOverride, expiredOverride] });
    getPlanEntitlements.mockResolvedValue(planEntitlements);
    setTenantOverride.mockResolvedValue({ id: "13" });
    revokeTenantOverride.mockResolvedValue({ companyId: "1", featureCode: "limit.seats" });
  });

  it("列出訂閱概況與例外；投影為方案 ⊕ 例外，且已過期的例外不列入生效值", async () => {
    renderAt();
    const projection = await ready();

    const overrides = screen.getByRole("table", { name: "功能例外（override）" });
    expect(within(overrides).getAllByRole("row")).toHaveLength(3); // 表頭 + 2 筆
    expect(rowOf(overrides, "limit.seats").textContent).toContain("簽約承諾");
    expect(screen.getAllByText("已過期")).toHaveLength(1);

    // 生效值：方案的 10 被未逾期例外的 30 覆寫。
    const seats = rowOf(projection, "limit.seats");
    expect(seats.textContent).toContain("上限 10");
    expect(seats.textContent).toContain("上限 30");
    // 未設定權益的維度（方案停用）不得被過期例外改成啟用。
    const printing = rowOf(projection, "feature.printing");
    expect(printing.textContent).toContain("停用");
    expect(within(projection).getAllByRole("row")).toHaveLength(3);
  });

  it("同一功能有兩筆未逾期的例外：標示多筆、不挑一筆當答案", async () => {
    getTenant.mockResolvedValue({
      tenant,
      overrides: [
        activeOverride,
        { ...activeOverride, id: "13", limitValue: 45n, reason: "第二次承諾" },
      ],
    });
    renderAt();
    const projection = await ready();

    // 前端拿不到 created_at，重建不出後端的判定順序 → 只能說「有多筆、以後端為準」：
    // 生效欄不給答案、例外欄標出筆數，區塊上方再說明一次。
    await waitFor(() =>
      expect(within(projection).getAllByText(/多筆例外/)).toHaveLength(2),
    );
    expect(within(projection).getByText("多筆例外：以後端判定為準")).toBeTruthy();
    expect(screen.getByText(/最終以後端判定為準/)).toBeTruthy();
  });

  it("新增例外關閉後重開：欄位不殘留（上一筆的承諾不得帶到下一筆）", async () => {
    renderAt();
    await ready();

    fireEvent.click(screen.getByRole("button", { name: "新增例外" }));
    fireEvent.input(await screen.findByLabelText(/功能代碼/), { target: { value: "limit.seats" } });
    fireEvent.input(screen.getByLabelText(/負責人/), { target: { value: "ops@example.com" } });
    fireEvent.click(screen.getByRole("button", { name: "取消" }));
    await waitFor(() => expect(screen.queryByRole("button", { name: "建立例外" })).toBeNull());

    fireEvent.click(screen.getByRole("button", { name: "新增例外" }));
    const featureCode = await screen.findByLabelText(/功能代碼/);
    expect((featureCode as HTMLInputElement).value).toBe("");
    expect((screen.getByLabelText(/負責人/) as HTMLInputElement).value).toBe("");
  });

  it("沒有例外時顯示空狀態（不是空表格）", async () => {
    getTenant.mockResolvedValue({ tenant, overrides: [] });
    renderAt();
    await screen.findByText("甲公司");

    await waitFor(() => expect(screen.getByText(/沒有例外/)).toBeTruthy());
    expect(screen.queryByRole("table", { name: "功能例外（override）" })).toBeNull();
  });

  it("查詢失敗顯示 ErrorInfo 的訊息（SYS-4001 附可行動說明）", async () => {
    getTenant.mockRejectedValue(
      connectErrorWithInfo("SYS-4001", {
        message: "缺少權限",
        details: { required_role: "admin" },
      }),
    );
    renderAt();

    const alert = await screen.findByRole("alert");
    expect(alert.textContent).toContain("SYS-4001");
    expect(alert.textContent).toMatch(/operator 帳號/);
  });

  it("新增例外：reason 空字串在送出前就被擋；填了才送出並使查詢失效", async () => {
    renderAt();
    await ready();
    expect(getTenant).toHaveBeenCalledTimes(1);

    fireEvent.click(screen.getByRole("button", { name: "新增例外" }));
    fireEvent.input(await screen.findByLabelText(/功能代碼/), {
      target: { value: "limit.seats" },
    });
    fireEvent.click(screen.getByRole("checkbox", { name: "指定上限" }));
    // 勾了才出現上限輸入（維度是 *_set，不是 0 值）。
    fireEvent.input(await screen.findByLabelText(/上限值/), { target: { value: "30" } });
    fireEvent.input(screen.getByLabelText(/負責人/), { target: { value: "ops@example.com" } });

    fireEvent.click(screen.getByRole("button", { name: "建立例外" }));
    await waitFor(() => expect(screen.getByText(/原因必填/)).toBeTruthy());
    expect(setTenantOverride).not.toHaveBeenCalled();

    // 前後空白要 trim 後才送出：稽核文字不該帶空白（後端原樣入庫）。
    fireEvent.input(screen.getByLabelText(/原因/), { target: { value: "  簽約承諾  " } });
    fireEvent.click(screen.getByRole("button", { name: "建立例外" }));

    await waitFor(() => expect(setTenantOverride).toHaveBeenCalledTimes(1));
    expect(setTenantOverride).toHaveBeenCalledWith({
      companyId: "1",
      featureCode: "limit.seats",
      enabledSet: false,
      enabled: false,
      limitSet: true,
      limitValue: 30n,
      owner: "ops@example.com",
      expiresAt: "",
      reason: "簽約承諾",
    });
    // 寫入成功 → 相關查詢自動失效（畫面重查，不必手動重整）。
    await waitFor(() => expect(getTenant).toHaveBeenCalledTimes(2));
  });

  it("負的上限先擋（後端守衛會回 SYS-1001），不送出", async () => {
    renderAt();
    await ready();

    fireEvent.click(screen.getByRole("button", { name: "新增例外" }));
    fireEvent.input(await screen.findByLabelText(/功能代碼/), {
      target: { value: "limit.seats" },
    });
    fireEvent.click(screen.getByRole("checkbox", { name: "指定上限" }));
    fireEvent.input(await screen.findByLabelText(/上限值/), { target: { value: "-5" } });
    fireEvent.input(screen.getByLabelText(/負責人/), { target: { value: "ops@example.com" } });
    fireEvent.input(screen.getByLabelText(/原因/), { target: { value: "測試" } });

    fireEvent.click(screen.getByRole("button", { name: "建立例外" }));

    await waitFor(() => expect(screen.getByText(/不得為負/)).toBeTruthy());
    expect(setTenantOverride).not.toHaveBeenCalled();
  });

  it("撤銷例外：reason 必填，成功後查詢自動失效", async () => {
    renderAt();
    await ready();

    const overrides = screen.getByRole("table", { name: "功能例外（override）" });
    fireEvent.click(within(rowOf(overrides, "limit.seats")).getByRole("button", { name: "撤銷" }));
    await screen.findByRole("button", { name: "確認撤銷" });

    fireEvent.click(screen.getByRole("button", { name: "確認撤銷" }));
    await waitFor(() => expect(screen.getByText(/原因必填/)).toBeTruthy());
    expect(revokeTenantOverride).not.toHaveBeenCalled();

    fireEvent.input(screen.getByLabelText(/原因/), { target: { value: " 承諾到期 " } });
    fireEvent.click(screen.getByRole("button", { name: "確認撤銷" }));

    await waitFor(() =>
      expect(revokeTenantOverride).toHaveBeenCalledWith({
        overrideId: "11",
        reason: "承諾到期",
      }),
    );
    await waitFor(() => expect(getTenant).toHaveBeenCalledTimes(2));
  });

  it("寫入失敗顯示後端訊息（不覆寫成前端自編文案），且不重查", async () => {
    setTenantOverride.mockRejectedValue(
      connectErrorWithInfo("SYS-1001", {
        message: "參數驗證失敗",
        connectCode: Code.InvalidArgument,
      }),
    );
    renderAt();
    await ready();

    fireEvent.click(screen.getByRole("button", { name: "新增例外" }));
    fireEvent.input(await screen.findByLabelText(/功能代碼/), { target: { value: "limit.seats" } });
    fireEvent.click(screen.getByRole("checkbox", { name: "指定啟用" }));
    // Ark 的勾選狀態在下一個 microtask 才回報：等它真的亮出來再送出。
    await screen.findByRole("checkbox", { name: "啟用" });
    fireEvent.input(screen.getByLabelText(/負責人/), { target: { value: "ops@example.com" } });
    fireEvent.input(screen.getByLabelText(/原因/), { target: { value: "測試" } });
    fireEvent.submit(screen.getByRole("button", { name: "建立例外" }).closest("form")!);

    const alerts = await screen.findAllByRole("alert");
    const text = alerts.map((a) => a.textContent).join(" ");
    expect(text).toContain("SYS-1001");
    expect(text).toContain("參數驗證失敗");
    // 失敗不得讓查詢失效（沒有成功的寫入就沒有要重查的資料）。
    expect(getTenant).toHaveBeenCalledTimes(1);
  });
});
