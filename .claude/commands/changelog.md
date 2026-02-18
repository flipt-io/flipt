---
argument-hint: [version, e.g. 2.5.0]
---

Update the changelog for Flipt version: $ARGUMENTS

## Steps

### Step 1: Gather changes

- Identify the most recent version in CHANGELOG.md
- Gather all commits and pull requests between that version and HEAD using `git log` and `gh`
- Group changes by category (features, fixes, etc.)

### Step 2: Generate changelog entry

- Update CHANGELOG.md with the new version entry, matching the format in CHANGELOG.template.md
- Include links to the PRs that implemented each change, matching the existing style in CHANGELOG.md
- Do not include detailed dependency updates — keep them brief unless they are security-related or major version changes

### Step 3: Review

- Show the user the generated changelog entry and ask them to review it before proceeding
