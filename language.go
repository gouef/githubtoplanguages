package main

import (
	"sort"
	"strings"
)

type Language struct {
	Name       string
	Color      string
	Size       int
	Percentage float64
}

func isImageLanguage(langName string) bool {
	langLower := strings.ToLower(langName)
	return langLower == "svg" || langLower == "png" || langLower == "image" || langLower == "jpg"
}

func shouldIgnoreLanguage(langName string, ignoredLangs []string) bool {
	langLower := strings.ToLower(langName)
	for _, ignored := range ignoredLangs {
		if langLower == strings.ToLower(ignored) {
			return true
		}
	}
	return false
}

func sortLanguages(languages map[string]int, limit int, ignoredSize int) []*Language {
	type kv struct {
		Key   string
		Value int
	}

	var sorted []kv
	var totalSize int

	for k, v := range languages {
		sorted = append(sorted, kv{k, v})
		totalSize += v
	}
	totalSize += ignoredSize

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	hasOthers := ignoredSize > 0 || len(sorted) > limit

	mainLimit := limit
	if hasOthers && mainLimit > 1 {
		mainLimit = limit - 1
	}

	if mainLimit > len(sorted) {
		mainLimit = len(sorted)
	}

	result := make([]*Language, 0)

	for i := 0; i < mainLimit; i++ {
		size := sorted[i].Value
		result = append(result, &Language{
			Name:       sorted[i].Key,
			Size:       size,
			Percentage: (float64(size) / float64(totalSize)) * 100,
		})
	}

	var othersSize int
	for i := mainLimit; i < len(sorted); i++ {
		othersSize += sorted[i].Value
	}

	othersSize += ignoredSize

	if othersSize > 0 {
		result = append(result, &Language{
			Name:       "Others",
			Size:       othersSize,
			Percentage: (float64(othersSize) / float64(totalSize)) * 100,
			Color:      "#d4d4d4",
		})
	}

	return result
}
