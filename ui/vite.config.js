import { defineConfig } from "vite";
import { createVuePlugin as vue } from "vite-plugin-vue2";

const path = require("path");

// https://vitejs.dev/config/
export default defineConfig(() => {
  // get FLIPT_HTTP_SERVER_HOST from environment variable or default to localhost
  const host = process.env.FLIPT_HTTP_SERVER_HOST || "localhost";
  // get FLIPT_HTTP_SERVER_PORT from environment variable or default to 8080
  const port = process.env.FLIPT_HTTP_SERVER_PORT || 8080;
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
      proxy: {
        "/api": `http://${host}:${port}`,
        "/meta": `http://${host}:${port}`,
      },
    },
  };
});
