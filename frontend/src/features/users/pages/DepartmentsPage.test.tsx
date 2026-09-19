import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";

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

/** 渲染頁面並等列表載入完成（modal 的測試都要先有列表可點）。 */
async function renderPage() {
  render(() => <DepartmentsPage />);
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
