import { createClient } from "@connectrpc/connect";
import { queryOptions } from "@tanstack/solid-query";
import {
  CustomerService,
  type ListCustomersResponse,
} from "~/lib/proto/customers/v1/customer_pb";
import {
  RouteService,
  WarehouseService,
  type ListRoutesResponse,
  type ListWarehousesResponse,
} from "~/lib/proto/masters/v1/master_pb";
import {
  PrintService,
  type ListLogsResponse,
} from "~/lib/proto/products/v1/print_pb";
import { transport } from "~/lib/transport";

/**
 * 單據列印查詢契約；樣板 = `features/users/queries.ts` 檔頭契約（不另發明）。
 *
 * 與其他清單的三處差異：
 * - **`ListLogsRequest` 沒有 `sort`/`desc`**（見 proto）→ 本頁不開排序（後端也不解析）。
 * - 總數在回應的 `total`（不是 `pagination.total`）。
 * - **只涵蓋 `print_logs`**：printing spec §151 要求同時查 `print_previews`，
 *   但 proto 未定義 `ListPreviews`（Go 端 `PrintPreview` 只有 `Create`、無查詢路徑）
 *   → 預覽紀錄列表**待後端補 RPC 後再做**，此處不假裝有。
 */
export const printClient = createClient(PrintService, transport);

/** 列印記錄每頁筆數（後端上限 100，與其他清單同用 20）。 */
export const PRINT_LOG_PAGE_SIZE = 20;

/** 列印記錄查詢參數（全部參數都進 queryKey）。 */
export interface PrintLogsParams {
  page: number;
  pageSize: number;
  /** YYYY-MM-DD，可空（後端 `date_from`/`date_to` 均為此格式）。 */
  dateFrom?: string;
  dateTo?: string;
  documentType?: string;
  routeId?: string;
}

/** 列印記錄查詢選項；`createQuery(() => printLogsQueryOptions(params))`。 */
export const printLogsQueryOptions = (params: PrintLogsParams) =>
  queryOptions<ListLogsResponse>({
    queryKey: [
      "printLogs",
      {
        page: params.page,
        pageSize: params.pageSize,
        dateFrom: params.dateFrom,
        dateTo: params.dateTo,
        documentType: params.documentType,
        routeId: params.routeId,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      printClient.listLogs({
        page: params.page,
        pageSize: params.pageSize,
        dateFrom: params.dateFrom ?? "",
        dateTo: params.dateTo ?? "",
        documentType: params.documentType ?? "",
        routeId: params.routeId ?? "",
      }),
  });

/**
 * 車次查詢選項：**列印頁的車次顯示與選擇都取自這裡**。
 *
 * 為什麼需要它：`PrintRequest.route_id` 與 `ListLogsRequest.route_id` 收的是**內部 id**
 * （car route.id，例 "12"），但全站其他地方只顯示「1車（R01）」這種名稱＋代碼，使用者在畫面上
 * 無從得知 id。原本的作法是把「車次 ID」做成自由輸入框，等於要求使用者背誦一個看不見的值，
 * 打錯只能等後端 `InvalidArgument`。
 *
 * 兩份清單各有用途、不可合併：
 * - `includeDeleted: false` → **列印表單的選項**：只能對現存車次列印。
 * - `includeDeleted: true` → **紀錄表格的顯示與篩選**：`print_logs` 是歷史，車次可能已停用／
 *   刪除，若只查未刪除的，舊紀錄的車次欄會查不到名稱而退回顯示裸 id。
 *
 * `ROUTE_PAGE_SIZE` 與派車看板同值（後端上限）：車次是部門級主檔、筆數有限；超過時表格
 * 退回顯示裸 id（見 `PrintPage` 的 `routeLabel`），不靜默假裝車次不存在。
 */
export const printRouteClient = createClient(RouteService, transport);

/** 車次單頁筆數（與派車看板同值，見檔頭說明）。 */
export const PRINT_ROUTE_PAGE_SIZE = 100;

/** 車次查詢選項；`includeDeleted` 決定是要「可選的車次」還是「含歷史的對照清單」。 */
export const printRoutesQueryOptions = (includeDeleted: boolean) =>
  queryOptions<ListRoutesResponse>({
    queryKey: ["printRoutes", { includeDeleted }],
    queryFn: () =>
      printRouteClient.listRoutes({
        page: 1,
        pageSize: PRINT_ROUTE_PAGE_SIZE,
        keyword: "",
        includeDeleted,
      }),
  });

/**
 * 店家與倉別選項：對點單（`delivery_note`）與揀貨單（`picking_list`）**必填其一**，
 * 而後端的必填檢查在 `Assemble`（空選擇器回 `InvalidArgument`）。
 *
 * 這兩份下拉原本是空殼（只有「不指定」）：<select> 有標籤、有 `FieldLabel`，
 * 卻沒有 `<For>`，所以「對點單／揀貨單」在 UI 上**永遠送不出合法請求**——
 * 使用者只會看到後端回的 InvalidArgument，而沒有任何線索指出是選項沒載入。
 *
 * 只取未刪除的：列印是對現存店家／倉別發生的動作（與車次選項同原則；車次的歷史對照另由
 * `includeDeleted: true` 的那份負責表格顯示）。
 */
export const printCustomerClient = createClient(CustomerService, transport);
export const printWarehouseClient = createClient(WarehouseService, transport);

/** 列印頁下拉單頁筆數（店家與倉別同值；兩者都是營運主檔，筆數有限）。 */
export const PRINT_OPTION_PAGE_SIZE = 100;

/** 店家選項（未刪除）。 */
export const printCustomerOptionsQueryOptions = () =>
  queryOptions<ListCustomersResponse>({
    queryKey: ["printCustomers", { pageSize: PRINT_OPTION_PAGE_SIZE }],
    queryFn: () =>
      printCustomerClient.listCustomers({
        page: 1,
        pageSize: PRINT_OPTION_PAGE_SIZE,
        keyword: "",
        includeDeleted: false,
      }),
  });

/** 倉別選項（未刪除）。 */
export const printWarehouseOptionsQueryOptions = () =>
  queryOptions<ListWarehousesResponse>({
    queryKey: ["printWarehouses", { pageSize: PRINT_OPTION_PAGE_SIZE }],
    queryFn: () =>
      printWarehouseClient.listWarehouses({
        page: 1,
        pageSize: PRINT_OPTION_PAGE_SIZE,
        keyword: "",
        includeDeleted: false,
      }),
  });
