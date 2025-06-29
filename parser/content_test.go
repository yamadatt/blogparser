package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractContent(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
		wantErr  bool
	}{
		{
			name: "articleタグからの抽出",
			html: `<html><body>
				<article>` + strings.Repeat("a", 100) + `</article>
			</body></html>`,
			expected: strings.Repeat("a", 100),
			wantErr:  false,
		},
		{
			name: "div.article-body-innerからの抽出",
			html: `<html><body>
				<div class="article-body-inner">` + strings.Repeat("b", 100) + `</div>
			</body></html>`,
			expected: strings.Repeat("b", 100),
			wantErr:  false,
		},
		{
			name: "div.skin-entryBodyからの抽出",
			html: `<html><body>
				<div class="skin-entryBody">` + strings.Repeat("c", 100) + `</div>
			</body></html>`,
			expected: strings.Repeat("c", 100),
			wantErr:  false,
		},
		{
			name: "mainタグからの抽出",
			html: `<html><body>
				<main>` + strings.Repeat("d", 100) + `</main>
			</body></html>`,
			expected: strings.Repeat("d", 100),
			wantErr:  false,
		},
		{
			name: "bodyタグからの抽出（最後の手段）",
			html: `<html><body>` + strings.Repeat("e", 100) + `</body></html>`,
			expected: strings.Repeat("e", 100),
			wantErr:  false,
		},
		{
			name: "複数セレクターがある場合は最初の有効なものを使用",
			html: `<html><body>
				<div class="article-body-inner">` + strings.Repeat("f", 100) + `</div>
				<article>` + strings.Repeat("g", 100) + `</article>
			</body></html>`,
			expected: strings.Repeat("f", 100),
			wantErr:  false,
		},
		{
			name: "短すぎるコンテンツは無効",
			html: `<html><body>
				<article>短い</article>
			</body></html>`,
			expected: "",
			wantErr:  true,
		},
		{
			name: "空のコンテンツ",
			html: `<html><body>
				<article></article>
			</body></html>`,
			expected: "",
			wantErr:  true,
		},
		{
			name: "該当するセレクターがない",
			html: `<html><body>
				<div class="unknown">` + strings.Repeat("h", 100) + `</div>
			</body></html>`,
			expected: `<div class="unknown">` + strings.Repeat("h", 100) + `</div>`,
			wantErr:  false,
		},
		{
			name:     "nilドキュメント",
			html:     "",
			expected: "",
			wantErr:  true,
		},
		{
			name: "HTMLパースエラーでも処理継続",
			html: `<html><body>
				<article><div>` + strings.Repeat("i", 100) + `</article>
			</body></html>`,
			expected: `<div>` + strings.Repeat("i", 100) + `</div>`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc *goquery.Document
			var err error
			
			if tt.html == "" {
				doc = nil
			} else {
				doc, err = goquery.NewDocumentFromReader(strings.NewReader(tt.html))
				if err != nil {
					t.Fatalf("HTMLのパースに失敗: %v", err)
				}
			}

			result, err := extractContent(doc)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("extractContent() error = nil, want error")
				}
				return
			}
			
			if err != nil {
				t.Errorf("extractContent() error = %v, want nil", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("extractContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "改行の正規化",
			input:    "<p>line1</p>\r\n<p>line2</p>\n\n<p>line3</p>",
			expected: "<p>line1</p>\n<p>line2</p>\n<p>line3</p>",
		},
		{
			name:     "空行の削除",
			input:    "<p>line1</p>\n\n\n<p>line2</p>\n\n",
			expected: "<p>line1</p>\n<p>line2</p>",
		},
		{
			name:     "前後の空白削除",
			input:    "  <p>content</p>  ",
			expected: "<p>content</p>",
		},
		{
			name:     "空文字列",
			input:    "",
			expected: "",
		},
		{
			name:     "空白のみ",
			input:    "   \n\n   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeHTML() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIsValidContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "有効なコンテンツ（100文字以上）",
			input:    strings.Repeat("a", 100),
			expected: true,
		},
		{
			name:     "有効なコンテンツ（200文字）",
			input:    strings.Repeat("b", 200),
			expected: true,
		},
		{
			name:     "無効なコンテンツ（空文字列）",
			input:    "",
			expected: false,
		},
		{
			name:     "無効なコンテンツ（短すぎる）",
			input:    "短い",
			expected: false,
		},
		{
			name:     "無効なコンテンツ（99文字）",
			input:    strings.Repeat("c", 99),
			expected: false,
		},
		{
			name:     "境界値（100文字ちょうど）",
			input:    strings.Repeat("d", 100),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidContent(tt.input)
			if result != tt.expected {
				t.Errorf("isValidContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}
