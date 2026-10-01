import { render } from "@solidjs/testing-library";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import type { JSX } from "solid-js";
import { ConfirmProvider } from "@ui";

/**
 * 測試專用的 provider 疊法：`QueryClientProvider` 之外再包 `ConfirmProvider`。
 *
 * 為什麼需要：破壞性操作改用 `useConfirm()`（見 `components/ui/confirm.tsx`），
 * 它不在 provider 之內會直接拋錯、不靜默放行；而 App 的 provider 疊法（main → App）
 * 在測試裡不存在，每個測試檔各自手寫一層只會漏掉 `ConfirmProvider`。
 *
 * 疊序與正式路徑一致：QueryClient 在外，Confirm 在內。
 */
export function withProviders(
  client: QueryClient,
  children: () => JSX.Element,
): JSX.Element {
  return (
    <QueryClientProvider client={client}>
      <ConfirmProvider>{children()}</ConfirmProvider>
    </QueryClientProvider>
  );
}

/**
 * `withProviders` ＋ `render`，回傳實際使用的 client（呼叫端要 spy `invalidateQueries` 時用得上）。
 *
 * 未指定 client 時每呼叫一次配一份新的（`retry: false`）：快取不跨測試殘留，
 * 失敗的 query 也不會自動重試而拖慢斷言。
 */
export function renderWithProviders(
  ui: () => JSX.Element,
  client: QueryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } }),
): QueryClient {
  render(() => withProviders(client, ui));
  return client;
}
