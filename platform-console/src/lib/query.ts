import { QueryClient } from "@tanstack/solid-query";

/**
 * 平台查詢的 QueryClient：**不重試**。
 *
 * 平台工具是 operator 的即時操作面：401/403 重試只會延後登入頁的出現，
 * 而平台寫入（收款、改方案）重試本身就有語意風險。要重試的是「使用者」，不是客戶端。
 */
export const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});
