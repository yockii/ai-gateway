package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rs/xid"
	"github.com/yockii/ai-gateway/internal/auth"
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

	// 配置连接池（优化高性能场景）
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 设置连接池参数（针对高并发优化）
	sqlDB.SetMaxIdleConns(20)             // 增加空闲连接数
	sqlDB.SetMaxOpenConns(200)            // 增加最大连接数以支持 10000+ QPS
	sqlDB.SetConnMaxLifetime(time.Hour)   // 连接最大生存时间
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // 空闲连接最大生存时间

	database := &DB{DB: db}

	// 注册查询性能监控回调
	registerQueryCallbacks(db)

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
		&models.MembershipTier{},
		&models.MembershipDiscount{},
		&models.UserMembership{},
		&models.ModelMapping{},
		&models.SupplierCostPricingExtended{},
		&models.UserGroupPricingExtended{},
		&models.Bill{},
		&models.BillItem{},
		&models.EnterprisePricing{}, // Phase 5: 大客户独立定价
		&models.SupplierApiKey{},     // Phase 6: 供应商 API Key
		&models.SupplierModel{},      // Phase 6: 供应商模型关联
		&models.PriceHistory{},       // Phase 6: 价格变更历史
		&models.SupplierFailureEvent{}, // Phase 8: 供应商失败事件
		&models.FailoverEvent{},      // Phase 8: 故障转移事件
		&models.HealthCheckHistory{}, // Phase 8: 健康检查历史
			&models.AuditLog{},           // Phase 9: 审计日志
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

	// UsageRecord 模型和时间索引
	// 用于按模型统计使用情况
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_usage_records_model_created ON usage_records(model_id, created_at)").Error; err != nil {
		log.Printf("警告: 创建索引失败 (可能已存在): %v", err)
	}

	// UsageRecord 供应商索引
	// 用于供应商统计
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_usage_records_supplier_created ON usage_records(supplier_id, created_at)").Error; err != nil {
		log.Printf("警告: 创建索引失败 (可能已存在): %v", err)
	}

	// UserAPIKey 复合索引
	// 用于查询用户的活跃 API Key
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_user_api_keys_user_active ON user_api_keys(user_id, is_active)").Error; err != nil {
		log.Printf("警告: 创建索引失败 (可能已存在): %v", err)
	}

	// UserAPIKey 过期时间索引
	// 用于查询过期密钥
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_user_api_keys_expires_active ON user_api_keys(expires_at, is_active)").Error; err != nil {
		log.Printf("警告: 创建索引失败 (可能已存在): %v", err)
	}

	// Bill 复合索引
	// 用于查询用户账单
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_bills_user_period ON bills(user_id, period)").Error; err != nil {
		log.Printf("警告: 创建索引失败 (可能已存在): %v", err)
	}

	// Bill 状态索引
	// 用于查询待支付账单
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_bills_status ON bills(status)").Error; err != nil {
		log.Printf("警告: 创建索引失败 (可能已存在): %v", err)
	}

	// BillItem 账单明细索引
	// 用于查询账单明细
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_bill_items_bill_id ON bill_items(bill_id)").Error; err != nil {
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

	// UserMembership 复合索引
	// 用于查询用户会员状态
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_user_memberships_user_tier_active ON user_memberships(user_id, membership_tier_id, is_active)").Error; err != nil {
		log.Printf("警告: 创建索引失败 (可能已存在): %v", err)
	}

	// UserMembership 过期时间索引
	// 用于查询过期会员
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_user_memberships_expires_active ON user_memberships(expires_at, is_active)").Error; err != nil {
		log.Printf("警告: 创建索引失败 (可能已存在): %v", err)
	}

	// ExternalModel 类型索引
	// 用于按类型查询模型
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_external_models_type_active ON external_models(model_type, is_active)").Error; err != nil {
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

// queryStartTimeKey 用于在 context 中存储查询开始时间
type queryStartTimeKey struct{}

// registerQueryCallbacks 注册查询性能监控回调
func registerQueryCallbacks(db *gorm.DB) {
	// 记录查询开始时间
	db.Callback().Query().Before("gorm:query").Register("query:start_time", func(db *gorm.DB) {
		db.InstanceSet("query:start_time", time.Now())
	})

	// 记录慢查询
	db.Callback().Query().After("gorm:query").Register("query:log_slow", func(db *gorm.DB) {
		if startTime, ok := db.InstanceGet("query:start_time"); ok {
			if start, ok := startTime.(time.Time); ok {
			elapsed := time.Since(start)
			// 记录超过 100ms 的查询
			if elapsed > 100*time.Millisecond {
				log.Printf("⚠️  Slow query detected: %v took %v", db.Statement.SQL.String(), elapsed)
			}
			}
		}
	})

	// 记录创建开始时间
	db.Callback().Create().Before("gorm:create").Register("create:start_time", func(db *gorm.DB) {
		db.InstanceSet("create:start_time", time.Now())
	})

	// 记录慢创建
	db.Callback().Create().After("gorm:create").Register("create:log_slow", func(db *gorm.DB) {
		if startTime, ok := db.InstanceGet("create:start_time"); ok {
			if start, ok := startTime.(time.Time); ok {
			elapsed := time.Since(start)
			if elapsed > 100*time.Millisecond {
				log.Printf("⚠️  Slow create detected: %v took %v", db.Statement.SQL.String(), elapsed)
			}
			}
		}
	})

	// 记录更新开始时间
	db.Callback().Update().Before("gorm:update").Register("update:start_time", func(db *gorm.DB) {
		db.InstanceSet("update:start_time", time.Now())
	})

	// 记录慢更新
	db.Callback().Update().After("gorm:update").Register("update:log_slow", func(db *gorm.DB) {
		if startTime, ok := db.InstanceGet("update:start_time"); ok {
			if start, ok := startTime.(time.Time); ok {
			elapsed := time.Since(start)
			if elapsed > 100*time.Millisecond {
				log.Printf("⚠️  Slow update detected: %v took %v", db.Statement.SQL.String(), elapsed)
			}
			}
		}
	})
}

// WithContext 创建带 context 的 DB 实例
func (db *DB) WithContext(ctx context.Context) *gorm.DB {
	return db.DB.WithContext(ctx)
}

// InitializeDefaultAdmin 初始化默认管理员
// 如果数据库中不存在任何管理员，则创建默认管理员
func (db *DB) InitializeDefaultAdmin() error {
	ctx := context.Background()

	// 检查是否已有管理员
	var count int64
	if err := db.WithContext(ctx).Model(&models.Admin{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check admin count: %w", err)
	}

	// 如果已有管理员，不创建默认管理员
	if count > 0 {
		log.Printf("数据库中已有 %d 个管理员，跳过默认管理员创建", count)
		return nil
	}

	// 创建默认管理员
	defaultPassword := "admin123456" // 默认密码，首次登录后应修改
	hashedPassword, err := auth.HashPassword(defaultPassword)
	if err != nil {
		return fmt.Errorf("failed to hash default password: %w", err)
	}

	admin := &models.Admin{
		ID:        xid.New().String(),
		Email:     "admin@example.com",
		Password:  hashedPassword,
		Name:      "默认管理员",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.WithContext(ctx).Create(admin).Error; err != nil {
		return fmt.Errorf("failed to create default admin: %w", err)
	}

	log.Println("⚠️  默认管理员已创建:")
	log.Println("   邮箱: admin@example.com")
	log.Println("   密码: admin123456")
	log.Println("   ⚠️  请在生产环境中立即修改默认密码！")

	return nil
}

// NewTestDB 创建测试数据库实例（使用环境变量配置）
func NewTestDB() (*DB, error) {
	// 使用环境变量配置测试数据库
	testDSN := BuildDSN(
		"localhost",
		"5432",
		"test",
		"test",
		"ai_gateway_test",
	)

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	db, err := gorm.Open(postgres.Open(testDSN), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	database := &DB{DB: db}

	// 测试数据库自动迁移
	if err := database.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to auto migrate test database: %w", err)
	}

	return database, nil
}
