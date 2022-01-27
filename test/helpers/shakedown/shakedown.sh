#!/usr/bin/env bash
set -u

# default values can be overriden by setting env vars
BASE_URL=${SHAKEDOWN_URL:-""}
CREDENTIALS=${SHAKEDOWN_CREDENTIALS:-""}
CONNECT_TIMEOUT=${SHAKEDOWN_CONNECT_TIMEOUT:-"5"}
MAX_TIME=${SHAKEDOWN_MAX_TIME:-"30"}

_usage() {
  echo '
usage: $0 [options...]
Options:
  -u <base URL>         Base URL to test.
  -c <user:password>    Credentials for HTTP authentication.
'
  exit 1
}

while getopts 'u:c:' OPTION
do
  case $OPTION in
    u) BASE_URL="$OPTARG";;
    c) CREDENTIALS="$OPTARG";;
    *) _usage;;
  esac
done


echo "Starting shakedown of ${BASE_URL:-"[base URL not set]"}"

STATE=""
FAIL_COUNT=0
PASS_COUNT=0
WORKING_DIR=$(mktemp -d -t shakedown.XXXXXX)
RESPONSE_BODY="${WORKING_DIR}/body"
RESPONSE_HEADERS="${WORKING_DIR}/headers"

AUTH=""
if [ -n "${CREDENTIALS}" ]; then
  AUTH="--anyauth --user ${CREDENTIALS}"
fi

CURL="curl -sS ${AUTH} -D ${RESPONSE_HEADERS} --connect-timeout ${CONNECT_TIMEOUT} --max-time ${MAX_TIME}"

CRED=$(tput setaf 1 2> /dev/null)
CGREEN=$(tput setaf 2 2> /dev/null)
CDEFAULT=$(tput sgr0 2> /dev/null)

_pass() {
  echo " ${CGREEN}✔ ${1}${CDEFAULT}"
}

_fail() {
  STATE="fail"
  echo " ${CRED}✘ ${1}${CDEFAULT}"
}

_start_test() {
  _finish_test
  STATE="pass"
}

_finish_test() {
  if [ "$STATE" =  "pass" ]; then
    PASS_COUNT=$((PASS_COUNT + 1))
  elif [ "$STATE" =  "fail" ]; then
    FAIL_COUNT=$((FAIL_COUNT + 1))
  fi
}

_finish() {
  _finish_test
  rm -rf "${WORKING_DIR}"
  echo
  MSG="Shakedown complete. ${PASS_COUNT} passed, ${FAIL_COUNT} failed."
  [[ ${FAIL_COUNT} -eq 0 ]] && echo "${CGREEN}${MSG}${CDEFAULT}" || echo "${CRED}${MSG} You're busted.${CDEFAULT}"
  exit ${FAIL_COUNT}
}

trap _finish EXIT

# start test
# $1 METHOD
# $2 URL
# $3..$n Custom CURL options
shakedown() {
  _start_test
  METHOD="$1"
  URL="$2"
  if ! [[ $URL == http* ]]; then
    URL="${BASE_URL}${URL}"
  fi
  echo
  echo "${METHOD} ${URL}"
  METHOD_OPT="-X ${METHOD}"
  if [ "${METHOD}" = "HEAD" ]; then
    METHOD_OPT="-I"
  fi
  ${CURL} ${METHOD_OPT} "${@:3}" "${URL}" > ${RESPONSE_BODY}
}

# assertions

header() {
  grep -iFq "${1}" "${RESPONSE_HEADERS}" && _pass "header ${1}" || _fail "header ${1}"
}

no_header() {
  grep -iFq "${1}" "${RESPONSE_HEADERS}" && _fail "no_header ${1}" || _pass "no_header ${1}"
}

status() {
  STATUS_CODE=$(grep -Eo "^HTTP.+ [1-5][0-9][0-9] " ${RESPONSE_HEADERS} | grep -Eo '[1-5][0-9][0-9]' | tail -n1)
  [[ "${STATUS_CODE}" = "${1}" ]] && _pass "status ${1}" || _fail "status ${1} (actual: ${STATUS_CODE})"
}

contains() {
  MSG="contains \"${1}\""
  grep -Fq "${1}" "${RESPONSE_BODY}" && _pass "${MSG}" || _fail "${MSG}"
}

matches() {
  MSG="matches \"${1}\""
  grep -Eq "${1}" "${RESPONSE_BODY}" && _pass "${MSG}" || _fail "${MSG}"
}

content_type() {
  CT_HEADER="$(_get_header 'Content-Type')"
  echo "${CT_HEADER}" | grep -Fq "${1}" && _pass "Content-Type: ${1}" || _fail "Content-Type: ${1} (actual: ${CT_HEADER})"
}

header_contains() {
	HEADER_NAME=${1}
  HEADER="$(_get_header $HEADER_NAME)"
  echo "${HEADER}" | grep -Fq "${2}" && _pass "${HEADER_NAME}: ${2}" || _fail "${HEADER_NAME}: ${2} (actual: ${HEADER})"
}

_get_header() {
  grep -i -F "${1}" "${RESPONSE_HEADERS}" | tr -d '\r'
}

# debug

print_headers() {
  cat "${RESPONSE_HEADERS}"
}

print_body() {
  cat "${RESPONSE_BODY}"
}
