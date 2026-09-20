import { createClient } from "@connectrpc/connect";
import { queryOptions } from "@tanstack/solid-query";
import {
  TenantEntitlementService,
  type GetTenantEntitlementsResponse,
} from "~/lib/proto/platform/v1/platform_pb";
import { transport } from "~/lib/transport";

/**
 * 租戶端的唯讀權益投影（Plan B Task 10）：服務掛在**租戶** `apiMux`
 * （`/api/v1/platform.v1.TenantEntitlementService/GetTenantEntitlements`），以租戶 session 身分為準、
 * 只回自己公司。前端只讀——沒有任何寫入路徑，`platform.*` 能力也不出現在租戶端（S11）。
 */
const client = createClient(TenantEntitlementService, transport);

export async function loadTenantEntitlements(): Promise<GetTenantEntitlementsResponse> {
  return client.getTenantEntitlements({});
}

/**
 * 卡片與 shell 提示條共用同一個查詢：同一個 queryKey ⇒ 同一個 QueryClient 只打一次，
 * 提示條不會讓每個頁面多一趟 RPC。
 */
export const tenantEntitlementsQueryOptions = () =>
  queryOptions({
    queryKey: ["tenant-entitlements"],
    // 提示條掛在 shell 上，每個頁面都會掛載一次；60 秒內不重打（投影本身在後端也有快取）。
    staleTime: 60_000,
    queryFn: loadTenantEntitlements,
  });
