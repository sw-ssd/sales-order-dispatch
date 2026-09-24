// 直接以 JSON-RPC over HTTP 呼叫 Pixso MCP（127.0.0.1:3667），繞過 eval 工具的 30s 逾時。
const URL_ = "http://127.0.0.1:3667/mcp";
const TIMEOUT_MS = Number(process.env.TK_TIMEOUT_MS || 180_000);
let sessionId = null;
let nextId = 1;

async function rpc(method, params, { notify = false } = {}) {
  const body = { jsonrpc: "2.0", method, ...(notify ? {} : { id: nextId++ }), ...(params ? { params } : {}) };
  const headers = {
    "Content-Type": "application/json",
    Accept: "application/json, text/event-stream",
    ...(sessionId ? { "mcp-session-id": sessionId } : {}),
  };
  // Node 的 fetch 沒有預設逾時；Pixso 端連線卡住時會永久掛住，故一律帶 abort。
  const res = await fetch(URL_, {
    method: "POST", headers, body: JSON.stringify(body), signal: AbortSignal.timeout(TIMEOUT_MS),
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}: ${await res.text()}`);
  const sid = res.headers.get("mcp-session-id");
  if (sid) sessionId = sid;
  const text = await res.text();
  if (notify) return null;
  // SSE 或純 JSON
  if (text.startsWith("event:") || text.includes("\ndata:")) {
    for (const line of text.split("\n")) {
      if (line.startsWith("data: ")) {
        const msg = JSON.parse(line.slice(6));
        if (msg.id !== undefined && msg.error) throw new Error(JSON.stringify(msg.error));
        if (msg.id !== undefined) return msg.result;
      }
    }
    return null;
  }
  const msg = JSON.parse(text);
  if (msg.error) throw new Error(JSON.stringify(msg.error));
  return msg.result;
}

export async function connect() {
  await rpc("initialize", {
    protocolVersion: "2024-11-05",
    capabilities: {},
    clientInfo: { name: "tailkit-pixso", version: "1.0.0" },
  });
  await rpc("notifications/initialized", {}, { notify: true });
  return sessionId;
}

export async function callTool(name, args) {
  const r = await rpc("tools/call", { name, arguments: args });
  if (r?.isError) throw new Error(`tool ${name} error: ${JSON.stringify(r)}`);
  const text = r?.content?.map((c) => c.text ?? "").join("") ?? "";
  return text;
}

// 自我測試：只讀不寫，避免污染畫布。
// 用法：node pixso-mcp.mjs selftest
if (process.argv[2] === "selftest") {
  console.log("session:", await connect());
  const t = await callTool("eval_script", {
    script: "return {file: pixso.root.name, page: pixso.currentPage.name, n: pixso.currentPage.children.length};",
  });
  console.log("eval_script ->", t.slice(0, 300));
}
