package models

import (
	"regexp"
	"strings"
	"time"
	"unicode"
)

// BlogPostはブログ記事を表現する構造体です。
type BlogPost struct {
	Title      string    // タイトル
	Author     string    // 著者名
	Content    string    // 本文
	Summary    string    // 要約
	Tags       []string  // タグ
	Categories []string  // カテゴリ
	CreatedAt  time.Time // 作成日時
	UpdatedAt  time.Time // 更新日時
	Published  bool      // 公開フラグ
	Slug       string    // URL用スラッグ
	FirstImage string    // 記事内で最初に登場する画像のURL
}

// SetSlug はTitleからSlugを生成してセットするメソッド
func (b *BlogPost) SetSlug() {
	slug := strings.ToLower(b.Title)
	slug = removeNonASCII(slug)

	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")

	re2 := regexp.MustCompile(`-+`)
	slug = re2.ReplaceAllString(slug, "-")

	slug = strings.Trim(slug, "-")

	b.Slug = slug
}

// removeNonASCII はASCII以外の文字を除去する
func removeNonASCII(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r <= unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
