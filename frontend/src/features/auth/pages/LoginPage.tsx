import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { createForm } from "@tanstack/solid-form";
import { useNavigate } from "@tanstack/solid-router";
import {
  Button,
  Field,
  FieldError,
  FieldLabel,
  Input,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "~/components/ui";
import { createSignal, Show, type JSX } from "solid-js";
import { AuthService } from "~/lib/proto/salesorder/v1/auth_pb";
import GoogleLoginButton from "../components/GoogleLoginButton";
import { fieldValidators, loginSchema } from "../schemas";

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
 * 欄位錯誤訊息只取第一則；`meta.errors` 的型別是 `unknown[]`，
 * 而 validator（`fieldValidators`）回傳的必定是字串，非字串一律不顯示。
 */
function firstMessage(errors: unknown[]): string | undefined {
  const [first] = errors;
  return typeof first === "string" ? first : undefined;
}

/**
 * 登入頁(/login)：Tailkit Boxed Sign In 版面(a-p-sign-in-01)——
 * 頁底 `bg-muted`、置中單欄卡片、頁首標題＋副標。
 * 卡片內沿用 T6 的 Tabs;顏色一律語意 token(不含 Tailkit 的色階字面值與深色變體)。
 *
 * 「店家」表單的欄位值、欄位驗證與提交狀態由 `createForm` 持有：
 * 驗證時機為 `onBlur` + `onSubmit`（輸入過程不標紅），客戶端錯誤落在該欄下方；
 * 伺服器錯誤（登入失敗）不對應特定欄位，由提交流程設進表單層 banner。
 */
export default function LoginPage() {
  const navigate = useNavigate();
  const [tab, setTab] = createSignal<LoginTab>("employee");
  // 伺服器錯誤不是驗證狀態（不對應任何欄位），由提交流程設定，落在表單層 banner。
  const [serverError, setServerError] = createSignal<string | undefined>();

  const customerCodeValidators = fieldValidators(loginSchema.entries.customerCode);
  const passwordValidators = fieldValidators(loginSchema.entries.password);

  const form = createForm(() => ({
    defaultValues: { customerCode: "", password: "" },
    onSubmit: async ({ value }) => {
      try {
        await authClient.login(value);
        navigate({ to: "/", replace: true });
      } catch (err) {
        setServerError(errorMessage(err));
      }
    },
  }));

  const isSubmitting = form.useSelector((state) => state.isSubmitting);

  const handleStoreSubmit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (event) => {
    event.preventDefault();
    // 客戶端驗證失敗時 `onSubmit` 不會被呼叫，舊的伺服器錯誤 banner 必須在這裡先清掉。
    setServerError(undefined);
    void form.handleSubmit();
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
                {/*
                  `novalidate`：必填規則已由 valibot 鏡射，原生驗證會用瀏覽器泡泡擋下 submit
                  並讓自訂的繁中欄位錯誤沒有機會顯示；`required` 保留作為必填的語意標記。
                */}
                <form class="space-y-6" novalidate onSubmit={handleStoreSubmit}>
                  <form.Field name="customerCode" validators={customerCodeValidators}>
                    {(field) => (
                      <Field invalid={!field().state.meta.isValid}>
                        <FieldLabel for="customer_code">客戶編號</FieldLabel>
                        <Input
                          id="customer_code"
                          name="customer_code"
                          required
                          autocomplete="username"
                          value={field().state.value}
                          onBlur={field().handleBlur}
                          onInput={(e) => field().handleChange(e.currentTarget.value)}
                        />
                        <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                      </Field>
                    )}
                  </form.Field>

                  <form.Field name="password" validators={passwordValidators}>
                    {(field) => (
                      <Field invalid={!field().state.meta.isValid}>
                        <FieldLabel for="password">密碼</FieldLabel>
                        <Input
                          id="password"
                          name="password"
                          type="password"
                          required
                          autocomplete="current-password"
                          value={field().state.value}
                          onBlur={field().handleBlur}
                          onInput={(e) => field().handleChange(e.currentTarget.value)}
                        />
                        <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                      </Field>
                    )}
                  </form.Field>

                  <Show when={serverError()}>
                    {(message) => <FieldError>{message()}</FieldError>}
                  </Show>

                  <Button type="submit" class="w-full" loading={isSubmitting()}>
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
