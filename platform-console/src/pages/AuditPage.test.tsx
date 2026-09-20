import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { fireEvent, render, screen, waitFor } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it } from "vitest";
import { stubPlatformWire, type WireCall } from "../test-helpers";
import AuditPage from "./AuditPage";

/**
 * 契約：稽核是**後端分頁與後端篩選**（不是把整份軌跡抓到前端再過濾）——篩選與頁碼都必須
 * 進請求參數；每列要看得見「誰（operator）、做了什麼（action）、對誰（target）、為什麼（reason）、
 * 什麼時候」，因為平台寫入的稽核就是查帳的唯一依據。
 */
const ENTRIES = [
  {
    id: "9",
    operatorEmail: "ops@example.com",
    action: "record_payment",
    targetType: "subscription",
    targetId: "5",
    reason: "匯款入帳",
    createdAt: "2026-09-20T10:00:00Z",
  },
  {
    id: "8",
    operatorEmail: "admin@example.com",
    action: "plan.entitlement_set",
    targetType: "plan",
    targetId: "7",
    reason: "年度調價",
    createdAt: "2026-09-19T08:30:00Z",
  },
];

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(() => (
    <QueryClientProvider client={client}>
      <AuditPage />
    </QueryClientProvider>
  ));
}

const auditCalls = (calls: WireCall[]) => calls.filter((c) => c.method === "ListPlatformAudit");
const lastBody = (calls: WireCall[]) => auditCalls(calls).at(-1)?.body;

let calls: WireCall[] = [];

beforeEach(() => {
  calls = stubPlatformWire({
    ListPlatformAudit: () => ({ entries: ENTRIES, pagination: { page: 1, pageSize: 20, total: 41 } }),
  });
});

describe("AuditPage", () => {
  it("每列顯示 operator、動作、目標、原因與時間（表頭＋2 列）", async () => {
    renderPage();

    await screen.findByText("ops@example.com");
    expect(screen.getAllByRole("row")).toHaveLength(3);
    expect(screen.getByText("admin@example.com")).toBeTruthy();
    expect(screen.getByText("匯款入帳")).toBeTruthy();
    expect(screen.getByText("record_payment")).toBeTruthy();
    expect(screen.getByText("subscription：5")).toBeTruthy();
    expect(screen.getByText("2026-09-20T10:00:00Z")).toBeTruthy();
  });

  it("首次載入就帶後端分頁參數（不是前端切頁）", async () => {
    renderPage();

    await waitFor(() => expect(auditCalls(calls)).toHaveLength(1));
    expect(lastBody(calls)).toEqual({ page: 1, pageSize: 20 });
    // 分頁文字被 Solid 的動態節點切開，故比對整段文字。
    await screen.findByText((text) => text === "第 1 / 3 頁（共 41 筆）");
  });

  it("篩選與翻頁都進請求參數（篩選後回到第 1 頁）", async () => {
    renderPage();
    await screen.findByText("ops@example.com");

    fireEvent.input(screen.getByLabelText(/目標類型/), { target: { value: "plan" } });
    fireEvent.input(screen.getByLabelText(/目標代碼/), { target: { value: "7" } });
    fireEvent.click(screen.getByRole("button", { name: "查詢" }));

    await waitFor(() => expect(auditCalls(calls)).toHaveLength(2));
    expect(lastBody(calls)).toEqual({ page: 1, pageSize: 20, targetType: "plan", targetId: "7" });

    fireEvent.click(await screen.findByRole("button", { name: "下一頁" }));
    await waitFor(() => expect(auditCalls(calls)).toHaveLength(3));
    expect(lastBody(calls)).toEqual({ page: 2, pageSize: 20, targetType: "plan", targetId: "7" });

    fireEvent.click(await screen.findByRole("button", { name: "清除" }));
    await waitFor(() => expect(auditCalls(calls)).toHaveLength(4));
    expect(lastBody(calls)).toEqual({ page: 1, pageSize: 20 });
    expect((screen.getByLabelText(/目標類型/) as HTMLInputElement).value).toBe("");
  });

  it("沒有紀錄時顯示空狀態（不出現表格）", async () => {
    stubPlatformWire({
      ListPlatformAudit: () => ({ entries: [], pagination: { page: 1, pageSize: 20, total: 0 } }),
    });
    renderPage();

    await waitFor(() => expect(screen.getByText(/尚無平台操作紀錄/)).toBeTruthy());
    expect(screen.queryByRole("table", { name: "平台稽核紀錄" })).toBeNull();
  });
});
