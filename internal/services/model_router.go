package services

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
)

// ModelRouter 模型路由器 (per D-09: 利润优化路由)
type ModelRouter struct {
	db                *database.DB
	membershipService *MembershipService
}

// NewModelRouter 创建模型路由器
func NewModelRouter(db *database.DB, ms *MembershipService) (*ModelRouter, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	router := &ModelRouter{
		db:                db,
		membershipService: ms,
	}

	log.Println("✅ 模型路由器初始化成功")
	return router, nil
}

// SupplierSelection 供应商选择结果
type SupplierSelection struct {
	SupplierID      string
	SupplierName    string
	ActualModelName string
	Priority        int
	CostPrice       float64
	SellingPrice    float64
	Profit          float64
	ProfitMargin    float64
}

// SelectBestSupplier 选择最优供应商 (per D-09)
func (mr *ModelRouter) SelectBestSupplier(ctx context.Context, userID, externalModelID string) (*SupplierSelection, error) {
	// 1. 获取用户售价 (会员价)
	var userPricing models.UserGroupPricing
	err := mr.db.WithContext(ctx).
		Where("model_id = ? AND is_active = ?", externalModelID, true).
		First(&userPricing).Error

	if err != nil {
		return nil, fmt.Errorf("no pricing found for model %s: %w", externalModelID, err)
	}

	// 应用会员折扣
	sellingPrice, err := mr.membershipService.CalculatePrice(
		ctx, userID, externalModelID, userPricing.InputPrice)
	if err != nil {
		sellingPrice = userPricing.InputPrice // 失败时使用原价
	}

	// 2. 获取所有可用的供应商路由
	routes, err := mr.getAvailableRoutes(ctx, externalModelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("no available routes for model %s", externalModelID)
	}

	// 3. 评估每个供应商并选择最优
	var bestSupplier *SupplierSelection
	bestScore := math.Inf(-1)

	for _, route := range routes {
		selection, err := mr.evaluateSupplier(ctx, route, sellingPrice)
		if err != nil {
			log.Printf("警告: 评估供应商 %s 失败: %v", route.SupplierID, err)
			continue
		}

		// 评分: 利润率 * 100 + (10 - 优先级)
		// 利润率越高越好，优先级越小越好
		score := selection.ProfitMargin*100 + float64(10-selection.Priority)

		if score > bestScore {
			bestScore = score
			bestSupplier = selection
		}
	}

	if bestSupplier == nil {
		return nil, fmt.Errorf("no suitable supplier found")
	}

	log.Printf("路由选择: 模型=%s 供应商=%s 利润=%.4f 利润率=%.2f%%",
		externalModelID, bestSupplier.SupplierName, bestSupplier.Profit, bestSupplier.ProfitMargin*100)

	return bestSupplier, nil
}

// GetAvailableRoutes 获取可用的路由
func (mr *ModelRouter) getAvailableRoutes(ctx context.Context, externalModelID string) ([]models.ModelMapping, error) {
	var routes []models.ModelMapping
	err := mr.db.WithContext(ctx).
		Where("external_model_id = ? AND is_active = ?", externalModelID, true).
		Order("priority ASC").
		Find(&routes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch routes: %w", err)
	}

	return routes, nil
}

// evaluateSupplier 评估供应商
func (mr *ModelRouter) evaluateSupplier(ctx context.Context, route models.ModelMapping, sellingPrice float64) (*SupplierSelection, error) {
	// 获取供应商成本价
	var costPricing models.SupplierCostPricing
	err := mr.db.WithContext(ctx).
		Where("supplier_id = ? AND model_id = ? AND is_active = ?",
			route.SupplierID, route.ExternalModelID, true).
		First(&costPricing).Error

	if err != nil {
		return nil, fmt.Errorf("no cost pricing found: %w", err)
	}

	costPrice := costPricing.InputCost

	// 计算利润
	profit := sellingPrice - costPrice
	profitMargin := 0.0
	if sellingPrice > 0 {
		profitMargin = profit / sellingPrice
	}

	// 获取供应商信息
	var supplier models.Supplier
	err = mr.db.WithContext(ctx).Where("id = ?", route.SupplierID).First(&supplier).Error
	if err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}

	return &SupplierSelection{
		SupplierID:      route.SupplierID,
		SupplierName:    supplier.DisplayName,
		ActualModelName: route.ActualModelName,
		Priority:        route.Priority,
		CostPrice:       costPrice,
		SellingPrice:    sellingPrice,
		Profit:          profit,
		ProfitMargin:    profitMargin,
	}, nil
}

// CalculateProfit 计算利润 (per D-09)
func (mr *ModelRouter) CalculateProfit(sellingPrice, costPrice float64) float64 {
	return sellingPrice - costPrice
}

// ValidateProfitMargin 验证利润率
func (mr *ModelRouter) ValidateProfitMargin(profit, sellingPrice, minMargin float64) error {
	if sellingPrice <= 0 {
		return fmt.Errorf("selling price must be positive")
	}

	profitMargin := profit / sellingPrice
	if profitMargin < minMargin {
		return fmt.Errorf("profit margin %.2f%% below minimum %.2f%%",
			profitMargin*100, minMargin*100)
	}

	return nil
}

// CheckSupplierHealth 检查供应商健康状态
func (mr *ModelRouter) CheckSupplierHealth(ctx context.Context, supplierID string) bool {
	// TODO: 实现健康检查逻辑
	// 1. 检查最近的错误率
	// 2. 检查响应时间
	// 3. 检查 QPS 限制
	return true
}

// IsOverloaded 检查供应商是否过载
func (mr *ModelRouter) IsOverloaded(ctx context.Context, route models.ModelMapping) bool {
	// TODO: 实现 QPS 检查
	if route.MaxQPS <= 0 {
		return false
	}
	// 检查当前 QPS 是否超过限制
	return false
}
