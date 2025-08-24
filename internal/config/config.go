package config

import (
	"github.com/Koshsky/subs-service/auth-service/internal/utils"
	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RabbitMQConfig struct {
	URL      string
	Exchange string
}

type Config struct {
	Database    DBConfig
	RabbitMQ    RabbitMQConfig
	JWTSecret   string
	Port        string
	TLSCertFile string
	TLSKeyFile  string
	EnableTLS   bool
}

func LoadConfig() *Config {
	godotenv.Load()

	db := DBConfig{
		Host:     utils.GetEnv("AUTH_DB_HOST", "auth-db"),
		Port:     utils.GetEnvRequiredWithValidation("AUTH_DB_PORT", utils.ValidatePort),
		User:     utils.GetEnvRequired("AUTH_DB_USER"),
		Password: utils.GetEnvRequired("AUTH_DB_PASSWORD"),
		DBName:   utils.GetEnvRequired("AUTH_DB_NAME"),
		SSLMode:  utils.GetEnv("AUTH_DB_SSLMODE", "disable"),
	}

	rabbitmq := RabbitMQConfig{
		URL:      utils.GetEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		Exchange: utils.GetEnv("RABBITMQ_EXCHANGE", "user_events"),
	}

	return &Config{
		Database:    db,
		RabbitMQ:    rabbitmq,
		JWTSecret:   utils.GetEnvRequiredWithValidation("JWT_SECRET", utils.ValidateMinLength(32)),
		Port:        utils.GetEnvRequiredWithValidation("AUTH_SERVICE_PORT", utils.ValidatePort),
		TLSCertFile: utils.GetEnv("TLS_CERT_FILE", "certs/server-cert.pem"),
		TLSKeyFile:  utils.GetEnv("TLS_KEY_FILE", "certs/server-key.pem"),
		EnableTLS:   utils.GetEnvBool("ENABLE_TLS", false),
	}
}
