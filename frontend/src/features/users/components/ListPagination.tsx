import { For, Show } from "solid-js";
import { cn } from "~/lib/cn";

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

const triggerClass =
  "rounded-md border border-gray-300 px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50";

type PageItem = { type: "page"; value: number } | { type: "ellipsis" };

// 分頁項目演算法照 @zag-js/pagination getRange 複製(siblingCount=1),
// 取代不相容 solid-js 2 的 ark Pagination,維持相同頁碼/省略號配置。
function range(start: number, end: number): number[] {
  return Array.from({ length: end - start + 1 }, (_, i) => i + start);
}

function getPageItems(page: number, totalPages: number, siblingCount = 1): PageItem[] {
  const totalPageNumbers = Math.min(2 * siblingCount + 5, totalPages);
  const firstPageIndex = 1;
  const lastPageIndex = totalPages;
  const leftSiblingIndex = Math.max(page - siblingCount, firstPageIndex);
  const rightSiblingIndex = Math.min(page + siblingCount, lastPageIndex);
  const showLeftEllipsis = leftSiblingIndex > firstPageIndex + 1;
  const showRightEllipsis = rightSiblingIndex < lastPageIndex - 1;
  const itemCount = totalPageNumbers - 2;

  let items: number[];
  if (!showLeftEllipsis && showRightEllipsis) {
    items = [...range(1, itemCount), -1, lastPageIndex];
  } else if (showLeftEllipsis && !showRightEllipsis) {
    items = [firstPageIndex, -1, ...range(lastPageIndex - itemCount + 1, lastPageIndex)];
  } else if (showLeftEllipsis && showRightEllipsis) {
    items = [firstPageIndex, -1, ...range(leftSiblingIndex, rightSiblingIndex), -1, lastPageIndex];
  } else {
    items = range(firstPageIndex, lastPageIndex);
  }
  return items.map((v) => (v === -1 ? { type: "ellipsis" } : { type: "page", value: v }));
}

/** 列表頁共用分頁列:顯示目前範圍(第 X–Y 筆,共 N 筆)與頁碼控制(原生實作,取代 ark Pagination)。 */
export function ListPagination(props: ListPaginationProps) {
  const rangeStart = () =>
    props.total === 0
      ? 0
      : Math.min((props.page - 1) * props.pageSize + 1, props.total);
  const rangeEnd = () => Math.min(props.page * props.pageSize, props.total);
  const totalPages = () => Math.max(1, Math.ceil(props.total / props.pageSize));
  const pageItems = () => getPageItems(props.page, totalPages());

  return (
    <Show when={props.total > props.pageSize}>
      <div class="flex flex-wrap items-center justify-between gap-3 border-t border-gray-200 px-4 py-3">
        <p class="text-sm text-gray-600">
          {props.total === 0
            ? "共 0 筆"
            : `第 ${rangeStart()}–${rangeEnd()} 筆,共 ${props.total} 筆`}
        </p>
        <nav aria-label="分頁" class="flex items-center gap-1">
          <button
            type="button"
            class={triggerClass}
            disabled={props.page <= 1}
            onClick={() => props.onPageChange(props.page - 1)}
          >
            上一頁
          </button>
          <For each={pageItems()}>
            {(item) =>
              item.type === "page" ? (
                <button
                  type="button"
                  aria-label={`第 ${item.value} 頁`}
                  aria-current={item.value === props.page ? "page" : undefined}
                  onClick={() => props.onPageChange(item.value)}
                  class={cn(
                    "rounded-md border px-3 py-1.5 text-sm",
                    item.value === props.page
                      ? "border-blue-600 bg-blue-600 text-white"
                      : "border-gray-300 text-gray-700 hover:bg-gray-50",
                  )}
                >
                  {item.value}
                </button>
              ) : (
                <span class="px-1 text-sm text-gray-500">…</span>
              )
            }
          </For>
          <button
            type="button"
            class={triggerClass}
            disabled={props.page >= totalPages()}
            onClick={() => props.onPageChange(props.page + 1)}
          >
            下一頁
          </button>
        </nav>
      </div>
    </Show>
  );
}
