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
import type {
  ReturnRequestEntry,
  ReturnRequestItemView,
} from "~/lib/proto/salesorder/v1/returns_pb";
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
  customerNameQueryOptions,
  returnCertificateQueryOptions,
  returnClient,
  returnRequestQueryOptions,
  returnRequestsQueryOptions,
  RETURN_PAGE_SIZE,
} from "../queries";

/**
 * 退貨表格的 table 功能集：**只有分頁**。
 *
 * `ListReturnRequestsRequest` 沒有 `sort`/`desc`（見 proto）→ 這頁不開排序（同 `PrintPage`）。
 */
const RETURN_TABLE_FEATURES = tableFeatures({ rowPaginationFeature });

const returnColumnHelper = createColumnHelper<typeof RETURN_TABLE_FEATURES, ReturnRequestEntry>();

/** 沒有資料時的穩定空陣列：`data` 每次回傳新陣列會讓 table 的 row model 反覆重算。 */
const NO_RETURNS: ReturnRequestEntry[] = [];

/** 退貨狀態標籤（後端僅此三個值）。 */
const STATUS_LABELS: Record<string, string> = {
  pending: "待審核",
  approved: "已核准",
  rejected: "已駁回",
};

const STATUS_VARIANTS: Record<
  string,
  "success" | "warning" | "secondary" | "destructive" | "info"
> = {
  pending: "warning",
  approved: "success",
  rejected: "destructive",
};

/** 明細來源標籤：快照要說得出這一行是「訂單來的」還是「客戶自己的商品」。 */
const SOURCE_LABELS: Record<string, string> = {
  order_item: "來自訂單",
  customer_product: "專屬商品",
};

/**
 * 錯誤訊息對照；樣板 = `CustomersPage`/`PrintPage` 的同一份 switch。
 * `InvalidArgument` 必須原樣透傳 `rawMessage`：版本衝突（「資料已變更，請重新載入」）
 * 與「僅待審核可審」都是後端給的中文說明，改寫會蓋掉唯一有辨識度的訊息。
 */
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

/** 日期時間顯示：RFC3339 的 `T` 換成空白、秒後切除（同 `PrintPage` 的列印時間）。 */
function formatDateTime(value: string): string {
  return value ? value.slice(0, 19).replace("T", " ") : "—";
}

/**
 * 退貨管理頁(/returns)。
 *
 * 版型同 `CustomersPage`/`PrintPage`（Page Headings + 篩選卡片色帶 + 表格 `Card`）。
 * 本頁**不提供建單**：`CreateReturnRequest` 只接受客戶子帳號身份（D25），
 * 員工與主帳號一律被後端拒絕 → 提供按鈕只會是必定失敗的陷阱，改以說明文字告知來源。
 *
 * 審核走樂觀鎖：`expectedVersion` 只能取自 `getReturnRequest` 剛讀到的 `version`
 * （寫死 `"0"` 會讓每次審核都撞鎖）。核准前以 `window.confirm` 確認（同 `CustomersPage`
 * 刪除的破壞性操作慣例）；駁回則要求原因（後端 `reject_reason` 必填）。
 */
export default function ReturnsPage() {
  // 篩選草稿：改下拉只動這顆 signal，不進 query key。
  const [statusDraft, setStatusDraft] = createSignal("");
  // 已套用的篩選是 query key 的來源：只有送出篩選會動它。
  const [statusFilter, setStatusFilter] = createSignal("");
  // 頁碼與每頁筆數的唯一真相＝table 的 pagination state（受控）。
  const [pagination, setPagination] = createSignal<PaginationState>({
    pageIndex: 0,
    pageSize: RETURN_PAGE_SIZE,
  });

  const client = useQueryClient();

  const columns = returnColumnHelper.columns([
    returnColumnHelper.accessor("status", {
      header: "狀態",
      cell: (info) => (
        <Badge variant={STATUS_VARIANTS[info.getValue()] ?? "secondary"}>
          {STATUS_LABELS[info.getValue()] ?? info.getValue()}
        </Badge>
      ),
    }),
    returnColumnHelper.accessor("customerId", {
      header: "客戶",
      // 列表只顯示代號：`ReturnRequestEntry` 不帶客戶名稱，而逐列查詢會變 N+1；
      // 真名在明細對話框（一次 `getCustomer`）與退貨證明（快照自帶）裡看。
      cell: (info) => <span class="text-muted-foreground">#{info.getValue()}</span>,
    }),
    returnColumnHelper.accessor("remark", {
      header: "備註",
      cell: (info) => (
        <span class="text-muted-foreground">{info.getValue() || "—"}</span>
      ),
    }),
    returnColumnHelper.accessor("createdAt", {
      header: "建立時間",
      cell: (info) => (
        <span class="text-muted-foreground">{formatDateTime(info.getValue())}</span>
      ),
    }),
    returnColumnHelper.display({
      id: "actions",
      header: () => <span class="block text-right">操作</span>,
      cell: (info) => (
        <div class="flex justify-end gap-3">
          <button
            type="button"
            onClick={() => openDetail(info.row.original)}
            class="font-medium text-primary hover:underline"
          >
            查看
          </button>
          <Show when={info.row.original.status === "approved"}>
            {/* 證明是唯讀快照 → 掛在列上、明細框維持唯讀：兩個 Ark 對話框不同時開
                （同 `OrdersPage` 檔頭的「動作只掛在列上」慣例，避免對話框交接）。 */}
            <button
              type="button"
              onClick={() => setCertId(info.row.original.id)}
              class="font-medium text-primary hover:underline"
            >
              證明
            </button>
          </Show>
        </div>
      ),
    }),
  ]);

  const query = createQuery(() =>
    returnRequestsQueryOptions({
      status: statusFilter(),
      page: pagination().pageIndex + 1,
      pageSize: pagination().pageSize,
    })
  );

  const total = () => queryData(query, (d) => d?.total) ?? 0;

  const table = createTable({
    features: RETURN_TABLE_FEATURES,
    columns,
    get data() {
      return queryData(query, (d) => d?.entries) ?? NO_RETURNS;
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

  // 明細／證明對話框：id 為 null＝關閉；query 以 `enabled` 掛在開關上。
  const [detailId, setDetailId] = createSignal<string | null>(null);
  const [detailOpen, setDetailOpen] = createSignal(false);
  const [certId, setCertId] = createSignal<string | null>(null);
  // 審核相關的本地狀態：錯誤／成功訊息與駁回原因草稿。
  const [reviewError, setReviewError] = createSignal<string | null>(null);
  const [successMessage, setSuccessMessage] = createSignal<string | null>(null);
  const [rejecting, setRejecting] = createSignal(false);
  const [rejectReason, setRejectReason] = createSignal("");
  const [rejectError, setRejectError] = createSignal<string | undefined>();
  const [reviewing, setReviewing] = createSignal(false);

  const detailQuery = createQuery(() => ({
    ...returnRequestQueryOptions({ id: detailId() ?? "" }),
    enabled: detailOpen(),
  }));

  const certQuery = createQuery(() => ({
    ...returnCertificateQueryOptions({ id: certId() ?? "" }),
    enabled: certId() !== null,
  }));

  /**
   * 客戶名稱：查失敗（權限／跨部門）只退回代號，**不阻斷明細**——
   * 審核才是本頁的主任務，客戶名稱是附加上下文。
   */
  const customerQuery = createQuery(() => ({
    ...customerNameQueryOptions({ id: queryData(detailQuery, (d) => d?.customerId) ?? "" }),
    enabled: detailOpen() && detailQuery.data !== undefined,
  }));

  const customerLabel = () => {
    const customer = queryData(customerQuery, (d) => d?.customer);
    if (customer) return customer.name;
    if (customerQuery.error) return `客戶 #${queryData(detailQuery, (d) => d?.customerId) ?? ""}`;
    return "載入中…";
  };

  /** 送出篩選：把草稿套進 query key 並回第 1 頁（同一個 `batch` 內，避免兩次 RPC）。 */
  const submitFilter: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (e) => {
    e.preventDefault();
    setSuccessMessage(null);
    batch(() => {
      setStatusFilter(statusDraft());
      table.setPageIndex(0);
    });
  };

  /** 開啟明細：清掉上一筆審核的殘留狀態，避免舊錯誤疊在新明細上。 */
  const openDetail = (entry: ReturnRequestEntry) => {
    setSuccessMessage(null);
    setReviewError(null);
    setRejecting(false);
    setRejectReason("");
    setRejectError(undefined);
    setDetailId(entry.id);
    setDetailOpen(true);
  };

  /**
   * 審核：`expectedVersion` 取自剛讀到的明細 `version`（樂觀鎖的唯一合法來源）。
   * 失敗時 `InvalidArgument`（版本衝突／非待審核）一律失效 `["returnRequests"]`，
   * 讓使用者看到後端的最新狀態再重試。
   */
  const review = async (decision: "approved" | "rejected") => {
    const detail = queryData(detailQuery, (d) => d);
    if (!detail || reviewing()) return;
    setReviewError(null);
    setReviewing(true);
    try {
      await returnClient.reviewReturnRequest({
        id: detail.id,
        decision,
        rejectReason: decision === "rejected" ? rejectReason().trim() : "",
        expectedVersion: detail.version,
      });
      setDetailOpen(false);
      setRejecting(false);
      setRejectReason("");
      setRejectError(undefined);
      setSuccessMessage(decision === "approved" ? "已核准退貨申請" : "已駁回退貨申請");
      await client.invalidateQueries({ queryKey: ["returnRequests"] });
    } catch (err) {
      setReviewError(errorMessage(err));
      if (err instanceof ConnectError && err.code === Code.InvalidArgument) {
        await client.invalidateQueries({ queryKey: ["returnRequests"] });
      }
    } finally {
      setReviewing(false);
    }
  };

  /** 核准是破壞性操作 → 走 `window.confirm`（同 `CustomersPage` 刪除的慣例）。 */
  const approve = async () => {
    if (!window.confirm("確定核准此退貨申請？")) return;
    await review("approved");
  };

  const submitReject = async () => {
    const reason = rejectReason().trim();
    // 後端要求駁回必附原因；前端先擋，避免吃一次必然失敗的來回。
    if (!reason) {
      setRejectError("請輸入駁回原因");
      return;
    }
    setRejectError(undefined);
    await review("rejected");
  };

  const cancelReject = () => {
    setRejecting(false);
    setRejectReason("");
    setRejectError(undefined);
  };

  /** 明細／證明共用的品項快照表（兩處顯示的欄位與標籤必須一致，不複製兩份）。 */
  const itemsTable = (items: readonly ReturnRequestItemView[]) => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>商品名稱</TableHead>
          <TableHead>規格</TableHead>
          <TableHead>單位</TableHead>
          <TableHead>數量</TableHead>
          <TableHead>退貨原因</TableHead>
          <TableHead>來源</TableHead>
          <TableHead>照片</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <For each={items}>
          {(item) => (
            <TableRow>
              <TableCell class="font-medium text-foreground">{item.productName}</TableCell>
              <TableCell>{item.spec || "—"}</TableCell>
              <TableCell>{item.unit || "—"}</TableCell>
              <TableCell>{item.quantity}</TableCell>
              <TableCell>{item.reason || "—"}</TableCell>
              <TableCell>{SOURCE_LABELS[item.sourceType] ?? item.sourceType}</TableCell>
              <TableCell>
                <Show
                  when={item.photoUrls.length > 0}
                  fallback={<span class="text-muted-foreground">無照片</span>}
                >
                  <For each={item.photoUrls}>
                    {(url) => (
                      <a
                        href={url}
                        target="_blank"
                        rel="noopener noreferrer"
                        class="block font-medium text-primary hover:underline"
                      >
                        {url}
                      </a>
                    )}
                  </For>
                </Show>
              </TableCell>
            </TableRow>
          )}
        </For>
      </TableBody>
    </Table>
  );

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">退貨管理</h1>
          <p class="mt-1 text-sm text-muted-foreground">退貨申請清單（共 {total()} 筆）</p>
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

      {/* 建單入口的替代說明：發起權只在客戶子帳號（D25），員工側只看與審。 */}
      <p class="mb-4 rounded-lg bg-muted px-3 py-2 text-sm text-muted-foreground">
        退貨申請由客戶 App 的子帳號發起（員工與主帳號一律被拒絕），本頁只負責查看與審核。
      </p>

      <Card>
        <form
          class="flex flex-wrap items-end gap-3 border-b border-border bg-muted px-3 py-3"
          onSubmit={submitFilter}
        >
          <Field class="w-full sm:w-40">
            <FieldLabel for="return-status">狀態</FieldLabel>
            <select
              id="return-status"
              value={statusDraft()}
              onChange={(e) => setStatusDraft(e.currentTarget.value)}
            >
              <option value="">全部</option>
              <option value="pending">待審核</option>
              <option value="approved">已核准</option>
              <option value="rejected">已駁回</option>
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
                      <TableHead>
                        {flexRender(h.column.columnDef.header, h.getContext())}
                      </TableHead>
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
                  <TableCell colSpan={columns.length}>
                    {query.isFetching ? "載入中…" : "尚無退貨申請"}
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

      <Dialog open={detailOpen()} onOpenChange={setDetailOpen}>
        <DialogContent class="max-w-3xl">
          <DialogHeader>
            <DialogTitle>退貨申請明細</DialogTitle>
            <DialogDescription>查看品項快照並進行核准或駁回</DialogDescription>
          </DialogHeader>

          <Show when={detailQuery.error}>
            {(err) => (
              <p
                class="rounded-lg bg-destructive/15 px-3 py-2 text-sm text-destructive"
                role="alert"
              >
                {errorMessage(err())}
              </p>
            )}
          </Show>

          <Show when={reviewError()}>
            {(message) => (
              <p
                class="rounded-lg bg-destructive/15 px-3 py-2 text-sm text-destructive"
                role="alert"
              >
                {message()}
              </p>
            )}
          </Show>

          <Show when={queryData(detailQuery, (d) => d)}>
            {(detail) => (
              <div class="space-y-3 text-sm">
                <div class="grid gap-2 sm:grid-cols-2">
                  <div>
                    <span class="text-muted-foreground">客戶：</span>
                    {customerLabel()}
                  </div>
                  <div class="flex items-center gap-2">
                    <span class="text-muted-foreground">狀態：</span>
                    <Badge variant={STATUS_VARIANTS[detail().status] ?? "secondary"}>
                      {STATUS_LABELS[detail().status] ?? detail().status}
                    </Badge>
                  </div>
                  <div class="sm:col-span-2">
                    <span class="text-muted-foreground">備註：</span>
                    {detail().remark || "—"}
                  </div>
                  <Show when={detail().rejectReason}>
                    <div class="sm:col-span-2 rounded-lg bg-destructive/10 px-3 py-2 text-destructive">
                      <span>駁回原因：</span>
                      {detail().rejectReason}
                    </div>
                  </Show>
                  <div class="sm:col-span-2">
                    <span class="text-muted-foreground">建立時間：</span>
                    {formatDateTime(detail().createdAt)}
                  </div>
                </div>
                {itemsTable(detail().items)}
              </div>
            )}
          </Show>

          <Show when={rejecting()}>
            <Field invalid={rejectError() !== undefined}>
              <FieldLabel for="reject-reason">駁回原因 *</FieldLabel>
              <textarea
                id="reject-reason"
                rows={3}
                class="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground"
                value={rejectReason()}
                onInput={(e) => setRejectReason(e.currentTarget.value)}
                placeholder="例：商品已拆封無法退換"
              />
              <FieldError>{rejectError()}</FieldError>
            </Field>
            <div class="flex justify-end gap-3">
              <Button type="button" variant="outline" onClick={cancelReject}>
                取消
              </Button>
              <Button type="button" onClick={() => void submitReject()}>
                確認駁回
              </Button>
            </div>
          </Show>

          <DialogFooter>
            <Show when={queryData(detailQuery, (d) => d?.status) === "pending"}>
              <Button type="button" onClick={() => void approve()} disabled={reviewing()}>
                核准
              </Button>
              <Button type="button" variant="outline" onClick={() => setRejecting(true)}>
                駁回
              </Button>
            </Show>
            <DialogClose class={buttonVariants({ variant: "outline" })}>關閉</DialogClose>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={certId() !== null}
        onOpenChange={(open) => {
          if (!open) setCertId(null);
        }}
      >
        <DialogContent class="max-w-3xl">
          <DialogHeader>
            <DialogTitle>退貨證明</DialogTitle>
            <DialogDescription>核准後產生的證明快照</DialogDescription>
          </DialogHeader>

          <Show when={certQuery.error}>
            {(err) => (
              <p
                class="rounded-lg bg-destructive/15 px-3 py-2 text-sm text-destructive"
                role="alert"
              >
                {errorMessage(err())}
              </p>
            )}
          </Show>

          <Show when={queryData(certQuery, (d) => d)}>
            {(cert) => (
              <div class="space-y-3 text-sm">
                <div class="grid gap-2 sm:grid-cols-2">
                  <div>
                    <span class="text-muted-foreground">客戶編號：</span>
                    {cert().customerCode}
                  </div>
                  <div>
                    <span class="text-muted-foreground">客戶名稱：</span>
                    {cert().customerName}
                  </div>
                  <div>
                    <span class="text-muted-foreground">建立時間：</span>
                    {formatDateTime(cert().createdAt)}
                  </div>
                  <div class="flex items-center gap-2">
                    <span class="text-muted-foreground">狀態：</span>
                    <Badge variant={STATUS_VARIANTS[cert().status] ?? "secondary"}>
                      {STATUS_LABELS[cert().status] ?? cert().status}
                    </Badge>
                  </div>
                  <div>
                    <span class="text-muted-foreground">審核人：</span>
                    {cert().reviewerName || "—"}
                  </div>
                  <div>
                    <span class="text-muted-foreground">審核時間：</span>
                    {formatDateTime(cert().reviewedAt)}
                  </div>
                </div>
                {itemsTable(cert().items)}
              </div>
            )}
          </Show>

          <DialogFooter>
            <DialogClose class={buttonVariants({ variant: "outline" })}>關閉</DialogClose>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </main>
  );
}
