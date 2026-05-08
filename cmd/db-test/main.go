package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/yockii/ai-gateway/internal/database"
)

func main() {
	log.Println("=== 数据库连接和迁移测试 ===")

	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("警告: .env 文件未找到，使用环境变量")
	}

	// 从环境变量获取数据库配置
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "ai_gateway")

	// 构建 DSN
	dsn := database.BuildDSN(dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("数据库配置:")
	log.Printf("  Host: %s", dbHost)
	log.Printf("  Port: %s", dbPort)
	log.Printf("  User: %s", dbUser)
	log.Printf("  Database: %s", dbName)
	log.Printf("  SSL Mode: disable")
	log.Println("")

	// 初始化数据库
	log.Println("正在连接数据库...")
	db, err := database.New(dsn)
	if err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	defer db.Close()

	log.Println("✅ 数据库连接成功")
	log.Println("")
	log.Println("测试完成！")
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
