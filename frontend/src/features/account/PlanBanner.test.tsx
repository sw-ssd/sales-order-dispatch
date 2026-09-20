import { render, screen, waitFor } from "@solidjs/testing-library";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import type { JSX } from "solid-js";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { GetTenantEntitlementsResponse, Usage } from "~/lib/proto/platform/v1/platform_pb";

/**
 * 契約（spec §2.4 規則 2）：租戶後台的訂閱資訊**唯讀且不吵** —— banner 只在用量達 80/90%
 * 或試用將到期時出現，其餘狀態（未達門檻／載入中／載入失敗）一律不渲染任何提示，
 * 也不阻擋任何操作。
 */

const { loadMock } = vi.hoisted(() => ({ loadMock: vi.fn() }));

vi.mock("./queries", () => ({
  loadTenantEntitlements: loadMock,
  tenantEntitlementsQueryOptions: () => ({ queryKey: ["tenant-entitlements"], queryFn: loadMock }),
}));

// banner 的連結是 TanStack 的 `Link`（需要 router context）；測試只關心它指向哪裡。
vi.mock("@tanstack/solid-router", () => ({
  Link: (props: { to: string; children: JSX.Element }) => <a href={props.to}>{props.children}</a>,
}));

import PlanBanner from "./PlanBanner";

const DAY_MS = 24 * 60 * 60 * 1000;

function seats(limitValue: number, used: number): Usage {
  return {
    featureCode: "limit.seats",
    enabled: true,
    limitSet: true,
    limitValue,
    used,
  } as unknown as Usage;
}

function entitlements(over: {
  status?: string;
  trialEndsAt?: string;
  usage?: Usage[];
} = {}): GetTenantEntitlementsResponse {
  return {
    planCode: "std",
    planName: "標準",
    status: "active",
    trialEndsAt: "",
    usage: [],
    ...over,
  } as unknown as GetTenantEntitlementsResponse;
}

function renderBanner() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(() => (
    <QueryClientProvider client={client}>
      <PlanBanner />
    </QueryClientProvider>
  ));
  return client;
}

describe("PlanBanner", () => {
  beforeEach(() => {
    loadMock.mockReset();
  });

  it("用量達 80% 時出現提示，並連到帳號／方案頁", async () => {
    loadMock.mockResolvedValue(entitlements({ usage: [seats(10, 8)] }));

    renderBanner();

    const banner = await screen.findByRole("status");
    expect(banner.textContent).toMatch(/接近上限/);
    expect(screen.getByRole("link", { name: "查看方案" }).getAttribute("href")).toBe("/account");
  });

  it("試用將到期時出現提示", async () => {
    loadMock.mockResolvedValue(
      entitlements({
        status: "trialing",
        trialEndsAt: new Date(Date.now() + 2 * DAY_MS).toISOString(),
      })
    );

    renderBanner();

    await waitFor(() =>
      expect(screen.getByRole("status").textContent).toMatch(/試用將於 2 天後到期/)
    );
  });

  it("未達門檻時不渲染任何提示（不吵）", async () => {
    loadMock.mockResolvedValue(
      entitlements({
        status: "trialing",
        trialEndsAt: new Date(Date.now() + 10 * DAY_MS).toISOString(),
        usage: [seats(10, 7)],
      })
    );

    const client = renderBanner();

    // 等查詢真的落地才斷言，否則「沒有 banner」可能只是「還沒載入」。
    await waitFor(() =>
      expect(client.getQueryState(["tenant-entitlements"])?.status).toBe("success")
    );
    expect(screen.queryByRole("status")).toBeNull();
  });

  it("載入失敗時不渲染提示（失敗不擋操作）", async () => {
    loadMock.mockRejectedValue(new Error("boom"));

    const client = renderBanner();

    await waitFor(() => expect(client.getQueryState(["tenant-entitlements"])?.status).toBe("error"));
    expect(screen.queryByRole("status")).toBeNull();
  });
});
