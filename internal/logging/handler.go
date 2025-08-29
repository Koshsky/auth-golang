package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/Koshsky/subs-service/auth-service/internal/utils"
)

// contextHandler wraps slog.Handler to extract context attributes
type contextHandler struct {
	base slog.Handler
}

// newContextHandler creates a new contextHandler
func newContextHandler(base slog.Handler) *contextHandler {
	return &contextHandler{base: base}
}

// Enabled reports whether the handler handles records at the given level
func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

// Handle handles the Record by extracting context attributes
func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	contextAttrs := extractContextAttrs(ctx)

	allAttrs := make([]slog.Attr, 0, len(contextAttrs)+r.NumAttrs())
	allAttrs = append(allAttrs, contextAttrs...)

	// Add original attributes from record
	r.Attrs(func(a slog.Attr) bool {
		allAttrs = append(allAttrs, a)
		return true
	})

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
