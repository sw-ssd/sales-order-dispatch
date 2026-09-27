import { describe, expect, it } from "vitest";
import { connectErrorWithInfo } from "../test-helpers";
import { describeError } from "./errors";

/**
 * 契約：`PLAT-3002`（收款衝突）由後端 billing 產生，**同一個碼承載兩種不同的語意**：
 * ① 輸入金額與期別快照金額不符；② 期別已付款但交易號不同（重複收款／溢收）。
 * 後端把語意放在 `ErrorInfo.details.reason` 那段自由文字裡，故行動指引必須照它分流；
 * 而且**兩者都不是「不可重試」**（後端刻意不把該碼當成不可重試，見 platformWriteError 的註解）。
 *
 * 斷言用詞刻意挑後端訊息裡**沒有**的字（「留空」「沿用原交易號」）：否則就算沒有指引，
 * 光是後端訊息本身也會讓斷言變綠，測不出指引有沒有分流。
 */
const AMOUNT_REASON = "輸入金額 3000.00 與期別金額 1500.00 不符（不支援部分付款；差異請記於備註）";
const REF_REASON = '期別 2 已付款（交易號 "TX-1"），本次交易號 "TX-2" 不同；請確認是否重複收款，或改用人工對帳處理溢收';

describe("describeError(PLAT-3002)", () => {
  it("金額不符 → 指引改用期別快照（留空）或依期別金額輸入", () => {
    const text = describeError(
      connectErrorWithInfo("PLAT-3002", { message: `收款衝突：${AMOUNT_REASON}`, details: { reason: AMOUNT_REASON } }),
    );

    expect(text).toContain("PLAT-3002");
    expect(text).toMatch(/留空/);
    expect(text).toContain(AMOUNT_REASON);
  });

  it("期別已付款且交易號不同 → 指引確認是否重複收款／沿用原交易號", () => {
    const text = describeError(
      connectErrorWithInfo("PLAT-3002", { message: `收款衝突：${REF_REASON}`, details: { reason: REF_REASON } }),
    );

    expect(text).toContain("PLAT-3002");
    expect(text).toMatch(/沿用原交易號/);
    expect(text).toContain(REF_REASON);
  });

  // 未結項 #30：kind 鍵優先 —— 有 kind 無 reason 時照樣分流（回退路徑由上兩測覆蓋）。
  it("kind 鍵優先分流：無 reason 時 kind 照樣給出對應指引", () => {
    const amount = describeError(
      connectErrorWithInfo("PLAT-3002", { details: { kind: "amount_mismatch" } }),
    );
    const ref = describeError(
      connectErrorWithInfo("PLAT-3002", { details: { kind: "ref_mismatch" } }),
    );
    const cross = describeError(
      connectErrorWithInfo("PLAT-3002", { details: { kind: "cross_period" } }),
    );
    expect(amount).toMatch(/留空/);
    expect(ref).toMatch(/沿用原交易號/);
    expect(cross).toMatch(/重新整理/);
  });

  it("兩種語意給出不同的行動指引，且都不叫人放棄", () => {
    const amount = describeError(
      connectErrorWithInfo("PLAT-3002", { details: { reason: AMOUNT_REASON } }),
    );
    const ref = describeError(connectErrorWithInfo("PLAT-3002", { details: { reason: REF_REASON } }));

    expect(amount).toMatch(/留空/);
    expect(ref).not.toMatch(/留空/);
    expect(ref).toMatch(/沿用原交易號/);
    expect(amount).not.toMatch(/沿用原交易號/);
    for (const text of [amount, ref]) {
      expect(text).not.toMatch(/不可重試|無法重試|請勿重試/);
    }
  });
});

/**
 * 契約：`SYS-1001`（參數驗證失敗）在平台寫入路徑**一定**帶 `details.field`，指出哪個欄位被拒。
 * 碼表那句「參數驗證失敗」對 operator 沒有行動資訊 —— 不讀 `field` 的話，他只看得到
 * 「SYS-1001：參數驗證失敗」然後自己猜是哪一格。
 */
describe("describeError(SYS-1001)", () => {
  it("帶已知 field → 說出是哪個欄位（顯示中文標籤，不是後端欄位名）", () => {
    const text = describeError(
      connectErrorWithInfo("SYS-1001", { message: "參數驗證失敗", details: { field: "seat_count" } }),
    );

    expect(text).toContain("席位數");
    // 內部欄位名不得攤到 operator 面前。
    expect(text).not.toContain("seat_count");
  });

  it("帶未知 field → 通用說法，不硬湊翻譯也不吞掉訊息", () => {
    const text = describeError(
      connectErrorWithInfo("SYS-1001", { message: "參數驗證失敗", details: { field: "zzz_unknown" } }),
    );

    expect(text).toContain("SYS-1001");
    expect(text).toContain("參數驗證失敗");
    expect(text).toContain("請檢查送出的欄位內容");
  });
});
