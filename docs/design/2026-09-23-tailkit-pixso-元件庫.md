# Tailkit 元件庫 → Pixso 設計資源

> **狀態**：轉換進行中（2026-09-23 起）
> **目標檔**：Pixso 檔 `TailKit`（**專屬檔，非** `多公司訂出貨系統` UI 稿）
> **需求**：使用者指示「將 tailkit mcp server 取得元件與 packages 轉成 pixso 設計文件資源透過 pixso mcp server」

---

## 1. 成果範圍

| 項目 | 內容 |
|---|---|
| 來源 | Tailkit MCP：3 packages、88 subcategories、**646 個元件** |
| 產物 | Pixso 檔內 page「Tailkit 元件庫」：88 個 subcategory frame，內含 646 個 component |
| 變體 | 同 subcategory 的 N 個獨立元件再合成 **component set**（單一 `style` 軸，值 = 元件 Title） |
| 對照表 | 工作目錄 `manifest.json`（identifier ↔ Pixso node id ↔ 尺寸 ↔ subcategory） |

覆蓋率以官方 `tailkits://catalog/overview` 的每子類別計數逐項核對，646/646 相符。

## 2. 為什麼需要自建 CSS

直接餵 Tailkit 的 HTML 給 Pixso 只會得到無色空殼：Pixso 的 `code_to_design` **不執行 Tailwind**。
實測（`a-c-alerts-01`）確認未編譯時 `rounded-xl`、`p-4`、`gap-2` 全部失效、`bg-secondary-50` 完全不上色。

因此管線在送件前先自行編譯 Tailwind v4 CSS（含 `@tailwindcss/forms`／`typography`），並內嵌進 HTML 的 `<style>`：

```css
@theme {
  /* Tailkit 用自訂色階 secondary-*（stock Tailwind 沒有），別名到 gray */
  --color-secondary-50 … --color-secondary-950: var(--color-gray-*);
  --container-8xl/9xl/10xl: 90rem/105rem/120rem;
  --animate-spin-slow: spin-slow 8s linear infinite;
}
```

補了這兩組後：`max-w-10xl`（44 處）、`max-w-7xl`、`animate-spin-slow`、`prose` 才正確生成。

## 3. 管線

```
tailkit browse_catalog（offset 分頁）→ catalog.json（646）
  → get_component_code（批次 10）→ raw/*.html
  → 剝除 Tailkit 專有 icon font class（hi-micro hi-x-mark 等，Pixso 無此字型；inline SVG 保留）
  → Tailwind v4 CLI 編譯 → out3.css
  → 每批 20 個元件包成 <div id="tk-slot-N">…</div> 送 code_to_design
  → eval_script 拆解：slot → 所屬 subcategory frame → createComponentFromNode → 改名為 identifier
  → 同 subcategory combineAsVariants → component set
```

工具與可重跑步驟見 `docs/design/tailkit-pixso/README.md`；腳本同目錄。

## 4. 實作時確認的 Pixso 行為（本輪新發現，補 §7.3 既有清單）

1. **`code_to_design` 是單一 shot、不看 viewport 高度**：artboard 高 900 或 1600，同一元件量到的 slot 高度完全相同 → 不必為高度調整 artboard。
2. **但超長畫布會被拒**：`height: 40000` 回 `Page is too large`（Chrome `captureScreenshot` 限制），故 batch 需搭配 900 高的 viewport。
3. **slot 需要 `display:flow-root`**：不加時 `a-c-dividers-01` 的 slot 只有 28px 而內容 92px（float／margin 塌陷）。加了之後每個 slot 自成 BFC，高度正確。
4. **批次失敗會整批回滾**（`res is not iterable`）：須有逐件重試路徑，否則整批要重做。
5. **`currentPage` 決定 `code_to_design` 的落點**，切換 page 後不必重新初始化連線。
6. **`get_export_image` 需要完整 `exportSettings`**：`{ imageType: 1, constraint: { type: 1, value: 1 } }`（1=scale、2=寬、3=高），少給會回 `invalid_type`。
7. **`pixso.currentPage` 不能直接刪**：要先把 `currentPage` 切到別的 page 才能 `remove()`。
8. **`combineAsVariants` 在 auto-layout 父層內可行**，但要先把父層 `layoutMode` 暫時設為 `NONE`，合成後還原。

## 5. 授權邊界

Tailkit 為付費授權產品（`frontend/AGENTS.md` §2：不得把 Tailkit 產物原樣入庫、不得寫進 `package.json`）。
因此：

- **不進版控**：`raw/`、`scan/`、`out3.css`、`manifest.json`（Tailkit 原文與衍生 CSS）
- **進版控**：`docs/design/tailkit-pixso/` 的**工具腳本**與本文件（我方撰寫的管線程式碼）
- 產出物落在 Pixso 檔內，屬設計稿用途，與程式碼庫的 npm 依賴無關

## 6. 後續

- 轉換完成後執行 `10-variants.mjs` 合成 component set
- `21-verify.mjs` 逐元件檢查空元件／文字溢出／尺寸異常
- `31-shots.mjs <id>` 抽樣匯出 PNG 目視
