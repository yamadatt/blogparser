package parser

import (
	"context"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

type parseTest struct {
	file       string
	title      string
	length     int
	categories []string
	tags       []string
	tagCount   int
	firstImage string
	createdAt  time.Time
}

func TestParseFileSamples(t *testing.T) {
	tz := time.FixedZone("JST", 9*3600)
	tests := []parseTest{
		{
			file:       filepath.Join("..", "sample", "test", "testdata", "12403291408.html"),
			title:      "『●あなた仕様にカスタマイズ！！『思い込み』書き換え・マンツーマン講座』",
			length:     43317,
			categories: []string{"マンツーマン講座"},
			tags:       []string{"不登校"},
			tagCount:   1,
			firstImage: "https://stat.ameba.jp/user_images/20180907/17/akinakai/eb/9a/j/o0480047014261879529.jpg",
			createdAt:  time.Date(2024, 5, 22, 12, 39, 1, 0, tz),
		},
		{
			file:       filepath.Join("..", "sample", "test", "testdata", "12887862927.html"),
			title:      "『ルーティーン』",
			length:     18409,
			categories: []string{"ブログ"},
			tags:       []string{"認知症介護", "認知症の母", "認知症"},
			tagCount:   3,
			firstImage: "https://stat.ameba.jp/user_images/20250412/13/macb2b37/d3/da/j/o1024102415565487103.jpg",
			createdAt:  time.Date(2025, 4, 13, 18, 18, 5, 0, tz),
		},
		{
			file:       filepath.Join("..", "sample", "test", "testdata", "16274503.html"),
			title:      "月山に思いを馳せる満月の夜",
			length:     11715,
			categories: nil,
			tags:       nil,
			tagCount:   0,
			firstImage: "https://pds.exblog.jp/pds/1/201109/12/14/b0207514_21282826.jpg",
			createdAt:  time.Date(2011, 9, 12, 23, 31, 0, 0, tz),
		},
		{
			file:       filepath.Join("..", "sample", "test", "testdata", "9994362.html"),
			title:      "【衝撃】最近、某宗教団体が窃盗を働いているという噂があった。そんなある日、Aさん宅の玄関先でその宗教の人達が勧誘していた→嫌だなと思いながらA宅の角を曲がると･･･",
			length:     9574,
			categories: []string{"セコママ・泥ママ", "キチママ"},
			tags:       nil,
			tagCount:   68,
			firstImage: "https://parts.blog.livedoor.jp/img/usr/cmn/ogp_image/livedoor.png",
			createdAt:  time.Date(2018, 6, 17, 2, 17, 45, 0, tz),
		},
	}

	p := New()
	ctx := context.Background()
	for _, tt := range tests {
		post, err := p.ParseFile(ctx, tt.file)
		if err != nil {
			t.Fatalf("ParseFile(%s) error: %v", tt.file, err)
		}
		if post.Title != tt.title {
			t.Errorf("%s title=%q want %q", tt.file, post.Title, tt.title)
		}
		if len(post.Content) != tt.length {
			t.Errorf("%s length=%d want %d", tt.file, len(post.Content), tt.length)
		}
		if !reflect.DeepEqual(post.Categories, tt.categories) {
			t.Errorf("%s categories=%v want %v", tt.file, post.Categories, tt.categories)
		}
		if tt.tags != nil {
			if !reflect.DeepEqual(post.Tags, tt.tags) {
				t.Errorf("%s tags=%v want %v", tt.file, post.Tags, tt.tags)
			}
		} else if len(post.Tags) != tt.tagCount {
			t.Errorf("%s tag count=%d want %d", tt.file, len(post.Tags), tt.tagCount)
		}
		if post.FirstImage != tt.firstImage {
			t.Errorf("%s first image=%q want %q", tt.file, post.FirstImage, tt.firstImage)
		}
		if !post.CreatedAt.Equal(tt.createdAt) {
			t.Errorf("%s createdAt=%v want %v", tt.file, post.CreatedAt, tt.createdAt)
		}
	}
}

func TestParseFileErrors(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "存在しないファイル",
			path:    "nonexistent.html",
			wantErr: true,
		},
		{
			name:    "ディレクトリを指定",
			path:    ".",
			wantErr: true,
		},
	}

	p := New()
	ctx := context.Background()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.ParseFile(ctx, tt.path)
			if tt.wantErr && err == nil {
				t.Errorf("ParseFile() error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ParseFile() error = %v, want nil", err)
			}
		})
	}
}

func TestParseFileWithCancelledContext(t *testing.T) {
	p := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // コンテキストをキャンセル

	_, err := p.ParseFile(ctx, "test.html")
	if err == nil {
		t.Error("ParseFile() with cancelled context should return error")
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		wantErr bool
	}{
		{
			name: "有効なHTML",
			html: `<html>
				<head><title>テストタイトル</title></head>
				<body>
					<article>` + strings.Repeat("テスト本文", 50) + `</article>
				</body>
			</html>`,
			wantErr: false,
		},
		{
			name: "タイトルなし",
			html: `<html>
				<body>
					<article>` + strings.Repeat("テスト本文", 50) + `</article>
				</body>
			</html>`,
			wantErr: true,
		},
		{
			name: "コンテンツなし",
			html: `<html>
				<head><title>テストタイトル</title></head>
				<body></body>
			</html>`,
			wantErr: true,
		},
		{
			name: "短すぎるコンテンツ",
			html: `<html>
				<head><title>テストタイトル</title></head>
				<body><article>短い</article></body>
			</html>`,
			wantErr: true,
		},
		{
			name: "無効なタイトル",
			html: `<html>
				<head><title>   </title></head>
				<body>
					<article>` + strings.Repeat("テスト本文", 50) + `</article>
				</body>
			</html>`,
			wantErr: true,
		},
	}

	p := New()
	ctx := context.Background()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.html)
			_, err := p.Parse(ctx, reader)
			
			if tt.wantErr && err == nil {
				t.Errorf("Parse() error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Parse() error = %v, want nil", err)
			}
		})
	}
}

func TestParseWithCancelledContext(t *testing.T) {
	p := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // コンテキストをキャンセル

	html := `<html><head><title>テスト</title></head><body><article>` + strings.Repeat("テスト", 50) + `</article></body></html>`
	reader := strings.NewReader(html)
	
	_, err := p.Parse(ctx, reader)
	if err == nil {
		t.Error("Parse() with cancelled context should return error")
	}
}

func TestParseInvalidHTML(t *testing.T) {
	p := New()
	ctx := context.Background()
	
	// 無効なHTMLでもgoqueryは処理できるため、エラーにはならない
	invalidHTML := `<html><head><title>テスト</title></head><body><article>` + strings.Repeat("テスト", 50)
	reader := strings.NewReader(invalidHTML)
	
	post, err := p.Parse(ctx, reader)
	if err != nil {
		t.Errorf("Parse() with invalid HTML should not error, got: %v", err)
	}
	if post == nil {
		t.Error("Parse() should return a post even with invalid HTML")
	}
}

// ファイルクローズエラーのテスト用のモックReader
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestParseWithReadError(t *testing.T) {
	p := New()
	ctx := context.Background()
	
	reader := &errorReader{}
	_, err := p.Parse(ctx, reader)
	if err == nil {
		t.Error("Parse() with read error should return error")
	}
}

func TestNew(t *testing.T) {
	parser := New()
	if parser == nil {
		t.Error("New() should return non-nil parser")
	}
	
	// HTMLParserの型チェック
	if _, ok := parser.(*HTMLParser); !ok {
		t.Error("New() should return *HTMLParser")
	}
}
