import { render, waitFor } from "@solidjs/testing-library";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import BrandLogo from "./BrandLogo";

// /me 走原生 fetch（lib/me.ts），以全域替身攔截；每測試自行設定回應。
const fetchMock = vi.fn();

/** 掛上元件並回傳 client（「沒有 img」的斷言要先確認查詢已 settled，見 meSettled）。 */
function mount() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(() => (
    <QueryClientProvider client={client}>
      <BrandLogo />
    </QueryClientProvider>
  ));
  return client;
}

/**
 * 等 /me 查詢真正結束（fetch 發出 → 回應消費完）：absence 斷言若在查詢還沒回來時執行，
 * 永遠看不到 img → 得到必綠的假測試；兩段 waitFor 封掉這個窗口。
 */
async function meSettled(client: QueryClient) {
  await waitFor(() => expect(fetchMock).toHaveBeenCalled());
  await waitFor(() => expect(client.isFetching()).toBe(0));
}

const LOGO_ME = {
  user_id: "9",
  role: "staff",
  company: { id: "7", name: "測試公司", logo_url: "/api/v1/files/l1/download" },
};

beforeEach(() => {
  fetchMock.mockReset();
  vi.stubGlobal("fetch", fetchMock);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("<BrandLogo>", () => {
  it("公司有 Logo：顯示公司 Logo（src=logo_url、alt=公司名），預設圖示消失", async () => {
    fetchMock.mockResolvedValue({ ok: true, status: 200, json: async () => LOGO_ME });
    mount();
    // waitFor 的回調必須「不成立就丟錯」——直接回 querySelector 會把 null 當成功、立刻結束。
    await waitFor(() => expect(document.querySelector("img")).toBeTruthy());
    const img = document.querySelector("img");
    expect(img!.getAttribute("src")).toBe("/api/v1/files/l1/download");
    expect(img!.getAttribute("alt")).toBe("測試公司");
  });

  it("公司無 Logo（logo_url 空）：settled 後維持預設圖示、不渲染 img", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ ...LOGO_ME, company: { ...LOGO_ME.company, logo_url: "" } }),
    });
    const client = mount();
    await meSettled(client);
    expect(document.querySelector("img")).toBeNull();
    expect(document.querySelector("svg")).toBeTruthy(); // Truck fallback 仍在
  });

  it("未登入（401 → null）：settled 後維持預設圖示", async () => {
    fetchMock.mockResolvedValue({ ok: false, status: 401, json: async () => ({}) });
    const client = mount();
    await meSettled(client);
    expect(document.querySelector("img")).toBeNull();
  });
});
