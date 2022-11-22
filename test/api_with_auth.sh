#!/bin/bash

set -eo pipefail

if [[ -z "$CI" ]]; then
    echo "This script is meant to run in CI only" 1>&2
    exit 1
fi

_api_test_hook() {
  FLIPT_TOKEN=$(cat out.log | jq -r '. | select(.M=="access token created") | .client_token')
  export FLIPT_TOKEN
}

run() {
  export -f _api_test_hook

  ./test/api.sh "test-with-auth.yml"
}

run
