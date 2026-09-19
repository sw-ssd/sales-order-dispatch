import { For, Show } from "solid-js";

/**
 * `/ui` 元件庫展示頁（dev-only）。
 *
 * 路由只在開發環境註冊（見 `router/index.tsx`），這裡再以 `import.meta.env.DEV`
 * 把關一次：正式站即使用手動輸入網址也拿不到展示內容。
 * 本 task 只做骨架（元件清單），各元件的變體與狀態展示由後續 task 補齊。
 */
const componentNames = [
  "badge",
  "button",
  "card",
  "checkbox",
  "dialog",
  "field",
  "input",
  "label",
  "pagination",
  "scroll-area",
  "spinner",
  "table",
  "tabs",
];

export default function UiDemoPage() {
  return (
    <Show
      when={import.meta.env.DEV}
      fallback={
        <main class="p-8">
          <h1 class="text-xl font-semibold text-foreground">元件庫展示</h1>
          <p class="mt-2 text-muted-foreground">此頁僅在開發環境提供。</p>
        </main>
      }
    >
      <main class="min-h-screen bg-background p-8 text-foreground">
        <h1 class="text-2xl font-bold">元件庫展示</h1>
        <p class="mt-2 text-muted-foreground">
          僅開發環境可用；各元件的變體與狀態展示由後續 task 補齊。
        </p>
        <ul class="mt-6 grid grid-cols-2 gap-2 text-sm sm:grid-cols-3 lg:grid-cols-4">
          <For each={componentNames}>
            {(name) => (
              <li class="rounded-lg border border-border bg-card px-3 py-2">{name}</li>
            )}
          </For>
        </ul>
      </main>
    </Show>
  );
}
