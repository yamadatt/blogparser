package parser

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

// extractCategories はHTMLドキュメントからカテゴリを抽出します。
// 以下の優先順位で抽出を試みます：
// 1. 一般的なブログプラットフォームのセレクタ
// 2. ld_blog_varsのarticles[0].categories
// 3. meta[property='article:section']
// 4. .category クラスを持つ要素
func extractCategories(doc *goquery.Document) ([]string, error) {
	if doc == nil {
		return nil, errors.New("ドキュメントがnilです")
	}

	var categories []string

	// 1. 一般的なブログプラットフォームのセレクタから抽出
	selectors := []string{
		// アメブロ固有
		".skin-categoryLabel",                      // アメブロのカテゴリラベル
		"[data-uranus-component='theme']",          // 別のパターン
		".skin-entryThemes a",                      // 新しいアメブロのテーマリンク
		".skin-categoryTag",                        // カテゴリタグ
		"[data-analytics-index-name='theme'] span", // データ属性を使ったテーマ
		"div.theme a",                              // テーマがdivクラスに入っている場合
		".skinTheme",                               // スキンテーマクラス
		"li.theme a",                               // リスト要素のテーマ
		".subHeader-theme",                         // サブヘッダーのテーマ
		"a.theme-link",                             // テーマリンク
		"dd.article-category1",                     //livedoorのカテゴリ
		"dd.article-category2",                     //livedoorのカテゴリ

		// エキサイトブログ固有
		".POST_TAIL .TIME a[href*=\"/i\"]", // エキサイトブログのカテゴリリンク
		".articleTheme",                    // 別のパターン
		"a[rel='category']",
		".category a",
		".cat-links a",
		".entry-categories a",
		".post-categories a",
		"[itemprop='articleSection']",
		".tags a", // カテゴリとタグの区別がない場合の対応

		// 一般的なブログプラットフォーム
		"a[rel='category tag']", // カテゴリタグリンク
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			category := strings.TrimSpace(s.Text())
			if category != "" && !containsString(categories, category) {
				categories = append(categories, category)
			}
		})
	}

	// セレクタからカテゴリが見つかった場合は返す
	if len(categories) > 0 {
		return categories, nil
	}

	// 2. ld_blog_varsからカテゴリを抽出
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		if script := s.Text(); strings.Contains(script, "ld_blog_vars") {
			// カテゴリを抽出するための正規表現
			re := regexp.MustCompile(`categories\s*:\s*\[\s*\{\s*[^}]*name\s*:\s*'([^']*)'`)
			matches := re.FindAllStringSubmatch(script, -1)
			for _, match := range matches {
				if len(match) > 1 {
					category := strings.TrimSpace(match[1])
					if category != "" && !containsString(categories, category) {
						categories = append(categories, category)
					}
				}
			}
		}
	})

	// ld_blog_varsからカテゴリが見つかった場合は返す
	if len(categories) > 0 {
		return categories, nil
	}

	// 3. meta[property='article:section']から抽出
	doc.Find("meta[property='article:section']").Each(func(i int, s *goquery.Selection) {
		if category, exists := s.Attr("content"); exists {
			category = strings.TrimSpace(category)
			if category != "" && !containsString(categories, category) {
				categories = append(categories, category)
			}
		}
	})

	// 4. .category クラスから抽出
	doc.Find(".category").Each(func(i int, s *goquery.Selection) {
		category := strings.TrimSpace(s.Text())
		if category != "" && !containsString(categories, category) {
			categories = append(categories, category)
		}
	})

	// カテゴリが1つも見つからなかった場合はエラー
	if len(categories) == 0 {
		return nil, errors.New("カテゴリが見つかりません")
	}

	return categories, nil
}

// cleanCategory はカテゴリ名を整形します。
func cleanCategory(category string) string {
	// 改行を削除
	category = strings.ReplaceAll(category, "\r\n", "\n")
	category = strings.ReplaceAll(category, "\r", "\n")
	category = strings.ReplaceAll(category, "\n", " ")

	// 連続する空白を1つに
	category = strings.Join(strings.Fields(category), " ")

	// 不要なプレフィックスのリスト
	prefixes := []string{
		"テーマ：",   // アメブロのテーマプレフィックス
		"テーマ:",   // コロンの全角/半角両方に対応
		"Theme：", // 英語表記の可能性
		"Theme:", // 英語表記（半角コロン）
	}

	// プレフィックスを削除
	for _, prefix := range prefixes {
		if strings.HasPrefix(category, prefix) {
			category = strings.TrimPrefix(category, prefix)
			break // 一つのプレフィックスを削除したら終了
		}
	}

	// 前後の空白を削除
	return strings.TrimSpace(category)
}

// isValidCategory はカテゴリが有効かどうかを判定します。
func isValidCategory(category string) bool {
	// 空文字列でないこと
	if category == "" {
		return false
	}

	// HTMLタグを含まないこと
	if strings.Contains(category, "<") || strings.Contains(category, ">") {
		return false
	}

	// 制御文字を含まないこと
	for _, r := range category {
		if r < ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}

	return true
}

// containsString は文字列スライスに特定の文字列が含まれているかを確認します。
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
