import { createClient } from "@connectrpc/connect";
import { queryOptions } from "@tanstack/solid-query";
import {
  NotificationService,
  type ListNotificationsResponse,
} from "~/lib/proto/salesorder/v1/notifications_pb";
import { transport } from "~/lib/transport";

/**
 * 通知中心查詢契約；樣板 = `features/printing/queries.ts` 檔頭契約（不另發明）：
 * - `queryKey` 一律 `["notifications", { …全部查詢參數 }]`，前綴 `["notifications"]`
 *   即失效單位 → 標記已讀後一次 `invalidateQueries({ queryKey: ["notifications"] })`
 *   就同時重載清單與未讀數。
 * - `placeholderData: (prev) => prev`；不寫 `retry`（共用謂詞在 `lib/query-client.ts`）。
 * - `queryFn` 回傳 proto 原生回應；**沒有排序參數**（`ListNotificationsRequest` 未定義
 *   `sort`/`desc`，後端固定 `created_at` 倒序）→ 本頁不開排序，同 `PrintPage`。
 * - 篩選在頁面是 signal「草稿」，**只在 submit 時**進 query key。
 *
 * **不另呼叫 `UnreadCount`**：`ListNotificationsResponse` 自帶 `unread_count`，
 * 同一次列表請求就能顯示未讀數 —— 再開一支 RPC 等於讓兩個數字各有自己的快取時序，
 * 標記已讀後會看到角標與列表互相打架（規格 §5.4 的角標輪詢是給 App 掛在鈴上的，
 * 本頁沒有那個入口）。
 */
export const notificationClient = createClient(NotificationService, transport);

/** 通知清單每頁筆數（後端上限 100，與其他清單同用 20）。 */
export const NOTIFICATION_PAGE_SIZE = 20;

/** 通知清單查詢參數（全部參數都進 queryKey）。 */
export interface NotificationListParams {
  /** 只看未讀（後端把 `pending`/`sent` 視為未讀；`read`/`failed` 不算）。 */
  unreadOnly: boolean;
  /** 空字串＝全部；其餘合法值 `in_app`/`fcm`（其他值後端回 InvalidArgument）。 */
  channel: string;
  page: number;
  pageSize: number;
}

/** 通知清單查詢選項；`createQuery(() => notificationsQueryOptions(params))`。 */
export const notificationsQueryOptions = (params: NotificationListParams) =>
  queryOptions<ListNotificationsResponse>({
    queryKey: [
      "notifications",
      {
        unreadOnly: params.unreadOnly,
        channel: params.channel,
        page: params.page,
        pageSize: params.pageSize,
      },
    ],
    placeholderData: (prev) => prev,
    queryFn: () =>
      notificationClient.listNotifications({
        unreadOnly: params.unreadOnly,
        channel: params.channel,
        page: params.page,
        pageSize: params.pageSize,
      }),
  });
