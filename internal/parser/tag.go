package parser

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

// extractTags はHTMLドキュメントからタグを抽出します。
// 以下の優先順位で抽出を試みます：
// 1. よく使われるタグ用セレクタ
// 2. ld_blog_varsのarticles[0].tags
// 3. meta[name="keywords"]
// 4. .tag, .tags, .entry-tags, .post-tags など
func extractTags(doc *goquery.Document) ([]string, error) {
	if doc == nil {
		return nil, errors.New("ドキュメントがnilです")
	}

	var tags []string

	// 1. よく使われるタグ用セレクタ
	selectors := []string{
		".skin-tagLabel",    // アメブロのタグラベル
		".skin-entryTags a", // アメブロのタグリンク
		".skin-tag",         // アメブロのタグ
		".tag a",            // 一般的なタグリンク
		".tags a",
		".entry-tags a",
		".post-tags a",
		".blog-tags a",
		".article-tags a",
		".taglist a",
		".entryTag a",
		".entry_tag a",
		".blogTag a",
		".blog_tag a",
		".label a",
		".labels a",
		".post-labels a",
		".post_label a",
		".entry-labels a",
		".entry_label a",
		".tagcloud a",
		".tagCloud a",
		".tag-list a",
		".tagList a",
		".tag_links a",
		".tagLinks a",
		".tag a[rel='tag']",
		".hashtag-module__item__text", // ハッシュタグspan対応
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			tag := strings.TrimSpace(s.Text())
			if tag != "" && !containsString(tags, tag) {
				tags = append(tags, tag)
			}
		})
	}

	// セレクタからタグが見つかった場合は返す
	if len(tags) > 0 {
		return tags, nil
	}

	// 2. ld_blog_varsからtagsを抽出
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		if script := s.Text(); strings.Contains(script, "ld_blog_vars") {
			// タグを抽出するための正規表現
			re := regexp.MustCompile(`tags\s*:\s*\[([^\]]*)\]`)
			if matches := re.FindStringSubmatch(script); len(matches) > 1 {
				tagsStr := matches[1]
				tagRe := regexp.MustCompile(`'([^']*)'`)
				tagMatches := tagRe.FindAllStringSubmatch(tagsStr, -1)
				for _, tm := range tagMatches {
					if len(tm) > 1 {
						tag := strings.TrimSpace(tm[1])
						if tag != "" && !containsString(tags, tag) {
							tags = append(tags, tag)
						}
					}
				}
			}
		}
	})

	if len(tags) > 0 {
		return tags, nil
	}

	// 3. meta[name="keywords"]から抽出
	if keywords, exists := doc.Find("meta[name='keywords']").Attr("content"); exists {
		for _, tag := range strings.Split(keywords, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" && !containsString(tags, tag) {
				tags = append(tags, tag)
			}
		}
	}

	if len(tags) > 0 {
		return tags, nil
	}

	// 4. .tag, .tags, .entry-tags, .post-tags などのテキストから抽出
	textSelectors := []string{
		".tag", ".tags", ".entry-tags", ".post-tags",
	}
	for _, selector := range textSelectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			tag := strings.TrimSpace(s.Text())
			if tag != "" && !containsString(tags, tag) {
				tags = append(tags, tag)
			}
		})
	}

	if len(tags) == 0 {
		return nil, errors.New("タグが見つかりません")
	}

	return tags, nil
}
