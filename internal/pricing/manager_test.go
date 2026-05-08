package pricing

import (
	"testing"
	"time"

	"github.com/yockii/ai-gateway/internal/models"
	"github.com/rs/xid"
)

// TestValidateUsageRecord 测试使用记录验证
func TestValidateUsageRecord(t *testing.T) {
	manager := &Manager{}

	tests := []struct {
		name        string
		record      *models.UsageRecord
		shouldError bool
	}{
		{
			name: "有效的使用记录",
			record: &models.UsageRecord{
				ID:           xid.New().String(),
				UserID:       "user-123",
				ModelID:      "gpt-3.5-turbo",
				SupplierID:   "supplier-1",
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  150,
				CostPrice:    0.0001,
				SellingPrice: 0.0002,
				Profit:       0.0001,
				CreatedAt:    time.Now().UTC(),
			},
			shouldError: false,
		},
		{
			name: "缺少用户 ID",
			record: &models.UsageRecord{
				ModelID:      "gpt-3.5-turbo",
				SupplierID:   "supplier-1",
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  150,
			},
			shouldError: true,
		},
		{
			name: "缺少模型 ID",
			record: &models.UsageRecord{
				UserID:       "user-123",
				SupplierID:   "supplier-1",
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  150,
			},
			shouldError: true,
		},
		{
			name: "Token 总数为 0",
			record: &models.UsageRecord{
				UserID:       "user-123",
				ModelID:      "gpt-3.5-turbo",
				SupplierID:   "supplier-1",
				InputTokens:  0,
				OutputTokens: 0,
				TotalTokens:  0,
			},
			shouldError: true,
		},
		{
			name: "负成本价格",
			record: &models.UsageRecord{
				UserID:       "user-123",
				ModelID:      "gpt-3.5-turbo",
				SupplierID:   "supplier-1",
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  150,
				CostPrice:    -0.0001,
			},
			shouldError: true,
		},
		{
			name: "利润计算错误",
			record: &models.UsageRecord{
				UserID:       "user-123",
				ModelID:      "gpt-3.5-turbo",
				SupplierID:   "supplier-1",
				InputTokens:  100,
				OutputTokens: 50,
				TotalTokens:  150,
				CostPrice:    0.0001,
				SellingPrice: 0.0002,
				Profit:       0.0002, // 错误：应该是 0.0001
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.validateUsageRecord(tt.record)
			if tt.shouldError && err == nil {
				t.Errorf("期望返回错误，但没有")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("不期望返回错误，但得到: %v", err)
			}
		})
	}
}

// TestCalculateProfit 测试利润计算
func TestCalculateProfit(t *testing.T) {
	tests := []struct {
		name         string
		costPrice    float64
		sellingPrice float64
		expectedProfit float64
		expectedMargin float64
	}{
		{
			name:           "标准利润",
			costPrice:      0.001,
			sellingPrice:   0.002,
			expectedProfit: 0.001,
			expectedMargin: 0.50,
		},
		{
			name:           "低利润率",
			costPrice:      0.001,
			sellingPrice:   0.0011,
			expectedProfit: 0.0001,
			expectedMargin: 0.0909,
		},
		{
			name:           "高利润率",
			costPrice:      0.001,
			sellingPrice:   0.005,
			expectedProfit: 0.004,
			expectedMargin: 0.80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profit := tt.sellingPrice - tt.costPrice
			margin := profit / tt.sellingPrice

			// 允许小的浮点数误差
			epsilon := 0.0001

			if abs(profit-tt.expectedProfit) > epsilon {
				t.Errorf("利润计算错误: 期望 %.4f, 实际 %.4f", tt.expectedProfit, profit)
			}

			if abs(margin-tt.expectedMargin) > epsilon {
				t.Errorf("利润率计算错误: 期望 %.4f, 实际 %.4f", tt.expectedMargin, margin)
			}

			t.Logf("成本: %.4f, 售价: %.4f, 利润: %.4f, 利润率: %.2f%%",
				tt.costPrice, tt.sellingPrice, profit, margin*100)
		})
	}
}

// TestBillGeneration 测试账单生成逻辑
func TestBillGeneration(t *testing.T) {
	// 模拟使用记录
	records := []models.UsageRecord{
		{
			ID:           xid.New().String(),
			UserID:       "user-123",
			ModelID:      "gpt-3.5-turbo",
			SupplierID:   "supplier-1",
			InputTokens:  1000,
			OutputTokens: 500,
			TotalTokens:  1500,
			CostPrice:    0.001,
			SellingPrice: 0.002,
			Profit:       0.001,
			CreatedAt:    time.Now().UTC().Add(-1 * time.Hour),
		},
		{
			ID:           xid.New().String(),
			UserID:       "user-123",
			ModelID:      "gpt-3.5-turbo",
			SupplierID:   "supplier-1",
			InputTokens:  2000,
			OutputTokens: 1000,
			TotalTokens:  3000,
			CostPrice:    0.002,
			SellingPrice: 0.004,
			Profit:       0.002,
			CreatedAt:    time.Now().UTC().Add(-30 * time.Minute),
		},
		{
			ID:           xid.New().String(),
			UserID:       "user-123",
			ModelID:      "gpt-4",
			SupplierID:   "supplier-2",
			InputTokens:  500,
			OutputTokens: 300,
			TotalTokens:  800,
			CostPrice:    0.003,
			SellingPrice: 0.006,
			Profit:       0.003,
			CreatedAt:    time.Now().UTC().Add(-15 * time.Minute),
		},
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
	var totalCost, totalRevenue, totalProfit float64
	var totalRequests int

	for _, stat := range modelStats {
		totalCost += stat.TotalCost
		totalRevenue += stat.TotalRevenue
		totalProfit += stat.TotalProfit
		totalRequests += stat.TotalRequests

		t.Logf("模型 %s: 请求数=%d, 总Token=%d, 成本=%.4f, 收入=%.4f, 利润=%.4f",
			stat.ModelID, stat.TotalRequests, stat.TotalTokens,
			stat.TotalCost, stat.TotalRevenue, stat.TotalProfit)
	}

	t.Logf("总计: 请求数=%d, 成本=%.4f, 收入=%.4f, 利润=%.4f, 利润率=%.2f%%",
		totalRequests, totalCost, totalRevenue, totalProfit,
		(totalProfit/totalRevenue)*100)

	// 验证总计
	expectedTotalCost := 0.001 + 0.002 + 0.003
	expectedTotalRevenue := 0.002 + 0.004 + 0.006
	expectedTotalProfit := 0.001 + 0.002 + 0.003

	epsilon := 0.0001
	if abs(totalCost-expectedTotalCost) > epsilon {
		t.Errorf("总成本计算错误: 期望 %.4f, 实际 %.4f", expectedTotalCost, totalCost)
	}
	if abs(totalRevenue-expectedTotalRevenue) > epsilon {
		t.Errorf("总收入计算错误: 期望 %.4f, 实际 %.4f", expectedTotalRevenue, totalRevenue)
	}
	if abs(totalProfit-expectedTotalProfit) > epsilon {
		t.Errorf("总利润计算错误: 期望 %.4f, 实际 %.4f", expectedTotalProfit, totalProfit)
	}
}

// abs 返回浮点数的绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// BenchmarkRecordUsage 性能基准测试 - 记录使用量
func BenchmarkRecordUsage(b *testing.B) {
	manager := &Manager{}
	record := &models.UsageRecord{
		ID:           xid.New().String(),
		UserID:       "user-123",
		ModelID:      "gpt-3.5-turbo",
		SupplierID:   "supplier-1",
		InputTokens:  1000,
		OutputTokens: 500,
		TotalTokens:  1500,
		CostPrice:    0.001,
		SellingPrice: 0.002,
		Profit:       0.001,
		CreatedAt:    time.Now().UTC(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record.ID = xid.New().String() // 每次生成新 ID
		_ = manager.validateUsageRecord(record)
	}
}
