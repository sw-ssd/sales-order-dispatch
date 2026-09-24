import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 商品／分類／倉別／字典 API 以 spy 取代：頁面在模組層建立 connect client，
// 以 createClient 的替身同時攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const {
  listProductsSpy,
  getProductSpy,
  createProductSpy,
  updateProductSpy,
  deleteProductSpy,
  restoreProductSpy,
  listProductCategoriesSpy,
  listWarehousesSpy,
  listOptionsSpy,
  listProcessingSpecsSpy,
} = vi.hoisted(() => ({
  listProductsSpy: vi.fn(),
  getProductSpy: vi.fn(),
  createProductSpy: vi.fn(),
  updateProductSpy: vi.fn(),
  deleteProductSpy: vi.fn(),
  restoreProductSpy: vi.fn(),
  listProductCategoriesSpy: vi.fn(),
  listWarehousesSpy: vi.fn(),
  listOptionsSpy: vi.fn(),
  listProcessingSpecsSpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listProducts: listProductsSpy,
    getProduct: getProductSpy,
    createProduct: createProductSpy,
    updateProduct: updateProductSpy,
    deleteProduct: deleteProductSpy,
    restoreProduct: restoreProductSpy,
    listProductCategories: listProductCategoriesSpy,
    listWarehouses: listWarehousesSpy,
    listOptions: listOptionsSpy,
    listProcessingSpecs: listProcessingSpecsSpy,
  }),
}));

import ProductsPage from "./ProductsPage";

const EXISTING_PRODUCT = {
  id: "p-1",
  companyId: "1",
  departmentId: "1",
  code: "AP-001",
  name: "蘋果",
  categoryId: "cat-1",
  inventoryWarehouseId: "",
  pickingWarehouseId: "",
  description: "進口",
  isActive: true,
  units: [
    {
      unitCode: "KG",
      conversionRate: "1",
      isBase: true,
      sortOrder: 0,
      sizeDesc: "",
    },
    {
      unitCode: "BOX",
      conversionRate: "5",
      isBase: false,
      sortOrder: 1,
      sizeDesc: "1盒=5kg",
    },
  ],
  processingSpecs: [
    { processingSpecId: "spec-1", attributes: { cutWidth: 3, note: "去骨" } },
    { processingSpecId: "spec-2" },
  ],
  createdAt: "2026-09-22T10:00:00Z",
  updatedAt: "2026-09-22T10:00:00Z",
  deletedAt: "",
};

/** 部門分切規格主檔（編輯器的選項來源）；spec-1/spec-2 已被 EXISTING_PRODUCT 關聯。 */
const PROCESSING_SPEC_OPTIONS = [
  { id: "spec-1", code: "SP1", name: "切塊", isActive: true, deletedAt: "" },
  { id: "spec-2", code: "SP2", name: "去骨", isActive: true, deletedAt: "" },
  { id: "spec-3", code: "SP3", name: "切片", isActive: true, deletedAt: "" },
];

const DELETED_PRODUCT = {
  ...EXISTING_PRODUCT,
  id: "p-2",
  code: "OLD-001",
  name: "已刪商品",
  deletedAt: "2026-09-01T00:00:00Z",
};

/** 每個測試一份全新的 `QueryClient`：快取不跨測試殘留（retry 關閉，理由同 CustomersPage）。 */
function newClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function mountPage(client: QueryClient = newClient()) {
  render(() => (
    <QueryClientProvider client={client}>
      <ProductsPage />
    </QueryClientProvider>
  ));
}

async function renderPage() {
  mountPage();
  await waitFor(() => expect(screen.getByText("AP-001")).toBeTruthy());
}

/** 開建立對話框並回傳它的 DOM（欄位查詢收斂到這個 dialog）。 */
async function openCreateDialog(): Promise<HTMLElement> {
  fireEvent.click(screen.getByRole("button", { name: "新增商品" }));
  return await waitFor(() => screen.getByRole("dialog"));
}

beforeEach(() => {
  vi.clearAllMocks();
  listProductsSpy.mockResolvedValue({
    products: [EXISTING_PRODUCT],
    pagination: { total: 1 },
  });
  getProductSpy.mockResolvedValue({ product: EXISTING_PRODUCT });
  listProductCategoriesSpy.mockResolvedValue({
    productCategories: [{ id: "cat-1", name: "水果", code: "F", isActive: true, sortOrder: 1 }],
    pagination: { total: 1 },
  });
  listWarehousesSpy.mockResolvedValue({
    warehouses: [{ id: "w-1", name: "主倉", code: "WH1", isActive: true, address: "" }],
    pagination: { total: 1 },
  });
  listOptionsSpy.mockResolvedValue({
    options: [
      { code: "KG", displayName: "公斤" },
      { code: "G", displayName: "公克" },
      { code: "BOX", displayName: "盒" },
    ],
  });
  listProcessingSpecsSpy.mockResolvedValue({
    processingSpecs: PROCESSING_SPEC_OPTIONS,
    pagination: { total: PROCESSING_SPEC_OPTIONS.length },
  });
});

describe("ProductsPage", () => {
  it("清單載入：以 pagination.total 推總筆數，且不送 sort/desc（proto 沒有這兩個參數）", async () => {
    await renderPage();
    expect(listProductsSpy).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: 20, keyword: "", categoryId: "", includeDeleted: false })
    );
    expect(screen.getByText(/商品主檔\(共 1 筆\)/)).toBeTruthy();
    const sent = listProductsSpy.mock.calls[0][0] as Record<string, unknown>;
    expect("sort" in sent).toBe(false);
    expect("desc" in sent).toBe(false);
  });

  it("分類欄由分類下拉解析名稱（收斂到表格，篩選下拉也有同名選項）", async () => {
    await renderPage();
    const table = screen.getByRole("table");
    expect(await within(table).findByText("水果")).toBeTruthy();
  });

  it("送出篩選才進 query key", async () => {
    await renderPage();
    listProductsSpy.mockClear();
    fireEvent.input(screen.getByLabelText("關鍵字"), { target: { value: "AP" } });
    await Promise.resolve();
    expect(listProductsSpy).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listProductsSpy).toHaveBeenCalledWith(expect.objectContaining({ keyword: "AP" }))
    );
  });

  it("含已刪除切換會換 query key", async () => {
    await renderPage();
    listProductsSpy.mockClear();
    fireEvent.click(screen.getByLabelText("含已刪除"));
    await waitFor(() =>
      expect(listProductsSpy).toHaveBeenCalledWith(expect.objectContaining({ includeDeleted: true }))
    );
  });

  it("未選單位代碼不打建單 API，且錯誤顯示在單位區", async () => {
    await renderPage();
    createProductSpy.mockResolvedValue({ product: EXISTING_PRODUCT });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("商品代號 *"), {
      target: { value: "AP-002" },
    });
    fireEvent.input(within(dialog).getByLabelText("商品名稱 *"), {
      target: { value: "香蕉" },
    });
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(screen.getByText("單位代碼不可為空")).toBeTruthy());
    expect(createProductSpy).not.toHaveBeenCalled();
  });

  it("名稱空白時不打建單 API", async () => {
    await renderPage();
    createProductSpy.mockResolvedValue({ product: EXISTING_PRODUCT });
    const dialog = await openCreateDialog();
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(screen.getByText("請輸入商品名稱")).toBeTruthy());
    expect(createProductSpy).not.toHaveBeenCalled();
  });

  it("勾第二個單位為基本單位：其餘自動降為換算單位，且換算率恆為 1", async () => {
    await renderPage();
    const dialog = await openCreateDialog();
    // 再加一列 → 兩列；把第二列設為基本單位。
    fireEvent.click(within(dialog).getByRole("button", { name: "加單位" }));
    const baseRadios = within(dialog).getAllByRole("radio", { name: /基本單位/ });
    expect(baseRadios).toHaveLength(2);
    fireEvent.click(baseRadios[1]);
    // 第二列成為基本單位（率鎖 1、欄位停用），第一列降為換算單位（率可編輯）。
    const rates = within(dialog).getAllByLabelText("換算率 *") as HTMLInputElement[];
    expect(rates[1].value).toBe("1");
    expect(rates[1].disabled).toBe(true);
    expect(rates[0].disabled).toBe(false);
  });

  it("單位齊備才建單，並組成 ProductUnit payload", async () => {
    await renderPage();
    createProductSpy.mockResolvedValue({ product: EXISTING_PRODUCT });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("商品代號 *"), { target: { value: "AP-002" } });
    fireEvent.input(within(dialog).getByLabelText("商品名稱 *"), { target: { value: "香蕉" } });
    const codeSelect = within(dialog).getAllByLabelText("單位代碼 *")[0] as HTMLSelectElement;
    fireEvent.change(codeSelect, { target: { value: "KG" } });
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(createProductSpy).toHaveBeenCalledOnce());
    const sent = createProductSpy.mock.calls[0][0] as {
      code: string;
      name: string;
      isActive: boolean;
      categoryId: string;
      inventoryWarehouseId: string;
      units: { unitCode: string; conversionRate: string; isBase: boolean; sortOrder: number }[];
    };
    expect(sent.code).toBe("AP-002");
    expect(sent.name).toBe("香蕉");
    expect(sent.isActive).toBe(true);
    expect(sent.categoryId).toBe("");
    expect(sent.inventoryWarehouseId).toBe("");
    expect(sent.units).toHaveLength(1);
    expect(sent.units[0]).toMatchObject({
      unitCode: "KG",
      conversionRate: "1",
      isBase: true,
      sortOrder: 0,
    });
  });

  it("編輯帶出既有值並走 updateProduct（單位整組替換）", async () => {
    await renderPage();
    updateProductSpy.mockResolvedValue({ product: EXISTING_PRODUCT });
    fireEvent.click(screen.getAllByRole("button", { name: "編輯" })[0]);
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    await waitFor(() =>
      expect((within(dialog).getByLabelText("商品代號 *") as HTMLInputElement).value).toBe("AP-001")
    );
    const codeSelects = within(dialog).getAllByLabelText("單位代碼 *") as HTMLSelectElement[];
    // 既有商品帶兩個單位 → 編輯器應原樣帶入兩列。
    expect(codeSelects).toHaveLength(2);
    // 改第二列（KG 基本單位留在第一列）；改成 BOX 會與既有 BOX 衝突，故改 G。
    fireEvent.change(codeSelects[1], { target: { value: "G" } });
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(updateProductSpy).toHaveBeenCalledOnce());
    const sent = updateProductSpy.mock.calls[0][0] as {
      id: string;
      units: { unitCode: string; conversionRate: string; isBase: boolean }[];
    };
    expect(sent.id).toBe("p-1");
    expect(sent.units).toHaveLength(2);
    // 第一列維持既有基本單位（率恆 1），第二列換成新單位且非基本。
    expect(sent.units[0]).toMatchObject({ unitCode: "KG", conversionRate: "1", isBase: true });
    expect(sent.units[1]).toMatchObject({ unitCode: "G", isBase: false });
  });

  it("編輯分切規格：增減關聯後，仍被選取之規格的 attributes 原樣帶出", async () => {
    await renderPage();
    updateProductSpy.mockResolvedValue({ product: EXISTING_PRODUCT });
    fireEvent.click(screen.getAllByRole("button", { name: "編輯" })[0]);
    const dialog = await waitFor(() => screen.getByRole("dialog"));

    // 既有兩筆關聯（spec-1 帶 attributes、spec-2 不帶）都應勾選。
    // 無障礙名稱由 label 內文組成，`SP1` 與名稱之間會被插入空白 → 用 regex 比對。
    const spec1 = (await within(dialog).findByRole("checkbox", {
      name: /SP1/,
    })) as HTMLInputElement;
    const spec2 = within(dialog).getByRole("checkbox", { name: /SP2/ }) as HTMLInputElement;
    await waitFor(() => expect(spec1.checked && spec2.checked).toBe(true));

    // 減一筆、加一筆：spec-1 留著（attributes 必須原樣），spec-2 取消，spec-3 新增。
    fireEvent.click(spec2);
    fireEvent.click(within(dialog).getByRole("checkbox", { name: /SP3/ }));
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(updateProductSpy).toHaveBeenCalledOnce());
    const sent = updateProductSpy.mock.calls[0][0] as {
      processingSpecs: { processingSpecId: string; attributes?: Record<string, unknown> }[];
    };
    // 整組替換語意：欄位必須「有送」，否則使用者的取消選取不會生效。
    expect(Array.isArray(sent.processingSpecs)).toBe(true);
    expect(sent.processingSpecs).toHaveLength(2);
    // 核心風險點：保留項的配對層覆寫指令原樣帶回，新選項給空物件。
    expect(sent.processingSpecs[0]).toEqual({
      processingSpecId: "spec-1",
      attributes: { cutWidth: 3, note: "去骨" },
    });
    expect(sent.processingSpecs[1]).toEqual({ processingSpecId: "spec-3", attributes: {} });
  });

  it("編輯時全部取消規格 → 帶 clearProcessingSpecs 旗標（空陣列清不掉）", async () => {
    await renderPage();
    updateProductSpy.mockResolvedValue({ product: EXISTING_PRODUCT });
    fireEvent.click(screen.getAllByRole("button", { name: "編輯" })[0]);
    const dialog = await waitFor(() => screen.getByRole("dialog"));

    const spec1 = (await within(dialog).findByRole("checkbox", { name: /SP1/ })) as HTMLInputElement;
    const spec2 = within(dialog).getByRole("checkbox", { name: /SP2/ }) as HTMLInputElement;
    await waitFor(() => expect(spec1.checked && spec2.checked).toBe(true));

    // 兩筆都取消 → 走旗標。proto3 repeated 無 presence，空陣列與「未提供」同義，
    // 只送 [] 不會清掉既有關聯（後端維持現值），故必須帶旗標。
    fireEvent.click(spec1);
    fireEvent.click(spec2);
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(updateProductSpy).toHaveBeenCalledOnce());
    const sent = updateProductSpy.mock.calls[0][0] as {
      clearProcessingSpecs?: boolean;
      processingSpecs?: unknown;
    };
    expect(sent.clearProcessingSpecs).toBe(true);
    // 兩者互斥：帶旗標時不得同時帶非空陣列（後端會回 invalid_argument）。
    expect(sent.processingSpecs ?? []).toHaveLength(0);
  });

  it("新增商品時未選規格 → 送空陣列（＝不關聯）", async () => {
    await renderPage();
    createProductSpy.mockResolvedValue({ product: EXISTING_PRODUCT });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("商品代號 *"), { target: { value: "AP-002" } });
    fireEvent.input(within(dialog).getByLabelText("商品名稱 *"), { target: { value: "香蕉" } });
    fireEvent.change(within(dialog).getAllByLabelText("單位代碼 *")[0], { target: { value: "KG" } });
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(createProductSpy).toHaveBeenCalledOnce());
    const sent = createProductSpy.mock.calls[0][0] as { processingSpecs: unknown[] };
    expect(sent.processingSpecs).toEqual([]);
  });

  it("刪除先確認再呼叫並失效 products 前綴", async () => {
    await renderPage();
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(false);
    deleteProductSpy.mockResolvedValue({});
    fireEvent.click(screen.getAllByRole("button", { name: "刪除" })[0]);
    await Promise.resolve();
    expect(deleteProductSpy).not.toHaveBeenCalled();
    confirmSpy.mockRestore();

    vi.spyOn(window, "confirm").mockReturnValue(true);
    listProductsSpy.mockClear();
    fireEvent.click(screen.getAllByRole("button", { name: "刪除" })[0]);
    await waitFor(() => expect(deleteProductSpy).toHaveBeenCalledOnce());
    await waitFor(() => expect(listProductsSpy).toHaveBeenCalled());
    (window.confirm as unknown as { mockRestore: () => void }).mockRestore();
  });

  it("已刪除的商品顯示還原而非刪除", async () => {
    listProductsSpy.mockResolvedValue({ products: [DELETED_PRODUCT], pagination: { total: 1 } });
    mountPage();
    await waitFor(() => expect(screen.getByText("OLD-001")).toBeTruthy());
    expect(screen.queryByRole("button", { name: "刪除" })).toBeNull();
    restoreProductSpy.mockResolvedValue({ product: DELETED_PRODUCT });
    fireEvent.click(screen.getByRole("button", { name: "還原" }));
    await waitFor(() => expect(restoreProductSpy).toHaveBeenCalledOnce());
  });

  it("無權限時顯示 banner", async () => {
    listProductsSpy.mockRejectedValue(new ConnectError("denied", Code.PermissionDenied));
    mountPage();
    await waitFor(() => expect(screen.getByRole("alert")).toBeTruthy());
    expect(screen.getByRole("alert").textContent).toContain("沒有權限");
  });
});
