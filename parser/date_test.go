package parser

import (
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractDate(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected time.Time
		wantErr  bool
	}{
		{
			name: "JSON-LDからの日付抽出",
			html: `<html><head>
				<script type="application/ld+json">
				{
					"@type": "BlogPosting",
					"datePublished": "2023-12-01T10:30:00+09:00"
				}
				</script>
			</head></html>`,
			expected: time.Date(2023, 12, 1, 10, 30, 0, 0, time.FixedZone("JST", 9*3600)),
			wantErr:  false,
		},
		{
			name: "timeタグのdatetime属性からの抽出",
			html: `<html><body>
				<time datetime="2023-11-15T14:20:00Z">2023年11月15日</time>
			</body></html>`,
			expected: time.Date(2023, 11, 15, 14, 20, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name: "timeタグのテキストからの抽出",
			html: `<html><body>
				<time>2023-10-20</time>
			</body></html>`,
			expected: time.Time{},
			wantErr:  true,
		},
		{
			name: "meta property article:published_timeからの抽出",
			html: `<html><head>
				<meta property="article:published_time" content="2023-09-10T08:15:00+09:00">
			</head></html>`,
			expected: time.Date(2023, 9, 10, 8, 15, 0, 0, time.FixedZone("JST", 9*3600)),
			wantErr:  false,
		},
		{
			name: "meta name pubdateからの抽出",
			html: `<html><head>
				<meta name="pubdate" content="2023-08-05">
			</head></html>`,
			expected: time.Date(2023, 8, 5, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name: "meta name dateからの抽出",
			html: `<html><head>
				<meta name="date" content="2023/07/25">
			</head></html>`,
			expected: time.Date(2023, 7, 25, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name: "class dateからの抽出",
			html: `<html><body>
				<span class="date">2023年6月12日</span>
			</body></html>`,
			expected: time.Date(2023, 6, 12, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name: "日付が見つからない場合",
			html: `<html><body>
				<p>日付情報なし</p>
			</body></html>`,
			expected: time.Time{},
			wantErr:  true,
		},
		{
			name:     "nilドキュメント",
			html:     "",
			expected: time.Time{},
			wantErr:  true,
		},
		{
			name: "無効な日付フォーマット",
			html: `<html><body>
				<time datetime="invalid-date">無効な日付</time>
			</body></html>`,
			expected: time.Time{},
			wantErr:  true,
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

			result, err := extractDate(doc)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("extractDate() error = nil, want error")
				}
				return
			}
			
			if err != nil {
				t.Errorf("extractDate() error = %v, want nil", err)
				return
			}
			
			if !result.Equal(tt.expected) {
				t.Errorf("extractDate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractDatePublishedFromJSONLD(t *testing.T) {
	tests := []struct {
		name     string
		jsonText string
		expected string
	}{
		{
			name:     "正常なJSON-LD",
			jsonText: `{"@type": "BlogPosting", "datePublished": "2023-12-01T10:30:00+09:00"}`,
			expected: "2023-12-01T10:30:00+09:00",
		},
		{
			name:     "スペース付きJSON-LD",
			jsonText: `{"@type": "BlogPosting", "datePublished" : "2023-11-15T14:20:00Z"}`,
			expected: "2023-11-15T14:20:00Z",
		},
		{
			name:     "datePublishedが存在しない",
			jsonText: `{"@type": "BlogPosting", "title": "テスト記事"}`,
			expected: "",
		},
		{
			name:     "空のJSON",
			jsonText: `{}`,
			expected: "",
		},
		{
			name:     "無効なJSON構造",
			jsonText: `{"datePublished": 123}`,
			expected: "",
		},
		{
			name:     "閉じクォートなし",
			jsonText: `{"datePublished": "2023-12-01T10:30:00+09:00}`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDatePublishedFromJSONLD(tt.jsonText)
			if result != tt.expected {
				t.Errorf("extractDatePublishedFromJSONLD() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseDateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "RFC3339フォーマット",
			input:    "2023-12-01T10:30:00+09:00",
			expected: time.Date(2023, 12, 1, 10, 30, 0, 0, time.FixedZone("JST", 9*3600)),
			wantErr:  false,
		},
		{
			name:     "ISO8601フォーマット",
			input:    "2023-11-15T14:20:00Z",
			expected: time.Date(2023, 11, 15, 14, 20, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "日付のみ",
			input:    "2023-10-20",
			expected: time.Date(2023, 10, 20, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "スラッシュ区切り",
			input:    "2023/09/15",
			expected: time.Date(2023, 9, 15, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "日本語フォーマット",
			input:    "2023年8月10日",
			expected: time.Date(2023, 8, 10, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "ドット区切り",
			input:    "2023.07.25",
			expected: time.Date(2023, 7, 25, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "時刻付き日本語",
			input:    "2023年6月12日 15:30",
			expected: time.Date(2023, 6, 12, 15, 30, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "無効なフォーマット",
			input:    "invalid-date",
			expected: time.Time{},
			wantErr:  true,
		},
		{
			name:     "空文字列",
			input:    "",
			expected: time.Time{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDateString(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDateString() error = nil, want error")
				}
				return
			}
			
			if err != nil {
				t.Errorf("parseDateString() error = %v, want nil", err)
				return
			}
			
			if !result.Equal(tt.expected) {
				t.Errorf("parseDateString() = %v, want %v", result, tt.expected)
			}
		})
	}
}