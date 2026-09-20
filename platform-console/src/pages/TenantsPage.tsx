import { createQuery } from "@tanstack/solid-query";
import { Link } from "@tanstack/solid-router";
import { createSignal, For, Show } from "solid-js";
import { Badge } from "@ui/badge";
import { Button } from "@ui/button";
import { Field, FieldLabel } from "@ui/field";
import { Input } from "@ui/input";
import { Pagination, PaginationSummary } from "@ui/pagination";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@ui/table";
import { EmptyState, PageShell, queryBoundary } from "../components/page";
import { Select } from "../components/select";
import { SUBSCRIPTION_STATUSES, SubscriptionBadge } from "../components/status";
import { platform } from "../lib/api";

/** 每頁筆數（與後端 `defaultPageSize` 一致）。 */
const PAGE_SIZE = 20;

/**
 * 租戶列表：跨租戶概況。
 *
 * 篩選與分頁**一律由後端執行**（`ListTenants` 的 keyword／status／page）：console 不做前端過濾，
 * 否則「這一頁沒有」與「全部沒有」分不出來，頁數也算不出來。草稿（輸入框）與已套用的篩選分開，
 * 按下「查詢」才打後端並回到第 1 頁。
 *
 * 三態（載入／空／錯誤）由 `queryBoundary` 統一呈現——錯誤經 `describeError`，
 * SYS-4001 會帶可行動說明（見 lib/errors.ts）。
 */
export default function TenantsPage() {
  const [page, setPage] = createSignal(1);
  const [keywordDraft, setKeywordDraft] = createSignal("");
  const [statusDraft, setStatusDraft] = createSignal("");
  const [filter, setFilter] = createSignal({ keyword: "", status: "" });

  const tenants = createQuery(() => ({
    queryKey: ["tenants", page(), filter().keyword, filter().status],
    queryFn: () =>
      platform.listTenants({
        page: page(),
        pageSize: PAGE_SIZE,
        keyword: filter().keyword,
        status: filter().status,
      }),
  }));

  return (
    <PageShell
      title="租戶"
      description="跨租戶概況：方案、席位、訂閱狀態（僅 operator 可讀；授權由後端判定）。"
    >
      <form
        class="flex flex-wrap items-end gap-3"
        onSubmit={(e) => {
          e.preventDefault();
          setPage(1);
          setFilter({ keyword: keywordDraft(), status: statusDraft() });
        }}
      >
        <Field class="w-full sm:w-64">
          <FieldLabel for="tenant-keyword">公司關鍵字</FieldLabel>
          <Input
            id="tenant-keyword"
            value={keywordDraft()}
            placeholder="公司名稱或識別碼"
            onInput={(e) => setKeywordDraft(e.currentTarget.value)}
          />
        </Field>
        <Field class="w-full sm:w-48">
          <FieldLabel for="tenant-status">訂閱狀態</FieldLabel>
          <Select
            id="tenant-status"
            value={statusDraft()}
            onChange={(e) => setStatusDraft(e.currentTarget.value)}
          >
            <option value="">全部</option>
            <For each={SUBSCRIPTION_STATUSES}>
              {(it) => <option value={it.value}>{it.label}</option>}
            </For>
          </Select>
        </Field>
        <Button type="submit" size="sm">
          查詢
        </Button>
      </form>

      {queryBoundary(tenants, (data) => {
        const total = data.pagination?.total ?? 0;

        if (data.tenants.length === 0) {
          return (
            <EmptyState>
              {filter().keyword || filter().status ? "沒有符合條件的租戶。" : "目前沒有租戶。"}
            </EmptyState>
          );
        }

        return (
          <div class="space-y-3">
            <Table aria-label="租戶清單">
              <TableHeader>
                <TableRow>
                  <TableHead scope="col">公司</TableHead>
                  <TableHead scope="col">方案</TableHead>
                  <TableHead scope="col">訂閱狀態</TableHead>
                  <TableHead scope="col">席位</TableHead>
                  <TableHead scope="col">當期到期日</TableHead>
                  <TableHead scope="col">待收款</TableHead>
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
                          class="font-medium underline underline-offset-4"
                        >
                          {t.companyName}
                        </Link>
                      </TableCell>
                      <TableCell>{t.planName || t.planCode}</TableCell>
                      <TableCell>
                        <SubscriptionBadge status={t.subscriptionStatus} />
                      </TableCell>
                      <TableCell>{t.seatCount}</TableCell>
                      <TableCell>{t.currentPeriodEnd || "—"}</TableCell>
                      <TableCell>
                        {t.overdue ? <Badge variant="destructive">逾期</Badge> : "—"}
                      </TableCell>
                    </TableRow>
                  )}
                </For>
              </TableBody>
            </Table>

            <Show when={total > PAGE_SIZE}>
              <div class="flex flex-wrap items-center justify-between gap-3">
                <PaginationSummary>
                  {/* 範圍照「要求的視窗」算（與後端分頁一致）：最後一頁不超過總筆數。 */}
                  {`第 ${total === 0 ? 0 : Math.min((page() - 1) * PAGE_SIZE + 1, total)}–${Math.min(
                    page() * PAGE_SIZE,
                    total,
                  )} 筆,共 ${total} 筆`}
                </PaginationSummary>
                <Pagination
                  class="mx-0 w-auto justify-start"
                  count={total}
                  page={page()}
                  pageSize={PAGE_SIZE}
                  onPageChange={setPage}
                />
              </div>
            </Show>
          </div>
        );
      })}
    </PageShell>
  );
}
