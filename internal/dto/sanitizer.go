package dto

import (
	"html"
	"strings"
)

// SanitizeString strips dangerous HTML tags and escapes characters to prevent XSS injection
func SanitizeString(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}

	// Escape HTML special characters (<, >, &, ", ')
	escaped := html.EscapeString(trimmed)
	return escaped
}
