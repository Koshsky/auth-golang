package logging

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/Koshsky/subs-service/auth-service/internal/utils"
)

// Config holds logger configuration
type Config struct {
	ServiceName string
	LogLevel    slog.Level
	Environment string
	Version     string
	Output      io.Writer
}

// LogCtx holds context data for logging
type LogCtx struct {
	UserID    interface{} `json:"user_id,omitempty"`
	RequestID interface{} `json:"request_id,omitempty"`
	TraceID   interface{} `json:"trace_id,omitempty"`
	Operation interface{} `json:"operation,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Email     interface{} `json:"email,omitempty"`
}

// contextHandler wraps slog.Handler to extract context attributes
type contextHandler struct {
	base slog.Handler
}

// NewLogger creates a new structured logger with JSON output
func NewLogger(config Config) (*slog.Logger, error) {
	if config.ServiceName == "" {
		return nil, errors.New("ServiceName is required")
	}
	if config.Output == nil {
		config.Output = os.Stdout
	}

	// Create base JSON handler with Kibana-compatible format
	baseHandler := slog.NewJSONHandler(config.Output, &slog.HandlerOptions{
		Level:     config.LogLevel,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				return slog.Attr{Key: "@timestamp", Value: a.Value}
			case slog.LevelKey:
				level := a.Value.Any().(slog.Level)
				return slog.Attr{Key: "level", Value: slog.StringValue(level.String())}
			case slog.MessageKey:
				return slog.Attr{Key: "message", Value: a.Value}
			case slog.SourceKey:
				return slog.Attr{Key: "source", Value: a.Value}
			}
			return a
		},
	})

	// Add service metadata
	handlerWithAttrs := baseHandler.WithAttrs([]slog.Attr{
		slog.String("service", config.ServiceName),
		slog.String("environment", config.Environment),
		slog.String("version", config.Version),
		slog.String("hostname", getHostname()),
	})

	// Wrap with context handler
	handler := &contextHandler{base: handlerWithAttrs}

	return slog.New(handler), nil
}

// WithLogCtx adds LogCtx to the context
func WithLogCtx(ctx context.Context, logCtx *LogCtx) context.Context {
	return context.WithValue(ctx, "log_ctx", logCtx)
}

// Enabled reports whether the handler handles records at the given level
func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

// Handle handles the Record by extracting context attributes
func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Extract context attributes
	contextAttrs := extractContextAttrs(ctx)

	// Collect all attributes
	allAttrs := make([]slog.Attr, 0, len(contextAttrs)+r.NumAttrs())
	allAttrs = append(allAttrs, contextAttrs...)

	// Add original attributes from record
	r.Attrs(func(a slog.Attr) bool {
		allAttrs = append(allAttrs, a)
		return true
	})

	// Create new record with all attributes
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	newRecord.AddAttrs(allAttrs...)

	return h.base.Handle(ctx, newRecord)
}

// WithAttrs returns a new Handler with additional attributes
func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{base: h.base.WithAttrs(attrs)}
}

// WithGroup returns a new Handler with a group
func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{base: h.base.WithGroup(name)}
}

// extractContextAttrs extracts attributes from LogCtx in context
func extractContextAttrs(ctx context.Context) []slog.Attr {
	var attrs []slog.Attr

	logCtx, ok := ctx.Value("log_ctx").(*LogCtx)
	if !ok || logCtx == nil {
		return attrs
	}

	if logCtx.UserID != nil {
		attrs = append(attrs, slog.Any("user_id", logCtx.UserID))
	}
	if logCtx.RequestID != nil {
		attrs = append(attrs, slog.Any("request_id", logCtx.RequestID))
	}
	if logCtx.TraceID != nil {
		attrs = append(attrs, slog.Any("trace_id", logCtx.TraceID))
	}
	if logCtx.Operation != nil {
		attrs = append(attrs, slog.Any("operation", logCtx.Operation))
	}
	if logCtx.Error != nil {
		attrs = append(attrs, slog.Any("error", logCtx.Error))
	}
	if logCtx.Email != nil {
		// Mask email for security
		maskedEmail := utils.MaskEmail(logCtx.Email)
		attrs = append(attrs, slog.String("email", maskedEmail))
	}

	return attrs
}

// getHostname returns system hostname
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
