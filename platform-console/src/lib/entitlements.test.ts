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

  it("同一功能多筆未逾期的例外：標示為不明確，不宣稱「生效值＝後端判定」", () => {
    // 後端判定的例外順序是 `ORDER BY created_at DESC`（store.go）＋迴圈後者勝＝**最舊者勝**，
    // 而 console 拿到的清單是 `ORDER BY feature_code`（admin.go）、proto 的 TenantOverride 沒有
    // created_at（也不得為此加欄位）→ 前端**不可能**重建判定順序。因此這裡只標示「有多筆」，
    // 由畫面說出「最終以後端判定為準」，不靜默挑一筆當答案。
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
    expect(seats?.ambiguous).toBe(true);
    expect(seats?.overrideCount).toBe(2);
    // 只有一筆（或零筆）時不得無故示警。
    const printing = rows.find((r) => r.featureCode === "feature.printing");
    expect(printing?.ambiguous).toBe(false);
    expect(printing?.overrideCount).toBe(0);
  });

  it("恰一筆未逾期的例外：ambiguous=false，且投影值＝該筆覆寫後的結果", async () => {
    // 未結項 #29：真實部署中同功能至多一筆未撤銷例外（tenant_overrides_active_unique），
    // 「恰一筆」才是常態路徑 —— 必須斷言它不示警、且值真的被覆寫（否則常態路徑無覆蓋）。
    const { projectTenantEntitlements: project } = await import("./entitlements");
    const rows = project({
      features,
      entitlements: planEntitlements,
      overrides: [
        {
          featureCode: "limit.seats",
          enabledSet: false,
          enabled: false,
          limitSet: true,
          limitValue: 40n,
          reason: "唯一承諾",
          owner: "b@example.com",
          expiresAt: "",
        },
      ],
      now,
    });
    const seats = rows.find((r) => r.featureCode === "limit.seats");
    expect(seats?.ambiguous).toBe(false);
    expect(seats?.overrideCount).toBe(1);
    expect(seats?.limitSet).toBe(true);
    expect(seats?.limitValue).toBe(40n);
    expect(seats?.override?.reason).toBe("唯一承諾");
  });
});
