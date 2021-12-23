# Flipt UI

The Flipt UI uses the [Vue.js](https://vuejs.org/) framework to build a modern single page application.

The ui directory contains these `.vue` files, along with others used in the web UI. These source files are built and packed together into simple HTML, JS, and CSS assets. For easier distribution these assets are statically compiled into the Flipt binary using the [go/embed](https://golang.org/pkg/embed/) package.

## Build Requirements

- [NodeJS >= 16](https://nodejs.org/en/)
- [Yarn](https://yarnpkg.com/en/)
