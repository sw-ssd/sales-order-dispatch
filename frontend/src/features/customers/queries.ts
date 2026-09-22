import { createClient } from "@connectrpc/connect";
import { queryOptions } from "@tanstack/solid-query";
import {
  CustomerService,
  type ListCustomersResponse,
} from "~/lib/proto/customers/v1/customer_pb";
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

export type { ListCustomersResponse };
