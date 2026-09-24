// 去除 Tailkit 專有 icon font class（Pixso 無此字型），保留 inline SVG。
// 產出 scan/ 供 Tailwind 掃描與後續轉換。
import { readFileSync, writeFileSync, mkdirSync, readdirSync } from "node:fs";

const ROOT = process.env.TK_DIR || "/tmp/tailkit-pixso";
const ICON_CLASS = /\s*hi-(micro|mini|solid|outline|regular)\s+hi-[a-z0-9-]+/g;

mkdirSync(`${ROOT}/scan`, { recursive: true });
let n = 0;
for (const f of readdirSync(`${ROOT}/raw`)) {
  if (!f.endsWith(".html")) continue;
  const t = readFileSync(`${ROOT}/raw/${f}`, "utf8");
  writeFileSync(`${ROOT}/scan/${f}`, t.replace(ICON_CLASS, ""));
  n++;
}
console.log("cleaned", n);
