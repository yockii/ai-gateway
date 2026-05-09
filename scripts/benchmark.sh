#!/usr/bin/env bash
set -euo pipefail

echo "======================================"
echo "  AI Gateway Performance Benchmark"
echo "======================================"
echo ""

# Configuration
BASE_URL="${API_BASE_URL:-http://localhost:8080}"
BENCHMARK_DIR="benchmarks/results"
SERVER_PID=""
REGRESSION_THRESHOLD=10  # Percentage

# Create results directory
mkdir -p "$BENCHMARK_DIR"

# Check server is running
check_server() {
    echo "Checking if server is running at $BASE_URL..."
    if ! curl -sf "$BASE_URL/health" > /dev/null 2>&1; then
        echo "Server is not running. Starting server..."

        if [ -f "cmd/ai-gateway/main.go" ]; then
            go run cmd/ai-gateway/main.go &
            SERVER_PID=$!
            echo "Server started with PID: $SERVER_PID"

            # Wait for server to be ready
            echo "Waiting for server to be ready..."
            for i in {1..60}; do
                if curl -sf "$BASE_URL/health" > /dev/null 2>&1; then
                    echo "Server is ready!"
                    return 0
                fi
                if [ $i -eq 60 ]; then
                    echo "Server failed to start within 60 seconds"
                    exit 1
                fi
                sleep 1
            done
        else
            echo "Error: cmd/ai-gateway/main.go not found. Please start the server manually."
            exit 1
        fi
    else
        echo "Server is already running"
    fi
}

# Cleanup function
cleanup() {
    if [ -n "$SERVER_PID" ]; then
        echo ""
        echo "Stopping server (PID: $SERVER_PID)..."
        kill "$SERVER_PID" 2>/dev/null || true
        wait "$SERVER_PID" 2>/dev/null || true
        echo "Server stopped"
    fi
}
trap cleanup EXIT

# Run benchmark
run_benchmark() {
    local name=$1
    local pattern=$2
    local output_file="$BENCHMARK_DIR/${name}.txt"

    echo ""
    echo "Running: $name"
    echo "----------------------------------------"
    go test -bench="$pattern" -benchmem -benchtime=5s ./benchmarks/ | tee "$output_file"

    # Check for regression
    if [ -f "$BENCHMARK_DIR/${name}.baseline" ]; then
        echo "Comparing against baseline..."
        if ! diff -u "$BENCHMARK_DIR/${name}.baseline" "$output_file"; then
            echo "WARNING: Results differ from baseline"
        fi
    fi
}

# Compare performance
check_regression() {
    local current=$1
    local baseline=$2
    local threshold=$3

    if [ ! -f "$baseline" ]; then
        echo "No baseline found, creating baseline..."
        cp "$current" "$baseline"
        return 0
    fi

    # Extract ns/op from current and baseline
    local current_ns=$(grep -oP '\d+(?= ns/op)' "$current" | tail -1)
    local baseline_ns=$(grep -oP '\d+(?= ns/op)' "$baseline" | tail -1)

    if [ -n "$current_ns" ] && [ -n "$baseline_ns" ]; then
        local diff=$((current_ns - baseline_ns))
        local percent=$((diff * 100 / baseline_ns))

        echo "Performance delta: $diff ns/op ($percent%)"

        if [ "$percent" -gt "$threshold" ]; then
            echo "ERROR: Performance regression detected! ($percent% > $threshold%)"
            return 1
        fi
    fi

    return 0
}

# Main execution
check_server

echo ""
echo "======================================"
echo "  Running Go Benchmarks"
echo "======================================"
echo ""

# Run all benchmarks
run_benchmark "listmodels" "BenchmarkListModels"
run_benchmark "healthcheck" "BenchmarkHealthCheck"
run_benchmark "middleware" "BenchmarkMiddleware"
run_benchmark "auth" "BenchmarkAuthMiddleware"
run_benchmark "cache" "BenchmarkCacheMiddleware"
run_benchmark "dbquery" "BenchmarkDBQuery"
run_benchmark "redis" "BenchmarkRedis"
run_benchmark "json" "BenchmarkJSON"
run_benchmark "concurrent" "BenchmarkConcurrent"
run_benchmark "chat" "BenchmarkChatCompletions"

echo ""
echo "======================================"
echo "  Running Response Time Tests"
echo "======================================"
echo ""

go test -v -run=TestAPIResponseTime ./benchmarks/ 2>&1 | tee "$BENCHMARK_DIR/response-time.txt"

echo ""
echo "======================================"
echo "  Running Load Test (if enabled)"
echo "======================================"
echo ""

if [ -n "${RUN_LOAD_TEST:-}" ]; then
    echo "Running QPS load test..."
    go test -v -run=TestQPSLoadTest ./benchmarks/ 2>&1 | tee "$BENCHMARK_DIR/loadtest.txt"
else
    echo "Skipping load test (set RUN_LOAD_TEST=1 to enable)"
fi

echo ""
echo "======================================"
echo "  Benchmark Complete"
echo "======================================"
echo ""

# Generate summary
cat > "$BENCHMARK_DIR/SUMMARY.md" << EOF
# AI Gateway Performance Benchmark Summary

**Generated:** $(date)
**Base URL:** $BASE_URL

## Benchmark Results

| Benchmark | ns/op | B/op | allocs/op | Status |
|-----------|-------|------|-----------|--------|
EOF

# Extract results for summary table
for file in "$BENCHMARK_DIR"/*.txt; do
    if [ -f "$file" ] && [ "$file" != "${BENCHMARK_DIR}/SUMMARY.md" ]; then
        name=$(basename "$file" .txt)
        if grep -q "Benchmark" "$file" 2>/dev/null; then
            stats=$(grep -E "Benchmark.*\d+\s+ns/op" "$file" | tail -1)
            if [ -n "$stats" ]; then
                echo "| $name | $stats |" >> "$BENCHMARK_DIR/SUMMARY.md"
            fi
        fi
    fi
done

cat >> "$BENCHMARK_DIR/SUMMARY.md" << 'EOF'

## Performance Targets

- **API Gateway Response Time:** < 20ms (P95)
- **API Gateway QPS:** 10000+
- **Memory:** Stable, no leaks
- **Regression Threshold:** 10%

## How to Run

### Quick Benchmark (5 seconds per test)
```bash
./scripts/benchmark.sh
```

### Full Load Test
```bash
RUN_LOAD_TEST=1 ./scripts/benchmark.sh
```

### Specific Benchmark
```bash
go test -bench=BenchmarkListModels -benchmem ./benchmarks/
```

### Update Baseline
```bash
cp benchmarks/results/*.txt benchmarks/results/*.baseline
```

## Notes

- Benchmarks run against a local server on port 8080
- The script will start the server if not running
- Results are saved in \`benchmarks/results/\` directory
- Load test is disabled by default (enable with RUN_LOAD_TEST=1)
EOF

echo "Results saved to: $BENCHMARK_DIR/"
echo "  - Individual benchmark results (*.txt)"
echo "  - SUMMARY.md"
echo ""
echo "View summary: cat $BENCHMARK_DIR/SUMMARY.md"

# Exit with error if regression detected
if ! check_regression "$BENCHMARK_DIR/listmodels.txt" "$BENCHMARK_DIR/listmodels.baseline" "$REGRESSION_THRESHOLD"; then
    exit 1
fi
