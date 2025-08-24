package utils

import (
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestValidatePassword(t *testing.T) {
	v := validator.New()
	if err := RegisterCustomValidations(v); err != nil {
		t.Fatalf("Failed to register custom validations: %v", err)
	}

	tests := []struct {
		name     string
		password string
		isValid  bool
	}{
		{
			name:     "Valid password with all requirements",
			password: "Password123!",
			isValid:  true,
		},
		{
			name:     "Password too short",
			password: "Pass123!",
			isValid:  false,
		},
		{
			name:     "Pasword too long",
			password: strings.Repeat("Aa1!", 19),
			isValid:  false,
		},
		{
			name:     "Password missing uppercase",
			password: "password123!",
			isValid:  false,
		},
		{
			name:     "Password missing lowercase",
			password: "PASSWORD123!",
			isValid:  false,
		},
		{
			name:     "Password missing special character",
			password: "Password123",
			isValid:  false,
		},
		{
			name:     "Password with underscore",
			password: "My_Pass123",
			isValid:  true,
		},
		{
			name:     "Password with dash",
			password: "My-Pass123",
			isValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Var(tt.password, "password")
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	assert.NotNil(t, validator)

	// Test that custom validations are registered
	err := validator.Var("Password123!", "password")
	assert.NoError(t, err)

	err = validator.Var("pass", "password")
	assert.Error(t, err)
}
