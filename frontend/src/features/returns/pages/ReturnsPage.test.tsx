import { fireEvent, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code } from "@connectrpc/connect";
import { apiError } from "@/test-api-error";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient } from "@tanstack/solid-query";
import { renderWithProviders } from "~/test-render";

// 退貨 API 以 spy 取代：ReturnsPage 在模組層建立 connect client，
// 以 createClient 的替身同時攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const {
  listReturnRequestsSpy,
  getReturnRequestSpy,
  reviewReturnRequestSpy,
  getReturnCertificateSpy,
  getCustomerSpy,
} = vi.hoisted(() => ({
  listReturnRequestsSpy: vi.fn(),
  getReturnRequestSpy: vi.fn(),
  reviewReturnRequestSpy: vi.fn(),
  getReturnCertificateSpy: vi.fn(),
  getCustomerSpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listReturnRequests: listReturnRequestsSpy,
    getReturnRequest: getReturnRequestSpy,
    reviewReturnRequest: reviewReturnRequestSpy,
    getReturnCertificate: getReturnCertificateSpy,
    getCustomer: getCustomerSpy,
  }),
}));

import ReturnsPage from "./ReturnsPage";

const PENDING_RETURN = {
  id: "rr-1",
  customerId: "cu-1",
  status: "pending",
  remark: "外箱破損",
  createdAt: "2026-09-20T10:00:00Z",
};

const APPROVED_RETURN = {
  id: "rr-2",
  customerId: "cu-1",
  status: "approved",
  remark: "客戶要求更換",
  createdAt: "2026-09-18T09:30:00Z",
};

const PENDING_DETAIL = {
  id: "rr-1",
  customerId: "cu-1",
  status: "pending",
  remark: "外箱破損",
  rejectReason: "",
  createdAt: "2026-09-20T10:00:00Z",
  items: [
    {
      id: "ri-1",
      sourceType: "order_item",
      productName: "洋芋片",
      spec: "大包",
      unit: "箱",
      quantity: 2,
      reason: "外包裝凹陷",
      photoUrls: ["https://example.com/photo-1.jpg"],
    },
  ],
  // 樂觀鎖版本：契約測試要證明審核送出的是這個值、絕非寫死的 "0"。
  version: "7",
};

const APPROVED_DETAIL = {
  ...PENDING_DETAIL,
  id: "rr-2",
  status: "approved",
  version: "3",
};

const CERTIFICATE = {
  id: "rr-2",
  customerCode: "C0001",
  customerName: "○○食品行",
  createdAt: "2026-09-18T09:30:00Z",
  status: "approved",
  reviewerName: "王大明",
  reviewedAt: "2026-09-19T08:00:00Z",
  items: APPROVED_DETAIL.items,
};

/** 每個測試一份全新的 `QueryClient`：快取不跨測試殘留（retry 關閉，理由同 CustomersPage）。 */
function newClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function mountPage(client: QueryClient = newClient()) {
  renderWithProviders(() => <ReturnsPage />, client);
}

async function renderPage() {
  mountPage();
  await waitFor(() => expect(screen.getByText("外箱破損")).toBeTruthy());
}

/** 開啟第一列的明細，等到 getReturnRequest 打完**且資料已渲染**（footer 按鈕依資料現身）。 */
async function openDetail() {
  fireEvent.click(screen.getByRole("button", { name: "查看" }));
  await waitFor(() => expect(getReturnRequestSpy).toHaveBeenCalledWith({ id: "rr-1" }));
  await waitFor(() => expect(screen.getByRole("button", { name: "核准" })).toBeTruthy());
}

/**
 * 核准的確認框：明細框開著時兩者並存於 DOM，所以依 accessible name（＝標題）取，
 * 不用 `getByRole("dialog")`（會同時命中兩個而拋錯）。
 */
async function confirmDialog() {
  return await waitFor(() =>
    screen.getByRole("dialog", { name: "核准退貨申請" })
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  listReturnRequestsSpy.mockResolvedValue({
    entries: [PENDING_RETURN],
    page: 1,
    pageSize: 20,
    total: 1,
  });
  getReturnRequestSpy.mockResolvedValue(PENDING_DETAIL);
  getCustomerSpy.mockResolvedValue({ customer: { id: "cu-1", name: "○○食品行" } });
  reviewReturnRequestSpy.mockResolvedValue({
    id: "rr-1",
    status: "approved",
    reviewedAt: "2026-09-21T00:00:00Z",
  });
  getReturnCertificateSpy.mockResolvedValue(CERTIFICATE);
});

describe("ReturnsPage", () => {
  it("清單載入後顯示狀態、備註與建立時間", async () => {
    await renderPage();
    expect(listReturnRequestsSpy).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, status: "" })
    );
    // 篩選下拉也有「待審核」字樣 → 狀態徽章斷言限縮在表格內。
    const table = screen.getByRole("table");
    expect(within(table).getByText("待審核")).toBeTruthy();
    expect(within(table).getByText("外箱破損")).toBeTruthy();
    expect(within(table).getByText("2026-09-20 18:00:00")).toBeTruthy();
    expect(within(table).getByText("#cu-1")).toBeTruthy();
  });

  it("送出狀態篩選才進 query key 並回第 1 頁", async () => {
    await renderPage();
    listReturnRequestsSpy.mockClear();
    fireEvent.change(screen.getByLabelText("狀態"), { target: { value: "pending" } });
    // 沒送出前不得重查（草稿只進 signal）。
    await Promise.resolve();
    expect(listReturnRequestsSpy).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listReturnRequestsSpy).toHaveBeenCalledWith(
        expect.objectContaining({ status: "pending", page: 1 })
      )
    );
  });

  it("開啟查看會呼叫 getReturnRequest 並顯示品項快照", async () => {
    await renderPage();
    await openDetail();
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    await waitFor(() => expect(within(dialog).getByText("洋芋片")).toBeTruthy());
    expect(within(dialog).getByText("大包")).toBeTruthy();
    expect(within(dialog).getByText("外包裝凹陷")).toBeTruthy();
    expect(within(dialog).getByText("來自訂單")).toBeTruthy();
    const photo = within(dialog).getByRole("link", { name: /photo-1/ });
    expect(photo.getAttribute("target")).toBe("_blank");
    expect(photo.getAttribute("rel")).toContain("noopener");
    // pending 沒有證明可看（證明入口只在已核准列的操作欄）；核准／駁回鈕則要現身。
    expect(within(dialog).queryByRole("button", { name: /證明/ })).toBeNull();
    expect(within(dialog).getByRole("button", { name: "核准" })).toBeTruthy();
  });

  it("核准送出明細回應的 version 當 expectedVersion（不是寫死的 0）", async () => {
    await renderPage();
    await openDetail();
    // 核准是不可逆的 → 走共用確認對話框（useConfirm），確認鈕是動詞「核准」。
    fireEvent.click(screen.getByRole("button", { name: "核准" }));
    const dialog = await confirmDialog();
    expect(within(dialog).getByText("核准退貨申請")).toBeTruthy();
    fireEvent.click(within(dialog).getByRole("button", { name: "核准" }));
    await waitFor(() => expect(reviewReturnRequestSpy).toHaveBeenCalledOnce());
    const arg = reviewReturnRequestSpy.mock.calls[0][0] as {
      id: string;
      decision: string;
      expectedVersion: string;
      rejectReason: string;
    };
    expect(arg).toEqual(
      expect.objectContaining({ id: "rr-1", decision: "approved", expectedVersion: "7" })
    );
    expect(arg.expectedVersion).not.toBe("0");
  });

  it("核准前要確認；拒絕確認就不打 API", async () => {
    await renderPage();
    await openDetail();
    fireEvent.click(screen.getByRole("button", { name: "核准" }));
    const dialog = await confirmDialog();
    fireEvent.click(within(dialog).getByRole("button", { name: "取消" }));
    await waitFor(() => expect(screen.getAllByRole("dialog")).toHaveLength(1));
    expect(reviewReturnRequestSpy).not.toHaveBeenCalled();
  });

  it("駁回：空白原因不打 API 並顯示欄位錯誤，填了才送出原因", async () => {
    await renderPage();
    await openDetail();
    fireEvent.click(screen.getByRole("button", { name: "駁回" }));
    const textarea = await waitFor(() => screen.getByLabelText(/駁回原因/));
    fireEvent.click(screen.getByRole("button", { name: "確認駁回" }));
    await waitFor(() => expect(screen.getByText("請輸入駁回原因")).toBeTruthy());
    expect(reviewReturnRequestSpy).not.toHaveBeenCalled();

    fireEvent.input(textarea, { target: { value: "品質瑕疵" } });
    fireEvent.click(screen.getByRole("button", { name: "確認駁回" }));
    await waitFor(() =>
      expect(reviewReturnRequestSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          id: "rr-1",
          decision: "rejected",
          rejectReason: "品質瑕疵",
        })
      )
    );
  });

  it("審核回 PermissionDenied 顯示沒有權限執行此操作", async () => {
    await renderPage();
    await openDetail();
    reviewReturnRequestSpy.mockRejectedValue(
      apiError("SYS-4001", "缺少權限", {}, Code.PermissionDenied)
    );
    fireEvent.click(screen.getByRole("button", { name: "核准" }));
    const dialog = await confirmDialog();
    fireEvent.click(within(dialog).getByRole("button", { name: "核准" }));
    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain("沒有權限執行此操作")
    );
  });

  it("版本衝突顯示資料已變更，請重新載入", async () => {
    await renderPage();
    await openDetail();
    // 清掉初次載入的呼叫，讓後面的斷言只對「衝突後的失效重取」成立。
    listReturnRequestsSpy.mockClear();
    reviewReturnRequestSpy.mockRejectedValue(
      apiError("SYS-1001", "參數驗證失敗", { reason: "資料已變更，請重新載入" }, Code.InvalidArgument)
    );
    fireEvent.click(screen.getByRole("button", { name: "核准" }));
    const dialog = await confirmDialog();
    fireEvent.click(within(dialog).getByRole("button", { name: "核准" }));
    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain("資料已變更，請重新載入")
    );
    // 衝突後要失效清單，讓使用者看到後端最新狀態。
    await waitFor(() =>
      expect(listReturnRequestsSpy).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1, status: "" })
      )
    );
  });

  it("已核准列才有證明入口，證明顯示審核人與審核時間", async () => {
    listReturnRequestsSpy.mockResolvedValue({
      entries: [APPROVED_RETURN],
      page: 1,
      pageSize: 20,
      total: 1,
    });
    getReturnRequestSpy.mockResolvedValue(APPROVED_DETAIL);
    mountPage();
    await waitFor(() => expect(screen.getByText("客戶要求更換")).toBeTruthy());
    // 憑證是唯讀快照 → 入口在列上（明細框維持唯讀，不同時開兩個對話框）。
    const certButton = await waitFor(() =>
      screen.getByRole("button", { name: "證明" })
    );
    fireEvent.click(certButton);
    await waitFor(() => expect(getReturnCertificateSpy).toHaveBeenCalledWith({ id: "rr-2" }));
    await waitFor(() => expect(screen.getByText("王大明")).toBeTruthy());
    expect(screen.getByText("2026-09-19 16:00:00")).toBeTruthy();
    expect(screen.getByText("C0001")).toBeTruthy();
    // 證明對話框開著時，明細對話框不該同時存在。
    expect(getReturnRequestSpy).not.toHaveBeenCalled();
  });

  it("列表顯示客戶代號，讓審核佇列能分辨是誰的申請", async () => {
    await renderPage();
    expect(within(screen.getByRole("table")).getByText("#cu-1")).toBeTruthy();
  });

  it("顯示子帳號發起說明且頁面沒有新增／建立入口", async () => {
    await renderPage();
    expect(screen.getByText(/退貨申請由客戶 App/)).toBeTruthy();
    expect(screen.queryByRole("button", { name: /新增|建立|建單/ })).toBeNull();
  });
});
