import { describe, expect, it } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { createSignal } from "solid-js";
import type { JSX } from "@solidjs/web";
import { subject } from "@casl/ability";
import { AbilityProvider } from "./context";
import { createAppAbility } from "./service";
import { Can } from "./Can";

function renderWithAbility(
  rules: Parameters<typeof createAppAbility>[0],
  ui: () => JSX.Element,
) {
  const [ability] = createSignal(createAppAbility(rules));
  return render(() => <AbilityProvider ability={ability}>{ui()}</AbilityProvider>);
}

describe("<Can>", () => {
  it("允許時顯示 children", () => {
    renderWithAbility([{ action: "read", subject: "sales_order" }], () => (
      <Can I="read" a="sales_order"><button>列表</button></Can>
    ));
    expect(screen.getByText("列表")).toBeTruthy();
  });
  it("拒絕時顯示 fallback", () => {
    renderWithAbility([], () => (
      <Can I="read" a="sales_order" fallback={<span>無權限</span>}><button>列表</button></Can>
    ));
    expect(screen.queryByText("列表")).toBeNull();
    expect(screen.getByText("無權限")).toBeTruthy();
  });
  it("instance 條件判斷(按鈕依狀態灰化)", () => {
    const order = subject("sales_order", { status: "processing" });
    renderWithAbility([{ action: "cancel", subject: "sales_order", conditions: { status: "pending" } }], () => (
      <Can I="cancel" a={order} fallback={<button disabled>取消</button>}><button>取消</button></Can>
    ));
    expect((screen.getByText("取消") as HTMLButtonElement).disabled).toBe(true);
  });
});
