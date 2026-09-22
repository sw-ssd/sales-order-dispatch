import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 車次 API 以 spy 取代：頁面在模組層建立 connect client，
// 以 createClient 的替身同時攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const {
  listRoutesSpy,
  createRouteSpy,
  updateRouteSpy,
  deleteRouteSpy,
  restoreRouteSpy,
} = vi.hoisted(() => ({
  listRoutesSpy: vi.fn(),
  createRouteSpy: vi.fn(),
  updateRouteSpy: vi.fn(),
  deleteRouteSpy: vi.fn(),
  restoreRouteSpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listRoutes: listRoutesSpy,
    createRoute: createRouteSpy,
    updateRoute: updateRouteSpy,
    deleteRoute: deleteRouteSpy,
    restoreRoute: restoreRouteSpy,
  }),
}));

import RoutesPage from "./RoutesPage";

const EXISTING_ROUTE = {
  id: "r-1",
  companyId: "c-1",
  departmentId: "d-1",
  code: "R1",
  name: "一號路線",
  description: "市區",
  sortOrder: 1,
  isActive: true,
  createdAt: "2026-09-22T10:00:00Z",
  updatedAt: "2026-09-22T10:00:00Z",
  deletedAt: "",
};

const DELETED_ROUTE = {
  ...EXISTING_ROUTE,
  id: "r-2",
  code: "R9",
  name: "已刪路線",
  deletedAt: "2026-09-01T00:00:00Z",
};

/** 每個測試一份全新的 `QueryClient`：快取不跨測試殘留（retry 關閉，理由同 CustomersPage）。 */
function newClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function mountPage(client: QueryClient = newClient()) {
  render(() => (
    <QueryClientProvider client={client}>
      <RoutesPage />
    </QueryClientProvider>
  ));
  return client;
}

async function renderPage() {
  mountPage();
  await waitFor(() => expect(screen.getByText("R1")).toBeTruthy());
}

/** 開新增對話框並回傳它的 DOM（欄位查詢收斂到這個 dialog）。 */
async function openCreateDialog(): Promise<HTMLElement> {
  fireEvent.click(screen.getByRole("button", { name: "新增車次" }));
  return await waitFor(() => screen.getByRole("dialog"));
}

beforeEach(() => {
  vi.clearAllMocks();
  listRoutesSpy.mockResolvedValue({
    routes: [EXISTING_ROUTE],
    pagination: { total: 1 },
  });
});

describe("RoutesPage", () => {
  it("清單載入：以 pagination.total 推總筆數，且不送 sort/desc（proto 沒有這兩個參數）", async () => {
    await renderPage();
    expect(listRoutesSpy).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: 20, keyword: "", includeDeleted: false })
    );
    expect(screen.getByText(/車次即派車看板的欄位\(共 1 筆\)/)).toBeTruthy();
    const sent = listRoutesSpy.mock.calls[0][0] as Record<string, unknown>;
    expect("sort" in sent).toBe(false);
    expect("desc" in sent).toBe(false);
  });

  it("啟用欄顯示停用狀態（停用＝看板不顯示此欄）", async () => {
    listRoutesSpy.mockResolvedValue({
      routes: [{ ...EXISTING_ROUTE, isActive: false }],
      pagination: { total: 1 },
    });
    mountPage();
    await waitFor(() => expect(screen.getByText("R1")).toBeTruthy());
    // 收斂到儲存格：表頭也叫「啟用」，整頁比對會打到表頭（表頭是 columnheader，不是 cell）。
    const cells = screen.getAllByRole("cell").map((c) => c.textContent ?? "");
    expect(cells).toContain("停用");
    expect(cells.join("|")).not.toContain("啟用");
  });

  it("送出關鍵字才進 query key", async () => {
    await renderPage();
    listRoutesSpy.mockClear();
    fireEvent.input(screen.getByLabelText("關鍵字"), { target: { value: "R1" } });
    // 沒送出前不得重查（D5：不得每按一鍵就查詢）。
    await Promise.resolve();
    expect(listRoutesSpy).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));
    await waitFor(() =>
      expect(listRoutesSpy).toHaveBeenCalledWith(expect.objectContaining({ keyword: "R1" }))
    );
  });

  it("含已刪除切換會換 query key", async () => {
    await renderPage();
    listRoutesSpy.mockClear();
    fireEvent.click(screen.getByLabelText("含已刪除"));
    await waitFor(() =>
      expect(listRoutesSpy).toHaveBeenCalledWith(expect.objectContaining({ includeDeleted: true }))
    );
  });

  it("名稱空白時不打建立 API", async () => {
    await renderPage();
    createRouteSpy.mockResolvedValue({ route: EXISTING_ROUTE });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("車次代號 *"), { target: { value: "R2" } });
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(screen.getByText("請輸入車次名稱")).toBeTruthy());
    expect(createRouteSpy).not.toHaveBeenCalled();
  });

  it("建立成功送出整份欄位，非數字排序歸 0（後端不驗格式）", async () => {
    await renderPage();
    createRouteSpy.mockResolvedValue({ route: EXISTING_ROUTE });
    const dialog = await openCreateDialog();
    fireEvent.input(within(dialog).getByLabelText("車次代號 *"), { target: { value: "R2" } });
    fireEvent.input(within(dialog).getByLabelText("車次名稱 *"), { target: { value: "二號路線" } });
    fireEvent.input(within(dialog).getByLabelText("說明"), { target: { value: "郊區" } });
    fireEvent.input(within(dialog).getByLabelText("排序"), { target: { value: "abc" } });
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(createRouteSpy).toHaveBeenCalledOnce());
    expect(createRouteSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        code: "R2",
        name: "二號路線",
        description: "郊區",
        sortOrder: 0,
        isActive: true,
      })
    );
    // 成功後失效 routes 前綴（並連帶讓看板車次重查）。
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });

  it("編輯帶出既有值並走 updateRoute（欄位式 presence：整份送出）", async () => {
    await renderPage();
    updateRouteSpy.mockResolvedValue({ route: EXISTING_ROUTE });
    fireEvent.click(screen.getByRole("button", { name: "編輯" }));
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    await waitFor(() =>
      expect((within(dialog).getByLabelText("車次代號 *") as HTMLInputElement).value).toBe("R1")
    );
    expect((within(dialog).getByLabelText("車次名稱 *") as HTMLInputElement).value).toBe("一號路線");
    expect((within(dialog).getByLabelText("排序") as HTMLInputElement).value).toBe("1");

    fireEvent.input(within(dialog).getByLabelText("車次名稱 *"), {
      target: { value: "一號路線（改）" },
    });
    fireEvent.submit(dialog.querySelector("form")!);
    await waitFor(() => expect(updateRouteSpy).toHaveBeenCalledOnce());
    expect(updateRouteSpy).toHaveBeenCalledWith(
      expect.objectContaining({ id: "r-1", name: "一號路線（改）", code: "R1", sortOrder: 1 })
    );
  });

  it("刪除先確認再呼叫，並失效看板車次查詢", async () => {
    const client = await (async () => {
      const c = newClient();
      mountPage(c);
      return c;
    })();
    await waitFor(() => expect(screen.getByText("R1")).toBeTruthy());
    const invalidate = vi.spyOn(client, "invalidateQueries");
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(false);
    deleteRouteSpy.mockResolvedValue({});
    fireEvent.click(screen.getByRole("button", { name: "刪除" }));
    await Promise.resolve();
    // 未確認 → 不得呼叫 API。
    expect(deleteRouteSpy).not.toHaveBeenCalled();
    confirmSpy.mockRestore();

    vi.spyOn(window, "confirm").mockReturnValue(true);
    fireEvent.click(screen.getByRole("button", { name: "刪除" }));
    await waitFor(() => expect(deleteRouteSpy).toHaveBeenCalledOnce());
    // 車次是看板的欄位來源 → 主檔變更要連帶讓看板重查。
    await waitFor(() =>
      expect(invalidate).toHaveBeenCalledWith({ queryKey: ["boardRoutes"] })
    );
    (window.confirm as unknown as { mockRestore: () => void }).mockRestore();
  });

  it("已刪除的車次顯示還原而非刪除", async () => {
    listRoutesSpy.mockResolvedValue({ routes: [DELETED_ROUTE], pagination: { total: 1 } });
    mountPage();
    await waitFor(() => expect(screen.getByText("R9")).toBeTruthy());
    expect(screen.queryByRole("button", { name: "刪除" })).toBeNull();
    restoreRouteSpy.mockResolvedValue({ route: DELETED_ROUTE });
    fireEvent.click(screen.getByRole("button", { name: "還原" }));
    await waitFor(() => expect(restoreRouteSpy).toHaveBeenCalledOnce());
  });

  it("無權限時顯示 banner", async () => {
    listRoutesSpy.mockRejectedValue(new ConnectError("denied", Code.PermissionDenied));
    mountPage();
    await waitFor(() => expect(screen.getByRole("alert")).toBeTruthy());
    expect(screen.getByRole("alert").textContent).toContain("沒有權限");
  });
});
