import { createClient } from "@connectrpc/connect";
import { queryOptions } from "@tanstack/solid-query";
import {
  ReturnService,
  type GetReturnCertificateResponse,
  type GetReturnRequestResponse,
  type ListReturnRequestsResponse,
} from "~/lib/proto/salesorder/v1/returns_pb";
import { transport } from "~/lib/transport";
import { customerClient } from "../customers/queries";

/**
 * 退貨查詢契約；樣板 = `features/users/queries.ts` 檔頭契約（不另發明）：
 * - `queryKey` 一律 `["returnRequests", { …全部查詢參數 }]`，前綴 `["returnRequests"]`
 *   即失效單位 → 審核成功後一次 `invalidateQueries({ queryKey: ["returnRequests"] })`
 *   就同時重載清單、明細與證明（三者都在此前綴下）。
 * - `placeholderData: (prev) => prev`；不寫 `retry`（共用謂詞在 `lib/query-client.ts`）。
 * - `queryFn` 回傳 proto 原生回應；**沒有排序參數**（`ListReturnRequestsRequest` 未定義
 *   `sort`/`desc`，後端也不解析 → 本頁不開排序，同 `PrintPage`）。
 * - 狀態篩選在頁面是 signal「草稿」，**只在 submit 時**進 query key。
 * - 客戶名稱跨 feature 復用 `customers/queries` 的 `customerClient`（同 `OrdersPage`
 *   匯入 `customerDropdownQueryOptions` 的先例）；查不到時頁面退回「客戶 #id」，
 *   明細對話框不因客戶查詢失敗而打不開。
 */
export const returnClient = createClient(ReturnService, transport);

/** 退貨清單每頁筆數（後端上限 100，與其他清單同用 20）。 */
export const RETURN_PAGE_SIZE = 20;

/** 退貨清單查詢參數（全部參數都進 queryKey）。 */
export interface ReturnRequestsParams {
  /** 空字串＝全部；其餘合法值 `pending`/`approved`/`rejected`（其他值後端回 InvalidArgument）。 */
  status: string;
  page: number;
  pageSize: number;
}

/** 退貨清單查詢選項；`createQuery(() => returnRequestsQueryOptions(params))`。 */
export const returnRequestsQueryOptions = (params: ReturnRequestsParams) =>
  queryOptions<ListReturnRequestsResponse>({
    queryKey: [
      "returnRequests",
      { status: params.status, page: params.page, pageSize: params.pageSize },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      returnClient.listReturnRequests({
        status: params.status,
        page: params.page,
        pageSize: params.pageSize,
      }),
  });

/**
 * 退貨明細查詢選項。`version`（樂觀鎖）只存在於這份回應裡 →
 * 審核的 `expectedVersion` 必須取自這裡，絕不可寫死。
 */
export const returnRequestQueryOptions = ({ id }: { id: string }) =>
  queryOptions<GetReturnRequestResponse>({
    queryKey: ["returnRequests", "detail", { id }],
    placeholderData: (prev) => prev,
    queryFn: () => returnClient.getReturnRequest({ id }),
  });

/** 退貨證明查詢選項（僅 `approved` 有證明；其餘狀態後端回「尚無證明可出示」）。 */
export const returnCertificateQueryOptions = ({ id }: { id: string }) =>
  queryOptions<GetReturnCertificateResponse>({
    queryKey: ["returnRequests", "cert", { id }],
    placeholderData: (prev) => prev,
    queryFn: () => returnClient.getReturnCertificate({ id }),
  });

/** 客戶名稱查詢（明細對話框用）；失敗時頁面退回「客戶 #id」，不阻斷審核。 */
export const customerNameQueryOptions = ({ id }: { id: string }) =>
  queryOptions({
    queryKey: ["customers", "name", { id }],
    placeholderData: (prev) => prev,
    queryFn: () => customerClient.getCustomer({ id }),
  });
