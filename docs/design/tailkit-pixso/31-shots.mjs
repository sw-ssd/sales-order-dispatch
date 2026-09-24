// 抽樣截圖：對指定元件匯出 PNG 到 shots/，供人工目視驗證。
// 用法：node 31-shots.mjs <identifier> [identifier...]
import { connect, callTool } from "./pixso-mcp.mjs";
import { mkdirSync, writeFileSync } from "node:fs";
import { execFileSync } from "node:child_process";

const ROOT = process.env.TK_DIR || "/tmp/tailkit-pixso";
await connect();
const ev = async (s) => {
  const t = await callTool("eval_script", { script: s });
  try { return JSON.parse(t); } catch { return t; }
};

const wanted = process.argv.slice(2);
if (!wanted.length) { console.log("usage: node 31-shots.mjs <id>..."); process.exit(1); }

const targets = await ev(`
const page = pixso.root.children.find((p) => p.name === "Tailkit 元件庫");
const want = ${JSON.stringify(wanted)};
const out = [];
for (const sec of page.children) for (const c of sec.children) if (want.includes(c.name)) out.push({ name: c.name, id: c.id, w: Math.round(c.width), h: Math.round(c.height) });
return out;
`);
mkdirSync(`${ROOT}/shots`, { recursive: true });
for (const t of targets) {
  const url = String(await callTool("get_export_image", {
    guid: t.id, exportSettings: { imageType: 1, constraint: { type: 1, value: 2 } },
  })).split("\n")[0].trim();
  const out = `${ROOT}/shots/${t.name}.png`;
  execFileSync("curl", ["-s", "-o", out, url]);
  console.log(`${t.name} ${t.w}x${t.h} -> ${out}`);
}
console.log("done", targets.length);
