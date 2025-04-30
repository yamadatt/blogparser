package parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/yourusername/blogparser/pkg/models"
)

// Parser はブログ記事をパースするためのインターフェースです。
type Parser interface {
	// ParseFiles は複数のファイルパスからブログ記事を解析します。
	ParseFiles(paths []string) ([]*models.BlogPost, error)
	// Parse はio.Readerからブログ記事を解析します。
	Parse(r io.Reader) (*models.BlogPost, error)
}

// HTMLParser はHTMLファイルからブログ記事を解析するパーサーです。
type HTMLParser struct{}

// New は新しいHTMLParserを作成します。
func New() Parser {
	return &HTMLParser{}
}

// ParseFiles は複数のファイルパスからブログ記事を解析します。
func (p *HTMLParser) ParseFiles(paths []string) ([]*models.BlogPost, error) {
	var posts []*models.BlogPost
	var errMsgs []string

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			errMsgs = append(errMsgs, fmt.Sprintf("ファイル %s を開けません: %v", path, err))
			continue
		}

		post, err := p.Parse(f)
		f.Close()
		if err != nil {
			errMsgs = append(errMsgs, fmt.Sprintf("ファイル %s の解析に失敗: %v", path, err))
			continue
		}

		// ファイル名（拡張子あり）をSlugにセット
		post.Slug = filepath.Base(path)

		posts = append(posts, post)
	}

	if len(errMsgs) > 0 {
		return posts, errors.Errorf("%d 件のエラーが発生しました:\n- %s",
			len(errMsgs), strings.Join(errMsgs, "\n- "))
	}

	return posts, nil
}

// Parse はio.Readerからブログ記事を解析します。
func (p *HTMLParser) Parse(r io.Reader) (*models.BlogPost, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, errors.Wrap(err, "HTMLのパースに失敗しました")
	}

	title, err := extractTitle(doc)
	if err != nil {
		return nil, errors.Wrap(err, "タイトルの抽出に失敗しました")
	}

	title = cleanTitle(title)
	if !isValidTitle(title) {
		return nil, errors.New("無効なタイトルです")
	}

	content, err := extractContent(doc)
	if err != nil {
		return nil, errors.Wrap(err, "コンテンツの抽出に失敗しました")
	}

	if !isValidContent(content) {
		return nil, errors.New("無効なコンテンツです")
	}

	categories, err := extractCategories(doc)
	if err != nil {
		return nil, errors.Wrap(err, "カテゴリの抽出に失敗しました")
	}

	// カテゴリの検証
	var validCategories []string
	for _, category := range categories {
		category = cleanCategory(category)
		if isValidCategory(category) {
			validCategories = append(validCategories, category)
		}
	}

	tags, err := extractTags(doc)
	if err != nil {
		return nil, errors.Wrap(err, "タグの抽出に失敗しました")
	}

	var validTags []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" && !containsString(validTags, tag) {
			validTags = append(validTags, tag)
		}
	}

	createdAt, err := extractDate(doc)
	if err != nil {
		createdAt = time.Time{} // 日付が見つからない場合はゼロ値
	}

	post := &models.BlogPost{
		Title:      title,
		Content:    content,
		Categories: validCategories,
		Tags:       validTags,
		CreatedAt:  createdAt,
	}

	return post, nil
}
