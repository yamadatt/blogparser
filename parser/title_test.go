package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestCleanTitle(t *testing.T) {
	cases := []struct{ in, want string }{
		{"  サンプル\nタイトル  | 心理カウンセラー・中井亜紀『成長の記録』", "サンプル タイトル"},
		{"\"quoted\"", "\\\"quoted\\\""},
		{" multiple   spaces ", "multiple spaces"},
	}
	for _, c := range cases {
		got := cleanTitle(c.in)
		if got != c.want {
			t.Errorf("cleanTitle(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}

func TestIsValidTitle(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"Valid Title", true},
		{"", false},
		{"<b>bad</b>", false},
		{"bad\x01", false},
	}
	for _, c := range cases {
		got := isValidTitle(c.in)
		if got != c.want {
			t.Errorf("isValidTitle(%q) = %v; want %v", c.in, got, c.want)
		}
	}
}

func TestExtractTitle(t *testing.T) {
	cases := []struct{ html, want string }{
		{`<script>var ld_blog_vars={articles:[{title:'Script Title'}]};</script>` +
			`<meta property='og:title' content='OG Title'>` +
			`<h1>H1 Title</h1>` +
			`<title>Doc Title</title>` +
			`<meta name='title' content='Meta Title'>`, "Script Title"},
		{`<meta property='og:title' content='OG Title'><h1>H1 Title</h1>`, "OG Title"},
		{`<h1>H1 Title</h1><title>Doc Title</title>`, "H1 Title"},
		{`<title>Doc Title</title>`, "Doc Title"},
		{`<meta name='title' content='Meta Title'>`, "Meta Title"},
	}
	for i, c := range cases {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(c.html))
		if err != nil {
			t.Fatalf("case %d: %v", i, err)
		}
		got, err := extractTitle(doc)
		if err != nil {
			t.Fatalf("case %d: extractTitle error: %v", i, err)
		}
		if got != c.want {
			t.Errorf("case %d: extractTitle()=%q want %q", i, got, c.want)
		}
	}

	// error when nothing found
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(`<div></div>`))
	if _, err := extractTitle(doc); err == nil {
		t.Error("expected error when title not found")
	}
}
