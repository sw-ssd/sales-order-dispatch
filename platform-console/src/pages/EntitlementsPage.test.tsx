import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { render, screen, waitFor } from "@solidjs/testing-library";
import { beforeEach, describe, expect, it } from "vitest";
import { stubPlatformWire, type WireCall } from "../test-helpers";
import EntitlementsPage from "./EntitlementsPage";

/**
 * 契約：權益矩陣是**唯讀**的檢視面（方案權益的寫入在方案頁），格子要能一眼看出
 * 「哪個方案、哪個功能、啟用與否、上限多少」，未設定的功能一律「未開通」（fail-closed），
 * 而 `limit_set=false` 對方案權益的語意是**不限額**（不是上限 0）。
 */
const FEATURES = [
  { code: "limit.seats", type: "integer", unit: "席", description: "席位" },
  { code: "limit.customers", type: "integer", unit: "客戶", description: "客戶" },
  { code: "limit.products", type: "integer", unit: "商品", description: "商品" },
  { code: "limit.departments", type: "integer", unit: "部門", description: "部門" },
  { code: "limit.storage_gb", type: "integer", unit: "GB", description: "空間" },
  { code: "feature.printing", type: "boolean", unit: "", description: "列印" },
  { code: "feature.dispatch", type: "boolean", unit: "", description: "派車" },
  { code: "feature.returns", type: "boolean", unit: "", description: "退貨" },
];

const PLANS = [
  { id: "1", code: "free", name: "免費版", status: "active", sortOrder: 1, prices: [] },
  { id: "2", code: "std", name: "標準版", status: "active", sortOrder: 2, prices: [] },
  { id: "3", code: "pro", name: "專業版", status: "active", sortOrder: 3, prices: [] },
];

/** 每個方案各自被設定的權益（未列出的功能＝未開通）。 */
const BY_PLAN: Record<string, { featureCode: string; enabled: boolean; limitSet: boolean; limitValue: string }[]> = {
  free: [
    { featureCode: "limit.seats", enabled: true, limitSet: true, limitValue: "3" },
    { featureCode: "limit.customers", enabled: true, limitSet: true, limitValue: "50" },
  ],
  std: [
    { featureCode: "limit.seats", enabled: true, limitSet: true, limitValue: "10" },
    // limit_set=false 對方案權益＝不限額。
    { featureCode: "limit.storage_gb", enabled: true, limitSet: false, limitValue: "0" },
    { featureCode: "feature.printing", enabled: false, limitSet: false, limitValue: "0" },
  ],
  pro: [
    { featureCode: "limit.seats", enabled: true, limitSet: true, limitValue: "999" },
    { featureCode: "feature.printing", enabled: true, limitSet: false, limitValue: "0" },
    { featureCode: "feature.dispatch", enabled: true, limitSet: false, limitValue: "0" },
  ],
};

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(() => (
    <QueryClientProvider client={client}>
      <EntitlementsPage />
    </QueryClientProvider>
  ));
}

function matrixTable(): HTMLElement {
  return screen.getByRole("table", { name: "方案權益矩陣" });
}

/** 依功能代碼找列（列上會同時出現說明與代碼，兩者都可比對）。 */
function rowOf(code: string): HTMLElement {
  const rows = screen.getAllByRole("row");
  const row = rows.find((r) => r.textContent?.includes(code));
  if (!row) throw new Error(`找不到 ${code} 的列`);
  return row;
}

/** 第 n 個方案欄（0＝功能欄，1＝第一個方案）。 */
function cellText(row: HTMLElement, column: number): string {
  return row.querySelectorAll("td,th")[column]?.textContent ?? "";
}

let calls: WireCall[] = [];

beforeEach(() => {
  calls = stubPlatformWire({
    ListPlans: () => ({ plans: PLANS }),
    GetPlanEntitlements: (body) => ({
      features: FEATURES,
      entitlements: BY_PLAN[body.planCode as string] ?? [],
    }),
  });
});

describe("EntitlementsPage", () => {
  it("矩陣為 8 列 × 3 方案（含表頭共 9 列），且逐方案取權益", async () => {
    renderPage();

    await waitFor(() => expect(screen.getAllByRole("row")).toHaveLength(9));
    for (const row of screen.getAllByRole("row")) {
      expect(row.querySelectorAll("td,th")).toHaveLength(4);
    }

    // 欄順序＝方案清單順序，每欄真的是各自方案的權益（不是同一份結果貼三次）。
    const header = cellText(screen.getAllByRole("row")[0], 1);
    expect(header).toContain("免費版");
    expect(
      calls.filter((c) => c.method === "GetPlanEntitlements").map((c) => c.body.planCode),
    ).toEqual(["free", "std", "pro"]);
  });

  it("未設定的格顯示未開通；啟用／停用與上限（含不限）逐格呈現", async () => {
    renderPage();
    await waitFor(() => expect(screen.getAllByRole("row")).toHaveLength(9));

    // 席位：免費版 3 席、標準版 10 席、專業版 999 席。
    const seats = rowOf("limit.seats");
    expect(cellText(seats, 1)).toContain("3 席");
    expect(cellText(seats, 2)).toContain("10 席");
    expect(cellText(seats, 3)).toContain("999 席");

    // limit_set=false ＝ 不限額（不是 0）。
    expect(cellText(rowOf("limit.storage_gb"), 2)).toContain("不限");

    // 未設定＝未開通；標準版停用、專業版啟用。
    expect(cellText(rowOf("feature.printing"), 1)).toContain("未開通");
    expect(cellText(rowOf("feature.printing"), 2)).toContain("停用");
    expect(cellText(rowOf("feature.printing"), 3)).toContain("啟用");

    expect(cellText(rowOf("feature.dispatch"), 1)).toContain("未開通");
  });

  it("唯讀：矩陣上沒有任何寫入介面（方案權益的變更在方案頁）", async () => {
    renderPage();
    await waitFor(() => expect(matrixTable()).toBeTruthy());

    expect(screen.queryAllByRole("button")).toHaveLength(0);
    expect(screen.queryAllByRole("textbox")).toHaveLength(0);
    expect(document.querySelectorAll("form")).toHaveLength(0);
  });

  it("沒有方案時顯示空狀態（不畫空矩陣）", async () => {
    stubPlatformWire({ ListPlans: () => ({ plans: [] }) });
    renderPage();

    await waitFor(() => expect(screen.getByText(/尚未定義任何方案/)).toBeTruthy());
    expect(screen.queryByRole("table", { name: "方案權益矩陣" })).toBeNull();
  });
});
