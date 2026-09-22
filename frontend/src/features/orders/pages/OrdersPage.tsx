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
import { batch, createEffect, createMemo, createSignal, For, Show, type JSX } from "solid-js";
import type { Customer } from "~/lib/proto/customers/v1/customer_pb";
import type { Product } from "~/lib/proto/products/v1/product_pb";
import type { SalesOrder } from "~/lib/proto/salesorder/v1/salesorder_pb";
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
import { customerDropdownQueryOptions } from "../../customers/queries";
import { productDropdownQueryOptions } from "../../products/queries";
import { ListPagination } from "../../users/components/ListPagination";
import {
  ORDER_PAGE_SIZE,
  orderClient,
  orderEventsQueryOptions,
  orderQueryOptions,
  ordersQueryOptions,
  orderSourceOptionsQueryOptions,
} from "../queries";
import {
  emptyItem,
  orderSchema,
  toOrderItemInput,
  validateOrderItems,
  type OrderItemDraft,
} from "../schemas";

/**
 * 訂單表格的 table 功能集：**只有分頁**。
 *
 * `ListOrdersRequest` 沒有 `sort`/`desc`（見 proto），所以這頁刻意不開排序、欄位一律
 * `enableSorting: false`——開了排序就會送後端不解析的參數（樣板 = `UsersPage`，同理同因）。
 */
const ORDER_TABLE_FEATURES = tableFeatures({ rowPaginationFeature });

const orderColumnHelper = createColumnHelper<typeof ORDER_TABLE_FEATURES, SalesOrder>();

/** 沒有資料時的穩定空陣列：`data` 每次回傳新陣列會讓 table 的 row model 反覆重算。 */
const NO_ORDERS: SalesOrder[] = [];

/** 訂單狀態（D13 狀態機的五個狀態值）。 */
const STATUS_LABELS: Record<string, string> = {
  pending: "待處理",
  processing: "處理中",
  completed: "已完成",
  cancelled: "已取消",
  voided: "已作廢",
};

const STATUS_VARIANTS: Record<
  string,
  "success" | "warning" | "secondary" | "destructive" | "info"
> = {
  pending: "info",
  processing: "warning",
  completed: "success",
  cancelled: "secondary",
  voided: "destructive",
};

/** 事件類型標籤（`SalesOrderEvent.event_type`）。 */
const EVENT_LABELS: Record<string, string> = {
  create: "建立",
  edit: "編輯",
  dispatch: "派車",
  dispatch_cancel: "取消派車",
  cancel: "取消",
  complete: "完成",
  void: "作廢",
};

/** 新增／編輯 modal 的欄位預設值（`form.reset` 要求整份 values，理由見 CompaniesPage 檔頭）。 */
const EMPTY_ORDER_VALUES = {
  customerId: "",
  source: "",
  expectedDeliveryDate: "",
  note: "",
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
        // 狀態機拒絕（非 pending 不能取消、非 processing 不能完成…）走這裡。
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
 * 訂單管理頁(/orders)。
 *
 * 版型同 `CustomersPage`（Page Headings + 篩選卡片色帶 + 表格 `Card`；內距由 AppShell 內容區負責）。
 * 清單資料一律來自 `../orders/queries.ts` 的 `ordersQueryOptions`＋`createQuery`；
 * 頁面只持有「查詢輸入」（篩選草稿、已套用篩選、table 的 pagination state）。
 *
 * 三處契約差異（相較 customers/users 兩頁）：
 * - **總數在 `data.total` 不是 `pagination.total`**（`ListOrdersResponse` 是 `int32 total`）。
 * - **無排序**（`ListOrdersRequest` 沒有 `sort`/`desc`）。
 * - 明細是動態陣列 → 放在 form 之外用 `items` signal 管理，送出前個別驗證
 *   （`validateOrderItems`），錯誤顯示在明細區塊而非欄位級。
 *
 * 動作只掛在列上、詳情框唯讀：讓 Ark 對話框不需在「一個關、另一個開」的同一 tick 交接
 * （那個交接會對新開的對話框回呼 `onOpenChange(false)`，見 CustomersPage `deliveryPayload` 註解）。
 */
export default function OrdersPage() {
  // 篩選草稿：輸入過程只動這些 signal，不進 query key（D5：不得每按一鍵就查詢）。
  const [keywordDraft, setKeywordDraft] = createSignal("");
  const [statusDraft, setStatusDraft] = createSignal("");
  const [customerDraft, setCustomerDraft] = createSignal("");
  const [sourceDraft, setSourceDraft] = createSignal("");
  const [includeDeleted, setIncludeDeleted] = createSignal(false);
  // 已套用的篩選是 query key 的來源：只有送出篩選會動它。
  const [filter, setFilter] = createSignal({
    keyword: "",
    status: "",
    customerId: "",
    source: "",
  });
  // 頁碼與每頁筆數的唯一真相＝table 的 pagination state（受控；後端無排序參數故無 sorting）。
  const [pagination, setPagination] = createSignal<PaginationState>({
    pageIndex: 0,
    pageSize: ORDER_PAGE_SIZE,
  });

  const client = useQueryClient();

  /** 客戶名稱解析：從客戶下拉（本頁篩選也用它）的已載入頁攤平成 id → 名稱。 */
  const customerOptions = createInfiniteQuery(() =>
    customerDropdownQueryOptions({})
  );
  const customerNameById = createMemo(() => {
    const map = new Map<string, string>();
    for (const page of customerOptions.data?.pages ?? []) {
      for (const c of page.customers) map.set(c.id, c.name);
    }
    return map;
  });
  const customerLabel = (id: string) =>
    customerNameById().get(id) ?? (id === "" ? "—" : `#${id}`);

  const columns = orderColumnHelper.columns([
    orderColumnHelper.accessor("orderNo", {
      header: "訂單編號",
      cell: (info) => <span class="font-medium text-foreground">{info.getValue()}</span>,
    }),
    orderColumnHelper.accessor("customerId", {
      header: "客戶",
      cell: (info) => <span class="text-muted-foreground">{customerLabel(info.getValue())}</span>,
    }),
    orderColumnHelper.accessor("source", {
      header: "來源",
      cell: (info) => <span class="text-muted-foreground">{info.getValue() || "—"}</span>,
    }),
    orderColumnHelper.accessor("status", {
      header: "狀態",
      cell: (info) => {
        const s = info.getValue();
        return <Badge variant={STATUS_VARIANTS[s] ?? "secondary"}>{STATUS_LABELS[s] ?? s}</Badge>;
      },
    }),
    orderColumnHelper.accessor("expectedDeliveryDate", {
      header: "出貨日",
      cell: (info) => <span class="text-muted-foreground">{info.getValue() || "—"}</span>,
    }),
    orderColumnHelper.accessor("createdAt", {
      header: "建立時間",
      cell: (info) => (
        <span class="text-muted-foreground">{info.getValue().slice(0, 19).replace("T", " ")}</span>
      ),
    }),
    orderColumnHelper.display({
      id: "actions",
      header: () => <span class="block text-right">操作</span>,
      cell: (info) => {
        const o = info.row.original;
        return (
          <div class="flex justify-end gap-3 whitespace-nowrap">
            <button
              type="button"
              onClick={() => void openDetail(o)}
              class="font-medium text-primary hover:underline"
            >
              詳情
            </button>
            <Show when={o.status === "pending"}>
              <button
                type="button"
                onClick={() => void openEdit(o)}
                class="font-medium text-primary hover:underline"
              >
                編輯
              </button>
              <button
                type="button"
                onClick={() => void transition(o, "cancel")}
                class="font-medium text-destructive hover:underline"
              >
                取消
              </button>
            </Show>
            <Show when={o.status === "processing"}>
              <button
                type="button"
                onClick={() => void transition(o, "complete")}
                class="font-medium text-primary hover:underline"
              >
                完成
              </button>
            </Show>
            <Show when={o.status === "completed"}>
              <button
                type="button"
                onClick={() => openVoid(o)}
                class="font-medium text-destructive hover:underline"
              >
                作廢
              </button>
            </Show>
            <Show when={o.status === "pending" || o.status === "cancelled"}>
              <button
                type="button"
                onClick={() => void remove(o)}
                class="font-medium text-destructive hover:underline"
              >
                刪除
              </button>
            </Show>
          </div>
        );
      },
    }),
  ]);

  const query = createQuery(() =>
    ordersQueryOptions({
      page: pagination().pageIndex + 1,
      pageSize: pagination().pageSize,
      status: filter().status || undefined,
      customerId: filter().customerId || undefined,
      source: filter().source || undefined,
      keyword: filter().keyword || undefined,
      includeDeleted: includeDeleted() || undefined,
    })
  );

  const total = () => query.data?.total ?? 0;

  const table = createTable({
    features: ORDER_TABLE_FEATURES,
    columns,
    get data() {
      return query.data?.orders ?? NO_ORDERS;
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

  // —— 建單／編輯 ——
  const [dialogOpen, setDialogOpen] = createSignal(false);
  const [editing, setEditing] = createSignal<SalesOrder | null>(null);
  const [items, setItems] = createSignal<OrderItemDraft[]>([emptyItem()]);
  const [itemError, setItemError] = createSignal<string | undefined>();
  const [serverError, setServerError] = createSignal<string | undefined>();
  const [actionError, setActionError] = createSignal<string | null>(null);

  // —— 詳情（唯讀） ——
  const [detailId, setDetailId] = createSignal<string | null>(null);

  // —— 作廢（需原因） ——
  const [voidTarget, setVoidTarget] = createSignal<SalesOrder | null>(null);
  const [voidReason, setVoidReason] = createSignal("");

  const sourceOptions = createQuery(() => orderSourceOptionsQueryOptions());
  const productOptions = createInfiniteQuery(() => productDropdownQueryOptions({}));

  /** 商品 id → 商品（明細顯示名與單位選項來源）。 */
  const productById = createMemo(() => {
    const map = new Map<string, Product>();
    for (const page of productOptions.data?.pages ?? []) {
      for (const p of page.products) map.set(p.id, p);
    }
    return map;
  });

  const bannerError = () => (query.error ? errorMessage(query.error) : actionError());

  createEffect(() => {
    if (query.isFetching) setActionError(null);
  });

  const customerIdValidators = fieldValidators(orderSchema.entries.customerId);
  const sourceValidators = fieldValidators(orderSchema.entries.source);

  /** 送出篩選：把草稿套進 query key 並回第 1 頁（同一個 `batch` 內，避免兩次 RPC）。 */
  const submitFilter: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    batch(() => {
      setFilter({
        keyword: keywordDraft(),
        status: statusDraft(),
        customerId: customerDraft(),
        source: sourceDraft(),
      });
      table.setPageIndex(0);
    });
  };

  const form = createForm(() => ({
    ...appFormOptions,
    defaultValues: { ...EMPTY_ORDER_VALUES },
    onSubmit: async ({ value }) => {
      const itemsErr = validateOrderItems(items());
      if (itemsErr) {
        setItemError(itemsErr);
        return;
      }
      const payload = items().map((it) => {
        // 送出當下的商品快照（非反應式）：品名在按下建立的那一刻定案。
        const staticProduct = it.productId === "" ? undefined : productById().get(it.productId);
        return toOrderItemInput(it, staticProduct ? staticProduct.name : it.manualName);
      });
      try {
        const current = editing();
        if (current) {
          await orderClient.updateOrder({
            id: current.id,
            version: current.version,
            expectedDeliveryDate: value.expectedDeliveryDate,
            note: value.note,
            items: payload,
          });
        } else {
          await orderClient.createOrder({
            customerId: value.customerId,
            source: value.source,
            expectedDeliveryDate: value.expectedDeliveryDate,
            note: value.note,
            items: payload,
          });
        }
        setDialogOpen(false);
        await client.invalidateQueries({ queryKey: ["orders"] });
      } catch (err) {
        setServerError(errorMessage(err));
      }
    },
  }));

  const isSubmitting = form.useSelector((state) => state.isSubmitting);

  const openCreate = () => {
    setEditing(null);
    setServerError(undefined);
    setItemError(undefined);
    setItems([emptyItem()]);
    form.reset({ ...EMPTY_ORDER_VALUES });
    setDialogOpen(true);
  };

  /** 編輯：先取該單明細（清單回應不含 items），載入成功才開對話框。 */
  const openEdit = async (o: SalesOrder) => {
    setServerError(undefined);
    setItemError(undefined);
    try {
      const res = await orderClient.getOrder({ id: o.id });
      setEditing(o);
      form.reset({
        customerId: o.customerId,
        source: o.source,
        expectedDeliveryDate: o.expectedDeliveryDate,
        note: o.note,
      });
      setItems(
        (res.items ?? []).map((it) => ({
          productId: it.productId,
          manualName: it.productId === "" ? it.displayName : "",
          qty: it.qty,
          unit: it.unit,
          processingSpecId: it.processingSpecId,
          specialCutNote: it.specialCutNote,
          saveAlias: false,
        }))
      );
      setDialogOpen(true);
    } catch (err) {
      setActionError(errorMessage(err));
    }
  };

  const submit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    if (isSubmitting()) return;
    setServerError(undefined);
    setItemError(undefined);
    void form.handleSubmit();
  };

  const addRow = () => {
    setItems((prev) => [...prev, emptyItem()]);
    setItemError(undefined);
  };
  const removeRow = (index: number) => {
    setItems((prev) => (prev.length === 1 ? [emptyItem()] : prev.filter((_, i) => i !== index)));
  };
  const patchRow = (index: number, patch: Partial<OrderItemDraft>) => {
    setItems((prev) => prev.map((it, i) => (i === index ? { ...it, ...patch } : it)));
    setItemError(undefined);
  };

  /** 狀態轉移（取消／完成）：只帶 id；失敗（狀態不符、跨租戶）顯示在頁面 banner。 */
  const transition = async (o: SalesOrder, kind: "cancel" | "complete") => {
    const label = kind === "cancel" ? "取消" : "完成";
    if (!window.confirm(`確定${label}訂單「${o.orderNo}」?`)) return;
    setActionError(null);
    try {
      if (kind === "cancel") await orderClient.cancelOrder({ id: o.id });
      else await orderClient.completeOrder({ id: o.id });
      await client.invalidateQueries({ queryKey: ["orders"] });
      await client.invalidateQueries({ queryKey: ["order"] });
    } catch (err) {
      setActionError(errorMessage(err));
    }
  };

  const openVoid = (o: SalesOrder) => {
    setVoidTarget(o);
    setVoidReason("");
    setActionError(null);
  };

  const submitVoid = async () => {
    const target = voidTarget();
    if (!target) return;
    const reason = voidReason().trim();
    if (reason === "") {
      setActionError("作廢需填原因");
      return;
    }
    setActionError(null);
    try {
      await orderClient.voidOrder({ id: target.id, reason });
      setVoidTarget(null);
      await client.invalidateQueries({ queryKey: ["orders"] });
      await client.invalidateQueries({ queryKey: ["order"] });
    } catch (err) {
      setActionError(errorMessage(err));
    }
  };

  const remove = async (o: SalesOrder) => {
    if (!window.confirm(`確定刪除訂單「${o.orderNo}」?`)) return;
    setActionError(null);
    try {
      await orderClient.deleteOrder({ id: o.id });
      await client.invalidateQueries({ queryKey: ["orders"] });
    } catch (err) {
      setActionError(errorMessage(err));
    }
  };

  const openDetail = (o: SalesOrder) => setDetailId(o.id);

  const detail = createQuery(() => ({
    ...orderQueryOptions(detailId() ?? ""),
    enabled: detailId() !== null,
  }));
  const events = createQuery(() => ({
    ...orderEventsQueryOptions({ orderId: detailId() ?? "", page: 1, pageSize: 50 }),
    enabled: detailId() !== null,
  }));

  const detailTotal = () => detail.data?.items?.length ?? 0;

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">訂單管理</h1>
          <p class="mt-1 text-sm text-muted-foreground">銷售訂單(共 {total()} 筆)</p>
        </div>
        <Button type="button" onClick={openCreate}>
          新增訂單
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
          <Field class="w-full sm:w-56">
            <FieldLabel for="order-keyword">關鍵字</FieldLabel>
            <Input
              id="order-keyword"
              value={keywordDraft()}
              onInput={(e) => setKeywordDraft(e.currentTarget.value)}
              placeholder="單號 / 備註"
            />
          </Field>
          <Field class="w-full sm:w-36">
            <FieldLabel for="order-status">狀態</FieldLabel>
            <select
              id="order-status"
              value={statusDraft()}
              onChange={(e) => setStatusDraft(e.currentTarget.value)}
            >
              <option value="">全部狀態</option>
              <For each={Object.keys(STATUS_LABELS)}>
                {(s) => <option value={s}>{STATUS_LABELS[s]}</option>}
              </For>
            </select>
          </Field>
          <Field class="w-full sm:w-56">
            <FieldLabel for="order-customer">客戶</FieldLabel>
            <select
              id="order-customer"
              value={customerDraft()}
              onChange={(e) => setCustomerDraft(e.currentTarget.value)}
            >
              <option value="">全部客戶</option>
              <For each={(customerOptions.data?.pages ?? []).flatMap((p) => p.customers)}>
                {(c: Customer) => <option value={c.id}>{c.name}</option>}
              </For>
            </select>
          </Field>
          <Field class="w-full sm:w-40">
            <FieldLabel for="order-source">來源</FieldLabel>
            <select
              id="order-source"
              value={sourceDraft()}
              onChange={(e) => setSourceDraft(e.currentTarget.value)}
            >
              <option value="">全部來源</option>
              <For each={sourceOptions.data?.options ?? []}>
                {(o) => <option value={o.code}>{o.displayName}</option>}
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
              when={query.data?.orders?.length}
              fallback={
                <TableRow>
                  <TableCell colSpan={7}>
                    {query.isFetching ? "載入中…" : "尚無訂單"}
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

      {/* 建單／編輯 */}
      <Dialog open={dialogOpen()} onOpenChange={setDialogOpen}>
        <DialogContent class="max-w-3xl">
          <DialogHeader>
            <DialogTitle>{editing() ? "編輯訂單" : "新增訂單"}</DialogTitle>
            <DialogDescription>
              {editing()
                ? "僅待處理可編輯；明細非空即整單替換"
                : "訂單編號與金額由系統處理（本頁不輸入金額）"}
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
              <form.Field name="customerId" validators={customerIdValidators}>
                {(field) => (
                  <Field invalid={!field().state.meta.isValid}>
                    <FieldLabel for="order-form-customer">客戶 *</FieldLabel>
                    <select
                      id="order-form-customer"
                      value={field().state.value}
                      disabled={editing() !== null}
                      onChange={(e) => field().handleChange(e.currentTarget.value)}
                    >
                      <option value="">請選擇</option>
                      <For each={(customerOptions.data?.pages ?? []).flatMap((p) => p.customers)}>
                        {(c: Customer) => <option value={c.id}>{c.name}</option>}
                      </For>
                    </select>
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="source" validators={sourceValidators}>
                {(field) => (
                  <Field invalid={!field().state.meta.isValid}>
                    <FieldLabel for="order-form-source">訂單來源 *</FieldLabel>
                    <select
                      id="order-form-source"
                      value={field().state.value}
                      disabled={editing() !== null}
                      onChange={(e) => field().handleChange(e.currentTarget.value)}
                    >
                      <option value="">請選擇</option>
                      <For each={sourceOptions.data?.options ?? []}>
                        {(o) => <option value={o.code}>{o.displayName}</option>}
                      </For>
                    </select>
                    <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                  </Field>
                )}
              </form.Field>

              <form.Field name="expectedDeliveryDate">
                {(field) => (
                  <Field>
                    <FieldLabel for="order-form-date">出貨日</FieldLabel>
                    <input
                      id="order-form-date"
                      type="date"
                      value={field().state.value}
                      onChange={(e) => field().handleChange(e.currentTarget.value)}
                    />
                  </Field>
                )}
              </form.Field>

              <form.Field name="note">
                {(field) => (
                  <Field>
                    <FieldLabel for="order-form-note">備註</FieldLabel>
                    <Input
                      id="order-form-note"
                      value={field().state.value}
                      onInput={(e) => field().handleChange(e.currentTarget.value)}
                    />
                  </Field>
                )}
              </form.Field>
            </div>

            {/* 明細（動態陣列，form 之外管理） */}
            <section class="space-y-3 border-t border-border pt-4">
              <div class="flex items-center justify-between">
                <h3 class="text-sm font-semibold text-foreground">明細</h3>
                <Button type="button" variant="outline" size="sm" onClick={addRow}>
                  加一項
                </Button>
              </div>

              <Show when={itemError()}>
                {(message) => (
                  <p class="text-sm text-destructive" role="alert">
                    {message()}
                  </p>
                )}
              </Show>

              <For each={items()}>
                {(row, index) => {
                  const product = () =>
                    row.productId === "" ? undefined : productById().get(row.productId);
                  return (
                    <div class="grid gap-3 rounded-lg border border-border p-3 sm:grid-cols-6">
                      <Field class="sm:col-span-2">
                        <FieldLabel for={`item-product-${index()}`}>商品</FieldLabel>
                        <select
                          id={`item-product-${index()}`}
                          value={row.productId}
                          onChange={(e) =>
                            patchRow(index(), { productId: e.currentTarget.value, saveAlias: false })
                          }
                        >
                          <option value="">（手打品名）</option>
                          <For each={(productOptions.data?.pages ?? []).flatMap((p) => p.products)}>
                            {(p: Product) => <option value={p.id}>{p.name}</option>}
                          </For>
                        </select>
                      </Field>

                      <Show
                        when={row.productId === ""}
                        fallback={
                          <Field class="sm:col-span-2">
                            <FieldLabel>單位</FieldLabel>
                            <select
                              value={row.unit}
                              onChange={(e) => patchRow(index(), { unit: e.currentTarget.value })}
                            >
                              <option value="">請選擇</option>
                              <For each={product()?.units ?? []}>
                                {(u) => <option value={u.unitCode}>{u.unitCode}</option>}
                              </For>
                            </select>
                          </Field>
                        }
                      >
                        <Field class="sm:col-span-2">
                          <FieldLabel for={`item-manual-${index()}`}>品名（手打）</FieldLabel>
                          <Input
                            id={`item-manual-${index()}`}
                            value={row.manualName}
                            onInput={(e) => patchRow(index(), { manualName: e.currentTarget.value })}
                          />
                        </Field>
                        <Field class="sm:col-span-1">
                          <FieldLabel for={`item-unit-${index()}`}>單位</FieldLabel>
                          <Input
                            id={`item-unit-${index()}`}
                            value={row.unit}
                            onInput={(e) => patchRow(index(), { unit: e.currentTarget.value })}
                          />
                        </Field>
                      </Show>

                      <Field class="sm:col-span-1">
                        <FieldLabel for={`item-qty-${index()}`}>數量 *</FieldLabel>
                        <Input
                          id={`item-qty-${index()}`}
                          inputMode="decimal"
                          value={row.qty}
                          onInput={(e) => patchRow(index(), { qty: e.currentTarget.value })}
                        />
                      </Field>

                      <Field class="sm:col-span-2">
                        <FieldLabel for={`item-spec-${index()}`}>分切規格</FieldLabel>
                        <select
                          id={`item-spec-${index()}`}
                          value={row.processingSpecId}
                          onChange={(e) =>
                            patchRow(index(), { processingSpecId: e.currentTarget.value })
                          }
                        >
                          <option value="">無</option>
                          <For each={product()?.processingSpecs ?? []}>
                            {(s) => <option value={s.processingSpecId}>{s.processingSpecId}</option>}
                          </For>
                        </select>
                      </Field>

                      <Field class="sm:col-span-4">
                        <FieldLabel for={`item-cut-${index()}`}>分切備註</FieldLabel>
                        <Input
                          id={`item-cut-${index()}`}
                          value={row.specialCutNote}
                          onInput={(e) => patchRow(index(), { specialCutNote: e.currentTarget.value })}
                        />
                      </Field>

                      <Show when={row.productId !== ""}>
                        <label class="flex items-end gap-2 pb-2 text-sm text-foreground">
                          <input
                            type="checkbox"
                            checked={row.saveAlias}
                            onChange={(e) => patchRow(index(), { saveAlias: e.currentTarget.checked })}
                          />
                          存為客戶別名
                        </label>
                      </Show>

                      <div class="flex items-end justify-end pb-2">
                        <button
                          type="button"
                          onClick={() => removeRow(index())}
                          class="text-sm font-medium text-destructive hover:underline"
                        >
                          移除
                        </button>
                      </div>
                    </div>
                  );
                }}
              </For>
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

      {/* 詳情（唯讀：明細＋事件軌跡） */}
      <Dialog
        open={detailId() !== null}
        onOpenChange={(open) => {
          if (!open) setDetailId(null);
        }}
      >
        <DialogContent class="max-w-3xl">
          <DialogHeader>
            <DialogTitle>訂單詳情</DialogTitle>
            <DialogDescription>{detail.data?.order?.orderNo}</DialogDescription>
          </DialogHeader>

          <Show when={detail.data?.order} keyed>
            {(o) => (
              <div class="space-y-4 text-sm">
                <dl class="grid gap-3 sm:grid-cols-3">
                  <div>
                    <dt class="text-muted-foreground">狀態</dt>
                    <dd>
                      <Badge variant={STATUS_VARIANTS[o.status] ?? "secondary"}>
                        {STATUS_LABELS[o.status] ?? o.status}
                      </Badge>
                    </dd>
                  </div>
                  <div>
                    <dt class="text-muted-foreground">客戶</dt>
                    <dd>{customerLabel(o.customerId)}</dd>
                  </div>
                  <div>
                    <dt class="text-muted-foreground">來源</dt>
                    <dd>{o.source}</dd>
                  </div>
                  <div>
                    <dt class="text-muted-foreground">出貨日</dt>
                    <dd>{o.expectedDeliveryDate || "—"}</dd>
                  </div>
                  <div>
                    <dt class="text-muted-foreground">建立時間</dt>
                    <dd>{o.createdAt.slice(0, 19).replace("T", " ")}</dd>
                  </div>
                  <div>
                    <dt class="text-muted-foreground">備註</dt>
                    <dd>{o.note || "—"}</dd>
                  </div>
                </dl>

                <section>
                  <h3 class="mb-2 font-semibold text-foreground">明細（共 {detailTotal()} 項）</h3>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>品名</TableHead>
                        <TableHead>數量</TableHead>
                        <TableHead>單位</TableHead>
                        <TableHead>基本數量</TableHead>
                        <TableHead>分切備註</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      <For each={detail.data?.items ?? []}>
                        {(it) => (
                          <TableRow>
                            <TableCell>{it.displayName}</TableCell>
                            <TableCell>{it.qty}</TableCell>
                            <TableCell>{it.unit}</TableCell>
                            <TableCell>{it.baseQty}</TableCell>
                            <TableCell class="text-muted-foreground">
                              {it.specialCutNote || "—"}
                            </TableCell>
                          </TableRow>
                        )}
                      </For>
                    </TableBody>
                  </Table>
                </section>

                <section>
                  <h3 class="mb-2 font-semibold text-foreground">事件軌跡</h3>
                  <Show
                    when={(events.data?.events ?? []).length > 0}
                    fallback={<p class="text-muted-foreground">尚無事件</p>}
                  >
                    <ul class="space-y-2">
                      <For each={events.data?.events ?? []}>
                        {(ev) => (
                          <li class="flex flex-wrap gap-2 text-muted-foreground">
                            <span class="text-foreground">
                              {EVENT_LABELS[ev.eventType] ?? ev.eventType}
                            </span>
                            <span>{ev.createdAt.slice(0, 19).replace("T", " ")}</span>
                            <Show when={ev.reason}>
                              <span>原因：{ev.reason}</span>
                            </Show>
                          </li>
                        )}
                      </For>
                    </ul>
                  </Show>
                </section>
              </div>
            )}
          </Show>

          <DialogFooter>
            <DialogClose class={buttonVariants({ variant: "outline" })}>關閉</DialogClose>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 作廢（需原因） */}
      <Dialog
        open={voidTarget() !== null}
        onOpenChange={(open) => {
          if (!open) setVoidTarget(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>作廢訂單</DialogTitle>
            <DialogDescription>{voidTarget()?.orderNo}</DialogDescription>
          </DialogHeader>
          <Field>
            <FieldLabel for="void-reason">作廢原因 *</FieldLabel>
            <Input
              id="void-reason"
              value={voidReason()}
              onInput={(e) => setVoidReason(e.currentTarget.value)}
            />
          </Field>
          <DialogFooter>
            <DialogClose class={buttonVariants({ variant: "outline" })}>取消</DialogClose>
            <Button type="button" onClick={() => void submitVoid()}>
              確認作廢
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </main>
  );
}
