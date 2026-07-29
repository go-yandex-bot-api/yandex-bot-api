// Package format provides functions to format text in Yandex Messenger Markdown.
package format

import (
	"fmt"
	"strings"
)

// MarkdownEscaper defines characters that need to be escaped in Yandex Messenger Markdown.
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
