import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { PlatformAdminService } from "./proto/platform/v1/platform_pb";

/**
 * 平台 API 的唯一入口：console **只**呼叫 `platform/v1`（PlatformAdminService）。
 *
 * 認證沿用後端既有的 operator session（`internal/platform/operatorauth`），不另立一套：
 * 登入走 OIDC（`/platform/auth/google`），成功後後端以 **HttpOnly** cookie
 * `platform_session`（Path=/platform）承載 JWT。
 *
 * 兩個必然結果：
 * 1. 每個請求都要帶 cookie → `credentials: "include"`（前端看不到 cookie 內容，
 *    因此「有沒有登入」只能問後端，見 lib/session.ts）。
 * 2. base URL 必須可設定（VITE_API_BASE_URL）：dev 由 Vite proxy 把 /platform 轉到 API；
 *    正式部署若 console 與 API 不同 origin，就設成 API 來源。
 */
const API_BASE = (import.meta.env.VITE_API_BASE_URL ?? "").replace(/\/+$/, "");

/** 平台登入端點（後端 operatorauth.LoginPath）；登入成功後導回 config.Platform.ConsoleURL。 */
export const LOGIN_PATH = "/platform/auth/google";

export const apiBaseUrl = API_BASE;
export const loginUrl = `${API_BASE}${LOGIN_PATH}`;

export const transport = createConnectTransport({
  baseUrl: API_BASE === "" ? "/" : `${API_BASE}/`,
  fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
});

export const platform = createClient(PlatformAdminService, transport);
