package cost

import (
	"testing"
)

// TestProfitMarginCalculation 测试利润率计算
func TestProfitMarginCalculation(t *testing.T) {
	tests := []struct {
		name           string
		costPrice      float64
		sellingPrice   float64
		expectedMargin float64
	}{
		{
			name:           "50% 利润率",
			costPrice:      0.001,
			sellingPrice:   0.002,
			expectedMargin: 0.50,
		},
		{
			name:           "10% 利润率",
			costPrice:      0.001,
			sellingPrice:   0.001111,
			expectedMargin: 0.09991,
		},
		{
			name:           "80% 利润率",
			costPrice:      0.001,
			sellingPrice:   0.005,
			expectedMargin: 0.80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profit := tt.sellingPrice - tt.costPrice
			margin := profit / tt.sellingPrice

			epsilon := 0.001
			if margin-tt.expectedMargin > epsilon {
				t.Errorf("利润率计算错误: 期望 %.4f, 实际 %.4f", tt.expectedMargin, margin)
			}

			t.Logf("成本: %.4f, 售价: %.4f, 利润: %.4f, 利润率: %.2f%%",
				tt.costPrice, tt.sellingPrice, profit, margin*100)
		})
	}
}

// TestMinProfitMarginValidation 测试最小利润率验证
func TestMinProfitMarginValidation(t *testing.T) {
	minProfitMargin := 0.10

	tests := []struct {
		name          string
		costPrice     float64
		sellingPrice  float64
		shouldBeValid bool
	}{
		{
			name:          "正常利润率",
			costPrice:     0.001,
			sellingPrice:  0.002,
			shouldBeValid: true,
		},
		{
			name:          "刚好满足最小利润率",
			costPrice:     0.001,
			sellingPrice:  0.001112, // 略微提高售价以确保 > 10%
			shouldBeValid: true,
		},
		{
			name:          "低于最小利润率",
			costPrice:     0.001,
			sellingPrice:  0.00105,
			shouldBeValid: false,
		},
		{
			name:          "零利润",
			costPrice:     0.001,
			sellingPrice:  0.001,
			shouldBeValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profit := tt.sellingPrice - tt.costPrice
			profitMargin := profit / tt.sellingPrice

			isValid := profitMargin >= minProfitMargin

			if isValid != tt.shouldBeValid {
				t.Errorf("利润率验证失败: 成本=%.4f, 售价=%.4f, 利润率=%.2f%%, 期望有效=%v, 实际=%v",
					tt.costPrice, tt.sellingPrice, profitMargin*100, tt.shouldBeValid, isValid)
			}
		})
	}
}
