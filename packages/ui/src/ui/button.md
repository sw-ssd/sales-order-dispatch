# Button

按鈕。行為沿用原生 `<button>`（沒有 Ark primitive），視覺照 Tailkit（`a-c-buttons-01/02/03/06/08`）
改寫為語意 token。

## API

| Prop | 型別 | 預設 | 說明 |
| --- | --- | --- | --- |
| `variant` | `default \| destructive \| outline \| secondary \| ghost \| link \| success \| warning \| info` | `default` | 顏色語意 |
| `size` | `xs \| sm \| default \| lg \| icon` | `default` | `icon` 是 `size-9` 的方形鈕 |
| `loading` | `boolean` | `false` | 顯示 `Spinner`（`size="sm"`）、同時 `disabled`，並掛 `aria-busy="true"` |
| `disabled` | `boolean` | — | 直通原生 `disabled` |
| `class` | `string` | — | 追加類名 |
| 其餘 | `JSX.ButtonHTMLAttributes<HTMLButtonElement>` | — | 直通 `<button>` |

`default` 的 hover 用 `bg-primary/95`（不是 `/90`）：白字疊在 90% 主色上只有 4.36:1，未達 AA。
`[&_svg]:size-4`：放進按鈕的 lucide 圖示自動統一尺寸。

## Ark 對應

無（`ark: []`）。要「按下切換」的語意請用 Ark `toggle-group`（見 `theme.md` 的 `ThemeSwitcher`），
不要用 `Button` 自己存布林值假裝分段控制。

## 可及性

- 只有圖示時**必須**給 `aria-label`（`size="icon"` 的常見誤用）。
- `loading` 會同時 `disabled`，避免重複送出；`aria-busy` 讓讀屏知道正在處理。
- `link` 變體看起來像連結但仍是 `<button>`；真正的導覽請用 `<a>`（或路由的 `Link`）配
  `buttonVariants()`，不要用 `onClick` 假導覽。
- 焦點環 `focus-visible:ring-3 focus-visible:ring-ring`，disabled 時 `pointer-events-none` + `opacity-50`。

## 範例

```tsx
import { Button } from "~/components/ui";

<Button>儲存</Button>
<Button variant="outline" size="sm">取消</Button>
<Button variant="destructive" loading={submitting()}>刪除</Button>
<Button size="icon" aria-label="切換側邊欄"><PanelLeft /></Button>
```
