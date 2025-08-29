package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/Koshsky/subs-service/auth-service/internal/config"
	"github.com/stretchr/testify/suite"
)

type LoggingTestSuite struct {
	suite.Suite
	validConfig config.LogConfig
	buffer      *bytes.Buffer
	ctx         context.Context
}

func (s *LoggingTestSuite) SetupTest() {
	s.buffer = &bytes.Buffer{}
	s.ctx = context.Background()
	s.validConfig = config.LogConfig{
		ServiceName: "test-service",
		LogLevel:    "INFO",
		Environment: "test",
		Version:     "1.0.0",
	}
	ResetGlobalLogger()
}

func (s *LoggingTestSuite) parseLogOutput() map[string]interface{} {
	var logEntry map[string]interface{}
	err := json.Unmarshal(s.buffer.Bytes(), &logEntry)
	s.Require().NoError(err)
	return logEntry
}

func (s *LoggingTestSuite) createSampleLogCtx() *LogCtx {
	return &LogCtx{
		UserID:    123,
		RequestID: "req-456",
		Email:     "user@example.com",
		Operation: "test",
		TraceID:   "trace-789",
	}
}

func (s *LoggingTestSuite) initLogging() {
	err := InitLoggingWithOutput(s.validConfig, s.buffer)
	s.NoError(err)
}

func (s *LoggingTestSuite) getLogCtxFromContext(ctx context.Context) *LogCtx {
	logCtx, ok := ctx.Value("log_ctx").(*LogCtx)
	s.True(ok)
	return logCtx
}

func (s *LoggingTestSuite) assertLogCtxFields(logCtx *LogCtx, expected map[string]interface{}) {
	if userID, ok := expected["user_id"]; ok {
		s.Equal(userID, logCtx.UserID)
	}
	if requestID, ok := expected["request_id"]; ok {
		s.Equal(requestID, logCtx.RequestID)
	}
	if email, ok := expected["email"]; ok {
		s.Equal(email, logCtx.Email)
	}
	if operation, ok := expected["operation"]; ok {
		s.Equal(operation, logCtx.Operation)
	}
	if traceID, ok := expected["trace_id"]; ok {
		s.Equal(traceID, logCtx.TraceID)
	}

}

func (s *LoggingTestSuite) TestInitLogging_ValidConfig() {
	s.initLogging()
	slog.Info("test message")

	logEntry := s.parseLogOutput()
	s.Equal("test message", logEntry["message"])
	s.Equal("INFO", logEntry["level"])
}

func (s *LoggingTestSuite) TestInitLogging_InvalidConfig() {
	invalidConfig := config.LogConfig{
		LogLevel:    "INFO",
		Environment: "test",
		Version:     "1.0.0",
	}

	err := InitLoggingWithOutput(invalidConfig, s.buffer)
	s.Error(err)
	s.Contains(err.Error(), "ServiceName is required")
}

func (s *LoggingTestSuite) TestInitLogging_CalledTwice() {
	s.initLogging()
	slog.Info("first message")
	firstOutput := s.buffer.String()
	s.buffer.Reset()

	differentConfig := config.LogConfig{
		ServiceName: "different-service",
		LogLevel:    "DEBUG",
		Environment: "different",
		Version:     "2.0.0",
	}

	err := InitLoggingWithOutput(differentConfig, &bytes.Buffer{})
	s.NoError(err)

	slog.Info("second message")
	secondOutput := s.buffer.String()
	s.NotEmpty(firstOutput)
	s.NotEmpty(secondOutput)

	var firstEntry, secondEntry map[string]interface{}
	json.Unmarshal([]byte(firstOutput), &firstEntry)
	json.Unmarshal([]byte(secondOutput), &secondEntry)
	s.Equal("test-service", firstEntry["service"])
	s.Equal("test-service", secondEntry["service"])
}

func (s *LoggingTestSuite) TestCreateLogger_ValidConfig() {
	logger, err := createLogger(s.validConfig, s.buffer)
	s.NoError(err)
	s.NotNil(logger)
}

func (s *LoggingTestSuite) TestCreateLogger_InvalidConfig() {
	invalidConfig := config.LogConfig{
		LogLevel:    "INFO",
		Environment: "test",
		Version:     "1.0.0",
	}

	logger, err := createLogger(invalidConfig, s.buffer)
	s.Error(err)
	s.Nil(logger)
	s.Contains(err.Error(), "ServiceName is required")
}

func (s *LoggingTestSuite) TestCreateLogger_NilOutput() {
	logger, err := createLogger(s.validConfig, nil)
	s.NoError(err)
	s.NotNil(logger)
}

func (s *LoggingTestSuite) TestResetGlobalLogger() {
	s.initLogging()
	slog.Info("before reset")
	s.NotEmpty(s.buffer.String())

	ResetGlobalLogger()

	newBuffer := &bytes.Buffer{}
	newConfig := config.LogConfig{
		ServiceName: "new-service",
		LogLevel:    "DEBUG",
		Environment: "new-env",
		Version:     "2.0.0",
	}

	err := InitLoggingWithOutput(newConfig, newBuffer)
	s.NoError(err)

	slog.Info("after reset")
	var afterEntry map[string]interface{}
	json.Unmarshal(newBuffer.Bytes(), &afterEntry)

	s.Equal("new-service", afterEntry["service"])
	s.Equal("new-env", afterEntry["environment"])
	s.Equal("2.0.0", afterEntry["version"])
}

func (s *LoggingTestSuite) TestWithLogCtx() {
	logCtx := s.createSampleLogCtx()
	ctx := WithLogCtx(s.ctx, logCtx)

	retrievedLogCtx := s.getLogCtxFromContext(ctx)
	s.assertLogCtxFields(retrievedLogCtx, map[string]interface{}{
		"user_id":    123,
		"request_id": "req-456",
		"email":      "user@example.com",
		"operation":  "test",
		"trace_id":   "trace-789",
		"error":      "test error",
	})
}

func (s *LoggingTestSuite) TestWithLogCtx_NilLogCtx() {
	ctx := WithLogCtx(s.ctx, nil)
	s.Nil(ctx.Value("log_ctx"))
}

func (s *LoggingTestSuite) TestWithUserID() {
	ctx := WithUserID(s.ctx, 12345)
	logCtx := s.getLogCtxFromContext(ctx)
	s.assertLogCtxFields(logCtx, map[string]interface{}{"user_id": 12345})
	s.Nil(logCtx.RequestID)
	s.Nil(logCtx.Email)
}

func (s *LoggingTestSuite) TestWithUserID_UpdateExistingContext() {
	existingLogCtx := &LogCtx{RequestID: "existing-req"}
	ctx := WithLogCtx(s.ctx, existingLogCtx)
	ctx = WithUserID(ctx, 99999)

	logCtx := s.getLogCtxFromContext(ctx)
	s.assertLogCtxFields(logCtx, map[string]interface{}{
		"user_id":    99999,
		"request_id": "existing-req",
	})
}

func (s *LoggingTestSuite) TestWithRequestID() {
	ctx := WithRequestID(s.ctx, "req-test-123")
	logCtx := s.getLogCtxFromContext(ctx)
	s.assertLogCtxFields(logCtx, map[string]interface{}{"request_id": "req-test-123"})
	s.Nil(logCtx.UserID)
}

func (s *LoggingTestSuite) TestWithTraceID() {
	ctx := WithTraceID(s.ctx, "trace-xyz-789")
	logCtx := s.getLogCtxFromContext(ctx)
	s.assertLogCtxFields(logCtx, map[string]interface{}{"trace_id": "trace-xyz-789"})
}

func (s *LoggingTestSuite) TestWithOperation() {
	ctx := WithOperation(s.ctx, "user_authentication")
	logCtx := s.getLogCtxFromContext(ctx)
	s.assertLogCtxFields(logCtx, map[string]interface{}{"operation": "user_authentication"})
}

func (s *LoggingTestSuite) TestWithEmail() {
	ctx := WithEmail(s.ctx, "test@example.com")
	logCtx := s.getLogCtxFromContext(ctx)
	s.assertLogCtxFields(logCtx, map[string]interface{}{"email": "test@example.com"})
}

func (s *LoggingTestSuite) TestChainedContextMethods() {
	ctx := s.ctx
	ctx = WithUserID(ctx, 555)
	ctx = WithRequestID(ctx, "req-chain-test")
	ctx = WithOperation(ctx, "chained_operation")
	ctx = WithEmail(ctx, "chain@test.com")

	logCtx := s.getLogCtxFromContext(ctx)
	s.assertLogCtxFields(logCtx, map[string]interface{}{
		"user_id":    555,
		"request_id": "req-chain-test",
		"operation":  "chained_operation",
		"email":      "chain@test.com",
	})
	s.Nil(logCtx.TraceID)

}

func (s *LoggingTestSuite) TestGetOrCreateLogCtx_NewContext() {
	logCtx, newCtx := getOrCreateLogCtx(s.ctx)
	s.NotNil(logCtx)
	s.NotNil(newCtx)
	s.IsType(&LogCtx{}, logCtx)
	s.Equal(logCtx, newCtx.Value("log_ctx"))
}

func (s *LoggingTestSuite) TestGetOrCreateLogCtx_ExistingContext() {
	existingLogCtx := &LogCtx{UserID: 789}
	ctx := WithLogCtx(s.ctx, existingLogCtx)

	logCtx, newCtx := getOrCreateLogCtx(ctx)

	// Should return a copy, not the original pointer (to prevent data races)
	s.True(existingLogCtx != logCtx, "Should return a copy to prevent data races - different pointer addresses")
	s.NotEqual(ctx, newCtx, "Should return new context with copied LogCtx")
	s.Equal(789, logCtx.UserID, "Should preserve the UserID value")

	// Verify the copy has the same values but different pointer
	s.Equal(existingLogCtx.UserID, logCtx.UserID)
	s.Equal(existingLogCtx.RequestID, logCtx.RequestID)
	s.Equal(existingLogCtx.TraceID, logCtx.TraceID)
	s.Equal(existingLogCtx.Operation, logCtx.Operation)
	s.Equal(existingLogCtx.Email, logCtx.Email)
}

func (s *LoggingTestSuite) TestGetOrCreateLogCtx_NilExistingContext() {
	ctx := context.WithValue(s.ctx, "log_ctx", nil)
	logCtx, newCtx := getOrCreateLogCtx(ctx)
	s.NotNil(logCtx)
	s.IsType(&LogCtx{}, logCtx)
	s.Equal(logCtx, newCtx.Value("log_ctx"))
}

func (s *LoggingTestSuite) TestWithMultiple_AllFields() {
	fields := map[string]interface{}{
		"user_id":    123,
		"request_id": "req-multi-test",
		"trace_id":   "trace-multi-456",
		"operation":  "multi_operation",
		"error":      "multi error",
		"email":      "multi@test.com",
	}

	ctx := WithMultiple(s.ctx, fields)
	logCtx := s.getLogCtxFromContext(ctx)
	s.assertLogCtxFields(logCtx, fields)
}

func (s *LoggingTestSuite) TestWithMultiple_PartialFields() {
	fields := map[string]interface{}{
		"user_id":   456,
		"operation": "partial_op",
	}

	ctx := WithMultiple(s.ctx, fields)
	logCtx := s.getLogCtxFromContext(ctx)
	s.assertLogCtxFields(logCtx, fields)
	s.Nil(logCtx.RequestID)
	s.Nil(logCtx.TraceID)

	s.Nil(logCtx.Email)
}

func (s *LoggingTestSuite) TestWithMultiple_UnknownFields() {
	fields := map[string]interface{}{
		"user_id":       789,
		"unknown_field": "should_be_ignored",
		"invalid_key":   "also_ignored",
		"request_id":    "req-known",
	}

	ctx := WithMultiple(s.ctx, fields)
	logCtx := s.getLogCtxFromContext(ctx)
	s.assertLogCtxFields(logCtx, map[string]interface{}{
		"user_id":    789,
		"request_id": "req-known",
	})
	s.Nil(logCtx.TraceID)
	s.Nil(logCtx.Operation)

	s.Nil(logCtx.Email)
}

func (s *LoggingTestSuite) TestWithMultiple_EmptyFields() {
	ctx := WithMultiple(s.ctx, map[string]interface{}{})
	logCtx := s.getLogCtxFromContext(ctx)

	s.Nil(logCtx.UserID)
	s.Nil(logCtx.RequestID)
	s.Nil(logCtx.TraceID)
	s.Nil(logCtx.Operation)

	s.Nil(logCtx.Email)
}

func (s *LoggingTestSuite) TestWithMultiple_UpdateExistingContext() {
	existingLogCtx := &LogCtx{
		UserID:    999,
		RequestID: "existing-req",
	}
	ctx := WithLogCtx(s.ctx, existingLogCtx)

	fields := map[string]interface{}{
		"user_id":   111,
		"operation": "new_operation",
		"email":     "new@email.com",
	}

	ctx = WithMultiple(ctx, fields)
	logCtx := s.getLogCtxFromContext(ctx)
	s.assertLogCtxFields(logCtx, map[string]interface{}{
		"user_id":    111,
		"request_id": "existing-req",
		"operation":  "new_operation",
		"email":      "new@email.com",
	})
}

func (s *LoggingTestSuite) TestNewContextHandler() {
	baseHandler := slog.NewJSONHandler(s.buffer, &slog.HandlerOptions{Level: slog.LevelInfo})
	handler := newContextHandler(baseHandler)
	s.NotNil(handler)
	s.IsType(&contextHandler{}, handler)
}

func (s *LoggingTestSuite) TestContextHandler_Enabled() {
	handler := &contextHandler{
		base: slog.NewJSONHandler(s.buffer, &slog.HandlerOptions{Level: slog.LevelInfo}),
	}

	s.True(handler.Enabled(s.ctx, slog.LevelInfo))
	s.True(handler.Enabled(s.ctx, slog.LevelError))
	s.False(handler.Enabled(s.ctx, slog.LevelDebug))
}

func (s *LoggingTestSuite) TestContextHandler_WithAttrs() {
	handler := &contextHandler{
		base: slog.NewJSONHandler(s.buffer, &slog.HandlerOptions{Level: slog.LevelInfo}),
	}

	newHandler := handler.WithAttrs([]slog.Attr{slog.String("test", "value")})
	s.NotNil(newHandler)
	s.IsType(&contextHandler{}, newHandler)
}

func (s *LoggingTestSuite) TestContextHandler_WithGroup() {
	handler := &contextHandler{
		base: slog.NewJSONHandler(s.buffer, &slog.HandlerOptions{Level: slog.LevelInfo}),
	}

	newHandler := handler.WithGroup("test-group")
	s.NotNil(newHandler)
	s.IsType(&contextHandler{}, newHandler)
}

func (s *LoggingTestSuite) TestExtractContextAttrs_AllFields() {
	logCtx := s.createSampleLogCtx()
	ctx := WithLogCtx(s.ctx, logCtx)

	attrs := extractContextAttrs(ctx)
	s.Len(attrs, 5)

	attrMap := make(map[string]interface{})
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	s.Equal(int64(123), attrMap["user_id"])
	s.Equal("req-456", attrMap["request_id"])
	s.Equal("trace-789", attrMap["trace_id"])
	s.Equal("test", attrMap["operation"])
	s.Equal("u***@example.com", attrMap["email"])
}

func (s *LoggingTestSuite) TestExtractContextAttrs_EmptyLogCtx() {
	emptyLogCtx := &LogCtx{}
	ctx := WithLogCtx(s.ctx, emptyLogCtx)

	attrs := extractContextAttrs(ctx)
	s.Len(attrs, 0)
}

func (s *LoggingTestSuite) TestExtractContextAttrs_NoLogCtx() {
	attrs := extractContextAttrs(s.ctx)
	s.Len(attrs, 0)
}

func (s *LoggingTestSuite) TestExtractContextAttrs_PartialFields() {
	partialLogCtx := &LogCtx{
		UserID: 999,
		Email:  "partial@test.com",
	}
	ctx := WithLogCtx(s.ctx, partialLogCtx)

	attrs := extractContextAttrs(ctx)
	s.Len(attrs, 2)

	attrMap := make(map[string]interface{})
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	s.Equal(int64(999), attrMap["user_id"])
	s.Equal("p******@test.com", attrMap["email"])
	s.NotContains(attrMap, "request_id")
	s.NotContains(attrMap, "trace_id")
	s.NotContains(attrMap, "operation")
	s.NotContains(attrMap, "error")
}

func (s *LoggingTestSuite) TestGetHostname() {
	hostname := getHostname()

	s.NotEmpty(hostname)
	// Should either be actual hostname or "unknown"
	if hostname != "unknown" {
		realHostname, err := os.Hostname()
		s.NoError(err)
		s.Equal(realHostname, hostname)
	}
}

func (s *LoggingTestSuite) TestContextMethods_WithNilValues() {
	ctx := s.ctx
	ctx = WithUserID(ctx, nil)
	ctx = WithRequestID(ctx, nil)
	ctx = WithEmail(ctx, nil)

	logCtx := s.getLogCtxFromContext(ctx)
	s.Nil(logCtx.UserID)
	s.Nil(logCtx.RequestID)
	s.Nil(logCtx.Email)
}

func (s *LoggingTestSuite) TestContextMethods_WithDifferentTypes() {
	ctx := s.ctx
	ctx = WithUserID(ctx, "string_user_id")
	ctx = WithRequestID(ctx, 12345)

	logCtx := s.getLogCtxFromContext(ctx)
	s.Equal("string_user_id", logCtx.UserID)
	s.Equal(12345, logCtx.RequestID)

}

func (s *LoggingTestSuite) TestWithMultiple_WithNilMap() {
	ctx := WithMultiple(s.ctx, nil)
	logCtx := s.getLogCtxFromContext(ctx)

	s.Nil(logCtx.UserID)
	s.Nil(logCtx.RequestID)
	s.Nil(logCtx.TraceID)
	s.Nil(logCtx.Operation)

	s.Nil(logCtx.Email)
}

func (s *LoggingTestSuite) TestContextModification_Immutability() {
	originalCtx := WithUserID(s.ctx, 123)
	newCtx := WithRequestID(originalCtx, "new-req")

	originalLogCtx, _ := originalCtx.Value("log_ctx").(*LogCtx)
	newLogCtx, _ := newCtx.Value("log_ctx").(*LogCtx)

	// Original context should not be modified
	s.Equal(123, originalLogCtx.UserID)
	s.Empty(originalLogCtx.RequestID) // Original should not have RequestID

	// New context should have both values
	s.Equal(123, newLogCtx.UserID)
	s.Equal("new-req", newLogCtx.RequestID)

	// LogCtx pointers should be different (no sharing)
	s.NotEqual(originalLogCtx, newLogCtx, "LogCtx should be copied to prevent data races")
}

func (s *LoggingTestSuite) TestConcurrentContextModification() {
	s.initLogging()

	baseCtx := WithUserID(s.ctx, "user123")

	done := make(chan bool, 2)

	// Goroutine 1: modifies request ID
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			ctx := WithRequestID(baseCtx, fmt.Sprintf("req-%d", i))
			slog.InfoContext(ctx, "Request processed")
		}
	}()

	// Goroutine 2: modifies trace ID
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			ctx := WithTraceID(baseCtx, fmt.Sprintf("trace-%d", i))
			slog.InfoContext(ctx, "Trace logged")
		}
	}()

	// Wait for completion
	<-done
	<-done

	s.NotEmpty(s.buffer.String())
}

// TestThreadSafetyDataRace tests that context operations are thread-safe
func (s *LoggingTestSuite) TestThreadSafetyDataRace() {
	s.initLogging()

	baseCtx := WithUserID(s.ctx, "base-user")

	// Create a shared buffer to collect race-free operations
	results := make(chan string, 200)
	done := make(chan bool, 2)

	// Goroutine 1: Creates many derived contexts and logs
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			ctx := WithRequestID(baseCtx, fmt.Sprintf("req-%d", i))
			ctx = WithOperation(ctx, fmt.Sprintf("op-%d", i))

			// Extract the context data to verify no corruption
			if logCtx, ok := ctx.Value("log_ctx").(*LogCtx); ok {
				results <- fmt.Sprintf("G1: user=%v req=%v op=%v",
					logCtx.UserID, logCtx.RequestID, logCtx.Operation)
			}
		}
	}()

	// Goroutine 2: Creates different derived contexts simultaneously
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			ctx := WithTraceID(baseCtx, fmt.Sprintf("trace-%d", i))
			ctx = WithEmail(ctx, fmt.Sprintf("user%d@test.com", i))

			// Extract the context data to verify no corruption
			if logCtx, ok := ctx.Value("log_ctx").(*LogCtx); ok {
				results <- fmt.Sprintf("G2: user=%v trace=%v email=%v",
					logCtx.UserID, logCtx.TraceID, logCtx.Email)
			}
		}
	}()

	// Wait for completion
	<-done
	<-done
	close(results)

	// Verify we got all expected results without corruption
	resultCount := 0
	for result := range results {
		resultCount++
		s.Contains(result, "user=base-user", "Base user should be preserved: %s", result)
		// If there was data corruption, we might see mixed values or nil pointers
	}

	s.Equal(200, resultCount, "Should have 200 race-free operations")
}

// TestContextImmutability tests that context modifications don't affect original context
func (s *LoggingTestSuite) TestContextImmutability() {
	s.initLogging()

	// Create base context with user ID
	baseCtx := WithUserID(s.ctx, "original-user")

	// Create derived context with different user ID
	derivedCtx := WithUserID(baseCtx, "derived-user")

	// Verify that LogCtx pointers are different (no sharing)
	baseLogCtx, _ := baseCtx.Value("log_ctx").(*LogCtx)
	derivedLogCtx, _ := derivedCtx.Value("log_ctx").(*LogCtx)

	s.NotEqual(baseLogCtx, derivedLogCtx, "LogCtx pointers should be different to prevent data races")

	// Log from base context
	s.buffer.Reset()
	slog.InfoContext(baseCtx, "Base context message")
	baseOutput := s.buffer.String()

	// Log from derived context
	s.buffer.Reset()
	slog.InfoContext(derivedCtx, "Derived context message")
	derivedOutput := s.buffer.String()

	// Parse JSON outputs
	var baseLog, derivedLog map[string]interface{}
	err := json.Unmarshal([]byte(baseOutput), &baseLog)
	s.NoError(err)
	err = json.Unmarshal([]byte(derivedOutput), &derivedLog)
	s.NoError(err)

	// Verify base context wasn't modified
	s.Equal("original-user", baseLog["user_id"])
	s.Equal("derived-user", derivedLog["user_id"])

	// They should be different
	s.NotEqual(baseLog["user_id"], derivedLog["user_id"])
}

func (s *LoggingTestSuite) TestLoggingIntegration_WithIndividualMethods() {
	s.initLogging()

	ctx := s.ctx
	ctx = WithUserID(ctx, 888)
	ctx = WithRequestID(ctx, "req-integration-test")
	ctx = WithOperation(ctx, "integration_operation")
	ctx = WithTraceID(ctx, "trace-integration-123")

	slog.InfoContext(ctx, "Integration test with individual methods")

	logEntry := s.parseLogOutput()
	s.Equal(float64(888), logEntry["user_id"])
	s.Equal("req-integration-test", logEntry["request_id"])
	s.Equal("integration_operation", logEntry["operation"])
	s.Equal("trace-integration-123", logEntry["trace_id"])
	s.Equal("Integration test with individual methods", logEntry["message"])
	s.Equal("INFO", logEntry["level"])
	s.Equal("test-service", logEntry["service"])
}

func (s *LoggingTestSuite) TestLoggingIntegration_WithMultipleMethod() {
	s.initLogging()

	fields := map[string]interface{}{
		"user_id":    777,
		"request_id": "req-multi-integration",
		"email":      "integration@test.com",
		"operation":  "multi_integration_test",
	}

	ctx := WithMultiple(s.ctx, fields)
	slog.WarnContext(ctx, "Integration test with WithMultiple method", "extra_field", "extra_value")

	logEntry := s.parseLogOutput()
	s.Equal(float64(777), logEntry["user_id"])
	s.Equal("req-multi-integration", logEntry["request_id"])
	s.Equal("i**********@test.com", logEntry["email"])
	s.Equal("multi_integration_test", logEntry["operation"])
	s.Equal("Integration test with WithMultiple method", logEntry["message"])
	s.Equal("WARN", logEntry["level"])
	s.Equal("extra_value", logEntry["extra_field"])
}

func (s *LoggingTestSuite) TestLoggingIntegration_MixedMethods() {
	s.initLogging()

	ctx := WithMultiple(s.ctx, map[string]interface{}{
		"user_id":    666,
		"request_id": "req-mixed-test",
	})

	ctx = WithOperation(ctx, "mixed_operation")
	ctx = WithEmail(ctx, "mixed@example.com")
	ctx = WithMultiple(ctx, map[string]interface{}{
		"trace_id": "trace-mixed-456",
	})

	slog.ErrorContext(ctx, "Integration test mixing methods")

	logEntry := s.parseLogOutput()
	s.Equal(float64(666), logEntry["user_id"])
	s.Equal("req-mixed-test", logEntry["request_id"])
	s.Equal("mixed_operation", logEntry["operation"])
	s.Equal("m****@example.com", logEntry["email"])
	s.Equal("trace-mixed-456", logEntry["trace_id"])
	s.Equal("Integration test mixing methods", logEntry["message"])
	s.Equal("ERROR", logEntry["level"])
}

func (s *LoggingTestSuite) TestLoggingIntegration_EmptyAndNilFields() {
	s.initLogging()

	ctx := s.ctx
	ctx = WithUserID(ctx, nil)
	ctx = WithRequestID(ctx, "")
	ctx = WithOperation(ctx, "test_op")

	slog.InfoContext(ctx, "Test with empty and nil fields")

	logEntry := s.parseLogOutput()
	s.NotContains(logEntry, "user_id")
	s.Contains(logEntry, "request_id")
	s.Equal("", logEntry["request_id"])
	s.Equal("test_op", logEntry["operation"])
}

func (s *LoggingTestSuite) TestContextLogging_WithLogCtx() {
	s.initLogging()

	logCtx := s.createSampleLogCtx()
	ctx := WithLogCtx(s.ctx, logCtx)

	slog.InfoContext(ctx, "Test message with context")

	logEntry := s.parseLogOutput()
	s.Equal(float64(123), logEntry["user_id"])
	s.Equal("req-456", logEntry["request_id"])
	s.Equal("u***@example.com", logEntry["email"])
	s.Equal("test", logEntry["operation"])
	s.Equal("trace-789", logEntry["trace_id"])
	s.Equal("test-service", logEntry["service"])
	s.Equal("test", logEntry["environment"])
	s.Equal("1.0.0", logEntry["version"])
	s.Contains(logEntry, "hostname")
	s.Equal("Test message with context", logEntry["message"])
	s.Equal("INFO", logEntry["level"])
	s.Contains(logEntry, "@timestamp")
}

func (s *LoggingTestSuite) TestContextLogging_EmptyLogCtx() {
	s.initLogging()

	emptyLogCtx := &LogCtx{}
	ctx := WithLogCtx(s.ctx, emptyLogCtx)

	slog.InfoContext(ctx, "Test with empty context")

	logEntry := s.parseLogOutput()
	s.NotContains(logEntry, "user_id")
	s.NotContains(logEntry, "email")
	s.NotContains(logEntry, "request_id")
	s.Equal("Test with empty context", logEntry["message"])
}

func (s *LoggingTestSuite) TestContextLogging_NoLogCtx() {
	s.initLogging()

	slog.InfoContext(s.ctx, "Test without context")

	logEntry := s.parseLogOutput()
	s.NotContains(logEntry, "user_id")
	s.NotContains(logEntry, "email")
	s.NotContains(logEntry, "request_id")
	s.Equal("Test without context", logEntry["message"])
}

func (s *LoggingTestSuite) TestSlogLevels() {
	debugConfig := s.validConfig
	debugConfig.LogLevel = "DEBUG"
	err := InitLoggingWithOutput(debugConfig, s.buffer)
	s.NoError(err)

	testCases := []struct {
		level   slog.Level
		logFunc func(string, ...any)
		name    string
	}{
		{slog.LevelDebug, slog.Debug, "DEBUG"},
		{slog.LevelInfo, slog.Info, "INFO"},
		{slog.LevelWarn, slog.Warn, "WARN"},
		{slog.LevelError, slog.Error, "ERROR"},
	}

	for _, tc := range testCases {
		s.buffer.Reset()
		tc.logFunc(tc.name+" message", "key", "value")

		logEntry := s.parseLogOutput()
		s.Equal(tc.name+" message", logEntry["message"])
		s.Equal(tc.name, logEntry["level"])
		s.Equal("value", logEntry["key"])
	}
}

func (s *LoggingTestSuite) TestSlogContextLevels() {
	s.initLogging()

	testCases := []struct {
		logFunc func(context.Context, string, ...any)
		name    string
		logCtx  *LogCtx
	}{
		{slog.InfoContext, "INFO", &LogCtx{UserID: 789}},
		{slog.WarnContext, "WARN", &LogCtx{Operation: "test-op"}},
		{slog.ErrorContext, "ERROR", &LogCtx{Operation: "error-op"}},
	}

	for _, tc := range testCases {
		s.buffer.Reset()
		ctx := WithLogCtx(s.ctx, tc.logCtx)
		tc.logFunc(ctx, tc.name+" message with context", "key", "value")

		logEntry := s.parseLogOutput()
		s.Equal(tc.name+" message with context", logEntry["message"])
		s.Equal(tc.name, logEntry["level"])
		s.Equal("value", logEntry["key"])
	}
}

func (s *LoggingTestSuite) TestSlogLog() {
	s.initLogging()

	ctx := WithLogCtx(s.ctx, &LogCtx{TraceID: "custom-trace"})
	slog.Log(ctx, slog.LevelWarn, "Custom log message", "custom", "attribute")

	logEntry := s.parseLogOutput()
	s.Equal("Custom log message", logEntry["message"])
	s.Equal("WARN", logEntry["level"])
	s.Equal("custom-trace", logEntry["trace_id"])
	s.Equal("attribute", logEntry["custom"])
}

func (s *LoggingTestSuite) TestJSONFormatting() {
	s.initLogging()

	slog.Info("Test JSON formatting", "string_field", "test", "int_field", 42, "bool_field", true)

	logEntry := s.parseLogOutput()
	s.Equal("test-service", logEntry["service"])
	s.Equal("test", logEntry["environment"])
	s.Equal("1.0.0", logEntry["version"])
	s.Contains(logEntry, "@timestamp")
	s.Equal("INFO", logEntry["level"])
	s.Equal("Test JSON formatting", logEntry["message"])
	s.Contains(logEntry, "source")
	s.Equal("test", logEntry["string_field"])
	s.Equal(float64(42), logEntry["int_field"])
	s.Equal(true, logEntry["bool_field"])
}

// TestParseLogLevel tests the parseLogLevel function
func (s *LoggingTestSuite) TestParseLogLevel() {
	testCases := []struct {
		input    string
		expected slog.Level
	}{
		{"DEBUG", slog.LevelDebug},
		{"debug", slog.LevelDebug},
		{"Debug", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{"info", slog.LevelInfo},
		{"Info", slog.LevelInfo},
		{"WARN", slog.LevelWarn},
		{"warn", slog.LevelWarn},
		{"WARNING", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"error", slog.LevelError},
		{"Error", slog.LevelError},
		{"INVALID", slog.LevelInfo}, // Default to INFO
		{"", slog.LevelInfo},        // Default to INFO
		{"random", slog.LevelInfo},  // Default to INFO
	}

	for _, tc := range testCases {
		result := parseLogLevel(tc.input)
		s.Equal(tc.expected, result, "Failed for input: %s", tc.input)
	}
}

// TestConfigIntegration tests integration with config package
func (s *LoggingTestSuite) TestConfigIntegration() {
	testCases := []struct {
		name        string
		configLevel string
		logFunc     func(string, ...any)
		shouldLog   bool
	}{
		{"DEBUG level allows all", "DEBUG", slog.Debug, true},
		{"INFO level blocks DEBUG", "INFO", slog.Debug, false},
		{"INFO level allows INFO", "INFO", slog.Info, true},
		{"WARN level blocks INFO", "WARN", slog.Info, false},
		{"WARN level allows WARN", "WARN", slog.Warn, true},
		{"ERROR level blocks WARN", "ERROR", slog.Warn, false},
		{"ERROR level allows ERROR", "ERROR", slog.Error, true},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ResetGlobalLogger()
			s.buffer.Reset()

			cfg := config.LogConfig{
				ServiceName: "test-service",
				LogLevel:    tc.configLevel,
				Environment: "test",
				Version:     "1.0.0",
			}

			err := InitLoggingWithOutput(cfg, s.buffer)
			s.NoError(err)

			tc.logFunc("test message")

			if tc.shouldLog {
				s.NotEmpty(s.buffer.String(), "Should have logged for level %s", tc.configLevel)
			} else {
				s.Empty(s.buffer.String(), "Should not have logged for level %s", tc.configLevel)
			}
		})
	}
}

// TestInvalidLogLevel tests handling of invalid log levels
func (s *LoggingTestSuite) TestInvalidLogLevel() {
	cfg := config.LogConfig{
		ServiceName: "test-service",
		LogLevel:    "INVALID_LEVEL",
		Environment: "test",
		Version:     "1.0.0",
	}

	err := InitLoggingWithOutput(cfg, s.buffer)
	s.NoError(err)

	// Should default to INFO level, so DEBUG should not log
	slog.Debug("debug message")
	s.Empty(s.buffer.String())

	// But INFO should log
	slog.Info("info message")
	s.NotEmpty(s.buffer.String())
}

// TestLoggingSuite runs the complete logging test suite
func TestLoggingSuite(t *testing.T) {
	suite.Run(t, new(LoggingTestSuite))
}
