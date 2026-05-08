package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type MonitoringHandler struct {
	prometheusAPI v1.API
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(prometheusURL string) *MonitoringHandler {
	client, err := api.NewClient(api.Config{
		Address: prometheusURL,
	})
	if err != nil {
		// In production, handle this error properly
		// For now, we'll panic to fail fast
		panic(err)
	}

	return &MonitoringHandler{
		prometheusAPI: v1.NewAPI(client),
	}
}

// GetSystemMetrics 获取系统指标
func (h *MonitoringHandler) GetSystemMetrics(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	now := time.Now()

	// 查询 QPS
	qps, _, err := h.prometheusAPI.Query(ctx, "rate(http_requests_total[5m])", now)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 查询平均延迟
	latency, _, err := h.prometheusAPI.Query(ctx, "histogram_quantile(0.95, http_request_duration_seconds_bucket)", now)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 查询错误率
	errorRate, _, err := h.prometheusAPI.Query(ctx, "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m])", now)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 查询活跃用户数
	activeUsers, _, err := h.prometheusAPI.Query(ctx, "active_users", now)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"qps":          extractSampleValue(qps),
		"avg_latency":  extractSampleValue(latency),
		"error_rate":   extractSampleValue(errorRate),
		"active_users": extractSampleValue(activeUsers),
		"timestamp":    now,
	})
}

// GetAlerts 获取告警列表
func (h *MonitoringHandler) GetAlerts(c fiber.Ctx) error {
	// 从 Prometheus 查询告警
	// 这里简化实现，实际应从 Alertmanager 查询
	return c.JSON([]fiber.Map{
		{
			"level":       "warning",
			"title":       "High Latency",
			"description": "P95 latency exceeds 1s",
			"timestamp":   time.Now().Add(-5 * time.Minute),
		},
	})
}

// GetLogs 获取日志列表
func (h *MonitoringHandler) GetLogs(c fiber.Ctx) error {
	// 从 Loki 查询日志
	// 这里简化实现
	return c.JSON([]fiber.Map{
		{
			"timestamp":  time.Now().Add(-1 * time.Minute),
			"level":      "info",
			"message":    "Request processed successfully",
			"request_id": "req-123",
		},
	})
}

// GetMetricsOverview 获取指标概览
func (h *MonitoringHandler) GetMetricsOverview(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	now := time.Now()

	// 查询总请求数
	totalRequests, _, err := h.prometheusAPI.Query(ctx, "sum(http_requests_total)", now)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 查询总 Token 使用量
	totalTokens, _, err := h.prometheusAPI.Query(ctx, "sum(api_tokens_total)", now)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 查询总费用
	totalCost, _, err := h.prometheusAPI.Query(ctx, "sum(api_cost_total)", now)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 查询当前并发连接数
	concurrentConns, _, err := h.prometheusAPI.Query(ctx, "concurrent_connections", now)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"total_requests":     extractSampleValue(totalRequests),
		"total_tokens":       extractSampleValue(totalTokens),
		"total_cost":         extractSampleValue(totalCost),
		"concurrent_connections": extractSampleValue(concurrentConns),
		"timestamp":          now,
	})
}

// GetModelMetrics 获取模型级别指标
func (h *MonitoringHandler) GetModelMetrics(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	now := time.Now()

	// 查询按模型分组的请求数
	requestsByModel, _, err := h.prometheusAPI.Query(ctx, "sum by (model) (api_requests_total)", now)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 查询按模型分组的 Token 使用量
	tokensByModel, _, err := h.prometheusAPI.Query(ctx, "sum by (model) (api_tokens_total)", now)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"requests_by_model": extractVectorValues(requestsByModel),
		"tokens_by_model":   extractVectorValues(tokensByModel),
		"timestamp":         now,
	})
}

// extractSampleValue extracts the value from a Prometheus query result
func extractSampleValue(result interface{}) float64 {
	if vector, ok := result.(model.Vector); ok && len(vector) > 0 {
		return float64(vector[0].Value)
	}
	return 0
}

// extractVectorValues extracts all values from a Vector result
func extractVectorValues(result interface{}) map[string]float64 {
	values := make(map[string]float64)
	if vector, ok := result.(model.Vector); ok {
		for _, sample := range vector {
			// Extract model label from metric
			for labelName, labelValue := range sample.Metric {
				if string(labelName) == "model" {
					values[string(labelValue)] = float64(sample.Value)
				}
			}
		}
	}
	return values
}
