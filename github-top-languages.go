package main

import (
	"flag"
	"github.com/gouef/githubtoplanguages/requests"
	"github.com/gouef/utils"
	"github.com/joho/godotenv"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	godotenv.Load()

	tokenFlag := flag.String("token", "", "Github API token")
	userFlag := flag.String("user", "", "Github username")
	limitFlag := flag.Int("limit", 6, "Limit of languages")
	outputFlag := flag.String("output", "", "Name of file (without .svg")
	ignoredOrgsFlag := flag.String("ignore-orgs", "", "Comma-separated list of ignored organizations")
	ignoredReposFlag := flag.String("ignore-repos", "", "Comma-separated list of ignored repositories")
	flag.Parse()

	token := *tokenFlag
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	if token == "" {
		log.Fatal("GITHUB_TOKEN is not set")
	}

	user := *userFlag
	if user == "" {
		user = os.Getenv("GITHUB_USERNAME")
	}

	if user == "" {
		log.Fatal("GITHUB_USERNAME is not set")
	}

	output := *outputFlag

	if output == "" {
		output = "toplanguages"
	}

	ignoredOrganizations := explode(",", getPriorityValue(*ignoredOrgsFlag, "GITHUB_IGNORE_ORGANIZATIONS"))
	ignoredRepositories := explode(",", getPriorityValue(*ignoredReposFlag, "GITHUB_IGNORE_REPOS"))
	ignored := append(ignoredOrganizations, ignoredRepositories...)

	limitEnv := os.Getenv("GITHUB_TOP_LIMIT")
	limit := *limitFlag

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

	generateSvg(resultLanguages, output)

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

func getPriorityValue(flagValue, envKey string) string {
	if flagValue != "" {
		return flagValue
	}
	return os.Getenv(envKey)
}

func explode(delimiter, str string) []string {
	if str == "" {
		return []string{}
	}
	parts := strings.Split(str, delimiter)
	var cleaned []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}
