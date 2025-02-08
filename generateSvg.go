package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"math"
	"os"
	"strings"
)

//go:embed svgTemplate.gohtml
var SvgTemplate string

type LanguageToSvg struct {
	Left   []*LangSvg
	Right  []*LangSvg
	All    []*Language
	Size   int
	Sizes  []LanguageSize
	Height int
}

type LanguageSize struct {
	Language *Language
	Size     float64
	X        float64
}

type LangSvg struct {
	Language   *Language
	Percentage string
	Offset     int
	Delay      int
}

func generateSvg(languages []*Language) {
	languageSvg := &LanguageToSvg{Size: 250, All: languages, Height: 95}
	half := math.Round(float64(len(languages)) / 2)
	x := 0.0

	offsetL := 0
	offsetR := 0
	delayL := 450
	delayR := 450
	for i, lang := range languages {
		if float64(i) < half {
			languageSvg.Left = append(languageSvg.Left, &LangSvg{Language: lang, Offset: offsetL, Percentage: fmt.Sprintf("%.2f", lang.Percentage), Delay: delayL})
			offsetL += 25
			delayL += 150
		} else {
			languageSvg.Right = append(languageSvg.Right, &LangSvg{Language: lang, Offset: offsetR, Percentage: fmt.Sprintf("%.2f", lang.Percentage), Delay: delayR})
			offsetR += 25
			delayR += 150
		}
		size := float64(languageSvg.Size) * (lang.Percentage / 100)
		languageSvg.Sizes = append(languageSvg.Sizes, LanguageSize{Language: lang, Size: size, X: x})
		x += size
	}

	if offsetL > offsetR {
		languageSvg.Height += offsetL
	} else {
		languageSvg.Height += offsetR
	}
	content := SvgTemplate
	tp := template.New("svg")
	tpl, err := tp.Parse(content)
	if err != nil {
		log.Fatalf("Failed to generate svg: %v", err)
	}

	var builder strings.Builder

	err = tpl.Execute(&builder, languageSvg)
	if err != nil {
		log.Fatalf("Failed to generate svg: %v", err)
	}

	result := builder.String()

	file, err := os.OpenFile("./toplanguages.svg", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Failed to generate svg: %v", err)
	}
	defer file.Close()
	if _, err := file.WriteString(result); err != nil {
		log.Fatalf("failed to write log entry: %w", err)
	}
}
