package supplier

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/yockii/ai-gateway/internal/models"
)

// HealthChecker 健康检查器
type HealthChecker struct {
	manager    *Manager
	httpClient *http.Client
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(manager *Manager) *HealthChecker {
	return &HealthChecker{
		manager:   manager,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// PerformRealHealthCheck 执行真实的健康检查
func (hc *HealthChecker) PerformRealHealthCheck(ctx context.Context, supplierID string, getAPIKey func(supplierID string) (string, error)) (*HealthCheckResult, error) {
	// 获取供应商信息
	supplier, err := hc.manager.GetSupplier(ctx, supplierID)
	if err != nil {
		return &HealthCheckResult{
			SupplierID: supplierID,
			IsHealthy:  false,
			Error:      fmt.Sprintf("supplier not found: %v", err),
		}, nil
	}

	// 检查供应商是否活跃
	if !supplier.IsActive {
		return &HealthCheckResult{
			SupplierID: supplierID,
			IsHealthy:  false,
			Error:      "supplier is not active",
		}, nil
	}

	// 获取 API Key（通过回调函数避免循环依赖）
	apiKeyValue, err := getAPIKey(supplierID)
	if err != nil {
		return &HealthCheckResult{
			SupplierID: supplierID,
			IsHealthy:  false,
			Error:      fmt.Sprintf("no available API key: %v", err),
		}, nil
	}

	// 根据供应商类型执行不同的检查
	startTime := time.Now()
	var checkErr error
	var statusCode int

	switch supplier.Provider {
	case "openai":
		checkErr, statusCode = hc.checkOpenAI(ctx, apiKeyValue)
	case "anthropic":
		checkErr, statusCode = hc.checkAnthropic(ctx, apiKeyValue)
	default:
		checkErr, statusCode = hc.checkGeneric(ctx, apiKeyValue)
	}

	latency := time.Since(startTime)

	// 评估结果
	isHealthy := checkErr == nil && statusCode < 500
	errorMsg := ""
	if checkErr != nil {
		errorMsg = checkErr.Error()
	}

	result := &HealthCheckResult{
		SupplierID: supplierID,
		IsHealthy:  isHealthy,
		Latency:    latency,
		Error:      errorMsg,
	}

	// 记录健康检查历史
	hc.recordHealthCheck(ctx, supplierID, isHealthy, latency.Milliseconds(), errorMsg)

	// 更新健康状态
	hc.manager.updateHealthStatus(result)

	if isHealthy {
		log.Printf("✅ 健康检查成功: %s 延迟=%v", supplierID, latency)
	} else {
		log.Printf("❌ 健康检查失败: %s 错误=%s 状态码=%d", supplierID, errorMsg, statusCode)
	}

	return result, nil
}

// checkOpenAI 检查 OpenAI 供应商（修复 WR-05: 确保有超时）
func (hc *HealthChecker) checkOpenAI(ctx context.Context, apiKey string) (error, int) {
	// 确保我们有超时
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	baseURL := "https://api.openai.com"

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/v1/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w"), 0
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w"), 0
	}
	defer resp.Body.Close()

	_, _ = io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API returned status %d", resp.StatusCode), resp.StatusCode
	}

	return nil, resp.StatusCode
}

// checkAnthropic 检查 Anthropic 供应商（修复 WR-05: 确保有超时）
func (hc *HealthChecker) checkAnthropic(ctx context.Context, apiKey string) (error, int) {
	// 确保我们有超时
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	baseURL := "https://api.anthropic.com"

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/v1/messages", nil)
	if err != nil {
		return err, 0
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w"), 0
	}
	defer resp.Body.Close()

	_, _ = io.ReadAll(resp.Body)

	if resp.StatusCode >= 500 {
		return fmt.Errorf("API returned status %d", resp.StatusCode), resp.StatusCode
	}

	if resp.StatusCode == 401 {
		return fmt.Errorf("authentication failed"), resp.StatusCode
	}

	return nil, resp.StatusCode
}

// checkGeneric 通用健康检查
func (hc *HealthChecker) checkGeneric(ctx context.Context, apiKey string) (error, int) {
	return fmt.Errorf("generic health check not implemented"), 0
}

// recordHealthCheck 记录健康检查历史
func (hc *HealthChecker) recordHealthCheck(ctx context.Context, supplierID string, isHealthy bool, latency int64, errorMsg string) {
	history := &models.HealthCheckHistory{
		ID:         generateID(),
		SupplierID: supplierID,
		IsHealthy:  isHealthy,
		Latency:    latency,
		Error:      errorMsg,
		Timestamp:  time.Now(),
	}

	go func() {
		if err := hc.manager.db.Create(history).Error; err != nil {
			log.Printf("警告: 记录健康检查历史失败: %v", err)
		}
	}()
}

// generateID 生成唯一 ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
