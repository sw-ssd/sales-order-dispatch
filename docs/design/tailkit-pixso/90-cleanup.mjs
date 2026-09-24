// 還原：把 Pixso 檔內所有臨時頁面清掉，只留「Tailkit 元件庫」。
import { connect, callTool } from "./pixso-mcp.mjs";
await connect();
const r = await callTool("eval_script", { script: `
const root = pixso.root;
const main = root.children.find((x) => x.name === "Tailkit 元件庫");
if (!main) return { error: "main page not found", pages: root.children.map((p) => p.name) };
pixso.currentPage = main;
const removed = [];
for (const p of root.children.slice()) {
  if (p.id === main.id) continue;
  try { p.remove(); removed.push(p.name); } catch (e) { removed.push(p.name + " FAILED"); }
}
// 清掉中轉 artboard
let stray = 0;
for (const c of main.children.slice()) if (c.name === "html") { try { c.remove(); stray++; } catch (e) {} }
return { removed, stray, pages: pixso.root.children.map((p) => p.name + ":" + p.children.length) };
` });
console.log(r);
