import { render, screen, waitFor } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 使用者 API 以 spy 取代：UsersPage 在模組層建立 connect client，
// 以 createClient 的替身同時攔截（錯誤訊息對照保持真實）。
const {
  listUsersSpy,
  createUserSpy,
  deactivateSpy,
  assignRoleSpy,
  listCompaniesSpy,
  listDepartmentsSpy,
} = vi.hoisted(() => ({
  listUsersSpy: vi.fn(),
  createUserSpy: vi.fn(),
  deactivateSpy: vi.fn(),
  assignRoleSpy: vi.fn(),
  listCompaniesSpy: vi.fn(),
  listDepartmentsSpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listUsers: listUsersSpy,
    createUser: createUserSpy,
    deactivate: deactivateSpy,
    assignRole: assignRoleSpy,
    listCompanies: listCompaniesSpy,
    listDepartments: listDepartmentsSpy,
  }),
}));

import UsersPage from "./UsersPage";

const EXISTING_USER = {
  id: "u-1",
  name: "王小明",
  email: "a@t.com",
  role: "staff",
  status: "active",
  companyId: "c-1",
};

function newClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function mountPage(client: QueryClient = newClient()) {
  render(() => (
    <QueryClientProvider client={client}>
      <UsersPage />
    </QueryClientProvider>
  ));
}

beforeEach(() => {
  vi.clearAllMocks();
  listUsersSpy.mockResolvedValue({
    users: [EXISTING_USER],
    pagination: { total: 1 },
  });
  listCompaniesSpy.mockResolvedValue({ companies: [], pagination: { total: 0 } });
  listDepartmentsSpy.mockResolvedValue({ departments: [], pagination: { total: 0 } });
});

describe("UsersPage", () => {
  it("清單載入後顯示使用者", async () => {
    mountPage();
    await waitFor(() => expect(screen.getByText("王小明")).toBeTruthy());
    expect(listUsersSpy).toHaveBeenCalledOnce();
  });

  it("篩選送出後進 query 並回第 1 頁", async () => {
    mountPage();
    await waitFor(() => expect(screen.getByText("王小明")).toBeTruthy());
    listUsersSpy.mockClear();
    const selects = screen.getAllByRole("combobox");
    // 角色篩選選 staff
    const roleSelect = selects[1] as HTMLSelectElement;
    roleSelect.value = "staff";
    roleSelect.dispatchEvent(new Event("change", { bubbles: true }));
    const filterForm = screen.getByRole("button", { name: "篩選" }).closest("form")!;
    filterForm.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await waitFor(() =>
      expect(listUsersSpy).toHaveBeenCalledWith(
        expect.objectContaining({ role: "staff", page: 1 })
      )
    );
  });

  it("停用後失效 users 前綴", async () => {
    mountPage();
    await waitFor(() => expect(screen.getByText("王小明")).toBeTruthy());
    vi.spyOn(window, "confirm").mockReturnValue(true);
    deactivateSpy.mockResolvedValue({});
    listUsersSpy.mockClear();
    const buttons = screen.getAllByRole("button", { name: "停用" });
    buttons[0].click();
    await waitFor(() => expect(deactivateSpy).toHaveBeenCalledOnce());
    await waitFor(() =>
      expect(listUsersSpy).toHaveBeenCalledWith(expect.objectContaining({ page: 1 }))
    );
    (window.confirm as unknown as { mockRestore: () => void }).mockRestore();
  });

  it("無權限時顯示 banner", async () => {
    listUsersSpy.mockRejectedValue(
      new ConnectError("denied", Code.PermissionDenied)
    );
    mountPage();
    await waitFor(() => expect(screen.getByRole("alert")).toBeTruthy());
    expect(screen.getByRole("alert").textContent).toContain("沒有權限");
  });
});
