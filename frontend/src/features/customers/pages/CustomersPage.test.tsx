import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 客戶 API 以 spy 取代：CustomersPage 在模組層建立 connect client，
// 以 createClient 的替身同時攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const {
  listCustomersSpy,
  createCustomerSpy,
  updateCustomerSpy,
  deleteCustomerSpy,
  restoreCustomerSpy,
} = vi.hoisted(() => ({
  listCustomersSpy: vi.fn(),
  createCustomerSpy: vi.fn(),
  updateCustomerSpy: vi.fn(),
  deleteCustomerSpy: vi.fn(),
  restoreCustomerSpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listCustomers: listCustomersSpy,
    createCustomer: createCustomerSpy,
    updateCustomer: updateCustomerSpy,
    deleteCustomer: deleteCustomerSpy,
    restoreCustomer: restoreCustomerSpy,
  }),
}));

import CustomersPage from "./CustomersPage";

const EXISTING_CUSTOMER = {
  id: "cu-1",
  name: "○○食品行",
  customerCode: "C0001",
  taxId: "12345678",
  deletedAt: "",
};

const DELETED_CUSTOMER = {
  id: "cu-2",
  name: "已刪客戶",
  customerCode: "C0002",
  taxId: "",
  deletedAt: "2026-09-01T00:00:00Z",
};

/** 每個測試一份全新的 `QueryClient`：快取不跨測試殘留（retry 關閉，理由同 UsersPage）。 */
function newClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function mountPage(client: QueryClient = newClient()) {
  render(() => (
    <QueryClientProvider client={client}>
      <CustomersPage />
    </QueryClientProvider>
  ));
}

async function renderPage() {
  mountPage();
  await waitFor(() => expect(screen.getByText(/○○食品行/)).toBeTruthy());
}

beforeEach(() => {
  vi.clearAllMocks();
  listCustomersSpy.mockResolvedValue({
    customers: [EXISTING_CUSTOMER],
    pagination: { total: 1 },
  });
});

describe("CustomersPage", () => {
  it("清單載入後顯示客戶", async () => {
    await renderPage();
    expect(listCustomersSpy).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, sort: "", desc: false })
    );
    expect(screen.getByText("C0001")).toBeTruthy();
  });

  it("點表頭排序：sort/desc 顯式進請求並回第 1 頁", async () => {
    await renderPage();
    listCustomersSpy.mockClear();
    // 客戶名稱表頭是可排序的按鈕；第一次點＝升冪。
    fireEvent.click(screen.getByRole("button", { name: /客戶名稱/ }));
    await waitFor(() =>
      expect(listCustomersSpy).toHaveBeenCalledWith(
        expect.objectContaining({ sort: "name", desc: false, page: 1 })
      )
    );
  });

  it("送出關鍵字才進 query key", async () => {
    await renderPage();
    listCustomersSpy.mockClear();
    fireEvent.input(screen.getByLabelText("關鍵字"), {
      target: { value: "食品" },
    });
    // 沒送出前不得重查（D5：不得每按一鍵就查詢）。
    await Promise.resolve();
    expect(listCustomersSpy).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listCustomersSpy).toHaveBeenCalledWith(
        expect.objectContaining({ keyword: "食品" })
      )
    );
  });

  it("刪除先確認再呼叫並失效 customers 前綴", async () => {
    await renderPage();
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(false);
    deleteCustomerSpy.mockResolvedValue({});
    fireEvent.click(screen.getByRole("button", { name: "刪除" }));
    await Promise.resolve();
    // 未確認 → 不得呼叫 API。
    expect(deleteCustomerSpy).not.toHaveBeenCalled();
    confirmSpy.mockRestore();

    vi.spyOn(window, "confirm").mockReturnValue(true);
    listCustomersSpy.mockClear();
    fireEvent.click(screen.getByRole("button", { name: "刪除" }));
    await waitFor(() => expect(deleteCustomerSpy).toHaveBeenCalledOnce());
    await waitFor(() => expect(listCustomersSpy).toHaveBeenCalled());
    (window.confirm as unknown as { mockRestore: () => void }).mockRestore();
  });

  it("已刪除的客戶顯示還原而非刪除", async () => {
    listCustomersSpy.mockResolvedValue({
      customers: [DELETED_CUSTOMER],
      pagination: { total: 1 },
    });
    mountPage();
    await waitFor(() => expect(screen.getByText(/已刪客戶/)).toBeTruthy());
    expect(screen.getByText("已刪除")).toBeTruthy();
    expect(screen.queryByRole("button", { name: "刪除" })).toBeNull();
    const restore = screen.getByRole("button", { name: "還原" });
    restoreCustomerSpy.mockResolvedValue({});
    fireEvent.click(restore);
    await waitFor(() => expect(restoreCustomerSpy).toHaveBeenCalledOnce());
  });

  it("含已刪除切換會換 query key", async () => {
    await renderPage();
    listCustomersSpy.mockClear();
    fireEvent.click(screen.getByLabelText(/含已刪除/));
    await waitFor(() =>
      expect(listCustomersSpy).toHaveBeenCalledWith(
        expect.objectContaining({ includeDeleted: true })
      )
    );
  });

  it("建立成功後顯示僅此一次的帳號交付對話框", async () => {
    await renderPage();
    createCustomerSpy.mockResolvedValue({
      customer: { id: "cu-9" },
      primaryAccountName: "○○食品行",
      primaryTempPassword: "tmp-pass-1",
      salesRepAccountName: "○○食品行(業務)",
      salesRepTempPassword: "tmp-pass-2",
      accountManageUrl: "https://demo/customer_account_manage",
    });
    fireEvent.click(screen.getByRole("button", { name: "新增客戶" }));
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    fireEvent.input(screen.getByLabelText(/客戶名稱/), {
      target: { value: "新客戶" },
    });
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() =>
      expect(createCustomerSpy).toHaveBeenCalledWith(
        expect.objectContaining({ name: "新客戶" })
      )
    );
    // 臨時密碼只此一次 → 對話框要顯示兩組帳號與管理連結。
    const delivery = await waitFor(() => screen.getByText(/帳號已建立/));
    const box = delivery.closest('[role="dialog"]')!;
    expect(within(box as HTMLElement).getByText(/tmp-pass-1/)).toBeTruthy();
    expect(within(box as HTMLElement).getByText(/tmp-pass-2/)).toBeTruthy();
    expect(within(box as HTMLElement).getByText(/customer_account_manage/)).toBeTruthy();
  });

  it("名稱空白時不打 API", async () => {
    await renderPage();
    createCustomerSpy.mockResolvedValue({});
    fireEvent.click(screen.getByRole("button", { name: "新增客戶" }));
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(screen.getByText("請輸入客戶名稱")).toBeTruthy());
    expect(createCustomerSpy).not.toHaveBeenCalled();
  });

  it("無權限時顯示 banner", async () => {
    listCustomersSpy.mockRejectedValue(new ConnectError("denied", Code.PermissionDenied));
    mountPage();
    await waitFor(() => expect(screen.getByRole("alert")).toBeTruthy());
    expect(screen.getByRole("alert").textContent).toContain("沒有權限");
  });
});
