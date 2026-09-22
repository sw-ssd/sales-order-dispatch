import { createClient } from "@connectrpc/connect";
import {
  infiniteQueryOptions,
  queryOptions,
  type InfiniteData,
  type QueryKey,
} from "@tanstack/solid-query";
import {
  MetadictService,
  type ListOptionsResponse,
} from "~/lib/proto/metadict/v1/metadict_pb";
import {
  ProductCategoryService,
  WarehouseService,
  type ListProductCategoriesResponse,
  type ListWarehousesResponse,
} from "~/lib/proto/masters/v1/master_pb";
import {
  ProductService,
  type GetProductResponse,
  type ListProductsResponse,
} from "~/lib/proto/products/v1/product_pb";
import { transport } from "~/lib/transport";

/**
 * 商品查詢契約；樣板 = `features/users/queries.ts` 檔頭契約（不另發明）。
 *
 * `ListProductsRequest` **沒有 `sort`/`desc`**（見 proto），所以這裡刻意沒有排序參數 ——
 * 送了後端也不會解析。失效前綴 `["products"]`。
 */
export const productClient = createClient(ProductService, transport);

/** 商品下拉每頁筆數：比清單的 20 筆大，減少「載入更多」次數。 */
const PRODUCT_DROPDOWN_PAGE_SIZE = 50;

/** 商品下拉查詢參數（全部參數都進 queryKey）。 */
export interface ProductDropdownParams {
  keyword?: string;
}

/**
 * 商品下拉查詢選項（累積式）；`createInfiniteQuery(() => productDropdownQueryOptions(params))`。
 *
 * `getNextPageParam` 以「已載入筆數 vs 總筆數」判斷是否還有下一頁；回 `undefined` 即
 * `hasNextPage === false`。樣板 = `features/users/queries.ts` 的 `companyDropdownQueryOptions`。
 *
 * 恆帶 `includeDeleted: false`：建單只能綁仍存在的商品。
 */
export const productDropdownQueryOptions = (params: ProductDropdownParams) =>
  infiniteQueryOptions<
    ListProductsResponse,
    Error,
    InfiniteData<ListProductsResponse>,
    QueryKey,
    number
  >({
    queryKey: [
      "products",
      "options",
      { pageSize: PRODUCT_DROPDOWN_PAGE_SIZE, keyword: params.keyword },
    ],
    initialPageParam: 1,
    placeholderData: (prev) => prev,
    queryFn: ({ pageParam }) =>
      productClient.listProducts({
        page: pageParam,
        pageSize: PRODUCT_DROPDOWN_PAGE_SIZE,
        keyword: params.keyword ?? "",
        categoryId: "",
        includeDeleted: false,
      }),
    getNextPageParam: (lastPage, allPages) =>
      allPages.reduce((loaded, page) => loaded + page.products.length, 0) <
      Number(lastPage.pagination?.total ?? 0)
        ? allPages.length + 1
        : undefined,
  });

/** 字典 client（商品頁的單位字值 `type=unit` 來源）。 */
export const metadictClient = createClient(MetadictService, transport);

/** 部門級主檔 client：分類下拉（商品「所屬分類」）與倉別下拉（進貨／揀貨倉）共用。 */
export const productCategoryClient = createClient(ProductCategoryService, transport);
export const warehouseClient = createClient(WarehouseService, transport);

/** 商品清單每頁筆數（與其他清單共用同一個值）。 */
export const PRODUCT_PAGE_SIZE = 20;

/** 商品清單查詢參數（全部參數都進 queryKey）。 */
export interface ProductListParams {
  page: number;
  pageSize: number;
  keyword?: string;
  categoryId?: string;
  includeDeleted?: boolean;
}

/**
 * 商品清單查詢選項；`createQuery(() => productsQueryOptions(params))`。
 *
 * 總數在 `pagination.total`（與訂單頁的 `total` 不同）；無排序參數（見檔頭）。
 */
export const productsQueryOptions = (params: ProductListParams) =>
  queryOptions({
    queryKey: [
      "products",
      "list",
      {
        page: params.page,
        pageSize: params.pageSize,
        keyword: params.keyword,
        categoryId: params.categoryId,
        includeDeleted: params.includeDeleted,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      productClient.listProducts({
        page: params.page,
        pageSize: params.pageSize,
        keyword: params.keyword ?? "",
        categoryId: params.categoryId ?? "",
        includeDeleted: params.includeDeleted ?? false,
      }),
  });

/** 單筆商品（編輯前補載單位與分類設定）；key `["product", id]`。 */
export const productQueryOptions = (id: string) =>
  queryOptions<GetProductResponse>({
    queryKey: ["product", id],
    queryFn: () => productClient.getProduct({ id }),
  });

/**
 * 單位字值下拉（metadicts `type=unit`）。
 *
 * `ListOptionsRequest` 沒有分頁參數 —— 字典是一次回全的，故用 `queryOptions` 而非
 * 累積式（分頁語意在這裡不存在）。恆帶 `type: "unit"`（後端 `validateUnitCode`
 * 只認這個 type，且含部門擴充）。
 */
export const unitOptionsQueryOptions = () =>
  queryOptions<ListOptionsResponse>({
    queryKey: ["metadictOptions", "unit"],
    queryFn: () => metadictClient.listOptions({ type: "unit", keyword: "" }),
  });

/** 分類下拉每頁筆數（與其他下拉一致：比清單的 20 筆大，減少「載入更多」次數）。 */
const MASTER_DROPDOWN_PAGE_SIZE = 50;

/**
 * 商品分類下拉（累積式）；樣板 = `features/users/queries.ts` 的 `companyDropdownQueryOptions`。
 * `getNextPageParam` 以「已載入筆數 vs 總筆數」判斷是否還有下一頁；回 `undefined` 即無下一頁。
 */
export const categoryDropdownQueryOptions = () =>
  infiniteQueryOptions<
    ListProductCategoriesResponse,
    Error,
    InfiniteData<ListProductCategoriesResponse>,
    QueryKey,
    number
  >({
    queryKey: ["productCategories", "options", { pageSize: MASTER_DROPDOWN_PAGE_SIZE }],
    initialPageParam: 1,
    placeholderData: (prev) => prev,
    queryFn: ({ pageParam }) =>
      productCategoryClient.listProductCategories({
        page: pageParam,
        pageSize: MASTER_DROPDOWN_PAGE_SIZE,
        keyword: "",
        includeDeleted: false,
      }),
    getNextPageParam: (lastPage, allPages) =>
      allPages.reduce((loaded, page) => loaded + page.productCategories.length, 0) <
      Number(lastPage.pagination?.total ?? 0)
        ? allPages.length + 1
        : undefined,
  });

/** 倉別下拉（累積式）；商品的進貨／揀貨倆倉共用同一份選項。 */
export const warehouseDropdownQueryOptions = () =>
  infiniteQueryOptions<
    ListWarehousesResponse,
    Error,
    InfiniteData<ListWarehousesResponse>,
    QueryKey,
    number
  >({
    queryKey: ["warehouses", "options", { pageSize: MASTER_DROPDOWN_PAGE_SIZE }],
    initialPageParam: 1,
    placeholderData: (prev) => prev,
    queryFn: ({ pageParam }) =>
      warehouseClient.listWarehouses({
        page: pageParam,
        pageSize: MASTER_DROPDOWN_PAGE_SIZE,
        keyword: "",
        includeDeleted: false,
      }),
    getNextPageParam: (lastPage, allPages) =>
      allPages.reduce((loaded, page) => loaded + page.warehouses.length, 0) <
      Number(lastPage.pagination?.total ?? 0)
        ? allPages.length + 1
        : undefined,
  });
