import { createQuery } from "@tanstack/solid-query";
import { For } from "solid-js";
import { Badge } from "@ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@ui/table";
import { EmptyState, PageShell, queryBoundary } from "../components/page";
import { platform } from "../lib/api";

/**
 * 方案權益矩陣：完整功能清單 × 該方案的權益（未設定的格子也要顯示）。
 * 目前取第一個方案（T12 會加上方案切換）。
 */
export default function EntitlementsPage() {
  const plans = createQuery(() => ({ queryKey: ["plans"], queryFn: () => platform.listPlans({}) }));
  const planCode = () => plans.data?.plans[0]?.code ?? "";

  const entitlements = createQuery(() => ({
    queryKey: ["entitlements", planCode()],
    queryFn: () => platform.getPlanEntitlements({ planCode: planCode() }),
    enabled: planCode() !== "",
  }));

  return (
    <PageShell
      title="方案權益"
      description={planCode() ? `方案 ${planCode()} 的功能與上限（未設定的格子以「未設定」表示）。` : "功能與上限。"}
    >
      {queryBoundary(plans, (data) =>
        data.plans.length === 0 ? (
          <EmptyState>尚未定義任何方案。</EmptyState>
        ) : (
          queryBoundary(entitlements, (matrix) => (
            <Table aria-label="方案權益矩陣">
              <TableHeader>
                <TableRow>
                  <TableHead scope="col">功能</TableHead>
                  <TableHead scope="col">類型</TableHead>
                  <TableHead scope="col">啟用</TableHead>
                  <TableHead scope="col">上限</TableHead>
                  <TableHead scope="col">說明</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <For each={matrix.features}>
                  {(f) => {
                    const row = matrix.entitlements.find((e) => e.featureCode === f.code);
                    return (
                      <TableRow>
                        <TableCell>{f.code}</TableCell>
                        <TableCell>{f.type === "integer" ? "數量" : "開關"}</TableCell>
                        <TableCell>
                          <Badge variant={row?.enabled ? "success" : "secondary"}>
                            {row?.enabled ? "啟用" : "停用"}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {row?.limitSet ? String(row.limitValue) : "未設定"}
                        </TableCell>
                        <TableCell>{f.description}</TableCell>
                      </TableRow>
                    );
                  }}
                </For>
              </TableBody>
            </Table>
          ))
        ),
      )}
    </PageShell>
  );
}
