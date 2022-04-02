#!/bin/bash

set -euo pipefail

cd "$(dirname "$0")/.." || exit

run()
{
    # run any pending db migrations
    ./bin/flipt migrate --config ./config/local.yml &> /dev/null

    ./bin/flipt --config ./config/local.yml &> /dev/null &

    sleep 5

    flipt_host="127.0.0.1:8080"

    echo -e "\e[32m                \e[0m"
    echo -e "\e[32m===========================================\e[0m"
    echo -e "\e[32mStart UI testing $flipt_host\e[0m"
    echo -e "\e[32m===========================================\e[0m"

    ./test/helpers/wait-for-it/wait-for-it.sh "$flipt_host" -t 30

    cd "ui" && yarn install --frozen-lockfile
    yarn test
}

run
