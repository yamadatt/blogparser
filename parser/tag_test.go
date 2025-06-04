package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestCleanTag(t *testing.T) {
	cases := []struct{ in, want string }{
		{"  #Go \nブログ", "Go"},
		{"心理カウンセラー・中井亜紀『成長の記録』タグ", "タグ"},
		{"multi   space", "multi space"},
	}
	for _, c := range cases {
		got := cleanTag(c.in)
		if got != c.want {
			t.Errorf("cleanTag(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestExtractTags(t *testing.T) {
	html := `
<meta name='keywords' content='kw1, kw2'>
<div class='skin-tagLabel'>TagA</div>
<script>var ld_blog_vars={tags:['TagB','TagC']};</script>
<div class='tags'><a>TagD</a></div>
<div class='tag'>TagE</div>
`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("doc error: %v", err)
	}
	tags, err := extractTags(doc)
	if err != nil {
		t.Fatalf("extractTags error: %v", err)
	}
	expected := []string{"TagA", "TagB", "TagC", "kw1", "kw2", "TagD", "TagE"}
	if len(tags) != len(expected) {
		t.Fatalf("got %d tags want %d", len(tags), len(expected))
	}
	for _, e := range expected {
		if !containsString(tags, e) {
			t.Errorf("expected tag %q", e)
		}
	}
}

func TestExtractTagsNil(t *testing.T) {
	if _, err := extractTags(nil); err == nil {
		t.Error("expected error with nil document")
	}
}
