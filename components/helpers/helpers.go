// Package helpers provides utility functions for gomponents templates.
// These mirror the helper functions available in Plush templates.
package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"net/url"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	g "maragu.dev/gomponents"
)

// String helpers

// StartsWith returns true if s starts with prefix
func StartsWith(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// EndsWith returns true if s ends with suffix
func EndsWith(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Contains returns true if s contains sub
func Contains(s, sub string) bool {
	return strings.Contains(s, sub)
}

// Matches returns true if s matches the regex pattern
func Matches(s, pattern string) bool {
	re := regexp.MustCompile(pattern)
	return re.MatchString(s)
}

// Replace replaces the first occurrence of old with new in s
func Replace(s, old, new string) string {
	return strings.Replace(s, old, new, 1)
}

// ReplaceAll replaces all occurrences of old with new in s
func ReplaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// ReplacePattern replaces all matches of pattern with replacement in s
func ReplacePattern(s, pattern, replacement string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(s, replacement)
}

// Upper returns s in uppercase
func Upper(s string) string {
	return strings.ToUpper(s)
}

// Lower returns s in lowercase
func Lower(s string) string {
	return strings.ToLower(s)
}

// Truncate truncates s to max characters with ellipsis
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// HTML/Markdown helpers

// Markdown converts markdown input to HTML using goldmark
func Markdown(input string) template.HTML {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(input), &buf); err != nil {
		return ""
	}
	return template.HTML(buf.String())
}

// HTML marks input as safe HTML (unescaped)
func HTML(input string) template.HTML {
	return template.HTML(input)
}

// UnescapeString unescapes HTML entities in input
func UnescapeString(input string) string {
	return html.UnescapeString(input)
}

// EscapeString escapes HTML special characters in input
func EscapeString(input string) string {
	return html.EscapeString(input)
}

// Encoding helpers

// URLEncode percent-encodes input for use in URLs
func URLEncode(input string) string {
	return url.QueryEscape(input)
}

// URLDecode decodes a percent-encoded string
func URLDecode(input string) string {
	decoded, err := url.QueryUnescape(input)
	if err != nil {
		return input
	}
	return decoded
}

// Stringify marshals data to a JSON string
func Stringify(data any) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

// StringifyPretty marshals data to a formatted JSON string
func StringifyPretty(data any) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

// Time helpers

// SecondsToISO8601 converts a duration in seconds to ISO8601 duration format
func SecondsToISO8601(seconds uint64) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	result := "PT"
	if hours > 0 {
		result += fmt.Sprintf("%dH", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dM", minutes)
	}
	if secs > 0 || result == "PT" {
		result += fmt.Sprintf("%dS", secs)
	}
	return result
}

// Gomponents-specific helpers

// Raw returns a gomponents Node containing raw HTML content
func Raw(htmlContent string) g.Node {
	return g.Raw(htmlContent)
}

// MarkdownNode converts markdown to a gomponents Node
func MarkdownNode(input string) g.Node {
	return g.Raw(string(Markdown(input)))
}

// Text returns a gomponents Text node (HTML-escaped)
func Text(s string) g.Node {
	return g.Text(s)
}

// Textf returns a formatted gomponents Text node (HTML-escaped)
func Textf(format string, args ...any) g.Node {
	return g.Textf(format, args...)
}

// If returns node if condition is true, otherwise nil
func If(condition bool, node g.Node) g.Node {
	return g.If(condition, node)
}

// Iff returns the result of fn() if condition is true
func Iff(condition bool, fn func() g.Node) g.Node {
	if condition {
		return fn()
	}
	return nil
}

// Map transforms a slice into a slice of Nodes
func Map[T any](items []T, fn func(T) g.Node) []g.Node {
	return g.Map(items, fn)
}

// Group combines multiple nodes into one
func Group(nodes []g.Node) g.Node {
	return g.Group(nodes)
}
