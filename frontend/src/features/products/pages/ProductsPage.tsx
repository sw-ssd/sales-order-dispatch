import { Code, ConnectError } from "@connectrpc/connect";
import { createForm } from "@tanstack/solid-form";
import {
  createInfiniteQuery,
  createQuery,
  useQueryClient,
} from "@tanstack/solid-query";
import {
  createColumnHelper,
  createTable,
  flexRender,
  rowPaginationFeature,
  tableFeatures,
  type PaginationState,
} from "@tanstack/solid-table";
import { batch, createEffect, createMemo, createSignal, For, Index, Show, type JSX } from "solid-js";
import type { ProductCategory } from "~/lib/proto/masters/v1/master_pb";
import type { Product } from "~/lib/proto/products/v1/product_pb";
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
import { queryData } from "~/lib/query-data";
import {
  PRODUCT_PAGE_SIZE,
  categoryDropdownQueryOptions,
  productClient,
  productsQueryOptions,
  unitOptionsQueryOptions,
  warehouseDropdownQueryOptions,
} from "../queries";
import {
  baseUnitDraft,
  derivedUnitDraft,
  productSchema,
  toProductUnitInput,
  validateProductUnits,
  type ProductUnitDraft,
} from "../schemas";

/**
 * 商品表格的 table 功能集：**只有分頁**。
 *
 * `ListProductsRequest` 沒有 `sort`/`desc`（見 proto）→ 這頁刻意不開排序、欄位一律
 * 非可排序表頭（同 `UsersPage`／`OrdersPage`：送了後端也不解析）。
 */
const PRODUCT_TABLE_FEATURES = tableFeatures({ rowPaginationFeature });

const productColumnHelper = createColumnHelper<typeof PRODUCT_TABLE_FEATURES, Product>();

/** 沒有資料時的穩定空陣列（避免 row model 反覆重算）。 */
const NO_PRODUCTS: Product[] = [];

/** 新增／編輯 modal 的欄位預設值（`form.reset` 要求整份 values，理由見 CompaniesPage 檔頭）。 */
const EMPTY_PRODUCT_VALUES = {
  code: "",
  name: "",
  description: "",
  isActive: true,
  categoryId: "",
  inventoryWarehouseId: "",
  pickingWarehouseId: "",
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
      case Code.FailedPrecondition:
        // 配額上限（PLAT-5001）與狀態衝突都走這裡，後端訊息已帶用量。
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
 * 商品主檔頁(/products)。
 *
 * 版型同 `CustomersPage`（Page Headings + 篩選卡片色帶 + 表格 `Card`）。
 * 清單資料一律來自 `../queries.ts` 的 `productsQueryOptions`＋`createQuery`；
 * 頁面只持有「查詢輸入」（篩選草稿、已套用篩選、table 的 pagination state）。
 *
 * 三處契約（相較其他頁）：
 * - **總數在 `pagination.total`**（`ListProductsResponse` 有 pagination；訂單頁是 `total`）。
 * - **無排序**（`ListProductsRequest` 沒有 `sort`/`desc`）。
 * - 單位是動態陣列 → 放在 form 之外用 `units` signal 管理，送出前個別驗證
 *   （`validateProductUnits`），錯誤顯示在單位區塊而非欄位級。
 *
 * **刻意未做**：`processing_specs`（分切規格關聯）的編輯。該陣列在更新時是「提供即整組替換」，
 * 而規格列帶 `attributes`（配對層覆寫指令，任意 JSON）——本頁若以空值送出會覆掉既有屬性。
 * 建立時送空陣列＝不關聯、更新時省略該欄位＝維持現值（後端 `req.Msg.ProcessingSpecs != nil` 守衛），
 * 因此不編輯它不會造成資料損失，等規格頁有編輯器時再補。
 */
export default function ProductsPage() {
  // 篩選草稿：輸入過程只動這些 signal，不進 query key（D5：不得每按一鍵就查詢）。
  const [keywordDraft, setKeywordDraft] = createSignal("");
  const [categoryDraft, setCategoryDraft] = createSignal("");
  const [includeDeleted, setIncludeDeleted] = createSignal(false);
  // 已套用的篩選是 query key 的來源。
  const [filter, setFilter] = createSignal({ keyword: "", categoryId: "" });
  // 頁碼與每頁筆數的唯一真相＝table 的 pagination state。
  const [pagination, setPagination] = createSignal<PaginationState>({
    pageIndex: 0,
    pageSize: PRODUCT_PAGE_SIZE,
  });

  const client = useQueryClient();

  const categories = createInfiniteQuery(() => categoryDropdownQueryOptions());
  const warehouses = createInfiniteQuery(() => warehouseDropdownQueryOptions());
  const unitOptions = createQuery(() => unitOptionsQueryOptions());

  /** 分類 id → 名稱（清單顯示與表單選項共用同一份資料）。 */
  const categoryNameById = createMemo(() => {
    const map = new Map<string, string>();
    for (const page of queryData(categories, (d) => d?.pages) ?? []) {
      for (const c of page.productCategories) map.set(c.id, c.name);
    }
    return map;
  });
  const categoryLabel = (id: string) =>
    categoryNameById().get(id) ?? (id === "" ? "—" : `#${id}`);

  const columns = productColumnHelper.columns([
    productColumnHelper.accessor("code", {
      header: "商品代號",
      cell: (info) => <span class="font-medium text-foreground">{info.getValue()}</span>,
    }),
    productColumnHelper.accessor("name", {
      header: "商品名稱",
      cell: (info) => <span class="text-foreground">{info.getValue()}</span>,
    }),
    productColumnHelper.accessor("categoryId", {
      header: "分類",
      cell: (info) => <span class="text-muted-foreground">{categoryLabel(info.getValue())}</span>,
    }),
    productColumnHelper.accessor("description", {
      header: "說明",
      cell: (info) => (
        <span class="text-muted-foreground">{info.getValue() || "—"}</span>
      ),
    }),
    productColumnHelper.accessor("isActive", {
      header: "啟用",
      cell: (info) =>
        info.getValue() ? (
          <Badge variant="success">啟用</Badge>
        ) : (
          <Badge variant="secondary">停用</Badge>
        ),
    }),
    productColumnHelper.accessor("createdAt", {
      header: "建立時間",
      cell: (info) => (
        <span class="text-muted-foreground">{info.getValue().slice(0, 19).replace("T", " ")}</span>
      ),
    }),
    productColumnHelper.display({
      id: "actions",
      header: () => <span class="block text-right">操作</span>,
      cell: (info) => {
        const p = info.row.original;
        return (
          <div class="flex justify-end gap-3 whitespace-nowrap">
            <button
              type="button"
              onClick={() => void openEdit(p)}
              class="font-medium text-primary hover:underline"
            >
              編輯
            </button>
            <Show
              when={p.deletedAt}
              fallback={
                <button
                  type="button"
                  onClick={() => void remove(p)}
                  class="font-medium text-destructive hover:underline"
                >
                  刪除
                </button>
              }
            >
              <button
                type="button"
                onClick={() => void restore(p)}
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
    productsQueryOptions({
      page: pagination().pageIndex + 1,
      pageSize: pagination().pageSize,
      keyword: filter().keyword || undefined,
      categoryId: filter().categoryId || undefined,
      includeDeleted: includeDeleted() || undefined,
    })
  );

  const total = () => Number(queryData(query, (d) => d?.pagination?.total) ?? 0);

  const table = createTable({
    features: PRODUCT_TABLE_FEATURES,
    columns,
    get data() {
      return queryData(query, (d) => d?.products) ?? NO_PRODUCTS;
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

  // —— 建立／編輯 ——
  const [dialogOpen, setDialogOpen] = createSignal(false);
  const [editing, setEditing] = createSignal<Product | null>(null);
  const [units, setUnits] = createSignal<ProductUnitDraft[]>([baseUnitDraft()]);
  const [unitError, setUnitError] = createSignal<string | undefined>();
  const [serverError, setServerError] = createSignal<string | undefined>();
  const [actionError, setActionError] = createSignal<string | null>(null);

  const bannerError = () => (query.error ? errorMessage(query.error) : actionError());

  createEffect(() => {
    if (query.isFetching) setActionError(null);
  });

  const codeValidators = fieldValidators(productSchema.entries.code);
  const nameValidators = fieldValidators(productSchema.entries.name);

  /** 送出篩選：把草稿套進 query key 並回第 1 頁（同一個 `batch` 內，避免兩次 RPC）。 */
  const submitFilter: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    batch(() => {
      setFilter({ keyword: keywordDraft(), categoryId: categoryDraft() });
      table.setPageIndex(0);
    });
  };

  const form = createForm(() => ({
    ...appFormOptions,
    defaultValues: { ...EMPTY_PRODUCT_VALUES },
    onSubmit: async ({ value }) => {
      const unitsErr = validateProductUnits(units());
      if (unitsErr) {
        setUnitError(unitsErr);
        return;
      }
      const payload = units().map(toProductUnitInput);
      try {
        const current = editing();
        if (current) {
          await productClient.updateProduct({
            id: current.id,
            code: value.code.trim(),
            name: value.name.trim(),
            description: value.description,
            isActive: value.isActive,
            categoryId: value.categoryId,
            inventoryWarehouseId: value.inventoryWarehouseId,
            pickingWarehouseId: value.pickingWarehouseId,
            units: payload,
          });
        } else {
          await productClient.createProduct({
            code: value.code.trim(),
            name: value.name.trim(),
            description: value.description,
            isActive: value.isActive,
            categoryId: value.categoryId,
            inventoryWarehouseId: value.inventoryWarehouseId,
            pickingWarehouseId: value.pickingWarehouseId,
            units: payload,
          });
        }
        setDialogOpen(false);
        await client.invalidateQueries({ queryKey: ["products"] });
      } catch (err) {
        setServerError(errorMessage(err));
      }
    },
  }));

  const isSubmitting = form.useSelector((state) => state.isSubmitting);

  const openCreate = () => {
    setEditing(null);
    setServerError(undefined);
    setUnitError(undefined);
    setUnits([baseUnitDraft()]);
    form.reset({ ...EMPTY_PRODUCT_VALUES });
    setDialogOpen(true);
  };

  /** 編輯：先取該商品（清單回應已含 units，但分類/倉別 ref 與說明以單筆為準），載入成功才開對話框。 */
  const openEdit = async (p: Product) => {
    setServerError(undefined);
    setUnitError(undefined);
    try {
      const res = await productClient.getProduct({ id: p.id });
      const fresh = res.product ?? p;
      setEditing(fresh);
      form.reset({
        code: fresh.code,
        name: fresh.name,
        description: fresh.description,
        isActive: fresh.isActive,
        categoryId: fresh.categoryId,
        inventoryWarehouseId: fresh.inventoryWarehouseId,
        pickingWarehouseId: fresh.pickingWarehouseId,
      });
      // 永保至少一列：單位是必填且「恰一個基本單位」，空編輯器會讓使用者無從修正。
      const loaded = (fresh.units ?? []).map((u, i) => ({
        unitCode: u.unitCode,
        conversionRate: u.conversionRate,
        isBase: u.isBase,
        sortOrder: u.sortOrder || i,
        sizeDesc: u.sizeDesc,
      }));
      setUnits(loaded.length > 0 ? loaded : [baseUnitDraft()]);
      setDialogOpen(true);
    } catch (err) {
      setActionError(errorMessage(err));
    }
  };

  const submit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    if (isSubmitting()) return;
    setServerError(undefined);
    setUnitError(undefined);
    void form.handleSubmit();
  };

  const addUnit = () => {
    setUnits((prev) => [...prev, derivedUnitDraft(prev.length)]);
    setUnitError(undefined);
  };
  const removeUnit = (index: number) => {
    // 只留一列時改為重設（不能移除到 0 列，後端要求恰一個基本單位）。
    setUnits((prev) => (prev.length === 1 ? [baseUnitDraft()] : prev.filter((_, i) => i !== index)));
    setUnitError(undefined);
  };
  const patchUnit = (index: number, patch: Partial<ProductUnitDraft>) => {
    setUnits((prev) => prev.map((u, i) => (i === index ? { ...u, ...patch } : u)));
    setUnitError(undefined);
  };
  /** 勾選為基本單位時，其換算率恆為 1（後端硬性要求），其他列降為換算單位。 */
  const markBase = (index: number) => {
    setUnits((prev) =>
      prev.map((u, i) =>
        i === index
          ? { ...u, isBase: true, conversionRate: "1" }
          : { ...u, isBase: false }
      )
    );
    setUnitError(undefined);
  };

  const remove = async (p: Product) => {
    if (!window.confirm(`確定刪除商品「${p.name}」?`)) return;
    setActionError(null);
    try {
      await productClient.deleteProduct({ id: p.id });
      await client.invalidateQueries({ queryKey: ["products"] });
    } catch (err) {
      setActionError(errorMessage(err));
    }
  };

  const restore = async (p: Product) => {
    setActionError(null);
    try {
      await productClient.restoreProduct({ id: p.id });
      await client.invalidateQueries({ queryKey: ["products"] });
    } catch (err) {
      setActionError(errorMessage(err));
    }
  };

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">商品總表</h1>
          <p class="mt-1 text-sm text-muted-foreground">商品主檔(共 {total()} 筆)</p>
        </div>
        <Button type="button" onClick={openCreate}>
          新增商品
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
            <FieldLabel for="product-keyword">關鍵字</FieldLabel>
            <Input
              id="product-keyword"
              value={keywordDraft()}
              onInput={(e) => setKeywordDraft(e.currentTarget.value)}
              placeholder="代號 / 名稱"
            />
          </Field>
          <Field class="w-full sm:w-56">
            <FieldLabel for="product-category-filter">分類</FieldLabel>
            <select
              id="product-category-filter"
              value={categoryDraft()}
              onChange={(e) => setCategoryDraft(e.currentTarget.value)}
            >
              <option value="">全部分類</option>
              <For each={(queryData(categories, (d) => d?.pages) ?? []).flatMap((p) => p.productCategories)}>
                {(c: ProductCategory) => <option value={c.id}>{c.name}</option>}
              </For>
            </select>
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
              when={queryData(query, (d) => d?.products?.length)}
              fallback={
                <TableRow>
                  <TableCell colSpan={7}>
                    {query.isFetching ? "載入中…" : "尚無商品"}
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
        <DialogContent class="max-w-3xl">
          <DialogHeader>
            <DialogTitle>{editing() ? "編輯商品" : "新增商品"}</DialogTitle>
            <DialogDescription>
              {editing()
                ? "單位為整組替換：送出即以表單內容取代既有單位"
                : "商品代號與名稱必填；單位至少一個且恰一個基本單位"}
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
                    <FieldLabel for="product-form-code">商品代號 *</FieldLabel>
                    <Input
                      id="product-form-code"
                      value={field().state.value}
                      onBlur={field().handleBlur}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                    />
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="name" validators={nameValidators}>
                {(field) => (
                  <Field invalid={!field().state.meta.isValid}>
                    <FieldLabel for="product-form-name">商品名稱 *</FieldLabel>
                    <Input
                      id="product-form-name"
                      value={field().state.value}
                      onBlur={field().handleBlur}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                    />
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="categoryId">
                {(field) => (
                  <Field>
                    <FieldLabel for="product-form-category">所屬分類</FieldLabel>
                    <select
                      id="product-form-category"
                      value={field().state.value}
                      onChange={(e) => field().handleChange(e.currentTarget.value)}
                    >
                      <option value="">無</option>
                      <For each={(queryData(categories, (d) => d?.pages) ?? []).flatMap((p) => p.productCategories)}>
                        {(c: ProductCategory) => <option value={c.id}>{c.name}</option>}
                      </For>
                    </select>
                  </Field>
                )}
              </form.Field>

              <form.Field name="description">
                {(field) => (
                  <Field>
                    <FieldLabel for="product-form-desc">說明</FieldLabel>
                    <Input
                      id="product-form-desc"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                    />
                  </Field>
                )}
              </form.Field>

              <form.Field name="inventoryWarehouseId">
                {(field) => (
                  <Field>
                    <FieldLabel for="product-form-inv-wh">進貨倉</FieldLabel>
                    <select
                      id="product-form-inv-wh"
                      value={field().state.value}
                      onChange={(e) => field().handleChange(e.currentTarget.value)}
                    >
                      <option value="">無</option>
                      <For each={(queryData(warehouses, (d) => d?.pages) ?? []).flatMap((p) => p.warehouses)}>
                        {(w) => <option value={w.id}>{w.name}</option>}
                      </For>
                    </select>
                  </Field>
                )}
              </form.Field>

              <form.Field name="pickingWarehouseId">
                {(field) => (
                  <Field>
                    <FieldLabel for="product-form-pick-wh">揀貨倉</FieldLabel>
                    <select
                      id="product-form-pick-wh"
                      value={field().state.value}
                      onChange={(e) => field().handleChange(e.currentTarget.value)}
                    >
                      <option value="">無</option>
                      <For each={(queryData(warehouses, (d) => d?.pages) ?? []).flatMap((p) => p.warehouses)}>
                        {(w) => <option value={w.id}>{w.name}</option>}
                      </For>
                    </select>
                  </Field>
                )}
              </form.Field>

              <form.Field name="isActive">
                {(field) => (
                  <Field>
                    <label class="flex items-center gap-2 text-sm text-foreground">
                      <input
                        id="product-form-active"
                        type="checkbox"
                        checked={field().state.value}
                        onChange={(e) => field().handleChange(e.currentTarget.checked)}
                      />
                      啟用（停用後建單不可選）
                    </label>
                  </Field>
                )}
              </form.Field>
            </div>

            {/* 單位（動態陣列，form 之外管理） */}
            <section class="space-y-3 border-t border-border pt-4">
              <div class="flex items-center justify-between">
                <h3 class="text-sm font-semibold text-foreground">單位換算</h3>
                <Button type="button" variant="outline" size="sm" onClick={addUnit}>
                  加單位
                </Button>
              </div>

              <Show when={unitError()}>
                {(message) => (
                  <p class="text-sm text-destructive" role="alert">
                    {message()}
                  </p>
                )}
              </Show>

              <Index each={units()}>
                {(row, index) => (
                  <div class="grid gap-3 rounded-lg border border-border p-3 sm:grid-cols-6">
                    <Field class="sm:col-span-2">
                      <FieldLabel for={`unit-code-${index}`}>單位代碼 *</FieldLabel>
                      <select
                        id={`unit-code-${index}`}
                        value={row().unitCode}
                        onChange={(e) => patchUnit(index, { unitCode: e.currentTarget.value })}
                      >
                        <option value="">請選擇</option>
                        <For each={queryData(unitOptions, (d) => d?.options) ?? []}>
                          {(o) => <option value={o.code}>{o.displayName}</option>}
                        </For>
                      </select>
                    </Field>

                    <Field class="sm:col-span-1">
                      <FieldLabel for={`unit-rate-${index}`}>換算率 *</FieldLabel>
                      <Input
                        id={`unit-rate-${index}`}
                        inputMode="decimal"
                        value={row().conversionRate}
                        disabled={row().isBase}
                        onInput={(e) => patchUnit(index, { conversionRate: e.currentTarget.value })}
                      />
                    </Field>

                    <Field class="sm:col-span-1">
                      <FieldLabel for={`unit-sort-${index}`}>排序</FieldLabel>
                      <Input
                        id={`unit-sort-${index}`}
                        inputMode="numeric"
                        value={String(row().sortOrder)}
                        onInput={(e) =>
                          patchUnit(index, { sortOrder: Number(e.currentTarget.value) || 0 })
                        }
                      />
                    </Field>

                    <Field class="sm:col-span-2">
                      <FieldLabel for={`unit-size-${index}`}>規格說明</FieldLabel>
                      <Input
                        id={`unit-size-${index}`}
                        value={row().sizeDesc}
                        placeholder="如「1盒=3kg」"
                        onInput={(e) => patchUnit(index, { sizeDesc: e.currentTarget.value })}
                      />
                    </Field>

                    <div class="flex items-center gap-4 sm:col-span-5">
                      <label class="flex items-center gap-2 text-sm text-foreground">
                        <input
                          type="radio"
                          name="base-unit"
                          checked={row().isBase}
                          onChange={() => markBase(index)}
                        />
                        基本單位（換算率恆為 1）
                      </label>
                    </div>

                    <div class="flex items-center justify-end sm:col-span-1">
                      <button
                        type="button"
                        onClick={() => removeUnit(index)}
                        class="text-sm font-medium text-destructive hover:underline"
                      >
                        移除
                      </button>
                    </div>
                  </div>
                )}
              </Index>
            </section>

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
