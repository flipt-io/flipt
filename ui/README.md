# Flipt UI

The Flipt UI uses the [Vue.js](https://vuejs.org/) framework to build a modern single page application.

The ui directory contains these `.vue` files, along with others used in the web UI. These source files are built and packed together into simple HTML, JS, and CSS assets. For easier distribution these assets are statically compiled into the Flipt binary using the [vfsgen](https://github.com/shurcooL/vfsgen) library.

## Build Requirements

* [NodeJS](https://nodejs.org/en/)
* [Yarn](https://yarnpkg.com/en/)

## Development

During development it is more convenient to always use the files on disk to directly see changes without recompiling. To make this work, run `make dev` after `make ui`.

Example:

```shell
$ make ui
$ make dev
```

This would run Flipt in development mode and serve the UI assets directly from the filesystem.

## Distribution

For distribution these assets must be included in the Flipt binary so that users do not need to copy these files along with the Flipt binary.

To build before committing, run:

```shell
$ make ui
$ make generate
```

This will build and update the generated inline versions of the files to be included for release.
