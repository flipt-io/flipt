#!/bin/bash
set -euo pipefail

# Extract integration test cases from AllCases map in integration.go
# This script parses the Go file to find test case names automatically

INTEGRATION_FILE="build/testing/integration.go"

if [[ ! -f "$INTEGRATION_FILE" ]]; then
    echo "Error: $INTEGRATION_FILE not found" >&2
    exit 1
fi

# Extract test case names from the AllCases map
# Look for lines between "AllCases = map[string]testCaseFn{" and the closing "}"
# Then extract the quoted strings (test case names)
grep -A 20 "AllCases = map\[string\]testCaseFn{" "$INTEGRATION_FILE" | \
    grep -B 20 "^\t}" | \
    grep -o '"[^"]*":' | \
    sed 's/"//g' | \
    sed 's/://g' | \
    sort