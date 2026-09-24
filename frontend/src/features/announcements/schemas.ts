import * as vali from "valibot";

/** 公告三型別（spec announcements 資料模型;非法值後端亦拒）。 */
export const ANNOUNCEMENT_TYPES = ["banner", "news", "article"] as const;

/** 三型別的中文標籤（表格 badge 與表單 select 共用）。 */
export const ANNOUNCEMENT_TYPE_LABELS: Record<(typeof ANNOUNCEMENT_TYPES)[number], string> = {
  banner: "輪播",
  news: "最新消息",
  article: "圖文文章",
};

/** 發佈範圍選擇（auto = 自動歸屬:非 super 由後端補自己的公司/部門,見 proto 註）。 */
export const ANNOUNCEMENT_SCOPES = ["auto", "system", "company", "department"] as const;

/**
 * 公告表單的欄位規則（新增／編輯共用）。
 *
 * 只鏡射後端必填：`type`（三值之一,表單以 select 供值）與 `title`（trim 後非空）。
 * 時間欄位是 `datetime-local` 字串,送出前經 `toRFC3339` 轉時刻;留空＝不帶
 * （publish 空 → 立即,unpublish 空 → 不自動下架,見 proto）。
 * `sortOrder` 為字串表單值,送出前 `toSortOrder` 轉 int32（同 masters 慣例:
 * 空或非數字歸 0,後端不驗格式 —— 不製造後端沒有的限制）。
 */
export const announcementSchema = vali.object({
  scope: vali.picklist(ANNOUNCEMENT_SCOPES),
  companyId: vali.string(),
  departmentId: vali.string(),
  type: vali.picklist(ANNOUNCEMENT_TYPES),
  title: vali.pipe(vali.string(), vali.trim(), vali.nonEmpty("請輸入標題")),
  content: vali.string(),
  imageUrl: vali.string(),
  linkUrl: vali.string(),
  publishAt: vali.string(),
  unpublishAt: vali.string(),
  sortOrder: vali.string(),
  isActive: vali.boolean(),
  deployWeb: vali.boolean(),
  deployApp: vali.boolean(),
});

/**
 * `datetime-local` 值 → RFC3339（本地時刻 → UTC 顯示為 Z;空字串回空,
 * 由後端套「空 = 立即 / 不下架」語意）。
 */
export function toRFC3339(local: string): string {
  const raw = local.trim();
  if (raw === "") return "";
  const d = new Date(raw);
  return Number.isNaN(d.getTime()) ? "" : d.toISOString();
}

/** RFC3339 → `datetime-local` 值（本地時刻;空/無效回空）。 */
export function fromRFC3339(rfc: string): string {
  const raw = rfc.trim();
  if (raw === "") return "";
  const d = new Date(raw);
  if (Number.isNaN(d.getTime())) return "";
  const pad = (n: number) => String(n).padStart(2, "0");
  return (
    `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}` +
    `T${pad(d.getHours())}:${pad(d.getMinutes())}`
  );
}

/** 範圍選擇 → 建立請求的 company/department 欄位（auto/system → 空字串）。 */
export function scopeTargets(scope: (typeof ANNOUNCEMENT_SCOPES)[number], companyId: string, departmentId: string): {
  companyId: string;
  departmentId: string;
} {
  if (scope === "company") return { companyId: companyId.trim(), departmentId: "" };
  if (scope === "department") {
    return { companyId: companyId.trim(), departmentId: departmentId.trim() };
  }
  return { companyId: "", departmentId: "" };
}
