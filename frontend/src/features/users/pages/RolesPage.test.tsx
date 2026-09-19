import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
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
    expect(listRolesSpy).toHaveBeenCalledWith({ page: 1, pageSize: PAGE_SIZE, sort: "", desc: false });
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
    expect(listRolesSpy).toHaveBeenLastCalledWith({ page: 2, pageSize: PAGE_SIZE, sort: "", desc: false });

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
    expect(listRolesSpy).toHaveBeenLastCalledWith({ page: 1, pageSize: PAGE_SIZE, sort: "", desc: false });
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

describe("<RolesPage> 角色清單表格（TanStack Table，manual 分頁）", () => {
  /** 第 n 個角色（n 從 1 起算）：代碼／名稱帶序號，用來辨識畫面拿到的是哪一頁的資料。 */
  function role(n: number) {
    return {
      id: `r-${n}`,
      code: `role_${n}`,
      name: `角色 ${n}`,
      dataScope: "all",
      // 系統角色與停用各留一筆（第 1 筆＝內建啟用、第 2 筆＝自訂停用）：兩種列渲染都要對。
      isSystem: n === 1,
      isActive: n !== 2,
    };
  }

  /** 角色清單表格（同頁另有權限矩陣那張表，故以可及名稱指名）。 */
  const listTable = () => screen.getByRole("table", { name: "角色清單" });

  /** `thead` 的欄位表頭。 */
  const headerCells = () =>
    [...listTable().querySelectorAll("thead th")] as HTMLTableCellElement[];

  /** `tbody` 的列（含載入列與空狀態列）。 */
  const tableRows = () =>
    [...listTable().querySelectorAll("tbody tr")] as HTMLTableRowElement[];

  /** 伺服器端分頁的替身：每頁回 `PAGE_SIZE` 筆（最後一頁可能更少），`total` 固定。 */
  function mockPages(total: number) {
    listRolesSpy.mockImplementation((req: { page: number }) =>
      Promise.resolve({
        roles: Array.from(
          { length: Math.max(0, Math.min(PAGE_SIZE, total - (req.page - 1) * PAGE_SIZE)) },
          (_, i) => role((req.page - 1) * PAGE_SIZE + i + 1)
        ),
        pagination: { total },
      })
    );
  }

  /**
   * 掛載頁面並等第 1 頁的資料上畫面（以列的選取按鈕為準：名稱同時出現在 `<td>` 與其內層
   * `<span>`，純文字查詢會撞名；按鈕的可及名稱反而唯一且正是使用者會操作的東西）。
   */
  async function renderPage1(total = 45) {
    mockPages(total);
    mountPage();
    await waitFor(() => expect(selectButton("角色 1")).toBeTruthy());
  }

  /** 列的選取控制項：以可及名稱精確定位（名稱含角色名 → 每列的按鈕名稱唯一）。 */
  function selectButton(name: string) {
    return screen.getByRole("button", { name: `選取 ${name}` }) as HTMLButtonElement;
  }

  it("表頭與 columns 一致：六欄（角色代碼/名稱/系統角色/狀態/ID/操作），資料列的格子數相同", async () => {
    await renderPage1();

    const headers = headerCells();
    expect(headers.map((h) => h.textContent?.trim())).toEqual([
      "角色代碼",
      "名稱",
      "系統角色",
      "狀態",
      "ID",
      "操作",
    ]);

    const firstRow = tableRows()[0];
    expect(firstRow.querySelectorAll("td")).toHaveLength(headers.length);
    const cells = [...firstRow.querySelectorAll("td")].map((td) => td.textContent?.trim());
    // 角色 1：內建、啟用。
    expect(cells.slice(0, 5)).toEqual(["role_1", "角色 1", "內建", "啟用", "r-1"]);

    // 角色 2：自訂、停用（列渲染的兩種值都要正確）。
    const cells2 = [...tableRows()[1].querySelectorAll("td")].map((td) =>
      td.textContent?.trim()
    );
    expect(cells2.slice(0, 5)).toEqual(["role_2", "角色 2", "自訂", "停用", "r-2"]);
  });

  it("同頁兩張表都有可及名稱：清單表與權限矩陣都能被指名", async () => {
    await renderPage1(2);
    await waitFor(() => expect(screen.getAllByRole("table")).toHaveLength(2));

    expect(screen.getByRole("table", { name: "角色清單" })).toBeTruthy();
    expect(screen.getByRole("table", { name: "權限矩陣" })).toBeTruthy();
  });

  it("點列內「選取」按鈕：載入該角色的權限矩陣，且選取態跟著移到那一列", async () => {
    await renderPage1(2);
    // 首屏自動選取第一筆。
    await waitFor(() => expect(selectedRoleHeading()?.textContent).toBe("角色 1"));
    expect(selectButton("角色 1").getAttribute("aria-pressed")).toBe("true");

    fireEvent.click(selectButton("角色 2"));

    await waitFor(() => expect(getRolePermissionsSpy).toHaveBeenLastCalledWith({ roleId: "r-2" }));
    expect(selectedRoleHeading()?.textContent).toBe("角色 2");
    expect(selectButton("角色 2").getAttribute("aria-pressed")).toBe("true");
    expect(selectButton("角色 1").getAttribute("aria-pressed")).toBe("false");
    // 矩陣內容確實換成該角色的權限（r-2 只有「角色 檢視」）。
    await waitFor(() => expect(granted("角色 檢視")).toBe(true));
    expect(granted("角色 管理")).toBe(false);
  });

  it("鍵盤可達：「選取」是原生可聚焦按鈕且有名稱；整列本身不得可點", async () => {
    await renderPage1();
    await waitFor(() => expect(selectedRoleHeading()?.textContent).toBe("角色 1"));

    const button = selectButton("角色 2");
    expect(button.tagName).toBe("BUTTON");
    expect(button.getAttribute("type")).toBe("button");
    expect(button.disabled).toBe(false);
    expect(button.tabIndex).toBe(0);
    // 可聚焦：焦點真的停在這個控制項上（不是整列）。
    button.focus();
    expect(document.activeElement).toBe(button);

    // 整列不可點：點第 2 列的非按鈕儲存格，選取不得改變、也不得多打一次 RPC。
    const row = tableRows()[1];
    expect(row.getAttribute("tabindex")).toBeNull();
    expect(row.getAttribute("role")).toBeNull();
    fireEvent.click(row.querySelectorAll("td")[1]);
    await settle();
    expect(selectedRoleHeading()?.textContent).toBe("角色 1");
    expect(getRolePermissionsSpy).toHaveBeenCalledTimes(1);
  });

  it("換頁：清空舊選取（面板退回未選取）→ 新頁第一筆自動被選取並載入其矩陣", async () => {
    const secondPage = Promise.withResolvers<unknown>();
    listRolesSpy.mockReset();
    listRolesSpy.mockResolvedValueOnce({
      roles: [role(1), role(2)],
      pagination: { total: 45 },
    });
    listRolesSpy.mockReturnValueOnce(secondPage.promise);
    mountPage();
    await waitFor(() => expect(selectedRoleHeading()?.textContent).toBe("角色 1"));

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));

    // 換頁立即清掉舊選取（新頁的角色清單與舊頁無關）。
    await waitFor(() => expect(selectedRoleHeading()).toBeNull());
    expect(screen.getByText("請選擇角色")).toBeTruthy();

    secondPage.resolve({ roles: [role(21), role(22)], pagination: { total: 45 } });

    // 新頁資料到達後自動選取該頁第一筆（與改寫前的 `loadRoles` 同義），列上也標示出選取態。
    await waitFor(() => expect(selectedRoleHeading()?.textContent).toBe("角色 21"));
    await waitFor(() => expect(getRolePermissionsSpy).toHaveBeenLastCalledWith({ roleId: "r-21" }));
    expect(selectButton("角色 21").getAttribute("aria-pressed")).toBe("true");
    expect(selectButton("角色 22").getAttribute("aria-pressed")).toBe("false");
  });

  it("分頁 UI 的頁碼與總數取自查詢結果（頁數由 rowCount 決定、頁碼來自 table）", async () => {
    await renderPage1(45);

    expect(screen.getByText(/第 1–20 筆,共 45 筆/)).toBeTruthy();
    expect(screen.getByRole("button", { name: "第 1 頁" }).getAttribute("aria-current")).toBe(
      "page"
    );
    expect(screen.getByRole("button", { name: "第 2 頁" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "第 3 頁" })).toBeTruthy();
    // 45 筆 / 20 = 3 頁：不得出現第 4 頁（頁數來自後端的 total，不是當前頁的列數）。
    expect(screen.queryByRole("button", { name: "第 4 頁" })).toBeNull();

    fireEvent.click(screen.getByRole("button", { name: "第 3 頁" }));

    await waitFor(() => expect(listRolesSpy).toHaveBeenLastCalledWith({ page: 3, pageSize: PAGE_SIZE, sort: "", desc: false }));
    await waitFor(() => expect(selectButton("角色 41")).toBeTruthy());
    expect(screen.getByText(/第 41–45 筆,共 45 筆/)).toBeTruthy();
    expect(screen.getByRole("button", { name: "第 3 頁" }).getAttribute("aria-current")).toBe(
      "page"
    );
    // 第 3 頁只渲染它自己的 5 列（伺服器已分頁，table 不得再切一次）。
    expect(tableRows()).toHaveLength(5);
  });

  it("clamp 守衛：換頁失敗（沒有屬於當前查詢的資料）時不得改寫頁碼，也不得多發請求", async () => {
    const failure = Promise.withResolvers<unknown>();
    listRolesSpy.mockReset();
    listRolesSpy.mockResolvedValueOnce({ roles: [role(1)], pagination: { total: 45 } });
    listRolesSpy.mockReturnValueOnce(failure.promise);
    mountPage();
    await waitFor(() => expect(selectButton("角色 1")).toBeTruthy());

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));
    await waitFor(() => expect(listRolesSpy).toHaveBeenCalledTimes(2));
    failure.reject(new ConnectError("伺服器暫時無法使用", Code.Internal));

    await waitFor(() => expect(screen.getByRole("alert").textContent).toBe("伺服器暫時無法使用"));

    // 錯誤狀態沒有 placeholder 可保留（`placeholderData` 只在 status==="pending" 且
    // data===undefined 時套用）→ `total()` 會算成 0；若據以夾頁碼就會把頁碼改寫成第 1 頁
    // 並多打一次請求（移除 effect 的 `!query.data` 守衛時本斷言變紅）。
    await settle();
    expect(listRolesSpy).toHaveBeenCalledTimes(2);
  });

  it("載入列與空狀態列的 colspan 由欄位數推導（兩者都等於表頭欄位數）", async () => {
    const pending = Promise.withResolvers<unknown>();
    listRolesSpy.mockReturnValue(pending.promise);
    mountPage();

    const columnCount = headerCells().length;
    expect(columnCount).toBe(6);

    const loadingRow = (await waitFor(() => screen.getByText("載入中…"))).closest("tr");
    expect(loadingRow?.querySelector("td")?.getAttribute("colspan")).toBe(String(columnCount));
    expect(loadingRow?.querySelectorAll("td")).toHaveLength(1);

    pending.resolve({ roles: [], pagination: { total: 0 } });

    const emptyRow = (await waitFor(() => screen.getByText("尚無角色"))).closest("tr");
    expect(emptyRow?.querySelector("td")?.getAttribute("colspan")).toBe(String(columnCount));
  });
});

describe("<RolesPage> 表頭排序（伺服器端）", () => {
  /** 第 n 個角色（n 從 1 起算）：代碼／名稱帶序號，用來辨識畫面拿到的是哪一頁的資料。 */
  function role(n: number) {
    return {
      id: `r-${n}`,
      code: `role_${n}`,
      name: `角色 ${n}`,
      dataScope: "all",
      isSystem: n === 1,
      isActive: n !== 2,
    };
  }

  /** 角色清單表格（同頁另有權限矩陣那張表，故以可及名稱指名）。 */
  const listTable = () => screen.getByRole("table", { name: "角色清單" });

  /** 面板目前顯示的角色（＝被選取的角色）。 */
  const selectedRoleHeading = () => screen.queryByRole("heading", { level: 2 });

  /** 伺服器端分頁的替身（45 筆＝3 頁）；排序由後端負責，替身不回傳已排序的資料。 */
  function mockPages(total = 45) {
    listRolesSpy.mockImplementation((req: { page: number }) =>
      Promise.resolve({
        roles: Array.from(
          { length: Math.max(0, Math.min(PAGE_SIZE, total - (req.page - 1) * PAGE_SIZE)) },
          (_, i) => role((req.page - 1) * PAGE_SIZE + i + 1)
        ),
        pagination: { total },
      })
    );
  }

  /** 掛載頁面並等第 1 頁的資料上畫面（以列內選取按鈕為準，首筆已被自動選取）。 */
  async function renderPage1() {
    mockPages();
    mountPage();
    await waitFor(() => expect(screen.getByRole("button", { name: "選取 角色 1" })).toBeTruthy());
  }

  /** 表頭儲存格：以可及名稱定位（只從清單表格取；方向圖示 `aria-hidden` 不進名稱）。 */
  function headerCell(label: string): HTMLTableCellElement {
    return within(listTable()).getByRole("columnheader", {
      name: label,
    }) as HTMLTableCellElement;
  }

  /** 表頭上的排序控制項（可及名稱＝欄位名，不含方向指示）。 */
  function sortButton(label: string): HTMLButtonElement {
    return within(headerCell(label)).getByRole("button", { name: label }) as HTMLButtonElement;
  }

  /** 清單請求的完整 payload（排序波次後 query key 一律帶 sort/desc）。 */
  const payload = (page: number, sort: string, desc: boolean) => ({
    page,
    pageSize: PAGE_SIZE,
    sort,
    desc,
  });

  it("點 name 表頭：請求帶 sort:name、desc:false，且從第 2 頁回到第 1 頁（同一次更新只查一次）", async () => {
    await renderPage1();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));
    await waitFor(() => expect(listRolesSpy).toHaveBeenCalledTimes(2));

    fireEvent.click(sortButton("名稱"));

    await waitFor(() => expect(listRolesSpy).toHaveBeenCalledTimes(3));
    expect(listRolesSpy).toHaveBeenLastCalledWith(payload(1, "name", false));
    // 排序與回第 1 頁必須同批：分開寫會先以「舊頁碼＋新排序」多打一次。
    await settle();
    expect(listRolesSpy).toHaveBeenCalledTimes(3);
  });

  it("再點同欄：desc 反轉為 true", async () => {
    await renderPage1();

    fireEvent.click(sortButton("名稱"));
    await waitFor(() => expect(listRolesSpy).toHaveBeenCalledTimes(2));

    fireEvent.click(sortButton("名稱"));

    await waitFor(() => expect(listRolesSpy).toHaveBeenCalledTimes(3));
    expect(listRolesSpy).toHaveBeenLastCalledWith(payload(1, "name", true));
  });

  it("點另一欄：改用新欄且 desc 從 false 起算", async () => {
    await renderPage1();

    fireEvent.click(sortButton("名稱"));
    await waitFor(() => expect(listRolesSpy).toHaveBeenCalledTimes(2));

    fireEvent.click(sortButton("角色代碼"));

    await waitFor(() => expect(listRolesSpy).toHaveBeenCalledTimes(3));
    expect(listRolesSpy).toHaveBeenLastCalledWith(payload(1, "code", false));
    expect(headerCell("名稱").getAttribute("aria-sort")).toBe("none");
  });

  it("aria-sort 掛在 <th> 且三態正確；控制項是可聚焦按鈕，排序後焦點不亂跳", async () => {
    await renderPage1();

    // 初始（未排序）：可排序欄位皆為 none。
    expect(headerCell("名稱").getAttribute("aria-sort")).toBe("none");
    expect(headerCell("ID").getAttribute("aria-sort")).toBe("none");

    const button = sortButton("名稱");
    expect(button.tagName).toBe("BUTTON");
    expect(button.getAttribute("type")).toBe("button");
    expect(button.tabIndex).toBe(0);
    button.focus();
    expect(document.activeElement).toBe(button);

    fireEvent.click(button);
    await waitFor(() => expect(headerCell("名稱").getAttribute("aria-sort")).toBe("ascending"));
    // 方向只掛在被排的那一欄。
    expect(headerCell("ID").getAttribute("aria-sort")).toBe("none");

    fireEvent.click(sortButton("名稱"));
    await waitFor(() => expect(headerCell("名稱").getAttribute("aria-sort")).toBe("descending"));
    // 排序變更不得重建表頭控制項：焦點留在同一顆按鈕上。
    expect(document.activeElement).toBe(button);
    expect(sortButton("名稱")).toBe(button);
  });

  it("可排序欄位僅限白名單：系統角色、狀態與操作欄沒有排序控制項，也不帶 aria-sort", async () => {
    await renderPage1();

    for (const label of ["系統角色", "狀態", "操作"]) {
      const cell = headerCell(label);
      expect(cell.querySelector("button")).toBeNull();
      // 不可排序的欄位不帶 `aria-sort`（`none` 只代表「可排序但未排」）。
      expect(cell.getAttribute("aria-sort")).toBeNull();
    }
    for (const label of ["角色代碼", "名稱", "ID"]) {
      expect(sortButton(label).tagName).toBe("BUTTON");
      expect(headerCell(label).getAttribute("aria-sort")).toBe("none");
    }
  });

  it("換頁後排序保持：請求仍帶同一組 sort/desc", async () => {
    await renderPage1();

    fireEvent.click(sortButton("名稱"));
    await waitFor(() => expect(listRolesSpy).toHaveBeenCalledTimes(2));

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));

    await waitFor(() => expect(listRolesSpy).toHaveBeenCalledTimes(3));
    expect(listRolesSpy).toHaveBeenLastCalledWith(payload(2, "name", false));
    await waitFor(() => expect(selectedRoleHeading()?.textContent).toBe("角色 21"));
    expect(headerCell("名稱").getAttribute("aria-sort")).toBe("ascending");
  });

  it("排序變更回第 1 頁：清空舊選取並改選新第 1 頁第一筆（面板不得停在找不到的角色）", async () => {
    await renderPage1();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));
    await waitFor(() => expect(selectedRoleHeading()?.textContent).toBe("角色 21"));

    fireEvent.click(sortButton("名稱"));

    // 回到第 1 頁的新清單與舊選取無關 → 與換頁同一條規則：清空並自動選取該頁第一筆。
    await waitFor(() => expect(selectedRoleHeading()?.textContent).toBe("角色 1"));
    expect(getRolePermissionsSpy).toHaveBeenLastCalledWith({ roleId: "r-1" });
    expect(listRolesSpy).toHaveBeenLastCalledWith(payload(1, "name", false));
  });
});
