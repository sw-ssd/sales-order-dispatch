import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 部門／公司 API 以 spy 取代：DepartmentsPage 在模組層建立兩個 connect client，
// 因此以 createClient 的替身同時攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const {
  listDepartmentsSpy,
  createDepartmentSpy,
  updateDepartmentSpy,
  deleteDepartmentSpy,
  listCompaniesSpy,
  getCompanySpy,
} = vi.hoisted(() => ({
  listDepartmentsSpy: vi.fn(),
  createDepartmentSpy: vi.fn(),
  updateDepartmentSpy: vi.fn(),
  deleteDepartmentSpy: vi.fn(),
  listCompaniesSpy: vi.fn(),
  getCompanySpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listDepartments: listDepartmentsSpy,
    createDepartment: createDepartmentSpy,
    updateDepartment: updateDepartmentSpy,
    deleteDepartment: deleteDepartmentSpy,
    listCompanies: listCompaniesSpy,
    getCompany: getCompanySpy,
  }),
}));

import DepartmentsPage from "./DepartmentsPage";

const EXISTING_COMPANY = { id: "c-1", name: "既有公司" };

const EXISTING_DEPARTMENT = {
  id: "d-1",
  name: "業務部",
  companyId: "c-1",
  companyName: "既有公司",
};

/** 每個測試一份全新的 `QueryClient`：快取不跨測試殘留。 */
function newClient() {
  // retry 關閉——測試裡的失敗都是刻意安排的，退避重試只會讓呼叫次數與時間變得不確定
  // （retry 謂詞本身由 `lib/query-client.test.ts` 守著）。
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

/** 在 provider 內掛載頁面（部門清單與公司下拉一律經 query client 取得）。 */
function mountPage(client: QueryClient = newClient()) {
  render(() => (
    <QueryClientProvider client={client}>
      <DepartmentsPage />
    </QueryClientProvider>
  ));
}

/** 渲染頁面並等列表載入完成（modal 的測試都要先有列表可點）。 */
async function renderPage() {
  mountPage();
  await waitFor(() => expect(screen.getByText("業務部")).toBeTruthy());
}

/** 開啟中的 modal 控制代號：欄位節點在 modal 開著期間不變，可在開啟時一次查好。 */
interface DepartmentModal {
  form: HTMLFormElement;
  name: HTMLInputElement;
  company: HTMLSelectElement;
  saveButton: HTMLButtonElement;
}

/**
 * 開 modal 並回傳 modal 內的 form 與欄位。
 * 查詢一律收斂到對話框內：頁面上另有兩個篩選表單，欄位標籤（所屬公司）會與 modal 撞名。
 */
async function openDialog(buttonName: string | RegExp): Promise<DepartmentModal> {
  fireEvent.click(screen.getByRole("button", { name: buttonName }));
  const dialog = await waitFor(() => screen.getByRole("dialog"));
  const form = dialog.querySelector("form");
  if (!form) throw new Error("對話框內尚未渲染表單");
  return {
    form,
    name: within(dialog).getByLabelText(/部門名稱/) as HTMLInputElement,
    company: within(dialog).getByLabelText(/所屬公司/) as HTMLSelectElement,
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
function fillDepartment(modal: DepartmentModal, values: { name?: string; company?: string }) {
  if (values.name !== undefined) {
    fireEvent.input(modal.name, { target: { value: values.name } });
  }
  if (values.company !== undefined) {
    fireEvent.change(modal.company, { target: { value: values.company } });
  }
}

beforeEach(() => {
  for (const spy of [
    listDepartmentsSpy,
    createDepartmentSpy,
    updateDepartmentSpy,
    deleteDepartmentSpy,
    listCompaniesSpy,
    getCompanySpy,
  ]) {
    spy.mockReset();
  }
  listDepartmentsSpy.mockResolvedValue({
    departments: [EXISTING_DEPARTMENT],
    pagination: { total: 1 },
  });
  listCompaniesSpy.mockResolvedValue({
    companies: [EXISTING_COMPANY],
    pagination: { total: 1 },
  });
});

describe("<DepartmentsPage> 部門 modal 表單", () => {
  it("未填直接提交：兩欄各自出現繁中必填訊息，aria 關聯指向該錯誤元素，且不呼叫 API", async () => {
    // 公司清單為空 → 新增模式的公司預設值是空字串（「請選擇公司」選項），必填規則才有機會擋下
    listCompaniesSpy.mockResolvedValue({ companies: [], pagination: { total: 0 } });
    await renderPage();
    const modal = await openDialog("新增部門");

    fireEvent.submit(modal.form);

    await waitFor(() => expect(screen.getByText("請輸入部門名稱")).toBeTruthy());
    expect(screen.getByText("請選擇所屬公司")).toBeTruthy();

    for (const [control, message] of [
      [modal.name, "請輸入部門名稱"],
      [modal.company, "請選擇所屬公司"],
    ] as const) {
      expect(control.getAttribute("aria-invalid")).toBe("true");
      const describedBy = control.getAttribute("aria-describedby");
      expect(describedBy).toBeTruthy();
      expect(document.getElementById(describedBy!)?.textContent).toContain(message);
    }

    expect(createDepartmentSpy).not.toHaveBeenCalled();
    expect(updateDepartmentSpy).not.toHaveBeenCalled();
  });

  it("驗證時機：輸入與選擇過程不標紅，blur 只驗該欄而不連帶標紅另一欄", async () => {
    listCompaniesSpy.mockResolvedValue({ companies: [], pagination: { total: 0 } });
    await renderPage();
    const modal = await openDialog("新增部門");

    fireEvent.input(modal.name, { target: { value: "客" } });
    fireEvent.input(modal.name, { target: { value: "" } });
    expect(screen.queryByText("請輸入部門名稱")).toBeNull();

    fireEvent.blur(modal.name);
    await waitFor(() => expect(screen.getByText("請輸入部門名稱")).toBeTruthy());

    expect(screen.queryByText("請選擇所屬公司")).toBeNull();
    expect(modal.company.getAttribute("aria-invalid")).toBeNull();

    fireEvent.blur(modal.company);
    await waitFor(() => expect(screen.getByText("請選擇所屬公司")).toBeTruthy());
  });

  it("新增：填合法值 → 錯誤消失、以正確 payload（含 companyId）建立部門並關閉 modal", async () => {
    createDepartmentSpy.mockResolvedValue({});
    await renderPage();
    const modal = await openDialog("新增部門");

    // 公司預設帶入清單第一筆，先提交一次製造名稱的欄位錯誤
    fireEvent.submit(modal.form);
    await waitFor(() => expect(screen.getByText("請輸入部門名稱")).toBeTruthy());
    await settle();

    fillDepartment(modal, { name: " 客服部 " });
    fireEvent.submit(modal.form);

    await waitFor(() =>
      expect(createDepartmentSpy).toHaveBeenCalledWith({ companyId: "c-1", name: "客服部" })
    );
    expect(screen.queryByText("請輸入部門名稱")).toBeNull();
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });

  it("編輯：帶入既有值、公司不可修改，提交以 {departmentId, name} 呼叫更新", async () => {
    updateDepartmentSpy.mockResolvedValue({});
    await renderPage();
    const modal = await openDialog("編輯");

    expect(modal.name.value).toBe("業務部");
    expect(modal.company.value).toBe("c-1");
    expect(modal.company.disabled).toBe(true);

    fillDepartment(modal, { name: "業務一部" });
    fireEvent.submit(modal.form);

    await waitFor(() =>
      expect(updateDepartmentSpy).toHaveBeenCalledWith({ departmentId: "d-1", name: "業務一部" })
    );
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });

  it("伺服器錯誤：以 role=alert 的表單層 banner 呈現，欄位不被誤掛錯誤", async () => {
    createDepartmentSpy.mockRejectedValue(new ConnectError("部門名稱重複", Code.AlreadyExists));
    await renderPage();
    const modal = await openDialog("新增部門");

    fillDepartment(modal, { name: "業務部" });
    fireEvent.submit(modal.form);

    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain("部門名稱重複")
    );
    expect(modal.name.getAttribute("aria-invalid")).toBeNull();
    expect(modal.company.getAttribute("aria-invalid")).toBeNull();
    expect(screen.getByRole("dialog")).toBeTruthy();
  });

  it("客戶端驗證失敗時清掉前一次留下的伺服器錯誤 banner", async () => {
    createDepartmentSpy.mockRejectedValue(new ConnectError("部門名稱重複", Code.AlreadyExists));
    await renderPage();
    const modal = await openDialog("新增部門");

    fillDepartment(modal, { name: "業務部" });
    fireEvent.submit(modal.form);
    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain("部門名稱重複")
    );
    await settle();

    fillDepartment(modal, { name: "" });
    fireEvent.submit(modal.form);

    await waitFor(() => expect(screen.getByText("請輸入部門名稱")).toBeTruthy());
    expect(screen.queryByRole("alert")).toBeNull();
  });

  it("關閉再開：欄位錯誤、touched 與伺服器錯誤 banner 皆已重設", async () => {
    createDepartmentSpy.mockRejectedValue(new ConnectError("部門名稱重複", Code.AlreadyExists));
    await renderPage();
    const modal = await openDialog("新增部門");

    // 留下欄位錯誤（未填提交）
    fireEvent.submit(modal.form);
    await waitFor(() => expect(screen.getByText("請輸入部門名稱")).toBeTruthy());
    await settle();

    // 填合法值後送出 → 伺服器錯誤 banner（公司明確選定，讓後續步驟不依賴新增模式的預設值）
    fillDepartment(modal, { name: "業務部", company: "c-1" });
    fireEvent.submit(modal.form);
    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain("部門名稱重複")
    );
    await settle();

    // 清空名稱後 blur：欄位錯誤與 banner 同時在畫面上（modal 的「髒」狀態）
    fillDepartment(modal, { name: "" });
    fireEvent.blur(modal.name);
    await waitFor(() => expect(screen.getByText("請輸入部門名稱")).toBeTruthy());
    expect(screen.getByRole("alert")).toBeTruthy();

    await closeDialog();
    const reopened = await openDialog("新增部門");

    expect(screen.queryByRole("alert")).toBeNull();
    expect(screen.queryByText("請輸入部門名稱")).toBeNull();
    expect(reopened.name.getAttribute("aria-invalid")).toBeNull();
    expect(reopened.company.getAttribute("aria-invalid")).toBeNull();
    expect(reopened.name.value).toBe("");
    expect(reopened.company.value).toBe("c-1");

    // 重開後表單仍可用：未填提交會重新出現錯誤（而不是被 touched 或殘留狀態擋住）
    fireEvent.submit(reopened.form);
    await waitFor(() => expect(screen.getByText("請輸入部門名稱")).toBeTruthy());
  });

  it("名稱已被 blur 標紅時送出：兩個必填欄位都要標紅（F3 迴歸，機制層）", async () => {
    // 真瀏覽器的 F3 路徑：modal 開啟時自動聚焦部門名稱，使用者點「儲存」造成該欄 blur
    // （touched + invalid），此時「這一次」送出不能因此跳過所屬公司的驗證。
    listCompaniesSpy.mockResolvedValue({ companies: [], pagination: { total: 0 } });
    await renderPage();
    const modal = await openDialog("新增部門");

    fireEvent.blur(modal.name);
    await waitFor(() => expect(screen.getByText("請輸入部門名稱")).toBeTruthy());

    fireEvent.submit(modal.form);

    await waitFor(() => expect(screen.getByText("請選擇所屬公司")).toBeTruthy());
    expect(modal.company.getAttribute("aria-invalid")).toBe("true");
    expect(modal.name.getAttribute("aria-invalid")).toBe("true");
    expect(createDepartmentSpy).not.toHaveBeenCalled();
  });

  it("重開後空送出：兩個必填欄位都要標紅並掛上 aria 關聯（F3 迴歸）", async () => {
    listCompaniesSpy.mockResolvedValue({ companies: [], pagination: { total: 0 } });
    await renderPage();
    const first = await openDialog("新增部門");
    fireEvent.submit(first.form);
    await waitFor(() => expect(screen.getByText("請選擇所屬公司")).toBeTruthy());
    await settle();

    await closeDialog();
    const reopened = await openDialog("新增部門");

    fireEvent.submit(reopened.form);

    await waitFor(() => expect(screen.getByText("請輸入部門名稱")).toBeTruthy());
    await waitFor(() => expect(screen.getByText("請選擇所屬公司")).toBeTruthy());

    for (const [control, message] of [
      [reopened.name, "請輸入部門名稱"],
      [reopened.company, "請選擇所屬公司"],
    ] as const) {
      expect(control.getAttribute("aria-invalid")).toBe("true");
      const describedBy = control.getAttribute("aria-describedby");
      expect(describedBy).toBeTruthy();
      expect(document.getElementById(describedBy!)?.textContent).toContain(message);
    }

    expect(createDepartmentSpy).not.toHaveBeenCalled();
  });

  it("編輯後再新增：不會殘留上一筆的欄位值", async () => {
    await renderPage();
    const editing = await openDialog("編輯");
    expect(editing.name.value).toBe("業務部");

    await closeDialog();
    const creating = await openDialog("新增部門");

    expect(creating.name.value).toBe("");
    expect(creating.company.value).toBe("c-1");
    expect(creating.company.disabled).toBe(false);
  });

  it("提交中由表單的 isSubmitting 驅動按鈕載入狀態，且再次提交不會重複送 API", async () => {
    const pending = Promise.withResolvers<unknown>();
    createDepartmentSpy.mockReturnValue(pending.promise);
    await renderPage();
    const modal = await openDialog("新增部門");

    fillDepartment(modal, { name: "客服部" });
    fireEvent.submit(modal.form);

    await waitFor(() => expect(createDepartmentSpy).toHaveBeenCalledTimes(1));
    expect(modal.saveButton.disabled).toBe(true);
    expect(modal.saveButton.getAttribute("aria-busy")).toBe("true");

    fireEvent.submit(modal.form);

    pending.resolve({});
    await waitFor(() => expect(modal.saveButton.disabled).toBe(false));
    // 等非同步工作排空，確認第二次提交是真的沒送出（而不是還沒輪到）。
    await settle();
    expect(createDepartmentSpy).toHaveBeenCalledTimes(1);
  });
});

describe("<DepartmentsPage> 部門清單與公司下拉查詢", () => {
  const SECOND_PAGE_DEPARTMENT = {
    id: "d-2",
    name: "第二頁部門",
    companyId: "c-1",
    companyName: "既有公司",
  };
  /** 公司下拉的真實規模：每頁 50 筆、共 120 家 = 3 頁（第 3 頁 21 筆）。 */
  const companyAt = (n: number) => ({ id: `c-${n}`, name: `公司 ${n}` });

  /** 45 筆 = 3 頁；第 2 頁回另一筆部門，用來證明真的換了資料而不是沿用快取。 */
  function mockDepartmentPages() {
    listDepartmentsSpy.mockImplementation((req: { page: number }) =>
      Promise.resolve(
        req.page === 2
          ? { departments: [SECOND_PAGE_DEPARTMENT], pagination: { total: 45 } }
          : { departments: [EXISTING_DEPARTMENT], pagination: { total: 45 } }
      )
    );
  }

  /** 部門篩選表單（頁面上另有公司搜尋表單，且 modal 內也有一個「所屬公司」下拉 → 以 id 收斂）。 */
  function departmentFilterForm(): HTMLFormElement {
    const form = document.getElementById("department-company-filter")?.closest("form");
    if (!form) throw new Error("找不到部門篩選表單");
    return form;
  }

  /** 公司下拉目前渲染出的選項 id（排除「全部公司」的空值選項）。 */
  function companyOptionIds(): string[] {
    const select = document.getElementById("department-company-filter");
    if (!select) throw new Error("找不到公司下拉");
    return Array.from(select.querySelectorAll("option"))
      .map((option) => option.value)
      .filter((value) => value !== "");
  }

  it("首屏：以 page 1、pageSize 20 與未設定的 companyId 查詢部門", async () => {
    await renderPage();

    expect(listDepartmentsSpy).toHaveBeenCalledTimes(1);
    expect(listDepartmentsSpy).toHaveBeenCalledWith({
      page: 1,
      pageSize: 20,
      sort: "",
      desc: false,
      companyId: undefined,
    });
  });

  it("換到第 2 頁：以 page 2 重新查詢，且新資料到達前舊列仍在（不閃空）", async () => {
    const secondPage = Promise.withResolvers<unknown>();
    listDepartmentsSpy.mockResolvedValueOnce({
      departments: [EXISTING_DEPARTMENT],
      pagination: { total: 45 },
    });
    listDepartmentsSpy.mockReturnValueOnce(secondPage.promise);
    await renderPage();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));

    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(2));
    expect(listDepartmentsSpy).toHaveBeenLastCalledWith({
      page: 2,
      pageSize: 20,
      sort: "",
      desc: false,
      companyId: undefined,
    });

    // 第 2 頁還在飛：placeholderData 讓舊頁資料留在畫面上，不退回載入列。
    expect(screen.getByText("業務部")).toBeTruthy();
    expect(screen.queryByText("載入中…")).toBeNull();

    secondPage.resolve({ departments: [SECOND_PAGE_DEPARTMENT], pagination: { total: 45 } });

    await waitFor(() => expect(screen.getByText("第二頁部門")).toBeTruthy());
    expect(screen.queryByText("業務部")).toBeNull();
  });

  it("換頁失敗：banner 顯示錯誤，且不得把「沒拿到資料」誤顯示成空狀態", async () => {
    const failure = Promise.withResolvers<unknown>();
    listDepartmentsSpy.mockResolvedValueOnce({
      departments: [EXISTING_DEPARTMENT],
      pagination: { total: 45 },
    });
    listDepartmentsSpy.mockReturnValueOnce(failure.promise);
    await renderPage();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));
    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(2));
    failure.reject(new ConnectError("伺服器暫時無法使用", Code.Internal));

    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toBe("伺服器暫時無法使用")
    );
    expect(screen.queryByText("尚無部門資料")).toBeNull();
  });

  it("所屬公司篩選送出：草稿不查詢、送出後帶入參數並回到第 1 頁（只查一次）", async () => {
    mockDepartmentPages();
    await renderPage();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));
    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(2));

    const form = departmentFilterForm();
    const select = within(form).getByLabelText("所屬公司");
    fireEvent.change(select, { target: { value: "c-1" } });
    // 草稿只存在頁面 signal 裡，未送出不得查詢。
    expect(listDepartmentsSpy).toHaveBeenCalledTimes(2);

    fireEvent.submit(form);

    await waitFor(() =>
      expect(listDepartmentsSpy).toHaveBeenLastCalledWith({
        page: 1,
        pageSize: 20,
        sort: "",
        desc: false,
        companyId: "c-1",
      })
    );
    // 套用篩選與回第 1 頁是同一次更新 → 只觸發一次查詢。
    await settle();
    expect(listDepartmentsSpy).toHaveBeenCalledTimes(3);
  });

  it("超頁退回：回傳的 total 讓目前頁碼超界時夾回合法頁碼，且請求次數有界", async () => {
    // 第 2 頁的結果只剩 1 頁（例：該頁的資料被刪光），目前頁碼因此超界。
    listDepartmentsSpy.mockImplementation((req: { page: number }) =>
      Promise.resolve(
        req.page === 2
          ? { departments: [SECOND_PAGE_DEPARTMENT], pagination: { total: 1 } }
          : { departments: [EXISTING_DEPARTMENT], pagination: { total: 21 } }
      )
    );
    await renderPage();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));

    // 夾回第 1 頁 → 只重取一次（第 3 次呼叫），且不再繼續長。
    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(3));
    expect(listDepartmentsSpy).toHaveBeenLastCalledWith({
      page: 1,
      pageSize: 20,
      sort: "",
      desc: false,
      companyId: undefined,
    });
    await settle();
    expect(listDepartmentsSpy).toHaveBeenCalledTimes(3);
    await waitFor(() =>
      expect(screen.getByRole("button", { name: "第 1 頁" }).getAttribute("aria-current")).toBe(
        "page"
      )
    );
  });

  it("公司下拉：首屏載入一頁 50 筆，「載入更多」累積後續頁且選項不重複", async () => {
    // 120 家＝3 頁（50/50/21）；第 2 頁刻意回一筆與第 1 頁重複的公司（c-50），
    // 用來釘住累積時的去重（改寫前的手寫累積版就會去重）。
    listCompaniesSpy.mockImplementation((req: { page: number }) =>
      Promise.resolve({
        companies:
          req.page === 3
            ? Array.from({ length: 21 }, (_, i) => companyAt(100 + i))
            : req.page === 2
              ? [companyAt(50), ...Array.from({ length: 49 }, (_, i) => companyAt(51 + i))]
              : Array.from({ length: 50 }, (_, i) => companyAt(1 + i)),
        pagination: { total: 120 },
      })
    );
    await renderPage();

    expect(listCompaniesSpy).toHaveBeenCalledTimes(1);
    expect(listCompaniesSpy).toHaveBeenCalledWith({
      page: 1,
      pageSize: 50,
      keyword: undefined,
    });
    const firstPageIds = Array.from({ length: 50 }, (_, i) => `c-${i + 1}`);
    expect(companyOptionIds()).toEqual(firstPageIds);

    const loadMore = screen.getByRole("button", { name: "載入更多" }) as HTMLButtonElement;
    fireEvent.click(loadMore);

    await waitFor(() => expect(listCompaniesSpy).toHaveBeenCalledTimes(2));
    expect(listCompaniesSpy).toHaveBeenLastCalledWith({
      page: 2,
      pageSize: 50,
      keyword: undefined,
    });

    // 累積：選項單調增加、第 1 頁仍在最前面，且不重複（第 2 頁的 c-50 不新增第二個選項）。
    await waitFor(() => expect(companyOptionIds().length).toBeGreaterThan(50));
    expect(companyOptionIds().length).toBe(99);
    expect(companyOptionIds().slice(0, 50)).toEqual(firstPageIds);
    expect(new Set(companyOptionIds()).size).toBe(companyOptionIds().length);

    fireEvent.click(loadMore);

    await waitFor(() => expect(companyOptionIds().length).toBe(120));
    expect(new Set(companyOptionIds()).size).toBe(120);
    expect(document.body.textContent).toContain("已載入 120 家,共 120 家");

    // 全部載完 → 「載入更多」停用（不再有下一頁）。
    await waitFor(() => expect(loadMore.disabled).toBe(true));
  });

  it("編輯不在已載入分頁內的公司：補載後 modal 的所屬公司仍能選中它", async () => {
    // 部門頁的下拉同時是 modal 的選項來源：該部門的公司若不在已載入分頁內，
    // modal 的 <select> 不能顯示成別的選項（補載的那筆必須出現在選項裡）。
    listDepartmentsSpy.mockResolvedValue({
      departments: [{ id: "d-9", name: "外部部門", companyId: "c-9", companyName: "未載入公司" }],
      pagination: { total: 1 },
    });
    getCompanySpy.mockResolvedValue({ company: { id: "c-9", name: "未載入公司" } });
    mountPage();
    await waitFor(() => expect(screen.getByText("外部部門")).toBeTruthy());

    fireEvent.click(screen.getByRole("button", { name: "編輯" }));

    const dialog = await waitFor(() => screen.getByRole("dialog"));
    const select = within(dialog).getByLabelText(/所屬公司/) as HTMLSelectElement;
    await waitFor(() => expect(select.value).toBe("c-9"));
    expect(getCompanySpy).toHaveBeenCalledWith({ companyId: "c-9" });
  });

  it("公司關鍵字搜尋送出後：補載釘住的公司不再殘留在選項裡", async () => {
    // 先造出「補載釘住」的狀態：編輯的部門，其公司不在已載入的分頁內。
    listDepartmentsSpy.mockResolvedValue({
      departments: [{ id: "d-9", name: "外部部門", companyId: "c-9", companyName: "未載入公司" }],
      pagination: { total: 1 },
    });
    getCompanySpy.mockResolvedValue({ company: { id: "c-9", name: "未載入公司" } });
    mountPage();
    await waitFor(() => expect(screen.getByText("外部部門")).toBeTruthy());

    fireEvent.click(screen.getByRole("button", { name: "編輯" }));
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    await waitFor(() => expect(within(dialog).getByLabelText(/所屬公司/)).toBeTruthy());
    expect(companyOptionIds()).toContain("c-9");
    await closeDialog();

    // 換一組關鍵字送出：補載的那筆不屬於這組關鍵字的結果，必須被清掉。
    listCompaniesSpy.mockResolvedValue({
      companies: [{ id: "c-1", name: "關鍵字命中公司" }],
      pagination: { total: 1 },
    });
    const searchForm = document.getElementById("company-search-keyword")?.closest("form");
    if (!searchForm) throw new Error("找不到公司搜尋表單");
    fireEvent.input(within(searchForm).getByLabelText("公司關鍵字"), {
      target: { value: "命中" },
    });
    fireEvent.submit(searchForm);

    await waitFor(() => expect(companyOptionIds()).toEqual(["c-1"]));
  });
});

describe("<DepartmentsPage> 表格（TanStack Table，manual 分頁）", () => {
  const PAGE_SIZE = 20;

  /** 第 n 筆部門（n 從 1 起算）：名稱帶序號，用來辨識畫面拿到的是哪一頁的資料。 */
  function department(n: number) {
    return {
      id: `d-${n}`,
      name: `部門 ${n}`,
      companyId: "c-1",
      companyName: "既有公司",
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
    listDepartmentsSpy.mockImplementation((req: { page: number }) =>
      Promise.resolve({
        departments: Array.from(
          { length: Math.max(0, Math.min(PAGE_SIZE, total - (req.page - 1) * PAGE_SIZE)) },
          (_, i) => department((req.page - 1) * PAGE_SIZE + i + 1)
        ),
        pagination: { total },
      })
    );
  }

  /** 掛載頁面並等第 1 頁的資料上畫面。 */
  async function renderPage1() {
    mockPages(45);
    mountPage();
    await waitFor(() => expect(screen.getByText("部門 1")).toBeTruthy());
  }

  it("表頭與 columns 一致：四欄（部門名稱/所屬公司/ID/操作），資料列的格子數相同", async () => {
    await renderPage1();

    const headers = headerCells();
    expect(headers.map((h) => h.textContent?.trim())).toEqual([
      "部門名稱",
      "所屬公司",
      "ID",
      "操作",
    ]);
    expect(tableRows()[0].querySelectorAll("td")).toHaveLength(headers.length);
  });

  it("列數等於資料數（順序照查詢結果）", async () => {
    listDepartmentsSpy.mockResolvedValue({
      departments: [department(1), department(2), department(3)],
      pagination: { total: 3 },
    });
    mountPage();
    await waitFor(() => expect(screen.getByText("部門 3")).toBeTruthy());

    const rows = tableRows();
    expect(rows).toHaveLength(3);
    expect(rows.map((r) => r.querySelector("td")?.textContent)).toEqual([
      "部門 1",
      "部門 2",
      "部門 3",
    ]);
  });

  it("載入列與空狀態列的 colspan 由欄位數推導（兩者都等於表頭欄位數）", async () => {
    const pending = Promise.withResolvers<unknown>();
    listDepartmentsSpy.mockReturnValue(pending.promise);
    mountPage();

    const columnCount = headerCells().length;
    expect(columnCount).toBe(4);

    const loadingRow = (await waitFor(() => screen.getByText("載入中…"))).closest("tr");
    expect(loadingRow?.querySelector("td")?.getAttribute("colspan")).toBe(String(columnCount));
    expect(loadingRow?.querySelectorAll("td")).toHaveLength(1);

    pending.resolve({ departments: [], pagination: { total: 0 } });

    const emptyRow = (await waitFor(() => screen.getByText("尚無部門資料"))).closest("tr");
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
    await waitFor(() => expect(screen.getByText("部門 41")).toBeTruthy());

    expect(screen.getByText(/第 41–45 筆,共 45 筆/)).toBeTruthy();
    expect(screen.getByRole("button", { name: "第 3 頁" }).getAttribute("aria-current")).toBe(
      "page"
    );
  });

  it("點「下一頁」：以 page 2 查詢，且第 2 頁只渲染它自己的列（伺服器已分頁，table 不再切一次）", async () => {
    await renderPage1();

    fireEvent.click(screen.getByRole("button", { name: "下一頁" }));

    await waitFor(() =>
      expect(listDepartmentsSpy).toHaveBeenLastCalledWith({
        page: 2,
        pageSize: 20,
        sort: "",
        desc: false,
        companyId: undefined,
      })
    );
    await waitFor(() => expect(screen.getByText("部門 21")).toBeTruthy());

    // 第 2 頁回 20 筆 → 全部渲染（若 table 又依 pageIndex 切一次，這裡會是 0 列）；
    // 頁碼也不得被 table 的 autoResetPageIndex 打回第 1 頁。
    expect(tableRows()).toHaveLength(20);
    expect(screen.getByRole("button", { name: "第 2 頁" }).getAttribute("aria-current")).toBe(
      "page"
    );
  });

  it("clamp 守衛：換頁失敗（沒有屬於當前查詢的資料）時不得改寫頁碼，也不得多發請求", async () => {
    const failure = Promise.withResolvers<unknown>();
    listDepartmentsSpy.mockResolvedValueOnce({
      departments: [department(1)],
      pagination: { total: 45 },
    });
    listDepartmentsSpy.mockReturnValueOnce(failure.promise);
    mountPage();
    await waitFor(() => expect(screen.getByText("部門 1")).toBeTruthy());

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));
    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(2));
    failure.reject(new ConnectError("伺服器暫時無法使用", Code.Internal));

    await waitFor(() => expect(screen.getByRole("alert").textContent).toBe("伺服器暫時無法使用"));

    // 錯誤沒有 placeholder 可保留 → `query.data` 為 undefined，`total()` 會算成 0；
    // 若據以夾頁碼就會多打一次 page 1 的請求（移除守衛時本斷言變紅）。
    await settle();
    expect(listDepartmentsSpy).toHaveBeenCalledTimes(2);
  });

  it("編輯成功後重取的資料會換到畫面上（row model 不得黏住舊列）", async () => {
    updateDepartmentSpy.mockResolvedValue({});
    listDepartmentsSpy.mockResolvedValueOnce({
      departments: [department(1)],
      pagination: { total: 1 },
    });
    mountPage();
    await waitFor(() => expect(screen.getByText("部門 1")).toBeTruthy());

    // 重取回同一頁但內容已變（名稱改了）：table 的 row model 必須跟著換。
    listDepartmentsSpy.mockResolvedValueOnce({
      departments: [{ ...department(1), name: "部門 1（改名後）" }],
      pagination: { total: 1 },
    });

    fireEvent.click(screen.getByRole("button", { name: "編輯" }));
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    fireEvent.input(within(dialog).getByLabelText(/部門名稱/), {
      target: { value: "部門 1（改名後）" },
    });
    fireEvent.submit(dialog.querySelector("form")!);

    await waitFor(() => expect(screen.getByText("部門 1（改名後）")).toBeTruthy());
    expect(screen.queryByText("部門 1")).toBeNull();
  });
});

describe("<DepartmentsPage> 表頭排序（伺服器端）", () => {
  const PAGE_SIZE = 20;

  /** 第 n 筆部門（n 從 1 起算）：名稱帶序號，用來辨識畫面拿到的是哪一頁的資料。 */
  function department(n: number) {
    return {
      id: `d-${n}`,
      name: `部門 ${n}`,
      companyId: "c-1",
      companyName: "既有公司",
    };
  }

  /** 伺服器端分頁的替身（45 筆＝3 頁）；排序由後端負責，替身不回傳已排序的資料。 */
  function mockPages(total = 45) {
    listDepartmentsSpy.mockImplementation((req: { page: number }) =>
      Promise.resolve({
        departments: Array.from(
          { length: Math.max(0, Math.min(PAGE_SIZE, total - (req.page - 1) * PAGE_SIZE)) },
          (_, i) => department((req.page - 1) * PAGE_SIZE + i + 1)
        ),
        pagination: { total },
      })
    );
  }

  /** 掛載頁面並等第 1 頁的資料上畫面。 */
  async function renderPage1() {
    mockPages();
    mountPage();
    await waitFor(() => expect(screen.getByText("部門 1")).toBeTruthy());
  }

  /** 表頭儲存格：以可及名稱定位（方向指示圖示 `aria-hidden`，不進名稱）。 */
  function headerCell(label: string): HTMLTableCellElement {
    return screen.getByRole("columnheader", { name: label }) as HTMLTableCellElement;
  }

  /** 表頭上的排序控制項（可及名稱＝欄位名，不含方向指示）。 */
  function sortButton(label: string): HTMLButtonElement {
    return within(headerCell(label)).getByRole("button", { name: label }) as HTMLButtonElement;
  }

  /** 帶方向語意的表頭（`aria-sort` 非 none）：ARIA 下一張表最多只能有一個排序列。 */
  function sortedHeaderCells(): HTMLTableCellElement[] {
    return screen.getAllByRole("columnheader").filter((cell) => {
      const value = cell.getAttribute("aria-sort");
      return value !== null && value !== "none";
    }) as HTMLTableCellElement[];
  }

  /** 清單請求的完整 payload（排序波次後 query key 一律帶 sort/desc）。 */
  const payload = (page: number, sort: string, desc: boolean) => ({
    page,
    pageSize: PAGE_SIZE,
    companyId: undefined,
    sort,
    desc,
  });

  it("點 name 表頭：請求帶 sort:name、desc:false，且從第 2 頁回到第 1 頁（同一次更新只查一次）", async () => {
    await renderPage1();

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));
    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(2));

    fireEvent.click(sortButton("部門名稱"));

    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(3));
    expect(listDepartmentsSpy).toHaveBeenLastCalledWith(payload(1, "name", false));
    // 排序與回第 1 頁必須同批：分開寫會先以「舊頁碼＋新排序」多打一次。
    await settle();
    expect(listDepartmentsSpy).toHaveBeenCalledTimes(3);
  });

  it("再點同欄：desc 反轉為 true", async () => {
    await renderPage1();

    fireEvent.click(sortButton("部門名稱"));
    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(2));

    fireEvent.click(sortButton("部門名稱"));

    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(3));
    expect(listDepartmentsSpy).toHaveBeenLastCalledWith(payload(1, "name", true));
  });

  it("點另一欄：改用新欄且 desc 從 false 起算", async () => {
    await renderPage1();

    fireEvent.click(sortButton("部門名稱"));
    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(2));

    fireEvent.click(sortButton("ID"));

    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(3));
    expect(listDepartmentsSpy).toHaveBeenLastCalledWith(payload(1, "id", false));
    expect(headerCell("部門名稱").getAttribute("aria-sort")).toBe("none");
  });

  it("aria-sort 掛在 <th> 且三態正確；控制項是可聚焦按鈕，排序後焦點不亂跳", async () => {
    await renderPage1();

    // 初始（未排序）：可排序欄位皆為 none。
    expect(headerCell("部門名稱").getAttribute("aria-sort")).toBe("none");
    expect(headerCell("ID").getAttribute("aria-sort")).toBe("none");

    const button = sortButton("部門名稱");
    expect(button.tagName).toBe("BUTTON");
    expect(button.getAttribute("type")).toBe("button");
    expect(button.tabIndex).toBe(0);
    button.focus();
    expect(document.activeElement).toBe(button);

    fireEvent.click(button);
    await waitFor(() => expect(headerCell("部門名稱").getAttribute("aria-sort")).toBe("ascending"));
    // 方向只掛在被排的那一欄。
    expect(headerCell("ID").getAttribute("aria-sort")).toBe("none");

    fireEvent.click(sortButton("部門名稱"));
    await waitFor(() =>
      expect(headerCell("部門名稱").getAttribute("aria-sort")).toBe("descending")
    );
    // 排序變更不得重建表頭控制項：焦點留在同一顆按鈕上。
    expect(document.activeElement).toBe(button);
    expect(sortButton("部門名稱")).toBe(button);
  });

  it("方向指示器：升冪顯示 ▲、降冪顯示 ▼、未排序兩者皆無（同一顆表頭按鈕上更新）", async () => {
    await renderPage1();

    const button = sortButton("部門名稱");
    // 未排序：兩個方向都不出現（顯示任一箭頭都是誤導）。
    expect(button.textContent).not.toContain("▲");
    expect(button.textContent).not.toContain("▼");

    fireEvent.click(button);
    await waitFor(() => expect(headerCell("部門名稱").getAttribute("aria-sort")).toBe("ascending"));
    await waitFor(() => expect(button.textContent).toContain("▲"));
    expect(button.textContent).not.toContain("▼");

    fireEvent.click(sortButton("部門名稱"));
    await waitFor(() =>
      expect(headerCell("部門名稱").getAttribute("aria-sort")).toBe("descending")
    );
    await waitFor(() => expect(button.textContent).toContain("▼"));
    expect(button.textContent).not.toContain("▲");
  });

  it("可排序欄位僅限白名單：所屬公司與操作欄沒有排序控制項，也不帶 aria-sort", async () => {
    await renderPage1();

    for (const label of ["所屬公司", "操作"]) {
      const cell = headerCell(label);
      expect(cell.querySelector("button")).toBeNull();
      // 不可排序的欄位不帶 `aria-sort`（`none` 只代表「可排序但未排」）。
      expect(cell.getAttribute("aria-sort")).toBeNull();
    }
    for (const label of ["部門名稱", "ID"]) {
      expect(sortButton(label).tagName).toBe("BUTTON");
      expect(headerCell(label).getAttribute("aria-sort")).toBe("none");
    }
  });

  it("換頁後排序保持：請求仍帶同一組 sort/desc", async () => {
    await renderPage1();

    fireEvent.click(sortButton("部門名稱"));
    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(2));

    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));

    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(3));
    expect(listDepartmentsSpy).toHaveBeenLastCalledWith(payload(2, "name", false));
    await waitFor(() => expect(screen.getByText("部門 21")).toBeTruthy());
    expect(headerCell("部門名稱").getAttribute("aria-sort")).toBe("ascending");
  });

  it("shift+click 第二欄：多欄排序已停用，只有一個排序列（aria-sort 與請求一致、未排欄位不得出現箭頭）", async () => {
    await renderPage1();

    fireEvent.click(sortButton("部門名稱"));
    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(2));

    // v9 預設把 shift+click 當成「加入多欄排序」（第二欄 append 進 sorting state），但三頁只送
    // `sorting()[0]` → 第二欄會出現假的 aria-sort 與箭頭，而請求完全不變。`enableMultiSort: false`
    // 讓它退化成一般的單欄取代排序（請求與顯示一致）。
    fireEvent.click(sortButton("ID"), { shiftKey: true });

    // ① 只有一組排序，且只發一次請求：新欄從 desc:false 起算。
    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(3));
    expect(listDepartmentsSpy).toHaveBeenLastCalledWith(payload(1, "id", false));
    await settle();
    expect(listDepartmentsSpy).toHaveBeenCalledTimes(3);

    // ② 只有一個表頭帶方向語意（舊欄回 none）。
    expect(sortedHeaderCells()).toHaveLength(1);
    expect(headerCell("ID").getAttribute("aria-sort")).toBe("ascending");
    expect(headerCell("部門名稱").getAttribute("aria-sort")).toBe("none");

    // ③ 未排的那一欄沒有方向指示器。
    expect(sortButton("ID").textContent).toContain("▲");
    expect(sortButton("部門名稱").textContent).not.toContain("▲");
    expect(sortButton("部門名稱").textContent).not.toContain("▼");
  });

  it("第三次點同一欄＝取消排序：請求回 sort 空字串、aria-sort 回 none、指示器消失，且回第 1 頁", async () => {
    await renderPage1();

    fireEvent.click(sortButton("部門名稱"));
    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(2));
    fireEvent.click(sortButton("部門名稱"));
    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(3));
    expect(listDepartmentsSpy).toHaveBeenLastCalledWith(payload(1, "name", true));

    // 先離開第 1 頁，才能驗「取消排序也回第 1 頁」（否則斷言的 page:1 沒有鑑別力）。
    fireEvent.click(screen.getByRole("button", { name: "第 2 頁" }));
    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(4));

    fireEvent.click(sortButton("部門名稱"));

    // D1：`sort` 空 → 服務預設排序（`desc` 被忽略，前端一律送 false）。
    await waitFor(() => expect(listDepartmentsSpy).toHaveBeenCalledTimes(5));
    expect(listDepartmentsSpy).toHaveBeenLastCalledWith(payload(1, "", false));
    await settle();
    expect(listDepartmentsSpy).toHaveBeenCalledTimes(5);

    expect(headerCell("部門名稱").getAttribute("aria-sort")).toBe("none");
    expect(sortedHeaderCells()).toHaveLength(0);
    expect(sortButton("部門名稱").textContent).not.toContain("▲");
    expect(sortButton("部門名稱").textContent).not.toContain("▼");
  });
});
