import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 公告 API 以 spy 取代：頁面在模組層建立 connect client,
// 以 createClient 的替身同時攔截（ConnectError/Code 保持真實,錯誤訊息對照才有效）。
const {
  listAnnouncementsSpy,
  createAnnouncementSpy,
  updateAnnouncementSpy,
  deleteAnnouncementSpy,
} = vi.hoisted(() => ({
  listAnnouncementsSpy: vi.fn(),
  createAnnouncementSpy: vi.fn(),
  updateAnnouncementSpy: vi.fn(),
  deleteAnnouncementSpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listAnnouncements: listAnnouncementsSpy,
    createAnnouncement: createAnnouncementSpy,
    updateAnnouncement: updateAnnouncementSpy,
    deleteAnnouncement: deleteAnnouncementSpy,
  }),
}));

import AnnouncementsPage from "./AnnouncementsPage";

const EXISTING_ANNOUNCEMENT = {
  id: "a-1",
  companyId: "c-1",
  departmentId: "",
  type: "news",
  title: "中秋節停送公告",
  content: "中秋連假暫停配送",
  imageUrl: "",
  linkUrl: "",
  publishAt: "2026-09-20T02:00:00Z",
  unpublishAt: "2026-09-21T02:00:00Z",
  sortOrder: 3,
  isActive: true,
  deployWeb: true,
  deployApp: true,
  createdAt: "2026-09-19T10:00:00Z",
  updatedAt: "2026-09-19T10:00:00Z",
};

/**
 * 掛上頁面：每次呼叫都給全新的 `QueryClient`（快取不跨測試殘留;retry 關閉,
 * 理由同 WarehousesPage —— 失敗查詢不得自動重試拖慢失敗斷言）。
 */
function mountPage(
  client: QueryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
) {
  render(() => (
    <QueryClientProvider client={client}>
      <AnnouncementsPage />
    </QueryClientProvider>
  ));
  return client;
}

/** 標準情境：以預設 mock（一列公告）掛頁面並等清單落地。 */
async function renderPage() {
  mountPage();
  await waitFor(() => expect(screen.getByText("中秋節停送公告")).toBeTruthy());
}

/** 開新增對話框並回傳它的 DOM（欄位查詢收斂到這個 dialog）。 */
async function openCreateDialog(): Promise<HTMLElement> {
  fireEvent.click(screen.getByRole("button", { name: "新增公告" }));
  return await waitFor(() => screen.getByRole("dialog"));
}

beforeEach(() => {
  vi.clearAllMocks();
  listAnnouncementsSpy.mockResolvedValue({
    announcements: [EXISTING_ANNOUNCEMENT],
    total: 1,
  });
});

describe("AnnouncementsPage", () => {
  it("清單載入：以 total 推總筆數,型別參數進 query key,不送 sort/desc（proto 沒有）", async () => {
    await renderPage();
    expect(listAnnouncementsSpy).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: 20, type: "", includeDeleted: false })
    );
    expect(screen.getByText(/投放 Web 中台與 App 首頁（共 1 筆）/)).toBeTruthy();
    const sent = listAnnouncementsSpy.mock.calls[0][0] as Record<string, unknown>;
    expect("sort" in sent).toBe(false);
    expect("desc" in sent).toBe(false);
  });

  it("清單列渲染標題、類型徽章、範圍與投放", async () => {
    await renderPage();
    expect(screen.getByText("中秋節停送公告")).toBeTruthy();
    const cells = screen.getAllByRole("cell").map((c) => c.textContent ?? "");
    expect(cells.join("|")).toContain("最新消息");
    expect(cells.join("|")).toContain("公司 #c-1");
    expect(cells.join("|")).toContain("Web");
    expect(cells.join("|")).toContain("App");
  });

  it("類型篩選換 query key（select 直接生效,回第 1 頁）", async () => {
    await renderPage();
    listAnnouncementsSpy.mockClear();
    fireEvent.change(screen.getByLabelText("類型"), { target: { value: "banner" } });
    await waitFor(() =>
      expect(listAnnouncementsSpy).toHaveBeenCalledWith(
        expect.objectContaining({ type: "banner", page: 1 })
      )
    );
  });

  it("標題空白時不打建立 API（欄位級驗證攔下）", async () => {
    await renderPage();
    createAnnouncementSpy.mockResolvedValue({ announcement: EXISTING_ANNOUNCEMENT });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("標題 *"), { target: { value: "   " } });
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(screen.getByText("請輸入標題")).toBeTruthy());
    expect(createAnnouncementSpy).not.toHaveBeenCalled();
  });

  it("建立成功送整份欄位:自動歸屬 → company/department 為空字串,時間轉 RFC3339", async () => {
    await renderPage();
    createAnnouncementSpy.mockResolvedValue({ announcement: EXISTING_ANNOUNCEMENT });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("標題 *"), { target: { value: "促銷開跑" } });
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(createAnnouncementSpy).toHaveBeenCalledOnce());
    const sent = createAnnouncementSpy.mock.calls[0][0] as Record<string, unknown>;
    expect(sent).toMatchObject({
      companyId: "",
      departmentId: "",
      type: "news",
      title: "促銷開跑",
      isActive: true,
      deployWeb: true,
      deployApp: true,
      sortOrder: 0,
      unpublishAt: "",
    });
    // 上架時間預設「現在」→ 送出 RFC3339（ISO 8601 Z 形）。
    expect(typeof sent.publishAt).toBe("string");
    expect(sent.publishAt as string).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });

  it("編輯帶出既有值、範圍唯讀,update 不帶範圍欄位（proto:範圍不可改）", async () => {
    await renderPage();
    updateAnnouncementSpy.mockResolvedValue({ announcement: EXISTING_ANNOUNCEMENT });
    fireEvent.click(screen.getByRole("button", { name: "編輯" }));
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    await waitFor(() =>
      expect((within(dialog).getByLabelText("標題 *") as HTMLInputElement).value).toBe(
        "中秋節停送公告"
      )
    );
    // 範圍不可改:建立用的範圍 select 不出現,改顯示唯讀範圍文字。
    expect(within(dialog).queryByLabelText("範圍")).toBeNull();
    expect(within(dialog).getByText("發佈範圍")).toBeTruthy();

    fireEvent.input(within(dialog).getByLabelText("標題 *"), {
      target: { value: "中秋節停送公告（改）" },
    });
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(updateAnnouncementSpy).toHaveBeenCalledOnce());
    const sent = updateAnnouncementSpy.mock.calls[0][0] as Record<string, unknown>;
    expect(sent).toMatchObject({ id: "a-1", title: "中秋節停送公告（改）" });
    expect("companyId" in sent).toBe(false);
    expect("departmentId" in sent).toBe(false);
  });

  it("刪除先確認：拒絕不打 API,接受才呼叫 deleteAnnouncement({id})", async () => {
    await renderPage();
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(false);
    deleteAnnouncementSpy.mockResolvedValue({});
    fireEvent.click(screen.getByRole("button", { name: "刪除" }));
    await Promise.resolve();
    expect(deleteAnnouncementSpy).not.toHaveBeenCalled();

    confirmSpy.mockReturnValue(true);
    fireEvent.click(screen.getByRole("button", { name: "刪除" }));
    await waitFor(() => expect(deleteAnnouncementSpy).toHaveBeenCalledWith({ id: "a-1" }));
    confirmSpy.mockRestore();
  });
});
