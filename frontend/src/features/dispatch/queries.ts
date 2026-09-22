import { createClient } from "@connectrpc/connect";
import {
  infiniteQueryOptions,
  queryOptions,
  type InfiniteData,
  type QueryKey,
} from "@tanstack/solid-query";
import {
  RouteService,
  type ListRoutesResponse,
} from "~/lib/proto/masters/v1/master_pb";
import { DispatchService } from "~/lib/proto/salesorder/v1/dispatch_pb";
import {
  SalesOrderService,
  type ListOrdersResponse,
} from "~/lib/proto/salesorder/v1/salesorder_pb";
import { transport } from "~/lib/transport";

/**
 * 派車看板查詢契約；樣板 = `features/users/queries.ts` 檔頭契約（不另發明）。
 *
 * 與其他頁的兩處差異：
 * - **看板以日期為主軸**：`ListOrders.expected_delivery_date` 是必帶條件（dispatch spec：
 *   看板僅顯示所選日期的訂單）。該欄是本次為看板補上的後端過濾，前端不可改以「撈全部再前端過濾」
 *   代替 —— 後端 `maxPageSize = 100`，日期外的訂單會把該日的擠出分頁。
 * - **事件只作失效提示**（spec：收到 `BoardEvent` 後使看板查詢失效並全量重查），
 *   所以這裡只有查詢、沒有事件快取；失效單位 = `["boardOrders"]`。
 */
export const dispatchClient = createClient(DispatchService, transport);
const orderClient = createClient(SalesOrderService, transport);
export const routeClient = createClient(RouteService, transport);

/** 看板單頁筆數＝後端 `maxPageSize` 上限（`shared_service.go`）。 */
export const BOARD_PAGE_SIZE = 100;

/** 車次單頁筆數；車次是部門級主檔、筆數有限，超過此數看板會顯示截斷警告（不靜默）。 */
export const ROUTE_PAGE_SIZE = 100;

/** 看板訂單查詢（累積式）：日期進 key，整頁補齊以避免單頁上限把該日訂單擠掉。 */
export const boardOrdersQueryOptions = (date: string) =>
  infiniteQueryOptions<
    ListOrdersResponse,
    Error,
    InfiniteData<ListOrdersResponse>,
    QueryKey,
    number
  >({
    queryKey: ["boardOrders", { date, pageSize: BOARD_PAGE_SIZE }],
    initialPageParam: 1,
    placeholderData: (prev) => prev,
    queryFn: ({ pageParam }) =>
      orderClient.listOrders({
        page: pageParam,
        pageSize: BOARD_PAGE_SIZE,
        status: "",
        customerId: "",
        source: "",
        keyword: "",
        includeDeleted: false,
        expectedDeliveryDate: date,
      }),
    getNextPageParam: (lastPage, allPages) =>
      allPages.reduce((loaded, page) => loaded + page.orders.length, 0) < lastPage.total
        ? allPages.length + 1
        : undefined,
  });

/**
 * 看板車次查詢（單頁）：看板把車次畫成欄，欄位無法「載入更多」，故單頁取回；
 * 超過 `ROUTE_PAGE_SIZE` 由頁面顯示截斷警告（見檔頭：不靜默）。
 */
export const boardRoutesQueryOptions = () =>
  queryOptions<ListRoutesResponse>({
    queryKey: ["boardRoutes", { pageSize: ROUTE_PAGE_SIZE }],
    queryFn: () =>
      routeClient.listRoutes({
        page: 1,
        pageSize: ROUTE_PAGE_SIZE,
        keyword: "",
        includeDeleted: false,
      }),
  });
