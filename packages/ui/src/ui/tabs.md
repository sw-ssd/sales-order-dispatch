# Tabs

分頁籤。選取狀態、鍵盤操作、ARIA 關聯一律交給 Ark `tabs`；結構照 Tailkit
（`a-c-tabs-11` In Card Alternate：膠囊在淺底容器內）。

## API

| 部件 | 底層 | 主要 props |
| --- | --- | --- |
| `Tabs` | `ArkTabs.Root` | `value`（受控）、`defaultValue`、`onValueChange: (value: string) => void`（純值）、`orientation`（`horizontal` 預設／`vertical`）、`class`；其餘直通 Ark Root |
| `TabsList` | `ArkTabs.List` | `class`（`bg-muted` 膠囊容器） |
| `TabsTrigger` | `ArkTabs.Trigger` | `value`（**必填**）、`class` |
| `TabsContent` | `ArkTabs.Content` | `value`（**必填**）、`class` |

`lazyMount` + `unmountOnExit` 由包裝層固定開啟，**故意不開放覆寫**：維持「未選中的 panel 不在 DOM」
的既有語意（`value`／`defaultValue`／`onValueChange` 因此也被包裝層接走）。

## Ark 對應

`tabs`：`Root` / `List` / `Trigger` / `Content` / `Indicator`。

## 可及性

- `role="tablist"`／`role="tab"`／`role="tabpanel"` 與 `aria-controls`／`aria-labelledby` 由 Ark 產生。
- 鍵盤：左右（或上下）方向鍵切換、Home/End 跳頭尾；未選中的 panel 不在 DOM，因此也不在 tab 順序裡。
- 觸發鈕在 `variant="vertical"` 時靠 `data-[orientation=vertical]` 變體改為左靠排列，不需要另寫樣式。

## 範例

```tsx
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui";

<Tabs value={tab()} onValueChange={setTab}>
  <TabsList>
    <TabsTrigger value="employee">員工</TabsTrigger>
    <TabsTrigger value="store">店家</TabsTrigger>
  </TabsList>
  <TabsContent value="employee">…</TabsContent>
  <TabsContent value="store">…</TabsContent>
</Tabs>;
```
