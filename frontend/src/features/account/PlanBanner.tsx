import { createQuery } from "@tanstack/solid-query";
import { Link } from "@tanstack/solid-router";
import { For, Show } from "solid-js";
import { cn } from "~/lib/cn";
import { ALERT_CLASS, entitlementAlerts } from "./entitlements";
import { tenantEntitlementsQueryOptions } from "./queries";

/**
 * shell 的權益提示條（spec §2.4 規則 2）：**只在**用量達 80/90% 或試用將到期時出現，
 * 其餘（未達門檻／載入中／載入失敗／尚未開通計費）一律不渲染任何東西——不吵是它的契約。
 *
 * 與帳號頁的卡片共用同一個查詢（同一個 queryKey ⇒ 只打一次）；失敗時靜默，不影響任何操作。
 */
export default function PlanBanner() {
  const query = createQuery(() => tenantEntitlementsQueryOptions());
  /** 沒有資料（載入中／失敗）就沒有提示——提示條只能由後端回的投影決定。 */
  const alerts = () => (query.data ? entitlementAlerts(query.data) : []);

  return (
    <Show when={alerts().length > 0}>
      <div
        role="status"
        class={cn(
          "flex flex-wrap items-center gap-x-3 gap-y-1 border-b px-4 py-2 text-sm lg:px-6",
          // 多筆提示時以最嚴重者為準（警示不因另一筆較輕而被淡化）。
          ALERT_CLASS[alerts().some((alert) => alert.severity === "urgent") ? "urgent" : "warning"]
        )}
      >
        <For each={alerts()}>{(alert) => <span class="font-semibold">{alert.text}</span>}</For>
        <Link to="/account" class="ms-auto underline underline-offset-4">
          查看方案
        </Link>
      </div>
    </Show>
  );
}
