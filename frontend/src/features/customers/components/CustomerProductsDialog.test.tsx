import { fireEvent, render, screen, waitFor } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import type { Customer } from "~/lib/proto/customers/v1/customer_pb";

// 同 AddressBookDialog.test：攔 createClient（三支 client 都在 queries.ts 模組層建立），
// ConnectError/Code 保持真實（errorMessage 的對照才有效）。
const {
  listProductsSpy,
  listCustomerProductsSpy,
  addCustomerProductSpy,
  updateCustomerProductSpy,
  deleteCustomerProductSpy,
} = vi.hoisted(() => ({
  listProductsSpy: vi.fn(),
  listCustomerProductsSpy: vi.fn(),
  addCustomerProductSpy: vi.fn(),
  updateCustomerProductSpy: vi.fn(),
  deleteCustomerProductSpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listProducts: listProductsSpy,
    listCustomerProducts: listCustomerProductsSpy,
    addCustomerProduct: addCustomerProductSpy,
    updateCustomerProduct: updateCustomerProductSpy,
    deleteCustomerProduct: deleteCustomerProductSpy,
  }),
}));

import CustomerProductsDialog from "./CustomerProductsDialog";

const CUSTOMER = {
  id: "c1",
  companyId: "1",
  departmentId: "2",
  customerCode: "C0001",
  name: "○○食品行",
  taxId: "",
  paymentMethodId: "",
  settlementMethodId: "",
  customerTypeId: "",
  invoiceTypeId: "",
  defaultSalesRepId: "",
  preferredDeliveryDays: [],
  promoTagIds: [],
  createdAt: "2026-09-01T00:00:00Z",
  updatedAt: "2026-09-01T00:00:00Z",
  deletedAt: "",
} as unknown as Customer;

const PRODUCT = { id: "p9", code: "SKU-009", name: "牛腱" };

// defaultQty "0" 是「保留不顯示」語意 → UI 呈現「—」（不是 0）；
// cutNote 給非空值，讓頁面上唯一的「—」就是數量欄（空 cutNote 也會渲染「—」）。
const ROW = {
  id: "cp1",
  customerId: "c1",
  productId: "p9",
  aliasName: "牛肉片",
  defaultQty: "0",
  cutNote: "去筋膜",
  promoTagIds: [],
  createdAt: "2026-09-02T00:00:00Z",
  updatedAt: "2026-09-02T00:00:00Z",
};

// 每個測試一份全新 QueryClient：快取不跨測試殘留（retry 關閉，理由同 AddressBookDialog.test）。
async function renderDialog(
  client: QueryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
) {
  render(() => (
    <QueryClientProvider client={client}>
      <CustomerProductsDialog customer={CUSTOMER} open={true} onOpenChange={() => {}} />
    </QueryClientProvider>
  ));
  await waitFor(() => expect(screen.getByText("牛肉片")).toBeTruthy());
}

beforeEach(() => {
  vi.clearAllMocks();
  listProductsSpy.mockResolvedValue({ products: [PRODUCT], total: 1 });
  listCustomerProductsSpy.mockResolvedValue({ products: [ROW], total: 1 });
  addCustomerProductSpy.mockResolvedValue({ product: ROW, created: true });
  updateCustomerProductSpy.mockResolvedValue({ product: ROW });
  deleteCustomerProductSpy.mockResolvedValue({});
});

describe("CustomerProductsDialog", () => {
  it("以 customerId 列出專屬清單，0 數量顯示「—」", async () => {
    await renderDialog();
    expect(listCustomerProductsSpy).toHaveBeenCalledTimes(1);
    expect(listCustomerProductsSpy).toHaveBeenCalledWith(
      expect.objectContaining({ customerId: "c1" })
    );
    expect(screen.getByText("#p9")).toBeTruthy();
    // defaultQty "0" → 「保留不顯示」，UI 給「—」而不是 0。
    expect(screen.getByText("—")).toBeTruthy();
    expect(screen.queryByText("尚無專屬商品")).toBeNull();
  });

  it("清單為空時顯示尚無專屬商品", async () => {
    listCustomerProductsSpy.mockResolvedValue({ products: [], total: 0 });
    render(() => (
      <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
      >
        <CustomerProductsDialog customer={CUSTOMER} open={true} onOpenChange={() => {}} />
      </QueryClientProvider>
    ));
    await waitFor(() => expect(screen.getByText("尚無專屬商品")).toBeTruthy());
  });

  it("關鍵字打字不打 API，按「查詢產品」才打（草稿與生效值分離）", async () => {
    await renderDialog();
    // 開啟時以空關鍵字撈一次挑選器。
    await waitFor(() => expect(listProductsSpy).toHaveBeenCalledTimes(1));

    fireEvent.click(screen.getByRole("button", { name: "新增專屬商品" }));
    fireEvent.input(screen.getByLabelText("產品關鍵字"), { target: { value: "牛" } });
    // 排空微任務：證明「只有打字」不會觸發查詢（草稿若直接進 queryKey，這裡已多打一次）。
    const flushed = Promise.withResolvers<void>();
    setTimeout(flushed.resolve, 10);
    await flushed.promise;
    expect(listProductsSpy).toHaveBeenCalledTimes(1);

    fireEvent.click(screen.getByRole("button", { name: "查詢產品" }));
    await waitFor(() => expect(listProductsSpy).toHaveBeenCalledTimes(2));
    // 第二次請求帶的是新關鍵字（不是空字串重打一次）。
    expect(listProductsSpy.mock.calls[1][0]).not.toEqual(listProductsSpy.mock.calls[0][0]);
  });

  it("新增送出完整欄位（別名去空白）並重載清單", async () => {
    await renderDialog();
    fireEvent.click(screen.getByRole("button", { name: "新增專屬商品" }));
    await waitFor(() => expect(listProductsSpy).toHaveBeenCalledTimes(1));

    fireEvent.change(screen.getByLabelText("產品 *"), { target: { value: "p9" } });
    fireEvent.input(screen.getByLabelText("別名（留空＝用商品名）"), {
      target: { value: "  牛肉片  " },
    });
    fireEvent.input(screen.getByLabelText("預設數量（留空＝不預設）"), { target: { value: "2.5" } });
    fireEvent.click(screen.getByRole("button", { name: "新增" }));

    await waitFor(() =>
      expect(addCustomerProductSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          customerId: "c1",
          productId: "p9",
          aliasName: "牛肉片",
          defaultQty: "2.5",
        })
      )
    );
    expect(updateCustomerProductSpy).not.toHaveBeenCalled();
    // 寫入後 invalidate ["customers"] → 專屬清單重載。
    await waitFor(() => expect(listCustomerProductsSpy).toHaveBeenCalledTimes(2));
  });

  it("未選產品不送出並顯示欄位錯誤", async () => {
    await renderDialog();
    fireEvent.click(screen.getByRole("button", { name: "新增專屬商品" }));
    // submit 事件而非點按鈕：jsdom 原生 required 約束會擋下 TanStack 驗證
    //（同 AddressBookDialog.test 做法）。
    fireEvent.submit(document.querySelector("form")!);
    await waitFor(() => expect(screen.getByText("請選擇產品")).toBeTruthy());
    expect(addCustomerProductSpy).not.toHaveBeenCalled();
  });

  it("數量格式非法不送出", async () => {
    await renderDialog();
    fireEvent.click(screen.getByRole("button", { name: "新增專屬商品" }));
    fireEvent.change(screen.getByLabelText("產品 *"), { target: { value: "p9" } });
    fireEvent.input(screen.getByLabelText("預設數量（留空＝不預設）"), { target: { value: "abc" } });
    fireEvent.submit(document.querySelector("form")!);
    await waitFor(() =>
      expect(screen.getByText("數量須為十進位數字（可小數）")).toBeTruthy()
    );
    expect(addCustomerProductSpy).not.toHaveBeenCalled();
  });

  it("編輯不呈現產品欄，送出只帶 id 與三個可改欄位", async () => {
    await renderDialog();
    fireEvent.click(screen.getByRole("button", { name: "編輯" }));
    await waitFor(() => expect(screen.getByText(/產品：不可修改/)).toBeTruthy());
    expect(screen.queryByLabelText("產品 *")).toBeNull();

    fireEvent.input(screen.getByLabelText("別名（留空＝用商品名）"), {
      target: { value: "牛腱片" },
    });
    fireEvent.click(screen.getByRole("button", { name: "儲存" }));

    await waitFor(() =>
      expect(updateCustomerProductSpy).toHaveBeenCalledWith(
        expect.objectContaining({ id: "cp1", aliasName: "牛腱片" })
      )
    );
    expect(addCustomerProductSpy).not.toHaveBeenCalled();
    // 不可改欄位不進請求體。
    expect(updateCustomerProductSpy.mock.calls[0][0]).not.toHaveProperty("customerId");
    expect(updateCustomerProductSpy.mock.calls[0][0]).not.toHaveProperty("productId");
  });

  it("移除要確認；拒絕確認不打 API，接受才刪", async () => {
    await renderDialog();
    const decline = vi.spyOn(window, "confirm").mockReturnValue(false);
    fireEvent.click(screen.getByRole("button", { name: "移除" }));
    // 同 AddressBookDialog.test：confirm 是同步分支，一個微任務足以證明沒打 API。
    await Promise.resolve();
    expect(deleteCustomerProductSpy).not.toHaveBeenCalled();

    decline.mockReturnValue(true);
    fireEvent.click(screen.getByRole("button", { name: "移除" }));
    await waitFor(() => expect(deleteCustomerProductSpy).toHaveBeenCalledWith({ id: "cp1" }));
    // 刪後清單重載。
    await waitFor(() => expect(listCustomerProductsSpy).toHaveBeenCalledTimes(2));
  });
});
