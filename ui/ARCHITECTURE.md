# Flipt UI Architecture

## Overview

This document provides an overview of the key elements and structure of the Flipt React/Typescript client application (UI).

The application is a modern web application built using React and TypeScript with Redux for global state management.

Read more about the reasoning behind the choice of React/TypeScript on our [blog](https://www.flipt.io/blog/new-look-and-authentication-options).

## Technology Stack

- [React](https://reactjs.org/)
- [TypeScript](https://www.typescriptlang.org/)
- [Redux](https://redux.js.org/)
- [React Router](https://reactrouter.com/en/main)
- [Tailwind CSS](https://tailwindcss.com/)
- [Jest](https://jestjs.io/)
- [Playwright](https://playwright.dev/)

## Key Elements

1. **App**: These files contain the 'pages for the application. This folder is further divided into subfolders for each page. The `App.tsx` file is the entry point for the application. It contains the routes for the application and the top-level components that are rendered on each page.

   We use [React Router](https://reactrouter.com/web/guides/quick-start) for routing. [HashRouter](https://reactrouter.com/web/api/HashRouter) is used for client-side routing to maintain backward compatability with the previous version of the UI.

2. **Components**: These are reusable pieces of code that return a React element to be rendered to the page. The components for this application are located in the `components` directory. This directory is further divided into subdirectories which roughly correspond to the pages of the application. For example, the `namespaces` directory contains components that are used on the Namespaces page.

3. **State**: Redux is used for global state management in the application. The `store.ts` file is where the Redux store is configured. This store is the single source of truth for global state data in the application (ie: `namespacesSlice.ts` contains the Redux slice for managing the namespaces state).

   Most other state in the application is managed locally by React components via the `useState` hook and the [`Context`](https://react.dev/learn/passing-data-deeply-with-context) API.

4. **Data**: This folder contains `api.ts` which is our thin wrapper around the [Flipt REST API](https://www.flipt.io/docs/reference/overview). It also contains [custom `hooks/`](https://react.dev/learn/reusing-logic-with-custom-hooks#extracting-your-own-custom-hook-from-a-component) and other 'data' related functionality such as `validations.ts` for Yup validation.

5. **Types**: TypeScript types are used throughout the application to enforce type safety.

6. **Utils**: The `utils` directory contains utility functions that are used across the application.

## Authentication

TODO

## Development

See: <https://github.com/flipt-io/flipt/blob/main/DEVELOPMENT.md#ui> for instructions on how to run the UI locally.

### Testing

We use [Jest](https://jestjs.io/) for unit testing and [Playwright](https://playwright.dev/) for end-to-end testing. The unit tests are located alongside the code they test. The end-to-end tests are located in the `tests` directory.

The end-to-end tests are run in a Docker container via Dagger. To run the tests locally, you will need to have Docker installed. You can run the tests locally by running the following command from the root of the repository:

    mage dagger:run test:ui

### Linting

We use [ESLint](https://eslint.org/) for linting. The configuration for ESLint is located in the `.eslintrc.json`. The linting rules are enforced via CI (GitHub Actions) and locally via the `lint` script in the `package.json` file (`npm run lint`).

### Formatting

We use [Prettier](https://prettier.io/) for formatting. The configuration for Prettier is located in the `prettier.config.cjs` file. Run `npm run format` to format the code.
