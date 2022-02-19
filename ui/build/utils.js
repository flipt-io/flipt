"use strict";
import path from "path";
import { dev, build } from "../config/index.js";
import notifier from "node-notifier";

export const assetsPath = function (_path) {
  const assetsSubDirectory =
    process.env.NODE_ENV === "production"
      ? build.assetsSubDirectory
      : dev.assetsSubDirectory;

  return path.posix.join(assetsSubDirectory, _path);
};

export const cssLoaders = function (options) {
  options = options || {};

  const cssLoader = {
    loader: "css-loader",
    options: {
      sourceMap: options.sourceMap,
    },
  };

  const postcssLoader = {
    loader: "postcss-loader",
    options: {
      sourceMap: options.sourceMap,
    },
  };

  // generate loader string to be used with extract text plugin
  function generateLoaders(loader, loaderOptions) {
    const loaders = options.usePostCSS
      ? [cssLoader, postcssLoader]
      : [cssLoader];

    if (loader) {
      loaders.push({
        loader: loader + "-loader",
        options: Object.assign({}, loaderOptions, {
          sourceMap: options.sourceMap,
        }),
      });
    }

    // If extract is used, then rely on miniCSS being set in production config
    if (!options.extract) {
      return ["vue-style-loader"].concat(loaders);
    }
    return loaders;
  }

  // https://vue-loader.vuejs.org/en/configurations/extract-css.html
  return {
    css: generateLoaders(),
    postcss: generateLoaders(),
    less: generateLoaders("less"),
    sass: generateLoaders("sass", { indentedSyntax: true }),
    scss: generateLoaders("sass"),
    stylus: generateLoaders("stylus"),
    styl: generateLoaders("stylus"),
  };
};

// Generate loaders for standalone style files (outside of .vue)
export const styleLoaders = function (options) {
  const output = [];
  const loaders = cssLoaders(options);

  for (const extension in loaders) {
    const loader = loaders[extension];
    output.push({
      test: new RegExp("\\." + extension + "$"),
      use: loader,
    });
  }
  return output;
};

export const createNotifierCallback = () => {
  return (severity, errors) => {
    if (severity !== "error") return;

    const error = errors[0];
    const filename = error.file && error.file.split("!").pop();

    notifier.notify({
      title: "flipt-ui",
      message: severity + ": " + error.name,
      subtitle: filename || "",
      icon: path.join(__dirname, "logo.png"),
    });
  };
};
