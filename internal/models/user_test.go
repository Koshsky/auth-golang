package models

import (
	"testing"

	"github.com/Koshsky/subs-service/auth-service/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestData contains test data for reuse
type TestData struct {
	validEmail    string
	validPassword string
}

var testData = TestData{
	validEmail:    "test@example.com",
	validPassword: "Password123!",
}

// TestEmailValidation tests email validation
func TestEmailValidation(t *testing.T) {
	validate := utils.NewValidator()

	testCases := []struct {
		name        string
		email       string
		expectValid bool
		description string
	}{
		{
			name:        "valid_email",
			email:       "test@example.com",
			expectValid: true,
			description: "Standard email format",
		},
		{
			name:        "email_with_subdomain",
			email:       "test@sub.example.com",
			expectValid: true,
			description: "Email with subdomain",
		},
		{
			name:        "email_with_numbers",
			email:       "test123@example.com",
			expectValid: true,
			description: "Email with numbers in local part",
		},
		{
			name:        "email_with_dots",
			email:       "test.name@example.com",
			expectValid: true,
			description: "Email with dots in local part",
		},
		{
			name:        "missing_at_symbol",
			email:       "invalid-email",
			expectValid: false,
			description: "Email without @ symbol",
		},
		{
			name:        "missing_domain",
			email:       "test@",
			expectValid: false,
			description: "Email without domain",
		},
		{
			name:        "missing_local_part",
			email:       "@example.com",
			expectValid: false,
			description: "Email without local part",
		},
		{
			name:        "contains_spaces",
			email:       "test @example.com",
			expectValid: false,
			description: "Email with spaces",
		},
		{
			name:        "empty_email",
			email:       "",
			expectValid: false,
			description: "Empty email string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			user := createTestUser(tc.email, testData.validPassword)

			// Act
			err := validate.Struct(user)

			// Assert
			if tc.expectValid {
				assert.NoError(t, err, "Expected valid email: %s", tc.description)
			} else {
				assert.Error(t, err, "Expected invalid email: %s", tc.description)
			}
		})
	}
}

// TestPasswordValidation tests password validation
func TestPasswordValidation(t *testing.T) {
	validate := utils.NewValidator()

	testCases := []struct {
		name        string
		password    string
		expectValid bool
		description string
	}{
		// Valid passwords
		{
			name:        "valid_password",
			password:    "Password123!",
			expectValid: true,
			description: "Password meets all requirements",
		},
		{
			name:        "password_with_at_symbol",
			password:    "MyPass@word1",
			expectValid: true,
			description: "Password with @ symbol",
		},
		{
			name:        "password_with_underscore",
			password:    "My_Pass123",
			expectValid: true,
			description: "Password with underscore",
		},
		{
			name:        "password_with_dash",
			password:    "My-Pass123",
			expectValid: true,
			description: "Password with dash",
		},
		{
			name:        "password_with_hash",
			password:    "My#Pass123",
			expectValid: true,
			description: "Password with hash symbol",
		},
		// Invalid passwords
		{
			name:        "empty_password",
			password:    "",
			expectValid: false,
			description: "Empty password",
		},
		{
			name:        "too_short_password",
			password:    "123",
			expectValid: false,
			description: "Password too short (less than 10 characters)",
		},
		{
			name:        "exactly_9_characters",
			password:    "Pass123!@",
			expectValid: false,
			description: "Password exactly 9 characters",
		},
		{
			name:        "missing_uppercase",
			password:    "password123!",
			expectValid: false,
			description: "Password missing uppercase letters",
		},
		{
			name:        "missing_lowercase",
			password:    "PASSWORD123!",
			expectValid: false,
			description: "Password missing lowercase letters",
		},
		{
			name:        "missing_special_character",
			password:    "Password123",
			expectValid: false,
			description: "Password missing special characters",
		},
		{
			name:        "only_uppercase_and_special",
			password:    "PASSWORD123!",
			expectValid: false,
			description: "Password with only uppercase and special characters",
		},
		{
			name:        "only_lowercase_and_special",
			password:    "password123!",
			expectValid: false,
			description: "Password with only lowercase and special characters",
		},
		{
			name:        "only_numbers",
			password:    "1234567890",
			expectValid: false,
			description: "Password with only numbers",
		},
		{
			name:        "only_lowercase",
			password:    "abcdefghijklmnopqrstuvwxyz",
			expectValid: false,
			description: "Password with only lowercase letters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			user := createTestUser(testData.validEmail, tc.password)

			// Act
			err := validate.Struct(user)

			// Assert
			if tc.expectValid {
				assert.NoError(t, err, "Expected valid password: %s", tc.description)
			} else {
				assert.Error(t, err, "Expected invalid password: %s", tc.description)
			}
		})
	}
}

// TestUserModelIntegration tests the integration of all model components
func TestUserModelIntegration(t *testing.T) {
	t.Run("complete_valid_user", func(t *testing.T) {
		// Arrange
		validate := utils.NewValidator()
		user := createTestUser(testData.validEmail, testData.validPassword)

		// Act
		validationErr := validate.Struct(user)

		// Assert
		assert.NoError(t, validationErr, "Valid user should pass validation")
		assert.Equal(t, uuid.Nil, user.ID, "User ID should be nil initially (UUID generation moved to repository)")
	})
}

// Helper functions

// createTestUser creates a test user with the given email and password
func createTestUser(email, password string) *User {
	return &User{
		Email:    email,
		Password: password,
	}
}

// Benchmark tests

func BenchmarkPasswordValidation(b *testing.B) {
	validate := utils.NewValidator()
	user := createTestUser(testData.validEmail, testData.validPassword)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validate.Struct(user)
	}
}

func BenchmarkEmailValidation(b *testing.B) {
	validate := utils.NewValidator()
	user := createTestUser(testData.validEmail, testData.validPassword)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validate.Struct(user)
	}
}
