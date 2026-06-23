# AGENTS

Guidance for AI agents working in this repository. Full docs live at <https://docs.flipt.io> — weigh **v2** docs over v1.

Flipt v2 is a Git-native feature management platform (Go server + React UI) in a monorepo.

## Branch model (read first)

- **v2 code** lives on the `v2` branch — the **default** branch. Most work happens here.
- **v1 code** lives on `main`.
- Base PRs on the matching branch and apply the matching label (`v2` or `v1`).
- License: Fair Core License (server), MIT (client code).

## Tooling

Tool versions are pinned in `.mise.toml` (**Go 1.26**, **Node 24**). Install [mise](https://mise.jdx.dev/), then:

```bash
mise install        # install pinned Go/Node
mise run bootstrap  # install dev dependencies
```

All common tasks run through mise (`mise tasks` to list). Key ones:

| Task | Does |
|------|------|
| `mise run dev` | Run Go server (HTTP :8080, gRPC :9000) with local config |
| `mise run ui:dev` | Run UI dev server on :5173 (proxies API to :8080) |
| `mise run build` | Release-style build (Go + bundled UI assets) |
| `mise run test` | Go unit tests (`go test -v {path} -run {name}` for one) |
| `mise run lint` / `mise run fmt` | golangci-lint / goimports |
| `mise run go:modernize` | Apply modern Go idioms |
| `mise run proto` | Regenerate protobuf + gRPC stubs |
| `mise run go:mockery` | Regenerate mocks (config in `.mockery.yml`) |

Portable agent prompts for Codex and other tools live in `.agents/`: commands in `.agents/commands/` (`changelog.md`, `release.md`) and reviewer personas in `.agents/personas/` (`go-reviewer.md`).

## Detailed rules — read the relevant file before working in that area

- **UI / React / TypeScript** (components, Redux, Tailwind, API slices) → read `ui/AGENTS.md`
- **Integration tests & the Dagger build pipeline** → read `build/CLAUDE.md`

## Architecture (the non-obvious parts)

- **gRPC-first**: gRPC (:9000) is the primary interface; the REST API (:8080) is an HTTP gateway generated over it. Add/change RPCs in `rpc/`, run `mise run proto`.
- `cmd/flipt/` is the entry point. `internal/` holds private app logic (server, storage, config); `core/` holds validation/business logic; `errors/` holds shared error types.
- **OSS vs Pro boundary**: `internal/coss/` and the Pro-tier providers (e.g. `internal/secrets/` Vault provider) are commercial features gated by license — keep them isolated from OSS paths.
- **Storage** (`internal/storage/`) is multi-backend: Git repos are primary for v2; filesystem for local dev; cloud (S3/GCS/Azure) is for Flipt config only. Environments map to Git branches/directories/repos.
- **APIs**: Management CRUD, Evaluation (`/api/v1/evaluate`, OpenFeature/OFREP at `/ofrep/v1`), Analytics (`/api/v2/analytics`), Environments (`/api/v2/environments`).

## Go conventions

Follow the [Google Go Style Guide](https://google.github.io/styleguide/go/guide); `goimports` and `golangci-lint` enforce the mechanical parts. The concise project-specific rules below are the always-on essentials; for a full review checklist with examples, use `.agents/personas/go-reviewer.md`.

- **Use the custom error types** where they fit: `errs.ErrNotFoundf("flag %q not found", key)`, `errs.ErrInvalidf(...)`. Fall back to `fmt.Errorf("...: %w", err)` otherwise.
- **Prefer `const` over `var`/`:=`** for values that won't change — including in tests.
- Prefer inline error checks: `if err := doSomething(); err != nil { ... }`.
- Imports in three groups: stdlib, third-party, local (`go.flipt.io/flipt/...`).
- Structured logging with `zap`; `Debug` is the common level. Be selective.
- Keep interfaces in the domain root, implementations in subdirectories; co-locate `*_test.go` and use `testdata/` for fixtures.

## Testing

Tests must verify **behavior, not just absence of error**. Don't stop at `require.NoError`; assert the values that matter and verify data flowing through mocks with `mock.MatchedBy`, then `mock.AssertExpectations(t)`. Go tests use testcontainers for anything needing external tools.

## Before you commit

Run the relevant checks — CI enforces all of these:

```bash
mise run fmt && mise run lint && mise run go:modernize   # Go
mise run ui:fmt && mise run ui:lint                      # UI
npx prettier -w <file>.md                                # Markdown
```

Hard constraints (CI will fail otherwise):

- `go mod tidy` must leave a clean git tree — tidy and commit `go.mod`/`go.sum` changes.
- Shell scripts must pass shellcheck.
- Never hand-edit generated code (protobuf output, mocks) — regenerate it.

## Commits & PRs

- Conventional commits: `type(scope): description` (`feat`, `fix`, `docs`, `refactor`, `test`, `chore`, …).
- **Always sign commits with `-s`.**
- Keep PRs focused; explain the *why*; reference issues with `Fixes #123`.
- Base on the correct branch and label `v2`/`v1` (see Branch model above).
