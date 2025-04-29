package parser

import (
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

// extractDate はHTMLドキュメントから公開日時を抽出します。
// 以下の優先順位で抽出を試みます：
// 1. <time datetime="...">
// 2. <meta property="article:published_time" content="...">
// 3. <meta name="pubdate" content="...">
// 4. <meta name="date" content="...">
// 5. <span class="date">...</span> など
// 6. script[type="application/ld+json"]内の"datePublished"
func extractDate(doc *goquery.Document) (time.Time, error) {
	if doc == nil {
		return time.Time{}, errors.New("ドキュメントがnilです")
	}

	// 1. script[type="application/ld+json"]内の"datePublished"
	var datePublished string
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		jsonText := s.Text()
		if dateStr := extractDatePublishedFromJSONLD(jsonText); dateStr != "" {
			datePublished = dateStr
		}
	})
	if datePublished != "" {
		parsed, err := parseDateString(datePublished)
		if err == nil {
			return parsed, nil
		}
	}

	// 2. <time datetime="...">
	if t := doc.Find("time[datetime]").First(); t.Length() > 0 {
		if dt, exists := t.Attr("datetime"); exists {
			parsed, err := parseDateString(dt)
			if err == nil {
				return parsed, nil
			}
		}
		// timeタグのテキストも試す
		text := strings.TrimSpace(t.Text())
		parsed, err := parseDateString(text)
		if err == nil {
			return parsed, nil
		}
	}

	// 3. <meta property="article:published_time" content="...">
	if content, exists := doc.Find("meta[property='article:published_time']").Attr("content"); exists {
		parsed, err := parseDateString(content)
		if err == nil {
			return parsed, nil
		}
	}

	// 4. <meta name="pubdate" content="...">
	if content, exists := doc.Find("meta[name='pubdate']").Attr("content"); exists {
		parsed, err := parseDateString(content)
		if err == nil {
			return parsed, nil
		}
	}

	// 5. <meta name="date" content="...">
	if content, exists := doc.Find("meta[name='date']").Attr("content"); exists {
		parsed, err := parseDateString(content)
		if err == nil {
			return parsed, nil
		}
	}

	// 6. <span class="date">...</span> など
	if s := doc.Find(".date").First(); s.Length() > 0 {
		text := strings.TrimSpace(s.Text())
		parsed, err := parseDateString(text)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, errors.New("公開日時が見つかりません")
}

// extractDatePublishedFromJSONLDはJSON-LDテキストから"datePublished"値を抽出する
func extractDatePublishedFromJSONLD(jsonText string) string {
	idx := strings.Index(jsonText, "\"datePublished\"")
	if idx == -1 {
		return ""
	}
	remain := jsonText[idx+len("\"datePublished\""):]
	// コロンとスペースをスキップ
	remain = strings.TrimLeft(remain, ": ")
	if len(remain) == 0 || remain[0] != '"' {
		return ""
	}
	remain = remain[1:]
	endIdx := strings.Index(remain, "\"")
	if endIdx == -1 {
		return ""
	}
	return remain[:endIdx]
}

// parseDateString は様々な日付文字列をtime.Timeに変換します。
func parseDateString(s string) (time.Time, error) {
	// よく使われる日付フォーマットを試す
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02",
		"2006年1月2日",
		"2006年01月02日",
		"2006.01.02",
		"2006/1/2",
		"2006-1-2",
		"2006年1月2日 15:04",
		"2006/01/02 15:04",
		"2006-01-02 15:04",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.Errorf("日付のパースに失敗: %s", s)
}
