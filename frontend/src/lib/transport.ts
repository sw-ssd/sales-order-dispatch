import { createConnectTransport } from "@connectrpc/connect-web";

// 全站共用 Connect transport(與既有頁面 inline 設定一致):前端相對路徑
// /api/v1 由 dev proxy(Vite)與正式部署反代轉發至後端。
export const transport = createConnectTransport({ baseUrl: "/api/v1" });
