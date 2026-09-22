import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 列印 API 以 spy 取代：頁面在模組層建立 connect client，
// 以 createClient 的替身同時攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const {
  listLogsSpy,
  previewSpy,
  printSpy,
  listRoutesSpy,
  listWarehousesSpy,
} = vi.hoisted(() => ({
  listLogsSpy: vi.fn(),
  previewSpy: vi.fn(),
  printSpy: vi.fn(),
  listRoutesSpy: vi.fn(),
  listWarehousesSpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listLogs: listLogsSpy,
    preview: previewSpy,
    print: printSpy,
    listRoutes: listRoutesSpy,
    listWarehouses: listWarehousesSpy,
  }),
}));

import PrintPage from "./PrintPage";

const ENTRY = {
  id: "pl-1",
  documentType: "dispatch_summary",
  routeId: "r-1",
  targetDate: "2026-09-22",
  printedBy: "7",
  printedAt: "2026-09-22T10:00:00Z",
  isReprint: false,
  reprintReason: "",
  downloadUrl: "/api/v1/files/dispatch_summary_2026-09-22.pdf/download",
};

/** 每個測試一份全新的 `QueryClient`：快取不跨測試殘留（retry 關閉，理由同 CustomersPage）。 */
function newClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function mountPage(client: QueryClient = newClient()) {
  render(() => (
    <QueryClientProvider client={client}>
      <PrintPage />
    </QueryClientProvider>
  ));
  return client;
}

/** 開列印對話框並回傳其 DOM（欄位查詢收斂到這個 dialog）。 */
async function openPrintDialog(): Promise<HTMLElement> {
  fireEvent.click(screen.getByRole("button", { name: "列印" }));
  return await waitFor(() => screen.getByRole("dialog"));
}

beforeEach(() => {
  vi.clearAllMocks();
  listLogsSpy.mockResolvedValue({
    entries: [ENTRY],
    page: 1,
    pageSize: 20,
    total: 1,
  });
  previewSpy.mockResolvedValue({
    previewId: "pv-1",
    fileAssetId: "fa-1",
    downloadUrl: ENTRY.downloadUrl,
  });
  printSpy.mockResolvedValue({
    printLogId: "pl-2",
    fileAssetId: "fa-2",
    downloadUrl: ENTRY.downloadUrl,
    isReprint: false,
  });
  listRoutesSpy.mockResolvedValue({ routes: [], pagination: { total: 0 } });
  listWarehousesSpy.mockResolvedValue({ warehouses: [], pagination: { total: 0 } });
});

describe("PrintPage", () => {
  it("列印紀錄載入：總數取 total（不是 pagination），且不送 sort/desc（proto 沒有這兩個參數）", async () => {
    mountPage();
    await waitFor(() => expect(screen.getByRole("link", { name: "下載" })).toBeTruthy());
    expect(listLogsSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        page: 1,
        pageSize: 20,
        documentType: "",
        routeId: "",
        dateFrom: "",
        dateTo: "",
      })
    );
    // 總數文案取自 `total`（`ListLogsResponse` 沒有 `pagination`，取錯會永遠顯示 0）。
    expect(screen.getByText(/共 1 筆列印紀錄/)).toBeTruthy();
    const sent = listLogsSpy.mock.calls[0][0] as Record<string, unknown>;
    expect("sort" in sent).toBe(false);
    expect("desc" in sent).toBe(false);
  });

  it("送出篩選才進 query key（D5：不得每按一鍵就查詢）", async () => {
    mountPage();
    await waitFor(() => expect(screen.getByRole("link", { name: "下載" })).toBeTruthy());
    listLogsSpy.mockClear();

    fireEvent.change(screen.getByLabelText("單據類型"), {
      target: { value: "picking_list" },
    });
    await Promise.resolve();
    expect(listLogsSpy).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listLogsSpy).toHaveBeenCalledWith(
        expect.objectContaining({ documentType: "picking_list" })
      )
    );
    // 篩選換了 query key → 回第 1 頁。
    expect(listLogsSpy).toHaveBeenCalledWith(expect.objectContaining({ page: 1 }));
  });

  it("下載連結直接用 downloadUrl 當 href（後端產的是相對 API 路徑，同源 cookie 認證）", async () => {
    mountPage();
    await waitFor(() => expect(screen.getByRole("link", { name: "下載" })).toBeTruthy());
    const link = screen.getByRole("link", { name: "下載" });
    expect(link.getAttribute("href")).toBe(ENTRY.downloadUrl);
    expect(link.getAttribute("target")).toBe("_blank");
  });

  it("首印狀態顯示「首印」，重印狀態顯示「重印」（由後端 isReprint 決定，前端不自行推導）", async () => {
    listLogsSpy.mockResolvedValue({
      entries: [
        ENTRY,
        { ...ENTRY, id: "pl-2", documentType: "picking_list", isReprint: true },
      ],
      page: 1,
      pageSize: 20,
      total: 2,
    });
    mountPage();
    await waitFor(() => expect(screen.getByText("首印")).toBeTruthy());
    expect(screen.getByText("重印")).toBeTruthy();
    // 前端不推導：同一批資料若把 isReprint 全設 false，就不該出現「重印」。
    expect(screen.getAllByText("重印").length).toBe(1);
  });

  it("對點單未選店家：前端先擋，不打 preview API", async () => {
    const openSpy = vi.spyOn(window, "open").mockReturnValue(null);
    mountPage();
    const dialog = await openPrintDialog();
    fireEvent.change(within(dialog).getByLabelText("單據類型 *"), {
      target: { value: "delivery_note" },
    });
    fireEvent.input(within(dialog).getByLabelText("車次 *"), {
      target: { value: "r-1" },
    });
    fireEvent.change(within(dialog).getByLabelText("出貨日期 *"), {
      target: { value: "2026-09-22" },
    });
    fireEvent.click(within(dialog).getByRole("button", { name: "預覽" }));

    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain("對點單需選擇店家")
    );
    expect(previewSpy).not.toHaveBeenCalled();
    openSpy.mockRestore();
  });

  it("預覽成功：呼 preview 並用 downloadUrl 開新分頁（預覽不寫 print_logs，由後端保證）", async () => {
    const openSpy = vi.spyOn(window, "open").mockReturnValue(null);
    mountPage();
    const dialog = await openPrintDialog();
    fireEvent.change(within(dialog).getByLabelText("單據類型 *"), {
      target: { value: "dispatch_summary" },
    });
    fireEvent.input(within(dialog).getByLabelText("車次 *"), {
      target: { value: "r-1" },
    });
    fireEvent.change(within(dialog).getByLabelText("出貨日期 *"), {
      target: { value: "2026-09-22" },
    });
    fireEvent.click(within(dialog).getByRole("button", { name: "預覽" }));

    await waitFor(() => expect(previewSpy).toHaveBeenCalledOnce());
    expect(previewSpy).toHaveBeenCalledWith({
      documentType: "dispatch_summary",
      routeId: "r-1",
      targetDate: "2026-09-22",
      customerId: "",
      warehouseId: "",
    });
    await waitFor(() => expect(openSpy).toHaveBeenCalled());
    expect(openSpy).toHaveBeenCalledWith(ENTRY.downloadUrl, "_blank", "noopener,noreferrer");
    // 預覽不動 printLogs（不寫 print_logs → 不該失效列印紀錄查詢）。
    expect(printSpy).not.toHaveBeenCalled();
    openSpy.mockRestore();
  });

  it("首印成功：顯示「首印完成」並失效 printLogs；重印欄位不出現", async () => {
    const client = newClient();
    const invalidate = vi.spyOn(client, "invalidateQueries");
    mountPage(client);
    const dialog = await openPrintDialog();
    fireEvent.change(within(dialog).getByLabelText("單據類型 *"), {
      target: { value: "dispatch_summary" },
    });
    fireEvent.input(within(dialog).getByLabelText("車次 *"), {
      target: { value: "r-1" },
    });
    fireEvent.change(within(dialog).getByLabelText("出貨日期 *"), {
      target: { value: "2026-09-22" },
    });
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(printSpy).toHaveBeenCalledOnce());
    expect(printSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        documentType: "dispatch_summary",
        routeId: "r-1",
        targetDate: "2026-09-22",
        reprintReason: "",
      })
    );
    await waitFor(() => expect(screen.getByText("首印完成")).toBeTruthy());
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ["printLogs"] });
    // 首印不該出現重印原因欄。
    expect(within(dialog).getByLabelText("重印原因")).toBeTruthy();
  });

  it("重印原因欄常駐，且原因原樣送給後端（是否重印由後端判斷）", async () => {
    mountPage();
    const dialog = await openPrintDialog();
    // 欄位一開表單就在：重印與否只有後端 hasPrintLog 知道，
    // 等收到 isReprint 才現身的檢查永遠來不及（那時已印完）。
    expect(within(dialog).getByLabelText("重印原因")).toBeTruthy();

    fireEvent.change(within(dialog).getByLabelText("單據類型 *"), {
      target: { value: "dispatch_summary" },
    });
    fireEvent.input(within(dialog).getByLabelText("車次 *"), {
      target: { value: "r-1" },
    });
    fireEvent.change(within(dialog).getByLabelText("出貨日期 *"), {
      target: { value: "2026-09-22" },
    });
    fireEvent.input(within(dialog).getByLabelText("重印原因"), {
      target: { value: "司機未帶單" },
    });
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(printSpy).toHaveBeenCalledOnce());
    expect(printSpy).toHaveBeenCalledWith(
      expect.objectContaining({ reprintReason: "司機未帶單" })
    );
  });

  it("無權限時顯示 banner", async () => {
    listLogsSpy.mockRejectedValue(new ConnectError("denied", Code.PermissionDenied));
    mountPage();
    await waitFor(() => expect(screen.getByRole("alert")).toBeTruthy());
    expect(screen.getByRole("alert").textContent).toContain("沒有權限");
  });
});
