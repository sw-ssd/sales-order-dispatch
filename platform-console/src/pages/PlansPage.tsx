import { createQuery } from "@tanstack/solid-query";
import { For } from "solid-js";
import { Badge } from "@ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@ui/table";
import { EmptyState, PageShell, queryBoundary } from "../components/page";
import { platform } from "../lib/api";

/** 方案與價目：唯讀檢視（價目維護在 T12 的寫入流程）。 */
export default function PlansPage() {
  const plans = createQuery(() => ({
    queryKey: ["plans"],
    queryFn: () => platform.listPlans({}),
  }));

  return (
    <PageShell title="方案與價目" description="方案定義與各計費週期價格（金額為後端 money.ParseCents 的十進位字串）。">
      {queryBoundary(plans, (data) =>
        data.plans.length === 0 ? (
          <EmptyState>尚未定義任何方案。</EmptyState>
        ) : (
          <div class="space-y-4">
            <For each={data.plans}>
              {(plan) => (
                <section class="space-y-2">
                  <h2 class="flex items-center gap-2 text-lg font-semibold">
                    {plan.name}
                    <Badge variant={plan.status === "active" ? "success" : "secondary"}>
                      {plan.status === "active" ? "上架中" : plan.status}
                    </Badge>
                  </h2>
                  {plan.prices.length === 0 ? (
                    <EmptyState>尚未設定價目。</EmptyState>
                  ) : (
                    <Table aria-label={`${plan.name} 價目`}>
                      <TableHeader>
                        <TableRow>
                          <TableHead scope="col">計費週期</TableHead>
                          <TableHead scope="col">基本價</TableHead>
                          <TableHead scope="col">單席價</TableHead>
                          <TableHead scope="col">幣別</TableHead>
                          <TableHead scope="col">生效日</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        <For each={plan.prices}>
                          {(price) => (
                            <TableRow>
                              <TableCell>{price.billingCycle}</TableCell>
                              <TableCell>{price.basePrice}</TableCell>
                              <TableCell>{price.seatPrice}</TableCell>
                              <TableCell>{price.currency || "TWD"}</TableCell>
                              <TableCell>{price.effectiveFrom}</TableCell>
                            </TableRow>
                          )}
                        </For>
                      </TableBody>
                    </Table>
                  )}
                </section>
              )}
            </For>
          </div>
        ),
      )}
    </PageShell>
  );
}
