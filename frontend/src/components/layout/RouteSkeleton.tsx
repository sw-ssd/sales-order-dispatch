import { For } from "solid-js";

/**
 * 路由層 Suspense 的 fallback（`router.options.defaultPendingComponent`）。
 *
 * 為什麼需要它：`@tanstack/solid-router` 在 client 把每條路由包進一個 `Solid.Suspense`
 * （`Matches` 的頂層邊界），而 `createQuery().data` 是 Solid resource —— **新 query key 首次
 * 載入**時讀 `.data` 會 suspend（`useBaseQuery` 的 proxy 只在該 key 已有值時改讀
 * `resource.latest`）。沒有 fallback 就是 `null`：整塊路由內容區在載入期間直接消失
 * （側邊欄還在，因為它在 `<Outlet/>` 之外）。
 *
 * 這裡用骨架而非 spinner（`perf-skeleton-suspense`）：形狀照頁面的「標題列 + 篩選卡 + 表格卡」
 * 排，載入完成時不做版面跳動。
 */
export default function RouteSkeleton() {
  return (
    <main aria-busy="true" aria-label="頁面載入中">
      <div class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4">
        <div class="h-7 w-48 animate-pulse rounded bg-muted" />
        <div class="h-4 w-32 animate-pulse rounded bg-muted" />
      </div>
      <div class="mb-4 h-20 animate-pulse rounded-lg border border-border bg-muted/50" />
      <div class="rounded-lg border border-border bg-card p-4">
        <For each={Array.from({ length: 6 }, (_, i) => i)}>
          {() => <div class="mb-3 h-5 animate-pulse rounded bg-muted last:mb-0" />}
        </For>
      </div>
    </main>
  );
}
