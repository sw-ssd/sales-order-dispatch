# Input

文字輸入框。視覺照 Tailkit（`a-c-form-elements-01/16`、`a-c-form-layouts-02`）改寫為語意 token，
並收斂 `@tailwindcss/forms` 對原生 `input` 的三項副作用（外掛寫在 base layer，本元件的 utilities 一律覆蓋）：

1. 外掛硬編碼 `background-color: #fff` → 以 `bg-card` 覆蓋，深色模式不再白底。
2. 外掛在 `:focus` 加 1px 藍框 → 以 `focus-visible:border-primary` + `ring-3 ring-primary/50` 覆蓋。
3. 外掛固定 placeholder 灰階 → 以 `placeholder:text-muted-foreground` 覆蓋。

## API

| Prop | 型別 | 預設 | 說明 |
| --- | --- | --- | --- |
| `type` | `string` | `"text"` | 直通原生 `type` |
| `class` | `string` | — | 追加類名 |
| 其餘 | `JSX.InputHTMLAttributes<HTMLInputElement>` | — | 直通 `<input>`（`value`、`onInput`、`required`、`disabled`…） |

## Ark 對應

`ark: []`——**它沒有包裝任何 Ark primitive**（是原生 `<input>`）。但它 import 了
`@ark-ui/solid/field` 的 `useFieldContext`：放在 `Field` 內時，會自行補上等同 Ark `Field.Input`
`getControlProps()` 的 `aria-invalid`／`aria-describedby`。因此 `deps` 有 `@ark-ui/solid` 而 `ark` 為空。
這條規則的實際好處：`Field` 的值錯誤態與說明文字只要寫在 `<Field>` 內就會自動關聯，不必手寫 id。

## 可及性

- 一律搭配 `FieldLabel for` 使用（只有 placeholder 沒有標籤，讀屏與自動填入都會退化）。
- `aria-invalid` 由 `Field` 的 `invalid` 驅動（`aria-invalid:border-destructive` 也會跟著變紅）。
- `disabled` 時 `cursor-not-allowed` + `opacity-50`；`type="file"` 的按鈕樣式已用 `file:` 變體統一。

## 範例

```tsx
import { Field, FieldLabel, Input } from "~/components/ui";

<Field invalid={!!error()}>
  <FieldLabel for="email">電子郵件</FieldLabel>
  <Input id="email" type="email" autocomplete="email" placeholder="you@example.com" />
</Field>;
```
