import { Show } from "solid-js";
import { Pagination, PaginationSummary } from "~/components/ui/pagination";

interface ListPaginationProps {
  /** 總筆數(後端 pagination.total) */
  total: number;
  /** 每頁筆數 */
  pageSize: number;
  /** 目前頁碼(1-based) */
  page: number;
  /** 頁碼變更(載入該頁資料) */
  onPageChange: (page: number) => void;
}

/** 列表頁共用分頁列:顯示目前範圍(第 X–Y 筆,共 N 筆)與分頁控制(頁碼範圍由 Ark 產生)。 */
export function ListPagination(props: ListPaginationProps) {
  const rangeStart = () =>
    props.total === 0
      ? 0
      : Math.min((props.page - 1) * props.pageSize + 1, props.total);
  const rangeEnd = () => Math.min(props.page * props.pageSize, props.total);

  return (
    <Show when={props.total > props.pageSize}>
      <div class="flex flex-wrap items-center justify-between gap-3 border-t border-border px-3 py-3">
        <PaginationSummary>
          {props.total === 0
            ? "共 0 筆"
            : `第 ${rangeStart()}–${rangeEnd()} 筆,共 ${props.total} 筆`}
        </PaginationSummary>
        <Pagination
          class="mx-0 w-auto justify-start"
          count={props.total}
          page={props.page}
          pageSize={props.pageSize}
          onPageChange={props.onPageChange}
        />
      </div>
    </Show>
  );
}
