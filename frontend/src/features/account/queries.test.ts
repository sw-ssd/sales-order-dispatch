import { describe, expect, it } from "vitest";
import { TENANT_ENTITLEMENTS_QUERY_KEY, tenantEntitlementsQueryOptions } from "./queries";

/**
 * 未結項 #32：queryKey 是 Banner 與卡片之間的契約 —— 改了它，兩者就各打各的
 * （提示條讓每個頁面多一趟 RPC）。此測試把 key 的字面值與「選項真的用它」一起釘住：
 * 改 key 必須同步改兩處測試 mock（PlanBanner.test.tsx／PlanCard.test.tsx），
 * 否則這裡先紅。
 */
describe("tenantEntitlementsQueryOptions（共用查詢契約）", () => {
  it("queryKey 字面值固定", () => {
    expect([...TENANT_ENTITLEMENTS_QUERY_KEY]).toEqual(["tenant-entitlements"]);
  });

  it("選項用的就是該 key（改 key 不得只改一半）", () => {
    expect(tenantEntitlementsQueryOptions().queryKey).toEqual([
      ...TENANT_ENTITLEMENTS_QUERY_KEY,
    ]);
  });
});
