#!/bin/bash
set -euo pipefail

# Extract integration test cases from AllCases map in integration.go
# This script parses the Go file to find test case names automatically

INTEGRATION_FILE="build/testing/integration.go"

if [[ ! -f "$INTEGRATION_FILE" ]]; then
    echo "Error: $INTEGRATION_FILE not found" >&2
    exit 1
fi

# Extract test case names using sed
# Find the AllCases map and extract quoted strings
sed -n '/AllCases = map\[string\]testCaseFn{/,/^\t}/p' "$INTEGRATION_FILE" | \
    sed -n 's/^\t\t"\([^"]*\)".*/\1/p' | \
    sort