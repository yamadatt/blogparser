package parser

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ImageInfo は画像の情報を保持する構造体です
type ImageInfo struct {
	URL         string
	Alt         string
	Width       string
	Height      string
	Description string
}

// ExtractImages はHTML内のすべての画像情報を抽出します
func (p *HTMLParser) ExtractImages(content string) []ImageInfo {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return nil
	}

	var images []ImageInfo

	// 1. まずOGP画像を探す
	ogImage, _ := doc.Find("meta[property='og:image']").Attr("content")
	if ogImage != "" {
		// OGP画像の情報を収集
		ogImageInfo := ImageInfo{
			URL:         normalizeImageURL(ogImage),
			Alt:         "OGP Image",
			Description: doc.Find("meta[property='og:description']").AttrOr("content", ""),
		}
		if ogImageInfo.URL != "" {
			images = append(images, ogImageInfo)
		}
	}

	// 2. 次にTwitter Card画像を探す（OGP画像がない場合）
	if len(images) == 0 {
		twitterImage, _ := doc.Find("meta[name='twitter:image']").Attr("content")
		if twitterImage != "" {
			twitterImageInfo := ImageInfo{
				URL:         normalizeImageURL(twitterImage),
				Alt:         "Twitter Card Image",
				Description: doc.Find("meta[name='twitter:description']").AttrOr("content", ""),
			}
			if twitterImageInfo.URL != "" {
				images = append(images, twitterImageInfo)
			}
		}
	}

	// 3. 通常の画像を探す
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		// data-src属性（遅延読み込み用）またはsrc属性を取得
		imgURL, _ := s.Attr("data-src")
		if imgURL == "" {
			imgURL, _ = s.Attr("src")
		}

		// 画像URLの正規化
		imgURL = normalizeImageURL(imgURL)
		if imgURL == "" {
			return
		}

		// 画像情報の収集
		alt, _ := s.Attr("alt")
		width, _ := s.Attr("width")
		height, _ := s.Attr("height")

		// 画像の説明文を取得（親要素のfigcaptionなど）
		description := ""
		if parent := s.Parent(); parent.Is("figure") {
			description = parent.Find("figcaption").Text()
		}

		images = append(images, ImageInfo{
			URL:         imgURL,
			Alt:         alt,
			Width:       width,
			Height:      height,
			Description: strings.TrimSpace(description),
		})
	})

	return images
}

// GetFirstImage は最初の有効な画像のURLを返します
func (p *HTMLParser) GetFirstImage(content string) string {
	images := p.ExtractImages(content)
	if len(images) > 0 {
		return images[0].URL
	}
	return ""
}

// normalizeImageURL は画像URLを正規化します
func normalizeImageURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	// データURLは除外
	if strings.HasPrefix(rawURL, "data:") {
		return ""
	}

	// URLの正規化
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// アメブロの画像URLを最適化
	if strings.Contains(u.Host, "ameblo.jp") {
		// 小さいサイズの画像URLを元のサイズに変換
		rawURL = strings.Replace(rawURL, "_s.", ".", 1)
		rawURL = strings.Replace(rawURL, "_m.", ".", 1)
	}

	return rawURL
}
