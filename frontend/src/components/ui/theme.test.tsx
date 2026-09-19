import { fireEvent, render, screen, waitFor } from "@solidjs/testing-library";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  THEME_DARK_CLASS,
  THEME_DARK_QUERY,
  THEME_STORAGE_KEY,
  ThemeProvider,
  ThemeSwitcher,
} from "./theme";

/**
 * jsdom 不評估媒體查詢（`matches` 永遠是 false），這裡補上可控 stub（含 `change` 事件），
 * 讓測試能決定系統現在偏好淺色還是深色、並模擬使用者在系統設定裡改偏好。
 */
function stubSystemDark(initialMatches: boolean) {
  const listeners = new Set<(event: MediaQueryListEvent) => void>();
  const query = {
    matches: initialMatches,
    media: THEME_DARK_QUERY,
    addEventListener: (_type: "change", listener: (event: MediaQueryListEvent) => void) => {
      listeners.add(listener);
    },
    removeEventListener: (_type: "change", listener: (event: MediaQueryListEvent) => void) => {
      listeners.delete(listener);
    },
  };
  vi.stubGlobal("matchMedia", vi.fn(() => query));
  return {
    setMatches(matches: boolean) {
      query.matches = matches;
      listeners.forEach((listener) => listener({ matches } as MediaQueryListEvent));
    },
  };
}

/** 最小組合：容器 + 切換器（切換器只從容器拿狀態，兩者一起才是可用的機制）。 */
const App = () => (
  <ThemeProvider>
    <ThemeSwitcher />
  </ThemeProvider>
);

describe("Theme", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    // 深色 class 掛在全域的 documentElement 上（不是元件的 DOM），由測試自己收拾。
    document.documentElement.classList.remove(THEME_DARK_CLASS);
  });

  it("預設跟隨系統：系統偏好深色就套用 .dark，且不覆寫 localStorage", async () => {
    stubSystemDark(true);
    render(App);

    await waitFor(() =>
      expect(document.documentElement.classList.contains(THEME_DARK_CLASS)).toBe(true),
    );
    // 掛載時只讀取、不覆寫：使用者沒選過就沒有存檔。
    expect(localStorage.getItem(THEME_STORAGE_KEY)).toBeNull();
    expect(screen.getByRole("radio", { name: "跟隨系統" }).getAttribute("aria-checked")).toBe(
      "true",
    );
  });

  it("切到深色／淺色：套用 class 並寫入 localStorage；再點已選中的項目不會被取消選取", async () => {
    stubSystemDark(false);
    render(App);

    await waitFor(() =>
      expect(document.documentElement.classList.contains(THEME_DARK_CLASS)).toBe(false),
    );

    fireEvent.click(screen.getByRole("radio", { name: "深色" }));

    await waitFor(() =>
      expect(document.documentElement.classList.contains(THEME_DARK_CLASS)).toBe(true),
    );
    expect(localStorage.getItem(THEME_STORAGE_KEY)).toBe("dark");

    // `deselectable={false}`：再點一次不會變成「沒有主題」。
    fireEvent.click(screen.getByRole("radio", { name: "深色" }));
    await waitFor(() =>
      expect(screen.getByRole("radio", { name: "深色" }).getAttribute("aria-checked")).toBe(
        "true",
      ),
    );
    expect(localStorage.getItem(THEME_STORAGE_KEY)).toBe("dark");

    fireEvent.click(screen.getByRole("radio", { name: "淺色" }));

    await waitFor(() =>
      expect(document.documentElement.classList.contains(THEME_DARK_CLASS)).toBe(false),
    );
    expect(localStorage.getItem(THEME_STORAGE_KEY)).toBe("light");
  });

  it("system 時跟著系統偏好的變化即時切換（兩個方向）", async () => {
    const system = stubSystemDark(false);
    render(App);

    await waitFor(() =>
      expect(document.documentElement.classList.contains(THEME_DARK_CLASS)).toBe(false),
    );

    system.setMatches(true);
    await waitFor(() =>
      expect(document.documentElement.classList.contains(THEME_DARK_CLASS)).toBe(true),
    );

    system.setMatches(false);
    await waitFor(() =>
      expect(document.documentElement.classList.contains(THEME_DARK_CLASS)).toBe(false),
    );
    // 跟隨系統不會寫入偏好。
    expect(localStorage.getItem(THEME_STORAGE_KEY)).toBeNull();
  });

  it("重整後維持上次的偏好：存了 dark 就套用，不看當下的系統偏好", async () => {
    stubSystemDark(false);
    localStorage.setItem(THEME_STORAGE_KEY, "dark");
    render(App);

    await waitFor(() =>
      expect(document.documentElement.classList.contains(THEME_DARK_CLASS)).toBe(true),
    );
    expect(screen.getByRole("radio", { name: "深色" }).getAttribute("aria-checked")).toBe("true");
  });

  it("localStorage 的值不合法時回退 system", async () => {
    stubSystemDark(true);
    localStorage.setItem(THEME_STORAGE_KEY, "midnight");
    render(App);

    await waitFor(() =>
      expect(document.documentElement.classList.contains(THEME_DARK_CLASS)).toBe(true),
    );
    expect(screen.getByRole("radio", { name: "跟隨系統" }).getAttribute("aria-checked")).toBe(
      "true",
    );
  });
});
