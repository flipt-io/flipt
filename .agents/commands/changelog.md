# Changelog Command

Update the changelog for a Flipt release.

## Required Input

- `version`: semver version without a `v` prefix, for example `2.5.0`

## Task

Update `CHANGELOG.md` for Flipt version `<version>`.

## Workflow

### 1. Gather changes

- Identify the most recent version already present in `CHANGELOG.md`.
- Gather all commits and pull requests between that version and `HEAD` using
  `git log` and `gh`.
- Group changes by category such as features, fixes, docs, refactors, and
  maintenance.

### 2. Generate the entry

- Update `CHANGELOG.md` with a new version entry.
- Match the existing style in `CHANGELOG.md` and the structure in
  `CHANGELOG.template.md`.
- Include pull request links for each item when available.
- Keep dependency updates brief unless they are security-related or major
  version changes.

### 3. Review checkpoint

- Show the generated changelog entry to the user.
- Stop and ask the user to review it before making any follow-up release
  changes.

## Output Expectations

- A proposed `CHANGELOG.md` diff or summary of the new entry.
- Any assumptions or missing PR metadata that prevented a complete entry.
