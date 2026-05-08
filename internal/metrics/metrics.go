package metrics

import (
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal HTTP 请求总数
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTPRequestDuration HTTP 请求持续时间
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// APIRequestsTotal API 请求计数（按模型）
	APIRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"model", "provider"},
	)

	// APITokensTotal API Token 使用量
	APITokensTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_tokens_total",
			Help: "Total number of API tokens consumed",
		},
		[]string{"model", "type"}, // type: input/output
	)

	// APICostTotal API 费用
	APICostTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_cost_total",
			Help: "Total API cost in currency",
		},
		[]string{"model", "currency"},
	)

	// ConcurrentConnections 并发连接数
	ConcurrentConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "concurrent_connections",
			Help: "Current number of concurrent connections",
		},
	)

	// ActiveUsers 活跃用户数
	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Current number of active users",
		},
	)

	// ErrorsTotal 错误率
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "location"},
	)
)

// Go runtime metrics - using custom names to avoid conflicts with standard collectors
var (
	// GoGoroutines 当前 goroutine 数量
	GoGoroutines = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "ai_gateway_go_goroutines",
			Help: "Number of goroutines",
		},
		func() float64 {
			return float64(getGoroutineCount())
		},
	)

	// GoMemstatsAllocBytes 当前内存分配
	GoMemstatsAllocBytes = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "ai_gateway_go_memstats_alloc_bytes",
			Help: "Current memory allocation in bytes",
		},
		func() float64 {
			return float64(getAllocBytes())
		},
	)
)

// getGoroutineCount 获取当前 goroutine 数量
func getGoroutineCount() int {
	return runtime.NumGoroutine()
}

// getAllocBytes 获取当前内存分配
func getAllocBytes() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}
