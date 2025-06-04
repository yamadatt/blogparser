package models

import "testing"

func TestSetSlug(t *testing.T) {
	cases := []struct{ title, want string }{
		{"Hello World!", "hello-world"},
		{"日本語タイトル", ""},
		{"Mixed 123 Title", "mixed-123-title"},
		{"  spaces  here  ", "spaces-here"},
	}
	for _, c := range cases {
		b := &BlogPost{Title: c.title}
		b.SetSlug()
		if b.Slug != c.want {
			t.Errorf("SetSlug(%q)=%q want %q", c.title, b.Slug, c.want)
		}
	}
}

func TestRemoveNonASCII(t *testing.T) {
	got := removeNonASCII("aあb1")
	if got != "ab1" {
		t.Errorf("removeNonASCII unexpected %q", got)
	}
}
