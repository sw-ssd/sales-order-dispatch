# Dialog

對話框。開關、Escape、焦點鎖定、外部點擊一律交給 Ark `dialog`；卡片結構照 Tailkit
（`a-c-modals-01/02`，遮罩為 `bg-foreground/75`）。

## API

### `Dialog`

| Prop | 型別 | 預設 | 說明 |
| --- | --- | --- | --- |
| `open` | `boolean` | — | 受控開關 |
| `defaultOpen` | `boolean` | — | 非受控初始值（直通 Ark） |
| `onOpenChange` | `(open: boolean) => void` | — | 收斂成純值 |
| `closeOnOutsideClick` | `boolean` | `true` | 對應 Ark `closeOnInteractOutside` |
| 其餘 | `ArkDialog.RootProps` | — | 直通（`closeOnEscape`、`trapFocus`、`modal`…） |

### 部件

| 部件 | 元素 | 備註 |
| --- | --- | --- |
| `DialogContent` | `Portal` + `Backdrop` + `Positioner` + `Content` | 置中交給 Ark `Positioner`（不用 `translate` 手動偏移），內容包在 `ScrollArea`（`max-h-[80vh]`）裡 |
| `DialogClose` | `ArkDialog.CloseTrigger` | 直接 re-export，用於頁尾「取消」等自訂關閉鈕 |
| `DialogHeader` / `DialogFooter` | `div` | 版面用 |
| `DialogTitle` | `ArkDialog.Title` | 必填（見下） |
| `DialogDescription` | `ArkDialog.Description` | 選填 |

`DialogContent` 額外收 `showCloseButton`（預設 `true`，右上角 X）與 `blur`（預設 `true`，
遮罩 `backdrop-blur-sm`）。

## Ark 對應

`dialog`：`Root` / `Trigger` / `Backdrop` / `Positioner` / `Content` / `Title` / `Description` /
`CloseTrigger`。`Trigger` 沒有包裝層，需要時直接從 `@ark-ui/solid` 取用或自行以 `Button` 控制 `open`。

## 可及性

- `DialogTitle` 一定要放：Ark 會把 `aria-labelledby` 指向它，缺了就會留下指不到元素的屬性。
- 焦點鎖定、Esc 關閉、關閉後歸還焦點都由 Ark 處理；外部點擊可用 `closeOnOutsideClick={false}` 關掉
  （破壞性操作確認框建議關掉，避免誤觸關閉）。
- 右上角關閉鈕的無障礙名稱是「關閉」，寫死在元件內、呼叫端無法覆寫——本階段已知限制。
- 遮罩只用顏色 + 模糊，沒有 `aria-hidden` 需求（Ark 掌管底層）。

## 範例

```tsx
import { Button, buttonVariants, Dialog, DialogClose, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "~/components/ui";

const [open, setOpen] = createSignal(false);

<Button onClick={() => setOpen(true)}>新增</Button>
<Dialog open={open()} onOpenChange={setOpen}>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>新增公司</DialogTitle>
    </DialogHeader>
    {/* 表單 */}
    <DialogFooter>
      <DialogClose class={buttonVariants({ variant: "outline" })}>取消</DialogClose>
      <Button type="submit">儲存</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>;
```
