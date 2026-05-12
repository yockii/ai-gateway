package pricing_test

import (
	"context"
	"testing"
	"time"

	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/yockii/ai-gateway/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestEnterprisePricingService(t *testing.T) {
	// 初始化测试数据库
	db, err := database.New("host=localhost port=5432 user=llm_gateway password=llm_gateway dbname=llm_gateway sslmode=disable")
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
		return
	}

	// 创建服务
	enterpriseService, err := services.NewEnterprisePricingService(db)
	assert.NoError(t, err)

	membershipService, err := services.NewMembershipService(db)
	assert.NoError(t, err)

	pricingService, err := services.NewPricingService(db, membershipService, enterpriseService)
	assert.NoError(t, err)

	ctx := context.Background()

	// 测试 1: 创建大客户定价
	t.Run("CreateEnterprisePricing", func(t *testing.T) {
		pricing := &models.EnterprisePricing{
			CustomerID:      "test-customer-001",
			CustomerName:    "测试企业",
			ModelID:         "gpt-4",
			InputPrice:      12.0,
			OutputPrice:     24.0,
			MinProfitMargin: 0.1,
			MaxCostPrice:    10.0,
			EffectiveDate:   time.Now().UTC(),
			IsActive:        true,
		}

		err := enterpriseService.SetEnterprisePrice(ctx, pricing, "admin-001")
		assert.NoError(t, err)
		assert.NotEmpty(t, pricing.ID)
	})

	// 测试 2: 查询大客户定价
	t.Run("GetEnterprisePrice", func(t *testing.T) {
		pricing, err := enterpriseService.GetEnterprisePrice(ctx, "test-customer-001", "gpt-4")
		assert.NoError(t, err)
		assert.Equal(t, "测试企业", pricing.CustomerName)
		assert.Equal(t, 12.0, pricing.InputPrice)
		assert.Equal(t, 24.0, pricing.OutputPrice)
	})

	// 测试 3: 价格优先级
	t.Run("PricePriority", func(t *testing.T) {
		// 创建基础用户组和定价
		basicGroup := &models.UserGroup{
			ID:          "basic-test",
			Name:        "basic",
			DisplayName: "普通用户",
			IsActive:    true,
		}
		db.WithContext(ctx).Create(basicGroup)

		basicPricing := &models.UserGroupPricing{
			UserGroupID:     "basic-test",
			ModelID:          "gpt-4",
			InputPrice:       18.0,
			OutputPrice:      36.0,
			MinProfitMargin:  0.1,
			EffectiveDate:   time.Now().UTC(),
			IsActive:        true,
		}
		db.WithContext(ctx).Create(basicPricing)

		// 大客户定价应该优先
		result, err := pricingService.GetUserPrice(ctx, "test-customer-001", "gpt-4")
		assert.NoError(t, err)
		assert.Equal(t, "enterprise", result.AppliedRule)
		assert.Equal(t, 12.0, result.InputPrice)
	})

	// 测试 4: 利润率验证
	t.Run("ProfitMarginValidation", func(t *testing.T) {
		// 创建供应商成本价
		costPricing := &models.SupplierCostPricing{
			SupplierID:    "test-supplier",
			ModelID:       "gpt-4",
			InputCost:     10.0,
			OutputCost:    20.0,
			EffectiveDate: time.Now().UTC(),
			IsActive:      true,
		}
		db.WithContext(ctx).Create(costPricing)

		// 计算带利润的价格
		result, err := pricingService.CalculatePriceWithProfit(
			ctx, "test-customer-001", "gpt-4", "test-supplier")
		assert.NoError(t, err)
		assert.True(t, result.IsValid)
		assert.Equal(t, 2.0, result.Profit) // 12 - 10
		assert.True(t, result.ProfitMargin >= 0.1)
	})

	// 测试 5: 利润率不足应该失败
	t.Run("InsufficientProfitMargin", func(t *testing.T) {
		// 设置一个很高的成本价
		highCostPricing := &models.SupplierCostPricing{
			SupplierID:    "test-supplier-2",
			ModelID:       "gpt-4",
			InputCost:     15.0, // 高于售价 12
			OutputCost:    30.0,
			EffectiveDate: time.Now().UTC(),
			IsActive:      true,
		}
		db.WithContext(ctx).Create(highCostPricing)

		result, err := pricingService.CalculatePriceWithProfit(
			ctx, "test-customer-001", "gpt-4", "test-supplier-2")
		assert.NoError(t, err)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.ValidationErrors)
	})

	// 清理测试数据
	db.WithContext(ctx).Where("customer_id = ?", "test-customer-001").Delete(&models.EnterprisePricing{})
	db.WithContext(ctx).Where("id = ?", "basic-test").Delete(&models.UserGroup{})
	db.WithContext(ctx).Where("supplier_id LIKE ?", "test-supplier%").Delete(&models.SupplierCostPricing{})
}
