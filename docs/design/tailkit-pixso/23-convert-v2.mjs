// 全量轉換：Tailkit HTML → Pixso component。
// 每個 subcategory 一個 section frame，逐批（BATCH 個元件一次）呼叫 code_to_design，再拆解歸位。
// 支援中斷續跑：進度寫入 manifest.json，已完成的 id 不重做。
import { connect, callTool } from "./pixso-mcp.mjs";
import { readFileSync, writeFileSync, existsSync, appendFileSync } from "node:fs";

const ROOT = process.env.TK_DIR || "/tmp/tailkit-pixso";
const MANIFEST = `${ROOT}/manifest.json`;
const LOG = `${ROOT}/convert.log`;
const PAGE_NAME = "Tailkit 元件庫";
const BATCH = Number(process.env.BATCH || 20);
const PARALLEL = Number(process.env.PARALLEL || 3);
const VIEW_W = 1280;
const VIEW_H = 900;
const ROW_H = 4000;
const COLS = 12;

const log = (m) => appendFileSync(LOG, `${new Date().toISOString()} ${m}\n`);

const css = readFileSync(`${ROOT}/out3.css`, "utf8");
const index = JSON.parse(readFileSync(`${ROOT}/catalog.index.json`, "utf8"));

function buildDoc(htmlSlots) {
  const body = htmlSlots.map((h, i) => `<div id="tk-slot-${i}">${h}</div>`).join("\n");
  // flow-root：每個 slot 自成 BFC，否則 float／margin 塌陷會讓 slot 高度小於內容（dividers 28px vs 92px）
  const slotCss = htmlSlots.map((_, i) => `#tk-slot-${i}`).join(",") + "{display:flow-root}";
  return `<!doctype html><html><head><meta charset="utf-8"><style>*{box-sizing:border-box}body{margin:0;padding:16px;font-family:Inter,system-ui,sans-serif;background:#fff}${css}${slotCss}</style></head><body>${body}</body></html>`;
}

await connect();
const ev = async (script) => {
  const t = await callTool("eval_script", { script });
  try { return JSON.parse(t); } catch { return t; }
};

const state = existsSync(MANIFEST) ? JSON.parse(readFileSync(MANIFEST, "utf8")) : {};

// ---- page + section frames ----
const subKeys = [];
for (const it of index) {
  const k = `${it.pkg}/${it.cat}/${it.sub}`;
  if (!subKeys.includes(k)) subKeys.push(k);
}

const pageId = await ev(`
const root = pixso.root;
let page = root.children.find((p) => p.name === ${JSON.stringify(PAGE_NAME)});
if (!page) { page = pixso.createPage(); page.name = ${JSON.stringify(PAGE_NAME)}; }
pixso.currentPage = page;
return page.id;
`);

const specs = subKeys.map((key, i) => ({
  key,
  x: (i % COLS) * 1400,
  y: Math.floor(i / COLS) * ROW_H,
}));

const frameIds = await ev(`
const page = pixso.getNodeById(${JSON.stringify(pageId)});
const made = {};
for (const sp of ${JSON.stringify(specs)}) {
  let f = page.children.find((c) => c.name === sp.key);
  if (!f) {
    f = pixso.createFrame();
    f.name = sp.key;
    page.appendChild(f);
    f.fills = [];
    f.layoutMode = "VERTICAL";
    f.itemSpacing = 48;
    f.paddingTop = 48; f.paddingBottom = 48; f.paddingLeft = 40; f.paddingRight = 40;
    f.primaryAxisSizingMode = "AUTO";
    f.counterAxisSizingMode = "AUTO";
  }
  f.x = sp.x; f.y = sp.y;
  f.setPluginData("tk_section", sp.key);
  made[sp.key] = f.id;
}
return made;
`);
log(`page ${pageId} frames ${Object.keys(frameIds).length} resume ${Object.keys(state).length}`);

const subKeyOf = (it) => `${it.pkg}/${it.cat}/${it.sub}`;
const pending = index.filter((it) => !state[it.id]);
log(`pending ${pending.length} / ${index.length}`);

const groups = new Map();
for (const it of pending) {
  const k = subKeyOf(it);
  if (!groups.has(k)) groups.set(k, []);
  groups.get(k).push(it);
}

async function placeSlots(artId, items) {
  const map = items.map((it, i) => ({ id: it.id, slot: `div.tk-slot-${i}`, frame: frameIds[subKeyOf(it)] }));
  const res = await ev(`
const art = pixso.getNodeById(${JSON.stringify(artId)});
const out = [];
for (const m of ${JSON.stringify(map)}) {
  let slot = null;
  (function f(n) { if (slot) return; if (n.name === m.slot) { slot = n; return; } if ('children' in n) for (const c of n.children) f(c); })(art);
  if (!slot) { out.push({ id: m.id, error: "slot not found" }); continue; }
  const frame = pixso.getNodeById(m.frame);
  const w = Math.round(slot.width), h = Math.round(slot.height);
  frame.appendChild(slot);
  slot.name = m.id;
  const comp = pixso.createComponentFromNode(slot);
  comp.name = m.id;
  comp.setPluginData("tk_id", m.id);
  out.push({ id: m.id, pixso_id: comp.id, w, h });
}
// 拆解後移除中轉 artboard，避免殘留空框
try { art.remove(); } catch (e) {}
return out;
`);
  // eval_script 失敗時回傳的是 { error } 物件而非陣列；呼叫端據此走逐件重試
  if (!Array.isArray(res)) throw new Error(`placeSlots failed: ${JSON.stringify(res).slice(0, 200)}`);
  return res;
}

async function convertOne(it) {
  const r = await callTool("code_to_design", {
    htmlStr: buildDoc([readFileSync(`${ROOT}/scan/${it.id}.html`, "utf8")]),
    width: VIEW_W, height: VIEW_H,
  });
  const aid = JSON.parse(r).id;
  const res = await placeSlots(aid, [it]);
  const x = res[0];
  if (!x || x.error) throw new Error(x?.error ?? "no result");
  return { ...x, title: it.title, pkg: it.pkg, cat: it.cat, sub: it.sub };
}

let processed = 0;
for (const [key, items] of groups) {
  for (let i = 0; i < items.length; i += BATCH) {
    const chunk = items.slice(i, i + BATCH);
    let handled = false;
    let artId = null;
    try {
      const r = await callTool("code_to_design", {
        htmlStr: buildDoc(chunk.map((it) => readFileSync(`${ROOT}/scan/${it.id}.html`, "utf8"))),
        width: VIEW_W, height: VIEW_H,
      });
      artId = JSON.parse(r).id;
      const res = await placeSlots(artId, chunk);
      for (const x of res) {
        const it = chunk.find((c) => c.id === x.id);
        if (x.error) { log(`slot missing ${x.id}`); continue; }
        state[x.id] = { ...x, title: it.title, pkg: it.pkg, cat: it.cat, sub: it.sub };
      }
      handled = true;
    } catch (e) {
      log(`batch fail ${key}: ${String(e).slice(0, 200)}`);
      // 拆解失敗會留下孤立的 html artboard，直接刪掉避免殘留
      if (artId) await ev(`const n = pixso.getNodeById(${JSON.stringify(artId)}); if (n) n.remove(); return 1;`);
    }
    if (!handled) {
      // 批次失敗 → 平行逐個重試
      for (let j = 0; j < chunk.length; j += PARALLEL) {
        const sub = chunk.slice(j, j + PARALLEL);
        const results = await Promise.all(sub.map((it) => convertOne(it).catch((e) => {
          log(`single fail ${it.id}: ${String(e).slice(0, 200)}`);
          return null;
        })));
        for (const x of results) if (x) state[x.id] = x;
      }
    }
    processed += chunk.length;
    writeFileSync(MANIFEST, JSON.stringify(state, null, 1));
    log(`ok ${key} ${processed}/${pending.length} total=${Object.keys(state).length}`);
    process.stdout.write(`\r${Object.keys(state).length}/${index.length} ${key}   `);
  }
}
console.log("\nDONE", Object.keys(state).length);
log(`DONE ${Object.keys(state).length}`);
