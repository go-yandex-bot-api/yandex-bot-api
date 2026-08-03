// Package format provides functions to format text in Yandex Messenger Markdown and build complex messages.
package format

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// DefaultMaxMessageLength is the default maximum text length for Yandex Messenger text messages (6000 chars).
const DefaultMaxMessageLength = 6000

// DefaultDivider is the default horizontal separator line.
const DefaultDivider = "━━━━━━"

const (
	maxHeaderLevel    = 6
	paragraphBreakLen = 2
)

// markdownReplacer defines characters that need to be escaped in Yandex Messenger Markdown.
var markdownReplacer = strings.NewReplacer(
	"*", "\\*",
	"_", "\\_",
	"~", "\\~",
	"+", "\\+",
	"`", "\\`",
	"[", "\\[",
	"]", "\\]",
	"\\", "\\\\",
)

// Escape escapes special characters in a string so it can be safely used in Markdown.
func Escape(text string) string {
	return markdownReplacer.Replace(text)
}

// Bold returns text formatted as bold (**text**).
func Bold(text string) string {
	return fmt.Sprintf("**%s**", text)
}

// Italic returns text formatted as italic (__text__).
func Italic(text string) string {
	return fmt.Sprintf("__%s__", text)
}

// Strikethrough returns text formatted as strikethrough (~~text~~).
func Strikethrough(text string) string {
	return fmt.Sprintf("~~%s~~", text)
}

// Underline returns text formatted as underlined (++text++).
func Underline(text string) string {
	return fmt.Sprintf("++%s++", text)
}

// Code returns text formatted as inline code (`text`).
func Code(text string) string {
	return fmt.Sprintf("`%s`", text)
}

// CodeBlock returns text formatted as a multi-line code block with an optional language.
func CodeBlock(text, language string) string {
	return fmt.Sprintf("```%s\n%s\n```", language, text)
}

// Link returns text formatted as a clickable URL link ([text](url)).
func Link(text, url string) string {
	return fmt.Sprintf("[%s](%s)", text, url)
}

// Quote returns text formatted as a quote block (> text). Multi-line text will have > on each line.
func Quote(text string) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		lines[i] = "> " + l
	}
	return strings.Join(lines, "\n")
}

// BulletList formats a slice of items into a bulleted list (• item).
func BulletList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = "• " + item
	}
	return strings.Join(lines, "\n")
}

// NumberedList formats a slice of items into a numbered list (1. item).
func NumberedList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = fmt.Sprintf("%d. %s", i+1, item)
	}
	return strings.Join(lines, "\n")
}

// HeaderLevel returns text formatted as a header of a specific level (1-6).
func HeaderLevel(level int, text string) string {
	if level < 1 {
		level = 1
	}
	if level > maxHeaderLevel {
		level = maxHeaderLevel
	}
	return fmt.Sprintf("%s %s", strings.Repeat("#", level), text)
}

// Header returns text formatted as a top-level header (# text).
func Header(text string) string {
	return HeaderLevel(1, text)
}

// KeyVal formats a key-value pair as bold key and plain value (**Key:** Value).
func KeyVal(key, value string) string {
	return fmt.Sprintf("**%s:** %s", key, value)
}

// Divider returns a horizontal line divider.
func Divider() string {
	return DefaultDivider
}

// Split divides a text into slices of strings where each slice does not exceed maxLen.
// If maxLen is <= 0, DefaultMaxMessageLength (6000) is used.
// It attempts to split at paragraph breaks (\n\n), line breaks (\n), or spaces (' '),
// ensuring UTF-8 multi-byte characters are not corrupted.
func Split(text string, maxLen int) []string {
	if maxLen <= 0 {
		maxLen = DefaultMaxMessageLength
	}
	if text == "" {
		return nil
	}
	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	remaining := text

	for remaining != "" {
		if len(remaining) <= maxLen {
			chunks = append(chunks, remaining)
			break
		}

		sub := remaining[:maxLen]
		splitIdx := -1

		if idx := strings.LastIndex(sub, "\n\n"); idx > 0 {
			splitIdx = idx + paragraphBreakLen
		} else if idx := strings.LastIndex(sub, "\n"); idx > 0 {
			splitIdx = idx + 1
		} else if idx := strings.LastIndex(sub, " "); idx > 0 {
			splitIdx = idx + 1
		}

		if splitIdx <= 0 {
			splitIdx = maxLen
			for splitIdx > 0 && !utf8.RuneStart(remaining[splitIdx]) {
				splitIdx--
			}
			if splitIdx == 0 {
				splitIdx = maxLen
			}
		}

		chunks = append(chunks, remaining[:splitIdx])
		remaining = remaining[splitIdx:]
	}

	return chunks
}

// SplitDefault splits text using the default Yandex Messenger limit (6000 characters).
func SplitDefault(text string) []string {
	return Split(text, DefaultMaxMessageLength)
}

// Builder helps construct Markdown formatted strings using a fluent API.
type Builder struct {
	buf strings.Builder
}

// NewBuilder creates a new Builder instance.
func NewBuilder() *Builder {
	return &Builder{}
}

// Text appends plain text to the buffer.
func (b *Builder) Text(text string) *Builder {
	b.buf.WriteString(text)
	return b
}

// TextEscaped appends escaped plain text to the buffer.
func (b *Builder) TextEscaped(text string) *Builder {
	b.buf.WriteString(Escape(text))
	return b
}

// Bold appends bold text (**text**).
func (b *Builder) Bold(text string) *Builder {
	b.buf.WriteString(Bold(text))
	return b
}

// BoldEscaped appends escaped bold text (**escapedText**).
func (b *Builder) BoldEscaped(text string) *Builder {
	b.buf.WriteString(Bold(Escape(text)))
	return b
}

// Italic appends italic text (__text__).
func (b *Builder) Italic(text string) *Builder {
	b.buf.WriteString(Italic(text))
	return b
}

// ItalicEscaped appends escaped italic text (__escapedText__).
func (b *Builder) ItalicEscaped(text string) *Builder {
	b.buf.WriteString(Italic(Escape(text)))
	return b
}

// Strikethrough appends strikethrough text (~~text~~).
func (b *Builder) Strikethrough(text string) *Builder {
	b.buf.WriteString(Strikethrough(text))
	return b
}

// Underline appends underlined text (++text++).
func (b *Builder) Underline(text string) *Builder {
	b.buf.WriteString(Underline(text))
	return b
}

// Code appends inline code (`text`).
func (b *Builder) Code(text string) *Builder {
	b.buf.WriteString(Code(text))
	return b
}

// CodeBlock appends a multi-line code block (```language\ntext\n```).
func (b *Builder) CodeBlock(text, language string) *Builder {
	b.buf.WriteString(CodeBlock(text, language))
	return b
}

// Link appends a clickable URL link ([text](url)).
func (b *Builder) Link(text, url string) *Builder {
	b.buf.WriteString(Link(text, url))
	return b
}

// Quote appends a quote block (> text).
func (b *Builder) Quote(text string) *Builder {
	b.buf.WriteString(Quote(text))
	return b
}

// BulletList appends a bulleted list.
func (b *Builder) BulletList(items []string) *Builder {
	b.buf.WriteString(BulletList(items))
	return b
}

// NumberedList appends a numbered list.
func (b *Builder) NumberedList(items []string) *Builder {
	b.buf.WriteString(NumberedList(items))
	return b
}

// Header appends a top-level header (# text).
func (b *Builder) Header(text string) *Builder {
	b.buf.WriteString(Header(text))
	return b
}

// HeaderLevel appends a header of specified level.
func (b *Builder) HeaderLevel(level int, text string) *Builder {
	b.buf.WriteString(HeaderLevel(level, text))
	return b
}

// KeyVal appends a key-value pair (**key:** value).
func (b *Builder) KeyVal(key, value string) *Builder {
	b.buf.WriteString(KeyVal(key, value))
	return b
}

// Divider appends a divider line.
func (b *Builder) Divider() *Builder {
	b.buf.WriteString(Divider())
	return b
}

// NewLine appends a newline character (\n).
func (b *Builder) NewLine() *Builder {
	b.buf.WriteString("\n")
	return b
}

// NewLines appends n newline characters.
func (b *Builder) NewLines(n int) *Builder {
	for range n {
		b.buf.WriteString("\n")
	}
	return b
}

// Space appends a space character.
func (b *Builder) Space() *Builder {
	b.buf.WriteString(" ")
	return b
}

// Reset clears the builder buffer.
func (b *Builder) Reset() *Builder {
	b.buf.Reset()
	return b
}

// Len returns the byte length of accumulated content.
func (b *Builder) Len() int {
	return b.buf.Len()
}

// String returns the built string content.
func (b *Builder) String() string {
	return b.buf.String()
}

