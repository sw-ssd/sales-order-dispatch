import { afterEach, describe, expect, it } from "vitest";
import { LOGIN_PATH, loginUrl, PLATFORM_PATH, platform } from "./api";

/**
 * 產生的 proto 與 Connect client 真的被用來打 `platform/v1`：這個測試不 mock 我們的
 * API 模組，而是攔下 **fetch**，因此走的是完整的生成程式碼路徑（訊息序列化、URL 組裝）。
 */
const realFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = realFetch;
});

describe("platform client", () => {
  it("打 platform.v1.PlatformAdminService/ListTenants，帶 cookie，並解析生成的訊息型別", async () => {
    const calls: { input: unknown; init?: RequestInit }[] = [];
    globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
      calls.push({ input, init });
      return new Response(
        JSON.stringify({
          tenants: [
            {
              companyId: "7",
              companyName: "甲公司",
              planCode: "std",
              planName: "標準",
              subscriptionStatus: "active",
              seatCount: "8",
              overdue: false,
            },
          ],
          pagination: { page: 1, pageSize: 1, total: 1 },
        }),
        { status: 200, headers: { "content-type": "application/json" } },
      );
    }) as typeof fetch;

    const res = await platform.listTenants({ page: 1, pageSize: 1 });

    // 後端把平台 RPC 掛在字面 /platform/ 之下（domains.go: Mount(operatorauth.CookiePath, StripPrefix(...))），
    // 少了這一段就是 404，而且 cookie 的 Path=/platform 依 RFC 6265 逐段前綴比對也不涵蓋
    // /platform.v1.…（未涵蓋的第一個字元是 "."）→ 瀏覽器根本不會送出 cookie。
    // 後端那一側由 internal/server/platform_admin_mount_integration_test.go 的
    // TestIntegrationPlatformAdminMount 釘住（含「少了前綴的 procedure 必須 404」）。
    expect(String(calls[0]?.input)).toBe("/platform/platform.v1.PlatformAdminService/ListTenants");
    // operator session 是 HttpOnly cookie：每個請求都必須帶上。
    expect(calls[0]?.init?.credentials).toBe("include");
    expect(res.tenants).toHaveLength(1);
    expect(res.tenants[0]?.companyName).toBe("甲公司");
    expect(res.tenants[0]?.seatCount).toBe(8);
    expect(res.pagination?.total).toBe(1);
  });

  it("base 路徑恰為 /platform（後端掛載前綴＝cookie Path；登入端點同一個前綴）", () => {
    // 兩端的字面值必須一致：後端是 operatorauth.CookiePath／LoginPath，
    // 前端是 PLATFORM_PATH。任一側改了而另一側沒改 → cookie 不送出或 404。
    expect(PLATFORM_PATH).toBe("/platform");
    expect(LOGIN_PATH).toBe("/platform/auth/google");
    expect(loginUrl).toBe("/platform/auth/google");
  });
});
