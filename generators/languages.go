package generators

import "fmt"

type Language struct {
	Name       string
	Color      string
	Percentage float64
}

type LangSvg struct {
	Language   *Language
	Percentage string
	Offset     int
	Delay      int
}

type LanguageSize struct {
	Language *Language
	Size     float64
	X        float64
}

type LanguagesSection struct {
	Sizes  []LanguageSize
	Left   []*LangSvg
	Right  []*LangSvg
	Height int
}

func NewLanguagesSection(languages []*Language, barWidth int) *LanguagesSection {
	section := &LanguagesSection{}
	if len(languages) == 0 {
		return section
	}

	half := int(len(languages)+1) / 2
	x := 0.0
	offsetL, offsetR := 0, 0
	delayL, delayR := 450, 450

	for i, lang := range languages {
		size := float64(barWidth) * (lang.Percentage / 100)
		section.Sizes = append(section.Sizes, LanguageSize{Language: lang, Size: size, X: x})
		x += size

		if i < half {
			section.Left = append(section.Left, &LangSvg{Language: lang, Offset: offsetL, Percentage: fmt.Sprintf("%.2f", lang.Percentage), Delay: delayL})
			offsetL += 22
			delayL += 100
		} else {
			section.Right = append(section.Right, &LangSvg{Language: lang, Offset: offsetR, Percentage: fmt.Sprintf("%.2f", lang.Percentage), Delay: delayR})
			offsetR += 22
			delayR += 100
		}
	}

	if offsetL > offsetR {
		section.Height = offsetL
	} else {
		section.Height = offsetR
	}

	return section
}
