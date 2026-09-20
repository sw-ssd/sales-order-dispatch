import { Match, Switch } from "solid-js";
import { Badge } from "@ui/badge";

/**
 * 訂閱狀態徽章：文案與色階只在此處定義（後端回的是 `platform.v1.TenantSummary.subscription_status`
 * 的字面值：trialing | active | past_due | suspended | cancelled | none）。
 * 純呈現——狀態能不能改由後端決定（見 lib/guard.ts 的說明）。
 */
const STATUS: Record<
  string,
  { label: string; variant: "success" | "info" | "warning" | "destructive" | "secondary" }
> = {
  trialing: { label: "試用中", variant: "info" },
  active: { label: "訂閱中", variant: "success" },
  past_due: { label: "逾期未付", variant: "warning" },
  suspended: { label: "已停用", variant: "destructive" },
  cancelled: { label: "已取消", variant: "secondary" },
  none: { label: "無訂閱", variant: "secondary" },
};

export function SubscriptionBadge(props: { status: string }) {
  return (
    <Switch fallback={<Badge variant="outline">{props.status}</Badge>}>
      <Match when={STATUS[props.status]}>
        {(known) => <Badge variant={known().variant}>{known().label}</Badge>}
      </Match>
    </Switch>
  );
}
