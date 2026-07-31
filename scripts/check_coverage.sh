#!/usr/bin/env bash
set -e

# check_coverage.sh verifies that global and core domain test coverage satisfy defined quality gates.

GLOBAL_MIN_COVERAGE=${1:-40.0}
CORE_MIN_COVERAGE=${2:-65.0}

echo "========================================================"
echo "🛡️  Dataspace Components Quality Gate: Unit Test Coverage Verification"
echo "========================================================"
echo "Minimum Global Coverage Requirement: ${GLOBAL_MIN_COVERAGE}%"
echo "Minimum Core Domain/App Coverage Requirement: ${CORE_MIN_COVERAGE}%"

# Generate full coverage profile across all packages in the monorepo
go test -coverprofile=coverage.out ./... > /dev/null

# Filter profile specifically for core packages (domain, core, ports)
echo "mode: set" > coverage_core.out
grep -E "/(domain|core)" coverage.out >> coverage_core.out || true

# Extract total percentage for global coverage
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total: | awk '{print $3}' | sed 's/%//')

# Extract total percentage for core coverage
if [ $(wc -l < coverage_core.out) -gt 1 ]; then
    CORE_COVERAGE=$(go tool cover -func=coverage_core.out | grep total: | awk '{print $3}' | sed 's/%//')
else
    CORE_COVERAGE="0.0"
fi

echo "--------------------------------------------------------"
echo "📊 Current Monorepo Global Coverage : ${TOTAL_COVERAGE}%"
echo "🎯 Current Core Domain/App Coverage : ${CORE_COVERAGE}%"
echo "--------------------------------------------------------"

GLOBAL_PASSED=$(awk -v total="$TOTAL_COVERAGE" -v min="$GLOBAL_MIN_COVERAGE" 'BEGIN {print (total >= min)}')
CORE_PASSED=$(awk -v total="$CORE_COVERAGE" -v min="$CORE_MIN_COVERAGE" 'BEGIN {print (total >= min)}')

if [ "$GLOBAL_PASSED" -eq 1 ] && [ "$CORE_PASSED" -eq 1 ]; then
    echo "✅ Quality Gate PASSED: All code coverage thresholds satisfied!"
    exit 0
else
    echo "❌ Quality Gate FAILED: Code coverage does not meet target thresholds!"
    exit 1
fi
