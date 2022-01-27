"use strict";
import { createNotifierCallback, styleLoaders } from "./utils.js";
import webpack from "webpack";
import { dev } from "../config/index.js";
import merge from "webpack-merge";
import path from "path";
import { fileURLToPath } from "url";
import baseWebpackConfig from "./webpack.base.conf.js";
import CopyWebpackPlugin from "copy-webpack-plugin";
import HtmlWebpackPlugin from "html-webpack-plugin";
import FriendlyErrorsPlugin from "friendly-errors-webpack-plugin";
import portfinder from "portfinder";
import env from "../config/dev.env.js";

const HOST = process.env.HOST;
const PORT = process.env.PORT && Number(process.env.PORT);

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const devWebpackConfig = merge(baseWebpackConfig, {
  mode: "development",
  module: {
    rules: styleLoaders({
      sourceMap: dev.cssSourceMap,
      usePostCSS: true,
    }),
  },
  // cheap-module-eval-source-map is faster for development
  devtool: dev.devtool,

  // these devServer options should be customized in /config/index.js
  devServer: {
    historyApiFallback: {
      rewrites: [
        {
          from: /.*/,
          to: path.posix.join(dev.assetsPublicPath, "index.html"),
        },
      ],
    },
    hot: true,
    compress: true,
    host: HOST || dev.host,
    port: PORT || dev.port,
    open: dev.autoOpenBrowser,
    proxy: [
      {
        context: ["/api", "/meta"],
        target: "http://localhost:8080",
      },
    ],
  },
  plugins: [
    new webpack.DefinePlugin({
      "process.env": env,
    }),
    new webpack.HotModuleReplacementPlugin(),
    new webpack.NamedModulesPlugin(), // HMR shows correct file names in console on update.
    new webpack.NoEmitOnErrorsPlugin(),
    // https://github.com/ampedandwired/html-webpack-plugin
    new HtmlWebpackPlugin({
      filename: "index.html",
      template: "index.html",
      inject: true,
    }),
    // copy custom static assets
    new CopyWebpackPlugin({
      patterns: [
        {
          from: path.resolve(__dirname, "../static"),
          to: dev.assetsSubDirectory,
          globOptions: {
            ignore: [".*"],
          },
        },
      ],
    }),
  ],
});

export default new Promise((resolve, reject) => {
  portfinder.basePort = process.env.PORT || dev.port;
  portfinder.getPort((err, port) => {
    if (err) {
      reject(err);
    } else {
      // publish the new Port, necessary for e2e tests
      process.env.PORT = port;
      // add port to devServer config
      devWebpackConfig.devServer.port = port;

      // Add FriendlyErrorsPlugin
      devWebpackConfig.plugins.push(
        new FriendlyErrorsPlugin({
          compilationSuccessInfo: {
            messages: [
              `Your application is running here: http://${devWebpackConfig.devServer.host}:${port}`,
            ],
          },
          onErrors: dev.notifyOnErrors ? createNotifierCallback() : undefined,
        })
      );

      resolve(devWebpackConfig);
    }
  });
});
