// 編譯 Tailkit 元件所需的 Tailwind v4 CSS（含 secondary 色階與 8xl/9xl/10xl 容器上限）。
import { execFileSync } from "node:child_process";

const ROOT = process.env.TK_DIR || "/tmp/tailkit-pixso";
execFileSync("bunx", ["@tailwindcss/cli@4.3.3", "-i", "tailkit.css", "-o", "out3.css"], {
  cwd: ROOT, stdio: "inherit",
});
