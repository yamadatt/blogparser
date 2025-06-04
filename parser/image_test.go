package parser

import "testing"

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
