package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rs/xid"
	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
)

// EnterprisePricingService 大客户定价服务 (Phase 5)
type EnterprisePricingService struct {
	db *database.DB
}

// NewEnterprisePricingService 创建大客户定价服务
func NewEnterprisePricingService(db *database.DB) (*EnterprisePricingService, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	service := &EnterprisePricingService{db: db}
	log.Println("✅ 大客户定价服务初始化成功")
	return service, nil
}

// GetEnterprisePrice 获取大客户价格
func (s *EnterprisePricingService) GetEnterprisePrice(
	ctx context.Context,
	customerID, modelID string,
) (*models.EnterprisePricing, error) {
	var pricing models.EnterprisePricing
	err := s.db.WithContext(ctx).
		Where("customer_id = ? AND model_id = ? AND is_active = ?", customerID, modelID, true).
		First(&pricing).Error

	if err != nil {
		return nil, fmt.Errorf("enterprise pricing not found for customer=%s model=%s: %w", customerID, modelID, err)
	}

	// 检查是否在有效期内
	if !pricing.IsActiveAt(time.Now()) {
		return nil, fmt.Errorf("enterprise pricing is not currently active")
	}

	return &pricing, nil
}

// SetEnterprisePrice 设置大客户价格
func (s *EnterprisePricingService) SetEnterprisePrice(
	ctx context.Context,
	pricing *models.EnterprisePricing,
	adminID string,
) error {
	// 验证利润率
	if pricing.MinProfitMargin <= 0 || pricing.MinProfitMargin >= 1 {
		return fmt.Errorf("invalid min_profit_margin: must be between 0 and 1")
	}

	// 检查是否已存在
	var existing models.EnterprisePricing
	err := s.db.WithContext(ctx).
		Where("customer_id = ? AND model_id = ?", pricing.CustomerID, pricing.ModelID).
		First(&existing).Error

	now := time.Now().UTC()

	if err == nil {
		// 更新现有记录
		existing.InputPrice = pricing.InputPrice
		existing.OutputPrice = pricing.OutputPrice
		existing.MinProfitMargin = pricing.MinProfitMargin
		existing.MaxCostPrice = pricing.MaxCostPrice
		existing.EffectiveDate = pricing.EffectiveDate
		existing.ExpiryDate = pricing.ExpiryDate
		existing.IsActive = pricing.IsActive
		existing.UpdatedAt = now
		existing.UpdatedBy = adminID

		if err := s.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return fmt.Errorf("failed to update enterprise pricing: %w", err)
		}

		log.Printf("✅ 大客户定价更新成功: 客户=%s 模型=%s", pricing.CustomerID, pricing.ModelID)
		return nil
	}

	// 创建新记录
	pricing.ID = xid.New().String()
	pricing.CreatedAt = now
	pricing.UpdatedAt = now
	pricing.CreatedBy = adminID
	pricing.UpdatedBy = adminID

	if err := s.db.WithContext(ctx).Create(pricing).Error; err != nil {
		return fmt.Errorf("failed to create enterprise pricing: %w", err)
	}

	log.Printf("✅ 大客户定价创建成功: 客户=%s 模型=%s 输入价格=%.4f 输出价格=%.4f",
		pricing.CustomerID, pricing.ModelID, pricing.InputPrice, pricing.OutputPrice)
	return nil
}

// ListEnterprisePricing 列出大客户定价
func (s *EnterprisePricingService) ListEnterprisePricing(
	ctx context.Context,
	filter *models.EnterprisePricingFilter,
) ([]*models.EnterprisePricing, error) {
	query := s.db.WithContext(ctx).Model(&models.EnterprisePricing{})

	// 应用筛选条件
	if filter != nil {
		if filter.CustomerID != "" {
			query = query.Where("customer_id = ?", filter.CustomerID)
		}
		if filter.ModelID != "" {
			query = query.Where("model_id = ?", filter.ModelID)
		}
		if filter.IsActive != nil {
			query = query.Where("is_active = ?", *filter.IsActive)
		}
		if filter.EffectiveAfter != nil {
			query = query.Where("effective_date >= ?", *filter.EffectiveAfter)
		}
		if filter.EffectiveBefore != nil {
			query = query.Where("effective_date <= ?", *filter.EffectiveBefore)
		}
	}

	// 分页
	if filter != nil && filter.Limit > 0 {
		query = query.Limit(filter.Limit)
		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}
	}

	var pricings []*models.EnterprisePricing
	err := query.Order("created_at DESC").Find(&pricings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list enterprise pricing: %w", err)
	}

	return pricings, nil
}

// UpdateEnterprisePrice 更新大客户价格
func (s *EnterprisePricingService) UpdateEnterprisePrice(
	ctx context.Context,
	id string,
	pricing *models.EnterprisePricing,
	adminID string,
) error {
	var existing models.EnterprisePricing
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&existing).Error
	if err != nil {
		return fmt.Errorf("enterprise pricing not found: %w", err)
	}

	// 验证利润率
	if pricing.MinProfitMargin <= 0 || pricing.MinProfitMargin >= 1 {
		return fmt.Errorf("invalid min_profit_margin: must be between 0 and 1")
	}

	// 更新字段
	existing.InputPrice = pricing.InputPrice
	existing.OutputPrice = pricing.OutputPrice
	existing.MinProfitMargin = pricing.MinProfitMargin
	existing.MaxCostPrice = pricing.MaxCostPrice
	existing.EffectiveDate = pricing.EffectiveDate
	existing.ExpiryDate = pricing.ExpiryDate
	existing.IsActive = pricing.IsActive
	existing.UpdatedAt = time.Now().UTC()
	existing.UpdatedBy = adminID

	if err := s.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return fmt.Errorf("failed to update enterprise pricing: %w", err)
	}

	log.Printf("✅ 大客户定价更新成功: ID=%s", id)
	return nil
}

// DeleteEnterprisePrice 删除大客户价格
func (s *EnterprisePricingService) DeleteEnterprisePrice(
	ctx context.Context,
	id string,
) error {
	result := s.db.WithContext(ctx).Delete(&models.EnterprisePricing{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete enterprise pricing: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("enterprise pricing not found: %s", id)
	}

	log.Printf("✅ 大客户定价删除成功: ID=%s", id)
	return nil
}

// ValidateProfitMargin 验证利润率
func (s *EnterprisePricingService) ValidateProfitMargin(
	sellingPrice, costPrice, minMargin float64,
) error {
	if sellingPrice <= 0 {
		return fmt.Errorf("selling price must be positive")
	}

	profitMargin := (sellingPrice - costPrice) / sellingPrice
	if profitMargin < minMargin {
		return fmt.Errorf("profit margin %.2f%% below minimum %.2f%% (cost: %.4f, selling: %.4f)",
			profitMargin*100, minMargin*100, costPrice, sellingPrice)
	}

	return nil
}

// CheckPriceEligibility 检查用户是否有大客户定价
func (s *EnterprisePricingService) CheckPriceEligibility(
	ctx context.Context,
	customerID string,
) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&models.EnterprisePricing{}).
		Where("customer_id = ? AND is_active = ?", customerID, true).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check pricing eligibility: %w", err)
	}

	return count > 0, nil
}
