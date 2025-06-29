package parser

import (
	"math"
	"strings"
	"testing"
)

func TestCalculateTF(t *testing.T) {
	doc := []Word{{Lemma: "go"}, {Lemma: "lang"}, {Lemma: "go"}}
	if tf := calculateTF(doc, "go"); tf != 2 {
		t.Errorf("calculateTF go=%.1f want 2", tf)
	}
	if tf := calculateTF(doc, "lang"); tf != 1 {
		t.Errorf("calculateTF lang=%.1f want 1", tf)
	}
	if tf := calculateTF(doc, "py"); tf != 0 {
		t.Errorf("calculateTF py=%.1f want 0", tf)
	}
}

func TestContainsWord(t *testing.T) {
	doc := []Word{{Lemma: "go"}, {Lemma: "lang"}}
	if !containsWord(doc, "go") {
		t.Error("containsWord should find existing word")
	}
	if containsWord(doc, "py") {
		t.Error("containsWord should not find missing word")
	}
}

func TestTruncateSummary(t *testing.T) {
	short := strings.Repeat("a", 10)
	if got := truncateSummary(short); got != short {
		t.Errorf("truncateSummary short=%q", got)
	}
	long := strings.Repeat("b", 305)
	got := truncateSummary(long)
	if !strings.HasSuffix(got, "・・・") || len([]rune(got)) != 303 {
		t.Errorf("truncateSummary long unexpected: %d %q", len([]rune(got)), got)
	}
}

func TestSplitSentences(t *testing.T) {
	p := &HTMLParser{}
	s := p.splitSentences("今日は晴れです。 明日も晴れ。")
	if len(s) != 2 || s[0] != "今日は晴れです" || s[1] != "明日も晴れ" {
		t.Errorf("splitSentences unexpected: %v", s)
	}
}

func TestIsSentenceEnd(t *testing.T) {
	if !isSentenceEnd("。", nil) || !isSentenceEnd("?", nil) {
		t.Error("isSentenceEnd punctuation failed")
	}
	if !isSentenceEnd("", []string{"記号", "句点"}) {
		t.Error("isSentenceEnd features failed")
	}
	if isSentenceEnd("a", []string{"名詞"}) {
		t.Error("isSentenceEnd non end failed")
	}
}

func TestGetWordWeight(t *testing.T) {
	if w := getWordWeight("名詞-固有名詞"); w != 2.0 {
		t.Errorf("getWordWeight 固有名詞=%.1f", w)
	}
	if w := getWordWeight("動詞-接尾"); w != 0.9 {
		t.Errorf("getWordWeight default verb=%.1f", w)
	}
	if w := getWordWeight("記号-一般"); w != 0 {
		t.Errorf("getWordWeight symbol=%.1f", w)
	}
}

func TestCalculateBM25Score(t *testing.T) {
	doc1 := []Word{{Lemma: "go", Weight: 1}}
	doc2 := []Word{{Lemma: "python", Weight: 1}}
	doc3 := []Word{{Lemma: "java", Weight: 1}}
	docs := [][]Word{doc1, doc2, doc3}
	score := calculateBM25Score(doc1, docs, 1)
	expected := math.Log((3-1+0.5)/(1+0.5)) * (1 * (k1 + 1)) / (1 + k1*(1-b+b*1/1))
	if math.Abs(score-expected) > 1e-6 {
		t.Errorf("BM25Score got %f want %f", score, expected)
	}
	docsAllSame := [][]Word{doc1, doc1, doc1}
	score = calculateBM25Score(doc1, docsAllSame, 1)
	if score != 0 {
		t.Errorf("BM25Score expected 0 when idf negative, got %f", score)
	}
}

func TestGenerateSummary(t *testing.T) {
	p := &HTMLParser{}
	html := `<html><body>今日は天気です。明日は雨です。明後日は晴れです。</body></html>`
	sum, err := p.GenerateSummary(html)
	if err != nil {
		t.Fatalf("GenerateSummary error: %v", err)
	}
	if sum != "今日は天気です明日は雨です" {
		t.Errorf("GenerateSummary unexpected: %q", sum)
	}
	if _, err := p.GenerateSummary(""); err == nil {
		t.Error("GenerateSummary empty content should error")
	}
}

func TestGenerateSummaryEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		minLen   int
	}{
		{
			name:    "短いコンテンツ",
			content: `<html><body>短い。</body></html>`,
			wantErr: false,
			minLen:  1,
		},
		{
			name:    "長いコンテンツ",
			content: `<html><body>` + strings.Repeat("これは長い文章です。", 100) + `</body></html>`,
			wantErr: false,
			minLen:  1,
		},
		{
			name:    "HTMLタグのみ",
			content: `<html><body><div><span></span></div></body></html>`,
			wantErr: false,
			minLen:  0,
		},
		{
			name:    "特殊文字を含む",
			content: `<html><body>これは&lt;特殊&gt;文字です。&amp;記号も含みます。</body></html>`,
			wantErr: false,
			minLen:  1,
		},
		{
			name:    "数字のみ",
			content: `<html><body>123456789。</body></html>`,
			wantErr: false,
			minLen:  1,
		},
		{
			name:    "英語コンテンツ",
			content: `<html><body>This is English content. It should work too.</body></html>`,
			wantErr: false,
			minLen:  1,
		},
		{
			name:    "混合言語",
			content: `<html><body>これはJapanese and English混合です。Very interesting content.</body></html>`,
			wantErr: false,
			minLen:  1,
		},
	}

	p := &HTMLParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.GenerateSummary(tt.content)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateSummary() error = nil, want error")
				}
				return
			}
			
			if err != nil {
				t.Errorf("GenerateSummary() error = %v, want nil", err)
				return
			}
			
			if len(result) < tt.minLen {
				t.Errorf("GenerateSummary() result length = %d, want >= %d", len(result), tt.minLen)
			}
		})
	}
}

func TestTokenizeEdgeCases(t *testing.T) {
	p := &HTMLParser{}
	
	tests := []struct {
		name     string
		text     string
		expected int // 期待される単語数の最小値
	}{
		{
			name:     "空文字列",
			text:     "",
			expected: 0,
		},
		{
			name:     "空白のみ",
			text:     "   \n\t  ",
			expected: 0,
		},
		{
			name:     "日本語文章",
			text:     "今日は良い天気です。",
			expected: 1,
		},
		{
			name:     "英語文章",
			text:     "This is a test.",
			expected: 1,
		},
		{
			name:     "数字のみ",
			text:     "123456",
			expected: 0,
		},
		{
			name:     "記号のみ",
			text:     "！？。、",
			expected: 0,
		},
		{
			name:     "非常に長いテキスト",
			text:     strings.Repeat("これは長いテキストです。", 1000),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words, err := p.tokenize(tt.text)
			if err != nil {
				t.Errorf("tokenize() error = %v, want nil", err)
				return
			}
			if len(words) < tt.expected {
				t.Errorf("tokenize() returned %d words, want >= %d", len(words), tt.expected)
			}
		})
	}
}

func TestProcessVectors(t *testing.T) {
	p := &HTMLParser{}
	
	// 空のベクターリスト
	vectors := make([][]Word, 0)
	sentences := []string{}
	err := p.processVectors(vectors, sentences)
	if err != nil {
		t.Errorf("processVectors() with empty input should not error, got: %v", err)
	}
	
	// 単一のベクター
	singleVectors := make([][]Word, 1)
	singleSentences := []string{"テスト文章"}
	err = p.processVectors(singleVectors, singleSentences)
	if err != nil {
		t.Errorf("processVectors() with single vector should not error, got: %v", err)
	}
	
	// 複数のベクター
	multipleVectors := make([][]Word, 3)
	multipleSentences := []string{"テスト文章1", "テスト文章2", "テスト文章3"}
	err = p.processVectors(multipleVectors, multipleSentences)
	if err != nil {
		t.Errorf("processVectors() with multiple vectors should not error, got: %v", err)
	}
}
