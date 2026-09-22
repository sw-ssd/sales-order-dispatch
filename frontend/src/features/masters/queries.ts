import { createClient } from "@connectrpc/connect";
import { queryOptions } from "@tanstack/solid-query";
import {
  ProcessingSpecService,
  ProductCategoryService,
  RouteService,
  WarehouseService,
  type ListProcessingSpecsResponse,
  type ListProductCategoriesResponse,
  type ListRoutesResponse,
  type ListWarehousesResponse,
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
 *
 * 04 Task 3.4 後本檔再承載倉別／商品分類／分切規格三個部門級主檔：`ListXxxRequest`
 * 同樣沒有 `sort`/`desc`、總數同在 `pagination.total`；守衛（見 router）為倉別
 * `read, dispatch`（商品進貨／揀貨倉的來源）、分類與規格 `read, product`（商品主檔的來源）。
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

/** 倉別主檔 client（商品「進貨倉／揀貨倉」的來源）。 */
export const warehouseClient = createClient(WarehouseService, transport);

/** 倉別清單每頁筆數（與其他清單共用同一個值）。 */
export const WAREHOUSES_PAGE_SIZE = 20;

/** 倉別清單查詢參數（全部參數都進 queryKey）。 */
export interface WarehouseListParams {
  page: number;
  pageSize: number;
  keyword?: string;
  includeDeleted?: boolean;
}

/** 倉別清單查詢選項；`createQuery(() => warehousesQueryOptions(params))`。 */
export const warehousesQueryOptions = (params: WarehouseListParams) =>
  queryOptions<ListWarehousesResponse>({
    queryKey: [
      "warehouses",
      {
        page: params.page,
        pageSize: params.pageSize,
        keyword: params.keyword,
        includeDeleted: params.includeDeleted,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      warehouseClient.listWarehouses({
        page: params.page,
        pageSize: params.pageSize,
        keyword: params.keyword ?? "",
        includeDeleted: params.includeDeleted ?? false,
      }),
  });

/** 商品分類 client（商品主檔「所屬分類」的來源）。 */
export const productCategoryClient = createClient(ProductCategoryService, transport);

/** 分類清單每頁筆數（與其他清單共用同一個值）。 */
export const PRODUCT_CATEGORIES_PAGE_SIZE = 20;

/** 分類清單查詢參數（全部參數都進 queryKey）。 */
export interface ProductCategoryListParams {
  page: number;
  pageSize: number;
  keyword?: string;
  includeDeleted?: boolean;
}

/** 分類清單查詢選項；`createQuery(() => productCategoriesQueryOptions(params))`。 */
export const productCategoriesQueryOptions = (params: ProductCategoryListParams) =>
  queryOptions<ListProductCategoriesResponse>({
    queryKey: [
      "productCategories",
      {
        page: params.page,
        pageSize: params.pageSize,
        keyword: params.keyword,
        includeDeleted: params.includeDeleted,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      productCategoryClient.listProductCategories({
        page: params.page,
        pageSize: params.pageSize,
        keyword: params.keyword ?? "",
        includeDeleted: params.includeDeleted ?? false,
      }),
  });

/** 分切規格 client（商品主檔勾選的加工／揀貨條件來源）。 */
export const processingSpecClient = createClient(ProcessingSpecService, transport);

/** 規格清單每頁筆數（與其他清單共用同一個值）。 */
export const PROCESSING_SPECS_PAGE_SIZE = 20;

/** 規格清單查詢參數（全部參數都進 queryKey）。 */
export interface ProcessingSpecListParams {
  page: number;
  pageSize: number;
  keyword?: string;
  includeDeleted?: boolean;
}

/** 規格清單查詢選項；`createQuery(() => processingSpecsQueryOptions(params))`。 */
export const processingSpecsQueryOptions = (params: ProcessingSpecListParams) =>
  queryOptions<ListProcessingSpecsResponse>({
    queryKey: [
      "processingSpecs",
      {
        page: params.page,
        pageSize: params.pageSize,
        keyword: params.keyword,
        includeDeleted: params.includeDeleted,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      processingSpecClient.listProcessingSpecs({
        page: params.page,
        pageSize: params.pageSize,
        keyword: params.keyword ?? "",
        includeDeleted: params.includeDeleted ?? false,
      }),
  });
