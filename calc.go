package main

import "fmt"

func Calc() {
	data := map[string]int{
		"Go":           50841,
		"TypeScript":   16794,
		"HTML":         29942,
		"CSS":          191965,
		"Makefile":     7932,
		"Dockerfile":   87854,
		"Vala":         228356,
		"Shell":        59461,
		"CoffeeScript": 3326,
		"Latte":        7215,
		"AMPL":         239,
		"Procfile":     40,
		//"Lua":          3786894,
		"PHP":        483471,
		"JavaScript": 217779,
	}

	// Spočítat celkový počet bajtů
	total := 0
	for _, v := range data {
		total += v
	}

	// Vytvořit mapu s procenty
	percentages := make(map[string]float64)
	for lang, bytes := range data {
		percentages[lang] = (float64(bytes) / float64(total)) * 100
	}

	// Výpis výsledku
	fmt.Println("Podíl jazyků v procentech:")
	for lang, percent := range percentages {
		fmt.Printf("%s: %.2f%%\n", lang, percent)
	}
}
