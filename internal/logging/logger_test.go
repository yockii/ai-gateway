package logging

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

// TestInitLogger tests logger initialization
func TestInitLogger(t *testing.T) {
	// Test development mode
	err := InitLogger("development")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	if logger == nil {
		t.Error("Expected logger to be initialized")
	}

	if sugar == nil {
		t.Error("Expected sugar logger to be initialized")
	}
}

// TestRequestID tests request ID handling
func TestRequestID(t *testing.T) {
	ctx := context.Background()

	// Test adding request ID
	ctx = WithRequestID(ctx)
	reqID := GetRequestID(ctx)
	if reqID == "" {
		t.Error("Expected request ID to be generated")
	}

	// Test getting existing request ID
	existingReqID := "test-req-123"
	ctx = context.WithValue(ctx, "request_id", existingReqID)
	retrievedReqID := GetRequestID(ctx)
	if retrievedReqID != existingReqID {
		t.Errorf("Expected %s, got %s", existingReqID, retrievedReqID)
	}
}

// TestLogFunctions tests various logging functions
func TestLogFunctions(t *testing.T) {
	// Initialize logger
	if err := InitLogger("development"); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Test Info
	Info("Test info message", zap.String("key", "value"))

	// Test Error
	Error("Test error message", zap.String("error", "test error"))

	// Test Warn
	Warn("Test warning message", zap.String("warning", "test warning"))

	// Test Debug
	Debug("Test debug message", zap.String("debug", "test debug"))

	// Test LogWithFields
	ctx := context.Background()
	ctx = WithRequestID(ctx)
	LogWithFields(ctx, "Test message with fields", zap.String("test", "value"))
}

// TestSync tests logger sync
func TestSync(t *testing.T) {
	if err := InitLogger("development"); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	err := Sync()
	if err != nil {
		t.Errorf("Expected sync to succeed, got error: %v", err)
	}
}
