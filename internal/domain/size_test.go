// internal/domain/size_test.go
package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeSizeInt(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		// invalid / default
		{"Zero input", 0, DefaultSize},
		{"Negative input", -1, DefaultSize},
		{"Very small input", 1, 16},

		// exact matches
		{"Exact match min", 16, 16},
		{"Exact match mid", 64, 64},
		{"Exact match large", 128, 128},
		{"Exact match max", 320, 320},

		// bucket-up behavior
		{"Between 16 and 32", 17, 32},
		{"Between 32 and 64", 40, 64},
		{"Between 64 and 128", 100, 128},
		{"Between 128 and 320", 200, 320},

		// overflow
		{"Just over max", 321, 320},
		{"Much larger than max", 1000, 320},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeSizeInt(tt.input)
			assert.Equal(t, tt.expected, result, "input=%d", tt.input)
		})
	}
}

func TestNormalizeSizeQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		// invalid
		{"Empty string", "", maxSize()},
		{"Invalid string (text)", "abc", maxSize()},
		{"Invalid string (mixed)", "12px", maxSize()},

		// valid
		{"Valid string (bucket up)", "15", 16},
		{"Valid string (exact)", "64", 64},
		{"Valid string (mid bucket)", "100", 128},
		{"Valid string (overflow)", "500", 320},
		{"Valid string (zero)", "0", 320},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeSizeQuery(tt.input)
			assert.Equal(t, tt.expected, result, "input=%s", tt.input)
		})
	}
}
