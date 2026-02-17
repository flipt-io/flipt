---
argument-hint: [version]
---

Please update the current changelog @CHANGELOG.md with details about this upcoming version.

version: $ARGUMENTS

## Steps

- Gather the commits/pull requests between the current top version in @CHANGELOG.md and the
  new version
- Update the changelog accordingly, matching the format in @CHANGELOG.template.md
- Do not include detailed dependency updates! Keep dependency updates brief unless they're security-related or major version changes
- Include links to the PRs that implemented the feature/fix/change where applicable, matching the existing style/examples in @CHANGELOG.md
