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
import type { Warehouse } from "~/lib/proto/masters/v1/master_pb";
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
import { WAREHOUSES_PAGE_SIZE, warehousesQueryOptions, warehouseClient } from "../queries";
import { warehouseSchema } from "../schemas";
import { queryData } from "~/lib/query-data";

/**
 * 倉別表格的 table 功能集：**只有分頁**。
 *
 * `ListWarehousesRequest` 沒有 `sort`/`desc`（見 proto）→ 這頁刻意不開排序（同 `RoutesPage`
 * ／`ProductsPage`：送了後端也不解析）。
 */
const WAREHOUSE_TABLE_FEATURES = tableFeatures({ rowPaginationFeature });

const warehouseColumnHelper = createColumnHelper<typeof WAREHOUSE_TABLE_FEATURES, Warehouse>();

/** 沒有資料時的穩定空陣列（避免 row model 反覆重算）。 */
const NO_WAREHOUSES: Warehouse[] = [];

/** 新增／編輯 modal 的欄位預設值（`form.reset` 要求整份 values，理由見 CompaniesPage 檔頭）。 */
const EMPTY_WAREHOUSE_VALUES = {
  code: "",
  name: "",
  address: "",
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
 * 倉別主檔頁(/masters/warehouses)。
 *
 * 版型同 `RoutesPage`（Page Headings + 篩選卡片色帶 + 表格 `Card`）。
 * 倉別是商品「進貨倉／揀貨倉」的來源（products 的 `validateWarehouseRef` 對空字串回 nil
 * → 商品可不掛倉）；停用只是一個標記，列仍會顯示（badge「停用」），軟刪除才移出預設清單。
 *
 * 三處契約：
 * - **總數在 `pagination.total`**（`ListWarehousesResponse` 有 pagination；int64 → `Number(...)`）。
 * - **無排序**（`ListWarehousesRequest` 沒有 `sort`/`desc` → 表頭一律非可排序，同 `RoutesPage`）。
 * - **寫入為欄位式 presence**（`UpdateWarehouse` 逐欄判 `!= nil`），故送出整份表單值，
 *   想清空的欄位（地址）送空字串即代表清除。
 */
export default function WarehousesPage() {
  const [keywordDraft, setKeywordDraft] = createSignal("");
  const [includeDeleted, setIncludeDeleted] = createSignal(false);
  const [filter, setFilter] = createSignal({ keyword: "" });
  const [pagination, setPagination] = createSignal<PaginationState>({
    pageIndex: 0,
    pageSize: WAREHOUSES_PAGE_SIZE,
  });

  const client = useQueryClient();

  const columns = warehouseColumnHelper.columns([
    warehouseColumnHelper.accessor("code", {
      header: "倉別代號",
      cell: (info) => <span class="font-medium text-foreground">{info.getValue()}</span>,
    }),
    warehouseColumnHelper.accessor("name", {
      header: "名稱",
      cell: (info) => <span class="text-foreground">{info.getValue()}</span>,
    }),
    warehouseColumnHelper.accessor("address", {
      header: "地址",
      cell: (info) => <span class="text-muted-foreground">{info.getValue() || "—"}</span>,
    }),
    warehouseColumnHelper.accessor("isActive", {
      header: "啟用",
      cell: (info) =>
        info.getValue() ? (
          <Badge variant="success">啟用</Badge>
        ) : (
          <Badge variant="secondary">停用</Badge>
        ),
    }),
    warehouseColumnHelper.display({
      id: "actions",
      header: () => <span class="block text-right">操作</span>,
      cell: (info) => {
        const w = info.row.original;
        return (
          <div class="flex justify-end gap-3 whitespace-nowrap">
            <button
              type="button"
              onClick={() => openEdit(w)}
              class="font-medium text-primary hover:underline"
            >
              編輯
            </button>
            <Show
              when={w.deletedAt}
              fallback={
                <button
                  type="button"
                  onClick={() => void remove(w)}
                  class="font-medium text-destructive hover:underline"
                >
                  刪除
                </button>
              }
            >
              <button
                type="button"
                onClick={() => void restore(w)}
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
    warehousesQueryOptions({
      page: pagination().pageIndex + 1,
      pageSize: pagination().pageSize,
      keyword: filter().keyword || undefined,
      includeDeleted: includeDeleted() || undefined,
    })
  );

  const total = () => Number(queryData(query, (d) => d?.pagination?.total) ?? 0);

  const table = createTable({
    features: WAREHOUSE_TABLE_FEATURES,
    columns,
    get data() {
      return queryData(query, (d) => d?.warehouses) ?? NO_WAREHOUSES;
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
  const [editing, setEditing] = createSignal<Warehouse | null>(null);
  const [serverError, setServerError] = createSignal<string | undefined>();
  const [actionError, setActionError] = createSignal<string | null>(null);

  createEffect(() => {
    if (query.isFetching) setActionError(null);
  });

  const codeValidators = fieldValidators(warehouseSchema.entries.code);
  const nameValidators = fieldValidators(warehouseSchema.entries.name);

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
    defaultValues: { ...EMPTY_WAREHOUSE_VALUES },
    onSubmit: async ({ value }) => {
      try {
        const current = editing();
        if (current) {
          await warehouseClient.updateWarehouse({
            id: current.id,
            code: value.code.trim(),
            name: value.name.trim(),
            address: value.address,
            isActive: value.isActive,
          });
        } else {
          await warehouseClient.createWarehouse({
            code: value.code.trim(),
            name: value.name.trim(),
            address: value.address,
            isActive: value.isActive,
          });
        }
        setDialogOpen(false);
        await client.invalidateQueries({ queryKey: ["warehouses"] });
      } catch (err) {
        setServerError(errorMessage(err));
      }
    },
  }));

  const isSubmitting = form.useSelector((state) => state.isSubmitting);

  const openCreate = () => {
    setEditing(null);
    setServerError(undefined);
    form.reset({ ...EMPTY_WAREHOUSE_VALUES });
    setDialogOpen(true);
  };

  /** 編輯：倉別清單已帶全部欄位，直接以該列值開表單（無需另取單筆 —— `WarehouseService` 也沒有 Get）。 */
  const openEdit = (w: Warehouse) => {
    setEditing(w);
    setServerError(undefined);
    form.reset({
      code: w.code,
      name: w.name,
      address: w.address,
      isActive: w.isActive,
    });
    setDialogOpen(true);
  };

  const submit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    if (isSubmitting()) return;
    setServerError(undefined);
    void form.handleSubmit();
  };

  const remove = async (w: Warehouse) => {
    if (!window.confirm(`確定刪除倉別「${w.name}」？刪除後可用「含已刪除」查回並還原。`)) return;
    setActionError(null);
    try {
      await warehouseClient.deleteWarehouse({ id: w.id });
      await client.invalidateQueries({ queryKey: ["warehouses"] });
    } catch (err) {
      setActionError(errorMessage(err));
    }
  };

  const restore = async (w: Warehouse) => {
    setActionError(null);
    try {
      await warehouseClient.restoreWarehouse({ id: w.id });
      await client.invalidateQueries({ queryKey: ["warehouses"] });
    } catch (err) {
      setActionError(errorMessage(err));
    }
  };

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">倉別主檔</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            倉別即商品的進貨／揀貨倉來源（共 {total()} 筆）
          </p>
        </div>
        <Button type="button" onClick={openCreate}>
          新增倉別
        </Button>
      </header>

      <Show when={query.error ? errorMessage(query.error) : actionError()}>
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
            <FieldLabel for="warehouse-keyword">關鍵字</FieldLabel>
            <Input
              id="warehouse-keyword"
              value={keywordDraft()}
              onInput={(e) => setKeywordDraft(e.currentTarget.value)}
              placeholder="代號 / 名稱"
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
              when={queryData(query, (d) => d?.warehouses?.length)}
              fallback={
                <TableRow>
                  <TableCell colSpan={5}>
                    {query.isFetching ? "載入中…" : "尚無倉別"}
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
            <DialogTitle>{editing() ? "編輯倉別" : "新增倉別"}</DialogTitle>
            <DialogDescription>
              {editing()
                ? "停用不會刪除，資料仍會列出"
                : "建立後即可做為商品的進貨／揀貨倉"}
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
                    <FieldLabel for="warehouse-form-code">倉別代號 *</FieldLabel>
                    <Input
                      id="warehouse-form-code"
                      value={field().state.value}
                      onBlur={field().handleBlur}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      placeholder="WH1"
                    />
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="name" validators={nameValidators}>
                {(field) => (
                  <Field invalid={!field().state.meta.isValid}>
                    <FieldLabel for="warehouse-form-name">名稱 *</FieldLabel>
                    <Input
                      id="warehouse-form-name"
                      value={field().state.value}
                      onBlur={field().handleBlur}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      placeholder="主倉"
                    />
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="address">
                {(field) => (
                  <Field>
                    <FieldLabel for="warehouse-form-address">地址（選填）</FieldLabel>
                    <Input
                      id="warehouse-form-address"
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
                        id="warehouse-form-active"
                        type="checkbox"
                        checked={field().state.value}
                        onChange={(e) => field().handleChange(e.currentTarget.checked)}
                      />
                      啟用
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
