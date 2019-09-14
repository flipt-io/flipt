#!/usr/bin/env bats

load test_helper

@test 'assert() <expression>: returns 0 if <expression> evaluates to TRUE' {
  run assert true
  [ "$status" -eq 0 ]
  [ "${#lines[@]}" -eq 0 ]
}

@test 'assert() <expression>: returns 1 and displays <expression> if it evaluates to FALSE' {
  run assert false
  [ "$status" -eq 1 ]
  [ "${#lines[@]}" -eq 3 ]
  [ "${lines[0]}" == '-- assertion failed --' ]
  [ "${lines[1]}" == 'expression : false' ]
  [ "${lines[2]}" == '--' ]
}
