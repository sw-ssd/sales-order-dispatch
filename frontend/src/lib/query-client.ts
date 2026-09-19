import { Code, ConnectError } from "@connectrpc/connect";
import { QueryClient } from "@tanstack/solid-query";

// 共用查詢 retry 謂詞(Phase 3 D):只重試暫時性錯誤。
// Unavailable/Unknown/DeadlineExceeded 為連線層/逾時類錯誤,退避重試有意義;
// 確定性錯誤(NotFound/InvalidArgument/PermissionDenied/Unauthenticated/AlreadyExists/
// FailedPrecondition)重試只會退避重打、延遲錯誤顯示(403 要重打 4 次),一律不重試。
// Code 值以 @connectrpc/connect 2.2.0 的列舉為準
// (node_modules/@connectrpc/connect/dist/esm/code.js:Unavailable=14/Unknown=2/DeadlineExceeded=4)。
// 非 ConnectError(非 RPC 的程式錯誤)不重試:重試不會改變結果。
export function isRetryableQueryError(error: unknown): boolean {
  return (
    error instanceof ConnectError &&
    (error.code === Code.Unavailable ||
      error.code === Code.Unknown ||
      error.code === Code.DeadlineExceeded)
  );
}

// 重試上限 3 次(= TanStack Query 數字形式 3 的既有預設,共 4 次嘗試)。
// retry 改為函式後必須自行封頂:query-core 對「函式回 true」不設次數上限
// (query-core retryer.js:`retry(failureCount, error)` 回 true 即續重試),
// 否則 Unavailable 會無限退避重試。
const MAX_RETRIES = 3;

// 全站共用 QueryClient(T13;WEB-INF-04)。retry 設為全域預設查詢選項,
// 清單 queryOptions 不必各自重複。權限異動後主動重載慣例:
//   queryClient.invalidateQueries({ queryKey: ["ability"] })
// (權限設置頁儲存成功後呼叫,2.7/2.11 前端計畫引用)。
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => failureCount < MAX_RETRIES && isRetryableQueryError(error),
    },
  },
});
