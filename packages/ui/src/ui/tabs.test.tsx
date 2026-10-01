import { render, screen, fireEvent, waitFor } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";
import { createSignal } from "solid-js";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./tabs";

/** 等一個真實的巨任務週期，讓 Ark 的 microtask 狀態回報有機會跑完。 */
const settle = () => {
  const { promise, resolve } = Promise.withResolvers<void>();
  setTimeout(resolve, 0);
  return promise;
};

const Harness = (props: { onValueChange?: (v: string) => void; disableSecond?: boolean }) => {
  const [value, setValue] = createSignal("employee");
  return (
    <Tabs
      value={value()}
      onValueChange={(next) => {
        setValue(next);
        props.onValueChange?.(next);
      }}
    >
      <TabsList>
        <TabsTrigger value="employee">員工</TabsTrigger>
        <TabsTrigger value="store" disabled={props.disableSecond}>
          店家
        </TabsTrigger>
      </TabsList>
      <TabsContent value="employee">員工內容</TabsContent>
      <TabsContent value="store">店家內容</TabsContent>
    </Tabs>
  );
};

describe("Tabs", () => {
  it("受控 value 決定顯示的 panel", () => {
    render(() => <Harness />);
    expect(screen.getByText("員工內容")).toBeTruthy();
    expect(screen.queryByText("店家內容")).toBeNull();
  });

  it("點第二個 tab 帶回新值並切換內容", async () => {
    const onValueChange = vi.fn();
    render(() => <Harness onValueChange={onValueChange} />);
    fireEvent.click(screen.getByText("店家"));
    // Ark 的狀態機與卸載都要等下一個 microtask／畫面才回報，互動後的斷言一律用真實等待。
    await waitFor(() => {
      expect(screen.getByText("店家內容")).toBeTruthy();
      expect(screen.queryByText("員工內容")).toBeNull();
    });
    expect(onValueChange).toHaveBeenCalledWith("store");
  });

  it("disabled trigger 不切換", async () => {
    const onValueChange = vi.fn();
    render(() => <Harness disableSecond onValueChange={onValueChange} />);
    const store = screen.getByRole("tab", { name: "店家" });
    expect(store.hasAttribute("disabled")).toBe(true);
    fireEvent.click(store);
    // 讓 Ark 的 microtask 跑完，再斷言「確實沒有切換」。
    await settle();
    expect(onValueChange).not.toHaveBeenCalled();
    expect(screen.getByText("員工內容")).toBeTruthy();
    expect(screen.queryByText("店家內容")).toBeNull();
  });

  it("選中的 trigger 對輔助科技標示為已選", async () => {
    render(() => <Harness />);
    expect(screen.getByRole("tab", { selected: true }).textContent).toBe("員工");
    fireEvent.click(screen.getByText("店家"));
    await waitFor(() =>
      expect(screen.getByRole("tab", { selected: true }).textContent).toBe("店家")
    );
  });
});
