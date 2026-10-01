# Table（多部件）

表格。結構取自 Tailkit（`a-c-tables-01` Bordered、`a-c-tables-13` In Card Alternate）：表頭色帶、
較緊湊的 cell 內距；外框與圓角留給外層容器（`Card`）。

## API

所有部件都只額外收 `class`，其餘屬性直通該部件渲染的元素。

| 部件 | 元素 | 預設樣式 |
| --- | --- | --- |
| `Table` | `div` + `table` | 外層 `div` 負責 `overflow-x-auto`；`table` 為 `min-w-full text-sm whitespace-nowrap` |
| `TableHeader` | `thead` | `[&_tr]:border-b [&_tr]:border-border` |
| `TableBody` | `tbody` | `[&_tr:last-child]:border-0` |
| `TableFooter` | `tfoot` | `border-t border-border bg-muted font-medium` |
| `TableRow` | `tr` | `border-b` + `hover:bg-muted/50` + `data-[state=selected]:bg-muted` |
| `TableHead` | `th` | `bg-muted px-3 py-4 text-left font-semibold text-foreground` |
| `TableCell` | `td` | `p-3 align-middle` |
| `TableCaption` | `caption` | `mt-4 text-xs text-muted-foreground pb-2` |

`Table` 渲染的是 `div > table`（不是單一 `<table>`），class 下在 `<table>` 上。

## Ark 對應

無（`ark: []`）。Ark 沒有 table primitive；排序／選取等行為由 Phase 2 的 TanStack Table 接手。

## 可及性

- 用 `TableCaption` 或 `aria-label` 給表格名稱；資料表請用 `TableHead` 配 `scope="col"`（需要時 `scope="row"`）。
- 表頭色帶是純視覺，`th` 語意本身才是重點。
- 水平溢出由外層 `div` 的 `overflow-x-auto` 處理（`whitespace-nowrap` 讓內容不換行）；
  容器可聚焦與否由呼叫端決定（鍵盤使用者若無法捲動，請給容器 `tabindex="0"`）。
- `data-[state=selected]` 與 `[role=checkbox]` 的 micro-layout 已內建，配合 `Checkbox` 使用時不用另調。

## 範例

```tsx
import { Table, TableBody, TableCaption, TableCell, TableHead, TableHeader, TableRow } from "~/components/ui";

<Table>
  <TableCaption>客戶清單</TableCaption>
  <TableHeader>
    <TableRow>
      <TableHead scope="col">名稱</TableHead>
      <TableHead scope="col">統一編號</TableHead>
    </TableRow>
  </TableHeader>
  <TableBody>
    <TableRow>
      <TableCell>範例公司</TableCell>
      <TableCell>12345678</TableCell>
    </TableRow>
  </TableBody>
</Table>;
```
