package format

import (
	"strings"
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
		{"QuoteSingle", Quote, "test", "> test"},
		{"QuoteMulti", Quote, "line1\nline2", "> line1\n> line2"},
		{"Header", Header, "Title", "# Title"},
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

func TestHeaderLevel(t *testing.T) {
	if res := HeaderLevel(1, "Title"); res != "# Title" {
		t.Errorf("HeaderLevel(1) = %q; want %q", res, "# Title")
	}
	if res := HeaderLevel(3, "Sub"); res != "### Sub" {
		t.Errorf("HeaderLevel(3) = %q; want %q", res, "### Sub")
	}
	if res := HeaderLevel(0, "Low"); res != "# Low" {
		t.Errorf("HeaderLevel(0) = %q; want %q", res, "# Low")
	}
	if res := HeaderLevel(10, "High"); res != "###### High" {
		t.Errorf("HeaderLevel(10) = %q; want %q", res, "###### High")
	}
}

func TestKeyVal(t *testing.T) {
	res := KeyVal("Status", "OK")
	expected := "**Status:** OK"
	if res != expected {
		t.Errorf("KeyVal() = %q; want %q", res, expected)
	}
}

func TestDivider(t *testing.T) {
	if res := Divider(); res != DefaultDivider {
		t.Errorf("Divider() = %q; want %q", res, DefaultDivider)
	}
}

func TestLists(t *testing.T) {
	items := []string{"First", "Second"}

	bullet := BulletList(items)
	expectedBullet := "• First\n• Second"
	if bullet != expectedBullet {
		t.Errorf("BulletList() = %q; want %q", bullet, expectedBullet)
	}
	if BulletList(nil) != "" {
		t.Errorf("BulletList(nil) should be empty string")
	}

	numbered := NumberedList(items)
	expectedNumbered := "1. First\n2. Second"
	if numbered != expectedNumbered {
		t.Errorf("NumberedList() = %q; want %q", numbered, expectedNumbered)
	}
	if NumberedList(nil) != "" {
		t.Errorf("NumberedList(nil) should be empty string")
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

func TestSplit(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		if res := Split("", 100); res != nil {
			t.Errorf("Split empty string = %v; want nil", res)
		}
	})

	t.Run("UnderLimit", func(t *testing.T) {
		res := Split("short string", 100)
		if len(res) != 1 || res[0] != "short string" {
			t.Errorf("Split under limit = %v; want ['short string']", res)
		}
	})

	t.Run("ParagraphSplit", func(t *testing.T) {
		text := "Paragraph 1\n\nParagraph 2\n\nParagraph 3"
		res := Split(text, 25)
		if len(res) < 2 {
			t.Errorf("Split by paragraph failed, got %d chunks: %v", len(res), res)
		}
		joined := strings.Join(res, "")
		if joined != text {
			t.Errorf("Joined chunks %q != original %q", joined, text)
		}
	})

	t.Run("DefaultLimit", func(t *testing.T) {
		longText := strings.Repeat("A", 7000)
		res := SplitDefault(longText)
		if len(res) != 2 {
			t.Errorf("SplitDefault got %d chunks; want 2", len(res))
		}
		if len(res[0]) != 6000 || len(res[1]) != 1000 {
			t.Errorf("SplitDefault chunk sizes = [%d, %d]; want [6000, 1000]", len(res[0]), len(res[1]))
		}
	})
}

func TestBuilder(t *testing.T) {
	b := NewBuilder()
	res := b.Header("Dashboard").
		NewLine().
		Divider().
		NewLine().
		KeyVal("Status", "Active").
		NewLine().
		Quote("All good").
		NewLine().
		BulletList([]string{"Node 1", "Node 2"}).
		NewLine().
		BoldEscaped("*test*").
		Space().
		ItalicEscaped("_test_").
		NewLine().
		Strikethrough("old").
		Space().
		Underline("new").
		NewLine().
		Code("val := 1").
		NewLine().
		CodeBlock("main()", "go").
		NewLine().
		Link("Google", "https://google.com").
		String()

	if b.Len() == 0 {
		t.Errorf("Builder Len() = 0")
	}

	if !strings.Contains(res, "# Dashboard") {
		t.Errorf("Builder missing Header")
	}
	if !strings.Contains(res, DefaultDivider) {
		t.Errorf("Builder missing Divider")
	}
	if !strings.Contains(res, "**Status:** Active") {
		t.Errorf("Builder missing KeyVal")
	}
	if !strings.Contains(res, "> All good") {
		t.Errorf("Builder missing Quote")
	}
	if !strings.Contains(res, "• Node 1\n• Node 2") {
		t.Errorf("Builder missing BulletList")
	}
	if !strings.Contains(res, "**\\*test\\***") {
		t.Errorf("Builder missing BoldEscaped")
	}

	b.Reset()
	if b.Len() != 0 || b.String() != "" {
		t.Errorf("Builder Reset() failed")
	}
}

