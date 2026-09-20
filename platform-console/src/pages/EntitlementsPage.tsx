import { createQuery } from "@tanstack/solid-query";
import { For, Show } from "solid-js";
import { Badge } from "@ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@ui/table";
import { EmptyState, PageShell, queryBoundary } from "../components/page";
import { platform } from "../lib/api";
import type { Feature, FeatureEntitlement } from "../lib/proto/platform/v1/platform_pb";

/**
 * 權益矩陣（**唯讀**）：完整功能清單（列）× 各方案（欄）的權益。
 *
 * 這裡**只讀不寫**：方案權益的變更在「方案」頁（`SetPlanEntitlement` 已在那裡接好必填原因與
 * 快取失效）。同一件事有兩個寫入入口，就一定會有一個變成半套的。
 *
 * 兩條與後端一致的語意（各方案都要看得出差別）：
 * 1. **未設定的功能＝未開通**（fail-closed；`GetPlanEntitlements` 會回完整功能清單正是為此）。
 * 2. 方案權益的 `limit_set=false` ＝**不限額**（與租戶例外的「不覆寫限額」不同，見 proto 註解），
 *    所以「不限」與「上限 0」在畫面上必須是兩個不同的字。
 *
 * 逐方案各查一次 `GetPlanEntitlements`（v1 沒有「一次回全部方案權益」的 RPC）。
 */
export default function EntitlementsPage() {
  const plans = createQuery(() => ({ queryKey: ["plans"], queryFn: () => platform.listPlans({}) }));
  const planCodes = () => plans.data?.plans.map((plan) => plan.code) ?? [];

  const matrix = createQuery(() => ({
    queryKey: ["plan-entitlements", planCodes()],
    queryFn: () =>
      Promise.all(planCodes().map((planCode) => platform.getPlanEntitlements({ planCode }))),
    enabled: plans.data !== undefined,
  }));

  /** 功能清單取各方案回應的聯集（同一份全域目錄，聯集只是不讓某一份回應決定列集合）。 */
  const features = (): Feature[] => {
    const byCode = new Map<string, Feature>();
    for (const response of matrix.data ?? []) {
      for (const feature of response.features) {
        if (!byCode.has(feature.code)) byCode.set(feature.code, feature);
      }
    }
    return [...byCode.values()];
  };

  /** 一格＝某方案對某功能的權益：未設定（未開通），或「啟用／停用 ＋ 上限」。 */
  const cell = (feature: Feature, entitlement: FeatureEntitlement | undefined) => (
    <TableCell>
      <Show when={entitlement} fallback={<Badge variant="secondary">未開通</Badge>}>
        {(known) => (
          <div class="flex flex-col items-start gap-1">
            <Badge variant={known().enabled ? "success" : "secondary"}>
              {known().enabled ? "啟用" : "停用"}
            </Badge>
            <Show when={feature.type === "integer"}>
              <span class="text-xs text-muted-foreground">
                {known().limitSet ? `${known().limitValue} ${feature.unit}`.trim() : "不限"}
              </span>
            </Show>
          </div>
        )}
      </Show>
    </TableCell>
  );

  return (
    <PageShell
      title="方案權益矩陣"
      description="功能 × 方案；未設定的功能視為未開通，上限「不限」表示該方案不限額。變更請到「方案」頁（本頁唯讀）。"
    >
      {queryBoundary(plans, (data) =>
        data.plans.length === 0 ? (
          <EmptyState>尚未定義任何方案。</EmptyState>
        ) : (
          queryBoundary(matrix, (perPlan) => (
            <Table aria-label="方案權益矩陣">
              <TableHeader>
                <TableRow>
                  <TableHead scope="col">功能</TableHead>
                  <For each={data.plans}>
                    {(plan) => (
                      <TableHead scope="col">
                        {plan.name}
                        <span class="ml-1 font-normal text-muted-foreground">{plan.code}</span>
                      </TableHead>
                    )}
                  </For>
                </TableRow>
              </TableHeader>
              <TableBody>
                <For each={features()}>
                  {(feature) => (
                    <TableRow>
                      <TableCell>
                        {feature.description}
                        <span class="ml-2 text-xs text-muted-foreground">{feature.code}</span>
                      </TableCell>
                      <For each={perPlan}>
                        {(response) =>
                          cell(
                            feature,
                            response.entitlements.find((it) => it.featureCode === feature.code),
                          )
                        }
                      </For>
                    </TableRow>
                  )}
                </For>
              </TableBody>
            </Table>
          ))
        ),
      )}
    </PageShell>
  );
}
