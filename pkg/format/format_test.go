package format

import (
	"testing"
)

func TestFormatting(t *testing.T) {
	tests := []struct {
		name     string
		function func(string) string
		input    string
		expected string
	}{
		{"Bold", Bold, "test", "**test**"},
		{"Italic", Italic, "test", "__test__"},
		{"Strikethrough", Strikethrough, "test", "~~test~~"},
		{"Underline", Underline, "test", "++test++"},
		{"Code", Code, "test", "`test`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.input)
			if result != tt.expected {
				t.Errorf("%s(%q) = %q; want %q", tt.name, tt.input, result, tt.expected)
			}
		})
	}
}

func TestCodeBlock(t *testing.T) {
	result := CodeBlock("fmt.Println()", "go")
	expected := "```go\nfmt.Println()\n```"
	if result != expected {
		t.Errorf("CodeBlock() = %q; want %q", result, expected)
	}
}

func TestLink(t *testing.T) {
	result := Link("Yandex", "https://yandex.ru")
	expected := "[Yandex](https://yandex.ru)"
	if result != expected {
		t.Errorf("Link() = %q; want %q", result, expected)
	}
}

func TestEscape(t *testing.T) {
	input := "*hello_world~ `code` [link] \\"
	expected := "\\*hello\\_world\\~ \\`code\\` \\[link\\] \\\\"
	result := Escape(input)

	if result != expected {
		t.Errorf("Escape() = %q; want %q", result, expected)
	}
}
