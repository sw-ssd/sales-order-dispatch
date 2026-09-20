import { createQuery } from "@tanstack/solid-query";
import { For } from "solid-js";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@ui/table";
import { EmptyState, PageShell, queryBoundary } from "../components/page";
import { platform } from "../lib/api";

/** 待收款：未付期別（金額為後端 money.ParseCents 的十進位字串；收款動作在 T12）。 */
export default function ReceivablesPage() {
  const receivables = createQuery(() => ({
    queryKey: ["receivables"],
    queryFn: () => platform.listReceivables({ page: 1, pageSize: 50 }),
  }));

  return (
    <PageShell title="待收款" description="所有未付期別（平台自營公司不算租戶，不列入）。">
      {queryBoundary(receivables, (data) =>
        data.rows.length === 0 ? (
          <EmptyState>沒有未付期別。</EmptyState>
        ) : (
          <Table aria-label="待收款清單">
            <TableHeader>
              <TableRow>
                <TableHead scope="col">公司</TableHead>
                <TableHead scope="col">方案</TableHead>
                <TableHead scope="col">期別</TableHead>
                <TableHead scope="col">金額</TableHead>
                <TableHead scope="col">期末</TableHead>
                <TableHead scope="col">狀態</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <For each={data.rows}>
                {(r) => (
                  <TableRow>
                    <TableCell>{r.companyName}</TableCell>
                    <TableCell>{r.planCode}</TableCell>
                    <TableCell>第 {r.periodNo} 期</TableCell>
                    <TableCell>{r.amount}</TableCell>
                    <TableCell>{r.periodEnd}</TableCell>
                    <TableCell>{r.status === "open" ? "未付" : r.status}</TableCell>
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
