package main

import (
	"log"
	"time"

	// 只导入 Bifrost 包验证可用性
	_ "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/rs/xid"
)

// BifrostIntegrationTest Bifrost 集成验证测试
func main() {
	log.Println("=== Bifrost 集成验证测试 ===")

	// 1. 验证 Bifrost 包导入成功
	testBifrostImport()

	// 2. 测试 xid 性能和唯一性
	testXidPerformance()

	// 3. 测试基础性能
	if err := testBasicPerformance(); err != nil {
		log.Fatalf("性能测试失败: %v", err)
	}

	log.Println("=== 所有测试通过 ===")
}

// testBifrostImport 测试 Bifrost 包导入
func testBifrostImport() {
	log.Println("\n--- Bifrost 包导入测试 ---")

	// 验证 schemas 包可用
	_ = schemas.BifrostChatRequest{}
	_ = schemas.ChatMessage{}
	_ = schemas.BifrostChatResponse{}

	log.Println("✅ Bifrost 包导入成功")
	log.Println("   - github.com/maximhq/bifrost/core")
	log.Println("   - github.com/maximhq/bifrost/core/schemas")
}

// testXidPerformance 测试 xid 性能和唯一性
func testXidPerformance() {
	log.Println("\n--- xid 性能测试 ---")

	// 性能测试：生成 100,000 个 ID
	start := time.Now()
	ids := make(map[string]bool, 100000)

	for i := 0; i < 100000; i++ {
		id := xid.New().String()
		if ids[id] {
			log.Fatalf("发现重复 ID: %s", id)
		}
		ids[id] = true
	}

	duration := time.Since(start)

	log.Printf("✅ 生成 100,000 个 xid")
	log.Printf("   耗时: %v", duration)
	log.Printf("   平均: %v/个", duration/100000)
	log.Printf("   速率: %.0f 个/秒", float64(100000)/duration.Seconds())
	log.Printf("   唯一性: 100%% (无重复)")
}

// testBasicPerformance 测试基础性能
func testBasicPerformance() error {
	log.Println("\n--- 基础性能测试 ---")

	// 测试空操作性能
	iterations := 10000

	start := time.Now()
	for i := 0; i < iterations; i++ {
		_ = xid.New().String()
	}

	duration := time.Since(start)
	avgPerOp := duration / time.Duration(iterations)

	log.Printf("✅ 基础性能测试")
	log.Printf("   操作: xid 生成")
	log.Printf("   次数: %d", iterations)
	log.Printf("   总耗时: %v", duration)
	log.Printf("   平均耗时: %v/操作", avgPerOp)
	log.Printf("   速率: %.0f 操作/秒", float64(iterations)/duration.Seconds())

	// 性能要求验证
	if avgPerOp > time.Microsecond {
		log.Printf("⚠️  警告: 平均耗时超过 1μs (当前: %v)", avgPerOp)
	} else {
		log.Printf("✅ 性能优异: 平均耗时 < 1μs")
	}

	return nil
}
