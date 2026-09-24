// 直連 tailkit.com/mcp 的 MCP client：token 由本機 omp credential store 讀取（不落地、不入對話）
import { execFileSync } from "node:child_process";
import { copyFileSync } from "node:fs";

const DB = process.env.OMP_AGENT_DB || `${process.env.HOME}/.omp/agent/agent.db`;
const COPY = `${process.env.TK_DIR || "/tmp/tailkit-pixso"}/.agent-db-read`;
const URL_ = "https://tailkit.com/mcp";

function loadToken() {
  copyFileSync(DB, COPY);
  const out = execFileSync("sqlite3", [
    COPY,
    "select data from auth_credentials where provider like 'mcp_oauth:%tailkit.com/mcp' order by id desc limit 1",
  ], { encoding: "utf8" }).trim();
  const cred = JSON.parse(out);
  return { access: cred.access, refresh: cred.refresh, tokenUrl: cred.tokenUrl, clientId: cred.clientId, expires: cred.expires };
}

let token = loadToken();
let sessionId = null;
let nextId = 1;

async function refreshToken() {
  const body = new URLSearchParams({
    grant_type: "refresh_token",
    refresh_token: token.refresh,
    client_id: token.clientId,
    resource: URL_,
  });
  const res = await fetch(token.tokenUrl, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body,
  });
  if (!res.ok) throw new Error(`refresh failed ${res.status}: ${await res.text()}`);
  const j = await res.json();
  token.access = j.access_token ?? j.access;
  if (j.refresh_token) token.refresh = j.refresh_token;
  if (j.expires_in) token.expires = Date.now() + j.expires_in * 1000;
  return token;
}

async function rpc(method, params, { notify = false, retry = true } = {}) {
  const body = { jsonrpc: "2.0", method, ...(notify ? {} : { id: nextId++ }), ...(params ? { params } : {}) };
  const res = await fetch(URL_, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json, text/event-stream",
      Authorization: `Bearer ${token.access}`,
      ...(sessionId ? { "mcp-session-id": sessionId } : {}),
    },
    body: JSON.stringify(body),
    signal: AbortSignal.timeout(Number(process.env.TK_TIMEOUT_MS || 120_000)),
  });
  if (res.status === 401 && retry) {
    await refreshToken();
    sessionId = null;
    return rpc(method, params, { notify, retry: false });
  }
  if (res.status === 429) {
    const wait = Number(res.headers.get("retry-after") || 0) * 1000 || 2000;
    await new Promise((r) => setTimeout(r, wait));
    return rpc(method, params, { notify, retry });
  }
  if (!res.ok) throw new Error(`HTTP ${res.status}: ${await res.text()}`);
  const sid = res.headers.get("mcp-session-id");
  if (sid) sessionId = sid;
  const text = await res.text();
  if (notify) return null;
  if (text.includes("data:")) {
    for (const line of text.split("\n")) {
      if (line.startsWith("data: ")) {
        const msg = JSON.parse(line.slice(6));
        if (msg.error) throw new Error(JSON.stringify(msg.error));
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
}

export async function callTool(name, args) {
  await new Promise((r) => setTimeout(r, 350));
  const r = await rpc("tools/call", { name, arguments: args });
  if (r?.isError) throw new Error(`tool ${name} error: ${JSON.stringify(r).slice(0, 400)}`);
  return (r?.content ?? []).map((c) => c.text ?? "").join("");
}

if (process.argv[2] === "selftest") {
  await connect();
  const cat = await callTool("browse_catalog", { level: "packages" });
  console.log("catalog head:", cat.slice(0, 200).replace(/\n/g, " | "));
  const code = await callTool("get_component_code", { identifiers: ["a-c-alerts-01"], tech: "html" });
  console.log("code len:", code.length);
}
