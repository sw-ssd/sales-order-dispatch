import { createClient } from "@connectrpc/connect";
import { queryOptions } from "@tanstack/solid-query";
import {
  MetadictService,
  type ListOptionsResponse,
} from "~/lib/proto/metadict/v1/metadict_pb";
import {
  SalesOrderService,
  type ListOrderEventsResponse,
  type ListOrdersResponse,
} from "~/lib/proto/salesorder/v1/salesorder_pb";
import { transport } from "~/lib/transport";

/**
 * 訂單查詢契約；樣板 = `features/users/queries.ts` 檔頭契約（不另發明）。三處不同：
 * - **回傳總數在 `total` 不是 `pagination`**（`ListOrdersResponse`／`ListOrderEventsResponse`
 *   都是 `int32 total`），所以頁面從 `data.total` 推總筆數，不看 `pagination`。
 * - **`ListOrdersRequest` 沒有 `sort`/`desc`**（見 proto）→ 清單刻意不開排序，這裡沒有排序參數。
 * - 狀態／客戶／來源／關鍵字都在 key；`includeDeleted` 亦然（含已刪除是檢視切換）。
 */
export const orderClient = createClient(SalesOrderService, transport);
export const metadictClient = createClient(MetadictService, transport);

/** 清單每頁筆數（與其他清單共用同一個值）。 */
export const ORDER_PAGE_SIZE = 20;

/** 訂單清單查詢參數（全部參數都進 queryKey）。 */
export interface OrderListParams {
  page: number;
  pageSize: number;
  status?: string;
  customerId?: string;
  source?: string;
  keyword?: string;
  includeDeleted?: boolean;
}

/** 訂單清單查詢選項；`createQuery(() => ordersQueryOptions(params))`。 */
export const ordersQueryOptions = (params: OrderListParams) =>
  queryOptions({
    queryKey: [
      "orders",
      {
        page: params.page,
        pageSize: params.pageSize,
        status: params.status,
        customerId: params.customerId,
        source: params.source,
        keyword: params.keyword,
        includeDeleted: params.includeDeleted,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      orderClient.listOrders({
        page: params.page,
        pageSize: params.pageSize,
        status: params.status ?? "",
        customerId: params.customerId ?? "",
        source: params.source ?? "",
        keyword: params.keyword ?? "",
        includeDeleted: params.includeDeleted ?? false,
      }),
  });

/** 單筆訂單＋明細；key 前綴 `["order", id]`（與清單的 `["orders"]` 分開失效單位）。 */
export const orderQueryOptions = (id: string) =>
  queryOptions({
    queryKey: ["order", id],
    queryFn: () => orderClient.getOrder({ id }),
  });

/** 訂單事件軌跡查詢參數（全部參數都進 queryKey）。 */
export interface OrderEventsParams {
  orderId: string;
  page: number;
  pageSize: number;
}

/** 訂單事件軌跡查詢選項；失效前綴 `["orderEvents", orderId]`。 */
export const orderEventsQueryOptions = (params: OrderEventsParams) =>
  queryOptions<ListOrderEventsResponse>({
    queryKey: [
      "orderEvents",
      params.orderId,
      { page: params.page, pageSize: params.pageSize },
    ],
    queryFn: () =>
      orderClient.listOrderEvents({
        salesOrderId: params.orderId,
        page: params.page,
        pageSize: params.pageSize,
      }),
  });

/**
 * 訂單來源下拉（metadicts `order_source`，系統級固定 W/A）。
 *
 * 恆帶 `type: "order_source"`：建單的 `source` 必須是該字典的啟用值
 * （後端 `validateOrderSource` 只認這個 type，且不讀部門擴充）。
 */
export const orderSourceOptionsQueryOptions = () =>
  queryOptions<ListOptionsResponse>({
    queryKey: ["metadictOptions", "order_source"],
    queryFn: () => metadictClient.listOptions({ type: "order_source", keyword: "" }),
  });

export type { ListOrdersResponse, ListOrderEventsResponse };
