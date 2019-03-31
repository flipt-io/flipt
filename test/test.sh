#!/bin/bash

source /vendor/shakedown/shakedown.sh

uuid_str()
{
    cat /dev/urandom | tr -dc '[:lower:]' | fold -w 16 | head -n 1
}

step_1_test_health()
{
    shakedown GET "$flipt_host/health"
        status 200
}

step_2_test_flags_and_variants()
{
    # create flag
    flag_key=$(uuid_str)
    flag_name_1=$(uuid_str)

    shakedown POST "$flipt_url/flags" -H 'Content-Type:application/json' -d "{\"key\":\"$flag_key\",\"name\":\"$flag_name_1\",\"description\":\"description\",\"enabled\":true}"
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$flag_name_1\""
        matches '"enabled":true'

    # get flag
    shakedown GET "$flipt_url/flags/$flag_key" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$flag_name_1\""

    # update flag
    flag_name_2=$(uuid_str)

    shakedown PUT "$flipt_url/flags/$flag_key" -H 'Content-Type:application/json' -d "{\"key\":\"$flag_key\",\"name\":\"$flag_name_2\",\"description\":\"description\",\"enabled\":true}"
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$flag_name_2\""

    # list flags
    shakedown GET "$flipt_url/flags" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$flag_name_2\""

    # create variants
    variant_key_1=$(uuid_str)
    variant_key_2=$(uuid_str)

    shakedown POST "$flipt_url/flags/$flag_key/variants" -H 'Content-Type:application/json' -d "{\"key\":\"$variant_key_1\"}"
        status 200
        matches "\"key\":\"$variant_key_1\""

    shakedown POST "$flipt_url/flags/$flag_key/variants" -H 'Content-Type:application/json' -d "{\"key\":\"$variant_key_2\"}"
        status 200
        matches "\"key\":\"$variant_key_2\""

    variant_id=$(curl -sS "$flipt_url/flags/$flag_key" | jq '.variants | .[0].id')
    variant_id=$(eval echo "$variant_id")

    # update variant
    variant_name_1=$(uuid_str)

    shakedown PUT "$flipt_url/flags/$flag_key/variants/$variant_id" -H 'Content-Type:application/json' -d "{\"key\":\"$variant_key_1\",\"name\":\"$variant_name_1\"}"
        status 200
        matches "\"key\":\"$variant_key_1\""
        matches "\"name\":\"$variant_name_1\""

    # get flag w/ variants
    shakedown GET "$flipt_url/flags/$flag_key" -H 'Content-Type:application/json'
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

    shakedown POST "$flipt_url/segments" -H 'Content-Type:application/json' -d "{\"key\":\"$segment_key\",\"name\":\"$segment_name_1\",\"description\":\"description\"}"
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$segment_name_1\""

    # get segment
    shakedown GET "$flipt_url/segments/$segment_key" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$segment_name_1\""

    # update segment
    segment_name_2=$(uuid_str)

    shakedown PUT "$flipt_url/segments/$segment_key" -H 'Content-Type:application/json' -d "{\"key\":\"$segment_key\",\"name\":\"$segment_name_2\",\"description\":\"description\"}"
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$segment_name_2\""

    # list segments
    shakedown GET "$flipt_url/segments" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$segment_name_2\""

    # create constraints
    shakedown POST "$flipt_url/segments/$segment_key/constraints" -H 'Content-Type:application/json' -d "{\"type\":\"STRING_COMPARISON_TYPE\",\"property\":\"foo\",\"operator\":\"eq\",\"value\":\"bar\"}"
        status 200
        matches "\"property\":\"foo\""
        matches "\"operator\":\"eq\""
        matches "\"value\":\"bar\""

    shakedown POST "$flipt_url/segments/$segment_key/constraints" -H 'Content-Type:application/json' -d "{\"type\":\"STRING_COMPARISON_TYPE\",\"property\":\"fizz\",\"operator\":\"neq\",\"value\":\"buzz\"}"
        status 200
        matches "\"property\":\"fizz\""
        matches "\"operator\":\"neq\""
        matches "\"value\":\"buzz\""

    constraint_id=$(curl -sS "$flipt_url/segments/$segment_key" | jq '.constraints | .[0].id')
    constraint_id=$(eval echo "$constraint_id")

    # update constraint
    shakedown PUT "$flipt_url/segments/$segment_key/constraints/$constraint_id" -H 'Content-Type:application/json' -d "{\"type\":\"STRING_COMPARISON_TYPE\",\"property\":\"foo\",\"operator\":\"eq\",\"value\":\"baz\"}"
        status 200
        matches "\"property\":\"foo\""
        matches "\"operator\":\"eq\""
        matches "\"value\":\"baz\""

    # get segment w/ constraints
    shakedown GET "$flipt_url/segments/$segment_key" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$segment_key\""
        contains "baz"
        contains "buzz"
}

step_4_test_rules_and_distributions()
{
    # create rule
    shakedown POST "$flipt_url/flags/$flag_key/rules" -H 'Content-Type:application/json' -d "{\"segment_key\":\"$segment_key\",\"rank\":1}"
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"segmentKey\":\"$segment_key\""
        matches "\"rank\":1"

    # list rules
    shakedown GET "$flipt_url/flags/$flag_key/rules" -H 'Content-Type:application/json'
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"segmentKey\":\"$segment_key\""
        matches "\"rank\":1"

    rule_id=$(curl -sS "$flipt_url/flags/$flag_key/rules" | jq '.rules | .[0].id')
    rule_id=$(eval echo "$rule_id")

    # get rule
    shakedown GET "$flipt_url/flags/$flag_key/rules/$rule_id" -H 'Content-Type:application/json'
        status 200
        matches "\"id\":\"$rule_id\""
        matches "\"flagKey\":\"$flag_key\""
        matches "\"segmentKey\":\"$segment_key\""
        matches "\"rank\":1"

    # create distribution
    shakedown POST "$flipt_url/flags/$flag_key/rules/$rule_id/distributions" -H 'Content-Type:application/json' -d "{\"variant_id\":\"$variant_id\",\"rollout\":100}"
        status 200
        matches "\"ruleId\":\"$rule_id\""
        matches "\"variantId\":\"$variant_id\""
        matches "\"rollout\":100"
}

step_5_test_evaluation()
{
    # evaluate
    shakedown POST "$flipt_url/evaluate" -H 'Content-Type:application/json' -d "{\"flag_key\":\"$flag_key\",\"entity_id\":\"$(uuid_str)\",\"context\":{\"foo\":\"baz\"}}"
        status 200
        matches "\"flagKey\":\"$flag_key\""
        matches "\"segmentKey\":\"$segment_key\""
        matches "\"match\":true"
        matches "\"value\":\"$variant_key_1\""
}

step_6_test_delete()
{
    # delete rule
    shakedown DELETE "$flipt_url/flags/$flag_key/rules/$rule_id" -H 'Content-Type:application/json'
        status 200

    # delete flag
    shakedown DELETE "$flipt_url/flags/$flag_key" -H 'Content-Type:application/json'
        status 200

    # delete segment
    shakedown DELETE "$flipt_url/segments/$segment_key" -H 'Content-Type:application/json'
        status 200
}

start()
{
    flipt_host="$1:8080"

    echo -e "\e[32m                \e[0m"
    echo -e "\e[32m===========================================\e[0m"
    echo -e "\e[32mStart testing $flipt_host\e[0m"
    echo -e "\e[32m===========================================\e[0m"

    /vendor/wait-for-it/wait-for-it.sh "$flipt_host" -t 30

    flipt_url=$flipt_host/api/v1

    step_1_test_health
    step_2_test_flags_and_variants
    step_3_test_segments_and_constraints
    step_4_test_rules_and_distributions
    step_5_test_evaluation
    step_6_test_delete
}

start flipt
