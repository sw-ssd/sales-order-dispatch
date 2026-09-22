import { Code, ConnectError } from "@connectrpc/connect";
import { createInfiniteQuery, createQuery, useQueryClient } from "@tanstack/solid-query";
import { createEffect, createMemo, createSignal, For, onCleanup, Show, type JSX } from "solid-js";
import type { Route } from "~/lib/proto/masters/v1/master_pb";
import type { SalesOrder } from "~/lib/proto/salesorder/v1/salesorder_pb";
import {
  Badge,
  Button,
  buttonVariants,
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Field,
  FieldLabel,
  Input,
} from "~/components/ui";
import { boardOrdersQueryOptions, boardRoutesQueryOptions, dispatchClient } from "../queries";

/** 看板只處理這兩個狀態（dispatch spec：看板僅顯示 pending 或 processing 的訂單）。 */
const BOARD_STATUSES = new Set(["pending", "processing"]);

function errorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    switch (err.code) {
      case Code.NotFound:
        return "資料不存在或已被刪除";
      case Code.InvalidArgument:
        return err.rawMessage || "輸入資料有誤,請檢查後再試";
      case Code.FailedPrecondition:
        // 樂觀鎖衝突與狀態衝突都走這裡；後端訊息已含「資料已變更，請重新載入」。
        return err.rawMessage || "資料已被他人變更，請重新整理看板";
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

/** 今天（本地時區）的 YYYY-MM-DD；看板預設落在今天。 */
function today(): string {
  const d = new Date();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${d.getFullYear()}-${mm}-${dd}`;
}

/**
 * 派車看板頁(/dispatch)。
 *
 * 版型是看板而非清單：車次為欄、訂單為卡片（依 `delivery_sequence` 排序）、未指派獨立欄；
 * 日期是頁面主軸（非篩選草稿）——選日期即全量重查，符合 dispatch spec 的「依日期篩選檢視看板」。
 *
 * 三處契約（相較其他頁）：
 * - **日期由後端過濾**（`ListOrders.expected_delivery_date`）：不能改前端過濾，
 *   後端 `maxPageSize=100`，日期外訂單會把該日的擠出分頁。
 * - **樂觀鎖**：每次指派都帶讀取時的 `version`；失敗一律重查看板（spec：衝突後重新整理至最新狀態）。
 * - **事件僅失效提示**：`WatchBoard` 收到非 heartbeat 事件就使 `["boardOrders"]` 失效並全量重查，
 *   事件內容本身不進 UI 快取（spec：狀態正確性以重查結果為準）。
 */
export default function DispatchPage() {
  const [date, setDate] = createSignal(today());
  const [dragging, setDragging] = createSignal<SalesOrder | null>(null);
  const [actionError, setActionError] = createSignal<string | null>(null);
  // 取消派車：目標訂單、原因、是否已確認重印。
  const [cancelTarget, setCancelTarget] = createSignal<SalesOrder | null>(null);
  const [cancelReason, setCancelReason] = createSignal("");
  const [reprintAck, setReprintAck] = createSignal(false);
  // 批次確認結果（部分失敗語義需逐筆呈現）。
  const [confirmResult, setConfirmResult] = createSignal<
    { successCount: number; failures: { orderId: string; reason: string }[] } | null
  >(null);

  const client = useQueryClient();

  const orders = createInfiniteQuery(() => boardOrdersQueryOptions(date()));
  const routes = createQuery(() => boardRoutesQueryOptions());

  /**
   * 頁首錯誤：查詢失敗優先於操作失敗。
   *
   * 查詢失敗（403／斷線）若只留在 `orders.error`，看板會永遠停在「載入中」而無任何提示
   * —— 使用者看不出是沒權限還是後端掛了，所以兩類錯誤都必須浮到 banner。
   */
  const bannerError = () =>
    orders.error
      ? errorMessage(orders.error)
      : routes.error
        ? errorMessage(routes.error)
        : actionError();

  /** 該日可上板的訂單：只留 pending/processing，並排除未指派的處理中單（它們已在車次欄）。 */
  const boardable = createMemo(() =>
    (orders.data?.pages ?? []).flatMap((p) => p.orders).filter((o) => BOARD_STATUSES.has(o.status))
  );

  /** 活躍車次＝看板的欄；停用車次不作派車目標。 */
  const activeRoutes = createMemo(
    () => (routes.data?.routes ?? []).filter((r) => r.isActive && !r.deletedAt)
  );

  /** 未指派欄：pending 且未綁車次。 */
  const unassigned = createMemo(() =>
    boardable()
      .filter((o) => o.status === "pending" && !o.routeId)
      .sort((a, b) => a.orderNo.localeCompare(b.orderNo))
  );

  /** 各車次欄的卡片：同車次內依配送順位升冪（spec：卡片依 delivery_sequence 排序）。 */
  const byRoute = createMemo(() => {
    const map = new Map<string, SalesOrder[]>();
    for (const o of boardable()) {
      if (!o.routeId) continue;
      const list = map.get(o.routeId) ?? [];
      list.push(o);
      map.set(o.routeId, list);
    }
    for (const list of map.values()) {
      list.sort((a, b) => a.deliverySequence - b.deliverySequence || a.orderNo.localeCompare(b.orderNo));
    }
    return map;
  });

  /** 車次欄截斷警告：車次取回筆數達單頁上限時，前端不假裝只有這些車次。 */
  const routesTruncated = createMemo(
    () => Number(routes.data?.pagination?.total ?? 0) > (routes.data?.routes.length ?? 0)
  );

  /** 所有操作的共用收尾：一律重查看板（樂觀鎖失敗也要回到伺服器最新狀態，spec 要求）。 */
  const refreshBoard = async () => {
    await client.invalidateQueries({ queryKey: ["boardOrders"] });
  };

  const assignTo = async (order: SalesOrder, routeId: string, deliverySequence: string) => {
    setActionError(null);
    try {
      await dispatchClient.assignRoute({
        salesOrderId: order.id,
        routeId,
        deliverySequence,
        // 樂觀鎖：帶讀取時的 version；後端比對不符即 failed_precondition。
        version: String(order.version),
        expectedDeliveryDate: date(),
      });
    } catch (err) {
      setActionError(errorMessage(err));
    } finally {
      setDragging(null);
      await refreshBoard();
    }
  };

  /** 拖進車次欄尾：新順位 = 該欄最大順位 + 1（後端會把 ≥ 該順位者依序後移）。 */
  const dropAtEnd = (routeId: string) => {
    const order = dragging();
    if (!order) return;
    const inColumn = byRoute().get(routeId) ?? [];
    const maxSeq = inColumn.reduce((m, o) => Math.max(m, o.deliverySequence), 0);
    void assignTo(order, routeId, String(maxSeq + 1));
  };

  /** 拖到某張卡之前：以該卡順位插入，後端 shiftSequence 把它與之後的往後移。 */
  const dropBefore = (target: SalesOrder) => {
    const order = dragging();
    if (!order || order.id === target.id) return;
    if (!target.routeId) {
      // 未指派欄的卡片不是插入錨點（它們沒有順位）→ 退回欄尾語意不適用,忽略。
      setDragging(null);
      return;
    }
    void assignTo(order, target.routeId, String(Math.max(1, target.deliverySequence)));
  };

  /** 拖回未指派：route_id 帶空字串即清除車次與順位（後端 ClearRouteID）。 */
  const dropUnassigned = () => {
    const order = dragging();
    if (!order) return;
    void assignTo(order, "", "");
  };

  const confirmRoute = async (routeId: string) => {
    setActionError(null);
    setConfirmResult(null);
    try {
      const res = await dispatchClient.confirmDispatch({
        routeId,
        expectedDeliveryDate: date(),
      });
      const failures = (res.items ?? [])
        .filter((it) => !it.success)
        .map((it) => ({ orderId: it.salesOrderId, reason: it.failReason || "未知原因" }));
      setConfirmResult({ successCount: res.successCount, failures });
    } catch (err) {
      setActionError(errorMessage(err));
    } finally {
      await refreshBoard();
    }
  };

  const openCancel = (order: SalesOrder) => {
    setCancelTarget(order);
    setCancelReason("");
    setReprintAck(false);
    setActionError(null);
  };

  const submitCancel = async () => {
    const target = cancelTarget();
    if (!target) return;
    const reason = cancelReason().trim();
    if (reason === "") {
      setActionError("取消派車需填原因");
      return;
    }
    setActionError(null);
    try {
      const res = await dispatchClient.cancelDispatch({
        salesOrderId: target.id,
        reason,
        acknowledgeReprint: reprintAck(),
      });
      if (res.reprintWarning && !reprintAck()) {
        // 已列印車次：後端只回警告、不執行；前端補確認後再送一次（spec 的重印確認）。
        setReprintAck(true);
        setActionError("該車次當日已列印，請確認重印後再送出一次");
        return;
      }
      setCancelTarget(null);
    } catch (err) {
      setActionError(errorMessage(err));
    } finally {
      await refreshBoard();
    }
  };

  /**
   * 看板訂閱：日期變更即重新訂閱；收到非 heartbeat 事件就使看板失效並全量重查。
   *
   * 斷線不重試（spec：事件僅失效提示、不保證補發；狀態正確性以重查結果為準），
   * 看板仍由查詢驅動，斷線期間的變更會在下次查詢帶回。
   */
  createEffect(() => {
    // 訂閱日期是本輪的快照:串流必須綁單一日期,翻日由本 effect 重跑改訂閱(不可在迴圈內再讀 date)。
    const staticSubscribedDate = date();
    const ac = new AbortController();
    let stopped = false;
    void (async () => {
      try {
        for await (const ev of dispatchClient.watchBoard(
          { expectedDeliveryDate: staticSubscribedDate },
          { signal: ac.signal }
        )) {
          if (stopped) break;
          if (ev.type === "heartbeat") continue;
          void client.invalidateQueries({ queryKey: ["boardOrders"] });
        }
      } catch {
        // 連線中止或伺服器關閉：非錯誤路徑（看板以查詢為準），刻意不提示。
      }
    })();
    onCleanup(() => {
      stopped = true;
      ac.abort();
    });
  });

  /** 欄位級放置目標（`<section>` 是 HTMLElement，型別要對齊它的元素）。 */
  const dropTarget = (onDrop: () => void): JSX.HTMLAttributes<HTMLElement> => ({
    onDragOver: (e) => e.preventDefault(),
    onDrop,
  });

  /** 訂單卡片：可拖、可作為「插到我前面」的放置目標（事件不冒泡到欄位）。 */
  const cardAttrs = (order: SalesOrder): JSX.HTMLAttributes<HTMLDivElement> => ({
    draggable: true,
    onDragStart: () => setDragging(order),
    onDragOver: (e) => e.preventDefault(),
    onDrop: (e) => {
      e.stopPropagation();
      dropBefore(order);
    },
  });

  return (
    <main>
      <header class="mb-6 flex flex-col gap-4 border-b-2 border-border pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-foreground">派車看板</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            {date()}・待派 {unassigned().length} 筆、車次內 {boardable().length - unassigned().length} 筆
          </p>
        </div>
        <div class="flex items-end gap-3">
          <Field class="w-44">
            <FieldLabel for="board-date">預計出貨日</FieldLabel>
            <input
              id="board-date"
              type="date"
              value={date()}
              onChange={(e) => setDate(e.currentTarget.value)}
            />
          </Field>
          <Button
            type="button"
            variant="outline"
            loading={orders.isFetching}
            onClick={() => void refreshBoard()}
          >
            重新整理
          </Button>
        </div>
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

      <Show when={confirmResult()}>
        {(r) => (
          <p
            class="mb-4 rounded-lg bg-success/15 px-3 py-2 text-sm font-medium text-success"
            role="status"
          >
            批次確認完成：成功 {r().successCount} 筆
            <Show when={r().failures.length > 0}>
              <span class="text-destructive">
                、失敗 {r().failures.length} 筆（
                {r()
                  .failures.map((f) => `${f.orderId}: ${f.reason}`)
                  .join("；")}
                ）
              </span>
            </Show>
          </p>
        )}
      </Show>

      <Show when={routesTruncated()}>
        <p class="mb-4 rounded-lg bg-warning/15 px-3 py-2 text-sm font-medium text-warning" role="status">
          車次筆數超出單頁上限，僅顯示前 {routes.data?.routes.length} 筆。
        </p>
      </Show>

      <Show
        when={!orders.isPending && !routes.isPending}
        fallback={<p class="text-sm text-muted-foreground">看板載入中…</p>}
      >
        <div class="flex gap-4 overflow-x-auto pb-4">
          {/* 未指派欄：pending 且未綁車次（spec：獨立的「未指派」區域）。 */}
          <section
            class="w-72 shrink-0 rounded-lg border border-dashed border-border bg-muted/50 p-3"
            aria-label="未指派"
            {...dropTarget(dropUnassigned)}
          >
            <h2 class="mb-3 flex items-center justify-between text-sm font-semibold text-foreground">
              未指派
              <Badge variant="secondary">{unassigned().length}</Badge>
            </h2>
            <div class="space-y-2">
              <For each={unassigned()}>
                {(o) => (
                  <div class="cursor-grab rounded-md border border-border bg-card p-2 text-sm" {...cardAttrs(o)}>
                    <div class="font-medium text-foreground">{o.orderNo}</div>
                    <div class="text-xs text-muted-foreground">
                      出貨日 {o.expectedDeliveryDate || "—"}
                    </div>
                    {/*
                      非拖曳指派路徑。HTML5 drag-and-drop 在觸控裝置（倉庫端的 iPad）與純鍵盤下
                      完全不會觸發，只剩拖曳一條路時這兩種使用者根本無法派車 —— 下拉是同一顆
                      assignTo 的另一個入口，語意等於「拖到該欄尾」（順位＝最大順位 +1）。
                      一律顯示（不限定 pointer-coarse）：鍵盤使用者同樣沒有拖曳。
                    */}
                    <select
                      class="mt-2 w-full cursor-pointer rounded-md border border-border bg-background px-2 py-1 text-xs text-foreground"
                      aria-label={`指派 ${o.orderNo} 到車次`}
                      onChange={(e) => {
                        const routeId = e.currentTarget.value;
                        if (!routeId) return;
                        const maxSeq = (byRoute().get(routeId) ?? []).reduce(
                          (m, x) => Math.max(m, x.deliverySequence),
                          0
                        );
                        void assignTo(o, routeId, String(maxSeq + 1));
                      }}
                    >
                      <option value="">指派到車次…</option>
                      <For each={activeRoutes()}>
                        {(r) => <option value={r.id}>{r.name}（{r.code}）</option>}
                      </For>
                    </select>
                  </div>
                )}
              </For>
              <Show when={unassigned().length === 0}>
                <p class="text-xs text-muted-foreground">無待派訂單</p>
              </Show>
            </div>
          </section>

          <For each={activeRoutes()}>
            {(route: Route) => {
              const cards = () => byRoute().get(route.id) ?? [];
              const canConfirm = () => cards().some((o) => o.status === "pending");
              return (
                <section
                  class="w-72 shrink-0 rounded-lg border border-border bg-card p-3"
                  aria-label={route.name}
                  {...dropTarget(() => dropAtEnd(route.id))}
                >
                  <h2 class="mb-3 flex items-center justify-between text-sm font-semibold text-foreground">
                    <span class="truncate">
                      {route.name}
                      <span class="ml-1 font-normal text-muted-foreground">({route.code})</span>
                    </span>
                    <Badge variant="secondary">{cards().length}</Badge>
                  </h2>
                  <div class="space-y-2">
                    <For each={cards()}>
                      {(o) => (
                        <div class="cursor-grab rounded-md border border-border bg-background p-2 text-sm" {...cardAttrs(o)}>
                          <div class="flex items-center justify-between gap-2">
                            <span class="font-medium text-foreground">{o.orderNo}</span>
                            <Badge variant={o.status === "processing" ? "success" : "info"}>
                              {o.status === "processing" ? "已派車" : "待派"}
                            </Badge>
                          </div>
                          <div class="mt-1 text-xs text-muted-foreground">
                            順位 {o.deliverySequence || "—"}・{o.source}
                          </div>
                          <div class="mt-2 flex justify-end gap-3 text-xs">
                            <Show when={o.status === "pending"}>
                              <button
                                type="button"
                                class="font-medium text-destructive hover:underline"
                                onClick={() => assignTo(o, "", "")}
                              >
                                退回未指派
                              </button>
                            </Show>
                            <Show when={o.status === "processing"}>
                              <button
                                type="button"
                                class="font-medium text-destructive hover:underline"
                                onClick={() => openCancel(o)}
                              >
                                取消派車
                              </button>
                            </Show>
                          </div>
                        </div>
                      )}
                    </For>
                    <Show when={cards().length === 0}>
                      <p class="text-xs text-muted-foreground">拖入訂單以指派此車次</p>
                    </Show>
                  </div>
                  <div class="mt-3 border-t border-border pt-2">
                    <Button
                      type="button"
                      size="sm"
                      class="w-full"
                      disabled={!canConfirm()}
                      onClick={() => void confirmRoute(route.id)}
                    >
                      批次確認派車
                    </Button>
                  </div>
                </section>
              );
            }}
          </For>

          <Show when={activeRoutes().length === 0}>
            <p class="text-sm text-muted-foreground">尚無啟用中的車次，請先於部門主檔建立路線。</p>
          </Show>
        </div>
      </Show>

      {/* 取消派車（必填原因；已列印車次需確認重印） */}
      <Dialog
        open={cancelTarget() !== null}
        onOpenChange={(open) => {
          if (!open) setCancelTarget(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>取消派車</DialogTitle>
            <DialogDescription>
              {cancelTarget()?.orderNo}（保留車次與順位，退回待派）
            </DialogDescription>
          </DialogHeader>
          <Show when={reprintAck()}>
            <p class="rounded-lg bg-warning/15 px-3 py-2 text-sm font-medium text-warning" role="alert">
              該車次當日已列印，確認後將產生重印。
            </p>
          </Show>
          <Field>
            <FieldLabel for="cancel-reason">取消原因 *</FieldLabel>
            <Input
              id="cancel-reason"
              value={cancelReason()}
              onInput={(e) => setCancelReason(e.currentTarget.value)}
            />
          </Field>
          <DialogFooter>
            <DialogClose class={buttonVariants({ variant: "outline" })}>
              取消
            </DialogClose>
            <Button type="button" onClick={() => void submitCancel()}>
              {reprintAck() ? "確認取消並重印" : "確認取消"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </main>
  );
}
