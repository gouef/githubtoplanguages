package generators

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/gouef/githubtoplanguages/requests"
)

//go:embed svgTemplate.gohtml
var SvgTemplate string

type CardData struct {
	Languages *LanguagesSection
	Stats     *StatsSection
	Streaks   *StreaksSection
	Height    int
}

func GenerateCard(languages []*Language, streak *requests.StreakStats, stats *requests.GlobalStats, showStreaks bool, showStats bool, statsFeatures string, output string) string {
	log.Printf("Generating card with %d languages, streak: %v, stats: %v", len(languages), showStreaks, showStats)

	langSec := NewLanguagesSection(languages, 450)
	statsSec := NewStatsSection(stats, showStats, statsFeatures, formatCountHelper)
	streakSec := NewStreaksSection(streak, showStreaks)

	middleHeight := langSec.Height
	if showStats {
		statsListHeight := statsSec.Height
		if statsListHeight < 115 {
			statsListHeight = 115
		}
		if statsListHeight > middleHeight {
			middleHeight = statsListHeight
		}
	} else {
		if middleHeight < 25 && len(languages) > 0 {
			middleHeight = 25
		}
	}

	totalHeight := 95 + middleHeight + streakSec.Height

	cardData := &CardData{
		Languages: langSec,
		Stats:     statsSec,
		Streaks:   streakSec,
		Height:    totalHeight,
	}

	funcMap := template.FuncMap{
		"subtract": func(a, b int) int { return a - b },
		"multiply": func(a, b int) int { return a * b },
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
	tpl, err := tp.Parse(SvgTemplate)
	if err != nil {
		log.Fatalf("Failed to parse svg template: %v", err)
	}

	var builder strings.Builder
	if err := tpl.Execute(&builder, cardData); err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	saveToFile(builder.String(), output)

	return builder.String()
}

func formatCountHelper(count int) string {
	if count >= 1000 {
		return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintf("%.1f", float64(count)/1000.0), ".0"), ".") + "K"
	}
	return fmt.Sprintf("%d", count)
}

func saveToFile(content, output string) {
	if output == "" {
		output = "./toplanguages.svg"
	} else {
		output = "./" + output + ".svg"
	}
	_ = os.WriteFile(output, []byte(content), 0644)
}
