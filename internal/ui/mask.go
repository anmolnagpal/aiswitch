package ui

import "strings"

// MaskSecret returns a display-safe version of a secret string.
// The first 8 and last 4 characters are preserved; everything in between
// is replaced with "...". Strings of 12 characters or fewer are fully masked.
func MaskSecret(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 12 {
		return strings.Repeat("*", len(s))
	}
	return s[:8] + "..." + s[len(s)-4:]
}
