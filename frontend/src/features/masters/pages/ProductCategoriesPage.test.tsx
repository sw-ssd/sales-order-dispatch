import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 分類 API 以 spy 取代：頁面在模組層建立 connect client，
// 以 createClient 的替身同時攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const {
  listProductCategoriesSpy,
  createProductCategorySpy,
  updateProductCategorySpy,
  deleteProductCategorySpy,
  restoreProductCategorySpy,
} = vi.hoisted(() => ({
  listProductCategoriesSpy: vi.fn(),
  createProductCategorySpy: vi.fn(),
  updateProductCategorySpy: vi.fn(),
  deleteProductCategorySpy: vi.fn(),
  restoreProductCategorySpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listProductCategories: listProductCategoriesSpy,
    createProductCategory: createProductCategorySpy,
    updateProductCategory: updateProductCategorySpy,
    deleteProductCategory: deleteProductCategorySpy,
    restoreProductCategory: restoreProductCategorySpy,
  }),
}));

import ProductCategoriesPage from "./ProductCategoriesPage";

const EXISTING_CATEGORY = {
  id: "c-1",
  companyId: "co-1",
  departmentId: "d-1",
  code: "C1",
  name: "冷藏類",
  sortOrder: 3,
  isActive: true,
  createdAt: "2026-09-22T10:00:00Z",
  updatedAt: "2026-09-22T10:00:00Z",
  deletedAt: "",
};

const DELETED_CATEGORY = {
  ...EXISTING_CATEGORY,
  id: "c-2",
  code: "C9",
  name: "已刪分類",
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
      <ProductCategoriesPage />
    </QueryClientProvider>
  ));
  return client;
}

/** 標準情境：以預設 mock（一列 C1）掛頁面並等清單落地。 */
async function renderPage() {
  mountPage();
  await waitFor(() => expect(screen.getByText("C1")).toBeTruthy());
}

/** 開新增對話框並回傳它的 DOM（欄位查詢收斂到這個 dialog）。 */
async function openCreateDialog(): Promise<HTMLElement> {
  fireEvent.click(screen.getByRole("button", { name: "新增分類" }));
  return await waitFor(() => screen.getByRole("dialog"));
}

beforeEach(() => {
  vi.clearAllMocks();
  listProductCategoriesSpy.mockResolvedValue({
    productCategories: [EXISTING_CATEGORY],
    pagination: { total: 1 },
  });
});

describe("ProductCategoriesPage", () => {
  it("清單載入：以 pagination.total 推總筆數，且不送 sort/desc（proto 沒有這兩個參數）", async () => {
    await renderPage();
    expect(listProductCategoriesSpy).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: 20, keyword: "", includeDeleted: false })
    );
    expect(screen.getByText(/商品分類即商品主檔的所屬分類（共 1 筆）/)).toBeTruthy();
    const sent = listProductCategoriesSpy.mock.calls[0][0] as Record<string, unknown>;
    expect("sort" in sent).toBe(false);
    expect("desc" in sent).toBe(false);
  });

  it("清單列渲染代號、名稱、排序與啟用徽章", async () => {
    await renderPage();
    expect(screen.getByText("C1")).toBeTruthy();
    expect(screen.getByText("冷藏類")).toBeTruthy();
    expect(screen.getByText("3")).toBeTruthy();
    const cells = screen.getAllByRole("cell").map((c) => c.textContent ?? "");
    expect(cells).toContain("啟用");
  });

  it("停用列顯示停用徽章（停用＝僅標記，列仍在清單中）", async () => {
    listProductCategoriesSpy.mockResolvedValue({
      productCategories: [{ ...EXISTING_CATEGORY, isActive: false }],
      pagination: { total: 1 },
    });
    mountPage();
    await waitFor(() => expect(screen.getByText("C1")).toBeTruthy());
    // 收斂到儲存格：表頭也叫「啟用」，整頁比對會打到表頭（表頭是 columnheader，不是 cell）。
    const cells = screen.getAllByRole("cell").map((c) => c.textContent ?? "");
    expect(cells).toContain("停用");
    expect(cells.join("|")).not.toContain("啟用");
  });

  it("送出關鍵字才進 query key（草稿狀態不得觸發查詢，送出後帶 keyword 與 page 1）", async () => {
    await renderPage();
    listProductCategoriesSpy.mockClear();
    fireEvent.input(screen.getByLabelText("關鍵字"), { target: { value: "冷藏" } });
    // 沒送出前不得重查（D5：不得每按一鍵就查詢）。
    await Promise.resolve();
    expect(listProductCategoriesSpy).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listProductCategoriesSpy).toHaveBeenCalledWith(
        expect.objectContaining({ keyword: "冷藏", page: 1 })
      )
    );
  });

  it("含已刪除切換會換 query key", async () => {
    await renderPage();
    listProductCategoriesSpy.mockClear();
    fireEvent.click(screen.getByLabelText("含已刪除"));
    await waitFor(() =>
      expect(listProductCategoriesSpy).toHaveBeenCalledWith(
        expect.objectContaining({ includeDeleted: true })
      )
    );
  });

  it("名稱空白時不打建立 API", async () => {
    await renderPage();
    createProductCategorySpy.mockResolvedValue({ productCategory: EXISTING_CATEGORY });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("分類代號 *"), { target: { value: "C2" } });
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(screen.getByText("請輸入分類名稱")).toBeTruthy());
    expect(createProductCategorySpy).not.toHaveBeenCalled();
  });

  it("建立成功送出整份欄位（字串排序經 toSortOrder 轉 int32）", async () => {
    await renderPage();
    createProductCategorySpy.mockResolvedValue({ productCategory: EXISTING_CATEGORY });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("分類代號 *"), { target: { value: "C2" } });
    fireEvent.input(within(dialog).getByLabelText("名稱 *"), { target: { value: "冷凍類" } });
    fireEvent.input(within(dialog).getByLabelText("排序"), { target: { value: "5" } });
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(createProductCategorySpy).toHaveBeenCalledOnce());
    expect(createProductCategorySpy).toHaveBeenCalledWith({
      code: "C2",
      name: "冷凍類",
      sortOrder: 5,
      isActive: true,
    });
    // 成功後關閉對話框並失效 productCategories 前綴。
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });

  it("編輯帶出既有值並走 updateProductCategory（欄位式 presence：整份送出）", async () => {
    await renderPage();
    updateProductCategorySpy.mockResolvedValue({ productCategory: EXISTING_CATEGORY });
    fireEvent.click(screen.getByRole("button", { name: "編輯" }));
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    await waitFor(() =>
      expect((within(dialog).getByLabelText("分類代號 *") as HTMLInputElement).value).toBe("C1")
    );
    expect((within(dialog).getByLabelText("名稱 *") as HTMLInputElement).value).toBe("冷藏類");
    expect((within(dialog).getByLabelText("排序") as HTMLInputElement).value).toBe("3");

    fireEvent.input(within(dialog).getByLabelText("名稱 *"), {
      target: { value: "冷藏類（改）" },
    });
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(updateProductCategorySpy).toHaveBeenCalledOnce());
    expect(updateProductCategorySpy).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "c-1",
        code: "C1",
        name: "冷藏類（改）",
        sortOrder: 3,
        isActive: true,
      })
    );
  });

  it("刪除先確認：拒絕不打 API，接受才呼叫 deleteProductCategory({id})", async () => {
    await renderPage();
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(false);
    deleteProductCategorySpy.mockResolvedValue({});
    fireEvent.click(screen.getByRole("button", { name: "刪除" }));
    await Promise.resolve();
    // 未確認 → 不得呼叫 API。
    expect(deleteProductCategorySpy).not.toHaveBeenCalled();

    confirmSpy.mockReturnValue(true);
    fireEvent.click(screen.getByRole("button", { name: "刪除" }));
    await waitFor(() => expect(deleteProductCategorySpy).toHaveBeenCalledWith({ id: "c-1" }));
    confirmSpy.mockRestore();
  });

  it("已刪除的分類顯示還原而非刪除，還原呼叫 restoreProductCategory({id})", async () => {
    listProductCategoriesSpy.mockResolvedValue({
      productCategories: [DELETED_CATEGORY],
      pagination: { total: 1 },
    });
    mountPage();
    await waitFor(() => expect(screen.getByText("C9")).toBeTruthy());
    expect(screen.queryByRole("button", { name: "刪除" })).toBeNull();
    restoreProductCategorySpy.mockResolvedValue({ productCategory: DELETED_CATEGORY });
    fireEvent.click(screen.getByRole("button", { name: "還原" }));
    await waitFor(() => expect(restoreProductCategorySpy).toHaveBeenCalledWith({ id: "c-2" }));
  });

  it("無權限時顯示「沒有權限執行此操作」", async () => {
    listProductCategoriesSpy.mockRejectedValue(
      new ConnectError("denied", Code.PermissionDenied)
    );
    mountPage();
    await waitFor(() => expect(screen.getByRole("alert")).toBeTruthy());
    expect(screen.getByRole("alert").textContent).toContain("沒有權限執行此操作");
  });
});
