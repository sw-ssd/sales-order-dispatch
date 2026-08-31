import path from "node:path";
import { defineConfig } from "vitest/config";
import solid from "vite-plugin-solid";

// vitest 專用設定:與 vite.config.ts 同源的 solid plugin 與 ~ alias,
// DOM 測試以 jsdom 環境執行。獨立檔避免改動共用 build 設定。
export default defineConfig({
  plugins: [solid()],
  resolve: {
    alias: {
      "~": path.resolve(__dirname, "src"),
    },
  },
  test: {
    environment: "jsdom",
  },
});
