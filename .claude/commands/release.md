---
argument-hint: [version, e.g. 2.5.0]
---

Release Flipt v2 with version: $ARGUMENTS

This is a step-by-step release process. Confirm with the user before executing each step.

## Prerequisites

- The version argument should be a semver version WITHOUT the `v` prefix (e.g. `2.5.0`)
- The working tree must be clean
- You must be on the `v2` branch and up to date with the remote

Validate all prerequisites before proceeding. If any fail, stop and tell the user.

## Steps

### Step 1: Create release branch

- Create and checkout a new branch: `release/v<version>` from `v2`
- Confirm with the user before proceeding

### Step 2: Generate changelog

- Run the `/changelog` command with the version argument to update CHANGELOG.md
- After the changelog is generated, ask the user to review it before proceeding

### Step 3: Commit changes

- Stage CHANGELOG.md (and any other modified files)
- Commit with: `git commit -s -m "chore: release v<version>"`
- Show the user the commit before proceeding

### Step 4: Push and create PR

- Push the release branch to origin
- Create a PR targeting `v2` with:
  - Title: `chore: release v<version>`
  - Body: Summary of the release (key changes from the changelog)
  - Label: `v2`
- Share the PR URL with the user

### Step 5: Wait for merge

- Tell the user to review and merge the PR
- STOP here and wait for the user to confirm the PR has been merged before continuing

### Step 6: Create annotated tag

- Switch to `v2` branch and pull latest
- Create an annotated tag: `git tag -a v<version> -m "Release v<version>"`
- Show the tag to the user and confirm before pushing

### Step 7: Push tag

- Push the tag: `git push origin v<version>`
- Confirm the tag was pushed successfully
- Tell the user that CI will now build and publish the release
