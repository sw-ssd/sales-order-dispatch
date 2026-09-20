import { describe, expect, it } from "vitest";
import { projectTenantEntitlements } from "./entitlements";

/**
 * 投影＝後端 `entitlements.resolveFeature` 的**唯讀**版本：方案 ⊕ 未逾期的例外。
 *
 * 這些斷言必須真的會紅：把投影改成「方案值直接當生效值」（拿掉例外覆寫）或「不過濾過期例外」，
 * 下面的測試都會失敗。**最終判定仍在後端**（訂閱狀態、用量都不是這裡算的），
 * 畫面只呈現「設定值」，並在頁面上明說這件事。
 */
const features = [
  { code: "limit.seats", type: "integer", unit: "席", description: "席位" },
  { code: "feature.printing", type: "boolean", unit: "", description: "列印" },
];

const planEntitlements = [
  { featureCode: "limit.seats", enabled: true, limitSet: true, limitValue: 10n },
  { featureCode: "feature.printing", enabled: false, limitSet: false, limitValue: 0n },
];

const now = new Date("2026-09-21T00:00:00Z");

describe("projectTenantEntitlements", () => {
  it("未設定權益的功能一律視為未開通（fail-closed，不是放行）", () => {
    const rows = projectTenantEntitlements({ features, entitlements: [], overrides: [], now });

    expect(rows.map((r) => r.featureCode)).toEqual(["limit.seats", "feature.printing"]);
    expect(rows.every((r) => !r.enabled)).toBe(true);
    expect(rows.every((r) => !r.planConfigured)).toBe(true);
  });

  it("例外覆寫方案值，且只覆寫它有指定的維度", () => {
    const rows = projectTenantEntitlements({
      features,
      entitlements: planEntitlements,
      overrides: [
        // 只給限額：不得把 enabled 一起改掉（後端以 NULL 表達「此維度不覆寫」）。
        {
          featureCode: "limit.seats",
          enabledSet: false,
          enabled: false,
          limitSet: true,
          limitValue: 30n,
          reason: "簽約承諾",
          owner: "ops@example.com",
          expiresAt: "2027-01-01T00:00:00Z",
        },
        // 只給 enabled：限額維度維持方案的設定。
        {
          featureCode: "feature.printing",
          enabledSet: true,
          enabled: true,
          limitSet: false,
          limitValue: 0n,
          reason: "POC",
          owner: "ops@example.com",
          expiresAt: "",
        },
      ],
      now,
    });

    const seats = rows.find((r) => r.featureCode === "limit.seats");
    expect(seats?.planLimitValue).toBe(10n);
    expect(seats?.limitValue).toBe(30n);
    expect(seats?.enabled).toBe(true);
    expect(seats?.override?.reason).toBe("簽約承諾");

    const printing = rows.find((r) => r.featureCode === "feature.printing");
    expect(printing?.enabled).toBe(true);
    expect(printing?.limitSet).toBe(false);
    expect(printing?.override?.owner).toBe("ops@example.com");
  });

  it("已過期的例外不生效（生效值回到方案值，來源標為未生效）", () => {
    const rows = projectTenantEntitlements({
      features,
      entitlements: planEntitlements,
      overrides: [
        {
          featureCode: "limit.seats",
          enabledSet: false,
          enabled: false,
          limitSet: true,
          limitValue: 99n,
          reason: "已到期的優惠",
          owner: "ops@example.com",
          expiresAt: "2026-09-01T00:00:00Z",
        },
      ],
      now,
    });

    const seats = rows.find((r) => r.featureCode === "limit.seats");
    expect(seats?.limitValue).toBe(10n);
    expect(seats?.override).toBeUndefined();
  });

  it("limit_set=false 的例外是「不覆寫限額」，不是「不限額」", () => {
    const rows = projectTenantEntitlements({
      features,
      entitlements: planEntitlements,
      overrides: [
        {
          featureCode: "limit.seats",
          enabledSet: true,
          enabled: false,
          limitSet: false,
          limitValue: 0n,
          reason: "關掉席位管理",
          owner: "ops@example.com",
          expiresAt: "",
        },
      ],
      now,
    });

    const seats = rows.find((r) => r.featureCode === "limit.seats");
    expect(seats?.enabled).toBe(false);
    expect(seats?.limitSet).toBe(true);
    expect(seats?.limitValue).toBe(10n);
  });

  it("同一功能多筆未逾期的例外：後一筆勝（與後端逐筆覆寫的順序一致）", () => {
    const rows = projectTenantEntitlements({
      features,
      entitlements: planEntitlements,
      overrides: [
        {
          featureCode: "limit.seats",
          enabledSet: false,
          enabled: false,
          limitSet: true,
          limitValue: 20n,
          reason: "第一次承諾",
          owner: "a@example.com",
          expiresAt: "",
        },
        {
          featureCode: "limit.seats",
          enabledSet: false,
          enabled: false,
          limitSet: true,
          limitValue: 40n,
          reason: "第二次承諾",
          owner: "b@example.com",
          expiresAt: "",
        },
      ],
      now,
    });

    const seats = rows.find((r) => r.featureCode === "limit.seats");
    expect(seats?.limitValue).toBe(40n);
    expect(seats?.override?.reason).toBe("第二次承諾");
  });
});
