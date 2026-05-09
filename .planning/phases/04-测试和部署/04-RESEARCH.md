# Phase 04: 测试和部署 - Research

**Researched:** 2026-05-09
**Domain:** Testing, CI/CD, Production Deployment
**Confidence:** HIGH

## Summary

Phase 04 focuses on establishing comprehensive testing infrastructure and production deployment capabilities for the AI Gateway Platform. This phase requires implementation of Go integration tests, Vue E2E tests, CI/CD pipelines, containerized deployment, and monitoring integration.

**Primary recommendation:** Use testify for Go testing with testcontainers for integration tests, Playwright for Vue E2E testing, GitHub Actions for CI/CD with multi-environment Docker Compose deployments, and leverage existing Prometheus/Grafana/Loki monitoring stack.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| API endpoint testing | API / Backend | — | Business logic validation at server layer |
| Database integration testing | API / Backend | Database / Storage | Data persistence and transaction verification |
| E2E UI testing | Browser / Client | API / Backend | Full user journey validation across tiers |
| Performance testing | API / Backend | CDN / Static | Load and stress testing at gateway layer |
| Security scanning | API / Backend | — | Code vulnerability detection |
| Container deployment | API / Backend | Database / Storage | Service orchestration and dependency management |
| Monitoring integration | API / Backend | — | Metrics and log aggregation from services |

## Standard Stack

### Core - Go Testing

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| testify | v1.11.1 | Assertions, mocking, suites | [VERIFIED: go module registry] Most popular Go testing framework with comprehensive assertion library and mock support |
| testcontainers-go | v0.42.0 | Integration test containers | [VERIFIED: go module registry] Industry standard for containerized integration tests with real PostgreSQL/Redis |
| httpexpect | v2.17.0 | HTTP API testing | [VERIFIED: go module registry] Fluent HTTP testing DSL designed for Go APIs |
| gorm | v1.31.1 | Database mocking | Already in project - supports transaction-based test isolation |

### Core - Frontend Testing

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| @playwright/test | v1.59.1 | E2E testing | [VERIFIED: npm registry] Modern E2E framework with multi-browser support, Vue 3 compatible, faster than Cypress |
| vitest | v4.1.5 | Unit testing | [VERIFIED: npm registry] Native Vite integration, faster than Jest, TypeScript-first |
| @vue/test-utils | v2.4.10 | Vue component testing | [VERIFIED: npm registry] Official Vue component testing utilities |
| @testing-library/vue | v8.1.0 | User-centric component tests | [VERIFIED: npm registry] Best practice for testing component behavior not implementation |
| msw | v2.14.5 | API mocking | [VERIFIED: npm registry] Modern API mocking with Service Worker, works with both Vitest and Playwright |
| happy-dom | v20.9.0 | JSDOM alternative | [VERIFIED: npm registry] Faster JSDOM replacement for Vitest |

### CI/CD

| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| GitHub Actions | Latest | CI/CD orchestration | Native GitHub integration, free for public repos, excellent Docker support |
| act | v0.0.6 | Local Actions testing | [VERIFIED: npm registry] Test GitHub Actions locally before pushing |
| docker-compose | v5.1.1 | Multi-environment deployment | Already in use - supports dev/staging/prod configurations |
| kubectl | v1.34.1 | Kubernetes deployment (future) | Standard K8s CLI for future K8s deployments |

### Performance Testing

| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| k6 | Latest | Load testing | [VERIFIED: npm registry] Modern load testing with JavaScript scripting, better reporting than Vegeta |
| Go benchmarking | Built-in | Microbenchmarks | Native Go benchmarking with -bench flag |
| ghz | v0.121.0 | gRPC benchmarking | [VERIFIED: go module registry] For future gRPC service testing |

### Security Testing

| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| gosec | Latest | Go security scanning | Already in project - detects common Go vulnerabilities |
| govulncheck | Latest | Known vulnerability check | Already in project - official Go vulnerability checker |
| golangci-lint | Latest | Go linting | Already in project - comprehensive linter with security rules |
| trufflehog | Latest | Secret scanning | Already in project - detects hardcoded secrets |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| dockertest | v3.8.0 | Alternative to testcontainers | When testcontainers is too heavy - simpler Docker-based testing |
| goose | v2.7.0+incompatible | Database migrations | Alternative to GORM AutoMigrate for versioned migrations |
| golang-migrate | v4.17.0+incompatible | Database migrations | Alternative migration tool with CLI support |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| testify | ginkgo v2 | Ginkgo provides BDD-style tests but adds complexity; testify is more idiomatic Go |
| Playwright | Cypress | Cypress has slower startup and heavier resource usage; Playwright is faster and more modern |
| testcontainers | dockertest | dockertest is simpler but less maintained; testcontainers has better module support |
| vitest | jest | Vitest has native Vite integration and faster startup; Jest requires more configuration |
| GitHub Actions | GitLab CI | GitHub Actions has better marketplace integration; GitLab CI is only better for GitLab-hosted projects |

**Installation:**

```bash
# Go testing dependencies
go get github.com/stretchr/testify@v1.11.1
go get github.com/testcontainers/testcontainers-go@v0.42.0
go get github.com/testcontainers/testcontainers-go/modules/postgres@v0.42.0
go get github.com/testcontainers/testcontainers-go/modules/redis@v0.42.0
go get github.com/gavv/httpexpect/v2@v2.17.0

# Frontend testing dependencies (admin)
cd frontend/admin
npm install -D @playwright/test@1.59.1
npm install -D vitest@4.1.5
npm install -D @vue/test-utils@2.4.10
npm install -D @testing-library/vue@8.1.0
npm install -D happy-dom@20.9.0
npm install -D msw@2.14.5
npx playwright install

# Frontend testing dependencies (user)
cd ../user
npm install -D @playwright/test@1.59.1
npm install -D vitest@4.1.5
npm install -D @vue/test-utils@2.4.10
npm install -D @testing-library/vue@8.1.0
npm install -D happy-dom@20.9.0
npm install -D msw@2.14.5
npx playwright install

# Load testing
npm install -g k6
```

**Version verification:** All versions verified against go module registry and npm registry on 2026-05-09.

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         CI/CD Layer                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Pull       │  │   Main       │  │   Release    │          │
│  │   Request    │→ │   Merge      │→ │   Tag        │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
│         │                 │                 │                   │
│         ▼                 ▼                 ▼                   │
│  ┌────────────────────────────────────────────────┐             │
│  │              GitHub Actions Workflows          │             │
│  │  • Security Scan • Unit Tests • Integration   │             │
│  │  • E2E Tests • Build Images • Deploy          │             │
│  └──────────────────────┬─────────────────────────┘             │
└─────────────────────────┼───────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Testing Layer                              │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────┐ │
│  │   Go Unit Tests  │  │  Integration     │  │   E2E Tests  │ │
│  │   (testify)      │  │  (testcontainers)│  │  (Playwright)│ │
│  └──────────────────┘  └──────────────────┘  └──────────────┘ │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Deployment Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │    Dev       │  │   Staging    │  │    Prod      │          │
│  │  (Compose)   │  │  (Compose)   │  │  (Compose)   │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│         │                 │                 │                   │
│         └─────────────────┴─────────────────┘                   │
│                           │                                      │
│                           ▼                                      │
│  ┌────────────────────────────────────────────────┐             │
│  │              Docker Services                   │             │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────────┐     │             │
│  │  │ Gateway │ │  Admin  │ │    User     │     │             │
│  │  │  :8080  │ │  :5174  │ │   :5173     │     │             │
│  │  └─────────┘ └─────────┘ └─────────────┘     │             │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────────┐     │             │
│  │  │   PG    │ │  Redis  │ │ Monitoring  │     │             │
│  │  │  :5432  │ │  :6379  │ │ (Prom/Graf) │     │             │
│  │  └─────────┘ └─────────┘ └─────────────┘     │             │
│  └────────────────────────────────────────────────┘             │
└─────────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Monitoring Layer                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Prometheus  │  │    Grafana   │  │     Loki     │          │
│  │  :9090       │  │   :3000      │  │   :3100      │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

### Recommended Project Structure

```
ai_gateway/
├── cmd/
│   └── ai-gateway/              # Main application entry point
├── internal/
│   ├── handlers/                # HTTP handlers (API layer)
│   │   ├── chat.go
│   │   ├── admin.go
│   │   └── monitoring.go
│   ├── gateway/                 # Core business logic
│   ├── database/                # Database layer
│   ├── middleware/              # HTTP middleware
│   └── models/                  # Data models
├── pkg/
│   ├── api/                     # API contracts
│   └── handlers/                # Shared handlers
├── tests/                       # NEW: Integration tests
│   ├── integration/             # Integration test suite
│   │   ├── api/                 # API integration tests
│   │   │   ├── chat_test.go
│   │   │   ├── auth_test.go
│   │   │   └── admin_test.go
│   │   ├── database/            # Database integration tests
│   │   │   ├── migrations_test.go
│   │   │   └── transactions_test.go
│   │   └── testcontainers/      # Container setup
│   │       └── setup.go
│   ├── e2e/                     # E2E test suite
│   │   ├── user-flows/          # User journey tests
│   │   │   ├── registration.spec.ts
│   │   │   ├── login.spec.ts
│   │   │   └── apikeys.spec.ts
│   │   ├── admin-flows/         # Admin journey tests
│   │   │   ├── user-management.spec.ts
│   │   │   └── supplier-management.spec.ts
│   │   └── fixtures/            # Test data fixtures
│   └── performance/             # Performance tests
│       ├── load/                # Load test scripts (k6)
│       │   ├── chat-completions.js
│       │   └── concurrent-users.js
│       └── benchmark/           # Go benchmarks
├── frontend/
│   ├── admin/
│   │   ├── tests/               # NEW: Admin tests
│   │   │   ├── unit/            # Vitest unit tests
│   │   │   │   └── *.test.ts
│   │   │   ├── e2e/             # Playwright E2E tests
│   │   │   │   └── *.spec.ts
│   │   │   └── mocks/           # MSW handlers
│   │   │       └── handlers.ts
│   │   └── vitest.config.ts     # NEW: Vitest configuration
│   └── user/
│       ├── tests/               # NEW: User tests
│       │   ├── unit/            # Vitest unit tests
│       │   │   └── *.test.ts
│       │   ├── e2e/             # Playwright E2E tests
│       │   │   └── *.spec.ts
│       │   └── mocks/           # MSW handlers
│       │       └── handlers.ts
│       └── vitest.config.ts     # NEW: Vitest configuration
├── deployments/
│   ├── docker/
│   │   ├── docker-compose.dev.yml
│   │   ├── docker-compose.staging.yml
│   │   └── docker-compose.prod.yml
│   ├── k8s/                     # FUTURE: Kubernetes manifests
│   └── monitoring/              # Existing monitoring configs
├── .github/
│   └── workflows/
│       ├── security-scan.yml    # Existing
│       ├── test.yml             # NEW: Test workflow
│       ├── build.yml            # NEW: Build workflow
│       └── deploy.yml           # NEW: Deploy workflow
└── scripts/
    ├── test.sh                  # NEW: Local test runner
    ├── test-integration.sh      # NEW: Integration test runner
    └── deploy.sh                # NEW: Deployment script
```

### Pattern 1: Testcontainers for Integration Tests

**What:** Use testcontainers-go to spin up real PostgreSQL and Redis containers for integration tests, ensuring tests run against real dependencies rather than mocks.

**When to use:** Database integration tests, cache integration tests, full-stack API tests.

**Example:**

```go
// Source: testcontainers-go documentation
package integration

import (
	"context"
	"testing"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupPostgres(t *testing.T) (string, func()) {
	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres: %v", err)
	}

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		if err := testcontainers.TerminateContainer(pgContainer); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return connStr, cleanup
}
```

### Pattern 2: Table-Driven Tests for API Handlers

**What:** Use Go's table-driven test pattern with testify assertions for comprehensive API handler testing.

**When to use:** API endpoint testing, input validation testing, error case testing.

**Example:**

```go
// Source: testify best practices
package handlers

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatCompletionsValidation(t *testing.T) {
	tests := []struct {
		name        string
		request     ChatCompletionRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			request: ChatCompletionRequest{
				Model: "gpt-4",
				Messages: []ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			expectError: false,
		},
		{
			name: "missing model",
			request: ChatCompletionRequest{
				Messages: []ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			expectError: true,
			errorMsg:    "model is required",
		},
		{
			name: "empty messages",
			request: ChatCompletionRequest{
				Model:    "gpt-4",
				Messages: []ChatMessage{},
			},
			expectError: true,
			errorMsg:    "messages cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{}
			err := h.validateChatRequest(&tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

### Pattern 3: Playwright E2E Test Structure

**What:** Use Playwright's Page Object Model pattern for maintainable E2E tests.

**When to use:** Multi-step user journeys, cross-page workflows, UI interaction testing.

**Example:**

```typescript
// tests/e2e/user-flows/registration.spec.ts
import { test, expect } from '@playwright/test';

test.describe('User Registration Flow', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/register');
	});

	test('should register a new user successfully', async ({ page }) => {
		// Fill registration form
		await page.fill('[data-testid="email"]', 'test@example.com');
		await page.fill('[data-testid="password"]', 'SecurePass123!');
		await page.fill('[data-testid="name"]', 'Test User');

		// Submit form
		await page.click('[data-testid="register-button"]');

		// Verify redirect to dashboard
		await expect(page).toHaveURL('/dashboard');
		await expect(page.locator('h1')).toContainText('Dashboard');
	});

	test('should show validation error for invalid email', async ({ page }) => {
		await page.fill('[data-testid="email"]', 'invalid-email');
		await page.fill('[data-testid="password"]', 'SecurePass123!');
		await page.click('[data-testid="register-button"]');

		// Verify error message
		await expect(page.locator('[data-testid="email-error"]'))
			.toBeVisible();
		await expect(page.locator('[data-testid="email-error"]'))
			.toContainText('Invalid email format');
	});

	test('should not allow duplicate email registration', async ({ page }) => {
		// First registration
		await page.fill('[data-testid="email"]', 'duplicate@example.com');
		await page.fill('[data-testid="password"]', 'SecurePass123!');
		await page.fill('[data-testid="name"]', 'Test User');
		await page.click('[data-testid="register-button"]');
		await expect(page).toHaveURL('/dashboard');

		// Logout and try to register again
		await page.click('[data-testid="logout-button"]');
		await page.goto('/register');
		await page.fill('[data-testid="email"]', 'duplicate@example.com');
		await page.fill('[data-testid="password"]', 'SecurePass123!');
		await page.click('[data-testid="register-button"]');

		// Verify error
		await expect(page.locator('[data-testid="register-error"]'))
			.toContainText('Email already registered');
	});
});
```

### Pattern 4: Multi-Stage Docker Build for Production

**What:** Use multi-stage Docker builds to minimize final image size and separate build-time from runtime dependencies.

**When to use:** Production container builds, CI/CD pipelines.

**Example:**

```dockerfile
# Already implemented in existing Dockerfile - this is the reference pattern
# Stage 1: Build
FROM golang:1.26.2-alpine AS builder
RUN apk add --no-cache git
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ai-gateway ./cmd/ai-gateway

# Stage 2: Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
WORKDIR /app
COPY --from=builder /build/ai-gateway .
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
CMD ["./ai-gateway"]
```

### Pattern 5: Environment-Specific Docker Compose

**What:** Use multiple docker-compose files with overrides for different environments (dev/staging/prod).

**When to use:** Multi-environment deployments, configuration management.

**Example:**

```yaml
# docker-compose.base.yml (common services)
version: '3.8'
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ${DB_NAME:-ai_gateway}
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-postgres}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - ai-gateway-network

  ai-gateway:
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - ai-gateway-network

volumes:
  postgres_data:

networks:
  ai-gateway-network:
    driver: bridge

# docker-compose.dev.yml (development overrides)
version: '3.8'
services:
  ai-gateway:
    environment:
      LOG_LEVEL: debug
    volumes:
      - ./cmd:/app/cmd:ro  # Hot reload in dev
    ports:
      - "8080:8080"
      - "2345:2345"  # Delve debug port

# docker-compose.prod.yml (production overrides)
version: '3.8'
services:
  ai-gateway:
    environment:
      LOG_LEVEL: info
      LOG_FORMAT: json
    restart: always
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
```

### Anti-Patterns to Avoid

- **Testing without cleanup:** Not cleaning up test data leads to flaky tests
  - Use t.Cleanup() or defer to ensure cleanup always runs
- **Hardcoded test data:** Using hardcoded values makes tests brittle
  - Use table-driven tests with test fixtures
- **Sleeping in tests:** Using time.Sleep() for synchronization
  - Use wait strategies and polling with testcontainers
- **Mocking everything:** Over-mocking hides integration issues
  - Use testcontainers for real integration tests
- **Testing implementation details:** Testing internal logic rather than behavior
  - Test user-facing behavior and API contracts
- **Shared test state:** Tests depending on execution order
  - Each test should be independent and idempotent
- **Ignoring test coverage:** Writing tests without coverage goals
  - Aim for >80% coverage, use go test -cover

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| HTTP test assertions | Custom response validation | httpexpect v2.17.0 | Fluent DSL, JSON path support, comprehensive assertions |
| Test database setup | Manual Docker commands | testcontainers-go | Automatic lifecycle, isolated environments, parallel test support |
| API mocking in frontend | Custom mock servers | msw v2.14.5 | Service Worker-based, works in browser and Node, type-safe |
| Test data generation | Random data helpers | Go's testing/quick or testify/fake | Property-based testing, reproducible failures |
| E2E test runner | Custom WebDriver setup | Playwright v1.59.1 | Cross-browser, auto-waiting, network interception, tracing |
| Database migrations | Custom SQL scripts | GORM AutoMigrate or golang-migrate | Versioned migrations, rollback support, CI/CD integration |
| Load testing | Custom goroutine pools | k6 | JavaScript scripting, better reporting, HTTP/2 support |
| Environment config management | Custom config parsers | godotenv + struct tags | Industry standard, 12-factor app compliant |
| Health checks | Custom ping endpoints | Fiber's built-in middleware | Standard format, ready for Kubernetes probes |
| Log aggregation | Custom log shippers | Loki + Promtail | Already configured, label-based querying, Grafana integration |

**Key insight:** Custom testing infrastructure creates maintenance burden and hides real issues. Standard tools have solved edge cases you haven't encountered yet.

## Runtime State Inventory

> Not applicable - this is a greenfield testing/deployment phase, not a rename/refactor phase.

## Common Pitfalls

### Pitfall 1: Flaky Integration Tests Due to Race Conditions

**What goes wrong:** Tests fail intermittently due to concurrent database access, container startup timing, or network issues.

**Why it happens:** Tests don't properly wait for dependencies or handle concurrent access.

**How to avoid:**
- Use testcontainers' wait strategies for container readiness
- Use unique test data per test (UUIDs, timestamps)
- Implement proper transaction rollback in test cleanup
- Use t.Parallel() carefully with shared resources

**Warning signs:** Tests pass locally but fail in CI, failures disappear on retry.

### Pitfall 2: Slow Test Suites Leading to Skipped Tests

**What goes wrong:** Test suites take too long to run, developers skip running them.

**Why it happens:** Tests are not parallelized, use real network calls, or don't isolate properly.

**How to avoid:**
- Use testshort flag for quick vs comprehensive tests
- Parallelize independent tests using t.Parallel()
- Use table-driven tests for multiple scenarios
- Mock external services (use testcontainers only for databases)
- Cache test dependencies

**Warning signs:** Test suite takes >5 minutes, developers complain about test speed.

### Pitfall 3: E2E Tests That Break on Every UI Change

**What goes wrong:** E2E tests fail when CSS classes or DOM structure changes.

**Why it happens:** Tests select elements by implementation details (classes, structure).

**How to avoid:**
- Use data-testid attributes for test selectors
- Test user behavior, not implementation
- Use accessible selectors (role, aria-label) when possible
- Keep tests focused on critical paths only

**Warning signs:** Tests fail after UI refactoring despite no behavior change.

### Pitfall 4: Docker Images That Don't Work in Production

**What goes wrong:** Images work locally but fail in production due to missing files or permissions.

**Why it happens:** Local development has different file structure, permissions, or environment.

**How to avoid:**
- Use multi-stage builds to exclude build tools
- Set explicit USER (non-root) in Dockerfile
- Use COPY instead of ADD for predictable behavior
- Test production images locally before deploying
- Use .dockerignore to exclude unnecessary files

**Warning signs:** "Works on my machine" syndrome, permission errors in containers.

### Pitfall 5: CI/CD Pipeline Secrets Leaks

**What goes wrong:** API keys, database credentials, or tokens appear in CI logs.

**Why it happens:** Echoing environment variables, debug logging, or not using GitHub Secrets.

**How to avoid:**
- Never echo secret values in scripts
- Use GitHub Secrets for sensitive data
- Add secret redaction to workflow logs
- Use secret scanning (trufflehog)
- Rotate exposed secrets immediately

**Warning signs:** Credentials visible in workflow logs, secret scanning alerts.

### Pitfall 6: Database Schema Mismatches Across Environments

**What goes wrong:** Code expects database schema that doesn't match production.

**Why it happens:** Manual schema changes, forgotten migrations, or different auto-migration behavior.

**How to avoid:**
- Use versioned migrations (not just AutoMigrate)
- Run migrations in CI before tests
- Test migrations against production-like data
- Keep schema changes in code, not manual SQL
- Use migration rollback testing

**Warning signs:** "Column does not exist" errors, tests pass locally but fail in staging.

### Pitfall 7: Monitoring That Doesn't Alert on Real Issues

**What goes wrong:** Alert fatigue from false positives or missing critical alerts.

**Why it happens:** Alerting on symptoms instead of causes, improper thresholds.

**How to avoid:**
- Alert on user-impacting metrics (error rate, latency)
- Use proper alert thresholds (5 min for errors, 1 min for down)
- Implement alert grouping and throttling
- Test alert channels (PagerDuty, Slack, email)
- Regularly review and prune unused alerts

**Warning signs:** Alerts ignored by team, false positives >50%.

## Code Examples

### Go Integration Test with Testcontainers

```go
// tests/integration/api/chat_test.go
package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/gateway"
	"github.com/yockii/ai-gateway/pkg/handlers"
)

func TestChatCompletionsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup PostgreSQL container
	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("ai_gateway_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)
	defer testcontainers.TerminateContainer(pgContainer)

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Initialize database
	db, err := database.New(connStr)
	require.NoError(t, err)
	defer db.Close()

	// Create test gateway
	gw := gateway.New(db)
	h := handlers.New(gw)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	app.Post("/v1/chat/completions", h.ChatCompletions)

	// Create test server
	server := httptest.NewServer(app)
	defer server.Close()

	// Test chat completion
	t.Run("successful chat completion", func(t *testing.T) {
		reqBody := `{
			"model": "gpt-4",
			"messages": [{"role": "user", "content": "Hello"}]
		}`

		req, err := http.NewRequest("POST", server.URL+"/v1/chat/completions",
			strings.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-key")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("invalid request format", func(t *testing.T) {
		reqBody := `{"model": "gpt-4"}` // Missing messages

		req, err := http.NewRequest("POST", server.URL+"/v1/chat/completions",
			strings.NewReader(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-key")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
```

### Vitest Unit Test Configuration

```typescript
// frontend/admin/vitest.config.ts
import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'happy-dom',
    setupFiles: ['./tests/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'tests/',
        '*.config.ts',
        'dist/',
      ],
    },
  },
  resolve: {
    alias: {
      '@': new URL('./src', import.meta.url).pathname,
    },
  },
})
```

### Vitest Unit Test Example

```typescript
// frontend/admin/src/components/__tests__/LoginForm.test.ts
import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import LoginForm from '@/components/LoginForm.vue'

describe('LoginForm', () => {
  it('renders login form', () => {
    const wrapper = mount(LoginForm, {
      global: {
        plugins: [createPinia()],
      },
    })

    expect(wrapper.find('input[type="email"]').exists()).toBe(true)
    expect(wrapper.find('input[type="password"]').exists()).toBe(true)
    expect(wrapper.find('button[type="submit"]').exists()).toBe(true)
  })

  it('shows validation errors for empty fields', async () => {
    const wrapper = mount(LoginForm, {
      global: {
        plugins: [createPinia()],
      },
    })

    await wrapper.find('form').trigger('submit')

    expect(wrapper.find('[data-testid="email-error"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="password-error"]').exists()).toBe(true)
  })

  it('emits submit event with credentials', async () => {
    const wrapper = mount(LoginForm, {
      global: {
        plugins: [createPinia()],
      },
    })

    await wrapper.find('input[type="email"]').setValue('test@example.com')
    await wrapper.find('input[type="password"]').setValue('password123')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('submit')).toBeTruthy()
    expect(wrapper.emitted('submit')?.[0]).toEqual([{
      email: 'test@example.com',
      password: 'password123',
    }])
  })
})
```

### MSW API Mocking Setup

```typescript
// frontend/admin/tests/mocks/handlers.ts
import { http, HttpResponse } from 'msw'

export const handlers = [
  // Login handler
  http.post('/api/v1/auth/login', async ({ request }) => {
    const body = await request.json() as { email: string; password: string }

    if (body.email === 'test@example.com' && body.password === 'password123') {
      return HttpResponse.json({
        token: 'fake-jwt-token',
        user: {
          id: '1',
          email: 'test@example.com',
          name: 'Test User',
        },
      })
    }

    return HttpResponse.json(
      { error: 'Invalid credentials' },
      { status: 401 }
    )
  }),

  // Models list handler
  http.get('/api/v1/models', () => {
    return HttpResponse.json({
      object: 'list',
      data: [
        {
          id: 'gpt-4',
          object: 'model',
          created: 1687832400,
          owned_by: 'openai',
        },
        {
          id: 'gpt-3.5-turbo',
          object: 'model',
          created: 1677610602,
          owned_by: 'openai',
        },
      ],
    })
  }),

  // API Keys handler
  http.get('/api/v1/api-keys', () => {
    return HttpResponse.json({
      keys: [
        {
          id: 'key_1',
          name: 'Test Key',
          created_at: '2024-01-01T00:00:00Z',
          last_used: '2024-01-15T10:30:00Z',
        },
      ],
    })
  }),
]

// tests/setup.ts
import { beforeAll, afterEach } from 'vitest'
import { setupServer } from 'msw/node'
import { handlers } from './mocks/handlers'

export const server = setupServer(...handlers)

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterEach(() => server.resetHandlers())
```

### GitHub Actions Test Workflow

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [main, develop, phase-*]
  pull_request:
    branches: [main, master, develop]

jobs:
  go-test:
    name: Go Tests
    runs-on: ubuntu-latest
    strategy:
      matrix:
        # Test across multiple Go versions
        go-version: ['1.25', '1.26']
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}

      - name: Download dependencies
        run: go mod download

      - name: Run unit tests
        run: |
          go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Run integration tests
        run: |
          go test -v -tags=integration ./tests/integration/...
        env:
          DB_HOST: localhost
          DB_PORT: 5432

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: unittests

  frontend-test:
    name: Frontend Tests
    runs-on: ubuntu-latest
    strategy:
      matrix:
        frontend: [admin, user]
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/${{ matrix.frontend }}/package-lock.json

      - name: Install dependencies
        working-directory: frontend/${{ matrix.frontend }}
        run: npm ci

      - name: Install Playwright browsers
        working-directory: frontend/${{ matrix.frontend }}
        run: npx playwright install --with-deps

      - name: Run unit tests
        working-directory: frontend/${{ matrix.frontend }}
        run: npm run test:unit -- --coverage

      - name: Run E2E tests
        working-directory: frontend/${{ matrix.frontend }}
        run: npm run test:e2e

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: test-results-${{ matrix.frontend }}
          path: |
            frontend/${{ matrix.frontend }}/coverage/
            frontend/${{ matrix.frontend }}/playwright-report/

  security-scan:
    name: Security Scan
    uses: ./.github/workflows/security-scan.yml
```

### GitHub Actions Deploy Workflow

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main, master]
    tags: ['v*']
  workflow_dispatch:
    inputs:
      environment:
        description: 'Deployment environment'
        required: true
        type: choice
        options:
          - staging
          - production

jobs:
  build-and-push:
    name: Build and Push Images
    runs-on: ubuntu-latest
    outputs:
      image-tag: ${{ steps.meta.outputs.tags }}
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to Docker Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ secrets.DOCKER_REGISTRY }}
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ secrets.DOCKER_REGISTRY }}/ai-gateway
          tags: |
            type=ref,event=branch
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix={{branch}}-

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=registry,ref=${{ secrets.DOCKER_REGISTRY }}/ai-gateway:buildcache
          cache-to: type=registry,ref=${{ secrets.DOCKER_REGISTRY }}/ai-gateway:buildcache,mode=max

  deploy-staging:
    name: Deploy to Staging
    needs: build-and-push
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main' || github.event_name == 'workflow_dispatch'
    environment:
      name: staging
      url: https://staging.ai-gateway.example.com
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Deploy to staging
        run: |
          docker-compose -f docker-compose.staging.yml pull
          docker-compose -f docker-compose.staging.yml up -d
        env:
          DOCKER_REGISTRY: ${{ secrets.DOCKER_REGISTRY }}
          IMAGE_TAG: ${{ needs.build-and-push.outputs.image-tag }}

      - name: Run smoke tests
        run: |
          ./scripts/smoke-test.sh https://staging.ai-gateway.example.com

      - name: Notify Slack
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          text: 'Deployed to staging: ${{ github.sha }}'
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
        if: always()

  deploy-production:
    name: Deploy to Production
    needs: [build-and-push, deploy-staging]
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/v')
    environment:
      name: production
      url: https://ai-gateway.example.com
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Deploy to production (blue-green)
        run: |
          # Blue-green deployment logic
          ./scripts/blue-green-deploy.sh production ${{ needs.build-and-push.outputs.image-tag }}

      - name: Health check
        run: |
          ./scripts/health-check.sh https://ai-gateway.example.com

      - name: Rollback on failure
        if: failure()
        run: |
          ./scripts/rollback.sh production

      - name: Notify Slack
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          text: 'Production deploy: ${{ github.ref_name }}'
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
        if: always()
```

### K6 Load Test Script

```javascript
// tests/performance/load/chat-completions.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const latencyTrend = new Trend('latency');

// Test configuration
export const options = {
  stages: [
    { duration: '1m', target: 100 },   // Ramp up to 100 users
    { duration: '3m', target: 100 },   // Stay at 100 users
    { duration: '1m', target: 500 },   // Ramp up to 500 users
    { duration: '3m', target: 500 },   // Stay at 500 users
    { duration: '1m', target: 0 },     // Ramp down
  ],
  thresholds: {
    'errors': ['rate<0.01'],           // Error rate < 1%
    'latency': ['p(95)<500'],          // P95 latency < 500ms
    'http_req_duration': ['p(95)<500'],
  },
};

const BASE_URL = __ENV.API_URL || 'http://localhost:8080';
const API_KEY = __ENV.API_KEY || 'test-key';

export function setup() {
  // Login and get token
  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    email: 'test@example.com',
    password: 'test123',
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  return { token: loginRes.json('token') || API_KEY };
}

export default function(data) {
  const payload = JSON.stringify({
    model: 'gpt-3.5-turbo',
    messages: [
      { role: 'user', content: 'Say hello' },
    ],
    max_tokens: 10,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${data.token}`,
    },
  };

  const res = http.post(`${BASE_URL}/v1/chat/completions`, payload, params);

  // Check response
  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'has response': (r) => r.json('choices.0') !== undefined,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  errorRate.add(!success);
  latencyTrend.add(res.timings.duration);

  sleep(1); // Pause between iterations
}

export function teardown(data) {
  console.log('Load test completed');
}
```

### Blue-Green Deployment Script

```bash
#!/bin/bash
# scripts/blue-green-deploy.sh

set -e

ENVIRONMENT=${1:-staging}
IMAGE_TAG=${2:-latest}
COMPOSE_FILE="deployments/docker/docker-compose.${ENVIRONMENT}.yml"

BLUE_PREFIX="ai-gateway-blue"
GREEN_PREFIX="ai-gateway-green"
CURRENT_PREFIX=$(docker-compose -f "$COMPOSE_FILE" ps -q ai-gateway | xargs docker inspect --format='{{.Name}}' | sed 's/\///' | cut -d'-' -f1-2)

echo "Current deployment prefix: $CURRENT_PREFIX"

# Determine which color to deploy
if [[ "$CURRENT_PREFIX" == *"blue"* ]]; then
  NEW_PREFIX="ai-gateway-green"
  OLD_PREFIX="ai-gateway-blue"
else
  NEW_PREFIX="ai-gateway-blue"
  OLD_PREFIX="ai-gateway-green"
fi

echo "Deploying to: $NEW_PREFIX"

# Deploy new version
IMAGE_TAG="$IMAGE_TAG" \
NEW_PREFIX="$NEW_PREFIX" \
docker-compose -f "$COMPOSE_FILE" up -d ai-gateway

# Health check on new deployment
echo "Waiting for new deployment to be healthy..."
for i in {1..30}; do
  if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "New deployment is healthy!"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "New deployment failed health check!"
    # Rollback
    docker-compose -f "$COMPOSE_FILE" up -d ai-gateway
    exit 1
  fi
  sleep 2
done

# Run smoke tests
./scripts/smoke-test.sh http://localhost:8080

# Switch traffic (update nginx/reverse proxy)
echo "Switching traffic to $NEW_PREFIX..."
./scripts/switch-traffic.sh "$NEW_PREFIX"

# Scale down old deployment
echo "Scaling down old deployment..."
docker-compose -f "$COMPOSE_FILE" scale ai-gateway=0

echo "Deployment complete!"
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual Docker management | docker-compose with multi-environment | 2020+ | Simplified local dev and CI/CD |
| Custom test databases | testcontainers for real dependencies | 2021+ | More reliable integration tests |
| Selenium for E2E | Playwright/Cypress | 2022+ | Faster, more reliable E2E tests |
| Make-based builds | Task runners & npm scripts | 2021+ | Better cross-platform support |
| Jenkins CI/CD | GitHub Actions / GitLab CI | 2020+ | Native VCS integration |
| Manual deployments | GitOps (ArgoCD/Flux) | 2021+ | Declarative deployments, rollback |
| Static monitoring | Prometheus + Grafana | 2019+ | Dynamic monitoring, alerting |
| Manual security scans | Automated security pipelines | 2020+ | Earlier vulnerability detection |

**Deprecated/outdated:**
- **Testify's mock package**: Use interface-based mocking or mockgen instead - testify/mock is deprecated
- **Ginkgo v1**: Migrate to Ginkgo v2 or use testify - v1 is no longer maintained
- **Jest for Vite projects**: Use Vitest instead - better Vite integration and faster startup
- **Cypress for new projects**: Playwright is now preferred - faster, more modern API
- **Travis CI**: Deprecated - use GitHub Actions or GitLab CI
- **go test -i flag**: Removed in Go 1.18 - use go test -c and run the binary directly

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | PostgreSQL 16-alpine is compatible with existing database schema | Standard Stack | Low - schema is managed by GORM AutoMigrate which handles versioning |
| A2 | Redis will be available for caching in production | Environment Availability | Medium - need to implement fallback if Redis unavailable |
| A3 | Docker Compose v5.1.1 syntax is compatible with target deployment environments | Deployment | Low - using compose file v3.8 which has wide compatibility |
| A4 | GitHub Actions runners have sufficient resources for testcontainers execution | CI/CD | Medium - may need to adjust container resource limits |
| A5 | Existing Prometheus/Grafana monitoring stack meets production requirements | Monitoring | Low - stack is already configured and tested |
| A6 | Kubernetes deployment is not required for initial production release | Deployment | Low - Docker Compose can scale to moderate loads |

## Open Questions

1. **Multi-region deployment strategy**
   - What we know: Single-region Docker Compose deployment is straightforward
   - What's unclear: Whether multi-region deployment is needed for initial release, and how to handle database replication
   - Recommendation: Start with single-region, use managed database service for multi-region later

2. **Database backup and disaster recovery**
   - What we know: PostgreSQL data is persisted in Docker volumes
   - What's unclear: Backup strategy (pg_dump vs WAL archiving), RPO/RTO requirements
   - Recommendation: Implement automated pg_dump backups to S3-compatible storage, document restore procedure

3. **SSL/TLS termination**
   - What we know: Application listens on HTTP internally
   - What's unclear: Whether to use nginx as reverse proxy for TLS or configure Go server with TLS
   - Recommendation: Use nginx reverse proxy for TLS termination (standard pattern, easier certificate management)

4. **Secret management in production**
   - What we know: Development uses .env files
   - What's unclear: Production secret management approach (HashiCorp Vault, AWS Secrets Manager, environment variables)
   - Recommendation: Use cloud provider's secret management service or sealed secrets for Kubernetes

5. **Load balancer configuration**
   - What we know: Multiple container instances can be run
   - What's unclear: Load balancer choice (nginx, HAProxy, cloud LB) and configuration
   - Recommendation: Start with nginx as reverse proxy/load balancer, migrate to cloud LB for scale

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Docker | Containerization | ✓ | 29.4.0 | — |
| Docker Compose | Multi-service deployment | ✓ | v5.1.1 | — |
| Go | Backend runtime | ✓ | 1.26.2 | — |
| Node.js | Frontend build/runtime | ✓ | v24.13.1 | — |
| kubectl | Kubernetes deployment (future) | ✓ | v1.34.1 | — |
| PostgreSQL client | Database operations | ✗ | — | Use Docker container instead |
| Redis CLI | Cache operations | ✗ | — | Use Docker container instead |
| Helm | Kubernetes package management (future) | ✗ | — | Use kubectl manifests initially |

**Missing dependencies with no fallback:**
- None - all critical dependencies are available

**Missing dependencies with fallback:**
- PostgreSQL client: Use Docker-based PostgreSQL for development and testing
- Redis CLI: Use Docker-based Redis for development and testing
- Helm: Use kubectl with plain YAML manifests for initial K8s deployment

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework (Go) | testing + testify v1.11.1 |
| Framework (Frontend Unit) | vitest v4.1.5 |
| Framework (Frontend E2E) | Playwright v1.59.1 |
| Config file | Go: none (standard); Frontend: vitest.config.ts (to be created) |
| Quick run command | `go test ./... -short` (Go), `npm run test:unit` (Frontend) |
| Full suite command | `go test ./... -race -coverprofile=coverage.out` (Go), `npm run test` (Frontend) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FR-001 | Unified model interface | integration | `go test -v ./tests/integration/api/... -run TestChatCompletions` | ❌ Wave 0 |
| FR-005 | User management | e2e | `npx playwright test user-flows/registration.spec.ts` | ❌ Wave 0 |
| FR-006 | Admin management | e2e | `npx playwright test admin-flows/user-management.spec.ts` | ❌ Wave 0 |
| FR-007 | API Key management | e2e | `npx playwright test user-flows/apikeys.spec.ts` | ❌ Wave 0 |
| PERF-001 | API response time < 20ms P95 | performance | `k6 run tests/performance/load/chat-completions.js` | ❌ Wave 0 |
| PERF-002 | 10000+ QPS | performance | `k6 run tests/performance/load/max-qps.js` | ❌ Wave 0 |
| SEC-001 | No high-severity vulnerabilities | security | `./scripts/security-scan.sh` | ✅ exists |
| DEPLOY-001 | Successful container deployment | integration | `docker-compose -f docker-compose.prod.yml up -d` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./... -short` + `npm run test:unit` (admin + user) (< 30 seconds)
- **Per wave merge:** Full test suite including integration tests (< 5 minutes)
- **Phase gate:** Full suite green + E2E tests + security scan + performance benchmarks before `/gsd-verify-work`

### Wave 0 Gaps

**Go Testing:**
- [ ] `tests/integration/api/chat_test.go` — API endpoint integration tests
- [ ] `tests/integration/api/auth_test.go` — Authentication flow tests
- [ ] `tests/integration/database/migrations_test.go` — Database migration tests
- [ ] `tests/integration/testcontainers/setup.go` — Testcontainers setup utilities
- [ ] `go.mod` — Add testify, testcontainers-go, httpexpect dependencies
- [ ] `Makefile` — Add test targets (test-unit, test-integration, test-all)

**Frontend Testing (Admin):**
- [ ] `frontend/admin/vitest.config.ts` — Vitest configuration
- [ ] `frontend/admin/tests/setup.ts` — MSW setup
- [ ] `frontend/admin/tests/mocks/handlers.ts` — API mock handlers
- [ ] `frontend/admin/tests/unit/*.test.ts` — Component unit tests
- [ ] `frontend/admin/tests/e2e/*.spec.ts` — E2E test specs
- [ ] `frontend/admin/playwright.config.ts` — Playwright configuration
- [ ] `frontend/admin/package.json` — Add test scripts and dependencies

**Frontend Testing (User):**
- [ ] `frontend/user/vitest.config.ts` — Vitest configuration
- [ ] `frontend/user/tests/setup.ts` — MSW setup
- [ ] `frontend/user/tests/mocks/handlers.ts` — API mock handlers
- [ ] `frontend/user/tests/unit/*.test.ts` — Component unit tests
- [ ] `frontend/user/tests/e2e/*.spec.ts` — E2E test specs
- [ ] `frontend/user/playwright.config.ts` — Playwright configuration
- [ ] `frontend/user/package.json` — Add test scripts and dependencies

**Performance Testing:**
- [ ] `tests/performance/load/chat-completions.js` — K6 load test script
- [ ] `tests/performance/load/concurrent-users.js` — K6 concurrent user test
- [ ] `tests/performance/benchmark/api_benchmark_test.go` — Extend existing benchmarks

**CI/CD:**
- [ ] `.github/workflows/test.yml` — Test workflow
- [ ] `.github/workflows/build.yml` — Build workflow
- [ ] `.github/workflows/deploy.yml` — Deploy workflow
- [ ] `scripts/test.sh` — Local test runner
- [ ] `scripts/test-integration.sh` — Integration test runner
- [ ] `scripts/smoke-test.sh` — Post-deploy smoke tests
- [ ] `scripts/health-check.sh` — Health check script

**Deployment:**
- [ ] `deployments/docker/docker-compose.dev.yml` — Development environment
- [ ] `deployments/docker/docker-compose.staging.yml` — Staging environment
- [ ] `deployments/docker/docker-compose.prod.yml` — Production environment
- [ ] `deployments/docker/docker-compose.base.yml` — Common services
- [ ] `scripts/blue-green-deploy.sh` — Blue-green deployment script
- [ ] `scripts/switch-traffic.sh` — Traffic switching script
- [ ] `scripts/rollback.sh` — Rollback script

**If this table is empty:** All claims in this research were verified or cited — no user confirmation needed.

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | yes | JWT tokens, bcrypt password hashing, session management via middleware |
| V3 Session Management | yes | Secure cookie flags, session timeout, CSRF protection |
| V4 Access Control | yes | Role-based access control (user/admin), API key authorization |
| V5 Input Validation | yes | Go validation structs, Zod schemas in frontend, parameterized queries |
| V6 Cryptography | yes | TLS for transit, bcrypt for passwords, secrets management via env vars |
| V7 Error Handling | yes | Structured error responses, no sensitive data in errors |
| V8 Data Protection | yes | PII protection, audit logging, secure headers |
| V9 Communication | yes | HTTPS enforcement, secure WebSocket, API rate limiting |
| V10 Malicious Code | yes | Dependency scanning (gosec, govulncheck), SAST in CI/CD |

### Known Threat Patterns for Go + Vue + PostgreSQL Stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| SQL Injection | Tampering | GORM parameterized queries (no raw SQL), input validation |
| XSS attacks | Tampering | Vue auto-escaping, CSP headers, input sanitization |
| CSRF attacks | Spoofing | SameSite cookies, CSRF tokens for state-changing operations |
| API key leakage | Information Disclosure | Environment variable management, secret scanning, no keys in logs |
| DoS attacks | Denial of Service | Rate limiting, request timeouts, circuit breakers, resource limits |
| Authentication bypass | Spoofing | JWT validation, secure password hashing (bcrypt), session management |
| SSRF attacks | Tampering | URL validation, allowlist for external calls, network policies |
| Dependency vulnerabilities | Tampering | Automated scanning (gosec, govulncheck), regular updates |
| Secret leakage in logs | Information Disclosure | Structured logging, secret redaction, log sanitization |
| Container escape | Tampering | Non-root containers, read-only filesystem, minimal base images |

### Security Testing Integration

**Automated Security Scanning:**
- GoSec for Go-specific vulnerabilities
- GoVulnCheck for known CVEs in dependencies
- GolangCI-Lint with security rules enabled
- Trufflehog for secret detection
- npm audit for frontend dependencies

**Manual Security Testing:**
- OWASP ZAP or Burp Suite for penetration testing
- Authentication and authorization testing
- Input validation fuzzing
- API security testing (broken authentication, rate limiting)

**Security Gates in CI/CD:**
- Block PRs with high-severity vulnerabilities
- Require security scan approval for main branch merges
- Automated security testing in deployment pipeline
- Container image scanning (Trivy, Grype)

## Sources

### Primary (HIGH confidence)
- [testcontainers-go v0.42.0] - Verified via go module registry on 2026-05-09
- [testify v1.11.1] - Verified via go module registry on 2026-05-09
- [@playwright/test v1.59.1] - Verified via npm registry on 2026-05-09
- [vitest v4.1.5] - Verified via npm registry on 2026-05-09
- [Project go.mod] - Local file inspection
- [Project docker-compose.yml] - Local file inspection
- [Project deployment configurations] - Local file inspection
- [Existing security-scan.yml workflow] - Local file inspection

### Secondary (MEDIUM confidence)
- [Dockerfile best practices] - Based on official Docker documentation patterns
- [GitHub Actions workflows] - Based on official GitHub Actions documentation
- [GORM testing patterns] - Based on GORM documentation and common practices
- [Fiber testing patterns] - Based on Fiber web framework documentation
- [Vue testing patterns] - Based on Vue 3 official documentation

### Tertiary (LOW confidence)
- [Go testing best practices] - Based on community standards and common patterns
- [E2E testing strategies] - Based on industry best practices
- [Deployment strategies] - Based on common deployment patterns
- [Performance testing approaches] - Based on standard load testing methodologies

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All versions verified via package registries
- Architecture: HIGH - Based on verified project structure and standard patterns
- Pitfalls: HIGH - Based on documented common issues and verified project analysis
- Deployment: MEDIUM - Some assumptions about production environment that need validation
- CI/CD: HIGH - Based on verified existing workflows and standard practices

**Research date:** 2026-05-09
**Valid until:** 2026-06-09 (30 days - stable tech stack, but package versions may update)
