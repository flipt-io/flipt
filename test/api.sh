#!/bin/bash

set -o pipefail

if [[ -z "$CI" ]]; then
    echo "This script is meant to run in CI only" 1>&2
    exit 1
fi

cd "$(dirname "$0")/.." || exit

export SHAKEDOWN_URL="http://0.0.0.0:8080"

source ./test/helpers/shakedown/shakedown.sh

CONFIG_FILE=${1:-"test.yml"}
FLIPT_PID="/tmp/flipt.api.pid"

finish() {
  _finish # shakedown trap that sets exit code correctly
  [[ -f "$FLIPT_PID" ]] && kill -9 `cat $FLIPT_PID`
}

trap finish EXIT

uuid_str()
{
    uuidgen
}

shakedownJSON() {
  shakedown "$@" -H "Accept: application/json"
}

authedShakedown() {
  shakedownJSON "$@" -H "Authorization: Bearer ${FLIPT_TOKEN:-""}"
}

_curl() {
  curl "$@" -H "Authorization: Bearer ${FLIPT_TOKEN:-""}"
}

# header_matches takes two arguments "key" and "substring".
# it passes if the value for the associated key matches the substring.
# shakedown includes header_contains which is exact match
header_matches() {
	HEADER_NAME=${1}
  HEADER="$(_get_header $HEADER_NAME)"
  echo "${HEADER}" | grep -Eq "${2}" && _pass "${HEADER_NAME}: ${2}" || _fail "${HEADER_NAME}: ${2} (actual: ${HEADER})"
}

# does_not_contain is equivalent to !contains
# shakedown doesn't appear to support the negative
# cases out of the box.
does_not_contain() {
  MSG="does not contain \"${1}\""
  print_body | grep -Fq "${1}" && _fail "${MSG}" || _pass "${MSG}"
}

step_1_test_health()
{
    shakedown GET "/health"
        status 200
}

step_2_test_flags_and_variants()
{
    # create flag
    flag_key=$(uuid_str)
    flag_name_1=$(uuid_str)

    authedShakedown POST "/api/v1/flags" -H 'Content-Type:application/json' -d "{\"key\":\"$flag_key\",\"name\":\"$flag_name_1\",\"description\":\"description\",\"enabled\":true}"
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$flag_name_1\""
        matches '"enabled":true'

    # get flag
    authedShakedown GET "/api/v1/flags/$flag_key" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$flag_name_1\""

    # update flag
    flag_name_2=$(uuid_str)

    authedShakedown PUT "/api/v1/flags/$flag_key" -H 'Content-Type:application/json' -d "{\"key\":\"$flag_key\",\"name\":\"$flag_name_2\",\"description\":\"description\",\"enabled\":true}"
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$flag_name_2\""

    # list flags
    authedShakedown GET "/api/v1/flags" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$flag_name_2\""

    # create variants
    variant_key_1=$(uuid_str)
    variant_key_2=$(uuid_str)

    authedShakedown POST "/api/v1/flags/$flag_key/variants" -H 'Content-Type:application/json' -d "{\"key\":\"$variant_key_1\"}"
        status 200
        matches "\"key\":\"$variant_key_1\""

    authedShakedown POST "/api/v1/flags/$flag_key/variants" -H 'Content-Type:application/json' -d "{\"key\":\"$variant_key_2\"}"
        status 200
        matches "\"key\":\"$variant_key_2\""

    variant_id=$(_curl -sS "$SHAKEDOWN_URL/api/v1/flags/$flag_key" | jq '.variants | .[0].id')
    variant_id=$(eval echo "$variant_id")

    # update variant
    variant_name_1=$(uuid_str)

    authedShakedown PUT "/api/v1/flags/$flag_key/variants/$variant_id" -H 'Content-Type:application/json' -d "{\"key\":\"$variant_key_1\",\"name\":\"$variant_name_1\"}"
        status 200
        matches "\"key\":\"$variant_key_1\""
        matches "\"name\":\"$variant_name_1\""

    # get flag w/ variants
    authedShakedown GET "/api/v1/flags/$flag_key" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$flag_key\""
        contains "$variant_key_1"
        contains "$variant_key_2"
}

step_3_test_segments_and_constraints()
{
    # create segment
    segment_key=$(uuid_str)
    segment_name_1=$(uuid_str)

    authedShakedown POST "/api/v1/segments" -H 'Content-Type:application/json' -d "{\"key\":\"$segment_key\",\"name\":\"$segment_name_1\",\"description\":\"description\"}"
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$segment_name_1\""
        matches "\"matchType\":\"ALL_MATCH_TYPE\""

    # get segment
    authedShakedown GET "/api/v1/segments/$segment_key" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$segment_name_1\""
        matches "\"matchType\":\"ALL_MATCH_TYPE\""

    # update segment
    segment_name_2=$(uuid_str)

    authedShakedown PUT "/api/v1/segments/$segment_key" -H 'Content-Type:application/json' -d "{\"key\":\"$segment_key\",\"name\":\"$segment_name_2\",\"matchType\":\"ANY_MATCH_TYPE\",\"description\":\"description\"}"
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$segment_name_2\""
        matches "\"matchType\":\"ANY_MATCH_TYPE\""

    # list segments
    authedShakedown GET "/api/v1/segments" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$segment_name_2\""

    # create constraints
    authedShakedown POST "/api/v1/segments/$segment_key/constraints" -H 'Content-Type:application/json' -d "{\"type\":\"STRING_COMPARISON_TYPE\",\"property\":\"foo\",\"operator\":\"eq\",\"value\":\"bar\"}"
        status 200
        matches "\"property\":\"foo\""
        matches "\"operator\":\"eq\""
        matches "\"value\":\"bar\""

    authedShakedown POST "/api/v1/segments/$segment_key/constraints" -H 'Content-Type:application/json' -d "{\"type\":\"STRING_COMPARISON_TYPE\",\"property\":\"fizz\",\"operator\":\"neq\",\"value\":\"buzz\"}"
        status 200
        matches "\"property\":\"fizz\""
        matches "\"operator\":\"neq\""
        matches "\"value\":\"buzz\""

    constraint_id=$(_curl -sS "$SHAKEDOWN_URL/api/v1/segments/$segment_key" | jq '.constraints | .[0].id')
    constraint_id=$(eval echo "$constraint_id")

    # update constraint
    authedShakedown PUT "/api/v1/segments/$segment_key/constraints/$constraint_id" -H 'Content-Type:application/json' -d "{\"type\":\"STRING_COMPARISON_TYPE\",\"property\":\"foo\",\"operator\":\"eq\",\"value\":\"baz\"}"
        status 200
        matches "\"property\":\"foo\""
        matches "\"operator\":\"eq\""
        matches "\"value\":\"baz\""

    # get segment w/ constraints
    authedShakedown GET "/api/v1/segments/$segment_key" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$segment_key\""
        contains "baz"
        contains "buzz"
}

step_4_test_rules_and_distributions()
{
    # create rule
    authedShakedown POST "/api/v1/flags/$flag_key/rules" -H 'Content-Type:application/json' -d "{\"segment_key\":\"$segment_key\",\"rank\":1}"
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"segmentKey\":\"$segment_key\""
        matches "\"rank\":1"

    # list rules
    authedShakedown GET "/api/v1/flags/$flag_key/rules" -H 'Content-Type:application/json'
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"segmentKey\":\"$segment_key\""
        matches "\"rank\":1"

    rule_id_1=$(_curl -sS "$SHAKEDOWN_URL/api/v1/flags/$flag_key/rules" | jq '.rules | .[0].id')
    rule_id_1=$(eval echo "$rule_id_1")

    # get rule
    authedShakedown GET "/api/v1/flags/$flag_key/rules/$rule_id_1" -H 'Content-Type:application/json'
        status 200
        matches "\"id\":\"$rule_id_1\""
        matches "\"flagKey\":\"$flag_key\""
        matches "\"segmentKey\":\"$segment_key\""
        matches "\"rank\":1"

    # create another rule
    authedShakedown POST "/api/v1/flags/$flag_key/rules" -H 'Content-Type:application/json' -d "{\"segment_key\":\"$segment_key\",\"rank\":2}"
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"segmentKey\":\"$segment_key\""
        matches "\"rank\":2"

    rule_id_2=$(_curl -sS "$SHAKEDOWN_URL/api/v1/flags/$flag_key/rules" | jq '.rules | .[1].id')
    rule_id_2=$(eval echo "$rule_id_2")

    # reorder rules
    authedShakedown PUT "/api/v1/flags/$flag_key/rules/order" -H 'Content-Type:application/json' -d "{\"ruleIds\":[\"$rule_id_2\",\"$rule_id_1\"]}"
        status 200

    # create distribution
    authedShakedown POST "/api/v1/flags/$flag_key/rules/$rule_id_2/distributions" -H 'Content-Type:application/json' -d "{\"variant_id\":\"$variant_id\",\"rollout\":100}"
        status 200
        matches "\"ruleId\":\"$rule_id_2\""
        matches "\"variantId\":\"$variant_id\""
        matches "\"rollout\":100"
}

step_5_test_evaluation()
{
    # evaluate
    authedShakedown POST "/api/v1/evaluate" -H 'Content-Type:application/json' -d "{\"flag_key\":\"$flag_key\",\"entity_id\":\"$(uuid_str)\",\"context\":{\"foo\":\"baz\",\"fizz\":\"bozz\"}}"
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"segmentKey\":\"$segment_key\""
        matches "\"match\":true"
        matches "\"value\":\"$variant_key_1\""
        matches "\"reason\":\"MATCH_EVALUATION_REASON\""

    # evaluate no match
    authedShakedown POST "/api/v1/evaluate" -H 'Content-Type:application/json' -d "{\"flag_key\":\"$flag_key\",\"entity_id\":\"$(uuid_str)\",\"context\":{\"fizz\":\"buzz\"}}"
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"match\":false"

    # evaluate handles null value
    # re: #664
    authedShakedown POST "/api/v1/evaluate" -H 'Content-Type:application/json' -d "{\"flag_key\":\"$flag_key\",\"entity_id\":\"$(uuid_str)\",\"context\":{\"cohort\":null}}"
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"match\":false"

    # evaluate returns 400 plus user friendly error message
    authedShakedown POST "/api/v1/evaluate" -H 'Content-Type:application/json' -d "{\"flag_key\":\"$flag_key\",\"entity_id\":\"$(uuid_str)\",\"context\":\"hello\"}"
        status 400
        matches "\"code\":3"
        contains "\"message\":\"invalid values for key(s) in json body\""
}

step_6_test_batch_evaluation()
{
    # evaluate
    authedShakedown POST "/api/v1/batch-evaluate" -H 'Content-Type:application/json' -d "{\"requests\": [{\"flag_key\":\"$flag_key\",\"entity_id\":\"$(uuid_str)\",\"context\":{\"foo\":\"baz\",\"fizz\":\"bozz\"}}]}"
        status 200
        contains "\"flagKey\":\"$flag_key\""
        contains "\"segmentKey\":\"$segment_key\""
        contains "\"match\":true"
        contains "\"value\":\"$variant_key_1\""

    # evaluate no match
    authedShakedown POST "/api/v1/batch-evaluate" -H 'Content-Type:application/json' -d "{\"requests\": [{\"flag_key\":\"$flag_key\",\"entity_id\":\"$(uuid_str)\",\"context\":{\"fizz\":\"buzz\"}}]}"
        status 200
        contains "\"flagKey\":\"$flag_key\""
        contains "\"match\":false"
}

step_7_test_delete()
{
    # delete rules and distributions
    authedShakedown DELETE "/api/v1/flags/$flag_key/rules/$rule_id_1" -H 'Content-Type:application/json'
        status 200

    authedShakedown DELETE "/api/v1/flags/$flag_key/rules/$rule_id_2" -H 'Content-Type:application/json'
        status 200

    # delete flag and variants
    authedShakedown DELETE "/api/v1/flags/$flag_key" -H 'Content-Type:application/json'
        status 200

    # delete segment and constraints
    authedShakedown DELETE "/api/v1/segments/$segment_key" -H 'Content-Type:application/json'
        status 200
}

step_8_test_meta()
{
    authedShakedown GET "/meta/info"
        status 200
        contains "\"buildDate\""
        contains "\"goVersion\""

    authedShakedown GET "/meta/config"
        status 200
        contains "\"log\""
        contains "\"ui\""
        contains "\"cache\""
        contains "\"server\""
        contains "\"db\""

    # in the authentication enabled case we check that
    # the returned config does not contain the CSRF key
    if [ -n "${TEST_FLIPT_API_AUTH_REQUIRED:-}" ]; then
      key=$(yq eval '.authentication.session.csrf.key' ./test/config/test-with-auth.yml | tr -d '\r\n')
      does_not_contain "${key}"

      # ensure CSRF cookie is present
      header_matches "Set-Cookie" "_gorilla_csrf"

      # ensure unauthenticated request returns 401
      shakedownJSON GET "/meta/info"
        status 401

      shakedownJSON GET "/meta/config"
        status 401
    fi
}

step_9_test_metrics()
{
    shakedown GET "/metrics"
        status 200
}

step_10_test_auths()
{
    # create a new token for the purpose of this test
    tokenPayload=$(_curl -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer ${FLIPT_TOKEN:-''}" -sS "$SHAKEDOWN_URL/auth/v1/method/token" -d '{"name":"token","description":"an access token"}')
    tokenID=$(echo "${tokenPayload}" | jq -r '.authentication.id')

    # replace FLIPT_TOKEN with created token
    FLIPT_TOKEN=$(echo "${tokenPayload}" | jq -r '.clientToken')
    export FLIPT_TOKEN

    # /auth/v1/method is always public
    shakedownJSON GET '/auth/v1/method'
        status 200

    # token should succeed when used via authorization header to list flags
    # (both when auth is required and not)
    authedShakedown GET '/api/v1/flags'
        status 200

    # listing tokens includes the created token
    authedShakedown GET "/auth/v1/tokens"
        status 200
        matches "\"id\":\"${tokenID}\""

    if [ -n "${TEST_FLIPT_API_AUTH_REQUIRED:-}" ]; then
        # getting self using token returns expected ID
        authedShakedown GET '/auth/v1/self'
        status 200
        matches "\"id\":\"${tokenID}\""

        # cookie based auth is configured and should also work
        shakedownJSON GET '/auth/v1/self' -H "Cookie: flipt_client_token=${FLIPT_TOKEN}"
        status 200
        matches "\"id\":\"${tokenID}\""

        # expiring self token should return 200
        authedShakedown PUT "/auth/v1/self/expire"
        header_matches "Set-Cookie" "flipt_client_token=.*Max-Age=0"
        status 200

        # getting self using expired token should return 401
        authedShakedown GET '/auth/v1/self'
        status 401

        # all attempts to use an expired cookie cause the cookie to be cleared
        # via Set-Cookie
        shakedownJSON GET '/auth/v1/self' -H "Cookie: flipt_client_token=${FLIPT_TOKEN}"
        header_matches "Set-Cookie" "flipt_client_token=.*Max-Age=0"
        status 401
    else
        # there is no self when authentication is disabled
        authedShakedown GET '/auth/v1/self'
        status 401
    fi
}

run()
{
    # run any pending db migrations
    ./bin/flipt migrate ---config "./test/config/$CONFIG_FILE" &> /dev/null

    ./bin/flipt --config "./test/config/$CONFIG_FILE" &> out.log &
    echo $! > "$FLIPT_PID"

    sleep 5

    echo -e "\e[32m                \e[0m"
    echo -e "\e[32m===========================================\e[0m"
    echo -e "\e[32mStart testing $SHAKEDOWN_URL\e[0m"
    echo -e "\e[32m===========================================\e[0m"

    ./test/helpers/wait-for-it/wait-for-it.sh "0.0.0.0:8080" -t 30

    # allows api with auth to extract FLIPT_TOKEN from out log
    [[ $(type -t _api_test_hook) == function ]] && _api_test_hook

    step_1_test_health
    step_2_test_flags_and_variants
    step_3_test_segments_and_constraints
    step_4_test_rules_and_distributions
    step_5_test_evaluation
    step_6_test_batch_evaluation
    step_7_test_delete
    step_8_test_meta
    step_9_test_metrics
    step_10_test_auths
}

run
