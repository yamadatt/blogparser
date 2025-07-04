package parser

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/yamadatt/blogparser/pkg/models"
	"go.uber.org/zap"
)

// Parser はブログ記事をパースするためのインターフェースです。
type Parser interface {
	// ParseFile はファイルパスからブログ記事を解析します。
	ParseFile(ctx context.Context, path string) (*models.BlogPost, error)
	// Parse はio.Readerからブログ記事を解析します。
	Parse(ctx context.Context, r io.Reader) (*models.BlogPost, error)
}

// HTMLParser はHTMLファイルからブログ記事を解析するパーサーです。
type HTMLParser struct {
	logger *zap.Logger
}

// New は新しいHTMLParserを作成します。
func New() Parser {
	return &HTMLParser{
		logger: zap.NewNop(),
	}
}

// ParseFile はファイルパスからブログ記事を解析します。
func (p *HTMLParser) ParseFile(ctx context.Context, path string) (*models.BlogPost, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ファイル %s を開けません: %w", path, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			p.logger.Warn("ファイルのクローズに失敗しました", zap.String("path", path), zap.Error(closeErr))
		}
	}()

	post, err := p.Parse(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("ファイル %s の解析に失敗: %w", path, err)
	}

	post.Slug = filepath.Base(path)
	return post, nil
}

// Parse はio.Readerからブログ記事を解析します。
func (p *HTMLParser) Parse(ctx context.Context, r io.Reader) (*models.BlogPost, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("HTMLのパースに失敗しました: %w", err)
	}

	title, err := extractTitle(doc)
	if err != nil {
		return nil, fmt.Errorf("タイトルの抽出に失敗しました: %w", err)
	}

	title = cleanTitle(title)
	if !isValidTitle(title) {
		return nil, errors.New("無効なタイトルです")
	}

	content, err := extractContent(doc)
	if err != nil {
		return nil, fmt.Errorf("コンテンツの抽出に失敗しました: %w", err)
	}

	// コンテンツのクリーニング
	content, err = p.CleanContent(content)
	if err != nil {
		return nil, fmt.Errorf("コンテンツのクリーニングに失敗しました: %w", err)
	}

	// サマリ生成
	summary, err := p.GenerateSummary(content)
	if err != nil {
		return nil, fmt.Errorf("サマリの生成に失敗しました: %w", err)
	}

	if !isValidContent(content) {
		return nil, errors.New("無効なコンテンツです")
	}

	categories, err := extractCategories(doc)
	if err != nil {
		return nil, fmt.Errorf("カテゴリの抽出に失敗しました: %w", err)
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
		return nil, fmt.Errorf("タグの抽出に失敗しました: %w", err)
	}

	var validTags []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" && !slices.Contains(validTags, tag) {
			validTags = append(validTags, tag)
		}
	}

	createdAt, err := extractDate(doc)
	if err != nil {
		createdAt = time.Time{} // 日付が見つからない場合はゼロ値
	}

	html, _ := doc.Html()
	images := p.ExtractImages(html)
	firstImage := ""
	if len(images) > 0 {
		firstImage = images[0].URL
	}

	post := &models.BlogPost{
		Title:      title,
		Content:    content,
		Summary:    summary,
		Categories: validCategories,
		Tags:       validTags,
		CreatedAt:  createdAt,
		FirstImage: firstImage,
	}

	return post, nil
}
