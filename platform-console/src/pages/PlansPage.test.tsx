import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { connectErrorWithInfo } from "../test-helpers";

const listPlans = vi.fn();
const getPlanEntitlements = vi.fn();
const upsertPlanPrice = vi.fn();
const setPlanEntitlement = vi.fn();
vi.mock("../lib/api", () => ({
  platform: {
    listPlans: (...args: unknown[]) => listPlans(...args),
    getPlanEntitlements: (...args: unknown[]) => getPlanEntitlements(...args),
    upsertPlanPrice: (...args: unknown[]) => upsertPlanPrice(...args),
    setPlanEntitlement: (...args: unknown[]) => setPlanEntitlement(...args),
  },
}));

import PlansPage from "./PlansPage";

/**
 * 契約：價目是**價格史**（調價下一期生效，既有期別已快照）——畫面必須說出來，
 * 免得營運以為改了就影響當期；寫入的 `reason` 空字串在送出前就被擋；
 * 金額只做格式檢查（真偽由後端 money.ParseCents 決定）；成功後相關查詢自動失效。
 */
const plans = {
  plans: [
    {
      id: "1",
      code: "std",
      name: "標準",
      status: "active",
      sortOrder: 1,
      prices: [
        {
          billingCycle: "monthly",
          basePrice: "1000.00",
          seatPrice: "200.00",
          currency: "TWD",
          effectiveFrom: "2026-09-01T00:00:00Z",
        },
        {
          billingCycle: "yearly",
          basePrice: "10000.00",
          seatPrice: "2000.00",
          currency: "TWD",
          effectiveFrom: "2026-09-01T00:00:00Z",
        },
      ],
    },
    {
      id: "2",
      code: "free",
      name: "免費",
      status: "active",
      sortOrder: 2,
      prices: [],
    },
  ],
};

const entitlements = {
  entitlements: [
    { featureCode: "limit.seats", enabled: true, limitSet: true, limitValue: 10n },
    { featureCode: "feature.printing", enabled: false, limitSet: false, limitValue: 0n },
  ],
  features: [
    { code: "limit.seats", type: "integer", unit: "席", description: "席位" },
    { code: "feature.printing", type: "boolean", unit: "", description: "列印" },
  ],
};

function renderPage() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(() => (
    <QueryClientProvider client={client}>
      <PlansPage />
    </QueryClientProvider>
  ));
}

async function ready() {
  await screen.findByText("標準");
  return screen.getByRole("table", { name: "標準 價目" });
}

describe("PlansPage", () => {
  beforeEach(() => {
    for (const fn of [listPlans, getPlanEntitlements, upsertPlanPrice, setPlanEntitlement]) {
      fn.mockReset();
    }
    listPlans.mockResolvedValue(plans);
    getPlanEntitlements.mockResolvedValue(entitlements);
    upsertPlanPrice.mockResolvedValue({});
    setPlanEntitlement.mockResolvedValue({});
  });

  it("列出方案與現行價目；沒有價目的方案顯示空狀態", async () => {
    renderPage();
    const table = await ready();

    expect(within(table).getAllByRole("row")).toHaveLength(3); // 表頭 + 月繳 + 年繳
    expect(within(table).getByText("10000.00")).toBeTruthy();
    expect(screen.getByText(/尚未設定價目/)).toBeTruthy();
  });

  it("明說生效語意：金額下一期生效、配額與開關立即生效", async () => {
    renderPage();
    await ready();

    // 金額／期別：新增一列價格史，已開立的期別金額有快照 → 不回溯。
    expect(screen.getByText(/下一期生效/)).toBeTruthy();
    expect(screen.getByText(/不會被回溯改帳/)).toBeTruthy();
    // 權益（配額與功能開關）：後端 SetPlanEntitlement 會 invalidateAll → 立即改變判定。
    expect(screen.getByText(/立即生效/)).toBeTruthy();
  });

  it("載入中顯示載入狀態（不是白屏）", () => {
    const { promise } = Promise.withResolvers<never>();
    listPlans.mockReturnValue(promise);
    renderPage();

    expect(screen.getByText("載入中…")).toBeTruthy();
    expect(screen.queryByRole("table")).toBeNull();
  });

  it("空清單顯示空狀態而非空白表格", async () => {
    listPlans.mockResolvedValue({ plans: [] });
    renderPage();

    await waitFor(() => expect(screen.getByText(/尚未定義任何方案/)).toBeTruthy());
    expect(screen.queryAllByRole("row")).toHaveLength(0);
  });

  it("查詢失敗顯示 ErrorInfo 的訊息（SYS-4001 附可行動說明）", async () => {
    listPlans.mockRejectedValue(
      connectErrorWithInfo("SYS-4001", {
        message: "缺少權限",
        details: { required_role: "admin" },
      }),
    );
    renderPage();

    const alert = await screen.findByRole("alert");
    expect(alert.textContent).toContain("SYS-4001");
    expect(alert.textContent).toMatch(/operator 帳號/);
  });

  it("調價對話框關閉後重開：欄位不殘留（A 方案的金額不得帶到 B 方案）", async () => {
    renderPage();
    await ready();

    fireEvent.click(screen.getByRole("button", { name: "調價：標準" }));
    fireEvent.input(await screen.findByLabelText(/基本價/), { target: { value: "1500.00" } });
    fireEvent.input(screen.getByLabelText(/單席價/), { target: { value: "200.00" } });
    fireEvent.input(screen.getByLabelText(/原因/), { target: { value: "年度調價" } });
    fireEvent.click(screen.getByRole("button", { name: "取消" }));
    await waitFor(() => expect(screen.queryByRole("button", { name: "儲存價目" })).toBeNull());

    fireEvent.click(screen.getByRole("button", { name: "調價：免費" }));
    const base = await screen.findByLabelText(/基本價/);
    expect((base as HTMLInputElement).value).toBe("");
    expect((screen.getByLabelText(/單席價/) as HTMLInputElement).value).toBe("");
    expect((screen.getByLabelText(/原因/) as HTMLInputElement).value).toBe("");
  });

  it("調價：reason 空白與金額格式錯誤都先擋；成功後價目重查", async () => {
    renderPage();
    await ready();

    fireEvent.click(screen.getByRole("button", { name: "調價：標準" }));
    fireEvent.change(await screen.findByLabelText(/計費週期/), { target: { value: "monthly" } });
    fireEvent.input(screen.getByLabelText(/基本價/), { target: { value: "1,500" } });
    fireEvent.input(screen.getByLabelText(/單席價/), { target: { value: "100" } });
    fireEvent.input(screen.getByLabelText(/幣別/), { target: { value: "TWD" } });
    fireEvent.input(screen.getByLabelText(/原因/), { target: { value: "年度調價" } });

    fireEvent.click(screen.getByRole("button", { name: "儲存價目" }));
    await waitFor(() => expect(screen.getByText(/金額格式/)).toBeTruthy());
    expect(upsertPlanPrice).not.toHaveBeenCalled();

    // 格式對了但原因空白 → 仍然不送出。
    fireEvent.input(screen.getByLabelText(/基本價/), { target: { value: "1500.00" } });
    fireEvent.input(screen.getByLabelText(/原因/), { target: { value: "" } });
    fireEvent.click(screen.getByRole("button", { name: "儲存價目" }));
    await waitFor(() => expect(screen.getByText(/原因必填/)).toBeTruthy());
    expect(upsertPlanPrice).not.toHaveBeenCalled();

    fireEvent.input(screen.getByLabelText(/原因/), { target: { value: "年度調價" } });
    fireEvent.click(screen.getByRole("button", { name: "儲存價目" }));

    await waitFor(() => expect(upsertPlanPrice).toHaveBeenCalledTimes(1));
    expect(upsertPlanPrice).toHaveBeenCalledWith({
      planCode: "std",
      billingCycle: "monthly",
      basePrice: "1500.00",
      seatPrice: "100",
      currency: "TWD",
      reason: "年度調價",
    });
    // 寫入成功 → 價目查詢自動失效（重查一次）。
    await waitFor(() => expect(listPlans).toHaveBeenCalledTimes(2));
  });

  it("方案權益：顯示目前設定、reason 必填、負上限先擋；成功後關閉並失效權益查詢", async () => {
    renderPage();
    await ready();

    fireEvent.click(screen.getByRole("button", { name: "權益：標準" }));
    fireEvent.change(await screen.findByLabelText(/^功能/), { target: { value: "feature.printing" } });
    await waitFor(() => expect(screen.getByText(/目前設定：停用・不限/)).toBeTruthy());

    fireEvent.click(screen.getByRole("checkbox", { name: "啟用" }));
    fireEvent.click(screen.getByRole("checkbox", { name: "指定上限" }));
    fireEvent.input(await screen.findByLabelText(/上限值/), { target: { value: "-1" } });
    fireEvent.input(screen.getByLabelText(/原因/), { target: { value: "POC" } });
    fireEvent.click(screen.getByRole("button", { name: "儲存權益" }));

    await waitFor(() => expect(screen.getByText(/不得為負/)).toBeTruthy());
    expect(setPlanEntitlement).not.toHaveBeenCalled();

    fireEvent.input(screen.getByLabelText(/上限值/), { target: { value: "50" } });
    fireEvent.input(screen.getByLabelText(/原因/), { target: { value: "" } });
    fireEvent.click(screen.getByRole("button", { name: "儲存權益" }));
    await waitFor(() => expect(screen.getByText(/原因必填/)).toBeTruthy());
    expect(setPlanEntitlement).not.toHaveBeenCalled();

    fireEvent.input(screen.getByLabelText(/原因/), { target: { value: "POC" } });
    fireEvent.click(screen.getByRole("button", { name: "儲存權益" }));

    await waitFor(() => expect(setPlanEntitlement).toHaveBeenCalledTimes(1));
    expect(setPlanEntitlement).toHaveBeenCalledWith({
      planCode: "std",
      featureCode: "feature.printing",
      enabled: true,
      limitSet: true,
      limitValue: 50n,
      reason: "POC",
    });
    // 成功後對話框關閉；權益查詢已失效（`refresh` 會 invalidate ["entitlements"]），
    // 下次開啟或切到權益頁時取到的是新值。
    await waitFor(() => expect(screen.queryByRole("button", { name: "儲存權益" })).toBeNull());
  });

  it("權益寫入失敗顯示後端訊息（不覆寫成前端自編文案）", async () => {
    setPlanEntitlement.mockRejectedValue(
      connectErrorWithInfo("SYS-4001", {
        message: "缺少權限",
        details: { required_role: "admin" },
      }),
    );
    renderPage();
    await ready();

    fireEvent.click(screen.getByRole("button", { name: "權益：標準" }));
    fireEvent.change(await screen.findByLabelText(/^功能/), { target: { value: "feature.printing" } });
    fireEvent.click(screen.getByRole("checkbox", { name: "啟用" }));
    fireEvent.input(screen.getByLabelText(/原因/), { target: { value: "POC" } });
    fireEvent.click(screen.getByRole("button", { name: "儲存權益" }));

    const alerts = await screen.findAllByRole("alert");
    const text = alerts.map((a) => a.textContent).join(" ");
    expect(text).toContain("SYS-4001");
    expect(text).toContain("缺少權限");
    expect(text).toMatch(/operator 帳號/);
  });
});
