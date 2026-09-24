// 全量轉換（v3）：Tailkit HTML → Pixso component。
//
// 與 v2 的差別，全部來自 2026-09-23 那次「跑 48 分鐘後 plugin bridge 卡死、文件被回溯到更早版本」
// 的實際故障：
//   1. 每批寫入後**回讀驗證**：新元件必須真的出現在目標 frame 的子節點裡，才算成功。
//   2. 每批開始前**完整性閘門**：目標 page 與 88 個 section frame 必須都在，否則立即中止
//      （文件被回溯時繼續寫只會把資料寫進不存在的父層）。
//   3. manifest 只在驗證通過後落盤；啟動時先**對帳**，把節點已不存在的紀錄剔除再續跑。
//   4. 批次縮小（預設 8）、失敗退避重試，降低遠端轉換服務（code_to_design 走雲端）的斷線機率。
//
// 用法：
//   BATCH=8 LIMIT=20 node 23-convert-v3.mjs     # 試跑 20 個
//   BATCH=8 node 23-convert-v3.mjs              # 全集（可中斷續跑）
import { connect, callTool } from "./pixso-mcp.mjs";
import { readFileSync, writeFileSync, existsSync, appendFileSync } from "node:fs";

const ROOT = process.env.TK_DIR || "/tmp/tailkit-pixso";
const MANIFEST = `${ROOT}/manifest.json`;
const LOG = `${ROOT}/convert.log`;
const PAGE_NAME = "Tailkit 元件庫";
const BATCH = Number(process.env.BATCH || 8);
const LIMIT = Number(process.env.LIMIT || 0); // 0 = 不限
const RETRIES = Number(process.env.RETRIES || 3);
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

// 斷線（terminated／socket hang up）才重試；誤用／驗證失敗不重試，避免掩蓋真正的問題。
const isTransient = (e) => /terminated|socket hang up|ECONNRESET|ETIMEDOUT|429|Too Many/i.test(String(e));
async function withRetry(label, fn) {
  for (let i = 1; ; i++) {
    try { return await fn(); } catch (e) {
      if (i > RETRIES || !isTransient(e)) throw e;
      const wait = 3000 * i;
      log(`${label} retry ${i}/${RETRIES} in ${wait}ms: ${String(e).slice(0, 120)}`);
      await new Promise((r) => setTimeout(r, wait));
    }
  }
}

const subKeyOf = (it) => `${it.pkg}/${it.cat}/${it.sub}`;

const subKeys = [];
for (const it of index) {
  const k = subKeyOf(it);
  if (!subKeys.includes(k)) subKeys.push(k);
}

// ---- 目標 page 與 section frames ----
const specs = subKeys.map((key, i) => ({ key, x: (i % COLS) * 1400, y: Math.floor(i / COLS) * ROW_H }));

const pageId = await withRetry("ensurePage", () => ev(`
const root = pixso.root;
let page = root.children.find((p) => p.name === ${JSON.stringify(PAGE_NAME)});
if (!page) { page = pixso.createPage(); page.name = ${JSON.stringify(PAGE_NAME)}; }
pixso.currentPage = page;
return page.id;
`));

const frameIds = await withRetry("ensureFrames", () => ev(`
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
`));
if (!frameIds || !frameIds[subKeys[0]]) throw new Error("section frames not established");
log(`page ${pageId} frames ${Object.keys(frameIds).length}`);

// ---- 對帳：manifest 內節點已不存在者剔除（文件回溯後必經此步）----
let state = existsSync(MANIFEST) ? JSON.parse(readFileSync(MANIFEST, "utf8")) : {};
const present = new Set(JSON.parse(await withRetry("liveIds", () => ev(`
const page = pixso.getNodeById(${JSON.stringify(pageId)});
const ids = [];
for (const sec of page.children) for (const c of sec.children) ids.push(c.id);
return ids;
`))) || []);
const before = Object.keys(state).length;
for (const [id, v] of Object.entries(state)) if (!present.has(v.pixso_id)) delete state[id];
if (Object.keys(state).length !== before) {
  log(`reconcile: dropped ${before - Object.keys(state).length} stale entries, kept ${Object.keys(state).length}`);
  writeFileSync(MANIFEST, JSON.stringify(state, null, 1));
}
log(`resume ${Object.keys(state).length} / ${index.length}`);

// ---- 完整性閘門：每批開始前確認 page 與 frames 仍在 ----
async function assertIntegrity() {
  const r = JSON.parse(await ev(`
const root = pixso.root;
const page = root.children.find((p) => p.name === ${JSON.stringify(PAGE_NAME)});
if (!page) return { ok: false, why: "page missing" };
const names = page.children.map((c) => c.name);
const missing = ${JSON.stringify(subKeys)}.filter((k) => !names.includes(k));
return { ok: missing.length === 0, why: missing.length ? "frames missing: " + missing.slice(0, 3).join(",") : "" };
`));
  if (!r.ok) throw new Error(`DOCUMENT UNSTABLE: ${r.why}`);
}

// ---- 送件與歸位（含回讀驗證）----
async function placeSlots(artId, items) {
  const map = items.map((it, i) => ({ id: it.id, slot: `div.tk-slot-${i}`, frame: frameIds[subKeyOf(it)] }));
  const res = await ev(`
const art = pixso.getNodeById(${JSON.stringify(artId)});
if (!art) return { error: "artboard missing" };
const out = [];
for (const m of ${JSON.stringify(map)}) {
  let slot = null;
  (function f(n) { if (slot) return; if (n.name === m.slot) { slot = n; return; } if ('children' in n) for (const c of n.children) f(c); })(art);
  if (!slot) { out.push({ id: m.id, error: "slot not found" }); continue; }
  const frame = pixso.getNodeById(m.frame);
  if (!frame) { out.push({ id: m.id, error: "frame missing" }); continue; }
  const w = Math.round(slot.width), h = Math.round(slot.height);
  frame.appendChild(slot);
  slot.name = m.id;
  const comp = pixso.createComponentFromNode(slot);
  comp.name = m.id;
  comp.setPluginData("tk_id", m.id);
  out.push({ id: m.id, pixso_id: comp.id, w, h });
}
try { art.remove(); } catch (e) {}
return out;
`);
  if (!Array.isArray(res)) throw new Error(`placeSlots: ${JSON.stringify(res).slice(0, 160)}`);

  // 回讀驗證：新元件必須真的掛在目標 frame 底下才算成功
  const expect = res.filter((x) => !x.error);
  const verify = JSON.parse(await ev(`
const ids = ${JSON.stringify(expect.map((x) => x.pixso_id))};
const frames = ${JSON.stringify([...new Set(expect.map((x) => frameIds[subKeyOf(items.find((i) => i.id === x.id))]))])};
const have = new Set();
for (const fid of frames) {
  const f = pixso.getNodeById(fid);
  if (f) for (const c of f.children) have.add(c.id);
}
return ids.filter((id) => !have.has(id));
`));
  if (verify.length) throw new Error(`verify failed: ${verify.length} component(s) not attached`);
  return res.filter((x) => !x.error);
}

async function convertOne(it) {
  const r = await callTool("code_to_design", {
    htmlStr: buildDoc([readFileSync(`${ROOT}/scan/${it.id}.html`, "utf8")]),
    width: VIEW_W, height: VIEW_H,
  });
  const artId = JSON.parse(r).id;
  const res = await placeSlots(artId, [it]);
  const x = res[0];
  if (!x || x.error) throw new Error(x?.error ?? "no result");
  return { ...x, title: it.title, pkg: it.pkg, cat: it.cat, sub: it.sub };
}

// ---- 待辦分組（同 subcategory 相鄰，維持變體順序）----
let pending = index.filter((it) => !state[it.id]);
if (LIMIT) pending = pending.slice(0, LIMIT);
log(`pending ${pending.length}${LIMIT ? ` (LIMIT=${LIMIT})` : ""} / ${index.length}`);

const groups = new Map();
for (const it of pending) {
  const k = subKeyOf(it);
  if (!groups.has(k)) groups.set(k, []);
  groups.get(k).push(it);
}

let processed = 0;
for (const [key, items] of groups) {
  for (let i = 0; i < items.length; i += BATCH) {
    await assertIntegrity();
    const chunk = items.slice(i, i + BATCH);
    let handled = false;
    let artId = null;
    try {
      const r = await withRetry(`c2d:${key}`, () => callTool("code_to_design", {
        htmlStr: buildDoc(chunk.map((it) => readFileSync(`${ROOT}/scan/${it.id}.html`, "utf8"))),
        width: VIEW_W, height: VIEW_H,
      }));
      artId = JSON.parse(r).id;
      const res = await placeSlots(artId, chunk);
      for (const x of res) {
        const it = chunk.find((c) => c.id === x.id);
        state[x.id] = { ...x, title: it.title, pkg: it.pkg, cat: it.cat, sub: it.sub };
      }
      handled = true;
    } catch (e) {
      log(`batch fail ${key}: ${String(e).slice(0, 200)}`);
      if (String(e).includes("DOCUMENT UNSTABLE")) throw e; // 文件出事 → 立即中止
      if (artId) await ev(`const n = pixso.getNodeById(${JSON.stringify(artId)}); if (n) n.remove(); return 1;`).catch(() => {});
    }
    if (!handled) {
      // 批次失敗 → 逐件重試（仍逐一驗證）
      for (const it of chunk) {
        try {
          const x = await withRetry(`one:${it.id}`, () => convertOne(it));
          state[x.id] = x;
        } catch (e) {
          log(`single fail ${it.id}: ${String(e).slice(0, 160)}`);
          if (String(e).includes("DOCUMENT UNSTABLE")) throw e;
        }
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
