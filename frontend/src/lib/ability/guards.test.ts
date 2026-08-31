import { describe, expect, it, vi, beforeEach } from "vitest";
import { QueryClient } from "@tanstack/solid-query";
import type { RoutePreloadFuncArgs } from "@solidjs/router";

const mockGetAbility = vi.fn();
vi.mock("./service", async (orig) => {
  const mod = await orig<typeof import("./service")>();
  return {
    ...mod,
    abilityQueryOptions: {
      ...mod.abilityQueryOptions,
      queryFn: async () => (await mockGetAbility()).rules,
    },
  };
});

import { makeRequireAbility } from "./guards";

// 守衛只讀 args.intent,測試以最小形狀補齊型別。
const args = (intent: RoutePreloadFuncArgs["intent"]) =>
  ({ intent }) as RoutePreloadFuncArgs;

describe("requireAbility", () => {
  beforeEach(() => mockGetAbility.mockReset());

  it("有權限時放行(不導向)", async () => {
    mockGetAbility.mockResolvedValue({ rules: [{ action: "read", subject: "sales_order" }] });
    const navigate = vi.fn();
    const guard = makeRequireAbility(new QueryClient(), () => navigate)("read", "sales_order");
    await guard(args("navigate"));
    expect(navigate).not.toHaveBeenCalled();
  });

  it("無權限時 navigate /403(replace)", async () => {
    mockGetAbility.mockResolvedValue({ rules: [] });
    const navigate = vi.fn();
    const guard = makeRequireAbility(new QueryClient(), () => navigate)("read", "sales_order");
    await guard(args("navigate"));
    expect(navigate).toHaveBeenCalledWith("/403", { replace: true });
  });

  it("hover 預載(intent=preload)不查詢也不導向", async () => {
    const navigate = vi.fn();
    const guard = makeRequireAbility(new QueryClient(), () => navigate)("read", "sales_order");
    await guard(args("preload"));
    expect(mockGetAbility).not.toHaveBeenCalled();
    expect(navigate).not.toHaveBeenCalled();
  });
});
