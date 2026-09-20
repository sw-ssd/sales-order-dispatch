import type { GetTenantEntitlementsResponse, Usage } from "~/lib/proto/platform/v1/platform_pb";

/**
 * 權益投影（`GetTenantEntitlements`）的**呈現契約**：文案、門檻與格式化都只在此處定義，
 * 卡片（`PlanCard`）與 shell 的提示條（`PlanBanner`）共用同一份真相。
 *
 * 這裡全是唯讀的呈現；能不能用一律由後端判定（spec §4.6「UI 投影與守衛的分工」）。
 */

/**
 * 用量提示門檻（spec §2.4 規則 2：「用量達 80/90%」）：達 80% 是一般提醒、達 90% 升級為急。
 */
const USAGE_WARN_RATIO = 0.8;
const USAGE_URGENT_RATIO = 0.9;

/**
 * 試用到期提示視窗（天）。spec 只寫「試用將到期」，沒定天數——取 3 天：試用預設 14 天，
 * 只在尾端才吵。要調整門檻只改這一個常數。
 */
const TRIAL_ALERT_DAYS = 3;

/** 提示的嚴重度（決定色階）；文字一律由此模組產生，兩個呼叫端不各自造句。 */
export type PlanAlertSeverity = "warning" | "urgent";

export interface PlanAlert {
  severity: PlanAlertSeverity;
  text: string;
}

/** 提示的色階（語意 token，深淺色自動翻轉）：卡片與提示條共用，不各自配色。 */
export const ALERT_CLASS: Record<PlanAlertSeverity, string> = {
  warning: "border-warning/30 bg-warning/10 text-warning",
  urgent: "border-destructive/30 bg-destructive/10 text-destructive",
};

/**
 * 訂閱狀態文案：與平台營運工具同一套字面值（`trialing|active|past_due|suspended|cancelled|none`）
 * 與同一套文案，兩個 surface 對同一份狀態不該說不同的話。
 */
const STATUS: Record<string, { label: string; variant: "success" | "info" | "warning" | "destructive" | "secondary" }> = {
  trialing: { label: "試用中", variant: "info" },
  active: { label: "訂閱中", variant: "success" },
  past_due: { label: "逾期未付", variant: "warning" },
  suspended: { label: "已停用", variant: "destructive" },
  cancelled: { label: "已取消", variant: "secondary" },
  none: { label: "無訂閱", variant: "secondary" },
};

/** 未知狀態照後端字面值顯示（不猜語意），色階用中性。 */
export function subscriptionStatus(status: string) {
  return STATUS[status] ?? { label: status || "未知狀態", variant: "outline" as const };
}

/**
 * v1 feature 的顯示名稱（spec §4.5 的 7 項；`limit.storage_gb` 已移除）。
 * 投影只帶 `feature_code`、不帶 `description`／`unit`，所以名稱在此對照；
 * 沒對照到的**照 code 顯示**（新 feature 上線時顯示代碼，不猜也不隱藏）。
 */
const FEATURE_LABELS: Record<string, string> = {
  "limit.seats": "席位",
  "limit.customers": "客戶",
  "limit.products": "商品",
  "limit.departments": "部門",
  "feature.printing": "單據列印",
  "feature.dispatch": "派車看板",
  "feature.returns": "退貨申請與審核",
};

export function featureLabel(code: string): string {
  return FEATURE_LABELS[code] ?? code;
}

/**
 * 單筆用量的顯示值。
 *
 * 種類只能由命名慣例推導（投影不帶 `type`）：`limit.*` 是數值配額、`feature.*` 是布林功能
 * （見 spec §4.5 的 v1 清單與 `cmd/seed/platform.go`）。`limit_set=false` 對數值配額的語意是
 * **不限**（不是上限 0），布林則沒有上限可言。
 */
export function usageValue(usage: Usage): string {
  if (!usage.enabled) return "方案未含";
  const limit = Number(usage.limitValue);
  const used = Number(usage.used);
  if (usage.limitSet && limit > 0) return `${used}/${limit}`;
  return usage.featureCode.startsWith("limit.") ? `${used}（不限）` : "已開通";
}

/**
 * 是否為「尚未開通計費」：沒有未取消的訂閱列時，後端的投影回 `status="none"`、`plan_code=""`，
 * 且**仍會列出各 feature 的 `enabled=false`**，但守衛其實不施加任何限制（Plan B F-7 裁定的
 * 語意，未結項 #13）。前端分不出這個狀態與「方案不含」，所以這種情況一律不列功能、也不說
 * 「方案未含」。
 */
export function isUnprovisioned(entitlements: GetTenantEntitlementsResponse): boolean {
  return entitlements.status === "none" || entitlements.planCode === "";
}

/** 試用到期日（UTC 日期）；後端傳 RFC3339，nil 為空字串。無法解析時照原字串顯示。 */
export function formatTrialEndsAt(trialEndsAt: string): string {
  if (!trialEndsAt) return "";
  const ms = Date.parse(trialEndsAt);
  return Number.isNaN(ms) ? trialEndsAt : new Date(ms).toISOString().slice(0, 10);
}

/**
 * 需要提醒的事項（用量近上限、試用將到期）。沒達門檻一律回空陣列——
 * 「不吵」是這個函式的契約：卡片與 banner 都只在這裡拿到東西時才顯示提示。
 */
export function entitlementAlerts(
  entitlements: GetTenantEntitlementsResponse,
  now: Date = new Date()
): PlanAlert[] {
  if (isUnprovisioned(entitlements)) return [];

  const alerts: PlanAlert[] = [];
  for (const usage of entitlements.usage) {
    const limit = Number(usage.limitValue);
    if (!usage.enabled || !usage.limitSet || limit <= 0) continue;
    const used = Number(usage.used);
    const ratio = used / limit;
    if (ratio < USAGE_WARN_RATIO) continue;
    const urgent = ratio >= USAGE_URGENT_RATIO;
    alerts.push({
      severity: urgent ? "urgent" : "warning",
      // 90% 不只在色階上升級，文案也說得更重（「還有多少」與「快沒了」是不同的事）。
      text: `${featureLabel(usage.featureCode)}用量 ${used}/${limit}（${Math.floor(
        ratio * 100
      )}%），接近上限${urgent ? "（即將額滿）" : ""}`,
    });
  }

  const trialEndsAt = Date.parse(entitlements.trialEndsAt);
  if (entitlements.status === "trialing" && !Number.isNaN(trialEndsAt)) {
    const daysLeft = Math.ceil((trialEndsAt - now.getTime()) / 86_400_000);
    if (daysLeft <= TRIAL_ALERT_DAYS) {
      alerts.push({
        severity: "warning",
        text: daysLeft > 0 ? `試用將於 ${daysLeft} 天後到期` : "試用已到期",
      });
    }
  }
  return alerts;
}
