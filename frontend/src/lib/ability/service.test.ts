import { subject } from "@casl/ability";
import { describe, expect, it } from "vitest";
import { createAppAbility } from "./service";

describe("createAppAbility", () => {
  it("由 proto 規則建 ability,支援 conditions 與 inverted", () => {
    const ability = createAppAbility([
      { action: "read", subject: "sales_order", conditions: { status: "pending" } },
      { action: "cancel", subject: "sales_order", inverted: true },
    ] as any);
    expect(ability.can("read", "sales_order")).toBe(true);
    expect(ability.can("cancel", "sales_order")).toBe(false);
    expect(ability.can("read", "customer")).toBe(false);
  });
  it("conditions 參與 instance 判斷", () => {
    const ability = createAppAbility([
      { action: "cancel", subject: "sales_order", conditions: { status: "pending" } },
    ] as any);
    expect(ability.can("cancel", subject("sales_order", { status: "pending" }))).toBe(true);
    expect(ability.can("cancel", subject("sales_order", { status: "processing" }))).toBe(false);
  });
  it("空規則 fail-closed", () => {
    const ability = createAppAbility([]);
    expect(ability.can("read", "sales_order")).toBe(false);
  });
});
