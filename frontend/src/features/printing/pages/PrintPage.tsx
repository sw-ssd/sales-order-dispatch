import { Code, ConnectError } from "@connectrpc/connect";
import { createForm } from "@tanstack/solid-form";
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
import type { LogEntry } from "~/lib/proto/products/v1/print_pb";
import {
  Badge,
  Button,
  buttonVariants,
  Card,
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Field,
  FieldError,
  FieldLabel,
  Input,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui";
import { appFormOptions, fieldValidators, firstMessage } from "../../form-helpers";
import { ListPagination } from "../../users/components/ListPagination";
import { PRINT_LOG_PAGE_SIZE, printClient, printLogsQueryOptions } from "../queries";
import { queryData } from "~/lib/query-data";
import {
  DOC_TYPES,
  DOC_TYPE_LABELS,
  printSchema,
  validateScope,
} from "../schemas";

/**
 * 列印紀錄表格的 table 功能集：**只有分頁**。
 *
 * `ListLogsRequest` 沒有 `sort`/`desc`（見 proto）→ 這頁不開排序（同 `OrdersPage`）。
 */
const PRINT_TABLE_FEATURES = tableFeatures({ rowPaginationFeature });

const logColumnHelper = createColumnHelper<typeof PRINT_TABLE_FEATURES, LogEntry>();

/** 沒有資料時的穩定空陣列（避免 row model 反覆重算）。 */
const NO_LOGS: LogEntry[] = [];

/** 列印表單的欄位預設值（`form.reset` 要求整份 values，理由見 CompaniesPage 檔頭）。 */
const EMPTY_PRINT_VALUES = {
  documentType: "",
  routeId: "",
  targetDate: "",
  customerId: "",
  warehouseId: "",
};

function errorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    switch (err.code) {
      case Code.NotFound:
        return "資料不存在或已被刪除";
      case Code.InvalidArgument:
        // 後端「無可列印資料」「非 processing」「重印缺原因」都走這裡，訊息原樣透傳。
        return err.rawMessage || "輸入資料有誤,請檢查後再試";
      case Code.FailedPrecondition:
        return err.rawMessage || "目前狀態不允許此操作";
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

/**
 * 單據列印頁(/printing)。
 *
 * 版型同 `CustomersPage`（Page Headings + 篩選卡片色帶 + 表格 `Card`）。
 * 兩段式操作：**預覽**（任何訂單狀態、不寫 `print_logs`）與**正式列印**
 * （僅 `processing`、寫 `print_logs`）；後端 `hasPrintLog` 以
 * 部門/文件/車次/日期/選用選擇器判首印或重印，故前端**不自行判斷**、
 * 只把回應的 `isReprint` 顯示出來（避免兩處推導漂移）。
 *
 * 三處契約：
 * - **總數在 `total`**（`ListLogsResponse` 沒有 `pagination`；訂單頁同為 `total`）。
 * - **無排序**（`ListLogsRequest` 沒有 `sort`/`desc`）。
 * - **只列 `print_logs`**：spec §151 要求同時查 `print_previews`，但 proto 未定義
 *   `ListPreviews`（Go 端 `PrintPreview` 只有 `Create`）→ 預覽紀錄列表待補 RPC，本頁不做。
 *
 * 下載用 `download_url`（後端產出的是相對 API 路徑 `/api/v1/files/…/download`，
 * 同源 + cookie 認證）→ 直接當 `<a href>` 即可，不另開 API 代理。
 */
export default function PrintPage() {
  const [filter, setFilter] = createSignal({
    documentType: "",
    routeId: "",
    dateFrom: "",
    dateTo: "",
  });
  const [filterDraft, setFilterDraft] = createSignal({
    documentType: "",
    routeId: "",
    dateFrom: "",
    dateTo: "",
  });
  const [pagination, setPagination] = createSignal<PaginationState>({
    pageIndex: 0,
    pageSize: PRINT_LOG_PAGE_SIZE,
  });

  const client = useQueryClient();

  const columns = logColumnHelper.columns([
    logColumnHelper.accessor("documentType", {
      header: "單據類型",
      cell: (info) => (
        <span class="text-foreground">{DOC_TYPE_LABELS[info.getValue()] ?? info.getValue()}</span>
      ),
    }),
    logColumnHelper.accessor("routeId", {
      header: "車次",
      cell: (info) => (
        <span class="text-muted-foreground">{info.getValue() ? `#${info.getValue()}` : "—"}</span>
      ),
    }),
    logColumnHelper.accessor("targetDate", {
      header: "出貨日期",
      cell: (info) => <span class="text-muted-foreground">{info.getValue()}</span>,
    }),
    logColumnHelper.accessor("printedAt", {
      header: "列印時間",
      cell: (info) => (
        <span class="text-muted-foreground">{info.getValue().slice(0, 19).replace("T", " ")}</span>
      ),
    }),
    logColumnHelper.accessor("isReprint", {
      header: "狀態",
      cell: (info) =>
        info.getValue() ? (
          <Badge variant="warning">重印</Badge>
        ) : (
          <Badge variant="success">首印</Badge>
        ),
    }),
    logColumnHelper.accessor("printedBy", {
      header: "列印人",
      cell: (info) => (
        <span class="text-muted-foreground">{info.getValue() ? `#${info.getValue()}` : "—"}</span>
      ),
    }),
    logColumnHelper.display({
      id: "actions",
      header: () => <span class="block text-right">操作</span>,
      cell: (info) => (
        <div class="flex justify-end gap-3 whitespace-nowrap">
          <a
            href={info.row.original.downloadUrl}
            target="_blank"
            rel="noreferrer"
            class="font-medium text-primary hover:underline"
          >
            下載
          </a>
        </div>
      ),
    }),
  ]);

  const query = createQuery(() =>
    printLogsQueryOptions({
      page: pagination().pageIndex + 1,
      pageSize: pagination().pageSize,
      documentType: filter().documentType || undefined,
      routeId: filter().routeId || undefined,
      dateFrom: filter().dateFrom || undefined,
      dateTo: filter().dateTo || undefined,
    })
  );

  const total = () => queryData(query, (d) => d?.total) ?? 0;

  const table = createTable({
    features: PRINT_TABLE_FEATURES,
    columns,
    get data() {
      return queryData(query, (d) => d?.entries) ?? NO_LOGS;
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

  const [printDialogOpen, setPrintDialogOpen] = createSignal(false);
  const [serverError, setServerError] = createSignal<string | undefined>();
  const [actionError, setActionError] = createSignal<string | null>(null);
  /** 最近一次正式列印的結果（`isReprint` 顯示；spec 要求保留可下載連結）。 */
  const [printResult, setPrintResult] = createSignal<{
    downloadUrl: string;
    isReprint: boolean;
    fileId: string;
  } | null>(null);
  /** 重印原因：常駐欄位。是否重印只有後端 `hasPrintLog` 知道（比對部門/文件/車次/日期/
   * 選用選擇器），前端沒有可預先判定的 RPC —— 等收到 `isReprint` 才現身的檢查永遠
   * 來不及（那時已印完），故常駐、原樣送出，由後端強制（spec「未填 MUST 被拒絕」）。 */
  const [reprintReason, setReprintReason] = createSignal("");

  const bannerError = () => (query.error ? errorMessage(query.error) : actionError());

  createEffect(() => {
    if (query.isFetching) setActionError(null);
  });

  const documentTypeValidators = fieldValidators(printSchema.entries.documentType);
  const routeIdValidators = fieldValidators(printSchema.entries.routeId);
  const targetDateValidators = fieldValidators(printSchema.entries.targetDate);

  const submitFilter: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    batch(() => {
      setFilter({ ...filterDraft() });
      table.setPageIndex(0);
    });
  };

  const form = createForm(() => ({
    ...appFormOptions,
    defaultValues: { ...EMPTY_PRINT_VALUES },
    onSubmit: async ({ value }) => {
      const scopeErr = validateScope(value.documentType, value.customerId, value.warehouseId);
      if (scopeErr) {
        setServerError(scopeErr);
        return;
      }
      try {
        const res = await printClient.print({
          documentType: value.documentType,
          routeId: value.routeId,
          targetDate: value.targetDate,
          customerId: value.customerId,
          warehouseId: value.warehouseId,
          reprintReason: reprintReason(),
        });
        setPrintResult({
          downloadUrl: res.downloadUrl,
          isReprint: res.isReprint,
          fileId: res.fileAssetId,
        });
        await client.invalidateQueries({ queryKey: ["printLogs"] });
      } catch (err) {
        setServerError(errorMessage(err));
      }
    },
  }));

  const isSubmitting = form.useSelector((state) => state.isSubmitting);

  /** 開啟列印表單（每次開啟清空重印原因與上次結果）。 */
  const openPrint = () => {
    setServerError(undefined);
    setPrintResult(null);
    setReprintReason("");
    form.reset({ ...EMPTY_PRINT_VALUES });
    setPrintDialogOpen(true);
  };

  const submit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    if (isSubmitting()) return;
    setServerError(undefined);
    void form.handleSubmit();
  };

  /** 預覽：任何狀態可印、不寫 `print_logs`，回 `downloadUrl` 直接開新分頁。 */
  const doPreview = async () => {
    const value = form.state.values;
    const scopeErr = validateScope(value.documentType, value.customerId, value.warehouseId);
    if (scopeErr) {
      setServerError(scopeErr);
      return;
    }
    setServerError(undefined);
    try {
      const res = await printClient.preview(value);
      window.open(res.downloadUrl, "_blank", "noopener,noreferrer");
    } catch (err) {
      setServerError(errorMessage(err));
    }
  };

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">單據列印</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            預覽不限狀態；正式列印限 processing（共 {total()} 筆列印紀錄）
          </p>
        </div>
        <Button type="button" onClick={openPrint}>
          列印
        </Button>
      </header>

      <Show when={bannerError()}>
        {(message) => (
          <p
            class="mb-4 rounded-lg bg-destructive/15 px-3 py-2 text-sm font-medium text-destructive"
            role="alert"
          >
            {message()}
          </p>
        )}
      </Show>

      <Show when={printResult()}>
        {(r) => (
          <p
            class="mb-4 rounded-lg bg-success/15 px-3 py-2 text-sm font-medium text-success"
            role="status"
          >
            {r().isReprint ? "重印完成" : "首印完成"}
            <a
              href={r().downloadUrl}
              target="_blank"
              rel="noreferrer"
              class="ml-2 font-medium underline"
            >
              下載 PDF
            </a>
          </p>
        )}
      </Show>

      <Card>
        <form
          class="flex flex-wrap items-end gap-3 border-b border-border bg-muted px-3 py-3"
          onSubmit={submitFilter}
        >
          <Field class="w-full sm:w-40">
            <FieldLabel for="log-doc-type">單據類型</FieldLabel>
            <select
              id="log-doc-type"
              value={filterDraft().documentType}
              onChange={(e) =>
                setFilterDraft({ ...filterDraft(), documentType: e.currentTarget.value })
              }
            >
              <option value="">全部類型</option>
              <For each={DOC_TYPES}>
                {(d) => <option value={d.code}>{d.label}</option>}
              </For>
            </select>
          </Field>
          <Field class="w-full sm:w-40">
            <FieldLabel for="log-route">車次</FieldLabel>
            <Input
              id="log-route"
              value={filterDraft().routeId}
              onInput={(e) => setFilterDraft({ ...filterDraft(), routeId: e.currentTarget.value })}
              placeholder="車次 ID"
            />
          </Field>
          <Field class="w-full sm:w-40">
            <FieldLabel for="log-from">起日</FieldLabel>
            <input
              id="log-from"
              type="date"
              value={filterDraft().dateFrom}
              onChange={(e) => setFilterDraft({ ...filterDraft(), dateFrom: e.currentTarget.value })}
            />
          </Field>
          <Field class="w-full sm:w-40">
            <FieldLabel for="log-to">迄日</FieldLabel>
            <input
              id="log-to"
              type="date"
              value={filterDraft().dateTo}
              onChange={(e) => setFilterDraft({ ...filterDraft(), dateTo: e.currentTarget.value })}
            />
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
              when={queryData(query, (d) => d?.entries?.length)}
              fallback={
                <TableRow>
                  <TableCell colSpan={7}>
                    {query.isFetching ? "載入中…" : "尚無列印紀錄"}
                  </TableCell>
                </TableRow>
              }
            >
              <For each={table.getRowModel().rows}>
                {(row) => (
                  <TableRow>
                    <For each={row.getAllCells()}>
                      {(cell) => (
                        <TableCell>
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </TableCell>
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

      <Dialog open={printDialogOpen()} onOpenChange={setPrintDialogOpen}>
        <DialogContent class="max-w-3xl">
          <DialogHeader>
            <DialogTitle>列印單據</DialogTitle>
            <DialogDescription>
              預覽不限狀態；正式列印僅限 processing，已列印過則需填重印原因
            </DialogDescription>
          </DialogHeader>

          <Show when={serverError()}>
            {(message) => (
              <p
                class="rounded-lg bg-destructive/15 px-3 py-2 text-sm text-destructive"
                role="alert"
              >
                {message()}
              </p>
            )}
          </Show>

          <form class="space-y-4" onSubmit={submit}>
            <div class="grid gap-4 sm:grid-cols-2">
              <form.Field name="documentType" validators={documentTypeValidators}>
                {(field) => (
                  <Field invalid={!field().state.meta.isValid}>
                    <FieldLabel for="print-doc-type">單據類型 *</FieldLabel>
                    <select
                      id="print-doc-type"
                      value={field().state.value}
                      onBlur={field().handleBlur}
                      onChange={(e) => field().handleChange(e.currentTarget.value)}
                    >
                      <option value="">請選擇</option>
                      <For each={DOC_TYPES}>
                        {(d) => <option value={d.code}>{d.label}</option>}
                      </For>
                    </select>
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="routeId" validators={routeIdValidators}>
                {(field) => (
                  <Field invalid={!field().state.meta.isValid}>
                    <FieldLabel for="print-route">車次 *</FieldLabel>
                    <Input
                      id="print-route"
                      value={field().state.value}
                      onBlur={field().handleBlur}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      placeholder="車次 ID"
                    />
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="targetDate" validators={targetDateValidators}>
                {(field) => (
                  <Field invalid={!field().state.meta.isValid}>
                    <FieldLabel for="print-target-date">出貨日期 *</FieldLabel>
                    <input
                      id="print-target-date"
                      type="date"
                      value={field().state.value}
                      onBlur={field().handleBlur}
                      onChange={(e) => field().handleChange(e.currentTarget.value)}
                    />
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="customerId">
                {(field) => (
                  <Field>
                    <FieldLabel for="print-customer">店家（對點單必填）</FieldLabel>
                    <select
                      id="print-customer"
                      value={field().state.value}
                      onChange={(e) => field().handleChange(e.currentTarget.value)}
                    >
                      <option value="">不指定</option>
                    </select>
                  </Field>
                )}
              </form.Field>

              <form.Field name="warehouseId">
                {(field) => (
                  <Field>
                    <FieldLabel for="print-warehouse">倉別（揀貨單必填）</FieldLabel>
                    <select
                      id="print-warehouse"
                      value={field().state.value}
                      onChange={(e) => field().handleChange(e.currentTarget.value)}
                    >
                      <option value="">不指定</option>
                    </select>
                  </Field>
                )}
              </form.Field>

                <Field>
                  <FieldLabel for="print-reason">重印原因</FieldLabel>
                  <Input
                    id="print-reason"
                    value={reprintReason()}
                    onInput={(e) => setReprintReason(e.currentTarget.value)}
                    placeholder="例：司機未帶單"
                  />
                </Field>

            </div>

            <DialogFooter>
              <DialogClose class={buttonVariants({ variant: "outline" })}>
                取消
              </DialogClose>
              <Button
                type="button"
                variant="outline"
                onClick={() => void doPreview()}
              >
                預覽
              </Button>
              <Button type="submit" loading={isSubmitting()}>
                正式列印
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </main>
  );
}
