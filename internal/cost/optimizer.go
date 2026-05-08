package cost

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/yockii/ai-gateway/internal/config"
	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
)

// Optimizer 成本优化器
type Optimizer struct {
	db    *database.DB
	config config.CostConfig

	// 性能监控缓存
	performanceCache map[string]*SupplierPerformance
	cacheMutex       sync.RWMutex
}

// SupplierPerformance 供应商性能指标
type SupplierPerformance struct {
	SupplierID      string
	AvgResponseTime float64 // 平均响应时间（毫秒）
	SuccessRate     float64 // 成功率 (0-1)
	RequestCount    int64   // 请求总数
	LastUpdateTime  int64   // 最后更新时间（Unix timestamp）
}

// SupplierSelection 供应商选择结果
type SupplierSelection struct {
	Supplier       *models.Supplier
	CostPrice      float64
	SellingPrice   float64
	Profit         float64
	ProfitMargin   float64
	ActualModelName string
}

// NewOptimizer 创建成本优化器
func NewOptimizer(db *database.DB, cfg config.CostConfig) (*Optimizer, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	optimizer := &Optimizer{
		db:               db,
		config:           cfg,
		performanceCache: make(map[string]*SupplierPerformance),
	}

	// 初始化性能监控缓存
	if err := optimizer.loadPerformanceMetrics(); err != nil {
		log.Printf("警告: 加载性能指标失败: %v", err)
		// 不影响启动，继续运行
	}

	log.Println("✅ 成本优化器初始化成功")
	return optimizer, nil
}

// SelectBestSupplier 选择最优供应商
// 根据成本、性能、可用性综合评分选择
func (o *Optimizer) SelectBestSupplier(ctx context.Context, userID, modelID string) (*SupplierSelection, error) {
	// 1. 获取用户所属群体
	userGroupID, err := o.getUserGroupID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user group: %w", err)
	}

	// 2. 获取所有可用供应商
	suppliers, err := o.getAvailableSuppliers(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get suppliers: %w", err)
	}

	if len(suppliers) == 0 {
		return nil, fmt.Errorf("no available suppliers for model: %s", modelID)
	}

	// 3. 评估每个供应商并选择最优
	var bestSupplier *SupplierSelection
	bestScore := math.Inf(-1)

	for _, supplier := range suppliers {
		selection, err := o.evaluateSupplier(ctx, supplier, userGroupID, modelID)
		if err != nil {
			log.Printf("警告: 评估供应商 %s 失败: %v", supplier.Name, err)
			continue
		}

		score := o.calculateScore(selection)
		if score > bestScore {
			bestScore = score
			bestSupplier = selection
		}
	}

	if bestSupplier == nil {
		return nil, fmt.Errorf("no suitable supplier found")
	}

	log.Printf("选择供应商: %s (成本: %.4f, 售价: %.4f, 利润率: %.2f%%, 评分: %.2f)",
		bestSupplier.Supplier.Name,
		bestSupplier.CostPrice,
		bestSupplier.SellingPrice,
		bestSupplier.ProfitMargin*100,
		bestScore)

	return bestSupplier, nil
}

// SelectBackupSupplier 选择备用供应商（故障转移）
func (o *Optimizer) SelectBackupSupplier(ctx context.Context, userID, modelID string, primarySupplierID string) *SupplierSelection {
	suppliers, _ := o.getAvailableSuppliers(ctx, modelID)

	for _, supplier := range suppliers {
		// 跳过主要供应商
		if supplier.ID == primarySupplierID {
			continue
		}

		userGroupID, _ := o.getUserGroupID(ctx, userID)
		selection, err := o.evaluateSupplier(ctx, supplier, userGroupID, modelID)
		if err != nil {
			continue
		}

		log.Printf("选择备用供应商: %s", supplier.Name)
		return selection
	}

	return nil
}

// evaluateSupplier 评估供应商
func (o *Optimizer) evaluateSupplier(ctx context.Context, supplier *models.Supplier, userGroupID, modelID string) (*SupplierSelection, error) {
	// 获取供应商成本定价
	costPricing, err := o.getSupplierCostPricing(ctx, supplier.ID, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost pricing: %w", err)
	}

	// 获取用户群体定价
	groupPricing, err := o.getUserGroupPricing(ctx, userGroupID, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group pricing: %w", err)
	}

	// 计算成本和售价
	costPrice := costPricing.InputCost // 简化计算，使用输入成本
	sellingPrice := groupPricing.InputPrice

	// 验证利润率
	profit := sellingPrice - costPrice
	profitMargin := profit / sellingPrice

	if profitMargin < o.config.MinProfitMargin {
		return nil, fmt.Errorf("profit margin %.2f%% below minimum %.2f%%",
			profitMargin*100, o.config.MinProfitMargin*100)
	}

	return &SupplierSelection{
		Supplier:        supplier,
		CostPrice:       costPrice,
		SellingPrice:    sellingPrice,
		Profit:          profit,
		ProfitMargin:    profitMargin,
		ActualModelName: modelID, // 简化处理
	}, nil
}

// calculateScore 计算供应商综合评分
// 评分考虑：成本、性能、可用性
func (o *Optimizer) calculateScore(selection *SupplierSelection) float64 {
	// 基础分数：利润率（越高越好）
	score := selection.ProfitMargin * 100

	// 性能加分：根据历史性能数据
	o.cacheMutex.RLock()
	perf, exists := o.performanceCache[selection.Supplier.ID]
	o.cacheMutex.RUnlock()

	if exists {
		// 成功率加分 (0-30 分)
		score += perf.SuccessRate * 30

		// 响应时间加分 (0-20 分，响应越快分数越高)
		// 假设 1000ms 为基准，每快 100ms 加 2 分
		responseBonus := math.Max(0, (1000-perf.AvgResponseTime)/100*2)
		score += math.Min(responseBonus, 20)
	}

	return score
}

// getUserGroupID 获取用户所属群体
func (o *Optimizer) getUserGroupID(ctx context.Context, userID string) (string, error) {
	var user models.User
	result := o.db.WithContext(ctx).Where("id = ? AND is_active = ?", userID, true).First(&user)
	if result.Error != nil {
		return "", result.Error
	}
	return user.UserGroupID, nil
}

// getAvailableSuppliers 获取可用供应商列表
func (o *Optimizer) getAvailableSuppliers(ctx context.Context, modelID string) ([]*models.Supplier, error) {
	var suppliers []*models.Supplier

	// 查询活跃的且有成本定价的供应商
	subQuery := o.db.WithContext(ctx).
		Table("supplier_cost_pricings").
		Select("supplier_id").
		Where("model_id = ? AND is_active = ?", modelID, true)

	result := o.db.WithContext(ctx).
		Where("is_active = ? AND id IN (?)", true, subQuery).
		Find(&suppliers)

	if result.Error != nil {
		return nil, result.Error
	}

	return suppliers, nil
}

// getSupplierCostPricing 获取供应商成本定价
func (o *Optimizer) getSupplierCostPricing(ctx context.Context, supplierID, modelID string) (*models.SupplierCostPricing, error) {
	var pricing models.SupplierCostPricing
	result := o.db.WithContext(ctx).
		Where("supplier_id = ? AND model_id = ? AND is_active = ?", supplierID, modelID, true).
		First(&pricing)

	if result.Error != nil {
		return nil, result.Error
	}

	return &pricing, nil
}

// getUserGroupPricing 获取用户群体定价
func (o *Optimizer) getUserGroupPricing(ctx context.Context, userGroupID, modelID string) (*models.UserGroupPricing, error) {
	var pricing models.UserGroupPricing
	result := o.db.WithContext(ctx).
		Where("user_group_id = ? AND model_id = ? AND is_active = ?", userGroupID, modelID, true).
		First(&pricing)

	if result.Error != nil {
		return nil, result.Error
	}

	return &pricing, nil
}

// loadPerformanceMetrics 加载性能指标
func (o *Optimizer) loadPerformanceMetrics() error {
	// TODO: 从 usage_records 表加载历史性能数据
	// 当前为简化实现，使用默认值
	return nil
}

// UpdatePerformanceMetric 更新供应商性能指标
func (o *Optimizer) UpdatePerformanceMetric(supplierID string, responseTime float64, success bool) {
	o.cacheMutex.Lock()
	defer o.cacheMutex.Unlock()

	perf, exists := o.performanceCache[supplierID]
	if !exists {
		perf = &SupplierPerformance{
			SupplierID:     supplierID,
			SuccessRate:    1.0,
			RequestCount:   0,
			LastUpdateTime: 0,
		}
		o.performanceCache[supplierID] = perf
	}

	// 更新性能指标（简化算法）
	perf.RequestCount++

	// 指数移动平均 (EMA)
	alpha := 0.2
	perf.AvgResponseTime = alpha*responseTime + (1-alpha)*perf.AvgResponseTime

	// 更新成功率
	if success {
		perf.SuccessRate = alpha*1.0 + (1-alpha)*perf.SuccessRate
	} else {
		perf.SuccessRate = alpha*0.0 + (1-alpha)*perf.SuccessRate
	}
}
