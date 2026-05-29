# AGENTS — UI

The Flipt UI: a React 19 + TypeScript single-page app that is **embedded into the Flipt server binary** for production. Follow the root `AGENTS.md` for repo-wide rules; this file covers the `ui/` package.

## Stack

React 19 · TypeScript · Redux Toolkit + RTK Query · React Router v7 (hash router) · Tailwind CSS **v4** · Radix UI primitives · Vite 6 · Jest (unit) + Playwright (e2e) · ESLint + Prettier.

## Commands

Node 20 (pinned in root `.mise.toml`). From the repo root:

```bash
mise run ui:dev    # dev server on :5173, proxies API to backend :8080
mise run ui:build  # production build (embedded into the server binary)
mise run ui:fmt    # prettier
mise run ui:lint   # eslint
```

Or from `ui/`: `npm run dev | build | lint | format | test | knip`.
`npm run build` runs `tsc` first, so **type errors fail the build**.

## Architecture (the non-obvious parts)

- **`src/app/<feature>/`** = feature/route modules — pages, layouts, and the feature's **co-located RTK Query slice** (`flagsApi.ts`, `authApi.ts`, `environmentsApi.ts`, …). This is where most logic lives.
- **`src/components/`** = shared, reusable components (built on Radix UI primitives).
- **`src/data/`** = cross-cutting data layer: `api.ts` (shared base URLs, `APIError`, fetch/browser helpers, session) and `hooks/` (shared hooks like `useChangesStream`).
- **`src/store.ts`** = Redux store (`configureStore`); `src/types/`, `src/utils/`, `src/hooks/`.
- **Routing**: `createHashRouter` from `react-router` (hash-based URLs).
- **Path alias**: `~/` maps to `src/` (e.g. `import { FlagType } from '~/types/Flag'`).
- **Backend API is v2-first**: base paths `api/v2/environments`, `auth/v1`, `evaluate/v1`, `internal/v2`, `meta/info`. Requests carry an `x-csrf-token` header.
- **Real-time** flag changes use **SSE** (`EventSource` in `useChangesStream`), not WebSocket.
- **Styling**: Tailwind **v4** — config is CSS-first in `src/index.css`; there is **no `tailwind.config.js`**. Dark mode is driven by the Redux `preferencesSlice` theme.

## Conventions

Concise essentials below; for a full review checklist with examples use `.agents/personas/react-reviewer.md`.

- Functional components with hooks; strict TypeScript, avoid `any`; type all props and API responses.
- RTK Query slices co-located per feature under `src/app/<feature>/<feature>Api.ts`; invalidate cache tags after mutations.
- Every API call handles loading and error states.
- File naming: Components `PascalCase.tsx`, hooks `useThing.ts`, API slices `thingApi.ts`, types `Thing.ts`, utils `camelCase.ts`.
- Accessibility: ARIA labels and keyboard navigation.

## Testing

- Unit (Jest): `cd ui && npm run test` (config in `jest.config.ts`).
- E2E (Playwright): specs in `ui/tests/*.spec.ts`, run with `npx playwright test` (requires a running backend).
- `npm run knip` finds unused files/exports/deps.

## Before you commit

`mise run ui:fmt && mise run ui:lint` — both are enforced by CI (the UI lint job runs eslint + prettier check).

## Notes

- Production embeds the built assets into the server binary; dev runs separately and proxies to :8080.
- Maintain backward compatibility with the v1 API where the UI still depends on it.
