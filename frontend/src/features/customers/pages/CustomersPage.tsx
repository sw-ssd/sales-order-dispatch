import { Code, ConnectError } from "@connectrpc/connect";
import { createForm } from "@tanstack/solid-form";
import { createQuery, useQueryClient } from "@tanstack/solid-query";
import {
  createColumnHelper,
  createTable,
  flexRender,
  rowPaginationFeature,
  rowSortingFeature,
  tableFeatures,
  type PaginationState,
  type SortingState,
} from "@tanstack/solid-table";
import { batch, createEffect, createSignal, For, Show, type JSX } from "solid-js";
import type { Customer } from "~/lib/proto/customers/v1/customer_pb";
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
import { queryData } from "~/lib/query-data";
import { appFormOptions, fieldValidators, firstMessage } from "../../form-helpers";
import { ListPagination } from "../../users/components/ListPagination";
import { createSortableHeaders } from "../../users/components/SortableHeader";
import { PAGE_SIZE } from "../../users/queries";
import { customerClient, customersQueryOptions } from "../queries";
import { customerSchema } from "../schemas";
import AddressBookDialog from "../components/AddressBookDialog";
import CustomerProductsDialog from "../components/CustomerProductsDialog";

/**
 * 客戶表格的 table 功能集：分頁 ＋ 排序（`manualSorting`，見下方 table）。
 * features 必須是穩定的靜態值——每個元件都自己 `tableFeatures({...})` 會多一份無用的定義。
 */
const CUSTOMER_TABLE_FEATURES = tableFeatures({ rowPaginationFeature, rowSortingFeature });

/** 欄位定義工具：features 已綁定，`accessor` 的值型別因此跟著功能集推導。 */
const customerColumnHelper = createColumnHelper<typeof CUSTOMER_TABLE_FEATURES, Customer>();

/** 沒有資料時的穩定空陣列：`data` 每次回傳新陣列會讓 table 的 row model 反覆重算。 */
const NO_CUSTOMERS: Customer[] = [];

/** 新增／編輯 modal 的欄位預設值（`form.reset` 要求整份 values，理由見 CompaniesPage 檔頭）。 */
const EMPTY_CUSTOMER_VALUES = { name: "", taxId: "" };

function errorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    switch (err.code) {
      case Code.NotFound:
        return "資料不存在或已被刪除";
      case Code.AlreadyExists:
        return err.rawMessage || "客戶已存在";
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

/**
 * 建檔連動帳號交付的警示區（D22/§9.4）：
 * 臨時密碼**只在 Create 的那一次回應出現**，系統不留明文——一旦關掉就再也查不到，
 * 所以它必須以「待複製的對話框」呈現，且不能只是個 dismiss 掉就沒了的 banner。
 */
/**
 * 建檔連動交付的帳號資料（D22）：僅 `CreateCustomer` 的**那一次**回應帶回，
 * 後端不留明文臨時密碼 —— 所以 UI 必須在成功當下完整顯示，關掉就再也取不到。
 */
interface AccountDelivery {
  primaryAccountName: string;
  primaryTempPassword: string;
  salesRepAccountName: string;
  salesRepTempPassword: string;
  accountManageUrl: string;
}

/**
 * 客戶主檔頁(/customers)。
 *
 * 版型同 `CompaniesPage`（Page Headings + 篩選卡片色帶 + 表格 `Card`；內距由 AppShell 內容區負責）。
 * 清單資料一律來自 `../customers/queries.ts` 的 `customersQueryOptions`＋`createQuery`；
 * 頁面只持有「查詢輸入」（關鍵字草稿、已套用關鍵字、table 的 pagination/sorting state）。
 *
 * 差異於 users 三頁：客戶有 **軟刪除**（D10），所以操作是 刪除／還原，且有「含已刪除」檢視切換；
 * 建檔會回一份**僅此一次**的帳號交付資料（見 `AccountDelivery`）。
 */
export default function CustomersPage() {
  // 篩選草稿：輸入過程只動這顆 signal，不進 query key。
  const [keywordDraft, setKeywordDraft] = createSignal("");
  // 已套用的篩選是 query key 的來源：只有送出篩選會動它。
  const [filter, setFilter] = createSignal("");
  const [includeDeleted, setIncludeDeleted] = createSignal(false);
  // 頁碼與每頁筆數的唯一真相＝table 的 pagination state（受控）。
  const [pagination, setPagination] = createSignal<PaginationState>({
    pageIndex: 0,
    pageSize: PAGE_SIZE,
  });
  // 排序狀態的唯一真相＝table 的 sorting state（受控）：空陣列＝未排序＝後端預設排序。
  const [sorting, setSorting] = createSignal<SortingState>([]);

  const client = useQueryClient();
  const sortableHeader = createSortableHeaders();

  const columns = customerColumnHelper.columns([
    customerColumnHelper.accessor("name", {
      header: (ctx) => sortableHeader(ctx.column, "客戶名稱"),
      cell: (info) => (
        <span class="font-medium text-foreground">
          {info.getValue()}
          <Show when={info.row.original.deletedAt}>
            <Badge variant="secondary" class="ml-2">
              已刪除
            </Badge>
          </Show>
        </span>
      ),
    }),
    customerColumnHelper.accessor("customerCode", {
      header: (ctx) => sortableHeader(ctx.column, "客戶編號"),
      cell: (info) => <span class="text-muted-foreground">{info.getValue()}</span>,
    }),
    customerColumnHelper.accessor("taxId", {
      enableSorting: false,
      header: (ctx) => sortableHeader(ctx.column, "統一編號"),
      cell: (info) => <span class="text-muted-foreground">{info.getValue() || "—"}</span>,
    }),
    customerColumnHelper.accessor("createdAt", {
      header: (ctx) => sortableHeader(ctx.column, "建立時間"),
      cell: (info) => <span class="text-muted-foreground">{info.getValue() || "—"}</span>,
    }),
    customerColumnHelper.display({
      id: "actions",
      header: () => <span class="block text-right">操作</span>,
      cell: (info) => (
        <div class="text-right">
          <button
            type="button"
            onClick={() => void openDialog(info.row.original)}
            class="font-medium text-primary hover:underline"
          >
            編輯
          </button>
          <Show when={!info.row.original.deletedAt}>
            {/* 地址簿/聯絡人是客戶的從屬資料：已刪除的客戶不給入口（後端 requireCustomer
                也只認未刪除的客戶，放了只會是必定 404 的按鈕）。 */}
            <button
              type="button"
              onClick={() => setBook(info.row.original)}
              class="ml-3 font-medium text-primary hover:underline"
            >
              地址簿
            </button>
            <button
              type="button"
              onClick={() => setProductsDialog(info.row.original)}
              class="ml-3 font-medium text-primary hover:underline"
            >
              專屬商品
            </button>
          </Show>
          <Show
            when={info.row.original.deletedAt}
            fallback={
              <button
                type="button"
                onClick={() => void remove(info.row.original)}
                class="ml-3 font-medium text-destructive hover:underline"
              >
                刪除
              </button>
            }
          >
            <button
              type="button"
              onClick={() => void restore(info.row.original)}
              class="ml-3 font-medium text-primary hover:underline"
            >
              還原
            </button>
          </Show>
        </div>
      ),
    }),
  ]);

  const query = createQuery(() =>
    customersQueryOptions({
      page: pagination().pageIndex + 1,
      pageSize: pagination().pageSize,
      sort: sorting()[0]?.id ?? "",
      desc: sorting()[0]?.desc ?? false,
      keyword: filter() || undefined,
      includeDeleted: includeDeleted() || undefined,
    })
  );

  const total = () => Number(queryData(query, (d) => d?.pagination?.total) ?? 0);

  /**
   * table 實例。分頁與排序一律手動（`manualPagination`/`manualSorting`）：
   * 資料永遠只有當前頁且已由後端排好序，table 不得再切一次或再排一次。
   */
  const table = createTable({
    features: CUSTOMER_TABLE_FEATURES,
    columns,
    get data() {
      return queryData(query, (d) => d?.customers ?? NO_CUSTOMERS);
    },
    get rowCount() {
      return total();
    },
    manualPagination: true,
    manualSorting: true,
    // 白名單欄位一律從「升冪」起算（v9 的第一方向預設依資料推測，空資料時會變降冪）。
    sortDescFirst: false,
    // 本專案是**單欄排序**契約（只有一欄的 id/desc 會進請求）→ 關掉多欄事件。
    enableMultiSort: false,
    get state() {
      return { pagination: pagination(), sorting: sorting() };
    },
    onPaginationChange: setPagination,
    /**
     * 排序變更必須與「回第 1 頁」同批（D1）：分開寫會先以「舊頁碼＋新排序」查一次、
     * 再以「第 1 頁＋新排序」查一次（兩次 RPC）。
     */
    onSortingChange: (updater) =>
      batch(() => {
        setSorting((prev) => (typeof updater === "function" ? updater(prev) : updater));
        table.setPageIndex(0);
      }),
  });

  /**
   * 超頁退回：回傳的 total 讓目前頁碼超界時（例：該頁資料被刪光），把頁碼夾到合法值。
   * `page` 是 query key 的一部分 → `setPageIndex` 自己就會觸發重取，不必再手動重載。
   * `!query.data` 那半邊防的是「首次載入／錯誤時 total 算成 0 → 靜默改寫頁碼」。
   */
  createEffect(() => {
    if (query.isPlaceholderData || !query.data) return;
    const maxPage = Math.max(1, Math.ceil(total() / pagination().pageSize));
    if (pagination().pageIndex + 1 > maxPage) table.setPageIndex(maxPage - 1);
  });

  const [dialogOpen, setDialogOpen] = createSignal(false);
  const [editing, setEditing] = createSignal<Customer | null>(null);
  const [serverError, setServerError] = createSignal<string | undefined>();
  const [deleteError, setDeleteError] = createSignal<string | null>(null);
  // 建檔帳號交付（僅此一次；見 AccountDelivery 註解）。
  const [delivery, setDelivery] = createSignal<AccountDelivery | null>(null);
  // 地址簿/聯絡人對話框的目標客戶（null＝關閉）。
  const [book, setBook] = createSignal<Customer | null>(null);
  // 專屬商品對話框的目標客戶（null＝關閉；同一個客戶頁一次只開一個對話框）。
  const [productsDialog, setProductsDialog] = createSignal<Customer | null>(null);
  /**
   * 建立回應後、交付對話框開啟前的中轉（非渲染狀態）。
   *
   * Ark 對「同一 tick 內 A 關、B 開」會對新開的 B 回呼一次 `onOpenChange(false)`
   * （群組焦點交接），而交付對話框的 `onOpenChange(false)` 負責清 `delivery`
   * —— 直接同步 `setDelivery(payload)` 會被那一次回呼立刻清掉、內容永遠不顯示。
   * 所以先關建立框、再於 microtask 開交付框，讓兩者落在不同 tick。
   */
  let deliveryPayload: AccountDelivery | null = null;

  const bannerError = () =>
    query.error ? errorMessage(query.error) : deleteError();

  createEffect(() => {
    if (query.isFetching) setDeleteError(null);
  });

  const nameValidators = fieldValidators(customerSchema.entries.name);
  const taxIdValidators = fieldValidators(customerSchema.entries.taxId);

  /** 送出篩選：把草稿套進 query key 並回第 1 頁（同一個 `batch` 內，避免兩次 RPC）。 */
  const submitFilter: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    batch(() => {
      setFilter(keywordDraft());
      table.setPageIndex(0);
    });
  };

  const form = createForm(() => ({
    ...appFormOptions,
    defaultValues: { ...EMPTY_CUSTOMER_VALUES },
    onSubmit: async ({ value }) => {
      const current = editing();
      const name = value.name.trim();
      try {
        if (current) {
          await customerClient.updateCustomer({
            id: current.id,
            name,
            taxId: value.taxId,
          });
        } else {
          const res = await customerClient.createCustomer({
            name,
            taxId: value.taxId,
          });
          // 帳號交付資料只在建立時出現一次（見 AccountDelivery）。
          deliveryPayload = {
            primaryAccountName: res.primaryAccountName,
            primaryTempPassword: res.primaryTempPassword,
            salesRepAccountName: res.salesRepAccountName,
            salesRepTempPassword: res.salesRepTempPassword,
            accountManageUrl: res.accountManageUrl,
          };
        }
        setDialogOpen(false);
        if (deliveryPayload) {
          const payload = deliveryPayload;
          deliveryPayload = null;
          // 先關建立框、下個 tick 再開交付框（理由見 deliveryPayload 註解）。
          queueMicrotask(() => setDelivery(payload));
        }
        await client.invalidateQueries({ queryKey: ["customers"] });
      } catch (err) {
        setServerError(errorMessage(err));
      }
    },
  }));

  const isSubmitting = form.useSelector((state) => state.isSubmitting);

  /**
   * 開啟 modal（新增傳 `null`）。`form.reset(values)` 整份取代（理由見 CompaniesPage 檔頭），
   * `serverError` 是元件層 signal、不受 `reset()` 影響，這裡顯式清掉。
   */
  const openDialog = (customer: Customer | null) => {
    setEditing(customer);
    setServerError(undefined);
    form.reset(
      customer
        ? { name: customer.name, taxId: customer.taxId }
        : { ...EMPTY_CUSTOMER_VALUES }
    );
    setDialogOpen(true);
  };

  const submit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    if (isSubmitting()) return;
    setServerError(undefined);
    void form.handleSubmit();
  };

  const remove = async (c: Customer) => {
    if (!window.confirm(`確定刪除客戶「${c.name}」?`)) return;
    setDeleteError(null);
    try {
      await customerClient.deleteCustomer({ id: c.id });
      await client.invalidateQueries({ queryKey: ["customers"] });
    } catch (err) {
      setDeleteError(errorMessage(err));
    }
  };

  const restore = async (c: Customer) => {
    setDeleteError(null);
    try {
      await customerClient.restoreCustomer({ id: c.id });
      await client.invalidateQueries({ queryKey: ["customers"] });
    } catch (err) {
      setDeleteError(errorMessage(err));
    }
  };

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">客戶管理</h1>
          <p class="mt-1 text-sm text-muted-foreground">客戶主檔(共 {total()} 筆)</p>
        </div>
        <Button type="button" onClick={() => openDialog(null)}>
          新增客戶
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
            <FieldLabel for="customer-keyword">關鍵字</FieldLabel>
            <Input
              id="customer-keyword"
              value={keywordDraft()}
              onInput={(e) => setKeywordDraft(e.currentTarget.value)}
              placeholder="名稱 / 客戶編號 / 統編"
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
              when={queryData(query, (d) => d?.customers?.length)}
              fallback={
                <TableRow>
                  <TableCell colSpan={5}>
                    {query.isFetching ? "載入中…" : "尚無客戶"}
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
            <DialogTitle>{editing() ? "編輯客戶" : "新增客戶"}</DialogTitle>
            <DialogDescription>
              {editing() ? "修改名稱或統一編號" : "建立客戶主檔（系統自動取號並交付帳號）"}
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
          <form onSubmit={submit}>
            <form.Field name="name" validators={nameValidators}>
              {(field) => (
                <Field invalid={!field().state.meta.isValid}>
                  <FieldLabel for="customer-name">客戶名稱 *</FieldLabel>
                  <Input
                    id="customer-name"
                    required
                    value={field().state.value}
                    onBlur={field().handleBlur}
                    onInput={(e) => field().handleChange(e.currentTarget.value)}
                    placeholder="○○食品行"
                  />
                  <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                </Field>
              )}
            </form.Field>
            <form.Field name="taxId" validators={taxIdValidators}>
              {(field) => (
                <Field invalid={!field().state.meta.isValid}>
                  <FieldLabel for="customer-tax-id">統一編號（選填）</FieldLabel>
                  <Input
                    id="customer-tax-id"
                    value={field().state.value}
                    onBlur={field().handleBlur}
                    onInput={(e) => field().handleChange(e.currentTarget.value)}
                    placeholder="12345678"
                  />
                  <FieldError>{firstMessage(field().state.meta.errors)}</FieldError>
                </Field>
              )}
            </form.Field>
            <DialogFooter>
              <DialogClose class={buttonVariants({ variant: "outline" })}>
                取消
              </DialogClose>
              <Button type="submit" loading={isSubmitting()}>
                儲存
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* 帳號交付對話框（僅此一次；關掉即無法再查，所以刻意不提供「稍後再看」）。 */}
      <Dialog
        open={delivery() !== null}
        onOpenChange={(open) => {
          if (!open) setDelivery(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>帳號已建立</DialogTitle>
            <DialogDescription>
              臨時密碼僅本次顯示（24 小時內首登強制改密），請立即交付並妥存。
            </DialogDescription>
          </DialogHeader>
          <Show when={delivery()}>
            {(d) => (
              <div class="space-y-3 text-sm">
                <div class="rounded-lg bg-muted px-3 py-2">
                  <div class="font-medium">主帳號</div>
                  <div>帳號：{d().primaryAccountName}</div>
                  <div>臨時密碼：{d().primaryTempPassword}</div>
                </div>
                <div class="rounded-lg bg-muted px-3 py-2">
                  <div class="font-medium">業務子帳號</div>
                  <div>帳號：{d().salesRepAccountName}</div>
                  <div>臨時密碼：{d().salesRepTempPassword}</div>
                </div>
                <div class="break-all text-muted-foreground">
                  帳號管理：{d().accountManageUrl}
                </div>
              </div>
            )}
          </Show>
          <DialogFooter>
            <DialogClose class={buttonVariants({ variant: "outline" })}>
              我已複製
            </DialogClose>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 地址簿／聯絡人（04 3.2.1／3.2.2）：從列上開啟，與其他對話框不同時開
          （同 OrdersPage「動作只掛在列上、詳情框唯讀」的慣例，避免對話框交接）。 */}
      <AddressBookDialog
        customer={book()}
        open={book() !== null}
        onOpenChange={(open) => {
          if (!open) setBook(null);
        }}
      />

      {/* 專屬商品（04 Task 3.5）：同地址簿，從列上開、不與其他對話框同開。 */}
      <CustomerProductsDialog
        customer={productsDialog()}
        open={productsDialog() !== null}
        onOpenChange={(open) => {
          if (!open) setProductsDialog(null);
        }}
      />
    </main>
  );
}
