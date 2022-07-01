import { defineConfig } from "vite";
import path from "path";
import { createVuePlugin as vue } from "vite-plugin-vue2";
import envCompatible from "vite-plugin-env-compatible";
import { injectHtml } from "vite-plugin-html";

// https://vitejs.dev/config/
export default defineConfig({
  resolve: {
    alias: [
      {
        find: /^~/,
        replacement: "",
      },
      {
        find: "@",
        replacement: path.resolve(__dirname, "src"),
      },
    ],
    extensions: [".mjs", ".js", ".ts", ".jsx", ".tsx", ".json", ".vue"],
  },
  plugins: [vue(), envCompatible(), injectHtml()],
  build: {},
});
