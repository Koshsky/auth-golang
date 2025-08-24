package utils

import (
	"fmt"
	"os"
	"strconv"
)

// GetEnv gets an environment variable with default value
// Use this for non-critical variables that can have defaults
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// GetEnvRequired gets a critical environment variable and panics if not set
// Use this for critical variables like passwords, secrets, ports
func GetEnvRequired(key string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	panic(fmt.Sprintf("CRITICAL ERROR: Environment variable %s is not set", key))
}

// GetEnvRequiredWithValidation gets a critical environment variable with validation
func GetEnvRequiredWithValidation(key string, validator func(string) error) string {
	value := GetEnvRequired(key)
	if err := validator(value); err != nil {
		panic(fmt.Sprintf("CRITICAL ERROR: Environment variable %s validation failed: %v", key, err))
	}
	return value
}

// GetEnvBool gets an environment variable as a boolean
func GetEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return boolValue
	}
	return defaultValue
}

// GetEnvBoolRequired gets a critical boolean environment variable
func GetEnvBoolRequired(key string) bool {
	if value, exists := os.LookupEnv(key); exists {
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			panic(fmt.Sprintf("CRITICAL ERROR: Environment variable %s is not a valid boolean", key))
		}
		return boolValue
	}
	panic(fmt.Sprintf("CRITICAL ERROR: Environment variable %s is not set", key))
}

// GetEnvInt gets an environment variable as an integer
func GetEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return intValue
	}
	return defaultValue
}

// GetEnvIntRequired gets a critical integer environment variable
func GetEnvIntRequired(key string) int {
	if value, exists := os.LookupEnv(key); exists {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			panic(fmt.Sprintf("CRITICAL ERROR: Environment variable %s is not a valid integer", key))
		}
		return intValue
	}
	panic(fmt.Sprintf("CRITICAL ERROR: Environment variable %s is not set", key))
}

// ValidatePort validates that a string is a valid port number
func ValidatePort(port string) error {
	if port == "" {
		return fmt.Errorf("port cannot be empty")
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("port must be a number")
	}

	if portNum < 1024 || portNum > 65535 {
		return fmt.Errorf("port must be between 1024 and 65535")
	}

	return nil
}

// ValidateNonEmpty validates that a string is not empty
func ValidateNonEmpty(value string) error {
	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

// ValidateMinLength validates that a string meets minimum length requirement
func ValidateMinLength(minLength int) func(string) error {
	return func(value string) error {
		if len(value) < minLength {
			return fmt.Errorf("value must be at least %d characters long", minLength)
		}
		return nil
	}
}
