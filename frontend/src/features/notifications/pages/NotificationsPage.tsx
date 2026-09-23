import { Code, ConnectError } from "@connectrpc/connect";
import { createQuery, useQueryClient } from "@tanstack/solid-query";
import {
  createColumnHelper,
  createTable,
  flexRender,
  rowPaginationFeature,
  tableFeatures,
  type PaginationState,
} from "@tanstack/solid-table";
import { batch, createEffect, createSignal, For, Show, type JSX } from "solid-js";
import type { NotificationView } from "~/lib/proto/salesorder/v1/notifications_pb";
import {
  Badge,
  Button,
  Card,
  Field,
  FieldLabel,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui";
import { ListPagination } from "../../users/components/ListPagination";
import { queryData } from "~/lib/query-data";
import {
  NOTIFICATION_PAGE_SIZE,
  notificationClient,
  notificationsQueryOptions,
} from "../queries";

/**
 * 通知表格的 table 功能集：**只有分頁**。
 *
 * `ListNotificationsRequest` 沒有 `sort`/`desc`（見 proto）→ 這頁不開排序（同 `PrintPage`）；
 * 後端固定 `created_at` 倒序，最新的在最上面。
 */
const NOTIFICATION_TABLE_FEATURES = tableFeatures({ rowPaginationFeature });

const notificationColumnHelper = createColumnHelper<
  typeof NOTIFICATION_TABLE_FEATURES,
  NotificationView
>();

/** 沒有資料時的穩定空陣列：`data` 每次回傳新陣列會讓 table 的 row model 反覆重算。 */
const NO_NOTIFICATIONS: NotificationView[] = [];

/**
 * 狀態標籤與樣式。
 *
 * 後端四個狀態（`notifications.status`）：`pending` 已入列尚未送出、`sent` 已送達通道、
 * `read` 已讀、`failed` 送達失敗。前端把前兩個合併顯示成「未讀」—— 它們對使用者的
 * 意義相同（我還沒看），而「待發送 vs 已發送」是後台內部進度，列在使用者頁面只是噪音；
 * `failed` 必須單獨顯示，因為那代表**對方沒收到**（且 `MarkRead` 刻意不讓它轉已讀）。
 */
const STATUS_META: Record<string, { label: string; variant: "warning" | "success" | "destructive" }> = {
  pending: { label: "未讀", variant: "warning" },
  sent: { label: "未讀", variant: "warning" },
  read: { label: "已讀", variant: "success" },
  failed: { label: "送達失敗", variant: "destructive" },
};

/** 通道標籤（proto 是字串而非 enum）。 */
const CHANNEL_LABELS: Record<string, string> = {
  in_app: "站內",
  fcm: "推播",
};

/**
 * 未讀＝`pending`/`sent`（與後端 `MarkRead` 可轉已讀的狀態同一組）；
 * `failed` 不算未讀也**不可**轉已讀，`read` 已是終態。
 */
const UNREAD_STATUSES = new Set(["pending", "sent"]);

/** 錯誤訊息對照；樣板 = `CustomersPage`/`PrintPage` 的同一份 switch。 */
function errorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    switch (err.code) {
      case Code.NotFound:
        return "資料不存在或已被刪除";
      case Code.InvalidArgument:
        return err.rawMessage || "輸入資料有誤,請檢查後再試";
      case Code.FailedPrecondition:
        return err.rawMessage || "資料狀態不允許此操作";
      case Code.PermissionDenied:
        return "沒有權限執行此操作";
      case Code.Unauthenticated:
        return "請先登入";
      case Code.Unavailable:
        return "無法連線至伺服器,請確認後端服務已啟動";
      default:
        return err.rawMessage || "操作失敗,請稍後再試";
    }
  }
  return "無法連線至伺服器,請確認後端服務已啟動";
}

/** 日期時間顯示已在 createdAt 欄位內聯（單一使用點，不另立包裝函式）。 */

/**
 * 通知中心頁(/notifications)。
 *
 * 版型同 `CustomersPage`/`PrintPage`（Page Headings + 篩選卡片色帶 + 表格 `Card`）。
 *
 * 三處契約：
 * - **總數在 `total`**（`ListNotificationsResponse` 沒有 `pagination`；同 orders/print）。
 * - **無排序**（`ListNotificationsRequest` 未定義 `sort`/`desc`）。
 * - **資料是本人的**：`NotificationService` 只回 `user_id` = 自己的列，他人的通知在
 *   `MarkRead` 視同 `not_found` → 本頁不做任何「看別人的通知」入口。
 *
 * 標記已讀有兩種粒度：**單列**（點那列的「標記已讀」）與**本頁**（把目前這頁的未讀一次
 * 標掉）。刻意**沒有「全部已讀」**：後端 `MarkRead` 只接受明列的 id（上限 100）且沒有
 * 「列出全部未讀 id」的 RPC，做「全部」必然只標到本頁卻宣稱全部 —— 那是會騙人的按鈕。
 */
export default function NotificationsPage() {
  // 篩選草稿：改下拉只動這顆 signal，不進 query key。
  const [unreadOnlyDraft, setUnreadOnlyDraft] = createSignal(false);
  const [channelDraft, setChannelDraft] = createSignal("");
  // 已套用的篩選是 query key 的來源：只有送出篩選會動它。
  const [filter, setFilter] = createSignal({ unreadOnly: false, channel: "" });
  // 頁碼與每頁筆數的唯一真相＝table 的 pagination state（受控）。
  const [pagination, setPagination] = createSignal<PaginationState>({
    pageIndex: 0,
    pageSize: NOTIFICATION_PAGE_SIZE,
  });
  // 操作訊息：成功回饋與失敗原因（兩者互斥，同一時間只會有一個）。
  const [actionError, setActionError] = createSignal<string | null>(null);
  const [successMessage, setSuccessMessage] = createSignal<string | null>(null);
  // 標記中的列 id：防止連點重送（也讓按鈕有明確的按下載入感）。
  const [marking, setMarking] = createSignal<ReadonlySet<string>>(new Set());

  const client = useQueryClient();

  const columns = notificationColumnHelper.columns([
    notificationColumnHelper.accessor("status", {
      header: "狀態",
      cell: (info) => {
        const meta = STATUS_META[info.getValue()];
        return (
          <Badge variant={meta?.variant ?? "secondary"}>{meta?.label ?? info.getValue()}</Badge>
        );
      },
    }),
    notificationColumnHelper.accessor("title", {
      header: "標題",
      cell: (info) => (
        <div class="min-w-0">
          <div class="font-medium text-foreground">{info.getValue() || "（無標題）"}</div>
          <div class="truncate text-sm text-muted-foreground">{info.row.original.content}</div>
        </div>
      ),
    }),
    notificationColumnHelper.accessor("channel", {
      header: "通道",
      cell: (info) => (
        <span class="text-muted-foreground">{CHANNEL_LABELS[info.getValue()] ?? info.getValue()}</span>
      ),
    }),
    notificationColumnHelper.accessor("createdAt", {
      header: "建立時間",
      cell: (info) => {
        const value = info.getValue();
        // RFC3339 的 `T` 換成空白、秒後切除（同 `PrintPage` 的列印時間）。
        return <span class="text-muted-foreground">{value ? value.slice(0, 19).replace("T", " ") : "—"}</span>;
      },
    }),
    notificationColumnHelper.display({
      id: "actions",
      header: () => <span class="block text-right">操作</span>,
      cell: (info) => (
        <div class="text-right">
          <Show when={UNREAD_STATUSES.has(info.row.original.status)}>
            <button
              type="button"
              disabled={marking().has(info.row.original.id)}
              onClick={() => void markRead([info.row.original.id])}
              class="font-medium text-primary hover:underline disabled:cursor-not-allowed disabled:opacity-60"
            >
              標記已讀
            </button>
          </Show>
        </div>
      ),
    }),
  ]);

  const query = createQuery(() =>
    notificationsQueryOptions({
      unreadOnly: filter().unreadOnly,
      channel: filter().channel,
      page: pagination().pageIndex + 1,
      pageSize: pagination().pageSize,
    })
  );

  const total = () => queryData(query, (d) => d?.total) ?? 0;
  const unreadCount = () => queryData(query, (d) => d?.unreadCount) ?? 0;

  const table = createTable({
    features: NOTIFICATION_TABLE_FEATURES,
    columns,
    get data() {
      return queryData(query, (d) => d?.notifications) ?? NO_NOTIFICATIONS;
    },
    get rowCount() {
      return total();
    },
    manualPagination: true,
    get state() {
      return { pagination: pagination() };
    },
    onPaginationChange: setPagination,
  });

  /** 超頁退回：total 讓目前頁碼超界時把頁碼夾到合法值（`page` 在 key 內 → 夾了就重取）。 */
  createEffect(() => {
    if (query.isPlaceholderData || !query.data) return;
    const maxPage = Math.max(1, Math.ceil(total() / pagination().pageSize));
    if (pagination().pageIndex + 1 > maxPage) table.setPageIndex(maxPage - 1);
  });

  /** 送出篩選：把草稿套進 query key 並回第 1 頁（同一個 `batch` 內，避免兩次 RPC）。 */
  const submitFilter: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    setActionError(null);
    setSuccessMessage(null);
    batch(() => {
      setFilter({ unreadOnly: unreadOnlyDraft(), channel: channelDraft() });
      table.setPageIndex(0);
    });
  };

  /**
   * 標記已讀：一次 RPC 帶整批 id（後端上限 100，本頁最多 20 → 不會超）。
   * 成功後失效整個 `["notifications"]` 前綴，讓清單（狀態）與未讀數同時刷新。
   */
  const markRead = async (ids: string[]) => {
    if (ids.length === 0) return;
    setActionError(null);
    setSuccessMessage(null);
    setMarking(new Set(ids));
    try {
      const res = await notificationClient.markRead({ notificationIds: ids });
      setSuccessMessage(
        res.markedCount === ids.length
          ? `已標記 ${res.markedCount} 筆已讀`
          : `已標記 ${res.markedCount} 筆已讀（其餘狀態不需或無法轉為已讀）`
      );
      await client.invalidateQueries({ queryKey: ["notifications"] });
    } catch (err) {
      setActionError(errorMessage(err));
    } finally {
      setMarking(new Set<string>());
    }
  };

  /** 本頁未讀（後端只認 `pending`/`sent` 可轉已讀，`failed` 刻意排除）。 */
  const pageUnreadIds = () =>
    (queryData(query, (d) => d?.notifications) ?? []).filter((n) => UNREAD_STATUSES.has(n.status)).map((n) => n.id);

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">通知中心</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            未讀 {unreadCount()} 筆（共 {total()} 筆）
          </p>
        </div>
        <Show when={pageUnreadIds().length > 0}>
          <Button
            type="button"
            variant="outline"
            disabled={marking().size > 0}
            onClick={() => void markRead(pageUnreadIds())}
          >
            標記本頁已讀（{pageUnreadIds().length}）
          </Button>
        </Show>
      </header>

      <Show when={query.error}>
        {(err) => (
          <p
            class="mb-4 rounded-lg bg-destructive/15 px-3 py-2 text-sm font-medium text-destructive"
            role="alert"
          >
            {errorMessage(err())}
          </p>
        )}
      </Show>

      <Show when={actionError()}>
        {(message) => (
          <p
            class="mb-4 rounded-lg bg-destructive/15 px-3 py-2 text-sm font-medium text-destructive"
            role="alert"
          >
            {message()}
          </p>
        )}
      </Show>

      <Show when={successMessage()}>
        {(message) => (
          <p
            class="mb-4 rounded-lg bg-success/15 px-3 py-2 text-sm font-medium text-success"
            role="status"
          >
            {message()}
          </p>
        )}
      </Show>

      <Card>
        <form
          class="flex flex-wrap items-end gap-3 border-b border-border bg-muted px-3 py-3"
          onSubmit={submitFilter}
        >
          <Field class="w-full sm:w-44">
            <FieldLabel for="notification-unread">已讀狀態</FieldLabel>
            <select
              id="notification-unread"
              value={unreadOnlyDraft() ? "unread" : ""}
              onChange={(e) => setUnreadOnlyDraft(e.currentTarget.value === "unread")}
            >
              <option value="">全部</option>
              <option value="unread">只看未讀</option>
            </select>
          </Field>
          <Field class="w-full sm:w-40">
            <FieldLabel for="notification-channel">通道</FieldLabel>
            <select
              id="notification-channel"
              value={channelDraft()}
              onChange={(e) => setChannelDraft(e.currentTarget.value)}
            >
              <option value="">全部</option>
              <option value="in_app">站內</option>
              <option value="fcm">推播</option>
            </select>
          </Field>
          <Button type="submit" variant="outline">
            查詢
          </Button>
        </form>

        <Table>
          <TableHeader>
            <For each={table.getHeaderGroups()}>
              {(hg) => (
                <TableRow>
                  <For each={hg.headers}>
                    {(h) => (
                      <TableHead>{flexRender(h.column.columnDef.header, h.getContext())}</TableHead>
                    )}
                  </For>
                </TableRow>
              )}
            </For>
          </TableHeader>
          <TableBody>
            <Show
              when={queryData(query, (d) => d?.notifications?.length)}
              fallback={
                <TableRow>
                  <TableCell colSpan={columns.length}>
                    {query.isFetching ? "載入中…" : "尚無通知"}
                  </TableCell>
                </TableRow>
              }
            >
              <For each={table.getRowModel().rows}>
                {(row) => (
                  <TableRow>
                    <For each={row.getAllCells()}>
                      {(cell) => (
                        <TableCell>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                      )}
                    </For>
                  </TableRow>
                )}
              </For>
            </Show>
          </TableBody>
        </Table>

        <ListPagination
          total={table.getRowCount()}
          pageSize={table.atoms.pagination.get().pageSize}
          page={table.atoms.pagination.get().pageIndex + 1}
          onPageChange={(page) => table.setPageIndex(page - 1)}
        />
      </Card>
    </main>
  );
}
