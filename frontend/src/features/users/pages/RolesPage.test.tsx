import { fireEvent, render, screen, waitFor } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { abilityQueryOptions } from "~/lib/ability/service";

// 角色 API 以 spy 取代：RolesPage 的 client 在 `queries.ts` 模組層建立，
// 因此以 createClient 的替身攔截（ConnectError/Code 保持真實）。
const { listRolesSpy, getRolePermissionsSpy, updateRolePermissionsSpy, getAbilitySpy } =
  vi.hoisted(() => ({
    listRolesSpy: vi.fn(),
    getRolePermissionsSpy: vi.fn(),
    updateRolePermissionsSpy: vi.fn(),
    getAbilitySpy: vi.fn(),
  }));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listRoles: listRolesSpy,
    getRolePermissions: getRolePermissionsSpy,
    updateRolePermissions: updateRolePermissionsSpy,
    getAbility: getAbilitySpy,
  }),
}));

import RolesPage from "./RolesPage";

const ADMIN = {
  id: "r-1",
  name: "系統管理員",
  code: "super",
  dataScope: "all",
  isSystem: true,
  isActive: true,
};
const COMPANY_ADMIN = {
  id: "r-2",
  name: "公司管理員",
  code: "company_admin",
  dataScope: "company",
  isSystem: false,
  isActive: true,
};
const STAFF = {
  id: "r-3",
  name: "一般員工",
  code: "staff",
  dataScope: "self",
  isSystem: false,
  isActive: true,
};
const ACCOUNTING = {
  id: "r-4",
  name: "會計",
  code: "accounting",
  dataScope: "company",
  isSystem: false,
  isActive: true,
};

const PAGE_SIZE = 20;

/** 每個角色的權限規則（矩陣勾選狀態由此推導）。 */
const PERMISSIONS: Record<string, { resource: string; action: string; sortOrder: number }[]> = {
  "r-1": [
    { resource: "role", action: "read", sortOrder: 0 },
    { resource: "role", action: "manage", sortOrder: 1 },
  ],
  "r-2": [{ resource: "role", action: "read", sortOrder: 0 }],
  "r-3": [{ resource: "company", action: "read", sortOrder: 0 }],
};

/** 每個測試一份全新的 `QueryClient`：快取不跨測試殘留（retry 關閉，呼叫次數才確定）。 */
function newClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

/** 在 provider 內掛載頁面（清單資料一律經 query client 取得）；回傳 client 供測試觀察失效。 */
function mountPage() {
  const client = newClient();
  render(() => (
    <QueryClientProvider client={client}>
      <RolesPage />
    </QueryClientProvider>
  ));
  return client;
}

/** 渲染頁面並等首屏的角色選取與權限矩陣都就緒（第一筆角色被自動選取）。 */
async function renderPage() {
  const client = mountPage();
  await waitFor(() => expect(selectedRoleHeading()?.textContent).toBe("系統管理員"));
  return client;
}

/** 角色的清單列（列內按鈕）以角色名稱定位；面板標題是 heading，不會撞名。 */
function roleButton(name: string) {
  return screen.getByRole("button", { name: new RegExp(name) });
}

/** 目前面板顯示的角色（＝被選取的角色）。 */
function selectedRoleHeading() {
  return screen.queryByRole("heading", { level: 2 });
}

/** 矩陣某格的勾選狀態（Ark 的 HiddenInput 是真的 checkbox）。 */
function granted(label: string) {
  return (screen.getByRole("checkbox", { name: label }) as HTMLInputElement).checked;
}

/** 等非同步工作排空（後續斷言「次數不再增加」用）。 */
async function settle() {
  const flushed = Promise.withResolvers<void>();
  setTimeout(flushed.resolve, 0);
  await flushed.promise;
}

beforeEach(() => {
  listRolesSpy.mockReset();
  getRolePermissionsSpy.mockReset();
  updateRolePermissionsSpy.mockReset();
  getAbilitySpy.mockReset();
  listRolesSpy.mockResolvedValue({
    roles: [ADMIN, COMPANY_ADMIN],
    pagination: { total: 2 },
  });
  getRolePermissionsSpy.mockImplementation((req: { roleId: string }) =>
    Promise.resolve({ permissions: PERMISSIONS[req.roleId] ?? [] })
  );
});

describe("<RolesPage> 角色清單查詢", () => {
  it("首屏：以 page 1 與 pageSize 20 查詢，並自動選取第一筆載入矩陣", async () => {
    mountPage();

    await waitFor(() =>
      expect(getRolePermissionsSpy).toHaveBeenCalledWith({ roleId: "r-1" })
    );
    expect(listRolesSpy).toHaveBeenCalledTimes(1);
    expect(listRolesSpy).toHaveBeenCalledWith({ page: 1, pageSize: PAGE_SIZE });
    expect(selectedRoleHeading()?.textContent).toBe("系統管理員");
    expect(granted("角色 管理")).toBe(true);
  });

  it("換到第 2 頁：以 page 2 重新查詢，且新資料到達前舊角色仍在（不閃空）", async () => {
    const secondPage = Promise.withResolvers<unknown>();
    listRolesSpy.mockReset();
    listRolesSpy.mockResolvedValueOnce({
      roles: [ADMIN, COMPANY_ADMIN],
      pagination: { total: 45 },
    });
    listRolesSpy.mockReturnValueOnce(secondPage.promise);
    await renderPage();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));

    await waitFor(() => expect(listRolesSpy).toHaveBeenCalledTimes(2));
    expect(listRolesSpy).toHaveBeenLastCalledWith({ page: 2, pageSize: PAGE_SIZE });

    // 第 2 頁還在飛：placeholderData 讓舊頁的角色留在畫面上，不退回載入列。
    expect(roleButton("系統管理員")).toBeTruthy();
    expect(screen.queryByText("載入中…")).toBeNull();

    secondPage.resolve({ roles: [STAFF, ACCOUNTING], pagination: { total: 45 } });

    await waitFor(() => expect(roleButton("一般員工")).toBeTruthy());
    expect(screen.queryByRole("button", { name: /系統管理員/ })).toBeNull();
  });

  it("換頁失敗：banner 顯示錯誤，且不得把「沒拿到資料」誤顯示成空狀態", async () => {
    const failure = Promise.withResolvers<unknown>();
    listRolesSpy.mockReset();
    listRolesSpy.mockResolvedValueOnce({
      roles: [ADMIN, COMPANY_ADMIN],
      pagination: { total: 45 },
    });
    listRolesSpy.mockReturnValueOnce(failure.promise);
    await renderPage();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));
    await waitFor(() => expect(listRolesSpy).toHaveBeenCalledTimes(2));
    failure.reject(new ConnectError("伺服器暫時無法使用", Code.Internal));

    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toBe("伺服器暫時無法使用")
    );
    // 清單不得被誤判成空的（BASE 保留上一頁角色；改寫後至少不顯示空狀態字樣）。
    expect(screen.queryByText("尚無角色")).toBeNull();
  });

  it("超頁退回：回傳的 total 讓頁碼超界時夾回合法頁碼，且請求次數有界", async () => {
    // 第 2 頁的結果只剩 1 頁（例：該頁資料被刪光），目前頁碼因此超界。
    listRolesSpy.mockReset();
    listRolesSpy.mockImplementation((req: { page: number }) =>
      Promise.resolve(
        req.page === 2
          ? { roles: [STAFF], pagination: { total: 1 } }
          : { roles: [ADMIN, COMPANY_ADMIN], pagination: { total: 21 } }
      )
    );
    await renderPage();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));

    // 夾回第 1 頁 → 只重取一次（第 3 次呼叫），且不再繼續長。
    await waitFor(() => expect(listRolesSpy).toHaveBeenCalledTimes(3));
    expect(listRolesSpy).toHaveBeenLastCalledWith({ page: 1, pageSize: PAGE_SIZE });
    await settle();
    expect(listRolesSpy).toHaveBeenCalledTimes(3);
    await waitFor(() =>
      expect(screen.getByRole("button", { name: "第 1 頁" }).getAttribute("aria-current")).toBe(
        "page"
      )
    );
    expect(roleButton("系統管理員")).toBeTruthy();

    // clamp 之後面板必須是**新頁（第 1 頁）的第一筆**。若自動選取 effect 搶先用「被放棄的
    // 第 2 頁」資料選了 STAFF，`selectedId` 會指向新頁清單中不存在的角色 → 面板永久停在
    // 「請選擇角色」、矩陣不再自動載入（本回回歸），且多打一次 `getRolePermissions`。
    await waitFor(() => expect(selectedRoleHeading()?.textContent).toBe("系統管理員"));
    // 首屏 r-1 ＋ 夾回後 r-1 = 2 次；被放棄的 r-3 不得被選取（多一次 RPC 是回歸的副產品）。
    await settle();
    expect(getRolePermissionsSpy).toHaveBeenCalledTimes(2);
    expect(getRolePermissionsSpy).toHaveBeenLastCalledWith({ roleId: "r-1" });
  });

  it("換頁清空已選角色：舊選取立即失效，新頁資料到達後改選該頁第一筆", async () => {
    const secondPage = Promise.withResolvers<unknown>();
    listRolesSpy.mockReset();
    listRolesSpy.mockResolvedValueOnce({
      roles: [ADMIN, COMPANY_ADMIN],
      pagination: { total: 45 },
    });
    listRolesSpy.mockReturnValueOnce(secondPage.promise);
    await renderPage();

    // 先改選第二筆，確認選取真的跟著使用者走（不是永遠第一筆）。
    fireEvent.click(roleButton("公司管理員"));
    await waitFor(() => expect(selectedRoleHeading()?.textContent).toBe("公司管理員"));

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));

    // 換頁後舊選取立即清空（矩陣退回「請選擇角色」），不會殘留前一頁的角色。
    await waitFor(() => expect(selectedRoleHeading()).toBeNull());
    expect(screen.getByText("請選擇角色")).toBeTruthy();
    expect(getRolePermissionsSpy).toHaveBeenLastCalledWith({ roleId: "r-2" });

    secondPage.resolve({ roles: [STAFF, ACCOUNTING], pagination: { total: 45 } });

    // 新頁資料到達後自動選取該頁第一筆（與改寫前的 loadRoles 行為一致）。
    await waitFor(() => expect(selectedRoleHeading()?.textContent).toBe("一般員工"));
    await waitFor(() => expect(getRolePermissionsSpy).toHaveBeenLastCalledWith({ roleId: "r-3" }));
    expect(granted("公司 檢視")).toBe(true);
  });

  it("選取角色：載入該角色的權限矩陣並標示為選取", async () => {
    await renderPage();

    fireEvent.click(roleButton("公司管理員"));

    await waitFor(() =>
      expect(getRolePermissionsSpy).toHaveBeenLastCalledWith({ roleId: "r-2" })
    );
    expect(selectedRoleHeading()?.textContent).toBe("公司管理員");
    await waitFor(() => expect(granted("角色 檢視")).toBe(true));
    expect(granted("角色 管理")).toBe(false);
  });

  it("儲存權限：全量更新後回讀矩陣，並讓 ability 快取真的重取", async () => {
    updateRolePermissionsSpy.mockResolvedValue({});
    getAbilitySpy.mockResolvedValue({ rules: [] });
    const client = await renderPage();
    // 能力規則由路由守衛的 `ensureQueryData` 填入快取（這裡走同一條路徑、同一顆注入的 client）。
    await client.ensureQueryData(abilityQueryOptions);
    expect(getAbilitySpy).toHaveBeenCalledTimes(1);

    // 勾一格讓表單進入 dirty（未 dirty 時儲存鈕是停用的）。
    fireEvent.click(screen.getByRole("checkbox", { name: "角色 新增" }));
    const saveButton = () => screen.getByRole("button", { name: "儲存變更" }) as HTMLButtonElement;
    await waitFor(() => expect(saveButton().disabled).toBe(false));

    fireEvent.click(saveButton());

    await waitFor(() => expect(updateRolePermissionsSpy).toHaveBeenCalledTimes(1));
    expect(updateRolePermissionsSpy.mock.calls[0][0]).toMatchObject({ roleId: "r-1" });
    // 效果（不是「`invalidateQueries` 被呼叫」）：注入 client 上的 `["ability"]` 快取被標記失效
    // ——失效落到錯的 key／錯的 client 這裡就看不到。
    expect(client.getQueryState(["ability"])?.isInvalidated).toBe(true);
    // 而且真的重取：同一條取用路徑再拿一次時，`staleTime`(60s) 內的舊快取不得被回傳。
    await client.ensureQueryData({ ...abilityQueryOptions, revalidateIfStale: true });
    await waitFor(() => expect(getAbilitySpy).toHaveBeenCalledTimes(2));
    // 回讀（sort_order 正規化後）。
    await waitFor(() => expect(getRolePermissionsSpy).toHaveBeenCalledTimes(2));
  });
});
