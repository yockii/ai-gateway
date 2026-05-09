---
phase: 04-测试和部署
plan: 03
subsystem: CI/CD Pipeline and Performance Testing
tags: [ci-cd, github-actions, docker, k6, benchmarks, testing, automation]
wave: 3
dependency_graph:
  requires: [04-01, 04-02]
  provides: [04-04]
  affects: [.github/workflows, tests/performance, scripts]
tech_stack:
  added:
    - "K6 load testing framework"
    - "GitHub Actions workflows (test.yml, build.yml, deploy.yml)"
    - "Docker Buildx for multi-platform builds"
    - "Trivy for container vulnerability scanning"
    - "Go benchmarking with b.ReportAllocs()"
  patterns:
    - "CI/CD pipeline with test, build, deploy stages"
    - "Multi-platform Docker image building (amd64/arm64)"
    - "Blue-green deployment for production"
    - "Performance regression detection"
    - "Automated smoke testing after deployment"
key_files:
  created:
    - path: tests/performance/load/chat-completions.js
      purpose: K6 load test for chat API with 100-500 concurrent users
    - path: tests/performance/load/concurrent-users.js
      purpose: K6 realistic user behavior simulation with mixed API calls
    - path: tests/performance/load/max-qps.js
      purpose: K6 maximum QPS test from 1000 to 15000 QPS
    - path: .github/workflows/test.yml
      purpose: CI test workflow with Go, frontend, performance, and security tests
    - path: .github/workflows/build.yml
      purpose: Docker build workflow with multi-platform support and vulnerability scanning
    - path: .github/workflows/deploy.yml
      purpose: Deployment workflow with staging and production environments
    - path: scripts/test.sh
      purpose: Run all tests with coverage reporting
    - path: scripts/smoke-test.sh
      purpose: Post-deployment smoke tests for all endpoints
    - path: scripts/health-check.sh
      purpose: Wait for service to be healthy with timeout
  modified:
    - path: benchmarks/api_benchmark_test.go
      changes: Added 8 new benchmarks (auth, cache, DB, Redis, response sizes, concurrent chat)
    - path: scripts/benchmark.sh
      changes: Added regression detection, summary report, and baseline comparison
decisions:
  - "Use K6 for load testing: Industry standard, good CI integration, JavaScript-based"
  - "Use GitHub Actions for CI/CD: Native integration, free for public repos, good community"
  - "Multi-platform Docker builds: Support both amd64 and arm64 architectures"
  - "Blue-green deployment: Zero-downtime deployments with automatic rollback"
  - "Performance regression threshold: 10% to avoid false positives"
metrics:
  duration: "0:09:24"
  completed_date: "2026-05-09T02:30:36Z"
  tasks_completed: 6
  files_created: 9
  test_files_created: 3
  workflows_created: 3
---

# Phase 04 Plan 03: CI/CD Pipeline and Performance Testing Summary

**One-liner:** Complete CI/CD pipeline established with GitHub Actions workflows for testing, building, and deploying Docker images, plus K6 load tests and Go benchmarks for performance validation.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed MaxTokens type error in benchmark**
- **Found during:** Task 2 - Go benchmark compilation
- **Issue:** MaxTokens field expects *int but benchmark passed int literal
- **Fix:** Changed to use pointer variable (maxTokens := 50; MaxTokens: &maxTokens)
- **Files modified:** benchmarks/api_benchmark_test.go
- **Commit:** 9359948

## What Was Built

### Task 1: K6 Load Test Scripts

#### chat-completions.js
- Ramp up to 100 users over 1 minute, sustain for 3 minutes
- Ramp up to 500 users over 1 minute, sustain for 3 minutes
- Ramp down to 0
- Custom metrics: error rate, chat latency
- Thresholds: error rate < 1%, P95 latency < 500ms
- Environment variables: API_URL, API_KEY

#### concurrent-users.js
- Realistic user behavior simulation
- Mix of API calls: chat (50%), models (20%), usage (20%), health (10%)
- Think time between requests (1-3 seconds)
- Stages: 10, 50, 100, 500 concurrent users
- Thresholds: P95 latency < 200ms overall, < 100ms for non-chat

#### max-qps.js
- Constant load test at 1000 RPS
- Ramping load test from 1000 to 15000 RPS
- Find breaking point through gradual ramp
- 80% health checks (fastest path), 20% models
- Thresholds: error rate < 5%, P95 latency < 100ms

### Task 2: Extended Go Benchmark Suite

#### New Benchmarks Added
1. **BenchmarkAuthMiddleware** - JWT validation overhead
2. **BenchmarkCacheMiddleware/Hit** - Cache hit performance
3. **BenchmarkCacheMiddleware/Miss** - Cache miss performance
4. **BenchmarkDBQuery/ListModels** - Database query for models
5. **BenchmarkDBQuery/UsageStats** - Database query for usage
6. **BenchmarkRedisGet** - Redis GET performance
7. **BenchmarkRedisSet** - Redis SET performance
8. **BenchmarkResponseSizes/SmallResponse** - Small response handling
9. **BenchmarkResponseSizes/MediumResponse** - Medium response handling
10. **BenchmarkConcurrentChatCompletions** - Concurrent chat requests

All benchmarks use:
- b.ResetTimer() to exclude setup time
- b.ReportAllocs() to report memory allocations
- b.RunParallel() for concurrent execution
- Realistic data sizes

#### Updated benchmark.sh
- Regression detection against baseline
- Summary table generation
- Proper error handling with set -euo pipefail
- Better output formatting

### Task 3: GitHub Actions Test Workflow

#### Jobs
1. **go-test** - Matrix testing on Go 1.25 and 1.26
   - Unit tests with coverage
   - Integration tests (non-PR)
   - Codecov upload
   - Coverage artifacts

2. **frontend-test** - Matrix for admin and user
   - Unit tests with coverage
   - E2E tests with Playwright (admin)
   - Test results artifacts
   - Codecov upload

3. **performance-test** - On main branch only
   - Run all benchmarks
   - Check for regression
   - Comment PR with results
   - Benchmark artifacts

4. **security-scan** - Runs in parallel
   - GoSec scanning
   - GoVulnCheck
   - Security scan script
   - SARIF upload to GitHub Security

5. **test-summary** - Aggregate results
   - Summarize all job results
   - Fail on critical failures

Features:
- Concurrency control (cancel outdated runs)
- Dependency caching for faster builds
- Timeout limits to prevent hanging
- Artifact uploads on failure

### Task 4: GitHub Actions Build Workflow

#### Jobs
1. **build-and-push** - Main application image
   - Multi-platform builds (linux/amd64, linux/arm64)
   - Docker Buildx with layer caching
   - Metadata extraction (tags, labels)
   - Push to GitHub Container Registry
   - Trivy vulnerability scanning
   - SARIF upload to GitHub Security

2. **build-frontend** - Frontend images
   - Matrix build for admin and user
   - Separate Dockerfile support
   - Build args for environment variables

3. **build-summary** - Aggregate build results

Image tags:
- Branch name
- SHA-based (branch-sha)
- Semver (for releases)
- Latest (default branch)

### Task 5: GitHub Actions Deploy Workflow

#### Jobs
1. **build-and-push** - Build images for deployment
2. **deploy-staging** - Automatic staging deployment
   - Pull latest images
   - Docker Compose deployment
   - Health check wait
   - Smoke tests
   - Failure notification (Slack)

3. **deploy-production** - Manual approval required
   - Blue-green deployment
   - Zero-downtime switching
   - Smoke tests on new environment
   - Automatic rollback on failure
   - Metrics verification
   - Old environment cleanup
   - Status notification

4. **post-deploy-test** - Post-deployment verification
   - Integration tests
   - Load tests with K6
   - Test result artifacts

Features:
- Environment-specific configurations
- Manual approval for production
- Blue-green deployment pattern
- Automatic rollback on failure
- Slack notification support

### Task 6: Test Automation Scripts

#### scripts/test.sh
- Run all tests (unit + integration)
- Support for short mode (-s)
- Verbose output (-v)
- Coverage reporting (default)
- Color-coded output
- Summary with pass/fail/skipped counts
- Exit code reflects test results

#### scripts/smoke-test.sh
- Test core endpoints (/health, /metrics, /v1/models)
- Test authentication scenarios (no key, invalid key)
- Test API endpoints (chat, usage)
- Table-formatted results
- Configurable base URL and API key
- Exit code indicates smoke test status

#### scripts/health-check.sh
- Wait for service to be healthy
- Configurable timeout (default 5 minutes)
- Progress reporting during wait
- Troubleshooting tips on timeout
- Useful for deployment scripts
- Exit code 0 when healthy, 1 on timeout

All scripts:
- Use proper shebang (#!/usr/bin/env bash)
- Use set -euo pipefail for error handling
- Have helpful usage messages (-h/--help)
- Are executable (chmod +x)

## Threat Surface Scan

### New Security-Relevant Surface

| Flag | File | Description |
|------|------|-------------|
| threat_flag: docker_registry | .github/workflows/build.yml | CI pushes to Docker registry with GITHUB_TOKEN |
| threat_flag: deployment_credentials | .github/workflows/deploy.yml | CI deploys to staging/production environments |
| threat_flag: load_test_targets | tests/performance/load/*.js | Load tests could target production if misconfigured |

### Mitigations Applied
- **T-04-11 (Spoofing)**: Branch protection required for production deployments
- **T-04-12 (Tampering)**: Trivy vulnerability scanning in build pipeline
- **T-04-13 (Information Disclosure)**: GitHub Actions secrets masking enabled
- **T-04-14 (DoS)**: Load tests default to localhost, API_URL env var required
- **T-04-15 (Escalation)**: Production deployment requires manual approval

## Known Stubs

None - all deliverables are complete and functional.

## Self-Check: PASSED

**Files created:**
- tests/performance/load/chat-completions.js ✓
- tests/performance/load/concurrent-users.js ✓
- tests/performance/load/max-qps.js ✓
- .github/workflows/test.yml ✓
- .github/workflows/build.yml ✓
- .github/workflows/deploy.yml ✓
- scripts/test.sh ✓
- scripts/smoke-test.sh ✓
- scripts/health-check.sh ✓

**Files modified:**
- benchmarks/api_benchmark_test.go ✓
- scripts/benchmark.sh ✓

**Commits verified:**
- d39f1af: create K6 load test scripts ✓
- 9359948: extend Go benchmark suite ✓
- 6bb12b6: create GitHub Actions test workflow ✓
- 94a1a6a: create GitHub Actions build workflow ✓
- d6f92c4: create GitHub Actions deploy workflow ✓
- 9bc13dc: create test automation scripts ✓
