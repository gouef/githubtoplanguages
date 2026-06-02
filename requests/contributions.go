package requests

import (
	"encoding/json"
	"time"
)

type ContributionResponse struct {
	Data struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					TotalContributions int `json:"totalContributions"`
					Weeks              []struct {
						ContributionDays []struct {
							ContributionCount int    `json:"contributionCount"`
							Date              string `json:"date"`
						} `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	} `json:"data"`
}

type StreakStats struct {
	TotalContributions int
	CurrentStreak      int
	LongestStreak      int
	CurrentStart       string
	CurrentEnd         string
	LongestStart       string
	LongestEnd         string
}

func FetchContributionStats(username, token string) (*StreakStats, error) {
	query := `query($user: String!) {
		user(login: $user) {
			contributionsCollection {
				contributionCalendar {
					totalContributions
					weeks {
						contributionDays {
							contributionCount
							date
						}
					}
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"user": username,
	}

	resp, err := Request(token, query, variables)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res ContributionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	calendar := res.Data.User.ContributionsCollection.ContributionCalendar
	stats := &StreakStats{
		TotalContributions: calendar.TotalContributions,
	}

	var allDays []struct {
		Count int
		Date  string
	}
	for _, week := range calendar.Weeks {
		for _, day := range week.ContributionDays {
			allDays = append(allDays, struct {
				Count int
				Date  string
			}{day.ContributionCount, day.Date})
		}
	}

	currentStreak := 0
	longestStreak := 0

	var curStart, curEnd, longStart, longEnd string
	var inStreak bool

	for _, day := range allDays {
		if day.Count > 0 {
			if !inStreak {
				curStart = day.Date
				inStreak = true
			}
			currentStreak++
			curEnd = day.Date

			if currentStreak > longestStreak {
				longestStreak = currentStreak
				longStart = curStart
				longEnd = curEnd
			}
		} else {
			todayStr := time.Now().Format("2006-01-02")
			if day.Date == todayStr && currentStreak > 0 {
				continue
			}

			currentStreak = 0
			inStreak = false
		}
	}

	stats.CurrentStreak = currentStreak
	stats.LongestStreak = longestStreak
	stats.CurrentStart = formatDateBounds(curStart)
	stats.CurrentEnd = formatDateBounds(curEnd)
	stats.LongestStart = formatDateBounds(longStart)
	stats.LongestEnd = formatDateBounds(longEnd)

	return stats, nil
}

func formatDateBounds(dateStr string) string {
	if dateStr == "" {
		return ""
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("Jan _2")
}
