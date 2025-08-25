package utils

import "testing"

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "Normal email",
			input:    "user@example.com",
			expected: "u***@example.com",
		},
		{
			name:     "Single character local part",
			input:    "a@test.com",
			expected: "a@test.com",
		},
		{
			name:     "Long email",
			input:    "testuser@domain.org",
			expected: "t*******@domain.org",
		},
		{
			name:     "Invalid email",
			input:    "invalid-email",
			expected: "*************",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: "",
		},
		{
			name:     "Empty local part",
			input:    "@example.com",
			expected: "*@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskEmail(tt.input)
			if result != tt.expected {
				t.Errorf("MaskEmail() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMaskSensitiveData(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    interface{}
		expected interface{}
	}{
		{
			name:     "Email field",
			key:      "email",
			value:    "user@example.com",
			expected: "u***@example.com",
		},
		{
			name:     "Password field",
			key:      "password",
			value:    "secret123",
			expected: "[REDACTED]",
		},
		{
			name:     "Token field",
			key:      "token",
			value:    "jwt-token-here",
			expected: "[REDACTED]",
		},
		{
			name:     "Regular field",
			key:      "user_id",
			value:    123,
			expected: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveData(tt.key, tt.value)
			if result != tt.expected {
				t.Errorf("MaskSensitiveData() = %v, want %v", result, tt.expected)
			}
		})
	}
}
