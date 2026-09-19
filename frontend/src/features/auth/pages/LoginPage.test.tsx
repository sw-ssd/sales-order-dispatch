import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import type * as ConnectRpc from "@connectrpc/connect";

// 登入 API 與導向都以 spy 取代：LoginPage 在模組層建立 connect client，
// 因此以 createClient 的替身攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const { loginSpy, navigateSpy } = vi.hoisted(() => ({
  loginSpy: vi.fn(),
  navigateSpy: vi.fn(),
}));

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({ login: loginSpy }),
}));

vi.mock("@tanstack/solid-router", () => ({ useNavigate: () => navigateSpy }));

import LoginPage from "./LoginPage";

/** 切到「店家」分頁並回傳該分頁的 form（Tabs 為 lazyMount + unmountOnExit，未選中的 panel 不在 DOM）。 */
async function openStoreForm() {
  const { container } = render(() => <LoginPage />);
  fireEvent.click(screen.getByRole("tab", { name: "店家" }));
  // Ark 的狀態機與 panel 掛載要等下一個 microtask 才回報，互動後的斷言一律用真實等待。
  return waitFor(() => {
    const form = container.querySelector("form");
    if (!form) throw new Error("店家分頁尚未渲染登入表單");
    return form;
  });
}

function fillCredentials(customerCode: string, password: string) {
  fireEvent.input(screen.getByLabelText("客戶編號"), { target: { value: customerCode } });
  fireEvent.input(screen.getByLabelText("密碼"), { target: { value: password } });
}

beforeEach(() => {
  loginSpy.mockReset();
  navigateSpy.mockReset();
});

describe("<LoginPage> 店家分頁表單", () => {
  it("未填直接提交：兩欄各自出現繁中必填訊息，aria 關聯指向該錯誤元素，且不呼叫登入", async () => {
    const form = await openStoreForm();

    fireEvent.submit(form);

    await waitFor(() => expect(screen.getByText("請輸入客戶編號")).toBeTruthy());
    expect(screen.getByText("請輸入密碼")).toBeTruthy();

    for (const [label, message] of [
      ["客戶編號", "請輸入客戶編號"],
      ["密碼", "請輸入密碼"],
    ] as const) {
      const input = screen.getByLabelText(label);
      expect(input.getAttribute("aria-invalid")).toBe("true");
      const describedBy = input.getAttribute("aria-describedby");
      expect(describedBy).toBeTruthy();
      expect(document.getElementById(describedBy!)?.textContent).toContain(message);
    }

    expect(loginSpy).not.toHaveBeenCalled();
    expect(navigateSpy).not.toHaveBeenCalled();
  });

  it("驗證時機：輸入過程不標紅，blur 後才出現該欄錯誤", async () => {
    await openStoreForm();
    const customerCode = screen.getByLabelText("客戶編號");

    fireEvent.input(customerCode, { target: { value: "S" } });
    fireEvent.input(customerCode, { target: { value: "" } });
    expect(screen.queryByText("請輸入客戶編號")).toBeNull();

    fireEvent.blur(customerCode);
    await waitFor(() => expect(screen.getByText("請輸入客戶編號")).toBeTruthy());
  });

  it("填入合法值：錯誤消失，並以正確 payload 登入後導向首頁", async () => {
    loginSpy.mockResolvedValue(undefined);
    const form = await openStoreForm();

    fireEvent.submit(form);
    await waitFor(() => expect(screen.getByText("請輸入客戶編號")).toBeTruthy());

    fillCredentials("S-001", "pw-12345");
    fireEvent.submit(form);

    await waitFor(() =>
      expect(loginSpy).toHaveBeenCalledWith({ customerCode: "S-001", password: "pw-12345" })
    );
    await waitFor(() => expect(navigateSpy).toHaveBeenCalledWith({ to: "/", replace: true }));
    expect(screen.queryByText("請輸入客戶編號")).toBeNull();
    expect(screen.getByLabelText("客戶編號").getAttribute("aria-invalid")).toBeNull();
  });

  it("伺服器錯誤：以 role=alert 的表單層 banner 呈現，且不導向", async () => {
    loginSpy.mockRejectedValue(new ConnectError("invalid credentials", Code.Unauthenticated));
    const form = await openStoreForm();

    fillCredentials("S-001", "bad-password");
    fireEvent.submit(form);

    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain("客戶編號或密碼錯誤")
    );
    expect(navigateSpy).not.toHaveBeenCalled();
  });

  it("提交中由表單的 isSubmitting 驅動按鈕載入狀態", async () => {
    const login = Promise.withResolvers<void>();
    loginSpy.mockReturnValue(login.promise);
    const form = await openStoreForm();
    const submitButton = within(form).getByRole("button", {
      name: /登入/,
    }) as HTMLButtonElement;

    fillCredentials("S-001", "pw-12345");
    fireEvent.submit(form);

    await waitFor(() => expect(loginSpy).toHaveBeenCalledTimes(1));
    expect(submitButton.disabled).toBe(true);
    expect(submitButton.getAttribute("aria-busy")).toBe("true");

    login.resolve();
    await waitFor(() => expect(submitButton.disabled).toBe(false));
  });
});
