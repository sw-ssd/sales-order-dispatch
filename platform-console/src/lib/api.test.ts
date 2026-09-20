import { afterEach, describe, expect, it } from "vitest";
import { platform } from "./api";

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

    expect(String(calls[0]?.input)).toBe("/platform.v1.PlatformAdminService/ListTenants");
    // operator session 是 HttpOnly cookie：每個請求都必須帶上。
    expect(calls[0]?.init?.credentials).toBe("include");
    expect(res.tenants).toHaveLength(1);
    expect(res.tenants[0]?.companyName).toBe("甲公司");
    expect(res.tenants[0]?.seatCount).toBe(8);
    expect(res.pagination?.total).toBe(1);
  });
});
