import { createClient } from "@connectrpc/connect";
import { queryOptions } from "@tanstack/solid-query";
import {
  RouteService,
  type ListRoutesResponse,
} from "~/lib/proto/masters/v1/master_pb";
import { transport } from "~/lib/transport";

/**
 * 車次主檔查詢契約；樣板 = `features/users/queries.ts` 檔頭契約（不另發明）。
 *
 * 車次是部門級主檔（masters.v1，04 Task 3.4），也是派車看板的欄位來源：
 * 看板顯示啟用中的車次，本頁負責建立／編輯／停用。
 *
 * `ListRoutesRequest` **沒有 `sort`/`desc`**（見 proto）→ 本頁刻意不開排序，
 * 欄位一律非可排序表頭（同 `OrdersPage`／`ProductsPage`：送了後端也不解析）。
 *
 * 註：`masters.v1` 的 RPC 不在 `server.protectedRPC` 表內（僅 Role/Company/Department/User
 * 四域在），故寫入的授權落在 handler 的 `requireAuth`＋`deptScope` 與 RLS；
 * 本頁守衛用 `read, dispatch`（見 router）—— 能開看板的人就能看其欄位來源。
 */
export const routeClient = createClient(RouteService, transport);

/** 清單每頁筆數（與其他清單共用同一個值）。 */
export const ROUTES_PAGE_SIZE = 20;

/** 車次清單查詢參數（全部參數都進 queryKey）。 */
export interface RouteListParams {
  page: number;
  pageSize: number;
  keyword?: string;
  includeDeleted?: boolean;
}

/** 車次清單查詢選項；`createQuery(() => routesQueryOptions(params))`。 */
export const routesQueryOptions = (params: RouteListParams) =>
  queryOptions<ListRoutesResponse>({
    queryKey: [
      "routes",
      {
        page: params.page,
        pageSize: params.pageSize,
        keyword: params.keyword,
        includeDeleted: params.includeDeleted,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      routeClient.listRoutes({
        page: params.page,
        pageSize: params.pageSize,
        keyword: params.keyword ?? "",
        includeDeleted: params.includeDeleted ?? false,
      }),
  });
