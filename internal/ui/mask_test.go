package ui

import (
	"strings"
	"testing"
)

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, got string)
	}{
		{
			name:  "short string fully masked",
			input: "abc",
			check: func(t *testing.T, got string) {
				if strings.Contains(got, "abc") {
					t.Errorf("short secret not masked: %q", got)
				}
				if got != "***" {
					t.Errorf("got %q, want ***", got)
				}
			},
		},
		{
			name:  "exactly 12 chars fully masked",
			input: "abcdefghijkl",
			check: func(t *testing.T, got string) {
				if got != "************" {
					t.Errorf("got %q, want 12 stars", got)
				}
			},
		},
		{
			name:  "long key shows prefix and suffix",
			input: "sk-ant-api03-LONGKEY1234",
			check: func(t *testing.T, got string) {
				if !strings.HasPrefix(got, "sk-ant-a") {
					t.Errorf("expected prefix sk-ant-a, got %q", got)
				}
				if !strings.HasSuffix(got, "1234") {
					t.Errorf("expected suffix 1234, got %q", got)
				}
				if !strings.Contains(got, "...") {
					t.Errorf("expected '...' separator, got %q", got)
				}
			},
		},
		{
			name:  "trims whitespace before masking",
			input: "  sk-ant-api03-LONGKEY1234  ",
			check: func(t *testing.T, got string) {
				if strings.HasPrefix(got, " ") || strings.HasSuffix(got, " ") {
					t.Errorf("leading/trailing space not trimmed: %q", got)
				}
			},
		},
		{
			name:  "empty string",
			input: "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("expected empty string, got %q", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskSecret(tt.input)
			tt.check(t, got)
		})
	}
}
