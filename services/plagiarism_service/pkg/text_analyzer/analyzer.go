package text_analyzer

import (
	"strings"
)

// TextAnalyzer сравнивает два текста на схожесть
type TextAnalyzer struct {
	nGramSize int
	threshold float64
}

func NewTextAnalyzer(nGramSize int, threshold float64) *TextAnalyzer {
	if nGramSize < 2 {
		nGramSize = 3
	}
	if threshold <= 0 {
		threshold = 0.7
	}

	return &TextAnalyzer{
		nGramSize: nGramSize,
		threshold: threshold,
	}
}

// CompareTexts сравнивает два текста, возвращает процент схожести (0.0-1.0)
func (a *TextAnalyzer) CompareTexts(text1, text2 string) float64 {
	if text1 == "" || text2 == "" {
		return 0.0
	}

	ngrams1 := a.extractNGrams(text1)
	ngrams2 := a.extractNGrams(text2)

	if len(ngrams1) < a.nGramSize || len(ngrams2) < a.nGramSize {
		return a.simpleCompare(text1, text2)
	}

	intersection := 0
	for gram := range ngrams1 {
		if _, exists := ngrams2[gram]; exists {
			intersection++
		}
	}

	union := len(ngrams1) + len(ngrams2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// extractNGrams разбивает текст на n-граммы
func (a *TextAnalyzer) extractNGrams(text string) map[string]int {
	words := strings.Fields(text)
	if len(words) < a.nGramSize {
		ngrams := make(map[string]int)
		ngrams[strings.Join(words, " ")] = 1
		return ngrams
	}

	ngrams := make(map[string]int)
	for i := 0; i <= len(words)-a.nGramSize; i++ {
		gram := strings.Join(words[i:i+a.nGramSize], " ")
		ngrams[gram]++
	}

	return ngrams
}

// simpleCompare простой метод сравнения для коротких текстов
func (a *TextAnalyzer) simpleCompare(text1, text2 string) float64 {
	words1 := strings.Fields(text1)
	words2 := strings.Fields(text2)

	set1 := make(map[string]bool)
	for _, w := range words1 {
		set1[w] = true
	}

	intersection := 0
	for _, w := range words2 {
		if set1[w] {
			intersection++
		}
	}

	union := len(set1) + len(words2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// IsPlagiarized проверяет, является ли схожесть плагиатом
func (a *TextAnalyzer) IsPlagiarized(similarity float64) bool {
	return similarity >= a.threshold
}

// FindCommonSections находит общие фрагменты в текстах
func (a *TextAnalyzer) FindCommonSections(text1, text2 string, minWords int) []string {
	words1 := strings.Fields(text1)
	words2 := strings.Fields(text2)

	var commonSections []string

	for i := 0; i < len(words1); i++ {
		for j := 0; j < len(words2); j++ {
			k := 0
			for i+k < len(words1) && j+k < len(words2) &&
				words1[i+k] == words2[j+k] {
				k++
			}

			if k >= minWords {
				section := strings.Join(words1[i:i+k], " ")
				commonSections = append(commonSections, section)
				i += k
				break
			}
		}
	}

	return commonSections
}

// PreprocessText предобрабатывает текст для сравнения
func (a *TextAnalyzer) PreprocessText(text string) string {
	text = strings.ToLower(text)
	text = strings.Join(strings.Fields(text), " ")
	return text
}
