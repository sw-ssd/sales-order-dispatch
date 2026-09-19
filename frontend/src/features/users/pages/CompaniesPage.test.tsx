import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";

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

/** 渲染頁面並等列表載入完成（modal 的測試都要先有列表可點）。 */
async function renderPage() {
  render(() => <CompaniesPage />);
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
