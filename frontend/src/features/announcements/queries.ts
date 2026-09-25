import { createClient } from "@connectrpc/connect";
import { queryOptions } from "@tanstack/solid-query";
import {
  AnnouncementService,
  type ListActiveAnnouncementsResponse,
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

/**
 * 前台公告查詢（spec「前台展示與排序」）：**只回當下可見**者（後端已套用
 * is_active / 上下架時間窗 / 平台投放過濾），故與管理列表是兩支 RPC、兩組 queryKey。
 *
 * 不共用 `announcementsQueryOptions`：管理面要看得到未上架/停用（否則無法預覽編輯），
 * 前台只能看到已上架的 —— 共用一個 queryKey 會讓首頁命中管理頁的快取而顯示未上架公告。
 */
export const activeAnnouncementsQueryOptions = (platform: "web" | "app") =>
  queryOptions<ListActiveAnnouncementsResponse>({
    // platform 進 queryKey：Web 與 App 的投放集合不同，快取不可互用。
    queryKey: ["announcements", "active", platform],
    queryFn: () => announcementClient.listActiveAnnouncements({ platform }),
  });
