#!/bin/bash
set -euo pipefail

# Decides which parts of CI a pull request needs, from its changed files
# (one path per line on stdin). Prints two lines for $GITHUB_OUTPUT:
#
#   go=true|false   Go build, lint, unit and integration jobs
#   ui=true|false   UI lint and unit test jobs
#
# A job is skipped only when every changed file is known not to affect it.
# Anything this script does not recognise runs everything, and so does empty
# input, which means the file list could not be read.

go=false
ui=false
seen=false

while IFS= read -r file || [[ -n "$file" ]]; do
    [[ -z "$file" ]] && continue
    seen=true
    case "$file" in
        # Go code embeds these two, so they are part of the Go build too.
        ui/*.go | ui/index.dev.html)
            go=true
            ui=true
            ;;
        ui/*)
            ui=true
            ;;
        # Go sources, modules and their fixtures (including any .md inside them).
        cmd/* | internal/* | core/* | rpc/* | sdk/* | errors/* | config/* | build/* | \
            go.mod | go.sum | .golangci.yml | .mockery.yml | buf.yaml | buf.lock)
            go=true
            ;;
        # Docs: agent guidance, the licence, and Markdown in the repo root only.
        .agents/* | LICENSE)
            ;;
        *.md)
            if [[ "$file" == */* ]]; then
                go=true
                ui=true
            fi
            ;;
        # Workflows, tool versions, Dockerfiles and anything unrecognised.
        *)
            go=true
            ui=true
            ;;
    esac
done

if [[ "$seen" == false ]]; then
    go=true
    ui=true
fi

echo "go=$go"
echo "ui=$ui"
