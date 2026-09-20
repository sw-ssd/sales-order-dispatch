import { redirect } from "@tanstack/solid-router";
import { ensureSession } from "./session";

/**
 * 路由守衛：未登入（探針失敗）一律導向登入頁。
 *
 * **前端守衛不構成授權**：它只是 UX——少一次「進得去但每個 RPC 都 401」的畫面。
 * 唯一決策者是後端：`PlatformAdminService` 由 `operatorauth.Interceptor` 擋下，
 * 每個請求驗 cookie、回查 operator 白名單；就算有人改前端碼或直接打 API，
 * 拿不到任何跨租戶資料。
 *
 * preload（hover/focus 預熱）不導向、也不觸發探針的副作用。
 */
export function requireOperator() {
  return async ({ preload }: { preload: boolean }) => {
    if (preload) return;
    if (!(await ensureSession())) throw redirect({ to: "/login" });
  };
}
