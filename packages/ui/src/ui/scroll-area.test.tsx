import { render, screen, fireEvent, waitFor } from "@solidjs/testing-library";
import { describe, expect, it } from "vitest";
import { ScrollArea } from "./scroll-area";

/**
 * jsdom 沒有 layout：量測輸入一律由 `fakeLayout` 假造，再以真實事件驅動 Ark 的量測，
 * 斷言留給 Ark 依量測自行算出的結果（observer 由 src/test-setup.ts 補上，這裡不碰）。
 */
const fakeLayout = (el: Element, sizes: Record<string, number>) => {
  for (const [prop, value] of Object.entries(sizes)) {
    Object.defineProperty(el, prop, { value, configurable: true });
  }
};

const VIEWPORT = '[data-part="viewport"]';
const ROOT = '[data-part="root"]';

describe("ScrollArea", () => {
  it("內容超出高度時可捲動：捲軸上的滾輪會捲動 Ark 的 viewport", async () => {
    render(() => (
      <ScrollArea class="h-20">
        <div style={{ height: "400px" }}>內容</div>
      </ScrollArea>
    ));

    const viewport = document.querySelector<HTMLElement>(VIEWPORT)!;
    // 內容必須落在 Ark 的 viewport 內，捲動層才是這個元素（Ark 會為它設 overflow: auto）。
    expect(viewport.contains(screen.getByText("內容"))).toBe(true);
    fakeLayout(viewport, {
      scrollHeight: 400,
      clientHeight: 100,
      scrollWidth: 100,
      clientWidth: 100,
    });

    fireEvent.wheel(
      document.querySelector<HTMLElement>('[data-part="scrollbar"][data-orientation="vertical"]')!,
      { deltaY: 100 }
    );

    await waitFor(() => expect(viewport.scrollTop).toBe(100));
  });

  it("量測後只標示真正溢出的方向", async () => {
    render(() => (
      <ScrollArea class="h-20">
        <div style={{ height: "400px" }}>內容</div>
      </ScrollArea>
    ));

    const viewport = document.querySelector<HTMLElement>(VIEWPORT)!;
    // Ark 預設兩軸都溢出；假造「只有垂直溢出」後必須自行修正。
    fakeLayout(viewport, {
      scrollHeight: 400,
      clientHeight: 100,
      scrollWidth: 100,
      clientWidth: 100,
    });
    fireEvent.scroll(viewport);

    const root = document.querySelector<HTMLElement>(ROOT)!;
    await waitFor(() => {
      expect(root.hasAttribute("data-overflow-y")).toBe(true);
      expect(root.hasAttribute("data-overflow-x")).toBe(false);
    });
  });

  it("內容放得下時不標示溢出", async () => {
    render(() => (
      <ScrollArea class="h-20">
        <div>內容</div>
      </ScrollArea>
    ));

    const viewport = document.querySelector<HTMLElement>(VIEWPORT)!;
    fakeLayout(viewport, {
      scrollHeight: 100,
      clientHeight: 100,
      scrollWidth: 100,
      clientWidth: 100,
    });
    fireEvent.scroll(viewport);

    const root = document.querySelector<HTMLElement>(ROOT)!;
    await waitFor(() => expect(root.hasAttribute("data-overflow-y")).toBe(false));
  });

  it("orientation 決定掛載哪一（幾）向的捲軸", () => {
    render(() => (
      <ScrollArea orientation="horizontal">
        <div>內容</div>
      </ScrollArea>
    ));

    expect(
      document.querySelector('[data-part="scrollbar"][data-orientation="horizontal"]')
    ).not.toBeNull();
    expect(document.querySelector('[data-part="scrollbar"][data-orientation="vertical"]')).toBeNull();
  });
});
