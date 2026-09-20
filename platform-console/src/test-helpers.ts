import { create, toBinary } from "@bufbuild/protobuf";
import { Code, ConnectError } from "@connectrpc/connect";
import { vi } from "vitest";
import { ErrorInfoSchema } from "./lib/proto/salesorder/v1/common_pb";

/**
 * 造一個**線路形狀**的錯誤：後端 `errcode.Code.Error` 的產物就是帶 ErrorInfo detail 的
 * ConnectError。測試要驗的是「console 顯示的是 ErrorInfo 的碼／訊息／details」，
 * 所以不能用裸 `new Error()` 代替——那會走不到正確的分支，測出來的綠是假的。
 */
export function connectErrorWithInfo(
  code: string,
  options: { message?: string; details?: Record<string, string>; connectCode?: Code } = {},
): ConnectError {
  return new ConnectError(
    options.message ?? code,
    options.connectCode ?? Code.PermissionDenied,
    undefined,
    [
      {
        desc: ErrorInfoSchema,
        value: { code, message: options.message ?? "", details: options.details ?? {} },
      },
    ],
  );
}

/** 以**例外**表達「後端拒絕」：`stubPlatformWire` 會把它轉成線路上的 Connect 錯誤回應。 */
export class WireFault extends Error {
  readonly code: string;
  readonly details: Record<string, string>;
  constructor(info: { code: string; message?: string; details?: Record<string, string> }) {
    super(info.message ?? info.code);
    this.code = info.code;
    this.details = info.details ?? {};
  }
}

/** 一次線路呼叫：方法名（`PlatformAdminService` 的哪個 RPC）、URL、解碼後的 JSON 請求內容。 */
export type WireCall = { method: string; url: string; body: Record<string, unknown> };

/**
 * 在 **fetch** 上假造 platform/v1 的線路（不 mock `lib/api`）：走真的 transport，因此
 * URL 前綴（`PLATFORM_PATH`）、協定（Connect ＋ JSON）、以及**預設值被省略**的欄位
 * 全都是真的 —— 「留空的金額沒有送上線」這種斷言才有意義。
 *
 * handler 依 RPC 方法名分派，回傳值即回應訊息；要模擬失敗就 `throw new WireFault(...)`。
 * 回傳的 `calls` 依序記錄每次呼叫，斷言直接看它。
 */
export function stubPlatformWire(
  handlers: Record<string, (body: Record<string, unknown>) => unknown>,
): WireCall[] {
  const calls: WireCall[] = [];
  vi.stubGlobal("fetch", async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    const method = url.split("/").pop() ?? "";
    const body: Record<string, unknown> = init?.body
      ? JSON.parse(new TextDecoder().decode(init.body as Uint8Array))
      : {};
    calls.push({ method, url, body });

    const handler = handlers[method];
    if (!handler) {
      return new Response(JSON.stringify({ code: "unimplemented", message: `未假造 ${method}` }), {
        status: 404,
        headers: { "content-type": "application/json" },
      });
    }
    try {
      return new Response(JSON.stringify((await handler(body)) ?? {}), {
        status: 200,
        headers: { "content-type": "application/json" },
      });
    } catch (err) {
      if (!(err instanceof WireFault)) throw err;
      const info = create(ErrorInfoSchema, {
        code: err.code,
        message: err.message,
        details: err.details,
      });
      // 後端（connect-go）把 ErrorInfo 放進 Any：type 是不含前綴的完整名稱、value 是二進位。
      const binary = toBinary(ErrorInfoSchema, info);
      let base64 = "";
      for (const byte of binary) base64 += String.fromCharCode(byte);
      return new Response(
        JSON.stringify({
          code: "unknown",
          message: err.message,
          details: [{ type: "salesorder.v1.ErrorInfo", value: btoa(base64) }],
        }),
        { status: 400, headers: { "content-type": "application/json" } },
      );
    }
  });
  return calls;
}
