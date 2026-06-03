package requests

import "encoding/json"

type GlobalStats struct {
	TotalStars   int
	TotalCommits int
	TotalPRs     int
	TotalIssues  int
	TotalReviews int
	MergedPRs    int
	Followers    int
	TotalRepos   int
	TotalForks   int
}

type GlobalStatsResponse struct {
	Data struct {
		User struct {
			Followers struct {
				TotalCount int `json:"totalCount"`
			} `json:"followers"`
			Issues struct {
				TotalCount int `json:"totalCount"`
			} `json:"issues"`
			PullRequests struct {
				TotalCount int `json:"totalCount"`
			} `json:"pullRequests"`
			MergedPRs struct {
				TotalCount int `json:"totalCount"`
			} `json:"mergedPRs"`
			ContributionsCollection struct {
				TotalPullRequestReviewContributions int `json:"totalPullRequestReviewContributions"`
				TotalCommitContributions            int `json:"totalCommitContributions"`
			} `json:"contributionsCollection"`
			Repositories struct {
				TotalCount int `json:"totalCount"`
				Nodes      []struct {
					StargazerCount int `json:"stargazerCount"`
					ForkCount      int `json:"forkCount"`
				} `json:"nodes"`
			} `json:"repositories"`
		} `json:"user"`
	} `json:"data"`
}

func FetchGlobalStats(username, token string) (*GlobalStats, error) {
	query := `query($user: String!) {
		user(login: $user) {
			followers {
				totalCount
			}
			issues {
				totalCount
			}
			pullRequests {
				totalCount
			}
			mergedPRs: pullRequests(states: MERGED) {
				totalCount
			}
			contributionsCollection {
				totalPullRequestReviewContributions
				totalCommitContributions
			}
			repositories(first: 100, ownerAffiliations: OWNER) {
				totalCount
				nodes {
					stargazerCount
					forkCount
				}
			}
		}
	}`

	variables := map[string]interface{}{"user": username}
	resp, err := Request(token, query, variables)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res GlobalStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	totalStars := 0
	totalForks := 0
	for _, repo := range res.Data.User.Repositories.Nodes {
		totalStars += repo.StargazerCount
		totalForks += repo.ForkCount
	}

	u := res.Data.User
	stats := &GlobalStats{
		TotalStars:   totalStars,
		TotalCommits: u.ContributionsCollection.TotalCommitContributions,
		TotalPRs:     u.PullRequests.TotalCount,
		TotalIssues:  u.Issues.TotalCount,
		TotalReviews: u.ContributionsCollection.TotalPullRequestReviewContributions,
		MergedPRs:    u.MergedPRs.TotalCount,
		Followers:    u.Followers.TotalCount,
		TotalRepos:   u.Repositories.TotalCount,
		TotalForks:   totalForks,
	}

	return stats, nil
}
