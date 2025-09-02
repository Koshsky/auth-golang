package utils

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// MaskingTestSuite holds common test data and setup
type MaskingTestSuite struct {
	suite.Suite
}

// emailTestCase represents test case for email masking
type emailTestCase struct {
	name     string
	input    interface{}
	expected string
}

// sensitiveDataTestCase represents test case for sensitive data masking
type sensitiveDataTestCase struct {
	name     string
	key      string
	value    interface{}
	expected interface{}
}

// getEmailTestCases returns test cases for email masking
func (s *MaskingTestSuite) getEmailTestCases() []emailTestCase {
	return []emailTestCase{
		{
			name:     "Normal email",
			input:    "user@example.com",
			expected: "u***@example.com",
		},
		{
			name:     "Single character local part",
			input:    "a@test.com",
			expected: "*@test.com",
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
}

// getSensitiveDataTestCases returns test cases for sensitive data masking
func (s *MaskingTestSuite) getSensitiveDataTestCases() []sensitiveDataTestCase {
	return []sensitiveDataTestCase{
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
			name:     "Secret field",
			key:      "secret",
			value:    "my-secret",
			expected: "[REDACTED]",
		},
		{
			name:     "API Key field",
			key:      "api_key",
			value:    "api-key-123",
			expected: "[REDACTED]",
		},
		{
			name:     "Regular field",
			key:      "user_id",
			value:    123,
			expected: 123,
		},
	}
}

func (s *MaskingTestSuite) TestMaskEmail() {
	testCases := s.getEmailTestCases()

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result := MaskEmail(tc.input)
			s.Equal(tc.expected, result, "MaskEmail() returned unexpected result")
		})
	}
}

func (s *MaskingTestSuite) TestMaskSensitiveData() {
	testCases := s.getSensitiveDataTestCases()

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result := MaskSensitiveData(tc.key, tc.value)
			s.Equal(tc.expected, result, "MaskSensitiveData() returned unexpected result")
		})
	}
}

// TestMaskEmail_EdgeCases tests additional edge cases
func (s *MaskingTestSuite) TestMaskEmail_EdgeCases() {
	s.Run("Multiple @ symbols", func() {
		result := MaskEmail("user@domain@com")
		s.Equal("***************", result) // 15 symbols total
	})

	s.Run("Non-string input", func() {
		result := MaskEmail(12345)
		s.Equal("*****", result)
	})

	s.Run("Very short email - single char", func() {
		result := MaskEmail("u@d.c")
		s.Equal("*@d.c", result)
	})

	s.Run("Two character local part", func() {
		result := MaskEmail("ab@test.com")
		s.Equal("**@test.com", result)
	})

	s.Run("Three character local part", func() {
		result := MaskEmail("abc@test.com")
		s.Equal("a**@test.com", result)
	})
}

// TestMaskSensitiveData_CaseInsensitivity tests case insensitivity
func (s *MaskingTestSuite) TestMaskSensitiveData_CaseInsensitivity() {
	testCases := []struct {
		name     string
		key      string
		value    interface{}
		expected interface{}
	}{
		// Email variations
		{"email lowercase", "email", "test@example.com", "t***@example.com"},
		{"EMAIL uppercase", "EMAIL", "test@example.com", "t***@example.com"},
		{"Email mixed case", "Email", "test@example.com", "t***@example.com"},
		{"eMaIl random case", "eMaIl", "test@example.com", "t***@example.com"},

		// Password variations
		{"password lowercase", "password", "secret123", "[REDACTED]"},
		{"PASSWORD uppercase", "PASSWORD", "secret123", "[REDACTED]"},
		{"Password mixed case", "Password", "secret123", "[REDACTED]"},
		{"PaSSwoRd random case", "PaSSwoRd", "secret123", "[REDACTED]"},

		// Token variations
		{"token lowercase", "token", "abc123xyz", "[REDACTED]"},
		{"TOKEN uppercase", "TOKEN", "abc123xyz", "[REDACTED]"},
		{"Token mixed case", "Token", "abc123xyz", "[REDACTED]"},
		{"ToKeN random case", "ToKeN", "abc123xyz", "[REDACTED]"},

		// Secret variations
		{"secret lowercase", "secret", "mysecret", "[REDACTED]"},
		{"SECRET uppercase", "SECRET", "mysecret", "[REDACTED]"},
		{"Secret mixed case", "Secret", "mysecret", "[REDACTED]"},
		{"SeCrEt random case", "SeCrEt", "mysecret", "[REDACTED]"},

		// API key variations
		{"api_key lowercase", "api_key", "key123", "[REDACTED]"},
		{"API_KEY uppercase", "API_KEY", "key123", "[REDACTED]"},
		{"Api_Key mixed case", "Api_Key", "key123", "[REDACTED]"},
		{"ApI_kEy random case", "ApI_kEy", "key123", "[REDACTED]"},

		// Non-sensitive keys should not be masked
		{"non-sensitive lowercase", "username", "john_doe", "john_doe"},
		{"non-sensitive uppercase", "USERNAME", "john_doe", "john_doe"},
		{"non-sensitive mixed", "UserName", "john_doe", "john_doe"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result := MaskSensitiveData(tc.key, tc.value)
			s.Equal(tc.expected, result)
		})
	}
}

// TestMaskSensitiveData_EdgeCases tests edge cases for case insensitive masking
func (s *MaskingTestSuite) TestMaskSensitiveData_EdgeCases() {
	s.Run("Empty key", func() {
		result := MaskSensitiveData("", "value")
		s.Equal("value", result)
	})

	s.Run("Whitespace in key", func() {
		result := MaskSensitiveData(" password ", "secret")
		// Should not match because of whitespace
		s.Equal("secret", result)
	})

	s.Run("Partial match", func() {
		result := MaskSensitiveData("passwords", "secret")
		// Should not match - only exact matches
		s.Equal("secret", result)
	})

	s.Run("Nil value with sensitive key", func() {
		result := MaskSensitiveData("PASSWORD", nil)
		s.Equal("[REDACTED]", result)
	})

	s.Run("Empty string value with sensitive key", func() {
		result := MaskSensitiveData("TOKEN", "")
		s.Equal("[REDACTED]", result)
	})

	s.Run("Non-string value with sensitive key", func() {
		result := MaskSensitiveData("SECRET", 12345)
		s.Equal("[REDACTED]", result)
	})

	s.Run("Email with uppercase key and complex email", func() {
		result := MaskSensitiveData("EMAIL", "very.long.email.address@sub.domain.example.com")
		s.Equal("v*******************@sub.domain.example.com", result)
	})
}

// TestMaskingTestSuite runs the test suite
func TestMaskingTestSuite(t *testing.T) {
	suite.Run(t, new(MaskingTestSuite))
}
