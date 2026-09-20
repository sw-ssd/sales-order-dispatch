import { describe, expect, it } from "vitest";
import { receivableStatus, receivablesCsv } from "./receivables";

/**
 * 契約：CSV 的組字串是**純函式**（可斷言輸出字串），跳脫照 RFC 4180 ——
 * 逗號、雙引號、換行都會讓一整列歪掉，而這份 CSV 是給營運貼進試算表對帳的。
 * 「逾期」不是後端狀態，是同一列上期末已過的事實（見 ListReceivables 的註解），
 * 故 `now` 由呼叫端注入，狀態才有唯一答案。
 */
const NOW = Date.parse("2026-11-15T00:00:00Z");

const row = {
  companyName: "甲公司",
  planCode: "std",
  periodNo: 2,
  amount: "1500.00",
  periodEnd: "2026-10-31T00:00:00Z",
  status: "open",
};

describe("receivablesCsv", () => {
  it("表頭固定六欄，金額保留後端的兩位小數字串", () => {
    const lines = receivablesCsv([row], NOW).split("\n");
    expect(lines[0]).toBe("公司,方案,期別,金額,到期日,狀態");
    expect(lines[1]).toBe("甲公司,std,2,1500.00,2026-10-31T00:00:00Z,逾期");
  });

  it("逗號／雙引號／換行都跳脫（RFC 4180），欄位數不因內容而位移", () => {
    const csv = receivablesCsv(
      [
        { ...row, companyName: "甲,乙公司", status: "open", periodEnd: "2999-01-01T00:00:00Z" },
        { ...row, companyName: '甲"乙"公司' },
        { ...row, companyName: "甲\n乙公司" },
      ],
      NOW,
    );
    const lines = csv.split("\n");
    expect(lines[1]).toBe('"甲,乙公司",std,2,1500.00,2999-01-01T00:00:00Z,未付');
    expect(lines[2]).toBe('"甲""乙""公司",std,2,1500.00,2026-10-31T00:00:00Z,逾期');
    // 欄位內的換行被包在引號裡：整份 CSV 因此多一個實體行，但那一列的欄位數不變。
    expect(lines).toHaveLength(6);
    expect(csv).toContain('"甲\n乙公司",std,2,1500.00,2026-10-31T00:00:00Z,逾期');
  });

  it("沒有資料時只有表頭（不是空字串）", () => {
    expect(receivablesCsv([], NOW)).toBe("公司,方案,期別,金額,到期日,狀態\n");
  });
});

describe("receivableStatus", () => {
  it("open 且期末已過＝逾期；期末未到＝未付", () => {
    expect(receivableStatus({ status: "open", periodEnd: "2026-10-31T00:00:00Z" }, NOW)).toBe("逾期");
    expect(receivableStatus({ status: "open", periodEnd: "2026-12-31T00:00:00Z" }, NOW)).toBe("未付");
  });

  it("非 open 的狀態原樣呈現（後端才是狀態的定義者）", () => {
    expect(receivableStatus({ status: "paid", periodEnd: "2026-01-01T00:00:00Z" }, NOW)).toBe("paid");
  });
});
