# Tailkit → Pixso 元件庫管線

把 Tailkit MCP 的 646 個元件（3 個 package）轉成 Pixso 設計檔內的設計資源。
本目錄只放**工具腳本**；Tailkit 原始 HTML 為付費授權產物，**不得入庫**（見 `frontend/AGENTS.md` §2）。

## 產出

| 位置 | 內容 |
|---|---|
| Pixso 檔 `TailKit`（file key 見開檔連結） | page「Tailkit 元件庫」：88 個 subcategory frame，內含 646 個 component（同 subcategory 合成 component set） |
| `manifest.json`（工作目錄，不入庫） | identifier ↔ Pixso node id ↔ 尺寸 ↔ 所屬 subcategory |
| `preview/`（工作目錄，不入庫） | 646 張可獨立開啟的預覽頁 `<id>.html`（連 `../out3.css`）＋ `index.html` 索引 |

## 前置

- Pixso 桌面版開著、開好目標檔（MCP 只作用在**當前檔案**；本管線用一個專屬檔 `TailKit`，不碰 `多公司訂出貨系統`）
- Tailkit MCP 已授權（token 由 omp credential store 讀取，不落地）
- Node 24、`bunx`（只為跑 tailwind CLI）

所有腳本的工作目錄由 `TK_DIR` 決定（預設 `/tmp/tailkit-pixso`），可任意指定：

```bash
export TK_DIR=~/tailkit-work && mkdir -p "$TK_DIR"
```

## 執行順序

```bash
cd <工作目錄>
node 1-catalog.mjs      # 爬 catalog → catalog.json（含 offset 分頁，646 筆）
node 2-fetch.mjs        # 抓 646 份 HTML → raw/（每批 10 個）+ catalog.index.json
node 3-clean.mjs        # 去 Tailkit 專有 icon class → scan/
node 4-css.mjs          # 編譯 Tailwind v4 CSS → out3.css
node 5-preview.mjs      # scan/ → preview/<id>.html + index.html（瀏覽器可直接開啟的預覽頁）
node 23-convert-v2.mjs  # 逐批 code_to_design → 拆解成 component 歸位；進度寫 manifest.json
node 10-variants.mjs    # 同 subcategory 合成 component set
node 31-shots.mjs <id>… # 抽樣匯出 PNG 供目視
```

`23-convert-v2.mjs` 可中斷續跑：已轉換的 id 不重做。環境變數 `BATCH`（一次送幾個元件，預設 20）、`PARALLEL`（批次失敗時逐個重試的併發數，預設 3）。

## 為什麼不是一條 tailwind css 就夠

Tailkit 用**自訂色階 `secondary-*`**（stock Tailwind 沒有），且 `max-w-10xl` 來自 Tailkit 的 `@theme`。
沒補這兩組，整份稿會是無色無框的空殼（見 `docs/design/2026-09-23-tailkit-pixso-元件庫.md` §4）。

## 已知限制

- slot 需 `display:flow-root`（BFC），否則 float／margin 塌陷會讓高度小於內容
- Pixso 的 `code_to_design` 對超長畫布（height > 約 16000px）會 `Page is too large`，故每批 20 個並把 viewport 壓在 900
- 元件內的 `x-show`／`x-ref`（Alpine 指令）與 Tailkit icon font class 被剝除或忽略；inline SVG 一律保留並轉成 VECTOR
