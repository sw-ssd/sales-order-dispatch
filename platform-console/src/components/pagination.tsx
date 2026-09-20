import type { JSX } from "solid-js";
import { Button } from "@ui/button";

/**
 * 後端分頁的控制列。
 *
 * 頁數**不由前端算資料筆數**：`total` 是後端回的總筆數（`PlatformPagination`），
 * 前端只知道現在這一頁；自己拿目前列數推算頁數會在篩選與並發寫入下說謊。
 */
export function Pagination(props: {
  page: number;
  pageSize: number;
  total: number;
  onPage: (page: number) => void;
}): JSX.Element {
  const lastPage = () => Math.max(1, Math.ceil(props.total / props.pageSize));

  return (
    <div class="flex items-center gap-3 text-sm text-muted-foreground">
      <span>
        第 {props.page} / {lastPage()} 頁（共 {props.total} 筆）
      </span>
      <Button
        variant="outline"
        size="sm"
        disabled={props.page <= 1}
        onClick={() => props.onPage(props.page - 1)}
      >
        上一頁
      </Button>
      <Button
        variant="outline"
        size="sm"
        disabled={props.page >= lastPage()}
        onClick={() => props.onPage(props.page + 1)}
      >
        下一頁
      </Button>
    </div>
  );
}
