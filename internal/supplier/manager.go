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

	// 健康检查缓存
	healthStatus      map[string]*HealthStatus
	healthMutex       sync.RWMutex

	// 健康检查配置
	checkInterval     time.Duration
	timeout           time.Duration
	failureThreshold  int
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

	manager := &Manager{
		db:               db,
		healthStatus:     make(map[string]*HealthStatus),
		checkInterval:    30 * time.Second,  // 默认 30 秒检查一次
		timeout:          10 * time.Second,  // 默认 10 秒超时
		failureThreshold: 3,                 // 默认连续失败 3 次标记为不健康
	}

	// 加载现有供应商的健康状态
	if err := manager.loadHealthStatus(); err != nil {
		log.Printf("警告: 加载健康状态失败: %v", err)
	}

	// 启动后台健康检查
	go manager.backgroundHealthCheck()

	log.Println("✅ 供应商管理器初始化成功")
	return manager, nil
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
	var healthySuppliers []*models.Suppliers
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

// PerformHealthCheck 执行健康检查
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

	// TODO: 实际的健康检查逻辑
	// 这里应该调用供应商的 API 进行健康检查
	// 当前为简化实现，假设所有供应商都是健康的

	startTime := time.Now()

	// 模拟健康检查延迟
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

// backgroundHealthCheck 后台健康检查
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

				_, err := m.PerformHealthCheck(ctx, sid)
				if err != nil {
					log.Printf("健康检查: 供应商 %s 检查失败: %v", sid, err)
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
