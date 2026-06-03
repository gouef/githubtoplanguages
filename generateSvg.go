package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"math"
	"os"
	"strings"

	"github.com/gouef/githubtoplanguages/requests"
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
	Streak *requests.StreakStats
	Stats  *requests.GlobalStats
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

func generateSvg(languages []*Language, output string, streak *requests.StreakStats, stats *requests.GlobalStats) string {

	panelHeight := 95
	languageSvg := &LanguageToSvg{Size: 450, All: languages, Height: panelHeight, Streak: streak, Stats: stats}
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

	maxLangOffset := offsetL
	if offsetR > offsetL {
		maxLangOffset = offsetR
	}

	statsListHeight := 155

	if offsetL > statsListHeight {
		languageSvg.Height = panelHeight + maxLangOffset
	} else {
		languageSvg.Height = panelHeight + statsListHeight
	}

	if languageSvg.Streak != nil {
		languageSvg.Height += 130
	}

	content := SvgTemplate
	funcMap := template.FuncMap{
		"subtract": func(a, b int) int {
			return a - b
		},
		"multiply": func(a, b int) int {
			return a * b
		},
		"formatPercent": func(f float64) string {
			return fmt.Sprintf("%.2f", f)
		},
		"formatCount": func(count int) string {
			if count >= 1000 {
				return fmt.Sprintf("%.1fK", float64(count)/1000.0)
			}
			return fmt.Sprintf("%d", count)
		},
	}

	tp := template.New("svg").Funcs(funcMap)
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

	if output == "" {
		output = "./toplanguages.svg"
	} else {
		output = "./" + output + ".svg"
	}

	file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Failed to generate svg: %v", err)
	}
	defer file.Close()
	if _, err := file.WriteString(result); err != nil {
		log.Fatalf("failed to write log entry: %v", err)
	}

	return result
}
