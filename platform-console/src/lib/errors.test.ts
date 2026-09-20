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
