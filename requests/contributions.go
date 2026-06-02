package requests

import (
	"encoding/json"
	"time"
)

type HistoryResponse struct {
	Data struct {
		User struct {
			CreatedAt string `json:"createdAt"`
		} `json:"user"`
	} `json:"data"`
}

type YearlyContributionResponse struct {
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
	AccountCreated     string
	CurrentStart       string
	CurrentEnd         string
	LongestStart       string
	LongestEnd         string
}

func FetchContributionStats(username, token string) (*StreakStats, error) {
	createdQuery := `query($user: String!) {
		user(login: $user) {
			createdAt
		}
	}`

	variables := map[string]interface{}{"user": username}
	resp, err := Request(token, createdQuery, variables)
	if err != nil {
		return nil, err
	}

	var histRes HistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&histRes); err != nil {
		resp.Body.Close()
		return nil, err
	}
	resp.Body.Close()

	createdTime, err := time.Parse(time.RFC3339, histRes.Data.User.CreatedAt)
	if err != nil {
		createdTime = time.Now().AddDate(-1, 0, 0)
	}

	startYear := createdTime.Year()
	currentYear := time.Now().Year()

	var allDays []struct {
		Count int
		Date  string
	}
	totalContributions := 0

	yearlyQuery := `query($user: String!, $from: DateTime!, $to: DateTime!) {
		user(login: $user) {
			contributionsCollection(from: $from, to: $to) {
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

	for year := startYear; year <= currentYear; year++ {
		fromTime := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
		toTime := time.Date(year, time.December, 31, 23, 59, 59, 0, time.UTC).Format(time.RFC3339)

		yearVars := map[string]interface{}{
			"user": username,
			"from": fromTime,
			"to":   toTime,
		}

		yResp, err := Request(token, yearlyQuery, yearVars)
		if err != nil {
			return nil, err
		}

		var yRes YearlyContributionResponse
		if err := json.NewDecoder(yResp.Body).Decode(&yRes); err != nil {
			yResp.Body.Close()
			return nil, err
		}
		yResp.Body.Close()

		calendar := yRes.Data.User.ContributionsCollection.ContributionCalendar
		totalContributions += calendar.TotalContributions

		var yearDays []struct {
			Count int
			Date  string
		}
		for _, week := range calendar.Weeks {
			for _, day := range week.ContributionDays {
				yearDays = append(yearDays, struct {
					Count int
					Date  string
				}{day.ContributionCount, day.Date})
			}
		}

		allDays = append(allDays, yearDays...)
	}

	stats := &StreakStats{
		TotalContributions: totalContributions,
		AccountCreated:     createdTime.Format("Jan _2, 2006"),
	}

	currentStreak := 0
	longestStreak := 0

	var curStart, curEnd, longStart, longEnd string
	var inStreak bool
	todayStr := time.Now().Format("2006-01-02")
	yesterdayStr := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	lastLiveStreak := 0
	lastLiveStart := ""
	lastLiveEnd := ""
	var lastActiveDate string

	for _, day := range allDays {
		if day.Count > 0 {
			lastActiveDate = day.Date

			if !inStreak {
				curStart = day.Date
				inStreak = true
			}
			currentStreak++
			curEnd = day.Date

			lastLiveStreak = currentStreak
			lastLiveStart = curStart
			lastLiveEnd = curEnd

			if currentStreak > longestStreak {
				longestStreak = currentStreak
				longStart = curStart
				longEnd = curEnd
			}
		} else {
			if day.Date == todayStr {
				continue
			}

			currentStreak = 0
			inStreak = false
		}
	}

	if lastActiveDate == todayStr || lastActiveDate == yesterdayStr {
		if lastActiveDate == yesterdayStr {
			currentStreak = lastLiveStreak + 1
			curStart = lastLiveStart
			curEnd = todayStr

			if currentStreak > longestStreak {
				longestStreak = currentStreak
				longStart = curStart
				longEnd = curEnd
			}
		} else {
			currentStreak = lastLiveStreak
			curStart = lastLiveStart
			curEnd = lastLiveEnd
		}
	} else {
		currentStreak = 0
		curStart = ""
		curEnd = ""
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
