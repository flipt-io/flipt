#!/bin/bash

set -o pipefail

if [[ -z "$CI" ]]; then
    echo "This script is meant to run in CI only" 1>&2
    exit 1
fi

cd "$(dirname "$0")/.." || exit

export SHAKEDOWN_URL="http://127.0.0.1:8080"

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
    LC_ALL=C; cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 16 | head -n 1
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

    shakedown POST "/api/v1/flags" -H 'Content-Type:application/json' -d "{\"key\":\"$flag_key\",\"name\":\"$flag_name_1\",\"description\":\"description\",\"enabled\":true}"
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$flag_name_1\""
        matches '"enabled":true'

    # get flag
    shakedown GET "/api/v1/flags/$flag_key" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$flag_name_1\""

    # update flag
    flag_name_2=$(uuid_str)

    shakedown PUT "/api/v1/flags/$flag_key" -H 'Content-Type:application/json' -d "{\"key\":\"$flag_key\",\"name\":\"$flag_name_2\",\"description\":\"description\",\"enabled\":true}"
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$flag_name_2\""

    # list flags
    shakedown GET "/api/v1/flags" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$flag_name_2\""

    # create variants
    variant_key_1=$(uuid_str)
    variant_key_2=$(uuid_str)

    shakedown POST "/api/v1/flags/$flag_key/variants" -H 'Content-Type:application/json' -d "{\"key\":\"$variant_key_1\"}"
        status 200
        matches "\"key\":\"$variant_key_1\""

    shakedown POST "/api/v1/flags/$flag_key/variants" -H 'Content-Type:application/json' -d "{\"key\":\"$variant_key_2\"}"
        status 200
        matches "\"key\":\"$variant_key_2\""

    variant_id=$(curl -sS "$SHAKEDOWN_URL/api/v1/flags/$flag_key" | jq '.variants | .[0].id')
    variant_id=$(eval echo "$variant_id")

    # update variant
    variant_name_1=$(uuid_str)

    shakedown PUT "/api/v1/flags/$flag_key/variants/$variant_id" -H 'Content-Type:application/json' -d "{\"key\":\"$variant_key_1\",\"name\":\"$variant_name_1\"}"
        status 200
        matches "\"key\":\"$variant_key_1\""
        matches "\"name\":\"$variant_name_1\""

    # get flag w/ variants
    shakedown GET "/api/v1/flags/$flag_key" -H 'Content-Type:application/json'
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

    shakedown POST "/api/v1/segments" -H 'Content-Type:application/json' -d "{\"key\":\"$segment_key\",\"name\":\"$segment_name_1\",\"description\":\"description\"}"
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$segment_name_1\""
        matches "\"matchType\":\"ALL_MATCH_TYPE\""

    # get segment
    shakedown GET "/api/v1/segments/$segment_key" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$segment_name_1\""
        matches "\"matchType\":\"ALL_MATCH_TYPE\""

    # update segment
    segment_name_2=$(uuid_str)

    shakedown PUT "/api/v1/segments/$segment_key" -H 'Content-Type:application/json' -d "{\"key\":\"$segment_key\",\"name\":\"$segment_name_2\",\"matchType\":\"ANY_MATCH_TYPE\",\"description\":\"description\"}"
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$segment_name_2\""
        matches "\"matchType\":\"ANY_MATCH_TYPE\""

    # list segments
    shakedown GET "/api/v1/segments" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$segment_name_2\""

    # create constraints
    shakedown POST "/api/v1/segments/$segment_key/constraints" -H 'Content-Type:application/json' -d "{\"type\":\"STRING_COMPARISON_TYPE\",\"property\":\"foo\",\"operator\":\"eq\",\"value\":\"bar\"}"
        status 200
        matches "\"property\":\"foo\""
        matches "\"operator\":\"eq\""
        matches "\"value\":\"bar\""

    shakedown POST "/api/v1/segments/$segment_key/constraints" -H 'Content-Type:application/json' -d "{\"type\":\"STRING_COMPARISON_TYPE\",\"property\":\"fizz\",\"operator\":\"neq\",\"value\":\"buzz\"}"
        status 200
        matches "\"property\":\"fizz\""
        matches "\"operator\":\"neq\""
        matches "\"value\":\"buzz\""

    constraint_id=$(curl -sS "$SHAKEDOWN_URL/api/v1/segments/$segment_key" | jq '.constraints | .[0].id')
    constraint_id=$(eval echo "$constraint_id")

    # update constraint
    shakedown PUT "/api/v1/segments/$segment_key/constraints/$constraint_id" -H 'Content-Type:application/json' -d "{\"type\":\"STRING_COMPARISON_TYPE\",\"property\":\"foo\",\"operator\":\"eq\",\"value\":\"baz\"}"
        status 200
        matches "\"property\":\"foo\""
        matches "\"operator\":\"eq\""
        matches "\"value\":\"baz\""

    # get segment w/ constraints
    shakedown GET "/api/v1/segments/$segment_key" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$segment_key\""
        contains "baz"
        contains "buzz"
}

step_4_test_rules_and_distributions()
{
    # create rule
    shakedown POST "/api/v1/flags/$flag_key/rules" -H 'Content-Type:application/json' -d "{\"segment_key\":\"$segment_key\",\"rank\":1}"
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"segmentKey\":\"$segment_key\""
        matches "\"rank\":1"

    # list rules
    shakedown GET "/api/v1/flags/$flag_key/rules" -H 'Content-Type:application/json'
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"segmentKey\":\"$segment_key\""
        matches "\"rank\":1"

    rule_id_1=$(curl -sS "$SHAKEDOWN_URL/api/v1/flags/$flag_key/rules" | jq '.rules | .[0].id')
    rule_id_1=$(eval echo "$rule_id_1")

    # get rule
    shakedown GET "/api/v1/flags/$flag_key/rules/$rule_id_1" -H 'Content-Type:application/json'
        status 200
        matches "\"id\":\"$rule_id_1\""
        matches "\"flagKey\":\"$flag_key\""
        matches "\"segmentKey\":\"$segment_key\""
        matches "\"rank\":1"

    # create another rule
    shakedown POST "/api/v1/flags/$flag_key/rules" -H 'Content-Type:application/json' -d "{\"segment_key\":\"$segment_key\",\"rank\":2}"
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"segmentKey\":\"$segment_key\""
        matches "\"rank\":2"

    rule_id_2=$(curl -sS "$SHAKEDOWN_URL/api/v1/flags/$flag_key/rules" | jq '.rules | .[1].id')
    rule_id_2=$(eval echo "$rule_id_2")

    # reorder rules
    shakedown PUT "/api/v1/flags/$flag_key/rules/order" -H 'Content-Type:application/json' -d "{\"ruleIds\":[\"$rule_id_2\",\"$rule_id_1\"]}"
        status 200

    # create distribution
    shakedown POST "/api/v1/flags/$flag_key/rules/$rule_id_2/distributions" -H 'Content-Type:application/json' -d "{\"variant_id\":\"$variant_id\",\"rollout\":100}"
        status 200
        matches "\"ruleId\":\"$rule_id_2\""
        matches "\"variantId\":\"$variant_id\""
        matches "\"rollout\":100"
}

step_5_test_evaluation()
{
    # evaluate
    shakedown POST "/api/v1/evaluate" -H 'Content-Type:application/json' -d "{\"flag_key\":\"$flag_key\",\"entity_id\":\"$(uuid_str)\",\"context\":{\"foo\":\"baz\",\"fizz\":\"bozz\"}}"
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"segmentKey\":\"$segment_key\""
        matches "\"match\":true"
        matches "\"value\":\"$variant_key_1\""

    # evaluate no match
    shakedown POST "/api/v1/evaluate" -H 'Content-Type:application/json' -d "{\"flag_key\":\"$flag_key\",\"entity_id\":\"$(uuid_str)\",\"context\":{\"fizz\":\"buzz\"}}"
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"match\":false"

    # evaluate handles null value
    # re: #664
    shakedown POST "/api/v1/evaluate" -H 'Content-Type:application/json' -d "{\"flag_key\":\"$flag_key\",\"entity_id\":\"$(uuid_str)\",\"context\":{\"cohort\":null}}"
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"match\":false"
}

step_6_test_batch_evaluation()
{
    # evaluate
    shakedown POST "/api/v1/batch-evaluate" -H 'Content-Type:application/json' -d "{\"requests\": [{\"flag_key\":\"$flag_key\",\"entity_id\":\"$(uuid_str)\",\"context\":{\"foo\":\"baz\",\"fizz\":\"bozz\"}}]}"
        status 200
        contains "\"flagKey\":\"$flag_key\""
        contains "\"segmentKey\":\"$segment_key\""
        contains "\"match\":true"
        contains "\"value\":\"$variant_key_1\""

    # evaluate no match
    shakedown POST "/api/v1/batch-evaluate" -H 'Content-Type:application/json' -d "{\"requests\": [{\"flag_key\":\"$flag_key\",\"entity_id\":\"$(uuid_str)\",\"context\":{\"fizz\":\"buzz\"}}]}"
        status 200
        contains "\"flagKey\":\"$flag_key\""
        contains "\"match\":false"
}

step_7_test_delete()
{
    # delete rules and distributions
    shakedown DELETE "/api/v1/flags/$flag_key/rules/$rule_id_1" -H 'Content-Type:application/json'
        status 200

    shakedown DELETE "/api/v1/flags/$flag_key/rules/$rule_id_2" -H 'Content-Type:application/json'
        status 200

    # delete flag and variants
    shakedown DELETE "/api/v1/flags/$flag_key" -H 'Content-Type:application/json'
        status 200

    # delete segment and constraints
    shakedown DELETE "/api/v1/segments/$segment_key" -H 'Content-Type:application/json'
        status 200
}

step_8_test_meta()
{
    shakedown GET "/meta/info"
        status 200
        contains "\"buildDate\""
        contains "\"goVersion\""

    shakedown GET "/meta/config"
        status 200
        contains "\"log\""
        contains "\"ui\""
        contains "\"cache\""
        contains "\"server\""
        contains "\"database\""
}

step_9_test_metrics()
{
    shakedown GET "/metrics"
        status 200
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

    ./test/helpers/wait-for-it/wait-for-it.sh "127.0.0.1:8080" -t 30

    step_1_test_health
    step_2_test_flags_and_variants
    step_3_test_segments_and_constraints
    step_4_test_rules_and_distributions
    step_5_test_evaluation
    step_6_test_batch_evaluation
    step_7_test_delete
    step_8_test_meta
    step_9_test_metrics
}

run
