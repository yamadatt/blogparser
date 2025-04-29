package models

import "time"

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
