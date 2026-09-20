import { createSignal } from "solid-js";
import { render, waitFor } from "@solidjs/testing-library";
import { describe, expect, it } from "vitest";
import { createSortableHeaders } from "./SortableHeader";

/**
 * `createSortableHeaders()` 的 owner 生命週期。
 *
 * 為什麼不是從頁面測：`createSortableHeaders()` 的 `onCleanup` 掛在**呼叫它的 owner**（三頁都在
 * 頁面 body 呼叫），而排序狀態由 table 持有；頁面卸載後元件內的 `column`／state 就不可達，外部
 * 無從再觸發一次狀態變更。這裡改以最小 harness 提供 owner（`render` 的 root）與狀態（本檔的
 * signal），對**公開的** `createSortableHeaders()` 與它渲染出的 DOM 節點下斷言。
 *
 * 守的是 F2 修法唯一沒被守住的那一半（f2-fix-report §9.2）：快取節點建在專屬 `createRoot` 下，
 * 只有 `onCleanup` 收得掉它；少了它，owner 釋放後那顆 root 仍活著並持續訂閱排序狀態，
 * 寫入一個已脫離文件的節點（洩漏）。
 */

/** 手造的 column 替身：只提供 `SortableHeader` 用到的子集（結構型別，與任何頁面無關）。 */
function columnOf(direction: () => false | "asc" | "desc") {
  return {
    id: "name",
    getCanSort: () => true,
    getIsSorted: direction,
    getToggleSortingHandler: () => undefined,
  };
}

/** 等非同步工作排空（與各頁測試的 `settle()` 同義）。 */
async function settle() {
  const flushed = Promise.withResolvers<void>();
  setTimeout(flushed.resolve, 0);
  await flushed.promise;
}

describe("createSortableHeaders 的快取節點生命週期", () => {
  it("owner 釋放後 dispose 生效：同一顆表頭節點不再隨排序狀態更新", async () => {
    const [direction, setDirection] = createSignal<false | "asc" | "desc">(false);

    // harness＝owner：`createSortableHeaders()` 在此註冊 `onCleanup`，卸載時必須把快取節點的
    // 專屬 root 收乾淨（否則節點雖脫離文件，仍被訂閱著）。
    function Harness() {
      const createHeader = createSortableHeaders();
      return <div>{createHeader(columnOf(direction), "名稱")}</div>;
    }

    const { container, unmount } = render(() => <Harness />);
    const button = container.querySelector("button");
    if (button === null) throw new Error("harness 必須渲染出表頭按鈕");

    // 掛載期間是響應式的：狀態一變，同一顆表頭節點的文字就跟著變（F2 的修法在卸載前必須成立）。
    setDirection("asc");
    await waitFor(() => expect(button.textContent).toBe("名稱▲"));

    unmount();

    // owner 已釋放：再改狀態不得再寫回這顆節點（拿掉 `onCleanup` 時這裡會變成「名稱▼」）。
    setDirection("desc");
    await settle();
    expect(button.textContent).toBe("名稱▲");
  });
});
