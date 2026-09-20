import { createQuery } from "@tanstack/solid-query";
import { For, Match, Show, Switch } from "solid-js";
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Spinner,
} from "~/components/ui";
import { cn } from "~/lib/cn";
import {
  ALERT_CLASS,
  entitlementAlerts,
  featureLabel,
  featureListNote,
  subscriptionStatus,
  trialEndsAtFor,
  usageValue,
} from "./entitlements";
import { tenantEntitlementsQueryOptions } from "./queries";

/**
 * 租戶後台的**唯讀**權益卡片（spec §2.4 第②面、§4.6）：方案、狀態、配額用量（`8/10`）與試用到期。
 *
 * 三條規矩：
 * 1. **唯讀且不吵**：沒有任何寫入；提示只在用量達 80/90% 或試用將到期時出現（`entitlementAlerts`）。
 * 2. **不誤導**：載入中／載入失敗都不會長得像「沒有權益」；無訂閱列（`status=none`）不寫成
 *    「你的方案不含此功能」——前端分不出那個狀態與「方案未含」，判定在後端。
 * 3. **前端不構成授權**：卡片只是投影，按鈕是否可用一律由後端擋（載入失敗也不阻擋其他操作）。
 */
export default function PlanCard() {
  const query = createQuery(() => tenantEntitlementsQueryOptions());

  return (
    <Card class="max-w-2xl">
      <CardHeader>
        <CardTitle>方案與用量</CardTitle>
        <CardDescription>唯讀資訊；變更方案或席位請洽營運。</CardDescription>
      </CardHeader>

      <CardContent class="flex flex-col gap-4">
        <Switch>
          <Match when={query.isPending}>
            <p class="flex items-center gap-2 text-sm text-muted-foreground">
              <Spinner size="sm" label="載入方案資訊" />
              方案資訊載入中…
            </p>
          </Match>

          <Match when={query.isError}>
            <p class="text-sm font-semibold">方案資訊暫時無法取得</p>
            <p class="text-sm text-muted-foreground">
              本卡片載入失敗不影響其他操作；前端不構成授權判定，功能可用性一律由後端決定。
            </p>
            <Button
              variant="outline"
              size="sm"
              class="self-start"
              onClick={() => void query.refetch()}
            >
              重試
            </Button>
          </Match>

          <Match when={query.data}>
            {(entitlements) => (
              <>
                <div class="flex flex-wrap items-center gap-2">
                  <p class="text-lg font-semibold">
                    {entitlements().planName || entitlements().planCode || "未指派方案"}
                  </p>
                  <Badge variant={subscriptionStatus(entitlements().status).variant}>
                    {subscriptionStatus(entitlements().status).label}
                  </Badge>
                </div>

                <Show when={trialEndsAtFor(entitlements().status, entitlements().trialEndsAt)}>
                  {(date) => (
                    <p class="text-sm text-muted-foreground">
                      試用到期：<span class="font-medium">{date()}</span>（UTC）
                    </p>
                  )}
                </Show>

                <Show when={entitlementAlerts(entitlements()).length > 0}>
                  <ul class="flex flex-col gap-2">
                    <For each={entitlementAlerts(entitlements())}>
                      {(alert) => (
                        <li
                          class={cn(
                            "rounded-lg border px-3 py-2 text-sm font-semibold",
                            ALERT_CLASS[alert.severity]
                          )}
                        >
                          {alert.text}
                        </li>
                      )}
                    </For>
                  </ul>
                </Show>

                <Show
                  when={featureListNote(entitlements())}
                  fallback={
                    <ul class="flex flex-col divide-y divide-border">
                      <For each={entitlements().usage}>
                        {(usage) => (
                          <li class="flex items-center justify-between gap-4 py-2 text-sm">
                            <span class="text-muted-foreground">
                              {featureLabel(usage.featureCode)}
                            </span>
                            <span class="font-semibold tabular-nums">{usageValue(usage)}</span>
                          </li>
                        )}
                      </For>
                    </ul>
                  }
                >
                  {(note) => <p class="text-sm text-muted-foreground">{note()}</p>}
                </Show>

                <p class="text-xs text-muted-foreground">
                  前端不構成授權判定：本卡片僅供參考，功能可用性由後端決定。
                </p>
              </>
            )}
          </Match>
        </Switch>
      </CardContent>
    </Card>
  );
}
