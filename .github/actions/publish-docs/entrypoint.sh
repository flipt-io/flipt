#!/bin/sh

set -eu

if [ -z "${PERSONAL_TOKEN}" ]; then
    echo "error: not found PERSONAL_TOKEN"
    exit 1
fi

if [ -z "${GITHUB_ACTOR}" ]; then
    echo "error: not found GITHUB_ACTOR"
    exit 1
fi

git config user.name "${GITHUB_ACTOR}"
git config user.email "${GITHUB_ACTOR}@users.noreply.github.com"
git remote add gh-token "https://${GITHUB_ACTOR}:${PERSONAL_TOKEN}@github.com/markphelps/flipt.git"
git fetch gh-token && git fetch gh-token gh-pages:gh-pages
mkdocs gh-deploy --clean --remote-name gh-token
