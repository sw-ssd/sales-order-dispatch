import path from "node:path";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vitest/config";
import solid from "vite-plugin-solid";

// 租戶 SPA 的目錄：console 的 UI 元件庫與元件庫內部相依的 `@/lib/cn`、`~/components/ui`
// 都在那裡。**只共用元件庫，不共用路由與路由守衛**（那兩者被 console 自己的
// src/router.tsx 與 src/lib/guard.ts 取代；S11：租戶 SPA 不得出現平台能力）。
const FRONTEND_SRC = path.resolve(import.meta.dirname, "../frontend/src");
const UI_DIR = path.join(FRONTEND_SRC, "components/ui");

export default defineConfig({
  plugins: [tailwindcss(), solid()],
  resolve: {
    // console 自家程式碼一律相對路徑匯入（`./lib/api`）；這些 alias 只為讓
    // 共用元件庫的既有 import 原樣解析：`@ui/*`（元件庫入口）、
    // `@/lib/cn`、`~/components/ui/...`（元件庫彼此的內部引用）。
    alias: {
      "@ui/": `${UI_DIR}/`,
      "@ui": path.join(UI_DIR, "index.ts"),
      "@/": `${FRONTEND_SRC}/`,
      "~/": `${FRONTEND_SRC}/`,
    },
  },
  server: {
    port: 5173,
    strictPort: true,
    // dev 時 console 與 API 不同 origin：只代理平台路徑（登入端點 ＋ RPC）。
    // 租戶 API（/api/v1）不代理 —— console 只走 platform/v1。
    proxy: {
      "/platform": { target: "http://localhost:3080", changeOrigin: true },
    },
    // 共用元件庫在 ../frontend（pnpm workspace 根目錄之外再上一層），明示允許。
    fs: { allow: [".."] },
  },
  build: { outDir: "dist" },
  // 測試設定與 alias 同源：另開 vitest.config.ts 會讓 alias 出現第二份定義而漂移。
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["src/test-setup.ts"],
    include: ["src/**/*.test.{ts,tsx}"],
  },
});
