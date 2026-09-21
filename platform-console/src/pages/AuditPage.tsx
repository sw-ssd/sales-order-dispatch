import { createQuery } from "@tanstack/solid-query";
import { createSignal, For, Show } from "solid-js";
import { Button } from "@ui/button";
import { Field, FieldDescription, FieldLabel } from "@ui/field";
import { Input } from "@ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@ui/table";
import { Pagination } from "../components/pagination";
import { EmptyState, PageShell, queryBoundary } from "../components/page";
import { platform } from "../lib/api";

/**
 * 平台稽核：平台端每一個寫入的軌跡（誰、對誰、做了什麼、為什麼）。
 *
 * **篩選與分頁都在後端**（`ListPlatformAudit` 收 target_type／target_id＋page／page_size）：
 * 稽核是只會長大的表，抓到前端再過濾等於一開始就註定漏掉舊資料。
 *
 * 每列都要看得見 `reason`：平台寫入的稽核是查帳的唯一依據，少了它只剩「有人改過」。
 * 篩選條件是**按下去才生效**（表單送出），不是每個字都打一次請求。
 */
const PAGE_SIZE = 20;

/**
 * `target_type` 在後端是**自由文字且精確比對**（`ListPlatformAudit` 只 TrimSpace），
 * 這些是實際上會出現的值（各寫入路徑的 `writeTx` 呼叫點 ＋ billing 的 subscription）。
 * 用 `datalist` 而不是 `select`：清單只是防打錯字，不是白名單 —— 後端加了新種類仍可篩。
 */
const TARGET_TYPES = ["settings", "company", "plan", "operator", "subscription"];

/**
 * 未結項 #4：同 trace_id 的列為同一請求所寫 —— 相鄰且 trace 相同的列併為一組，
 * 續列標「同上筆請求」。trace 為空（排程／歷史列）不成組：各列獨立顯示。
 */
export function groupedEntries<T extends { traceId: string }>(entries: readonly T[]) {
  const groups: { traceId: string; entries: T[] }[] = [];
  for (const entry of entries) {
    const last = groups[groups.length - 1];
    if (entry.traceId !== "" && last !== undefined && last.traceId === entry.traceId) {
      last.entries.push(entry);
    } else {
      groups.push({ traceId: entry.traceId, entries: [entry] });
    }
  }
  return groups;
}

export default function AuditPage() {
  // 送出中的草稿與「已套用」的查詢條件分開：打字的每個字不該各打一次後端。
  const [targetType, setTargetType] = createSignal("");
  const [targetId, setTargetId] = createSignal("");
  // 頁碼與篩選放**同一個 signal**：一次操作就是一次查詢條件變更（分開寫會先送出一個中間狀態，
  // 例：按「清除」時先送 page=1＋舊篩選，再送 page=1＋無篩選 —— 兩趟白工的請求）。
  const [query, setQuery] = createSignal({ page: 1, targetType: "", targetId: "" });

  const audit = createQuery(() => ({
    queryKey: ["platform-audit", query()],
    queryFn: () =>
      platform.listPlatformAudit({
        page: query().page,
        pageSize: PAGE_SIZE,
        targetType: query().targetType,
        targetId: query().targetId,
      }),
  }));

  const applyFilter = () => {
    setQuery({ page: 1, targetType: targetType().trim(), targetId: targetId().trim() });
  };

  /** 已套用的篩選條件（給空狀態用：打錯字不能長得像「平台從來沒被寫入過」）。 */
  const appliedFilter = () => {
    const parts: string[] = [];
    if (query().targetType) parts.push(`目標類型「${query().targetType}」`);
    if (query().targetId) parts.push(`目標代碼「${query().targetId}」`);
    return parts.join("、");
  };

  return (
    <PageShell
      title="平台稽核"
      description="平台端寫入的稽核軌跡（operator、動作、目標、原因、時間），新到舊；篩選與分頁皆由後端執行。"
    >
      <form
        class="grid gap-3 sm:grid-cols-[1fr_1fr_auto]"
        aria-label="稽核篩選"
        onSubmit={(e) => {
          e.preventDefault();
          applyFilter();
        }}
      >
        <Field>
          <FieldLabel for="audit-target-type">目標類型</FieldLabel>
          <Input
            id="audit-target-type"
            list="audit-target-types"
            value={targetType()}
            placeholder="例：company、subscription、plan、operator"
            onInput={(e) => setTargetType(e.currentTarget.value)}
          />
          <datalist id="audit-target-types">
            <For each={TARGET_TYPES}>{(type) => <option value={type} />}</For>
          </datalist>
          <FieldDescription>精確比對目標種類；留空＝不限。</FieldDescription>
        </Field>

        <Field>
          <FieldLabel for="audit-target-id">目標代碼</FieldLabel>
          <Input
            id="audit-target-id"
            value={targetId()}
            placeholder="例：公司 ID、期別所屬訂閱 ID"
            onInput={(e) => setTargetId(e.currentTarget.value)}
          />
        </Field>

        <div class="flex items-end gap-2">
          <Button type="submit" size="sm">
            查詢
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => {
              setTargetType("");
              setTargetId("");
              setQuery({ page: 1, targetType: "", targetId: "" });
            }}
          >
            清除
          </Button>
        </div>
      </form>

      {queryBoundary(audit, (data) => (
        <div class="space-y-4">
          <Show
            when={data.entries.length > 0}
            fallback={
              <EmptyState>
                {appliedFilter()
                  ? `沒有符合條件的紀錄（${appliedFilter()}）。`
                  : "尚無平台操作紀錄。"}
              </EmptyState>
            }
          >
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
                <For each={groupedEntries(data.entries)}>
                  {(group) => (
                    <>
                      <For each={group.entries}>
                        {(entry, i) => (
                          <TableRow>
                            <TableCell>{entry.createdAt}</TableCell>
                            <TableCell>{entry.operatorEmail}</TableCell>
                            <TableCell>{entry.action}</TableCell>
                            <TableCell>
                              {entry.targetType}
                              {entry.targetId ? `：${entry.targetId}` : ""}
                            </TableCell>
                            <TableCell>
                              {entry.reason}
                              {/* 未結項 #4：同請求多列共用 trace_id —— 續列標「同上筆請求」，否則一次請求寫 2 列看起來像兩次操作。 */}
                              <Show when={i() > 0 && group.traceId !== ""}>
                                <span class="text-muted-foreground">（同上筆請求）</span>
                              </Show>
                            </TableCell>
                          </TableRow>
                        )}
                      </For>
                    </>
                  )}
                </For>
              </TableBody>
            </Table>

          </Show>

          {/* 分頁控制**永遠**在（含篩選無結果時）：關在「有資料」分支裡，畫面就只剩一句
              「沒有紀錄」，看的人分不出是篩選問題還是真的沒有。 */}
          <Pagination
            page={query().page}
            pageSize={data.pagination?.pageSize ?? PAGE_SIZE}
            total={data.pagination?.total ?? data.entries.length}
            onPage={(page) => setQuery({ ...query(), page })}
          />
        </div>
      ))}
    </PageShell>
  );
}
