package parser

import (
	"strings"
	"testing"
)

func TestNormalizeImageURL(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"data:image/png;base64,xyz", ""},
		{"http://example.com/img.jpg", "http://example.com/img.jpg"},
		{"https://stat.ameblo.jp/foo_s.jpg", "https://stat.ameblo.jp/foo.jpg"},
		{":bad url", ""},
	}
	for _, c := range cases {
		got := normalizeImageURL(c.in)
		if got != c.want {
			t.Errorf("normalizeImageURL(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestGetFirstImage(t *testing.T) {
	html := `<meta property='og:image' content='http://ex.com/og.jpg'>` +
		`<img src='http://ex.com/a.jpg'><img src='http://ex.com/b.jpg'>`
	p := &HTMLParser{}
	first := p.GetFirstImage(html)
	if first != "http://ex.com/og.jpg" {
		t.Errorf("first image=%q", first)
	}

	html2 := `<img src='a.jpg'><img src='b.jpg'>`
	if p.GetFirstImage(html2) != "a.jpg" {
		t.Errorf("GetFirstImage normal failed")
	}

	if p.GetFirstImage("") != "" {
		t.Errorf("expected empty for no image")
	}
}

func TestExtractImages(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected int // 期待される画像数
	}{
		{
			name: "複数の画像",
			html: `<html><body>
				<img src="image1.jpg" alt="画像1">
				<img src="image2.png" alt="画像2">
				<img src="image3.gif" alt="画像3">
			</body></html>`,
			expected: 3,
		},
		{
			name: "data URLは除外",
			html: `<html><body>
				<img src="data:image/png;base64,xyz" alt="データ画像">
				<img src="image1.jpg" alt="通常画像">
			</body></html>`,
			expected: 1,
		},
		{
			name: "無効なURLは除外",
			html: `<html><body>
				<img src=":invalid" alt="無効画像">
				<img src="image1.jpg" alt="有効画像">
			</body></html>`,
			expected: 1,
		},
		{
			name: "画像なし",
			html: `<html><body>
				<p>テキストのみ</p>
			</body></html>`,
			expected: 0,
		},
		{
			name: "空のsrc属性",
			html: `<html><body>
				<img src="" alt="空画像">
				<img src="image1.jpg" alt="有効画像">
			</body></html>`,
			expected: 1,
		},
		{
			name: "src属性なし",
			html: `<html><body>
				<img alt="src属性なし">
				<img src="image1.jpg" alt="有効画像">
			</body></html>`,
			expected: 1,
		},
	}

	p := &HTMLParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			images := p.ExtractImages(tt.html)
			if len(images) != tt.expected {
				t.Errorf("ExtractImages() returned %d images, want %d", len(images), tt.expected)
			}
		})
	}
}

func TestGetFirstImageEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "og:imageが優先される",
			html: `<html><head>
				<meta property="og:image" content="og-image.jpg">
			</head><body>
				<img src="first-img.jpg">
			</body></html>`,
			expected: "og-image.jpg",
		},
		{
			name: "twitter:imageが使用される",
			html: `<html><head>
				<meta name="twitter:image" content="twitter-image.jpg">
			</head><body>
				<img src="first-img.jpg">
			</body></html>`,
			expected: "twitter-image.jpg",
		},
		{
			name: "og:imageもtwitter:imageもない場合は最初のimg",
			html: `<html><body>
				<img src="first-img.jpg">
				<img src="second-img.jpg">
			</body></html>`,
			expected: "first-img.jpg",
		},
		{
			name: "アメブロのサムネイル正規化",
			html: `<html><body>
				<img src="https://stat.ameblo.jp/user_images/20230101/12/test/ab/cd/j/o0480047014261879529_s.jpg">
			</body></html>`,
			expected: "https://stat.ameblo.jp/user_images/20230101/12/test/ab/cd/j/o0480047014261879529.jpg",
		},
		{
			name: "無効なHTMLでも処理継続",
			html: `<img src="image.jpg"><div><span>`,
			expected: "image.jpg",
		},
		{
			name: "非常に長いHTML",
			html: `<html><body>` + strings.Repeat("<p>テキスト</p>", 1000) + `<img src="deep-image.jpg"></body></html>`,
			expected: "deep-image.jpg",
		},
	}

	p := &HTMLParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.GetFirstImage(tt.html)
			if result != tt.expected {
				t.Errorf("GetFirstImage() = %v, want %v", result, tt.expected)
			}
		})
	}
}
