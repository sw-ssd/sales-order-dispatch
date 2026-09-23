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
import type { Route } from "~/lib/proto/masters/v1/master_pb";
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
import { ROUTES_PAGE_SIZE, routesQueryOptions, routeClient } from "../queries";
import { routeSchema, toSortOrder } from "../schemas";
import { queryData } from "~/lib/query-data";

/**
 * 車次表格的 table 功能集：**只有分頁**。
 *
 * `ListRoutesRequest` 沒有 `sort`/`desc`（見 proto）→ 這頁刻意不開排序（同 `OrdersPage`
 * ／`ProductsPage`：送了後端也不解析）。
 */
const ROUTE_TABLE_FEATURES = tableFeatures({ rowPaginationFeature });

const routeColumnHelper = createColumnHelper<typeof ROUTE_TABLE_FEATURES, Route>();

/** 沒有資料時的穩定空陣列（避免 row model 反覆重算）。 */
const NO_ROUTES: Route[] = [];

/** 新增／編輯 modal 的欄位預設值（`form.reset` 要求整份 values，理由見 CompaniesPage 檔頭）。 */
const EMPTY_ROUTE_VALUES = {
  code: "",
  name: "",
  description: "",
  sortOrder: "0",
  isActive: true,
};

function errorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    switch (err.code) {
      case Code.NotFound:
        return "資料不存在或已被刪除";
      case Code.AlreadyExists:
        return err.rawMessage || "資料已存在";
      case Code.InvalidArgument:
        return err.rawMessage || "輸入資料有誤,請檢查後再試";
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
 * 車次主檔頁(/masters/routes)。
 *
 * 版型同 `CustomersPage`（Page Headings + 篩選卡片色帶 + 表格 `Card`）。
 * 車次是派車看板的欄位來源（dispatch spec：以車次為欄）；停用的車次不會出現在看板，
 * 但仍可在此列出並還原 —— 所以停用是「暫不上板」，不是刪除。
 *
 * 三處契約：
 * - **總數在 `pagination.total`**（`ListRoutesResponse` 有 pagination；訂單頁是 `total`）。
 * - **無排序**（`ListRoutesRequest` 沒有 `sort`/`desc`）。
 * - **寫入為欄位式 presence**（`UpdateRoute` 逐欄判 `!= nil`），故送出整份表單值，
 *   想清空的欄位（說明）送空字串即代表清除。
 */
export default function RoutesPage() {
  const [keywordDraft, setKeywordDraft] = createSignal("");
  const [includeDeleted, setIncludeDeleted] = createSignal(false);
  const [filter, setFilter] = createSignal({ keyword: "" });
  const [pagination, setPagination] = createSignal<PaginationState>({
    pageIndex: 0,
    pageSize: ROUTES_PAGE_SIZE,
  });

  const client = useQueryClient();

  const columns = routeColumnHelper.columns([
    routeColumnHelper.accessor("code", {
      header: "車次代號",
      cell: (info) => <span class="font-medium text-foreground">{info.getValue()}</span>,
    }),
    routeColumnHelper.accessor("name", {
      header: "車次名稱",
      cell: (info) => <span class="text-foreground">{info.getValue()}</span>,
    }),
    routeColumnHelper.accessor("description", {
      header: "說明",
      cell: (info) => <span class="text-muted-foreground">{info.getValue() || "—"}</span>,
    }),
    routeColumnHelper.accessor("sortOrder", {
      header: "排序",
      cell: (info) => <span class="text-muted-foreground">{info.getValue()}</span>,
    }),
    routeColumnHelper.accessor("isActive", {
      header: "啟用",
      cell: (info) =>
        info.getValue() ? (
          <Badge variant="success">啟用</Badge>
        ) : (
          <Badge variant="secondary">停用</Badge>
        ),
    }),
    routeColumnHelper.display({
      id: "actions",
      header: () => <span class="block text-right">操作</span>,
      cell: (info) => {
        const r = info.row.original;
        return (
          <div class="flex justify-end gap-3 whitespace-nowrap">
            <button
              type="button"
              onClick={() => openEdit(r)}
              class="font-medium text-primary hover:underline"
            >
              編輯
            </button>
            <Show
              when={r.deletedAt}
              fallback={
                <button
                  type="button"
                  onClick={() => void remove(r)}
                  class="font-medium text-destructive hover:underline"
                >
                  刪除
                </button>
              }
            >
              <button
                type="button"
                onClick={() => void restore(r)}
                class="font-medium text-primary hover:underline"
              >
                還原
              </button>
            </Show>
          </div>
        );
      },
    }),
  ]);

  const query = createQuery(() =>
    routesQueryOptions({
      page: pagination().pageIndex + 1,
      pageSize: pagination().pageSize,
      keyword: filter().keyword || undefined,
      includeDeleted: includeDeleted() || undefined,
    })
  );

  const total = () => Number(queryData(query, (d) => d?.pagination?.total) ?? 0);

  const table = createTable({
    features: ROUTE_TABLE_FEATURES,
    columns,
    get data() {
      return queryData(query, (d) => d?.routes) ?? NO_ROUTES;
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

  const [dialogOpen, setDialogOpen] = createSignal(false);
  const [editing, setEditing] = createSignal<Route | null>(null);
  const [serverError, setServerError] = createSignal<string | undefined>();
  const [actionError, setActionError] = createSignal<string | null>(null);

  const bannerError = () => (query.error ? errorMessage(query.error) : actionError());

  createEffect(() => {
    if (query.isFetching) setActionError(null);
  });

  const codeValidators = fieldValidators(routeSchema.entries.code);
  const nameValidators = fieldValidators(routeSchema.entries.name);

  /** 送出篩選：把草稿套進 query key 並回第 1 頁（同一個 `batch` 內，避免兩次 RPC）。 */
  const submitFilter: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    batch(() => {
      setFilter({ keyword: keywordDraft() });
      table.setPageIndex(0);
    });
  };

  const form = createForm(() => ({
    ...appFormOptions,
    defaultValues: { ...EMPTY_ROUTE_VALUES },
    onSubmit: async ({ value }) => {
      try {
        const current = editing();
        if (current) {
          await routeClient.updateRoute({
            id: current.id,
            code: value.code.trim(),
            name: value.name.trim(),
            description: value.description,
            sortOrder: toSortOrder(value.sortOrder),
            isActive: value.isActive,
          });
        } else {
          await routeClient.createRoute({
            code: value.code.trim(),
            name: value.name.trim(),
            description: value.description,
            sortOrder: toSortOrder(value.sortOrder),
            isActive: value.isActive,
          });
        }
        setDialogOpen(false);
        await client.invalidateQueries({ queryKey: ["routes"] });
      } catch (err) {
        setServerError(errorMessage(err));
      }
    },
  }));

  const isSubmitting = form.useSelector((state) => state.isSubmitting);

  const openCreate = () => {
    setEditing(null);
    setServerError(undefined);
    form.reset({ ...EMPTY_ROUTE_VALUES });
    setDialogOpen(true);
  };

  /** 編輯：車次清單已帶全部欄位，直接以該列值開表單（無需另取單筆 —— `RouteService` 也沒有 Get）。 */
  const openEdit = (r: Route) => {
    setEditing(r);
    setServerError(undefined);
    form.reset({
      code: r.code,
      name: r.name,
      description: r.description,
      sortOrder: String(r.sortOrder),
      isActive: r.isActive,
    });
    setDialogOpen(true);
  };

  const submit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    if (isSubmitting()) return;
    setServerError(undefined);
    void form.handleSubmit();
  };

  const remove = async (r: Route) => {
    if (!window.confirm(`確定刪除車次「${r.name}」？刪除後看板不再顯示此欄。`)) return;
    setActionError(null);
    try {
      await routeClient.deleteRoute({ id: r.id });
      await client.invalidateQueries({ queryKey: ["routes"] });
      // 看板以車次為欄 → 主檔變更要讓看板重查（看板 query 前綴獨立）。
      await client.invalidateQueries({ queryKey: ["boardRoutes"] });
    } catch (err) {
      setActionError(errorMessage(err));
    }
  };

  const restore = async (r: Route) => {
    setActionError(null);
    try {
      await routeClient.restoreRoute({ id: r.id });
      await client.invalidateQueries({ queryKey: ["routes"] });
      await client.invalidateQueries({ queryKey: ["boardRoutes"] });
    } catch (err) {
      setActionError(errorMessage(err));
    }
  };

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">車次主檔</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            車次即派車看板的欄位(共 {total()} 筆)
          </p>
        </div>
        <Button type="button" onClick={openCreate}>
          新增車次
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

      <Card>
        <form
          class="flex flex-wrap items-end gap-3 border-b border-border bg-muted px-3 py-3"
          onSubmit={submitFilter}
        >
          <Field class="w-full sm:w-64">
            <FieldLabel for="route-keyword">關鍵字</FieldLabel>
            <Input
              id="route-keyword"
              value={keywordDraft()}
              onInput={(e) => setKeywordDraft(e.currentTarget.value)}
              placeholder="代號 / 名稱 / 說明"
            />
          </Field>
          <label class="flex h-9 cursor-pointer items-center gap-2 text-sm text-foreground">
            <input
              type="checkbox"
              checked={includeDeleted()}
              onChange={(e) => setIncludeDeleted(e.currentTarget.checked)}
            />
            含已刪除
          </label>
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
              when={queryData(query, (d) => d?.routes?.length)}
              fallback={
                <TableRow>
                  <TableCell colSpan={6}>
                    {query.isFetching ? "載入中…" : "尚無車次"}
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

      <Dialog open={dialogOpen()} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing() ? "編輯車次" : "新增車次"}</DialogTitle>
            <DialogDescription>
              {editing()
                ? "停用不會刪除，只讓看板不再顯示此欄"
                : "建立後即為派車看板的一欄"}
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
              <form.Field name="code" validators={codeValidators}>
                {(field) => (
                  <Field invalid={!field().state.meta.isValid}>
                    <FieldLabel for="route-form-code">車次代號 *</FieldLabel>
                    <Input
                      id="route-form-code"
                      value={field().state.value}
                      onBlur={field().handleBlur}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      placeholder="R1"
                    />
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="name" validators={nameValidators}>
                {(field) => (
                  <Field invalid={!field().state.meta.isValid}>
                    <FieldLabel for="route-form-name">車次名稱 *</FieldLabel>
                    <Input
                      id="route-form-name"
                      value={field().state.value}
                      onBlur={field().handleBlur}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      placeholder="一號路線"
                    />
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="description">
                {(field) => (
                  <Field>
                    <FieldLabel for="route-form-desc">說明</FieldLabel>
                    <Input
                      id="route-form-desc"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                    />
                  </Field>
                )}
              </form.Field>

              <form.Field name="sortOrder">
                {(field) => (
                  <Field>
                    <FieldLabel for="route-form-sort">排序</FieldLabel>
                    <Input
                      id="route-form-sort"
                      inputMode="numeric"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                    />
                  </Field>
                )}
              </form.Field>

              <form.Field name="isActive">
                {(field) => (
                  <Field>
                    <label class="flex items-center gap-2 text-sm text-foreground">
                      <input
                        id="route-form-active"
                        type="checkbox"
                        checked={field().state.value}
                        onChange={(e) => field().handleChange(e.currentTarget.checked)}
                      />
                      啟用（停用後看板不顯示此欄）
                    </label>
                  </Field>
                )}
              </form.Field>
            </div>

            <DialogFooter>
              <DialogClose class={buttonVariants({ variant: "outline" })}>取消</DialogClose>
              <Button type="submit" loading={isSubmitting()}>
                {editing() ? "儲存" : "建立"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </main>
  );
}
