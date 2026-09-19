# Badge

靜態狀態標籤（純呈現、無互動）。視覺結構取自 Tailkit（`a-c-badges-01/03`：`rounded-sm`、`px-2`、`py-1`、
`text-xs`、無邊框軟底），顏色一律語意 token。

## API

| Prop | 型別 | 預設 | 說明 |
| --- | --- | --- | --- |
| `variant` | `default \| secondary \| destructive \| outline \| success \| warning \| info` | `default` | 顏色語意 |
| `class` | `string` | — | 追加類名（走 `twMerge`） |
| `children` | `JSX.Element` | — | 標籤文字 |
| 其餘 | `JSX.HTMLAttributes<HTMLDivElement>` | — | 直通根層 `<div>` |

`variant` 對應 token：`default` → `--primary`／`--primary-foreground`；`secondary` → `--muted`／`--muted-foreground`；
`destructive`／`success`／`warning`／`info` → 各自的 `/15` 軟底 + 同色文字；`outline` → `--border` + `--foreground`。

## Ark 對應

無。Ark 沒有等價 primitive；本元件是純樣式 `<div>`（`ark: []`）。

## 可及性

- 內容是純文字，讀屏會直接念出；**不要**只靠顏色傳達狀態，文字本身要寫清楚（例：「已取消」而非「—」）。
- 有 `focus:ring-*` 樣式但 `<div>` 不可聚焦，因此沒有焦點環；要能聚焦請用 `tabindex="0"` 或改用 `Button`。
- 顏色對比：軟底變體（`bg-*/15`）的文字色與底色的對比已由 token 決定；深色模式交由 token 翻轉，
  沒有任何顏色相關的 `dark:` 變體。

## 範例

```tsx
import { Badge } from "~/components/ui";

<Badge>預設</Badge>
<Badge variant="success">已完成</Badge>
<Badge variant="destructive">已取消</Badge>
<Badge variant="outline">草稿</Badge>
```

狀態清單請用 `<ul>`／`<li>` 包起來，不要用一排 `<div>` 假裝清單。
