#!/usr/bin/env bats

load test_helper

@test 'refute() <expression>: returns 0 if <expression> evaluates to FALSE' {
  run refute false
  [ "$status" -eq 0 ]
  [ "${#lines[@]}" -eq 0 ]
}

@test 'refute() <expression>: returns 1 and displays <expression> if it evaluates to TRUE' {
  run refute true
  [ "$status" -eq 1 ]
  [ "${#lines[@]}" -eq 3 ]
  [ "${lines[0]}" == '-- assertion succeeded, but it was expected to fail --' ]
  [ "${lines[1]}" == 'expression : true' ]
  [ "${lines[2]}" == '--' ]
}
