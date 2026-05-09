#!/usr/bin/env bash
set -euo pipefail

# AI Gateway Test Runner
# Runs all tests (unit + integration) with coverage reporting

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

GO_TEST_FLAGS="${GO_TEST_FLAGS:-}"
COVERAGE_FILE="${COVERAGE_FILE:-coverage.out}"
COVERAGE_HTML="${COVERAGE_HTML:-coverage.html}"
TEST_TIMEOUT="${TEST_TIMEOUT:-10m}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print usage
usage() {
    cat << EOF
Usage: $(basename "$0") [OPTIONS]

Run all tests for the AI Gateway project.

OPTIONS:
    -s, --short       Run quick tests (skip integration and load tests)
    -v, --verbose     Enable verbose output
    -c, --coverage    Generate coverage report (default)
    --no-coverage     Skip coverage report
    -h, --help        Show this help message

ENVIRONMENT VARIABLES:
    GO_TEST_FLAGS     Additional flags to pass to go test
    TEST_TIMEOUT      Test timeout (default: 10m)
    COVERAGE_FILE     Coverage output file (default: coverage.out)

EXAMPLES:
    $(basename "$0")                    # Run all tests with coverage
    $(basename "$0") -s                 # Run quick tests only
    $(basename "$0") -v -c              # Verbose output with coverage

EOF
    exit 0
}

# Parse arguments
SHORT=false
VERBOSE=false
COVERAGE=true
SKIP_INTEGRATION=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--short)
            SHORT=true
            SKIP_INTEGRATION=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            GO_TEST_FLAGS="$GO_TEST_FLAGS -v"
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        --no-coverage)
            COVERAGE=false
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Print header
echo "========================================"
echo "  AI Gateway Test Runner"
echo "========================================"
echo ""

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Run unit tests
echo "Running unit tests..."
echo "----------------------------------------"

UNIT_TEST_CMD="go test $GO_TEST_FLAGS -timeout $TEST_TIMEOUT"
if [ "$COVERAGE" = true ]; then
    UNIT_TEST_CMD="$UNIT_TEST_CMD -coverprofile=$COVERAGE_FILE -covermode=atomic"
fi

UNIT_TEST_CMD="$UNIT_TEST_CMD ./..."

if eval "$UNIT_TEST_CMD"; then
    ((PASSED_TESTS++))
    echo -e "${GREEN}Unit tests: PASSED${NC}"
else
    ((FAILED_TESTS++))
    echo -e "${RED}Unit tests: FAILED${NC}"
    exit 1
fi

# Count packages tested
PACKETS=$(go list ./... | wc -l)
((TOTAL_TESTS += PACKETS))

echo ""

# Run integration tests
if [ "$SKIP_INTEGRATION" = false ]; then
    echo "Running integration tests..."
    echo "----------------------------------------"

    if [ -d "tests/integration" ]; then
        INTEGRATION_CMD="go test $GO_TEST_FLAGS -timeout $TEST_TIMEOUT -tags=integration ./tests/integration/..."

        if eval "$INTEGRATION_CMD"; then
            ((PASSED_TESTS++))
            echo -e "${GREEN}Integration tests: PASSED${NC}"
        else
            ((FAILED_TESTS++))
            echo -e "${YELLOW}Integration tests: FAILED (may require environment setup)${NC}"
        fi
    else
        echo -e "${YELLOW}No integration tests found${NC}"
        ((SKIPPED_TESTS++))
    fi
else
    echo -e "${YELLOW}Skipping integration tests (short mode)${NC}"
    ((SKIPPED_TESTS++))
fi

echo ""

# Generate coverage report
if [ "$COVERAGE" = true ] && [ -f "$COVERAGE_FILE" ]; then
    echo "Generating coverage report..."
    echo "----------------------------------------"

    # Get coverage percentage
    COVERAGE_PERCENT=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}')
    echo "Total coverage: $COVERAGE_PERCENT"

    # Generate HTML report
    go tool cover -html="$COVERAGE_FILE" -o "$COVERAGE_HTML"
    echo "HTML coverage report: $COVERAGE_HTML"

    # Check coverage threshold
    COVERAGE_NUM=$(echo "$COVERAGE_PERCENT" | sed 's/%//')
    if (( $(echo "$COVERAGE_NUM < 50" | bc -l) )); then
        echo -e "${YELLOW}Warning: Coverage is below 50%${NC}"
    fi

    echo ""
fi

# Print summary
echo "========================================"
echo "  Test Summary"
echo "========================================"
echo ""
echo "Tests run:     $TOTAL_TESTS"
echo -e "Passed:        ${GREEN}$PASSED_TESTS${NC}"
if [ "$FAILED_TESTS" -gt 0 ]; then
    echo -e "Failed:        ${RED}$FAILED_TESTS${NC}"
else
    echo "Failed:        $FAILED_TESTS"
fi
if [ "$SKIPPED_TESTS" -gt 0 ]; then
    echo -e "Skipped:       ${YELLOW}$SKIPPED_TESTS${NC}"
fi

if [ "$COVERAGE" = true ]; then
    echo "Coverage:      $COVERAGE_PERCENT"
fi

echo ""

# Exit with appropriate code
if [ "$FAILED_TESTS" -gt 0 ]; then
    exit 1
else
    exit 0
fi
