import { Code, ConnectError } from "@connectrpc/connect";
import type { JsonObject } from "@bufbuild/protobuf";
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
import type { ProcessingSpec } from "~/lib/proto/masters/v1/master_pb";
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
import {
  PROCESSING_SPECS_PAGE_SIZE,
  processingSpecClient,
  processingSpecsQueryOptions,
} from "../queries";
import { processingSpecSchema, toSortOrder } from "../schemas";
import { queryData } from "~/lib/query-data";

/**
 * 規格表格的 table 功能集：**只有分頁**。
 *
 * `ListProcessingSpecsRequest` 沒有 `sort`/`desc`（見 proto）→ 這頁刻意不開排序；
 * 列表順序由後端 `sort_order, id` 決定（`specListSource`），使用者在「排序」欄位調節。
 */
const PROCESSING_SPEC_TABLE_FEATURES = tableFeatures({ rowPaginationFeature });

const processingSpecColumnHelper = createColumnHelper<
  typeof PROCESSING_SPEC_TABLE_FEATURES,
  ProcessingSpec
>();

/** 沒有資料時的穩定空陣列（避免 row model 反覆重算）。 */
const NO_PROCESSING_SPECS: ProcessingSpec[] = [];

/** 新增／編輯 modal 的欄位預設值（`form.reset` 要求整份 values，理由見 CompaniesPage 檔頭）。 */
const EMPTY_PROCESSING_SPEC_VALUES = {
  code: "",
  name: "",
  kind: "",
  // 兩個 appliesTo 預設皆勾：後端 `validateSpecFlags` 要求至少其一為 true，
  // 預設兩真可讓開表單就送出（且規格同時服務加工室與配送兩條揀貨路徑是合法態）。
  appliesToProcessing: true,
  appliesToPicking: true,
  sortOrder: "0",
  isActive: true,
  attributes: "",
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
 * 分切規格頁(/masters/processing-specs)。
 *
 * 版型同 `RoutesPage`（Page Headings + 篩選卡片色帶 + 表格 `Card`）。
 * 規格泛化自 cutting_specs（proto 3.4.3）：商品勾選規格，決定套用於加工室揀／配送揀
 * （兩者可同時 true）；`attributes` 是自由結構指令 —— 後端原樣儲存不解析。
 *
 * 四處契約：
 * - **總數在 `pagination.total`**（`ListProcessingSpecsResponse` 有 pagination；
 *   int64 → `Number(...)`）。
 * - **無排序**（`ListProcessingSpecsRequest` 沒有 `sort`/`desc` → 表頭一律非可排序）。
 * - **寫入為欄位式 presence，`attributes` 也是 presence**（Go 側 `req.Msg.Attributes != nil`
 *   才套用、且 `master_pb.d.ts` 欄位型別是 `JsonObject`）→ 一律送整份表單解析出的物件：
 *   空白欄位送 `{}`（空 Struct，＝清除屬性），非空送 `JSON.parse` 結果；若「留空就不送」
 *   會靜默保留舊值，與其他欄位「整份送出」的語意不一致。
 * - `kind` 是開放集（metadicts `processing_kind` 背書）→ 自由輸入，後端空值歸 `other`。
 */
export default function ProcessingSpecsPage() {
  const [keywordDraft, setKeywordDraft] = createSignal("");
  const [includeDeleted, setIncludeDeleted] = createSignal(false);
  const [filter, setFilter] = createSignal({ keyword: "" });
  const [pagination, setPagination] = createSignal<PaginationState>({
    pageIndex: 0,
    pageSize: PROCESSING_SPECS_PAGE_SIZE,
  });

  const client = useQueryClient();

  const columns = processingSpecColumnHelper.columns([
    processingSpecColumnHelper.accessor("code", {
      header: "規格代號",
      cell: (info) => <span class="font-medium text-foreground">{info.getValue()}</span>,
    }),
    processingSpecColumnHelper.accessor("name", {
      header: "名稱",
      cell: (info) => <span class="text-foreground">{info.getValue()}</span>,
    }),
    processingSpecColumnHelper.accessor("kind", {
      header: "類型",
      cell: (info) => <span class="text-muted-foreground">{info.getValue() || "—"}</span>,
    }),
    processingSpecColumnHelper.accessor("appliesToProcessing", {
      header: "加工室揀",
      cell: (info) => (
        <span class="text-muted-foreground">{info.getValue() ? "是" : "否"}</span>
      ),
    }),
    processingSpecColumnHelper.accessor("appliesToPicking", {
      header: "配送揀",
      cell: (info) => (
        <span class="text-muted-foreground">{info.getValue() ? "是" : "否"}</span>
      ),
    }),
    processingSpecColumnHelper.accessor("sortOrder", {
      header: "排序",
      cell: (info) => <span class="text-muted-foreground">{info.getValue()}</span>,
    }),
    processingSpecColumnHelper.accessor("isActive", {
      header: "啟用",
      cell: (info) =>
        info.getValue() ? (
          <Badge variant="success">啟用</Badge>
        ) : (
          <Badge variant="secondary">停用</Badge>
        ),
    }),
    processingSpecColumnHelper.display({
      id: "actions",
      header: () => <span class="block text-right">操作</span>,
      cell: (info) => {
        const s = info.row.original;
        return (
          <div class="flex justify-end gap-3 whitespace-nowrap">
            <button
              type="button"
              onClick={() => openEdit(s)}
              class="font-medium text-primary hover:underline"
            >
              編輯
            </button>
            <Show
              when={s.deletedAt}
              fallback={
                <button
                  type="button"
                  onClick={() => void remove(s)}
                  class="font-medium text-destructive hover:underline"
                >
                  刪除
                </button>
              }
            >
              <button
                type="button"
                onClick={() => void restore(s)}
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
    processingSpecsQueryOptions({
      page: pagination().pageIndex + 1,
      pageSize: pagination().pageSize,
      keyword: filter().keyword || undefined,
      includeDeleted: includeDeleted() || undefined,
    })
  );

  const total = () => Number(queryData(query, (d) => d?.pagination?.total) ?? 0);

  const table = createTable({
    features: PROCESSING_SPEC_TABLE_FEATURES,
    columns,
    get data() {
      return queryData(query, (d) => d?.processingSpecs) ?? NO_PROCESSING_SPECS;
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
  const [editing, setEditing] = createSignal<ProcessingSpec | null>(null);
  const [serverError, setServerError] = createSignal<string | undefined>();
  const [actionError, setActionError] = createSignal<string | null>(null);

  createEffect(() => {
    if (query.isFetching) setActionError(null);
  });

  const codeValidators = fieldValidators(processingSpecSchema.entries.code);
  const nameValidators = fieldValidators(processingSpecSchema.entries.name);
  const attributesValidators = fieldValidators(processingSpecSchema.entries.attributes);

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
    defaultValues: { ...EMPTY_PROCESSING_SPEC_VALUES },
    onSubmit: async ({ value }) => {
      try {
        const current = editing();
        // 表單字串 → 送出用的 Struct 物件：空白＝空 Struct（清除），其餘必已過欄位層 JSON 驗證。
        const attributes: JsonObject =
          value.attributes.trim() === "" ? {} : (JSON.parse(value.attributes) as JsonObject);
        if (current) {
          await processingSpecClient.updateProcessingSpec({
            id: current.id,
            code: value.code.trim(),
            name: value.name.trim(),
            kind: value.kind.trim(),
            appliesToProcessing: value.appliesToProcessing,
            appliesToPicking: value.appliesToPicking,
            attributes,
            sortOrder: toSortOrder(value.sortOrder),
            isActive: value.isActive,
          });
        } else {
          await processingSpecClient.createProcessingSpec({
            code: value.code.trim(),
            name: value.name.trim(),
            kind: value.kind.trim(),
            appliesToProcessing: value.appliesToProcessing,
            appliesToPicking: value.appliesToPicking,
            attributes,
            sortOrder: toSortOrder(value.sortOrder),
            isActive: value.isActive,
          });
        }
        setDialogOpen(false);
        await client.invalidateQueries({ queryKey: ["processingSpecs"] });
      } catch (err) {
        setServerError(errorMessage(err));
      }
    },
  }));

  const isSubmitting = form.useSelector((state) => state.isSubmitting);

  const openCreate = () => {
    setEditing(null);
    setServerError(undefined);
    form.reset({ ...EMPTY_PROCESSING_SPEC_VALUES });
    setDialogOpen(true);
  };

  /** 編輯：規格清單已帶全部欄位（含 attributes Struct），直接以該列值開表單（無需另取單筆）。 */
  const openEdit = (s: ProcessingSpec) => {
    setEditing(s);
    setServerError(undefined);
    form.reset({
      code: s.code,
      name: s.name,
      kind: s.kind,
      appliesToProcessing: s.appliesToProcessing,
      appliesToPicking: s.appliesToPicking,
      sortOrder: String(s.sortOrder),
      isActive: s.isActive,
      // Struct → 表單字串：list 回的 JsonObject 直接 JSON.stringify（空值歸 {} 與送出側對稱）。
      attributes: JSON.stringify(s.attributes ?? {}),
    });
    setDialogOpen(true);
  };

  const submit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    if (isSubmitting()) return;
    setServerError(undefined);
    void form.handleSubmit();
  };

  const remove = async (s: ProcessingSpec) => {
    if (!window.confirm(`確定刪除規格「${s.name}」？刪除後可用「含已刪除」查回並還原。`)) return;
    setActionError(null);
    try {
      await processingSpecClient.deleteProcessingSpec({ id: s.id });
      await client.invalidateQueries({ queryKey: ["processingSpecs"] });
    } catch (err) {
      setActionError(errorMessage(err));
    }
  };

  const restore = async (s: ProcessingSpec) => {
    setActionError(null);
    try {
      await processingSpecClient.restoreProcessingSpec({ id: s.id });
      await client.invalidateQueries({ queryKey: ["processingSpecs"] });
    } catch (err) {
      setActionError(errorMessage(err));
    }
  };

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">分切規格</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            分切規格即商品可勾選的加工／揀貨條件（共 {total()} 筆）
          </p>
        </div>
        <Button type="button" onClick={openCreate}>
          新增規格
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
            <FieldLabel for="spec-keyword">關鍵字</FieldLabel>
            <Input
              id="spec-keyword"
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
              when={queryData(query, (d) => d?.processingSpecs?.length)}
              fallback={
                <TableRow>
                  <TableCell colSpan={8}>
                    {query.isFetching ? "載入中…" : "尚無規格"}
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
            <DialogTitle>{editing() ? "編輯規格" : "新增規格"}</DialogTitle>
            <DialogDescription>
              {editing()
                ? "停用不會刪除，資料仍會列出"
                : "建立後可在商品主檔勾選此規格"}
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
                    <FieldLabel for="spec-form-code">規格代號 *</FieldLabel>
                    <Input
                      id="spec-form-code"
                      value={field().state.value}
                      onBlur={field().handleBlur}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      placeholder="S1"
                    />
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="name" validators={nameValidators}>
                {(field) => (
                  <Field invalid={!field().state.meta.isValid}>
                    <FieldLabel for="spec-form-name">名稱 *</FieldLabel>
                    <Input
                      id="spec-form-name"
                      value={field().state.value}
                      onBlur={field().handleBlur}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      placeholder="五分切"
                    />
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="kind">
                {(field) => (
                  <Field>
                    <FieldLabel for="spec-form-kind">類型</FieldLabel>
                    <Input
                      id="spec-form-kind"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      placeholder="如 cutting（開放集，字典 processing_kind 背書）"
                    />
                  </Field>
                )}
              </form.Field>

              <form.Field name="sortOrder">
                {(field) => (
                  <Field>
                    <FieldLabel for="spec-form-sort">排序</FieldLabel>
                    <Input
                      id="spec-form-sort"
                      inputMode="numeric"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                    />
                  </Field>
                )}
              </form.Field>

              <form.Field name="appliesToProcessing">
                {(field) => (
                  <Field>
                    <label class="flex items-center gap-2 text-sm text-foreground">
                      <input
                        id="spec-form-proc"
                        type="checkbox"
                        checked={field().state.value}
                        onChange={(e) => field().handleChange(e.currentTarget.checked)}
                      />
                      加工室揀
                    </label>
                  </Field>
                )}
              </form.Field>

              <form.Field name="appliesToPicking">
                {(field) => (
                  <Field>
                    <label class="flex items-center gap-2 text-sm text-foreground">
                      <input
                        id="spec-form-pick"
                        type="checkbox"
                        checked={field().state.value}
                        onChange={(e) => field().handleChange(e.currentTarget.checked)}
                      />
                      配送揀
                    </label>
                  </Field>
                )}
              </form.Field>

              <form.Field name="attributes" validators={attributesValidators}>
                {(field) => (
                  <Field invalid={!field().state.meta.isValid} class="sm:col-span-2">
                    <FieldLabel for="spec-form-attrs">屬性 JSON（選填，留空＝不帶）</FieldLabel>
                    <textarea
                      id="spec-form-attrs"
                      rows={4}
                      class="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground"
                      value={field().state.value}
                      onBlur={field().handleBlur}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                      placeholder='{"depth": 3}'
                    />
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="isActive">
                {(field) => (
                  <Field>
                    <label class="flex items-center gap-2 text-sm text-foreground">
                      <input
                        id="spec-form-active"
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
