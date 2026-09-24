// 驗證：逐元件檢查是否為空（無文字且無影像）、尺寸異常、文字溢出。
// 用法：TK_DIR=... node 21-verify.mjs
import { connect, callTool } from "./pixso-mcp.mjs";
import { writeFileSync } from "node:fs";

const ROOT = process.env.TK_DIR || "/tmp/tailkit-pixso";
await connect();
const ev = async (s) => {
  const t = await callTool("eval_script", { script: s });
  try { return JSON.parse(t); } catch { return t; }
};

const report = await ev(`
const page = pixso.root.children.find((p) => p.name === "Tailkit 元件庫");
if (!page) return { error: "page not found" };
const issues = [];
const stats = { total: 0, empty: 0, tiny: 0, textOverflow: 0, images: 0 };
for (const sec of page.children) {
  for (const comp of sec.children) {
    if (comp.type !== "COMPONENT" && comp.type !== "COMPONENT_SET") continue;
    stats.total++;
    const texts = [], imgs = [], visuals = [];
    (function f(n) {
      if (n.type === "TEXT") texts.push(n);
      if (n.type === "VECTOR" || n.type === "RECTANGLE" || n.type === "ELLIPSE") {
        if (n.width >= 2 && n.height >= 2) visuals.push(n);
      }
      if ('fills' in n && Array.isArray(n.fills) && n.fills.some((x) => x.type === "IMAGE")) imgs.push(n);
      if ('children' in n) for (const c of n.children) f(c);
    })(comp);
    if (imgs.length) stats.images++;
    // 空元件 = 無文字、無影像、也無可見向量／色塊（純圖示或純色塊的元件是正常的）
    if (texts.length === 0 && imgs.length === 0 && visuals.length === 0) {
      stats.empty++; issues.push(comp.name + " | empty");
    }
    if (comp.height < 10) { stats.tiny++; issues.push(comp.name + " | h=" + Math.round(comp.height)); }
    for (const t of texts) {
      if (t.width > comp.width + 8) { stats.textOverflow++; issues.push(comp.name + " | text wider than comp: " + Math.round(t.width)); break; }
    }
  }
}
return { stats, issues: issues.slice(0, 60), issueCount: issues.length, sections: page.children.length };
`);
console.log(JSON.stringify(report, null, 1).slice(0, 5000));
writeFileSync(`${ROOT}/verify.json`, JSON.stringify(report, null, 1));
