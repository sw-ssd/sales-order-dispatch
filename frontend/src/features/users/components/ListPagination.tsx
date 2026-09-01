import { For, Show } from "solid-js";
import { createPagination } from "~/hooks/create-pagination";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
} from "~/components/ui/pagination";

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

/** 列表頁共用分頁列:顯示目前範圍(第 X–Y 筆,共 N 筆)與分頁控制(含省略號)。 */
export function ListPagination(props: ListPaginationProps) {
  const pagination = createPagination({
    count: () => props.total,
    pageSize: () => props.pageSize,
    page: () => props.page,
    onChange: (p) => props.onPageChange(p),
  });

  const rangeStart = () =>
    props.total === 0
      ? 0
      : Math.min((props.page - 1) * props.pageSize + 1, props.total);
  const rangeEnd = () => Math.min(props.page * props.pageSize, props.total);

  return (
    <Show when={props.total > props.pageSize}>
      <div class="flex flex-wrap items-center justify-between gap-3 border-t border-gray-200 px-4 py-3">
        <p class="text-sm text-gray-600">
          {props.total === 0
            ? "共 0 筆"
            : `第 ${rangeStart()}–${rangeEnd()} 筆,共 ${props.total} 筆`}
        </p>
        <Pagination class="mx-0 w-auto justify-start">
          <PaginationContent>
            <PaginationItem>
              <PaginationLink
                aria-label="上一頁"
                disabled={!pagination.hasPrevious()}
                onClick={pagination.previous}
              >
                上一頁
              </PaginationLink>
            </PaginationItem>
            <For each={pagination.range()}>
              {(item) => (
                <PaginationItem>
                  {item === "ellipsis" ? (
                    <PaginationEllipsis />
                  ) : (
                    <PaginationLink
                      isActive={item === props.page}
                      aria-label={`第 ${item} 頁`}
                      onClick={() =>
                        item !== props.page && pagination.setPage(item)
                      }
                    >
                      {item}
                    </PaginationLink>
                  )}
                </PaginationItem>
              )}
            </For>
            <PaginationItem>
              <PaginationLink
                aria-label="下一頁"
                disabled={!pagination.hasNext()}
                onClick={pagination.next}
              >
                下一頁
              </PaginationLink>
            </PaginationItem>
          </PaginationContent>
        </Pagination>
      </div>
    </Show>
  );
}
