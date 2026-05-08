package database

import (
	"os"
	"testing"

	"github.com/yockii/ai-gateway/internal/models"
)

// TestDatabaseIntegration 数据库集成测试
func TestDatabaseIntegration(t *testing.T) {
	// 从环境变量获取数据库配置
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "ai_gateway_test")

	// 如果没有设置数据库密码，跳过测试
	if dbPassword == "" {
		t.Skip("跳过数据库集成测试：未设置 DB_PASSWORD")
	}

	// 构建测试数据库 DSN
	dsn := BuildDSN(dbHost, dbPort, dbUser, dbPassword, dbName)

	// 初始化数据库
	db, err := New(dsn)
	if err != nil {
		t.Fatalf("数据库初始化失败: %v", err)
	}
	defer db.Close()

	t.Log("✅ 数据库连接成功")

	// 测试自动迁移
	t.Run("AutoMigrate", func(t *testing.T) {
		// 表应该已经自动创建
		// 验证表是否存在
		if err := db.Exec("SELECT 1 FROM users LIMIT 1").Error; err != nil {
			t.Errorf("users 表不存在或查询失败: %v", err)
		}

		if err := db.Exec("SELECT 1 FROM usage_records LIMIT 1").Error; err != nil {
			t.Errorf("usage_records 表不存在或查询失败: %v", err)
		}

		t.Log("✅ 数据库迁移成功")
	})

	// 测试插入和查询
	t.Run("InsertAndQuery", func(t *testing.T) {
		// 创建测试用户
		user := &models.User{
			Email:       "test@example.com",
			Name:        "Test User",
			UserGroupID: "default",
			IsActive:    true,
		}

		// 插入用户
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("插入用户失败: %v", err)
		}

		t.Logf("✅ 用户插入成功: ID=%s", user.ID)

		// 查询用户
		var queriedUser models.User
		result := db.Where("email = ?", "test@example.com").First(&queriedUser)
		if result.Error != nil {
			t.Fatalf("查询用户失败: %v", result.Error)
		}

		if queriedUser.Email != user.Email {
			t.Errorf("用户邮箱不匹配: 期望 %s, 实际 %s", user.Email, queriedUser.Email)
		}

		t.Log("✅ 用户查询成功")

		// 清理测试数据
		db.Delete(user)
	})
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
