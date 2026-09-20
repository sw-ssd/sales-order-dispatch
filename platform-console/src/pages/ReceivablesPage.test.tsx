import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { fireEvent, render, screen, waitFor } from "@solidjs/testing-library";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import * as receivablesLib from "../lib/receivables";
import { WireFault, stubPlatformWire, type WireCall } from "../test-helpers";
import ReceivablesPage from "./ReceivablesPage";

/**
 * 契約：
 * - 清單與分頁由後端決定（不是前端全量載入）；「逾期」是期末已過的事實，前端標記。
 * - 記收款走 T9 的 `RecordPayment`：金額**留空＝採用期別快照**（線路上不送 amount 欄位），
 *   填了就送該金額（後端比對期別金額，不符回 PLAT-3002）。
 * - `reason` 是必填：空字串與全空白都在送出前擋掉，**線路上不會出現任何 RecordPayment**。
 * - `PLAT-3002` 依 `details.reason` 給不同的行動指引，且不是「不可重試」。
 * - 成功後付款清單自動失效（不靠手動重整）。
 * - **金額 0 先擋**：後端把 `AmountCents=0` 當成「採期別快照」，填 0 會把整期記成已收（不是「本期不收」）。
 * - 最後一頁被收款清空時，**分頁控制必須還在**（否則就是死路：看不到「沒有未付期別」以外的線索）。
 */
const ROWS = [
  {
    companyId: "42",
    companyName: "甲公司",
    planCode: "std",
    periodNo: 2,
    amount: "1500.00",
    periodEnd: "2026-08-31T00:00:00Z", // 期末已過 → 逾期
    status: "open",
  },
  {
    companyId: "43",
    companyName: "乙公司",
    planCode: "pro",
    periodNo: 1,
    amount: "3000.00",
    periodEnd: "2999-09-30T00:00:00Z",
    status: "open",
  },
];

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(() => (
    <QueryClientProvider client={client}>
      <ReceivablesPage />
    </QueryClientProvider>
  ));
  return client;
}

const receivablesCalls = (calls: WireCall[]) => calls.filter((c) => c.method === "ListReceivables");
const paymentCalls = (calls: WireCall[]) => calls.filter((c) => c.method === "RecordPayment");

/** 開甲公司的「記收款」對話框（dialog 內容是 Portal，故用 screen 查）。 */
async function openPaymentDialog() {
  await screen.findByText("甲公司");
  fireEvent.click(screen.getAllByRole("button", { name: "記收款" })[0]);
  await screen.findByRole("button", { name: "確認收款" });
}

/** 送出前填原因（其餘欄位由各測試自行填）。 */
function fillReason(value: string) {
  fireEvent.input(screen.getByLabelText(/原因/), { target: { value } });
}

let calls: WireCall[] = [];

afterEach(() => {
  vi.restoreAllMocks();
});

beforeEach(() => {
  calls = stubPlatformWire({
    ListReceivables: () => ({ rows: ROWS, pagination: { page: 1, pageSize: 50, total: 2 } }),
    RecordPayment: () => ({ periodNo: 2, status: "paid" }),
  });
});

describe("ReceivablesPage", () => {
  it("列出待收款並標記逾期（後端分頁）", async () => {
    renderPage();

    await screen.findByText("甲公司");
    expect(screen.getByText("乙公司")).toBeTruthy();
    expect(screen.getByText("1500.00")).toBeTruthy();
    expect(screen.getByText("逾期")).toBeTruthy();
    expect(screen.getByText("未付")).toBeTruthy();
    expect(cellTextOf("甲公司", 1)).toBe("std");
    expect(receivablesCalls(calls)[0].body).toEqual({ page: 1, pageSize: 50 });
  });

  it("沒有未付期別時顯示空狀態（不出現表格）", async () => {
    stubPlatformWire({
      ListReceivables: () => ({ rows: [], pagination: { page: 1, pageSize: 50, total: 0 } }),
    });
    renderPage();

    await waitFor(() => expect(screen.getByText(/沒有未付期別/)).toBeTruthy());
    expect(screen.queryByRole("table", { name: "待收款清單" })).toBeNull();
  });

  it("金額留空＝採用期別快照：線路上不送 amount，並帶期別、發票號、交易號與 trim 後的原因", async () => {
    const before = receivablesCalls(calls).length;
    renderPage();
    await openPaymentDialog();

    fireEvent.input(screen.getByLabelText(/發票號/), { target: { value: "AB12345678" } });
    fireEvent.input(screen.getByLabelText(/交易號/), { target: { value: "TX-1" } });
    fireEvent.input(screen.getByLabelText(/備註/), { target: { value: "溢收 100 元" } });
    fillReason("  匯款入帳  ");
    fireEvent.click(screen.getByRole("button", { name: "確認收款" }));

    await waitFor(() => expect(paymentCalls(calls)).toHaveLength(1));
    const call = paymentCalls(calls)[0];
    expect(call.url).toBe("/platform/platform.v1.PlatformAdminService/RecordPayment");
    // 金額留空 → proto3 的預設值不會上線：沒有 amount 這個欄位，後端才採期別快照。
    expect(call.body).toEqual({
      companyId: "42",
      periodNo: 2,
      provider: "manual",
      externalRef: "TX-1",
      invoiceNo: "AB12345678",
      note: "溢收 100 元",
      reason: "匯款入帳",
    });

    // 成功 → 清單自動失效（後端那筆已不是未付）。
    await waitFor(() => expect(receivablesCalls(calls).length).toBeGreaterThan(before));
  });

  it("金額 0 先擋：後端把 0 當成採期別快照，填 0 會把整期記成已收", async () => {
    renderPage();
    await openPaymentDialog();

    for (const zero of ["0", "0.00"]) {
      fireEvent.input(screen.getByLabelText(/金額/), { target: { value: zero } });
      fillReason("客戶說本期不收");
      fireEvent.click(screen.getByRole("button", { name: "確認收款" }));

      await waitFor(() => expect(screen.getByText(/金額必須大於 0/)).toBeTruthy());
      expect(paymentCalls(calls)).toHaveLength(0);
    }
  });

  it("最後一頁被收款清空：仍看得見分頁控制，且空狀態說的是「這一頁」", async () => {
    // 共 51 筆、每頁 50 → 第 2 頁只有一列；收款成功後那一頁變空。
    let page2Paid = false;
    stubPlatformWire({
      ListReceivables: (body) =>
        body.page === 2
          ? page2Paid
            ? { rows: [], pagination: { page: 2, pageSize: 50, total: 50 } }
            : { rows: [ROWS[1]], pagination: { page: 2, pageSize: 50, total: 51 } }
          : { rows: [ROWS[0]], pagination: { page: 1, pageSize: 50, total: 51 } },
      RecordPayment: () => {
        page2Paid = true;
        return { periodNo: 1, status: "paid" };
      },
    });
    renderPage();
    await screen.findByText("甲公司");

    fireEvent.click(await screen.findByRole("button", { name: "下一頁" }));
    await screen.findByText("乙公司");

    fireEvent.click(screen.getByRole("button", { name: "記收款" }));
    await screen.findByRole("button", { name: "確認收款" });
    fillReason("匯款入帳");
    fireEvent.click(screen.getByRole("button", { name: "確認收款" }));

    await waitFor(() => expect(screen.getByText(/第 2 頁沒有未付期別/)).toBeTruthy());
    // 死路的判準：分頁控制必須還在，否則營運回不了第 1 頁。
    expect(screen.getByRole("button", { name: "上一頁" })).toBeTruthy();
    expect(screen.getByText(/共 50 筆/)).toBeTruthy();
  });

  it("金額手動輸入時原樣送上線；格式錯誤先擋（後端 ParseCents 才是真偽決定者）", async () => {
    renderPage();
    await openPaymentDialog();

    fireEvent.input(screen.getByLabelText(/金額/), { target: { value: "1,500" } });
    fillReason("匯款入帳");
    fireEvent.click(screen.getByRole("button", { name: "確認收款" }));

    await waitFor(() => expect(screen.getByText(/金額格式錯誤/)).toBeTruthy());
    expect(paymentCalls(calls)).toHaveLength(0);

    fireEvent.input(screen.getByLabelText(/金額/), { target: { value: "1500.00" } });
    fireEvent.click(screen.getByRole("button", { name: "確認收款" }));

    await waitFor(() => expect(paymentCalls(calls)).toHaveLength(1));
    expect(paymentCalls(calls)[0].body.amount).toBe("1500.00");
  });

  it("reason 空字串與全空白都會被擋，且線路上沒有任何 RecordPayment", async () => {
    renderPage();
    await openPaymentDialog();

    fireEvent.click(screen.getByRole("button", { name: "確認收款" }));
    await waitFor(() => expect(screen.getByText(/原因必填/)).toBeTruthy());

    fillReason("   ");
    fireEvent.click(screen.getByRole("button", { name: "確認收款" }));

    await waitFor(() => expect(screen.getByText(/原因必填/)).toBeTruthy());
    expect(paymentCalls(calls)).toHaveLength(0);
  });

  it("PLAT-3002（金額不符）：指引採用期別快照／依期別金額輸入，不是「不可重試」", async () => {
    stubPlatformWire({
      ListReceivables: () => ({ rows: ROWS, pagination: { page: 1, pageSize: 50, total: 2 } }),
      RecordPayment: () => {
        throw new WireFault({
          code: "PLAT-3002",
          message: "收款衝突：輸入金額 3000.00 與期別金額 1500.00 不符（不支援部分付款；差異請記於備註）",
          details: {
            reason: "輸入金額 3000.00 與期別金額 1500.00 不符（不支援部分付款；差異請記於備註）",
          },
        });
      },
    });
    renderPage();
    await openPaymentDialog();

    fireEvent.input(screen.getByLabelText(/金額/), { target: { value: "3000.00" } });
    fillReason("客戶匯款");
    fireEvent.click(screen.getByRole("button", { name: "確認收款" }));

    const alert = await screen.findByRole("alert");
    expect(alert.textContent).toContain("PLAT-3002");
    // 指引的字（「留空」）不在後端訊息裡：有它才證明指引有分流。
    expect(alert.textContent).toMatch(/留空/);
    expect(alert.textContent).not.toMatch(/不可重試|請勿重試/);
  });

  it("PLAT-3002（期別已付款且交易號不同）：指引確認是否重複收款，不是同一段話", async () => {
    stubPlatformWire({
      ListReceivables: () => ({ rows: ROWS, pagination: { page: 1, pageSize: 50, total: 2 } }),
      RecordPayment: () => {
        throw new WireFault({
          code: "PLAT-3002",
          message: '收款衝突：期別 2 已付款（交易號 "TX-1"），本次交易號 "TX-2" 不同',
          details: { reason: '期別 2 已付款（交易號 "TX-1"），本次交易號 "TX-2" 不同' },
        });
      },
    });
    renderPage();
    await openPaymentDialog();

    fireEvent.input(screen.getByLabelText(/交易號/), { target: { value: "TX-2" } });
    fillReason("客戶匯款");
    fireEvent.click(screen.getByRole("button", { name: "確認收款" }));

    const alert = await screen.findByRole("alert");
    expect(alert.textContent).toMatch(/沿用原交易號/);
    expect(alert.textContent).not.toMatch(/留空/);
    expect(alert.textContent).not.toMatch(/不可重試|請勿重試/);
  });

  it("匯出 CSV：以純函式組出目前列出的資料（不碰 Blob 的時序）", async () => {
    const download = vi.spyOn(receivablesLib, "downloadCsv").mockImplementation(() => {});
    renderPage();
    await screen.findByText("甲公司");

    fireEvent.click(screen.getByRole("button", { name: /匯出 CSV/ }));

    expect(download).toHaveBeenCalledTimes(1);
    const [filename, csv] = download.mock.calls[0];
    expect(filename).toBe("receivables.csv");
    // 開頭的 UTF-8 BOM：沒有它，Windows Excel 會把中文顯示成亂碼。
    expect(csv.startsWith("\uFEFF")).toBe(true);
    expect(csv.split("\n")[0]).toBe("\uFEFF公司,方案,期別,金額,到期日,狀態");
    expect(csv.split("\n")[1]).toContain("甲公司,std,2,1500.00");
  });
});

/** 該公司那一列的第 n 格文字（0＝公司）。 */
function cellTextOf(company: string, column: number): string {
  const row = screen
    .getAllByRole("row")
    .find((r) => r.textContent?.includes(company) && r.querySelectorAll("td").length > 0);
  return row?.querySelectorAll("td")[column]?.textContent ?? "";
}
