import { createClient } from "@connectrpc/connect";
import {
  infiniteQueryOptions,
  queryOptions,
  type InfiniteData,
  type QueryKey,
} from "@tanstack/solid-query";
import {
  CompanyService,
  DepartmentService,
  type ListCompaniesResponse,
} from "~/lib/proto/salesorder/v1/company_pb";
import { RoleService } from "~/lib/proto/salesorder/v1/role_pb";
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
 * - **排序（D1）**：`sort`（白名單欄位名）＋`desc` 一律**顯式**進 key 與請求，未排序時送
 *   `sort: ""`／`desc: false`。理由：(a) `sort` 空字串就是後端契約裡的「服務預設排序（並忽略
 *   `desc`）」，是這個參數的**正典空值**，不必用 `undefined` 再表示一次；(b) key 因此對每個
 *   參數都有值，兩顆不同 key 不會因為「省略 vs undefined」而語意重疊；(c) 前端永遠不會產生
 *   「`sort` 空＋`desc: true`」這種後端會忽略的組合（未排序時 table 的 sorting state 為空）。
 *   `desc` 在 `sort` 為空時仍照送，只是後端會忽略——請求 payload 與後端語意一致。
 * - **累積式下拉**（部門頁的「所屬公司」）用 `infiniteQueryOptions`＋`createInfiniteQuery`：
 *   key 為 `["companies", "options", { … }]`（仍以 `["companies"]` 為失效前綴），
 *   `getNextPageParam` 由「已載入頁數 × 每頁筆數 vs 總筆數」決定；頁面把各頁攤平後去重。
 *
 * 頁面用法（3B 之後的實際形狀；`pagination`／`sorting` 都是 table 的 state getter）：
 *   const query = createQuery(() =>
 *     companiesQueryOptions({
 *       page: pagination().pageIndex + 1,
 *       pageSize: pagination().pageSize,
 *       sort: sorting()[0]?.id ?? "",
 *       desc: sorting()[0]?.desc ?? false,
 *       keyword: filter().keyword || undefined,
 *       status: filter().status || undefined,
 *     })
 *   );
 *   const options = createInfiniteQuery(() =>
 *     companyDropdownQueryOptions({ keyword: companySearch() || undefined })
 *   );
 */

/** 清單每頁筆數（三張清單共用同一個值）。 */
export const PAGE_SIZE = 20;

/** 公司服務 client：清單查詢與 modal 的建立／更新／刪除共用同一個實例。 */
export const companyClient = createClient(CompanyService, transport);

/** 公司清單查詢參數（全部參數都進 queryKey）。 */
export interface CompanyListParams {
  page: number;
  pageSize: number;
  /** 排序白名單欄位名（`name`/`identifier`/`tax_id`/`id`）；**空字串＝服務預設排序**。 */
  sort: string;
  /** 是否降冪；後端在 `sort` 為空時忽略（仍照送，見檔頭契約）。 */
  desc: boolean;
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
        sort: params.sort,
        desc: params.desc,
        keyword: params.keyword,
        status: params.status,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      companyClient.listCompanies({
        page: params.page,
        pageSize: params.pageSize,
        sort: params.sort,
        desc: params.desc,
        keyword: params.keyword,
        status: params.status,
      }),
  });

/** 累積式公司下拉（部門頁「所屬公司」）每頁筆數：比清單的 20 筆大，減少「載入更多」次數。 */
const COMPANY_DROPDOWN_PAGE_SIZE = 50;

/** 公司下拉查詢參數（全部參數都進 queryKey）。 */
export interface CompanyDropdownParams {
  keyword?: string;
}

/**
 * 公司下拉查詢選項（累積式）；`createInfiniteQuery(() => companyDropdownQueryOptions(params))`。
 *
 * `getNextPageParam` 以「已載入筆數 vs 總筆數」判斷是否還有下一頁（等同改寫前「載入更多」的
 * 停用條件 `已載入 >= 總筆數`）；回 `undefined` 即 `hasNextPage === false`。
 */
export const companyDropdownQueryOptions = (params: CompanyDropdownParams) =>
  infiniteQueryOptions<
    ListCompaniesResponse,
    Error,
    InfiniteData<ListCompaniesResponse>,
    QueryKey,
    number
  >({
    queryKey: [
      "companies",
      "options",
      { pageSize: COMPANY_DROPDOWN_PAGE_SIZE, keyword: params.keyword },
    ],
    initialPageParam: 1,
    placeholderData: (prev) => prev,
    queryFn: ({ pageParam }) =>
      companyClient.listCompanies({
        page: pageParam,
        pageSize: COMPANY_DROPDOWN_PAGE_SIZE,
        keyword: params.keyword,
      }),
    getNextPageParam: (lastPage, allPages) =>
      allPages.reduce((loaded, page) => loaded + page.companies.length, 0) <
      Number(lastPage.pagination?.total ?? 0)
        ? allPages.length + 1
        : undefined,
  });

/** 部門服務 client：清單查詢與 modal 的建立／更新／刪除共用同一個實例。 */
export const departmentClient = createClient(DepartmentService, transport);

/** 部門清單查詢參數（全部參數都進 queryKey）。 */
export interface DepartmentListParams {
  page: number;
  pageSize: number;
  /** 排序白名單欄位名（`name`/`id`）；**空字串＝服務預設排序**。 */
  sort: string;
  /** 是否降冪；後端在 `sort` 為空時忽略（仍照送，見檔頭契約）。 */
  desc: boolean;
  companyId?: string;
}

/** 部門清單查詢選項；`createQuery(() => departmentsQueryOptions(params))`。 */
export const departmentsQueryOptions = (params: DepartmentListParams) =>
  queryOptions({
    queryKey: [
      "departments",
      {
        page: params.page,
        pageSize: params.pageSize,
        sort: params.sort,
        desc: params.desc,
        companyId: params.companyId,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      departmentClient.listDepartments({
        page: params.page,
        pageSize: params.pageSize,
        sort: params.sort,
        desc: params.desc,
        companyId: params.companyId,
      }),
  });

/** 角色服務 client：清單查詢與權限矩陣的讀（`getRolePermissions`）／寫（`updateRolePermissions`）共用同一個實例。 */
export const roleClient = createClient(RoleService, transport);

/** 角色清單查詢參數（全部參數都進 queryKey）；角色沒有篩選條件，只有分頁與排序。 */
export interface RoleListParams {
  page: number;
  pageSize: number;
  /** 排序白名單欄位名（`code`/`name`/`id`）；**空字串＝服務預設排序**（`id ASC`）。 */
  sort: string;
  /** 是否降冪；後端在 `sort` 為空時忽略（仍照送，見檔頭契約）。 */
  desc: boolean;
}

/** 角色清單查詢選項；`createQuery(() => rolesQueryOptions(params))`。 */
export const rolesQueryOptions = (params: RoleListParams) =>
  queryOptions({
    queryKey: ["roles", { page: params.page, pageSize: params.pageSize, sort: params.sort, desc: params.desc }],
    placeholderData: (prev) => prev,
    queryFn: () =>
      roleClient.listRoles({
        page: params.page,
        pageSize: params.pageSize,
        sort: params.sort,
        desc: params.desc,
      }),
  });
