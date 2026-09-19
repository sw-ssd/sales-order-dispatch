import { createClient } from "@connectrpc/connect";
import { queryOptions } from "@tanstack/solid-query";
import { CompanyService } from "~/lib/proto/salesorder/v1/company_pb";
import { transport } from "~/lib/transport";

/**
 * 三張清單（公司／部門／角色）的共用查詢慣例；樣板 = `lib/ability/service.ts`。
 *
 * 契約（T3／T4 沿用同一套，不要另發明）：
 * - `queryKey` 一律 `[<複數資源名>, { …全部查詢參數 }]`：**每個參數都進 key**（含 `pageSize`），
 *   前綴即失效單位 → mutation 成功後 `invalidateQueries({ queryKey: ["companies"] })`。
 * - `placeholderData: (prev) => prev`（v5 的 keepPreviousData 等價物）：換頁／換篩選時
 *   保留前一筆結果，不閃空；據此，「目前資料是否屬於當前 key」一律看 `isPlaceholderData`。
 * - **不寫 `retry`**：`lib/query-client.ts` 已把共用謂詞接在 `defaultOptions.queries.retry`
 *   （只重試暫時性錯誤、封頂 3 次），逐個 `queryOptions` 再寫一份會變成兩份真相。
 * - `queryFn` 回傳 proto 原生回應，頁面自行由 `data.pagination.total` 推導總筆數；
 *   頁面**不得**再留一份 signal 快取（兩份真相）。
 * - 篩選條件在頁面是 signal「草稿」，**只在 submit 時**進 query key；`page` 由頁面持有
 *   （3B 起改由 table 的 pagination state 持有，query key 一律取自該狀態）。
 *
 * 頁面用法：
 *   const query = createQuery(() => companiesQueryOptions({ page: page(), pageSize: PAGE_SIZE, … }));
 */

/** 清單每頁筆數（三張清單共用同一個值）。 */
export const PAGE_SIZE = 20;

/** 公司服務 client：清單查詢與 modal 的建立／更新／刪除共用同一個實例。 */
export const companyClient = createClient(CompanyService, transport);

/** 公司清單查詢參數（全部參數都進 queryKey）。 */
export interface CompanyListParams {
  page: number;
  pageSize: number;
  keyword?: string;
  status?: string;
}

/** 公司清單查詢選項；`createQuery(() => companiesQueryOptions(params))`。 */
export const companiesQueryOptions = (params: CompanyListParams) =>
  queryOptions({
    queryKey: [
      "companies",
      {
        page: params.page,
        pageSize: params.pageSize,
        keyword: params.keyword,
        status: params.status,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      companyClient.listCompanies({
        page: params.page,
        pageSize: params.pageSize,
        keyword: params.keyword,
        status: params.status,
      }),
  });
