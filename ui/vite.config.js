import { defineConfig, loadEnv } from "vite";
import { createVuePlugin as vue } from "vite-plugin-vue2";

const path = require("path");

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  Object.assign(process.env, loadEnv(mode, process.cwd(), ""));
  // get FLIPT_SERVER_HTTP_HOST from environment variable or default to localhost
  const host = process.env.FLIPT_SERVER_HTTP_HOST || "localhost";
  // get FLIPT_SERVER_HTTP_PORT from environment variable or default to 8080
  const port = process.env.FLIPT_SERVER_HTTP_PORT || 8080;
  return {
    plugins: [vue()],
    resolve: {
      extensions: [".mjs", ".js", ".ts", ".jsx", ".tsx", ".json", ".vue"],
      alias: [
        {
          // this is required for the SCSS modules
          find: /^~(.*)$/,
          replacement: "$1",
        },
        {
          find: "@",
          replacement: path.resolve(__dirname, "./src"),
        },
      ],
    },
    server: {
      port: 8081,
      proxy: {
        "/api": `http://${host}:${port}`,
        "/meta": `http://${host}:${port}`,
      },
    },
  };
});
