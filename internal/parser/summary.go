package parser

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"go.uber.org/zap"
)

// BM25のパラメータ
const (
	k1 = 1.2  // 単語頻度の飽和度パラメータ
	b  = 0.75 // 文書長正規化パラメータ
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
	TF      float64 // その文での出現頻度
	IDF     float64 // 逆文書頻度
}

// BM25Score は文のBM25スコアを計算します
func calculateBM25Score(doc []Word, docs [][]Word, avgDocLen float64) float64 {
	score := 0.0
	docLen := float64(len(doc))

	for _, word := range doc {
		// 単語の文書頻度を計算
		df := 0
		for _, d := range docs {
			if containsWord(d, word.Lemma) {
				df++
			}
		}

		// IDFの計算
		docsLen := float64(len(docs))
		dfFloat := float64(df)
		idf := math.Log((docsLen - dfFloat + 0.5) / (dfFloat + 0.5))
		if idf < 0 {
			idf = 0 // IDFが負になるのを防ぐ
		}

		// TFの計算（文書内での単語頻度）
		tf := calculateTF(doc, word.Lemma)

		// BM25スコアの計算
		numerator := tf * (k1 + 1)
		denominator := tf + k1*(1-b+b*docLen/avgDocLen)
		score += idf * numerator / denominator * word.Weight // 品詞の重みも考慮
	}

	return score
}

// calculateTF は単語の出現頻度を計算します
func calculateTF(doc []Word, lemma string) float64 {
	count := 0.0
	for _, word := range doc {
		if word.Lemma == lemma {
			count++
		}
	}
	return count
}

// containsWord は文書に指定された単語が含まれているかチェックします
func containsWord(doc []Word, lemma string) bool {
	for _, word := range doc {
		if word.Lemma == lemma {
			return true
		}
	}
	return false
}

// GenerateSummary は記事本文からサマリ（要約）を生成します。
func (p *HTMLParser) GenerateSummary(content string) (string, error) {
	if content == "" {
		return "", ErrEmptyContent
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("HTMLのパース中にエラーが発生しました: %w", err)
	}
	text := doc.Find("body").Text()
	text = p.normalizeWhitespace(text)

	sentences := p.splitSentences(text)
	if len(sentences) <= 2 {
		return text, nil
	}

	// 形態素解析と文のベクトル化
	vectors := make([][]Word, len(sentences))

	// エラー処理を追加
	if err := p.processVectors(vectors, sentences); err != nil {
		return "", fmt.Errorf("文のベクトル化に失敗しました: %w", err)
	}

	// 平均文長の計算
	var totalLen float64
	for _, sent := range sentences {
		totalLen += float64(len(sent))
	}
	avgDocLength := totalLen / float64(len(sentences))

	// BM25スコアの計算
	scores := make([]float64, len(sentences))
	for i, vec := range vectors {
		scores[i] = calculateBM25Score(vec, vectors, avgDocLength)
	}

	// スコアの高い文を選択
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

	// 上位2文を元の順序で結合
	var summary []string
	for i := 0; i < len(sentences) && len(summary) < 2; i++ {
		for _, r := range ranked {
			if r.index == i {
				summary = append(summary, sentences[i])
				break
			}
		}
	}

	return strings.Join(summary, ""), nil
}

// processVectors は文のベクトル化を行います
func (p *HTMLParser) processVectors(vectors [][]Word, sentences []string) error {
	for i, sentence := range sentences {
		words, err := p.tokenize(sentence)
		if err != nil {
			return fmt.Errorf("形態素解析に失敗しました: %w", err)
		}
		vectors[i] = words
	}
	return nil
}

// tokenize は文を形態素解析します
func (p *HTMLParser) tokenize(text string) ([]Word, error) {
	t, err := tokenizer.New(ipa.Dict())
	if err != nil {
		if p.logger != nil {
			p.logger.Error("形態素解析器の初期化に失敗しました",
				zap.Error(err),
			)
		}
		return nil, fmt.Errorf("%w: 形態素解析器の初期化に失敗しました", ErrTokenizer)
	}

	var words []Word
	tokens := t.Tokenize(text)

	for _, token := range tokens {
		features := token.Features()
		if len(features) < 7 {
			continue
		}

		pos := features[0]
		if len(features) > 1 {
			pos += "-" + features[1]
		}

		weight := getWordWeight(pos)
		if weight > 0 {
			word := Word{
				Surface: token.Surface,
				Lemma:   features[6],
				POS:     pos,
				Weight:  weight,
			}
			words = append(words, word)
		}
	}

	return words, nil
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

// splitSentences は文を分割します
func (p *HTMLParser) splitSentences(text string) []string {
	sentences := strings.Split(text, "。")
	var result []string
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

// isSentenceEnd は文末かどうかを判定します
func isSentenceEnd(surface string, features []string) bool {
	// 句読点チェック
	if surface == "。" || surface == "！" || surface == "？" ||
		surface == "." || surface == "!" || surface == "?" {
		return true
	}

	// 品詞チェック
	if len(features) > 1 && features[0] == "記号" &&
		(features[1] == "句点" || features[1] == "終助詞") {
		return true
	}

	return false
}
