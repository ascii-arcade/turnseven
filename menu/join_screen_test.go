package menu

import "testing"

func TestFormatCode(t *testing.T) {
	tests := []struct {
		before string
		value  string
		want   string
	}{
		{"A", "ab", "AB"},
		{"AB", "ABc", "ABC-"},
		{"ABC-", "ABC-d", "ABC-D"},
		{"ABC-D", "ABC-", "ABC-"},
		{"ABC-", "ABC", "AB"},
		{"AB", "A", "A"},
		{"", "abc-def", "ABC-DEF"},
		{"", "abcdef", "ABC-DEF"},
	}
	for _, tt := range tests {
		if got := formatCode(tt.before, tt.value); got != tt.want {
			t.Errorf("formatCode(%q, %q) = %q, want %q", tt.before, tt.value, got, tt.want)
		}
	}
}
