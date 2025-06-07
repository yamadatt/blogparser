package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	// 削除対象のタグ
	removeTags = []string{
		"script",
		"style",
		"iframe",
		".google-auto-placed",
		"dl.article-tags",
		"div.blogroll1",
		"div.rss2-title",
		"a[href*='newresu1.blog.fc2.com']",
		"div.ad-entry-bottom",
		"div.POST_TAIL",
		"hr[style*='191970']",
	}

	// アメブロ特有の削除対象要素
	amebloRemoveSelectors = map[string][]string{
		".skin-entryBody, .skin-entryBody2": {
			// 広告関連の要素
			".google-auto-placed",
			".adsbygoogle",
			".blogroll-ad",
			// SNSボタン関連の要素
			".social-btn",
			".share-btn",
			".twitter-share-button",
		},
	}

	// 正規表現による削除パターン
	regexPatterns = []struct {
		pattern     string
		description string
	}{
		{`<!--[\s\S]*?-->`, "HTMLコメントを削除"},
		{`[１-９一二三四五六七八九十]位：`, "順位表記を削除"},
	}
)

// CleanContent はHTMLコンテンツをクリーニングし、HTMLのまま返します。
func (p *HTMLParser) CleanContent(content string) (string, error) {
	if content == "" {
		return "", ErrEmptyContent
	}

	// 正規表現による削除
	content = p.removeByRegex(content)

	// HTMLをパース
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrParseHTML, err)
	}

	// 不要なタグを削除
	for _, selector := range removeTags {
		doc.Find(selector).Remove()
	}

	// アメブロ特有の不要な要素を削除
	for parentSelector, childSelectors := range amebloRemoveSelectors {
		doc.Find(parentSelector).Each(func(i int, s *goquery.Selection) {
			for _, selector := range childSelectors {
				s.Find(selector).Remove()
			}
		})
	}

	// HTMLとして取得
	html, err := doc.Find("body").Html()
	if err != nil {
		html, err = doc.Html()
		if err != nil {
			return "", fmt.Errorf("HTMLの生成に失敗しました: %w", err)
		}
	}

	if html == "" {
		return "", ErrEmptyContent
	}

	return html, nil
}

// removeByRegex は正規表現パターンに基づいてコンテンツを削除します
func (p *HTMLParser) removeByRegex(content string) string {
	for _, pattern := range regexPatterns {
		re := regexp.MustCompile(pattern.pattern)
		content = re.ReplaceAllString(content, "")
	}
	return content
}

// normalizeWhitespace は空白や改行を正規化します
func (p *HTMLParser) normalizeWhitespace(content string) string {
	// 連続する改行を1つの改行に
	re := regexp.MustCompile(`\n\s*\n`)
	content = re.ReplaceAllString(content, "\n")

	// 行頭と行末の空白を削除
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}

	// 空の行を削除
	var nonEmptyLines []string
	for _, line := range lines {
		if line != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	// 改行で結合
	content = strings.Join(nonEmptyLines, "\n")

	return content
}
