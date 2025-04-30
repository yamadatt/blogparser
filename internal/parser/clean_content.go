package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// CleanContent はHTMLコンテンツをクリーニングし、HTMLのまま返します。
func (p *HTMLParser) CleanContent(content string) string {
	// HTMLをパース
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		fmt.Println("HTMLのパースに失敗しました:", err)
		return content
	}

	// 不要なタグを削除
	doc.Find("script").Remove()
	doc.Find("style").Remove()
	doc.Find("iframe").Remove()
	doc.Find(".google-auto-placed").Remove()

	// アメブロ特有の不要な要素を削除
	doc.Find(".skin-entryBody, .skin-entryBody2").Each(func(i int, s *goquery.Selection) {
		// 広告関連の要素を削除
		s.Find(".google-auto-placed, .adsbygoogle, .blogroll-ad").Remove()
		// SNSボタン関連の要素を削除
		s.Find(".social-btn, .share-btn, .twitter-share-button").Remove()
	})

	// HTMLとして取得
	html, err := doc.Find("body").Html()
	if err != nil || html == "" {
		html, _ = doc.Html()
	}

	// 空白行を正規化（HTMLなので不要ならコメントアウト可）
	// html = p.normalizeWhitespace(html)

	return html
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
