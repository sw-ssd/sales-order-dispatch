import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 公司 API 以 spy 取代：CompaniesPage 在模組層建立 connect client，
// 因此以 createClient 的替身攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const { listCompaniesSpy, createCompanySpy, updateCompanySpy, deleteCompanySpy } = vi.hoisted(
  () => ({
    listCompaniesSpy: vi.fn(),
    createCompanySpy: vi.fn(),
    updateCompanySpy: vi.fn(),
    deleteCompanySpy: vi.fn(),
  })
);

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listCompanies: listCompaniesSpy,
    createCompany: createCompanySpy,
    updateCompany: updateCompanySpy,
    deleteCompany: deleteCompanySpy,
  }),
}));

import CompaniesPage from "./CompaniesPage";

const EXISTING_COMPANY = {
  id: "c-1",
  name: "既有公司",
  identifier: "C-001",
  taxId: "12345678",
  status: "active",
};

/** 每個測試一份全新的 `QueryClient`：快取不跨測試殘留。 */
function newClient() {
  // retry 關閉——測試裡的失敗都是刻意安排的，退避重試只會讓呼叫次數與時間變得不確定
  // （retry 謂詞本身由 `lib/query-client.test.ts` 守著）。
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

/** 在 provider 內掛載頁面（頁面的清單資料一律經 query client 取得）。 */
function mountPage(client: QueryClient = newClient()) {
  render(() => (
    <QueryClientProvider client={client}>
      <CompaniesPage />
    </QueryClientProvider>
  ));
}

/** 渲染頁面並等列表載入完成（modal 的測試都要先有列表可點）。 */
async function renderPage() {
  mountPage();
  await waitFor(() => expect(screen.getByText("既有公司")).toBeTruthy());
}

/** 開啟中的 modal 控制代號：欄位節點在 modal 開著期間不變，可在開啟時一次查好。 */
interface CompanyModal {
  form: HTMLFormElement;
  name: HTMLInputElement;
  identifier: HTMLInputElement;
  taxId: HTMLInputElement;
  status: HTMLSelectElement;
  saveButton: HTMLButtonElement;
}

/**
 * 開 modal 並回傳 modal 內的 form 與欄位。
 * 查詢一律收斂到對話框內：頁面上另有一個篩選表單，欄位標籤（狀態）會與 modal 撞名。
 */
async function openDialog(buttonName: string | RegExp): Promise<CompanyModal> {
  fireEvent.click(screen.getByRole("button", { name: buttonName }));
  const dialog = await waitFor(() => screen.getByRole("dialog"));
  const form = dialog.querySelector("form");
  if (!form) throw new Error("對話框內尚未渲染表單");
  const field = (label: RegExp) => within(dialog).getByLabelText(label);
  return {
    form,
    name: field(/公司名稱/) as HTMLInputElement,
    identifier: field(/識別碼/) as HTMLInputElement,
    taxId: field(/統一編號/) as HTMLInputElement,
    status: field(/狀態/) as HTMLSelectElement,
    saveButton: within(form).getByRole("button", { name: /儲存/ }) as HTMLButtonElement,
  };
}

async function closeDialog() {
  fireEvent.click(screen.getByRole("button", { name: "取消" }));
  await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
}

/**
 * 等表單的非同步工作排空。`isSubmitting` 要等到 `handleSubmit` 的驗證鏈跑完才會轉回 false，
 * 而錯誤訊息在那之前就出現了；不等它，下一次提交會被頁面的「提交中不重送」守門擋掉
 * （真實使用者的節奏是等按鈕重新可用後才再按一次，這裡把同一件事顯式化）。
 */
async function settle() {
  const flushed = Promise.withResolvers<void>();
  setTimeout(flushed.resolve, 0);
  await flushed.promise;
}

/** 只填有給的欄位；值一律走真實的 input/change 事件（不是直接改 DOM 值）。 */
function fillCompany(modal: CompanyModal, values: Partial<Record<"name" | "identifier" | "taxId", string>> & { status?: string }) {
  for (const key of ["name", "identifier", "taxId"] as const) {
    const value = values[key];
    if (value !== undefined) fireEvent.input(modal[key], { target: { value } });
  }
  if (values.status !== undefined) {
    fireEvent.change(modal.status, { target: { value: values.status } });
  }
}

beforeEach(() => {
  listCompaniesSpy.mockReset();
  createCompanySpy.mockReset();
  updateCompanySpy.mockReset();
  deleteCompanySpy.mockReset();
  listCompaniesSpy.mockResolvedValue({
    companies: [EXISTING_COMPANY],
    pagination: { total: 1 },
  });
});

describe("<CompaniesPage> 公司 modal 表單", () => {
  it("未填直接提交：兩欄各自出現繁中必填訊息，aria 關聯指向該錯誤元素，且不呼叫 API", async () => {
    await renderPage();
    const modal = await openDialog("新增公司");

    fireEvent.submit(modal.form);

    await waitFor(() => expect(screen.getByText("請輸入公司名稱")).toBeTruthy());
    expect(screen.getByText("請輸入識別碼(identifier)")).toBeTruthy();

    for (const [input, message] of [
      [modal.name, "請輸入公司名稱"],
      [modal.identifier, "請輸入識別碼(identifier)"],
    ] as const) {
      expect(input.getAttribute("aria-invalid")).toBe("true");
      const describedBy = input.getAttribute("aria-describedby");
      expect(describedBy).toBeTruthy();
      expect(document.getElementById(describedBy!)?.textContent).toContain(message);
    }

    expect(createCompanySpy).not.toHaveBeenCalled();
    expect(updateCompanySpy).not.toHaveBeenCalled();
  });

  it("驗證時機：輸入過程不標紅，blur 只驗該欄而不連帶標紅另一欄", async () => {
    await renderPage();
    const modal = await openDialog("新增公司");

    fireEvent.input(modal.name, { target: { value: "新" } });
    fireEvent.input(modal.name, { target: { value: "" } });
    expect(screen.queryByText("請輸入公司名稱")).toBeNull();

    fireEvent.blur(modal.name);
    await waitFor(() => expect(screen.getByText("請輸入公司名稱")).toBeTruthy());

    expect(screen.queryByText("請輸入識別碼(identifier)")).toBeNull();
    expect(modal.identifier.getAttribute("aria-invalid")).toBeNull();
  });

  it("新增：填合法值 → 錯誤消失、以正確 payload 建立公司並關閉 modal", async () => {
    createCompanySpy.mockResolvedValue({});
    await renderPage();
    const modal = await openDialog("新增公司");

    fireEvent.submit(modal.form);
    await waitFor(() => expect(screen.getByText("請輸入公司名稱")).toBeTruthy());
    await settle();

    fillCompany(modal, { name: "新公司", identifier: "C-002", taxId: " 87654321 ", status: "inactive" });
    fireEvent.submit(modal.form);

    await waitFor(() =>
      expect(createCompanySpy).toHaveBeenCalledWith({
        name: "新公司",
        taxId: "87654321",
        identifier: "C-002",
        status: "inactive",
      })
    );
    expect(screen.queryByText("請輸入公司名稱")).toBeNull();
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });

  it("編輯：帶入既有值、識別碼不可修改，提交以 {companyId, name, taxId, status} 呼叫更新", async () => {
    updateCompanySpy.mockResolvedValue({});
    await renderPage();
    const modal = await openDialog("編輯");

    expect(modal.name.value).toBe("既有公司");
    expect(modal.identifier.value).toBe("C-001");
    expect(modal.identifier.disabled).toBe(true);

    fillCompany(modal, { name: "改名後", taxId: "", status: "suspended" });
    fireEvent.submit(modal.form);

    await waitFor(() =>
      expect(updateCompanySpy).toHaveBeenCalledWith({
        companyId: "c-1",
        name: "改名後",
        taxId: "",
        status: "suspended",
      })
    );
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });

  it("伺服器錯誤：以 role=alert 的表單層 banner 呈現，欄位不被誤掛錯誤", async () => {
    createCompanySpy.mockRejectedValue(new ConnectError("already exists", Code.AlreadyExists));
    await renderPage();
    const modal = await openDialog("新增公司");

    fillCompany(modal, { name: "新公司", identifier: "C-002" });
    fireEvent.submit(modal.form);

    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain("識別碼(identifier)已存在")
    );
    expect(modal.name.getAttribute("aria-invalid")).toBeNull();
    expect(modal.identifier.getAttribute("aria-invalid")).toBeNull();
    expect(screen.getByRole("dialog")).toBeTruthy();
  });

  it("客戶端驗證失敗時清掉前一次留下的伺服器錯誤 banner", async () => {
    createCompanySpy.mockRejectedValue(new ConnectError("already exists", Code.AlreadyExists));
    await renderPage();
    const modal = await openDialog("新增公司");

    fillCompany(modal, { name: "新公司", identifier: "C-002" });
    fireEvent.submit(modal.form);
    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain("識別碼(identifier)已存在")
    );
    await settle();

    fillCompany(modal, { identifier: "" });
    fireEvent.submit(modal.form);

    await waitFor(() => expect(screen.getByText("請輸入識別碼(identifier)")).toBeTruthy());
    expect(screen.queryByRole("alert")).toBeNull();
  });

  it("關閉再開：欄位錯誤、touched 與伺服器錯誤 banner 皆已重設", async () => {
    createCompanySpy.mockRejectedValue(new ConnectError("already exists", Code.AlreadyExists));
    await renderPage();
    const modal = await openDialog("新增公司");

    // 留下欄位錯誤（未填提交）
    fireEvent.submit(modal.form);
    await waitFor(() => expect(screen.getByText("請輸入公司名稱")).toBeTruthy());
    await settle();

    // 填合法值後送出 → 伺服器錯誤 banner
    fillCompany(modal, { name: "新公司", identifier: "C-002", taxId: "87654321", status: "inactive" });
    fireEvent.submit(modal.form);
    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain("識別碼(identifier)已存在")
    );
    await settle();

    // 清空識別碼後 blur：欄位錯誤與 banner 同時在畫面上（modal 的「髒」狀態）
    fillCompany(modal, { identifier: "" });
    fireEvent.blur(modal.identifier);
    await waitFor(() => expect(screen.getByText("請輸入識別碼(identifier)")).toBeTruthy());
    expect(screen.getByRole("alert")).toBeTruthy();

    await closeDialog();
    const reopened = await openDialog("新增公司");

    expect(screen.queryByRole("alert")).toBeNull();
    expect(screen.queryByText("請輸入公司名稱")).toBeNull();
    expect(screen.queryByText("請輸入識別碼(identifier)")).toBeNull();
    expect(reopened.name.getAttribute("aria-invalid")).toBeNull();
    expect(reopened.identifier.getAttribute("aria-invalid")).toBeNull();
    expect(reopened.name.value).toBe("");
    expect(reopened.identifier.value).toBe("");
    expect(reopened.taxId.value).toBe("");
    expect(reopened.status.value).toBe("active");

    // 重開後表單仍可用：未填提交會重新出現錯誤（而不是被 touched 或殘留狀態擋住）
    fireEvent.submit(reopened.form);
    await waitFor(() => expect(screen.getByText("請輸入公司名稱")).toBeTruthy());
  });

  it("第一欄已被 blur 標紅時送出：兩個必填欄位都要標紅（F3 迴歸，機制層）", async () => {
    // 真瀏覽器的 F3 路徑：modal 開啟時自動聚焦第一欄，使用者點「儲存」造成第一欄 blur
    // （touched + invalid），此時「這一次」送出不能因此跳過其他欄位的驗證。
    await renderPage();
    const modal = await openDialog("新增公司");

    fireEvent.blur(modal.name);
    await waitFor(() => expect(screen.getByText("請輸入公司名稱")).toBeTruthy());

    fireEvent.submit(modal.form);

    await waitFor(() => expect(screen.getByText("請輸入識別碼(identifier)")).toBeTruthy());
    expect(modal.identifier.getAttribute("aria-invalid")).toBe("true");
    expect(modal.name.getAttribute("aria-invalid")).toBe("true");
    expect(createCompanySpy).not.toHaveBeenCalled();
  });

  it("重開後空送出：兩個必填欄位都要標紅並掛上 aria 關聯（F3 迴歸）", async () => {
    await renderPage();
    const first = await openDialog("新增公司");
    fireEvent.submit(first.form);
    await waitFor(() => expect(screen.getByText("請輸入識別碼(identifier)")).toBeTruthy());
    await settle();

    await closeDialog();
    const reopened = await openDialog("新增公司");

    fireEvent.submit(reopened.form);

    await waitFor(() => expect(screen.getByText("請輸入公司名稱")).toBeTruthy());
    await waitFor(() => expect(screen.getByText("請輸入識別碼(identifier)")).toBeTruthy());

    for (const [input, message] of [
      [reopened.name, "請輸入公司名稱"],
      [reopened.identifier, "請輸入識別碼(identifier)"],
    ] as const) {
      expect(input.getAttribute("aria-invalid")).toBe("true");
      const describedBy = input.getAttribute("aria-describedby");
      expect(describedBy).toBeTruthy();
      expect(document.getElementById(describedBy!)?.textContent).toContain(message);
    }

    expect(createCompanySpy).not.toHaveBeenCalled();
  });

  it("編輯後再新增：不會殘留上一筆的欄位值", async () => {
    await renderPage();
    const editing = await openDialog("編輯");
    expect(editing.name.value).toBe("既有公司");

    await closeDialog();
    const creating = await openDialog("新增公司");

    expect(creating.name.value).toBe("");
    expect(creating.identifier.value).toBe("");
    expect(creating.identifier.disabled).toBe(false);
  });

  it("提交中由表單的 isSubmitting 驅動按鈕載入狀態，且再次提交不會重複送 API", async () => {
    const pending = Promise.withResolvers<unknown>();
    createCompanySpy.mockReturnValue(pending.promise);
    await renderPage();
    const modal = await openDialog("新增公司");

    fillCompany(modal, { name: "新公司", identifier: "C-002" });
    fireEvent.submit(modal.form);

    await waitFor(() => expect(createCompanySpy).toHaveBeenCalledTimes(1));
    expect(modal.saveButton.disabled).toBe(true);
    expect(modal.saveButton.getAttribute("aria-busy")).toBe("true");

    fireEvent.submit(modal.form);

    pending.resolve({});
    await waitFor(() => expect(modal.saveButton.disabled).toBe(false));
    // 等非同步工作排空，確認第二次提交是真的沒送出（而不是還沒輪到）。
    await settle();
    expect(createCompanySpy).toHaveBeenCalledTimes(1);
  });
});

describe("<CompaniesPage> 公司清單查詢", () => {
  const SECOND_PAGE_COMPANY = {
    id: "c-2",
    name: "第二頁公司",
    identifier: "C-002",
    taxId: "",
    status: "inactive",
  };

  /** 45 筆 = 3 頁；第 2 頁回另一家公司，用來證明真的換了資料而不是沿用快取。 */
  function mockCompanyPages() {
    listCompaniesSpy.mockImplementation((req: { page: number }) =>
      Promise.resolve(
        req.page === 2
          ? { companies: [SECOND_PAGE_COMPANY], pagination: { total: 45 } }
          : { companies: [EXISTING_COMPANY], pagination: { total: 45 } }
      )
    );
  }

  /** 篩選表單（頁面上唯一的 form；modal 未開時）。 */
  function filterForm(): HTMLFormElement {
    const form = screen.getByLabelText("關鍵字").closest("form");
    if (!form) throw new Error("找不到篩選表單");
    return form;
  }

  it("首屏：以 page 1 與未設定的 keyword/status 查詢", async () => {
    await renderPage();

    expect(listCompaniesSpy).toHaveBeenCalledTimes(1);
    expect(listCompaniesSpy).toHaveBeenCalledWith({
      page: 1,
      pageSize: 20,
      sort: "",
      desc: false,
      status: undefined,
      keyword: undefined,
    });
  });

  it("換到第 2 頁：以 page 2 重新查詢，且新資料到達前舊列仍在（不閃空）", async () => {
    const secondPage = Promise.withResolvers<unknown>();
    listCompaniesSpy.mockResolvedValueOnce({
      companies: [EXISTING_COMPANY],
      pagination: { total: 45 },
    });
    listCompaniesSpy.mockReturnValueOnce(secondPage.promise);
    await renderPage();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));

    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(2));
    expect(listCompaniesSpy).toHaveBeenLastCalledWith({
      page: 2,
      pageSize: 20,
      sort: "",
      desc: false,
      status: undefined,
      keyword: undefined,
    });

    // 第 2 頁還在飛：placeholderData 讓舊頁資料留在畫面上，不退回載入列。
    expect(screen.getByText("既有公司")).toBeTruthy();
    expect(screen.queryByText("載入中…")).toBeNull();

    secondPage.resolve({ companies: [SECOND_PAGE_COMPANY], pagination: { total: 45 } });

    await waitFor(() => expect(screen.getByText("第二頁公司")).toBeTruthy());
    expect(screen.queryByText("既有公司")).toBeNull();
  });

  it("換頁失敗：banner 顯示錯誤，且不得把「沒拿到資料」誤顯示成空狀態", async () => {
    const failure = Promise.withResolvers<unknown>();
    listCompaniesSpy.mockResolvedValueOnce({
      companies: [EXISTING_COMPANY],
      pagination: { total: 45 },
    });
    listCompaniesSpy.mockReturnValueOnce(failure.promise);
    await renderPage();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));
    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(2));
    failure.reject(new ConnectError("伺服器暫時無法使用", Code.Internal));

    // 失敗要看得到（banner）……
    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toBe("伺服器暫時無法使用")
    );
    // ……但清單不得被誤判成空的（BASE 保留上一頁列；改寫後至少不顯示空狀態字樣）。
    expect(screen.queryByText("尚無公司資料")).toBeNull();
  });

  it("篩選送出：草稿不查詢、送出後帶入參數並回到第 1 頁（只查一次）", async () => {
    mockCompanyPages();
    await renderPage();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));
    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(2));

    const form = filterForm();
    // modal 的欄位在關閉時仍掛在 DOM 裡，查詢一律收斂到篩選表單內。
    fireEvent.input(within(form).getByLabelText("關鍵字"), { target: { value: "宏" } });
    fireEvent.change(within(form).getByLabelText("狀態"), { target: { value: "active" } });
    // 草稿只存在頁面 signal 裡，輸入過程不得查詢。
    expect(listCompaniesSpy).toHaveBeenCalledTimes(2);

    fireEvent.submit(form);

    await waitFor(() =>
      expect(listCompaniesSpy).toHaveBeenLastCalledWith({
        page: 1,
        pageSize: 20,
        sort: "",
        desc: false,
        keyword: "宏",
        status: "active",
      })
    );
    // 套用篩選與回第 1 頁是同一次更新 → 只觸發一次查詢。
    await settle();
    expect(listCompaniesSpy).toHaveBeenCalledTimes(3);
  });

  it("超頁退回：回傳的 total 讓目前頁碼超界時夾回合法頁碼，且請求次數有界", async () => {
    // 第 2 頁的結果只剩 1 頁（例：該頁的資料被刪光），目前頁碼因此超界。
    listCompaniesSpy.mockImplementation((req: { page: number }) =>
      Promise.resolve(
        req.page === 2
          ? { companies: [SECOND_PAGE_COMPANY], pagination: { total: 1 } }
          : { companies: [EXISTING_COMPANY], pagination: { total: 21 } }
      )
    );
    await renderPage();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));

    // 夾回第 1 頁 → 只重取一次（第 3 次呼叫），且不再繼續長。
    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(3));
    expect(listCompaniesSpy).toHaveBeenLastCalledWith({
      page: 1,
      pageSize: 20,
      sort: "",
      desc: false,
      status: undefined,
      keyword: undefined,
    });
    await settle();
    expect(listCompaniesSpy).toHaveBeenCalledTimes(3);
    await waitFor(() =>
      expect(screen.getByRole("button", { name: "第 1 頁" }).getAttribute("aria-current")).toBe(
        "page"
      )
    );
  });

  it("建立／編輯／刪除成功後清單被重新取得", async () => {
    createCompanySpy.mockResolvedValue({});
    updateCompanySpy.mockResolvedValue({});
    deleteCompanySpy.mockResolvedValue({});
    vi.spyOn(window, "confirm").mockReturnValue(true);
    await renderPage();
    expect(listCompaniesSpy).toHaveBeenCalledTimes(1);

    const creating = await openDialog("新增公司");
    fillCompany(creating, { name: "新公司", identifier: "C-002" });
    fireEvent.submit(creating.form);
    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(2));

    const editing = await openDialog("編輯");
    fillCompany(editing, { name: "改名後" });
    fireEvent.submit(editing.form);
    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(3));

    fireEvent.click(screen.getByRole("button", { name: "刪除" }));
    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(4));
  });

  it("清單載入失敗：錯誤由 query 狀態驅動，顯示在頁面層 banner", async () => {
    listCompaniesSpy.mockRejectedValue(new ConnectError("查詢被拒", Code.PermissionDenied));

    mountPage();

    await waitFor(() => expect(screen.getByRole("alert").textContent).toContain("查詢被拒"));
    expect(screen.queryByText("載入中…")).toBeNull();
    expect(listCompaniesSpy).toHaveBeenCalledTimes(1);
  });
});

describe("<CompaniesPage> 表格（TanStack Table，manual 分頁）", () => {
  const PAGE_SIZE = 20;

  /** 第 n 筆公司（n 從 1 起算）：名稱帶序號，用來辨識畫面拿到的是哪一頁的資料。 */
  function company(n: number) {
    return {
      id: `c-${n}`,
      name: `公司 ${n}`,
      identifier: `C-${n}`,
      taxId: "",
      status: "active",
    };
  }

  /** `thead` 的欄位表頭。 */
  const headerCells = () => [
    ...document.querySelectorAll("thead th"),
  ] as HTMLTableCellElement[];

  /** `tbody` 的列（含載入列與空狀態列）。 */
  const tableRows = () => [
    ...document.querySelectorAll("tbody tr"),
  ] as HTMLTableRowElement[];

  /** 伺服器端分頁的替身：每頁回 `PAGE_SIZE` 筆（最後一頁可能更少），`total` 固定。 */
  function mockPages(total: number) {
    listCompaniesSpy.mockImplementation((req: { page: number }) =>
      Promise.resolve({
        companies: Array.from(
          { length: Math.max(0, Math.min(PAGE_SIZE, total - (req.page - 1) * PAGE_SIZE)) },
          (_, i) => company((req.page - 1) * PAGE_SIZE + i + 1)
        ),
        pagination: { total },
      })
    );
  }

  /** 掛載頁面並等第 1 頁的資料上畫面。 */
  async function renderPage1() {
    mockPages(45);
    mountPage();
    await waitFor(() => expect(screen.getByText("公司 1")).toBeTruthy());
  }

  it("表頭與 columns 一致：六欄（名稱/識別碼/統一編號/狀態/ID/操作），資料列的格子數相同", async () => {
    await renderPage1();

    const headers = headerCells();
    expect(headers.map((h) => h.textContent?.trim())).toEqual([
      "名稱",
      "識別碼",
      "統一編號",
      "狀態",
      "ID",
      "操作",
    ]);
    expect(tableRows()[0].querySelectorAll("td")).toHaveLength(headers.length);
  });

  it("列數等於資料數（順序照查詢結果）", async () => {
    listCompaniesSpy.mockResolvedValue({
      companies: [company(1), company(2), company(3)],
      pagination: { total: 3 },
    });
    mountPage();
    await waitFor(() => expect(screen.getByText("公司 3")).toBeTruthy());

    const rows = tableRows();
    expect(rows).toHaveLength(3);
    expect(rows.map((r) => r.querySelector("td")?.textContent)).toEqual([
      "公司 1",
      "公司 2",
      "公司 3",
    ]);
  });

  it("載入列與空狀態列的 colspan 由欄位數推導（兩者都等於表頭欄位數）", async () => {
    const pending = Promise.withResolvers<unknown>();
    listCompaniesSpy.mockReturnValue(pending.promise);
    mountPage();

    const columnCount = headerCells().length;
    expect(columnCount).toBe(6);

    const loadingRow = (await waitFor(() => screen.getByText("載入中…"))).closest("tr");
    expect(loadingRow?.querySelector("td")?.getAttribute("colspan")).toBe(String(columnCount));
    expect(loadingRow?.querySelectorAll("td")).toHaveLength(1);

    pending.resolve({ companies: [], pagination: { total: 0 } });

    const emptyRow = (await waitFor(() => screen.getByText("尚無公司資料"))).closest("tr");
    expect(emptyRow?.querySelector("td")?.getAttribute("colspan")).toBe(String(columnCount));
  });

  it("分頁 UI 的頁碼與總數取自查詢結果（rowCount 決定頁數、頁碼來自 table）", async () => {
    await renderPage1();

    expect(screen.getByText(/第 1–20 筆,共 45 筆/)).toBeTruthy();
    expect(screen.getByRole("button", { name: "第 1 頁" }).getAttribute("aria-current")).toBe(
      "page"
    );
    expect(screen.getByRole("button", { name: "第 2 頁" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "第 3 頁" })).toBeTruthy();
    // 45 筆 / 20 = 3 頁：不得出現第 4 頁（頁數來自後端的 total，不是當前頁的列數）。
    expect(screen.queryByRole("button", { name: "第 4 頁" })).toBeNull();

    fireEvent.click(screen.getByRole("button", { name: "第 3 頁" }));
    await waitFor(() => expect(screen.getByText("公司 41")).toBeTruthy());

    expect(screen.getByText(/第 41–45 筆,共 45 筆/)).toBeTruthy();
    expect(screen.getByRole("button", { name: "第 3 頁" }).getAttribute("aria-current")).toBe(
      "page"
    );
  });

  it("點「下一頁」：以 page 2 查詢，且第 2 頁只渲染它自己的列（伺服器已分頁，table 不再切一次）", async () => {
    await renderPage1();

    fireEvent.click(screen.getByRole("button", { name: "下一頁" }));

    await waitFor(() =>
      expect(listCompaniesSpy).toHaveBeenLastCalledWith({
        page: 2,
        pageSize: 20,
        sort: "",
        desc: false,
        status: undefined,
        keyword: undefined,
      })
    );
    await waitFor(() => expect(screen.getByText("公司 21")).toBeTruthy());

    // 第 2 頁回 20 筆 → 全部渲染（若 table 又依 pageIndex 切一次，這裡會是 0 列）；
    // 頁碼也不得被 table 的 autoResetPageIndex 打回第 1 頁。
    expect(tableRows()).toHaveLength(20);
    expect(screen.getByRole("button", { name: "第 2 頁" }).getAttribute("aria-current")).toBe(
      "page"
    );
  });

  it("clamp 守衛：換頁失敗（沒有屬於當前查詢的資料）時不得改寫頁碼，也不得多發請求", async () => {
    const failure = Promise.withResolvers<unknown>();
    listCompaniesSpy.mockResolvedValueOnce({ companies: [company(1)], pagination: { total: 45 } });
    listCompaniesSpy.mockReturnValueOnce(failure.promise);
    mountPage();
    await waitFor(() => expect(screen.getByText("公司 1")).toBeTruthy());

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));
    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(2));
    failure.reject(new ConnectError("伺服器暫時無法使用", Code.Internal));

    await waitFor(() => expect(screen.getByRole("alert").textContent).toBe("伺服器暫時無法使用"));

    // 錯誤沒有 placeholder 可保留 → `query.data` 為 undefined，`total()` 會算成 0；
    // 若據以夾頁碼就會多打一次 page 1 的請求（移除守衛時本斷言變紅）。
    await settle();
    expect(listCompaniesSpy).toHaveBeenCalledTimes(2);
  });

  it("編輯成功後重取的資料會換到畫面上（row model 不得黏住舊列）", async () => {
    updateCompanySpy.mockResolvedValue({});
    listCompaniesSpy.mockResolvedValueOnce({
      companies: [company(1)],
      pagination: { total: 1 },
    });
    mountPage();
    await waitFor(() => expect(screen.getByText("公司 1")).toBeTruthy());

    // 重取回同一頁但內容已變（名稱改了）：table 的 row model 必須跟著換。
    listCompaniesSpy.mockResolvedValueOnce({
      companies: [{ ...company(1), name: "公司 1（改名後）" }],
      pagination: { total: 1 },
    });

    fireEvent.click(screen.getByRole("button", { name: "編輯" }));
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    fireEvent.input(within(dialog).getByLabelText(/公司名稱/), {
      target: { value: "公司 1（改名後）" },
    });
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(screen.getByText("公司 1（改名後）")).toBeTruthy());
    expect(screen.queryByText("公司 1")).toBeNull();
  });
});

describe("<CompaniesPage> 表頭排序（伺服器端）", () => {
  const PAGE_SIZE = 20;

  /** 第 n 筆公司（n 從 1 起算）：名稱帶序號，用來辨識畫面拿到的是哪一頁的資料。 */
  function company(n: number) {
    return {
      id: `c-${n}`,
      name: `公司 ${n}`,
      identifier: `C-${n}`,
      taxId: "",
      status: "active",
    };
  }

  /** 伺服器端分頁的替身（45 筆＝3 頁）；排序由後端負責，替身不回傳已排序的資料。 */
  function mockPages(total = 45) {
    listCompaniesSpy.mockImplementation((req: { page: number }) =>
      Promise.resolve({
        companies: Array.from(
          { length: Math.max(0, Math.min(PAGE_SIZE, total - (req.page - 1) * PAGE_SIZE)) },
          (_, i) => company((req.page - 1) * PAGE_SIZE + i + 1)
        ),
        pagination: { total },
      })
    );
  }

  /** 掛載頁面並等第 1 頁的資料上畫面。 */
  async function renderPage1() {
    mockPages();
    mountPage();
    await waitFor(() => expect(screen.getByText("公司 1")).toBeTruthy());
  }

  /** 表頭儲存格：以可及名稱定位（方向指示圖示 `aria-hidden`，不進名稱）。 */
  function headerCell(label: string): HTMLTableCellElement {
    return screen.getByRole("columnheader", { name: label }) as HTMLTableCellElement;
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
    keyword: undefined,
    status: undefined,
  });

  it("點 name 表頭：請求帶 sort:name、desc:false，且從第 2 頁回到第 1 頁（同一次更新只查一次）", async () => {
    await renderPage1();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));
    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(2));

    fireEvent.click(sortButton("名稱"));

    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(3));
    expect(listCompaniesSpy).toHaveBeenLastCalledWith(payload(1, "name", false));
    // 排序與回第 1 頁必須同批：分開寫會先以「舊頁碼＋新排序」多打一次。
    await settle();
    expect(listCompaniesSpy).toHaveBeenCalledTimes(3);
  });

  it("再點同欄：desc 反轉為 true", async () => {
    await renderPage1();

    fireEvent.click(sortButton("名稱"));
    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(2));

    fireEvent.click(sortButton("名稱"));

    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(3));
    expect(listCompaniesSpy).toHaveBeenLastCalledWith(payload(1, "name", true));
  });

  it("點另一欄：改用新欄且 desc 從 false 起算", async () => {
    await renderPage1();

    fireEvent.click(sortButton("名稱"));
    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(2));

    fireEvent.click(sortButton("ID"));

    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(3));
    expect(listCompaniesSpy).toHaveBeenLastCalledWith(payload(1, "id", false));
    expect(headerCell("名稱").getAttribute("aria-sort")).toBe("none");
  });

  it("統一編號送出後端白名單的 tax_id（欄位 id 與 accessor 名不同，不得直接送 taxId）", async () => {
    await renderPage1();

    fireEvent.click(sortButton("統一編號"));

    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(2));
    expect(listCompaniesSpy).toHaveBeenLastCalledWith(payload(1, "tax_id", false));
  });

  it("aria-sort 掛在 <th> 且三態正確；控制項是可聚焦按鈕，排序後焦點不亂跳", async () => {
    await renderPage1();

    // 初始（未排序）：可排序欄位皆為 none，方向不掛在 td 或 button 上。
    expect(headerCell("名稱").getAttribute("aria-sort")).toBe("none");
    expect(headerCell("識別碼").getAttribute("aria-sort")).toBe("none");

    const button = sortButton("名稱");
    expect(button.tagName).toBe("BUTTON");
    expect(button.getAttribute("type")).toBe("button");
    expect(button.tabIndex).toBe(0);
    button.focus();
    expect(document.activeElement).toBe(button);

    fireEvent.click(button);
    await waitFor(() => expect(headerCell("名稱").getAttribute("aria-sort")).toBe("ascending"));
    // 方向只掛在被排的那一欄。
    expect(headerCell("識別碼").getAttribute("aria-sort")).toBe("none");

    fireEvent.click(sortButton("名稱"));
    await waitFor(() => expect(headerCell("名稱").getAttribute("aria-sort")).toBe("descending"));
    // 排序變更不得重建表頭控制項：焦點留在同一顆按鈕上。
    expect(document.activeElement).toBe(button);
    expect(sortButton("名稱")).toBe(button);
  });

  it("方向指示器：升冪顯示 ▲、降冪顯示 ▼、未排序兩者皆無（同一顆表頭按鈕上更新）", async () => {
    await renderPage1();

    const button = sortButton("名稱");
    // 未排序：兩個方向都不出現（顯示任一箭頭都是誤導）。
    expect(button.textContent).not.toContain("▲");
    expect(button.textContent).not.toContain("▼");

    fireEvent.click(button);
    await waitFor(() => expect(headerCell("名稱").getAttribute("aria-sort")).toBe("ascending"));
    await waitFor(() => expect(button.textContent).toContain("▲"));
    expect(button.textContent).not.toContain("▼");

    fireEvent.click(sortButton("名稱"));
    await waitFor(() => expect(headerCell("名稱").getAttribute("aria-sort")).toBe("descending"));
    await waitFor(() => expect(button.textContent).toContain("▼"));
    expect(button.textContent).not.toContain("▲");
  });

  it("可排序欄位僅限白名單：狀態與操作欄沒有排序控制項，也不帶 aria-sort", async () => {
    await renderPage1();

    for (const label of ["狀態", "操作"]) {
      const cell = headerCell(label);
      expect(cell.querySelector("button")).toBeNull();
      // 不可排序的欄位不帶 `aria-sort`（`none` 只代表「可排序但未排」）。
      expect(cell.getAttribute("aria-sort")).toBeNull();
    }
    for (const label of ["名稱", "識別碼", "統一編號", "ID"]) {
      expect(sortButton(label).tagName).toBe("BUTTON");
      expect(headerCell(label).getAttribute("aria-sort")).toBe("none");
    }
  });

  it("換頁後排序保持：請求仍帶同一組 sort/desc", async () => {
    await renderPage1();

    fireEvent.click(sortButton("名稱"));
    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(2));

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));

    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(3));
    expect(listCompaniesSpy).toHaveBeenLastCalledWith(payload(2, "name", false));
    await waitFor(() => expect(screen.getByText("公司 21")).toBeTruthy());
    expect(headerCell("名稱").getAttribute("aria-sort")).toBe("ascending");
  });
});
