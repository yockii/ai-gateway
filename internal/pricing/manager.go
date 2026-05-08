package pricing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/rs/xid"
)

// Manager 定价管理器
type Manager struct {
	db *database.DB
}

// NewManager 创建定价管理器
func NewManager(db *database.DB) (*Manager, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	manager := &Manager{
		db: db,
	}

	log.Println("✅ 定价管理器初始化成功")
	return manager, nil
}

// RecordUsage 记录使用量
func (m *Manager) RecordUsage(ctx context.Context, usage *models.UsageRecord) error {
	if usage.ID == "" {
		usage.ID = xid.New().String()
	}

	if usage.CreatedAt.IsZero() {
		usage.CreatedAt = time.Now().UTC()
	}

	// 验证数据完整性
	if err := m.validateUsageRecord(usage); err != nil {
		return fmt.Errorf("invalid usage record: %w", err)
	}

	// 保存到数据库
	result := m.db.WithContext(ctx).Create(usage)
	if result.Error != nil {
		return fmt.Errorf("failed to save usage record: %w", result.Error)
	}

	log.Printf("✅ 使用量记录成功: 用户=%s 模型=%s 成本=%.4f 售价=%.4f 利润=%.4f",
		usage.UserID, usage.ModelID, usage.CostPrice, usage.SellingPrice, usage.Profit)

	return nil
}

// validateUsageRecord 验证使用记录数据完整性
func (m *Manager) validateUsageRecord(usage *models.UsageRecord) error {
	if usage.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	if usage.ModelID == "" {
		return fmt.Errorf("model_id is required")
	}

	if usage.SupplierID == "" {
		return fmt.Errorf("supplier_id is required")
	}

	if usage.TotalTokens <= 0 {
		return fmt.Errorf("total_tokens must be greater than 0")
	}

	if usage.CostPrice < 0 {
		return fmt.Errorf("cost_price cannot be negative")
	}

	if usage.SellingPrice < 0 {
		return fmt.Errorf("selling_price cannot be negative")
	}

	if usage.Profit != usage.SellingPrice-usage.CostPrice {
		return fmt.Errorf("profit calculation mismatch")
	}

	return nil
}

// GenerateBill 生成账单
func (m *Manager) GenerateBill(ctx context.Context, userID string, startDate, endDate time.Time) (*Bill, error) {
	// 查询用户的使用记录
	var records []models.UsageRecord
	result := m.db.WithContext(ctx).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startDate, endDate).
		Order("created_at ASC").
		Find(&records)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch usage records: %w", result.Error)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no usage records found for the given period")
	}

	// 生成账单
	bill := &Bill{
		ID:        xid.New().String(),
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
		CreatedAt: time.Now().UTC(),
	}

	// 按模型分组统计
	modelStats := make(map[string]*ModelUsage)
	for _, record := range records {
		if stat, exists := modelStats[record.ModelID]; exists {
			stat.TotalRequests++
			stat.TotalTokens += record.TotalTokens
			stat.InputTokens += record.InputTokens
			stat.OutputTokens += record.OutputTokens
			stat.TotalCost += record.CostPrice
			stat.TotalRevenue += record.SellingPrice
			stat.TotalProfit += record.Profit
		} else {
			modelStats[record.ModelID] = &ModelUsage{
				ModelID:       record.ModelID,
				TotalRequests: 1,
				TotalTokens:   record.TotalTokens,
				InputTokens:   record.InputTokens,
				OutputTokens:  record.OutputTokens,
				TotalCost:     record.CostPrice,
				TotalRevenue:  record.SellingPrice,
				TotalProfit:   record.Profit,
			}
		}
	}

	// 计算总计
	for _, stat := range modelStats {
		bill.Items = append(bill.Items, stat)
		bill.TotalCost += stat.TotalCost
		bill.TotalRevenue += stat.TotalRevenue
		bill.TotalProfit += stat.TotalProfit
	}

	bill.TotalRequests = int64(len(records))

	return bill, nil
}

// GetUserUsageSummary 获取用户使用量摘要
func (m *Manager) GetUserUsageSummary(ctx context.Context, userID string, days int) (*UsageSummary, error) {
	// 计算时间范围
	endDate := time.Now().UTC()
	startDate := endDate.AddDate(0, 0, -days)

	// 查询使用记录
	var result struct {
		TotalRequests int64
		TotalTokens   int64
		TotalCost     float64
		TotalRevenue  float64
		TotalProfit   float64
	}

	err := m.db.WithContext(ctx).
		Table("usage_records").
		Select(
			"COUNT(*) as total_requests",
			"SUM(total_tokens) as total_tokens",
			"SUM(cost_price) as total_cost",
			"SUM(selling_price) as total_revenue",
			"SUM(profit) as total_profit",
		).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startDate, endDate).
		Scan(&result).
		Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch usage summary: %w", err)
	}

	summary := &UsageSummary{
		UserID:        userID,
		StartDate:     startDate,
		EndDate:       endDate,
		TotalRequests: result.TotalRequests,
		TotalTokens:   result.TotalTokens,
		TotalCost:     result.TotalCost,
		TotalRevenue:  result.TotalRevenue,
		TotalProfit:   result.TotalProfit,
	}

	if summary.TotalRevenue > 0 {
		summary.ProfitMargin = summary.TotalProfit / summary.TotalRevenue
	}

	return summary, nil
}

// GetModelUsageStats 获取模型使用统计
func (m *Manager) GetModelUsageStats(ctx context.Context, modelID string, days int) (*ModelStats, error) {
	endDate := time.Now().UTC()
	startDate := endDate.AddDate(0, 0, -days)

	var result struct {
		TotalRequests int64
		TotalTokens   int64
		TotalCost     float64
		TotalRevenue  float64
		AvgCostPer1kTokens float64
	}

	err := m.db.WithContext(ctx).
		Table("usage_records").
		Select(
			"COUNT(*) as total_requests",
			"SUM(total_tokens) as total_tokens",
			"SUM(cost_price) as total_cost",
			"SUM(selling_price) as total_revenue",
			"(SUM(cost_price) / SUM(total_tokens)) * 1000 as avg_cost_per_1k_tokens",
		).
		Where("model_id = ? AND created_at >= ? AND created_at < ?", modelID, startDate, endDate).
		Scan(&result).
		Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch model stats: %w", err)
	}

	return &ModelStats{
		ModelID:            modelID,
		StartDate:          startDate,
		EndDate:            endDate,
		TotalRequests:      result.TotalRequests,
		TotalTokens:        result.TotalTokens,
		TotalCost:          result.TotalCost,
		TotalRevenue:       result.TotalRevenue,
		AvgCostPer1kTokens: result.AvgCostPer1kTokens,
	}, nil
}

// 账单相关数据结构

// Bill 账单
type Bill struct {
	ID           string         `json:"id"`
	UserID       string         `json:"user_id"`
	StartDate    time.Time      `json:"start_date"`
	EndDate      time.Time      `json:"end_date"`
	Items        []*ModelUsage  `json:"items"`
	TotalRequests int64         `json:"total_requests"`
	TotalCost    float64        `json:"total_cost"`
	TotalRevenue float64        `json:"total_revenue"`
	TotalProfit  float64        `json:"total_profit"`
	CreatedAt    time.Time      `json:"created_at"`
}

// ModelUsage 模型使用详情
type ModelUsage struct {
	ModelID       string  `json:"model_id"`
	TotalRequests int     `json:"total_requests"`
	TotalTokens   int32   `json:"total_tokens"`
	InputTokens   int32   `json:"input_tokens"`
	OutputTokens  int32   `json:"output_tokens"`
	TotalCost     float64 `json:"total_cost"`
	TotalRevenue  float64 `json:"total_revenue"`
	TotalProfit   float64 `json:"total_profit"`
}

// UsageSummary 使用量摘要
type UsageSummary struct {
	UserID        string    `json:"user_id"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	TotalRequests int64     `json:"total_requests"`
	TotalTokens   int64     `json:"total_tokens"`
	TotalCost     float64   `json:"total_cost"`
	TotalRevenue  float64   `json:"total_revenue"`
	TotalProfit   float64   `json:"total_profit"`
	ProfitMargin  float64   `json:"profit_margin"`
}

// ModelStats 模型统计
type ModelStats struct {
	ModelID            string    `json:"model_id"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
	TotalRequests      int64     `json:"total_requests"`
	TotalTokens        int64     `json:"total_tokens"`
	TotalCost          float64   `json:"total_cost"`
	TotalRevenue       float64   `json:"total_revenue"`
	AvgCostPer1kTokens float64   `json:"avg_cost_per_1k_tokens"`
}
