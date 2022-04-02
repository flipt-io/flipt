#!/usr/bin/env bats

load 'helpers/bats-support/load'
load 'helpers/bats-assert/load'

@test "unknown command results in error" {
    run ./bin/flipt foo
    assert_failure
    assert_equal "${lines[0]}" "Error: unknown command \"foo\" for \"flipt\""
    assert_equal "${lines[1]}" "Run 'flipt --help' for usage."
}

@test "config file does not exists results in error" {
    run ./bin/flipt --config /foo/bar.yml
    assert_failure
    assert_output -p "loading configuration: open /foo/bar.yml: no such file or directory"
}

@test "config file not yaml results in error" {
    run ./bin/flipt --config /tmp
    assert_failure
    assert_output -p "loading configuration: Unsupported Config Type"
}

@test "help flag prints usage" {
    run ./bin/flipt --help
    assert_success
    assert_equal "${lines[0]}" "Flipt is a modern feature flag solution"
    assert_equal "${lines[1]}" "Usage:"
    assert_equal "${lines[2]}" "  flipt [flags]"
    assert_equal "${lines[3]}" "  flipt [command]"
    assert_equal "${lines[4]}" "Available Commands:"
    assert_equal "${lines[5]}" "  export      Export flags/segments/rules to file/stdout"
    assert_equal "${lines[6]}" "  help        Help about any command"
    assert_equal "${lines[7]}" "  import      Import flags/segments/rules from file"
    assert_equal "${lines[8]}" "  migrate     Run pending database migrations"
    assert_equal "${lines[9]}" "Flags:" ]
    assert_equal "${lines[10]}" "      --config string   path to config file (default \"/etc/flipt/config/default.yml\")"
}

@test "version flag prints version info" {
    run ./bin/flipt --version
    assert_success
    assert_output -p "Commit:"
    assert_output -e "Build Date: [0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z"
    assert_output -e "Go Version: go[0-9]+\.[0-9]+\.[0-9]"
}

@test "import with empty database from STDIN" {
    run bash -c "rm ./test/flipt.db; cat ./test/flipt.yml | ./bin/flipt --config ./test/config/test.yml import --stdin"
    assert_success
}

@test "import existing data from file not unique results in error" {
    run ./bin/flipt --config ./test/config/test.yml import ./test/flipt.yml
    assert_output -p "is not unique"
    assert_failure
}

@test "import with invalid data from STDIN results in error" {
    run bash -c "echo FOOBAR | ./bin/flipt --config ./test/config/test.yml import --stdin"
    assert_failure
}

@test "import with file that doesnt exist results in error" {
    run ./bin/flipt --config ./test/config/test.yml import foo
    assert_output -p "opening import file: open foo: no such file or directory"
    assert_failure
}

@test "import with existing data not unique and --drop flag is used" {
    run ./bin/flipt --config ./test/config/test.yml import ./test/flipt.yml --drop
    assert_success
}

@test "export outputs to STDOUT" {
    run ./bin/flipt --config ./test/config/test.yml export
    assert_output -p "flags:"
    assert_output -p "- key: zUFtS7D0UyMeueYu"
    assert_output -p "  variants:"
    assert_output -p "  rules:"
    assert_output -p "segments:"
    assert_output -p "- key: 08UoVJ96LhZblPEx"
    assert_output -p "  constraints:"
    assert_success
}

@test "export outputs to file" {
    run ./bin/flipt --config ./test/config/test.yml export -o /tmp/flipt.yml
    run test -f "/tmp/flipt.yml"
    assert_success
}

@test "migrate with empty db" {
    run bash -c "rm ./test/flipt.db; ./bin/flipt --config ./test/config/test.yml migrate"
    assert_success
}
