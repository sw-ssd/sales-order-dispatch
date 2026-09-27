import { describe, expect, it } from "vitest";
import { EMPTY_DATETIME, formatDate, formatDateTime } from "./datetime";

/**
 * 這一組測的核心是「**印出的值不得隨輸入的時區偏移改變**」。
 * 後端有 5 個服務輸出 `...Z`、15 個輸出 `+08:00`（見 datetime.ts 的說明），
 * `slice(0, 19)` 會讓同一個瞬間差八小時；這裡把那件事釘死。
 */
describe("formatDateTime", () => {
  it("同一個瞬間不論後端用 Z 還是 +08:00 序列化，印出來完全相同", () => {
    const instant = "2026-09-24T17:43:09Z";
    const sameInstant = "2026-09-25T01:43:09+08:00";
    expect(formatDateTime(instant)).toBe(formatDateTime(sameInstant));
    // 營業時區 UTC+8：17:43Z 是隔天凌晨 01:43。
    expect(formatDateTime(instant)).toBe("2026-09-25 01:43:09");
  });

  it("帶負偏移也正確（後端若改以其他偏移序列化不會走鐘）", () => {
    expect(formatDateTime("2026-09-24T12:43:09-05:00")).toBe("2026-09-25 01:43:09");
  });

  it("輸出的格式固定為 YYYY-MM-DD HH:mm:ss（切秒、T 換空白）", () => {
    expect(formatDateTime("2026-01-05T02:03:04Z")).toBe("2026-01-05 10:03:04");
  });

  it("跨日與跨月跨年邊界以營業時區判定", () => {
    // 16:00Z = 隔天 00:00(+08)。若誤用 UTC 會顯示成前一天。
    expect(formatDateTime("2026-12-31T16:00:00Z")).toBe("2027-01-01 00:00:00");
    expect(formatDateTime("2026-12-31T15:59:59Z")).toBe("2026-12-31 23:59:59");
  });

  it("空值與無法解析的值顯示佔位字元，不丟錯也不顯示 NaN", () => {
    expect(formatDateTime(undefined)).toBe(EMPTY_DATETIME);
    expect(formatDateTime(null)).toBe(EMPTY_DATETIME);
    expect(formatDateTime("")).toBe(EMPTY_DATETIME);
    expect(formatDateTime("not-a-date")).toBe(EMPTY_DATETIME);
  });

  it("formatDate 取營業時區的日期（瞬間跨日時不可用 UTC 的日期）", () => {
    expect(formatDate("2026-12-31T16:00:00Z")).toBe("2027-01-01");
    expect(formatDate(undefined)).toBe(EMPTY_DATETIME);
  });
});
