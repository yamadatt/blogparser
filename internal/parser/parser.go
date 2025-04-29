package parser

import (
	"io"
	"os"
	"path/filepath"
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
	var errs []error

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "ファイル %s を開けません", path))
			continue
		}

		post, err := p.Parse(f)
		f.Close()
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "ファイル %s の解析に失敗しました", path))
			continue
		}

		// ファイル名（拡張子あり）をSlugにセット
		post.Slug = filepath.Base(path)

		posts = append(posts, post)
	}

	if len(errs) > 0 {
		return posts, errors.Errorf("%d 件のエラーが発生しました", len(errs))
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

	createdAt, err := extractDate(doc)
	if err != nil {
		createdAt = time.Time{} // 日付が見つからない場合はゼロ値
	}

	post := &models.BlogPost{
		Title:     title,
		CreatedAt: createdAt,
	}

	return post, nil
}
