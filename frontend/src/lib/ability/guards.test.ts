import { describe, expect, it, vi, beforeEach } from "vitest";
import { QueryClient } from "@tanstack/solid-query";

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

describe("requireAbility", () => {
  beforeEach(() => mockGetAbility.mockReset());

  it("有權限時放行", async () => {
    mockGetAbility.mockResolvedValue({ rules: [{ action: "read", subject: "sales_order" }] });
    const guard = makeRequireAbility(new QueryClient())("read", "sales_order");
    await expect(guard()).resolves.toBeUndefined();
  });
  it("無權限時 redirect /403", async () => {
    mockGetAbility.mockResolvedValue({ rules: [] });
    const guard = makeRequireAbility(new QueryClient())("read", "sales_order");
    // @solidjs/router redirect 拋出物件的形狀隨版本不同(href/url/Response headers)。
    // 註:Response.url 在此環境為空字串,?? 鏈會短路,故取任一含 /403 的候選。
    try {
      await guard();
      expect.unreachable("應拋出 redirect");
    } catch (e) {
      // e 為 unknown;redirect 拋出物件的形狀跨版本漂移,取任一含 /403 的候選。
      const candidates: unknown[] = [];
      if (e && typeof e === "object") {
        const obj = e as Record<string, unknown>; // 庫拋出物件(Response 或 redirect 物件),於此收斂
        candidates.push(obj.href, obj.url);
        const headers = obj.headers;
        if (headers && typeof headers === "object" && "get" in headers) {
          const getter = headers as { get(name: string): unknown }; // in 收斂後僅餘 get 可呼叫
          candidates.push(getter.get("Location"));
        }
      }
      const href = candidates.find((v): v is string => typeof v === "string" && v.includes("/403"));
      expect(href).toBeTruthy();
    }
  });
});
