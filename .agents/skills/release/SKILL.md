---
name: release
description: Run the Flipt v2 stable release process, from changelog preparation through the annotated tag. Use when creating a new Flipt v2 release.
---

# Release Flipt v2

Use this skill for a stable Flipt v2 release. Ask for the version when it is
not provided. The version must be semver without the `v` prefix, such as
`2.12.0`.

This process has confirmation checkpoints. Stop at each checkpoint and wait
for the user.

## 1. Check prerequisites

Before making changes, verify all of the following:

- The version is valid semver without the `v` prefix.
- The working tree is clean.
- The current branch is `v2`.
- Local `v2` is up to date with `origin/v2`.

If a check fails, stop and tell the user what to fix.

## 2. Create the release branch

Create and check out `release/v<version>` from `v2`:

```sh
git switch -c release/v<version>
```

Stop and ask the user to confirm before continuing.

## 3. Prepare the changelog

Find the newest version in `CHANGELOG.md`. Review commits between that release
tag and `HEAD`:

```sh
git log --oneline v<previous-version>..HEAD
gh pr list --state merged --base v2 --limit 100 --json number,title,url
```

Add a new entry at the top of `CHANGELOG.md`:

- Use the existing Keep a Changelog format.
- Link the version to the GitHub release tag.
- Use the release date.
- Group user-facing changes under `Added`, `Changed`, `Fixed`, and
  `Documentation` as needed.
- Add pull request links or numbers when available.
- Keep dependency updates brief unless they are security-related or major.
- Do not invent changes that are not supported by commits or pull requests.

Show the generated entry and ask the user to review it. Stop until the user
confirms it is ready.

## 4. Commit the release changes

Run a whitespace check, then stage only intended release files and create a
signed-off commit:

```sh
git diff --check
git add CHANGELOG.md
git commit -s -m "chore: release v<version>"
```

Show the commit details. Stop and ask the user to confirm before pushing.

## 5. Push and open the pull request

Push the branch and open a pull request targeting `v2`:

```sh
git push -u origin release/v<version>
gh pr create --base v2 --head release/v<version> \
  --title "chore: release v<version>" \
  --body "$(cat <<'EOF'
## Summary

Release Flipt v<version>.

Highlights:

- Summarize the main changelog changes here.
EOF
)" \
  --label v2
```

Use a real multiline body. Do not pass literal `\\n` sequences as the body.
Share the pull request URL with the user.

Tell the user to review and squash-merge the pull request. Stop and wait for
confirmation that it was merged.

## 6. Tag the merged commit

The release pull request is squash-merged. The local release branch commit is
not the merged commit. Always tag the squashed commit on `origin/v2`:

```sh
git fetch origin
git switch v2
git reset --hard origin/v2
git log -1 --oneline
```

Confirm that `HEAD` is the squashed release commit. Then create an annotated
tag:

```sh
git tag -a v<version> -m "Release v<version>"
git show v<version>
```

Show the tag details. Stop and ask the user to confirm before pushing the tag.

## 7. Verify and push the tag

Before pushing, verify that the tag is reachable from `origin/v2`:

```sh
git merge-base --is-ancestor v<version> origin/v2 \
  && echo "OK: tag is reachable from origin/v2" \
  || echo "FAIL: tag is NOT on origin/v2 — do not push"
```

Only if the check prints `OK`, push the tag:

```sh
git push origin v<version>
```

Tell the user that CI should build and publish the release artifacts.

If the reachability check fails, do not push. Delete the tag with
`git tag -d v<version>`, sync `v2`, and repeat step 6.

## Safety rules

- Do not skip confirmation checkpoints.
- Do not continue after opening the pull request until the user confirms it
  was merged.
- Never push a tag that is not an ancestor of `origin/v2`.
- Surface any conflict with repository policy before continuing.
