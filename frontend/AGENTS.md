# Frontend（SolidJS）Development Guidelines

## 1. Strict SolidJS Reactivity Rules

1. **NEVER Destructure Props Directly**:
   - `const { variant, class: className } = props;` -> ❌ **FORBIDDEN** (breaks SolidJS fine-grained signal tracking).
   - `const [local, others] = splitProps(props, ["variant", "class"]);` -> ✅ **REQUIRED**.

2. **Children Inspection & Tab Hydration**:
   - ALWAYS wrap `props.children` with SolidJS's native `children(() => props.children)` memoization helper when inspecting, iterating, or rendering dynamic JSX child nodes.

3. **SSR Safety Guards**:
   - Always include `typeof window !== "undefined"` and `typeof document !== "undefined"` guards inside browser event listeners or DOM access logic.

## 2. UI 元件庫：Ark UI（行為）× Tailkit（視覺）

- 互動元件一律以 **Ark UI primitives**（`@ark-ui/solid`）實作行為（狀態機、焦點管理、鍵盤互動、ARIA），放在 `src/components/ui/`。
- **Tailkit 只作為結構與樣式的參考來源，不是 npm 依賴**：取其版面／間距／層級思路，改寫成 SolidJS 元件，不得把 Tailkit 產物原樣入庫、不得寫進 `package.json`。
- 狀態一律以 Ark 的 `data-state` / `data-selected` / `data-orientation` 表達，不要自寫 `aria-*` 或 `role`。
- 元件對外 API 照 Ark 語意，且不外洩 `details` 物件：回呼吃純值（`onOpenChange(open: boolean)`、`onValueChange(value: string)`、`onCheckedChange(checked: boolean)`）。

## 3. Tailwind CSS v4 Native Design Tokens

- All components must use semantic design tokens (`bg-background`, `text-foreground`, `bg-card`, `border-border`, `bg-primary`, etc.).
- 不得使用色階字面值（`secondary-*`、`orange-*`、`emerald-*`、`zinc-*` 等）；token 已在 `src/index.css` 定義並於深色模式翻轉，因此也不再寫顏色相關的 `dark:` 變體。

## 4. Code Search / Index via codebase-memory-mcp

All code indexing and querying in this project **must go through codebase-memory-mcp first** (see root `docs/AGENTS.md` §4.0): locate definitions/implementations/callers, trace call paths, analyze blast radius, and traverse across projects via the codebase-memory knowledge graph; fall back to grep/direct file reads only when insufficient. Confirm index coverage of a target file before modifying existing code.

## 5. 測試：斷言必須「能失敗」

本專案前端測試的用途是**契約守門**,不是覆蓋率裝飾。寫測試時逐條自問「**什麼改動會讓它紅?**」——答不出來就是假綠。已在本專案抓到的假綠形狀(勿再犯):

1. **以「某文案不存在」代表狀態** → 文案一改就失去鑑別力。改為**結構判定**(`tbody` 列數、`queryAllByRole("row")`)或斷言**應該出現**的內容。
2. **期望值由被測畫面自己產生**(`expect(x).toHaveLength(headers.length)`、`colspan == headers.length`) → 兩邊同源,整欄被刪也全綠。改為**字面值**(`toHaveLength(6)`、`colspan === "6"`)。
3. **原生元素的預設值當契約**(`button.tabIndex === 0`) → jsdom 預設即 0。刪掉,或改斷言真正的行為(`tagName`、`document.activeElement`、點擊後焦點仍在)。
4. **斷言「不存在的屬性」**(`tr.getAttribute("role") === null`) → 任何非互動元素都成立。改為斷言緊接的**行為**(點列不會選取、不會發請求)。
5. **函式形式的選項只看單點**(`retry(0)=true`、`retry(3)=false` 卻號稱「封頂 3 次」) → 上限改成 1/2 也會過。補齊中間值。
6. **突變驗證**:新增或強化的斷言,交付前至少手動把對應實作改壞一次確認它會紅(並在報告寫下該實驗);**不能只靠「測試全綠」當證據**。

