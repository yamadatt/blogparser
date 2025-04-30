package parser

import (
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// GenerateSummary は記事本文からサマリ（要約）を生成します。
// TextRankアルゴリズムを使用して重要な文を抽出します。
func (p *HTMLParser) GenerateSummary(content string) string {
	// 1. HTML → テキスト変換
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return ""
	}
	text := doc.Find("body").Text()

	// 2. テキストの正規化
	text = p.normalizeWhitespace(text)

	// 3. 文分割
	sentences := splitSentences(text)
	if len(sentences) <= 2 {
		return text // 文が少ない場合は全文を返す
	}

	// 4. 形態素解析と文のベクトル化
	vectors := make([][]string, len(sentences))
	t, err := tokenizer.New(ipa.Dict())
	if err != nil {
		return text
	}

	for i, sentence := range sentences {
		tokens := t.Tokenize(sentence)
		words := make([]string, 0)
		for _, token := range tokens {
			features := token.Features()
			if len(features) > 0 && (features[0] == "名詞" || features[0] == "動詞" || features[0] == "形容詞") {
				words = append(words, token.Surface)
			}
		}
		vectors[i] = words
	}

	// 5. TextRankの計算
	scores := calculateTextRank(vectors)

	// 6. 上位2文を選択
	type SentenceScore struct {
		index int
		score float64
	}
	var ranked []SentenceScore
	for i, score := range scores {
		ranked = append(ranked, SentenceScore{i, score})
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})

	// 7. 文の順序を維持して結合
	var summary []string
	for i := 0; i < len(sentences) && len(summary) < 2; i++ {
		for _, r := range ranked {
			if r.index == i {
				summary = append(summary, sentences[i])
				break
			}
		}
	}

	return strings.Join(summary, "")
}

// calculateTextRank は文のベクトル表現からTextRankスコアを計算します
func calculateTextRank(vectors [][]string) []float64 {
	n := len(vectors)
	if n == 0 {
		return nil
	}

	// 類似度行列の作成
	similarity := make([][]float64, n)
	for i := range similarity {
		similarity[i] = make([]float64, n)
	}

	// コサイン類似度の計算
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i != j {
				similarity[i][j] = calculateSimilarity(vectors[i], vectors[j])
			}
		}
	}

	// TextRankの反復計算
	scores := make([]float64, n)
	for i := range scores {
		scores[i] = 1.0
	}

	d := 0.85 // ダンピング係数
	iterations := 30

	for iter := 0; iter < iterations; iter++ {
		newScores := make([]float64, n)
		for i := 0; i < n; i++ {
			sum := 0.0
			for j := 0; j < n; j++ {
				if i != j {
					// 正規化された重みを計算
					weightSum := 0.0
					for k := 0; k < n; k++ {
						if k != j {
							weightSum += similarity[j][k]
						}
					}
					if weightSum > 0 {
						sum += similarity[j][i] * scores[j] / weightSum
					}
				}
			}
			newScores[i] = (1 - d) + d*sum
		}
		scores = newScores
	}

	return scores
}

// calculateSimilarity は2つの文ベクトル間のコサイン類似度を計算します
func calculateSimilarity(vec1, vec2 []string) float64 {
	// 単語の出現をカウント
	count1 := make(map[string]int)
	count2 := make(map[string]int)

	for _, word := range vec1 {
		count1[word]++
	}
	for _, word := range vec2 {
		count2[word]++
	}

	// 内積の計算
	dotProduct := 0.0
	for word, count := range count1 {
		if count2[word] > 0 {
			dotProduct += float64(count * count2[word])
		}
	}

	// ベクトルの大きさを計算
	magnitude1 := 0.0
	for _, count := range count1 {
		magnitude1 += float64(count * count)
	}
	magnitude2 := 0.0
	for _, count := range count2 {
		magnitude2 += float64(count * count)
	}

	// コサイン類似度を計算
	if magnitude1 == 0 || magnitude2 == 0 {
		return 0
	}
	return dotProduct / (sqrt(magnitude1) * sqrt(magnitude2))
}

// sqrt は平方根を計算します
func sqrt(x float64) float64 {
	// 簡易的なニュートン法による平方根の計算
	z := 1.0
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

// stripHTMLTags はHTMLタグを除去します
func stripHTMLTags(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(html, "")
}

// splitSentences は句点（。や.）で文を分割します
func splitSentences(text string) []string {
	re := regexp.MustCompile(`.*?[。.]`)
	return re.FindAllString(text, -1)
}
