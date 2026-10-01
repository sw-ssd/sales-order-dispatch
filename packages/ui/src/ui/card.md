# Card（多部件）

卡片容器。結構取自 Tailkit（`a-c-cards-01/03/09`：`flex-col`、`overflow-hidden`、`rounded-lg`、
`shadow-xs`，標題／頁尾為 `bg-muted` 色帶）。

## API

所有部件都只額外收 `class`，其餘屬性直通該部件渲染的元素。

| 部件 | 元素 | 預設樣式 |
| --- | --- | --- |
| `Card` | `div` | `flex flex-col overflow-hidden rounded-lg border border-border bg-card text-card-foreground shadow-xs` |
| `CardHeader` | `div` | `flex flex-col gap-1.5 bg-muted px-5 py-4` |
| `CardTitle` | `h3` | `font-semibold leading-none tracking-tight text-lg` |
| `CardDescription` | `p` | `text-sm text-muted-foreground` |
| `CardContent` | `div` | `grow p-5` |
| `CardFooter` | `div` | `flex items-center gap-2 bg-muted px-5 py-4 text-sm text-muted-foreground` |

## Ark 對應

無（`ark: []`）。純版面容器。

## 可及性

- `CardTitle` 固定渲染 `<h3>`：放進頁面時要確認標題層級（頁面若已有 `h2`，卡片用 `h3` 剛好；
  若卡片直接掛在 `h1` 底下就會跳級）——需要別層級時傳 `class` 不夠，請用原生標題元素自行組版，
  或把 `CardTitle` 放在正確的層級位置。
- 卡片本身沒有 `role`；若整張卡可點，請放一個真正的 `<a>`／`<button>` 在裡面，
  不要只掛 `onClick` 到 `div`。

## 範例

```tsx
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle, Button } from "~/components/ui";

<Card>
  <CardHeader>
    <CardTitle>客戶資料</CardTitle>
    <CardDescription>共 42 筆</CardDescription>
  </CardHeader>
  <CardContent>…</CardContent>
  <CardFooter>
    <Button size="sm">匯出</Button>
  </CardFooter>
</Card>
```
