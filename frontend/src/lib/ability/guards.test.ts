import { describe, expect, it, vi, beforeEach } from "vitest";
import { QueryClient } from "@tanstack/solid-query";
import { isRedirect } from "@tanstack/solid-router";

const mockGetAbility = vi.fn();
vi.mock("./service", async (orig) => {
  const mod = await orig<typeof import("./service")>();
  return {
    ...mod,
    abilityQueryOptions: {
      ...mod.abilityQueryOptions,
      queryFn: async () => createPermissions(mockGetAbility()),
    },
  };
});

import { createPermissions, hasPermission } from "./permissions";
import { makeRequireAbility } from "./guards";

describe("permissions", () => {
  it("hasPermission 依集合查詢", () => {
    const perms = createPermissions([{ resource: "sales_order", action: "read" }]);
    expect(hasPermission(perms, "sales_order", "read")).toBe(true);
    expect(hasPermission(perms, "sales_order", "write")).toBe(false);
  });
});

describe("requireAbility", () => {
  beforeEach(() => mockGetAbility.mockReset());

  it("有權限時放行(不拋 redirect)", async () => {
    mockGetAbility.mockReturnValue([{ resource: "sales_order", action: "read" }]);
    const guard = makeRequireAbility(new QueryClient())("read", "sales_order");
    await expect(guard({ preload: false })).resolves.toBeUndefined();
  });

  it("無權限時拋 redirect /403", async () => {
    mockGetAbility.mockReturnValue([]);
    const guard = makeRequireAbility(new QueryClient())("read", "sales_order");
    try {
      await guard({ preload: false });
      expect.unreachable("應拋出 redirect");
    } catch (e) {
      expect(isRedirect(e)).toBe(true);
      if (e && typeof e === "object" && "options" in e) {
        const options = e.options;
        if (options && typeof options === "object" && "to" in options) {
          expect(options.to).toBe("/403");
        } else {
          expect.unreachable("redirect options 缺少 to");
        }
      } else {
        expect.unreachable("redirect Response 缺少 options");
      }
    }
  });

  it("hover 預載(preload)不查詢也不導向", async () => {
    const guard = makeRequireAbility(new QueryClient())("read", "sales_order");
    await expect(guard({ preload: true })).resolves.toBeUndefined();
    expect(mockGetAbility).not.toHaveBeenCalled();
  });
});
