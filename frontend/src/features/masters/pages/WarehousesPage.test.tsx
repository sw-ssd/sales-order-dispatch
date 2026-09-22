import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 倉別 API 以 spy 取代：頁面在模組層建立 connect client，
// 以 createClient 的替身同時攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const {
  listWarehousesSpy,
  createWarehouseSpy,
  updateWarehouseSpy,
  deleteWarehouseSpy,
  restoreWarehouseSpy,
} = vi.hoisted(() => ({
  listWarehousesSpy: vi.fn(),
  createWarehouseSpy: vi.fn(),
  updateWarehouseSpy: vi.fn(),
  deleteWarehouseSpy: vi.fn(),
  restoreWarehouseSpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listWarehouses: listWarehousesSpy,
    createWarehouse: createWarehouseSpy,
    updateWarehouse: updateWarehouseSpy,
    deleteWarehouse: deleteWarehouseSpy,
    restoreWarehouse: restoreWarehouseSpy,
  }),
}));

import WarehousesPage from "./WarehousesPage";

const EXISTING_WAREHOUSE = {
  id: "w-1",
  companyId: "c-1",
  departmentId: "d-1",
  code: "WH1",
  name: "主倉",
  address: "台北市",
  isActive: true,
  createdAt: "2026-09-22T10:00:00Z",
  updatedAt: "2026-09-22T10:00:00Z",
  deletedAt: "",
};

const DELETED_WAREHOUSE = {
  ...EXISTING_WAREHOUSE,
  id: "w-2",
  code: "WH9",
  name: "已刪倉別",
  deletedAt: "2026-09-01T00:00:00Z",
};

/**
 * 掛上頁面：每次呼叫都給全新的 `QueryClient`（快取不跨測試殘留；retry 關閉，
 * 理由同 CustomersPage —— 失敗查詢不得自動重試拖慢失敗斷言）。
 * 回傳 client 供需要 spy `invalidateQueries` 的測試使用。
 */
function mountPage(
  client: QueryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
) {
  render(() => (
    <QueryClientProvider client={client}>
      <WarehousesPage />
    </QueryClientProvider>
  ));
  return client;
}

/** 標準情境：以預設 mock（一列 WH1）掛頁面並等清單落地。 */
async function renderPage() {
  mountPage();
  await waitFor(() => expect(screen.getByText("WH1")).toBeTruthy());
}

/** 開新增對話框並回傳它的 DOM（欄位查詢收斂到這個 dialog）。 */
async function openCreateDialog(): Promise<HTMLElement> {
  fireEvent.click(screen.getByRole("button", { name: "新增倉別" }));
  return await waitFor(() => screen.getByRole("dialog"));
}

beforeEach(() => {
  vi.clearAllMocks();
  listWarehousesSpy.mockResolvedValue({
    warehouses: [EXISTING_WAREHOUSE],
    pagination: { total: 1 },
  });
});

describe("WarehousesPage", () => {
  it("清單載入：以 pagination.total 推總筆數，且不送 sort/desc（proto 沒有這兩個參數）", async () => {
    await renderPage();
    expect(listWarehousesSpy).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: 20, keyword: "", includeDeleted: false })
    );
    expect(screen.getByText(/倉別即商品的進貨／揀貨倉來源（共 1 筆）/)).toBeTruthy();
    const sent = listWarehousesSpy.mock.calls[0][0] as Record<string, unknown>;
    expect("sort" in sent).toBe(false);
    expect("desc" in sent).toBe(false);
  });

  it("清單列渲染代號、名稱、地址與啟用徽章", async () => {
    await renderPage();
    expect(screen.getByText("WH1")).toBeTruthy();
    expect(screen.getByText("主倉")).toBeTruthy();
    expect(screen.getByText("台北市")).toBeTruthy();
    const cells = screen.getAllByRole("cell").map((c) => c.textContent ?? "");
    expect(cells).toContain("啟用");
  });

  it("停用列顯示停用徽章（停用＝僅標記，列仍在清單中）", async () => {
    listWarehousesSpy.mockResolvedValue({
      warehouses: [{ ...EXISTING_WAREHOUSE, isActive: false }],
      pagination: { total: 1 },
    });
    mountPage();
    await waitFor(() => expect(screen.getByText("WH1")).toBeTruthy());
    // 收斂到儲存格：表頭也叫「啟用」，整頁比對會打到表頭（表頭是 columnheader，不是 cell）。
    const cells = screen.getAllByRole("cell").map((c) => c.textContent ?? "");
    expect(cells).toContain("停用");
    expect(cells.join("|")).not.toContain("啟用");
  });

  it("送出關鍵字才進 query key（草稿狀態不得觸發查詢，送出後帶 keyword 與 page 1）", async () => {
    await renderPage();
    listWarehousesSpy.mockClear();
    fireEvent.input(screen.getByLabelText("關鍵字"), { target: { value: "WH1" } });
    // 沒送出前不得重查（D5：不得每按一鍵就查詢）。
    await Promise.resolve();
    expect(listWarehousesSpy).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listWarehousesSpy).toHaveBeenCalledWith(
        expect.objectContaining({ keyword: "WH1", page: 1 })
      )
    );
  });

  it("含已刪除切換會換 query key", async () => {
    await renderPage();
    listWarehousesSpy.mockClear();
    fireEvent.click(screen.getByLabelText("含已刪除"));
    await waitFor(() =>
      expect(listWarehousesSpy).toHaveBeenCalledWith(
        expect.objectContaining({ includeDeleted: true })
      )
    );
  });

  it("名稱空白時不打建立 API", async () => {
    await renderPage();
    createWarehouseSpy.mockResolvedValue({ warehouse: EXISTING_WAREHOUSE });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("倉別代號 *"), { target: { value: "WH2" } });
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(screen.getByText("請輸入倉別名稱")).toBeTruthy());
    expect(createWarehouseSpy).not.toHaveBeenCalled();
  });

  it("建立成功送出整份欄位（地址一併送，isActive 布林）", async () => {
    await renderPage();
    createWarehouseSpy.mockResolvedValue({ warehouse: EXISTING_WAREHOUSE });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("倉別代號 *"), { target: { value: "WH2" } });
    fireEvent.input(within(dialog).getByLabelText("名稱 *"), { target: { value: "副倉" } });
    fireEvent.input(within(dialog).getByLabelText("地址（選填）"), {
      target: { value: "新北市" },
    });
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(createWarehouseSpy).toHaveBeenCalledOnce());
    expect(createWarehouseSpy).toHaveBeenCalledWith({
      code: "WH2",
      name: "副倉",
      address: "新北市",
      isActive: true,
    });
    // 成功後關閉對話框並失效 warehouses 前綴。
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });

  it("編輯帶出既有值並走 updateWarehouse（欄位式 presence：整份送出）", async () => {
    await renderPage();
    updateWarehouseSpy.mockResolvedValue({ warehouse: EXISTING_WAREHOUSE });
    fireEvent.click(screen.getByRole("button", { name: "編輯" }));
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    await waitFor(() =>
      expect((within(dialog).getByLabelText("倉別代號 *") as HTMLInputElement).value).toBe("WH1")
    );
    expect((within(dialog).getByLabelText("名稱 *") as HTMLInputElement).value).toBe("主倉");
    expect((within(dialog).getByLabelText("地址（選填）") as HTMLInputElement).value).toBe("台北市");

    fireEvent.input(within(dialog).getByLabelText("名稱 *"), {
      target: { value: "主倉（改）" },
    });
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(updateWarehouseSpy).toHaveBeenCalledOnce());
    expect(updateWarehouseSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "w-1",
        code: "WH1",
        name: "主倉（改）",
        address: "台北市",
        isActive: true,
      })
    );
  });

  it("刪除先確認：拒絕不打 API，接受才呼叫 deleteWarehouse({id})", async () => {
    await renderPage();
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(false);
    deleteWarehouseSpy.mockResolvedValue({});
    fireEvent.click(screen.getByRole("button", { name: "刪除" }));
    await Promise.resolve();
    // 未確認 → 不得呼叫 API。
    expect(deleteWarehouseSpy).not.toHaveBeenCalled();

    confirmSpy.mockReturnValue(true);
    fireEvent.click(screen.getByRole("button", { name: "刪除" }));
    await waitFor(() => expect(deleteWarehouseSpy).toHaveBeenCalledWith({ id: "w-1" }));
    confirmSpy.mockRestore();
  });

  it("已刪除的倉別顯示還原而非刪除，還原呼叫 restoreWarehouse({id})", async () => {
    listWarehousesSpy.mockResolvedValue({
      warehouses: [DELETED_WAREHOUSE],
      pagination: { total: 1 },
    });
    mountPage();
    await waitFor(() => expect(screen.getByText("WH9")).toBeTruthy());
    expect(screen.queryByRole("button", { name: "刪除" })).toBeNull();
    restoreWarehouseSpy.mockResolvedValue({ warehouse: DELETED_WAREHOUSE });
    fireEvent.click(screen.getByRole("button", { name: "還原" }));
    await waitFor(() => expect(restoreWarehouseSpy).toHaveBeenCalledWith({ id: "w-2" }));
  });

  it("無權限時顯示「沒有權限執行此操作」", async () => {
    listWarehousesSpy.mockRejectedValue(new ConnectError("denied", Code.PermissionDenied));
    mountPage();
    await waitFor(() => expect(screen.getByRole("alert")).toBeTruthy());
    expect(screen.getByRole("alert").textContent).toContain("沒有權限執行此操作");
  });
});
