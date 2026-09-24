// 把同一 subcategory 的獨立元件合成 component set（單一 style 軸）。
// 合成時暫時解除父層 auto-layout，完成後還原。失敗者保持獨立。
import { connect, callTool } from "./pixso-mcp.mjs";
import { readFileSync, writeFileSync, existsSync } from "node:fs";

const ROOT = process.env.TK_DIR || "/tmp/tailkit-pixso";
await connect();
const ev = async (s) => {
  const t = await callTool("eval_script", { script: s });
  try { return JSON.parse(t); } catch { return t; }
};

const manifest = JSON.parse(readFileSync(`${ROOT}/manifest.json`, "utf8"));

// variant 值不可含 = 或換行；取 Title，缺就用 id
const clean = (s) => String(s ?? "").replace(/[=\n\r]/g, " ").replace(/\s+/g, " ").trim().slice(0, 60) || "default";

const groups = new Map();
for (const [id, v] of Object.entries(manifest)) {
  const k = `${v.pkg}/${v.cat}/${v.sub}`;
  if (!groups.has(k)) groups.set(k, []);
  groups.get(k).push({ id, ...v });
}

const out = existsSync(`${ROOT}/variant-sets.json`)
  ? JSON.parse(readFileSync(`${ROOT}/variant-sets.json`, "utf8"))
  : { sets: [], standalone: [], failed: [] };
const doneKeys = new Set(out.sets.map((s) => s.key));

for (const [key, members] of groups) {
  if (doneKeys.has(key)) continue;
  if (members.length < 2) { if (!out.standalone.includes(key)) out.standalone.push(key); continue; }
  const spec = members.map((m) => ({ pixso_id: m.pixso_id, id: m.id, style: clean(m.title) }));
  const res = await ev(`
function find(n, id) { if (n.id === id) return n; if ('children' in n) for (const c of n.children) { const r = find(c, id); if (r) return r; } return null; }
const page = pixso.root.children.find((p) => p.name === "Tailkit 元件庫");
const spec = ${JSON.stringify(spec)};
const found = [];
for (const s of spec) { const n = find(page, s.pixso_id); if (n) found.push({ node: n, s }); }
if (found.length < 2) return { skipped: true, found: found.length };
const parent = found[0].node.parent;
const saved = { mode: parent.layoutMode, spacing: parent.itemSpacing };
try {
  if (parent.layoutMode && parent.layoutMode !== "NONE") parent.layoutMode = "NONE";
  const set = pixso.combineAsVariants(found.map((x) => x.node), parent);
  set.name = ${JSON.stringify(key)};
  set.setPluginData("tk_section", ${JSON.stringify(key)});
  const renamed = [];
  for (const c of set.children) {
    const m = found.find((x) => x.node.id === c.id);
    c.name = "style=" + (m ? m.s.style : "v" + c.id);
    c.setPluginData("tk_id", m ? m.s.id : "");
    renamed.push(c.name);
  }
  return { setId: set.id, type: set.type, variants: renamed };
} finally {
  try { if (saved.mode && saved.mode !== "NONE") { parent.layoutMode = saved.mode; parent.itemSpacing = saved.spacing; } } catch (e) {}
}
`);
  if (res && res.setId) {
    out.sets.push({ key, setId: res.setId, count: members.length, variants: res.variants });
    console.log(`SET ${key} (${members.length}) -> ${res.setId}`);
  } else {
    out.failed.push({ key, res });
    console.log(`FAIL ${key} ${JSON.stringify(res).slice(0, 160)}`);
  }
  writeFileSync(`${ROOT}/variant-sets.json`, JSON.stringify(out, null, 1));
}
console.log("sets:", out.sets.length, "standalone:", out.standalone.length, "failed:", out.failed.length);
