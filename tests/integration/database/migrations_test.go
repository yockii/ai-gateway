package database

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/yockii/ai-gateway/tests/integration/testcontainers"
	"gorm.io/gorm"
)

// setupDBTest creates a test database for migration testing
func setupDBTest(t *testing.T) *gorm.DB {
	gormDB, _, _ := testcontainers.SetupFullStack(t)
	t.Cleanup(func() {
		// Containers are cleaned up by testcontainers
	})
	return gormDB
}

func TestAutoMigrate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupDBTest(t)

	// Wrap in database.DB
	testDB := &database.DB{DB: db}

	// Run AutoMigrate
	err := testDB.AutoMigrate()
	require.NoError(t, err, "AutoMigrate should complete without errors")

	// Verify all tables exist
	tables := []string{
		"users",
		"admins",
		"user_groups",
		"suppliers",
		"supplier_cost_pricings",
		"supplier_cost_pricing_extended",
		"user_group_pricings",
		"user_group_pricing_extended",
		"external_models",
		"model_mappings",
		"user_api_keys",
		"usage_records",
		"membership_tiers",
		"membership_discounts",
		"user_memberships",
		"bills",
		"bill_items",
	}

	for _, table := range tables {
		// Check if table exists by attempting to query it
		err := db.Raw(fmt.Sprintf("SELECT 1 FROM %s LIMIT 1", table)).Error
		// We expect an error if the table is empty, but not a "does not exist" error
		if err != nil {
			// Check if it's a "does not exist" error
			if fmt.Sprintf("%v", err) == "ERROR: relation \""+table+"\" does not exist (SQLSTATE 42P01)" ||
				fmt.Sprintf("%v", err) == "ERROR: relation \""+table+"\" does not exist" {
				t.Errorf("Table %s does not exist after migration", table)
			}
		}
	}
}

func TestForeignKeyConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupDBTest(t)

	// Run migrations
	err := db.AutoMigrate(&models.User{}, &models.UserAPIKey{}, &models.UsageRecord{})
	require.NoError(t, err)

	// Create a test user
	user := &models.User{
		ID:       "test-user-1",
		Email:    "test@example.com",
		Password: "hashed",
		Name:     "Test User",
		IsActive: true,
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	// Test creating a UserAPIKey with valid user ID
	apiKey := &models.UserAPIKey{
		ID:       "test-key-1",
		UserID:   "test-user-1",
		KeyValue: "sk-test-key",
		Name:     "Test Key",
		IsActive: true,
	}
	err = db.Create(apiKey).Error
	require.NoError(t, err, "Should create API key with valid user ID")

	// Test creating a UserAPIKey with invalid user ID
	invalidAPIKey := &models.UserAPIKey{
		ID:       "test-key-2",
		UserID:   "non-existent-user",
		KeyValue: "sk-invalid-key",
		Name:     "Invalid Key",
		IsActive: true,
	}
	// Note: Foreign key constraints are disabled in production (DisableForeignKeyConstraintWhenMigrating: true)
	// So this test documents the current behavior
	err = db.Create(invalidAPIKey).Error
	// With foreign keys disabled, this will succeed
	// If foreign keys were enabled, this would fail
	assert.NoError(t, err, "With foreign keys disabled, invalid user ID is allowed")
}

func TestIndexes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupDBTest(t)

	// Run migrations
	err := db.AutoMigrate(&models.User{}, &models.ExternalModel{}, &models.UsageRecord{})
	require.NoError(t, err)

	// Create wrapper to use createIndexes
	testDB := &database.DB{DB: db}
	err = testDB.AutoMigrate()
	require.NoError(t, err)

	// Check that indexes exist by querying pg_indexes
	var indexes []string

	// Check users email index (from uniqueIndex tag)
	err = db.Raw(`
		SELECT indexname
		FROM pg_indexes
		WHERE tablename = 'users' AND indexname LIKE '%email%'
	`).Scan(&indexes).Error
	require.NoError(t, err)
	assert.NotEmpty(t, indexes, "Users table should have email index")

	// Check usage_records composite index
	err = db.Raw(`
		SELECT indexname
		FROM pg_indexes
		WHERE tablename = 'usage_records' AND indexname LIKE '%user_created%'
	`).Scan(&indexes).Error
	require.NoError(t, err)
	assert.NotEmpty(t, indexes, "Usage records table should have user_created index")
}

func TestUniqueConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupDBTest(t)

	// Run migrations
	err := db.AutoMigrate(&models.User{}, &models.Admin{}, &models.ExternalModel{})
	require.NoError(t, err)

	// Test user email unique constraint
	user1 := &models.User{
		ID:       "test-user-1",
		Email:    "unique@example.com",
		Password: "hashed",
		Name:     "User 1",
		IsActive: true,
	}
	err = db.Create(user1).Error
	require.NoError(t, err)

	user2 := &models.User{
		ID:       "test-user-2",
		Email:    "unique@example.com", // Duplicate email
		Password: "hashed",
		Name:     "User 2",
		IsActive: true,
	}
	err = db.Create(user2).Error
	assert.Error(t, err, "Duplicate email should fail unique constraint")

	// Test admin email unique constraint
	admin1 := &models.Admin{
		ID:       "test-admin-1",
		Email:    "admin@example.com",
		Password: "hashed",
		Name:     "Admin 1",
		IsActive: true,
	}
	err = db.Create(admin1).Error
	require.NoError(t, err)

	admin2 := &models.Admin{
		ID:       "test-admin-2",
		Email:    "admin@example.com", // Duplicate email
		Password: "hashed",
		Name:     "Admin 2",
		IsActive: true,
	}
	err = db.Create(admin2).Error
	assert.Error(t, err, "Duplicate admin email should fail unique constraint")

	// Test external model name unique constraint
	model1 := &models.ExternalModel{
		ID:          "test-model-1",
		Name:        "gpt-4",
		DisplayName: "GPT-4",
		ModelType:   "chat",
		IsActive:    true,
	}
	err = db.Create(model1).Error
	require.NoError(t, err)

	model2 := &models.ExternalModel{
		ID:          "test-model-2",
		Name:        "gpt-4", // Duplicate name
		DisplayName: "GPT-4 Copy",
		ModelType:   "chat",
		IsActive:    true,
	}
	err = db.Create(model2).Error
	assert.Error(t, err, "Duplicate model name should fail unique constraint")
}

func TestModelTimestamps(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupDBTest(t)

	// Run migrations
	err := db.AutoMigrate(&models.User{})
	require.NoError(t, err)

	// Create a user
	user := &models.User{
		ID:       "test-user-1",
		Email:    "timestamp@example.com",
		Password: "hashed",
		Name:     "Test User",
		IsActive: true,
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	// Verify CreatedAt and UpdatedAt are set
	assert.NotZero(t, user.CreatedAt, "CreatedAt should be set")
	assert.NotZero(t, user.UpdatedAt, "UpdatedAt should be set")

	// Update the user
	user.Name = "Updated User"
	err = db.Save(user).Error
	require.NoError(t, err)

	// Verify UpdatedAt was updated
	var fetchedUser models.User
	err = db.First(&fetchedUser, "id = ?", user.ID).Error
	require.NoError(t, err)

	assert.True(t, fetchedUser.UpdatedAt.After(user.CreatedAt), "UpdatedAt should be after CreatedAt")
}
