import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 規格 API 以 spy 取代：頁面在模組層建立 connect client，
// 以 createClient 的替身同時攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const {
  listProcessingSpecsSpy,
  createProcessingSpecSpy,
  updateProcessingSpecSpy,
  deleteProcessingSpecSpy,
  restoreProcessingSpecSpy,
} = vi.hoisted(() => ({
  listProcessingSpecsSpy: vi.fn(),
  createProcessingSpecSpy: vi.fn(),
  updateProcessingSpecSpy: vi.fn(),
  deleteProcessingSpecSpy: vi.fn(),
  restoreProcessingSpecSpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listProcessingSpecs: listProcessingSpecsSpy,
    createProcessingSpec: createProcessingSpecSpy,
    updateProcessingSpec: updateProcessingSpecSpy,
    deleteProcessingSpec: deleteProcessingSpecSpy,
    restoreProcessingSpec: restoreProcessingSpecSpy,
  }),
}));

import ProcessingSpecsPage from "./ProcessingSpecsPage";

const EXISTING_SPEC = {
  id: "s-1",
  companyId: "co-1",
  departmentId: "d-1",
  code: "S1",
  name: "五分切",
  kind: "cutting",
  appliesToProcessing: true,
  appliesToPicking: true,
  attributes: { depth: 3 },
  sortOrder: 1,
  isActive: true,
  createdAt: "2026-09-22T10:00:00Z",
  updatedAt: "2026-09-22T10:00:00Z",
  deletedAt: "",
};

const DELETED_SPEC = {
  ...EXISTING_SPEC,
  id: "s-2",
  code: "S9",
  name: "已刪規格",
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
      <ProcessingSpecsPage />
    </QueryClientProvider>
  ));
  return client;
}

/** 標準情境：以預設 mock（一列 S1）掛頁面並等清單落地。 */
async function renderPage() {
  mountPage();
  await waitFor(() => expect(screen.getByText("S1")).toBeTruthy());
}

/** 開新增對話框並回傳它的 DOM（欄位查詢收斂到這個 dialog）。 */
async function openCreateDialog(): Promise<HTMLElement> {
  fireEvent.click(screen.getByRole("button", { name: "新增規格" }));
  return await waitFor(() => screen.getByRole("dialog"));
}

beforeEach(() => {
  vi.clearAllMocks();
  listProcessingSpecsSpy.mockResolvedValue({
    processingSpecs: [EXISTING_SPEC],
    pagination: { total: 1 },
  });
});

describe("ProcessingSpecsPage", () => {
  it("清單載入：以 pagination.total 推總筆數，且不送 sort/desc（proto 沒有這兩個參數）", async () => {
    await renderPage();
    expect(listProcessingSpecsSpy).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: 20, keyword: "", includeDeleted: false })
    );
    expect(screen.getByText(/分切規格即商品可勾選的加工／揀貨條件（共 1 筆）/)).toBeTruthy();
    const sent = listProcessingSpecsSpy.mock.calls[0][0] as Record<string, unknown>;
    expect("sort" in sent).toBe(false);
    expect("desc" in sent).toBe(false);
  });

  it("清單列渲染代號、名稱、類型與啟用徽章", async () => {
    await renderPage();
    expect(screen.getByText("S1")).toBeTruthy();
    expect(screen.getByText("五分切")).toBeTruthy();
    expect(screen.getByText("cutting")).toBeTruthy();
    const cells = screen.getAllByRole("cell").map((c) => c.textContent ?? "");
    expect(cells).toContain("啟用");
  });

  it("停用列顯示停用徽章（停用＝僅標記，列仍在清單中）", async () => {
    listProcessingSpecsSpy.mockResolvedValue({
      processingSpecs: [{ ...EXISTING_SPEC, isActive: false }],
      pagination: { total: 1 },
    });
    mountPage();
    await waitFor(() => expect(screen.getByText("S1")).toBeTruthy());
    // 收斂到儲存格：表頭也叫「啟用」，整頁比對會打到表頭（表頭是 columnheader，不是 cell）。
    const cells = screen.getAllByRole("cell").map((c) => c.textContent ?? "");
    expect(cells).toContain("停用");
    expect(cells.join("|")).not.toContain("啟用");
  });

  it("送出關鍵字才進 query key（草稿狀態不得觸發查詢，送出後帶 keyword 與 page 1）", async () => {
    await renderPage();
    listProcessingSpecsSpy.mockClear();
    fireEvent.input(screen.getByLabelText("關鍵字"), { target: { value: "五分" } });
    // 沒送出前不得重查（D5：不得每按一鍵就查詢）。
    await Promise.resolve();
    expect(listProcessingSpecsSpy).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listProcessingSpecsSpy).toHaveBeenCalledWith(
        expect.objectContaining({ keyword: "五分", page: 1 })
      )
    );
  });

  it("含已刪除切換會換 query key", async () => {
    await renderPage();
    listProcessingSpecsSpy.mockClear();
    fireEvent.click(screen.getByLabelText("含已刪除"));
    await waitFor(() =>
      expect(listProcessingSpecsSpy).toHaveBeenCalledWith(
        expect.objectContaining({ includeDeleted: true })
      )
    );
  });

  it("名稱空白時不打建立 API", async () => {
    await renderPage();
    createProcessingSpecSpy.mockResolvedValue({ processingSpec: EXISTING_SPEC });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("規格代號 *"), { target: { value: "S2" } });
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(screen.getByText("請輸入規格名稱")).toBeTruthy());
    expect(createProcessingSpecSpy).not.toHaveBeenCalled();
  });

  it("建立成功送出整份欄位（布林旗標、字串排序轉 int32、空白 attributes 送空 Struct）", async () => {
    await renderPage();
    createProcessingSpecSpy.mockResolvedValue({ processingSpec: EXISTING_SPEC });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("規格代號 *"), { target: { value: "S2" } });
    fireEvent.input(within(dialog).getByLabelText("名稱 *"), { target: { value: "十分切" } });
    fireEvent.input(within(dialog).getByLabelText("類型"), { target: { value: "pack" } });
    fireEvent.input(within(dialog).getByLabelText("排序"), { target: { value: "2" } });
    // 取消其中一個旗標：布林欄位要如實反映勾選狀態（後端至少其一 true 的規則另由 banner 呈現）。
    fireEvent.click(within(dialog).getByLabelText("配送揀"));
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(createProcessingSpecSpy).toHaveBeenCalledOnce());
    expect(createProcessingSpecSpy).toHaveBeenCalledWith({
      code: "S2",
      name: "十分切",
      kind: "pack",
      appliesToProcessing: true,
      appliesToPicking: false,
      attributes: {},
      sortOrder: 2,
      isActive: true,
    });
    // 成功後關閉對話框並失效 processingSpecs 前綴。
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });

  it("屬性不是合法 JSON 物件時標欄位錯誤且不打 API", async () => {
    await renderPage();
    createProcessingSpecSpy.mockResolvedValue({ processingSpec: EXISTING_SPEC });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("規格代號 *"), { target: { value: "S2" } });
    fireEvent.input(within(dialog).getByLabelText("名稱 *"), { target: { value: "十分切" } });
    fireEvent.input(within(dialog).getByLabelText("屬性 JSON（選填，留空＝不帶）"), {
      target: { value: "{oops" },
    });
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() =>
      expect(screen.getByText("屬性須為 JSON 物件，或留空")).toBeTruthy()
    );
    expect(createProcessingSpecSpy).not.toHaveBeenCalled();
  });

  it("合法 JSON 屬性經 JSON.parse 成 Struct 送出（字串表單值不得原樣上送）", async () => {
    await renderPage();
    createProcessingSpecSpy.mockResolvedValue({ processingSpec: EXISTING_SPEC });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("規格代號 *"), { target: { value: "S3" } });
    fireEvent.input(within(dialog).getByLabelText("名稱 *"), { target: { value: "三分切" } });
    fireEvent.input(within(dialog).getByLabelText("屬性 JSON（選填，留空＝不帶）"), {
      target: { value: '{"depth": 3}' },
    });
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(createProcessingSpecSpy).toHaveBeenCalledOnce());
    expect(createProcessingSpecSpy).toHaveBeenCalledWith({
      code: "S3",
      name: "三分切",
      kind: "",
      appliesToProcessing: true,
      appliesToPicking: true,
      attributes: { depth: 3 },
      sortOrder: 0,
      isActive: true,
    });
  });

  it("編輯帶出既有值（attributes 還原成 JSON 字串）並整份送進 updateProcessingSpec", async () => {
    await renderPage();
    updateProcessingSpecSpy.mockResolvedValue({ processingSpec: EXISTING_SPEC });
    fireEvent.click(screen.getByRole("button", { name: "編輯" }));
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    await waitFor(() =>
      expect((within(dialog).getByLabelText("規格代號 *") as HTMLInputElement).value).toBe("S1")
    );
    expect((within(dialog).getByLabelText("名稱 *") as HTMLInputElement).value).toBe("五分切");
    expect((within(dialog).getByLabelText("類型") as HTMLInputElement).value).toBe("cutting");
    expect((within(dialog).getByLabelText("排序") as HTMLInputElement).value).toBe("1");
    expect(
      (within(dialog).getByLabelText("屬性 JSON（選填，留空＝不帶）") as HTMLTextAreaElement).value
    ).toBe('{"depth":3}');

    // 取消其中一個旗標：update 是欄位式 presence，兩個旗標仍整份送出（後端以最終值合併驗證）。
    fireEvent.click(within(dialog).getByLabelText("加工室揀"));
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(updateProcessingSpecSpy).toHaveBeenCalledOnce());
    expect(updateProcessingSpecSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "s-1",
        code: "S1",
        name: "五分切",
        kind: "cutting",
        appliesToProcessing: false,
        appliesToPicking: true,
        attributes: { depth: 3 },
        sortOrder: 1,
        isActive: true,
      })
    );
  });

  it("刪除先確認：拒絕不打 API，接受才呼叫 deleteProcessingSpec({id})", async () => {
    await renderPage();
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(false);
    deleteProcessingSpecSpy.mockResolvedValue({});
    fireEvent.click(screen.getByRole("button", { name: "刪除" }));
    await Promise.resolve();
    // 未確認 → 不得呼叫 API。
    expect(deleteProcessingSpecSpy).not.toHaveBeenCalled();

    confirmSpy.mockReturnValue(true);
    fireEvent.click(screen.getByRole("button", { name: "刪除" }));
    await waitFor(() => expect(deleteProcessingSpecSpy).toHaveBeenCalledWith({ id: "s-1" }));
    confirmSpy.mockRestore();
  });

  it("已刪除的規格顯示還原而非刪除，還原呼叫 restoreProcessingSpec({id})", async () => {
    listProcessingSpecsSpy.mockResolvedValue({
      processingSpecs: [DELETED_SPEC],
      pagination: { total: 1 },
    });
    mountPage();
    await waitFor(() => expect(screen.getByText("S9")).toBeTruthy());
    expect(screen.queryByRole("button", { name: "刪除" })).toBeNull();
    restoreProcessingSpecSpy.mockResolvedValue({ processingSpec: DELETED_SPEC });
    fireEvent.click(screen.getByRole("button", { name: "還原" }));
    await waitFor(() => expect(restoreProcessingSpecSpy).toHaveBeenCalledWith({ id: "s-2" }));
  });

  it("無權限時顯示「沒有權限執行此操作」", async () => {
    listProcessingSpecsSpy.mockRejectedValue(new ConnectError("denied", Code.PermissionDenied));
    mountPage();
    await waitFor(() => expect(screen.getByRole("alert")).toBeTruthy());
    expect(screen.getByRole("alert").textContent).toContain("沒有權限執行此操作");
  });
});
