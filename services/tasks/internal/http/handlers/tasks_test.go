package handlers

import "testing"

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain text", "hello world", "hello world"},
		{"simple script tag", "<script>alert(1)</script>", "alert(1)"},
		{"multiple tags", "<b>bold</b> and <i>italic</i>", "bold and italic"},
		{"nested tags", "<div><p>text</p></div>", "text"},
		{"empty string", "", ""},
		{"only tags", "<><>", ""},
		{"tag with attributes", "<a href='#'>click</a>", "click"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeString(tt.input); got != tt.want {
				t.Errorf("sanitizeString() = %q, want %q", got, tt.want)
			}
		})
	}
}
