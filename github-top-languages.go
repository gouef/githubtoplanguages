package main

import (
	"flag"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gouef/githubtoplanguages/requests"
	"github.com/gouef/utils"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	tokenFlag := flag.String("gh-token", "", "Github API token")
	userFlag := flag.String("user", "", "Github username")
	limitFlag := flag.Int("limit", 6, "Limit of languages")
	outputFlag := flag.String("output", "", "Name of file (without .svg")
	ignoredOrgsFlag := flag.String("ignore-orgs", "", "Comma-separated list of ignored organizations")
	ignoredReposFlag := flag.String("ignore-repos", "", "Comma-separated list of ignored repositories")
	ignoredLangsFlag := flag.String("ignore-langs", "", "Comma-separated list of ignored languages")
	withForksFlag := flag.String("with-forks", "false", "Include forked repositories in the analysis (true/false)")

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
	withForks, _ := strconv.ParseBool(*withForksFlag)

	if os.Getenv("GITHUB_WITH_FORKS") != "" {
		withForks = os.Getenv("GITHUB_WITH_FORKS") == "true"
	}

	output := *outputFlag

	if output == "" {
		output = "toplanguages"
	}

	ignoredOrganizations := explode(",", getPriorityValue(*ignoredOrgsFlag, "GITHUB_IGNORE_ORGANIZATIONS"))
	ignoredRepositories := explode(",", getPriorityValue(*ignoredReposFlag, "GITHUB_IGNORE_REPOS"))
	ignoredLanguages := explode(",", getPriorityValue(*ignoredLangsFlag, "GITHUB_IGNORE_LANGS"))
	ignored := append(ignoredOrganizations, ignoredRepositories...)

	limitEnv := os.Getenv("GITHUB_TOP_LIMIT")
	limit := *limitFlag

	if limitEnv != "" {
		limit, _ = strconv.Atoi(limitEnv)
	}

	var ignoredSize int

	// Organizations
	result, err := requests.FetchOrganizations(user, token, ignored...)

	if err != nil {
		log.Fatalf("Failed to fetch organizations: %v", err)
	}

	var repositories []string
	languages := make(map[string]int)
	colors := make(map[string]string)

	log.Println("Processing organizations..." + strconv.Itoa(len(result.Repositories)))

	for _, repoList := range result.Repositories {
		repositories = append(repositories, repoList.Name)

		for _, lang := range repoList.Languages {
			if isImageLanguage(lang.Name) {
				continue
			}
			if shouldIgnoreLanguage(lang.Name, ignoredLanguages) {
				ignoredSize += lang.Size
				continue
			}
			languages[lang.Name] += lang.Size
			colors[lang.Name] = lang.Color
		}
	}

	// User
	result, err = requests.FetchUser(user, token, false, ignored...)

	if err != nil {
		log.Fatalf("Failed to fetch user: %v", err)
	}

	for _, repoList := range result.Repositories {
		if utils.InArray(repoList.Name, repositories) {
			continue
		}

		repositories = append(repositories, repoList.Name)

		for _, lang := range repoList.Languages {
			if isImageLanguage(lang.Name) {
				continue
			}
			if shouldIgnoreLanguage(lang.Name, ignoredLanguages) {
				ignoredSize += lang.Size
				continue
			}
			languages[lang.Name] += lang.Size
			colors[lang.Name] = lang.Color
		}
	}

	// Forks
	if withForks {
		result, err = requests.FetchUser(user, token, true, ignored...)
		if err != nil {
			log.Fatalf("Failed to fetch user forks: %v", err)
		}
		for _, repoList := range result.Repositories {
			for _, lang := range repoList.Languages {
				if isImageLanguage(lang.Name) {
					continue
				}
				if shouldIgnoreLanguage(lang.Name, ignoredLanguages) {
					ignoredSize += lang.Size
					continue
				}
				languages[lang.Name] += lang.Size
				colors[lang.Name] = lang.Color
			}
		}
	}

	log.Println("Downloading latest language definitions from GitHub Linguist...")
	extensionMap, err := requests.LoadLinguistLanguages()
	if err != nil {
		log.Fatalf("Failed to load dynamic languages: %v", err)
	}

	//PRs
	prResult, err := requests.FetchUserPRLanguages(token, extensionMap, ignored...)
	if err != nil {
		log.Fatalf("Failed to fetch PR languages: %v", err)
	}
	for _, repoList := range prResult.Repositories {
		for _, lang := range repoList.Languages {
			if isImageLanguage(lang.Name) {
				continue
			}
			if shouldIgnoreLanguage(lang.Name, ignoredLanguages) {
				ignoredSize += lang.Size
				continue
			}
			languages[lang.Name] += lang.Size
			colors[lang.Name] = lang.Color
		}
	}

	resultLanguages := sortLanguages(languages, limit, ignoredSize)

	for _, lang := range resultLanguages {
		lang.Color = colors[lang.Name]
	}

	generateSvg(resultLanguages, output)

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

type Language struct {
	Name       string
	Color      string
	Size       int
	Percentage float64
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
		})
	}

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
