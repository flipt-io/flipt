# Release Command

Run the Flipt v2 release process for a specific version.

This is a step-by-step workflow. Validate prerequisites first, then stop at
each checkpoint that requires user confirmation.

## Required Input

- `version`: semver version without a `v` prefix, for example `2.5.0`

## Task

Release Flipt v2 version `<version>`.

## Prerequisites

Validate all of these before proceeding. If any check fails, stop and tell the
user exactly what needs to be fixed.

- The version input is valid semver without the `v` prefix.
- The git working tree is clean.
- The current branch is `v2`.
- The local `v2` branch is up to date with `origin/v2`.

## Workflow

### 1. Create the release branch

- Create and check out `release/v<version>` from `v2`.
- Stop and confirm with the user before moving on.

### 2. Generate the changelog

- Use `.agents/commands/changelog.md` with the same version input to update
  `CHANGELOG.md`.
- After generating the changelog, stop and ask the user to review it before
  proceeding.

### 3. Commit the release changes

- Stage `CHANGELOG.md` and any other intended release files.
- Commit with:

```text
git commit -s -m "chore: release v<version>"
```

- Show the commit details to the user.
- Stop and confirm before proceeding.

### 4. Push and open the pull request

- Push the release branch to `origin`.
- Create a pull request targeting `v2` with:
  - Title: `chore: release v<version>`
  - Body: a short summary of the release based on the changelog
  - Label: `v2`
- Share the PR URL with the user.

### 5. Wait for merge

- Tell the user to review and merge the PR.
- Stop here and wait until the user confirms the PR has been merged.

### 6. Create the annotated tag

- Switch back to `v2`.
- Pull the latest changes.
- Create the annotated tag:

```text
git tag -a v<version> -m "Release v<version>"
```

- Show the tag details to the user.
- Stop and confirm before pushing the tag.

### 7. Push the tag

- Push the tag to `origin`.
- Confirm the push succeeded.
- Tell the user that CI should now build and publish the release artifacts.

## Agent Notes

- Do not skip confirmation checkpoints.
- Do not continue after the PR is opened until the user confirms it was merged.
- If repository policy conflicts with any step, surface the conflict before
  proceeding.
