import { render, screen, fireEvent, waitFor } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";
import { createSignal } from "solid-js";
import { Pagination, PaginationContent, PaginationNext, PaginationPrevious } from "./pagination";

/** 頁碼鈕（Ark 以 `itemLabel` 產生無障礙名稱，繁中預設為「第 N 頁」）。 */
const pageButtons = () => screen.getAllByRole("button", { name: /^第 \d+ 頁$/ });

/** 省略號由 Ark 產生，`data-part` 是 Ark 公開的樣式契約。 */
const ellipsisCount = () => document.querySelectorAll('[data-part="ellipsis"]').length;

interface HarnessProps {
  count?: number;
  pageSize?: number;
  siblingCount?: number;
  initialPage?: number;
  onPageChange?: (page: number) => void;
}

/** 受控用法：與 ListPagination 相同——頁碼由外部狀態持有，點擊後回寫。 */
const Harness = (props: HarnessProps) => {
  const [page, setPage] = createSignal(props.initialPage ?? 1);
  return (
    <Pagination
      count={props.count ?? 50}
      pageSize={props.pageSize ?? 10}
      siblingCount={props.siblingCount ?? 1}
      page={page()}
      onPageChange={(next) => {
        setPage(next);
        props.onPageChange?.(next);
      }}
    />
  );
};

describe("Pagination", () => {
  it("點下一頁回報新頁碼", async () => {
    const onPageChange = vi.fn();
    render(() => <Harness onPageChange={onPageChange} />);

    expect(screen.getByRole("button", { name: "第 1 頁" }).getAttribute("aria-current")).toBe(
      "page"
    );

    fireEvent.click(screen.getByRole("button", { name: "下一頁" }));

    // Ark 的狀態機要到下一個 microtask 才回報。
    await waitFor(() => expect(onPageChange).toHaveBeenCalledWith(2));
    await waitFor(() =>
      expect(screen.getByRole("button", { name: "第 2 頁" }).getAttribute("aria-current")).toBe(
        "page"
      )
    );
  });

  it("首頁時上一頁不可用", () => {
    render(() => <Harness initialPage={1} />);
    expect((screen.getByRole("button", { name: "上一頁" }) as HTMLButtonElement).disabled).toBe(
      true
    );
    expect((screen.getByRole("button", { name: "下一頁" }) as HTMLButtonElement).disabled).toBe(
      false
    );
  });

  it("最後一頁時下一頁不可用", () => {
    render(() => <Harness initialPage={5} />);
    expect((screen.getByRole("button", { name: "下一頁" }) as HTMLButtonElement).disabled).toBe(
      true
    );
    expect((screen.getByRole("button", { name: "上一頁" }) as HTMLButtonElement).disabled).toBe(
      false
    );
  });

  it("頁數少時全部頁碼可見、沒有省略號", () => {
    render(() => <Harness count={50} />);
    expect(pageButtons().map((button) => button.textContent)).toEqual([
      "1",
      "2",
      "3",
      "4",
      "5",
    ]);
    expect(ellipsisCount()).toBe(0);
  });

  it("超過可視窗的頁數時保留首尾邊界並以省略號收斂", () => {
    render(() => <Harness count={1000} initialPage={5} />);
    expect(pageButtons().map((button) => button.textContent)).toEqual([
      "1",
      "4",
      "5",
      "6",
      "100",
    ]);
    expect(ellipsisCount()).toBe(2);
  });

  it("頁碼鈕數量隨 siblingCount 變化", async () => {
    const [siblingCount, setSiblingCount] = createSignal(0);
    render(() => <Harness count={1000} initialPage={5} siblingCount={siblingCount()} />);

    expect(pageButtons().map((button) => button.textContent)).toEqual(["1", "5", "100"]);

    setSiblingCount(2);
    await waitFor(() =>
      expect(pageButtons().map((button) => button.textContent)).toEqual([
        "1",
        "2",
        "3",
        "4",
        "5",
        "6",
        "7",
        "100",
      ])
    );
  });

  it("自訂子節點可取代預設版型，樣式部件仍組裝在 Ark 的狀態上", async () => {
    const onPageChange = vi.fn();
    render(() => (
      <Pagination
        count={50}
        page={1}
        pageSize={10}
        onPageChange={onPageChange}
        aria-label="自訂分頁"
      >
        <PaginationContent>
          <li>
            <PaginationPrevious />
          </li>
          <li>
            <PaginationNext hideText />
          </li>
        </PaginationContent>
      </Pagination>
    ));

    expect(screen.getByRole("navigation", { name: "自訂分頁" })).toBeTruthy();
    // 預設版型的頁碼鈕整組被取代，只剩自訂的兩個觸發鈕。
    expect(screen.queryByRole("button", { name: /^第 \d+ 頁$/ })).toBeNull();
    expect((screen.getByRole("button", { name: "上一頁" }) as HTMLButtonElement).disabled).toBe(
      true
    );

    fireEvent.click(screen.getByRole("button", { name: "下一頁" }));
    await waitFor(() => expect(onPageChange).toHaveBeenCalledWith(2));
  });
});
