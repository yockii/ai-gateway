package database

import (
	"fmt"
	"log"
	"time"

	"github.com/yockii/ai-gateway/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// BuildDSN 构建 PostgreSQL DSN
func BuildDSN(host, port, user, password, dbname string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
}

// DB 数据库封装
type DB struct {
	*gorm.DB
}

// New 创建数据库实例
func New(dsn string) (*DB, error) {
	// 配置 GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		// 禁用外键约束（性能优化）
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 连接 PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	database := &DB{DB: db}

	// 自动迁移表结构
	if err := database.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("✅ Database initialized successfully")

	return database, nil
}

// AutoMigrate 自动迁移表结构
func (db *DB) AutoMigrate() error {
	log.Println("开始数据库表结构自动迁移...")

	// 定义所有需要迁移的模型
	models := []interface{}{
		&models.User{},
		&models.Admin{},
		&models.UserGroup{},
		&models.Supplier{},
		&models.SupplierCostPricing{},
		&models.UserGroupPricing{},
		&models.ExternalModel{},
		&models.UsageRecord{},
		&models.UserAPIKey{},
	}

	// 逐个迁移模型
	for _, model := range models {
		if err := db.DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
		log.Printf("✅ 迁移表: %T", model)
	}

	// 创建索引
	if err := db.createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("✅ 数据库表结构自动迁移完成")
	return nil
}

// createIndexes 创建额外的索引
func (db *DB) createIndexes() error {
	// UsageRecord 复合索引
	// 用于查询用户使用记录和时间范围统计
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_usage_records_user_created ON usage_records(user_id, created_at)").Error; err != nil {
		log.Printf("警告: 创建索引失败 (可能已存在): %v", err)
	}

	// SupplierCostPricing 复合索引
	// 用于查询供应商的活跃定价
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_supplier_cost_supplier_model_active ON supplier_cost_pricings(supplier_id, model_id, is_active)").Error; err != nil {
		log.Printf("警告: 创建索引失败 (可能已存在): %v", err)
	}

	// UserGroupPricing 复合索引
	// 用于查询用户群体的活跃定价
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_user_group_pricing_group_model_active ON user_group_pricings(user_group_id, model_id, is_active)").Error; err != nil {
		log.Printf("警告: 创建索引失败 (可能已存在): %v", err)
	}

	return nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
