import { createClient } from "@connectrpc/connect";
import { queryOptions } from "@tanstack/solid-query";
import {
  AnnouncementService,
  type ListAnnouncementsResponse,
} from "~/lib/proto/salesorder/v1/announcement_pb";
import { transport } from "~/lib/transport";

/**
 * 公告 CMS 查詢契約；樣板 = `features/users/queries.ts` 檔頭契約（不另發明）。
 *
 * `ListAnnouncementsRequest` 沒有 `sort`/`desc`（見 proto）→ 本頁刻意不開排序；
 * 總數在 `total`（`ListAnnouncementsResponse.total`,int32）。
 * 管理列表只列**可管理範圍**（後端收斂）,且預設排除已刪除（spec:無復原入口,故前端
 * 不提供「含已刪除」開關）。
 */
export const announcementClient = createClient(AnnouncementService, transport);

/** 清單每頁筆數（與其他清單共用同一個值）。 */
export const ANNOUNCEMENTS_PAGE_SIZE = 20;

/** 公告清單查詢參數（全部參數都進 queryKey）。 */
export interface AnnouncementListParams {
  page: number;
  pageSize: number;
  type?: string;
}

/** 公告清單查詢選項；`createQuery(() => announcementsQueryOptions(params))`。 */
export const announcementsQueryOptions = (params: AnnouncementListParams) =>
  queryOptions<ListAnnouncementsResponse>({
    queryKey: [
      "announcements",
      {
        page: params.page,
        pageSize: params.pageSize,
        type: params.type,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      announcementClient.listAnnouncements({
        page: params.page,
        pageSize: params.pageSize,
        type: params.type ?? "",
        includeDeleted: false,
      }),
  });
