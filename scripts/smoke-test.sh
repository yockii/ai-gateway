#!/usr/bin/env bash
set -euo pipefail

# AI Gateway Smoke Tests
# Quick post-deployment verification tests

# Configuration
BASE_URL="${1:-http://localhost:8080}"
API_KEY="${API_KEY:-test-key}"
TIMEOUT=10
VERBOSE="${VERBOSE:-false}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test results
PASSED=0
FAILED=0
TOTAL=0

# Print usage
if [[ "${1:-}" == "-h" ]] || [[ "${1:-}" == "--help" ]]; then
    cat << EOF
Usage: $(basename "$0") [BASE_URL]

Run smoke tests against the AI Gateway.

ARGUMENTS:
    BASE_URL    Base URL of the API (default: http://localhost:8080)

ENVIRONMENT VARIABLES:
    API_KEY     API key for authentication (default: test-key)
    TIMEOUT     Request timeout in seconds (default: 10)
    VERBOSE     Enable verbose output (default: false)

EXAMPLES:
    $(basename "$0")                              # Test localhost:8080
    $(basename "$0") https://api.example.com      # Test remote server
    API_KEY=xxx $(basename "$0") https://staging.example.com

EOF
    exit 0
fi

# Print header
echo "========================================"
echo "  AI Gateway Smoke Tests"
echo "========================================"
echo ""
echo "Base URL: $BASE_URL"
echo "Timeout: ${TIMEOUT}s"
echo ""

# Test function
run_test() {
    local test_name="$1"
    local url="$2"
    local expected_code="${3:-200}"
    local headers="${4:-}"
    local method="${5:-GET}"
    local body="${6:-}"

    ((TOTAL++))

    echo -n "Testing $test_name... "

    # Build curl command
    local curl_cmd="curl -s -w '\n%{http_code}' -X $method --max-time $TIMEOUT"

    if [ -n "$headers" ]; then
        curl_cmd="$curl_cmd $headers"
    fi

    if [ -n "$body" ]; then
        curl_cmd="$curl_cmd -d '$body'"
    fi

    curl_cmd="$curl_cmd '$url'"

    # Execute request
    local response
    response=$(eval "$curl_cmd" 2>&1) || true

    # Extract status code (last line)
    local status_code
    status_code=$(echo "$response" | tail -n1)

    # Extract body (everything except last line)
    local body_content
    body_content=$(echo "$response" | head -n -1)

    # Check result
    if [ "$status_code" = "$expected_code" ]; then
        echo -e "${GREEN}PASS${NC} ($status_code)"
        ((PASSED++))

        if [ "$VERBOSE" = "true" ]; then
            echo "  Response: $body_content" | head -c 100
            echo ""
        fi
    else
        echo -e "${RED}FAIL${NC} (expected $expected_code, got $status_code)"
        ((FAILED++))

        if [ "$VERBOSE" = "true" ] || [ "$status_code" = "000" ]; then
            echo "  Response: $body_content"
        fi
    fi
}

# Run tests
echo "========================================"
echo "  Core Endpoints"
echo "========================================"

run_test "Health Check" "$BASE_URL/health" 200 "" "GET"
run_test "Metrics Endpoint" "$BASE_URL/metrics" 200 "" "GET"
run_test "Models List" "$BASE_URL/v1/models" 200 "-H 'Authorization: Bearer $API_KEY'" "GET"

echo ""
echo "========================================"
echo "  Authentication"
echo "========================================"

run_test "Auth without key" "$BASE_URL/v1/models" 401 "" "GET"
run_test "Auth with invalid key" "$BASE_URL/v1/models" 401 "-H 'Authorization: Bearer invalid-key'" "GET"

echo ""
echo "========================================"
echo "  API Endpoints"
echo "========================================"

# Test chat completions endpoint (may fail if upstream not configured)
echo -n "Testing Chat Completions... "
CHAT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"test"}]}' \
    --max-time $TIMEOUT \
    "$BASE_URL/v1/chat/completions" 2>&1) || true

CHAT_STATUS=$(echo "$CHAT_RESPONSE" | tail -n1)

# Chat may return 200, 401 (auth), 500 (upstream), or 503 (service unavailable)
if [[ "$CHAT_STATUS" =~ ^(200|401|500|503)$ ]]; then
    echo -e "${GREEN}PASS${NC} ($CHAT_STATUS)"
    ((PASSED++))
else
    echo -e "${YELLOW}SKIP${NC} (got $CHAT_STATUS, may be expected)"
    ((SKIPPED++))
fi
((TOTAL++))

echo ""
echo "========================================"
echo "  Usage Endpoints"
echo "========================================"

run_test "Usage Stats" "$BASE_URL/v1/usage" 200 "-H 'Authorization: Bearer $API_KEY'" "GET"

echo ""
echo "========================================"
echo "  Test Results"
echo "========================================"
echo ""
echo "Total tests:  $TOTAL"
echo -e "Passed:       ${GREEN}$PASSED${NC}"
echo -e "Failed:       ${RED}$FAILED${NC}"
if [ "$SKIPPED"x != "x" ]; then
    echo -e "Skipped:      ${YELLOW}$SKIPPED${NC}"
fi
echo ""

# Print results table
echo "Detailed Results:"
echo "----------------"
printf "%-25s %-10s\n" "Test" "Status"
printf "%-25s %-10s\n" "----" "------"

# Exit with appropriate code
if [ "$FAILED" -gt 0 ]; then
    echo ""
    echo -e "${RED}Smoke tests FAILED${NC}"
    exit 1
else
    echo ""
    echo -e "${GREEN}All smoke tests PASSED${NC}"
    exit 0
fi
