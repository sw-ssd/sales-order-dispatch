import { render, screen, fireEvent, waitFor } from "@solidjs/testing-library";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { GetTenantEntitlementsResponse, Usage } from "~/lib/proto/platform/v1/platform_pb";

/**
 * 契約（spec §2.4 第②面「租戶後台」、§4.6）：
 * 卡片唯讀、且**不吵** —— 只有用量達 80/90% 或試用將到期才出現提示；
 * 載入中／載入失敗／無訂閱列三種狀態都必須可讀且不誤導（不得讓失敗看起來像「沒有權益」，
 * 也不得把「尚未開通計費」寫成「你的方案不含此功能」——前端分不出這兩者，判定在後端）。
 */

/** 權益投影的假 loader：卡片與 banner 共用 `queries.ts` 的查詢選項，測試不打網路。 */
const { loadMock } = vi.hoisted(() => ({ loadMock: vi.fn() }));

vi.mock("./queries", () => ({
  loadTenantEntitlements: loadMock,
  tenantEntitlementsQueryOptions: () => ({ queryKey: ["tenant-entitlements"], queryFn: loadMock }),
}));

import PlanCard from "./PlanCard";

const DAY_MS = 24 * 60 * 60 * 1000;

/** 一筆用量投影（測試用數字；正式路徑是 int64 → bigint）。 */
function usage(input: {
  featureCode: string;
  enabled?: boolean;
  limitSet?: boolean;
  limitValue?: number;
  used?: number;
}): Usage {
  return {
    enabled: true,
    limitSet: false,
    limitValue: 0,
    used: 0,
    ...input,
  } as unknown as Usage;
}

function entitlements(
  over: {
    planCode?: string;
    planName?: string;
    status?: string;
    trialEndsAt?: string;
    usage?: Usage[];
  } = {}
): GetTenantEntitlementsResponse {
  return {
    planCode: "std",
    planName: "標準",
    status: "active",
    trialEndsAt: "",
    usage: [],
    ...over,
  } as unknown as GetTenantEntitlementsResponse;
}

/** 席位用量 8/10 的訂閱（`limitValue` 為上限、`used` 為現用量）。 */
function seats(over: { limitValue?: number; used?: number } = {}) {
  return entitlements({
    usage: [usage({ featureCode: "limit.seats", limitSet: true, limitValue: 10, ...over })],
  });
}

function renderCard() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(() => (
    <QueryClientProvider client={client}>
      <PlanCard />
    </QueryClientProvider>
  ));
}

describe("PlanCard", () => {
  beforeEach(() => {
    loadMock.mockReset();
  });

  it("顯示方案名稱、狀態、免責語意；用量達 80% 時出現提示", async () => {
    loadMock.mockResolvedValue(seats({ used: 8 }));

    renderCard();

    await waitFor(() => expect(screen.getByText("8/10")).toBeTruthy());
    expect(screen.getByText("標準")).toBeTruthy();
    expect(screen.getByText("訂閱中")).toBeTruthy();
    // 「這不是授權判定」必須在卡片上說清楚（與既有「前端守衛不構成授權」一致）。
    expect(screen.getByText(/不構成授權/)).toBeTruthy();
    expect(screen.getByText(/8\/10（80%）.*接近上限/)).toBeTruthy();
  });

  it("用量達 90% 時提示升級為「即將額滿」", async () => {
    loadMock.mockResolvedValue(seats({ used: 9 }));

    renderCard();

    await waitFor(() => expect(screen.getByText(/9\/10（90%）.*即將額滿/)).toBeTruthy());
  });

  it("用量未達門檻時不提示（卡片不吵）", async () => {
    loadMock.mockResolvedValue(seats({ used: 7 }));

    renderCard();

    await waitFor(() => expect(screen.getByText("7/10")).toBeTruthy());
    expect(screen.queryByText(/接近上限/)).toBeNull();
  });

  it("布林功能顯示開通狀態、無數值上限的功能顯示不限", async () => {
    loadMock.mockResolvedValue(
      entitlements({
        usage: [
          usage({ featureCode: "feature.printing" }),
          usage({ featureCode: "limit.customers", used: 12 }),
        ],
      })
    );

    renderCard();

    await waitFor(() => expect(screen.getByText("單據列印")).toBeTruthy());
    expect(screen.getByText("已開通")).toBeTruthy();
    expect(screen.getByText("12（不限）")).toBeTruthy();
  });

  it("試用 2 天後到期時提示，並顯示到期日", async () => {
    const trialEndsAt = new Date(Date.now() + 2 * DAY_MS).toISOString();
    loadMock.mockResolvedValue(entitlements({ status: "trialing", trialEndsAt }));

    renderCard();

    await waitFor(() => expect(screen.getByText(/試用將於 2 天後到期/)).toBeTruthy());
    expect(screen.getByText(trialEndsAt.slice(0, 10))).toBeTruthy();
  });

  it("試用還有 10 天時不提示（卡片不吵）", async () => {
    const trialEndsAt = new Date(Date.now() + 10 * DAY_MS).toISOString();
    loadMock.mockResolvedValue(entitlements({ status: "trialing", trialEndsAt }));

    renderCard();

    await waitFor(() => expect(screen.getByText(trialEndsAt.slice(0, 10))).toBeTruthy());
    expect(screen.queryByText(/試用將於/)).toBeNull();
  });

  it("載入中顯示載入狀態，不誤稱未開通或未含", () => {
    const pending = Promise.withResolvers<GetTenantEntitlementsResponse>();
    loadMock.mockImplementation(() => pending.promise);

    renderCard();

    expect(screen.getByText(/方案資訊載入中/)).toBeTruthy();
    expect(screen.queryByText(/尚未開通計費/)).toBeNull();
    expect(screen.queryByText(/方案未含/)).toBeNull();
  });

  it("載入失敗時顯示可讀狀態、不顯示任何額度，且重試可恢復", async () => {
    loadMock
      .mockRejectedValueOnce(new Error("boom"))
      .mockResolvedValue(seats({ used: 8 }));

    renderCard();

    await waitFor(() => expect(screen.getByText(/方案資訊暫時無法取得/)).toBeTruthy());
    // 失敗不得看起來像「沒有權益」：沒有額度、沒有未含、也沒有未開通。
    expect(screen.queryByText(/\d+\/\d+/)).toBeNull();
    expect(screen.queryByText(/方案未含/)).toBeNull();
    expect(screen.queryByText(/尚未開通計費/)).toBeNull();

    fireEvent.click(screen.getByRole("button", { name: "重試" }));

    await waitFor(() => expect(screen.getByText("8/10")).toBeTruthy());
  });

  it("無訂閱列（尚未開通計費）：不宣稱方案未含，也不列功能開通狀態", async () => {
    loadMock.mockResolvedValue(
      entitlements({
        planCode: "",
        planName: "",
        status: "none",
        usage: [
          usage({ featureCode: "limit.seats", enabled: false }),
          usage({ featureCode: "feature.printing", enabled: false }),
        ],
      })
    );

    renderCard();

    await waitFor(() => expect(screen.getByText(/尚未開通計費/)).toBeTruthy());
    // 投影在無訂閱列時仍會列 `enabled=false`，但守衛其實不施加限制 →
    // 前端不得把這個狀態寫成「你的方案不含此功能」。
    expect(screen.queryByText(/方案未含/)).toBeNull();
    expect(screen.queryByText("席位")).toBeNull();
    expect(screen.getByText(/不構成授權/)).toBeTruthy();
  });
});
