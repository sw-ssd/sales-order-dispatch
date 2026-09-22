import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 訂單／客戶／商品／字典 API 以 spy 取代：頁面在模組層建立 connect client，
// 以 createClient 的替身同時攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const {
  listOrdersSpy,
  getOrderSpy,
  createOrderSpy,
  updateOrderSpy,
  cancelOrderSpy,
  completeOrderSpy,
  voidOrderSpy,
  deleteOrderSpy,
  listOrderEventsSpy,
  listOptionsSpy,
  listProductsSpy,
  listCustomersSpy,
} = vi.hoisted(() => ({
  listOrdersSpy: vi.fn(),
  getOrderSpy: vi.fn(),
  createOrderSpy: vi.fn(),
  updateOrderSpy: vi.fn(),
  cancelOrderSpy: vi.fn(),
  completeOrderSpy: vi.fn(),
  voidOrderSpy: vi.fn(),
  deleteOrderSpy: vi.fn(),
  listOrderEventsSpy: vi.fn(),
  listOptionsSpy: vi.fn(),
  listProductsSpy: vi.fn(),
  listCustomersSpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listOrders: listOrdersSpy,
    getOrder: getOrderSpy,
    createOrder: createOrderSpy,
    updateOrder: updateOrderSpy,
    cancelOrder: cancelOrderSpy,
    completeOrder: completeOrderSpy,
    voidOrder: voidOrderSpy,
    deleteOrder: deleteOrderSpy,
    listOrderEvents: listOrderEventsSpy,
    listOptions: listOptionsSpy,
    listProducts: listProductsSpy,
    listCustomers: listCustomersSpy,
  }),
}));

import OrdersPage from "./OrdersPage";

/** 五個狀態各一筆（D13 狀態機），用來驗證動作依狀態開關。 */
const ROWS = [
  { id: "o-1", orderNo: "W-000001", customerId: "cu-1", source: "W", status: "pending", version: 1, expectedDeliveryDate: "2026-09-25", createdAt: "2026-09-22T10:00:00Z", note: "", deletedAt: "" },
  { id: "o-2", orderNo: "W-000002", customerId: "cu-1", source: "W", status: "processing", version: 1, expectedDeliveryDate: "", createdAt: "2026-09-22T11:00:00Z", note: "", deletedAt: "" },
  { id: "o-3", orderNo: "W-000003", customerId: "cu-1", source: "A", status: "completed", version: 1, expectedDeliveryDate: "", createdAt: "2026-09-22T12:00:00Z", note: "", deletedAt: "" },
  { id: "o-4", orderNo: "W-000004", customerId: "cu-1", source: "A", status: "voided", version: 1, expectedDeliveryDate: "", createdAt: "2026-09-22T13:00:00Z", note: "", deletedAt: "" },
];

const CUSTOMER = { id: "cu-1", name: "○○食品行", customerCode: "C0001", deletedAt: "" };
const PRODUCT = {
  id: "p-1",
  name: "蘋果",
  code: "P001",
  deletedAt: "",
  units: [{ unitCode: "KG", conversionRate: "1", isBase: true, sortOrder: 1, sizeDesc: "" }],
  processingSpecs: [],
};

/** 每個測試一份全新的 `QueryClient`：快取不跨測試殘留（retry 關閉，理由同 CustomersPage）。 */
function newClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function mountPage(client: QueryClient = newClient()) {
  render(() => (
    <QueryClientProvider client={client}>
      <OrdersPage />
    </QueryClientProvider>
  ));
}

async function renderPage() {
  mountPage();
  await waitFor(() => expect(screen.getByText("W-000001")).toBeTruthy());
}

/** 開建單對話框並回傳它的 DOM（欄位查詢都收斂到這個 dialog）。 */
async function openCreateDialog(): Promise<HTMLElement> {
  fireEvent.click(screen.getByRole("button", { name: "新增訂單" }));
  return await waitFor(() => screen.getByRole("dialog"));
}

beforeEach(() => {
  vi.clearAllMocks();
  // `ListOrdersResponse` 的總數在 `total`（不是 `pagination.total`）——契約差異要被釘住。
  listOrdersSpy.mockResolvedValue({ orders: ROWS, total: 4 });
  listCustomersSpy.mockResolvedValue({
    customers: [CUSTOMER],
    pagination: { total: 1 },
  });
  listProductsSpy.mockResolvedValue({ products: [PRODUCT], pagination: { total: 1 } });
  listOptionsSpy.mockResolvedValue({
    options: [
      { code: "W", displayName: "Web 中台" },
      { code: "A", displayName: "App" },
    ],
  });
  listOrderEventsSpy.mockResolvedValue({ events: [], total: 0 });
});

describe("OrdersPage", () => {
  it("清單載入：以 total 推總筆數，且不送 sort/desc（proto 沒有這兩個參數）", async () => {
    await renderPage();
    expect(listOrdersSpy).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: 20, status: "", customerId: "", source: "", keyword: "", includeDeleted: false })
    );
    // `共 {total()} 筆` 被 JSX 拆成多個文字節點，整段以正則比對。
    expect(screen.getByText(/銷售訂單\(共 4 筆\)/)).toBeTruthy();
    const sent = listOrdersSpy.mock.calls[0][0] as Record<string, unknown>;
    expect("sort" in sent).toBe(false);
    expect("desc" in sent).toBe(false);
  });

  it("動作依狀態開關：pending 可編輯/取消/刪除、processing 可完成、completed 可作廢、voided 無動作", async () => {
    await renderPage();
    // 各狀態各一筆 → 每個按鈕各只出現一次（有重複就代表狀態判斷錯）。
    expect(screen.getAllByRole("button", { name: "編輯" })).toHaveLength(1);
    expect(screen.getAllByRole("button", { name: "取消" })).toHaveLength(1);
    expect(screen.getAllByRole("button", { name: "刪除" })).toHaveLength(1);
    expect(screen.getAllByRole("button", { name: "完成" })).toHaveLength(1);
    expect(screen.getAllByRole("button", { name: "作廢" })).toHaveLength(1);
    // voided 是終態 → 除「詳情」外沒有任何動作。
    expect(screen.getAllByRole("button", { name: "詳情" })).toHaveLength(4);
  });

  it("送出篩選才進 query key", async () => {
    await renderPage();
    listOrdersSpy.mockClear();
    fireEvent.input(screen.getByLabelText("關鍵字"), { target: { value: "W-000001" } });
    // 沒送出前不得重查（D5：不得每按一鍵就查詢）。
    await Promise.resolve();
    expect(listOrdersSpy).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listOrdersSpy).toHaveBeenCalledWith(expect.objectContaining({ keyword: "W-000001" }))
    );
  });

  it("含已刪除切換會換 query key", async () => {
    await renderPage();
    listOrdersSpy.mockClear();
    fireEvent.click(screen.getByLabelText("含已刪除"));
    await waitFor(() =>
      expect(listOrdersSpy).toHaveBeenCalledWith(expect.objectContaining({ includeDeleted: true }))
    );
  });

  it("明細未填數量不打建單 API，且錯誤顯示在明細區", async () => {
    await renderPage();
    createOrderSpy.mockResolvedValue({ order: ROWS[0] });
    const dialog = await openCreateDialog();
    const customerSelect = within(dialog).getByLabelText("客戶 *") as HTMLSelectElement;
    fireEvent.change(customerSelect, { target: { value: "cu-1" } });
    const sourceSelect = within(dialog).getByLabelText("訂單來源 *") as HTMLSelectElement;
    fireEvent.change(sourceSelect, { target: { value: "W" } });
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(screen.getByText("數量不可為空")).toBeTruthy());
    expect(createOrderSpy).not.toHaveBeenCalled();
  });

  it("明細齊備才建單，並把手打列組成 OrderItemInput", async () => {
    await renderPage();
    createOrderSpy.mockResolvedValue({ order: ROWS[0] });
    const dialog = await openCreateDialog();
    const customerSelect = within(dialog).getByLabelText("客戶 *") as HTMLSelectElement;
    fireEvent.change(customerSelect, { target: { value: "cu-1" } });
    const sourceSelect = within(dialog).getByLabelText("訂單來源 *") as HTMLSelectElement;
    fireEvent.change(sourceSelect, { target: { value: "W" } });

    // 手打列：品名 + 單位 + 數量（商品留空即手打，後端視「帶 manualName」為手打）。
    fireEvent.input(within(dialog).getByLabelText("品名（手打）"), {
      target: { value: "手打品" },
    });
    fireEvent.input(within(dialog).getByLabelText("單位"), {
      target: { value: "KG" },
    });
    fireEvent.input(within(dialog).getByLabelText("數量 *"), {
      target: { value: "3" },
    });
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(createOrderSpy).toHaveBeenCalledOnce());
    const sent = createOrderSpy.mock.calls[0][0] as {
      customerId: string;
      source: string;
      items: { manualName: string; productId: string; qty: string; unit: string; displayName: string }[];
    };
    expect(sent.customerId).toBe("cu-1");
    expect(sent.source).toBe("W");
    expect(sent.items).toHaveLength(1);
    expect(sent.items[0]).toMatchObject({
      productId: "",
      manualName: "手打品",
      displayName: "手打品",
      qty: "3",
      unit: "KG",
    });
  });

  it("取消先確認再呼叫並失效 orders 前綴", async () => {
    await renderPage();
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(false);
    cancelOrderSpy.mockResolvedValue({});
    fireEvent.click(screen.getByRole("button", { name: "取消" }));
    await Promise.resolve();
    expect(cancelOrderSpy).not.toHaveBeenCalled();
    confirmSpy.mockRestore();

    vi.spyOn(window, "confirm").mockReturnValue(true);
    listOrdersSpy.mockClear();
    fireEvent.click(screen.getByRole("button", { name: "取消" }));
    await waitFor(() => expect(cancelOrderSpy).toHaveBeenCalledOnce());
    await waitFor(() => expect(listOrdersSpy).toHaveBeenCalled());
    (window.confirm as unknown as { mockRestore: () => void }).mockRestore();
  });

  it("作廢空原因不打 API", async () => {
    await renderPage();
    voidOrderSpy.mockResolvedValue({});
    fireEvent.click(screen.getByRole("button", { name: "作廢" }));
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    // 作廢對話框的確認是 button（沒有 <form>），直接點。
    fireEvent.click(within(dialog).getByRole("button", { name: "確認作廢" }));
    await waitFor(() => expect(screen.getByText("作廢需填原因")).toBeTruthy());
    expect(voidOrderSpy).not.toHaveBeenCalled();
  });

  it("詳情框顯示明細與事件軌跡", async () => {
    await renderPage();
    getOrderSpy.mockResolvedValue({
      order: ROWS[0],
      items: [
        {
          id: "i-1",
          productId: "",
          displayName: "手打品",
          qty: "3",
          unit: "KG",
          baseQty: "3",
          processingSpecId: "",
          specialCutNote: "去皮",
          warehouseId: "",
          sortOrder: 0,
        },
      ],
    });
    listOrderEventsSpy.mockResolvedValue({
      events: [
        {
          id: "e-1",
          eventType: "create",
          actorId: "1",
          reason: "",
          payload: "",
          createdAt: "2026-09-22T10:00:00Z",
        },
      ],
      total: 1,
    });
    fireEvent.click(screen.getAllByRole("button", { name: "詳情" })[0]);
    await waitFor(() => expect(screen.getByText("訂單詳情")).toBeTruthy());
    expect(await screen.findByText("手打品")).toBeTruthy();
    expect(screen.getByText("去皮")).toBeTruthy();
    // 「建立」同時是建單按鈕與事件標籤 → 收斂到事件軌跡區塊內比對。
    const eventsSection = screen.getByText("事件軌跡").closest("section")!;
    expect(within(eventsSection as HTMLElement).getByText("建立")).toBeTruthy();
  });

  it("無權限時顯示 banner", async () => {
    listOrdersSpy.mockRejectedValue(new ConnectError("denied", Code.PermissionDenied));
    mountPage();
    await waitFor(() => expect(screen.getByRole("alert")).toBeTruthy());
    expect(screen.getByRole("alert").textContent).toContain("沒有權限");
  });
});
