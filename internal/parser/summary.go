package parser

import (
	"regexp"
	"strings"
)

// GenerateSummary は記事本文からサマリ（要約）を生成します。
// 現状は最初の2文をサマリとします。
func (p *HTMLParser) GenerateSummary(content string) string {
	// 改行やHTMLタグを除去し、ピリオド・句点で分割
	plain := stripHTMLTags(content)
	plain = strings.ReplaceAll(plain, "\n", "")
	plain = strings.ReplaceAll(plain, "\r", "")

	sentences := splitSentences(plain)
	if len(sentences) == 0 {
		return ""
	}
	if len(sentences) == 1 {
		return sentences[0]
	}
	return sentences[0] + sentences[1]
}

// stripHTMLTags はHTMLタグを除去します
func stripHTMLTags(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(html, "")
}

// splitSentences は句点（。や.）で文を分割します
func splitSentences(text string) []string {
	re := regexp.MustCompile(`.*?[。.]`)
	return re.FindAllString(text, -1)
}
