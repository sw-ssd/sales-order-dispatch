# Spinner

載入指示器（純呈現）。

## API

| Prop | 型別 | 預設 | 說明 |
| --- | --- | --- | --- |
| `size` | `sm \| default \| lg` | `default` | `size-3` / `size-4` / `size-6` |
| `label` | `string` | `"載入中"` | 讀屏念出的文字（`aria-label`） |
| `class` | `string` | — | 追加類名；用 `text-current` 可以吃掉 `text-primary` 跟隨父層顏色 |
| 其餘 | `Omit<JSX.HTMLAttributes<HTMLSpanElement>, "role">` | — | `role` 被鎖住（固定 `status`） |

## Ark 對應

無（`ark: []`）。CSS `animate-spin` + 邊框。

## 可及性

- `role="status"` + `aria-label`：讀屏會公告「載入中」。預設 label 就是繁中 `載入中`，
  需要別的說法時再傳 `label`。
- 放在 `Button` 的 `loading` 內時不需要另行處理：按鈕會同時 `disabled` 並掛 `aria-busy="true"`。

## 範例

```tsx
import { Spinner } from "~/components/ui";

<Spinner label="載入中" />
<Spinner size="lg" label="載入中" class="text-current" />
```
