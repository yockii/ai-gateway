package supplier

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/yockii/ai-gateway/internal/models"
)

// HealthChecker 健康检查器（修复 CR-06: worker pool 限制 goroutine）
type HealthChecker struct {
	manager    *Manager
	httpClient *http.Client
	recordChan chan *models.HealthCheckHistory // 修复 CR-06: 使用通道代替直接创建 goroutine
	workers    int
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(manager *Manager) *HealthChecker {
	hc := &HealthChecker{
		manager:    manager,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		recordChan: make(chan *models.HealthCheckHistory, 1000), // 缓冲通道
		workers:    5,  // 固定数量的 worker
		stopChan:   make(chan struct{}),
	}

	// 启动 worker pool
	for i := 0; i < hc.workers; i++ {
		hc.wg.Add(1)
		go hc.recordWorker(i)
	}

	return hc
}

// recordWorker worker pool 工作协程（修复 CR-06）
func (hc *HealthChecker) recordWorker(id int) {
	defer hc.wg.Done()
	for {
		select {
		case history, ok := <-hc.recordChan:
			if !ok {
				// 通道已关闭
				return
			}
			if err := hc.manager.db.Create(history).Error; err != nil {
				log.Printf("警告: [worker-%d] 记录健康检查历史失败: %v", id, err)
			}
		case <-hc.stopChan:
			return
		}
	}
}

// Stop 停止 health checker
func (hc *HealthChecker) Stop() {
	close(hc.stopChan)
	close(hc.recordChan)
	hc.wg.Wait()
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

	// 记录健康检查历史（修复 CR-06: 使用通道代替直接创建 goroutine）
	hc.recordHealthCheck(supplierID, isHealthy, latency.Milliseconds(), errorMsg)

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

// recordHealthCheck 记录健康检查历史（修复 CR-06: 使用通道发送到 worker pool）
func (hc *HealthChecker) recordHealthCheck(supplierID string, isHealthy bool, latency int64, errorMsg string) {
	history := &models.HealthCheckHistory{
		ID:         generateID(),
		SupplierID: supplierID,
		IsHealthy:  isHealthy,
		Latency:    latency,
		Error:      errorMsg,
		Timestamp:  time.Now(),
	}

	select {
	case hc.recordChan <- history:
		// 成功发送到通道
	default:
		// 通道已满，丢弃记录
		log.Printf("警告: 健康检查记录通道已满，丢弃记录: 供应商=%s", supplierID)
	}
}

// generateID 生成唯一 ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
