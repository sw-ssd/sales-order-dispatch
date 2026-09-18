import { describe, expect, it } from "vitest";
import { createPermissions, hasPermission } from "./permissions";

describe("createPermissions", () => {
  it("由規則建集合(resource:action key)", () => {
    const perms = createPermissions([
      { resource: "sales_order", action: "read" },
      { resource: "sales_order", action: "write" },
      { resource: "customer", action: "read" },
    ]);
    expect(perms.size).toBe(3);
    expect(hasPermission(perms, "sales_order", "read")).toBe(true);
    expect(hasPermission(perms, "sales_order", "write")).toBe(true);
    expect(hasPermission(perms, "customer", "read")).toBe(true);
  });
  it("無權限 fail-closed false", () => {
    const perms = createPermissions([]);
    expect(hasPermission(perms, "sales_order", "read")).toBe(false);
  });
  it("未授予的 resource:action 回 false", () => {
    const perms = createPermissions([{ resource: "sales_order", action: "read" }]);
    expect(hasPermission(perms, "sales_order", "write")).toBe(false);
    expect(hasPermission(perms, "customer", "read")).toBe(false);
  });
});
