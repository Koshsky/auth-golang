package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestNewLogger(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
		LogLevel:    slog.LevelInfo,
		Environment: "test",
		Version:     "1.0.0",
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	if logger == nil {
		t.Error("NewLogger() returned nil logger")
	}
}

func TestNewLogger_InvalidConfig(t *testing.T) {
	config := Config{
		// Missing ServiceName
		LogLevel:    slog.LevelInfo,
		Environment: "test",
		Version:     "1.0.0",
	}

	_, err := NewLogger(config)
	if err == nil {
		t.Error("NewLogger() should return error for invalid config")
	}
}

func TestContextLogging(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		ServiceName: "test-service",
		LogLevel:    slog.LevelInfo,
		Environment: "test",
		Version:     "1.0.0",
		Output:      &buf,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	// Create context with logging data
	logCtx := &LogCtx{
		UserID:    123,
		RequestID: "req-456",
		Email:     "user@example.com",
		Operation: "test",
	}

	ctx := WithLogCtx(context.Background(), logCtx)

	// Log a message
	logger.InfoContext(ctx, "Test message")

	// Parse output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Check if context attributes are present
	if logEntry["user_id"] != float64(123) {
		t.Errorf("Expected user_id=123, got %v", logEntry["user_id"])
	}
	if logEntry["request_id"] != "req-456" {
		t.Errorf("Expected request_id=req-456, got %v", logEntry["request_id"])
	}

	// Check if email is masked
	if logEntry["email"] != "u***@example.com" {
		t.Errorf("Expected masked email, got %v", logEntry["email"])
	}

	// Check service metadata
	if logEntry["service"] != "test-service" {
		t.Errorf("Expected service=test-service, got %v", logEntry["service"])
	}
}

func TestWithLogCtx(t *testing.T) {
	logCtx := &LogCtx{
		UserID:    123,
		RequestID: "req-456",
	}

	ctx := WithLogCtx(context.Background(), logCtx)

	// Retrieve the context value
	retrievedCtx := ctx.Value("log_ctx")
	if retrievedCtx == nil {
		t.Error("LogCtx not found in context")
	}

	retrievedLogCtx, ok := retrievedCtx.(*LogCtx)
	if !ok {
		t.Error("Retrieved context is not of type *LogCtx")
	}

	if retrievedLogCtx.UserID != 123 {
		t.Errorf("Expected UserID 123, got %v", retrievedLogCtx.UserID)
	}
}
