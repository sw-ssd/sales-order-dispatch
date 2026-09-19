import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { useNavigate } from "@tanstack/solid-router";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "~/components/ui/tabs";
import { createSignal, Show, type JSX } from "solid-js";
import { AuthService } from "~/lib/proto/salesorder/v1/auth_pb";
import { Button } from "~/components/ui/button";
import { Field, FieldError, FieldLabel } from "~/components/ui/field";
import { Input } from "~/components/ui/input";
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

/**
 * 登入頁(/login)：Tailkit Boxed Sign In 版面(a-p-sign-in-01)——
 * 頁底 `bg-muted`、置中單欄卡片、頁首標題＋副標。
 * 卡片內沿用 T6 的 Tabs;顏色一律語意 token(不含 Tailkit 的色階字面值與深色變體)。
 */
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
      navigate({ to: "/", replace: true });
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <main class="flex min-h-dvh w-full flex-col items-center justify-center bg-muted p-4 lg:p-8">
      <section class="w-full max-w-lg py-6">
        <header class="mb-8 text-center">
          <h1 class="text-2xl font-bold text-foreground">登入</h1>
          <p class="mt-2 text-sm font-medium text-muted-foreground">多公司訂出貨系統</p>
        </header>

        <div class="overflow-hidden rounded-lg border border-border bg-card text-card-foreground shadow-xs">
          <div class="p-5 md:px-12 md:py-10">
            <Tabs
              value={tab()}
              onValueChange={(value) => setTab(value as LoginTab)}
            >
              <TabsList class="h-auto w-full">
                <TabsTrigger value="employee" class="flex-1 py-2">
                  員工
                </TabsTrigger>
                <TabsTrigger value="store" class="flex-1 py-2">
                  店家
                </TabsTrigger>
              </TabsList>

              <TabsContent value="employee" class="mt-5">
                <p class="mb-4 text-center text-sm text-muted-foreground">
                  員工請使用公司 Google 帳號登入
                </p>
                <GoogleLoginButton />
              </TabsContent>

              <TabsContent value="store" class="mt-5">
                <form class="space-y-6" onSubmit={handleStoreSubmit}>
                  <Field>
                    <FieldLabel for="customer_code">客戶編號</FieldLabel>
                    <Input
                      id="customer_code"
                      name="customer_code"
                      required
                      autocomplete="username"
                      value={customerCode()}
                      onInput={(e) => setCustomerCode(e.currentTarget.value)}
                    />
                  </Field>
                  <Field>
                    <FieldLabel for="password">密碼</FieldLabel>
                    <Input
                      id="password"
                      name="password"
                      type="password"
                      required
                      autocomplete="current-password"
                      value={password()}
                      onInput={(e) => setPassword(e.currentTarget.value)}
                    />
                  </Field>

                  <Show when={error()}>
                    {(message) => <FieldError>{message()}</FieldError>}
                  </Show>

                  <Button type="submit" class="w-full" loading={submitting()}>
                    登入
                  </Button>
                </form>
              </TabsContent>
            </Tabs>
          </div>
        </div>
      </section>
    </main>
  );
}
