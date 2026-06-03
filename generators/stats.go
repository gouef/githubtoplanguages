package generators

import (
	"strings"

	"github.com/gouef/githubtoplanguages/requests"
)

type StatsFeature string

const (
	Repos   StatsFeature = "repos"
	Stars   StatsFeature = "stars"
	Forks   StatsFeature = "forks"
	Commits StatsFeature = "commits"
	PRs     StatsFeature = "prs"
	Reviews StatsFeature = "reviews"
	Issues  StatsFeature = "issues"
)

const ALL_STATS = "commits,prs,reviews,issues,stars,repos,followers"

type StatItem struct {
	Label string
	Value string
	Y     int
}

type StatsSection struct {
	Items  []StatItem
	Height int
}

func NewStatsSection(stats *requests.GlobalStats, show bool, featuresFilter string, formatCount func(int) string) *StatsSection {
	section := &StatsSection{}
	if !show || stats == nil {
		return section
	}

	if featuresFilter == "" {
		featuresFilter = ALL_STATS
	}

	features := strings.Split(strings.ToLower(featuresFilter), ",")
	yOffset := 2

	for _, f := range features {
		f = strings.TrimSpace(f)
		switch f {
		case "commits":
			section.Items = append(section.Items, StatItem{Label: "Total Commits:", Value: formatCount(stats.TotalCommits), Y: yOffset})
			yOffset += 22
		case "prs":
			section.Items = append(section.Items, StatItem{Label: "Merged PRs:", Value: formatCount(stats.MergedPRs), Y: yOffset})
			yOffset += 22
		case "reviews":
			section.Items = append(section.Items, StatItem{Label: "Code Reviews:", Value: formatCount(stats.TotalReviews), Y: yOffset})
			yOffset += 22
		case "issues":
			section.Items = append(section.Items, StatItem{Label: "Total Issues:", Value: formatCount(stats.TotalIssues), Y: yOffset})
			yOffset += 22
		case "stars":
			section.Items = append(section.Items, StatItem{Label: "Total Stars:", Value: formatCount(stats.TotalStars), Y: yOffset})
			yOffset += 22
		case "repos":
			section.Items = append(section.Items, StatItem{Label: "Total Repositories:", Value: formatCount(stats.TotalRepos), Y: yOffset})
			yOffset += 22
		case "followers":
			section.Items = append(section.Items, StatItem{Label: "Followers:", Value: formatCount(stats.Followers), Y: yOffset})
			yOffset += 22
		}
	}

	section.Height = yOffset
	return section
}
