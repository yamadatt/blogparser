package parser

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

// extractTitle はHTMLドキュメントからタイトルを抽出します。
// 以下の優先順位で抽出を試みます：
// 1. 最初のh1タグのテキスト
// 2. titleタグのテキスト
// 3. og:titleメタタグの内容
// 4. titleメタタグの内容
func extractTitle(doc *goquery.Document) (string, error) {
	if doc == nil {
		return "", errors.New("ドキュメントがnilです")
	}

	// 1. h1タグから抽出
	if h1 := doc.Find("h1").First(); h1.Length() > 0 {
		title := strings.TrimSpace(h1.Text())
		if title != "" {
			return title, nil
		}
	}

	// 2. titleタグから抽出
	if title := doc.Find("title").First(); title.Length() > 0 {
		text := strings.TrimSpace(title.Text())
		if text != "" {
			return text, nil
		}
	}

	// 3. og:titleメタタグから抽出
	if ogTitle, exists := doc.Find("meta[property='og:title']").Attr("content"); exists {
		title := strings.TrimSpace(ogTitle)
		if title != "" {
			return title, nil
		}
	}

	// 4. titleメタタグから抽出
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
