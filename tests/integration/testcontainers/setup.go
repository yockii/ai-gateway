package testcontainers

import (
	"context"
	"fmt"
	"testing"
	"time"

	redisClient "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	redisContainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupPostgres starts a PostgreSQL container and returns connection string and cleanup function
func SetupPostgres(t *testing.T) (string, func()) {
	ctx := context.Background()

	// Generate unique database name per test to avoid conflicts
	dbName := fmt.Sprintf("test_%s", t.Name())

	// Configure PostgreSQL container
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Register cleanup function
	cleanup := func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate PostgreSQL container: %v", err)
		}
	}
	t.Cleanup(cleanup)

	return connStr, cleanup
}

// SetupRedis starts a Redis container and returns connection string and cleanup function
func SetupRedis(t *testing.T) (string, func()) {
	ctx := context.Background()

	// Configure Redis container
	container, err := redisContainer.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}

	// Get connection string
	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Failed to get Redis connection string: %v", err)
	}

	// Register cleanup function
	cleanup := func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate Redis container: %v", err)
		}
	}
	t.Cleanup(cleanup)

	return connStr, cleanup
}

// SetupFullStack starts both PostgreSQL and Redis containers, initializes GORM and Redis client
func SetupFullStack(t *testing.T) (*gorm.DB, *redisClient.Client, func()) {
	ctx := context.Background()

	// Start PostgreSQL
	pgConnStr, pgCleanup := SetupPostgres(t)

	// Initialize GORM DB
	db, err := gorm.Open(gormpostgres.Open(pgConnStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		t.Fatalf("Failed to initialize GORM: %v", err)
	}

	// Configure connection pool for tests
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get database instance: %v", err)
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Start Redis
	redisConnStr, redisCleanup := SetupRedis(t)

	// Initialize Redis client (connection string from testcontainers includes redis:// prefix, need to strip it)
	redisAddr := redisConnStr
	if len(redisAddr) > 8 && redisAddr[:8] == "redis://" {
		redisAddr = redisAddr[8:]
	}
	redisClient := redisClient.NewClient(&redisClient.Options{
		Addr: redisAddr,
	})

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Combined cleanup function
	cleanup := func() {
		if sqlDB != nil {
			sqlDB.Close()
		}
		if redisClient != nil {
			redisClient.Close()
		}
		redisCleanup()
		pgCleanup()
	}

	return db, redisClient, cleanup
}

// CleanupDB drops all tables in the database (useful for test isolation)
func CleanupDB(t *testing.T, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Drop all tables
	if err := db.Migrator().DropTable(
		"users", "admins", "user_groups",
		"suppliers", "supplier_cost_pricings", "supplier_cost_pricing_extended",
		"user_group_pricings", "user_group_pricing_extended",
		"external_models", "model_mappings",
		"user_api_keys",
		"usage_records",
		"membership_tiers", "membership_discounts", "user_memberships",
		"bills", "bill_items",
	); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	// Close and reopen connection to clear any cached state
	sqlDB.Close()
	return nil
}

// TruncateDB truncates all tables (faster than DropTable for test isolation)
func TruncateDB(t *testing.T, db *gorm.DB) error {
	tables := []string{
		"bill_items", "bills",
		"user_memberships", "membership_discounts", "membership_tiers",
		"usage_records",
		"user_api_keys",
		"model_mappings", "external_models",
		"user_group_pricing_extended", "user_group_pricings",
		"supplier_cost_pricing_extended", "supplier_cost_pricings", "suppliers",
		"user_groups", "admins", "users",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			// Table might not exist, continue
			continue
		}
	}

	return nil
}
