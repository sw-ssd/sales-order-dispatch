import { fireEvent, render, screen, waitFor } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import type { Customer } from "~/lib/proto/customers/v1/customer_pb";

// 地址簿/聯絡人 API 以 spy 取代：元件在模組層建立 connect client，
// 以 createClient 的替身同時攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const {
  listAddressesSpy,
  addAddressSpy,
  updateAddressSpy,
  deleteAddressSpy,
  listContactsSpy,
  addContactSpy,
  updateContactSpy,
  deleteContactSpy,
} = vi.hoisted(() => ({
  listAddressesSpy: vi.fn(),
  addAddressSpy: vi.fn(),
  updateAddressSpy: vi.fn(),
  deleteAddressSpy: vi.fn(),
  listContactsSpy: vi.fn(),
  addContactSpy: vi.fn(),
  updateContactSpy: vi.fn(),
  deleteContactSpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listAddresses: listAddressesSpy,
    addAddress: addAddressSpy,
    updateAddress: updateAddressSpy,
    deleteAddress: deleteAddressSpy,
    listContacts: listContactsSpy,
    addContact: addContactSpy,
    updateContact: updateContactSpy,
    deleteContact: deleteContactSpy,
  }),
}));

import AddressBookDialog from "./AddressBookDialog";

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

const ADDRESS = {
  id: "a1",
  customerId: "c1",
  type: "shipping",
  recipientName: "王小明",
  phone: "0912345678",
  addressLine: "台北市信義路一段 1 號",
  city: "台北市",
  postalCode: "100",
  isDefault: true,
  createdAt: "2026-09-02T00:00:00Z",
  updatedAt: "2026-09-02T00:00:00Z",
  deletedAt: "",
};

const CONTACT = {
  id: "p1",
  customerId: "c1",
  name: "李小姐",
  title: "採購",
  email: "lee@example.com",
  phone: "0212345678",
  isDefault: true,
  createdAt: "2026-09-02T00:00:00Z",
  updatedAt: "2026-09-02T00:00:00Z",
  deletedAt: "",
};

/** 每個測試一份全新的 `QueryClient`：快取不跨測試殘留（retry 關閉，理由同 CustomersPage）。 */
function newClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

async function renderDialog(client: QueryClient = newClient()) {
  render(() => (
    <QueryClientProvider client={client}>
      <AddressBookDialog customer={CUSTOMER} open={true} onOpenChange={() => {}} />
    </QueryClientProvider>
  ));
  // 兩段清單都要到位（dialog 內的表格）。
  await waitFor(() => expect(screen.getByText("王小明")).toBeTruthy());
  await waitFor(() => expect(screen.getByText("李小姐")).toBeTruthy());
}

beforeEach(() => {
  vi.clearAllMocks();
  listAddressesSpy.mockResolvedValue({ addresses: [ADDRESS] });
  listContactsSpy.mockResolvedValue({ contacts: [CONTACT] });
  addAddressSpy.mockResolvedValue({ address: ADDRESS });
  updateAddressSpy.mockResolvedValue({ address: ADDRESS });
  deleteAddressSpy.mockResolvedValue({});
  addContactSpy.mockResolvedValue({ contact: CONTACT });
  updateContactSpy.mockResolvedValue({ contact: CONTACT });
  deleteContactSpy.mockResolvedValue({});
});

describe("AddressBookDialog", () => {
  it("同時列出地址簿與聯絡人，並顯示類型與預設標籤", async () => {
    await renderDialog();
    expect(screen.getByText("出貨")).toBeTruthy();
    expect(screen.getByText("台北市信義路一段 1 號")).toBeTruthy();
    // 預設徽章兩段各一（地址與聯絡人都是 isDefault）。
    expect(screen.getAllByText("預設")).toHaveLength(2);
    expect(screen.getByText("採購")).toBeTruthy();
    expect(screen.getByText("lee@example.com")).toBeTruthy();
    // 兩支清單 RPC 都只呼叫一次（各自的 query）。
    expect(listAddressesSpy).toHaveBeenCalledTimes(1);
    expect(listContactsSpy).toHaveBeenCalledTimes(1);
  });

  it("新增地址送出完整欄位並失效客戶快取", async () => {
    await renderDialog();
    fireEvent.click(screen.getByRole("button", { name: "新增地址" }));
    fireEvent.input(screen.getByLabelText("收件人 *"), { target: { value: "陳小弟" } });
    fireEvent.input(screen.getByLabelText("地址 *"), { target: { value: "新北市中正路 2 號" } });
    fireEvent.click(screen.getByRole("button", { name: "新增" }));
    await waitFor(() =>
      expect(addAddressSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          customerId: "c1",
          type: "shipping",
          recipientName: "陳小弟",
          addressLine: "新北市中正路 2 號",
        })
      )
    );
    // 寫入後清單重載（invalidate ["customers"] 前綴涵蓋 addresses key）。
    await waitFor(() => expect(listAddressesSpy).toHaveBeenCalledTimes(2));
  });

  it("必填欄位空白時不打 API，顯示欄位錯誤", async () => {
    await renderDialog();
    fireEvent.click(screen.getByRole("button", { name: "新增地址" }));
    // 用 submit 事件而非點按鈕：jsdom 對帶 `required` 的表單會以原生約束擋下送出，
    // TanStack 的欄位驗證就不會跑（同 `CustomersPage.test` 的做法）。
    fireEvent.submit(document.querySelector("form")!);
    await waitFor(() => expect(screen.getByText("請輸入收件人")).toBeTruthy());
    expect(addAddressSpy).not.toHaveBeenCalled();
    expect(screen.getByText("請輸入地址")).toBeTruthy();
  });

  it("編輯地址預填既有值，儲存時帶 id", async () => {
    await renderDialog();
    const editButtons = screen.getAllByRole("button", { name: "編輯" });
    fireEvent.click(editButtons[0]); // 第一段＝地址
    await waitFor(() =>
      expect((screen.getByLabelText("收件人 *") as HTMLInputElement).value).toBe("王小明")
    );
    fireEvent.input(screen.getByLabelText("電話"), { target: { value: "0999999999" } });
    fireEvent.click(screen.getByRole("button", { name: "儲存" }));
    await waitFor(() =>
      expect(updateAddressSpy).toHaveBeenCalledWith(
        expect.objectContaining({ id: "a1", recipientName: "王小明", phone: "0999999999" })
      )
    );
    expect(addAddressSpy).not.toHaveBeenCalled();
  });

  it("刪除地址要確認；拒絕確認就不打 API", async () => {
    await renderDialog();
    const decline = vi.spyOn(window, "confirm").mockReturnValue(false);
    const delButtons = screen.getAllByRole("button", { name: "刪除" });
    fireEvent.click(delButtons[0]);
    await Promise.resolve();
    expect(deleteAddressSpy).not.toHaveBeenCalled();

    decline.mockReturnValue(true);
    fireEvent.click(screen.getAllByRole("button", { name: "刪除" })[0]);
    await waitFor(() => expect(deleteAddressSpy).toHaveBeenCalledWith({ id: "a1" }));
    await waitFor(() => expect(listAddressesSpy).toHaveBeenCalledTimes(2));
    decline.mockRestore();
  });

  it("email 格式非法時前端先擋，不打 API", async () => {
    await renderDialog();
    fireEvent.click(screen.getByRole("button", { name: "新增聯絡人" }));
    fireEvent.input(screen.getByLabelText("姓名 *"), { target: { value: "新聯絡人" } });
    fireEvent.input(screen.getByLabelText("Email"), { target: { value: "not-an-email" } });
    fireEvent.submit(document.querySelector("form")!);
    await waitFor(() => expect(screen.getByText("email 格式非法")).toBeTruthy());
    expect(addContactSpy).not.toHaveBeenCalled();
  });

  it("新增聯絡人送出欄位", async () => {
    await renderDialog();
    fireEvent.click(screen.getByRole("button", { name: "新增聯絡人" }));
    fireEvent.input(screen.getByLabelText("姓名 *"), { target: { value: "新聯絡人" } });
    fireEvent.click(screen.getByRole("button", { name: "新增" }));
    await waitFor(() =>
      expect(addContactSpy).toHaveBeenCalledWith(
        expect.objectContaining({ customerId: "c1", name: "新聯絡人" })
      )
    );
    await waitFor(() => expect(listContactsSpy).toHaveBeenCalledTimes(2));
  });

  it("後端回 PermissionDenied 顯示沒有權限執行此操作", async () => {
    await renderDialog();
    addAddressSpy.mockRejectedValue(new ConnectError("denied", Code.PermissionDenied));
    fireEvent.click(screen.getByRole("button", { name: "新增地址" }));
    fireEvent.input(screen.getByLabelText("收件人 *"), { target: { value: "陳小弟" } });
    fireEvent.input(screen.getByLabelText("地址 *"), { target: { value: "新北市中正路 2 號" } });
    fireEvent.click(screen.getByRole("button", { name: "新增" }));
    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain("沒有權限執行此操作")
    );
  });

  it("兩段皆空時顯示各自的空狀態", async () => {
    listAddressesSpy.mockResolvedValue({ addresses: [] });
    listContactsSpy.mockResolvedValue({ contacts: [] });
    render(() => (
      <QueryClientProvider client={newClient()}>
        <AddressBookDialog customer={CUSTOMER} open={true} onOpenChange={() => {}} />
      </QueryClientProvider>
    ));
    await waitFor(() => expect(screen.getByText("尚無地址")).toBeTruthy());
    expect(screen.getByText("尚無聯絡人")).toBeTruthy();
  });
});
