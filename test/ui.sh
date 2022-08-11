#!/bin/bash

set -euo pipefail

if [[ -z "$CI" ]]; then
    echo "This script is meant to run in CI only" 1>&2
    exit 1
fi

cd "$(dirname "$0")/.." || exit

FLIPT_PID="/tmp/flipt.ui.pid"

finish() {
  [[ -f "$FLIPT_PID" ]] && kill -9 `cat $FLIPT_PID`
}

run()
{
    # run any pending db migrations
    ./bin/flipt migrate --config ./test/config/test.yml &> /dev/null

    ./bin/flipt --config ./test/config/test.yml &> /dev/null &
    echo $! > "$FLIPT_PID"

    port="${FLIPT_SERVER_HTTP_PORT:-8080}"

    sleep 5

    flipt_host="127.0.0.1:${port}"

    echo -e "\e[32m                \e[0m"
    echo -e "\e[32m===========================================\e[0m"
    echo -e "\e[32mStart UI testing $flipt_host\e[0m"
    echo -e "\e[32m===========================================\e[0m"

    ./test/helpers/wait-for-it/wait-for-it.sh "$flipt_host" -t 30

    cd "ui" && npm ci && npx playwright install chromium
    npm test && npx playwright test
}

run
