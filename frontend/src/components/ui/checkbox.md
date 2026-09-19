# Checkbox

勾選框。勾選狀態機、鍵盤操作與 ARIA 由 Ark `checkbox` 提供；結構照 Tailkit
（`a-c-form-elements-16`：方框 `size-4 rounded-sm`，勾選態 `bg-primary`）。

## API

| Prop | 型別 | 預設 | 說明 |
| --- | --- | --- | --- |
| `checked` | `boolean` | — | 受控勾選狀態 |
| `defaultChecked` | `boolean` | — | 非受控初始值 |
| `onCheckedChange` | `(checked: boolean) => void` | — | 收斂成純值（不外洩 Ark 的 `details`） |
| `disabled` | `boolean` | — | 直通 Ark |
| `class` | `string` | — | 追加到 Ark `Checkbox.Root`（渲染成 `<label>`） |
| 其餘 | `JSX.LabelHTMLAttributes<HTMLLabelElement>` | — | 直通 `<label>`；**`children` 被 `Omit`**（見「已知限制」） |

`Root` 渲染 `<label>`、`HiddenInput` 才是真正的 `<input type="checkbox">`；`Control` 是視覺方框，
`Indicator` 只在勾選時渲染。

## Ark 對應

`checkbox`：`Root` / `HiddenInput` / `Control` / `Indicator`（五個部件中最常用的四個）。

## 可及性

- 真正的 `<input>` 由 Ark 產生，**可辨識名稱一律用 `aria-label`**（目前不支援可見標籤文字）。
  用在權限矩陣這種「列／欄表頭已經說明語意」的場合剛好；需要可見文字時請由呼叫端自己在旁邊渲染。
- 焦點環畫在 `Control` 上，用 `peer` 指向同層、位於 `Control` 之前的 `HiddenInput`；
  `HiddenInput` 另加 `sr-only`，避免 `@tailwindcss/forms` 為原生 checkbox 上底色與藍框。
- 鍵盤：Space 切換（Ark 提供）。

## 範例

```tsx
import { Checkbox } from "~/components/ui";

<Checkbox aria-label="客戶 檢視" defaultChecked />
<Checkbox aria-label="全選" checked={all()} onCheckedChange={(next) => setAll(next)} />
<Checkbox aria-label="停用" disabled />
```

## 已知限制

- `children` 被型別 `Omit` 掉、實作也沒渲染（`{...rest}` 傳進 Ark 的 `children` 會被明確的 JSX 子節點
  覆蓋），因此**目前沒有可見標籤**；要加上請改 `CheckboxProps` 與 `Root` 的 JSX，兩處要一起改。
