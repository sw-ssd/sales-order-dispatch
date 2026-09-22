import { createConnectTransport } from "@connectrpc/connect-web";

// 全站共用 Connect transport(與既有頁面 inline 設定一致):前端相對路徑
// /api/v1 由 dev proxy(Vite)與正式部署反代轉發至後端。
//
// API_BASE 為同一個前綴的**單一定義處**:非 Connect 的 REST 呼叫(原生 fetch,如 lib/me.ts、
// Logo 上傳)必須由這裡取用,不得各自硬編 —— 改前綴時才不會漏掉幾處。
export const API_BASE = "/api/v1";

export const transport = createConnectTransport({ baseUrl: API_BASE });
