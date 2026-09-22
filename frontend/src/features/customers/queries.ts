import { createClient } from "@connectrpc/connect";
import {
  infiniteQueryOptions,
  queryOptions,
  type InfiniteData,
  type QueryKey,
} from "@tanstack/solid-query";
import {
  CustomerService,
  type ListCustomersResponse,
} from "~/lib/proto/customers/v1/customer_pb";
import {
  CustomerProductService,
  ProductService,
} from "~/lib/proto/products/v1/product_pb";
import { transport } from "~/lib/transport";

/**
 * 客戶主檔清單查詢契約；樣板 = `features/users/queries.ts` 檔頭契約（不另發明）：
 * - `queryKey` 一律 `["customers", { …全部查詢參數 }]`（含 `sort`／`desc`，失效前綴 `["customers"]`）。
 * - `placeholderData: (prev) => prev`；不寫 `retry`（共用謂詞在 `lib/query-client.ts`）。
 * - `queryFn` 回傳 proto 原生回應，頁面自行由 `pagination.total` 推導總筆數。
 * - **排序（D1）**：`sort`（白名單 `name`/`customer_code`/`created_at`）＋`desc` 顯式進 key 與請求，
 *   未排序時送 `sort: ""`/`desc: false`（空字串即「服務預設排序」的正典空值）。
 * - 關鍵字在頁面是 signal「草稿」，**只在 submit 時**進 query key；頁碼唯一真相＝table 的 pagination state。
 */
export const customerClient = createClient(CustomerService, transport);

/**
 * 專屬清單與產品挑選器的 client：兩者都在 `products.v1`（`CustomerProductService`
 * 與 `ProductService` 同一份 `product.proto`）。本檔不跨 feature import `products/queries`
 * —— 那會讓客戶頁的型別依賴商品頁的實作檔案（同 `metadictClient` 在兩個 feature 各自
 * 一份的既有做法）。
 */
export const customerProductClient = createClient(CustomerProductService, transport);
export const productClient = createClient(ProductService, transport);

/** 客戶清單查詢參數（全部參數都進 queryKey）。 */
export interface CustomerListParams {
  page: number;
  pageSize: number;
  /** 排序白名單欄位名（`name`/`customer_code`/`created_at`）；**空字串＝服務預設排序**。 */
  sort: string;
  /** 是否降冪；後端在 `sort` 為空時忽略（仍照送，見檔頭契約）。 */
  desc: boolean;
  keyword?: string;
  includeDeleted?: boolean;
}

/** 客戶清單查詢選項；`createQuery(() => customersQueryOptions(params))`。 */
export const customersQueryOptions = (params: CustomerListParams) =>
  queryOptions({
    queryKey: [
      "customers",
      {
        page: params.page,
        pageSize: params.pageSize,
        sort: params.sort,
        desc: params.desc,
        keyword: params.keyword,
        includeDeleted: params.includeDeleted,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      customerClient.listCustomers({
        page: params.page,
        pageSize: params.pageSize,
        keyword: params.keyword ?? "",
        includeDeleted: params.includeDeleted ?? false,
        sort: params.sort,
        desc: params.desc,
      }),
  });

/** 累積式客戶下拉（訂單頁「客戶」）每頁筆數：比清單的 20 筆大，減少「載入更多」次數。 */
const CUSTOMER_DROPDOWN_PAGE_SIZE = 50;

/** 客戶下拉查詢參數（全部參數都進 queryKey）。 */
export interface CustomerDropdownParams {
  keyword?: string;
}

/**
 * 客戶下拉查詢選項（累積式）；`createInfiniteQuery(() => customerDropdownQueryOptions(params))`。
 *
 * `getNextPageParam` 以「已載入筆數 vs 總筆數」判斷是否還有下一頁；回 `undefined` 即
 * `hasNextPage === false`。樣板 = `features/users/queries.ts` 的 `companyDropdownQueryOptions`。
 */
export const customerDropdownQueryOptions = (params: CustomerDropdownParams) =>
  infiniteQueryOptions<
    ListCustomersResponse,
    Error,
    InfiniteData<ListCustomersResponse>,
    QueryKey,
    number
  >({
    queryKey: [
      "customers",
      "options",
      { pageSize: CUSTOMER_DROPDOWN_PAGE_SIZE, keyword: params.keyword },
    ],
    initialPageParam: 1,
    placeholderData: (prev) => prev,
    queryFn: ({ pageParam }) =>
      customerClient.listCustomers({
        page: pageParam,
        pageSize: CUSTOMER_DROPDOWN_PAGE_SIZE,
        keyword: params.keyword ?? "",
        includeDeleted: false,
        sort: "",
        desc: false,
      }),
    getNextPageParam: (lastPage, allPages) =>
      allPages.reduce((loaded, page) => loaded + page.customers.length, 0) <
      Number(lastPage.pagination?.total ?? 0)
        ? allPages.length + 1
        : undefined,
  });

export type { ListCustomersResponse };

/**
 * 地址簿／聯絡人清單查詢選項（`AddressBookDialog` 用）。
 *
 * 兩者的 `queryKey` 都掛在 `["customers", …]` 前綴下 → 任何客戶相關 mutation 之後
 * 一句 `invalidateQueries({ queryKey: ["customers"] })` 就會一起重載（同 customers 慣例）。
 * 兩支 RPC 都**不帶 `includeDeleted`**：軟刪除的地址/聯絡人在 3.2 是不可見的資料
 * （只有 Restore 類 RPC 才會帶它），本頁不做「含已刪除」檢視。
 */
export const addressesQueryOptions = (customerId: string) =>
  queryOptions({
    queryKey: ["customers", "addresses", { customerId }],
    placeholderData: (prev) => prev,
    queryFn: () => customerClient.listAddresses({ customerId, includeDeleted: false }),
  });

export const contactsQueryOptions = (customerId: string) =>
  queryOptions({
    queryKey: ["customers", "contacts", { customerId }],
    placeholderData: (prev) => prev,
    queryFn: () => customerClient.listContacts({ customerId, includeDeleted: false }),
  });

/**
 * 客戶專屬清單查詢選項（`CustomerProductsDialog` 用）。
 *
 * 樣板同上：key 掛 `["customers"]` 前綴 → 任何客戶層 mutation 一次失效就帶到這裡。
 * **不帶 `forOrder`**：那支旗標是「下單用途，排除 default_qty=0」的語意（訂單入口用），
 * 管理頁要看全部列 —— 否則管理員看不到自己剛設的 0 數量預設列，會以為新增沒生效。
 */
export const customerProductsQueryOptions = (customerId: string) =>
  queryOptions({
    queryKey: ["customers", "customerProducts", { customerId }],
    placeholderData: (prev) => prev,
    queryFn: () =>
      customerProductClient.listCustomerProducts({
        customerId,
        forOrder: false,
        includeDeleted: false,
      }),
  });

/**
 * 產品挑選器的清單查詢（只撈啟用中產品、上限 100）。
 *
 * 沒有「選擇全部產品」的負擔：這是一次性下拉，關鍵字只在 submit 進 key
 * （同清單頁草稿慣例）。`total` 是 int64 → `Number()`。
 */
export const pickerProductsQueryOptions = (keyword: string) =>
  queryOptions({
    queryKey: ["customers", "pickerProducts", { keyword }],
    placeholderData: (prev) => prev,
    queryFn: () =>
      productClient.listProducts({
        page: 1,
        pageSize: 100,
        keyword,
        categoryId: "",
        includeDeleted: false,
      }),
  });
