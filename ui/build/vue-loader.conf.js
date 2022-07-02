"use strict";
import { cssLoaders } from "./utils.js";
import { dev, build } from "../config/index.js";
const isProduction = process.env.NODE_ENV === "production";
const sourceMapEnabled = isProduction
  ? build.productionSourceMap
  : dev.cssSourceMap;

export default {
  loaders: cssLoaders({
    sourceMap: sourceMapEnabled,
    extract: isProduction,
  }),
  cssSourceMap: sourceMapEnabled,
  cacheBusting: dev.cacheBusting,
  transformToRequire: {
    video: ["src", "poster"],
    source: "src",
    img: "src",
    image: "xlink:href",
  },
};
