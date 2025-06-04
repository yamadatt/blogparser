package parser

import "testing"

func TestNormalizeHTML(t *testing.T) {
	html := "<p>line1</p>\r\n<p>line2</p>\n\n<p>line3</p>"
	got := normalizeHTML(html)
	want := "<p>line1</p>\n<p>line2</p>\n<p>line3</p>"
	if got != want {
		t.Errorf("normalizeHTML got %q want %q", got, want)
	}
}

func TestIsValidContent(t *testing.T) {
	long := make([]byte, 100)
	for i := range long {
		long[i] = 'a'
	}
	if !isValidContent(string(long)) {
		t.Error("expected valid content")
	}
	if isValidContent("") {
		t.Error("empty should be invalid")
	}
	if isValidContent("short") {
		t.Error("short text should be invalid")
	}
}
