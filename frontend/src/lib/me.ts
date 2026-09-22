import { queryOptions } from "@tanstack/solid-query";
import { API_BASE } from "~/lib/transport";

/** GET /api/v1/me 回應：session 身分與所屬公司品牌（後端 handlers.Me）。 */
export interface Me {
  user_id: string;
  role: string;
  /** 所屬公司；無公司、公司不存在或已軟刪一律 null（後端不區分，前端照樣降級）。 */
  company: { id: string; name: string; logo_url: string } | null;
}

/**
 * 身分查詢（queryKey `["me"]`：Logo 上傳成功後與 `["companies"]` 一起失效）。
 * 401（未登入）回 null 而非錯誤——沒有身分不是故障，側邊欄與上傳鈕各自降級為預設行為；
 * 其餘非 2xx 才 throw（走全域 retry 謂詞，失敗時查詢為 error 狀態同樣降級）。
 */
export const meQueryOptions = queryOptions({
  queryKey: ["me"],
  staleTime: 60_000,
  queryFn: async (): Promise<Me | null> => {
    const res = await fetch(`${API_BASE}/me`);
    if (res.status === 401) return null;
    if (!res.ok) throw new Error(`me 查詢失敗: ${res.status}`);
    return (await res.json()) as Me;
  },
});
