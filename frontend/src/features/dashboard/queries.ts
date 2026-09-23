import { createClient } from "@connectrpc/connect";
import { queryOptions } from "@tanstack/solid-query";
import {
  SalesOrderService,
  type ListOrdersResponse,
} from "~/lib/proto/salesorder/v1/salesorder_pb";
import { transport } from "~/lib/transport";

/**
 * 首頁 Dashboard 查詢契約；樣板 = `features/orders/queries.ts` 檔頭契約（不另發明）。
 *
 * 首頁要的是「今日待出貨」與「待處理訂單數量」（規格 §8.2），兩者都是 `ListOrders` 的既有
 * 篩選，因此**不新增後端 RPC** —— `total` 由後端算（`ListOrders` 先 `Count` 再取頁），
 * 所以每個數字都只是一頁一筆的查詢（`pageSize: 1`），不必把清單撈回來自己數。
 *
 * 三處契約：
 * - **`pageSize: 1`**：只需要 `total`，列本身不用（`orders` 會被忽略）。
 * - **日期用本地日**：`expected_delivery_date` 是 `YYYY-MM-DD`（§82），值由呼叫端的 `today()`
 *   產生 —— 與派車看板同一套 local-date 慣例，不可改用 UTC 日期（跨時區會少一天）。
 * - **一個查詢只有一個 status**：`ListOrdersRequest.status` 是單一字串，不是重複欄位，
 *   所以「pending 或 processing」這種集合條件必須拆成兩支查詢後相加（見 `DashboardPage`）。
 */
export const salesOrderClient = createClient(SalesOrderService, transport);

/** 首頁查詢只取總數，列不取（見檔頭）。 */
const COUNT_ONLY = 1;

/**
 * 訂單數查詢（單一狀態 + 可選日期）。
 *
 * `date` 給空字串即「不限日期」—— 這與「日期＝今天」是不同問題：
 * 前者問「還沒動的有幾張」（含未來日期的單），後者問「今天要送的」。
 */
export const orderCountQueryOptions = (params: { key: string; status: string; date: string }) =>
  queryOptions<ListOrdersResponse>({
    queryKey: ["dashboard", params.key, { status: params.status, date: params.date }],
    queryFn: () =>
      salesOrderClient.listOrders({
        page: 1,
        pageSize: COUNT_ONLY,
        status: params.status,
        customerId: "",
        source: "",
        keyword: "",
        includeDeleted: false,
        expectedDeliveryDate: params.date,
      }),
  });

/** 今日待出貨拆成兩支：`pending` 與 `processing`（單一 status 欄位無法表達集合，見檔頭）。 */
export const todayPendingQueryOptions = (date: string) =>
  orderCountQueryOptions({ key: "todayPending", status: "pending", date });

export const todayProcessingQueryOptions = (date: string) =>
  orderCountQueryOptions({ key: "todayProcessing", status: "processing", date });

/** 待處理訂單數：`status = pending`，**不限日期**（包含未來日期的單）。 */
export const pendingOrderCountQueryOptions = () =>
  orderCountQueryOptions({ key: "pendingCount", status: "pending", date: "" });

/** 處理中訂單數：`status = processing`（已派車、完成前），不限日期。 */
export const processingOrderCountQueryOptions = () =>
  orderCountQueryOptions({ key: "processingCount", status: "processing", date: "" });
