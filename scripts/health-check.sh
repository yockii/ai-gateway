#!/usr/bin/env bash
set -euo pipefail

# AI Gateway Health Check
# Wait for the service to be healthy

# Configuration
BASE_URL="${1:-http://localhost:8080}"
TIMEOUT="${2:-300}"  # Default 5 minutes
INTERVAL="${HEALTH_CHECK_INTERVAL:-5}"  # Check every 5 seconds

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Print usage
if [[ "${1:-}" == "-h" ]] || [[ "${1:-}" == "--help" ]]; then
    cat << EOF
Usage: $(basename "$0") [BASE_URL] [TIMEOUT]

Wait for the AI Gateway service to become healthy.

ARGUMENTS:
    BASE_URL    Base URL of the API (default: http://localhost:8080)
    TIMEOUT     Maximum wait time in seconds (default: 300)

ENVIRONMENT VARIABLES:
    HEALTH_CHECK_INTERVAL  Check interval in seconds (default: 5)

EXIT CODES:
    0    Service is healthy
    1    Timeout waiting for service
    2    Invalid arguments

EXAMPLES:
    $(basename "$0")                              # Wait for localhost:8080 (5 min)
    $(basename "$0") http://localhost:8080 60     # Wait 1 minute only
    $(basename "$0") https://api.example.com 600  # Wait 10 minutes for remote

EOF
    exit 0
fi

# Validate URL
if [[ ! "$BASE_URL" =~ ^https?:// ]]; then
    echo -e "${RED}Error: Invalid URL format: $BASE_URL${NC}"
    echo "URL must start with http:// or https://"
    exit 2
fi

# Print header
echo "========================================"
echo "  AI Gateway Health Check"
echo "========================================"
echo ""
echo "Base URL:  $BASE_URL"
echo "Timeout:   ${TIMEOUT}s"
echo "Interval:  ${INTERVAL}s"
echo ""

# Calculate end time
END_TIME=$(($(date +%s) + TIMEOUT))

# Wait loop
ATTEMPT=0
while true; do
    ((ATTEMPT++))
    CURRENT_TIME=$(date +%s)

    # Check timeout
    if [ "$CURRENT_TIME" -ge "$END_TIME" ]; then
        echo ""
        echo -e "${RED}Timeout: Service did not become healthy within ${TIMEOUT}s${NC}"
        echo ""
        echo "Troubleshooting:"
        echo "  1. Check if the service is running: docker ps | grep ai-gateway"
        echo "  2. Check service logs: docker logs ai-gateway"
        echo "  3. Verify the service is listening on the correct port"
        echo "  4. Check firewall rules if testing a remote server"
        exit 1
    fi

    # Print status
    REMAINING=$((END_TIME - CURRENT_TIME))
    echo -ne "\rChecking health... (attempt $ATTEMPT, ${REMAINING}s remaining) "

    # Perform health check
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time 5 "$BASE_URL/health" 2>&1) || true

    # Extract status code
    STATUS_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n -1)

    # Check if healthy
    if [ "$STATUS_CODE" = "200" ]; then
        # Optionally verify body content
        if [ -n "$BODY" ] && [[ ! "$BODY" =~ (OK|healthy|{"status) ]]; then
            # Response doesn't look right, but status is 200
            :
        fi

        echo ""
        echo ""
        echo -e "${GREEN}Service is healthy!${NC}"
        echo ""
        echo "Details:"
        echo "  Attempts:    $ATTEMPT"
        echo "  Status code: $STATUS_CODE"
        echo "  Response:    $BODY"
        echo ""
        exit 0
    fi

    # Log progress every minute
    if [ $((ATTEMPT * INTERVAL % 60)) -eq 0 ] && [ "$ATTEMPT" -gt 1 ]; then
        echo ""
        echo -e "${YELLOW}Still waiting... (status: $STATUS_CODE)${NC}"
    fi

    # Wait before next check
    sleep "$INTERVAL"
done
