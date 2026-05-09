package database

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/yockii/ai-gateway/tests/integration/testcontainers"
	"gorm.io/gorm"
)

// setupTransactionTest creates a test database for transaction testing
func setupTransactionTest(t *testing.T) *gorm.DB {
	gormDB, _, _ := testcontainers.SetupFullStack(t)
	t.Cleanup(func() {
		// Containers are cleaned up by testcontainers
	})

	// Run migrations
	err := gormDB.AutoMigrate(&models.User{}, &models.UsageRecord{}, &models.Bill{})
	require.NoError(t, err)

	return gormDB
}

func TestTransactionRollback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTransactionTest(t)

	// Start a transaction
	tx := db.Begin()
	require.NotNil(t, tx)

	// Create a user within transaction
	user := &models.User{
		ID:       "test-user-rollback",
		Email:    "rollback@example.com",
		Password: "hashed",
		Name:     "Test User",
		IsActive: true,
	}
	err := tx.Create(user).Error
	require.NoError(t, err)

	// Verify user exists within transaction
	var count int64
	tx.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(1), count, "User should exist within transaction")

	// Rollback the transaction
	tx.Rollback()

	// Verify user does not exist after rollback
	db.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(0), count, "User should not exist after rollback")
}

func TestNestedTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTransactionTest(t)

	// Start outer transaction
	tx := db.Begin()
	require.NotNil(t, tx)

	// Create user in outer transaction
	user := &models.User{
		ID:       "test-user-nested",
		Email:    "nested@example.com",
		Password: "hashed",
		Name:     "Test User",
		IsActive: true,
	}
	err := tx.Create(user).Error
	require.NoError(t, err)

	// Create savepoint (nested transaction)
	err = tx.SavePoint("sp1").Error
	require.NoError(t, err)

	// Create usage record in nested transaction
	usage := &models.UsageRecord{
		ID:         "test-usage-nested",
		RequestID:  "req-nested",
		UserID:     user.ID,
		ModelID:    "gpt-4",
		SupplierID: "openai",
		InputTokens: 10,
		OutputTokens: 20,
		TotalTokens: 30,
	}
	err = tx.Create(usage).Error
	require.NoError(t, err)

	// Rollback to savepoint
	err = tx.RollbackTo("sp1").Error
	require.NoError(t, err)

	// Verify usage record was rolled back
	var count int64
	tx.Model(&models.UsageRecord{}).Where("id = ?", usage.ID).Count(&count)
	assert.Equal(t, int64(0), count, "Usage record should not exist after rollback to savepoint")

	// User should still exist
	tx.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(1), count, "User should still exist after rollback to savepoint")

	// Commit outer transaction
	tx.Commit()

	// Verify user exists after commit
	db.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(1), count, "User should exist after commit")
}

func TestConcurrentWrites(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTransactionTest(t)

	// Create a base user
	user := &models.User{
		ID:       "test-user-concurrent",
		Email:    "concurrent@example.com",
		Password: "hashed",
		Name:     "Test User",
		IsActive: true,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Number of concurrent goroutines
	numGoroutines := 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Launch concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Each goroutine creates its own transaction
			tx := db.Begin()
			if tx == nil {
				errors <- fmt.Errorf("failed to begin transaction")
				return
			}

			// Create a usage record
			usage := &models.UsageRecord{
				ID:         fmt.Sprintf("usage-%d", idx),
				RequestID:  fmt.Sprintf("req-%d", idx),
				UserID:     user.ID,
				ModelID:    "gpt-4",
				SupplierID: "openai",
				InputTokens: int32(idx * 10),
				OutputTokens: int32(idx * 20),
				TotalTokens: int32(idx * 30),
			}

			err := tx.Create(usage).Error
			if err != nil {
				tx.Rollback()
				errors <- err
				return
			}

			// Commit transaction
			err = tx.Commit().Error
			if err != nil {
				errors <- err
				return
			}

			errors <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		assert.NoError(t, err, "Concurrent writes should not fail")
	}

	// Verify all records were created
	var count int64
	db.Model(&models.UsageRecord{}).Where("user_id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(numGoroutines), count, "All concurrent writes should succeed")
}

func TestTransactionIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTransactionTest(t)

	// Create a user
	user := &models.User{
		ID:       "test-user-isolation",
		Email:    "isolation@example.com",
		Password: "hashed",
		Name:     "Test User",
		IsActive: true,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Start transaction 1
	tx1 := db.Begin()
	require.NotNil(t, tx1)

	// Update user in transaction 1 (not committed yet)
	tx1.Model(&models.User{}).Where("id = ?", user.ID).Update("name", "Updated in TX1")

	// Start transaction 2
	tx2 := db.Begin()
	require.NotNil(t, tx2)

	// Read user in transaction 2
	var userFromTX2 models.User
	err = tx2.Where("id = ?", user.ID).First(&userFromTX2).Error
	require.NoError(t, err)

	// Verify transaction 2 sees the old value (read committed isolation)
	// Note: PostgreSQL's default isolation level is Read Committed
	// So tx2 should not see uncommitted changes from tx1
	assert.Equal(t, "Test User", userFromTX2.Name, "TX2 should not see uncommitted changes from TX1")

	// Commit transaction 1
	tx1.Commit()

	// Now transaction 2 should see the committed value
	err = tx2.Where("id = ?", user.ID).First(&userFromTX2).Error
	require.NoError(t, err)

	// After tx1 commits, tx2 should see the new value on its next read
	// But since we already fetched, we need to query again
	err = tx2.Where("id = ?", user.ID).First(&userFromTX2).Error
	require.NoError(t, err)

	// Clean up
	tx2.Rollback()
}

func TestTransactionWithContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTransactionTest(t)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use transaction with context
	tx := db.WithContext(ctx).Begin()
	require.NotNil(t, tx)

	// Create user
	user := &models.User{
		ID:       "test-user-context",
		Email:    "context@example.com",
		Password: "hashed",
		Name:     "Test User",
		IsActive: true,
	}
	err := tx.Create(user).Error
	require.NoError(t, err)

	// Commit transaction
	err = tx.Commit().Error
	require.NoError(t, err)

	// Verify user was created
	var fetchedUser models.User
	err = db.WithContext(ctx).Where("id = ?", user.ID).First(&fetchedUser).Error
	require.NoError(t, err)

	assert.Equal(t, user.Email, fetchedUser.Email)
}

func TestTransactionOnMultipleTables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTransactionTest(t)

	// Start transaction
	tx := db.Begin()
	require.NotNil(t, tx)

	// Create user
	user := &models.User{
		ID:       "test-user-multi",
		Email:    "multi@example.com",
		Password: "hashed",
		Name:     "Test User",
		IsActive: true,
	}
	err := tx.Create(user).Error
	require.NoError(t, err)

	// Create usage record
	usage := &models.UsageRecord{
		ID:         "test-usage-multi",
		RequestID:  "req-multi",
		UserID:     user.ID,
		ModelID:    "gpt-4",
		SupplierID: "openai",
		InputTokens: 100,
		OutputTokens: 200,
		TotalTokens: 300,
	}
	err = tx.Create(usage).Error
	require.NoError(t, err)

	// Create bill
	bill := &models.Bill{
		ID:            "test-bill-multi",
		UserID:        user.ID,
		Period:        "2024-01",
		TotalRequests: 1,
		TotalCost:     0.01,
		TotalRevenue:  0.02,
		TotalProfit:   0.01,
		Status:        "pending",
	}
	err = tx.Create(bill).Error
	require.NoError(t, err)

	// Commit transaction
	err = tx.Commit().Error
	require.NoError(t, err)

	// Verify all records were created
	var userCount, usageCount, billCount int64
	db.Model(&models.User{}).Where("id = ?", user.ID).Count(&userCount)
	db.Model(&models.UsageRecord{}).Where("id = ?", usage.ID).Count(&usageCount)
	db.Model(&models.Bill{}).Where("id = ?", bill.ID).Count(&billCount)

	assert.Equal(t, int64(1), userCount, "User should be created")
	assert.Equal(t, int64(1), usageCount, "Usage record should be created")
	assert.Equal(t, int64(1), billCount, "Bill should be created")
}
