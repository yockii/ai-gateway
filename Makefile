.PHONY: help build run test-unit test-integration test-quick test-all test-coverage clean

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  test-unit     - Run unit tests with coverage"
	@echo "  test-integration - Run integration tests with testcontainers"
	@echo "  test-quick    - Run quick tests (unit only, short mode)"
	@echo "  test-all      - Run all tests (unit + integration)"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"

# Build the application
build:
	@echo "Building ai-gateway..."
	@go build -o bin/ai-gateway ./cmd/ai-gateway
	@echo "Build complete: bin/ai-gateway"

# Run the application
run: build
	@echo "Running ai-gateway..."
	@./bin/ai-gateway

# Unit tests (quick, no containers)
test-unit:
	@echo "Running unit tests..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Unit tests complete. Coverage report: coverage.html"

# Integration tests (with containers)
test-integration:
	@echo "Running integration tests..."
	@go test -v -tags=integration -timeout 5m ./tests/integration/...
	@echo "Integration tests complete."

# Quick test for development (unit only, short mode)
test-quick:
	@echo "Running quick tests..."
	@go test -v -short ./...
	@echo "Quick tests complete."

# All tests
test-all: test-unit test-integration
	@echo "All tests complete."

# Test with coverage report
test-coverage: test-unit
	@echo ""
	@echo "Coverage Summary:"
	@go tool cover -func=coverage.out | grep total
	@echo ""
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete."
