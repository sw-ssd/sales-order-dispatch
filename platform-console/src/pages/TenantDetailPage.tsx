import { createQuery } from "@tanstack/solid-query";
import { useParams } from "@tanstack/solid-router";
import { For } from "solid-js";
import { Card, CardContent, CardHeader, CardTitle } from "@ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@ui/table";
import { EmptyState, PageShell, queryBoundary } from "../components/page";
import { SubscriptionBadge } from "../components/status";
import { platform } from "../lib/api";

/** 租戶詳情：概況 ＋ 平台對該租戶的功能覆寫（override）。 */
export default function TenantDetailPage() {
  const params = useParams({ from: "/tenants/$tenantId" });
  const tenant = createQuery(() => ({
    queryKey: ["tenant", params().tenantId],
    queryFn: () => platform.getTenant({ companyId: params().tenantId }),
  }));

  return (
    <PageShell title="租戶詳情" description={`公司識別碼 ${params().tenantId}`}>
      {queryBoundary(tenant, (data) => (
        <div class="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>{data.tenant?.companyName ?? params().tenantId}</CardTitle>
            </CardHeader>
            <CardContent class="space-y-2 text-sm">
              <p>方案：{data.tenant?.planName || data.tenant?.planCode}</p>
              <p>
                狀態：<SubscriptionBadge status={data.tenant?.subscriptionStatus ?? "none"} />
              </p>
              <p>席次：{data.tenant?.seatCount}</p>
              <p>本期結束：{data.tenant?.currentPeriodEnd}</p>
            </CardContent>
          </Card>

          <section class="space-y-2">
            <h2 class="text-lg font-semibold">功能覆寫</h2>
            {data.overrides.length === 0 ? (
              <EmptyState>沒有覆寫：此租戶一律依方案權益。</EmptyState>
            ) : (
              <Table aria-label="功能覆寫">
                <TableHeader>
                  <TableRow>
                    <TableHead scope="col">功能</TableHead>
                    <TableHead scope="col">啟用</TableHead>
                    <TableHead scope="col">上限</TableHead>
                    <TableHead scope="col">原因</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <For each={data.overrides}>
                    {(o) => (
                      <TableRow>
                        <TableCell>{o.featureCode}</TableCell>
                        <TableCell>
                          {o.enabledSet ? (o.enabled ? "是" : "否") : "未指定"}
                        </TableCell>
                        <TableCell>{o.limitSet ? String(o.limitValue) : "—"}</TableCell>
                        <TableCell>{o.reason}</TableCell>
                      </TableRow>
                    )}
                  </For>
                </TableBody>
              </Table>
            )}
          </section>
        </div>
      ))}
    </PageShell>
  );
}
