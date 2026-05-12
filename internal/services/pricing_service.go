package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
)

type PricingService struct {
	db                *database.DB
	membershipService *MembershipService
	enterpriseService *EnterprisePricingService
}

func NewPricingService(db *database.DB, ms *MembershipService, es *EnterprisePricingService) (*PricingService, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	log.Println("✅ 统一定价服务初始化成功")
	return &PricingService{db: db, membershipService: ms, enterpriseService: es}, nil
}

type UserPricingResult struct {
	UserID          string    `json:"user_id"`
	ModelID         string    `json:"model_id"`
	InputPrice      float64   `json:"input_price"`
	OutputPrice     float64   `json:"output_price"`
	AppliedRule     string    `json:"applied_rule"`
	DiscountRate    float64   `json:"discount_rate"`
	EffectiveDate   time.Time `json:"effective_date"`
	MinProfitMargin float64   `json:"min_profit_margin"`
}

type PriceCalculationResult struct {
	CostPrice        float64  `json:"cost_price"`
	SellingPrice     float64  `json:"selling_price"`
	Profit           float64  `json:"profit"`
	ProfitMargin     float64  `json:"profit_margin"`
	IsValid          bool     `json:"is_valid"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
	AppliedRule      string   `json:"applied_rule"`
}

func (s *PricingService) GetUserPrice(ctx context.Context, userID, modelID string) (*UserPricingResult, error) {
	now := time.Now()
	
	pricing, err := s.enterpriseService.GetEnterprisePrice(ctx, userID, modelID)
	if err == nil {
		log.Printf("定价应用: 用户=%s 模型=%s 规则=enterprise", userID, modelID)
		return &UserPricingResult{UserID: userID, ModelID: modelID, InputPrice: pricing.InputPrice,
			OutputPrice: pricing.OutputPrice, AppliedRule: "enterprise", EffectiveDate: pricing.EffectiveDate,
			MinProfitMargin: pricing.MinProfitMargin}, nil
	}
	
	membership, err := s.membershipService.GetUserMembership(ctx, userID)
	if err == nil {
		discount, _ := s.membershipService.GetModelDiscount(ctx, userID, modelID)
		if discount > 0 {
			basePrice, err := s.getBasePrice(ctx, modelID)
			if err == nil {
				finalPrice := s.membershipService.ApplyDiscount(basePrice.InputPrice, discount)
				return &UserPricingResult{UserID: userID, ModelID: modelID, InputPrice: finalPrice,
					OutputPrice: finalPrice * 2, AppliedRule: "membership", DiscountRate: discount,
					EffectiveDate: membership.EffectiveAt, MinProfitMargin: 0.1}, nil
			}
		}
	}
	
	var user models.User
	err = s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err == nil && user.UserGroupID != "" {
		var groupPricing models.UserGroupPricing
		err = s.db.WithContext(ctx).Where("user_group_id = ? AND model_id = ? AND is_active = ?",
			user.UserGroupID, modelID, true).First(&groupPricing).Error
		if err == nil {
			return &UserPricingResult{UserID: userID, ModelID: modelID, InputPrice: groupPricing.InputPrice,
				OutputPrice: groupPricing.OutputPrice, AppliedRule: "group",
				EffectiveDate: groupPricing.EffectiveDate, MinProfitMargin: groupPricing.MinProfitMargin}, nil
		}
	}
	
	basePrice, err := s.getBasePrice(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("no pricing available: %w", err)
	}
	
	return &UserPricingResult{UserID: userID, ModelID: modelID, InputPrice: basePrice.InputPrice,
		OutputPrice: basePrice.OutputPrice, AppliedRule: "basic", EffectiveDate: now, MinProfitMargin: 0.1}, nil
}

func (s *PricingService) CalculatePriceWithProfit(ctx context.Context, userID, modelID, supplierID string) (*PriceCalculationResult, error) {
	userPricing, err := s.GetUserPrice(ctx, userID, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user pricing: %w", err)
	}
	
	var costPricing models.SupplierCostPricing
	err = s.db.WithContext(ctx).Where("supplier_id = ? AND model_id = ? AND is_active = ?", supplierID, modelID, true).First(&costPricing).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier cost pricing: %w", err)
	}
	
	sellingPrice := userPricing.InputPrice
	costPrice := costPricing.InputCost
	profit := sellingPrice - costPrice
	profitMargin := 0.0
	if sellingPrice > 0 {
		profitMargin = profit / sellingPrice
	}
	
	result := &PriceCalculationResult{CostPrice: costPrice, SellingPrice: sellingPrice, Profit: profit,
		ProfitMargin: profitMargin, AppliedRule: userPricing.AppliedRule}
	
	if profitMargin < userPricing.MinProfitMargin {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors,
			fmt.Sprintf("利润率 %.2f%% 低于最低要求 %.2f%%", profitMargin*100, userPricing.MinProfitMargin*100))
	} else {
		result.IsValid = true
	}
	
	return result, nil
}

func (s *PricingService) getBasePrice(ctx context.Context, modelID string) (*models.UserGroupPricing, error) {
	var basicGroup models.UserGroup
	err := s.db.WithContext(ctx).Where("name = ?", "basic").First(&basicGroup).Error
	if err != nil {
		return nil, fmt.Errorf("basic user group not found: %w", err)
	}
	
	var pricing models.UserGroupPricing
	err = s.db.WithContext(ctx).Where("user_group_id = ? AND model_id = ? AND is_active = ?", basicGroup.ID, modelID, true).First(&pricing).Error
	if err != nil {
		return nil, fmt.Errorf("basic pricing not found: %w", err)
	}
	
	return &pricing, nil
}

func (s *PricingService) ValidatePriceWithProfit(ctx context.Context, userID, modelID, supplierID string, sellingPrice float64) (*PriceCalculationResult, error) {
	var costPricing models.SupplierCostPricing
	err := s.db.WithContext(ctx).Where("supplier_id = ? AND model_id = ? AND is_active = ?", supplierID, modelID, true).First(&costPricing).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier cost pricing: %w", err)
	}
	
	costPrice := costPricing.InputCost
	profit := sellingPrice - costPrice
	profitMargin := 0.0
	if sellingPrice > 0 {
		profitMargin = profit / sellingPrice
	}
	
	userPricing, err := s.GetUserPrice(ctx, userID, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user pricing: %w", err)
	}
	
	result := &PriceCalculationResult{CostPrice: costPrice, SellingPrice: sellingPrice, Profit: profit,
		ProfitMargin: profitMargin, AppliedRule: userPricing.AppliedRule}
	
	if profitMargin < userPricing.MinProfitMargin {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors,
			fmt.Sprintf("利润率 %.2f%% 低于最低要求 %.2f%%", profitMargin*100, userPricing.MinProfitMargin*100))
	} else {
		result.IsValid = true
	}
	
	return result, nil
}

type PriceApplicationRecord struct {
	ID           string    `json:"id"`
	RequestID    string    `json:"request_id"`
	UserID       string    `json:"user_id"`
	ModelID      string    `json:"model_id"`
	SupplierID   string    `json:"supplier_id"`
	AppliedRule  string    `json:"applied_rule"`
	CostPrice    float64   `json:"cost_price"`
	SellingPrice float64   `json:"selling_price"`
	Profit       float64   `json:"profit"`
	ProfitMargin float64   `json:"profit_margin"`
	CreatedAt    time.Time `json:"created_at"`
}

func (s *PricingService) RecordPriceApplication(ctx context.Context, record *PriceApplicationRecord) error {
	log.Printf("价格应用: 请求=%s 用户=%s 模型=%s 规则=%s 成本=%.4f 售价=%.4f 利润=%.4f 利润率=%.2f%%",
		record.RequestID, record.UserID, record.ModelID, record.AppliedRule,
		record.CostPrice, record.SellingPrice, record.Profit, record.ProfitMargin*100)
	return nil
}
