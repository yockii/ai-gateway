package supplier

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yockii/ai-gateway/internal/crypto"
	"github.com/yockii/ai-gateway/internal/database"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/yockii/ai-gateway/internal/services"
)

func TestEncryptionService(t *testing.T) {
	cryptoSvc, err := crypto.NewEncryptionService()
	assert.NoError(t, err)

	plaintext := "sk-test-key-12345678901234567890"

	// Test encrypt
	ciphertext, err := cryptoSvc.Encrypt(plaintext)
	assert.NoError(t, err)
	assert.NotEmpty(t, ciphertext)
	assert.NotEqual(t, plaintext, ciphertext)

	// Test decrypt
	decrypted, err := cryptoSvc.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestCreateApiKey(t *testing.T) {
	// This test requires a database connection
	// Skip in CI/CD if no database available
	db, err := database.New("host=localhost port=5432 user=postgres password=postgres dbname=ai_gateway sslmode=disable")
	if err != nil {
		t.Skip("Database not available")
		return
	}

	ctx := context.Background()
	svc, _ := services.NewSupplierApiKeyService(db)

	key := &models.SupplierApiKey{
		SupplierID:  "test-supplier",
		Name:        "Test Key",
		Priority:    1,
		IsActive:    true,
	}

	err = svc.CreateApiKey(ctx, key, "sk-test-key-12345")
	assert.NoError(t, err)
	assert.NotEmpty(t, key.ID)
	assert.NotEmpty(t, key.KeyPrefix)
	assert.NotEmpty(t, key.KeyValueEncrypted)
}

func TestGetBestApiKey(t *testing.T) {
	db, err := database.New("host=localhost port=5432 user=postgres password=postgres dbname=ai_gateway sslmode=disable")
	if err != nil {
		t.Skip("Database not available")
		return
	}

	ctx := context.Background()
	svc, _ := services.NewSupplierApiKeyService(db)

	// Test with non-existent supplier
	_, plaintext, err := svc.GetBestApiKey(ctx, "non-existent")
	assert.Error(t, err)
	assert.Empty(t, plaintext)
}

func TestKeyPrefix(t *testing.T) {
	apiKey := "sk-test-key-12345678901234567890"
	prefix := crypto.GenerateKeyPrefix(apiKey)
	
	assert.Contains(t, prefix, "sk-****")
	assert.Contains(t, prefix, "7890")
}

func TestShouldRotate(t *testing.T) {
	now := time.Now().UTC()
	
	tests := []struct {
		name     string
		key      models.SupplierApiKey
		expected bool
	}{
		{
			name: "Active key within limits",
			key: models.SupplierApiKey{
				MaxRequests:     1000,
				CurrentRequests: 500,
				ExpireAt:        &[]time.Time{now.Add(24 * time.Hour)}[0],
			},
			expected: false,
		},
		{
			name: "Expired key",
			key: models.SupplierApiKey{
				MaxRequests:     1000,
				CurrentRequests: 500,
				ExpireAt:        &[]time.Time{now.Add(-1 * time.Hour)}[0],
			},
			expected: true,
		},
		{
			name: "Key at request limit",
			key: models.SupplierApiKey{
				MaxRequests:     1000,
				CurrentRequests: 1000,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.key.ShouldRotate()
			assert.Equal(t, tt.expected, result)
		})
	}
}
