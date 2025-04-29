package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

// extractContent はHTMLドキュメントから記事の本文を抽出します。
// 以下の優先順位で抽出を試みます：
// 1. article タグ内のコンテンツ
// 2. main タグ内のコンテンツ
// 3. .content または .article クラスを持つ要素のコンテンツ
func extractContent(doc *goquery.Document) (string, error) {
	if doc == nil {
		return "", errors.New("ドキュメントがnilです")
	}

	var extractionAttempts []string

	// よく使用されるブログプラットフォームのセレクター
	selectors := []string{
		"div.article-body-inner",
		"div.skin-entryBody",
		"div.articleText",
		"div.post-main",
		"div.post-body",
		"div.entry-content",
		"div.POST_BODY",
		"article",
		"[itemprop='articleBody']",
		".entry-content",
		".post-content",
		".article-content",
		"#content",
		"#main-content",
		".content",
	}

	// 指定されたセレクターで抽出を試みる
	for _, selector := range selectors {
		if element := doc.Find(selector).First(); element.Length() > 0 {
			html, err := element.Html()
			if err != nil {
				extractionAttempts = append(extractionAttempts,
					fmt.Sprintf("%s: HTMLの抽出に失敗: %v", selector, err))
				continue
			}

			content := normalizeHTML(html)
			if content != "" {
				if isValidContent(content) {
					return content, nil
				}
				extractionAttempts = append(extractionAttempts,
					fmt.Sprintf("%s: コンテンツが無効です", selector))
			} else {
				extractionAttempts = append(extractionAttempts,
					fmt.Sprintf("%s: コンテンツが空です", selector))
			}
		} else {
			extractionAttempts = append(extractionAttempts,
				fmt.Sprintf("%s: 見つかりません", selector))
		}
	}

	// main タグから抽出
	if main := doc.Find("main").First(); main.Length() > 0 {
		html, err := main.Html()
		if err != nil {
			extractionAttempts = append(extractionAttempts,
				fmt.Sprintf("main タグ: HTMLの抽出に失敗: %v", err))
		} else {
			content := normalizeHTML(html)
			if content != "" {
				if isValidContent(content) {
					return content, nil
				}
				extractionAttempts = append(extractionAttempts, "main タグ: コンテンツが無効です")
			} else {
				extractionAttempts = append(extractionAttempts, "main タグ: コンテンツが空です")
			}
		}
	} else {
		extractionAttempts = append(extractionAttempts, "main タグ: 見つかりません")
	}

	// bodyから抽出（最後の手段）
	body := doc.Find("body").First()
	if body.Length() > 0 {
		html, err := body.Html()
		if err != nil {
			extractionAttempts = append(extractionAttempts,
				fmt.Sprintf("body タグ: HTMLの抽出に失敗: %v", err))
		} else {
			content := normalizeHTML(html)
			if content != "" {
				if isValidContent(content) {
					return content, nil
				}
				extractionAttempts = append(extractionAttempts, "body タグ: コンテンツが無効です")
			} else {
				extractionAttempts = append(extractionAttempts, "body タグ: コンテンツが空です")
			}
		}
	} else {
		extractionAttempts = append(extractionAttempts, "body タグ: 見つかりません")
	}

	return "", errors.Errorf("コンテンツ抽出に失敗しました。試行結果:\n%s", strings.Join(extractionAttempts, "\n- "))
}

// normalizeHTML はHTML文字列を整形します。
func normalizeHTML(html string) string {
	// 改行の正規化
	html = strings.ReplaceAll(html, "\r\n", "\n")
	html = strings.ReplaceAll(html, "\r", "\n")

	// 連続する空行を1つに
	lines := strings.Split(html, "\n")
	var cleanedLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}

	// 行を結合して1つの文字列に
	html = strings.Join(cleanedLines, "\n")

	// 前後の空白を削除
	return strings.TrimSpace(html)
}

// isValidContent はコンテンツが有効かどうかを判定します。
func isValidContent(content string) bool {
	// 空文字列でないこと
	if content == "" {
		return false
	}

	// 最小文字数（例：100文字）を満たすこと
	if len(content) < 100 {
		return false
	}

	return true
}
