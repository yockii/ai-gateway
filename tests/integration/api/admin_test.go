package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yockii/ai-gateway/internal/models"
	"github.com/yockii/ai-gateway/pkg/api"
	"github.com/yockii/ai-gateway/pkg/handlers"
	"github.com/yockii/ai-gateway/tests/integration/testcontainers"
	"gorm.io/gorm"
)

// setupAdminTestApp creates a test Fiber app for admin testing
func setupAdminTestApp(t *testing.T) (*fiber.App, *gorm.DB) {
	gormDB, _, _ := testcontainers.SetupFullStack(t)

	// Run AutoMigrate for all models
	require.NoError(t, gormDB.AutoMigrate(&models.User{}, &models.Admin{}, &models.ExternalModel{}, &models.Supplier{}))

	gw := createTestGateway(t, gormDB)
	app := fiber.New(fiber.Config{
		AppName: "AI Gateway Admin Test",
	})

	handler := handlers.New(gw)

	// Set up admin routes with simulated auth
	app.Post("/v1/admin/models", func(c fiber.Ctx) error {
		// Simulate admin auth - check X-User-Role header
		role := c.Get("X-User-Role", "")
		if role != "admin" {
			return c.Status(http.StatusForbidden).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Admin access required",
					Type:    "permission_error",
					Code:    http.StatusForbidden,
				},
			})
		}
		return handler.CreateModel(c)
	})

	app.Put("/v1/admin/models/:id", func(c fiber.Ctx) error {
		role := c.Get("X-User-Role", "")
		if role != "admin" {
			return c.Status(http.StatusForbidden).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Admin access required",
					Type:    "permission_error",
					Code:    http.StatusForbidden,
				},
			})
		}
		return handler.UpdateModel(c)
	})

	app.Delete("/v1/admin/models/:id", func(c fiber.Ctx) error {
		role := c.Get("X-User-Role", "")
		if role != "admin" {
			return c.Status(http.StatusForbidden).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Admin access required",
					Type:    "permission_error",
					Code:    http.StatusForbidden,
				},
			})
		}
		return handler.DeleteModel(c)
	})

	app.Get("/v1/admin/models", func(c fiber.Ctx) error {
		role := c.Get("X-User-Role", "")
		if role != "admin" {
			return c.Status(http.StatusForbidden).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Admin access required",
					Type:    "permission_error",
					Code:    http.StatusForbidden,
				},
			})
		}
		return handler.AdminListModels(c)
	})

	app.Post("/v1/admin/suppliers", func(c fiber.Ctx) error {
		role := c.Get("X-User-Role", "")
		if role != "admin" {
			return c.Status(http.StatusForbidden).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Admin access required",
					Type:    "permission_error",
					Code:    http.StatusForbidden,
				},
			})
		}
		return handler.CreateSupplier(c)
	})

	app.Get("/v1/admin/suppliers", func(c fiber.Ctx) error {
		role := c.Get("X-User-Role", "")
		if role != "admin" {
			return c.Status(http.StatusForbidden).JSON(api.ErrorResponse{
				Error: api.ErrorDetail{
					Message: "Admin access required",
					Type:    "permission_error",
					Code:    http.StatusForbidden,
				},
			})
		}
		return handler.ListSuppliers(c)
	})

	t.Cleanup(func() {
		app.Shutdown()
	})

	return app, gormDB
}

func TestAdminLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	app, db := setupAdminTestApp(t)

	// Create test admin
	adminID := uuid.New().String()
	admin := &models.Admin{
		ID:       adminID,
		Email:    fmt.Sprintf("admin-%s@example.com", adminID),
		Password: "hashed_password",
		Name:     "Test Admin",
		IsActive: true,
	}
	require.NoError(t, db.Create(admin).Error)

	// Test admin access with proper role
	req := httptest.NewRequest("GET", "/v1/admin/models", nil)
	req.Header.Set("X-User-Role", "admin")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUnauthorizedAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	app, _ := setupAdminTestApp(t)

	// Test without admin role
	req := httptest.NewRequest("GET", "/v1/admin/models", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestUserCannotAccessAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	app, db := setupAdminTestApp(t)

	// Create regular user
	userID := uuid.New().String()
	user := &models.User{
		ID:       userID,
		Email:    fmt.Sprintf("user-%s@example.com", userID),
		Password: "hashed_password",
		Name:     "Regular User",
		IsActive: true,
	}
	require.NoError(t, db.Create(user).Error)

	// Test regular user trying to access admin endpoints
	req := httptest.NewRequest("GET", "/v1/admin/models", nil)
	req.Header.Set("X-User-Role", "user") // Not admin

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestModelManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	app, db := setupAdminTestApp(t)

	// Create test admin
	adminID := uuid.New().String()
	admin := &models.Admin{
		ID:       adminID,
		Email:    fmt.Sprintf("admin-%s@example.com", adminID),
		Password: "hashed_password",
		Name:     "Test Admin",
		IsActive: true,
	}
	require.NoError(t, db.Create(admin).Error)

	// Test creating a model
	createReq := api.CreateModelRequest{
		Name:        "gpt-4-test",
		DisplayName: "GPT-4 Test",
		ModelType:   "chat",
		Capabilities: []string{"chat", "streaming"},
		IsActive:    true,
	}

	bodyBytes, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/v1/admin/models", bytes.NewReader(bodyBytes))
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result["id"], "model-")
	assert.Equal(t, "gpt-4-test", result["name"])
	assert.Equal(t, "GPT-4 Test", result["display_name"])
	assert.Equal(t, "chat", result["model_type"])
}

func TestSupplierManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	app, db := setupAdminTestApp(t)

	// Create test admin
	adminID := uuid.New().String()
	admin := &models.Admin{
		ID:       adminID,
		Email:    fmt.Sprintf("admin-%s@example.com", adminID),
		Password: "hashed_password",
		Name:     "Test Admin",
		IsActive: true,
	}
	require.NoError(t, db.Create(admin).Error)

	// Test creating a supplier
	createReq := api.CreateSupplierRequest{
		Name:        "openai-test",
		DisplayName: "OpenAI Test",
		Provider:    "openai",
		IsActive:    true,
	}

	bodyBytes, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/v1/admin/suppliers", bytes.NewReader(bodyBytes))
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result["id"], "supplier-")
	assert.Equal(t, "openai-test", result["name"])
	assert.Equal(t, "OpenAI Test", result["display_name"])
	assert.Equal(t, "openai", result["provider"])
}

func TestAdminListModelValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	app, db := setupAdminTestApp(t)

	// Create test admin
	adminID := uuid.New().String()
	admin := &models.Admin{
		ID:       adminID,
		Email:    fmt.Sprintf("admin-%s@example.com", adminID),
		Password: "hashed_password",
		Name:     "Test Admin",
		IsActive: true,
	}
	require.NoError(t, db.Create(admin).Error)

	tests := []struct {
		name       string
		request    api.CreateModelRequest
		wantStatus int
	}{
		{
			name: "missing name",
			request: api.CreateModelRequest{
				Name:        "",
				DisplayName: "Test Model",
				ModelType:   "chat",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid model type",
			request: api.CreateModelRequest{
				Name:        "test-model",
				DisplayName: "Test Model",
				ModelType:   "invalid-type",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid model",
			request: api.CreateModelRequest{
				Name:        "gpt-3.5-turbo-test",
				DisplayName: "GPT-3.5 Turbo Test",
				ModelType:   "chat",
			},
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/v1/admin/models", bytes.NewReader(bodyBytes))
			req.Header.Set("X-User-Role", "admin")
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.wantStatus, resp.StatusCode, "Expected status %d for %s", tt.wantStatus, tt.name)
		})
	}
}
