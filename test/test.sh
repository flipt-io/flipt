#!/bin/bash

source /vendor/shakedown/shakedown.sh

uuid_str()
{
    cat /dev/urandom | tr -dc 'a-z' | fold -w 16 | head -n 1
}

step_1_test_health()
{
    flipt_url=$1:8080

    shakedown GET "$flipt_url/health"
        status 200
}

step_2_test_flags_and_variants()
{
    flipt_url=$1:8080/api/v1

    # create flag
    flag_key=$(uuid_str)
    name=$(uuid_str)

    shakedown POST "$flipt_url/flags" -H 'Content-Type:application/json' -d "{\"key\":\"$flag_key\",\"name\":\"$name\",\"description\":\"description\",\"enabled\":true}"

        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$name\""
        matches '"enabled":true'

    # get flag
    shakedown GET "$flipt_url/flags/$flag_key" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$name\""

    # update flag
    name_2=$(uuid_str)

    shakedown PUT "$flipt_url/flags/$flag_key" -H 'Content-Type:application/json' -d "{\"key\":\"$flag_key\",\"name\":\"$name_2\",\"description\":\"description\",\"enabled\":false}"
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$name_2\""

    # list flags
    shakedown GET "$flipt_url/flags" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$flag_key\""
        matches "\"name\":\"$name_2\""

    # create variants
    variant_key_1=$(uuid_str)
    variant_key_2=$(uuid_str)

    shakedown POST "$flipt_url/flags/$flag_key/variants" -H 'Content-Type:application/json' -d "{\"key\": \"$variant_key_1\"}"
        status 200
        matches "\"key\":\"$variant_key_1\""

    shakedown POST "$flipt_url/flags/$flag_key/variants" -H 'Content-Type:application/json' -d "{\"key\": \"$variant_key_2\"}"
        status 200
        matches "\"key\":\"$variant_key_2\""
}

step_4_test_segments_and_constraints()
{
    flipt_url=$1:8080/api/v1

    # create segment
    segment_key=$(uuid_str)
    name=$(uuid_str)

    shakedown POST "$flipt_url/segments" -H 'Content-Type:application/json' -d "{\"key\":\"$segment_key\",\"name\":\"$name\",\"description\":\"description\"}"
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$name\""

    # get segment
    shakedown GET "$flipt_url/segments/$segment_key" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$name\""

    # update segment
    name_2=$(uuid_str)

    shakedown PUT "$flipt_url/segments/$segment_key" -H 'Content-Type:application/json' -d "{\"key\":\"$segment_key\",\"name\":\"$name_2\",\"description\":\"description\"}"
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$name_2\""

    # list segments
    shakedown GET "$flipt_url/segments" -H 'Content-Type:application/json'
        status 200
        matches "\"key\":\"$segment_key\""
        matches "\"name\":\"$name_2\""
}

step_10_test_delete() {
    # delete flag
    shakedown DELETE "$flipt_url/flags/$flag_key" -H 'Content-Type:application/json'
        status 200

    # delete segment
    shakedown DELETE "$flipt_url/segments/$segment_key" -H 'Content-Type:application/json'
        status 200
}

start()
{
    flipt_host=$1
    echo -e "\e[32m                \e[0m"
    echo -e "\e[32m===========================================\e[0m"
    echo -e "\e[32mStart testing $1\e[0m"
    echo -e "\e[32m===========================================\e[0m"

    /vendor/wait-for-it/wait-for-it.sh "$flipt_host:8080" -t 30

    step_1_test_health "$flipt_host"
    step_2_test_flags_and_variants "$flipt_host"
    step_4_test_segments_and_constraints "$flipt_host"
    step_10_test_delete "$flipt_host"
}

start flipt
