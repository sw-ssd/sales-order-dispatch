import { Pagination } from "@ark-ui/solid/pagination";
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

/** 列表頁共用分頁列:顯示目前範圍(第 X–Y 筆,共 N 筆)與 Ark Pagination 控制。 */
export function ListPagination(props: ListPaginationProps) {
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
        <Pagination.Root
          count={props.total}
          pageSize={props.pageSize}
          page={props.page}
          onPageChange={(details) => props.onPageChange(details.page)}
          class="flex items-center gap-1"
        >
          <Pagination.PrevTrigger class={triggerClass}>
            上一頁
          </Pagination.PrevTrigger>
          <Pagination.Context>
            {(pagination) => (
              <For each={pagination().pages}>
                {(item, index) =>
                  item.type === "page" ? (
                    <Pagination.Item
                      {...item}
                      aria-label={`第 ${item.value} 頁`}
                      class={cn(
                        "rounded-md border px-3 py-1.5 text-sm",
                        item.value === pagination().page
                          ? "border-blue-600 bg-blue-600 text-white"
                          : "border-gray-300 text-gray-700 hover:bg-gray-50",
                      )}
                    >
                      {item.value}
                    </Pagination.Item>
                  ) : (
                    <Pagination.Ellipsis
                      index={index()}
                      class="px-1 text-sm text-gray-500"
                    >
                      …
                    </Pagination.Ellipsis>
                  )
                }
              </For>
            )}
          </Pagination.Context>
          <Pagination.NextTrigger class={triggerClass}>
            下一頁
          </Pagination.NextTrigger>
        </Pagination.Root>
      </div>
    </Show>
  );
}
