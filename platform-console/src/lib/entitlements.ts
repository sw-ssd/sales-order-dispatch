/**
 * 租戶權益的**唯讀投影**：方案權益 ⊕ 租戶例外（override）。
 *
 * 這裡**不是判定層**。真正的決策在後端 `internal/platform/entitlements.resolveFeature`：
 * 訂閱狀態（非可用狀態一律拒絕、試用中一律開放）、用量計數、快取，都是後端的事。
 * 本檔只把「設定值」攤開給 operator 看——尤其**已逾期的例外不生效**這件事，
 * 不投影出來的話，營運會看到一筆躺在清單裡的承諾而誤以為它還在生效。
 *
 * **同功能多筆例外時，這裡算不出後端的答案**（`ambiguous`）：
 * 後端判定用的例外順序是 `Store.Overrides` 的 `ORDER BY created_at DESC` 逐筆覆寫（最後一筆＝最舊者勝），
 * 而 console 拿到的清單是 `Admin.GetTenant` 的 `ORDER BY feature_code`，且 proto 的 `TenantOverride`
 * **沒有 created_at**（不為此加欄位）。前端因此不可能重建判定順序 —— 這種情況一律標示
 * `ambiguous`，由畫面說「最終以後端判定為準」，**不靜默挑一筆當答案**。
 * 防禦性標示：`tenant_overrides_active_unique`（`WHERE revoked_at IS NULL`）保證同功能至多
 * 一筆未撤銷例外，故真實部署中此分支不可達 —— 但若該約束被放寬或資料繞過約束寫入，
 * 畫面仍不會靜默給出錯誤答案。
 *
 * 與後端一致的兩條語意（少一條就會顯示錯的值）：
 * 1. `limit_set=false` 對**例外**是「不覆寫限額」，對**方案權益**是「不限額」——proto 的註解如此定義。
 * 2. 未設定的功能一律視為未開通（fail-closed），不是放行。
 *
 * 型別刻意寫成結構型（不引用生成的 Message 型別）：生成的訊息多帶欄位仍可指派進來，
 * 而這裡只需要這幾個欄位就能算出投影。
 */

/** 功能定義（`Feature`）。 */
export type Feature = {
  code: string;
  type?: string;
  unit?: string;
  description?: string;
};

/** 方案對某功能的權益（`FeatureEntitlement`）：這裡的 `limitSet=false` ＝ 不限額。 */
export type FeatureEntitlement = {
  featureCode: string;
  enabled: boolean;
  limitSet: boolean;
  limitValue: bigint;
};

/** 租戶例外（`TenantOverride`）：這裡的 `limitSet=false` ＝ 不覆寫限額維度。 */
export type TenantOverride = {
  featureCode: string;
  enabledSet: boolean;
  enabled: boolean;
  limitSet: boolean;
  limitValue: bigint;
  reason?: string;
  owner?: string;
  expiresAt?: string;
};

export type ProjectedEntitlement = {
  featureCode: string;
  description: string;
  /** 方案是否有這一格的設定（沒有＝未開通）。 */
  planConfigured: boolean;
  planEnabled: boolean;
  planLimitSet: boolean;
  planLimitValue: bigint;
  /** 生效中的例外（未逾期）；已逾期的例外不列入。同功能多筆時是最後被套用的一筆，見 `ambiguous`。 */
  override?: TenantOverride;
  /** 未逾期的例外筆數（同一功能）。 */
  overrideCount: number;
  /** 同一功能有 >1 筆未逾期的例外＝前端重建不出後端的判定順序，畫面必須說出「以後端為準」。 */
  ambiguous: boolean;
  /** 方案 ⊕ 例外後的設定值（`ambiguous` 時只是一種可能，不是答案）。 */
  enabled: boolean;
  limitSet: boolean;
  limitValue: bigint;
};

/**
 * 例外是否已逾期。`expires_at` 空字串＝不過期（proto 的約定）。
 *
 * 用前端時鐘判斷有幾秒級的偏差：這只影響「畫面顯示」，不影響任何決定——
 * 後端以伺服器時間為準（`resolveFeature` 也是同一條規則）。
 */
export function isExpired(expiresAt: string | undefined, now: Date): boolean {
  if (!expiresAt) return false;
  const at = new Date(expiresAt);
  return !Number.isNaN(at.getTime()) && at <= now;
}

/**
 * 逐功能算出投影。
 *
 * 同一功能有多筆未逾期的例外時，套用順序只是「清單順序」——**那不是後端判定的順序**（見檔頭），
 * 故標示 `ambiguous`；呼叫端必須把這件事顯示出來，不能拿算出來的值當答案。
 */
export function projectTenantEntitlements(input: {
  features: Feature[];
  entitlements: FeatureEntitlement[];
  overrides: TenantOverride[];
  /** 可注入，測試用；預設為現在。 */
  now?: Date;
}): ProjectedEntitlement[] {
  const now = input.now ?? new Date();

  return input.features.map((feature) => {
    const plan = input.entitlements.find((e) => e.featureCode === feature.code);

    let enabled = plan?.enabled ?? false;
    let limitSet = plan?.limitSet ?? false;
    let limitValue = plan?.limitValue ?? 0n;
    let applied: TenantOverride | undefined;
    let appliedCount = 0;

    for (const override of input.overrides) {
      if (override.featureCode !== feature.code) continue;
      if (isExpired(override.expiresAt, now)) continue;
      if (override.enabledSet) enabled = override.enabled;
      if (override.limitSet) {
        limitSet = true;
        limitValue = override.limitValue;
      }
      applied = override;
      appliedCount += 1;
    }

    return {
      featureCode: feature.code,
      description: feature.description ?? "",
      planConfigured: plan !== undefined,
      planEnabled: plan?.enabled ?? false,
      planLimitSet: plan?.limitSet ?? false,
      planLimitValue: plan?.limitValue ?? 0n,
      override: applied,
      overrideCount: appliedCount,
      ambiguous: appliedCount > 1,
      enabled,
      limitSet,
      limitValue,
    };
  });
}
