#!/bin/bash
set -e

echo "======================================"
echo "  AI Gateway Performance Benchmark"
echo "======================================"
echo ""

# 检查服务器是否运行
echo "Checking if server is running..."
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "Server is not running. Starting server..."

    # 检查是否有 main.go
    if [ -f "cmd/main.go" ]; then
        go run cmd/main.go &
        SERVER_PID=$!
        echo "Server started with PID: $SERVER_PID"

        # 等待服务器启动
        echo "Waiting for server to be ready..."
        for i in {1..30}; do
            if curl -s http://localhost:8080/health > /dev/null 2>&1; then
                echo "Server is ready!"
                break
            fi
            if [ $i -eq 30 ]; then
                echo "Server failed to start within 30 seconds"
                exit 1
            fi
            sleep 1
        done
    else
        echo "Error: cmd/main.go not found. Please start the server manually."
        exit 1
    fi
else
    echo "Server is already running"
    SERVER_PID=""
fi

# 清理函数
cleanup() {
    if [ -n "$SERVER_PID" ]; then
        echo ""
        echo "Stopping server (PID: $SERVER_PID)..."
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        echo "Server stopped"
    fi
}
trap cleanup EXIT

# 创建结果目录
mkdir -p benchmark-results
cd benchmark-results

echo ""
echo "======================================"
echo "  Running Go Benchmarks"
echo "======================================"
echo ""

# 运行基准测试
echo "1. Running ListModels benchmark..."
go test -bench=BenchmarkListModels -benchmem -benchtime=5s ../benchmarks/ | tee listmodels.txt

echo ""
echo "2. Running HealthCheck benchmark..."
go test -bench=BenchmarkHealthCheck -benchmem -benchtime=5s ../benchmarks/ | tee healthcheck.txt

echo ""
echo "3. Running Middleware benchmark..."
go test -bench=BenchmarkMiddleware -benchmem -benchtime=5s ../benchmarks/ | tee middleware.txt

echo ""
echo "4. Running JSON serialization benchmark..."
go test -bench=BenchmarkJSON -benchmem -benchtime=5s ../benchmarks/ | tee json.txt

echo ""
echo "5. Running concurrent benchmark..."
go test -bench=BenchmarkConcurrent -benchmem -benchtime=5s ../benchmarks/ | tee concurrent.txt

echo ""
echo "======================================"
echo "  Running Response Time Tests"
echo "======================================"
echo ""

# 运行响应时间测试
echo "Testing API response times..."
go test -v -run=TestAPIResponseTime ../benchmarks/ 2>&1 | tee response-time.txt || true

echo ""
echo "======================================"
echo "  Running Load Test (if enabled)"
echo "======================================"
echo ""

# 运行负载测试（仅在非短模式下）
if [ -n "$RUN_LOAD_TEST" ]; then
    echo "Running QPS load test..."
    go test -v -run=TestQPSLoadTest ../benchmarks/ 2>&1 | tee loadtest.txt || true
else
    echo "Skipping load test (set RUN_LOAD_TEST=1 to enable)"
fi

echo ""
echo "======================================"
echo "  Benchmark Complete"
echo "======================================"
echo ""

# 生成摘要报告
echo "Generating summary report..."
cat > BENCHMARK_SUMMARY.md << EOF
# AI Gateway Performance Benchmark Summary

Generated: $(date)

## Benchmark Results

### ListModels Endpoint
- See \`listmodels.txt\` for detailed results

### HealthCheck Endpoint
- See \`healthcheck.txt\` for detailed results

### Middleware Performance
- See \`middleware.txt\` for detailed results

### JSON Serialization
- See \`json.txt\` for detailed results

### Concurrent Requests
- See \`concurrent.txt\` for detailed results

### Response Time Tests
- See \`response-time.txt\` for detailed results

### Load Test
- See \`loadtest.txt\` for detailed results (if enabled)

## Performance Targets

- API Gateway Response Time: < 20ms (P95)
- API Gateway QPS: 10000+
- Concurrent Limit Check: < 1ms
- Memory: Stable, no leaks

## How to Run

### Quick Benchmark (5 seconds per test)
\`\`\`bash
./scripts/benchmark.sh
\`\`\`

### Full Load Test
\`\`\`bash
RUN_LOAD_TEST=1 ./scripts/benchmark.sh
\`\`\`

### Specific Benchmark
\`\`\`bash
go test -bench=BenchmarkListModels -benchmem ./benchmarks/
\`\`\`

### Response Time Test
\`\`\`bash
go test -v -run=TestAPIResponseTime ./benchmarks/
\`\`\`

## Notes

- Benchmarks run against a local server on port 8080
- The script will start the server if not running
- Results are saved in \`benchmark-results/\` directory
- Load test is disabled by default (enable with RUN_LOAD_TEST=1)
EOF

echo "Results saved to: benchmark-results/"
echo "  - Individual benchmark results"
echo "  - BENCHMARK_SUMMARY.md"
echo ""
echo "View summary: cat benchmark-results/BENCHMARK_SUMMARY.md"
