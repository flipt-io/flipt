# Flipt Issue Health

Flipt is a Git-native feature management platform (Go server + React UI).
Analyze newly opened issues for actionability and routing, not code review.

## Versions

Flipt v2 is the current focus and lives on the `v2` (default) branch; v1 is in
maintenance on `main`. Apply `v2` or `v1` when the issue clearly concerns one
version (e.g. mentions Git-native environments, v2 config, or a specific
version string); default to `v2` when unstated. Point reporters at
<https://docs.flipt.io>, preferring v2 docs.

## Routing

If `targetRepoLabels` contains a matching label, suggest it:

- `bug` for defect reports. Ask for missing reproduction steps, Flipt version,
  and deployment/storage backend, and suggest `needs more info` when a bug
  report lacks them.
- `enhancement` for feature requests, and `proposal` for open-ended ideas or
  design discussions.
- `ui` for the React UI, `go` for the Go server/API, and `dx` for developer
  experience or tooling.
- `needs docs` when the request implies documentation updates.

Do not suggest `good first issue` or `help wanted` — maintainers triage those.

## Support

Direct usage questions and general help to the Discord community at
<https://flipt.io/discord> rather than the issue tracker.
