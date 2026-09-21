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
  // 未結項 #31:當頁被清空時（例：最後一頁的唯一一列被收款），後端回的 page 仍是舊頁碼，
  // 而 total 已縮水 → 顯示會變成「第 2 / 1 頁」的矛盾頁碼。顯示值夾住（不改呼叫端的 signal）：
  // 按鈕的 enabled 仍以真實 page 為準（「上一頁」可按），只有「第 N 頁」這個文字不說謊。
  const shownPage = () => Math.min(props.page, lastPage());

  return (
    <div class="flex items-center gap-3 text-sm text-muted-foreground">
      <span>
        第 {shownPage()} / {lastPage()} 頁（共 {props.total} 筆）
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
