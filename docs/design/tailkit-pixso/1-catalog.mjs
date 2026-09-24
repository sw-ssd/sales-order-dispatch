// 擷取 Tailkit 完整 catalog（packages → categories → subcategories → components）成 JSON。
import { writeFileSync } from "node:fs";
import { connect, callTool } from "./tailkit-mcp.mjs";

const ROOT = process.env.TK_DIR || "/tmp/tailkit-pixso";

await connect();

function parseComponents(md) {
  // 條目區塊：## `id`\n**Title**\n描述...\n- Formats: ...
  const out = [];
  const re = /^## `([^`]+)`\n\*\*(.+?)\*\*/gm;
  let m;
  while ((m = re.exec(md))) out.push({ id: m[1], title: m[2] });
  return out;
}

const pkgsJson = await callTool("browse_catalog", { level: "packages" });
const pkgSlugs = [...pkgsJson.matchAll(/`(application-ui|marketing|ecommerce)`/g)].map((m) => m[1]);
const packages = [...new Set(pkgSlugs)];

const catalog = { packages: [] };
let total = 0;

for (const pkg of packages) {
  const catMd = await callTool("browse_catalog", { level: "categories", package: pkg });
  // 只取 `- **Name** (\`slug\`)` 形式的分類列，避免撈到單字母 letter code
  const cats = [...new Set([...catMd.matchAll(/^\s*-\s+\*\*.+?\*\*\s+\(`([a-z][a-z-]+)`\)/gm)].map((m) => m[1]))];
  const pkgNode = { slug: pkg, categories: [] };
  catalog.packages.push(pkgNode);
  for (const cat of cats) {
    const subMd = await callTool("browse_catalog", { level: "subcategories", package: pkg, category: cat });
    const subs = [...new Set([...subMd.matchAll(/`([a-z0-9-]+)`/g)].map((m) => m[1]))].filter(
      (s) => !["packages", "categories", "subcategories", "components", pkg, cat].includes(s),
    );
    const catNode = { slug: cat, subcategories: [] };
    pkgNode.categories.push(catNode);
    for (const sub of subs) {
      const comps = [];
      for (let offset = 0; ; offset += 20) {
        const compMd = await callTool("browse_catalog", {
          level: "components", package: pkg, category: cat, subcategory: sub,
          ...(offset ? { offset } : {}),
        });
        const page = parseComponents(compMd);
        comps.push(...page);
        if (!/Use `offset=\d+` to see next page/.test(compMd) || page.length === 0) break;
      }
      catNode.subcategories.push({ slug: sub, components: comps });
      total += comps.length;
      console.log(`${pkg}/${cat}/${sub}: ${comps.length}`);
    }
  }
}

writeFileSync(`${ROOT}/catalog.json`, JSON.stringify(catalog, null, 1));
console.log("TOTAL", total);
