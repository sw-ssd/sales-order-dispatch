import { describe, expect, it } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { createSignal, type JSX } from "solid-js";
import { AbilityProvider } from "./context";
import { Can } from "./Can";
import { createPermissions } from "./permissions";

function renderWithPerms(
  rules: ReadonlyArray<{ resource: string; action: string }>,
  ui: () => JSX.Element,
) {
  const [perms] = createSignal(createPermissions(rules));
  return render(() => <AbilityProvider ability={perms}>{ui()}</AbilityProvider>);
}

describe("<Can>", () => {
  it("允許時顯示 children", () => {
    renderWithPerms([{ resource: "sales_order", action: "read" }], () => (
      <Can I="read" a="sales_order"><button>列表</button></Can>
    ));
    expect(screen.getByText("列表")).toBeTruthy();
  });
  it("拒絕時顯示 fallback", () => {
    renderWithPerms([], () => (
      <Can I="read" a="sales_order" fallback={<span>無權限</span>}><button>列表</button></Can>
    ));
    expect(screen.queryByText("列表")).toBeNull();
    expect(screen.getByText("無權限")).toBeTruthy();
  });
  it("未授予的 action 拒絕", () => {
    renderWithPerms([{ resource: "sales_order", action: "read" }], () => (
      <Can I="write" a="sales_order" fallback={<button disabled>新增</button>}><button>新增</button></Can>
    ));
    expect((screen.getByText("新增") as HTMLButtonElement).disabled).toBe(true);
  });
});
