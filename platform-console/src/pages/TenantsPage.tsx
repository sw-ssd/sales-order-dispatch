import { createQuery } from "@tanstack/solid-query";
import { Link } from "@tanstack/solid-router";
import { For } from "solid-js";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@ui/table";
import { EmptyState, PageShell, queryBoundary } from "../components/page";
import { SubscriptionBadge } from "../components/status";
import { platform } from "../lib/api";

/** 租戶列表：跨租戶概況（T11 會把篩選／分頁補上；本頁目前固定取第一頁）。 */
export default function TenantsPage() {
  const tenants = createQuery(() => ({
    queryKey: ["tenants"],
    queryFn: () => platform.listTenants({ page: 1, pageSize: 50 }),
  }));

  return (
    <PageShell title="租戶" description="跨租戶概況：方案、席次、訂閱狀態（僅 operator 可讀）。">
      {queryBoundary(tenants, (data) =>
        data.tenants.length === 0 ? (
          <EmptyState>沒有任何租戶。</EmptyState>
        ) : (
          <Table aria-label="租戶清單">
            <TableHeader>
              <TableRow>
                <TableHead scope="col">公司</TableHead>
                <TableHead scope="col">方案</TableHead>
                <TableHead scope="col">狀態</TableHead>
                <TableHead scope="col">席次</TableHead>
                <TableHead scope="col">本期結束</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <For each={data.tenants}>
                {(t) => (
                  <TableRow>
                    <TableCell>
                      <Link
                        to="/tenants/$tenantId"
                        params={{ tenantId: t.companyId }}
                        class="underline underline-offset-4"
                      >
                        {t.companyName}
                      </Link>
                      {t.overdue ? <span class="ml-2 text-xs text-destructive">有未付期別</span> : null}
                    </TableCell>
                    <TableCell>{t.planName || t.planCode}</TableCell>
                    <TableCell>
                      <SubscriptionBadge status={t.subscriptionStatus} />
                    </TableCell>
                    <TableCell>{t.seatCount}</TableCell>
                    <TableCell>{t.currentPeriodEnd}</TableCell>
                  </TableRow>
                )}
              </For>
            </TableBody>
          </Table>
        ),
      )}
    </PageShell>
  );
}
