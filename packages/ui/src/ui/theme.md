# Theme（深色模式）

全 app **唯一**的主題機制（Tailkit class-based dark mode）：偏好三態 `light`／`dark`／`system`，
持久化在 localStorage `ui:theme`，套用結果是 `document.documentElement` 上的 `.dark`。

## 匯出

| 名稱 | 型別 | 說明 |
| --- | --- | --- |
| `ThemeProvider` | `ParentComponent` | 掛載時只**讀** localStorage；偏好真的被切換過才寫回 |
| `useTheme` | `() => { theme, setTheme }` | 不在 `ThemeProvider` 內時直接拋錯 |
| `ThemeSwitcher` | `Component<{ class?: string }>` | 分段控制（淺色／跟隨系統／深色），形狀是 Ark `toggle-group` |
| `THEME_STORAGE_KEY` | `"ui:theme"` | 持久化 key |
| `THEME_DARK_QUERY` | `"(prefers-color-scheme: dark)"` | 系統偏好查詢 |
| `THEME_DARK_CLASS` | `"dark"` | 掛在 `documentElement` 的 class |
| `isTheme` | `(value: unknown) => value is Theme` | 值守衛 |

兩段接力，缺一不可：

1. `index.html` 的 anti-FOUC 內聯腳本在任何 CSS 與 bundle 之前同步套用同一規則
   （不引用外部檔，因此那段邏輯與這裡**刻意重複兩份**：改規則時兩邊都要改）。
2. `theme.tsx` 接手之後的切換（`system` 時跟著 `prefers-color-scheme` 即時變）。

## Ark 對應

`toggle-group`（`ThemeSwitcher`）：`Root value={[theme()]}` + `deselectable={false}` +
每個 `ToggleGroup.Item`。

## 可及性

- `ToggleGroup.Root` 帶 `aria-label="主題"`，每項 `role="radio"` 且各有 `aria-label`（`淺色`／
  `跟隨系統`／`深色`），可 Tab 進入、方向鍵切換、永遠有一項選中。
- 深色模式只換 token（`index.css` 的 `.dark`），**不得**在元件內寫顏色相關的 `dark:` 變體。
- `ThemeProvider` 不看 `chromeless`：登入頁／403 也要跟著使用者偏好。

## 範例

```tsx
import { ThemeProvider, ThemeSwitcher } from "~/components/ui";

<ThemeProvider>
  <ThemeSwitcher />
  …app…
</ThemeProvider>;
```

深色預覽一律用這個切換器；**不要**在測試或 demo 頁自己 `classList.add("dark")`，
那會變成第二套主題機制。
