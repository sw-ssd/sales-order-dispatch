import { fireEvent, render, screen, waitFor } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 前台公告以 spy 取代 API（同 AnnouncementsPage.test 的手法）。
const { listActiveSpy } = vi.hoisted(() => ({ listActiveSpy: vi.fn() }));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({ listActiveAnnouncements: listActiveSpy }),
}));

import AnnouncementBoard from "./AnnouncementBoard";

function banner(overrides: Record<string, unknown> = {}) {
  return {
    id: "b-1",
    type: "banner",
    title: "年節休假公告",
    content: "初一至初三暫停配送",
    imageUrl: "",
    linkUrl: "",
    publishAt: "2026-09-20T02:00:00Z",
    unpublishAt: "",
    sortOrder: 1,
    isActive: true,
    deployWeb: true,
    deployApp: true,
    companyId: "",
    departmentId: "",
    createdAt: "2026-09-19T10:00:00Z",
    updatedAt: "2026-09-19T10:00:00Z",
    ...overrides,
  };
}

function mount(response: unknown) {
  listActiveSpy.mockResolvedValue(response);
  render(() => (
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <AnnouncementBoard />
    </QueryClientProvider>
  ));
}

const EMPTY = { banners: [], news: [], articles: [] };

beforeEach(() => {
  vi.clearAllMocks();
});

describe("AnnouncementBoard", () => {
  it("以 platform=web 查前台公告 —— 平台不是後端預設，必須明確指定", async () => {
    mount({ banners: [banner()], news: [], articles: [] });
    await waitFor(() => expect(screen.getByText("年節休假公告")).toBeTruthy());
    // 不傳 platform 後端會 invalid_argument；傳錯平台則會顯示不該投放的公告。
    expect(listActiveSpy).toHaveBeenCalledWith({ platform: "web" });
  });

  it("完全沒有公告時整塊不渲染（不佔首頁、不顯示噪音空態）", async () => {
    mount(EMPTY);
    await waitFor(() => expect(listActiveSpy).toHaveBeenCalled());
    expect(screen.queryByText("最新消息")).toBeNull();
    expect(screen.queryByText(/沒有公告/)).toBeNull();
  });

  it("news 與 article 都進「最新消息」列表；article 提供圖文全文", async () => {
    mount({
      banners: [],
      news: [banner({ id: "n-1", type: "news", title: "本週菜價調整" })],
      articles: [
        banner({
          id: "a-1",
          type: "article",
          title: "食材保存指南",
          content: "冷藏 0-7 度",
        }),
      ],
    });

    await waitFor(() => expect(screen.getByText("最新消息")).toBeTruthy());
    expect(screen.getByText("本週菜價調整")).toBeTruthy();
    expect(screen.getByText("食材保存指南")).toBeTruthy();
    // article 才有全文展開（news 直接顯示內容，不需展開）。
    expect(screen.getByText("閱讀全文")).toBeTruthy();
  });

  it("banner 有 link_url 時整張可點並開新分頁；無 link_url 時不是連結", async () => {
    mount({
      banners: [
        banner({ id: "b-1", title: "有連結", linkUrl: "https://example.com/promo" }),
        banner({ id: "b-2", title: "無連結", sortOrder: 2 }),
      ],
      news: [],
      articles: [],
    });

    await waitFor(() => expect(screen.getByText("有連結")).toBeTruthy());
    const link = screen.getByRole("link");
    expect(link.getAttribute("href")).toBe("https://example.com/promo");
    // rel=noopener：外部連結不得取得 window.opener。
    expect(link.getAttribute("rel")).toContain("noopener");
  });

  it("多張 banner 可切換且依 sort_order（切換按鈕只在多張時出現）", async () => {
    mount({
      banners: [
        banner({ id: "b-1", title: "第一則", sortOrder: 1 }),
        banner({ id: "b-2", title: "第二則", sortOrder: 2 }),
      ],
      news: [],
      articles: [],
    });

    await waitFor(() => expect(screen.getByText("第一則")).toBeTruthy());
    expect(screen.getByText("1 / 2")).toBeTruthy();
    // 順序由後端決定；前端照收到的次序顯示（不再自行排序）。
    fireEvent.click(screen.getByLabelText("下一則"));
    await waitFor(() => expect(screen.getByText("第二則")).toBeTruthy());
    // 環繞：最後一張再往後回到第一張。
    fireEvent.click(screen.getByLabelText("下一則"));
    await waitFor(() => expect(screen.getByText("第一則")).toBeTruthy());
  });

  it("單張 banner 不顯示切換控制（按了不動的按鈕比沒有按鈕更糟）", async () => {
    mount({ banners: [banner()], news: [], articles: [] });
    await waitFor(() => expect(screen.getByText("年節休假公告")).toBeTruthy());
    expect(screen.queryByLabelText("下一則")).toBeNull();
    expect(screen.queryByLabelText("上一則")).toBeNull();
  });

  it("查詢失敗不讓首頁出現錯誤區塊（公告是附加資訊）", async () => {
    listActiveSpy.mockRejectedValue(new Error("boom"));
    render(() => (
      <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
      >
        <AnnouncementBoard />
      </QueryClientProvider>
    ));
    await waitFor(() => expect(listActiveSpy).toHaveBeenCalled());
    expect(screen.queryByText(/載入失敗/)).toBeNull();
    expect(screen.queryByText("最新消息")).toBeNull();
  });
});
