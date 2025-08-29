package utils

import (
	"log/slog"
	"strings"
)

// MaskEmail masks email address to prevent information leakage.
// For local parts with 3+ characters, keeps only first character.
// For very short local parts (1-2 chars), masks completely for security.
func MaskEmail(email interface{}) string {
	if email == nil {
		return ""
	}

	emailStr := ""
	switch v := email.(type) {
	case string:
		emailStr = v
	default:
		emailStr = slog.AnyValue(email).String()
	}

	if emailStr == "" {
		return ""
	}

	parts := strings.Split(emailStr, "@")
	if len(parts) != 2 {
		// Invalid email, mask completely
		return strings.Repeat("*", len(emailStr))
	}

	localPart := parts[0]
	domain := parts[1]

	// Mask local part to prevent information leakage
	if len(localPart) == 0 {
		return "*@" + domain
	} else if len(localPart) <= 2 {
		// For very short local parts (1-2 chars), mask completely to prevent info leakage
		return strings.Repeat("*", len(localPart)) + "@" + domain
	} else {
		// For longer local parts, keep first character and mask the rest
		maskedLocal := string(localPart[0]) + strings.Repeat("*", len(localPart)-1)
		return maskedLocal + "@" + domain
	}
}

// MaskSensitiveData masks sensitive data based on field name
func MaskSensitiveData(key string, value interface{}) interface{} {
	switch key {
	case "email":
		return MaskEmail(value)
	case "password", "token", "secret", "api_key":
		return "[REDACTED]"
	default:
		return value
	}
}
