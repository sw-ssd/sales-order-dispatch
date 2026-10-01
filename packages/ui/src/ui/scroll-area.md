# ScrollArea

捲動容器。量測、拖曳、鍵盤與滾輪行為全部交給 Ark `scroll-area`。

## API

| Prop | 型別 | 預設 | 說明 |
| --- | --- | --- | --- |
| `orientation` | `vertical \| horizontal \| both` | `vertical` | 決定掛哪幾個 `Scrollbar` |
| `class` | `string` | — | 追加到 Ark `Root`；**高度一律下在這裡**（`max-h-*`／`h-*`） |
| `children` | `JSX.Element` | — | 內容 |
| 其餘 | `JSX.HTMLAttributes<HTMLDivElement>` | — | 直通 Root |

內部解剖：`Root`（`relative flex flex-col overflow-hidden`）> `Viewport`（`overflow: auto`，Ark 設定）>
`Content`，加上依 `orientation` 掛載的 `Scrollbar` / `Thumb` 與 `Corner`。

## Ark 對應

`scroll-area`：`Root` / `Viewport` / `Content` / `Scrollbar(orientation)` / `Thumb` / `Corner`。

## 可及性

- `Viewport` 可聚焦（`focus-visible:ring-3 focus-visible:ring-ring`），因此鍵盤（PageUp/PageDown、
  方向鍵、Home/End）能捲動；原生的 `::-webkit-scrollbar` 與 `scrollbar-width` 被藏起來，改用 Ark 的
  自繪捲軸。
- 捲軸顯隱交給 Ark 的 `data-hover`／`data-scrolling`（指標在 Root 內就會標記），並帶 `opacity-0`
  過場；純鍵盤使用者不受影響。
- 顏色只有 `--border`（thumb）、`--muted-foreground`（thumb hover）、`--ring`。

## 範例

```tsx
import { ScrollArea } from "~/components/ui";

// 垂直：高度下在 Root 上
<ScrollArea class="max-h-64 w-full">
  <div class="space-y-2 p-4">{/* 長內容 */}</div>
</ScrollArea>;

// 兩向
<ScrollArea orientation="both" class="h-64 w-full">
  <div class="w-[80rem]">…</div>
</ScrollArea>;
```

## 已知限制

- `Viewport` **不能**改用 Ark 範例的 `h-full`：`height: 100%` 在 `max-h` 父層下會量到比可見區更高的高度
  （Edge 實測 viewport 472px / 可見 378px），尾端內容會被 Root 的 `overflow-hidden` 裁掉而捲不到。
- 只有 `Vertical`／`Horizontal` 兩向的 `Scrollbar`；`Corner` 不帶樣式（兩向同時出現時由 thumb 交會處填滿）。
