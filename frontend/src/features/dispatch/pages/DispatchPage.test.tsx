import { fireEvent, render, screen, waitFor, within } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import { ErrorInfoSchema } from "~/lib/proto/salesorder/v1/common_pb";
import type * as ConnectRpc from "@connectrpc/connect";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

// 派車／訂單／車次 API 以 spy 取代：頁面在模組層建立 connect client，
// 以 createClient 的替身同時攔截（ConnectError/Code 保持真實，錯誤訊息對照才有效）。
const {
  listOrdersSpy,
  listRoutesSpy,
  assignRouteSpy,
  confirmDispatchSpy,
  cancelDispatchSpy,
  watchBoardSpy,
} = vi.hoisted(() => {
  // 看板訂閱：一個只 yield 一則事件就結束的 async iterable（頁面 for-await 會因此自然結束）。
  const boardEvent = {
    type: "route_assign",
    salesOrderId: "1",
    routeId: "r-1",
    deliverySequence: "1",
    version: "3",
    departmentId: "d-1",
  };
  return {
    listOrdersSpy: vi.fn(),
    listRoutesSpy: vi.fn(),
    assignRouteSpy: vi.fn(),
    confirmDispatchSpy: vi.fn(),
    cancelDispatchSpy: vi.fn(),
    watchBoardSpy: vi.fn(async function* () {
      yield boardEvent;
    }),
  };
});

vi.mock("@connectrpc/connect", async (importOriginal) => ({
  ...(await importOriginal<typeof ConnectRpc>()),
  createClient: () => ({
    listOrders: listOrdersSpy,
    listRoutes: listRoutesSpy,
    assignRoute: assignRouteSpy,
    confirmDispatch: confirmDispatchSpy,
    cancelDispatch: cancelDispatchSpy,
    watchBoard: watchBoardSpy,
  }),
}));

import DispatchPage, { boardDensity, streamRetryDelay } from "./DispatchPage";

const PENDING_UNASSIGNED = {
  id: "1",
  orderNo: "W000001",
  source: "W",
  status: "pending",
  expectedDeliveryDate: "2026-07-20",
  routeId: "",
  deliverySequence: 0,
  version: 1,
};

const PENDING_IN_ROUTE = {
  id: "2",
  orderNo: "W000002",
  source: "W",
  status: "pending",
  expectedDeliveryDate: "2026-07-20",
  routeId: "r-1",
  deliverySequence: 1,
  version: 2,
};

const PROCESSING_IN_ROUTE = {
  id: "3",
  orderNo: "W000003",
  source: "W",
  status: "processing",
  expectedDeliveryDate: "2026-07-20",
  routeId: "r-1",
  deliverySequence: 2,
  version: 3,
};

/** 不該上看板的狀態（spec：僅 pending 或 processing）。 */
const COMPLETED = {
  id: "4",
  orderNo: "W000004",
  source: "W",
  status: "completed",
  expectedDeliveryDate: "2026-07-20",
  routeId: "r-1",
  deliverySequence: 3,
  version: 4,
};

const ROUTE = {
  id: "r-1",
  code: "R1",
  name: "一號路線",
  description: "",
  sortOrder: 1,
  isActive: true,
  companyId: "c-1",
  departmentId: "d-1",
  createdAt: "",
  updatedAt: "",
  deletedAt: "",
};

/** 每個測試一份全新的 `QueryClient`：快取不跨測試殘留（retry 關閉，理由同 CustomersPage）。 */
function newClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function mountPage(client: QueryClient = newClient()) {
  render(() => (
    <QueryClientProvider client={client}>
      <DispatchPage />
    </QueryClientProvider>
  ));
  return client;
}

async function renderPage() {
  mountPage();
  // 看板預設「今天」，fixture 釘在 2026-07-20 → 先切日期。
  // 這一步同時驗證 spec 的「依日期篩選檢視看板」：改日期即以新日期重查。
  const dateInput = await waitFor(() => screen.getByLabelText("預計出貨日"));
  fireEvent.change(dateInput, { target: { value: "2026-07-20" } });
  await waitFor(() =>
    expect(listOrdersSpy).toHaveBeenCalledWith(
      expect.objectContaining({ expectedDeliveryDate: "2026-07-20" })
    )
  );
  // 送到請求 ≠ 已渲染：新日期 key 落地前看板是 pending（`<Show>` 會藏起看板），要等到內容回來。
  await waitFor(() => expect(screen.getByText("W000001")).toBeTruthy());
}

beforeEach(() => {
  vi.clearAllMocks();
  listOrdersSpy.mockResolvedValue({
    orders: [PENDING_UNASSIGNED, PENDING_IN_ROUTE, PROCESSING_IN_ROUTE, COMPLETED],
    total: 4,
  });
  listRoutesSpy.mockResolvedValue({
    routes: [ROUTE],
    pagination: { total: 1 },
  });
  assignRouteSpy.mockResolvedValue({
    salesOrderId: "1",
    routeId: "r-1",
    deliverySequence: "1",
    version: "2",
  });
  confirmDispatchSpy.mockResolvedValue({
    items: [{ salesOrderId: "2", success: true, failReason: "" }],
    successCount: 1,
  });
  cancelDispatchSpy.mockResolvedValue({
    salesOrderId: "3",
    status: "pending",
    routeId: "r-1",
    deliverySequence: "2",
    reprintWarning: false,
  });
});

describe("DispatchPage", () => {
  it("看板依日期查詢：expectedDeliveryDate 進請求，且只顯示 pending/processing", async () => {
    await renderPage();
    expect(listOrdersSpy).toHaveBeenCalledWith(
      expect.objectContaining({ expectedDeliveryDate: "2026-07-20", pageSize: 100, status: "" })
    );
    // pending/processing 上板。
    expect(screen.getByText("W000001")).toBeTruthy();
    expect(screen.getByText("W000003")).toBeTruthy();
    // completed 不上板。
    expect(screen.queryByText("W000004")).toBeNull();
  });

  it("未指派與車次欄分流，卡片依順位排序", async () => {
    await renderPage();
    const unassigned = screen.getByLabelText("未指派");
    const routeCol = screen.getByLabelText("一號路線");
    expect(within(unassigned).getByText("W000001")).toBeTruthy();
    expect(within(routeCol).getByText("W000002")).toBeTruthy();
    expect(within(routeCol).getByText("W000003")).toBeTruthy();
    // 順位升冪：1 號卡在 2 號卡之前。
    const texts = within(routeCol)
      .getAllByText(/^W00000\d$/)
      .map((n) => n.textContent);
    expect(texts).toEqual(["W000002", "W000003"]);
  });

  it("密度門檻：comfortable 放得下就用，放不下降 compact", () => {
    // 寬度未量到（ResizeObserver 尚未回報）→ 預設 comfortable，不縮欄。
    expect(boardDensity(2, 0)).toBe("comfortable");
    // 2 欄 = 2*288+16 = 592 剛好放得下 → comfortable；差 1px 就降 compact。
    expect(boardDensity(2, 592)).toBe("comfortable");
    expect(boardDensity(2, 591)).toBe("compact");
    // 8 欄需 8*288+7*16 = 2416，2000px 放不下 → compact。
    expect(boardDensity(8, 2000)).toBe("compact");
  });

  it("退回未指派：assignRoute 帶空車次、讀取時的 version 與日期，並失效看板", async () => {
    await renderPage();
    const routeCol = screen.getByLabelText("一號路線");
    fireEvent.click(within(routeCol).getByRole("button", { name: "退回未指派" }));
    await waitFor(() => expect(assignRouteSpy).toHaveBeenCalledOnce());
    expect(assignRouteSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        salesOrderId: "2",
        routeId: "",
        deliverySequence: "",
        version: "2",
        expectedDeliveryDate: "2026-07-20",
      })
    );
    // 樂觀鎖失敗也要回伺服器最新狀態 → 必須重查（spec：衝突後重新整理看板）。
    await waitFor(() => expect(listOrdersSpy.mock.calls.length).toBeGreaterThan(1));
  });

  it("樂觀鎖失敗：顯示後端訊息且仍重查看板", async () => {
    // 後端樂觀鎖失敗是 errcode 錯誤：connect 碼 InvalidArgument、可行動的原因放在
    // ErrorInfo.details.reason（訊息本體只有樣板「參數驗證失敗」）。前端必須顯示 reason。
    assignRouteSpy.mockRejectedValue(
      new ConnectError("參數驗證失敗", Code.InvalidArgument, undefined, [
        {
          desc: ErrorInfoSchema,
          value: {
            code: "SYS-1001",
            message: "參數驗證失敗",
            details: { reason: "資料已變更，請重新載入" },
          },
        },
      ])
    );
    await renderPage();
    const routeCol = screen.getByLabelText("一號路線");
    fireEvent.click(within(routeCol).getByRole("button", { name: "退回未指派" }));
    await waitFor(() => expect(screen.getByRole("alert")).toBeTruthy());
    expect(screen.getByRole("alert").textContent).toContain("資料已變更，請重新載入");
    await waitFor(() => expect(listOrdersSpy.mock.calls.length).toBeGreaterThan(1));
  });

  it("批次確認：confirmDispatch 帶車次與日期，成功筆數顯示", async () => {
    await renderPage();
    const routeCol = screen.getByLabelText("一號路線");
    fireEvent.click(within(routeCol).getByRole("button", { name: "批次確認派車" }));
    await waitFor(() => expect(confirmDispatchSpy).toHaveBeenCalledOnce());
    expect(confirmDispatchSpy).toHaveBeenCalledWith(
      expect.objectContaining({ routeId: "r-1", expectedDeliveryDate: "2026-07-20" })
    );
    await waitFor(() => expect(screen.getByRole("status").textContent).toContain("成功 1 筆"));
  });

  it("批次確認部分失敗：逐筆帶出失敗原因", async () => {
    confirmDispatchSpy.mockResolvedValue({
      items: [
        { salesOrderId: "2", success: false, failReason: "狀態已變更" },
        { salesOrderId: "3", success: true, failReason: "" },
      ],
      successCount: 1,
    });
    await renderPage();
    const routeCol = screen.getByLabelText("一號路線");
    fireEvent.click(within(routeCol).getByRole("button", { name: "批次確認派車" }));
    await waitFor(() => expect(screen.getByRole("status").textContent).toContain("失敗 1 筆"));
    expect(screen.getByRole("status").textContent).toContain("狀態已變更");
  });

  it("取消派車空原因不打 API", async () => {
    await renderPage();
    const routeCol = screen.getByLabelText("一號路線");
    fireEvent.click(within(routeCol).getByRole("button", { name: "取消派車" }));
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    fireEvent.click(within(dialog).getByRole("button", { name: "確認取消" }));
    await waitFor(() => expect(screen.getByRole("alert").textContent).toContain("需填原因"));
    expect(cancelDispatchSpy).not.toHaveBeenCalled();
  });

  it("重印警告：第一次不帶確認旗且不關窗，確認後第二次帶旗送出", async () => {
    await renderPage();
    const routeCol = screen.getByLabelText("一號路線");
    fireEvent.click(within(routeCol).getByRole("button", { name: "取消派車" }));
    const dialog = await waitFor(() => screen.getByRole("dialog"));
    fireEvent.input(within(dialog).getByLabelText(/取消原因/), {
      target: { value: "司機臨時請假" },
    });

    cancelDispatchSpy.mockResolvedValue({
      salesOrderId: "3",
      status: "processing",
      routeId: "r-1",
      deliverySequence: "2",
      reprintWarning: true,
    });
    fireEvent.click(within(dialog).getByRole("button", { name: "確認取消" }));
    await waitFor(() => expect(cancelDispatchSpy).toHaveBeenCalledOnce());
    expect(cancelDispatchSpy).toHaveBeenCalledWith(
      expect.objectContaining({ reason: "司機臨時請假", acknowledgeReprint: false })
    );
    // 警告出現且對話框不關（尚未確認重印）。
    expect(within(dialog).getByText(/已列印/)).toBeTruthy();

    cancelDispatchSpy.mockResolvedValue({
      salesOrderId: "3",
      status: "pending",
      routeId: "r-1",
      deliverySequence: "2",
      reprintWarning: false,
    });
    fireEvent.click(within(dialog).getByRole("button", { name: "確認取消並重印" }));
    await waitFor(() => expect(cancelDispatchSpy).toHaveBeenCalledTimes(2));
    expect(cancelDispatchSpy).toHaveBeenLastCalledWith(
      expect.objectContaining({ acknowledgeReprint: true })
    );
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });

  it("串流事件到達即使看板失效並全量重查（事件僅作失效提示）", async () => {
    await renderPage();
    await waitFor(() => expect(watchBoardSpy).toHaveBeenCalled());
    // 訂閱綁「當前」看板日期：翻日要重新訂閱到新日期（舊訂閱由 cleanup 中止）。
    expect(watchBoardSpy).toHaveBeenLastCalledWith(
      expect.objectContaining({ expectedDeliveryDate: "2026-07-20" }),
      expect.anything()
    );
    // 事件 yield → invalidate → 重查。
    await waitFor(() => expect(listOrdersSpy.mock.calls.length).toBeGreaterThan(1));
  });

  it("重連退避有上限（spec Task 5.2 Step 4：指數退避＋降級輪詢）", () => {
    // 沒有上限的指數退避會在幾分鐘內把間隔撐到小時級，等於實質停止重連。
    expect(streamRetryDelay(1)).toBe(1_000);
    expect(streamRetryDelay(2)).toBe(2_000);
    expect(streamRetryDelay(5)).toBe(16_000);
    expect(streamRetryDelay(6)).toBe(30_000);
    expect(streamRetryDelay(50)).toBe(30_000);
  });

  it("串流連續失敗後降級輪詢，並在頁首說明（不讓看板靜靜變舊）", async () => {
    // 假時鐘驅動退避（1s→2s→4s），真時鐘要等 7 秒才到降級。
    // 這裡不用 `waitFor`：它在假時鐘下與 TanStack 的計時器互相等待，改成自己推進時鐘。
    vi.useFakeTimers();
    try {
      watchBoardSpy.mockImplementation(async function* () {
        throw new Error("watch board down");
      });
      mountPage();
      await vi.advanceTimersByTimeAsync(0); // 首批查詢（純微任務）落地
      // 尚未達門檻：單次斷線由重連處理，不該驚動使用者。
      expect(screen.queryByText(/即時更新不可用/)).toBeNull();
      for (const backoff of [1_000, 2_000, 4_000]) await vi.advanceTimersByTimeAsync(backoff);
      expect(screen.getByText(/即時更新不可用/)).toBeTruthy();
      // 降級後仍持續重連（不因降級而停止嘗試回到即時）。
      expect(watchBoardSpy.mock.calls.length).toBeGreaterThanOrEqual(4);
    } finally {
      vi.useRealTimers();
    }
  });

  it("重連成功收到訊息後自動收回降級提示", async () => {
    vi.useFakeTimers();
    try {
      let attempts = 0;
      watchBoardSpy.mockImplementation(async function* () {
        attempts += 1;
        if (attempts <= 3) throw new Error("watch board down");
        // 第 4 次連上：後端訂閱後第一個位元組是事件或 heartbeat，收到即算恢復。
        yield {
          type: "heartbeat",
          salesOrderId: "",
          routeId: "",
          deliverySequence: "",
          version: "",
          departmentId: "68",
        };
      });
      mountPage();
      await vi.advanceTimersByTimeAsync(0);
      // t=0 失敗、t=1s 失敗、t=3s 第三次失敗 → 降級（第 4 次重連排在 t=7s）。
      await vi.advanceTimersByTimeAsync(1_000);
      await vi.advanceTimersByTimeAsync(2_000);
      expect(screen.getByText(/即時更新不可用/)).toBeTruthy();
      // t=7s 第 4 次連上並收到 heartbeat → 提示收回、恢復即時模式。
      await vi.advanceTimersByTimeAsync(4_000);
      expect(screen.queryByText(/即時更新不可用/)).toBeNull();
    } finally {
      vi.useRealTimers();
    }
  });

  it("非拖曳指派：未指派卡片的下拉選車次即指派到該欄尾", async () => {
    // 觸控（倉庫 iPad）與純鍵盤都沒有 HTML5 拖放事件，下拉是唯一的派車入口；
    // 這條釘住它送出的順位與拖到欄尾同一語意（該欄最大順位 +1）。
    await renderPage();
    const unassigned = screen.getByLabelText("未指派");
    const picker = within(unassigned).getByLabelText("指派 W000001 到車次");
    fireEvent.change(picker, { target: { value: "r-1" } });
    await waitFor(() => expect(assignRouteSpy).toHaveBeenCalledOnce());
    expect(assignRouteSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        salesOrderId: "1",
        routeId: "r-1",
        // 該欄已有順位 1（W000002）與 2（W000003）→ 新卡接在 3，與拖到欄尾同語意。
        deliverySequence: "3",
        version: "1",
        expectedDeliveryDate: "2026-07-20",
      })
    );
  });

  it("非拖曳指派：未選車次（空值）不打 API", async () => {
    await renderPage();
    const picker = within(screen.getByLabelText("未指派")).getByLabelText("指派 W000001 到車次");
    fireEvent.change(picker, { target: { value: "" } });
    // 給事件迴圈機會跑（若誤觸發，這裡就會看到呼叫）。
    const { promise, resolve } = Promise.withResolvers<void>();
    setTimeout(resolve, 0);
    await promise;
    expect(assignRouteSpy).not.toHaveBeenCalled();
  });

  it("無權限時顯示 banner", async () => {
    listOrdersSpy.mockRejectedValue(new ConnectError("denied", Code.PermissionDenied));
    mountPage();
    await waitFor(() => expect(screen.getByRole("alert")).toBeTruthy());
    expect(screen.getByRole("alert").textContent).toContain("沒有權限");
  });
});
