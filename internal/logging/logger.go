package logging

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/Koshsky/subs-service/auth-service/internal/config"
)

var (
	once sync.Once
)

// InitLogging initializes global slog logger configuration
func InitLogging(config config.LogConfig) error {
	return InitLoggingWithOutput(config, os.Stdout)
}

// InitLoggingWithOutput initializes logging with custom output (for testing)
func InitLoggingWithOutput(logConfig config.LogConfig, output io.Writer) error {
	var err error
	once.Do(func() {
		var logger *slog.Logger
		logger, err = createLogger(logConfig, output)
		if err == nil {
			slog.SetDefault(logger)
		}
	})
	return err
}

// validateLogLevel validates log level string and returns the level with validity flag
func validateLogLevel(levelStr string) (slog.Level, bool) {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug, true
	case "info":
		return slog.LevelInfo, true
	case "warn", "warning":
		return slog.LevelWarn, true
	case "error":
		return slog.LevelError, true
	default:
		return slog.LevelInfo, false
	}
}

// createLogger creates a new structured logger with JSON output
func createLogger(config config.LogConfig, output io.Writer) (*slog.Logger, error) {
	if config.ServiceName == "" {
		return nil, errors.New("ServiceName is required")
	}
	if output == nil {
		output = os.Stdout
	}

	// Validate log level string before using it
	level, valid := validateLogLevel(config.LogLevel)
	if !valid {
		fmt.Fprintf(os.Stderr, "WARNING: Unrecognized log level '%s', defaulting to INFO\n", config.LogLevel)
		level = slog.LevelInfo
	}

	// Create base JSON handler with Kibana-compatible format
	baseHandler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
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
	handler := newContextHandler(handlerWithAttrs)

	return slog.New(handler), nil
}
