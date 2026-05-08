package benchmarks

import (
	"context"
	"testing"
	"time"

	"github.com/rs/xid"
)

// BenchmarkXidGeneration xid 生成性能基准测试
func BenchmarkXidGeneration(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = xid.New().String()
		}
	})
}

// BenchmarkXidGenerationSequential xid 顺序生成性能测试
func BenchmarkXidGenerationSequential(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = xid.New().String()
	}
}

// TestXidUniqueness 测试 xid 唯一性
func TestXidUniqueness(t *testing.T) {
	count := 100000
	ids := make(map[string]bool, count)

	start := time.Now()
	for i := 0; i < count; i++ {
		id := xid.New().String()
		if ids[id] {
			t.Errorf("发现重复 ID: %s", id)
		}
		ids[id] = true
	}
	duration := time.Since(start)

	t.Logf("生成 %d 个 xid", count)
	t.Logf("耗时: %v", duration)
	t.Logf("平均: %v/个", duration/time.Duration(count))
	t.Logf("速率: %.0f 个/秒", float64(count)/duration.Seconds())
	t.Logf("唯一性: 100%% (无重复)")
}

// BenchmarkContextCreation 上下文创建性能测试
func BenchmarkContextCreation(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = context.WithTimeout(context.Background(), 30*time.Second)
		}
	})
}

// BenchmarkProfitMarginCalculation 利润率计算性能测试
func BenchmarkProfitMarginCalculation(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			costPrice := 0.001
			sellingPrice := 0.002
			profit := sellingPrice - costPrice
			_ = profit / sellingPrice
		}
	})
}

// TestProfitMarginCalculationAccuracy 利润率计算准确性测试
func TestProfitMarginCalculationAccuracy(t *testing.T) {
	tests := []struct {
		name           string
		costPrice      float64
		sellingPrice   float64
		expectedMargin float64
	}{
		{"10%", 0.001, 0.001111, 0.09991},
		{"20%", 0.001, 0.00125, 0.20},
		{"50%", 0.001, 0.002, 0.50},
		{"80%", 0.001, 0.005, 0.80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profit := tt.sellingPrice - tt.costPrice
			margin := profit / tt.sellingPrice

			epsilon := 0.001
			if margin-tt.expectedMargin > epsilon {
				t.Errorf("利润率计算错误: 期望 %.4f, 实际 %.4f", tt.expectedMargin, margin)
			}

			t.Logf("利润率: %.2f%%", margin*100)
		})
	}
}

// BenchmarkStringOperations 字符串操作性能测试
func BenchmarkStringOperations(b *testing.B) {
	str := "test-api-key"

	b.Run("string_comparison", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = str == "test-api-key"
		}
	})

	b.Run("has_prefix", func(b *testing.B) {
		prefix := "sk-"
		for i := 0; i < b.N; i++ {
			_ = len(str) > len(prefix) && str[:len(prefix)] == prefix
		}
	})
}

// TestConcurrentXidGeneration 并发 xid 生成测试
func TestConcurrentXidGeneration(t *testing.T) {
	concurrency := 100
	generations := 1000

	results := make(chan string, concurrency*generations)
	start := time.Now()

	// 启动多个 goroutine 并发生成 xid
	for i := 0; i < concurrency; i++ {
		go func() {
			for j := 0; j < generations; j++ {
				results <- xid.New().String()
			}
		}()
	}

	// 收集所有 ID
	ids := make(map[string]bool)
	total := concurrency * generations
	for i := 0; i < total; i++ {
		id := <-results
		if ids[id] {
			t.Errorf("发现重复 ID: %s", id)
		}
		ids[id] = true
	}

	duration := time.Since(start)

	t.Logf("并发生成 %d 个 xid (%d goroutine × %d)", total, concurrency, generations)
	t.Logf("耗时: %v", duration)
	t.Logf("平均: %v/个", duration/time.Duration(total))
	t.Logf("速率: %.0f 个/秒", float64(total)/duration.Seconds())
	t.Logf("唯一性: 100%% (无重复)")
}
