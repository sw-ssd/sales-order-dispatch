import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { useNavigate } from "@solidjs/router";
import { createSignal, Show, type JSX } from "solid-js";
import { AuthService } from "~/lib/proto/salesorder/v1/auth_pb";
import GoogleLoginButton from "../components/GoogleLoginButton";

const authClient = createClient(
  AuthService,
  createConnectTransport({ baseUrl: "/api/v1" }),
);

type LoginTab = "employee" | "store";

function errorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    if (err.code === Code.Unimplemented) {
      return "後端尚未實作登入功能,請稍後再試";
    }
    if (err.code === Code.Unauthenticated) {
      return "客戶編號或密碼錯誤";
    }
    if (err.code === Code.Unavailable || err.code === Code.Unknown) {
      // 後端未啟動時,vite proxy 會回 HTTP 500,connect-web 對應 Code.Unknown
      return "無法連線到伺服器,請確認後端服務已啟動";
    }
    return `登入失敗:${err.message}`;
  }
  return "無法連線到伺服器,請確認後端服務已啟動";
}

export default function LoginPage() {
  const navigate = useNavigate();
  const [tab, setTab] = createSignal<LoginTab>("employee");
  const [customerCode, setCustomerCode] = createSignal("");
  const [password, setPassword] = createSignal("");
  const [submitting, setSubmitting] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);

  const handleStoreSubmit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = async (event) => {
    event.preventDefault();
    if (submitting()) return;
    setSubmitting(true);
    setError(null);
    try {
      await authClient.login({
        customerCode: customerCode(),
        password: password(),
      });
      navigate("/", { replace: true });
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <main class="flex min-h-screen items-center justify-center bg-gray-50 p-4">
      <div class="w-full max-w-sm rounded-lg bg-white p-6 shadow">
        <h1 class="text-center text-2xl font-bold text-gray-900">登入</h1>
        <p class="mt-1 text-center text-sm text-gray-500">多公司訂出貨系統</p>

        <div class="mt-6 flex gap-1 rounded-lg bg-gray-100 p-1" role="tablist">
          <button
            type="button"
            role="tab"
            aria-selected={tab() === "employee"}
            class={`flex-1 rounded-md px-4 py-2 text-sm font-medium ${
              tab() === "employee" ? "bg-white text-gray-900 shadow" : "text-gray-500 hover:text-gray-700"
            }`}
            onClick={() => setTab("employee")}
          >
            員工
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={tab() === "store"}
            class={`flex-1 rounded-md px-4 py-2 text-sm font-medium ${
              tab() === "store" ? "bg-white text-gray-900 shadow" : "text-gray-500 hover:text-gray-700"
            }`}
            onClick={() => setTab("store")}
          >
            店家
          </button>
        </div>

        <Show when={error()}>
          {(message) => (
            <p class="mt-4 rounded-md bg-red-50 px-3 py-2 text-sm text-red-700" role="alert">
              {message()}
            </p>
          )}
        </Show>

        <Show when={tab() === "employee"}>
          <div class="mt-6">
            <p class="mb-3 text-center text-sm text-gray-500">員工請使用公司 Google 帳號登入</p>
            <GoogleLoginButton />
          </div>
        </Show>

        <Show when={tab() === "store"}>
          <form class="mt-6 space-y-4" onSubmit={handleStoreSubmit}>
            <div>
              <label for="customer_code" class="block text-sm font-medium text-gray-700">
                客戶編號
              </label>
              <input
                id="customer_code"
                name="customer_code"
                type="text"
                required
                autocomplete="username"
                value={customerCode()}
                onInput={(e) => setCustomerCode(e.currentTarget.value)}
                class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
              />
            </div>
            <div>
              <label for="password" class="block text-sm font-medium text-gray-700">
                密碼
              </label>
              <input
                id="password"
                name="password"
                type="password"
                required
                autocomplete="current-password"
                value={password()}
                onInput={(e) => setPassword(e.currentTarget.value)}
                class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
              />
            </div>
            <button
              type="submit"
              disabled={submitting()}
              class="w-full rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {submitting() ? "登入中…" : "登入"}
            </button>
          </form>
        </Show>
      </div>
    </main>
  );
}
