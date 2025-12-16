package plagiarism_analyzer

import (
	"github.com/Nikita-Smirnov-idk/plagiarism-service/pkg/text_analyzer"
	"github.com/Nikita-Smirnov-idk/plagiarism-service/pkg/text_extractor"
)

// PlagiarismChecker основной интерфейс для проверки плагиата
type PlagiarismChecker struct {
	extractor *text_extractor.TextExtractor
	analyzer  *text_analyzer.TextAnalyzer
}

func NewPlagiarismChecker(nGramSize int, threshold float64) *PlagiarismChecker {
	return &PlagiarismChecker{
		extractor: text_extractor.NewTextExtractor(),
		analyzer:  text_analyzer.NewTextAnalyzer(nGramSize, threshold),
	}
}

// CompareFiles сравнивает два файла по их URL
func (p *PlagiarismChecker) CompareFiles(fileURL1, fileURL2 string) (float64, error) {
	text1, err := p.extractor.ExtractFromURL(fileURL1)
	if err != nil {
		return 0.0, err
	}

	text2, err := p.extractor.ExtractFromURL(fileURL2)
	if err != nil {
		return 0.0, err
	}

	cleanText1 := p.extractor.CleanText(text1)
	cleanText2 := p.extractor.CleanText(text2)

	similarity := p.analyzer.CompareTexts(cleanText1, cleanText2)

	return similarity, nil
}

// IsPlagiarized проверяет, является ли схожесть плагиатом
func (p *PlagiarismChecker) IsPlagiarized(similarity float64) bool {
	return p.analyzer.IsPlagiarized(similarity)
}
