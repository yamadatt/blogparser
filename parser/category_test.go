package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestCleanCategory(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"テーマ：テスト", "テスト"},
		{"テーマ:テスト", "テスト"},
		{"Theme：Test", "Test"},
		{"Theme:Test", "Test"},
		{"  テーマ：  テスト  ", "テスト"},
		{"テスト\nカテゴリ", "テスト カテゴリ"},
		{"  テスト  ", "テスト"},
	}
	for _, c := range cases {
		got := cleanCategory(c.in)
		if got != c.want {
			t.Errorf("cleanCategory(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}

func TestIsValidCategory(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"カテゴリ", true},
		{"", false},
		{"<b>カテゴリ</b>", false},
		{"カテゴリ\x01", false},
		{"カテゴリ\n", true},
	}
	for _, c := range cases {
		got := isValidCategory(c.in)
		if got != c.want {
			t.Errorf("isValidCategory(%q) = %v; want %v", c.in, got, c.want)
		}
	}
}

func TestContainsString(t *testing.T) {
	slice := []string{"a", "b", "c"}
	if !containsString(slice, "a") {
		t.Error("containsString should return true for existing element")
	}
	if containsString(slice, "d") {
		t.Error("containsString should return false for non-existing element")
	}
}

func TestExtractCategories_Selectors(t *testing.T) {
	html := `<div class="skin-categoryLabel">カテゴリ1</div><div class="skin-categoryLabel">カテゴリ2</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("goquery.NewDocumentFromReader error: %v", err)
	}
	cats, err := extractCategories(doc)
	if err != nil {
		t.Fatalf("extractCategories error: %v", err)
	}
	if len(cats) != 2 || cats[0] != "カテゴリ1" || cats[1] != "カテゴリ2" {
		t.Errorf("extractCategories selectors failed: got %v", cats)
	}
}

func TestExtractCategories_LdBlogVars(t *testing.T) {
	html := `<script>var ld_blog_vars = {articles:[{categories:[{name:'カテゴリA'}]}]};</script>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("goquery.NewDocumentFromReader error: %v", err)
	}
	cats, err := extractCategories(doc)
	if err != nil {
		t.Fatalf("extractCategories error: %v", err)
	}
	if len(cats) != 1 || cats[0] != "カテゴリA" {
		t.Errorf("extractCategories ld_blog_vars failed: got %v", cats)
	}
}

func TestExtractCategories_MetaSection(t *testing.T) {
	html := `<meta property='article:section' content='カテゴリB'>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("goquery.NewDocumentFromReader error: %v", err)
	}
	cats, err := extractCategories(doc)
	if err != nil {
		t.Fatalf("extractCategories error: %v", err)
	}
	if len(cats) != 1 || cats[0] != "カテゴリB" {
		t.Errorf("extractCategories meta section failed: got %v", cats)
	}
}

func TestExtractCategories_CategoryClass(t *testing.T) {
	html := `<div class='category'>カテゴリC</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("goquery.NewDocumentFromReader error: %v", err)
	}
	cats, err := extractCategories(doc)
	if err != nil {
		t.Fatalf("extractCategories error: %v", err)
	}
	if len(cats) != 1 || cats[0] != "カテゴリC" {
		t.Errorf("extractCategories .category class failed: got %v", cats)
	}
}

func TestExtractCategories_NilDoc(t *testing.T) {
	_, err := extractCategories(nil)
	if err == nil {
		t.Error("extractCategories(nil) should return error")
	}
}
