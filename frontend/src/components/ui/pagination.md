# Pagination（多部件）

分頁。頁碼狀態（`count`／`page`／`pageSize`）與 `api().pages` 產生的頁碼範圍一律交給 Ark `pagination`；
部件只決定外觀（Tailkit `a-c-pagination-01` 的結構 + 語意 token）。

**本階段唯一變更的對外 API**：`Pagination` 直接收 `count`／`page`／`pageSize`／`onPageChange`。

## API

### `Pagination`（Root）

| Prop | 型別 | 說明 |
| --- | --- | --- |
| `count` | `number` | 總筆數（後端 `pagination.total`），**必填** |
| `page` | `number` | 目前頁碼（1-based，受控），**必填** |
| `pageSize` | `number` | 每頁筆數，**必填** |
| `onPageChange` | `(page: number) => void` | 回報新頁碼（純值），**必填** |
| `translations` | Ark `translations` 的部分覆寫 | 未給的項目沿用內建繁中字串 |
| `class` / `children` | — | 不傳 `children` 時渲染預設版型（上一頁／頁碼範圍／下一頁） |

其餘 Ark `RootProps`（`siblingCount`、`boundaryCount`、`type`、`getPageUrl`、`onPageSizeChange`…）直通。

### 部件

| 部件 | 底層 | 說明 |
| --- | --- | --- |
| `PaginationContent` | `ul` | 相鄰項目共用邊框（`-space-x-px`），只有群組頭尾圓角 |
| `PaginationItem` | `ArkPagination.Item` | `type="page"` 由包裝層固定；選中狀態（`data-selected`、`aria-current="page"`）由 Ark 決定，**不再需要 `isActive`** |
| `PaginationEllipsis` | `ArkPagination.Ellipsis` | `index` 必須唯一（Ark 用它產生元素 id） |
| `PaginationPrevious` / `Next` | `ArkPagination.PrevTrigger` / `NextTrigger` | 文字預設「上一頁」／「下一頁」，可換 `children`；`hideText` 只留圖示 |
| `PaginationFirst` / `Last` | `ArkPagination.FirstTrigger` / `LastTrigger` | 純圖示，名稱由 Ark 的 `firstTriggerLabel`／`lastTriggerLabel` 提供 |
| `PaginationSummary` | `div` | 範圍摘要（第 X–Y 筆，共 N 筆），`font-mono` |

`PaginationItem`／`Previous`／`Next`／`First`／`Last` 都收 `variant`（`default | outline | ghost | flat`）
與 `size`（`default | sm | lg | icon`）。

## Ark 對應

`pagination`：`Root` / `Item` / `Ellipsis` / `PrevTrigger` / `NextTrigger` / `FirstTrigger` /
`LastTrigger` / `Context`（預設版型用 `Context` 取得 `api().pages`）。

## 可及性

- `<nav>` 語意與 `aria-current="page"`、頁碼鈕的無障礙名稱（`第 N 頁`）都由 Ark 產生。
- 第一頁／最後一頁時 Ark 直接標 `disabled`，不用自己算。
- 內建 `translations` 是繁中；要改字串用 `translations` prop 逐項覆寫，不要改 DOM。
- `PaginationEllipsis` 標了 `aria-hidden="true"`（裝飾性）。

## 範例

```tsx
import { Pagination } from "~/components/ui";

<Pagination count={total()} page={page()} pageSize={20} onPageChange={setPage} />;
```

列表頁另有共用列 `ListPagination`（範圍摘要 + 控制）。

## 已知限制

- `hideText` 的圖示鈕仍保留 Ark 給的無障礙名稱，所以不需要另加 `aria-label`。
- 預設版型不含 first/last 與頁碼輸入框；需要時傳 `children` 用部件自行組裝（`PaginationContent` 的
  頭尾圓角只認第一／最後一個 `<li>` 的直系子節點）。
