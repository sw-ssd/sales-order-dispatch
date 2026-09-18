import { describe, expect, it } from "vitest";
import { createPermissions, hasPermission } from "./service";

describe("ability service (OpenFGA proxy)", () => {
  it("createPermissions 建集合並可查詢", () => {
    const perms = createPermissions([
      { action: "read", resource: "sales_order" },
      { action: "write", resource: "sales_order" },
    ]);
    expect(hasPermission(perms, "sales_order", "read")).toBe(true);
    expect(hasPermission(perms, "sales_order", "write")).toBe(true);
    expect(hasPermission(perms, "customer", "read")).toBe(false);
  });
  it("空規則 fail-closed", () => {
    const perms = createPermissions([]);
    expect(hasPermission(perms, "sales_order", "read")).toBe(false);
  });
});
