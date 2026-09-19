# Field（多部件）

單一欄位的包裝層（標籤／控件／說明／錯誤）。a11y 關聯全部交由 Ark `Field`；版面沿用 Tailkit 表單的
`div.space-y-1`（`a-c-form-elements-01`、`a-c-form-layouts-02`）。

## API

| 部件 | 底層 | 主要 props |
| --- | --- | --- |
| `Field` | `ArkField.Root` | `invalid?: boolean`（見下）、`class`，其餘是 `JSX.HTMLAttributes<HTMLDivElement>` |
| `FieldLabel` | `ArkField.Label` | `for`（直通，維持與頁面控件的 id 關聯）、`class` |
| `FieldDescription` | `ArkField.HelperText` | `class`（內建 `block`，維持原本 `<p>` 的區塊版面） |
| `FieldError` | `ArkField.ErrorText` | `class`；**不在 `Field` 內**時退回獨立 `<span role="alert">` |

`invalid` 必須由呼叫端明傳：Solid 的 JSX 在編譯期就把子元件展開成 DOM，包裝層拿不到子元件型別，
無法「偵測有沒有 `FieldError` 子節點」。

`required`／`disabled`／`readOnly` **沒有**從包裝層開放（`FieldProps` 只到 `div` 屬性，`{...rest}`
雖然會 spread 進 Ark Root 但過不了型別）：請下在控件本身（`Input` 的 `required`／`disabled`）。

## Ark 對應

`field`：`Root`（由 `invalid` 推導 `aria-invalid`、錯誤訊息 id、`data-invalid`）/ `Label` /
`HelperText` / `ErrorText` / `Input` / `RequiredIndicator`。

## 可及性

- `FieldLabel for` + 控件的 `id` 要對上；`Input`／`Checkbox` 這類有接 Ark field context 的控件會自動
  取得 `aria-describedby`（說明 + 錯誤）與 `aria-invalid`（見 `input.md`）。
- `FieldError` 只在 `Field` 為 invalid 時渲染（Ark 的 `ErrorText` 行為）；放 `Field` 之外時由包裝層
  自帶 `role="alert"`，適合登入頁那種「表單層單一錯誤」。
- 錯誤訊息用文字說清楚原因，不要只把邊框變紅。

## 範例

```tsx
import { Field, FieldDescription, FieldError, FieldLabel, Input } from "~/components/ui";

<Field>
  <FieldLabel for="name">公司名稱</FieldLabel>
  <Input id="name" value={name()} onInput={(e) => setName(e.currentTarget.value)} />
  <FieldDescription>發票上的正式名稱</FieldDescription>
</Field>;

<Field invalid={!!error()}>
  <FieldLabel for="code">統一編號</FieldLabel>
  <Input id="code" value={code()} onInput={(e) => setCode(e.currentTarget.value)} />
  <FieldError>{error()}</FieldError>
</Field>;
```
