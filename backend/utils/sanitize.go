package utils

import (
	"html"
	"regexp"
	"strings"
)

var (
	chineseParenRegex = regexp.MustCompile(`（([^）]*)）`)
	englishParenRegex = regexp.MustCompile(`\(([^)]*)\)`)
)

func SanitizeInput(input string) string {
	cleaned := html.EscapeString(input)
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}

func SanitizeContent(content string) string {
	content = strings.ReplaceAll(content, "<script", "&lt;script")
	content = strings.ReplaceAll(content, "<", "&lt;")
	content = strings.ReplaceAll(content, ">", "&gt;")
	content = strings.ReplaceAll(content, "\"", "&quot;")
	content = strings.ReplaceAll(content, "'", "&#39;")
	return content
}

func SanitizeEmotionalTags(text string) string {
	text = chineseParenRegex.ReplaceAllString(text, "<emotional>$1</emotional>")
	text = englishParenRegex.ReplaceAllString(text, "<emotional>$1</emotional>")
	return text
}
