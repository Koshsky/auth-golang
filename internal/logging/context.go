package logging

import "context"

/*
Package logging provides context-aware structured logging capabilities.

Example usage:

1. Using individual methods:
   ctx := WithUserID(context.Background(), "12345")
   ctx = WithRequestID(ctx, "req-abc-123")
   ctx = WithOperation(ctx, "user_login")

   slog.InfoContext(ctx, "User logged in successfully")

2. Using multiple fields at once:
   ctx := WithMultiple(context.Background(), map[string]interface{}{
       "user_id": "12345",
       "request_id": "req-abc-123",
       "operation": "user_login",
   })

   slog.InfoContext(ctx, "User logged in successfully")

All methods automatically create a new LogCtx if one doesn't exist in the context,
or update the existing one if it does.
*/

// LogCtx holds context data for logging
type LogCtx struct {
	UserID    interface{} `json:"user_id,omitempty"`
	RequestID interface{} `json:"request_id,omitempty"`
	TraceID   interface{} `json:"trace_id,omitempty"`
	Operation interface{} `json:"operation,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Email     interface{} `json:"email,omitempty"`
}

// WithLogCtx adds LogCtx to the context
func WithLogCtx(ctx context.Context, logCtx *LogCtx) context.Context {
	return context.WithValue(ctx, "log_ctx", logCtx)
}

// getOrCreateLogCtx gets existing LogCtx from context or creates a new one
func getOrCreateLogCtx(ctx context.Context) (*LogCtx, context.Context) {
	if existing, ok := ctx.Value("log_ctx").(*LogCtx); ok && existing != nil {
		return existing, ctx
	}
	newLogCtx := &LogCtx{}
	return newLogCtx, context.WithValue(ctx, "log_ctx", newLogCtx)
}

// WithUserID adds user ID to the logging context
func WithUserID(ctx context.Context, userID interface{}) context.Context {
	logCtx, newCtx := getOrCreateLogCtx(ctx)
	logCtx.UserID = userID
	return newCtx
}

// WithRequestID adds request ID to the logging context
func WithRequestID(ctx context.Context, requestID interface{}) context.Context {
	logCtx, newCtx := getOrCreateLogCtx(ctx)
	logCtx.RequestID = requestID
	return newCtx
}

// WithTraceID adds trace ID to the logging context
func WithTraceID(ctx context.Context, traceID interface{}) context.Context {
	logCtx, newCtx := getOrCreateLogCtx(ctx)
	logCtx.TraceID = traceID
	return newCtx
}

// WithOperation adds operation name to the logging context
func WithOperation(ctx context.Context, operation interface{}) context.Context {
	logCtx, newCtx := getOrCreateLogCtx(ctx)
	logCtx.Operation = operation
	return newCtx
}

// WithError adds error to the logging context
func WithError(ctx context.Context, err interface{}) context.Context {
	logCtx, newCtx := getOrCreateLogCtx(ctx)
	logCtx.Error = err
	return newCtx
}

// WithEmail adds email to the logging context
func WithEmail(ctx context.Context, email interface{}) context.Context {
	logCtx, newCtx := getOrCreateLogCtx(ctx)
	logCtx.Email = email
	return newCtx
}

// WithMultiple adds multiple fields to the logging context at once
func WithMultiple(ctx context.Context, fields map[string]interface{}) context.Context {
	logCtx, newCtx := getOrCreateLogCtx(ctx)

	for key, value := range fields {
		switch key {
		case "user_id":
			logCtx.UserID = value
		case "request_id":
			logCtx.RequestID = value
		case "trace_id":
			logCtx.TraceID = value
		case "operation":
			logCtx.Operation = value
		case "error":
			logCtx.Error = value
		case "email":
			logCtx.Email = value
		}
	}

	return newCtx
}
