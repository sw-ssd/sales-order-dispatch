import { createClient } from "@connectrpc/connect";
import { queryOptions } from "@tanstack/solid-query";
import {
  AuditService,
  type ListAuditLogsResponse,
} from "~/lib/proto/audit/v1/audit_pb";
import { transport } from "~/lib/transport";

/**
 * 稽核日誌查詢契約；樣板 = `features/printing/queries.ts` 檔頭契約（不另發明）：
 * - `queryKey` 一律 `["auditLogs", { …全部查詢參數 }]`，每個參數都進 key。
 * - `placeholderData: (prev) => prev`；不寫 `retry`（共用謂詞在 `lib/query-client.ts`）。
 * - `queryFn` 回傳 proto 原生回應；**總數在 `pagination.total`**（同 customers/users），
 *   且它是 int64 → 頁面一律 `Number(...)` 轉換（proto int64 在 TS 是 bigint）。
 * - **無排序參數**：後端固定 `created_at` 倒序（＋`id` 次序鍵避免同值群逐頁重複/遺漏）。
 * - 篩選在頁面是 signal「草稿」，**只在 submit 時**進 query key。
 *
 * `from`/`to` 是 **RFC3339**（不是 `YYYY-MM-DD`）：兩者都空時後端套用「近 3 個月」
 * 的預設保留窗（D27），所以「清空時間」有明確語意，本頁不自行算三個月再送出去。
 */
export const auditClient = createClient(AuditService, transport);

/** 稽核清單每頁筆數（與其他清單同用 20）。 */
export const AUDIT_PAGE_SIZE = 20;

/** 稽核清單查詢參數（全部參數都進 queryKey；空字串＝該篩選不套用）。 */
export interface AuditListParams {
  from: string;
  to: string;
  action: string;
  resourceType: string;
  resourceId: string;
  userId: string;
  page: number;
  pageSize: number;
}

/** 稽核清單查詢選項；`createQuery(() => auditLogsQueryOptions(params))`。 */
export const auditLogsQueryOptions = (params: AuditListParams) =>
  queryOptions<ListAuditLogsResponse>({
    queryKey: [
      "auditLogs",
      {
        from: params.from,
        to: params.to,
        action: params.action,
        resourceType: params.resourceType,
        resourceId: params.resourceId,
        userId: params.userId,
        page: params.page,
        pageSize: params.pageSize,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      auditClient.listAuditLogs({
        page: params.page,
        pageSize: params.pageSize,
        from: params.from,
        to: params.to,
        companyId: "",
        action: params.action,
        resourceType: params.resourceType,
        resourceId: params.resourceId,
        userId: params.userId,
      }),
  });
