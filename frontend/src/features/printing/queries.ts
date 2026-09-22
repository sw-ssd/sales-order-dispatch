import { createClient } from "@connectrpc/connect";
import { queryOptions } from "@tanstack/solid-query";
import {
  PrintService,
  type ListLogsResponse,
} from "~/lib/proto/products/v1/print_pb";
import { transport } from "~/lib/transport";

/**
 * 單據列印查詢契約；樣板 = `features/users/queries.ts` 檔頭契約（不另發明）。
 *
 * 與其他清單的三處差異：
 * - **`ListLogsRequest` 沒有 `sort`/`desc`**（見 proto）→ 本頁不開排序（後端也不解析）。
 * - 總數在回應的 `total`（不是 `pagination.total`）。
 * - **只涵蓋 `print_logs`**：printing spec §151 要求同時查 `print_previews`，
 *   但 proto 未定義 `ListPreviews`（Go 端 `PrintPreview` 只有 `Create`、無查詢路徑）
 *   → 預覽紀錄列表**待後端補 RPC 後再做**，此處不假裝有。
 */
export const printClient = createClient(PrintService, transport);

/** 列印記錄每頁筆數（後端上限 100，與其他清單同用 20）。 */
export const PRINT_LOG_PAGE_SIZE = 20;

/** 列印記錄查詢參數（全部參數都進 queryKey）。 */
export interface PrintLogsParams {
  page: number;
  pageSize: number;
  /** YYYY-MM-DD，可空（後端 `date_from`/`date_to` 均為此格式）。 */
  dateFrom?: string;
  dateTo?: string;
  documentType?: string;
  routeId?: string;
}

/** 列印記錄查詢選項；`createQuery(() => printLogsQueryOptions(params))`。 */
export const printLogsQueryOptions = (params: PrintLogsParams) =>
  queryOptions<ListLogsResponse>({
    queryKey: [
      "printLogs",
      {
        page: params.page,
        pageSize: params.pageSize,
        dateFrom: params.dateFrom,
        dateTo: params.dateTo,
        documentType: params.documentType,
        routeId: params.routeId,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      printClient.listLogs({
        page: params.page,
        pageSize: params.pageSize,
        dateFrom: params.dateFrom ?? "",
        dateTo: params.dateTo ?? "",
        documentType: params.documentType ?? "",
        routeId: params.routeId ?? "",
      }),
  });
