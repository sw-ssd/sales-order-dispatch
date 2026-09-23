import { createQuery } from "@tanstack/solid-query";
import { Show } from "solid-js";
import { Link } from "@tanstack/solid-router";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Spinner,
} from "~/components/ui";
import { Can } from "~/lib/ability/Can";
import { useAbility } from "~/lib/ability/context";
import { hasPermission } from "~/lib/ability/permissions";
import type { NavRoute } from "~/components/layout/Sidebar";
import {
  pendingOrderCountQueryOptions,
  processingOrderCountQueryOptions,
  todayPendingQueryOptions,
  todayProcessingQueryOptions,
} from "../queries";

/** 今天（本地時區）的 YYYY-MM-DD；與派車看板同一套 local-date 慣例（見 queries.ts 檔頭）。 */
function today(): string {
  const d = new Date();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${d.getFullYear()}-${mm}-${dd}`;
}

/**
 * 數字卡：標題、大數字、前往該清單的連結。
 *
 * `count` 為 undefined 代表查詢失敗（呼叫端已把 `total` 收斂成 undefined）——
 * 顯示 `—` 而非 0：0 是「真的沒有」，兩者在營運判斷上意義相反。
 */
function StatCard(props: {
  title: string;
  count: number | undefined;
  to: string;
  hint: string;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{props.title}</CardTitle>
        <CardDescription>{props.hint}</CardDescription>
      </CardHeader>
      <CardContent class="flex items-end justify-between gap-4">
        <p class="text-3xl font-bold tabular-nums text-foreground">
          {props.count ?? "—"}
        </p>
        <Link
          to={props.to}
          class="text-sm font-medium text-primary underline-offset-4 hover:underline"
        >
          檢視清單
        </Link>
      </CardContent>
    </Card>
  );
}

/**
 * 首頁 Dashboard（規格 §8.2：今日待出貨、待處理訂單數量、快速連結）。
 *
 * 三處契約：
 * - **每張卡各自一次查詢**（`pageSize: 1`，只取 `total`）：不合併成一個 RPC，因為條件不同、
 *   失效時機也不同；單卡失敗不影響其他卡（失敗顯示 `—`）。今日待出貨是兩支查詢相加，
 *   理由見 `todayTotal`。
 * - **失敗顯示 `—` 不顯示 0**：`0` 與「查不到」在營運上是相反的結論（見 `StatCard`）。
 * - **快捷連結隨權限出現**：`Can` 包住，沒有該頁權限的人不會看到連過去必定 403 的入口
 *   （前端守衛不構成授權，但能避免誤導；實際授權仍在後端）。
 */
export default function DashboardPage() {
  const ability = useAbility();
  /**
   * 沒有 `sales_order:read` 的人（例如 `customer` 角色）不該打這些查詢 ——
   * 打了必定 403，於是整頁只剩錯誤訊息。`enabled` 讓查詢根本不送出。
   */
  const canReadOrders = () => hasPermission(ability(), "sales_order", "read");

  /**
   * 今日待出貨＝今日 pending ＋ 今日 processing 兩支查詢相加。
   *
   * `ListOrdersRequest.status` 是**單一字串**（proto 無重複欄位），表達不了「pending 或
   * processing」；而今日出貨的單不可能已完成（completed 是已出貨、cancelled/voided 不會出），
   * 所以這兩個狀態正好就是全部要出貨的。前端相加，後端不必為首頁加聚合 RPC。
   */
  const todayPending = createQuery(() => ({
    ...todayPendingQueryOptions(today()),
    enabled: canReadOrders(),
  }));
  const todayProcessing = createQuery(() => ({
    ...todayProcessingQueryOptions(today()),
    enabled: canReadOrders(),
  }));
  const pending = createQuery(() => ({
    ...pendingOrderCountQueryOptions(),
    enabled: canReadOrders(),
  }));
  const processing = createQuery(() => ({
    ...processingOrderCountQueryOptions(),
    enabled: canReadOrders(),
  }));

  /** 今日待出貨：兩支都成功才有值；任一失敗回 undefined → 顯示「—」。 */
  const todayTotal = (): number | undefined => {
    const a = todayPending.data?.total;
    const b = todayProcessing.data?.total;
    return a === undefined || b === undefined ? undefined : a + b;
  };

  const isPending = () =>
    todayPending.isPending ||
    todayProcessing.isPending ||
    pending.isPending ||
    processing.isPending;

  const isError = () =>
    todayPending.isError ||
    todayProcessing.isError ||
    pending.isError ||
    processing.isError;

  return (
    <main>
      <header class="mb-6 border-b-2 border-border pb-4">
        <h1 class="text-2xl font-bold text-foreground">多公司訂出貨系統</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          今日概況與常用入口
        </p>
      </header>

      <section class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Can I="read" a="sales_order">
          <StatCard
            title="今日待出貨"
            count={todayTotal()}
            to="/orders"
            hint={`出貨日為今天（${today()}）且尚未出貨的訂單`}
          />
          <StatCard
            title="待處理訂單"
            count={pending.data?.total}
            to="/orders"
            hint="尚未派車的訂單（不限日期）"
          />
          <StatCard
            title="處理中訂單"
            count={processing.data?.total}
            to="/dispatch"
            hint="已派車、尚未完成的訂單"
          />
        </Can>
      </section>

      <Show when={canReadOrders() && isPending()}>
        <p class="mt-4 flex items-center gap-2 text-sm text-muted-foreground">
          <Spinner size="sm" label="載入概況" />
          概況載入中…
        </p>
      </Show>

      <Show when={canReadOrders() && isError()}>
        <p class="mt-4 text-sm text-muted-foreground">
          部分概況暫時無法取得（顯示為 <span class="font-semibold">—</span>）；
          這不影響其他操作，請稍後重新整理。
        </p>
      </Show>

      <section class="mt-8">
        <h2 class="mb-3 text-lg font-semibold text-foreground">快速連結</h2>
        <ul class="flex flex-wrap gap-3">
          <Can I="read" a="sales_order">
            <QuickLink to="/orders" label="訂單管理" />
          </Can>
          <Can I="read" a="dispatch">
            <QuickLink to="/dispatch" label="派車規劃" />
          </Can>
          <Can I="read" a="print">
            <QuickLink to="/printing" label="單據列印" />
          </Can>
          <Can I="read" a="customer">
            <QuickLink to="/customers" label="客戶總表" />
          </Can>
          <Can I="read" a="product">
            <QuickLink to="/products" label="商品總表" />
          </Can>
          <Can I="read" a="notification">
            <QuickLink to="/notifications" label="通知中心" />
          </Can>
        </ul>
      </section>
    </main>
  );
}

/** 快速連結：中性外框按鈕，與卡片的大數字視覺區隔。 */
function QuickLink(props: { to: NavRoute; label: string }) {
  return (
    <li>
      <Link
        to={props.to}
        class="inline-flex items-center rounded-lg border border-border bg-card px-4 py-2 text-sm font-medium text-foreground shadow-xs transition-colors hover:bg-muted"
      >
        {props.label}
      </Link>
    </li>
  );
}
