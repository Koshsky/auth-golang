package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		setValue     string
		expected     string
	}{
		{
			name:         "Variable exists",
			key:          "TEST_VAR",
			defaultValue: "default",
			setValue:     "custom_value",
			expected:     "custom_value",
		},
		{
			name:         "Variable does not exist",
			key:          "NONEXISTENT_VAR",
			defaultValue: "default",
			setValue:     "",
			expected:     "default",
		},
		{
			name:         "Variable exists but empty",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			setValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up after test
			defer os.Unsetenv(tt.key)

			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
			}

			result := GetEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvRequired(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		setValue    string
		shouldPanic bool
		expected    string
	}{
		{
			name:        "Variable exists and not empty",
			key:         "REQUIRED_VAR",
			setValue:    "required_value",
			shouldPanic: false,
			expected:    "required_value",
		},
		{
			name:        "Variable does not exist",
			key:         "NONEXISTENT_REQUIRED_VAR",
			setValue:    "",
			shouldPanic: true,
			expected:    "",
		},
		{
			name:        "Variable exists but empty",
			key:         "EMPTY_REQUIRED_VAR",
			setValue:    "",
			shouldPanic: true,
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up after test
			defer os.Unsetenv(tt.key)

			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
			}

			if tt.shouldPanic {
				assert.PanicsWithValue(t, "CRITICAL ERROR: Environment variable "+tt.key+" is not set", func() {
					GetEnvRequired(tt.key)
				})
			} else {
				result := GetEnvRequired(tt.key)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetEnvRequiredWithValidation(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		setValue    string
		validator   func(string) error
		shouldPanic bool
		expected    string
	}{
		{
			name:        "Valid port number",
			key:         "VALID_PORT",
			setValue:    "8080",
			validator:   ValidatePort,
			shouldPanic: false,
			expected:    "8080",
		},
		{
			name:        "Invalid port number - too low",
			key:         "INVALID_PORT_LOW",
			setValue:    "80",
			validator:   ValidatePort,
			shouldPanic: true,
			expected:    "",
		},
		{
			name:        "Invalid port number - too high",
			key:         "INVALID_PORT_HIGH",
			setValue:    "70000",
			validator:   ValidatePort,
			shouldPanic: true,
			expected:    "",
		},
		{
			name:        "Invalid port number - not a number",
			key:         "INVALID_PORT_NAN",
			setValue:    "not_a_number",
			validator:   ValidatePort,
			shouldPanic: true,
			expected:    "",
		},
		{
			name:        "Variable does not exist",
			key:         "NONEXISTENT_VALIDATED_VAR",
			setValue:    "",
			validator:   ValidateNonEmpty,
			shouldPanic: true,
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up after test
			defer os.Unsetenv(tt.key)

			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
			}

			if tt.shouldPanic {
				assert.Panics(t, func() {
					GetEnvRequiredWithValidation(tt.key, tt.validator)
				})
			} else {
				result := GetEnvRequiredWithValidation(tt.key, tt.validator)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		setValue     string
		expected     bool
	}{
		{
			name:         "True value",
			key:          "BOOL_TRUE",
			defaultValue: false,
			setValue:     "true",
			expected:     true,
		},
		{
			name:         "False value",
			key:          "BOOL_FALSE",
			defaultValue: true,
			setValue:     "false",
			expected:     false,
		},
		{
			name:         "Invalid boolean value",
			key:          "BOOL_INVALID",
			defaultValue: true,
			setValue:     "not_a_bool",
			expected:     true, // Should return default
		},
		{
			name:         "Variable does not exist",
			key:          "NONEXISTENT_BOOL",
			defaultValue: false,
			setValue:     "",
			expected:     false,
		},
		{
			name:         "Case insensitive true",
			key:          "BOOL_TRUE_CASE",
			defaultValue: false,
			setValue:     "TRUE",
			expected:     true,
		},
		{
			name:         "Case insensitive false",
			key:          "BOOL_FALSE_CASE",
			defaultValue: true,
			setValue:     "FALSE",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up after test
			defer os.Unsetenv(tt.key)

			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
			}

			result := GetEnvBool(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvBoolRequired(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		setValue    string
		shouldPanic bool
		expected    bool
	}{
		{
			name:        "True value",
			key:         "REQUIRED_BOOL_TRUE",
			setValue:    "true",
			shouldPanic: false,
			expected:    true,
		},
		{
			name:        "False value",
			key:         "REQUIRED_BOOL_FALSE",
			setValue:    "false",
			shouldPanic: false,
			expected:    false,
		},
		{
			name:        "Invalid boolean value",
			key:         "REQUIRED_BOOL_INVALID",
			setValue:    "not_a_bool",
			shouldPanic: true,
			expected:    false,
		},
		{
			name:        "Variable does not exist",
			key:         "NONEXISTENT_REQUIRED_BOOL",
			setValue:    "",
			shouldPanic: true,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up after test
			defer os.Unsetenv(tt.key)

			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
			}

			if tt.shouldPanic {
				assert.Panics(t, func() {
					GetEnvBoolRequired(tt.key)
				})
			} else {
				result := GetEnvBoolRequired(tt.key)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		setValue     string
		expected     int
	}{
		{
			name:         "Valid integer",
			key:          "INT_VALID",
			defaultValue: 0,
			setValue:     "42",
			expected:     42,
		},
		{
			name:         "Zero value",
			key:          "INT_ZERO",
			defaultValue: 100,
			setValue:     "0",
			expected:     0,
		},
		{
			name:         "Negative integer",
			key:          "INT_NEGATIVE",
			defaultValue: 0,
			setValue:     "-10",
			expected:     -10,
		},
		{
			name:         "Invalid integer",
			key:          "INT_INVALID",
			defaultValue: 100,
			setValue:     "not_a_number",
			expected:     100, // Should return default
		},
		{
			name:         "Variable does not exist",
			key:          "NONEXISTENT_INT",
			defaultValue: 50,
			setValue:     "",
			expected:     50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up after test
			defer os.Unsetenv(tt.key)

			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
			}

			result := GetEnvInt(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvIntRequired(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		setValue    string
		shouldPanic bool
		expected    int
	}{
		{
			name:        "Valid integer",
			key:         "REQUIRED_INT_VALID",
			setValue:    "42",
			shouldPanic: false,
			expected:    42,
		},
		{
			name:        "Zero value",
			key:         "REQUIRED_INT_ZERO",
			setValue:    "0",
			shouldPanic: false,
			expected:    0,
		},
		{
			name:        "Invalid integer",
			key:         "REQUIRED_INT_INVALID",
			setValue:    "not_a_number",
			shouldPanic: true,
			expected:    0,
		},
		{
			name:        "Variable does not exist",
			key:         "NONEXISTENT_REQUIRED_INT",
			setValue:    "",
			shouldPanic: true,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up after test
			defer os.Unsetenv(tt.key)

			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
			}

			if tt.shouldPanic {
				assert.Panics(t, func() {
					GetEnvIntRequired(tt.key)
				})
			} else {
				result := GetEnvIntRequired(tt.key)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name        string
		port        string
		expectError bool
	}{
		{
			name:        "Valid port in range",
			port:        "8080",
			expectError: false,
		},
		{
			name:        "Minimum valid port",
			port:        "1024",
			expectError: false,
		},
		{
			name:        "Maximum valid port",
			port:        "65535",
			expectError: false,
		},
		{
			name:        "Port too low",
			port:        "80",
			expectError: true,
		},
		{
			name:        "Port too high",
			port:        "70000",
			expectError: true,
		},
		{
			name:        "Empty port",
			port:        "",
			expectError: true,
		},
		{
			name:        "Not a number",
			port:        "not_a_port",
			expectError: true,
		},
		{
			name:        "Negative port",
			port:        "-1",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.port)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNonEmpty(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
	}{
		{
			name:        "Non-empty string",
			value:       "some value",
			expectError: false,
		},
		{
			name:        "Empty string",
			value:       "",
			expectError: true,
		},
		{
			name:        "Whitespace only",
			value:       "   ",
			expectError: false, // Whitespace is considered non-empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNonEmpty(tt.value)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMinLength(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		minLength   int
		expectError bool
	}{
		{
			name:        "String meets minimum length",
			value:       "hello",
			minLength:   5,
			expectError: false,
		},
		{
			name:        "String exceeds minimum length",
			value:       "hello world",
			minLength:   5,
			expectError: false,
		},
		{
			name:        "String shorter than minimum",
			value:       "hi",
			minLength:   5,
			expectError: true,
		},
		{
			name:        "Empty string with minimum length 1",
			value:       "",
			minLength:   1,
			expectError: true,
		},
		{
			name:        "Empty string with minimum length 0",
			value:       "",
			minLength:   0,
			expectError: false,
		},
		{
			name:        "Unicode string",
			value:       "привет",
			minLength:   6,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := ValidateMinLength(tt.minLength)
			err := validator(tt.value)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
