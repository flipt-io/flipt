#!/bin/sh

set -eu

git config user.name "Mark Phelps"
git config user.email "mark.aaron.phelps@gmail.com"
git remote add gh-token "https://${GITHUB_TOKEN}@github.com/markphelps/flipt.git"
git fetch gh-token && git fetch gh-token gh-pages:gh-pages
mkdocs gh-deploy --clean --remote-name gh-token
