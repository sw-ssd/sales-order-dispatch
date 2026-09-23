import { Code, ConnectError } from "@connectrpc/connect";
import { createQuery } from "@tanstack/solid-query";
import {
  createColumnHelper,
  createTable,
  flexRender,
  rowPaginationFeature,
  tableFeatures,
  type PaginationState,
} from "@tanstack/solid-table";
import { batch, createEffect, createSignal, For, Show, type JSX } from "solid-js";
import type { AuditLog } from "~/lib/proto/audit/v1/audit_pb";
import {
  Badge,
  Button,
  Card,
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  Field,
  FieldLabel,
  Input,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  buttonVariants,
} from "~/components/ui";
import { ListPagination } from "../../users/components/ListPagination";
import { AUDIT_PAGE_SIZE, auditLogsQueryOptions } from "../queries";
import { queryData } from "~/lib/query-data";

/**
 * 稽核表格的 table 功能集：**只有分頁**。
 *
 * `ListAuditLogsRequest` 沒有 `sort`/`desc` → 後端固定 `created_at` 倒序（＋`id`
 * 次序鍵），本頁不開排序，同 `PrintPage`／`ReturnsPage`。
 */
const AUDIT_TABLE_FEATURES = tableFeatures({ rowPaginationFeature });

const auditColumnHelper = createColumnHelper<typeof AUDIT_TABLE_FEATURES, AuditLog>();

/** 沒有資料時的穩定空陣列：`data` 每次回傳新陣列會讓 table 的 row model 反覆重算。 */
const NO_AUDIT_LOGS: AuditLog[] = [];

/**
 * 動作標籤（`audit_logs.action` 的十個合法值，後端 `validAuditActions` 是權威白名單；
 * 下拉只給這十項 → 前端不可能送出白名單外的值）。
 */
const ACTION_LABELS: Record<string, string> = {
  create: "新增",
  update: "更新",
  delete: "刪除",
  login: "登入",
  logout: "登出",
  print: "列印",
  force_logout: "強制登出",
  role_change: "角色變更",
  dispatch_cancel: "取消派車",
  void: "作廢",
};

/** 動作徽章樣式：刪改與強制動作標紅，登入類為中性。 */
const ACTION_VARIANTS: Record<string, "destructive" | "warning" | "info" | "secondary"> = {
  delete: "destructive",
  force_logout: "destructive",
  void: "destructive",
  role_change: "warning",
  dispatch_cancel: "warning",
  update: "info",
  create: "info",
  login: "secondary",
  logout: "secondary",
  print: "secondary",
};

/** 錯誤訊息對照；樣板 = `CustomersPage`/`PrintPage` 的同一份 switch。 */
function errorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    switch (err.code) {
      case Code.NotFound:
        return "資料不存在或已被刪除";
      case Code.InvalidArgument:
        // 後端「無效的 from/to/action」都走這裡（訊息原樣透傳，改寫會蓋掉辨識度）。
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

/** 快照是 JSON 字串；解析失敗就原樣顯示（快照由後端 marshal，失敗代表內容異常）。 */
function prettyJson(raw: string): string {
  if (!raw) return "";
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

/**
 * 稽核日誌頁(/audit)。
 *
 * 版型同 `CustomersPage`/`PrintPage`（Page Headings + 篩選卡片色帶 + 表格 `Card`）。
 *
 * 三處契約：
 * - **總數在 `pagination.total`**（int64 → `Number()`），同 customers/users。
 * - **無排序**（`ListAuditLogsRequest` 未定義 `sort`/`desc`）。
 * - **時間是 RFC3339**：`from`/`to` 都空時後端套「近 3 個月」（D27 保留窗）——
 *   本頁不自行算三個月再送，避免前後端各有一份保留期定義。
 *
 * 只讀：稽核紀錄由各服務在**同一業務交易**內寫入（D18），前端沒有任何修改入口；
 * 明細（before/after 快照）來自**清單那一列自帶的欄位**，不另開 RPC ——
 * `AuditService` 只有 `ListAuditLogs`，沒有按 id 查詢的接口。
 */
export default function AuditPage() {
  // 篩選草稿：改輸入只動這顆 signal，不進 query key。
  const [filterDraft, setFilterDraft] = createSignal({
    from: "",
    to: "",
    action: "",
    resourceType: "",
    resourceId: "",
    userId: "",
  });
  // 已套用的篩選是 query key 的來源：只有送出篩選會動它。
  const [filter, setFilter] = createSignal({
    from: "",
    to: "",
    action: "",
    resourceType: "",
    resourceId: "",
    userId: "",
  });
  // 頁碼與每頁筆數的唯一真相＝table 的 pagination state（受控）。
  const [pagination, setPagination] = createSignal<PaginationState>({
    pageIndex: 0,
    pageSize: AUDIT_PAGE_SIZE,
  });
  // 明細對話框：null＝關閉；資料取自列上（無第二支 RPC）。
  const [detail, setDetail] = createSignal<AuditLog | null>(null);

  const columns = auditColumnHelper.columns([
    auditColumnHelper.accessor("createdAt", {
      header: "時間",
      cell: (info) => {
        const value = info.getValue();
        // RFC3339 的 `T` 換成空白、秒後切除（同 `PrintPage` 的列印時間）。
        return <span class="text-muted-foreground">{value ? value.slice(0, 19).replace("T", " ") : "—"}</span>;
      },
    }),
    auditColumnHelper.accessor("userName", {
      header: "操作者",
      cell: (info) => (
        <span class="text-foreground">
          {/* 查不到使用者（已刪除帳號）時退回 id，避免整欄空白。 */}
          {info.getValue() || `#${info.row.original.userId}`}
        </span>
      ),
    }),
    auditColumnHelper.accessor("action", {
      header: "動作",
      cell: (info) => (
        <Badge variant={ACTION_VARIANTS[info.getValue()] ?? "secondary"}>
          {ACTION_LABELS[info.getValue()] ?? info.getValue()}
        </Badge>
      ),
    }),
    auditColumnHelper.accessor("resourceType", {
      header: "資源",
      cell: (info) => {
        const id = info.row.original.resourceId;
        return (
          <span class="text-muted-foreground">
            {info.getValue() ? `${info.getValue()}${id ? ` #${id}` : ""}` : "—"}
          </span>
        );
      },
    }),
    auditColumnHelper.accessor("ipAddress", {
      header: "IP",
      cell: (info) => <span class="text-muted-foreground">{info.getValue() || "—"}</span>,
    }),
    auditColumnHelper.display({
      id: "actions",
      header: () => <span class="block text-right">操作</span>,
      cell: (info) => (
        <div class="text-right">
          <button
            type="button"
            onClick={() => setDetail(info.row.original)}
            class="font-medium text-primary hover:underline"
          >
            檢視
          </button>
        </div>
      ),
    }),
  ]);

  const query = createQuery(() =>
    auditLogsQueryOptions({ ...filter(), page: pagination().pageIndex + 1, pageSize: pagination().pageSize })
  );

  const total = () => Number(queryData(query, (d) => d?.pagination?.total) ?? 0);

  const table = createTable({
    features: AUDIT_TABLE_FEATURES,
    columns,
    get data() {
      return queryData(query, (d) => d?.items) ?? NO_AUDIT_LOGS;
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

  /**
   * 送出篩選：把草稿套進 query key 並回第 1 頁（同一個 `batch` 內，避免兩次 RPC）。
   *
   * `YYYY-MM-DD` 轉 RFC3339：起日取當地 00:00:00、迄日取當地 23:59:59 ——
   * 「9/1 到 9/22」在使用者時區裡是完整兩天，轉成 UTC 也只是同一個瞬間的另一種寫法。
   */
  const submitFilter: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    batch(() => {
      // 事件處理器內的一次性快照（非反應式訂閱）：送出去的就是按下滑鼠那一刻的條件。
      const staticDraft = filterDraft();
      setFilter({
        from: staticDraft.from ? new Date(`${staticDraft.from}T00:00:00`).toISOString() : "",
        to: staticDraft.to ? new Date(`${staticDraft.to}T23:59:59`).toISOString() : "",
        action: staticDraft.action,
        resourceType: staticDraft.resourceType,
        resourceId: staticDraft.resourceId,
        userId: staticDraft.userId,
      });
      table.setPageIndex(0);
    });
  };

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">稽核日誌</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            共 {total()} 筆・未填時間即查近 3 個月（保留期 D27）
          </p>
        </div>
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

      <Card>
        <form
          class="grid gap-3 border-b border-border bg-muted px-3 py-3 sm:grid-cols-2 lg:grid-cols-4"
          onSubmit={submitFilter}
        >
          <Field>
            <FieldLabel for="audit-from">起日</FieldLabel>
            <Input
              id="audit-from"
              type="date"
              value={filterDraft().from}
              onInput={(e) => setFilterDraft({ ...filterDraft(), from: e.currentTarget.value })}
            />
          </Field>
          <Field>
            <FieldLabel for="audit-to">迄日</FieldLabel>
            <Input
              id="audit-to"
              type="date"
              value={filterDraft().to}
              onInput={(e) => setFilterDraft({ ...filterDraft(), to: e.currentTarget.value })}
            />
          </Field>
          <Field>
            <FieldLabel for="audit-action">動作</FieldLabel>
            <select
              id="audit-action"
              value={filterDraft().action}
              onChange={(e) => setFilterDraft({ ...filterDraft(), action: e.currentTarget.value })}
            >
              <option value="">全部</option>
              <For each={Object.entries(ACTION_LABELS)}>
                {([value, label]) => <option value={value}>{label}</option>}
              </For>
            </select>
          </Field>
          <Field>
            <FieldLabel for="audit-resource-type">資源類型</FieldLabel>
            <Input
              id="audit-resource-type"
              value={filterDraft().resourceType}
              placeholder="例：customer"
              onInput={(e) =>
                setFilterDraft({ ...filterDraft(), resourceType: e.currentTarget.value.trim() })
              }
            />
          </Field>
          <Field>
            <FieldLabel for="audit-resource-id">資源 ID</FieldLabel>
            <Input
              id="audit-resource-id"
              value={filterDraft().resourceId}
              inputMode="numeric"
              onInput={(e) =>
                setFilterDraft({ ...filterDraft(), resourceId: e.currentTarget.value.trim() })
              }
            />
          </Field>
          <Field>
            <FieldLabel for="audit-user-id">操作者 ID</FieldLabel>
            <Input
              id="audit-user-id"
              value={filterDraft().userId}
              inputMode="numeric"
              onInput={(e) => setFilterDraft({ ...filterDraft(), userId: e.currentTarget.value.trim() })}
            />
          </Field>
          <div class="flex items-end gap-3 lg:col-span-2">
            <Button type="submit" variant="outline">
              查詢
            </Button>
            <Button
              type="button"
              variant="ghost"
              onClick={() =>
                setFilterDraft({
                  from: "",
                  to: "",
                  action: "",
                  resourceType: "",
                  resourceId: "",
                  userId: "",
                })
              }
            >
              清空條件
            </Button>
          </div>
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
              when={queryData(query, (d) => d?.items?.length)}
              fallback={
                <TableRow>
                  <TableCell colSpan={columns.length}>
                    {query.isFetching ? "載入中…" : "尚無稽核紀錄"}
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

      <Dialog open={detail() !== null} onOpenChange={(open) => !open && setDetail(null)}>
        <DialogContent class="max-w-3xl">
          <DialogHeader>
            <DialogTitle>稽核明細</DialogTitle>
            <DialogDescription>這一筆紀錄的來源與變更快照（唯讀）</DialogDescription>
          </DialogHeader>
          <Show when={detail()}>
            {(row) => (
              <div class="space-y-3 text-sm">
                <div class="grid gap-2 sm:grid-cols-2">
                  <div>
                    <span class="text-muted-foreground">時間：</span>
                    {row().createdAt ? row().createdAt.slice(0, 19).replace("T", " ") : "—"}
                  </div>
                  <div>
                    <span class="text-muted-foreground">操作者：</span>
                    {row().userName || `#${row().userId}`}
                  </div>
                  <div>
                    <span class="text-muted-foreground">動作：</span>
                    <Badge variant={ACTION_VARIANTS[row().action] ?? "secondary"}>
                      {ACTION_LABELS[row().action] ?? row().action}
                    </Badge>
                  </div>
                  <div>
                    <span class="text-muted-foreground">資源：</span>
                    {row().resourceType ? `${row().resourceType} #${row().resourceId}` : "—"}
                  </div>
                  <div>
                    <span class="text-muted-foreground">IP：</span>
                    {row().ipAddress || "—"}
                  </div>
                  <div class="sm:col-span-2 break-all">
                    <span class="text-muted-foreground">User-Agent：</span>
                    {row().userAgent || "—"}
                  </div>
                </div>
                <div class="grid gap-3 sm:grid-cols-2">
                  <div>
                    <div class="mb-1 font-medium text-foreground">變更前</div>
                    <pre class="max-h-64 overflow-auto rounded-lg bg-muted p-2 text-xs">
                      {row().beforeSnapshot ? prettyJson(row().beforeSnapshot) : "（無）"}
                    </pre>
                  </div>
                  <div>
                    <div class="mb-1 font-medium text-foreground">變更後</div>
                    <pre class="max-h-64 overflow-auto rounded-lg bg-muted p-2 text-xs">
                      {row().afterSnapshot ? prettyJson(row().afterSnapshot) : "（無）"}
                    </pre>
                  </div>
                </div>
              </div>
            )}
          </Show>
          <DialogClose class={buttonVariants({ variant: "outline" })}>關閉</DialogClose>
        </DialogContent>
      </Dialog>
    </main>
  );
}
