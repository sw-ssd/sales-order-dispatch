import { Code, ConnectError } from "@connectrpc/connect";
import { beforeEach, describe, expect, it, vi } from "vitest";

const listTenants = vi.fn();
vi.mock("./api", () => ({
  platform: { listTenants: (...args: unknown[]) => listTenants(...args) },
  loginUrl: "/platform/auth/google",
}));

import { ensureSession, logout, PROBE_TTL_MS, resetSession, sessionStatus } from "./session";

describe("ensureSession（session 探針）", () => {
  beforeEach(() => {
    listTenants.mockReset();
    resetSession();
  });

  it("後端回 Unauthenticated → false，狀態 anonymous（fail-closed）", async () => {
    listTenants.mockRejectedValue(new ConnectError("未登入", Code.Unauthenticated));
    await expect(ensureSession()).resolves.toBe(false);
    expect(sessionStatus()).toBe("anonymous");
  });

  it("連線失敗（非 401）一樣視為未登入：寧可擋在登入頁", async () => {
    listTenants.mockRejectedValue(new ConnectError("後端不可用", Code.Unavailable));
    await expect(ensureSession()).resolves.toBe(false);
    expect(sessionStatus()).toBe("anonymous");
  });

  it("探針成功 → true，狀態 authenticated", async () => {
    listTenants.mockResolvedValue({ tenants: [] });
    await expect(ensureSession()).resolves.toBe(true);
    expect(sessionStatus()).toBe("authenticated");
  });

  it("單一探針：並行與後續呼叫合計只打後端一次", async () => {
    listTenants.mockResolvedValue({ tenants: [] });
    await Promise.all([ensureSession(), ensureSession()]);
    await ensureSession();
    expect(listTenants).toHaveBeenCalledTimes(1);
  });

  it("成功結果不得永久快取：過了 TTL 會重新探針", async () => {
    const now = vi.spyOn(Date, "now");
    try {
      now.mockReturnValue(0);
      listTenants.mockResolvedValue({ tenants: [] });
      await ensureSession();

      now.mockReturnValue(PROBE_TTL_MS + 1);
      await ensureSession();
      expect(listTenants).toHaveBeenCalledTimes(2);
    } finally {
      now.mockRestore();
    }
  });

  it("舊探針晚回不得覆寫新探針的狀態（TTL 替換競態）", async () => {
    // 未結項 #25 中段：TTL 一過就同步替換 probe —— 舊探針晚回會覆寫新探針已寫的狀態。
    const first = Promise.withResolvers<{ tenants: never[] }>();
    const second = Promise.withResolvers<{ tenants: never[] }>();
    const now = vi.spyOn(Date, "now");
    try {
      now.mockReturnValue(0);
      listTenants.mockReturnValueOnce(first.promise);
      const p1 = ensureSession();
      now.mockReturnValue(PROBE_TTL_MS + 1);
      listTenants.mockReturnValueOnce(second.promise);
      const p2 = ensureSession();
      // 新探針先失敗 → anonymous；舊探針後成功 → 不得翻回 authenticated。
      second.reject(new ConnectError("後端不可用", Code.Unavailable));
      await expect(p2).resolves.toBe(false);
      expect(sessionStatus()).toBe("anonymous");
      first.resolve({ tenants: [] });
      await expect(p1).resolves.toBe(false);
      expect(sessionStatus()).toBe("anonymous");
      expect(listTenants).toHaveBeenCalledTimes(2);
    } finally {
      now.mockRestore();
    }
  });
});

describe("logout", () => {
  beforeEach(() => {
    listTenants.mockReset();
    resetSession();
  });

  it("登出後狀態回 anonymous，且不再重新探針（cookie 是 HttpOnly，前端清不掉）", async () => {
    listTenants.mockResolvedValue({ tenants: [] });
    await ensureSession();
    logout();
    expect(sessionStatus()).toBe("anonymous");
    await expect(ensureSession()).resolves.toBe(false);
    expect(listTenants).toHaveBeenCalledTimes(1);
  });

  it("登出後即使過了 TTL 也不再探針（否則會用仍在的 cookie 把人放回主控台）", async () => {
    const now = vi.spyOn(Date, "now");
    try {
      now.mockReturnValue(0);
      listTenants.mockResolvedValue({ tenants: [] });
      await ensureSession();
      logout();

      now.mockReturnValue(PROBE_TTL_MS * 10);
      await expect(ensureSession()).resolves.toBe(false);
      expect(listTenants).toHaveBeenCalledTimes(1);
    } finally {
      now.mockRestore();
    }
  });

  it("登出時已有飛行中的探針：其成功回來不得把狀態翻回 authenticated", async () => {
    // 未結項 #25 前半：logout 不取消已在飛行的探針 —— 其 .then 回來會 setStatus("authenticated")。
    const { promise, resolve } = Promise.withResolvers<{ tenants: never[] }>();
    listTenants.mockReturnValue(promise);
    const inFlight = ensureSession();
    logout();
    resolve({ tenants: [] });
    await expect(inFlight).resolves.toBe(false);
    expect(sessionStatus()).toBe("anonymous");
  });
});
