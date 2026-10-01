import { render, screen, fireEvent, waitFor } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";
import { Checkbox } from "./checkbox";

const isChecked = () => {
  const el = screen.getByRole("checkbox") as HTMLInputElement;
  return el.tagName === "INPUT" ? el.checked : el.getAttribute("aria-checked") === "true";
};

/** 等一個真實的巨任務週期，讓 Ark 的 microtask 狀態回報有機會跑完。 */
const settle = () => {
  const { promise, resolve } = Promise.withResolvers<void>();
  setTimeout(resolve, 0);
  return promise;
};

describe("Checkbox", () => {
  it("受控 checked=false 點擊回報 true", async () => {
    const onCheckedChange = vi.fn();
    render(() => <Checkbox checked={false} onCheckedChange={onCheckedChange} aria-label="客戶 檢視" />);
    fireEvent.click(screen.getByRole("checkbox"));
    // Ark 的狀態機在 microtask 後才回報，用真實等待而非同步斷言。
    await waitFor(() => expect(onCheckedChange).toHaveBeenCalledWith(true));
  });

  it("受控 checked=true 點擊回報 false", async () => {
    const onCheckedChange = vi.fn();
    render(() => <Checkbox checked onCheckedChange={onCheckedChange} aria-label="客戶 檢視" />);
    fireEvent.click(screen.getByRole("checkbox"));
    await waitFor(() => expect(onCheckedChange).toHaveBeenCalledWith(false));
  });

  it("非受控 defaultChecked 可切換", async () => {
    const onCheckedChange = vi.fn();
    render(() => <Checkbox defaultChecked onCheckedChange={onCheckedChange} aria-label="商品 檢視" />);
    expect(isChecked()).toBe(true);
    fireEvent.click(screen.getByRole("checkbox"));
    // 同時斷言回呼：只看 input.checked 會被瀏覽器原生的切換掩蓋成假綠。
    await waitFor(() => {
      expect(onCheckedChange).toHaveBeenCalledWith(false);
      expect(isChecked()).toBe(false);
    });
  });

  it("disabled 不觸發", async () => {
    const onCheckedChange = vi.fn();
    render(() => <Checkbox disabled onCheckedChange={onCheckedChange} aria-label="訂單 檢視" />);
    fireEvent.click(screen.getByRole("checkbox"));
    // 讓 Ark 的 microtask 跑完，再斷言「確實沒有切換」。
    await settle();
    expect(onCheckedChange).not.toHaveBeenCalled();
  });

  it("aria-label 成為可辨識名稱", () => {
    render(() => <Checkbox aria-label="客戶 檢視" />);
    expect(screen.getByRole("checkbox", { name: "客戶 檢視" })).toBeTruthy();
  });
});
