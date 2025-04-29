package parser

import (
	"io"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/yourusername/blogparser/pkg/models"
)

// Parser はブログ記事をパースするためのインターフェースです。
type Parser interface {
	// ParseFile はファイルパスからブログ記事を解析します。
	ParseFile(path string) (*models.BlogPost, error)
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

// ParseFile はファイルパスからブログ記事を解析します。
func (p *HTMLParser) ParseFile(path string) (*models.BlogPost, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "ファイル %s を開けません", path)
	}
	defer f.Close()

	return p.Parse(f)
}

// ParseFiles は複数のファイルパスからブログ記事を解析します。
func (p *HTMLParser) ParseFiles(paths []string) ([]*models.BlogPost, error) {
	var posts []*models.BlogPost
	var errs []error

	for _, path := range paths {
		post, err := p.ParseFile(path)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "ファイル %s の解析に失敗しました", path))
			continue
		}
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

	return &models.BlogPost{
		Title: title,
	}, nil
}
