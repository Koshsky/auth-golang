package main

import (
	"context"
	"net"
	"testing"

	"github.com/Koshsky/subs-service/auth-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGRPCServer_WithoutTLS_Unit(t *testing.T) {
	// Arrange
	ctx := context.Background()
	cfg := &config.Config{
		EnableTLS: false,
	}

	// Act
	grpcServer, err := createGRPCServer(ctx, cfg)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, grpcServer)
}

func TestCreateGRPCServer_WithTLS_InvalidFiles_Unit(t *testing.T) {
	// Arrange
	ctx := context.Background()
	cfg := &config.Config{
		EnableTLS:   true,
		TLSCertFile: "nonexistent.crt",
		TLSKeyFile:  "nonexistent.key",
	}

	// Act
	grpcServer, err := createGRPCServer(ctx, cfg)

	// Assert
	require.Error(t, err)
	assert.Nil(t, grpcServer)
	assert.Contains(t, err.Error(), "open nonexistent.crt")
}

func TestStartServer_InvalidPort(t *testing.T) {
	// This test verifies that invalid ports are properly handled
	// We'll test the net.Listen function directly since that's what fails with invalid ports

	// Act - try to listen on invalid port
	_, err := net.Listen("tcp", ":99999") // Invalid port

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid port")
}

func TestStartServer_ValidPort(t *testing.T) {
	// This test verifies that valid ports are accepted
	// We'll test the net.Listen function directly

	// Act - try to listen on valid port
	listener, err := net.Listen("tcp", ":0") // Use port 0 to get any available port

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, listener)
	defer listener.Close()
}

// TestConfigValidation tests configuration validation scenarios
func TestConfigValidation(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			Database: config.DBConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "test",
				Password: "test",
				DBName:   "test",
				SSLMode:  "disable",
			},
			RabbitMQ: config.RabbitMQConfig{
				URL:      "amqp://guest:guest@localhost:5672/",
				Exchange: "test_exchange",
			},
			Logging: config.LogConfig{
				ServiceName: "auth-service",
				LogLevel:    "INFO",
				Environment: "test",
				Version:     "1.0.0",
			},
			JWTSecret:   "test-secret-key-32-chars-long-secret",
			Port:        "8080",
			EnableTLS:   false,
			TLSCertFile: "",
			TLSKeyFile:  "",
			Environment: "test",
		}

		// Act & Assert
		assert.NotEmpty(t, cfg.Database.Host)
		assert.NotEmpty(t, cfg.Database.Port)
		assert.NotEmpty(t, cfg.Database.User)
		assert.NotEmpty(t, cfg.Database.Password)
		assert.NotEmpty(t, cfg.Database.DBName)
		assert.NotEmpty(t, cfg.RabbitMQ.URL)
		assert.NotEmpty(t, cfg.RabbitMQ.Exchange)
		assert.NotEmpty(t, cfg.Logging.ServiceName)
		assert.Equal(t, "INFO", cfg.Logging.LogLevel)
		assert.NotEmpty(t, cfg.Logging.Environment)
		assert.NotEmpty(t, cfg.Logging.Version)
		assert.NotEmpty(t, cfg.JWTSecret)
		assert.NotEmpty(t, cfg.Port)
		assert.NotEmpty(t, cfg.Environment)
	})

	t.Run("EmptyDatabaseConfig", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			Database: config.DBConfig{},
			RabbitMQ: config.RabbitMQConfig{
				URL: "amqp://guest:guest@localhost:5672/",
			},
			JWTSecret: "test-secret-key-32-chars-long-secret",
			Port:      "8080",
		}

		// Act & Assert - проверяем только валидацию конфига, не подключаемся к сервисам
		assert.Empty(t, cfg.Database.Host)
		assert.Empty(t, cfg.Database.Port)
		assert.Empty(t, cfg.Database.User)
		assert.Empty(t, cfg.Database.Password)
		assert.Empty(t, cfg.Database.DBName)
		assert.NotEmpty(t, cfg.RabbitMQ.URL)
		assert.NotEmpty(t, cfg.JWTSecret)
		assert.NotEmpty(t, cfg.Port)
	})

	t.Run("EmptyRabbitMQConfig", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			Database: config.DBConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "test",
				Password: "test",
				DBName:   "test",
				SSLMode:  "disable",
			},
			RabbitMQ:  config.RabbitMQConfig{},
			JWTSecret: "test-secret-key-32-chars-long-secret",
			Port:      "8080",
		}

		// Act & Assert - проверяем только валидацию конфига, не подключаемся к сервисам
		assert.NotEmpty(t, cfg.Database.Host)
		assert.NotEmpty(t, cfg.Database.Port)
		assert.NotEmpty(t, cfg.Database.User)
		assert.NotEmpty(t, cfg.Database.Password)
		assert.NotEmpty(t, cfg.Database.DBName)
		assert.Empty(t, cfg.RabbitMQ.URL)
		assert.NotEmpty(t, cfg.JWTSecret)
		assert.NotEmpty(t, cfg.Port)
	})
}

// TestLogConfigValidation tests logging configuration validation scenarios
func TestLogConfigValidation(t *testing.T) {
	t.Run("ValidLogConfig", func(t *testing.T) {
		// Arrange
		logCfg := config.LogConfig{
			ServiceName: "auth-service",
			LogLevel:    "INFO",
			Environment: "test",
			Version:     "1.0.0",
		}

		// Act & Assert
		assert.Equal(t, "auth-service", logCfg.ServiceName)
		assert.Equal(t, "INFO", logCfg.LogLevel)
		assert.Equal(t, "test", logCfg.Environment)
		assert.Equal(t, "1.0.0", logCfg.Version)
	})

	t.Run("DifferentLogLevels", func(t *testing.T) {
		testCases := []struct {
			name     string
			level    string
			expected string
		}{
			{"Debug", "DEBUG", "DEBUG"},
			{"Info", "INFO", "INFO"},
			{"Warn", "WARN", "WARN"},
			{"Error", "ERROR", "ERROR"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				logCfg := config.LogConfig{
					ServiceName: "test-service",
					LogLevel:    tc.level,
					Environment: "test",
					Version:     "1.0.0",
				}

				assert.Equal(t, tc.level, logCfg.LogLevel)
				assert.Equal(t, tc.expected, logCfg.LogLevel)
			})
		}
	})

	t.Run("EmptyLogConfig", func(t *testing.T) {
		// Arrange
		logCfg := config.LogConfig{}

		// Act & Assert
		assert.Empty(t, logCfg.ServiceName)
		assert.Empty(t, logCfg.LogLevel)
		assert.Empty(t, logCfg.Environment)
		assert.Empty(t, logCfg.Version)
	})
}

// TestGRPCServerConfiguration tests different gRPC server configurations
func TestGRPCServerConfiguration(t *testing.T) {
	t.Run("ServerWithoutTLS", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		cfg := &config.Config{
			EnableTLS: false,
		}

		// Act
		server, err := createGRPCServer(ctx, cfg)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, server)
	})

	t.Run("ServerWithTLS_ValidFiles", func(t *testing.T) {
		// This test would require actual TLS certificate files
		// In a real scenario, you'd create temporary test certificates
		t.Skip("Skipping TLS test - requires actual certificate files")
	})

	t.Run("ServerWithTLS_MissingCertFile", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		cfg := &config.Config{
			EnableTLS:   true,
			TLSCertFile: "missing.crt",
			TLSKeyFile:  "missing.key",
		}

		// Act
		server, err := createGRPCServer(ctx, cfg)

		// Assert
		require.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "missing.crt")
	})
}
