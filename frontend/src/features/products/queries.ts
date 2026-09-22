import { createClient } from "@connectrpc/connect";
import {
  infiniteQueryOptions,
  type InfiniteData,
  type QueryKey,
} from "@tanstack/solid-query";
import {
  ProductService,
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
