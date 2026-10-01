import { defineConfig } from "vite";
import solid from "vite-plugin-solid";
import tailwindcss from "@tailwindcss/vite";
import path from "node:path";

export default defineConfig({
  plugins: [tailwindcss(), solid()],
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "src"),
      "~": path.resolve(import.meta.dirname, "src"),
      "@ui": path.resolve(import.meta.dirname, "../packages/ui/src/ui"),
      "@salesorder/ui": path.resolve(import.meta.dirname, "../packages/ui/src/ui"),
      "~/components/ui": path.resolve(import.meta.dirname, "../packages/ui/src/ui"),
      "~/components/ui/": path.resolve(import.meta.dirname, "../packages/ui/src/ui/"),
    },
  },
  server: {
    port: 3000,
    proxy: {
      "/api": {
        target: "http://localhost:3080",
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: "dist",
  },
});
