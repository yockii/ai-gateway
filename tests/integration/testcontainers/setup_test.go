package testcontainers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupPostgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	connStr, _ := SetupPostgres(t)

	assert.NotEmpty(t, connStr, "Connection string should not be empty")
	assert.Contains(t, connStr, "postgres", "Connection string should contain postgres")
	assert.Contains(t, connStr, "test_", "Connection string should contain test database name")

	// Cleanup is called automatically by t.Cleanup
}

func TestSetupRedis(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	connStr, _ := SetupRedis(t)

	assert.NotEmpty(t, connStr, "Connection string should not be empty")
	assert.Contains(t, connStr, "localhost", "Connection string should contain localhost")

	// Cleanup is called automatically by t.Cleanup
}

func TestSetupFullStack(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	db, redisClient, _ := SetupFullStack(t)

	assert.NotNil(t, db, "DB should not be nil")
	assert.NotNil(t, redisClient, "Redis client should not be nil")

	// Test DB connection
	sqlDB, err := db.DB()
	assert.NoError(t, err, "Should be able to get SQL DB")
	assert.NoError(t, sqlDB.Ping(), "Should be able to ping database")

	// Test Redis connection
	ctx := t.Context()
	assert.NoError(t, redisClient.Ping(ctx).Err(), "Should be able to ping Redis")

	// Cleanup is called automatically by t.Cleanup
}
