package logging

import (
	"errors"
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

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

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

// createLogger creates a new structured logger with JSON output
func createLogger(config config.LogConfig, output io.Writer) (*slog.Logger, error) {
	if config.ServiceName == "" {
		return nil, errors.New("ServiceName is required")
	}
	if output == nil {
		output = os.Stdout
	}

	// Create base JSON handler with Kibana-compatible format
	baseHandler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level:     parseLogLevel(config.LogLevel),
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
	handler := newContextHandler(handlerWithAttrs)

	return slog.New(handler), nil
}

// ResetGlobalLogger resets the global logger state - used for testing only
func ResetGlobalLogger() {
	once = sync.Once{}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
}
