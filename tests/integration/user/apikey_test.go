package user_test

import (
	"context"
	"testing"
	"time"

	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/yockii/ai-gateway/internal/services"
	"github.com/yockii/ai-gateway/tests/integration/testcontainers"
)

// TestKeyManager_CreateKey 测试创建 API Key
func TestKeyManager_CreateKey(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	km, _ := services.NewKeyManager(db)

	userID := "test-user-001"

	key, err := km.CreateKey(ctx, userID, "Test Key", services.CreateKeyOptions{
		QuotaDaily:       1000,
		QuotaMonthly:     30000,
		ConcurrencyLimit: 10,
	})

	if err != nil {
		t.Fatalf("Failed to create key: %v", err)
	}

	if key.ID == "" {
		t.Error("Key ID should not be empty")
	}

	if key.KeyValue == "" {
		t.Error("KeyValue should not be empty")
	}

	if !key.IsActive {
		t.Error("New key should be active")
	}

	if key.Name != "Test Key" {
		t.Errorf("Expected name 'Test Key', got '%s'", key.Name)
	}
}

// TestKeyManager_ValidateKey 测试验证 API Key
func TestKeyManager_ValidateKey(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	km, _ := services.NewKeyManager(db)

	userID := "test-user-002"
	key, _ := km.CreateKey(ctx, userID, "Validation Test", services.CreateKeyOptions{})

	// 测试有效 key
	keyInfo, err := km.ValidateKey(ctx, key.KeyValue)
	if err != nil {
		t.Fatalf("Failed to validate valid key: %v", err)
	}

	if keyInfo.UserID != userID {
		t.Errorf("Expected userID '%s', got '%s'", userID, keyInfo.UserID)
	}

	if keyInfo.KeyID != key.ID {
		t.Errorf("Expected keyID '%s', got '%s'", key.ID, keyInfo.KeyID)
	}

	// 测试无效 key
	_, err = km.ValidateKey(ctx, "sk-invalid-key")
	if err == nil {
		t.Error("Should return error for invalid key")
	}
}

// TestKeyManager_DisableKey 测试禁用 Key
func TestKeyManager_DisableKey(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	km, _ := services.NewKeyManager(db)

	userID := "test-user-003"
	key, _ := km.CreateKey(ctx, userID, "Disable Test", services.CreateKeyOptions{})

	// 禁用 key
	err := km.UpdateKeyStatus(ctx, key.ID, userID, false)
	if err != nil {
		t.Fatalf("Failed to disable key: %v", err)
	}

	// 验证禁用后无法使用
	_, err = km.ValidateKey(ctx, key.KeyValue)
	if err == nil {
		t.Error("Disabled key should be invalid")
	}
}

// TestKeyManager_UpdateKey 测试更新 Key 配置
func TestKeyManager_UpdateKey(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	km, _ := services.NewKeyManager(db)

	userID := "test-user-004"
	key, _ := km.CreateKey(ctx, userID, "Update Test", services.CreateKeyOptions{
		QuotaDaily: 1000,
	})

	// 更新配置
	err := km.UpdateKey(ctx, key.ID, userID, services.UpdateKeyOptions{
		Name:       "Updated Key",
		QuotaDaily: 5000,
	})

	if err != nil {
		t.Fatalf("Failed to update key: %v", err)
	}

	// 验证更新
	keys, _ := km.GetUserKeys(ctx, userID)
	if len(keys) != 1 {
		t.Fatalf("Expected 1 key, got %d", len(keys))
	}

	if keys[0].Name != "Updated Key" {
		t.Errorf("Expected name 'Updated Key', got '%s'", keys[0].Name)
	}

	if keys[0].QuotaDaily != 5000 {
		t.Errorf("Expected QuotaDaily 5000, got %d", keys[0].QuotaDaily)
	}
}

// TestKeyManager_DeleteKey 测试删除 Key
func TestKeyManager_DeleteKey(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	km, _ := services.NewKeyManager(db)

	userID := "test-user-005"
	key, _ := km.CreateKey(ctx, userID, "Delete Test", services.CreateKeyOptions{})

	// 删除 key
	err := km.DeleteKey(ctx, key.ID, userID)
	if err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	// 验证删除后无法获取
	keys, _ := km.GetUserKeys(ctx, userID)
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys after deletion, got %d", len(keys))
	}
}

// TestKeyManager_QuotaCheck 测试额度检查
func TestKeyManager_QuotaCheck(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	km, _ := services.NewKeyManager(db)

	userID := "test-user-006"
	key, _ := km.CreateKey(ctx, userID, "Quota Test", services.CreateKeyOptions{
		QuotaDaily: 10,
	})

	// 创建一些使用记录
	for i := 0; i < 5; i++ {
		usage := &models.UsageRecord{
			ID:         generateID(),
			RequestID:  generateID(),
			UserID:     userID,
			KeyID:      key.ID,
			ModelID:    "gpt-4",
			SupplierID: "openai",
			InputTokens: 100,
			OutputTokens: 50,
			TotalTokens: 150,
			CostPrice: 0.001,
			SellingPrice: 0.002,
			Profit: 0.001,
			CreatedAt: time.Now(),
		}
		db.WithContext(ctx).Create(usage)
	}

	// 应该仍然在限额内
	err := km.CheckQuota(ctx, key.ID)
	if err != nil {
		t.Errorf("Should be within quota: %v", err)
	}
}

// TestKeyManager_GetStats 测试获取统计
func TestKeyManager_GetStats(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	km, _ := services.NewKeyManager(db)

	userID := "test-user-007"
	key, _ := km.CreateKey(ctx, userID, "Stats Test", services.CreateKeyOptions{
		QuotaDaily:  1000,
		QuotaMonthly: 30000,
	})

	// 创建使用记录
	usage := &models.UsageRecord{
		ID:         generateID(),
		RequestID:  generateID(),
		UserID:     userID,
		KeyID:      key.ID,
		ModelID:    "gpt-4",
		SupplierID: "openai",
		InputTokens: 500,
		OutputTokens: 300,
		TotalTokens: 800,
		CostPrice: 0.01,
		SellingPrice: 0.02,
		Profit: 0.01,
		CreatedAt: time.Now(),
	}
	db.WithContext(ctx).Create(usage)

	// 获取统计
	stats, err := km.GetKeyStats(ctx, key.ID)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TotalRequests != 1 {
		t.Errorf("Expected 1 request, got %d", stats.TotalRequests)
	}

	if stats.TotalTokens != 800 {
		t.Errorf("Expected 800 tokens, got %d", stats.TotalTokens)
	}
}

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *database.DB {
	ctx := context.Background()
	
	// 使用 testcontainers 启动 PostgreSQL
	container, err := testcontainers.SetupPostgreSQL(ctx)
	if err != nil {
		t.Skipf("Skipping test: failed to setup PostgreSQL: %v", err)
	}
	t.Cleanup(func() {
		container.Terminate(ctx)
	})

	// 连接到测试数据库
	db, err := database.New(container.GetDSN())
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// generateID 生成测试 ID
func generateID() string {
	return time.Now().Format("20060102150405") + "-test"
}
