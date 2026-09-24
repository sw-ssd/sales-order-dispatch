// 批次抓取全部 646 個元件的 HTML 原始碼（每批 10 個）。
import { writeFileSync, mkdirSync, existsSync, readFileSync } from "node:fs";
import { connect, callTool } from "./tailkit-mcp.mjs";

const ROOT = process.env.TK_DIR || "/tmp/tailkit-pixso";

await connect();
const cat = JSON.parse(readFileSync(`${ROOT}/catalog.json`, "utf8"));
const all = [];
for (const p of cat.packages)
  for (const c of p.categories)
    for (const s of c.subcategories)
      for (const comp of s.components)
        all.push({ ...comp, pkg: p.slug, cat: c.slug, sub: s.slug });

mkdirSync(`${ROOT}/raw`, { recursive: true });

function splitBlocks(md) {
  const out = {};
  let cur = null, collecting = false, buf = [];
  for (const line of md.split("\n")) {
    const m = line.match(/^# .*\(`([^`]+)`\)/);
    if (m) { cur = m[1]; collecting = false; continue; }
    if (cur && line.trim() === "```html") { buf = []; collecting = true; continue; }
    if (collecting && line.trim() === "```") { out[cur] = buf.join("\n"); collecting = false; continue; }
    if (collecting) buf.push(line);
  }
  return out;
}

const missing = [];
let done = 0;
for (let i = 0; i < all.length; i += 10) {
  const batch = all.slice(i, i + 10);
  const ids = batch.map((b) => b.id);
  const md = await callTool("get_component_code", { identifiers: ids, tech: "html" });
  const blocks = splitBlocks(md);
  for (const b of batch) {
    const html = blocks[b.id];
    if (!html) { missing.push(b.id); continue; }
    writeFileSync(`${ROOT}/raw/${b.id}.html`, html);
  }
  done += batch.length;
  process.stdout.write(`\r${done}/${all.length}`);
}
console.log("\nmissing:", missing.length, missing.slice(0, 20).join(" "));
writeFileSync(`${ROOT}/catalog.index.json`, JSON.stringify(all, null, 1));
