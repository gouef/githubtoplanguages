package main

import (
	"github.com/gouef/githubtoplanguages/requests"
	"github.com/gouef/utils"
	"github.com/joho/godotenv"
	"log"
	"os"
	"sort"
	"strconv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN is not set")
	}
	user := os.Getenv("GITHUB_USERNAME")
	if user == "" {
		log.Fatal("GITHUB_USERNAME is not set")
	}

	ignoredOrganizationsEnv := os.Getenv("GITHUB_IGNORE_ORGANIZATIONS")
	ignoredRepositoriesEnv := os.Getenv("GITHUB_IGNORE_REPOS")

	ignoredOrganizations := utils.Explode(",", ignoredOrganizationsEnv)
	ignoredRepositories := utils.Explode(",", ignoredRepositoriesEnv)
	ignored := ignoredOrganizations
	ignored = append(ignored, ignoredRepositories...)

	limitEnv := os.Getenv("GITHUB_TOP_LIMIT")
	limit := 10

	if limitEnv != "" {
		limit, _ = strconv.Atoi(limitEnv)
	}

	result, err := requests.FetchOrganizations(user, token, ignored...)

	if err != nil {
		log.Fatalf("Failed to fetch organizations: %v", err)
	}

	var repositories []string
	languages := make(map[string]int)
	colors := make(map[string]string)

	for _, repoList := range result.Repositories {
		repositories = append(repositories, repoList.Name)

		for _, lang := range repoList.Languages {
			languages[lang.Name] += lang.Size
			colors[lang.Name] = lang.Color
		}
	}

	result, err = requests.FetchUser(user, token, ignored...)

	if err != nil {
		log.Fatalf("Failed to fetch user: %v", err)
	}

	for _, repoList := range result.Repositories {
		if utils.InArray(repoList.Name, repositories) {
			continue
		}

		repositories = append(repositories, repoList.Name)

		for _, lang := range repoList.Languages {
			languages[lang.Name] += lang.Size
			colors[lang.Name] = lang.Color
		}
	}

	resultLanguages := sortLanguages(languages, limit)

	for _, lang := range resultLanguages {
		lang.Color = colors[lang.Name]
	}

	generateSvg(resultLanguages)

}

type Language struct {
	Name       string
	Color      string
	Size       int
	Percentage float64
}

func sortLanguages(languages map[string]int, limit int) []*Language {
	type kv struct {
		Key   string
		Value int
	}

	var sorted []kv
	for k, v := range languages {
		sorted = append(sorted, kv{k, v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	if limit > len(sorted) {
		limit = len(sorted)
	}

	sortedMap := make(map[string]int, limit)

	var languagesSize int
	for i := 0; i < limit; i++ {
		sortedMap[sorted[i].Key] = sorted[i].Value
		languagesSize += sorted[i].Value
	}
	result := make([]*Language, 0)

	for key, size := range sortedMap {
		result = append(result, &Language{Name: key, Size: size, Percentage: (float64(size) / float64(languagesSize)) * 100})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Percentage > result[j].Percentage
	})

	return result
}
