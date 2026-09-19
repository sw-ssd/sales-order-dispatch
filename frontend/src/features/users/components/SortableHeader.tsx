import type { Component, JSX } from "solid-js";

/**
 * 可排序表頭的控制項（三張清單共用：公司／部門／角色）。
 *
 * 契約（設計 D1／D11）：
 * - **點表頭＝改排序**，一律送服務端排序（`manualSorting`）；前端不做本地排序。
 * - 表頭以**原生 `<button type="button">`** 承載：可聚焦、可及名稱就是欄位名（方向圖示
 *   `aria-hidden`，不進名稱），鍵盤與滑鼠路徑同一顆控制項。
 * - 方向語意掛在 `<th aria-sort>`（`ascending`／`descending`／`none`，不可排序的欄位則不帶
 *   這個屬性）——由 `ariaSort()` 產生，頁面在 `<TableHead>` 上接線；`ui/table.tsx` 的
 *   `TableHead` 會把其餘 props 展開到 `<th>`，所以不必動元件庫。
 * - 切換循環用 v9 的預設（`enableSortingRemoval` 未關）：未排 → 升冪 → 降冪 → 未排。
 *   取消排序＝`sort` 空字串＝服務預設排序（D1 已定義），語意一致。
 * - 表頭必須以 `sortDescFirst: false`（table 層，見各頁 features 旁的註解）起算：v9 的
 *   「第一個方向」預設是**依資料推測**（非字串欄位會從降冪起），資料為空時更會退化成降冪。
 *   白名單欄位一律從**升冪**起算才是可預期的行為。
 */

/**
 * 這個控制項用得到的 tanstack v9 `Column` 子集（結構型別）。
 * 三張表的 features／列型別各不相同，只取用得到的幾個成員即可，不必為每個 feature set
 * 各寫一次泛型元件（欄位定義本身仍留在各頁，見 D8）。
 */
export interface SortableColumn {
  id: string;
  getCanSort: () => boolean;
  getIsSorted: () => false | "asc" | "desc";
  getToggleSortingHandler: () => undefined | ((event: unknown) => void);
}

/** 方向圖示：只給視覺，語意方向由 `<th aria-sort>` 承載。 */
const SORT_ARROWS: Record<string, string> = { asc: "▲", desc: "▼" };

/** `<th>` 的 `aria-sort`；不可排序的欄位回 `undefined`（不帶這屬性）。 */
export function ariaSort(column: SortableColumn): "ascending" | "descending" | "none" | undefined {
  if (!column.getCanSort()) return undefined;
  const direction = column.getIsSorted();
  if (direction === "asc") return "ascending";
  if (direction === "desc") return "descending";
  return "none";
}

export interface SortableHeaderProps {
  column: SortableColumn;
  label: string;
}

/**
 * 可排序表頭（按鈕）。點擊直接走 table 自己的 toggle handler——排序 state 的寫入者仍是
 * table（各頁以 `state.sorting` ＋ `onSortingChange` 受控），這一層不自己動 state。
 */
export const SortableHeader: Component<SortableHeaderProps> = (props) => {
  const direction = () => props.column.getIsSorted();

  return (
    <button
      type="button"
      onClick={(event) => props.column.getToggleSortingHandler()?.(event)}
      // 負邊距＋等量內距：焦點環貼著文字，左緣仍與 `<th>` 的內距對齊。
      class="-mx-1 inline-flex items-center gap-1 rounded px-1 transition-colors hover:text-primary focus-visible:ring-3 focus-visible:ring-ring focus-visible:outline-none"
    >
      {props.label}
      <span aria-hidden="true" class="text-xs">
        {SORT_ARROWS[direction() || ""]}
      </span>
    </button>
  );
};

/** 建立該頁專用的表頭控制項產生器（快取生命週期＝一個頁面實例）。 */
export function createSortableHeaders() {
  /**
   * 每個欄位只建一次的表頭控制項。
   *
   * 為什麼要快取**節點**：`TableHead` 把 children 當 getter 插入，而 children 來自
   * `flexRender(columnDef.header, ctx)` → `createComponent(...)`；getter 每次求值都會建一個
   * **新的元件實例**，所以 table state 一變（換頁／排序）就會重新插入 children、以新節點取代
   * 舊節點——新舊不同時 Solid 會重建 DOM，**焦點就掉回 `<body>`**（之後的 Tab 從頁首重來，
   * D11 的「排序變更後焦點不亂跳」即失效）。實測：回傳同一個節點時，Solid 的
   * `insertExpression` 因值相同直接沿用，DOM 不動、焦點留在同一顆按鈕上。
   *
   * 快取必須是**每個頁面實例一份**（不可放模組層）：否則測試的反覆掛載會共用上一輪的節點。
   */
  const nodes = new Map<string, JSX.Element>();

  return (column: SortableColumn, label: string): JSX.Element => {
    const node = nodes.get(column.id) ?? <SortableHeader column={column} label={label} />;
    nodes.set(column.id, node);
    return node;
  };
}
