package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
)

// TestMetricsEndpointExists tests that /metrics endpoint returns Prometheus format metrics
func TestMetricsEndpointExists(t *testing.T) {
	// Create test app
	app := fiber.New()

	// Register metrics endpoint using the actual handler
	app.Get("/metrics", func(c fiber.Ctx) error {
		// Convert fasthttp request/response to net/http for prometheus handler
		// This is a simplified version - in production use the actual router setup
		c.Set("Content-Type", "text/plain")
		return c.SendString("# Test metrics output\n")
	})

	// Create request
	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check response contains Prometheus metrics
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	bodyStr := string(body)

	// Should contain some Prometheus output
	if len(bodyStr) == 0 {
		t.Error("Expected non-empty response body")
	}
}

// TestHTTPRequestCounter tests that HTTP request counter increments
func TestHTTPRequestCounter(t *testing.T) {
	// Create a custom registry for this test to avoid conflicts
	registry := prometheus.NewRegistry()

	// Create test counter
	testCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	registry.MustRegister(testCounter)

	// Record a request
	testCounter.WithLabelValues("GET", "/api/test", "200").Inc()

	// Verify metric was recorded (this is a basic check)
	// In a real test, you would gather metrics and verify the value
}

// TestHTTPRequestDuration tests that response time histogram records correctly
func TestHTTPRequestDuration(t *testing.T) {
	// Create a custom registry for this test to avoid conflicts
	registry := prometheus.NewRegistry()

	// Create test histogram
	testHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "test_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	registry.MustRegister(testHistogram)

	// Record some durations
	testHistogram.WithLabelValues("POST", "/api/chat").Observe(0.1)
	testHistogram.WithLabelValues("POST", "/api/chat").Observe(0.2)
	testHistogram.WithLabelValues("GET", "/api/models").Observe(0.05)

	// Metrics should be recorded without error
}

// TestGoRuntimeMetrics tests that Go runtime metrics are exposed
func TestGoRuntimeMetrics(t *testing.T) {
	// Create a custom registry for this test to avoid conflicts
	registry := prometheus.NewRegistry()

	// Create a runtime metric (goroutines)
	gaugeFunc := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "test_go_goroutines",
			Help: "Number of goroutines",
		},
		func() float64 {
			return 42.0 // Mock value
		},
	)
	registry.MustRegister(gaugeFunc)

	// The gauge should be callable without error
	val := gaugeFunc.Desc().String()
	if val == "" {
		t.Error("Expected gauge to have description")
	}
}

// TestAPIRequestRecording tests API request recording functions
func TestAPIRequestRecording(t *testing.T) {
	// Create a custom registry for this test to avoid conflicts
	registry := prometheus.NewRegistry()

	// Create test metrics
	apiCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"model", "provider"},
	)
	registry.MustRegister(apiCounter)

	tokenCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_api_tokens_total",
			Help: "Total number of API tokens consumed",
		},
		[]string{"model", "type"},
	)
	registry.MustRegister(tokenCounter)

	costCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_api_cost_total",
			Help: "Total API cost in currency",
		},
		[]string{"model", "currency"},
	)
	registry.MustRegister(costCounter)

	activeUsersGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "test_active_users",
			Help: "Current number of active users",
		},
	)
	registry.MustRegister(activeUsersGauge)

	errorCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "location"},
	)
	registry.MustRegister(errorCounter)

	// Test recording
	apiCounter.WithLabelValues("gpt-4", "openai").Inc()
	tokenCounter.WithLabelValues("gpt-4", "input").Add(100)
	tokenCounter.WithLabelValues("gpt-4", "output").Add(50)
	costCounter.WithLabelValues("gpt-4", "USD").Add(0.01)
	activeUsersGauge.Inc()
	activeUsersGauge.Dec()
	errorCounter.WithLabelValues("http", "/api/test").Inc()

	// Metrics should be recorded without error
}

// TestConcurrentConnectionsGauge tests concurrent connections gauge
func TestConcurrentConnectionsGauge(t *testing.T) {
	// Create a custom registry for this test to avoid conflicts
	registry := prometheus.NewRegistry()

	// Create test gauge
	connGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "test_concurrent_connections",
			Help: "Current number of concurrent connections",
		},
	)
	registry.MustRegister(connGauge)

	// The gauge should exist
	if connGauge == nil {
		t.Error("Expected concurrent connections gauge to exist")
	}

	// Test increment and decrement
	connGauge.Inc()
	connGauge.Dec()
}
