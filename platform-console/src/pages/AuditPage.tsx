import { createQuery } from "@tanstack/solid-query";
import { For } from "solid-js";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@ui/table";
import { EmptyState, PageShell, queryBoundary } from "../components/page";
import { platform } from "../lib/api";

/** 平台稽核：誰（operator）對哪個目標做了什麼（T12 會加上目標篩選與分頁）。 */
export default function AuditPage() {
  const audit = createQuery(() => ({
    queryKey: ["platform-audit"],
    queryFn: () => platform.listPlatformAudit({ page: 1, pageSize: 50 }),
  }));

  return (
    <PageShell title="平台稽核" description="平台端寫入的稽核軌跡（operator、動作、目標、原因）。">
      {queryBoundary(audit, (data) =>
        data.entries.length === 0 ? (
          <EmptyState>尚無平台操作紀錄。</EmptyState>
        ) : (
          <Table aria-label="平台稽核紀錄">
            <TableHeader>
              <TableRow>
                <TableHead scope="col">時間</TableHead>
                <TableHead scope="col">操作者</TableHead>
                <TableHead scope="col">動作</TableHead>
                <TableHead scope="col">目標</TableHead>
                <TableHead scope="col">原因</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <For each={data.entries}>
                {(e) => (
                  <TableRow>
                    <TableCell>{e.createdAt}</TableCell>
                    <TableCell>{e.operatorEmail}</TableCell>
                    <TableCell>{e.action}</TableCell>
                    <TableCell>
                      {e.targetType}
                      {e.targetId ? `：${e.targetId}` : ""}
                    </TableCell>
                    <TableCell>{e.reason}</TableCell>
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
