#!/usr/bin/env bash

# Integration Test Runner Script
# This script runs integration tests with proper environment setup

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
TEST_TIMEOUT="5m"
VERBOSE=false
SHORT=false
COVERAGE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -s|--short)
            SHORT=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -t|--timeout)
            TEST_TIMEOUT="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -v, --verbose      Enable verbose output"
            echo "  -s, --short       Run in short mode (skip long tests)"
            echo "  -c, --coverage    Generate coverage report"
            echo "  -t, --timeout     Set test timeout (default: 5m)"
            echo "  -h, --help        Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                 # Run all integration tests"
            echo "  $0 -v              # Run with verbose output"
            echo "  $0 -s              # Run in short mode"
            echo "  $0 -c              # Run with coverage"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Check prerequisites
echo -e "${GREEN}Checking prerequisites...${NC}"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed or not in PATH${NC}"
    echo "Please install Docker Desktop from https://www.docker.com/products/docker-desktop"
    exit 1
fi

# Check if Docker is running
if ! docker info &> /dev/null; then
    echo -e "${RED}Error: Docker is not running${NC}"
    echo "Please start Docker Desktop and try again"
    exit 1
fi

echo -e "${GREEN}Docker is available and running${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}"
    exit 1
fi

echo -e "${GREEN}Go is available: $(go version)${NC}"

# Set environment variables for test configuration
export TEST_INTEGRATION=true
export CGO_ENABLED=0

# Build test command
TEST_CMD="go test"

if [ "$VERBOSE" = true ]; then
    TEST_CMD="$TEST_CMD -v"
fi

if [ "$SHORT" = true ]; then
    TEST_CMD="$TEST_CMD -short"
fi

if [ "$COVERAGE" = true ]; then
    TEST_CMD="$TEST_CMD -coverprofile=coverage-integration.out -covermode=atomic"
fi

TEST_CMD="$TEST_CMD -tags=integration -timeout $TEST_TIMEOUT ./tests/integration/..."

# Print test configuration
echo ""
echo -e "${YELLOW}Test Configuration:${NC}"
echo "  Command: $TEST_CMD"
echo "  Timeout: $TEST_TIMEOUT"
echo "  Short Mode: $SHORT"
echo "  Coverage: $COVERAGE"
echo ""

# Run integration tests
echo -e "${GREEN}Running integration tests...${NC}"
echo ""

START_TIME=$(date +%s)

if eval "$TEST_CMD"; then
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))

    echo ""
    echo -e "${GREEN}Integration tests passed successfully!${NC}"
    echo -e "Duration: ${DURATION}s"

    # Generate coverage report if requested
    if [ "$COVERAGE" = true ]; then
        echo ""
        echo -e "${GREEN}Generating coverage report...${NC}"
        go tool cover -html=coverage-integration.out -o coverage-integration.html
        echo -e "Coverage report: coverage-integration.html"
        echo ""
        go tool cover -func=coverage-integration.out | grep total
    fi

    echo ""
    echo -e "${GREEN}Integration tests complete.${NC}"
    exit 0
else
    EXIT_CODE=$?

    echo ""
    echo -e "${RED}Integration tests failed with exit code $EXIT_CODE${NC}"
    echo ""
    echo -e "${YELLOW}Troubleshooting tips:${NC}"
    echo "  1. Make sure Docker Desktop is running"
    echo "  2. Check that port 5432 and 6379 are not in use"
    echo "  3. Try running 'docker ps' to verify Docker is working"
    echo "  4. Check test logs above for specific errors"

    exit $EXIT_CODE
fi
