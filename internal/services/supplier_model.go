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

type SupplierModelService struct {
	db *database.DB
}

func NewSupplierModelService(db *database.DB) (*SupplierModelService, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return &SupplierModelService{db: db}, nil
}

func (s *SupplierModelService) AddModelToSupplier(ctx context.Context, supplierID, modelID string, inputCost, outputCost float64) error {
	// 检查是否已存在
	var existing models.SupplierModel
	err := s.db.WithContext(ctx).
		Where("supplier_id = ? AND model_id = ?", supplierID, modelID).
		First(&existing).Error
	
	now := time.Now().UTC()
	
	if err == nil {
		// 记录历史
		s.recordPriceHistory(ctx, &existing, inputCost, outputCost, "system")
		
		// 更新现有记录
		existing.InputCost = inputCost
		existing.OutputCost = outputCost
		existing.UpdatedAt = now
		return s.db.WithContext(ctx).Save(&existing).Error
	}
	
	// 创建新记录
	supplierModel := &models.SupplierModel{
		ID:            xid.New().String(),
		SupplierID:    supplierID,
		ModelID:       modelID,
		InputCost:     inputCost,
		OutputCost:    outputCost,
		IsActive:      true,
		EffectiveDate: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	
	if err := s.db.WithContext(ctx).Create(supplierModel).Error; err != nil {
		return fmt.Errorf("failed to add model to supplier: %w", err)
	}
	
	log.Printf("✅ 模型添加到供应商: 供应商=%s 模型=%s", supplierID, modelID)
	return nil
}

func (s *SupplierModelService) ListSupplierModels(ctx context.Context, supplierID string) ([]*models.SupplierModel, error) {
	var models []*models.SupplierModel
	err := s.db.WithContext(ctx).
		Where("supplier_id = ?", supplierID).
		Order("model_id ASC").
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to list supplier models: %w", err)
	}
	
	return models, nil
}

func (s *SupplierModelService) UpdateModelCost(ctx context.Context, id string, inputCost, outputCost float64) error {
	var existing models.SupplierModel
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&existing).Error
	if err != nil {
		return fmt.Errorf("failed to get supplier model: %w", err)
	}
	
	// 记录历史
	s.recordPriceHistory(ctx, &existing, inputCost, outputCost, "system")
	
	// 更新价格
	now := time.Now().UTC()
	result := s.db.WithContext(ctx).
		Model(&models.SupplierModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"input_cost":  inputCost,
			"output_cost": outputCost,
			"updated_at":  now,
		})
	
	if result.Error != nil {
		return fmt.Errorf("failed to update model cost: %w", result.Error)
	}
	
	log.Printf("✅ 模型价格更新: ID=%s 输入=%.4f 输出=%.4f", id, inputCost, outputCost)
	return nil
}

func (s *SupplierModelService) RemoveModelFromSupplier(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.SupplierModel{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to remove model from supplier: %w", result.Error)
	}
	
	log.Printf("✅ 模型从供应商移除: ID=%s", id)
	return nil
}

func (s *SupplierModelService) GetPriceHistory(ctx context.Context, supplierID, modelID string) ([]*models.PriceHistory, error) {
	var history []*models.PriceHistory
	err := s.db.WithContext(ctx).
		Where("supplier_id = ? AND model_id = ?", supplierID, modelID).
		Order("changed_at DESC").
		Limit(100).
		Find(&history).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}
	
	return history, nil
}

func (s *SupplierModelService) recordPriceHistory(ctx context.Context, existing *models.SupplierModel, newInputCost, newOutputCost float64, changedBy string) {
	history := &models.PriceHistory{
		ID:              xid.New().String(),
		SupplierModelID: existing.ID,
		SupplierID:      existing.SupplierID,
		ModelID:         existing.ModelID,
		OldInputCost:    existing.InputCost,
		OldOutputCost:   existing.OutputCost,
		NewInputCost:    newInputCost,
		NewOutputCost:   newOutputCost,
		ChangedAt:       time.Now().UTC(),
		ChangedBy:       changedBy,
	}

	if err := s.db.WithContext(ctx).Create(history).Error; err != nil {
		log.Printf("警告: 记录价格历史失败: %v", err)
	}
}

// GetModelCost 获取模型成本（用于审计日志）
func (s *SupplierModelService) GetModelCost(ctx context.Context, id string) (*models.SupplierModel, error) {
	var model models.SupplierModel
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier model: %w", err)
	}
	return &model, nil
}

// GetModelInfo 获取模型信息（用于审计日志）
func (s *SupplierModelService) GetModelInfo(ctx context.Context, id string) (*models.SupplierModel, error) {
	return s.GetModelCost(ctx, id)
}
