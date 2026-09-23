import type { QueryObserverResult } from "@tanstack/solid-query";

/**
 * 讀取查詢資料，且**不觸發 Suspense**。
 *
 * `@tanstack/solid-query` 的 query 是包在 Solid resource 上的 proxy（`useBaseQuery` 的 handler）：
 * 只有當該 query key **從未有過值**時，`.data` 才落到 `resource()` —— 在 pending 期間讀它就會
 * suspend，把最近的 `Suspense` fallback 交出來；一旦有過值就改讀 `resource.latest`，永不 suspend。
 *
 * 本專案的 router 把整條路由包在一個 Suspense 邊界內（`Matches` 的頂層邊界），因此任何一個
 * 「順手在 JSX 讀 `.data`」的查詢，在**首次載入**時都會把整個內容區換成 fallback：
 * 導覽到 `/customers` 時畫面空白、訂單詳情對話框一開整頁消失，都是這條路徑。
 *
 * 這個 helper 把「pending 期間不碰 `.data`」集中成唯一入口：pending（含未啟用）時把
 * `undefined` 交給取值函式，其餘情況給真正的資料。`isPending` 是 store 上的普通欄位，
 * 讀它不會 suspend；`placeholderData` 讓前一 key 的資料在切 key 期間維持 `status = success`
 * （`isPending` 為 false），所以「換頁保留舊頁資料」的行為不受影響。
 *
 * 用法：把 `query.data?.customers ?? []` 改寫成 `queryData(query, (d) => d?.customers ?? [])`。
 */
export function queryData<TData, TResult>(
  query: Pick<QueryObserverResult<TData, unknown>, "isPending" | "data">,
  read: (data: TData | undefined) => TResult
): TResult {
  return read(query.isPending ? undefined : query.data);
}
