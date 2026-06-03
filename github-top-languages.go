package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gouef/githubtoplanguages/generators"
	"github.com/gouef/githubtoplanguages/requests"
	"github.com/gouef/utils"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	tokenFlag := flag.String("gh-token", "", "Github API token")
	userFlag := flag.String("user", "", "Github username")
	limitFlag := flag.Int("limit", 12, "Limit of languages")
	outputFlag := flag.String("output", "", "Name of file (without .svg")
	ignoredOrgsFlag := flag.String("ignore-orgs", "", "Comma-separated list of ignored organizations")
	ignoredReposFlag := flag.String("ignore-repos", "", "Comma-separated list of ignored repositories")
	ignoredLangsFlag := flag.String("ignore-langs", "", "Comma-separated list of ignored languages")
	withForksFlag := flag.String("with-forks", "false", "Include forked repositories in the analysis (true/false)")
	withStreakFlag := flag.String("with-streak", "true", "Include streak statistics of your github account (true/false)")
	showStatsFlag := flag.String("show-stats", "false", "Show statistics section")
	statsFeaturesFlag := flag.String("stats", "", "Comma-separated list of stats features to include")

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

	withForks, _ := strconv.ParseBool(getPriorityValue(*withForksFlag, "GITHUB_WITH_FORKS"))
	withStreak, _ := strconv.ParseBool(getPriorityValue(*withStreakFlag, "GITHUB_WITH_STREAK"))
	showStats, _ := strconv.ParseBool(getPriorityValue(*showStatsFlag, "GITHUB_SHOW_STATS"))
	statsFeatures := getPriorityValue(*statsFeaturesFlag, "GITHUB_STATS_FEATURES")

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
		if lang.Name == "Others" {
			lang.Color = "#d4d4d4"
			continue
		}
		lang.Color = colors[lang.Name]
	}

	var streakStats *requests.StreakStats = nil
	if withStreak {
		streakStats, err = requests.FetchContributionStats(user, token)
		if err != nil {
			log.Printf("Warning: Failed to fetch contribution streaks: %v", err)
			streakStats = &requests.StreakStats{}
		}
	}

	var globalStats *requests.GlobalStats = nil
	if showStats {
		globalStats, err = requests.FetchGlobalStats(user, token)
		if err != nil {
			log.Printf("Warning: Failed to fetch global stats: %v", err)
			globalStats = &requests.GlobalStats{}
		}
	}

	var generatorLanguages []*generators.Language
	for _, lang := range resultLanguages {
		generatorLanguages = append(generatorLanguages, &generators.Language{
			Name:       lang.Name,
			Color:      lang.Color,
			Percentage: lang.Percentage,
		})
	}

	generators.GenerateCard(generatorLanguages, streakStats, globalStats, withStreak, showStats, statsFeatures, output)

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
