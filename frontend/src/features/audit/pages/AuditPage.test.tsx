import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 稽核 API 以 spy 取代：AuditPage 在模組層建立 connect client，
// 以 createClient 的替身同時攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const { listAuditLogsSpy } = vi.hoisted(() => ({ listAuditLogsSpy: vi.fn() }));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({ listAuditLogs: listAuditLogsSpy }),
}));

import AuditPage from "./AuditPage";

const LOG = {
  id: "9",
  companyId: "1",
  departmentId: "",
  userId: "5",
  userName: "王小明",
  action: "delete",
  resourceType: "customer",
  resourceId: "42",
  beforeSnapshot: '{"name":"舊名稱"}',
  afterSnapshot: '{"deleted_at":"2026-09-20T10:00:00Z"}',
  ipAddress: "10.0.0.9",
  userAgent: "Mozilla/5.0 Test",
  createdAt: "2026-09-20T10:00:00Z",
};

/** 每個測試一份全新的 `QueryClient`：快取不跨測試殘留（retry 關閉，理由同 CustomersPage）。 */
function newClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function mountPage(client: QueryClient = newClient()) {
  render(() => (
    <QueryClientProvider client={client}>
      <AuditPage />
    </QueryClientProvider>
  ));
}

async function renderPage(
  items: unknown[] = [LOG],
  total = 1,
  waitText = "王小明"
) {
  listAuditLogsSpy.mockResolvedValue({
    items,
    pagination: { page: 1, pageSize: 20, total },
  });
  mountPage();
  await waitFor(() => expect(screen.getByText(waitText)).toBeTruthy());
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("AuditPage", () => {
  it("清單顯示時間、操作者、動作標籤、資源與 IP", async () => {
    await renderPage();
    expect(listAuditLogsSpy).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: 20, from: "", to: "", companyId: "" })
    );
    const table = screen.getByRole("table");
    expect(within(table).getByText("王小明")).toBeTruthy();
    expect(within(table).getByText("刪除")).toBeTruthy();
    expect(within(table).getByText("customer #42")).toBeTruthy();
    expect(within(table).getByText("10.0.0.9")).toBeTruthy();
    expect(within(table).getByText("2026-09-20 10:00:00")).toBeTruthy();
    // 保留期提示：未填時間即近 3 個月（D27）。
    expect(screen.getByText(/近 3 個月/)).toBeTruthy();
  });

  it("送出動作篩選才進 query key 並回第 1 頁", async () => {
    await renderPage();
    listAuditLogsSpy.mockClear();
    fireEvent.change(screen.getByLabelText("動作"), { target: { value: "update" } });
    // 沒送出前不得重查（草稿只進 signal）。
    await Promise.resolve();
    expect(listAuditLogsSpy).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listAuditLogsSpy).toHaveBeenCalledWith(
        expect.objectContaining({ action: "update", page: 1 })
      )
    );
  });

  it("起迄日轉成當地整天的 RFC3339（起 00:00、迄 23:59:59）", async () => {
    await renderPage();
    listAuditLogsSpy.mockClear();
    fireEvent.input(screen.getByLabelText("起日"), { target: { value: "2026-09-01" } });
    fireEvent.input(screen.getByLabelText("迄日"), { target: { value: "2026-09-22" } });
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listAuditLogsSpy).toHaveBeenCalledWith(expect.objectContaining({ page: 1 }))
    );
    const { from, to } = listAuditLogsSpy.mock.calls[0][0] as { from: string; to: string };
    // 斷言「當地日」而非字串樣式：字串會隨機器時區不同，當地日才契約語意。
    const fromDate = new Date(from);
    expect([fromDate.getFullYear(), fromDate.getMonth(), fromDate.getDate(), fromDate.getHours()]).toEqual([
      2026, 8, 1, 0,
    ]);
    const toDate = new Date(to);
    expect([toDate.getFullYear(), toDate.getMonth(), toDate.getDate(), toDate.getHours(), toDate.getMinutes(), toDate.getSeconds()]).toEqual([
      2026, 8, 22, 23, 59, 59,
    ]);
  });

  it("清空時間後送出空字串，交由後端套用近 3 個月預設窗", async () => {
    await renderPage();
    listAuditLogsSpy.mockClear();
    fireEvent.input(screen.getByLabelText("起日"), { target: { value: "2026-09-01" } });
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listAuditLogsSpy).toHaveBeenCalledWith(expect.objectContaining({ from: expect.any(String) }))
    );
    fireEvent.input(screen.getByLabelText("起日"), { target: { value: "" } });
    fireEvent.click(screen.getByRole("button", { name: "清空條件" }));
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listAuditLogsSpy).toHaveBeenCalledWith(
        expect.objectContaining({ from: "", to: "", action: "", resourceId: "", userId: "" })
      )
    );
  });

  it("檢視明細顯示格式化快照與 User-Agent，且不另開 RPC", async () => {
    await renderPage();
    fireEvent.click(screen.getByRole("button", { name: "檢視" }));
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    await waitFor(() => expect(within(dialog).getByText(/"name": "舊名稱"/)).toBeTruthy());
    expect(within(dialog).getByText(/"deleted_at"/)).toBeTruthy();
    expect(within(dialog).getByText(/Mozilla\/5\.0 Test/)).toBeTruthy();
    // AuditService 只有 ListAuditLogs：明細取自列上欄位，整段流程只該有一次請求。
    expect(listAuditLogsSpy).toHaveBeenCalledTimes(1);
  });

  it("後端回 PermissionDenied 顯示沒有權限執行此操作", async () => {
    listAuditLogsSpy.mockRejectedValue(new ConnectError("denied", Code.PermissionDenied));
    mountPage();
    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain("沒有權限執行此操作")
    );
  });

  it("無資料顯示尚無稽核紀錄，且頁面沒有任何寫入入口", async () => {
    await renderPage([], 0, "尚無稽核紀錄");
    expect(screen.getByText("尚無稽核紀錄")).toBeTruthy();
    expect(screen.queryByRole("button", { name: /新增|編輯|刪除/ })).toBeNull();
  });

  it("時間格式錯誤時原樣透傳後端訊息", async () => {
    listAuditLogsSpy.mockRejectedValue(
      new ConnectError('無效的 from "2026-99-99"', Code.InvalidArgument)
    );
    mountPage();
    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain('無效的 from "2026-99-99"')
    );
  });
});
