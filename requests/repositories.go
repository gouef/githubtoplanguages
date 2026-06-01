package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gouef/utils"
)

type GraphQLUserResponse struct {
	Data DataUser `json:"data"`
}

type DataUser struct {
	Viewer ViewerUser `json:"viewer"`
}

type ViewerUser struct {
	Repositories UserRepositories `json:"repositories"`
	PullRequests UserPullRequests `json:"pullRequests"` // Přidáno pro získání PR nezávisle na repozitářích
}

type UserRepositories struct {
	Edges    []UserRepositoryEdge `json:"edges"`
	PageInfo PageInfo             `json:"pageInfo"`
}

type UserRepositoryEdge struct {
	Node UserRepositoryNode `json:"node"`
}

type UserRepositoryNode struct {
	Name            string          `json:"name"`
	NameWithOwner   string          `json:"nameWithOwner"`
	IsFork          bool            `json:"isFork"`
	PrimaryLanguage PrimaryLanguage `json:"primaryLanguage"`
	Languages       Languages       `json:"languages"`
}

// Struktury pro Pull Requesty a jejich zdrojové repozitáře
type UserPullRequests struct {
	Edges    []UserPullRequestEdge `json:"edges"`
	PageInfo PageInfo              `json:"pageInfo"`
}

type UserPullRequestEdge struct {
	Node UserPullRequestNode `json:"node"`
}

type UserPullRequestNode struct {
	Title      string             `json:"title"`
	Repository PRRepositoryDetail `json:"repository"` // Repozitář, do kterého PR směřuje (nebo z něj pochází)
}

type PRRepositoryDetail struct {
	Name          string    `json:"name"`
	NameWithOwner string    `json:"nameWithOwner"`
	Languages     Languages `json:"languages"`
}

func FetchUser(loginName, token string, withFork bool, ignored ...string) (*Result, error) {
	var isForkStr string
	if withFork {
		isForkStr = "true"
	} else {
		isForkStr = "false"
	}

	// Dotaz pro běžné repozitáře a forky (zůstává podobný, ale optimalizovaný)
	query := fmt.Sprintf(`query($after: String) {
		  viewer {
			repositories(first: 100, isFork: %s, affiliations: OWNER, ownerAffiliations: OWNER, after: $after) {
			  edges {
				node {
				  name
				  nameWithOwner
				  isFork
				  primaryLanguage {
					name
				  }
				  languages(first: 5, after: null) {
					edges {
					  node {
						name
						color
					  }
					  size
					}
				  }
				}
			  } 
			  pageInfo {
				hasNextPage
				endCursor
			  }
			}
		  }
		}`, isForkStr)

	finalResult := &Result{}
	var cursor interface{} = nil

	for {
		variables := map[string]interface{}{
			"after": cursor,
		}

		resp, err := Request(token, query, variables)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("GitHub API returned non-200 status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
		}

		var result GraphQLUserResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		repoData := result.Data.Viewer.Repositories

		for _, r := range repoData.Edges {
			if utils.InArray(r.Node.Name, ignored) {
				continue
			}
			
			resultRepository := &ResultRepository{
				Name:   r.Node.NameWithOwner,
				IsFork: r.Node.IsFork,
			}

			var languages []*ResultLanguage
			for _, l := range r.Node.Languages.Edges {
				languages = append(languages, &ResultLanguage{Name: l.Node.Name, Size: l.Size, Color: l.Node.Color})
			}
			resultRepository.Languages = languages

			finalResult.Repositories = append(finalResult.Repositories, resultRepository)
		}

		if !repoData.PageInfo.HasNextPage {
			break
		}

		cursor = repoData.PageInfo.EndCursor
	}

	return finalResult, nil
}

// Nová samostatná funkce pro vytažení repozitářů skrze uživatelovy Pull Requesty
func FetchUserPRLanguages(token string, ignored ...string) (*Result, error) {
	query := `query($after: String) {
		viewer {
			pullRequests(first: 50, after: $after, states: [OPEN, MERGED]) {
				edges {
					node {
						title
						repository {
							name
							nameWithOwner
							languages(first: 5) {
								edges {
									node {
										name
										color
									}
									size
								}
							}
						}
					}
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	}`

	finalResult := &Result{}
	var cursor interface{} = nil

	for {
		variables := map[string]interface{}{
			"after": cursor,
		}

		resp, err := Request(token, query, variables)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("GitHub API returned non-200 status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
		}

		var result GraphQLUserResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		prData := result.Data.Viewer.PullRequests

		for _, edge := range prData.Edges {
			repo := edge.Node.Repository
			if repo.Name == "" || utils.InArray(repo.Name, ignored) {
				continue
			}

			resultRepository := &ResultRepository{
				Name:   repo.NameWithOwner,
				IsPR:   true,
			}

			var languages []*ResultLanguage
			for _, l := range repo.Languages.Edges {
				languages = append(languages, &ResultLanguage{Name: l.Node.Name, Size: l.Size, Color: l.Node.Color})
			}
			resultRepository.Languages = languages

			finalResult.Repositories = append(finalResult.Repositories, resultRepository)
		}

		if !prData.PageInfo.HasNextPage {
			break
		}
		cursor = prData.PageInfo.EndCursor
	}

	return finalResult, nil
}