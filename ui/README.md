# Flipt UI

The Flipt UI uses the [Vue.js](https://vuejs.org/) framework to build a modern single page application.

The ui directory contains these `.vue` files, along with others used in the web UI. These source files are built and packed together into simple HTML, JS, and CSS assets. For easier distribution these assets are statically compiled into the Flipt binary using the [packr](https://github.com/gobuffalo/packr) library.

## Build Requirements

* [NodeJS](https://nodejs.org/en/)
* [Yarn](https://yarnpkg.com/en/)

## Distribution

For distribution these assets must be included in the Flipt binary so that users do not need to copy these files along with the Flipt binary.

To build before committing, run:

```shell
$ make assets
```

This will build and update the generated in-line versions of the files to be included for release.
