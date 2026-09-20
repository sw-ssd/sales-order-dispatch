import { buttonVariants } from "@ui/button";
import { loginUrl } from "../lib/api";

/**
 * 登入頁：console 不自己收帳密——operator 一律走後端 OIDC（`/platform/auth/google`），
 * 成功後由後端設 HttpOnly cookie 並導回 console 根路徑。
 */
export default function LoginPage() {
  return (
    <main class="mx-auto max-w-md space-y-4 p-6">
      <h1 class="text-2xl font-bold">平台營運主控台</h1>
      <p class="text-sm text-muted-foreground">
        此工具需要 operator 身分：公司網域 OIDC 登入 ＋ 在 <code>platform.operators</code> 白名單內。
      </p>
      <a href={loginUrl} class={buttonVariants({ size: "lg" })}>
        以 Google 登入
      </a>
      <p class="text-xs text-muted-foreground">
        登入後由後端設定 cookie 並導回主控台。此頁的導向只是 UX：每一個平台 RPC 都由後端
        重新驗證 operator 身分與白名單，前端的呈現不構成授權。
      </p>
    </main>
  );
}
