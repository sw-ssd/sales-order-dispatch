import { buttonVariants } from "~/components/ui/button";

/**
 * 403 頁(/403)：Tailkit Errors 版面(a-p-errors-01)——
 * 卡片底 + 左右斜角色帶(品牌淺底)、大號狀態碼、分隔短槓、標題／說明、回首頁。
 * 顏色一律語意 token;文案與導向(<a href="/">)與改版前相同。
 */
export default function ForbiddenPage() {
  return (
    <main class="relative flex min-h-dvh items-center overflow-hidden bg-card text-card-foreground">
      <div
        class="absolute top-0 bottom-0 left-0 -ml-44 w-48 bg-primary/10 md:-ml-28 md:skew-x-6"
        aria-hidden="true"
      />
      <div
        class="absolute top-0 right-0 bottom-0 -mr-44 w-48 bg-primary/10 md:-mr-28 md:skew-x-6"
        aria-hidden="true"
      />

      <div class="relative mx-auto w-full max-w-lg space-y-6 px-8 py-16 text-center">
        <p class="text-6xl font-extrabold text-primary md:text-7xl">403</p>
        <div class="mx-auto h-1.5 w-12 rounded-lg bg-border" aria-hidden="true" />
        <h1 class="text-2xl font-extrabold text-foreground md:text-3xl">無權限存取</h1>
        <p class="font-medium text-muted-foreground">你的角色不允許存取此頁面</p>
        <a href="/" class={buttonVariants()}>
          回首頁
        </a>
      </div>
    </main>
  );
}
