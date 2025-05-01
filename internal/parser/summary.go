package parser

import (
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// 品詞の重み付け
var posWeights = map[string]float64{
	"名詞-固有名詞": 2.0,
	"名詞-一般":   1.5,
	"動詞-自立":   1.2,
	"形容詞-自立":  1.2,
	"副詞-一般":   0.8,
	"名詞-副詞可能": 0.7,
}

// Word は単語とその重みを表します
type Word struct {
	Surface string  // 表層形
	Lemma   string  // 基本形
	POS     string  // 品詞
	Weight  float64 // 重み
}

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
	vectors := make([][]Word, len(sentences))
	t, err := tokenizer.New(ipa.Dict())
	if err != nil {
		return text
	}

	for i, sentence := range sentences {
		tokens := t.Tokenize(sentence)
		var words []Word
		for _, token := range tokens {
			features := token.Features()
			if len(features) < 7 {
				continue
			}

			// 品詞情報の取得
			pos := features[0]
			if len(features) > 1 {
				pos += "-" + features[1]
			}

			// 重要な品詞のみを対象とする
			weight := getWordWeight(pos)
			if weight > 0 {
				word := Word{
					Surface: token.Surface,
					Lemma:   features[6], // 基本形
					POS:     pos,
					Weight:  weight,
				}
				words = append(words, word)
			}
		}
		vectors[i] = words
	}

	// 5. TextRankの計算
	scores := calculateTextRank(vectors)

	// 6. 上位2文を選択（文の長さで重み付け）
	type SentenceScore struct {
		index int
		score float64
	}
	var ranked []SentenceScore
	for i, score := range scores {
		// 極端に短い文や長い文にペナルティを与える
		length := len(sentences[i])
		if length < 10 {
			score *= 0.5
		} else if length > 200 {
			score *= 0.8
		}
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

// getWordWeight は品詞に基づいて単語の重要度を返します
func getWordWeight(pos string) float64 {
	if weight, exists := posWeights[pos]; exists {
		return weight
	}
	// デフォルトの重み
	switch {
	case strings.HasPrefix(pos, "名詞"):
		return 1.0
	case strings.HasPrefix(pos, "動詞"):
		return 0.9
	case strings.HasPrefix(pos, "形容詞"):
		return 0.9
	}
	return 0
}

// calculateTextRank は文のベクトル表現からTextRankスコアを計算します
func calculateTextRank(vectors [][]Word) []float64 {
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

// calculateSimilarity は2つの文ベクトル間の重み付きコサイン類似度を計算します
func calculateSimilarity(vec1, vec2 []Word) float64 {
	// 単語の重み付き出現をカウント
	count1 := make(map[string]float64)
	count2 := make(map[string]float64)

	for _, word := range vec1 {
		// 基本形をキーとして使用
		count1[word.Lemma] += word.Weight
	}
	for _, word := range vec2 {
		count2[word.Lemma] += word.Weight
	}

	// 内積の計算
	dotProduct := 0.0
	for word, count := range count1 {
		if count2[word] > 0 {
			dotProduct += count * count2[word]
		}
	}

	// ベクトルの大きさを計算
	magnitude1 := 0.0
	for _, count := range count1 {
		magnitude1 += count * count
	}
	magnitude2 := 0.0
	for _, count := range count2 {
		magnitude2 += count * count
	}

	// コサイン類似度を計算
	if magnitude1 == 0 || magnitude2 == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(magnitude1) * math.Sqrt(magnitude2))
}

// splitSentences は文を分割します
func splitSentences(text string) []string {
	// 1. 括弧内の句点を一時的に置換
	text = replaceBracketContent(text)

	// 2. 文分割のパターン
	patterns := []string{
		`[。．.！!？?]`,   // 基本的な句読点
		`。」`,          // 会話文の終わり
		`。）`,          // 括弧付きの文の終わり
		`[;\n]`,       // セミコロンと改行
		`。[\s]*[\n]+`, // 句点+改行
		`^[\s]*[•●・]`, // 箇条書きの開始
	}

	// 3. パターンを組み合わせて分割
	pattern := "(" + strings.Join(patterns, "|") + ")"
	re := regexp.MustCompile(pattern)
	parts := re.Split(text, -1)

	// 4. 空の文を除去し、意味のある文のみを保持
	var sentences []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) > 0 {
			// 括弧内の句点を復元
			part = restoreBracketContent(part)
			sentences = append(sentences, part)
		}
	}

	return sentences
}

// replaceBracketContent は括弧内のテキストを一時的に置換します
func replaceBracketContent(text string) string {
	re := regexp.MustCompile(`[（(][^）)]*[）)]`)
	return re.ReplaceAllStringFunc(text, func(s string) string {
		return strings.ReplaceAll(s, "。", "@@PERIOD@@")
	})
}

// restoreBracketContent は括弧内のテキストを復元します
func restoreBracketContent(text string) string {
	return strings.ReplaceAll(text, "@@PERIOD@@", "。")
}
