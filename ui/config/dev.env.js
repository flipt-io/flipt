"use strict";
import merge from "webpack-merge";
import prodEnv from "./prod.env.js";

export default merge(prodEnv, {
  NODE_ENV: '"development"',
});
