package parser

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

// extractTitle はHTMLドキュメントからタイトルを抽出します。
// 以下の優先順位で抽出を試みます：
// 1. ld_blog_varsのarticles[0].title
// 2. og:titleメタタグの内容
// 3. 最初のh1タグのテキスト
// 4. titleタグのテキスト
// 5. titleメタタグの内容
func extractTitle(doc *goquery.Document) (string, error) {
	if doc == nil {
		return "", errors.New("ドキュメントがnilです")
	}

	// 1. ld_blog_varsからタイトルを抽出
	var foundTitle string
	doc.Find("script").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if script := s.Text(); strings.Contains(script, "ld_blog_vars") {
			// タイトルを抽出するための正規表現
			re := regexp.MustCompile(`articles\s*:\s*\[\s*\{\s*[^}]*title\s*:\s*'([^']*)'`)
			if matches := re.FindStringSubmatch(script); len(matches) > 1 {
				foundTitle = strings.TrimSpace(matches[1])
				return false // 検索を終了
			}
		}
		return true // 検索を続行
	})
	if foundTitle != "" {
		return foundTitle, nil
	}

	// 2. og:titleメタタグから抽出
	if ogTitle, exists := doc.Find("meta[property='og:title']").Attr("content"); exists {
		title := strings.TrimSpace(ogTitle)
		if title != "" {
			return title, nil
		}
	}

	// 3. h1タグから抽出
	if h1 := doc.Find("h1").First(); h1.Length() > 0 {
		title := strings.TrimSpace(h1.Text())
		if title != "" {
			return title, nil
		}
	}

	// 4. titleタグから抽出
	if title := doc.Find("title").First(); title.Length() > 0 {
		text := strings.TrimSpace(title.Text())
		if text != "" {
			return text, nil
		}
	}

	// 5. titleメタタグから抽出
	if metaTitle, exists := doc.Find("meta[name='title']").Attr("content"); exists {
		title := strings.TrimSpace(metaTitle)
		if title != "" {
			return title, nil
		}
	}

	return "", errors.New("タイトルが見つかりません")
}

// cleanTitle はタイトルテキストを整形します。
func cleanTitle(title string) string {
	// 改行を削除
	title = strings.ReplaceAll(title, "\n", " ")
	// 連続する空白を1つに
	title = strings.Join(strings.Fields(title), " ")
	// ダブルクォーテーションをエスケープ
	title = strings.ReplaceAll(title, "\"", "\\\"")
	// 特定の文字列を削除
	title = strings.ReplaceAll(title, " | 心理カウンセラー・中井亜紀『成長の記録』", "")
	// 前後の空白を削除
	return strings.TrimSpace(title)
}

// isValidTitle はタイトルが有効かどうかを判定します。
func isValidTitle(title string) bool {
	// 空文字列でないこと
	if title == "" {
		return false
	}

	// HTMLタグを含まないこと
	if strings.Contains(title, "<") || strings.Contains(title, ">") {
		return false
	}

	// 制御文字を含まないこと
	for _, r := range title {
		if r < ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}

	return true
}
