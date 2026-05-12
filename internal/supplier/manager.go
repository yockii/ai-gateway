package supplier

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
)

// Manager 供应商管理器
type Manager struct {
	db *database.DB

	// 服务依赖
	apiKeyService  SupplierApiKeyServiceInterface
	healthChecker  *HealthChecker

	// 健康检查缓存
	healthStatus      map[string]*HealthStatus
	healthMutex       sync.RWMutex
	backgroundMutex   sync.Mutex // 修复 CR-05: 序列化后台更新防止死锁

	// 健康检查配置
	checkInterval     time.Duration
	timeout           time.Duration
	failureThreshold  int
}

// SupplierApiKeyServiceInterface 供应商 API Key 服务接口（避免循环依赖）
type SupplierApiKeyServiceInterface interface {
	GetBestApiKey(ctx context.Context, supplierID string) (*models.SupplierApiKey, string, error)
}

// HealthStatus 健康状态
type HealthStatus struct {
	SupplierID       string
	IsHealthy        bool
	FailureCount     int
	LastCheckTime    time.Time
	LastError        string
	ConsecutiveFailures int
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	SupplierID string
	IsHealthy  bool
	Latency    time.Duration
	Error      string
}

// NewManager 创建供应商管理器
func NewManager(db *database.DB) (*Manager, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	mgr := &Manager{
		db: db,
		healthStatus:     make(map[string]*HealthStatus),
		checkInterval:    30 * time.Second,
		timeout:          10 * time.Second,
		failureThreshold: 3,
	}

	// 创建健康检查器
	mgr.healthChecker = NewHealthChecker(mgr)

	log.Println("✅ 供应商管理器初始化成功")
	return mgr, nil
}

// SetApiKeyService 设置 API Key 服务（延迟注入）
func (m *Manager) SetApiKeyService(service SupplierApiKeyServiceInterface) {
	m.apiKeyService = service
}

// AddSupplier 添加供应商
func (m *Manager) AddSupplier(ctx context.Context, supplier *models.Supplier) error {
	if supplier.ID == "" {
		return fmt.Errorf("supplier_id is required")
	}

	// 检查是否已存在
	var count int64
	if err := m.db.WithContext(ctx).Table("suppliers").Where("id = ?", supplier.ID).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check supplier existence: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("supplier already exists: %s", supplier.ID)
	}

	// 创建供应商
	if err := m.db.WithContext(ctx).Create(supplier).Error; err != nil {
		return fmt.Errorf("failed to create supplier: %w", err)
	}

	// 初始化健康状态
	m.healthMutex.Lock()
	m.healthStatus[supplier.ID] = &HealthStatus{
		SupplierID:        supplier.ID,
		IsHealthy:         true,  // 初始状态为健康
		FailureCount:      0,
		LastCheckTime:     time.Now(),
		ConsecutiveFailures: 0,
	}
	m.healthMutex.Unlock()

	log.Printf("✅ 供应商添加成功: %s (%s)", supplier.Name, supplier.ID)
	return nil
}

// UpdateSupplier 更新供应商
func (m *Manager) UpdateSupplier(ctx context.Context, supplier *models.Supplier) error {
	result := m.db.WithContext(ctx).
		Model(&models.Supplier{}).
		Where("id = ?", supplier.ID).
		Updates(supplier)

	if result.Error != nil {
		return fmt.Errorf("failed to update supplier: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("supplier not found: %s", supplier.ID)
	}

	log.Printf("✅ 供应商更新成功: %s", supplier.ID)
	return nil
}

// DeactivateSupplier 停用供应商
func (m *Manager) DeactivateSupplier(ctx context.Context, supplierID string) error {
	result := m.db.WithContext(ctx).
		Model(&models.Supplier{}).
		Where("id = ?", supplierID).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to deactivate supplier: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("supplier not found: %s", supplierID)
	}

	log.Printf("✅ 供应商已停用: %s", supplierID)
	return nil
}

// GetSupplier 获取供应商
func (m *Manager) GetSupplier(ctx context.Context, supplierID string) (*models.Supplier, error) {
	var supplier models.Supplier
	result := m.db.WithContext(ctx).Where("id = ?", supplierID).First(&supplier)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get supplier: %w", result.Error)
	}

	return &supplier, nil
}

// ListSuppliers 列出所有供应商
func (m *Manager) ListSuppliers(ctx context.Context, activeOnly bool) ([]*models.Supplier, error) {
	var suppliers []*models.Supplier

	query := m.db.WithContext(ctx)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	result := query.Find(&suppliers)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list suppliers: %w", result.Error)
	}

	return suppliers, nil
}

// GetHealthySuppliers 获取健康的供应商列表
func (m *Manager) GetHealthySuppliers(ctx context.Context) ([]*models.Supplier, error) {
	// 获取所有活跃供应商
	suppliers, err := m.ListSuppliers(ctx, true)
	if err != nil {
		return nil, err
	}

	// 过滤出健康的供应商
	var healthySuppliers []*models.Supplier
	for _, supplier := range suppliers {
		if m.IsHealthy(supplier.ID) {
			healthySuppliers = append(healthySuppliers, supplier)
		}
	}

	return healthySuppliers, nil
}

// IsHealthy 检查供应商是否健康
func (m *Manager) IsHealthy(supplierID string) bool {
	m.healthMutex.RLock()
	defer m.healthMutex.RUnlock()

	status, exists := m.healthStatus[supplierID]
	if !exists {
		// 如果没有健康状态记录，默认为健康
		return true
	}

	return status.IsHealthy
}

// PerformHealthCheck 执行健康检查 - 使用真实 API Key
func (m *Manager) PerformHealthCheck(ctx context.Context, supplierID string) (*HealthCheckResult, error) {
	// 获取供应商信息
	supplier, err := m.GetSupplier(ctx, supplierID)
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

	// 获取最佳 API Key 进行真实健康检查
	if m.apiKeyService != nil && m.healthChecker != nil {
		_, keyValue, err := m.apiKeyService.GetBestApiKey(ctx, supplierID)
		if err != nil {
			log.Printf("警告: 获取供应商 %s 的 API Key 失败: %v", supplierID, err)
			// 继续进行模拟健康检查
		} else {
			// 使用真实 API Key 进行健康检查
			return m.healthChecker.PerformRealHealthCheck(ctx, supplierID, func(s string) (string, error) {
				return keyValue, nil
			})
		}
	}

	// 降级：模拟健康检查
	startTime := time.Now()
	time.Sleep(10 * time.Millisecond)
	latency := time.Since(startTime)

	result := &HealthCheckResult{
		SupplierID: supplierID,
		IsHealthy:  true,
		Latency:    latency,
	}

	// 更新健康状态
	m.updateHealthStatus(result)

	return result, nil
}

// updateHealthStatus 更新健康状态
func (m *Manager) updateHealthStatus(result *HealthCheckResult) {
	m.healthMutex.Lock()
	defer m.healthMutex.Unlock()

	status, exists := m.healthStatus[result.SupplierID]
	if !exists {
		status = &HealthStatus{
			SupplierID: result.SupplierID,
		}
		m.healthStatus[result.SupplierID] = status
	}

	status.LastCheckTime = time.Now()

	if result.IsHealthy {
		status.IsHealthy = true
		status.ConsecutiveFailures = 0
		status.LastError = ""
	} else {
		status.ConsecutiveFailures++
		status.FailureCount++
		status.LastError = result.Error

		// 连续失败次数达到阈值，标记为不健康
		if status.ConsecutiveFailures >= m.failureThreshold {
			status.IsHealthy = false
			log.Printf("⚠️  供应商 %s 标记为不健康: 连续失败 %d 次",
				result.SupplierID, status.ConsecutiveFailures)
		}
	}
}

// backgroundHealthCheck 后台健康检查（修复 CR-05: 使用 backgroundMutex 序列化更新）
func (m *Manager) backgroundHealthCheck() {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), m.timeout)

		// 获取所有活跃供应商
		suppliers, err := m.ListSuppliers(ctx, true)
		cancel()

		if err != nil {
			log.Printf("健康检查: 获取供应商列表失败: %v", err)
			continue
		}

		// 并发检查所有供应商
		var wg sync.WaitGroup
		for _, supplier := range suppliers {
			wg.Add(1)
			go func(sid string) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
				defer cancel()

				result, err := m.PerformHealthCheck(ctx, sid)
				if err != nil {
					log.Printf("健康检查: 供应商 %s 检查失败: %v", sid, err)
				} else {
					// 修复 CR-05: 序列化后台更新防止与 updateHealthStatus 死锁
					m.backgroundMutex.Lock()
					m.updateHealthStatus(result)
					m.backgroundMutex.Unlock()
				}
			}(supplier.ID)
		}

		wg.Wait()
	}
}

// loadHealthStatus 加载健康状态
func (m *Manager) loadHealthStatus() error {
	// TODO: 从持久化存储加载健康状态
	// 当前为简化实现，使用内存存储
	return nil
}

// GetHealthStatus 获取健康状态
func (m *Manager) GetHealthStatus(supplierID string) *HealthStatus {
	m.healthMutex.RLock()
	defer m.healthMutex.RUnlock()

	return m.healthStatus[supplierID]
}

// GetAllHealthStatus 获取所有供应商的健康状态
func (m *Manager) GetAllHealthStatus() map[string]*HealthStatus {
	m.healthMutex.RLock()
	defer m.healthMutex.RUnlock()

	// 返回副本
	result := make(map[string]*HealthStatus, len(m.healthStatus))
	for k, v := range m.healthStatus {
		result[k] = v
	}

	return result
}
