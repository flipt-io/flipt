# Flipt UI

The Flipt UI uses the [Vue.js](https://vuejs.org/) framework to build a modern single page application.

This directory contains these `.vue` files, along with others used in the web UI. These source files are built and packed together into simple HTML, JS, and CSS assets using the [Vite](https://vitejs.dev/) build tool.

For easier distribution these assets are then statically compiled into the Flipt binary using the [go/embed](https://golang.org/pkg/embed/) package.

## Build Requirements

- [NodeJS >= 18](https://nodejs.org/en/)
- [Go >= 1.17](https://golang.org/doc/install/source.html)