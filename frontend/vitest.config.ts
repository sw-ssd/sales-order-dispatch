import path from "node:path";
import { defineConfig } from "vitest/config";
import solid from "vite-plugin-solid";

// vitest 專用設定:與 vite.config.ts 同源的 solid plugin 與 ~ alias,
// DOM 測試以 jsdom 環境執行。獨立檔避免改動共用 build 設定。
export default defineConfig({
  plugins: [solid()],
  resolve: {
    alias: {
      "~": path.resolve(import.meta.dirname, "src"),
    },
  },
  test: {
    environment: "jsdom",
    // @solidjs/testing-library 依全域 afterEach 自動 cleanup,避免測試間 DOM 殘留。
    globals: true,
  },
});
