// 把 scan/ 元件片段包成可獨立開啟的預覽頁 → preview/<id>.html（+ index.html 索引）。
// 產物在 TK_DIR，不入庫；樣式用相對路徑連 out3.css。
import { readFileSync, writeFileSync, mkdirSync, readdirSync, existsSync } from "node:fs";

const ROOT = process.env.TK_DIR || "/tmp/tailkit-pixso";
const SRC = existsSync(`${ROOT}/scan`) ? `${ROOT}/scan` : `${ROOT}/raw`;
const esc = (s) => s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");

const files = readdirSync(SRC).filter((f) => f.endsWith(".html"));
mkdirSync(`${ROOT}/preview`, { recursive: true });

for (const f of files) {
  const id = f.slice(0, -5);
  const body = readFileSync(`${SRC}/${f}`, "utf8");
  writeFileSync(
    `${ROOT}/preview/${f}`,
    `<!doctype html>
<html lang="zh-Hant">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>${esc(id)}</title>
<link rel="stylesheet" href="../out3.css">
</head>
<body>
${body}
</body>
</html>
`,
  );
}

// 索引：有 catalog.index.json 就依 package/category/subcategory 分組，否則平鋪。
const idxPath = `${ROOT}/catalog.index.json`;
const groups = new Map();
if (existsSync(idxPath)) {
  for (const c of JSON.parse(readFileSync(idxPath, "utf8"))) {
    const key = `${c.pkg} / ${c.cat} / ${c.sub}`;
    if (!groups.has(key)) groups.set(key, []);
    groups.get(key).push(c.id);
  }
} else {
  groups.set("", files.map((f) => f.slice(0, -5)));
}

const li = [...groups.entries()].flatMap(([key, ids]) =>
  key ? [`<h2>${esc(key)}</h2>`, ...ids.map((id) => `<li><a href="${esc(id)}.html">${esc(id)}</a></li>`)]
      : ids.map((id) => `<li><a href="${esc(id)}.html">${esc(id)}</a></li>`),
);
writeFileSync(
  `${ROOT}/preview/index.html`,
  `<!doctype html>
<html lang="zh-Hant">
<head><meta charset="utf-8"><title>Tailkit 元件預覽索引</title></head>
<body>
<h1>Tailkit 元件預覽（${files.length}）</h1>
${li.join("\n")}
</body>
</html>
`,
);
console.log("preview", files.length, "from", SRC);
