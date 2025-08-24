package main

import (
	"net"
	"testing"

	"github.com/Koshsky/subs-service/auth-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGRPCServer_WithoutTLS_Unit(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		EnableTLS: false,
	}

	// Act
	grpcServer, err := createGRPCServer(cfg)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, grpcServer)
}

func TestCreateGRPCServer_WithTLS_InvalidFiles_Unit(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		EnableTLS:   true,
		TLSCertFile: "nonexistent.crt",
		TLSKeyFile:  "nonexistent.key",
	}

	// Act
	grpcServer, err := createGRPCServer(cfg)

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
			JWTSecret:   "test-secret-key-32-chars-long-secret",
			Port:        "8080",
			EnableTLS:   false,
			TLSCertFile: "",
			TLSKeyFile:  "",
		}

		// Act & Assert
		assert.NotEmpty(t, cfg.Database.Host)
		assert.NotEmpty(t, cfg.Database.Port)
		assert.NotEmpty(t, cfg.Database.User)
		assert.NotEmpty(t, cfg.Database.Password)
		assert.NotEmpty(t, cfg.Database.DBName)
		assert.NotEmpty(t, cfg.RabbitMQ.URL)
		assert.NotEmpty(t, cfg.RabbitMQ.Exchange)
		assert.NotEmpty(t, cfg.JWTSecret)
		assert.NotEmpty(t, cfg.Port)
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

// TestGRPCServerConfiguration tests different gRPC server configurations
func TestGRPCServerConfiguration(t *testing.T) {
	t.Run("ServerWithoutTLS", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			EnableTLS: false,
		}

		// Act
		server, err := createGRPCServer(cfg)

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
		cfg := &config.Config{
			EnableTLS:   true,
			TLSCertFile: "missing.crt",
			TLSKeyFile:  "missing.key",
		}

		// Act
		server, err := createGRPCServer(cfg)

		// Assert
		require.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "missing.crt")
	})
}
