package utils

import (
	"log/slog"
	"strings"
)

// MaskEmail masks email address, keeping only first character of local part
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

	// Keep first character of local part
	if len(localPart) == 0 {
		return "*@" + domain
	} else if len(localPart) == 1 {
		return localPart + "@" + domain
	} else {
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
