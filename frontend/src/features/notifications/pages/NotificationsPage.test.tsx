import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 通知 API 以 spy 取代：NotificationsPage 在模組層建立 connect client，
// 以 createClient 的替身同時攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const { listNotificationsSpy, markReadSpy } = vi.hoisted(() => ({
  listNotificationsSpy: vi.fn(),
  markReadSpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listNotifications: listNotificationsSpy,
    markRead: markReadSpy,
  }),
}));

import NotificationsPage from "./NotificationsPage";

const UNREAD = {
  id: "n-1",
  channel: "in_app",
  title: "訂單已成立",
  content: "訂單 SO-000123 已由業務確認",
  payload: '{"sales_order_id":"123"}',
  status: "sent",
  sentAt: "2026-09-20T10:00:05Z",
  readAt: "",
  createdAt: "2026-09-20T10:00:00Z",
};

const READ = {
  ...UNREAD,
  id: "n-2",
  title: "退貨已核准",
  content: "你的退貨申請已核准",
  status: "read",
  readAt: "2026-09-20T11:00:00Z",
  createdAt: "2026-09-19T09:00:00Z",
};

const FAILED = {
  ...UNREAD,
  id: "n-3",
  channel: "fcm",
  title: "推播失敗示例",
  content: "這則推播沒有送達裝置",
  status: "failed",
  sentAt: "",
  createdAt: "2026-09-18T08:00:00Z",
};

/** 每個測試一份全新的 `QueryClient`：快取不跨測試殘留（retry 關閉，理由同 CustomersPage）。 */
function newClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function mountPage(client: QueryClient = newClient()) {
  render(() => (
    <QueryClientProvider client={client}>
      <NotificationsPage />
    </QueryClientProvider>
  ));
}

async function renderPage(
  list: unknown[] = [UNREAD, READ, FAILED],
  total = 3,
  unread = 1,
  waitText = "訂單已成立"
) {
  listNotificationsSpy.mockResolvedValue({
    notifications: list,
    page: 1,
    pageSize: 20,
    total,
    unreadCount: unread,
  });
  mountPage();
  await waitFor(() => expect(screen.getByText(waitText)).toBeTruthy());
}

beforeEach(() => {
  vi.clearAllMocks();
  markReadSpy.mockResolvedValue({ markedCount: 1 });
});

describe("NotificationsPage", () => {
  it("清單顯示未讀數、狀態標籤、標題內文與通道", async () => {
    await renderPage();
    // 未讀數取自 ListNotificationsResponse.unread_count（不另呼叫 UnreadCount）。
    expect(screen.getByText(/未讀 1 筆/)).toBeTruthy();
    const table = screen.getByRole("table");
    expect(within(table).getByText("未讀")).toBeTruthy();
    expect(within(table).getByText("已讀")).toBeTruthy();
    expect(within(table).getByText("送達失敗")).toBeTruthy();
    expect(within(table).getByText("訂單 SO-000123 已由業務確認")).toBeTruthy();
    // 通道標籤：UNREAD/READ 都是 in_app（2 個「站內」）、FAILED 是 fcm（1 個「推播」）。
    expect(within(table).getAllByText("站內")).toHaveLength(2);
    expect(within(table).getAllByText("推播")).toHaveLength(1);
    expect(within(table).getByText("2026-09-20 10:00:00")).toBeTruthy();
  });

  it("送出只看未讀才進 query key 並回第 1 頁", async () => {
    await renderPage();
    listNotificationsSpy.mockClear();
    fireEvent.change(screen.getByLabelText("已讀狀態"), { target: { value: "unread" } });
    // 沒送出前不得重查（草稿只進 signal）。
    await Promise.resolve();
    expect(listNotificationsSpy).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listNotificationsSpy).toHaveBeenCalledWith(
        expect.objectContaining({ unreadOnly: true, page: 1 })
      )
    );
  });

  it("送出通道篩選才把 channel 放進請求", async () => {
    await renderPage();
    listNotificationsSpy.mockClear();
    fireEvent.change(screen.getByLabelText("通道"), { target: { value: "in_app" } });
    await Promise.resolve();
    expect(listNotificationsSpy).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listNotificationsSpy).toHaveBeenCalledWith(
        expect.objectContaining({ channel: "in_app", page: 1 })
      )
    );
  });

  it("單列標記已讀只送那一筆，並失效清單讓未讀數刷新", async () => {
    await renderPage();
    const buttons = screen.getAllByRole("button", { name: "標記已讀" });
    // 三個狀態裡只有 pending/sent 算未讀 → 一列可標（read/failed 沒有按鈕）。
    expect(buttons).toHaveLength(1);
    fireEvent.click(buttons[0]);
    await waitFor(() =>
      expect(markReadSpy).toHaveBeenCalledWith({ notificationIds: ["n-1"] })
    );
    await waitFor(() =>
      expect(listNotificationsSpy).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1 })
      )
    );
    expect(screen.getByRole("status").textContent).toContain("已標記 1 筆已讀");
  });

  it("標記本頁已讀把本頁未讀 id 一次送出（failed 不含在內）", async () => {
    listNotificationsSpy.mockResolvedValue({
      notifications: [
        UNREAD,
        {
          ...UNREAD,
          id: "n-4",
          title: "專屬商品已上架",
          content: "你專屬的 3 項商品已可訂購",
          status: "pending",
        },
        READ,
        FAILED,
      ],
      page: 1,
      pageSize: 20,
      total: 4,
      unreadCount: 2,
    });
    mountPage();
    await waitFor(() => expect(screen.getByText("訂單已成立")).toBeTruthy());
    fireEvent.click(screen.getByRole("button", { name: /標記本頁已讀/ }));
    await waitFor(() =>
      expect(markReadSpy).toHaveBeenCalledWith({ notificationIds: ["n-1", "n-4"] })
    );
    const sent = markReadSpy.mock.calls[0][0].notificationIds as string[];
    expect(sent).not.toContain("n-3"); // failed 不可轉已讀，後端也只會略過
    expect(sent).not.toContain("n-2"); // 已讀是終態
  });

  it("沒有未讀時不出現標記本頁已讀", async () => {
    await renderPage([READ], 1, 0, "退貨已核准");
    expect(screen.queryByRole("button", { name: /標記本頁已讀/ })).toBeNull();
    expect(screen.queryByRole("button", { name: "標記已讀" })).toBeNull();
  });

  it("後端說部分未被異動時，回饋訊息不能宣稱全部完成", async () => {
    await renderPage();
    markReadSpy.mockResolvedValue({ markedCount: 0 });
    fireEvent.click(screen.getAllByRole("button", { name: "標記已讀" })[0]);
    await waitFor(() =>
      expect(screen.getByRole("status").textContent).toContain(
        "已標記 0 筆已讀（其餘狀態不需或無法轉為已讀）"
      )
    );
  });

  it("標記已讀回 PermissionDenied 顯示沒有權限執行此操作", async () => {
    await renderPage();
    markReadSpy.mockRejectedValue(new ConnectError("denied", Code.PermissionDenied));
    fireEvent.click(screen.getAllByRole("button", { name: "標記已讀" })[0]);
    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain("沒有權限執行此操作")
    );
  });

  it("清單失敗顯示錯誤；無資料顯示尚無通知", async () => {
    listNotificationsSpy.mockRejectedValueOnce(
      new ConnectError("offline", Code.Unavailable)
    );
    mountPage();
    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain("無法連線至伺服器")
    );

    listNotificationsSpy.mockResolvedValue({
      notifications: [],
      page: 1,
      pageSize: 20,
      total: 0,
      unreadCount: 0,
    });
    mountPage();
    await waitFor(() => expect(screen.getByText("尚無通知")).toBeTruthy());
  });
});
