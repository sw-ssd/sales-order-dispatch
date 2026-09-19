import { Code, ConnectError } from "@connectrpc/connect";
import { describe, expect, it } from "vitest";
import { isRetryableQueryError, queryClient } from "./query-client";

describe("共用查詢 retry 謂詞", () => {
  it("重試暫時性錯誤:Unavailable / Unknown / DeadlineExceeded", () => {
    for (const code of [Code.Unavailable, Code.Unknown, Code.DeadlineExceeded]) {
      expect(isRetryableQueryError(new ConnectError("暫時性失敗", code)), Code[code]).toBe(true);
    }
  });

  it("不重試確定性錯誤", () => {
    const deterministic = [
      Code.NotFound,
      Code.InvalidArgument,
      Code.PermissionDenied,
      Code.Unauthenticated,
      Code.AlreadyExists,
      Code.FailedPrecondition,
    ];
    for (const code of deterministic) {
      expect(isRetryableQueryError(new ConnectError("確定性失敗", code)), Code[code]).toBe(false);
    }
  });

  it("白名單以外的 ConnectError 一律不重試", () => {
    const others = [
      Code.Canceled,
      Code.ResourceExhausted,
      Code.Aborted,
      Code.OutOfRange,
      Code.Unimplemented,
      Code.Internal,
      Code.DataLoss,
    ];
    for (const code of others) {
      expect(isRetryableQueryError(new ConnectError("其他", code)), Code[code]).toBe(false);
    }
  });

  it("非 ConnectError 不重試", () => {
    expect(isRetryableQueryError(new Error("boom"))).toBe(false);
    expect(isRetryableQueryError("boom")).toBe(false);
    expect(isRetryableQueryError(undefined)).toBe(false);
    expect(isRetryableQueryError(null)).toBe(false);
  });

  it("QueryClient 預設 retry 依謂詞判斷且封頂 3 次", () => {
    const retry = queryClient.getDefaultOptions().queries?.retry;
    if (typeof retry !== "function") {
      throw new Error("預設 retry 必須是函式");
    }
    const unavailable = new ConnectError("後端未啟動", Code.Unavailable);
    const forbidden = new ConnectError("無權限", Code.PermissionDenied);
    // 第 1 次失敗(未達上限)且為暫時性錯誤 → 重試。
    expect(retry(0, unavailable)).toBe(true);
    // 已重試 3 次 → 停止(否則 Unavailable 會無限退避重試)。
    expect(retry(3, unavailable)).toBe(false);
    expect(retry(0, forbidden)).toBe(false);
  });
});
