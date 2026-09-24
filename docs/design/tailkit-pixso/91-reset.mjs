// 清空 Pixso 檔：只留單一目標 page，刪掉所有頁面與其內容。
// 用法：node 91-reset.mjs [keepPageName]     （keepPageName 預設「页面 1」）
// 設計：先快取名字字串，再刪除節點 —— 節點被 remove() 後再讀 .name 會拋 "No node with the given GUID!"。
import { connect, callTool } from "./pixso-mcp.mjs";

const KEEP = process.argv[2] || "页面 1";

await connect();
const r = await callTool("eval_script", { script: `
const root = pixso.root;
const names = root.children.map((p) => ({ id: p.id, name: p.name }));
let keep = root.children.find((p) => p.name === ${JSON.stringify(KEEP)}) || root.children[0];
pixso.currentPage = keep;
const log = [];
for (const it of names) {
  if (it.id === keep.id) continue;
  try { const p = pixso.getNodeById(it.id); if (p) { p.remove(); log.push("page " + it.name); } }
  catch (e) { log.push("pageFAIL " + it.name + " :: " + e.message); }
}
const kids = keep.children.map((c) => ({ id: c.id, name: c.name }));
for (const it of kids) {
  try { const c = pixso.getNodeById(it.id); if (c) { c.remove(); log.push("frame " + it.name); } }
  catch (e) { log.push("frameFAIL " + it.name + " :: " + e.message); }
}
return { kept: keep.name, log, pages: pixso.root.children.map((p) => p.name + ":" + p.children.length) };
` });
console.log(r);
